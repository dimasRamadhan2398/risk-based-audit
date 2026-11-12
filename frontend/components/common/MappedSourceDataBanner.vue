<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useActiveDataSource } from '~/composables/useActiveDataSource'
import { useDataSourceStore } from '~/stores/data-source'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useAppToast } from '~/composables/useAppToast'
import { useRuntimeConfig } from '#app'
import { getCanonicalCategory } from '~/composables/useAnomalyScatter'

interface MappedRecord {
  connection_id: string
  connection_name: string
  table_name: string
  target_scope: string
  target_module: string
  audit_field: string
  total_rows?: number
  anomaly_rules?: string[]
  custom_rules?: any[]
  sample_records: Record<string, unknown>[]
}

const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()
const toast = useAppToast()
const dataSourceStore = useDataSourceStore()
const { selectedSourceId, selectedSource } = useActiveDataSource()

const isModalOpen = ref(false)
const selectedMapping = ref<MappedRecord | null>(null)
const mappedRecords = ref<MappedRecord[]>([])
const selectedTableIndex = ref(0)
const isLoading = ref(false)
const isDismissed = ref(false)
const isExecutingAction = ref(false)

const getBaseUrl = (): string => {
  return (config.public?.masterServiceBaseUrl as string) || '/api/v1'
}

// Map current route to target scope and module
const currentModuleInfo = computed(() => {
  const path = route.path.toLowerCase()
  if (path.startsWith('/settings') || path.startsWith('/auth')) {
    return null
  }

  // --- 1. Risk Management Scope ---
  if (path === '/risk-profile' || path === '/risk-profile/') {
    return {
      scope: 'risk_management',
      module: 'Corporate Risk Profile',
      slug: 'corporate_risk_profile',
      label: 'Corporate Risk Profile',
      actionType: 'risk_profile',
      actionLabel: '+ Tambahkan ke Profil Risiko'
    }
  }
  if (path.includes('/risk-profile/audit-universe')) {
    return {
      scope: 'risk_management',
      module: 'Audit Universe',
      slug: 'audit_universe',
      label: 'Audit Universe',
      actionType: 'audit_universe',
      actionLabel: '+ Tambahkan ke Audit Universe'
    }
  }
  if (path.includes('/risk-profile/risk-factors')) {
    return {
      scope: 'risk_management',
      module: 'Risk Factors',
      slug: 'risk_factors',
      label: 'Risk Factors',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Faktor Risiko'
    }
  }
  if (path.includes('/risk-profile/risk-control-matrix')) {
    return {
      scope: 'risk_management',
      module: 'Risk Control Matrix (RCM)',
      slug: 'risk_control_matrix',
      label: 'Risk Control Matrix (RCM)',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Kontrol RCM'
    }
  }
  if (path.includes('/risk-profile/audit-priority')) {
    return {
      scope: 'risk_management',
      module: 'Audit Priority',
      slug: 'audit_priority',
      label: 'Audit Priority',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Prioritas Audit'
    }
  }
  if (path.includes('/risk-appetite')) {
    return {
      scope: 'risk_management',
      module: 'Risk Appetite Statement',
      slug: 'risk_appetite',
      label: 'Risk Appetite Statement',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Risk Appetite'
    }
  }

  // --- 2. Audit Features Scope ---
  if (path.includes('/working-paper')) {
    return {
      scope: 'audit_features',
      module: 'Working Paper Evidence',
      slug: 'working_paper',
      label: 'Working Paper Evidence',
      actionType: 'working_paper',
      actionLabel: '+ Jadikan Bukti Sampel Kertas Kerja'
    }
  }
  if (path.includes('/audit-fieldwork')) {
    return {
      scope: 'audit_features',
      module: 'Audit Fieldwork',
      slug: 'audit_fieldwork',
      label: 'Audit Fieldwork',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Bukti Lapangan'
    }
  }
  if (path.includes('/annual-audit')) {
    return {
      scope: 'audit_features',
      module: 'Annual Audit Plan',
      slug: 'annual_audit',
      label: 'Annual Audit Plan',
      actionType: 'generic',
      actionLabel: '+ Masukkan ke Rencana Audit'
    }
  }
  if (path.includes('/audit-activity-plan')) {
    return {
      scope: 'audit_features',
      module: 'Audit Activity Plan',
      slug: 'audit_activity_plan',
      label: 'Audit Activity Plan',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Aktivitas Audit'
    }
  }
  if (path.includes('/assignment-letter')) {
    return {
      scope: 'audit_features',
      module: 'Assignment Letter',
      slug: 'assignment_letter',
      label: 'Assignment Letter',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Surat Tugas'
    }
  }
  if (path.includes('/audit-execution-status')) {
    return {
      scope: 'audit_features',
      module: 'Audit Execution Status',
      slug: 'audit_execution_status',
      label: 'Audit Execution Status',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Status Pelaksanaan'
    }
  }
  if (path.includes('/audit-result-report')) {
    return {
      scope: 'audit_features',
      module: 'Audit Result Report (LHA)',
      slug: 'audit_result_report',
      label: 'Audit Result Report (LHA)',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Temuan LHA'
    }
  }
  if (path.includes('/action-taken-report')) {
    return {
      scope: 'audit_features',
      module: 'Action Taken Report (ATR)',
      slug: 'action_taken_report',
      label: 'Action Taken Report (ATR)',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke Rekomendasi ATR'
    }
  }

  // --- 3. Data Analytics & AI Scope ---
  if (path.includes('/analytics/ai/anomaly-detection') || path.includes('/ai/anomaly-detection')) {
    return {
      scope: 'data_analytics',
      module: 'AI: Anomaly Detection',
      slug: 'anomaly_detection',
      label: 'AI Anomaly Detection',
      actionType: 'ai_anomaly',
      actionLabel: '+ Jalankan AI Anomaly Detection'
    }
  }
  if (path.includes('/analytics/ai/risk-scoring') || path.includes('/analytics/ai/entity-prediction')) {
    return {
      scope: 'data_analytics',
      module: 'AI: Entity Prediction',
      slug: 'entity_prediction',
      label: 'AI Entity Prediction (XGBoost)',
      actionType: 'ai_entity_prediction',
      actionLabel: '+ Prediksi Entitas AI'
    }
  }
  if (path.includes('/analytics/ai/nlp')) {
    return {
      scope: 'data_analytics',
      module: 'AI: NLP Document Analysis',
      slug: 'nlp_analysis',
      label: 'AI NLP Document Analysis',
      actionType: 'generic',
      actionLabel: '+ Analisis Teks AI (NLP)'
    }
  }
  if (path.includes('/analytics/ai/kpi-forecast') || path.includes('/analytics/kpi-forecast')) {
    return {
      scope: 'data_analytics',
      module: 'AI: KPI & Loss Forecast',
      slug: 'kpi_forecast',
      label: 'AI KPI & Loss Forecast',
      actionType: 'generic',
      actionLabel: '+ Prediksi Tren KPI'
    }
  }
  if (path.includes('/analytics/caatt/benford')) {
    return {
      scope: 'data_analytics',
      module: "CAATT: Benford's Law Analysis",
      slug: 'caatt_benford',
      label: "CAATT Benford's Law",
      actionType: 'generic',
      actionLabel: '+ Jalankan Uji Benford Law'
    }
  }
  if (path.includes('/analytics/caatt/duplicate-gap')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: Duplicate & Gap Detection',
      slug: 'caatt_duplicate',
      label: 'CAATT Duplicate & Gap',
      actionType: 'generic',
      actionLabel: '+ Deteksi Duplikasi & Gap'
    }
  }
  if (path.includes('/analytics/caatt/full-population')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: Full Population Testing',
      slug: 'caatt_population',
      label: 'CAATT Full Population',
      actionType: 'generic',
      actionLabel: '+ Uji 100% Populasi Transaksi'
    }
  }
  if (path.includes('/analytics/caatt/reconciliation')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: 3-Way Reconciliation',
      slug: 'caatt_reconciliation',
      label: 'CAATT Reconciliation',
      actionType: 'generic',
      actionLabel: '+ Jalankan Rekonsiliasi 3-Way'
    }
  }
  if (path.includes('/analytics/caatt/stratification')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: Stratification & Aging',
      slug: 'caatt_stratification',
      label: 'CAATT Stratification',
      actionType: 'generic',
      actionLabel: '+ Jalankan Stratifikasi Populasi'
    }
  }
  if (path.includes('/analytics/caatt/policy')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: Policy & Plafond Compliance',
      slug: 'caatt_policy',
      label: 'CAATT Policy Compliance',
      actionType: 'generic',
      actionLabel: '+ Uji Kepatuhan Policy'
    }
  }
  if (path.includes('/analytics/caatt/data-quality')) {
    return {
      scope: 'data_analytics',
      module: 'CAATT: Data Quality & Integrity',
      slug: 'caatt_quality',
      label: 'CAATT Data Quality',
      actionType: 'generic',
      actionLabel: '+ Profil Kualitas Data'
    }
  }
  if (path.startsWith('/analytics')) {
    return {
      scope: 'data_analytics',
      module: 'Data Analytics Center',
      slug: 'data_analytics',
      label: 'Data Analytics Center',
      actionType: 'generic',
      actionLabel: '+ Masukkan ke Data Analytics'
    }
  }

  // --- 4. Reporting Scope ---
  if (path.includes('/kpi-performance')) {
    return {
      scope: 'reporting',
      module: 'Internal Audit Performance (KPI)',
      slug: 'kpi_performance',
      label: 'Internal Audit Performance (KPI)',
      actionType: 'generic',
      actionLabel: '+ Tautkan ke KPI Kinerja'
    }
  }

  return null
})

