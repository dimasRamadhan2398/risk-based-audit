package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"audit-service/models"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// derivedFields are read-only response fields that clients must not set
var derivedFields = []string{"isOverdue", "is_overdue", "daysOverdue", "days_overdue"}

// NormalizeRequest runs before the generic crud Create/Update handlers for
// action taken reports. It rewrites "status" to its stored upper snake form
// ("In Progress" -> "IN_PROGRESS"), rejects unknown statuses with 400, drops an
// empty/null status (create falls back to the PLANNED default, update leaves it
// unchanged) and strips the derived isOverdue/daysOverdue fields.
//
// Bodies that are not a JSON object are passed through untouched so the crud
// handler reports the binding error as before.
func NormalizeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}
		raw, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		if err != nil {
			response.BadRequest(c, "Failed to read request body")
			c.Abort()
			return
		}

		var body map[string]json.RawMessage
		if err := json.Unmarshal(raw, &body); err != nil || body == nil {
			setBody(c, raw)
			c.Next()
			return
		}

		for key, value := range body {
			// encoding/json binds keys case-insensitively and crud.Update snake-cases
			// them, so "Status" reaches the same column as "status"
			if isDerivedField(key) {
				delete(body, key)
				continue
			}
			if !strings.EqualFold(key, "status") {
				continue
			}
			if string(bytes.TrimSpace(value)) == "null" {
				delete(body, key)
				continue
			}
			var s string
			if err := json.Unmarshal(value, &s); err != nil {
				response.BadRequest(c, "Invalid status: must be a string, one of "+strings.Join(models.ATRStatuses, ", "))
				c.Abort()
				return
			}
			if strings.TrimSpace(s) == "" {
				delete(body, key)
				continue
			}
			status, ok := models.NormalizeATRStatus(s)
			if !ok {
				response.BadRequest(c, "Invalid status \""+s+"\": must be one of "+strings.Join(models.ATRStatuses, ", "))
				c.Abort()
				return
			}
			encoded, _ := json.Marshal(status)
			body[key] = encoded
		}

		rewritten, err := json.Marshal(body)
		if err != nil {
			response.InternalServerError(c, "Failed to process request body")
			c.Abort()
			return
		}
		setBody(c, rewritten)
		c.Next()
	}
}

func isDerivedField(key string) bool {
	for _, f := range derivedFields {
		if strings.EqualFold(key, f) {
			return true
		}
	}
	return false
}

func setBody(c *gin.Context, b []byte) {
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	c.Request.ContentLength = int64(len(b))
}
