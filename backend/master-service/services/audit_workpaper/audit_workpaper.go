package audit_workpaper
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_workpaper"
	"github.com/google/uuid"
)
type AuditWorkpaperServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditWorkpaper, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditWorkpaper, error)
	Create(ctx *base.BaseService, wp *models.AuditWorkpaper) (*models.AuditWorkpaper, error)
	Update(ctx *base.BaseService, id string, wp *models.AuditWorkpaper) (*models.AuditWorkpaper, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditWorkpaperService struct { wpRepo repo.IAuditWorkpaperRepository }
func NewAuditWorkpaperService(wpRepo repo.IAuditWorkpaperRepository) AuditWorkpaperServiceInterface { return &AuditWorkpaperService{wpRepo: wpRepo} }
func (s *AuditWorkpaperService) Create(ctx *base.BaseService, wp *models.AuditWorkpaper) (*models.AuditWorkpaper, error) {
	if _, err := s.wpRepo.FindByCode(wp.WorkpaperCode); err == nil {
		return nil, apperrors.Wrap("WORKPAPER_CODE_ALREADY_EXISTS", "Workpaper code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate workpaper code", 500, err)
	}
	if err := s.wpRepo.Create(wp); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit workpaper", 500, err) }
	return s.wpRepo.FindByID(wp.ID)
}
func (s *AuditWorkpaperService) Delete(ctx *base.BaseService, id string) error {
	wpID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_WORKPAPER_ID", "Invalid workpaper ID format", 400, err) }
	if _, err := s.wpRepo.FindByID(wpID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find workpaper", 500, err)
	}
	if err := s.wpRepo.Delete(wpID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete workpaper", 500, err) }
	return nil
}
func (s *AuditWorkpaperService) FindAll(ctx *base.BaseService) (*[]models.AuditWorkpaper, error) {
	wps, err := s.wpRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit workpapers", 500, err) }
	result := make([]models.AuditWorkpaper, 0, len(wps))
	for _, wp := range wps { if wp != nil { result = append(result, *wp) } }
	return &result, nil
}
func (s *AuditWorkpaperService) FindById(ctx *base.BaseService, id string) (*models.AuditWorkpaper, error) {
	wpID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_WORKPAPER_ID", "Invalid workpaper ID format", 400, err) }
	wp, err := s.wpRepo.FindByID(wpID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit workpaper", 500, err)
	}
	return wp, nil
}
func (s *AuditWorkpaperService) Update(ctx *base.BaseService, id string, wp *models.AuditWorkpaper) (*models.AuditWorkpaper, error) {
	wpID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_WORKPAPER_ID", "Invalid workpaper ID format", 400, err) }
	existingWp, err := s.wpRepo.FindByID(wpID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit workpaper", 500, err)
	}
	if existingWp.WorkpaperCode != wp.WorkpaperCode {
		if _, err := s.wpRepo.FindByCode(wp.WorkpaperCode); err == nil {
			return nil, apperrors.Wrap("WORKPAPER_CODE_ALREADY_EXISTS", "Workpaper code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate workpaper code", 500, err)
		}
	}
	existingWp.WorkpaperCode = wp.WorkpaperCode
	existingWp.Title = wp.Title
	existingWp.WorkpaperType = wp.WorkpaperType
	existingWp.Description = wp.Description
	existingWp.AuditPlanID = wp.AuditPlanID
	existingWp.AuditScopeID = wp.AuditScopeID
	existingWp.ControlAssessmentID = wp.ControlAssessmentID
	existingWp.AuditFindingID = wp.AuditFindingID
	existingWp.Content = wp.Content
	existingWp.AttachmentURL = wp.AttachmentURL
	existingWp.Status = wp.Status
	existingWp.Version = wp.Version
	existingWp.AuditorID = wp.AuditorID
	existingWp.ReviewerID = wp.ReviewerID
	existingWp.ReviewedAt = wp.ReviewedAt
	existingWp.ReviewNotes = wp.ReviewNotes
	if err := s.wpRepo.Update(existingWp); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit workpaper", 500, err) }
	return s.wpRepo.FindByID(wpID)
}
