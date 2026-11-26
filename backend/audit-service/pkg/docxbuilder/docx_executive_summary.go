package docxbuilder

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"audit-service/models"
)

type execFollowUpRow struct {
	Status     string  `json:"status"`
	Jumlah     int     `json:"jumlah"`
	Persentase float64 `json:"persentase"`
	Keterangan string  `json:"keterangan"`
}

type execTopFinding struct {
	UnitDivision string `json:"unitDivision"`
	JudulTemuan  string `json:"judulTemuan"`
	Risiko       string `json:"risiko"`
	StatusTL     string `json:"statusTL"`
	Usulan       string `json:"usulan"`
}

type execMatrixRow struct {
	Nomor        string  `json:"nomor"`
	Division     string  `json:"division"`
	UnitKerja    string  `json:"unitKerja"`
	ProsesBisnis string  `json:"prosesBisnis"`
	JudulTemuan  string  `json:"judulTemuan"`
	NilaiRisiko  string  `json:"nilaiRisiko"`
	Rekomendasi  string  `json:"rekomendasi"`
	DueDate      string  `json:"dueDate"`
	PICUnit      string  `json:"picUnit"`
	Progres      float64 `json:"progres"`
	Status       string  `json:"status"`
	BuktiTL      string  `json:"buktiTL"`
}

// ExecutiveSummaryDocxOptions carries optional linked LHA metadata for an
// Individual Executive Summary .docx and explicit variant selection.
type ExecutiveSummaryDocxOptions struct {
	IsCompilation bool
	ReportTitle   string
	AuditObject   string
	Department    string
}

func execRiskLabelID(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "-"
	}
	switch strings.ToLower(v) {
	case "very significant", "significant", "high", "moderate-high", "moderate to high", "tinggi", "sangat signifikan", "signifikan":
		return "Tinggi"
	case "quite significant", "moderate", "sedang", "cukup signifikan":
		return "Sedang"
	case "not significant", "low", "low-moderate", "low to moderate", "rendah", "tidak signifikan":
		return "Rendah"
	default:
		return v
	}
}

func execQuarterLabel(q int) string {
	switch q {
	case 1:
		return "JANUARI – MARET (TRIWULAN I)"
	case 2:
		return "APRIL – JUNI (TRIWULAN II)"
	case 3:
		return "JULI – SEPTEMBER (TRIWULAN III)"
	case 4:
		return "OKTOBER – DESEMBER (TRIWULAN IV)"
	default:
		return "TRIWULAN I"
	}
}

func execFormatDateID(raw string, fallbackYear int) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		if fallbackYear <= 0 {
			fallbackYear = 2026
		}
		return fmt.Sprintf("................ %d", fallbackYear)
	}
	datePart := strings.Split(s, "T")[0]
	parts := strings.Split(datePart, "-")
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	if len(parts) == 3 {
		var y, m, d int
		_, _ = fmt.Sscanf(parts[0], "%d", &y)
		_, _ = fmt.Sscanf(parts[1], "%d", &m)
		_, _ = fmt.Sscanf(parts[2], "%d", &d)
		if y > 0 && m >= 1 && m <= 12 && d >= 1 && d <= 31 {
			return fmt.Sprintf("%d %s %d", d, months[m-1], y)
		}
	}
	return s
}

func execCenterPara(text string, sizeHalfPt int, bold bool) string {
	bTag := ""
	if bold {
		bTag = "<w:b/>"
	}
	return fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="40" w:after="40"/></w:pPr><w:r><w:rPr>%s<w:sz w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, bTag, sizeHalfPt, xmlEsc(text))
}

