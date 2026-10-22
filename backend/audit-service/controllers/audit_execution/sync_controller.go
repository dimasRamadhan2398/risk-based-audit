package audit_execution

import (
	"net/http"

	"audit-service/pkg/response"
	"audit-service/services/audit_execution_sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SyncController struct {
	db *gorm.DB
}

func NewSyncController(db *gorm.DB) *SyncController {
	return &SyncController{db: db}
}

// SyncFromSources first creates the missing audit execution of every live
// assignment letter (backfill, idempotent), then recomputes progress and status
// for every audit execution from the annual plan, assignment letter, fieldwork,
// working paper and LHA data. created is the number of executions the backfill
// added.
// POST /audit-executions/sync
func (c *SyncController) SyncFromSources(ctx *gin.Context) {
	created, err := audit_execution_sync.EnsureExecutionsForAllLetters(ctx.Request.Context(), c.db)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "SYNC_FAILED", "Failed to create audit executions for assignment letters", err.Error())
		return
	}
	results, err := audit_execution_sync.SyncAll(ctx.Request.Context(), c.db)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "SYNC_FAILED", "Failed to sync audit executions", err.Error())
		return
	}
	// Source rows whose letter reference matches no assignment letter. They are
	// not counted in any execution; a non-zero count means a reference is broken.
	unmatched, err := audit_execution_sync.UnmatchedSourceRows(ctx.Request.Context(), c.db)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "SYNC_FAILED", "Failed to check unmatched source rows", err.Error())
		return
	}
	response.Success(ctx, http.StatusOK, "Audit executions synced from source modules", gin.H{
		"synced":                len(results),
		"created":               created,
		"items":                 results,
		"unmatched_source_rows": unmatched,
	})
}
