package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// COSO 2013 mapping on the RCM (docs/coso-2013-component-assessment-plan.md
// 4.1, 8.1): request parsing and validation, tag reads/writes, filters, the
// catalogue endpoint and the risk-side signals endpoint.

// optionalString tells an omitted JSON field (Set=false) from an explicit
// null (Set=true, Value=nil) and a value.
type optionalString struct {
	Set   bool
	Value *string
}

func (o *optionalString) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// fieldErrors is a 422 with one message per request field.
type fieldErrors map[string]string

func respondValidation(c *gin.Context, fields fieldErrors) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "Invalid request",
			"fields":  fields,
		},
	})
}

func respondDB(c *gin.Context, what string, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "DB_ERROR",
			"message": what + ": " + err.Error(),
		},
	})
}

func respondDepartmentsUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "DEPARTMENT_MASTER_UNAVAILABLE",
			"message": "Department master data is unavailable; try again shortly",
		},
	})
}

func respondLocationsUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "LOCATION_MASTER_UNAVAILABLE",
			"message": "Location master data is unavailable; try again shortly",
		},
	})
}

// cosoCatalogue is a snapshot of the COSO 2013 rows (22 rows, read per request).
type cosoCatalogue struct {
	components map[string]models.CosoComponent
	principles map[string]models.CosoPrinciple // by code
	numberOf   map[string]int
}

func loadCosoCatalogue(db *gorm.DB) (*cosoCatalogue, error) {
	var comps []models.CosoComponent
	if err := db.Where("framework = ?", models.CosoFramework2013).Find(&comps).Error; err != nil {
		return nil, err
	}
	var prins []models.CosoPrinciple
	if err := db.Where("framework = ?", models.CosoFramework2013).Find(&prins).Error; err != nil {
		return nil, err
	}
	cat := &cosoCatalogue{
		components: map[string]models.CosoComponent{},
		principles: map[string]models.CosoPrinciple{},
		numberOf:   map[string]int{},
	}
	for _, c := range comps {
		cat.components[c.Code] = c
	}
	for _, p := range prins {
		cat.principles[p.Code] = p
		cat.numberOf[p.Code] = p.Number
	}
	return cat, nil
}

