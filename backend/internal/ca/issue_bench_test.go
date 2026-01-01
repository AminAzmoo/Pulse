package ca

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pulse/backend/internal/db/models"
)

const benchNodeIdentity = "bench-node"

func BenchmarkIssueCertificate(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Certificate{})
	svc := NewCAService(db)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = svc.IssueCertificate(benchNodeIdentity)
	}
}
