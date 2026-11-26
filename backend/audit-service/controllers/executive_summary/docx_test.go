package executivesummary_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/controllers/crud"
	executivesummary "audit-service/controllers/executive_summary"
	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupDocxTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t, &models.ExecutiveSummary{}, &models.AuditResultReport{})
}

func readDocumentXMLFromDocx(t *testing.T, data []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open docx zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open word/document.xml: %v", err)
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read word/document.xml: %v", err)
			}
			return string(b)
		}
	}
	t.Fatal("word/document.xml not found in docx")
	return ""
}

func TestDownloadDocxRequiresApprovedStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupDocxTestDB(t)

	draft := models.ExecutiveSummary{
		ID:           uuid.New(),
		NomorDokumen: "ESC-001/2026",
		Quarter:      1,
		Tahun:        2026,
		Status:       "Draft",
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("create draft: %v", err)
	}

	r := gin.New()
	r.GET("/api/v1/executive-summaries/:id/download-docx", executivesummary.DownloadDocxHandler(db))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executive-summaries/"+draft.ID.String()+"/download-docx", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for Draft summary, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "EXECUTIVE_SUMMARY_NOT_APPROVED") {
		t.Fatalf("expected EXECUTIVE_SUMMARY_NOT_APPROVED error code, got %s", w.Body.String())
	}
}

func TestDownloadDocxApprovedCompilationAndIndividual(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupDocxTestDB(t)

	comp := models.ExecutiveSummary{
		ID:                  uuid.New(),
		NomorDokumen:        "ESC-001/2026",
		Quarter:             1,
		PeriodeBulan:        "Kuartal I",
		Tahun:               2026,
		Status:              "Approved",
		JumlahLaporan:       3,
		RisikoHigh:          2,
		RisikoModerate:      3,
		RisikoLow:           1,
		JumlahRekomendasi:   8,
		FollowUpTable:       `[{"status":"Closed","jumlah":4,"persentase":66.7,"keterangan":"Telah divalidasi SPI"},{"status":"In Progress","jumlah":1,"persentase":16.7,"keterangan":"On track"},{"status":"Overdue","jumlah":1,"persentase":16.7,"keterangan":"1 Risiko Tinggi"}]`,
		TopFindings:         `[{"unitDivision":"TI","judulTemuan":"Celah Otentikasi Aplikasi RAS","risiko":"Very Significant","statusTL":"Overdue","usulan":"Eskalasi Direksi"}]`,
		MatriksKompilasi:    `[{"nomor":"001/SPI/2026","division":"OP","unitKerja":"Division Operation Personnel","prosesBisnis":"O&M","judulTemuan":"Kalibrasi Alat Ukur Gas Terlambat","nilaiRisiko":"Significant","rekomendasi":"Lakukan kalibrasi Q2","dueDate":"30/06/2026","picUnit":"Manager O&M","progres":60,"status":"In Progress","buktiTL":"BA Kalibrasi 1"}]`,
		AkarMasalah:         "Tema berulang periode ini: 1. Pengelolaan Aset Tetap; 2. Kepatuhan Kontrak.",
		Kesimpulan:          "Secara umum pengendalian membaik, indikator on-time di atas target.",
		SignatureTempat:     "Jakarta",
		SignatureTanggal:    "2026-03-31",
		SignatureNamaKepala: "Budi Santoso, CIA",
		SignatureNIK:        "SPI-77621",
	}
	if err := db.Create(&comp).Error; err != nil {
		t.Fatalf("create compilation: %v", err)
	}

	lha := models.AuditResultReport{
		ID:                 uuid.New(),
		ReportNumber:       "LHA-001/SKAI/2026",
		AssignmentLetterID: "ST-001/SKAI/2026",
		ReportTitle:        "Laporan Hasil Audit Operasional Keuangan",
		AuditObject:        "Operasional Keuangan",
		Department:         "Finance (FIN)",
		Status:             "Final",
	}
	if err := db.Create(&lha).Error; err != nil {
		t.Fatalf("create lha: %v", err)
	}

	ind := models.ExecutiveSummary{
		ID:                  uuid.New(),
		AssignmentLetterID:  "ST-001/SKAI/2026",
		NomorDokumen:        "LHA-001/SKAI/2026",
		Quarter:             1,
		PeriodeBulan:        "Januari – Februari 2026",
		Tahun:               2026,
		Status:              "Approved",
		JumlahLaporan:       1,
		RisikoHigh:          1,
		RisikoModerate:      1,
		RisikoLow:           0,
		JumlahRekomendasi:   2,
		TopFindings:         `[{"unitDivision":"Finance (FIN)","judulTemuan":"Selisih Rekonsiliasi Kas","risiko":"Significant","statusTL":"In Progress","usulan":"Perbaikan SOP"}]`,
		AkarMasalah:         "Kelemahan monitoring rekonsiliasi.",
		Kesimpulan:          "Pengendalian memadai dengan tindak lanjut rekonsiliasi.",
		SignatureTempat:     "Jakarta",
		SignatureTanggal:    "2026-03-15",
		SignatureNamaKepala: "Budi Santoso, CIA",
		SignatureNIK:        "SPI-77621",
	}
	if err := db.Create(&ind).Error; err != nil {
		t.Fatalf("create individual: %v", err)
	}

	r := gin.New()
	r.GET("/api/v1/executive-summaries/:id/download-docx", executivesummary.DownloadDocxHandler(db))

	// 1. Compilation DOCX
	wComp := httptest.NewRecorder()
	reqComp := httptest.NewRequest(http.MethodGet, "/api/v1/executive-summaries/"+comp.ID.String()+"/download-docx", nil)
	r.ServeHTTP(wComp, reqComp)

	if wComp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Approved Compilation, got %d: %s", wComp.Code, wComp.Body.String())
	}
	compXML := readDocumentXMLFromDocx(t, wComp.Body.Bytes())
	for _, want := range []string{
		"EXECUTIVE SUMMARY COMPILATION",
		"HASIL AUDIT TRIWULANAN",
		"II. STATISTIK KOMPILASI",
		"III. STATUS TINDAK LANJUT",
		"IV. TOP 5 TEMUAN SIGNIFIKAN",
		"V. ANALISIS TEMUAN BERULANG",
		"VI. TREN &amp; GRAFIK",
		"VII. KESIMPULAN &amp; REKOMENDASI UNTUK MANAJEMEN",
		"MATRIKS INDUK KOMPILASI TEMUAN",
		"Kalibrasi Alat Ukur Gas Terlambat",
		"Budi Santoso, CIA",
	} {
		if !strings.Contains(compXML, want) {
			t.Errorf("compilation docx missing %q", want)
		}
	}

	// 2. Individual DOCX
	wInd := httptest.NewRecorder()
	reqInd := httptest.NewRequest(http.MethodGet, "/api/v1/executive-summaries/"+ind.ID.String()+"/download-docx", nil)
	r.ServeHTTP(wInd, reqInd)

	if wInd.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Approved Individual, got %d: %s", wInd.Code, wInd.Body.String())
	}
	indXML := readDocumentXMLFromDocx(t, wInd.Body.Bytes())
	for _, want := range []string{
		"EXECUTIVE SUMMARY INDIVIDUAL",
		"LAPORAN HASIL AUDIT (LHA)",
		"Nomor LHA: LHA-001/SKAI/2026",
		"II. INFORMASI &amp; STATISTIK LHA INDIVIDUAL",
		"Laporan Hasil Audit Operasional Keuangan",
		"Selisih Rekonsiliasi Kas",
	} {
		if !strings.Contains(indXML, want) {
			t.Errorf("individual docx missing %q", want)
		}
	}
	if strings.Contains(indXML, "MATRIKS INDUK KOMPILASI TEMUAN") {
		t.Error("individual docx must not include the compilation matrix attachment")
	}
}

