package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	hash "github.com/evildead81/metrics-and-alerts/internal/hash"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

func UpdateMetricByParamsHandler(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricTypeParam := r.PathValue("metricType")
		metricNameParam := r.PathValue("metricName")
		metricValueParam := r.PathValue("metricValue")
		switch {
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

func UpdateMetricByJSONHandler(storage storages.Storage, key string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		var metric contracts.Metrics
		if unmarshalErr := json.Unmarshal(buf.Bytes(), &metric); unmarshalErr != nil {

			http.Error(rw, unmarshalErr.Error(), http.StatusBadRequest)
			return
		}

		if len(key) != 0 {
			hashReqHeaderVal := r.Header.Get(hash.HashHeaderKey)

			hashedRequest, err := hash.Hash(buf.Bytes(), key)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			if hashReqHeaderVal != hashedRequest {
				http.Error(rw, "Incorrect hash header", http.StatusBadRequest)
				return
			}
		}

		newMetric := contracts.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
		}

		switch {
		case metric.MType == consts.Gauge:
			storage.UpdateGauge(metric.ID, *metric.Value)
			newMetric.Value = metric.Value
		case metric.MType == consts.Counter:
			storage.UpdateCounter(metric.ID, *metric.Delta)
			newMetric.Delta = metric.Delta
		default:
			http.Error(rw, "Incorrect type", http.StatusBadRequest)
			return
		}

		bytes, err := json.MarshalIndent(newMetric, "", "   ")
		if err != nil {
			http.Error(rw, "Server error", http.StatusInternalServerError)
			return
		}

		hashedResponse, err := hash.Hash(bytes, key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Add("Content-type", "application/json")
		rw.Header().Add("Accept-Encoding", "gzip")
		rw.Header().Add(hash.HashHeaderKey, hashedResponse)
		rw.WriteHeader(http.StatusOK)
		rw.Write(bytes)
	}
}

func GetMetricByParamsHandler(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricTypeParam := r.PathValue("metricType")
		metricNameParam := r.PathValue("metricName")

		switch {
		case metricTypeParam == consts.Gauge:
			value, err := storage.GetGaugeValueByName(metricNameParam)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusNotFound)
			} else {
				rw.WriteHeader(http.StatusOK)
				io.WriteString(rw, strconv.FormatFloat(value, 'f', -1, 64))
			}
		case metricTypeParam == consts.Counter:
			value, err := storage.GetCountValueByName(metricNameParam)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusNotFound)
			} else {
				rw.WriteHeader(http.StatusOK)
				io.WriteString(rw, strconv.FormatInt(value, 10))
			}
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}
}

func GetMetricByJSONHandler(storage storages.Storage, key string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric contracts.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if metric.MType != consts.Gauge && metric.MType != consts.Counter {
			http.Error(rw, "Incorrect type", http.StatusNotFound)
			return
		}

		result := contracts.Metrics{}
		result.ID = metric.ID
		result.MType = metric.MType
		if metric.MType == consts.Gauge {
			value, err := storage.GetGaugeValueByName(metric.ID)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			}
			result.Value = &value
		}

		if metric.MType == consts.Counter {
			value, err := storage.GetCountValueByName(metric.ID)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			}
			result.Delta = &value
		}

		bytes, err := json.MarshalIndent(result, "", "   ")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		rw.Header().Add("Content-Type", "application/json")
		rw.Header().Add("Accept-Encoding", "gzip")
		rw.WriteHeader(http.StatusOK)
		rw.Write(bytes)
	}
}

func GetPageHandler(storage storages.Storage) http.HandlerFunc {
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
		rw.Header().Add("Content-Type", "text/html")
		rw.Header().Add("Accept-Encoding", "gzip")
		rw.WriteHeader(http.StatusOK)
		io.WriteString(rw, sb.String())
	}
}

func Ping(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		err := storage.Ping()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func UpdateMetrics(storage storages.Storage, key string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		var metrics []contracts.Metrics
		if unmarshalErr := json.Unmarshal(buf.Bytes(), &metrics); unmarshalErr != nil {
			http.Error(rw, unmarshalErr.Error(), http.StatusBadRequest)
			return
		}

		if len(key) != 0 {
			hashReqHeaderVal := r.Header.Get(hash.HashHeaderKey)

			hashedRequest, err := hash.Hash(buf.Bytes(), key)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			if hashReqHeaderVal != hashedRequest {
				http.Error(rw, "Incorrect hash header", http.StatusBadRequest)
				return
			}
		}

		err = storage.UpdateMetrics(metrics)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		hashedResponse, err := hash.Hash(make([]byte, 0), key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set(hash.HashHeaderKey, hashedResponse)

		rw.WriteHeader(http.StatusOK)
	}
}
