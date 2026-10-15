package audit_execution

import (
	"fmt"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/services/audit_execution_sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListItem is an audit execution in the list response, with the source signals
// it was derived from. The embedded fields keep the execution's own JSON shape.
type ListItem struct {
	models.AuditExecution
	Sources audit_execution_sync.Signals `json:"sources"`
}

// ListWithSources is GET /audit-executions. It is the generic list (same filters,
// order, search and pagination) with each row's progress and status derived from
// its sources, so the list agrees with the sources it returns. It is read-only:
// the stored row is not changed here.
func ListWithSources(db *gorm.DB) gin.HandlerFunc {
	return crud.ListWithItems(db, "AuditExecution", func() interface{} { return &[]models.AuditExecution{} },
		func(c *gin.Context, items interface{}) (interface{}, error) {
			execs, ok := items.(*[]models.AuditExecution)
			if !ok {
				return nil, fmt.Errorf("unexpected list type %T", items)
			}
			out := make([]ListItem, 0, len(*execs))
			for i := range *execs {
				exec := (*execs)[i]
				sources, _, err := audit_execution_sync.Collect(c.Request.Context(), db, &exec)
				if err != nil {
					return nil, err
				}
				progress := audit_execution_sync.DeriveProgress(sources)
				exec.Progress = progress
				exec.Status, exec.StatusDetail = models.MapProgressToStatus(progress)
				out = append(out, ListItem{AuditExecution: exec, Sources: sources})
			}
			return out, nil
		})
}
