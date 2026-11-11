// Branch filter for the Risk Control Matrix. RCM items have no branch of their
// own: an item's branch is the Location master branch of its linked Corporate
// Risk Profile risk. Items with no linked risk, a risk that cannot be found, or
// a risk without a location fall under "Unassigned" (same as the CRP filter).

// Same sentinel values as ALL_BRANCHES / UNASSIGNED_BRANCH in stores/risk-profile.ts.
export const RCM_ALL_BRANCHES = 'All Branches'
export const RCM_UNASSIGNED_BRANCH = '__unassigned__'

export interface RcmBranchRisk {
  id: string | number
  location_id?: string | null
  branch?: string | null
  assessments?: { year?: number, location_id?: string | null }[]
}

export interface RcmBranchItem {
  risk_id?: string | null
  year?: number
}

/**
 * Location master ID an RCM item belongs to, or null when unassigned.
 * The risk's assessment for the item's year wins (the CRP reads a risk as it
 * stood that year); otherwise the risk's own location (via riskLocationId,
 * which also matches legacy rows that only carry a branch name).
 */
export const resolveRcmItemLocationId = (
  item: RcmBranchItem,
  findRisk: (riskId: string) => RcmBranchRisk | undefined,
  riskLocationId: (risk: RcmBranchRisk) => string | null
): string | null => {
  if (!item.risk_id) return null
  const risk = findRisk(String(item.risk_id))
  if (!risk) return null
  const yearLocation = item.year
    ? risk.assessments?.find(a => a.year === item.year)?.location_id
    : undefined
  if (yearLocation) return String(yearLocation)
  return riskLocationId(risk) || null
}

/** Whether a resolved location ID passes the selected branch filter value. */
export const matchesBranchFilter = (locationId: string | null, selected: string): boolean => {
  if (!selected || selected === RCM_ALL_BRANCHES) return true
  if (selected === RCM_UNASSIGNED_BRANCH) return !locationId
  return locationId === selected
}

export const filterRcmItemsByBranch = <T extends RcmBranchItem>(
  items: T[],
  selected: string,
  locationOf: (item: T) => string | null
): T[] => {
  if (!selected || selected === RCM_ALL_BRANCHES) return items
  return items.filter(item => matchesBranchFilter(locationOf(item), selected))
}

/** "All", the active Location master rows, and "Unassigned" only when something is unassigned. */
export const buildRcmBranchOptions = (
  locations: { id: string, name: string }[],
  hasUnassigned: boolean,
  labels: { all: string, unassigned: string }
) => [
  { label: labels.all, value: RCM_ALL_BRANCHES },
  ...locations.map(l => ({ label: l.name, value: l.id })),
  ...(hasUnassigned ? [{ label: labels.unassigned, value: RCM_UNASSIGNED_BRANCH }] : [])
]
