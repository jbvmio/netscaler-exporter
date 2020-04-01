package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawGSLBVServerStats is the payload as returned by the Nitro API.
type RawGSLBVServerStats []byte

// Len returns the size of the underlying []byte.
func (r RawGSLBVServerStats) Len() int {
	return len(r)
}

// GSLBVServerStats represents the data returned from the /stat/service Nitro API endpoint
type GSLBVServerStats struct {
	Name                     string             `json:"name"`
	EstablishedConnections   string             `json:"establishedconn"`
	State                    CurState           `json:"state"`
	Health                   string             `json:"vslbhealth"`
	InactiveServices         string             `json:"inactsvcs"`
	ActiveServices           string             `json:"actsvcs"`
	TotalHits                string             `json:"tothits"`
	TotalRequests            string             `json:"totalrequests"`
	TotalResponses           string             `json:"totalresponses"`
	TotalRequestBytes        string             `json:"totalrequestbytes"`
	TotalResponseBytes       string             `json:"totalresponsebytes"`
	CurrentClientConnections string             `json:"curclntconnections"`
	CurrentServerConnections string             `json:"cursrvrconnections"`
	GSLBService              []GSLBServiceStats `json:"gslbservice"`
}

// GSLBServiceStats represents the data returned from the /stat/service Nitro API endpoint
type GSLBServiceStats struct {
	GSLBName                 string   `json:"gslbname"`
	ServiceName              string   `json:"servicename"`
	EstablishedConnections   string   `json:"establishedconn"`
	State                    CurState `json:"state"`
	TotalRequests            string   `json:"totalrequests"`
	TotalResponses           string   `json:"totalresponses"`
	TotalRequestBytes        string   `json:"totalrequestbytes"`
	TotalResponseBytes       string   `json:"totalresponsebytes"`
	CurrentServerConnections string   `json:"cursrvrconnections"`
	CurrentLoad              string   `json:"curload"`
	ServiceHits              string   `json:"vsvrservicehits"`
}

// NitroType implements the NitroData interface.
func (s GSLBVServerStats) NitroType() string {
	return gslbVServerSubsystem
}

func processGSLBVServerStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := gslbVServerSubsystem
	switch {
	case P.metricFlipBit[thisSS].good():
		defer P.metricFlipBit[thisSS].flip()
		switch {
		case P.stopped:
			P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
		default:
			P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
			gslbvServers, err := GetGSLBServerServiceStats(P)
			switch {
			case err != nil:
				P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
				P.insertBackoff(thisSS)
			default:
				P.logger.Debug("processing lbservice stats", zap.String("subSystem", thisSS), zap.Int("number of lbvservers", len(gslbvServers)))
				for _, svr := range gslbvServers {
					req := newNitroDataReq(svr)
					success := P.submit(req)
					if !success {
						exporterProcessingFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
					}
				}
				go TK.set(P.nsInstance, thisSS, float64(time.Now().UnixNano()))
				P.logger.Debug("subSystem stat collection Complete", zap.String("subSystem", thisSS))
			}
		}
	default:
		P.logger.Debug("subSystem stat collection already in progress", zap.String("subSystem", thisSS))
	}
}

// GetGSLBServerServiceStats retrieves stats for both GSLBServers and GSLBServices.
func GetGSLBServerServiceStats(P *Pool) ([]GSLBVServerStats, error) {
	var gslbVServers []GSLBVServerStats
	servers, err := getGSLBVServerStats(P.client)
	if err != nil {
		exporterAPICollectFailures.WithLabelValues(P.nsInstance, gslbVServerSubsystem).Inc()
		return gslbVServers, err
	}
	for _, svr := range servers {
		var retries int
		s, err := getGSLBVServerStats(P.client, svr.Name)
	retryLoop:
		for err != nil {
			exporterAPICollectFailures.WithLabelValues(P.nsInstance, gslbVServerSvcSubsystem).Inc()
			if retries >= 3 {
				break retryLoop
			}
			time.Sleep(time.Millisecond * 100)
			s, err = getGSLBVServerStats(P.client, svr.Name)
			retries++
		}
		if err == nil {
			gslbVServers = append(gslbVServers, s...)
		}
	}
	return gslbVServers, nil
}

func getGSLBVServerStats(client *netscaler.NitroClient, target ...string) ([]GSLBVServerStats, error) {
	var gslbVServers []GSLBVServerStats
	var b []byte
	var err error
	switch len(target) {
	case 0:
		b, err = client.GetAll(netscaler.StatsTypeGSLBVServer)
	default:
		svr := target[0]
		b, err = client.Get(netscaler.StatsTypeGSLBVServer, svr+`?statbindings=yes`)
	}
	if err != nil {
		return gslbVServers, err
	}
	tmp := struct {
		Target *[]GSLBVServerStats `json:"gslbvserver"`
	}{Target: &gslbVServers}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return gslbVServers, err
	}
	return gslbVServers, nil
}
