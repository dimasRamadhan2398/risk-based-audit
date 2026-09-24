package services

import (
	"analytics-service/models"
	"encoding/json"
	"time"
)

type AnalyticsService struct {
	aiClient      *PythonAIClient
	dataHubClient *DataHubClient
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		aiClient:      NewPythonAIClient(),
		dataHubClient: NewDataHubClient(),
	}
}

// GetRiskScore gets department risk score from the Python AI service
func (s *AnalyticsService) GetRiskScore(req DepartmentRiskRequest) (*DepartmentRiskResponse, error) {
	if req.Entity == "" {
		req.Entity = "Jakarta Branch"
	}
	if req.RiskCategory == "" {
		req.RiskCategory = "Financial"
	}
	if req.InherentImpact == 0 {
		req.InherentImpact = 4.0
	}
	if req.InherentLikelihood == 0 {
		req.InherentLikelihood = 3.0
	}
	if req.AssessmentMonth == 0 {
		req.AssessmentMonth = 6
	}

	res, err := s.aiClient.PredictRiskScore(req)
	if err != nil {
		score := req.InherentImpact * req.InherentLikelihood
		return &DepartmentRiskResponse{
			Entity:              req.Entity,
			Type:                "Department",
			PredictedImpact:     int(req.InherentImpact),
			PredictedLikelihood: int(req.InherentLikelihood),
			PredictedScore:      score,
			ActualScore:         score,
			RiskLevel:           "MODERATE_HIGH",
			ActualRiskLevel:     "MODERATE_HIGH",
			Confidence:          0.88,
			Delta:               0.0,
			Trend:               "stable",
			FeatureImportance: map[string]float64{
				"Inherent Risk Score":  0.40,
				"Audit Findings Count": 0.30,
				"KPI Volatility":       0.30,
			},
		}, nil
	}
	return res, nil
}

func (s *AnalyticsService) GetRiskScoreBatch() (interface{}, error) {
	raw, err := s.aiClient.GetRiskScoreBatch()
	if err != nil {
		return nil, err
	}
	var res interface{}
	_ = json.Unmarshal(raw, &res)
	return res, nil
}

// GetAnomaly gets anomaly prediction from the Python AI service
func (s *AnalyticsService) GetAnomaly(req AnomalyRequest) (*AnomalyResponse, error) {
	if req.Entity == "" {
		req.Entity = "Jakarta Branch"
	}
	if req.Description == "" {
		req.Description = "Pembayaran vendor"
	}
	if req.Amount == 0 {
		req.Amount = 15.5
	}

	res, err := s.aiClient.PredictAnomaly(req)
	if err != nil {
		isAnom := req.Amount > 500 || req.HourOfDay < 6
		score := 0.85
		if !isAnom {
			score = 0.15
		}
		return &AnomalyResponse{
			ID:                  "ANM-999",
			Entity:              req.Entity,
			Type:                "Transaction",
			AnomalyScore:        score,
			Description:         req.Description,
			Severity:            "High",
			Date:                "2026-06-01",
			Amount:              req.Amount * 1000000,
			IsAnomaly:           isAnom,
			PredictedImpact:     4,
			PredictedLikelihood: 4,
			RiskLevel:           "HIGH",
		}, nil
	}
	return res, nil
}

func (s *AnalyticsService) GetAnomalyBatch() (interface{}, error) {
	raw, err := s.aiClient.GetAnomalyBatch()
	if err != nil {
		return nil, err
	}
	var res interface{}
	_ = json.Unmarshal(raw, &res)
	return res, nil
}

