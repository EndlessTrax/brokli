package parser

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(htmlNode)

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(htmlNode)

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		links := findLinks(htmlNode)

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(htmlNode, *baseUrl)

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(htmlNode, *baseUrl)

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
		htmlNode := createTestNode("html", []html.Attribute{}, body)

		results := GetPageResults(htmlNode, *baseUrl)

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
			_, _ = w.Write([]byte(testHTML))
		}))
		defer server.Close()

		baseUrl, _ := url.Parse(server.URL)

		// Step 1: Parse HTML (from fetcher in real use)
		doc, err := ParseHTML([]byte(testHTML))
		if err != nil {
			t.Errorf("ParseHTML failed: %v", err)
		}

		// Step 2: Get page results
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
