package instance

import (
	"net/http"

	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
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

func (t ServerInstance) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.MemStorageUpdateHandler(t.storage))
	err := http.ListenAndServe(":"+t.port, mux)
	if err != nil {
		panic(err)
	}
}
