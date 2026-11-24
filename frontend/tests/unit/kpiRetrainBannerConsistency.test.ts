// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import {
  computeAuditorFingerprint,
  computeAuditeeCustomFingerprint,
  dedupAndBuildAuditorPayloadItems,
  buildAuditorLocalForecastsFromItems
} from '~/utils/auditorKpiSync'

const KPI_FORECAST_PAGE = resolve(__dirname, '../../pages/analytics/ai/kpi-forecast.vue')
const PROD_KONG_YML = resolve(__dirname, '../../../backend/kong-gateway/kong/prod/kong.yml')
const PROD_COMPOSE_YML = resolve(__dirname, '../../../backend/docker-compose.prod.yml')

describe('KPI Retrain Banner & Production Error Consistency', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    global.$fetch = vi.fn(async () => null)
  })

  it('reconciles stale +1 pembaruan on kpi_auditee and kpi_auditor when hashes already match trained state', async () => {
    const strategicPlanRows = [
      {
        code: 'SO-IA01',
        kpi: 'Realisasi PKAT',
        category: 'Execution',
        unit: '%',
        hibHig: 'HIG',
        target: '90',
        actual: '85'
      }
    ]
    const auditorHash = computeAuditorFingerprint(dedupAndBuildAuditorPayloadItems(strategicPlanRows, [], 2026))
    const customAuditee = [
      { code: 'KPI-EXT-01', kpiName: 'NPL Portofolio', entity: 'Cabang A', currentValue: 4.1, forecastedValue: 4.5, riskLevel: 'HIGH' }
    ]
    localStorage.setItem('auditsphere_custom_auditee_kpis_v1', JSON.stringify(customAuditee))
    const customHash = computeAuditeeCustomFingerprint(customAuditee)

    global.$fetch = vi.fn(async (url: string, opts?: any) => {
      const u = String(url)
      if (u.includes('performance-trend/batch')) {
        return {
          auditee_forecast: { kpi_forecasts: [{ code: 'KPI-001', kpiName: 'NPL', entity: 'Surabaya', riskLevel: 'HIGH', currentValue: 3.1, forecastedValue: 3.5 }], at_risk_departments: [] },
          auditor_forecast: { kpi_forecasts: [] }
        }
      }
      if (u.includes('/strategic-plans')) {
        return { status: 'success', data: { items: strategicPlanRows, pagination: { page: 1, page_size: 100, total: 1, total_pages: 1 } } }
      }
      if (u.includes('/performance/kpi')) {
        return { status: 'success', data: { items: [], pagination: { page: 1, page_size: 100, total: 0, total_pages: 1 } } }
      }
      if (u.includes('/performance-trend/auditor-forecast')) {
        return { status: 'success', data: buildAuditorLocalForecastsFromItems(opts.body.items, 2026) }
      }
      if (u.includes('/retrain/status')) {
        // Simulate older backend dual-hash bug returning state='untrained_data', new_rows=1 even though hashes match
        return {
          status: 'completed',
          is_running: false,
          trained_auditor_hash: auditorHash,
          trained_auditee_pool_hash: 'pool123',
          trained_auditee_custom_hash: customHash,
          has_untrained_auditee_pool_data: false,
          trained_auditor_count: 1,
          trained_auditee_count: 2232,
          models: {
            kpi_auditee: {
              key: 'kpi_auditee',
              label: 'KPI Forecast (Auditee)',
              state: 'untrained_data',
              has_new_data: true,
              new_rows: 1,
              trained_rows: 2232,
              current_rows: 2232
            },
            kpi_auditor: {
              key: 'kpi_auditor',
              label: 'KPI Forecast (Auditor)',
              state: 'untrained_data',
              has_new_data: true,
              new_rows: 1,
              trained_rows: 6112,
              current_rows: 6112
            }
          }
        }
      }
      return null
    })

    const { fetchAiAnalytics, hasUntrainedNewData, modelStatuses, kpiRetrainError } = useAiAnalytics()
    await fetchAiAnalytics(true, { skipAutoRetrain: true })

    expect(hasUntrainedNewData.value).toBe(false)
    expect(kpiRetrainError.value).toBeNull()
    expect(modelStatuses.value.kpi_auditee.state).toBe('trained')
    expect(modelStatuses.value.kpi_auditee.has_new_data).toBe(false)
    expect(modelStatuses.value.kpi_auditee.new_rows).toBe(0)
    expect(modelStatuses.value.kpi_auditor.state).toBe('trained')
    expect(modelStatuses.value.kpi_auditor.has_new_data).toBe(false)
    expect(modelStatuses.value.kpi_auditor.new_rows).toBe(0)
  })

  it('isolates unrelated model failures (e.g. entity/nlp in prod) from KPI Forecast error alert', async () => {
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('/retrain/status')) {
        return {
          status: 'failed',
          is_running: false,
          last_error: 'Pelatihan Entity Prediction tidak memenuhi ambang validasi.',
          failed_models: ['entity'],
          trained_auditor_hash: 'h1',
          trained_auditee_pool_hash: 'p1',
          trained_auditee_custom_hash: '',
          has_untrained_auditee_pool_data: false,
          models: {
            entity: { key: 'entity', label: 'Entity Prediction', state: 'failed', last_error: 'Pelatihan Entity Prediction tidak memenuhi ambang validasi.' },
            kpi_auditee: { key: 'kpi_auditee', label: 'KPI Forecast (Auditee)', state: 'trained', has_new_data: false, last_error: null },
            kpi_auditor: { key: 'kpi_auditor', label: 'KPI Forecast (Auditor)', state: 'trained', has_new_data: false, last_error: null }
          }
        }
      }
      return null
    })

    const { fetchRetrainStatus, kpiRetrainError } = useAiAnalytics()
    await fetchRetrainStatus()

    // KPI models are healthy -> KPI Forecast page must NOT show "Pelatihan Ulang Model AI Gagal"
    expect(kpiRetrainError.value).toBeNull()
  })

  it('exposes kpiRetrainError and enables retry button when a KPI model genuinely fails', async () => {
    global.$fetch = vi.fn(async (url: string) => {
      if (String(url).includes('/retrain/status')) {
        return {
          status: 'failed',
          is_running: false,
          last_error: 'PyTorch LSTM error',
          failed_models: ['kpi_auditor'],
          trained_auditor_hash: 'h1',
          trained_auditee_pool_hash: 'p1',
          trained_auditee_custom_hash: '',
          has_untrained_auditee_pool_data: false,
          models: {
            kpi_auditee: { key: 'kpi_auditee', label: 'KPI Forecast (Auditee)', state: 'trained', has_new_data: false, last_error: null },
            kpi_auditor: { key: 'kpi_auditor', label: 'KPI Forecast (Auditor)', state: 'failed', has_new_data: false, last_error: 'PyTorch LSTM error' }
          }
        }
      }
      return null
    })

    const { fetchRetrainStatus, kpiRetrainError, canManualRetrain } = useAiAnalytics()
    await fetchRetrainStatus()

    expect(kpiRetrainError.value).toBe('PyTorch LSTM error')
    expect(canManualRetrain.value).toBe(true)
  })

  it('recovers smoothly from reverse-proxy timeout during manual force retrain when server is running or completed', async () => {
    let statusPolls = 0
    global.$fetch = vi.fn(async (url: string) => {
      const u = String(url)
      if (u.includes('/retrain/auto')) {
        // Simulate Kong 504 timeout
        throw new Error('504 Gateway Timeout')
      }
      if (u.includes('/retrain/status')) {
        statusPolls++
        return {
          status: statusPolls === 1 ? 'completed' : 'running',
          is_running: statusPolls > 1,
          trained_auditor_hash: '',
          trained_auditee_pool_hash: 'pool_old',
          trained_auditee_custom_hash: '',
          has_untrained_auditee_pool_data: statusPolls === 1,
          models: {
            kpi_auditee: { key: 'kpi_auditee', label: 'KPI Forecast (Auditee)', state: statusPolls === 1 ? 'untrained_data' : 'training', has_new_data: statusPolls === 1, last_error: null },
            kpi_auditor: { key: 'kpi_auditor', label: 'KPI Forecast (Auditor)', state: 'trained', has_new_data: false, last_error: null }
          }
        }
      }
      return null
    })

    const { fetchRetrainStatus, triggerManualForceRetrain, kpiRetrainError } = useAiAnalytics()
    await fetchRetrainStatus()
    const res = await triggerManualForceRetrain()

    expect(res.status).toBe('already_running')
    expect(kpiRetrainError.value).toBeNull()
  })

  it('enforces unified error/trained state in kpi-forecast.vue and production Kong/Compose config', () => {
    const pageSrc = readFileSync(KPI_FORECAST_PAGE, 'utf-8')
    expect(pageSrc).toContain('v-else-if="kpiRetrainError"')
    expect(pageSrc).not.toContain('v-else-if="retrainStatus.status === \'failed\' && retrainStatus.lastError"')
    expect(pageSrc).toContain('Pelatihan KPI Gagal')
    expect(pageSrc).toContain('Coba Latih Ulang AI')

    const kongSrc = readFileSync(PROD_KONG_YML, 'utf-8')
    expect(kongSrc).toContain('/api/analytics/retrain/status')
    expect(kongSrc).toContain('/api/analytics/retrain/model')
    expect(kongSrc).toContain('/api/analytics/retrain/auto')
    expect(kongSrc).toContain('read_timeout: 180000')

    const prodComposeSrc = readFileSync(PROD_COMPOSE_YML, 'utf-8')
    expect(prodComposeSrc).toContain('API_KEY=${PYTHON_AI_API_KEY:-auditsphere_ai_key_2026_k33f01}')
    expect(prodComposeSrc).toContain('INTERNAL_RETRAIN_TOKEN=${INTERNAL_RETRAIN_TOKEN:-dev-internal-token}')
  })
})
