export interface AuditableEntityOption {
  /** corporate audit universe id; this is what the activity stores as auditUniverseId */
  id?: string
  name: string
  label: string
  value: string
  riskLevel: string
  riskIndex: number
  auditPriority: boolean
  source: 'yearly' | 'corporate'
}

export interface YearlyUniverseEntry {
  id?: string
  corporate_audit_universe_id?: string
  corporate_audit_universe?: { id?: string, name?: string } | null
  risk_level?: string
  risk_index?: number | string
  audit_priority?: boolean
}

export interface CorporateUniverseNode {
  id?: string
  name?: string
  children?: unknown[]
}

/**
 * Builds the "Auditable Entity" dropdown options from the real Audit Universe data.
 * 1. Entities scored for the selected plan year (yearly universe) come first.
 * 2. Then every corporate universe node, parents and sub-entities, not already listed.
 * Names are de-duplicated case-insensitively.
 */
export const buildAuditableEntityOptions = (
  yearly: YearlyUniverseEntry[] | null | undefined,
  corporate: CorporateUniverseNode[] | null | undefined
): AuditableEntityOption[] => {
  const items: AuditableEntityOption[] = []
  const seenNames = new Set<string>()

  for (const item of yearly || []) {
    const name = item.corporate_audit_universe?.name
    if (!name || seenNames.has(name.toLowerCase())) continue
    seenNames.add(name.toLowerCase())

    // The entity identity is the corporate node, not the per-year row.
    const id = item.corporate_audit_universe_id || item.corporate_audit_universe?.id
    items.push({
      id,
      name,
      label: name,
      value: name,
      riskLevel: item.risk_level || 'Not Scored',
      riskIndex: Number(item.risk_index) || 0,
      auditPriority: !!item.audit_priority,
      source: 'yearly'
    })
  }

  const traverse = (node: CorporateUniverseNode, parentName?: string) => {
    if (!node || !node.name) return
    const lowerName = node.name.toLowerCase()
    if (!seenNames.has(lowerName)) {
      seenNames.add(lowerName)
      items.push({
        id: node.id,
        name: node.name,
        label: parentName ? `${node.name} (${parentName})` : node.name,
        value: node.name,
        riskLevel: 'Not Scored',
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

  return items
}
