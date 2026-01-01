package models

import "time"

const (
	NodeStatusActive  = "active"
	NodeStatusRevoked = "revoked"
)

type Node struct {
	ID        uint   `gorm:"primaryKey"`
	Identity  string `gorm:"uniqueIndex;not null"`
	Status    string `gorm:"not null;default:'active'"`
	
	// Inventory
	Hostname        string
	OS              string
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	KernelVersion   string
	Arch            string
	AgentVersion    string
	UptimeSeconds   uint64
	IPAddresses     string // JSON array or comma-separated

	LastSeenAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
