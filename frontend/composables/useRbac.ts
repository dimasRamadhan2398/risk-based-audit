import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

/**
 * Role-Based Access Control composable.
 * F-04: Provides helpers to check user roles and permissions.
 */
export const useRbac = () => {
  const authStore = useAuthStore()

  const normalize = (r: string | UserRole) => String(r || '').toLowerCase().trim()

  /**
   * Check if the current user has a specific role (case-insensitive).
   */
  const hasRole = (role: UserRole | string): boolean => {
    if (!authStore.user?.roles || !Array.isArray(authStore.user.roles)) return false
    const target = normalize(role)
    return authStore.user.roles.some(r => normalize(r) === target)
  }

  /**
   * Check if the current user has any of the provided roles (case-insensitive).
   */
  const hasAnyRole = (roles: (UserRole | string)[]): boolean => {
    if (!authStore.user?.roles || !Array.isArray(authStore.user.roles)) return false
    const targets = roles.map(normalize)
    return authStore.user.roles.some(r => targets.includes(normalize(r)))
  }

  /**
   * Check if the current user has all of the provided roles (case-insensitive).
   */
  const hasAllRoles = (roles: (UserRole | string)[]): boolean => {
    if (!authStore.user?.roles || !Array.isArray(authStore.user.roles)) return false
    const userRoles = authStore.user.roles.map(normalize)
    return roles.every(role => userRoles.includes(normalize(role)))
  }

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
    canEditRiskAppetite,
    canEditRiskFactors,
    canEditAuditUniverse,
    canManageStrategicPlan,
    canManageAnnualPlan,
    canManageAssignmentLetter,
    canImportPlanDocs,
    primaryRole,
  }
}
