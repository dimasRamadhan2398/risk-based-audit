// Employee API Composable
import type {
  Employee,
  CreateEmployeeRequest,
  UpdateEmployeeRequest,
  ListEmployeesResponse
} from '~/types/master'
import { getMasterServiceBaseUrl } from '~/composables/useApiUrl'

export const useEmployeeApi = () => {
  const getBaseUrl = () => getMasterServiceBaseUrl()

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

    // Handle different response formats: response.data as array, or response.data.employees, or response.employees
    let employeesList: Employee[] = Array.isArray(response.data)
      ? response.data
      : (response.data?.employees || response.employees || response.data?.data || [])

    // Support client-side search filtering if query provided
    if (params?.search && params.search.trim()) {
      const q = params.search.trim().toLowerCase()
      employeesList = employeesList.filter((emp: any) =>
        emp.full_name?.toLowerCase().includes(q) ||
        emp.employee_code?.toLowerCase().includes(q) ||
        emp.email?.toLowerCase().includes(q) ||
        emp.phone?.toLowerCase().includes(q)
      )
    }

    const paginationData = response.data?.pagination || response.pagination
    const total = paginationData?.total ?? employeesList.length
    const pageSize = paginationData?.page_size ?? (params?.page_size || 10)
    const page = paginationData?.page ?? (params?.page || 1)
    const totalPages = paginationData?.total_pages ?? (Math.ceil(total / pageSize) || 1)

    return {
      employees: employeesList,
      pagination: {
        page,
        page_size: pageSize,
        total,
        total_pages: totalPages
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

    const employees = Array.isArray(response.data)
      ? response.data
      : (response.data?.employees || response.employees || response.data?.data || [])
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
