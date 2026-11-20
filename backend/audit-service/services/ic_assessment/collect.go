package ic_assessment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/authclient"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/riskclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sources are the other services an assessment reads, on behalf of the
// caller (each call forwards the caller's token). Tests replace it.
type Sources interface {
	Department(ctx context.Context, id string) (*masterclient.DepartmentDetail, error)
	Location(ctx context.Context, id string) (*masterclient.Location, error)
	ActiveEmployees(ctx context.Context, departmentID string) ([]masterclient.Employee, error)
	EmployeeByCode(ctx context.Context, code string) (*masterclient.Employee, error)
	Acceptance(ctx context.Context, employeeCodes []string) ([]authclient.Acceptance, error)
	CosoSignals(ctx context.Context, q riskclient.CosoSignalQuery) (*riskclient.CosoSignals, error)
	// PofCatalogue reads one COSO PoF catalogue version ("" = current) for
	// QUANTITATIVE_POF assessments
	PofCatalogue(ctx context.Context, version string) (*riskclient.PofCatalogue, error)
}

// HTTPSources calls master-, auth- and risk-service with one caller's
// Authorization header
type HTTPSources struct {
	Authorization string
	Master        *masterclient.MasterClient
	Auth          *authclient.Client
	Risk          *riskclient.RiskClient
}

// NewHTTPSources returns the sources of this stack for one request
func NewHTTPSources(authorization string) *HTTPSources {
	return &HTTPSources{
		Authorization: authorization,
		Master:        masterclient.GetClient(),
		Auth:          authclient.New(),
		Risk:          riskclient.NewFromEnv(nil),
	}
}

func (s *HTTPSources) Department(ctx context.Context, id string) (*masterclient.DepartmentDetail, error) {
	return s.Master.GetDepartment(ctx, s.Authorization, id)
}

func (s *HTTPSources) Location(ctx context.Context, id string) (*masterclient.Location, error) {
	return s.Master.GetLocation(ctx, s.Authorization, id)
}

func (s *HTTPSources) ActiveEmployees(ctx context.Context, departmentID string) ([]masterclient.Employee, error) {
	return s.Master.ListActiveDepartmentEmployees(ctx, s.Authorization, departmentID)
}

func (s *HTTPSources) EmployeeByCode(ctx context.Context, code string) (*masterclient.Employee, error) {
	return s.Master.GetEmployeeByCode(ctx, s.Authorization, code)
}

func (s *HTTPSources) Acceptance(ctx context.Context, codes []string) ([]authclient.Acceptance, error) {
	return s.Auth.AcceptanceSummary(ctx, s.Authorization, codes)
}

func (s *HTTPSources) PofCatalogue(ctx context.Context, version string) (*riskclient.PofCatalogue, error) {
	return s.Risk.GetPofCatalogue(ctx, s.Authorization, version)
}

func (s *HTTPSources) CosoSignals(ctx context.Context, q riskclient.CosoSignalQuery) (*riskclient.CosoSignals, error) {
	return s.Risk.GetCosoSignals(ctx, s.Authorization, q)
}

// Scope is what one proposal run measures
type Scope struct {
	DepartmentID   string
	DepartmentName string
	LocationID     string // "" = whole department
	LocationName   string
	Year           int
	Quarter        int
	PeriodStart    time.Time // UTC midnight
	PeriodEnd      time.Time // UTC midnight, inclusive
	AsOf           time.Time // UTC midnight, inclusive
	Settings       Settings
}

// PeriodBounds returns the calendar period of (year, quarter); quarter 0 is
// the whole year (fiscal year = calendar year)
func PeriodBounds(year, quarter int) (time.Time, time.Time) {
	if quarter < 1 || quarter > 4 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	start := time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 3, -1)
}

// AsOfDate is min(period end, today)
func AsOfDate(periodEnd, now time.Time) time.Time {
	y, m, d := now.UTC().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if today.Before(periodEnd) {
		return today
	}
	return periodEnd
}

