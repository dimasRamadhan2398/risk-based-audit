package activityplanform

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Status of a newly inserted RKA row and header.
const statusPlanned = "PLANNED"

// Page size limits of GET /activity-plans.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var yearRe = regexp.MustCompile(`(19|20|21)\d{2}`)

// parseYear reads the plan year stored as text ("2026", "FY 2026"); 0 if none.
func parseYear(s string) int {
	if y := yearRe.FindString(s); y != "" {
		n, _ := strconv.Atoi(y)
		return n
	}
	return 0
}

// uuidText is the id as text, "" for nil or the zero UUID (old rows store the
// zero UUID for "none").
func uuidText(id *uuid.UUID) string {
	if id == nil || *id == uuid.Nil {
		return ""
	}
	return id.String()
}

func strValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// pkatFields are the PKAT-owned values of a row, live or from the snapshot.
type pkatFields struct {
	object, category, schedule, supervisor string
	universe                               UniverseRef
	department                             DepartmentRef
	teamSize, mandays                      int
	source                                 string
}

// resolve returns the PKAT-owned fields of an RKA row: live from the PKAT row
// when the PKAT is live and still has the row, else the snapshot columns.
func resolve(a *models.AuditActivity, pkat *PKAT, pkatDeleted bool) pkatFields {
	if pkat != nil && !pkatDeleted {
		if r, _, ok := pkat.Row(strValue(a.AnnualActivityID)); ok {
			return pkatFields{
				object:     r.Name,
				category:   r.Category,
				schedule:   r.TimelineText,
				supervisor: r.SupervisorName,
				universe:   UniverseRef{ID: r.AuditUniverseID, Name: r.RiskName, RiskLevel: r.RiskLevel},
				department: DepartmentRef{ID: r.DepartmentID, Name: r.Department},
				teamSize:   r.AuditorCount,
				mandays:    r.TotalMandays,
				source:     "pkat",
			}
		}
	}
	return pkatFields{
		object:     a.Title,
		category:   a.AuditType,
		universe:   UniverseRef{ID: uuidText(a.AuditUniverseID)},
		department: DepartmentRef{ID: uuidText(a.TargetUnitID)},
		teamSize:   a.TeamSize,
		mandays:    a.TotalMandays,
		source:     "snapshot",
	}
}

func notFound() error {
	return Fail(CodeNotFound, http.StatusNotFound, "activity plan not found", Details{})
}

// loadPlan reads a live activity plan, 404 NOT_FOUND when missing.
func loadPlan(tx *gorm.DB, id uuid.UUID, lock bool) (*models.ActivityPlan, error) {
	q := tx
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var plan models.ActivityPlan
	if err := q.First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFound()
		}
		return nil, err
	}
	return &plan, nil
}

