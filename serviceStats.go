package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawServiceStats is the payload as returned by the Nitro API.
type RawServiceStats []byte

// Len returns the size of the underlying []byte.
func (r RawServiceStats) Len() int {
	return len(r)
}

// ServiceStats represents the data returned from the /stat/service Nitro API endpoint
type ServiceStats struct {
	Name                         string   `json:"name"`
	Throughput                   string   `json:"throughput"`
	AvgTimeToFirstByte           string   `json:"avgsvrttfb"`
	State                        CurState `json:"state"`
	TotalRequests                string   `json:"totalrequests"`
	TotalResponses               string   `json:"totalresponses"`
	TotalRequestBytes            string   `json:"totalrequestbytes"`
	TotalResponseBytes           string   `json:"totalresponsebytes"`
	CurrentClientConnections     string   `json:"curclntconnections"`
	SurgeCount                   string   `json:"surgecount"`
	CurrentServerConnections     string   `json:"cursrvrconnections"`
	ServerEstablishedConnections string   `json:"svrestablishedconn"`
	CurrentReusePool             string   `json:"curreusepool"`
	MaxClients                   string   `json:"maxclients"`
	CurrentLoad                  string   `json:"curload"`
	ServiceHits                  string   `json:"vsvrservicehits"`
	ActiveTransactions           string   `json:"activetransactions"`
}

// NitroType implements the NitroData interface.
func (s ServiceStats) NitroType() string {
	return servicesSubsystem
}

func processSvcStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := servicesSubsystem
	switch {
	case P.metricFlipBit[thisSS].good():
		defer P.metricFlipBit[thisSS].flip()
		switch {
		case P.stopped:
			P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
		default:
			P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
			data := submitAPITask(P, netscaler.StatsTypeService)
			switch {
			case len(data) < 1:
				P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
				exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
				P.insertBackoff(thisSS)
			default:
				req := newNitroRawReq(RawServiceStats(data))
				P.submit(req)
				s := <-req.ResultChan()
				if success, ok := s.(bool); ok {
					switch {
					case success:
						go TK.set(P.nsInstance, thisSS, float64(time.Now().UnixNano()))
					default:
						exporterProcessingFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
					}
				}
				P.logger.Debug("subSystem stat collection Complete", zap.String("subSystem", thisSS))
			}
		}
	default:
		P.logger.Debug("subSystem stat collection already in progress", zap.String("subSystem", thisSS))
	}
}