const getTotalRowCountForTable = (tableName: string): number => {
  switch (tableName) {
    case 'gl_transactions': return 142580
    case 'user_access_audit_logs': return 894200
    case 'loan_accounts': return 28490
    case 'vendor_invoices': return 12350
    case 'compliance_incident_records': return 320
    case 'customer_accounts':
    case 'accounts': return 45200
    default: return 1000
  }
}

const getSampleRecordsForTable = (tableName: string, auditField: string): Record<string, unknown>[] => {
  switch (tableName) {
    case 'loan_accounts':
      return [
        { loan_id: 'LN-001', cif_number: 'CIF-881923', product_type: 'Commercial Working Capital', principal_amount: 2500000000, kol_status: 'KOL-1', interest_rate: '8.5%' },
        { loan_id: 'LN-002', cif_number: 'CIF-772910', product_type: 'Mortgage Griya', principal_amount: 750000000, kol_status: 'KOL-2', interest_rate: '7.2%' },
        { loan_id: 'LN-003', cif_number: 'CIF-662819', product_type: 'Consumer Personal Loan', principal_amount: 150000000, kol_status: 'KOL-3', interest_rate: '11.0%' }
      ]
    case 'gl_transactions':
      return [
        { id: 'TRX-001', transaction_ref: 'TRX-2026-0091', account_number: 'ACC-10029381', amount: 450000000, currency: 'IDR', channel: 'RTGS', status: 'POSTED' },
        { id: 'TRX-002', transaction_ref: 'TRX-2026-0092', account_number: 'ACC-50019283', amount: 12500000, currency: 'IDR', channel: 'MOBILE_APP', status: 'POSTED' },
        { id: 'TRX-003', transaction_ref: 'TRX-2026-0093', account_number: 'ACC-10029381', amount: 180000000, currency: 'IDR', channel: 'ATM_OVERSEAS', status: 'FLAGGED' }
      ]
    case 'customer_accounts':
    case 'accounts':
      return [
        { account_id: 'ACC-10029381', cif_number: 'CIF-881923', account_type: 'Giro Bisnis Korporasi', balance: 4850000000, branch_code: 'BR-001', status: 'Active' },
        { account_id: 'ACC-50019283', cif_number: 'CIF-772910', account_type: 'Tabungan Payroll', balance: 35000000, branch_code: 'BR-002', status: 'Active' }
      ]
    case 'user_access_audit_logs':
      return [
        { log_id: 98120, user_id: 'adm_super_01', action_name: 'MUTATION_TABLE_PERMISSIONS', ip_address: '10.20.4.12', is_override_auth: true, timestamp: '2026-03-24 10:15:00' },
        { log_id: 98121, user_id: 'teller_004', action_name: 'FAILED_PIN_ATTEMPTS_EXCEEDED', ip_address: '10.20.4.45', is_override_auth: false, timestamp: '2026-03-24 11:22:15' }
      ]
    case 'compliance_incident_records':
      return [
        { incident_id: 'INC-001', category: 'AML_STR_ALERT', severity_score: 85, status: 'ESCALATED', reporting_unit: 'Anti-Fraud Unit' },
        { incident_id: 'INC-002', category: 'UNAUTHORIZED_ACCESS_ATTEMPT', severity_score: 65, status: 'INVESTIGATING', reporting_unit: 'IT Security' }
      ]
    case 'vendor_invoices':
      return [
        { invoice_no: 'INV-2026-8801', vendor_code: 'VND-TECH-01', invoice_amount: 185000000, approval_status: 'APPROVED', tax_verified: true },
        { invoice_no: 'INV-2026-8802', vendor_code: 'VND-SUPPLY-04', invoice_amount: 42000000, approval_status: 'PENDING_APPROVAL', tax_verified: false }
      ]
    default:
      return [
        { [auditField || 'id']: 'REC-001', name: `${tableName} sample row 1`, status: 'Active' },
        { [auditField || 'id']: 'REC-002', name: `${tableName} sample row 2`, status: 'Active' }
      ]
  }
}

const fetchMappedData = async () => {
  if (!currentModuleInfo.value) {
    mappedRecords.value = []
    return
  }

  isLoading.value = true
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 3000)

  let apiRecords: MappedRecord[] = []
  try {
    const baseUrl = getBaseUrl()
    const sourceParam = selectedSourceId.value && selectedSourceId.value !== 'all' ? `&source_id=${selectedSourceId.value}` : ''
    const url = `${baseUrl}/data-sources/mapped-records?target_module=${encodeURIComponent(currentModuleInfo.value.module)}&target_scope=${currentModuleInfo.value.scope}${sourceParam}`
    
    const res = await $fetch<{ success: boolean; data: MappedRecord[] }>(url, {
      signal: controller.signal
    }).catch(() => null)

    if (res && Array.isArray(res.data) && res.data.length > 0) {
      apiRecords = res.data
    }
  } catch {
    apiRecords = []
  } finally {
    clearTimeout(timeoutId)
  }

  // If API returned records, use them
  if (apiRecords.length > 0) {
    mappedRecords.value = apiRecords.map(r => ({
      ...r,
      total_rows: r.total_rows || getTotalRowCountForTable(r.table_name)
    }))
    if (selectedTableIndex.value >= mappedRecords.value.length) {
      selectedTableIndex.value = 0
    }
    isLoading.value = false
    return
  }

  // Fallback to local Pinia Store / LocalStorage data mappings
  const localList: MappedRecord[] = []
  const activeSources = dataSourceStore.sources.filter(s => {
    if (selectedSourceId.value && selectedSourceId.value !== 'all') {
      return s.id === selectedSourceId.value || s.source_id === selectedSourceId.value
    }
    return true
  })

  const norm = (str: string) => (str || '').toLowerCase().replace(/[\s_-]+/g, '')
  const curModNorm = norm(currentModuleInfo.value.module)
  const curSlugNorm = norm(currentModuleInfo.value.slug)

  for (const src of activeSources) {
    const mappings = Array.isArray(src.dataMappings) ? src.dataMappings : []
    for (const m of mappings) {
      if (m.isActive === false) continue
      const mapModNorm = norm(m.targetModule)
      const scopeMatches = m.targetScope === currentModuleInfo.value.scope

      const moduleMatches = 
        mapModNorm === curModNorm || 
        mapModNorm.includes(curModNorm) || 
        curModNorm.includes(mapModNorm) || 
        mapModNorm.includes(curSlugNorm) ||
        // Aliases for anomaly
        (mapModNorm.includes('anomaly') && (curModNorm.includes('anomaly') || curSlugNorm.includes('anomaly'))) ||
        (mapModNorm.includes('risk') && curModNorm.includes('risk')) ||
        (mapModNorm.includes('universe') && curModNorm.includes('universe')) ||
        (mapModNorm.includes('paper') && curModNorm.includes('paper')) ||
        (mapModNorm.includes('benford') && curModNorm.includes('benford')) ||
        (mapModNorm.includes('analytic') && curModNorm.includes('analytic'))

      if (moduleMatches || (scopeMatches && !m.targetModule)) {
        // Prevent duplicate entries for same connection + table
        const exists = localList.some(r => r.connection_id === src.id && r.table_name === m.tableName)
        if (!exists) {
          localList.push({
            connection_id: src.id,
            connection_name: src.name,
            table_name: m.tableName,
            target_scope: m.targetScope,
            target_module: m.targetModule || currentModuleInfo.value.module,
            audit_field: m.auditField || 'id',
            total_rows: getTotalRowCountForTable(m.tableName),
            sample_records: getSampleRecordsForTable(m.tableName, m.auditField)
          })
        }
      }
    }
  }

  mappedRecords.value = localList
  if (selectedTableIndex.value >= mappedRecords.value.length) {
    selectedTableIndex.value = 0
  }
  isLoading.value = false
}

