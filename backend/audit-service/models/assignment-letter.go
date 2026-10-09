package models

import (
	"encoding/json"
	"fmt"
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
	CompanyID          *uuid.UUID          `gorm:"type:uuid;index" json:"companyId,omitempty"`
	CompanyName        string              `gorm:"type:varchar(255)" json:"companyName,omitempty"`
	ExecutionPeriod    string              `gorm:"type:varchar(255)" json:"executionPeriod"`
	AuditPurpose       string              `gorm:"type:text" json:"auditPurpose"`
	LetterDate         *time.Time          `json:"letterDate"`
	CAESignature       string              `gorm:"type:text" json:"caeSignature"`
	MembersList        []LetterMember      `gorm:"serializer:json" json:"membersList"`
	PurposeList        []string            `gorm:"serializer:json" json:"purposeList"`
	ScopeList          []string            `gorm:"serializer:json" json:"scopeList"`
	CcList             []string            `gorm:"serializer:json" json:"ccList"`
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
		CompanyID  *string `json:"companyId"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.CompanyID != nil {
		cidStr := strings.TrimSpace(*aux.CompanyID)
		if cidStr == "" || cidStr == "null" {
			a.CompanyID = nil
		} else if parsedID, err := uuid.Parse(cidStr); err == nil {
			a.CompanyID = &parsedID
		} else {
			a.CompanyID = nil
		}
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

// JoinAuditPurposes joins the non-blank entries of a letter's purposeList
// into one line, the form used for a single "audit purpose" value.
func JoinAuditPurposes(purposes []string) string {
	parts := make([]string, 0, len(purposes))
	for _, p := range purposes {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "; ")
}

// EffectiveAuditPurpose is the letter's audit purpose: the auditPurpose
// column when set, otherwise its purposeList. The Assignment Letter form only
// edits purposeList ("Audit Purpose" in the UI), so letters created there
// store an empty audit_purpose.
func (a *AssignmentLetter) EffectiveAuditPurpose() string {
	if p := strings.TrimSpace(a.AuditPurpose); p != "" {
		return p
	}
	return JoinAuditPurposes(a.PurposeList)
}

// AfterFind fills an empty AuditPurpose from PurposeList, so every read
// (list, detail, Working Paper sync, reports) returns the letter's purpose.
// Nothing is written back.
func (a *AssignmentLetter) AfterFind(tx *gorm.DB) error {
	if strings.TrimSpace(a.AuditPurpose) == "" {
		a.AuditPurpose = JoinAuditPurposes(a.PurposeList)
	}
	return nil
}

// FindAssignmentLetter looks a letter up by letter number, or by id when ref
// is a UUID. Working papers and fieldwork store either form.
func FindAssignmentLetter(tx *gorm.DB, ref string) (*AssignmentLetter, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, false
	}
	var letter AssignmentLetter
	if err := tx.Where("letter_number = ?", ref).First(&letter).Error; err == nil {
		return &letter, true
	}
	if _, err := uuid.Parse(ref); err == nil {
		if err := tx.Where("id = ?", ref).First(&letter).Error; err == nil {
			return &letter, true
		}
	}
	return nil, false
}

func (a *AssignmentLetter) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(a.LetterNumber) == "" {
		year := strings.TrimSpace(a.AuditYear)
		if len(year) > 4 {
			year = year[:4]
		}
		if year == "" {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
		team := strings.TrimSpace(a.AuditTeam)
		if team == "" {
			team = "SKAI"
		}

		var letters []AssignmentLetter
		tx.Model(&AssignmentLetter{}).Unscoped().Select("letter_number").Find(&letters)

		maxSeq := 0
		prefix := "ST-"
		for _, l := range letters {
			numStr := strings.ToUpper(strings.TrimSpace(l.LetterNumber))
			if strings.HasPrefix(numStr, prefix) {
				parts := strings.Split(numStr, "/")
				if len(parts) > 0 {
					var seq int
					numPart := strings.TrimPrefix(parts[0], prefix)
					if _, err := fmt.Sscanf(numPart, "%d", &seq); err == nil && seq > maxSeq {
						maxSeq = seq
					}
				}
			}
		}
		a.LetterNumber = fmt.Sprintf("ST-%03d/%s/%s", maxSeq+1, strings.ToUpper(team), year)
	}
	return nil
}

