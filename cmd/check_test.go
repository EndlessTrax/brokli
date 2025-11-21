package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrintGitHubAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		text     string
		status   int
		wantType string
	}{
		{
			name:     "404 error with text",
			url:      "https://example.com/404",
			text:     "Broken Link",
			status:   404,
			wantType: "error",
		},
		{
			name:     "500 error without text",
			url:      "https://example.com/500",
			text:     "",
			status:   500,
			wantType: "error",
		},
		{
			name:     "unchecked link",
			url:      "",
			text:     "Email",
			status:   -1,
			wantType: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the function doesn't panic
			// In real usage, output would be captured but for now we just test execution
			printGitHubAnnotation(tt.url, tt.text, tt.status)
		})
	}
}

func TestWriteGitHubOutput(t *testing.T) {
	tests := []struct {
		name        string
		brokenCount int
		totalCount  int
		setupEnv    bool
		wantErr     bool
	}{
		{
			name:        "no GITHUB_OUTPUT env var",
			brokenCount: 5,
			totalCount:  10,
			setupEnv:    false,
			wantErr:     false,
		},
		{
			name:        "with broken links",
			brokenCount: 5,
			totalCount:  10,
			setupEnv:    true,
			wantErr:     false,
		},
		{
			name:        "no broken links",
			brokenCount: 0,
			totalCount:  10,
			setupEnv:    true,
			wantErr:     false,
		},
		{
			name:        "all links broken",
			brokenCount: 10,
			totalCount:  10,
			setupEnv:    true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var tmpFile string
			if tt.setupEnv {
				tmpDir := t.TempDir()
				tmpFile = filepath.Join(tmpDir, "github_output.txt")
				os.Setenv("GITHUB_OUTPUT", tmpFile)
				defer os.Unsetenv("GITHUB_OUTPUT")
			} else {
				os.Unsetenv("GITHUB_OUTPUT")
			}

			// Execute
			err := writeGitHubOutput(tt.brokenCount, tt.totalCount)

			// Verify
			if (err != nil) != tt.wantErr {
				t.Errorf("writeGitHubOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.setupEnv && err == nil {
				// Read the file and verify contents
				content, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				output := string(content)

				// Check for expected values
				expectedBrokenCount := ""
				expectedTotalCount := ""
				expectedHasBrokenLinks := ""

				// Using fmt.Sprintf to match the actual format
				expectedBrokenCount = "broken_links_count="
				expectedTotalCount = "total_links_count="
				expectedHasBrokenLinks = "has_broken_links="

				if !contains(output, expectedBrokenCount) {
					t.Errorf("Output missing broken_links_count field")
				}
				if !contains(output, expectedTotalCount) {
					t.Errorf("Output missing total_links_count field")
				}
				if !contains(output, expectedHasBrokenLinks) {
					t.Errorf("Output missing has_broken_links field")
				}

				// Verify has_broken_links value
				if tt.brokenCount > 0 {
					if !contains(output, "has_broken_links=true") {
						t.Errorf("Expected has_broken_links=true when brokenCount > 0, got: %s", output)
					}
				} else {
					if !contains(output, "has_broken_links=false") {
						t.Errorf("Expected has_broken_links=false when brokenCount == 0, got: %s", output)
					}
				}
			}
		})
	}
}

func TestWriteGitHubOutput_InvalidPath(t *testing.T) {
	// Set GITHUB_OUTPUT to an invalid path
	os.Setenv("GITHUB_OUTPUT", "/invalid/path/that/does/not/exist/output.txt")
	defer os.Unsetenv("GITHUB_OUTPUT")

	err := writeGitHubOutput(5, 10)
	if err == nil {
		t.Error("Expected error when writing to invalid path, got nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
