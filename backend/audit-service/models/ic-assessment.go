package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// COSO 2013 internal control component assessment (docs/coso-2013-component-
// assessment-plan.md section 4.2). One assessment per master department,
// optional branch (location), year and optional quarter. Principle and
// component codes are business keys validated against the constant catalogue
// in services/ic_assessment (it mirrors the risk-service seed); there is no
// cross-service foreign key. Every *_by is a JWT user id plus display name,
// because users live in auth-service.
//
// All tables are new; "audit migrate up" creates them (additive only).

// ICAssessment is the header of one assessment
type ICAssessment struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// Scope. department_* and company_id are a snapshot taken at create and
	// never rewritten. location_key is the location id or "" so the scope
	// unique index needs no expression.
	DepartmentID   uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:uq_ic_assessments_scope,priority:1,where:deleted_at IS NULL" json:"department_id"`
	DepartmentCode string     `gorm:"type:varchar(50)" json:"department_code"`
	DepartmentName string     `gorm:"type:varchar(255)" json:"department_name"`
	CompanyID      *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	LocationID     *uuid.UUID `gorm:"type:uuid" json:"location_id"`
	LocationKey    string     `gorm:"type:varchar(36);not null;default:'';uniqueIndex:uq_ic_assessments_scope,priority:2,where:deleted_at IS NULL" json:"-"`
	LocationName   string     `gorm:"type:varchar(255)" json:"location_name"`
	Year           int        `gorm:"not null;index;uniqueIndex:uq_ic_assessments_scope,priority:3,where:deleted_at IS NULL" json:"year"`
	Quarter        int        `gorm:"type:smallint;not null;default:0;check:chk_ic_assessments_quarter,quarter BETWEEN 0 AND 4;uniqueIndex:uq_ic_assessments_scope,priority:4,where:deleted_at IS NULL" json:"quarter"`
	PeriodStart    time.Time  `gorm:"type:date" json:"period_start"`
	PeriodEnd      time.Time  `gorm:"type:date" json:"period_end"`
	AsOfDate       *time.Time `gorm:"type:date" json:"as_of_date"`

	Status      string `gorm:"type:varchar(20);not null;index" json:"status"`
	Version     int    `gorm:"not null;default:1" json:"version"`
	ProposalRun int    `gorm:"not null;default:0" json:"proposal_run"`
	// EvidenceHash is the hash of the scored signal inputs of the latest
	// proposal run ("data changed since approval" compares against it)
	EvidenceHash string `gorm:"type:varchar(64)" json:"evidence_hash"`

	DerivedConclusion       string `gorm:"type:varchar(20)" json:"derived_conclusion"`
	FinalConclusion         string `gorm:"type:varchar(20)" json:"final_conclusion"`
	OperateTogether         *bool  `json:"operate_together"`
	OperateTogetherNote     string `gorm:"type:text" json:"operate_together_note"`
	ConclusionText          string `gorm:"type:text" json:"conclusion_text"`
	ConclusionJustification string `gorm:"type:text" json:"conclusion_justification"`

	CreatedByUserID   string     `gorm:"type:varchar(64)" json:"created_by_user_id"`
	CreatedByName     string     `gorm:"type:varchar(200)" json:"created_by_name"`
	SubmittedByUserID string     `gorm:"type:varchar(64)" json:"submitted_by_user_id"`
	SubmittedByName   string     `gorm:"type:varchar(200)" json:"submitted_by_name"`
	SubmittedAt       *time.Time `json:"submitted_at"`
	ApprovedByUserID  string     `gorm:"type:varchar(64)" json:"approved_by_user_id"`
	ApprovedByName    string     `gorm:"type:varchar(200)" json:"approved_by_name"`
	ApprovedByRole    string     `gorm:"type:varchar(100)" json:"approved_by_role"`
	ApprovedAt        *time.Time `json:"approved_at"`
	ApprovalNote      string     `gorm:"type:text" json:"approval_note"`

	// Optional approval signature: a PNG/JPEG data URL (same validation and
	// storage as the executive summary approval). Never returned in JSON; it
	// is served only by GET /ic-assessments/:id/signature/view.
	SignatureImage string `gorm:"type:text" json:"-"`
	SignerName     string `gorm:"type:varchar(200)" json:"signer_name"`
	SignerRole     string `gorm:"type:varchar(100)" json:"signer_role"`

	ReopenCount      int    `gorm:"not null;default:0" json:"reopen_count"`
	LastReopenReason string `gorm:"type:text" json:"last_reopen_reason"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Components []ICAssessmentComponent `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE" json:"components,omitempty"`
	Principles []ICAssessmentPrinciple `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE" json:"principles,omitempty"`
}

