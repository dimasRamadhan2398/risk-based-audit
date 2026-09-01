// Company Store - Pinia State Management
import type { TableColumn } from '@nuxt/ui'
import type {
  Company,
  CompanyFormState,
  CreateCompanyRequest,
  UpdateCompanyRequest,
  PaginationMeta,
  CompanyType
} from '~/types/master'

export const useCompanyStore = defineStore('company', () => {
  // ============= State =============
  const companies = ref<Company[]>([])
  const loading = ref(false)
  const errorMsg = ref('')

  // Modal State
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  // Pagination State
  const pagination = ref<PaginationMeta>({
    page: 1,
    page_size: 10,
    total: 0,
    total_pages: 0
  })
  const search = ref('')

  // Form State
  const form = reactive<CompanyFormState>({
    company_code: '',
    company_name: '',
    legal_name: '',
    tax_id: '',
    company_type: 'SUBSIDIARY' as CompanyType,
    parent_id: '',
    location_id: '',
    phone: '',
    email: '',
    website: '',
    is_active: true,
    established_at: ''
  })

  // Table Columns
  const columns: TableColumn<Company>[] = [
    { accessorKey: 'company_code', header: 'Code' },
    { accessorKey: 'company_name', header: 'Company Name' },
    { accessorKey: 'legal_name', header: 'Legal Name' },
    { accessorKey: 'company_type', header: 'Type' },
    { accessorKey: 'tax_id', header: 'Tax ID' },
    { accessorKey: 'is_active', header: 'Status' },
    { accessorKey: 'actions', header: '' }
  ]

  // ============= Getters =============
  const totalPages = computed(() => pagination.value.total_pages)
  const hasNextPage = computed(() => pagination.value.page < pagination.value.total_pages)
  const hasPrevPage = computed(() => pagination.value.page > 1)

  // ============= Actions =============

  /**
   * Fetch companies with pagination
   */
  const fetchCompanies = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useCompanyApi()
      const response = await api.getCompanies({
        page: pagination.value.page,
        page_size: pagination.value.page_size,
        search: search.value || undefined
      })

      companies.value = response.companies || []
      pagination.value = response.pagination || pagination.value
    } catch (error: any) {
      console.error('Failed to fetch companies:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil data company.'
      companies.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * Get single company by ID
   */
  const getCompanyById = async (id: string): Promise<Company | null> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useCompanyApi()
      return await api.getCompanyById(id)
    } catch (error: any) {
      console.error('Failed to fetch company:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil detail company.'
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * Create new company
   */
  const createCompany = async (): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useCompanyApi()
      const payload: CreateCompanyRequest = {
        company_code: form.company_code,
        company_name: form.company_name,
        legal_name: form.legal_name || undefined,
        tax_id: form.tax_id || undefined,
        company_type: form.company_type,
        parent_id: form.parent_id || undefined,
        location_id: form.location_id || undefined,
        phone: form.phone || undefined,
        email: form.email || undefined,
        website: form.website || undefined,
        is_active: form.is_active,
        established_at: form.established_at || undefined
      }

      await api.createCompany(payload)
      await fetchCompanies()
      return true
    } catch (error: any) {
      console.error('Failed to create company:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal membuat company baru.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Update existing company
   */
  const updateCompany = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useCompanyApi()
      const payload: UpdateCompanyRequest = {
        company_name: form.company_name || undefined,
        legal_name: form.legal_name || undefined,
        tax_id: form.tax_id || undefined,
        company_type: form.company_type || undefined,
        parent_id: form.parent_id || undefined,
        location_id: form.location_id || undefined,
        phone: form.phone || undefined,
        email: form.email || undefined,
        website: form.website || undefined,
        is_active: form.is_active,
        established_at: form.established_at || undefined
      }

      await api.updateCompany(id, payload)
      await fetchCompanies()
      return true
    } catch (error: any) {
      console.error('Failed to update company:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal memperbarui company.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete company
   */
  const deleteCompany = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useCompanyApi()
      await api.deleteCompany(id)
      await fetchCompanies()
      return true
    } catch (error: any) {
      console.error('Failed to delete company:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal menghapus company.'
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============

  /**
   * Open modal for creating new company
   */
  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    resetForm()
    showModal.value = true
  }

  /**
   * Open modal for editing existing company
   */
  const handleEdit = (company: Company) => {
    isEditing.value = true
    editingId.value = company.id

    // Fill form with existing data
    form.company_code = company.company_code
    form.company_name = company.company_name
    form.legal_name = company.legal_name || ''
    form.tax_id = company.tax_id || ''
    form.company_type = company.company_type
    form.parent_id = company.parent_id || ''
    form.location_id = company.location_id || ''
    form.phone = company.phone || ''
    form.email = company.email || ''
    form.website = company.website || ''
    form.is_active = company.is_active
    form.established_at = company.established_at || ''

    showModal.value = true
  }

  /**
   * Close modal and reset form
   */
  const closeModal = () => {
    showModal.value = false
    isEditing.value = false
    editingId.value = null
    errorMsg.value = ''
    resetForm()
  }

  /**
   * Reset form to default values
   */
  const resetForm = () => {
    form.company_code = ''
    form.company_name = ''
    form.legal_name = ''
    form.tax_id = ''
    form.company_type = 'SUBSIDIARY'
    form.parent_id = ''
    form.location_id = ''
    form.phone = ''
    form.email = ''
    form.website = ''
    form.is_active = true
    form.established_at = ''
  }

  /**
   * Handle form submit
   */
  const handleSubmit = async () => {
    // Validation
    if (!form.company_code.trim()) {
      errorMsg.value = 'Company code is required.'
      return
    }
    if (!form.company_name.trim()) {
      errorMsg.value = 'Company name is required.'
      return
    }
    if (!form.company_type) {
      errorMsg.value = 'Company type is required.'
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        const success = await updateCompany(editingId.value)
        if (success) {
          alert('Company berhasil diperbarui!')
          closeModal()
        }
      } else {
        const success = await createCompany()
        if (success) {
          alert('Company berhasil dibuat!')
          closeModal()
        }
      }
    } catch {
      // Error already handled in create/update
    }
  }

  /**
   * Handle delete with confirmation
   */
  const handleDelete = async (company: Company) => {
    if (!await useGlobalModalStore().confirmDelete({ description: `Are you sure you want to delete company "${company.company_name}"?` })) {
      return
    }

    const success = await deleteCompany(company.id)
    if (success) {
      alert('Company berhasil dihapus!')
    }
  }

  /**
   * Pagination controls
   */
  const setPage = (page: number) => {
    pagination.value.page = page
    fetchCompanies()
  }

  const setPageSize = (size: number) => {
    pagination.value.page_size = size
    pagination.value.page = 1
    fetchCompanies()
  }

  const setSearch = (value: string) => {
    search.value = value
    pagination.value.page = 1
    fetchCompanies()
  }

  const nextPage = () => {
    if (hasNextPage.value) {
      setPage(pagination.value.page + 1)
    }
  }

  const prevPage = () => {
    if (hasPrevPage.value) {
      setPage(pagination.value.page - 1)
    }
  }

  // ============= Return =============
  return {
    // State
    companies,
    loading,
    errorMsg,
    showModal,
    isEditing,
    editingId,
    pagination,
    search,
    form,
    columns,

    // Getters
    totalPages,
    hasNextPage,
    hasPrevPage,

    // Actions
    fetchCompanies,
    getCompanyById,
    createCompany,
    updateCompany,
    deleteCompany,
    openCreateModal,
    handleEdit,
    handleDelete,
    closeModal,
    handleSubmit,
    resetForm,
    setPage,
    setPageSize,
    setSearch,
    nextPage,
    prevPage
  }
})
