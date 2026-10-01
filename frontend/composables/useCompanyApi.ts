// Company API Composable
import type {
  Company,
  CreateCompanyRequest,
  UpdateCompanyRequest,
  ListCompaniesResponse
} from '~/types/master'
import { getMasterServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuthStore } from '~/stores/auth'

export const useCompanyApi = () => {
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

  const getBaseUrl = () => {
    return getMasterServiceBaseUrl()
  }

  const normalizeCompany = (c: any): Company => ({
    ...c,
    company_code: c.company_code || c.code || '',
    company_name: c.company_name || c.name || '',
    code: c.code || c.company_code || '',
    name: c.name || c.company_name || ''
  })

  /**
   * Get all companies (with pagination & search)
   * GET /api/v1/companies
   */
  const getCompanies = async (params?: {
    page?: number
    page_size?: number
    search?: string
  }): Promise<ListCompaniesResponse> => {
    // Base URLs are relative (/api/v1) in production, so build the query with
    // ofetch's `query` option instead of `new URL()`, which throws on relative URLs.
    // ofetch drops undefined values, so unset params are simply omitted.
    const response = await $fetch<any>(`${getBaseUrl()}/companies`, {
      method: 'GET',
      query: {
        page: params?.page || undefined,
        page_size: params?.page_size || undefined,
        search: params?.search || undefined
      }
    })

    const rawData = response.data
    let list: any[] = Array.isArray(rawData) ? rawData : (rawData?.companies || response.companies || [])

    // If API returned un-filtered array and search was requested, filter client-side as fallback
    if (params?.search && params.search.trim()) {
      const q = params.search.toLowerCase().trim()
      list = list.filter((c: any) =>
        (c.company_name || c.name || '').toLowerCase().includes(q) ||
        (c.company_code || c.code || '').toLowerCase().includes(q) ||
        (c.legal_name || '').toLowerCase().includes(q) ||
        (c.tax_id || '').toLowerCase().includes(q)
      )
    }

    const normalized = list.map(normalizeCompany)

    const page = params?.page || 1
    const pageSize = params?.page_size || 10
    const total = normalized.length

    return {
      companies: normalized,
      pagination: (!Array.isArray(rawData) && rawData?.pagination) || response.pagination || {
        page,
        page_size: pageSize,
        total,
        total_pages: Math.max(1, Math.ceil(total / pageSize))
      }
    }
  }

  /**
   * Get single company by ID
   * GET /api/v1/companies/:id
   */
  const getCompanyById = async (id: string): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies/${id}`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    const raw = response.data || response
    return normalizeCompany(raw)
  }

  /**
   * Create new company
   * POST /api/v1/companies
   */
  const createCompany = async (payload: CreateCompanyRequest): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: payload
    })

    const raw = response.data || response
    return normalizeCompany(raw)
  }

  /**
   * Update existing company
   * PUT /api/v1/companies/:id
   */
  const updateCompany = async (id: string, payload: UpdateCompanyRequest): Promise<Company> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: payload
    })

    const raw = response.data || response
    return normalizeCompany(raw)
  }

  /**
   * Delete company
   * DELETE /api/v1/companies/:id
   */
  const deleteCompany = async (id: string): Promise<void> => {
    await $fetch(`${getBaseUrl()}/companies/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
  }

  /**
   * Get all companies (no pagination - for dropdowns)
   * GET /api/v1/companies
   */
  const getAllCompanies = async (): Promise<Company[]> => {
    const response = await $fetch<any>(`${getBaseUrl()}/companies`, {
      method: 'GET',
      headers: getAuthHeaders(),
      params: { page: 1, page_size: 1000 }
    })

    const list = Array.isArray(response.data) ? response.data : (response.data?.companies || response.companies || [])
    return Array.isArray(list) ? list.map(normalizeCompany) : []
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