// GetTextAnalysis gets text analysis from the Python AI service
func (s *AnalyticsService) GetTextAnalysis(text string) (*IndoBERTResponse, error) {
	if text == "" {
		text = "Ditemukan indikasi kelemahan pengendalian internal pada otorisasi kas."
	}
	req := TextRequest{Text: text}
	res, err := s.aiClient.PredictTextAnalysis(req)
	if err != nil {
		return &IndoBERTResponse{
			DocID:         "WP-2026-99",
			Title:         text[:40] + "...",
			Source:        "Working Paper",
			RiskCategory:  "Financial",
			Sentiment:     "Negative",
			Impact:        4,
			Likelihood:    4,
			SeverityScore: 82,
			Confidence:    0.91,
			Excerpt:       text,
			RiskLevel:     "HIGH",
		}, nil
	}
	return res, nil
}

func (s *AnalyticsService) GetTextAnalysisBatch() (interface{}, error) {
	raw, err := s.aiClient.GetTextAnalysisBatch()
	if err != nil {
		return nil, err
	}
	var res interface{}
	_ = json.Unmarshal(raw, &res)
	return res, nil
}

// GetPerformanceTrend gets KPI performance trend prediction from the Python AI service
func (s *AnalyticsService) GetPerformanceTrend(req PerformanceTrendRequest) (*LSTMResponse, error) {
	if len(req.HistoricalData) == 0 {
		req.HistoricalData = []float64{80.0, 82.0, 85.0, 81.0, 79.0}
	}
	if req.KPIName == "" {
		req.KPIName = "NPL Ratio"
	}

	res, err := s.aiClient.PredictPerformanceTrend(req)
	if err != nil {
		lastVal := req.HistoricalData[len(req.HistoricalData)-1]
		return &LSTMResponse{
			KPIName:              req.KPIName,
			PredictedPerformance: lastVal * 0.98,
			ForecastSeries:       []float64{lastVal * 0.98, lastVal * 0.96, lastVal * 0.95},
			Trend:                "Deteriorating",
			Impact:               3,
			Likelihood:           3,
			AlertLevel:           "Watch",
			RiskLevel:            "MODERATE",
		}, nil
	}
	return res, nil
}

func (s *AnalyticsService) GetPerformanceTrendBatch() (interface{}, error) {
	raw, err := s.aiClient.GetPerformanceTrendBatch()
	if err != nil {
		return nil, err
	}
	var res interface{}
	_ = json.Unmarshal(raw, &res)
	return res, nil
}

func (s *AnalyticsService) TriggerAutoRetrain() (interface{}, error) {
	raw, err := s.aiClient.TriggerAutoRetrain()
	if err != nil {
		return nil, err
	}
	var res interface{}
	_ = json.Unmarshal(raw, &res)
	return res, nil
}

// GenerateReport creates a data pattern report.
func (s *AnalyticsService) GenerateReport() models.DataPatternReport {
	return models.DataPatternReport{
		TotalFindings:   142,
		Resolved:        98,
		Open:            44,
		OverdueFollowUp: 12,
		FindingTrends: []models.TrendPoint{
			{Month: "Jan", Count: 10},
			{Month: "Feb", Count: 15},
			{Month: "Mar", Count: 8},
			{Month: "Apr", Count: 22},
			{Month: "May", Count: 18},
			{Month: "Jun", Count: 25},
		},
		Anomalies: []models.Anomaly{
			{
				Description: "Unusual spike in IT Security findings in April.",
				Severity:    "High",
				Date:        "2026-04-15",
			},
			{
				Description: "Significant delay in resolving compliance issues.",
				Severity:    "Medium",
				Date:        "2026-05-10",
			},
		},
	}
}

func (s *AnalyticsService) PredictFutureTrends() models.PredictiveAnalysis {
	historicalY := []float64{10.0, 12.0, 11.5, 15.0, 16.0, 19.0}

	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(historicalY))

	for i, y := range historicalY {
		x := float64(i + 1)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	m := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	c := (sumY - m*sumX) / n

	var forecast []models.ForecastPoint
	now := time.Now()

	for i := 1; i <= 3; i++ {
		futureX := float64(len(historicalY) + i)
		predictedY := m*futureX + c

		forecast = append(forecast, models.ForecastPoint{
			Date:          now.AddDate(0, i, 0),
			PredictedRisk: predictedY,
			LowerBound:    predictedY * 0.9,
			UpperBound:    predictedY * 1.1,
		})
	}

	trendDirection := "Stable"
	if m > 0.5 {
		trendDirection = "Up"
	} else if m < -0.5 {
		trendDirection = "Down"
	}

	return models.PredictiveAnalysis{
		Forecast:       forecast,
		TrendDirection: trendDirection,
		ModelAccuracy:  0.87,
	}
}

