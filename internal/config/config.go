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
	Foreground      bool
	Stop            bool
	ShowLogs        bool
	Status          bool
}

// ParseConfig parses the command-line flags and returns the validated application configuration.
func ParseConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RootDir, "dir", ".", "Root directory to browse")
	flag.StringVar(&cfg.RootDir, "directory", ".", "Root directory to browse")
	flag.IntVar(&cfg.Port, "port", 8080, "Port to run the web server on")
	flag.BoolVar(&cfg.DisableOpen, "disable-open", false, "Disable automatically opening the default browser on start")
	flag.IntVar(&cfg.RefreshInterval, "refresh-interval", 5, "Interval in seconds to check for directory changes")
	flag.BoolVar(&cfg.Foreground, "foreground", false, "Run the application in the foreground instead of background daemonizing")
	flag.BoolVar(&cfg.Stop, "stop", false, "Stop the currently running background instance of Markdown Browser")
	flag.BoolVar(&cfg.ShowLogs, "show-logs", false, "Show logs of the background Markdown Browser instance")
	flag.BoolVar(&cfg.Status, "status", false, "Show currently running Markdown Browser instances")

	flag.Parse()

	// If stop, status, or show-logs is requested, bypass directory checks immediately
	if cfg.Stop || cfg.Status || cfg.ShowLogs {
		return cfg, nil
	}

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
