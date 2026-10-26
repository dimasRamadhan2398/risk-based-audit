package docxbuilder_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"audit-service/models"
	"audit-service/pkg/docxbuilder"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// An unresolved company name must not be replaced by a hardcoded company: the
// cover shows "-" and the signature line is just "AUDIT INTERNAL".
func TestGenerateAuditReportDocx_EmptyCompanyNameHasNoHardcodedFallback(t *testing.T) {
	report := &models.AuditResultReport{
		ID:           uuid.New(),
		ReportNumber: "LHA-002/SKAI/2026",
		ReportTitle:  "Laporan",
	}
	st := &models.AssignmentLetter{LetterNumber: "ST-002/SKAI/2026"}

	docxBytes, err := docxbuilder.GenerateAuditReportDocx(report, st, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	require.NoError(t, err)
	var doc string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			require.NoError(t, err)
			b, err := io.ReadAll(rc)
			require.NoError(t, err)
			_ = rc.Close()
			doc = string(b)
		}
	}
	require.NotEmpty(t, doc)

	require.NotContains(t, strings.ToUpper(doc), "AIFL")
	require.Contains(t, doc, `<w:t xml:space="preserve">-</w:t>`, "cover company line should be '-'")
	require.Contains(t, doc, `<w:t xml:space="preserve">AUDIT INTERNAL</w:t>`, "signature line without a company")
}
