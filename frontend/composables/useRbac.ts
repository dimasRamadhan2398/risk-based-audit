import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

/**
 * Role-Based Access Control composable.
 * F-04: Provides helpers to check user roles and permissions.
 */
export const useRbac = () => {
  const authStore = useAuthStore()

  const cleanStr = (r: string | UserRole) => String(r || '').toLowerCase().replace(/[\s_-]+/g, '')
  const normalize = (r: string | UserRole) => String(r || '').toLowerCase().trim()

  const getUserRoles = (): string[] => {
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

    // CAE / Chief Audit Executive / Executive
    const caeAliases = ['chiefauditexecutive', 'cae', 'executive', 'kepalaspi', 'headofskai']
    if (caeAliases.includes(t) && caeAliases.includes(u)) return true

    // Admin
    const adminAliases = ['admin', 'superadmin', 'administrator']
    if (adminAliases.includes(t) && adminAliases.includes(u)) return true

    // Audit Manager
    const managerAliases = ['auditmanager', 'manager', 'manageraudit']
    if (managerAliases.includes(t) && managerAliases.includes(u)) return true

    // Auditor
    const auditorAliases = ['auditor', 'auditstaff', 'staffaudit', 'leadauditor']
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
    const roles = (authStore.user?.roles ?? []).map(normalize)
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
      if (roles.includes(role)) return role
    }
    return authStore.user?.roles?.[0] ?? 'user'
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

  return {
    hasRole,
    hasAnyRole,
    hasAllRoles,
    isAdmin,
    isAuditor,
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
    primaryRole,
  }
}
