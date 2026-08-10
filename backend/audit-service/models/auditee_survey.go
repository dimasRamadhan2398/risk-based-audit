package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditeeSurvey represents feedback and satisfaction ratings collected from auditees
type AuditeeSurvey struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AuditExecutionID     *uuid.UUID     `gorm:"type:uuid;index" json:"audit_execution_id,omitempty"`
	AuditExecution       *AuditExecution `gorm:"foreignKey:AuditExecutionID" json:"audit_execution,omitempty"`
	AuditeeName          string         `gorm:"type:varchar(255);not null" json:"auditee_name"`
	Department           string         `gorm:"type:varchar(255)" json:"department"`
	Year                 int            `gorm:"type:int;not null" json:"year"`
	Month                int            `gorm:"type:int;not null" json:"month"` // 1-12
	RatingClarity        int            `gorm:"type:int;default:5" json:"rating_clarity"`        // 1-5
	RatingProfessionalism int           `gorm:"type:int;default:5" json:"rating_professionalism"` // 1-5
	RatingTimeliness     int            `gorm:"type:int;default:5" json:"rating_timeliness"`     // 1-5
	OverallScore         float64        `gorm:"type:decimal(3,2);not null" json:"overall_score"` // 1.00 - 5.00
	Comments             string         `gorm:"type:text" json:"comments"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditeeSurvey) TableName() string {
	return "auditee_surveys"
}

func (s *AuditeeSurvey) BeforeSave(tx *gorm.DB) (err error) {
	s.OverallScore = float64(s.RatingClarity+s.RatingProfessionalism+s.RatingTimeliness) / 3.0
	return
}

func (s *AuditeeSurvey) BeforeCreate(tx *gorm.DB) (err error) {
	if s.AuditExecutionID != nil {
		var count int64
		tx.Model(&AuditeeSurvey{}).Where("audit_execution_id = ?", s.AuditExecutionID).Count(&count)
		if count > 0 {
			return fmt.Errorf("a survey has already been submitted for this audit execution")
		}
	}
	return
}
