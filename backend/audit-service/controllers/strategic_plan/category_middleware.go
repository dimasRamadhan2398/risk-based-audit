// Package strategicplan holds request handling specific to /strategic-plans that
// the generic crud handlers do not cover.
package strategicplan

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"audit-service/models"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// ValidateCategory runs before the generic crud Create/Update handlers for
// strategic plans. It rejects a "category" outside models.StrategicPlanCategories
// with 400 and rewrites a known one to its stored form ("financial" -> "Financial").
//
//   - absent: create stores "", update leaves the category unchanged
//   - "" (or blank): stored as "", so an update with "" clears the category
//   - null: dropped, i.e. treated as absent (create "", update unchanged)
//
// Bodies that are not a JSON object are passed through untouched so the crud
// handler reports the binding error as before.
func ValidateCategory() gin.HandlerFunc {
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

		changed := false
		for key, value := range body {
			// encoding/json binds keys case-insensitively and crud.Update snake-cases
			// them, so "Category" reaches the same column as "category"
			if !strings.EqualFold(key, "category") {
				continue
			}
			changed = true
			if string(bytes.TrimSpace(value)) == "null" {
				delete(body, key)
				continue
			}
			var s string
			if err := json.Unmarshal(value, &s); err != nil {
				response.BadRequest(c, "Invalid category: must be a string, one of "+allowedList())
				c.Abort()
				return
			}
			category, ok := models.NormalizeStrategicPlanCategory(s)
			if !ok {
				response.BadRequest(c, "Invalid category \""+s+"\": must be one of "+allowedList())
				c.Abort()
				return
			}
			encoded, _ := json.Marshal(category)
			body[key] = encoded
		}
		if !changed {
			setBody(c, raw)
			c.Next()
			return
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

func allowedList() string {
	return strings.Join(models.StrategicPlanCategories, ", ") + " or empty"
}

func setBody(c *gin.Context, b []byte) {
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	c.Request.ContentLength = int64(len(b))
}
