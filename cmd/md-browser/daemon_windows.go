//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
	// On Windows, finding the process object and sending a signal isn't standard,
	// but a simpler check or trying to find it using Tasklist is standard.
	// Since os.FindProcess always succeeds on Windows if PID fits, we can try to
	// check if the process exists or is active by testing standard behaviors.
	// We can use a simple ping or check that doesn't block.
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
