package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/bvankampen/md-browser/internal/config"
	"github.com/bvankampen/md-browser/web"
)

// Server represents the Markdown Browser web server.
type Server struct {
	config *config.Config
}

// NewServer creates a new Server instance.
func NewServer(cfg *config.Config) *Server {
	return &Server{config: cfg}
}

// Start configures routes and launches the HTTP server.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/list", s.handleList)
	mux.HandleFunc("/api/view", s.handleView)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/raw", s.handleRaw)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/", s.handleIndex)

	// Bind to localhost port
	addr := fmt.Sprintf("127.0.0.1:%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// If 127.0.0.1 fails, try port directly
		addr = fmt.Sprintf(":%d", s.config.Port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
		}
	}

	fmt.Printf("Web-based Markdown Browser started.\n")
	fmt.Printf("Serving: %s\n", s.config.RootDir)
	url := fmt.Sprintf("http://%s", listener.Addr().String())
	fmt.Printf("URL: %s\n", url)

	if !s.config.DisableOpen {
		go func() {
			if err := openBrowser(url); err != nil {
				log.Printf("Warning: failed to open browser automatically: %v", err)
			}
		}()
	}

	return http.Serve(listener, mux)
}

// handleIndex serves the embedded frontend files from the web package.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := web.FS.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// openBrowser opens the specified URL in the default browser of the user's platform.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
