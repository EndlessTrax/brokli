package checker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/endlesstrax/brokli/pkg/link"
)

// Config holds configuration for the HTTP checker
type Config struct {
	// MaxWorkers is the number of concurrent HTTP requests
	MaxWorkers int
	// Timeout is the maximum time to wait for a single request
	Timeout time.Duration
	// UserAgent is the User-Agent header to send with requests
	UserAgent string
	// MaxRedirects is the maximum number of redirects to follow (default: 10)
	MaxRedirects int
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		MaxWorkers:   10,
		Timeout:      10 * time.Second,
		UserAgent:    "Brokli/0.1.0 (Broken Link Checker)",
		MaxRedirects: 10,
	}
}

// CheckableLink is an interface for types that can have their HTTP status checked
type CheckableLink interface {
	GetURL() string
	SetStatus(statusCode int)
}

// AnchorTagWrapper wraps link.AnchorTag to implement CheckableLink
type AnchorTagWrapper struct {
	*link.AnchorTag
}

func (a *AnchorTagWrapper) GetURL() string {
	return a.AbsoluteUrl.String()
}

func (a *AnchorTagWrapper) SetStatus(statusCode int) {
	a.Status = statusCode
}

// SitemapUrlWrapper wraps link.SitemapUrl to implement CheckableLink
type SitemapUrlWrapper struct {
	*link.SitemapUrl
}

func (s *SitemapUrlWrapper) GetURL() string {
	return s.AbsoluteUrl.String()
}

func (s *SitemapUrlWrapper) SetStatus(statusCode int) {
	s.Status = statusCode
}

// CheckLinks checks the HTTP status of multiple links concurrently
func CheckLinks(ctx context.Context, links []CheckableLink, config Config) error {
	if len(links) == 0 {
		return nil
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow up to MaxRedirects redirects
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		},
	}

	// Create channels for work distribution
	jobs := make(chan CheckableLink, len(links))
	results := make(chan error, len(links))

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := config.MaxWorkers
	if numWorkers > len(links) {
		numWorkers = len(links)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, client, jobs, results, &wg, config.UserAgent)
	}

	// Send jobs to workers
	for _, link := range links {
		jobs <- link
	}
	close(jobs)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during checking", len(errors))
	}

	return nil
}

// worker processes links from the jobs channel
func worker(ctx context.Context, client *http.Client, jobs <-chan CheckableLink, results chan<- error, wg *sync.WaitGroup, userAgent string) {
	defer wg.Done()

	for link := range jobs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
		}

		// Skip empty URLs
		url := link.GetURL()
		if url == "" {
			link.SetStatus(-1)
			results <- nil
			continue
		}

		// Make HEAD request to check status
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			link.SetStatus(-1)
			results <- fmt.Errorf("failed to create request for %s: %w", url, err)
			continue
		}

		// Set User-Agent header
		req.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(req)
		if err != nil {
			link.SetStatus(-1)
			results <- fmt.Errorf("failed to check %s: %w", url, err)
			continue
		}
		// Close body immediately - we only need the status code from HEAD request
		_ = resp.Body.Close()

		// Set the status code
		link.SetStatus(resp.StatusCode)
		results <- nil
	}
}

// CheckAnchorTags is a convenience function for checking anchor tags
func CheckAnchorTags(ctx context.Context, tags []*link.AnchorTag, config Config) error {
	links := make([]CheckableLink, len(tags))
	for i := range tags {
		links[i] = &AnchorTagWrapper{tags[i]}
	}
	return CheckLinks(ctx, links, config)
}

// CheckSitemapUrls is a convenience function for checking sitemap URLs
func CheckSitemapUrls(ctx context.Context, urls []*link.SitemapUrl, config Config) error {
	links := make([]CheckableLink, len(urls))
	for i := range urls {
		links[i] = &SitemapUrlWrapper{urls[i]}
	}
	return CheckLinks(ctx, links, config)
}
