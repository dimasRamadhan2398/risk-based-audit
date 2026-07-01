package control
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/control"
	"github.com/google/uuid"
)
type ControlServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Control, error)
	FindById(ctx *base.BaseService, id string) (*models.Control, error)
	Create(ctx *base.BaseService, control *models.Control) (*models.Control, error)
	Update(ctx *base.BaseService, id string, control *models.Control) (*models.Control, error)
	Delete(ctx *base.BaseService, id string) error
}
type ControlService struct { controlRepo repo.IControlRepository }
func NewControlService(controlRepo repo.IControlRepository) ControlServiceInterface { return &ControlService{controlRepo: controlRepo} }
func (s *ControlService) Create(ctx *base.BaseService, control *models.Control) (*models.Control, error) {
	if _, err := s.controlRepo.FindByCode(control.ControlCode); err == nil {
		return nil, apperrors.Wrap("CONTROL_CODE_ALREADY_EXISTS", "Control code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate control code", 500, err)
	}
	if err := s.controlRepo.Create(control); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create control", 500, err) }
	return s.controlRepo.FindByID(control.ID)
}
func (s *ControlService) Delete(ctx *base.BaseService, id string) error {
	controlID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_CONTROL_ID", "Invalid control ID format", 400, err) }
	if _, err := s.controlRepo.FindByID(controlID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find control", 500, err)
	}
	if err := s.controlRepo.Delete(controlID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete control", 500, err) }
	return nil
}
func (s *ControlService) FindAll(ctx *base.BaseService) (*[]models.Control, error) {
	controls, err := s.controlRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch controls", 500, err) }
	result := make([]models.Control, 0, len(controls))
	for _, control := range controls { if control != nil { result = append(result, *control) } }
	return &result, nil
}
func (s *ControlService) FindById(ctx *base.BaseService, id string) (*models.Control, error) {
	controlID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CONTROL_ID", "Invalid control ID format", 400, err) }
	control, err := s.controlRepo.FindByID(controlID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch control", 500, err)
	}
	return control, nil
}
func (s *ControlService) Update(ctx *base.BaseService, id string, control *models.Control) (*models.Control, error) {
	controlID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CONTROL_ID", "Invalid control ID format", 400, err) }
	existingControl, err := s.controlRepo.FindByID(controlID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch control", 500, err)
	}
	if existingControl.ControlCode != control.ControlCode {
		if _, err := s.controlRepo.FindByCode(control.ControlCode); err == nil {
			return nil, apperrors.Wrap("CONTROL_CODE_ALREADY_EXISTS", "Control code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate control code", 500, err)
		}
	}
	existingControl.ControlCode = control.ControlCode
	existingControl.ControlName = control.ControlName
	existingControl.Description = control.Description
	existingControl.ControlType = control.ControlType
	existingControl.ControlCategory = control.ControlCategory
	existingControl.RiskRegisterID = control.RiskRegisterID
	existingControl.OwnerID = control.OwnerID
	existingControl.DepartmentID = control.DepartmentID
	existingControl.Frequency = control.Frequency
	existingControl.Documentation = control.Documentation
	existingControl.IsKeyControl = control.IsKeyControl
	existingControl.IsActive = control.IsActive
	if err := s.controlRepo.Update(existingControl); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update control", 500, err) }
	return s.controlRepo.FindByID(controlID)
}
