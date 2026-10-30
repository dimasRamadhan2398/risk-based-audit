import { isUniverseYearScored } from './riskScoring'

export interface AuditableEntityOption {
  /** corporate audit universe id; this is what the activity stores as auditUniverseId */
  id?: string
  name: string
  label: string
  /** corporate audit universe id (falls back to the name only for nodes without an id) */
  value: string
  riskLevel: string
  riskIndex: number
  auditPriority: boolean
  source: 'yearly' | 'corporate'
}

export interface YearlyUniverseEntry {
  id?: string
  corporate_audit_universe_id?: string
  corporate_audit_universe?: { id?: string, name?: string, parent_id?: string | null } | null
  risk_level?: string
  risk_index?: number | string
  audit_priority?: boolean
  /** null = never scored; missing on older backends (treated as scored) */
  scored_at?: string | null
}

export interface CorporateUniverseNode {
  id?: string
  name?: string
  parent_id?: string | null
  children?: unknown[]
}

export const NOT_SCORED_RISK_LEVEL = 'Not Scored'

/**
 * Builds the "Auditable Entity" dropdown options from the real Audit Universe data.
 * 1. Entities established for the selected plan year (yearly universe) come first.
 * 2. Then every corporate universe node, parents and sub-entities, not already listed.
 * Entities are identified by their corporate node id, so two entities with the same name stay
 * separate; their labels are disambiguated with the parent name when it is known.
 * Unscored entities are never presented as prioritized.
 */
export const buildAuditableEntityOptions = (
  yearly: YearlyUniverseEntry[] | null | undefined,
  corporate: CorporateUniverseNode[] | null | undefined
): AuditableEntityOption[] => {
  const items: AuditableEntityOption[] = []
  const seenKeys = new Set<string>()
  // Nodes without an id (should not happen) fall back to a name key so they are still listed once.
  const keyOf = (id: string | undefined, name: string) => id ? `id:${id}` : `name:${name.toLowerCase()}`

  // Parent names from the corporate tree, to label sub-entities and duplicate names.
  const parentNameById = new Map<string, string>()
  const nameById = new Map<string, string>()
  const collect = (node: CorporateUniverseNode, parentName?: string) => {
    if (!node) return
    if (node.id && node.name) {
      nameById.set(node.id, node.name)
      if (parentName) parentNameById.set(node.id, parentName)
    }
    if (Array.isArray(node.children)) {
      for (const child of node.children) collect(child as CorporateUniverseNode, node.name)
    }
  }
  for (const node of corporate || []) collect(node)

  const parentNameOf = (id?: string, parentId?: string | null) =>
    (id && parentNameById.get(id)) || (parentId ? nameById.get(parentId) : undefined)

  for (const item of yearly || []) {
    const name = item.corporate_audit_universe?.name
    if (!name) continue
    // The entity identity is the corporate node, not the per-year row.
    const id = item.corporate_audit_universe_id || item.corporate_audit_universe?.id
    const key = keyOf(id, name)
    if (seenKeys.has(key)) continue
    seenKeys.add(key)

    const scored = isUniverseYearScored(item)
    items.push({
      id,
      name,
      label: name,
      value: id || name,
      riskLevel: (scored && item.risk_level) || NOT_SCORED_RISK_LEVEL,
      riskIndex: scored ? Number(item.risk_index) || 0 : 0,
      auditPriority: scored && !!item.audit_priority,
      source: 'yearly'
    })
    const parentName = parentNameOf(id, item.corporate_audit_universe?.parent_id)
    if (id && parentName) parentNameById.set(id, parentName)
  }

  const traverse = (node: CorporateUniverseNode, parentName?: string) => {
    if (!node || !node.name) return
    const key = keyOf(node.id, node.name)
    if (!seenKeys.has(key)) {
      seenKeys.add(key)
      items.push({
        id: node.id,
        name: node.name,
        label: parentName ? `${node.name} (${parentName})` : node.name,
        value: node.id || node.name,
        riskLevel: NOT_SCORED_RISK_LEVEL,
        riskIndex: 0,
        auditPriority: false,
        source: 'corporate'
      })
    }
    if (Array.isArray(node.children)) {
      for (const child of node.children) {
        traverse(child as CorporateUniverseNode, node.name)
      }
    }
  }

  for (const node of corporate || []) {
    traverse(node)
  }

  // Same name twice: add the parent name to yearly labels that do not have it yet.
  const nameCounts = new Map<string, number>()
  for (const item of items) {
    const lower = item.name.toLowerCase()
    nameCounts.set(lower, (nameCounts.get(lower) || 0) + 1)
  }
  for (const item of items) {
    if ((nameCounts.get(item.name.toLowerCase()) || 0) < 2 || item.label !== item.name) continue
    const parentName = item.id ? parentNameById.get(item.id) : undefined
    if (parentName) item.label = `${item.name} (${parentName})`
  }

  return items
}

/**
 * Corporate audit universe id (lower-case) -> entity name, over the whole tree (parents and
 * sub-entities). The PKAT row stores this id as auditUniverseId.
 */
export const buildCorporateUniverseNameMap = (
  corporate: CorporateUniverseNode[] | null | undefined
): Map<string, string> => {
  const names = new Map<string, string>()
  const collect = (node: CorporateUniverseNode) => {
    if (!node) return
    if (node.id && node.name) names.set(node.id.trim().toLowerCase(), node.name)
    if (Array.isArray(node.children)) {
      for (const child of node.children) collect(child as CorporateUniverseNode)
    }
  }
  for (const node of corporate || []) collect(node)
  return names
}

/**
 * Resolves what the picker emitted (an option object or its value) to an option, by corporate id.
 * Never matches by name, so entities sharing a name cannot be confused.
 */
export const resolveAuditableEntitySelection = (
  options: AuditableEntityOption[],
  selected: unknown
): AuditableEntityOption | null => {
  if (selected === null || selected === undefined || selected === '') return null
  if (typeof selected === 'object') {
    const opt = selected as Partial<AuditableEntityOption>
    const byId = opt.id ? options.find(o => o.id === opt.id) : undefined
    if (byId) return byId
    const byValue = opt.value ? options.find(o => o.value === opt.value) : undefined
    if (byValue) return byValue
    return opt.name ? (opt as AuditableEntityOption) : null
  }
  const value = String(selected)
  return options.find(o => o.value === value) || options.find(o => o.id === value) || null
}

export interface SavedAuditableEntityRef {
  auditUniverseId?: string | null
  riskName?: string | null
  riskLevel?: string | null
}

/**
 * The option to pre-select for a saved PKAT row: by auditUniverseId; by name only when the row has no id
 * (and the name is unambiguous). A row whose entity is not in the loaded lists gets a synthetic option.
 */
export const findAuditableEntityForActivity = (
  options: AuditableEntityOption[],
  activity: SavedAuditableEntityRef
): AuditableEntityOption | undefined => {
  if (activity.auditUniverseId) {
    const byId = options.find(o => o.id && o.id === activity.auditUniverseId)
    if (byId) return byId
  } else if (activity.riskName) {
    const lower = activity.riskName.toLowerCase()
    const byName = options.filter(o => o.name.toLowerCase() === lower)
    if (byName.length === 1) return byName[0]
  }
  if (!activity.riskName) return undefined
  return {
    id: activity.auditUniverseId || undefined,
    name: activity.riskName,
    label: activity.riskName,
    value: activity.auditUniverseId || activity.riskName,
    riskLevel: activity.riskLevel || NOT_SCORED_RISK_LEVEL,
    riskIndex: 0,
    auditPriority: false,
    source: 'corporate'
  }
}