func dayOf(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// maxDetailIDs bounds the record ids kept per signal for drill-down
const maxDetailIDs = 200

func capIDs(ids []string) []string {
	if len(ids) > maxDetailIDs {
		return ids[:maxDetailIDs]
	}
	return ids
}

// Collection is the result of one collection
type Collection struct {
	Signals    []Signal
	Principles []riskclient.CosoPrincipleMean
	RiskOK     bool
}

// Collect gathers every signal of the scope. A source that fails gives
// SOURCE_UNAVAILABLE for its signals; Collect itself fails only on a local
// database error. It never writes.
func Collect(ctx context.Context, db *gorm.DB, src Sources, sc Scope) (*Collection, error) {
	out := &Collection{}
	auditNote := ""
	if sc.LocationID != "" {
		auditNote = "department-wide, no branch data"
	}

	out.Signals = append(out.Signals, collectIntegrityPact(ctx, src, sc))
	out.Signals = append(out.Signals, collectUnitAccountability(ctx, src, sc))
	charter, err := collectCharter(db)
	if err != nil {
		return nil, err
	}
	out.Signals = append(out.Signals, charter)

	risk, principles, riskOK := collectRisk(ctx, src, sc)
	out.Signals = append(out.Signals, risk...)
	out.Principles, out.RiskOK = principles, riskOK

	audit, err := collectAuditSide(db, sc, auditNote)
	if err != nil {
		return nil, err
	}
	out.Signals = append(out.Signals, audit...)

	exec, err := collectExecutiveSummary(db, sc)
	if err != nil {
		return nil, err
	}
	out.Signals = append(out.Signals, exec, Signal{
		Key: ContextThresholds, Component: ComponentIC, SourceService: "audit", SourceRef: "ic_assessment_settings",
		State: StateOK, ScopeNote: "thresholds used by this proposal",
		Detail: map[string]interface{}{
			"lha_due_days":                 sc.Settings.LHADueDays,
			"management_response_due_days": sc.Settings.ManagementResponseDueDays,
			"source":                       sc.Settings.Source,
		},
	})
	return out, nil
}

// newSignal starts a signal from its definition
func newSignal(key, source, ref string) Signal {
	d, _ := SignalDefByKey(key)
	return Signal{Key: key, Component: d.Component, PrincipleHint: d.PrincipleHint, Weight: d.Weight,
		SourceService: source, SourceRef: ref, Detail: map[string]interface{}{}}
}

func unavailable(s Signal, err error) Signal {
	s.State = StateSourceUnavailable
	s.ValuePct, s.Numerator, s.Denominator = nil, nil, nil
	if err != nil {
		s.Detail["error"] = err.Error()
	}
	return s
}

func withRatio(s Signal, num, den int) Signal {
	s.Numerator, s.Denominator = intPtr(num), intPtr(den)
	if den <= 0 {
		s.State = StateNoData
		s.ValuePct = nil
		return s
	}
	s.State = StateOK
	s.ValuePct = pct(num, den)
	return s
}

// ce.integrity_pact: linked users of the department's active employees with
// an accepted, unexpired SYSTEM agreement at the current version / linked users
func collectIntegrityPact(ctx context.Context, src Sources, sc Scope) Signal {
	s := newSignal(SignalCEIntegrityPact, "master,auth",
		"GET /employees?department_id=&is_active=true; GET /confidentiality/acceptance-summary")
	if sc.LocationID != "" {
		s.ScopeNote = "department-wide, no branch data"
	}
	emps, err := src.ActiveEmployees(ctx, sc.DepartmentID)
	if err != nil {
		s.SourceService = "master"
		return unavailable(s, err)
	}
	s.Detail["active_employees"] = len(emps)
	if len(emps) == 0 {
		s.State = StateNoData
		s.Detail["reason"] = "no active employees in the department"
		return s
	}
	codes := make([]string, 0, len(emps))
	for _, e := range emps {
		if c := strings.TrimSpace(e.EmployeeCode); c != "" {
			codes = append(codes, c)
		}
	}
	rows, err := src.Acceptance(ctx, codes)
	if err != nil {
		s.SourceService = "auth"
		return unavailable(s, err)
	}
	linked, accepted := 0, 0
	var notAccepted []string
	for _, r := range rows {
		if !r.HasUser {
			continue
		}
		linked++
		if r.Accepted {
			accepted++
		} else {
			notAccepted = append(notAccepted, r.EmployeeCode)
		}
	}
	s.Detail["linked_users"] = linked
	s.Detail["without_user"] = len(codes) - linked
	s.Detail["not_accepted_employee_codes"] = capIDs(notAccepted)
	s = withRatio(s, accepted, linked)
	if linked == 0 {
		s.Detail["reason"] = "no employee of the department has a login account"
	}
	return s
}

// ce.unit_accountability: 100 if the department is active and has a PIC
func collectUnitAccountability(ctx context.Context, src Sources, sc Scope) Signal {
	s := newSignal(SignalCEUnitAccountability, "master", "GET /departments/:id")
	d, err := src.Department(ctx, sc.DepartmentID)
	if err != nil {
		return unavailable(s, err)
	}
	ok := d.IsActive && d.HasPIC()
	s.Detail["is_active"] = d.IsActive
	s.Detail["has_pic"] = d.HasPIC()
	n := 0
	if ok {
		n = 1
	}
	return withRatio(s, n, 1)
}

// ce.audit_charter: 100 if an active charter exists (entity-level)
func collectCharter(db *gorm.DB) (Signal, error) {
	s := newSignal(SignalCEAuditCharter, "audit", "audit_charters.is_active")
	s.ScopeNote = "entity-level"
	var n int64
	if err := db.Model(&models.AuditCharter{}).Where("is_active = ?", true).Count(&n).Error; err != nil {
		return s, fmt.Errorf("count active charters: %w", err)
	}
	s.Detail["active_charters"] = n
	v := 0
	if n > 0 {
		v = 1
	}
	return withRatio(s, v, 1), nil
}

var riskSignalKeys = []string{SignalRAQuarterlyCoverage, SignalRAAboveAppetite, SignalCARCMEffectiveness}

// collectRisk reads the RA / CA signals from risk-service GET /rcm/coso-signals
func collectRisk(ctx context.Context, src Sources, sc Scope) ([]Signal, []riskclient.CosoPrincipleMean, bool) {
	ans, err := src.CosoSignals(ctx, riskclient.CosoSignalQuery{
		DepartmentID: sc.DepartmentID, Year: sc.Year, Quarter: sc.Quarter,
		LocationID: sc.LocationID, AsOf: sc.AsOf.Format("2006-01-02"),
	})
	out := make([]Signal, 0, len(riskSignalKeys))
	for _, key := range riskSignalKeys {
		s := newSignal(key, "risk", "GET /rcm/coso-signals")
		if err != nil {
			out = append(out, unavailable(s, err))
			continue
		}
		r := ans.Signal(key)
		if r == nil {
			out = append(out, unavailable(s, fmt.Errorf("signal missing in the risk-service answer")))
			continue
		}
		for k, v := range r.Detail {
			s.Detail[k] = v
		}
		s.ScopeNote = r.ScopeNote
		if sc.Quarter > 0 && s.ScopeNote == "" && key == SignalCARCMEffectiveness {
			s.ScopeNote = "annual data"
		}
		s.Numerator, s.Denominator = r.Numerator, r.Denominator
		switch strings.ToUpper(strings.TrimSpace(r.State)) {
		case StateOK:
			if r.ValuePct == nil {
				s.State = StateNoData
			} else {
				v := round2(clampPct(*r.ValuePct))
				s.State, s.ValuePct = StateOK, &v
			}
		case StateNoData:
			s.State = StateNoData
		case StateNotApplicable:
			s.State = StateNotApplicable
		default:
			s = unavailable(s, fmt.Errorf("unknown state %q from risk-service", r.State))
		}
		out = append(out, s)
	}
	if err != nil {
		return out, nil, false
	}
	return out, ans.PrincipleMeans(), true
}

func clampPct(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// finishPeriodLayouts are the date forms assignment_letters.finish_period is
// written in
var finishPeriodLayouts = []string{
	"2006-01-02", time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05",
	"02/01/2006", "2/1/2006", "02-01-2006", "2 January 2006", "02 January 2006",
	"January 2, 2006", "2 Jan 2006", "02 Jan 2006",
}

// ParseFinishPeriod parses a letter's finish_period
func ParseFinishPeriod(raw string) (time.Time, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range finishPeriodLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return dayOf(t), true
		}
	}
	if len(s) > 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			return dayOf(t), true
		}
	}
	return time.Time{}, false
}

