package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/ns/ns/ns/

const nsSubsystem = `ns`

var (
	nsLabels      = []string{netscalerInstance}
	nsCPUUsagePct = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "cpu_usage_pct",
			Help:      "CPU utilization percentage",
		},
		nsLabels,
	)

	nsMemUsagePct = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "mem_usage_pct",
			Help:      "Percentage of memory utilization on NetScaler",
		},
		nsLabels,
	)

	nsMgmtCPUUsagePct = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "mgmt_cpu_usage_pct",
			Help:      "Average Management CPU utilization percentage",
		},
		nsLabels,
	)

	nsPktCPUUsagePct = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "pkt_cpu_usage_pct",
			Help:      "Average CPU utilization percentage for all packet engines excluding management PE",
		},
		nsLabels,
	)

	nsFlashPartUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "flash_usage_pct",
			Help:      "Used space in /flash partition of the disk, as a percentage. This is a critical counter",
		},
		nsLabels,
	)

	nsVarPartUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "var_usage_pct",
			Help:      "Used space in /var partition of the disk, as a percentage. This is a critical counter",
		},
		nsLabels,
	)

	nsTotalRxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "received_bytes_total",
			Help:      "Number of bytes received by the NetScaler appliance",
		},
		nsLabels,
	)

	nsTotalTxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "transmitted_bytes_total",
			Help:      "Number of bytes transmitted by the NetScaler appliance",
		},
		nsLabels,
	)

	nsHTTPReqsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests received",
		},
		nsLabels,
	)

	nsHTTPRespTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "http_responses_total",
			Help:      "Total number of HTTP responses sent",
		},
		nsLabels,
	)

	nsTCPCurClientConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "client_connections",
			Help:      "Client connections, including connections in the Opening, Established, and Closing state.",
		},
		nsLabels,
	)

	nsTCPCurClientConnsEst = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "client_connections_established",
			Help:      "Current client connections in the Established state, which indicates that data transfer can occur between the NetScaler and the client.",
		},
		nsLabels,
	)

	nsTCPCurServerConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "server_connections",
			Help:      "Server connections, including connections in the Opening, Established, and Closing state.",
		},
		nsLabels,
	)

	nsTCPCurServerConnsEst = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: nsSubsystem,
			Name:      "server_connections_established",
			Help:      "Current server connections in the Established state, which indicates that data transfer can occur between the NetScaler and the server.",
		},
		nsLabels,
	)
)

func (P *Pool) promNSStats(ss NSStats) {
	nsCPUUsagePct.WithLabelValues(P.nsInstance).Set(ss.CPUUsagePct)
	nsMemUsagePct.WithLabelValues(P.nsInstance).Set(ss.MemUsagePct)
	nsMgmtCPUUsagePct.WithLabelValues(P.nsInstance).Set(ss.MgmtCPUUsagePct)
	nsPktCPUUsagePct.WithLabelValues(P.nsInstance).Set(ss.PktCPUUsagePct)
	nsFlashPartUsage.WithLabelValues(P.nsInstance).Set(ss.FlashPartitionUsage)
	nsVarPartUsage.WithLabelValues(P.nsInstance).Set(ss.VarPartitionUsage)
	nsTotalRxBytes.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalReceivedMB) * 1024 * 1024)
	nsTotalTxBytes.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalTransmitMB) * 1024 * 1024)
	nsHTTPReqsTotal.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.HTTPRequests))
	nsHTTPRespTotal.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.HTTPResponses))
	nsTCPCurClientConns.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentClientConnections))
	nsTCPCurClientConnsEst.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentClientConnectionsEstablished))
	nsTCPCurServerConns.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentServerConnections))
	nsTCPCurServerConnsEst.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TCPCurrentServerConnectionsEstablished))
	P.labelTTLs.setTTL(nsStatCollection, P.nsInstance)
}

var nsStatCollection = gaugeCollection{
	nsCPUUsagePct,
	nsMemUsagePct,
	nsMgmtCPUUsagePct,
	nsPktCPUUsagePct,
	nsFlashPartUsage,
	nsVarPartUsage,
	nsTotalRxBytes,
	nsTotalTxBytes,
	nsHTTPReqsTotal,
	nsHTTPRespTotal,
	nsTCPCurClientConns,
	nsTCPCurClientConnsEst,
	nsTCPCurServerConns,
	nsTCPCurServerConnsEst,
}
