// Package executivesummary serves the read-only aggregates behind the
// "Statistik Kompilasi" and "Status Tindak Lanjut" sections of the Executive
// Summary (compilation) form. Nothing here is stored: every value is counted
// from the source tables for one audit year, or for one quarter of it.
//
// Scope, from the top:
//   - LHA (audit_result_reports): approved only (models.LHAApprovedStatuses).
//     Without a quarter, the LHAs of the year's assignment letters (letters
//     with audit_year = year, as the performance statistics do). With a
//     quarter, the LHAs whose report_date falls in that quarter (UTC, as the
//     date is stored); an LHA without a report date is in the year but in no
//     quarter.
//   - Everything else is derived from the in-scope LHAs: the assignment letters
//     they reference (by letter number or letter UUID), the working paper risks
//     and test controls of those letters (each row counted once), and the
//     action taken reports of those LHAs.
package executivesummary

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Follow-up status labels, as the frontend table uses them
const (
	FollowUpClosed     = "Closed"
	FollowUpInProgress = "In Progress"
	FollowUpOverdue    = "Overdue"
)

// ErrInvalidQuarter is returned for a quarter outside 1-4 (0 means the full year)
var ErrInvalidQuarter = errors.New("quarter must be between 1 and 4")

// FollowUpRow is one row of the follow-up status table
type FollowUpRow struct {
	Status string `json:"status"`
	Jumlah int64  `json:"jumlah"`
}

// CompilationStats is the data payload of GET /executive-summary/compilation-stats
type CompilationStats struct {
	Year int `json:"year"`
	// Quarter is 1-4, or 0 for the full year
	Quarter int `json:"quarter"`
	// SyncedAt is the time the figures were counted
	SyncedAt          time.Time `json:"syncedAt"`
	JumlahLaporan     int64     `json:"jumlahLaporan"`
	JumlahRekomendasi int64     `json:"jumlahRekomendasi"`
	// Breakdown by the five working paper risk levels (see RiskBucket);
	// RisikoLainnya counts empty or unrecognised levels and is not part of TotalTemuan
	TotalTemuan        int64         `json:"totalTemuan"`
	RisikoHigh         int64         `json:"risikoHigh"`
	RisikoModerateHigh int64         `json:"risikoModerateHigh"`
	RisikoModerate     int64         `json:"risikoModerate"`
	RisikoLowModerate  int64         `json:"risikoLowModerate"`
	RisikoLow          int64         `json:"risikoLow"`
	RisikoLainnya      int64         `json:"risikoLainnya"`
	FollowUpTable      []FollowUpRow `json:"followUpTable"`
	FollowUpCancelled  int64         `json:"followUpCancelled"`
}

// Risk levels of the compilation, as the frontend stores them in
// working_paper_risks.risk_level
const (
	RiskHigh         = "high"
	RiskModerateHigh = "moderate-high"
	RiskModerate     = "moderate"
	RiskLowModerate  = "low-moderate"
	RiskLow          = "low"
)

// RiskBucket maps a working paper risk level onto one of the five levels,
// case-insensitively; "_", "-" and repeated spaces are equivalent. Legacy
// three-level values map to high / moderate / low. ok is false for an empty or
// unrecognised level.
func RiskBucket(raw string) (bucket string, ok bool) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.Join(strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(s)), " ")
	switch s {
	case "HIGH", "CRITICAL", "VERY HIGH", "TINGGI", "SANGAT TINGGI":
		return RiskHigh, true
	case "MODERATE HIGH", "MODERATE TO HIGH", "MEDIUM HIGH", "MEDIUM TO HIGH":
		return RiskModerateHigh, true
	case "MEDIUM", "MODERATE", "SEDANG":
		return RiskModerate, true
	case "LOW MODERATE", "LOW TO MODERATE", "LOW MEDIUM", "LOW TO MEDIUM":
		return RiskLowModerate, true
	case "LOW", "RENDAH", "SANGAT RENDAH":
		return RiskLow, true
	}
	return "", false
}

// quarterBounds returns the half-open UTC interval [start, end) of a quarter.
// report_date is stored as UTC midnight for date-only input, so UTC bounds
// select the calendar quarter the frontend reads from the date string.
func quarterBounds(year, quarter int) (start, end time.Time) {
	start = time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 3, 0)
}

// lhaScope is the id and letter reference of one in-scope LHA
type lhaScope struct {
	ID                 uuid.UUID
	AssignmentLetterID string
}

// letterKeysForYear returns every key the year's letters are referenced by:
// the letter number and the letter UUID (LHAs, test controls and working paper
// risks reference a letter by either).
func letterKeysForYear(db *gorm.DB, year int) ([]string, error) {
	var letters []models.AssignmentLetter
	if err := db.Select("id", "letter_number").
		Where("audit_year = ?", strconv.Itoa(year)).
		Find(&letters).Error; err != nil {
		return nil, err
	}
	keys := newKeySet()
	for _, l := range letters {
		keys.add(l.LetterNumber)
		keys.add(l.ID.String())
	}
	return keys.list, nil
}

// letterKeysFor expands the letter references of LHAs to every key of their
// letter: the reference itself plus the number and UUID of the live letter it
// resolves to. A reference that resolves to nothing (deleted or unknown letter)
// is kept as is.
func letterKeysFor(db *gorm.DB, refs []string) ([]string, error) {
	keys := newKeySet()
	var uuids []uuid.UUID
	for _, ref := range refs {
		keys.add(ref)
		if u, err := uuid.Parse(ref); err == nil {
			uuids = append(uuids, u)
		}
	}
	if len(refs) == 0 {
		return keys.list, nil
	}

	q := db.Select("id", "letter_number")
	if len(uuids) > 0 {
		q = q.Where("(letter_number IN ? OR id IN ?)", refs, uuids)
	} else {
		q = q.Where("letter_number IN ?", refs)
	}
	var letters []models.AssignmentLetter
	if err := q.Find(&letters).Error; err != nil {
		return nil, err
	}
	for _, l := range letters {
		keys.add(l.LetterNumber)
		keys.add(l.ID.String())
	}
	return keys.list, nil
}

