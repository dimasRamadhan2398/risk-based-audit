package role

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	pkgKafka "auth-service/pkg/kafka"
	"auth-service/pkg/redis"
	role "auth-service/repositories"
	"context"

	casbin "auth-service/pkg/permissions"

	"github.com/google/uuid"
)

type RoleService struct {
	roleRepo role.RoleRepository
	publisher pkgKafka.IEventPublisher
	enforcer *casbin.CasbinEnforcer
	redis *redis.Client
}

func NewRoleService(roleRepo role.RoleRepository, publisher pkgKafka.IEventPublisher, redis *redis.Client) IRoleService {
	return &RoleService{
		roleRepo: roleRepo,
		publisher: publisher,
		redis: redis,
	}
}

type IRoleService interface{
	CreateRole(ctx context.Context, req models.CreateRoleRequest) (*models.Role, error)
    UpdateRole(ctx context.Context, id uuid.UUID, req models.UpdateRoleRequest) (*models.Role, error)
    DeleteRole(ctx context.Context, id uuid.UUID) error
    GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
    GetAllRoles(ctx context.Context) ([]*models.Role, error)
    GetRoles(ctx context.Context, offset, limit int, search string) (*models.PaginatedResponse, error)
    AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
    RemovePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error	
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, req models.CreateRoleRequest) (*models.Role, error) {
    // check duplicate name
    existing, _ := s.roleRepo.FindByName(req.Name)
    if existing != nil {
        return nil, apperrors.Wrap("ROLE_ALREADY_EXISTS", "Role with this name already exists", 400, nil)
    }

    role := &models.Role{
        ID:          uuid.New(),
        Name:        req.Name,
        Description: req.Description,
    }

    if err := s.roleRepo.Create(role); err != nil {
        return nil, apperrors.Wrap("CREATE_ROLE_FAILED", "Failed to create role", 500, nil)
    }

    s.publisher.PublishRoleEvent(ctx, "", "", role.Name, pkgKafka.EventRoleAssigned)

    return role, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, req models.UpdateRoleRequest) (*models.Role, error) {
    role, err := s.roleRepo.FindByID(id)
    if err != nil {
        return nil, apperrors.Wrap("ROLE_NOT_FOUND", "Role not found", 404, nil)
    }

    // check duplicate name if name is being changed
    if req.Name != "" && req.Name != role.Name {
        existing, _ := s.roleRepo.FindByName(req.Name)
        if existing != nil {
            return nil, apperrors.Wrap("ROLE_ALREADY_EXISTS", "Role with this name already exists", 400, nil)
        }
        role.Name = req.Name
    }

    if req.Description != "" {
        role.Description = req.Description
    }

    if err := s.roleRepo.Update(role); err != nil {
        return nil, apperrors.Wrap("UPDATE_ROLE_FAILED", "Failed to update role", 500, nil)
    }

    return role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
    role, err := s.roleRepo.FindByID(id)
    if err != nil {
        return apperrors.Wrap("ROLE_NOT_FOUND", "Role not found", 404, nil)
    }

    // prevent deleting system roles
    systemRoles := map[string]bool{
        "admin":           true,
        "auditor":         true,
        "department_head": true,
        "auditee":         true,
        "viewer":          true,
    }
    if systemRoles[role.Name] {
        return apperrors.Wrap("CANNOT_DELETE_SYSTEM_ROLE", "System roles cannot be deleted", 400, nil)
    }

    if err := s.roleRepo.Delete(id); err != nil {
        return apperrors.Wrap("DELETE_ROLE_FAILED", "Failed to delete role", 500, nil)
    }

    s.publisher.PublishRoleEvent(ctx, "", "", role.Name, pkgKafka.EventRoleRevoked)

    return nil
}

// GetRoleByID gets a role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
    role, err := s.roleRepo.FindByID(id)
    if err != nil {
        return nil, apperrors.Wrap("ROLE_NOT_FOUND", "Role not found", 404, nil)
    }
    return role, nil
}

// GetAllRoles returns all roles without pagination — for dropdowns
func (s *RoleService) GetAllRoles(ctx context.Context) ([]*models.Role, error) {
    roles, err := s.roleRepo.FindAll()
    if err != nil {
        return nil, apperrors.Wrap("FETCH_ROLES_FAILED", "Failed to fetch roles", 500, nil)
    }
    return roles, nil
}

// GetRoles returns paginated roles with optional search
func (s *RoleService) GetRoles(ctx context.Context, offset, limit int, search string) (*models.PaginatedResponse, error) {
    roles, err := s.roleRepo.FindMany(offset, limit, search)
    if err != nil {
        return nil, apperrors.Wrap("FETCH_ROLES_FAILED", "Failed to fetch roles", 500, nil)
    }

    total, err := s.roleRepo.Count(search)
    if err != nil {
        return nil, apperrors.Wrap("COUNT_ROLES_FAILED", "Failed to count roles", 500, nil)
    }

    return &models.PaginatedResponse{
        Data:   roles,
        Total:  total,
        Offset: offset,
        Limit:  limit,
    }, nil
}

// AssignPermissions assigns permissions to a role and syncs to Casbin
func (s *RoleService) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
    role, err := s.roleRepo.FindByID(roleID)
    if err != nil {
        return apperrors.Wrap("ROLE_NOT_FOUND", "Role not found", 404, nil)
    }

    if err := s.roleRepo.AssignPermissions(roleID, permissionIDs); err != nil {
        return apperrors.Wrap("ASSIGN_PERMISSIONS_FAILED", "Failed to assign permissions", 500, nil)
    }

    // sync to casbin — reload policy so new permissions take effect immediately
    if err := s.enforcer.LoadPolicy(); err != nil {
        // don't fail — DB is source of truth, casbin will sync on next reload
        s.publisher.PublishSecurityEvent(ctx, "", "", pkgKafka.EventPermissionDenied, "", map[string]interface{}{
            "reason":  "casbin policy reload failed after permission assignment",
            "role":    role.Name,
        })
    }

    s.publisher.PublishRoleEvent(ctx, "", "", role.Name, pkgKafka.EventRoleAssigned)

    return nil
}

// RemovePermissions removes permissions from a role and syncs to Casbin
func (s *RoleService) RemovePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
    role, err := s.roleRepo.FindByID(roleID)
    if err != nil {
        return apperrors.Wrap("ROLE_NOT_FOUND", "Role not found", 404, nil)
    }

    if err := s.roleRepo.RemovePermissions(roleID, permissionIDs); err != nil {
        return apperrors.Wrap("REMOVE_PERMISSIONS_FAILED", "Failed to remove permissions", 500, nil)
    }

    // sync casbin immediately — removing permissions is a security action
    if err := s.enforcer.LoadPolicy(); err != nil {
        s.publisher.PublishSecurityEvent(ctx, "", "", pkgKafka.EventPermissionDenied, "", map[string]interface{}{
            "reason": "casbin policy reload failed after permission removal",
            "role":   role.Name,
        })
    }

    s.publisher.PublishRoleEvent(ctx, "", "", role.Name, pkgKafka.EventRoleRevoked)

    return nil
}