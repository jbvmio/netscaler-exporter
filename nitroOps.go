package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"

	"github.com/jbvmio/netscaler"
	"github.com/jbvmio/work"
)

const modelRegexStr = `[A-Z]{4,}[ -][0-9]+`

// SvcBind represents a service bind configuration.
type SvcBind struct {
	Name        string   `json:"name"`
	ServiceName string   `json:"servicename"`
	Curstate    CurState `json:"curstate"`
	colTime     int64
}

// NSHardware represents NSHardware data returned from the Nitro API
type NSHardware struct {
	HWDescription   string `json:"hwdescription"`
	ManufactureYear int    `json:"manufactureyear"`
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

// GetNSInfo returns the model, verion and manufacture year for the Netscaler Appliance.
func GetNSInfo(client *netscaler.NitroClient) (model, version string, year int, err error) {
	version, err = GetNSVersion(client)
	if err != nil {
		return
	}
	model, year, err = GetNSModelAndYear(client)
	if err != nil {
		return
	}
	return
}

// GetNSVersion take a NitroClient and returns the Netscaler Version.
func GetNSVersion(client *netscaler.NitroClient) (string, error) {
	b, err := client.GetAll(netscaler.ConfigTypeNSVersion)
	if err != nil {
		return "", err
	}
	tmp := struct {
		Target struct {
			Version string `json:"version"`
		} `json:"nsversion"`
	}{}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return "", err
	}
	return tmp.Target.Version, nil
}

// GetNSModelAndYear take a NitroClient and returns the Netscaler Model and Year.
func GetNSModelAndYear(client *netscaler.NitroClient) (string, int, error) {
	b, err := client.GetAll(netscaler.ConfigTypeNSHardware)
	if err != nil {
		return "", 0, err
	}
	var nsh NSHardware
	tmp := struct {
		Data *NSHardware `json:"nshardware"`
	}{
		Data: &nsh,
	}
	err = json.Unmarshal(b, &tmp)
	if err != nil {
		return "", 0, err
	}
	regex := regexp.MustCompile(modelRegexStr)
	model := `unknown`
	if regex.MatchString(nsh.HWDescription) {
		groups := regex.FindStringSubmatch(nsh.HWDescription)
		if len(groups) > 0 {
			model = groups[0]
		}
	}
	return model, nsh.ManufactureYear, nil
}

func collectInfoHandler(w http.ResponseWriter, r *http.Request) {
	go pools.collectNSInfo()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"request sent"}`))
}
