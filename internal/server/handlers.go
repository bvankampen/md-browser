package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bvankampen/md-browser/internal/markdown"
)

type FileItem struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	HasNoMD bool   `json:"has_no_md,omitempty"`
}

// safePath validates and returns the absolute target path within the root directory.
func (s *Server) safePath(userPath string) (string, error) {
	// Join and clean the path with RootDir
	targetPath := filepath.Join(s.config.RootDir, userPath)

	// Check for path traversal using filepath.Rel
	rel, err := filepath.Rel(s.config.RootDir, targetPath)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", errors.New("access denied: path traversal detected")
	}

	return targetPath, nil
}

// hasMDFiles checks recursively if a directory contains any Markdown files.
func hasMDFiles(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files/directories (starting with dot)
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			if hasMDFiles(filepath.Join(dirPath, name)) {
				return true
			}
		} else {
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".md" || ext == ".markdown" {
				return true
			}
		}
	}
	return false
}

// handleList lists files in the given directory path.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userPath := r.URL.Query().Get("path")
	showAll := r.URL.Query().Get("all") == "true"

	targetDir, err := s.safePath(userPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	var items []FileItem
	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files/directories (starting with dot) unless they are explicitly requested
		// We always skip .git directory
		if strings.HasPrefix(name, ".") {
			continue
		}

		isDir := entry.IsDir()
		var size int64
		if !isDir {
			info, err := entry.Info()
			if err == nil {
				size = info.Size()
			}
		}

		// Filter logic:
		// Keep all directories.
		// For files, if showAll is true, keep everything.
		// If showAll is false, only keep files with .md or .markdown extensions.
		if !isDir && !showAll {
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".md" && ext != ".markdown" {
				continue
			}
		}

		// Get relative path of this item
		relPath := filepath.Join(userPath, name)

		var hasNoMD bool
		if isDir {
			hasNoMD = !hasMDFiles(filepath.Join(targetDir, name))
		}

		items = append(items, FileItem{
			Name:    name,
			Path:    relPath,
			IsDir:   isDir,
			Size:    size,
			HasNoMD: hasNoMD,
		})
	}

	// Ensure empty slice is serialized as [] instead of null
	if items == nil {
		items = []FileItem{}
	}

	json.NewEncoder(w).Encode(items)
}

// handleConfig returns the server configuration (e.g. refresh interval).
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"refresh_interval": s.config.RefreshInterval,
	}
	json.NewEncoder(w).Encode(response)
}

// handleView renders and returns the requested markdown file content as JSON HTML, or raw source text.
func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userPath := r.URL.Query().Get("path")
	if userPath == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	targetFile, err := s.safePath(userPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		http.Error(w, "Path is a directory", http.StatusBadRequest)
		return
	}

	isBin, size, err := isBinaryFile(targetFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to inspect file: %v", err), http.StatusInternalServerError)
		return
	}

	if isBin {
		response := map[string]interface{}{
			"path":      userPath,
			"title":     filepath.Base(targetFile),
			"is_binary": true,
			"size":      size,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(targetFile))
	if ext == ".md" || ext == ".markdown" {
		htmlData, err := markdown.Convert(content)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse markdown: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"path":        userPath,
			"title":       filepath.Base(targetFile),
			"is_binary":   false,
			"is_markdown": true,
			"html":        string(htmlData),
			"content":     string(content),
			"ext":         strings.TrimPrefix(ext, "."),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// For other text/source files, return raw content and file extension (for highlight.js)
	response := map[string]interface{}{
		"path":        userPath,
		"title":       filepath.Base(targetFile),
		"is_binary":   false,
		"is_markdown": false,
		"content":     string(content),
		"ext":         strings.TrimPrefix(ext, "."),
	}
	json.NewEncoder(w).Encode(response)
}

// isBinaryFile checks if the file at path is a binary file and returns its size.
func isBinaryFile(path string) (bool, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, 0, err
	}
	size := info.Size()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, 0, err
	}

	// 1. Check for null bytes (standard way to spot binary files)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, size, nil
		}
	}

	// 2. Sniff content type
	contentType := http.DetectContentType(buf[:n])
	if strings.HasPrefix(contentType, "text/") || contentType == "application/json" || contentType == "application/javascript" || contentType == "application/xml" {
		return false, size, nil
	}

	if contentType == "application/octet-stream" {
		return true, size, nil
	}

	return false, size, nil
}

// handleRaw serves a raw static file (e.g., images, diagrams) securely.
func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
	userPath := r.URL.Query().Get("path")
	if userPath == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	targetFile, err := s.safePath(userPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to access file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if info.IsDir() {
		http.Error(w, "Path is a directory", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, targetFile)
}
