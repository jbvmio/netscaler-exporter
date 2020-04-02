package main

import (
	"log"

	"github.com/spf13/viper"
)

// Config parameters for the exporter:
type Config struct {
	LogLevel  string     `yaml:"loglevel"`
	LBServers []LBServer `yaml:"lbservers"`
}

// LBServer details for a Netscaler LB:
type LBServer struct {
	URL             string   `yaml:"url"`
	User            string   `yaml:"user"`
	Pass            string   `yaml:"pass"`
	IgnoreCert      bool     `yaml:"ignoreCert"`
	PoolWorkers     int      `yaml:"poolWorkers"`
	PoolWorkerQueue int      `yaml:"poolWorkerQueue"`
	CollectMappings bool     `yaml:"collectMappings"`
	MappingsURL     string   `yaml:"mappingsUrl"`
	Metrics         []string `yaml:"metrics"`
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
	viper.SetDefault(`loglevel`, `info`)
	C.LogLevel = viper.GetString(`loglevel`)
	for _, c := range C.LBServers {
		if c.PoolWorkers < len(c.Metrics)*10 {
			c.PoolWorkers = len(c.Metrics) * 10
		}
		if c.PoolWorkerQueue < 1000 {
			c.PoolWorkerQueue = 1000
		}
	}
	return &C
}
