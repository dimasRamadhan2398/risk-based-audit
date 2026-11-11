/**
 * Composable providing simulated/mock data for all 4 AI-powered analytics modules.
 * Data references actual branches, categories, and KPIs from the existing stores.
 */

import { RiskLevel } from '~/types/risk'
import { BANK_ANOMALIES_SEED, BANK_SCATTER_SEED } from './bankAnomalyDataset'

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
  /** Transaction hour (0-23); the scatter's Y axis. */
  hour?: number | null
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
  source: string
  sourceType?: 'Imported Document' | 'Internal Note' | 'External Database' | string
  fileName?: string
  fileType?: string
  extractionMethod?: string
  autoCategory: string
  confidence: number
  sentiment: 'Positive' | 'Neutral' | 'Negative'
  severityScore: number
  controlIndicator?: string
  extractedFinding?: string
  extractedRecommendation?: string
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

export interface AuditorKPIForecast {
  code: string
  kpiName: string
  category: string
  strategicObjective: string
  hibHig: 'HIG' | 'HIB'
  unit: string
  targetHorizon: string
  targetValue: number
  currentValue: number
  forecastedValue: number
  achievementRate: number
  forecastedAchievementRate: number
  currentStatus: 'Good' | 'Moderate' | 'Poor'
  predictedStatus: 'Good' | 'Moderate' | 'Poor'
  trend: 'Improving' | 'Deteriorating' | 'Stable'
  confidenceLower?: number
  confidenceUpper?: number
  aiNarrative?: string
  recommendedAction?: string
}

export interface AuditorAtRiskPillar {
  pillar: string
  code: string
  kpi: string
  target: number
  predicted: number
  achievementRate: number
  predictedStatus: 'Good' | 'Moderate' | 'Poor'
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
  const anomalies: AnomalyRecord[] = [...BANK_ANOMALIES_SEED]
  const scatterData: ScatterPoint[] = [...BANK_SCATTER_SEED]

  const summary: AnomalySummary = {
    totalScanned: 1200,
    anomaliesFound: anomalies.length,
    contaminationRate: 0.125,
    topCategory: 'IT Control',
  }

  return { anomalies, scatterData, summary }
}