func TestCreateExecutiveSummaryAcceptsEmptyApprovedAtAndPersists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupDocxTestDB(t)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-1")
		c.Set("username", "admin")
		c.Set("roles", []string{"ADMIN"})
		c.Next()
	})
	g := r.Group("/api/v1/executive-summaries")
	g.GET("", crud.ListWithItems(db, "ExecutiveSummary", func() interface{} { return &[]models.ExecutiveSummary{} }, executivesummary.ListDecorator))
	g.POST("", crud.CreateWithHook(db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, executivesummary.GuardCreateHook))
	g.POST("/:id/approve", executivesummary.ApproveHandler(db))

	// Simulate frontend payload that sends empty strings for timestamp fields
	body := `{"id":"","nomorDokumen":"ESC-001/2026","quarter":1,"periodeBulan":"Kuartal I","tahun":2026,"status":"Draft","approvedAt":"","created_at":"","updated_at":"","narrative":"Test narrative"}`
	wCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/executive-summaries", strings.NewReader(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created even when approvedAt is empty string, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createResp struct {
		Data models.ExecutiveSummary `json:"data"`
	}
	if err := json.Unmarshal(wCreate.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Data.ID == uuid.Nil {
		t.Fatal("expected backend to assign a valid UUID")
	}

	// Verify it persists on refresh (GET list)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, httptest.NewRequest(http.MethodGet, "/api/v1/executive-summaries?page=1&page_size=100", nil))
	if wList.Code != http.StatusOK || !strings.Contains(wList.Body.String(), "ESC-001/2026") {
		t.Fatalf("expected saved compilation summary to persist in GET list, got %d: %s", wList.Code, wList.Body.String())
	}

	// Verify approve populates signature_nama_kepala from signerName
	png := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
	approveBody := `{"signaturePlace":"Jakarta","signatureDate":"2026-03-31","signatureImage":"` + png + `","signerName":"Budi Santoso","signerRole":"Kepala SPI"}`
	wApprove := httptest.NewRecorder()
	reqApprove := httptest.NewRequest(http.MethodPost, "/api/v1/executive-summaries/"+createResp.Data.ID.String()+"/approve", strings.NewReader(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wApprove, reqApprove)
	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on approve, got %d: %s", wApprove.Code, wApprove.Body.String())
	}
	var saved models.ExecutiveSummary
	if err := db.First(&saved, "id = ?", createResp.Data.ID).Error; err != nil {
		t.Fatalf("fetch approved summary: %v", err)
	}
	if saved.Status != "Approved" || saved.SignatureNamaKepala != "Budi Santoso" {
		t.Fatalf("expected Approved with SignatureNamaKepala=Budi Santoso, got status=%q nama=%q", saved.Status, saved.SignatureNamaKepala)
	}
}
