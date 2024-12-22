package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddlewareWithoutGzipSupport(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("plain response"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "identity")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "" {
		t.Errorf("Expected no Content-Encoding header, but got '%s'", rec.Header().Get("Content-Encoding"))
	}

	if rec.Body.String() != "plain response" {
		t.Errorf("Expected response body to be 'plain response', but got '%s'", rec.Body.String())
	}
}

func TestGzipMiddlewareWithGzipSupport(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("gzip response"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding header to be 'gzip', but got '%s'", rec.Header().Get("Content-Encoding"))
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	uncompressedBody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read uncompressed response: %v", err)
	}

	if string(uncompressedBody) != "gzip response" {
		t.Errorf("Expected response body to be 'gzip response', but got '%s'", string(uncompressedBody))
	}
}

func TestGzipResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	gz := gzip.NewWriter(rec)

	gzw := gzipResponseWriter{
		ResponseWriter: rec,
		Writer:         gz,
	}

	_, err := gzw.Write([]byte("test response"))
	if err != nil {
		t.Fatalf("Failed to write gzip response: %v", err)
	}

	err = gz.Close()
	if err != nil {
		t.Fatalf("Failed to close gzip.Writer: %v", err)
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	uncompressedBody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read uncompressed response: %v", err)
	}

	if string(uncompressedBody) != "test response" {
		t.Errorf("Expected uncompressed body to be 'test response', but got '%s'", string(uncompressedBody))
	}
}
