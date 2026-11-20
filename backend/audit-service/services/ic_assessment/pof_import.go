package ic_assessment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"audit-service/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Import workflow: preview -> (remap) -> apply (docs/coso-pof-assessment-api.md
// section 8) and the mapping profiles of section 9.

const maxPreviewRows = 1000

// ImportSheetInfo describes one sheet of the uploaded workbook
type ImportSheetInfo struct {
	Name      string `json:"name"`
	Rows      int    `json:"rows"`
	Selected  bool   `json:"selected"`
	HeaderRow int    `json:"header_row"`
	Source    string `json:"mapping_source"` // fixed | profile | request | detected | none
}

// ImportColumn is one column of the header row with what it was mapped to
type ImportColumn struct {
	Index    int    `json:"index"`
	Letter   string `json:"letter"`
	Header   string `json:"header"`
	MappedTo string `json:"mapped_to"`
}

// ImportProfileRef tells which saved profile was used
type ImportProfileRef struct {
	ID      *uuid.UUID `json:"id"`
	Name    string     `json:"name"`
	Matched bool       `json:"matched"`
}

// PreviewResult is the answer of preview and remap
type PreviewResult struct {
	ImportID  uuid.UUID         `json:"import_id"`
	Status    string            `json:"status"`
	Format    string            `json:"format"`
	FileName  string            `json:"file_name"`
	SHA256    string            `json:"sha256"`
	Sheets    []ImportSheetInfo `json:"sheets"`
	Sheet     string            `json:"sheet"`
	HeaderRow int               `json:"header_row"`
	Profile   ImportProfileRef  `json:"profile"`
	Mapping   *ImportMapping    `json:"mapping"`
	Columns   []ImportColumn    `json:"columns"`
	Summary   PlanSummary       `json:"summary"`
	Rows      []PlanRow         `json:"rows"`
	RowsShown int               `json:"rows_shown"`
	Unmatched []UnmatchedRow    `json:"unmatched"`
	Warnings  []string          `json:"warnings"`
}

// PreviewInput is one uploaded workbook
type PreviewInput struct {
	FileName  string
	Content   []byte
	Format    string // auto | auditsphere | external
	ProfileID *uuid.UUID
	Mapping   *ImportMapping
	Sheet     string
}

type interpretation struct {
	format   string
	mapping  *ImportMapping
	sheets   []ImportSheetInfo
	parsed   []ParsedRow
	warnings []string
	profile  ImportProfileRef
	columns  []ImportColumn
	selected string
	// sheetMaps is the mapping used for each parsed sheet; stored with the
	// import so apply reads the file exactly as the preview did
	sheetMaps map[string]*ImportMapping
}

func columnLetter(i int) string {
	n, _ := excelize.ColumnNumberToName(i)
	return n
}

func headerColumns(rows [][]string, m *ImportMapping) []ImportColumn {
	if m.HeaderRow < 1 || m.HeaderRow > len(rows) {
		return []ImportColumn{}
	}
	rev := map[int]string{}
	for k, c := range m.Columns {
		if c > 0 {
			rev[c] = k
		}
	}
	var out []ImportColumn
	row := rows[m.HeaderRow-1]
	for i, h := range row {
		if h == "" && rev[i+1] == "" {
			continue
		}
		out = append(out, ImportColumn{Index: i + 1, Letter: columnLetter(i + 1), Header: h, MappedTo: rev[i+1]})
	}
	if out == nil {
		out = []ImportColumn{}
	}
	return out
}

func isComponentSheet(name string) bool {
	return IsComponentCode(strings.ToUpper(strings.TrimSpace(name)))
}

