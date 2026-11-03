/**
 * Expanded state of the significant-finding cards (components/audit-result-report/SignificantFindingCards.vue).
 * Cards are keyed by finding object identity, not by index, so removing a card never moves the open state to another card.
 * Nothing is expanded on load: every existing card starts collapsed. Only expand() opens a card programmatically.
 */
export interface FindingExpansion {
  isExpanded: (finding: object) => boolean
  setExpanded: (finding: object, open: boolean) => void
  expand: (finding: object | undefined) => void
  forget: (finding: object) => void
}

export const createFindingExpansion = (expanded: Set<object> = new Set<object>()): FindingExpansion => ({
  isExpanded: finding => expanded.has(finding),
  setExpanded: (finding, open) => {
    if (open) expanded.add(finding)
    else expanded.delete(finding)
  },
  expand: (finding) => {
    if (finding) expanded.add(finding)
  },
  forget: (finding) => {
    expanded.delete(finding)
  }
})
