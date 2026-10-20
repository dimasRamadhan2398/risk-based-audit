package activityplanform

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"audit-service/models"
	"audit-service/pkg/activitycode"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Limits of the POST/PUT body (contract 6.4).
const (
	MaxTitleLen     = 255
	MaxRows         = 200
	MaxAuditKindLen = 100
	MaxCodeLen      = 50
	MaxLeaderLen    = 255
)

var codePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// RowInput is one row of the POST/PUT /activity-plans body.
type RowInput struct {
	ID                   *string  `json:"id"`
	AnnualActivityID     string   `json:"annualActivityId"`
	AssignmentCode       string   `json:"assignmentCode"`
	AuditKind            string   `json:"auditKind"`
	Justification        []string `json:"justification"`
	Objectives           []string `json:"objectives"`
	Focus                []string `json:"focus"`
	TeamLeaderEmployeeID *string  `json:"teamLeaderEmployeeId"`
	TeamLeaderName       string   `json:"teamLeaderName"`
}

// SaveInput is the POST/PUT /activity-plans body. Rows is a pointer so that a
// body without "rows" (the old client shape) can be told apart and rejected.
// Attachments nil (absent) keeps the stored list on PUT.
type SaveInput struct {
	AnnualPlanID       string                          `json:"annualPlanId"`
	ActivityType       string                          `json:"activityType"`
	Title              string                          `json:"title"`
	UpdatedAt          *string                         `json:"updatedAt"`
	Rows               *[]RowInput                     `json:"rows"`
	AttachmentCategory *string                         `json:"attachmentCategory"`
	Attachments        *[]models.AnnualAuditAttachment `json:"attachments"`
}

// Create stores a new RKA and returns its id. actor is the signed-in user's
// name (created_by, attachment uploader).
func Create(db *gorm.DB, in SaveInput, actor string, now time.Time) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		id, err = save(tx, nil, in, actor, now)
		return err
	})
	return id, err
}

// Update replaces an RKA's header and rows. All-or-nothing.
func Update(db *gorm.DB, id uuid.UUID, in SaveInput, actor string, now time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		plan, err := loadPlan(tx, id, true)
		if err != nil {
			return err
		}
		if !isRKA(plan) {
			return Fail(CodeActivityPlanLegacyReadonly, http.StatusConflict,
				"old-format activity plans cannot be edited", Details{ActivityPlanID: plan.ID.String()})
		}
		if in.UpdatedAt != nil && strings.TrimSpace(*in.UpdatedAt) != "" {
			t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*in.UpdatedAt))
			if err != nil {
				return invalid("updatedAt", "updatedAt must be an RFC 3339 timestamp")
			}
			if t.Unix() != plan.UpdatedAt.Unix() {
				return Fail(CodeActivityPlanStale, http.StatusConflict,
					"the activity plan was changed by someone else; reload it", Details{ActivityPlanID: plan.ID.String()})
			}
		}
		_, err = save(tx, plan, in, actor, now)
		return err
	})
}

// rowPlan is one validated request row and what the save does with it.
type rowPlan struct {
	idx        int // 0-based position in the request
	key        string
	givenID    *uuid.UUID
	code       string // as sent, trimmed ("" = generate / keep)
	kind       string
	just       []string
	obj        []string
	focus      []string
	leaderID   *uuid.UUID
	leaderName string
	pkatRow    *models.AnnualAuditActivity

	target  *models.AuditActivity // existing row kept, adopted or restored; nil = insert
	restore bool
}

func (r *rowPlan) no() int { return r.idx + 1 }
func (r *rowPlan) field(name string) string {
	return fmt.Sprintf("rows[%d].%s", r.idx, name)
}