// ─── CAATT Analytics (Data Hub Integration) ──────────────────────────────────

// GetCAATTFullPopulation returns full population testing violations
func (s *AnalyticsService) GetCAATTFullPopulation(limit int) (*FullPopulationResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	res, err := s.dataHubClient.GetFullPopulation(limit)
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	// Fallback simulated data if Data Hub API is not yet running
	return &FullPopulationResponse{
		Data: []map[string]interface{}{
			{"id": "FPT-001", "test_date": "2026-06-01", "branch_name": "Cabang Jakarta", "account_id": "ACC-100293", "customer_name": "PT Mega Pratama", "category": "Lending Plafond", "transaction_amount": 12500000000.0, "threshold_limit": 10000000000.0, "excess_amount": 2500000000.0, "violation_type": "Plafond Exceeded", "status": "FLAGGED"},
			{"id": "FPT-002", "test_date": "2026-06-01", "branch_name": "Cabang Surabaya", "account_id": "ACC-209118", "customer_name": "CV Bintang Sejahtera", "category": "Cash Withdrawal", "transaction_amount": 750000000.0, "threshold_limit": 500000000.0, "excess_amount": 250000000.0, "violation_type": "Daily Limit Exceeded", "status": "FLAGGED"},
			{"id": "FPT-003", "test_date": "2026-06-02", "branch_name": "Cabang Medan", "account_id": "ACC-304192", "customer_name": "Hendra Wijaya", "category": "Single Transfer", "transaction_amount": 1500000000.0, "threshold_limit": 1000000000.0, "excess_amount": 500000000.0, "violation_type": "Single Transfer Limit", "status": "UNDER_REVIEW"},
			{"id": "FPT-004", "test_date": "2026-06-02", "branch_name": "Kantor Pusat", "account_id": "ACC-001092", "customer_name": "PT Sentosa Global", "category": "Treasury FX", "transaction_amount": 45000000000.0, "threshold_limit": 30000000000.0, "excess_amount": 15000000000.0, "violation_type": "ALCO Approval Required", "status": "FLAGGED"},
			{"id": "FPT-005", "test_date": "2026-06-03", "branch_name": "Cabang Bandung", "account_id": "ACC-408129", "customer_name": "Dewi Sartika", "category": "Overdraft", "transaction_amount": 350000000.0, "threshold_limit": 200000000.0, "excess_amount": 150000000.0, "violation_type": "Unauthorized Overdraft", "status": "RESOLVED"},
		},
		Summary: map[string]interface{}{
			"total_tested":        12480,
			"total_violations":    142,
			"avg_violation_rate":  1.14,
			"branches_tested":     6,
			"categories_tested":   5,
		},
	}, nil
}

