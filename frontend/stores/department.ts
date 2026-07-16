// Department Store - Pinia State Management
import type { TableColumn } from '@nuxt/ui'
import type {
  Department,
  DepartmentFormState,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  PaginationMeta
} from '~/types/master'

export const useDepartmentStore = defineStore('department', () => {
  // ============= State =============
  const departments = ref<Department[]>([])
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
  const form = reactive<DepartmentFormState>({
    department_code: '',
    department_name: '',
    department_description: '',
    pic_id: '',
    level: 1,
    company_id: '',
    business_unit_id: '',
    is_active: true
  })

  // Table Columns
  const columns: TableColumn<Department>[] = [
    { accessorKey: 'department_code', header: 'Code' },
    { accessorKey: 'department_name', header: 'Department Name' },
    { accessorKey: 'department_description', header: 'Description' },
    { accessorKey: 'level', header: 'Level' },
    { accessorKey: 'is_active', header: 'Status' },
    { accessorKey: 'actions', header: '' }
  ]

  // ============= Getters =============
  const totalPages = computed(() => pagination.value.total_pages)
  const hasNextPage = computed(() => pagination.value.page < pagination.value.total_pages)
  const hasPrevPage = computed(() => pagination.value.page > 1)

  // ============= Actions =============

  /**
   * Fetch departments with pagination
   */
  const fetchDepartments = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useDepartmentApi()
      const response = await api.getDepartments({
        page: pagination.value.page,
        page_size: pagination.value.page_size,
        search: search.value || undefined
      })

      departments.value = response.departments || []
      pagination.value = response.pagination || pagination.value
    } catch (error: any) {
      console.error('Failed to fetch departments:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil data department.'
      departments.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * Get single department by ID
   */
  const getDepartmentById = async (id: string): Promise<Department | null> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useDepartmentApi()
      return await api.getDepartmentById(id)
    } catch (error: any) {
      console.error('Failed to fetch department:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil detail department.'
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * Create new department
   */
  const createDepartment = async (): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useDepartmentApi()
      const payload: CreateDepartmentRequest = {
        department_code: form.department_code,
        department_name: form.department_name,
        department_description: form.department_description || '',
        pic_id: form.pic_id,
        level: form.level,
        company_id: form.company_id,
        business_unit_id: form.business_unit_id || undefined,
        is_active: form.is_active
      }

      await api.createDepartment(payload)
      await fetchDepartments()
      return true
    } catch (error: any) {
      console.error('Failed to create department:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal membuat department baru.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Update existing department
   */
  const updateDepartment = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useDepartmentApi()
      const payload: UpdateDepartmentRequest = {
        department_name: form.department_name || undefined,
        department_description: form.department_description || undefined,
        pic_id: form.pic_id || undefined,
        level: form.level || undefined,
        company_id: form.company_id || undefined,
        business_unit_id: form.business_unit_id || undefined,
        is_active: form.is_active
      }

      await api.updateDepartment(id, payload)
      await fetchDepartments()
      return true
    } catch (error: any) {
      console.error('Failed to update department:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal memperbarui department.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete department
   */
  const deleteDepartment = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useDepartmentApi()
      await api.deleteDepartment(id)
      await fetchDepartments()
      return true
    } catch (error: any) {
      console.error('Failed to delete department:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal menghapus department.'
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============

  /**
   * Open modal for creating new department
   */
  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    resetForm()
    showModal.value = true
  }

  /**
   * Open modal for editing existing department
   */
  const handleEdit = (department: Department) => {
    isEditing.value = true
    editingId.value = department.id

    // Fill form with existing data
    form.department_code = department.department_code
    form.department_name = department.department_name
    form.department_description = department.department_description || ''
    form.pic_id = department.pic_id
    form.level = department.level
    form.company_id = department.company_id
    form.business_unit_id = department.business_unit_id || ''
    form.is_active = department.is_active

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
    form.department_code = ''
    form.department_name = ''
    form.department_description = ''
    form.pic_id = ''
    form.level = 1
    form.company_id = ''
    form.business_unit_id = ''
    form.is_active = true
  }

  /**
   * Handle form submit
   */
  const handleSubmit = async () => {
    // Validation
    if (!form.department_code.trim()) {
      errorMsg.value = 'Department code is required.'
      return
    }
    if (!form.department_name.trim()) {
      errorMsg.value = 'Department name is required.'
      return
    }
    if (!form.company_id.trim()) {
      errorMsg.value = 'Company is required.'
      return
    }
    if (!form.pic_id.trim()) {
      errorMsg.value = 'Person in Charge (PIC) is required.'
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        const success = await updateDepartment(editingId.value)
        if (success) {
          alert('Department berhasil diperbarui!')
          closeModal()
        }
      } else {
        const success = await createDepartment()
        if (success) {
          alert('Department berhasil dibuat!')
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
  const handleDelete = async (department: Department) => {
    if (!confirm(`Are you sure you want to delete department "${department.department_name}"?`)) {
      return
    }

    const success = await deleteDepartment(department.id)
    if (success) {
      alert('Department berhasil dihapus!')
    }
  }

  /**
   * Pagination controls
   */
  const setPage = (page: number) => {
    pagination.value.page = page
    fetchDepartments()
  }

  const setPageSize = (size: number) => {
    pagination.value.page_size = size
    pagination.value.page = 1
    fetchDepartments()
  }

  const setSearch = (value: string) => {
    search.value = value
    pagination.value.page = 1
    fetchDepartments()
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
    departments,
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
    fetchDepartments,
    getDepartmentById,
    createDepartment,
    updateDepartment,
    deleteDepartment,
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
