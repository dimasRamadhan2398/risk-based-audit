/**
 * Numbered lists of the Formulir Rencana Kegiatan Audit (Justifikasi, Tujuan, Fokus).
 *
 * Mirrors the audit-service normaliser (services/activity_plan_form): trim each item, strip a
 * leading "1." or "1)" and drop empty items. The server stores the items joined by "\n".
 */

/** Limits enforced by the server per list. */
export const NUMBERED_LIST_MAX_ITEMS = 20
export const NUMBERED_LIST_MAX_ITEM_LENGTH = 1000

const LEADING_NUMBER = /^\s*\d+[.)]\s*/

/** Trims an item and strips a leading "1." / "1)". */
export const normalizeListItem = (item: unknown): string =>
  String(item ?? '').trim().replace(LEADING_NUMBER, '').trim()

/** Normalises every item and drops the empty ones. */
export const normalizeListItems = (items: readonly unknown[] | null | undefined): string[] =>
  (items || []).map(normalizeListItem).filter(Boolean)

/** Splits text (pasted, or stored joined by "\n") into normalised items. */
export const splitNumberedText = (text: string | null | undefined): string[] =>
  normalizeListItems(String(text ?? '').split(/\r?\n/))

/** Joins normalised items the way the server stores them. */
export const joinListItems = (items: readonly unknown[] | null | undefined): string =>
  normalizeListItems(items).join('\n')

/**
 * List value from the API: an array is normalised, a string (legacy, joined by "\n") is split.
 */
export const toListItems = (value: unknown): string[] => {
  if (Array.isArray(value)) return normalizeListItems(value)
  if (typeof value === 'string') return splitNumberedText(value)
  return []
}

/**
 * Pasting into item `index`: multi-line text becomes several items. The pasted items replace an
 * empty item, or are inserted after a filled one. Returns null when the text is a single line, so
 * the normal paste happens.
 */
export const applyPastedText = (items: readonly string[], index: number, text: string): string[] | null => {
  if (!/\r?\n/.test(text.trim())) return null
  const pasted = splitNumberedText(text)
  if (!pasted.length) return null
  const next = [...items]
  const current = next[index] ?? ''
  if (current.trim()) next.splice(index + 1, 0, ...pasted)
  else next.splice(index, 1, ...pasted)
  return next
}

/** Problems of one list against the server limits. */
export interface NumberedListCheck {
  /** No non-empty item. */
  empty: boolean
  /** More than NUMBERED_LIST_MAX_ITEMS items. */
  tooMany: boolean
  /** At least one item longer than NUMBERED_LIST_MAX_ITEM_LENGTH. */
  tooLong: boolean
}

export const checkNumberedList = (items: readonly unknown[] | null | undefined): NumberedListCheck => {
  const normalized = normalizeListItems(items)
  return {
    empty: normalized.length === 0,
    tooMany: normalized.length > NUMBERED_LIST_MAX_ITEMS,
    tooLong: normalized.some(item => item.length > NUMBERED_LIST_MAX_ITEM_LENGTH)
  }
}
