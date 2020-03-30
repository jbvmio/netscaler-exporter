package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jbvmio/netscaler"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	exporterName      = `netscaler-exporter`
	namespace         = "citrixadc"
	netscalerInstance = `citrixadc_instance`
	mappingsDir       = `/tmp/mappings`
)

var (
	pools PoolCollection

	configPath string
	buildTime  string
	commitHash string
)

func main() {
	pf := pflag.NewFlagSet(exporterName, pflag.ExitOnError)
	pf.StringVarP(&configPath, `config`, `c`, "./config.yaml", `Path to config file.`)
	pf.Parse(os.Args[1:])

	config := GetConfig(configPath)
	var L *zap.Logger
	switch config.LogLevel {
	case `trace`:
		L = ConfigureLogger(`debug`, os.Stdout)
	default:
		L = ConfigureLogger(config.LogLevel, os.Stdout)
	}
	L.Info("Starting ...", zap.String(`Version`, buildTime), zap.String(`Commit`, commitHash))
	err := createDir(mappingsDir)
	if err != nil {
		L.Error("unable to create mappings directory", zap.Error(err))
	}
	for _, lbs := range config.LBServers {
		client := netscaler.NewClient(lbs.URL, lbs.User, lbs.Pass, lbs.IgnoreCert)
		P := newPool(lbs, L, config.LogLevel)
		P.nsVersion = nsVersion(GetNSVersion(client))
		P.client = client
		pools = append(pools, P)
	}
	if len(pools) < 1 {
		L.Fatal("no valid nsInstances available, exiting")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	pools.collectMappings(nil)

	R := makeCounterRegistry()
	handleProm := makeProm(R, L)
	r := mux.NewRouter()
	r.Handle(`/metrics`, handleProm)
	r.PathPrefix(`/mappings/`).Handler(http.StripPrefix(`/mappings/`, http.FileServer(http.Dir(mappingsDir))))
	httpSrv := http.Server{
		Handler:      r,
		Addr:         `:9280`,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	api := newAPI(L)
	api.start(&httpSrv)
	pools.startCollecting(L)

	<-sigChan

	L.Warn("interrupt received ... stopping", zap.String(`process`, exporterName))
	pools.stopCollecting()
	api.stop(&httpSrv)

}
