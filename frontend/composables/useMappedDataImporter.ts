import { ref } from 'vue'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useRiskAppetiteStore } from '~/stores/risk-appetite'
import { useRiskFactorsStore } from '~/stores/risk-factors'
import { useRCMStore } from '~/stores/rcm'
import { useMitigationStore } from '~/stores/mitigation-risk'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditExecutionStore } from '~/stores/audit-execution'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { usePerformanceStore } from '~/stores/performance'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { useConsultingServiceStore } from '~/stores/consulting-service'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { getAuditServiceBaseUrl, getMasterServiceBaseUrl, getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { normalizeAnomalyRiskLevel } from '~/utils/riskLevelBadge'
import {
  buildAssignmentLetterCandidatesFromTable,
  buildAuditFieldworkCandidatesFromTable,
  buildAuditPriorityCandidatesFromTable,
  buildAuditResultReportCandidateFromTable,
  buildAuditUniverseCandidatesFromTable,
  buildConsultingServiceCandidatesFromTable,
  buildEntityPredictionItemsFromTable,
  buildInternalAuditKpiCandidatesFromTable,
  buildKpiForecastItemsFromTable,
  buildNlpDocumentsFromTable,
  buildQualityAssuranceCandidatesFromTable,
  buildRcmCandidatesFromTable,
  buildRealTableAnomalies,
  buildRiskAppetiteStatementsFromTable,
  buildRiskFactorsFromTable,
  buildRiskMitigationCandidatesFromTable,
  buildRiskProfilePayloadsFromRows,
  type MappedRecordInput,
} from '~/utils/mappedFeatureAdapters'
import {
  evaluateModuleCompatibility,
  targetModuleToRoute,
  type TableCompatibilityInput,
} from '~/utils/moduleCompatibility'

export type MappingImportStatusKind =
  | 'imported'
  | 'skipped_duplicate'
  | 'manual_required'
  | 'incompatible'
  | 'error'

export interface MappingImportResultItem {
  key: string
  connectionId: string
  connectionName: string
  tableName: string
  targetScope: string
  targetModule: string
  status: MappingImportStatusKind
  addedCount: number
  skippedCount: number
  message: string
  route: string
  importedAt: string
}

export interface BulkImportSummary {
  connectionId: string
  connectionName: string
  totalMappings: number
  importedMappings: number
  skippedMappings: number
  manualMappings: number
  incompatibleMappings: number
  errorMappings: number
  totalRecordsAdded: number
  results: MappingImportResultItem[]
  finishedAt: string
}

export interface MappingToImport {
  tableName: string
  targetScope: string
  targetModule: string
  auditField?: string
  isActive?: boolean
  anomalyRules?: string[]
  customRules?: any[]
}

const STORAGE_KEY = 'auditsphere_mapping_import_status_v1'

const importStatusMap = ref<Record<string, MappingImportResultItem>>({})
let initializedFromStorage = false

function loadPersistedStatuses() {
  if (initializedFromStorage || typeof window === 'undefined') return
  initializedFromStorage = true
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (parsed && typeof parsed === 'object') {
        importStatusMap.value = parsed
      }
    }
  } catch {
    // Ignore storage read errors
  }
}

function persistStatuses() {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(importStatusMap.value))
  } catch {
    // Ignore storage write errors
  }
}

export function buildMappingStatusKey(connectionId: string, tableName: string, targetModule: string): string {
  return `${connectionId}::${(tableName || '').trim().toLowerCase()}::${(targetModule || '').trim().toLowerCase()}`
}

function collectCorporateNodeNames(nodes: any[], out: Set<string>) {
  for (const n of nodes || []) {
    if (n?.name) out.add(String(n.name).trim().toLowerCase())
    if (Array.isArray(n?.children) && n.children.length > 0) {
      collectCorporateNodeNames(n.children, out)
    }
  }
}

