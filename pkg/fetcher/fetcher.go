package fetcher

import (
	"io"
	"net/http"
)

// GetHTML fetches the HTML content from the specified URL and returns it as a byte slice
func GetHTML(url string) ([]byte, error) {
	// Fetch the HTML from the URL
	// #nosec G107 -- URL is provided by user and is expected to be variable
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
