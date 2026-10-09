// Package workingpaper holds request handling specific to /working-papers
// that the generic crud handlers do not cover.
package workingpaper

import (
	"fmt"
	"strings"

	"audit-service/models"

	"gorm.io/gorm"
)

// SyncHeaderOnUpdate is a crud.UpdateHook for WorkingPaperHeader. The header's
// audit purpose comes from its Assignment Letter (the field is read-only in
// the UI):
//   - when the update links a different letter, audit_purpose is set to that
//     letter's purpose, replacing the previous letter's value (cleared when
//     the new letter has none, so a stale purpose is not kept);
//   - otherwise, when the purpose after this update would be empty, it is
//     filled from the linked letter.
//
// A non-empty purpose sent for the same letter is kept. Unknown letters leave
// the columns untouched.
func SyncHeaderOnUpdate(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	header, ok := existing.(*models.WorkingPaperHeader)
	if !ok {
		return fmt.Errorf("workingpaper: unexpected entity %T", existing)
	}

	letterRef := strings.TrimSpace(header.AssignmentLetterID)
	letterChanged := false
	if raw, present := columns["assignment_letter_id"]; present {
		s, _ := raw.(string)
		s = strings.TrimSpace(s)
		letterChanged = s != letterRef
		letterRef = s
	}
	if letterRef == "" {
		return nil
	}

	// existing.AuditPurpose may already hold the letter's purpose filled in by
	// AfterFind; a purpose sent in the request takes precedence.
	purpose := header.AuditPurpose
	if raw, present := columns["audit_purpose"]; present {
		purpose, _ = raw.(string)
	}
	if !letterChanged && strings.TrimSpace(purpose) != "" {
		return nil
	}

	letter, found := models.FindAssignmentLetter(tx, letterRef)
	if !found {
		return nil
	}
	if p := letter.EffectiveAuditPurpose(); p != "" || letterChanged {
		columns["audit_purpose"] = p
	}
	return nil
}
