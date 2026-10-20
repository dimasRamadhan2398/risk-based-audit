import type { AnnualAuditAttachment, AnnualPlanForm, AuditActivities } from '~/types/audit'

/** Number input value -> integer; "", null and NaN fall back to `fallback`. */
export const toCount = (value: unknown, fallback = 0): number => {
  const n = typeof value === 'number' ? value : parseInt(String(value ?? '').trim(), 10)
  return Number.isFinite(n) ? n : fallback
}

/**
 * Payload for one annual activity row (activities[] item).
 *
 * riskName, riskLevel and department (name) are server snapshots refreshed from the
 * Audit Universe and master data, so they are never sent: resending a stale copy
 * would overwrite the refreshed value. The row is identified by its id (kept on
 * edit) and its Audit Universe entity by auditUniverseId.
 *
 * Kode Penugasan (activityCode), Justifikasi, Tujuan and Fokus moved to the Audit
 * Activity Plan: they are not sent, and the server keeps any stored value.
 *
 * departmentId and auditUniverseId are only sent when set, so a row loaded from the
 * server without them (status change, revision) does not blank them out.
 */
export const buildAnnualActivityPayload = (activity: AuditActivities, index: number) => {
  const payload: Record<string, unknown> = {
    itemNumber: String(index + 1),
    name: activity.name || '',
    category: activity.category,
    timelineText: activity.timelineText || '',
    auditorCount: toCount(activity.auditorCount, 1),
    totalMandays: toCount(activity.totalMandays, 0),
    supervisorName: activity.supervisorName || ''
  }
  if (activity.id) payload.id = activity.id
  if (activity.departmentId) payload.departmentId = activity.departmentId
  if (activity.auditUniverseId) payload.auditUniverseId = activity.auditUniverseId
  return payload
}

/** Sum of Total Mandays over the activity rows. Plan-level mandays are no longer stored. */
export const sumActivityMandays = (activities: Pick<AuditActivities, 'totalMandays'>[] | null | undefined): number =>
  (activities || []).reduce((sum, a) => sum + toCount(a.totalMandays, 0), 0)

export interface AnnualPlanPayloadExtras {
  quarters: string[]
  attachments: AnnualAuditAttachment[]
}

/**
 * Body for POST/PUT /annual-audit-plans. Plan header: title (Judul PKAT), year. Rows:
 * activities[]. Plan-level auditor, days, supervisor and total mandays are not sent: the
 * rows are the only entry point. Jenis Kegiatan Audit (activityType) belongs to the Audit
 * Activity Plan and is not sent; PUT is partial, so the stored value is untouched.
 */
export const buildAnnualPlanPayload = (form: AnnualPlanForm, extras: AnnualPlanPayloadExtras) => ({
  title: (form.title || '').trim(),
  version: form.version || 'v1.0',
  revisionHistory: form.revisionHistory || [],
  activities: form.activities.map((activity, index) => buildAnnualActivityPayload(activity, index)),
  status: form.status,
  selectedMonths: [...form.selectedMonths].sort((a, b) => a - b),
  quarters: extras.quarters,
  notes: form.notes || '',
  year: toCount(form.year, 2026),
  attachmentCategory: form.attachmentCategory,
  attachments: extras.attachments,
  attachmentUploadedBy: form.attachmentUploadedBy,
  attachmentUploadDate: form.attachmentUploadDate,
  isActive: form.isActive
})
