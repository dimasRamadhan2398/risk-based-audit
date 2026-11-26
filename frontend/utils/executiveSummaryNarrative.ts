/**
 * Executive summary narrative helpers shared by the executive summary store, the Individual and
 * Compilation forms and the unit tests.
 *
 * The generated narrative always starts with the period it covers:
 *   Individual:  "Periode Januari 2026, SPI telah menerbitkan ..." (audit period of the linked LHA)
 *   Compilation: "Periode Kuartal 4 2026, SPI telah menerbitkan ..." (selected quarter and year)
 * Texts come from i18n (executiveSummaryForm.narrative.*); month names from common.datePicker.months.
 * When no period is known the prefix is left out and the narrative starts with the template body.
 */

export type Translate = (key: string, params?: Record<string, string | number>) => string

export const NARRATIVE_KEYS = {
  periodPrefix: 'executiveSummaryForm.narrative.periodPrefix',
  templateBody: 'executiveSummaryForm.narrative.templateBody',
  syncedBody: 'executiveSummaryForm.narrative.syncedBody',
  quarterPeriod: 'executiveSummaryForm.narrative.quarterPeriod',
  quarterPeriodNoYear: 'executiveSummaryForm.narrative.quarterPeriodNoYear'
} as const

const MONTH_KEY = 'common.datePicker.months'
const RANGE_SEPARATOR = ' – '

export interface MonthYear {
  year: number
  /** 1-12 */
  month: number
}

export interface PeriodRange {
  start: MonthYear | null
  end: MonthYear | null
}

export interface NarrativeSyncData {
  lhaCount: number
  totalFindings: number
  recommendationCount: number
  onTimeRate: number | string
  overdueHighRiskCount: number
  overdueTopic: string
  recurringProcess: string
}

export interface NarrativeSourceInput {
  isCompilation?: boolean
  nomorDokumen?: string | null
  jumlahLaporan?: number | null
  totalTemuan?: number | null
  risikoHigh?: number | null
  risikoModerateHigh?: number | null
  risikoModerate?: number | null
  risikoLowModerate?: number | null
  risikoLow?: number | null
  jumlahRekomendasi?: number | null
  followUpTable?: ReadonlyArray<{ status?: string, jumlah?: number | string | null, persentase?: number | string | null }> | null
  topFindings?: ReadonlyArray<{
    unitDivision?: string | null
    judulTemuan?: string | null
    risiko?: string | null
    statusTL?: string | null
    usulan?: string | null
  }> | null
  matriksKompilasi?: ReadonlyArray<{
    nomor?: string | null
    division?: string | null
    unitKerja?: string | null
    prosesBisnis?: string | null
    judulTemuan?: string | null
    nilaiRisiko?: string | null
    rekomendasi?: string | null
    status?: string | null
  }> | null
  akarMasalah?: string | null
  linkedLha?: {
    reportNumber?: string | null
    reportTitle?: string | null
    auditObject?: string | null
    department?: string | null
    findingsCount?: number | null
    findings?: ReadonlyArray<{
      id?: string | null
      title?: string | null
      finding?: string | null
      category?: string | null
      significanceLevel?: string | null
      risiko?: string | null
      action?: string | null
      recommendation?: string | null
      recommendations?: ReadonlyArray<string> | null
      actionPlans?: ReadonlyArray<unknown> | null
      unitDivision?: string | null
      department?: string | null
    }> | null
  } | null
  quarterlyLhas?: ReadonlyArray<{
    reportNumber?: string | null
    reportTitle?: string | null
    auditObject?: string | null
    department?: string | null
    findingsCount?: number | null
    findings?: ReadonlyArray<{
      id?: string | null
      title?: string | null
      finding?: string | null
      category?: string | null
      significanceLevel?: string | null
      risiko?: string | null
      action?: string | null
      recommendation?: string | null
      recommendations?: ReadonlyArray<string> | null
      actionPlans?: ReadonlyArray<unknown> | null
    }> | null
  }> | null
  formatDepartmentName?: (val?: string) => string
}

