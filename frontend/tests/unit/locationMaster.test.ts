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
import { useRiskProfileStore, fallbackBranches } from '~/stores/risk-profile'

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

  it('takes its branch options from the Location master data', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()

    // Active master branches first, then any branch already used by a risk.
    expect(store.branches).toContain('AIFL Headquarters')
    expect(store.branches).toContain('Surabaya Branch')
    // Inactive and unused by any risk → not offered.
    expect(store.branches).not.toContain('Medan Branch (Closed)')
  })

  it('keeps branches that existing risks already reference', async () => {
    global.$fetch = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse([seededLocations[1]]) : okResponse([])
    )

    const store = useRiskProfileStore()
    await store.fetchBranches()
    await vi.waitFor(() => expect(store.rawRisks.length).toBeGreaterThan(0))

    // Seed risks live in 'Head Office', which is not in the master list above —
    // it must still be selectable so those records stay editable.
    expect(store.branches).toContain('Surabaya Branch')
    expect(store.branches).toContain('Head Office')
  })

  it('falls back to the seed list when no master data and no risks are available', async () => {
    global.$fetch = vi.fn(async () => {
      throw new Error('offline')
    })

    const store = useRiskProfileStore()
    await store.fetchBranches()
    store.rawRisks = []

    expect(store.branches.length).toBeGreaterThan(0)
  })
})

describe('Risks carry a location reference', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('sends location_id alongside the branch name when saving a risk', async () => {
    const fetchMock = vi.fn(async (url: string) =>
      String(url).includes('/locations') ? okResponse(seededLocations) : okResponse([])
    )
    global.$fetch = fetchMock

    const store = useRiskProfileStore()
    await store.fetchBranches()

    await store.addRisk({
      name: 'Kegagalan rekonsiliasi kas cabang',
      category: 'Operations',
      branch: 'Surabaya Branch',
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

  it('omits location_id for a branch that is not master data', async () => {
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
    expect(post[1].body.branch).toBe('Nowhere Branch')
  })
})

describe('Branch names match the backend location seeds', () => {
  const SEEDER = resolve(
    __dirname,
    '../../../backend/master-service/pkg/database/seeders/seeder.go'
  )

  it('every CRP fallback branch exists in LocationSeeds', () => {
    const seeder = readFileSync(SEEDER, 'utf-8')
    const seedBlock = seeder.slice(
      seeder.indexOf('var LocationSeeds = []models.Location{'),
      seeder.indexOf('var renamedLocations')
    )
    const seededNames = Array.from(seedBlock.matchAll(/Name:\s*"([^"]+)"/g)).map((m) => m[1])

    expect(seededNames.length).toBeGreaterThan(0)
    for (const branch of fallbackBranches) {
      expect(
        seededNames.includes(branch),
        `branch "${branch}" is offered by the risk profile but not seeded as a location`
      ).toBe(true)
    }
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
