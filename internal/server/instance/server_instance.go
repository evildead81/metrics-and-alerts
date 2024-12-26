package instance

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/pprof"
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
	key           string
	privateKey    *rsa.PrivateKey
}

// New создает инстанс сервера.
func New(
	endpoint string,
	storage *storages.Storage,
	storeInterval time.Duration,
	key string,
	cryptoKeyPath string,
) *ServerInstance {
	instance := ServerInstance{
		endpoint:      endpoint,
		storage:       *storage,
		storeInterval: storeInterval,
		key:           key,
	}

	if len(cryptoKeyPath) != 0 {
		privateKeyPEM, err := os.ReadFile(cryptoKeyPath)
		if err != nil {
			panic(err)
		}

		privateKeyBlock, _ := pem.Decode(privateKeyPEM)
		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			panic(err)
		}

		instance.privateKey = privateKey
	}

	return &instance
}

// Run - запускает сервер.
func (t ServerInstance) Run() {
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	r.Use(middlewares.GzipMiddleware)
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricByParamsHandler(t.storage))
		r.Post("/", handlers.UpdateMetricByJSONHandler(t.storage, t.key, t.privateKey))
	})
	r.Post("/updates/", handlers.UpdateMetrics(t.storage, t.key, t.privateKey))
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", handlers.GetMetricByParamsHandler(t.storage))
		r.Post("/", handlers.GetMetricByJSONHandler(t.storage, t.key))
	})
	r.Get("/", handlers.GetPageHandler(t.storage))
	r.Get("/ping", handlers.Ping(t.storage))
	r.Mount("/debug", chiMid.Profiler())

	rtProf := chi.NewRouter()

	rtProf.HandleFunc("/", pprof.Index)
	rtProf.HandleFunc("/cmdline", pprof.Cmdline)
	rtProf.HandleFunc("/profile", pprof.Profile)
	rtProf.HandleFunc("/symbol", pprof.Symbol)
	rtProf.HandleFunc("/trace", pprof.Trace)

	rtProf.Handle("/goroutine", pprof.Handler("goroutine"))
	rtProf.Handle("/heap", pprof.Handler("heap"))
	rtProf.Handle("/threadcreate", pprof.Handler("threadcreate"))
	rtProf.Handle("/block", pprof.Handler("block"))
	rtProf.Handle("/mutex", pprof.Handler("mutex"))
	rtProf.Handle("/allocs", pprof.Handler("allocs"))

	r.Mount("/debug/pprof", rtProf)

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