const toMonthYear = (year: number, month: number): MonthYear | null =>
  Number.isInteger(year) && year >= 1000 && Number.isInteger(month) && month >= 1 && month <= 12
    ? { year, month }
    : null

/** Reads the month and year of 'YYYY-MM', 'YYYY-MM-DD', an ISO datetime or 'DD/MM/YYYY' / 'DD-MM-YYYY'. */
export const parseMonthYear = (value?: string | null): MonthYear | null => {
  const text = String(value ?? '').trim()
  if (!text) return null
  const iso = text.match(/^(\d{4})-(\d{1,2})(?:-\d{1,2})?(?:[T\s].*)?$/)
  if (iso) return toMonthYear(Number(iso[1]), Number(iso[2]))
  const dmy = text.match(/^\d{1,2}[/-](\d{1,2})[/-](\d{4})$/)
  if (dmy) return toMonthYear(Number(dmy[2]), Number(dmy[1]))
  return null
}

/** Splits an assignment letter executionPeriod ("2026-01-05 to 2026-03-20", "... s/d ...") into its two dates. */
export const parseExecutionPeriod = (value?: string | null): PeriodRange => {
  const parts = String(value ?? '').split(/\s+(?:to|s\/d|sd|until|hingga|sampai|–|—|-)\s+/i)
  return { start: parseMonthYear(parts[0]), end: parseMonthYear(parts[1]) }
}

/** Audit period of an assignment letter: startPeriod/finishPeriod, else the executionPeriod text. */
export const auditPeriodFromAssignmentLetter = (letter?: {
  startPeriod?: string | null
  finishPeriod?: string | null
  executionPeriod?: string | null
} | null): PeriodRange => {
  if (!letter) return { start: null, end: null }
  const fromExecution = parseExecutionPeriod(letter.executionPeriod)
  return {
    start: parseMonthYear(letter.startPeriod) ?? fromExecution.start,
    end: parseMonthYear(letter.finishPeriod) ?? fromExecution.end
  }
}

const monthIndex = (value: MonthYear) => value.year * 12 + (value.month - 1)

/** Localized month name (id: Januari, en: January). */
export const monthName = (t: Translate, month: number): string => t(`${MONTH_KEY}.${month - 1}`)

/**
 * Formats an audit period as month(s) and year:
 *   same month           -> "Januari 2026"
 *   same year            -> "Januari – Maret 2026"
 *   different years      -> "Desember 2025 – Januari 2026"
 * Returns '' when neither date is known. A single known date is shown as one month.
 */
export const formatAuditPeriod = (t: Translate, range: PeriodRange): string => {
  let start = range.start ?? range.end
  let end = range.end ?? range.start
  if (!start || !end) return ''
  if (monthIndex(end) < monthIndex(start)) [start, end] = [end, start]
  const endLabel = `${monthName(t, end.month)} ${end.year}`
  if (monthIndex(start) === monthIndex(end)) return endLabel
  const startLabel = start.year === end.year
    ? monthName(t, start.month)
    : `${monthName(t, start.month)} ${start.year}`
  return `${startLabel}${RANGE_SEPARATOR}${endLabel}`
}

/** Compilation period: "Kuartal 4 2026" / "Q4 2026". */
export const formatQuarterPeriod = (t: Translate, quarter: number, year?: number | null): string => {
  const q = Number(quarter)
  if (!Number.isInteger(q) || q < 1 || q > 4) return ''
  const y = Number(year)
  return y > 0
    ? t(NARRATIVE_KEYS.quarterPeriod, { quarter: q, year: y })
    : t(NARRATIVE_KEYS.quarterPeriodNoYear, { quarter: q })
}

/** "Periode Januari 2026," / "For the period of January 2026,", or '' without a period. */
export const narrativePrefix = (t: Translate, period: string): string =>
  period ? t(NARRATIVE_KEYS.periodPrefix, { period }) : ''

