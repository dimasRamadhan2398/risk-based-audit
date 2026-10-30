// Package orgmode resolves which organisation-structure model this process
// serves (docs/org-structure-model-plan.md section 3.2).
//
// Effective mode = ORG_STRUCTURE_MODE=flexible AND the org_levels table exists.
// Everything else is legacy: flat departments, no org-structure table access.
// The mode is resolved once per process; nothing here runs DDL.
package orgmode

import (
	"fmt"

	"gorm.io/gorm"
)

// Mode is the effective organisation-structure mode of a tenant.
type Mode string

const (
	Legacy   Mode = "legacy"
	Flexible Mode = "flexible"
)

// MaxLevels is the maximum number of live org levels, and therefore the
// maximum depth of the department tree.
const MaxLevels = 6

// LevelsTable is the persisted marker of a flexible tenant. It is created only
// by the gated DDL in pkg/database/orgstructure.
const LevelsTable = "org_levels"

// IsFlexible reports whether m is the flexible mode.
func (m Mode) IsFlexible() bool { return m == Flexible }

// Resolve returns the effective mode. flagOn is the parsed ORG_STRUCTURE_MODE
// (config.OrgStructureConfig.FlexibleRequested). The only database access is
// one catalog lookup for the org_levels table. warn, when not nil, receives a
// message for the two drift cases (flag without table, table without flag).
func Resolve(db *gorm.DB, flagOn bool, warn func(string)) Mode {
	if warn == nil {
		warn = func(string) {}
	}
	hasTable := db != nil && db.Migrator().HasTable(LevelsTable)
	switch {
	case flagOn && hasTable:
		return Flexible
	case flagOn && !hasTable:
		warn(fmt.Sprintf("ORG_STRUCTURE_MODE=flexible but table %s does not exist: serving legacy until `master migrate` has run", LevelsTable))
	case !flagOn && hasTable:
		warn(fmt.Sprintf("table %s exists but ORG_STRUCTURE_MODE is not flexible: serving legacy (flag drift)", LevelsTable))
	}
	return Legacy
}