// interpret reads the workbook with the mapping rules: a fixed layout for an
// AuditSphere workbook; otherwise the explicit mapping, else the best saved
// profile, else header detection, per sheet
func interpret(wb *WorkbookData, format string, explicit *ImportMapping, profile *models.ICPoFImportProfile, profiles []models.ICPoFImportProfile, sheetName string) (*interpretation, error) {
	it := &interpretation{format: wb.Format, warnings: append([]string{}, wb.Skipped...), sheetMaps: map[string]*ImportMapping{}}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "auto":
	case "auditsphere":
		if wb.Format != FormatAuditSphere {
			return nil, FieldErrors{"format": "the workbook is not an AuditSphere export (no _meta marker)"}
		}
	case "external":
		it.format = FormatExternal
	default:
		return nil, FieldErrors{"format": "must be auto, auditsphere or external"}
	}
	trusted := it.format == FormatAuditSphere
	var re *regexp.Regexp
	if explicit != nil && explicit.SheetPattern != "" {
		re, _ = regexp.Compile(explicit.SheetPattern)
	}
	var parsedProfiles []*ImportMapping
	for i := range profiles {
		if m, err := ParseMappingJSON(profiles[i].Mapping); err == nil {
			parsedProfiles = append(parsedProfiles, m)
		} else {
			parsedProfiles = append(parsedProfiles, nil)
		}
	}
	var profMapping *ImportMapping
	if profile != nil {
		profMapping, _ = ParseMappingJSON(profile.Mapping)
		it.profile = ImportProfileRef{ID: &profile.ID, Name: profile.Name, Matched: true}
	}
	if explicit != nil && explicit.Sheet != "" && sheetName == "" {
		sheetName = explicit.Sheet
	}
	for _, sd := range wb.Sheets {
		info := ImportSheetInfo{Name: sd.Name, Rows: len(sd.Rows), Source: "none"}
		skip := false
		switch {
		case sheetName != "" && !strings.EqualFold(sd.Name, sheetName):
			skip = true
		case trusted && !isComponentSheet(sd.Name):
			skip = true
		case re != nil && !re.MatchString(sd.Name):
			skip = true
		}
		if skip {
			it.sheets = append(it.sheets, info)
			continue
		}
		var m *ImportMapping
		switch {
		case trusted:
			m, info.Source = AuditSphereMapping(), "fixed"
		case explicit != nil:
			cp := *explicit
			m, info.Source = &cp, "request"
		case profMapping != nil:
			cp := *profMapping
			m, info.Source = &cp, "profile"
		default:
			// best saved profile of the tenant, then detection
			bestFit := 0.0
			var bestIdx = -1
			for i, pm := range parsedProfiles {
				if pm == nil {
					continue
				}
				if pm.SheetPattern != "" {
					if r, err := regexp.Compile(pm.SheetPattern); err != nil || !r.MatchString(sd.Name) {
						continue
					}
				}
				if fit := ProfileFit(sd.Rows, pm); fit > bestFit {
					bestFit, bestIdx = fit, i
				}
			}
			if bestIdx >= 0 && bestFit >= 0.6 {
				cp := *parsedProfiles[bestIdx]
				m, info.Source = &cp, "profile"
				if it.profile.ID == nil {
					it.profile = ImportProfileRef{ID: &profiles[bestIdx].ID, Name: profiles[bestIdx].Name, Matched: true}
				}
			} else if dm, ok := DetectHeader(sd.Rows, nil); ok {
				m, info.Source = dm, "detected"
			}
		}
		if m == nil {
			it.warnings = append(it.warnings, fmt.Sprintf("sheet %q skipped: no header row found", sd.Name))
			it.sheets = append(it.sheets, info)
			continue
		}
		if m.HeaderRow > len(sd.Rows) {
			it.warnings = append(it.warnings, fmt.Sprintf("sheet %q skipped: the header row %d is past the end of the sheet", sd.Name, m.HeaderRow))
			it.sheets = append(it.sheets, info)
			continue
		}
		if sd.Truncated {
			it.warnings = append(it.warnings, fmt.Sprintf("sheet %q has more than %d rows; only the first %d were read", sd.Name, MaxImportRows, MaxImportRows))
		}
		info.Selected, info.HeaderRow = true, m.HeaderRow
		cpm := *m
		it.sheetMaps[sd.Name] = &cpm
		rows := ParseSheet(sd, m, trusted)
		it.parsed = append(it.parsed, rows...)
		if it.mapping == nil {
			cp := *m
			it.mapping = &cp
			it.selected = sd.Name
			it.columns = headerColumns(sd.Rows, m)
		}
		it.sheets = append(it.sheets, info)
	}
	if it.mapping == nil {
		return it, FieldErrors{"file": "no sheet with a recognisable header row was found; send a mapping (header_row and columns) or choose a saved profile"}
	}
	if trusted {
		if v := wb.Meta["catalog_version"]; v != "" {
			it.warnings = append(it.warnings, "workbook catalogue version: "+v)
		}
	}
	return it, nil
}

// storedMapping is what an import row keeps in its mapping column
type storedMapping struct {
	Format string                    `json:"format"`
	Top    *ImportMapping            `json:"top"`
	Sheets map[string]*ImportMapping `json:"sheets"`
}

func (it *interpretation) stored() string {
	b, _ := json.Marshal(storedMapping{Format: it.format, Top: it.mapping, Sheets: it.sheetMaps})
	return string(b)
}

