package main

import (
	"sync"
	"time"
)

var (
	poolWG              sync.WaitGroup
	poolStop            chan struct{}
	poolLock            *sync.Mutex
	flipBit             collectBit
	collectIntervalSecs = 5
)

// PoolCollection is collection of Pool.
type PoolCollection []*Pool

type collectBit bool

func (c collectBit) good() bool {
	var ok bool
	poolLock.Lock()
	if !c {
		c = true
		ok = true
	}
	poolLock.Unlock()
	return ok
}

func (c collectBit) flip() {
	poolLock.Lock()
	if c {
		c = false
	} else {
		c = true
	}
	poolLock.Unlock()
}

func (p PoolCollection) startCollecting() {
	p.startTeams()
	poolStop = make(chan struct{})
	poolLock = &sync.Mutex{}
	poolWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(collectIntervalSecs) * time.Second)
	collectLoop:
		for {
			select {
			case <-poolStop:
				break collectLoop
			case <-ticker.C:
				poolWG.Add(1)
				go p.processAll(&poolWG)
			}
		}
	}(&poolWG)
}

func (p PoolCollection) stopCollecting() {
	close(poolStop)
	for _, P := range p {
		P.stopped = true
	}
	poolWG.Wait()
	p.stopTeams()
}

func (p PoolCollection) startTeams() {
	for _, P := range p {
		P.team.Start()
	}
}

func (p PoolCollection) stopTeams() {
	for _, P := range p {
		P.team.Stop()
		P.closeClientPool()
	}
}

func (p PoolCollection) processAll(wg *sync.WaitGroup) {
	defer wg.Done()
	w := sync.WaitGroup{}
	for _, P := range p {
		w.Add(1)
		go P.collectMetrics(&w)
	}
	w.Wait()
}

func (p PoolCollection) processSvcStats(wg *sync.WaitGroup) {
	switch {
	case wg != nil:
		defer wg.Done()
		w := sync.WaitGroup{}
		for _, P := range p {
			w.Add(1)
			go processSvcStats(P, &w)
		}
		w.Wait()
	default:
		for _, P := range p {
			go processSvcStats(P, nil)
		}
	}
}

func (p PoolCollection) collectMappings(wg *sync.WaitGroup) {
	switch {
	case wg != nil:
		defer wg.Done()
		w := sync.WaitGroup{}
		for _, P := range p {
			w.Add(1)
			go collectMappings(P, &w)
		}
		w.Wait()
	default:
		for _, P := range p {
			go collectMappings(P, nil)
		}
	}
}
