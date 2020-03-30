package main

import (
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"go.uber.org/zap"
)

// RawNSStats is the payload as returned by the Nitro API.
type RawNSStats []byte

// Len returns the size of the underlying []byte.
func (r RawNSStats) Len() int {
	return len(r)
}

// NSStats represents the data returned from the /stat/service Nitro API endpoint
type NSStats struct {
	CPUUsagePct                            float64 `json:"cpuusagepcnt"`
	MemUsagePct                            float64 `json:"memusagepcnt"`
	MgmtCPUUsagePct                        float64 `json:"mgmtcpuusagepcnt"`
	PktCPUUsagePct                         float64 `json:"pktcpuusagepcnt"`
	FlashPartitionUsage                    float64 `json:"disk0perusage"`
	VarPartitionUsage                      float64 `json:"disk1perusage"`
	TotalReceivedMB                        string  `json:"totrxmbits"`
	TotalTransmitMB                        string  `json:"tottxmbits"`
	HTTPRequests                           string  `json:"httptotrequests"`
	HTTPResponses                          string  `json:"httptotresponses"`
	TCPCurrentClientConnections            string  `json:"tcpcurclientconn"`
	TCPCurrentClientConnectionsEstablished string  `json:"tcpcurclientconnestablished"`
	TCPCurrentServerConnections            string  `json:"tcpcurserverconn"`
	TCPCurrentServerConnectionsEstablished string  `json:"tcpcurserverconnestablished"`
}

// NitroType implements the NitroData interface.
func (s NSStats) NitroType() string {
	return nsSubsystem
}

func processNSStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := nsSubsystem
	switch {
	case P.stopped:
		P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
	default:
		P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
		data := submitAPITask(P, netscaler.StatsTypeNS)
		switch {
		case len(data) < 1:
			P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
			exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
			P.insertBackoff(thisSS)
		default:
			req := newNitroRawReq(RawNSStats(data))
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
