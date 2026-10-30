// Dashboard calendar widget: event shape and pure date helpers.
// Interim source: dates the dashboards already load (Surat Tugas period, fieldwork dates, ATR due dates).
// TODO(B4): replace the per-dashboard event builders with GET /api/v1/me/calendar (docs/audit-calendar-plan.md).

export type CalendarEventKind = 'engagement' | 'atrDue' | 'interview' | 'observation' | 'sample' | 'pbcDue' | 'testControlDue'

export interface CalendarEvent {
  id: string
  kind: CalendarEventKind
  /** First day, 'YYYY-MM-DD'. */
  start: string
  /** Last day for range events, 'YYYY-MM-DD'. Omitted for single-day events. */
  end?: string
  title: string
  subtitle?: string
  overdue?: boolean
  to?: string
}

export interface CalendarDayCell {
  day: string
  inMonth: boolean
}

const DAY_MS = 24 * 60 * 60 * 1000

/**
 * Calendar day of a stored date as 'YYYY-MM-DD'.
 * Accepts 'YYYY-MM-DD', RFC3339 (date part kept as written, no time-zone shift) and 'DD/MM/YYYY'.
 * Returns null for empty or unparseable values.
 */
export const toCalendarDay = (value: unknown): string | null => {
  if (typeof value !== 'string') return null
  const raw = value.trim()
  if (!raw) return null
  const iso = /^(\d{4})-(\d{2})-(\d{2})/.exec(raw)
  const dmy = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(raw)
  let y: number, m: number, d: number
  if (iso) {
    [y, m, d] = [Number(iso[1]), Number(iso[2]), Number(iso[3])]
  } else if (dmy) {
    [y, m, d] = [Number(dmy[3]), Number(dmy[2]), Number(dmy[1])]
  } else {
    return null
  }
  const date = new Date(Date.UTC(y, m - 1, d))
  if (date.getUTCFullYear() !== y || date.getUTCMonth() !== m - 1 || date.getUTCDate() !== d) return null
  return date.toISOString().slice(0, 10)
}

export const addDays = (day: string, days: number): string =>
  new Date(Date.parse(`${day}T00:00:00Z`) + days * DAY_MS).toISOString().slice(0, 10)

/** Builds a range or single-day event. Returns null when the start (or, for a range, both ends) cannot be parsed. */
export const makeCalendarEvent = (
  base: Omit<CalendarEvent, 'start' | 'end'>,
  startValue: unknown,
  endValue?: unknown
): CalendarEvent | null => {
  const start = toCalendarDay(startValue)
  const end = toCalendarDay(endValue)
  if (!start && !end) return null
  const first = start || end!
  const last = end && start && end < start ? start : end
  if (!last || last === first) return { ...base, start: first }
  return { ...base, start: first, end: last }
}

/** Last day an event covers. */
export const eventEnd = (event: CalendarEvent) => event.end || event.start

export const eventCoversDay = (event: CalendarEvent, day: string) =>
  event.start <= day && day <= eventEnd(event)

/** True when the event touches the [from, to] window (inclusive). */
export const eventInWindow = (event: CalendarEvent, from: string, to: string) =>
  event.start <= to && eventEnd(event) >= from

/** Overdue first, then by start day, then single-day before range. */
export const compareCalendarEvents = (a: CalendarEvent, b: CalendarEvent) => {
  if (!!a.overdue !== !!b.overdue) return a.overdue ? -1 : 1
  if (a.start !== b.start) return a.start.localeCompare(b.start)
  if (!a.end !== !b.end) return a.end ? 1 : -1
  return a.title.localeCompare(b.title)
}

/**
 * Six-week month grid starting on Monday, for a month given as 'YYYY-MM'.
 * Always 42 cells so the widget height does not jump between months.
 */
export const buildMonthGrid = (month: string): CalendarDayCell[] => {
  const first = `${month}-01`
  // getUTCDay: 0 = Sunday. Shift so Monday = 0.
  const weekday = (new Date(`${first}T00:00:00Z`).getUTCDay() + 6) % 7
  const gridStart = addDays(first, -weekday)
  return Array.from({ length: 42 }, (_, i) => {
    const day = addDays(gridStart, i)
    return { day, inMonth: day.startsWith(month) }
  })
}

export const shiftMonth = (month: string, delta: number): string => {
  const [y, m] = month.split('-').map(Number)
  const date = new Date(Date.UTC(y!, m! - 1 + delta, 1))
  return date.toISOString().slice(0, 7)
}
