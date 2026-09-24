/**
 * Composable providing simulated/mock data for all 4 AI-powered analytics modules.
 * Data references actual branches, categories, and KPIs from the existing stores.
 */

import { RiskLevel } from '~/types/risk'

// ─── Types ──────────────────────────────────────────────────────────────────

export interface RiskScorePrediction {
  entity: string
  type: 'Branch' | 'Department'
  riskCategory: string
  targetTimeline: string
  predictedLikelihood: number
  predictedImpact: number
  predictedScore: number
  actualScore: number
  delta: number
  trend: 'up' | 'down' | 'stable'
  predictedRiskLevel: RiskLevel
  actualRiskLevel: RiskLevel
}

export interface FeatureImportance {
  feature: string
  importance: number
}

export interface ModelMetrics {
  accuracy: number
  precision: number
  recall: number
  f1Score: number
  auc: number
}

export interface AnomalyRecord {
  id: string
  entity: string
  type: string
  category?: string
  anomalyScore: number
  description: string
  severity: 'Critical' | 'High' | 'Medium' | 'Low'
  date: string
  amount?: number
  frequency?: number
  isAnomaly: boolean
  riskLevel: RiskLevel
  xMetric?: number
}

export interface ScatterPoint {
  x: number
  y: number
  type?: string
  isAnomaly: boolean
  label?: string
}

export interface AnomalySummary {
  totalScanned: number
  anomaliesFound: number
  contaminationRate: number
  topCategory: string
}

export interface NLPDocumentResult {
  docId: string
  title: string
  source: 'Working Paper' | 'Audit Result Report'
  autoCategory: string
  confidence: number
  sentiment: 'Positive' | 'Neutral' | 'Negative'
  severityScore: number
  date: string
  excerpt: string
  riskLevel: RiskLevel
}

export interface SentimentDistribution {
  positive: number
  neutral: number
  negative: number
}



export interface KPIForecast {
  kpiName: string
  code: string
  unit: string
  entity: string
  entityType: 'Department' | 'Branch'
  targetHorizon: string
  currentValue: number
  forecastedValue: number
  trend: 'Improving' | 'Declining' | 'Stable'
  recommendedAction: string
  riskLevel: RiskLevel
}

export interface TimeSeriesPoint {
  period: string
  actual: number | null
  forecast: number | null
  upperBound: number | null
  lowerBound: number | null
}

export interface AtRiskDepartment {
  department: string
  kpi: string
  currentTrend: number
  predictedQ3: number
  riskLevel: RiskLevel
}

// ─── Data Generators ────────────────────────────────────────────────────────