func execFollowUpTableXML(rawJSON string, highRisk int) (string, float64) {
	var rows []execFollowUpRow
	if strings.TrimSpace(rawJSON) != "" {
		_ = json.Unmarshal([]byte(rawJSON), &rows)
	}
	if len(rows) == 0 {
		rows = []execFollowUpRow{
			{Status: "Closed", Jumlah: 0},
			{Status: "In Progress", Jumlah: 0},
			{Status: "Overdue", Jumlah: 0},
		}
	}
	total := 0
	closedCount := 0
	for _, r := range rows {
		total += r.Jumlah
		if strings.EqualFold(strings.TrimSpace(r.Status), "Closed") {
			closedCount += r.Jumlah
		}
	}
	onTimePct := 0.0
	if total > 0 {
		onTimePct = float64(closedCount) / float64(total) * 100
	}

	widths := []int{2200, 1400, 1400, 4600}
	var sb strings.Builder
	sb.WriteString(tableOpen(widths))
	sb.WriteString(headerRow(widths, []string{"Status", "Jumlah", "%", "Keterangan"}))
	for _, r := range rows {
		if strings.EqualFold(strings.TrimSpace(r.Status), "No Action") && r.Jumlah <= 0 {
			continue
		}
		pct := r.Persentase
		if total > 0 {
			pct = float64(r.Jumlah) / float64(total) * 100
		}
		note := strings.TrimSpace(r.Keterangan)
		if note == "" {
			switch strings.TrimSpace(r.Status) {
			case "Closed":
				if r.Jumlah > 0 {
					note = "Telah divalidasi SPI"
				}
			case "In Progress":
				if r.Jumlah > 0 {
					note = "On track pelaksanaan tindak lanjut"
				}
			case "Overdue":
				if r.Jumlah > 0 && highRisk > 0 {
					hr := r.Jumlah
					if highRisk < hr {
						hr = highRisk
					}
					note = fmt.Sprintf("%d Risiko Tinggi", hr)
				} else if r.Jumlah > 0 {
					note = "Melewati batas waktu"
				}
			}
		}
		sb.WriteString("<w:tr>")
		sb.WriteString(tableCell(widths[0], r.Status, false, "", ""))
		sb.WriteString(tableCell(widths[1], fmt.Sprintf("%d", r.Jumlah), false, "", ""))
		sb.WriteString(tableCell(widths[2], fmt.Sprintf("%.1f%%", pct), false, "", ""))
		sb.WriteString(tableCell(widths[3], note, false, "", ""))
		sb.WriteString("</w:tr>")
	}
	totalPct := "0%"
	if total > 0 {
		totalPct = "100%"
	}
	sb.WriteString("<w:tr>")
	sb.WriteString(tableCell(widths[0], "Total", true, "F2F4F8", ""))
	sb.WriteString(tableCell(widths[1], fmt.Sprintf("%d", total), true, "F2F4F8", ""))
	sb.WriteString(tableCell(widths[2], totalPct, true, "F2F4F8", ""))
	sb.WriteString(tableCell(widths[3], "-", true, "F2F4F8", ""))
	sb.WriteString("</w:tr></w:tbl><w:p><w:pPr><w:spacing w:after=\"80\"/></w:pPr></w:p>")
	return sb.String(), onTimePct
}

