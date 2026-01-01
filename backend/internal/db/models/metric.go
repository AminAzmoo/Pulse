package models

import "time"

type Metric struct {
	ID        uint      `gorm:"primaryKey"`
	NodeID    uint      `gorm:"index;not null"`
	CPU       float64   `gorm:"type:decimal(5,2)"` // Percentage 0-100.00
	Memory    float64   `gorm:"type:decimal(5,2)"` // Percentage 0-100.00
	Disk      float64   `gorm:"type:decimal(5,2)"` // Percentage 0-100.00
	Timestamp time.Time `gorm:"index;not null"`
}
