package ic_assessment

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"audit-service/models"

	"github.com/xuri/excelize/v2"
)

// Excel workbook of a QUANTITATIVE_POF assessment (docs/coso-pof-assessment-api.md
// section 7). Layout is shared by export and by the import of the AuditSphere
// format, so a file exported here is read back by the exact layout.
//
// Safety: every text is written with SetCellStr (a string cell, never a
// formula) and a text that starts with = + - @ tab or CR is prefixed with a
// single quote (anti formula injection). Scores are written as values.

// Workbook format markers
const (
	WorkbookMarker        = "AUDITSPHERE_COSO_POF"
	WorkbookFormatVersion = "1"
	MetaSheet             = "_meta"
	// HeaderRow is the row of the column headers on a component sheet
	HeaderRow = 4
)

// Column positions (1-based) of a component sheet
const (
	ColNo = iota + 1
	ColDescription
	ColType
	ColMethods
	ColWeight
	ColAchievement
	ColScore
	ColExplanation
	ColRecommendation
)

type langText struct{ id, en string }

var xlText = map[string]langText{
	"no": {"No", "No"}, "desc": {"Uraian", "Description"}, "type": {"Tipe", "Type"}, "methods": {"Metode (D/K/W/O)", "Method (D/K/W/O)"},
	"weight": {"Bobot", "Weight"}, "ach": {"Capaian", "Achievement"}, "score": {"Skor", "Score"},
	"expl": {"Penjelasan", "Explanation"}, "rec": {"Saran", "Recommendation"}, "status": {"Status", "Status"},
	"summary": {"Ringkasan", "Summary"}, "total": {"Total", "Total"}, "present": {"Present", "Present"}, "function": {"Function", "Function"},
	"component": {"Komponen", "Component"}, "title": {"Penilaian Pengendalian Intern COSO 2013 (Point of Focus)", "COSO 2013 internal control assessment (points of focus)"},
	"dept": {"Unit", "Unit"}, "period": {"Periode", "Period"}, "catalog": {"Versi katalog", "Catalogue version"}, "note": {"Catatan", "Note"},
}

func tr(lang, key string) string {
	t := xlText[key]
	if lang == "en" {
		return t.en
	}
	return t.id
}

// SafeCell neutralises a text that a spreadsheet could evaluate as a formula
func SafeCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

// UnsafeCell undoes SafeCell on import of the AuditSphere format
func UnsafeCell(s string) string {
	if len(s) >= 2 && s[0] == '\'' {
		switch s[1] {
		case '=', '+', '-', '@', '\t', '\r':
			return s[1:]
		}
	}
	return s
}

// WorkbookMeta describes where a workbook came from
type WorkbookMeta struct {
	AssessmentID   string
	Title          string // unit line
	Period         string
	CatalogVersion string
	Lang           string
	Blank          bool
}

type exportStyles struct{ title, head, bold, wrap, num, group, tot int }

func newStyles(f *excelize.File) (exportStyles, error) {
	var st exportStyles
	var err error
	mk := func(s *excelize.Style) int {
		if err != nil {
			return 0
		}
		var id int
		id, err = f.NewStyle(s)
		return id
	}
	st.title = mk(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	st.head = mk(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	st.bold = mk(&excelize.Style{Font: &excelize.Font{Bold: true}, Alignment: &excelize.Alignment{Vertical: "top", WrapText: true}})
	st.wrap = mk(&excelize.Style{Alignment: &excelize.Alignment{Vertical: "top", WrapText: true}})
	st.num = mk(&excelize.Style{NumFmt: 2, Alignment: &excelize.Alignment{Vertical: "top"}}) // 0.00
	st.group = mk(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"DDEBF7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true}})
	st.tot = mk(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"FFE699"}, Pattern: 1}})
	return st, err
}

func cell(col, row int) string {
	n, _ := excelize.CoordinatesToCellName(col, row)
	return n
}

func setStr(f *excelize.File, sheet string, col, row int, v string, style int) {
	c := cell(col, row)
	_ = f.SetCellStr(sheet, c, SafeCell(v))
	if style != 0 {
		_ = f.SetCellStyle(sheet, c, c, style)
	}
}

func setNum(f *excelize.File, sheet string, col, row int, v float64, style int) {
	c := cell(col, row)
	_ = f.SetCellFloat(sheet, c, v, -1, 64)
	if style != 0 {
		_ = f.SetCellStyle(sheet, c, c, style)
	}
}

func fptrNum(f *excelize.File, sheet string, col, row int, v *float64, style int) {
	if v != nil {
		setNum(f, sheet, col, row, *v, style)
	}
}

func numberingOf(code string) string { return code }

// componentNumber returns 1..5 for a component code
func componentNumber(code string) int { return componentOrder(code) + 1 }

