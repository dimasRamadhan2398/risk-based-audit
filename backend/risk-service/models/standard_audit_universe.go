package models

import (
	"time"

	"github.com/google/uuid"
)

type StandardAuditUniverse struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string                 `gorm:"type:varchar(200);not null" json:"name"`
	ParentID  *uuid.UUID             `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent    *StandardAuditUniverse `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []StandardAuditUniverse `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Category  string                 `gorm:"type:varchar(100)" json:"category"` // For IT headings (e.g. IT Governance, Cybersecurity, Applications, etc.)
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func (StandardAuditUniverse) TableName() string {
	return "standard_audit_universe"
}
