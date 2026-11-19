package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QUANTITATIVE_POF method of the COSO 2013 assessment: weighted points of
// focus (PoF) with fulfilment elements (docs/coso-pof-assessment-api.md).
//
// All tables are new and additive; "audit migrate up" creates them. The
// catalogue (names, weights, default owners) is copied into the assessment when
// it is created, so an assessment never changes when the catalogue is revised
// and needs no cross-service call to be read or scored.

// Assessment methods
const (
	ICMethodQualitative     = "QUALITATIVE"
	ICMethodQuantitativePoF = "QUANTITATIVE_POF"
)

// ICBranchMethod is the method setting of one branch. LocationKey is the
// location id, or "" for the no-branch scope (which also acts as the tenant
// default). No row means QUALITATIVE.
type ICBranchMethod struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LocationKey     string     `gorm:"type:varchar(36);not null;default:'';uniqueIndex:uq_ic_branch_methods_location" json:"-"`
	LocationID      *uuid.UUID `gorm:"type:uuid" json:"location_id"`
	LocationName    string     `gorm:"type:varchar(255)" json:"location_name"`
	Method          string     `gorm:"type:varchar(20);not null" json:"method"`
	UpdatedByUserID string     `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName   string     `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ICBranchMethod) TableName() string { return "ic_branch_methods" }

func (m *ICBranchMethod) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICPofOwnerOverride replaces the catalogue's default owner labels of one
// principle for this tenant. OwnerRoles is a comma separated label list.
type ICPofOwnerOverride struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PrincipleCode   string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_ic_pof_owner_overrides_principle" json:"principle_code"`
	OwnerRoles      string    `gorm:"type:varchar(200);not null" json:"-"`
	UpdatedByUserID string    `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName   string    `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ICPofOwnerOverride) TableName() string { return "ic_pof_owner_overrides" }

func (m *ICPofOwnerOverride) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICAssessmentPoF is one point of focus of one assessment (a snapshot of the
// catalogue row plus the PoF-level answer).
type ICAssessmentPoF struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_ic_assessment_pofs,priority:1" json:"assessment_id"`
	PofCode       string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_ic_assessment_pofs,priority:2" json:"pof_code"`
	PrincipleCode string    `gorm:"type:varchar(10);not null" json:"principle_code"`
	ComponentCode string    `gorm:"type:varchar(10);not null" json:"component_code"`
	SortOrder     int       `gorm:"type:int" json:"sort_order"`
	NameEn        string    `gorm:"type:varchar(255)" json:"name_en"`
	NameID        string    `gorm:"type:varchar(255)" json:"name_id"`
	Weight        float64   `gorm:"type:numeric(9,4);not null" json:"weight"`
	// DefaultOwnerRole: comma separated labels (catalogue default, or the
	// tenant override at creation time). Information only.
	DefaultOwnerRole string `gorm:"type:varchar(200)" json:"default_owner_role"`

	// Achievement is the PoF-level value (0..1). It counts only while no
	// element of the PoF has a value.
	Achievement    *float64 `gorm:"type:numeric(5,4)" json:"achievement"`
	Explanation    string   `gorm:"type:text" json:"explanation"`
	Recommendation string   `gorm:"type:text" json:"recommendation"`
	Methods        string   `gorm:"type:varchar(16)" json:"methods"`

	UpdatedByUserID string    `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName   string    `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ICAssessmentPoF) TableName() string { return "ic_assessment_pofs" }

