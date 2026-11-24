package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"audit-service/models"
	"audit-service/pkg/middleware"
	"audit-service/pkg/response"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LHAApproverRoles are the canonical roles (services/action_taken_report RBAC)
// that may approve an audit result report (LHA), i.e. set it to "Final" (or
// "Approved"/"Published"), or take it back out of that state: Admin, Audit
// Manager and the Chief Audit Executive. The CAE is the seeded EXECUTIVE role.
// Auditors and team leaders may draft and edit a report but not approve it.
//
// The frontend mirrors this set in useRbac canPublishLha.
var LHAApproverRoles = []string{atrsvc.RoleAdmin, atrsvc.RoleAuditManager, atrsvc.RoleCAE}

// CanApproveLHA reports whether a caller with the given JWT roles may approve
// an audit result report
func CanApproveLHA(roles []string) bool {
	return atrsvc.HasCanonicalRole(roles, LHAApproverRoles...)
}

// lhaApprovalGuard keeps approval of an audit result report with LHAApproverRoles.
// Everyone else may still create and edit reports, but may not
//   - create a report that is already approved,
//   - move a report into an approved status, or
//   - change the status of a report that is already approved (re-opening it
//     would cancel its action taken reports).
//
// Statuses are compared case-insensitively because the column defaults to
// "DRAFT" while the form sends "Draft". A request without a "status" field
// does not change the status and passes.
func lhaApprovalGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if CanApproveLHA(middleware.ContextRoles(c)) {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.BadRequest(c, "Failed to read request body")
			c.Abort()
			return
		}
		// Put the body back for the handler that binds it
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		var payload struct {
			Status *string `json:"status"`
		}
		if err := json.Unmarshal(body, &payload); err != nil || payload.Status == nil {
			// Malformed JSON is rejected by the handler itself; no status means no change
			c.Next()
			return
		}
		requested := strings.TrimSpace(*payload.Status)

		current := ""
		if c.Request.Method != http.MethodPost {
			var existing struct{ Status string }
			if err := db.Table("audit_result_reports").Select("status").Where("id = ? AND deleted_at IS NULL", c.Param("id")).Take(&existing).Error; err != nil {
				// Let the handler report not-found / invalid ID as usual
				c.Next()
				return
			}
			current = existing.Status
		}

		switch {
		case models.IsLHAApproved(current) && !strings.EqualFold(requested, strings.TrimSpace(current)):
			response.Forbidden(c, "Only an Admin, Audit Manager or Chief Audit Executive can change the status of an approved audit result report")
			c.Abort()
		case models.IsLHAApproved(requested) && !models.IsLHAApproved(current):
			response.Forbidden(c, "Only an Admin, Audit Manager or Chief Audit Executive can approve an audit result report")
			c.Abort()
		default:
			c.Next()
		}
	}
}
