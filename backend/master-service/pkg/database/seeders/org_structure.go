package seeders

import (
	"fmt"

	"master-service/models"

	"github.com/google/uuid"
)

// Flexible-mode demo departments (docs/org-structure-model-plan.md 4.2 and
// 4.2.1). Only reached with ORG_STRUCTURE_MODE=flexible; idempotent by
// department_code, and an existing row is never changed except that a NULL
// parent is filled in (so admin edits survive a re-run).

func (s *Seeder) liveOrgLevels() ([]models.OrgLevel, error) {
	var levels []models.OrgLevel
	if err := s.DB.Where("is_active = ?", true).Order("level ASC").Find(&levels).Error; err != nil {
		return nil, fmt.Errorf("load org levels: %w", err)
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("ORG_STRUCTURE_MODE=flexible but org_levels is empty: run `master seed-org-levels --preset ...` first")
	}
	return levels, nil
}

// departmentLevel is the level of the "department" key, else the deepest level.
func departmentLevel(levels []models.OrgLevel) int {
	for _, l := range levels {
		if l.NameKey != nil && *l.NameKey == models.OrgLevelKeyDepartment {
			return l.Level
		}
	}
	return levels[len(levels)-1].Level
}

// bankLevels maps name_key -> level when all four bank keys exist.
func bankLevels(levels []models.OrgLevel) (map[string]int, bool) {
	out := map[string]int{}
	for _, l := range levels {
		if l.NameKey != nil {
			out[*l.NameKey] = l.Level
		}
	}
	for _, k := range models.OrgLevelNameKeys {
		if _, ok := out[k]; !ok {
			return nil, false
		}
	}
	return out, true
}

func (s *Seeder) seedOrgDepartments(seeds []models.Department, companyID, corpBU, opsBU uuid.UUID) error {
	levels, err := s.liveOrgLevels()
	if err != nil {
		return err
	}
	deptLevel := departmentLevel(levels)
	for i := range seeds {
		seeds[i].Level = deptLevel
		if err := s.findOrCreateDepartment(&seeds[i]); err != nil {
			return err
		}
	}
	bank, ok := bankLevels(levels)
	if !ok {
		return nil
	}
	return s.seedBankTree(bank, companyID, corpBU, opsBU)
}

type treeSeed struct {
	code, name, parent, key string
	bu                      string // "corp" or "ops"
}

// bankTree is the D7 demo: a level skip (Departemen directly under a
// Direktorat) and the same leaf name under two parents.
var bankTree = []treeSeed{
	{code: "DIR-UTM", name: "Direktorat Utama", key: models.OrgLevelKeyDirectorate, bu: "corp"},
	{code: "DIR-KEU", name: "Direktorat Keuangan", key: models.OrgLevelKeyDirectorate, bu: "corp"},
	{code: "DIR-OPS", name: "Direktorat Operasional", key: models.OrgLevelKeyDirectorate, bu: "ops"},
	{code: "DIV-AKT", name: "Divisi Akuntansi", parent: "DIR-KEU", key: models.OrgLevelKeyDivision, bu: "corp"},
	{code: "DIV-TRS", name: "Divisi Treasury", parent: "DIR-KEU", key: models.OrgLevelKeyDivision, bu: "corp"},
	{code: "DIV-TI", name: "Divisi Teknologi Informasi", parent: "DIR-OPS", key: models.OrgLevelKeyDivision, bu: "corp"},
	{code: "DEPT-003", name: "Internal Audit", parent: "DIR-UTM", key: models.OrgLevelKeyDepartment, bu: "corp"},
	{code: "DEPT-001", name: "Finance & Accounting", parent: "DIV-AKT", key: models.OrgLevelKeyDepartment, bu: "corp"},
	{code: "DEPT-PJK-AKT", name: "Departemen Pajak", parent: "DIV-AKT", key: models.OrgLevelKeyDepartment, bu: "corp"},
	{code: "DEPT-PJK-TRS", name: "Departemen Pajak", parent: "DIV-TRS", key: models.OrgLevelKeyDepartment, bu: "corp"},
	{code: "DEPT-004", name: "Operations", parent: "DIR-OPS", key: models.OrgLevelKeyDepartment, bu: "ops"},
	{code: "DEPT-002", name: "Information Technology", parent: "DIV-TI", key: models.OrgLevelKeyDepartment, bu: "corp"},
	{code: "SKS-INFRA", name: "Seksi Infrastruktur", parent: "DEPT-002", key: models.OrgLevelKeySection, bu: "corp"},
}

func (s *Seeder) seedBankTree(bank map[string]int, companyID, corpBU, opsBU uuid.UUID) error {
	ids := map[string]uuid.UUID{}
	for _, t := range bankTree {
		bu := corpBU
		if t.bu == "ops" {
			bu = opsBU
		}
		d := models.Department{
			DepartmentCode:        t.code,
			DepartmentName:        t.name,
			DepartmentDescription: t.name,
			Level:                 bank[t.key],
			CompanyID:             companyID,
			BusinessUnitID:        &bu,
			IsActive:              true,
		}
		if err := s.findOrCreateDepartment(&d); err != nil {
			return fmt.Errorf("seed %s: %w", t.code, err)
		}
		ids[t.code] = d.ID
		if t.parent == "" {
			continue
		}
		parentID, ok := ids[t.parent]
		if !ok {
			return fmt.Errorf("seed %s: parent %s not seeded first", t.code, t.parent)
		}
		// Fill the parent only where none is set (never overwrite an edit).
		if err := s.DB.Model(&models.DepartmentOrg{}).
			Where("id = ? AND parent_department_id IS NULL", d.ID).
			Update("parent_department_id", parentID).Error; err != nil {
			return fmt.Errorf("seed %s parent: %w", t.code, err)
		}
	}
	return nil
}

// findOrCreateDepartment loads the live department with d's code into d, or
// creates d (with a client-side id) when none exists. Existing rows are never
// modified. (FirstOrCreate would also match on a preset id and miss the row.)
func (s *Seeder) findOrCreateDepartment(d *models.Department) error {
	var existing models.Department
	err := s.DB.Where("department_code = ?", d.DepartmentCode).Limit(1).Find(&existing).Error
	if err != nil {
		return err
	}
	if existing.ID != uuid.Nil {
		*d = existing
		return nil
	}
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return s.DB.Omit("Company", "BusinessUnit").Create(d).Error
}
