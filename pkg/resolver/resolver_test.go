package resolver

import (
	"net/url"
	"testing"
)

func TestResolveAbsoluteUrl(t *testing.T) {
	baseUrl, _ := url.Parse("https://example.com/path/")

	t.Run("relative URL", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("page.html", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "https://example.com/path/page.html"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("absolute URL", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("https://other.com/page", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "https://other.com/page"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("empty href", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != baseUrl.String() {
			t.Errorf("Expected base URL '%s', got '%s'", baseUrl.String(), result.String())
		}
	})

	t.Run("mailto link", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("mailto:test@example.com", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for mailto, got '%s'", result.String())
		}
	})

	t.Run("javascript link", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("javascript:void(0)", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for javascript, got '%s'", result.String())
		}
	})

	t.Run("fragment link", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("#section1", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.String() != "" {
			t.Errorf("Expected empty URL for fragment, got '%s'", result.String())
		}
	})

	t.Run("protocol-relative URL", func(t *testing.T) {
		result, err := ResolveAbsoluteUrl("//other.com/path", *baseUrl)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		expected := "http://other.com/path"
		if result.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result.String())
		}
	})

	t.Run("invalid href", func(t *testing.T) {
		_, err := ResolveAbsoluteUrl("://invalid-url", *baseUrl)
		if err == nil {
			t.Error("Expected error for invalid href")
		}
	})
}

func TestIsSpecialLink(t *testing.T) {
	tests := []struct {
		name     string
		href     string
		expected bool
	}{
		{"mailto link", "mailto:test@example.com", true},
		{"javascript link", "javascript:void(0)", true},
		{"fragment link", "#section", true},
		{"empty string", "", false},
		{"normal link", "https://example.com", false},
		{"relative link", "/about", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSpecialLink(tt.href)
			if result != tt.expected {
				t.Errorf("IsSpecialLink(%q) = %v, expected %v", tt.href, result, tt.expected)
			}
		})
	}
}
