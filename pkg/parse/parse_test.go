package parse

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/html"
)

// Helper function to create HTML nodes for testing
func createTestNode(tagName string, attributes []html.Attribute, children ...*html.Node) *html.Node {
	node := &html.Node{
		Type: html.ElementNode,
		Data: tagName,
		Attr: attributes,
	}

	for _, child := range children {
		node.AppendChild(child)
	}

	return node
}

func createTextNode(text string) *html.Node {
	return &html.Node{
		Type: html.TextNode,
		Data: text,
	}
}

func TestFindLinks(t *testing.T) {
	t.Run("document with multiple anchor tags", func(t *testing.T) {
		// Create a test HTML structure with multiple anchor tags
		link1 := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://example.com"},
		}, createTextNode("Link 1"))

		link2 := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "/about"},
		}, createTextNode("About"))

		div := createTestNode("div", []html.Attribute{}, link1)
		p := createTestNode("p", []html.Attribute{}, link2)

		body := createTestNode("body", []html.Attribute{}, div, p)
		html := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(html)

		if len(links) != 2 {
			t.Errorf("Expected 2 links, got %d", len(links))
		}

		// Verify the links are correct
		if links[0].Data != "a" {
			t.Errorf("Expected first link to be an 'a' tag, got '%s'", links[0].Data)
		}
		if links[1].Data != "a" {
			t.Errorf("Expected second link to be an 'a' tag, got '%s'", links[1].Data)
		}
	})

	t.Run("document with no anchor tags", func(t *testing.T) {
		div := createTestNode("div", []html.Attribute{}, createTextNode("No links here"))
		body := createTestNode("body", []html.Attribute{}, div)
		html := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(html)

		if len(links) != 0 {
			t.Errorf("Expected 0 links, got %d", len(links))
		}
	})

	t.Run("nested anchor tags", func(t *testing.T) {
		innerLink := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "#inner"},
		}, createTextNode("Inner"))

		outerLink := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "#outer"},
		}, createTextNode("Outer"), innerLink)

		body := createTestNode("body", []html.Attribute{}, outerLink)
		html := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(html)

		if len(links) != 2 {
			t.Errorf("Expected 2 links (nested), got %d", len(links))
		}
	})

	t.Run("nil document", func(t *testing.T) {
		// The function panics on nil input, which is expected behavior
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil document")
			}
		}()

		findLinks(nil)
	})
}

func TestGetPageResults(t *testing.T) {
	baseUrl, _ := url.Parse("https://example.com/")

	t.Run("document with valid links", func(t *testing.T) {
		link1 := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://google.com"},
		}, createTextNode("Google"))

		link2 := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "/about"},
		}, createTextNode("About"))

		body := createTestNode("body", []html.Attribute{}, link1, link2)
		html := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(html, *baseUrl)

		if len(results.Links) != 2 {
			t.Errorf("Expected 2 anchor tags, got %d", len(results.Links))
		}

		// Check first link
		if results.Links[0].AbsoluteUrl.String() != "https://google.com" {
			t.Errorf("Expected first link URL to be 'https://google.com', got '%s'", results.Links[0].AbsoluteUrl.String())
		}
		if results.Links[0].Text != "Google" {
			t.Errorf("Expected first link text to be 'Google', got '%s'", results.Links[0].Text)
		}

		// Check second link (relative URL resolution)
		if results.Links[1].AbsoluteUrl.String() != "https://example.com/about" {
			t.Errorf("Expected second link URL to be 'https://example.com/about', got '%s'", results.Links[1].AbsoluteUrl.String())
		}
		if results.Links[1].Text != "About" {
			t.Errorf("Expected second link text to be 'About', got '%s'", results.Links[1].Text)
		}
	})

	t.Run("document with no links", func(t *testing.T) {
		div := createTestNode("div", []html.Attribute{}, createTextNode("No links"))
		body := createTestNode("body", []html.Attribute{}, div)
		html := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(html, *baseUrl)

		if len(results.Links) != 0 {
			t.Errorf("Expected 0 anchor tags, got %d", len(results.Links))
		}
	})

	t.Run("document with special links", func(t *testing.T) {
		mailtoLink := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "mailto:test@example.com"},
		}, createTextNode("Email"))

		jsLink := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "javascript:void(0)"},
		}, createTextNode("JS Link"))

		fragmentLink := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "#section1"},
		}, createTextNode("Section"))

		body := createTestNode("body", []html.Attribute{}, mailtoLink, jsLink, fragmentLink)
		html := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(html, *baseUrl)

		if len(results.Links) != 3 {
			t.Errorf("Expected 3 anchor tags, got %d", len(results.Links))
		}

		// All special links should have empty AbsoluteUrl
		for i, link := range results.Links {
			if link.AbsoluteUrl.String() != "" {
				t.Errorf("Expected empty AbsoluteUrl for special link %d, got '%s'", i, link.AbsoluteUrl.String())
			}
		}
	})

	t.Run("nil document", func(t *testing.T) {
		// GetPageResults should handle nil input gracefully or panic consistently
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil document")
			}
		}()

		GetPageResults(nil, *baseUrl)
	})
}

