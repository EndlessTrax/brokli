package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHTML(t *testing.T) {
	t.Run("successful HTTP request", func(t *testing.T) {
		// Create a test server
		testHTML := `<html><body><a href="/test">Test Link</a></body></html>`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(testHTML))
		}))
		defer server.Close()

		html, err := GetHTML(server.URL)
		if err != nil {
			t.Errorf("Expected no error for successful request, got %v", err)
		}

		if string(html) != testHTML {
			t.Errorf("Expected HTML content '%s', got '%s'", testHTML, string(html))
		}
	})

	t.Run("HTTP error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		// GetHTML should still return the content even for 404 (it's up to the caller to check status)
		html, err := GetHTML(server.URL)
		if err != nil {
			t.Errorf("Expected no error for HTTP error response, got %v", err)
		}

		if string(html) != "Not Found" {
			t.Errorf("Expected error content 'Not Found', got '%s'", string(html))
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := GetHTML("not-a-valid-url")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("network error", func(t *testing.T) {
		// Use a URL that will cause a network error
		_, err := GetHTML("http://localhost:99999/nonexistent")
		if err == nil {
			t.Error("Expected error for network error")
		}
	})
}
