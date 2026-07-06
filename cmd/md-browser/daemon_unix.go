//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"
)

// getPIDFilePath returns the platform-specific PID file path securely isolated per local OS user.
func getPIDFilePath() string {
	username := "default"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("md-browser-%s.pid", username))
}

// getLogFilePath returns the platform-specific logs file path.
func getLogFilePath() string {
	username := "default"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("md-browser-%s.log", username))
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
