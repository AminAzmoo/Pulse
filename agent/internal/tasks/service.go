package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ServiceAction handles service control (start, stop, restart, status).
// Params:
//   - service: string (required)
//   - action: string (required) - start, stop, restart, status
func ServiceAction(params map[string]interface{}) error {
	service, ok := params["service"].(string)
	if !ok || service == "" {
		return errors.New("missing or invalid 'service' parameter")
	}

	action, ok := params["action"].(string)
	if !ok || action == "" {
		return errors.New("missing or invalid 'action' parameter")
	}

	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"status":  true,
	}

	if !validActions[action] {
		return fmt.Errorf("invalid action '%s'. Must be one of: start, stop, restart, status", action)
	}

	zap.L().Info("Executing service action", zap.String("service", service), zap.String("action", action))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if runtime.GOOS == "windows" {
		return handleWindowsService(ctx, service, action)
	}
	return handleLinuxService(ctx, service, action)
}

func handleLinuxService(ctx context.Context, service, action string) error {
	// systemctl <action> <service>
	cmd := exec.CommandContext(ctx, "systemctl", action, service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl failed: %s: %w", string(out), err)
	}
	return nil
}

func handleWindowsService(ctx context.Context, service, action string) error {
	// Using PowerShell for consistent behavior
	// Start-Service, Stop-Service, Restart-Service, Get-Service
	
	var psCmd string
	switch action {
	case "start":
		psCmd = fmt.Sprintf("Start-Service -Name '%s'", service)
	case "stop":
		psCmd = fmt.Sprintf("Stop-Service -Name '%s' -Force", service)
	case "restart":
		psCmd = fmt.Sprintf("Restart-Service -Name '%s' -Force", service)
	case "status":
		psCmd = fmt.Sprintf("Get-Service -Name '%s'", service)
	}

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// PowerShell error output can be verbose, try to capture it
		return fmt.Errorf("powershell failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
