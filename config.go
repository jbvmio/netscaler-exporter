package main

import (
	"log"

	"github.com/spf13/viper"
)

// Config parameters for the exporter:
type Config struct {
	LBServers []LBServer `yaml:"lbservers"`
}

// LBServer details for a Netscaler LB:
type LBServer struct {
	URL              string   `yaml:"url"`
	User             string   `yaml:"user"`
	Pass             string   `yaml:"pass"`
	IgnoreCert       bool     `yaml:"ignoreCert"`
	CollectorThreads int      `yaml:"collectorThreads"`
	PoolWorkers      int      `yaml:"poolWorkers"`
	PoolWorkerQueue  int      `yaml:"poolWorkerQueue"`
	LogLevel         string   `yaml:"loglevel"`
	Metrics          []string `yaml:"metrics"`
}

// GetConfig reads in a config file and returns a Config.
func GetConfig(filePath string) *Config {
	viper.SetConfigFile(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Unable to Read Config: %v\n", err)
	}
	var C Config
	viper.Unmarshal(&C)
	for _, c := range C.LBServers {
		if c.CollectorThreads < 1 {
			c.CollectorThreads = 1
		}
		if c.PoolWorkers < 1 {
			c.PoolWorkers = 10
		}
		if c.PoolWorkerQueue < 20 {
			c.PoolWorkerQueue = 20
		}
		if c.LogLevel == "" {
			c.LogLevel = `info`
		}
	}
	return &C
}
