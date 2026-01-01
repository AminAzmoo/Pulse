package models

import "time"

type AlertRule struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"not null"`
	Metric    string  `gorm:"not null"` // "cpu", "memory", "disk"
	Operator  string  `gorm:"not null"` // ">", "<", ">=", "<="
	Threshold float64 `gorm:"not null"`
	Duration  string  `gorm:"not null"`                   // e.g., "5m" (trigger if violated for 5 mins) - MVP: ignore duration, instant trigger
	Severity  string  `gorm:"not null;default:'warning'"` // "info", "warning", "critical"
	Enabled   bool    `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Alert struct {
	ID          uint      `gorm:"primaryKey"`
	NodeID      uint      `gorm:"index;not null"`
	RuleID      uint      `gorm:"index;not null"`
	Rule        AlertRule `gorm:"foreignKey:RuleID"`
	Value       float64   `gorm:"not null"`
	Status      string    `gorm:"not null;default:'firing'"` // "firing", "resolved"
	TriggeredAt time.Time `gorm:"index;not null"`
	ResolvedAt  *time.Time
}