/** Default narrative: the period prefix followed by the template body. */
export const buildDefaultNarrative = (t: Translate, period: string): string =>
  [narrativePrefix(t, period), t(NARRATIVE_KEYS.templateBody)].filter(Boolean).join(' ')

const HIGH_RISK_VALUES = new Set([
  'very significant',
  'significant',
  'tinggi',
  'sangat tinggi',
  'sangat signifikan',
  'signifikan',
  'high',
  'moderate-high',
  'moderate to high',
  'moderatehigh',
  'critical'
])

const OVERDUE_VALUES = new Set([
  'overdue',
  'terlambat',
  'overdue/telat',
  'telat'
])

export const isHighRiskValue = (val?: string | null): boolean =>
  HIGH_RISK_VALUES.has(String(val ?? '').trim().toLowerCase())

export const isOverdueStatusValue = (val?: string | null): boolean =>
  OVERDUE_VALUES.has(String(val ?? '').trim().toLowerCase())

export const cleanProcessLabel = (raw?: string | null): string => {
  const trimmed = String(raw ?? '').trim()
  if (!trimmed || trimmed === '-') return ''
  const stripped = trimmed
    .replace(/^(?:laporan\s+hasil\s+audit|lha)\s+(?:atas\s+|pada\s+)?/i, '')
    .replace(/^audit\s+(?:atas\s+|pada\s+)?/i, '')
    .trim()
  return stripped || trimmed
}

const countReportRecommendations = (findings?: ReadonlyArray<{
  action?: string | null
  recommendation?: string | null
  recommendations?: ReadonlyArray<string> | null
  actionPlans?: ReadonlyArray<unknown> | null
}> | null): number => {
  if (!Array.isArray(findings) || findings.length === 0) return 0
  return findings.reduce((sum, f) => {
    if (Array.isArray(f.recommendations) && f.recommendations.length > 0) return sum + f.recommendations.length
    if (Array.isArray(f.actionPlans) && f.actionPlans.length > 0) return sum + f.actionPlans.length
    return sum + 1
  }, 0)
}

/**
 * Extracts synchronized narrative metrics and topics from the connected Executive Summary
 * (Individual or Compilation) state: LHA count, total findings, recommendation count,
 * on-time completion rate from followUpTable, overdue/late high-risk findings count & related topic,
 * and the most frequently recurring process.
 */
