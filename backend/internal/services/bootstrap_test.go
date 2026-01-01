package services

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	_ = db.AutoMigrate(&models.InstallToken{}, &models.Node{}, &models.AlertRule{}, &models.Alert{})
	return db
}

func TestGenerateInstallToken(t *testing.T) {
	db := setupTestDB()
	svc := NewIdentityService(db)

	node := models.Node{Identity: "test-node"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}

	token, err := svc.GenerateInstallToken(node.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token, got empty string")
	}

	var installToken models.InstallToken
	if err := db.First(&installToken, "node_id = ?", node.ID).Error; err != nil {
		t.Fatal(err)
	}

	if installToken.Token != token {
		t.Errorf("Expected %s, got %s", token, installToken.Token)
	}
	if installToken.Used {
		t.Error("Token should not be marked as used initially")
	}
	if time.Since(installToken.CreatedAt) > time.Second {
		t.Error("CreatedAt not set correctly")
	}
	if installToken.ExpiresAt.Sub(installToken.CreatedAt) != 24*time.Hour {
		t.Error("ExpiresAt not set to 24 hours")
	}
}

func TestValidateInstallToken(t *testing.T) {
	db := setupTestDB()
	svc := NewIdentityService(db)

	node := models.Node{Identity: "test-node"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}

	token, _ := svc.GenerateInstallToken(node.ID)

	// Valid token
	valid, err := svc.ValidateInstallToken(token)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if valid == 0 {
		t.Error("Expected valid token")
	}

	// Replay attack
	valid, err = svc.ValidateInstallToken(token)
	if err == nil {
		t.Error("Expected error for used token")
	}
	if valid != 0 {
		t.Error("Expected invalid for replay")
	}

	// Check audit event - assuming logging.Audit is called on rejection
	// (Test logging separately if needed)
}

func TestExpiredInstallToken(t *testing.T) {
	db := setupTestDB()
	svc := NewIdentityService(db)

	node := models.Node{Identity: "test-node"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}

	tokenID := uuid.New()
	expiredToken := models.InstallToken{
		ID:        tokenID,
		NodeID:    node.ID,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}
	if err := db.Create(&expiredToken).Error; err != nil {
		t.Fatal(err)
	}

	valid, err := svc.ValidateInstallToken("expired-token")
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if valid != 0 {
		t.Error("Expected invalid for expired")
	}
}
