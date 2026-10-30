package assignmentletter

import (
	atrsvc "audit-service/services/action_taken_report"
)

// PublisherRoles are the canonical roles (services/action_taken_report RBAC)
// that may publish a letter or otherwise change its status: Admin, Audit
// Manager and the Chief Audit Executive. The CAE is the seeded EXECUTIVE role
// in every tenant; CHIEF_AUDIT_EXECUTIVE (and spellings such as "CAE") only
// exist when an admin creates them, and all of them canonicalise to RoleCAE.
// Auditors and team leaders may draft and edit letters but not publish them.
//
// The frontend mirrors this set in useRbac canPublishAssignmentLetter.
var PublisherRoles = []string{atrsvc.RoleAdmin, atrsvc.RoleAuditManager, atrsvc.RoleCAE}

// CanPublish reports whether a caller with the given JWT roles may change a
// letter's status
func CanPublish(roles []string) bool {
	return atrsvc.HasCanonicalRole(roles, PublisherRoles...)
}