// planRows reads the live rows of an RKA in display order.
func planRows(tx *gorm.DB, planID uuid.UUID) ([]models.AuditActivity, error) {
	var rows []models.AuditActivity
	err := tx.Where("activity_plan_id = ?", planID).
		Order("sort_order ASC").Order("created_at ASC").Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func isRKA(p *models.ActivityPlan) bool { return p.FormVersion == models.ActivityPlanFormRKA }

func attachmentsOrEmpty(a []models.AnnualAuditAttachment) []models.AnnualAuditAttachment {
	if a == nil {
		return []models.AnnualAuditAttachment{}
	}
	return a
}

// Get builds GET /activity-plans/:id.
func Get(db *gorm.DB, id uuid.UUID) (*Detail, error) {
	plan, err := loadPlan(db, id, false)
	if err != nil {
		return nil, err
	}
	d := &Detail{
		ID:                   plan.ID.String(),
		FormVersion:          plan.FormVersion,
		Title:                plan.PlanTitle,
		Status:               plan.Status,
		Rows:                 []Row{},
		AttachmentCategory:   plan.AttachmentCategory,
		Attachments:          attachmentsOrEmpty(plan.Attachments),
		AttachmentUploadedBy: plan.AttachmentUploadedBy,
		AttachmentUploadDate: plan.AttachmentUploadDate,
		CreatedBy:            plan.CreatedBy,
		CreatedAt:            plan.CreatedAt,
		UpdatedAt:            plan.UpdatedAt,
	}
	if !isRKA(plan) {
		d.Year = parseYear(plan.PlanYear)
		d.Totals.RowCount = len(plan.PlannedActivities)
		acts := plan.PlannedActivities
		if acts == nil {
			acts = []models.PlannedActivity{}
		}
		auditors := plan.ResourceAuditors
		if auditors == nil {
			auditors = []models.ResourceAuditor{}
		}
		d.Legacy = &Legacy{
			PlanTitle:         plan.PlanTitle,
			PlanYear:          plan.PlanYear,
			PlanPeriodStart:   plan.PlanPeriodStart,
			PlanPeriodEnd:     plan.PlanPeriodEnd,
			Department:        plan.Department,
			CreationDate:      plan.CreationDate,
			PlannedActivities: acts,
			ResourceAuditors:  auditors,
			Budget:            plan.Budget,
			Review:            plan.Review,
		}
		return d, nil
	}

	d.ActivityType = plan.ActivityType
	d.Year = parseYear(plan.PlanYear)
	var pkat *PKAT
	pkatDeleted := false
	if s := uuidText(plan.AnnualPlanID); s != "" {
		d.AnnualPlanID = &s
		pkats, deleted, err := loadPKATs(db, []uuid.UUID{*plan.AnnualPlanID})
		if err != nil {
			return nil, err
		}
		if p, ok := pkats[*plan.AnnualPlanID]; ok {
			pkat, pkatDeleted = p, deleted[*plan.AnnualPlanID]
			d.AnnualPlan = &AnnualPlanRef{
				ID: p.Plan.ID.String(), Code: p.Plan.Code, Title: p.Plan.Title,
				Year: p.Plan.Year, Status: p.Plan.Status, IsActive: p.Plan.IsActive,
			}
			d.Year = p.Plan.Year
		}
	}

	rows, err := planRows(db, plan.ID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	inUse, err := InUse(db, ids)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		a := &rows[i]
		f := resolve(a, pkat, pkatDeleted)
		row := Row{
			ID:                     a.ID.String(),
			No:                     i + 1,
			AnnualActivityID:       strValue(a.AnnualActivityID),
			AssignmentCode:         a.ProjectCode,
			AssignmentCodeEditable: !inUse[a.ID],
			AuditObject:            f.object,
			AuditCategory:          f.category,
			AuditKind:              a.AuditKind,
			AuditUniverse:          f.universe,
			Department:             f.department,
			Justification:          Split(a.Justification),
			Objectives:             Split(a.AuditPurpose),
			Focus:                  Split(a.AuditFocus),
			Schedule:               f.schedule,
			TeamSize:               f.teamSize,
			TotalMandays:           f.mandays,
			PkatSupervisorName:     f.supervisor,
			Status:                 a.Status,
			InUse:                  inUse[a.ID],
			Source:                 f.source,
		}
		if leader := uuidText(a.TeamLeaderID); leader != "" {
			row.TeamLeader = &TeamLeaderRef{EmployeeID: leader, Name: a.TeamLeaderName}
		}
		d.Rows = append(d.Rows, row)
		d.Totals.RowCount++
		d.Totals.TeamSize += f.teamSize
		d.Totals.TotalMandays += f.mandays
	}
	return d, nil
}

// ListParams are the GET /activity-plans query parameters.
type ListParams struct {
	Page          int
	PageSize      int
	Search        string
	Year          *int
	ActivityType  string
	AnnualPlanID  *uuid.UUID
	IncludeLegacy bool
}

// ListResult is one page of GET /activity-plans.
type ListResult struct {
	Items    []ListItem
	Page     int
	PageSize int
	Total    int64
}

// List returns one page of activity plans, sorted by year (desc), activity
// type, then newest first. All conditions use fixed column names; request
// values are bound as parameters.
func List(db *gorm.DB, p ListParams) (*ListResult, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}

	q := db.Model(&models.ActivityPlan{})
	if s := strings.TrimSpace(p.Search); s != "" {
		q = q.Where("LOWER(plan_title) LIKE ?", "%"+strings.ToLower(s)+"%")
	}
	if p.Year != nil {
		q = q.Where("plan_year = ?", strconv.Itoa(*p.Year))
	}
	if t := strings.TrimSpace(p.ActivityType); t != "" {
		q = q.Where("activity_type = ?", t)
	}
	if p.AnnualPlanID != nil {
		q = q.Where("annual_plan_id = ?", *p.AnnualPlanID)
	}
	if !p.IncludeLegacy {
		q = q.Where("form_version = ?", models.ActivityPlanFormRKA)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var plans []models.ActivityPlan
	if err := q.Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		{Column: clause.Column{Name: "plan_year"}, Desc: true},
		{Column: clause.Column{Name: "activity_type"}},
		{Column: clause.Column{Name: "created_at"}, Desc: true},
		{Column: clause.Column{Name: "id"}},
	}}).Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize).Find(&plans).Error; err != nil {
		return nil, err
	}

	// PKATs and rows of the RKAs on this page: two queries.
	var pkatIDs, rkaIDs []uuid.UUID
	for i := range plans {
		if isRKA(&plans[i]) {
			rkaIDs = append(rkaIDs, plans[i].ID)
			if uuidText(plans[i].AnnualPlanID) != "" {
				pkatIDs = append(pkatIDs, *plans[i].AnnualPlanID)
			}
		}
	}
	pkats, deleted, err := loadPKATs(db, pkatIDs)
	if err != nil {
		return nil, err
	}
	rowsByPlan := map[uuid.UUID][]models.AuditActivity{}
	if len(rkaIDs) > 0 {
		var rows []models.AuditActivity
		if err := db.Select("id", "activity_plan_id", "annual_activity_id", "title", "audit_type",
			"audit_universe_id", "target_unit_id", "team_size", "total_mandays").
			Where("activity_plan_id IN ?", rkaIDs).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			if r.ActivityPlanID != nil {
				rowsByPlan[*r.ActivityPlanID] = append(rowsByPlan[*r.ActivityPlanID], r)
			}
		}
	}

	items := make([]ListItem, 0, len(plans))
	for i := range plans {
		plan := &plans[i]
		item := ListItem{
			ID:              plan.ID.String(),
			FormVersion:     plan.FormVersion,
			Year:            parseYear(plan.PlanYear),
			Title:           plan.PlanTitle,
			Status:          plan.Status,
			AttachmentCount: len(plan.Attachments),
			CreatedBy:       plan.CreatedBy,
			CreatedAt:       plan.CreatedAt,
			UpdatedAt:       plan.UpdatedAt,
		}
		if !isRKA(plan) {
			item.Totals.RowCount = len(plan.PlannedActivities)
			items = append(items, item)
			continue
		}
		item.ActivityType = plan.ActivityType
		var pkat *PKAT
		if uuidText(plan.AnnualPlanID) != "" {
			if pk, ok := pkats[*plan.AnnualPlanID]; ok {
				pkat = pk
				item.Year = pk.Plan.Year
				item.AnnualPlan = &AnnualPlanSummaryRef{
					ID: pk.Plan.ID.String(), Code: pk.Plan.Code, Title: pk.Plan.Title,
					Status: pk.Plan.Status, IsActive: pk.Plan.IsActive,
				}
			}
		}
		for j := range rowsByPlan[plan.ID] {
			f := resolve(&rowsByPlan[plan.ID][j], pkat, pkat != nil && deleted[pkat.Plan.ID])
			item.Totals.RowCount++
			item.Totals.TeamSize += f.teamSize
			item.Totals.TotalMandays += f.mandays
		}
		items = append(items, item)
	}
	return &ListResult{Items: items, Page: p.Page, PageSize: p.PageSize, Total: total}, nil
}