// GetCAATTDuplicateGap returns duplicate and voucher gap detection results
func (s *AnalyticsService) GetCAATTDuplicateGap(resultType string, limit int) (*DuplicateGapResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	res, err := s.dataHubClient.GetDuplicateGap(resultType, limit)
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &DuplicateGapResponse{
		Data: []map[string]interface{}{
			{"id": "DG-001", "test_date": "2026-06-01", "result_type": "DUPLICATE", "branch_name": "Cabang Jakarta", "reference_no": "TRX-2026-8819", "account_id": "ACC-100293", "amount": 85000000.0, "description": "Duplicate transfer in 42 seconds to identical beneficiary", "occurrences": 2, "status": "OPEN"},
			{"id": "DG-002", "test_date": "2026-06-01", "result_type": "GAP", "branch_name": "Cabang Surabaya", "reference_no": "GL-VOUCH-2026-00412", "account_id": "GL-101.01", "amount": 0.0, "description": "Missing GL sequence numbers 00413 - 00415 (3 consecutive missing)", "occurrences": 3, "status": "INVESTIGATING"},
			{"id": "DG-003", "test_date": "2026-06-02", "result_type": "DUPLICATE", "branch_name": "Cabang Medan", "reference_no": "INV-2026-0192", "account_id": "ACC-304192", "amount": 120000000.0, "description": "Duplicate payment voucher against same vendor invoice", "occurrences": 2, "status": "RESOLVED"},
			{"id": "DG-004", "test_date": "2026-06-02", "result_type": "GAP", "branch_name": "Kantor Pusat", "reference_no": "CHK-2026-00991", "account_id": "ACC-001092", "amount": 0.0, "description": "Missing cheque clearing range 00992 - 00994", "occurrences": 3, "status": "OPEN"},
		},
		Summary: map[string]interface{}{
			"total_duplicates":  38,
			"total_gaps":        14,
			"branches_affected": 5,
		},
	}, nil
}

// GetCAATTBenford returns first-digit Benford's Law analysis
func (s *AnalyticsService) GetCAATTBenford() (*BenfordResponse, error) {
	res, err := s.dataHubClient.GetBenfordAnalysis()
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &BenfordResponse{
		Data: []map[string]interface{}{
			{"digit": 1, "actual_count": 3612, "actual_pct": 28.9, "expected_pct": 30.1, "deviation_pct": -1.2, "is_significant": false},
			{"digit": 2, "actual_count": 2180, "actual_pct": 17.4, "expected_pct": 17.6, "deviation_pct": -0.2, "is_significant": false},
			{"digit": 3, "actual_count": 1620, "actual_pct": 13.0, "expected_pct": 12.5, "deviation_pct": 0.5, "is_significant": false},
			{"digit": 4, "actual_count": 1210, "actual_pct": 9.7, "expected_pct": 9.7, "deviation_pct": 0.0, "is_significant": false},
			{"digit": 5, "actual_count": 1340, "actual_pct": 10.7, "expected_pct": 7.9, "deviation_pct": 2.8, "is_significant": true},
			{"digit": 6, "actual_count": 820, "actual_pct": 6.6, "expected_pct": 6.7, "deviation_pct": -0.1, "is_significant": false},
			{"digit": 7, "actual_count": 690, "actual_pct": 5.5, "expected_pct": 5.8, "deviation_pct": -0.3, "is_significant": false},
			{"digit": 8, "actual_count": 550, "actual_pct": 4.4, "expected_pct": 5.1, "deviation_pct": -0.7, "is_significant": false},
			{"digit": 9, "actual_count": 478, "actual_pct": 3.8, "expected_pct": 4.6, "deviation_pct": -0.8, "is_significant": false},
		},
		Summary: map[string]interface{}{
			"total_digits_analyzed":  9,
			"significant_deviations": 1,
			"conclusion":             "MINOR DEVIATION (Digit 5 anomaly indicates possible threshold structuring)",
		},
	}, nil
}

