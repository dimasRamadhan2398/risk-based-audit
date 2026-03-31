package repositories

import (
	"fmt"

	apperrors "auth-service/pkg/errors"

	"gorm.io/gorm"
)

// BaseRepository provides common database operations for all repositories
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// Create creates a new record
func (r *BaseRepository) Create(entity interface{}) error {
	if err := r.DB.Create(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// CreateTx creates a new record within a transaction
func (r *BaseRepository) CreateTx(tx *gorm.DB, entity interface{}) error {
	if err := tx.Create(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Update updates a record
func (r *BaseRepository) Update(entity interface{}) error {
	if err := r.DB.Save(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// UpdateTx updates a record within a transaction
func (r *BaseRepository) UpdateTx(tx *gorm.DB, entity interface{}) error {
	if err := tx.Save(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Delete deletes a record (soft delete if using DeletedAt)
func (r *BaseRepository) Delete(entity interface{}) error {
	if err := r.DB.Delete(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// HardDelete permanently deletes a record
func (r *BaseRepository) HardDelete(entity interface{}) error {
	if err := r.DB.Unscoped().Delete(entity).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// FindByID finds a record by ID
func (r *BaseRepository) FindByID(entity interface{}, id interface{}) error {
	if err := r.DB.First(entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}

// FindByIDWithPreload finds a record by ID with preloaded relationships
func (r *BaseRepository) FindByIDWithPreload(entity interface{}, id interface{}, preloads ...string) error {
	db := r.DB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.First(entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}

// FindOne finds a single record matching the given conditions
func (r *BaseRepository) FindOne(entity interface{}, condition interface{}, args ...interface{}) error {
	if err := r.DB.Where(condition, args...).First(entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}

// FindMany finds multiple records with pagination
func (r *BaseRepository) FindMany(entities interface{}, conditions map[string]interface{}) error {
	query := r.DB.Model(entities)

	// Apply filters if provided
	if where, ok := conditions["where"].(map[string]interface{}); ok {
		query = query.Where(where)
	}

	// Apply search if provided
	if search, ok := conditions["search"].(string); ok && search != "" {
		if fields, ok := conditions["search_fields"].([]string); ok && len(fields) > 0 {
			searchPattern := "%" + search + "%"
			query = query.Where(fmt.Sprintf("%s ILIKE ?", fields[0]), searchPattern)
			for i := 1; i < len(fields); i++ {
				query = query.Or(fmt.Sprintf("%s ILIKE ?", fields[i]), searchPattern)
			}
		}
	}

	// Apply pagination
	if offset, ok := conditions["offset"].(int); ok {
		query = query.Offset(offset)
	}
	if limit, ok := conditions["limit"].(int); ok && limit > 0 {
		query = query.Limit(limit)
	}

	// Apply ordering
	if order, ok := conditions["order"].(string); ok && order != "" {
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Find(entities).Error; err != nil {
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}

// Count counts records matching the given conditions
func (r *BaseRepository) Count(model interface{}, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.DB.Model(model)

	// Apply filters if provided
	if where, ok := conditions["where"].(map[string]interface{}); ok {
		query = query.Where(where)
	}

	// Apply search if provided
	if search, ok := conditions["search"].(string); ok && search != "" {
		if fields, ok := conditions["search_fields"].([]string); ok && len(fields) > 0 {
			searchPattern := "%" + search + "%"
			query = query.Where(fmt.Sprintf("%s ILIKE ?", fields[0]), searchPattern)
			for i := 1; i < len(fields); i++ {
				query = query.Or(fmt.Sprintf("%s ILIKE ?", fields[i]), searchPattern)
			}
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return count, nil
}

// Exists checks if a record exists
func (r *BaseRepository) Exists(model interface{}, condition interface{}, args ...interface{}) (bool, error) {
	var count int64
	if err := r.DB.Model(model).Where(condition, args...).Count(&count).Error; err != nil {
		return false, fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return count > 0, nil
}

// WithTx starts a new transaction
func (r *BaseRepository) WithTx(fn func(tx *gorm.DB) error) error {
	return r.DB.Transaction(fn)
}

// GetDB returns the underlying DB instance
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.DB
}

// Raw executes a raw SQL query
func (r *BaseRepository) Raw(result interface{}, sql string, args ...interface{}) error {
	if err := r.DB.Raw(sql, args...).Scan(result).Error; err != nil {
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}

// Exec executes a raw SQL statement
func (r *BaseRepository) Exec(sql string, args ...interface{}) error {
	if err := r.DB.Exec(sql, args...).Error; err != nil {
		return fmt.Errorf("%w: %v", apperrors.ErrDatabase, err)
	}
	return nil
}
