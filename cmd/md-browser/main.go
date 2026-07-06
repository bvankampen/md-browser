package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bvankampen/md-browser/internal/config"
	"github.com/bvankampen/md-browser/internal/server"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	pidFile := getPIDFilePath()

	// Handle Stop command
	if cfg.Stop {
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Println("Markdown Browser is not running.")
			return
		}

		pidStr := strings.TrimSpace(string(pidData))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Println("Markdown Browser is not running (invalid PID file). Removing corrupt file.")
			os.Remove(pidFile)
			return
		}

		if isProcessRunning(pid) {
			if err := stopProcess(pid); err != nil {
				fmt.Printf("Failed to stop process %d: %v\n", pid, err)
				return
			}
			os.Remove(pidFile)
			fmt.Println("Markdown Browser stopped.")
		} else {
			os.Remove(pidFile)
			fmt.Println("Markdown Browser was not running (stale PID file cleaned up).")
		}
		return
	}

	// Prevent running multiple background processes simultaneously
	if pidData, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if pid != os.Getpid() && isProcessRunning(pid) {
				fmt.Printf("Markdown Browser is already running in background (PID: %d).\n", pid)
				fmt.Println("To stop it, run: md-browser -stop")
				os.Exit(0)
			}
		}
	}

	// Verify port availability and auto-resolve conflicts
	if !isPortAvailable(cfg.Port) {
		freePort := findNextFreePort(cfg.Port + 1)
		fmt.Printf("⚠️ Port %d is already in use!\n", cfg.Port)
		fmt.Printf("💡 Automatically resolving conflict: Starting Markdown Browser on next available port: %d\n\n", freePort)
		cfg.Port = freePort
	}

	// Check if running in Foreground or as a spawned Background child
	if cfg.Foreground || os.Getenv("MD_BROWSER_DAEMON") == "1" {
		// Save current PID to PID file
		_ = os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)

		srv := server.NewServer(cfg)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
		return
	}

	// Daemonize and spawn background copy of the current process
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to resolve executable path: %v", err)
	}

	// Build the exact CLI arguments to preserve state (forcing foreground inside the spawned copy)
	args := []string{
		"-dir", cfg.RootDir,
		"-port", strconv.Itoa(cfg.Port),
		"-refresh-interval", strconv.Itoa(cfg.RefreshInterval),
		"-foreground",
	}
	if cfg.DisableOpen {
		args = append(args, "-disable-open")
	}

	cmd := exec.Command(execPath, args...)
	cmd.Env = append(os.Environ(), "MD_BROWSER_DAEMON=1")

	// Redirect daemon outputs to a user-isolated background log file
	logFile, err := os.OpenFile(getLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	} else {
		fmt.Printf("Warning: failed to open log file, background logging disabled: %v\n", err)
	}

	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start background daemon: %v", err)
	}

	// Write daemon process PID to PID file
	_ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)

	fmt.Printf("Markdown Browser started in background (PID: %d).\n", cmd.Process.Pid)
	fmt.Printf("Logs are written to: %s\n", getLogFilePath())
}

// isPortAvailable checks if a local TCP port is open to bind.
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	l.Close()
	return true
}

// findNextFreePort increments the port number from startPort until finding a free TCP port.
func findNextFreePort(startPort int) int {
	port := startPort
	for {
		if isPortAvailable(port) {
			return port
		}
		port++
	}
}
