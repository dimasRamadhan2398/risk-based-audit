package ic_assessment

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"audit-service/models"

	"github.com/xuri/excelize/v2"
)

// Excel import: reading and interpreting a workbook (docs/coso-pof-assessment-api.md
// sections 8 and 9). Deterministic heuristics only, no ML:
//
//  1. read cells as TEXT (raw cached values; formulas and macros are never run);
//  2. find the header row and the columns from header words (or take a saved
//     mapping profile, or a mapping sent by the client);
//  3. read the hierarchy from the numbering (x.x.x.x) and the Present / Function
//     blocks;
//  4. match each row to the assessment by number, else by text similarity.
//
// Rows that cannot be matched are reported, never written silently.

// Import limits
const (
	MaxImportBytes   = 5 << 20
	MaxImportSheets  = 25
	MaxImportRows    = 5000
	MaxImportCols    = 40
	maxZipEntries    = 2000
	maxUnzipTotal    = 64 << 20
	maxUnzipXMLPart  = 16 << 20
	maxCellRunes     = 4000
	headerScanRows   = 30
	TextMatchMin     = 0.60
	TextConflictMax  = 0.30
	textAmbiguityGap = 0.05
)

// Import formats
const (
	FormatAuditSphere = "AUDITSPHERE"
	FormatExternal    = "EXTERNAL"
)

// Column keys of a mapping
var columnKeys = []string{"no", "description", "type", "methods", "weight", "achievement", "score", "explanation", "recommendation"}

// ImportMapping tells the parser where things are. Columns are 1-based.
type ImportMapping struct {
	Sheet        string              `json:"sheet,omitempty"`
	SheetPattern string              `json:"sheet_pattern,omitempty"`
	HeaderRow    int                 `json:"header_row"`
	Numbering    string              `json:"numbering,omitempty"`
	Columns      map[string]int      `json:"columns"`
	Headers      map[string]string   `json:"headers,omitempty"`
	HeaderWords  map[string][]string `json:"header_words,omitempty"`
}

// ValidateMapping checks a mapping sent by a client; keys are prefixed "mapping."
func ValidateMapping(m *ImportMapping, prefix string) FieldErrors {
	f := FieldErrors{}
	if m.HeaderRow < 1 || m.HeaderRow > 1000 {
		f[prefix+"header_row"] = "must be between 1 and 1000"
	}
	if len(m.Columns) == 0 {
		f[prefix+"columns"] = "is required"
	}
	known := map[string]bool{}
	for _, k := range columnKeys {
		known[k] = true
	}
	for k, v := range m.Columns {
		if !known[k] {
			f[prefix+"columns."+k] = "is not a known column (" + strings.Join(columnKeys, ", ") + ")"
		} else if v < 0 || v > MaxImportCols {
			f[prefix+"columns."+k] = fmt.Sprintf("must be a column number between 1 and %d (0 = not present)", MaxImportCols)
		}
	}
	if m.Columns["no"] == 0 && m.Columns["description"] == 0 {
		f[prefix+"columns"] = "map at least the no or the description column"
	}
	if m.Columns["achievement"] == 0 && m.Columns["score"] == 0 && m.Columns["explanation"] == 0 && m.Columns["recommendation"] == 0 {
		f[prefix+"columns"] = "map at least one of achievement, score, explanation or recommendation"
	}
	if m.SheetPattern != "" {
		if len(m.SheetPattern) > 200 {
			f[prefix+"sheet_pattern"] = "is too long"
		} else if _, err := regexp.Compile(m.SheetPattern); err != nil {
			f[prefix+"sheet_pattern"] = "is not a valid regular expression"
		}
	}
	if len(m.Sheet) > 100 {
		f[prefix+"sheet"] = "is too long"
	}
	return f
}

// ---------------------------------------------------------------------------
// Reading the workbook
// ---------------------------------------------------------------------------

// SheetData is one worksheet read as text
type SheetData struct {
	Name      string
	Rows      [][]string
	Truncated bool
}

// WorkbookData is a workbook read as text
type WorkbookData struct {
	Format  string
	Meta    map[string]string
	Sheets  []SheetData
	Skipped []string
}

// ImportError is a refused file (HTTP 400 / 415)
type ImportError struct {
	Code    string
	Message string
	Status  int
}

func (e *ImportError) Error() string { return e.Message }

func badFile(code, msg string, status int) *ImportError {
	return &ImportError{Code: code, Message: msg, Status: status}
}

