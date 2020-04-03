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
	case P.metricFlipBit[thisSS].good():
		defer P.metricFlipBit[thisSS].flip()
		timeBegin := time.Now().UnixNano()
		switch {
		case P.stopped:
			P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
		default:
			P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
			data := submitAPITask(P, netscaler.StatsTypeSSL)
			switch {
			case len(data) < 1:
				P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
				exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
				P.insertBackoff(thisSS)
			default:
				req := newNitroRawReq(RawSSLStats(data))
				P.submit(req)
				s := <-req.ResultChan()
				if success, ok := s.(bool); ok {
					switch {
					case success:
						go TK.set(P.nsInstance, thisSS, float64(time.Now().UnixNano()))
						timeEnd := time.Now().UnixNano()
						exporterPromProcessingTime.WithLabelValues(P.nsInstance, thisSS).Set(float64((timeEnd - timeBegin) / nanoSecond))
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
