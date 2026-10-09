package crud

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// CRUDHandler provides generic CRUD operations for any GORM model
type CRUDHandler struct {
	DB        *gorm.DB
	ModelName string
}

// NewCRUDHandler creates a new generic CRUD handler
func NewCRUDHandler(db *gorm.DB, modelName string) *CRUDHandler {
	return &CRUDHandler{DB: db, ModelName: modelName}
}

// List returns a paginated list of records
func List(db *gorm.DB, modelName string, newSlice func() interface{}, preloads ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		search := c.Query("search")

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		offset := (page - 1) * pageSize

		items := newSlice()

		// Column names coming from the request (filters, order, search) are only
		// accepted when they exist on the model, and are passed to GORM as quoted
		// columns rather than raw SQL, to rule out SQL injection.
		columns, err := modelColumns(db, items)
		if err != nil {
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		orderBy, ok := parseOrder(c.DefaultQuery("order", "created_at DESC"), columns)
		if !ok {
			response.BadRequest(c, "Invalid order parameter")
			return
		}

		// Sort keys such as created_at are not unique (a batch insert gives every
		// row the same timestamp), and Postgres does not keep ties in a stable
		// order across LIMIT/OFFSET queries, so pages could skip or repeat rows.
		// Break ties on the primary key.
		orderBy = withIDTiebreaker(orderBy, columns)

		query := db.Model(items)

		for _, p := range preloads {
			query = query.Preload(p)
		}

		// Apply search on common text fields the model actually has
		if search != "" {
			match := ilikeAny{Value: "%" + search + "%"}
			for _, col := range []string{"title", "name"} {
				if columns[col] {
					match.Columns = append(match.Columns, col)
				}
			}
			if len(match.Columns) > 0 {
				query = query.Where(match)
			}
		}

		// Apply filters from query params; unknown columns are ignored
		for key, values := range c.Request.URL.Query() {
			if key == "page" || key == "page_size" || key == "search" || key == "order" {
				continue
			}
			col := toSnakeCase(key)
			if !columns[col] {
				continue
			}
			if len(values) > 0 && values[0] != "" {
				query = query.Where(clause.Eq{Column: clause.Column{Name: col}, Value: values[0]})
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		if err := query.Order(orderBy).Offset(offset).Limit(pageSize).Find(items).Error; err != nil {
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		response.OK(c, modelName+" fetched successfully", gin.H{
			"items": items,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// modelColumns returns the set of database column names of a model (or a
// pointer to a slice of models)
func modelColumns(db *gorm.DB, model interface{}) (map[string]bool, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return nil, err
	}
	columns := make(map[string]bool, len(stmt.Schema.Fields))
	for _, field := range stmt.Schema.Fields {
		if field.DBName != "" {
			columns[field.DBName] = true
		}
	}
	return columns, nil
}

// parseOrder turns "column [ASC|DESC]" (comma separated for several columns)
// into an ORDER BY clause, rejecting anything that is not a known column
func parseOrder(raw string, columns map[string]bool) (clause.OrderBy, bool) {
	var orderBy clause.OrderBy
	for _, part := range strings.Split(raw, ",") {
		fields := strings.Fields(part)
		if len(fields) == 0 || len(fields) > 2 {
			return orderBy, false
		}
		col := toSnakeCase(fields[0])
		if !columns[col] {
			return orderBy, false
		}
		desc := false
		if len(fields) == 2 {
			switch strings.ToUpper(fields[1]) {
			case "ASC":
			case "DESC":
				desc = true
			default:
				return orderBy, false
			}
		}
		orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc})
	}
	return orderBy, true
}

// withIDTiebreaker appends "id" to the ORDER BY when the model has an id column
// that is not already part of the ordering, so the order is total
func withIDTiebreaker(orderBy clause.OrderBy, columns map[string]bool) clause.OrderBy {
	if !columns["id"] {
		return orderBy
	}
	for _, col := range orderBy.Columns {
		if col.Column.Name == "id" {
			return orderBy
		}
	}
	orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	return orderBy
}

// ilikeAny matches Value case-insensitively against any of Columns:
// ("col1" ILIKE ? OR "col2" ILIKE ?). Column names are quoted, never interpolated.
type ilikeAny struct {
	Columns []string
	Value   string
}

func (e ilikeAny) Build(builder clause.Builder) {
	builder.WriteByte('(')
	for i, col := range e.Columns {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteQuoted(clause.Column{Name: col})
		builder.WriteString(" ILIKE ")
		builder.AddVar(builder, e.Value)
	}
	builder.WriteByte(')')
}

// GetByID returns a single record by ID
func GetByID(db *gorm.DB, modelName string, newEntity func() interface{}, preloads ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			response.BadRequest(c, "Invalid "+modelName+" ID")
			return
		}

		entity := newEntity()
		query := db

		for _, p := range preloads {
			query = query.Preload(p)
		}

		if err := query.First(entity, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.NotFound(c, modelName+" not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		response.OK(c, modelName+" fetched successfully", entity)
	}
}

// CreateHook runs inside the create transaction, after binding and before the
// INSERT. It may change entity. Returning an *errors.AppError sends that status.
type CreateHook func(tx *gorm.DB, entity interface{}) error

// UpdateHook runs inside the update transaction, before the UPDATE. existing is
// the stored record; columns are the snake_case column values about to be
// written (JSON columns as []byte), and may be changed.
type UpdateHook func(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error

func writeHookError(c *gin.Context, modelName, action string, err error) {
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.InternalServerError(c, "Failed to "+action+" "+modelName+": "+err.Error())
}

// AfterHook runs inside the write transaction, right after the INSERT or
// UPDATE. entity is the written record (for updates: the stored record with
// the written columns applied). Returning an *errors.AppError sends that
// status; any error rolls the write back.
type AfterHook func(tx *gorm.DB, entity interface{}) error

// Create creates a new record
func Create(db *gorm.DB, modelName string, newEntity func() interface{}) gin.HandlerFunc {
	return CreateWithHooks(db, modelName, newEntity, nil, nil)
}

// CreateWithHook is Create with a hook run in the same transaction as the INSERT.
func CreateWithHook(db *gorm.DB, modelName string, newEntity func() interface{}, hook CreateHook) gin.HandlerFunc {
	return CreateWithHooks(db, modelName, newEntity, hook, nil)
}

// CreateWithHooks is Create with hooks run in the same transaction as the
// INSERT: before (may change the entity) and after (sees the created row).
func CreateWithHooks(db *gorm.DB, modelName string, newEntity func() interface{}, before CreateHook, after AfterHook) gin.HandlerFunc {
	return func(c *gin.Context) {
		entity := newEntity()

		if err := c.ShouldBindJSON(entity); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		if before != nil || after != nil {
			if err := db.Transaction(func(tx *gorm.DB) error {
				if before != nil {
					if err := before(tx, entity); err != nil {
						return err
					}
				}
				if err := tx.Create(entity).Error; err != nil {
					return err
				}
				if after != nil {
					return after(tx, entity)
				}
				return nil
			}); err != nil {
				writeHookError(c, modelName, "create", err)
				return
			}
			response.Created(c, modelName+" created successfully", entity)
			return
		}

		if err := db.Create(entity).Error; err != nil {
			response.InternalServerError(c, "Failed to create "+modelName+": "+err.Error())
			return
		}

		response.Created(c, modelName+" created successfully", entity)
	}
}

// Update updates an existing record
func Update(db *gorm.DB, modelName string, newEntity func() interface{}) gin.HandlerFunc {
	return UpdateWithHooks(db, modelName, newEntity, nil, nil)
}

// UpdateWithHook is Update with a hook run in the same transaction as the UPDATE.
func UpdateWithHook(db *gorm.DB, modelName string, newEntity func() interface{}, hook UpdateHook) gin.HandlerFunc {
	return UpdateWithHooks(db, modelName, newEntity, hook, nil)
}

// UpdateWithHooks is Update with hooks run in the same transaction as the
// UPDATE: before (may change the columns) and after (sees the updated record).
func UpdateWithHooks(db *gorm.DB, modelName string, newEntity func() interface{}, before UpdateHook, after AfterHook) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			response.BadRequest(c, "Invalid "+modelName+" ID")
			return
		}

		// Find existing record
		existing := newEntity()
		if err := db.First(existing, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.NotFound(c, modelName+" not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		// Bind update data
		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// Remove protected fields
		delete(updateData, "id")
		delete(updateData, "created_at")
		delete(updateData, "deleted_at")

		// Convert camelCase keys to snake_case column names for GORM
		snakeData := make(map[string]interface{})
		for k, v := range updateData {
			switch val := v.(type) {
			case []interface{}, map[string]interface{}:
				if b, err := json.Marshal(val); err == nil {
					snakeData[toSnakeCase(k)] = b
				} else {
					snakeData[toSnakeCase(k)] = v
				}
			default:
				snakeData[toSnakeCase(k)] = v
			}
		}

		// Parse GORM schema to filter out non-existent columns and format dates
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(existing); err == nil {
			fieldsByDBName := make(map[string]*schema.Field)
			for _, field := range stmt.Schema.Fields {
				fieldsByDBName[field.DBName] = field
			}
			for col, val := range snakeData {
				field, exists := fieldsByDBName[col]
				if !exists {
					delete(snakeData, col)
					continue
				}
				// Parse date strings into time.Time only for actual time/date model fields
				isTimeField := field.DataType == schema.Time || field.FieldType.String() == "time.Time" || field.FieldType.String() == "*time.Time"
				if isTimeField && (strings.HasSuffix(col, "_date") || strings.HasSuffix(col, "_at") || strings.Contains(col, "date") || strings.Contains(col, "time")) {
					if strVal, ok := val.(string); ok {
						trimmed := strings.TrimSpace(strVal)
						if trimmed == "" || trimmed == "null" {
							snakeData[col] = nil
						} else if t, err := time.Parse("2006-01-02", trimmed); err == nil {
							snakeData[col] = t
						} else if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
							snakeData[col] = t
						} else if len(trimmed) >= 10 {
							if t, err := time.Parse("2006-01-02", trimmed[:10]); err == nil {
								snakeData[col] = t
							}
						}
					}
				}
			}
		}

		if before != nil || after != nil {
			if err := db.Transaction(func(tx *gorm.DB) error {
				if before != nil {
					if err := before(tx, existing, snakeData); err != nil {
						return err
					}
				}
				if err := tx.Model(existing).Updates(snakeData).Error; err != nil {
					return err
				}
				if after != nil {
					return after(tx, existing)
				}
				return nil
			}); err != nil {
				writeHookError(c, modelName, "update", err)
				return
			}
		} else if err := db.Model(existing).Updates(snakeData).Error; err != nil {
			response.InternalServerError(c, "Failed to update "+modelName+": "+err.Error())
			return
		}

		// Reload
		db.First(existing, "id = ?", id)
		response.OK(c, modelName+" updated successfully", existing)
	}
}

// Delete soft-deletes a record
func Delete(db *gorm.DB, modelName string, newEntity func() interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			response.BadRequest(c, "Invalid "+modelName+" ID")
			return
		}

		entity := newEntity()
		if err := db.First(entity, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.NotFound(c, modelName+" not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		if err := db.Delete(entity).Error; err != nil {
			response.InternalServerError(c, "Failed to delete "+modelName)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": modelName + " deleted successfully",
		})
	}
}
