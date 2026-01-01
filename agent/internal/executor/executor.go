package executor

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"pulse/agent/internal/tasks"
)

const (
	statusInProgress           = "in_progress"
	statusAwaitingConfirmation = "awaiting_confirmation"
	statusCompleted            = "completed"
	statusFailed               = "failed"
)

type Envelope struct {
	TaskID         uint                   `json:"task_id"`
	Identity       string                 `json:"identity"`
	Type           string                 `json:"type"`
	Class          string                 `json:"class"`
	IdempotencyKey string                 `json:"idempotency_key"`
	TimeoutSeconds int                    `json:"timeout_seconds"`
	Params         map[string]interface{} `json:"params"`
}

type Reporter interface {
	ReportStatus(taskID uint, status string, errMsg string) error
}

type Executor struct {
	reg         *tasks.Registry
	heavySlots  int
	lightSlots  int
	idemPath    string
	mu          sync.Mutex
	appliedIdem map[string]string
	heavyChan   chan struct{}
	lightChan   chan struct{}
}

func NewExecutor(reg *tasks.Registry, heavySlots, lightSlots int, statePath string) *Executor {
	e := &Executor{
		reg:         reg,
		heavySlots:  heavySlots,
		lightSlots:  lightSlots,
		idemPath:    statePath + ".idempotency.json",
		appliedIdem: make(map[string]string),
		heavyChan:   make(chan struct{}, heavySlots),
		lightChan:   make(chan struct{}, lightSlots),
	}
	e.loadIdem()
	return e
}

func (e *Executor) Capacity() (int, int) {
	return e.heavySlots - len(e.heavyChan), e.lightSlots - len(e.lightChan)
}

func (e *Executor) Execute(env Envelope, r Reporter) {
	// idempotency check
	e.mu.Lock()
	if _, exists := e.appliedIdem[env.IdempotencyKey]; exists {
		e.mu.Unlock()
		_ = r.ReportStatus(env.TaskID, statusCompleted, "idempotent")
		return
	}
	e.mu.Unlock()

	// concurrency slot
	if env.Class == "heavy" {
		e.heavyChan <- struct{}{}
		defer func() { <-e.heavyChan }()
	} else {
		e.lightChan <- struct{}{}
		defer func() { <-e.lightChan }()
	}

	_ = r.ReportStatus(env.TaskID, statusInProgress, "")

	done := make(chan error, 1)
	go func() { done <- e.reg.Exec(env.Type, env.Params) }()

	select {
	case err := <-done:
		if err != nil {
			_ = r.ReportStatus(env.TaskID, statusFailed, err.Error())
			return
		}
		e.markApplied(env.IdempotencyKey)
		_ = r.ReportStatus(env.TaskID, statusCompleted, "")
	case <-time.After(time.Duration(env.TimeoutSeconds) * time.Second):
		_ = r.ReportStatus(env.TaskID, statusFailed, "timeout")
	}
}

func (e *Executor) markApplied(key string) {
	e.mu.Lock()
	e.appliedIdem[key] = "completed"
	e.mu.Unlock()
	e.saveIdem()
}

func (e *Executor) loadIdem() {
	b, err := os.ReadFile(e.idemPath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &e.appliedIdem)
}

func (e *Executor) saveIdem() {
	b, _ := json.Marshal(e.appliedIdem)
	_ = os.WriteFile(e.idemPath, b, 0o600)
}
