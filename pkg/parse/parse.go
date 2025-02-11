package parse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"

	t "github.com/endlesstrax/brokli/pkg/types"
)

func GetPageResults(links []html.Node) t.PageResults {
	var pageResults t.PageResults

	for _, link := range links {
		var a t.AnchorTag
		err := a.New(&link)
		if err != nil {
			panic(err) // TODO: Handle this error better
		}
		pageResults.Links = append(pageResults.Links, a)
	}

	return pageResults
}

func GetHTML(url string) []byte {
	// Fetch the HTML from the URL
	resp, err := http.Get(url)
	if err != nil {
		panic(err) // TODO: Handle this error better
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err) // TODO: Handle this error better
	}

	return body
}

func FindAllLinks(doc *html.Node) []*html.Node {
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


func ParseHTML(b []byte) *html.Node {
	// Parse the HTML into a tree
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		panic(err) // TODO: Handle this error better
	}
	return doc
}


func TestLink(url string) int {
	resp, err := http.Get(url)
	if err != nil {
		panic(err) // TODO: Handle this error better
	}	
	
	return resp.StatusCode
}

