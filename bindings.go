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

// GetSvcBindings take a NitroClient and returns Service Bindings
func GetSvcBindings(client *netscaler.NitroClient) ([]SvcBind, error) {
	var svcBinds []SvcBind
	b, err := client.GetAll(netscaler.ConfigTypeLBVSSvcBinding)
	if err != nil {
		return svcBinds, err
	}
	tmp := struct {
		Target *[]SvcBind `json:"lbvserver_service_binding"`
	}{Target: &svcBinds}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return svcBinds, err
	}
	return svcBinds, nil
}
