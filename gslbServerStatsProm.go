package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/gslb/gslbvserver/
// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/gslb/gslbservice/

const (
	gslbVServerSubsystem    = `gslbvserver`
	gslbVServerSvcSubsystem = `gslbservice`
)

var (
	gslbvserverLabels           = []string{netscalerInstance, `citrixadc_gslb_name`}
	gslbvserverEstablishedConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "established_connections",
			Help:      "Number of client connections in ESTABLISHED state",
		},
		gslbvserverLabels,
	)

	gslbvserverState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "state",
			Help:      "Current state of the server. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		gslbvserverLabels,
	)

	gslbvserverHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "health",
			Help:      "Health of the vserver. Gives percentage of UP services bound to this vserver",
		},
		gslbvserverLabels,
	)

	gslbvserverActiveServices = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "active_services",
			Help:      "Number of ACTIVE services bound to a vserver",
		},
		gslbvserverLabels,
	)

	gslbvserverTotalHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "hits_total",
			Help:      "Total vserver hits",
		},
		gslbvserverLabels,
	)

	gslbvserverTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this virtual server",
		},
		gslbvserverLabels,
	)

	gslbvserverTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this virtual server",
		},
		gslbvserverLabels,
	)
)

var (
	gslbvserverSvcLabels = []string{netscalerInstance, `citrixadc_gslb_name`, `citrixadc_gslbservice_name`}

	gslbservicesEstablishedConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "established_connections",
			Help:      "Number of client connections in ESTABLISHED state",
		},
		gslbvserverSvcLabels,
	)

	gslbservicesState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "state",
			Help:      "Current state of the service. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		gslbvserverSvcLabels,
	)

	gslbservicesTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this service",
		},
		gslbvserverSvcLabels,
	)

	gslbservicesTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this service",
		},
		gslbvserverSvcLabels,
	)

	gsservicesVirtualServerServiceHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "vserver_service_hits",
			Help:      "Number of times that the service has been provided",
		},
		gslbvserverSvcLabels,
	)
)

func (P *Pool) promGSLBVServerStats(ss GSLBVServerStats) {
	gslbvserverEstablishedConns.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.EstablishedConnections))
	gslbvserverState.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.State.Value()))
	gslbvserverHealth.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.Health))
	gslbvserverActiveServices.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.ActiveServices))
	gslbvserverTotalHits.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalHits))
	gslbvserverTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalRequestBytes))
	gslbvserverTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.TotalResponseBytes))
	for _, svc := range ss.GSLBService {
		gslbservicesState.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName).Set(cast.ToFloat64(svc.State.Value()))
		gslbservicesEstablishedConns.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName).Set(cast.ToFloat64(svc.EstablishedConnections))
		gslbservicesTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName).Set(cast.ToFloat64(svc.TotalRequestBytes))
		gslbservicesTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName).Set(cast.ToFloat64(svc.TotalResponseBytes))
		gsservicesVirtualServerServiceHits.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName).Set(cast.ToFloat64(svc.ServiceHits))
	}
}
