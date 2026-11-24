package docxbuilder

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"audit-service/models"
)

type docxSignature struct {
	Name     string
	Role     string
	HasImage bool
	RelID    string
	DocPrID  int
}

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

	// Process Signatures and extract images
	sigItems := []docxSignature{}
	type sigImgItem struct {
		relID    string
		filename string
		data     []byte
	}
	sigImages := []sigImgItem{}

	if report != nil && len(report.Signatures) > 0 {
		for i, s := range report.Signatures {
			item := docxSignature{
				Name:    s.Name,
				Role:    s.Role,
				DocPrID: 1000 + i + 1,
			}
			if item.Role == "" {
				item.Role = "Team Member"
			}
			sigStr := strings.TrimSpace(s.Signature)
			if idx := strings.Index(sigStr, "base64,"); idx != -1 {
				base64Data := sigStr[idx+7:]
				imgData, err := base64.StdEncoding.DecodeString(base64Data)
				if err == nil && len(imgData) > 0 {
					relID := fmt.Sprintf("rId_sig_%d", i+2)
					ext := "png"
					if strings.Contains(sigStr, "image/jpeg") || strings.Contains(sigStr, "image/jpg") {
						ext = "jpg"
					}
					filename := fmt.Sprintf("word/media/sig_%d.%s", i+1, ext)
					sigImages = append(sigImages, sigImgItem{
						relID:    relID,
						filename: filename,
						data:     imgData,
					})
					item.HasImage = true
					item.RelID = relID
				}
			}
			sigItems = append(sigItems, item)
		}
	} else if st != nil && len(st.MembersList) > 0 {
		for i, m := range st.MembersList {
			role := m.Role
			if role == "" {
				role = "Team Member"
			}
			sigItems = append(sigItems, docxSignature{
				Name:    m.Name,
				Role:    role,
				DocPrID: 1000 + i + 1,
			})
		}
	} else {
		prepBy := "Auditor"
		revBy := "Ketua Tim"
		if report != nil {
			if report.PreparedBy != "" {
				prepBy = report.PreparedBy
			}
			if report.ReviewedBy != "" {
				revBy = report.ReviewedBy
			}
		}
		sigItems = append(sigItems,
			docxSignature{Name: revBy, Role: "Ketua Tim / Reviewer", DocPrID: 1001},
			docxSignature{Name: prepBy, Role: "Auditor / Prepared By", DocPrID: 1002},
		)
	}

	// 1. [Content_Types].xml
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

	// 2. _rels/.rels
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if err := addZipFile(zw, "_rels/.rels", rels); err != nil {
		return nil, err
	}

	// Write media files to zip
	for _, img := range sigImages {
		if err := addZipBinary(zw, img.filename, img.data); err != nil {
			return nil, err
		}
	}

	// 3. word/_rels/document.xml.rels
	docRelsBuf := new(bytes.Buffer)
	docRelsBuf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`)
	for _, img := range sigImages {
		target := strings.TrimPrefix(img.filename, "word/")
		docRelsBuf.WriteString(fmt.Sprintf(`
  <Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`, img.relID, target))
	}
	docRelsBuf.WriteString("\n</Relationships>")

	if err := addZipFile(zw, "word/_rels/document.xml.rels", docRelsBuf.String()); err != nil {
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
	docXml, err := buildDocumentXML(report, st, interviews, observations, fieldworkDocs, fieldworkSamples, wpHeader, wpRisks, wpSamples, wpCauses, wpPlans, importedWPs, sigItems)
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

func addZipBinary(zw *zip.Writer, filename string, content []byte) error {
	w, err := zw.Create(filename)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
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
	sigItems []docxSignature,
) (string, error) {
	var body bytes.Buffer

	// No hardcoded company: an unresolved name renders as "-" on the cover
	// and is left out of the "AUDIT INTERNAL" signature line.
	companyName := ""
	if report != nil && report.CompanyName != "" {
		companyName = report.CompanyName
	} else if st != nil && st.CompanyName != "" {
		companyName = st.CompanyName
	}
	coverCompanyName := companyName
	if strings.TrimSpace(coverCompanyName) == "" {
		coverCompanyName = "-"
	}
	auditInternalLine := "AUDIT INTERNAL"
	if strings.TrimSpace(companyName) != "" {
		auditInternalLine += " " + strings.ToUpper(companyName)
	}

	repNumber := xmlEsc(report.ReportNumber)
	if repNumber == "" {
		repNumber = "LHA-020/SKAI/2023"
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
		auditObj = "Departemen Keuangan &amp; Operasional"
	}

	auditTitle := auditObj
	if st != nil && st.AuditTitle != "" {
		auditTitle = xmlEsc(st.AuditTitle)
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
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="28"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, auditTitle))
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/></w:rPr><w:t xml:space="preserve">PERIODE : %s</w:t></w:r></w:p>`, auditPeriod))

	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr></w:p>`)
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">SATUAN INTERNAL AUDIT</w:t></w:r></w:p>`)
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(coverCompanyName)))
	body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Sistem Pengendalian Intern &amp; Audit Internal</w:t></w:r></w:p>`)

	// Page Break
	body.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)

	body.WriteString(buildLHAContent(report, st, wpHeader, wpRisks, wpCauses))

	// --- PENUTUP & SIGNATURES ---
	body.WriteString(`<w:p><w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">PENUTUP</w:t></w:r></w:p>`)
	conclusionText := xmlEsc(report.Conclusion)
	if conclusionText == "" {
		conclusionText = "Demikian Laporan Hasil Audit ini disampaikan untuk dapat dipergunakan sebagai bahan perbaikan tata kelola dan sistem pengendalian internal perusahaan secara berkelanjutan."
	}
	body.WriteString(fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, conclusionText))

	sigPlace := report.SignaturePlace
	if sigPlace == "" {
		sigPlace = "Jakarta"
	}
	sigDateStr := escapedDateStr
	if report.SignatureDate != nil {
		sigDateStr = report.SignatureDate.Format("02 January 2006")
	}
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:spacing w:before="360"/><w:jc w:val="right"/></w:pPr><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%s, %s</w:t></w:r></w:p>`, xmlEsc(sigPlace), sigDateStr))
	body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:spacing w:after="240"/><w:jc w:val="right"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(auditInternalLine)))

	body.WriteString(`<w:p><w:pPr><w:spacing w:before="120" w:after="240"/><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="22"/></w:rPr><w:t>LEMBAR PENGESAHAN TIM AUDIT</w:t></w:r></w:p>`)

	if len(sigItems) > 0 {
		body.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="9600" w:type="dxa"/><w:tblBorders><w:top w:val="none"/><w:left w:val="none"/><w:bottom w:val="none"/><w:right w:val="none"/><w:insideH w:val="none"/><w:insideV w:val="none"/></w:tblBorders><w:tblCellMar><w:top w:w="120" w:type="dxa"/><w:bottom w:w="120" w:type="dxa"/><w:left w:w="120" w:type="dxa"/><w:right w:w="120" w:type="dxa"/></w:tblCellMar></w:tblPr>`)

		chunkSize := 3
		for i := 0; i < len(sigItems); i += chunkSize {
			end := i + chunkSize
			if end > len(sigItems) {
				end = len(sigItems)
			}
			chunk := sigItems[i:end]
			cellWidth := 9600 / len(chunk)

			body.WriteString("<w:tr>")
			for _, m := range chunk {
				body.WriteString(fmt.Sprintf(`<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr>`, cellWidth))

				// 1. Role
				body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:after="80"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="444444"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(m.Role)))

				// 2. Signature Drawing / Line
				if m.HasImage {
					body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="60" w:after="60"/></w:pPr><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="1440000" cy="720000"/><wp:effectExtent l="0" t="0" r="0" b="0"/><wp:docPr id="%d" name="Sig%d"/><wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:nvPicPr><pic:cNvPr id="%d" name="Sig%d"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="1440000" cy="720000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`, m.DocPrID, m.DocPrID, m.DocPrID, m.DocPrID, m.RelID))
				} else {
					body.WriteString(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="300" w:after="300"/></w:pPr><w:r><w:rPr><w:color w:val="888888"/><w:i/><w:sz w:val="18"/></w:rPr><w:t xml:space="preserve">( Tanda Tangan )</w:t></w:r></w:p>`)
				}

				// 3. Member Name (Underlined, bold)
				body.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="60"/></w:pPr><w:r><w:rPr><w:b/><w:u w:val="single"/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, xmlEsc(m.Name)))

				body.WriteString("</w:tc>")
			}
			body.WriteString("</w:tr>")
		}
		body.WriteString("</w:tbl>")
	}

	// Return document.xml
	docXml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"
            xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math"
            xmlns:v="urn:schemas-microsoft-com:vml"
            xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
            xmlns:w10="urn:schemas-microsoft-com:office:word"
            xmlns:sl="http://schemas.openxmlformats.org/schemaLibrary/2006/main"
            xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
            xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">
  <w:body>
    %s
  </w:body>
</w:document>`, body.String())

	_ = time.Now()
	return docXml, nil
}

func formatTestResult(val any) string {
	if val == nil {
		return "-"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "Pass"
		}
		return "Fail"
	case string:
		l := strings.ToLower(strings.TrimSpace(v))
		if l == "pass" || l == "true" {
			return "Pass"
		}
		if l == "fail" || l == "false" {
			return "Fail"
		}
		if v == "" {
			return "-"
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isSampleDocEffective(s models.SampleDoc) bool {
	isFail := func(val any) bool {
		if val == nil {
			return false
		}
		switch v := val.(type) {
		case bool:
			return !v
		case string:
			l := strings.ToLower(strings.TrimSpace(v))
			return l == "fail" || l == "false"
		default:
			return false
		}
	}
	return !isFail(s.L1) && !isFail(s.L2) && !isFail(s.L3)
}

