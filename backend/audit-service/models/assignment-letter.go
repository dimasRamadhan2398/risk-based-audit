package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LetterMember struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type AssignmentLetter struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LetterNumber       string              `gorm:"type:varchar(100);uniqueIndex;not null" json:"letterNumber"`
	Status             string              `gorm:"type:varchar(50);default:'Draft'" json:"status"`
	AuditTitle         string              `gorm:"type:varchar(255)" json:"auditTitle"`
	Leader             string              `gorm:"type:varchar(200)" json:"leader"`
	Category           string              `gorm:"type:varchar(100)" json:"category"`
	AuditYear          string              `gorm:"type:varchar(10)" json:"auditYear"`
	AuditTeam          string              `gorm:"type:varchar(100)" json:"auditTeam"`
	StartPeriod        string              `gorm:"type:varchar(100)" json:"startPeriod"`
	FinishPeriod       string              `gorm:"type:varchar(100)" json:"finishPeriod"`
	WorkingUnit        string              `gorm:"type:varchar(255)" json:"workingUnit"`
	ExecutionPeriod    string              `gorm:"type:varchar(255)" json:"executionPeriod"`
	AuditPurpose       string              `gorm:"type:text" json:"auditPurpose"`
	LetterDate         *time.Time          `json:"letterDate"`
	CAESignature       string              `gorm:"type:text" json:"caeSignature"`
	MembersList        []LetterMember      `gorm:"serializer:json" json:"membersList"`
	PurposeList        []string            `gorm:"serializer:json" json:"purposeList"`
	ScopeList          []string            `gorm:"serializer:json" json:"scopeList"`
	CcList             []string            `gorm:"serializer:json" json:"ccList"`
	ActionTakenReports []ActionTakenReport `gorm:"foreignKey:AssignmentLetterID" json:"actionTakenReports,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	DeletedAt          gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (AssignmentLetter) TableName() string {
	return "assignment_letters"
}

func (a *AssignmentLetter) UnmarshalJSON(data []byte) error {
	type Alias AssignmentLetter
	aux := struct {
		LetterDate *string `json:"letterDate"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.LetterDate != nil {
		dateStr := strings.TrimSpace(*aux.LetterDate)
		if dateStr == "" || dateStr == "null" {
			a.LetterDate = nil
		} else {
			formats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05Z07:00",
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				"2006-01-02",
				"02/01/2006",
			}
			var parsed bool
			for _, f := range formats {
				if t, err := time.Parse(f, dateStr); err == nil {
					a.LetterDate = &t
					parsed = true
					break
				}
			}
			if !parsed && len(dateStr) >= 10 {
				if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
					a.LetterDate = &t
					parsed = true
				}
			}
			if !parsed {
				a.LetterDate = nil
			}
		}
	} else {
		a.LetterDate = nil
	}

	return nil
}

