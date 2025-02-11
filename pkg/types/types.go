package types

import (
	"golang.org/x/net/html"
)


type PageResults struct {
	Links []AnchorTag
	SourceUrl string
}

type AnchorTag struct {
	RawLink string
	Href	string
	Text	string
	Status	int
	// Attributes map[string]string
}

func (a *AnchorTag) New(tag *html.Node) error {
	// Takes an anchor tag and marshalls it to a AnchorTag struct

	for _, attr := range tag.Attr {
		if attr.Key == "href" {
			a.Href = attr.Val
		}
	}

	// TODO: Parse the tag and set the Href, Text, and Status fields
	
	return nil
}