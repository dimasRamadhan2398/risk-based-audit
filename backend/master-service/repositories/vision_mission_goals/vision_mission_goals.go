package vision_mission_goals

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IVisionMissionGoalsRepository = VisionMissionGoalsRepositoryInterface

// VisionMissionGoalsRepositoryInterface defines the VMG repository interface
type VisionMissionGoalsRepositoryInterface interface {
	Create(vmg *models.VisionMissionGoals) error
	Update(vmg *models.VisionMissionGoals) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.VisionMissionGoals, error)
	FindAll() ([]*models.VisionMissionGoals, error)
	FindMany(offset, limit int, search string) ([]*models.VisionMissionGoals, error)
	Count(search string) (int64, error)
	FindByCompanyID(companyID uuid.UUID) ([]*models.VisionMissionGoals, error)

	// Goal operations
	CreateGoal(goal *models.VmgGoal) error
	UpdateGoal(goal *models.VmgGoal) error
	DeleteGoal(id uuid.UUID) error
	DeleteGoalsByVmgID(vmgID uuid.UUID) error
}

// VisionMissionGoalsRepository handles VMG data operations
type VisionMissionGoalsRepository struct {
	*repositories.BaseRepository
}

var _ IVisionMissionGoalsRepository = (*VisionMissionGoalsRepository)(nil)

// NewVisionMissionGoalsRepository creates a new VMG repository
func NewVisionMissionGoalsRepository(db *gorm.DB) IVisionMissionGoalsRepository {
	return &VisionMissionGoalsRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

// Create creates a new VMG record
func (r *VisionMissionGoalsRepository) Create(vmg *models.VisionMissionGoals) error {
	return r.GetDB().Create(vmg).Error
}

// Update updates an existing VMG record
func (r *VisionMissionGoalsRepository) Update(vmg *models.VisionMissionGoals) error {
	return r.GetDB().Save(vmg).Error
}

// Delete deletes a VMG record (soft delete via GORM)
func (r *VisionMissionGoalsRepository) Delete(id uuid.UUID) error {
	return r.GetDB().Delete(&models.VisionMissionGoals{ID: id}).Error
}

// FindByID finds a VMG by ID with related Company and Goals
func (r *VisionMissionGoalsRepository) FindByID(id uuid.UUID) (*models.VisionMissionGoals, error) {
	var vmg models.VisionMissionGoals
	if err := r.GetDB().
		Preload("Company").
		Preload("Goals").
		First(&vmg, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &vmg, nil
}

// FindAll finds all VMG records
func (r *VisionMissionGoalsRepository) FindAll() ([]*models.VisionMissionGoals, error) {
	var vmgs []*models.VisionMissionGoals
	if err := r.GetDB().
		Preload("Company").
		Preload("Goals").
		Find(&vmgs).Error; err != nil {
		return nil, err
	}
	return vmgs, nil
}

// FindMany finds VMG records with pagination
func (r *VisionMissionGoalsRepository) FindMany(offset, limit int, search string) ([]*models.VisionMissionGoals, error) {
	var vmgs []*models.VisionMissionGoals
	query := r.GetDB().Model(&models.VisionMissionGoals{}).
		Preload("Company").
		Preload("Goals")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"period LIKE ? OR vision LIKE ? OR mission LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&vmgs).Error; err != nil {
		return nil, err
	}
	return vmgs, nil
}

// Count counts VMG records with optional search filter
func (r *VisionMissionGoalsRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.VisionMissionGoals{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"period LIKE ? OR vision LIKE ? OR mission LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindByCompanyID finds VMG records by company ID
func (r *VisionMissionGoalsRepository) FindByCompanyID(companyID uuid.UUID) ([]*models.VisionMissionGoals, error) {
	var vmgs []*models.VisionMissionGoals
	if err := r.GetDB().
		Preload("Company").
		Preload("Goals").
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&vmgs).Error; err != nil {
		return nil, err
	}
	return vmgs, nil
}

// CreateGoal creates a new goal record
func (r *VisionMissionGoalsRepository) CreateGoal(goal *models.VmgGoal) error {
	return r.GetDB().Create(goal).Error
}

// UpdateGoal updates an existing goal record
func (r *VisionMissionGoalsRepository) UpdateGoal(goal *models.VmgGoal) error {
	return r.GetDB().Save(goal).Error
}

// DeleteGoal deletes a goal record
func (r *VisionMissionGoalsRepository) DeleteGoal(id uuid.UUID) error {
	return r.GetDB().Delete(&models.VmgGoal{ID: id}).Error
}

// DeleteGoalsByVmgID deletes all goals for a VMG
func (r *VisionMissionGoalsRepository) DeleteGoalsByVmgID(vmgID uuid.UUID) error {
	return r.GetDB().Where("vmg_id = ?", vmgID).Delete(&models.VmgGoal{}).Error
}
