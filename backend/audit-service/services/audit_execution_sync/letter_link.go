package audit_execution_sync

import (
	"context"
	"strings"
	"unicode/utf8"

	"audit-service/models"
	"audit-service/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Letter-driven executions. Every live assignment letter has an audit execution:
// it is created when the letter is created (or by the backfill in
// EnsureExecutionsForAllLetters), follows the letter on update, and is
// soft-deleted with the letter.
//
// The link is the same one Collect uses: an execution belongs to a letter when
// its ref is the letter number or the letter UUID. New executions use the letter
// number (the UUID only when the number is blank).
//
// Field ownership. The progress sync owns progress, status, status_detail and
// the activity link. From the letter, ref, name, category and lead_auditor are
// filled in when the execution is created or when they are blank. On a letter
// update a field follows the letter only while it still holds the value derived
// from the letter before the update; a value a user has edited is kept.

// Column limits of models.AuditExecution.
const (
	maxNameLen     = 255
	maxCategoryLen = 100
	maxLeadLen     = 200
)

// letterFields is what an execution takes from its letter.
type letterFields struct {
	Name        string
	Category    string
	LeadAuditor string
}

// executionRef is the ref a new execution of the letter gets.
func executionRef(l *models.AssignmentLetter) string {
	if n := strings.TrimSpace(l.LetterNumber); n != "" {
		return n
	}
	return l.ID.String()
}

// fieldsFromLetter maps a letter onto execution fields: audit title to name (the
// letter number when the title is blank), category as is, and the team leader to
// lead auditor. The team leader is the letter's leader, else the first member
// whose role is Chairperson (the role the Assignment Letter form gives the team
// leader).
func fieldsFromLetter(l *models.AssignmentLetter) letterFields {
	name := strings.TrimSpace(l.AuditTitle)
	if name == "" {
		name = executionRef(l)
	}
	return letterFields{
		Name:        truncate(name, maxNameLen),
		Category:    truncate(strings.TrimSpace(l.Category), maxCategoryLen),
		LeadAuditor: truncate(teamLeader(l), maxLeadLen),
	}
}

func teamLeader(l *models.AssignmentLetter) string {
	if s := strings.TrimSpace(l.Leader); s != "" {
		return s
	}
	for _, m := range l.MembersList {
		if isLeaderRole(m.Role) {
			if s := strings.TrimSpace(m.Name); s != "" {
				return s
			}
		}
	}
	return ""
}

func isLeaderRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "chairperson", "team leader", "leader", "ketua", "ketua tim":
		return true
	}
	return false
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}

// lockLetter serialises execution creation for one letter until the end of the
// transaction, so two concurrent writers cannot both create its execution.
// Postgres only; sqlite (tests) uses a single connection.
func lockLetter(tx *gorm.DB, letterID uuid.UUID) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "audit_execution:"+letterID.String()).Error
}

