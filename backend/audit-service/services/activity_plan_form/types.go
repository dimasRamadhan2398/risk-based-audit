package activityplanform

import (
	"strings"

	"audit-service/models"
	"audit-service/pkg/activitycode"
)

// typeCodes maps the RKA Jenis Kegiatan Audit to the activitycode type code of
// the PKAT category it accepts (contract 4.3).
var typeCodes = map[string]string{
	models.ActivityTypeAssurance:     "ASR", // Assurance
	models.ActivityTypeSpecialAudit:  "SPC", // Special Audit
	models.ActivityTypeSpecificAudit: "SPR", // Specific Reason
	models.ActivityTypeInvestigation: "INV", // Investigation
}

// ValidActivityType reports whether t is one of models.ActivityPlanActivityTypes
// (exact match after trimming).
func ValidActivityType(t string) bool {
	_, ok := typeCodes[strings.TrimSpace(t)]
	return ok
}

// CategoryMatches reports whether a PKAT row category may be added to an RKA of
// the given Jenis Kegiatan Audit. Categories are compared by their
// activitycode.TypeCode, so legacy spellings such as "ASSURANCE" or
// "SPECIAL_AUDIT" match too. Categories without an RKA type (Consulting
// Services, Quality Assurance Review, Follow-Up Audit, empty) never match.
func CategoryMatches(activityType, category string) bool {
	want, ok := typeCodes[strings.TrimSpace(activityType)]
	if !ok || strings.TrimSpace(category) == "" {
		return false
	}
	return activitycode.TypeCode(category) == want
}
