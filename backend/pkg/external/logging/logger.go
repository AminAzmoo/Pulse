package logging

import (
    "encoding/json"
    "os"
	"strings"

	models "pulse/backend/internal/db/models"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
)

var Logger *zap.Logger = zap.NewNop()
var db *gorm.DB

var sensitiveKeys = []string{"password", "token", "secret", "key", "credential", "auth"}

func InitLogger(logPath string, database *gorm.DB) error {
	db = database

	// File Writer
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})

	// Console Writer (stdout)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Combine writers based on environment
	var core zapcore.Core
	
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	
	// In development, we might want a pretty console encoder
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	if os.Getenv("GO_ENV") == "development" {
		// Log to both file (JSON) and console (Pretty)
		core = zapcore.NewTee(
			zapcore.NewCore(fileEncoder, fileWriter, zap.InfoLevel),
			zapcore.NewCore(consoleEncoder, consoleWriter, zap.DebugLevel),
		)
	} else {
		// Production: Log to file (JSON) only (or stdout JSON if containerized, but sticking to request)
		core = zapcore.NewCore(fileEncoder, fileWriter, zap.InfoLevel)
	}

	Logger = zap.New(core, zap.AddCaller())
	return nil
}

func Audit(message string, metadata map[string]interface{}) {
    jsonMetadata := sanitizeMetadata(metadata)
    Logger.Info(message, zap.String("metadata", jsonMetadata), zap.String("type", "audit"))

	// Persist to DB
    if db != nil {
        event := models.Event{
            Level:    models.EventLevelAudit,
            Message:  message,
            Metadata: jsonMetadata,
        }
        if err := db.Create(&event).Error; err != nil {
            Logger.Error("Failed to persist audit event", zap.Error(err))
        }
    }
}

func sanitizeMetadata(metadata map[string]interface{}) string {
	sanitized := make(map[string]interface{})
	for k, v := range metadata {
		keyLower := strings.ToLower(k)
		isSensitive := false
		for _, sk := range sensitiveKeys {
			if strings.Contains(keyLower, sk) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}
	jsonStr, _ := json.Marshal(sanitized)
	return string(jsonStr)
}
