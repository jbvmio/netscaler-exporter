package main

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	ioprom "github.com/prometheus/client_model/go"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const (
	exporterSubsystem = "exporter"
	nanoSecond        = 1000000000
	promResetVal      = 0
)

// TK keeps track of latest collect times for each nsInstance and subSystem.
var TK = &timekeeper{
	last: make(map[string]map[string]float64),
	lock: sync.Mutex{},
}

var (
	exporterLabels             = []string{netscalerInstance, `citrixadc_subsystem`}
	nsInfoLabels               = []string{netscalerInstance, `citrixadc_ns_model`, `citrixadc_ns_version`, `citrixadc_ns_year`}
	exporterAPICollectFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `api_collect_failures_total`,
			Help:      `The total number of failures encountered while querying the netscaler API`,
		},
		exporterLabels,
	)
	exporterProcessingFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `processing_failures_total`,
			Help:      `The total number of failures encountered while processing data returned from the netscaler API`,
		},
		exporterLabels,
	)
	exporterMissedMetrics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `missed_metrics_total`,
			Help:      `The total number of metrics that were missed and not able to get collected`,
		},
		exporterLabels,
	)
	exporterPromCollectFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `prometheus_collect_failures_total`,
			Help:      `The total number of failures encountered sending metrics to prometheus`,
		},
		exporterLabels,
	)
	exporterPromProcessingTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `processing_time_seconds`,
			Help:      `Duration in seconds gathering subsystem metrics from the last collection if successful`,
		},
		exporterLabels,
	)
	exporterScrapeLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporterSubsystem,
			Name:      `scrape_lag_seconds`,
			Help:      `The number of seconds between the subsystem metric collection and the metric scrape`,
		},
		exporterLabels,
	)
	exporterScrapeLagDesc = prometheus.NewDesc(
		namespace+`_`+exporterSubsystem+`_scrape_lag_seconds`,
		`The number of seconds between the subsystem metric collection and the metric scrape`,
		exporterLabels,
		nil,
	)
	exporterNSYear = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: `ns`,
			Name:      `info`,
			Help:      `A metric with a constant '1' value labeled with model, manufactured year, and version of the Netscaler.`,
		},
		nsInfoLabels,
	)
	exporterNSYearDesc = prometheus.NewDesc(
		namespace+`_ns_info`,
		`A metric with a constant '1' value labeled with model, manufactured year, and version of the Netscaler.`,
		nsInfoLabels,
		nil,
	)
)

type exporter struct {
	counterRegistry *prometheus.Registry
	scrapeLagDesc   *prometheus.Desc
	nsYearDesc      *prometheus.Desc
	lock            sync.Mutex
	logger          *zap.Logger
}

// Describe implements prometheus.Collector.
func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.scrapeLagDesc
	ch <- e.nsYearDesc
}

// Collect implements prometheus.Collector.
/*
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go pools.collectNSYear(e.nsYearDesc, ch, &wg)
	wg.Add(1)
	go e.collectCounters(ch, &wg)
	wg.Add(1)
	timeNow := float64(time.Now().UnixNano())
	times := TK.retrieve()
	for ins, sub := range times {
		for s, T := range sub {
			if T > 0 {
				ch <- prometheus.MustNewConstMetric(e.scrapeLagDesc, prometheus.GaugeValue, (timeNow-T)/nanoSecond, ins, s)
			}
		}
	}
	wg.Wait()
}
*/

// Collect implements prometheus.Collector.
func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	pools.collectNSYear(e.nsYearDesc, ch, nil)
	e.collectCounters(ch, nil)
	timeNow := float64(time.Now().UnixNano())
	times := TK.retrieve()
	for ins, sub := range times {
		for s, T := range sub {
			if T > 0 {
				ch <- prometheus.MustNewConstMetric(e.scrapeLagDesc, prometheus.GaugeValue, (timeNow-T)/nanoSecond, ins, s)
			}
		}
	}
}

func (e *exporter) collectCounters(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	fams, err := e.getCounterFamilies()
	if err != nil {
		exporterProcessingFailures.WithLabelValues(`all`, `exporter`).Inc()
		e.logger.Error("error gathering counters", zap.Error(err))
		return
	}
	for _, fam := range fams {
		metrics := fam.GetMetric()
		if len(metrics) > 0 {
			var labels []string
			for _, l := range metrics[0].GetLabel() {
				labels = append(labels, l.GetName())
			}
			desc := prometheus.NewDesc(fam.GetName(), fam.GetHelp(), labels, nil)
			for _, metric := range metrics {
				var labelVals []string
				for _, l := range metric.GetLabel() {
					labelVals = append(labelVals, l.GetValue())
				}
				ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, metric.GetGauge().GetValue(), labelVals...)
			}
		}
	}
}

func (e *exporter) getCounterFamilies() (fams []*ioprom.MetricFamily, err error) {
	e.lock.Lock()
	fams, err = e.counterRegistry.Gather()
	e.lock.Unlock()
	return
}

func (p PoolCollection) collectNSYear(desc *prometheus.Desc, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	for _, P := range p {
		if P.nsVersion != "" {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1, P.nsInstance, P.nsModel, P.nsVersion, cast.ToString(P.nsYear))
		}
	}
}

func newExporter(cr *prometheus.Registry, l *zap.Logger) *exporter {
	return &exporter{
		counterRegistry: cr,
		scrapeLagDesc:   exporterScrapeLagDesc,
		nsYearDesc:      exporterNSYearDesc,
		lock:            sync.Mutex{},
		logger:          l.With(zap.String("process", "exporter")),
	}
}

type timekeeper struct {
	last map[string]map[string]float64
	lock sync.Mutex
}

func (t *timekeeper) set(instance, subSystem string, T float64) {
	t.lock.Lock()
	_, ok := t.last[instance]
	if !ok {
		t.last[instance] = make(map[string]float64)
	}
	t.last[instance][subSystem] = T
	t.lock.Unlock()
}

func (t *timekeeper) get(instance, subSystem string) float64 {
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

func (t *timekeeper) retrieve() map[string]map[string]float64 {
	tmp := make(map[string]map[string]float64)
	t.lock.Lock()
	for ins, sub := range t.last {
		tmp[ins] = make(map[string]float64)
		for s, T := range sub {
			tmp[ins][s] = T
		}
	}
	t.lock.Unlock()
	return tmp
}