func (ICAssessment) TableName() string { return "ic_assessments" }

func (a *ICAssessment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// ICAssessmentComponent is one of the 5 component rows of an assessment
type ICAssessmentComponent struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_ic_assessment_components,priority:1" json:"assessment_id"`
	ComponentCode string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_ic_assessment_components,priority:2" json:"component_code"`

	ProposedScorePct        *float64 `gorm:"type:decimal(5,2)" json:"proposed_score_pct"`
	ProposedRating          *int     `gorm:"type:smallint" json:"proposed_rating"`
	ProposedStatus          string   `gorm:"type:varchar(20)" json:"proposed_status"`
	ProposedSeverity        string   `gorm:"type:varchar(20)" json:"proposed_severity"`
	ProposalDataCoveragePct *float64 `gorm:"type:decimal(5,2)" json:"proposal_data_coverage_pct"`

	FinalRating   *int   `gorm:"type:smallint" json:"final_rating"`
	FinalStatus   string `gorm:"type:varchar(20)" json:"final_status"`
	FinalSeverity string `gorm:"type:varchar(20)" json:"final_severity"`

	DerivedStatus   string `gorm:"type:varchar(20)" json:"derived_status"`
	DerivedSeverity string `gorm:"type:varchar(20)" json:"derived_severity"`

	Conclusion     string `gorm:"type:text" json:"conclusion"`
	OverrideReason string `gorm:"type:text" json:"override_reason"`
	IsOverride     bool   `gorm:"not null;default:false" json:"is_override"`

	ReviewedByUserID string     `gorm:"type:varchar(64)" json:"reviewed_by_user_id"`
	ReviewedByName   string     `gorm:"type:varchar(200)" json:"reviewed_by_name"`
	ReviewedAt       *time.Time `json:"reviewed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ICAssessmentComponent) TableName() string { return "ic_assessment_components" }

func (c *ICAssessmentComponent) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ICAssessmentPrinciple is one of the 17 principle rows of an assessment
type ICAssessmentPrinciple struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_ic_assessment_principles,priority:1" json:"assessment_id"`
	ComponentCode       string    `gorm:"type:varchar(10);not null" json:"component_code"`
	PrincipleCode       string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_ic_assessment_principles,priority:2" json:"principle_code"`
	TaggedControlCount  int       `gorm:"not null;default:0" json:"tagged_control_count"`
	TaggedControlAvgPct *float64  `gorm:"type:decimal(5,2)" json:"tagged_control_avg_pct"`
	Status              string    `gorm:"type:varchar(20);not null;default:'NOT_ASSESSED'" json:"status"`
	Severity            string    `gorm:"type:varchar(20);not null;default:'NONE'" json:"severity"`
	Rationale           string    `gorm:"type:text" json:"rationale"`
	UpdatedByUserID     string    `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName       string    `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (ICAssessmentPrinciple) TableName() string { return "ic_assessment_principles" }

func (p *ICAssessmentPrinciple) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ICAssessmentEvidence is one signal of one proposal run. Append-only: the
// current evidence is the rows of the assessment's latest proposal_run.
type ICAssessmentEvidence struct {
	ID            uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID  uuid.UUID              `gorm:"type:uuid;not null;index:idx_ic_assessment_evidence_run,priority:1" json:"assessment_id"`
	ProposalRun   int                    `gorm:"not null;index:idx_ic_assessment_evidence_run,priority:2" json:"proposal_run"`
	ComponentCode string                 `gorm:"type:varchar(10);not null" json:"component_code"`
	SignalKey     string                 `gorm:"type:varchar(64);not null" json:"signal_key"`
	PrincipleHint string                 `gorm:"type:varchar(50)" json:"principle_hint"`
	SourceService string                 `gorm:"type:varchar(30)" json:"source_service"`
	SourceRef     string                 `gorm:"type:varchar(255)" json:"source_ref"`
	State         string                 `gorm:"type:varchar(20);not null" json:"state"`
	ValuePct      *float64               `gorm:"type:decimal(5,2)" json:"value_pct"`
	Numerator     *int                   `json:"numerator"`
	Denominator   *int                   `json:"denominator"`
	Weight        float64                `gorm:"type:decimal(6,2);not null;default:0" json:"weight"`
	ScopeNote     string                 `gorm:"type:text" json:"scope_note"`
	Detail        map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"detail"`
	CollectedAt   time.Time              `json:"collected_at"`
}

func (ICAssessmentEvidence) TableName() string { return "ic_assessment_evidence" }

func (e *ICAssessmentEvidence) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// ICAssessmentEvent is the append-only audit trail. AssessmentID is NULL for
// period lock/unlock and settings changes. There is no update or delete route.
type ICAssessmentEvent struct {
	ID           uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID *uuid.UUID             `gorm:"type:uuid;index" json:"assessment_id"`
	Action       string                 `gorm:"type:varchar(30);not null" json:"action"`
	FromStatus   string                 `gorm:"type:varchar(20)" json:"from_status"`
	ToStatus     string                 `gorm:"type:varchar(20)" json:"to_status"`
	ActorUserID  string                 `gorm:"type:varchar(64)" json:"actor_user_id"`
	ActorName    string                 `gorm:"type:varchar(200)" json:"actor_name"`
	ActorRoles   string                 `gorm:"type:varchar(255)" json:"actor_roles"`
	Reason       string                 `gorm:"type:text" json:"reason"`
	Before       map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"before"`
	After        map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"after"`
	IP           string                 `gorm:"type:varchar(64)" json:"ip"`
	UserAgent    string                 `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt    time.Time              `gorm:"index" json:"created_at"`
}

func (ICAssessmentEvent) TableName() string { return "ic_assessment_events" }

func (e *ICAssessmentEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// ICAssessmentPeriod is a period lock. No row means open. quarter 0 is the
// whole year; locking the year also locks its quarters.
type ICAssessmentPeriod struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Year           int        `gorm:"not null;uniqueIndex:uq_ic_assessment_periods,priority:1" json:"year"`
	Quarter        int        `gorm:"type:smallint;not null;default:0;uniqueIndex:uq_ic_assessment_periods,priority:2" json:"quarter"`
	Status         string     `gorm:"type:varchar(10);not null" json:"status"`
	LockedByUserID string     `gorm:"type:varchar(64)" json:"locked_by_user_id"`
	LockedByName   string     `gorm:"type:varchar(200)" json:"locked_by_name"`
	LockedAt       *time.Time `json:"locked_at"`
	LockNote       string     `gorm:"type:text" json:"lock_note"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ICAssessmentPeriod) TableName() string { return "ic_assessment_periods" }

func (p *ICAssessmentPeriod) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ICAssessmentSettingsID is the id of the single settings row of a stack
// (each tenant has its own audit database, so one row is one tenant)
const ICAssessmentSettingsID = 1

// ICAssessmentSettings holds the tenant's proposal thresholds. "audit migrate
// up" inserts the row once (from IC_LHA_DUE_DAYS / IC_MGMT_RESPONSE_DUE_DAYS
// when set, else the defaults) and never overwrites it; ADMIN changes it
// through PUT /ic-assessments/settings.
type ICAssessmentSettings struct {
	ID                        int       `gorm:"primaryKey;autoIncrement:false" json:"-"`
	LHADueDays                int       `gorm:"column:lha_due_days;not null" json:"lha_due_days"`
	ManagementResponseDueDays int       `gorm:"column:management_response_due_days;not null" json:"management_response_due_days"`
	UpdatedByUserID           string    `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName             string    `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

func (ICAssessmentSettings) TableName() string { return "ic_assessment_settings" }

// ICAssessmentModels lists the models "audit migrate up" creates, in order
func ICAssessmentModels() []interface{} {
	return []interface{}{
		&ICAssessment{},
		&ICAssessmentComponent{},
		&ICAssessmentPrinciple{},
		&ICAssessmentEvidence{},
		&ICAssessmentEvent{},
		&ICAssessmentPeriod{},
		&ICAssessmentSettings{},
	}
}
