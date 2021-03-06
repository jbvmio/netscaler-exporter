package main

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	backoffTime = 3
)

func (p *Pool) collectMetrics(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	switch {
	case p.stopped:
		p.logger.Info("unable to collect metrics, process is stopping")
	case p.poolFlipBit.good():
		defer p.poolFlipBit.flip()
		for s, f := range p.metricHandlers {
			switch {
			case p.hasBackoff(s):
				go handleBackoff(p, s)
			default:
				switch s {
				case lbvserviceSubsystem:
					go f(p, nil)
				default:
					f(p, nil)
				}
			}
		}
	default:
		p.logger.Debug(("metric collection already in progress"))
	}
}

func (p *Pool) insertBackoff(subSystem string) {
	p.backoff.Update(subSystem, &FlipBit{lock: sync.Mutex{}})
}

func (p *Pool) removeBackoff(subSystem string) {
	p.backoff.Remove(subSystem)
}

func (p *Pool) hasBackoff(subSystem string) bool {
	return p.backoff.Exists(subSystem)
}

func handleBackoff(p *Pool, subSystem string) {
	fb, ok := p.backoff.Get(subSystem).(*FlipBit)
	switch {
	case !ok:
		p.logger.Error("recieved nil backoff", zap.String("subSystem", subSystem))
	case fb.good():
		defer p.backoff.Remove(subSystem)
		p.logger.Info("performing backoff for subSystem metric collection", zap.String("subSystem", subSystem))
		time.Sleep(time.Second * backoffTime)
	}
}
