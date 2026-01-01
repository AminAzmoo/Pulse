package migrations

import (
	"pulse/backend/internal/db/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Admin{},
		&models.Node{},
		&models.Certificate{},
		&models.InstallToken{},
		&models.Event{},
		&models.Task{},
		&models.TaskAttempt{},
		&models.IdempotencyRecord{},
		&models.Metric{},
		&models.AlertRule{},
		&models.Alert{},
	)
}
