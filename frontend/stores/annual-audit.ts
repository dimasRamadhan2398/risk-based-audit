// stores/annual-plan.ts
import type { TableColumn } from '@nuxt/ui'
import { defineStore } from 'pinia'
import { ref, reactive, computed, watch } from 'vue'
import { AnnualAuditPlanStatus, AuditDepartment, AuditCategory, type AnnualAuditPlan, type AnnualPlanForm } from '~/types/audit'
import { getAuditServiceBaseUrl, getRiskServiceBaseUrl } from '~/composables/useApiUrl'
import type { TablePagination } from '~/types/common'

export const useAnnualPlanStore = defineStore('annual-audit', () => {
  const showModal = ref(false)
  const errorMsg = ref('')
  const progressAudit = ref(50)
  const loading = ref(false)
  const validationErrors = reactive({
    activityDetail: '',
    timeline: '',
    auditor: ''
  })

  // Constants
  const monthsList = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const attachmentCategoryOptions = ['Plan', 'Evidence', 'Charter', 'Other']
  const paginations = ref<TablePagination>({
    pageSize: 10,
    currentIndex: 1,
    totalPages: 1
  })

  const approvalStepperItems = ref([
    {
      slot: 'draft' as const,
      label: 'Draft',
      description: 'Audit Staff',
      icon: 'i-lucide-check'
    }, {
      slot: 'review' as const,
      label: 'Under Review',
      description: 'Audit Manager',
      icon: 'i-lucide-clock'
    }, {
      slot: 'approval' as const,
      label: 'Pending Approval',
      description: 'Chief Audit Executive',
      icon: 'i-lucide-clock'
    }, {
      slot: 'approved' as const,
      label: 'Approved',
      description: 'System',
      icon: 'i-lucide-check-circle'
    }
  ]) as any

  const handleDownload = (plan: any) => {
    if (!plan) return

    const dataStr = JSON.stringify(plan, null, 2)
    const blob = new Blob([dataStr], { type: 'application/json' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')

    a.href = url
    a.download = `${plan.code}-annual-audit-plan.json`
    document.body.appendChild(a)
    a.click()

    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
  }

  const columns: TableColumn<AnnualAuditPlan>[] = [
    { accessorKey: 'activity', header: 'Activity' },
    { accessorKey: 'department', header: 'Department' },
    { accessorKey: 'riskName', header: 'Risk Name' },
    { accessorKey: 'riskLevel', header: 'Risk Level' },
    { accessorKey: 'timeline', header: 'Timeline' },
    { accessorKey: 'progress', header: 'Progress' },
    { accessorKey: 'status', header: 'Status' },
    { accessorKey: 'actions', header: 'Actions' }
  ]

  const isEditing = ref(false)
  const editingId = ref<string | null>(null)
  // --- STATE UNTUK MODAL VIEW ---
  const showViewModal = ref(false)
  const selectedPlan = ref<any>(null)

  // --- STATE UNTUK FILTER ---
  const searchCode = ref('')
  const selectedDepartment = ref<AuditDepartment | undefined>(undefined)
  const selectedStatus = ref<AnnualAuditPlanStatus | undefined>(undefined)

  // --- OPSI UNTUK DROPDOWN FILTER ---
  const yearOptions = ['2026', '2027', '2028', '2029', '2030']
  const departmentOptions = Object.values(AuditDepartment)
  const statusOptions = Object.values(AnnualAuditPlanStatus)

  const initialMockPlans: any[] = [
    {
      id: 'mock-1',
      code: 'PKAT-2026-ASR-001',
      version: 'v1.0',
      status: AnnualAuditPlanStatus.WORK_IN_PROGRESS,
      year: '2026',
      quarters: ['Q1', 'Q2'],
      selectedMonths: [0, 1, 3, 4],
      auditorCount: 3,
      daysPerAuditor: 10,
      totalMandays: 30,
      supervisorId: 'S01',
      supervisorName: 'Budi Santoso (Mgr)',
      notes: 'Audit operasional TI dan infrastruktur data center',
      isActive: true,
      activities: [
        {
          name: 'Audit Keamanan Siber & IT Infrastructure',
          category: AuditCategory.ASSURANCE,
          department: AuditDepartment.IT,
          riskName: 'Kelemahan Kontrol Akses & Server Downtime',
          riskLevel: 'High'
        }
      ]
    },
    {
      id: 'mock-2',
      code: 'PKAT-2026-ASR-002',
      version: 'v1.0',
      status: AnnualAuditPlanStatus.DONE,
      year: '2026',
      quarters: ['Q2', 'Q3'],
      selectedMonths: [4, 5, 7],
      auditorCount: 2,
      daysPerAuditor: 8,
      totalMandays: 16,
      supervisorId: 'S02',
      supervisorName: 'Siti Aminah (Snr)',
      notes: 'Audit kepatuhan regulasi OJK & laporan keuangan bulanan',
      isActive: true,
      activities: [
        {
          name: 'Audit Kepatuhan Keuangan & Pajak',
          category: AuditCategory.ASSURANCE,
          department: AuditDepartment.FINANCE,
          riskName: 'Penyimpangan Rekonsiliasi Bank',
          riskLevel: 'Medium'
        }
      ]
    }
  ]

  const normalizePlans = (rawPlans: any[]): AnnualAuditPlan[] => {
    if (!Array.isArray(rawPlans)) return []
    return rawPlans.map(p => ({
      ...p,
      code: p.code || p.plan_code || '',
      title: p.title || p.plan_title || '',
      activities: Array.isArray(p.activities)
        ? p.activities.map((a: any) => ({
            ...a,
            name: a.name || a.title || '',
            category: a.category || AuditCategory.ASSURANCE,
            department: a.department || (a.involved_departments && a.involved_departments[0]?.name) || AuditDepartment.IT,
            involvedDepartments: a.involvedDepartments || a.involved_departments || [],
            timelineText: a.timeline_text || a.timelineText || '',
            auditorCount: a.auditor_count || a.auditorCount || 1,
            totalMandays: a.total_mandays || a.totalMandays || 0,
            supervisorName: a.supervisor_name || a.supervisorName || ''
          }))
        : [],
      selectedMonths: Array.isArray(p.selectedMonths) ? p.selectedMonths : (Array.isArray(p.selected_months) ? p.selected_months : []),
      quarters: Array.isArray(p.quarters) ? p.quarters : []
    }))
  }

  const plans = ref<AnnualAuditPlan[]>(normalizePlans(initialMockPlans))

  // --- COMPUTED: FILTER DATA ---
  const filteredPlans = computed(() => {
    return (plans.value || []).filter((plan) => {
      const matchCode = !searchCode.value
        || (plan.code || '').toLowerCase().includes(searchCode.value.toLowerCase())
        || (plan.activities || []).some(act => (act.name || '').toLowerCase().includes(searchCode.value.toLowerCase()))

      const matchDept = !selectedDepartment.value
        || (plan.activities || []).some(act => act.department === selectedDepartment.value)

      const matchStatus = !selectedStatus.value || plan.status === selectedStatus.value

      return matchCode && matchDept && matchStatus
    })
  })

  // Fungsi Reset Filter
  const clearFilters = () => {
    searchCode.value = ''
    selectedDepartment.value = undefined
    selectedStatus.value = undefined
  }

  // --- ACTIONS UNTUK MODAL VIEW ---
  const openViewModal = (plan: any) => {
    selectedPlan.value = plan
    showViewModal.value = true
  }

  const closeViewModal = () => {
    showViewModal.value = false
    setTimeout(() => { selectedPlan.value = null }, 200)
  }

  watch(showViewModal, (isOpen) => {
    if (!isOpen) {
      setTimeout(() => { selectedPlan.value = null }, 200)
    }
  })

  const handleEditFromView = (plan: any) => {
    closeViewModal()
    handleEdit(plan)
  }

  const getSupervisorName = (id: string) => {
    if (!id) return '-'
    const supervisor = supervisors.value.find(s => s.id === id)
    return supervisor ? supervisor.name : id
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Done': return 'bg-green-500'
      case 'Work In Progress': return 'bg-amber-500'
      case 'Not Available': return 'bg-gray-400'
      default: return 'bg-gray-200'
    }
  }

  const getRiskLevelColor = (level?: string) => {
    if (!level) return 'neutral'
    const lvl = level.toLowerCase()
    if (lvl.includes('high')) return 'error'
    if (lvl.includes('mod') || lvl.includes('medium')) return 'warning'
    if (lvl.includes('low')) return 'success'
    return 'neutral'
  }

  // State untuk menyimpan ID baris yang sedang "Read More"
  const expandedNotesRows = ref(new Set<string>())
  const expandedDetailRows = ref(new Set<string>())

  const toggleNotesReadMore = (id: string) => {
    if (expandedNotesRows.value.has(id)) {
      expandedNotesRows.value.delete(id)
    } else {
      expandedNotesRows.value.add(id)
    }
  }

  const toggleDetailReadMore = (id: string) => {
    if (expandedDetailRows.value.has(id)) {
      expandedDetailRows.value.delete(id)
    } else {
      expandedDetailRows.value.add(id)
    }
  }

  // Helper untuk mengecek apakah teks lebih dari 100 karakter
  const isNotesLongText = (text: string) => text && text.length > 100
  const isDetailLongText = (text: string) => text && text.length > 100

  // Form State
  const form = reactive<AnnualPlanForm>({
    code: '',
    version: 'v1.0',
    revisionHistory: [],
    activities: [
      { name: '', category: AuditCategory.ASSURANCE, department: AuditDepartment.IT }
    ],
    status: AnnualAuditPlanStatus.NOT_AVAILABLE,
    selectedMonths: [],
    auditorCount: 2,
    daysPerAuditor: 5,
    supervisorId: '',
    attachmentCategory: '',
    attachmentUploadedBy: '',
    attachmentUploadDate: '',
    file: [],
    notes: '',
    staffApprovalNote: '',
    managerApprovalNote: '',
    chiefApprovalNote: '',
    isActive: true,
    year: ''
  })

  // --- COMPUTED LOGIC (Real-time Validation) ---
  const totalMandays = computed(() => form.auditorCount * form.daysPerAuditor)
  const selectedSupervisor = computed(() => supervisors.value.find(s => s.id === form.supervisorId))
  const utilizationData = computed(() => checkUtilization(totalMandays.value))
  const computedQuarters = computed(() => calculateQuarters(form.selectedMonths))
  const scheduleWarning = computed(() => checkScheduleGaps(form.selectedMonths))

  const quarterAlert = computed(() => {
    const q1Count = form.selectedMonths.filter(m => m <= 2).length
    const totalCount = form.selectedMonths.length
    if (totalCount > 0 && (q1Count / totalCount) > 0.4 && totalCount > 3) {
      return 'Beban kerja Triwulan I terlalu tinggi (>40%). Mohon ratakan jadwal.'
    }
    return null
  })

  const supervisorOptions = computed(() => {
    return supervisors.value.map(s => ({
      id: s.id,
      label: `${s.name} (Workload: ${s.workload})`
    }))
  })

  const toggleMonth = (idx: number) => {
    if (form.selectedMonths.includes(idx)) {
      form.selectedMonths = form.selectedMonths.filter(m => m !== idx)
    } else {
      form.selectedMonths.push(idx)
    }
  }

  const addActivity = () => {
    form.activities.push({
      name: '',
      category: AuditCategory.ASSURANCE,
      department: AuditDepartment.IT
    })
  }

  const removeActivity = (index: number) => {
    if (form.activities.length > 1) {
      form.activities.splice(index, 1)
    }
  }

  const getCategoryCode = (category?: string): string => {
    if (!category) return 'ASR'
    const cat = category.toLowerCase()
    if (cat.includes('assurance') && !cat.includes('quality')) return 'ASR'
    if (cat.includes('special')) return 'SPE'
    if (cat.includes('specific')) return 'SPR'
    if (cat.includes('consulting')) return 'CON'
    if (cat.includes('investigation')) return 'INV'
    if (cat.includes('quality') || cat.includes('qar')) return 'QAR'
    if (cat.includes('follow')) return 'FLW'
    const cleaned = category.replace(/[^a-zA-Z]/g, '')
    return cleaned.length >= 3 ? cleaned.slice(0, 3).toUpperCase() : 'ASR'
  }

  const generateAutoCode = () => {
    if (isEditing.value) return

    const yr = form.year || yearOptions[0] || new Date().getFullYear().toString()
    const catCode = getCategoryCode(form.activities[0]?.category)

    const seq = (plans.value.length + 1).toString().padStart(3, '0')
    form.code = `PKAT-${yr}-${catCode}-${seq}`
  }

  watch(
    [() => form.year, () => form.activities[0]?.category],
    () => {
      if (!isEditing.value && showModal.value) {
        generateAutoCode()
      }
    }
  )

  const openModal = () => {
    isEditing.value = false
    editingId.value = null

    clearValidationErrors()
    const defaultYear = yearOptions[0] || new Date().getFullYear().toString()

    Object.assign(form, {
      code: '',
      version: 'v1.0',
      revisionHistory: [],
      activities: [
        { name: '', category: AuditCategory.ASSURANCE, department: AuditDepartment.IT }
      ],
      status: AnnualAuditPlanStatus.NOT_AVAILABLE,
      selectedMonths: [],
      auditorCount: 2,
      daysPerAuditor: 5,
      supervisorId: '',
      notes: '',
      staffApprovalNote: '',
      managerApprovalNote: '',
      chiefApprovalNote: '',
      file: [],
      attachmentCategory: '',
      attachmentUploadedBy: '',
      attachmentUploadDate: '',
      year: defaultYear
    })

    generateAutoCode()
    showModal.value = true
  }

  const closeModal = () => showModal.value = false

  const yearlyUniverse = ref<any[]>([])

  const fetchYearlyUniverse = async (year: number) => {
    try {
      const baseUrl = getRiskServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-universe/year/${year}`)
      if (response && response.success) {
        yearlyUniverse.value = response.data
      }
    } catch (error) {
      console.error('Failed to fetch yearly universe for annual plan:', error)
    }
  }

  watch(
    () => filteredPlans.value.length,
    (total) => {
      const size = paginations.value.pageSize || 10
      const pages = Math.ceil(total / size) || 1
      paginations.value.totalPages = pages
      if (paginations.value.currentIndex > pages) {
        paginations.value.currentIndex = Math.max(1, pages)
      }
    },
    { immediate: true }
  )

  const fetchPlans = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response = await $fetch<Record<string, unknown>>(`${baseUrl}/annual-audit-plans`, {
        method: 'GET'
      })
      if (response && response.data && typeof response.data === 'object' && Array.isArray((response.data as Record<string, unknown>).items) && ((response.data as Record<string, unknown>).items as unknown[]).length > 0) {
        plans.value = normalizePlans((response.data as Record<string, unknown>).items as unknown[])
      } else if (response && response.data && Array.isArray(response.data) && response.data.length > 0) {
        plans.value = normalizePlans(response.data)
      } else if (response && Array.isArray(response.items) && response.items.length > 0) {
        plans.value = normalizePlans(response.items)
      } else if (Array.isArray(response) && response.length > 0) {
        plans.value = normalizePlans(response)
      }

      // Bind and update paginations data after fetching plans
      const resPagination = response?.pagination as Record<string, unknown> | undefined
      const totalItems = (resPagination?.total as number | undefined) ?? (response?.total as number | undefined) ?? plans.value.length
      const pageSize = (resPagination?.pageSize as number | undefined) ?? (resPagination?.page_size as number | undefined) ?? paginations.value.pageSize ?? 10
      const currentIndex = (resPagination?.currentIndex as number | undefined) ?? (resPagination?.page as number | undefined) ?? paginations.value.currentIndex ?? 1
      const totalPages = (resPagination?.totalPages as number | undefined) ?? (resPagination?.total_pages as number | undefined) ?? Math.ceil(totalItems / pageSize) ?? 1

      paginations.value = {
        pageSize,
        currentIndex,
        totalPages
      }
    } catch (error: unknown) {
      console.warn('Backend API unavailable, using initial mock annual plans:', error)
      const totalItems = plans.value.length
      const pageSize = paginations.value.pageSize || 10
      paginations.value = {
        pageSize,
        currentIndex: 1,
        totalPages: Math.ceil(totalItems / pageSize) || 1
      }
    } finally {
      loading.value = false
    }
  }

  const clearValidationErrors = () => {
    validationErrors.activityDetail = ''
    validationErrors.timeline = ''
    validationErrors.auditor = ''
  }

  const validateForm = () => {
    clearValidationErrors()

    const activityErrors: string[] = []
    const timelineErrors: string[] = []
    const auditorErrors: string[] = []

    // 1. Activity Detail
    if (!form.code.trim()) {
      activityErrors.push('Activity Code wajib diisi.')
    }

    if (!form.activities.length) {
      activityErrors.push('Minimal harus ada 1 aktivitas.')
    }

    form.activities.forEach((activity, index) => {
      if (!activity.name?.trim()) {
        activityErrors.push(
          `Activity Name pada Sub-Aktivitas ${index + 1} wajib diisi.`
        )
      }

      if (!activity.category) {
        activityErrors.push(
          `Category pada Sub-Aktivitas ${index + 1} wajib dipilih.`
        )
      }

      if (!activity.department) {
        activityErrors.push(
          `Department pada Sub-Aktivitas ${index + 1} wajib dipilih.`
        )
      }

      if (!activity.riskName?.trim()) {
        activityErrors.push(
          `Auditable Entity pada Sub-Aktivitas ${index + 1} wajib dipilih.`
        )
      }
    })

    // 2. Timeline
    if (!form.year) {
      timelineErrors.push('Year wajib dipilih.')
    }

    if (form.selectedMonths.length === 0) {
      timelineErrors.push('Minimal 1 bulan pelaksanaan wajib dipilih.')
    }

    // 3. Auditor
    if (
      !Number.isInteger(form.auditorCount)
      || form.auditorCount < 1
      || form.auditorCount > 10
    ) {
      auditorErrors.push('Number of Auditors harus antara 1 sampai 10.')
    }

    if (
      !Number.isFinite(form.daysPerAuditor)
      || form.daysPerAuditor < 1
    ) {
      auditorErrors.push('Duration harus minimal 1 hari.')
    }

    if (!form.supervisorId) {
      auditorErrors.push('Supervisor wajib dipilih.')
    }

    validationErrors.activityDetail = activityErrors.join(' ')
    validationErrors.timeline = timelineErrors.join(' ')
    validationErrors.auditor = auditorErrors.join(' ')

    return (
      activityErrors.length === 0
      && timelineErrors.length === 0
      && auditorErrors.length === 0
    )
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      alert('⚠️ Mohon lengkapi data pada Activity Detail, Timeline, dan Auditor.')
      return
    }

    try {
      if (isEditing.value && editingId.value) {
        await updatePlan(editingId.value, { ...form })
        alert('Data Rencana Audit Berhasil Diperbarui!')
      } else {
        await addPlan({ ...form })
        alert('Data Rencana Audit Berhasil Disimpan!')
      }

      closeModal()
    } catch (err: any) {
      alert('Gagal menyimpan data: ' + err.message)
    }
  }

  const handleEdit = (plan: any) => {
    isEditing.value = true
    editingId.value = plan.id

    clearValidationErrors()

    form.code = plan.code
    form.version = plan.version || 'v1.0'
    form.revisionHistory = plan.revisionHistory ? [...plan.revisionHistory] : []
    form.activities = plan.activities.map((act: any) => ({ ...act }))
    form.status = plan.status
    form.selectedMonths = [...plan.selectedMonths]
    form.auditorCount = plan.auditorCount
    form.daysPerAuditor = plan.daysPerAuditor
    form.supervisorId = plan.supervisorId
    form.notes = plan.notes || ''
    form.file = []
    form.attachmentCategory = plan.attachmentCategory || ''
    form.attachmentUploadedBy = plan.attachmentUploadedBy || ''
    form.attachmentUploadDate = plan.attachmentUploadDate || ''
    form.isActive = plan.isActive
    form.year = plan.year || ''

    showModal.value = true
  }

  const handleDelete = async (id: string | undefined) => {
    if (!id) return
    try {
      await deletePlan(id)
      closeViewModal()
    } catch (error) {
      alert('Gagal menghapus data: ' + error)
    }
  }

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    if (file.size > 5 * 1024 * 1024) {
      errorMsg.value = 'File terlalu besar! Maksimal 5MB.'
      form.file = null
      target.value = ''
      return
    }

    const allowedTypes = [
      'application/pdf',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'application/msword'
    ]

    if (!allowedTypes.includes(file.type)) {
      errorMsg.value = 'Format file tidak valid. Gunakan PDF atau DOCX.'
      form.file = null
      target.value = ''
      return
    }

    errorMsg.value = ''
    form.file = [file]
  }

  const supervisors = ref([
    { id: 'S01', name: 'Budi Santoso (Mgr)', workload: 5 },
    { id: 'S02', name: 'Siti Aminah (Snr)', workload: 8 },
    { id: 'S03', name: 'John Doe (Mgr)', workload: 2 }
  ])

  const TEAM_CAPACITY_PER_YEAR = 10 * 220 * 0.7

  const calculateQuarters = (months: number[]) => {
    const qSet = new Set<string>()
    months.forEach((m) => {
      if (m >= 0 && m <= 2) qSet.add('Q1')
      else if (m >= 3 && m <= 5) qSet.add('Q2')
      else if (m >= 6 && m <= 8) qSet.add('Q3')
      else qSet.add('Q4')
    })
    return Array.from(qSet).sort()
  }

  const checkUtilization = (additionalMandays: number) => {
    const currentUsed = plans.value.reduce((sum, p) => sum + (p.totalMandays || 0), 0)
    const totalAfterAdd = currentUsed + additionalMandays
    const utilization = (totalAfterAdd / TEAM_CAPACITY_PER_YEAR) * 100

    if (utilization > 95) return { color: 'red', percent: utilization, msg: '🔴 OVERLOAD (>95%)' }
    if (utilization > 80) return { color: 'yellow', percent: utilization, msg: '🟡 High Load (80-95%)' }
    return { color: 'green', percent: utilization, msg: '🟢 Optimal (60-80%)' }
  }

  const checkScheduleGaps = (months: number[]) => {
    if (months.length === 0) return 'Wajib pilih minimal 1 bulan.'
    const sorted = [...months].sort((a, b) => a - b)
    for (let i = 0; i < sorted.length - 1; i++) {
      if (sorted[i + 1]! - sorted[i]! > 1) {
        return '⚠️ Warning: Ada gap bulan kosong. Pastikan continuous coverage jika diperlukan.'
      }
    }
    return null
  }

  const addPlan = async (form: AnnualPlanForm) => {
    const quarters = calculateQuarters(form.selectedMonths)
    const totalMandays = form.auditorCount * form.daysPerAuditor
    const supervisor = supervisors.value.find(s => s.id === form.supervisorId)

    const payload = {
      code: form.code,
      version: 'v1.0',
      revisionHistory: [],
      activities: form.activities,
      status: form.status,
      selectedMonths: form.selectedMonths.sort((a, b) => a - b),
      quarters: quarters,
      auditorCount: form.auditorCount,
      daysPerAuditor: form.daysPerAuditor,
      totalMandays: totalMandays,
      supervisorId: form.supervisorId,
      supervisorName: supervisor?.name || 'Unknown',
      notes: form.notes,
      year: parseInt(form.year) || 2026,
      attachmentCategory: form.attachmentCategory,
      attachments: form.file && form.file.length > 0
        ? form.file.map((f: any) => ({
            name: f.name,
            size: Math.round(f.size / 1024) + ' KB',
            url: '#'
          }))
        : [],
      attachmentUploadedBy: form.attachmentUploadedBy,
      attachmentUploadDate: form.attachmentUploadDate,
      isActive: form.isActive
    }

    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/annual-audit-plans`, {
      method: 'POST',
      body: payload
    })
    await fetchPlans()
  }

  const updatePlan = async (id: string, updatedData: AnnualPlanForm) => {
    const baseUrl = getAuditServiceBaseUrl()
    const quarters = calculateQuarters(updatedData.selectedMonths)
    const totalMandays = updatedData.auditorCount * updatedData.daysPerAuditor
    const supervisor = supervisors.value.find(s => s.id === updatedData.supervisorId)

    const fileList = updatedData.file && updatedData.file.length > 0
      ? updatedData.file.map((f: any) => ({
          name: f.name,
          size: Math.round(f.size / 1024) + ' KB',
          url: '#'
        }))
      : []

    const payload = {
      ...updatedData,
      quarters: quarters,
      totalMandays: totalMandays,
      supervisorName: supervisor?.name || 'Unknown',
      year: parseInt(updatedData.year) || 2026,
      attachments: (updatedData.attachments || []).concat(fileList)
    }

    await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchPlans()
  }

  const deletePlan = async (id: string) => {
    const plan = plans.value.find(a => a.id === id)
    if (!plan) return

    const baseUrl = getAuditServiceBaseUrl()
    if (plan.isUsed) {
      const confirmed = confirm('Rencana audit ini sudah digunakan dan tidak dapat dihapus permanen. Apakah Anda ingin menonaktifkannya?')

      if (!confirmed) return

      const payload = {
        ...plan,
        isActive: false
      }

      await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
        method: 'PUT',
        body: payload
      })
    } else {
      const confirmed = confirm('Apakah Anda yakin ingin menghapus rencana ini secara permanen dari server?')

      if (!confirmed) return

      await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
        method: 'DELETE'
      })
    }
    await fetchPlans()
  }

  const createRevision = async (planId: string, changesNote: string, user: string) => {
    const plan = plans.value.find(p => p.id === planId)
    if (!plan) return

    const currentVersion = plan.version || 'v1.0'
    const parts = currentVersion.replace('v', '').split('.')
    const minor = parseInt(parts[1] || '0')
    const newVersion = `v${parts[0]}.${minor + 1}`

    const historyEntry = {
      date: new Date().toISOString().split('T')[0] || '',
      version: newVersion,
      changes: changesNote,
      user: user
    }

    const revisionHistory = plan.revisionHistory ? [...plan.revisionHistory] : []
    revisionHistory.unshift(historyEntry)

    const payload = {
      ...plan,
      version: newVersion,
      status: AnnualAuditPlanStatus.NOT_AVAILABLE,
      revisionHistory: revisionHistory
    }

    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/annual-audit-plans/${planId}`, {
      method: 'PUT',
      body: payload
    })
    await fetchPlans()
    alert(`Revised RKAT created (Version ${newVersion}).`)
  }

  const updatePlanStatus = async (id: string, status: AnnualAuditPlanStatus) => {
    const plan = plans.value.find(p => p.id === id)
    if (!plan) return

    const baseUrl = getAuditServiceBaseUrl()
    const payload = { ...plan, status: status }
    await $fetch(`${baseUrl}/annual-audit-plans/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchPlans()
  }

  const handleStaffApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.WORK_IN_PROGRESS)
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id)
    }
  }

  const handleStaffReject = () => {
    if (selectedPlan.value) {
      selectedPlan.value.staffApprovalNote = ''
    }
  }

  const handleManagerApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.PENDING_APPROVAL)
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id)
    }
  }

  const handleManagerReject = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.NOT_AVAILABLE)
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id)
    }
  }

  const handleChiefApprove = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.DONE)
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id)
    }
  }

  const handleChiefReject = async () => {
    if (selectedPlan.value) {
      await updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.WORK_IN_PROGRESS)
      selectedPlan.value = plans.value.find(p => p.id === selectedPlan.value.id)
    }
  }

  return {
    plans, supervisors, monthsList, yearOptions, supervisorOptions, attachmentCategoryOptions,
    showModal, isEditing, editingId, showViewModal, selectedPlan, progressAudit, approvalStepperItems,
    searchCode, selectedDepartment, selectedStatus, form, columns, errorMsg, loading, validationErrors,
    filteredPlans, totalMandays, selectedSupervisor, quarterAlert, scheduleWarning, utilizationData,
    computedQuarters, paginations, pagination: paginations,
    clearFilters, openViewModal, closeViewModal, departmentOptions, statusOptions,
    toggleMonth, addActivity, removeActivity, handleDownload,
    openModal, closeModal, handleSubmit, handleEdit, handleEditFromView, handleDelete, getSupervisorName,
    getStatusColor, handleFileChange, fetchPlans, generateAutoCode, getCategoryCode,
    handleStaffApprove, handleStaffReject, handleManagerApprove, handleManagerReject, handleChiefApprove, handleChiefReject,
    createRevision, getRiskLevelColor, yearlyUniverse, fetchYearlyUniverse
  }
})
