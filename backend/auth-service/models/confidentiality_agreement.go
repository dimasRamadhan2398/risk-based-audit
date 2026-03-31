package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgreementType represents the type of confidentiality agreement
type AgreementType string

const (
	AgreementTypeSystem  AgreementType = "SYSTEM"
	AgreementTypeRole   AgreementType = "ROLE"
	AgreementTypeProject AgreementType = "PROJECT"
)

// ConfidentialityAgreement represents a user's acceptance of confidentiality agreement
type ConfidentialityAgreement struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"-"`
	AgreementType  AgreementType  `gorm:"type:varchar(20);not null" json:"agreement_type"`
	Title          string         `gorm:"type:varchar(200);not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Version        string         `gorm:"type:varchar(20);not null" json:"version"`
	IsAccepted     bool           `gorm:"default:false" json:"is_accepted"`
	AcceptedAt     *time.Time     `json:"accepted_at,omitempty"`
	IPAddress      string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent      string         `gorm:"type:varchar(255)" json:"user_agent"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name
func (ConfidentialityAgreement) TableName() string {
	return "confidentiality_agreements"
}
