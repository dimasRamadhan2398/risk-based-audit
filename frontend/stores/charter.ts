import type { TableColumn } from '@nuxt/ui'
import { defineStore } from 'pinia'
import { computed, reactive, ref, watch } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import type { AuditCharter, CharterFormState } from '~/types/audit'
import { extractErrorMessage, getErrorStatus, getUserErrorMessage, parseBlobErrorBody } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'
import { useAuthStore } from '~/stores/auth'

export const ROLE_DEFAULT_FULL_NAMES: Record<string, string> = {
  admin: 'System Administrator',
  administrator: 'System Administrator',
  auditor: 'John Auditor',
  audit_manager: 'Audit Manager',
  chief_audit_executive: 'Chief Audit Executive',
  cae: 'Chief Audit Executive',
  department_head: 'Alice Department Head',
  auditee: 'Bob Auditee',
  viewer: 'Carol Viewer',
  audit_staff: 'Audit Staff',
}

export const getRoleFullName = (user?: any): string => {
  const fullName = user?.fullName || user?.full_name
  if (fullName && typeof fullName === 'string' && fullName.trim() !== '') {
    return fullName.trim()
  }

  const roles: string[] = []
  if (Array.isArray(user?.roles)) roles.push(...user.roles)
  if (user?.role) roles.push(user.role)
  if (user?.position) roles.push(user.position)

  for (const r of roles) {
    const key = String(r || '').toLowerCase().trim()
    if (ROLE_DEFAULT_FULL_NAMES[key]) {
      return ROLE_DEFAULT_FULL_NAMES[key]
    }
    const clean = key.replace(/[\s-]+/g, '_')
    if (ROLE_DEFAULT_FULL_NAMES[clean]) {
      return ROLE_DEFAULT_FULL_NAMES[clean]
    }
  }

  if (user?.username) {
    const un = String(user.username).toLowerCase().trim()
    if (ROLE_DEFAULT_FULL_NAMES[un]) {
      return ROLE_DEFAULT_FULL_NAMES[un]
    }
    return user.username
  }

  return 'System Administrator'
}

// GET /audit-charters pages its results: page_size defaults to 10 and is capped
// at 100, and the response is
// { data: { charters, pagination: { page, page_size, total_count, total_pages } } }.
// Without page/page_size only the 10 newest charters came back.
export const CHARTER_FETCH_PAGE_SIZE = 100
// Rows per page in the history table (paged client-side).
export const CHARTER_HISTORY_PAGE_SIZE = 10

/** total_pages from the list response; computed from total_count when it is missing. */
type CharterPaginationMeta = { total_pages?: unknown, total_count?: unknown, total?: unknown, page_size?: unknown }
export const extractCharterTotalPages = (response: unknown): number => {
  const res = response as { data?: { pagination?: CharterPaginationMeta }, pagination?: CharterPaginationMeta } | null | undefined
  const meta = res?.data?.pagination ?? res?.pagination
  const totalPages = Number(meta?.total_pages)
  if (Number.isFinite(totalPages) && totalPages > 0) return totalPages
  const totalCount = Number(meta?.total_count ?? meta?.total)
  const pageSize = Number(meta?.page_size) || CHARTER_FETCH_PAGE_SIZE
  if (Number.isFinite(totalCount) && totalCount > 0) return Math.ceil(totalCount / pageSize)
  return 1
}

