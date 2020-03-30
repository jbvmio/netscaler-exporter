package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/load-balancing/lbvserver/lbvserver/

const lbvserverSubsystem = `lbvserver`

var (
	lbvserverLabels    = []string{netscalerInstance, `citrixadc_lb_name`}
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

func (P *Pool) promLBVServerStats(ss LBVServerStats) {
	lbvserverAveCLTTLB.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.AvgTimeClientTTLB))
	lbvserverState.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.State.Value))
	lbvserverTotalRequests.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalRequests))
	lbvserverTotalResponses.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalResponses))
	lbvserverTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalRequestBytes))
	lbvserverTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalResponseBytes))
	lbvserverTotalClientTTLBTrans.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalClientTTLBTransactions))
	lbvserverActiveServices.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.ActiveServices))
	lbvserverTotalHits.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalHits))
	lbvserverTotalPktsRx.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalPktsReceived))
	lbvserverTotalPktsTx.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalPktsSent))
	lbvserverSurgeCount.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.SurgeCount))
	lbvserverSvcSurgeCount.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.SvcSurgeCount))
	lbvserverVSvrSurgeCount.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.VSvrSurgeCount))
}
