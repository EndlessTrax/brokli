package parser

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/endlesstrax/brokli/pkg/link"
)

// XMLSitemap represents the structure of a sitemap XML file
type XMLSitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []XMLURL `xml:"url"`
}

// XMLURL represents a single URL entry in a sitemap
type XMLURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// ParseSitemap parses the given byte slice into sitemap data
func ParseSitemap(b []byte) (*XMLSitemap, error) {
	var sitemap XMLSitemap
	err := xml.Unmarshal(b, &sitemap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap XML: %w", err)
	}
	return &sitemap, nil
}

// GetSitemapResults extracts all URLs from the provided sitemap XML and returns them as SitemapResults
func GetSitemapResults(sitemapData *XMLSitemap, sourceUrl string) link.SitemapResults {
	var sitemapResults link.SitemapResults
	sitemapResults.SourceUrl = sourceUrl

	for _, xmlUrl := range sitemapData.URLs {
		sitemapUrl, err := newSitemapUrl(xmlUrl.Loc, xmlUrl.LastMod, xmlUrl.ChangeFreq, xmlUrl.Priority)
		if err != nil {
			fmt.Printf("Skipping URL due to error: %v\n", err)
			continue
		}
		sitemapResults.Urls = append(sitemapResults.Urls, sitemapUrl)
	}

	return sitemapResults
}

// newSitemapUrl creates a new SitemapUrl from the provided URL string and metadata
func newSitemapUrl(urlStr, lastMod, changeFreq, priority string) (link.SitemapUrl, error) {
	if urlStr == "" {
		return link.SitemapUrl{}, fmt.Errorf("URL string cannot be empty")
	}

	// Parse the URL directly since sitemap URLs are already absolute
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return link.SitemapUrl{}, fmt.Errorf("failed to parse URL '%s': %w", urlStr, err)
	}

	// Create and return the SitemapUrl
	sitemapUrl := link.SitemapUrl{
		AbsoluteUrl: *parsedUrl,
		LastMod:     lastMod,
		ChangeFreq:  changeFreq,
		Priority:    priority,
		Status:      -1, // Initialize to -1 (unknown)
	}

	return sitemapUrl, nil
}
