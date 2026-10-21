package executivesummary

import (
	"net/http"
	"strings"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"gorm.io/gorm"
)

// Executive summary statuses. Draft -> Approved is only possible through
// POST /executive-summaries/:id/approve, which requires a signature.
const (
	StatusDraft    = "Draft"
	StatusApproved = "Approved"
	StatusRejected = "Rejected"
)

const (
	errCodeApprovedLocked  = "EXECUTIVE_SUMMARY_APPROVED"
	errCodeApproveByStatus = "APPROVE_REQUIRES_SIGNATURE"
	approveEndpointHint    = "use POST /executive-summaries/:id/approve to approve with a signature"
	approvedLockedMessage  = "An approved executive summary cannot be changed"
)

// serverManagedColumns are written only by the approve endpoint. A create or
// update payload that sets them is ignored, so an approval signature cannot be
// forged or overwritten through the generic CRUD routes.
var serverManagedColumns = []string{
	"signature_image",
	"signature_name",
	"signature_role",
	"approved_by",
	"approved_by_role",
	"approved_at",
}

// isApprovedStatus reports whether status is Approved (any casing)
func isApprovedStatus(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), StatusApproved)
}

// GuardCreateHook is the crud.CreateHook of POST /executive-summaries. A new
// summary cannot start as Approved, and server-managed approval fields are
// cleared.
func GuardCreateHook(_ *gorm.DB, entity interface{}) error {
	e, ok := entity.(*models.ExecutiveSummary)
	if !ok {
		return nil
	}
	if isApprovedStatus(e.Status) {
		return apperrors.New(errCodeApproveByStatus, approveEndpointHint, http.StatusBadRequest)
	}
	e.SignatureImage = ""
	e.SignatureName = ""
	e.SignatureRole = ""
	e.ApprovedBy = ""
	e.ApprovedByRole = ""
	e.ApprovedAt = nil
	return nil
}

// GuardUpdateHook is the first part of the crud.UpdateHook of PUT
// /executive-summaries/:id:
//   - an Approved summary is locked: any update is rejected with 409, so the
//     signature and approval record cannot be overwritten;
//   - a payload that sets status Approved is rejected with 400 (approval goes
//     through the approve endpoint);
//   - server-managed approval columns are removed from the payload.
func GuardUpdateHook(_ *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	e, ok := existing.(*models.ExecutiveSummary)
	if !ok {
		return nil
	}
	if isApprovedStatus(e.Status) {
		return apperrors.New(errCodeApprovedLocked, approvedLockedMessage, http.StatusConflict)
	}
	if raw, ok := columns["status"]; ok {
		if s, isString := raw.(string); isString && isApprovedStatus(s) {
			return apperrors.New(errCodeApproveByStatus, approveEndpointHint, http.StatusBadRequest)
		}
	}
	for _, col := range serverManagedColumns {
		delete(columns, col)
	}
	return nil
}

// UpdateHook is the crud.UpdateHook of PUT /executive-summaries/:id: the
// status guard first, then the risk count sync.
func UpdateHook(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	if err := GuardUpdateHook(tx, existing, columns); err != nil {
		return err
	}
	return SyncRiskColumnsBeforeUpdate(tx, existing, columns)
}
