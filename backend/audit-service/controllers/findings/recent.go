package findings

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	DefaultRecentLimit = 5
	MaxRecentLimit     = 50
)

// RecentItem is one entry of GET /audit-result-reports/recent-findings
type RecentItem struct {
	Title              string  `json:"title"`
	Category           string  `json:"category"`
	Action             string  `json:"action"`
	Source             string  `json:"source"`
	AssignmentLetterID string  `json:"assignmentLetterId"`
	ReportID           *string `json:"reportId"`
	Date               string  `json:"date"` // RFC3339, UTC
}

// RecentResult is the data payload of the recent-findings response
type RecentResult struct {
	Items []RecentItem `json:"items"`
	Total int          `json:"total"`
}

// NormalizeCategory maps the category/severity vocabularies used across the
// app onto the four categories the dashboard shows. Matching ignores case,
// surrounding whitespace and '_'/'-' separators.
//
//	Very Significant, HIGH, CRITICAL, VERY HIGH     -> Very Significant
//	Significant, MEDIUM, MODERATE                   -> Significant
//	Quite Significant, Moderately Significant, LOW  -> Quite Significant
//	Not Significant, Insignificant                  -> Not Significant
//	empty / anything else                           -> Significant
func NormalizeCategory(raw string) string {
	s := strings.NewReplacer("_", " ", "-", " ").Replace(raw)
	switch strings.ToUpper(strings.Join(strings.Fields(s), " ")) {
	case "VERY SIGNIFICANT", "HIGH", "CRITICAL", "VERY HIGH":
		return "Very Significant"
	case "SIGNIFICANT", "MEDIUM", "MODERATE":
		return "Significant"
	case "QUITE SIGNIFICANT", "MODERATELY SIGNIFICANT", "LOW":
		return "Quite Significant"
	case "NOT SIGNIFICANT", "INSIGNIFICANT":
		return "Not Significant"
	default:
		return "Significant"
	}
}

// ParseRecentLimit returns the limit query value, or the default when it is
// missing, not a number or outside [1, MaxRecentLimit].
func ParseRecentLimit(raw string) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 || n > MaxRecentLimit {
		return DefaultRecentLimit
	}
	return n
}

// LetterFindings is the live (KKA + fieldwork) findings of one assignment letter
type LetterFindings struct {
	AssignmentLetterID string
	Findings           []Finding
}

type candidate struct {
	item       RecentItem
	date       time.Time
	fromReport bool
}

