package actiontakenreport

import (
	"encoding/json"
	"strings"

	"audit-service/models"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// PUT /action-taken-reports/:id
// ---------------------------------------------------------------------------

// updateRequest holds the editable ATR fields. Omitted fields are left as
// they are; due_date null or "" clears it.
type updateRequest struct {
	ActionPlan *string         `json:"action_plan"`
	Progress   *int            `json:"progress"`
	DueDate    json.RawMessage `json:"due_date"`
}

// dueDateValue reads a due_date JSON value: nil clears it, otherwise a
// normalised date. errMsg is set when the value is not a date.
func dueDateValue(raw json.RawMessage) (value interface{}, errMsg string) {
	s := strings.TrimSpace(string(raw))
	if s == "null" || s == `""` {
		return nil, ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return nil, "due_date must be a date string (YYYY-MM-DD or RFC3339) or null"
	}
	d, ok := models.ParseATRDueDate(str)
	if !ok {
		return nil, "due_date must be a date (YYYY-MM-DD or RFC3339)"
	}
	return d, ""
}

// Update edits the ATR's action plan, progress and due date in one call.
//
//   - action_plan and progress: the assigned PIC (or an admin) while the ATR is
//     PLANNED or IN_PROGRESS. Setting an action plan moves PLANNED to IN_PROGRESS,
//     as PUT /:id/action-plan does.
//   - due_date: the assign permission, on any open ATR.
//
// Report, finding, PIC and status are not editable here (the LHA, assignment,
// submit/review and cancel endpoints own them). Completed and cancelled ATRs
// cannot be changed.
func (ctl *Controller) Update(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	var req updateRequest
	if !bindJSON(c, &req) {
		return
	}
	hasPlan := req.ActionPlan != nil || req.Progress != nil
	hasDue := len(req.DueDate) > 0
	if !hasPlan && !hasDue {
		invalid(c, "Nothing to update: send action_plan, progress or due_date")
		return
	}
	if models.IsATRClosed(r.Status) {
		conflict(c, "A "+r.Status+" action taken report cannot be edited")
		return
	}

	updates := map[string]interface{}{}
	from := models.ATROpenStatuses
	if hasPlan {
		if !a.CanActAsPIC(r) {
			forbidden(c, "Only the assigned PIC can update the action plan")
			return
		}
		if !statusIn(r.Status, models.ATRStatusPlanned, models.ATRStatusInProgress) {
			conflict(c, "The action plan cannot be changed while the action taken report is "+r.Status)
			return
		}
		from = []string{models.ATRStatusPlanned, models.ATRStatusInProgress}
		if req.ActionPlan != nil {
			plan := strings.TrimSpace(*req.ActionPlan)
			if plan == "" {
				invalid(c, "action_plan is required")
				return
			}
			if len([]rune(plan)) > maxActionPlanLen {
				invalid(c, "action_plan is too long")
				return
			}
			updates["action_plan"] = plan
			updates["status"] = models.ATRStatusInProgress
		}
		if req.Progress != nil {
			if *req.Progress < 0 || *req.Progress > 100 {
				invalid(c, "progress must be between 0 and 100")
				return
			}
			updates["progress"] = *req.Progress
		}
	}
	if hasDue {
		if !a.Perms[atrsvc.PermAssign] {
			forbidden(c, "You are not allowed to change the due date of action taken reports")
			return
		}
		value, errMsg := dueDateValue(req.DueDate)
		if errMsg != "" {
			invalid(c, errMsg)
			return
		}
		updates["due_date"] = value
	}

	if !ctl.transition(c, r.ID, from, updates) {
		return
	}
	ctl.respondOne(c, a, r.ID, "Action taken report updated")
}