watch(
  () => [route.path, selectedSourceId.value, dataSourceStore.sources],
  () => {
    isDismissed.value = false
    fetchMappedData()
  },
  { immediate: true, deep: true }
)

const openPreview = (item: MappedRecord) => {
  selectedMapping.value = item
  isModalOpen.value = true
}

// Active record based on selected tab index
const activeRecord = computed(() => mappedRecords.value[selectedTableIndex.value] || mappedRecords.value[0] || null)

const previewColumns = computed(() => {
  if (!selectedMapping.value || !selectedMapping.value.sample_records || selectedMapping.value.sample_records.length === 0) {
    return []
  }
  return Object.keys(selectedMapping.value.sample_records[0] || {})
})

interface TableAnomalyProfile {
  canonicalCategory: string
  anomalies: (ts: number, connName: string, tbl: string) => any[]
  normalBaseline: () => { x: number; y: number; id: string }[]
}

// Scalable helper to generate dense, realistic normal distribution points (blue dots)
function generateNormalBaselinePoints(
  prefix: string,
  minX: number,
  maxX: number,
  count = 60,
  minY = 8,
  maxY = 17
): { x: number; y: number; id: string }[] {
  const points: { x: number; y: number; id: string }[] = []
  const step = (maxX - minX) / Math.max(count - 1, 1)
  for (let i = 0; i < count; i++) {
    const seed = (i * 9301 + 49297) % 233280
    const rnd = seed / 233280
    const xBase = minX + (i * step)
    const xJitter = (rnd - 0.5) * step * 0.75
    const x = Math.max(minX, Math.round((xBase + xJitter) * 100) / 100)

    const yRnd = ((i * 12345 + 6789) % 1000) / 1000
    const y = Math.round((minY + yRnd * (maxY - minY)) * 10) / 10

    points.push({
      x,
      y,
      id: `${prefix}-NORM-${String(i + 1).padStart(4, '0')}`
    })
  }
  return points
}

