package services

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

func setupAlertTestDB(t *testing.T) *gorm.DB {
	// Use shared cache for in-memory DB to ensure all connections see the same data
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	// Ensure all models are migrated
	err = db.AutoMigrate(&models.Node{}, &models.Metric{}, &models.AlertRule{}, &models.Alert{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestAlerting_FireAndResolve(t *testing.T) {
	db := setupAlertTestDB(t)
	// AlertService is initialized within IdentityService, but we need to seed rules manually if using fresh DB
	// NewIdentityService calls NewAlertService which seeds defaults
	svc := NewIdentityService(db)

	node, _ := svc.CreateNode("alert-node")

	// 1. Normal metric - no alert
	normal := &models.Metric{CPU: 50.0, Memory: 50.0, Disk: 50.0}
	err := svc.RecordMetrics("alert-node", normal)
	assert.NoError(t, err)

	// Check alerts - should be empty
	alerts, err := svc.alertSvc.GetActiveAlerts(node.ID)
	assert.NoError(t, err)
	assert.Len(t, alerts, 0)

	// 2. High CPU metric - trigger alert
	highCPU := &models.Metric{CPU: 95.0, Memory: 50.0, Disk: 50.0}
	err = svc.RecordMetrics("alert-node", highCPU)
	assert.NoError(t, err)

	// We MUST manually evaluate here because RecordMetrics launches Evaluate in a goroutine
	// which might not finish before our assertion in this test environment.
	// But RecordMetrics calls Evaluate async.
	// To test reliably, we should wait or mock.
	// For simplicity, we call Evaluate explicitly again synchronously.
	svc.alertSvc.Evaluate(node.ID, highCPU)

	alerts, err = svc.alertSvc.GetActiveAlerts(node.ID)
	assert.NoError(t, err)
	if assert.NotEmpty(t, alerts) {
		assert.Equal(t, "High CPU", alerts[0].Rule.Name)
		assert.Equal(t, 95.0, alerts[0].Value)
		assert.Equal(t, "firing", alerts[0].Status)
	}

	// 3. Resolve alert
	normal2 := &models.Metric{CPU: 50.0, Memory: 50.0, Disk: 50.0}
	err = svc.RecordMetrics("alert-node", normal2)
	assert.NoError(t, err)
	svc.alertSvc.Evaluate(node.ID, normal2)

	alerts, err = svc.alertSvc.GetActiveAlerts(node.ID)
	assert.NoError(t, err)
	assert.Len(t, alerts, 0)

	// Check DB for resolved record
	var resolved models.Alert
	err = db.Where("node_id = ? AND status = ?", node.ID, "resolved").First(&resolved).Error
	assert.NoError(t, err)
	assert.NotNil(t, resolved.ResolvedAt)
}

func TestAlerting_CustomRule(t *testing.T) {
	db := setupAlertTestDB(t)
	svc := NewIdentityService(db)
	node, _ := svc.CreateNode("custom-rule-node")

	// Create custom rule: Memory > 10%
	rule := models.AlertRule{
		Name:      "Low Memory Test",
		Metric:    "memory",
		Operator:  ">",
		Threshold: 10.0,
		Enabled:   true,
	}
	_, err := svc.alertSvc.CreateRule(rule)
	assert.NoError(t, err)

	// Trigger it
	m := &models.Metric{CPU: 1.0, Memory: 15.0, Disk: 1.0}
	svc.alertSvc.Evaluate(node.ID, m)

	alerts, _ := svc.alertSvc.GetActiveAlerts(node.ID)
	assert.NotEmpty(t, alerts)
	found := false
	for _, a := range alerts {
		if a.Rule.Name == "Low Memory Test" {
			found = true
			break
		}
	}
	assert.True(t, found)
}