func sameDepartment(text, name string) bool {
	a, b := strings.TrimSpace(text), strings.TrimSpace(name)
	return a != "" && b != "" && strings.EqualFold(a, b)
}

func inDays(t, start, end time.Time) bool {
	d := dayOf(t)
	return !d.Before(start) && !d.After(end)
}

// collectAuditSide computes the IC and MA signals from LHAs and ATRs
func collectAuditSide(db *gorm.DB, sc Scope, scopeNote string) ([]Signal, error) {
	var reports []models.AuditResultReport
	if err := db.Select("id", "assignment_letter_id", "report_number", "report_date", "status", "department").
		Where("UPPER(TRIM(status)) IN ?", models.LHAApprovedStatuses).
		Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("load approved LHAs: %w", err)
	}

	var inPeriod, allDept []models.AuditResultReport
	unmatched := 0
	for _, r := range reports {
		periodOK := r.ReportDate != nil && inDays(*r.ReportDate, sc.PeriodStart, sc.AsOf)
		if !sameDepartment(r.Department, sc.DepartmentName) {
			if periodOK {
				unmatched++
			}
			continue
		}
		allDept = append(allDept, r)
		if periodOK {
			inPeriod = append(inPeriod, r)
		}
	}

	// ic.lha_timeliness
	lha := newSignal(SignalICLHATimeliness, "audit", "audit_result_reports, assignment_letters.finish_period")
	lha.ScopeNote = scopeNote
	lha.Detail["threshold_days"] = sc.Settings.LHADueDays
	lha.Detail["lha_in_period"] = len(inPeriod)
	lha.Detail["lha_in_period_unmatched_department"] = unmatched
	onTime, comparable, unparseable, noLetter := 0, 0, 0, 0
	var late []string
	for _, r := range inPeriod {
		letter, ok := models.FindAssignmentLetter(db, r.AssignmentLetterID)
		if !ok {
			noLetter++
			continue
		}
		finish, ok := ParseFinishPeriod(letter.FinishPeriod)
		if !ok {
			unparseable++
			continue
		}
		comparable++
		if !dayOf(*r.ReportDate).After(finish.AddDate(0, 0, sc.Settings.LHADueDays)) {
			onTime++
		} else {
			late = append(late, r.ID.String())
		}
	}
	lha.Detail["unparseable_finish_period"] = unparseable
	lha.Detail["letter_not_found"] = noLetter
	lha.Detail["late_report_ids"] = capIDs(late)
	lha = withRatio(lha, onTime, comparable)

	// ATRs of the department's approved LHAs
	ids := make([]uuid.UUID, 0, len(allDept))
	periodIDs := map[uuid.UUID]bool{}
	for _, r := range allDept {
		ids = append(ids, r.ID)
	}
	for _, r := range inPeriod {
		periodIDs[r.ID] = true
	}
	var atrs []models.ActionTakenReport
	if len(ids) > 0 {
		if err := db.Select("id", "audit_result_report_id", "status", "due_date", "reviewed_at", "created_at", "action_plan_submitted_at").
			Where("audit_result_report_id IN ?", ids).
			Find(&atrs).Error; err != nil {
			return nil, fmt.Errorf("load action taken reports: %w", err)
		}
	}
	var ofPeriod []models.ActionTakenReport // ATRs of LHAs in the period, not cancelled
	var ofDept []models.ActionTakenReport   // ATRs of every department LHA, not cancelled
	for _, a := range atrs {
		if strings.EqualFold(strings.TrimSpace(a.Status), models.ATRStatusCancelled) {
			continue
		}
		ofDept = append(ofDept, a)
		if a.AuditResultReportID != nil && periodIDs[*a.AuditResultReportID] {
			ofPeriod = append(ofPeriod, a)
		}
	}

	mgmt := managementResponse(ofPeriod, sc, scopeNote)
	onTimeSig := completedOnTime(ofDept, sc, scopeNote)
	notOverdue := notOverdueSignal(ofPeriod, sc, scopeNote)
	return []Signal{lha, mgmt, onTimeSig, notOverdue}, nil
}

