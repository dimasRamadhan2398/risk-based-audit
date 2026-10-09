// Package activitycode generates Activity IDs of the form
//
//	{audit type code}-{year}-{sequence}   e.g. ASR-2026-003
//
// where the sequence is the n-th activity of that audit type in that year,
// zero-padded to 3 digits (it simply grows to 4 digits after 999).
//
// The audit type is the audit category used across the app (frontend
// AuditCategory). Short codes:
//
//	Assurance                 ASR
//	Special Audit             SPC
//	Specific Reason           SPR
//	Consulting Services       CNS
//	Investigation             INV
//	Quality Assurance Review  QAR
//	Follow-Up Audit           FUA
//	(empty)                   AUD
//
// Matching ignores case and punctuation and also accepts the enum keys
// (SPECIAL_AUDIT, CONSULTING_SERVICES, ...) and short forms seen in stored data
// (CONSULTING, SPECIAL, QAR, FOLLOW UP). Any other value gets a derived code:
// the initials of a multi-word value or the first three letters of a single
// word, uppercased; "AUD" if that would clash with one of the codes above.
//
// Numbers come from the activity_code_sequences counter, bumped with an
// INSERT ... ON CONFLICT DO UPDATE ... RETURNING inside the caller's
// transaction. That statement takes a row lock on the prefix until commit, so
// concurrent creates are serialised per prefix, and a rollback returns the
// number. The counter never goes down, so deleting an activity does not free
// its number. When bumping, the counter is first raised to the highest code
// already stored for the prefix, so codes written outside the counter (seed
// data) are never handed out again.
package activitycode

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
)

// GenericTypeCode is used when the audit type is empty.
const GenericTypeCode = "AUD"

var knownTypeCodes = map[string]string{
	"ASSURANCE":                "ASR",
	"SPECIAL AUDIT":            "SPC",
	"SPECIAL":                  "SPC",
	"SPECIFIC REASON":          "SPR",
	"CONSULTING SERVICES":      "CNS",
	"CONSULTING SERVICE":       "CNS",
	"CONSULTING":               "CNS",
	"INVESTIGATION":            "INV",
	"QUALITY ASSURANCE REVIEW": "QAR",
	"QAR":                      "QAR",
	"FOLLOW UP AUDIT":          "FUA",
	"FOLLOWUP AUDIT":           "FUA",
	"FOLLOW UP":                "FUA",
}

var reservedCodes = func() map[string]bool {
	m := map[string]bool{GenericTypeCode: true}
	for _, c := range knownTypeCodes {
		m[c] = true
	}
	return m
}()

var (
	nonAlnum = regexp.MustCompile(`[^A-Z0-9]+`)
	codeRe   = regexp.MustCompile(`^([A-Z0-9]{1,6})-(\d{4})-(\d{3,})$`)
)

// TypeCode returns the short code for an audit type / category.
func TypeCode(auditType string) string {
	norm := strings.TrimSpace(nonAlnum.ReplaceAllString(strings.ToUpper(auditType), " "))
	if norm == "" {
		return GenericTypeCode
	}
	if c, ok := knownTypeCodes[norm]; ok {
		return c
	}
	// "Quality Assurance Review (QAR)" and similar labels with a suffix.
	if c, ok := knownTypeCodes[strings.TrimSuffix(norm, " QAR")]; ok {
		return c
	}
	words := strings.Fields(norm)
	var code string
	if len(words) > 1 {
		for _, w := range words {
			code += w[:1]
		}
	} else {
		code = words[0]
	}
	if len(code) > 3 && len(words) == 1 {
		code = code[:3]
	}
	if len(code) > 6 {
		code = code[:6]
	}
	if reservedCodes[code] {
		return GenericTypeCode
	}
	return code
}

// Prefix is the "{type}-{year}" part a sequence is counted under.
func Prefix(typeCode string, year int) string {
	return fmt.Sprintf("%s-%04d", typeCode, year)
}

// Format builds the full Activity ID.
func Format(typeCode string, year, seq int) string {
	return fmt.Sprintf("%s-%03d", Prefix(typeCode, year), seq)
}

// Parse splits an Activity ID. ok is false for anything not in the format,
// e.g. the legacy "ACT-2026-001" style is parsed (ACT, 2026, 1) but a client
// row key such as "1728371234567" is not.
func Parse(code string) (typeCode string, year, seq int, ok bool) {
	m := codeRe.FindStringSubmatch(strings.TrimSpace(code))
	if m == nil {
		return "", 0, 0, false
	}
	year, _ = strconv.Atoi(m[2])
	seq, err := strconv.Atoi(m[3])
	if err != nil {
		return "", 0, 0, false
	}
	return m[1], year, seq, true
}

