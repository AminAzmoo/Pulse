package models

import "time"

const (
	CertStatusActive  = "active"
	CertStatusRevoked = "revoked"
)

type Certificate struct {
	ID        uint   `gorm:"primaryKey"`
	Identity  string `gorm:"index"`
	Status    string `gorm:"not null;default:'active'"`
	IssuedAt  time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}
