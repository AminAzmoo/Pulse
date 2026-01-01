package models

import "time"

const (
	EventLevelInfo  = "info"
	EventLevelWarn  = "warn"
	EventLevelError = "error"
	EventLevelAudit = "audit"
)

type Event struct {
	ID        uint      `gorm:"primaryKey"`
	Level     string    `gorm:"not null"`
	Message   string    `gorm:"not null"`
	Metadata  string    `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"index"`
}