// approvedLHAs returns the approved, non-deleted LHAs in scope (see the package
// comment). An empty year yields no rows.
func approvedLHAs(db *gorm.DB, year, quarter int) ([]lhaScope, error) {
	q := db.Model(&models.AuditResultReport{}).
		Select("id, assignment_letter_id").
		Where("UPPER(TRIM(status)) IN ?", models.LHAApprovedStatuses)

	if quarter == 0 {
		keys, err := letterKeysForYear(db, year)
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			return nil, nil
		}
		q = q.Where("assignment_letter_id IN ?", keys)
	} else {
		start, end := quarterBounds(year, quarter)
		q = q.Where("report_date >= ? AND report_date < ?", start, end)
	}

	var rows []lhaScope
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// LoadCompilationStats counts the statistics for the given audit year and
// quarter (0 = full year). now is the clock for the overdue test of the action
// taken reports and becomes SyncedAt. Every query is unpaged, so no count is
// cut off by a page size.
func LoadCompilationStats(db *gorm.DB, year, quarter int, now time.Time) (CompilationStats, error) {
	stats := CompilationStats{
		Year:     year,
		Quarter:  quarter,
		SyncedAt: now,
		FollowUpTable: []FollowUpRow{
			{Status: FollowUpClosed},
			{Status: FollowUpInProgress},
			{Status: FollowUpOverdue},
		},
	}
	if quarter < 0 || quarter > 4 {
		return stats, ErrInvalidQuarter
	}

	// Jumlah Laporan: approved LHAs in scope. Everything below derives from them.
	lhas, err := approvedLHAs(db, year, quarter)
	if err != nil {
		return stats, err
	}
	stats.JumlahLaporan = int64(len(lhas))
	if len(lhas) == 0 {
		return stats, nil
	}

	reportIDs := make([]uuid.UUID, 0, len(lhas))
	refs := make([]string, 0, len(lhas))
	seenRef := map[string]bool{}
	for _, l := range lhas {
		reportIDs = append(reportIDs, l.ID)
		if ref := strings.TrimSpace(l.AssignmentLetterID); ref != "" && !seenRef[ref] {
			seenRef[ref] = true
			refs = append(refs, ref)
		}
	}

	// Keys of the letters of the in-scope LHAs. Each key set is distinct, so a
	// test control or risk row is matched, and counted, once even when several
	// LHAs share its letter.
	keys, err := letterKeysFor(db, refs)
	if err != nil {
		return stats, err
	}

	// Jumlah Rekomendasi: test controls with a non-blank recommendation, per letter
	if len(keys) > 0 {
		if err := db.Model(&models.FieldworkTestControl{}).
			Where("assignment_letter_id IN ? AND COALESCE(TRIM(recommendation), '') <> ''", keys).
			Count(&stats.JumlahRekomendasi).Error; err != nil {
			return stats, err
		}
	}

	// Breakdown Temuan Berdasarkan Risiko: working paper (KKA F02) risks by level
	if len(keys) > 0 {
		var risks []struct{ RiskLevel string }
		if err := db.Model(&models.WorkingPaperRisk{}).
			Select("COALESCE(risk_level, '') AS risk_level").
			Where("working_paper_id IN ?", keys).
			Scan(&risks).Error; err != nil {
			return stats, err
		}
		for _, r := range risks {
			switch bucket, _ := RiskBucket(r.RiskLevel); bucket {
			case RiskHigh:
				stats.RisikoHigh++
			case RiskModerateHigh:
				stats.RisikoModerateHigh++
			case RiskModerate:
				stats.RisikoModerate++
			case RiskLowModerate:
				stats.RisikoLowModerate++
			case RiskLow:
				stats.RisikoLow++
			default:
				stats.RisikoLainnya++
			}
		}
		stats.TotalTemuan = stats.RisikoHigh + stats.RisikoModerateHigh + stats.RisikoModerate + stats.RisikoLowModerate + stats.RisikoLow
	}

	// Status Tindak Lanjut: ATRs of the in-scope LHAs (one per finding)
	var atrs []struct {
		Status  string
		DueDate *time.Time
	}
	if err := db.Model(&models.ActionTakenReport{}).
		Select("COALESCE(status, '') AS status, due_date").
		Where("audit_result_report_id IN ?", reportIDs).
		Scan(&atrs).Error; err != nil {
		return stats, err
	}
	for _, a := range atrs {
		status, _ := models.NormalizeATRStatus(a.Status)
		switch {
		case status == models.ATRStatusCompleted:
			stats.FollowUpTable[0].Jumlah++
		case status == models.ATRStatusCancelled:
			stats.FollowUpCancelled++
		default:
			// Any open status (or an unrecognised legacy value) is open
			if overdue, _ := models.ComputeATROverdue(a.DueDate, a.Status, now); overdue {
				stats.FollowUpTable[2].Jumlah++
			} else {
				stats.FollowUpTable[1].Jumlah++
			}
		}
	}

	return stats, nil
}

// keySet is an insertion-ordered set of non-blank strings
type keySet struct {
	seen map[string]bool
	list []string
}

func newKeySet() *keySet {
	return &keySet{seen: map[string]bool{}}
}

func (k *keySet) add(v string) {
	v = strings.TrimSpace(v)
	if v == "" || k.seen[v] {
		return
	}
	k.seen[v] = true
	k.list = append(k.list, v)
}