export const useCharterStore = defineStore('charter', () => {
  // Modal State
  const showModal = ref(false)
  const errorMsg = ref('')
  const loading = ref(false)

  const isEditing = ref(false)
  const editingId = ref<string | null>(null)
  const toast = useToastNotification()
  const { t } = useI18n()

  const columns: (TableColumn<AuditCharter> & { class?: string })[] = [
    { accessorKey: 'version', header: 'Version', class: 'w-16 whitespace-nowrap text-center' },
    { accessorKey: 'title', header: 'Charter Name', class: 'w-48' },
    { accessorKey: 'date', header: 'Date', class: 'w-28 whitespace-nowrap' },
    { accessorKey: 'approvedBy', header: 'Approved By', class: 'w-36' },
    { accessorKey: 'uploadedBy', header: 'Uploaded By', class: 'w-36' },
    { accessorKey: 'actions', header: '', class: 'w-14 whitespace-nowrap text-center' },
  ]

  const authStore = useAuthStore()

  const currentUserFullName = computed(() => {
    let u: any = authStore.user
    if (!u) {
      try {
        const userCookie = useCookie<any>('auth-user')
        if (userCookie.value) {
          u = typeof userCookie.value === 'string' ? JSON.parse(userCookie.value) : userCookie.value
        }
      } catch {}
      if (!u && typeof window !== 'undefined') {
        try {
          const stored = localStorage.getItem('auth-user') || localStorage.getItem('user')
          if (stored) u = JSON.parse(stored)
        } catch {}
      }
    }
    return getRoleFullName(u)
  })

  // Form State
  const form = reactive<CharterFormState>({
    title: '',
    version: '', // Tidak perlu diisi user
    content: '',
    date: new Date().toISOString().split('T')[0] || '',
    uploadedBy: currentUserFullName.value,
    approvedBy: '',
    isActive: true,
    file: null,
    fileName: '',
  })

  // Keep form.uploadedBy in sync with Account Settings Full Name for current role
  watch(currentUserFullName, (newName, oldName) => {
    if (!isEditing.value && newName) {
      form.uploadedBy = newName
    }
    if (newName && oldName && charters.value.length > 0) {
      for (const c of charters.value) {
        if (!c.uploadedBy || c.uploadedBy === oldName || c.uploadedBy === 'Dimas (HIA)') {
          c.uploadedBy = newName
        }
      }
    }
  }, { immediate: true })



  /**
   * Mapper dari response backend ke format yang dipakai frontend.
   *
   * Backend audit-service kemungkinan menggunakan snake_case:
   * - filename
   * - is_active
   * - created_at
   *
   * Frontend menggunakan camelCase:
   * - fileName
   * - isActive
   * - uploadedBy
   * - approvedBy
   */
  const getCharterDownloadUrl = (id: string) => {
    const baseUrl = getAuditServiceBaseUrl()
    return `${baseUrl}/audit-charters/${id}/download`
  }

  const mapBackendToFrontend = (item: any): AuditCharter => {
    const dateValue =
      item.date ||
      item.created_at ||
      item.createdAt ||
      item.updated_at ||
      item.updatedAt ||
      new Date().toISOString()

    const fileSize =
      item.file_size || item.fileSize
        ? `${(Number(item.file_size || item.fileSize) / 1024 / 1024).toFixed(2)} MB`
        : item.fileSize || '-'

    const approvedBy =
      item.approved_by ||
      item.approvedBy ||
      (item.content && item.content !== '-' && !item.content.includes('establishes the authority')
        ? item.content
        : '-')

    const uploadedBy =
      (item.uploaded_by && item.uploaded_by !== 'Dimas (HIA)' && item.uploaded_by !== '-')
        ? item.uploaded_by
        : (item.uploadedBy && item.uploadedBy !== 'Dimas (HIA)' && item.uploadedBy !== '-')
          ? item.uploadedBy
          : currentUserFullName.value

    return {
      id: String(item.id),
      title: item.title || item.filename || item.fileName || '-',
      version: item.version || '-',
      content: item.content || '-',
      date: item.date || new Date(dateValue).toISOString().split('T')[0] || '',
      uploadedBy,
      approvedBy,
      isActive: item.is_active ?? item.isActive ?? false,
      fileName: item.filename || item.file_name || item.fileName || '-',
      fileSize,
      fileUrl: item.file_url || item.fileUrl || '#',
    }
  }

  /**
   * Helper untuk membaca bentuk response backend yang berbeda-beda.
   *
   * Kemungkinan response:
   * 1. { data: [...] }
   * 2. { data: { items: [...] } }
   * 3. { data: { data: [...] } }
   * 4. [...]
   */
  const extractItemsFromResponse = (response: any) => {
    if (Array.isArray(response?.data?.charters)) return response.data.charters
    if (Array.isArray(response?.data?.items)) return response.data.items
    if (Array.isArray(response?.data)) return response.data
    if (Array.isArray(response?.charters)) return response.charters
    if (Array.isArray(response)) return response
    return []
  }

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement

    // Ambil file dengan aman menggunakan optional chaining
    // target.files?[0] akan return 'File | undefined'
    const file = target.files?.[0]

    // GUARD CLAUSE (PENTING):
    // Jika file undefined, langsung berhenti.
    // Setelah baris ini, TypeScript tahu 'file' pasti bertipe 'File' (bukan undefined).
    if (!file) return

    // --- Mulai Validasi ---

    // Validasi Ukuran (Max 5MB)
    // Sekarang 'file.size' aman diakses karena file dijamin ada
    if (file.size > 5 * 1024 * 1024) {
      errorMsg.value = t('common.fileInput.tooLarge', { size: 5 })
      form.file = null

      // Reset input value agar user bisa pilih file ulang
      target.value = ''
      return
    }

    // Validasi Tipe
    const allowedTypes = [
      'application/pdf',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'application/msword',
    ]

    if (!allowedTypes.includes(file.type)) {
      errorMsg.value = t('common.fileInput.invalidPdfDocx')
      form.file = null
      target.value = ''
      return
    }

    // Jika lolos semua audit
    errorMsg.value = ''
    form.file = file
  }

  const charters = ref<AuditCharter[]>([])

  // --- LOGIC BARU: GETTER OTOMATISASI VERSI ---
  const nextVersion = computed(() => {
    // Skenario 1: Jika belum ada data sama sekali, mulai dari 1.0
    if (charters.value.length === 0) return '1.0'

    // Skenario 2: Ambil versi dari data paling atas (index 0 / terbaru)
    // Kita asumsikan data selalu di-unshift / terbaru di atas
    const latestVerStr = charters.value[0]?.version
    const latestVerNum = parseFloat(latestVerStr || '1.0')

    if (Number.isNaN(latestVerNum)) return '1.0'

    // Tambah 0.1 lalu jadikan string dengan 1 angka di belakang koma
    // Contoh: 1.1 + 0.1 = 1.2
    return (latestVerNum + 0.1).toFixed(1)
  })

  // Getters Lainnya
  const activeCharter = computed(() => charters.value.find(c => c.isActive))
  const historyCharters = computed(() => charters.value.filter(c => !c.isActive))

  // History table page. Kept here, not in the table, because the card swaps the
  // table for a skeleton while loading, which would reset it to page 1 on every
  // refetch (edit, delete, download).
  const historyPage = ref(1)
  const historyPageCount = computed(() =>
    Math.max(1, Math.ceil(historyCharters.value.length / CHARTER_HISTORY_PAGE_SIZE))
  )
  // A delete can remove the last row of the last page; move back to a page that exists.
  watch(historyPageCount, (pages) => {
    if (historyPage.value > pages) historyPage.value = pages
  })

  /**
   * GET /api/v1/audit-charters
   *
   * Dipanggil saat halaman Audit Charter dibuka,
   * dan setelah create/update/delete berhasil.
   *
   * Loads every charter: the first page gives total_pages, the rest are fetched
   * in parallel. The page needs the full list for the active card, the history
   * table and nextVersion.
   */
  const fetchCharters = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const url = `${getAuditServiceBaseUrl()}/audit-charters`
      // no-store: the backend sends this list with Cache-Control: public,
      // max-age=1800, so the refetch after a create/update/delete could
      // otherwise get the old list back from the browser cache
      const fetchPage = (page: number) => $fetch<unknown>(url, {
        method: 'GET',
        query: { page, page_size: CHARTER_FETCH_PAGE_SIZE },
        cache: 'no-store'
      })

      const first = await fetchPage(1)
      const totalPages = extractCharterTotalPages(first)
      const rest = totalPages > 1
        ? await Promise.all(Array.from({ length: totalPages - 1 }, (_, i) => fetchPage(i + 2)))
        : []

      // De-duplicate by id in case rows shift between page requests.
      const seen = new Set<string>()
      const items = [first, ...rest].flatMap(extractItemsFromResponse).filter((item: { id?: unknown } | null) => {
        if (item?.id === undefined || item?.id === null) return true
        const key = String(item.id)
        if (seen.has(key)) return false
        seen.add(key)
        return true
      })

      charters.value = items.map(mapBackendToFrontend)
    } catch (error: any) {
      console.error('Failed to fetch audit charters:', error)
      errorMsg.value = extractErrorMessage(error, t('auditCharter.toasts.charter.loadFailed'))
      charters.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * Helper to convert a File to base64 string.
   */
  const fileToBase64 = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => {
        // Remove the data URL prefix (e.g., "data:application/pdf;base64,")
        const base64 = (reader.result as string).split(',')[1] || ''
        resolve(base64)
      }
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
  }

  /**
   * POST /api/v1/audit-charters
   *
   * Ini masih CRUD JSON / metadata.
   * Belum upload file PDF asli.
   * Uploads file binary to Google Drive first, then saves charter metadata.
   */
  const itemFileSizeToBytes = (sizeStr: string | undefined): number => {
    if (!sizeStr || sizeStr === '-') return 0
    if (sizeStr.includes(' MB')) {
      return Math.round(parseFloat(sizeStr.replace(' MB', '')) * 1024 * 1024)
    }
    return parseFloat(sizeStr) || 0
  }

  const addCharter = async (form: CharterFormState) => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()
      const autoVersion = nextVersion.value

      const formData = new FormData()
      if (form.file) {
        formData.append('file', form.file)
      }
      formData.append('title', form.title)
      formData.append('version', autoVersion)
      formData.append('content', form.approvedBy || form.content || '')
      formData.append('approvedBy', form.approvedBy || '')
      formData.append('approved_by', form.approvedBy || '')
      const uploader = form.uploadedBy || currentUserFullName.value
      formData.append('uploadedBy', uploader)
      formData.append('uploaded_by', uploader)
      formData.append('isActive', String(form.isActive))
      formData.append('date', form.date)

      const authStore = useAuthStore()
      await $fetch(`${baseUrl}/audit-charters`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        },
        body: formData,
      })

      toast.showSuccess(t('auditCharter.toasts.charter.added'))
      // Newest first: the charter this one replaced is now at the top of page 1.
      historyPage.value = 1
      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to create audit charter:', error)
      const detail = extractErrorMessage(error, t('auditCharter.toasts.charter.addFailed'))
      errorMsg.value = detail
      toast.showError(t('auditCharter.toasts.charter.addFailed'), detail)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * PUT /api/v1/audit-charters/:id
   *
   * Updates metadata and optional new file binary.
   */
  const updateCharter = async (id: string, form: CharterFormState) => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()

      const formData = new FormData()
      if (form.file) {
        formData.append('file', form.file)
      }
      formData.append('title', form.title)
      formData.append('version', form.version || '')
      formData.append('content', form.approvedBy || form.content || '')
      formData.append('approvedBy', form.approvedBy || '')
      formData.append('approved_by', form.approvedBy || '')
      const uploader = form.uploadedBy || currentUserFullName.value
      formData.append('uploadedBy', uploader)
      formData.append('uploaded_by', uploader)
      formData.append('isActive', String(form.isActive))
      formData.append('date', form.date)

      const authStore = useAuthStore()
      await $fetch(`${baseUrl}/audit-charters/${id}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        },
        body: formData,
      })

      toast.showSuccess(t('auditCharter.toasts.charter.updated'))
      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to update audit charter:', error)
      const detail = extractErrorMessage(error, t('auditCharter.toasts.charter.updateFailed'))
      errorMsg.value = detail
      toast.showError(t('auditCharter.toasts.charter.updateFailed'), detail)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * DELETE /api/v1/audit-charters/:id
   */
  const deleteCharter = async (id: string) => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()

      const authStore = useAuthStore()
      await $fetch(`${baseUrl}/audit-charters/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${authStore.token}`
        }
      })

      toast.showSuccess(t('auditCharter.toasts.charter.deleted'))
      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to delete audit charter:', error)
      const detail = extractErrorMessage(error, t('auditCharter.toasts.charter.deleteFailed'))
      errorMsg.value = detail
      toast.showError(t('auditCharter.toasts.charter.deleteFailed'), detail)
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * GET /api/v1/audit-charters/:id/download
   *
   * Downloads the audit charter file by ID.
   */
  const downloadCharter = async (id: string, filename?: string, fileUrl?: string) => {
    loading.value = true
    errorMsg.value = ''

    try {
      if (fileUrl && (fileUrl.startsWith('http://') || fileUrl.startsWith('https://'))) {
        window.open(fileUrl, '_blank')
        return
      }

      // Fetched with $fetch so the auth token is sent; a bare window.open to
      // the API would be rejected with 401
      const baseUrl = getAuditServiceBaseUrl()
      const blob = await $fetch<Blob>(`${baseUrl}/audit-charters/${id}/download`, { responseType: 'blob' })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename || 'audit-charter.pdf'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
    } catch (error: any) {
      console.error('Failed to download audit charter:', error)
      // The blob request reads the error body as a Blob; parse it so the
      // status and backend message are visible. Not written to errorMsg: that
      // shows the "Failed to load Audit Charter" alert, which is wrong here.
      const parsedError = await parseBlobErrorBody(error)
      const notFound = getErrorStatus(parsedError) === 404
      const detail = notFound
        ? t('auditCharter.errors.fileMissing')
        : getUserErrorMessage(parsedError, t, { fallbackKey: 'auditCharter.errors.download' })
      toast.showError(t('auditCharter.errors.downloadTitle'), detail)
      // 404 can also mean the charter was deleted since the list loaded;
      // refetch so a stale row disappears
      if (notFound) await fetchCharters()
      throw error
    } finally {
      loading.value = false
    }
  }

  // Submit Handler
  const handleSubmit = async () => {
    if (!isEditing.value && !form.file) {
      errorMsg.value = t('auditCharter.form.fileRequired')
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        // MODE EDIT
        await updateCharter(editingId.value, { ...form })
        toast.showSuccess(t('auditCharter.toasts.charter.updated'))
      } else {
        // MODE ADD
        await addCharter({ ...form })
        toast.showSuccess(t('auditCharter.toasts.charter.uploaded'))
      }
      closeModal()
    } catch {
      // errorMsg sudah diisi di addCharter/updateCharter
    }
  }

  const closeModal = () => {
    showModal.value = false
    errorMsg.value = ''
    isEditing.value = false
    editingId.value = null

    // Reset Form
    form.title = ''
    form.version = ''
    form.content = ''
    form.date = new Date().toISOString().split('T')[0] || ''
    form.uploadedBy = currentUserFullName.value
    form.approvedBy = ''
    form.isActive = false
    form.file = null
    form.fileName = ''
  }

  const handleEdit = (charter: any) => {
    isEditing.value = true
    editingId.value = charter.id
    showModal.value = true

    // Isi form dengan data lama
    form.title = charter.title
    form.version = charter.version
    form.content = charter.content && charter.content !== '-' ? charter.content : ''
    form.date = charter.date
    form.uploadedBy = charter.uploadedBy && charter.uploadedBy !== '-' && charter.uploadedBy !== 'Dimas (HIA)' ? charter.uploadedBy : currentUserFullName.value
    form.approvedBy = charter.approvedBy && charter.approvedBy !== '-' ? charter.approvedBy : ''
    form.isActive = charter.isActive
    form.file = null // Reset file input karena file tidak wajib diisi saat edit
    form.fileName = charter.fileName || ''

    // Reset error
    errorMsg.value = ''
  }

  /**
   * Catatan:
   * definePageMeta sebaiknya tidak diletakkan di Pinia store.
   * Pindahkan ke file page yang menampilkan Audit Charter.
   *
   * Contoh:
   *
   * definePageMeta({
   *   layout: 'default',
   *   // middleware: 'auth'
   * })
   */

  return {
    charters,
    activeCharter,
    historyCharters,
    historyPage,
    historyPageCount,
    historyPageSize: CHARTER_HISTORY_PAGE_SIZE,
    nextVersion,
    showModal,
    loading,

    columns,
    form,
    isEditing,
    errorMsg,
    currentUserFullName,
    getRoleFullName,

    fetchCharters,
    addCharter,
    updateCharter,
    deleteCharter,
    downloadCharter,

    getCharterDownloadUrl,
    handleEdit,
    handleSubmit,
    closeModal,
    handleFileChange,
  }
})