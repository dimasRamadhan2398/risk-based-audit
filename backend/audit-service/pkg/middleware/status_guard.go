package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HasAnyRole reports whether the authenticated caller has one of the given
// roles (case-insensitive). Requires Authenticate to have run first.
func HasAnyRole(c *gin.Context, roles ...string) bool {
	userRoles, _ := c.Get("roles")
	roleList, _ := userRoles.([]string)
	for _, userRole := range roleList {
		for _, role := range roles {
			if strings.EqualFold(userRole, role) {
				return true
			}
		}
	}
	return false
}

// RequireRolesForStatusChange guards the "status" field of a record on the
// generic CRUD routes: other fields stay editable by anyone who can reach the
// route, but only the given roles may create a record with a status other
// than initialStatus, or change the status of an existing record.
//
// table is the database table used to look up the current status on PUT.
func RequireRolesForStatusChange(db *gorm.DB, table, initialStatus string, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasAnyRole(c, roles...) {
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
		requested := *payload.Status

		current := initialStatus
		if c.Request.Method != http.MethodPost {
			var existing struct{ Status string }
			if err := db.Table(table).Select("status").Where("id = ? AND deleted_at IS NULL", c.Param("id")).Take(&existing).Error; err != nil {
				// Let the handler report not-found / invalid ID as usual
				c.Next()
				return
			}
			current = existing.Status
		}

		if requested != current {
			response.Forbidden(c, "You are not allowed to change the status to "+requested)
			c.Abort()
			return
		}

		c.Next()
	}
}
