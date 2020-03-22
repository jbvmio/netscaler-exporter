package main

import (
	"log"
	"sync"

	"github.com/jbvmio/netscaler"
)

// RawServiceStats is the payload as returned by the Nitro API.
type RawServiceStats []byte

// Len returns the size of the underlying []byte.
func (r RawServiceStats) Len() int {
	return len(r)
}

// ServiceStats represents the data returned from the /stat/service Nitro API endpoint
type ServiceStats struct {
	Name                         string   `json:"name"`
	ServiceName                  string   `json:"servicename"`
	Throughput                   string   `json:"throughput"`
	AvgTimeToFirstByte           string   `json:"avgsvrttfb"`
	State                        CurState `json:"state"`
	TotalRequests                string   `json:"totalrequests"`
	TotalResponses               string   `json:"totalresponses"`
	TotalRequestBytes            string   `json:"totalrequestbytes"`
	TotalResponseBytes           string   `json:"totalresponsebytes"`
	CurrentClientConnections     string   `json:"curclntconnections"`
	SurgeCount                   string   `json:"surgecount"`
	CurrentServerConnections     string   `json:"cursrvrconnections"`
	ServerEstablishedConnections string   `json:"svrestablishedconn"`
	CurrentReusePool             string   `json:"curreusepool"`
	MaxClients                   string   `json:"maxclients"`
	CurrentLoad                  string   `json:"curload"`
	ServiceHits                  string   `json:"vsvrservicehits"`
	ActiveTransactions           string   `json:"activetransactions"`
}

// NitroType implements the NitroData interface.
func (s ServiceStats) NitroType() string {
	return `service`
}

func processSvcStats(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	P.logger.Info("Processing Service Stats")
	data, err := P.client.GetAll(netscaler.StatsTypeService)
	if err != nil {
		log.Fatalf("error retrieving data: %v\n", err)
	}
	req := newNitroRawReq(RawServiceStats(data))
	P.team.Submit(req)
	<-req.ResultChan()
	P.logger.Info("Service Stat Processing Complete")
}
