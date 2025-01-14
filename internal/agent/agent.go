package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	cryproRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	hash "github.com/evildead81/metrics-and-alerts/internal/hash"
	pb "github.com/evildead81/metrics-and-alerts/internal/proto"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"google.golang.org/grpc"
)

type Agent struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	counter        int64
	host           string
	pollInterval   time.Duration
	reportInterval time.Duration
	mutex          *sync.Mutex
	ctx            context.Context
	sendAttempts   uint8
	key            string
	rateLimit      int
	publicKey      *rsa.PublicKey
	localIPAddress string
	gRPCClient     pb.MetricsServiceClient
}

// New создает инстанс агента.
func New(
	host string,
	pollInterval time.Duration,
	reportInterval time.Duration,
	ctx context.Context,
	key string,
	rateLimit int,
	cryptoKeyPath string,
) *Agent {
	agent := &Agent{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		counter:        0,
		host:           "http://" + host,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		mutex:          &sync.Mutex{},
		ctx:            ctx,
		key:            key,
		rateLimit:      rateLimit,
		localIPAddress: getLocalIP(),
	}

	if len(cryptoKeyPath) != 0 {
		publicKeyPEM, err := os.ReadFile(cryptoKeyPath)
		if err != nil {
			panic(err)
		}
		publicKeyBlock, _ := pem.Decode(publicKeyPEM)
		publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
		if err != nil {
			panic(err)
		}

		agent.publicKey = publicKey
	}

	return agent
}

// Run запускает процесс отправки метрик на указаннй эндпоинт.
func (t Agent) Run() error {
	if t.rateLimit == 0 {
		go func() error {
			for {
				time.Sleep(t.reportInterval)
				t.sendMeticList()
			}
		}()

		for {
			select {
			case <-t.ctx.Done():
				return nil
			default:
				t.refreshMetrics()
				time.Sleep(t.pollInterval)
			}
		}
	} else {
		jobs := make(chan contracts.Metrics, 100)

		var wg sync.WaitGroup

		for i := 1; i <= t.rateLimit; i++ {
			wg.Add(1)
			go t.worker(jobs, &wg)
		}

		go t.readMetircs(jobs)
		go t.readAdditionalMetrics(jobs)
		time.Sleep(t.reportInterval)

		close(jobs)
		wg.Wait()

	}

	return nil
}

func (t Agent) RunRPC() error {
	conn, err := grpc.NewClient(t.host)
	if err != nil {
		logger.Logger.Fatal()
	}
	defer conn.Close()

	t.gRPCClient = pb.NewMetricsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.ctx = ctx

	defer cancel()

	if t.rateLimit == 0 {
		go func() error {
			for {
				time.Sleep(t.reportInterval)
				t.postMetricsByRPC()
			}
		}()

		for {
			select {
			case <-t.ctx.Done():
				return nil
			default:
				t.refreshMetrics()
				time.Sleep(t.pollInterval)
			}
		}
	} else {
		jobs := make(chan contracts.Metrics, 100)

		var wg sync.WaitGroup

		for i := 1; i <= t.rateLimit; i++ {
			wg.Add(1)
			go t.rpcWorker(jobs, &wg)
		}

		go t.readMetircs(jobs)
		go t.readAdditionalMetrics(jobs)
		time.Sleep(t.reportInterval)

		close(jobs)
		wg.Wait()

	}

	return nil
}

func (t *Agent) worker(jobs <-chan contracts.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		t.serializeMetricAndPost(&j)
	}
}

func (t *Agent) rpcWorker(jobs <-chan contracts.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for j := range jobs {
		t.postMetricByRPC(&j)
	}
}

func (t *Agent) sendGauge(name string, value float64) {
	metric := contracts.Metrics{
		ID:    name,
		Value: &value,
		MType: consts.Gauge,
	}
	t.serializeMetricAndPost(&metric)
}

func (t *Agent) sendCounter(name string, delta int64) {
	metric := contracts.Metrics{
		ID:    name,
		Delta: &delta,
		MType: consts.Counter,
	}
	t.serializeMetricAndPost(&metric)
}

