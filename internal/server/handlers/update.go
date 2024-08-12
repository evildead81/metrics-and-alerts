package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

func UpdateMetricHandler(storage storages.Storage) http.HandlerFunc {
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

func GetMetric(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricTypeParam := r.PathValue("metricType")
		metricNameParam := r.PathValue("metricName")

		switch {
		case metricTypeParam == consts.Gauge:
			value, err := storage.GetGaugeValueByName(metricNameParam)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
			}
			rw.WriteHeader(http.StatusOK)
			io.WriteString(rw, strconv.FormatFloat(value, 'f', -1, 64))
		case metricTypeParam == consts.Counter:
			value, err := storage.GetCountValueByName(metricNameParam)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
			}
			rw.WriteHeader(http.StatusOK)
			io.WriteString(rw, strconv.FormatInt(value, 10))
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}
}

func GetPage(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		sb := strings.Builder{}
		gauges := storage.GetGauges()
		counters := storage.GetCounters()

		sb.WriteString("<!DOCTYPE html>")
		sb.WriteString("<html lang=\"en\">")
		sb.WriteString("<head><title>Metrics</title></head>")
		sb.WriteString("<body>")
		sb.WriteString("<h3>Gauge metrics</h3><br>")
		sb.WriteString("<div style=\"display:flex;flex-direction:column;gap:8px\">")
		for name, value := range gauges {
			sb.WriteString("<span> Name: ")
			sb.WriteString(name)
			sb.WriteString(", Value: ")
			sb.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			sb.WriteString("</span>")
		}
		sb.WriteString("</div>")
		sb.WriteString("<h3>Gauge metrics</h3><br>")
		sb.WriteString("<div style=\"display:flex;flex-direction:column;gap:8px\">")
		for name, value := range counters {
			sb.WriteString("<span> Name: ")
			sb.WriteString(name)
			sb.WriteString(", Value: ")
			sb.WriteString(strconv.FormatInt(value, 10))
			sb.WriteString("</span>")
		}
		sb.WriteString("</div>")
		sb.WriteString("</body>")
		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, sb.String())
	}
}
