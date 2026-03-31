package repositories

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ISysAlertLogRepository is an alias for the sys alert log repository interface
type ISysAlertLogRepository = SysAlertLogRepositoryInterface

// SysAlertLogRepositoryInterface defines the sys alert log repository interface
type SysAlertLogRepositoryInterface interface {
	Create(log *models.SysAlertLog) error
	FindByID(id uuid.UUID) (*models.SysAlertLog, error)
	FindByUserID(userID uuid.UUID, offset, limit int) ([]*models.SysAlertLog, error)
	FindByAction(action models.SysAlertAction, offset, limit int) ([]*models.SysAlertLog, error)
	FindByIPAddress(ipAddress string, offset, limit int) ([]*models.SysAlertLog, error)
	FindRecentByUserID(userID uuid.UUID, limit int) ([]*models.SysAlertLog, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	CountByAction(action models.SysAlertAction) (int64, error)
	DeleteOldLogs(days int) error
}

// SysAlertLogRepository handles system alert log data operations
type SysAlertLogRepository struct {
	*BaseRepository
}

// NewSysAlertLogRepository creates a new sys alert log repository
func NewSysAlertLogRepository(db *gorm.DB) ISysAlertLogRepository {
	return &SysAlertLogRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new system alert log
func (r *SysAlertLogRepository) Create(log *models.SysAlertLog) error {
	return r.BaseRepository.Create(log)
}

// FindByID finds a system alert log by ID
func (r *SysAlertLogRepository) FindByID(id uuid.UUID) (*models.SysAlertLog, error) {
	var log models.SysAlertLog
	if err := r.GetDB().Preload("User").First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

// FindByUserID finds system alert logs by user ID
func (r *SysAlertLogRepository) FindByUserID(userID uuid.UUID, offset, limit int) ([]*models.SysAlertLog, error) {
	var logs []*models.SysAlertLog
	if err := r.GetDB().Preload("User").Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByAction finds system alert logs by action
func (r *SysAlertLogRepository) FindByAction(action models.SysAlertAction, offset, limit int) ([]*models.SysAlertLog, error) {
	var logs []*models.SysAlertLog
	if err := r.GetDB().Preload("User").Where("action = ?", action).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByIPAddress finds system alert logs by IP address
func (r *SysAlertLogRepository) FindByIPAddress(ipAddress string, offset, limit int) ([]*models.SysAlertLog, error) {
	var logs []*models.SysAlertLog
	if err := r.GetDB().Preload("User").Where("ip_address = ?", ipAddress).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindRecentByUserID finds recent system alert logs by user ID
func (r *SysAlertLogRepository) FindRecentByUserID(userID uuid.UUID, limit int) ([]*models.SysAlertLog, error) {
	var logs []*models.SysAlertLog
	if err := r.GetDB().Preload("User").Where("user_id = ?", userID).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// CountByUserID counts system alert logs by user ID
func (r *SysAlertLogRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.GetDB().Model(&models.SysAlertLog{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByAction counts system alert logs by action
func (r *SysAlertLogRepository) CountByAction(action models.SysAlertAction) (int64, error) {
	var count int64
	if err := r.GetDB().Model(&models.SysAlertLog{}).Where("action = ?", action).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteOldLogs deletes logs older than specified days
func (r *SysAlertLogRepository) DeleteOldLogs(days int) error {
	result := r.GetDB().Where("created_at < NOW() - INTERVAL '? days'", days).Delete(&models.SysAlertLog{})
	return result.Error
}
