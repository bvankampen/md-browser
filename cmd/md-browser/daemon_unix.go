//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// getPIDFilePath returns the platform-specific PID file path securely isolated under ~/.local/md-browser/log/
func getPIDFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "md-browser.pid")
	}
	dir := filepath.Join(home, ".local", "md-browser", "log")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "md-browser.pid")
}

// getLogFilePath returns the platform-specific logs file path securely isolated under ~/.local/md-browser/log/
func getLogFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "md-browser.log")
	}
	dir := filepath.Join(home, ".local", "md-browser", "log")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "md-browser.log")
}

// isProcessRunning checks if the background process with given PID is currently active.
func isProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, sending signal 0 checks if the process exists and is active.
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// stopProcess sends a graceful SIGTERM signal to the process to allow cleanup.
func stopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}

// setSysProcAttr configures process attributes to run as an independent background daemon session.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Creates a new session so closing terminal doesn't kill the child daemon
	}
}
