// Department API Composable
import type {
  Department,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  ListDepartmentsResponse
} from '~/types/master'

export const useDepartmentApi = () => {
  const config = useRuntimeConfig()

  /**
   * Get base URL - with fallback to auditServiceBaseUrl if masterServiceBaseUrl not configured
   */
  const getBaseUrl = () => {
    return getAuditServiceBaseUrl()
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
    const url = new URL(`${getBaseUrl()}/departments`)

    if (params?.page) url.searchParams.set('page', String(params.page))
    if (params?.page_size) url.searchParams.set('page_size', String(params.page_size))
    if (params?.search) url.searchParams.set('search', params.search)

    const response = await $fetch<any>(url.toString(), {
      method: 'GET'
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
    const response = await $fetch<any>(`${getBaseUrl()}/departments`, {
      method: 'GET',
      params: { page: 1, page_size: 1000 }
    })

    const departments = response.data?.departments || response.departments || []
    return Array.isArray(departments) ? departments : []
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
