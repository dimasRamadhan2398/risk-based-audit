package services

import (
	"audit-service/models"
	"audit-service/pkg/errors"
	"audit-service/pkg/logger"
	"audit-service/pkg/utils"
	"audit-service/repositories"
	"audit-service/services/base"
	"context"

	"github.com/google/uuid"
)

// AuditMandateServiceInterface defines the audit mandate service interface
type AuditMandateServiceInterface interface {
	CreateMandate(ctx context.Context, req *models.CreateAuditMandateRequest) (*models.AuditMandateResponse, error)
	UpdateMandate(ctx context.Context, id uuid.UUID, req *models.UpdateAuditMandateRequest) (*models.AuditMandateResponse, error)
	DeleteMandate(ctx context.Context, id uuid.UUID) error
	GetMandate(ctx context.Context, id uuid.UUID) (*models.AuditMandateResponse, error)
	GetMandateByReference(ctx context.Context, refNumber string) (*models.AuditMandateResponse, error)
	GetActiveMandate(ctx context.Context) (*models.AuditMandateResponse, error)
	ListMandates(ctx context.Context, req *models.ListAuditMandatesRequest) ([]*models.AuditMandateResponse, *utils.PaginationResponse, error)
}

// AuditMandateService handles audit mandate business logic
type AuditMandateService struct {
	*base.BaseService
	repo repositories.AuditMandateRepositoryInterface
}

// NewAuditMandateService creates a new audit mandate service
func NewAuditMandateService(repo repositories.AuditMandateRepositoryInterface) AuditMandateServiceInterface {
	return &AuditMandateService{
		BaseService: base.NewBaseService(),
		repo:        repo,
	}
}

// CreateMandate creates a new audit mandate
func (s *AuditMandateService) CreateMandate(ctx context.Context, req *models.CreateAuditMandateRequest) (*models.AuditMandateResponse, error) {
	// Check if reference number already exists
	if _, err := s.repo.FindByReferenceNumber(req.ReferenceNumber); err == nil {
		return nil, errors.ErrConflict
	}

	mandate := &models.AuditMandate{
		Title:           req.Title,
		ReferenceNumber: req.ReferenceNumber,
		MandateSource:   req.MandateSource,
		LegalBasis:      req.LegalBasis,
		EffectiveDate:   req.EffectiveDate,
		ExpiryDate:      req.ExpiryDate,
		IsActive:        true,
	}

	if err := s.repo.Create(mandate); err != nil {
		s.LogError("Failed to create audit mandate", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit mandate created", logger.LogField("id", mandate.ID))
	return s.toResponse(mandate), nil
}

// UpdateMandate updates an audit mandate
func (s *AuditMandateService) UpdateMandate(ctx context.Context, id uuid.UUID, req *models.UpdateAuditMandateRequest) (*models.AuditMandateResponse, error) {
	mandate, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		mandate.Title = *req.Title
	}
	if req.MandateSource != nil {
		mandate.MandateSource = *req.MandateSource
	}
	if req.LegalBasis != nil {
		mandate.LegalBasis = *req.LegalBasis
	}
	if req.EffectiveDate != nil {
		mandate.EffectiveDate = *req.EffectiveDate
	}
	if req.ExpiryDate != nil {
		mandate.ExpiryDate = req.ExpiryDate
	}
	if req.IsActive != nil {
		mandate.IsActive = *req.IsActive
	}

	if err := s.repo.Update(mandate); err != nil {
		s.LogError("Failed to update audit mandate", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit mandate updated", logger.LogField("id", mandate.ID))
	return s.toResponse(mandate), nil
}

// DeleteMandate deletes an audit mandate
func (s *AuditMandateService) DeleteMandate(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		s.LogError("Failed to delete audit mandate", logger.LogField("error", err))
		return err
	}

	s.LogInfo("Audit mandate deleted", logger.LogField("id", id))
	return nil
}

// GetMandate retrieves an audit mandate by ID
func (s *AuditMandateService) GetMandate(ctx context.Context, id uuid.UUID) (*models.AuditMandateResponse, error) {
	mandate, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mandate), nil
}

// GetMandateByReference retrieves an audit mandate by reference number
func (s *AuditMandateService) GetMandateByReference(ctx context.Context, refNumber string) (*models.AuditMandateResponse, error) {
	mandate, err := s.repo.FindByReferenceNumber(refNumber)
	if err != nil {
		return nil, err
	}

	return s.toResponse(mandate), nil
}

// GetActiveMandate retrieves the active audit mandate
func (s *AuditMandateService) GetActiveMandate(ctx context.Context) (*models.AuditMandateResponse, error) {
	mandate, err := s.repo.FindActive()
	if err != nil {
		return nil, err
	}

	return s.toResponse(mandate), nil
}

// ListMandates retrieves a list of audit mandates with pagination
func (s *AuditMandateService) ListMandates(ctx context.Context, req *models.ListAuditMandatesRequest) ([]*models.AuditMandateResponse, *utils.PaginationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize, "created_at", "desc")

	mandates, err := s.repo.FindMany(pagination.GetOffset(), pagination.GetLimit(), req.Search, req.IsActive)
	if err != nil {
		s.LogError("Failed to list audit mandates", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	totalCount, err := s.repo.Count(req.Search, req.IsActive)
	if err != nil {
		s.LogError("Failed to count audit mandates", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	responses := make([]*models.AuditMandateResponse, len(mandates))
	for i, mandate := range mandates {
		responses[i] = s.toResponse(mandate)
	}

	paginationResp := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, totalCount)
	return responses, paginationResp, nil
}

// toResponse converts an audit mandate model to a response DTO
func (s *AuditMandateService) toResponse(mandate *models.AuditMandate) *models.AuditMandateResponse {
	resp := &models.AuditMandateResponse{
		ID:              mandate.ID.String(),
		Title:           mandate.Title,
		ReferenceNumber: mandate.ReferenceNumber,
		MandateSource:   mandate.MandateSource,
		LegalBasis:      mandate.LegalBasis,
		EffectiveDate:   mandate.EffectiveDate.Format("2006-01-02"),
		IsActive:        mandate.IsActive,
		CreatedAt:       mandate.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       mandate.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if mandate.ExpiryDate != nil {
		resp.ExpiryDate = mandate.ExpiryDate.Format("2006-01-02")
	}

	return resp
}