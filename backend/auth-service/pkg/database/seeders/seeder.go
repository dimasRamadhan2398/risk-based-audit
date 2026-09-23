package seeders

import (
	"fmt"

	"auth-service/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Seeder struct {
	DB *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{DB: db}
}

func (s *Seeder) RunAll() error {
	if err := s.SeedPermissions(); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	if err := s.SeedRoles(); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	if err := s.SeedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	if err := s.SeedRbacMatrix(); err != nil {
		return fmt.Errorf("failed to seed RBAC matrix: %w", err)
	}

	return nil
}

// SeedPermissions seeds permission data
func (s *Seeder) SeedPermissions() error {
	permissions := []models.Permission{
		// Users & Roles
		{Name: "view_users", Resource: "users", Action: "read", Description: "View users"},
		{Name: "create_users", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "update_users", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "delete_users", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "view_roles", Resource: "roles", Action: "read", Description: "View roles"},
		{Name: "manage_roles", Resource: "roles", Action: "manage", Description: "Manage roles"},

		// Master Data
		{Name: "view_companies", Resource: "companies", Action: "read", Description: "View companies"},
		{Name: "manage_companies", Resource: "companies", Action: "manage", Description: "Manage companies"},
		{Name: "view_departments", Resource: "departments", Action: "read", Description: "View departments"},
		{Name: "manage_departments", Resource: "departments", Action: "manage", Description: "Manage departments"},
		{Name: "view_employees", Resource: "employees", Action: "read", Description: "View employees"},
		{Name: "manage_employees", Resource: "employees", Action: "manage", Description: "Manage employees"},

		// Audit Charter & Risk Profile
		{Name: "view_audit_charter", Resource: "audit_charter", Action: "read", Description: "View audit charter"},
		{Name: "manage_audit_charter", Resource: "audit_charter", Action: "manage", Description: "Manage audit charter"},
		{Name: "view_risk_register", Resource: "risk_register", Action: "read", Description: "View risk register"},
		{Name: "create_risk_register", Resource: "risk_register", Action: "create", Description: "Create risk register"},
		{Name: "update_risk_register", Resource: "risk_register", Action: "update", Description: "Update risk register"},
		{Name: "delete_risk_register", Resource: "risk_register", Action: "delete", Description: "Delete risk register"},

		// Audit Planning & Execution
		{Name: "view_strategic_plan", Resource: "strategic_plan", Action: "read", Description: "View strategic audit plan"},
		{Name: "manage_strategic_plan", Resource: "strategic_plan", Action: "manage", Description: "Manage strategic audit plan"},
		{Name: "view_annual_plan", Resource: "annual_plan", Action: "read", Description: "View annual audit plan"},
		{Name: "manage_annual_plan", Resource: "annual_plan", Action: "manage", Description: "Manage annual audit plan"},
		{Name: "view_activity_plan", Resource: "activity_plan", Action: "read", Description: "View audit activity plan"},
		{Name: "manage_activity_plan", Resource: "activity_plan", Action: "manage", Description: "Manage audit activity plan"},
		{Name: "view_assignment_letter", Resource: "assignment_letter", Action: "read", Description: "View assignment letter"},
		{Name: "manage_assignment_letter", Resource: "assignment_letter", Action: "manage", Description: "Manage assignment letter"},
		{Name: "view_working_paper", Resource: "working_paper", Action: "read", Description: "View working paper"},
		{Name: "create_working_paper", Resource: "working_paper", Action: "create", Description: "Create working paper"},
		{Name: "update_working_paper", Resource: "working_paper", Action: "update", Description: "Update working paper"},
		{Name: "delete_working_paper", Resource: "working_paper", Action: "delete", Description: "Delete working paper"},

		// Reporting & Follow-Up
		{Name: "view_audit_report", Resource: "audit_report", Action: "read", Description: "View audit result report (LHA)"},
		{Name: "manage_audit_report", Resource: "audit_report", Action: "manage", Description: "Manage audit result report (LHA)"},
		{Name: "view_executive_summary", Resource: "executive_summary", Action: "read", Description: "View executive summary report"},
		{Name: "manage_executive_summary", Resource: "executive_summary", Action: "manage", Description: "Manage executive summary report"},
		{Name: "view_action_taken_report", Resource: "action_taken_report", Action: "read", Description: "View action taken report (ATR)"},
		{Name: "update_action_taken_report", Resource: "action_taken_report", Action: "update", Description: "Update action taken report (ATR)"},

		// Executive, Performance & Quality
		{Name: "view_kpi", Resource: "kpi", Action: "read", Description: "View KPI performance"},
		{Name: "manage_kpi", Resource: "kpi", Action: "manage", Description: "Manage KPI performance"},
		{Name: "view_consulting", Resource: "consulting", Action: "read", Description: "View consulting service"},
		{Name: "manage_consulting", Resource: "consulting", Action: "manage", Description: "Manage consulting service"},
		{Name: "view_quality_assurance", Resource: "quality_assurance", Action: "read", Description: "View quality assurance"},
		{Name: "manage_quality_assurance", Resource: "quality_assurance", Action: "manage", Description: "Manage quality assurance"},
		{Name: "view_analytics", Resource: "analytics", Action: "read", Description: "View analytics dashboard"},

		// System
		{Name: "view_system_logs", Resource: "sys_alert_logs", Action: "read", Description: "View alert logs"},
		{Name: "export_system_logs", Resource: "sys_alert_logs", Action: "manage", Description: "Export alert logs"},
		{Name: "delete_system_logs", Resource: "sys_alert_logs", Action: "delete", Description: "Delete alert logs"},
	}

	for i := range permissions {
		s.DB.FirstOrCreate(&permissions[i], models.Permission{Name: permissions[i].Name})
	}

	return nil
}

