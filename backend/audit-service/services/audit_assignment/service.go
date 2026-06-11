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

// AuditAssignmentServiceInterface defines the audit assignment service interface
type AuditAssignmentServiceInterface interface {
	CreateAssignment(ctx context.Context, req *models.CreateAuditAssignmentRequest) (*models.AuditAssignmentResponse, error)
	UpdateAssignment(ctx context.Context, id uuid.UUID, req *models.UpdateAuditAssignmentRequest) (*models.AuditAssignmentResponse, error)
	DeleteAssignment(ctx context.Context, id uuid.UUID) error
	GetAssignment(ctx context.Context, id uuid.UUID) (*models.AuditAssignmentResponse, error)
	GetAssignmentsByAuditor(ctx context.Context, auditorID uuid.UUID) ([]*models.AuditAssignmentResponse, error)
	GetAssignmentsByAuditPlan(ctx context.Context, auditPlanID uuid.UUID) ([]*models.AuditAssignmentResponse, error)
	ListAssignments(ctx context.Context, req *models.ListAuditAssignmentsRequest) ([]*models.AuditAssignmentResponse, *utils.PaginationResponse, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.AssignmentStatus) (*models.AuditAssignmentResponse, error)
}

// AuditAssignmentService handles audit assignment business logic
type AuditAssignmentService struct {
	*base.BaseService
	repo repositories.AuditAssignmentRepositoryInterface
}

// NewAuditAssignmentService creates a new audit assignment service
func NewAuditAssignmentService(repo repositories.AuditAssignmentRepositoryInterface) AuditAssignmentServiceInterface {
	return &AuditAssignmentService{
		BaseService: base.NewBaseService(),
		repo:        repo,
	}
}

// CreateAssignment creates a new audit assignment
func (s *AuditAssignmentService) CreateAssignment(ctx context.Context, req *models.CreateAuditAssignmentRequest) (*models.AuditAssignmentResponse, error) {
	assignment := &models.AuditAssignment{
		AssignmentTitle: req.AssignmentTitle,
		Description:     req.Description,
		AuditorID:       req.AuditorID,
		AuditPlanID:     req.AuditPlanID,
		Status:          models.AssignmentStatusPending,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Notes:           req.Notes,
	}

	if err := s.repo.Create(assignment); err != nil {
		s.LogError("Failed to create audit assignment", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit assignment created", logger.LogField("id", assignment.ID))
	return s.toResponse(assignment), nil
}

// UpdateAssignment updates an audit assignment
func (s *AuditAssignmentService) UpdateAssignment(ctx context.Context, id uuid.UUID, req *models.UpdateAuditAssignmentRequest) (*models.AuditAssignmentResponse, error) {
	assignment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.AssignmentTitle != nil {
		assignment.AssignmentTitle = *req.AssignmentTitle
	}
	if req.Description != nil {
		assignment.Description = *req.Description
	}
	if req.StartDate != nil {
		assignment.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		assignment.EndDate = req.EndDate
	}
	if req.Notes != nil {
		assignment.Notes = *req.Notes
	}

	if err := s.repo.Update(assignment); err != nil {
		s.LogError("Failed to update audit assignment", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit assignment updated", logger.LogField("id", assignment.ID))
	return s.toResponse(assignment), nil
}

// DeleteAssignment deletes an audit assignment
func (s *AuditAssignmentService) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		s.LogError("Failed to delete audit assignment", logger.LogField("error", err))
		return err
	}

	s.LogInfo("Audit assignment deleted", logger.LogField("id", id))
	return nil
}

// GetAssignment retrieves an audit assignment by ID
func (s *AuditAssignmentService) GetAssignment(ctx context.Context, id uuid.UUID) (*models.AuditAssignmentResponse, error) {
	assignment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(assignment), nil
}

// GetAssignmentsByAuditor retrieves audit assignments by auditor ID
func (s *AuditAssignmentService) GetAssignmentsByAuditor(ctx context.Context, auditorID uuid.UUID) ([]*models.AuditAssignmentResponse, error) {
	assignments, err := s.repo.FindByAuditorID(auditorID)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.AuditAssignmentResponse, len(assignments))
	for i, assignment := range assignments {
		responses[i] = s.toResponse(assignment)
	}

	return responses, nil
}

// GetAssignmentsByAuditPlan retrieves audit assignments by audit plan ID
func (s *AuditAssignmentService) GetAssignmentsByAuditPlan(ctx context.Context, auditPlanID uuid.UUID) ([]*models.AuditAssignmentResponse, error) {
	assignments, err := s.repo.FindByAuditPlanID(auditPlanID)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.AuditAssignmentResponse, len(assignments))
	for i, assignment := range assignments {
		responses[i] = s.toResponse(assignment)
	}

	return responses, nil
}

// ListAssignments retrieves a list of audit assignments with pagination
func (s *AuditAssignmentService) ListAssignments(ctx context.Context, req *models.ListAuditAssignmentsRequest) ([]*models.AuditAssignmentResponse, *utils.PaginationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize, "created_at", "desc")

	assignments, err := s.repo.FindMany(pagination.GetOffset(), pagination.GetLimit(), req.Search, req.AuditorID, req.AuditPlanID, req.Status)
	if err != nil {
		s.LogError("Failed to list audit assignments", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	totalCount, err := s.repo.Count(req.Search, req.AuditorID, req.AuditPlanID, req.Status)
	if err != nil {
		s.LogError("Failed to count audit assignments", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	responses := make([]*models.AuditAssignmentResponse, len(assignments))
	for i, assignment := range assignments {
		responses[i] = s.toResponse(assignment)
	}

	paginationResp := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, totalCount)
	return responses, paginationResp, nil
}

// UpdateStatus updates the status of an audit assignment
func (s *AuditAssignmentService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AssignmentStatus) (*models.AuditAssignmentResponse, error) {
	assignment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	assignment.Status = status
	if err := s.repo.Update(assignment); err != nil {
		s.LogError("Failed to update audit assignment status", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit assignment status updated", logger.LogField("id", id), logger.LogField("status", status))
	return s.toResponse(assignment), nil
}

// toResponse converts an audit assignment model to a response DTO
func (s *AuditAssignmentService) toResponse(assignment *models.AuditAssignment) *models.AuditAssignmentResponse {
	resp := &models.AuditAssignmentResponse{
		ID:              assignment.ID.String(),
		AssignmentTitle: assignment.AssignmentTitle,
		Description:     assignment.Description,
		AuditorID:       assignment.AuditorID.String(),
		AuditPlanID:     assignment.AuditPlanID.String(),
		Status:          string(assignment.Status),
		Notes:           assignment.Notes,
		CreatedAt:       assignment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       assignment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if assignment.StartDate != nil {
		startDate := assignment.StartDate.Format("2006-01-02")
		resp.StartDate = &startDate
	}

	if assignment.EndDate != nil {
		endDate := assignment.EndDate.Format("2006-01-02")
		resp.EndDate = &endDate
	}

	return resp
}