package handlers

import (
	"net/http"
	"strconv"

	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

func MemStorageUpdateHandler(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricTypeParam := r.PathValue("metricType")
		metricNameParam := r.PathValue("metricName")
		metricValueParam := r.PathValue("metricValue")

		switch {
		case r.Method != http.MethodPost:
			rw.WriteHeader(http.StatusMethodNotAllowed)
		case len(metricNameParam) == 0:
			rw.WriteHeader(http.StatusNotFound)
		case metricTypeParam == consts.Gauge:
			if parsed, err := strconv.ParseFloat(metricValueParam, 64); err == nil {
				storage.UpdateGauge(metricNameParam, parsed)
				rw.WriteHeader(http.StatusOK)
			} else {
				rw.WriteHeader(http.StatusBadRequest)
			}
		case metricTypeParam == consts.Counter:
			if parsed, err := strconv.ParseInt(metricValueParam, 10, 64); err == nil {
				storage.UpdateCounter(metricNameParam, parsed)
				rw.WriteHeader(http.StatusOK)
			} else {
				rw.WriteHeader(http.StatusBadRequest)
			}
		default:
			rw.WriteHeader(http.StatusBadRequest)
		}
	}
}
