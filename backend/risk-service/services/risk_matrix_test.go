package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestRiskLevelMatrixParity compares RiskLevelFor with getRiskLevel in
// frontend/stores/risk-profile.ts cell by cell (plan gap G3). It reads the
// TypeScript source, so a change on either side fails here. Skipped only when
// the frontend tree is not present (e.g. a service-only build context).
func TestRiskLevelMatrixParity(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "stores", "risk-profile.ts")
	src, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("%s not present", path)
	}
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	start := strings.Index(s, "const getRiskLevel")
	if start < 0 {
		t.Fatal("getRiskLevel not found in risk-profile.ts")
	}
	end := strings.Index(s[start:], "return matrix")
	if end < 0 {
		t.Fatal("getRiskLevel matrix end not found")
	}
	block := s[start : start+end]

	rowRe := regexp.MustCompile(`(\d)\s*:\s*\{([^}]*)\}`)
	cellRe := regexp.MustCompile(`(\d)\s*:\s*RiskLevel\.([A-Z_]+)`)
	rows := rowRe.FindAllStringSubmatch(block, -1)
	if len(rows) != 5 {
		t.Fatalf("parsed %d matrix rows from the frontend, want 5", len(rows))
	}
	cells := 0
	for _, row := range rows {
		lik, _ := strconv.Atoi(row[1])
		for _, cell := range cellRe.FindAllStringSubmatch(row[2], -1) {
			imp, _ := strconv.Atoi(cell[1])
			if got := RiskLevelFor(lik, imp); got != cell[2] {
				t.Errorf("likelihood %d impact %d: backend %s, frontend %s", lik, imp, got, cell[2])
			}
			cells++
		}
	}
	if cells != 25 {
		t.Fatalf("parsed %d cells, want 25", cells)
	}
	if !strings.Contains(block, "if (!likelihood || !impact) return RiskLevel.LOW") {
		t.Error("frontend no longer returns LOW for missing likelihood/impact; update RiskLevelFor")
	}
}

func TestRiskLevelForOutOfRange(t *testing.T) {
	for _, c := range [][2]int{{0, 3}, {3, 0}, {6, 1}, {1, 6}, {-1, -1}} {
		if got := RiskLevelFor(c[0], c[1]); got != RiskLevelCodeLow {
			t.Errorf("RiskLevelFor(%d,%d) = %s, want LOW", c[0], c[1], got)
		}
	}
}

func TestWithinAppetite(t *testing.T) {
	want := map[string]bool{
		RiskLevelCodeLow: true, RiskLevelCodeLowModerate: true,
		RiskLevelCodeModerate: false, RiskLevelCodeModerateHigh: false, RiskLevelCodeHigh: false,
	}
	for lvl, w := range want {
		if WithinAppetite(lvl) != w {
			t.Errorf("WithinAppetite(%s) = %v, want %v", lvl, !w, w)
		}
	}
}
