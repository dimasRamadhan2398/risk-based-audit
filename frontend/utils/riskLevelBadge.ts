// Single source of truth for risk level and control effectiveness badge colours across the app.
// Canonical keys are the backend `risk_level` values; other spellings (RiskLevel enum, "Moderate to High",
// AI labels, ...) are normalised onto them. Class names are written out in full so Tailwind detects them.
// Light mode uses a lighter (-400) shade with black text, dark mode a deeper (-700) shade of the same hue with white text.

export type PriorityRiskLevel = 'High' | 'Medium to High' | 'Medium' | 'Low to Medium' | 'Low'

export const PRIORITY_RISK_LEVELS: PriorityRiskLevel[] = ['High', 'Medium to High', 'Medium', 'Low to Medium', 'Low']

export const RISK_LEVEL_TEXT_CLASS = 'text-black dark:text-white'

export const RISK_LEVEL_BG_CLASSES: Record<PriorityRiskLevel, string> = {
  'High': 'bg-red-400 dark:bg-red-700',
  'Medium to High': 'bg-orange-400 dark:bg-orange-700',
  'Medium': 'bg-yellow-400 dark:bg-yellow-700',
  'Low to Medium': 'bg-lime-400 dark:bg-lime-700',
  'Low': 'bg-green-400 dark:bg-green-700'
}

export const RISK_LEVEL_FALLBACK_BG_CLASS = 'bg-slate-200 dark:bg-slate-700'

// i18n keys under riskFactors.priority.levels.*
export const RISK_LEVEL_LABEL_KEYS: Record<PriorityRiskLevel, string> = {
  'High': 'riskFactors.priority.levels.high',
  'Medium to High': 'riskFactors.priority.levels.moderateToHigh',
  'Medium': 'riskFactors.priority.levels.moderate',
  'Low to Medium': 'riskFactors.priority.levels.lowToModerate',
  'Low': 'riskFactors.priority.levels.low'
}

// ─── Risk index thresholds ─────────────────────────────────────────
// Mirrors the risk-service scoring (audit_universe_controller.go): the level is chosen from the risk index
// and Medium to High / High are audit priorities. Used by the scoring preview and the reference table.
export interface RiskLevelThreshold {
  level: PriorityRiskLevel
  /** inclusive lower bound of the risk index (%) */
  min: number
  /** upper bound shown in the reference table (%) */
  max: number
  prioritized: boolean
}

export const RISK_LEVEL_THRESHOLDS: RiskLevelThreshold[] = [
  { level: 'High', min: 80, max: 100, prioritized: true },
  { level: 'Medium to High', min: 60, max: 79, prioritized: true },
  { level: 'Medium', min: 40, max: 59, prioritized: false },
  { level: 'Low to Medium', min: 20, max: 39, prioritized: false },
  { level: 'Low', min: 0, max: 19, prioritized: false }
]

/** Risk level for a risk index (%), using the same thresholds as the backend. */
export const getRiskLevelForIndex = (riskIndex: number): PriorityRiskLevel => {
  const value = Number.isFinite(riskIndex) ? riskIndex : 0
  return (RISK_LEVEL_THRESHOLDS.find(th => value >= th.min) ?? RISK_LEVEL_THRESHOLDS[RISK_LEVEL_THRESHOLDS.length - 1]!).level
}

/** Whether a risk level makes an entity an audit priority (Medium to High or High). */
export const isPrioritizedRiskLevel = (level?: string | null): boolean => {
  const key = normalizeRiskLevel(level)
  return !!key && !!RISK_LEVEL_THRESHOLDS.find(th => th.level === key)?.prioritized
}

/** "80 - 100%" style label for the reference table. */
export const formatRiskLevelRange = (threshold: RiskLevelThreshold): string => `${threshold.min} - ${threshold.max}%`

const RISK_LEVEL_BADGE_BASE_CLASS ='inline-flex items-center justify-center font-bold px-3 py-1 rounded text-xs w-36 shadow-sm'

export const isPriorityRiskLevel = (level?: string | null): level is PriorityRiskLevel =>
  !!level && Object.prototype.hasOwnProperty.call(RISK_LEVEL_BG_CLASSES, level)

/**
 * Maps any risk level spelling onto the canonical level, or null when it isn't a risk level
 * (e.g. "Not Scored", "Custom"). Combined levels are checked before single words so
 * "Moderate to High" never collapses into "High" or "Moderate".
 */
export const normalizeRiskLevel = (level?: string | null): PriorityRiskLevel | null => {
  if (!level) return null
  if (isPriorityRiskLevel(level)) return level
  const raw = String(level).trim().toLowerCase().replace(/[\s\-_]+/g, '')
  if (!raw) return null
  if (/^(moderate|medium)(to)?high$/.test(raw)) return 'Medium to High'
  if (/^low(to)?(moderate|medium)$/.test(raw)) return 'Low to Medium'
  if (['high', 'veryhigh', 'critical', 'extreme'].includes(raw)) return 'High'
  if (['moderate', 'medium', 'watch'].includes(raw)) return 'Medium'
  if (['low', 'verylow'].includes(raw)) return 'Low'
  return null
}

/** Background + text colour classes for a risk level; unknown/empty levels get a neutral fallback. */
export const getRiskLevelColorClass = (level?: string | null): string => {
  const key = normalizeRiskLevel(level)
  const bg = key ? RISK_LEVEL_BG_CLASSES[key] : RISK_LEVEL_FALLBACK_BG_CLASS
  return `${bg} ${RISK_LEVEL_TEXT_CLASS}`
}

/** Background-only classes, for colour dots/swatches next to a risk level label. */
export const getRiskLevelDotClass = (level?: string | null): string => {
  const key = normalizeRiskLevel(level)
  return key ? RISK_LEVEL_BG_CLASSES[key] : RISK_LEVEL_FALLBACK_BG_CLASS
}

/** Full badge classes (shape + colour) used by the Audit Priority page. */
export const getRiskLevelBadgeClass = (level?: string | null): string =>
  `${RISK_LEVEL_BADGE_BASE_CLASS} ${getRiskLevelColorClass(level)}`

// ─── Control effectiveness ─────────────────────────────────────────
// Reuses the risk palette in reverse: the more effective the control, the lower the residual risk colour.
export type ControlEffectivenessRating
  = 'Highly Effective' | 'Effective' | 'Moderately Effective' | 'Partially Effective' | 'Weak' | 'Ineffective'

export const CONTROL_EFFECTIVENESS_RISK_EQUIVALENT: Record<ControlEffectivenessRating, PriorityRiskLevel> = {
  'Highly Effective': 'Low',
  'Effective': 'Low to Medium',
  'Moderately Effective': 'Medium',
  'Partially Effective': 'Medium',
  'Weak': 'Medium to High',
  'Ineffective': 'High'
}

const normalizeControlEffectiveness = (rating?: string | null): ControlEffectivenessRating | null => {
  if (!rating) return null
  const raw = String(rating).trim().toLowerCase().replace(/[\s\-_]+/g, '')
  const match = (Object.keys(CONTROL_EFFECTIVENESS_RISK_EQUIVALENT) as ControlEffectivenessRating[])
    .find(key => key.toLowerCase().replace(/\s+/g, '') === raw)
  return match || null
}

/** Background + text colour classes for a control effectiveness rating or test result; "Not Tested"/unknown get the neutral fallback. */
export const getControlEffectivenessColorClass = (rating?: string | null): string => {
  const key = normalizeControlEffectiveness(rating)
  return getRiskLevelColorClass(key ? CONTROL_EFFECTIVENESS_RISK_EQUIVALENT[key] : null)
}