// CheckWorkbookBytes validates the container before any parser sees it: size,
// zip structure (entry count, total uncompressed size), no macros
func CheckWorkbookBytes(content []byte) *ImportError {
	if len(content) == 0 {
		return badFile("FILE_EMPTY", "The file is empty", 400)
	}
	if len(content) > MaxImportBytes {
		return badFile("FILE_TOO_LARGE", fmt.Sprintf("The file is larger than %d MB", MaxImportBytes>>20), 413)
	}
	if !bytes.HasPrefix(content, []byte("PK\x03\x04")) {
		return badFile("FILE_TYPE_NOT_ALLOWED", "Only .xlsx workbooks can be imported (.xls, .xlsm, .xlsb and CSV are not accepted)", 415)
	}
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return badFile("FILE_CORRUPT", "The file is not a readable .xlsx workbook", 400)
	}
	if len(zr.File) > maxZipEntries {
		return badFile("FILE_CORRUPT", "The workbook has too many parts", 400)
	}
	var total uint64
	hasWorkbook := false
	for _, zf := range zr.File {
		name := strings.ToLower(zf.Name)
		if strings.Contains(name, "vbaproject") || strings.HasSuffix(name, ".bin") && strings.Contains(name, "vba") {
			return badFile("MACROS_NOT_ALLOWED", "Workbooks with macros are not accepted", 415)
		}
		if name == "xl/workbook.xml" {
			hasWorkbook = true
		}
		total += zf.UncompressedSize64
		if zf.UncompressedSize64 > maxUnzipXMLPart*2 {
			return badFile("FILE_TOO_LARGE", "A part of the workbook is too large when unpacked", 413)
		}
	}
	if total > maxUnzipTotal {
		return badFile("FILE_TOO_LARGE", "The workbook is too large when unpacked", 413)
	}
	if !hasWorkbook {
		return badFile("FILE_TYPE_NOT_ALLOWED", "The file is not an Excel workbook (.xlsx)", 415)
	}
	for _, zf := range zr.File {
		if strings.EqualFold(zf.Name, "[Content_Types].xml") {
			rc, err := zf.Open()
			if err == nil {
				b := make([]byte, 0, 4096)
				buf := make([]byte, 4096)
				for len(b) < 1<<20 {
					n, rerr := rc.Read(buf)
					b = append(b, buf[:n]...)
					if rerr != nil {
						break
					}
				}
				rc.Close()
				if bytes.Contains(bytes.ToLower(b), []byte("macroenabled")) {
					return badFile("MACROS_NOT_ALLOWED", "Workbooks with macros are not accepted", 415)
				}
			}
		}
	}
	return nil
}

