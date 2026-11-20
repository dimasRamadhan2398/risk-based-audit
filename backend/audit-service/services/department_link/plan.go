// Package department_link links legacy assignment letters and LHAs to
// master-service departments by id (`audit migrate backfill-department-ids`).
//
// A letter is linked when its Unit Kerja text names exactly one master
// department: the id itself, the code, the name or "Name (CODE)" (trimmed,
// whitespace collapsed, case-insensitive). An LHA takes its letter's
// department (existing or planned); an LHA whose letter has none is matched
// on its own department text with the same rule. Ambiguous and unmatched
// values are reported and left alone. Only rows whose id is still NULL are
// considered, so re-running is a no-op.
package department_link

import (
	"sort"
	"strings"

	"audit-service/models"
	"audit-service/pkg/masterclient"

	"github.com/google/uuid"
)

// Link is one planned update
type Link struct {
	Table        string // assignment_letters | audit_result_reports
	ID           uuid.UUID
	Ref          string // letter number / report number
	FromValue    string
	DepartmentID uuid.UUID
	Via          string // text | letter
}

// Miss is a row that stays unlinked
type Miss struct {
	Table     string
	ID        uuid.UUID
	Ref       string
	FromValue string
	Reason    string
}

func normalize(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

type index struct {
	byKey map[string]map[uuid.UUID]bool
}

func newIndex(depts []masterclient.Department) *index {
	ix := &index{byKey: map[string]map[uuid.UUID]bool{}}
	add := func(key string, id uuid.UUID) {
		if key = normalize(key); key == "" {
			return
		}
		if ix.byKey[key] == nil {
			ix.byKey[key] = map[uuid.UUID]bool{}
		}
		ix.byKey[key][id] = true
	}
	for _, d := range depts {
		id, err := uuid.Parse(strings.TrimSpace(d.ID))
		if err != nil || id == uuid.Nil {
			continue
		}
		add(id.String(), id)
		add(d.DepartmentCode, id)
		add(d.DepartmentName, id)
		if c := strings.TrimSpace(d.DepartmentCode); c != "" {
			add(strings.TrimSpace(d.DepartmentName)+" ("+c+")", id)
		}
	}
	return ix
}

// match returns the single department text names, or a reason
func (ix *index) match(text string) (uuid.UUID, string) {
	key := normalize(text)
	if key == "" {
		return uuid.Nil, "no department text"
	}
	ids := ix.byKey[key]
	switch len(ids) {
	case 0:
		return uuid.Nil, "no master department with this id, code or name"
	case 1:
		for id := range ids {
			return id, ""
		}
	}
	return uuid.Nil, "matches more than one master department"
}

// Plan decides which letters and reports can be linked
func Plan(letters []models.AssignmentLetter, reports []models.AuditResultReport, depts []masterclient.Department) ([]Link, []Miss) {
	ix := newIndex(depts)
	var links []Link
	var misses []Miss

	// department of each letter (existing id or planned), by number and id
	letterDept := map[string]uuid.UUID{}
	remember := func(l models.AssignmentLetter, id uuid.UUID) {
		letterDept[strings.TrimSpace(l.LetterNumber)] = id
		letterDept[l.ID.String()] = id
	}
	for _, l := range letters {
		if l.WorkingUnitDepartmentID != nil {
			remember(l, *l.WorkingUnitDepartmentID)
			continue
		}
		id, reason := ix.match(l.WorkingUnit)
		if reason != "" {
			misses = append(misses, Miss{Table: "assignment_letters", ID: l.ID, Ref: l.LetterNumber, FromValue: l.WorkingUnit, Reason: reason})
			continue
		}
		remember(l, id)
		links = append(links, Link{Table: "assignment_letters", ID: l.ID, Ref: l.LetterNumber, FromValue: l.WorkingUnit, DepartmentID: id, Via: "text"})
	}

	for _, r := range reports {
		if r.DepartmentID != nil {
			continue
		}
		if id, ok := letterDept[strings.TrimSpace(r.AssignmentLetterID)]; ok {
			links = append(links, Link{Table: "audit_result_reports", ID: r.ID, Ref: r.ReportNumber, FromValue: r.AssignmentLetterID, DepartmentID: id, Via: "letter"})
			continue
		}
		id, reason := ix.match(r.Department)
		if reason != "" {
			misses = append(misses, Miss{Table: "audit_result_reports", ID: r.ID, Ref: r.ReportNumber, FromValue: r.Department, Reason: reason})
			continue
		}
		links = append(links, Link{Table: "audit_result_reports", ID: r.ID, Ref: r.ReportNumber, FromValue: r.Department, DepartmentID: id, Via: "text"})
	}

	sort.SliceStable(links, func(i, j int) bool {
		if links[i].Table != links[j].Table {
			return links[i].Table < links[j].Table
		}
		return links[i].ID.String() < links[j].ID.String()
	})
	sort.SliceStable(misses, func(i, j int) bool {
		if misses[i].Table != misses[j].Table {
			return misses[i].Table < misses[j].Table
		}
		return misses[i].ID.String() < misses[j].ID.String()
	})
	return links, misses
}