// reparse reads a workbook with the sheet mappings saved at preview time
func reparse(wb *WorkbookData, stored string, profiles []models.ICPoFImportProfile) (*interpretation, error) {
	var sm storedMapping
	if err := json.Unmarshal([]byte(stored), &sm); err != nil || sm.Top == nil {
		return nil, fmt.Errorf("the stored import mapping is not readable")
	}
	it := &interpretation{format: sm.Format, mapping: sm.Top, sheetMaps: sm.Sheets, warnings: append([]string{}, wb.Skipped...)}
	trusted := sm.Format == FormatAuditSphere
	for _, sd := range wb.Sheets {
		info := ImportSheetInfo{Name: sd.Name, Rows: len(sd.Rows), Source: "none"}
		if m := sm.Sheets[sd.Name]; m != nil && m.HeaderRow <= len(sd.Rows) {
			info.Selected, info.HeaderRow, info.Source = true, m.HeaderRow, "stored"
			it.parsed = append(it.parsed, ParseSheet(sd, m, trusted)...)
			if it.selected == "" {
				it.selected = sd.Name
				it.columns = headerColumns(sd.Rows, m)
			}
		}
		it.sheets = append(it.sheets, info)
	}
	return it, nil
}

func (it *interpretation) toResult(imp *models.ICPoFImport, plan *Plan) *PreviewResult {
	res := &PreviewResult{ImportID: imp.ID, Status: imp.Status, Format: it.format, FileName: imp.FileName, SHA256: imp.Sha256,
		Sheets: it.sheets, Sheet: it.selected, HeaderRow: it.mapping.HeaderRow, Profile: it.profile, Mapping: it.mapping,
		Columns: it.columns, Summary: plan.Summary, Unmatched: plan.Unmatched, Warnings: append(append([]string{}, it.warnings...), plan.Warnings...)}
	res.Rows = plan.Rows
	res.RowsShown = len(plan.Rows)
	if len(res.Rows) > maxPreviewRows {
		res.Rows = res.Rows[:maxPreviewRows]
		res.Warnings = append(res.Warnings, fmt.Sprintf("only the first %d rows are listed; counts cover all rows", maxPreviewRows))
		res.RowsShown = maxPreviewRows
	}
	if res.Sheets == nil {
		res.Sheets = []ImportSheetInfo{}
	}
	return res
}

func (s *Service) loadQuantEditable(tx *gorm.DB, actor Actor, id uuid.UUID) (*models.ICAssessment, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can import values")
	}
	a, err := loadFull(tx, id)
	if err != nil {
		return nil, err
	}
	if a.Method != MethodQuantitativePoF {
		return nil, methodMismatch(a.Method, "PoF import exists only for the QUANTITATIVE_POF method")
	}
	if !isEditable(a.Status) {
		return nil, conflictf("INVALID_STATUS", "Values can only be imported while the assessment is DRAFT or REOPENED (current status: %s)", a.Status)
	}
	if err := s.mustBeOpen(tx, a.Year, a.Quarter); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) loadProfiles() ([]models.ICPoFImportProfile, error) {
	var ps []models.ICPoFImportProfile
	err := s.DB.Order("name_key").Find(&ps).Error
	return ps, err
}

func (s *Service) profileByID(id uuid.UUID) (*models.ICPoFImportProfile, error) {
	var p models.ICPoFImportProfile
	if err := s.DB.Where("id = ?", id).Take(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, FieldErrors{"profile_id": "is not a saved import profile"}
		}
		return nil, err
	}
	return &p, nil
}

// PreviewImport stores the workbook and returns the interpretation of it
func (s *Service) PreviewImport(ctx context.Context, actor Actor, id uuid.UUID, in PreviewInput) (*PreviewResult, error) {
	a, err := s.loadQuantEditable(s.DB, actor, id)
	if err != nil {
		return nil, err
	}
	if in.Mapping != nil {
		if fe := ValidateMapping(in.Mapping, "mapping."); len(fe) > 0 {
			return nil, fe
		}
	}
	wb, err := ReadWorkbook(in.Content)
	if err != nil {
		return nil, err
	}
	var profile *models.ICPoFImportProfile
	if in.ProfileID != nil {
		if profile, err = s.profileByID(*in.ProfileID); err != nil {
			return nil, err
		}
	}
	profiles, err := s.loadProfiles()
	if err != nil {
		return nil, err
	}
	it, err := interpret(wb, in.Format, in.Mapping, profile, profiles, in.Sheet)
	if err != nil {
		return nil, err
	}
	pofs, els, err := loadPoFRows(s.DB, a.ID)
	if err != nil {
		return nil, err
	}
	plan := BuildPlan(it.parsed, pofs, els, it.format == FormatAuditSphere)
	if it.format == FormatAuditSphere {
		if aid := wb.Meta["assessment_id"]; aid != "" && aid != a.ID.String() {
			plan.Warnings = append(plan.Warnings, "the workbook was exported from another assessment; rows are matched by code only")
		}
	}
	sum := sha256.Sum256(in.Content)
	imp := &models.ICPoFImport{AssessmentID: a.ID, Status: models.ICImportPreviewed, Format: it.format, ProfileID: it.profile.ID,
		FileName: truncate(in.FileName, 255), SizeBytes: int64(len(in.Content)), Sha256: hex.EncodeToString(sum[:]), Content: in.Content,
		Mapping: it.stored(), Sheet: truncate(in.Sheet, 100), CreatedByUserID: actor.UserID, CreatedByName: actor.Name, CreatedAt: s.now()}
	res := it.toResult(imp, plan)
	pj, _ := json.Marshal(res.Summary)
	imp.Preview = string(pj)
	if err := s.DB.Create(imp).Error; err != nil {
		return nil, err
	}
	res.ImportID = imp.ID
	return res, nil
}

