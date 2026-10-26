import { describe, expect, it, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'node:fs'
import path from 'node:path'
import en from '../../locales/en/common.json'
import id from '../../locales/id/common.json'
import { useExecutiveSummaryStore } from '../../stores/executive-summary'
import {
  auditPeriodFromAssignmentLetter,
  buildDefaultNarrative,
  formatAuditPeriod,
  formatQuarterPeriod,
  parseExecutionPeriod,
  parseMonthYear,
  prefixNarrative,
  refreshNarrativePeriod,
  type Translate
} from '../../utils/executiveSummaryNarrative'

const read = (relativePath: string) => fs.readFileSync(path.resolve(__dirname, '../../', relativePath), 'utf8')

// Same lookup and {param} replacement as composables/useI18n.ts, for a fixed locale
const translator = (messages: unknown): Translate => (key, params) => {
  let value: unknown = messages
  for (const k of key.split('.')) {
    if (value && typeof value === 'object' && k in value) value = (value as Record<string, unknown>)[k]
    else return key
  }
  if (typeof value !== 'string') return key
  return params ? value.replace(/\{(\w+)\}/g, (m, p) => params[p]?.toString() || m) : value
}

const tId = translator(id)
const tEn = translator(en)

const range = (start: string, end: string) => ({ start: parseMonthYear(start), end: parseMonthYear(end) })

describe('Executive summary narrative period', () => {
  describe('date parsing', () => {
    it('reads month and year from ISO dates, datetimes and DD/MM/YYYY', () => {
      expect(parseMonthYear('2026-01-15')).toEqual({ year: 2026, month: 1 })
      expect(parseMonthYear('2026-03')).toEqual({ year: 2026, month: 3 })
      expect(parseMonthYear('2025-12-01T00:00:00Z')).toEqual({ year: 2025, month: 12 })
      expect(parseMonthYear('20/02/2026')).toEqual({ year: 2026, month: 2 })
      expect(parseMonthYear('')).toBeNull()
      expect(parseMonthYear(undefined)).toBeNull()
      expect(parseMonthYear('Kuartal I')).toBeNull()
      expect(parseMonthYear('2026-13-01')).toBeNull()
    })

    it('splits an assignment letter executionPeriod', () => {
      expect(parseExecutionPeriod('2026-01-05 to 2026-03-20')).toEqual({
        start: { year: 2026, month: 1 },
        end: { year: 2026, month: 3 }
      })
      expect(parseExecutionPeriod('2026-01-05 s/d 2026-02-20').end).toEqual({ year: 2026, month: 2 })
      expect(parseExecutionPeriod('Q1 2026')).toEqual({ start: null, end: null })
    })

    it('prefers startPeriod/finishPeriod and falls back to executionPeriod', () => {
      expect(auditPeriodFromAssignmentLetter({ startPeriod: '2026-01-05', finishPeriod: '2026-02-10', executionPeriod: '2025-01-01 to 2025-01-31' }))
        .toEqual({ start: { year: 2026, month: 1 }, end: { year: 2026, month: 2 } })
      expect(auditPeriodFromAssignmentLetter({ executionPeriod: '2025-12-01 to 2026-01-31' }))
        .toEqual({ start: { year: 2025, month: 12 }, end: { year: 2026, month: 1 } })
      expect(auditPeriodFromAssignmentLetter(null)).toEqual({ start: null, end: null })
    })
  })

  describe('formatAuditPeriod', () => {
    it('shows a single month when the audit starts and ends in the same month', () => {
      expect(formatAuditPeriod(tId, range('2026-01-05', '2026-01-28'))).toBe('Januari 2026')
      expect(formatAuditPeriod(tEn, range('2026-01-05', '2026-01-28'))).toBe('January 2026')
    })

    it('shows a month range within one year', () => {
      expect(formatAuditPeriod(tId, range('2026-01-05', '2026-03-20'))).toBe('Januari – Maret 2026')
      expect(formatAuditPeriod(tEn, range('2026-01-05', '2026-03-20'))).toBe('January – March 2026')
    })

    it('shows both years when the audit crosses a year', () => {
      expect(formatAuditPeriod(tId, range('2025-12-01', '2026-01-31'))).toBe('Desember 2025 – Januari 2026')
      expect(formatAuditPeriod(tEn, range('2025-12-01', '2026-01-31'))).toBe('December 2025 – January 2026')
    })

    it('handles one known date, swapped dates and no dates', () => {
      expect(formatAuditPeriod(tId, range('2026-05-02', ''))).toBe('Mei 2026')
      expect(formatAuditPeriod(tId, range('', '2026-08-02'))).toBe('Agustus 2026')
      expect(formatAuditPeriod(tId, range('2026-03-01', '2026-01-01'))).toBe('Januari – Maret 2026')
      expect(formatAuditPeriod(tId, { start: null, end: null })).toBe('')
    })
  })

  describe('formatQuarterPeriod', () => {
    it('formats the selected quarter and year per locale', () => {
      expect(formatQuarterPeriod(tId, 4, 2026)).toBe('Kuartal 4 2026')
      expect(formatQuarterPeriod(tEn, 4, 2026)).toBe('Q4 2026')
      expect(formatQuarterPeriod(tId, 2)).toBe('Kuartal 2')
      expect(formatQuarterPeriod(tId, 0, 2026)).toBe('')
    })
  })

  describe('narrative text', () => {
    it('starts the Individual narrative with the audit period', () => {
      const text = buildDefaultNarrative(tId, formatAuditPeriod(tId, range('2026-01-05', '2026-01-28')))
      expect(text.startsWith('Periode Januari 2026, SPI telah menerbitkan .... LHA. Total temuan')).toBe(true)
    })

    it('starts the Compilation narrative with the quarter', () => {
      expect(buildDefaultNarrative(tId, formatQuarterPeriod(tId, 4, 2026)).startsWith('Periode Kuartal 4 2026, SPI telah menerbitkan .... LHA')).toBe(true)
      expect(buildDefaultNarrative(tEn, formatQuarterPeriod(tEn, 4, 2026)).startsWith('For the period of Q4 2026, SPI has issued .... LHA')).toBe(true)
    })

    it('leaves the prefix out when no period is known', () => {
      const text = buildDefaultNarrative(tId, '')
      expect(text.startsWith('SPI telah menerbitkan')).toBe(true)
      expect(text).not.toContain('undefined')
      expect(text).not.toContain('Periode')
    })

    it('prefixes an LHA executive summary once and keeps acronyms', () => {
      expect(prefixNarrative(tId, 'Berdasarkan hasil audit ditemukan 3 temuan.', 'Januari 2026'))
        .toBe('Periode Januari 2026, berdasarkan hasil audit ditemukan 3 temuan.')
      const once = prefixNarrative(tId, 'SPI menemukan 3 temuan.', 'Januari 2026')
      expect(once).toBe('Periode Januari 2026, SPI menemukan 3 temuan.')
      expect(prefixNarrative(tId, once, 'Januari 2026')).toBe(once)
      expect(prefixNarrative(tId, once, 'Februari – Maret 2026')).toBe('Periode Februari – Maret 2026, SPI menemukan 3 temuan.')
      expect(prefixNarrative(tId, once, '')).toBe('SPI menemukan 3 temuan.')
    })

    it('updates the period of generated text when the quarter changes, keeping edits after the prefix', () => {
      const q3 = buildDefaultNarrative(tId, formatQuarterPeriod(tId, 3, 2026))
      const q4Period = formatQuarterPeriod(tId, 4, 2026)
      expect(refreshNarrativePeriod(tId, q3, q4Period)).toBe(buildDefaultNarrative(tId, q4Period))

      const edited = 'Periode Kuartal 3 2026, SPI telah menerbitkan 12 LHA dengan catatan khusus.'
      expect(refreshNarrativePeriod(tId, edited, q4Period)).toBe('Periode Kuartal 4 2026, SPI telah menerbitkan 12 LHA dengan catatan khusus.')
    })

    it('adds the period to the untouched template once an audit period is known', () => {
      const noPeriod = buildDefaultNarrative(tId, '')
      expect(refreshNarrativePeriod(tId, noPeriod, 'Januari 2026')).toBe(buildDefaultNarrative(tId, 'Januari 2026'))
    })

    it('leaves free text and an empty narrative unchanged', () => {
      expect(refreshNarrativePeriod(tId, 'Ringkasan ditulis sendiri oleh auditor.', 'Kuartal 4 2026')).toBe('Ringkasan ditulis sendiri oleh auditor.')
      expect(refreshNarrativePeriod(tId, 'Periode ini, SPI fokus pada TI.', 'Kuartal 4 2026')).toBe('Periode ini, SPI fokus pada TI.')
      expect(refreshNarrativePeriod(tId, '', 'Kuartal 4 2026')).toBe('')
    })
  })

  describe('store and forms', () => {
    beforeEach(() => {
      setActivePinia(createPinia())
    })

    it('opens a new compilation with the selected quarter and year at the start of the narrative', () => {
      const store = useExecutiveSummaryStore()
      store.openNewCompilationForm(4, 2025)
      // Tests run with the default locale (en)
      expect(store.form.narrative.startsWith('For the period of Q4 2025, SPI has issued')).toBe(true)
    })

    it('opens a new individual summary without a period until an LHA is selected', () => {
      const store = useExecutiveSummaryStore()
      store.openNewForm(1)
      expect(store.form.narrative.startsWith('SPI has issued')).toBe(true)
      expect(store.defaultNarrativeTemplate('January – March 2026').startsWith('For the period of January – March 2026, SPI has issued')).toBe(true)
    })

    it('wires the period into both forms', () => {
      const individual = read('components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
      expect(individual).toContain('prefixNarrative(t, item.executiveSummary, auditPeriodForReport(item))')
      expect(individual).toContain('refreshNarrativePeriod(t, store.form.narrative, period)')
      expect(individual).toContain('store.defaultNarrativeTemplate(narrativePeriod.value)')

      const compilation = read('components/audit-result-report/ExecutiveSummaryCompilationForm.vue')
      expect(compilation).toContain('formatQuarterPeriod(t, Number(store.form.quarter) || 1, Number(store.form.tahun) || null)')
      expect(compilation).toContain('refreshNarrativePeriod(t, store.form.narrative, period)')
      expect(compilation).toContain('store.defaultNarrativeTemplate(narrativePeriod.value)')
    })

    it('keeps the narrative keys in both locales', () => {
      for (const messages of [en, id]) {
        const narrative = messages.executiveSummaryForm.narrative as Record<string, string>
        expect(narrative.periodPrefix).toContain('{period}')
        expect(narrative.quarterPeriod).toContain('{quarter}')
        expect(narrative.quarterPeriod).toContain('{year}')
        expect(narrative.quarterPeriodNoYear).toContain('{quarter}')
        expect(narrative.templateBody).toContain('SPI')
      }
    })
  })
})
