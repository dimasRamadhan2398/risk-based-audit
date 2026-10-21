// Employee API Composable
import type {
  Employee,
  CreateEmployeeRequest,
  UpdateEmployeeRequest,
  ListEmployeesResponse,
  NextEmployeeCodeResponse
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
    // Base URLs are relative (/api/v1) in production, so build the query with
    // ofetch's `query` option instead of `new URL()`, which throws on relative URLs.
    // ofetch drops undefined values, so unset params are simply omitted.
    const response = await $fetch<any>(`${getBaseUrl()}/employees`, {
      method: 'GET',
      query: {
        page: params?.page || undefined,
        page_size: params?.page_size || undefined,
        search: params?.search || undefined
      }
    })

    // Handle different response formats: response.data as array, or response.data.employees, or response.employees
    let employeesList: Employee[] = Array.isArray(response.data)
      ? response.data
      : (response.data?.employees || response.employees || response.data?.data || [])

    // Support client-side search filtering if query provided
    if (params?.search && params.search.trim()) {
      const q = params.search.trim().toLowerCase()
      employeesList = employeesList.filter((emp: any) =>
        emp.full_name?.toLowerCase().includes(q)
        || emp.employee_code?.toLowerCase().includes(q)
        || emp.email?.toLowerCase().includes(q)
        || emp.phone?.toLowerCase().includes(q)
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
   * Preview the code the next created employee will receive (reserves nothing)
   * GET /api/v1/employees/next-code
   */
  const getNextEmployeeCode = async (): Promise<string> => {
    const response = await $fetch<{ data?: NextEmployeeCodeResponse }>(`${getBaseUrl()}/employees/next-code`, {
      method: 'GET'
    })

    return response.data?.employee_code || ''
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
   * GET /api/v1/employees?page=N&page_size=100, walking pages until total_pages.
   * The backend caps page_size at 100, so a single request cannot fetch everything.
   */
  const getAllEmployees = async (): Promise<Employee[]> => {
    const pageSize = 100
    const all: Employee[] = []
    let page = 1
    let totalPages = 1

    do {
      const response = await $fetch<any>(`${getBaseUrl()}/employees`, {
        method: 'GET',
        params: { page, page_size: pageSize }
      })

      const employees = Array.isArray(response.data)
        ? response.data
        : (response.data?.employees || response.employees || response.data?.data || [])
      if (!Array.isArray(employees) || employees.length === 0) break

      all.push(...employees)

      // Without pagination info we cannot know whether more pages exist, so stop after this one.
      totalPages = response.data?.pagination?.total_pages ?? response.pagination?.total_pages ?? page
      page++
    } while (page <= totalPages)

    return all
  }

  return {
    getEmployees,
    getEmployeeById,
    getNextEmployeeCode,
    createEmployee,
    updateEmployee,
    deleteEmployee,
    getAllEmployees
  }
}
