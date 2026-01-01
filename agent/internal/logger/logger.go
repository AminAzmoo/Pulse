package logger

import (
    "go.uber.org/zap"
)

var L *zap.Logger

func Setup() error {
    lg, err := zap.NewProduction()
    if err != nil {
        return err
    }
    L = lg
    return nil
}

func SanitizeSecret(in string) string {
    if in == "" {
        return ""
    }
    return "[REDACTED]"
}

