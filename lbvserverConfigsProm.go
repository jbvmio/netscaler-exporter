package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

const lbvserverConfigSubsystem = `lbvserver_cfg`

var (
	lbvserverConfigLabels        = []string{netscalerInstance, `citrixadc_lb_name`}
	lbvserverLastStateChangeSecs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: lbvserverConfigSubsystem,
			Name:      "last_state_change_seconds",
			Help:      "Seconds since the last state change.",
		},
		lbvserverConfigLabels,
	)
)

func (P *Pool) promLBVServerConfigs(ss LBVServerConfigs) {
	lbvserverLastStateChangeSecs.WithLabelValues(P.nsInstance, ss.Name).Set(cast.ToFloat64(ss.StateChangeTimeSeconds))
	P.labelTTLs.setTTL(lbvserverConfigCollection, P.nsInstance, ss.Name)
}

var lbvserverConfigCollection = gaugeCollection{
	lbvserverLastStateChangeSecs,
}
