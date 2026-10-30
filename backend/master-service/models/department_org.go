package models

import "github.com/google/uuid"

// DepartmentOrg is the flexible-mode view of the departments table: only the
// columns added by the gated org-structure DDL. It is NEVER auto-migrated
// (parent_department_id exists only on flexible tenants), so read and write it
// with explicit Select/Update calls.
type DepartmentOrg struct {
	ID                 uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	ParentDepartmentID *uuid.UUID `gorm:"column:parent_department_id;type:uuid"`
}

func (DepartmentOrg) TableName() string { return "departments" }

// DepartmentPathSeg is one step of a department's derived path, root first,
// the department itself last. The UI joins the names with " › ".
type DepartmentPathSeg struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Level     int       `json:"level"`
	LevelName string    `json:"level_name"`
	NameKey   *string   `json:"name_key"`
}

// DepartmentTreeNode is a department with its children, for GET /departments/tree.
type DepartmentTreeNode struct {
	Department
	Children []*DepartmentTreeNode `json:"children"`
}
