// Package fieldwork serves read-only aggregates over the fieldwork tables.
package fieldwork

import (
	"strings"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MaxRecommendationSummaryIDs caps audit_result_report_ids per request
const MaxRecommendationSummaryIDs = 200

// RecommendationCount is the recommendation count of one LHA
type RecommendationCount struct {
	AuditResultReportID string `json:"auditResultReportId"`
	// AssignmentLetterID is the LHA's stored audit_result_reports.assignment_letter_id
	// (letter number or letter UUID); "" for an unknown or deleted LHA.
	AssignmentLetterID string `json:"assignmentLetterId"`
	Count              int64  `json:"count"`
}

// RecommendationSummary is the data payload of
// GET /fieldwork/test-controls/recommendations/summary
type RecommendationSummary struct {
	Total int64                 `json:"total"`
	Items []RecommendationCount `json:"items"`
}

// ParseReportIDs splits the comma separated audit_result_report_ids value into
// unique, trimmed ids in request order.
func ParseReportIDs(raw string) []string {
	seen := map[string]bool{}
	var ids []string
	for _, part := range strings.Split(raw, ",") {
		id := strings.TrimSpace(part)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

// LoadRecommendationSummary counts, per LHA id, the live fieldwork test
// controls with a non-blank recommendation of the LHA's assignment letter.
//
// Test controls and LHAs reference their letter by the same string key
// (letter number, or letter UUID for some rows), so a test control belongs to
// an LHA when its assignment_letter_id is the LHA's stored key, or the letter
// number / id of the letter that key resolves to. Three queries regardless of
// the number of ids. Unknown, deleted or malformed ids yield a 0 item.
func LoadRecommendationSummary(db *gorm.DB, ids []string) (RecommendationSummary, error) {
	out := RecommendationSummary{Items: make([]RecommendationCount, 0, len(ids))}

	var reportIDs []uuid.UUID
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			reportIDs = append(reportIDs, u)
		}
	}

	// 1. LHA -> stored letter key (soft-deleted LHAs excluded)
	letterRefOf := map[string]string{}
	if len(reportIDs) > 0 {
		var reports []models.AuditResultReport
		if err := db.Select("id", "assignment_letter_id").
			Where("id IN ?", reportIDs).
			Find(&reports).Error; err != nil {
			return out, err
		}
		for _, r := range reports {
			letterRefOf[r.ID.String()] = strings.TrimSpace(r.AssignmentLetterID)
		}
	}

	// 2. Every key a test control may hold for each stored letter key
	refSet := map[string]bool{}
	var refs []string
	var refUUIDs []uuid.UUID
	for _, ref := range letterRefOf {
		if ref == "" || refSet[ref] {
			continue
		}
		refSet[ref] = true
		refs = append(refs, ref)
		if u, err := uuid.Parse(ref); err == nil {
			refUUIDs = append(refUUIDs, u)
		}
	}
	keysOf := map[string][]string{}
	for _, ref := range refs {
		keysOf[ref] = []string{ref}
	}
	if len(refs) > 0 {
		// One grouped condition so the soft-delete filter applies to both branches
		q := db.Select("id", "letter_number")
		if len(refUUIDs) > 0 {
			q = q.Where("letter_number IN ? OR id IN ?", refs, refUUIDs)
		} else {
			q = q.Where("letter_number IN ?", refs)
		}
		var letters []models.AssignmentLetter
		if err := q.Find(&letters).Error; err != nil {
			return out, err
		}
		for _, l := range letters {
			number := strings.TrimSpace(l.LetterNumber)
			id := l.ID.String()
			for _, ref := range refs {
				if ref == number || strings.EqualFold(ref, id) {
					keysOf[ref] = appendUnique(keysOf[ref], number, id)
				}
			}
		}
	}

	// 3. Recommendation counts per test control letter key
	countOf := map[string]int64{}
	var allKeys []string
	seenKey := map[string]bool{}
	for _, keys := range keysOf {
		for _, k := range keys {
			if !seenKey[k] {
				seenKey[k] = true
				allKeys = append(allKeys, k)
			}
		}
	}
	if len(allKeys) > 0 {
		var rows []struct {
			AssignmentLetterID string
			N                  int64
		}
		if err := db.Model(&models.FieldworkTestControl{}).
			Select("assignment_letter_id, COUNT(*) AS n").
			Where("assignment_letter_id IN ?", allKeys).
			Where("TRIM(COALESCE(recommendation, '')) <> ''").
			Group("assignment_letter_id").
			Scan(&rows).Error; err != nil {
			return out, err
		}
		for _, r := range rows {
			countOf[r.AssignmentLetterID] += r.N
		}
	}

	// Total counts each letter key once: LHAs that share an assignment letter
	// (or reference it by number and by UUID) must not repeat its test controls.
	// Items keep the per-LHA count, which is the letter's count for every LHA of it.
	totalCounted := map[string]bool{}
	for _, id := range ids {
		item := RecommendationCount{AuditResultReportID: id}
		if u, err := uuid.Parse(id); err == nil {
			if ref, ok := letterRefOf[u.String()]; ok {
				item.AssignmentLetterID = ref
				// Keys are distinct and each test control has one key: no double count
				for _, k := range keysOf[ref] {
					item.Count += countOf[k]
					if !totalCounted[k] {
						totalCounted[k] = true
						out.Total += countOf[k]
					}
				}
			}
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func appendUnique(list []string, values ...string) []string {
	for _, v := range values {
		if v == "" {
			continue
		}
		dup := false
		for _, x := range list {
			if x == v {
				dup = true
				break
			}
		}
		if !dup {
			list = append(list, v)
		}
	}
	return list
}

// RecommendationSummaryHandler handles
// GET /api/v1/fieldwork/test-controls/recommendations/summary?audit_result_report_ids=<id1>,<id2>
func RecommendationSummaryHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ids := ParseReportIDs(c.Query("audit_result_report_ids"))
		if len(ids) == 0 {
			response.BadRequest(c, "audit_result_report_ids is required")
			return
		}
		if len(ids) > MaxRecommendationSummaryIDs {
			response.BadRequest(c, "audit_result_report_ids accepts at most 200 ids")
			return
		}

		summary, err := LoadRecommendationSummary(db.WithContext(c.Request.Context()), ids)
		if err != nil {
			if logger.Log != nil {
				logger.Error("failed to load recommendation summary", zap.Error(err))
			}
			response.InternalServerError(c, "Failed to fetch recommendation summary")
			return
		}
		response.OK(c, "Recommendation summary retrieved successfully", summary)
	}
}