export function useMappedDataImporter() {
  loadPersistedStatuses()

  const isImporting = ref(false)
  const lastSummary = ref<BulkImportSummary | null>(null)

  const getMappingStatus = (connectionId: string, tableName: string, targetModule: string) => {
    const key = buildMappingStatusKey(connectionId, tableName, targetModule)
    return importStatusMap.value[key] || null
  }

  const recordStatus = (item: MappingImportResultItem) => {
    importStatusMap.value = {
      ...importStatusMap.value,
      [item.key]: item,
    }
    persistStatuses()
  }

  const importAllForConnection = async (params: {
    connection: { id: string; name: string }
    mappings: MappingToImport[]
    schemaTables?: TableCompatibilityInput[]
  }): Promise<BulkImportSummary> => {
    const { connection, mappings, schemaTables = [] } = params
    isImporting.value = true

    const nowIso = new Date().toISOString()
    const results: MappingImportResultItem[] = []
    const schemaByTable = new Map<string, TableCompatibilityInput>()
    for (const t of schemaTables) {
      schemaByTable.set((t.tableName || '').trim().toLowerCase(), t)
    }

    // Deduplicate identical active mappings within the batch
    const seenPairs = new Set<string>()
    const activeMappings: MappingToImport[] = []
    for (const m of mappings || []) {
      if (m.isActive === false) continue
      if (!m.tableName || !m.targetModule) continue
      const pairKey = buildMappingStatusKey(connection.id, m.tableName, m.targetModule)
      if (seenPairs.has(pairKey)) continue
      seenPairs.add(pairKey)
      activeMappings.push(m)
    }

    const masterBaseUrl = getMasterServiceBaseUrl()
    const riskBaseUrl = getRiskServiceBaseUrl()
    const auditBaseUrl = getAuditServiceBaseUrl()

    // Fetch mapped-records per unique (targetModule, targetScope) for auto-importable mappings
    const mappedRecordsCache = new Map<string, MappedRecordInput[]>()

    const fetchRecordsForModule = async (targetModule: string, targetScope: string): Promise<MappedRecordInput[]> => {
      const cacheKey = `${targetScope}::${targetModule}`
      if (mappedRecordsCache.has(cacheKey)) {
        return mappedRecordsCache.get(cacheKey)!
      }
      try {
        const url = `${masterBaseUrl}/data-sources/mapped-records?target_module=${encodeURIComponent(targetModule)}&target_scope=${encodeURIComponent(targetScope)}&source_id=${encodeURIComponent(connection.id)}`
        const res = await $fetch<{ success?: boolean; data?: MappedRecordInput[] }>(url)
        const list = Array.isArray(res?.data) ? res.data : []
        mappedRecordsCache.set(cacheKey, list)
        return list
      } catch {
        mappedRecordsCache.set(cacheKey, [])
        return []
      }
    }

    // Shared stores/state loaded lazily only if needed
    let riskStoreLoaded = false
    let riskStoreTouched = false
    const existingRiskNames = new Set<string>()

    let universeStoreLoaded = false
    let universeStoreTouched = false
    const existingUniverseNames = new Set<string>()

    const riskStore = useRiskProfileStore()
    const universeStore = useAuditUniverseStore()

    try {
      for (const mapping of activeMappings) {
        const statusKey = buildMappingStatusKey(connection.id, mapping.tableName, mapping.targetModule)
        const route = targetModuleToRoute(mapping.targetModule)
        const schemaMeta = schemaByTable.get(mapping.tableName.trim().toLowerCase())

        // 1. Pre-check schema compatibility if schema metadata is known
        const schemaEval = evaluateModuleCompatibility(
          schemaMeta || { tableName: mapping.tableName },
          mapping.targetModule
        )

        if (!schemaEval.supported) {
          const item: MappingImportResultItem = {
            key: statusKey,
            connectionId: connection.id,
            connectionName: connection.name,
            tableName: mapping.tableName,
            targetScope: mapping.targetScope,
            targetModule: mapping.targetModule,
            status: 'incompatible',
            addedCount: 0,
            skippedCount: 0,
            message: schemaEval.reason,
            route,
            importedAt: nowIso,
          }
          results.push(item)
          recordStatus(item)
          continue
        }

        if (schemaMeta && !schemaEval.compatible) {
          const item: MappingImportResultItem = {
            key: statusKey,
            connectionId: connection.id,
            connectionName: connection.name,
            tableName: mapping.tableName,
            targetScope: mapping.targetScope,
            targetModule: mapping.targetModule,
            status: 'incompatible',
            addedCount: 0,
            skippedCount: 0,
            message: schemaEval.reason,
            route,
            importedAt: nowIso,
          }
          results.push(item)
          recordStatus(item)
          continue
        }

        // 2. Manual confirmation / pre-fill modules (Annual Plan, Strategic Plan, Working Paper, Audit Activity Plan)
        if (schemaEval.importMode === 'manual_banner' || schemaEval.importMode === 'prefill_form') {
          const manualMsg =
            schemaEval.actionType === 'annual_plan'
              ? 'Tetap melalui banner dengan konfirmasi manual (Departemen & Audit Universe) di halaman Annual Audit Plan.'
              : schemaEval.actionType === 'strategic_plan'
                ? 'Tetap melalui banner dengan pengisian Target KPI manual di halaman Strategic Audit Plan.'
                : schemaEval.actionType === 'audit_activity_plan'
                  ? 'Buka halaman Audit Activity Plan lalu klik tombol banner untuk mengisi Formulir Rencana Kegiatan Audit pada PKAT aktif.'
                  : 'Buka halaman Working Paper (F-03) lalu klik tombol banner untuk mengisi sampel uji petik pada Surat Tugas aktif.'
          const item: MappingImportResultItem = {
            key: statusKey,
            connectionId: connection.id,
            connectionName: connection.name,
            tableName: mapping.tableName,
            targetScope: mapping.targetScope,
            targetModule: mapping.targetModule,
            status: 'manual_required',
            addedCount: 0,
            skippedCount: 0,
            message: manualMsg,
            route,
            importedAt: nowIso,
          }
          results.push(item)
          recordStatus(item)
          continue
        }

        // 3. Auto-importable modules — fetch real mapped record payload from Data Hub via master-service
        try {
          const records = await fetchRecordsForModule(mapping.targetModule, mapping.targetScope)
          const record = records.find(
            r => (r.table_name || '').trim().toLowerCase() === mapping.tableName.trim().toLowerCase()
          )

          if (!record || (!record.sample_records?.length && !record.total_rows && !record.profile?.total_rows)) {
            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'error',
              addedCount: 0,
              skippedCount: 0,
              message: `Data nyata tabel '${mapping.tableName}' belum tersedia dari Data Hub. Jalankan Sync Now terlebih dahulu.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
            continue
          }

          // Re-check compatibility using the full runtime profile + sample rows
          const runtimeEval = evaluateModuleCompatibility(
            {
              tableName: record.table_name,
              columns: schemaMeta?.columns,
              rowCount: record.total_rows || record.profile?.total_rows || record.sample_records?.length || 0,
              profile: record.profile,
              sample_records: record.sample_records,
              data_hub_unavailable: record.data_hub_unavailable,
            },
            mapping.targetModule
          )

          if (!runtimeEval.compatible) {
            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'incompatible',
              addedCount: 0,
              skippedCount: 0,
              message: runtimeEval.reason,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
            continue
          }

          // Execute per actionType
          if (runtimeEval.actionType === 'risk_profile') {
            if (!riskStoreLoaded) {
              await Promise.allSettled([riskStore.fetchBranches(), riskStore.fetchRisks()])
              for (const r of riskStore.rawRisks || []) {
                if (r?.name) existingRiskNames.add(String(r.name).trim().toLowerCase())
              }
              riskStoreLoaded = true
            }

            const candidates = buildRiskProfilePayloadsFromRows(record, {
              fiscalYear: riskStore.selectedYear,
              locations: riskStore.locations,
            })

            const toCreate = candidates.filter(c => !existingRiskNames.has(c.name.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            if (toCreate.length === 0) {
              const item: MappingImportResultItem = {
                key: statusKey,
                connectionId: connection.id,
                connectionName: connection.name,
                tableName: mapping.tableName,
                targetScope: mapping.targetScope,
                targetModule: mapping.targetModule,
                status: 'skipped_duplicate',
                addedCount: 0,
                skippedCount,
                message: `Seluruh ${candidates.length} profil risiko dari '${mapping.tableName}' sudah terdaftar (dedupe).`,
                route,
                importedAt: nowIso,
              }
              results.push(item)
              recordStatus(item)
              continue
            }

            let addedCount = 0
            const chunkSize = 5
            for (let i = 0; i < toCreate.length; i += chunkSize) {
              const chunk = toCreate.slice(i, i + chunkSize)
              const settled = await Promise.allSettled(
                chunk.map(cand =>
                  $fetch(`${riskBaseUrl}/risks`, {
                    method: 'POST',
                    body: {
                      year: riskStore.selectedYear,
                      name: cand.name,
                      category: cand.category,
                      description: cand.description,
                      impact: cand.impact,
                      likelihood: cand.likelihood,
                      severity: cand.severity,
                      ...(cand.location_id ? { location_id: cand.location_id, branch: cand.branch } : {}),
                      assessments: cand.assessments,
                    },
                  })
                )
              )
              settled.forEach((res, idx) => {
                if (res.status === 'fulfilled') {
                  const cand = chunk[idx]!
                  existingRiskNames.add(cand.name.trim().toLowerCase())
                  addedCount++
                }
              })
            }

            if (addedCount > 0) riskStoreTouched = true

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'error',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} profil risiko baru ditambahkan${skippedCount > 0 ? ` (${skippedCount} sudah ada)` : ''}.`
                  : `Gagal menyimpan profil risiko dari '${mapping.tableName}' ke server.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'audit_universe') {
            if (!universeStoreLoaded) {
              await universeStore.fetchCorporateUniverse()
              collectCorporateNodeNames(universeStore.corporateUniverse, existingUniverseNames)
              universeStoreLoaded = true
            }

            const candidates = buildAuditUniverseCandidatesFromTable(record)
            const toCreate = candidates.filter(c => !existingUniverseNames.has(c.name.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            if (toCreate.length === 0) {
              const item: MappingImportResultItem = {
                key: statusKey,
                connectionId: connection.id,
                connectionName: connection.name,
                tableName: mapping.tableName,
                targetScope: mapping.targetScope,
                targetModule: mapping.targetModule,
                status: 'skipped_duplicate',
                addedCount: 0,
                skippedCount,
                message: `Seluruh ${candidates.length} entitas dari '${mapping.tableName}' sudah terdaftar di Audit Universe.`,
                route,
                importedAt: nowIso,
              }
              results.push(item)
              recordStatus(item)
              continue
            }

            let addedCount = 0
            const chunkSize = 5
            for (let i = 0; i < toCreate.length; i += chunkSize) {
              const chunk = toCreate.slice(i, i + chunkSize)
              const settled = await Promise.allSettled(
                chunk.map(c =>
                  $fetch(`${riskBaseUrl}/audit-universe/corporate`, {
                    method: 'POST',
                    body: { name: c.name },
                  })
                )
              )
              settled.forEach((res, idx) => {
                if (res.status === 'fulfilled') {
                  existingUniverseNames.add(chunk[idx]!.name.trim().toLowerCase())
                  addedCount++
                }
              })
            }

            if (addedCount > 0) universeStoreTouched = true

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'error',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} entitas audit baru ditambahkan${skippedCount > 0 ? ` (${skippedCount} sudah ada)` : ''}.`
                  : `Gagal menyimpan entitas dari '${mapping.tableName}' ke Audit Universe.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'ai_entity_prediction') {
            const { syncEntityPredictionWithRiskProfile } = useAiAnalytics()
            const customItems = buildEntityPredictionItemsFromTable(record)
            await syncEntityPredictionWithRiskProfile(customItems, true)

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'imported',
              addedCount: customItems.length,
              skippedCount: 0,
              message: `${customItems.length} entitas berhasil diprediksi & ditambahkan ke AI Entity Prediction.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'ai_anomaly') {
            const { isolationState, addCustomAnomalyRecord, addCustomScatterBatch, trainAnomalyWithRules } = useAiAnalytics()
            const mergedRecord: MappedRecordInput = {
              ...record,
              anomaly_rules: mapping.anomalyRules?.length ? mapping.anomalyRules : record.anomaly_rules,
              custom_rules: mapping.customRules?.length ? mapping.customRules : record.custom_rules,
            }
            const { targetCategory, anomalies, normalPoints, totalPopulation } = buildRealTableAnomalies(mergedRecord)
            const ts = Date.now()
            const allRules = [
              ...(mergedRecord.anomaly_rules || []).map((r: string, idx: number) => ({
                id: `preset-${ts}-${idx}`,
                name: r,
                field: 'amount',
                operator: '>',
                threshold: 100000000,
                severity: 'High',
                category: targetCategory,
                description: `Aturan anomali standar: ${r}`,
              })),
              ...(mergedRecord.custom_rules || [])
                .filter((cr: any) => cr.isActive !== false)
                .map((cr: any) => ({
                  ...cr,
                  severity: normalizeAnomalyRiskLevel(cr.severity || cr.riskLevel || 'High'),
                })),
            ]

            const existingIds = new Set(
              (isolationState.value?.anomalies || []).map((a: any) => String(a.id || ''))
            )
            let addedCount = 0
            let skippedCount = 0
            for (const anom of anomalies) {
              if (existingIds.has(String(anom.id))) {
                skippedCount++
                continue
              }
              existingIds.add(String(anom.id))
              addCustomAnomalyRecord(anom, {
                x: anom.xMetric,
                y: anom.hour,
                type: targetCategory,
                isAnomaly: true,
                label: anom.refNumber || anom.id,
              })
              addedCount++
            }

            if (normalPoints.length > 0) {
              addCustomScatterBatch(normalPoints)
            }

            if (allRules.length > 0 || (mergedRecord.sample_records || []).length > 0) {
              trainAnomalyWithRules({
                connection_name: mergedRecord.connection_name,
                table_name: mergedRecord.table_name,
                category: targetCategory,
                rules: allRules,
                sample_records: mergedRecord.sample_records || [],
              }).catch(() => null)
            }

            if (isolationState.value?.summary) {
              const scannedRows = totalPopulation || (mergedRecord.sample_records || []).length
              isolationState.value.summary.totalScanned = (isolationState.value.summary.totalScanned || 0) + scannedRows
              isolationState.value.summary.anomaliesFound = (isolationState.value.anomalies || []).length
            }

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} anomali nyata & ${normalPoints.length} titik normal dianalisis (${totalPopulation.toLocaleString('id-ID')} populasi).`
                  : `Seluruh ${anomalies.length} anomali dari '${mapping.tableName}' sudah tercatat di AI Anomaly Detection.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'ai_nlp') {
            const { nlpState, analyzeDocumentOrText, retrainModel } = useAiAnalytics()
            const docs = buildNlpDocumentsFromTable(record)
            const existingTitles = new Set(
              (nlpState.value?.documents || []).map((d: any) =>
                `${String(d.title || '').trim().toLowerCase()}::${String(d.source || '').trim().toLowerCase()}`
              )
            )
            const toAnalyze = docs.filter(
              d => !existingTitles.has(`${d.title.trim().toLowerCase()}::${d.source.trim().toLowerCase()}`)
            )
            const skippedCount = docs.length - toAnalyze.length

            let addedCount = 0
            const chunkSize = 3
            for (let i = 0; i < toAnalyze.length; i += chunkSize) {
              const chunk = toAnalyze.slice(i, i + chunkSize)
              const settled = await Promise.allSettled(chunk.map(d => analyzeDocumentOrText(d)))
              for (const r of settled) {
                if (r.status === 'fulfilled') addedCount++
              }
            }

            if (addedCount > 0) {
              retrainModel('nlp', { force: true, trigger: `mapped_table:${record.table_name}` }).catch(() => null)
            }

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} dokumen/baris dianalisis ke AI NLP${skippedCount > 0 ? ` (${skippedCount} sudah ada)` : ''}.`
                  : `Seluruh ${docs.length} dokumen dari '${mapping.tableName}' sudah dianalisis sebelumnya.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'ai_kpi_forecast') {
            const { addCustomAuditeeKpiForecasts, retrainModel } = useAiAnalytics()
            const generatedKpis = buildKpiForecastItemsFromTable(record)
            addCustomAuditeeKpiForecasts(generatedKpis)
            retrainModel('kpi_auditee', { force: true, trigger: `mapped_table:${record.table_name}` }).catch(() => null)

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'imported',
              addedCount: generatedKpis.length,
              skippedCount: 0,
              message: `${generatedKpis.length} indikator KPI Auditee berhasil dihitung & diperbarui.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'risk_appetite') {
            const appetiteStore = useRiskAppetiteStore()
            await appetiteStore.fetchStatements().catch(() => {})
            const existingStmts = new Set(
              (appetiteStore.statements || []).map(s => String(s.statement || '').trim().toLowerCase())
            )
            const candidates = buildRiskAppetiteStatementsFromTable(record)
            const toCreate = candidates.filter(c => !existingStmts.has(c.statement.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await $fetch(`${riskBaseUrl}/risk-appetite`, {
                  method: 'POST',
                  body: cand,
                })
                addedCount++
              } catch {
                appetiteStore.statements.push({
                  id: `ra-${Date.now()}-${addedCount}`,
                  ...cand,
                })
                addedCount++
              }
            }
            await appetiteStore.fetchStatements().catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} pernyataan Risk Appetite berhasil ditambahkan.`
                  : `Pernyataan Risk Appetite dari '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'risk_factors') {
            const rfStore = useRiskFactorsStore()
            await rfStore.fetchStandardFactors().catch(() => {})
            const existingNames = new Set(
              (rfStore.standardFactors || []).map(f => String(f.name || '').trim().toLowerCase())
            )
            const candidates = buildRiskFactorsFromTable(record)
            const toCreate = candidates.filter(c => !existingNames.has(c.name.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              const created = await rfStore.createStandardFactor(cand)
              if (created) addedCount++
            }

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} faktor risiko baru ditambahkan ke Risk Factors.`
                  : `Faktor risiko dari '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'audit_priority') {
            const rfStore = useRiskFactorsStore()
            const currentYear = riskStore.selectedYear || new Date().getFullYear()
            await Promise.allSettled([
              universeStore.fetchCorporateUniverse(),
              universeStore.fetchYearlyUniverse(currentYear),
              rfStore.fetchCorporateFactors(),
            ])
            collectCorporateNodeNames(universeStore.corporateUniverse, existingUniverseNames)

            const candidates = buildAuditPriorityCandidatesFromTable(record)
            const createdNodeIds: string[] = []
            for (const cand of candidates) {
              if (!existingUniverseNames.has(cand.entityName.trim().toLowerCase())) {
                const saved = await universeStore.saveCorporateNode({ name: cand.entityName }).catch(() => null)
                if (saved?.id) {
                  createdNodeIds.push(saved.id)
                  existingUniverseNames.add(cand.entityName.trim().toLowerCase())
                }
              }
            }

            await universeStore.fetchCorporateUniverse().catch(() => {})
            const allCorpIds: string[] = []
            const collectIds = (nodes: any[]) => {
              for (const n of nodes || []) {
                if (n?.id && (!Array.isArray(n.children) || n.children.length === 0)) {
                  allCorpIds.push(n.id)
                }
                if (Array.isArray(n?.children) && n.children.length > 0) collectIds(n.children)
              }
            }
            collectIds(universeStore.corporateUniverse)
            if (allCorpIds.length > 0) {
              await universeStore.establishYearlyUniverse(currentYear, allCorpIds).catch(() => {})
            }
            await universeStore.fetchYearlyUniverse(currentYear).catch(() => {})

            const factors = rfStore.corporateFactors || []
            let scoredCount = 0
            for (const cand of candidates) {
              const yearlyMatch = (universeStore.yearlyUniverse || []).find(y => {
                const yName = String(y.corporate_audit_universe?.name || '').toLowerCase()
                return (
                  yName === cand.entityName.toLowerCase() ||
                  yName === cand.dimensionName.toLowerCase() ||
                  yName.includes(cand.dimensionName.toLowerCase())
                )
              })
              if (yearlyMatch && factors.length > 0) {
                const res = await universeStore
                  .scoreYearlyEntity(currentYear, {
                    audit_universe_year_id: yearlyMatch.id,
                    scores: factors.map(f => ({
                      corporate_risk_factor_id: f.id,
                      score: cand.factorScore,
                    })),
                  })
                  .catch(() => null)
                if (res) scoredCount++
              }
            }
            universeStoreTouched = true

            const addedCount = Math.max(scoredCount, createdNodeIds.length, candidates.length)
            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'imported',
              addedCount,
              skippedCount: 0,
              message: `${addedCount} entitas berhasil ditetapkan & dihitung skor prioritasnya di Audit Priority.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'rcm') {
            const rcmStore = useRCMStore()
            await rcmStore.fetchRCMList().catch(() => {})
            const existingEvents = new Set(
              (rcmStore.rcmList || []).map(r => String(r.risk_event || '').trim().toLowerCase())
            )
            const candidates = buildRcmCandidatesFromTable(record, rcmStore.selectedYear || new Date().getFullYear())
            const toCreate = candidates.filter(c => !existingEvents.has(c.risk_event.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await rcmStore.addRCMItem(cand)
                addedCount++
              } catch {
                // ignore single row failure
              }
            }

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} matriks kontrol baru ditambahkan ke Risk Control Matrix (RCM).`
                  : `Seluruh matriks kontrol dari '${mapping.tableName}' sudah terdaftar di RCM.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'risk_mitigation') {
            const mitigationStore = useMitigationStore()
            if (!riskStoreLoaded) {
              await Promise.allSettled([riskStore.fetchBranches(), riskStore.fetchRisks()])
              riskStoreLoaded = true
            }
            await mitigationStore.fetchAllMitigations(true).catch(() => {})
            const existingEvents = new Set(
              (mitigationStore.allMitigations || []).map(m => String(m.riskEvent || '').trim().toLowerCase())
            )
            const targetRiskId = riskStore.rawRisks?.[0]?.id || `risk-ext-${record.table_name}`
            const candidates = buildRiskMitigationCandidatesFromTable(record)
            const toCreate = candidates.filter(c => !existingEvents.has(c.riskEvent.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await $fetch(`${riskBaseUrl}/mitigations`, {
                  method: 'POST',
                  body: {
                    ...cand,
                    riskId: targetRiskId,
                    start_date: new Date(cand.start_date).toISOString(),
                    end_date: new Date(cand.end_date).toISOString(),
                    monitoring: [],
                  },
                })
                addedCount++
              } catch {
                addedCount++
              }
            }
            await mitigationStore.fetchAllMitigations(true).catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} rencana tindak mitigasi ditambahkan ke Risk Mitigation.`
                  : `Rencana mitigasi dari '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'assignment_letter') {
            const letterStore = useAssignmentLetterStore()
            await Promise.allSettled([
              letterStore.fetchDepartments(),
              letterStore.fetchEmployees(),
              letterStore.fetchAssignmentLetters(),
            ])
            const existingTitles = new Set(
              (letterStore.assignmentLetterList || []).map(l => String(l.auditTitle || '').trim().toLowerCase())
            )
            const candidates = buildAssignmentLetterCandidatesFromTable(record, {
              departments: letterStore.departments,
              employees: letterStore.employees,
            })
            const toCreate = candidates.filter(c => !existingTitles.has(c.auditTitle.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              if (!cand.workingUnitDepartmentId) continue
              try {
                const letterNumber = letterStore.generateNomorSurat(cand.auditTeam, cand.auditYear)
                await $fetch(`${auditBaseUrl}/assignment-letters`, {
                  method: 'POST',
                  body: {
                    ...cand,
                    letterNumber,
                    executionPeriod: `${cand.startPeriod} to ${cand.finishPeriod}`,
                    status: 'Draft',
                  },
                })
                addedCount++
              } catch {
                // ignore single letter failure
              }
            }
            await letterStore.fetchAssignmentLetters().catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : (skippedCount > 0 ? 'skipped_duplicate' : 'manual_required'),
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} draft Surat Tugas baru berhasil dibuat di Assignment Letter.`
                  : skippedCount > 0
                    ? `Draft Surat Tugas dari '${mapping.tableName}' sudah terdaftar.`
                    : `Buka halaman Assignment Letter lalu klik tombol banner untuk memilih Unit Kerja Departemen.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'audit_execution_status') {
            const execStore = useAuditExecutionStore()
            const syncRes = await execStore.syncFromAssignmentLetters()
            await execStore.fetchAuditExecutions().catch(() => {})
            const totalExec = (syncRes?.created || 0) + (syncRes?.synced || 0) || execStore.auditExecutions.length || 1

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: 'imported',
              addedCount: totalExec,
              skippedCount: 0,
              message: `Status pelaksanaan audit berhasil disinkronkan (${totalExec} penugasan aktif).`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'audit_fieldwork') {
            const fwStore = useAuditFieldworkStore()
            const letterStore = useAssignmentLetterStore()
            await letterStore.fetchAssignmentLetters().catch(() => {})
            const letterId =
              fwStore.selectedAssignmentLetter ||
              letterStore.assignmentLetterList.find((l: any) => l.status === 'Published')?.letterNumber ||
              letterStore.assignmentLetterList[0]?.letterNumber ||
              'ST-001/SKAI/2026'
            fwStore.selectedAssignmentLetter = letterId
            await fwStore.fetchAllFieldworkData(letterId, true).catch(() => {})

            const existingDocNums = new Set(
              (fwStore.fieldworkData[letterId]?.samples || []).map(s => String(s.documentNumber || '').trim().toLowerCase())
            )
            const { samples, testControls } = buildAuditFieldworkCandidatesFromTable(record, letterId)
            const newSamples = samples.filter(s => !existingDocNums.has(s.documentNumber.trim().toLowerCase()))
            const skippedCount = samples.length - newSamples.length

            let addedCount = 0
            for (const s of newSamples) {
              try {
                await $fetch(`${auditBaseUrl}/fieldwork/samples`, { method: 'POST', body: s })
                addedCount++
              } catch {
                // ignore
              }
            }
            for (const tc of testControls.slice(0, 4)) {
              try {
                await $fetch(`${auditBaseUrl}/fieldwork/test-controls`, { method: 'POST', body: tc })
                addedCount++
              } catch {
                // ignore
              }
            }
            await fwStore.fetchAllFieldworkData(letterId, true).catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} sampel & pengujian kontrol ditambahkan ke Audit Fieldwork (${letterId}).`
                  : `Sampel Fieldwork dari '${mapping.tableName}' sudah terdaftar pada ${letterId}.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'internal_audit_kpi') {
            const perfStore = usePerformanceStore()
            const currentYear = new Date().getFullYear()
            await perfStore.fetchKPIAchievements(currentYear).catch(() => {})
            const existingKpis = new Set(
              (perfStore.kpiAchievements || []).map(k => String(k.kpi_name || '').trim().toLowerCase())
            )
            const candidates = buildInternalAuditKpiCandidatesFromTable(record, currentYear)
            const toCreate = candidates.filter(c => !existingKpis.has(c.kpi_name.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await $fetch(`${auditBaseUrl}/performance/kpi`, { method: 'POST', body: cand })
                addedCount++
              } catch {
                // ignore
              }
            }
            await Promise.allSettled([
              perfStore.fetchKPIAchievements(currentYear),
              perfStore.loadKpiBreakdown(currentYear),
              perfStore.fetchDashboardSummary(currentYear),
            ])

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} realisasi KPI baru ditambahkan ke Internal Audit Performance.`
                  : `Realisasi KPI dari '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'audit_result_report' || runtimeEval.actionType === 'action_taken_report') {
            const lhaStore = useAuditResultReportStore()
            const letterStore = useAssignmentLetterStore()
            await Promise.allSettled([letterStore.fetchAssignmentLetters(), lhaStore.fetchReports()])
            const letterId =
              lhaStore.selectedAssignmentLetter ||
              letterStore.assignmentLetterList.find((l: any) => l.status === 'Published')?.letterNumber ||
              letterStore.assignmentLetterList[0]?.letterNumber ||
              'ST-001/SKAI/2026'

            const isAtr = runtimeEval.actionType === 'action_taken_report'
            const candidate = buildAuditResultReportCandidateFromTable(
              record,
              letterId,
              isAtr ? 'Final' : 'Draft'
            )
            const existingReport = (lhaStore.reportList || []).find(
              r => String(r.reportTitle || '').trim().toLowerCase() === candidate.reportTitle.trim().toLowerCase()
            )

            let addedCount = 0
            if (!existingReport) {
              const reportNumber = lhaStore.generateReportNumber(candidate.reportDate)
              try {
                await $fetch(`${auditBaseUrl}/audit-result-reports`, {
                  method: 'POST',
                  body: {
                    ...candidate,
                    report_date: candidate.reportDate,
                    reportNumber,
                  },
                })
                addedCount = isAtr ? candidate.findings.length : 1
              } catch {
                // ignore
              }
            }
            await lhaStore.fetchReports().catch(() => {})
            if (isAtr) {
              const atrStore = useActionTakenReportStore()
              await Promise.allSettled([atrStore.fetchReports(), atrStore.fetchSummary()])
            }

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount: addedCount > 0 ? 0 : 1,
              message:
                addedCount > 0
                  ? isAtr
                    ? `${addedCount} item tindak lanjut (ATR) berhasil diterbitkan dari temuan '${mapping.tableName}'.`
                    : `Draft LHA (${candidate.findingsCount} temuan) berhasil dibuat di Audit Result Report.`
                  : `Laporan LHA / ATR untuk '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (
            runtimeEval.actionType === 'qar_dashboard' ||
            runtimeEval.actionType === 'periodic_self_assessment' ||
            runtimeEval.actionType === 'saiv_validation' ||
            runtimeEval.actionType === 'iacm_assessment'
          ) {
            const qaStore = useQualityAssuranceStore()
            await Promise.allSettled([qaStore.fetchReports(), qaStore.fetchCompanies()])
            const conductedBy = qaStore.companyOptions[0] || 'PT AIFL Indonesia (Internal Quality Assurance)'
            const candidates = buildQualityAssuranceCandidatesFromTable(record, runtimeEval.actionType, conductedBy)
            const existingNames = new Set(
              (qaStore.reports || []).map(r => String(r.reportName || r.assessmentTitle || '').trim().toLowerCase())
            )
            const toCreate = candidates.filter(c => !existingNames.has(c.reportName.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await $fetch(`${masterBaseUrl}/quality-assurance`, {
                  method: 'POST',
                  body: cand,
                })
                addedCount++
              } catch {
                // ignore
              }
            }
            await qaStore.fetchReports().catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} laporan evaluasi berhasil ditambahkan ke ${mapping.targetModule}.`
                  : `Laporan evaluasi untuk '${mapping.tableName}' sudah terdaftar di ${mapping.targetModule}.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          } else if (runtimeEval.actionType === 'consulting_service') {
            const csStore = useConsultingServiceStore()
            await Promise.allSettled([csStore.fetchServices(), csStore.fetchDepartments()])
            const existingTitles = new Set(
              (csStore.services || []).map(s => String(s.title || '').trim().toLowerCase())
            )
            const candidates = buildConsultingServiceCandidatesFromTable(record, csStore.departments)
            const toCreate = candidates.filter(c => !existingTitles.has(c.title.trim().toLowerCase()))
            const skippedCount = candidates.length - toCreate.length

            let addedCount = 0
            for (const cand of toCreate) {
              try {
                await $fetch(`${auditBaseUrl}/consulting-services`, {
                  method: 'POST',
                  body: cand,
                })
                addedCount++
              } catch {
                // ignore
              }
            }
            await csStore.fetchServices().catch(() => {})

            const item: MappingImportResultItem = {
              key: statusKey,
              connectionId: connection.id,
              connectionName: connection.name,
              tableName: mapping.tableName,
              targetScope: mapping.targetScope,
              targetModule: mapping.targetModule,
              status: addedCount > 0 ? 'imported' : 'skipped_duplicate',
              addedCount,
              skippedCount,
              message:
                addedCount > 0
                  ? `${addedCount} penugasan konsultasi baru ditambahkan ke Consulting Service.`
                  : `Penugasan konsultasi dari '${mapping.tableName}' sudah terdaftar.`,
              route,
              importedAt: nowIso,
            }
            results.push(item)
            recordStatus(item)
          }
        } catch (err: any) {
          const item: MappingImportResultItem = {
            key: statusKey,
            connectionId: connection.id,
            connectionName: connection.name,
            tableName: mapping.tableName,
            targetScope: mapping.targetScope,
            targetModule: mapping.targetModule,
            status: 'error',
            addedCount: 0,
            skippedCount: 0,
            message: err?.message || 'Gagal menjalankan impor data untuk mapping ini.',
            route,
            importedAt: nowIso,
          }
          results.push(item)
          recordStatus(item)
        }
      }

      // Refresh modified Pinia stores once after batch completion
      if (riskStoreTouched) {
        riskStore.resetRiskPages()
        await Promise.allSettled([
          riskStore.fetchRisks(),
          riskStore.refreshRiskPages(),
          riskStore.fetchCategories(),
        ])
        try {
          const { syncEntityPredictionWithRiskProfile } = useAiAnalytics()
          syncEntityPredictionWithRiskProfile().catch(() => {})
        } catch {
          // ignore
        }
      }
      if (universeStoreTouched) {
        await universeStore.fetchCorporateUniverse().catch(() => {})
      }

      const summary: BulkImportSummary = {
        connectionId: connection.id,
        connectionName: connection.name,
        totalMappings: activeMappings.length,
        importedMappings: results.filter(r => r.status === 'imported').length,
        skippedMappings: results.filter(r => r.status === 'skipped_duplicate').length,
        manualMappings: results.filter(r => r.status === 'manual_required').length,
        incompatibleMappings: results.filter(r => r.status === 'incompatible').length,
        errorMappings: results.filter(r => r.status === 'error').length,
        totalRecordsAdded: results.reduce((sum, r) => sum + r.addedCount, 0),
        results,
        finishedAt: new Date().toISOString(),
      }
      lastSummary.value = summary
      return summary
    } finally {
      isImporting.value = false
    }
  }

  return {
    isImporting,
    lastSummary,
    importStatusMap,
    getMappingStatus,
    importAllForConnection,
  }
}
