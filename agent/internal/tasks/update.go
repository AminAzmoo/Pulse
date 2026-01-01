package tasks

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"time"

	"go.uber.org/zap"
)

func SelfUpdate(params map[string]interface{}) error {
	url, ok := params["url"].(string)
	if !ok || url == "" {
		return errors.New("missing or invalid 'url' parameter")
	}

	hash, ok := params["sha256"].(string)
	if !ok || hash == "" {
		return errors.New("missing or invalid 'sha256' parameter")
	}

	// 1. Identify current executable
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks to find the real binary
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	zap.L().Info("Starting self-update",
		zap.String("current_exe", currentExe),
		zap.String("url", url),
	)

	// 2. Download new binary
	newExe := currentExe + ".new"
	if err := doDownload(url, newExe, hash); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(newExe) // Cleanup if we fail before rename

	// 3. Prepare for replacement
	if runtime.GOOS != "windows" {
		if err := os.Chmod(newExe, 0755); err != nil {
			return fmt.Errorf("failed to chmod new binary: %w", err)
		}
	}

	// 4. Atomic Replace
	// On Linux: Rename overwrites.
	// On Windows: Cannot overwrite running exe. Needs separate handling (rename current -> old, then new -> current).
	if runtime.GOOS == "windows" {
		oldExe := currentExe + ".old"
		_ = os.Remove(oldExe) // Remove any previous backup

		if err := os.Rename(currentExe, oldExe); err != nil {
			return fmt.Errorf("failed to rename current exe to .old: %w", err)
		}

		if err := os.Rename(newExe, currentExe); err != nil {
			// Rollback
			_ = os.Rename(oldExe, currentExe)
			return fmt.Errorf("failed to move new exe to current location: %w", err)
		}
	} else {
		if err := os.Rename(newExe, currentExe); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	}

	zap.L().Info("Binary replaced successfully. Restarting service...")

	// 5. Restart Service
	// We use a non-blocking command to restart systemd service,
	// because if we restart synchronously, we get killed before returning.
	// Actually, systemctl restart will kill us.
	// We want to return 'success' to the backend so the task is marked as done.
	// Then restart.

	go func() {
		// Wait a bit to allow the response to be sent
		time.Sleep(2 * time.Second)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			// Powershell restart service? Or exit and let Service Manager restart?
			// If running as service, exit(0) might not trigger restart if configured 'on-failure'.
			// Safest is to use net stop/start or sc, but we are the service.
			// Just exiting with non-zero might trigger restart.
			os.Exit(1)
		} else {
			// Linux systemd
			cmd = exec.Command("systemctl", "restart", "pulse-agent")
			// If we are not root/sudo, this fails. Agent usually runs as root.
			if err := cmd.Run(); err != nil {
				zap.L().Error("Failed to restart service", zap.Error(err))
				// Fallback: exit and hope systemd restarts us
				os.Exit(1)
			}
		}
	}()

	return nil
}
