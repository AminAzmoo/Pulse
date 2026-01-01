package models

import (
	"time"

	"github.com/google/uuid"
)

type InstallToken struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	NodeID    uint      `gorm:"index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	Used      bool      `gorm:"default:false"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