func (s *Service) loadImport(tx *gorm.DB, assessmentID, importID uuid.UUID) (*models.ICPoFImport, error) {
	var imp models.ICPoFImport
	if err := tx.Where("id = ? AND assessment_id = ?", importID, assessmentID).Take(&imp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrorsNotFound("Import not found")
		}
		return nil, err
	}
	return &imp, nil
}

// RemapRequest is the body of POST /:id/pof-import/:import_id/remap
type RemapRequest struct {
	Mapping   *ImportMapping `json:"mapping"`
	Sheet     string         `json:"sheet"`
	ProfileID *uuid.UUID     `json:"profile_id"`
}

// RemapImport re-reads the stored workbook with another mapping or profile
func (s *Service) RemapImport(ctx context.Context, actor Actor, id, importID uuid.UUID, req RemapRequest) (*PreviewResult, error) {
	a, err := s.loadQuantEditable(s.DB, actor, id)
	if err != nil {
		return nil, err
	}
	imp, err := s.loadImport(s.DB, id, importID)
	if err != nil {
		return nil, err
	}
	if imp.Status != models.ICImportPreviewed {
		return nil, conflictf("IMPORT_ALREADY_APPLIED", "This import was already applied")
	}
	if req.Mapping == nil && req.ProfileID == nil {
		return nil, FieldErrors{"mapping": "send a mapping or a profile_id"}
	}
	if req.Mapping != nil {
		if fe := ValidateMapping(req.Mapping, "mapping."); len(fe) > 0 {
			return nil, fe
		}
	}
	wb, err := ReadWorkbook(imp.Content)
	if err != nil {
		return nil, err
	}
	var profile *models.ICPoFImportProfile
	if req.ProfileID != nil && req.Mapping == nil {
		if profile, err = s.profileByID(*req.ProfileID); err != nil {
			return nil, err
		}
	}
	profiles, err := s.loadProfiles()
	if err != nil {
		return nil, err
	}
	format := "external"
	if imp.Format == FormatAuditSphere {
		format = "auditsphere"
	}
	it, err := interpret(wb, format, req.Mapping, profile, profiles, req.Sheet)
	if err != nil {
		return nil, err
	}
	pofs, els, err := loadPoFRows(s.DB, a.ID)
	if err != nil {
		return nil, err
	}
	plan := BuildPlan(it.parsed, pofs, els, it.format == FormatAuditSphere)
	imp.Mapping, imp.Sheet, imp.ProfileID = it.stored(), truncate(req.Sheet, 100), it.profile.ID
	pj, _ := json.Marshal(plan.Summary)
	imp.Preview = string(pj)
	if err := s.DB.Model(&models.ICPoFImport{}).Where("id = ?", imp.ID).Updates(map[string]interface{}{
		"mapping": imp.Mapping, "sheet": imp.Sheet, "profile_id": imp.ProfileID, "preview": imp.Preview}).Error; err != nil {
		return nil, err
	}
	return it.toResult(imp, plan), nil
}

// GetImportPreview recomputes the preview of a stored import (current values)
func (s *Service) GetImportPreview(ctx context.Context, actor Actor, id, importID uuid.UUID) (*PreviewResult, error) {
	a, err := s.loadQuantEditable(s.DB, actor, id)
	if err != nil {
		return nil, err
	}
	imp, err := s.loadImport(s.DB, id, importID)
	if err != nil {
		return nil, err
	}
	wb, err := ReadWorkbook(imp.Content)
	if err != nil {
		return nil, err
	}
	it, err := reparse(wb, imp.Mapping, nil)
	if err != nil {
		return nil, err
	}
	pofs, els, err := loadPoFRows(s.DB, a.ID)
	if err != nil {
		return nil, err
	}
	return it.toResult(imp, BuildPlan(it.parsed, pofs, els, it.format == FormatAuditSphere)), nil
}

