// Job Role API Composable
import type { JobRole } from '~/types/master'
import { getMasterServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuthStore } from '~/stores/auth'

export const useJobRoleApi = () => {
  const getBaseUrl = () => getMasterServiceBaseUrl()

  /**
   * Helper to get auth headers
   */
  const getAuthHeaders = (): Record<string, string> => {
    const headers: Record<string, string> = {}
    try {
      const authStore = useAuthStore()
      if (authStore.token) {
        headers['Authorization'] = `Bearer ${authStore.token}`
      }
    } catch {
      // Store might not be ready yet in SSR
    }
    return headers
  }

  /**
   * Get all job roles (for dropdowns)
   * GET /api/v1/job-roles
   */
  const getAllJobRoles = async (): Promise<JobRole[]> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/job-roles`, {
        method: 'GET',
        headers: getAuthHeaders()
      })

      const roles = Array.isArray(response.data)
        ? response.data
        : (response.data?.job_roles || response.job_roles || response.data?.data || [])
      return Array.isArray(roles) ? roles : []
    } catch (err) {
      console.error('Failed to fetch job roles:', err)
      return []
    }
  }

  /**
   * Get single job role by ID
   * GET /api/v1/job-roles/:id
   */
  const getJobRoleById = async (id: string): Promise<JobRole> => {
    const response = await $fetch<any>(`${getBaseUrl()}/job-roles/${id}`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    return response.data || response
  }

  return {
    getAllJobRoles,
    getJobRoleById
  }
}
