// @ts-nocheck
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useActiveDataSource } from '~/composables/useActiveDataSource'
import { useDataSourceStore } from '~/stores/data-source'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      masterServiceUrl: 'http://localhost:8001/api/v1'
    }
  })
}))

global.$fetch = vi.fn()

describe('Composable: useActiveDataSource', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    if (typeof localStorage !== 'undefined') {
      localStorage.clear()
    }
  })

  it('should expose reactive refs and methods from useDataSourceStore', () => {
    const { sources, selectedSourceId, isAllSources, selectedSource, setSelectedSourceId } = useActiveDataSource()

    expect(sources.value).toEqual([])
    expect(selectedSourceId.value).toBe('all')
    expect(isAllSources.value).toBe(true)
    expect(selectedSource.value).toBeNull()

    const store = useDataSourceStore()
    store.sources = [
      {
        id: 'ds-test',
        name: 'Test DB',
        source_type: 'postgres',
        host: '127.0.0.1',
        port: 5432,
        database: 'test_db'
      }
    ]

    setSelectedSourceId('ds-test')
    expect(selectedSourceId.value).toBe('ds-test')
    expect(isAllSources.value).toBe(false)
    expect(selectedSource.value?.name).toBe('Test DB')
  })
})