// DownloadImport returns the original uploaded file (assessors)
func (s *Service) DownloadImport(ctx context.Context, actor Actor, id, importID uuid.UUID) (*models.ICPoFImport, error) {
	if !actor.CanRead() || actor.IsScopedReader() {
		return nil, forbiddenf("Only assessors and approvers can download an imported file")
	}
	imp, err := s.loadImport(s.DB, id, importID)
	if err != nil {
		return nil, err
	}
	return imp, nil
}

// DiscardImport removes a PREVIEWED import (the stored file is deleted)
func (s *Service) DiscardImport(ctx context.Context, actor Actor, id, importID uuid.UUID) error {
	if !actor.IsAssessor() {
		return forbiddenf("Only assessors can discard an import")
	}
	imp, err := s.loadImport(s.DB, id, importID)
	if err != nil {
		return err
	}
	if imp.Status != models.ICImportPreviewed {
		return conflictf("IMPORT_ALREADY_APPLIED", "An applied import is kept as evidence and cannot be discarded")
	}
	return s.DB.Where("id = ?", imp.ID).Delete(&models.ICPoFImport{}).Error
}

// OnlyRow selects one row of one sheet
type OnlyRow struct {
	Sheet string `json:"sheet"`
	Row   int    `json:"row"`
}

// ApplyRequest is the body of POST /:id/pof-import/:import_id/apply
type ApplyRequest struct {
	Mode              string    `json:"mode"`
	AcceptTextMatches bool      `json:"accept_text_matches"`
	AcceptConflicts   []string  `json:"accept_conflicts"`
	OnlyRows          []OnlyRow `json:"only_rows"`
	SaveProfileAs     string    `json:"save_profile_as"`
}