export const deriveNarrativeSyncData = (source: NarrativeSourceInput = {}): NarrativeSyncData => {
  const isCompilation = Boolean(source.isCompilation)
  const fmtDept = (val?: string | null): string => {
    const raw = String(val ?? '').trim()
    if (!raw || raw === '-') return ''
    if (source.formatDepartmentName) {
      const formatted = source.formatDepartmentName(raw)
      if (formatted && formatted !== '-') return formatted
    }
    return raw
  }

  // 1. LHA Count
  const explicitLaporan = Number(source.jumlahLaporan) || 0
  let lhaCount = explicitLaporan
  if (lhaCount <= 0) {
    if (isCompilation) {
      lhaCount = source.quarterlyLhas?.length ?? 0
    } else if (source.linkedLha || (source.nomorDokumen && !String(source.nomorDokumen).toUpperCase().startsWith('ESC-'))) {
      lhaCount = 1
    }
  }

  // 2. Total Findings
  const riskSum =
    (Number(source.risikoHigh) || 0) +
    (Number(source.risikoModerateHigh) || 0) +
    (Number(source.risikoModerate) || 0) +
    (Number(source.risikoLowModerate) || 0) +
    (Number(source.risikoLow) || 0)
  const fuRows = Array.isArray(source.followUpTable) ? source.followUpTable : []
  const fuTotal = fuRows.reduce((sum, r) => sum + (Number(r?.jumlah) || 0), 0)
  const linkedFindings = Array.isArray(source.linkedLha?.findings) ? source.linkedLha!.findings! : []
  const linkedFindingsCount = linkedFindings.length || (Number(source.linkedLha?.findingsCount) || 0)
  const quarterlyFindingsCount = (source.quarterlyLhas || []).reduce(
    (sum, r) => sum + ((Array.isArray(r.findings) ? r.findings.length : 0) || (Number(r.findingsCount) || 0)),
    0
  )
  const matrixRows = Array.isArray(source.matriksKompilasi) ? source.matriksKompilasi : []
  const topRows = Array.isArray(source.topFindings) ? source.topFindings : []

  let totalFindings = Number(source.totalTemuan) || 0
  if (totalFindings <= 0) {
    if (!isCompilation && linkedFindingsCount > 0) totalFindings = linkedFindingsCount
    else if (riskSum > 0) totalFindings = riskSum
    else if (fuTotal > 0) totalFindings = fuTotal
    else if (matrixRows.length > 0) totalFindings = matrixRows.length
    else if (quarterlyFindingsCount > 0) totalFindings = quarterlyFindingsCount
    else if (topRows.length > 0) totalFindings = topRows.length
  }

  // 3. Recommendation Count
  let recommendationCount = Number(source.jumlahRekomendasi) || 0
  if (recommendationCount <= 0) {
    if (!isCompilation && linkedFindings.length > 0) {
      recommendationCount = countReportRecommendations(linkedFindings)
    } else if (isCompilation && Array.isArray(source.quarterlyLhas) && source.quarterlyLhas.length > 0) {
      recommendationCount = source.quarterlyLhas.reduce(
        (sum, r) => sum + (countReportRecommendations(r.findings) || (Number(r.findingsCount) || 0)),
        0
      )
    } else if (matrixRows.length > 0) {
      recommendationCount = matrixRows.filter(r => String(r.rekomendasi ?? '').trim() !== '').length || matrixRows.length
    }
    if (recommendationCount <= 0) {
      recommendationCount = totalFindings
    }
  }

  // 4. On-time completion rate (%) from Follow-up Status Table (Closed share)
  const closedRow = fuRows.find(r => String(r?.status ?? '').trim().toLowerCase() === 'closed')
  const closedCount = Number(closedRow?.jumlah) || 0
  let onTimeRate: number = 0
  if (fuTotal > 0) {
    onTimeRate = Number(((closedCount / fuTotal) * 100).toFixed(1))
  } else if (closedRow && Number(closedRow.persentase) > 0) {
    onTimeRate = Number(Number(closedRow.persentase).toFixed(1))
  }

  // 5. Overdue/late high-risk findings count & related topic
  interface CandidateFinding {
    title: string
    risk: string
    status: string
    process: string
  }
  const candidates: CandidateFinding[] = []
  const seenTitles = new Set<string>()
  const addCandidate = (title?: string | null, risk?: string | null, status?: string | null, process?: string | null) => {
    const tStr = String(title ?? '').trim()
    const key = tStr.toLowerCase()
    if (tStr && seenTitles.has(key)) return
    if (tStr) seenTitles.add(key)
    candidates.push({
      title: tStr,
      risk: String(risk ?? '').trim(),
      status: String(status ?? '').trim(),
      process: String(process ?? '').trim()
    })
  }

  for (const r of matrixRows) {
    addCandidate(
      r.judulTemuan,
      r.nilaiRisiko,
      r.status,
      cleanProcessLabel(r.prosesBisnis) || fmtDept(r.unitKerja || r.division)
    )
  }
  for (const f of topRows) {
    addCandidate(
      f.judulTemuan,
      f.risiko,
      f.statusTL,
      fmtDept(f.unitDivision)
    )
  }
  if (!isCompilation && linkedFindings.length > 0) {
    for (const f of linkedFindings) {
      addCandidate(
        f.title || f.finding,
        f.category || f.significanceLevel || f.risiko,
        '',
        fmtDept(f.unitDivision || f.department || source.linkedLha?.department)
      )
    }
  } else if (isCompilation && Array.isArray(source.quarterlyLhas)) {
    for (const lha of source.quarterlyLhas) {
      for (const f of lha.findings || []) {
        addCandidate(
          f.title || f.finding,
          f.category || f.significanceLevel || f.risiko,
          '',
          cleanProcessLabel(lha.auditObject || lha.reportTitle) || fmtDept(lha.department)
        )
      }
    }
  }

  const overdueHighRisk = candidates.filter(c => isOverdueStatusValue(c.status) && isHighRiskValue(c.risk))
  const overdueAny = candidates.filter(c => isOverdueStatusValue(c.status))
  const highRiskAny = candidates.filter(c => isHighRiskValue(c.risk))

  const overdueRowCount = Number(fuRows.find(r => String(r?.status ?? '').trim().toLowerCase() === 'overdue')?.jumlah) || 0
  const highRiskStatCount = (Number(source.risikoHigh) || 0) + (Number(source.risikoModerateHigh) || 0)

  let overdueHighRiskCount = overdueHighRisk.length
  if (overdueHighRiskCount === 0 && overdueRowCount > 0) {
    const hasClassifiedStatuses = candidates.some(c => c.status !== '')
    const hasClassifiedRisks = candidates.some(c => c.risk !== '')
    if (!hasClassifiedStatuses || !hasClassifiedRisks) {
      overdueHighRiskCount = highRiskStatCount > 0 ? Math.min(overdueRowCount, highRiskStatCount) : overdueRowCount
    }
  }

  const pickDistinctTitles = (items: CandidateFinding[], limit = 2): string => {
    const out: string[] = []
    for (const it of items) {
      const tStr = it.title.trim()
      if (tStr && !out.includes(tStr)) {
        out.push(tStr)
        if (out.length >= limit) break
      }
    }
    return out.join(', ')
  }

  let overdueTopic =
    pickDistinctTitles(overdueHighRisk) ||
    pickDistinctTitles(overdueAny) ||
    pickDistinctTitles(highRiskAny) ||
    pickDistinctTitles(candidates) ||
    cleanProcessLabel(source.linkedLha?.auditObject || source.linkedLha?.reportTitle) ||
    fmtDept(source.linkedLha?.department) ||
    '-'

  // 6. Most frequently recurring process
  const processFreq = new Map<string, { label: string, count: number }>()
  const bumpProcess = (raw?: string | null) => {
    const label = String(raw ?? '').trim()
    if (!label || label === '-') return
    const key = label.toLowerCase()
    const cur = processFreq.get(key)
    if (cur) cur.count++
    else processFreq.set(key, { label, count: 1 })
  }

  for (const r of matrixRows) {
    bumpProcess(cleanProcessLabel(r.prosesBisnis) || fmtDept(r.unitKerja || r.division))
  }
  for (const f of topRows) {
    bumpProcess(fmtDept(f.unitDivision))
  }
  if (!isCompilation && source.linkedLha) {
    bumpProcess(cleanProcessLabel(source.linkedLha.auditObject || source.linkedLha.reportTitle) || fmtDept(source.linkedLha.department))
  }
  if (isCompilation && Array.isArray(source.quarterlyLhas)) {
    for (const lha of source.quarterlyLhas) {
      const weight = (Array.isArray(lha.findings) ? lha.findings.length : 0) || (Number(lha.findingsCount) || 1)
      const label = cleanProcessLabel(lha.auditObject || lha.reportTitle) || fmtDept(lha.department)
      for (let i = 0; i < weight; i++) bumpProcess(label)
    }
  }

  let recurringProcess = ''
  let bestCount = 0
  for (const entry of processFreq.values()) {
    if (entry.count > bestCount) {
      bestCount = entry.count
      recurringProcess = entry.label
    }
  }
  if (!recurringProcess && source.akarMasalah) {
    const match = String(source.akarMasalah).match(/:\s*([^.;\n]+)/)
    if (match?.[1]) recurringProcess = match[1].trim()
  }
  if (!recurringProcess) {
    recurringProcess = overdueTopic !== '-' ? overdueTopic : '-'
  }

  return {
    lhaCount,
    totalFindings,
    recommendationCount,
    onTimeRate,
    overdueHighRiskCount,
    overdueTopic,
    recurringProcess
  }
}

