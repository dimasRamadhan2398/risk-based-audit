package utils

import "go.uber.org/zap"

// LogField creates a zap field for logging
func LogField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