// ic.management_response. With action_plan_submitted_at data: ATRs whose
// action plan came within the threshold / ATRs with a plan time or still
// PLANNED past the threshold. Without it (history): response coverage, ATRs
// not PLANNED / all.
func managementResponse(atrs []models.ActionTakenReport, sc Scope, scopeNote string) Signal {
	s := newSignal(SignalICManagementResponse, "audit", "action_taken_reports")
	s.ScopeNote = scopeNote
	days := sc.Settings.ManagementResponseDueDays
	s.Detail["threshold_days"] = days
	s.Detail["atrs"] = len(atrs)
	timed := false
	for _, a := range atrs {
		if a.ActionPlanSubmittedAt != nil {
			timed = true
			break
		}
	}
	asOfEnd := sc.AsOf.AddDate(0, 0, 1)
	if timed {
		s.Detail["method"] = "response_time"
		num, den, skipped := 0, 0, 0
		var late []string
		for _, a := range atrs {
			deadline := a.CreatedAt.AddDate(0, 0, days)
			switch {
			case a.ActionPlanSubmittedAt != nil:
				den++
				if !a.ActionPlanSubmittedAt.After(deadline) {
					num++
				} else {
					late = append(late, a.ID.String())
				}
			case a.Status == models.ATRStatusPlanned && deadline.Before(asOfEnd):
				den++
				late = append(late, a.ID.String())
			default:
				skipped++
			}
		}
		s.Detail["skipped_without_plan_time"] = skipped
		s.Detail["late_atr_ids"] = capIDs(late)
		return withRatio(s, num, den)
	}
	s.Detail["method"] = "coverage"
	num := 0
	var planned []string
	for _, a := range atrs {
		if a.Status != models.ATRStatusPlanned {
			num++
		} else {
			planned = append(planned, a.ID.String())
		}
	}
	s.Detail["planned_atr_ids"] = capIDs(planned)
	return withRatio(s, num, len(atrs))
}

