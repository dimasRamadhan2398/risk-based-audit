package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditSop struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string          `gorm:"type:varchar(255);not null" json:"name"`
	GuidelineID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"guideline_id"`
	Guideline     *AuditGuideline `gorm:"foreignKey:GuidelineID" json:"guideline,omitempty"`
	Status        string          `gorm:"type:varchar(50);default:'Aktif'" json:"status"` // "Aktif" or "Sedang Diperbarui"
	EffectiveDate string          `gorm:"type:varchar(50);not null" json:"effective_date"` // e.g. "2026-06"
	FileUrl       string          `gorm:"type:text" json:"file_url"`
	FileName      string          `gorm:"type:varchar(255)" json:"file_name"`
	FileSize      int64           `gorm:"type:bigint" json:"file_size"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (AuditSop) TableName() string {
	return "audit_sops"
}
