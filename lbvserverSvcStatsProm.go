package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const lbVSvrSvcSubsystem = `lbvsvrsvc`

var (
	lbVSvrSvcLabels     = []string{netscalerInstance, `citrixadc_service_name`, `citrixadc_lb_name`}
	lbVSvrSvcThroughput = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "throughput_bytes",
			Help:      "Number of bytes received or sent by this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcAvgTTFB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "average_time_to_first_byte_seconds",
			Help:      "Average TTFB between the NetScaler appliance and the server. TTFB is the time interval between sending the request packet to a service and receiving the first response from the service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "state",
			Help:      "Current state of the service. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "requests_total",
			Help:      "Total number of requests received on this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcTotalResponses = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "responses_total",
			Help:      "Total number of responses received on this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcCurrentClientConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "current_client_connections",
			Help:      "Number of current client connections",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "surge_queue",
			Help:      "Number of requests in the surge queue",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcCurrentServerConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "current_server_connections",
			Help:      "Number of current connections to the actual servers",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcServerEstablishedConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "server_established_connections",
			Help:      "Number of server connections in ESTABLISHED state",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcCurrentReusePool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "current_reuse_pool",
			Help:      "Number of requests in the idle queue/reuse pool.",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcMaxClients = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "max_clients",
			Help:      "Maximum open connections allowed on this service",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcCurrentLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "current_load",
			Help:      "Load on the service that is calculated from the bound load based monitor",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcVirtualServerServiceHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "vserver_service_hits",
			Help:      "Number of times that the service has been provided",
		},
		lbVSvrSvcLabels,
	)

	lbVSvrSvcActiveTransactions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbVSvrSvcSubsystem,
			Name:      "active_transactions",
			Help:      "Number of active transactions handled by this service. (Including those in the surge queue.) Active Transaction means number of transactions currently served by the server including those waiting in the SurgeQ",
		},
		lbVSvrSvcLabels,
	)
)

func (P *Pool) promLBVSvrSvcStats(ss ServiceStats) {
	P.logger.Debug("recieved ServiceStat", zap.String("Received", fmt.Sprintf("%+v", ss)))
	// Value is in megabytes. Convert to base unit of bytes.
	lbVSvrSvcThroughput.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.Throughput) * 1024 * 1024)
	// Value is in milliseconds. Convert to base unit of seconds.
	lbVSvrSvcAvgTTFB.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.AvgTimeToFirstByte) * 0.001)
	lbVSvrSvcState.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(ss.State.Value())
	lbVSvrSvcTotalRequests.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalRequests))
	lbVSvrSvcTotalResponses.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalResponses))
	lbVSvrSvcTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalRequestBytes))
	lbVSvrSvcTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalResponseBytes))
	lbVSvrSvcCurrentClientConns.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentClientConnections))
	lbVSvrSvcCurrentServerConns.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentServerConnections))
	lbVSvrSvcSurgeCount.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.SurgeCount))
	lbVSvrSvcServerEstablishedConnections.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ServerEstablishedConnections))
	lbVSvrSvcCurrentReusePool.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentReusePool))
	lbVSvrSvcMaxClients.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.MaxClients))
	lbVSvrSvcCurrentLoad.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentLoad))
	lbVSvrSvcVirtualServerServiceHits.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ServiceHits))
	lbVSvrSvcActiveTransactions.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ActiveTransactions))
}
