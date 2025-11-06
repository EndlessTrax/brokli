package resolver

import (
	"fmt"
	"net/url"
)

// ResolveAbsoluteUrl resolves the href attribute to an absolute URL
func ResolveAbsoluteUrl(href string, baseUrl url.URL) (url.URL, error) {
	// If href is empty, return the base URL
	if href == "" {
		return baseUrl, nil
	}

	// Handle special schemes that shouldn't be resolved
	if IsSpecialLink(href) {
		return url.URL{}, nil
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

// IsSpecialLink returns true if the link is a special type that shouldn't be resolved
// (mailto:, javascript:, or fragment links starting with #)
func IsSpecialLink(href string) bool {
	if len(href) == 0 {
		return false
	}

	// Check for mailto:, javascript:, or fragment links
	if (len(href) >= 7 && href[:7] == "mailto:") ||
		(len(href) >= 11 && href[:11] == "javascript:") ||
		(len(href) >= 1 && href[0] == '#') {
		return true
	}

	return false
}
