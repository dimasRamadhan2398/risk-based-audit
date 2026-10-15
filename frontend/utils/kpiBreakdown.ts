// Pure helpers for the KPI Detailed Breakdown table (GET /performance/kpi-breakdown).
// The backend filters, pages and computes gap/status; these only build the query and format what it returned.

import type { TranslateFn } from '~/utils/kpiPerformanceLabels'

export type KpiBreakdownSource = 'strategic_plan' | 'kpi_achievement'
export type KpiBreakdownStatus = 'Exceeded' | 'On Track' | 'Needs Attention' | 'No Target'

export interface KpiBreakdownItem {
  id: string
  source: KpiBreakdownSource
  metric: string
  category: string
  period: string
  unit: string
  target: number
  actual: number
  gap: number
  gapIsPositive: boolean
  achievementRate: number | null
  status: KpiBreakdownStatus | string
}

export interface KpiBreakdownPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

/** Filter options for the year (`data.filters`): every distinct non-empty value, independent of the active filters and page. */
export interface KpiBreakdownFilterOptions {
  categories: string[]
  periods: string[]
}

export interface KpiBreakdownQuery {
  year: number
  page: number
  pageSize: number
  search: string
  category: string
  status: string
  period: string
}

export const KPI_BREAKDOWN_PAGE_SIZES = [10, 25, 50]
export const KPI_BREAKDOWN_DEFAULT_PAGE_SIZE = 10
export const KPI_BREAKDOWN_MAX_PAGE_SIZE = 100
export const KPI_BREAKDOWN_STATUSES: KpiBreakdownStatus[] = ['Exceeded', 'On Track', 'Needs Attention', 'No Target']

export const emptyKpiBreakdownPagination = (pageSize = KPI_BREAKDOWN_DEFAULT_PAGE_SIZE): KpiBreakdownPagination =>
  ({ page: 1, page_size: pageSize, total: 0, total_pages: 0 })

/** Query string for the endpoint; empty filters are left out. */
export const buildKpiBreakdownParams = (q: KpiBreakdownQuery): Record<string, string | number> => {
  const params: Record<string, string | number> = { year: q.year, page: q.page, page_size: q.pageSize }
  const search = q.search.trim()
  if (search) params.search = search
  if (q.category) params.category = q.category
  if (q.status) params.status = q.status
  if (q.period) params.period = q.period
  return params
}

export const emptyKpiBreakdownFilterOptions = (): KpiBreakdownFilterOptions => ({ categories: [], periods: [] })

const stringList = (value: unknown): string[] => {
  if (!Array.isArray(value)) return []
  const out: string[] = []
  for (const v of value) {
    const s = typeof v === 'string' ? v.trim() : ''
    if (s && !out.includes(s)) out.push(s)
  }
  return out
}

/**
 * Read `data.filters` from the response, keeping the server's order (it sorts them).
 * A missing or malformed `filters` (older backend) gives empty lists, so the menus are hidden instead of guessed.
 */
export const parseKpiBreakdownFilters = (raw: unknown): KpiBreakdownFilterOptions => {
  const f = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>
  return { categories: stringList(f.categories), periods: stringList(f.periods) }
}

/** Menu values for a filter: the server's options, plus the active value so it can still be seen and cleared. */
export const kpiFilterMenuValues = (options: string[], selected?: string): string[] =>
  mergeDistinct(options, [], selected)

const toInt = (value: unknown, fallback: number) => {
  const n = Number(value)
  return Number.isFinite(n) ? Math.trunc(n) : fallback
}

/** Read `data.pagination` from the response; missing fields fall back to the request / zero. */
export const parseKpiBreakdownPagination = (
  raw: Partial<Record<keyof KpiBreakdownPagination, unknown>> | null | undefined,
  requested: { page: number, pageSize: number }
): KpiBreakdownPagination => {
  const pageSize = Math.max(1, toInt(raw?.page_size, requested.pageSize))
  const total = Math.max(0, toInt(raw?.total, 0))
  const totalPages = raw?.total_pages !== undefined ? Math.max(0, toInt(raw.total_pages, 0)) : Math.ceil(total / pageSize)
  return { page: Math.max(1, toInt(raw?.page, requested.page)), page_size: pageSize, total, total_pages: totalPages }
}

/** 1-based row range of the current page; { from: 0, to: 0 } when there is nothing to show. */
export const kpiBreakdownRange = (p: KpiBreakdownPagination): { from: number, to: number } => {
  if (p.total <= 0) return { from: 0, to: 0 }
  const from = (p.page - 1) * p.page_size + 1
  if (from > p.total) return { from: 0, to: 0 }
  return { from, to: Math.min(p.page * p.page_size, p.total) }
}

/** Footer text: "Showing 1-10 of 104", or the no-data text when the range is empty. */
export const kpiBreakdownRangeText = (t: TranslateFn, p: KpiBreakdownPagination): string => {
  const { from, to } = kpiBreakdownRange(p)
  if (!from) return t('kpiPerformance.table.showingNone')
  return t('kpiPerformance.table.showing', { from, to, total: p.total })
}

/** Number with up to 2 decimals, plus the unit: "92.5%", "Rp 1,500,000", "14 Day", "4.7". */
export const formatKpiValue = (value: number | null | undefined, unit?: string, locale = 'en-US'): string => {
  if (value === null || value === undefined || !Number.isFinite(Number(value))) return '-'
  const text = new Intl.NumberFormat(locale, { maximumFractionDigits: 2 }).format(Number(value))
  const u = (unit || '').trim()
  if (u === '%') return `${text}%`
  if (u === 'Rp') return `Rp ${text}`
  if (!u || u === 'Score' || u === 'Amount') return text
  return `${text} ${u}`
}

/** Gap with an explicit sign: "+2.5%", "-3 Day", "0%". */
export const formatKpiGap = (gap: number | null | undefined, unit?: string, locale = 'en-US'): string => {
  if (gap === null || gap === undefined || !Number.isFinite(Number(gap))) return '-'
  // Round first so a tiny float gap does not render as "+0".
  const n = Math.round(Number(gap) * 100) / 100
  if (n === 0) return formatKpiValue(0, unit, locale)
  return `${n > 0 ? '+' : '-'}${formatKpiValue(Math.abs(n), unit, locale)}`
}

/** Text colour of the gap cell: neutral at zero or without a target, otherwise by the backend's gapIsPositive. */
export const kpiGapClass = (item: Pick<KpiBreakdownItem, 'gap' | 'gapIsPositive' | 'status'>): string => {
  if (item.status === 'No Target' || !Math.round(Number(item.gap) * 100)) return 'text-gray-900 dark:text-white'
  return item.gapIsPositive ? 'text-emerald-500' : 'text-red-500'
}

// Keyed by the raw status value from the API, not by its translated label.
const STATUS_COLORS: Record<string, string> = {
  'Exceeded': 'bg-secondary-500',
  'On Track': 'bg-emerald-500',
  'Needs Attention': 'bg-red-500',
  'No Target': 'bg-gray-400'
}

export const kpiStatusColor = (status?: string): string => (status && STATUS_COLORS[status]) || 'bg-gray-400'

/** Category text; the API sends "" when it does not know the category. */
export const kpiCategoryText = (label: string): string => label || '-'

/** Distinct non-empty values, keeping `extra` (e.g. the selected filter) so it never disappears from its menu. */
export const mergeDistinct = (known: string[], incoming: Array<string | null | undefined>, extra?: string): string[] => {
  const out = [...known]
  for (const v of [...incoming, extra]) {
    const s = (v ?? '').trim()
    if (s && !out.includes(s)) out.push(s)
  }
  return out
}
