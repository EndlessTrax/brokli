package output

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/endlesstrax/brokli/pkg/link"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name         string
		format       Format
		expectedType string
	}{
		{"default format", FormatDefault, "*output.DefaultFormatter"},
		{"verbose format", FormatVerbose, "*output.VerboseFormatter"},
		{"github format", FormatGitHub, "*output.GitHubFormatter"},
		{"unknown format defaults", "unknown", "*output.DefaultFormatter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			if formatter == nil {
				t.Error("NewFormatter() returned nil")
			}
		})
	}
}

func TestDefaultFormatter_FormatPageResults(t *testing.T) {
	tests := []struct {
		name           string
		links          []*link.AnchorTag
		expectedOutput []string
	}{
		{
			name: "all working links",
			links: []*link.AnchorTag{
				{Status: 200, Text: "Home", AbsoluteUrl: mustParseURL("http://example.com")},
				{Status: 200, Text: "About", AbsoluteUrl: mustParseURL("http://example.com/about")},
			},
			expectedOutput: []string{"All links are working"},
		},
		{
			name: "some broken links",
			links: []*link.AnchorTag{
				{Status: 200, Text: "Home", AbsoluteUrl: mustParseURL("http://example.com")},
				{Status: 404, Text: "Missing", AbsoluteUrl: mustParseURL("http://example.com/404")},
			},
			expectedOutput: []string{"Broken Links:", "Missing", "404", "Summary: 1 broken links found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &DefaultFormatter{}
			results := &link.PageResults{}

			err := formatter.FormatPageResults(&buf, results, tt.links)
			if err != nil {
				t.Errorf("FormatPageResults() error = %v", err)
				return
			}

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("FormatPageResults() output missing expected string %q\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestVerboseFormatter_FormatPageResults(t *testing.T) {
	links := []*link.AnchorTag{
		{Status: 200, Text: "Home", AbsoluteUrl: mustParseURL("http://example.com")},
		{Status: 404, Text: "Missing", AbsoluteUrl: mustParseURL("http://example.com/404")},
	}

	var buf bytes.Buffer
	formatter := &VerboseFormatter{}
	results := &link.PageResults{}

	err := formatter.FormatPageResults(&buf, results, links)
	if err != nil {
		t.Errorf("FormatPageResults() error = %v", err)
		return
	}

	output := buf.String()
	expectedStrings := []string{"All Links:", "Home", "Missing", "Summary:"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("FormatPageResults() output missing expected string %q", expected)
		}
	}
}

func TestGitHubFormatter_FormatPageResults(t *testing.T) {
	links := []*link.AnchorTag{
		{Status: 200, Text: "Home", AbsoluteUrl: mustParseURL("http://example.com")},
		{Status: 404, Text: "Missing", AbsoluteUrl: mustParseURL("http://example.com/404")},
	}

	var buf bytes.Buffer
	formatter := &GitHubFormatter{}
	results := &link.PageResults{}

	err := formatter.FormatPageResults(&buf, results, links)
	if err != nil {
		t.Errorf("FormatPageResults() error = %v", err)
		return
	}

	output := buf.String()
	expectedStrings := []string{"::error", "Broken Link", "404", "Found 1 broken links"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("FormatPageResults() output missing expected string %q\nGot: %s", expected, output)
		}
	}
}

func TestGitHubFormatter_WriteGitHubOutput(t *testing.T) {
	// Setup temp file for GITHUB_OUTPUT
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "github_output.txt")
	os.Setenv("GITHUB_OUTPUT", tmpFile)
	defer os.Unsetenv("GITHUB_OUTPUT")

	err := writeGitHubOutput(5, 10)
	if err != nil {
		t.Errorf("writeGitHubOutput() error = %v", err)
		return
	}

	// Read the file and verify contents
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	expectedStrings := []string{
		"broken_links_count=5",
		"total_links_count=10",
		"has_broken_links=true",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected string %q\nGot: %s", expected, output)
		}
	}
}

func TestGitHubFormatter_WriteGitHubOutput_NoBrokenLinks(t *testing.T) {
	// Setup temp file for GITHUB_OUTPUT
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "github_output.txt")
	os.Setenv("GITHUB_OUTPUT", tmpFile)
	defer os.Unsetenv("GITHUB_OUTPUT")

	err := writeGitHubOutput(0, 10)
	if err != nil {
		t.Errorf("writeGitHubOutput() error = %v", err)
		return
	}

	// Read the file and verify contents
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "has_broken_links=false") {
		t.Errorf("Expected has_broken_links=false when brokenCount == 0, got: %s", output)
	}
}

func TestCountBrokenLinks(t *testing.T) {
	links := []*link.AnchorTag{
		{Status: 200},
		{Status: 404},
		{Status: 500},
		{Status: 301},
		{Status: -1},
	}

	count := countBrokenLinks(links)
	if count != 3 { // 404, 500, and -1 are broken
		t.Errorf("countBrokenLinks() = %d, want 3", count)
	}
}

func TestIsBrokenLink(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"200 OK", 200, false},
		{"301 Redirect", 301, false},
		{"404 Not Found", 404, true},
		{"500 Server Error", 500, true},
		{"-1 Unchecked", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBrokenLink(tt.status); got != tt.want {
				t.Errorf("isBrokenLink(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// Helper function to parse URLs in tests
func mustParseURL(rawURL string) url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return *u
}
