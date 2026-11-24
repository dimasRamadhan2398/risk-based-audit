package docxbuilder

import (
	"fmt"
	"strings"

	"audit-service/models"
)

// rootCauseCategories are the RCA categories in report order. Unknown methods
// are appended after them, unchanged.
var rootCauseCategories = []string{"People", "Process", "Policy", "System", "Eksternal"}

// CategoryLabelID maps a finding category to its Indonesian document label.
// The app uses Very Significant / Significant / Quite Significant / Not
// Significant; unknown or legacy values (e.g. "Major") are returned unchanged
// (trimmed), and an empty value yields "-".
func CategoryLabelID(category string) string {
	trimmed := strings.TrimSpace(category)
	if trimmed == "" {
		return "-"
	}
	key := strings.ToLower(strings.Join(strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(trimmed)), " "))
	switch key {
	case "very significant":
		return "Sangat Signifikan"
	case "significant":
		return "Signifikan"
	case "quite significant":
		return "Cukup Signifikan"
	case "not significant":
		return "Tidak Signifikan"
	}
	return trimmed
}

// tableCell renders one bordered cell; each line of text becomes a paragraph.
// Empty text renders "-". vMerge is "", "restart" or "continue".
func tableCell(width int, text string, bold bool, shade string, vMerge string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>`, width))
	switch vMerge {
	case "restart":
		sb.WriteString(`<w:vMerge w:val="restart"/>`)
	case "continue":
		sb.WriteString(`<w:vMerge/>`)
	}
	if shade != "" {
		sb.WriteString(fmt.Sprintf(`<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, shade))
	}
	sb.WriteString(`</w:tcPr>`)

	rpr := `<w:rPr><w:sz w:val="18"/></w:rPr>`
	if bold {
		rpr = `<w:rPr><w:b/><w:sz w:val="18"/></w:rPr>`
	}
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	wrote := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		wrote = true
		sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:spacing w:after="40"/></w:pPr><w:r>%s<w:t xml:space="preserve">%s</w:t></w:r></w:p>`, rpr, xmlEsc(line)))
	}
	if !wrote {
		// a vertically merged continuation cell still needs one paragraph
		dash := "-"
		if vMerge == "continue" {
			dash = ""
		}
		sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:spacing w:after="40"/></w:pPr><w:r>%s<w:t xml:space="preserve">%s</w:t></w:r></w:p>`, rpr, dash))
	}
	sb.WriteString(`</w:tc>`)
	return sb.String()
}

func tableOpen(widths []int) string {
	var sb strings.Builder
	sb.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="9600" w:type="dxa"/><w:tblBorders>`)
	for _, side := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		sb.WriteString(fmt.Sprintf(`<w:%s w:val="single" w:sz="4" w:space="0" w:color="808080"/>`, side))
	}
	sb.WriteString(`</w:tblBorders><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="40" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:bottom w:w="40" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		sb.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, w))
	}
	sb.WriteString(`</w:tblGrid>`)
	return sb.String()
}

func headerRow(widths []int, titles []string) string {
	var sb strings.Builder
	sb.WriteString(`<w:tr><w:trPr><w:tblHeader/></w:trPr>`)
	for i, t := range titles {
		sb.WriteString(tableCell(widths[i], t, true, "D9E2F3", ""))
	}
	sb.WriteString(`</w:tr>`)
	return sb.String()
}

var rootCauseWidths = []int{1800, 1200, 2200, 2200, 2200}
var actionPlanWidths = []int{600, 3000, 1400, 1500, 1500, 1600}

