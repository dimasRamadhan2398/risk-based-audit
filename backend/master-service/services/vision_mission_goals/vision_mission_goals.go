package vision_mission_goals

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	vmgsRepo "master-service/repositories/vision_mission_goals"

	"github.com/google/uuid"
)

type IVisionMissionGoalsRepository = vmgsRepo.VisionMissionGoalsRepositoryInterface

type VisionMissionGoalsServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.VisionMissionGoals, error)
	FindById(ctx *base.BaseService, id string) (*models.VisionMissionGoals, error)
	FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.VisionMissionGoals, int64, error)
	FindByCompanyId(ctx *base.BaseService, companyId string) (*[]models.VisionMissionGoals, error)
	Create(ctx *base.BaseService, vmg *models.VisionMissionGoals, createdBy string) (*models.VisionMissionGoals, error)
	Update(ctx *base.BaseService, id string, vmg *models.VisionMissionGoals, modifiedBy string) (*models.VisionMissionGoals, error)
	Delete(ctx *base.BaseService, id string) error
}

type VisionMissionGoalsService struct {
	repo IVisionMissionGoalsRepository
}

var _ VisionMissionGoalsServiceInterface = (*VisionMissionGoalsService)(nil)

func NewVisionMissionGoalsService(repo IVisionMissionGoalsRepository) VisionMissionGoalsServiceInterface {
	return &VisionMissionGoalsService{repo: repo}
}

// FindAll retrieves all VMG records
func (s *VisionMissionGoalsService) FindAll(ctx *base.BaseService) (*[]models.VisionMissionGoals, error) {
	vmgs, err := s.repo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch Vision, Mission & Goals", 500, err)
	}

	result := make([]models.VisionMissionGoals, 0, len(vmgs))
	for _, vmg := range vmgs {
		if vmg != nil {
			result = append(result, *vmg)
		}
	}
	return &result, nil
}

// FindById retrieves a VMG record by ID
func (s *VisionMissionGoalsService) FindById(ctx *base.BaseService, id string) (*models.VisionMissionGoals, error) {
	vmgID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_ID", "Invalid VMG ID format", 400, err)
	}

	vmg, err := s.repo.FindByID(vmgID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch Vision, Mission & Goals", 500, err)
	}

	return vmg, nil
}

// FindMany retrieves VMG records with pagination
func (s *VisionMissionGoalsService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.VisionMissionGoals, int64, error) {
	vmgs, err := s.repo.FindMany(offset, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch Vision, Mission & Goals", 500, err)
	}

	count, err := s.repo.Count(search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to count Vision, Mission & Goals", 500, err)
	}

	result := make([]models.VisionMissionGoals, 0, len(vmgs))
	for _, vmg := range vmgs {
		if vmg != nil {
			result = append(result, *vmg)
		}
	}
	return &result, count, nil
}

// FindByCompanyId retrieves VMG records by company ID
func (s *VisionMissionGoalsService) FindByCompanyId(ctx *base.BaseService, companyId string) (*[]models.VisionMissionGoals, error) {
	companyID, err := uuid.Parse(companyId)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_ID", "Invalid Company ID format", 400, err)
	}

	vmgs, err := s.repo.FindByCompanyID(companyID)
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch Vision, Mission & Goals by company", 500, err)
	}

	result := make([]models.VisionMissionGoals, 0, len(vmgs))
	for _, vmg := range vmgs {
		if vmg != nil {
			result = append(result, *vmg)
		}
	}
	return &result, nil
}

// Create creates a new VMG record with goals
func (s *VisionMissionGoalsService) Create(ctx *base.BaseService, vmg *models.VisionMissionGoals, createdBy string) (*models.VisionMissionGoals, error) {
	// Validate required fields
	if vmg.CompanyID == uuid.Nil {
		return nil, apperrors.Wrap("VALIDATION_ERROR", "Company ID is required", 400, nil)
	}
	if vmg.Period == "" {
		return nil, apperrors.Wrap("VALIDATION_ERROR", "Period is required", 400, nil)
	}

	// Set audit fields
	vmg.CreatedBy = createdBy
	vmg.ModifiedBy = createdBy
	if vmg.Status == "" {
		vmg.Status = models.VmgStatusDraft
	}
	if vmg.Version == "" {
		vmg.Version = "v1.0"
	}

	// Create the main VMG record
	if err := s.repo.Create(vmg); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create Vision, Mission & Goals", 500, err)
	}

	// Create goals if provided
	for i := range vmg.Goals {
		vmg.Goals[i].VmgID = vmg.ID
		vmg.Goals[i].ID = uuid.New()
		if err := s.repo.CreateGoal(&vmg.Goals[i]); err != nil {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create goal", 500, err)
		}
	}

	return vmg, nil
}

// Update updates an existing VMG record with goals
func (s *VisionMissionGoalsService) Update(ctx *base.BaseService, id string, vmg *models.VisionMissionGoals, modifiedBy string) (*models.VisionMissionGoals, error) {
	vmgID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_ID", "Invalid VMG ID format", 400, err)
	}

	// Check if record exists
	existing, err := s.repo.FindByID(vmgID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch Vision, Mission & Goals", 500, err)
	}

	// Update main fields
	existing.Period = vmg.Period
	existing.EffectiveDate = vmg.EffectiveDate
	existing.Vision = vmg.Vision
	existing.Mission = vmg.Mission
	existing.Version = vmg.Version
	existing.Status = vmg.Status
	existing.Notes = vmg.Notes
	existing.ModifiedBy = modifiedBy

	// Update the main record
	if err := s.repo.Update(existing); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update Vision, Mission & Goals", 500, err)
	}

	// Delete existing goals and create new ones
	if err := s.repo.DeleteGoalsByVmgID(vmgID); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to delete existing goals", 500, err)
	}

	for i := range vmg.Goals {
		vmg.Goals[i].VmgID = vmgID
		vmg.Goals[i].ID = uuid.New()
		if err := s.repo.CreateGoal(&vmg.Goals[i]); err != nil {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create goal", 500, err)
		}
	}

	// Fetch updated record with relations
	updated, err := s.repo.FindByID(vmgID)
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch updated record", 500, err)
	}

	return updated, nil
}

// Delete deletes a VMG record and its goals
func (s *VisionMissionGoalsService) Delete(ctx *base.BaseService, id string) error {
	vmgID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_ID", "Invalid VMG ID format", 400, err)
	}

	// Check if record exists
	if _, err := s.repo.FindByID(vmgID); err != nil {
		return err
	}

	// Delete goals first
	if err := s.repo.DeleteGoalsByVmgID(vmgID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete goals", 500, err)
	}

	// Delete main record
	if err := s.repo.Delete(vmgID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete Vision, Mission & Goals", 500, err)
	}

	return nil
}
