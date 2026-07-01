package audit_period
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_period"
	"github.com/google/uuid"
)
type AuditPeriodServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditPeriod, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditPeriod, error)
	Create(ctx *base.BaseService, p *models.AuditPeriod) (*models.AuditPeriod, error)
	Update(ctx *base.BaseService, id string, p *models.AuditPeriod) (*models.AuditPeriod, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditPeriodService struct { pRepo repo.IAuditPeriodRepository }
func NewAuditPeriodService(pRepo repo.IAuditPeriodRepository) AuditPeriodServiceInterface { return &AuditPeriodService{pRepo: pRepo} }
func (s *AuditPeriodService) Create(ctx *base.BaseService, p *models.AuditPeriod) (*models.AuditPeriod, error) {
	if _, err := s.pRepo.FindByCode(p.PeriodCode); err == nil {
		return nil, apperrors.Wrap("AUDIT_PERIOD_CODE_ALREADY_EXISTS", "Audit period code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit period code", 500, err)
	}
	if err := s.pRepo.Create(p); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit period", 500, err) }
	return p, nil
}
func (s *AuditPeriodService) Delete(ctx *base.BaseService, id string) error {
	pID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_AUDIT_PERIOD_ID", "Invalid audit period ID format", 400, err) }
	if _, err := s.pRepo.FindByID(pID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find audit period", 500, err)
	}
	if err := s.pRepo.Delete(pID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete audit period", 500, err) }
	return nil
}
func (s *AuditPeriodService) FindAll(ctx *base.BaseService) (*[]models.AuditPeriod, error) {
	ps, err := s.pRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit periods", 500, err) }
	result := make([]models.AuditPeriod, 0, len(ps))
	for _, p := range ps { if p != nil { result = append(result, *p) } }
	return &result, nil
}
func (s *AuditPeriodService) FindById(ctx *base.BaseService, id string) (*models.AuditPeriod, error) {
	pID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_PERIOD_ID", "Invalid audit period ID format", 400, err) }
	p, err := s.pRepo.FindByID(pID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit period", 500, err)
	}
	return p, nil
}
func (s *AuditPeriodService) Update(ctx *base.BaseService, id string, p *models.AuditPeriod) (*models.AuditPeriod, error) {
	pID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_PERIOD_ID", "Invalid audit period ID format", 400, err) }
	existingP, err := s.pRepo.FindByID(pID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit period", 500, err)
	}
	if existingP.PeriodCode != p.PeriodCode {
		if _, err := s.pRepo.FindByCode(p.PeriodCode); err == nil {
			return nil, apperrors.Wrap("AUDIT_PERIOD_CODE_ALREADY_EXISTS", "Audit period code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit period code", 500, err)
		}
	}
	existingP.PeriodCode = p.PeriodCode
	existingP.PeriodName = p.PeriodName
	existingP.Year = p.Year
	existingP.Quarter = p.Quarter
	existingP.StartDate = p.StartDate
	existingP.EndDate = p.EndDate
	existingP.Status = p.Status
	existingP.CompanyID = p.CompanyID
	existingP.IsActive = p.IsActive
	if err := s.pRepo.Update(existingP); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit period", 500, err) }
	return existingP, nil
}
