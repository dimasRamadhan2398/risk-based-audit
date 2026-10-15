import type { ActionTakenReport } from '~/types/audit'
import { useRbac } from '~/composables/useRbac'
import {
  atrCanAssignStatus,
  atrCanCancelStatus,
  atrCanEditPlanStatus,
  atrCanEditStatus,
  atrCanRemindStatus,
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
// The server's name for the reminder action (not part of AtrAvailableActions).
const SERVER_ACTION_REMIND = 'remind'

/** The server's allowed_actions, when present, must list the action; without the list the local rules decide. */
const serverAllows = (item: ActionTakenReport, name: string) =>
  !Array.isArray(item.allowed_actions) || item.allowed_actions.includes(name)

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

  /**
   * Which edit-form fields the caller may change. Due date: assign rights (auditor/manager/admin) on an open item.
   * Action plan and progress: PIC rights (or admin) on a PLANNED/IN_PROGRESS item. Both also need the server's
   * allowed_actions to list the matching action, when the list is present.
   */
  const editableFields = (item: ActionTakenReport | null | undefined): { dueDate: boolean, actionPlan: boolean, progress: boolean } => {
    if (!item || !atrCanEditStatus(item.status)) return { dueDate: false, actionPlan: false, progress: false }
    const dueDate = rbac.canAssignAtr.value && serverAllows(item, SERVER_ACTION.assign)
    const plan = rbac.canWorkOnAtr(item) && atrCanEditPlanStatus(item.status) && serverAllows(item, SERVER_ACTION.actionPlan)
    return { dueDate, actionPlan: plan, progress: plan }
  }

  /** The edit form is offered when at least one of its fields can be changed. */
  const canEdit = (item: ActionTakenReport | null | undefined): boolean => {
    const fields = editableFields(item)
    return fields.dueDate || fields.actionPlan
  }

  /** Send reminder (auditee and department head): any open item, not cancelled; assign rights (auditor/manager/admin). */
  const canRemind = (item: ActionTakenReport | null | undefined): boolean => {
    if (!item || !atrCanRemindStatus(item.status)) return false
    return rbac.canAssignAtr.value && serverAllows(item, SERVER_ACTION_REMIND)
  }

  return { availableActions, canEdit, editableFields, canRemind }
}
