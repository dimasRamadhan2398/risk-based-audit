package control_assessment
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/control_assessment"
	"github.com/google/uuid"
)
type ControlAssessmentServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.ControlAssessment, error)
	FindById(ctx *base.BaseService, id string) (*models.ControlAssessment, error)
	Create(ctx *base.BaseService, assessment *models.ControlAssessment) (*models.ControlAssessment, error)
	Update(ctx *base.BaseService, id string, assessment *models.ControlAssessment) (*models.ControlAssessment, error)
	Delete(ctx *base.BaseService, id string) error
}
type ControlAssessmentService struct { assessmentRepo repo.IControlAssessmentRepository }
func NewControlAssessmentService(assessmentRepo repo.IControlAssessmentRepository) ControlAssessmentServiceInterface { return &ControlAssessmentService{assessmentRepo: assessmentRepo} }
func (s *ControlAssessmentService) Create(ctx *base.BaseService, assessment *models.ControlAssessment) (*models.ControlAssessment, error) {
	if _, err := s.assessmentRepo.FindByCode(assessment.AssessmentCode); err == nil {
		return nil, apperrors.Wrap("ASSESSMENT_CODE_ALREADY_EXISTS", "Assessment code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate assessment code", 500, err)
	}
	if err := s.assessmentRepo.Create(assessment); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create control assessment", 500, err) }
	return s.assessmentRepo.FindByID(assessment.ID)
}
func (s *ControlAssessmentService) Delete(ctx *base.BaseService, id string) error {
	assessmentID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_ASSESSMENT_ID", "Invalid assessment ID format", 400, err) }
	if _, err := s.assessmentRepo.FindByID(assessmentID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find assessment", 500, err)
	}
	if err := s.assessmentRepo.Delete(assessmentID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete assessment", 500, err) }
	return nil
}
func (s *ControlAssessmentService) FindAll(ctx *base.BaseService) (*[]models.ControlAssessment, error) {
	assessments, err := s.assessmentRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch control assessments", 500, err) }
	result := make([]models.ControlAssessment, 0, len(assessments))
	for _, assessment := range assessments { if assessment != nil { result = append(result, *assessment) } }
	return &result, nil
}
func (s *ControlAssessmentService) FindById(ctx *base.BaseService, id string) (*models.ControlAssessment, error) {
	assessmentID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ASSESSMENT_ID", "Invalid assessment ID format", 400, err) }
	assessment, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch control assessment", 500, err)
	}
	return assessment, nil
}
func (s *ControlAssessmentService) Update(ctx *base.BaseService, id string, assessment *models.ControlAssessment) (*models.ControlAssessment, error) {
	assessmentID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_ASSESSMENT_ID", "Invalid assessment ID format", 400, err) }
	existingAssessment, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch control assessment", 500, err)
	}
	if existingAssessment.AssessmentCode != assessment.AssessmentCode {
		if _, err := s.assessmentRepo.FindByCode(assessment.AssessmentCode); err == nil {
			return nil, apperrors.Wrap("ASSESSMENT_CODE_ALREADY_EXISTS", "Assessment code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate assessment code", 500, err)
		}
	}
	existingAssessment.AssessmentCode = assessment.AssessmentCode
	existingAssessment.ControlID = assessment.ControlID
	existingAssessment.AuditPlanID = assessment.AuditPlanID
	existingAssessment.AuditScopeID = assessment.AuditScopeID
	existingAssessment.AssessmentStatus = assessment.AssessmentStatus
	existingAssessment.TestingStatus = assessment.TestingStatus
	existingAssessment.Effectiveness = assessment.Effectiveness
	existingAssessment.TestMethod = assessment.TestMethod
	existingAssessment.TestPeriodStart = assessment.TestPeriodStart
	existingAssessment.TestPeriodEnd = assessment.TestPeriodEnd
	existingAssessment.SampleSize = assessment.SampleSize
	existingAssessment.ExceptionsFound = assessment.ExceptionsFound
	existingAssessment.Finding = assessment.Finding
	existingAssessment.RootCause = assessment.RootCause
	existingAssessment.Recommendation = assessment.Recommendation
	existingAssessment.AuditorID = assessment.AuditorID
	existingAssessment.AssessmentDate = assessment.AssessmentDate
	existingAssessment.CompletedAt = assessment.CompletedAt
	existingAssessment.IsActive = assessment.IsActive
	if err := s.assessmentRepo.Update(existingAssessment); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update control assessment", 500, err) }
	return s.assessmentRepo.FindByID(assessmentID)
}