func TestParseHTML(t *testing.T) {
	t.Run("valid HTML", func(t *testing.T) {
		htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
	<div>
		<a href="https://example.com">Link</a>
		<p>Some text</p>
	</div>
</body>
</html>`

		doc, err := ParseHTML([]byte(htmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing valid HTML, got %v", err)
		}

		if doc == nil {
			t.Error("Expected non-nil document")
		}

		// Verify we can find elements in the parsed document
		links := findLinks(doc)
		if len(links) != 1 {
			t.Errorf("Expected 1 link in parsed document, got %d", len(links))
		}
	})

	t.Run("minimal HTML", func(t *testing.T) {
		htmlContent := `<html><body><a href="/test">Test</a></body></html>`

		doc, err := ParseHTML([]byte(htmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing minimal HTML, got %v", err)
		}

		if doc == nil {
			t.Error("Expected non-nil document")
		}

		links := findLinks(doc)
		if len(links) != 1 {
			t.Errorf("Expected 1 link in parsed document, got %d", len(links))
		}
	})

	t.Run("malformed HTML", func(t *testing.T) {
		htmlContent := `<html><body><a href="/test">Unclosed link<p>Other content</html>`

		// html.Parse is very tolerant and should still parse this
		doc, err := ParseHTML([]byte(htmlContent))
		if err != nil {
			t.Errorf("Expected no error parsing malformed HTML (html.Parse is tolerant), got %v", err)
		}

		if doc == nil {
			t.Error("Expected non-nil document even for malformed HTML")
		}
	})

	t.Run("empty HTML", func(t *testing.T) {
		doc, err := ParseHTML([]byte(""))
		if err != nil {
			t.Errorf("Expected no error parsing empty HTML, got %v", err)
		}

		if doc == nil {
			t.Error("Expected non-nil document for empty HTML")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		doc, err := ParseHTML(nil)
		if err != nil {
			t.Errorf("Expected no error parsing nil input, got %v", err)
		}

		if doc == nil {
			t.Error("Expected non-nil document for nil input")
		}
	})
}

func TestGetHTML(t *testing.T) {
	t.Run("successful HTTP request", func(t *testing.T) {
		// Create a test server
		testHTML := `<html><body><a href="/test">Test Link</a></body></html>`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testHTML))
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
			w.Write([]byte("Not Found"))
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

// Integration test that combines multiple functions
func TestIntegration(t *testing.T) {
	t.Run("full pipeline test", func(t *testing.T) {
		// Create a test server with HTML content
		testHTML := `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<div>
		<a href="https://google.com">Google</a>
		<a href="/about">About Us</a>
		<a href="mailto:test@example.com">Contact</a>
	</div>
</body>
</html>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testHTML))
		}))
		defer server.Close()

		baseUrl, _ := url.Parse(server.URL)

		// Step 1: Get HTML
		htmlBytes, err := GetHTML(server.URL)
		if err != nil {
			t.Errorf("GetHTML failed: %v", err)
		}

		// Step 2: Parse HTML
		doc, err := ParseHTML(htmlBytes)
		if err != nil {
			t.Errorf("ParseHTML failed: %v", err)
		}

		// Step 3: Get page results
		results := GetPageResults(doc, *baseUrl)

		// Verify results
		if len(results.Links) != 3 {
			t.Errorf("Expected 3 links, got %d", len(results.Links))
		}

		// Check that we have the expected links
		expectedTexts := []string{"Google", "About Us", "Contact"}
		for i, expectedText := range expectedTexts {
			if i < len(results.Links) && results.Links[i].Text != expectedText {
				t.Errorf("Expected link %d text '%s', got '%s'", i, expectedText, results.Links[i].Text)
			}
		}

		// Check URL resolution
		if results.Links[0].AbsoluteUrl.String() != "https://google.com" {
			t.Errorf("Expected first link to be 'https://google.com', got '%s'", results.Links[0].AbsoluteUrl.String())
		}

		expectedAboutUrl := server.URL + "/about"
		if results.Links[1].AbsoluteUrl.String() != expectedAboutUrl {
			t.Errorf("Expected second link to be '%s', got '%s'", expectedAboutUrl, results.Links[1].AbsoluteUrl.String())
		}

		// mailto link should have empty URL
		if results.Links[2].AbsoluteUrl.String() != "" {
			t.Errorf("Expected mailto link to have empty URL, got '%s'", results.Links[2].AbsoluteUrl.String())
		}
	})
}

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
		}

		if len(sitemap.URLs) != 2 {
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
			w.Write([]byte(xmlContent))
		}))
		defer server.Close()

		// Step 1: Get XML content
		xmlBytes, err := GetHTML(server.URL)
		if err != nil {
			t.Errorf("GetHTML failed: %v", err)
		}

		// Step 2: Parse sitemap XML
		sitemap, err := ParseSitemap(xmlBytes)
		if err != nil {
			t.Errorf("ParseSitemap failed: %v", err)
		}

		// Step 3: Get sitemap results
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
