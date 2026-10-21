/**
 * useEmployeeApi.getAllEmployees: GET /employees caps page_size at 100 and returns
 * data.pagination. The dropdown loader used to ask for page_size=1000 in one request, so
 * employees past the first 100 silently disappeared. It now walks every page.
 */
import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { useEmployeeApi } from '~/composables/useEmployeeApi'

type FetchOpts = { method?: string, params?: Record<string, number | string> }
type FetchFn = (url: string, opts?: FetchOpts) => Promise<unknown>
const g = globalThis as unknown as { $fetch: Mock<FetchFn> }

const BACKEND_CAP = 100

/** Mimics master-service: page_size defaults to 10, capped at 100, pages are 1-based. */
const backend = (total: number) => (url: string, opts?: FetchOpts) => {
  expect(url.endsWith('/employees')).toBe(true)
  const pageSize = Math.min(Number(opts?.params?.page_size ?? 10), BACKEND_CAP)
  const page = Number(opts?.params?.page ?? 1)
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const start = (page - 1) * pageSize
  const count = Math.max(0, Math.min(pageSize, total - start))
  const employees = Array.from({ length: count }, (_, i) => ({
    id: `e${start + i + 1}`,
    full_name: `Employee ${start + i + 1}`,
    is_active: true
  }))
  return Promise.resolve({
    success: true,
    data: { employees, pagination: { page, page_size: pageSize, total, total_pages: totalPages } }
  })
}

describe('useEmployeeApi.getAllEmployees', () => {
  beforeEach(() => {
    g.$fetch = vi.fn(backend(0)) as unknown as Mock<FetchFn>
  })

  it('collects every page: 250 employees across 3 pages returns 250', async () => {
    g.$fetch = vi.fn(backend(250)) as unknown as Mock<FetchFn>

    const list = await useEmployeeApi().getAllEmployees()

    expect(list).toHaveLength(250)
    expect(list[0]?.id).toBe('e1')
    expect(list[249]?.id).toBe('e250')
    expect(g.$fetch).toHaveBeenCalledTimes(3)
    const pages = (g.$fetch as unknown as Mock).mock.calls.map(c => (c[1] as FetchOpts).params)
    expect(pages).toEqual([
      { page: 1, page_size: 100 },
      { page: 2, page_size: 100 },
      { page: 3, page_size: 100 }
    ])
  })

  it('makes a single request when everything fits on one page', async () => {
    g.$fetch = vi.fn(backend(42)) as unknown as Mock<FetchFn>

    const list = await useEmployeeApi().getAllEmployees()

    expect(list).toHaveLength(42)
    expect(g.$fetch).toHaveBeenCalledTimes(1)
  })

  it('returns an empty array without looping when there are no employees', async () => {
    g.$fetch = vi.fn(backend(0)) as unknown as Mock<FetchFn>

    const list = await useEmployeeApi().getAllEmployees()

    expect(list).toEqual([])
    expect(g.$fetch).toHaveBeenCalledTimes(1)
  })

  it('stops on an empty page even if total_pages claims more pages exist', async () => {
    g.$fetch = vi.fn(async (_url: string, opts?: FetchOpts) => {
      const page = Number(opts?.params?.page ?? 1)
      const employees = page === 1 ? [{ id: 'e1', full_name: 'Only One' }] : []
      return { data: { employees, pagination: { page, page_size: 100, total: 1, total_pages: 5 } } }
    }) as unknown as Mock<FetchFn>

    const list = await useEmployeeApi().getAllEmployees()

    expect(list.map(e => e.id)).toEqual(['e1'])
    expect(g.$fetch).toHaveBeenCalledTimes(2)
  })

  it('stops after one page when the response carries no pagination info', async () => {
    g.$fetch = vi.fn(async () => ({ data: [{ id: 'e1' }, { id: 'e2' }] })) as unknown as Mock<FetchFn>

    const list = await useEmployeeApi().getAllEmployees()

    expect(list.map(e => e.id)).toEqual(['e1', 'e2'])
    expect(g.$fetch).toHaveBeenCalledTimes(1)
  })
})