/** Builds the synchronized narrative text with all placeholders filled from NarrativeSyncData. */
export const buildSyncedNarrative = (t: Translate, period: string, data: NarrativeSyncData): string => {
  const body = t(NARRATIVE_KEYS.syncedBody, {
    lhaCount: data.lhaCount,
    totalFindings: data.totalFindings,
    recommendationCount: data.recommendationCount,
    onTimeRate: data.onTimeRate,
    overdueHighRiskCount: data.overdueHighRiskCount,
    overdueTopic: data.overdueTopic || '-',
    recurringProcess: data.recurringProcess || '-'
  })
  return [narrativePrefix(t, period), body].filter(Boolean).join(' ')
}

const escapeRegExp = (value: string) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

// A generated prefix of the current locale. The period part must contain a digit (year or quarter),
// so ordinary text such as "Periode ini, ..." is not taken for a generated prefix.
const prefixPattern = (t: Translate): RegExp | null => {
  const template = t(NARRATIVE_KEYS.periodPrefix)
  const [before, after] = template.split('{period}')
  if (before === undefined || after === undefined) return null
  return new RegExp(`^\\s*${escapeRegExp(before)}[^,\\n]*\\d[^,\\n]*?${escapeRegExp(after)}\\s*`)
}

export const hasPeriodPrefix = (t: Translate, text: string): boolean =>
  prefixPattern(t)?.test(text ?? '') ?? false