func execTopFindingsTableXML(rawJSON string) (string, []execTopFinding) {
	var list []execTopFinding
	if strings.TrimSpace(rawJSON) != "" {
		_ = json.Unmarshal([]byte(rawJSON), &list)
	}
	widths := []int{600, 1600, 2800, 1300, 1400, 1900}
	var sb strings.Builder
	sb.WriteString(tableOpen(widths))
	sb.WriteString(headerRow(widths, []string{"No", "Unit Division", "Judul Temuan", "Risiko", "Status TL", "Usulan"}))
	if len(list) == 0 {
		sb.WriteString("<w:tr>")
		sb.WriteString(tableCell(widths[0], "-", false, "", ""))
		sb.WriteString(tableCell(widths[1], "-", false, "", ""))
		sb.WriteString(tableCell(widths[2], "Tidak ada temuan signifikan pada periode ini", false, "", ""))
		sb.WriteString(tableCell(widths[3], "-", false, "", ""))
		sb.WriteString(tableCell(widths[4], "-", false, "", ""))
		sb.WriteString(tableCell(widths[5], "-", false, "", ""))
		sb.WriteString("</w:tr>")
	} else {
		for i, f := range list {
			statusTL := strings.TrimSpace(f.StatusTL)
			if strings.EqualFold(statusTL, "Overdue") {
				statusTL = "Overdue/Telat"
			}
			sb.WriteString("<w:tr>")
			sb.WriteString(tableCell(widths[0], fmt.Sprintf("%d", i+1), false, "", ""))
			sb.WriteString(tableCell(widths[1], f.UnitDivision, false, "", ""))
			sb.WriteString(tableCell(widths[2], f.JudulTemuan, false, "", ""))
			sb.WriteString(tableCell(widths[3], execRiskLabelID(f.Risiko), false, "", ""))
			sb.WriteString(tableCell(widths[4], statusTL, false, "", ""))
			sb.WriteString(tableCell(widths[5], f.Usulan, false, "", ""))
			sb.WriteString("</w:tr>")
		}
	}
	sb.WriteString("</w:tbl><w:p><w:pPr><w:spacing w:after=\"80\"/></w:pPr></w:p>")
	return sb.String(), list
}

func execMatrixTableXML(rawJSON string) (string, []execMatrixRow) {
	var rows []execMatrixRow
	if strings.TrimSpace(rawJSON) != "" {
		_ = json.Unmarshal([]byte(rawJSON), &rows)
	}
	widths := []int{800, 650, 950, 850, 1250, 650, 1050, 750, 750, 550, 650, 700}
	var sb strings.Builder
	sb.WriteString(tableOpen(widths))
	sb.WriteString(headerRow(widths, []string{"Nomor", "Division", "Unit Kerja", "Proses Bisnis", "Judul Temuan", "Nilai Risiko", "Rekomendasi", "Due Date", "PIC Unit", "% Progres", "Status", "Bukti TL"}))
	sb.WriteString(headerRow(widths, []string{"(A)", "(B)", "(C)", "(D)", "(E)", "(F)", "(G)", "(H)", "(I)", "(J)", "(K)", "(L)"}))
	if len(rows) == 0 {
		sb.WriteString("<w:tr>")
		for i, w := range widths {
			val := "-"
			if i == 9 {
				val = "0%"
			}
			sb.WriteString(tableCell(w, val, false, "", ""))
		}
		sb.WriteString("</w:tr>")
	} else {
		for _, r := range rows {
			unitKerja := firstNonEmpty(r.UnitKerja, r.Division, "-")
			sb.WriteString("<w:tr>")
			sb.WriteString(tableCell(widths[0], r.Nomor, false, "", ""))
			sb.WriteString(tableCell(widths[1], r.Division, false, "", ""))
			sb.WriteString(tableCell(widths[2], unitKerja, false, "", ""))
			sb.WriteString(tableCell(widths[3], r.ProsesBisnis, false, "", ""))
			sb.WriteString(tableCell(widths[4], r.JudulTemuan, false, "", ""))
			sb.WriteString(tableCell(widths[5], execRiskLabelID(r.NilaiRisiko), false, "", ""))
			sb.WriteString(tableCell(widths[6], r.Rekomendasi, false, "", ""))
			sb.WriteString(tableCell(widths[7], r.DueDate, false, "", ""))
			sb.WriteString(tableCell(widths[8], r.PICUnit, false, "", ""))
			sb.WriteString(tableCell(widths[9], fmt.Sprintf("%.0f%%", r.Progres), false, "", ""))
			sb.WriteString(tableCell(widths[10], r.Status, false, "", ""))
			sb.WriteString(tableCell(widths[11], r.BuktiTL, false, "", ""))
			sb.WriteString("</w:tr>")
		}
	}
	sb.WriteString("</w:tbl><w:p><w:pPr><w:spacing w:after=\"80\"/></w:pPr></w:p>")
	return sb.String(), rows
}

