package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
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
	ServiceName                  string   `json:"servicename"`
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
	switch {
	case P.stopped:
		P.logger.Info("Skipping " + servicesSubsystem + " Stats, process is stopping")
	case !P.mappingsLoaded:
		P.logger.Info("unable to collect " + servicesSubsystem + " metrics, mapping not yet complete")
	default:
		P.logger.Info("Processing " + servicesSubsystem + " Stats")
		data := submitAPITask(P, netscaler.StatsTypeService)
		switch {
		case len(data) < 1:
			P.logger.Error("error retrieving data for " + servicesSubsystem + " Stats")
			exporterFailuresTotal.WithLabelValues(P.nsInstance, servicesSubsystem).Inc()
		default:
			req := newNitroRawReq(RawServiceStats(data))
			P.submit(req)
			s := <-req.ResultChan()
			if success, ok := s.(bool); ok {
				switch {
				case success:
					go TK.set(P.nsInstance, servicesSubsystem, float64(time.Now().UnixNano()))
				default:
					exporterFailuresTotal.WithLabelValues(P.nsInstance, servicesSubsystem).Inc()
				}
			}
			P.logger.Info(servicesSubsystem + " Stat Processing Complete")
		}
	}
}
