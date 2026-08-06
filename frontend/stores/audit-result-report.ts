import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'

export interface FindingItem {
  title: string
  category: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
  action?: string
}

export interface AuditResultReport {
  id: string
  reportNumber: string
  assignmentLetterId: string
  reportTitle: string
  executiveSummary: string
  findings?: FindingItem[]
  reportDate: string
  status: 'Draft' | 'Final'
  findingsCount: number
  category?: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
  department?: string
}

export const useAuditResultReportStore = defineStore('audit-result-report', () => {
  const assignmentLetterStore = useAssignmentLetterStore()

  // State
  const selectedAssignmentLetter = ref<string>('')
  const reportList = ref<AuditResultReport[]>([
    {
      id: 'R-001',
      reportNumber: '020/LHA/01/KS IAD/2023',
      assignmentLetterId: '020/ST/01/KSIAD/2023',
      reportTitle: 'Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau',
      executiveSummary: 'Audit Operasional Tahun 2023 di Unit Pelaksana Pengendalian Pembangkitan Kepulauan Riau meliputi ketersediaan pembangkit, K3LH, manajemen risiko, dan pengadaan barang/jasa.',
      category: 'Significant',
      reportDate: '2023-09-22',
      status: 'Final',
      findingsCount: 8,
      findings: [
        { title: 'Pengelolaan Manajemen Risiko Belum Sepenuhnya Sesuai Kebijakan Masa Transisi HSH', category: 'Very Significant', action: 'Perbaikan SOP' },
        { title: 'Pelaksanaan Overhaul UPDK KEPRI Belum Optimal Terjadi PE 6 Hari pada ME+ PLTU TBK #1', category: 'Very Significant', action: 'Evaluasi jadwal' },
        { title: 'Peralatan Lab Milik Perusahaan Belum Digunakan Secara Optimal Sebagai Pembanding Surveyor', category: 'Significant', action: 'Kalibrasi ulang' },
        { title: 'Data Maturity Level Manajemen Aset Belum Lengkap Terbatalnya Fitur Maximo WPC', category: 'Significant', action: 'Update Maximo' },
        { title: 'Terdapat Penyusunan HPS dan Pemanfaatan ERP Tidak Sesuai Ketentuan SCM', category: 'Significant', action: 'Review HPS' },
        { title: 'Program Pemeliharaan Aset Tetap Belum Diakui Kepemilikannya Menggunakan Anggaran Operasi', category: 'Quite Significant', action: 'Inventarisasi aset' },
        { title: 'Terdapat Kontrak Pekerjaan Sejenis Yang Tidak Digabungkan (Strategi Squeezing)', category: 'Quite Significant', action: 'Review kontrak' },
        { title: 'Pengelolaan K3 dan Keamanan di UPDK KEPRI Belum Optimal (Fire Fighting & Lightning)', category: 'Significant', action: 'Audit K3' }
      ]
    },
    {
      id: 'R-002',
      reportNumber: '021/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-001/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Operasional Keuangan 2025',
      executiveSummary: 'Audit dilakukan untuk mengevaluasi efektivitas ICOFR dan kepatuhan terhadap SOP pembayaran.',
      category: 'Significant',
      reportDate: '2026-04-15',
      status: 'Draft',
      findingsCount: 5,
      findings: [
        { title: 'Keterlambatan rekonsiliasi kas harian cabang utama', category: 'Very Significant', action: 'Perbaikan jadwal harian' },
        { title: 'Kelemahan kontrol otorisasi transaksi di atas Rp 500jt', category: 'Very Significant', action: 'Review limit otorisasi' },
        { title: 'Selisih pencatatan inventaris fisik vs buku besar', category: 'Very Significant', action: 'Stok opname ulang' },
        { title: 'Dokumentasi bukti transfer eksternal tidak lengkap', category: 'Significant', action: 'Lengkapi berkas transfer' },
        { title: 'Akses user kasir tidak di-nonaktifkan setelah mutasi', category: 'Significant', action: 'Nonaktifkan akun user' }
      ]
    },
    {
      id: 'R-003',
      reportNumber: '022/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-002/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Keamanan Sistem Informasi & ERP 2026',
      executiveSummary: 'Audit mengevaluasi tata kelola akses pengguna dan keamanan database ERP serta backup data.',
      category: 'Very Significant',
      reportDate: '2026-05-02',
      status: 'Final',
      findingsCount: 4,
      findings: [
        { title: 'Keterlambatan patch keamanan server database ERP', category: 'Very Significant', action: 'Update patch rutin' },
        { title: 'Akses Superadmin ERP belum menggunakan Multi-Factor Authentication', category: 'Very Significant', action: 'Implementasi MFA mandatory' },
        { title: 'Prosedur Backup Data belum diuji pemulihannya secara berkala', category: 'Significant', action: 'Jadwalkan DRC drill' },
        { title: 'Log audit aktivitas sistem informasi belum di-review mingguan', category: 'Quite Significant', action: 'Setup SOC log alert' }
      ]
    },
    {
      id: 'R-004',
      reportNumber: '023/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-003/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Operasional Gudang & Persediaan Logistik 2026',
      executiveSummary: 'Audit mengevaluasi akurasi pencatatan stok gudang persediaan dan pengelolaan distribusi.',
      category: 'Significant',
      reportDate: '2026-08-10',
      status: 'Draft',
      findingsCount: 3,
      findings: [
        { title: 'Selisih fisik barang material persediaan gudang cabang', category: 'Significant', action: 'Investigasi selisih stok' },
        { title: 'Suhu penyimpanan gudang bahan kimia belum terpantau 24/7', category: 'Quite Significant', action: 'Pasang IoT sensor suhu' },
        { title: 'Pengeluaran material proyek tanpa Work Order yang disetujui', category: 'Significant', action: 'Kunci sistem release barang' }
      ]
    },
    {
      id: 'R-005',
      reportNumber: '024/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-004/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Kepatuhan Procurement & SCM 2026',
      executiveSummary: 'Audit evaluasi pelaksanaan rekomendasi audit internal dan kepatuhan pengadaan SCM.',
      category: 'Quite Significant',
      reportDate: '2026-08-18',
      status: 'Final',
      findingsCount: 2,
      findings: [
        { title: 'Penyusunan HPS pengadaan komponen turbin belum melampirkan kertas kerja survei harga', category: 'Quite Significant', action: 'Lampirkan bukti survei HPS' },
        { title: 'Monitoring pencairan jaminan bank vendor belum terintegrasi ERP', category: 'Not Significant', action: 'Fitur reminder otomatis ERP' }
      ]
    },
    {
      id: 'R-006',
      reportNumber: '025/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-005/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit 2026',
      executiveSummary: 'Executive Summary Individual DOC-EXSUM-Q1-2026 untuk Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit.',
      category: 'Very Significant',
      reportDate: '2026-09-15',
      status: 'Final',
      findingsCount: 3,
      findings: [
        { title: 'Inspeksi berkala sistem pemadam kebakaran hidran belum 100% terlaksana', category: 'Very Significant', action: 'Jadwalkan pemeliharaan hidran' },
        { title: 'Sertifikasi K3LH teknisi pemeliharaan pembangkit belum di-renew', category: 'Significant', action: 'Daftarkan pelatihan sertifikasi' },
        { title: 'APBD K3 belum memadai untuk instalasi area berisiko tinggi', category: 'Quite Significant', action: 'Pengadaan APD tambahan' }
      ]
    }
  ])
  const showModal = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const reportForm = reactive({
    reportNumber: '',
    assignmentLetterId: '',
    reportTitle: '',
    reportDate: new Date().toISOString().split('T')[0] as string,
    status: 'Draft' as 'Draft' | 'Final',
    findingsCount: 0,
    findings: [] as FindingItem[]
  })

  // Computed
  const publishedAssignmentLetters = computed(() => {
    return assignmentLetterStore.assignmentLetterList
      .filter((st: any) => st.status === 'Published')
      .map((st: any) => st.letterNumber)
  })

  const filteredReports = computed(() => {
    if (!selectedAssignmentLetter.value) return reportList.value
    return reportList.value.filter(r => r.assignmentLetterId === selectedAssignmentLetter.value)
  })

  const hasSelectedAssignmentLetter = computed(() => !!selectedAssignmentLetter.value && selectedAssignmentLetter.value !== '')

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const loading = ref(false)
  const errorMsg = ref('')

  const mockReports: AuditResultReport[] = [
    {
      id: 'R-001',
      reportNumber: '020/LHA/01/KS IAD/2023',
      assignmentLetterId: '020/ST/01/KSIAD/2023',
      reportTitle: 'Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau',
      executiveSummary: 'Audit Operasional Tahun 2023 di Unit Pelaksana Pengendalian Pembangkitan Kepulauan Riau meliputi ketersediaan pembangkit, K3LH, manajemen risiko, dan pengadaan barang/jasa.',
      category: 'Significant',
      reportDate: '2023-09-22',
      status: 'Final',
      findingsCount: 8,
      findings: [
        { title: 'Pengelolaan Manajemen Risiko Belum Sepenuhnya Sesuai Kebijakan Masa Transisi HSH', category: 'Very Significant', action: 'Perbaikan SOP' },
        { title: 'Pelaksanaan Overhaul UPDK KEPRI Belum Optimal Terjadi PE 6 Hari pada ME+ PLTU TBK #1', category: 'Very Significant', action: 'Evaluasi jadwal' },
        { title: 'Peralatan Lab Milik Perusahaan Belum Digunakan Secara Optimal Sebagai Pembanding Surveyor', category: 'Significant', action: 'Kalibrasi ulang' },
        { title: 'Data Maturity Level Manajemen Aset Belum Lengkap Terbatalnya Fitur Maximo WPC', category: 'Significant', action: 'Update Maximo' },
        { title: 'Terdapat Penyusunan HPS dan Pemanfaatan ERP Tidak Sesuai Ketentuan SCM', category: 'Significant', action: 'Review HPS' },
        { title: 'Program Pemeliharaan Aset Tetap Belum Diakui Kepemilikannya Menggunakan Anggaran Operasi', category: 'Quite Significant', action: 'Inventarisasi aset' },
        { title: 'Terdapat Kontrak Pekerjaan Sejenis Yang Tidak Digabungkan (Strategi Squeezing)', category: 'Quite Significant', action: 'Review kontrak' },
        { title: 'Pengelolaan K3 dan Keamanan di UPDK KEPRI Belum Optimal (Fire Fighting & Lightning)', category: 'Significant', action: 'Audit K3' }
      ]
    },
    {
      id: 'R-002',
      reportNumber: '021/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-001/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Operasional Keuangan 2025',
      executiveSummary: 'Audit dilakukan untuk mengevaluasi efektivitas ICOFR dan kepatuhan terhadap SOP pembayaran.',
      category: 'Significant',
      reportDate: '2026-04-15',
      status: 'Draft',
      findingsCount: 5,
      findings: [
        { title: 'Keterlambatan rekonsiliasi kas harian cabang utama', category: 'Very Significant', action: 'Perbaikan jadwal harian' },
        { title: 'Kelemahan kontrol otorisasi transaksi di atas Rp 500jt', category: 'Very Significant', action: 'Review limit otorisasi' },
        { title: 'Selisih pencatatan inventaris fisik vs buku besar', category: 'Very Significant', action: 'Stok opname ulang' },
        { title: 'Dokumentasi bukti transfer eksternal tidak lengkap', category: 'Significant', action: 'Lengkapi berkas transfer' },
        { title: 'Akses user kasir tidak di-nonaktifkan setelah mutasi', category: 'Significant', action: 'Nonaktifkan akun user' }
      ]
    },
    {
      id: 'R-003',
      reportNumber: '022/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-002/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Keamanan Sistem Informasi & ERP 2026',
      executiveSummary: 'Audit mengevaluasi tata kelola akses pengguna dan keamanan database ERP serta backup data.',
      category: 'Very Significant',
      reportDate: '2026-05-02',
      status: 'Final',
      findingsCount: 4,
      findings: [
        { title: 'Keterlambatan patch keamanan server database ERP', category: 'Very Significant', action: 'Update patch rutin' },
        { title: 'Akses Superadmin ERP belum menggunakan Multi-Factor Authentication', category: 'Very Significant', action: 'Implementasi MFA mandatory' },
        { title: 'Prosedur Backup Data belum diuji pemulihannya secara berkala', category: 'Significant', action: 'Jadwalkan DRC drill' },
        { title: 'Log audit aktivitas sistem informasi belum di-review mingguan', category: 'Quite Significant', action: 'Setup SOC log alert' }
      ]
    },
    {
      id: 'R-004',
      reportNumber: '023/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-003/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Operasional Gudang & Persediaan Logistik 2026',
      executiveSummary: 'Audit mengevaluasi akurasi pencatatan stok gudang persediaan dan pengelolaan distribusi.',
      category: 'Significant',
      reportDate: '2026-08-10',
      status: 'Draft',
      findingsCount: 3,
      findings: [
        { title: 'Selisih fisik barang material persediaan gudang cabang', category: 'Significant', action: 'Investigasi selisih stok' },
        { title: 'Suhu penyimpanan gudang bahan kimia belum terpantau 24/7', category: 'Quite Significant', action: 'Pasang IoT sensor suhu' },
        { title: 'Pengeluaran material proyek tanpa Work Order yang disetujui', category: 'Significant', action: 'Kunci sistem release barang' }
      ]
    },
    {
      id: 'R-005',
      reportNumber: '024/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-004/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit Kepatuhan Procurement & SCM 2026',
      executiveSummary: 'Audit evaluasi pelaksanaan rekomendasi audit internal dan kepatuhan pengadaan SCM.',
      category: 'Quite Significant',
      reportDate: '2026-08-18',
      status: 'Final',
      findingsCount: 2,
      findings: [
        { title: 'Penyusunan HPS pengadaan komponen turbin belum melampirkan kertas kerja survei harga', category: 'Quite Significant', action: 'Lampirkan bukti survei HPS' },
        { title: 'Monitoring pencairan jaminan bank vendor belum terintegrasi ERP', category: 'Not Significant', action: 'Fitur reminder otomatis ERP' }
      ]
    },
    {
      id: 'R-006',
      reportNumber: '025/LHA/01/KS IAD/2026',
      assignmentLetterId: 'ST-005/SKAI/2026',
      reportTitle: 'Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit 2026',
      executiveSummary: 'Executive Summary Individual DOC-EXSUM-Q1-2026 untuk Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit.',
      category: 'Very Significant',
      reportDate: '2026-09-15',
      status: 'Final',
      findingsCount: 3,
      findings: [
        { title: 'Inspeksi berkala sistem pemadam kebakaran hidran belum 100% terlaksana', category: 'Very Significant', action: 'Jadwalkan pemeliharaan hidran' },
        { title: 'Sertifikasi K3LH teknisi pemeliharaan pembangkit belum di-renew', category: 'Significant', action: 'Daftarkan pelatihan sertifikasi' },
        { title: 'APBD K3 belum memadai untuk instalasi area berisiko tinggi', category: 'Quite Significant', action: 'Pengadaan APD tambahan' }
      ]
    }
  ]

  const mapReportItem = (item: any): AuditResultReport => {
    let dateVal = item.reportDate || item.report_date || item.created_at || ''
    if (typeof dateVal === 'string' && dateVal.includes('T')) {
      dateVal = dateVal.split('T')[0]
    }

    const findingsArr = item.findings || item.Findings || []

    // Map legacy severity to category
    const mappedFindings = findingsArr.map((f: any) => {
      let cat = f.category || f.severity || 'Quite Significant'
      if (cat === 'Moderately Significant') cat = 'Quite Significant'
      if (cat === 'Insignificant') cat = 'Not Significant'
      return {
        ...f,
        category: cat
      }
    })

    return {
      ...item,
      reportNumber: item.reportNumber || item.report_number || '020/LHA/01/KS IAD/2023',
      findingsCount: item.findingsCount || item.findings_count || mappedFindings.length || 0,
      findings: mappedFindings,
      reportDate: dateVal || new Date().toISOString().split('T')[0]
    }
  }

  const fetchReports = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/audit-result-reports`, { method: 'GET' })
      let items: any[] = []
      if (response && response.data && Array.isArray(response.data.items)) {
        items = response.data.items
      } else if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      if (items.length > 0) {
        reportList.value = items.map(mapReportItem)
      } else {
        reportList.value = [...mockReports]
      }
    } catch (error) {
      console.error('Failed to fetch reports, falling back to mock data:', error)
      errorMsg.value = 'Failed to load audit result reports.'
      reportList.value = [...mockReports]
    } finally {
      loading.value = false
    }
  }

  // Fetch on initialization
  fetchReports()

  // Actions
  const openModal = () => {
    resetForm()
    isEditing.value = false
    editingId.value = null
    showModal.value = true
  }

  const closeModal = () => {
    showModal.value = false
  }

  const resetForm = () => {
    Object.assign(reportForm, {
      assignmentLetterId: selectedAssignmentLetter.value,
      reportTitle: '',
      reportDate: new Date().toISOString().split('T')[0] as string,
      status: 'Draft',
      findingsCount: 0,
      findings: []
    })
  }

  const saveReport = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload: any = {
        assignmentLetterId: reportForm.assignmentLetterId || selectedAssignmentLetter.value || 'ST-001/SKAI/2026',
        reportTitle: reportForm.reportTitle,
        reportDate: reportForm.reportDate,
        report_date: reportForm.reportDate,
        reportNumber: reportForm.reportNumber,
        findingsCount: Number(reportForm.findingsCount || reportForm.findings?.length || 0),
        findings: (reportForm.findings || []).map(f => ({
          title: f.title,
          category: f.category,
          action: f.action || ''
        })),
        status: reportForm.status
      }
      if (isEditing.value && editingId.value) {
        await $fetch(`${baseUrl}/audit-result-reports/${editingId.value}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        await $fetch(`${baseUrl}/audit-result-reports`, {
          method: 'POST',
          body: payload
        })
      }
      closeModal()
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to save report:', error)
      alert('Failed to save report: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  const editReport = (report: AuditResultReport) => {
    Object.assign(reportForm, {
      ...report,
      findings: report.findings ? JSON.parse(JSON.stringify(report.findings)) : []
    })
    isEditing.value = true
    editingId.value = report.id
    showModal.value = true
  }

  const deleteReport = async (id: string) => {
    if (!confirm('Are you sure you want to delete this report?')) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/audit-result-reports/${id}`, {
        method: 'DELETE'
      })
      await fetchReports()
    } catch (error: any) {
      console.error('Failed to delete report:', error)
      alert('Failed to delete report: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  const downloadDocx = async (id: string, reportNumber: string) => {
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const blob = await $fetch<Blob>(`${baseUrl}/audit-result-reports/${id}/download-docx`, {
        method: 'GET',
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      const cleanNum = (reportNumber || 'LHA').replace(/[\/\s]/g, '_')
      a.download = `LHA_${cleanNum}.docx`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      window.URL.revokeObjectURL(url)
    } catch (err: any) {
      console.error('Failed to download docx:', err)
      alert('Gagal mengunduh dokumen Word LHA: ' + (err.message || err))
    }
  }

  const syncExecutiveSummaryField = async (reportNumber: string, narrative: string, findingsCount?: number) => {
    const found = reportList.value.find(r => (r.reportNumber || (r as any).report_number) === reportNumber)
    if (found) {
      found.executiveSummary = narrative
      if (typeof findingsCount === 'number' && findingsCount > 0) {
        found.findingsCount = findingsCount
      }
      try {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/audit-result-reports/${found.id}`, {
          method: 'PUT',
          body: {
            executive_summary: narrative,
            findingsCount: found.findingsCount
          }
        })
      } catch (e) {
        console.warn('Silent sync to backend failed:', e)
      }
    }
  }

  const clearExecutiveSummaryField = async (reportNumber: string) => {
    const found = reportList.value.find(r => (r.reportNumber || (r as any).report_number) === reportNumber)
    if (found) {
      found.executiveSummary = ''
      try {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/audit-result-reports/${found.id}`, {
          method: 'PUT',
          body: {
            executive_summary: ''
          }
        })
      } catch (e) {
        console.warn('Silent clear executive summary failed:', e)
      }
    }
  }

  return {
    selectedAssignmentLetter,
    reportList,
    showModal,
    isEditing,
    reportForm,
    publishedAssignmentLetters,
    filteredReports,
    hasSelectedAssignmentLetter,
    openModal,
    closeModal,
    saveReport,
    editReport,
    deleteReport,
    downloadDocx,
    syncExecutiveSummaryField,
    clearExecutiveSummaryField,
    loading,
    errorMsg,
    fetchReports
  }
})
