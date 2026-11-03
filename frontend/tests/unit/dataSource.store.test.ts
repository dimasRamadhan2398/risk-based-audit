// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useDataSourceStore } from '~/stores/data-source'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      masterServiceUrl: 'http://localhost:8001/api/v1'
    }
  })
}))

global.$fetch = vi.fn()

describe('DataSource Store (useDataSourceStore)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    if (typeof localStorage !== 'undefined') {
      localStorage.clear()
    }
  })

  it('should initialize with default state', () => {
    const store = useDataSourceStore()
    expect(store.sources).toEqual([])
    expect(store.selectedSourceId).toBe('all')
    expect(store.isAllSources).toBe(true)
    expect(store.selectedSource).toBeNull()
    expect(store.isLoading).toBe(false)
  })

  it('should setSelectedSourceId and update selectedSource and localStorage', () => {
    const store = useDataSourceStore()
    store.sources = [
      {
        id: 'ds-1',
        name: 'Core Banking Prod Postgres',
        source_type: 'postgres',
        host: '10.20.0.14',
        port: 5432,
        database: 'core_banking_db',
        status: 'Connected'
      },
      {
        id: 'ds-2',
        name: 'ERP SAP Database',
        source_type: 'postgres',
        host: '10.20.0.88',
        port: 5432,
        database: 'sap_erp_prod',
        status: 'Connected'
      }
    ]

    store.setSelectedSourceId('ds-1')

    expect(store.selectedSourceId).toBe('ds-1')
    expect(store.isAllSources).toBe(false)
    expect(store.selectedSource).toBeDefined()
    expect(store.selectedSource?.name).toBe('Core Banking Prod Postgres')
    expect(localStorage.getItem('auditsphere_selected_source_id')).toBe('ds-1')

    // Switching back to all
    store.setSelectedSourceId('all')
    expect(store.selectedSourceId).toBe('all')
    expect(store.isAllSources).toBe(true)
    expect(store.selectedSource).toBeNull()
    expect(localStorage.getItem('auditsphere_selected_source_id')).toBe('all')
  })

  it('fetchSources should retrieve connections from API and populate sources', async () => {
    const store = useDataSourceStore()

    const mockDataSources = [
      {
        id: '11111111-2222-3333-4444-555555555555',
        name: 'Core Banking DB',
        type: 'postgres',
        host: 'localhost',
        port: 5432,
        database: 'bank_db',
        status: 'Connected'
      }
    ]

    vi.mocked($fetch).mockResolvedValueOnce({
      success: true,
      data: mockDataSources
    })

    await store.fetchSources()

    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/data-sources'),
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
    expect(store.sources).toHaveLength(1)
    expect(store.sources[0].name).toBe('Core Banking DB')
    expect(store.isLoading).toBe(false)
  })

  it('fetchSources should handle failure gracefully without crashing', async () => {
    const store = useDataSourceStore()

    vi.mocked($fetch).mockRejectedValueOnce(new Error('Network error'))

    await store.fetchSources()

    expect(store.sources).toEqual([])
    expect(store.isLoading).toBe(false)
  })
})