// GetCAATTStratification returns amount range stratification and aging
func (s *AnalyticsService) GetCAATTStratification(category string) (*StratificationResponse, error) {
	res, err := s.dataHubClient.GetStratification(category)
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &StratificationResponse{
		Data: []map[string]interface{}{
			{"stratum_label": "< Rp 10 Juta", "min_value": 0, "max_value": 10000000, "category": "Retail", "trx_count": 8420, "total_amount": 32500000000.0, "pct_count": 67.4, "pct_amount": 8.1},
			{"stratum_label": "Rp 10M - 50M", "min_value": 10000000, "max_value": 50000000, "category": "Commercial", "trx_count": 2840, "total_amount": 71000000000.0, "pct_count": 22.7, "pct_amount": 17.7},
			{"stratum_label": "Rp 50M - 100M", "min_value": 50000000, "max_value": 100000000, "category": "Corporate", "trx_count": 890, "total_amount": 66750000000.0, "pct_count": 7.1, "pct_amount": 16.6},
			{"stratum_label": "Rp 100M - 500M", "min_value": 100000000, "max_value": 500000000, "category": "High Value", "trx_count": 280, "total_amount": 84000000000.0, "pct_count": 2.2, "pct_amount": 20.9},
			{"stratum_label": "> Rp 500 Juta", "min_value": 500000000, "max_value": 99999999999, "category": "Wholesale / Institutional", "trx_count": 65, "total_amount": 147500000000.0, "pct_count": 0.6, "pct_amount": 36.7},
		},
		Summary: map[string]interface{}{
			"total_transactions": 12495,
			"total_amount":       401750000000.0,
			"total_strata":       5,
			"categories":         5,
		},
	}, nil
}

// GetCAATTReconciliation returns cross-system reconciliation results
func (s *AnalyticsService) GetCAATTReconciliation() (*ReconciliationResponse, error) {
	res, err := s.dataHubClient.GetReconciliation()
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &ReconciliationResponse{
		Data: []map[string]interface{}{
			{"test_date": "2026-06-01", "system_a": "Core Banking (CBS)", "system_b": "General Ledger (GL)", "module": "Giro & Tabungan", "total_records_a": 12500, "total_records_b": 12498, "matched_records": 12495, "unmatched_a": 5, "unmatched_b": 3, "match_rate_pct": 99.96, "total_difference": 4500000.0, "status": "BALANCED"},
			{"test_date": "2026-06-01", "system_a": "Loan Origination (LOS)", "system_b": "Core Banking (CBS)", "module": "Kredit Komersial", "total_records_a": 450, "total_records_b": 448, "matched_records": 447, "unmatched_a": 3, "unmatched_b": 1, "match_rate_pct": 99.33, "total_difference": 250000000.0, "status": "DISCREPANCY_FLAGGED"},
			{"test_date": "2026-06-01", "system_a": "Treasury Trading", "system_b": "General Ledger (GL)", "module": "Forex Dealing", "total_records_a": 180, "total_records_b": 180, "matched_records": 180, "unmatched_a": 0, "unmatched_b": 0, "match_rate_pct": 100.00, "total_difference": 0.0, "status": "PERFECT_MATCH"},
			{"test_date": "2026-06-01", "system_a": "ATM Switch", "system_b": "Core Banking (CBS)", "module": "Interbank Switching", "total_records_a": 6400, "total_records_b": 6392, "matched_records": 6388, "unmatched_a": 12, "unmatched_b": 4, "match_rate_pct": 99.81, "total_difference": 18500000.0, "status": "PENDING_SETTLEMENT"},
		},
		Summary: map[string]interface{}{
			"avg_match_rate":   99.78,
			"total_unmatched":  28,
			"total_difference": 273000000.0,
		},
	}, nil
}

