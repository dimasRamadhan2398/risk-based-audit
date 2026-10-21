package seeders

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// masterSeederPath points at the master-service seeder from this package's
// directory. auth-service and master-service are separate Go modules, so the
// master seed codes are read from source rather than imported.
var masterSeederPath = filepath.Join("..", "..", "..", "..", "master-service", "pkg", "database", "seeders", "seeder.go")

var (
	// authEmployeeIDPattern matches the employee_id assignments in this package,
	// e.g. `emp001 := "EMP-0013"`.
	authEmployeeIDPattern = regexp.MustCompile(`\bemp\w*\s*:=\s*"([^"]+)"`)
	// masterEmployeeCodePattern matches the employee_code literals in the master
	// seeder, e.g. `EmployeeCode:      "EMP-0013",`.
	masterEmployeeCodePattern = regexp.MustCompile(`EmployeeCode:\s*"([^"]+)"`)
)

// TestSeedUserEmployeeIDsExistInMasterSeed guards the link between users.employee_id
// and employees.employee_code: a seeded user must point at an employee that the
// master seed also creates, otherwise the account has no employee record.
func TestSeedUserEmployeeIDsExistInMasterSeed(t *testing.T) {
	masterSrc, err := os.ReadFile(masterSeederPath)
	if os.IsNotExist(err) {
		t.Skipf("master-service seeder not found at %s (checkout layout differs)", masterSeederPath)
	}
	if err != nil {
		t.Fatalf("read master seeder: %v", err)
	}
	authSrc, err := os.ReadFile("seeder.go")
	if err != nil {
		t.Fatalf("read auth seeder: %v", err)
	}

	masterCodes := map[string]bool{}
	for _, m := range masterEmployeeCodePattern.FindAllStringSubmatch(string(masterSrc), -1) {
		masterCodes[m[1]] = true
	}
	if len(masterCodes) == 0 {
		t.Fatalf("no employee_code literals found in %s", masterSeederPath)
	}

	authIDs := authEmployeeIDPattern.FindAllStringSubmatch(string(authSrc), -1)
	// Six seeded users carry an employee_id (admin, auditor, dept_head,
	// new_dept_head, auditee, viewer). A lower count means the pattern stopped
	// matching and the check below would pass vacuously.
	if len(authIDs) != 6 {
		t.Fatalf("found %d auth employee_id assignments, want 6", len(authIDs))
	}
	for _, m := range authIDs {
		if !masterCodes[m[1]] {
			t.Errorf("auth seed employee_id %q has no matching master seed employee_code", m[1])
		}
	}
}
