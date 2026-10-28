package routes

import "risk-service/middleware"

// auditUniverseWriters may write the audit universe (corporate entries,
// sync-master, yearly establish/score, events/resync) and the standard risk
// factor library without department scoping. Rights are the union of a
// caller's roles; VIEWER, AUDITOR and AUDITEE never appear.
var auditUniverseWriters = middleware.UniverseFullWriters

// departmentScopedWriters may additionally reach ONLY the corporate audit
// universe write routes (POST /corporate, DELETE /corporate/:id). The
// controller (controllers/audit_universe_controller.go, authorizeNodeWrite)
// then limits them to entries of their own department and fails closed when
// the department cannot be resolved. Do not add this role to any other group:
// sync-master, year establish/score, events/resync and the standard libraries
// are global and stay on auditUniverseWriters. Flipping the role off for the
// corporate routes is removing it from corporateWriters below.
var departmentScopedWriters = []string{middleware.RoleDepartmentHead}

// corporateWriters is who reaches the corporate audit universe write handlers.
var corporateWriters = append(append([]string{}, auditUniverseWriters...), departmentScopedWriters...)
