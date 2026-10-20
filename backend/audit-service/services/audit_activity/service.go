package services

import (
	"audit-service/models"
	"audit-service/pkg/errors"
	"audit-service/pkg/logger"
	"audit-service/pkg/utils"
	"audit-service/repositories"
	"audit-service/services/base"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// AuditActivityServiceInterface defines the audit activity service interface
type AuditActivityServiceInterface interface {
	CreateActivity(ctx context.Context, req *models.CreateActivityPlanRequest) (*models.ActivityPlanResponse, error)
	UpdateActivity(ctx context.Context, id uuid.UUID, req *models.UpdateActivityPlanRequest) (*models.ActivityPlanResponse, error)
	DeleteActivity(ctx context.Context, id uuid.UUID) error
	GetActivity(ctx context.Context, id uuid.UUID) (*models.ActivityPlanResponse, error)
	GetActivityByProjectCode(ctx context.Context, projectCode string) (*models.ActivityPlanResponse, error)
	ListActivities(ctx context.Context, req *models.ListActivityPlansRequest) ([]*models.ActivityPlanResponse, *utils.PaginationResponse, error)
}

// AuditActivityService handles audit activity business logic
type AuditActivityService struct {
	*base.BaseService
	repo repositories.AuditActivityRepositoryInterface
}

// NewAuditActivityService creates a new audit activity service
func NewAuditActivityService(repo repositories.AuditActivityRepositoryInterface) AuditActivityServiceInterface {
	return &AuditActivityService{
		BaseService: base.NewBaseService(),
		repo:        repo,
	}
}

// CreateActivity creates a new audit activity. The project code (Activity ID)
// is generated: {audit type code}-{annual plan year}-{n}, e.g. ASR-2026-003.
// A project_code sent by the client is ignored.
func (s *AuditActivityService) CreateActivity(ctx context.Context, req *models.CreateActivityPlanRequest) (*models.ActivityPlanResponse, error) {
	year, err := s.repo.AnnualPlanYear(req.AnnualPlanID)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.New(errors.ErrValidation.Code, "Annual plan not found", errors.ErrValidation.StatusCode)
		}
		s.LogError("Failed to read annual plan year", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}
	if year <= 0 {
		year = req.PlannedStart.Year()
	}

	status := req.Status
	if status == "" {
		status = "PLANNED"
	}

	targetUnitID := req.TargetUnitID
	activity := &models.AuditActivity{
		AnnualPlanID: req.AnnualPlanID,
		TargetUnitID: &targetUnitID,
		AuditType:    strings.TrimSpace(req.AuditType),
		Title:        req.Title,
		Objective:    req.Objective,
		Scope:        req.Scope,
		PlannedStart: req.PlannedStart,
		PlannedEnd:   req.PlannedEnd,
		Status:       status,
	}

	if err := s.repo.CreateWithGeneratedCode(activity, year); err != nil {
		s.LogError("Failed to create audit activity", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit activity created", logger.LogField("id", activity.ID))
	return s.toResponse(activity), nil
}

// ErrCodeManagedByPlan is returned when a change to a row that belongs to an
// Audit Activity Plan (RKA) touches a field the plan or its PKAT owns.
const ErrCodeManagedByPlan = "ACTIVITY_ROW_MANAGED_BY_PLAN"

// planManagedChange names the first field of req that would change a row
// managed by an RKA (title, target unit, planned dates), or "" if none. Those
// values follow the PKAT and the RKA form; status stays writable because the
// execution sync and the audit lifecycle drive it.
func planManagedChange(activity *models.AuditActivity, req *models.UpdateActivityPlanRequest) string {
	if activity.ActivityPlanID == nil || *activity.ActivityPlanID == uuid.Nil {
		return ""
	}
	if req.Title != nil && *req.Title != activity.Title {
		return "title"
	}
	if req.TargetUnitID != nil && (activity.TargetUnitID == nil || *req.TargetUnitID != *activity.TargetUnitID) {
		return "target_unit_id"
	}
	if req.PlannedStart != nil && !req.PlannedStart.Equal(activity.PlannedStart) {
		return "planned_start"
	}
	if req.PlannedEnd != nil && !req.PlannedEnd.Equal(activity.PlannedEnd) {
		return "planned_end"
	}
	return ""
}

// UpdateActivity updates an audit activity
func (s *AuditActivityService) UpdateActivity(ctx context.Context, id uuid.UUID, req *models.UpdateActivityPlanRequest) (*models.ActivityPlanResponse, error) {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if field := planManagedChange(activity, req); field != "" {
		return nil, errors.New(ErrCodeManagedByPlan,
			field+" of this activity is managed by its Audit Activity Plan and the PKAT; change it there",
			http.StatusConflict)
	}

	if req.TargetUnitID != nil {
		targetUnitID := *req.TargetUnitID
		activity.TargetUnitID = &targetUnitID
	}
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Objective != nil {
		activity.Objective = *req.Objective
	}
	if req.Scope != nil {
		activity.Scope = *req.Scope
	}
	if req.PlannedStart != nil {
		activity.PlannedStart = *req.PlannedStart
	}
	if req.PlannedEnd != nil {
		activity.PlannedEnd = *req.PlannedEnd
	}
	if req.Status != nil {
		activity.Status = *req.Status
	}

	if err := s.repo.Update(activity); err != nil {
		s.LogError("Failed to update audit activity", logger.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Audit activity updated", logger.LogField("id", activity.ID))
	return s.toResponse(activity), nil
}

// DeleteActivity deletes an audit activity
func (s *AuditActivityService) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		s.LogError("Failed to delete audit activity", logger.LogField("error", err))
		return err
	}

	s.LogInfo("Audit activity deleted", logger.LogField("id", id))
	return nil
}

// GetActivity retrieves an audit activity by ID
func (s *AuditActivityService) GetActivity(ctx context.Context, id uuid.UUID) (*models.ActivityPlanResponse, error) {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(activity), nil
}

// GetActivityByProjectCode retrieves an audit activity by project code
func (s *AuditActivityService) GetActivityByProjectCode(ctx context.Context, projectCode string) (*models.ActivityPlanResponse, error) {
	activity, err := s.repo.FindByProjectCode(projectCode)
	if err != nil {
		return nil, err
	}

	return s.toResponse(activity), nil
}

// ListActivities retrieves a list of audit activities with pagination
func (s *AuditActivityService) ListActivities(ctx context.Context, req *models.ListActivityPlansRequest) ([]*models.ActivityPlanResponse, *utils.PaginationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize, "created_at", "desc")

	activities, err := s.repo.FindMany(pagination.GetOffset(), pagination.GetLimit(), req.Search, req.AnnualPlanID, req.TargetUnitID, req.Status)
	if err != nil {
		s.LogError("Failed to list audit activities", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	totalCount, err := s.repo.Count(req.Search, req.AnnualPlanID, req.TargetUnitID, req.Status)
	if err != nil {
		s.LogError("Failed to count audit activities", logger.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	responses := make([]*models.ActivityPlanResponse, len(activities))
	for i, activity := range activities {
		responses[i] = s.toResponse(activity)
	}

	paginationResp := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, totalCount)
	return responses, paginationResp, nil
}

// uuidText is the id as text, "" for nil or the zero UUID.
func uuidText(id *uuid.UUID) string {
	if id == nil || *id == uuid.Nil {
		return ""
	}
	return id.String()
}

// toResponse converts an audit activity model to a response DTO
func (s *AuditActivityService) toResponse(activity *models.AuditActivity) *models.ActivityPlanResponse {
	return &models.ActivityPlanResponse{
		ID:           activity.ID.String(),
		AnnualPlanID: activity.AnnualPlanID.String(),
		TargetUnitID: uuidText(activity.TargetUnitID),
		ProjectCode:  activity.ProjectCode,
		AuditType:    activity.AuditType,
		Title:        activity.Title,
		Objective:    activity.Objective,
		Scope:        activity.Scope,
		PlannedStart: activity.PlannedStart.Format("2006-01-02"),
		PlannedEnd:   activity.PlannedEnd.Format("2006-01-02"),
		Status:       activity.Status,
		CreatedAt:    activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    activity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
