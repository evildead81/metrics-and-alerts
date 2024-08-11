package instance

import (
	"net/http"
	"strconv"

	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

type ServerInstance struct {
	port    string
	storage storages.Storage
}

func New(port string) *ServerInstance {
	instance := ServerInstance{
		port:    port,
		storage: storages.New(),
	}
	return &instance
}

func (t *ServerInstance) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", http.HandlerFunc(t.update))
	err := http.ListenAndServe(":"+t.port, mux)
	if err != nil {
		panic(err)
	}
}

func (t *ServerInstance) update(w http.ResponseWriter, r *http.Request) {
	metricTypeParam := r.PathValue("metricType")
	metricNameParam := r.PathValue("metricName")
	metricValueParam := r.PathValue("metricValue")

	switch {
	case r.Method != http.MethodPost:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case len(metricNameParam) == 0:
		w.WriteHeader(http.StatusNotFound)
	case metricTypeParam == consts.Gauge:
		if parsed, err := strconv.ParseFloat(metricValueParam, 64); err == nil {
			t.storage.UpdateGauge(metricNameParam, parsed)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case metricTypeParam == consts.Counter:
		if parsed, err := strconv.ParseInt(metricValueParam, 10, 64); err == nil {
			t.storage.UpdateCounter(metricNameParam, parsed)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
