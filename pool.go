package main

import (
	"container/ring"
	"encoding/json"
	"sync"
	"time"

	"github.com/jbvmio/netscaler"
	"github.com/jbvmio/work"
	"go.uber.org/zap"
)

// Pool for exporting metrics for a lbserver.
type Pool struct {
	team            *work.Team
	client          *netscaler.NitroClient
	clientPool      []*netscaler.NitroClient
	poolIdx         *ring.Ring
	poolLock        *sync.Mutex
	poolWG          sync.WaitGroup
	backoff         *MiscMap
	poolFlipBit     *FlipBit
	metricClients   map[string]*netscaler.NitroClient
	metricHandlers  map[string]metricHandleFunc
	metricFlipBit   map[string]*FlipBit
	vipMap          VIPMap
	lbserver        LBServer
	nsInstance      string
	collectMappings bool
	mappingsLoaded  bool
	stopped         bool
	nsVersion       string
	logger          *zap.Logger
}

func newPool(lbs LBServer, logger *zap.Logger, loglevel string) *Pool {
	noClients := lbs.PoolWorkers
	conf := work.NewTeamConfig()
	conf.Name = lbs.URL
	conf.Workers = lbs.PoolWorkers
	conf.WorkerQueueSize = lbs.PoolWorkerQueue
	team := work.NewTeam(conf)
	pool := Pool{
		team:            team,
		poolIdx:         ring.New(noClients),
		poolLock:        &sync.Mutex{},
		poolWG:          sync.WaitGroup{},
		lbserver:        lbs,
		collectMappings: lbs.CollectMappings,
		poolFlipBit:     &FlipBit{lock: sync.Mutex{}},
		metricFlipBit:   make(map[string]*FlipBit, len(lbs.Metrics)),
		backoff: &MiscMap{
			data: make(map[string]interface{}, len(lbs.Metrics)),
			lock: sync.Mutex{},
		},
		nsInstance: nsInstance(lbs.URL),
		logger:     logger.With(zap.String(`nsInstance`, nsInstance(lbs.URL))),
	}
	if loglevel == `trace` {
		team.Logger = pool.logger
	}
	pool.logger.Info("registered netscaler instance")
	pool.logger.Info("registered lbserverUrl", zap.String("lbserverUrl", lbs.URL))
	pool.vipMap = VIPMap{
		mappings: make(map[string]map[string]string),
		lock:     sync.Mutex{},
	}
	pool.logger.Info("registering metrics")
	metricHandlers := make(map[string]metricHandleFunc, len(lbs.Metrics))
	for _, m := range lbs.Metrics {
		_, ok := metricsMap[m]
		switch {
		case ok:
			pool.logger.Info("registering metric", zap.String("metric", m))
			metricHandlers[m] = metricsMap[m]
			pool.metricFlipBit[m] = &FlipBit{lock: sync.Mutex{}}
		default:
			pool.logger.Warn("invalid metric", zap.String("metric", m))
		}
	}
	if _, there := metricHandlers[lbvserviceSubsystem]; there {
		if _, hasIt := metricHandlers[servicesSubsystem]; hasIt {
			pool.logger.Info("unregistering metric", zap.String("metric", servicesSubsystem))
			delete(metricHandlers, servicesSubsystem)
			delete(pool.metricFlipBit, servicesSubsystem)
		}
		if _, hasIt := metricHandlers[lbvserverSubsystem]; hasIt {
			pool.logger.Info("unregistering metric", zap.String("metric", lbvserverSubsystem))
			delete(metricHandlers, lbvserverSubsystem)
		}
		pool.collectMappings = false
	}
	pool.metricHandlers = metricHandlers
	clientPool := make([]*netscaler.NitroClient, noClients)
	for i := 0; i < noClients; i++ {
		client := netscaler.NewClient(lbs.URL, lbs.User, lbs.Pass, lbs.IgnoreCert)
		client.WithHTTPTimeout(time.Second * 30)
		clientPool[i] = client
		pool.poolIdx.Value = i
		pool.poolIdx = pool.poolIdx.Next()
	}
	pool.clientPool = clientPool
	pool.team.AddTask(int(nitroTaskAPI), pool.nitroAPITask)
	pool.team.AddTask(int(nitroTaskRaw), pool.nitroRawTask)
	pool.team.AddTask(int(nitroTaskData), pool.nitroDataTask)
	pool.team.AddTask(int(nitroProm), pool.nitroPromTask)
	return &pool
}

func (p *Pool) startTeam(wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Info("starting workers")
	p.team.Start()
}

func (p *Pool) stopTeam(wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Warn("stopping workers")
	p.team.Stop()
}

func (p *Pool) closeClientPool(wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Warn("disconnecting clients")
	for _, client := range p.clientPool {
		client.Disconnect()
	}
	p.client.Disconnect()
	p.logger.Warn("disconnecting clients complete")
}