// SeedRoles seeds role data
func (s *Seeder) SeedRoles() error {
	var permissions []models.Permission
	s.DB.Find(&permissions)

	permMap := make(map[string]models.Permission)
	for _, p := range permissions {
		permMap[p.Name] = p
	}

	getPerms := func(names ...string) []models.Permission {
		var list []models.Permission
		for _, name := range names {
			if p, ok := permMap[name]; ok {
				list = append(list, p)
			}
		}
		return list
	}

	roles := []models.Role{
		{
			Name:        "ADMIN",
			Description: "System Administrator with full access",
			Permissions: permissions,
		},
		{
			Name:        "AUDITOR",
			Description: "Internal Auditor with access to audit planning, fieldwork, risk registers, and reporting",
			Permissions: getPerms(
				"view_users", "view_companies", "view_departments", "view_employees",
				"view_audit_charter", "manage_audit_charter",
				"view_risk_register", "create_risk_register", "update_risk_register", "delete_risk_register",
				"view_strategic_plan", "manage_strategic_plan",
				"view_annual_plan", "manage_annual_plan",
				"view_activity_plan", "manage_activity_plan",
				"view_assignment_letter", "manage_assignment_letter",
				"view_working_paper", "create_working_paper", "update_working_paper", "delete_working_paper",
				"view_audit_report", "manage_audit_report",
				"view_executive_summary", "manage_executive_summary",
				"view_action_taken_report", "update_action_taken_report",
				"view_kpi", "manage_kpi",
				"view_consulting", "manage_consulting",
				"view_quality_assurance", "manage_quality_assurance",
				"view_analytics",
			),
		},
		{
			Name:        "EXECUTIVE",
			Description: "Executive Management with access to executive summary, analytics, and high-level reports",
			Permissions: getPerms(
				"view_users", "view_roles", "view_companies", "view_departments", "view_employees",
				"view_audit_charter", "view_risk_register",
				"view_strategic_plan", "view_annual_plan",
				"view_audit_report", "view_executive_summary", "manage_executive_summary",
				"view_action_taken_report", "view_kpi", "view_quality_assurance",
				"view_analytics", "view_system_logs",
			),
		},
		{
			Name:        "AUDITEE",
			Description: "Auditee unit responsible for viewing assigned findings and submitting Action Taken Reports",
			Permissions: getPerms(
				"view_users", "view_companies", "view_departments",
				"view_audit_charter", "view_risk_register",
				"view_action_taken_report", "update_action_taken_report",
			),
		},
		{
			Name:        "VIEWER",
			Description: "Viewer with read-only access to published reports and dashboards",
			Permissions: getPerms(
				"view_users", "view_companies", "view_departments",
				"view_audit_charter", "view_risk_register",
				"view_audit_report", "view_executive_summary", "view_analytics",
			),
		},
		{
			Name:        "DEPARTMENT_HEAD",
			Description: "Head of Department overseeing department risk and audit actions",
			Permissions: getPerms(
				"view_users", "view_companies", "view_departments", "view_employees",
				"view_risk_register", "create_risk_register", "update_risk_register", "delete_risk_register",
				"view_audit_report", "view_action_taken_report", "update_action_taken_report",
			),
		},
	}

	for i := range roles {
		var existing models.Role
		if err := s.DB.Where("name = ?", roles[i].Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			s.DB.Create(&roles[i])
		} else {
			s.DB.Model(&existing).Association("Permissions").Replace(roles[i].Permissions)
		}
	}

	return nil
}

