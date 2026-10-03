package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamMember struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type ActivityItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type WorkingPaperHeader struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	AuditPurpose       string         `gorm:"type:text" json:"auditPurpose"`
	BusinessProcess    string         `gorm:"type:varchar(255)" json:"businessProcess"`
	Period             string         `gorm:"type:varchar(100)" json:"period"`
	Location           string         `gorm:"type:varchar(255)" json:"location"`
	TeamMembers        []TeamMember   `gorm:"serializer:json" json:"teamMembers"`
	Activities         []ActivityItem `gorm:"serializer:json" json:"activities"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkingPaperHeader) TableName() string {
	return "working_paper_headers"
}

type WorkingPaperRisk struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkingPaperID     string         `gorm:"type:varchar(100);index" json:"workingPaperId"` // Link back if needed, or assignment letter ID
	Risk               string         `gorm:"type:text" json:"risk"`
	Taxonomy           string         `gorm:"type:varchar(100)" json:"taxonomy"`
	RiskLevel          string         `gorm:"type:varchar(50)" json:"riskLevel"`
	ControlDescription string         `gorm:"type:text" json:"controlDescription"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkingPaperRisk) TableName() string {
	return "working_paper_risks"
}

type SampleDoc struct {
	ID                int64  `json:"id"`
	FieldworkDocument string `json:"fieldworkDocument,omitempty"`
	Document          string `json:"document"`
	Step1             string `json:"step1,omitempty"`
	Step2             string `json:"step2,omitempty"`
	Step3             string `json:"step3,omitempty"`
	L1                any    `json:"l1"`
	L2                any    `json:"l2"`
	L3                any    `json:"l3"`
}

// UnmarshalJSON handles flexible parsing of L1, L2, L3 which can be boolean, string ("Pass"/"Fail"), or nil
func (s *SampleDoc) UnmarshalJSON(data []byte) error {
	type Alias SampleDoc
	aux := &struct {
		L1 any `json:"l1"`
		L2 any `json:"l2"`
		L3 any `json:"l3"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	parseResult := func(val any) any {
		if val == nil {
			return nil
		}
		switch v := val.(type) {
		case bool:
			return v
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			if lower == "pass" || lower == "true" {
				return true
			}
			if lower == "fail" || lower == "false" {
				return false
			}
			return v
		default:
			return v
		}
	}

	s.L1 = parseResult(aux.L1)
	s.L2 = parseResult(aux.L2)
	s.L3 = parseResult(aux.L3)
	return nil
}

type WorkingPaperSample struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkingPaperID string         `gorm:"type:varchar(100);index" json:"workingPaperId"`
	Population     *string        `gorm:"type:varchar(255)" json:"population"`
	SampleSize     *int           `gorm:"type:int" json:"sampleSize"`
	Samples        []SampleDoc    `gorm:"serializer:json" json:"samples"`
	Conclusion     string         `gorm:"type:text" json:"conclusion"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *WorkingPaperSample) UnmarshalJSON(data []byte) error {
	type Alias WorkingPaperSample
	aux := &struct {
		Population interface{} `json:"population"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Population != nil {
		switch v := aux.Population.(type) {
		case string:
			w.Population = &v
		case float64:
			str := fmt.Sprintf("%.0f", v)
			w.Population = &str
		case int:
			str := fmt.Sprintf("%d", v)
			w.Population = &str
		default:
			str := fmt.Sprintf("%v", v)
			w.Population = &str
		}
	} else {
		w.Population = nil
	}
	return nil
}

func (WorkingPaperSample) TableName() string {
	return "working_paper_samples"
}

type RootCauseItem struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	W1     string `json:"w1"`
	W2     string `json:"w2"`
	W3     string `json:"w3"`
}

type WorkingPaperCause struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkingPaperID string          `gorm:"type:varchar(100);index" json:"workingPaperId"`
	Condition      string          `gorm:"type:text" json:"condition"`
	Criteria       string          `gorm:"type:text" json:"criteria"`
	Impact         string          `gorm:"type:text" json:"impact"`
	RootCause      []RootCauseItem `gorm:"serializer:json" json:"rootCause"`
	EvidenceFile   string          `gorm:"type:varchar(255)" json:"evidenceFile"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (WorkingPaperCause) TableName() string {
	return "working_paper_causes"
}

type WorkingPaperPlan struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkingPaperID    string         `gorm:"type:varchar(100);index" json:"workingPaperId"`
	Recommendation    string         `gorm:"type:text" json:"recommendation"`
	Response          string         `gorm:"type:text" json:"response"`
	ActionDescription string         `gorm:"type:text" json:"actionDescription"`
	PIC               string         `gorm:"type:varchar(200)" json:"pic"`
	PeriodAction      string         `gorm:"type:varchar(100)" json:"periodAction"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkingPaperPlan) TableName() string {
	return "working_paper_plans"
}

// Keep original WorkingPaper struct as well
type WorkingPaper struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID uuid.UUID      `gorm:"type:uuid;not null;index" json:"activity_plan_id"`
	ActivityPlan   AuditActivity  `gorm:"foreignKey:ActivityPlanID" json:"activity_plan"`
	PreparedByID   uuid.UUID      `gorm:"type:uuid;not null" json:"prepared_by_id"`
	ReviewedByID   *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by_id"`
	PaperCode      string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"paper_code"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Methodology    string         `gorm:"type:text" json:"methodology"`
	TestResults    string         `gorm:"type:text" json:"test_results"`
	Conclusion     string         `gorm:"type:text" json:"conclusion"`
	RiskID         *uuid.UUID     `gorm:"type:uuid" json:"risk_id"`
	ControlID      *uuid.UUID     `gorm:"type:uuid" json:"control_id"`
	Status         string         `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}