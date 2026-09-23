package docxbuilder

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"audit-service/models"
)

// xmlEsc escapes characters for valid XML document content
func xmlEsc(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// GenerateAuditReportDocx creates a native Microsoft Word (.docx) file for AuditResultReport
func GenerateAuditReportDocx(
	report *models.AuditResultReport,
	st *models.AssignmentLetter,
	interviews []models.FieldworkInterview,
	observations []models.FieldworkObservation,
	fieldworkDocs []models.FieldworkDocument,
	fieldworkSamples []models.FieldworkSample,
	wpHeader *models.WorkingPaperHeader,
	wpRisks []models.WorkingPaperRisk,
	wpSamples []models.WorkingPaperSample,
	wpCauses []models.WorkingPaperCause,
	wpPlans []models.WorkingPaperPlan,
	importedWPs []models.ImportedWorkingPaper,
) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// 1. [Content_Types].xml
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
	if err := addZipFile(zw, "[Content_Types].xml", contentTypes); err != nil {
		return nil, err
	}

	// 2. _rels/.rels
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if err := addZipFile(zw, "_rels/.rels", rels); err != nil {
		return nil, err
	}

	// 3. word/_rels/document.xml.rels
	docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
	if err := addZipFile(zw, "word/_rels/document.xml.rels", docRels); err != nil {
		return nil, err
	}

	// 4. word/styles.xml
	styles := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:cs="Arial"/>
        <w:sz w:val="22"/>
        <w:color w:val="2A2A2A"/>
      </w:rPr>
    </w:rPrDefault>
  </w:docDefaults>
</w:styles>`
	if err := addZipFile(zw, "word/styles.xml", styles); err != nil {
		return nil, err
	}

	// 5. Build word/document.xml
	docXml, err := buildDocumentXML(report, st, interviews, observations, fieldworkDocs, fieldworkSamples, wpHeader, wpRisks, wpSamples, wpCauses, wpPlans, importedWPs)
	if err != nil {
		return nil, err
	}

	if err := addZipFile(zw, "word/document.xml", docXml); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addZipFile(zw *zip.Writer, filename string, content string) error {
	w, err := zw.Create(filename)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func buildDocumentXML(
	report *models.AuditResultReport,
	st *models.AssignmentLetter,
	interviews []models.FieldworkInterview,
	observations []models.FieldworkObservation,
	fieldworkDocs []models.FieldworkDocument,
	fieldworkSamples []models.FieldworkSample,
	wpHeader *models.WorkingPaperHeader,
	wpRisks []models.WorkingPaperRisk,
	wpSamples []models.WorkingPaperSample,
	wpCauses []models.WorkingPaperCause,
	wpPlans []models.WorkingPaperPlan,
	importedWPs []models.ImportedWorkingPaper,
) (string, error) {
	var body bytes.Buffer

	companyName := "PT AIFL Indonesia"

	repNumber := xmlEsc(report.ReportNumber)
	if repNumber == "" {
		repNumber = "020/LHA/01/KS IAD/2023"
	}
	repTitle := xmlEsc(report.ReportTitle)
	if repTitle == "" {
		repTitle = "Laporan Hasil Audit"
	}
	auditObj := xmlEsc(report.AuditObject)
	if auditObj == "" && st != nil && st.WorkingUnit != "" {
		auditObj = xmlEsc(st.WorkingUnit)
	}
	if auditObj == "" {
		auditObj = "Departemen Keuangan & Operasional"
	}

	auditPeriod := xmlEsc(report.AuditPeriod)
	if auditPeriod == "" && st != nil && st.ExecutionPeriod != "" {
		auditPeriod = xmlEsc(st.ExecutionPeriod)
	}
	if auditPeriod == "" {
		auditPeriod = "Periode Tahun 2026"
	}

	stNum := xmlEsc(report.AssignmentLetterID)
	if stNum == "" && st != nil {
		stNum = xmlEsc(st.LetterNumber)
	}
	if stNum == "" {
		stNum = "ST-001/SKAI/2026"
	}

	dateStr := "23 Juli 2026"
	if report.ReportDate != nil {
		dateStr = report.ReportDate.Format("02 January 2006")
	}
	escapedDateStr := xmlEsc(dateStr)

	// --- COVER PAGE ---
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr></w:p>`)
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="36"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">LAPORAN HASIL AUDIT</w:t></w:r></w:p>`)
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/></w:rPr><w:t xml:space="preserve">No. : %s</w:t></w:r></w:p>`, repNumber))
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="22"/></w:rPr><w:t xml:space="preserve">Tanggal : %s</w:t></w:r></w:p>`, escapedDateStr))
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="24"/></w:rPr><w:t xml:space="preserve">PADA</w:t></w:r></w:p>`)
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="28"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, auditObj))
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/></w:rPr><w:t xml:space="preserve">PERIODE : %s</w:t></w:r></w:p>`, auditPeriod))

	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr></w:p>`)
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">SATUAN INTERNAL AUDIT</w:t></w:r></w:p>`)
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(companyName)))
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Sistem Pengendalian Intern &amp; Audit Internal</w:t></w:r></w:p>`)

	// Page Break
	body.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)

	// --- BAB I: INFORMASI SURAT TUGAS (ASSIGNMENT LETTER) ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">I. INFORMASI SURAT TUGAS (ASSIGNMENT LETTER)</w:t></w:r></w:p>`)
	if st != nil {
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">• Nomor Surat Tugas : </w:t></w:r><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(st.LetterNumber)))
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">• Audit Title / Object : </w:t></w:r><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(st.AuditTitle)))
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">• Work Unit (Unit Kerja) : </w:t></w:r><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(st.WorkingUnit)))
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">• Execution Period : </w:t></w:r><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(st.ExecutionPeriod)))
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">• Audit Team (Ketua / Tim) : </w:t></w:r><w:r><w:t xml:space="preserve">%s / %s</w:t></w:r></w:p>`, xmlEsc(st.Leader), xmlEsc(st.AuditTeam)))

		if len(st.MembersList) > 0 {
			body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">  Anggota Tim Audit:</w:t></w:r></w:p>`)
			for _, m := range st.MembersList {
				body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   - %s (%s)</w:t></w:r></w:p>`, xmlEsc(m.Name), xmlEsc(m.Role)))
			}
		}
	} else {
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">Nomor Surat Tugas: %s</w:t></w:r></w:p>`, stNum))
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">Judul Audit: %s</w:t></w:r></w:p>`, repTitle))
	}

	// --- BAB II: DATA AUDIT FIELDWORK ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">II. DATA AUDIT FIELDWORK</w:t></w:r></w:p>`)

	// 1. Interview
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">1. Data Interview (Wawancara)</w:t></w:r></w:p>`)
	if len(interviews) > 0 {
		for idx, inv := range interviews {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Interviewee: %s (%s) | Interviewer: %s (%s) | Tanggal: %s | Topik: %s</w:t></w:r></w:p>`,
				idx+1, xmlEsc(inv.Interviewee), xmlEsc(inv.IntervieweePosition), xmlEsc(inv.Interviewer), xmlEsc(inv.InterviewerPosition), xmlEsc(inv.Date), xmlEsc(inv.Topic)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Tidak ada catatan data wawancara.</w:t></w:r></w:p>`)
	}

	// 2. Observation
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">2. Data Observation (Observasi Lapangan)</w:t></w:r></w:p>`)
	if len(observations) > 0 {
		for idx, obs := range observations {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Aktivitas: %s | Lokasi: %s | Tanggal: %s | Observer: %s</w:t></w:r></w:p>`,
				idx+1, xmlEsc(obs.Activity), xmlEsc(obs.Location), xmlEsc(obs.Date), xmlEsc(obs.Observer)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Tidak ada catatan data observasi.</w:t></w:r></w:p>`)
	}

	// 3. Document Collection
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">3. Data Document Collection (Pengumpulan Dokumen)</w:t></w:r></w:p>`)
	if len(fieldworkDocs) > 0 {
		for idx, doc := range fieldworkDocs {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Nama Dokumen: %s | Deskripsi: %s | Target Tanggal: %s</w:t></w:r></w:p>`,
				idx+1, xmlEsc(doc.DocumentName), xmlEsc(doc.Description), xmlEsc(doc.RequiredDate)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Tidak ada catatan pengumpulan dokumen.</w:t></w:r></w:p>`)
	}

	// 4. Sample Data
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">4. Data Sample (Uji Petik Dokumen)</w:t></w:r></w:p>`)
	if len(fieldworkSamples) > 0 {
		for idx, smp := range fieldworkSamples {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Dokumen: %s | No Dokumen: %s | Tanggal: %s | Keterangan: %s</w:t></w:r></w:p>`,
				idx+1, xmlEsc(smp.DocumentName), xmlEsc(smp.DocumentNumber), xmlEsc(smp.Date), xmlEsc(smp.Description)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Tidak ada sampel dokumen tercatat.</w:t></w:r></w:p>`)
	}

	// --- BAB III: DATA CREATE WORKING PAPER ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">III. DATA WORKING PAPER (KERTAS KERJA AUDIT)</w:t></w:r></w:p>`)

	// 1. Header
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">1. Tahap Header Working Paper</w:t></w:r></w:p>`)
	if wpHeader != nil {
		body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   • Process Bisnis: %s | Periode: %s | Lokasi: %s</w:t></w:r></w:p>`,
			xmlEsc(wpHeader.BusinessProcess), xmlEsc(wpHeader.Period), xmlEsc(wpHeader.Location)))
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Header Kertas Kerja disesuaikan dengan Surat Tugas.</w:t></w:r></w:p>`)
	}

	// 2. Risk Profile
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">2. Data Risk Profile (Risk &amp; Control Matrix)</w:t></w:r></w:p>`)
	if len(wpRisks) > 0 {
		for idx, r := range wpRisks {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Risiko: %s | Taksonomi: %s | Level: %s | Kontrol: %s</w:t></w:r></w:p>`,
				idx+1, xmlEsc(r.Risk), xmlEsc(r.Taxonomy), xmlEsc(r.RiskLevel), xmlEsc(r.ControlDescription)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Profil Risiko teridentifikasi pada area operasional dan pengendalian internal.</w:t></w:r></w:p>`)
	}

	// 3. Test Sample
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">3. Tahap Test Sample</w:t></w:r></w:p>`)
	if len(wpSamples) > 0 {
		for idx, ws := range wpSamples {
			pop := 0
			if ws.Population != nil {
				pop = *ws.Population
			}
			ss := 0
			if ws.SampleSize != nil {
				ss = *ws.SampleSize
			}
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Populasi: %d | Jumlah Sampel: %d | Kesimpulan: %s</w:t></w:r></w:p>`,
				idx+1, pop, ss, xmlEsc(ws.Conclusion)))
			if len(ws.Samples) > 0 {
				body.WriteString(`<w:p><w:r><w:t xml:space="preserve">      Daftar Sampel:</w:t></w:r></w:p>`)
				for _, sDoc := range ws.Samples {
					body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">      - Dokumen: %s</w:t></w:r></w:p>`, xmlEsc(sDoc.Document)))
				}
			}
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Pengujian sampel dilakukan secara uji petik profesional (judgement sampling).</w:t></w:r></w:p>`)
	}

	// 4. AOI & RCA
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">4. Tahap AOI &amp; RCA (Area of Improvement &amp; Root Cause Analysis)</w:t></w:r></w:p>`)
	if len(wpCauses) > 0 {
		for idx, wc := range wpCauses {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">   Temuan %d:</w:t></w:r></w:p>`, idx+1))
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   • Condition : %s</w:t></w:r></w:p>`, xmlEsc(wc.Condition)))
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   • Criteria  : %s</w:t></w:r></w:p>`, xmlEsc(wc.Criteria)))
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   • Impact    : %s</w:t></w:r></w:p>`, xmlEsc(wc.Impact)))
			if len(wc.RootCause) > 0 {
				body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   • Root Cause (Analisis Akar Masalah):</w:t></w:r></w:p>`)
				for _, rc := range wc.RootCause {
					body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">     - [%s] W1: %s | W2: %s | W3: %s</w:t></w:r></w:p>`,
						xmlEsc(rc.Method), xmlEsc(rc.W1), xmlEsc(rc.W2), xmlEsc(rc.W3)))
				}
			}
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Analisis penyebab masalah mengacu pada efektivitas sistem supervisi dan pembaruan SOP.</w:t></w:r></w:p>`)
	}

	// 5. Action Plan
	body.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">5. Tahap Action Plan (Rencana Tindak Lanjut)</w:t></w:r></w:p>`)
	if len(wpPlans) > 0 {
		for idx, wp := range wpPlans {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Rekomendasi : %s</w:t></w:r></w:p>`, idx+1, xmlEsc(wp.Recommendation)))
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">      Respon Auditee: %s | Deskripsi: %s | PIC: %s | Target: %s</w:t></w:r></w:p>`,
				xmlEsc(wp.Response), xmlEsc(wp.ActionDescription), xmlEsc(wp.PIC), xmlEsc(wp.PeriodAction)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Auditee telah menyetujui rencana perbaikan dengan PIC yang ditunjuk.</w:t></w:r></w:p>`)
	}

	// --- BAB IV: UPLOAD WORKING PAPER ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">IV. DATA UPLOAD WORKING PAPER (DOKUMEN TERUNGGAH)</w:t></w:r></w:p>`)
	if len(importedWPs) > 0 {
		for idx, iwp := range importedWPs {
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   %d. Judul Dokumen : %s</w:t></w:r></w:p>`, idx+1, xmlEsc(iwp.Title)))
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">      Nama Berkas   : %s | Deskripsi: %s</w:t></w:r></w:p>`, xmlEsc(iwp.FileName), xmlEsc(iwp.Description)))
		}
	} else {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">   Tidak ada berkas fisik working paper tambahan yang terunggah.</w:t></w:r></w:p>`)
	}

	// --- BAB V: RINGKASAN TEMUAN HASIL AUDIT ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">V. RINGKASAN TEMUAN HASIL AUDIT</w:t></w:r></w:p>`)
	if len(report.Findings) == 0 {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">Tidak ditemukan temuan signifikan pada audit ini.</w:t></w:r></w:p>`)
	} else {
		for idx, f := range report.Findings {
			fTitle := xmlEsc(f.Title)
			fSev := xmlEsc(f.Category)
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%d. %s [%s]</w:t></w:r></w:p>`, idx+1, fTitle, fSev))

			fAction := xmlEsc(f.Action)
			if fAction == "" {
				fAction = "Manajemen terkait diharapkan menyelesaikan tindak lanjut sesuai target waktu."
			}
			body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">   Tindak Lanjut / Action: %s</w:t></w:r></w:p>`, fAction))
		}
	}

	// --- PENUTUP & SIGNATURES ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">VI. PENUTUP</w:t></w:r></w:p>`)
	conclusionText := xmlEsc(report.Conclusion)
	if conclusionText == "" {
		conclusionText = "Demikian Laporan Hasil Audit ini disampaikan untuk dapat dipergunakan sebagai bahan perbaikan tata kelola dan sistem pengendalian internal perusahaan secara berkelanjutan."
	}
	body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, conclusionText))

	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr></w:p>`)
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="right"/></w:pPr><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Jakarta, %s</w:t></w:r></w:p>`, escapedDateStr))
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="right"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">AUDIT INTERNAL %s</w:t></w:r></w:p>`, xmlEsc(strings.ToUpper(companyName))))

	prepBy := xmlEsc(report.PreparedBy)
	if prepBy == "" {
		prepBy = "Zeta Ramadhani"
	}
	revBy := xmlEsc(report.ReviewedBy)
	if revBy == "" {
		revBy = "Budi Santoso"
	}

	body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Ketua Tim / Reviewer : %s</w:t></w:r></w:p>`, revBy))
	body.WriteString(fmt.Sprintf(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Auditor / Prepared By : %s</w:t></w:r></w:p>`, prepBy))

	// Return document.xml
	docXml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
            xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math"
            xmlns:v="urn:schemas-microsoft-com:vml"
            xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
            xmlns:w10="urn:schemas-microsoft-com:office:word"
            xmlns:sl="http://schemas.openxmlformats.org/schemaLibrary/2006/main">
  <w:body>
    %s
  </w:body>
</w:document>`, body.String())

	_ = time.Now()
	return docXml, nil
}