func (t *Agent) sendMetricsByOne() error {
	for name, value := range t.gaugeMetrics {
		t.sendGauge(name, value)
	}
	for name, value := range t.counterMetrics {
		t.sendCounter(name, value)
	}
	return nil
}

func (t *Agent) sendMeticList() error {
	metrics := make([]contracts.Metrics, 0)
	for name, value := range t.gaugeMetrics {
		metrics = append(metrics, contracts.Metrics{
			ID:    name,
			Value: &value,
			MType: consts.Gauge,
		})
	}
	for name, value := range t.counterMetrics {
		metrics = append(metrics, contracts.Metrics{
			ID:    name,
			Delta: &value,
			MType: consts.Counter,
		})
	}

	err := t.serializeMetricsAndPost(&metrics)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			if t.sendAttempts < 3 {
				t.sendAttempts += 1
				time.Sleep(time.Duration(t.sendAttempts*2-1) * time.Second)
				t.sendMeticList()
			}
		}
		return err
	}
	return nil
}

func (t *Agent) serializeMetricAndPost(metric *contracts.Metrics) error {
	url := t.host + "/update/"
	serialized, serErr := json.Marshal(metric)
	if serErr != nil {
		return serErr
	}

	var encryptedData []byte
	if t.publicKey != nil {
		encryptedData, serErr = rsa.EncryptPKCS1v15(cryproRand.Reader, t.publicKey, serialized)
		if serErr != nil {
			return serErr
		}
	} else {
		encryptedData = serialized
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encryptedData))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, zipErr := zb.Write(encryptedData)
	if zipErr != nil {
		return zipErr
	}
	defer zb.Close()

	if len(t.key) != 0 {
		hashStr, err := hash.Hash(encryptedData, t.key)
		if err != nil {
			return err
		}
		req.Header.Set(hash.HashHeaderKey, hashStr)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Real-IP", t.localIPAddress)
	response, reqErr := http.DefaultClient.Do(req)
	if reqErr != nil {
		return reqErr
	}
	defer response.Body.Close()

	return nil
}

func (t *Agent) postMetricByRPC(metric *contracts.Metrics) error {
	_, err := t.gRPCClient.UpdateMetric(t.ctx, &pb.UpdateMetricRequest{Metric: &pb.Metrics{
		Id:    metric.ID,
		Type:  metric.MType,
		Delta: *metric.Delta,
		Value: *metric.Value,
	}})

	if err != nil {
		return err
	}

	return nil
}

func (t *Agent) serializeMetricsAndPost(metrics *[]contracts.Metrics) error {
	url := t.host + "/updates/"
	serialized, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var encryptedData []byte
	if t.publicKey != nil {
		encryptedData, err = rsa.EncryptPKCS1v15(cryproRand.Reader, t.publicKey, serialized)
		if err != nil {
			return err
		}
	} else {
		encryptedData = serialized
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encryptedData))
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, zipErr := zb.Write(encryptedData)
	if zipErr != nil {
		return zipErr
	}
	defer zb.Close()

	if len(t.key) != 0 {
		hashStr, err := hash.Hash(encryptedData, t.key)
		if err != nil {
			return err
		}
		req.Header.Set(hash.HashHeaderKey, hashStr)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Real-IP", t.localIPAddress)
	response, reqErr := http.DefaultClient.Do(req)
	if reqErr != nil {
		return reqErr
	}
	defer response.Body.Close()
	return nil
}

func (t *Agent) postMetricsByRPC() error {
	pbMetrics := make([]*pb.Metrics, 0)
	for name, value := range *&t.gaugeMetrics {
		pbMetrics = append(pbMetrics, &pb.Metrics{
			Id:    name,
			Type:  consts.Gauge,
			Value: value,
		})
	}
	for name, value := range *&t.counterMetrics {
		pbMetrics = append(pbMetrics, &pb.Metrics{
			Id:    name,
			Type:  consts.Counter,
			Delta: value,
		})
	}

	_, err := t.gRPCClient.UpdateMetrics(t.ctx, &pb.UpdateMetricsRequest{Metrics: pbMetrics})

	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			if t.sendAttempts < 3 {
				t.sendAttempts += 1
				time.Sleep(time.Duration(t.sendAttempts*2-1) * time.Second)
				t.sendMeticList()
			}
		}
		return err
	}
	return nil
}

