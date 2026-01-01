package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

type CommandExecutor struct {
	allowedCommands map[string]bool
}

func NewCommandExecutor(allowed []string) *CommandExecutor {
	m := make(map[string]bool)
	for _, cmd := range allowed {
		m[cmd] = true
	}
	return &CommandExecutor{allowedCommands: m}
}

func (e *CommandExecutor) Execute(params map[string]interface{}) error {
	cmdName, ok := params["command"].(string)
	if !ok {
		return errors.New("missing or invalid 'command' parameter")
	}

	if !e.allowedCommands[cmdName] {
		return fmt.Errorf("command '%s' is not allowed", cmdName)
	}

	argsRaw, ok := params["args"].([]interface{})
	var args []string
	if ok {
		for _, arg := range argsRaw {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	timeout := 30 * time.Second
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, args...)
	output, err := cmd.CombinedOutput()
	
	zap.L().Info("Executed command", 
		zap.String("command", cmdName),
		zap.Strings("args", args),
		zap.String("output", string(output)),
		zap.Error(err),
	)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New("command timed out")
		}
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}

	return nil
}
