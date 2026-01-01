package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model
	FirebaseUID string `gorm:"uniqueIndex;not null"`
	Email       string `gorm:"uniqueIndex;not null"`
	Name        string
	LastLogin   time.Time
}
