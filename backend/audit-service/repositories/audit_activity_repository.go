package repositories

import (
	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditActivityRepositoryInterface defines the audit activity repository interface
type AuditActivityRepositoryInterface interface {
	Create(activity *models.ActivityPlan) error
	Update(activity *models.ActivityPlan) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.ActivityPlan, error)
	FindByProjectCode(projectCode string) (*models.ActivityPlan, error)
	FindMany(offset, limit int, search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) ([]*models.ActivityPlan, error)
	Count(search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) (int64, error)
}

// AuditActivityRepository handles audit activity data operations
type AuditActivityRepository struct {
	*BaseRepository
}

// NewAuditActivityRepository creates a new audit activity repository
func NewAuditActivityRepository(baseRepo *BaseRepository) AuditActivityRepositoryInterface {
	return &AuditActivityRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new audit activity
func (r *AuditActivityRepository) Create(activity *models.ActivityPlan) error {
	if err := r.DB.Create(activity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Update updates an audit activity
func (r *AuditActivityRepository) Update(activity *models.ActivityPlan) error {
	if err := r.DB.Save(activity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Delete deletes an audit activity (soft delete)
func (r *AuditActivityRepository) Delete(id uuid.UUID) error {
	result := r.DB.Delete(&models.ActivityPlan{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.ErrDatabase
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// FindByID finds an audit activity by ID
func (r *AuditActivityRepository) FindByID(id uuid.UUID) (*models.ActivityPlan, error) {
	var activity models.ActivityPlan
	if err := r.DB.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// FindByProjectCode finds an audit activity by project code
func (r *AuditActivityRepository) FindByProjectCode(projectCode string) (*models.ActivityPlan, error) {
	var activity models.ActivityPlan
	if err := r.DB.Where("project_code = ?", projectCode).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// FindMany finds multiple audit activities with filters
func (r *AuditActivityRepository) FindMany(offset, limit int, search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) ([]*models.ActivityPlan, error) {
	var activities []*models.ActivityPlan
	query := r.DB.Model(&models.ActivityPlan{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("project_code ILIKE ? OR title ILIKE ? OR objective ILIKE ? OR scope ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if annualPlanID != nil {
		query = query.Where("annual_plan_id = ?", *annualPlanID)
	}

	if targetUnitID != nil {
		query = query.Where("target_unit_id = ?", *targetUnitID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error; err != nil {
		return nil, err
	}

	return activities, nil
}

// Count counts audit activities with filters
func (r *AuditActivityRepository) Count(search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) (int64, error) {
	var count int64
	query := r.DB.Model(&models.ActivityPlan{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("project_code ILIKE ? OR title ILIKE ? OR objective ILIKE ? OR scope ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if annualPlanID != nil {
		query = query.Where("annual_plan_id = ?", *annualPlanID)
	}

	if targetUnitID != nil {
		query = query.Where("target_unit_id = ?", *targetUnitID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