// validateShape checks the header and rows without touching the database.
func validateShape(in SaveInput) (title, activityType string, pkatID uuid.UUID, rows []*rowPlan, err error) {
	title = strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLen {
		return "", "", uuid.Nil, nil, invalid("title", fmt.Sprintf("title is required (at most %d characters)", MaxTitleLen))
	}
	activityType = strings.TrimSpace(in.ActivityType)
	if !ValidActivityType(activityType) {
		return "", "", uuid.Nil, nil, invalid("activityType",
			"activityType must be one of "+strings.Join(models.ActivityPlanActivityTypes, ", "))
	}
	pkatID, perr := uuid.Parse(strings.TrimSpace(in.AnnualPlanID))
	if perr != nil || pkatID == uuid.Nil {
		return "", "", uuid.Nil, nil, invalid("annualPlanId", "annualPlanId must be the id of an annual audit plan (PKAT)")
	}
	if in.Rows == nil {
		return "", "", uuid.Nil, nil, invalid("rows", "rows is required")
	}
	if n := len(*in.Rows); n == 0 || n > MaxRows {
		return "", "", uuid.Nil, nil, invalid("rows", fmt.Sprintf("rows must have between 1 and %d entries", MaxRows))
	}

	seenKey := map[string]bool{}
	seenID := map[uuid.UUID]bool{}
	seenCode := map[string]bool{}
	for i, in := range *in.Rows {
		r := &rowPlan{idx: i}
		r.key = RowKey(in.AnnualActivityID)
		if r.key == "" {
			return fail(invalidRow(r.no(), r.field("annualActivityId"), fmt.Sprintf("assignment %d: annualActivityId is required", r.no())))
		}
		if seenKey[r.key] {
			return fail(invalidRow(r.no(), r.field("annualActivityId"), fmt.Sprintf("assignment %d: this PKAT activity is already in the form", r.no())))
		}
		seenKey[r.key] = true

		if in.ID != nil && strings.TrimSpace(*in.ID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(*in.ID))
			if err != nil || seenID[id] {
				return fail(invalidRow(r.no(), r.field("id"), fmt.Sprintf("assignment %d: invalid row id", r.no())))
			}
			seenID[id] = true
			r.givenID = &id
		}

		r.code = strings.TrimSpace(in.AssignmentCode)
		if r.code != "" {
			if utf8.RuneCountInString(r.code) > MaxCodeLen || !codePattern.MatchString(r.code) {
				return fail(invalidRow(r.no(), r.field("assignmentCode"),
					fmt.Sprintf("assignment %d: the code may use letters, digits, . _ / - (at most %d characters)", r.no(), MaxCodeLen)))
			}
			if seenCode[r.code] {
				return fail(Fail(CodeAssignmentCodeTaken, http.StatusConflict,
					fmt.Sprintf("assignment %d: code %s is used twice in this form", r.no(), r.code),
					Details{Field: r.field("assignmentCode"), Row: r.no(), AssignmentCode: r.code}))
			}
			seenCode[r.code] = true
		}

		r.kind = strings.TrimSpace(in.AuditKind)
		if r.kind == "" || utf8.RuneCountInString(r.kind) > MaxAuditKindLen {
			return fail(invalidRow(r.no(), r.field("auditKind"),
				fmt.Sprintf("assignment %d: auditKind is required (at most %d characters)", r.no(), MaxAuditKindLen)))
		}

		r.just, r.obj, r.focus = Normalize(in.Justification), Normalize(in.Objectives), Normalize(in.Focus)
		if err := checkList(r.no(), r.field("justification"), "justification", r.just); err != nil {
			return fail(err)
		}
		if err := checkList(r.no(), r.field("objectives"), "objectives", r.obj); err != nil {
			return fail(err)
		}
		if err := checkList(r.no(), r.field("focus"), "focus", r.focus); err != nil {
			return fail(err)
		}

		if in.TeamLeaderEmployeeID != nil && strings.TrimSpace(*in.TeamLeaderEmployeeID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(*in.TeamLeaderEmployeeID))
			if err != nil || id == uuid.Nil {
				return fail(invalidRow(r.no(), r.field("teamLeaderEmployeeId"), fmt.Sprintf("assignment %d: invalid team leader id", r.no())))
			}
			r.leaderID = &id
			r.leaderName = strings.TrimSpace(in.TeamLeaderName)
			if utf8.RuneCountInString(r.leaderName) > MaxLeaderLen {
				return fail(invalidRow(r.no(), r.field("teamLeaderName"),
					fmt.Sprintf("assignment %d: teamLeaderName is at most %d characters", r.no(), MaxLeaderLen)))
			}
		}
		rows = append(rows, r)
	}
	return title, activityType, pkatID, rows, nil
}

func fail(err error) (string, string, uuid.UUID, []*rowPlan, error) {
	return "", "", uuid.Nil, nil, err
}

