package parse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	t "github.com/endlesstrax/brokli/pkg/types"
)

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

// GetPageResults extracts all anchor tags from the provided HTML content and returns them as PageResults
func GetPageResults(htmlContent *html.Node, baseUrl url.URL) t.PageResults {
	var pageResults t.PageResults

	links := findLinks(htmlContent)

	for _, link := range links {
		a, err := t.NewAnchorTag(link, baseUrl)
		if err != nil {
			fmt.Println("Skipping link due to error:", err)
			continue
		}
		pageResults.Links = append(pageResults.Links, a)
	}

	return pageResults
}

// GetHTML fetches the HTML content from the specified URL and returns it as a byte slice
func GetHTML(url string) ([]byte, error) {
	// Fetch the HTML from the URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// ParseHTML parses the given byte slice into an HTML node tree
func ParseHTML(b []byte) (*html.Node, error) {
	// Parse the HTML into a tree
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