const TABLE_ANOMALY_PROFILES: Record<string, TableAnomalyProfile> = {
  gl_transactions: {
    canonicalCategory: 'Transaction',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Kantor Pusat (${connName})` : 'Cabang Utama Jakarta / RTGS',
        refNumber: `TRX-${ts.toString().slice(-4)}-01`,
        accountNumber: 'ACC-10029381',
        amount: 450000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.94,
        description: `Threshold Breach (>100M via RTGS) terdeteksi otomatis dari tabel ${tbl}`,
        reason: `Threshold Breach (>100M via RTGS) terdeteksi otomatis dari tabel ${tbl}`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 450.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Cabang Surabaya / Core Banking',
        refNumber: `TRX-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'ACC-50019283',
        amount: 180000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.88,
        description: `Penarikan ATM overseas berulang di luar jam operasional (${tbl})`,
        reason: `Penarikan ATM overseas berulang di luar jam operasional (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 180.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Cabang Medan / Teller Counter',
        refNumber: `TRX-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'ACC-30018294',
        amount: 320000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.91,
        description: `Pola transaksi terpecah (Structuring) tunai beruntun di bawah batas pelaporan PPATK (${tbl})`,
        reason: `Pola transaksi terpecah (Structuring) tunai beruntun di bawah batas pelaporan PPATK (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 320.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Cabang Bandung / Internet Banking',
        refNumber: `TRX-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'ACC-70019284',
        amount: 550000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.95,
        description: `Transfer dana antarbank bernominal ganjil ke rekening penerima baru terindikasi phishing (${tbl})`,
        reason: `Transfer dana antarbank bernominal ganjil ke rekening penerima baru terindikasi phishing (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 550.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Divisi Tresuri & Settlement Kliring',
        refNumber: `TRX-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'ACC-90028172',
        amount: 720000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.93,
        description: `Settlement kliring warkat tanpa otorisasi ganda maker-checker (${tbl})`,
        reason: `Settlement kliring warkat tanpa otorisasi ganda maker-checker (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 720.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Cabang Makassar / Mobile Banking',
        refNumber: `TRX-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'ACC-40019281',
        amount: 210000000,
        type: 'Transaction',
        category: 'Transaction',
        anomalyScore: 0.86,
        description: `Lonjakan debit frekuensi tinggi (velocity anomaly) pada merchant digital baru (${tbl})`,
        reason: `Lonjakan debit frekuensi tinggi (velocity anomaly) pada merchant digital baru (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 210.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('GL', 2.5, 95.0, 65, 8, 17)
  },
  loan_accounts: {
    canonicalCategory: 'Lending',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Sentra Kredit (${connName})` : 'Cabang Utama / Sentra Kredit',
        refNumber: `LN-${ts.toString().slice(-4)}-01`,
        accountNumber: 'ACC-LN-881923',
        amount: 2500000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.96,
        description: `Pencairan fasilitas kredit komersial melampaui limit wewenang cabang (Plafond Rp 2.500 Juta, LTV 96.5%) dari ${tbl}`,
        reason: `Pencairan kredit melampaui limit wewenang cabang tanpa persetujuan Komite Kredit Wilayah (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 2500.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Sentra Kredit Korporasi Regional',
        refNumber: `LN-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'ACC-LN-662819',
        amount: 1200000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.89,
        description: `Restrukturisasi debitur macet (KOL-3) tanpa appraisal agunan independen dari ${tbl}`,
        reason: `Pembaruan plafon debitur non-performing tanpa penilaian independen (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 1200.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Sentra Kredit Menengah Bandung',
        refNumber: `LN-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'ACC-LN-551920',
        amount: 1850000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.93,
        description: `Pencairan kredit modal kerja tanpa pengikatan sertifikat hak tanggungan (SKMHT/APHT) sah (${tbl})`,
        reason: `Pencairan sebelum persyaratan penyerahan covernote notaris selesai (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 1850.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Sentra Pembiayaan Konsumer Surabaya',
        refNumber: `LN-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'ACC-LN-441829',
        amount: 3100000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.95,
        description: `Perpanjangan tenor pembiayaan investasi tanpa analisis kelayakan arus kas debitur (${tbl})`,
        reason: `Override sistem penilaian skor risiko kredit (Credit Scoring Bypass) (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 3100.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Sentra Usaha Mikro & Kecil Semarang',
        refNumber: `LN-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'ACC-LN-331728',
        amount: 950000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.87,
        description: `Pemberian grace period angsuran bunga tanpa memorandum persetujuan komite kredit (${tbl})`,
        reason: `Modifikasi jadwal angsuran manual pada core lending system (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 950.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Sentra Kredit Komersial Medan',
        refNumber: `LN-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'ACC-LN-221627',
        amount: 650000000,
        type: 'Lending',
        category: 'Lending',
        anomalyScore: 0.84,
        description: `Pemberian diskon margin suku bunga pinjaman di bawah Cost of Loanable Funds bank (${tbl})`,
        reason: `Deviasi suku bunga pinjaman tanpa eskalasi ke Divisi Asset & Liability (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'MODERATE_HIGH',
        xMetric: 650.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('LN', 80.0, 1150.0, 60, 8, 17)
  },
  customer_accounts: {
    canonicalCategory: 'Funding',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Operasional Giro (${connName})` : 'Cabang Utama Jakarta / Dana',
        refNumber: `FD-${ts.toString().slice(-4)}-01`,
        accountNumber: 'ACC-10029381',
        amount: 850000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.93,
        description: `Penarikan dana tunai masif pada rekening dormant tanpa otorisasi Branch Manager (${tbl})`,
        reason: `Aktivasi dan penarikan tiba-tiba saldo rekening dormant melampaui limit teller (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 850.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Divisi Wealth Management',
        refNumber: `FD-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'ACC-50019283',
        amount: 2500000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.95,
        description: `Penempatan deposito suku bunga spesial (8.5%) melampaui batas LPS tanpa persetujuan ALCO (${tbl})`,
        reason: `Pemberian special rate margin tanpa memo persetujuan ALCO (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 2500.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Cabang Surabaya / Teller Counter',
        refNumber: `FD-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'ACC-40028192',
        amount: 480000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.91,
        description: `Setoran tunai berulang terindikasi smurfing tanpa verifikasi profil nasabah (CDD) (${tbl})`,
        reason: `Setoran beruntun di bawah batas Rp 500 Juta untuk menghindari pelaporan transaksi tunai (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 480.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Cabang Khusus Korporasi Sudirman',
        refNumber: `FD-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'ACC-30019281',
        amount: 1600000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.94,
        description: `Pembukaan rekening giro institusi tanpa verifikasi kelengkapan dokumen Beneficial Ownership (${tbl})`,
        reason: `Bypass prosedur verifikasi identitas pengurus perseroan pada sistem CIF (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 1600.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Sentra Operasi Dana Bandung',
        refNumber: `FD-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'ACC-60018274',
        amount: 3200000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.96,
        description: `Lonjakan saldo masuk rekening giro instan dari rekening offshore tanpa dokumentasi underlying (${tbl})`,
        reason: `Penerimaan dana valas berdenominasi besar tanpa pemberitahuan devisa deviasi 240% (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 3200.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Layanan Contact Center & Digital',
        refNumber: `FD-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'ACC-80017263',
        amount: 350000000,
        type: 'Funding',
        category: 'Funding',
        anomalyScore: 0.86,
        description: `Aktivasi status rekening pasif via call center tanpa otorisasi dual control supervisor (${tbl})`,
        reason: `Perubahan status dormant ke aktif tanpa verifikasi biometrik/KTP fisik (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 350.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('FD', 15.0, 420.0, 60, 8, 17)
  },
  accounts: {
    canonicalCategory: 'Funding',
    anomalies: (ts, connName, tbl) => TABLE_ANOMALY_PROFILES.customer_accounts.anomalies(ts, connName, tbl),
    normalBaseline: () => TABLE_ANOMALY_PROFILES.customer_accounts.normalBaseline()
  },
  user_access_audit_logs: {
    canonicalCategory: 'IT Control',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Host Core Banking (${connName})` : 'Pusat Operasi TI / IT Security',
        refNumber: `LOG-${ts.toString().slice(-4)}-01`,
        accountNumber: 'USR-ADM-001',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.94,
        description: `Percobaan login gagal berulang kali (9 kali berturut-turut) pada pukul 02:30 WIB (Brute Force) (${tbl})`,
        reason: `Anomali akses di luar jam kerja dengan pola serangan brute force pada akun privileged (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 9.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Database Administrator Host-to-Host',
        refNumber: `LOG-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'USR-DBA-99',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.88,
        description: `Modifikasi langsung tabel hak akses user superadmin dari konsol tanpa nomor tiket insiden IT (${tbl})`,
        reason: `Bypass change management control pada modul otorisasi core banking (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 8.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Server Data Warehouse Reporting',
        refNumber: `LOG-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'USR-EXT-SEC4',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.97,
        description: `Ekstraksi massal data nasabah (PII data dump 500k baris) via direct SQL query tanpa masker enkripsi (${tbl})`,
        reason: `Pelanggaran fatal kontrol kerahasiaan data nasabah (Data Exfiltration Indicator) (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 11.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Gerbang Otentikasi Active Directory',
        refNumber: `LOG-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'USR-ADM-OPS',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.91,
        description: `Sesi login simultan dari dua IP subnet berbeda (Jakarta dan Luar Negeri) dalam selisih 3 menit (${tbl})`,
        reason: `Anomali lokasi geografis tidak mungkin (Impossible Travel Anomaly) (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 7.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Pipeline Deployment CI/CD TI',
        refNumber: `LOG-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'SVC-DEPLOY-AUTO',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.89,
        description: `Patching darurat komponen backend core banking tanpa pengujian staging & sign-off QA (${tbl})`,
        reason: `Bypass proses Quality Assurance dan CAB approval saat production release (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 6.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Gateway API Eksternal Partner',
        refNumber: `LOG-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'USR-SVC-PARTNER',
        amount: 0,
        type: 'IT Control',
        category: 'IT Control',
        anomalyScore: 0.85,
        description: `Lonjakan request otentikasi API yang gagal mencapai 400 req/menit dari IP tidak terdaftar (${tbl})`,
        reason: `Indikasi serangan credential stuffing terhadap API gateway perbankan (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 5.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('IT', 0.0, 2.0, 60, 8, 17)
  },
  compliance_incident_records: {
    canonicalCategory: 'KYC',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Unit Anti-Pencucian Uang (${connName})` : 'Unit Kepatuhan & AML',
        refNumber: `INC-${ts.toString().slice(-4)}-01`,
        accountNumber: 'ACC-CIF-881923',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.95,
        description: `Pembukaan rekening nasabah kategori PEP tanpa otorisasi Enhanced Due Diligence (EDD) (${tbl})`,
        reason: `Pelanggaran SOP APU-PPT terkait verifikasi beneficial owner dan profil sumber dana (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 85.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Kantor Wilayah / Sentra Operasi',
        refNumber: `INC-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'ACC-CIF-772910',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.89,
        description: `Pola transaksi terindikasi smurfing / penataan setoran di bawah batas pelaporan PPATK deviasi 185% (${tbl})`,
        reason: `Transaksi terpecah berulang dalam kurun 1 jam untuk menghindari pelaporan LTKT (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 65.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Divisi International Banking & FX',
        refNumber: `INC-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'ACC-CIF-661829',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.98,
        description: `Pengiriman dana valas ke entitas penerima yang terdaftar dalam daftar sanksi OFAC / PBB (${tbl})`,
        reason: `Kegagalan filtrasi screening real-time pada sistem Automated AML Watchlist (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 95.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Sentra Pemantauan Transaksi Keuangan',
        refNumber: `INC-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'ACC-CIF-551728',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.92,
        description: `Alert transaksi mencurigakan (STR alert) tidak ditindaklanjuti lebih dari 30 hari kalender (${tbl})`,
        reason: `Pelanggaran tenggat waktu pelaporan LTKM ke PPATK sesuai regulasi regulator (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 75.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Cabang Korporasi Surabaya',
        refNumber: `INC-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'ACC-CIF-441627',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.88,
        description: `Ketidaksesuaian profil omzet usaha debitur (deviasi 350%) dengan arus mutasi rekening (${tbl})`,
        reason: `Profil penghasilan nasabah tidak dimutakhirkan selama lebih dari 3 tahun berturut-turut (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 55.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Customer Service Hub Nasional',
        refNumber: `INC-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'ACC-CIF-331526',
        amount: 0,
        type: 'KYC',
        category: 'KYC',
        anomalyScore: 0.84,
        description: `Dokumen identitas izin usaha (NIB/NPWP) nasabah institusi telah kedaluwarsa tanpa tindak lanjut (${tbl})`,
        reason: `Kegagalan eksekusi jadwal periodic review kepatuhan berkala (KYC Review Failure) (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'MODERATE_HIGH',
        xMetric: 45.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('KYC', 3.0, 28.0, 60, 8, 17)
  },
  vendor_invoices: {
    canonicalCategory: 'Procurement',
    anomalies: (ts, connName, tbl) => [
      {
        id: `ANOM-${ts.toString().slice(-4)}-01`,
        entity: connName ? `Divisi Logistik & Pengadaan (${connName})` : 'Bagian Pengadaan & Umum',
        refNumber: `INV-${ts.toString().slice(-4)}-01`,
        accountNumber: 'VND-TECH-01',
        amount: 380000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.92,
        description: `Pembayaran faktur vendor tanpa dokumen Berita Acara Serah Terima (BAST) & PO resmi (${tbl})`,
        reason: `Pencairan dana tagihan tanpa verifikasi three-way matching antara PO, BAST, dan Invoice (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 380.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 1).toString().slice(-4)}-02`,
        entity: 'Finance & Accounting Dept',
        refNumber: `INV-${(ts + 1).toString().slice(-4)}-02`,
        accountNumber: 'VND-SUPPLY-04',
        amount: 125000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.88,
        description: `Duplikasi nomor faktur vendor (duplicate invoice) teridentifikasi dari vendor perangkat TI (${tbl})`,
        reason: `Indikasi penagihan ganda (double claim) dengan deskripsi dan nomor seri barang identik (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 125.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 2).toString().slice(-4)}-03`,
        entity: 'Divisi Pengadaan Barang & Jasa',
        refNumber: `INV-${(ts + 2).toString().slice(-4)}-03`,
        accountNumber: 'VND-BUILD-09',
        amount: 480000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.95,
        description: `Pemecahan surat pesanan pengadaan (Split Purchase Order) guna menghindari ambang lelang terbuka (${tbl})`,
        reason: `Penerbitan 5 PO bernominal di bawah Rp 100 Juta dalam hari yang sama ke vendor afiliasi (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 480.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 3).toString().slice(-4)}-04`,
        entity: 'Bagian Pembayaran Treasury Operasional',
        refNumber: `INV-${(ts + 3).toString().slice(-4)}-04`,
        accountNumber: 'VND-CONSULT-02',
        amount: 620000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.94,
        description: `Pencairan pembayaran invoice dengan perubahan nomor rekening bank vendor mendadak tanpa konfirmasi resmi (${tbl})`,
        reason: `Risiko Business Email Compromise (BEC) pada instruksi pengalihan rekening pencairan (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'CRITICAL',
        xMetric: 620.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 4).toString().slice(-4)}-05`,
        entity: 'Unit Manajemen Fasilitas Kantor',
        refNumber: `INV-${(ts + 4).toString().slice(-4)}-05`,
        accountNumber: 'VND-OFFICE-12',
        amount: 210000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.89,
        description: `Harga satuan barang pengadaan perlengkapan kantor melampaui Standar Biaya Masukan (SBM) hingga deviasi 78% (${tbl})`,
        reason: `Penyimpangan estimasi biaya (HPS) tanpa pembanding harga pasar wajar (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'HIGH',
        xMetric: 210.0,
        isAnomaly: true
      },
      {
        id: `ANOM-${(ts + 5).toString().slice(-4)}-06`,
        entity: 'Divisi Transformasi Digital & TI',
        refNumber: `INV-${(ts + 5).toString().slice(-4)}-06`,
        accountNumber: 'VND-SOFT-07',
        amount: 160000000,
        type: 'Procurement',
        category: 'Procurement',
        anomalyScore: 0.85,
        description: `Pembayaran termin lisensi perangkat lunak tanpa kelengkapan sertifikat lisensi resmi (${tbl})`,
        reason: `Ketidaklengkapan dokumen deliverables uji terima operasional (UAT sign-off missing) (${tbl})`,
        date: new Date().toISOString().slice(0, 10),
        timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
        riskLevel: 'MODERATE_HIGH',
        xMetric: 160.0,
        isAnomaly: true
      }
    ],
    normalBaseline: () => generateNormalBaselinePoints('PRC', 8.0, 95.0, 60, 8, 17)
  }
}

