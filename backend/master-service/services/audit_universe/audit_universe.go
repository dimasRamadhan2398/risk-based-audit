package audit_universe
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_universe"
	"github.com/google/uuid"
)
type AuditUniverseServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditUniverse, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditUniverse, error)
	Create(ctx *base.BaseService, u *models.AuditUniverse) (*models.AuditUniverse, error)
	Update(ctx *base.BaseService, id string, u *models.AuditUniverse) (*models.AuditUniverse, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditUniverseService struct { uRepo repo.IAuditUniverseRepository }
func NewAuditUniverseService(uRepo repo.IAuditUniverseRepository) AuditUniverseServiceInterface { return &AuditUniverseService{uRepo: uRepo} }
func (s *AuditUniverseService) Create(ctx *base.BaseService, u *models.AuditUniverse) (*models.AuditUniverse, error) {
	if _, err := s.uRepo.FindByCode(u.EntityCode); err == nil {
		return nil, apperrors.Wrap("AUDIT_UNIVERSE_CODE_ALREADY_EXISTS", "Audit universe code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit universe code", 500, err)
	}
	if err := s.uRepo.Create(u); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit universe", 500, err) }
	return s.uRepo.FindByID(u.ID)
}
func (s *AuditUniverseService) Delete(ctx *base.BaseService, id string) error {
	uID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_AUDIT_UNIVERSE_ID", "Invalid audit universe ID format", 400, err) }
	if _, err := s.uRepo.FindByID(uID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find audit universe", 500, err)
	}
	if err := s.uRepo.Delete(uID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete audit universe", 500, err) }
	return nil
}
func (s *AuditUniverseService) FindAll(ctx *base.BaseService) (*[]models.AuditUniverse, error) {
	us, err := s.uRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit universe entities", 500, err) }
	result := make([]models.AuditUniverse, 0, len(us))
	for _, u := range us { if u != nil { result = append(result, *u) } }
	return &result, nil
}
func (s *AuditUniverseService) FindById(ctx *base.BaseService, id string) (*models.AuditUniverse, error) {
	uID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_UNIVERSE_ID", "Invalid audit universe ID format", 400, err) }
	u, err := s.uRepo.FindByID(uID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit universe entity", 500, err)
	}
	return u, nil
}
func (s *AuditUniverseService) Update(ctx *base.BaseService, id string, u *models.AuditUniverse) (*models.AuditUniverse, error) {
	uID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_UNIVERSE_ID", "Invalid audit universe ID format", 400, err) }
	existingU, err := s.uRepo.FindByID(uID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit universe entity", 500, err)
	}
	if existingU.EntityCode != u.EntityCode {
		if _, err := s.uRepo.FindByCode(u.EntityCode); err == nil {
			return nil, apperrors.Wrap("AUDIT_UNIVERSE_CODE_ALREADY_EXISTS", "Audit universe code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit universe code", 500, err)
		}
	}
	existingU.EntityCode = u.EntityCode
	existingU.EntityName = u.EntityName
	existingU.EntityType = u.EntityType
	existingU.Description = u.Description
	existingU.BusinessOwner = u.BusinessOwner
	existingU.ParentID = u.ParentID
	existingU.RiskRating = u.RiskRating
	existingU.LastAuditYear = u.LastAuditYear
	existingU.LastAuditDate = u.LastAuditDate
	existingU.LastAuditResult = u.LastAuditResult
	existingU.AuditFrequency = u.AuditFrequency
	existingU.IsMandatory = u.IsMandatory
	existingU.IsHighPriority = u.IsHighPriority
	existingU.DepartmentID = u.DepartmentID
	existingU.Status = u.Status
	if err := s.uRepo.Update(existingU); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit universe entity", 500, err) }
	return s.uRepo.FindByID(uID)
}