func (p *Pool) submit(request work.TaskRequest) bool {
	switch {
	case p.stopped:
		if request.ResultChan() != nil {
			request.ResultChan() <- false
			close(request.ResultChan())
		}
		return false
	default:
		return p.team.Submit(request)
	}
}

func (p *Pool) getNextClient() *netscaler.NitroClient {
	i := p.poolIdx.Value.(int)
	p.poolIdx = p.poolIdx.Next()
	p.logger.Debug("Retrieving Next Client in Client Pool", zap.Int("Client ID", i))
	return p.clientPool[i]
}

func (t nitroTaskReq) ReqType() work.RequestType {
	return t.taskID
}

func (t nitroTaskReq) ResultChan() chan interface{} {
	return t.result
}

func (t nitroTaskReq) Get() interface{} {
	return t.data
}

func (t nitroTaskReq) ConsistID() string {
	return t.taskID.String()
}

func (p *Pool) nitroAPITask(req work.TaskRequest) {
	timeNow := time.Now().UnixNano()
	p.logger.Debug("Recieved nitroAPI Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
	var b []byte
	var err error
	client := p.getNextClient()
	R := req.(*nitroTaskReq)
	switch len(R.targets) {
	case 0:
		p.logger.Debug("Sending GetAll API Req", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		b, err = client.GetAll(R.nitroID)
		if err != nil {
			p.logger.Error("error retrieving API data", zap.Error(err))
			R.ResultChan() <- []byte{}
			close(R.ResultChan())
			return
		}
	case 1:
		p.logger.Debug("Sending Targed API Req", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		t := R.targets[0]
		b, err = client.Get(R.nitroID, t)
		if err != nil {
			p.logger.Error("error retrieving API data", zap.Error(err))
			R.ResultChan() <- []byte{}
			close(R.ResultChan())
			return
		}
	default:
		payloads := make([]RawData, len(R.targets))
		for i := 0; i < len(R.targets); i++ {
			apiReq := newNitroAPIReq(netscaler.StatsType(R.nitroID), R.targets[i])
			p.submit(apiReq)
			data := <-apiReq.ResultChan()
			b, ok := data.([]byte)
			if ok {
				payloads[i] = RawData(b)
			}
		}
		R.ResultChan() <- payloads
		close(R.ResultChan())
		return
	}
	R.ResultChan() <- b
	close(R.ResultChan())
	p.logger.Debug("Completed nitroAPI Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
}

func (p *Pool) nitroRawTask(req work.TaskRequest) {
	timeNow := time.Now().UnixNano()
	var noErr = true
	p.logger.Debug("Recieved nitroRaw Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
	R := req.(*nitroTaskReq)
	switch data := R.data.(type) {
	case RawServiceStats:
		p.logger.Debug("Identified nitroRaw Task Type as RawServiceStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		var stats []ServiceStats
		tmp := struct {
			Target *[]ServiceStats `json:"service"`
		}{Target: &stats}
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			p.logger.Error("Recieved nitroRaw Task Error", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Error(err))
			R.ResultChan() <- false
			close(R.ResultChan())
			return
		}
		p.logger.Debug("Processed RawServiceStats", zap.String("TaskType", req.ReqType().String()), zap.Int("Number of Stats", len(stats)), zap.Int64("TaskTS", timeNow))
		for _, s := range stats {
			datReq := newNitroDataReq(s)
			success := p.submit(datReq)
			p.logger.Debug("Sending nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
			if !success {
				noErr = false
			}
		}
	case RawLBVServerStats:
		p.logger.Debug("Identified nitroRaw Task Type as RawLBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		var stats []LBVServerStats
		tmp := struct {
			Target *[]LBVServerStats `json:"lbvserver"`
		}{Target: &stats}
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			p.logger.Error("Recieved nitroRaw Task Error", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Error(err))
			R.ResultChan() <- false
			close(R.ResultChan())
			return
		}
		for _, s := range stats {
			datReq := newNitroDataReq(s)
			success := p.submit(datReq)
			p.logger.Debug("Sending nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
			if !success {
				noErr = false
			}
		}
		p.logger.Debug("Processed RawLBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int("Number of Stats", len(stats)), zap.Int64("TaskTS", timeNow))
	case RawNSStats:
		p.logger.Debug("Identified nitroRaw Task Type as RawNSStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		var stats NSStats
		tmp := struct {
			Target *NSStats `json:"ns"`
		}{Target: &stats}
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			p.logger.Error("Recieved nitroRaw Task Error", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Error(err))
			R.ResultChan() <- false
			close(R.ResultChan())
			return
		}
		p.logger.Debug("Processed RawNSStats", zap.String("TaskType", req.ReqType().String()), zap.Int("Number of Stats", 1), zap.Int64("TaskTS", timeNow))
		datReq := newNitroDataReq(stats)
		noErr = p.submit(datReq)
		p.logger.Debug("Sending nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", noErr))
	case RawSSLStats:
		p.logger.Debug("Identified nitroRaw Task Type as RawSSLStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		var stats SSLStats
		tmp := struct {
			Target *SSLStats `json:"ssl"`
		}{Target: &stats}
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			p.logger.Error("Recieved nitroRaw Task Error", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Error(err))
			R.ResultChan() <- false
			close(R.ResultChan())
			return
		}
		p.logger.Debug("Processed RawSSLStats", zap.String("TaskType", req.ReqType().String()), zap.Int("Number of Stats", 1), zap.Int64("TaskTS", timeNow))
		datReq := newNitroDataReq(stats)
		noErr = p.submit(datReq)
		p.logger.Debug("Sending nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", noErr))
	}
	R.ResultChan() <- noErr
	close(R.ResultChan())
	p.logger.Debug("Completed nitroRaw Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
}

func (p *Pool) nitroDataTask(req work.TaskRequest) {
	timeNow := time.Now().UnixNano()
	var success bool
	var sub string
	p.logger.Debug("Recieved nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
	R := req.(*nitroTaskReq)
	switch data := R.data.(type) {
	case ServiceStats:
		sub = servicesSubsystem
		p.logger.Debug("Identified nitroData Task Type as ServiceStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
	case LBVServerStats:
		sub = lbvserverSubsystem
		p.logger.Debug("Identified nitroData Task Type as LBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
	case GSLBVServerStats:
		sub = gslbVServerSubsystem
		p.logger.Debug("Identified nitroData Task Type as GSLBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
	case NSStats:
		sub = nsSubsystem
		p.logger.Debug("Identified nitroData Task Type as NSStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
	case SSLStats:
		sub = sslSubsystem
		p.logger.Debug("Identified nitroData Task Type as SSLStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
	}
	if R.ResultChan() != nil {
		close(R.ResultChan())
	}
	if !success {
		exporterPromCollectFailures.WithLabelValues(p.nsInstance, sub).Inc()
	}
	p.logger.Debug("Completed nitroData Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
}

func (p *Pool) nitroPromTask(req work.TaskRequest) {
	timeNow := time.Now().UnixNano()
	p.logger.Debug("Recieved nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
	R := req.(*nitroTaskReq)
	switch data := R.data.(type) {
	case ServiceStats:
		p.logger.Debug("Identified nitroProm Task Type as ServiceStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promLBVServerStats(data)
	case LBVServerStats:
		p.logger.Debug("Identified nitroProm Task Type as LBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promLBVServerStats(data)
	case GSLBVServerStats:
		p.logger.Debug("Identified nitroProm Task Type as GSLBVServerStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promGSLBVServerStats(data)
	case NSStats:
		p.logger.Debug("Identified nitroProm Task Type as NSStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promNSStats(data)
	case SSLStats:
		p.logger.Debug("Identified nitroProm Task Type as SSLStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promSSLStats(data)
	}
	if R.ResultChan() != nil {
		close(R.ResultChan())
	}
	p.logger.Debug("Completed nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
}

type nitroTaskReq struct {
	taskID  TaskID
	nitroID netscaler.StatsType
	targets []string
	data    interface{}
	result  chan interface{}
}

func newNitroAPIReq(id netscaler.StatsType, targets ...string) *nitroTaskReq {
	return &nitroTaskReq{
		taskID:  nitroTaskAPI,
		nitroID: id,
		targets: targets,
		result:  work.NewResultChannel(),
	}
}

func newNitroRawReq(n NitroRaw) *nitroTaskReq {
	return &nitroTaskReq{
		taskID: nitroTaskRaw,
		data:   n,
		result: work.NewResultChannel(),
	}
}

func newNitroDataReq(n NitroData) *nitroTaskReq {
	return &nitroTaskReq{
		taskID: nitroTaskData,
		data:   n,
		result: work.NewResultChannel(),
	}
}

func newPromTask(n NitroData) *nitroTaskReq {
	return &nitroTaskReq{
		taskID: nitroProm,
		data:   n,
		result: work.NewResultChannel(),
	}
}

func submitAPITask(P *Pool, stat netscaler.StatsType, targets ...string) []byte {
	var data []byte
	var valid bool
	apiReq := newNitroAPIReq(stat, targets...)
	success := P.submit(apiReq)
	if !success {
		return []byte{}
	}
	timeout := time.NewTimer(time.Second * 10)
	select {
	case b := <-apiReq.ResultChan():
		data, valid = b.([]byte)
	case <-timeout.C:
		valid = false
	}
	switch {
	case !valid:
		return []byte{}
	}
	return data
}
