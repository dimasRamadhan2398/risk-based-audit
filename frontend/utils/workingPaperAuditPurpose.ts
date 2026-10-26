/**
 * Audit Purpose of a Working Paper header (F01).
 *
 * The backend stores `auditPurpose` as one text value (see WorkingPaperHeader). The value is kept
 * as that string, with one purpose per line ("\n"), and the form shows it as a numbered list.
 * Splitting and joining reuse the Activity Plan rules in utils/numberedList.ts.
 */
import { joinListItems, normalizeListItems, splitNumberedText } from '~/utils/numberedList'

/** Source of the purpose: the Assignment Letter fields the store reads. */
export interface AuditPurposeSource {
  auditPurpose?: string | null
  purposeList?: readonly unknown[] | null
}

/**
 * Items of the letter's purpose. A saved single `auditPurpose` string is split by line. Letters
 * created in the Assignment Letter form only carry `purposeList`, so each entry is one item.
 */
export const auditPurposeItemsFromLetter = (letter: AuditPurposeSource | null | undefined): string[] => {
  const single = splitNumberedText(letter?.auditPurpose)
  if (single.length) return single
  return normalizeListItems(letter?.purposeList)
}

/** Stored text for a letter's purpose (one item per line). */
export const resolveAuditPurposeText = (letter: AuditPurposeSource | null | undefined): string =>
  joinListItems(auditPurposeItemsFromLetter(letter))

/** Items shown in the form for a stored working paper value. */
export const auditPurposeItemsFromText = (text: string | null | undefined): string[] =>
  splitNumberedText(text)

/** Value sent to the backend for a list of items. */
export const auditPurposeTextFromItems = (items: readonly unknown[] | null | undefined): string =>
  joinListItems(items)
