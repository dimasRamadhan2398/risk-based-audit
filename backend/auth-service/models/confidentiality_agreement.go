package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgreementType represents the type of confidentiality agreement
type AgreementType string

const (
	AgreementTypeSystem  AgreementType = "SYSTEM"
	AgreementTypeRole   AgreementType = "ROLE"
	AgreementTypeProject AgreementType = "PROJECT"
)

// ConfidentialityAgreement represents a user's acceptance of confidentiality agreement
type ConfidentialityAgreement struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"-"`
	AgreementType  AgreementType  `gorm:"type:varchar(20);not null" json:"agreement_type"`
	Title          string         `gorm:"type:varchar(200);not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Version        string         `gorm:"type:varchar(20);not null" json:"version"`
	IsAccepted     bool           `gorm:"default:false" json:"is_accepted"`
	AcceptedAt     *time.Time     `json:"accepted_at,omitempty"`
	IPAddress      string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent      string         `gorm:"type:varchar(255)" json:"user_agent"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name
func (ConfidentialityAgreement) TableName() string {
	return "confidentiality_agreements"
}

// CurrentSystemAgreementVersion is the SYSTEM agreement version users are asked
// to accept. The frontend posts it with POST /confidentiality/accept
// (frontend/stores/auth.ts, acceptConfidentiality); keep the two in step when
// the integrity pact text changes. Only an acceptance of this version counts in
// GET /confidentiality/acceptance-summary.
const CurrentSystemAgreementVersion = "1.0"

// MaxAcceptanceSummaryCodes caps the employee_codes of one acceptance-summary
// request.
const MaxAcceptanceSummaryCodes = 500

// MaxEmployeeCodeLength is the width of users.employee_id.
const MaxEmployeeCodeLength = 50

// ConfidentialityAcceptanceItem is one row of GET
// /confidentiality/acceptance-summary: whether the employee code has an
// active login account and whether that account holds an accepted, unexpired
// SYSTEM agreement at CurrentSystemAgreementVersion. No agreement content,
// IP address, user agent or account details are returned.
type ConfidentialityAcceptanceItem struct {
	EmployeeCode string     `json:"employee_code"`
	HasUser      bool       `json:"has_user"`
	Accepted     bool       `json:"accepted"`
	AcceptedAt   *time.Time `json:"accepted_at"`
	Version      *string    `json:"version"`
}