// save is the shared body of Create (existing nil) and Update, inside tx.
func save(tx *gorm.DB, existing *models.ActivityPlan, in SaveInput, actor string, now time.Time) (uuid.UUID, error) {
	creating := existing == nil
	title, activityType, pkatID, rows, err := validateShape(in)
	if err != nil {
		return uuid.Nil, err
	}
	if !creating && (existing.AnnualPlanID == nil || *existing.AnnualPlanID != pkatID) {
		return uuid.Nil, invalid("annualPlanId", "the annual audit plan (PKAT) of an activity plan cannot change")
	}

	pkat, err := LoadPKAT(tx, pkatID, true)
	if err != nil {
		return uuid.Nil, err
	}
	if creating && !pkat.Plan.IsActive {
		return uuid.Nil, Fail(CodeAnnualPlanInactive, http.StatusConflict,
			"the annual audit plan (PKAT) is deactivated", Details{Field: "annualPlanId"})
	}

	// I3: every row is a row of this PKAT whose category maps to the type.
	for _, r := range rows {
		row, _, ok := pkat.Row(r.key)
		if !ok {
			return uuid.Nil, Fail(CodeAnnualActivityNotFound, http.StatusBadRequest,
				fmt.Sprintf("assignment %d: the PKAT activity does not exist in this PKAT", r.no()),
				Details{Field: r.field("annualActivityId"), Row: r.no()})
		}
		if !CategoryMatches(activityType, row.Category) {
			return uuid.Nil, Fail(CodeAnnualActivityTypeMismatch, http.StatusConflict,
				fmt.Sprintf("assignment %d: PKAT category %q does not match %q", r.no(), row.Category, activityType),
				Details{Field: r.field("annualActivityId"), Row: r.no()})
		}
		r.pkatRow = row
	}

	// I2: one live RKA per (PKAT, activity type).
	dupQ := tx.Model(&models.ActivityPlan{}).
		Where("annual_plan_id = ? AND activity_type = ? AND form_version = ?", pkatID, activityType, models.ActivityPlanFormRKA)
	if !creating {
		dupQ = dupQ.Where("id <> ?", existing.ID)
	}
	var dup models.ActivityPlan
	if err := dupQ.Select("id", "plan_title").Limit(1).Find(&dup).Error; err != nil {
		return uuid.Nil, err
	}
	if dup.ID != uuid.Nil {
		return uuid.Nil, duplicatePlan(dup.ID.String(), dup.PlanTitle)
	}

	planID := uuid.New()
	var current []models.AuditActivity
	if !creating {
		planID = existing.ID
		if current, err = planRows(tx, planID); err != nil {
			return uuid.Nil, err
		}
	}
	if err := resolveTargets(tx, planID, creating, rows, current); err != nil {
		return uuid.Nil, err
	}

	// In-use state of the rows that already exist (kept, adopted, removed).
	targeted := map[uuid.UUID]bool{}
	var checkIDs []uuid.UUID
	for _, r := range rows {
		if r.target != nil && !r.restore {
			targeted[r.target.ID] = true
			checkIDs = append(checkIDs, r.target.ID)
		}
	}
	var removed []models.AuditActivity
	for i, a := range current {
		if !targeted[a.ID] {
			removed = append(removed, current[i])
			checkIDs = append(checkIDs, a.ID)
		}
	}
	inUse, err := InUse(tx, checkIDs)
	if err != nil {
		return uuid.Nil, err
	}
	for i, a := range removed {
		if inUse[a.ID] {
			no := a.SortOrder
			if no <= 0 {
				no = i + 1
			}
			return uuid.Nil, Fail(CodeActivityRowInUse, http.StatusConflict,
				fmt.Sprintf("assignment %d (%s) has started and cannot be removed", no, a.ProjectCode),
				Details{Row: no, AssignmentCode: a.ProjectCode, ActivityPlanID: planID.String()})
		}
	}

	// Kode Penugasan: locked when in use, unique across all rows.
	for _, r := range rows {
		if r.target != nil {
			if r.code == "" || r.code == r.target.ProjectCode {
				r.code = r.target.ProjectCode
				continue
			}
			if inUse[r.target.ID] {
				return uuid.Nil, Fail(CodeAssignmentCodeLocked, http.StatusConflict,
					fmt.Sprintf("assignment %d: the code cannot change after the engagement started", r.no()),
					Details{Field: r.field("assignmentCode"), Row: r.no(), AssignmentCode: r.target.ProjectCode})
			}
		}
		if r.code == "" {
			continue // generated below
		}
		q := tx.Unscoped().Model(&models.AuditActivity{}).Where("project_code = ?", r.code)
		if r.target != nil {
			q = q.Where("id <> ?", r.target.ID)
		}
		var n int64
		if err := q.Count(&n).Error; err != nil {
			return uuid.Nil, err
		}
		if n > 0 {
			return uuid.Nil, Fail(CodeAssignmentCodeTaken, http.StatusConflict,
				fmt.Sprintf("assignment %d: code %s is already used", r.no(), r.code),
				Details{Field: r.field("assignmentCode"), Row: r.no(), AssignmentCode: r.code})
		}
	}

	// Writes. Removed rows first: adopted rows (they existed before this RKA)
	// are detached, rows the RKA created are soft-deleted (re-adding the PKAT
	// row later restores them with their code).
	planCreatedAt := now
	if !creating {
		planCreatedAt = existing.CreatedAt
	}
	for _, a := range removed {
		if err := releaseRow(tx, a, planCreatedAt); err != nil {
			return uuid.Nil, err
		}
	}

	if err := writeHeader(tx, existing, planID, pkat, activityType, title, in, actor, now); err != nil {
		return uuid.Nil, err
	}

	// Rows with a code first, so generated numbers start above any code just
	// typed in this request.
	for _, r := range rows {
		if r.code != "" {
			if err := writeRow(tx, r, planID, pkat); err != nil {
				return uuid.Nil, err
			}
		}
	}
	for _, r := range rows {
		if r.code == "" {
			code, err := activitycode.Next(tx, r.pkatRow.Category, pkat.Plan.Year)
			if err != nil {
				return uuid.Nil, err
			}
			r.code = code
			if err := writeRow(tx, r, planID, pkat); err != nil {
				return uuid.Nil, err
			}
		}
	}
	return planID, nil
}

