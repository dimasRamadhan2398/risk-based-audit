package repositories

import (
	"strings"

	"audit-service/models"
	"audit-service/pkg/activitycode"
	apperrors "audit-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditActivityRepositoryInterface defines the audit activity repository interface
type AuditActivityRepositoryInterface interface {
	Create(activity *models.AuditActivity) error
	CreateWithGeneratedCode(activity *models.AuditActivity, year int) error
	AnnualPlanYear(annualPlanID uuid.UUID) (int, error)
	Update(activity *models.AuditActivity) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditActivity, error)
	FindByProjectCode(projectCode string) (*models.AuditActivity, error)
	FindMany(offset, limit int, search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) ([]*models.AuditActivity, error)
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
func (r *AuditActivityRepository) Create(activity *models.AuditActivity) error {
	if err := r.DB.Create(activity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// codeAttempts bounds retries when a generated project code collides with a
// row written outside the counter at the same moment (e.g. a seeder run).
const codeAttempts = 3

// CreateWithGeneratedCode sets activity.ProjectCode to the next Activity ID for
// activity.AuditType and year and inserts the row, in one transaction. The
// counter row stays locked until commit, so concurrent creates get distinct
// numbers; on a unique violation the whole transaction is retried.
func (r *AuditActivityRepository) CreateWithGeneratedCode(activity *models.AuditActivity, year int) error {
	var err error
	for attempt := 0; attempt < codeAttempts; attempt++ {
		err = r.DB.Transaction(func(tx *gorm.DB) error {
			code, err := activitycode.Next(tx, activity.AuditType, year)
			if err != nil {
				return err
			}
			activity.ProjectCode = code
			return tx.Create(activity).Error
		})
		if err == nil {
			return nil
		}
		activity.ProjectCode = ""
		if !isUniqueViolation(err) {
			break
		}
	}
	return apperrors.Wrap(apperrors.ErrDatabase.Code, apperrors.ErrDatabase.Message, apperrors.ErrDatabase.StatusCode, err)
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") || strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

// AnnualPlanYear returns the year of an annual audit plan, or ErrNotFound.
func (r *AuditActivityRepository) AnnualPlanYear(annualPlanID uuid.UUID) (int, error) {
	var plan models.AuditAnnual
	if err := r.DB.Select("id", "year").First(&plan, "id = ?", annualPlanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, apperrors.ErrNotFound
		}
		return 0, apperrors.ErrDatabase
	}
	return plan.Year, nil
}

// Update updates an audit activity
func (r *AuditActivityRepository) Update(activity *models.AuditActivity) error {
	if err := r.DB.Save(activity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Delete deletes an audit activity (soft delete)
func (r *AuditActivityRepository) Delete(id uuid.UUID) error {
	result := r.DB.Delete(&models.AuditActivity{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.ErrDatabase
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// FindByID finds an audit activity by ID
func (r *AuditActivityRepository) FindByID(id uuid.UUID) (*models.AuditActivity, error) {
	var activity models.AuditActivity
	if err := r.DB.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// FindByProjectCode finds an audit activity by project code
func (r *AuditActivityRepository) FindByProjectCode(projectCode string) (*models.AuditActivity, error) {
	var activity models.AuditActivity
	if err := r.DB.Where("project_code = ?", projectCode).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// FindMany finds multiple audit activities with filters
func (r *AuditActivityRepository) FindMany(offset, limit int, search string, annualPlanID *uuid.UUID, targetUnitID *uuid.UUID, status *string) ([]*models.AuditActivity, error) {
	var activities []*models.AuditActivity
	query := r.DB.Model(&models.AuditActivity{})

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
	query := r.DB.Model(&models.AuditActivity{})

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
