package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditMandate struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title           string         `gorm:"type:varchar(200);not null"`
	ReferenceNumber string         `gorm:"type:varchar(100);not null"`
	MandateSource   string         `gorm:"type:varchar(100)"`
	LegalBasis      string         `gorm:"type:text"`
	EffectiveDate   time.Time
	ExpiryDate      *time.Time
	IsActive        bool           `gorm:"default:true"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// Request DTOs
type CreateAuditMandateRequest struct {
	Title           string     `json:"title" binding:"required" validate:"required,max=200"`
	ReferenceNumber string     `json:"reference_number" binding:"required" validate:"required,max=100"`
	MandateSource   string     `json:"mandate_source" validate:"max=100"`
	LegalBasis      string     `json:"legal_basis"`
	EffectiveDate   time.Time  `json:"effective_date" binding:"required"`
	ExpiryDate      *time.Time `json:"expiry_date"`
}

type UpdateAuditMandateRequest struct {
	Title           *string    `json:"title" validate:"omitempty,max=200"`
	MandateSource   *string    `json:"mandate_source" validate:"omitempty,max=100"`
	LegalBasis      *string    `json:"legal_basis"`
	EffectiveDate   *time.Time `json:"effective_date"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	IsActive        *bool      `json:"is_active"`
}

type ListAuditMandatesRequest struct {
	Page     int    `form:"page" validate:"min=1"`
	PageSize int    `form:"page_size" validate:"min=1,max=100"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`
}

// Response DTOs
type AuditMandateResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	ReferenceNumber string `json:"reference_number"`
	MandateSource   string `json:"mandate_source"`
	LegalBasis      string `json:"legal_basis"`
	EffectiveDate   string `json:"effective_date"`
	ExpiryDate      string `json:"expiry_date,omitempty"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}