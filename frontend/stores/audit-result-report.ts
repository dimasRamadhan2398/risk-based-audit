import { defineStore } from 'pinia'
import { ref, computed, reactive, watch } from 'vue'
import { useAssignmentLetterStore } from './assignment-letter'
import { useWorkingPaperStore } from './working-paper'
import { useAuditFieldworkStore } from './audit-fieldwork'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'

export interface FindingItem {
  title: string
  category: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
  action?: string
  source?: string
  impact?: string
  criteria?: string
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
  const toast = useToastNotification()

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
        category: cat,
        source: f.source || 'Audit Features'
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
      errorMsg.value = extractErrorMessage(error, 'Failed to load audit result reports.')
      reportList.value = [...mockReports]
    } finally {
      loading.value = false
    }
  }

  // Fetch on initialization
  fetchReports()

  const isAutoDetecting = ref(false)

  // Local fallback aggregation from Working Paper and Fieldwork stores
  const getLocalAutoFindings = (targetLetter: string): FindingItem[] => {
    const findings: FindingItem[] = []
    const existingTitles = new Set<string>()

    const wpStore = useWorkingPaperStore()
    const fwStore = useAuditFieldworkStore()

    // 1. Digital Working Paper (AOI & RCA F04 + Plan F05 + Risk F02)
    const allCauses = (wpStore.dataF04 && wpStore.dataF04.length > 0) ? wpStore.dataF04 : (wpStore.mockF04 || [])
    const allPlans = (wpStore.dataF05 && wpStore.dataF05.length > 0) ? wpStore.dataF05 : (wpStore.mockF05 || [])
    const allRisks = (wpStore.dataF02 && wpStore.dataF02.length > 0) ? wpStore.dataF02 : (wpStore.mockF02 || [])

    const stCauses = allCauses.filter((c: any) => (c.workingPaperId || c.assignmentLetterId) === targetLetter)
    const stPlans = allPlans.filter((p: any) => (p.workingPaperId || p.assignmentLetterId) === targetLetter)
    const stRisks = allRisks.filter((r: any) => (r.workingPaperId || r.assignmentLetterId) === targetLetter)

    let defaultCat: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' = 'Significant'
    for (const r of stRisks) {
      const lvl = String(r.riskLevel || '').toUpperCase()
      if (lvl === 'HIGH' || lvl === 'CRITICAL') {
        defaultCat = 'Very Significant'
        break
      } else if (lvl === 'MODERATE' || lvl === 'MEDIUM') {
        defaultCat = 'Significant'
      } else if (lvl === 'LOW') {
        defaultCat = 'Quite Significant'
      }
    }

    stCauses.forEach((cause: any, idx: number) => {
      const cond = (cause.condition || '').trim()
      if (!cond) return
      const key = cond.toLowerCase()
      if (existingTitles.has(key)) return
      existingTitles.add(key)

      let action = ''
      if (stPlans[idx]) {
        action = stPlans[idx]?.actionDescription || stPlans[idx]?.recommendation || ''
      } else if (stPlans.length > 0 && stPlans[0]) {
        action = stPlans[0]?.actionDescription || stPlans[0]?.recommendation || ''
      }

      let cat = defaultCat
      const condLower = cond.toLowerCase()
      if (condLower.includes('kritis') || condLower.includes('overhaul') || condLower.includes('mfa') || condLower.includes('override') || condLower.includes('transisi')) {
        cat = 'Very Significant'
      }

      findings.push({
        title: cond,
        category: cat,
        action,
        source: 'Digital Working Paper (KKA - AOI & RCA)',
        impact: cause.impact || '',
        criteria: cause.criteria || ''
      })
    })

    // 2. Audit Fieldwork Test Controls
    if (typeof (fwStore as any).ensureDataExists === 'function') {
      (fwStore as any).ensureDataExists()
    }
    const testControlsList = (fwStore.fieldworkData && fwStore.fieldworkData[targetLetter]?.testControls?.length)
      ? fwStore.fieldworkData[targetLetter].testControls
      : ((fwStore.mockFieldwork && (fwStore.mockFieldwork as any)[targetLetter]?.testControls) || [])

    testControlsList.forEach((tc: any) => {
      const findingText = (tc.finding || '').trim()
      const resultUpper = (tc.testResult || '').toUpperCase()

      if (findingText || resultUpper === 'INEFFECTIVE' || resultUpper === 'PARTIALLY EFFECTIVE') {
        const title = findingText || `Kelemahan Kontrol: ${tc.controlName || 'Internal Control'}`
        const key = title.toLowerCase()
        if (existingTitles.has(key)) return
        existingTitles.add(key)

        const action = (tc.mitigationPlan || tc.recommendation || '').trim()
        let cat: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant' = 'Significant'
        if (resultUpper === 'INEFFECTIVE') {
          cat = 'Very Significant'
        } else if (resultUpper === 'PARTIALLY EFFECTIVE') {
          cat = 'Significant'
        }

        findings.push({
          title,
          category: cat,
          action,
          source: 'Audit Fieldwork (Test Controls)'
        })
      }
    })

    // Fallback to pre-existing reports if still empty
    if (findings.length === 0) {
      const existingReport = reportList.value.find(r => r.assignmentLetterId === targetLetter)
      if (existingReport && existingReport.findings && existingReport.findings.length > 0) {
        return existingReport.findings.map(f => ({
          ...f,
          source: (f as any).source || 'Audit Record'
        }))
      }
    }

    return findings
  }

  // Fetch auto-findings from Backend with fallback to local stores
  const fetchAutoFindings = async (stNumber?: string): Promise<FindingItem[]> => {
    const targetLetter = stNumber || selectedAssignmentLetter.value || reportForm.assignmentLetterId
    if (!targetLetter) return []

    isAutoDetecting.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const res: any = await $fetch(`${baseUrl}/audit-result-reports/auto-findings?assignmentLetterId=${encodeURIComponent(targetLetter)}`, {
        method: 'GET'
      })
      if (res && res.data && Array.isArray(res.data.findings) && res.data.findings.length > 0) {
        return res.data.findings.map((f: any) => ({
          title: f.title,
          category: f.category || 'Significant',
          action: f.action || '',
          source: f.source || 'Audit Features',
          impact: f.impact || '',
          criteria: f.criteria || ''
        }))
      }
    } catch (err) {
      console.warn('Backend auto-findings API not reachable or returned empty, falling back to local audit stores:', err)
    } finally {
      isAutoDetecting.value = false
    }

    return getLocalAutoFindings(targetLetter)
  }

  const autoPopulateFindings = async (stNumber?: string) => {
    const targetLetter = stNumber || selectedAssignmentLetter.value || reportForm.assignmentLetterId
    if (!targetLetter) return
    const autoFindings = await fetchAutoFindings(targetLetter)
    if (autoFindings.length > 0) {
      reportForm.findings = JSON.parse(JSON.stringify(autoFindings))
      reportForm.findingsCount = autoFindings.length
    }
  }

  const runAutoDetectFindings = async (mode: 'replace' | 'merge' = 'replace') => {
    const targetLetter = reportForm.assignmentLetterId || selectedAssignmentLetter.value
    if (!targetLetter) {
      toast.showWarning('Peringatan', 'Silakan pilih Assignment Letter terlebih dahulu.')
      return
    }
    const detected = await fetchAutoFindings(targetLetter)
    if (detected.length === 0) {
      toast.showWarning('Informasi', `Tidak ada temuan audit baru yang terdeteksi untuk ${targetLetter}.`)
      return
    }

    if (mode === 'replace' || !reportForm.findings || reportForm.findings.length === 0) {
      reportForm.findings = JSON.parse(JSON.stringify(detected))
    } else {
      const existing = new Set(reportForm.findings.map(f => f.title.toLowerCase().trim()))
      detected.forEach(d => {
        if (!existing.has(d.title.toLowerCase().trim())) {
          reportForm.findings.push(JSON.parse(JSON.stringify(d)))
          existing.add(d.title.toLowerCase().trim())
        }
      })
    }
    reportForm.findingsCount = reportForm.findings.length
    toast.showSuccess('Temuan Terisi Otomatis', `${detected.length} temuan berhasil ditarik dari modul KKA & Fieldwork.`)
  }

  // Actions
  const openModal = async () => {
    resetForm()
    isEditing.value = false
    editingId.value = null
    showModal.value = true

    if (selectedAssignmentLetter.value) {
      reportForm.assignmentLetterId = selectedAssignmentLetter.value
      const stData = assignmentLetterStore.assignmentLetterList.find(
        (st: any) => st.letterNumber === selectedAssignmentLetter.value
      )
      if (stData?.auditTitle) {
        reportForm.reportTitle = `Laporan Hasil Audit - ${stData.auditTitle}`
      } else {
        reportForm.reportTitle = `Laporan Hasil Audit - ${selectedAssignmentLetter.value}`
      }
      await autoPopulateFindings(selectedAssignmentLetter.value)
    }
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
        findingsCount: Number(reportForm.findings?.length || reportForm.findingsCount || 0),
        findings: (reportForm.findings || []).map(f => ({
          title: f.title,
          category: f.category,
          action: f.action || '',
          source: (f as any).source || 'Audit Features'
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
      toast.showSuccess('Report saved successfully')
    } catch (error: any) {
      console.error('Failed to save report:', error)
      const detail = extractErrorMessage(error, 'Failed to save report.')
      errorMsg.value = detail
      toast.showError('Failed to save report', detail)
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
    if (!await useGlobalModalStore().confirmDelete({ description: 'Are you sure you want to delete this report?' })) return
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/audit-result-reports/${id}`, {
        method: 'DELETE'
      })
      await fetchReports()
      toast.showSuccess('Report deleted successfully')
    } catch (error: any) {
      console.error('Failed to delete report:', error)
      const detail = extractErrorMessage(error, 'Failed to delete report.')
      errorMsg.value = detail
      toast.showError('Failed to delete report', detail)
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
      const detail = extractErrorMessage(err, 'Gagal mengunduh dokumen Word LHA.')
      toast.showError('Gagal mengunduh dokumen Word LHA', detail)
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
    fetchReports,
    isAutoDetecting,
    fetchAutoFindings,
    runAutoDetectFindings,
    autoPopulateFindings
  }
})