func (t *Agent) readMetircs(runtimeMetricsChan chan contracts.Metrics) {
	for {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		alloc := float64(stats.Alloc)
		runtimeMetricsChan <- contracts.Metrics{ID: "Alloc", Value: &alloc, MType: consts.Gauge}

		frees := float64(stats.Frees)
		runtimeMetricsChan <- contracts.Metrics{ID: "Frees", Value: &frees, MType: consts.Gauge}

		gccgpufraction := float64(stats.GCCPUFraction)
		runtimeMetricsChan <- contracts.Metrics{ID: "GCCPUFraction", Value: &gccgpufraction, MType: consts.Gauge}

		gcsys := float64(stats.GCSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "GCSys", Value: &gcsys, MType: consts.Gauge}

		heapalloc := float64(stats.HeapAlloc)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapAlloc", Value: &heapalloc, MType: consts.Gauge}

		heapidle := float64(stats.HeapIdle)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapIdle", Value: &heapidle, MType: consts.Gauge}

		heapinuse := float64(stats.HeapInuse)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapInuse", Value: &heapinuse, MType: consts.Gauge}

		heapobjects := float64(stats.HeapObjects)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapObjects", Value: &heapobjects, MType: consts.Gauge}

		headreleased := float64(stats.HeapReleased)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapReleased", Value: &headreleased, MType: consts.Gauge}

		heapsys := float64(stats.HeapSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "HeapSys", Value: &heapsys, MType: consts.Gauge}

		lastgc := float64(stats.LastGC)
		runtimeMetricsChan <- contracts.Metrics{ID: "LastGC", Value: &lastgc, MType: consts.Gauge}

		lookups := float64(stats.Lookups)
		runtimeMetricsChan <- contracts.Metrics{ID: "Lookups", Value: &lookups, MType: consts.Gauge}

		mcacheinuse := float64(stats.MCacheInuse)
		runtimeMetricsChan <- contracts.Metrics{ID: "MCacheInuse", Value: &mcacheinuse, MType: consts.Gauge}

		mcachesys := float64(stats.MCacheSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "MCacheSys", Value: &mcachesys, MType: consts.Gauge}

		mspaninuse := float64(stats.MSpanInuse)
		runtimeMetricsChan <- contracts.Metrics{ID: "MSpanInuse", Value: &mspaninuse, MType: consts.Gauge}

		mspansys := float64(stats.MSpanSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "MSpanSys", Value: &mspansys, MType: consts.Gauge}

		mallocs := float64(stats.Mallocs)
		runtimeMetricsChan <- contracts.Metrics{ID: "Mallocs", Value: &mallocs, MType: consts.Gauge}

		nextgc := float64(stats.NextGC)
		runtimeMetricsChan <- contracts.Metrics{ID: "NextGC", Value: &nextgc, MType: consts.Gauge}

		numforgedgc := float64(stats.NumForcedGC)
		runtimeMetricsChan <- contracts.Metrics{ID: "NumForcedGC", Value: &numforgedgc, MType: consts.Gauge}

		othersys := float64(stats.OtherSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "OtherSys", Value: &othersys, MType: consts.Gauge}

		pausetotalns := float64(stats.PauseTotalNs)
		runtimeMetricsChan <- contracts.Metrics{ID: "PauseTotalNs", Value: &pausetotalns, MType: consts.Gauge}

		stackinuse := float64(stats.StackInuse)
		runtimeMetricsChan <- contracts.Metrics{ID: "StackInuse", Value: &stackinuse, MType: consts.Gauge}

		stacksys := float64(stats.StackSys)
		runtimeMetricsChan <- contracts.Metrics{ID: "StackSys", Value: &stacksys, MType: consts.Gauge}

		sys := float64(stats.Sys)
		runtimeMetricsChan <- contracts.Metrics{ID: "Sys", Value: &sys, MType: consts.Gauge}

		totalalloc := float64(stats.TotalAlloc)
		runtimeMetricsChan <- contracts.Metrics{ID: "TotalAlloc", Value: &totalalloc, MType: consts.Gauge}

		randomvalue := rand.Float64()
		runtimeMetricsChan <- contracts.Metrics{ID: "RandomValue", Value: &randomvalue, MType: consts.Gauge}

		t.counterMetrics["PollCount"] += 1
		pollcount := t.counterMetrics["PollCount"]
		runtimeMetricsChan <- contracts.Metrics{ID: "PollCount", Delta: &pollcount, MType: consts.Counter}

		time.Sleep(t.pollInterval)
	}

}

