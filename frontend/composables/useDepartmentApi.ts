import type {
  Department,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  ListDepartmentsResponse
} from '~/types/master'
import { getMasterServiceBaseUrl } from '~/composables/useApiUrl'

export const useDepartmentApi = () => {
  /**
   * Get base URL from masterServiceBaseUrl
   */
  const getBaseUrl = () => {
    return getMasterServiceBaseUrl()
  }

  /**
   * Get all departments (with pagination)
   * GET /api/v1/departments?page=1&page_size=10&search=keyword
   */
  const getDepartments = async (params?: {
    page?: number
    page_size?: number
    search?: string
  }): Promise<ListDepartmentsResponse> => {
    // Base URLs are relative (/api/v1) in production, so build the query with
    // ofetch's `query` option instead of `new URL()`, which throws on relative URLs.
    // ofetch drops undefined values, so unset params are simply omitted.
    const response = await $fetch<any>(`${getBaseUrl()}/departments`, {
      method: 'GET',
      query: {
        page: params?.page || undefined,
        page_size: params?.page_size || undefined,
        search: params?.search || undefined
      }
    })

    // Handle different response formats
    return {
      departments: response.data?.departments || response.departments || [],
      pagination: response.data?.pagination || response.pagination || {
        page: params?.page || 1,
        page_size: params?.page_size || 10,
        total: 0,
        total_pages: 0
      }
    }
  }

  /**
   * Get single department by ID
   * GET /api/v1/departments/:id
   */
  const getDepartmentById = async (id: string): Promise<Department> => {
    const response = await $fetch<any>(`${getBaseUrl()}/departments/${id}`, {
      method: 'GET'
    })

    return response.data || response
  }

  /**
   * Create new department
   * POST /api/v1/departments
   */
  const createDepartment = async (payload: CreateDepartmentRequest): Promise<Department> => {
    const response = await $fetch<any>(`${getBaseUrl()}/departments`, {
      method: 'POST',
      body: payload
    })

    return response.data || response
  }

  /**
   * Update existing department
   * PUT /api/v1/departments/:id
   */
  const updateDepartment = async (id: string, payload: UpdateDepartmentRequest): Promise<Department> => {
    const response = await $fetch<any>(`${getBaseUrl()}/departments/${id}`, {
      method: 'PUT',
      body: payload
    })

    return response.data || response
  }

  /**
   * Delete department
   * DELETE /api/v1/departments/:id
   */
  const deleteDepartment = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/departments/${id}`, {
      method: 'DELETE'
    })
  }

  /**
   * Get all departments (no pagination - for dropdowns)
   * GET /api/v1/departments (fetch all)
   */
  const getAllDepartments = async (): Promise<Department[]> => {
    try {
      const response = await $fetch<any>(`${getBaseUrl()}/departments`, {
        method: 'GET',
        params: { page: 1, page_size: 1000 }
      })

      const departments = response.data?.departments || response.departments || (Array.isArray(response.data) ? response.data : [])
      return Array.isArray(departments) ? departments : []
    } catch (err) {
      console.error('Failed to fetch departments:', err)
      return []
    }
  }

  return {
    getDepartments,
    getDepartmentById,
    createDepartment,
    updateDepartment,
    deleteDepartment,
    getAllDepartments
  }
}
