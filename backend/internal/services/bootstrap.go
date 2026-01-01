package services

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
	"pulse/backend/pkg/external/logging"
)

func (s *IdentityService) GenerateInstallToken(nodeID uint) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	installToken := models.InstallToken{
		ID:        uuid.New(),
		NodeID:    nodeID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&installToken).Error; err != nil {
		return "", err
	}

	metadata := map[string]interface{}{"node_id": nodeID, "token_id": installToken.ID}
	logging.Audit("Install token generated", metadata)

	return token, nil
}

func (s *IdentityService) ValidateInstallToken(token string) (uint, error) {
	var installToken models.InstallToken
	if err := s.db.Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).First(&installToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logging.Logger.Warn("Invalid or expired install token", zap.String("token", "[REDACTED]"))
			metadata := map[string]interface{}{"token": "[REDACTED]"}
			logging.Audit("Install token validation failed", metadata)
		}
		return 0, err
	}

	installToken.Used = true
	if err := s.db.Save(&installToken).Error; err != nil {
		return 0, err
	}

	metadata := map[string]interface{}{"node_id": installToken.NodeID, "token_id": installToken.ID}
	logging.Audit("Install token validated and invalidated", metadata)

	return installToken.NodeID, nil
}
