package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawSSLStats is the payload as returned by the Nitro API.
type RawSSLStats []byte

// Len returns the size of the underlying []byte.
func (r RawSSLStats) Len() int {
	return len(r)
}

// SSLStats represents the data returned from the /stat/service Nitro API endpoint
type SSLStats struct {
	TotalSSLTransactions string `json:"ssltottransactions"`
	TotalSSLSessions     string `json:"ssltotsessions"`
	SSLSessions          string `json:"sslcursessions"`
}

// NitroType implements the NitroData interface.
func (s SSLStats) NitroType() string {
	return sslSubsystem
}

func processSSLStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := sslSubsystem
	switch {
	case P.stopped:
		P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
	default:
		P.logger.Info("Processing subSystem Stats", zap.String("subSystem", thisSS))
		data := submitAPITask(P, netscaler.StatsTypeSSL)
		switch {
		case len(data) < 1:
			P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
			exporterFailuresTotal.WithLabelValues(P.nsInstance, thisSS).Inc()
			P.insertBackoff(thisSS)
		default:
			req := newNitroRawReq(RawSSLStats(data))
			P.submit(req)
			s := <-req.ResultChan()
			if success, ok := s.(bool); ok {
				switch {
				case success:
					go TK.set(P.nsInstance, thisSS, float64(time.Now().UnixNano()))
				default:
					exporterFailuresTotal.WithLabelValues(P.nsInstance, thisSS).Inc()
				}
			}
			P.logger.Info("subSystem stat collection Complete", zap.String("subSystem", thisSS))
		}
	}
}
