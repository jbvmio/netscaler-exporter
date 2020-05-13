package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	mappingSubsystem = "mapping"
)

// VIPMap contains mappings.
type VIPMap struct {
	mappings map[string]map[string][]string
	lock     sync.Mutex
}

func (v *VIPMap) updateMappings(key string, maps map[string][]string, uploadConfig UploadConfig, l *zap.Logger) {
	var updated bool
	var err, uploadErr error
	l.Debug("Recieved Update Mapping Request", zap.Int("Mappings Recieved", len(maps)))
	v.lock.Lock()
	_, there := v.mappings[key]
	if !there {
		v.mappings[key] = make(map[string][]string)
	}
	switch {
	case !reflect.DeepEqual(v.mappings[key], maps):
		v.mappings[key] = maps
		updated = true
		l.Info("Updated Mappings", zap.Int("Total Mappings", len(maps)))
		err = saveMappingYaml(v.mappings, mappingsDir+`/`+key+`.yaml`)
		b, errd := yaml.Marshal(v.mappings)
		if errd != nil {
			l.Info("error converting mappings to yaml", zap.Error(err))
		}
		uploadErr = uploadMappingYaml(uploadConfig, b, l)
	default:
		l.Info("No New Mappings, skipping update ...", zap.Int("Total Mappings", len(maps)))
	}
	v.lock.Unlock()
	switch {
	case updated:
		switch {
		case err != nil:
			l.Error("error saving mappings", zap.String(`instance`, key), zap.Error(err))
			if uploadErr != nil {
				l.Error("error uploading mappings", zap.String(`instance`, key), zap.Error(err))
			}
		case uploadErr != nil:
			l.Error("error uploading mappings", zap.String(`instance`, key), zap.Error(err))
		default:
			l.Info("successfully saved mappings to " + mappingsDir + `/` + key + `.yaml`)
		}
	}
}

func (v *VIPMap) exists(nsInstance, key string) bool {
	var there bool
	v.lock.Lock()
	_, there = v.mappings[key]
	v.lock.Unlock()
	return there
}

func (v *VIPMap) getMappings(nsInstance, key string, l *zap.Logger) []string {
	var val []string
	v.lock.Lock()
	val = v.mappings[nsInstance][key]
	v.lock.Unlock()
	l.Debug("Recieved Mapping Request", zap.String("key", key), zap.Strings("found values", val))
	return val
}

func (v *VIPMap) getMappingYaml() (y []byte, err error) {
	v.lock.Lock()
	y, err = yaml.Marshal(v.mappings)
	v.lock.Unlock()
	return
}

func (v *VIPMap) loadMappingYaml(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &v.mappings)
	if err != nil {
		return err
	}
	if len(v.mappings) < 1 {
		return fmt.Errorf("loaded mappings file contained zero entries")
	}
	return nil
}

func (v *VIPMap) loadMappingFromURLYaml(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &v.mappings)
	if err != nil {
		return err
	}
	if len(v.mappings) < 1 {
		return fmt.Errorf("loaded mappings file contained zero entries")
	}
	return nil
}

func (v *VIPMap) saveMappingYaml(path string) error {
	return saveMappingYaml(v.mappings, path)
}

func saveMappingYaml(mappings map[string]map[string][]string, path string) error {
	b, err := yaml.Marshal(mappings)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, b, 0664)
	if err != nil {
		return err
	}
	return nil
}

func uploadMappingYaml(c UploadConfig, payload []byte, l *zap.Logger) error {
	if c.UploadURL == "" || len(payload) < 1 {
		return nil
	}
	req, err := http.NewRequest(c.Method, c.UploadURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	for k, v := range c.Headers {
		req.Header.Add(k, v)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 201:
		l.Info("successfully uploaded mappings", zap.String(`uploadURL`, c.UploadURL))
		return nil
	default:
		return fmt.Errorf("received non 201 status: %s", resp.Status)
	}
}

func collectMappings(P *Pool, force bool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	if P.stopped {
		P.logger.Info("Skipping Mapping Collection, process is stopping")
		return
	}
	switch {
	case P.mappingFlipBit.good():
		defer P.mappingFlipBit.flip()
		switch {
		case force:
			P.logger.Info("Refreshing Mappings")
		case P.lbserver.MappingsURL != "":
			err := P.vipMap.loadMappingFromURLYaml(P.lbserver.MappingsURL)
			switch {
			case err == nil:
				P.logger.Info("Loaded mappings from url", zap.Int("Total Mappings", len(P.vipMap.mappings[P.nsInstance])))
				P.mappingsLoaded = true
				P.vipMap.saveMappingYaml(mappingsDir + `/` + P.nsInstance + `.yaml`)
				return
			default:
				P.logger.Error("could not load mappings from url, received error, trying file ...", zap.Error(err))
				err := P.vipMap.loadMappingYaml(mappingsDir + `/` + P.nsInstance + `.yaml`)
				if err == nil {
					P.logger.Info("Loaded mappings from file", zap.Int("Total Mappings", len(P.vipMap.mappings[P.nsInstance])))
					P.mappingsLoaded = true
					return
				}
			}
		default:
			err := P.vipMap.loadMappingYaml(mappingsDir + `/` + P.nsInstance + `.yaml`)
			switch {
			case err == nil:
				P.logger.Info("Loaded mappings from file", zap.Int("Total Mappings", len(P.vipMap.mappings[P.nsInstance])))
				P.mappingsLoaded = true
				return
			default:
				P.logger.Warn("could not load default mappings from file, received error", zap.Error(err))
				P.logger.Info("Collecting Mappings")
			}
		}
		var pr bool
		svcB, err := GetSvcBindings(P.client)
		if err != nil {
			P.logger.Error("error retrieving data", zap.Error(err))
			exporterAPICollectFailures.WithLabelValues(P.nsInstance, mappingSubsystem).Inc()
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
			P.client.WithHTTPTimeout(time.Second * 120)
			svcB, err = GetSvcBindings(P.client)
			if err != nil {
				P.logger.Error("error retrieving data", zap.Error(err))
				exporterAPICollectFailures.WithLabelValues(P.nsInstance, mappingSubsystem).Inc()
			} else {
				pr = true
				P.client.WithHTTPTimeout(time.Second * 60)
			}
		}
		tmpMap := make(map[string][]string)
		for _, svc := range svcB {
			tmpMap[svc.ServiceName] = append(tmpMap[svc.ServiceName], svc.Name)
		}
		P.vipMap.updateMappings(P.nsInstance, tmpMap, P.lbserver.UploadConfig, P.logger)
		P.logger.Info("Mappings Collection Complete", zap.Int("Total Mappings", len(tmpMap)))
		P.mappingsLoaded = true
	default:
		P.logger.Info("Skipping Mapping Collection, already in progress")
	}
}
