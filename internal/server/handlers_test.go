package server

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bvankampen/md-browser/internal/config"
)

func TestSafePath(t *testing.T) {
	// Setup root directory for the test
	tmpDir := t.TempDir()

	// Resolve absolute path
	absRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	cfg := &config.Config{
		RootDir: absRoot,
	}
	s := NewServer(cfg)

	tests := []struct {
		name     string
		userPath string
		wantErr  bool
	}{
		{
			name:     "Valid simple file",
			userPath: "readme.md",
			wantErr:  false,
		},
		{
			name:     "Valid nested file",
			userPath: "docs/sub/readme.md",
			wantErr:  false,
		},
		{
			name:     "Empty path (maps to root)",
			userPath: "",
			wantErr:  false,
		},
		{
			name:     "Current directory dot",
			userPath: ".",
			wantErr:  false,
		},
		{
			name:     "Path traversal using .. prefix",
			userPath: "../outside.md",
			wantErr:  true,
		},
		{
			name:     "Path traversal deeply nested",
			userPath: "docs/../../outside.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.safePath(tt.userPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("safePath(%q) error = %v, wantErr %v", tt.userPath, err, tt.wantErr)
				return
			}
			if err == nil {
				// Make sure the resolved path is inside the root directory
				if !strings.HasPrefix(got, s.config.RootDir) {
					t.Errorf("safePath(%q) returned path %q which does not have rootDir prefix %q", tt.userPath, got, s.config.RootDir)
				}
			}
		})
	}
}
