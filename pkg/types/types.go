package types

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type PageResults struct {
	Links     []AnchorTag
	SourceUrl string
}

type SitemapResults struct {
	Urls      []SitemapUrl
	SourceUrl string
}

type AnchorTag struct {
	AbsoluteUrl url.URL
	Attributes  map[string]string
	RawTag      string
	Status      int
	Text        string
}

type SitemapUrl struct {
	AbsoluteUrl url.URL
	LastMod     string
	ChangeFreq  string
	Priority    string
	Status      int
}

// extractAttributes parses HTML attributes from a node and returns them as a map
func extractAttributes(tag *html.Node) (map[string]string, error) {
	if tag == nil {
		return nil, fmt.Errorf("tag cannot be nil")
	}

	attributes := make(map[string]string)
	for _, attr := range tag.Attr {
		if attr.Key == "" {
			continue // Skip empty attribute keys
		}
		attributes[attr.Key] = attr.Val
	}

	return attributes, nil
}

// extractText recursively extracts all text content from an HTML node
func extractText(tag *html.Node) (string, error) {
	if tag == nil {
		return "", fmt.Errorf("tag cannot be nil")
	}

	var text string
	var extractTextRecursive func(*html.Node)
	extractTextRecursive = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTextRecursive(c)
		}
	}

	extractTextRecursive(tag)
	return text, nil
}

// renderRawTag converts an HTML node back to its string representation
func renderRawTag(tag *html.Node) (string, error) {
	if tag == nil {
		return "", fmt.Errorf("tag cannot be nil")
	}

	var renderNode func(*html.Node) string
	renderNode = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return n.Data
		}
		result := "<" + n.Data
		for _, attr := range n.Attr {
			result += " " + attr.Key + `="` + attr.Val + `"`
		}
		result += ">"
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result += renderNode(c)
		}
		result += "</" + n.Data + ">"
		return result
	}

	return renderNode(tag), nil
}

// resolveAbsoluteUrl resolves the href attribute to an absolute URL
func resolveAbsoluteUrl(attributes map[string]string, baseUrl url.URL) (url.URL, error) {
	if attributes == nil {
		return url.URL{}, fmt.Errorf("attributes cannot be nil")
	}

	href := attributes["href"]

	// If href is empty, return the base URL
	if href == "" {
		return baseUrl, nil
	}

	// Handle special schemes that shouldn't be resolved
	if len(href) > 0 {
		// Check for mailto:, javascript:, or fragment links
		if (len(href) >= 7 && href[:7] == "mailto:") ||
			(len(href) >= 11 && href[:11] == "javascript:") ||
			(len(href) >= 1 && href[0] == '#') {
			return url.URL{}, nil
		}
	}

	// Handle protocol-relative URLs (//example.com)
	if len(href) > 1 && href[:2] == "//" {
		// Parse the URL without protocol to properly handle host and path
		protocolRelativeUrl, err := url.Parse("http:" + href)
		if err != nil {
			return url.URL{}, fmt.Errorf("failed to parse protocol-relative href '%s': %w", href, err)
		}
		return *protocolRelativeUrl, nil
	}

	// Parse the href to handle it properly
	hrefUrl, err := url.Parse(href)
	if err != nil {
		return url.URL{}, fmt.Errorf("failed to parse href '%s': %w", href, err)
	}

	// Resolve the href relative to the base URL
	resolvedUrl := baseUrl.ResolveReference(hrefUrl)
	return *resolvedUrl, nil
}

func NewAnchorTag(tag *html.Node, baseUrl url.URL) (AnchorTag, error) {
	if tag == nil {
		return AnchorTag{}, fmt.Errorf("tag cannot be nil")
	}

	// Extract attributes
	attributes, err := extractAttributes(tag)
	if err != nil {
		return AnchorTag{}, fmt.Errorf("failed to extract attributes: %w", err)
	}

	// Extract text content
	text, err := extractText(tag)
	if err != nil {
		return AnchorTag{}, fmt.Errorf("failed to extract text: %w", err)
	}

	// Render raw tag
	rawTag, err := renderRawTag(tag)
	if err != nil {
		return AnchorTag{}, fmt.Errorf("failed to render raw tag: %w", err)
	}

	// Resolve absolute URL
	absoluteUrl, err := resolveAbsoluteUrl(attributes, baseUrl)
	if err != nil {
		return AnchorTag{}, fmt.Errorf("failed to resolve absolute URL: %w", err)
	}

	// Create and return the AnchorTag
	a := AnchorTag{
		Attributes:  attributes,
		Text:        text,
		RawTag:      rawTag,
		AbsoluteUrl: absoluteUrl,
		Status:      -1, // Initialize to -1 (unknown)
	}

	return a, nil
}

// Gets the HTTP status code of the AnchorTag
func (a *AnchorTag) GetHttpStatus() error {

	if a.AbsoluteUrl.String() == "" {
		return nil
	}

	resp, err := http.Head(a.AbsoluteUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	a.Status = resp.StatusCode

	return nil
}

// NewSitemapUrl creates a new SitemapUrl from the provided URL string and metadata
func NewSitemapUrl(urlStr, lastMod, changeFreq, priority string) (SitemapUrl, error) {
	if urlStr == "" {
		return SitemapUrl{}, fmt.Errorf("URL string cannot be empty")
	}

	// Parse the URL
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return SitemapUrl{}, fmt.Errorf("failed to parse URL '%s': %w", urlStr, err)
	}

	// Create and return the SitemapUrl
	s := SitemapUrl{
		AbsoluteUrl: *parsedUrl,
		LastMod:     lastMod,
		ChangeFreq:  changeFreq,
		Priority:    priority,
		Status:      -1, // Initialize to -1 (unknown)
	}

	return s, nil
}

// Gets the HTTP status code of the SitemapUrl
func (s *SitemapUrl) GetHttpStatus() error {
	if s.AbsoluteUrl.String() == "" {
		return nil
	}

	resp, err := http.Head(s.AbsoluteUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	s.Status = resp.StatusCode

	return nil
}
