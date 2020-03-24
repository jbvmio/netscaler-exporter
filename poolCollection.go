package main

import (
	"sync"
	"time"
)

var (
	collectionWG        sync.WaitGroup
	collectionStop      chan struct{}
	collectionLock      *sync.Mutex
	flipBit             collectBit
	collectIntervalSecs = 5
)

// PoolCollection is collection of Pool.
type PoolCollection []*Pool

type collectBit bool

func (c collectBit) good() bool {
	var ok bool
	collectionLock.Lock()
	if !c {
		c = true
		ok = true
	}
	collectionLock.Unlock()
	return ok
}

func (c collectBit) flip() {
	collectionLock.Lock()
	if c {
		c = false
	} else {
		c = true
	}
	collectionLock.Unlock()
}

func (p PoolCollection) startCollecting() {
	p.startTeams()
	collectionStop = make(chan struct{})
	collectionLock = &sync.Mutex{}
	collectionWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(collectIntervalSecs) * time.Second)
	collectLoop:
		for {
			select {
			case <-collectionStop:
				break collectLoop
			case <-ticker.C:
				collectionWG.Add(1)
				go p.processAll(&collectionWG)
			}
		}
	}(&collectionWG)
}

func (p PoolCollection) stopCollecting() {
	close(collectionStop)
	for _, P := range p {
		P.stopped = true
	}
	collectionWG.Wait()
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
