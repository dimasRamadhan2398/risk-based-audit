package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityCategory string

const (
	CategoryAssurance          ActivityCategory = "ASSURANCE"
	CategoryConsulting         ActivityCategory = "CONSULTING"
	CategoryExternalAssistance ActivityCategory = "EXTERNAL_ASSISTANCE"
	CategoryOther              ActivityCategory = "OTHER"
)

type AnnualAuditPlanActivity struct {
	ID                 uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnualAuditPlanID  uuid.UUID        `gorm:"type:uuid;not null;index" json:"annual_audit_plan_id"`
	ItemNumber         string           `gorm:"type:varchar(20)" json:"item_number"` // "1.1", "1.2", "2.1"
	Category           ActivityCategory `gorm:"type:varchar(30);not null" json:"category"`
	GroupTitle         string           `gorm:"type:varchar(200)" json:"group_title"` // "Audit Kantor Pusat & Unit Kerja"
	Title              string           `gorm:"type:varchar(250);not null" json:"title"`
	
	// Involved Departments (Many-to-Many)
	InvolvedDepartments []Department    `gorm:"many2many:annual_audit_activity_departments;" json:"involved_departments"`

	TimelineText       string           `gorm:"type:varchar(100)" json:"timeline_text"` // "Jan - Feb", "Mei", "Tiap Kuartal"
	AuditorCount       int              `gorm:"type:int;default:1" json:"auditor_count"`
	TotalMandays       int              `gorm:"type:int;default:0" json:"total_mandays"`
	SupervisorName     string           `gorm:"type:varchar(150)" json:"supervisor_name"`
	NotesObjective     string           `gorm:"type:text" json:"notes_objective"`

	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}
