import type { ActionTakenReport } from '~/types/audit'
import { useRbac } from '~/composables/useRbac'
import {
  atrCanAssignStatus,
  atrCanCancelStatus,
  atrCanEditPlanStatus,
  atrCanReviewStatus
} from '~/utils/actionTakenReport'

export interface AtrAvailableActions {
  /** Open the action plan form (assigned PIC, or admin): PLANNED / IN_PROGRESS. */
  actionPlan: boolean
  /** Assign or reassign PIC and due date (auditor/manager/admin): PLANNED / IN_PROGRESS. */
  assign: boolean
  /** Approve or reject (auditor/manager/CAE/admin): PENDING_REVIEW. */
  review: boolean
  /** Cancel (manager/CAE/admin): any open status. */
  cancel: boolean
}

const NONE: AtrAvailableActions = { actionPlan: false, assign: false, review: false, cancel: false }

/**
 * What the logged-in user can do with an ATR: role (useRbac) AND status. The backend enforces
 * the same rules; this only decides which buttons and modals the UI offers.
 */
export const useAtrActions = () => {
  const rbac = useRbac()

  const availableActions = (item: ActionTakenReport | null | undefined): AtrAvailableActions => {
    if (!item) return NONE
    return {
      actionPlan: rbac.canWorkOnAtr(item) && atrCanEditPlanStatus(item.status),
      assign: rbac.canAssignAtr.value && atrCanAssignStatus(item.status),
      review: rbac.canReviewAtr.value && atrCanReviewStatus(item.status),
      cancel: rbac.canCancelAtr.value && atrCanCancelStatus(item.status)
    }
  }

  return { availableActions }
}
