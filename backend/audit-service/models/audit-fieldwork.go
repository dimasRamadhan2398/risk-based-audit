package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldworkInterview struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID  string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	Interviewee         string         `gorm:"type:varchar(200)" json:"interviewee"`
	IntervieweePosition string         `gorm:"type:varchar(200)" json:"intervieweePosition"`
	Interviewer         string         `gorm:"type:varchar(200)" json:"interviewer"`
	InterviewerPosition string         `gorm:"type:varchar(200)" json:"interviewerPosition"`
	Date                string         `gorm:"type:varchar(100)" json:"date"`
	Topic               string         `gorm:"type:text" json:"topic"`
	FileName            string         `gorm:"type:varchar(255)" json:"fileName,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FieldworkInterview) TableName() string {
	return "fieldwork_interviews"
}

type FieldworkObservation struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	Activity           string         `gorm:"type:text" json:"activity"`
	Location           string         `gorm:"type:varchar(255)" json:"location"`
	Date               string         `gorm:"type:varchar(100)" json:"date"`
	Observer           string         `gorm:"type:varchar(200)" json:"observer"`
	FileName           string         `gorm:"type:varchar(255)" json:"fileName,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FieldworkObservation) TableName() string {
	return "fieldwork_observations"
}

type FieldworkDocument struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	DocumentName       string         `gorm:"type:varchar(255)" json:"documentName"`
	Description        string         `gorm:"type:text" json:"description"`
	RequiredDate       string         `gorm:"type:varchar(100)" json:"requiredDate"`
	FileName           string         `gorm:"type:varchar(255)" json:"fileName,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FieldworkDocument) TableName() string {
	return "fieldwork_documents"
}

type FieldworkSample struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	DocumentName       string         `gorm:"type:varchar(255)" json:"documentName"`
	DocumentNumber     string         `gorm:"type:varchar(100)" json:"documentNumber"`
	Date               string         `gorm:"type:varchar(100)" json:"date"`
	Description        string         `gorm:"type:text" json:"description"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FieldworkSample) TableName() string {
	return "fieldwork_samples"
}

type FieldworkTestControl struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID string         `gorm:"type:varchar(100);index" json:"assignmentLetterId"`
	ControlName        string         `gorm:"type:varchar(255)" json:"controlName"`
	ControlDescription string         `gorm:"type:text" json:"controlDescription"`
	ControlType        string         `gorm:"type:varchar(100)" json:"controlType"`
	TestProcedure      string         `gorm:"type:text" json:"testProcedure"`
	TestResult         string         `gorm:"type:varchar(100)" json:"testResult"`
	Finding            string         `gorm:"type:text" json:"finding"`
	Recommendation     string         `gorm:"type:text" json:"recommendation"`
	MitigationPlan     string         `gorm:"type:text" json:"mitigationPlan"`
	PIC                string         `gorm:"type:varchar(200)" json:"pic"`
	DueDate            string         `gorm:"type:varchar(100)" json:"dueDate"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FieldworkTestControl) TableName() string {
	return "fieldwork_test_controls"
}
