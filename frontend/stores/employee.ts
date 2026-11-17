// Employee Store - Pinia State Management
import type { TableColumn } from '@nuxt/ui'
import type {
  Employee,
  EmployeeFormState,
  CreateEmployeeRequest,
  EmployeeLoginAccount,
  UpdateEmployeeRequest,
  PaginationMeta,
  UserSeats
} from '~/types/master'
import { getUserErrorMessage } from '~/utils/error'
import { getEmployeeActionError, isSeatDowngraded, type EmployeeAction } from '~/utils/employeeSeats'
import { useI18n } from '~/composables/useI18n'
import { useAppToast } from '~/composables/useAppToast'

export const useEmployeeStore = defineStore('employee', () => {
  // ============= State =============
  const employees = ref<Employee[]>([])
  const loading = ref(false)
  const errorMsg = ref('')
  const { t } = useI18n()

  // Modal State
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  // Create-mode preview of the next employee code (display only, not reserved)
  const nextEmployeeCode = ref('')
  const nextEmployeeCodeLoading = ref(false)
  const nextEmployeeCodeFailed = ref(false)
  // Guards against a slow response landing after the modal was reopened or closed
  let nextEmployeeCodeRequestId = 0

  // Code returned by the last successful create, used in the success message
  const createdEmployeeCode = ref('')

  // Login account of the last created employee. Holds the one-time temporary
  // password only until the credentials dialog is dismissed; never persisted.
  const createdAccount = ref<EmployeeLoginAccount | null>(null)
  const createdAccountEmployeeName = ref('')
  // Login of the last created employee was set to VIEWER because the seats were full
  const createdSeatDowngraded = ref(false)

  // User seat usage (GET /users/seats, ADMIN only). Stays null until the page
  // loads it for an admin; after that every create/update/delete refreshes it.
  const seats = ref<UserSeats | null>(null)
  let seatsTracked = false

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
    work_location_id: '',
    residence_address: '',
    residence_city: '',
    residence_province: '',
    residence_postal_code: '',
    manager_id: '',
    is_active: true,
    join_date: new Date().toISOString().split('T')[0] || ''
  })

  // Table Columns (computed so headers follow the active locale)
  const columns = computed<TableColumn<Employee>[]>(() => [
    { accessorKey: 'employee_code', header: t('masterData.employee.columns.code') },
    { accessorKey: 'full_name', header: t('masterData.employee.columns.fullName') },
    { accessorKey: 'email', header: t('masterData.employee.columns.email') },
    { accessorKey: 'phone', header: t('masterData.employee.columns.phone') },
    { accessorKey: 'is_active', header: t('masterData.employee.columns.status') },
    { accessorKey: 'actions', header: t('common.columns.actions') }
  ])

  // ============= Getters =============
  const totalPages = computed(() => pagination.value.total_pages)
  const hasNextPage = computed(() => pagination.value.page < pagination.value.total_pages)
  const hasPrevPage = computed(() => pagination.value.page > 1)

  // ============= Helpers =============

  /**
   * Keep only digits. Phone is numeric-only (no "+", spaces or dashes).
   */
  const normalizePhone = (value: string | null | undefined): string => (value || '').replace(/\D/g, '')

  /**
   * Form value for join_date. The API returns RFC 3339 ("2021-03-15T00:00:00Z")
   * but the date picker works with YYYY-MM-DD. Go's zero time (0001-01-01, left
   * by an earlier update bug) is treated as unset.
   */
  const toJoinDateFormValue = (value: string | null | undefined): string => {
    const day = (value || '').slice(0, 10)
    return /^\d{4}-\d{2}-\d{2}$/.test(day) && day !== '0001-01-01' ? day : ''
  }

  /**
   * join_date as the backend parses it (RFC 3339; a bare YYYY-MM-DD is rejected),
   * or undefined when the form has no date.
   */
  const toJoinDatePayload = (value: string | null | undefined): string | undefined => {
    const day = toJoinDateFormValue(value)
    return day ? `${day}T00:00:00Z` : undefined
  }

  // ============= Actions =============

  /**
   * Load the seat usage. Admin only: the page calls this once for admins, and
   * later mutations refresh it through refreshSeats. A failure keeps the last
   * known value; seats are informational and never block a save.
   */
  const fetchSeats = async () => {
    seatsTracked = true
    try {
      seats.value = await useEmployeeApi().getUserSeats()
    } catch (error) {
      console.error('Failed to fetch user seats:', error)
    }
  }

  /** Stop tracking the seats (e.g. the signed-in user is not an admin) */
  const resetSeats = () => {
    seatsTracked = false
    seats.value = null
  }

  /** Refresh the seats only when they were loaded before (i.e. for admins) */
  const refreshSeats = async () => {
    if (seatsTracked) await fetchSeats()
  }

  /** Error toast for a failed row action, listing where the employee is used if known */
  const toastActionError = (error: unknown, action: EmployeeAction) => {
    const { message, usages } = getEmployeeActionError(error, t, action)
    const description = usages.length ? t('masterData.employee.errors.usedIn', { list: usages.join(', ') }) : undefined
    useAppToast().error(message, description)
  }

  /**
   * Fetch employees with pagination
   */
  const fetchEmployees = async (page?: number) => {
    if (page !== undefined) {
      pagination.value.page = page
    }
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
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.fetchEmployees' })
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
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.fetchEmployee' })
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
    createdEmployeeCode.value = ''
    createdAccount.value = null
    createdSeatDowngraded.value = false

    try {
      const api = useEmployeeApi()
      const payload: CreateEmployeeRequest = {
        full_name: form.full_name,
        email: form.email,
        phone: normalizePhone(form.phone) || '',
        company_id: form.company_id,
        department_id: form.department_id,
        job_role_id: form.job_role_id,
        work_location_id: (form.work_location_id && form.work_location_id !== '__none__') ? form.work_location_id : undefined,
        residence_address: form.residence_address || undefined,
        residence_city: form.residence_city || undefined,
        residence_province: form.residence_province || undefined,
        residence_postal_code: form.residence_postal_code || undefined,
        manager_id: (form.manager_id && form.manager_id !== '__none__') ? form.manager_id : undefined,
        is_active: form.is_active,
        // Backend parses join_date as RFC 3339; a bare YYYY-MM-DD is rejected
        join_date: `${form.join_date}T00:00:00Z`
      }

      const created = await api.createEmployee(payload)
      createdEmployeeCode.value = created?.employee_code || ''
      createdAccount.value = created?.account || null
      createdSeatDowngraded.value = isSeatDowngraded(created)
      createdAccountEmployeeName.value = form.full_name
      await Promise.all([fetchEmployees(), refreshSeats()])
      return true
    } catch (error: any) {
      console.error('Failed to create employee:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.createEmployee' })
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Drop the one-time temporary password (call when the credentials dialog closes)
   */
  const clearCreatedAccount = () => {
    createdAccount.value = null
    createdAccountEmployeeName.value = ''
  }

  /**
   * Report the login account result of a create as a toast; the temporary
   * password is never put in the toast (the credentials dialog shows it).
   */
  const notifyAccountResult = () => {
    const account = createdAccount.value
    const toast = useAppToast()
    const message = createdEmployeeCode.value
      ? t('masterData.employee.createdWithCode', { code: createdEmployeeCode.value })
      : t('masterData.employee.messages.created')
    if (account?.status === 'created') {
      toast.success(message, t('masterData.employee.account.createdToast', { username: account.username }))
    } else if (account?.status === 'linked') {
      toast.success(message, t('masterData.employee.account.linkedToast', { username: account.username }))
    } else {
      toast.success(message)
    }
    if (createdSeatDowngraded.value) {
      toast.warning(
        t('masterData.employee.seats.downgradedTitle'),
        account?.username
          ? t('masterData.employee.seats.downgradedToast', { username: account.username })
          : t('masterData.employee.seats.downgradedToastNoUsername')
      )
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
        // Backend still requires the existing code on update; it is read-only in the form
        employee_code: form.employee_code,
        // Backend replaces company_id on update; omitting it writes an empty company and fails the FK check
        company_id: form.company_id,
        full_name: form.full_name || undefined,
        email: form.email || undefined,
        phone: normalizePhone(form.phone) || undefined,
        department_id: form.department_id || undefined,
        job_role_id: form.job_role_id || undefined,
        work_location_id: (form.work_location_id && form.work_location_id !== '__none__') ? form.work_location_id : undefined,
        residence_address: form.residence_address || undefined,
        residence_city: form.residence_city || undefined,
        residence_province: form.residence_province || undefined,
        residence_postal_code: form.residence_postal_code || undefined,
        manager_id: (form.manager_id && form.manager_id !== '__none__') ? form.manager_id : undefined,
        is_active: form.is_active,
        // Omitted when the form has no date; the backend then keeps the stored join_date
        join_date: toJoinDatePayload(form.join_date)
      }

      await api.updateEmployee(id, payload)
      // Status and job role changes can take or free a seat
      await Promise.all([fetchEmployees(), refreshSeats()])
      return true
    } catch (error: any) {
      console.error('Failed to update employee:', error)
      errorMsg.value = getEmployeeActionError(error, t, 'update').message
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete employee. The default soft delete deactivates the login and can be
   * restored; hard: true removes the employee permanently. Failures are
   * reported as a toast so the list stays visible.
   */
  const deleteEmployee = async (id: string, options?: { hard?: boolean }): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      await api.deleteEmployee(id, options)
      await Promise.all([fetchEmployees(), refreshSeats()])
      return true
    } catch (error: any) {
      console.error('Failed to delete employee:', error)
      toastActionError(error, options?.hard ? 'hardDelete' : 'delete')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * Activate or deactivate an employee; the backend does the same to the
   * login. Sends only the status plus the two fields the update endpoint
   * requires (employee_code, company_id), so nothing else is overwritten.
   */
  const setEmployeeActive = async (employee: Employee, active: boolean): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''

    try {
      const api = useEmployeeApi()
      await api.updateEmployee(employee.id, {
        employee_code: employee.employee_code,
        company_id: employee.company_id,
        is_active: active
      })
      await Promise.all([fetchEmployees(), refreshSeats()])
      return true
    } catch (error: any) {
      console.error('Failed to change employee status:', error)
      toastActionError(error, active ? 'reactivate' : 'deactivate')
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============

  /**
   * Fetch the next employee code for display. Failure is not fatal:
   * the backend generates the real code on create.
   */
  const fetchNextEmployeeCode = async () => {
    const requestId = ++nextEmployeeCodeRequestId
    nextEmployeeCode.value = ''
    nextEmployeeCodeFailed.value = false
    nextEmployeeCodeLoading.value = true

    try {
      const api = useEmployeeApi()
      const code = await api.getNextEmployeeCode()
      if (requestId !== nextEmployeeCodeRequestId) return
      if (!code) throw new Error('Next employee code response was empty')
      nextEmployeeCode.value = code
    } catch (error) {
      if (requestId !== nextEmployeeCodeRequestId) return
      console.error('Failed to fetch next employee code:', error)
      nextEmployeeCodeFailed.value = true
    } finally {
      if (requestId === nextEmployeeCodeRequestId) {
        nextEmployeeCodeLoading.value = false
      }
    }
  }

  /**
   * Open modal for creating new employee
   */
  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    resetForm()
    showModal.value = true
    fetchNextEmployeeCode()
    // Keep the seats-full alert current (no-op for non-admins)
    refreshSeats()
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
    form.phone = normalizePhone(employee.phone)
    form.company_id = employee.company_id
    form.department_id = employee.department_id
    form.job_role_id = employee.job_role_id
    form.work_location_id = employee.work_location_id || ''
    form.residence_address = employee.residence_address || ''
    form.residence_city = employee.residence_city || ''
    form.residence_province = employee.residence_province || ''
    form.residence_postal_code = employee.residence_postal_code || ''
    form.manager_id = employee.manager_id || ''
    form.is_active = employee.is_active
    form.join_date = toJoinDateFormValue(employee.join_date)

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
    // Invalidate any in-flight next-code request
    nextEmployeeCodeRequestId++
    nextEmployeeCodeLoading.value = false
    nextEmployeeCodeFailed.value = false
    nextEmployeeCode.value = ''
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
    // Validation (employee_code is auto-generated by the backend, not required here)
    if (!form.full_name.trim()) {
      errorMsg.value = t('masterData.validation.fullNameRequired')
      return
    }
    if (!form.email.trim()) {
      errorMsg.value = t('masterData.validation.emailRequired')
      return
    }
    if (!form.company_id) {
      errorMsg.value = t('masterData.validation.companyRequired')
      return
    }
    if (!form.department_id) {
      errorMsg.value = t('masterData.validation.departmentRequired')
      return
    }
    if (!form.job_role_id) {
      errorMsg.value = t('masterData.validation.jobRoleRequired')
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        const success = await updateEmployee(editingId.value)
        if (success) {
          useAppToast().success(t('masterData.employee.messages.updated'))
          closeModal()
        }
      } else {
        const success = await createEmployee()
        if (success) {
          notifyAccountResult()
          // Without a one-time password there is nothing left to show
          if (createdAccount.value?.status !== 'created' || !createdAccount.value.temporary_password) {
            clearCreatedAccount()
          }
          closeModal()
        }
      }
    } catch {
      // Error already handled in create/update
    }
  }

  /**
   * Soft delete with confirmation: deactivates the login, can be restored
   */
  const handleDelete = async (employee: Employee) => {
    const confirmed = await useGlobalModalStore().confirmDelete({
      description: t('masterData.employee.softDeleteConfirm', { name: employee.full_name }),
      body: [t('masterData.employee.softDeleteConfirmBody')],
      confirmLabel: t('masterData.employee.actions.softDelete')
    })
    if (!confirmed) return

    const success = await deleteEmployee(employee.id)
    if (success) {
      useAppToast().success(t('masterData.employee.messages.softDeleted'))
    }
  }

  /**
   * Permanent delete (ADMIN only in the UI; the backend enforces it too)
   */
  const handleHardDelete = async (employee: Employee) => {
    const confirmed = await useGlobalModalStore().confirmDelete({
      description: t('masterData.employee.hardDeleteConfirm', { name: employee.full_name }),
      body: [t('masterData.employee.hardDeleteConfirmBody'), t('masterData.employee.hardDeleteConfirmNote')],
      confirmLabel: t('masterData.employee.actions.hardDelete')
    })
    if (!confirmed) return

    const success = await deleteEmployee(employee.id, { hard: true })
    if (success) {
      useAppToast().success(t('masterData.employee.messages.hardDeleted'))
    }
  }

  /**
   * Deactivate / reactivate with confirmation; the login follows the status
   */
  const handleToggleActive = async (employee: Employee) => {
    const activate = !employee.is_active
    const modal = useGlobalModalStore()
    const confirmed = activate
      ? await modal.confirmSubmit({
        description: t('masterData.employee.reactivateConfirm', { name: employee.full_name }),
        body: [t('masterData.employee.reactivateConfirmBody')],
        confirmLabel: t('masterData.employee.actions.reactivate')
      })
      : await modal.confirmDelete({
        description: t('masterData.employee.deactivateConfirm', { name: employee.full_name }),
        body: [t('masterData.employee.deactivateConfirmBody')],
        confirmLabel: t('masterData.employee.actions.deactivate')
      })
    if (!confirmed) return

    const success = await setEmployeeActive(employee, activate)
    if (success) {
      useAppToast().success(t(activate ? 'masterData.employee.messages.reactivated' : 'masterData.employee.messages.deactivated'))
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
    nextEmployeeCode,
    nextEmployeeCodeLoading,
    nextEmployeeCodeFailed,
    createdAccount,
    createdAccountEmployeeName,
    createdSeatDowngraded,
    seats,
    pagination,
    search,
    form,
    columns,
    normalizePhone,

    // Getters
    totalPages,
    hasNextPage,
    hasPrevPage,

    // Actions
    fetchEmployees,
    getEmployeeById,
    fetchNextEmployeeCode,
    createEmployee,
    clearCreatedAccount,
    updateEmployee,
    deleteEmployee,
    setEmployeeActive,
    fetchSeats,
    refreshSeats,
    resetSeats,
    openCreateModal,
    handleEdit,
    handleDelete,
    handleHardDelete,
    handleToggleActive,
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
