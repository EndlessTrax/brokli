package types

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
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

func TestExtractAttributes(t *testing.T) {
	t.Run("valid node with attributes", func(t *testing.T) {
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://example.com"},
			{Key: "class", Val: "link"},
			{Key: "id", Val: "main-link"},
		})

		attrs, err := extractAttributes(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := map[string]string{
			"href":  "https://example.com",
			"class": "link",
			"id":    "main-link",
		}

		if !reflect.DeepEqual(attrs, expected) {
			t.Errorf("Expected %v, got %v", expected, attrs)
		}
	})

	t.Run("node with empty attribute key", func(t *testing.T) {
		node := createTestNode("a", []html.Attribute{
			{Key: "", Val: "should-be-skipped"},
			{Key: "href", Val: "https://example.com"},
		})

		attrs, err := extractAttributes(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := map[string]string{
			"href": "https://example.com",
		}

		if !reflect.DeepEqual(attrs, expected) {
			t.Errorf("Expected %v, got %v", expected, attrs)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		_, err := extractAttributes(nil)
		if err == nil {
			t.Error("Expected error for nil node")
		}
		if !strings.Contains(err.Error(), "tag cannot be nil") {
			t.Errorf("Expected 'tag cannot be nil' error, got %v", err)
		}
	})

	t.Run("node with no attributes", func(t *testing.T) {
		node := createTestNode("a", []html.Attribute{})

		attrs, err := extractAttributes(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(attrs) != 0 {
			t.Errorf("Expected empty map, got %v", attrs)
		}
	})
}

func TestExtractText(t *testing.T) {
	t.Run("simple text node", func(t *testing.T) {
		textNode := createTextNode("Hello World")
		node := createTestNode("a", []html.Attribute{}, textNode)

		text, err := extractText(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if text != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", text)
		}
	})

	t.Run("nested text nodes", func(t *testing.T) {
		boldNode := createTestNode("b", []html.Attribute{}, createTextNode("Bold"))
		textNode := createTextNode(" and ")
		italicNode := createTestNode("i", []html.Attribute{}, createTextNode("Italic"))

		node := createTestNode("a", []html.Attribute{}, boldNode, textNode, italicNode)

		text, err := extractText(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "Bold and Italic"
		if text != expected {
			t.Errorf("Expected '%s', got '%s'", expected, text)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		_, err := extractText(nil)
		if err == nil {
			t.Error("Expected error for nil node")
		}
		if !strings.Contains(err.Error(), "tag cannot be nil") {
			t.Errorf("Expected 'tag cannot be nil' error, got %v", err)
		}
	})

	t.Run("node with no text", func(t *testing.T) {
		node := createTestNode("a", []html.Attribute{})

		text, err := extractText(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if text != "" {
			t.Errorf("Expected empty string, got '%s'", text)
		}
	})
}

func TestRenderRawTag(t *testing.T) {
	t.Run("simple anchor tag", func(t *testing.T) {
		textNode := createTextNode("Link Text")
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://example.com"},
		}, textNode)

		rawTag, err := renderRawTag(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := `<a href="https://example.com">Link Text</a>`
		if rawTag != expected {
			t.Errorf("Expected '%s', got '%s'", expected, rawTag)
		}
	})

	t.Run("tag with multiple attributes", func(t *testing.T) {
		textNode := createTextNode("Link")
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://example.com"},
			{Key: "class", Val: "link"},
			{Key: "id", Val: "main"},
		}, textNode)

		rawTag, err := renderRawTag(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that all attributes are present (order may vary)
		if !strings.Contains(rawTag, `href="https://example.com"`) {
			t.Error("Missing href attribute")
		}
		if !strings.Contains(rawTag, `class="link"`) {
			t.Error("Missing class attribute")
		}
		if !strings.Contains(rawTag, `id="main"`) {
			t.Error("Missing id attribute")
		}
		if !strings.HasPrefix(rawTag, "<a ") {
			t.Error("Should start with '<a '")
		}
		if !strings.HasSuffix(rawTag, ">Link</a>") {
			t.Error("Should end with '>Link</a>'")
		}
	})

	t.Run("nil node", func(t *testing.T) {
		_, err := renderRawTag(nil)
		if err == nil {
			t.Error("Expected error for nil node")
		}
		if !strings.Contains(err.Error(), "tag cannot be nil") {
			t.Errorf("Expected 'tag cannot be nil' error, got %v", err)
		}
	})

	t.Run("nested tags", func(t *testing.T) {
		boldNode := createTestNode("b", []html.Attribute{}, createTextNode("Bold"))
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://example.com"},
		}, boldNode)

		rawTag, err := renderRawTag(node)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := `<a href="https://example.com"><b>Bold</b></a>`
		if rawTag != expected {
			t.Errorf("Expected '%s', got '%s'", expected, rawTag)
		}
	})
}

func TestResolveAbsoluteUrl(t *testing.T) {
	baseUrl, _ := url.Parse("https://example.com/path/")

	t.Run("relative URL", func(t *testing.T) {
		attrs := map[string]string{"href": "page.html"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "https://example.com/path/page.html"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("absolute URL", func(t *testing.T) {
		attrs := map[string]string{"href": "https://other.com/page"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "https://other.com/page"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("empty href", func(t *testing.T) {
		attrs := map[string]string{"href": ""}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != baseUrl.String() {
			t.Errorf("Expected base URL '%s', got '%s'", baseUrl.String(), result.String())
		}
	})

	t.Run("mailto link", func(t *testing.T) {
		attrs := map[string]string{"href": "mailto:test@example.com"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for mailto, got '%s'", result.String())
		}
	})

	t.Run("javascript link", func(t *testing.T) {
		attrs := map[string]string{"href": "javascript:void(0)"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for javascript, got '%s'", result.String())
		}
	})

	t.Run("fragment link", func(t *testing.T) {
		attrs := map[string]string{"href": "#section1"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for fragment, got '%s'", result.String())
		}
	})

	t.Run("protocol-relative URL", func(t *testing.T) {
		attrs := map[string]string{"href": "//other.com/path"}

		result, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		expected := "http://other.com/path"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("nil attributes", func(t *testing.T) {
		_, err := resolveAbsoluteUrl(nil, *baseUrl)
		if err == nil {
			t.Error("Expected error for nil attributes")
		}
		if !strings.Contains(err.Error(), "attributes cannot be nil") {
			t.Errorf("Expected 'attributes cannot be nil' error, got %v", err)
		}
	})

	t.Run("invalid href", func(t *testing.T) {
		attrs := map[string]string{"href": "://invalid-url"}

		_, err := resolveAbsoluteUrl(attrs, *baseUrl)
		if err == nil {
			t.Error("Expected error for invalid href")
		}
		if !strings.Contains(err.Error(), "failed to parse href") {
			t.Errorf("Expected 'failed to parse href' error, got %v", err)
		}
	})
}

func TestNewAnchorTag(t *testing.T) {
	baseUrl, _ := url.Parse("https://example.com/")

	t.Run("complete anchor tag", func(t *testing.T) {
		textNode := createTextNode("Google")
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "https://www.google.com"},
			{Key: "class", Val: "external"},
		}, textNode)

		anchorTag, err := NewAnchorTag(node, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check all fields
		if anchorTag.AbsoluteUrl.String() != "https://www.google.com" {
			t.Errorf("Expected AbsoluteUrl 'https://www.google.com', got '%s'", anchorTag.AbsoluteUrl.String())
		}

		if anchorTag.Attributes["href"] != "https://www.google.com" {
			t.Errorf("Expected href attribute 'https://www.google.com', got '%s'", anchorTag.Attributes["href"])
		}

		if anchorTag.Attributes["class"] != "external" {
			t.Errorf("Expected class attribute 'external', got '%s'", anchorTag.Attributes["class"])
		}

		if anchorTag.Text != "Google" {
			t.Errorf("Expected text 'Google', got '%s'", anchorTag.Text)
		}

		if !strings.Contains(anchorTag.RawTag, `href="https://www.google.com"`) ||
			!strings.Contains(anchorTag.RawTag, `class="external"`) ||
			!strings.Contains(anchorTag.RawTag, ">Google</a>") {
			t.Errorf("RawTag doesn't contain expected elements: %s", anchorTag.RawTag)
		}

		if anchorTag.Status != -1 {
			t.Errorf("Expected Status -1, got %d", anchorTag.Status)
		}
	})

	t.Run("relative link", func(t *testing.T) {
		textNode := createTextNode("About")
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "/about"},
		}, textNode)

		anchorTag, err := NewAnchorTag(node, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedUrl := "https://example.com/about"
		if anchorTag.AbsoluteUrl.String() != expectedUrl {
			t.Errorf("Expected AbsoluteUrl '%s', got '%s'", expectedUrl, anchorTag.AbsoluteUrl.String())
		}
	})

	t.Run("mailto link", func(t *testing.T) {
		textNode := createTextNode("Contact")
		node := createTestNode("a", []html.Attribute{
			{Key: "href", Val: "mailto:test@example.com"},
		}, textNode)

		anchorTag, err := NewAnchorTag(node, *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if anchorTag.AbsoluteUrl.String() != "" {
			t.Errorf("Expected empty AbsoluteUrl for mailto, got '%s'", anchorTag.AbsoluteUrl.String())
		}
	})

	t.Run("nil node", func(t *testing.T) {
		_, err := NewAnchorTag(nil, *baseUrl)
		if err == nil {
			t.Error("Expected error for nil node")
		}
		if !strings.Contains(err.Error(), "tag cannot be nil") {
			t.Errorf("Expected 'tag cannot be nil' error, got %v", err)
		}
	})
}

func TestNewSitemapUrl(t *testing.T) {
	t.Run("valid sitemap URL with all metadata", func(t *testing.T) {
		urlStr := "https://www.example.com/"
		lastMod := "2023-01-01"
		changeFreq := "daily"
		priority := "1.0"

		sitemapUrl, err := NewSitemapUrl(urlStr, lastMod, changeFreq, priority)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if sitemapUrl.AbsoluteUrl.String() != urlStr {
			t.Errorf("Expected AbsoluteUrl to be '%s', got '%s'", urlStr, sitemapUrl.AbsoluteUrl.String())
		}
		if sitemapUrl.LastMod != lastMod {
			t.Errorf("Expected LastMod to be '%s', got '%s'", lastMod, sitemapUrl.LastMod)
		}
		if sitemapUrl.ChangeFreq != changeFreq {
			t.Errorf("Expected ChangeFreq to be '%s', got '%s'", changeFreq, sitemapUrl.ChangeFreq)
		}
		if sitemapUrl.Priority != priority {
			t.Errorf("Expected Priority to be '%s', got '%s'", priority, sitemapUrl.Priority)
		}
		if sitemapUrl.Status != -1 {
			t.Errorf("Expected Status -1, got %d", sitemapUrl.Status)
		}
	})

	t.Run("valid sitemap URL with minimal metadata", func(t *testing.T) {
		urlStr := "https://www.example.com/about"

		sitemapUrl, err := NewSitemapUrl(urlStr, "", "", "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if sitemapUrl.AbsoluteUrl.String() != urlStr {
			t.Errorf("Expected AbsoluteUrl to be '%s', got '%s'", urlStr, sitemapUrl.AbsoluteUrl.String())
		}
		if sitemapUrl.LastMod != "" {
			t.Errorf("Expected empty LastMod, got '%s'", sitemapUrl.LastMod)
		}
		if sitemapUrl.ChangeFreq != "" {
			t.Errorf("Expected empty ChangeFreq, got '%s'", sitemapUrl.ChangeFreq)
		}
		if sitemapUrl.Priority != "" {
			t.Errorf("Expected empty Priority, got '%s'", sitemapUrl.Priority)
		}
	})

	t.Run("empty URL string", func(t *testing.T) {
		_, err := NewSitemapUrl("", "2023-01-01", "daily", "1.0")
		if err == nil {
			t.Error("Expected error for empty URL string")
		}
		if !strings.Contains(err.Error(), "URL string cannot be empty") {
			t.Errorf("Expected 'URL string cannot be empty' error, got %v", err)
		}
	})

	t.Run("invalid URL string", func(t *testing.T) {
		_, err := NewSitemapUrl("://invalid-url", "2023-01-01", "daily", "1.0")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if !strings.Contains(err.Error(), "failed to parse URL") {
			t.Errorf("Expected 'failed to parse URL' error, got %v", err)
		}
	})

	t.Run("relative URL", func(t *testing.T) {
		urlStr := "/about"

		sitemapUrl, err := NewSitemapUrl(urlStr, "", "", "")
		if err != nil {
			t.Errorf("Expected no error for relative URL, got %v", err)
		}

		if sitemapUrl.AbsoluteUrl.String() != urlStr {
			t.Errorf("Expected AbsoluteUrl to be '%s', got '%s'", urlStr, sitemapUrl.AbsoluteUrl.String())
		}
	})
}

func TestSitemapUrlGetHttpStatus(t *testing.T) {
	t.Run("successful HTTP request", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		sitemapUrl, err := NewSitemapUrl(server.URL, "", "", "")
		if err != nil {
			t.Errorf("Expected no error creating SitemapUrl, got %v", err)
		}

		err = sitemapUrl.GetHttpStatus()
		if err != nil {
			t.Errorf("Expected no error getting HTTP status, got %v", err)
		}

		if sitemapUrl.Status != http.StatusOK {
			t.Errorf("Expected Status %d, got %d", http.StatusOK, sitemapUrl.Status)
		}
	})

	t.Run("empty URL", func(t *testing.T) {
		// Create SitemapUrl with empty URL by manipulating the struct directly
		sitemapUrl := SitemapUrl{
			AbsoluteUrl: url.URL{},
			Status:      -1,
		}

		err := sitemapUrl.GetHttpStatus()
		if err != nil {
			t.Errorf("Expected no error for empty URL, got %v", err)
		}

		// Status should remain unchanged
		if sitemapUrl.Status != -1 {
			t.Errorf("Expected Status -1 for empty URL, got %d", sitemapUrl.Status)
		}
	})

	t.Run("network error", func(t *testing.T) {
		sitemapUrl, err := NewSitemapUrl("http://localhost:99999/nonexistent", "", "", "")
		if err != nil {
			t.Errorf("Expected no error creating SitemapUrl, got %v", err)
		}

		err = sitemapUrl.GetHttpStatus()
		if err == nil {
			t.Error("Expected error for network error")
		}

		// Status should remain unchanged on error
		if sitemapUrl.Status != -1 {
			t.Errorf("Expected Status -1 after error, got %d", sitemapUrl.Status)
		}
	})
}
