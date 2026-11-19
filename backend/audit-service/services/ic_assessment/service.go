package ic_assessment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/masterclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service runs the assessment workflow on one stack's audit database
type Service struct {
	DB  *gorm.DB
	Now func() time.Time
}

// New returns a service on db
func New(db *gorm.DB) *Service { return &Service{DB: db, Now: time.Now} }

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeEvent(tx *gorm.DB, actor Actor, id *uuid.UUID, action, from, to, reason string, before, after map[string]interface{}, now time.Time) error {
	ev := models.ICAssessmentEvent{
		AssessmentID: id, Action: action, FromStatus: from, ToStatus: to,
		ActorUserID: actor.UserID, ActorName: actor.Name, ActorRoles: actor.RolesString(),
		Reason: reason, Before: before, After: after, IP: actor.IP, UserAgent: truncate(actor.UserAgent, 500),
		CreatedAt: now,
	}
	return tx.Create(&ev).Error
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// IsPeriodLocked reports whether (year, quarter) is locked: the year row, or
// the quarter's own row
func IsPeriodLocked(db *gorm.DB, year, quarter int) (bool, error) {
	var n int64
	q := db.Model(&models.ICAssessmentPeriod{}).Where("year = ? AND status = ?", year, PeriodLocked)
	if quarter > 0 {
		q = q.Where("quarter IN ?", []int{0, quarter})
	}
	// A whole-year assessment is locked by the year row only
	if quarter == 0 {
		q = q.Where("quarter = ?", 0)
	}
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) mustBeOpen(tx *gorm.DB, year, quarter int) error {
	locked, err := IsPeriodLocked(tx, year, quarter)
	if err != nil {
		return err
	}
	if locked {
		return conflictf("PERIOD_LOCKED", "The period %s is locked", periodLabel(year, quarter))
	}
	return nil
}

func periodLabel(year, quarter int) string {
	if quarter == 0 {
		return fmt.Sprint(year)
	}
	return fmt.Sprintf("%d Q%d", year, quarter)
}

func componentOrder(code string) int {
	for i, c := range ComponentCodes {
		if c == code {
			return i
		}
	}
	return len(ComponentCodes)
}

func sortChildren(a *models.ICAssessment) {
	sort.SliceStable(a.Components, func(i, j int) bool {
		return componentOrder(a.Components[i].ComponentCode) < componentOrder(a.Components[j].ComponentCode)
	})
	sort.SliceStable(a.Principles, func(i, j int) bool { return a.Principles[i].PrincipleCode < a.Principles[j].PrincipleCode })
}

func loadFull(db *gorm.DB, id uuid.UUID) (*models.ICAssessment, error) {
	var a models.ICAssessment
	err := db.Preload("Components").Preload("Principles").Where("id = ?", id).Take(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notFound()
	}
	if err != nil {
		return nil, err
	}
	sortChildren(&a)
	return &a, nil
}

func isEditable(status string) bool { return status == StatusDraft || status == StatusReopened }

// ownDepartment resolves a scoped reader's department from the JWT
// employee_code through master-service
func ownDepartment(ctx context.Context, src Sources, actor Actor) (string, error) {
	code := strings.TrimSpace(actor.EmployeeCode)
	if code == "" {
		return "", forbiddenf("Your account is not linked to an employee, so your department is unknown")
	}
	e, err := src.EmployeeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, masterclient.ErrNotFound) {
			return "", forbiddenf("Your employee record was not found, so your department is unknown")
		}
		return "", unavailablef("Could not resolve your department from master-service")
	}
	id, perr := uuid.Parse(strings.TrimSpace(e.DepartmentID))
	if perr != nil || id == uuid.Nil {
		return "", forbiddenf("Your employee record has no department")
	}
	return id.String(), nil
}

// checkRead enforces read access on one assessment (404 for a scoped reader
// outside their scope, so existence is not disclosed)
func (s *Service) checkRead(ctx context.Context, src Sources, actor Actor, a *models.ICAssessment) error {
	if !actor.CanRead() {
		return forbiddenf("No role on this account may read internal control assessments")
	}
	if !actor.IsScopedReader() {
		return nil
	}
	dept, err := ownDepartment(ctx, src, actor)
	if err != nil {
		return err
	}
	if a.Status != StatusApproved || !strings.EqualFold(a.DepartmentID.String(), dept) {
		return notFound()
	}
	return nil
}

// bump performs the optimistic-lock header update: WHERE id, version and
// status match; version + 1. 409 when nothing matched.
func bump(tx *gorm.DB, a *models.ICAssessment, version int, from []string, updates map[string]interface{}, now time.Time) error {
	updates["version"] = gorm.Expr("version + 1")
	updates["updated_at"] = now
	res := tx.Model(&models.ICAssessment{}).
		Where("id = ? AND version = ? AND status IN ?", a.ID, version, from).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return conflictf("VERSION_CONFLICT", "The assessment was changed by someone else (or is no longer in a state that allows this); reload and try again")
	}
	return nil
}

