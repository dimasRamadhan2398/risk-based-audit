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
  /** Assign or reassign PIC and due date (auditor/manager/admin): any open status. */
  assign: boolean
  /** Approve or reject (auditor/manager/CAE/admin, not on an item they are the PIC of): PENDING_REVIEW. */
  review: boolean
  /** Cancel (manager/CAE/admin): any open status. */
  cancel: boolean
}

const NONE: AtrAvailableActions = { actionPlan: false, assign: false, review: false, cancel: false }

// Names in the item's `allowed_actions` (sent by the server) that unlock each UI action.
const SERVER_ACTION: Record<keyof AtrAvailableActions, string> = {
  actionPlan: 'update_action_plan',
  assign: 'assign',
  review: 'review',
  cancel: 'cancel'
}

/**
 * What the logged-in user can do with an ATR: role (useRbac) AND status. When the item carries the
 * server's `allowed_actions`, an action must also be listed there, so the UI never offers a button the
 * backend would refuse. The backend enforces the same rules; this only decides what the UI offers.
 */
export const useAtrActions = () => {
  const rbac = useRbac()

  const availableActions = (item: ActionTakenReport | null | undefined): AtrAvailableActions => {
    if (!item) return NONE
    const local: AtrAvailableActions = {
      actionPlan: rbac.canWorkOnAtr(item) && atrCanEditPlanStatus(item.status),
      assign: rbac.canAssignAtr.value && atrCanAssignStatus(item.status),
      // Segregation of duties: a reviewer does not review their own follow-up (admins excepted).
      review: rbac.canReviewAtr.value && (rbac.isAdmin.value || !rbac.isAtrPic(item)) && atrCanReviewStatus(item.status),
      cancel: rbac.canCancelAtr.value && atrCanCancelStatus(item.status)
    }
    const server = item.allowed_actions
    if (!Array.isArray(server)) return local
    return {
      actionPlan: local.actionPlan && server.includes(SERVER_ACTION.actionPlan),
      assign: local.assign && server.includes(SERVER_ACTION.assign),
      review: local.review && server.includes(SERVER_ACTION.review),
      cancel: local.cancel && server.includes(SERVER_ACTION.cancel)
    }
  }

  return { availableActions }
}
