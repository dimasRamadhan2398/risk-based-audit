package repositories

import (
	"strings"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AuditCharterRepositoryInterface defines the audit charter repository interface
type AuditCharterRepositoryInterface interface {
	Create(charter *models.AuditCharter) error
	Update(charter *models.AuditCharter) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditCharter, error)
	FindByVersion(version string) (*models.AuditCharter, error)
	FindActive() (*models.AuditCharter, error)
	FindMany(offset, limit int, search string, isActive *bool) ([]*models.AuditCharter, error)
	Count(search string, isActive *bool) (int64, error)
}

// AuditCharterRepository handles audit charter data operations
type AuditCharterRepository struct {
	*BaseRepository
}

// NewAuditCharterRepository creates a new audit charter repository
func NewAuditCharterRepository(baseRepo *BaseRepository) AuditCharterRepositoryInterface {
	return &AuditCharterRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new audit charter
func (r *AuditCharterRepository) Create(charter *models.AuditCharter) error {
	return r.BaseRepository.Create(charter)
}

// Update updates an audit charter
func (r *AuditCharterRepository) Update(charter *models.AuditCharter) error {
	return r.BaseRepository.Update(charter)
}

// Delete deletes an audit charter
func (r *AuditCharterRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.AuditCharter{ID: id})
}

// FindByID finds an audit charter by ID
func (r *AuditCharterRepository) FindByID(id uuid.UUID) (*models.AuditCharter, error) {
	var charter models.AuditCharter
	if err := r.GetDB().First(&charter, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// FindByVersion finds an audit charter by version
func (r *AuditCharterRepository) FindByVersion(version string) (*models.AuditCharter, error) {
	var charter models.AuditCharter
	if err := r.GetDB().Where("version = ?", version).First(&charter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// FindActive finds the active audit charter
func (r *AuditCharterRepository) FindActive() (*models.AuditCharter, error) {
	var charter models.AuditCharter
	if err := r.GetDB().Where("is_active = ?", true).Order("created_at DESC").First(&charter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// charterSearchColumns are the audit_charters columns matched by ?search=
var charterSearchColumns = []string{"title", "version", "filename"}

// charterListOrder is newest first, with id as a tiebreaker so pages never
// overlap or skip rows when created_at values are equal
var charterListOrder = clause.OrderBy{Columns: []clause.OrderByColumn{
	{Column: clause.Column{Name: "created_at"}, Desc: true},
	{Column: clause.Column{Name: "id"}, Desc: true},
}}

// lowerLikeAny matches Value case-insensitively against any of Columns:
// (LOWER("col1") LIKE ? OR LOWER("col2") LIKE ?). Portable across Postgres
// and SQLite; column names are quoted, never interpolated.
type lowerLikeAny struct {
	Columns []string
	Value   string
}

func (e lowerLikeAny) Build(builder clause.Builder) {
	builder.WriteByte('(')
	for i, col := range e.Columns {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString("LOWER(")
		builder.WriteQuoted(clause.Column{Name: col})
		builder.WriteString(") LIKE ")
		builder.AddVar(builder, e.Value)
	}
	builder.WriteByte(')')
}

// listQuery applies the shared FindMany/Count filters so rows and total agree
func (r *AuditCharterRepository) listQuery(search string, isActive *bool) *gorm.DB {
	query := r.GetDB().Model(&models.AuditCharter{})

	if search = strings.TrimSpace(search); search != "" {
		query = query.Where(lowerLikeAny{Columns: charterSearchColumns, Value: "%" + strings.ToLower(search) + "%"})
	}

	if isActive != nil {
		query = query.Where(clause.Eq{Column: clause.Column{Name: "is_active"}, Value: *isActive})
	}

	return query
}

// FindMany finds multiple audit charters with filters
func (r *AuditCharterRepository) FindMany(offset, limit int, search string, isActive *bool) ([]*models.AuditCharter, error) {
	var charters []*models.AuditCharter

	if err := r.listQuery(search, isActive).Order(charterListOrder).Offset(offset).Limit(limit).Find(&charters).Error; err != nil {
		return nil, err
	}

	return charters, nil
}

// Count counts audit charters with filters
func (r *AuditCharterRepository) Count(search string, isActive *bool) (int64, error) {
	var count int64

	if err := r.listQuery(search, isActive).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
