// Employee API Composable
import type {
  Employee,
  CreateEmployeeRequest,
  UpdateEmployeeRequest,
  ListEmployeesResponse
} from '~/types/master'

export const useEmployeeApi = () => {
  const config = useRuntimeConfig()

  // Base URL for master-service
  const baseUrl = config.public.masterServiceBaseUrl || 'http://localhost:8002/api/v1'

  /**
   * Get base URL - with fallback to auditServiceBaseUrl if masterServiceBaseUrl not configured
   */
  const getBaseUrl = () => {
    return getAuditServiceBaseUrl()
  }

  /**
   * Get all employees (with pagination)
   * GET /api/v1/employees?page=1&page_size=10&search=keyword
   */
  const getEmployees = async (params?: {
    page?: number
    page_size?: number
    search?: string
  }): Promise<ListEmployeesResponse> => {
    const url = new URL(`${getBaseUrl()}/employees`)

    if (params?.page) url.searchParams.set('page', String(params.page))
    if (params?.page_size) url.searchParams.set('page_size', String(params.page_size))
    if (params?.search) url.searchParams.set('search', params.search)

    const response = await $fetch<any>(url.toString(), {
      method: 'GET'
    })

    // Handle different response formats
    return {
      employees: response.data?.employees || response.employees || [],
      pagination: response.data?.pagination || response.pagination || {
        page: params?.page || 1,
        page_size: params?.page_size || 10,
        total: 0,
        total_pages: 0
      }
    }
  }

  /**
   * Get single employee by ID
   * GET /api/v1/employees/:id
   */
  const getEmployeeById = async (id: string): Promise<Employee> => {
    const response = await $fetch<any>(`${getBaseUrl()}/employees/${id}`, {
      method: 'GET'
    })

    return response.data || response
  }

  /**
   * Create new employee
   * POST /api/v1/employees
   */
  const createEmployee = async (payload: CreateEmployeeRequest): Promise<Employee> => {
    const response = await $fetch<any>(`${getBaseUrl()}/employees`, {
      method: 'POST',
      body: payload
    })

    return response.data || response
  }

  /**
   * Update existing employee
   * PUT /api/v1/employees/:id
   */
  const updateEmployee = async (id: string, payload: UpdateEmployeeRequest): Promise<Employee> => {
    const response = await $fetch<any>(`${getBaseUrl()}/employees/${id}`, {
      method: 'PUT',
      body: payload
    })

    return response.data || response
  }

  /**
   * Delete employee
   * DELETE /api/v1/employees/:id
   */
  const deleteEmployee = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/employees/${id}`, {
      method: 'DELETE'
    })
  }

  /**
   * Get all employees (no pagination - for dropdowns)
   * GET /api/v1/employees (fetch all)
   */
  const getAllEmployees = async (): Promise<Employee[]> => {
    const response = await $fetch<any>(`${getBaseUrl()}/employees`, {
      method: 'GET',
      params: { page: 1, page_size: 1000 }
    })

    const employees = response.data?.employees || response.employees || []
    return Array.isArray(employees) ? employees : []
  }

  return {
    getEmployees,
    getEmployeeById,
    createEmployee,
    updateEmployee,
    deleteEmployee,
    getAllEmployees
  }
}
