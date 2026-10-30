package errors

// constraintRegistry maps Postgres constraint/index names to client-facing
// codes and fields. Names follow GORM's defaults for the current models:
// uniqueIndex -> idx_<table>_<column>, belongs-to FK -> fk_<table>_<relation>.
// Unknown constraints fall back to DUPLICATE_ENTRY / REFERENCE_NOT_FOUND.
var constraintRegistry = map[string]ConstraintInfo{
	// Unique indexes
	"idx_employees_employee_code": {
		Code: CodeEmployeeCodeAlreadyExists, Message: "An employee with this employee code already exists.", Field: "employee_code",
	},
	"idx_employees_email": {
		Code: CodeEmployeeEmailAlreadyExists, Message: "An employee with this email already exists.", Field: "email",
	},
	"idx_departments_department_code": {
		Code: CodeDepartmentCodeAlreadyExists, Message: "A department with this code already exists.", Field: "department_code",
	},
	"idx_departments_department_name": {
		Code: CodeDepartmentNameAlreadyExists, Message: "A department with this name already exists.", Field: "department_name",
	},
	// Flexible org structure (gated DDL, pkg/database/orgstructure)
	"uq_departments_company_parent_name": {
		Code: CodeDepartmentNameAlreadyExists, Message: "A unit with this name already exists under the same parent.", Field: "department_name",
	},
	"uq_departments_code_live": {
		Code: CodeDepartmentCodeAlreadyExists, Message: "A department with this code already exists.", Field: "department_code",
	},
	"uq_departments_company_code_live": {
		Code: CodeDepartmentCodeAlreadyExists, Message: "A department with this code already exists in this company.", Field: "department_code",
	},
	"uq_org_levels_level": {
		Code: CodeOrgLevelAlreadyExists, Message: "An organisation level with this number already exists.", Field: "level",
	},
	"uq_org_levels_name": {
		Code: CodeOrgLevelNameAlreadyExists, Message: "An organisation level with this name already exists.", Field: "name",
	},
	"idx_companies_company_code": {
		Code: CodeCompanyCodeAlreadyExists, Message: "A company with this code already exists.", Field: "code",
	},
	"idx_companies_tax_id": {
		Code: CodeCompanyTaxIDAlreadyExists, Message: "A company with this tax ID already exists.", Field: "tax_id",
	},

	// Foreign keys (create/update: referenced record missing)
	"fk_employees_company":         {Field: "company_id"},
	"fk_employees_department":      {Field: "department_id"},
	"fk_employees_job_role":        {Field: "job_role_id"},
	"fk_employees_work_location":   {Field: "work_location_id"},
	"fk_employees_manager":         {Field: "manager_id"},
	"fk_departments_company":       {Field: "company_id"},
	"fk_departments_business_unit": {Field: "business_unit_id"},
	"fk_departments_parent":        {Field: "parent_department_id"},
	"fk_companies_location":        {Field: "location_id"},
	"fk_companies_children":        {Field: "parent_id"},
}
