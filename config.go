package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config parameters for the exporter:
type Config struct {
	LogLevel  string        `yaml:"loglevel"`
	Interval  time.Duration `yaml:"interval"`
	LBServers []LBServer    `yaml:"lbservers"`
}

// LBServer details for a Netscaler LB:
type LBServer struct {
	URL             string       `yaml:"url"`
	User            string       `yaml:"user"`
	Pass            string       `yaml:"pass"`
	IgnoreCert      bool         `yaml:"ignoreCert"`
	PoolWorkers     int          `yaml:"poolWorkers"`
	PoolWorkerQueue int          `yaml:"poolWorkerQueue"`
	CollectMappings bool         `yaml:"collectMappings"`
	MappingsURL     string       `yaml:"mappingsUrl"`
	UploadConfig    UploadConfig `yaml:"uploadConfig"`
	Metrics         []string     `yaml:"metrics"`
}

// UploadConfig is used for saving mappings to the MappingsURL.
type UploadConfig struct {
	UploadURL string            `yaml:"uploadUrl"`
	Method    string            `yaml:"method"`
	Headers   map[string]string `yaml:"headers"`
	Insecure  bool              `yaml:"insecure"`
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
	viper.SetDefault(`interval`, `5s`)
	C.LogLevel = viper.GetString(`loglevel`)
	C.Interval = viper.GetDuration(`interval`)
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
