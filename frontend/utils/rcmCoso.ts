// COSO 2013 mapping and master department on the Risk Control Matrix
// (docs/coso-2013-component-assessment-plan.md 9.3). Pure helpers, no store access.

export const RCM_FILTER_ALL = '__all'
export const RCM_FILTER_UNASSIGNED = 'unassigned'

export interface CosoPrinciple {
  code: string
  number: number
  component_code: string
  name_en: string
  is_active: boolean
}

export interface CosoComponent {
  code: string
  sort_order: number
  name_en: string
  principles: CosoPrinciple[]
}

export interface CosoMappingItem {
  department_id?: string | null
  coso_component_code?: string | null
  coso_principle_codes?: string[] | null
}

/** Field name -> message, from the backend 422 `error.fields`. */
export type RcmFieldErrors = Record<string, string>

/** Reads the per-field messages of a 422 (`{ error: { fields } }`); empty when there are none. */
export const parseRcmFieldErrors = (error: unknown): RcmFieldErrors => {
  const fields = (error as { data?: { error?: { fields?: unknown } } } | null)?.data?.error?.fields
  if (!fields || typeof fields !== 'object') return {}
  const out: RcmFieldErrors = {}
  for (const [key, value] of Object.entries(fields as Record<string, unknown>)) {
    if (typeof value === 'string' && value) out[key] = value
  }
  return out
}

export const parseCosoCatalogue = (catalogue: unknown): CosoComponent[] => {
  const root = (catalogue as { data?: { components?: unknown } } | null)?.data?.components
  if (!Array.isArray(root)) return []
  return root
    .filter((c): c is CosoComponent => !!c && typeof (c as CosoComponent).code === 'string')
    .map(c => ({
      code: c.code,
      sort_order: Number(c.sort_order) || 0,
      name_en: c.name_en || c.code,
      principles: Array.isArray(c.principles) ? c.principles : []
    }))
}

/** Component code of a principle code in the catalogue, or null. */
export const componentOfPrinciple = (catalogue: CosoComponent[], principleCode: string): string | null =>
  catalogue.find(c => c.principles.some(p => p.code === principleCode))?.code ?? null

/**
 * Principles the user may tag: active ones, or inactive ones the row already has
 * (the backend lets an existing tag stay). Optionally limited to one component.
 */
export const selectablePrinciples = (
  catalogue: CosoComponent[],
  opts: { componentCode?: string | null, keep?: string[] } = {}
): CosoPrinciple[] => {
  const keep = new Set(opts.keep ?? [])
  return catalogue
    .filter(c => !opts.componentCode || c.code === opts.componentCode)
    .flatMap(c => c.principles)
    .filter(p => p.is_active || keep.has(p.code))
    .sort((a, b) => a.number - b.number)
}

/** Distinct components of the given principle codes, in catalogue order. */
export const componentsOfPrinciples = (catalogue: CosoComponent[], principleCodes: string[]): string[] => {
  const wanted = new Set(principleCodes)
  return catalogue
    .filter(c => c.principles.some(p => wanted.has(p.code)))
    .map(c => c.code)
}

/**
 * Same rule the backend enforces: the primary component must be the component of at
 * least one tagged principle. Returns null when valid (including no component at all).
 */
export const validateCosoMapping = (
  catalogue: CosoComponent[],
  componentCode: string | null | undefined,
  principleCodes: string[]
): 'componentWithoutPrinciples' | 'componentMismatch' | null => {
  if (!componentCode) return null
  if (!principleCodes.length) return 'componentWithoutPrinciples'
  // Unknown catalogue (failed to load): leave the check to the backend.
  if (!catalogue.length) return null
  return componentsOfPrinciples(catalogue, principleCodes).includes(componentCode) ? null : 'componentMismatch'
}

/** Filter by master department: all, unassigned (no department_id) or one department id. */
export const matchesRcmDepartment = (item: CosoMappingItem, selected: string): boolean => {
  if (!selected || selected === RCM_FILTER_ALL) return true
  if (selected === RCM_FILTER_UNASSIGNED) return !item.department_id
  return item.department_id === selected
}

/** Filter by COSO component (primary component or the component of any tagged principle) and principle. */
export const matchesRcmCoso = (
  item: CosoMappingItem,
  catalogue: CosoComponent[],
  component: string,
  principle: string
): boolean => {
  const codes = item.coso_principle_codes ?? []
  if (component && component !== RCM_FILTER_ALL) {
    const viaPrinciple = componentsOfPrinciples(catalogue, codes).includes(component)
    if (item.coso_component_code !== component && !viaPrinciple) return false
  }
  if (principle && principle !== RCM_FILTER_ALL && !codes.includes(principle)) return false
  return true
}
