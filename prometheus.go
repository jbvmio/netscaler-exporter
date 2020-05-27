package main

import (
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type metricCollection interface {
	len() int
}
type gaugeCollection []*prometheus.GaugeVec

type counterCollection []*prometheus.CounterVec

func (c gaugeCollection) len() int {
	return len(c)
}

func (c counterCollection) len() int {
	return len(c)
}

// LabelValues registers labels with a timestamp.
type LabelValues struct {
	lastRegisteredAt time.Time
	gaugeVec         *prometheus.GaugeVec
	countVec         *prometheus.CounterVec
	labels           []string
}

// LabelTTLs contains LabelValues and a TTL value.
type LabelTTLs struct {
	labelValues map[uint64]map[uint64]*LabelValues
	ttl         time.Duration
	lock        sync.Mutex
}

func (L *LabelTTLs) setTTL(c metricCollection, labels ...string) {
	L.lock.Lock()
	timeNow := time.Now()
	switch c := c.(type) {
	case gaugeCollection:
		for _, C := range c {
			xxh := xxhash.New()
			xxh.WriteString(C.WithLabelValues(labels...).Desc().String()[:100])
			metric, hasMetric := L.labelValues[xxh.Sum64()]
			if !hasMetric {
				metric = make(map[uint64]*LabelValues)
				L.labelValues[xxh.Sum64()] = metric
			}
			for _, l := range labels {
				xxh.WriteString(l)
			}
			labelVals, ok := metric[xxh.Sum64()]
			if !ok {
				labelVals = &LabelValues{
					gaugeVec: C,
					labels:   labels,
				}
				metric[xxh.Sum64()] = labelVals
			}
			labelVals.lastRegisteredAt = timeNow
		}
	case counterCollection:
		for _, C := range c {
			xxh := xxhash.New()
			xxh.WriteString(C.WithLabelValues(labels...).Desc().String()[:100])
			metric, hasMetric := L.labelValues[xxh.Sum64()]
			if !hasMetric {
				metric = make(map[uint64]*LabelValues)
				L.labelValues[xxh.Sum64()] = metric
			}
			for _, l := range labels {
				xxh.WriteString(l)
			}
			labelVals, ok := metric[xxh.Sum64()]
			if !ok {
				labelVals = &LabelValues{
					countVec: C,
					labels:   labels,
				}
				metric[xxh.Sum64()] = labelVals
			}
			labelVals.lastRegisteredAt = timeNow
		}
	}
	L.lock.Unlock()
}

func (L *LabelTTLs) deleteStale() {
	L.lock.Lock()
	timeNow := time.Now()
	for metricHash := range L.labelValues {
		for hash, labels := range L.labelValues[metricHash] {
			if labels.lastRegisteredAt.Add(L.ttl).Before(timeNow) {
				switch {
				case labels.gaugeVec != nil:
					labels.gaugeVec.DeleteLabelValues(labels.labels...)
				case labels.countVec != nil:
					labels.countVec.DeleteLabelValues(labels.labels...)
				}
				delete(L.labelValues[metricHash], hash)
			}
		}
		if len(L.labelValues[metricHash]) < 1 {
			delete(L.labelValues, metricHash)
		}
	}
	L.lock.Unlock()
}
