import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import type { RCMItem } from '~/stores/rcm'
import {
  RCM_RATING_KEYS,
  createEmptyRcmForm,
  rcmFormFromItem,
  getMissingRatingKeys,
  hasAllRatings,
  calculateFormScorePercent,
  toRcmItemForSave
} from '~/utils/rcmForm'

const savedItem: RCMItem = {
  id: 'abc',
  risk_id: 'r1',
  risk_code: 'R-001',
  risk_event: 'Fraud',
  control_code: 'CTL-R-001',
  control_description: 'Dual approval',
  control_owner: 'Finance Manager',
  department: 'Head Office',
  year: 2026,
  design_effectiveness_weight: 20,
  design_effectiveness_rating: 5,
  operating_effectiveness_weight: 20,
  operating_effectiveness_rating: 2,
  coverage_completeness_weight: 20,
  coverage_completeness_rating: 4,
  timeliness_weight: 20,
  timeliness_rating: 1,
  automation_monitoring_weight: 20,
  automation_monitoring_rating: 3,
  total_weighted_score: 60
}

const fullRatings = (form: ReturnType<typeof createEmptyRcmForm>) => {
  for (const key of RCM_RATING_KEYS) form[key] = 3
  return form
}

describe('createEmptyRcmForm', () => {
  it('starts every dimension score empty (null, not 0 or a default)', () => {
    const form = createEmptyRcmForm(2026)
    for (const key of RCM_RATING_KEYS) expect(form[key]).toBeNull()
    expect(form.year).toBe(2026)
    expect(form.control_description).toBe('')
  })

  it('cannot be prefilled with scores through overrides', () => {
    const form = createEmptyRcmForm(2026, { risk_code: 'R-9', design_effectiveness_rating: 4 })
    expect(form.risk_code).toBe('R-9')
    expect(form.design_effectiveness_rating).toBeNull()
  })

  it('returns a fresh object each time, so a previous item never leaks in', () => {
    const first = fullRatings(createEmptyRcmForm(2026))
    first.control_description = 'leftover'
    const second = createEmptyRcmForm(2026)
    expect(second.control_description).toBe('')
    expect(getMissingRatingKeys(second)).toEqual([...RCM_RATING_KEYS])
  })
})

describe('rcmFormFromItem (edit)', () => {
  it('keeps the saved scores and does not mutate the source', () => {
    const form = rcmFormFromItem(savedItem)
    expect(form.design_effectiveness_rating).toBe(5)
    expect(form.timeliness_rating).toBe(1)
    expect(form.control_type).toBe('Preventive')
    form.design_effectiveness_rating = 1
    expect(savedItem.design_effectiveness_rating).toBe(5)
  })

  it('treats an out-of-range saved score as unscored', () => {
    const form = rcmFormFromItem({ ...savedItem, automation_monitoring_rating: 0 })
    expect(form.automation_monitoring_rating).toBeNull()
    expect(hasAllRatings(form)).toBe(false)
  })
})

describe('rating validation and score preview', () => {
  it('lists missing dimensions and rejects invalid values', () => {
    const form = createEmptyRcmForm(2026)
    form.design_effectiveness_rating = 4
    form.operating_effectiveness_rating = 0
    form.coverage_completeness_rating = 6
    form.timeliness_rating = 2.5
    expect(getMissingRatingKeys(form)).toEqual([
      'operating_effectiveness_rating',
      'coverage_completeness_rating',
      'timeliness_rating',
      'automation_monitoring_rating'
    ])
  })

  it('preview is null until all five are scored, then sums rating x 4%', () => {
    const form = createEmptyRcmForm(2026)
    expect(calculateFormScorePercent(form)).toBeNull()
    form.design_effectiveness_rating = 5
    form.operating_effectiveness_rating = 5
    form.coverage_completeness_rating = 5
    form.timeliness_rating = 5
    expect(calculateFormScorePercent(form)).toBeNull()
    form.automation_monitoring_rating = 1
    expect(calculateFormScorePercent(form)).toBe(84)
    expect(calculateFormScorePercent(rcmFormFromItem(savedItem))).toBe(60)
  })

  it('toRcmItemForSave blocks incomplete forms and passes complete ones with no nulls', () => {
    const form = createEmptyRcmForm(2026)
    form.design_effectiveness_rating = 2
    expect(toRcmItemForSave(form)).toBeNull()
    const item = toRcmItemForSave(fullRatings(form))
    expect(item).not.toBeNull()
    for (const key of RCM_RATING_KEYS) expect(item![key]).toBe(3)
  })
})

describe('i18n keys', () => {
  const load = (lang: string) =>
    JSON.parse(readFileSync(resolve(__dirname, `../../locales/${lang}/common.json`), 'utf8'))

  it.each(['en', 'id'])('%s has the dimension-required messages', (lang) => {
    const rcm = load(lang).rcm
    expect(rcm.toasts.dimensionsRequired).toBeTruthy()
    expect(rcm.modal.dimensionRequired).toBeTruthy()
  })
})
