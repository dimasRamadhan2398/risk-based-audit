package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPredictFutureTrends(t *testing.T) {
	service := NewAnalyticsService()
	result := service.PredictFutureTrends()

	if result.ModelAccuracy != 0.87 {
		t.Errorf("Expected model accuracy 0.87, got %v", result.ModelAccuracy)
	}

	if len(result.Forecast) != 3 {
		t.Errorf("Expected 3 forecast points, got %d", len(result.Forecast))
	}

	if result.TrendDirection != "Up" && result.TrendDirection != "Down" && result.TrendDirection != "Stable" {
		t.Errorf("Expected valid TrendDirection, got %v", result.TrendDirection)
	}
}

func TestGenerateReport(t *testing.T) {
	service := NewAnalyticsService()
	report := service.GenerateReport()

	if report.TotalFindings != 142 {
		t.Errorf("Expected total findings 142, got %d", report.TotalFindings)
	}

	if len(report.FindingTrends) != 6 {
		t.Errorf("Expected 6 trend points, got %d", len(report.FindingTrends))
	}
}

func TestCAATTOffline_DoesNotReturnDummyFallback(t *testing.T) {
	// Closed server simulates Data Hub offline without touching localhost:8100
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	t.Setenv("DATA_HUB_URL", deadURL)
	service := NewAnalyticsService()

	benford, err := service.GetCAATTBenford("ext-src-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(benford.Data) != 0 {
		t.Fatalf("expected 0 dummy rows when Data Hub offline, got %d", len(benford.Data))
	}
	if benford.Summary["data_hub_offline"] != true || benford.Summary["conclusion"] != "DATA_HUB_OFFLINE" {
		t.Fatalf("expected offline summary marker, got %v", benford.Summary)
	}

	fpt, err := service.GetCAATTFullPopulation("ext-src-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fpt.Data) != 0 {
		t.Fatalf("expected 0 dummy FPT rows when Data Hub offline, got %d", len(fpt.Data))
	}
}

func TestCAATTOnline_ForwardsSourceIDAndBenfordFindings(t *testing.T) {
	var seenQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"digit":                5,
					"actual_count":         420,
					"expected_count":       237.5,
					"count_diff":           182.5,
					"actual_pct":           14.0,
					"expected_pct":         7.9,
					"deviation_pct":        6.1,
					"chi_square":           140.2,
					"chi_square_share_pct": 88.4,
					"is_significant":       true,
					"dominant_branch":      "Cabang Surabaya",
					"finding_origin":       "Kelebihan +182 transaksi di Cabang Surabaya",
				},
			},
			"findings": []map[string]interface{}{
				{"trx_ref_number": "TRX-REAL-501", "digit": 5, "amount": 50000000.0, "branch_name": "Cabang Surabaya"},
			},
			"summary": map[string]interface{}{
				"sample_size":            3000,
				"total_chi_square":       158.6,
				"critical_value":         15.507,
				"significant_deviations": 1,
				"conclusion":             "ANOMALY DETECTED",
			},
		})
	}))
	defer srv.Close()

	t.Setenv("DATA_HUB_URL", srv.URL)
	service := NewAnalyticsService()

	res, err := service.GetCAATTBenford("real-source-uuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seenQuery != "source_id=real-source-uuid" {
		t.Fatalf("expected source_id forwarded in query, got %q", seenQuery)
	}
	if len(res.Data) != 1 || len(res.Findings) != 1 {
		t.Fatalf("expected real benford data and findings forwarded, got data=%d findings=%d", len(res.Data), len(res.Findings))
	}
	if res.Findings[0]["trx_ref_number"] != "TRX-REAL-501" {
		t.Fatalf("unexpected finding payload: %v", res.Findings[0])
	}

	resFiltered, err := service.GetCAATTBenfordWithParams("real-source-uuid", "5", "100")
	if err != nil {
		t.Fatalf("unexpected error on filtered call: %v", err)
	}
	if seenQuery != "digit=5&findings_limit=100&source_id=real-source-uuid" && seenQuery != "source_id=real-source-uuid&digit=5&findings_limit=100" {
		t.Fatalf("expected digit & findings_limit forwarded in query, got %q", seenQuery)
	}
	if len(resFiltered.Findings) != 1 {
		t.Fatalf("expected findings on filtered call, got %d", len(resFiltered.Findings))
	}
}