func normalizeCode(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

// sortCodes orders principle codes by principle number.
func (cat *cosoCatalogue) sortCodes(codes []string) {
	sort.SliceStable(codes, func(i, j int) bool {
		ni, nj := cat.numberOf[codes[i]], cat.numberOf[codes[j]]
		if ni != nj {
			return ni < nj
		}
		return codes[i] < codes[j]
	})
}

// cosoWrite is the validated COSO/department part of a create or update.
type cosoWrite struct {
	departmentID   *uuid.UUID
	departmentName string // master name, "" when not looked up
	locationID     *uuid.UUID
	componentCode  *string
	// tagsChanged: the request carried coso_principle_codes, so the row's
	// tags are replaced by principleIDs/codes in the write transaction.
	tagsChanged  bool
	principleIDs []uuid.UUID
	codes        []string // final codes (new or existing), sorted
}

// resolveCosoWrite validates the department and COSO fields of req against
// the current row state (zero values on create). It returns field errors
// (422), errDepartmentsUnavailable (503) or a DB error.
func (ctrl *RCMController) resolveCosoWrite(ctx context.Context, req *RCMRequest, cur *models.RiskControlMatrix, curCodes []string) (*cosoWrite, fieldErrors, error) {
	fields := fieldErrors{}
	w := &cosoWrite{departmentID: cur.DepartmentID, locationID: cur.LocationID, componentCode: cur.CosoComponentCode}

	// location_id: an active row of the Location master. A row keeps a
	// location that was since deactivated or removed; a new link must name a
	// live one.
	if req.LocationID.Set {
		raw := ""
		if req.LocationID.Value != nil {
			raw = strings.TrimSpace(*req.LocationID.Value)
		}
		if raw == "" {
			w.locationID = nil
		} else if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			fields["location_id"] = "must be the UUID of a Location master row"
		} else {
			unchanged := cur.LocationID != nil && *cur.LocationID == id
			if !unchanged {
				if ctrl.locations == nil {
					return nil, nil, errLocationsUnavailable
				}
				locs, err := ctrl.locations.ListLocations(ctx)
				if err != nil {
					return nil, nil, errLocationsUnavailable
				}
				found := false
				for _, l := range locs {
					if l.ID == id && l.IsActive {
						found = true
						break
					}
				}
				if !found {
					fields["location_id"] = "is not an active location in the master data"
				}
			}
			w.locationID = &id
		}
	}

	// department_id
	if req.DepartmentID.Set {
		raw := ""
		if req.DepartmentID.Value != nil {
			raw = strings.TrimSpace(*req.DepartmentID.Value)
		}
		if raw == "" {
			w.departmentID = nil
		} else if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			fields["department_id"] = "must be the UUID of a master department"
		} else {
			unchanged := cur.DepartmentID != nil && *cur.DepartmentID == id
			if ctrl.departments == nil {
				if !unchanged {
					return nil, nil, errDepartmentsUnavailable
				}
			} else {
				dept, err := ctrl.departments.LookupDepartment(ctx, id)
				switch {
				case err == nil:
					w.departmentName = dept.Name
				case errors.Is(err, masterclient.ErrDepartmentNotFound):
					// A row keeps a department that was since removed from
					// master; a new link must name a live one.
					if !unchanged {
						fields["department_id"] = "is not a department in the master data"
					}
				default:
					if !unchanged {
						return nil, nil, errDepartmentsUnavailable
					}
				}
			}
			w.departmentID = &id
		}
	}

	needCatalogue := req.CosoComponentCode.Set || req.CosoPrincipleCodes != nil || w.componentCode != nil
	var cat *cosoCatalogue
	if needCatalogue {
		var err error
		if cat, err = loadCosoCatalogue(ctrl.db.WithContext(ctx)); err != nil {
			return nil, nil, err
		}
	}

	// coso_component_code
	if req.CosoComponentCode.Set {
		code := ""
		if req.CosoComponentCode.Value != nil {
			code = normalizeCode(*req.CosoComponentCode.Value)
		}
		if code == "" {
			w.componentCode = nil
		} else if _, ok := cat.components[code]; !ok {
			fields["coso_component_code"] = fmt.Sprintf("unknown COSO 2013 component %q", code)
		} else {
			w.componentCode = &code
		}
	}

	// coso_principle_codes
	w.codes = append([]string{}, curCodes...)
	if req.CosoPrincipleCodes != nil {
		w.tagsChanged = true
		w.codes = nil
		had := map[string]bool{}
		for _, c := range curCodes {
			had[c] = true
		}
		seen := map[string]bool{}
		var unknown []string
		for _, raw := range *req.CosoPrincipleCodes {
			code := normalizeCode(raw)
			if code == "" || seen[code] {
				continue
			}
			seen[code] = true
			p, ok := cat.principles[code]
			// An inactive principle may stay on a row that already had it.
			if !ok || (!p.IsActive && !had[code]) {
				unknown = append(unknown, code)
				continue
			}
			w.codes = append(w.codes, code)
			w.principleIDs = append(w.principleIDs, p.ID)
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			fields["coso_principle_codes"] = "unknown COSO 2013 principle(s): " + strings.Join(unknown, ", ")
		}
	}
	if cat != nil {
		cat.sortCodes(w.codes)
	}

	// The primary component must be the component of a tagged principle.
	if w.componentCode != nil && fields["coso_component_code"] == "" && fields["coso_principle_codes"] == "" {
		match := false
		for _, code := range w.codes {
			if cat.principles[code].ComponentCode == *w.componentCode {
				match = true
				break
			}
		}
		if !match {
			fields["coso_component_code"] = fmt.Sprintf("component %s must be the component of at least one tagged principle", *w.componentCode)
		}
	}

	if len(fields) > 0 {
		return nil, fields, nil
	}
	return w, nil, nil
}

var (
	errDepartmentsUnavailable = errors.New("department master unavailable")
	errLocationsUnavailable   = errors.New("location master unavailable")
)