export const stripPeriodPrefix = (t: Translate, text: string): string => {
  const pattern = prefixPattern(t)
  return pattern ? (text ?? '').replace(pattern, '') : (text ?? '')
}

// "Berdasarkan hasil ..." -> "berdasarkan hasil ..." after the prefix; acronyms ("SPI", "LHA") stay as they are.
const lowerFirstWord = (text: string) => text.replace(/^(\p{Lu})(?=\p{Ll})/u, c => c.toLowerCase())

/**
 * Puts the period prefix in front of a narrative, replacing a generated prefix that is already there.
 * Without a period, a generated prefix is removed and the rest is returned as it is.
 */
export const prefixNarrative = (t: Translate, text: string, period: string): string => {
  const body = stripPeriodPrefix(t, text ?? '').trimStart()
  if (!period) return body
  return body ? `${narrativePrefix(t, period)} ${lowerFirstWord(body)}` : narrativePrefix(t, period)
}

/**
 * Updates the period of a narrative after the period changed (quarter/year selection, linked LHA).
 * Only text that still starts with a generated prefix or with the template body is changed;
 * free text written by the user and an empty narrative are returned unchanged.
 */
export const refreshNarrativePeriod = (t: Translate, text: string, period: string): string => {
  const current = text ?? ''
  if (!current.trim()) return current
  const lead = (t(NARRATIVE_KEYS.templateBody).split('....')[0] ?? '').trim()
  const startsWithTemplate = !!lead && current.trimStart().startsWith(lead)
  if (!hasPeriodPrefix(t, current) && !startsWithTemplate) return current
  return prefixNarrative(t, current, period)
}

/**
 * Reports whether a narrative string is empty, still contains un-filled template placeholders ("....", "…."),
 * or follows the auto-generated SPI narrative pattern (so it should stay synchronized when data changes).
 */
export const isAutoSyncableNarrative = (t: Translate, text?: string | null): boolean => {
  const current = String(text ?? '').trim()
  if (!current) return true
  if (current.includes('....') || current.includes('….') || current.includes('…..')) return true
  const body = stripPeriodPrefix(t, current).trimStart()
  const lead = (t(NARRATIVE_KEYS.templateBody).split('....')[0] ?? '').trim()
  if (lead && body.startsWith(lead)) return true
  if (/^(?:SPI\s+telah\s+menerbitkan|SPI\s+has\s+issued|Executive\s+Summary\s+(?:Individual\s+)?untuk)\b/i.test(body)) {
    return true
  }
  return false
}
