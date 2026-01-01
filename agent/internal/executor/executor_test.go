package executor

import (
	"sync/atomic"
	"testing"

	"pulse/agent/internal/tasks"

	"github.com/stretchr/testify/assert"
)

type mockReporter struct{ statuses []string }

func (m *mockReporter) ReportStatus(taskID uint, status string, errMsg string) error {
	m.statuses = append(m.statuses, status)
	return nil
}

func TestIdempotency_NoDoubleApply(t *testing.T) {
	reg := tasks.NewRegistry(nil) // nil means all allowed or default
	var count int32
	reg.Register("noop", func(params map[string]interface{}) error { atomic.AddInt32(&count, 1); return nil })

	tmp := t.TempDir() + "/state.json"
	exec := NewExecutor(reg, 1, 1, tmp)
	env := Envelope{TaskID: 1, Identity: "node-1", Type: "noop", Class: "light", IdempotencyKey: "k1", TimeoutSeconds: 2}
	rep := &mockReporter{}
	exec.Execute(env, rep)
	exec.Execute(env, rep)
	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
}

func TestHeavyConcurrency_OneAtATime(t *testing.T) {
	r := tasks.NewRegistry(nil)
	started := make(chan struct{}, 2)
	done := make(chan struct{}, 2)

	// custom long-running heavy task
	r.Register("long", func(params map[string]interface{}) error { started <- struct{}{}; <-done; return nil })
	tmp := t.TempDir() + "/state.json"
	exec2 := NewExecutor(r, 1, 2, tmp)
	rep := &mockReporter{}
	go exec2.Execute(Envelope{TaskID: 1, Identity: "n", Type: "long", Class: "heavy", IdempotencyKey: "k1", TimeoutSeconds: 5}, rep)
	go exec2.Execute(Envelope{TaskID: 2, Identity: "n", Type: "long", Class: "heavy", IdempotencyKey: "k2", TimeoutSeconds: 5}, rep)

	// first starts
	<-started
	// ensure second hasn't started yet
	select {
	case <-started:
		t.Fatalf("second heavy task should not start before first completes")
	default:
	}
	// allow first to complete
	done <- struct{}{}
}
