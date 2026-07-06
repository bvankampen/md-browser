package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
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

	pidFile := getPIDFilePath(cfg.Port)

	// Handle Show Logs command
	if cfg.ShowLogs {
		logFile := getLogFilePath(cfg.Port)
		logData, err := os.ReadFile(logFile)
		if err != nil {
			fmt.Printf("No logs found for Markdown Browser running on port %d.\n", cfg.Port)
			return
		}
		fmt.Print(string(logData))
		return
	}

	// Handle Status command (list all running instances)
	if cfg.Status {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Failed to resolve user home directory.")
			return
		}
		logDir := filepath.Join(home, ".local", "md-browser", "log")
		matches, err := filepath.Glob(filepath.Join(logDir, "md-browser-*.pid"))
		if err != nil || len(matches) == 0 {
			fmt.Println("No active md-browser instances found.")
			return
		}

		activeCount := 0
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			lines := strings.Split(string(content), "\n")
			if len(lines) < 3 {
				os.Remove(match)
				continue
			}
			pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
			if err != nil {
				os.Remove(match)
				continue
			}
			dir := strings.TrimSpace(lines[1])
			port, err := strconv.Atoi(strings.TrimSpace(lines[2]))
			if err != nil {
				os.Remove(match)
				continue
			}

			if isProcessRunning(pid) {
				if activeCount == 0 {
					fmt.Println("Running md-browser instances:")
				}
				fmt.Printf("  • PID: %d | Port: %d | Directory: %s\n", pid, port, dir)
				activeCount++
			} else {
				// Clean up stale file
				os.Remove(match)
				os.Remove(getLogFilePath(port))
			}
		}
		if activeCount == 0 {
			fmt.Println("No active md-browser instances found.")
		}
		return
	}

	// Handle Stop command
	if cfg.Stop {
		if !cfg.PortPassed {
			// Stop ALL running instances!
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Failed to resolve user home directory.")
				return
			}
			logDir := filepath.Join(home, ".local", "md-browser", "log")
			matches, err := filepath.Glob(filepath.Join(logDir, "md-browser-*.pid"))
			if err != nil || len(matches) == 0 {
				fmt.Println("No running Markdown Browser instances found.")
				return
			}

			stoppedCount := 0
			for _, match := range matches {
				content, err := os.ReadFile(match)
				if err != nil {
					continue
				}
				lines := strings.Split(string(content), "\n")
				if len(lines) < 3 {
					os.Remove(match)
					continue
				}
				pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
				if err != nil {
					os.Remove(match)
					continue
				}
				port, err := strconv.Atoi(strings.TrimSpace(lines[2]))
				if err != nil {
					os.Remove(match)
					continue
				}

				if isProcessRunning(pid) {
					if err := stopProcess(pid); err == nil {
						fmt.Printf("Stopped Markdown Browser running on port %d (PID %d).\n", port, pid)
						stoppedCount++
					} else {
						fmt.Printf("Failed to stop process %d on port %d: %v\n", pid, port, err)
					}
				}
				os.Remove(match)
				os.Remove(getLogFilePath(port))
			}

			if stoppedCount == 0 {
				fmt.Println("No running Markdown Browser instances found (stale files cleaned up).")
			} else {
				fmt.Printf("Successfully stopped %d Markdown Browser instances.\n", stoppedCount)
			}
			return
		}

		// Stop ONLY the specified port!
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Printf("Markdown Browser is not running on port %d.\n", cfg.Port)
			return
		}

		lines := strings.Split(string(pidData), "\n")
		if len(lines) == 0 {
			fmt.Printf("Markdown Browser is not running on port %d (invalid PID file).\n", cfg.Port)
			os.Remove(pidFile)
			return
		}

		pidStr := strings.TrimSpace(lines[0])
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Printf("Markdown Browser is not running on port %d (corrupt PID). Cleaning up PID file.\n", cfg.Port)
			os.Remove(pidFile)
			return
		}

		if isProcessRunning(pid) {
			if err := stopProcess(pid); err != nil {
				fmt.Printf("Failed to stop process %d: %v\n", pid, err)
				return
			}
			os.Remove(pidFile)
			os.Remove(getLogFilePath(cfg.Port))
			fmt.Printf("Markdown Browser on port %d stopped.\n", cfg.Port)
		} else {
			os.Remove(pidFile)
			os.Remove(getLogFilePath(cfg.Port))
			fmt.Printf("Markdown Browser on port %d was not running (stale files cleaned up).\n", cfg.Port)
		}
		return
	}

	// Prevent running multiple background processes simultaneously on the SAME directory
	home, err := os.UserHomeDir()
	if err == nil {
		logDir := filepath.Join(home, ".local", "md-browser", "log")
		matches, _ := filepath.Glob(filepath.Join(logDir, "md-browser-*.pid"))
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			lines := strings.Split(string(content), "\n")
			if len(lines) < 3 {
				continue
			}
			pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
			if err != nil {
				continue
			}
			dir := strings.TrimSpace(lines[1])
			port, err := strconv.Atoi(strings.TrimSpace(lines[2]))
			if err != nil {
				continue
			}

			if pid != os.Getpid() && isProcessRunning(pid) {
				if filepath.Clean(dir) == filepath.Clean(cfg.RootDir) {
					fmt.Printf("Markdown Browser is already running for directory %s (PID: %d) on port %d.\n", dir, pid, port)
					fmt.Printf("To stop it, run: md-browser --stop --port %d\n", port)
					os.Exit(0)
				}
			}
		}
	}

	// Verify port availability and auto-resolve conflicts
	if !isPortAvailable(cfg.Port) {
		freePort := findNextFreePort(cfg.Port + 1)
		fmt.Printf("⚠️ Port %d is already in use!\n", cfg.Port)
		fmt.Printf("💡 Automatically resolving conflict: Starting Markdown Browser on next available port: %d\n\n", freePort)
		cfg.Port = freePort
		// Re-evaluate the PID file path for the updated port
		pidFile = getPIDFilePath(cfg.Port)
	}

	// Prevent running multiple background processes simultaneously on the SAME port
	if pidData, err := os.ReadFile(pidFile); err == nil {
		lines := strings.Split(string(pidData), "\n")
		if len(lines) > 0 {
			pidStr := strings.TrimSpace(lines[0])
			if pid, err := strconv.Atoi(pidStr); err == nil {
				if pid != os.Getpid() && isProcessRunning(pid) {
					fmt.Printf("Markdown Browser is already running on port %d (PID: %d).\n", cfg.Port, pid)
					fmt.Printf("To stop it, run: md-browser --stop --port %d\n", cfg.Port)
					os.Exit(0)
				}
			}
		}
	}

	// Check if running in Foreground or as a spawned Background child
	if cfg.Foreground || os.Getenv("MD_BROWSER_DAEMON") == "1" {
		// Save current PID and metadata to PID file
		data := fmt.Sprintf("%d\n%s\n%d\n", os.Getpid(), cfg.RootDir, cfg.Port)
		_ = os.WriteFile(pidFile, []byte(data), 0644)

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
		"--dir", cfg.RootDir,
		"--port", strconv.Itoa(cfg.Port),
		"--refresh-interval", strconv.Itoa(cfg.RefreshInterval),
		"--foreground",
	}
	if cfg.DisableOpen {
		args = append(args, "--disable-open")
	}

	cmd := exec.Command(execPath, args...)
	cmd.Env = append(os.Environ(), "MD_BROWSER_DAEMON=1")

	// Redirect daemon outputs to a user-isolated background log file
	logFile, err := os.OpenFile(getLogFilePath(cfg.Port), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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

	// Write daemon process PID and metadata to PID file
	data := fmt.Sprintf("%d\n%s\n%d\n", cmd.Process.Pid, cfg.RootDir, cfg.Port)
	_ = os.WriteFile(pidFile, []byte(data), 0644)

	fmt.Printf("Markdown Browser started in background (PID: %d) on port %d.\n", cmd.Process.Pid, cfg.Port)
	fmt.Printf("Logs are written to: %s\n", getLogFilePath(cfg.Port))
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