// GenerateExecutiveSummaryDocx creates a native Microsoft Word (.docx) file
// for an Approved Executive Summary (Compilation or Individual).
func GenerateExecutiveSummaryDocx(es *models.ExecutiveSummary, opts ExecutiveSummaryDocxOptions) ([]byte, error) {
	if es == nil {
		return nil, fmt.Errorf("executive summary is required")
	}
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	var sigImgData []byte
	sigImgExt := "png"
	sigImgRelID := ""
	sigStr := strings.TrimSpace(es.SignatureImage)
	if idx := strings.Index(sigStr, "base64,"); idx != -1 {
		if decoded, err := base64.StdEncoding.DecodeString(sigStr[idx+7:]); err == nil && len(decoded) > 0 {
			sigImgData = decoded
			sigImgRelID = "rId_es_sig_1"
			if strings.Contains(sigStr, "image/jpeg") || strings.Contains(sigStr, "image/jpg") {
				sigImgExt = "jpg"
			}
		}
	}

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Default Extension="jpg" ContentType="image/jpeg"/>
  <Default Extension="jpeg" ContentType="image/jpeg"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
	if err := addZipFile(zw, "[Content_Types].xml", contentTypes); err != nil {
		return nil, err
	}

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if err := addZipFile(zw, "_rels/.rels", rels); err != nil {
		return nil, err
	}

	var docRels bytes.Buffer
	docRels.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`)
	if len(sigImgData) > 0 {
		mediaName := fmt.Sprintf("media/es_sig.%s", sigImgExt)
		if err := addZipBinary(zw, "word/"+mediaName, sigImgData); err != nil {
			return nil, err
		}
		docRels.WriteString(fmt.Sprintf(`
  <Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`, sigImgRelID, mediaName))
	}
	docRels.WriteString("\n</Relationships>")
	if err := addZipFile(zw, "word/_rels/document.xml.rels", docRels.String()); err != nil {
		return nil, err
	}

	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:cs="Arial"/>
        <w:sz w:val="20"/>
        <w:color w:val="2A2A2A"/>
      </w:rPr>
    </w:rPrDefault>
  </w:docDefaults>
</w:styles>`
	if err := addZipFile(zw, "word/styles.xml", styles); err != nil {
		return nil, err
	}

	docXML := buildExecutiveSummaryDocXML(es, opts, sigImgRelID)
	if err := addZipFile(zw, "word/document.xml", docXML); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildExecutiveSummaryDocXML(es *models.ExecutiveSummary, opts ExecutiveSummaryDocxOptions, sigRelID string) string {
	var body bytes.Buffer
	year := es.Tahun
	if year <= 0 {
		year = 2026
	}
	quarter := es.Quarter
	if quarter < 1 || quarter > 4 {
		quarter = 1
	}

	highRisk, medRisk, lowRisk := models.LegacyRiskCounts(es.RisikoHigh, es.RisikoModerateHigh, es.RisikoModerate, es.RisikoLowModerate, es.RisikoLow)
	if highRisk == 0 && medRisk == 0 && lowRisk == 0 {
		highRisk, medRisk, lowRisk = es.RisikoTinggi, es.RisikoSedang, es.RisikoRendah
	}

	followUpXML, onTimePct := execFollowUpTableXML(es.FollowUpTable, highRisk)
	topFindingsXML, topList := execTopFindingsTableXML(es.TopFindings)
	matrixXML, matrixList := execMatrixTableXML(es.MatriksKompilasi)

	totalFindings := highRisk + medRisk + lowRisk
	if totalFindings == 0 {
		if len(matrixList) > 0 {
			totalFindings = len(matrixList)
		} else if len(topList) > 0 {
			totalFindings = len(topList)
		}
	}
	recCount := es.JumlahRekomendasi
	if recCount <= 0 {
		recCount = totalFindings
	}

	statusDateRaw := es.SignatureTanggal
	if es.ApprovedAt != nil && !es.ApprovedAt.IsZero() {
		statusDateRaw = es.ApprovedAt.Format("2006-01-02")
	}
	statusUntilDate := execFormatDateID(statusDateRaw, year)

	// Ensure narrative has no raw "...." placeholders
	narrative := strings.TrimSpace(es.Narrative)
	if narrative == "" || strings.Contains(narrative, "....") || strings.Contains(narrative, "….") {
		lhaCount := es.JumlahLaporan
		if lhaCount <= 0 {
			if opts.IsCompilation {
				lhaCount = 0
			} else {
				lhaCount = 1
			}
		}
		topic := "-"
		if len(topList) > 0 && strings.TrimSpace(topList[0].JudulTemuan) != "" {
			topic = strings.TrimSpace(topList[0].JudulTemuan)
		} else if len(matrixList) > 0 && strings.TrimSpace(matrixList[0].JudulTemuan) != "" {
			topic = strings.TrimSpace(matrixList[0].JudulTemuan)
		} else if strings.TrimSpace(opts.ReportTitle) != "" {
			topic = strings.TrimSpace(opts.ReportTitle)
		}
		process := firstNonEmpty(opts.AuditObject, opts.ReportTitle, opts.Department)
		if process == "" && len(matrixList) > 0 {
			process = firstNonEmpty(matrixList[0].ProsesBisnis, matrixList[0].UnitKerja, matrixList[0].Division)
		}
		if process == "" && len(topList) > 0 {
			process = firstNonEmpty(topList[0].UnitDivision, topList[0].JudulTemuan)
		}
		if process == "" {
			process = "-"
		}
		periodLabel := firstNonEmpty(es.PeriodeBulan, fmt.Sprintf("Kuartal %d", quarter))
		narrative = fmt.Sprintf(
			"Periode %s %d, SPI telah menerbitkan %d LHA. Total temuan %d dengan %d rekomendasi. Tingkat penyelesaian on-time %.1f%%. Terdapat %d temuan risiko tinggi overdue/terlambat terkait %s yang perlu arahan Direksi. Temuan berulang tertinggi pada proses %s.",
			periodLabel, year, lhaCount, totalFindings, recCount, onTimePct, highRisk, topic, process,
		)
	}

	if opts.IsCompilation {
		periodHeader := fmt.Sprintf("BULAN %s TAHUN %d", execQuarterLabel(quarter), year)
		if strings.TrimSpace(es.PeriodeBulan) != "" {
			periodHeader = fmt.Sprintf("%s (%s) TAHUN %d", strings.ToUpper(strings.TrimSpace(es.PeriodeBulan)), execQuarterLabel(quarter), year)
		}
		docNum := firstNonEmpty(es.NomorDokumen, fmt.Sprintf("LKA/B/%03d/SPI/KAI/%d", quarter, year))

		body.WriteString(execCenterPara("EXECUTIVE SUMMARY COMPILATION", 28, true))
		body.WriteString(execCenterPara("HASIL AUDIT TRIWULANAN", 26, true))
		body.WriteString(execCenterPara("PERIODE: "+periodHeader, 22, true))
		body.WriteString(execCenterPara("Nomor: "+docNum, 22, true))
		body.WriteString(`<w:p><w:pPr><w:spacing w:after="120"/></w:pPr></w:p>`)

		body.WriteString(headingPara("I. EXECUTIVE SUMMARY"))
		body.WriteString(textPara(narrative))

		body.WriteString(headingPara("II. STATISTIK KOMPILASI"))
		body.WriteString(textPara(fmt.Sprintf("Jumlah Laporan: Internal %d LHA.", es.JumlahLaporan)))
		body.WriteString(textPara(fmt.Sprintf("Jumlah Temuan: %d temuan. Rincian: Risiko Tinggi %d, Sedang %d, Rendah %d.", totalFindings, highRisk, medRisk, lowRisk)))
		body.WriteString(textPara(fmt.Sprintf("Jumlah Rekomendasi: %d rekomendasi.", recCount)))

		body.WriteString(headingPara("III. STATUS TINDAK LANJUT s.d " + statusUntilDate))
		body.WriteString(followUpXML)

		body.WriteString(headingPara("IV. TOP 5 TEMUAN SIGNIFIKAN"))
		body.WriteString(topFindingsXML)

		body.WriteString(headingPara("V. ANALISIS TEMUAN BERULANG"))
		body.WriteString(textPara(firstNonEmpty(es.AkarMasalah, "-")))

		body.WriteString(headingPara("VI. TREN & GRAFIK"))
		body.WriteString(textPara("[Lampirkan Grafik: Jumlah Temuan per Bulan, Aging Rekomendasi, % On-time]"))
		body.WriteString(textPara(fmt.Sprintf("Ringkasan Tren Periode Triwulan %d Tahun %d: Total %d LHA diterbitkan dengan %d temuan (%d Risiko Tinggi, %d Sedang, %d Rendah) dan tingkat penyelesaian (On-time/Closed) sebesar %.1f%%.", quarter, year, es.JumlahLaporan, totalFindings, highRisk, medRisk, lowRisk, onTimePct)))

		body.WriteString(headingPara("VII. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN"))
		body.WriteString(textPara(firstNonEmpty(es.Kesimpulan, "-")))
		if strings.TrimSpace(es.ExecutiveNote) != "" {
			body.WriteString(textPara("Catatan Executive: " + strings.TrimSpace(es.ExecutiveNote)))
		}

		body.WriteString(buildExecSignatureBlockXML(es, year, sigRelID))

		body.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Lampiran.</w:t></w:r></w:p>`)
		body.WriteString(execCenterPara("MATRIKS INDUK KOMPILASI TEMUAN", 24, true))
		body.WriteString(execCenterPara("LAPORAN KOMPILASI HASIL AUDIT PERIODE: "+periodHeader, 20, true))
		body.WriteString(execCenterPara("Nomor: "+docNum, 20, true))
		body.WriteString(`<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>`)
		body.WriteString(matrixXML)
		body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="18"/></w:rPr><w:t xml:space="preserve">Keterangan:</w:t></w:r></w:p>`)
		body.WriteString(textPara("Kolom C: Unit Kerja\nKolom F: Tinggi; Sedang; Rendah\nKolom K: Closed; In Progress; Overdue"))
	} else {
		// Individual Executive Summary template adapted strictly to Individual feature data
		periodHeader := fmt.Sprintf("TAHUN %d", year)
		if strings.TrimSpace(es.PeriodeBulan) != "" {
			pb := strings.ToUpper(strings.TrimSpace(es.PeriodeBulan))
			if strings.Contains(pb, fmt.Sprintf("%d", year)) {
				periodHeader = pb
			} else {
				periodHeader = fmt.Sprintf("%s TAHUN %d", pb, year)
			}
		}
		lhaNum := firstNonEmpty(es.NomorDokumen, "-")
		stNum := firstNonEmpty(es.AssignmentLetterID, "-")
		repTitle := firstNonEmpty(opts.ReportTitle, opts.AuditObject, fmt.Sprintf("Laporan Hasil Audit (%s)", lhaNum))
		dept := firstNonEmpty(opts.Department)
		if dept == "" && len(topList) > 0 {
			dept = topList[0].UnitDivision
		}
		dept = firstNonEmpty(dept, "-")

		body.WriteString(execCenterPara("EXECUTIVE SUMMARY INDIVIDUAL", 28, true))
		body.WriteString(execCenterPara("LAPORAN HASIL AUDIT (LHA)", 26, true))
		body.WriteString(execCenterPara("PERIODE: "+periodHeader, 22, true))
		body.WriteString(execCenterPara("Nomor LHA: "+lhaNum, 22, true))
		if stNum != "-" {
			body.WriteString(execCenterPara("Surat Tugas: "+stNum, 20, false))
		}
		body.WriteString(`<w:p><w:pPr><w:spacing w:after="120"/></w:pPr></w:p>`)

		body.WriteString(headingPara("I. EXECUTIVE SUMMARY"))
		body.WriteString(textPara(narrative))

		body.WriteString(headingPara("II. INFORMASI & STATISTIK LHA INDIVIDUAL"))
		body.WriteString(textPara("Nomor LHA: " + lhaNum))
		body.WriteString(textPara("Nomor Surat Tugas: " + stNum))
		body.WriteString(textPara("Judul Laporan / Objek Audit: " + repTitle))
		body.WriteString(textPara("Unit / Divisi: " + dept))
		body.WriteString(textPara(fmt.Sprintf("Jumlah Temuan: %d temuan. Rincian: Risiko Tinggi (Signifikan) %d, Sedang (Cukup Signifikan) %d, Rendah (Tidak Signifikan) %d.", totalFindings, highRisk, medRisk, lowRisk)))
		body.WriteString(textPara(fmt.Sprintf("Jumlah Rekomendasi: %d rekomendasi.", recCount)))

		body.WriteString(headingPara("III. STATUS TINDAK LANJUT s.d " + statusUntilDate))
		body.WriteString(followUpXML)

		body.WriteString(headingPara("IV. TOP 5 TEMUAN SIGNIFIKAN"))
		body.WriteString(topFindingsXML)

		body.WriteString(headingPara("V. ANALISIS TEMUAN BERULANG & AKAR MASALAH"))
		body.WriteString(textPara(firstNonEmpty(es.AkarMasalah, "-")))

		body.WriteString(headingPara("VI. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN"))
		body.WriteString(textPara(firstNonEmpty(es.Kesimpulan, "-")))
		if strings.TrimSpace(es.ExecutiveNote) != "" {
			body.WriteString(textPara("Catatan Executive: " + strings.TrimSpace(es.ExecutiveNote)))
		}

		body.WriteString(buildExecSignatureBlockXML(es, year, sigRelID))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
            xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
            xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
            xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">
  <w:body>
    %s
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" w:header="708" w:footer="708" w:gutter="0"/>
    </w:sectPr>
  </w:body>
</w:document>`, body.String())
}

func buildExecSignatureBlockXML(es *models.ExecutiveSummary, year int, sigRelID string) string {
	place := firstNonEmpty(es.SignatureTempat, "Jakarta")
	dateLabel := execFormatDateID(es.SignatureTanggal, year)
	headName := firstNonEmpty(es.SignatureNamaKepala, es.SignatureName, "[Nama Kepala SPI]")
	nik := firstNonEmpty(es.SignatureNIK, "........................")

	var sb strings.Builder
	sb.WriteString(`<w:p><w:pPr><w:spacing w:before="240" w:after="40"/></w:pPr></w:p>`)
	sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="40"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(fmt.Sprintf("%s, %s", place, dateLabel))))
	if sigRelID != "" {
		sb.WriteString(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="80"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Kepala SPI,</w:t></w:r></w:p>`)
		sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="80"/></w:pPr><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="1371600" cy="548640"/><wp:docPr id="2001" name="Signature"/><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:nvPicPr><pic:cNvPr id="0" name="sig.png"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="1371600" cy="548640"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`, sigRelID))
	} else {
		sb.WriteString(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="600"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Kepala SPI,</w:t></w:r></w:p>`)
	}
	sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="40"/></w:pPr><w:r><w:rPr><w:b/><w:u w:val="single"/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(headName)))
	sb.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="120"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc("NIK. "+nik)))
	return sb.String()
}
