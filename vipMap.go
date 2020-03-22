package main

import (
	"io/ioutil"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// VIPMap contains mappings.
type VIPMap struct {
	mappings map[string]map[string]string
	lock     sync.Mutex
}

func (v *VIPMap) updateMappings(key string, maps map[string]string, l *zap.Logger) {
	l.Debug("Recieved Update Mapping Request", zap.Int("Mappings Recieved", len(maps)))
	v.lock.Lock()
	_, there := v.mappings[key]
	if !there {
		v.mappings[key] = make(map[string]string)
	}
	v.mappings[key] = maps
	l.Debug("Updated Mappings", zap.Int("Total Mappings", len(maps)))
	v.lock.Unlock()
}

func (v *VIPMap) exists(key string) bool {
	var there bool
	v.lock.Lock()
	_, there = v.mappings[key]
	v.lock.Unlock()
	return there
}

func (v *VIPMap) getMapping(url, key string, l *zap.Logger) string {
	var val string
	v.lock.Lock()
	val = v.mappings[url][key]
	v.lock.Unlock()
	l.Debug("Recieved Mapping Request", zap.String("key", key), zap.String("found value", val))
	return val
}

func (v *VIPMap) getMappingYaml() (y []byte, err error) {
	v.lock.Lock()
	y, err = yaml.Marshal(v.mappings)
	v.lock.Unlock()
	return
}

func (v *VIPMap) loadMappingYaml(path string) bool {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("error loading mapping file:", err)
		return false
	}
	err = yaml.Unmarshal(b, &v.mappings)
	if err != nil {
		log.Println("error unmarshaling mapping file:", err)
		return false
	}
	if len(v.mappings) > 0 {
		return true
	}
	return false
}

func collectMappings(P *Pool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	if P.stopped {
		P.logger.Info("Skipping Mapping Collection, process is stopping")
		return
	}
	var pr bool
	P.logger.Info("Collecting Mappings")
	svcB, err := GetSvcBindings(P.client)
	if err != nil {
		P.logger.Error("error retrieving data", zap.Error(err))
		if P.mappingsLoaded {
			return
		}
	} else {
		pr = true
	}
	for !pr {
		if P.stopped {
			P.logger.Info("Skipping Mapping Collection, process is stopping")
			return
		}
		time.Sleep(time.Second * 3)
		P.logger.Info("Retrying Mapping Collection")
		svcB, err = GetSvcBindings(P.client)
		if err != nil {
			P.logger.Error("error retrieving data", zap.Error(err))
		} else {
			pr = true
		}
	}
	tmpMap := make(map[string]string)
	for _, svc := range svcB {
		tmpMap[svc.ServiceName] = svc.Name
	}
	P.vipMap.updateMappings(P.lbserver.URL, tmpMap, P.logger)
	P.logger.Info("Mappings Collection Complete", zap.Int("Total Mappings", len(tmpMap)))
	P.mappingsLoaded = true
}
