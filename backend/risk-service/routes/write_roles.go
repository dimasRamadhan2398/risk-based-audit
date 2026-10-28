package routes

import "risk-service/middleware"

// auditUniverseWriters may write the audit universe (corporate entries,
// sync-master, yearly establish/score, events/resync) and the standard risk
// factor library. Rights are the union of a caller's roles; VIEWER, AUDITOR and
// AUDITEE never appear.
var auditUniverseWriters = []string{
	middleware.RoleAdmin,
	middleware.RoleExecutive,
	middleware.RoleChiefAuditExecutive,
}

// departmentScopedWriters is the role that is meant to edit the corporate audit
// universe and the standard libraries ONLY for its own department. It is NOT
// enforced and NOT in auditUniverseWriters, so DEPARTMENT_HEAD is refused
// (fail closed) until scoping can be built. Blocked on data that does not exist
// today: no department field on CorporateAuditUniverse, StandardAuditUniverse or
// StandardRiskFactor (models/), and no department in the JWT claims
// (middleware/auth.go Claims carries only user_id, username, roles). Once both
// exist, enforce the scope in the controller and then append this to the
// allow-list; do not append it alone, that would grant every department.
var departmentScopedWriters = []string{middleware.RoleDepartmentHead}
