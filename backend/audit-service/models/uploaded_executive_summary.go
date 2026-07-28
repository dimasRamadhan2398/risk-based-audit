package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedExecutiveSummary struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	FileName    string         `gorm:"type:varchar(255);not null" json:"fileName"`
	FilePath    string         `gorm:"type:varchar(500);not null" json:"filePath"`
	FileSize    int64          `gorm:"type:bigint" json:"fileSize"`
	FileType    string         `gorm:"type:varchar(100)" json:"fileType"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UploadedExecutiveSummary) TableName() string {
	return "uploaded_executive_summaries"
}
