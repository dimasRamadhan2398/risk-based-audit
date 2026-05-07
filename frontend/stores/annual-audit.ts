// stores/annual-plan.ts
import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import { ref, reactive, computed, watch } from 'vue'
import { AnnualAuditPlanStatus, AuditDepartment, AuditCategory, type AnnualAuditPlan, type AnnualPlanForm } from '~/types/audit'

export const useAnnualPlanStore = defineStore('annual-audit', () => {

  const showModal = ref(false)
  const errorMsg = ref("")
  const progressAudit = ref(50);

  // Constants
  const monthsList = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const attachmentCategoryOptions = ['Plan', 'Evidence', 'Charter', 'Other']
  const open = ref(false)

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
  ]) as any;

  const handleDownload = (plan: any) => {
    if (!plan) return;

    const dataStr = JSON.stringify(plan, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');

    a.href = url;
    a.download = `${plan.code}-annual-audit-plan.json`;
    document.body.appendChild(a);
    a.click();

    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  defineShortcuts({
    o: () => open.value = !open.value
  })

  const columns: TableColumn<AnnualAuditPlan>[] = [
    { accessorKey: 'activity', header: 'Activity' },
    { accessorKey: 'department', header: 'Department' },
    { accessorKey: 'timeline', header: 'Timeline' },
    { accessorKey: 'progress', header: 'Progress' },
    { accessorKey: 'status', header: 'Status' },
    { accessorKey: 'actions', header: 'Actions' },
  ]

  const isEditing = ref(false)
  const editingId = ref<string | null>(null)
  // --- STATE UNTUK MODAL VIEW ---
  const showViewModal = ref(false)
  const selectedPlan = ref<any>(null)

  // --- STATE UNTUK FILTER ---
  const searchCode = ref('')
  const selectedDepartment = ref<string | undefined>(undefined)
  const selectedStatus = ref<string | undefined>(undefined)

  // --- OPSI UNTUK DROPDOWN FILTER ---
  const yearOptions = ['2026', '2027', '2028', '2029', '2030']

  // --- COMPUTED: FILTER DATA ---
  // --- 3. UPDATE FILTERED PLANS ---
  const filteredPlans = computed(() => {
    // 1. TAMBAHKAN .value DI SINI
    return plans.value.filter(plan => {

      // 2. Opsional tapi disarankan: gunakan optional chaining (?.) untuk mencegah error jika activities kosong/undefined
      const matchCode = !searchCode.value ||
        plan.code.toLowerCase().includes(searchCode.value.toLowerCase()) ||
        plan.activities?.some(act => act.name.toLowerCase().includes(searchCode.value.toLowerCase()))

      const matchDept = !selectedDepartment.value ||
        plan.activities?.some(act => act.department === selectedDepartment.value)

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
    setTimeout(() => { selectedPlan.value = null }, 200) // Delay agar transisi tidak flicker
  }

  watch(showViewModal, (isOpen) => {
    if (!isOpen) {
      setTimeout(() => { selectedPlan.value = null }, 200)
    }
  })

  const handleEditFromView = (plan: any) => {
    closeViewModal() // Tutup modal detail
    handleEdit(plan) // Buka modal edit bawaan form
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

  // F-03: Auto Calculate Mandays
  const totalMandays = computed(() => form.auditorCount * form.daysPerAuditor)

  // F-03: Supervisor Check
  const selectedSupervisor = computed(() => supervisors.value.find(s => s.id === form.supervisorId))

  // F-03: Utilization Check
  const utilizationData = computed(() => checkUtilization(totalMandays.value))

  // F-02: Schedule Logic
  const computedQuarters = computed(() => calculateQuarters(form.selectedMonths))

  const scheduleWarning = computed(() => checkScheduleGaps(form.selectedMonths))

  const quarterAlert = computed(() => {
    // F-02: Alert jika Q1 > 40% (Simplifikasi: jika user pilih semua bulan Q1 dan hanya sedikit bulan lain)
    const q1Count = form.selectedMonths.filter(m => m <= 2).length
    const totalCount = form.selectedMonths.length
    if (totalCount > 0 && (q1Count / totalCount) > 0.4 && totalCount > 3) {
      return "Beban kerja Triwulan I terlalu tinggi (>40%). Mohon ratakan jadwal."
    }
    return null
  })

  const supervisorOptions = computed(() => {
    return supervisors.value.map(s => ({
      id: s.id,
      label: `${s.name} (Workload: ${s.workload})`
    }))
  })

  // --- ACTIONS ---

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
      department: AuditDepartment.IT,
    })
  }

  const removeActivity = (index: number) => {
    // Cegah penghapusan jika hanya tersisa 1 aktivitas
    if (form.activities.length > 1) {
      form.activities.splice(index, 1)
    }
  }

  const openModal = () => {
    isEditing.value = false
    editingId.value = null

    // Reset Form
    Object.assign(form, {
      code: '',
      version: 'v1.0',
      revisionHistory: [],
      activities: [ // Reset array activities kembali ke 1 baris kosong
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
      year: ''
    })
    showModal.value = true
  }

  const closeModal = () => showModal.value = false

  const handleSubmit = () => {
    // if (!isEditing.value && (!form.file || form.file.length === 0)) {
    //   errorMsg.value = "Mohon upload file charter.";
    //   return;
    // }

    // F-04: Final Validation
    if (form.selectedMonths.length === 0) {
      alert("⚠️ Wajib memilih minimal 1 bulan pelaksanaan.")
      return
    }
    if (!form.supervisorId) {
      alert("⚠️ Wajib memilih Supervisor.")
      return
    }

    if (isEditing.value && editingId.value) {
      // Mode EDIT
      updatePlan(editingId.value, { ...form })
      alert("Data Rencana Audit Berhasil Diperbarui!")
    } else {
      // Mode ADD
      addPlan({ ...form })
      alert("Data Rencana Audit Berhasil Disimpan!")
    }

    closeModal()
  }

  const handleEdit = (plan: any) => {
    isEditing.value = true
    editingId.value = plan.id

    // Isi form dengan data yang dipilih

    form.code = plan.code,
      form.version = plan.version || 'v1.0',
      form.revisionHistory = plan.revisionHistory ? [...plan.revisionHistory] : [],
      form.activities = plan.activities.map((act: any) => ({ ...act })),
      form.status = plan.status,
      form.selectedMonths = [...plan.selectedMonths], // Gunakan spread agar tidak reaktif terhubung langsung
      form.auditorCount = plan.auditorCount,
      form.daysPerAuditor = plan.daysPerAuditor,
      form.supervisorId = plan.supervisorId,
      form.notes = plan.notes || '',
      form.file = [],
      form.attachmentCategory = '',
      form.attachmentUploadedBy = '',
      form.attachmentUploadDate = '',
      form.isActive = plan.isActive,
      form.year = plan.year


    showModal.value = true
  }

  const handleDelete = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deletePlan(id)

      // 3. Tutup modal detail setelah berhasil menghapus
      closeViewModal()

    } catch (error) {
      alert('Gagal menghapus data: ' + error)
    }
  }

  const handleFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement;

    // Ambil file dengan aman menggunakan optional chaining
    // target.files?[0] akan return 'File | undefined'
    const file = target.files?.[0];

    // GUARD CLAUSE (PENTING):
    // Jika file undefined, langsung berhenti.
    // Setelah baris ini, TypeScript tahu 'file' pasti bertipe 'File' (bukan undefined).
    if (!file) return;

    // --- Mulai Validasi ---

    // Validasi Ukuran (Max 5MB)
    // Sekarang 'file.size' aman diakses karena file dijamin ada
    if (file.size > 5 * 1024 * 1024) {
      errorMsg.value = "File terlalu besar! Maksimal 5MB.";
      form.file = null;
      // Reset input value agar user bisa pilih file ulang
      target.value = "";
      return;
    }

    // Validasi Tipe
    const allowedTypes = [
      "application/pdf",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      "application/msword",
    ];

    if (!allowedTypes.includes(file.type)) {
      errorMsg.value = "Format file tidak valid. Gunakan PDF atau DOCX.";
      form.file = null;
      target.value = "";
      return;
    }

    // Jika lolos semua audit
    errorMsg.value = "";
    form.file = [file];
  };

  const supervisors = ref([
    { id: 'S01', name: 'Budi Santoso (Mgr)', workload: 5 },
    { id: 'S02', name: 'Siti Aminah (Snr)', workload: 8 }, // Overloaded
    { id: 'S03', name: 'John Doe (Mgr)', workload: 2 },
  ])

  // Mock Team Capacity (F-03 Logic)
  // Available = Total Auditor (misal 10) * 220 hari * 70%
  const TEAM_CAPACITY_PER_YEAR = 10 * 220 * 0.7 // 1540 Mandays
  // F-02: Auto-grouping Quarters
  const calculateQuarters = (months: number[]) => {
    const qSet = new Set<string>()
    months.forEach(m => {
      if (m >= 0 && m <= 2) qSet.add('Q1')
      else if (m >= 3 && m <= 5) qSet.add('Q2')
      else if (m >= 6 && m <= 8) qSet.add('Q3')
      else qSet.add('Q4')
    })
    return Array.from(qSet).sort()
  }

  // F-03: Utilization Check
  // Returns: { status: 'Green'|'Yellow'|'Red', percent: number, message: string }
  const checkUtilization = (additionalMandays: number) => {
    const currentUsed = plans.value.reduce((sum, p) => sum + p.totalMandays!, 0)
    const totalAfterAdd = currentUsed + additionalMandays
    const utilization = (totalAfterAdd / TEAM_CAPACITY_PER_YEAR) * 100

    if (utilization > 95) return { color: 'red', percent: utilization, msg: '🔴 OVERLOAD (>95%)' }
    if (utilization > 80) return { color: 'yellow', percent: utilization, msg: '🟡 High Load (80-95%)' }
    return { color: 'green', percent: utilization, msg: '🟢 Optimal (60-80%)' }
  }

  // F-04: Validate Schedule Gaps
  const checkScheduleGaps = (months: number[]) => {
    if (months.length === 0) return "Wajib pilih minimal 1 bulan."
    // Logic simple: Cek apakah bulan loncat (misal Jan & Mar dipilih, Feb kosong)
    const sorted = [...months].sort((a, b) => a - b)
    for (let i = 0; i < sorted.length - 1; i++) {
      if (sorted[i + 1]! - sorted[i]! > 1) {
        return "⚠️ Warning: Ada gap bulan kosong. Pastikan continuous coverage jika diperlukan."
      }
    }
    return null
  }

  const plans = ref<AnnualAuditPlan[]>([])

  // Actions
  const addPlan = (form: AnnualPlanForm) => {
    // 1. Calculate Quarters (F-02)
    const quarters = calculateQuarters(form.selectedMonths)

    // 2. Calculate Resource (F-03)
    const totalMandays = form.auditorCount * form.daysPerAuditor
    const supervisor = supervisors.value.find(s => s.id === form.supervisorId)

    const newPlan: AnnualAuditPlan = {
      id: Date.now().toString(),
      code: form.code,
      version: 'v1.0',
      revisionHistory: [],
      activities: [...form.activities],
      status: form.status,
      selectedMonths: form.selectedMonths.sort((a, b) => a - b),
      quarters: quarters,
      auditorCount: form.auditorCount,
      daysPerAuditor: form.daysPerAuditor,
      totalMandays: totalMandays,
      supervisorId: form.supervisorId,
      supervisorName: supervisor?.name || 'Unknown',
      notes: form.notes,
      year: form.year,
      attachmentCategory: form.attachmentCategory,
      attachments: form.file?.map(f => ({
        name: f.name,
        size: `${(f.size / 1024 / 1024).toFixed(2)} MB`,
        url: URL.createObjectURL(f)
      })) || [],
      attachmentUploadedBy: form.attachmentUploadedBy,
      attachmentUploadDate: form.attachmentUploadDate,
      isActive: form.isActive
    }
    plans.value.unshift(newPlan)
  }

  const updatePlan = (id: string, updatedData: AnnualPlanForm) => {
    const index = plans.value.findIndex(p => p.id === id)
    const isDuplicate = plans.value.some(a => a.code === updatedData.code && a.id !== id)
    if (isDuplicate) throw new Error('Kode Kegiatan sudah digunakan data lain!')
    if (index === -1) return

    const quarters = calculateQuarters(updatedData.selectedMonths)
    const totalMandays = updatedData.auditorCount * updatedData.daysPerAuditor
    const supervisor = supervisors.value.find(s => s.id === updatedData.supervisorId)
    const existingPlan = plans.value[index]!

    plans.value[index] = {
      ...updatedData,
      id: existingPlan.id,
      activities: [...updatedData.activities],
      quarters: quarters,
      totalMandays: totalMandays,
      supervisorName: supervisor?.name || 'Unknown',
      attachmentCategory: updatedData.attachmentCategory,
      // Jika ada file baru, ganti. Jika tidak, pertahankan yang lama.
      attachments: (updatedData.file && updatedData.file.length > 0)
        ? updatedData.file.map(f => ({
          name: f.name,
          size: `${(f.size / 1024 / 1024).toFixed(2)} MB`,
          url: URL.createObjectURL(f)
        }))
        : existingPlan.attachments,
    }

  }

  const deletePlan = (id: string) => {
    const plan = plans.value.find(a => a.id === id)
    if (!plan) return

    if (plan.isUsed) {
      alert('Data ini sudah digunakan dalam RKAT. Hanya status yang akan dinonaktifkan.')
      plan.isActive = false // Soft Delete / Deactivate
    } else {
      if (confirm('Apakah Anda yakin ingin menghapus permanen?')) {
        plans.value = plans.value.filter(a => a.id !== id)
      }
    }
  }

  const createRevision = (planId: string, changesNote: string, user: string) => {
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

    plan.version = newVersion
    plan.status = AnnualAuditPlanStatus.NOT_AVAILABLE
    if (!plan.revisionHistory) plan.revisionHistory = []
    plan.revisionHistory.unshift(historyEntry)

    if (selectedPlan.value && selectedPlan.value.id === planId) {
      selectedPlan.value.version = newVersion
      selectedPlan.value.status = AnnualAuditPlanStatus.NOT_AVAILABLE
      selectedPlan.value.revisionHistory = plan.revisionHistory
    }

    alert(`Revised RKAT created (Version ${newVersion}).`)
  }

  const updatePlanStatus = (id: string, status: AnnualAuditPlanStatus) => {
    const plan = plans.value.find(p => p.id === id)
    if (plan) {
      plan.status = status
    }
  }

  const handleStaffApprove = () => {
    if (selectedPlan.value) {
      selectedPlan.value.status = AnnualAuditPlanStatus.WORK_IN_PROGRESS
      updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.WORK_IN_PROGRESS)
    }
  }

  const handleStaffReject = () => {
    if (selectedPlan.value) {
      selectedPlan.value.staffApprovalNote = ''
    }
  }

  const handleManagerApprove = () => {
    if (selectedPlan.value) {
      selectedPlan.value.status = AnnualAuditPlanStatus.PENDING_APPROVAL
      updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.PENDING_APPROVAL)
    }
  }

  const handleManagerReject = () => {
    if (selectedPlan.value) {
      selectedPlan.value.status = AnnualAuditPlanStatus.NOT_AVAILABLE
      selectedPlan.value.managerApprovalNote = ''
      updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.NOT_AVAILABLE)
    }
  }

  const handleChiefApprove = () => {
    if (selectedPlan.value) {
      selectedPlan.value.status = AnnualAuditPlanStatus.DONE
      updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.DONE)
    }
  }

  const handleChiefReject = () => {
    if (selectedPlan.value) {
      selectedPlan.value.status = AnnualAuditPlanStatus.WORK_IN_PROGRESS
      selectedPlan.value.chiefApprovalNote = ''
      updatePlanStatus(selectedPlan.value.id, AnnualAuditPlanStatus.WORK_IN_PROGRESS)
    }
  }

  return {
    // State
    plans, supervisors, monthsList, yearOptions, supervisorOptions, attachmentCategoryOptions,
    showModal, isEditing, editingId, showViewModal, selectedPlan, progressAudit, approvalStepperItems,
    searchCode, selectedDepartment, selectedStatus, form, columns, errorMsg,

    // Computed
    filteredPlans, totalMandays, selectedSupervisor, quarterAlert, scheduleWarning, utilizationData,
    computedQuarters,
    // Actions
    clearFilters, openViewModal, closeViewModal, toggleMonth, addActivity, removeActivity, handleDownload,
    openModal, closeModal, handleSubmit, handleEdit, handleEditFromView, handleDelete, getSupervisorName,
    getStatusColor, handleFileChange,
    handleStaffApprove, handleStaffReject, handleManagerApprove, handleManagerReject, handleChiefApprove, handleChiefReject,
    createRevision
  }
})