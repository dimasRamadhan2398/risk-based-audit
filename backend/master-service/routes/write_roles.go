package routes

import "master-service/pkg/middleware"

// Who may write, by route group. Reads are not governed here. Rights are the
// union of a caller's roles; VIEWER never appears, and is refused entirely by
// DenyViewer unless the caller also holds another role.
var (
	// masterDataWriters: companies, business-units, departments (incl.
	// events/resync), employees, job-roles, locations, vision-mission-goals.
	masterDataWriters = []string{middleware.RoleAdmin}

	// auditReportWriters: quality-assurance and consulting-services
	// (create, update, delete, import).
	auditReportWriters = []string{
		middleware.RoleAdmin,
		middleware.RoleExecutive,
		middleware.RoleChiefAuditExecutive,
		middleware.RoleAuditManager,
	}
)

// Data sources are deliberately not listed: their write rule is undecided and
// they stay open to every non-VIEWER role.
