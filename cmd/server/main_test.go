package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	"github.com/stretchr/testify/assert"
)

func TestStatusHandler(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		params struct {
			metricType  string
			metricName  string
			metricValue string
		}
		want int
	}{
		{
			name: "positive case",
			url:  "/update",
			params: struct {
				metricType  string
				metricName  string
				metricValue string
			}{
				metricType:  "counter",
				metricName:  "someMetric",
				metricValue: "527",
			},
			want: http.StatusOK,
		},
		{
			name: "negative case #1",
			url:  "/update",
			params: struct {
				metricType  string
				metricName  string
				metricValue string
			}{
				metricType:  "incorrect_type",
				metricName:  "someMetric",
				metricValue: "527",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "negative case #2",
			url:  "/update",
			params: struct {
				metricType  string
				metricName  string
				metricValue string
			}{
				metricType:  "counter",
				metricName:  "",
				metricValue: "527",
			},
			want: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.url, nil)
			request.SetPathValue("metricType", test.params.metricType)
			request.SetPathValue("metricName", test.params.metricName)
			request.SetPathValue("metricValue", test.params.metricValue)
			w := httptest.NewRecorder()
			storage := storages.New()
			h := http.HandlerFunc(handlers.MemStorageUpdateHandler(storage))
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.want, res.StatusCode)
			res.Body.Close()
		})
	}
}
