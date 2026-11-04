// Job Role API Composable
import type { JobRole, CreateJobRoleRequest, UpdateJobRoleRequest } from '~/types/master'
import { FALLBACK_ACCESS_ROLES, normalizeAccessRoles } from '~/utils/jobRole'
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

  /**
   * Create job role
   * POST /api/v1/job-roles
   */
  const createJobRole = async (payload: CreateJobRoleRequest): Promise<JobRole> => {
    const response = await $fetch<any>(`${getBaseUrl()}/job-roles`, {
      method: 'POST',
      body: payload
    })
    return response.data || response
  }

  /**
   * Update job role (partial)
   * PUT /api/v1/job-roles/:id
   */
  const updateJobRole = async (id: string, payload: UpdateJobRoleRequest): Promise<JobRole> => {
    const response = await $fetch<any>(`${getBaseUrl()}/job-roles/${id}`, {
      method: 'PUT',
      body: payload
    })
    return response.data || response
  }

  /**
   * Delete job role
   * DELETE /api/v1/job-roles/:id
   */
  const deleteJobRole = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/job-roles/${id}`, {
      method: 'DELETE'
    })
  }

  /**
   * Allowed default access role names.
   * GET /api/v1/job-roles/access-roles, falling back to a constant list if it fails or is empty.
   */
  const getAccessRoles = async (): Promise<string[]> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/job-roles/access-roles`, { method: 'GET' })
      const roles = normalizeAccessRoles(response)
      return roles.length ? roles : [...FALLBACK_ACCESS_ROLES]
    } catch {
      return [...FALLBACK_ACCESS_ROLES]
    }
  }

  return {
    getAllJobRoles,
    getJobRoleById,
    createJobRole,
    updateJobRole,
    deleteJobRole,
    getAccessRoles
  }
}