func TestTriggerAutoRetrain_ForceBypassesThrottleAndInvalidatesCache(t *testing.T) {
	calls := 0
	var lastQuery string
	pySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/retrain/status" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      "completed",
				"is_running":  false,
				"trained_auditor_hash": "abc123",
			})
			return
		}
		calls++
		lastQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "completed",
			"is_running": false,
		})
	}))
	defer pySrv.Close()

	t.Setenv("PYTHON_AI_URL", pySrv.URL)
	kv := newFakeStore()
	kv.data[cacheKeyPrefix+"batch:performance-trend"] = []byte(`{"cached":true}`)

	service := &AnalyticsService{
		aiClient: NewPythonAIClient(),
		cache: &AnalyticsCache{
			store:           kv,
			batchTTL:        15 * time.Minute,
			retrainInterval: 6 * time.Hour,
		},
	}

	// 1st background auto-retrain succeeds and claims throttle lock
	if _, err := service.TriggerAutoRetrain(); err != nil {
		t.Fatalf("unexpected auto-retrain error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 python call on first auto retrain, got %d", calls)
	}

	// Simulate batch cache re-populated and re-claimed lock
	kv.data[cacheKeyPrefix+"retrain:auto:last"] = []byte("locked")
	kv.data[cacheKeyPrefix+"batch:performance-trend"] = []byte(`{"cached":true}`)

	// 2nd non-force auto-retrain is skipped by 6h cooldown
	skipRes, err := service.TriggerAutoRetrain()
	if err != nil {
		t.Fatalf("unexpected error on throttled call: %v", err)
	}
	if m, ok := skipRes.(map[string]string); !ok || m["status"] != "skipped" {
		t.Fatalf("expected throttled skip, got %T %v", skipRes, skipRes)
	}
	if calls != 1 {
		t.Fatalf("expected throttled call not to hit python, got %d calls", calls)
	}

	// Manual/Force retrain MUST bypass cooldown, wait synchronously, and invalidate batch cache
	forceRes, err := service.TriggerAutoRetrainWithOptions(true, map[string]interface{}{"force": true, "wait": true})
	if err != nil {
		t.Fatalf("unexpected force retrain error: %v", err)
	}
	if calls != 2 || lastQuery != "force=true&wait=true" {
		t.Fatalf("expected force=true&wait=true call to python, calls=%d query=%q", calls, lastQuery)
	}
	if m, ok := forceRes.(map[string]interface{}); !ok || m["status"] != "completed" {
		t.Fatalf("expected completed response, got %v", forceRes)
	}
	if _, exists := kv.data[cacheKeyPrefix+"batch:performance-trend"]; exists {
		t.Fatalf("expected batch cache invalidated after force retrain")
	}

	// Retrain status endpoint forwards status
	st, err := service.GetRetrainStatus("auditor_hash=abc123")
	if err != nil {
		t.Fatalf("unexpected status error: %v", err)
	}
	if m, ok := st.(map[string]interface{}); !ok || m["trained_auditor_hash"] != "abc123" {
		t.Fatalf("unexpected retrain status: %v", st)
	}
}

func TestTriggerModelRetrain_PerModelCacheInvalidation(t *testing.T) {
	var seenPath string
	pySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "queued",
			"model":           "anomaly",
			"models_enqueued": []string{"anomaly"},
			"job_id":          "job-anomaly-123",
			"queue_position":  1,
		})
	}))
	defer pySrv.Close()

	t.Setenv("PYTHON_AI_URL", pySrv.URL)
	kv := newFakeStore()
	kv.data[cacheKeyPrefix+"batch:anomaly"] = []byte(`{"cached":true}`)
	kv.data[cacheKeyPrefix+"batch:risk-score"] = []byte(`{"cached":true}`)
	kv.data[cacheKeyPrefix+"batch:performance-trend"] = []byte(`{"cached":true}`)

	service := &AnalyticsService{
		aiClient: NewPythonAIClient(),
		cache: &AnalyticsCache{
			store:           kv,
			batchTTL:        15 * time.Minute,
			retrainInterval: 6 * time.Hour,
		},
	}

	res, err := service.TriggerModelRetrain("anomaly", map[string]interface{}{"force": true})
	if err != nil {
		t.Fatalf("unexpected TriggerModelRetrain error: %v", err)
	}
	if seenPath != "/retrain/model/anomaly" {
		t.Fatalf("expected /retrain/model/anomaly, got %s", seenPath)
	}
	if m, ok := res.(map[string]interface{}); !ok || m["status"] != "queued" {
		t.Fatalf("unexpected response: %v", res)
	}
	if _, exists := kv.data[cacheKeyPrefix+"batch:anomaly"]; exists {
		t.Fatalf("expected anomaly batch cache invalidated")
	}
	if _, exists := kv.data[cacheKeyPrefix+"batch:risk-score"]; !exists {
		t.Fatalf("expected risk-score batch cache NOT invalidated when retraining anomaly")
	}
	if _, exists := kv.data[cacheKeyPrefix+"batch:performance-trend"]; !exists {
		t.Fatalf("expected performance-trend batch cache NOT invalidated when retraining anomaly")
	}
}

