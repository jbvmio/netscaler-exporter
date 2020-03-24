package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

const nsSubsystem = `ns`

var (
	nsLabels      = []string{netscalerInstance}
	nsCPUUsagePCT = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "cpu_usage_pct",
			Help:      "cpu utilization percentage",
		},
		nsLabels,
	)
	nsMemUsagePCT = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "mem_usage_pct",
			Help:      "percentage of memory utilization on netscaler",
		},
		nsLabels,
	)

	nsPktCPUUsagePCT = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "packet_cpu_usage_pct",
			Help:      "average cpu utilization for all packet engines excluding mgmt PE",
		},
		nsLabels,
	)

	nsFlashPartUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "flash_usage_pct",
			Help:      "used space of the /flash partition on disk as a percentage",
		},
		nsLabels,
	)

	nsVarPartUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "var_usage_pct",
			Help:      "used space of the /var partition on disk as a percentage",
		},
		nsLabels,
	)

	nsTotalRxMB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "mb_received_total",
			Help:      "number of megabytes received by the netscaler appliance",
		},
		nsLabels,
	)

	nsTotalTxMG = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "mb_transmitted_total",
			Help:      "number of megabytes transmitted by the netscaler appliance",
		},
		nsLabels,
	)

	nsHTTPReqsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "http_requests_total",
			Help:      "total number of HTTP requests received",
		},
		nsLabels,
	)

	nsHTTPRespTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "http_responses_total",
			Help:      "total number of HTTP responses sent",
		},
		nsLabels,
	)

	nsTCPCurClientConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "client_connections",
			Help:      "client connections, including connections in the opening, establised and closing state",
		},
		nsLabels,
	)

	nsTCPCurClientConnsEst = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "client_connections_established",
			Help:      "current client connections in the established state",
		},
		nsLabels,
	)

	nsTCPCurServerConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "server_connections",
			Help:      "server connections, including connections in the opening, establised and closing state",
		},
		nsLabels,
	)

	nsTCPCurServerConnsEst = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "server_connections_established",
			Help:      "current server connections in the established state",
		},
		nsLabels,
	)
)

func (P *Pool) promNSStats(ss NSStats) {
	nsCPUUsagePCT.WithLabelValues(P.nsInstance).Set(ss.CPUUsagePct)
	nsMemUsagePCT.WithLabelValues(P.nsInstance).Set(ss.MemUsagePct)
	nsPktCPUUsagePCT.WithLabelValues(P.nsInstance).Set(ss.PktCPUUsagePct)
	nsFlashPartUsage.WithLabelValues(P.nsInstance).Set(ss.FlashPartitionUsage)
	nsVarPartUsage.WithLabelValues(P.nsInstance).Set(ss.VarPartitionUsage)
	nsTotalRxMB.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalReceivedMB))
	nsTotalTxMG.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalTransmitMB))
	nsHTTPReqsTotal.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.HTTPRequests))
	nsHTTPRespTotal.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.HTTPResponses))
	nsTCPCurClientConns.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentClientConnections))
	nsTCPCurClientConnsEst.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentClientConnectionsEstablished))
	nsTCPCurServerConns.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentServerConnections))
	nsTCPCurServerConnsEst.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentServerConnectionsEstablished))
}
