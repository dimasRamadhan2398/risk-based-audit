package strategicplan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"audit-service/models"
	"audit-service/pkg/activitycode"
	apperrors "audit-service/pkg/errors"

	"gorm.io/gorm"
)

// CodePrefix is the prefix of every strategic plan code (SO-IA01, SO-IA02, ...).
// It is also the activity_code_sequences key the numbers are drawn from.
const CodePrefix = "SO-IA"

var (
	codeRe = regexp.MustCompile(`^` + CodePrefix + `(\d+)$`)
	// A KPI target or realization: a plain decimal number such as 85, 85.5 or -2.
	numberRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
)

// PrepareOnCreate is the crud.CreateHook for /strategic-plans. It checks the
// period and the KPI numbers, then assigns the code; a code sent by the client
// is ignored.
func PrepareOnCreate(tx *gorm.DB, entity interface{}) error {
	plan, ok := entity.(*models.StrategicPlan)
	if !ok {
		return fmt.Errorf("strategicplan: unexpected entity %T", entity)
	}
	if err := validatePeriod(plan.YearStart, plan.YearEnd); err != nil {
		return err
	}
	if err := validateNumbers(plan.KPITargets, plan.KPIActuals, plan.Target, plan.Actual); err != nil {
		return err
	}
	code, err := NextCode(tx)
	if err != nil {
		return err
	}
	plan.Code = code
	return nil
}

// PrepareOnUpdate is the crud.UpdateHook for /strategic-plans. The code never
// changes after create; the period (merged with the stored years) and the KPI
// numbers that are being written are checked as on create.
func PrepareOnUpdate(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	plan, ok := existing.(*models.StrategicPlan)
	if !ok {
		return fmt.Errorf("strategicplan: unexpected entity %T", existing)
	}
	delete(columns, "code")

	start, end := plan.YearStart, plan.YearEnd
	for col, year := range map[string]*int{"year_start": &start, "year_end": &end} {
		raw, ok := columns[col]
		if !ok || raw == nil {
			continue
		}
		y, err := toInt(raw)
		if err != nil {
			return validationError(col + " must be a year")
		}
		*year = y
	}
	if err := validatePeriod(start, end); err != nil {
		return err
	}

	var targets, actuals map[string]string
	for col, dst := range map[string]*map[string]string{"kpi_targets": &targets, "kpi_actuals": &actuals} {
		raw, ok := columns[col].([]byte)
		if !ok {
			continue
		}
		if err := json.Unmarshal(raw, dst); err != nil {
			return validationError(col + " must map each period to a number written as text")
		}
	}
	return validateNumbers(targets, actuals, columnString(columns, "target"), columnString(columns, "actual"))
}

// NextCode returns the next strategic plan code. The number comes from the
// activity_code_sequences counter (row-locked inside the caller's transaction,
// so concurrent creates get different numbers) and is never below the highest
// SO-IA number stored, soft-deleted and seeded rows included.
func NextCode(tx *gorm.DB) (string, error) {
	var codes []string
	if err := tx.Unscoped().Model(&models.StrategicPlan{}).
		Where("code LIKE ?", CodePrefix+"%").
		Pluck("code", &codes).Error; err != nil {
		return "", fmt.Errorf("strategicplan: read codes: %w", err)
	}
	floor := 0
	for _, c := range codes {
		m := codeRe.FindStringSubmatch(strings.TrimSpace(c))
		if m == nil {
			continue
		}
		if n, err := strconv.Atoi(m[1]); err == nil && n > floor {
			floor = n
		}
	}
	seq, err := activitycode.ReserveAbove(tx, CodePrefix, 1, floor)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%02d", CodePrefix, seq), nil
}

// validatePeriod rejects an end year before the start year. A year of 0 is
// unset and not compared.
func validatePeriod(start, end int) error {
	if start > 0 && end > 0 && end < start {
		return validationError(fmt.Sprintf("yearEnd (%d) must not be before yearStart (%d)", end, start))
	}
	return nil
}

// validateNumbers rejects a target or realization that is not a number. Empty
// values are allowed: a period may not have a target or realization yet.
func validateNumbers(targets, actuals map[string]string, target, actual string) error {
	check := func(label, value string) error {
		if v := strings.TrimSpace(value); v != "" && !numberRe.MatchString(v) {
			return validationError(fmt.Sprintf("%s must be a number (e.g. 85 or 85.5), got %q", label, value))
		}
		return nil
	}
	for _, m := range []struct {
		label  string
		values map[string]string
	}{{"Target", targets}, {"Realization", actuals}} {
		periods := make([]string, 0, len(m.values))
		for p := range m.values {
			periods = append(periods, p)
		}
		sort.Strings(periods)
		for _, p := range periods {
			if err := check(m.label+" "+p, m.values[p]); err != nil {
				return err
			}
		}
	}
	if err := check("Target", target); err != nil {
		return err
	}
	return check("Realization", actual)
}

func validationError(msg string) error {
	return apperrors.New("VALIDATION_ERROR", msg, http.StatusBadRequest)
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case float64:
		if n != float64(int(n)) {
			return 0, fmt.Errorf("not a whole number")
		}
		return int(n), nil
	case int:
		return n, nil
	case string:
		return strconv.Atoi(strings.TrimSpace(n))
	}
	return 0, fmt.Errorf("unexpected type %T", v)
}

// columnString is the column value as text ("" when absent or null). Numbers
// are formatted so they pass the number check.
func columnString(columns map[string]interface{}, col string) string {
	switch v := columns[col].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}
