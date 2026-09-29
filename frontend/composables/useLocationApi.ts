// Location API Composable
import type { Location } from '~/types/master'
import { getMasterServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuthStore } from '~/stores/auth'

export const useLocationApi = () => {
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
   * Get all locations (for dropdowns)
   * GET /api/v1/locations
   */
  const getAllLocations = async (): Promise<Location[]> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/locations`, {
        method: 'GET',
        headers: getAuthHeaders()
      })

      const locations = Array.isArray(response.data)
        ? response.data
        : (response.data?.locations || response.locations || response.data?.data || [])
      return Array.isArray(locations) ? locations : []
    } catch (err) {
      console.error('Failed to fetch locations:', err)
      return []
    }
  }

  /**
   * Get single location by ID
   * GET /api/v1/locations/:id
   */
  const getLocationById = async (id: string): Promise<Location> => {
    const response = await $fetch<any>(`${getBaseUrl()}/locations/${id}`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    return response.data || response
  }

  return {
    getAllLocations,
    getLocationById
  }
}
