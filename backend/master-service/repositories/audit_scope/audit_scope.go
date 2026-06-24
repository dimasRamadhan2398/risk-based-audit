package audit_scope
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditScopeRepository interface {
	Create(scope *models.AuditScope) error
	Update(scope *models.AuditScope) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditScope, error)
	FindByCode(code string) (*models.AuditScope, error)
	FindAll() ([]*models.AuditScope, error)
}
type AuditScopeRepository struct { *repositories.BaseRepository }
func NewAuditScopeRepository(db *gorm.DB) IAuditScopeRepository { return &AuditScopeRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditScopeRepository) Create(scope *models.AuditScope) error { return r.BaseRepository.Create(scope) }
func (r *AuditScopeRepository) Update(scope *models.AuditScope) error { return r.BaseRepository.Update(scope) }
func (r *AuditScopeRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditScope{ID: id}) }
func (r *AuditScopeRepository) FindByID(id uuid.UUID) (*models.AuditScope, error) {
	var scope models.AuditScope
	if err := r.GetDB().Preload("AuditPlan").Preload("Department").Preload("RiskRegister").First(&scope, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &scope, nil
}
func (r *AuditScopeRepository) FindByCode(code string) (*models.AuditScope, error) {
	var scope models.AuditScope
	if err := r.GetDB().Where("scope_code = ?", code).First(&scope).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &scope, nil
}
func (r *AuditScopeRepository) FindAll() ([]*models.AuditScope, error) {
	var scopes []*models.AuditScope
	if err := r.GetDB().Preload("AuditPlan").Preload("Department").Preload("RiskRegister").Find(&scopes).Error; err != nil { return nil, err }
	return scopes, nil
}