func (t *Agent) readAdditionalMetrics(additMetrics chan contracts.Metrics) error {
	for {
		v, _ := mem.VirtualMemory()
		totalmemory := float64(v.Total)
		additMetrics <- contracts.Metrics{ID: "TotalMemory", Value: &totalmemory, MType: consts.Gauge}

		freemomry := float64(v.Free)
		additMetrics <- contracts.Metrics{ID: "FreeMemory", Value: &freemomry, MType: consts.Gauge}

		utilization, err := cpu.Percent(t.pollInterval, true)
		if err != nil {
			for i := 0; i < len(utilization); i++ {
				cpuutil := utilization[i]
				additMetrics <- contracts.Metrics{ID: "CPUutilization" + strconv.FormatInt(int64(i), 10), Value: &cpuutil, MType: consts.Gauge}
			}
		}

		time.Sleep(t.pollInterval)
	}

}

func (t *Agent) refreshMetrics() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	t.mutex.Lock()
	t.gaugeMetrics["Alloc"] = float64(stats.Alloc)
	t.gaugeMetrics["BuckHashSys"] = float64(stats.BuckHashSys)
	t.gaugeMetrics["Frees"] = float64(stats.Frees)
	t.gaugeMetrics["GCCPUFraction"] = float64(stats.GCCPUFraction)
	t.gaugeMetrics["GCSys"] = float64(stats.GCSys)
	t.gaugeMetrics["HeapAlloc"] = float64(stats.HeapAlloc)
	t.gaugeMetrics["HeapIdle"] = float64(stats.HeapIdle)
	t.gaugeMetrics["HeapInuse"] = float64(stats.HeapInuse)
	t.gaugeMetrics["HeapObjects"] = float64(stats.HeapObjects)
	t.gaugeMetrics["HeapReleased"] = float64(stats.HeapReleased)
	t.gaugeMetrics["HeapSys"] = float64(stats.HeapSys)
	t.gaugeMetrics["LastGC"] = float64(stats.LastGC)
	t.gaugeMetrics["Lookups"] = float64(stats.Lookups)
	t.gaugeMetrics["MCacheInuse"] = float64(stats.MCacheInuse)
	t.gaugeMetrics["MCacheSys"] = float64(stats.MCacheSys)
	t.gaugeMetrics["MSpanInuse"] = float64(stats.MSpanInuse)
	t.gaugeMetrics["MSpanSys"] = float64(stats.MSpanSys)
	t.gaugeMetrics["Mallocs"] = float64(stats.Mallocs)
	t.gaugeMetrics["NextGC"] = float64(stats.NextGC)
	t.gaugeMetrics["NumForcedGC"] = float64(stats.NumForcedGC)
	t.gaugeMetrics["NumGC"] = float64(stats.NumGC)
	t.gaugeMetrics["OtherSys"] = float64(stats.OtherSys)
	t.gaugeMetrics["PauseTotalNs"] = float64(stats.PauseTotalNs)
	t.gaugeMetrics["StackInuse"] = float64(stats.StackInuse)
	t.gaugeMetrics["StackSys"] = float64(stats.StackSys)
	t.gaugeMetrics["Sys"] = float64(stats.Sys)
	t.gaugeMetrics["TotalAlloc"] = float64(stats.TotalAlloc)
	t.gaugeMetrics["RandomValue"] = rand.Float64()
	t.counterMetrics["PollCount"] += 1
	t.mutex.Unlock()
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Logger.Fatal(err)
		panic(err)

	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String()
}
