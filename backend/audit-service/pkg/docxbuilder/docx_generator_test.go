package docxbuilder_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/docxbuilder"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAuditReportDocx_WithUpdatedSampleDataAndWorkingPaper(t *testing.T) {
	now := time.Now()
	report := &models.AuditResultReport{
		ID:                 uuid.New(),
		ReportNumber:       "LHA-020/SKAI/2026",
		ReportTitle:        "Laporan Hasil Audit Operasional & Pengendalian",
		AssignmentLetterID: "ST-001/SKAI/2026",
		AuditPeriod:        "2026-03-01 s/d 2026-03-31",
		ReportDate:         &now,
		CompanyName:        "PT Test Company",
		PreparedBy:         "Zeta Ramadhani",
		ReviewedBy:         "Budi Santoso",
	}

	st := &models.AssignmentLetter{
		LetterNumber:    "ST-001/SKAI/2026",
		AuditTitle:      "Audit Pengendalian Kas & Operasional",
		WorkingUnit:     "Finance & Operations",
		ExecutionPeriod: "2026-03-01 to 2026-03-31",
		Leader:          "Zeta Ramadhani",
		AuditTeam:       "SKAI",
		MembersList: []models.LetterMember{
			{Name: "Zeta Ramadhani", Role: "Chairperson"},
			{Name: "Budi Santoso", Role: "Supervisor"},
		},
	}

	// 1. Audit Fieldwork: Sample Data with attached file
	fieldworkSamples := []models.FieldworkSample{
		{
			DocumentName:   "Bukti Pengeluaran Kas",
			DocumentNumber: "BPK-2026-001",
			Date:           "2026-03-10",
			Description:    "Uji petik sampel pengeluaran kas di atas 50 juta",
			FileName:       "bukti_pengeluaran_kas_001.pdf",
		},
	}

	// 2. Working Paper: Header with BusinessProcess, AuditPurpose, Activities, and TeamMembers
	wpHeader := &models.WorkingPaperHeader{
		AssignmentLetterID: "ST-001/SKAI/2026",
		AuditPurpose:       "Menilai kepatuhan prosedur pembayaran kas dan otorisasi manajemen.",
		BusinessProcess:    "Procurement-to-Pay",
		Period:             "2026-03-01 s/d 2026-03-31",
		Location:           "Kantor Pusat Jakarta",
		Activities: []models.ActivityItem{
			{ID: 1, Name: "Verifikasi dokumen voucher pembayaran kas"},
			{ID: 2, Name: "Konfirmasi otorisasi pejabat berwenang"},
		},
		TeamMembers: []models.TeamMember{
			{ID: 1, Name: "Zeta Ramadhani", Role: "Chairperson"},
			{ID: 2, Name: "Budi Santoso", Role: "Supervisor"},
			{ID: 3, Name: "Rina Wulandari", Role: "Member"},
		},
	}

	// 3. Working Paper: Test Sample with Step1, Step2, Step3, L1, L2, L3 and FieldworkDocument
	pop := "150 Transaksi"
	sampleSize := 10
	l1Pass := true
	l2Fail := false
	l3Pass := "Pass"

	wpSamples := []models.WorkingPaperSample{
		{
			WorkingPaperID: "ST-001/SKAI/2026",
			Population:     &pop,
			SampleSize:     &sampleSize,
			Conclusion:     "Terdapat 1 deviasi pada otorisasi level 2.",
			Samples: []models.SampleDoc{
				{
					ID:                1,
					FieldworkDocument: "Bukti Pengeluaran Kas (BPK-2026-001)",
					Document:          "Voucher Pembayaran V-101",
					Step1:             "Pengecekan Tanda Tangan Manager",
					L1:                l1Pass,
					Step2:             "Kesesuaian Lampiran Faktur Pajak",
					L2:                l2Fail,
					Step3:             "Validasi Pembukuan GL",
					L3:                l3Pass,
				},
			},
		},
	}

	docxBytes, err := docxbuilder.GenerateAuditReportDocx(
		report,
		st,
		nil,
		nil,
		nil,
		fieldworkSamples,
		wpHeader,
		nil,
		wpSamples,
		nil,
		nil,
		nil,
	)

	require.NoError(t, err)
	require.NotEmpty(t, docxBytes)

	// Unzip docx bytes and inspect word/document.xml
	zipReader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	require.NoError(t, err)

	var docXmlContent string
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			_ = rc.Close()
			docXmlContent = string(content)
			break
		}
	}

	require.NotEmpty(t, docXmlContent, "word/document.xml should be present in docx")

	// Verify Fieldwork Sample Data: Document Name, Number, Date, Description, and Attachment
	assert.True(t, strings.Contains(docXmlContent, "Bukti Pengeluaran Kas"))
	assert.True(t, strings.Contains(docXmlContent, "BPK-2026-001"))
	assert.True(t, strings.Contains(docXmlContent, "bukti_pengeluaran_kas_001.pdf"))

	// Verify Working Paper Header: Business Process, Audit Purpose, Period, Location, Activities, Team Members
	assert.True(t, strings.Contains(docXmlContent, "Procurement-to-Pay"))
	assert.True(t, strings.Contains(docXmlContent, "Menilai kepatuhan prosedur pembayaran kas"))
	assert.True(t, strings.Contains(docXmlContent, "Verifikasi dokumen voucher pembayaran kas"))
	assert.True(t, strings.Contains(docXmlContent, "Zeta Ramadhani (Chairperson)"))
	assert.True(t, strings.Contains(docXmlContent, "Rina Wulandari (Member)"))

	// Verify Working Paper Test Sample: Population, Sample Size, Conclusion, Step1, Step2, Step3, Results
	assert.True(t, strings.Contains(docXmlContent, "150 Transaksi"))
	assert.True(t, strings.Contains(docXmlContent, "Voucher Pembayaran V-101"))
	assert.True(t, strings.Contains(docXmlContent, "Bukti Pengeluaran Kas (BPK-2026-001)"))
	assert.True(t, strings.Contains(docXmlContent, "Tidak Efektif")) // because L2 is false
	assert.True(t, strings.Contains(docXmlContent, "Pengecekan Tanda Tangan Manager"))
	assert.True(t, strings.Contains(docXmlContent, "Pass"))
	assert.True(t, strings.Contains(docXmlContent, "Fail"))
}