// SeedUsers seeds user data
func (s *Seeder) SeedUsers() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []models.User{
		{
			EmployeeID:   "EMP001",
			Username:     "admin",
			Email:        "admin@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "System Administrator",
			Phone:        "+6281234567890",
			Department:   "IT",
			IsActive:     true,
		},
		{
			EmployeeID:   "EMP002",
			Username:     "auditor",
			Email:        "auditor@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "John Auditor",
			Phone:        "+6281234567891",
			Department:   "Internal Audit",
			IsActive:     true,
		},
		{
			EmployeeID:   "EMP003",
			Username:     "dept_head",
			Email:        "depthead@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Alice Department Head",
			Phone:        "+6281234567892",
			Department:   "Operations",
			IsActive:     true,
		},
		{
			EmployeeID:   "EMP007",
			Username:     "new_dept_head",
			Email:        "new_dept_head@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "New Department Head",
			Phone:        "+6281234567897",
			Department:   "Finance",
			IsActive:     true,
		},
		{
			EmployeeID:   "EMP004",
			Username:     "auditee",
			Email:        "auditee@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Bob Auditee",
			Phone:        "+6281234567893",
			Department:   "Finance",
			IsActive:     true,
		},
		{
			EmployeeID:   "EMP005",
			Username:     "viewer",
			Email:        "viewer@company.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Carol Viewer",
			Phone:        "+6281234567894",
			Department:   "Management",
			IsActive:     true,
		},
	}

	// Create users if they don't exist
	for i := range users {
		var existing models.User

		if err := s.DB.Where("username = ?", users[i].Username).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := s.DB.Create(&users[i]).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", users[i].Username, err)
			}
		}
	}

	// Fetch all roles
	var (
		adminRole     models.Role
		auditorRole   models.Role
		deptHeadRole  models.Role
		auditeeRole   models.Role
		viewerRole    models.Role
		executiveRole models.Role
	)

	roles := map[string]*models.Role{
		"ADMIN":           &adminRole,
		"AUDITOR":         &auditorRole,
		"DEPARTMENT_HEAD": &deptHeadRole,
		"AUDITEE":         &auditeeRole,
		"VIEWER":          &viewerRole,
		"EXECUTIVE":       &executiveRole,
	}

	for name, role := range roles {
		if err := s.DB.Where("name = ?", name).First(role).Error; err != nil {
			return fmt.Errorf("failed to find role %s: %w", name, err)
		}
	}

	// Fetch all users
	var (
		adminUser       models.User
		auditorUser     models.User
		deptHeadUser    models.User
		newDeptHeadUser models.User
		auditeeUser     models.User
		viewerUser      models.User
	)

	userRoleMap := map[string]*models.User{
		"admin":         &adminUser,
		"auditor":       &auditorUser,
		"dept_head":     &deptHeadUser,
		"new_dept_head": &newDeptHeadUser,
		"auditee":       &auditeeUser,
		"viewer":        &viewerUser,
	}

	for username, user := range userRoleMap {
		if err := s.DB.Where("username = ?", username).First(user).Error; err != nil {
			return fmt.Errorf("failed to find user %s: %w", username, err)
		}
	}

	// Assign roles to users
	assignments := []struct {
		user *models.User
		role *models.Role
	}{
		{&adminUser, &adminRole},
		{&auditorUser, &auditorRole},
		{&deptHeadUser, &deptHeadRole},
		{&newDeptHeadUser, &deptHeadRole},
		{&auditeeUser, &auditeeRole},
		{&viewerUser, &viewerRole},
	}

	for _, a := range assignments {
		if err := s.DB.Model(a.user).Association("Roles").Append(a.role); err != nil {
			return fmt.Errorf("failed to assign role to user %s: %w", a.user.Username, err)
		}
	}

	return nil
}