export function useIndoBERTData() {
  const documents: NLPDocumentResult[] = [
    {
      docId: 'WP-2026-001',
      title: 'Kertas Kerja Pemeriksaan Kas & Otorisasi Pembayaran Cabang Surabaya',
      source: 'Working Paper',
      sourceType: 'Imported Document',
      fileName: 'KKP_Kas_Cabang_Surabaya_2026.pdf',
      fileType: 'PDF',
      extractionMethod: 'PDF Text & OCR Engine',
      autoCategory: 'Financial',
      confidence: 0.96,
      sentiment: 'Negative',
      severityScore: 91,
      controlIndicator: 'Fraud Indication',
      extractedFinding: 'Ditemukan 14 transaksi pembayaran vendor senilai Rp 1,85 Miliar yang dipecah (split transaction) untuk menghindari batas otorisasi Pimpinan Cabang dengan bukti penerimaan barang yang terindikasi fiktif.',
      extractedRecommendation: 'Lakukan audit investigatif khusus atas 14 transaksi vendor tersebut, blokir sementara pembayaran ke rekening vendor terkait, dan terapkan validasi otomatis pendeteksi split invoice pada sistem ERP.',
      date: '2026-06-15',
      excerpt: 'Hasil ekstraksi KKP_Kas_Cabang_Surabaya_2026.pdf: Ditemukan 14 transaksi pembayaran vendor senilai Rp 1,85 Miliar yang dipecah (split transaction) untuk menghindari batas otorisasi Pimpinan Cabang dengan bukti penerimaan barang yang terindikasi fiktif.',
      riskLevel: RiskLevel.HIGH
    },
    {
      docId: 'LHA-2026-002',
      title: 'Laporan Hasil Audit Keamanan Akses Core Banking & Server Database',
      source: 'Audit Result Report',
      sourceType: 'Imported Document',
      fileName: 'LHA_IT_Security_CoreBanking_Q2.docx',
      fileType: 'DOCX',
      extractionMethod: 'OOXML Word Extractor',
      autoCategory: 'Technology',
      confidence: 0.95,
      sentiment: 'Negative',
      severityScore: 84,
      controlIndicator: 'IT & Access Control Weakness',
      extractedFinding: 'Terdapat 23 akun administrator database dan pegawai yang telah mutasi/resign masih berstatus aktif serta ditemukan penggunaan shared account pada server produksi tanpa pencatatan log individual.',
      extractedRecommendation: 'Cabut segera hak akses 23 akun pegawai mutasi/resign, hapus penggunaan shared account pada server produksi, dan integrasikan proses offboarding HR dengan pencabutan akses otomatis (IAM).',
      date: '2026-06-14',
      excerpt: 'Hasil ekstraksi LHA_IT_Security_CoreBanking_Q2.docx: Terdapat 23 akun administrator database dan pegawai yang telah mutasi/resign masih berstatus aktif serta ditemukan penggunaan shared account pada server produksi.',
      riskLevel: RiskLevel.HIGH
    },
    {
      docId: 'EXT-DB-003',
      title: 'Compliance Incident DB: Alert Transaksi Rekening Dormant & Eskalasi PPATK',
      source: 'External DB: Compliance Incident Records',
      sourceType: 'External Database',
      fileName: 'compliance_incident_records.investigator_notes',
      fileType: 'DB_RECORD',
      extractionMethod: 'External Data Source Sync',
      autoCategory: 'Compliance',
      confidence: 0.95,
      sentiment: 'Negative',
      severityScore: 92,
      controlIndicator: 'Fraud Indication',
      extractedFinding: 'Terjadi perputaran dana mencurigakan (rapid turnover of funds) sebesar Rp 4,2 Miliar pada rekening korporasi dormant tanpa pembaruan dokumen Enhanced Due Diligence (EDD) dan berstatus ESCALATED_TO_PPATK.',
      extractedRecommendation: 'Selesaikan pelaporan Laporan Transaksi Keuangan Mencurigakan (LTKM/STR) ke PPATK tepat waktu, lakukan pemblokiran sementara mutasi debit rekening dormant, dan wajibkan verifikasi ulang Beneficial Owner.',
      date: '2026-06-13',
      excerpt: 'Catatan Database Eksternal (compliance_incident_records.investigator_notes): Terjadi perputaran dana mencurigakan (rapid turnover of funds) sebesar Rp 4,2 Miliar pada rekening korporasi dormant tanpa pembaruan dokumen EDD.',
      riskLevel: RiskLevel.HIGH
    },
    {
      docId: 'EXT-DB-004',
      title: 'External Audit Log DB: Override Otorisasi Jurnal Buku Besar Tanpa Dual Control',
      source: 'External DB: User Access Audit Logs',
      sourceType: 'External Database',
      fileName: 'user_access_audit_logs.action_notes',
      fileType: 'DB_RECORD',
      extractionMethod: 'External Data Source Sync',
      autoCategory: 'Financial',
      confidence: 0.94,
      sentiment: 'Negative',
      severityScore: 85,
      controlIndicator: 'Segregation of Duties (SoD) Violation',
      extractedFinding: 'Pejabat yang sama menginisiasi sekaligus mengotorisasi penyesuaian jurnal buku besar (Same user initiated and authorized ledger adjustment) senilai Rp 890 Juta melalui jalur BRANCH_OVERRIDE.',
      extractedRecommendation: 'Terapkan pemisahan fungsi (Maker-Checker) secara sistemik pada modul jurnal buku besar, nonaktifkan fitur self-approval pada parameter BRANCH_OVERRIDE, dan lakukan verifikasi bukti pendukung atas jurnal Rp 890 Juta.',
      date: '2026-06-12',
      excerpt: 'Catatan Database Eksternal (user_access_audit_logs & gl_transactions): Pejabat yang sama menginisiasi sekaligus mengotorisasi penyesuaian jurnal buku besar (SOD_VIOLATION) senilai Rp 890 Juta.',
      riskLevel: RiskLevel.HIGH
    },
    {
      docId: 'NOTE-FW-005',
      title: 'Catatan Observasi Fieldwork: Pemeriksaan Fisik Persediaan & Kontrak Pengadaan Logistik',
      source: 'Internal Note: Fieldwork Observation',
      sourceType: 'Internal Note',
      fileName: 'fieldwork_observations.notes',
      fileType: 'NOTE',
      extractionMethod: 'Internal Feature Note Sync',
      autoCategory: 'Operational',
      confidence: 0.92,
      sentiment: 'Negative',
      severityScore: 74,
      controlIndicator: 'Internal Control Weakness',
      extractedFinding: 'Terdapat keterlambatan pengiriman barang pengadaan logistik rata-rata 18 hari kerja di 4 kantor wilayah serta kartu stok gudang belum direkonsiliasi dengan pencatatan modul persediaan selama 2 kuartal.',
      extractedRecommendation: 'Lakukan stock opname dan rekonsiliasi fisik persediaan setiap akhir bulan, serta kenakan klaim penalti keterlambatan (SLA) sesuai klausul kontrak kepada vendor logistik.',
      date: '2026-06-10',
      excerpt: 'Catatan Observasi Fieldwork: Terdapat keterlambatan pengiriman barang pengadaan logistik rata-rata 18 hari kerja di 4 kantor wilayah serta kartu stok gudang belum direkonsiliasi dengan modul persediaan.',
      riskLevel: RiskLevel.MODERATE_HIGH
    },
    {
      docId: 'PKPT-2026-006',
      title: 'Evaluasi Matriks Pengendalian Risiko (RCM) & Kepatuhan Pelaporan OJK',
      source: 'Annual Audit Plan',
      sourceType: 'Imported Document',
      fileName: 'Evaluasi_RCM_Kepatuhan_2026.xlsx',
      fileType: 'XLSX',
      extractionMethod: 'Spreadsheet Table Extractor',
      autoCategory: 'Compliance',
      confidence: 0.91,
      sentiment: 'Neutral',
      severityScore: 52,
      controlIndicator: 'Compliance Deviation',
      extractedFinding: 'Sebanyak 3 dari 28 kewajiban pelaporan berkala regulator memerlukan pembaruan format template sesuai SEOJK terbaru yang mulai berlaku pada Semester II.',
      extractedRecommendation: 'Perbarui template otomasi pelaporan pada modul Compliance sebelum tenggat waktu Semester II dan lakukan uji coba validasi pelaporan bersama Divisi Kepatuhan.',
      date: '2026-06-08',
      excerpt: 'Hasil ekstraksi Evaluasi_RCM_Kepatuhan_2026.xlsx: Sebanyak 3 dari 28 kewajiban pelaporan berkala regulator memerlukan pembaruan format template sesuai SEOJK terbaru.',
      riskLevel: RiskLevel.MODERATE
    },
    {
      docId: 'QA-2026-007',
      title: 'Laporan Quality Assurance (QAR) Pelaksanaan Standar Audit Internal',
      source: 'Quality Assurance',
      sourceType: 'Imported Document',
      fileName: 'Laporan_QAR_Internal_Audit_2026.pdf',
      fileType: 'PDF',
      extractionMethod: 'PDF Text & OCR Engine',
      autoCategory: 'Operational',
      confidence: 0.94,
      sentiment: 'Positive',
      severityScore: 22,
      controlIndicator: 'Adequate Control',
      extractedFinding: 'Pelaksanaan supervisi berjenjang dan dokumentasi kertas kerja audit pada 94% penugasan telah memadai dan berjalan efektif sesuai Standar Internasional Praktik Profesional Audit Internal.',
      extractedRecommendation: 'Pertahankan standar reviu berjenjang kertas kerja audit dan lanjutkan program sertifikasi berkelanjutan bagi seluruh auditor.',
      date: '2026-06-05',
      excerpt: 'Hasil ekstraksi Laporan_QAR_Internal_Audit_2026.pdf: Pelaksanaan supervisi berjenjang dan dokumentasi kertas kerja audit pada 94% penugasan telah memadai dan berjalan efektif sesuai standar.',
      riskLevel: RiskLevel.LOW
    }
  ]

  const sentimentDistribution: SentimentDistribution = {
    positive: 1,
    neutral: 1,
    negative: 5,
  }

  const categoryDistribution: Record<string, number> = {
    'Financial': 2,
    'Operational': 2,
    'Compliance': 2,
    'Technology': 1,
  }

  const controlIndicatorDistribution: Record<string, number> = {
    'Fraud Indication': 2,
    'Internal Control Weakness': 1,
    'Segregation of Duties (SoD) Violation': 1,
    'IT & Access Control Weakness': 1,
    'Compliance Deviation': 1,
    'Adequate Control': 1,
  }

  const sourceTypeDistribution: Record<string, number> = {
    'Imported Document': 4,
    'Internal Note': 1,
    'External Database': 2,
  }

  const taxonomy = {
    risk_categories: [
      { name: 'Financial', is_default: true },
      { name: 'Operational', is_default: true },
      { name: 'Compliance', is_default: true },
      { name: 'Technology', is_default: true }
    ],
    control_indicators: [
      { name: 'Fraud Indication', is_default: true },
      { name: 'Internal Control Weakness', is_default: true },
      { name: 'Segregation of Duties (SoD) Violation', is_default: true },
      { name: 'IT & Access Control Weakness', is_default: true },
      { name: 'Compliance Deviation', is_default: true },
      { name: 'Adequate Control', is_default: true }
    ]
  }

  return { documents, sentimentDistribution, categoryDistribution, controlIndicatorDistribution, sourceTypeDistribution, taxonomy }
}