func (m *ICAssessmentPoF) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICAssessmentElement is one fulfilment element of one assessment.
type ICAssessmentElement struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_ic_assessment_elements,priority:1;index:idx_ic_assessment_elements_pof,priority:1" json:"assessment_id"`
	ElementCode    string    `gorm:"type:varchar(30);not null;uniqueIndex:uq_ic_assessment_elements,priority:2" json:"element_code"`
	PofCode        string    `gorm:"type:varchar(20);not null;index:idx_ic_assessment_elements_pof,priority:2" json:"pof_code"`
	SortOrder      int       `gorm:"type:int" json:"sort_order"`
	Type           string    `gorm:"type:varchar(10);not null" json:"type"`
	DescriptionEn  string    `gorm:"type:text" json:"description_en"`
	DescriptionID  string    `gorm:"type:text" json:"description_id"`
	Weight         float64   `gorm:"type:numeric(9,4);not null" json:"weight"`
	DefaultMethods string    `gorm:"type:varchar(16)" json:"default_methods"`
	ScaleType      string    `gorm:"type:varchar(14);not null;default:'DISCRETE'" json:"scale_type"`
	// RequiresIndependentAssessor: the CAE may not fill this element
	RequiresIndependentAssessor bool `gorm:"not null;default:false" json:"requires_independent_assessor"`

	Achievement    *float64 `gorm:"type:numeric(5,4)" json:"achievement"`
	Methods        string   `gorm:"type:varchar(16)" json:"methods"`
	Explanation    string   `gorm:"type:text" json:"explanation"`
	Recommendation string   `gorm:"type:text" json:"recommendation"`

	UpdatedByUserID string    `gorm:"type:varchar(64)" json:"updated_by_user_id"`
	UpdatedByName   string    `gorm:"type:varchar(200)" json:"updated_by_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ICAssessmentElement) TableName() string { return "ic_assessment_elements" }

func (m *ICAssessmentElement) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Element types (mirrors risk-service coso_pof_elements.type)
const (
	CosoElementPresentName  = "PRESENT"
	CosoElementFunctionName = "FUNCTION"
)

// Scope types of assignments and evidence
const (
	ICScopePrinciple = "PRINCIPLE"
	ICScopePoF       = "POF"
	ICScopeElement   = "ELEMENT"
)

// ICAssessmentAssignment assigns a person (PIC) to a principle or a PoF. The
// person is an auth-service user id; RoleLabel is a free label or a role name
// that already exists in auth-service, never a new role.
type ICAssessmentAssignment struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_ic_assessment_assignments,priority:1" json:"assessment_id"`
	ScopeType        string     `gorm:"type:varchar(10);not null;uniqueIndex:uq_ic_assessment_assignments,priority:2" json:"scope_type"`
	ScopeCode        string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_ic_assessment_assignments,priority:3" json:"scope_code"`
	AssigneeUserID   string     `gorm:"type:varchar(64);not null;uniqueIndex:uq_ic_assessment_assignments,priority:4;index" json:"assignee_user_id"`
	AssigneeName     string     `gorm:"type:varchar(200)" json:"assignee_name"`
	RoleLabel        string     `gorm:"type:varchar(100)" json:"role_label"`
	DueDate          *time.Time `gorm:"type:date" json:"due_date"`
	Note             string     `gorm:"type:text" json:"note"`
	AssignedByUserID string     `gorm:"type:varchar(64)" json:"assigned_by_user_id"`
	AssignedByName   string     `gorm:"type:varchar(200)" json:"assigned_by_name"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (ICAssessmentAssignment) TableName() string { return "ic_assessment_assignments" }

func (m *ICAssessmentAssignment) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICPoFEvidence is a file or a reference attached to a PoF or an element. The
// file bytes live in the database (Content, never serialised) so evidence
// survives container recreation.
type ICPoFEvidence struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID uuid.UUID  `gorm:"type:uuid;not null;index:idx_ic_pof_evidence_scope,priority:1" json:"assessment_id"`
	ScopeType    string     `gorm:"type:varchar(10);not null" json:"scope_type"` // POF | ELEMENT | ASSESSMENT
	ScopeCode    string     `gorm:"type:varchar(30);index:idx_ic_pof_evidence_scope,priority:2" json:"scope_code"`
	Kind         string     `gorm:"type:varchar(10);not null" json:"kind"` // FILE | REFERENCE
	Title        string     `gorm:"type:varchar(255)" json:"title"`
	DocNumber    string     `gorm:"type:varchar(100)" json:"doc_number"`
	Note         string     `gorm:"type:text" json:"note"`
	FileName     string     `gorm:"type:varchar(255)" json:"file_name"`
	ContentType  string     `gorm:"type:varchar(100)" json:"content_type"`
	SizeBytes    int64      `json:"size_bytes"`
	Sha256       string     `gorm:"type:varchar(64)" json:"sha256"`
	Content      []byte     `gorm:"type:bytea" json:"-"`
	Source       string     `gorm:"type:varchar(10);not null;default:'MANUAL'" json:"source"` // MANUAL | IMPORT
	ImportID     *uuid.UUID `gorm:"type:uuid" json:"import_id"`

	UploadedByUserID string         `gorm:"type:varchar(64)" json:"uploaded_by_user_id"`
	UploadedByName   string         `gorm:"type:varchar(200)" json:"uploaded_by_name"`
	CreatedAt        time.Time      `json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ICPoFEvidence) TableName() string { return "ic_pof_evidence" }

