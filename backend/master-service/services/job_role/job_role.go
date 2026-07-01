package job_role
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/job_role"
	"github.com/google/uuid"
)
type JobRoleServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.JobRole, error)
	FindById(ctx *base.BaseService, id string) (*models.JobRole, error)
	Create(ctx *base.BaseService, role *models.JobRole) (*models.JobRole, error)
	Update(ctx *base.BaseService, id string, role *models.JobRole) (*models.JobRole, error)
	Delete(ctx *base.BaseService, id string) error
}
type JobRoleService struct { roleRepo repo.IJobRoleRepository }
func NewJobRoleService(roleRepo repo.IJobRoleRepository) JobRoleServiceInterface { return &JobRoleService{roleRepo: roleRepo} }
func (s *JobRoleService) Create(ctx *base.BaseService, role *models.JobRole) (*models.JobRole, error) {
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
func (s *JobRoleService) FindAll(ctx *base.BaseService) (*[]models.JobRole, error) {
	roles, err := s.roleRepo.FindAll()
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
	if existingRole.JobRoleCode != role.JobRoleCode {
		if _, err := s.roleRepo.FindByCode(role.JobRoleCode); err == nil {
			return nil, apperrors.Wrap("JOB_ROLE_CODE_ALREADY_EXISTS", "Job role code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate job role code", 500, err)
		}
	}
	existingRole.JobRoleCode = role.JobRoleCode
	existingRole.JobRoleName = role.JobRoleName
	existingRole.JobRoleDescription = role.JobRoleDescription
	existingRole.JobPositionType = role.JobPositionType
	existingRole.IsActive = role.IsActive
	if err := s.roleRepo.Update(existingRole); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update job role", 500, err) }
	return existingRole, nil
}
