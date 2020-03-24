package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

const sslSubsystem = `ssl`

var (
	sslLabels            = []string{netscalerInstance}
	sslTotalTransactions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: sslSubsystem,
			Name:      "transactions_total",
			Help:      "number of SSL transactions on the netscaler appliance",
		},
		sslLabels,
	)
	sslTotalSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: sslSubsystem,
			Name:      "sessions_total",
			Help:      "number of SSL sessions on the netscaler appliance",
		},
		sslLabels,
	)
	sslCurrentSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: sslSubsystem,
			Name:      "sessions",
			Help:      "number of active SSL sessions on the netscaler appliance",
		},
		sslLabels,
	)
)

func (P *Pool) promSSLStats(ss SSLStats) {
	sslTotalTransactions.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalSSLTransactions))
	sslTotalSessions.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.TotalSSLSessions))
	sslCurrentSessions.WithLabelValues(P.nsInstance).Set(cast.ToFloat64(ss.SSLSessions))
}
