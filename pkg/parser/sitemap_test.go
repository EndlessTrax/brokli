package parser

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSitemap(t *testing.T) {
	t.Run("valid sitemap XML", func(t *testing.T) {
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://www.example.com/</loc>
      <lastmod>2023-01-01</lastmod>
      <changefreq>daily</changefreq>
      <priority>1.0</priority>
   </url>
   <url>
      <loc>https://www.example.com/about</loc>
      <lastmod>2023-01-01</lastmod>
      <changefreq>weekly</changefreq>
      <priority>0.8</priority>
   </url>
</urlset>`

		sitemap, err := ParseSitemap([]byte(xmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing valid sitemap, got %v", err)
		}

		if sitemap == nil {
			t.Error("Expected non-nil sitemap")
		} else if len(sitemap.URLs) != 2 {
			t.Errorf("Expected 2 URLs in sitemap, got %d", len(sitemap.URLs))
		}

		// Check first URL
		firstUrl := sitemap.URLs[0]
		if firstUrl.Loc != "https://www.example.com/" {
			t.Errorf("Expected first URL to be 'https://www.example.com/', got '%s'", firstUrl.Loc)
		}
		if firstUrl.LastMod != "2023-01-01" {
			t.Errorf("Expected first URL lastmod to be '2023-01-01', got '%s'", firstUrl.LastMod)
		}
		if firstUrl.ChangeFreq != "daily" {
			t.Errorf("Expected first URL changefreq to be 'daily', got '%s'", firstUrl.ChangeFreq)
		}
		if firstUrl.Priority != "1.0" {
			t.Errorf("Expected first URL priority to be '1.0', got '%s'", firstUrl.Priority)
		}
	})

	t.Run("minimal sitemap XML", func(t *testing.T) {
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://www.example.com/</loc>
   </url>
</urlset>`

		sitemap, err := ParseSitemap([]byte(xmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing minimal sitemap, got %v", err)
		}

		if len(sitemap.URLs) != 1 {
			t.Errorf("Expected 1 URL in sitemap, got %d", len(sitemap.URLs))
		}

		// Optional fields should be empty
		firstUrl := sitemap.URLs[0]
		if firstUrl.LastMod != "" {
			t.Errorf("Expected empty lastmod, got '%s'", firstUrl.LastMod)
		}
		if firstUrl.ChangeFreq != "" {
			t.Errorf("Expected empty changefreq, got '%s'", firstUrl.ChangeFreq)
		}
		if firstUrl.Priority != "" {
			t.Errorf("Expected empty priority, got '%s'", firstUrl.Priority)
		}
	})

	t.Run("empty sitemap XML", func(t *testing.T) {
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
</urlset>`

		sitemap, err := ParseSitemap([]byte(xmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing empty sitemap, got %v", err)
		}

		if len(sitemap.URLs) != 0 {
			t.Errorf("Expected 0 URLs in empty sitemap, got %d", len(sitemap.URLs))
		}
	})

	t.Run("malformed XML", func(t *testing.T) {
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://www.example.com/</loc>
   <url>` // Missing closing tags

		_, err := ParseSitemap([]byte(xmlContent))
		if err == nil {
			t.Error("Expected error for malformed XML")
		}
	})

	t.Run("invalid XML", func(t *testing.T) {
		xmlContent := `This is not XML`

		_, err := ParseSitemap([]byte(xmlContent))
		if err == nil {
			t.Error("Expected error for invalid XML")
		}
	})
}

func TestGetSitemapResults(t *testing.T) {
	t.Run("sitemap with multiple URLs", func(t *testing.T) {
		sitemap := &XMLSitemap{
			URLs: []XMLURL{
				{
					Loc:        "https://www.example.com/",
					LastMod:    "2023-01-01",
					ChangeFreq: "daily",
					Priority:   "1.0",
				},
				{
					Loc:        "https://www.example.com/about",
					LastMod:    "2023-01-01",
					ChangeFreq: "weekly",
					Priority:   "0.8",
				},
			},
		}

		results := GetSitemapResults(sitemap, "https://example.com/sitemap.xml")

		if len(results.Urls) != 2 {
			t.Errorf("Expected 2 URLs in results, got %d", len(results.Urls))
		}

		if results.SourceUrl != "https://example.com/sitemap.xml" {
			t.Errorf("Expected SourceUrl to be 'https://example.com/sitemap.xml', got '%s'", results.SourceUrl)
		}

		// Check first URL
		firstUrl := results.Urls[0]
		if firstUrl.AbsoluteUrl.String() != "https://www.example.com/" {
			t.Errorf("Expected first URL to be 'https://www.example.com/', got '%s'", firstUrl.AbsoluteUrl.String())
		}
		if firstUrl.LastMod != "2023-01-01" {
			t.Errorf("Expected first URL lastmod to be '2023-01-01', got '%s'", firstUrl.LastMod)
		}
		if firstUrl.Status != -1 {
			t.Errorf("Expected Status -1, got %d", firstUrl.Status)
		}
	})

	t.Run("empty sitemap", func(t *testing.T) {
		sitemap := &XMLSitemap{
			URLs: []XMLURL{},
		}

		results := GetSitemapResults(sitemap, "https://example.com/sitemap.xml")

		if len(results.Urls) != 0 {
			t.Errorf("Expected 0 URLs in empty sitemap results, got %d", len(results.Urls))
		}
	})

	t.Run("sitemap with invalid URL", func(t *testing.T) {
		sitemap := &XMLSitemap{
			URLs: []XMLURL{
				{
					Loc:        "://invalid-url",
					LastMod:    "2023-01-01",
					ChangeFreq: "daily",
					Priority:   "1.0",
				},
			},
		}

		results := GetSitemapResults(sitemap, "https://example.com/sitemap.xml")

		// Invalid URL should be skipped
		if len(results.Urls) != 0 {
			t.Errorf("Expected 0 URLs due to invalid URL being skipped, got %d", len(results.Urls))
		}
	})
}

// Integration test for sitemap functionality
func TestSitemapIntegration(t *testing.T) {
	t.Run("complete sitemap processing", func(t *testing.T) {
		// Create a test server with sitemap XML
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://www.example.com/</loc>
      <lastmod>2023-01-01</lastmod>
      <changefreq>daily</changefreq>
      <priority>1.0</priority>
   </url>
   <url>
      <loc>https://www.example.com/about</loc>
      <lastmod>2023-01-01</lastmod>
      <changefreq>weekly</changefreq>
      <priority>0.8</priority>
   </url>
</urlset>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(xmlContent))
		}))
		defer server.Close()

		// Step 1: Parse sitemap XML (from fetcher in real use)
		sitemap, err := ParseSitemap([]byte(xmlContent))
		if err != nil {
			t.Errorf("ParseSitemap failed: %v", err)
		}

		// Step 2: Get sitemap results
		results := GetSitemapResults(sitemap, server.URL)

		// Verify results
		if len(results.Urls) != 2 {
			t.Errorf("Expected 2 URLs, got %d", len(results.Urls))
		}

		// Check that we have the expected URLs
		expectedUrls := []string{"https://www.example.com/", "https://www.example.com/about"}
		for i, expectedUrl := range expectedUrls {
			if i < len(results.Urls) && results.Urls[i].AbsoluteUrl.String() != expectedUrl {
				t.Errorf("Expected URL %d to be '%s', got '%s'", i, expectedUrl, results.Urls[i].AbsoluteUrl.String())
			}
		}

		// Check metadata
		if results.Urls[0].Priority != "1.0" {
			t.Errorf("Expected first URL priority to be '1.0', got '%s'", results.Urls[0].Priority)
		}
		if results.Urls[1].ChangeFreq != "weekly" {
			t.Errorf("Expected second URL changefreq to be 'weekly', got '%s'", results.Urls[1].ChangeFreq)
		}
	})
}
