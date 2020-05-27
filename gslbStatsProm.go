package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/gslb/gslbvserver/
// https://developer-docs.citrix.com/projects/netscaler-nitro-api/en/12.0/statistics/gslb/gslbservice/

const (
	gslbVServerSubsystem    = `gslb_vserver`
	gslbVServerSvcSubsystem = `gslb_service`
)

var (
	gslbVServerLabels           = []string{netscalerInstance, `citrixadc_gslb_name`, `citrixadc_gslb_type`}
	gslbVServerEstablishedConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "established_connections",
			Help:      "Number of client connections in ESTABLISHED state",
		},
		gslbVServerLabels,
	)

	gslbVServerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "state",
			Help:      "Current state of the server. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		gslbVServerLabels,
	)

	gslbVServerHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "health",
			Help:      "Health of the vserver. Gives percentage of UP services bound to this vserver",
		},
		gslbVServerLabels,
	)

	gslbVServerActiveServices = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "active_services",
			Help:      "Number of ACTIVE services bound to a vserver",
		},
		gslbVServerLabels,
	)

	gslbVServerTotalHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "hits_total",
			Help:      "Total vserver hits",
		},
		gslbVServerLabels,
	)

	gslbVServerTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this virtual server",
		},
		gslbVServerLabels,
	)

	gslbVServerTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this virtual server",
		},
		gslbVServerLabels,
	)
)

var (
	gslbVServerSvcLabels = []string{netscalerInstance, `citrixadc_gslb_name`, `citrixadc_gslb_service_name`, `citrixadc_gslb_service_type`}

	gslbServicesEstablishedConns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "established_connections",
			Help:      "Number of client connections in ESTABLISHED state",
		},
		gslbVServerSvcLabels,
	)

	gslbServicesState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "state",
			Help:      "Current state of the service. 0 = DOWN, 1 = UP, 2 = OUT OF SERVICE, 3 = UNKNOWN",
		},
		gslbVServerSvcLabels,
	)

	gslbServicesTotalRequestBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "request_bytes_total",
			Help:      "Total number of request bytes received on this service",
		},
		gslbVServerSvcLabels,
	)

	gslbServicesTotalResponseBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "response_bytes_total",
			Help:      "Total number of response bytes received on this service",
		},
		gslbVServerSvcLabels,
	)

	gslbServicesHits = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: gslbVServerSvcSubsystem,
			Name:      "hits_total",
			Help:      "Number of times that the service has been provided",
		},
		gslbVServerSvcLabels,
	)
)

func (P *Pool) promGSLBVServerStats(ss GSLBVServerStats) {
	gslbVServerEstablishedConns.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.EstablishedConnections))
	gslbVServerState.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.State.Value()))
	gslbVServerHealth.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.Health))
	gslbVServerActiveServices.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.ActiveServices))
	gslbVServerTotalHits.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalHits))
	gslbVServerTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalRequestBytes))
	gslbVServerTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, ss.Type).Set(cast.ToFloat64(ss.TotalResponseBytes))
	P.labelTTLs.setTTL(gslbvserverStatCollection, P.nsInstance, ss.Name, ss.Type)
	for _, svc := range ss.GSLBService {
		gslbServicesEstablishedConns.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType).Set(cast.ToFloat64(svc.EstablishedConnections))
		gslbServicesHits.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType).Set(cast.ToFloat64(svc.ServiceHits))
		gslbServicesState.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType).Set(cast.ToFloat64(svc.State.Value()))
		gslbServicesTotalRequestBytes.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType).Set(cast.ToFloat64(svc.TotalRequestBytes))
		gslbServicesTotalResponseBytes.WithLabelValues(P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType).Set(cast.ToFloat64(svc.TotalResponseBytes))
		P.labelTTLs.setTTL(gslbServiceCollection, P.nsInstance, ss.Name, svc.ServiceName, svc.ServiceType)
	}
}

var gslbvserverStatCollection = gaugeCollection{
	gslbVServerEstablishedConns,
	gslbVServerState,
	gslbVServerHealth,
	gslbVServerActiveServices,
	gslbVServerTotalHits,
	gslbVServerTotalRequestBytes,
	gslbVServerTotalResponseBytes,
}

var gslbServiceCollection = gaugeCollection{
	gslbServicesEstablishedConns,
	gslbServicesHits,
	gslbServicesState,
	gslbServicesTotalRequestBytes,
	gslbServicesTotalResponseBytes,
}
