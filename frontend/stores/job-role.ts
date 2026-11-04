// Job Role Store - Pinia State Management
// GET /job-roles is not paginated by the backend, so search and paging run client-side.
import type { TableColumn } from '@nuxt/ui'
import type {
  JobRole,
  JobRoleFormState,
  CreateJobRoleRequest,
  UpdateJobRoleRequest
} from '~/types/master'
import { getUserErrorMessage } from '~/utils/error'
import { FALLBACK_ACCESS_ROLES, suggestAccessRole } from '~/utils/jobRole'
import { useI18n } from '~/composables/useI18n'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useMasterOptionsStore } from '~/stores/master-options'
import { useGlobalModalStore } from '~/stores/global-modal'

export const useJobRoleStore = defineStore('job-role', () => {
  // ============= State =============
  const jobRoles = ref<JobRole[]>([])
  const accessRoles = ref<string[]>([...FALLBACK_ACCESS_ROLES])
  const loading = ref(false)
  const errorMsg = ref('')
  const { t } = useI18n()
  const toast = useToastNotification()

  // Modal State
  const showModal = ref(false)
  const isEditing = ref(false)
  // Stored code of the row being edited; a row with a blank code must stay editable so it can be repaired.
  const originalCode = ref('')
  const editingId = ref<string | null>(null)
  // True once the user picked (or cleared) the access role by hand; stops auto-suggestion.
  const accessRoleTouched = ref(false)

  // Client-side search / paging
  const search = ref('')
  const page = ref(1)
  const pageSize = ref(10)

  const form = reactive<JobRoleFormState>({
    job_role_code: '',
    job_role_name: '',
    job_role_description: '',
    job_position_type: '',
    default_access_role: '',
    is_active: true
  })

  const columns = computed<TableColumn<JobRole>[]>(() => [
    { accessorKey: 'job_role_code', header: t('masterData.jobRole.columns.code') },
    { accessorKey: 'job_role_name', header: t('masterData.jobRole.columns.name') },
    { accessorKey: 'job_position_type', header: t('masterData.jobRole.columns.positionType') },
    { accessorKey: 'default_access_role', header: t('masterData.jobRole.columns.accessRole') },
    { accessorKey: 'is_active', header: t('masterData.jobRole.columns.status') },
    { accessorKey: 'actions', header: t('common.columns.actions') }
  ])

  // ============= Getters =============
  const filtered = computed(() => {
    const q = search.value.trim().toLowerCase()
    if (!q) return jobRoles.value
    return jobRoles.value.filter(r =>
      [r.job_role_code, r.job_role_name, r.job_role_description, r.job_position_type, r.default_access_role]
        .some(v => (v || '').toLowerCase().includes(q))
    )
  })

  const paged = computed(() => {
    const start = (page.value - 1) * pageSize.value
    return filtered.value.slice(start, start + pageSize.value)
  })

  // ============= Actions =============
  const refreshOptions = async () => {
    try {
      await useMasterOptionsStore().fetchAll(true)
    } catch {
      // The employee form refetches on open; a failed refresh must not fail the save.
    }
  }

  const fetchJobRoles = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const api = useJobRoleApi()
      jobRoles.value = await api.getAllJobRoles()
    } catch (error: any) {
      console.error('Failed to fetch job roles:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.fetchJobRoles' })
      jobRoles.value = []
    } finally {
      loading.value = false
    }
  }

  const fetchAccessRoles = async () => {
    accessRoles.value = await useJobRoleApi().getAccessRoles()
  }

  const createJobRole = async (): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''
    try {
      const payload: CreateJobRoleRequest = {
        job_role_code: form.job_role_code.trim(),
        job_role_name: form.job_role_name.trim(),
        job_role_description: form.job_role_description || '',
        job_position_type: form.job_position_type as CreateJobRoleRequest['job_position_type'],
        default_access_role: form.default_access_role || undefined,
        is_active: form.is_active
      }
      await useJobRoleApi().createJobRole(payload)
      await fetchJobRoles()
      await refreshOptions()
      return true
    } catch (error: any) {
      console.error('Failed to create job role:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.createJobRole' })
      return false
    } finally {
      loading.value = false
    }
  }

  const updateJobRole = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''
    try {
      // '' clears the access role. The code is sent unchanged so the backend never sees it blank.
      const payload: UpdateJobRoleRequest = {
        job_role_code: form.job_role_code.trim(),
        job_role_name: form.job_role_name.trim(),
        job_role_description: form.job_role_description,
        job_position_type: form.job_position_type as UpdateJobRoleRequest['job_position_type'],
        default_access_role: form.default_access_role || '',
        is_active: form.is_active
      }
      await useJobRoleApi().updateJobRole(id, payload)
      await fetchJobRoles()
      await refreshOptions()
      return true
    } catch (error: any) {
      console.error('Failed to update job role:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.updateJobRole' })
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteJobRole = async (id: string): Promise<boolean> => {
    loading.value = true
    errorMsg.value = ''
    try {
      await useJobRoleApi().deleteJobRole(id)
      await fetchJobRoles()
      await refreshOptions()
      return true
    } catch (error: any) {
      console.error('Failed to delete job role:', error)
      errorMsg.value = getUserErrorMessage(error, t, { fallbackKey: 'masterData.errors.deleteJobRole' })
      return false
    } finally {
      loading.value = false
    }
  }

  // ============= UI Actions =============
  const resetForm = () => {
    form.job_role_code = ''
    originalCode.value = ''
    form.job_role_name = ''
    form.job_role_description = ''
    form.job_position_type = ''
    form.default_access_role = ''
    form.is_active = true
    accessRoleTouched.value = false
  }

  const openCreateModal = () => {
    isEditing.value = false
    editingId.value = null
    errorMsg.value = ''
    resetForm()
    showModal.value = true
  }

  const handleEdit = (role: JobRole) => {
    isEditing.value = true
    editingId.value = role.id
    errorMsg.value = ''
    form.job_role_code = role.job_role_code
    originalCode.value = role.job_role_code || ''
    form.job_role_name = role.job_role_name
    form.job_role_description = role.job_role_description || ''
    form.job_position_type = role.job_position_type
    form.default_access_role = role.default_access_role || ''
    form.is_active = role.is_active
    // An existing role keeps its stored access role until the user changes it.
    accessRoleTouched.value = true
    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
    isEditing.value = false
    editingId.value = null
    errorMsg.value = ''
    resetForm()
  }

  /** Called when the position type changes: fill the access role unless the user already chose one. */
  const onPositionTypeChange = (value: string) => {
    form.job_position_type = value as JobRoleFormState['job_position_type']
    if (!accessRoleTouched.value) {
      form.default_access_role = suggestAccessRole(value)
    }
  }

  const onAccessRoleChange = (value: string | null | undefined) => {
    form.default_access_role = value || ''
    accessRoleTouched.value = true
  }

  const handleSubmit = async () => {
    if (!form.job_role_code.trim()) {
      errorMsg.value = t('masterData.validation.jobRoleCodeRequired')
      return
    }
    if (!form.job_role_name.trim()) {
      errorMsg.value = t('masterData.validation.jobRoleNameRequired')
      return
    }
    if (!form.job_position_type) {
      errorMsg.value = t('masterData.validation.jobPositionTypeRequired')
      return
    }

    const wasEditing = isEditing.value
    const success = wasEditing && editingId.value
      ? await updateJobRole(editingId.value)
      : await createJobRole()
    if (success) {
      toast.showSuccess(t(wasEditing ? 'masterData.jobRole.updated' : 'masterData.jobRole.created'))
      closeModal()
    }
  }

  const handleDelete = async (role: JobRole) => {
    if (!await useGlobalModalStore().confirmDelete({ description: t('masterData.jobRole.deleteConfirm', { name: role.job_role_name }) })) {
      return
    }
    if (await deleteJobRole(role.id)) {
      toast.showSuccess(t('masterData.jobRole.deleted'))
    }
  }

  const setPage = (p: number) => {
    page.value = p
  }

  const setPageSize = (size: number) => {
    pageSize.value = size
    page.value = 1
  }

  const setSearch = (value: string) => {
    search.value = value
    page.value = 1
  }

  return {
    jobRoles,
    accessRoles,
    loading,
    errorMsg,
    showModal,
    isEditing,
    originalCode,
    editingId,
    accessRoleTouched,
    search,
    page,
    pageSize,
    form,
    columns,
    filtered,
    paged,
    fetchJobRoles,
    fetchAccessRoles,
    createJobRole,
    updateJobRole,
    deleteJobRole,
    openCreateModal,
    handleEdit,
    handleDelete,
    closeModal,
    handleSubmit,
    resetForm,
    onPositionTypeChange,
    onAccessRoleChange,
    setPage,
    setPageSize,
    setSearch
  }
})
