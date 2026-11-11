import type { RCMItem } from '~/stores/rcm'

// The five COSO dimension scores captured in step 3 of the RCM wizard.
export const RCM_RATING_KEYS = [
  'design_effectiveness_rating',
  'operating_effectiveness_rating',
  'coverage_completeness_rating',
  'timeliness_rating',
  'automation_monitoring_rating'
] as const

export type RCMRatingKey = typeof RCM_RATING_KEYS[number]

// Form state: same shape as an RCM item, except a dimension score may be
// null until the user picks it (null renders as "no scale button selected").
export type RCMFormData = Omit<Partial<RCMItem>, RCMRatingKey> & {
  [K in RCMRatingKey]: number | null
}

export const isValidRating = (value: unknown): value is number =>
  typeof value === 'number' && Number.isInteger(value) && value >= 1 && value <= 5

/** Fresh form for "add new RCM": every dimension score starts empty. */
export const createEmptyRcmForm = (year: number, overrides: Partial<RCMFormData> = {}): RCMFormData => ({
  risk_id: '',
  risk_code: '',
  risk_event: '',
  control_code: '',
  control_description: '',
  control_type: 'Preventive',
  control_owner: '',
  department: 'Head Office',
  year,
  design_effectiveness_weight: 20,
  operating_effectiveness_weight: 20,
  coverage_completeness_weight: 20,
  timeliness_weight: 20,
  automation_monitoring_weight: 20,
  notes: '',
  ...overrides,
  design_effectiveness_rating: null,
  operating_effectiveness_rating: null,
  coverage_completeness_rating: null,
  timeliness_rating: null,
  automation_monitoring_rating: null
})

/** Form for editing an existing item: a deep copy that keeps its saved scores. */
export const rcmFormFromItem = (item: RCMItem): RCMFormData => {
  const copy = JSON.parse(JSON.stringify(item)) as RCMItem
  const form = { ...copy, control_type: copy.control_type || 'Preventive' } as RCMFormData
  for (const key of RCM_RATING_KEYS) {
    form[key] = isValidRating(copy[key]) ? copy[key] : null
  }
  return form
}

export const getMissingRatingKeys = (form: Partial<Record<RCMRatingKey, number | null | undefined>>): RCMRatingKey[] =>
  RCM_RATING_KEYS.filter(key => !isValidRating(form[key]))

export const hasAllRatings = (form: Partial<Record<RCMRatingKey, number | null | undefined>>): boolean =>
  getMissingRatingKeys(form).length === 0

/** Total weighted score (each rating x 4%) or null while any dimension is unscored. */
export const calculateFormScorePercent = (form: Partial<Record<RCMRatingKey, number | null | undefined>>): number | null => {
  if (!hasAllRatings(form)) return null
  return RCM_RATING_KEYS.reduce((sum, key) => sum + (form[key] as number) * 4, 0)
}

/** Narrow a complete form to the item type the store expects; null if any score is missing. */
export const toRcmItemForSave = (form: RCMFormData): RCMItem | null => {
  if (!hasAllRatings(form)) return null
  return form as RCMItem
}
