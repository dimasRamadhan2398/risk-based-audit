package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SOPDocument struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SOPNumber string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"sop_number"`
	Title     string         `gorm:"type:varchar(200);not null" json:"title"`
	LinkURL   string         `gorm:"type:varchar(500);not null" json:"link_url"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SOPDocument) TableName() string {
	return "sop_document"
}
