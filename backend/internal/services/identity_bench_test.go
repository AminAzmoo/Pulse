package services

import (
	"strconv"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

const (
	benchNodeCount   = 10000
	benchTargetIndex = 5000
	benchNodePrefix  = "node-"
)

func BenchmarkIsAllowed(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Node{})
	for i := 0; i < benchNodeCount; i++ {
		n := models.Node{Identity: benchNodePrefix + strconv.Itoa(i), Status: models.NodeStatusActive}
		_ = db.Create(&n).Error
	}
	svc := NewIdentityService(db)
	target := benchNodePrefix + strconv.Itoa(benchTargetIndex)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.IsAllowed(target)
	}
}
