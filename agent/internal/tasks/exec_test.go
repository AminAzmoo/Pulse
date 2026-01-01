package tasks

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand_Allowed(t *testing.T) {
	cmd := "echo"
	args := []interface{}{"hello"}
	
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []interface{}{"/c", "echo", "hello"}
	}

	exec := NewCommandExecutor([]string{cmd})
	
	params := map[string]interface{}{
		"command": cmd,
		"args":    args,
	}

	err := exec.Execute(params)
	assert.NoError(t, err)
}

func TestExecCommand_Denied(t *testing.T) {
	exec := NewCommandExecutor([]string{"echo"})
	
	params := map[string]interface{}{
		"command": "rm",
		"args":    []interface{}{"-rf", "/"},
	}

	err := exec.Execute(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestExecCommand_Timeout(t *testing.T) {
	cmd := "sleep"
	args := []interface{}{"2"}
	
	if runtime.GOOS == "windows" {
		cmd = "timeout"
		args = []interface{}{"2"}
	}

	exec := NewCommandExecutor([]string{cmd})
	
	params := map[string]interface{}{
		"command": cmd,
		"args":    args,
		"timeout": 0.1, // Fail fast
	}

	err := exec.Execute(params)
	assert.Error(t, err)
	// Timeout error message depends on context cancellation, usually "command timed out" or "signal: killed"
}
