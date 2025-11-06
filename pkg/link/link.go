package link

import "net/url"

// PageResults contains all links extracted from a single web page
type PageResults struct {
	Links     []AnchorTag
	SourceUrl string
}

// SitemapResults contains all URLs extracted from a sitemap
type SitemapResults struct {
	Urls      []SitemapUrl
	SourceUrl string
}

// AnchorTag represents an HTML anchor tag with its metadata
type AnchorTag struct {
	AbsoluteUrl url.URL
	Attributes  map[string]string
	RawTag      string
	Status      int
	Text        string
}

// SitemapUrl represents a URL entry from a sitemap with its metadata
type SitemapUrl struct {
	AbsoluteUrl url.URL
	LastMod     string
	ChangeFreq  string
	Priority    string
	Status      int
}
