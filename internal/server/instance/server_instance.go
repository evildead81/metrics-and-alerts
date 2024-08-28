package instance

import (
	"net/http"

	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
	"github.com/evildead81/metrics-and-alerts/internal/server/middlewares"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	"github.com/go-chi/chi/v5"
)

type ServerInstance struct {
	endpoint string
	storage  storages.Storage
}

func New(endpoint string) *ServerInstance {
	instance := ServerInstance{
		endpoint: endpoint,
		storage:  storages.New(),
	}
	return &instance
}

func (t ServerInstance) Run() {
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricByParamsHandler(t.storage))
	r.Post("/update/", handlers.UpdateMetricByJsonHandler(t.storage))
	r.Get("/value/{metricType}/{metricName}", handlers.GetMetricByParams(t.storage))
	r.Post("/value/", handlers.GetMetricByJson(t.storage))
	r.Get("/", handlers.GetPage(t.storage))
	err := http.ListenAndServe(t.endpoint, r)
	if err != nil {
		panic(err)
	}
}