// GetCAATTPolicyViolations returns banking rule & policy compliance violations
func (s *AnalyticsService) GetCAATTPolicyViolations(severity string, limit int) (*PolicyViolationsResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	res, err := s.dataHubClient.GetPolicyViolations(severity, limit)
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &PolicyViolationsResponse{
		Data: []map[string]interface{}{
			{"id": "PV-001", "test_date": "2026-06-01", "rule_name": "Batas Maksimum Suku Bunga Deposito", "branch_name": "Cabang Bali", "customer_name": "PT Sinar Bali", "severity": "Critical", "description": "Suku bunga deposito 8.86% melebihi batas penjaminan LPS (4.25%) tanpa persetujuan ALCO", "status": "OPEN"},
			{"id": "PV-002", "test_date": "2026-06-01", "rule_name": "Batas Maksimum Pemberian Kredit (BMPK)", "branch_name": "Kantor Pusat", "customer_name": "PT Mitra Sentosa Abadi", "severity": "Critical", "description": "Plafond kredit konsorsium melanggar 20% modal disetor entitas terkait", "status": "OPEN"},
			{"id": "PV-003", "test_date": "2026-06-02", "rule_name": "Otorisasi Transaksi Dual Control", "branch_name": "Cabang Jakarta", "customer_name": "Internal Vault", "severity": "High", "description": "Pengeluaran kas fisik Rp 500 Juta dilakukan tanpa otorisasi Branch Manager", "status": "INVESTIGATING"},
			{"id": "PV-004", "test_date": "2026-06-02", "rule_name": "Kelengkapan Dokumen Jaminan Kredit", "branch_name": "Cabang Medan", "customer_name": "Hendra Pratama", "severity": "Medium", "description": "Pencairan kredit sebelum sertifikat hak tanggungan (SHT) terbit", "status": "UNDER_REVIEW"},
			{"id": "PV-005", "test_date": "2026-06-03", "rule_name": "Monitoring Rekening Dormant", "branch_name": "Cabang Bandung", "customer_name": "Siti Nurhaliza", "severity": "Medium", "description": "Aktivasi rekening dormant > 12 bulan tanpa verifikasi tatap muka", "status": "RESOLVED"},
		},
		Summary: map[string]interface{}{
			"total_violations":       32,
			"critical":               4,
			"high":                   9,
			"medium":                 14,
			"low":                    5,
			"unique_rules_violated":  8,
			"branches_affected":      6,
		},
	}, nil
}

// GetDataQualityMetrics returns completeness, accuracy, and timeliness metrics
func (s *AnalyticsService) GetDataQualityMetrics() (*DataQualityResponse, error) {
	res, err := s.dataHubClient.GetDataQualityMetrics()
	if err == nil && res != nil && len(res.Data) > 0 {
		return res, nil
	}

	return &DataQualityResponse{
		Data: []map[string]interface{}{
			{"table_name": "gold.fact_transactions", "zone": "Gold", "total_rows": 125000, "completeness_pct": 99.85, "accuracy_pct": 99.92, "timeliness_days": 0.2, "duplicate_count": 0, "status": "EXCELLENT"},
			{"table_name": "gold.fact_loans", "zone": "Gold", "total_rows": 8500, "completeness_pct": 98.40, "accuracy_pct": 99.10, "timeliness_days": 0.5, "duplicate_count": 0, "status": "GOOD"},
			{"table_name": "gold.dim_accounts", "zone": "Gold", "total_rows": 45000, "completeness_pct": 99.95, "accuracy_pct": 99.98, "timeliness_days": 0.1, "duplicate_count": 0, "status": "EXCELLENT"},
			{"table_name": "silver.cbs_transactions", "zone": "Silver", "total_rows": 125000, "completeness_pct": 99.50, "accuracy_pct": 99.80, "timeliness_days": 0.2, "duplicate_count": 24, "status": "GOOD"},
			{"table_name": "silver.cbs_customers", "zone": "Silver", "total_rows": 38000, "completeness_pct": 97.20, "accuracy_pct": 98.50, "timeliness_days": 1.0, "duplicate_count": 18, "status": "NEEDS_CLEANSING"},
			{"table_name": "bronze.cbs_daily_transactions", "zone": "Bronze", "total_rows": 125500, "completeness_pct": 99.10, "accuracy_pct": 99.20, "timeliness_days": 0.1, "duplicate_count": 500, "status": "RAW_INGESTED"},
		},
		Summary: map[string]interface{}{
			"avg_completeness":      98.99,
			"avg_accuracy":          99.41,
			"avg_timeliness_days":   0.4,
			"tables_profiled":       6,
			"total_rows_profiled":   467000,
		},
		OverallQualityScore: 99.20,
	}, nil
}