func checkVersion(a *models.ICAssessment, version int) error {
	if version <= 0 {
		return invalidf("version is required")
	}
	if version != a.Version {
		return conflictf("VERSION_CONFLICT", "The assessment is at version %d, not %d; reload and try again", a.Version, version)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Detail
// ---------------------------------------------------------------------------

// Detail is GET /ic-assessments/:id
type Detail struct {
	*models.ICAssessment
	Evidence       []models.ICAssessmentEvidence `json:"evidence"`
	HasSignature   bool                          `json:"has_signature"`
	PeriodLocked   bool                          `json:"period_locked"`
	AllowedActions []string                      `json:"allowed_actions"`
	// SubmitIssues lists what blocks a submit (empty when it would pass)
	SubmitIssues []string  `json:"submit_issues"`
	Thresholds   *Settings `json:"thresholds_used,omitempty"`
}

// AllowedActions lists what actor may do now (the endpoints enforce the same)
func AllowedActions(actor Actor, a *models.ICAssessment, locked bool) []string {
	out := []string{}
	switch a.Status {
	case StatusDraft, StatusReopened:
		if actor.IsAssessor() && !locked {
			out = append(out, "propose", "update", "submit")
			if a.Status == StatusDraft && a.SubmittedAt == nil && (actor.IsAdmin() || SameUser(actor.UserID, a.CreatedByUserID)) {
				out = append(out, "delete")
			}
		}
	case StatusSubmitted:
		if actor.IsApprover() {
			out = append(out, "return")
			if !SameUser(actor.UserID, a.SubmittedByUserID) {
				out = append(out, "approve")
			}
		}
	case StatusApproved:
		if actor.IsApprover() && !locked {
			out = append(out, "reopen")
		}
	}
	return out
}

func (s *Service) detail(db *gorm.DB, actor Actor, a *models.ICAssessment) (*Detail, error) {
	locked, err := IsPeriodLocked(db, a.Year, a.Quarter)
	if err != nil {
		return nil, err
	}
	ev, err := evidenceOf(db, a.ID, a.ProposalRun)
	if err != nil {
		return nil, err
	}
	d := &Detail{ICAssessment: a, Evidence: ev, HasSignature: a.SignatureImage != "", PeriodLocked: locked,
		AllowedActions: AllowedActions(actor, a, locked)}
	d.SubmitIssues = []string{}
	if isEditable(a.Status) {
		if a.Method == MethodQuantitativePoF {
			issues, err := quantSubmitIssues(db, a)
			if err != nil {
				return nil, err
			}
			d.SubmitIssues = issues
		} else if issues := ValidateForSubmit(a); issues != nil {
			d.SubmitIssues = issues
		}
	}
	d.Thresholds = thresholdsFromEvidence(ev)
	return d, nil
}

func evidenceOf(db *gorm.DB, id uuid.UUID, run int) ([]models.ICAssessmentEvidence, error) {
	out := []models.ICAssessmentEvidence{}
	if run <= 0 {
		return out, nil
	}
	err := db.Where("assessment_id = ? AND proposal_run = ?", id, run).Order("component_code").Order("signal_key").Find(&out).Error
	return out, err
}

func thresholdsFromEvidence(ev []models.ICAssessmentEvidence) *Settings {
	for _, e := range ev {
		if e.SignalKey != ContextThresholds {
			continue
		}
		st := Settings{Source: "evidence"}
		st.LHADueDays = toInt(e.Detail["lha_due_days"])
		st.ManagementResponseDueDays = toInt(e.Detail["management_response_due_days"])
		if !ValidThreshold(st.LHADueDays) || !ValidThreshold(st.ManagementResponseDueDays) {
			return nil
		}
		return &st
	}
	return nil
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

// Get returns one assessment
func (s *Service) Get(ctx context.Context, src Sources, actor Actor, id uuid.UUID) (*Detail, error) {
	a, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	if err := s.checkRead(ctx, src, actor, a); err != nil {
		return nil, err
	}
	return s.detail(s.DB, actor, a)
}

// Signature returns the approval signature data URL (read access applies)
func (s *Service) Signature(ctx context.Context, src Sources, actor Actor, id uuid.UUID) (string, error) {
	var a models.ICAssessment
	if err := s.DB.Where("id = ?", id).Take(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", notFound()
		}
		return "", err
	}
	if err := s.checkRead(ctx, src, actor, &a); err != nil {
		return "", err
	}
	if a.SignatureImage == "" {
		return "", apperrors.New("NOT_FOUND", "This assessment has no approval signature", http.StatusNotFound)
	}
	return a.SignatureImage, nil
}

// ---------------------------------------------------------------------------
// List and summary
// ---------------------------------------------------------------------------

// ListFilter is the typed query of GET /ic-assessments
type ListFilter struct {
	Year         *int
	Quarter      *int
	DepartmentID *uuid.UUID
	LocationID   *uuid.UUID
	Status       string
	Sort         string
	Desc         bool
	Page         int
	PageSize     int
}

// SortColumns whitelists the list sort fields (GORM column names)
var SortColumns = map[string]bool{
	"year": true, "quarter": true, "status": true, "department_name": true, "location_name": true,
	"created_at": true, "updated_at": true, "submitted_at": true, "approved_at": true,
}

// ListResult is one page
type ListResult struct {
	Items      []models.ICAssessment `json:"items"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

// List returns a page of assessments with their component rows
func (s *Service) List(ctx context.Context, src Sources, actor Actor, f ListFilter) (*ListResult, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	q := s.DB.Model(&models.ICAssessment{})
	if actor.IsScopedReader() {
		dept, err := ownDepartment(ctx, src, actor)
		if err != nil {
			return nil, err
		}
		q = q.Where("department_id = ? AND status = ?", dept, StatusApproved)
	}
	if f.Year != nil {
		q = q.Where("year = ?", *f.Year)
	}
	if f.Quarter != nil {
		q = q.Where("quarter = ?", *f.Quarter)
	}
	if f.DepartmentID != nil {
		q = q.Where("department_id = ?", *f.DepartmentID)
	}
	if f.LocationID != nil {
		q = q.Where("location_key = ?", f.LocationID.String())
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	sortCol := f.Sort
	if !SortColumns[sortCol] {
		sortCol = "updated_at"
	}
	page, size := f.Page, f.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 500 {
		size = 500
	}
	items := []models.ICAssessment{}
	err := q.Preload("Components").
		Order(clause.OrderByColumn{Column: clause.Column{Name: sortCol}, Desc: f.Desc}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).
		Offset((page - 1) * size).Limit(size).Find(&items).Error
	if err != nil {
		return nil, err
	}
	for i := range items {
		sortChildren(&items[i])
	}
	pages := int((total + int64(size) - 1) / int64(size))
	return &ListResult{Items: items, Page: page, PageSize: size, Total: total, TotalPages: pages}, nil
}

// SummaryRow is one approved assessment in the summary
type SummaryRow struct {
	AssessmentID    uuid.UUID          `json:"assessment_id"`
	DepartmentID    uuid.UUID          `json:"department_id"`
	DepartmentName  string             `json:"department_name"`
	LocationID      *uuid.UUID         `json:"location_id"`
	LocationName    string             `json:"location_name"`
	FinalConclusion string             `json:"final_conclusion"`
	ApprovedAt      *time.Time         `json:"approved_at"`
	Components      []SummaryComponent `json:"components"`
}

// SummaryComponent is one component of a SummaryRow
type SummaryComponent struct {
	ComponentCode string `json:"component_code"`
	FinalStatus   string `json:"final_status"`
	FinalRating   *int   `json:"final_rating"`
	FinalSeverity string `json:"final_severity"`
}

// Summary is GET /ic-assessments/summary
type Summary struct {
	Year         int                       `json:"year"`
	Quarter      int                       `json:"quarter"`
	Departments  []SummaryRow              `json:"departments"`
	ByComponent  map[string]map[string]int `json:"by_component"`
	ByConclusion map[string]int            `json:"by_conclusion"`
	Total        int                       `json:"total"`
}

// GetSummary returns the APPROVED assessments of (year, quarter), the
// department-level ones unless locationID is given, with counts
func (s *Service) GetSummary(ctx context.Context, src Sources, actor Actor, year, quarter int, locationID *uuid.UUID) (*Summary, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	q := s.DB.Preload("Components").Where("status = ? AND year = ? AND quarter = ?", StatusApproved, year, quarter)
	if locationID != nil {
		q = q.Where("location_key = ?", locationID.String())
	} else {
		q = q.Where("location_key = ?", "")
	}
	if actor.IsScopedReader() {
		dept, err := ownDepartment(ctx, src, actor)
		if err != nil {
			return nil, err
		}
		q = q.Where("department_id = ?", dept)
	}
	var rows []models.ICAssessment
	if err := q.Order("department_name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := &Summary{Year: year, Quarter: quarter, Departments: []SummaryRow{},
		ByComponent: map[string]map[string]int{}, ByConclusion: map[string]int{ConclusionEffective: 0, ConclusionNotEffective: 0}}
	for _, c := range ComponentCodes {
		out.ByComponent[c] = map[string]int{PrincipleYes: 0, PrinciplePartial: 0, PrincipleNo: 0}
	}
	for i := range rows {
		a := &rows[i]
		sortChildren(a)
		row := SummaryRow{AssessmentID: a.ID, DepartmentID: a.DepartmentID, DepartmentName: a.DepartmentName,
			LocationID: a.LocationID, LocationName: a.LocationName, FinalConclusion: a.FinalConclusion, ApprovedAt: a.ApprovedAt,
			Components: []SummaryComponent{}}
		for _, c := range a.Components {
			row.Components = append(row.Components, SummaryComponent{c.ComponentCode, c.FinalStatus, c.FinalRating, c.FinalSeverity})
			if m, ok := out.ByComponent[c.ComponentCode]; ok && c.FinalStatus != "" {
				m[c.FinalStatus]++
			}
		}
		out.ByConclusion[a.FinalConclusion]++
		out.Departments = append(out.Departments, row)
	}
	out.Total = len(rows)
	return out, nil
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

// CreateRequest is the body of POST /ic-assessments
type CreateRequest struct {
	DepartmentID string  `json:"department_id"`
	Year         int     `json:"year"`
	Quarter      int     `json:"quarter"`
	LocationID   *string `json:"location_id"`
}

// Create starts a DRAFT assessment with its 5 component and 17 principle rows
func (s *Service) Create(ctx context.Context, src Sources, actor Actor, req CreateRequest) (*Detail, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only AUDITOR, AUDIT_MANAGER, CHIEF_AUDIT_EXECUTIVE or ADMIN can create an assessment")
	}
	deptID, err := uuid.Parse(strings.TrimSpace(req.DepartmentID))
	if err != nil || deptID == uuid.Nil {
		return nil, invalidf("department_id must be a UUID")
	}
	if req.Year < 2000 || req.Year > 2100 {
		return nil, invalidf("year must be between 2000 and 2100")
	}
	if req.Quarter < 0 || req.Quarter > 4 {
		return nil, invalidf("quarter must be 0 (annual) or 1-4")
	}
	var locID *uuid.UUID
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		l, err := uuid.Parse(strings.TrimSpace(*req.LocationID))
		if err != nil || l == uuid.Nil {
			return nil, invalidf("location_id must be a UUID")
		}
		locID = &l
	}
	if err := s.mustBeOpen(s.DB, req.Year, req.Quarter); err != nil {
		return nil, err
	}

	dept, err := src.Department(ctx, deptID.String())
	if err != nil {
		if errors.Is(err, masterclient.ErrNotFound) || strings.Contains(err.Error(), "answered 404") {
			return nil, invalidf("department_id is not a department of this organisation")
		}
		return nil, unavailablef("Could not check the department in master-service")
	}
	if !dept.IsActive {
		return nil, invalidf("The department is not active")
	}
	locName := ""
	if locID != nil {
		l, err := src.Location(ctx, locID.String())
		if err != nil {
			if errors.Is(err, masterclient.ErrNotFound) || strings.Contains(err.Error(), "answered 404") {
				return nil, invalidf("location_id is not a location of this organisation")
			}
			return nil, unavailablef("Could not check the location in master-service")
		}
		locName = strings.TrimSpace(l.Name)
	}

	// The method comes from the branch setting and is fixed for the life of
	// the assessment. QUANTITATIVE_POF copies the catalogue into the assessment.
	locKey := ""
	if locID != nil {
		locKey = locID.String()
	}
	method, _, err := ResolveMethod(s.DB, locKey)
	if err != nil {
		return nil, err
	}
	var pofSetup *pofCreateSetup
	if method == MethodQuantitativePoF {
		if pofSetup, err = s.preparePoFCreate(ctx, src); err != nil {
			return nil, err
		}
	}

	now := s.now()
	start, end := PeriodBounds(req.Year, req.Quarter)
	a := &models.ICAssessment{
		DepartmentID: deptID, DepartmentCode: dept.DepartmentCode, DepartmentName: strings.TrimSpace(dept.DepartmentName),
		LocationID: locID, LocationName: locName, Year: req.Year, Quarter: req.Quarter,
		PeriodStart: start, PeriodEnd: end, Status: StatusDraft, Version: 1,
		DerivedConclusion: ConclusionIncomplete, Method: method,
		CreatedByUserID: actor.UserID, CreatedByName: actor.Name, CreatedAt: now, UpdatedAt: now,
	}
	if pofSetup != nil {
		pofSetup.applyHeader(a)
	}
	if c, err := uuid.Parse(strings.TrimSpace(dept.CompanyID)); err == nil && c != uuid.Nil {
		a.CompanyID = &c
	}
	if locID != nil {
		a.LocationKey = locID.String()
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&models.ICAssessment{}).
			Where("department_id = ? AND location_key = ? AND year = ? AND quarter = ?", deptID, a.LocationKey, req.Year, req.Quarter).
			Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return conflictf("ASSESSMENT_EXISTS", "An assessment for this department, branch and period already exists")
		}
		if err := tx.Omit("Components", "Principles").Create(a).Error; err != nil {
			if isUniqueViolation(err) {
				return conflictf("ASSESSMENT_EXISTS", "An assessment for this department, branch and period already exists")
			}
			return err
		}
		comps := make([]models.ICAssessmentComponent, 0, len(ComponentCodes))
		for _, code := range ComponentCodes {
			comps = append(comps, models.ICAssessmentComponent{AssessmentID: a.ID, ComponentCode: code,
				DerivedSeverity: SeverityNone, CreatedAt: now, UpdatedAt: now})
		}
		if err := tx.Create(&comps).Error; err != nil {
			return err
		}
		prins := make([]models.ICAssessmentPrinciple, 0, len(Principles))
		for _, p := range Principles {
			prins = append(prins, models.ICAssessmentPrinciple{AssessmentID: a.ID, ComponentCode: p.ComponentCode,
				PrincipleCode: p.Code, Status: PrincipleNotAssessed, Severity: SeverityNone, CreatedAt: now, UpdatedAt: now})
		}
		if err := tx.Create(&prins).Error; err != nil {
			return err
		}
		created := map[string]interface{}{
			"department_id": deptID.String(), "department_name": a.DepartmentName, "location_id": a.LocationKey,
			"year": req.Year, "quarter": req.Quarter, "method": method,
		}
		if pofSetup != nil {
			nPoF, nEl, err := pofSetup.insertRows(tx, a.ID, now)
			if err != nil {
				return err
			}
			created["catalog_version"] = pofSetup.Catalogue.Version
			created["pofs"], created["elements"] = nPoF, nEl
		}
		return writeEvent(tx, actor, &a.ID, ActionCreate, "", StatusDraft, "", nil, created, now)
	})
	if err != nil {
		return nil, err
	}
	full, err := loadFull(s.DB, a.ID)
	if err != nil {
		return nil, err
	}
	return s.detail(s.DB, actor, full)
}

func isUniqueViolation(err error) bool {
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "duplicate key") || strings.Contains(m, "unique constraint") || strings.Contains(m, "23505")
}

// ---------------------------------------------------------------------------
// Propose
// ---------------------------------------------------------------------------

// ProposeRequest is the optional body of POST /ic-assessments/:id/propose
type ProposeRequest struct {
	Version int `json:"version"`
}

// ScopeOf builds the collection scope of an assessment as of asOf
func ScopeOf(a *models.ICAssessment, asOf time.Time, st Settings) Scope {
	sc := Scope{DepartmentID: a.DepartmentID.String(), DepartmentName: a.DepartmentName, LocationName: a.LocationName,
		Year: a.Year, Quarter: a.Quarter, PeriodStart: dayOf(a.PeriodStart), PeriodEnd: dayOf(a.PeriodEnd), AsOf: dayOf(asOf), Settings: st}
	if a.LocationID != nil {
		sc.LocationID = a.LocationID.String()
	}
	return sc
}

// Propose collects the signals, stores them as a new proposal run and writes
// the proposed component scores. Final values are untouched.
func (s *Service) Propose(ctx context.Context, src Sources, actor Actor, id uuid.UUID, req ProposeRequest) (*Detail, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can run a proposal")
	}
	a, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	if a.Method == MethodQuantitativePoF {
		return nil, methodMismatch(a.Method, "a signal proposal exists only for the QUALITATIVE method")
	}
	if !isEditable(a.Status) {
		return nil, conflictf("INVALID_STATUS", "A proposal can only be run on a DRAFT or REOPENED assessment (current status: %s)", a.Status)
	}
	if req.Version > 0 {
		if err := checkVersion(a, req.Version); err != nil {
			return nil, err
		}
	}
	if err := s.mustBeOpen(s.DB, a.Year, a.Quarter); err != nil {
		return nil, err
	}
	st, err := LoadSettings(s.DB)
	if err != nil {
		return nil, err
	}
	now := s.now()
	asOf := AsOfDate(dayOf(a.PeriodEnd), now)
	coll, err := Collect(ctx, s.DB, src, ScopeOf(a, asOf, st))
	if err != nil {
		return nil, err
	}
	run := a.ProposalRun + 1
	hash := HashSignals(coll.Signals)

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.ICAssessment{}).
			Where("id = ? AND version = ? AND proposal_run = ? AND status IN ?", a.ID, a.Version, a.ProposalRun, []string{StatusDraft, StatusReopened}).
			Updates(map[string]interface{}{
				"proposal_run": run, "as_of_date": asOf, "evidence_hash": hash,
				"version": gorm.Expr("version + 1"), "updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return conflictf("VERSION_CONFLICT", "The assessment was changed by someone else; reload and try again")
		}
		rows := make([]models.ICAssessmentEvidence, 0, len(coll.Signals))
		var unavailableKeys []string
		for _, sig := range coll.Signals {
			rows = append(rows, models.ICAssessmentEvidence{
				AssessmentID: a.ID, ProposalRun: run, ComponentCode: sig.Component, SignalKey: sig.Key,
				PrincipleHint: sig.PrincipleHint, SourceService: sig.SourceService, SourceRef: sig.SourceRef,
				State: sig.State, ValuePct: sig.ValuePct, Numerator: sig.Numerator, Denominator: sig.Denominator,
				Weight: sig.Weight, ScopeNote: sig.ScopeNote, Detail: sig.Detail, CollectedAt: now,
			})
			if sig.State == StateSourceUnavailable {
				unavailableKeys = append(unavailableKeys, sig.Key)
			}
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}

		scores := map[string]interface{}{}
		a.ProposalRun = run
		for i := range a.Components {
			c := &a.Components[i]
			sc := ScoreComponent(c.ComponentCode, coll.Signals)
			c.ProposedScorePct, c.ProposedRating = sc.ScorePct, sc.Rating
			c.ProposedStatus, c.ProposedSeverity = sc.Status, sc.Severity
			cov := sc.CoveragePct
			c.ProposalDataCoveragePct = &cov
			scores[c.ComponentCode] = map[string]interface{}{"score_pct": sc.ScorePct, "rating": sc.Rating, "coverage_pct": cov}
		}
		means := map[string]struct {
			n   int
			avg *float64
		}{}
		for _, m := range coll.Principles {
			means[strings.ToUpper(strings.TrimSpace(m.PrincipleCode))] = struct {
				n   int
				avg *float64
			}{m.ControlCount, m.AvgPct}
		}
		for i := range a.Principles {
			p := &a.Principles[i]
			m := means[p.PrincipleCode]
			p.TaggedControlCount, p.TaggedControlAvgPct = m.n, m.avg
			if err := tx.Model(&models.ICAssessmentPrinciple{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
				"tagged_control_count": m.n, "tagged_control_avg_pct": m.avg,
			}).Error; err != nil {
				return err
			}
		}
		Recompute(a)
		if err := saveComponents(tx, a.Components, now); err != nil {
			return err
		}
		if err := tx.Model(&models.ICAssessment{}).Where("id = ?", a.ID).
			Update("derived_conclusion", a.DerivedConclusion).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionPropose, a.Status, a.Status, "", nil, map[string]interface{}{
			"proposal_run": run, "as_of_date": asOf.Format("2006-01-02"), "scores": scores,
			"partial": len(unavailableKeys) > 0, "source_unavailable": unavailableKeys, "evidence_hash": hash,
			"lha_due_days": st.LHADueDays, "management_response_due_days": st.ManagementResponseDueDays,
		}, now)
	})
	if err != nil {
		return nil, err
	}
	full, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	return s.detail(s.DB, actor, full)
}

func saveComponents(tx *gorm.DB, comps []models.ICAssessmentComponent, now time.Time) error {
	for _, c := range comps {
		if err := tx.Model(&models.ICAssessmentComponent{}).Where("id = ?", c.ID).Updates(map[string]interface{}{
			"proposed_score_pct": c.ProposedScorePct, "proposed_rating": c.ProposedRating,
			"proposed_status": c.ProposedStatus, "proposed_severity": c.ProposedSeverity,
			"proposal_data_coverage_pct": c.ProposalDataCoveragePct,
			"final_rating":               c.FinalRating, "final_status": c.FinalStatus, "final_severity": c.FinalSeverity,
			"derived_status": c.DerivedStatus, "derived_severity": c.DerivedSeverity,
			"conclusion": c.Conclusion, "override_reason": c.OverrideReason, "is_override": c.IsOverride,
			"reviewed_by_user_id": c.ReviewedByUserID, "reviewed_by_name": c.ReviewedByName, "reviewed_at": c.ReviewedAt,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// ComponentInput changes one component; nil fields are unchanged.
// final_rating 0 clears the rating; "" clears a status / severity.
type ComponentInput struct {
	ComponentCode  string  `json:"component_code"`
	FinalRating    *int    `json:"final_rating"`
	FinalStatus    *string `json:"final_status"`
	FinalSeverity  *string `json:"final_severity"`
	Conclusion     *string `json:"conclusion"`
	OverrideReason *string `json:"override_reason"`
}

// PrincipleInput changes one principle; nil fields are unchanged
type PrincipleInput struct {
	PrincipleCode string  `json:"principle_code"`
	Status        *string `json:"status"`
	Severity      *string `json:"severity"`
	Rationale     *string `json:"rationale"`
}

// UpdateRequest is the body of PUT /ic-assessments/:id
type UpdateRequest struct {
	Version                 int              `json:"version"`
	Components              []ComponentInput `json:"components"`
	Principles              []PrincipleInput `json:"principles"`
	FinalConclusion         *string          `json:"final_conclusion"`
	ConclusionText          *string          `json:"conclusion_text"`
	ConclusionJustification *string          `json:"conclusion_justification"`
}

func upper(p *string) string { return strings.ToUpper(strings.TrimSpace(*p)) }

func ratingVal(r *int) interface{} {
	if r == nil {
		return nil
	}
	return *r
}

// Update saves principle ratings, component final values and the conclusion
func (s *Service) Update(ctx context.Context, actor Actor, id uuid.UUID, req UpdateRequest) (*Detail, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can change an assessment")
	}
	if s.methodOf(id) == MethodQuantitativePoF {
		return s.updateQuantConclusion(actor, id, req)
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if !isEditable(a.Status) {
			return conflictf("INVALID_STATUS", "Only a DRAFT or REOPENED assessment can be changed (current status: %s)", a.Status)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		before, after := map[string]interface{}{}, map[string]interface{}{}
		changedPrinciples := map[int]bool{}
		changedComponents := map[int]bool{}

		pIndex := map[string]int{}
		for i, p := range a.Principles {
			pIndex[p.PrincipleCode] = i
		}
		for _, in := range req.Principles {
			code := strings.ToUpper(strings.TrimSpace(in.PrincipleCode))
			i, ok := pIndex[code]
			if !ok {
				return invalidf("unknown principle_code %q", in.PrincipleCode)
			}
			p := &a.Principles[i]
			old := *p
			if in.Status != nil {
				p.Status = upper(in.Status)
				if in.Severity == nil {
					// The only severity the status allows, or the lowest
					switch p.Status {
					case PrincipleNo:
						p.Severity = SeverityMajor
					case PrinciplePartial:
						if SeverityRank(p.Severity) < SeverityRank(SeverityDeficiency) || p.Severity == SeverityMajor {
							p.Severity = SeverityDeficiency
						}
					case PrincipleYes:
						if SeverityRank(p.Severity) > SeverityRank(SeverityDeficiency) {
							p.Severity = SeverityNone
						}
					default:
						p.Severity = SeverityNone
					}
				}
			}
			if in.Severity != nil {
				p.Severity = upper(in.Severity)
			}
			if in.Rationale != nil {
				p.Rationale = strings.TrimSpace(*in.Rationale)
			}
			if err := ValidatePrinciple(p); err != nil {
				return invalidf("%s", err.Error())
			}
			if old.Status != p.Status || old.Severity != p.Severity || old.Rationale != p.Rationale {
				changedPrinciples[i] = true
				before["principles."+code] = map[string]interface{}{"status": old.Status, "severity": old.Severity, "rationale": old.Rationale}
				after["principles."+code] = map[string]interface{}{"status": p.Status, "severity": p.Severity, "rationale": p.Rationale}
			}
		}

		cIndex := map[string]int{}
		for i, c := range a.Components {
			cIndex[c.ComponentCode] = i
		}
		ratingChanged := map[int]bool{}
		for _, in := range req.Components {
			code := strings.ToUpper(strings.TrimSpace(in.ComponentCode))
			i, ok := cIndex[code]
			if !ok {
				return invalidf("unknown component_code %q", in.ComponentCode)
			}
			c := &a.Components[i]
			old := *c
			if in.FinalRating != nil {
				if *in.FinalRating == 0 {
					c.FinalRating = nil
				} else {
					v := *in.FinalRating
					c.FinalRating = &v
				}
			}
			if in.FinalStatus != nil {
				c.FinalStatus = upper(in.FinalStatus)
			}
			if in.FinalSeverity != nil {
				c.FinalSeverity = upper(in.FinalSeverity)
			}
			if in.Conclusion != nil {
				c.Conclusion = strings.TrimSpace(*in.Conclusion)
			}
			if in.OverrideReason != nil {
				c.OverrideReason = strings.TrimSpace(*in.OverrideReason)
			}
			changed := ratingVal(old.FinalRating) != ratingVal(c.FinalRating) || old.FinalStatus != c.FinalStatus ||
				old.FinalSeverity != c.FinalSeverity || old.Conclusion != c.Conclusion || old.OverrideReason != c.OverrideReason
			if changed {
				changedComponents[i] = true
				ratingChanged[i] = ratingVal(old.FinalRating) != ratingVal(c.FinalRating)
				c.ReviewedByUserID, c.ReviewedByName = actor.UserID, actor.Name
				t := now
				c.ReviewedAt = &t
				before["components."+code] = map[string]interface{}{"final_rating": ratingVal(old.FinalRating), "final_status": old.FinalStatus,
					"final_severity": old.FinalSeverity, "conclusion": old.Conclusion, "override_reason": old.OverrideReason}
				after["components."+code] = map[string]interface{}{"final_rating": ratingVal(c.FinalRating), "final_status": c.FinalStatus,
					"final_severity": c.FinalSeverity, "conclusion": c.Conclusion, "override_reason": c.OverrideReason}
			}
		}

		hdrBefore := map[string]interface{}{"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification}
		if req.FinalConclusion != nil {
			a.FinalConclusion = upper(req.FinalConclusion)
		}
		if req.ConclusionText != nil {
			a.ConclusionText = strings.TrimSpace(*req.ConclusionText)
		}
		if req.ConclusionJustification != nil {
			a.ConclusionJustification = strings.TrimSpace(*req.ConclusionJustification)
		}
		for k, v := range map[string]string{"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification} {
			if hdrBefore[k] != v {
				before[k], after[k] = hdrBefore[k], v
			}
		}

		Recompute(a)
		for i := range a.Components {
			c := &a.Components[i]
			d := DeriveComponent(PrinciplesOf(c.ComponentCode, a.Principles))
			if d.Applicable == 0 {
				return invalidf("%s: at least one principle of each component must be applicable (COSO requires all five components)", c.ComponentCode)
			}
			if err := ValidateComponentFinal(c, d, changedComponents[i]); err != nil {
				return invalidf("%s", err.Error())
			}
		}
		if err := ValidateFinalConclusion(a, a.Components, a.Principles, false); err != nil {
			return invalidf("%s", err.Error())
		}
		if len(after) == 0 {
			return nil
		}

		for i := range changedPrinciples {
			p := a.Principles[i]
			if err := tx.Model(&models.ICAssessmentPrinciple{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
				"status": p.Status, "severity": p.Severity, "rationale": p.Rationale,
				"updated_by_user_id": actor.UserID, "updated_by_name": actor.Name, "updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		if err := saveComponents(tx, a.Components, now); err != nil {
			return err
		}
		if err := bump(tx, a, req.Version, []string{StatusDraft, StatusReopened}, map[string]interface{}{
			"final_conclusion": a.FinalConclusion, "conclusion_text": a.ConclusionText,
			"conclusion_justification": a.ConclusionJustification, "derived_conclusion": a.DerivedConclusion,
		}, now); err != nil {
			return err
		}
		after["derived_conclusion"] = a.DerivedConclusion
		if err := writeEvent(tx, actor, &a.ID, ActionUpdate, a.Status, a.Status, "", before, after, now); err != nil {
			return err
		}
		for i := range changedComponents {
			c := a.Components[i]
			if !c.IsOverride || !ratingChanged[i] {
				continue
			}
			if err := writeEvent(tx, actor, &a.ID, ActionOverride, a.Status, a.Status, c.OverrideReason,
				map[string]interface{}{"component_code": c.ComponentCode, "proposed_rating": ratingVal(c.ProposedRating)},
				map[string]interface{}{"component_code": c.ComponentCode, "final_rating": ratingVal(c.FinalRating)}, now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	full, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	return s.detail(s.DB, actor, full)
}

// ---------------------------------------------------------------------------
// Workflow transitions
// ---------------------------------------------------------------------------

// TransitionRequest is the body of submit / return / reopen
type TransitionRequest struct {
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}

// ApproveRequest is the body of POST /ic-assessments/:id/approve. The
// signature is optional; the HTTP layer validates the image.
type ApproveRequest struct {
	Version             int    `json:"version"`
	OperateTogether     *bool  `json:"operate_together"`
	OperateTogetherNote string `json:"operate_together_note"`
	ApprovalNote        string `json:"approval_note"`
	SignatureImage      string `json:"signature_image"`
	SignerName          string `json:"signer_name"`
	SignerRole          string `json:"signer_role"`
}

func (s *Service) finish(actor Actor, id uuid.UUID) (*Detail, error) {
	full, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	return s.detail(s.DB, actor, full)
}

// Submit moves DRAFT / REOPENED to SUBMITTED when every rule passes
func (s *Service) Submit(ctx context.Context, actor Actor, id uuid.UUID, req TransitionRequest) (*Detail, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can submit an assessment")
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if !isEditable(a.Status) {
			return conflictf("INVALID_STATUS", "Only a DRAFT or REOPENED assessment can be submitted (current status: %s)", a.Status)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		if a.Method == MethodQuantitativePoF {
			issues, err := quantSubmitIssues(tx, a)
			if err != nil {
				return err
			}
			if len(issues) > 0 {
				return invalidList("The assessment cannot be submitted yet", issues)
			}
		} else {
			Recompute(a)
			if issues := ValidateForSubmit(a); len(issues) > 0 {
				return invalidList("The assessment cannot be submitted yet", issues)
			}
		}
		if err := bump(tx, a, req.Version, []string{StatusDraft, StatusReopened}, map[string]interface{}{
			"status": StatusSubmitted, "submitted_by_user_id": actor.UserID, "submitted_by_name": actor.Name,
			"submitted_at": now, "derived_conclusion": a.DerivedConclusion,
		}, now); err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionSubmit, a.Status, StatusSubmitted, strings.TrimSpace(req.Reason), nil,
			map[string]interface{}{"derived_conclusion": a.DerivedConclusion, "final_conclusion": a.FinalConclusion}, now)
	})
	if err != nil {
		return nil, err
	}
	return s.finish(actor, id)
}

// Return sends a SUBMITTED assessment back to DRAFT (reason required)
func (s *Service) Return(ctx context.Context, actor Actor, id uuid.UUID, req TransitionRequest) (*Detail, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can return an assessment")
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, invalidf("reason is required")
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Status != StatusSubmitted {
			return conflictf("INVALID_STATUS", "Only a SUBMITTED assessment can be returned (current status: %s)", a.Status)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		if err := bump(tx, a, req.Version, []string{StatusSubmitted}, map[string]interface{}{"status": StatusDraft}, now); err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionReturn, StatusSubmitted, StatusDraft, reason, nil, nil, now)
	})
	if err != nil {
		return nil, err
	}
	return s.finish(actor, id)
}

// Approve moves SUBMITTED to APPROVED. The approver may not be the
// submitter (segregation of duties) and must answer operate_together; false
// forces NOT_EFFECTIVE.
func (s *Service) Approve(ctx context.Context, actor Actor, id uuid.UUID, req ApproveRequest) (*Detail, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can approve an assessment")
	}
	if req.OperateTogether == nil {
		return nil, invalidf("operate_together is required: confirm whether the five components operate together")
	}
	note := strings.TrimSpace(req.OperateTogetherNote)
	if !*req.OperateTogether && note == "" {
		return nil, invalidf("operate_together_note is required when the components do not operate together")
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Status != StatusSubmitted {
			return conflictf("INVALID_STATUS", "Only a SUBMITTED assessment can be approved (current status: %s)", a.Status)
		}
		if SameUser(actor.UserID, a.SubmittedByUserID) {
			return apperrors.New("SEGREGATION_OF_DUTIES", "The approver must not be the person who submitted the assessment", http.StatusForbidden)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		final := a.FinalConclusion
		if !*req.OperateTogether {
			final = ConclusionNotEffective
		}
		updates := map[string]interface{}{
			"status": StatusApproved, "operate_together": *req.OperateTogether, "operate_together_note": note,
			"final_conclusion": final, "approved_by_user_id": actor.UserID, "approved_by_name": actor.Name,
			"approved_by_role": actor.ApproverRoleName(), "approved_at": now, "approval_note": strings.TrimSpace(req.ApprovalNote),
			"signature_image": "", "signer_name": "", "signer_role": "",
		}
		if img := strings.TrimSpace(req.SignatureImage); img != "" {
			signer := strings.TrimSpace(req.SignerName)
			if signer == "" {
				signer = actor.Name
			}
			updates["signature_image"], updates["signer_name"], updates["signer_role"] = img, truncate(signer, 200), truncate(strings.TrimSpace(req.SignerRole), 100)
		}
		if err := bump(tx, a, req.Version, []string{StatusSubmitted}, updates, now); err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionApprove, StatusSubmitted, StatusApproved, strings.TrimSpace(req.ApprovalNote),
			map[string]interface{}{"final_conclusion": a.FinalConclusion},
			map[string]interface{}{"final_conclusion": final, "operate_together": *req.OperateTogether,
				"operate_together_note": note, "has_signature": updates["signature_image"] != "", "evidence_hash": a.EvidenceHash,
				"proposal_run": a.ProposalRun}, now)
	})
	if err != nil {
		return nil, err
	}
	return s.finish(actor, id)
}

// Reopen moves APPROVED to REOPENED (period open, reason required). The
// approval stays in the event log.
func (s *Service) Reopen(ctx context.Context, actor Actor, id uuid.UUID, req TransitionRequest) (*Detail, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can reopen an assessment")
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, invalidf("reason is required")
	}
	now := s.now()
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := loadFull(tx, id)
		if err != nil {
			return err
		}
		if a.Status != StatusApproved {
			return conflictf("INVALID_STATUS", "Only an APPROVED assessment can be reopened (current status: %s)", a.Status)
		}
		if err := checkVersion(a, req.Version); err != nil {
			return err
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		before := map[string]interface{}{
			"final_conclusion": a.FinalConclusion, "operate_together": a.OperateTogether, "approved_by_user_id": a.ApprovedByUserID,
			"approved_by_name": a.ApprovedByName, "approved_by_role": a.ApprovedByRole, "approved_at": a.ApprovedAt,
			"approval_note": a.ApprovalNote, "has_signature": a.SignatureImage != "", "signer_name": a.SignerName,
		}
		if err := bump(tx, a, req.Version, []string{StatusApproved}, map[string]interface{}{
			"status": StatusReopened, "reopen_count": gorm.Expr("reopen_count + 1"), "last_reopen_reason": reason,
			"operate_together": nil, "operate_together_note": "", "approved_by_user_id": "", "approved_by_name": "",
			"approved_by_role": "", "approved_at": nil, "approval_note": "", "signature_image": "", "signer_name": "", "signer_role": "",
		}, now); err != nil {
			return err
		}
		return writeEvent(tx, actor, &a.ID, ActionReopen, StatusApproved, StatusReopened, reason, before,
			map[string]interface{}{"reopen_count": a.ReopenCount + 1}, now)
	})
	if err != nil {
		return nil, err
	}
	return s.finish(actor, id)
}

// Delete soft-deletes a DRAFT that was never submitted (creator or ADMIN)
func (s *Service) Delete(ctx context.Context, actor Actor, id uuid.UUID) error {
	if !actor.IsAssessor() {
		return forbiddenf("Only the creator or an ADMIN can delete a draft assessment")
	}
	now := s.now()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var a models.ICAssessment
		if err := tx.Where("id = ?", id).Take(&a).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return notFound()
			}
			return err
		}
		if !actor.IsAdmin() && !SameUser(actor.UserID, a.CreatedByUserID) {
			return forbiddenf("Only the creator or an ADMIN can delete a draft assessment")
		}
		if a.Status != StatusDraft || a.SubmittedAt != nil {
			return conflictf("INVALID_STATUS", "Only a DRAFT that was never submitted can be deleted")
		}
		if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
			return err
		}
		res := tx.Where("id = ? AND status = ? AND submitted_at IS NULL", id, StatusDraft).Delete(&models.ICAssessment{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return conflictf("VERSION_CONFLICT", "The assessment was changed by someone else; reload and try again")
		}
		return writeEvent(tx, actor, &a.ID, ActionDelete, StatusDraft, "", "", map[string]interface{}{
			"department_id": a.DepartmentID.String(), "year": a.Year, "quarter": a.Quarter, "location_id": a.LocationKey,
		}, nil, now)
	})
}

// ---------------------------------------------------------------------------
// Evidence, events, drift
// ---------------------------------------------------------------------------

// Evidence returns the evidence rows of one run (0 = latest)
func (s *Service) Evidence(ctx context.Context, src Sources, actor Actor, id uuid.UUID, run int) ([]models.ICAssessmentEvidence, error) {
	var a models.ICAssessment
	if err := s.DB.Where("id = ?", id).Take(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound()
		}
		return nil, err
	}
	if err := s.checkRead(ctx, src, actor, &a); err != nil {
		return nil, err
	}
	if run == 0 {
		run = a.ProposalRun
	}
	if run < 0 || run > a.ProposalRun {
		return nil, invalidf("run must be between 1 and %d", a.ProposalRun)
	}
	return evidenceOf(s.DB, id, run)
}

// Events returns the audit trail (not for DEPARTMENT_HEAD / AUDITEE)
func (s *Service) Events(ctx context.Context, actor Actor, id uuid.UUID) ([]models.ICAssessmentEvent, error) {
	if !actor.CanRead() || actor.IsScopedReader() {
		return nil, forbiddenf("Your role cannot read the assessment history")
	}
	var n int64
	if err := s.DB.Unscoped().Model(&models.ICAssessment{}).Where("id = ?", id).Count(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, notFound()
	}
	out := []models.ICAssessmentEvent{}
	err := s.DB.Where("assessment_id = ?", id).Order("created_at").Order("id").Find(&out).Error
	return out, err
}

// DriftSignal compares one scored signal of the frozen run with today
type DriftSignal struct {
	SignalKey     string             `json:"signal_key"`
	ComponentCode string             `json:"component_code"`
	Frozen        *SignalFingerprint `json:"frozen"`
	Current       *SignalFingerprint `json:"current"`
	// Changed is null when today's value could not be read
	Changed *bool `json:"changed"`
}

// Drift is GET /ic-assessments/:id/drift: "source data changed since the
// proposal (approval)". It re-runs the collectors read-only with the frozen
// as-of date and thresholds; nothing is written.
type Drift struct {
	AssessmentID uuid.UUID     `json:"assessment_id"`
	Status       string        `json:"status"`
	ProposalRun  int           `json:"proposal_run"`
	AsOfDate     *time.Time    `json:"as_of_date"`
	FrozenHash   string        `json:"frozen_hash"`
	CurrentHash  string        `json:"current_hash"`
	Changed      bool          `json:"changed"`
	Complete     bool          `json:"complete"` // every source answered today
	Signals      []DriftSignal `json:"signals"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// CheckDrift compares the latest proposal's evidence with the data today
func (s *Service) CheckDrift(ctx context.Context, src Sources, actor Actor, id uuid.UUID) (*Drift, error) {
	if !actor.CanRead() || actor.IsScopedReader() {
		return nil, forbiddenf("Your role cannot run the data change check")
	}
	a, err := loadFull(s.DB, id)
	if err != nil {
		return nil, err
	}
	if a.ProposalRun < 1 || a.AsOfDate == nil {
		return nil, conflictf("NO_PROPOSAL", "The assessment has no proposal to compare with")
	}
	frozen, err := evidenceOf(s.DB, id, a.ProposalRun)
	if err != nil {
		return nil, err
	}
	st := thresholdsFromEvidence(frozen)
	if st == nil {
		cur, err := LoadSettings(s.DB)
		if err != nil {
			return nil, err
		}
		st = &cur
	}
	coll, err := Collect(ctx, s.DB, src, ScopeOf(a, *a.AsOfDate, *st))
	if err != nil {
		return nil, err
	}
	out := &Drift{AssessmentID: a.ID, Status: a.Status, ProposalRun: a.ProposalRun, AsOfDate: a.AsOfDate,
		FrozenHash: a.EvidenceHash, CurrentHash: HashSignals(coll.Signals), Complete: true, CheckedAt: s.now(), Signals: []DriftSignal{}}
	frozenBy := map[string]SignalFingerprint{}
	for _, e := range frozen {
		frozenBy[e.SignalKey] = SignalFingerprint{Key: e.SignalKey, State: e.State, ValuePct: e.ValuePct, Numerator: e.Numerator, Denominator: e.Denominator}
	}
	currentBy := map[string]SignalFingerprint{}
	for _, sig := range coll.Signals {
		currentBy[sig.Key] = fingerprintOf(sig)
	}
	for _, d := range SignalDefs {
		ds := DriftSignal{SignalKey: d.Key, ComponentCode: d.Component}
		f, hasF := frozenBy[d.Key]
		c, hasC := currentBy[d.Key]
		if hasF {
			ds.Frozen = &f
		}
		if hasC {
			ds.Current = &c
		}
		if !hasC || c.State == StateSourceUnavailable {
			out.Complete = false
		} else {
			changed := !hasF || f.String() != c.String()
			ds.Changed = &changed
			if changed {
				out.Changed = true
			}
		}
		out.Signals = append(out.Signals, ds)
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Periods
// ---------------------------------------------------------------------------

// PeriodRequest is the body of POST /ic-assessments/periods/lock and /unlock
type PeriodRequest struct {
	Year    int    `json:"year"`
	Quarter int    `json:"quarter"`
	Note    string `json:"note"`
	Reason  string `json:"reason"`
}

// ListPeriods returns the period rows of year (all years when 0)
func (s *Service) ListPeriods(actor Actor, year int) ([]models.ICAssessmentPeriod, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	out := []models.ICAssessmentPeriod{}
	q := s.DB.Order("year").Order("quarter")
	if year > 0 {
		q = q.Where("year = ?", year)
	}
	return out, q.Find(&out).Error
}

func validPeriod(year, quarter int) error {
	if year < 2000 || year > 2100 {
		return invalidf("year must be between 2000 and 2100")
	}
	if quarter < 0 || quarter > 4 {
		return invalidf("quarter must be 0 (whole year) or 1-4")
	}
	return nil
}

// LockPeriod locks (year, quarter)
func (s *Service) LockPeriod(actor Actor, req PeriodRequest) (*models.ICAssessmentPeriod, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can lock a period")
	}
	if err := validPeriod(req.Year, req.Quarter); err != nil {
		return nil, err
	}
	now := s.now()
	var row models.ICAssessmentPeriod
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("year = ? AND quarter = ?", req.Year, req.Quarter).Take(&row).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			row = models.ICAssessmentPeriod{Year: req.Year, Quarter: req.Quarter, CreatedAt: now}
		case err != nil:
			return err
		case row.Status == PeriodLocked:
			return conflictf("PERIOD_LOCKED", "The period %s is already locked", periodLabel(req.Year, req.Quarter))
		}
		t := now
		row.Status, row.LockedByUserID, row.LockedByName, row.LockedAt = PeriodLocked, actor.UserID, actor.Name, &t
		row.LockNote, row.UpdatedAt = strings.TrimSpace(req.Note), now
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, nil, ActionPeriodLock, PeriodOpen, PeriodLocked, row.LockNote, nil,
			map[string]interface{}{"year": req.Year, "quarter": req.Quarter}, now)
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UnlockPeriod opens (year, quarter); reason required
func (s *Service) UnlockPeriod(actor Actor, req PeriodRequest) (*models.ICAssessmentPeriod, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can unlock a period")
	}
	if err := validPeriod(req.Year, req.Quarter); err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, invalidf("reason is required")
	}
	now := s.now()
	var row models.ICAssessmentPeriod
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("year = ? AND quarter = ?", req.Year, req.Quarter).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && row.Status != PeriodLocked) {
			return conflictf("PERIOD_NOT_LOCKED", "The period %s is not locked", periodLabel(req.Year, req.Quarter))
		}
		if err != nil {
			return err
		}
		if req.Quarter > 0 {
			if yearLocked, err := IsPeriodLocked(tx, req.Year, 0); err != nil {
				return err
			} else if yearLocked {
				return conflictf("PERIOD_LOCKED", "The whole year %d is locked; unlock the year first", req.Year)
			}
		}
		row.Status, row.UpdatedAt = PeriodOpen, now
		if err := tx.Model(&models.ICAssessmentPeriod{}).Where("id = ?", row.ID).
			Updates(map[string]interface{}{"status": PeriodOpen, "updated_at": now}).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, nil, ActionPeriodUnlock, PeriodLocked, PeriodOpen, reason, nil,
			map[string]interface{}{"year": req.Year, "quarter": req.Quarter}, now)
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}
