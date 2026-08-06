import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { AuditCategory, AuditDepartment, AuditStatus, type ActionTakenReport } from '~/types/audit'

export const useActionTakenReportStore = defineStore('action-taken-report', () => {
  // State
  const mockReports: ActionTakenReport[] = [
    {
      id: '1',
      auditRef: 'ST-001/SKAI/2026',
      title: 'Rekonsiliasi Kas Harian dan Arus Kas',
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
      progressDescription: 'Selesai tepat waktu.',
      assignmentLetter: {
        id: 'letter-001',
        letterNumber: 'ST-001/SKAI/2026',
        auditTitle: 'Audit Pengendalian Keuangan & Pengeluaran Kas Q1 2026',
        leader: 'Zeta Ramadhani',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-03-01',
        finishPeriod: '2026-03-31',
        workingUnit: 'Finance',
        executionPeriod: '2026-03-01 to 2026-03-31',
        auditPurpose: 'Annual Audit',
        membersList: [{ name: 'Zeta Ramadhani', role: 'Chairperson' }],
        purposeList: [],
        scopeList: [],
        ccList: [],
        status: 'Published',
        createdAt: '2026-01-01'
      }
    },
    {
      id: '2',
      auditRef: 'ST-002/SKAI/2026',
      title: 'Pemasangan SMTP Alert Backup ERP',
      department: AuditDepartment.IT,
      deadline: '20 Apr 2026',
      status: AuditStatus.COMPLETED,
      auditObject: 'IT Infrastructure',
      findingCategory: AuditCategory.ASSURANCE,
      condition: 'Kegagalan backup database tidak memicu notifikasi otomatis.',
      criteria: 'SOP IT Recovery mensyaratkan notifikasi insiden langsung terkirim ke tim sysadmin.',
      recommendation: 'Konfigurasi SMTP server untuk mengirim log gagal.',
      pic: 'Rudi Hermawan',
      attachment: 'Config_Alert_SMTP.pdf',
      progressDescription: 'SMTP alert dikonfigurasi dan sukses diuji coba.',
      assignmentLetter: {
        id: 'letter-002',
        letterNumber: 'ST-002/SKAI/2026',
        auditTitle: 'Audit Keamanan Sistem Informasi & Infrastruktur ERP 2026',
        leader: 'Andi Firmansyah',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-04-01',
        finishPeriod: '2026-04-30',
        workingUnit: 'IT',
        executionPeriod: '2026-04-01 to 2026-04-30',
        auditPurpose: 'IT Security Audit',
        membersList: [{ name: 'Andi Firmansyah', role: 'Chairperson' }],
        purposeList: [],
        scopeList: [],
        ccList: [],
        status: 'Published',
        createdAt: '2026-01-10'
      }
    },
    {
      id: '3',
      auditRef: 'ST-003/SKAI/2026',
      title: 'Opname Stok Fisik Gudang Utama',
      department: AuditDepartment.OPS,
      deadline: '15 May 2026',
      status: AuditStatus.IN_PROGRESS,
      auditObject: 'Manajemen Gudang Utama',
      findingCategory: AuditCategory.SPECIAL_AUDIT,
      condition: 'Selisih stok fisik 5% (Gudang A)',
      criteria: 'SOP Inventori No. 12',
      recommendation: 'Opname stok ulang & kunci ganda.',
      pic: 'Departemen Logistik',
      attachment: 'Draft_Inventarisasi.xlsx',
      progressDescription: 'Sedang berjalan 50%.',
      assignmentLetter: {
        id: 'letter-003',
        letterNumber: 'ST-003/SKAI/2026',
        auditTitle: 'Audit Operasional Branch, Gudang & Logistik 2026',
        leader: 'Rina Wulandari',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-07-01',
        finishPeriod: '2026-07-31',
        workingUnit: 'Operations',
        executionPeriod: '2026-07-01 to 2026-07-31',
        auditPurpose: 'Operational Audit',
        membersList: [{ name: 'Rina Wulandari', role: 'Chairperson' }],
        purposeList: [],
        scopeList: [],
        ccList: [],
        status: 'Published',
        createdAt: '2026-01-25'
      }
    },
    {
      id: '4',
      auditRef: 'ST-004/SKAI/2026',
      title: 'Evaluasi Kontrak Vendor SCM',
      department: AuditDepartment.OPS,
      deadline: '01 Jun 2026',
      status: AuditStatus.PLANNED,
      auditObject: 'Manajemen Vendor',
      findingCategory: AuditCategory.CONSULTING_SERVICES,
      condition: 'Dokumen HPS vendor belum diperbarui',
      criteria: 'SOP Pengadaan No. 05',
      recommendation: 'Review dan perbarui HPS vendor secara periodik',
      pic: 'Departemen GA',
      progressDescription: 'Dalam perancangan.',
      assignmentLetter: {
        id: 'letter-004',
        letterNumber: 'ST-004/SKAI/2026',
        auditTitle: 'Audit Kepatuhan Manajemen Risiko & Pengadaan Barang/Jasa 2026',
        leader: 'Budi Santoso',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-08-01',
        finishPeriod: '2026-08-31',
        workingUnit: 'Procurement',
        executionPeriod: '2026-08-01 to 2026-08-31',
        auditPurpose: 'Compliance Audit',
        membersList: [{ name: 'Budi Santoso', role: 'Chairperson' }],
        purposeList: [],
        scopeList: [],
        ccList: [],
        status: 'Published',
        createdAt: '2026-01-26'
      }
    },
    {
      id: '5',
      auditRef: 'ST-005/SKAI/2026',
      title: 'Sertifikasi Peralatan Pemadam Hidran',
      department: AuditDepartment.OPS,
      deadline: '15 Jun 2026',
      status: AuditStatus.CANCELLED,
      auditObject: 'Fasilitas K3LH',
      findingCategory: AuditCategory.INVESTIGATION,
      condition: 'Masa berlaku sertifikasi hidran berakhir',
      criteria: 'SOP K3LH No. 08',
      recommendation: 'Pengujian dan resertifikasi hidran pabrik',
      pic: 'K3LH Team',
      attachment: 'Foto_Fisik.jpg',
      progressDescription: 'Dibatalkan karena restrukturisasi operasional unit.',
      assignmentLetter: {
        id: 'letter-005',
        letterNumber: 'ST-005/SKAI/2026',
        auditTitle: 'Audit Pengendalian K3LH & Manufaktur Pembangkit 2026',
        leader: 'Dewi Kusumawati',
        category: AuditCategory.ASSURANCE,
        auditYear: '2026',
        auditTeam: 'SKAI',
        startPeriod: '2026-09-01',
        finishPeriod: '2026-09-30',
        workingUnit: 'Maintenance',
        executionPeriod: '2026-09-01 to 2026-09-30',
        auditPurpose: 'HSE Audit',
        membersList: [{ name: 'Dewi Kusumawati', role: 'Chairperson' }],
        purposeList: [],
        scopeList: [],
        ccList: [],
        status: 'Published',
        createdAt: '2026-01-27'
      }
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
      let items: any[] = []
      if (response && response.data && Array.isArray(response.data.items)) {
        items = response.data.items
      } else if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      if (items.length > 0) {
        reportList.value = items.map((item: any) => ({
          ...item,
          assignmentLetter: item.assignment_letter || item.assignmentLetter
        }))
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
