// Package executivesummary serves the read-only aggregates behind the
// "Statistik Kompilasi" and "Status Tindak Lanjut" sections of the Executive
// Summary (compilation) form. Nothing here is stored: every value is counted
// from the source tables for one audit year.
package executivesummary

import (
	"strconv"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
)

// Follow-up status labels, as the frontend table uses them
const (
	FollowUpClosed     = "Closed"
	FollowUpInProgress = "In Progress"
	FollowUpOverdue    = "Overdue"
)

// FollowUpRow is one row of the follow-up status table
type FollowUpRow struct {
	Status string `json:"status"`
	Jumlah int64  `json:"jumlah"`
}

// CompilationStats is the data payload of GET /executive-summary/compilation-stats
type CompilationStats struct {
	Year              int   `json:"year"`
	JumlahLaporan     int64 `json:"jumlahLaporan"`
	JumlahRekomendasi int64 `json:"jumlahRekomendasi"`
	// Breakdown by the five working paper risk levels (see RiskBucket);
	// RisikoLainnya counts empty or unrecognised levels
	RisikoHigh         int64         `json:"risikoHigh"`
	RisikoModerateHigh int64         `json:"risikoModerateHigh"`
	RisikoModerate     int64         `json:"risikoModerate"`
	RisikoLowModerate  int64         `json:"risikoLowModerate"`
	RisikoLow          int64         `json:"risikoLow"`
	RisikoLainnya      int64         `json:"risikoLainnya"`
	FollowUpTable      []FollowUpRow `json:"followUpTable"`
	FollowUpCancelled  int64         `json:"followUpCancelled"`
}

// letterKeysForYear returns the assignment letter keys (letter numbers) of the
// year. Child rows (LHA, test controls, working paper risks) reference their
// letter by letter number, the same key the findings derivation uses. Letters
// are scoped with assignment_letters.audit_year, as the performance statistics do.
func letterKeysForYear(db *gorm.DB, year int) ([]string, error) {
	var keys []string
	err := db.Model(&models.AssignmentLetter{}).
		Where("audit_year = ?", strconv.Itoa(year)).
		Pluck("letter_number", &keys).Error
	return keys, err
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

// LoadCompilationStats counts the five values for the given audit year. now is
// the clock for the overdue test of the action taken reports.
func LoadCompilationStats(db *gorm.DB, year int, now time.Time) (CompilationStats, error) {
	stats := CompilationStats{
		Year: year,
		FollowUpTable: []FollowUpRow{
			{Status: FollowUpClosed},
			{Status: FollowUpInProgress},
			{Status: FollowUpOverdue},
		},
	}

	keys, err := letterKeysForYear(db, year)
	if err != nil {
		return stats, err
	}
	if len(keys) == 0 {
		return stats, nil
	}

	// Jumlah Laporan: non-deleted LHA of the year's letters
	if err := db.Model(&models.AuditResultReport{}).
		Where("assignment_letter_id IN ?", keys).
		Count(&stats.JumlahLaporan).Error; err != nil {
		return stats, err
	}

	// Jumlah Rekomendasi: fieldwork test controls that carry a recommendation
	if err := db.Model(&models.FieldworkTestControl{}).
		Where("assignment_letter_id IN ? AND COALESCE(TRIM(recommendation), '') <> ''", keys).
		Count(&stats.JumlahRekomendasi).Error; err != nil {
		return stats, err
	}

	// Breakdown Temuan Berdasarkan Risiko: working paper (KKA F02) risk rows by level
	var risks []models.WorkingPaperRisk
	if err := db.Select("risk_level").
		Where("working_paper_id IN ?", keys).
		Find(&risks).Error; err != nil {
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

	// Status Tindak Lanjut: ATRs of the year's LHAs (ATR -> LHA -> letter).
	// ATRs without an LHA link (legacy rows) are outside any year.
	var atrs []struct {
		Status  string
		DueDate *time.Time
	}
	if err := db.Model(&models.ActionTakenReport{}).
		Select("action_taken_reports.status, action_taken_reports.due_date").
		Joins("JOIN audit_result_reports ON audit_result_reports.id = action_taken_reports.audit_result_report_id AND audit_result_reports.deleted_at IS NULL").
		Where("audit_result_reports.assignment_letter_id IN ?", keys).
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
