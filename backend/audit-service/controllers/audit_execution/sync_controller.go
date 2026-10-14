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

// SyncFromSources recomputes progress and status for every audit execution
// from the annual plan, assignment letter, fieldwork, working paper and LHA data.
// POST /audit-executions/sync
func (c *SyncController) SyncFromSources(ctx *gin.Context) {
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
		"items":                 results,
		"unmatched_source_rows": unmatched,
	})
}
