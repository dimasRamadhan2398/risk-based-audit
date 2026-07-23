package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditCharter struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Filename  string         `gorm:"type:varchar(200);not null" json:"filename"`
	Version   string         `gorm:"type:varchar(20);not null" json:"version"`
	Title     string         `gorm:"type:varchar(200);not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	IsActive  bool           `gorm:"default:false" json:"is_active"`
	FileUrl   string         `gorm:"type:text" json:"file_url"`
	FileSize  int64          `gorm:"type:bigint" json:"file_size"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Request DTOs
//
//	type CreateAuditCharterRequest struct {
//		Filename string `json:"filename" binding:"required" validate:"required,max=200"`
//		Version  string `json:"version" binding:"required" validate:"required,max=20"`
//		Title    string `json:"title" binding:"required" validate:"required,max=200"`
//		Content  string `json:"content" binding:"required" validate:"required"`
//	}

// Request DTOs
type CreateAuditCharterRequest struct {
	Filename string `json:"filename" validate:"omitempty,max=200"`
	Version  string `json:"version" binding:"required" validate:"required,max=20"`
	Title    string `json:"title" binding:"required" validate:"required,max=200"`
	Content  string `json:"content"`
	IsActive *bool  `json:"is_active"`
	FileUrl  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

type UpdateAuditCharterRequest struct {
	Filename *string `json:"filename" validate:"omitempty,max=200"`
	Title    *string `json:"title" validate:"omitempty,max=200"`
	Content  *string `json:"content"`
	IsActive *bool   `json:"is_active"`
	FileUrl  *string `json:"file_url"`
	FileSize *int64  `json:"file_size"`
}

type ListAuditChartersRequest struct {
	Page     int    `form:"page" validate:"min=1"`
	PageSize int    `form:"page_size" validate:"min=1,max=100"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`
}

// Response DTOs
type AuditCharterResponse struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Version   string `json:"version"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	FileUrl   string `json:"file_url"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
