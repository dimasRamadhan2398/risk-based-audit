package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedPerformanceReport struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title           string         `gorm:"type:varchar(255);not null" json:"title"`
	Period          string         `gorm:"type:varchar(20);not null;default:'Tahunan'" json:"period"` // Q1, Q2, Q3, Q4, Tahunan
	Year            int            `gorm:"type:int;not null" json:"year"`
	Description     string         `gorm:"type:text" json:"description"`
	FileName        string         `gorm:"type:varchar(255);not null" json:"fileName"`
	FilePath        string         `gorm:"type:varchar(500);not null" json:"filePath"`
	FileSize        int64          `gorm:"type:bigint" json:"fileSize"`
	FileType        string         `gorm:"type:varchar(100)" json:"fileType"`
	FileContent []byte         `gorm:"type:bytea" json:"-"`
	Status          string         `gorm:"type:varchar(50);default:'Uploaded'" json:"status"`
	ParsedKPIsCount int            `gorm:"type:int;default:0" json:"parsedKpisCount"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UploadedPerformanceReport) TableName() string {
	return "uploaded_performance_reports"
}
