package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingResponseWriterWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	responseData := &responseData{}
	lrw := &loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   responseData,
	}

	data := []byte("Hello, world!")
	n, err := lrw.Write(data)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, but got %d", len(data), n)
	}

	if responseData.size != len(data) {
		t.Errorf("Expected response size %d, but got %d", len(data), responseData.size)
	}
}

func TestLoggingResponseWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	responseData := &responseData{}
	lrw := &loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   responseData,
	}

	statusCode := http.StatusCreated
	lrw.WriteHeader(statusCode)

	if responseData.status != statusCode {
		t.Errorf("Expected status code %d, but got %d", statusCode, responseData.status)
	}
}
