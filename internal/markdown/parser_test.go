package markdown

import (
	"strings"
	"testing"
)

func TestMarkdownConversion(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		contains string
	}{
		{
			name:     "Header rendering",
			markdown: "# Title Hello",
			contains: "<h1>Title Hello</h1>",
		},
		{
			name:     "GFM Strikethrough",
			markdown: "~~strikethrough~~",
			contains: "<del>strikethrough</del>",
		},
		{
			name:     "GFM Table rendering",
			markdown: "| Col 1 | Col 2 |\n| --- | --- |\n| Val 1 | Val 2 |",
			contains: "<table>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := Convert([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}
			got := string(gotBytes)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("Convert(%q) = %q; expected it to contain %q", tt.markdown, got, tt.contains)
			}
		})
	}
}
