package main

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

const (
	exporterSubsystem = "exporter"
	promResetVal      = 0
)

// TK keeps track of latest collect times for each nsInstance and subSystem.
var TK = &timekeeper{
	last: make(map[string]map[string]int64),
	lock: sync.Mutex{},
}

var (
	exporterLabels        = []string{netscalerInstance, `citrixadc_subsystem`}
	exporterFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `failures_total`,
			Help:      `The total number of failures encountered by the netscaler-exporter`,
		},
		exporterLabels,
	)
	exporterScrapeLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `scrape_lag`,
			Help:      `The number of milliseconds between the subsystem metric collection and the metric scrape`,
		},
		exporterLabels,
	)
	exporterScrapeLagDesc = prometheus.NewDesc(
		namespace+`_`+exporterSubsystem+`_scrape_lag`,
		`The number of milliseconds between the subsystem metric collection and the metric scrape`,
		exporterLabels,
		nil,
	)
)

type exporter struct {
	scrapeLagDesc *prometheus.Desc
}

// Describe implements prometheus.Collector.
func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.scrapeLagDesc
}

// Collect implements prometheus.Collector.
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	timeNow := time.Now().Unix() * 1000
	times := TK.retrieve()
	for ins, sub := range times {
		for s, T := range sub {
			if T > 0 {
				ch <- prometheus.MustNewConstMetric(e.scrapeLagDesc, prometheus.GaugeValue, cast.ToFloat64(timeNow-T), ins, s)
			}
		}
	}
}

func newExporter() *exporter {
	return &exporter{
		scrapeLagDesc: exporterScrapeLagDesc,
	}
}

type timekeeper struct {
	last map[string]map[string]int64
	lock sync.Mutex
}

func (t *timekeeper) set(instance, subSystem string, T int64) {
	t.lock.Lock()
	_, ok := t.last[instance]
	if !ok {
		t.last[instance] = make(map[string]int64)
	}
	t.last[instance][subSystem] = T
	t.lock.Unlock()
}

func (t *timekeeper) get(instance, subSystem string) int64 {
	t.lock.Lock()
	_, ok := t.last[instance]
	if !ok {
		return 0
	}
	_, ok = t.last[instance][subSystem]
	if !ok {
		return 0
	}
	T := t.last[instance][subSystem]
	t.lock.Unlock()
	return T
}

func (t *timekeeper) retrieve() map[string]map[string]int64 {
	tmp := make(map[string]map[string]int64)
	t.lock.Lock()
	for ins, sub := range t.last {
		tmp[ins] = make(map[string]int64)
		for s, T := range sub {
			tmp[ins][s] = T
		}
	}
	t.lock.Unlock()
	return tmp
}
