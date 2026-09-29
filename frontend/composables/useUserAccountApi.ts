// User Account API Composable
// Admin actions on login accounts (auth-service) linked to master-data employees
import { getAuthServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuthStore } from '~/stores/auth'
import { extractErrorMessage } from '~/utils/error'

export interface LinkedUserAccount {
  id: string
  employee_id: string
  username: string
  email: string
  full_name: string
  roles: string[]
  is_active: boolean
  must_change_password: boolean
}

export interface AdminResetPasswordResult {
  user_id: string
  username: string
  temporary_password: string
}

export const useUserAccountApi = () => {
  const getBaseUrl = () => getAuthServiceBaseUrl()

  const getAuthHeaders = (): Record<string, string> => {
    const authStore = useAuthStore()
    return authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {}
  }

  /**
   * Find the login account linked to an employee. Returns null when the
   * employee has no account.
   * GET /api/v1/users/lookup?employee_code=&email=
   */
  const findAccountByEmployee = async (employeeCode: string, email: string): Promise<LinkedUserAccount | null> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/users/lookup`, {
        method: 'GET',
        headers: getAuthHeaders(),
        query: { employee_code: employeeCode, email }
      })
      return response.data ?? response
    } catch (err: any) {
      const status = err?.status ?? err?.statusCode ?? err?.response?.status
      if (status === 404) return null
      throw new Error(extractErrorMessage(err, 'Failed to look up user account'))
    }
  }

  /**
   * Reset a user's password to a temporary one (ADMIN only).
   * POST /api/v1/users/:id/reset-password
   */
  const resetPassword = async (userId: string): Promise<AdminResetPasswordResult> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/users/${userId}/reset-password`, {
        method: 'POST',
        headers: getAuthHeaders()
      })
      return response.data ?? response
    } catch (err: any) {
      throw new Error(extractErrorMessage(err, 'Failed to reset password'))
    }
  }

  return {
    findAccountByEmployee,
    resetPassword
  }
}
