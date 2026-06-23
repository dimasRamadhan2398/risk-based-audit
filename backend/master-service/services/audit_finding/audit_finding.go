package audit_finding
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_finding"
	"github.com/google/uuid"
)
type AuditFindingServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditFinding, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditFinding, error)
	Create(ctx *base.BaseService, finding *models.AuditFinding) (*models.AuditFinding, error)
	Update(ctx *base.BaseService, id string, finding *models.AuditFinding) (*models.AuditFinding, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditFindingService struct { findingRepo repo.IAuditFindingRepository }
func NewAuditFindingService(findingRepo repo.IAuditFindingRepository) AuditFindingServiceInterface { return &AuditFindingService{findingRepo: findingRepo} }
func (s *AuditFindingService) Create(ctx *base.BaseService, finding *models.AuditFinding) (*models.AuditFinding, error) {
	if _, err := s.findingRepo.FindByCode(finding.FindingCode); err == nil {
		return nil, apperrors.Wrap("FINDING_CODE_ALREADY_EXISTS", "Finding code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate finding code", 500, err)
	}
	if err := s.findingRepo.Create(finding); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit finding", 500, err) }
	return s.findingRepo.FindByID(finding.ID)
}
func (s *AuditFindingService) Delete(ctx *base.BaseService, id string) error {
	findingID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_FINDING_ID", "Invalid finding ID format", 400, err) }
	if _, err := s.findingRepo.FindByID(findingID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find finding", 500, err)
	}
	if err := s.findingRepo.Delete(findingID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete finding", 500, err) }
	return nil
}
func (s *AuditFindingService) FindAll(ctx *base.BaseService) (*[]models.AuditFinding, error) {
	findings, err := s.findingRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit findings", 500, err) }
	result := make([]models.AuditFinding, 0, len(findings))
	for _, finding := range findings { if finding != nil { result = append(result, *finding) } }
	return &result, nil
}
func (s *AuditFindingService) FindById(ctx *base.BaseService, id string) (*models.AuditFinding, error) {
	findingID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_FINDING_ID", "Invalid finding ID format", 400, err) }
	finding, err := s.findingRepo.FindByID(findingID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit finding", 500, err)
	}
	return finding, nil
}
func (s *AuditFindingService) Update(ctx *base.BaseService, id string, finding *models.AuditFinding) (*models.AuditFinding, error) {
	findingID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_FINDING_ID", "Invalid finding ID format", 400, err) }
	existingFinding, err := s.findingRepo.FindByID(findingID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit finding", 500, err)
	}
	if existingFinding.FindingCode != finding.FindingCode {
		if _, err := s.findingRepo.FindByCode(finding.FindingCode); err == nil {
			return nil, apperrors.Wrap("FINDING_CODE_ALREADY_EXISTS", "Finding code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate finding code", 500, err)
		}
	}
	existingFinding.FindingCode = finding.FindingCode
	existingFinding.AuditPlanID = finding.AuditPlanID
	existingFinding.AuditScopeID = finding.AuditScopeID
	existingFinding.ControlAssessmentID = finding.ControlAssessmentID
	existingFinding.FindingType = finding.FindingType
	existingFinding.Severity = finding.Severity
	existingFinding.Title = finding.Title
	existingFinding.Description = finding.Description
	existingFinding.Facts = finding.Facts
	existingFinding.Criterion = finding.Criterion
	existingFinding.Condition = finding.Condition
	existingFinding.Cause = finding.Cause
	existingFinding.Effect = finding.Effect
	existingFinding.RiskRegisterID = finding.RiskRegisterID
	existingFinding.DepartmentID = finding.DepartmentID
	existingFinding.OwnerID = finding.OwnerID
	existingFinding.AuditorID = finding.AuditorID
	existingFinding.Status = finding.Status
	existingFinding.AgreedWithManagement = finding.AgreedWithManagement
	existingFinding.IssueDate = finding.IssueDate
	existingFinding.DueDate = finding.DueDate
	existingFinding.ClosedAt = finding.ClosedAt
	existingFinding.Evidence = finding.Evidence
	if err := s.findingRepo.Update(existingFinding); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit finding", 500, err) }
	return s.findingRepo.FindByID(findingID)
}
