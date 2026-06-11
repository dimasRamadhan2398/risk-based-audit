package repositories

import (
	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditAssignmentRepositoryInterface defines the audit assignment repository interface
type AuditAssignmentRepositoryInterface interface {
	Create(assignment *models.AuditAssignment) error
	Update(assignment *models.AuditAssignment) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditAssignment, error)
	FindByAuditorID(auditorID uuid.UUID) ([]*models.AuditAssignment, error)
	FindByAuditPlanID(auditPlanID uuid.UUID) ([]*models.AuditAssignment, error)
	FindMany(offset, limit int, search string, auditorID *uuid.UUID, auditPlanID *uuid.UUID, status *string) ([]*models.AuditAssignment, error)
	Count(search string, auditorID *uuid.UUID, auditPlanID *uuid.UUID, status *string) (int64, error)
}

// AuditAssignmentRepository handles audit assignment data operations
type AuditAssignmentRepository struct {
	*BaseRepository
}

// NewAuditAssignmentRepository creates a new audit assignment repository
func NewAuditAssignmentRepository(baseRepo *BaseRepository) AuditAssignmentRepositoryInterface {
	return &AuditAssignmentRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new audit assignment
func (r *AuditAssignmentRepository) Create(assignment *models.AuditAssignment) error {
	if err := r.DB.Create(assignment).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Update updates an audit assignment
func (r *AuditAssignmentRepository) Update(assignment *models.AuditAssignment) error {
	if err := r.DB.Save(assignment).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Delete deletes an audit assignment (soft delete)
func (r *AuditAssignmentRepository) Delete(id uuid.UUID) error {
	result := r.DB.Delete(&models.AuditAssignment{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.ErrDatabase
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// FindByID finds an audit assignment by ID
func (r *AuditAssignmentRepository) FindByID(id uuid.UUID) (*models.AuditAssignment, error) {
	var assignment models.AuditAssignment
	if err := r.DB.First(&assignment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &assignment, nil
}

// FindByAuditorID finds audit assignments by auditor ID
func (r *AuditAssignmentRepository) FindByAuditorID(auditorID uuid.UUID) ([]*models.AuditAssignment, error) {
	var assignments []*models.AuditAssignment
	if err := r.DB.Where("auditor_id = ?", auditorID).Order("created_at DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// FindByAuditPlanID finds audit assignments by audit plan ID
func (r *AuditAssignmentRepository) FindByAuditPlanID(auditPlanID uuid.UUID) ([]*models.AuditAssignment, error) {
	var assignments []*models.AuditAssignment
	if err := r.DB.Where("audit_plan_id = ?", auditPlanID).Order("created_at DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// FindMany finds multiple audit assignments with filters
func (r *AuditAssignmentRepository) FindMany(offset, limit int, search string, auditorID *uuid.UUID, auditPlanID *uuid.UUID, status *string) ([]*models.AuditAssignment, error) {
	var assignments []*models.AuditAssignment
	query := r.DB.Model(&models.AuditAssignment{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("assignment_title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if auditorID != nil {
		query = query.Where("auditor_id = ?", *auditorID)
	}

	if auditPlanID != nil {
		query = query.Where("audit_plan_id = ?", *auditPlanID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}

	return assignments, nil
}

// Count counts audit assignments with filters
func (r *AuditAssignmentRepository) Count(search string, auditorID *uuid.UUID, auditPlanID *uuid.UUID, status *string) (int64, error) {
	var count int64
	query := r.DB.Model(&models.AuditAssignment{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("assignment_title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if auditorID != nil {
		query = query.Where("auditor_id = ?", *auditorID)
	}

	if auditPlanID != nil {
		query = query.Where("audit_plan_id = ?", *auditPlanID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}