// MergeRecent merges saved-report findings with live findings, dedupes them
// by (assignmentLetterId, lower-cased trimmed title), sorts them by date DESC
// (tiebreak assignmentLetterId, then title) and returns the first limit items
// plus the number of distinct findings.
//
// On a duplicate the saved-report entry wins over a live one (between two of
// the same kind, the newer one wins, else the first seen); the kept entry
// gets the newer date of the two. Reports without an assignment letter are
// deduped per report so unrelated reports do not collapse into each other.
func MergeRecent(reports []models.AuditResultReport, live []LetterFindings, limit int) RecentResult {
	byKey := make(map[string]*candidate)
	var order []string

	add := func(dedupeScope string, c candidate) {
		title := strings.TrimSpace(c.item.Title)
		if title == "" {
			return
		}
		c.item.Title = title
		key := dedupeScope + "\x00" + strings.ToLower(title)
		prev, ok := byKey[key]
		if !ok {
			byKey[key] = &c
			order = append(order, key)
			return
		}
		newest := prev.date
		if c.date.After(newest) {
			newest = c.date
		}
		if (c.fromReport && !prev.fromReport) || (c.fromReport == prev.fromReport && c.date.After(prev.date)) {
			*prev = c
		}
		prev.date = newest
	}

	for _, r := range reports {
		letterID := strings.TrimSpace(r.AssignmentLetterID)
		scope := "letter:" + letterID
		if letterID == "" {
			scope = "report:" + r.ID.String()
		}
		date := r.CreatedAt
		if r.ReportDate != nil && !r.ReportDate.IsZero() {
			date = *r.ReportDate
		}
		reportID := r.ID.String()
		for _, f := range r.Findings {
			rid := reportID
			add(scope, candidate{
				item: RecentItem{
					Title:              f.Title,
					Category:           NormalizeCategory(f.Category),
					Action:             strings.TrimSpace(f.Action),
					Source:             SourceAuditResultReport,
					AssignmentLetterID: letterID,
					ReportID:           &rid,
				},
				date:       date,
				fromReport: true,
			})
		}
	}

	for _, lf := range live {
		letterID := strings.TrimSpace(lf.AssignmentLetterID)
		if letterID == "" {
			continue
		}
		for _, f := range lf.Findings {
			add("letter:"+letterID, candidate{
				item: RecentItem{
					Title:              f.Title,
					Category:           NormalizeCategory(f.Category),
					Action:             strings.TrimSpace(f.Action),
					Source:             f.Source,
					AssignmentLetterID: letterID,
				},
				date: f.Date,
			})
		}
	}

	all := make([]*candidate, 0, len(order))
	for _, k := range order {
		all = append(all, byKey[k])
	}
	sort.SliceStable(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if !a.date.Equal(b.date) {
			return a.date.After(b.date)
		}
		if a.item.AssignmentLetterID != b.item.AssignmentLetterID {
			return a.item.AssignmentLetterID < b.item.AssignmentLetterID
		}
		at, bt := strings.ToLower(a.item.Title), strings.ToLower(b.item.Title)
		if at != bt {
			return at < bt
		}
		return a.item.Title < b.item.Title
	})

	if limit < 1 || limit > MaxRecentLimit {
		limit = DefaultRecentLimit
	}
	n := len(all)
	if n > limit {
		n = limit
	}
	items := make([]RecentItem, 0, n)
	for _, c := range all[:n] {
		c.item.Date = c.date.UTC().Format(time.RFC3339)
		items = append(items, c.item)
	}
	return RecentResult{Items: items, Total: len(all)}
}

// LetterRow is the part of an assignment letter the letter filter needs
type LetterRow struct {
	ID           uuid.UUID
	LetterNumber string
	DeletedAt    gorm.DeletedAt
}

// LetterFilter decides which findings belong to a letter that is gone.
//
// Findings reference their assignment letter by a free-text key: the letter
// NUMBER in practice (the frontend saves letterNumber into
// audit_result_reports.assignment_letter_id, working_paper_*.working_paper_id
// and fieldwork_*.assignment_letter_id), occasionally the letter UUID. Keys
// are matched against both, trimmed and case-insensitively.
type LetterFilter struct {
	live    map[string]bool
	deleted map[string]bool
	// orphans are dropped only when at least one live letter exists, so an
	// empty or unreadable letters table never hides every finding
	dropOrphans bool
}

