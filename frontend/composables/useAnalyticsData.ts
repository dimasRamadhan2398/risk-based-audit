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
  anomalyScore: number
  description: string
  severity: 'Critical' | 'High' | 'Medium' | 'Low'
  date: string
  amount?: number
  frequency?: number
  isAnomaly: boolean
  riskLevel: RiskLevel
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
    { feature: 'Prior Audit Findings Count', importance: 0.28 },
    { feature: 'KPI Achievement Rate', importance: 0.22 },
    { feature: 'Transaction Volume', importance: 0.17 },
    { feature: 'Employee Turnover Rate', importance: 0.12 },
    { feature: 'Compliance Score', importance: 0.09 },
    { feature: 'Outstanding Mitigations', importance: 0.06 },
    { feature: 'Previous Risk Score', importance: 0.04 },
    { feature: 'External Audit Flags', importance: 0.02 },
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
    { id: 'ANM-001', entity: 'Jakarta Branch', type: 'Transaction', anomalyScore: -0.92, description: 'Unusual after-hours bulk payment processing — 47 transactions totaling Rp 2.3B within 12 minutes', severity: 'Critical', date: '2026-05-28', amount: 2300000000, frequency: 47, isAnomaly: true, riskLevel: RiskLevel.HIGH },
    { id: 'ANM-002', entity: 'Head Office', type: 'Access Pattern', anomalyScore: -0.87, description: 'Repeated failed authentication attempts on financial system by terminated employee credentials', severity: 'Critical', date: '2026-06-01', amount: 0, frequency: 156, isAnomaly: true, riskLevel: RiskLevel.HIGH },
    { id: 'ANM-003', entity: 'Finance Dept', type: 'Expense Report', anomalyScore: -0.81, description: 'Duplicate vendor invoices detected — same amounts, sequential invoice numbers, different dates', severity: 'High', date: '2026-05-15', amount: 450000000, frequency: 12, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH },
    { id: 'ANM-004', entity: 'Surabaya Branch', type: 'Fieldwork', anomalyScore: -0.76, description: 'Audit fieldwork completion time 4.2σ below historical average — potential superficial review', severity: 'High', date: '2026-05-20', amount: 0, frequency: 3, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH },
    { id: 'ANM-005', entity: 'Bali Branch', type: 'Procurement', anomalyScore: -0.72, description: 'Single-source procurement exceeding threshold without required competitive bidding documentation', severity: 'High', date: '2026-06-03', amount: 780000000, frequency: 5, isAnomaly: true, riskLevel: RiskLevel.MODERATE_HIGH },
    { id: 'ANM-006', entity: 'Bandung Branch', type: 'Travel Expense', anomalyScore: -0.68, description: 'Travel reimbursement patterns showing systematic rounding to exact Rp 500K increments', severity: 'Medium', date: '2026-05-10', amount: 45000000, frequency: 18, isAnomaly: true, riskLevel: RiskLevel.MODERATE },
    { id: 'ANM-007', entity: 'IT Dept', type: 'Data Access', anomalyScore: -0.63, description: 'Unusual data export volume from CRM system — 3x average daily export size', severity: 'Medium', date: '2026-06-05', amount: 0, frequency: 7, isAnomaly: true, riskLevel: RiskLevel.MODERATE },
    { id: 'ANM-008', entity: 'Operations Dept', type: 'Inventory', anomalyScore: -0.55, description: 'Inventory adjustment entries without supporting documentation in warehouse system', severity: 'Medium', date: '2026-05-22', amount: 120000000, frequency: 9, isAnomaly: true, riskLevel: RiskLevel.MODERATE },
  ]

  // Generate scatter plot data: normal + anomaly points
  const scatterData: ScatterPoint[] = []
  // Normal points
  for (let i = 0; i < 80; i++) {
    scatterData.push({
      x: Math.round((Math.random() * 40 + 10) * 100) / 100,
      y: Math.round((Math.random() * 30 + 5) * 100) / 100,
      isAnomaly: false,
    })
  }
  // Anomaly points (outliers)
  const anomalyPoints: ScatterPoint[] = [
    { x: 92.5, y: 47.0, isAnomaly: true, label: 'ANM-001' },
    { x: 5.0, y: 156.0, isAnomaly: true, label: 'ANM-002' },
    { x: 85.3, y: 12.0, isAnomaly: true, label: 'ANM-003' },
    { x: 15.0, y: 3.0, isAnomaly: true, label: 'ANM-004' },
    { x: 78.0, y: 5.0, isAnomaly: true, label: 'ANM-005' },
    { x: 45.0, y: 65.0, isAnomaly: true, label: 'ANM-006' },
    { x: 8.0, y: 85.0, isAnomaly: true, label: 'ANM-007' },
    { x: 62.0, y: 9.0, isAnomaly: true, label: 'ANM-008' },
  ]
  scatterData.push(...anomalyPoints)

  const summary: AnomalySummary = {
    totalScanned: 12847,
    anomaliesFound: anomalies.length,
    contaminationRate: 0.062,
    topCategory: 'Transaction',
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
    anomaliesDetected: 8,
    documentsAnalyzed: 10,
    kpiAlerts: 3,
  }
}
