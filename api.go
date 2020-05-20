package main

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// API for Promethues.
type API struct {
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mappingFB   *FlipBit
	infoFB      *FlipBit
	lastMapping time.Time
	lastInfo    time.Time
	logger      *zap.Logger
}

func newAPI(L *zap.Logger) *API {
	return &API{
		stopChan:  make(chan struct{}),
		mappingFB: &FlipBit{lock: sync.Mutex{}},
		infoFB:    &FlipBit{lock: sync.Mutex{}},
		wg:        sync.WaitGroup{},
		logger:    L.With(zap.String(`process`, `Metrics Exporter API`)),
	}
}

func (a *API) start(httpSrv *http.Server) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.logger.Info("Starting ... Listening on " + httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil {
			if !strings.Contains(err.Error(), `Server closed`) {
				a.logger.Fatal("http server encountered an error", zap.Error(err))
			}
		}
		a.logger.Info("Stopped.")
	}()
}

// Stop stops the API.
func (a *API) stop(httpSrv *http.Server) {
	close(a.stopChan)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := httpSrv.Shutdown(ctx)
	if err != nil {
		a.logger.Error("error shutting down", zap.Error(err))
	}
	a.logger.Info("Stopping ...")
	<-ctx.Done()
	a.wg.Wait()
	a.logger.Info("All Processes Stopped.")
}

func makeProm(cr *prometheus.Registry, l *zap.Logger) http.Handler {
	prom := prometheus.NewRegistry()
	e := newExporter(cr, l)
	prom.MustRegister(e)
	prom.MustRegister(allPromCollectors...)
	handleProm := promhttp.HandlerFor(prom, promhttp.HandlerOpts{})
	return handleProm
}

func makeCounterRegistry() *prometheus.Registry {
	prom := prometheus.NewRegistry()
	prom.MustRegister(counterPromCollectors...)
	return prom
}

var counterPromCollectors = []prometheus.Collector{
	gslbServicesHits,
	gslbServicesTotalRequestBytes,
	gslbServicesTotalResponseBytes,
	gslbVServerTotalHits,
	gslbVServerTotalRequestBytes,
	gslbVServerTotalResponseBytes,
	lbvserverTotalRequests,
	lbvserverTotalResponses,
	lbvserverTotalRequestBytes,
	lbvserverTotalResponseBytes,
	lbvserverTotalHits,
	lbvserverTotalPktsRx,
	lbvserverTotalPktsTx,
	lbvsvrServiceTotalRequests,
	lbvsvrServiceTotalResponses,
	lbvsvrServiceTotalRequestBytes,
	lbvsvrServiceTotalResponseBytes,
	sslTotalTransactions,
	sslTotalSessions,
	nsTotalRxBytes,
	nsTotalTxBytes,
	nsHTTPReqsTotal,
	nsHTTPRespTotal,
}

var allPromCollectors = []prometheus.Collector{
	exporterAPICollectFailures,
	exporterProcessingFailures,
	exporterMissedMetrics,
	exporterPromCollectFailures,
	exporterPromProcessingTime,
	gslbServicesEstablishedConns,
	gslbServicesState,
	gslbVServerActiveServices,
	gslbVServerEstablishedConns,
	gslbVServerHealth,
	gslbVServerState,
	lbvserverLastStateChangeSecs,
	lbvserverAveCLTTLB,
	lbvserverState,
	lbvserverTotalClientTTLBTrans,
	lbvserverActiveServices,
	lbvserverSurgeCount,
	lbvserverSvcSurgeCount,
	lbvserverVSvrSurgeCount,
	lbvsvrServiceThroughput,
	lbvsvrServiceAvgTTFB,
	lbvsvrServiceState,
	lbvsvrServiceCurrentClientConns,
	lbvsvrServiceCurrentServerConns,
	lbvsvrServiceSurgeCount,
	lbvsvrServiceServerEstablishedConnections,
	lbvsvrServiceCurrentReusePool,
	lbvsvrServiceMaxClients,
	lbvsvrServiceCurrentLoad,
	lbvsvrServiceVirtualServerServiceHits,
	lbvsvrServiceActiveTransactions,
	nsCPUUsagePct,
	nsMgmtCPUUsagePct,
	nsMemUsagePct,
	nsPktCPUUsagePct,
	nsFlashPartUsage,
	nsVarPartUsage,
	nsTCPCurClientConns,
	nsTCPCurClientConnsEst,
	nsTCPCurServerConns,
	nsTCPCurServerConnsEst,
	sslCurrentSessions,
}
