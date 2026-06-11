package base

import (
	"audit-service/pkg/logger"
	"context"

	"go.uber.org/zap"
)

// BaseService provides common service operations
type BaseService struct {
	ctx context.Context
}

// NewBaseService creates a new base service
func NewBaseService() *BaseService {
	return &BaseService{
		ctx: context.Background(),
	}
}

// WithContext returns a new service with context
func (s *BaseService) WithContext(ctx context.Context) *BaseService {
	return &BaseService{ctx: ctx}
}

// GetContext returns the current context
func (s *BaseService) GetContext() context.Context {
	return s.ctx
}

// LogInfo logs an info message
func (s *BaseService) LogInfo(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// LogError logs an error message
func (s *BaseService) LogError(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// LogWarning logs a warning message
func (s *BaseService) LogWarning(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// LogDebug logs a debug message
func (s *BaseService) LogDebug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}