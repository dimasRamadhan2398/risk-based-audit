package job_role
import (
	"strings"
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/job_role"
	"github.com/google/uuid"
)
type JobRoleServiceInterface interface {
	FindAll(ctx *base.BaseService, activeOnly bool) (*[]models.JobRole, error)
	AccessRoles() []string
	FindById(ctx *base.BaseService, id string) (*models.JobRole, error)
	Create(ctx *base.BaseService, role *models.JobRole) (*models.JobRole, error)
	Update(ctx *base.BaseService, id string, role *models.JobRole) (*models.JobRole, error)
	Delete(ctx *base.BaseService, id string) error
}
type JobRoleService struct { roleRepo repo.IJobRoleRepository }
func NewJobRoleService(roleRepo repo.IJobRoleRepository) JobRoleServiceInterface { return &JobRoleService{roleRepo: roleRepo} }
// validateInput normalizes the access role and checks the unique name. exclude
// is the role being updated (nil on create).
func (s *JobRoleService) validateInput(role *models.JobRole, exclude *models.JobRole) error {
	if role == nil { return apperrors.New(apperrors.CodeInvalidRequestBody, apperrors.MsgInvalidRequestBody, 400) }
	role.JobRoleCode = strings.TrimSpace(role.JobRoleCode)
	role.JobRoleName = strings.TrimSpace(role.JobRoleName)
	n, ok := models.NormalizeAccessRole(role.DefaultAccessRole)
	if !ok {
		return apperrors.ValidationFailed("default_access_role must be one of: "+strings.Join(models.AllowedAccessRoles(), ", ")+" (or empty for no mapping).",
			map[string]string{"default_access_role": apperrors.FieldInvalid})
	}
	role.DefaultAccessRole = n
	if existing, err := s.roleRepo.FindByName(role.JobRoleName); err == nil {
		if exclude == nil || existing.ID != exclude.ID {
			return apperrors.Wrap("JOB_ROLE_NAME_ALREADY_EXISTS", "Job role name already exists", 409, nil)
		}
	} else if err != apperrors.ErrNotFound {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate job role name", 500, err)
	}
	return nil
}

func (s *JobRoleService) Create(ctx *base.BaseService, role *models.JobRole) (*models.JobRole, error) {
	if err := s.validateInput(role, nil); err != nil { return nil, err }
	if _, err := s.roleRepo.FindByCode(role.JobRoleCode); err == nil {
		return nil, apperrors.Wrap("JOB_ROLE_CODE_ALREADY_EXISTS", "Job role code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate job role code", 500, err)
	}
	if err := s.roleRepo.Create(role); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create job role", 500, err) }
	return role, nil
}
func (s *JobRoleService) Delete(ctx *base.BaseService, id string) error {
	roleID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_JOB_ROLE_ID", "Invalid job role ID format", 400, err) }
	if _, err := s.roleRepo.FindByID(roleID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find job role", 500, err)
	}
	if err := s.roleRepo.Delete(roleID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete job role", 500, err) }
	return nil
}
func (s *JobRoleService) AccessRoles() []string { return models.AllowedAccessRoles() }

func (s *JobRoleService) FindAll(ctx *base.BaseService, activeOnly bool) (*[]models.JobRole, error) {
	roles, err := s.roleRepo.FindAll(activeOnly)
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch job roles", 500, err) }
	result := make([]models.JobRole, 0, len(roles))
	for _, role := range roles { if role != nil { result = append(result, *role) } }
	return &result, nil
}
func (s *JobRoleService) FindById(ctx *base.BaseService, id string) (*models.JobRole, error) {
	roleID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_JOB_ROLE_ID", "Invalid job role ID format", 400, err) }
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch job role", 500, err)
	}
	return role, nil
}
func (s *JobRoleService) Update(ctx *base.BaseService, id string, role *models.JobRole) (*models.JobRole, error) {
	roleID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_JOB_ROLE_ID", "Invalid job role ID format", 400, err) }
	existingRole, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch job role", 500, err)
	}
	// The code is immutable in the UI and may be omitted from the payload; keep the stored one.
	role.JobRoleCode = strings.TrimSpace(role.JobRoleCode)
	if role.JobRoleCode == "" { role.JobRoleCode = existingRole.JobRoleCode }
	if role.JobRoleCode == "" {
		return nil, apperrors.ValidationFailed("job_role_code is required.", map[string]string{"job_role_code": apperrors.FieldRequired})
	}
	if existingRole.JobRoleCode != role.JobRoleCode {
		if _, err := s.roleRepo.FindByCode(role.JobRoleCode); err == nil {
			return nil, apperrors.Wrap("JOB_ROLE_CODE_ALREADY_EXISTS", "Job role code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate job role code", 500, err)
		}
	}
	if err := s.validateInput(role, existingRole); err != nil { return nil, err }
	existingRole.JobRoleCode = role.JobRoleCode
	existingRole.DefaultAccessRole = role.DefaultAccessRole
	existingRole.JobRoleName = role.JobRoleName
	existingRole.JobRoleDescription = role.JobRoleDescription
	existingRole.JobPositionType = role.JobPositionType
	existingRole.IsActive = role.IsActive
	if err := s.roleRepo.Update(existingRole); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update job role", 500, err) }
	return existingRole, nil
}
