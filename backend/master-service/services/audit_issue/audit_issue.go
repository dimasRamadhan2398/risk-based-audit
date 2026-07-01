package audit_issue
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_issue"
	"github.com/google/uuid"
)
type AuditIssueServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditIssue, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditIssue, error)
	Create(ctx *base.BaseService, issue *models.AuditIssue) (*models.AuditIssue, error)
	Update(ctx *base.BaseService, id string, issue *models.AuditIssue) (*models.AuditIssue, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditIssueService struct { issueRepo repo.IAuditIssueRepository }
func NewAuditIssueService(issueRepo repo.IAuditIssueRepository) AuditIssueServiceInterface { return &AuditIssueService{issueRepo: issueRepo} }
func (s *AuditIssueService) Create(ctx *base.BaseService, issue *models.AuditIssue) (*models.AuditIssue, error) {
	if _, err := s.issueRepo.FindByCode(issue.IssueCode); err == nil {
		return nil, apperrors.Wrap("ISSUE_CODE_ALREADY_EXISTS", "Issue code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate issue code", 500, err)
	}
	if err := s.issueRepo.Create(issue); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit issue", 500, err) }
	return s.issueRepo.FindByID(issue.ID)
}
func (s *AuditIssueService) Delete(ctx *base.BaseService, id string) error {
	issueID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_ISSUE_ID", "Invalid issue ID format", 400, err) }
	if _, err := s.issueRepo.FindByID(issueID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find issue", 500, err)
	}
	if err := s.issueRepo.Delete(issueID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete issue", 500, err) }
	return nil
}
func (s *AuditIssueService) FindAll(ctx *base.BaseService) (*[]models.AuditIssue, error) {
	issues, err := s.issueRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit issues", 500, err) }
	result := make([]models.AuditIssue, 0, len(issues))
	for _, issue := range issues { if issue != nil { result = append(result, *issue) } }
	return &result, nil
}
func (s *AuditIssueService) FindById(ctx *base.BaseService, id string) (*models.AuditIssue, error) {
	issueID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ISSUE_ID", "Invalid issue ID format", 400, err) }
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit issue", 500, err)
	}
	return issue, nil
}
func (s *AuditIssueService) Update(ctx *base.BaseService, id string, issue *models.AuditIssue) (*models.AuditIssue, error) {
	issueID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ISSUE_ID", "Invalid issue ID format", 400, err) }
	existingIssue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit issue", 500, err)
	}
	if existingIssue.IssueCode != issue.IssueCode {
		if _, err := s.issueRepo.FindByCode(issue.IssueCode); err == nil {
			return nil, apperrors.Wrap("ISSUE_CODE_ALREADY_EXISTS", "Issue code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate issue code", 500, err)
		}
	}
	existingIssue.IssueCode = issue.IssueCode
	existingIssue.FindingID = issue.FindingID
	existingIssue.AuditPlanID = issue.AuditPlanID
	existingIssue.Category = issue.Category
	existingIssue.Priority = issue.Priority
	existingIssue.Title = issue.Title
	existingIssue.Description = issue.Description
	existingIssue.CurrentState = issue.CurrentState
	existingIssue.ExpectedState = issue.ExpectedState
	existingIssue.Impact = issue.Impact
	existingIssue.RiskRegisterID = issue.RiskRegisterID
	existingIssue.ControlID = issue.ControlID
	existingIssue.DepartmentID = issue.DepartmentID
	existingIssue.OwnerID = issue.OwnerID
	existingIssue.AuditorID = issue.AuditorID
	existingIssue.Status = issue.Status
	existingIssue.RootCause = issue.RootCause
	existingIssue.IssueDate = issue.IssueDate
	existingIssue.DueDate = issue.DueDate
	existingIssue.RemediatedDate = issue.RemediatedDate
	existingIssue.ValidatedDate = issue.ValidatedDate
	existingIssue.ClosedAt = issue.ClosedAt
	existingIssue.RemediationPlan = issue.RemediationPlan
	existingIssue.RemediationEvidence = issue.RemediationEvidence
	if err := s.issueRepo.Update(existingIssue); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit issue", 500, err) }
	return s.issueRepo.FindByID(issueID)
}
