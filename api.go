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
	stopChan chan struct{}
	wg       sync.WaitGroup
	logger   *zap.Logger
}

func newAPI(L *zap.Logger) *API {
	return &API{
		stopChan: make(chan struct{}),
		wg:       sync.WaitGroup{},
		logger:   L.With(zap.String(`process`, `Metrics Exporter API`)),
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
	servicesTotalRequestBytes,
	servicesTotalRequests,
	servicesTotalResponses,
	servicesTotalResponseBytes,
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
	exporterPromCollectFailures,
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
	servicesThroughput,
	servicesAvgTTFB,
	servicesState,
	servicesCurrentClientConns,
	servicesCurrentServerConns,
	servicesSurgeCount,
	servicesServerEstablishedConnections,
	servicesCurrentReusePool,
	servicesMaxClients,
	servicesCurrentLoad,
	servicesVirtualServerServiceHits,
	servicesActiveTransactions,
	sslCurrentSessions,
}
