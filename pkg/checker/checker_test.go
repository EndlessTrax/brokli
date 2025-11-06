package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/endlesstrax/brokli/pkg/link"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxWorkers != 10 {
		t.Errorf("Expected MaxWorkers to be 10, got %d", config.MaxWorkers)
	}

	if config.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout to be 10s, got %v", config.Timeout)
	}

	if config.UserAgent == "" {
		t.Error("Expected UserAgent to be set")
	}
}

func TestCheckLinks_Empty(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()

	err := CheckLinks(ctx, []CheckableLink{}, config)
	if err != nil {
		t.Errorf("Expected no error for empty links, got %v", err)
	}
}

func TestCheckLinks_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("Expected HEAD request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create anchor tag
	parsedURL, _ := url.Parse(server.URL)
	tag := &link.AnchorTag{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()
	config.MaxWorkers = 1

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tag.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, tag.Status)
	}
}

func TestCheckLinks_MultipleConcurrent(t *testing.T) {
	// Create test server
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		// Simulate some delay
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create multiple anchor tags
	tags := make([]*link.AnchorTag, 20)
	parsedURL, _ := url.Parse(server.URL)
	for i := range tags {
		tags[i] = &link.AnchorTag{
			AbsoluteUrl: *parsedURL,
			Status:      -1,
		}
	}

	ctx := context.Background()
	config := DefaultConfig()
	config.MaxWorkers = 5

	start := time.Now()
	err := CheckAnchorTags(ctx, tags, config)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check all links were processed
	for i, tag := range tags {
		if tag.Status != http.StatusOK {
			t.Errorf("Tag %d: expected status %d, got %d", i, http.StatusOK, tag.Status)
		}
	}

	// Concurrent execution should be faster than sequential
	// 20 requests * 10ms = 200ms sequential
	// With 5 workers: ~40ms (plus overhead)
	if duration > 150*time.Millisecond {
		t.Logf("Warning: Duration %v seems too slow for concurrent execution", duration)
	}
}

func TestCheckLinks_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		expectedStatus int
	}{
		{"OK", http.StatusOK, http.StatusOK},
		{"Not Found", http.StatusNotFound, http.StatusNotFound},
		{"Moved Permanently", http.StatusMovedPermanently, http.StatusMovedPermanently},
		{"Internal Server Error", http.StatusInternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			parsedURL, _ := url.Parse(server.URL)
			tag := &link.AnchorTag{
				AbsoluteUrl: *parsedURL,
				Status:      -1,
			}

			ctx := context.Background()
			config := DefaultConfig()

			_ = CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)

			if tag.Status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, tag.Status)
			}
		})
	}
}

func TestCheckLinks_EmptyURL(t *testing.T) {
	tag := &link.AnchorTag{
		AbsoluteUrl: url.URL{},
		Status:      100, // Set to non-default value
	}

	ctx := context.Background()
	config := DefaultConfig()

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)
	if err != nil {
		t.Errorf("Expected no error for empty URL, got %v", err)
	}

	if tag.Status != -1 {
		t.Errorf("Expected status -1 for empty URL, got %d", tag.Status)
	}
}

func TestCheckLinks_Timeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	tag := &link.AnchorTag{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()
	config.Timeout = 100 * time.Millisecond // Very short timeout

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)

	// Should encounter an error due to timeout
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if tag.Status != -1 {
		t.Errorf("Expected status -1 after timeout, got %d", tag.Status)
	}
}

func TestCheckLinks_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	tags := make([]*link.AnchorTag, 10)
	for i := range tags {
		tags[i] = &link.AnchorTag{
			AbsoluteUrl: *parsedURL,
			Status:      -1,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	config := DefaultConfig()
	config.MaxWorkers = 2

	err := CheckAnchorTags(ctx, tags, config)

	// Should encounter context cancellation
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestCheckLinks_Redirects(t *testing.T) {
	// Create redirect chain
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer finalServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusMovedPermanently)
	}))
	defer redirectServer.Close()

	parsedURL, _ := url.Parse(redirectServer.URL)
	tag := &link.AnchorTag{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)
	if err != nil {
		t.Errorf("Expected no error for redirect, got %v", err)
	}

	// After following redirect, should get final status
	if tag.Status != http.StatusOK {
		t.Errorf("Expected status %d after redirect, got %d", http.StatusOK, tag.Status)
	}
}

func TestCheckLinks_UserAgent(t *testing.T) {
	userAgentReceived := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgentReceived = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	tag := &link.AnchorTag{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()
	customUA := "Custom-User-Agent/1.0"
	config.UserAgent = customUA

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if userAgentReceived != customUA {
		t.Errorf("Expected User-Agent '%s', got '%s'", customUA, userAgentReceived)
	}
}

func TestCheckSitemapUrls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	sitemapUrl := &link.SitemapUrl{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()

	err := CheckSitemapUrls(ctx, []*link.SitemapUrl{sitemapUrl}, config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if sitemapUrl.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, sitemapUrl.Status)
	}
}

func TestCheckLinks_InvalidURL(t *testing.T) {
	parsedURL, _ := url.Parse("http://invalid-domain-that-does-not-exist-12345.com")
	tag := &link.AnchorTag{
		AbsoluteUrl: *parsedURL,
		Status:      -1,
	}

	ctx := context.Background()
	config := DefaultConfig()
	config.Timeout = 1 * time.Second

	err := CheckAnchorTags(ctx, []*link.AnchorTag{tag}, config)

	// Should encounter an error for invalid domain
	if err == nil {
		t.Error("Expected error for invalid domain, got nil")
	}

	if tag.Status != -1 {
		t.Errorf("Expected status -1 for failed request, got %d", tag.Status)
	}
}
