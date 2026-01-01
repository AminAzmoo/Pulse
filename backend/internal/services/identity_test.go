package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

func setupIdentityTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&models.Node{}, &models.Metric{}, &models.AlertRule{}, &models.Alert{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func TestIdentityService_GetNode(t *testing.T) {
	db := setupIdentityTestDB(t)
	svc := NewIdentityService(db)

	// Create a node
	node, err := svc.CreateNode("test-node-1")
	assert.NoError(t, err)

	// Update info
	info := &models.Node{
		Hostname:    "test-host",
		OS:          "linux",
		IPAddresses: "127.0.0.1",
	}
	err = svc.UpdateHostInfo("test-node-1", info)
	assert.NoError(t, err)

	// Get node
	fetched, err := svc.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test-host", fetched.Hostname)
	assert.Equal(t, "linux", fetched.OS)
}

func TestIdentityService_Metrics(t *testing.T) {
	db := setupIdentityTestDB(t)
	svc := NewIdentityService(db)

	node, _ := svc.CreateNode("test-node-metrics")

	// Record metrics
	m1 := &models.Metric{CPU: 10.5, Memory: 20.0, Disk: 30.0}
	err := svc.RecordMetrics("test-node-metrics", m1)
	assert.NoError(t, err)

	// Wait a bit or ensure timestamp diff if needed, but for now just order check
	time.Sleep(10 * time.Millisecond)

	m2 := &models.Metric{CPU: 15.0, Memory: 25.0, Disk: 35.0}
	err = svc.RecordMetrics("test-node-metrics", m2)
	assert.NoError(t, err)

	// Get metrics
	metrics, err := svc.GetNodeMetrics(node.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	// Should be descending order by timestamp
	assert.Equal(t, 15.0, metrics[0].CPU)
	assert.Equal(t, 10.5, metrics[1].CPU)
}
