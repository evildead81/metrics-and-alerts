package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	hash "github.com/evildead81/metrics-and-alerts/internal/hash"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

// UpdateMetricByParamsHandler - обновляет метрику, переданную в строке запроса.
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

// UpdateMetricByJSONHandler обновляет метрику, переданную в body в формате JSON.
func UpdateMetricByJSONHandler(storage storages.Storage, key string, privateKey *rsa.PrivateKey) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var reader io.ReadCloser
		var err error

		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, "failed to read gzip body", http.StatusBadRequest)
				logger.Logger.Error(err.Error())
				return
			}
			defer reader.Close()
		} else {
			reader = r.Body
		}

		encryptedData, err := io.ReadAll(reader)
		if err != nil {
			http.Error(rw, "failed to read request body", http.StatusBadRequest)
			logger.Logger.Error(err.Error())
			return
		}

		var decryptedData []byte
		if privateKey != nil {
			decryptedData, err = rsa.DecryptPKCS1v15(reader, privateKey, encryptedData)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				logger.Logger.Error(err.Error())
				return
			}
		} else {
			decryptedData = encryptedData
		}

		var metric contracts.Metrics
		if err := json.NewDecoder(bytes.NewReader(decryptedData)).Decode(&metric); err != nil {
			http.Error(rw, fmt.Sprintf("failed to decode JSON: %v", err), http.StatusBadRequest)
			logger.Logger.Error(err.Error())
			return
		}

		if len(key) != 0 {
			var buf bytes.Buffer
			_, err := buf.ReadFrom(reader)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				logger.Logger.Error(err.Error())
				return
			}
			hashReqHeaderVal := r.Header.Get(hash.HashHeaderKey)

			if len(hashReqHeaderVal) != 0 {

				hashedRequest, err := hash.Hash(buf.Bytes(), key)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					logger.Logger.Error(err.Error())
					return
				}

				if hashReqHeaderVal != hashedRequest {
					http.Error(rw, "Incorrect hash header", http.StatusBadRequest)
					logger.Logger.Error("Incorrect hash header")
					return
				}
			}
		}

		newMetric := contracts.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
		}

		switch {
		case metric.MType == consts.Gauge:
			err := storage.UpdateGauge(metric.ID, *metric.Value)
			if err != nil {
				http.Error(rw, "Server error", http.StatusInternalServerError)
				logger.Logger.Error(err.Error())
				return
			}
			newMetric.Value = metric.Value
		case metric.MType == consts.Counter:
			err := storage.UpdateCounter(metric.ID, *metric.Delta)
			if err != nil {
				http.Error(rw, "Server error", http.StatusInternalServerError)
				logger.Logger.Error(err.Error())
				return
			}
			updatedCounterValue, err := storage.GetCountValueByName(metric.ID)
			if err != nil {
				http.Error(rw, "Server error", http.StatusInternalServerError)
				logger.Logger.Error(err.Error())
				return
			}
			newMetric.Delta = &updatedCounterValue
		default:
			http.Error(rw, "Incorrect type", http.StatusBadRequest)
			logger.Logger.Error("Incorrect type")
			return
		}

		bytes, err := json.MarshalIndent(newMetric, "", "   ")
		if err != nil {
			http.Error(rw, "Server error", http.StatusInternalServerError)
			logger.Logger.Error(err.Error())
			return
		}

		hashedResponse, err := hash.Hash(bytes, key)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			logger.Logger.Error(err.Error())
			return
		}
		rw.Header().Add("Content-type", "application/json")
		rw.Header().Add("Accept-Encoding", "gzip")
		rw.Header().Add(hash.HashHeaderKey, hashedResponse)
		rw.WriteHeader(http.StatusOK)
		rw.Write(bytes)
	}
}

// GetMetricByParamsHandler возвращает метрику по указанным в строке запроса типу и имени.
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

// GetMetricByJSONHandler возвращает метрику по параметрам, переданным в body в формате JSON.
func GetMetricByJSONHandler(storage storages.Storage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var metric contracts.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if metric.MType == consts.Counter {
			value, err := storage.GetCountValueByName(metric.ID)
			if err != nil {
				http.Error(rw, "Metric not found", http.StatusNotFound)
				return
			}
			metric.Delta = &value
		}

		if metric.MType == consts.Gauge {
			value, err := storage.GetGaugeValueByName(metric.ID)
			if err != nil {
				http.Error(rw, "Metric not found", http.StatusNotFound)
				return
			}
			metric.Value = &value
		}

		response, err := json.Marshal(metric)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(response)
	}
}

// GetPageHandler возвращает html-страницу, где отображены текущие метрики.
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

// Ping проверяет доступность хранилища.
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

// UpdateMetrics обновляет список метрик, переданных в body в формате JSON.
func UpdateMetrics(storage storages.Storage, key string, privateKey *rsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reader io.ReadCloser
		var err error

		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to read gzip body", http.StatusBadRequest)
				return
			}
			defer reader.Close()
		} else {
			reader = r.Body
		}

		defer reader.Close()

		encryptedData, err := io.ReadAll(reader)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		var decryptedData []byte
		if privateKey != nil {
			decryptedData, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
			if err != nil {
				http.Error(w, "failed to decrypt data", http.StatusBadRequest)
				return
			}
		} else {
			decryptedData = encryptedData
		}

		var metrics []contracts.Metrics
		if err := json.NewDecoder(bytes.NewReader(decryptedData)).Decode(&metrics); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode JSON: %v", err), http.StatusBadRequest)
			return
		}

		if err := storage.UpdateMetrics(metrics); err != nil {
			fmt.Println("UPDATE METRICS ERROR", err)
			http.Error(w, fmt.Sprintf("failed to update metrics: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
