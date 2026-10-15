// Package assignmentletter holds request handling specific to
// /assignment-letters that the generic crud handlers do not cover.
package assignmentletter

import (
	"encoding/json"
	"fmt"
	"strings"

	"audit-service/models"

	"gorm.io/gorm"
)

// NormalizePurposeOnUpdate is a crud.UpdateHook for AssignmentLetter.
//
// Reads return auditPurpose filled from purposeList when the column is empty
// (models.AssignmentLetter.AfterFind), and the edit form posts that value
// back unchanged. Storing it would freeze the purpose at the old list, so an
// auditPurpose equal to the joined purposeList (before or after this update)
// is stored as empty and keeps following purposeList.
func NormalizePurposeOnUpdate(_ *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	letter, ok := existing.(*models.AssignmentLetter)
	if !ok {
		return fmt.Errorf("assignmentletter: unexpected entity %T", existing)
	}
	raw, present := columns["audit_purpose"]
	if !present {
		return nil
	}
	sent, _ := raw.(string)
	sent = strings.TrimSpace(sent)
	if sent == "" {
		return nil
	}

	derived := []string{models.JoinAuditPurposes(letter.PurposeList)}
	if rawList, present := columns["purpose_list"]; present {
		var next []string
		if b, ok := rawList.([]byte); ok && json.Unmarshal(b, &next) == nil {
			derived = append(derived, models.JoinAuditPurposes(next))
		}
	}
	for _, d := range derived {
		if d != "" && sent == d {
			columns["audit_purpose"] = ""
			return nil
		}
	}
	return nil
}
