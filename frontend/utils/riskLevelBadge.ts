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
  { level: 'Medium', min: 40, max: 59, prioritized: true },
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

const RISK_LEVEL_BADGE_BASE_CLASS = 'inline-flex items-center justify-center font-bold px-3 py-1 rounded text-xs w-36 shadow-sm'

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

// ─── Anomaly Detection 5-Level Normalization ────────────────────────
// Anomaly Detection strictly uses ONLY the 5 levels requested across trained, predicted,
// and external database data (no "Critical" or unmapped custom external risk labels).
export type AnomalyRiskLevelLabel = 'High' | 'Moderate to High' | 'Moderate' | 'Low to Moderate' | 'Low'

export const ANOMALY_RISK_LEVELS: AnomalyRiskLevelLabel[] = [
  'High',
  'Moderate to High',
  'Moderate',
  'Low to Moderate',
  'Low'
]

export const ANOMALY_RISK_LABEL_TO_CODE: Record<AnomalyRiskLevelLabel, string> = {
  'High': 'HIGH',
  'Moderate to High': 'MODERATE_HIGH',
  'Moderate': 'MODERATE',
  'Low to Moderate': 'LOW_MODERATE',
  'Low': 'LOW'
}

/**
 * Strictly maps ANY anomaly risk level / severity value (including legacy "Critical",
 * external DB custom ratings such as "Extreme"/"Very High"/"Major"/"Minor"/"KOL-3",
 * Indonesian ratings, or numeric scores) onto ONLY:
 *   "High" | "Moderate to High" | "Moderate" | "Low to Moderate" | "Low"
 */
export const normalizeAnomalyRiskLevel = (
  level?: unknown,
  fallback: AnomalyRiskLevelLabel = 'Moderate'
): AnomalyRiskLevelLabel => {
  if (typeof level === 'number' && Number.isFinite(level)) {
    if (level <= 5) {
      if (level >= 4.5) return 'High'
      if (level >= 3.5) return 'Moderate to High'
      if (level >= 2.5) return 'Moderate'
      if (level >= 1.5) return 'Low to Moderate'
      return 'Low'
    }
    if (level <= 25) {
      if (level >= 20) return 'High'
      if (level >= 15) return 'Moderate to High'
      if (level >= 10) return 'Moderate'
      if (level >= 5) return 'Low to Moderate'
      return 'Low'
    }
    if (level >= 80) return 'High'
    if (level >= 60) return 'Moderate to High'
    if (level >= 40) return 'Moderate'
    if (level >= 20) return 'Low to Moderate'
    return 'Low'
  }

  const str = level == null ? '' : String(level).trim()
  if (!str) return fallback

  if (/^-?\d+(\.\d+)?$/.test(str)) {
    return normalizeAnomalyRiskLevel(Number(str), fallback)
  }

  const shared = normalizeRiskLevel(str)
  if (shared === 'High') return 'High'
  if (shared === 'Medium to High') return 'Moderate to High'
  if (shared === 'Medium') return 'Moderate'
  if (shared === 'Low to Medium') return 'Low to Moderate'
  if (shared === 'Low') return 'Low'

  const lower = str.toLowerCase()
  const raw = lower.replace(/[\s\-_/]+/g, '')

  // Combined tiers first so "Moderate to High" / "Low to Moderate" never collapse into High / Moderate
  if (
    /(moderate|medium|sedang|menengah)(to|ke)?(high|tinggi)/.test(raw) ||
    /(high|tinggi)(to|ke)?(moderate|medium|sedang)/.test(raw) ||
    ['moderatehigh', 'mediumhigh', 'modhigh', 'sedangtinggi', 'menengahtinggi', 'elevated', 'significant', 'major', 'tier2', 'level4', 'kol3', 'p2', 'kuranglancar'].includes(raw)
  ) {
    return 'Moderate to High'
  }
  if (
    /(low|rendah)(to|ke)?(moderate|medium|sedang)/.test(raw) ||
    /(moderate|medium|sedang)(to|ke)?(low|rendah)/.test(raw) ||
    ['lowmoderate', 'lowmedium', 'lowmod', 'sedangrendah', 'rendahsedang', 'guarded', 'minor', 'tier4', 'level2', 'kol2', 'p4', 'dpk', 'dalamperhatiankhusus'].includes(raw)
  ) {
    return 'Low to Moderate'
  }
  if (
    ['high', 'highrisk', 'risikotinggi', 'veryhigh', 'veryhighrisk', 'critical', 'extreme', 'catastrophic', 'severe', 'urgent', 'blocker', 'fatal', 'sangattinggi', 'kritis', 'bahaya', 'tinggi', 'tier1', 'level5', 'kol4', 'kol5', 'p0', 'p1', 'diragukan', 'macet'].includes(raw) ||
    /crit|extr|catastroph|sever|blocker|fatal|sangat\s*tinggi|kritis|\bvery\s*high\b|\bhigh\s*risk\b|\brisiko\s*tinggi\b|\b(high|tinggi)\b/.test(lower)
  ) {
    return 'High'
  }
  if (
    ['low', 'lowrisk', 'risikorendah', 'verylow', 'verylowrisk', 'minimal', 'insignificant', 'negligible', 'trivial', 'safe', 'normal', 'info', 'rendah', 'sangatrendah', 'aman', 'tier5', 'level1', 'kol1', 'p5', 'lancar'].includes(raw) ||
    /insignif|neglig|trivial|minimal|sangat\s*rendah|\blow\s*risk\b|\brisiko\s*rendah\b|\b(low|rendah)\b/.test(lower)
  ) {
    return 'Low'
  }
  if (
    ['moderate', 'medium', 'mod', 'moderaterisk', 'mediumrisk', 'risikosedang', 'sedang', 'menengah', 'watch', 'warning', 'fair', 'average', 'tier3', 'level3', 'p3'].includes(raw) ||
    /\b(moderate|medium)\s*risk\b|\brisiko\s*sedang\b|\b(moderate|medium|sedang)\b/.test(lower)
  ) {
    return 'Moderate'
  }

  return fallback
}

export const toAnomalyRiskCode = (
  level?: unknown,
  fallback: AnomalyRiskLevelLabel = 'Moderate'
): string => ANOMALY_RISK_LABEL_TO_CODE[normalizeAnomalyRiskLevel(level, fallback)]

// Entity Prediction enforces the exact same 5 canonical risk levels:
// High, Moderate to High, Moderate, Low to Moderate, Low.
export type EntityRiskLevelLabel = AnomalyRiskLevelLabel
export const ENTITY_RISK_LEVELS = ANOMALY_RISK_LEVELS
export const ENTITY_RISK_LABEL_TO_CODE = ANOMALY_RISK_LABEL_TO_CODE
export const normalizeEntityRiskLevel = normalizeAnomalyRiskLevel
export const toEntityRiskCode = toAnomalyRiskCode

// NLP Analysis enforces the exact same 5 canonical risk levels:
// High, Moderate to High, Moderate, Low to Moderate, Low.
export type NlpRiskLevelLabel = AnomalyRiskLevelLabel
export const NLP_RISK_LEVELS = ANOMALY_RISK_LEVELS
export const NLP_RISK_LABEL_TO_CODE = ANOMALY_RISK_LABEL_TO_CODE
export const normalizeNlpRiskLevel = normalizeAnomalyRiskLevel
export const toNlpRiskCode = toAnomalyRiskCode
