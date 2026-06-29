package mitigation_action
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/mitigation_action"
	"github.com/google/uuid"
)
type MitigationActionServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.MitigationAction, error)
	FindById(ctx *base.BaseService, id string) (*models.MitigationAction, error)
	Create(ctx *base.BaseService, action *models.MitigationAction) (*models.MitigationAction, error)
	Update(ctx *base.BaseService, id string, action *models.MitigationAction) (*models.MitigationAction, error)
	Delete(ctx *base.BaseService, id string) error
}
type MitigationActionService struct { actionRepo repo.IMitigationActionRepository }
func NewMitigationActionService(actionRepo repo.IMitigationActionRepository) MitigationActionServiceInterface { return &MitigationActionService{actionRepo: actionRepo} }
func (s *MitigationActionService) Create(ctx *base.BaseService, action *models.MitigationAction) (*models.MitigationAction, error) {
	if _, err := s.actionRepo.FindByCode(action.ActionCode); err == nil {
		return nil, apperrors.Wrap("ACTION_CODE_ALREADY_EXISTS", "Mitigation action code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate action code", 500, err)
	}
	if err := s.actionRepo.Create(action); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create mitigation action", 500, err) }
	return s.actionRepo.FindByID(action.ID)
}
func (s *MitigationActionService) Delete(ctx *base.BaseService, id string) error {
	actionID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_ACTION_ID", "Invalid action ID format", 400, err) }
	if _, err := s.actionRepo.FindByID(actionID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find action", 500, err)
	}
	if err := s.actionRepo.Delete(actionID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete action", 500, err) }
	return nil
}
func (s *MitigationActionService) FindAll(ctx *base.BaseService) (*[]models.MitigationAction, error) {
	actions, err := s.actionRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch mitigation actions", 500, err) }
	result := make([]models.MitigationAction, 0, len(actions))
	for _, action := range actions { if action != nil { result = append(result, *action) } }
	return &result, nil
}
func (s *MitigationActionService) FindById(ctx *base.BaseService, id string) (*models.MitigationAction, error) {
	actionID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ACTION_ID", "Invalid action ID format", 400, err) }
	action, err := s.actionRepo.FindByID(actionID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch mitigation action", 500, err)
	}
	return action, nil
}
func (s *MitigationActionService) Update(ctx *base.BaseService, id string, action *models.MitigationAction) (*models.MitigationAction, error) {
	actionID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ACTION_ID", "Invalid action ID format", 400, err) }
	existingAction, err := s.actionRepo.FindByID(actionID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch mitigation action", 500, err)
	}
	if existingAction.ActionCode != action.ActionCode {
		if _, err := s.actionRepo.FindByCode(action.ActionCode); err == nil {
			return nil, apperrors.Wrap("ACTION_CODE_ALREADY_EXISTS", "Mitigation action code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate action code", 500, err)
		}
	}
	existingAction.ActionCode = action.ActionCode
	existingAction.RiskRegisterID = action.RiskRegisterID
	existingAction.ControlID = action.ControlID
	existingAction.IssueID = action.IssueID
	existingAction.ActionTitle = action.ActionTitle
	existingAction.Description = action.Description
	existingAction.ActionType = action.ActionType
	existingAction.Status = action.Status
	existingAction.Progress = action.Progress
	existingAction.CompletionPercent = action.CompletionPercent
	existingAction.OwnerID = action.OwnerID
	existingAction.DepartmentID = action.DepartmentID
	existingAction.BudgetEstimate = action.BudgetEstimate
	existingAction.ActualCost = action.ActualCost
	existingAction.PlannedStartDate = action.PlannedStartDate
	existingAction.PlannedEndDate = action.PlannedEndDate
	existingAction.ActualStartDate = action.ActualStartDate
	existingAction.ActualEndDate = action.ActualEndDate
	existingAction.VerifiedByID = action.VerifiedByID
	existingAction.VerifiedAt = action.VerifiedAt
	existingAction.VerificationResult = action.VerificationResult
	if err := s.actionRepo.Update(existingAction); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update mitigation action", 500, err) }
	return s.actionRepo.FindByID(actionID)
}