// rootCauseTable renders the Temuan | Kategori | Why-1..3 table: one row per
// category (People/Process/Policy/System/Eksternal, "-" cells when empty),
// then any other method as its own row. The finding text is vertically merged.
func rootCauseTable(findingText string, items []models.RootCauseItem) string {
	type row struct{ cat, w1, w2, w3 string }
	var rows []row
	used := make([]bool, len(items))
	for _, cat := range rootCauseCategories {
		found := false
		for i, it := range items {
			if !used[i] && strings.EqualFold(strings.TrimSpace(it.Method), cat) {
				used[i] = true
				found = true
				rows = append(rows, row{cat, it.W1, it.W2, it.W3})
			}
		}
		if !found {
			rows = append(rows, row{cat, "", "", ""})
		}
	}
	for i, it := range items {
		if !used[i] {
			rows = append(rows, row{strings.TrimSpace(it.Method), it.W1, it.W2, it.W3})
		}
	}

	var sb strings.Builder
	sb.WriteString(tableOpen(rootCauseWidths))
	sb.WriteString(headerRow(rootCauseWidths, []string{"Temuan", "Kategori", "Why-1", "Why-2", "Why-3"}))
	for i, r := range rows {
		merge := "continue"
		text := ""
		if i == 0 {
			merge = "restart"
			text = findingText
		}
		sb.WriteString(`<w:tr>`)
		sb.WriteString(tableCell(rootCauseWidths[0], text, false, "", merge))
		sb.WriteString(tableCell(rootCauseWidths[1], r.cat, false, "", ""))
		sb.WriteString(tableCell(rootCauseWidths[2], r.w1, false, "", ""))
		sb.WriteString(tableCell(rootCauseWidths[3], r.w2, false, "", ""))
		sb.WriteString(tableCell(rootCauseWidths[4], r.w3, false, "", ""))
		sb.WriteString(`</w:tr>`)
	}
	sb.WriteString(`</w:tbl>`)
	sb.WriteString(`<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>`)
	return sb.String()
}

// effectiveActionPlans returns the finding's action plans, or one row built
// from the legacy Action string. With neither, one empty row ("-" cells).
func effectiveActionPlans(f models.AuditReportFinding) []models.FindingActionPlan {
	if len(f.ActionPlans) > 0 {
		return f.ActionPlans
	}
	return []models.FindingActionPlan{{Action: f.Action}}
}

// actionPlanTable renders No | Rencana Aksi | Target Penyelesaian | PIC |
// Final Approver | Evidence.
func actionPlanTable(plans []models.FindingActionPlan) string {
	w := actionPlanWidths
	var sb strings.Builder
	sb.WriteString(tableOpen(w))
	sb.WriteString(headerRow(w, []string{"No", "Rencana Aksi", "Target Penyelesaian", "PIC", "Final Approver", "Evidence"}))
	for i, p := range plans {
		sb.WriteString(`<w:tr>`)
		sb.WriteString(tableCell(w[0], fmt.Sprintf("%d", i+1), false, "", ""))
		sb.WriteString(tableCell(w[1], p.Action, false, "", ""))
		sb.WriteString(tableCell(w[2], p.TargetDate, false, "", ""))
		sb.WriteString(tableCell(w[3], p.PIC, false, "", ""))
		sb.WriteString(tableCell(w[4], p.FinalApprover, false, "", ""))
		sb.WriteString(tableCell(w[5], p.Evidence, false, "", ""))
		sb.WriteString(`</w:tr>`)
	}
	sb.WriteString(`</w:tbl>`)
	sb.WriteString(`<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>`)
	return sb.String()
}

// findingTables renders the BAB V block of one finding after its title line.
func findingTables(f models.AuditReportFinding) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">Kategori Temuan: %s</w:t></w:r></w:p>`, xmlEsc(CategoryLabelID(f.Category))))
	sb.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Root Cause</w:t></w:r></w:p>`)
	sb.WriteString(rootCauseTable(f.Title, f.RootCause))
	sb.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Rencana Aksi (Setelah disepakati dengan auditi) / Rencana Tindak Lanjut</w:t></w:r></w:p>`)
	sb.WriteString(actionPlanTable(effectiveActionPlans(f)))
	return sb.String()
}
