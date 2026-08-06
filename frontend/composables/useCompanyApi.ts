// Company API Composable
import type {
  Company,
  CreateCompanyRequest,
  UpdateCompanyRequest,
  ListCompaniesResponse
} from '~/types/master'

export const useCompanyApi = () => {
  const config = useRuntimeConfig()

  /**
   * Get base URL - with fallback to auditServiceBaseUrl if masterServiceBaseUrl not configured
   */
  const getBaseUrl = () => {
    return getAuditServiceBaseUrl()
  }

  /**
   * Get all companies (with pagination)
   * GET /api/v1/companies?page=1&page_size=10&search=keyword
   */
  const getCompanies = async (params?: {
    page?: number
    page_size?: number
    search?: string
  }): Promise<ListCompaniesResponse> => {
    const url = new URL(`${getBaseUrl()}/companies`)

    if (params?.page) url.searchParams.set('page', String(params.page))
    if (params?.page_size) url.searchParams.set('page_size', String(params.page_size))
    if (params?.search) url.searchParams.set('search', params.search)

    const response = await $fetch<any>(url.toString(), {
      method: 'GET'
    })

    // Handle different response formats
    const rawData = response.data
    const companies = Array.isArray(rawData) ? rawData : (rawData?.companies || response.companies || [])
    return {
      companies,
      pagination: (!Array.isArray(rawData) && rawData?.pagination) || response.pagination || {
        page: params?.page || 1,
        page_size: params?.page_size || 10,
        total: companies.length,
        total_pages: 1
      }
    }
  }

  /**
   * Get single company by ID
   * GET /api/v1/companies/:id
   */
  const getCompanyById = async (id: string): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies/${id}`, {
      method: 'GET'
    })

    return response.data || response
  }

  /**
   * Create new company
   * POST /api/v1/companies
   */
  const createCompany = async (payload: CreateCompanyRequest): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies`, {
      method: 'POST',
      body: payload
    })

    return response.data || response
  }

  /**
   * Update existing company
   * PUT /api/v1/companies/:id
   */
  const updateCompany = async (id: string, payload: UpdateCompanyRequest): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies/${id}`, {
      method: 'PUT',
      body: payload
    })

    return response.data || response
  }

  /**
   * Delete company
   * DELETE /api/v1/companies/:id
   */
  const deleteCompany = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/companies/${id}`, {
      method: 'DELETE'
    })
  }

  /**
   * Get all companies (no pagination - for dropdowns)
   * GET /api/v1/companies (fetch all)
   */
  const getAllCompanies = async (): Promise<Company[]> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies`, {
      method: 'GET',
      params: { page: 1, page_size: 1000 }
    })

    const companiesList = Array.isArray(response.data) ? response.data : (response.data?.companies || response.companies || [])
    return Array.isArray(companiesList) ? companiesList : []
  }

  return {
    getCompanies,
    getCompanyById,
    createCompany,
    updateCompany,
    deleteCompany,
    getAllCompanies
  }
}
