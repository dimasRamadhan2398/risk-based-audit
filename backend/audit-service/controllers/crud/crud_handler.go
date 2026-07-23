package crud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
		query := db.Model(items)

		for _, p := range preloads {
			query = query.Preload(p)
		}

		// Apply search on common text fields
		if search != "" {
			query = query.Where("title ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
		}

		// Apply filters from query params
		for key, values := range c.Request.URL.Query() {
			if key == "page" || key == "page_size" || key == "search" || key == "order" {
				continue
			}
			if len(values) > 0 && values[0] != "" {
				query = query.Where(fmt.Sprintf("%s = ?", key), values[0])
			}
		}

		var total int64
		query.Count(&total)

		order := c.DefaultQuery("order", "created_at DESC")
		if err := query.Order(order).Offset(offset).Limit(pageSize).Find(items).Error; err != nil {
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

// Create creates a new record
func Create(db *gorm.DB, modelName string, newEntity func() interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		entity := newEntity()

		if err := c.ShouldBindJSON(entity); err != nil {
			response.BadRequest(c, err.Error())
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
			validCols := make(map[string]bool)
			for _, field := range stmt.Schema.Fields {
				validCols[field.DBName] = true
			}
			for col, val := range snakeData {
				if !validCols[col] {
					delete(snakeData, col)
					continue
				}
				// Parse date strings into time.Time for time/date fields
				if strings.HasSuffix(col, "_date") || strings.HasSuffix(col, "_at") {
					if strVal, ok := val.(string); ok && strVal != "" {
						if t, err := time.Parse("2006-01-02", strVal); err == nil {
							snakeData[col] = t
						} else if t, err := time.Parse(time.RFC3339, strVal); err == nil {
							snakeData[col] = t
						}
					}
				}
			}
		}

		if err := db.Model(existing).Updates(snakeData).Error; err != nil {
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