func duplicatePlan(id, title string) error {
	return Fail(CodeActivityPlanDuplicate, http.StatusConflict,
		"an activity plan of this type already exists for this PKAT",
		Details{Field: "activityType", ActivityPlanID: id, ActivityPlanTitle: title})
}

// resolveTargets decides, for every request row, which stored row it writes:
// the RKA's own row (by id, else by PKAT row), an adopted legacy row, a
// restored soft-deleted row of this RKA, or none (insert).
func resolveTargets(tx *gorm.DB, planID uuid.UUID, creating bool, rows []*rowPlan, current []models.AuditActivity) error {
	byID := map[uuid.UUID]*models.AuditActivity{}
	byKey := map[string]*models.AuditActivity{}
	for i := range current {
		byID[current[i].ID] = &current[i]
		byKey[RowKey(strValue(current[i].AnnualActivityID))] = &current[i]
	}

	var lookup []string
	for _, r := range rows {
		switch {
		case r.givenID != nil:
			a, ok := byID[*r.givenID]
			if !ok {
				return invalidRow(r.no(), r.field("id"), fmt.Sprintf("assignment %d: the row does not belong to this activity plan", r.no()))
			}
			if RowKey(strValue(a.AnnualActivityID)) != r.key {
				return invalidRow(r.no(), r.field("annualActivityId"),
					fmt.Sprintf("assignment %d: the PKAT activity of an existing row cannot change; remove the row and add a new one", r.no()))
			}
			r.target = a
		case byKey[r.key] != nil:
			r.target = byKey[r.key]
		default:
			lookup = append(lookup, r.key)
		}
	}
	if len(lookup) == 0 {
		return nil
	}

	// Live rows elsewhere that hold these PKAT rows, and their RKAs.
	var linked []models.AuditActivity
	if err := tx.Where("annual_activity_id IN ?", lookup).Find(&linked).Error; err != nil {
		return err
	}
	held := map[string]*models.AuditActivity{}
	var holderIDs []uuid.UUID
	for i := range linked {
		held[RowKey(strValue(linked[i].AnnualActivityID))] = &linked[i]
		if linked[i].ActivityPlanID != nil {
			holderIDs = append(holderIDs, *linked[i].ActivityPlanID)
		}
	}
	holders := map[uuid.UUID]models.ActivityPlan{}
	if len(holderIDs) > 0 {
		var plans []models.ActivityPlan
		if err := tx.Select("id", "plan_title").Where("id IN ?", holderIDs).Find(&plans).Error; err != nil {
			return err
		}
		for _, p := range plans {
			holders[p.ID] = p
		}
	}

	var restoreKeys []string
	for _, r := range rows {
		if r.target != nil {
			continue
		}
		if a, ok := held[r.key]; ok {
			if a.ActivityPlanID != nil {
				if h, live := holders[*a.ActivityPlanID]; live && h.ID != planID {
					return Fail(CodeAnnualActivityTaken, http.StatusConflict,
						fmt.Sprintf("assignment %d is already in %q", r.no(), h.PlanTitle),
						Details{Field: r.field("annualActivityId"), Row: r.no(),
							ActivityPlanID: h.ID.String(), ActivityPlanTitle: h.PlanTitle, AssignmentCode: a.ProjectCode})
				}
			}
			r.target = a // adopt
			continue
		}
		restoreKeys = append(restoreKeys, r.key)
	}
	if creating || len(restoreKeys) == 0 {
		return nil
	}

	var gone []models.AuditActivity
	if err := tx.Unscoped().
		Where("activity_plan_id = ? AND annual_activity_id IN ? AND deleted_at IS NOT NULL", planID, restoreKeys).
		Order("deleted_at DESC").Find(&gone).Error; err != nil {
		return err
	}
	latest := map[string]*models.AuditActivity{}
	for i := range gone {
		k := RowKey(strValue(gone[i].AnnualActivityID))
		if latest[k] == nil {
			latest[k] = &gone[i]
		}
	}
	for _, r := range rows {
		if r.target == nil && latest[r.key] != nil {
			r.target, r.restore = latest[r.key], true
		}
	}
	return nil
}

