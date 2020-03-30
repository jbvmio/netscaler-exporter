package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawLBVServerStats is the payload as returned by the Nitro API.
type RawLBVServerStats []byte

// Len returns the size of the underlying []byte.
func (r RawLBVServerStats) Len() int {
	return len(r)
}

// LBVServerStats represents the data returned from the /stat/service Nitro API endpoint
type LBVServerStats struct {
	Name                        string         `json:"name"`
	AvgTimeClientTTLB           string         `json:"avgcltttlb"`
	State                       CurState       `json:"state"`
	VSLBHealth                  string         `json:"vslbhealth"`
	TotalRequests               string         `json:"totalrequests"`
	TotalResponses              string         `json:"totalresponses"`
	TotalRequestBytes           string         `json:"totalrequestbytes"`
	TotalResponseBytes          string         `json:"totalresponsebytes"`
	TotalClientTTLBTransactions string         `json:"totcltttlbtransactions"`
	ActiveServices              string         `json:"actsvcs"`
	TotalHits                   string         `json:"tothits"`
	TotalPktsReceived           string         `json:"totalpktsrecvd"`
	TotalPktsSent               string         `json:"totalpktssent"`
	SurgeCount                  string         `json:"surgecount"`
	SvcSurgeCount               string         `json:"svcsurgecount"`
	VSvrSurgeCount              string         `json:"vsvrsurgecount"`
	Service                     []ServiceStats `json:"service"`
}

// NitroType implements the NitroData interface.
func (s LBVServerStats) NitroType() string {
	return lbvserverSubsystem
}

func processLBVServerStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := lbvserverSubsystem
	switch {
	case P.stopped:
		P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
	default:
		P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
		data := submitAPITask(P, netscaler.StatsTypeLBVServer)
		switch {
		case len(data) < 1:
			P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
			exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
			P.insertBackoff(thisSS)
		default:
			req := newNitroRawReq(RawLBVServerStats(data))
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
}
