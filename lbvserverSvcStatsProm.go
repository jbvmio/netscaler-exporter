package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const lbvserviceSubsystem = `lbservice`

var lbvserverSvcSubsystem = `service`
var (
	lbvsvrServiceLabels     = []string{netscalerInstance, `citrixadc_service_name`, `citrixadc_lb_name`, `citrixadc_service_type`}
	lbvsvrServiceThroughput = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "throughput_bytes",
			Help:      "Number of bytes received or sent by this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceAvgTTFB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "average_time_to_first_byte_seconds",
			Help:      "Average TTFB between the NetScaler appliance and the server. TTFB is the time interval between sending the request packet to a service and receiving the first response from the service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "state",
			Help:      "Current state of the service. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceTotalRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "requests_total",
			Help:      "Total number of requests received on this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceTotalResponses = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "responses_total",
			Help:      "Total number of responses received on this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceCurrentClientConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "current_client_connections",
			Help:      "Number of current client connections",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceSurgeCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "surge_queue",
			Help:      "Number of requests in the surge queue",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceCurrentServerConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "current_server_connections",
			Help:      "Number of current connections to the actual servers",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceServerEstablishedConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "server_established_connections",
			Help:      "Number of server connections in ESTABLISHED state",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceCurrentReusePool = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "current_reuse_pool",
			Help:      "Number of requests in the idle queue/reuse pool.",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceMaxClients = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "max_clients",
			Help:      "Maximum open connections allowed on this service",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceCurrentLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "current_load",
			Help:      "Load on the service that is calculated from the bound load based monitor",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceVirtualServerServiceHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "vserver_service_hits",
			Help:      "Number of times that the service has been provided",
		},
		lbvsvrServiceLabels,
	)

	lbvsvrServiceActiveTransactions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverSvcSubsystem,
			Name:      "active_transactions",
			Help:      "Number of active transactions handled by this service. (Including those in the surge queue.) Active Transaction means number of transactions currently served by the server including those waiting in the SurgeQ",
		},
		lbvsvrServiceLabels,
	)
)
