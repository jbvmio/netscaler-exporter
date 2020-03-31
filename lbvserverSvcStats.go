package main

// RawSSFromLBVS is a RawLBVServer payload as returned by the Nitro API used to process ServiceStats.
type RawSSFromLBVS []byte

// Len returns the size of the underlying []byte.
func (r RawSSFromLBVS) Len() int {
	return len(r)
}

/*
func processLBVServerSvcStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	thisSS := lbvserverSvcSubsystem
	switch {
	case P.stopped:
		P.logger.Info("Skipping sybSystem stat collection, process is stopping", zap.String("subSystem", thisSS))
	default:
		P.logger.Debug("Processing subSystem Stats", zap.String("subSystem", thisSS))
		data := submitAPITask(P, netscaler.StatsTypeLBVServer)
		switch {
		case len(data) < 1:
			P.logger.Error("error retrieving data for subSystem stat collection", zap.String("subSystem", thisSS))
			exporterAPICollectFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
			P.insertBackoff(thisSS)
		case P.metricFlipBit[thisSS].good():
			defer P.metricFlipBit[thisSS].flip()
			req := newNitroRawReq(RawSSFromLBVS(data))
			P.submit(req)
			s := <-req.ResultChan()
			if success, ok := s.(bool); ok {
				switch {
				case success:
					go TK.set(P.nsInstance, thisSS, float64(time.Now().UnixNano()))
				default:
					exporterProcessingFailures.WithLabelValues(P.nsInstance, thisSS).Inc()
					P.insertBackoff(thisSS)
				}
			}
			P.logger.Debug("subSystem stat collection Complete", zap.String("subSystem", thisSS))
		default:
			P.logger.Debug("subSystem metric collection already in progress", zap.String("subSystem", thisSS))

		}
	}
}
*/
