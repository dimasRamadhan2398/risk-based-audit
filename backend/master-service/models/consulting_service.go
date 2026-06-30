package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConsultingAttachment struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	UploadedAt string `json:"uploadedAt"`
	FilePath   string `json:"filePath,omitempty"`
}

type ConsultingService struct {
	ID                uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title             string                `gorm:"type:varchar(255);not null" json:"title"`
	Category          string                `gorm:"type:varchar(100);not null" json:"category"` // Operational Advisory, IT Advisory, Training, Policy Review
	RequestorDept     string                `gorm:"type:varchar(255);not null" json:"requestorDept"`
	Period            string                `gorm:"type:varchar(50);not null" json:"period"` // Q1 2026
	ConsultantName    string                `gorm:"type:varchar(255);not null" json:"consultantName"`
	Status            string                `gorm:"type:varchar(50);default:'Planned'" json:"status"` // Planned, In Progress, Completed
	Notes             string                `gorm:"type:text" json:"notes,omitempty"`
	Attachment        *ConsultingAttachment `gorm:"serializer:json" json:"attachment,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	DeletedAt         gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (ConsultingService) TableName() string {
	return "consulting_services"
}
