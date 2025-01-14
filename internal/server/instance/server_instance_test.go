package instance

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	memstorage "github.com/evildead81/metrics-and-alerts/internal/server/storages/mem-storage"
	"github.com/go-chi/chi/v5"
)

type MockStorage struct {
	writeCalled int32
	storage     storages.Storage
}

func (m *MockStorage) Write() error {
	atomic.AddInt32(&m.writeCalled, 1)
	return nil
}

func (m *MockStorage) Read() error {
	return nil
}

func TestNewInstance(t *testing.T) {
	storage := &MockStorage{
		storage: memstorage.New("./metrics.json", true),
	}
	instance := New(":8080", &storage.storage, 5*time.Second, "test-key", "", "")

	if instance.endpoint != ":8080" {
		t.Errorf("Expected endpoint ':8080', got '%s'", instance.endpoint)
	}
	if instance.storeInterval != 5*time.Second {
		t.Errorf("Expected storeInterval '5s', got '%v'", instance.storeInterval)
	}
	if instance.key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", instance.key)
	}
	t.Log("New instance created successfully")
}

func TestRoutes(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	t.Log("Route /ping works correctly")
}

func TestRunServer(t *testing.T) {
	storage := &MockStorage{}
	instance := New(":8080", &storage.storage, 5*time.Second, "test-key", "", "")

	go func() {
		defer func() {
			recover()
		}()
		instance.Run()
	}()

	time.Sleep(1 * time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(1 * time.Second)
	t.Log("Server started and stopped correctly")
}
