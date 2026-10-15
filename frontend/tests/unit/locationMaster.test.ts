// @ts-nocheck
/**
 * Branch / Location master data.
 *
 * The backend already serves branch master data at /api/v1/locations; these
 * tests cover the frontend module that exposes it: the store, the sidebar entry
 * and the i18n keys, plus the Corporate Risk Profile branch filter that should
 * follow the master data instead of a hardcoded list.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { setActivePinia, createPinia } from 'pinia'

import { useLocationStore } from '~/stores/location'
import { useRiskProfileStore, ALL_BRANCHES, UNASSIGNED_BRANCH } from '~/stores/risk-profile'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: { masterServiceBaseUrl: 'http://localhost:8080/api/v1' }
  })
}))

const LAYOUT = resolve(__dirname, '../../layouts/default.vue')
const PAGE = resolve(__dirname, '../../pages/master/location/index.vue')
const EN = resolve(__dirname, '../../locales/en/common.json')
const ID = resolve(__dirname, '../../locales/id/common.json')

const seededLocations = [
  { id: 'loc-1', name: 'AIFL Headquarters', address: 'Jl. Jend. Sudirman Kav. 52-53', city: 'Jakarta Selatan', province: 'DKI Jakarta', is_active: true },
  { id: 'loc-2', name: 'Surabaya Branch', address: 'Jl. Ahmad Yani No. 88', city: 'Surabaya', province: 'Jawa Timur', is_active: true },
  { id: 'loc-3', name: 'Bandung Branch', address: 'Jl. Asia Afrika No. 8', city: 'Bandung', province: 'Jawa Barat', is_active: false },
  // Closed branch that no risk references — must not be offered as an option.
  { id: 'loc-4', name: 'Medan Branch (Closed)', address: 'Jl. Putri Hijau No. 1', city: 'Medan', province: 'Sumatera Utara', is_active: false }
]

const okResponse = (data: any) => ({ success: true, message: 'ok', data })

describe('Location (branch) store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    global.$fetch = vi.fn(async () => okResponse(seededLocations))
  })

  it('loads branch master data from GET /locations', async () => {
    const store = useLocationStore()
    await store.fetchLocations()

    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/locations'),
      expect.objectContaining({ method: 'GET' })
    )
    expect(store.locations.map((l) => l.name)).toEqual([
      'AIFL Headquarters',
      'Surabaya Branch',
      'Bandung Branch',
      'Medan Branch (Closed)'
    ])
  })

  it('filters client-side by name, city or province', async () => {
    const store = useLocationStore()
    await store.fetchLocations()

    store.setSearch('surabaya')
    expect(store.filteredLocations.map((l) => l.id)).toEqual(['loc-2'])

    store.setSearch('Jawa')
    expect(store.filteredLocations.map((l) => l.id)).toEqual(['loc-2', 'loc-3'])

    store.setSearch('')
    expect(store.total).toBe(4)
  })

  it('exposes only active branches as branch names', async () => {
    const store = useLocationStore()
    await store.fetchLocations()

    expect(store.branchNames).toEqual(['AIFL Headquarters', 'Surabaya Branch'])
  })

  it('requires name, address and city before submitting', async () => {
    const store = useLocationStore()
    store.openCreateModal()

    expect(await store.handleSubmit()).toBe(false)
    expect(store.errorMsg).toBeTruthy()
    expect($fetch).not.toHaveBeenCalled()

    store.form.name = 'Medan Branch'
    expect(await store.handleSubmit()).toBe(false) // address still missing

    store.form.address = 'Jl. Putri Hijau No. 1'
    expect(await store.handleSubmit()).toBe(false) // city still missing
    expect($fetch).not.toHaveBeenCalled()
  })

  it('posts a complete payload on create and defaults the country', async () => {
    const store = useLocationStore()
    store.openCreateModal()
    store.form.name = 'Medan Branch'
    store.form.address = 'Jl. Putri Hijau No. 1'
    store.form.city = 'Medan'

    expect(await store.handleSubmit()).toBe(true)
    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/locations'),
      expect.objectContaining({
        method: 'POST',
        body: expect.objectContaining({
          name: 'Medan Branch',
          address: 'Jl. Putri Hijau No. 1',
          city: 'Medan',
          country: 'Indonesia',
          is_active: true
        })
      })
    )
    expect(store.showModal).toBe(false)
  })

  it('fills the form from the selected row and PUTs on update', async () => {
    const store = useLocationStore()
    store.handleEdit(seededLocations[1])

    expect(store.isEditing).toBe(true)
    expect(store.form.name).toBe('Surabaya Branch')
    expect(store.form.city).toBe('Surabaya')

    store.form.city = 'Sidoarjo'
    expect(await store.handleSubmit()).toBe(true)
    expect($fetch).toHaveBeenCalledWith(
      expect.stringContaining('/locations/loc-2'),
      expect.objectContaining({
        method: 'PUT',
        body: expect.objectContaining({ city: 'Sidoarjo' })
      })
    )
  })

  it('surfaces an error message when the master service fails', async () => {
    global.$fetch = vi.fn(async () => {
      throw new Error('503 Service Unavailable')
    })

    const store = useLocationStore()
    await store.fetchLocations()

    expect(store.locations).toEqual([])
    expect(store.errorMsg).toBeTruthy()
  })
})

describe('Corporate Risk Profile branch filter', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  const year = new Date().getFullYear()
  const risk = (id: string, extra: any) => ({
    id, name: id, category: 'Operations', impact: 3, likelihood: 3,
    assessments: [{ year }], ...extra
  })

  it('takes its branch options from active Location master rows only', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()

    expect(store.locations).toEqual([
      { id: 'loc-1', name: 'AIFL Headquarters' },
      { id: 'loc-2', name: 'Surabaya Branch' }
    ])
    expect(store.branches).toEqual(['AIFL Headquarters', 'Surabaya Branch'])
    expect(store.branchesError).toBe('')
  })

  it('does not add branch names found on risks to the options', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations')
        ? okResponse([seededLocations[1]])
        : okResponse([risk('r1', { branch: 'Legacy Dummy Branch', location_id: null })])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()
    await store.fetchRisks()

    expect(store.branches).toEqual(['Surabaya Branch'])
    expect(store.branches).not.toContain('Legacy Dummy Branch')
    expect(store.branches).not.toContain('Head Office')
  })

  it('offers no branches (and reports an error) when the master fails to load', async () => {
    global.$fetch = vi.fn(async () => {
      throw new Error('offline')
    })

    const store = useRiskProfileStore()
    await store.fetchBranches()

    expect(store.branches).toEqual([])
    expect(store.locations).toEqual([])
    expect(store.branchesError).toBeTruthy()
  })

  it('filters by location_id, matching the branch name only for unlinked risks', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations')
        ? okResponse(seededLocations)
        : okResponse([
          risk('linked', { location_id: 'loc-2', branch: 'Surabaya Branch' }),
          // location_id wins over a stale name
          risk('renamed', { location_id: 'loc-2', branch: 'Old Surabaya Name' }),
          risk('legacy', { location_id: null, branch: 'Surabaya Branch' }),
          risk('hq', { location_id: 'loc-1', branch: 'AIFL Headquarters' }),
          risk('orphan', { location_id: null, branch: null })
        ])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()
    await store.fetchRisks()

    const ids = () => store.filteredRisks.map((r: any) => r.id)

    store.selectedBranch = ALL_BRANCHES
    expect(ids()).toEqual(['linked', 'renamed', 'legacy', 'hq', 'orphan'])

    store.selectedBranch = 'loc-2'
    expect(ids()).toEqual(['linked', 'renamed', 'legacy'])

    store.selectedBranch = 'loc-1'
    expect(ids()).toEqual(['hq'])

    expect(store.hasUnassignedRisks).toBe(true)
    store.selectedBranch = UNASSIGNED_BRANCH
    expect(ids()).toEqual(['orphan'])
  })

  it('reports no unassigned risks when every risk has a location', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations')
        ? okResponse(seededLocations)
        : okResponse([risk('linked', { location_id: 'loc-2', branch: 'Surabaya Branch' })])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()
    await store.fetchRisks()

    expect(store.hasUnassignedRisks).toBe(false)
  })
})

describe('Corporate Risk Profile has no mock risks', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows an empty list, not seed data, when GET /risks returns nothing', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )

    const store = useRiskProfileStore()
    await store.fetchRisks()

    expect(store.rawRisks).toEqual([])
    expect(store.risksLoadError).toBe('')
  })

  it('shows an empty list and an error when GET /risks fails', async () => {
    global.$fetch = vi.fn(async () => {
      throw new Error('503 Service Unavailable')
    })

    const store = useRiskProfileStore()
    await store.fetchRisks()

    expect(store.rawRisks).toEqual([])
    expect(store.risksLoadError).toBeTruthy()
  })

  it('does not add a local risk when the backend rejects the POST', async () => {
    global.$fetch = vi.fn(async (url: string, opts: any) => {
      if (opts?.method === 'POST') throw new Error('400 Bad Request')
      return String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    })

    const store = useRiskProfileStore()
    await store.fetchRisks()

    const ok = await store.addRisk({ name: 'Ditolak', location_id: 'loc-2', assessments: [] })
    expect(ok).toBe(false)
    expect(store.rawRisks).toEqual([])
    expect(store.errorMsg).toBeTruthy()
  })
})

describe('Risks carry a location reference', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('sends the chosen location_id with its master name when saving a risk', async () => {
    const fetchMock = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )
    global.$fetch = fetchMock

    const store = useRiskProfileStore()
    await store.fetchBranches()

    await store.addRisk({
      name: 'Kegagalan rekonsiliasi kas cabang',
      category: 'Operations',
      location_id: 'loc-2',
      impact: 4,
      likelihood: 3,
      assessments: []
    })

    const post = fetchMock.mock.calls.find(([, opts]: any) => opts?.method === 'POST')
    expect(post, 'no POST /risks was issued').toBeTruthy()
    expect(post[1].body).toMatchObject({
      branch: 'Surabaya Branch',
      location_id: 'loc-2'
    })
  })

  it('resolves location_id from a master branch name', async () => {
    const fetchMock = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )
    global.$fetch = fetchMock

    const store = useRiskProfileStore()
    await store.fetchBranches()

    await store.addRisk({ name: 'Risiko lama', branch: 'Surabaya Branch', assessments: [] })

    const post = fetchMock.mock.calls.find(([, opts]: any) => opts?.method === 'POST')
    expect(post[1].body).toMatchObject({ branch: 'Surabaya Branch', location_id: 'loc-2' })
  })

  it('never sends a branch name that is not master data', async () => {
    const fetchMock = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )
    global.$fetch = fetchMock

    const store = useRiskProfileStore()
    await store.fetchBranches()

    expect(store.locationIdForBranch('Surabaya Branch')).toBe('loc-2')
    expect(store.locationIdForBranch('Nowhere Branch')).toBeUndefined()

    await store.addRisk({ name: 'Risiko tanpa cabang induk', branch: 'Nowhere Branch', assessments: [] })

    const post = fetchMock.mock.calls.find(([, opts]: any) => opts?.method === 'POST')
    expect(post[1].body.location_id).toBeUndefined()
    expect(post[1].body.branch).toBeUndefined()
  })

  it('PUTs the location picked in the edit modal and rolls back on failure', async () => {
    const year = new Date().getFullYear()
    let failPut = false
    const fetchMock = vi.fn(async (url: string, opts: any) => {
      if (opts?.method === 'PUT') {
        if (failPut) throw new Error('400 Bad Request')
        return okResponse({ id: 'r1' })
      }
      if (String(url).includes('/locations')) return okResponse(seededLocations)
      return okResponse([{ id: 'r1', name: 'R1', impact: 3, likelihood: 3, location_id: 'loc-1', branch: 'AIFL Headquarters', assessments: [{ year }] }])
    })
    global.$fetch = fetchMock

    const store = useRiskProfileStore()
    await store.fetchBranches()
    await store.fetchRisks()

    expect(await store.updateRisk({ ...store.rawRisks[0], location_id: 'loc-2' })).toBe(true)
    const put = fetchMock.mock.calls.find(([, opts]: any) => opts?.method === 'PUT')
    expect(put[1].body).toMatchObject({ location_id: 'loc-2', branch: 'Surabaya Branch' })
    expect(store.rawRisks[0].location_id).toBe('loc-2')

    failPut = true
    expect(await store.updateRisk({ ...store.rawRisks[0], location_id: 'loc-1' })).toBe(false)
    expect(store.rawRisks[0].location_id).toBe('loc-2')
  })
})

describe('Risk profile does not hardcode branches', () => {
  const STORE = resolve(__dirname, '../../stores/risk-profile.ts')
  const HEATMAP = resolve(__dirname, '../../components/risk-profile/RiskHeatMap.vue')

  it('has no fallback branch list or mock risk data', () => {
    const store = readFileSync(STORE, 'utf-8')
    expect(store).not.toContain('fallbackBranches')
    expect(store).not.toContain('initialRiskData')
    for (const name of ['Head Office', 'Bali Branch', 'Surabaya Branch', 'Bandung Branch', 'Jakarta Branch']) {
      expect(store, `store hardcodes "${name}"`).not.toContain(`'${name}'`)
    }
  })

  it('the heat map does not default new risks to a hardcoded branch', () => {
    expect(readFileSync(HEATMAP, 'utf-8')).not.toContain("'Head Office'")
  })
})

describe('Branch / Location navigation and i18n', () => {
  it('appears under Master Data in the sidebar', () => {
    const layout = readFileSync(LAYOUT, 'utf-8')
    const masterGroup = layout.slice(
      layout.indexOf("t('navigation.masterData')"),
      layout.indexOf("// // Settings")
    )

    expect(
      masterGroup.includes("t('navigation.location')"),
      'Master Data group has no Branch/Location entry'
    ).toBe(true)
    expect(
      masterGroup.includes("to: '/master/location'"),
      'Master Data group does not link to /master/location'
    ).toBe(true)
  })

  it('has a page at /master/location rendering the table and form', () => {
    const page = readFileSync(PAGE, 'utf-8')
    expect(page).toContain('MasterLocationTable')
    expect(page).toContain('MasterLocationForm')
  })

  it('defines every location key in both locales', () => {
    const en = JSON.parse(readFileSync(EN, 'utf-8'))
    const id = JSON.parse(readFileSync(ID, 'utf-8'))

    expect(en.navigation.location).toBeTruthy()
    expect(id.navigation.location).toBeTruthy()

    const enKeys = Object.keys(en.masterData.location).sort()
    const idKeys = Object.keys(id.masterData.location).sort()
    expect(idKeys).toEqual(enKeys)
    expect(Object.keys(id.masterData.location.columns).sort()).toEqual(
      Object.keys(en.masterData.location.columns).sort()
    )

    for (const key of ['fetchLocations', 'createLocation', 'updateLocation', 'deleteLocation']) {
      expect(en.masterData.errors[key], `en masterData.errors.${key}`).toBeTruthy()
      expect(id.masterData.errors[key], `id masterData.errors.${key}`).toBeTruthy()
    }
    for (const key of ['locationNameRequired', 'locationAddressRequired', 'locationCityRequired']) {
      expect(en.masterData.validation[key], `en masterData.validation.${key}`).toBeTruthy()
      expect(id.masterData.validation[key], `id masterData.validation.${key}`).toBeTruthy()
    }
  })
})
