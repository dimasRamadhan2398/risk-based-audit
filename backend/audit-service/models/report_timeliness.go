package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportTimeliness struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Year                   int       `json:"year"`
	Period                 string    `json:"period"`
	TotalReportsPlanned    int       `json:"total_reports_planned"`
	TotalReportsCompleted  int       `json:"total_reports_completed"`
	ReportsCompletedOnTime int       `json:"reports_completed_on_time"`
	TimelinessPercentage   float64   `json:"timeliness_percentage"`
	Remarks                string    `json:"remarks"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (report *ReportTimeliness) BeforeCreate(tx *gorm.DB) (err error) {
	report.ID = uuid.New()
	return
}
