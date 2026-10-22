package controllers

import (
	"context"
	"log"
	"math"
	"time"

	"risk-service/models"
	"risk-service/pkg/kafka"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// loadFactorWeights returns the current corporate weight (fraction) of every
// corporate risk factor, by corporate factor id.
func loadFactorWeights(tx *gorm.DB) (map[uuid.UUID]float64, error) {
	var factors []models.CorporateRiskFactor
	if err := tx.Find(&factors).Error; err != nil {
		return nil, err
	}
	weights := make(map[uuid.UUID]float64, len(factors))
	for _, f := range factors {
		weights[f.ID] = f.Weight
	}
	return weights, nil
}

// rescoreYearRow recalculates weighted_score of each score row from the current
// weights (writing the ones that differ) and sets row's risk_index, risk_level
// and audit_priority in memory. It does not save row.
//
// Scores whose factor is not in weights are ignored. If no factor with a
// positive weight has a score, row is reset to "not scored" (scored_at nil,
// index 0, level "", priority false). Otherwise row.ScoredAt is left as the
// caller set it.
//
// It reports whether any field that is published in year.scored, or the scored
// state, changed.
func rescoreYearRow(tx *gorm.DB, row *models.AuditUniverseYear, scores []models.AuditUniverseRiskScore, weights map[uuid.UUID]float64, now time.Time) (bool, error) {
	factors := make([]models.ScoredFactor, 0, len(scores))
	for i := range scores {
		sc := &scores[i]
		w, ok := weights[sc.CorporateRiskFactorID]
		if !ok {
			continue
		}
		ws := w * float64(sc.Score)
		if math.Abs(ws-sc.WeightedScore) > 1e-6 {
			if err := tx.Model(&models.AuditUniverseRiskScore{}).Where("id = ?", sc.ID).
				Updates(map[string]any{"weighted_score": ws, "updated_at": now}).Error; err != nil {
				return false, err
			}
			sc.WeightedScore = ws
		}
		factors = append(factors, models.ScoredFactor{Weight: w, Score: sc.Score})
	}

	before := *row
	if index, ok := models.ComputeRiskIndex(factors); ok {
		row.RiskIndex, row.RiskLevel, row.AuditPriority = models.ClassifyRiskIndex(index)
	} else {
		row.ScoredAt = nil
		row.RiskIndex, row.RiskLevel, row.AuditPriority = 0, "", false
	}

	changed := math.Abs(before.RiskIndex-row.RiskIndex) > 1e-9 ||
		before.RiskLevel != row.RiskLevel ||
		before.AuditPriority != row.AuditPriority ||
		(before.ScoredAt == nil) != (row.ScoredAt == nil)
	return changed, nil
}

// saveYearRowScore writes the derived columns of row. Select makes GORM write
// zero values and a nil scored_at too.
func saveYearRowScore(tx *gorm.DB, row *models.AuditUniverseYear) error {
	return tx.Model(&models.AuditUniverseYear{}).Where("id = ?", row.ID).
		Select("risk_index", "risk_level", "audit_priority", "scored_at", "updated_at").
		Updates(row).Error
}

// yearRowsScoredOn returns the ids of the year rows that have a score for any
// of corporateFactorIDs. Call it BEFORE deleting those scores and pass the
// result to recomputeScoredYears, so that a row which loses its last score in
// that transaction is reset to "not scored" instead of being skipped.
func yearRowsScoredOn(tx *gorm.DB, corporateFactorIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	if len(corporateFactorIDs) == 0 {
		return out, nil
	}
	var ids []uuid.UUID
	if err := tx.Model(&models.AuditUniverseRiskScore{}).
		Where("corporate_risk_factor_id IN ?", corporateFactorIDs).
		Distinct().Pluck("audit_universe_year_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// recomputeScoredYears re-derives weighted scores and the risk index, level
// and priority of every SCORED year row (scored_at not null) from the current
// corporate weights. Run it in the same transaction as any change to the
// corporate weights or the factor set. Unscored rows have no scores and are
// left alone. It returns the rows whose published values changed; the caller
// publishes year.scored for each after commit.
//
// A scored row with no score rows at all is left as it is, unless its scores
// were deleted in this transaction (lostScores, see yearRowsScoredOn). Such
// rows predate per-factor scoring: migrate backfilled scored_at for every row
// that existed before the column, and most of them carry an index and level
// with no factor scores behind them. There is nothing to re-derive them from,
// so a weight change must not wipe them to "not scored".
func recomputeScoredYears(tx *gorm.DB, now time.Time, lostScores map[uuid.UUID]bool) ([]models.AuditUniverseYear, error) {
	weights, err := loadFactorWeights(tx)
	if err != nil {
		return nil, err
	}

	var rows []models.AuditUniverseYear
	if err := tx.Where("scored_at IS NOT NULL").Order("year ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	var scores []models.AuditUniverseRiskScore
	if err := tx.Where("audit_universe_year_id IN (?)",
		tx.Model(&models.AuditUniverseYear{}).Select("id").Where("scored_at IS NOT NULL")).
		Find(&scores).Error; err != nil {
		return nil, err
	}
	byRow := make(map[uuid.UUID][]models.AuditUniverseRiskScore, len(rows))
	for _, sc := range scores {
		byRow[sc.AuditUniverseYearID] = append(byRow[sc.AuditUniverseYearID], sc)
	}

	var changed []models.AuditUniverseYear
	for i := range rows {
		row := &rows[i]
		if len(byRow[row.ID]) == 0 && !lostScores[row.ID] {
			continue
		}
		diff, err := rescoreYearRow(tx, row, byRow[row.ID], weights, now)
		if err != nil {
			return nil, err
		}
		if !diff {
			continue
		}
		row.UpdatedAt = now
		if err := saveYearRowScore(tx, row); err != nil {
			return nil, err
		}
		changed = append(changed, *row)
	}
	return changed, nil
}

// publishYearScores records one year.scored event per row. Like the other
// Audit Universe events it is written to the outbox after the data commit and
// is best-effort: a failure is logged, and POST /audit-universe/events/resync
// repairs any gap.
func publishYearScores(events *kafka.Publisher, rows []models.AuditUniverseYear) {
	for _, row := range rows {
		if err := events.PublishYearScore(context.Background(), yearScoreData(row)); err != nil {
			log.Printf("audit universe event %s for %s not published: %v", kafka.EventYearScored, row.CorporateAuditUniverseID, err)
		}
	}
}
