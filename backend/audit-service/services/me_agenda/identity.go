// Package me_agenda builds the Auditor dashboard's calendar and to-do list
// (GET /api/v1/me/calendar and /api/v1/me/todos).
//
// Team membership of an assignment letter is stored as names (Leader,
// MembersList) with optional employee ids, so deciding "is the caller on this
// audit" is a matching problem; identity.go holds it. The rest of the package is
// pure derivation over rows the loader fetched, so it can be tested without a
// database.
package me_agenda

import (
	"strings"
	"unicode"

	"audit-service/models"
)

// Caller is the logged-in user as far as audit-service can tell.
type Caller struct {
	UserID       string
	Username     string
	EmployeeID   string // JWT employee_id: master-service employee UUID
	EmployeeCode string // JWT employee_code: employees.employee_code
	FullName     string // from auth-service; empty when it could not be resolved
}

// LinkStatus says how much the caller can be matched on.
type LinkStatus string

const (
	// LinkEmployee: the account is linked to an employee (id and/or code).
	LinkEmployee LinkStatus = "employee"
	// LinkNameOnly: no employee link, but a display name or a two-word username
	// is available for name matching.
	LinkNameOnly LinkStatus = "name_only"
	// LinkNone: nothing to match on; mine is always false.
	LinkNone LinkStatus = "none"
)

// Role of a caller on an audit team.
const (
	RoleLeader = "LEADER"
	RoleMember = "MEMBER"
)

// How a membership was established, strongest first.
const (
	ViaEmployeeID   = "employee_id"
	ViaEmployeeCode = "employee_code"
	ViaName         = "name"
	ViaUsername     = "username"
)

// Membership is the result of matching a caller against one letter. Role is ""
// when the caller is not on the team.
type Membership struct {
	Role string
	Via  string
}

// Mine reports whether the caller is on the team.
func (m Membership) Mine() bool { return m.Role != "" }

// honorifics dropped from the start of a name before comparing.
var honorifics = map[string]bool{
	"dr": true, "drs": true, "dra": true, "ir": true, "prof": true,
	"h": true, "hj": true, "bapak": true, "ibu": true, "pak": true, "bu": true,
	"mr": true, "mrs": true, "ms": true,
}

// NormalizeName reduces a person's name to a comparison key: lower case,
// anything after the first comma dropped (Indonesian degrees are written
// "Budi Santoso, S.E., M.M."), punctuation turned into spaces, leading
// honorifics (Dr., Ir., H., ...) and extra spaces removed. Word order is kept:
// "Santoso Budi" does not match "Budi Santoso".
func NormalizeName(s string) string {
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for len(words) > 1 && honorifics[words[0]] {
		words = words[1:]
	}
	return strings.Join(words, " ")
}

func tokenCount(normalized string) int {
	if normalized == "" {
		return 0
	}
	return strings.Count(normalized, " ") + 1
}

// nameKeys are the normalized names the caller may be listed under, with the
// way each was obtained. A username only counts when it reads as a full name
// ("budi.santoso"); a single word ("andi", "admin") is too ambiguous.
type nameKeys struct {
	full     string
	username string
}

func (c Caller) names() nameKeys {
	k := nameKeys{full: NormalizeName(c.FullName)}
	if u := NormalizeName(c.Username); tokenCount(u) >= 2 {
		k.username = u
	}
	return k
}

// Link reports how the caller can be matched.
func (c Caller) Link() LinkStatus {
	if strings.TrimSpace(c.EmployeeID) != "" || strings.TrimSpace(c.EmployeeCode) != "" {
		return LinkEmployee
	}
	if k := c.names(); k.full != "" || k.username != "" {
		return LinkNameOnly
	}
	return LinkNone
}

func sameID(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	return a != "" && strings.EqualFold(a, b)
}

// matchPerson matches the caller against one listed person (a member or the
// leader). Ids win: a person listed with an id is only ever matched by id, so
// two employees who share a name cannot be mixed up. A person without an id is
// matched by normalized full name, then by a two-word username.
func (c Caller) matchPerson(name, employeeID, employeeCode string) (via string, ok bool) {
	hasID := strings.TrimSpace(employeeID) != "" || strings.TrimSpace(employeeCode) != ""
	if hasID {
		switch {
		case sameID(employeeID, c.EmployeeID):
			return ViaEmployeeID, true
		case sameID(employeeCode, c.EmployeeCode):
			return ViaEmployeeCode, true
		}
		return "", false
	}
	n := NormalizeName(name)
	if n == "" {
		return "", false
	}
	k := c.names()
	if k.full != "" && n == k.full {
		return ViaName, true
	}
	if k.username != "" && n == k.username {
		return ViaUsername, true
	}
	return "", false
}

var viaRank = map[string]int{ViaEmployeeID: 4, ViaEmployeeCode: 3, ViaName: 2, ViaUsername: 1}

// MatchLetter decides whether the caller is the leader or a member of the
// letter. When the caller is listed as both, LEADER wins; among several
// matches the strongest method is reported.
func MatchLetter(c Caller, l *models.AssignmentLetter) Membership {
	var best Membership
	consider := func(role, via string) {
		if best.Role == "" || (role == RoleLeader && best.Role != RoleLeader) ||
			(role == best.Role && viaRank[via] > viaRank[best.Via]) {
			best = Membership{Role: role, Via: via}
		}
	}
	if via, ok := c.matchPerson(l.Leader, l.LeaderEmployeeID, l.LeaderEmployeeCode); ok {
		consider(RoleLeader, via)
	}
	for _, m := range l.MembersList {
		if via, ok := c.matchPerson(m.Name, m.EmployeeID, m.EmployeeCode); ok {
			role := RoleMember
			if strings.EqualFold(strings.TrimSpace(m.Role), "leader") ||
				strings.EqualFold(strings.TrimSpace(m.Role), "chairperson") ||
				strings.EqualFold(strings.TrimSpace(m.Role), "ketua tim") {
				role = RoleLeader
			}
			consider(role, via)
		}
	}
	return best
}