/** Monthly periods ("YYYY-MM") ending at `last`, oldest first. */
const monthsBack = (last: string, count: number): string[] => {
  const [y, m] = last.split('-').map(Number) as [number, number]
  const base = y * 12 + (m - 1)
  return Array.from({ length: count }, (_, i) => {
    const idx = base - (count - 1 - i)
    return `${Math.floor(idx / 12)}-${String((idx % 12) + 1).padStart(2, '0')}`
  })
}

const monthsAhead = (last: string, count: number): string[] => {
  const [y, m] = last.split('-').map(Number) as [number, number]
  const base = y * 12 + (m - 1)
  return Array.from({ length: count }, (_, i) => {
    const idx = base + i + 1
    return `${Math.floor(idx / 12)}-${String((idx % 12) + 1).padStart(2, '0')}`
  })
}

const seedSeries = (last: string, actual: number[], forecast: number[]) => ({
  lastActualPeriod: last,
  history: monthsBack(last, actual.length).map((period, i) => ({ period, value: actual[i]! })),
  forecast: monthsAhead(last, forecast.length).map((period, i) => ({ period, value: forecast[i]! })),
})

/**
 * Offline seed, same shape as /performance-trend/batch: every KPI carries its
 * own monthly history and a forecast that starts the month after its last
 * actual value. Units follow utils/analyticsFormat (NPS is a score, CLV rupiah).
 */
