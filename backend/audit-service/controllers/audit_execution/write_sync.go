package audit_execution

import (
	"context"
	"net/http"
	"strings"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/services/audit_execution_sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Save-triggered sync. Every write to a letter-linked row re-derives the audit
// executions of that letter, so the Audit Execution Status follows the data
// without a manual POST /audit-executions/sync. A failed sync is logged and never
// fails the save.

// LetterKeysOf returns the assignment letter keys a row belongs to: the letter
// number or the letter UUID, whichever the row stores. Rows of the working paper
// tables store the key in WorkingPaperID, and letters carry both forms.
// Unknown types return nil.
func LetterKeysOf(entity interface{}) []string {
	switch e := entity.(type) {
	case *models.FieldworkInterview:
		return []string{e.AssignmentLetterID}
	case *models.FieldworkObservation:
		return []string{e.AssignmentLetterID}
	case *models.FieldworkDocument:
		return []string{e.AssignmentLetterID}
	case *models.FieldworkSample:
		return []string{e.AssignmentLetterID}
	case *models.FieldworkTestControl:
		return []string{e.AssignmentLetterID}
	case *models.WorkingPaperHeader:
		return []string{e.AssignmentLetterID}
	case *models.WorkingPaperRisk:
		return []string{e.WorkingPaperID}
	case *models.WorkingPaperSample:
		return []string{e.WorkingPaperID}
	case *models.WorkingPaperCause:
		return []string{e.WorkingPaperID}
	case *models.WorkingPaperPlan:
		return []string{e.WorkingPaperID}
	case *models.AuditResultReport:
		return []string{e.AssignmentLetterID}
	case *models.AssignmentLetter:
		keys := []string{e.LetterNumber}
		if e.ID != uuid.Nil {
			keys = append(keys, e.ID.String())
		}
		return keys
	}
	return nil
}

// SyncAfterCreate is the after-write hook of a create route. It runs inside the
// INSERT transaction, so the sync sees the new row. Each letter sync runs in a
// savepoint: a failure rolls back only the sync, is logged, and the save commits.
func SyncAfterCreate(keysOf func(interface{}) []string) crud.AfterHook {
	return func(tx *gorm.DB, entity interface{}) error {
		for _, key := range keysOf(entity) {
			if strings.TrimSpace(key) == "" {
				continue
			}
			err := tx.Transaction(func(sp *gorm.DB) error {
				_, err := audit_execution_sync.SyncExecutionsForLetter(context.Background(), sp, key)
				return err
			})
			if err != nil {
				logSyncFailure(key, err)
			}
		}
		return nil
	}
}

// SyncOnWrite wraps a PUT or DELETE handler. Before the write it reads the stored
// row (by the :id parameter) and its letter keys. After a 2xx response it reads
// the row again, so a letter change is covered for both letters, and re-derives
// the executions of every key. The sync runs after the write has committed; a
// failure is logged and the response is left as the handler wrote it.
//
// newEntity returns an empty model of the route's type; keysOf is LetterKeysOf
// or a variant for the model.
func SyncOnWrite(db *gorm.DB, newEntity func() interface{}, keysOf func(interface{}) []string, next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := storedLetterKeys(db, c.Param("id"), newEntity, keysOf)
		next(c)
		if status := c.Writer.Status(); status < http.StatusOK || status >= http.StatusMultipleChoices {
			return
		}
		keys = append(keys, storedLetterKeys(db, c.Param("id"), newEntity, keysOf)...)
		syncLetterKeys(c.Request.Context(), db, keys)
	}
}

// storedLetterKeys returns the letter keys of the stored row with the given id,
// or nil when the id is not a UUID or no live row matches.
func storedLetterKeys(db *gorm.DB, id string, newEntity func() interface{}, keysOf func(interface{}) []string) []string {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	entity := newEntity()
	if err := db.Where("id = ?", uid).First(entity).Error; err != nil {
		return nil
	}
	return keysOf(entity)
}

// syncLetterKeys re-derives the executions of each distinct, non-blank key.
func syncLetterKeys(ctx context.Context, db *gorm.DB, keys []string) {
	seen := map[string]bool{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		if _, err := audit_execution_sync.SyncExecutionsForLetter(ctx, db, key); err != nil {
			logSyncFailure(key, err)
		}
	}
}

func logSyncFailure(key string, err error) {
	logger.Warn("[audit_execution_sync] save-triggered sync failed",
		logger.LogField("letter", key),
		logger.LogField("error", err),
	)
}
