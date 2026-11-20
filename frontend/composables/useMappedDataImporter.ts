import { ref } from 'vue'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { getMasterServiceBaseUrl, getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import { normalizeAnomalyRiskLevel } from '~/utils/riskLevelBadge'
import {
  buildAuditUniverseCandidatesFromTable,
  buildEntityPredictionItemsFromTable,
  buildKpiForecastItemsFromTable,
  buildNlpDocumentsFromTable,
  buildRealTableAnomalies,
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

        // 2. Manual confirmation / pre-fill modules (Annual Plan, Strategic Plan, Working Paper)
        if (schemaEval.importMode === 'manual_banner' || schemaEval.importMode === 'prefill_form') {
          const manualMsg =
            schemaEval.actionType === 'annual_plan'
              ? 'Tetap melalui banner dengan konfirmasi manual (Departemen & Audit Universe) di halaman Annual Audit Plan.'
              : schemaEval.actionType === 'strategic_plan'
                ? 'Tetap melalui banner dengan pengisian Target KPI manual di halaman Strategic Audit Plan.'
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
