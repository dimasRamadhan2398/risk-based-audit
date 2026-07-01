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

	return nil
}

// SeedPermissions seeds permission data
func (s *Seeder) SeedPermissions() error {
	permissions := []models.Permission{
		{Name: "view_users", Resource: "users", Action: "read", Description: "View users"},
		{Name: "create_users", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "update_users", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "delete_users", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "view_roles", Resource: "roles", Action: "read", Description: "View roles"},
		{Name: "manage_roles", Resource: "roles", Action: "manage", Description: "Manage roles"},
		{Name: "view_companies", Resource: "companies", Action: "read", Description: "View companies"},
		{Name: "manage_companies", Resource: "companies", Action: "manage", Description: "Manage companies"},
		{Name: "view_departments", Resource: "departments", Action: "read", Description: "View departments"},
		{Name: "manage_departments", Resource: "departments", Action: "manage", Description: "Manage departments"},
		{Name: "view_employees", Resource: "employees", Action: "read", Description: "View employees"},
		{Name: "manage_employees", Resource: "employees", Action: "manage", Description: "Manage employees"},
		{Name: "view_risk_register", Resource: "risk_register", Action: "read", Description: "View risk register"},
		{Name: "create_risk_register", Resource: "risk_register", Action: "create", Description: "Create risk register"},
		{Name: "update_risk_register", Resource: "risk_register", Action: "update", Description: "Update risk register"},
		{Name: "delete_risk_register", Resource: "risk_register", Action: "delete", Description: "Delete risk register"},
		{Name: "view_audit_charter", Resource: "audit_charter", Action: "read", Description: "View audit charter"},
		{Name: "manage_audit_charter", Resource: "audit_charter", Action: "manage", Description: "Manage audit charter"},
		{Name: "manage_audit_charter", Resource: "audit_charter", Action: "manage", Description: "Manage audit charter"},
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

	roles := []models.Role{
		{
			Name:        "ADMIN",
			Description: "System Administrator",
			Permissions: permissions,
		},
		{
			Name:        "AUDITOR",
			Description: "Internal Auditor",
			Permissions: []models.Permission{
				permissions[0],  // view_users
				permissions[6],  // view_companies
				permissions[8],  // view_departments
				permissions[10], // view_employees
				permissions[12], // view_risk_register
				permissions[13], // create_risk_register
				permissions[14], // update_risk_register
				permissions[16], // view_audit_charter
			},
		},
		{
			Name:        "DEPARTMENT_HEAD",
			Description: "Head of Department who is responsible of the subordinates taking part in the audit process ",
			Permissions: []models.Permission{
				permissions[6],  // view_companies
				permissions[8],  // view_departments
				permissions[10], // view_employees
				permissions[12], // view_risk_register
				permissions[13], // create_risk_register
				permissions[14], // update_risk_register
				permissions[15], // delete_risk_register
			},
		},
		{
			Name:        "AUDITEE",
			Description: "Auditee is the unit that is audited by the auditor",
			Permissions: []models.Permission{
				permissions[0], // view_users
				permissions[6], // view_companies
				permissions[8], // view_departments
			},
		},
		{
			Name:        "VIEWER",
			Description: "Viewer with very limited access",
			Permissions: []models.Permission{
				permissions[0], // view_users
				permissions[6], // view_companies
				permissions[8], // view_departments
			},
		},
		{
			Name:        "EXECUTIVE",
			Description: "Executive management with high-level access",
			Permissions: []models.Permission{
				permissions[0],  // view_users
				permissions[4],  // view_roles
				permissions[6],  // view_companies
				permissions[8],  // view_departments
				permissions[10], // view_employees
				permissions[12], // view_risk_register
				permissions[16], // view_audit_charter
				permissions[19], // view_system_logs
			},
		},
	}

	for i := range roles {
		s.DB.FirstOrCreate(&roles[i], models.Role{Name: roles[i].Name})
		if roles[i].Name == "ADMIN" {
			s.DB.Model(&roles[i]).Association("Permissions").Replace(roles[i].Permissions)
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
		adminUser    models.User
		auditorUser  models.User
		deptHeadUser models.User
		auditeeUser  models.User
		viewerUser   models.User
	)

	userRoleMap := map[string]*models.User{
		"admin":     &adminUser,
		"auditor":   &auditorUser,
		"dept_head": &deptHeadUser,
		"auditee":   &auditeeUser,
		"viewer":    &viewerUser,
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