// RejectedRow is a row that was not written, with the reason
type RejectedRow struct {
	Row    int    `json:"row"`
	Sheet  string `json:"sheet"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// ApplyResult is the answer of apply
type ApplyResult struct {
	ImportID uuid.UUID `json:"import_id"`
	Status   string    `json:"status"`
	Applied  struct {
		Elements int `json:"elements"`
		Pofs     int `json:"pofs"`
	} `json:"applied"`
	Unchanged        int            `json:"unchanged"`
	SkippedForbidden []RejectedRow  `json:"skipped_forbidden"`
	Rejected         []RejectedRow  `json:"rejected"`
	NotSelected      int            `json:"not_selected"`
	Unmatched        []UnmatchedRow `json:"unmatched"`
	ProfileSaved     *struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	} `json:"profile_saved"`
	Total ScoreBlock `json:"total"`
}

// ApplyImport writes the selected rows of a previewed import, attaches the
// original file as evidence and optionally saves the mapping as a profile
func (s *Service) ApplyImport(ctx context.Context, src Sources, actor Actor, id, importID uuid.UUID, req ApplyRequest) (*ApplyResult, error) {
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "fill_empty"
	}
	fields := FieldErrors{}
	if mode != "fill_empty" && mode != "overwrite" {
		fields["mode"] = "must be fill_empty or overwrite"
	}
	profileName := strings.TrimSpace(req.SaveProfileAs)
	if profileName != "" && utf8.RuneCountInString(profileName) > 120 {
		fields["save_profile_as"] = "must be at most 120 characters"
	}
	if len(fields) > 0 {
		return nil, fields
	}
	only := map[string]bool{}
	for _, o := range req.OnlyRows {
		only[fmt.Sprintf("%s!%d", strings.ToLower(o.Sheet), o.Row)] = true
	}
	acceptConflict := map[string]bool{}
	for _, c := range req.AcceptConflicts {
		acceptConflict[strings.TrimSpace(c)] = true
	}
	now := s.now()
	var result *ApplyResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		a, err := s.loadQuantEditable(tx, actor, id)
		if err != nil {
			return err
		}
		imp, err := s.loadImport(tx, id, importID)
		if err != nil {
			return err
		}
		if imp.Status != models.ICImportPreviewed {
			return conflictf("IMPORT_ALREADY_APPLIED", "This import was already applied")
		}
		wb, err := ReadWorkbook(imp.Content)
		if err != nil {
			return err
		}
		it, err := reparse(wb, imp.Mapping, nil)
		if err != nil {
			return err
		}
		pofs, els, err := loadPoFRows(tx, a.ID)
		if err != nil {
			return err
		}
		plan := BuildPlan(it.parsed, pofs, els, it.format == FormatAuditSphere)
		tree := Calculate(pofs, els, thresholdsOf(a))
		locked := false
		out := &ApplyResult{ImportID: imp.ID, Status: models.ICImportApplied, Unmatched: plan.Unmatched,
			SkippedForbidden: []RejectedRow{}, Rejected: []RejectedRow{}}
		var vr ValuesRequest
		ac := scopeAccess{}
		for _, r := range plan.Rows {
			key := fmt.Sprintf("%s!%d", strings.ToLower(r.Sheet), r.Row)
			switch r.Action {
			case RowInvalid:
				code := ""
				if r.Target != nil {
					code = r.Target.Code
				}
				out.Rejected = append(out.Rejected, RejectedRow{Row: r.Row, Sheet: r.Sheet, Code: code, Reason: strings.Join(r.Issues, "; ")})
				continue
			case RowSame:
				out.Unchanged++
				continue
			case RowConflict:
				if r.ConflictKind != "TEXT" || r.Target == nil || !acceptConflict[r.Target.Code] {
					if r.Target != nil {
						out.Rejected = append(out.Rejected, RejectedRow{Row: r.Row, Sheet: r.Sheet, Code: r.Target.Code, Reason: "conflict: " + strings.Join(r.Issues, "; ")})
					}
					continue
				}
			case RowUpdate:
			default:
				continue
			}
			if len(only) > 0 && !only[key] {
				out.NotSelected++
				continue
			}
			if r.Match == MatchText && !req.AcceptTextMatches && !only[key] {
				out.NotSelected++
				continue
			}
			// what to write: fill_empty only where the assessment has no value
			cur := r.Current
			var ach optionalFloat
			var expl, rec *string
			if r.Achievement != nil && (mode == "overwrite" || cur.Achievement == nil) {
				ach = optionalFloat{Set: true, Value: r.Achievement}
			}
			if r.Explanation != "" && (mode == "overwrite" || cur.Explanation == "") {
				e := r.Explanation
				expl = &e
			}
			if r.Recommendation != "" && (mode == "overwrite" || cur.Recommendation == "") {
				e := r.Recommendation
				rec = &e
			}
			if !ach.Set && expl == nil && rec == nil {
				out.Unchanged++
				continue
			}
			if r.Target.Level == LevelElement {
				e := tree.Element(r.Target.Code)
				pc := tree.PoF(e.PofCode)
				if ferr := s.canEdit(a, locked, actor, ac, pc.Row.PrincipleCode, pc.Row.PofCode, e); ferr != nil {
					out.SkippedForbidden = append(out.SkippedForbidden, RejectedRow{Row: r.Row, Sheet: r.Sheet, Code: r.Target.Code, Reason: ferr.Error()})
					continue
				}
				vr.Elements = append(vr.Elements, ElementValueInput{ElementCode: r.Target.Code, Achievement: ach, Explanation: expl, Recommendation: rec})
			} else {
				vr.Pofs = append(vr.Pofs, PoFValueInput{PofCode: r.Target.Code, Achievement: ach, Explanation: expl, Recommendation: rec})
			}
		}
		out.Applied.Elements, out.Applied.Pofs = len(vr.Elements), len(vr.Pofs)
		if len(vr.Elements)+len(vr.Pofs) > 0 {
			res, err := s.saveValuesTx(tx, ctx, src, actor, id, vr, MaxImportRows)
			if err != nil {
				var fe FieldErrors
				if errors.As(err, &fe) {
					return invalidf("the import produced an invalid value: %s", fe.Error())
				}
				return err
			}
			out.Total = res.Total
		} else {
			full := Calculate(pofs, els, thresholdsOf(a))
			out.Total = full.Total
		}
		// the original file stays as evidence of what was imported
		ev := models.ICPoFEvidence{AssessmentID: id, ScopeType: "ASSESSMENT", Kind: "FILE", Title: "Import: " + imp.FileName, FileName: imp.FileName,
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", SizeBytes: imp.SizeBytes, Sha256: imp.Sha256,
			Content: imp.Content, Source: "IMPORT", ImportID: &imp.ID, UploadedByUserID: actor.UserID, UploadedByName: actor.Name, CreatedAt: now}
		if err := tx.Create(&ev).Error; err != nil {
			return err
		}
		if imp.ProfileID != nil {
			_ = tx.Model(&models.ICPoFImportProfile{}).Where("id = ?", *imp.ProfileID).Updates(map[string]interface{}{
				"use_count": gorm.Expr("use_count + 1"), "last_used_at": now}).Error
		}
		if profileName != "" {
			p, err := upsertProfile(tx, actor, profileName, "", it.mapping, now)
			if err != nil {
				return err
			}
			out.ProfileSaved = &struct {
				ID   uuid.UUID `json:"id"`
				Name string    `json:"name"`
			}{p.ID, p.Name}
		}
		if err := tx.Model(&models.ICPoFImport{}).Where("id = ?", imp.ID).Updates(map[string]interface{}{
			"status": models.ICImportApplied, "applied_by_name": actor.Name, "applied_at": now}).Error; err != nil {
			return err
		}
		result = out
		return writeEvent(tx, actor, &id, ActionPofImport, a.Status, a.Status, "", nil, map[string]interface{}{
			"import_id": imp.ID.String(), "file_name": imp.FileName, "sha256": imp.Sha256, "mode": mode, "format": imp.Format,
			"applied_elements": out.Applied.Elements, "applied_pofs": out.Applied.Pofs, "rejected": len(out.Rejected),
			"unmatched": len(out.Unmatched), "skipped_forbidden": len(out.SkippedForbidden)}, now)
	})
	if err != nil {
		var fe FieldErrors
		if errors.As(err, &fe) {
			return nil, fe
		}
		return nil, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Profiles
// ---------------------------------------------------------------------------

// ProfileDTO is one saved mapping profile
type ProfileDTO struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Format        string         `json:"format"`
	Mapping       *ImportMapping `json:"mapping"`
	UseCount      int            `json:"use_count"`
	LastUsedAt    *time.Time     `json:"last_used_at"`
	CreatedByName string         `json:"created_by_name"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toProfileDTO(p models.ICPoFImportProfile) ProfileDTO {
	m, _ := ParseMappingJSON(p.Mapping)
	return ProfileDTO{ID: p.ID, Name: p.Name, Description: p.Description, Format: p.Format, Mapping: m, UseCount: p.UseCount,
		LastUsedAt: p.LastUsedAt, CreatedByName: p.CreatedByName, CreatedAt: p.CreatedAt}
}

// ListProfiles returns the saved profiles (any reader)
func (s *Service) ListProfiles(actor Actor) ([]ProfileDTO, error) {
	if !actor.CanRead() {
		return nil, forbiddenf("No role on this account may read internal control assessments")
	}
	ps, err := s.loadProfiles()
	if err != nil {
		return nil, err
	}
	out := make([]ProfileDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, toProfileDTO(p))
	}
	return out, nil
}