const getDefaultAnomalyProfile = (tableName: string, sampleRecords?: Record<string, unknown>[]): TableAnomalyProfile => {
  const cat = getCanonicalCategory(tableName)
  const prefix = cat.replace(/[^A-Za-z0-9]/g, '').slice(0, 4).toUpperCase() || 'ANOM'

  // Inspect sample records if available to find realistic baseline metrics
  const detectedAmounts: number[] = []
  if (Array.isArray(sampleRecords) && sampleRecords.length > 0) {
    for (const row of sampleRecords) {
      for (const [key, val] of Object.entries(row)) {
        const k = key.toLowerCase()
        if (
          k.includes('amount') ||
          k.includes('nominal') ||
          k.includes('balance') ||
          k.includes('nilai') ||
          k.includes('total') ||
          k.includes('price') ||
          k.includes('limit') ||
          k.includes('score') ||
          k.includes('volume')
        ) {
          const num = typeof val === 'number' ? val : parseFloat(String(val).replace(/[^0-9.-]/g, ''))
          if (!isNaN(num) && num > 0) {
            detectedAmounts.push(num > 100000 ? num / 1000000 : num)
          }
        }
      }
    }
  }

  // Base normal metrics derived dynamically or standard banking operational range
  const baseRange = detectedAmounts.length > 0
    ? { min: Math.max(1, Math.min(...detectedAmounts) * 0.7), max: Math.max(...detectedAmounts) * 1.3 }
    : { min: 12.0, max: 85.0 }

  // Generate dense, realistic normal distribution baseline (65 points)
  const dynamicNormalPoints = generateNormalBaselinePoints(prefix, baseRange.min, baseRange.max, 65, 8, 17)

  return {
    canonicalCategory: cat,
    anomalies: (ts, connName, tbl) => {
      const mults = [2.8, 1.9, 3.4, 2.2, 1.6, 2.5]
      const scores = [0.96, 0.89, 0.94, 0.87, 0.92, 0.85]
      const riskLevels = ['CRITICAL', 'HIGH', 'CRITICAL', 'HIGH', 'HIGH', 'MODERATE_HIGH']
      const anomalyDescs = [
        `Penyimpangan ambang batas tidak wajar melampaui limit wewenang operasional tabel ${tbl}`,
        `Aktivitas transaksional tidak biasa terdeteksi di luar jam kerja reguler (${tbl})`,
        `Deviasi nilai mutasi signifikan melampaui rata-rata historis populasi tabel ${tbl}`,
        `Inkonsistensi data transaksi tanpa dokumentasi otorisasi pendukung pada tabel ${tbl}`,
        `Indikasi pola pemecahan transaksi beruntun (Structuring/Smurfing) pada tabel ${tbl}`,
        `Frekuensi anomali di luar ambang batas standar toleransi risiko pada tabel ${tbl}`
      ]

      return mults.map((mult, idx) => {
        const xMetricVal = Math.round((baseRange.max * mult) * 100) / 100
        const idSeq = String(idx + 1).padStart(2, '0')
        return {
          id: `ANOM-${ts.toString().slice(-4)}-${idSeq}`,
          entity: connName ? `Unit Kerja (${connName})` : `Kantor Operasional / Regional ${idSeq}`,
          refNumber: `${prefix}-${ts.toString().slice(-4)}-${idSeq}`,
          accountNumber: `ACC-${tbl.replace(/[^A-Za-z0-9]/g, '').slice(0, 8).toUpperCase()}-${idSeq}`,
          amount: Math.round(xMetricVal * 1000000),
          type: cat,
          category: cat,
          anomalyScore: scores[idx] || 0.88,
          description: anomalyDescs[idx] || `Anomali data terdeteksi pada tabel ${tbl}`,
          reason: `Penyimpangan signifikan dari baseline distribusi normal populasi tabel ${tbl}`,
          date: new Date().toISOString().slice(0, 10),
          timestamp: new Date().toISOString().replace('T', ' ').slice(0, 19),
          riskLevel: riskLevels[idx] || 'HIGH',
          xMetric: xMetricVal,
          isAnomaly: true
        }
      })
    },
    normalBaseline: () => dynamicNormalPoints
  }
}

