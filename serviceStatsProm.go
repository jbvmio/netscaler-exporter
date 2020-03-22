package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const servicesSubsystem = "service"

var (
	servicesLabels = []string{netscalerInstance, `citrixadc_service_name`, `citrixadc_lb_name`}
	// TODO - Convert megabytes to bytes
	servicesThroughput = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "throughput_bytes_total",
			Help:      "Number of bytes received or sent by this service",
		},
		servicesLabels,
	)

	servicesAvgTTFB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "average_time_to_first_byte_seconds",
			Help:      "Average TTFB between the NetScaler appliance and the server. TTFB is the time interval between sending the request packet to a service and receiving the first response from the service",
		},
		servicesLabels,
	)

	servicesState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "state",
			Help:      "Current state of the service. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		servicesLabels,
	)

	servicesTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "requests_total",
			Help:      "Total number of requests received on this service",
		},
		servicesLabels,
	)

	servicesTotalResponses = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "responses_total",
			Help:      "Total number of responses received on this service",
		},
		servicesLabels,
	)

	servicesTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this service",
		},
		servicesLabels,
	)

	servicesTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this service",
		},
		servicesLabels,
	)

	servicesCurrentClientConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "current_client_connections",
			Help:      "Number of current client connections",
		},
		servicesLabels,
	)

	servicesSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "surge_queue",
			Help:      "Number of requests in the surge queue",
		},
		servicesLabels,
	)

	servicesCurrentServerConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "current_server_connections",
			Help:      "Number of current connections to the actual servers",
		},
		servicesLabels,
	)

	servicesServerEstablishedConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "server_established_connections",
			Help:      "Number of server connections in ESTABLISHED state",
		},
		servicesLabels,
	)

	servicesCurrentReusePool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "current_reuse_pool",
			Help:      "Number of requests in the idle queue/reuse pool.",
		},
		servicesLabels,
	)

	servicesMaxClients = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "max_clients",
			Help:      "Maximum open connections allowed on this service",
		},
		servicesLabels,
	)

	servicesCurrentLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "current_load",
			Help:      "Load on the service that is calculated from the bound load based monitor",
		},
		servicesLabels,
	)

	servicesVirtualServerServiceHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "vserver_service_hits_total",
			Help:      "Number of times that the service has been provided",
		},
		servicesLabels,
	)

	servicesActiveTransactions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: servicesSubsystem,
			Name:      "active_transactions",
			Help:      "Number of active transactions handled by this service. (Including those in the surge queue.) Active Transaction means number of transactions currently served by the server including those waiting in the SurgeQ",
		},
		servicesLabels,
	)
)

func (P *Pool) promSvcStats(ss ServiceStats) {
	P.logger.Debug("recieved ServiceStat", zap.String("Received", fmt.Sprintf("%+v", ss)))
	servicesThroughput.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Add(cast.ToFloat64(ss.Throughput))
	servicesAvgTTFB.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.AvgTimeToFirstByte))
	servicesState.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(ss.State.Value())
	servicesTotalRequests.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalRequests))
	servicesTotalResponses.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalResponses))
	servicesTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalRequestBytes))
	servicesTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.TotalResponseBytes))
	servicesCurrentClientConns.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentClientConnections))
	servicesCurrentServerConns.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentServerConnections))
	servicesSurgeCount.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.SurgeCount))
	servicesServerEstablishedConnections.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ServerEstablishedConnections))
	servicesCurrentReusePool.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentReusePool))
	servicesMaxClients.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.MaxClients))
	servicesCurrentLoad.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.CurrentLoad))
	servicesVirtualServerServiceHits.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ServiceHits))
	servicesActiveTransactions.WithLabelValues(P.nsInstance, ss.Name, ss.ServiceName).Set(cast.ToFloat64(ss.ActiveTransactions))
}

// External Counters:
// https://github.com/prometheus-net/prometheus-net/issues/63
