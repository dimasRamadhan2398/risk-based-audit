import { ref, computed } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

// Module-level singleton to share active preview role across all components and pages
const activeRoleView = ref<string | null>(null)

/**
 * Role-Based Access Control composable.
 * F-04: Provides helpers to check user roles and permissions.
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
    const addRole = (val: any) => {
      if (!val) return
      if (typeof val === 'string') {
        roles.push(val)
      } else if (typeof val === 'object') {
        if (val.name) roles.push(String(val.name))
        if (val.role) roles.push(String(val.role))
        if (val.code) roles.push(String(val.code))
        if (val.title) roles.push(String(val.title))
        if (val.position) roles.push(String(val.position))
      }
    }

    if (Array.isArray(u.roles)) {
      u.roles.forEach(addRole)
    }
    addRole(u.role)
    addRole(u.position)
    return roles
  }

  const matchesRole = (userRole: string, targetRole: string): boolean => {
    const u = cleanStr(userRole)
    const t = cleanStr(targetRole)
    if (!u || !t) return false
    if (u === t) return true

    // Target is Admin
    if (t === 'admin' || t === 'superadmin' || t === 'administrator') {
      return u === 'admin' || u === 'superadmin' || u === 'administrator' || u.includes('admin')
    }

    // Target is CAE / Executive / Viewer
    if (t === 'chiefauditexecutive' || t === 'cae' || t === 'executive' || t === 'viewer' || t === 'direksi' || t === 'komiteaudit' || t === 'kepalaspi' || t === 'headofskai') {
      if (u.includes('admin')) return false
      const execKeywords = ['chiefauditexecutive', 'cae', 'executive', 'kepalaspi', 'headofskai', 'viewer', 'komiteaudit', 'direksi', 'board', 'boc', 'bod', 'komisaris']
      return execKeywords.some(kw => u === kw || u.includes(kw))
    }

    // Target is Audit Manager
    if (t === 'auditmanager' || t === 'manager' || t === 'manageraudit') {
      return (u.includes('manager') || u.includes('manageraudit')) && !u.includes('admin')
    }

    // Target is Auditor
    if (t === 'auditor' || t === 'auditstaff' || t === 'staffaudit' || t === 'leadauditor') {
      if (u.includes('manager') || u.includes('chief') || u.includes('cae') || u.includes('admin') || u.includes('auditee') || u.startsWith('pic')) {
        return false
      }
      return u.includes('auditor') || u.includes('auditstaff') || u.includes('staffaudit')
    }

    // Target is Auditee / PIC Unit Kerja
    if (t === 'auditee' || t === 'audite' || t === 'pic' || t === 'picunit' || t === 'departmenthead' || t === 'unitkerja' || t === 'picunitkerja') {
      if (u.includes('chief') || u.includes('executive') || u.includes('admin') || u.includes('manager') || u.includes('auditor')) {
        return false
      }
      const auditeeKeywords = ['auditee', 'audite', 'pic', 'picunit', 'departmenthead', 'depthead', 'unitkerja', 'picunitkerja', 'headofdepartment']
      return auditeeKeywords.some(kw => u === kw || (kw === 'pic' ? u.startsWith('pic') : u.includes(kw)))
    }

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
   * Check if current user is Admin, CAE, or Audit Manager for Executive Summary
   */
  const canReviewExecutiveSummary = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.CHIEF_AUDIT_EXECUTIVE,
    UserRole.AUDIT_MANAGER,
    'admin',
    'chief_audit_executive',
    'audit_manager',
    'cae',
    'executive',
    'head_of_skai',
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
    UserRole.AUDIT_STAFF,
    UserRole.AUDIT_MANAGER,
    UserRole.CHIEF_AUDIT_EXECUTIVE,
  ]))

  /**
   * Check if the user is an auditee / PIC unit kerja.
   */
  const isAuditee = computed(() => hasAnyRole([
    UserRole.AUDITEE,
    UserRole.DEPARTMENT_HEAD,
    'auditee',
    'pic',
    'picunit',
    'departmenthead',
    'unitkerja'
  ]))

  /**
   * Check if the user is an executive (CAE, Viewer, Direksi, Komite Audit).
   */
  const isExecutive = computed(() => hasAnyRole([
    UserRole.CHIEF_AUDIT_EXECUTIVE,
    UserRole.VIEWER,
    UserRole.AUDIT_MANAGER,
    'cae',
    'viewer',
    'executive',
    'direksi',
    'komite_audit'
  ]) && !isAdmin.value)

  /**
   * Check if user can access audit management features.
   */
  const canManageAudits = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.AUDIT_MANAGER,
    UserRole.CHIEF_AUDIT_EXECUTIVE,
  ]))

  /**
   * Check if user can view risk data.
   */
  const canViewRisks = computed(() => hasAnyRole([
    UserRole.ADMIN,
    UserRole.AUDITOR,
    UserRole.AUDIT_STAFF,
    UserRole.AUDIT_MANAGER,
    UserRole.CHIEF_AUDIT_EXECUTIVE,
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
      UserRole.CHIEF_AUDIT_EXECUTIVE,
      UserRole.AUDIT_MANAGER,
      UserRole.AUDIT_STAFF,
      UserRole.AUDITOR,
      UserRole.DEPARTMENT_HEAD,
      UserRole.AUDITEE,
      UserRole.VIEWER,
    ].map(normalize)
    for (const role of priority) {
      if (roles.some(r => matchesRole(r, role))) return role
    }
    return roles[0] ?? 'auditor'
  })

  /**
   * Module permission helpers matching Settings RBAC matrix
   */
  const canManageCharter = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER]))
  const canManageGuideline = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER]))
  const canManageSop = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER]))
  const canEditRiskAppetite = computed(() => isAdmin.value)
  const canEditRiskFactors = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))
  const canEditAuditUniverse = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))
  const canManageStrategicPlan = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))
  const canManageAnnualPlan = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))
  const canManageAssignmentLetter = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))
  const canImportPlanDocs = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.CHIEF_AUDIT_EXECUTIVE]))

  /**
   * Action Taken Report (ATR). The backend enforces the same rules; these only decide what the UI shows.
   */
  const canViewAllAtr = computed(() => hasAnyRole([
    UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER, UserRole.AUDITOR, UserRole.AUDIT_STAFF,
  ]))
  const canAssignAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.AUDIT_MANAGER, UserRole.AUDITOR, UserRole.AUDIT_STAFF]))
  const canReviewAtr = computed(() => hasAnyRole([
    UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER, UserRole.AUDITOR, UserRole.AUDIT_STAFF,
  ]))
  const canCancelAtr = computed(() => hasAnyRole([UserRole.ADMIN, UserRole.CHIEF_AUDIT_EXECUTIVE, UserRole.AUDIT_MANAGER]))
  /** True when the item is assigned to the logged-in user. */
  const isAtrPic = (item: { pic_user_id?: string | null } | null | undefined): boolean => {
    const me = getCurrentUserId()
    return !!me && !!item?.pic_user_id && String(item.pic_user_id) === me
  }
  /** The assigned PIC fills the action plan, uploads evidence and submits; admin may act on any item. */
  const canWorkOnAtr = (item: { pic_user_id?: string | null } | null | undefined): boolean =>
    isAtrPic(item) || (isAdmin.value && !!item?.pic_user_id)

  const setRoleView = (_view: string | null) => {
    activeRoleView.value = null
  }

  /**
   * Check if current user is strictly an Auditor (not Admin, Manager, or CAE).
   */
  const isPureAuditor = computed((): boolean => {
    const userRoles = getUserRoles()
    const hasAuditor = userRoles.some(r => matchesRole(r, 'auditor'))
    const isHigher = hasAnyRole([
      UserRole.ADMIN,
      UserRole.CHIEF_AUDIT_EXECUTIVE,
      UserRole.AUDIT_MANAGER,
    ])
    return hasAuditor && !isHigher
  })

  /**
   * Check if current user is strictly an Auditee / Department Head (not Internal Audit or Admin).
   */
  const isPureAuditee = computed((): boolean => {
    const userRoles = getUserRoles()
    const hasAuditee = userRoles.some(r => matchesRole(r, 'auditee'))
    const isInternalAuditOrAdmin = hasAnyRole([
      UserRole.ADMIN,
      UserRole.CHIEF_AUDIT_EXECUTIVE,
      UserRole.AUDIT_MANAGER,
      UserRole.AUDITOR,
      UserRole.AUDIT_STAFF,
    ])
    return hasAuditee && !isInternalAuditOrAdmin
  })

  /**
   * Evaluates if Auditor View mode should be active.
   */
  const isAuditorView = computed((): boolean => {
    return isPureAuditor.value || matchesRole(primaryRole.value, 'auditor')
  })

  /**
   * Evaluates if Auditee View mode should be active.
   */
  const isAuditeeView = computed((): boolean => {
    return isPureAuditee.value || matchesRole(primaryRole.value, 'auditee')
  })

  /**
   * Evaluates if Executive View mode should be active.
   */
  const isExecutiveView = computed((): boolean => {
    return isExecutive.value || matchesRole(primaryRole.value, 'chief_audit_executive') || matchesRole(primaryRole.value, 'viewer')
  })

  /**
   * Evaluates if Admin View mode should be active.
   */
  const isAdminView = computed((): boolean => {
    return isAdmin.value
  })

  /**
   * Determines which dedicated dashboard role to display:
   * 'admin' | 'auditor' | 'auditee' | 'executive'
   * Strictly evaluates the user's authentic role; roles only access their own dashboard.
   */
  const currentDashboardRole = computed<'admin' | 'auditor' | 'auditee' | 'executive'>(() => {
    if (isAdmin.value) return 'admin'
    if (isExecutive.value) return 'executive'
    if (isPureAuditee.value || (!isPureAuditor.value && isAuditee.value)) return 'auditee'
    if (isPureAuditor.value || isAuditor.value) return 'auditor'

    // Fallback based on primary role match
    if (matchesRole(primaryRole.value, 'admin')) return 'admin'
    if (matchesRole(primaryRole.value, 'executive') || matchesRole(primaryRole.value, 'viewer')) return 'executive'
    if (matchesRole(primaryRole.value, 'auditee')) return 'auditee'
    if (matchesRole(primaryRole.value, 'auditor')) return 'auditor'

    return 'admin'
  })

  /**
   * Authority to publish/authorize LHA report (Row 9-1: FULL for Admin/Manager, DRAFT for Auditor).
   * Per guide: Auditor does NOT have the publish/authorization button.
   */
  const canPublishLha = computed((): boolean => {
    if (isAuditorView.value || isAuditeeView.value) return false
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
    isAuditee,
    isPureAuditee,
    isAuditeeView,
    isExecutive,
    isExecutiveView,
    isAdminView,
    currentDashboardRole,
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
