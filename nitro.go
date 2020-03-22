package main

import (
	"sync"
)

// TaskID defines the differents tasks available working with the Nitro API.
type TaskID int

const (
	nitroTaskAPI TaskID = iota
	nitroTaskRaw
	nitroTaskData
	nitroProm
)

var nitroTaskStrings = [...]string{
	`nitroTaskAPI`,
	`nitroTaskRaw`,
	`nitroTaskData`,
	`nitroProm`,
}

// ID returns the int ID.
func (t TaskID) ID() int {
	return int(t)
}

// String returns the definition of the TaskID.
func (t TaskID) String() string {
	return nitroTaskStrings[t]
}

// NitroData identifies the type of resource or stat from the Nitro API.
type NitroData interface {
	NitroType() string
}

// NitroRaw represents raw data returned by the Nitro API.
type NitroRaw interface {
	Len() int
}

// RawData is any payload as returned by the Nitro API.
type RawData []byte

// Len returns the size of the underlying []byte.
func (r RawData) Len() int {
	return len(r)
}

type metricHandleFunc func(*Pool, *sync.WaitGroup)

func defaultMetricHandleFunc(P *Pool, wg *sync.WaitGroup) {
	wg.Done()
}

var metricsMap = map[string]metricHandleFunc{
	`service`: processSvcStats,
}

func (p *Pool) collectMetrics(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	switch {
	case p.stopped:
		p.logger.Info("unable to collect metrics, process is stopping")
	case !p.mappingsLoaded:
		p.logger.Info("unable to collect metrics, mapping not yet complete")
	case p.flipBit.good():
		defer p.flipBit.flip()
		for _, f := range p.metricHandlers {
			p.poolWG.Add(1)
			go f(p, &p.poolWG)
		}
		p.poolWG.Wait()
	default:
		p.logger.Info(("metric collection already in progress"))
	}
}
