package assignmentletter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"audit-service/models"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// workingUnitIDKey is the request key of the letter's master department
const workingUnitIDKey = "workingUnitDepartmentId"

// DepartmentGetter reads one master-service department with the caller's
// Authorization header (*masterclient.MasterClient)
type DepartmentGetter interface {
	GetDepartment(ctx context.Context, authorization, id string) (*masterclient.DepartmentDetail, error)
}

// WorkingUnitLabel is the display text of a department in Unit Kerja,
// the same "Name (CODE)" form the letter form shows
func WorkingUnitLabel(name, code string) string {
	name, code = strings.TrimSpace(name), strings.TrimSpace(code)
	if code == "" {
		return name
	}
	return name + " (" + code + ")"
}

// ResolveWorkingUnit validates workingUnitDepartmentId of a letter
// create/update against the master and rewrites workingUnit from the master
// department, so the letter stores the id with a matching display text.
//
//   - key omitted: the request is left alone (legacy clients);
//   - null or "": the link is cleared;
//   - a UUID: it must be an active master department; a letter may keep a
//     department it already has even if it was since deactivated or the
//     master cannot be read.
func ResolveWorkingUnit(db *gorm.DB, depts DepartmentGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.BadRequest(c, "Invalid request body")
			c.Abort()
			return
		}
		restore := func(b []byte) {
			c.Request.Body = io.NopCloser(bytes.NewReader(b))
			c.Request.ContentLength = int64(len(b))
		}
		restore(raw)

		var body map[string]interface{}
		if len(bytes.TrimSpace(raw)) == 0 || json.Unmarshal(raw, &body) != nil {
			c.Next() // the handler reports the bad body
			return
		}
		v, present := body[workingUnitIDKey]
		if !present {
			c.Next()
			return
		}
		s, _ := v.(string)
		s = strings.TrimSpace(s)
		if v != nil {
			if _, isString := v.(string); !isString {
				validationError(c, "must be the UUID of a master department")
				return
			}
		}

		if s == "" {
			body[workingUnitIDKey] = nil
		} else {
			id, perr := uuid.Parse(s)
			if perr != nil || id == uuid.Nil {
				validationError(c, "must be the UUID of a master department")
				return
			}
			unchanged := currentWorkingUnitID(db, c.Param("id")) == id.String()
			dept, gerr := depts.GetDepartment(c.Request.Context(), c.GetHeader("Authorization"), id.String())
			switch {
			case gerr == nil && (dept.IsActive || unchanged):
				body[workingUnitIDKey] = id.String()
				body["workingUnit"] = WorkingUnitLabel(dept.DepartmentName, dept.DepartmentCode)
			case gerr == nil:
				validationError(c, "the department is not active")
				return
			case unchanged:
				// keep the stored link and text as sent
				body[workingUnitIDKey] = id.String()
			case errors.Is(gerr, masterclient.ErrNotFound) || strings.Contains(gerr.Error(), "answered 404"):
				validationError(c, "is not a department in the master data")
				return
			default:
				response.Error(c, http.StatusServiceUnavailable, "DEPARTMENT_MASTER_UNAVAILABLE",
					"Department master data is unavailable; try again shortly", "")
				c.Abort()
				return
			}
		}

		out, err := json.Marshal(body)
		if err != nil {
			response.InternalServerError(c, "Failed to encode the request")
			c.Abort()
			return
		}
		restore(out)
		c.Next()
	}
}

func validationError(c *gin.Context, msg string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": workingUnitIDKey + " " + msg,
			"fields":  gin.H{workingUnitIDKey: msg},
		},
	})
	c.Abort()
}

// currentWorkingUnitID is the stored department id of letter id ("" when
// none, on create, or when the letter is not found)
func currentWorkingUnitID(db *gorm.DB, id string) string {
	id = strings.TrimSpace(id)
	if id == "" || db == nil {
		return ""
	}
	var l models.AssignmentLetter
	if err := db.Select("id", "working_unit_department_id").Where("id = ?", id).First(&l).Error; err != nil {
		return ""
	}
	if l.WorkingUnitDepartmentID == nil {
		return ""
	}
	return l.WorkingUnitDepartmentID.String()
}