// apply copies the validated fields onto the row. prevDepartment is the
// department text the row had before parseRCMRequest overwrote it.
func (w *cosoWrite) apply(target *models.RiskControlMatrix, prevDepartmentID *uuid.UUID, prevDepartment string) {
	target.DepartmentID = w.departmentID
	target.LocationID = w.locationID
	target.CosoComponentCode = w.componentCode
	target.CosoPrincipleCodes = w.codes
	if target.CosoPrincipleCodes == nil {
		target.CosoPrincipleCodes = []string{}
	}
	if w.departmentID == nil {
		return
	}
	// A linked row carries the master name in department.
	switch {
	case w.departmentName != "":
		target.Department = truncate(w.departmentName, 100)
	case prevDepartmentID != nil && *prevDepartmentID == *w.departmentID && prevDepartment != "":
		target.Department = prevDepartment
	}
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// writeTags replaces the row's tags inside tx.
func writeTags(tx *gorm.DB, rcmID uuid.UUID, principleIDs []uuid.UUID, replace bool) error {
	if replace {
		if err := tx.Where("rcm_id = ?", rcmID).Delete(&models.RcmCosoPrinciple{}).Error; err != nil {
			return fmt.Errorf("clear tags: %w", err)
		}
	}
	if len(principleIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]models.RcmCosoPrinciple, 0, len(principleIDs))
	for _, pid := range principleIDs {
		rows = append(rows, models.RcmCosoPrinciple{RcmID: rcmID, PrincipleID: pid, CreatedAt: now})
	}
	if err := tx.Omit(clause.Associations).Create(&rows).Error; err != nil {
		return fmt.Errorf("write tags: %w", err)
	}
	return nil
}

// attachTags fills CosoPrincipleCodes on every row ([] when untagged).
func attachTags(db *gorm.DB, list []models.RiskControlMatrix) error {
	ids := make([]uuid.UUID, 0, len(list))
	for _, r := range list {
		ids = append(ids, r.ID)
	}
	tags, err := services.LoadRCMPrincipleCodes(db, ids)
	if err != nil {
		return err
	}
	for i := range list {
		if codes, ok := tags[list[i].ID]; ok {
			list[i].CosoPrincipleCodes = codes
		} else {
			list[i].CosoPrincipleCodes = []string{}
		}
	}
	return nil
}

// rcmFilterValue is the special department_id filter for rows without one.
const rcmFilterUnassigned = "unassigned"

// applyRCMFilters adds the typed list/summary filters. Values are bound, never
// used as identifiers. Invalid values are field errors.
func applyRCMFilters(c *gin.Context, db *gorm.DB, query *gorm.DB) (*gorm.DB, fieldErrors, error) {
	fields := fieldErrors{}
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", y)
		}
	}
	if dept := c.Query("department"); dept != "" && dept != "All Departments" && dept != "All Branches" {
		query = query.Where("department = ?", dept)
	}
	if raw := strings.TrimSpace(c.Query("department_id")); raw != "" {
		if strings.EqualFold(raw, rcmFilterUnassigned) {
			query = query.Where("department_id IS NULL")
		} else if id, err := uuid.Parse(raw); err == nil {
			query = query.Where("department_id = ?", id)
		} else {
			fields["department_id"] = "must be a department UUID or \"unassigned\""
		}
	}
	component := normalizeCode(c.Query("component"))
	principle := normalizeCode(c.Query("principle"))
	if component != "" || principle != "" {
		cat, err := loadCosoCatalogue(db)
		if err != nil {
			return nil, nil, err
		}
		if component != "" {
			if _, ok := cat.components[component]; ok {
				query = query.Where("coso_component_code = ?", component)
			} else {
				fields["component"] = fmt.Sprintf("unknown COSO 2013 component %q", component)
			}
		}
		if principle != "" {
			if p, ok := cat.principles[principle]; ok {
				query = query.Where("id IN (?)", db.Model(&models.RcmCosoPrinciple{}).
					Select("rcm_id").Where("principle_id = ?", p.ID))
			} else {
				fields["principle"] = fmt.Sprintf("unknown COSO 2013 principle %q", principle)
			}
		}
	}
	if len(fields) > 0 {
		return nil, fields, nil
	}
	return query, nil, nil
}

// ---- GET /rcm/coso-principles ----

var frameworkPattern = regexp.MustCompile(`^[A-Z0-9_]{1,20}$`)

type cosoPrincipleDTO struct {
	Code          string `json:"code"`
	Number        int    `json:"number"`
	ComponentCode string `json:"component_code"`
	NameEn        string `json:"name_en"`
	IsActive      bool   `json:"is_active"`
}

type cosoComponentDTO struct {
	Code       string             `json:"code"`
	SortOrder  int                `json:"sort_order"`
	NameEn     string             `json:"name_en"`
	Principles []cosoPrincipleDTO `json:"principles"`
}