export function useXGBoostData() {
  const predictions: RiskScorePrediction[] = [
    { entity: 'Head Office', type: 'Branch', riskCategory: 'Financial', targetTimeline: 'Q3 2026', predictedLikelihood: 4.2, predictedImpact: 4.6, predictedScore: 19.3, actualScore: 20, delta: -0.7, trend: 'up', predictedRiskLevel: RiskLevel.HIGH, actualRiskLevel: RiskLevel.HIGH },
    { entity: 'Jakarta Branch', type: 'Branch', riskCategory: 'Operational', targetTimeline: 'Q3 2026', predictedLikelihood: 3.8, predictedImpact: 4.1, predictedScore: 15.6, actualScore: 16, delta: -0.4, trend: 'up', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
    { entity: 'Surabaya Branch', type: 'Branch', riskCategory: 'Technology', targetTimeline: 'Q3 2026', predictedLikelihood: 3.5, predictedImpact: 3.9, predictedScore: 13.7, actualScore: 14, delta: -0.3, trend: 'stable', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
    { entity: 'Bandung Branch', type: 'Branch', riskCategory: 'Compliance', targetTimeline: 'Q3 2026', predictedLikelihood: 3.1, predictedImpact: 3.4, predictedScore: 10.5, actualScore: 11, delta: -0.5, trend: 'down', predictedRiskLevel: RiskLevel.MODERATE, actualRiskLevel: RiskLevel.MODERATE },
    { entity: 'Bali Branch', type: 'Branch', riskCategory: 'Strategic', targetTimeline: 'Q3 2026', predictedLikelihood: 3.3, predictedImpact: 4.0, predictedScore: 13.2, actualScore: 13, delta: 0.2, trend: 'stable', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
    { entity: 'Finance Dept', type: 'Department', riskCategory: 'Financial', targetTimeline: 'Q3 2026', predictedLikelihood: 4.5, predictedImpact: 4.8, predictedScore: 21.6, actualScore: 22, delta: -0.4, trend: 'up', predictedRiskLevel: RiskLevel.HIGH, actualRiskLevel: RiskLevel.HIGH },
    { entity: 'IT Dept', type: 'Department', riskCategory: 'Technology', targetTimeline: 'Q3 2026', predictedLikelihood: 4.0, predictedImpact: 4.3, predictedScore: 17.2, actualScore: 18, delta: -0.8, trend: 'up', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
    { entity: 'HR Dept', type: 'Department', riskCategory: 'Reputational', targetTimeline: 'Q3 2026', predictedLikelihood: 2.8, predictedImpact: 3.0, predictedScore: 8.4, actualScore: 9, delta: -0.6, trend: 'down', predictedRiskLevel: RiskLevel.MODERATE, actualRiskLevel: RiskLevel.MODERATE },
    { entity: 'Operations Dept', type: 'Department', riskCategory: 'Operational', targetTimeline: 'Q3 2026', predictedLikelihood: 3.6, predictedImpact: 3.7, predictedScore: 13.3, actualScore: 13, delta: 0.3, trend: 'stable', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
    { entity: 'Legal & Compliance', type: 'Department', riskCategory: 'Legal', targetTimeline: 'Q3 2026', predictedLikelihood: 3.2, predictedImpact: 4.2, predictedScore: 13.4, actualScore: 14, delta: -0.6, trend: 'stable', predictedRiskLevel: RiskLevel.MODERATE_HIGH, actualRiskLevel: RiskLevel.MODERATE_HIGH },
  ]

  const featureImportance: FeatureImportance[] = [
    { feature: 'Prior Audit Findings Count', importance: 0.36 },
    { feature: 'KPI Achievement Rate', importance: 0.29 },
    { feature: 'Transaction Volume', importance: 0.22 },
    { feature: 'Outstanding Mitigations', importance: 0.08 },
    { feature: 'Previous Risk Score', importance: 0.05 },
  ]

  const modelMetrics: ModelMetrics = {
    accuracy: 0.912,
    precision: 0.895,
    recall: 0.928,
    f1Score: 0.911,
    auc: 0.947,
  }

  const chartLabels = predictions.map(p => p.entity)
  const predictedScores = predictions.map(p => p.predictedScore)
  const actualScores = predictions.map(p => p.actualScore)

  return { predictions, featureImportance, modelMetrics, chartLabels, predictedScores, actualScores }
}

export function useIsolationForestData() {
  const anomalies: AnomalyRecord[] = [
    { id: 'BANK-TRX-0001', entity: 'Cabang Medan', type: 'IT Control', anomalyScore: -0.95, description: 'Aktivitas akses administratif basis data core banking pada dini hari (02:00) tanpa tiket Change Request', severity: 'Critical', date: '2026-06-01', amount: 0, isAnomaly: true, riskLevel: RiskLevel.HIGH, xMetric: 7 },
    { id: 'BANK-TRX-0003', entity: 'Cabang Bali', type: 'Funding', anomalyScore: -0.91, description: 'Pemberian suku bunga deposito ekstrem (8.86%) di atas batas penjaminan LPS tanpa persetujuan ALCO', severity: 'Critical', date: '2026-06-01', amount: 21465500000, isAnomaly: true, riskLevel: RiskLevel.HIGH, xMetric: 21465.5 },
    { id: 'BANK-TRX-0007', entity: 'Kantor Pusat', type: 'Lending', anomalyScore: -0.89, description: 'Pencairan kredit bernilai Rp 66.662 Juta dengan bunga murah ekstrem (4.18%) melanggar BMPK', severity: 'Critical', date: '2026-06-01', amount: 66662330000, isAnomaly: true, riskLevel: RiskLevel.HIGH, xMetric: 66662.33 },
    { id: 'BANK-TRX-0009', entity: 'Cabang Jakarta', type: 'Treasury', anomalyScore: -0.84, description: 'Transaksi valas dealing room bernilai Rp 62.412 Juta di luar rentang kuotasi pasar resmi', severity: 'High', date: '2026-06-01', amount: 62412080000, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH, xMetric: 62412.08 },
    { id: 'BANK-TRX-0014', entity: 'Cabang Surabaya', type: 'KYC', anomalyScore: -0.82, description: 'Profil transaksi nasabah high risk deviasi 123.5% dari profil historis tanpa pembaruan EDD', severity: 'High', date: '2026-06-01', amount: 0, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH, xMetric: 123.53 },
    { id: 'BANK-TRX-0018', entity: 'Cabang Bandung', type: 'Payment', anomalyScore: -0.78, description: 'Transaksi pembayaran split-bill berulang dengan frekuensi tinggi mendekati batas threshold harian', severity: 'High', date: '2026-06-01', amount: 55950000, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH, xMetric: 55.95 },
  ]

  // Generate scatter plot data covering 6 bank categories
  const scatterData: ScatterPoint[] = []
  const bankCategories = ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control']

  // Normal points for each category
  bankCategories.forEach((cat) => {
    for (let i = 0; i < 15; i++) {
      let xVal = 0
      if (cat === 'Funding') xVal = Math.round(Math.random() * 800 + 50)
      else if (cat === 'Lending') xVal = Math.round(Math.random() * 3000 + 200)
      else if (cat === 'Treasury') xVal = Math.round(Math.random() * 5000 + 500)
      else if (cat === 'Payment') xVal = Math.round(Math.random() * 150 + 5)
      else if (cat === 'KYC') xVal = Math.round(Math.random() * 25 + 1)
      else xVal = Math.round(Math.random() * 3)

      scatterData.push({
        x: xVal,
        y: Math.round(Math.random() * 25 + 10),
        type: cat,
        isAnomaly: false,
        label: `NORM-${cat.substring(0, 3).toUpperCase()}-${i + 1}`
      })
    }
  })

  // Anomaly outlier points
  scatterData.push(
    { x: 21465.5, y: 55.0, type: 'Funding', isAnomaly: true, label: 'BANK-TRX-0003' },
    { x: 66662.33, y: 58.0, type: 'Lending', isAnomaly: true, label: 'BANK-TRX-0007' },
    { x: 62412.08, y: 52.0, type: 'Treasury', isAnomaly: true, label: 'BANK-TRX-0009' },
    { x: 55.95, y: 48.0, type: 'Payment', isAnomaly: true, label: 'BANK-TRX-0018' },
    { x: 123.53, y: 46.0, type: 'KYC', isAnomaly: true, label: 'BANK-TRX-0014' },
    { x: 7.0, y: 50.0, type: 'IT Control', isAnomaly: true, label: 'BANK-TRX-0001' }
  )

  const summary: AnomalySummary = {
    totalScanned: 1200,
    anomaliesFound: 150,
    contaminationRate: 0.125,
    topCategory: 'Funding',
  }

  return { anomalies, scatterData, summary }
}

export function useIndoBERTData() {
  const documents: NLPDocumentResult[] = [
    { docId: 'WP-2026-041', title: 'Evaluasi Pengendalian Internal atas Pengelolaan Kas', source: 'Working Paper', autoCategory: 'Financial', confidence: 0.96, sentiment: 'Negative', severityScore: 82, date: '2026-05-28', excerpt: 'Ditemukan kelemahan signifikan dalam prosedur otorisasi pembayaran kas besar...', riskLevel: RiskLevel.HIGH },
    { docId: 'ARR-2026-018', title: 'Laporan Hasil Audit Kepatuhan TI Cabang Jakarta', source: 'Audit Result Report', autoCategory: 'Technology', confidence: 0.93, sentiment: 'Negative', severityScore: 78, date: '2026-06-01', excerpt: 'Beberapa sistem kritis tidak memiliki prosedur backup yang memadai dan patch keamanan tertunda...', riskLevel: RiskLevel.MODERATE_HIGH },
    { docId: 'WP-2026-042', title: 'Review Proses Pengadaan Barang dan Jasa', source: 'Working Paper', autoCategory: 'Operations', confidence: 0.91, sentiment: 'Neutral', severityScore: 65, date: '2026-05-20', excerpt: 'Proses tender telah mengikuti SOP namun ditemukan keterlambatan dalam evaluasi penawaran...', riskLevel: RiskLevel.MODERATE },
    { docId: 'ARR-2026-019', title: 'Audit Kepatuhan Regulasi Anti Pencucian Uang', source: 'Audit Result Report', autoCategory: 'Compliance', confidence: 0.94, sentiment: 'Negative', severityScore: 88, date: '2026-05-15', excerpt: 'Prosedur KYC (Know Your Customer) belum sepenuhnya dilaksanakan untuk nasabah risiko tinggi...', riskLevel: RiskLevel.HIGH },
    { docId: 'WP-2026-043', title: 'Pemeriksaan Efektivitas Program Pelatihan SDM', source: 'Working Paper', autoCategory: 'Human Resources', confidence: 0.88, sentiment: 'Positive', severityScore: 35, date: '2026-06-03', excerpt: 'Program pelatihan menunjukkan peningkatan kompetensi yang terukur berdasarkan pre-post assessment...', riskLevel: RiskLevel.LOW_MODERATE },
    { docId: 'WP-2026-044', title: 'Evaluasi Tata Kelola Perusahaan (GCG)', source: 'Working Paper', autoCategory: 'Governance', confidence: 0.90, sentiment: 'Neutral', severityScore: 52, date: '2026-05-25', excerpt: 'Implementasi GCG secara umum telah berjalan baik, namun perlu penguatan fungsi whistle-blowing...', riskLevel: RiskLevel.MODERATE },
    { docId: 'ARR-2026-020', title: 'Laporan Audit Operasional Cabang Surabaya', source: 'Audit Result Report', autoCategory: 'Operations', confidence: 0.87, sentiment: 'Neutral', severityScore: 58, date: '2026-06-05', excerpt: 'Operasional cabang berjalan sesuai target namun utilisasi aset masih di bawah standar optimal...', riskLevel: RiskLevel.MODERATE },
    { docId: 'WP-2026-045', title: 'Analisis Risiko Kredit Portofolio Mikro', source: 'Working Paper', autoCategory: 'Financial', confidence: 0.95, sentiment: 'Negative', severityScore: 75, date: '2026-05-30', excerpt: 'NPL ratio untuk segmen mikro meningkat 2.3% YoY, mengindikasikan penurunan kualitas aset...', riskLevel: RiskLevel.MODERATE_HIGH },
    { docId: 'ARR-2026-021', title: 'Audit Keamanan Siber dan Perlindungan Data', source: 'Audit Result Report', autoCategory: 'Technology', confidence: 0.92, sentiment: 'Negative', severityScore: 85, date: '2026-06-08', excerpt: 'Penetration testing mengungkapkan 3 kerentanan kritis yang belum diperbaiki sejak audit terakhir...', riskLevel: RiskLevel.HIGH },
    { docId: 'WP-2026-046', title: 'Review Strategi Ekspansi Pasar Regional', source: 'Working Paper', autoCategory: 'Strategic', confidence: 0.86, sentiment: 'Positive', severityScore: 30, date: '2026-06-10', excerpt: 'Rencana ekspansi ke wilayah Kalimantan telah didukung studi kelayakan yang komprehensif...', riskLevel: RiskLevel.LOW },
  ]

  const sentimentDistribution: SentimentDistribution = {
    positive: 2,
    neutral: 3,
    negative: 5,
  }

  const categoryDistribution = {
    'Financial': 3,
    'Technology': 2,
    'Operations': 2,
    'Compliance': 1,
    'Human Resources': 1,
    'Governance': 1,
    'Strategic': 1,
  }

  return { documents, sentimentDistribution, categoryDistribution }
}

export function useTimeSeriesData() {
  const historicalKPI: TimeSeriesPoint[] = [
    { period: 'Q1 2025', actual: 72.5, forecast: null, upperBound: null, lowerBound: null },
    { period: 'Q2 2025', actual: 68.3, forecast: null, upperBound: null, lowerBound: null },
    { period: 'Q3 2025', actual: 75.1, forecast: null, upperBound: null, lowerBound: null },
    { period: 'Q4 2025', actual: 71.8, forecast: null, upperBound: null, lowerBound: null },
    { period: 'Q1 2026', actual: 69.4, forecast: null, upperBound: null, lowerBound: null },
    { period: 'Q2 2026', actual: 66.2, forecast: 67.0, upperBound: 70.5, lowerBound: 63.5 },
    { period: 'Q3 2026', actual: null, forecast: 63.8, upperBound: 68.2, lowerBound: 59.4 },
    { period: 'Q4 2026', actual: null, forecast: 61.5, upperBound: 66.8, lowerBound: 56.2 },
    { period: 'Q1 2027', actual: null, forecast: 59.2, upperBound: 65.1, lowerBound: 53.3 },
  ]

  const kpiForecasts: KPIForecast[] = [
    { kpiName: 'Revenue Operational Cost', code: 'SO-IA01', unit: '%', entity: 'Finance Dept', entityType: 'Department', targetHorizon: 'Q3 2026', currentValue: 33.33, forecastedValue: 28.5, trend: 'Declining', recommendedAction: 'Prioritize operational cost audit — projected KPI miss by 15% next quarter', riskLevel: RiskLevel.HIGH },
    { kpiName: 'Customer Satisfaction Index', code: 'SO-IA02', unit: 'Score', entity: 'Jakarta Branch', entityType: 'Branch', targetHorizon: 'Q3 2026', currentValue: 94.44, forecastedValue: 91.2, trend: 'Declining', recommendedAction: 'Monitor service quality metrics — moderate decline expected', riskLevel: RiskLevel.MODERATE },
    { kpiName: 'Audit Response Time', code: 'SO-IA03', unit: 'Hour', entity: 'HR Dept', entityType: 'Department', targetHorizon: 'Q3 2026', currentValue: 200.00, forecastedValue: 210.5, trend: 'Improving', recommendedAction: 'KPI on track — maintain current audit workflow efficiency', riskLevel: RiskLevel.LOW },
    { kpiName: 'Internal Control Effectiveness', code: 'SO-IA04', unit: '%', entity: 'Bandung Branch', entityType: 'Branch', targetHorizon: 'Q3 2026', currentValue: 82.0, forecastedValue: 78.3, trend: 'Declining', recommendedAction: 'Schedule comprehensive control testing — predicted effectiveness drop', riskLevel: RiskLevel.MODERATE_HIGH },
    { kpiName: 'Compliance Adherence Rate', code: 'SO-IA05', unit: '%', entity: 'Finance Dept', entityType: 'Department', targetHorizon: 'Q3 2026', currentValue: 91.5, forecastedValue: 93.1, trend: 'Improving', recommendedAction: 'Compliance trending positively — continue current regulatory monitoring', riskLevel: RiskLevel.LOW },
    { kpiName: 'Risk Mitigation Completion', code: 'SO-IA06', unit: '%', entity: 'Jakarta Branch', entityType: 'Branch', targetHorizon: 'Q3 2026', currentValue: 67.0, forecastedValue: 58.2, trend: 'Declining', recommendedAction: 'Escalate mitigation backlog to management — significant completion rate decline projected', riskLevel: RiskLevel.HIGH },
  ]

  const atRiskDepartments: AtRiskDepartment[] = [
    { department: 'Finance Dept', kpi: 'Revenue Operational Cost', currentTrend: -12.5, predictedQ3: -18.2, riskLevel: RiskLevel.HIGH },
    { department: 'Operations Dept', kpi: 'Risk Mitigation Completion', currentTrend: -8.3, predictedQ3: -15.1, riskLevel: RiskLevel.HIGH },
    { department: 'IT Dept', kpi: 'Internal Control Effectiveness', currentTrend: -5.7, predictedQ3: -9.4, riskLevel: RiskLevel.MODERATE },
  ]

  const forecastAccuracy = {
    mape: 4.8,
    rmse: 3.2,
    r2Score: 0.923,
  }

  return { historicalKPI, kpiForecasts, atRiskDepartments, forecastAccuracy }
}

// ─── Summary Stats ──────────────────────────────────────────────────────────

export function useAnalyticsSummary() {
  return {
    totalEntitiesScored: 10,
    anomaliesDetected: 150,
    documentsAnalyzed: 10,
    kpiAlerts: 3,
  }
}

// ─── CAATT Analytics Types & Generators ──────────────────────────────────────

// 1. Full Population Testing
export interface CAATTFullPopulationRecord {
  id: string
  testDate: string
  branchName: string
  accountId: string
  customerName: string
  category: string
  transactionAmount: number
  thresholdLimit: number
  excessAmount: number
  violationType: string
  status: 'FLAGGED' | 'UNDER_REVIEW' | 'RESOLVED'
}

export interface CAATTFullPopulationSummary {
  totalTested: number
  totalViolations: number
  avgViolationRate: number
  branchesTested: number
  categoriesTested: number
}

export function useCAATTFullPopulationData() {
  const records: CAATTFullPopulationRecord[] = [
    { id: 'FPT-001', testDate: '2026-06-01', branchName: 'Cabang Jakarta', accountId: 'ACC-100293', customerName: 'PT Mega Pratama', category: 'Lending Plafond', transactionAmount: 12500000000, thresholdLimit: 10000000000, excessAmount: 2500000000, violationType: 'Plafond Exceeded', status: 'FLAGGED' },
    { id: 'FPT-002', testDate: '2026-06-01', branchName: 'Cabang Surabaya', accountId: 'ACC-209118', customerName: 'CV Bintang Sejahtera', category: 'Cash Withdrawal', transactionAmount: 750000000, thresholdLimit: 500000000, excessAmount: 250000000, violationType: 'Daily Limit Exceeded', status: 'FLAGGED' },
    { id: 'FPT-003', testDate: '2026-06-02', branchName: 'Cabang Medan', accountId: 'ACC-304192', customerName: 'Hendra Wijaya', category: 'Single Transfer', transactionAmount: 1500000000, thresholdLimit: 1000000000, excessAmount: 500000000, violationType: 'Single Transfer Limit', status: 'UNDER_REVIEW' },
    { id: 'FPT-004', testDate: '2026-06-02', branchName: 'Kantor Pusat', accountId: 'ACC-001092', customerName: 'PT Sentosa Global', category: 'Treasury FX', transactionAmount: 45000000000, thresholdLimit: 30000000000, excessAmount: 15000000000, violationType: 'ALCO Approval Required', status: 'FLAGGED' },
    { id: 'FPT-005', testDate: '2026-06-03', branchName: 'Cabang Bandung', accountId: 'ACC-408129', customerName: 'Dewi Sartika', category: 'Overdraft', transactionAmount: 350000000, thresholdLimit: 200000000, excessAmount: 150000000, violationType: 'Unauthorized Overdraft', status: 'RESOLVED' },
    { id: 'FPT-006', testDate: '2026-06-03', branchName: 'Cabang Bali', accountId: 'ACC-501239', customerName: 'I Made Sudarta', category: 'Special Rate', transactionAmount: 8500000000, thresholdLimit: 5000000000, excessAmount: 3500000000, violationType: 'Interest Rate Cap', status: 'UNDER_REVIEW' },
  ]

  const summary: CAATTFullPopulationSummary = {
    totalTested: 12480,
    totalViolations: 142,
    avgViolationRate: 1.14,
    branchesTested: 6,
    categoriesTested: 5,
  }

  return { records, summary }
}

// 2. Duplicate & Gap Detection
export interface CAATTDuplicateGapRecord {
  id: string
  testDate: string
  resultType: 'DUPLICATE' | 'GAP'
  branchName: string
  referenceNo: string
  accountId: string
  amount: number
  description: string
  occurrences: number
  status: 'OPEN' | 'INVESTIGATING' | 'RESOLVED'
}

export interface CAATTDuplicateGapSummary {
  totalDuplicates: number
  totalGaps: number
  branchesAffected: number
}

export function useCAATTDuplicateGapData() {
  const records: CAATTDuplicateGapRecord[] = [
    { id: 'DG-001', testDate: '2026-06-01', resultType: 'DUPLICATE', branchName: 'Cabang Jakarta', referenceNo: 'TRX-2026-8819', accountId: 'ACC-100293', amount: 85000000, description: 'Transfer ganda dalam 42 detik ke rekening penerima identik', occurrences: 2, status: 'OPEN' },
    { id: 'DG-002', testDate: '2026-06-01', resultType: 'GAP', branchName: 'Cabang Surabaya', referenceNo: 'GL-VOUCH-2026-00412', accountId: 'GL-101.01', amount: 0, description: 'Nomor voucher GL hilang: 00413 s/d 00415 (3 nomor berurutan hilang)', occurrences: 3, status: 'INVESTIGATING' },
    { id: 'DG-003', testDate: '2026-06-02', resultType: 'DUPLICATE', branchName: 'Cabang Medan', referenceNo: 'INV-2026-0192', accountId: 'ACC-304192', amount: 120000000, description: 'Pencairan biaya ganda untuk nomor invoice vendor yang sama', occurrences: 2, status: 'RESOLVED' },
    { id: 'DG-004', testDate: '2026-06-02', resultType: 'GAP', branchName: 'Kantor Pusat', referenceNo: 'CHK-2026-00991', accountId: 'ACC-001092', amount: 0, description: 'Nomor warkat kliring cek hilang: 00992 s/d 00994', occurrences: 3, status: 'OPEN' },
    { id: 'DG-005', testDate: '2026-06-03', resultType: 'DUPLICATE', branchName: 'Cabang Bandung', referenceNo: 'TRX-2026-9014', accountId: 'ACC-408129', amount: 45000000, description: 'Debit rekening berulang tanpa otorisasi nasabah kedua', occurrences: 2, status: 'OPEN' },
  ]

  const summary: CAATTDuplicateGapSummary = {
    totalDuplicates: 38,
    totalGaps: 14,
    branchesAffected: 5,
  }

  return { records, summary }
}

// 3. Benford's Law
export interface CAATTBenfordRecord {
  digit: number
  actualCount: number
  actualPct: number
  expectedPct: number
  deviationPct: number
  isSignificant: boolean
}

export interface CAATTBenfordSummary {
  totalDigitsAnalyzed: number
  significantDeviations: number
  conclusion: string
}

export function useCAATTBenfordData() {
  const records: CAATTBenfordRecord[] = [
    { digit: 1, actualCount: 3612, actualPct: 28.9, expectedPct: 30.1, deviationPct: -1.2, isSignificant: false },
    { digit: 2, actualCount: 2180, actualPct: 17.4, expectedPct: 17.6, deviationPct: -0.2, isSignificant: false },
    { digit: 3, actualCount: 1620, actualPct: 13.0, expectedPct: 12.5, deviationPct: 0.5, isSignificant: false },
    { digit: 4, actualCount: 1210, actualPct: 9.7, expectedPct: 9.7, deviationPct: 0.0, isSignificant: false },
    { digit: 5, actualCount: 1340, actualPct: 10.7, expectedPct: 7.9, deviationPct: 2.8, isSignificant: true },
    { digit: 6, actualCount: 820, actualPct: 6.6, expectedPct: 6.7, deviationPct: -0.1, isSignificant: false },
    { digit: 7, actualCount: 690, actualPct: 5.5, expectedPct: 5.8, deviationPct: -0.3, isSignificant: false },
    { digit: 8, actualCount: 550, actualPct: 4.4, expectedPct: 5.1, deviationPct: -0.7, isSignificant: false },
    { digit: 9, actualCount: 478, actualPct: 3.8, expectedPct: 4.6, deviationPct: -0.8, isSignificant: false },
  ]

  const summary: CAATTBenfordSummary = {
    totalDigitsAnalyzed: 9,
    significantDeviations: 1,
    conclusion: 'MINOR DEVIATION (Lonjakan digit 5 mengindikasikan potensi structuring Rp 49.xxx.xxx)',
  }

  return { records, summary }
}

// 4. Stratification & Aging
export interface CAATTStratificationRecord {
  stratumLabel: string
  minValue: number
  maxValue: number
  category: string
  trxCount: number
  totalAmount: number
  pctCount: number
  pctAmount: number
}

export interface CAATTStratificationSummary {
  totalTransactions: number
  totalAmount: number
  totalStrata: number
  categories: number
}

export function useCAATTStratificationData() {
  const records: CAATTStratificationRecord[] = [
    { stratumLabel: '< Rp 10 Juta', minValue: 0, maxValue: 10000000, category: 'Retail', trxCount: 8420, totalAmount: 32500000000, pctCount: 67.4, pctAmount: 8.1 },
    { stratumLabel: 'Rp 10M - 50M', minValue: 10000000, maxValue: 50000000, category: 'Commercial', trxCount: 2840, totalAmount: 71000000000, pctCount: 22.7, pctAmount: 17.7 },
    { stratumLabel: 'Rp 50M - 100M', minValue: 50000000, maxValue: 100000000, category: 'Corporate', trxCount: 890, totalAmount: 66750000000, pctCount: 7.1, pctAmount: 16.6 },
    { stratumLabel: 'Rp 100M - 500M', minValue: 100000000, maxValue: 500000000, category: 'High Value', trxCount: 280, totalAmount: 84000000000, pctCount: 2.2, pctAmount: 20.9 },
    { stratumLabel: '> Rp 500 Juta', minValue: 500000000, maxValue: 99999999999, category: 'Wholesale / Inst', trxCount: 65, totalAmount: 147500000000, pctCount: 0.6, pctAmount: 36.7 },
  ]

  const summary: CAATTStratificationSummary = {
    totalTransactions: 12495,
    totalAmount: 401750000000,
    totalStrata: 5,
    categories: 5,
  }

  return { records, summary }
}

// 5. Cross-System Reconciliation
export interface CAATTReconciliationRecord {
  testDate: string
  systemA: string
  systemB: string
  module: string
  totalRecordsA: number
  totalRecordsB: number
  matchedRecords: number
  unmatchedA: number
  unmatchedB: number
  matchRatePct: number
  totalDifference: number
  status: 'BALANCED' | 'DISCREPANCY_FLAGGED' | 'PERFECT_MATCH' | 'PENDING_SETTLEMENT'
}

export interface CAATTReconciliationSummary {
  avgMatchRate: number
  totalUnmatched: number
  totalDifference: number
}

export function useCAATTReconciliationData() {
  const records: CAATTReconciliationRecord[] = [
    { testDate: '2026-06-01', systemA: 'Core Banking (CBS)', systemB: 'General Ledger (GL)', module: 'Giro & Tabungan', totalRecordsA: 12500, totalRecordsB: 12498, matchedRecords: 12495, unmatchedA: 5, unmatchedB: 3, matchRatePct: 99.96, totalDifference: 4500000, status: 'BALANCED' },
    { testDate: '2026-06-01', systemA: 'Loan Origination (LOS)', systemB: 'Core Banking (CBS)', module: 'Kredit Komersial', totalRecordsA: 450, totalRecordsB: 448, matchedRecords: 447, unmatchedA: 3, unmatchedB: 1, matchRatePct: 99.33, totalDifference: 250000000, status: 'DISCREPANCY_FLAGGED' },
    { testDate: '2026-06-01', systemA: 'Treasury Trading', systemB: 'General Ledger (GL)', module: 'Forex Dealing', totalRecordsA: 180, totalRecordsB: 180, matchedRecords: 180, unmatchedA: 0, unmatchedB: 0, matchRatePct: 100.00, totalDifference: 0, status: 'PERFECT_MATCH' },
    { testDate: '2026-06-01', systemA: 'ATM Switch', systemB: 'Core Banking (CBS)', module: 'Interbank Switching', totalRecordsA: 6400, totalRecordsB: 6392, matchedRecords: 6388, unmatchedA: 12, unmatchedB: 4, matchRatePct: 99.81, totalDifference: 18500000, status: 'PENDING_SETTLEMENT' },
  ]

  const summary: CAATTReconciliationSummary = {
    avgMatchRate: 99.78,
    totalUnmatched: 28,
    totalDifference: 273000000,
  }

  return { records, summary }
}

// 6. Policy & Rule Compliance
export interface CAATTPolicyRecord {
  id: string
  testDate: string
  ruleName: string
  branchName: string
  customerName: string
  severity: 'Critical' | 'High' | 'Medium' | 'Low'
  description: string
  status: 'OPEN' | 'INVESTIGATING' | 'UNDER_REVIEW' | 'RESOLVED'
}

export interface CAATTPolicySummary {
  totalViolations: number
  critical: number
  high: number
  medium: number
  low: number
  uniqueRulesViolated: number
  branchesAffected: number
}

export function useCAATTPolicyData() {
  const records: CAATTPolicyRecord[] = [
    { id: 'PV-001', testDate: '2026-06-01', ruleName: 'Batas Maksimum Suku Bunga Deposito', branchName: 'Cabang Bali', customerName: 'PT Sinar Bali', severity: 'Critical', description: 'Suku bunga deposito 8.86% melebihi batas penjaminan LPS (4.25%) tanpa persetujuan ALCO', status: 'OPEN' },
    { id: 'PV-002', testDate: '2026-06-01', ruleName: 'Batas Maksimum Pemberian Kredit (BMPK)', branchName: 'Kantor Pusat', customerName: 'PT Mitra Sentosa Abadi', severity: 'Critical', description: 'Plafond kredit konsorsium melanggar 20% modal disetor entitas terkait', status: 'OPEN' },
    { id: 'PV-003', testDate: '2026-06-02', ruleName: 'Otorisasi Transaksi Dual Control', branchName: 'Cabang Jakarta', customerName: 'Internal Vault', severity: 'High', description: 'Pengeluaran kas fisik Rp 500 Juta dilakukan tanpa otorisasi Branch Manager', status: 'INVESTIGATING' },
    { id: 'PV-004', testDate: '2026-06-02', ruleName: 'Kelengkapan Dokumen Jaminan Kredit', branchName: 'Cabang Medan', customerName: 'Hendra Pratama', severity: 'Medium', description: 'Pencairan kredit sebelum sertifikat hak tanggungan (SHT) terbit', status: 'UNDER_REVIEW' },
    { id: 'PV-005', testDate: '2026-06-03', ruleName: 'Monitoring Rekening Dormant', branchName: 'Cabang Bandung', customerName: 'Siti Nurhaliza', severity: 'Medium', description: 'Aktivasi rekening dormant > 12 bulan tanpa verifikasi tatap muka', status: 'RESOLVED' },
  ]

  const summary: CAATTPolicySummary = {
    totalViolations: 32,
    critical: 4,
    high: 9,
    medium: 14,
    low: 5,
    uniqueRulesViolated: 8,
    branchesAffected: 6,
  }

  return { records, summary }
}

// 7. Data Quality Dashboard
export interface CAATTDataQualityRecord {
  tableName: string
  zone: 'Bronze' | 'Silver' | 'Gold'
  totalRows: number
  completenessPct: number
  accuracyPct: number
  timelinessDays: number
  duplicateCount: number
  status: 'EXCELLENT' | 'GOOD' | 'NEEDS_CLEANSING' | 'RAW_INGESTED'
}

export interface CAATTDataQualitySummary {
  avgCompleteness: number
  avgAccuracy: number
  avgTimelinessDays: number
  tablesProfiled: number
  totalRowsProfiled: number
}

export function useCAATTDataQualityData() {
  const records: CAATTDataQualityRecord[] = [
    { tableName: 'gold.fact_transactions', zone: 'Gold', totalRows: 125000, completenessPct: 99.85, accuracyPct: 99.92, timelinessDays: 0.2, duplicateCount: 0, status: 'EXCELLENT' },
    { tableName: 'gold.fact_loans', zone: 'Gold', totalRows: 8500, completenessPct: 98.40, accuracyPct: 99.10, timelinessDays: 0.5, duplicateCount: 0, status: 'GOOD' },
    { tableName: 'gold.dim_accounts', zone: 'Gold', totalRows: 45000, completenessPct: 99.95, accuracyPct: 99.98, timelinessDays: 0.1, duplicateCount: 0, status: 'EXCELLENT' },
    { tableName: 'silver.cbs_transactions', zone: 'Silver', totalRows: 125000, completenessPct: 99.50, accuracyPct: 99.80, timelinessDays: 0.2, duplicateCount: 24, status: 'GOOD' },
    { tableName: 'silver.cbs_customers', zone: 'Silver', totalRows: 38000, completenessPct: 97.20, accuracyPct: 98.50, timelinessDays: 1.0, duplicateCount: 18, status: 'NEEDS_CLEANSING' },
    { tableName: 'bronze.cbs_daily_transactions', zone: 'Bronze', totalRows: 125500, completenessPct: 99.10, accuracyPct: 99.20, timelinessDays: 0.1, duplicateCount: 500, status: 'RAW_INGESTED' },
  ]

  const summary: CAATTDataQualitySummary = {
    avgCompleteness: 98.99,
    avgAccuracy: 99.41,
    avgTimelinessDays: 0.4,
    tablesProfiled: 6,
    totalRowsProfiled: 467000,
  }

  const overallQualityScore = 99.20

  return { records, summary, overallQualityScore }
}

