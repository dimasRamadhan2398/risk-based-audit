package audit_scope
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/audit_scope"
	"github.com/google/uuid"
)
type AuditScopeServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AuditScope, error)
	FindById(ctx *base.BaseService, id string) (*models.AuditScope, error)
	Create(ctx *base.BaseService, scope *models.AuditScope) (*models.AuditScope, error)
	Update(ctx *base.BaseService, id string, scope *models.AuditScope) (*models.AuditScope, error)
	Delete(ctx *base.BaseService, id string) error
}
type AuditScopeService struct { scopeRepo repo.IAuditScopeRepository }
func NewAuditScopeService(scopeRepo repo.IAuditScopeRepository) AuditScopeServiceInterface { return &AuditScopeService{scopeRepo: scopeRepo} }
func (s *AuditScopeService) Create(ctx *base.BaseService, scope *models.AuditScope) (*models.AuditScope, error) {
	if _, err := s.scopeRepo.FindByCode(scope.ScopeCode); err == nil {
		return nil, apperrors.Wrap("AUDIT_SCOPE_CODE_ALREADY_EXISTS", "Audit scope code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit scope code", 500, err)
	}
	if err := s.scopeRepo.Create(scope); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit scope", 500, err) }
	return s.scopeRepo.FindByID(scope.ID)
}
func (s *AuditScopeService) Delete(ctx *base.BaseService, id string) error {
	scopeID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_AUDIT_SCOPE_ID", "Invalid audit scope ID format", 400, err) }
	if _, err := s.scopeRepo.FindByID(scopeID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find audit scope", 500, err)
	}
	if err := s.scopeRepo.Delete(scopeID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete audit scope", 500, err) }
	return nil
}
func (s *AuditScopeService) FindAll(ctx *base.BaseService) (*[]models.AuditScope, error) {
	scopes, err := s.scopeRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit scopes", 500, err) }
	result := make([]models.AuditScope, 0, len(scopes))
	for _, scope := range scopes { if scope != nil { result = append(result, *scope) } }
	return &result, nil
}
func (s *AuditScopeService) FindById(ctx *base.BaseService, id string) (*models.AuditScope, error) {
	scopeID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_SCOPE_ID", "Invalid audit scope ID format", 400, err) }
	scope, err := s.scopeRepo.FindByID(scopeID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit scope", 500, err)
	}
	return scope, nil
}
func (s *AuditScopeService) Update(ctx *base.BaseService, id string, scope *models.AuditScope) (*models.AuditScope, error) {
	scopeID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_SCOPE_ID", "Invalid audit scope ID format", 400, err) }
	existingScope, err := s.scopeRepo.FindByID(scopeID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit scope", 500, err)
	}
	if existingScope.ScopeCode != scope.ScopeCode {
		if _, err := s.scopeRepo.FindByCode(scope.ScopeCode); err == nil {
			return nil, apperrors.Wrap("AUDIT_SCOPE_CODE_ALREADY_EXISTS", "Audit scope code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit scope code", 500, err)
		}
	}
	existingScope.ScopeCode = scope.ScopeCode
	existingScope.ScopeName = scope.ScopeName
	existingScope.ScopeType = scope.ScopeType
	existingScope.Description = scope.Description
	existingScope.AuditPlanID = scope.AuditPlanID
	existingScope.InScope = scope.InScope
	existingScope.OutOfScope = scope.OutOfScope
	existingScope.DepartmentID = scope.DepartmentID
	existingScope.RiskRegisterID = scope.RiskRegisterID
	existingScope.Objectives = scope.Objectives
	existingScope.IsActive = scope.IsActive
	if err := s.scopeRepo.Update(existingScope); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit scope", 500, err) }
	return s.scopeRepo.FindByID(scopeID)
}
