package ic_assessment

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/masterclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Event actions of the QUANTITATIVE_POF method
const (
	ActionBranchMethodUpdate = "BRANCH_METHOD_UPDATE"
	ActionPofSettingsUpdate  = "POF_SETTINGS_UPDATE"
	ActionPofUpdate          = "POF_UPDATE"
	ActionPofAssign          = "POF_ASSIGN"
	ActionPofEvidence        = "POF_EVIDENCE"
	ActionPofImport          = "POF_IMPORT"
)

// ---------------------------------------------------------------------------
// FieldErrors: 422 with one message per request field
// ---------------------------------------------------------------------------

// FieldErrors is a validation error with one message per field. The HTTP layer
// answers 422 {"error": {"code": "VALIDATION_ERROR", "fields": {...}}}.
type FieldErrors map[string]string

func (f FieldErrors) Error() string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("invalid request: ")
	for i, k := range keys {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(k + " " + f[k])
	}
	return b.String()
}

func methodMismatch(method, what string) *apperrors.AppError {
	return apperrors.New("METHOD_MISMATCH", "This assessment uses the "+method+" method: "+what, http.StatusConflict)
}

// methodOf returns the method of an assessment ("" when it does not exist)
func (s *Service) methodOf(id uuid.UUID) string {
	var m string
	if err := s.DB.Model(&models.ICAssessment{}).Where("id = ?", id).Select("method").Scan(&m).Error; err != nil {
		return ""
	}
	if m == "" {
		return MethodQualitative
	}
	return m
}

// ---------------------------------------------------------------------------
// Method per branch
// ---------------------------------------------------------------------------

// ResolveMethod returns the method for new assessments of a branch
// (locationKey "" = no branch) and where it comes from: "branch", "default"
// (the no-branch row) or "builtin" (QUALITATIVE).
func ResolveMethod(db *gorm.DB, locationKey string) (method, source string, err error) {
	var rows []models.ICBranchMethod
	keys := []string{""}
	if locationKey != "" {
		keys = append(keys, locationKey)
	}
	if err := db.Where("location_key IN ?", keys).Find(&rows).Error; err != nil {
		return "", "", err
	}
	byKey := map[string]string{}
	for _, r := range rows {
		byKey[r.LocationKey] = r.Method
	}
	if locationKey != "" {
		if m, ok := byKey[locationKey]; ok && IsMethod(m) {
			return m, "branch", nil
		}
	}
	if m, ok := byKey[""]; ok && IsMethod(m) {
		if locationKey == "" {
			return m, "branch", nil
		}
		return m, "default", nil
	}
	return MethodQualitative, "builtin", nil
}

