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
	team           *work.Team
	client         *netscaler.NitroClient
	clientPool     []*netscaler.NitroClient
	poolIdx        *ring.Ring
	poolLock       *sync.Mutex
	poolWG         sync.WaitGroup
	metricHandlers map[string]metricHandleFunc
	backoff        *MiscMap
	flipBit        *FlipBit
	lbserver       LBServer
	nsInstance     string
	vipMap         VIPMap
	mappingsLoaded bool
	stopped        bool
	logger         *zap.Logger
}

func newPool(lbs LBServer, metricsChan chan bool, logger *zap.Logger) *Pool {
	noClients := len(lbs.Metrics) + 1
	conf := work.NewTeamConfig()
	conf.Name = lbs.URL
	conf.Workers = lbs.PoolWorkers
	conf.WorkerQueueSize = lbs.PoolWorkerQueue
	team := work.NewTeam(conf)
	pool := Pool{
		team:     team,
		poolIdx:  ring.New(noClients),
		poolLock: &sync.Mutex{},
		poolWG:   sync.WaitGroup{},
		lbserver: lbs,
		flipBit:  &FlipBit{lock: sync.Mutex{}},
		backoff: &MiscMap{
			data: make(map[string]interface{}, len(lbs.Metrics)),
			lock: sync.Mutex{},
		},
		nsInstance: nsInstance(lbs.URL),
		logger:     logger.With(zap.String(`nsInstance`, nsInstance(lbs.URL))),
	}
	team.Logger = pool.logger
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
		default:
			pool.logger.Warn("invalid metric", zap.String("metric", m))
		}
	}
	pool.metricHandlers = metricHandlers
	clientPool := make([]*netscaler.NitroClient, noClients)
	for i := 0; i < noClients; i++ {
		client, err := netscaler.NewNitroClient(lbs.URL, lbs.User, lbs.Pass, lbs.IgnoreCert)
		if err != nil {
			pool.logger.Fatal("error creating additional client", zap.Error(err))
		}
		client.WithHTTPTimeout(time.Second * 30)
		err = client.Connect()
		if err != nil {
			pool.logger.Fatal("error connecting additional client", zap.Error(err))
		}
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
		p.logger.Debug("Sending MultiTargeted API Req - SHOULD NOT SEE!!", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		for _, t := range R.targets {
			apiReq := newNitroAPIReq(netscaler.StatsType(R.nitroID), t)
			p.submit(apiReq)
			data := <-apiReq.ResultChan()
			b := data.([]byte)
			rawReq := newNitroRawReq(RawData(b))
			p.submit(rawReq)
			<-rawReq.ResultChan()
		}
		R.ResultChan() <- true
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
		p.logger.Debug("Looking up Service VIP Name", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.String("Lookup", data.Name))
		data.ServiceName = p.vipMap.getMapping(p.lbserver.URL, data.Name, p.logger)
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
		return
	case NSStats:
		sub = nsSubsystem
		p.logger.Debug("Identified nitroData Task Type as NSStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		promReq := newPromTask(data)
		success = p.submit(promReq)
		p.logger.Debug("Sending nitroProm Task", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow), zap.Bool("successful", success))
		return
	}
	if R.ResultChan() != nil {
		close(R.ResultChan())
	}
	if !success {
		exporterFailuresTotal.WithLabelValues(p.nsInstance, sub).Inc()
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
		p.promSvcStats(data)
	case NSStats:
		p.logger.Debug("Identified nitroProm Task Type as NSStats", zap.String("TaskType", req.ReqType().String()), zap.Int64("TaskTS", timeNow))
		p.promNSStats(data)
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

func submitAPITask(P *Pool, stat netscaler.StatsType) []byte {
	var data []byte
	var valid bool
	apiReq := newNitroAPIReq(stat)
	success := P.submit(apiReq)
	if !success {
		return []byte{}
	}
	b := <-apiReq.ResultChan()
	data, valid = b.([]byte)
	switch {
	case !valid:
		return []byte{}
	}
	return data
}
