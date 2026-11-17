package ic_assessment

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Tenant threshold settings (decision Q4): the proposal formulas read them
// and every proposal records the values it used in its evidence.

// Environment variables read once, by "audit migrate up", to fill the
// settings row when it does not exist yet. They never overwrite a row.
const (
	EnvLHADueDays          = "IC_LHA_DUE_DAYS"
	EnvMgmtResponseDueDays = "IC_MGMT_RESPONSE_DUE_DAYS"
)

// Defaults when neither the row nor the environment says otherwise
const (
	DefaultLHADueDays          = 30
	DefaultMgmtResponseDueDays = 14
	MinThresholdDays           = 1
	MaxThresholdDays           = 365
)

// Settings are the thresholds in effect
type Settings struct {
	LHADueDays                int        `json:"lha_due_days"`
	ManagementResponseDueDays int        `json:"management_response_due_days"`
	Source                    string     `json:"source"` // "stored" or "default"
	UpdatedByName             string     `json:"updated_by_name,omitempty"`
	UpdatedAt                 *time.Time `json:"updated_at,omitempty"`
}

// DefaultSettings are the built-in defaults
func DefaultSettings() Settings {
	return Settings{LHADueDays: DefaultLHADueDays, ManagementResponseDueDays: DefaultMgmtResponseDueDays, Source: "default"}
}

// ValidThreshold reports whether days is within 1..365
func ValidThreshold(days int) bool { return days >= MinThresholdDays && days <= MaxThresholdDays }

// LoadSettings reads the settings row; without one (migrate not run yet) the
// defaults apply
func LoadSettings(db *gorm.DB) (Settings, error) {
	var row models.ICAssessmentSettings
	err := db.Where("id = ?", models.ICAssessmentSettingsID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	s := Settings{LHADueDays: row.LHADueDays, ManagementResponseDueDays: row.ManagementResponseDueDays, Source: "stored", UpdatedByName: row.UpdatedByName}
	if !row.UpdatedAt.IsZero() {
		t := row.UpdatedAt
		s.UpdatedAt = &t
	}
	// A row edited by hand outside 1..365 falls back per field
	if !ValidThreshold(s.LHADueDays) {
		s.LHADueDays = DefaultLHADueDays
	}
	if !ValidThreshold(s.ManagementResponseDueDays) {
		s.ManagementResponseDueDays = DefaultMgmtResponseDueDays
	}
	return s, nil
}

// envThreshold reads one env value: (value, true) when set and valid;
// a set but invalid value is reported in warn
func envThreshold(getenv func(string) string, key string, def int) (int, string) {
	raw := strings.TrimSpace(getenv(key))
	if raw == "" {
		return def, ""
	}
	n, err := strconv.Atoi(raw)
	if err != nil || !ValidThreshold(n) {
		return def, fmt.Sprintf("%s=%q is not a whole number of days in %d..%d; using %d", key, raw, MinThresholdDays, MaxThresholdDays, def)
	}
	return n, ""
}

// EnsureSettings inserts the settings row if it is missing, with the values
// of IC_LHA_DUE_DAYS / IC_MGMT_RESPONSE_DUE_DAYS when set and valid, else the
// defaults. An existing row is never changed. getenv nil uses os.Getenv.
// It returns whether a row was inserted and any warnings about the env.
func EnsureSettings(db *gorm.DB, getenv func(string) string, now time.Time) (bool, []string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	var warnings []string
	lha, w1 := envThreshold(getenv, EnvLHADueDays, DefaultLHADueDays)
	mgmt, w2 := envThreshold(getenv, EnvMgmtResponseDueDays, DefaultMgmtResponseDueDays)
	for _, w := range []string{w1, w2} {
		if w != "" {
			warnings = append(warnings, w)
		}
	}
	row := models.ICAssessmentSettings{
		ID:                        models.ICAssessmentSettingsID,
		LHADueDays:                lha,
		ManagementResponseDueDays: mgmt,
		UpdatedByName:             "system (migrate)",
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	res := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).Create(&row)
	if res.Error != nil {
		return false, warnings, res.Error
	}
	return res.RowsAffected > 0, warnings, nil
}

// SettingsUpdate is the body of PUT /ic-assessments/settings; omitted fields
// keep their value
type SettingsUpdate struct {
	LHADueDays                *int `json:"lha_due_days"`
	ManagementResponseDueDays *int `json:"management_response_due_days"`
}

// UpdateSettings validates and stores the thresholds (upsert) and logs a
// SETTINGS_UPDATE event in the same transaction
func UpdateSettings(db *gorm.DB, actor Actor, req SettingsUpdate, now time.Time) (Settings, error) {
	if req.LHADueDays == nil && req.ManagementResponseDueDays == nil {
		return Settings{}, invalidf("lha_due_days or management_response_due_days is required")
	}
	if req.LHADueDays != nil && !ValidThreshold(*req.LHADueDays) {
		return Settings{}, invalidf("lha_due_days must be between %d and %d", MinThresholdDays, MaxThresholdDays)
	}
	if req.ManagementResponseDueDays != nil && !ValidThreshold(*req.ManagementResponseDueDays) {
		return Settings{}, invalidf("management_response_due_days must be between %d and %d", MinThresholdDays, MaxThresholdDays)
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		before, err := LoadSettings(tx)
		if err != nil {
			return err
		}
		after := before
		if req.LHADueDays != nil {
			after.LHADueDays = *req.LHADueDays
		}
		if req.ManagementResponseDueDays != nil {
			after.ManagementResponseDueDays = *req.ManagementResponseDueDays
		}
		row := models.ICAssessmentSettings{
			ID:                        models.ICAssessmentSettingsID,
			LHADueDays:                after.LHADueDays,
			ManagementResponseDueDays: after.ManagementResponseDueDays,
			UpdatedByUserID:           actor.UserID,
			UpdatedByName:             actor.Name,
			CreatedAt:                 now,
			UpdatedAt:                 now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"lha_due_days", "management_response_due_days", "updated_by_user_id", "updated_by_name", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, nil, ActionSettingsUpdate, "", "", "",
			map[string]interface{}{"lha_due_days": before.LHADueDays, "management_response_due_days": before.ManagementResponseDueDays, "source": before.Source},
			map[string]interface{}{"lha_due_days": after.LHADueDays, "management_response_due_days": after.ManagementResponseDueDays}, now)
	})
	if err != nil {
		return Settings{}, err
	}
	return LoadSettings(db)
}
