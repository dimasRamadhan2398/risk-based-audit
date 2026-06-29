package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamMember struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type WorkingPaperHeader struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	AuditPurpose       string         `gorm:"type:text" json:"auditPurpose"`
	BusinessProcess    string         `gorm:"type:varchar(255)" json:"businessProcess"`
	Period             string         `gorm:"type:varchar(100)" json:"period"`
	Location           string         `gorm:"type:varchar(255)" json:"location"`
	TeamMembers        []TeamMember   `gorm:"serializer:json" json:"teamMembers"`
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
	ID       int64  `json:"id"`
	Document string `json:"document"`
	L1       *bool  `json:"l1"`
	L2       *bool  `json:"l2"`
	L3       *bool  `json:"l3"`
}

type WorkingPaperSample struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkingPaperID string         `gorm:"type:varchar(100);index" json:"workingPaperId"`
	Population     *int           `gorm:"type:int" json:"population"`
	SampleSize     *int           `gorm:"type:int" json:"sampleSize"`
	Samples        []SampleDoc    `gorm:"serializer:json" json:"samples"`
	Conclusion     string         `gorm:"type:text" json:"conclusion"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
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