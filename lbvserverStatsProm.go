package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/load-balancing/lbvserver/lbvserver/

const lbvserverSubsystem = `lbvserver`

var (
	lbvserverLabels    = []string{netscalerInstance, `citrixadc_lb_name`, `citrixadc_lb_type`}
	lbvserverAveCLTTLB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "average_client_time_to_lb",
			Help:      "Average time interval between sending the request packet to a service and receiving the ACK for response from the client",
		},
		lbvserverLabels,
	)

	lbvserverState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "state",
			Help:      "Current state of the server. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		lbvserverLabels,
	)

	lbvserverTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "requests_total",
			Help:      "Total number of requests received on this virtual server",
		},
		lbvserverLabels,
	)

	lbvserverTotalResponses = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "responses_total",
			Help:      "Total number of responses received on this virtual server",
		},
		lbvserverLabels,
	)

	lbvserverTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this virtual server",
		},
		lbvserverLabels,
	)

	lbvserverTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this virtual server",
		},
		lbvserverLabels,
	)

	lbvserverTotalClientTTLBTrans = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "client_ttlb_total",
			Help:      "Total transactions where client TTLB is calculated",
		},
		lbvserverLabels,
	)
	lbvserverActiveServices = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "active_services",
			Help:      "Number of ACTIVE services bound to a vserver",
		},
		lbvserverLabels,
	)
	lbvserverTotalHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "hits_total",
			Help:      "Total vserver hits",
		},
		lbvserverLabels,
	)
	lbvserverTotalPktsRx = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "pkts_received_total",
			Help:      "Total number of packets received on this virtual server",
		},
		lbvserverLabels,
	)
	lbvserverTotalPktsTx = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "pkts_sent_total",
			Help:      "Total number of packets sent on this virtual server",
		},
		lbvserverLabels,
	)

	lbvserverSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "surge_queue",
			Help:      "Number of requests in the surge queue",
		},
		lbvserverLabels,
	)
	lbvserverSvcSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "svc_surge_queue",
			Help:      "Total number of requests in the surge queues of all the services bound to this LB-vserver",
		},
		lbvserverLabels,
	)
	lbvserverVSvrSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSubsystem,
			Name:      "vsvr_surge_queue",
			Help:      "Number of requests waiting on this vserver",
		},
		lbvserverLabels,
	)
)

func (P *Pool) promLBVServerStats(N NitroData) {
	switch ss := N.(type) {
	case LBVServerStats:
		lbvserverAveCLTTLB.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.AvgTimeClientTTLB))
		lbvserverState.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.State.Value()))
		lbvserverTotalRequests.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalRequests))
		lbvserverTotalResponses.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalResponses))
		lbvserverTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalRequestBytes))
		lbvserverTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalResponseBytes))
		lbvserverTotalClientTTLBTrans.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalClientTTLBTransactions))
		lbvserverActiveServices.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.ActiveServices))
		lbvserverTotalHits.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalHits))
		lbvserverTotalPktsRx.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalPktsReceived))
		lbvserverTotalPktsTx.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalPktsSent))
		lbvserverSurgeCount.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.SurgeCount))
		lbvserverSvcSurgeCount.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.SvcSurgeCount))
		lbvserverVSvrSurgeCount.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.VSvrSurgeCount))
		for _, svc := range ss.LBService {
			lbvsvrServiceThroughput.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.Throughput) * 1024 * 1024)
			lbvsvrServiceAvgTTFB.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.AvgTimeToFirstByte) * 0.001)
			lbvsvrServiceState.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(svc.State.Value())
			lbvsvrServiceTotalRequests.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.TotalRequests))
			lbvsvrServiceTotalResponses.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.TotalResponses))
			lbvsvrServiceTotalRequestBytes.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.TotalRequestBytes))
			lbvsvrServiceTotalResponseBytes.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.TotalResponseBytes))
			lbvsvrServiceCurrentClientConns.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.CurrentClientConnections))
			lbvsvrServiceCurrentServerConns.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.CurrentServerConnections))
			lbvsvrServiceSurgeCount.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.SurgeCount))
			lbvsvrServiceServerEstablishedConnections.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.ServerEstablishedConnections))
			lbvsvrServiceCurrentReusePool.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.CurrentReusePool))
			lbvsvrServiceMaxClients.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.MaxClients))
			lbvsvrServiceCurrentLoad.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.CurrentLoad))
			lbvsvrServiceVirtualServerServiceHits.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.ServiceHits))
			lbvsvrServiceActiveTransactions.WithLabelValues(P.nsInstance, svc.Name, ss.Name, svc.ServiceType).Set(cast.ToFloat64(svc.ActiveTransactions))
		}
	case ServiceStats:
		svcNames := P.vipMap.getMappings(P.nsInstance, ss.Name, P.logger)
		switch len(svcNames) {
		case 0:
			P.logger.Debug("no bindings found for service", zap.String("service", ss.Name))
			/* Needs work. Leftover services needing cleaned up causes repeated updates. Use interval or handler for now.
			if P.collectMappings {
				go collectMappings(P, true, nil)
			}
			*/
		default:
			for _, svcName := range svcNames {
				lbvsvrServiceThroughput.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.Throughput) * 1024 * 1024)
				lbvsvrServiceAvgTTFB.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.AvgTimeToFirstByte) * 0.001)
				lbvsvrServiceState.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(ss.State.Value())
				lbvsvrServiceTotalRequests.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.TotalRequests))
				lbvsvrServiceTotalResponses.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.TotalResponses))
				lbvsvrServiceTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.TotalRequestBytes))
				lbvsvrServiceTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.TotalResponseBytes))
				lbvsvrServiceCurrentClientConns.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.CurrentClientConnections))
				lbvsvrServiceCurrentServerConns.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.CurrentServerConnections))
				lbvsvrServiceSurgeCount.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.SurgeCount))
				lbvsvrServiceServerEstablishedConnections.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.ServerEstablishedConnections))
				lbvsvrServiceCurrentReusePool.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.CurrentReusePool))
				lbvsvrServiceMaxClients.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.MaxClients))
				lbvsvrServiceCurrentLoad.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.CurrentLoad))
				lbvsvrServiceVirtualServerServiceHits.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.ServiceHits))
				lbvsvrServiceActiveTransactions.WithLabelValues(P.nsInstance, ss.Name, svcName, ss.ServiceType).Set(cast.ToFloat64(ss.ActiveTransactions))
			}
		}
	}
}
