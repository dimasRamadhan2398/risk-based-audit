package docxbuilder_test

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"audit-service/models"
	"audit-service/pkg/docxbuilder"

	"github.com/stretchr/testify/require"
)

func TestCategoryLabelID(t *testing.T) {
	cases := map[string]string{
		"Very Significant":   "Sangat Signifikan",
		"significant":        "Signifikan",
		" Quite Significant": "Cukup Signifikan",
		"NOT_SIGNIFICANT":    "Tidak Signifikan",
		"Major":              "Major",
		"":                   "-",
	}
	for in, want := range cases {
		require.Equal(t, want, docxbuilder.CategoryLabelID(in), in)
	}
}

func docXML(t *testing.T, b []byte) string {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	require.NoError(t, err)
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			require.NoError(t, err)
			defer rc.Close()
			data, err := io.ReadAll(rc)
			require.NoError(t, err)
			return string(data)
		}
	}
	t.Fatal("no document.xml")
	return ""
}

func TestGenerateAuditReportDocx_FindingTables(t *testing.T) {
	report := &models.AuditResultReport{
		ReportNumber: "LHA-001/SKAI/2026",
		Findings: []models.AuditReportFinding{
			{
				Title: "Temuan <A> & B", Category: "Very Significant",
				RootCause: []models.RootCauseItem{{Method: "process", W1: "baris1\nbaris2", W2: "w2 & x", W3: "w3"}},
				ActionPlans: []models.FindingActionPlan{
					{Action: "Perbaiki SOP", TargetDate: "2026-12-31", PIC: "Budi", FinalApprover: "Kepala", Evidence: "SOP v2"},
					{Action: "Training"},
				},
			},
			{Title: "Legacy", Category: "Major", Action: "Legacy action"},
		},
	}
	causes := []models.WorkingPaperCause{{Condition: "Kondisi X", RootCause: []models.RootCauseItem{{Method: "People", W1: "a", W2: "b", W3: "c"}}}}
	b, err := docxbuilder.GenerateAuditReportDocx(report, &models.AssignmentLetter{}, nil, nil, nil, nil, nil, nil, nil, causes, nil, nil)
	require.NoError(t, err)
	doc := docXML(t, b)

	for _, want := range []string{
		"Why-1", "Why-2", "Why-3", "Kategori Temuan: Sangat Signifikan", "Kategori Temuan: Major",
		"Rencana Aksi (Setelah disepakati dengan auditi) / Rencana Tindak Lanjut",
		">Rencana Aksi<", ">Target Penyelesaian<", ">PIC<", ">Final Approver<", ">Evidence<",
		"Temuan &lt;A&gt; &amp; B", "baris2", "Perbaiki SOP", "Legacy action", "1. Surat Tugas", "2. Obyek Audit", "3. Risiko yang diaudit", "4. Ruang Lingkup Audit", "5. Periode Audit", "6. Tanggal Audit", "7. HASIL (TEMUAN) AUDIT PERIODE INI",
		">Eksternal<", ">Policy<",
	} {
		require.Contains(t, doc, want)
	}
	require.NotContains(t, doc, "Tindak Lanjut / Action")

	// well-formed XML
	dec := xml.NewDecoder(strings.NewReader(doc))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
}
