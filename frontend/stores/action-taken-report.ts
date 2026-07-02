import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { AuditCategory, AuditDepartment, AuditStatus, type ActionTakenReport } from '~/types/audit'

export const useActionTakenReportStore = defineStore('action-taken-report', () => {
  // State
  const mockReports: ActionTakenReport[] = [
    {
      id: '1',
      auditRef: 'AUD-2023-009',
      title: 'Ipsum Lorem',
      department: AuditDepartment.FINANCE,
      deadline: '15 Apr 2026',
      status: AuditStatus.COMPLETED,
      auditObject: 'Manajemen Keuangan Utama',
      findingCategory: AuditCategory.ASSURANCE,
      condition: 'Selisih saldo kas 5%',
      criteria: 'SOP Keuangan No. 01',
      recommendation: 'Rekonsiliasi harian',
      pic: 'Departemen Finance',
      attachment: 'Bukti_Rekonsiliasi.pdf',
      progressDescription: 'Selesai tepat waktu.'
    },
    {
      id: '2',
      auditRef: 'AUD-2023-002',
      title: 'Ipsum Lorem',
      department: AuditDepartment.OPS,
      deadline: '15 Apr 2026',
      status: AuditStatus.IN_PROGRESS,
      auditObject: 'Manajemen Aset Tetap',
      findingCategory: AuditCategory.CONSULTING_SERVICES,
      condition: 'Aset tidak terdata',
      criteria: 'SOP Aset No. 12',
      recommendation: 'Inventarisasi ulang',
      pic: 'Departemen GA',
      attachment: 'Draft_Inventarisasi.xlsx',
      progressDescription: 'Sedang berjalan 50%.'
    },
    {
      id: '3',
      auditRef: 'AUD-2023-004',
      title: 'Ipsum Lorem',
      department: AuditDepartment.OPS,
      deadline: '15 Jun 2026',
      status: AuditStatus.COMPLETED,
      auditObject: 'Manajemen Gudang Utama',
      findingCategory: AuditCategory.SPECIAL_AUDIT,
      condition: 'Selisih stok fisik 5% (Gudang A)',
      criteria: 'OP Inventori No. 12',
      recommendation: 'Opname stok ulang & kunci ganda.',
      pic: 'Departemen Logistik',
      attachment: 'Foto_Fisik.jpg',
      progressDescription: 'Masih dalam penghitungan ulang. Kendala: Kurang personil shift malam.'
    },
    {
      id: '4',
      auditRef: 'AUD-2023-006',
      title: 'Ipsum Lorem',
      department: AuditDepartment.OPS,
      deadline: '15 Jun 2026',
      status: AuditStatus.PLANNED
    },
    {
      id: '5',
      auditRef: 'AUD-2023-0027',
      title: 'Ipsum Lorem',
      department: AuditDepartment.HR,
      deadline: '15 Jun 2026',
      status: AuditStatus.IN_PROGRESS
    },
    {
      id: '6',
      auditRef: 'AUD-2026-005',
      title: 'Ipsum Lorem',
      department: AuditDepartment.OPS,
      deadline: '20 Feb 2026',
      status: AuditStatus.CANCELLED,
      auditObject: 'Manajemen Gudang Utama',
      findingCategory: AuditCategory.INVESTIGATION,
      condition: 'Selisih stok fisik 5% (Gudang A)',
      criteria: 'OP Inventori No. 12',
      recommendation: 'Opname stok ulang & kunci ganda.',
      pic: 'Departemen Logistik',
      attachment: 'Foto_Fisik.jpg',
      progressDescription: 'Masih dalam penghitungan ulang. Kendala: Kurang personil shift malam.'
    }
  ]

  const reportList = ref<ActionTakenReport[]>([...mockReports])
  const searchQuery = ref('')
  const selectedDepartment = ref('')
  const selectedStatus = ref('')

  const showModal = ref(false)
  const selectedReport = ref<ActionTakenReport | null>(null)

  const loading = ref(false)
  const errorMsg = ref('')

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/action-taken-reports`, { method: 'GET' })
      let items: ActionTakenReport[] = []
      if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      if (items.length > 0) {
        reportList.value = items
      } else {
        reportList.value = mockReports
      }
    } catch (error) {
      console.error('Failed to fetch action taken reports:', error)
      errorMsg.value = 'Failed to load action taken reports.'
      reportList.value = mockReports
    } finally {
      loading.value = false
    }
  }

  // Load on init
  fetchReports()

  // Computed
  const filteredReports = computed(() => {
    return reportList.value.filter(report => {
      const matchesSearch = (report.auditRef?.toLowerCase() || '').includes(searchQuery.value.toLowerCase()) ||
        (report.title?.toLowerCase() || '').includes(searchQuery.value.toLowerCase())
      const matchesDept = !selectedDepartment.value || report.department === selectedDepartment.value
      const matchesStatus = !selectedStatus.value || report.status === selectedStatus.value
      return matchesSearch && matchesDept && matchesStatus
    })
  })

  const stats = computed(() => {
    const total = reportList.value.length
    if (total === 0) {
      return {
        donePercent: 0,
        wipPercent: 0,
        latePercent: 0
      }
    }
    const done = reportList.value.filter(r => r.status === AuditStatus.COMPLETED).length
    const wip = reportList.value.filter(r => r.status === AuditStatus.IN_PROGRESS).length
    const late = reportList.value.filter(r => r.status === AuditStatus.CANCELLED).length

    return {
      donePercent: Math.round((done / total) * 100),
      wipPercent: Math.round((wip / total) * 100),
      latePercent: Math.round((late / total) * 100)
    }
  })

  // Actions
  const openDetail = (report: ActionTakenReport) => {
    selectedReport.value = report
    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
    selectedReport.value = null
  }

  return {
    reportList,
    searchQuery,
    selectedDepartment,
    selectedStatus,
    showModal,
    selectedReport,
    filteredReports,
    stats,
    openDetail,
    closeModal,
    loading,
    errorMsg,
    fetchReports
  }
})
