package docxbuilder

import (
	"fmt"
	"strings"

	"audit-service/models"
)

func headingPara(text string) string {
	return fmt.Sprintf(`<w:p><w:pPr><w:spacing w:before="200" w:after="80"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(text))
}

func textPara(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "-"
	}
	var sb strings.Builder
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(strings.TrimSpace(line))))
	}
	return sb.String()
}

func numberedList(items []string) string {
	var sb strings.Builder
	n := 0
	for _, it := range items {
		if strings.TrimSpace(it) == "" {
			continue
		}
		n++
		sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="360" w:hanging="360"/></w:pPr><w:r><w:t xml:space="preserve">%d. %s</w:t></w:r></w:p>`, n, xmlEsc(strings.TrimSpace(it))))
	}
	if n == 0 {
		return textPara("-")
	}
	return sb.String()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// buildLHAContent renders the LHA body in the agreed order: 1 Surat Tugas
// (nomor, tanggal), 2 Obyek Audit, 3 Risiko yang diaudit, 4 Ruang Lingkup,
// 5 Periode Audit, 6 Tanggal Audit, 7 Hasil (Temuan) Audit periode ini.
func buildLHAContent(report *models.AuditResultReport, st *models.AssignmentLetter, wpHeader *models.WorkingPaperHeader, wpRisks []models.WorkingPaperRisk, wpCauses []models.WorkingPaperCause) string {
	var sb strings.Builder

	// 1. Surat Tugas
	letterNo, letterDate := "", ""
	if report != nil {
		letterNo = report.AssignmentLetterID
	}
	if st != nil {
		letterNo = firstNonEmpty(st.LetterNumber, letterNo)
		if st.LetterDate != nil {
			letterDate = st.LetterDate.Format("02 January 2006")
		}
	}
	sb.WriteString(headingPara("1. Surat Tugas"))
	sb.WriteString(textPara("Nomor : " + firstNonEmpty(letterNo, "-")))
	sb.WriteString(textPara("Tanggal : " + firstNonEmpty(letterDate, "-")))

	// 2. Obyek Audit
	obj := ""
	if report != nil {
		obj = report.AuditObject
	}
	if st != nil {
		obj = firstNonEmpty(obj, st.AuditTitle, st.WorkingUnit)
	}
	sb.WriteString(headingPara("2. Obyek Audit"))
	sb.WriteString(textPara(obj))

	// 3. Risiko yang diaudit
	sb.WriteString(headingPara("3. Risiko yang diaudit"))
	risks := make([]string, 0, len(wpRisks))
	for _, r := range wpRisks {
		risks = append(risks, r.Risk)
	}
	sb.WriteString(numberedList(risks))

	// 4. Ruang Lingkup Audit
	sb.WriteString(headingPara("4. Ruang Lingkup Audit"))
	var scope []string
	if wpHeader != nil {
		for _, a := range wpHeader.Activities {
			scope = append(scope, a.Name)
		}
	}
	if len(scope) == 0 && st != nil {
		scope = st.ScopeList
	}
	sb.WriteString(numberedList(scope))

	// 5. Periode Audit
	period := ""
	if report != nil {
		period = report.AuditPeriod
	}
	if wpHeader != nil {
		period = firstNonEmpty(period, wpHeader.Period)
	}
	if st != nil {
		period = firstNonEmpty(period, st.ExecutionPeriod)
	}
	sb.WriteString(headingPara("5. Periode Audit"))
	sb.WriteString(textPara(period))

	// 6. Tanggal Audit
	dates := ""
	if st != nil {
		s, f := strings.TrimSpace(st.StartPeriod), strings.TrimSpace(st.FinishPeriod)
		switch {
		case s != "" && f != "":
			dates = s + " s.d. " + f
		default:
			dates = firstNonEmpty(s, f)
		}
	}
	sb.WriteString(headingPara("6. Tanggal Audit"))
	sb.WriteString(textPara(dates))

	// 7. Hasil (Temuan) Audit periode ini
	sb.WriteString(headingPara("7. HASIL (TEMUAN) AUDIT PERIODE INI"))
	if report == nil || len(report.Findings) == 0 {
		sb.WriteString(textPara("Tidak ada temuan pada periode ini."))
		return sb.String()
	}
	for idx, f := range report.Findings {
		sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:spacing w:before="160"/></w:pPr><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%d. %s</w:t></w:r></w:p>`, idx+1, xmlEsc(f.Title)))
		if len(f.RootCause) == 0 {
			// findings saved before rootCause existed: use the working paper
			// cause whose condition equals the title
			for _, wc := range wpCauses {
				if strings.EqualFold(strings.TrimSpace(wc.Condition), strings.TrimSpace(f.Title)) && len(wc.RootCause) > 0 {
					f.RootCause = wc.RootCause
					break
				}
			}
		}
		sb.WriteString(findingTables(f))
	}
	return sb.String()
}