func letterKey(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// NewLetterFilter builds the filter from all letters, soft-deleted included
func NewLetterFilter(rows []LetterRow) LetterFilter {
	f := LetterFilter{live: map[string]bool{}, deleted: map[string]bool{}}
	for _, r := range rows {
		target := f.live
		if r.DeletedAt.Valid {
			target = f.deleted
		}
		if r.ID != uuid.Nil {
			target[letterKey(r.ID.String())] = true
		}
		if k := letterKey(r.LetterNumber); k != "" {
			target[k] = true
		}
	}
	f.dropOrphans = len(f.live) > 0
	return f
}

// Drop reports whether findings keyed by letterKey must be hidden: the key
// matches a soft-deleted letter (by id or number), or it matches no existing
// letter at all (orphan: hard-deleted or never-created letter). An empty key
// (report saved without a letter) is never dropped.
func (f LetterFilter) Drop(key string) bool {
	k := letterKey(key)
	if k == "" {
		return false
	}
	if f.deleted[k] {
		return true
	}
	return f.dropOrphans && !f.live[k]
}

// loadLetterFilter reads every assignment letter. If the table cannot be read
// the filter is disabled (logged) rather than failing or hiding everything.
func loadLetterFilter(db *gorm.DB) LetterFilter {
	var rows []LetterRow
	err := db.Unscoped().Model(&models.AssignmentLetter{}).
		Select("id", "letter_number", "deleted_at").Find(&rows).Error
	if err != nil {
		if logger.Log != nil {
			logger.Warn("recent findings: cannot read assignment letters, letter filter skipped", zap.Error(err))
		}
		return LetterFilter{}
	}
	f := NewLetterFilter(rows)
	if !f.dropOrphans && logger.Log != nil {
		logger.Warn("recent findings: no live assignment letters, orphan filter skipped")
	}
	return f
}

// LoadRecent reads saved reports, KKA and fieldwork rows (GORM excludes
// soft-deleted rows) and merges them, dropping findings of deleted or
// non-existent assignment letters (see LetterFilter).
func LoadRecent(db *gorm.DB, limit int) (RecentResult, error) {
	dropLetter := loadLetterFilter(db).Drop

	var reports []models.AuditResultReport
	if err := db.Select("id", "assignment_letter_id", "findings", "report_date", "created_at").
		Order("created_at ASC").Order("id ASC").Find(&reports).Error; err != nil {
		return RecentResult{}, err
	}
	keptReports := reports[:0]
	for _, r := range reports {
		if !dropLetter(r.AssignmentLetterID) {
			keptReports = append(keptReports, r)
		}
	}

	var causes []models.WorkingPaperCause
	var plans []models.WorkingPaperPlan
	var risks []models.WorkingPaperRisk
	var tcs []models.FieldworkTestControl
	for _, q := range []struct {
		dest interface{}
		col  string
	}{
		{&causes, "working_paper_id"},
		{&plans, "working_paper_id"},
		{&risks, "working_paper_id"},
		{&tcs, "assignment_letter_id"},
	} {
		if err := db.Order(q.col + " ASC").Order("created_at ASC").Order("id ASC").Find(q.dest).Error; err != nil {
			return RecentResult{}, err
		}
	}

	byLetter := make(map[string]*LiveSources)
	var letters []string
	get := func(id string) *LiveSources {
		id = strings.TrimSpace(id)
		if id == "" || dropLetter(id) {
			return nil
		}
		s, ok := byLetter[id]
		if !ok {
			s = &LiveSources{}
			byLetter[id] = s
			letters = append(letters, id)
		}
		return s
	}
	for _, r := range causes {
		if s := get(r.WorkingPaperID); s != nil {
			s.Causes = append(s.Causes, r)
		}
	}
	for _, r := range plans {
		if s := get(r.WorkingPaperID); s != nil {
			s.Plans = append(s.Plans, r)
		}
	}
	for _, r := range risks {
		if s := get(r.WorkingPaperID); s != nil {
			s.Risks = append(s.Risks, r)
		}
	}
	for _, r := range tcs {
		if s := get(r.AssignmentLetterID); s != nil {
			s.TestControls = append(s.TestControls, r)
		}
	}

	live := make([]LetterFindings, 0, len(letters))
	for _, id := range letters {
		live = append(live, LetterFindings{AssignmentLetterID: id, Findings: Derive(*byLetter[id])})
	}
	return MergeRecent(keptReports, live, limit), nil
}

// Recent handles GET /audit-result-reports/recent-findings?limit=5
func Recent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := LoadRecent(db.WithContext(c.Request.Context()), ParseRecentLimit(c.Query("limit")))
		if err != nil {
			if logger.Log != nil {
				logger.Error("failed to load recent findings", zap.Error(err))
			}
			response.InternalServerError(c, "Failed to fetch recent findings")
			return
		}
		response.OK(c, "Recent findings retrieved successfully", result)
	}
}