var yearRe = regexp.MustCompile(`(19|20|21)\d{2}`)

// PlanYear picks the year an activity is counted under: the first value that
// contains a 4-digit year, in the order given (plan year, then period start,
// ...). If none does, the current year.
func PlanYear(candidates ...string) int {
	for _, c := range candidates {
		if y := yearRe.FindString(c); y != "" {
			n, _ := strconv.Atoi(y)
			return n
		}
	}
	return time.Now().Year()
}

// Reserve hands out n consecutive numbers for prefix and returns the first.
// Call it inside the transaction that stores the codes.
func Reserve(tx *gorm.DB, prefix string, n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("activitycode: reserve %d numbers", n)
	}
	floor, err := highestStored(tx, prefix)
	if err != nil {
		return 0, err
	}
	var last int
	// Fixed SQL; prefix is server-derived ([A-Z0-9-]) and passed as a parameter.
	err = tx.Raw(`INSERT INTO activity_code_sequences (prefix, last_value, updated_at)
VALUES (?, CAST(? AS INTEGER), ?)
ON CONFLICT (prefix) DO UPDATE SET
  last_value = CASE WHEN activity_code_sequences.last_value < CAST(? AS INTEGER)
                    THEN CAST(? AS INTEGER)
                    ELSE activity_code_sequences.last_value END + CAST(? AS INTEGER),
  updated_at = excluded.updated_at
RETURNING last_value`,
		prefix, floor+n, time.Now().UTC(), floor, floor, n).Scan(&last).Error
	if err != nil {
		return 0, fmt.Errorf("activitycode: reserve %s: %w", prefix, err)
	}
	return last - n + 1, nil
}

// Next returns one new Activity ID for the audit type and year.
func Next(tx *gorm.DB, auditType string, year int) (string, error) {
	typeCode := TypeCode(auditType)
	seq, err := Reserve(tx, Prefix(typeCode, year), 1)
	if err != nil {
		return "", err
	}
	return Format(typeCode, year, seq), nil
}

// Request is one code to hand out: the audit type of the activity.
type Request struct{ AuditType string }

// NextMany returns one new Activity ID per request, in request order, all in
// the given year. Numbers are reserved per prefix in sorted prefix order, so
// two transactions needing several prefixes cannot deadlock on each other.
func NextMany(tx *gorm.DB, year int, reqs []Request) ([]string, error) {
	out := make([]string, len(reqs))
	if len(reqs) == 0 {
		return out, nil
	}
	byPrefix := map[string][]int{}
	typeOf := map[string]string{}
	for i, r := range reqs {
		tc := TypeCode(r.AuditType)
		p := Prefix(tc, year)
		byPrefix[p] = append(byPrefix[p], i)
		typeOf[p] = tc
	}
	prefixes := make([]string, 0, len(byPrefix))
	for p := range byPrefix {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)
	for _, p := range prefixes {
		idx := byPrefix[p]
		first, err := Reserve(tx, p, len(idx))
		if err != nil {
			return nil, err
		}
		for k, i := range idx {
			out[i] = Format(typeOf[p], year, first+k)
		}
	}
	return out, nil
}

// highestStored is the largest sequence already used under prefix, in either
// audit_activities.project_code or an activity_plans planned activity, soft
// deleted rows included (project_code's unique index covers them too).
func highestStored(tx *gorm.DB, prefix string) (int, error) {
	// prefix is [A-Z0-9-] only, so it has no LIKE wildcards.
	pattern := prefix + "-%"
	highest := 0
	note := func(code string) {
		if !strings.HasPrefix(code, prefix+"-") {
			return
		}
		if _, _, seq, ok := Parse(code); ok && seq > highest {
			highest = seq
		}
	}

	var codes []string
	if err := tx.Unscoped().Model(&models.AuditActivity{}).
		Where("project_code LIKE ?", pattern).
		Pluck("project_code", &codes).Error; err != nil {
		return 0, fmt.Errorf("activitycode: read audit activity codes: %w", err)
	}
	for _, c := range codes {
		note(c)
	}

	var blobs []string
	if err := tx.Unscoped().Model(&models.ActivityPlan{}).
		Where("planned_activities LIKE ?", "%"+pattern).
		Pluck("planned_activities", &blobs).Error; err != nil {
		return 0, fmt.Errorf("activitycode: read activity plan codes: %w", err)
	}
	for _, b := range blobs {
		var acts []struct {
			ActivityCode string `json:"activityCode"`
		}
		if json.Unmarshal([]byte(b), &acts) != nil {
			continue
		}
		for _, a := range acts {
			note(a.ActivityCode)
		}
	}
	return highest, nil
}
