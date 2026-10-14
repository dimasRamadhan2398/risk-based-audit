package audit_execution_sync

import (
	"context"
	"strings"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/services/audit_completion"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Result is what one sync run found and wrote for an execution.
type Result struct {
	ExecutionID uuid.UUID `json:"id"`
	Ref         string    `json:"ref"`
	Progress    int       `json:"progress"`
	Status      string    `json:"status"`
	Sources     Signals   `json:"sources"`
}

// Collect reads the source modules for one execution.
//
// The key is the assignment letter. Fieldwork, working paper and LHA rows store
// the letter in one of two forms: the letter number (ST-001/SKAI/2026) or the
// letter's UUID (see models.FindAssignmentLetter). Collect matches both forms, so
// rows written by either path count. It also returns the annual plan activity the
// execution belongs to, or nil.
func Collect(ctx context.Context, db *gorm.DB, exec *models.AuditExecution) (Signals, *uuid.UUID, error) {
	var s Signals
	db = db.WithContext(ctx)
	ref := strings.TrimSpace(exec.Ref)

	// 1. Annual Audit Plan: explicit link first, then project code or title
	activityID, err := findActivityID(db, exec)
	if err != nil {
		return s, nil, err
	}
	s.AnnualPlanLinked = activityID != nil

	if ref == "" {
		return s, activityID, nil
	}

	// 2. Assignment Letter: issued once it is no longer a draft. Its entry items
	// (scope, purpose, team) place the execution inside the 1..25 entry band.
	letters, err := findLetters(db, ref)
	if err != nil {
		return s, nil, err
	}
	keys := letterKeys(ref, letters)
	for _, l := range letters {
		if !letterIssued(l.Status) {
			continue
		}
		s.AssignmentLetterIssued = true
		if n := entryAlignment(l); n > s.EntryAlignment {
			s.EntryAlignment = n
		}
	}

	// 3. Audit Fieldwork: evidence records plus the test controls being tested
	for _, m := range []any{
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		&models.FieldworkSample{},
	} {
		var n int64
		if err := db.Model(m).Where(colAssignmentLetterID+" IN ?", keys).Count(&n).Error; err != nil {
			return s, nil, err
		}
		s.FieldworkRecords += int(n)
	}

	var controls []models.FieldworkTestControl
	if err := db.Where(colAssignmentLetterID+" IN ?", keys).Find(&controls).Error; err != nil {
		return s, nil, err
	}
	s.TestControls = len(controls)
	for _, c := range controls {
		if strings.TrimSpace(c.TestResult) != "" {
			s.TestControlsTested++
		}
		if strings.TrimSpace(c.Finding) != "" {
			s.TestControlsWithFinding++
		}
	}

	// Mitigation ("mitigation" in the product spec) is read by column name, not by
	// struct field, so the derivation depends on the column the schema creates.
	var withMitigation int64
	if err := db.Model(&models.FieldworkTestControl{}).
		Where(colAssignmentLetterID+" IN ?", keys).
		Where("TRIM(COALESCE(" + colMitigationPlan + ", '')) <> ''").
		Count(&withMitigation).Error; err != nil {
		return s, nil, err
	}
	s.TestControlsWithMitigation = int(withMitigation)

	// 4. Working Paper: causes and complete plans drive the draft findings band.
	// Header, risk and sample rows are counted for information only.
	var wpHeaders int64
	if err := db.Model(&models.WorkingPaperHeader{}).Where(colAssignmentLetterID+" IN ?", keys).Count(&wpHeaders).Error; err != nil {
		return s, nil, err
	}
	s.WorkingPaperRecords += int(wpHeaders)
	for _, m := range []any{&models.WorkingPaperRisk{}, &models.WorkingPaperSample{}} {
		var n int64
		if err := db.Model(m).Where(colWorkingPaperID+" IN ?", keys).Count(&n).Error; err != nil {
			return s, nil, err
		}
		s.WorkingPaperRecords += int(n)
	}
	var causes int64
	if err := db.Model(&models.WorkingPaperCause{}).Where(colWorkingPaperID+" IN ?", keys).Count(&causes).Error; err != nil {
		return s, nil, err
	}
	s.WorkingPaperCauses = int(causes)

	var plans []models.WorkingPaperPlan
	if err := db.Where(colWorkingPaperID+" IN ?", keys).Find(&plans).Error; err != nil {
		return s, nil, err
	}
	for _, p := range plans {
		if planComplete(p) {
			s.WorkingPaperPlansComplete++
		}
	}

	// 5. Result Report (LHA): rejected and cancelled reports do not count. Any
	// approved report completes the audit; an in-review one is at 88, or 90 when
	// it waits on CAE. Only approved reports are final.
	var reports []models.AuditResultReport
	if err := db.Where(colAssignmentLetterID+" IN ?", keys).Find(&reports).Error; err != nil {
		return s, nil, err
	}
	for _, r := range reports {
		if models.IsLHARejectedOrCancelled(r.Status) {
			continue
		}
		s.LHAExists = true
		switch {
		case models.IsLHAApproved(r.Status):
			s.LHAApproved = true
		case models.IsLHAInReview(r.Status):
			s.LHAInReview = true
			if models.IsLHAAwaitingCAE(r.Status) {
				s.LHAAwaitingCAE = true
			}
		}
	}

	return s, activityID, nil
}

// findLetters returns the assignment letters the execution ref points at: the
// letter number, or the letter id when ref is a UUID.
func findLetters(db *gorm.DB, ref string) ([]models.AssignmentLetter, error) {
	q := db.Where("letter_number = ?", ref)
	if _, err := uuid.Parse(ref); err == nil {
		q = db.Where("letter_number = ? OR id = ?", ref, ref)
	}
	var letters []models.AssignmentLetter
	if err := q.Find(&letters).Error; err != nil {
		return nil, err
	}
	return letters, nil
}

// letterKeys is every value a child row may hold for this execution's letter:
// the ref itself, and the letter number and UUID of each matched letter.
func letterKeys(ref string, letters []models.AssignmentLetter) []string {
	seen := map[string]bool{}
	var keys []string
	add := func(k string) {
		if k != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	add(ref)
	for _, l := range letters {
		add(l.LetterNumber)
		add(l.ID.String())
	}
	return keys
}

// letterIssued reports whether a letter status counts as issued: any status
// other than draft. A blank status is not treated as issued.
func letterIssued(status string) bool {
	s := strings.TrimSpace(status)
	return s != "" && !strings.EqualFold(s, "draft")
}

// entryAlignment counts the entry items recorded on a letter: scope, audit
// purpose and team. Each one is 0 or 1; the total is 0..entryItems.
func entryAlignment(l models.AssignmentLetter) int {
	n := 0
	for _, scope := range l.ScopeList {
		if strings.TrimSpace(scope) != "" {
			n++
			break
		}
	}
	if strings.TrimSpace(l.EffectiveAuditPurpose()) != "" {
		n++
	}
	for _, m := range l.MembersList {
		if strings.TrimSpace(m.Name) != "" {
			n++
			break
		}
	}
	return n
}

// planComplete reports whether a working paper plan has both the action
// description and the management recommendation filled in.
func planComplete(p models.WorkingPaperPlan) bool {
	return strings.TrimSpace(p.ActionDescription) != "" && strings.TrimSpace(p.Recommendation) != ""
}

// findActivityID returns the annual plan activity linked to the execution, or nil.
// A blank name or ref never matches: an empty title or project code in the plan
// must not link to an execution with no name or ref.
func findActivityID(db *gorm.DB, exec *models.AuditExecution) (*uuid.UUID, error) {
	if exec.ActivityID != nil {
		var act models.AuditActivity
		err := db.Select("id").First(&act, "id = ?", *exec.ActivityID).Error
		if err == nil {
			return &act.ID, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	name := strings.TrimSpace(exec.Name)
	ref := strings.TrimSpace(exec.Ref)
	var conds []string
	var args []interface{}
	if name != "" {
		conds = append(conds, "LOWER(title) = LOWER(?)")
		args = append(args, name)
	}
	if ref != "" {
		conds = append(conds, "LOWER(project_code) = LOWER(?)")
		args = append(args, ref)
	}
	if len(conds) == 0 {
		return nil, nil
	}

	var act models.AuditActivity
	err := db.Select("id").Where(strings.Join(conds, " OR "), args...).First(&act).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &act.ID, nil
}

// UnmatchedSourceRows counts rows in the source tables whose assignment letter
// reference matches no live assignment letter (by letter number or id). Such rows
// never reach an execution, so the count shows the gap. Only tables with unmatched
// rows appear in the map.
func UnmatchedSourceRows(ctx context.Context, db *gorm.DB) (map[string]int64, error) {
	db = db.WithContext(ctx)

	var letters []models.AssignmentLetter
	if err := db.Select("id", "letter_number").Find(&letters).Error; err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, l := range letters {
		known[l.LetterNumber] = true
		known[l.ID.String()] = true
	}

	type source struct {
		table string
		model any
		col   string
	}
	sources := []source{
		{"fieldwork_interviews", &models.FieldworkInterview{}, colAssignmentLetterID},
		{"fieldwork_observations", &models.FieldworkObservation{}, colAssignmentLetterID},
		{"fieldwork_documents", &models.FieldworkDocument{}, colAssignmentLetterID},
		{"fieldwork_samples", &models.FieldworkSample{}, colAssignmentLetterID},
		{"fieldwork_test_controls", &models.FieldworkTestControl{}, colAssignmentLetterID},
		{"working_paper_headers", &models.WorkingPaperHeader{}, colAssignmentLetterID},
		{"working_paper_risks", &models.WorkingPaperRisk{}, colWorkingPaperID},
		{"working_paper_samples", &models.WorkingPaperSample{}, colWorkingPaperID},
		{"working_paper_causes", &models.WorkingPaperCause{}, colWorkingPaperID},
		{"working_paper_plans", &models.WorkingPaperPlan{}, colWorkingPaperID},
		{"audit_result_reports", &models.AuditResultReport{}, colAssignmentLetterID},
	}

	type row struct {
		RowKey   string
		RowCount int64
	}
	out := map[string]int64{}
	for _, src := range sources {
		var rows []row
		err := db.Model(src.model).
			Select("COALESCE(" + src.col + ", '') AS row_key, COUNT(*) AS row_count").
			Group(src.col).
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if !known[r.RowKey] {
				out[src.table] += r.RowCount
			}
		}
	}
	for table, n := range out {
		if n == 0 {
			delete(out, table)
		}
	}
	return out, nil
}

// SyncExecution derives progress from the source modules, saves it on the
// execution, and pushes the status on to its annual plan activity.
// exec is updated in place. The execution row is written only when progress,
// status, status detail or the activity link changed.
func SyncExecution(ctx context.Context, db *gorm.DB, exec *models.AuditExecution) (*Result, error) {
	sources, activityID, err := Collect(ctx, db, exec)
	if err != nil {
		return nil, err
	}

	progress := DeriveProgress(sources)
	status, detail := models.MapProgressToStatus(progress)
	linkActivity := exec.ActivityID == nil && activityID != nil

	if exec.Progress != progress || exec.Status != status || exec.StatusDetail != detail || linkActivity {
		updates := map[string]interface{}{
			"progress":      progress,
			"status":        status,
			"status_detail": detail,
		}
		if linkActivity {
			updates["activity_id"] = *activityID
		}
		if err := db.WithContext(ctx).Model(exec).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	exec.Progress = progress
	exec.Status = status
	exec.StatusDetail = detail
	if linkActivity {
		exec.ActivityID = activityID
	}

	if err := audit_completion.SyncExecutionStatusToActivity(ctx, db, exec); err != nil {
		logger.Warn("[audit_execution_sync] activity status sync warning",
			logger.LogField("execution_id", exec.ID),
			logger.LogField("error", err),
		)
	}

	return &Result{
		ExecutionID: exec.ID,
		Ref:         exec.Ref,
		Progress:    progress,
		Status:      status,
		Sources:     sources,
	}, nil
}

// SyncExecutionsForLetter re-derives every live execution that belongs to one
// assignment letter. letterKey is the letter number or the letter UUID. An
// execution belongs to the letter when its ref equals the number or the UUID of
// a letter with that number (the same two forms Collect matches).
//
// It is the save-time sync: the write hooks of the letter-linked routes call it
// after a successful write. The first failure is returned, so a caller running
// inside a transaction can roll its savepoint back.
func SyncExecutionsForLetter(ctx context.Context, db *gorm.DB, letterKey string) ([]Result, error) {
	key := strings.TrimSpace(letterKey)
	if key == "" {
		return nil, nil
	}
	db = db.WithContext(ctx)

	letters, err := findLetters(db, key)
	if err != nil {
		return nil, err
	}
	var executions []models.AuditExecution
	if err := db.Where("ref IN ?", letterKeys(key, letters)).Find(&executions).Error; err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(executions))
	for i := range executions {
		res, err := SyncExecution(ctx, db, &executions[i])
		if err != nil {
			return results, err
		}
		results = append(results, *res)
	}
	return results, nil
}

// SyncAll syncs every live execution. A failing execution is logged and
// skipped so one bad row does not block the rest.
func SyncAll(ctx context.Context, db *gorm.DB) ([]Result, error) {
	var executions []models.AuditExecution
	if err := db.WithContext(ctx).Find(&executions).Error; err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(executions))
	for i := range executions {
		res, err := SyncExecution(ctx, db, &executions[i])
		if err != nil {
			logger.Warn("[audit_execution_sync] execution sync failed",
				logger.LogField("execution_id", executions[i].ID),
				logger.LogField("error", err),
			)
			continue
		}
		results = append(results, *res)
	}
	return results, nil
}