func cleanCell(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// ReadWorkbook reads every visible sheet as text. Cached values only: a formula
// cell yields the value the spreadsheet last computed, or "" when none.
func ReadWorkbook(content []byte) (*WorkbookData, error) {
	if e := CheckWorkbookBytes(content); e != nil {
		return nil, e
	}
	f, err := excelize.OpenReader(bytes.NewReader(content), excelize.Options{UnzipSizeLimit: maxUnzipTotal, UnzipXMLSizeLimit: maxUnzipXMLPart})
	if err != nil {
		return nil, badFile("FILE_CORRUPT", "The file could not be read as an .xlsx workbook", 400)
	}
	defer f.Close()
	names := f.GetSheetList()
	if len(names) > MaxImportSheets {
		return nil, badFile("TOO_MANY_SHEETS", fmt.Sprintf("The workbook has more than %d sheets", MaxImportSheets), 400)
	}
	wb := &WorkbookData{Format: FormatExternal, Meta: map[string]string{}}
	for _, name := range names {
		rows, err := f.GetRows(name, excelize.Options{RawCellValue: true})
		if err != nil {
			wb.Skipped = append(wb.Skipped, fmt.Sprintf("sheet %q could not be read", name))
			continue
		}
		if name == MetaSheet {
			for _, r := range rows {
				if len(r) >= 2 {
					wb.Meta[cleanCell(r[0])] = cleanCell(r[1])
				} else if len(r) == 1 {
					wb.Meta[cleanCell(r[0])] = ""
				}
			}
			continue
		}
		if visible, _ := f.GetSheetVisible(name); !visible {
			wb.Skipped = append(wb.Skipped, fmt.Sprintf("sheet %q is hidden and was skipped", name))
			continue
		}
		sd := SheetData{Name: name}
		if len(rows) > MaxImportRows {
			rows = rows[:MaxImportRows]
			sd.Truncated = true
		}
		for _, r := range rows {
			if len(r) > MaxImportCols {
				r = r[:MaxImportCols]
			}
			out := make([]string, len(r))
			for i, c := range r {
				c = cleanCell(c)
				if utf8.RuneCountInString(c) > maxCellRunes {
					c = string([]rune(c)[:maxCellRunes])
				}
				out[i] = c
			}
			sd.Rows = append(sd.Rows, out)
		}
		wb.Sheets = append(wb.Sheets, sd)
	}
	if _, ok := wb.Meta[WorkbookMarker]; ok {
		wb.Format = FormatAuditSphere
	}
	return wb, nil
}

// ---------------------------------------------------------------------------
// Header detection
// ---------------------------------------------------------------------------

func normHeader(s string) string {
	var b strings.Builder
	prevSpace := true
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
		} else if !prevSpace {
			b.WriteByte(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

var headerDict = map[string][]string{
	"score":          {"skor", "score", "nilai akhir", "nilai terbobot", "skor terbobot", "bobot x capaian", "bobot capaian", "weighted score"},
	"achievement":    {"capaian", "pencapaian", "achievement", "nilai", "hasil", "realisasi", "tingkat pemenuhan", "pemenuhan"},
	"weight":         {"bobot", "weight", "weighting"},
	"explanation":    {"penjelasan", "keterangan", "explanation", "catatan", "remarks", "alasan", "uraian penjelasan"},
	"recommendation": {"saran", "rekomendasi", "recommendation", "tindak lanjut", "suggestion", "recommendations"},
	"methods":        {"metode", "method", "metode evaluasi", "evaluasi", "dkwo", "d k w o"},
	"type":           {"tipe", "type", "jenis", "present function", "present fungsi"},
	"no":             {"no", "nomor", "no urut", "number", "kode", "code", "ref"},
	"description":    {"uraian", "deskripsi", "description", "indikator", "parameter", "pertanyaan", "unsur", "elemen", "element", "point of focus", "pof", "kriteria", "item"},
}

var classifyOrder = []string{"score", "achievement", "weight", "explanation", "recommendation", "methods", "type", "no", "description"}

func containsPhrase(h, phrase string) bool {
	return h == phrase || strings.HasPrefix(h, phrase+" ") || strings.HasSuffix(h, " "+phrase) || strings.Contains(h, " "+phrase+" ")
}

// classifyHeader maps a header cell to a column key ("" when unknown)
func classifyHeader(txt string, extra map[string][]string) string {
	h := normHeader(txt)
	if h == "" || len(h) > 60 {
		return ""
	}
	for _, key := range classifyOrder {
		for _, w := range headerDict[key] {
			if h == w {
				return key
			}
		}
		for _, w := range extra[key] {
			if h == normHeader(w) {
				return key
			}
		}
	}
	for _, key := range classifyOrder {
		for _, w := range headerDict[key] {
			if containsPhrase(h, w) {
				return key
			}
		}
		for _, w := range extra[key] {
			if containsPhrase(h, normHeader(w)) {
				return key
			}
		}
	}
	return ""
}

// DetectHeader finds the header row of a sheet: the first of the first 30 rows
// with at least 3 recognised headers, one of them an identifier (no or
// description) and one a value (achievement, score, explanation or weight)
func DetectHeader(rows [][]string, extra map[string][]string) (*ImportMapping, bool) {
	limit := headerScanRows
	if len(rows) < limit {
		limit = len(rows)
	}
	for r := 0; r < limit; r++ {
		cols := map[string]int{}
		headers := map[string]string{}
		for c, txt := range rows[r] {
			if key := classifyHeader(txt, extra); key != "" {
				if _, dup := cols[key]; !dup {
					cols[key] = c + 1
					headers[key] = normHeader(txt)
				}
			}
		}
		if len(cols) < 3 {
			continue
		}
		ident := cols["no"] > 0 || cols["description"] > 0
		value := cols["achievement"] > 0 || cols["score"] > 0 || cols["explanation"] > 0 || cols["weight"] > 0
		if ident && value {
			return &ImportMapping{HeaderRow: r + 1, Numbering: "no", Columns: cols, Headers: headers, HeaderWords: extra}, true
		}
	}
	return nil, false
}

// ProfileFit scores how well a saved mapping fits a sheet (0..1): the share of
// its recorded headers found at the same column of its header row
func ProfileFit(rows [][]string, m *ImportMapping) float64 {
	if m.HeaderRow < 1 || m.HeaderRow > len(rows) || len(m.Columns) == 0 {
		return 0
	}
	row := rows[m.HeaderRow-1]
	total, ok := 0, 0
	for key, col := range m.Columns {
		if col < 1 {
			continue
		}
		total++
		var cell string
		if col <= len(row) {
			cell = row[col-1]
		}
		if want := m.Headers[key]; want != "" {
			if normHeader(cell) == want {
				ok++
			}
		} else if classifyHeader(cell, m.HeaderWords) == key {
			ok++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(ok) / float64(total)
}

// AuditSphereMapping is the fixed layout of the exported workbook
func AuditSphereMapping() *ImportMapping {
	return &ImportMapping{HeaderRow: HeaderRow, Numbering: "no", Columns: map[string]int{
		"no": ColNo, "description": ColDescription, "type": ColType, "methods": ColMethods, "weight": ColWeight,
		"achievement": ColAchievement, "score": ColScore, "explanation": ColExplanation, "recommendation": ColRecommendation}}
}

// ---------------------------------------------------------------------------
// Row interpretation
// ---------------------------------------------------------------------------

// Levels of a numbered row
const (
	LevelComponent = "COMPONENT"
	LevelPrinciple = "PRINCIPLE"
	LevelPoF       = "POF"
	LevelElement   = "ELEMENT"
)

var numberingRe = regexp.MustCompile(`^\d+(\.\d+){0,3}$`)

// ParsedRow is one interpreted row of a sheet
type ParsedRow struct {
	Row            int      `json:"row"`
	Sheet          string   `json:"sheet"`
	No             string   `json:"no"`
	Description    string   `json:"description"`
	Level          string   `json:"level"`
	Block          string   `json:"block,omitempty"`
	ContextPoF     string   `json:"-"`
	Methods        string   `json:"-"`
	FileWeight     *float64 `json:"-"`
	Achievement    *float64 `json:"achievement"`
	Score          *float64 `json:"-"`
	Explanation    string   `json:"explanation"`
	Recommendation string   `json:"recommendation"`
	Issues         []string `json:"issues"`
	HasValue       bool     `json:"-"`
	ErrorCell      bool     `json:"-"`
}

// normNumbering returns the numbering of a cell ("" when it is not one)
func normNumbering(s string, trusted bool) string {
	s = strings.TrimSpace(s)
	if trusted {
		s = UnsafeCell(s)
	}
	s = strings.TrimRight(s, ". ")
	s = strings.ReplaceAll(s, " ", "")
	if numberingRe.MatchString(s) {
		return s
	}
	return ""
}

func levelOfNumbering(no string) string {
	switch strings.Count(no, ".") {
	case 0:
		return LevelComponent
	case 1:
		return LevelPrinciple
	case 2:
		return LevelPoF
	default:
		return LevelElement
	}
}

// ParseNumber reads a number from text: decimal comma or point, percent.
// ok is false for anything else. A value above 1 without a percent sign is
// reported through tooBig (it may be a percentage or a 0-100 scale).
func ParseNumber(s string) (v float64, ok bool, tooBig bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
	if s == "" {
		return 0, false, false
	}
	pct := strings.HasSuffix(s, "%")
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	if strings.Count(s, ",") == 1 && !strings.Contains(s, ".") {
		s = strings.Replace(s, ",", ".", 1)
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false, false
	}
	if pct {
		n /= 100
	} else if n > 1 {
		return n, true, true
	}
	return n, true, false
}

func isErrorCell(s string) bool {
	switch strings.ToUpper(s) {
	case "#N/A", "#VALUE!", "#REF!", "#DIV/0!", "#NAME?", "#NUM!", "#NULL!":
		return true
	}
	return false
}

func col(row []string, c int) string {
	if c < 1 || c > len(row) {
		return ""
	}
	return row[c-1]
}

func blockOf(s string) string {
	switch normHeader(s) {
	case "present", "ada", "keberadaan":
		return models.CosoElementPresentName
	case "function", "fungsi", "berfungsi":
		return models.CosoElementFunctionName
	}
	return ""
}

// ParseSheet interprets the rows below the header row of a sheet
func ParseSheet(sd SheetData, m *ImportMapping, trusted bool) []ParsedRow {
	var out []ParsedRow
	block := ""
	pofCtx := ""
	for r := m.HeaderRow; r < len(sd.Rows); r++ {
		row := sd.Rows[r]
		no := normNumbering(col(row, m.Columns["no"]), trusted)
		desc := col(row, m.Columns["description"])
		if trusted {
			desc = UnsafeCell(desc)
		}
		if no == "" && desc != "" {
			if b := blockOf(desc); b != "" && col(row, m.Columns["achievement"]) == "" {
				block = b
				continue
			}
		}
		pr := ParsedRow{Row: r + 1, Sheet: sd.Name, No: no, Description: desc, Issues: []string{}}
		if no != "" {
			pr.Level = levelOfNumbering(no)
		}
		if tcell := col(row, m.Columns["type"]); tcell != "" {
			if b := blockOf(tcell); b != "" {
				block = b
			}
		}
		pr.Block = block
		if pr.Level == LevelPoF {
			pofCtx = no
			block = ""
		}
		if pr.Level == LevelPrinciple || pr.Level == LevelComponent {
			pofCtx, block = "", ""
		}
		pr.ContextPoF = pofCtx
		pr.Methods = col(row, m.Columns["methods"])
		if w := col(row, m.Columns["weight"]); w != "" {
			if v, ok, big := ParseNumber(w); ok && !big || ok && big {
				pr.FileWeight = &v
			}
		}
		for _, k := range []string{"achievement", "score"} {
			txt := col(row, m.Columns[k])
			if txt == "" {
				continue
			}
			if isErrorCell(txt) {
				pr.ErrorCell = true
				pr.Issues = append(pr.Issues, fmt.Sprintf("%s cell holds the error %s; no value taken", k, txt))
				continue
			}
			v, ok, big := ParseNumber(txt)
			switch {
			case !ok:
				if !strings.EqualFold(txt, "n/a") && txt != "-" {
					pr.Issues = append(pr.Issues, fmt.Sprintf("%s %q is not a number", k, truncate(txt, 40)))
				}
			case big && k == "achievement":
				pr.Issues = append(pr.Issues, fmt.Sprintf("achievement %v is above 1 (percent? write 75%% or 0.75); no value taken", v))
			case big && k == "score":
				pr.Score = &v
			case k == "achievement":
				pr.Achievement = &v
			default:
				pr.Score = &v
			}
		}
		pr.Explanation = col(row, m.Columns["explanation"])
		pr.Recommendation = col(row, m.Columns["recommendation"])
		if trusted {
			pr.Explanation, pr.Recommendation = UnsafeCell(pr.Explanation), UnsafeCell(pr.Recommendation)
		}
		pr.HasValue = pr.Achievement != nil || pr.Score != nil || pr.Explanation != "" || pr.Recommendation != ""
		if no == "" && desc == "" && !pr.HasValue {
			continue
		}
		out = append(out, pr)
	}
	return out
}

// ---------------------------------------------------------------------------
// Text similarity
// ---------------------------------------------------------------------------

var stopwords = map[string]bool{}

func init() {
	for _, w := range strings.Fields("yang dan di ke dari untuk dengan atau pada adalah telah sudah oleh dalam serta ini itu the a an of and to in is are has have for by on that as with be or at it its") {
		stopwords[w] = true
	}
}

func tokens(s string) map[string]bool {
	out := map[string]bool{}
	for _, t := range strings.Fields(normHeader(s)) {
		if len(t) >= 2 && !stopwords[t] {
			out[t] = true
		}
	}
	return out
}

// Similarity is the Jaccard similarity of the token sets of two texts (0..1)
func Similarity(a, b string) float64 {
	ta, tb := tokens(a), tokens(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	inter := 0
	for t := range ta {
		if tb[t] {
			inter++
		}
	}
	union := len(ta) + len(tb) - inter
	return float64(inter) / float64(union)
}

// ---------------------------------------------------------------------------
// Matching and planning
// ---------------------------------------------------------------------------

// Match kinds and actions
const (
	MatchCode = "CODE"
	MatchText = "TEXT"
	MatchNone = "NONE"

	RowUpdate    = "UPDATE"
	RowSame      = "SAME"
	RowSkip      = "SKIP"
	RowConflict  = "CONFLICT"
	RowUnmatched = "UNMATCHED"
	RowInvalid   = "INVALID"
	RowAggregate = "AGGREGATE"
)

// RowTarget is the matched PoF or element
type RowTarget struct {
	Code  string `json:"code"`
	Level string `json:"level"`
}

// RowCurrent is the value the assessment has now
type RowCurrent struct {
	Achievement    *float64 `json:"achievement"`
	Explanation    string   `json:"explanation"`
	Recommendation string   `json:"recommendation"`
}

// TextCandidate is the best text match of a conflicting row
type TextCandidate struct {
	Code       string  `json:"code"`
	Similarity float64 `json:"similarity"`
}

// PlanRow is one row of the preview
type PlanRow struct {
	Row            int            `json:"row"`
	Sheet          string         `json:"sheet"`
	No             string         `json:"no"`
	Description    string         `json:"description"`
	Level          string         `json:"level"`
	Block          string         `json:"block,omitempty"`
	Target         *RowTarget     `json:"target"`
	Match          string         `json:"match"`
	Similarity     float64        `json:"similarity"`
	Achievement    *float64       `json:"achievement"`
	Explanation    string         `json:"explanation"`
	Recommendation string         `json:"recommendation"`
	Current        *RowCurrent    `json:"current"`
	Action         string         `json:"action"`
	Overwrites     bool           `json:"overwrites"`
	Issues         []string       `json:"issues"`
	Candidate      *TextCandidate `json:"text_candidate,omitempty"`
	// ConflictKind: TEXT (the number exists, the description differs; the user
	// may accept it) or ELEMENTS_HAVE_VALUES (cannot be applied)
	ConflictKind string `json:"conflict_kind,omitempty"`
}

// UnmatchedRow is a row that was not matched to the assessment
type UnmatchedRow struct {
	Row         int    `json:"row"`
	Sheet       string `json:"sheet"`
	No          string `json:"no"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
}

// PlanSummary counts the rows of a preview
type PlanSummary struct {
	RowsRead         int `json:"rows_read"`
	RowsWithNumber   int `json:"rows_with_number"`
	MatchedByCode    int `json:"matched_by_code"`
	MatchedByText    int `json:"matched_by_text"`
	Unmatched        int `json:"unmatched"`
	NoValue          int `json:"no_value"`
	WillUpdate       int `json:"will_update"`
	Overwrites       int `json:"overwrites"`
	Conflicts        int `json:"conflicts"`
	Invalid          int `json:"invalid"`
	WeightMismatches int `json:"weight_mismatches"`
}

// Plan is the interpretation of a workbook against an assessment
type Plan struct {
	Rows      []PlanRow      `json:"rows"`
	Unmatched []UnmatchedRow `json:"unmatched"`
	Summary   PlanSummary    `json:"summary"`
	Warnings  []string       `json:"warnings"`
}

type target struct {
	code, level string
	texts       []string
	typ         string
	weight      float64
	scale       string
	pof         *models.ICAssessmentPoF
	el          *models.ICAssessmentElement
}

func (t *target) similarity(desc string) float64 {
	best := 0.0
	for _, x := range t.texts {
		if s := Similarity(desc, x); s > best {
			best = s
		}
	}
	return best
}

func trimF(v *float64) *float64 {
	if v == nil {
		return nil
	}
	x := round(*v, 6)
	return &x
}

// BuildPlan matches parsed rows to the assessment's PoFs and elements.
// trusted (AuditSphere format) matches by code only and skips the conflict test.
func BuildPlan(parsed []ParsedRow, pofs []models.ICAssessmentPoF, els []models.ICAssessmentElement, trusted bool) *Plan {
	plan := &Plan{Rows: []PlanRow{}, Unmatched: []UnmatchedRow{}, Warnings: []string{}}
	pofBy := map[string]*target{}
	elBy := map[string]*target{}
	elsOfPoF := map[string][]*target{}
	var allPoF, allEl []*target
	elementsWithValues := map[string]bool{}
	for i := range pofs {
		p := &pofs[i]
		t := &target{code: p.PofCode, level: LevelPoF, texts: []string{p.NameEn, p.NameID}, weight: p.Weight, pof: p}
		pofBy[p.PofCode] = t
		allPoF = append(allPoF, t)
	}
	for i := range els {
		e := &els[i]
		t := &target{code: e.ElementCode, level: LevelElement, texts: []string{e.DescriptionEn, e.DescriptionID}, typ: e.Type, weight: e.Weight, scale: e.ScaleType, el: e}
		elBy[e.ElementCode] = t
		elsOfPoF[e.PofCode] = append(elsOfPoF[e.PofCode], t)
		allEl = append(allEl, t)
		if e.Achievement != nil {
			elementsWithValues[e.PofCode] = true
		}
	}

	bestText := func(desc string, cands []*target, block string) (*target, float64, bool) {
		var best, second *target
		var bs, ss float64
		for _, c := range cands {
			if block != "" && c.typ != "" && c.typ != block {
				continue
			}
			s := c.similarity(desc)
			switch {
			case s > bs:
				second, ss = best, bs
				best, bs = c, s
			case s > ss:
				second, ss = c, s
			}
		}
		_ = second
		if best == nil || bs < TextMatchMin {
			return nil, bs, false
		}
		if ss >= bs-textAmbiguityGap && ss >= TextMatchMin {
			return nil, bs, true // ambiguous
		}
		return best, bs, false
	}

	fileValueRowsOfPoF := map[string]bool{} // element rows with a value, per PoF code
	type pending struct {
		pr  ParsedRow
		row PlanRow
		t   *target
	}
	var rows []*pending

	for _, pr := range parsed {
		plan.Summary.RowsRead++
		if pr.No != "" {
			plan.Summary.RowsWithNumber++
		}
		row := PlanRow{Row: pr.Row, Sheet: pr.Sheet, No: pr.No, Description: pr.Description, Level: pr.Level, Block: pr.Block,
			Achievement: trimF(pr.Achievement), Explanation: pr.Explanation, Recommendation: pr.Recommendation, Match: MatchNone, Issues: append([]string{}, pr.Issues...)}
		if row.Achievement == nil && pr.Score != nil {
			// derive from the score when the file gives no achievement
			w := 0.0
			if pr.FileWeight != nil {
				w = *pr.FileWeight
			}
			if pr.No != "" {
				if t := pofBy[pr.No]; t != nil && w == 0 {
					w = t.weight
				} else if t := elBy[pr.No]; t != nil && w == 0 {
					w = t.weight
				}
			}
			if w > 0 {
				v := round(*pr.Score/w, 3)
				if v >= 0 && v <= 1 {
					row.Achievement = &v
					row.Issues = append(row.Issues, fmt.Sprintf("achievement derived from score %v / weight %v", *pr.Score, w))
				}
			}
		}
		hasValue := row.Achievement != nil || row.Explanation != "" || row.Recommendation != ""

		// aggregate rows (component, principle): not a target
		if pr.Level == LevelComponent || pr.Level == LevelPrinciple {
			row.Action = RowAggregate
			continue
		}
		var t *target
		switch {
		case pr.No != "" && (pr.Level == LevelPoF || pr.Level == LevelElement):
			if pr.Level == LevelPoF {
				t = pofBy[pr.No]
			} else {
				t = elBy[pr.No]
			}
			if t != nil {
				row.Match, row.Similarity = MatchCode, 1
				if !trusted && pr.Description != "" {
					sim := t.similarity(pr.Description)
					row.Similarity = round(sim, 3)
					if sim < TextConflictMax {
						row.Action = RowConflict
						row.ConflictKind = "TEXT"
						row.Issues = append(row.Issues, fmt.Sprintf("the number %s exists but its description differs from the catalogue (similarity %.2f)", pr.No, sim))
						cands := allEl
						if pr.Level == LevelPoF {
							cands = allPoF
						}
						if alt, as, _ := bestText(pr.Description, cands, ""); alt != nil && alt.code != pr.No {
							row.Candidate = &TextCandidate{Code: alt.code, Similarity: round(as, 3)}
						}
					}
				}
			}
		}
		if t == nil && pr.Description != "" && hasValue && pr.Level != "" && pr.No != "" && trusted {
			// trusted workbooks carry exact codes: an unknown one is reported
		} else if t == nil && pr.Description != "" && hasValue && !trusted {
			var cands []*target
			switch {
			case pr.Level == LevelPoF:
				cands = allPoF
			case pr.Level == LevelElement:
				cands = allEl
			case pr.ContextPoF != "" && len(elsOfPoF[pr.ContextPoF]) > 0:
				cands = elsOfPoF[pr.ContextPoF]
			default:
				cands = append(append([]*target{}, allEl...), allPoF...)
			}
			cand, sim, ambiguous := bestText(pr.Description, cands, pr.Block)
			if cand == nil && pr.ContextPoF != "" && pr.Level == "" {
				// the context PoF may be numbered differently: widen to every element and PoF
				cand, sim, ambiguous = bestText(pr.Description, append(append([]*target{}, allEl...), allPoF...), pr.Block)
			}
			if cand != nil {
				t = cand
				row.Match, row.Similarity = MatchText, round(sim, 3)
			} else {
				reason := fmt.Sprintf("no catalogue code and no text match at or above %.2f", TextMatchMin)
				if ambiguous {
					reason = "the description matches more than one catalogue item equally well"
				}
				if pr.No != "" {
					reason = fmt.Sprintf("the number %s is not in the assessment and %s", pr.No, strings.TrimPrefix(reason, "no catalogue code and "))
				}
				plan.Unmatched = append(plan.Unmatched, UnmatchedRow{Row: pr.Row, Sheet: pr.Sheet, No: pr.No, Description: truncate(pr.Description, 200), Reason: reason})
				row.Action = RowUnmatched
			}
		}
		if t == nil && row.Action == "" {
			switch {
			case pr.No != "" && hasValue:
				plan.Unmatched = append(plan.Unmatched, UnmatchedRow{Row: pr.Row, Sheet: pr.Sheet, No: pr.No, Description: truncate(pr.Description, 200),
					Reason: "the number is not a point of focus or element of this assessment"})
				row.Action = RowUnmatched
			case pr.No == "" && pr.Description == "" && hasValue:
				plan.Unmatched = append(plan.Unmatched, UnmatchedRow{Row: pr.Row, Sheet: pr.Sheet, Reason: "the row has a value but neither a number nor a description"})
				row.Action = RowUnmatched
			default:
				// a heading, a note or an empty target row: nothing to import
				continue
			}
		}
		if t != nil {
			row.Target = &RowTarget{Code: t.code, Level: t.level}
			row.Level = t.level
			if pr.FileWeight != nil && t.weight > 0 && math.Abs(*pr.FileWeight-t.weight) > 0.0005 && math.Abs(*pr.FileWeight/100-t.weight) > 0.0005 {
				plan.Summary.WeightMismatches++
				row.Issues = append(row.Issues, fmt.Sprintf("the file weight %v differs from the catalogue weight %v (weights come from the catalogue)", round4(*pr.FileWeight), round4(t.weight)))
			}
			if t.level == LevelElement && row.Achievement != nil {
				fileValueRowsOfPoF[t.el.PofCode] = true
			}
		}
		rows = append(rows, &pending{pr: pr, row: row, t: t})
	}

	for _, p := range rows {
		row := p.row
		t := p.t
		if row.Action == RowUnmatched {
			plan.Summary.Unmatched++
			plan.Rows = append(plan.Rows, row)
			continue
		}
		hasValue := row.Achievement != nil || row.Explanation != "" || row.Recommendation != ""
		if row.Match == MatchCode {
			plan.Summary.MatchedByCode++
		} else if row.Match == MatchText {
			plan.Summary.MatchedByText++
		}
		cur := &RowCurrent{}
		if t.level == LevelPoF {
			cur.Achievement, cur.Explanation, cur.Recommendation = t.pof.Achievement, t.pof.Explanation, t.pof.Recommendation
		} else {
			cur.Achievement, cur.Explanation, cur.Recommendation = t.el.Achievement, t.el.Explanation, t.el.Recommendation
		}
		row.Current = cur
		if row.Action == RowConflict {
			plan.Summary.Conflicts++
			plan.Rows = append(plan.Rows, row)
			continue
		}
		if !hasValue {
			row.Action = RowSkip
			plan.Summary.NoValue++
			plan.Rows = append(plan.Rows, row)
			continue
		}
		if row.Achievement != nil {
			var err error
			if t.level == LevelPoF {
				err = ValidateAchievement(*row.Achievement, "", true)
			} else {
				err = ValidateAchievement(*row.Achievement, t.scale, false)
			}
			if err != nil {
				row.Action = RowInvalid
				row.Issues = append(row.Issues, fmt.Sprintf("achievement %v: %s", *row.Achievement, err.Error()))
				plan.Summary.Invalid++
				plan.Rows = append(plan.Rows, row)
				continue
			}
		}
		if t.level == LevelPoF && row.Achievement != nil {
			if fileValueRowsOfPoF[t.code] {
				row.Action = RowSkip
				row.Issues = append(row.Issues, "the PoF value is derived from its elements in this file")
				plan.Summary.NoValue++
				plan.Rows = append(plan.Rows, row)
				continue
			}
			if elementsWithValues[t.code] {
				row.Action = RowConflict
				row.ConflictKind = "ELEMENTS_HAVE_VALUES"
				row.Issues = append(row.Issues, "the elements of this PoF have values in the assessment, so a PoF-level value cannot be set")
				plan.Summary.Conflicts++
				plan.Rows = append(plan.Rows, row)
				continue
			}
		}
		changes, overwrites := diffValues(cur, row)
		switch {
		case !changes:
			row.Action = RowSame
		default:
			row.Action = RowUpdate
			row.Overwrites = overwrites
			plan.Summary.WillUpdate++
			if overwrites {
				plan.Summary.Overwrites++
			}
		}
		plan.Rows = append(plan.Rows, row)
	}
	sort.SliceStable(plan.Unmatched, func(i, j int) bool { return plan.Unmatched[i].Row < plan.Unmatched[j].Row })
	return plan
}

// diffValues compares a row with the current value: whether anything differs,
// and whether a non-empty current value would be replaced
func diffValues(cur *RowCurrent, row PlanRow) (changes, overwrites bool) {
	if row.Achievement != nil && !floatEq(cur.Achievement, row.Achievement) {
		changes = true
		if cur.Achievement != nil {
			overwrites = true
		}
	}
	if row.Explanation != "" && row.Explanation != cur.Explanation {
		changes = true
		if cur.Explanation != "" {
			overwrites = true
		}
	}
	if row.Recommendation != "" && row.Recommendation != cur.Recommendation {
		changes = true
		if cur.Recommendation != "" {
			overwrites = true
		}
	}
	return
}

// MappingJSON / ParseMappingJSON store a mapping as text
func MappingJSON(m *ImportMapping) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// ParseMappingJSON decodes a stored mapping
func ParseMappingJSON(s string) (*ImportMapping, error) {
	var m ImportMapping
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	if m.Columns == nil {
		m.Columns = map[string]int{}
	}
	return &m, nil
}
