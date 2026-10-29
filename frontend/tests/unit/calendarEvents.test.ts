import { describe, expect, it } from 'vitest'
import {
  addDays,
  buildMonthGrid,
  compareCalendarEvents,
  eventCoversDay,
  eventInWindow,
  makeCalendarEvent,
  shiftMonth,
  toCalendarDay,
  type CalendarEvent
} from '~/utils/calendarEvents'

const base = { id: 'x', kind: 'atrDue' as const, title: 'T' }

describe('toCalendarDay', () => {
  it('reads ISO dates and keeps the written day of RFC3339 values', () => {
    expect(toCalendarDay('2026-10-22')).toBe('2026-10-22')
    expect(toCalendarDay('2026-10-22T23:30:00+07:00')).toBe('2026-10-22')
  })

  it('reads DD/MM/YYYY', () => {
    expect(toCalendarDay('05/11/2026')).toBe('2026-11-05')
  })

  it('rejects empty, invalid and impossible dates', () => {
    expect(toCalendarDay('')).toBeNull()
    expect(toCalendarDay(null)).toBeNull()
    expect(toCalendarDay('22 Oktober 2026')).toBeNull()
    expect(toCalendarDay('2026-02-30')).toBeNull()
  })
})

describe('makeCalendarEvent', () => {
  it('builds a single-day event', () => {
    expect(makeCalendarEvent(base, '2026-10-22')).toEqual({ ...base, start: '2026-10-22' })
  })

  it('builds a range and collapses a same-day range', () => {
    expect(makeCalendarEvent(base, '2026-10-22', '2026-10-31')).toMatchObject({ start: '2026-10-22', end: '2026-10-31' })
    expect(makeCalendarEvent(base, '2026-10-22', '2026-10-22')).not.toHaveProperty('end')
  })

  it('uses whichever end parses and ignores an end before the start', () => {
    expect(makeCalendarEvent(base, '', '2026-10-31')).toEqual({ ...base, start: '2026-10-31' })
    expect(makeCalendarEvent(base, '2026-10-22', '2026-10-01')).toEqual({ ...base, start: '2026-10-22' })
  })

  it('returns null when nothing parses', () => {
    expect(makeCalendarEvent(base, 'soon', undefined)).toBeNull()
  })
})

describe('event windows', () => {
  const range: CalendarEvent = { ...base, start: '2026-10-22', end: '2026-10-31' }

  it('covers every day of a range', () => {
    expect(eventCoversDay(range, '2026-10-22')).toBe(true)
    expect(eventCoversDay(range, '2026-10-31')).toBe(true)
    expect(eventCoversDay(range, '2026-11-01')).toBe(false)
  })

  it('overlaps a window that touches either end', () => {
    expect(eventInWindow(range, '2026-10-31', '2026-11-13')).toBe(true)
    expect(eventInWindow(range, '2026-11-01', '2026-11-13')).toBe(false)
  })
})

describe('compareCalendarEvents', () => {
  it('puts overdue first, then earlier days, then single-day before range', () => {
    const a: CalendarEvent = { ...base, id: 'a', start: '2026-10-25' }
    const b: CalendarEvent = { ...base, id: 'b', start: '2026-10-25', end: '2026-10-30' }
    const c: CalendarEvent = { ...base, id: 'c', start: '2026-10-28', overdue: true }
    const d: CalendarEvent = { ...base, id: 'd', start: '2026-10-20' }
    expect([a, b, c, d].sort(compareCalendarEvents).map(e => e.id)).toEqual(['c', 'd', 'a', 'b'])
  })
})

describe('month grid', () => {
  it('starts on Monday and always has 42 cells', () => {
    // 1 Oct 2026 is a Thursday, so the grid starts on Monday 28 Sep.
    const grid = buildMonthGrid('2026-10')
    expect(grid).toHaveLength(42)
    expect(grid[0]).toEqual({ day: '2026-09-28', inMonth: false })
    expect(grid[3]).toEqual({ day: '2026-10-01', inMonth: true })
  })

  it('shifts months across years and adds days', () => {
    expect(shiftMonth('2026-01', -1)).toBe('2025-12')
    expect(shiftMonth('2026-12', 1)).toBe('2027-01')
    expect(addDays('2026-10-31', 1)).toBe('2026-11-01')
  })
})