// SeedRbacMatrix seeds the AuditSphere RBAC matrix permissions feature set
func (s *Seeder) SeedRbacMatrix() error {
	features := []models.RbacMatrixFeature{
		{FeatureNumber: 1, Module: "Dashboard", Submodule: "Dashboard Main Overview", FeatureCode: "dashboard", Description: "Dashboard access & overview", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 2, Module: "Audit Charter", Submodule: "Audit Charter Framework", FeatureCode: "audit_charter", Description: "Audit charter framework", AdminAccess: "FULL", AuditManagerAccess: "READ", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 3, Module: "Risk Profile", Submodule: "Corporate Risk Profile", FeatureCode: "risk_profile_corporate", Description: "Corporate risk profile mapping", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "FULL", AuditeeAccess: "EDIT — Own Dept.", ViewerAccess: "READ"},
		{FeatureNumber: 3, Module: "Risk Profile", Submodule: "Risk Appetite Statement", FeatureCode: "risk_profile_appetite", Description: "Risk appetite statement", AdminAccess: "FULL", AuditManagerAccess: "READ", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 3, Module: "Risk Profile", Submodule: "Risk Factors & Scoring", FeatureCode: "risk_profile_factors", Description: "Risk scoring parameters", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 3, Module: "Risk Profile", Submodule: "Audit Universe", FeatureCode: "risk_profile_universe", Description: "Audit universe entities", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 3, Module: "Risk Profile", Submodule: "Risk Control Matrix (RCM)", FeatureCode: "risk_profile_rcm", Description: "Risk control matrix", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "FULL", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 4, Module: "Strategic Audit Plan", Submodule: "Strategic Audit Plan", FeatureCode: "strategic_audit_plan", Description: "Strategic audit plan", AdminAccess: "FULL", AuditManagerAccess: "EDIT", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 5, Module: "Annual Audit Plan", Submodule: "Create / Manage Plan", FeatureCode: "annual_plan_manage", Description: "Annual plan creation & management", AdminAccess: "FULL", AuditManagerAccess: "DRAFT", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 5, Module: "Annual Audit Plan", Submodule: "Import Plan Document", FeatureCode: "annual_plan_import", Description: "Import annual plan document", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "NONE", AuditeeAccess: "NONE", ViewerAccess: "NONE"},
		{FeatureNumber: 5, Module: "Annual Audit Plan", Submodule: "Execution Status", FeatureCode: "annual_plan_status", Description: "Annual plan execution status", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 6, Module: "Audit Activity Plan", Submodule: "Create Activity Plan", FeatureCode: "activity_plan_create", Description: "Create activity plan", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "DRAFT", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 6, Module: "Audit Activity Plan", Submodule: "Import Activity Plan", FeatureCode: "activity_plan_import", Description: "Import activity plan", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "NONE", AuditeeAccess: "NONE", ViewerAccess: "NONE"},
		{FeatureNumber: 7, Module: "Assignment Letter", Submodule: "Create / Publish Letter", FeatureCode: "assignment_letter_publish", Description: "Create/publish assignment letter", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 7, Module: "Assignment Letter", Submodule: "Import Letter", FeatureCode: "assignment_letter_import", Description: "Import assignment letter", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "NONE", AuditeeAccess: "NONE", ViewerAccess: "NONE"},
		{FeatureNumber: 8, Module: "Audit Fieldwork", Submodule: "Fieldwork — Interviews, Sampling, Testing, etc.", FeatureCode: "fieldwork_testing", Description: "Fieldwork interviews & testing", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "FULL", AuditeeAccess: "RESPOND", ViewerAccess: "READ"},
		{FeatureNumber: 8, Module: "Audit Fieldwork", Submodule: "Working Papers — F-01 to F-05", FeatureCode: "fieldwork_working_papers", Description: "Working papers F-01 to F-05", AdminAccess: "FULL", AuditManagerAccess: "REVIEW", AuditorAccess: "CREATE", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 8, Module: "Audit Fieldwork", Submodule: "Import Working Papers", FeatureCode: "fieldwork_import_papers", Description: "Import working papers", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "UPLOAD", AuditeeAccess: "NONE", ViewerAccess: "NONE"},
		{FeatureNumber: 9, Module: "Audit Result Report", Submodule: "Result Report / LHA", FeatureCode: "report_lha", Description: "Result report / LHA", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "DRAFT", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 9, Module: "Audit Result Report", Submodule: "Executive Summary", FeatureCode: "report_executive_summary", Description: "Executive summary report", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 9, Module: "Audit Result Report", Submodule: "Client Satisfaction Survey", FeatureCode: "report_survey", Description: "Client satisfaction survey", AdminAccess: "FULL", AuditManagerAccess: "READ", AuditorAccess: "READ", AuditeeAccess: "FILL SURVEY", ViewerAccess: "READ"},
		{FeatureNumber: 10, Module: "Action Taken Report (ATR)", Submodule: "Action Taken Report (ATR)", FeatureCode: "action_taken_report", Description: "Action taken report & CAPA tracking", AdminAccess: "FULL", AuditManagerAccess: "REVIEW", AuditorAccess: "READ", AuditeeAccess: "UPDATE CAPA", ViewerAccess: "READ"},
		{FeatureNumber: 11, Module: "KPI Performance", Submodule: "KPI Performance", FeatureCode: "kpi_performance", Description: "KPI performance metrics", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 12, Module: "Consulting Service", Submodule: "Consulting Service", FeatureCode: "consulting_service", Description: "Consulting service requests", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "FULL", AuditeeAccess: "REQUEST", ViewerAccess: "READ"},
		{FeatureNumber: 13, Module: "Quality Assurance (QA)", Submodule: "Quality Assurance (QA)", FeatureCode: "quality_assurance", Description: "Quality assurance reviews", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "NONE", ViewerAccess: "READ"},
		{FeatureNumber: 14, Module: "Analytics", Submodule: "Analytics", FeatureCode: "analytics", Description: "Analytics & audit dashboards", AdminAccess: "FULL", AuditManagerAccess: "FULL", AuditorAccess: "READ", AuditeeAccess: "READ", ViewerAccess: "READ"},
		{FeatureNumber: 15, Module: "Settings & User Management", Submodule: "Settings & User Management", FeatureCode: "settings_user_management", Description: "Settings & RBAC management", AdminAccess: "FULL", AuditManagerAccess: "NONE", AuditorAccess: "NONE", AuditeeAccess: "NONE", ViewerAccess: "NONE"},
	}

	for i := range features {
		var existing models.RbacMatrixFeature
		if err := s.DB.Where("feature_code = ?", features[i].FeatureCode).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := s.DB.Create(&features[i]).Error; err != nil {
				return fmt.Errorf("failed to create rbac feature %s: %w", features[i].FeatureCode, err)
			}
		} else {
			s.DB.Model(&existing).Updates(&features[i])
		}
	}

	return nil
}