// ListCosoPrinciples returns the catalogue: components with nested principles
// (inactive principles included, flagged is_active=false).
func (ctrl *RCMController) ListCosoPrinciples(c *gin.Context) {
	framework := normalizeCode(c.DefaultQuery("framework", models.CosoFramework2013))
	if !frameworkPattern.MatchString(framework) {
		respondValidation(c, fieldErrors{"framework": "must be 1-20 characters A-Z, 0-9 or _"})
		return
	}
	db := ctrl.db.WithContext(c.Request.Context())
	var comps []models.CosoComponent
	if err := db.Where("framework = ?", framework).Order("sort_order, code").Find(&comps).Error; err != nil {
		respondDB(c, "Failed to load COSO components", err)
		return
	}
	var prins []models.CosoPrinciple
	if err := db.Where("framework = ?", framework).Order("number, code").Find(&prins).Error; err != nil {
		respondDB(c, "Failed to load COSO principles", err)
		return
	}
	out := make([]cosoComponentDTO, 0, len(comps))
	index := map[string]int{}
	for _, comp := range comps {
		index[comp.Code] = len(out)
		out = append(out, cosoComponentDTO{Code: comp.Code, SortOrder: comp.SortOrder, NameEn: comp.NameEn, Principles: []cosoPrincipleDTO{}})
	}
	for _, p := range prins {
		i, ok := index[p.ComponentCode]
		if !ok {
			continue
		}
		out[i].Principles = append(out[i].Principles, cosoPrincipleDTO{
			Code: p.Code, Number: p.Number, ComponentCode: p.ComponentCode, NameEn: p.NameEn, IsActive: p.IsActive,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"framework": framework, "components": out},
	})
}

// ---- GET /rcm/coso-signals ----

// signalsTZ is the business time zone used for the default as_of date (same
// zone audit-service derives ATR overdue in).
var signalsTZ = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*3600)
}()

// departmentLister is the part of *masterclient.DepartmentClient that lists
// every department.
type departmentLister interface {
	ListDepartments(ctx context.Context) ([]masterclient.Department, error)
}

// legacyDepartmentName is the master name of id when legacy RCM rows (no
// department_id) may be matched to it by name: id must be active and the only
// active department with that name (the backfill-rcm-department rule). ""
// when the master cannot be read or the name is ambiguous.
func (ctrl *RCMController) legacyDepartmentName(ctx context.Context, id uuid.UUID) string {
	lister, ok := ctrl.departments.(departmentLister)
	if !ok || lister == nil {
		return ""
	}
	depts, err := lister.ListDepartments(ctx)
	if err != nil {
		return ""
	}
	name := ""
	for _, d := range depts {
		if d.ID == id && d.IsActive {
			name = d.Name
		}
	}
	key := services.NormalizeDepartmentName(name)
	if key == "" {
		return ""
	}
	n := 0
	for _, d := range depts {
		if d.IsActive && services.NormalizeDepartmentName(d.Name) == key {
			n++
		}
	}
	if n != 1 {
		return ""
	}
	return name
}

