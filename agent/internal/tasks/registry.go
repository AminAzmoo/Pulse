package tasks

import (
	"errors"
	"time"
)

type ExecutionFunc func(params map[string]interface{}) error

type Registry struct {
	handlers map[string]ExecutionFunc
}

func NewRegistry(allowedCommands []string) *Registry {
	r := &Registry{handlers: make(map[string]ExecutionFunc)}
	// Default handlers
	r.handlers["noop"] = func(_ map[string]interface{}) error { time.Sleep(100 * time.Millisecond); return nil }
	r.handlers["reload_service"] = func(_ map[string]interface{}) error { time.Sleep(500 * time.Millisecond); return nil }

	// Secure command execution
	cmdExec := NewCommandExecutor(allowedCommands)
	r.handlers["exec_command"] = cmdExec.Execute

	// File download
	r.handlers["file_download"] = DownloadFile

	// Self-update
	r.handlers["self_update"] = SelfUpdate

	// Service control
	r.handlers["service_action"] = ServiceAction

	return r
}

func (r *Registry) Exec(taskType string, params map[string]interface{}) error {
	fn, ok := r.handlers[taskType]
	if !ok {
		return errors.New("unknown task type")
	}
	return fn(params)
}

func (r *Registry) Register(taskType string, fn ExecutionFunc) {
	r.handlers[taskType] = fn
}