// releaseRow takes a row out of an RKA: a row that existed before the RKA
// (adopted) is detached, a row the RKA created is soft-deleted.
func releaseRow(tx *gorm.DB, a models.AuditActivity, planCreatedAt time.Time) error {
	if a.CreatedAt.Before(planCreatedAt) {
		return tx.Model(&models.AuditActivity{}).Where("id = ?", a.ID).
			Updates(map[string]interface{}{"activity_plan_id": nil, "sort_order": 0}).Error
	}
	return tx.Delete(&models.AuditActivity{}, "id = ?", a.ID).Error
}

func writeHeader(tx *gorm.DB, existing *models.ActivityPlan, planID uuid.UUID, pkat *PKAT,
	activityType, title string, in SaveInput, actor string, now time.Time) error {
	today := now.Format("2006-01-02")
	year := strconv.Itoa(pkat.Plan.Year)
	pkatID := pkat.Plan.ID

	if existing == nil {
		plan := models.ActivityPlan{
			ID:           planID,
			AnnualPlanID: &pkatID,
			ActivityType: activityType,
			FormVersion:  models.ActivityPlanFormRKA,
			PlanTitle:    title,
			PlanYear:     year,
			CreatedBy:    actor,
			CreationDate: today,
			Status:       statusPlanned,
			Attachments:  []models.AnnualAuditAttachment{},
		}
		if in.AttachmentCategory != nil {
			plan.AttachmentCategory = strings.TrimSpace(*in.AttachmentCategory)
		}
		if in.Attachments != nil && len(*in.Attachments) > 0 {
			plan.Attachments = *in.Attachments
			plan.AttachmentUploadedBy = actor
			plan.AttachmentUploadDate = today
		}
		if err := tx.Create(&plan).Error; err != nil {
			return classifyWriteError(err, 0, "")
		}
		return nil
	}

	// Struct update with an explicit column list, so the json serializer of
	// attachments applies and zero values are written.
	upd := *existing
	upd.ActivityType = activityType
	upd.PlanTitle = title
	upd.PlanYear = year
	upd.FormVersion = models.ActivityPlanFormRKA
	upd.UpdatedAt = now
	cols := []string{"activity_type", "plan_title", "plan_year", "form_version", "updated_at"}
	if in.AttachmentCategory != nil {
		upd.AttachmentCategory = strings.TrimSpace(*in.AttachmentCategory)
		cols = append(cols, "attachment_category")
	}
	if in.Attachments != nil {
		upd.Attachments = *in.Attachments
		if upd.Attachments == nil {
			upd.Attachments = []models.AnnualAuditAttachment{}
		}
		cols = append(cols, "attachments")
		if len(upd.Attachments) > len(existing.Attachments) {
			upd.AttachmentUploadedBy = actor
			upd.AttachmentUploadDate = today
			cols = append(cols, "attachment_uploaded_by", "attachment_upload_date")
		}
	}
	if err := tx.Model(&models.ActivityPlan{ID: existing.ID}).Select(cols).Updates(&upd).Error; err != nil {
		return classifyWriteError(err, 0, "")
	}
	return nil
}