// ma.atr_completed_on_time: ATRs due in the period (up to the as-of date)
// completed with reviewed_at on or before the due date / ATRs due
func completedOnTime(atrs []models.ActionTakenReport, sc Scope, scopeNote string) Signal {
	s := newSignal(SignalMAATRCompletedOnTime, "audit", "action_taken_reports")
	s.ScopeNote = scopeNote
	num, den := 0, 0
	var missed []string
	for _, a := range atrs {
		if a.DueDate == nil || a.DueDate.IsZero() || !inDays(*a.DueDate, sc.PeriodStart, sc.AsOf) {
			continue
		}
		den++
		due := dayOf(*a.DueDate)
		if a.Status == models.ATRStatusCompleted && a.ReviewedAt != nil && !models.ATRToday(*a.ReviewedAt).After(due) {
			num++
		} else {
			missed = append(missed, a.ID.String())
		}
	}
	s.Detail["missed_atr_ids"] = capIDs(missed)
	return withRatio(s, num, den)
}

// ma.atr_not_overdue: 100 - overdue / open ATRs with a due date, at the as-of
// date. No open ATR with a due date -> NOT_APPLICABLE.
func notOverdueSignal(atrs []models.ActionTakenReport, sc Scope, scopeNote string) Signal {
	s := newSignal(SignalMAATRNotOverdue, "audit", "action_taken_reports")
	s.ScopeNote = scopeNote
	// Midday UTC of the as-of date is the same calendar day in Asia/Jakarta
	at := sc.AsOf.Add(12 * time.Hour)
	open, overdue := 0, 0
	var ids []string
	for _, a := range atrs {
		if models.IsATRClosed(a.Status) || a.DueDate == nil || a.DueDate.IsZero() {
			continue
		}
		open++
		if od, _ := models.ComputeATROverdue(a.DueDate, a.Status, at); od {
			overdue++
			ids = append(ids, a.ID.String())
		}
	}
	s.Detail["open_with_due_date"] = open
	s.Detail["overdue"] = overdue
	s.Detail["overdue_atr_ids"] = capIDs(ids)
	if open == 0 {
		s.State = StateNotApplicable
		return s
	}
	return withRatio(s, open-overdue, open)
}

// ctx.executive_summary: approved executive summary for the period (context)
func collectExecutiveSummary(db *gorm.DB, sc Scope) (Signal, error) {
	s := Signal{Key: ContextExecutiveSummary, Component: ComponentIC, SourceService: "audit",
		SourceRef: "executive_summaries", ScopeNote: "entity-level, not scored", Detail: map[string]interface{}{}}
	base := func() *gorm.DB {
		q := db.Model(&models.ExecutiveSummary{}).Where("tahun = ?", sc.Year)
		if sc.Quarter > 0 {
			q = q.Where("quarter = ?", sc.Quarter)
		}
		return q
	}
	var total, approved int64
	if err := base().Count(&total).Error; err != nil {
		return s, fmt.Errorf("count executive summaries: %w", err)
	}
	if err := base().Where("UPPER(TRIM(status)) = ?", "APPROVED").Count(&approved).Error; err != nil {
		return s, fmt.Errorf("count approved executive summaries: %w", err)
	}
	s.Detail["total"] = total
	s.Detail["approved"] = approved
	s.State = StateOK
	if total == 0 {
		s.State = StateNoData
	}
	return s, nil
}
