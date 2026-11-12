/**
 * Display helpers for the internal control assessment pages: labels, icons and source links.
 * Status is always shown as icon + text, never by colour alone (plan 9.2).
 */
import type { ComponentStatus, EvidenceState, IcEvidence, PrincipleStatus } from '~/utils/icAssessment'

type T = (key: string, params?: Record<string, string | number>) => string

const translated = (t: T, key: string, params?: Record<string, string | number>): string | null => {
  const v = t(key, params)
  return v && v !== key ? v : null
}

export const periodLabel = (t: T, year: number, quarter: number): string =>
  quarter > 0 ? t('icAssessment.period.quarter', { quarter, year }) : t('icAssessment.period.annual', { year })

/** `ce.integrity_pact` -> icAssessment.signals.ce_integrity_pact; unknown keys still render. */
export const signalLabel = (t: T, key: string): string =>
  translated(t, `icAssessment.signals.${key.replace(/\./g, '_')}`) ?? t('icAssessment.signals.unknown', { key })

export const componentLabel = (t: T, code: string): string => translated(t, `coso.components.${code}`) ?? code

export const principleTitle = (t: T, code: string): string => translated(t, `coso.principles.${code}.title`) ?? code

export const ratingLabel = (t: T, rating: number | null): string =>
  rating === null ? t('icAssessment.grid.noRating') : (translated(t, `icAssessment.rating.${rating}`) ?? String(rating))

export interface StatusVisual {
  icon: string
  /** Text colour classes, light and dark. The text label carries the meaning. */
  cls: string
}

export const COMPONENT_STATUS_VISUAL: Record<ComponentStatus | 'NOT_ASSESSED', StatusVisual> = {
  YES: { icon: 'i-lucide-circle-check', cls: 'text-emerald-700 dark:text-emerald-400' },
  PARTIAL: { icon: 'i-lucide-circle-alert', cls: 'text-amber-700 dark:text-amber-400' },
  NO: { icon: 'i-lucide-circle-x', cls: 'text-red-700 dark:text-red-400' },
  NOT_ASSESSED: { icon: 'i-lucide-circle-dashed', cls: 'text-gray-500 dark:text-gray-400' }
}

export const principleStatusVisual = (s: PrincipleStatus): StatusVisual =>
  s === 'NOT_APPLICABLE'
    ? { icon: 'i-lucide-circle-slash', cls: 'text-gray-600 dark:text-gray-300' }
    : COMPONENT_STATUS_VISUAL[s]

export const EVIDENCE_STATE_VISUAL: Record<EvidenceState, StatusVisual> = {
  OK: { icon: 'i-lucide-database', cls: 'text-emerald-700 dark:text-emerald-400' },
  NO_DATA: { icon: 'i-lucide-database-zap', cls: 'text-gray-600 dark:text-gray-300' },
  NOT_APPLICABLE: { icon: 'i-lucide-circle-slash', cls: 'text-gray-600 dark:text-gray-300' },
  SOURCE_UNAVAILABLE: { icon: 'i-lucide-plug-zap', cls: 'text-red-700 dark:text-red-400' }
}

export const ASSESSMENT_STATUS_COLOR: Record<string, 'neutral' | 'info' | 'success' | 'warning'> = {
  DRAFT: 'neutral',
  SUBMITTED: 'info',
  APPROVED: 'success',
  REOPENED: 'warning'
}

export const CONCLUSION_COLOR: Record<string, 'neutral' | 'success' | 'error' | 'warning'> = {
  EFFECTIVE: 'success',
  NOT_EFFECTIVE: 'error',
  REQUIRES_JUDGMENT: 'warning',
  INCOMPLETE: 'neutral'
}

/** Page where the source records of a signal can be looked at (plan 9.2 evidence drawer). */
export const evidenceSourceLink = (e: Pick<IcEvidence, 'signalKey' | 'detail'>, departmentId: string): { path: string, query?: Record<string, string> } | null => {
  const key = e.signalKey
  if (key.startsWith('ca.')) return { path: '/risk-profile/risk-control-matrix', query: departmentId ? { department_id: departmentId } : undefined }
  if (key.startsWith('ra.')) return { path: '/risk-profile' }
  if (key === 'ic.lha_timeliness') return { path: '/audit-result-report' }
  if (key === 'ic.management_response' || key.startsWith('ma.')) return { path: '/action-taken-report' }
  if (key === 'ce.audit_charter') return { path: '/audit-charter' }
  if (key === 'ce.unit_accountability' || key === 'ce.integrity_pact') return { path: '/master/department' }
  return null
}
