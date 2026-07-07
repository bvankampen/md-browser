package server

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestHandleRaw(t *testing.T) {
	// Setup root directory for the test
	tmpDir := t.TempDir()

	// Resolve absolute path
	absRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Write a dummy test file in the tmp directory
	testFileName := "test-image.png"
	testFileContent := []byte("fake-image-bytes-1234")
	err = os.WriteFile(filepath.Join(absRoot, testFileName), testFileContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg := &config.Config{
		RootDir: absRoot,
	}
	s := NewServer(cfg)

	tests := []struct {
		name           string
		userPath       string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "Serve valid static file",
			userPath:       testFileName,
			wantStatusCode: http.StatusOK,
			wantBody:       string(testFileContent),
		},
		{
			name:           "File not found",
			userPath:       "doesnotexist.png",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "File not found\n",
		},
		{
			name:           "Path traversal attempted",
			userPath:       "../outside.png",
			wantStatusCode: http.StatusForbidden,
			wantBody:       "access denied: path traversal detected\n",
		},
		{
			name:           "Empty path requested",
			userPath:       "",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "Path is required\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/raw?path="+tt.userPath, nil)
			w := httptest.NewRecorder()

			s.handleRaw(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("handleRaw() status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			body := w.Body.String()
			if body != tt.wantBody {
				t.Errorf("handleRaw() body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestHasMDFiles(t *testing.T) {
	tmpDir := t.TempDir()

	emptyDir := filepath.Join(tmpDir, "empty_dir")
	noMdDir := filepath.Join(tmpDir, "no_md_dir")
	withMdDir := filepath.Join(tmpDir, "with_md_dir")
	nestedMdDir := filepath.Join(tmpDir, "nested_md_dir")

	_ = os.Mkdir(emptyDir, 0755)
	_ = os.Mkdir(noMdDir, 0755)
	_ = os.Mkdir(withMdDir, 0755)
	_ = os.Mkdir(nestedMdDir, 0755)

	_ = os.WriteFile(filepath.Join(noMdDir, "file.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(withMdDir, "readme.md"), []byte("# hello"), 0644)

	subDir := filepath.Join(nestedMdDir, "sub")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "doc.markdown"), []byte("# doc"), 0644)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Completely empty directory",
			path: emptyDir,
			want: false,
		},
		{
			name: "Directory with only non-markdown files",
			path: noMdDir,
			want: false,
		},
		{
			name: "Directory with direct markdown file",
			path: withMdDir,
			want: true,
		},
		{
			name: "Directory with nested markdown file",
			path: nestedMdDir,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasMDFiles(tt.path)
			if got != tt.want {
				t.Errorf("hasMDFiles(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestHandleSearch(t *testing.T) {
	tmpDir := t.TempDir()

	absRoot, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create some mock files
	_ = os.WriteFile(filepath.Join(absRoot, "file1.md"), []byte("This is a special test line.\nAnother line here."), 0644)
	_ = os.WriteFile(filepath.Join(absRoot, "file2.markdown"), []byte("No match here.\nActually, SPECIAL match here."), 0644)
	_ = os.WriteFile(filepath.Join(absRoot, "file3.txt"), []byte("This has a special line but it is not markdown."), 0644)

	cfg := &config.Config{
		RootDir: absRoot,
	}
	s := NewServer(cfg)

	tests := []struct {
		name           string
		query          string
		wantStatusCode int
		wantInBody     []string
		dontWantInBody []string
	}{
		{
			name:           "Valid search with matches",
			query:          "special",
			wantStatusCode: http.StatusOK,
			wantInBody:     []string{"file1.md", "file2.markdown", "special test line", "SPECIAL match"},
			dontWantInBody: []string{"file3.txt"},
		},
		{
			name:           "Case insensitive match",
			query:          "SpEcIaL",
			wantStatusCode: http.StatusOK,
			wantInBody:     []string{"file1.md", "file2.markdown", "special test line", "SPECIAL match"},
			dontWantInBody: []string{"file3.txt"},
		},
		{
			name:           "Empty query search",
			query:          "",
			wantStatusCode: http.StatusOK,
			wantInBody:     []string{"[]"},
			dontWantInBody: []string{"file1.md", "special"},
		},
		{
			name:           "No matching search",
			query:          "nonexistentkeyword",
			wantStatusCode: http.StatusOK,
			wantInBody:     []string{"[]"},
			dontWantInBody: []string{"file1.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/search?q="+tt.query, nil)
			w := httptest.NewRecorder()

			s.handleSearch(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("handleSearch() status = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			body := w.Body.String()
			for _, searchStr := range tt.wantInBody {
				if !strings.Contains(body, searchStr) {
					t.Errorf("handleSearch() body = %q; expected to contain %q", body, searchStr)
				}
			}
			for _, dontSearchStr := range tt.dontWantInBody {
				if strings.Contains(body, dontSearchStr) {
					t.Errorf("handleSearch() body = %q; expected NOT to contain %q", body, dontSearchStr)
				}
			}
		})
	}
}
