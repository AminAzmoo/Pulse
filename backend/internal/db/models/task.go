package models

import (
    "time"
)

const (
    TaskStatusPending             = "pending"
    TaskStatusInProgress          = "in_progress"
    TaskStatusAwaitingConfirmation = "awaiting_confirmation"
    TaskStatusCompleted           = "completed"
    TaskStatusFailed              = "failed"

    TaskClassHeavy = "heavy"
    TaskClassLight = "light"
)

type Task struct {
    ID              uint      `gorm:"primaryKey"`
    Identity        string    `gorm:"index;not null"`
    Type            string    `gorm:"not null"`
    Class           string    `gorm:"not null"`
    IdempotencyKey  string    `gorm:"uniqueIndex;not null"`
    Status          string    `gorm:"not null"`
    PayloadJSON     string    `gorm:"type:jsonb"`
    TimeoutSeconds  int       `gorm:"not null"`
    CreatedAt       time.Time `gorm:"index"`
    UpdatedAt       time.Time `gorm:"index"`
}

type TaskAttempt struct {
    ID        uint      `gorm:"primaryKey"`
    TaskID    uint      `gorm:"index;not null"`
    Status    string    `gorm:"not null"`
    Message   string    `gorm:"type:text"`
    StartedAt time.Time `gorm:"index"`
    EndedAt   *time.Time
}

type IdempotencyRecord struct {
    ID              uint      `gorm:"primaryKey"`
    IdempotencyKey  string    `gorm:"uniqueIndex;not null"`
    LastStatus      string    `gorm:"not null"`
    LastOutcomeJSON string    `gorm:"type:jsonb"`
    UpdatedAt       time.Time `gorm:"index"`
}
