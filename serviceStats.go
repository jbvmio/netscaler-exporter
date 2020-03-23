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
		P.logger.Info("Skipping Service Stats, process is stopping")
	default:
		P.logger.Info("Processing Service Stats")
		data, err := P.client.GetAll(netscaler.StatsTypeService)
		switch {
		case err != nil:
			P.logger.Error("error retrieving data for Service Stats", zap.Error(err))
			exporterFailuresTotal.WithLabelValues(P.nsInstance, servicesSubsystem).Inc()
		default:
			req := newNitroRawReq(RawServiceStats(data))
			P.team.Submit(req)
			s := <-req.ResultChan()
			if success, ok := s.(bool); ok {
				switch {
				case success:
					go TK.set(P.nsInstance, servicesSubsystem, float64(time.Now().UnixNano()))
				default:
					exporterFailuresTotal.WithLabelValues(P.nsInstance, servicesSubsystem).Inc()
				}
			}
			P.logger.Info("Service Stat Processing Complete")
		}
	}
}