export function useTimeSeriesData() {
  const LAST = '2025-12'
  const kpiForecasts: (KPIForecast & { lastActualPeriod: string, history: { period: string, value: number }[], forecast: { period: string, value: number }[] })[] = [
    {
      kpiName: 'Net Promoter Score', code: 'KPI-001', unit: 'score', entity: 'Jakarta Branch', entityType: 'Branch', targetHorizon: '2026-01..2026-03',
      currentValue: 47.7, forecastedValue: 47.2, trend: 'Declining', riskLevel: RiskLevel.MODERATE,
      recommendedAction: 'Lakukan survey mendalam kepuasan pelanggan — NPS menurun',
      ...seedSeries(LAST, [55.1, 53.8, 54.6, 52.9, 51.7, 52.4, 50.8, 49.9, 50.6, 48.8, 47.0, 47.7], [47.2, 46.6, 46.1]),
    },
    {
      kpiName: 'Customer Lifetime Value', code: 'KPI-002', unit: 'Rp', entity: 'Jakarta Branch', entityType: 'Branch', targetHorizon: '2026-01..2026-03',
      currentValue: 650.3, forecastedValue: 649.0, trend: 'Declining', riskLevel: RiskLevel.MODERATE,
      recommendedAction: 'Investigasi faktor penurunan lifetime value',
      ...seedSeries(LAST, [701.2, 694.5, 688.0, 690.7, 682.3, 676.9, 671.4, 668.2, 662.5, 659.8, 655.1, 650.3], [649.0, 647.7, 646.3]),
    },
    {
      kpiName: 'Sales Conversion Rate', code: 'KPI-003', unit: '%', entity: 'Bandung Branch', entityType: 'Branch', targetHorizon: '2026-01..2026-03',
      currentValue: 12.0, forecastedValue: 12.2, trend: 'Improving', riskLevel: RiskLevel.LOW,
      recommendedAction: 'Monitor kualitas lead dan efektivitas tim sales',
      ...seedSeries(LAST, [10.8, 11.0, 11.3, 10.9, 11.4, 11.6, 11.2, 11.7, 11.9, 11.5, 11.8, 12.0], [12.2, 12.3, 12.45]),
    },
    {
      kpiName: 'Customer Acquisition Cost', code: 'KPI-004', unit: 'Rp', entity: 'Jakarta Branch', entityType: 'Branch', targetHorizon: '2026-01..2026-03',
      currentValue: 45.5, forecastedValue: 45.2, trend: 'Improving', riskLevel: RiskLevel.LOW,
      recommendedAction: 'Lanjutkan optimasi funnel pemasaran',
      ...seedSeries(LAST, [48.2, 47.9, 48.5, 47.1, 46.8, 47.3, 46.4, 46.9, 46.0, 45.8, 46.1, 45.5], [45.2, 44.9, 44.6]),
    },
    {
      kpiName: 'Gross Profit Margin', code: 'KPI-005', unit: '%', entity: 'Finance Dept', entityType: 'Department', targetHorizon: '2026-01..2026-03',
      currentValue: 39.1, forecastedValue: 39.3, trend: 'Improving', riskLevel: RiskLevel.LOW,
      recommendedAction: 'Pertahankan kontrol biaya dan evaluasi strategi pricing',
      ...seedSeries(LAST, [37.2, 37.8, 38.1, 37.5, 38.4, 38.0, 38.6, 38.9, 38.3, 38.8, 39.0, 39.1], [39.3, 39.5, 39.7]),
    },
    {
      kpiName: 'Customer Retention Rate', code: 'KPI-006', unit: '%', entity: 'Bandung Branch', entityType: 'Branch', targetHorizon: '2026-01..2026-03',
      currentValue: 91.7, forecastedValue: 92.1, trend: 'Improving', riskLevel: RiskLevel.LOW,
      recommendedAction: 'Pertahankan program loyalitas dan kualitas layanan',
      ...seedSeries(LAST, [89.4, 89.9, 90.2, 89.8, 90.5, 90.9, 90.6, 91.1, 91.4, 91.0, 91.5, 91.7], [92.1, 92.5, 92.87]),
    },
  ]

  // Legacy single-series shape: first KPI, forecast line anchored at the last actual.
  const first = kpiForecasts[0]!
  const historicalKPI: TimeSeriesPoint[] = [
    ...first.history.map((p, i) => {
      const isLast = i === first.history.length - 1
      return { period: p.period, actual: p.value, forecast: isLast ? p.value : null, upperBound: isLast ? p.value : null, lowerBound: isLast ? p.value : null }
    }),
    ...first.forecast.map((p) => ({ period: p.period, actual: null, forecast: p.value, upperBound: null, lowerBound: null })),
  ]

  const atRiskDepartments: AtRiskDepartment[] = [
    { department: 'Jakarta Branch', kpi: 'Net Promoter Score', currentTrend: -3.3, predictedQ3: 47.2, riskLevel: RiskLevel.MODERATE },
    { department: 'Jakarta Branch', kpi: 'Customer Lifetime Value', currentTrend: -0.6, predictedQ3: 649.0, riskLevel: RiskLevel.MODERATE },
  ]

  const forecastAccuracy = {
    mape: 3.82,
    rmse: 1.45,
    r2Score: 0.942,
  }

  const initialAuditorKPIs: AuditorKPIForecast[] = [
    {
      code: 'SO-IA01',
      kpiName: 'Presentase Penyelesaian PKAT (Annual Audit Plan)',
      category: 'Pelaksanaan Audit & PKAT',
      strategicObjective: 'Pelaksanaan Audit & PKAT',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 100.0,
      currentValue: 85.0,
      forecastedValue: 82.4,
      achievementRate: 85.0,
      forecastedAchievementRate: 82.4,
      currentStatus: 'Moderate',
      predictedStatus: 'Moderate',
      trend: 'Deteriorating',
      confidenceLower: 78.5,
      confidenceUpper: 86.0,
      aiNarrative: 'Dengan kecepatan saat ini, diprediksi audit plan hanya selesai 82.4% di akhir tahun. Diperlukan akselerasi pada pilar Pelaksanaan Audit & PKAT.',
      recommendedAction: 'Akselerasi penerbitan surat tugas audit dan tingkatkan otomatisasi pengujian substantif dengan CAATT.'
    },
    {
      code: 'SO-IA02',
      kpiName: 'Ketepatan Waktu Penerbitan LHA',
      category: 'Kualitas & Pelaporan Audit',
      strategicObjective: 'Kualitas & Pelaporan Audit',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 90.0,
      currentValue: 88.0,
      forecastedValue: 91.5,
      achievementRate: 97.8,
      forecastedAchievementRate: 101.7,
      currentStatus: 'Moderate',
      predictedStatus: 'Good',
      trend: 'Improving',
      confidenceLower: 88.0,
      confidenceUpper: 94.5,
      aiNarrative: 'Dengan kecepatan saat ini, target Ketepatan Waktu Penerbitan LHA diproyeksikan tercapai sangat baik (101.7%, Good) di akhir tahun.',
      recommendedAction: 'Pertahankan efisiensi reviu berjenjang laporan hasil audit.'
    },
    {
      code: 'SO-IA03',
      kpiName: 'Tindak Lanjut Temuan Audit Selesai',
      category: 'Monitoring & Tindak Lanjut',
      strategicObjective: 'Monitoring & Tindak Lanjut',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 85.0,
      currentValue: 72.0,
      forecastedValue: 68.5,
      achievementRate: 84.7,
      forecastedAchievementRate: 80.6,
      currentStatus: 'Moderate',
      predictedStatus: 'Moderate',
      trend: 'Deteriorating',
      confidenceLower: 65.0,
      confidenceUpper: 72.0,
      aiNarrative: 'Dengan kecepatan saat ini, diprediksi target Tindak Lanjut Temuan Audit Selesai tercapai 80.6% (Moderate) di akhir tahun.',
      recommendedAction: 'Eskalasi komitmen auditee kepada Direksi untuk percepatan penyelesaian komitmen perbaikan jatuh tempo.'
    },
    {
      code: 'SO-IA04',
      kpiName: 'Utilisasi Jam Kerja Audit Efektif',
      category: 'Efisiensi Sumber Daya & Operasional',
      strategicObjective: 'Efisiensi Sumber Daya & Operasional',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 80.0,
      currentValue: 82.5,
      forecastedValue: 84.0,
      achievementRate: 103.1,
      forecastedAchievementRate: 105.0,
      currentStatus: 'Good',
      predictedStatus: 'Good',
      trend: 'Improving',
      confidenceLower: 81.0,
      confidenceUpper: 87.0,
      aiNarrative: 'Dengan kecepatan saat ini, target Utilisasi Jam Kerja Audit Efektif diproyeksikan tercapai sangat baik (105.0%, Good) di akhir tahun.',
      recommendedAction: 'Pertahankan alokasi jam kerja audit langsung di atas 80% per auditor per tahun.'
    },
    {
      code: 'SO-IA05',
      kpiName: 'Rata-rata Hari Penyelesaian Fieldwork',
      category: 'Pelaksanaan Audit & PKAT',
      strategicObjective: 'Pelaksanaan Audit & PKAT',
      hibHig: 'HIB',
      unit: 'Hari',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 15.0,
      currentValue: 16.5,
      forecastedValue: 17.8,
      achievementRate: 90.9,
      forecastedAchievementRate: 84.3,
      currentStatus: 'Moderate',
      predictedStatus: 'Moderate',
      trend: 'Deteriorating',
      confidenceLower: 16.0,
      confidenceUpper: 19.5,
      aiNarrative: 'Dengan kecepatan saat ini, diprediksi rata-rata fieldwork membutuhkan 17.8 hari (ketercapaian target 84.3%, Moderate).',
      recommendedAction: 'Optimalkan preliminary survey dan perjelas batasan audit scope sebelum fieldwork dimulai.'
    },
    {
      code: 'SO-IA06',
      kpiName: 'Cakupan Audit Berbasis Risiko (RBIA Coverage)',
      category: 'Tata Kelola Perusahaan & GCG',
      strategicObjective: 'Tata Kelola Perusahaan & GCG',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 100.0,
      currentValue: 95.0,
      forecastedValue: 96.0,
      achievementRate: 95.0,
      forecastedAchievementRate: 96.0,
      currentStatus: 'Moderate',
      predictedStatus: 'Moderate',
      trend: 'Improving',
      confidenceLower: 93.0,
      confidenceUpper: 98.0,
      aiNarrative: 'Dengan kecepatan saat ini, target Cakupan Audit Berbasis Risiko diproyeksikan tercapai 96.0% (Moderate) di akhir tahun.',
      recommendedAction: 'Prioritaskan audit entitas auditee berkategori risiko High dan Critical.'
    },
    {
      code: 'SO-IA07',
      kpiName: 'Indeks Kepuasan Auditee (CSAT Audit)',
      category: 'Kualitas & Pelaporan Audit',
      strategicObjective: 'Kualitas & Pelaporan Audit',
      hibHig: 'HIG',
      unit: 'Score',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 85.0,
      currentValue: 86.5,
      forecastedValue: 87.2,
      achievementRate: 101.8,
      forecastedAchievementRate: 102.6,
      currentStatus: 'Good',
      predictedStatus: 'Good',
      trend: 'Improving',
      confidenceLower: 84.5,
      confidenceUpper: 89.5,
      aiNarrative: 'Dengan kecepatan saat ini, target Indeks Kepuasan Auditee diproyeksikan tercapai sangat baik (102.6%, Good) di akhir tahun.',
      recommendedAction: 'Tingkatkan diskusi rekomendasi bernilai tambah (value-added audit findings).'
    },
    {
      code: 'SO-IA08',
      kpiName: 'Rasio Temuan Berulang (Repeat Findings)',
      category: 'Monitoring & Tindak Lanjut',
      strategicObjective: 'Monitoring & Tindak Lanjut',
      hibHig: 'HIB',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 5.0,
      currentValue: 6.8,
      forecastedValue: 7.5,
      achievementRate: 73.5,
      forecastedAchievementRate: 66.7,
      currentStatus: 'Moderate',
      predictedStatus: 'Poor',
      trend: 'Deteriorating',
      confidenceLower: 6.8,
      confidenceUpper: 8.5,
      aiNarrative: 'Dengan kecepatan saat ini, diprediksi target Rasio Temuan Berulang hanya tercapai 66.7% (Poor) di akhir tahun.',
      recommendedAction: 'Lakukan root-cause analysis mendalam dan reviu pengendalian internal preventif di unit bersangkutan.'
    },
    {
      code: 'SO-IA09',
      kpiName: 'Keikutsertaan Pelatihan & Sertifikasi Auditor',
      category: 'Pengembangan Kompetensi SDM',
      strategicObjective: 'Pengembangan Kompetensi SDM',
      hibHig: 'HIG',
      unit: 'Jam',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 40.0,
      currentValue: 42.0,
      forecastedValue: 45.0,
      achievementRate: 105.0,
      forecastedAchievementRate: 112.5,
      currentStatus: 'Good',
      predictedStatus: 'Good',
      trend: 'Improving',
      confidenceLower: 42.0,
      confidenceUpper: 48.0,
      aiNarrative: 'Dengan kecepatan saat ini, target Pelatihan & Sertifikasi Auditor diproyeksikan tercapai sangat baik (112.5%, Good) di akhir tahun.',
      recommendedAction: 'Fasilitasi sertifikasi profesi CIA/CISA/CRMA bagi auditor senior.'
    },
    {
      code: 'SO-IA10',
      kpiName: 'Implementasi Continuous Auditing & CAATT',
      category: 'Teknologi & Otomasi Audit',
      strategicObjective: 'Teknologi & Otomasi Audit',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 75.0,
      currentValue: 70.0,
      forecastedValue: 78.0,
      achievementRate: 93.3,
      forecastedAchievementRate: 104.0,
      currentStatus: 'Moderate',
      predictedStatus: 'Good',
      trend: 'Improving',
      confidenceLower: 73.0,
      confidenceUpper: 82.0,
      aiNarrative: 'Dengan kecepatan saat ini, target Implementasi Continuous Auditing diproyeksikan tercapai sangat baik (104.0%, Good) di akhir tahun.',
      recommendedAction: 'Perluas script audit terotomasi pada data transaksi pinjaman dan operasional cabang.'
    },
    {
      code: 'SO-IA11',
      kpiName: 'Validitas Temuan Deviasi Material & Fraud',
      category: 'Tata Kelola Perusahaan & GCG',
      strategicObjective: 'Tata Kelola Perusahaan & GCG',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 100.0,
      currentValue: 100.0,
      forecastedValue: 100.0,
      achievementRate: 100.0,
      forecastedAchievementRate: 100.0,
      currentStatus: 'Good',
      predictedStatus: 'Good',
      trend: 'Stable',
      confidenceLower: 98.0,
      confidenceUpper: 100.0,
      aiNarrative: 'Dengan kecepatan saat ini, target Validitas Temuan Fraud diproyeksikan tercapai 100.0% (Good) di akhir tahun.',
      recommendedAction: 'Pertahankan standar bukti audit forensik dan kecukupan dokumen pembuktian.'
    },
    {
      code: 'SO-IA12',
      kpiName: 'Kepatuhan terhadap Standar Profesi (QAR Internal)',
      category: 'Kualitas & Pelaporan Audit',
      strategicObjective: 'Kualitas & Pelaporan Audit',
      hibHig: 'HIG',
      unit: '%',
      targetHorizon: 'Akhir Tahun 2026',
      targetValue: 95.0,
      currentValue: 92.0,
      forecastedValue: 90.5,
      achievementRate: 96.8,
      forecastedAchievementRate: 95.3,
      currentStatus: 'Moderate',
      predictedStatus: 'Moderate',
      trend: 'Deteriorating',
      confidenceLower: 87.0,
      confidenceUpper: 93.5,
      aiNarrative: 'Dengan kecepatan saat ini, target Kepatuhan Standar Profesi diproyeksikan tercapai 95.3% (Moderate) di akhir tahun.',
      recommendedAction: 'Tingkatkan pelaksanaan Quality Assurance and Improvement Program (QAIP) berkala.'
    }
  ]

  const initialAuditorTimeSeries: TimeSeriesPoint[] = [
    { period: 'Q1 2025', actual: 88.5, forecast: 88.5, upperBound: 91.0, lowerBound: 86.0 },
    { period: 'Q2 2025', actual: 90.2, forecast: 90.2, upperBound: 93.0, lowerBound: 87.5 },
    { period: 'Q3 2025', actual: 89.0, forecast: 89.0, upperBound: 92.0, lowerBound: 86.0 },
    { period: 'Q4 2025', actual: 91.8, forecast: 91.8, upperBound: 94.5, lowerBound: 89.0 },
    { period: 'Q1 2026 (Pred)', actual: null, forecast: 86.4, upperBound: 89.5, lowerBound: 83.0 },
    { period: 'Akhir Tahun 2026 (Pred)', actual: null, forecast: 82.4, upperBound: 86.0, lowerBound: 78.5 }
  ]

  const initialAuditorAtRiskPillars: AuditorAtRiskPillar[] = [
    {
      pillar: 'Pelaksanaan Audit & PKAT',
      code: 'SO-IA01',
      kpi: 'Presentase Penyelesaian PKAT (Annual Audit Plan)',
      target: 100.0,
      predicted: 82.4,
      achievementRate: 82.4,
      predictedStatus: 'Moderate'
    },
    {
      pillar: 'Monitoring & Tindak Lanjut',
      code: 'SO-IA08',
      kpi: 'Rasio Temuan Berulang (Repeat Findings)',
      target: 5.0,
      predicted: 7.5,
      achievementRate: 66.7,
      predictedStatus: 'Poor'
    },
    {
      pillar: 'Pelaksanaan Audit & PKAT',
      code: 'SO-IA05',
      kpi: 'Rata-rata Hari Penyelesaian Fieldwork',
      target: 15.0,
      predicted: 17.8,
      achievementRate: 84.3,
      predictedStatus: 'Moderate'
    }
  ]

  return {
    historicalKPI,
    kpiForecasts,
    atRiskDepartments,
    forecastAccuracy,
    auditorKPIForecasts: initialAuditorKPIs,
    auditorTimeSeries: initialAuditorTimeSeries,
    auditorAtRiskPillars: initialAuditorAtRiskPillars
  }
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
    { stratumLabel: 'Rp 10 Juta - 50 Juta', minValue: 10000000, maxValue: 50000000, category: 'Commercial', trxCount: 2840, totalAmount: 71000000000, pctCount: 22.7, pctAmount: 17.7 },
    { stratumLabel: 'Rp 50 Juta - 100 Juta', minValue: 50000000, maxValue: 100000000, category: 'Corporate', trxCount: 890, totalAmount: 66750000000, pctCount: 7.1, pctAmount: 16.6 },
    { stratumLabel: 'Rp 100 Juta - 500 Juta', minValue: 100000000, maxValue: 500000000, category: 'High Value', trxCount: 280, totalAmount: 84000000000, pctCount: 2.2, pctAmount: 20.9 },
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