// BranchMethodItem is one stored setting
type BranchMethodItem struct {
	LocationID    *uuid.UUID `json:"location_id"`
	LocationName  string     `json:"location_name"`
	Method        string     `json:"method"`
	UpdatedByName string     `json:"updated_by_name"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// EffectiveMethod is the method a new assessment of a branch would get
type EffectiveMethod struct {
	LocationID *string `json:"location_id"`
	Method     string  `json:"method"`
	Source     string  `json:"source"`
}

// BranchMethods is GET /ic-assessments/branch-methods
type BranchMethods struct {
	Methods   []string           `json:"methods"`
	Default   EffectiveMethod    `json:"default"`
	Items     []BranchMethodItem `json:"items"`
	Effective *EffectiveMethod   `json:"effective,omitempty"`
}

// ListBranchMethods returns the settings; with locationKey != nil also the
// effective method of that scope ("" = no branch)
func (s *Service) ListBranchMethods(actor Actor, locationKey *string) (*BranchMethods, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	var rows []models.ICBranchMethod
	if err := s.DB.Order("location_name, location_key").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := &BranchMethods{Methods: []string{MethodQualitative, MethodQuantitativePoF}, Items: []BranchMethodItem{},
		Default: EffectiveMethod{Method: MethodQualitative, Source: "builtin"}}
	for _, r := range rows {
		if r.LocationKey == "" {
			out.Default = EffectiveMethod{Method: r.Method, Source: "stored"}
			continue
		}
		out.Items = append(out.Items, BranchMethodItem{LocationID: r.LocationID, LocationName: r.LocationName, Method: r.Method,
			UpdatedByName: r.UpdatedByName, UpdatedAt: r.UpdatedAt})
	}
	if locationKey != nil {
		m, src, err := ResolveMethod(s.DB, *locationKey)
		if err != nil {
			return nil, err
		}
		e := &EffectiveMethod{Method: m, Source: src}
		if *locationKey != "" {
			k := *locationKey
			e.LocationID = &k
		}
		out.Effective = e
	}
	return out, nil
}

// BranchMethodRequest is the body of PUT /ic-assessments/branch-methods
type BranchMethodRequest struct {
	LocationID *string `json:"location_id"`
	Method     string  `json:"method"`
}

// BranchMethodResult is the stored row with the count of existing assessments
// per method in that scope, so the UI can warn about mixing methods
type BranchMethodResult struct {
	BranchMethodItem
	ExistingAssessments map[string]int64 `json:"existing_assessments"`
}

// PutBranchMethod sets the method for new assessments of a branch (approvers)
func (s *Service) PutBranchMethod(ctx context.Context, src Sources, actor Actor, req BranchMethodRequest) (*BranchMethodResult, error) {
	if !actor.IsApprover() {
		return nil, forbiddenf("Only CHIEF_AUDIT_EXECUTIVE or ADMIN can change the assessment method of a branch")
	}
	fields := FieldErrors{}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if !IsMethod(method) {
		fields["method"] = "must be QUALITATIVE or QUANTITATIVE_POF"
	}
	var locID *uuid.UUID
	locKey, locName := "", ""
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		l, err := uuid.Parse(strings.TrimSpace(*req.LocationID))
		if err != nil || l == uuid.Nil {
			fields["location_id"] = "must be a UUID or null"
		} else {
			locID, locKey = &l, l.String()
		}
	}
	if len(fields) > 0 {
		return nil, fields
	}
	if locID != nil {
		l, err := src.Location(ctx, locKey)
		if err != nil {
			if errors.Is(err, masterclient.ErrNotFound) || strings.Contains(err.Error(), "answered 404") {
				return nil, FieldErrors{"location_id": "is not a location of this organisation"}
			}
			return nil, unavailablef("Could not check the location in master-service")
		}
		locName = strings.TrimSpace(l.Name)
	}
	now := s.now()
	var before string
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var cur models.ICBranchMethod
		if err := tx.Where("location_key = ?", locKey).Take(&cur).Error; err == nil {
			before = cur.Method
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		row := models.ICBranchMethod{LocationKey: locKey, LocationID: locID, LocationName: locName, Method: method,
			UpdatedByUserID: actor.UserID, UpdatedByName: actor.Name, CreatedAt: now, UpdatedAt: now}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "location_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"method", "location_name", "updated_by_user_id", "updated_by_name", "updated_at"}),
		}).Create(&row).Error; err != nil {
			return err
		}
		return writeEvent(tx, actor, nil, ActionBranchMethodUpdate, "", "", "",
			map[string]interface{}{"location_id": locKey, "method": before},
			map[string]interface{}{"location_id": locKey, "method": method}, now)
	})
	if err != nil {
		return nil, err
	}
	res := &BranchMethodResult{BranchMethodItem: BranchMethodItem{LocationID: locID, LocationName: locName, Method: method,
		UpdatedByName: actor.Name, UpdatedAt: now}, ExistingAssessments: map[string]int64{MethodQualitative: 0, MethodQuantitativePoF: 0}}
	type cnt struct {
		Method string
		N      int64
	}
	var counts []cnt
	if err := s.DB.Model(&models.ICAssessment{}).Select("method, count(*) as n").Where("location_key = ?", locKey).Group("method").Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, c := range counts {
		m := c.Method
		if m == "" {
			m = MethodQualitative
		}
		res.ExistingAssessments[m] += c.N
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// PoF settings: thresholds and owner overrides
// ---------------------------------------------------------------------------

// OwnerOverride is a tenant override of the default owner labels of a principle
type OwnerOverride struct {
	PrincipleCode string     `json:"principle_code"`
	OwnerRoles    []string   `json:"owner_roles"`
	UpdatedByName string     `json:"updated_by_name"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

// PofSettings is GET /ic-assessments/pof-settings
type PofSettings struct {
	PofYesThreshold     float64         `json:"pof_yes_threshold"`
	PofPartialThreshold float64         `json:"pof_partial_threshold"`
	Source              string          `json:"source"`
	OwnerOverrides      []OwnerOverride `json:"owner_overrides"`
}

// LoadPofThresholds reads the thresholds in effect (defaults without a row)
func LoadPofThresholds(db *gorm.DB) (Thresholds, string, error) {
	var row models.ICAssessmentSettings
	err := db.Where("id = ?", models.ICAssessmentSettingsID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultThresholds(), "default", nil
	}
	if err != nil {
		return Thresholds{}, "", err
	}
	th := Thresholds{Yes: row.PofYesThreshold, Partial: row.PofPartialThreshold}
	if th.Validate() != nil {
		return DefaultThresholds(), "default", nil
	}
	return th, "stored", nil
}

func ownerOverrideMap(db *gorm.DB) (map[string]string, error) {
	var rows []models.ICPofOwnerOverride
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, r := range rows {
		out[r.PrincipleCode] = r.OwnerRoles
	}
	return out, nil
}

// GetPofSettings returns the thresholds and the owner overrides
func (s *Service) GetPofSettings(actor Actor) (*PofSettings, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	th, src, err := LoadPofThresholds(s.DB)
	if err != nil {
		return nil, err
	}
	out := &PofSettings{PofYesThreshold: th.Yes, PofPartialThreshold: th.Partial, Source: src, OwnerOverrides: []OwnerOverride{}}
	var rows []models.ICPofOwnerOverride
	if err := s.DB.Order("principle_code").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		t := r.UpdatedAt
		out.OwnerOverrides = append(out.OwnerOverrides, OwnerOverride{PrincipleCode: r.PrincipleCode,
			OwnerRoles: SplitMethods(r.OwnerRoles), UpdatedByName: r.UpdatedByName, UpdatedAt: &t})
	}
	return out, nil
}