// ProfileRequest is the body of profile create / update
type ProfileRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Mapping     *ImportMapping `json:"mapping"`
}

func validateProfileRequest(req *ProfileRequest) (string, FieldErrors) {
	f := FieldErrors{}
	name := strings.TrimSpace(req.Name)
	if name == "" || utf8.RuneCountInString(name) > 120 {
		f["name"] = "is required (at most 120 characters)"
	}
	d := strings.TrimSpace(req.Description)
	checkText(f, "description", &d)
	if req.Mapping == nil {
		f["mapping"] = "is required"
	} else {
		for k, v := range ValidateMapping(req.Mapping, "mapping.") {
			f[k] = v
		}
	}
	return name, f
}

func upsertProfile(tx *gorm.DB, actor Actor, name, desc string, m *ImportMapping, now time.Time) (*models.ICPoFImportProfile, error) {
	key := strings.ToLower(name)
	var p models.ICPoFImportProfile
	err := tx.Where("name_key = ?", key).Take(&p).Error
	switch {
	case err == nil:
		upd := map[string]interface{}{"name": name, "mapping": MappingJSON(m), "updated_at": now}
		if desc != "" {
			upd["description"] = desc
		}
		if err := tx.Model(&models.ICPoFImportProfile{}).Where("id = ?", p.ID).Updates(upd).Error; err != nil {
			return nil, err
		}
		p.Name = name
		return &p, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		p = models.ICPoFImportProfile{Name: name, NameKey: key, Description: desc, Format: FormatExternal, Mapping: MappingJSON(m),
			CreatedByUserID: actor.UserID, CreatedByName: actor.Name, CreatedAt: now, UpdatedAt: now}
		if err := tx.Create(&p).Error; err != nil {
			return nil, err
		}
		return &p, nil
	default:
		return nil, err
	}
}

