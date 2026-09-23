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

// AuditCharterServiceInterface defines the audit charter service interface
type AuditCharterServiceInterface interface {
	CreateCharter(ctx context.Context, req *models.CreateAuditCharterRequest) (*models.AuditCharterResponse, error)
	UpdateCharter(ctx context.Context, id uuid.UUID, req *models.UpdateAuditCharterRequest) (*models.AuditCharterResponse, error)
	DeleteCharter(ctx context.Context, id uuid.UUID) error
	GetCharter(ctx context.Context, id uuid.UUID) (*models.AuditCharterResponse, error)
	GetCharterByVersion(ctx context.Context, version string) (*models.AuditCharterResponse, error)
	GetActiveCharter(ctx context.Context) (*models.AuditCharterResponse, error)
	ListCharters(ctx context.Context, req *models.ListAuditChartersRequest) ([]*models.AuditCharterResponse, *utils.PaginationResponse, error)
	SetActiveCharter(ctx context.Context, id uuid.UUID) error
}

// AuditCharterService handles audit charter business logic
type AuditCharterService struct {
	*base.BaseService
	repo repositories.AuditCharterRepositoryInterface
}

// NewAuditCharterService creates a new audit charter service
func NewAuditCharterService(repo repositories.AuditCharterRepositoryInterface) AuditCharterServiceInterface {
	return &AuditCharterService{
		BaseService: base.NewBaseService(),
		repo:        repo,
	}
}

// CreateCharter creates a new audit charter
func (s *AuditCharterService) CreateCharter(ctx context.Context, req *models.CreateAuditCharterRequest) (*models.AuditCharterResponse, error) {
	// If version already exists, append suffix to avoid hard failure
	if _, err := s.repo.FindByVersion(req.Version); err == nil && req.Version != "" {
		req.Version = req.Version + "-v2"
	}

	charter := &models.AuditCharter{
		Filename: req.Filename,
		Version:  req.Version,
		Title:    req.Title,
		Content:  req.Content,
		IsActive: req.IsActive != nil && *req.IsActive,
		FileUrl:  req.FileUrl,
		FileSize: req.FileSize,
	}

	if err := s.repo.Create(charter); err != nil {
		s.LogError("Failed to create audit charter", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	if charter.IsActive {
		// Deactivate others
		allCharters, err := s.repo.FindMany(0, 1000, "", nil)
		if err == nil {
			for _, c := range allCharters {
				if c.ID != charter.ID && c.IsActive {
					c.IsActive = false
					_ = s.repo.Update(c)
				}
			}
		}
	}

	s.LogInfo("Audit charter created", logger.LogField("id", charter.ID))
	return s.toResponse(charter), nil
}

// UpdateCharter updates an audit charter
func (s *AuditCharterService) UpdateCharter(ctx context.Context, id uuid.UUID, req *models.UpdateAuditCharterRequest) (*models.AuditCharterResponse, error) {
	charter, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Filename != nil {
		charter.Filename = *req.Filename
	}
	if req.Title != nil {
		charter.Title = *req.Title
	}
	if req.Content != nil {
		charter.Content = *req.Content
	}
	if req.IsActive != nil {
		charter.IsActive = *req.IsActive
	}
	if req.FileUrl != nil {
		charter.FileUrl = *req.FileUrl
	}
	if req.FileSize != nil {
		charter.FileSize = *req.FileSize
	}

	if err := s.repo.Update(charter); err != nil {
		s.LogError("Failed to update audit charter", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	if charter.IsActive {
		// Deactivate others
		allCharters, err := s.repo.FindMany(0, 1000, "", nil)
		if err == nil {
			for _, c := range allCharters {
				if c.ID != id && c.IsActive {
					c.IsActive = false
					_ = s.repo.Update(c)
				}
			}
		}
	}

	s.LogInfo("Audit charter updated", logger.LogField("id", charter.ID))
	return s.toResponse(charter), nil
}

// DeleteCharter deletes an audit charter
func (s *AuditCharterService) DeleteCharter(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		s.LogError("Failed to delete audit charter", logger.LogField("error", err))
		return err
	}

	s.LogInfo("Audit charter deleted", logger.LogField("id", id))
	return nil
}

// GetCharter retrieves an audit charter by ID
func (s *AuditCharterService) GetCharter(ctx context.Context, id uuid.UUID) (*models.AuditCharterResponse, error) {
	charter, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(charter), nil
}

// GetCharterByVersion retrieves an audit charter by version
func (s *AuditCharterService) GetCharterByVersion(ctx context.Context, version string) (*models.AuditCharterResponse, error) {
	charter, err := s.repo.FindByVersion(version)
	if err != nil {
		return nil, err
	}

	return s.toResponse(charter), nil
}

// GetActiveCharter retrieves the active audit charter
func (s *AuditCharterService) GetActiveCharter(ctx context.Context) (*models.AuditCharterResponse, error) {
	charter, err := s.repo.FindActive()
	if err != nil {
		return nil, err
	}

	return s.toResponse(charter), nil
}

// ListCharters retrieves a list of audit charters with pagination
func (s *AuditCharterService) ListCharters(ctx context.Context, req *models.ListAuditChartersRequest) ([]*models.AuditCharterResponse, *utils.PaginationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize, "created_at", "desc")

	charters, err := s.repo.FindMany(pagination.GetOffset(), pagination.GetLimit(), req.Search, req.IsActive)
	if err != nil {
		s.LogError("Failed to list audit charters", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	totalCount, err := s.repo.Count(req.Search, req.IsActive)
	if err != nil {
		s.LogError("Failed to count audit charters", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	responses := make([]*models.AuditCharterResponse, len(charters))
	for i, charter := range charters {
		responses[i] = s.toResponse(charter)
	}

	paginationResp := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, totalCount)
	return responses, paginationResp, nil
}

// SetActiveCharter sets an audit charter as active
func (s *AuditCharterService) SetActiveCharter(ctx context.Context, id uuid.UUID) error {
	charter, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Deactivate all other charters
	allCharters, err := s.repo.FindMany(0, 1000, "", nil)
	if err != nil {
		return errors.ErrInternalServer
	}

	for _, c := range allCharters {
		if c.ID != id && c.IsActive {
			c.IsActive = false
			if err := s.repo.Update(c); err != nil {
				return err
			}
		}
	}

	// Activate the selected charter
	charter.IsActive = true
	if err := s.repo.Update(charter); err != nil {
		return errors.ErrInternalServer
	}

	s.LogInfo("Audit charter set as active", logger.LogField("id", id))
	return nil
}

// toResponse converts an audit charter model to a response DTO
func (s *AuditCharterService) toResponse(charter *models.AuditCharter) *models.AuditCharterResponse {
	return &models.AuditCharterResponse{
		ID:        charter.ID.String(),
		Filename:  charter.Filename,
		Version:   charter.Version,
		Title:     charter.Title,
		Content:   charter.Content,
		IsActive:  charter.IsActive,
		FileUrl:   charter.FileUrl,
		FileSize:  charter.FileSize,
		CreatedAt: charter.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: charter.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}