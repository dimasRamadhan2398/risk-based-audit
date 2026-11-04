import { describe, it, expect } from 'vitest'
import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'
import {
  RCM_ALL_BRANCHES,
  RCM_UNASSIGNED_BRANCH,
  resolveRcmItemLocationId,
  matchesBranchFilter,
  filterRcmItemsByBranch,
  buildRcmBranchOptions,
  type RcmBranchRisk
} from '~/utils/rcmBranch'

const locations = [
  { id: 'loc-jkt', name: 'Jakarta Branch' },
  { id: 'loc-sby', name: 'Surabaya Branch' }
]

const risks: RcmBranchRisk[] = [
  { id: 'r-jkt', location_id: 'loc-jkt' },
  { id: 'r-none', location_id: null, branch: null },
  { id: 'r-legacy', branch: 'Surabaya Branch' },
  { id: 'r-moved', location_id: 'loc-jkt', assessments: [{ year: 2025, location_id: 'loc-sby' }, { year: 2026 }] }
]

const findRisk = (riskId: string) => risks.find(r => String(r.id) === riskId)
// Mirrors riskProfileStore.riskLocationId: location_id, else match a branch name to the master.
const riskLocationId = (risk: RcmBranchRisk) =>
  risk.location_id ? String(risk.location_id) : (locations.find(l => l.name === risk.branch)?.id ?? null)

const locate = (item: { risk_id?: string | null, year?: number }) =>
  resolveRcmItemLocationId(item, findRisk, riskLocationId)

describe('resolveRcmItemLocationId', () => {
  it('linked risk with a location -> that location', () => {
    expect(locate({ risk_id: 'r-jkt', year: 2026 })).toBe('loc-jkt')
  })

  it('legacy risk with only a branch name -> matched master location', () => {
    expect(locate({ risk_id: 'r-legacy' })).toBe('loc-sby')
  })

  it('no risk_id -> unassigned', () => {
    expect(locate({ risk_id: '' })).toBeNull()
    expect(locate({})).toBeNull()
  })

  it('risk without a location -> unassigned', () => {
    expect(locate({ risk_id: 'r-none' })).toBeNull()
  })

  it('risk not found -> unassigned', () => {
    expect(locate({ risk_id: 'r-deleted' })).toBeNull()
  })

  it('uses the risk\'s location for the item\'s year when that assessment has one', () => {
    expect(locate({ risk_id: 'r-moved', year: 2025 })).toBe('loc-sby')
    expect(locate({ risk_id: 'r-moved', year: 2026 })).toBe('loc-jkt')
  })
})

describe('branch filtering', () => {
  const items = [
    { id: 'a', risk_id: 'r-jkt', department: 'Finance' },
    { id: 'b', risk_id: 'r-legacy', department: 'Finance' },
    { id: 'c', risk_id: 'r-none', department: 'IT' },
    { id: 'd', risk_id: '', department: 'IT' },
    { id: 'e', risk_id: 'r-deleted', department: 'IT' }
  ]
  const ids = (selected: string) => filterRcmItemsByBranch(items, selected, locate).map(i => i.id)

  it('"All branches" keeps everything', () => {
    expect(ids(RCM_ALL_BRANCHES)).toEqual(['a', 'b', 'c', 'd', 'e'])
  })

  it('a branch keeps only items whose linked risk is there', () => {
    expect(ids('loc-jkt')).toEqual(['a'])
    expect(ids('loc-sby')).toEqual(['b'])
  })

  it('"Unassigned" keeps items with no risk, a missing risk, or a risk without location', () => {
    expect(ids(RCM_UNASSIGNED_BRANCH)).toEqual(['c', 'd', 'e'])
  })

  it('combines with a department filter applied first', () => {
    const it = items.filter(i => i.department === 'IT')
    expect(filterRcmItemsByBranch(it, RCM_UNASSIGNED_BRANCH, locate).map(i => i.id)).toEqual(['c', 'd', 'e'])
    expect(filterRcmItemsByBranch(it, 'loc-jkt', locate)).toEqual([])
  })

  it('matchesBranchFilter handles the sentinels', () => {
    expect(matchesBranchFilter(null, RCM_ALL_BRANCHES)).toBe(true)
    expect(matchesBranchFilter(null, RCM_UNASSIGNED_BRANCH)).toBe(true)
    expect(matchesBranchFilter('loc-jkt', RCM_UNASSIGNED_BRANCH)).toBe(false)
    expect(matchesBranchFilter('loc-jkt', 'loc-sby')).toBe(false)
  })
})

describe('buildRcmBranchOptions', () => {
  const labels = { all: 'All Branches', unassigned: 'Unassigned' }

  it('All first, then Location master rows by id, Unassigned only when needed', () => {
    expect(buildRcmBranchOptions(locations, false, labels)).toEqual([
      { label: 'All Branches', value: RCM_ALL_BRANCHES },
      { label: 'Jakarta Branch', value: 'loc-jkt' },
      { label: 'Surabaya Branch', value: 'loc-sby' }
    ])
    expect(buildRcmBranchOptions(locations, true, labels).at(-1)).toEqual({ label: 'Unassigned', value: RCM_UNASSIGNED_BRANCH })
  })
})

type Locale = { rcm: { filters: Record<string, string> }, riskProfile: { unassignedBranch: string } }
const enL = en as unknown as Locale
const idL = id as unknown as Locale

describe('RCM filter i18n', () => {
  it.each(['allDepartments', 'allBranches', 'unassignedBranch', 'branchLabel', 'branchesUnavailable'])('%s exists in en and id', (key) => {
    expect(enL.rcm.filters[key]).toBeTruthy()
    expect(idL.rcm.filters[key]).toBeTruthy()
  })

  it('Unassigned matches the Corporate Risk Profile label', () => {
    expect(enL.rcm.filters.unassignedBranch).toBe(enL.riskProfile.unassignedBranch)
    expect(idL.rcm.filters.unassignedBranch).toBe(idL.riskProfile.unassignedBranch)
  })
})