func TestNLPAnalysisRiskLevel5TierEnforcement(t *testing.T) {
	allowedCodes := map[string]bool{
		"HIGH":          true,
		"MODERATE_HIGH": true,
		"MODERATE":      true,
		"LOW_MODERATE":  true,
		"LOW":           true,
	}

	// 1. Offline fraud inference must NEVER return CRITICAL; strictly maps to HIGH
	_, _, _, _, _, _, fraudLevel := inferDocumentMetadata("Laporan Fraud Kas", "fraud.pdf", "Ditemukan transaksi fiktif", "")
	if fraudLevel != "HIGH" {
		t.Fatalf("expected HIGH for fraud metadata fallback, got %q", fraudLevel)
	}

	// 2. Canonical Go risk normalization eliminates external/legacy non-5-tier labels
	cases := map[string]string{
		"Critical":         "HIGH",
		"CRITICAL":         "HIGH",
		"Extreme":          "HIGH",
		"Very High":        "HIGH",
		"Major":            "MODERATE_HIGH",
		"Medium to High":   "MODERATE_HIGH",
		"KOL-3":            "MODERATE_HIGH",
		"Medium":           "MODERATE",
		"Watch":            "MODERATE",
		"Minor":            "LOW_MODERATE",
		"Low to Medium":    "LOW_MODERATE",
		"KOL-2":            "LOW_MODERATE",
		"Insignificant":    "LOW",
		"Safe":             "LOW",
		"KOL-1":            "LOW",
		"Not Scored":       "MODERATE",
	}
	for raw, expected := range cases {
		code, label := normalizeCanonicalRiskLevelGo(raw, "MODERATE")
		if code != expected {
			t.Fatalf("normalizeCanonicalRiskLevelGo(%q) = %q (%q), expected %q", raw, code, label, expected)
		}
		if !allowedCodes[code] {
			t.Fatalf("unexpected non-canonical code %q", code)
		}
	}

	// 3. Offline AnalyzeDocumentOrText adapts external DB riskLevel/severity strictly to 5 levels
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close()
	t.Setenv("PYTHON_AI_URL", deadURL)
	service := NewAnalyticsService()

	for raw, exp := range map[string]string{"Extreme": "HIGH", "Major": "MODERATE_HIGH", "Medium": "MODERATE", "Minor": "LOW_MODERATE", "Insignificant": "LOW"} {
		res, err := service.AnalyzeDocumentOrText(TextRequest{
			Title:      "External Record " + raw,
			Text:       "Temuan eksternal atas proses pengendalian internal cabang.",
			SourceType: "External Database",
			RiskLevel:  raw,
		})
		if err != nil {
			t.Fatalf("unexpected fallback error: %v", err)
		}
		if res.RiskLevel != exp {
			t.Fatalf("expected fallback RiskLevel %q for external %q, got %q", exp, raw, res.RiskLevel)
		}
	}

	// 4. Online batch response sanitizes legacy CRITICAL/MEDIUM/Minor records from AI or cache
	pySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"documents": []map[string]interface{}{
				{"docId": "EXT-1", "title": "Critical Doc", "risk_level": "CRITICAL"},
				{"docId": "EXT-2", "title": "Major Doc", "risk_level": "Major"},
				{"docId": "EXT-3", "title": "Medium Doc", "risk_level": "MEDIUM"},
				{"docId": "EXT-4", "title": "Minor Doc", "risk_level": "Minor"},
				{"docId": "EXT-5", "title": "Safe Doc", "risk_level": "Insignificant"},
			},
		})
	}))
	defer pySrv.Close()
	t.Setenv("PYTHON_AI_URL", pySrv.URL)
	serviceOnline := NewAnalyticsService()
	batchRes, err := serviceOnline.GetTextAnalysisBatch()
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}
	m, ok := batchRes.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map response, got %T", batchRes)
	}
	docs := m["documents"].([]interface{})
	expectedOrder := []string{"HIGH", "MODERATE_HIGH", "MODERATE", "LOW_MODERATE", "LOW"}
	for idx, item := range docs {
		doc := item.(map[string]interface{})
		got := doc["risk_level"].(string)
		if got != expectedOrder[idx] {
			t.Fatalf("doc[%d] risk_level = %q, expected %q", idx, got, expectedOrder[idx])
		}
		if !allowedCodes[got] {
			t.Fatalf("non-canonical risk_level %q leaked in batch response", got)
		}
	}
}
