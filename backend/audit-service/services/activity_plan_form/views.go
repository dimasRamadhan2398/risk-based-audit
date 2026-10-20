package activityplanform

import (
	"time"

	"audit-service/models"
)

// Response DTOs of contract sections 6.1-6.3. Field names are camelCase.

// AnnualPlanRef is the PKAT summary shown with an RKA.
type AnnualPlanRef struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Status   string `json:"status"`
	IsActive bool   `json:"isActive"`
}

// AnnualPlanSummaryRef is AnnualPlanRef without the year (list items).
type AnnualPlanSummaryRef struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	IsActive bool   `json:"isActive"`
}

// Totals sums the rows of an RKA.
type Totals struct {
	RowCount     int `json:"rowCount"`
	TeamSize     int `json:"teamSize"`
	TotalMandays int `json:"totalMandays"`
}

// UniverseRef is the Audit Universe of a PKAT row (name and risk level are the
// PKAT snapshots).
type UniverseRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RiskLevel string `json:"riskLevel"`
}

// DepartmentRef is the auditee department of a PKAT row.
type DepartmentRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TeamLeaderRef is the Ketua Tim of an RKA row.
type TeamLeaderRef struct {
	EmployeeID string `json:"employeeId"`
	Name       string `json:"name"`
}

// ListItem is one entry of GET /activity-plans.
type ListItem struct {
	ID              string                `json:"id"`
	FormVersion     int                   `json:"formVersion"`
	Year            int                   `json:"year"`
	ActivityType    string                `json:"activityType"`
	Title           string                `json:"title"`
	AnnualPlan      *AnnualPlanSummaryRef `json:"annualPlan"`
	Totals          Totals                `json:"totals"`
	Status          string                `json:"status"`
	AttachmentCount int                   `json:"attachmentCount"`
	CreatedBy       string                `json:"createdBy"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

// Row is one RKA row in GET /activity-plans/:id.
type Row struct {
	ID                     string         `json:"id"`
	No                     int            `json:"no"`
	AnnualActivityID       string         `json:"annualActivityId"`
	AssignmentCode         string         `json:"assignmentCode"`
	AssignmentCodeEditable bool           `json:"assignmentCodeEditable"`
	AuditObject            string         `json:"auditObject"`
	AuditCategory          string         `json:"auditCategory"`
	AuditKind              string         `json:"auditKind"`
	AuditUniverse          UniverseRef    `json:"auditUniverse"`
	Department             DepartmentRef  `json:"department"`
	Justification          []string       `json:"justification"`
	Objectives             []string       `json:"objectives"`
	Focus                  []string       `json:"focus"`
	Schedule               string         `json:"schedule"`
	TeamSize               int            `json:"teamSize"`
	TotalMandays           int            `json:"totalMandays"`
	TeamLeader             *TeamLeaderRef `json:"teamLeader"`
	PkatSupervisorName     string         `json:"pkatSupervisorName"`
	Status                 string         `json:"status"`
	InUse                  bool           `json:"inUse"`
	// Source is "pkat" when the PKAT-owned fields are live, "snapshot" when the
	// PKAT or its row no longer exists and the saved snapshot is shown.
	Source string `json:"source"`
}

// Legacy is the stored JSON of a form-version-1 plan, as-is.
type Legacy struct {
	PlanTitle         string                   `json:"planTitle"`
	PlanYear          string                   `json:"planYear"`
	PlanPeriodStart   string                   `json:"planPeriodStart"`
	PlanPeriodEnd     string                   `json:"planPeriodEnd"`
	Department        string                   `json:"department"`
	CreationDate      string                   `json:"creationDate"`
	PlannedActivities []models.PlannedActivity `json:"plannedActivities"`
	ResourceAuditors  []models.ResourceAuditor `json:"resourceAuditors"`
	Budget            models.PlanBudget        `json:"budget"`
	Review            models.PlanReview        `json:"review"`
}

// Detail is GET /activity-plans/:id and the POST/PUT response.
type Detail struct {
	ID                   string                         `json:"id"`
	FormVersion          int                            `json:"formVersion"`
	AnnualPlanID         *string                        `json:"annualPlanId"`
	AnnualPlan           *AnnualPlanRef                 `json:"annualPlan"`
	Year                 int                            `json:"year"`
	ActivityType         string                         `json:"activityType"`
	Title                string                         `json:"title"`
	Status               string                         `json:"status"`
	Rows                 []Row                          `json:"rows"`
	Totals               Totals                         `json:"totals"`
	AttachmentCategory   string                         `json:"attachmentCategory"`
	Attachments          []models.AnnualAuditAttachment `json:"attachments"`
	AttachmentUploadedBy string                         `json:"attachmentUploadedBy"`
	AttachmentUploadDate string                         `json:"attachmentUploadDate"`
	CreatedBy            string                         `json:"createdBy"`
	CreatedAt            time.Time                      `json:"createdAt"`
	UpdatedAt            time.Time                      `json:"updatedAt"`
	Legacy               *Legacy                        `json:"legacy"`
}

// UsedByRef names the RKA that holds a PKAT row.
type UsedByRef struct {
	ActivityPlanID    string `json:"activityPlanId"`
	ActivityPlanTitle string `json:"activityPlanTitle"`
	AssignmentCode    string `json:"assignmentCode"`
}

// Prefill holds the values applied once when a PKAT row is added to an RKA.
type Prefill struct {
	AssignmentCode string   `json:"assignmentCode"`
	Justification  []string `json:"justification"`
	Objectives     []string `json:"objectives"`
	Focus          []string `json:"focus"`
}

// ExistingRowRef is an audit_activities row already linked to the PKAT row; the
// RKA adopts it on save.
type ExistingRowRef struct {
	ID             string `json:"id"`
	AssignmentCode string `json:"assignmentCode"`
	Status         string `json:"status"`
}

// Option is one PKAT row in GET /activity-plans/annual-activity-options.
type Option struct {
	AnnualActivityID   string          `json:"annualActivityId"`
	ItemNumber         string          `json:"itemNumber"`
	AuditObject        string          `json:"auditObject"`
	AuditCategory      string          `json:"auditCategory"`
	AuditUniverse      UniverseRef     `json:"auditUniverse"`
	Department         DepartmentRef   `json:"department"`
	Schedule           string          `json:"schedule"`
	TeamSize           int             `json:"teamSize"`
	TotalMandays       int             `json:"totalMandays"`
	PkatSupervisorName string          `json:"pkatSupervisorName"`
	Available          bool            `json:"available"`
	UsedBy             *UsedByRef      `json:"usedBy"`
	Prefill            Prefill         `json:"prefill"`
	ExistingRow        *ExistingRowRef `json:"existingRow"`
}
