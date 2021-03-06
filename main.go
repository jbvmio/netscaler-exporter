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
	L.Info("Setting Collect Interval ...", zap.Duration("interval", config.Interval))
	collectInterval = config.Interval
	err := createDir(mappingsDir)
	if err != nil {
		L.Error("unable to create mappings directory", zap.Error(err))
	}
	for _, lbs := range config.LBServers {
		client := netscaler.NewClient(lbs.URL, lbs.User, lbs.Pass, lbs.IgnoreCert)
		//nv, err := GetNSVersion(client)
		model, ver, year, err := GetNSInfo(client)
		switch {
		case err != nil:
			L.Error("error validating client, skipping ...", zap.String(`nsInstance`, nsInstance(lbs.URL)), zap.Error(err))
		default:
			P := newPool(lbs, L, config.LogLevel)
			P.nsVersion = nsVersion(ver)
			P.nsModel = model
			P.nsYear = year
			P.client = client
			pools = append(pools, P)
		}
	}
	if len(pools) < 1 {
		L.Fatal("no valid nsInstances available, exiting")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	pools.collectMappings(nil, false)
	api := newAPI(L)

	R := makeCounterRegistry()
	handleProm := makeProm(R, L)
	r := mux.NewRouter()
	r.Handle(`/metrics`, handleProm)
	r.HandleFunc(`/ops`, api.opsHandler)
	r.HandleFunc(`/update/info`, api.collectInfoHandler)
	r.HandleFunc(`/update/mappings`, api.updateMappingsHandler)
	r.PathPrefix(`/mappings/`).Handler(http.StripPrefix(`/mappings/`, http.FileServer(http.Dir(mappingsDir))))
	httpSrv := http.Server{
		Handler:      r,
		Addr:         `:9280`,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	api.start(&httpSrv)
	pools.startCollecting(L)

	<-sigChan

	L.Warn("interrupt received ... stopping", zap.String(`process`, exporterName))
	pools.stopCollecting()
	api.stop(&httpSrv)

}