// Options lists the PKAT rows that can be added to an RKA of activityType
// (GET /activity-plans/annual-activity-options). excludePlanID is the RKA being
// edited: rows it holds count as available.
func Options(db *gorm.DB, annualPlanID uuid.UUID, activityType string, excludePlanID *uuid.UUID) ([]Option, error) {
	if !ValidActivityType(activityType) {
		return nil, invalid("activity_type", "activity_type must be one of "+strings.Join(models.ActivityPlanActivityTypes, ", "))
	}
	pkat, err := LoadPKAT(db, annualPlanID, false)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, key := range pkat.Order {
		r, _, _ := pkat.Row(key)
		if CategoryMatches(activityType, r.Category) {
			keys = append(keys, key)
		}
	}
	out := make([]Option, 0, len(keys))
	if len(keys) == 0 {
		return out, nil
	}

	// Live rows linked to these PKAT rows, and the RKAs holding them.
	var linked []models.AuditActivity
	if err := db.Select("id", "annual_activity_id", "activity_plan_id", "project_code", "status",
		"justification", "audit_purpose", "audit_focus").
		Where("annual_activity_id IN ?", keys).Find(&linked).Error; err != nil {
		return nil, err
	}
	byKey := map[string]*models.AuditActivity{}
	var holderIDs []uuid.UUID
	for i := range linked {
		byKey[RowKey(strValue(linked[i].AnnualActivityID))] = &linked[i]
		if linked[i].ActivityPlanID != nil {
			holderIDs = append(holderIDs, *linked[i].ActivityPlanID)
		}
	}
	holders := map[uuid.UUID]models.ActivityPlan{}
	if len(holderIDs) > 0 {
		var plans []models.ActivityPlan
		if err := db.Select("id", "plan_title").Where("id IN ?", holderIDs).Find(&plans).Error; err != nil {
			return nil, err
		}
		for _, p := range plans {
			holders[p.ID] = p
		}
	}

	for _, key := range keys {
		r, idx, _ := pkat.Row(key)
		itemNumber := strings.TrimSpace(r.ItemNumber)
		if itemNumber == "" {
			itemNumber = strconv.Itoa(idx + 1)
		}
		opt := Option{
			AnnualActivityID:   key,
			ItemNumber:         itemNumber,
			AuditObject:        r.Name,
			AuditCategory:      r.Category,
			AuditUniverse:      UniverseRef{ID: r.AuditUniverseID, Name: r.RiskName, RiskLevel: r.RiskLevel},
			Department:         DepartmentRef{ID: r.DepartmentID, Name: r.Department},
			Schedule:           r.TimelineText,
			TeamSize:           r.AuditorCount,
			TotalMandays:       r.TotalMandays,
			PkatSupervisorName: r.SupervisorName,
			Available:          true,
			Prefill: Prefill{
				AssignmentCode: strings.TrimSpace(r.ActivityCode),
				Justification:  Split(r.Justification),
				Objectives:     Split(r.AuditObjective),
				Focus:          Split(r.AuditFocus),
			},
		}
		if a, ok := byKey[key]; ok {
			holder, held := models.ActivityPlan{}, false
			if a.ActivityPlanID != nil {
				holder, held = holders[*a.ActivityPlanID]
			}
			own := held && excludePlanID != nil && holder.ID == *excludePlanID
			if held && !own {
				opt.Available = false
				opt.UsedBy = &UsedByRef{
					ActivityPlanID:    holder.ID.String(),
					ActivityPlanTitle: holder.PlanTitle,
					AssignmentCode:    a.ProjectCode,
				}
			} else {
				// A legacy row (or the edited RKA's own row): the save adopts or
				// keeps it, so its values come first.
				opt.ExistingRow = &ExistingRowRef{ID: a.ID.String(), AssignmentCode: a.ProjectCode, Status: a.Status}
				opt.Prefill.AssignmentCode = a.ProjectCode
				if v := Split(a.Justification); len(v) > 0 {
					opt.Prefill.Justification = v
				}
				if v := Split(a.AuditPurpose); len(v) > 0 {
					opt.Prefill.Objectives = v
				}
				if v := Split(a.AuditFocus); len(v) > 0 {
					opt.Prefill.Focus = v
				}
			}
		}
		out = append(out, opt)
	}
	return out, nil
}
