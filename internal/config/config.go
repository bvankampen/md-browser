package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application settings.
type Config struct {
	RootDir         string
	Port            int
	DisableOpen     bool
	RefreshInterval int
}

// ParseConfig parses the command-line flags and returns the validated application configuration.
func ParseConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RootDir, "dir", ".", "Root directory to browse")
	flag.StringVar(&cfg.RootDir, "directory", ".", "Root directory to browse")
	flag.IntVar(&cfg.Port, "port", 8080, "Port to run the web server on")
	flag.BoolVar(&cfg.DisableOpen, "disable-open", false, "Disable automatically opening the default browser on start")
	flag.IntVar(&cfg.RefreshInterval, "refresh-interval", 5, "Interval in seconds to check for directory changes")

	flag.Parse()

	// Resolve absolute path of RootDir
	absRoot, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}
	cfg.RootDir = absRoot

	// Verify root directory exists
	info, err := os.Stat(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("root directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", cfg.RootDir)
	}

	return cfg, nil
}
