package instance

import (
	"net/http"

	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
	"github.com/evildead81/metrics-and-alerts/internal/server/middlewares"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	"github.com/go-chi/chi/v5"
	chiMid "github.com/go-chi/chi/v5/middleware"
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
	r.Use(chiMid.Compress(5))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricByParamsHandler(t.storage))
		r.Post("/", handlers.UpdateMetricByJSONHandler(t.storage))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", handlers.GetMetricByParamsHandler(t.storage))
		r.Post("/", handlers.GetMetricByJSONHandler(t.storage))
	})
	r.Get("/", handlers.GetPageHandler(t.storage))
	err := http.ListenAndServe(t.endpoint, r)
	if err != nil {
		panic(err)
	}
}