// EnsureExecutionForLetter makes sure the live letter has a live execution and
// reports whether one was created. prev is the letter as stored before an
// update, or nil on create and backfill. It does not derive progress; call
// SyncExecutionsForLetter (or ReconcileLetter) for that.
//
// Existing executions are matched by the letter number or UUID, plus prev's
// number when the number changed (the execution is moved to the new number, so
// a renumbered letter keeps one execution). Several matching executions (rows
// created by hand) are all kept; none is added.
func EnsureExecutionForLetter(ctx context.Context, db *gorm.DB, letter *models.AssignmentLetter, prev *models.AssignmentLetter) (bool, error) {
	if letter == nil || letter.ID == uuid.Nil {
		return false, nil
	}
	created := false
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockLetter(tx, letter.ID); err != nil {
			return err
		}

		keys := letterKeys(executionRef(letter), []models.AssignmentLetter{*letter})
		renamedFrom := ""
		if prev != nil {
			old := strings.TrimSpace(prev.LetterNumber)
			if old != "" && old != strings.TrimSpace(letter.LetterNumber) {
				// Only claim the old number if no other live letter holds it.
				others, err := findLetters(tx, old)
				if err != nil {
					return err
				}
				if len(others) == 0 {
					renamedFrom = old
					keys = append(keys, old)
				}
			}
		}

		var execs []models.AuditExecution
		if err := tx.Where("ref IN ?", keys).Order("created_at ASC").Find(&execs).Error; err != nil {
			return err
		}

		want := fieldsFromLetter(letter)
		if len(execs) == 0 {
			status, detail := models.MapProgressToStatus(0)
			exec := models.AuditExecution{
				ID:           uuid.New(),
				Ref:          executionRef(letter),
				Name:         want.Name,
				Category:     want.Category,
				LeadAuditor:  want.LeadAuditor,
				Status:       status,
				StatusDetail: detail,
			}
			if err := tx.Create(&exec).Error; err != nil {
				return err
			}
			created = true
			return nil
		}

		var was *letterFields
		if prev != nil {
			f := fieldsFromLetter(prev)
			was = &f
		}
		for i := range execs {
			updates := letterUpdates(&execs[i], want, was)
			if renamedFrom != "" && execs[i].Ref == renamedFrom {
				updates["ref"] = executionRef(letter)
			}
			if len(updates) == 0 {
				continue
			}
			if err := tx.Model(&execs[i]).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

// letterUpdates returns the letter-owned columns to change on an existing
// execution. A blank field is filled. With was (the fields derived from the
// letter before an update), a field still equal to its old derived value
// follows the letter; anything else was edited by a user and is kept.
func letterUpdates(exec *models.AuditExecution, want letterFields, was *letterFields) map[string]interface{} {
	updates := map[string]interface{}{}
	apply := func(col, current, next, old string, oldOK bool) {
		if next == "" || current == next {
			return
		}
		if strings.TrimSpace(current) == "" || (oldOK && current == old) {
			updates[col] = next
		}
	}
	var old letterFields
	if was != nil {
		old = *was
	}
	apply("name", exec.Name, want.Name, old.Name, was != nil)
	apply("category", exec.Category, want.Category, old.Category, was != nil)
	apply("lead_auditor", exec.LeadAuditor, want.LeadAuditor, old.LeadAuditor, was != nil)
	return updates
}

// RetireExecutionsForLetter soft-deletes the live executions of a letter that is
// no longer live. Nothing happens while a live letter still has the number or
// id, so it is safe to call twice or for a letter that was not deleted. It
// returns the number of executions deleted.
func RetireExecutionsForLetter(ctx context.Context, db *gorm.DB, letter *models.AssignmentLetter) (int64, error) {
	if letter == nil || letter.ID == uuid.Nil {
		return 0, nil
	}
	db = db.WithContext(ctx)
	var live int64
	q := db.Model(&models.AssignmentLetter{})
	if n := strings.TrimSpace(letter.LetterNumber); n != "" {
		q = q.Where("id = ? OR letter_number = ?", letter.ID, n)
	} else {
		q = q.Where("id = ?", letter.ID)
	}
	if err := q.Count(&live).Error; err != nil {
		return 0, err
	}
	if live > 0 {
		return 0, nil
	}
	res := db.Where("ref IN ?", letterKeys(executionRef(letter), []models.AssignmentLetter{*letter})).
		Delete(&models.AuditExecution{})
	return res.RowsAffected, res.Error
}

// ReconcileLetter brings the executions of one letter in line after a write.
// letterID is the written letter; prev is the letter as stored before the write
// (nil on create). A live letter gets its execution ensured and synced; a letter
// that is gone (deleted) has its executions soft-deleted.
func ReconcileLetter(ctx context.Context, db *gorm.DB, letterID uuid.UUID, prev *models.AssignmentLetter) error {
	var letter models.AssignmentLetter
	err := db.WithContext(ctx).Where("id = ?", letterID).First(&letter).Error
	if err == gorm.ErrRecordNotFound {
		if prev == nil {
			return nil
		}
		_, err := RetireExecutionsForLetter(ctx, db, prev)
		return err
	}
	if err != nil {
		return err
	}
	if _, err := EnsureExecutionForLetter(ctx, db, &letter, prev); err != nil {
		return err
	}
	_, err = SyncExecutionsForLetter(ctx, db, letter.ID.String())
	return err
}

// EnsureExecutionsForAllLetters is the backfill: it creates the missing
// execution of every live letter (Draft included) and returns how many it
// created. A failing letter is logged and skipped. Progress is not derived here;
// SyncAll does that.
func EnsureExecutionsForAllLetters(ctx context.Context, db *gorm.DB) (int, error) {
	var letters []models.AssignmentLetter
	if err := db.WithContext(ctx).Order("created_at ASC").Order("id ASC").Find(&letters).Error; err != nil {
		return 0, err
	}
	created := 0
	for i := range letters {
		ok, err := EnsureExecutionForLetter(ctx, db, &letters[i], nil)
		if err != nil {
			logger.Warn("[audit_execution_sync] ensure execution for letter failed",
				logger.LogField("letter", letters[i].LetterNumber),
				logger.LogField("error", err),
			)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}
