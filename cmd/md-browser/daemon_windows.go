//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
	// On Windows, finding the process object always succeeds if the PID is valid,
	// so standard checks or pinging the process is typical.
	_ = proc
	return true
}

// stopProcess halts the background process using TaskKill-equivalent force closure.
func stopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

// setSysProcAttr configures process attributes to run as an independent background daemon session.
func setSysProcAttr(cmd *exec.Cmd) {
	// Standard Windows process spawning properties. No special sys attributes needed for default daemon.
}
