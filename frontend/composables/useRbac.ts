import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'

/**
 * Role-Based Access Control composable.
 * F-04: Provides helpers to check user roles and permissions.
 */
export const useRbac = () => {
  const authStore = useAuthStore()

  /**
   * Check if the current user has a specific role.
   */
  const hasRole = (role: UserRole | string): boolean => {
    if (!authStore.user?.roles) return false
    return authStore.user.roles.includes(role)
  }

  /**
   * Check if the current user has any of the provided roles.
   */
  const hasAnyRole = (roles: (UserRole | string)[]): boolean => {
    if (!authStore.user?.roles) return false
    return roles.some(role => authStore.user!.roles.includes(role))
  }

  /**
   * Check if the current user has all of the provided roles.
   */
  const hasAllRoles = (roles: (UserRole | string)[]): boolean => {
    if (!authStore.user?.roles) return false
    return roles.every(role => authStore.user!.roles.includes(role))
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
    const roles = authStore.user?.roles ?? []
    const priority = [
      UserRole.ADMIN,
      UserRole.CHIEF_AUDIT_EXECUTIVE,
      UserRole.AUDIT_MANAGER,
      UserRole.AUDIT_STAFF,
      UserRole.AUDITOR,
      UserRole.DEPARTMENT_HEAD,
      UserRole.AUDITEE,
      UserRole.VIEWER,
    ]
    for (const role of priority) {
      if (roles.includes(role)) return role
    }
    return roles[0] ?? 'user'
  })

  return {
    hasRole,
    hasAnyRole,
    hasAllRoles,
    isAdmin,
    isAuditor,
    canManageAudits,
    canViewRisks,
    primaryRole,
  }
}