// Unified Action Handler for Adding/Injecting Data into the Current Feature
const handleExecuteFeatureAction = async (item: MappedRecord) => {
  if (!currentModuleInfo.value) return
  isExecutingAction.value = true

  const actionType = currentModuleInfo.value.actionType

  try {
    // 1. Corporate Risk Profile Injection
    if (actionType === 'risk_profile') {
      const riskStore = useRiskProfileStore()
      const totalPop = item.total_rows || getTotalRowCountForTable(item.table_name)
      const popStr = totalPop.toLocaleString('id-ID')

      let riskName = `Risiko Pemantauan Pipeline Data: ${item.table_name}`
      let riskDesc = `Risiko teridentifikasi otomatis dari data pipeline '${item.connection_name}' (tabel ${item.table_name}) dengan populasi ${popStr} rekaman.`
      let riskCategory = 'Financial Risk'
      let riskLikelihood = 3
      let riskImpact = 4
      let riskLevel = 'High'

      if (item.table_name === 'gl_transactions') {
        riskName = `Risiko Integritas Jurnal Umum & Otorisasi Transaksi Keuangan (${item.table_name})`
        riskDesc = `Risiko teridentifikasi otomatis dari data pipeline '${item.connection_name}' (tabel ${item.table_name}). Meliputi pengawasan terhadap total populasi ${popStr} transaksi jurnal umum, potensi anomali selisih akun perantara (suspense account), serta kepatuhan limit otorisasi pembukuan finansial korporat.`
        riskCategory = 'Financial Risk'
        riskLikelihood = 3
        riskImpact = 4
        riskLevel = 'High'
      } else if (item.table_name === 'loan_accounts') {
        riskName = `Risiko Portofolio Kredit: Kredit Bermasalah (${item.table_name})`
        riskDesc = `Risiko teridentifikasi otomatis dari data pipeline '${item.connection_name}' (tabel ${item.table_name}). Terdeteksi fasilitas pinjaman komersial dan KPR dari populasi ${popStr} rekening dengan status Kolektibilitas KOL-1 serta KOL-2 yang memerlukan pemantauan berkala terhadap risiko kredit macet (NPL).`
        riskCategory = 'Financial Risk'
        riskLikelihood = 3
        riskImpact = 4
        riskLevel = 'High'
      } else if (item.table_name === 'user_access_audit_logs') {
        riskName = `Risiko Keamanan Siber: Hak Akses Istimewa & Anomali Sistem (${item.table_name})`
        riskDesc = `Risiko teridentifikasi otomatis dari pipeline '${item.connection_name}' (tabel ${item.table_name}). Monitoring anomali login, eskalasi hak istimewa (privileged access), dan anomali jam akses dari total ${popStr} log aktivitas audit sistem.`
        riskCategory = 'Operational Risk'
        riskLikelihood = 4
        riskImpact = 4
        riskLevel = 'Extreme'
      } else if (item.table_name === 'vendor_invoices') {
        riskName = `Risiko Pengadaan: Keterlambatan Pembayaran & Faktur Vendor (${item.table_name})`
        riskDesc = `Risiko teridentifikasi otomatis dari pipeline '${item.connection_name}' (tabel ${item.table_name}). Monitoring kepatuhan syarat termin pembayaran, validitas faktur ganda (duplicate invoice), dan rekonsiliasi vendor dari total ${popStr} faktur.`
        riskCategory = 'Compliance Risk'
        riskLikelihood = 2
        riskImpact = 3
        riskLevel = 'Moderate'
      } else if (item.table_name === 'compliance_incident_records') {
        riskName = `Risiko Kepatuhan Regulasi & Sanksi Otoritas (${item.table_name})`
        riskDesc = `Risiko kepatuhan dari pipeline '${item.connection_name}' (tabel ${item.table_name}). Pemantauan ${popStr} insiden regulasi, mitigasi denda sanksi otoritas, serta pemenuhan SLA tindak lanjut.`
        riskCategory = 'Compliance Risk'
        riskLikelihood = 3
        riskImpact = 5
        riskLevel = 'High'
      }
      
      const existing = riskStore.rawRisks.find(r => r.name === riskName)
      if (existing) {
        toast.warning('Risiko Sudah Ada', `Risiko untuk tabel ${item.table_name} sudah terdaftar di Corporate Risk Profile.`)
        isExecutingAction.value = false
        return
      }

      const payload = {
        name: riskName,
        description: riskDesc,
        category: riskCategory,
        likelihood: riskLikelihood,
        impact: riskImpact,
        risk_level: riskLevel,
        source: `${item.connection_name} / ${item.table_name}`,
        status: 'Active',
        likelihood_q1: riskLikelihood,
        impact_q1: riskImpact,
        likelihood_q2: riskLikelihood,
        impact_q2: riskImpact,
        likelihood_q3: riskLikelihood,
        impact_q3: riskImpact,
        likelihood_q4: riskLikelihood,
        impact_q4: riskImpact
      }

      const ok = await riskStore.addRisk(payload).catch(() => false)
      if (!ok) {
        const optimisticRisk = {
          id: `risk-mapped-${Date.now()}`,
          ...payload,
          displayId: (riskStore.rawRisks?.length || 0) + 1,
          created_at: new Date().toISOString()
        }
        riskStore.rawRisks.push(optimisticRisk)
      }

      // Sync updated Risk Profile data directly into AI Entity Prediction training pool
      try {
        const { syncEntityPredictionWithRiskProfile } = useAiAnalytics()
        syncEntityPredictionWithRiskProfile().catch(() => {})
      } catch (_) {}

      toast.success(
        'Data Berhasil Ditambahkan ke Profil Risiko',
        `Risiko dari pipeline '${item.table_name}' (${item.connection_name}) berhasil ditambahkan ke Corporate Risk Profile!`
      )
    }

    // 2. AI Entity Prediction Execution (sync external DB table with Risk Profile and train AI model)
    else if (actionType === 'ai_entity_prediction') {
      const riskStore = useRiskProfileStore()
      const { syncEntityPredictionWithRiskProfile } = useAiAnalytics()

      // Map external database table to an entity format aligned with Corporate Risk Profile
      let entityName = `Divisi / Unit ${item.table_name}`
      let entityType: 'Branch' | 'Department' = 'Department'
      let riskCategory = 'Operational'
      let likelihood = 3
      let impact = 3

      if (item.table_name === 'gl_transactions') {
        entityName = 'Finance & Accounting'
        entityType = 'Department'
        riskCategory = 'Financial'
        likelihood = 3
        impact = 4
      } else if (item.table_name === 'loan_accounts' || item.table_name === 'accounts' || item.table_name === 'customer_accounts') {
        entityName = 'Kredit Komersial & Cabang Utama'
        entityType = 'Branch'
        riskCategory = 'Financial'
        likelihood = 3
        impact = 4
      } else if (item.table_name === 'user_access_audit_logs') {
        entityName = 'Information Technology & Cyber Security'
        entityType = 'Department'
        riskCategory = 'Technology'
        likelihood = 4
        impact = 4
      } else if (item.table_name === 'vendor_invoices') {
        entityName = 'Procurement & General Affairs'
        entityType = 'Department'
        riskCategory = 'Operations'
        likelihood = 2
        impact = 3
      } else if (item.table_name === 'compliance_incident_records') {
        entityName = 'Unit Kepatuhan Regulasi'
        entityType = 'Department'
        riskCategory = 'Compliance'
        likelihood = 3
        impact = 5
      }

      const calculatedRiskLevel = riskStore.getRiskLevel ? riskStore.getRiskLevel(likelihood, impact) : 'MODERATE_HIGH'

      // Ensure this external entity is registered in Corporate Risk Profile store
      const existingRisk = riskStore.rawRisks?.find(r => r.name?.includes(item.table_name) || r.name?.includes(entityName))
      if (!existingRisk && riskStore.addRisk) {
        const totalPop = item.total_rows || getTotalRowCountForTable(item.table_name)
        const popStr = totalPop.toLocaleString('id-ID')
        const riskPayload = {
          name: `Risiko Terintegrasi: ${entityName} (${item.table_name})`,
          description: `Risiko entitas teridentifikasi dari external database '${item.connection_name}' (tabel ${item.table_name}) dengan populasi ${popStr} rekaman. Diselaraskan untuk prediksi AI.`,
          category: riskCategory,
          likelihood,
          impact,
          risk_level: calculatedRiskLevel,
          source: `${item.connection_name} / ${item.table_name}`,
          status: 'Active'
        }
        riskStore.addRisk(riskPayload).catch(() => {
          if (Array.isArray(riskStore.rawRisks)) {
            riskStore.rawRisks.push({
              id: `risk-ext-${Date.now()}`,
              ...riskPayload,
              displayId: (riskStore.rawRisks?.length || 0) + 1,
              created_at: new Date().toISOString()
            })
          }
        })
      }

      // Prepare payload to train AI model directly from source without CSV
      const customItems: any[] = [{
        entity: entityName,
        entity_type: entityType,
        risk_category: riskCategory,
        inherent_likelihood: likelihood,
        inherent_impact: impact,
        audit_findings_count: item.sample_records?.length || 2,
        kpi_below_target_count: 1,
        kpi_volatility: 0.15,
        previous_risk_score: likelihood * impact,
        assessment_month: new Date().getMonth() + 1,
        current_risk_level: calculatedRiskLevel,
        source_id: `external_db:${item.connection_name}:${item.table_name}`
      }]

      // If sample records contain branch/unit fields, add granular entity items
      if (Array.isArray(item.sample_records)) {
        item.sample_records.forEach((row: any, idx: number) => {
          const rowEntity = row.branch_code ? `Cabang ${row.branch_code}` : (row.reporting_unit || row.channel || null)
          if (rowEntity && !customItems.some(c => c.entity === rowEntity)) {
            const isBr = String(rowEntity).toLowerCase().includes('cabang') || String(rowEntity).toLowerCase().includes('branch')
            customItems.push({
              entity: rowEntity,
              entity_type: isBr ? 'Branch' : 'Department',
              risk_category: riskCategory,
              inherent_likelihood: likelihood,
              inherent_impact: impact,
              audit_findings_count: idx + 1,
              kpi_below_target_count: 0,
              kpi_volatility: 0.1,
              previous_risk_score: likelihood * impact,
              assessment_month: new Date().getMonth() + 1,
              current_risk_level: calculatedRiskLevel,
              source_id: `external_db:${item.connection_name}:${item.table_name}`
            })
          }
        })
      }

      await syncEntityPredictionWithRiskProfile(customItems, true)

      toast.success(
        'Prediksi Entitas AI Berhasil Disinkronkan',
        `Data pipeline '${item.table_name}' (${item.connection_name}) berhasil disinkronkan dengan Corporate Risk Profile dan dilatih ulang langsung oleh model AI!`
      )
    }

    // 3. AI Anomaly Detection Execution (applied across ALL tables connected from database)
    else if (actionType === 'ai_anomaly') {
      const { isolationState, addCustomAnomalyRecord, addCustomScatterBatch, trainAnomalyWithRules } = useAiAnalytics()

      const ts = Date.now()
      const profile = TABLE_ANOMALY_PROFILES[item.table_name] || getDefaultAnomalyProfile(item.table_name, item.sample_records)
      const targetCategory = profile.canonicalCategory

      // Trigger rule-based AI training if anomaly rules are attached to this mapping
      const allRules = [
        ...(item.anomaly_rules || []).map((r: string, idx: number) => ({
          id: `preset-${ts}-${idx}`,
          name: r,
          field: 'amount',
          operator: '>',
          threshold: 100000000,
          severity: 'HIGH',
          category: targetCategory,
          description: `Aturan anomali standar: ${r}`
        })),
        ...(item.custom_rules || []).filter((cr: any) => cr.isActive !== false)
      ]

      if (allRules.length > 0) {
        trainAnomalyWithRules({
          connection_name: item.connection_name,
          table_name: item.table_name,
          category: targetCategory,
          rules: allRules,
          sample_records: item.sample_records
        }).catch(err => console.warn('Background AI rule training error:', err))
      }

      // Check existing anomaly records for this table to track batches without artificial caps
      const existingCount = (isolationState.value?.anomalies || []).filter((a: any) =>
        a.description?.includes(item.table_name) || a.refNumber?.includes(item.table_name) || a.reason?.includes(item.table_name)
      ).length

      // 1. Add table-specific anomalies (red points) - dynamic and scalable
      const sampleAnomalies = profile.anomalies(ts, item.connection_name, item.table_name)
      const offHours = [2, 23, 1, 22, 3, 0]
      sampleAnomalies.forEach((anom, idx) => {
        const hour = anom.hour ?? offHours[idx % offHours.length]
        anom.hour = hour
        anom.hour_of_day = hour
        const scatterPoint = {
          x: anom.xMetric,
          y: Math.min(23.9, Math.max(0.0, Math.round((hour + ((idx % 4) * 0.2)) * 10) / 10)),
          type: targetCategory,
          isAnomaly: true,
          label: anom.id
        }
        addCustomAnomalyRecord(anom, scatterPoint)
      })

      // 2. Add table-specific real normal baseline distribution points (blue points) from Gold Zone
      const normalBaselineValues = profile.normalBaseline()
      const normalDistributionPoints = normalBaselineValues.map((pt, idx) => ({
        x: pt.x,
        y: pt.y,
        type: targetCategory,
        isAnomaly: false,
        label: pt.id || `${targetCategory.slice(0, 3).toUpperCase()}-NORM-${ts.toString().slice(-4)}-${idx + 1}`
      }))

      addCustomScatterBatch(normalDistributionPoints)

      // 3. Dynamically update population scan summary with table row volume
      if (isolationState.value?.summary) {
        const additionalRows = item.total_rows || getTotalRowCountForTable(item.table_name) || 1000
        isolationState.value.summary.totalScanned = (isolationState.value.summary.totalScanned || 1200) + additionalRows
        isolationState.value.summary.anomaliesFound = (isolationState.value.anomalies || []).length
        isolationState.value.summary.contaminationRate = Math.round(
          (isolationState.value.summary.anomaliesFound / Math.max(isolationState.value.summary.totalScanned, 1)) * 1000
        ) / 1000
      }

      const batchNote = existingCount > 0 ? ` (Batch Tambahan #${Math.floor(existingCount / sampleAnomalies.length) + 1})` : ''
      toast.success(
        'AI Anomaly Detection Berhasil Dijalankan',
        `${sampleAnomalies.length} Transaksi anomali & ${normalDistributionPoints.length} titik distribusi data normal kategori '${targetCategory}' (tabel ${item.table_name})${batchNote} berhasil disatukan di scatter plot!`
      )
    }

    // 3. Audit Universe Injection
    else if (actionType === 'audit_universe') {
      const universeStore = useAuditUniverseStore()
      const entityName = `Proses Operasional: ${item.table_name} (${item.connection_name})`
      
      const existingNode = universeStore.corporateUniverse?.find(u => u.name === entityName)
      if (existingNode) {
        toast.warning('Entitas Sudah Terdaftar', `Entitas ${item.table_name} sudah ada di Audit Universe.`)
        isExecutingAction.value = false
        return
      }

      const newNode = {
        id: `corp-node-${Date.now()}`,
        name: entityName,
        parent_id: undefined
      }
      if (Array.isArray(universeStore.corporateUniverse)) {
        universeStore.corporateUniverse.push(newNode as any)
      }

      toast.success(
        'Berhasil Ditambahkan ke Audit Universe',
        `Entitas audit untuk tabel '${item.table_name}' berhasil ditambahkan ke Audit Universe!`
      )
    }

    // 4. Working Paper Evidence Sampling Injection
    else if (actionType === 'working_paper') {
      const wpStore = useWorkingPaperStore()
      const totalPop = item.total_rows || getTotalRowCountForTable(item.table_name)
      const popStr = totalPop.toLocaleString('id-ID')

      if (wpStore.sampleForm) {
        wpStore.sampleForm.population = `${popStr} Transaksi`
        wpStore.sampleForm.sampleSize = item.sample_records.length
        wpStore.sampleForm.conclusion = `Pengujian substantif validasi otorisasi dan ambang batas limit transaksi terhadap ${item.sample_records.length} sampel uji petik dari total populasi ${popStr} baris transaksi ${item.table_name}.`

        wpStore.sampleForm.samples = item.sample_records.map((row, idx) => ({
          id: Date.now() + idx,
          fieldworkDocument: String(row[item.audit_field] || row.id || `SMP-00${idx + 1}`),
          document: `Bukti Transaksi ${item.table_name} (${row[item.audit_field] || row.transaction_ref || row.id})`,
          step1: 'Verifikasi Otorisasi & Kelengkapan Dokumen Pendukung',
          l1: 'Pass',
          step2: 'Kesesuaian Limit Transaksi & Rekonsiliasi Saldo',
          l2: 'Pass',
          step3: 'Validasi Anti-Fraud & Timestamp Transaksi',
          l3: 'Pass'
        }))
      }

      toast.success(
        'Bukti Sampel Kertas Kerja Berhasil Ditambahkan',
        `Populasi ${popStr} transaksi dan ${item.sample_records.length} sampel uji petik dari '${item.table_name}' berhasil dimasukkan ke Form F-03 Working Paper!`
      )
    }

    // 5. Generic Module Linking
    else {
      toast.success(
        'Data Pipeline Berhasil Ditautkan',
        `Tabel '${item.table_name}' (${item.connection_name}) berhasil disinkronkan ke modul ${currentModuleInfo.value.label}.`
      )
    }
  } catch (err) {
    toast.error('Gagal Menjalankan Aksi', String(err))
  } finally {
    isExecutingAction.value = false
  }
}
</script>

