// Employee Store - Pinia State Management
import type { TableColumn } from '@nuxt/ui'
import type {
  Employee,
  EmployeeFormState,
  CreateEmployeeRequest,
  UpdateEmployeeRequest,
  PaginationMeta
} from '~/types/master'

export const useEmployeeStore = defineStore('employee', () => {
  // ============= State =============
  const employees = ref<Employee[]>([])
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
  const form = reactive<EmployeeFormState>({
    employee_code: '',
    full_name: '',
    email: '',
    phone: '',
    company_id: '',
    department_id: '',
    job_role_id: '',
    level_grade: 1,
    work_location_id: '',
    residence_address: '',
    residence_city: '',
    residence_province: '',
    residence_postal_code: '',
    manager_id: '',
    is_active: true,
    join_date: new Date().toISOString().split('T')[0] || ''
  })

  // Table Columns
  const columns: TableColumn<Employee>[] = [
    { accessorKey: 'employee_code', header: 'Code' },
    { accessorKey: 'full_name', header: 'Full Name' },
    { accessorKey: 'email', header: 'Email' },
    { accessorKey: 'phone', header: 'Phone' },
    { accessorKey: 'level_grade', header: 'Level' },
    { accessorKey: 'is_active', header: 'Status' },
    { accessorKey: 'actions', header: '' }
  ]

  // ============= Getters =============
  const totalPages = computed(() => pagination.value.total_pages)
  const hasNextPage = computed(() => pagination.value.page < pagination.value.total_pages)
  const hasPrevPage = computed(() => pagination.value.page > 1)

  // ============= Actions =============

  /**
   * Fetch employees with pagination
   */
  const fetchEmployees = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      const response = await api.getEmployees({
        page: pagination.value.page,
        page_size: pagination.value.page_size,
        search: search.value || undefined
      })

      employees.value = response.employees || []
      pagination.value = response.pagination || pagination.value
    } catch (error: any) {
      console.error('Failed to fetch employees:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil data employee.'
      employees.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * Get single employee by ID
   */
  const getEmployeeById = async (id: string): Promise<Employee | null> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      return await api.getEmployeeById(id)
    } catch (error: any) {
      console.error('Failed to fetch employee:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal mengambil detail employee.'
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * Create new employee
   */
  const createEmployee = async (): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      const payload: CreateEmployeeRequest = {
        employee_code: form.employee_code,
        full_name: form.full_name,
        email: form.email,
        phone: form.phone || '',
        company_id: form.company_id,
        department_id: form.department_id,
        job_role_id: form.job_role_id,
        level_grade: form.level_grade,
        work_location_id: form.work_location_id || undefined,
        residence_address: form.residence_address || undefined,
        residence_city: form.residence_city || undefined,
        residence_province: form.residence_province || undefined,
        residence_postal_code: form.residence_postal_code || undefined,
        manager_id: form.manager_id || undefined,
        is_active: form.is_active,
        join_date: form.join_date
      }

      await api.createEmployee(payload)
      await fetchEmployees()
      return true
    } catch (error: any) {
      console.error('Failed to create employee:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal membuat employee baru.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Update existing employee
   */
  const updateEmployee = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      const payload: UpdateEmployeeRequest = {
        full_name: form.full_name || undefined,
        email: form.email || undefined,
        phone: form.phone || undefined,
        department_id: form.department_id || undefined,
        job_role_id: form.job_role_id || undefined,
        level_grade: form.level_grade || undefined,
        work_location_id: form.work_location_id || undefined,
        residence_address: form.residence_address || undefined,
        residence_city: form.residence_city || undefined,
        residence_province: form.residence_province || undefined,
        residence_postal_code: form.residence_postal_code || undefined,
        manager_id: form.manager_id || undefined,
        is_active: form.is_active
      }

      await api.updateEmployee(id, payload)
      await fetchEmployees()
      return true
    } catch (error: any) {
      console.error('Failed to update employee:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal memperbarui employee.'
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete employee
   */
  const deleteEmployee = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      await api.deleteEmployee(id)
      await fetchEmployees()
      return true
    } catch (error: any) {
      console.error('Failed to delete employee:', error)
      errorMsg.value = error?.data?.message || error?.message || 'Gagal menghapus employee.'
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============

  /**
   * Open modal for creating new employee
   */
  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    resetForm()
    showModal.value = true
  }

  /**
   * Open modal for editing existing employee
   */
  const handleEdit = (employee: Employee) => {
    isEditing.value = true
    editingId.value = employee.id

    // Fill form with existing data
    form.employee_code = employee.employee_code
    form.full_name = employee.full_name
    form.email = employee.email
    form.phone = employee.phone || ''
    form.company_id = employee.company_id
    form.department_id = employee.department_id
    form.job_role_id = employee.job_role_id
    form.level_grade = employee.level_grade
    form.work_location_id = employee.work_location_id || ''
    form.residence_address = employee.residence_address || ''
    form.residence_city = employee.residence_city || ''
    form.residence_province = employee.residence_province || ''
    form.residence_postal_code = employee.residence_postal_code || ''
    form.manager_id = employee.manager_id || ''
    form.is_active = employee.is_active
    form.join_date = employee.join_date

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
    form.employee_code = ''
    form.full_name = ''
    form.email = ''
    form.phone = ''
    form.company_id = ''
    form.department_id = ''
    form.job_role_id = ''
    form.level_grade = 1
    form.work_location_id = ''
    form.residence_address = ''
    form.residence_city = ''
    form.residence_province = ''
    form.residence_postal_code = ''
    form.manager_id = ''
    form.is_active = true
    form.join_date = new Date().toISOString().split('T')[0] || ''
  }

  /**
   * Handle form submit
   */
  const handleSubmit = async () => {
    // Validation
    if (!form.employee_code.trim()) {
      errorMsg.value = 'Employee code is required.'
      return
    }
    if (!form.full_name.trim()) {
      errorMsg.value = 'Full name is required.'
      return
    }
    if (!form.email.trim()) {
      errorMsg.value = 'Email is required.'
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        const success = await updateEmployee(editingId.value)
        if (success) {
          alert('Employee berhasil diperbarui!')
          closeModal()
        }
      } else {
        const success = await createEmployee()
        if (success) {
          alert('Employee berhasil dibuat!')
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
  const handleDelete = async (employee: Employee) => {
    if (!await useGlobalModalStore().confirmDelete({ description: `Are you sure you want to delete employee "${employee.full_name}"?` })) {
      return
    }

    const success = await deleteEmployee(employee.id)
    if (success) {
      alert('Employee berhasil dihapus!')
    }
  }

  /**
   * Pagination controls
   */
  const setPage = (page: number) => {
    pagination.value.page = page
    fetchEmployees()
  }

  const setPageSize = (size: number) => {
    pagination.value.page_size = size
    pagination.value.page = 1
    fetchEmployees()
  }

  const setSearch = (value: string) => {
    search.value = value
    pagination.value.page = 1
    fetchEmployees()
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
    employees,
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
    fetchEmployees,
    getEmployeeById,
    createEmployee,
    updateEmployee,
    deleteEmployee,
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