// GetCosoSignals returns the risk-side signals of the COSO 2013 assessment
// for one department, period and optional branch. Read-only.
func (ctrl *RCMController) GetCosoSignals(c *gin.Context) {
	fields := fieldErrors{}
	var p services.CosoSignalsParams

	if id, err := uuid.Parse(strings.TrimSpace(c.Query("department_id"))); err != nil || id == uuid.Nil {
		fields["department_id"] = "required, the UUID of a master department"
	} else {
		p.DepartmentID = id
	}
	if y, err := strconv.Atoi(strings.TrimSpace(c.Query("year"))); err != nil || y < 2000 || y > 2100 {
		fields["year"] = "required, a year between 2000 and 2100"
	} else {
		p.Year = y
	}
	if raw := strings.TrimSpace(c.Query("quarter")); raw != "" {
		if q, err := strconv.Atoi(raw); err != nil || q < 0 || q > 4 {
			fields["quarter"] = "must be 0 (annual) or 1-4"
		} else {
			p.Quarter = q
		}
	}
	if raw := strings.TrimSpace(c.Query("location_id")); raw != "" {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			fields["location_id"] = "must be the UUID of a Location master row"
		} else {
			p.LocationID = &id
		}
	}
	var asOf *time.Time
	if raw := strings.TrimSpace(c.Query("as_of")); raw != "" {
		if d, err := time.Parse("2006-01-02", raw); err != nil {
			fields["as_of"] = "must be a date YYYY-MM-DD"
		} else {
			asOf = &d
		}
	}
	if len(fields) > 0 {
		respondValidation(c, fields)
		return
	}

	// as_of defaults to today (Asia/Jakarta) and never runs past the period end.
	_, periodEnd := services.PeriodBounds(p.Year, p.Quarter)
	if asOf == nil {
		now := ctrl.now().In(signalsTZ)
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		asOf = &d
	}
	if asOf.After(periodEnd) {
		asOf = &periodEnd
	}
	p.AsOf = *asOf

	ctx := masterclient.WithAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
	p.LegacyDepartmentName = ctrl.legacyDepartmentName(ctx, p.DepartmentID)
	res, err := services.ComputeCosoSignals(ctx, ctrl.db, ctrl.locations, p)
	if err != nil {
		respondDB(c, "Failed to compute COSO signals", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// ---- GET /rcm/coso-pof-catalogue ----

// OwnerRoleLabel describes one default owner label of the PoF catalogue and
// the auth-service role it matches, if any. A label with no auth role is
// assigned to a user together with the free label; no role is ever created.
type ownerRoleLabelDTO struct {
	Label    string `json:"label"`
	NameEn   string `json:"name_en"`
	NameID   string `json:"name_id"`
	AuthRole string `json:"auth_role"`
}

func ownerRoleLabels() []ownerRoleLabelDTO {
	return []ownerRoleLabelDTO{
		{"EXECUTIVE", "Executive / Board of Directors", "Eksekutif / Direksi", "EXECUTIVE"},
		{"AUDIT_COMMITTEE", "Audit Committee", "Komite Audit", ""},
		{"CORPORATE_SECRETARY", "Corporate Secretary", "Sekretaris Perusahaan", ""},
		{"HEAD_OF_HR", "Head of Human Resources", "Kepala SDM", ""},
		{"RISK_MANAGEMENT", "Risk Management (second line)", "Manajemen Risiko (lini 2)", ""},
		{"DEPARTMENT_HEAD", "Head of Department", "Kepala Departemen", "DEPARTMENT_HEAD"},
		{"HEAD_OF_IT", "Head of IT", "Kepala TI", ""},
		{"CORPORATE_COMMUNICATION", "Corporate Communication", "Komunikasi Korporat", ""},
		{"MANAGEMENT", "Management", "Manajemen", ""},
	}
}

// GetCosoPofCatalogue returns one version of the PoF catalogue (default: the
// current one) with the weights, elements and default owner labels. Read-only,
// any authenticated role. audit-service reads it when it creates a
// quantitative assessment and copies it into the assessment (a snapshot per
// catalog_version).
func (ctrl *RCMController) GetCosoPofCatalogue(c *gin.Context) {
	framework := normalizeCode(c.DefaultQuery("framework", models.CosoPofFramework))
	if !frameworkPattern.MatchString(framework) {
		respondValidation(c, fieldErrors{"framework": "must be 1-20 characters A-Z, 0-9 or _"})
		return
	}
	version := strings.TrimSpace(c.Query("version"))
	if len(version) > 20 {
		respondValidation(c, fieldErrors{"version": "must be at most 20 characters"})
		return
	}
	cat, err := services.LoadCosoPofCatalogue(ctrl.db.WithContext(c.Request.Context()), framework, version)
	if errors.Is(err, services.ErrCatalogNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "COSO PoF catalogue version not found; run `risk up`"}})
		return
	}
	if err != nil {
		respondDB(c, "Failed to load the COSO PoF catalogue", err)
		return
	}
	issues := services.ValidateCosoPofCatalogue(cat)
	if len(issues) > 0 {
		// A stored catalogue that breaks an invariant is never served.
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{
			"code": "CATALOGUE_INVALID", "message": "The stored COSO PoF catalogue breaks a weight invariant", "details": issues}})
		return
	}
	pofs, els := services.CountPofElements(cat)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"catalogue": cat, "counts": gin.H{"pofs": pofs, "elements": els}, "owner_roles": ownerRoleLabels(),
	}})
}
