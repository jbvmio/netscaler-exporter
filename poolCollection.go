package main

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	collectionWG    sync.WaitGroup
	collectionStop  chan struct{}
	collectionLock  *sync.Mutex
	collectInterval time.Duration
)

// PoolCollection is collection of Pool.
type PoolCollection []*Pool

func (p PoolCollection) startCollecting(l *zap.Logger) {
	logger := l.With(zap.String("process", "Pool Collector"))
	p.startTeams()
	collectionStop = make(chan struct{})
	collectionLock = &sync.Mutex{}
	collectionWG.Add(1)
	go func(wg *sync.WaitGroup, logger *zap.Logger) {
		defer wg.Done()
		logger.Info("Starting Metric Collection")
		ticker := time.NewTicker(collectInterval)
	collectLoop:
		for {
			select {
			case <-collectionStop:
				logger.Warn("Stopping Metric Collection")
				break collectLoop
			case <-ticker.C:
				collectionWG.Add(1)
				go p.processAll(&collectionWG, logger)
			}
		}
		logger.Warn("Metric Collection Stopped")
		p.stopTeams()
	}(&collectionWG, logger)
}

func (p PoolCollection) stopCollecting() {
	close(collectionStop)
}

func (p PoolCollection) startTeams() {
	wg := sync.WaitGroup{}
	for _, P := range p {
		wg.Add(1)
		go P.startTeam(&wg)
	}
	wg.Wait()
}

func (p PoolCollection) stopTeams() {
	wg := sync.WaitGroup{}
	for _, P := range p {
		P.stopped = true
		wg.Add(1)
		go P.stopTeam(&wg)
	}
	wg.Wait()
	for _, P := range p {
		wg.Add(1)
		go P.closeClientPool(&wg)
	}
	wg.Wait()
}

func (p PoolCollection) processAll(wg *sync.WaitGroup, l *zap.Logger) {
	defer wg.Done()
	w := sync.WaitGroup{}
	for _, P := range p {
		w.Add(1)
		l.Debug("Start Collect", zap.String("nsInstance", P.nsInstance))
		go P.collectMetrics(&w)
	}
	w.Wait()
}

func (p PoolCollection) collectMappings(wg *sync.WaitGroup) {
	switch {
	case wg != nil:
		defer wg.Done()
		w := sync.WaitGroup{}
		for _, P := range p {
			if P.collectMappings {
				w.Add(1)
				go collectMappings(P, false, &w)
			}
		}
		w.Wait()
	default:
		for _, P := range p {
			if P.collectMappings {
				go collectMappings(P, false, nil)
			}
		}
	}
}
