package crud

import (
	"fmt"
	"net/http"
	"strconv"

	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
			pattern := "%" + search + "%"
			query = query.Where("title ILIKE ? OR CAST(id AS TEXT) ILIKE ?", pattern, pattern)
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
		countQuery := *query
		countQuery.Count(&total)

		order := c.DefaultQuery("order", "created_at DESC")
		if err := query.Order(order).Offset(offset).Limit(pageSize).Find(items).Error; err != nil {
			response.InternalServerError(c, "Failed to fetch "+modelName)
			return
		}

		response.OK(c, modelName+" retrieved successfully", gin.H{
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

		response.OK(c, modelName+" retrieved successfully", entity)
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
			response.InternalServerError(c, "Failed to create "+modelName)
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

		if err := db.Model(existing).Updates(updateData).Error; err != nil {
			response.InternalServerError(c, "Failed to update "+modelName)
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