// OwnerOverrideInput is one entry of PUT pof-settings owner_overrides
type OwnerOverrideInput struct {
	PrincipleCode string   `json:"principle_code"`
	OwnerRoles    []string `json:"owner_roles"`
}

// PofSettingsUpdate is the body of PUT /ic-assessments/pof-settings
type PofSettingsUpdate struct {
	PofYesThreshold     *float64             `json:"pof_yes_threshold"`
	PofPartialThreshold *float64             `json:"pof_partial_threshold"`
	OwnerOverrides      []OwnerOverrideInput `json:"owner_overrides"`
}

// UpdatePofSettings validates and stores the thresholds and owner overrides
// (ADMIN). Thresholds apply to assessments created afterwards.
func (s *Service) UpdatePofSettings(actor Actor, req PofSettingsUpdate) (*PofSettings, error) {
	if !actor.IsAdmin() {
		return nil, forbiddenf("Only ADMIN can change the PoF settings")
	}
	if req.PofYesThreshold == nil && req.PofPartialThreshold == nil && req.OwnerOverrides == nil {
		return nil, FieldErrors{"body": "send pof_yes_threshold, pof_partial_threshold or owner_overrides"}
	}
	fields := FieldErrors{}
	cur, _, err := LoadPofThresholds(s.DB)
	if err != nil {
		return nil, err
	}
	next := cur
	if req.PofYesThreshold != nil {
		next.Yes = *req.PofYesThreshold
	}
	if req.PofPartialThreshold != nil {
		next.Partial = *req.PofPartialThreshold
	}
	if req.PofYesThreshold != nil || req.PofPartialThreshold != nil {
		if err := next.Validate(); err != nil {
			key := "pof_yes_threshold"
			if req.PofYesThreshold == nil {
				key = "pof_partial_threshold"
			}
			fields[key] = err.Error()
		}
	}
	overrides := map[string]string{}
	for i, o := range req.OwnerOverrides {
		key := "owner_overrides[" + itoa(i) + "]"
		code := strings.ToUpper(strings.TrimSpace(o.PrincipleCode))
		if _, ok := PrincipleByCode(code); !ok {
			fields[key+".principle_code"] = "is not a COSO principle code (P01..P17)"
			continue
		}
		var labels []string
		seen := map[string]bool{}
		for _, l := range o.OwnerRoles {
			l = strings.ToUpper(strings.TrimSpace(l))
			if !OwnerLabelRe.MatchString(l) {
				fields[key+".owner_roles"] = "labels use A-Z, 0-9 and _ (2-40 characters), for example HEAD_OF_IT"
				break
			}
			if !seen[l] {
				seen[l] = true
				labels = append(labels, l)
			}
		}
		if len(strings.Join(labels, ",")) > 200 {
			fields[key+".owner_roles"] = "too many labels"
		}
		overrides[code] = strings.Join(labels, ",")
	}
	if len(fields) > 0 {
		return nil, fields
	}
	now := s.now()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if req.PofYesThreshold != nil || req.PofPartialThreshold != nil {
			// make sure the row exists (migrate creates it; tests and fresh stacks may not)
			if _, _, err := EnsureSettings(tx, func(string) string { return "" }, now); err != nil {
				return err
			}
			if err := tx.Model(&models.ICAssessmentSettings{}).Where("id = ?", models.ICAssessmentSettingsID).Updates(map[string]interface{}{
				"pof_yes_threshold": next.Yes, "pof_partial_threshold": next.Partial,
				"updated_by_user_id": actor.UserID, "updated_by_name": actor.Name, "updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		for code, labels := range overrides {
			if labels == "" {
				if err := tx.Where("principle_code = ?", code).Delete(&models.ICPofOwnerOverride{}).Error; err != nil {
					return err
				}
				continue
			}
			row := models.ICPofOwnerOverride{PrincipleCode: code, OwnerRoles: labels, UpdatedByUserID: actor.UserID,
				UpdatedByName: actor.Name, CreatedAt: now, UpdatedAt: now}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "principle_code"}},
				DoUpdates: clause.AssignmentColumns([]string{"owner_roles", "updated_by_user_id", "updated_by_name", "updated_at"}),
			}).Create(&row).Error; err != nil {
				return err
			}
		}
		return writeEvent(tx, actor, nil, ActionPofSettingsUpdate, "", "", "",
			map[string]interface{}{"pof_yes_threshold": cur.Yes, "pof_partial_threshold": cur.Partial},
			map[string]interface{}{"pof_yes_threshold": next.Yes, "pof_partial_threshold": next.Partial, "owner_overrides": overrides}, now)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPofSettings(actor)
}

func itoa(i int) string { return strconv.Itoa(i) }