// principleNumbering returns "1.1" for P01 etc.
func principleNumbering(p Principle) string {
	idx := 0
	for _, q := range Principles {
		if q.ComponentCode == p.ComponentCode {
			idx++
		}
		if q.Code == p.Code {
			break
		}
	}
	return fmt.Sprintf("%d.%d", componentNumber(p.ComponentCode), idx)
}

// BuildWorkbook writes the workbook for the given rows. With meta.Blank the
// values (achievement, explanation, recommendation, score) are left empty.
func BuildWorkbook(meta WorkbookMeta, pofs []models.ICAssessmentPoF, els []models.ICAssessmentElement, th Thresholds) ([]byte, error) {
	lang := meta.Lang
	if lang != "en" {
		lang = "id"
	}
	f := excelize.NewFile()
	defer f.Close()
	st, err := newStyles(f)
	if err != nil {
		return nil, err
	}
	rows, rowEls := pofs, els
	if meta.Blank {
		rows = make([]models.ICAssessmentPoF, len(pofs))
		copy(rows, pofs)
		rowEls = make([]models.ICAssessmentElement, len(els))
		copy(rowEls, els)
		for i := range rows {
			rows[i].Achievement, rows[i].Explanation, rows[i].Recommendation = nil, "", ""
		}
		for i := range rowEls {
			rowEls[i].Achievement, rowEls[i].Explanation, rowEls[i].Recommendation = nil, "", ""
		}
	}
	tree := Calculate(rows, rowEls, th)

	// Summary sheet
	sum := tr(lang, "summary")
	_ = f.SetSheetName("Sheet1", sum)
	setStr(f, sum, 1, 1, tr(lang, "title"), st.title)
	setStr(f, sum, 1, 2, tr(lang, "dept")+": "+meta.Title+"   "+tr(lang, "period")+": "+meta.Period+"   "+tr(lang, "catalog")+": "+meta.CatalogVersion, 0)
	heads := []string{tr(lang, "no"), tr(lang, "component"), tr(lang, "weight"), tr(lang, "score"), tr(lang, "ach"), tr(lang, "status")}
	for i, h := range heads {
		setStr(f, sum, i+1, HeaderRow, h, st.head)
	}
	r := HeaderRow + 1
	pct := func(v *float64) *float64 {
		if v == nil {
			return nil
		}
		x := *v
		return &x
	}
	for _, cc := range tree.Components {
		setStr(f, sum, 1, r, fmt.Sprint(componentNumber(cc.Code)), st.group)
		setStr(f, sum, 2, r, ComponentNames[cc.Code], st.group)
		setNum(f, sum, 3, r, cc.Weight, st.group)
		fptrNum(f, sum, 4, r, cc.Score, st.num)
		fptrNum(f, sum, 5, r, pct(cc.Achievement), st.num)
		setStr(f, sum, 6, r, statusLabel(cc.Status), st.group)
		r++
		for _, pr := range cc.Principles {
			setStr(f, sum, 1, r, principleNumbering(pr.Principle), 0)
			setStr(f, sum, 2, r, pr.Principle.Code+" - "+pr.Principle.NameEN, 0)
			setNum(f, sum, 3, r, pr.Weight, st.num)
			fptrNum(f, sum, 4, r, pr.Score, st.num)
			fptrNum(f, sum, 5, r, pr.Achievement, st.num)
			setStr(f, sum, 6, r, statusLabel(pr.Status), 0)
			r++
		}
	}
	setStr(f, sum, 2, r, tr(lang, "total"), st.tot)
	setNum(f, sum, 3, r, tree.Total.Weight, st.tot)
	fptrNum(f, sum, 4, r, tree.Total.Score, st.tot)
	fptrNum(f, sum, 5, r, tree.Total.Achievement, st.tot)
	setStr(f, sum, 6, r, statusLabel(tree.Total.Status), st.tot)
	_ = f.SetColWidth(sum, "A", "A", 8)
	_ = f.SetColWidth(sum, "B", "B", 60)
	_ = f.SetColWidth(sum, "C", "F", 14)

	// One sheet per component
	for _, cc := range tree.Components {
		sh := cc.Code
		if _, err := f.NewSheet(sh); err != nil {
			return nil, err
		}
		setStr(f, sh, 1, 1, tr(lang, "title")+" - "+ComponentNames[cc.Code], st.title)
		setStr(f, sh, 1, 2, tr(lang, "dept")+": "+meta.Title+"   "+tr(lang, "period")+": "+meta.Period, 0)
		hs := map[int]string{ColNo: tr(lang, "no"), ColDescription: tr(lang, "desc"), ColType: tr(lang, "type"), ColMethods: tr(lang, "methods"),
			ColWeight: tr(lang, "weight"), ColAchievement: tr(lang, "ach"), ColScore: tr(lang, "score"), ColExplanation: tr(lang, "expl"), ColRecommendation: tr(lang, "rec")}
		for c, h := range hs {
			setStr(f, sh, c, HeaderRow, h, st.head)
		}
		r := HeaderRow + 1
		setStr(f, sh, ColNo, r, fmt.Sprint(componentNumber(cc.Code)), st.group)
		setStr(f, sh, ColDescription, r, ComponentNames[cc.Code], st.group)
		setNum(f, sh, ColWeight, r, cc.Weight, st.group)
		fptrNum(f, sh, ColAchievement, r, cc.Achievement, st.num)
		fptrNum(f, sh, ColScore, r, cc.Score, st.num)
		r++
		for _, pr := range cc.Principles {
			setStr(f, sh, ColNo, r, principleNumbering(pr.Principle), st.group)
			setStr(f, sh, ColDescription, r, pr.Principle.Code+" - "+pr.Principle.NameEN, st.group)
			setNum(f, sh, ColWeight, r, pr.Weight, st.group)
			fptrNum(f, sh, ColAchievement, r, pr.Achievement, st.num)
			fptrNum(f, sh, ColScore, r, pr.Score, st.num)
			r++
			for _, pc := range pr.Pofs {
				p := pc.Row
				setStr(f, sh, ColNo, r, numberingOf(p.PofCode), st.bold)
				setStr(f, sh, ColDescription, r, descFor(lang, p.NameEn, p.NameID), st.bold)
				setStr(f, sh, ColMethods, r, p.Methods, st.wrap)
				setNum(f, sh, ColWeight, r, round4(p.Weight), st.num)
				if pc.Achievement != nil {
					setNum(f, sh, ColAchievement, r, round4(*pc.Achievement), st.num)
					setNum(f, sh, ColScore, r, round4(*pc.Score), st.num)
				}
				setStr(f, sh, ColExplanation, r, p.Explanation, st.wrap)
				setStr(f, sh, ColRecommendation, r, p.Recommendation, st.wrap)
				r++
				for _, e := range pc.Elements {
					setStr(f, sh, ColNo, r, e.ElementCode, st.wrap)
					setStr(f, sh, ColDescription, r, descFor(lang, e.DescriptionEn, e.DescriptionID), st.wrap)
					setStr(f, sh, ColType, r, typeLabel(lang, e.Type), st.wrap)
					m := e.Methods
					if m == "" {
						m = e.DefaultMethods
					}
					setStr(f, sh, ColMethods, r, m, st.wrap)
					setNum(f, sh, ColWeight, r, round4(e.Weight), st.num)
					if e.Achievement != nil {
						setNum(f, sh, ColAchievement, r, *e.Achievement, st.num)
						setNum(f, sh, ColScore, r, round4(e.Weight**e.Achievement), st.num)
					}
					setStr(f, sh, ColExplanation, r, e.Explanation, st.wrap)
					setStr(f, sh, ColRecommendation, r, e.Recommendation, st.wrap)
					r++
				}
			}
		}
		_ = f.SetColWidth(sh, "A", "A", 11)
		_ = f.SetColWidth(sh, "B", "B", 70)
		_ = f.SetColWidth(sh, "C", "D", 14)
		_ = f.SetColWidth(sh, "E", "G", 10)
		_ = f.SetColWidth(sh, "H", "I", 40)
		_ = f.SetPanes(sh, &excelize.Panes{Freeze: true, Split: true, XSplit: 2, YSplit: HeaderRow, TopLeftCell: cell(3, HeaderRow+1), ActivePane: "bottomRight"})
	}

	// Meta sheet (hidden): the marker the importer looks for
	if _, err := f.NewSheet(MetaSheet); err != nil {
		return nil, err
	}
	metaRows := [][2]string{{WorkbookMarker, WorkbookFormatVersion}, {"catalog_version", meta.CatalogVersion}, {"assessment_id", meta.AssessmentID},
		{"lang", lang}, {"blank", fmt.Sprint(meta.Blank)}, {"exported_at", time.Now().UTC().Format(time.RFC3339)}}
	for i, m := range metaRows {
		_ = f.SetCellStr(MetaSheet, cell(1, i+1), m[0])
		_ = f.SetCellStr(MetaSheet, cell(2, i+1), m[1])
	}
	_ = f.SetSheetVisible(MetaSheet, false)
	idx, _ := f.GetSheetIndex(sum)
	f.SetActiveSheet(idx)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func descFor(lang, en, id string) string {
	if lang == "en" {
		return en
	}
	if id != "" {
		return id
	}
	return en
}

func typeLabel(lang, t string) string {
	if t == models.CosoElementPresentName {
		return tr(lang, "present")
	}
	return tr(lang, "function")
}

func statusLabel(s string) string {
	return strings.ReplaceAll(s, "_", " ")
}
