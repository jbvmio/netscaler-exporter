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
	backoffTime       = 3
	defaultThreads    = 1
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

	L := ConfigureLogger(`info`, os.Stdout)
	metricsChan := make(chan bool)

	config := GetConfig(configPath)
	for _, lbs := range config.LBServers {
		client, err := netscaler.NewNitroClient(lbs.URL, lbs.User, lbs.Pass, lbs.IgnoreCert)
		if err != nil {
			L.Error("error creating client, skipping nsInstance", zap.String("nsInstance", nsInstance(lbs.URL)), zap.Error(err))
			continue
		}
		err = client.Connect()
		if err != nil {
			L.Error("error connecting client, skipping nsInstance", zap.String("nsInstance", nsInstance(lbs.URL)), zap.Error(err))
			continue
		}
		P := newPool(lbs, metricsChan, L)
		P.client = client
		pools = append(pools, P)
	}
	if len(pools) < 1 {
		L.Fatal("no valid nsInstances available, exiting")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	pools.collectMappings(nil)

	handleProm := makeProm()
	r := mux.NewRouter()
	r.Handle(`/metrics`, handleProm)
	httpSrv := http.Server{
		Handler:      r,
		Addr:         `:9280`,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	api := newAPI(L)
	api.start(&httpSrv)
	pools.startCollecting()

	<-sigChan

	L.Warn("interrupt received ... stopping", zap.String(`process`, exporterName))
	pools.stopCollecting()
	api.stop(&httpSrv)

}
