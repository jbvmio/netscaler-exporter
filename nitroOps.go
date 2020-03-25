package main

import (
	"encoding/json"
	"sync"

	"github.com/jbvmio/netscaler"
	"github.com/jbvmio/work"
)

// SvcBind represents a service bind configuration.
type SvcBind struct {
	Name        string   `json:"name"`
	ServiceName string   `json:"servicename"`
	Curstate    CurState `json:"curstate"`
	colTime     int64
}

func (s *SvcBind) svcStats(T *work.Team, w *sync.WaitGroup) RawServiceStats {
	defer w.Done()
	req := newNitroAPIReq(netscaler.StatsTypeService, s.ServiceName)
	T.Submit(req)
	data := <-req.result
	b := data.([]byte)
	return b
}

// GetSvcBindings take a NitroClient and returns Service Bindings.
func GetSvcBindings(client *netscaler.NitroClient) ([]SvcBind, error) {
	var target []SvcBind
	b, err := client.GetAll(netscaler.ConfigTypeLBVSSvcBinding)
	if err != nil {
		return target, err
	}
	tmp := struct {
		Target *[]SvcBind `json:"lbvserver_service_binding"`
	}{Target: &target}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return target, err
	}
	return target, nil
}

// GetNSVersion take a NitroClient and returns the Netscaler Version or an empty string if an error is encountered.
func GetNSVersion(client *netscaler.NitroClient) string {
	b, err := client.GetAll(netscaler.ConfigTypeNSVersion)
	if err != nil {
		return ""
	}
	tmp := struct {
		Target struct {
			Version string `json:"version"`
		} `json:"nsversion"`
	}{}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return ""
	}
	return tmp.Target.Version
}
