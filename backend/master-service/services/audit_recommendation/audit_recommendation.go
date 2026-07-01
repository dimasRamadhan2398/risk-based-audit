package audit_recommendation
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_recommendation"
	"github.com/google/uuid"
)
type AuditRecommendationServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditRecommendation, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditRecommendation, error)
	Create(ctx *base.BaseService, rec *models.AuditRecommendation) (*models.AuditRecommendation, error)
	Update(ctx *base.BaseService, id string, rec *models.AuditRecommendation) (*models.AuditRecommendation, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditRecommendationService struct { recRepo repo.IAuditRecommendationRepository }
func NewAuditRecommendationService(recRepo repo.IAuditRecommendationRepository) AuditRecommendationServiceInterface { return &AuditRecommendationService{recRepo: recRepo} }
func (s *AuditRecommendationService) Create(ctx *base.BaseService, rec *models.AuditRecommendation) (*models.AuditRecommendation, error) {
	if _, err := s.recRepo.FindByCode(rec.RecommendationCode); err == nil {
		return nil, apperrors.Wrap("RECOMMENDATION_CODE_ALREADY_EXISTS", "Recommendation code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate recommendation code", 500, err)
	}
	if err := s.recRepo.Create(rec); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit recommendation", 500, err) }
	return s.recRepo.FindByID(rec.ID)
}
func (s *AuditRecommendationService) Delete(ctx *base.BaseService, id string) error {
	recID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_RECOMMENDATION_ID", "Invalid recommendation ID format", 400, err) }
	if _, err := s.recRepo.FindByID(recID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find recommendation", 500, err)
	}
	if err := s.recRepo.Delete(recID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete recommendation", 500, err) }
	return nil
}
func (s *AuditRecommendationService) FindAll(ctx *base.BaseService) (*[]models.AuditRecommendation, error) {
	recs, err := s.recRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit recommendations", 500, err) }
	result := make([]models.AuditRecommendation, 0, len(recs))
	for _, rec := range recs { if rec != nil { result = append(result, *rec) } }
	return &result, nil
}
func (s *AuditRecommendationService) FindById(ctx *base.BaseService, id string) (*models.AuditRecommendation, error) {
	recID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RECOMMENDATION_ID", "Invalid recommendation ID format", 400, err) }
	rec, err := s.recRepo.FindByID(recID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit recommendation", 500, err)
	}
	return rec, nil
}
func (s *AuditRecommendationService) Update(ctx *base.BaseService, id string, rec *models.AuditRecommendation) (*models.AuditRecommendation, error) {
	recID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RECOMMENDATION_ID", "Invalid recommendation ID format", 400, err) }
	existingRec, err := s.recRepo.FindByID(recID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit recommendation", 500, err)
	}
	if existingRec.RecommendationCode != rec.RecommendationCode {
		if _, err := s.recRepo.FindByCode(rec.RecommendationCode); err == nil {
			return nil, apperrors.Wrap("RECOMMENDATION_CODE_ALREADY_EXISTS", "Recommendation code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate recommendation code", 500, err)
		}
	}
	existingRec.RecommendationCode = rec.RecommendationCode
	existingRec.FindingID = rec.FindingID
	existingRec.Recommendation = rec.Recommendation
	existingRec.Priority = rec.Priority
	existingRec.Benefits = rec.Benefits
	existingRec.ManagementResponse = rec.ManagementResponse
	existingRec.ManagementComment = rec.ManagementComment
	existingRec.ActionPlan = rec.ActionPlan
	existingRec.ResponsibleID = rec.ResponsibleID
	existingRec.DueDate = rec.DueDate
	existingRec.CompletedDate = rec.CompletedDate
	existingRec.ClosedAt = rec.ClosedAt
	existingRec.Status = rec.Status
	existingRec.Progress = rec.Progress
	existingRec.VerifiedByID = rec.VerifiedByID
	existingRec.VerifiedAt = rec.VerifiedAt
	existingRec.VerificationNote = rec.VerificationNote
	existingRec.CostEstimate = rec.CostEstimate
	existingRec.ActualCost = rec.ActualCost
	if err := s.recRepo.Update(existingRec); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit recommendation", 500, err) }
	return s.recRepo.FindByID(recID)
}