// snapshot is the PKAT-owned columns of an RKA row (I4).
func snapshot(row *models.AnnualAuditActivity) map[string]interface{} {
	return map[string]interface{}{
		"title":             row.Name,
		"audit_type":        row.Category,
		"audit_universe_id": nullableUUID(row.AuditUniverseID),
		"target_unit_id":    nullableUUID(row.DepartmentID),
		"team_size":         row.AuditorCount,
		"total_mandays":     row.TotalMandays,
	}
}

// nullableUUID is the parsed id, or nil (SQL NULL) when s is not a UUID.
func nullableUUID(s string) interface{} {
	if u, err := uuid.Parse(strings.TrimSpace(s)); err == nil && u != uuid.Nil {
		return u
	}
	return nil
}

func uuidPtr(s string) *uuid.UUID {
	if u, ok := nullableUUID(s).(uuid.UUID); ok {
		return &u
	}
	return nil
}

func writeRow(tx *gorm.DB, r *rowPlan, planID uuid.UUID, pkat *PKAT) error {
	pkatID := pkat.Plan.ID
	var leader interface{}
	if r.leaderID != nil {
		leader = *r.leaderID
	}
	cols := snapshot(r.pkatRow)
	cols["annual_plan_id"] = pkatID
	cols["activity_plan_id"] = planID
	cols["annual_activity_id"] = r.key
	cols["sort_order"] = r.no()
	cols["project_code"] = r.code
	cols["audit_kind"] = r.kind
	cols["justification"] = Join(r.just)
	cols["audit_purpose"] = Join(r.obj)
	cols["audit_focus"] = Join(r.focus)
	cols["team_leader_id"] = leader
	cols["team_leader_name"] = r.leaderName

	if r.target != nil {
		q := tx.Model(&models.AuditActivity{}).Where("id = ?", r.target.ID)
		if r.restore {
			q = tx.Unscoped().Model(&models.AuditActivity{}).Where("id = ?", r.target.ID)
			cols["deleted_at"] = nil
		}
		if err := q.Updates(cols).Error; err != nil {
			return classifyWriteError(err, r.no(), r.code)
		}
		return nil
	}

	key := r.key
	a := models.AuditActivity{
		ID:               uuid.New(),
		AnnualPlanID:     pkatID,
		ActivityPlanID:   &planID,
		AnnualActivityID: &key,
		SortOrder:        r.no(),
		AuditKind:        r.kind,
		TeamLeaderName:   r.leaderName,
		TeamLeaderID:     r.leaderID,
		TargetUnitID:     uuidPtr(r.pkatRow.DepartmentID),
		ProjectCode:      r.code,
		Title:            r.pkatRow.Name,
		AuditType:        r.pkatRow.Category,
		AuditUniverseID:  uuidPtr(r.pkatRow.AuditUniverseID),
		Justification:    Join(r.just),
		AuditPurpose:     Join(r.obj),
		AuditFocus:       Join(r.focus),
		TeamSize:         r.pkatRow.AuditorCount,
		TotalMandays:     r.pkatRow.TotalMandays,
		Status:           statusPlanned,
	}
	if err := tx.Omit(clause.Associations).Create(&a).Error; err != nil {
		return classifyWriteError(err, r.no(), r.code)
	}
	return nil
}

// classifyWriteError maps a unique violation raised by a concurrent writer
// (the pre-checks above saw no conflict) to the contract's 409 codes.
func classifyWriteError(err error, row int, code string) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "unique") && !strings.Contains(msg, "duplicate key") {
		return err
	}
	switch {
	case strings.Contains(msg, IndexAnnualActivityLive) || strings.Contains(msg, "audit_activities.annual_activity_id"):
		return Fail(CodeAnnualActivityTaken, http.StatusConflict,
			"the PKAT activity was just added to another activity plan", Details{Row: row})
	case strings.Contains(msg, IndexAnnualTypeLive) || strings.Contains(msg, "activity_plans.annual_plan_id"):
		return Fail(CodeActivityPlanDuplicate, http.StatusConflict,
			"an activity plan of this type already exists for this PKAT", Details{Field: "activityType"})
	case strings.Contains(msg, "project_code"):
		return Fail(CodeAssignmentCodeTaken, http.StatusConflict,
			fmt.Sprintf("code %s is already used", code), Details{Row: row, AssignmentCode: code})
	}
	return err
}
