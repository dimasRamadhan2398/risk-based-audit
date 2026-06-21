// stores/charter.ts
import type { TableColumn } from '@nuxt/ui'
import { defineStore } from 'pinia'
import type { AuditCharter, CharterFormState } from '~/types/audit'

export const useCharterStore = defineStore('charter', () => {
  // Modal State
  const showModal = ref(false)
  const errorMsg = ref('')
  const loading = ref(false)

  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const columns: TableColumn<AuditCharter>[] = [
    { accessorKey: 'version', header: 'Version' },
    { accessorKey: 'title', header: 'Charter Name' },
    { accessorKey: 'date', header: 'Date' },
    { accessorKey: 'approvedBy', header: 'Approved By' },
    { accessorKey: 'uploadedBy', header: 'Uploaded By' },
    { accessorKey: 'fileName', header: 'Actions' },
    { accessorKey: 'actions', header: '' },
  ]

  // Form State
  const form = reactive<CharterFormState>({
    title: '',
    version: '', // Tidak perlu diisi user
    date: new Date().toISOString().split('T')[0] || '',
    uploadedBy: 'Dimas (HIA)',
    approvedBy: '',
    isActive: true,
    file: null,
  })

  /**
   * Base URL audit-service.
   *
   * Pastikan di frontend/.env sudah ada:
   * NUXT_PUBLIC_AUDIT_SERVICE_BASE_URL=http://localhost:8002/api/v1
   */
  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

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

    return {
      id: String(item.id),
      title: item.title || item.filename || item.fileName || '-',
      version: item.version || '-',
      date: new Date(dateValue).toISOString().split('T')[0] || '',
      uploadedBy: item.uploaded_by || item.uploadedBy || 'Dimas (HIA)',
      approvedBy: item.approved_by || item.approvedBy || item.content || '-',
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
    if (Array.isArray(response?.data?.charters)) {
            return response.data.charters
        }
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
      errorMsg.value = 'File terlalu besar! Maksimal 5MB.'
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
      errorMsg.value = 'Format file tidak valid. Gunakan PDF atau DOCX.'
      form.file = null
      target.value = ''
      return
    }

    // Jika lolos semua audit
    errorMsg.value = ''
    form.file = file
  }

  /**
   * Data Audit Charter dari backend.
   *
   * Sebelumnya ini berisi mock data.
   * Sekarang dikosongkan, lalu diisi dari fetchCharters().
   */
  const mockCharters: AuditCharter[] = [
    {
      id: '2',
      title: 'Internal Audit Charter 2026',
      version: '1.1',
      date: '2026-01-15',
      uploadedBy: 'Dimas (HIA)',
      approvedBy: 'Audit Committee',
      isActive: true,
      fileName: 'audit-charter-v1.1_2026.pdf',
      fileSize: '3.1 MB',
      fileUrl: '#'
    },
    {
      id: '1',
      title: 'Internal Audit Charter 2025',
      version: '1.0',
      date: '2025-01-01',
      uploadedBy: 'Dimas (HIA)',
      approvedBy: 'Board of Directors',
      isActive: false,
      fileName: 'audit-charter-v1.0_2025.pdf',
      fileSize: '2.5 MB',
      fileUrl: '#'
    }
  ]

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

  /**
   * GET /api/v1/audit-charters
   *
   * Dipanggil saat halaman Audit Charter dibuka,
   * dan setelah create/update/delete berhasil.
   */
  const fetchCharters = async () => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()

      const response: any = await $fetch(`${baseUrl}/audit-charters`, {
        method: 'GET',
      })

      const items = extractItemsFromResponse(response)

      if (items.length > 0) {
        charters.value = items.map(mapBackendToFrontend)
      } else {
        charters.value = [...mockCharters]
      }
    } catch (error: any) {
      console.error('Failed to fetch audit charters, falling back to mock data:', error)
      errorMsg.value = 'Gagal mengambil data Audit Charter.'
      charters.value = [...mockCharters]
    } finally {
      loading.value = false
    }
  }

  /**
   * POST /api/v1/audit-charters
   *
   * Ini masih CRUD JSON / metadata.
   * Belum upload file PDF asli.
   *
   * Upload file PDF asli nanti perlu endpoint multipart/form-data
   * di backend, misalnya:
   * POST /api/v1/audit-charters/upload
   */
  const addCharter = async (form: CharterFormState) => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()

      const autoVersion = nextVersion.value
      const year = new Date(form.date).getFullYear()
      const fileExt = form.file?.name.split('.').pop() || 'pdf'

      // Gunakan autoVersion untuk penamaan file
      const generatedFileName = form.file?.name || `audit-charter-v${autoVersion}_${year}.${fileExt}`

      const payload = {
        filename: generatedFileName,
        version: autoVersion,
        title: form.title,
        content: form.approvedBy || '',
        is_active: form.isActive,
      }

      await $fetch(`${baseUrl}/audit-charters`, {
        method: 'POST',
        body: payload,
      })

      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to create audit charter:', error)
      errorMsg.value = 'Gagal menambahkan Audit Charter.'
      throw error
    } finally {
      loading.value = false
    }
  }

  /**
   * PUT /api/v1/audit-charters/:id
   *
   * Ini juga masih update JSON / metadata.
   * Jika form.file diisi, saat ini yang dikirim baru filename,
   * bukan file binary.
   */
  const updateCharter = async (id: string, form: CharterFormState) => {
    loading.value = true
    errorMsg.value = ''

    try {
      const baseUrl = getAuditServiceBaseUrl()

      const existingCharter = charters.value.find(c => c.id === id)

      const payload = {
        filename: form.file?.name || existingCharter?.fileName || undefined,
        version: form.version,
        title: form.title,
        content: form.approvedBy || '',
        is_active: form.isActive,
      }

      await $fetch(`${baseUrl}/audit-charters/${id}`, {
        method: 'PUT',
        body: payload,
      })

      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to update audit charter:', error)
      errorMsg.value = 'Gagal memperbarui Audit Charter.'
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

      await $fetch(`${baseUrl}/audit-charters/${id}`, {
        method: 'DELETE',
      })

      await fetchCharters()
    } catch (error: any) {
      console.error('Failed to delete audit charter:', error)
      errorMsg.value = 'Gagal menghapus Audit Charter.'
      throw error
    } finally {
      loading.value = false
    }
  }

  // Submit Handler
  const handleSubmit = async () => {
    /**
     * Catatan:
     * Saat ini endpoint backend yang dipakai masih POST JSON,
     * belum upload file multipart/form-data.
     *
     * Jadi validasi file wajib untuk mode ADD sementara dibuat warning saja,
     * agar CRUD JSON bisa dites dulu dari frontend.
     *
     * Nanti setelah endpoint upload PDF sudah dibuat,
     * validasi ini bisa dikembalikan menjadi wajib.
     */
    if (!isEditing.value && !form.file) {
      console.warn('File belum dikirim karena endpoint upload PDF belum dibuat.')
      // errorMsg.value = 'Mohon upload file charter.'
      // return
    }

    try {
      if (isEditing.value && editingId.value) {
        // MODE EDIT
        await updateCharter(editingId.value, { ...form })
        alert('Audit Charter berhasil diperbarui!')
      } else {
        // MODE ADD
        await addCharter({ ...form })
        alert('Audit Charter berhasil diupload!')
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
    form.date = new Date().toISOString().split('T')[0] || ''
    form.uploadedBy = 'Dimas (HIA)'
    form.approvedBy = ''
    form.isActive = false
    form.file = null
  }

  const handleEdit = (charter: any) => {
    isEditing.value = true
    editingId.value = charter.id
    showModal.value = true

    // Isi form dengan data lama
    form.title = charter.title
    form.version = charter.version
    form.date = charter.date
    form.uploadedBy = charter.uploadedBy
    form.approvedBy = charter.approvedBy
    form.isActive = charter.isActive
    form.file = null // Reset file input karena file tidak wajib diisi saat edit

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
    nextVersion,
    showModal,
    loading,

    columns,
    form,
    isEditing,
    errorMsg,

    fetchCharters,
    addCharter,
    updateCharter,
    deleteCharter,

    handleEdit,
    handleSubmit,
    closeModal,
    handleFileChange,
  }
})