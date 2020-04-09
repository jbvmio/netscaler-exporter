package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawLBVServerConfigs is the payload as returned by the Nitro API.
type RawLBVServerConfigs []byte

// Len returns the size of the underlying []byte.
func (r RawLBVServerConfigs) Len() int {
	return len(r)
}

// LBVServerConfigs represents the data returned from the /stat/service Nitro API endpoint
type LBVServerConfigs struct {
	Name                   string `json:"name"`
	StateChangeTimeSeconds string `json:"statechangetimeseconds"`
}

// NitroType implements the NitroData interface.
func (s LBVServerConfigs) NitroType() string {
	return lbvserverConfigSubsystem
}

func processLBVServerConfigs(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := lbvserverConfigSubsystem
	switch {
	case P.metricFlipBit[thisSS].good():
		defer P.metricFlipBit[thisSS].flip()
		timeBegin := time.Now().UnixNano()
		switch {
		case P.stopped:
			P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
		default:
			if P.collectMappings {
				if !P.mappingsLoaded {
					P.logger.Warn("unable to collect subSystem metrics, mapping not yet complete", zap.String("subSystem", thisSS))
					return
				}
			}
			P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
			data := submitAPITask(P, netscaler.ConfigTypeLBVServer)
			switch {
			case len(data) < 1:
				P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
				exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
				P.insertBackoff(thisSS)
			default:
				req := newNitroRawReq(RawLBVServerConfigs(data))
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
