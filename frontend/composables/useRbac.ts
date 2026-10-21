import { ref, computed } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

/**
 * Role-Based Access Control composable.
 * F-04: Provides helpers to check user roles and permissions.
 * Role names follow the backend seeder (auth-service): ADMIN, AUDITOR, EXECUTIVE, AUDITEE, VIEWER, DEPARTMENT_HEAD.
 */
export const useRbac = () => {
  const authStore = useAuthStore()

  const cleanStr = (r: string | UserRole) => String(r || '').toLowerCase().replace(/[\s_-]+/g, '')
  const normalize = (r: string | UserRole) => String(r || '').toLowerCase().trim()

  const getCurrentUser = (): any => {
    let u: any = authStore.user
    if (!u) {
      try {
        const userCookie = useCookie<any>('auth-user')
        if (userCookie.value) {
          u = typeof userCookie.value === 'string' ? JSON.parse(userCookie.value) : userCookie.value
        }
      } catch {}
      if (!u && typeof window !== 'undefined') {
        try {
          const stored = localStorage.getItem('auth-user') || localStorage.getItem('user')
          if (stored) u = JSON.parse(stored)
        } catch {}
      }
    }
    return u || null
  }

  /** Auth user id of the logged-in user ('' when unknown). */
  const getCurrentUserId = (): string => {
    const id = getCurrentUser()?.id
    return id ? String(id) : ''
  }

  const getUserRoles = (): string[] => {
    const u = getCurrentUser()
    if (!u) return []

    const roles: string[] = []
    if (Array.isArray(u.roles)) roles.push(...u.roles)
    if (u.role) roles.push(u.role)
    if (u.position) roles.push(u.position)
    return roles
  }

  const matchesRole = (userRole: string, targetRole: string): boolean => {
    const u = cleanStr(userRole)
    const t = cleanStr(targetRole)
    if (!u || !t) return false
    if (u === t) return true

    // EXECUTIVE is the seeded role that acts as the Chief Audit Executive (CAE); the backend aliases mirror this.
    const executiveAliases = ['executive', 'chiefauditexecutive', 'cae']
    if (executiveAliases.includes(t) && executiveAliases.includes(u)) return true

    // Admin
    const adminAliases = ['admin', 'superadmin', 'administrator']
    if (adminAliases.includes(t) && adminAliases.includes(u)) return true

    // Auditor (backend alias AUDITSTAFF -> AUDITOR)
    const auditorAliases = ['auditor', 'auditstaff']
    if (auditorAliases.includes(t) && auditorAliases.includes(u)) return true

    return false
  }

  /**
   * Check if the current user has a specific role (case-insensitive).
   */
  const hasRole = (role: UserRole | string): boolean => {
    const userRoles = getUserRoles()
    if (!userRoles.length) return false
    return userRoles.some(r => matchesRole(r, role))
  }

  /**
   * Check if the current user has any of the provided roles (case-insensitive).
   */
  const hasAnyRole = (roles: (UserRole | string)[]): boolean => {
    const userRoles = getUserRoles()
    if (!userRoles.length) return false
    return roles.some(target => userRoles.some(r => matchesRole(r, target)))
  }

  /**
   * Check if the current user has all of the provided roles (case-insensitive).
   */
  const hasAllRoles = (roles: (UserRole | string)[]): boolean => {
    const userRoles = getUserRoles()
    if (!userRoles.length) return false
    return userRoles.every(target => userRoles.some(r => matchesRole(r, target)))
  }

  /**
   * Check if current user is Admin or EXECUTIVE (CAE) for Executive Summary
   */
  const canReviewExecutiveSummary = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.EXECUTIVE,
  ]))

  /**
   * Check if the current user is an admin.
   */
  const isAdmin = computed(() => hasRole(UserRole.ADMIN))

  /**
   * Check if the user is an auditor or higher.
   */
  const isAuditor = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.AUDITOR,
    UserRole.EXECUTIVE,
  ]))

  /**
   * Check if user can access audit management features.
   */
  const canManageAudits = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.EXECUTIVE,
  ]))

  /**
   * Check if user can view risk data.
   */
  const canViewRisks = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.AUDITOR,
    UserRole.EXECUTIVE,
    UserRole.DEPARTMENT_HEAD,
    UserRole.VIEWER,
  ]))

  /**
   * Get the user's primary display role (highest privilege).
   */
  const primaryRole = computed((): string => {
    const roles = getUserRoles().map(normalize)
    const priority = [
      UserRole.ADMIN,
      UserRole.EXECUTIVE,
      UserRole.AUDITOR,
      UserRole.DEPARTMENT_HEAD,
      UserRole.AUDITEE,
      UserRole.VIEWER,
    ].map(normalize)
    for (const role of priority) {
      if (roles.includes(role)) return role
    }
    return roles[0] ?? 'user'
  })

  /**
   * Module permission helpers matching Settings RBAC matrix
   */
  const canManageCharter = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canManageGuideline = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canManageSop = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canEditRiskAppetite = computed(() => isAdmin.value)
  const canEditRiskFactors = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canEditAuditUniverse = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canManageStrategicPlan = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canManageAnnualPlan = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canManageAssignmentLetter = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  const canImportPlanDocs = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))

  /**
   * Action Taken Report (ATR). The backend enforces the same rules; these only decide what the UI shows.
   * Permission keys (audit-service ATR role matrix, mirrored in the auth-service seeder):
   *  view_action_taken_report (own items), view_all_action_taken_report, assign_action_taken_report,
   *  update_action_taken_report (PIC: action plan, evidence, submit), review_action_taken_report,
   *  cancel_action_taken_report. The user object carries roles only, so the checks below are role-based.
   *  ADMIN: all | EXECUTIVE (acts as CAE): view all, review, cancel
   *  AUDITOR: view all, assign, review | AUDITEE/PIC: own items, action plan, evidence, submit
   * Per item, useAtrActions also requires the action in the server's `allowed_actions` when present.
   */
  const canViewAllAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE, UserRole.AUDITOR]))
  const canAssignAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDITOR]))
  const canReviewAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE, UserRole.AUDITOR]))
  const canCancelAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.EXECUTIVE]))
  /** True when the item is assigned to the logged-in user. */
  const isAtrPic = (item: { pic_user_id?: string | null } | null | undefined): boolean => {
    const me = getCurrentUserId()
    return !!me && !!item?.pic_user_id && String(item.pic_user_id) === me
  }
  /** The assigned PIC fills the action plan, uploads evidence and submits; admin may act on any item. */
  const canWorkOnAtr = (item: { pic_user_id?: string | null } | null | undefined): boolean =>
    isAtrPic(item) || (isAdmin.value && !!item?.pic_user_id)

  /**
   * Active Role View Mode (used for preview & testing by admins/managers).
   */
  const activeRoleView = ref<string | null>(null)

  const setRoleView = (view: string | null) => {
    activeRoleView.value = view
  }

  /**
   * Check if current user is strictly an Auditor (not Admin or Manager).
   */
  const isPureAuditor = computed((): boolean => {
    const userRoles = getUserRoles()
    const hasAuditor = userRoles.some(r => {
      const c = cleanStr(r)
      return c === 'auditor' || c === 'auditstaff' || c === 'staffaudit' || c === 'leadauditor'
    })
    const isHigher = hasAnyRole([
      UserRole.ADMIN,
      UserRole.CHIEF_AUDIT_EXECUTIVE,
      UserRole.AUDIT_MANAGER,
    ])
    return hasAuditor && !isHigher
  })

  /**
   * Evaluates if Auditor View mode should be active.
   * - Pure Auditor: ALWAYS true.
   * - Admin/Manager: ONLY true if explicitly previewing via activeRoleView === 'auditor'.
   */
  const isAuditorView = computed((): boolean => {
    if (activeRoleView.value === 'auditor') return true
    if (activeRoleView.value && activeRoleView.value !== 'auditor') return false
    return isPureAuditor.value || primaryRole.value === 'auditor' || primaryRole.value === 'audit_staff'
  })

  /**
   * Check if current user is strictly an Auditee (AUDITEE and no higher or auditor role).
   * DEPARTMENT_HEAD is not included on purpose: it keeps the admin view (spec D4).
   */
  const isPureAuditee = computed((): boolean => {
    const userRoles = getUserRoles()
    const hasAuditee = userRoles.some(r => cleanStr(r) === 'auditee')
    if (!hasAuditee) return false
    const isHigher = hasAnyRole([
      UserRole.ADMIN,
      UserRole.EXECUTIVE,
      UserRole.AUDITOR,
      UserRole.DEPARTMENT_HEAD,
    ])
    return !isHigher && !isPureAuditor.value
  })

  /**
   * Evaluates if Auditee View mode should be active.
   * - Pure Auditee: ALWAYS true.
   * - Others: ONLY true if previewing via activeRoleView === 'auditee'.
   */
  const isAuditeeView = computed((): boolean => {
    if (activeRoleView.value === 'auditee') return true
    if (activeRoleView.value) return false
    return isPureAuditee.value
  })

  /**
   * Authority to publish/authorize LHA report (Row 9-1: FULL for Admin/Manager, DRAFT for Auditor).
   * Per guide: Auditor does NOT have the publish/authorization button.
   */
  const canPublishLha = computed((): boolean => {
    if (isAuditorView.value) return false
    return hasAnyRole([UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER])
  })

  return {
    hasRole,
    hasAnyRole,
    hasAllRoles,
    isAdmin,
    isAuditor,
    isPureAuditor,
    isAuditorView,
    isPureAuditee,
    isAuditeeView,
    activeRoleView,
    setRoleView,
    canPublishLha,
    canManageAudits,
    canViewRisks,
    canManageCharter,
    canManageGuideline,
    canManageSop,
    canEditRiskAppetite,
    canEditRiskFactors,
    canEditAuditUniverse,
    canManageStrategicPlan,
    canManageAnnualPlan,
    canManageAssignmentLetter,
    canImportPlanDocs,
    canReviewExecutiveSummary,
    canViewAllAtr,
    canAssignAtr,
    canReviewAtr,
    canCancelAtr,
    isAtrPic,
    canWorkOnAtr,
    getCurrentUserId,
    primaryRole,
  }
}