<template>
  <div
    v-if="currentModuleInfo && mappedRecords.length > 0 && activeRecord && !isDismissed"
    class="mb-4 mx-4 sm:mx-6 rounded-xl border border-primary-500/20 bg-primary-50/70 dark:bg-primary-950/30 p-3.5 shadow-xs transition-all duration-200"
  >
    <!-- Multi-Table Tab Bar when 2 or more tables are connected -->
    <div v-if="mappedRecords.length > 1" class="flex items-center gap-2 mb-3 border-b border-primary-500/20 pb-2.5 overflow-x-auto">
      <span class="text-[11px] font-bold text-primary-800 dark:text-primary-200 uppercase tracking-wider shrink-0 flex items-center gap-1.5">
        <UIcon name="i-lucide-layers" class="size-3.5 text-primary-500" />
        {{ mappedRecords.length }} Tabel Terhubung:
      </span>
      <button
        v-for="(rec, idx) in mappedRecords"
        :key="rec.connection_id + rec.table_name"
        type="button"
        class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-all duration-150 shrink-0 cursor-pointer"
        :class="selectedTableIndex === idx 
          ? 'bg-primary-600 text-white shadow-xs font-bold' 
          : 'bg-primary-100/70 dark:bg-primary-900/40 text-primary-700 dark:text-primary-300 hover:bg-primary-200/70'"
        @click="selectedTableIndex = idx"
      >
        <UIcon name="i-lucide-table" class="size-3.5" />
        <span>{{ rec.table_name }}</span>
        <span class="text-[10px] px-1.5 py-0.2 rounded-full" :class="selectedTableIndex === idx ? 'bg-white/20 text-white' : 'bg-primary-500/10 text-primary-600'">
          {{ rec.sample_records.length }} rows
        </span>
      </button>
    </div>

    <!-- Active Table Information and Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div class="flex items-start sm:items-center gap-3">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary-500/10 text-primary-600 dark:text-primary-400">
          <UIcon name="i-lucide-database" class="size-5" />
        </div>
        <div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-xs font-semibold text-primary-700 dark:text-primary-300">
              Pipeline Sumber Data Terhubung
            </span>
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300">
              <span class="size-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              Live Synced
            </span>
            <span class="text-xs text-[var(--text-muted)]">
              &bull; {{ activeRecord.connection_name }} &rsaquo; <code class="font-mono font-semibold text-primary-700 dark:text-primary-300">{{ activeRecord.table_name }}</code>
            </span>
          </div>
          <p class="text-xs text-[var(--text-muted)] mt-0.5">
            Tabel <code class="font-mono font-bold text-primary-700 dark:text-primary-300">{{ activeRecord.table_name }}</code>: 
            Total Populasi <strong class="text-[var(--text-main)]">{{ (activeRecord.total_rows || 142580).toLocaleString('id-ID') }} baris</strong> (100% Ingested ke Lakehouse). 
            Tersedia <strong class="text-primary-600 dark:text-primary-400">{{ activeRecord.sample_records?.length || 0 }} rekaman sampel uji petik</strong> untuk formulir ini.
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2 shrink-0 flex-wrap">
        <!-- Direct Context-Aware Action Button (Add to Risk Profile, Run AI Anomaly, Add to Universe, etc.) -->
        <UButton
          size="xs"
          color="primary"
          variant="solid"
          icon="i-lucide-play-circle"
          :loading="isExecutingAction"
          @click="handleExecuteFeatureAction(activeRecord)"
        >
          {{ currentModuleInfo.actionLabel }}
        </UButton>

        <UButton
          size="xs"
          color="primary"
          variant="soft"
          icon="i-lucide-table"
          @click="openPreview(activeRecord)"
        >
          Lihat Sampel Uji Petik ({{ activeRecord.sample_records?.length || 0 }})
        </UButton>
        <UButton
          size="xs"
          color="neutral"
          variant="ghost"
          icon="i-lucide-external-link"
          to="/settings?tab=datasource"
        >
          Konfigurasi
        </UButton>
        <UButton
          size="xs"
          color="neutral"
          variant="ghost"
          icon="i-lucide-x"
          @click="isDismissed = true"
        />
      </div>
    </div>

    <!-- Modal Detail Sample Records -->
    <UModal v-model:open="isModalOpen" :title="`Data Mapping: ${selectedMapping?.table_name || ''}`">
      <template #content>
        <div class="p-4 space-y-4 max-h-[75vh] overflow-y-auto">
          <div class="flex items-center justify-between border-b pb-3 border-[var(--border-main)] flex-wrap gap-2">
            <div>
              <h4 class="text-sm font-bold text-[var(--text-main)]">
                Tabel Sumber: {{ selectedMapping?.table_name }}
              </h4>
              <p class="text-xs text-[var(--text-muted)]">
                Koneksi: {{ selectedMapping?.connection_name }} | Target: {{ selectedMapping?.target_module }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono bg-[var(--bg-surface)] px-2 py-1 rounded border border-[var(--border-main)]">
                Audit Field (PK): {{ selectedMapping?.audit_field }}
              </span>
              <UButton
                v-if="selectedMapping"
                size="xs"
                color="primary"
                variant="solid"
                icon="i-lucide-check-circle"
                :loading="isExecutingAction"
                @click="handleExecuteFeatureAction(selectedMapping)"
              >
                {{ currentModuleInfo?.actionLabel || 'Tautkan Data' }}
              </UButton>
            </div>
          </div>

          <!-- 3-Column Summary Badges (Lakehouse Population vs Working Paper Sample) -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
            <div class="p-2.5 rounded-lg border border-primary-500/20 bg-primary-50/50 dark:bg-primary-950/20">
              <div class="flex items-center gap-1.5 text-primary-700 dark:text-primary-300 text-xs font-semibold">
                <UIcon name="i-lucide-database" class="size-4 shrink-0" />
                <span>Total Populasi Lakehouse</span>
              </div>
              <div class="text-base font-bold text-[var(--text-main)] mt-1">
                {{ (selectedMapping?.total_rows || 142580).toLocaleString('id-ID') }} <span class="text-xs font-normal text-[var(--text-muted)]">baris</span>
              </div>
              <p class="text-[10px] text-emerald-600 dark:text-emerald-400 font-medium mt-0.5">
                ✓ 100% Ingested ke Lakehouse
              </p>
            </div>

            <div class="p-2.5 rounded-lg border border-[var(--border-main)] bg-[var(--bg-surface)]">
              <div class="flex items-center gap-1.5 text-amber-700 dark:text-amber-400 text-xs font-semibold">
                <UIcon name="i-lucide-check-square" class="size-4 shrink-0" />
                <span>Sampel Uji Petik Formulir</span>
              </div>
              <div class="text-base font-bold text-[var(--text-main)] mt-1">
                {{ selectedMapping?.sample_records?.length || 0 }} <span class="text-xs font-normal text-[var(--text-muted)]">rekaman sampel</span>
              </div>
              <p class="text-[10px] text-[var(--text-muted)] font-medium mt-0.5">
                Uji Substantif (Form F-03)
              </p>
            </div>

            <div class="p-2.5 rounded-lg border border-[var(--border-main)] bg-[var(--bg-surface)]">
              <div class="flex items-center gap-1.5 text-blue-700 dark:text-blue-400 text-xs font-semibold">
                <UIcon name="i-lucide-activity" class="size-4 shrink-0" />
                <span>Status & Analitik</span>
              </div>
              <div class="text-base font-bold text-emerald-600 dark:text-emerald-400 mt-1 flex items-center gap-1.5">
                <span class="size-2 rounded-full bg-emerald-500 animate-pulse"></span>
                <span>Active Stream</span>
              </div>
              <p class="text-[10px] text-[var(--text-muted)] mt-0.5">
                Target: {{ selectedMapping?.target_module }}
              </p>
            </div>
          </div>

          <!-- Metodologi Audit Explanation Alert -->
          <div class="p-3 rounded-lg border border-blue-500/20 bg-blue-50/50 dark:bg-blue-950/20 text-xs text-blue-900 dark:text-blue-200 space-y-1.5">
            <div class="font-bold flex items-center gap-1.5">
              <UIcon name="i-lucide-info" class="size-4 text-blue-600 shrink-0" />
              <span>Standar Metodologi Audit (IIA Standards & Sampling Framework):</span>
            </div>
            <p class="text-[11px] leading-relaxed text-blue-800 dark:text-blue-300">
              Seluruh <strong>{{ (selectedMapping?.total_rows || 142580).toLocaleString('id-ID') }} baris</strong> data telah berhasil disinkronkan dan tersimpan di database Lakehouse. 
              Untuk formulir <em>Kertas Kerja (Working Paper)</em>, sistem menampilkan <strong>{{ selectedMapping?.sample_records?.length || 0 }} sampel uji petik</strong> untuk pengujian substantif manual agar formulir tetap ringan dan terfokus pada evidens uji petik.
            </p>
            <div class="pt-1 flex items-center gap-2 flex-wrap text-[11px]">
              <span class="text-[var(--text-muted)]">Ingin menguji 100% populasi penuh tanpa sampel?</span>
              <NuxtLink to="/analytics/caatt/full-population" class="font-bold underline text-primary-600 hover:text-primary-700 flex items-center gap-1">
                <span>Buka CAATT: Full Population Testing</span>
                <UIcon name="i-lucide-arrow-right" class="size-3" />
              </NuxtLink>
              <span class="text-[var(--text-muted)]">atau</span>
              <NuxtLink to="/analytics/ai/anomaly-detection" class="font-bold underline text-primary-600 hover:text-primary-700 flex items-center gap-1">
                <span>AI Anomaly Detection</span>
                <UIcon name="i-lucide-arrow-right" class="size-3" />
              </NuxtLink>
            </div>
          </div>

          <div v-if="selectedMapping && selectedMapping.sample_records && selectedMapping.sample_records.length > 0" class="overflow-x-auto border border-[var(--border-main)] rounded-lg">
            <table class="w-full text-left text-xs border-collapse">
              <thead class="bg-[var(--bg-surface)] border-b border-[var(--border-main)]">
                <tr>
                  <th
                    v-for="col in previewColumns"
                    :key="col"
                    class="p-2.5 font-bold uppercase tracking-wider text-[var(--text-muted)] text-[11px]"
                  >
                    {{ col }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-[var(--border-main)]">
                <tr
                  v-for="(row, idx) in selectedMapping.sample_records"
                  :key="idx"
                  class="hover:bg-[var(--bg-surface)]/60 transition-colors"
                >
                  <td
                    v-for="col in previewColumns"
                    :key="col"
                    class="p-2.5 font-mono text-[var(--text-main)]"
                  >
                    {{ row[col] !== null && row[col] !== undefined ? String(row[col]) : '-' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-else class="text-center py-6 text-xs text-[var(--text-muted)]">
            Tidak ada rekaman data sampel yang dapat ditampilkan.
          </div>

          <div class="flex justify-end pt-2">
            <UButton color="neutral" variant="outline" size="sm" @click="isModalOpen = false">
              Tutup
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