// CreateProfile saves a mapping profile (assessors)
func (s *Service) CreateProfile(actor Actor, req ProfileRequest) (*ProfileDTO, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can save import profiles")
	}
	name, f := validateProfileRequest(&req)
	if len(f) > 0 {
		return nil, f
	}
	var n int64
	if err := s.DB.Model(&models.ICPoFImportProfile{}).Where("name_key = ?", strings.ToLower(name)).Count(&n).Error; err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, FieldErrors{"name": "a profile with this name already exists"}
	}
	p, err := upsertProfile(s.DB, actor, name, strings.TrimSpace(req.Description), req.Mapping, s.now())
	if err != nil {
		return nil, err
	}
	d := toProfileDTO(*p)
	return &d, nil
}

// UpdateProfile changes a profile's name, description and mapping
func (s *Service) UpdateProfile(actor Actor, id uuid.UUID, req ProfileRequest) (*ProfileDTO, error) {
	if !actor.IsAssessor() {
		return nil, forbiddenf("Only assessors can change import profiles")
	}
	name, f := validateProfileRequest(&req)
	if len(f) > 0 {
		return nil, f
	}
	var p models.ICPoFImportProfile
	if err := s.DB.Where("id = ?", id).Take(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrorsNotFound("Import profile not found")
		}
		return nil, err
	}
	var n int64
	if err := s.DB.Model(&models.ICPoFImportProfile{}).Where("name_key = ? AND id <> ?", strings.ToLower(name), id).Count(&n).Error; err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, FieldErrors{"name": "a profile with this name already exists"}
	}
	now := s.now()
	if err := s.DB.Model(&models.ICPoFImportProfile{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": name, "name_key": strings.ToLower(name), "description": strings.TrimSpace(req.Description),
		"mapping": MappingJSON(req.Mapping), "updated_at": now}).Error; err != nil {
		return nil, err
	}
	if err := s.DB.Where("id = ?", id).Take(&p).Error; err != nil {
		return nil, err
	}
	d := toProfileDTO(p)
	return &d, nil
}

// DeleteProfile soft-deletes a profile
func (s *Service) DeleteProfile(actor Actor, id uuid.UUID) error {
	if !actor.IsAssessor() {
		return forbiddenf("Only assessors can delete import profiles")
	}
	res := s.DB.Where("id = ?", id).Delete(&models.ICPoFImportProfile{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrorsNotFound("Import profile not found")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

func safeFileName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	out := b.String()
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

// ExportAssessment builds the workbook of an assessment (blank = template)
func (s *Service) ExportAssessment(ctx context.Context, src Sources, actor Actor, id uuid.UUID, lang string, blank bool) ([]byte, string, error) {
	a, ac, err := s.loadQuant(ctx, src, actor, id)
	if err != nil {
		return nil, "", err
	}
	if ac.restricted {
		return nil, "", forbiddenf("The export covers the whole assessment; it is not available for an assignment-only view")
	}
	pofs, els, err := loadPoFRows(s.DB, id)
	if err != nil {
		return nil, "", err
	}
	meta := WorkbookMeta{AssessmentID: a.ID.String(), Title: a.DepartmentName, Period: periodLabel(a.Year, a.Quarter), CatalogVersion: a.CatalogVersion, Lang: lang, Blank: blank}
	if a.LocationName != "" {
		meta.Title += " / " + a.LocationName
	}
	b, err := BuildWorkbook(meta, pofs, els, thresholdsOf(a))
	if err != nil {
		return nil, "", err
	}
	kind := "export"
	if blank {
		kind = "template"
	}
	return b, fmt.Sprintf("coso-pof-%s-%s-%s.xlsx", safeFileName(a.DepartmentName), periodLabel(a.Year, a.Quarter), kind), nil
}

// ExportTemplate builds the empty workbook of the current catalogue
func (s *Service) ExportTemplate(ctx context.Context, src Sources, actor Actor, lang string) ([]byte, string, error) {
	if !actor.CanRead() || actor.IsScopedReader() {
		return nil, "", forbiddenf("Your role cannot export the template")
	}
	cat, err := src.PofCatalogue(ctx, "")
	if err != nil {
		return nil, "", unavailablef("Could not read the COSO PoF catalogue from risk-service")
	}
	if issues := ValidateCatalogue(cat); len(issues) > 0 {
		return nil, "", invalidList("The COSO PoF catalogue of risk-service breaks a weight invariant", issues)
	}
	pofs, els := MaterializeRows(uuid.Nil, cat, nil, s.now())
	th, _, err := LoadPofThresholds(s.DB)
	if err != nil {
		return nil, "", err
	}
	b, err := BuildWorkbook(WorkbookMeta{Title: "-", Period: "-", CatalogVersion: cat.Version, Lang: lang, Blank: true}, pofs, els, th)
	return b, "coso-pof-template.xlsx", err
}
