package services

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

func TestCreateTask_IdempotencyReplay(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Task{}, &models.IdempotencyRecord{}, &models.AlertRule{}, &models.Alert{}) // Add all models to be safe
	svc := NewTaskService(db)

	req := TaskRequest{Identity: "node-1", Type: "noop", Class: models.TaskClassLight, Params: map[string]interface{}{"x": 1}}
	task1, replay1, err1 := svc.CreateTask(req)
	assert.NoError(t, err1)
	assert.False(t, replay1)
	assert.Equal(t, models.TaskStatusPending, task1.Status)

	task2, replay2, err2 := svc.CreateTask(req)
	assert.NoError(t, err2)
	assert.True(t, replay2)
	assert.Equal(t, task1.ID, task2.ID)
}

func TestBackendCompletionRequiresAgentEvent(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Task{}, &models.IdempotencyRecord{}, &models.AlertRule{}, &models.Alert{})
	svc := NewTaskService(db)

	req := TaskRequest{Identity: "node-1", Type: "noop", Class: models.TaskClassLight, Params: map[string]interface{}{"x": 1}}
	task, _, _ := svc.CreateTask(req)
	assert.Equal(t, models.TaskStatusPending, task.Status)

	_ = svc.MarkInProgress(task.ID)
	var in models.Task
	_ = db.First(&in, task.ID).Error
	assert.Equal(t, models.TaskStatusInProgress, in.Status)

	// only agent event can mark completed
	err := svc.HandleStatusUpdate(TaskStatusUpdate{Kind: "status", TaskID: task.ID, Status: models.TaskStatusCompleted})
	assert.NoError(t, err)
	var out models.Task
	_ = db.First(&out, task.ID).Error
	assert.Equal(t, models.TaskStatusCompleted, out.Status)
}
