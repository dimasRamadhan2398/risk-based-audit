// Location API Composable
import type { Location, CreateLocationRequest, UpdateLocationRequest } from '~/types/master'
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
   * Get all locations
   * GET /api/v1/locations
   *
   * Throws on failure — callers that manage master data need to tell "the
   * service is down" apart from "there are no branches yet".
   */
  const getLocations = async (): Promise<Location[]> => {
    const response = await $fetch<any>(`${getBaseUrl()}/locations`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    const locations = Array.isArray(response.data)
      ? response.data
      : (response.data?.locations || response.locations || response.data?.data || [])
    return Array.isArray(locations) ? locations : []
  }

  /**
   * Get all locations (for dropdowns) — never throws, returns [] on failure.
   */
  const getAllLocations = async (): Promise<Location[]> => {
    try {
      return await getLocations()
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

  /**
   * Create new location
   * POST /api/v1/locations
   */
  const createLocation = async (payload: CreateLocationRequest): Promise<Location> => {
    const response = await $fetch<any>(`${getBaseUrl()}/locations`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: payload
    })

    return response.data || response
  }

  /**
   * Update existing location
   * PUT /api/v1/locations/:id
   */
  const updateLocation = async (id: string, payload: UpdateLocationRequest): Promise<Location> => {
    const response = await $fetch<any>(`${getBaseUrl()}/locations/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: payload
    })

    return response.data || response
  }

  /**
   * Delete location
   * DELETE /api/v1/locations/:id
   */
  const deleteLocation = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/locations/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
  }

  return {
    getLocations,
    getAllLocations,
    getLocationById,
    createLocation,
    updateLocation,
    deleteLocation
  }
}
