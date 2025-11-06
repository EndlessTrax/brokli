package parser

import (
	"bytes"
	"fmt"
	"net/url"

	"golang.org/x/net/html"

	"github.com/endlesstrax/brokli/pkg/link"
	"github.com/endlesstrax/brokli/pkg/resolver"
)

// ParseHTML parses the given byte slice into an HTML node tree
func ParseHTML(b []byte) (*html.Node, error) {
	// Parse the HTML into a tree
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// GetPageResults extracts all anchor tags from the provided HTML content and returns them as PageResults
func GetPageResults(htmlContent *html.Node, baseUrl url.URL) link.PageResults {
	var pageResults link.PageResults

	links := findLinks(htmlContent)

	for _, linkNode := range links {
		anchorTag, err := newAnchorTag(linkNode, baseUrl)
		if err != nil {
			fmt.Println("Skipping link due to error:", err)
			continue
		}
		pageResults.Links = append(pageResults.Links, anchorTag)
	}

	return pageResults
}

// findLinks recursively finds all anchor tags in the HTML document
func findLinks(doc *html.Node) []*html.Node {
	var links []*html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			links = append(links, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	fmt.Println(links)
	return links
}

// newAnchorTag creates a new AnchorTag from an HTML node
func newAnchorTag(tag *html.Node, baseUrl url.URL) (link.AnchorTag, error) {
	if tag == nil {
		return link.AnchorTag{}, fmt.Errorf("tag cannot be nil")
	}

	// Extract attributes
	attributes, err := extractAttributes(tag)
	if err != nil {
		return link.AnchorTag{}, fmt.Errorf("failed to extract attributes: %w", err)
	}

	// Extract text content
	text, err := extractText(tag)
	if err != nil {
		return link.AnchorTag{}, fmt.Errorf("failed to extract text: %w", err)
	}

	// Render raw tag
	rawTag, err := renderRawTag(tag)
	if err != nil {
		return link.AnchorTag{}, fmt.Errorf("failed to render raw tag: %w", err)
	}

	// Resolve absolute URL
	href := attributes["href"]
	absoluteUrl, err := resolver.ResolveAbsoluteUrl(href, baseUrl)
	if err != nil {
		return link.AnchorTag{}, fmt.Errorf("failed to resolve absolute URL: %w", err)
	}

	// Create and return the AnchorTag
	anchorTag := link.AnchorTag{
		Attributes:  attributes,
		Text:        text,
		RawTag:      rawTag,
		AbsoluteUrl: absoluteUrl,
		Status:      -1, // Initialize to -1 (unknown)
	}

	return anchorTag, nil
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
