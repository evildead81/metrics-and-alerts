package instance

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
	"github.com/evildead81/metrics-and-alerts/internal/server/middlewares"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	"github.com/go-chi/chi/v5"
	chiMid "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ServerInstance struct {
	endpoint      string
	storage       storages.Storage
	storeInterval time.Duration
	db            *sql.DB
}

func New(endpoint string, storage *storages.Storage, storeInterval time.Duration) *ServerInstance {
	instance := ServerInstance{
		endpoint:      endpoint,
		storage:       *storage,
		storeInterval: storeInterval,
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
	r.Get("/ping", handlers.PingDB(t.db))
	t.runSaver()

	srv := &http.Server{
		Addr:    t.endpoint,
		Handler: r,
	}
	srvErrs := make(chan error, 1)
	go func() {
		srvErrs <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	shutdown := t.gracefulShutdown(srv)

	select {
	case err := <-srvErrs:
		shutdown(err)
	case sig := <-quit:
		shutdown(sig)
	}
}

func (t ServerInstance) gracefulShutdown(srv *http.Server) func(reason interface{}) {
	return func(reason interface{}) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		t.storage.Write()
		srv.Shutdown(ctx)
	}
}

func (t ServerInstance) runSaver() {
	go func() {
		for {
			time.Sleep(t.storeInterval * time.Second)
			t.storage.Write()
		}
	}()
}
