package agent

import (
	"context"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
}

func New(host string, pollInterval time.Duration, reportInterval time.Duration, ctx context.Context) *Agent {
	return &Agent{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		counter:        0,
		host:           "http://" + host,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		mutex:          &sync.Mutex{},
		ctx:            ctx,
	}
}

func (t Agent) Run() error {
	go func() error {
		for {
			time.Sleep(t.reportInterval)
			err := t.sendMetrics()
			if err != nil {
				return err
			}
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
}

func (t *Agent) sendMetrics() error {
	for name, value := range t.gaugeMetrics {
		url := strings.Join([]string{t.host, "/update/gauge/", name, "/", strconv.FormatFloat(value, 'f', -1, 64)}, "")
		response, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
		response.Body.Close()
	}
	for name, value := range t.counterMetrics {
		url := strings.Join([]string{t.host, "/update/counter/", name, "/", strconv.FormatInt(value, 10)}, "")
		response, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
		response.Body.Close()
	}
	return nil
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
	t.gaugeMetrics["HeapIdle"] = float64(stats.HeapIdle)
	t.gaugeMetrics["HeapInuse"] = float64(stats.HeapInuse)
	t.gaugeMetrics["HeapObjects"] = float64(stats.HeapObjects)
	t.gaugeMetrics["HeapReleased"] = float64(stats.HeapReleased)
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