func (m *ICPoFEvidence) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICPoFImportProfile is a saved column mapping for importing external working
// papers: the software "learns" a format once and reuses it. Mapping is the
// JSON of the import mapping (see services/ic_assessment).
type ICPoFImportProfile struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name            string         `gorm:"type:varchar(120);not null" json:"name"`
	NameKey         string         `gorm:"type:varchar(120);not null;uniqueIndex:uq_ic_pof_import_profiles_name,where:deleted_at IS NULL" json:"-"`
	Description     string         `gorm:"type:text" json:"description"`
	Format          string         `gorm:"type:varchar(12);not null;default:'EXTERNAL'" json:"format"`
	Mapping         string         `gorm:"type:text;not null" json:"-"`
	UseCount        int            `gorm:"not null;default:0" json:"use_count"`
	LastUsedAt      *time.Time     `json:"last_used_at"`
	CreatedByUserID string         `gorm:"type:varchar(64)" json:"created_by_user_id"`
	CreatedByName   string         `gorm:"type:varchar(200)" json:"created_by_name"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ICPoFImportProfile) TableName() string { return "ic_pof_import_profiles" }

func (m *ICPoFImportProfile) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Import statuses
const (
	ICImportPreviewed = "PREVIEWED"
	ICImportApplied   = "APPLIED"
)

// ICPoFImport is one uploaded workbook: the original file, the mapping used
// and the preview. The file stays as the evidence of what was imported.
type ICPoFImport struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"assessment_id"`
	Status       string     `gorm:"type:varchar(12);not null" json:"status"`
	Format       string     `gorm:"type:varchar(12)" json:"format"`
	ProfileID    *uuid.UUID `gorm:"type:uuid" json:"profile_id"`
	FileName     string     `gorm:"type:varchar(255)" json:"file_name"`
	SizeBytes    int64      `json:"size_bytes"`
	Sha256       string     `gorm:"type:varchar(64)" json:"sha256"`
	Content      []byte     `gorm:"type:bytea" json:"-"`
	Mapping      string     `gorm:"type:text" json:"-"`
	Preview      string     `gorm:"type:text" json:"-"`
	Sheet        string     `gorm:"type:varchar(100)" json:"sheet"`

	CreatedByUserID string     `gorm:"type:varchar(64)" json:"created_by_user_id"`
	CreatedByName   string     `gorm:"type:varchar(200)" json:"created_by_name"`
	AppliedByName   string     `gorm:"type:varchar(200)" json:"applied_by_name"`
	AppliedAt       *time.Time `json:"applied_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (ICPoFImport) TableName() string { return "ic_pof_imports" }

func (m *ICPoFImport) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ICPoFModels lists the PoF tables "audit migrate up" creates.
func ICPoFModels() []interface{} {
	return []interface{}{
		&ICBranchMethod{},
		&ICPofOwnerOverride{},
		&ICAssessmentPoF{},
		&ICAssessmentElement{},
		&ICAssessmentAssignment{},
		&ICPoFEvidence{},
		&ICPoFImportProfile{},
		&ICPoFImport{},
	}
}
