import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import type {
  WorkingPaperHeaderForm, WorkingPaperRiskForm, WorkingPaperSampleForm,
  WorkingPaperCauseForm, WorkingPaperPlanForm, SampleItem, WorkingPaperHeader,
  WorkingPaperRisk,
  WorkingPaperSample,
  WorkingPaperCause,
  WorkingPaperPlan,
} from '~/types/audit'
import { ROOT_CAUSE_METHOD_OPTIONS, TEST_RESULT_OPTIONS } from '~/types/audit'
import { useAuditFieldworkStore } from './audit-fieldwork'
import { computed } from 'vue'
import { RiskLevel, RiskTaxonomy } from '../types/risk'
import type { StepperItem } from '@nuxt/ui'

import { z } from 'zod'

export const headerSchema = z.object({
  assignmentLetterId: z.string().min(1, 'Required'),
  businessProcess: z.string().min(1, 'Required'),
  periodStart: z.string().min(1, 'Required'),
  periodEnd: z.string().min(1, 'Required'),
  location: z.string().min(1, 'Required')
})

export const riskSchema = z.object({
  risk: z.string().min(1, 'Required'),
  taxonomy: z.string().min(1, 'Required'),
  riskLevel: z.string().min(1, 'Required'),
  controlDescription: z.string().min(1, 'Required')
})

export const sampleSchema = z.object({
  population: z.number().min(1, 'Required'),
  sampleSize: z.number().min(1, 'Required'),
  conclusion: z.string().min(1, 'Required')
})

export const causeSchema = z.object({
  condition: z.string().min(1, 'Required'),
  criteria: z.string().min(1, 'Required'),
  impact: z.string().min(1, 'Required')
})

export const planSchema = z.object({
  recommendation: z.string().min(1, 'Required'),
  response: z.string().min(1, 'Required'),
  actionDescription: z.string().min(1, 'Required'),
  pic: z.string().min(1, 'Required'),
  periodAction: z.string().min(1, 'Required')
})

export const useWorkingPaperStore = defineStore('working-paper', () => {
  const fieldworkStore = useAuditFieldworkStore()

  const fileInput = ref<HTMLInputElement | null>(null)

  // Picu klik pada input file yang tersembunyi
  const triggerUpload = () => {
    fileInput.value?.click()
  }

  // Handle perubahan file
  const onFileChange = (e: Event) => {
    const target = e.target as HTMLInputElement
    const files = target.files

    if (files && files[0]) {
      const file = files[0]

      // Validasi Ukuran (Contoh: 10MB)
      if (file.size > 10 * 1024 * 1024) {
        alert('File size too large! Maximum 10MB.')
        return
      }

      // Simpan ke Store
      causeForm.evidenceFile = file
      console.log('File chosen:', file.name)
    }
  }

  // Hapus file
  const removeFile = () => {
    causeForm.evidenceFile = null
    if (fileInput.value) {
      fileInput.value.value = '' // Reset input agar bisa upload file yang sama lagi
    }
  }

  // Fungsi untuk mendapatkan daftar anggota yang tersedia untuk baris tertentu
  const getAvailableMembers = (currentIndex: number) => {
    // 1. Ambil semua nama yang sudah dipilih di baris-baris LAIN
    const selectedNames = headerForm.teamMembers
      .filter((_, index) => index !== currentIndex) // Kecualikan baris yang sedang aktif
      .map(member => member.name)
      .filter(name => !!name) // Hanya ambil yang sudah ada isinya

    // 2. Filter master list PIC agar tidak menyertakan nama yang sudah dipilih di baris lain
    return options.pic.filter(pic => !selectedNames.includes(pic))
  }

  const isDateError = computed(() => {
    // Jika salah satu tanggal belum diisi, jangan anggap error dulu
    if (!headerForm.periodStart || !headerForm.periodEnd) return false

    const start = new Date(headerForm.periodStart)
    const end = new Date(headerForm.periodEnd)

    // Return true jika tanggal akhir LEBIH KECIL dari tanggal mulai
    return end < start
  })

  // Opsional: Pesan error dinamis
  const dateErrorMessage = computed(() => {
    return isDateError.value ? "The end date cannot be before the start date." : false
  })

  // Tabs Configuration
  const tabs = [
    { label: 'Header', slot: 'f01', icon: 'i-heroicons-document-text' },
    { label: 'Risk Profile', slot: 'f02', icon: 'i-heroicons-shield-exclamation' },
    { label: 'Test Sample', slot: 'f03', icon: 'i-heroicons-table-cells' },
    { label: 'AOI & RCA', slot: 'f04', icon: 'i-heroicons-magnifying-glass-circle' },
    { label: 'Action Plan', slot: 'f05', icon: 'i-heroicons-check-badge' }
  ]

  const workingItems = [
    {
      slot: 'f01' as const,
      title: 'Header',
      description: 'Add your header',
      icon: 'i-lucide-house'
    }, {
      slot: 'f02' as const,
      title: 'Risk Profile',
      description: 'Add your risk profile',
      icon: 'i-lucide-shield'
    }, {
      slot: 'f03' as const,
      title: 'Test Sample',
      description: 'Add your test sample',
      icon: 'i-lucide-table'
    }, {
      slot: 'f04' as const,
      title: 'AOI & RCA',
      description: 'Add your AOI & RCA',
      icon: 'i-lucide-file-search'
    }, {
      slot: 'f05' as const,
      title: 'Action Plan',
      description: 'Add your action plan',
      icon: 'i-lucide-check'
    }
  ] satisfies StepperItem[]

  const columnsF01 = [
    { accessorKey: 'assignmentLetterId', header: 'Assignment Letter' },
    { accessorKey: 'businessProcess', header: 'Business Process' },
    { accessorKey: 'period', header: 'Period' },
    { accessorKey: 'location', header: 'Location' },
    { accessorKey: 'teamMembers', header: 'Team' },
    { accessorKey: 'actions', header: 'Action' }
  ]

  const columnsF02 = [
    { accessorKey: 'risk', header: 'Risk' },
    { accessorKey: 'taxonomy', header: 'Risk Category' },
    { accessorKey: 'riskLevel', header: 'Risk Level' },
    { accessorKey: 'controlDescription', header: 'Control Description' },
    { accessorKey: 'actions', header: 'Action' }
  ]

  const columnsF03 = [
    { accessorKey: 'population', header: 'Population' },
    { accessorKey: 'sampleSize', header: 'Sample Size' },
    { accessorKey: 'samples', header: 'Sample List' },
    { accessorKey: 'conclusion', header: 'Conclusion' },
    { accessorKey: 'actions', header: 'Action' }
  ]

  const columnsF04 = [
    { accessorKey: 'condition', header: 'Condition' },
    { accessorKey: 'criteria', header: 'Criteria' },
    { accessorKey: 'impact', header: 'Impact' },
    { accessorKey: 'rootCause', header: 'Root Cause' },
    { accessorKey: 'evidenceFile', header: 'Evidence Document' },
    { accessorKey: 'actions', header: 'Action' }
  ]

  const columnsF05 = [
    { accessorKey: 'recommendation', header: 'Recommendation' },
    { accessorKey: 'response', header: 'Auditee Response' },
    { accessorKey: 'actionDescription', header: 'Description' },
    { accessorKey: 'pic', header: 'PIC' },
    { accessorKey: 'periodAction', header: 'Target Selesai' },
    { accessorKey: 'actions', header: 'Actions' },
  ]
  // --- STATE ---
  const headerForm = reactive<WorkingPaperHeaderForm>({
    assignmentLetterId: '',
    auditPurpose: '',
    businessProcess: '',
    periodStart: '',
    periodEnd: '',
    location: '',
    teamMembers: [
      { id: Date.now(), name: '', role: '' } // Inisialisasi 1 baris kosong
    ],
  })

  const riskForm = reactive<WorkingPaperRiskForm>({
    risk: '',
    taxonomy: RiskTaxonomy.Financial,
    riskLevel: RiskLevel.HIGH,
    controlDescription: '',
  })

  const sampleForm = reactive<WorkingPaperSampleForm>({
    population: null,
    sampleSize: null,
    samples: [
      { id: Date.now(), document: '', l1: undefined, l2: undefined, l3: undefined }
    ],
    conclusion: '',
  })

  const causeForm = reactive<WorkingPaperCauseForm>({
    condition: '',
    criteria: '',
    impact: '',
    evidenceFile: null,
    rootCause: [
      { id: Date.now(), method: 'People', w1: '', w2: '', w3: '' }
    ],
  })

  const planForm = reactive<WorkingPaperPlanForm>({
    recommendation: '',
    response: '',
    actionDescription: '',
    pic: '',
    periodAction: ''
  })

  const dataF01 = ref<WorkingPaperHeader[]>([])
  const dataF02 = ref<WorkingPaperRisk[]>([])
  const dataF03 = ref<WorkingPaperSample[]>([])
  const dataF04 = ref<WorkingPaperCause[]>([])
  const dataF05 = ref<WorkingPaperPlan[]>([])



  const loading = ref(false)
  const errorMsg = ref('')

  const extractItems = (res: any) => {
    if (res?.items && Array.isArray(res.items)) return res.items
    if (res?.data?.items && Array.isArray(res.data.items)) return res.data.items
    if (Array.isArray(res?.data)) return res.data
    if (Array.isArray(res)) return res
    return []
  }

  const fetchAllData = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()

      const resF01: any = await $fetch(`${baseUrl}/working-papers/headers`, { method: 'GET' })
      dataF01.value = extractItems(resF01)

      const resF02: any = await $fetch(`${baseUrl}/working-papers/risks`, { method: 'GET' })
      dataF02.value = extractItems(resF02)

      const resF03: any = await $fetch(`${baseUrl}/working-papers/samples`, { method: 'GET' })
      dataF03.value = extractItems(resF03)

      const resF04: any = await $fetch(`${baseUrl}/working-papers/causes`, { method: 'GET' })
      dataF04.value = extractItems(resF04)

      const resF05: any = await $fetch(`${baseUrl}/working-papers/plans`, { method: 'GET' })
      dataF05.value = extractItems(resF05)
    } catch (error) {
      console.error('Failed to fetch working papers:', error)
    } finally {
      loading.value = false
    }
  }

  const mockF01: WorkingPaperHeader[] = [
    { id: 'WP-H-001', assignmentLetterId: 'ST-001/SKAI/2026', auditPurpose: 'Annual Audit', businessProcess: 'Finance & Cash', period: '2026-03-01 s/d 2026-03-31', location: 'Head Office', teamMembers: [{ id: 1, name: 'Zeta Ramadhani', role: 'Chairperson' }] },
    { id: 'WP-H-002', assignmentLetterId: 'ST-002/SKAI/2026', auditPurpose: 'IT Security Audit', businessProcess: 'IT Operations & ERP', period: '2026-04-01 s/d 2026-04-30', location: 'Data Center', teamMembers: [{ id: 2, name: 'Andi Firmansyah', role: 'Chairperson' }] },
    { id: 'WP-H-003', assignmentLetterId: 'ST-003/SKAI/2026', auditPurpose: 'Operational Audit', businessProcess: 'Warehouse & Supply Chain', period: '2026-07-01 s/d 2026-07-31', location: 'Gudang Pusat', teamMembers: [{ id: 3, name: 'Rina Wulandari', role: 'Chairperson' }] },
    { id: 'WP-H-004', assignmentLetterId: 'ST-004/SKAI/2026', auditPurpose: 'Compliance Audit', businessProcess: 'Procurement', period: '2026-08-01 s/d 2026-08-31', location: 'Head Office', teamMembers: [{ id: 4, name: 'Budi Santoso', role: 'Chairperson' }] },
    { id: 'WP-H-005', assignmentLetterId: 'ST-005/SKAI/2026', auditPurpose: 'HSE Audit', businessProcess: 'K3LH & Maintenance', period: '2026-09-01 s/d 2026-09-30', location: 'Pembangkit PLTU', teamMembers: [{ id: 5, name: 'Dewi Kusumawati', role: 'Chairperson' }] },
    { id: 'WP-H-020', assignmentLetterId: '020/ST/01/KSIAD/2023', auditPurpose: 'Operational Audit', businessProcess: 'O&M Pembangkit', period: 'Januari 2023 s.d Agustus 2023', location: 'UPDK Kepulauan Riau', teamMembers: [{ id: 6, name: 'Tomy Afrilianto', role: 'Chairperson' }] }
  ]

  const mockF02: WorkingPaperRisk[] = [
    { id: 'WP-R-001', workingPaperId: 'ST-001/SKAI/2026', risk: 'FIN-001: Selisih pencatatan kas', taxonomy: RiskTaxonomy.Financial, riskLevel: RiskLevel.HIGH, controlDescription: 'Otorisasi transaksi kas di atas Rp 500jt' },
    { id: 'WP-R-002', workingPaperId: 'ST-002/SKAI/2026', risk: 'SEC-001: Serangan siber database ERP', taxonomy: RiskTaxonomy.Operational, riskLevel: RiskLevel.HIGH, controlDescription: 'Penggunaan MFA dan patch keamanan berkala' },
    { id: 'WP-R-003', workingPaperId: 'ST-003/SKAI/2026', risk: 'OPS-001: Selisih fisik stok gudang', taxonomy: RiskTaxonomy.Operational, riskLevel: RiskLevel.HIGH, controlDescription: 'Stock opname fisik bulanan' },
    { id: 'WP-R-004', workingPaperId: 'ST-004/SKAI/2026', risk: 'COM-001: Ketidakpatuhan SOP HPS Procurement', taxonomy: RiskTaxonomy.Operational, riskLevel: RiskLevel.MODERATE, controlDescription: 'Reviu kertas kerja survei HPS' },
    { id: 'WP-R-005', workingPaperId: 'ST-005/SKAI/2026', risk: 'ENV-001: Kerusakan fasilitas K3LH hidran', taxonomy: RiskTaxonomy.Operational, riskLevel: RiskLevel.HIGH, controlDescription: 'Inspeksi rutin hidran & APD' },
    { id: 'WP-R-020', workingPaperId: '020/ST/01/KSIAD/2023', risk: 'OPS-020: Overhaul PLTU PE 6 hari', taxonomy: RiskTaxonomy.Operational, riskLevel: RiskLevel.HIGH, controlDescription: 'Penyusunan DMR & sertifikasi O&M' }
  ]

  const mockF03: WorkingPaperSample[] = [
    { id: 'WP-S-001', workingPaperId: 'ST-001/SKAI/2026', population: 150, sampleSize: 20, samples: [{ id: 1, document: 'Voucher Kas Q1', l1: 'Pass', l2: 'Pass', l3: 'Fail' }], conclusion: 'Pengendalian kas memadai dengan catatan keterlambatan rekonsiliasi' },
    { id: 'WP-S-002', workingPaperId: 'ST-002/SKAI/2026', population: 80, sampleSize: 15, samples: [{ id: 2, document: 'Log User ERP', l1: 'Pass', l2: 'Fail', l3: 'Fail' }], conclusion: 'Autentikasi superadmin perlu ditingkatkan dengan MFA' },
    { id: 'WP-S-003', workingPaperId: 'ST-003/SKAI/2026', population: 300, sampleSize: 30, samples: [{ id: 3, document: 'Kartu Stok Gudang', l1: 'Pass', l2: 'Pass', l3: 'Pass' }], conclusion: 'Fisik barang sesuai dengan beberapa catatan selisih kecil' },
    { id: 'WP-S-004', workingPaperId: 'ST-004/SKAI/2026', population: 45, sampleSize: 10, samples: [{ id: 4, document: 'Dokumen HPS Vendor', l1: 'Pass', l2: 'Pass', l3: 'Fail' }], conclusion: 'Evaluasi HPS memenuhi syarat administrasi' },
    { id: 'WP-S-005', workingPaperId: 'ST-005/SKAI/2026', population: 50, sampleSize: 12, samples: [{ id: 5, document: 'Checklist Inspeksi K3', l1: 'Pass', l2: 'Pass', l3: 'Pass' }], conclusion: 'Peralatan K3 memadai' },
    { id: 'WP-S-020', workingPaperId: '020/ST/01/KSIAD/2023', population: 120, sampleSize: 25, samples: [{ id: 6, document: 'Laporan Overhaul ME+ PLTU', l1: 'Pass', l2: 'Fail', l3: 'Fail' }], conclusion: 'Overhaul terlaksana dengan catatan waktu ketersediaan' }
  ]

  const mockF04: WorkingPaperCause[] = [
    { id: 'WP-C-001', workingPaperId: 'ST-001/SKAI/2026', condition: 'Rekonsiliasi kas harian terlambat', criteria: 'SOP Kas Bab 4', impact: 'Risiko selisih kas', evidenceFile: null, rootCause: [{ id: 1, method: 'People', w1: 'Kasir belum input harian', w2: 'Beban kerja tinggi', w3: 'Belum ada sistem alert' }] },
    { id: 'WP-C-002', workingPaperId: 'ST-002/SKAI/2026', condition: 'Superadmin ERP tanpa MFA', criteria: 'Kebijakan Keamanan TI 2025', impact: 'Kerentanan pengambilalihan akun', evidenceFile: null, rootCause: [{ id: 2, method: 'System', w1: 'Modul MFA belum aktif', w2: 'Lisensi belum di-renew', w3: '' }] },
    { id: 'WP-C-003', workingPaperId: 'ST-003/SKAI/2026', condition: 'Suhu gudang bahan kimia belum terpantau 24/7', criteria: 'SOP K3 Storage', impact: 'Kerusakan bahan kimia', evidenceFile: null, rootCause: [{ id: 3, method: 'Process', w1: 'Sensor manual', w2: 'IoT belum terpasang', w3: '' }] },
    { id: 'WP-C-004', workingPaperId: 'ST-004/SKAI/2026', condition: 'Kertas kerja survei HPS belum dilampirkan', criteria: 'SOP SCM Bab 2', impact: 'Kurang akuratnya HPS', evidenceFile: null, rootCause: [{ id: 4, method: 'Process', w1: 'Formulir survei tidak wajib diunggah', w2: '', w3: '' }] },
    { id: 'WP-C-005', workingPaperId: 'ST-005/SKAI/2026', condition: 'Inspeksi hidran belum 100%', criteria: 'Standar K3LH Pembangkit', impact: 'Risiko kegagalan pemadaman', evidenceFile: null, rootCause: [{ id: 5, method: 'Process', w1: 'Jadwal pemeliharaan bentrok', w2: '', w3: '' }] },
    { id: 'WP-C-020', workingPaperId: '020/ST/01/KSIAD/2023', condition: 'Pelaksanaan Overhaul ME+ PLTU TBK #1 Terjadi PE 6 Hari', criteria: 'SOP Maintenance UPDK', impact: 'Penurunan ketersediaan EAF', evidenceFile: null, rootCause: [{ id: 6, method: 'Process', w1: 'Masa transisi holding sub-holding', w2: 'Keterbatasan lab batubara', w3: 'Fitur Maximo WPC terhenti' }] }
  ]

  const mockF05: WorkingPaperPlan[] = [
    { id: 'WP-P-001', workingPaperId: 'ST-001/SKAI/2026', recommendation: 'Lakukan rekonsiliasi harian dan alert otomatis', response: 'Disetujui', actionDescription: 'Pembuatan modul otomatisasi rekonsiliasi', pic: 'Budi - Finance', periodAction: '2026-05-15' },
    { id: 'WP-P-002', workingPaperId: 'ST-002/SKAI/2026', recommendation: 'Wajibkan MFA untuk seluruh akses Superadmin ERP', response: 'Disetujui', actionDescription: 'Pengadaan dan konfigurasi MFA', pic: 'Dimas - IT', periodAction: '2026-06-01' },
    { id: 'WP-P-003', workingPaperId: 'ST-003/SKAI/2026', recommendation: 'Pasang IoT sensor pemantau suhu 24/7', response: 'Disetujui', actionDescription: 'Instalasi sensor IoT gudang', pic: 'Siti - HR', periodAction: '2026-08-30' },
    { id: 'WP-P-004', workingPaperId: 'ST-004/SKAI/2026', recommendation: 'Lampirkan kertas kerja survei harga pada HPS', response: 'Disetujui', actionDescription: 'Update template HPS di ERP SCM', pic: 'Budi - Finance', periodAction: '2026-09-15' },
    { id: 'WP-P-005', workingPaperId: 'ST-005/SKAI/2026', recommendation: 'Lakukan pengujian dan sertifikasi hidran berkala', response: 'Disetujui', actionDescription: 'Jadwal rutin inspeksi K3', pic: 'Dewi - Maintenance', periodAction: '2026-10-15' },
    { id: 'WP-P-020', workingPaperId: '020/ST/01/KSIAD/2023', recommendation: 'Penyusunan DMR & Sertifikasi Pemeliharaan PLTU', response: 'Disetujui', actionDescription: 'Sertifikasi O&M dan pembaruan Maximo WPC', pic: 'Tomy - UPDK', periodAction: '2023-11-30' }
  ]

  const filteredDataF01 = computed(() => {
    const list = dataF01.value.length > 0 ? dataF01.value : mockF01
    if (!fieldworkStore.selectedAssignmentLetter) return list
    return list.filter(wp => wp.assignmentLetterId === fieldworkStore.selectedAssignmentLetter)
  })

  const filteredDataF02 = computed(() => {
    const list = dataF02.value.length > 0 ? dataF02.value : mockF02
    if (!fieldworkStore.selectedAssignmentLetter) return list
    return list.filter(wp => (wp.workingPaperId || (wp as any).assignmentLetterId) === fieldworkStore.selectedAssignmentLetter)
  })

  const filteredDataF03 = computed(() => {
    const list = dataF03.value.length > 0 ? dataF03.value : mockF03
    if (!fieldworkStore.selectedAssignmentLetter) return list
    return list.filter(wp => (wp.workingPaperId || (wp as any).assignmentLetterId) === fieldworkStore.selectedAssignmentLetter)
  })

  const filteredDataF04 = computed(() => {
    const list = dataF04.value.length > 0 ? dataF04.value : mockF04
    if (!fieldworkStore.selectedAssignmentLetter) return list
    return list.filter(wp => (wp.workingPaperId || (wp as any).assignmentLetterId) === fieldworkStore.selectedAssignmentLetter)
  })

  const filteredDataF05 = computed(() => {
    const list = dataF05.value.length > 0 ? dataF05.value : mockF05
    if (!fieldworkStore.selectedAssignmentLetter) return list
    return list.filter(wp => (wp.workingPaperId || (wp as any).assignmentLetterId) === fieldworkStore.selectedAssignmentLetter)
  })

  const isEditingF01 = ref(false)
  const isEditingF02 = ref(false)
  const isEditingF03 = ref(false)
  const isEditingF04 = ref(false)
  const isEditingF05 = ref(false)
  const editingIdF01 = ref<string | null>(null)
  const editingIdF02 = ref<string | null>(null)
  const editingIdF03 = ref<string | null>(null)
  const editingIdF04 = ref<string | null>(null)
  const editingIdF05 = ref<string | null>(null)
  const showModalF01 = ref(false)
  const showModalF02 = ref(false)
  const showModalF03 = ref(false)
  const showModalF04 = ref(false)
  const showModalF05 = ref(false)
  const closeModalF01 = () => showModalF01.value = false
  const closeModalF02 = () => showModalF02.value = false
  const closeModalF03 = () => showModalF03.value = false
  const closeModalF04 = () => showModalF04.value = false
  const closeModalF05 = () => showModalF05.value = false

  const addF01 = async (headerForm: WorkingPaperHeaderForm) => {
    const auditPeriod = `${headerForm.periodStart} s/d ${headerForm.periodEnd}`
    const newHeader = {
      assignmentLetterId: headerForm.assignmentLetterId,
      auditPurpose: headerForm.auditPurpose,
      businessProcess: headerForm.businessProcess,
      period: auditPeriod,
      location: headerForm.location,
      teamMembers: headerForm.teamMembers
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/headers`, {
      method: 'POST',
      body: newHeader
    })
    await fetchAllData()
  }

  const updateF01 = async (id: string, updatedHeaderData: WorkingPaperHeaderForm) => {
    const auditPeriod = `${updatedHeaderData.periodStart} s/d ${updatedHeaderData.periodEnd}`
    const payload = {
      assignmentLetterId: updatedHeaderData.assignmentLetterId,
      auditPurpose: updatedHeaderData.auditPurpose,
      businessProcess: updatedHeaderData.businessProcess,
      period: auditPeriod,
      location: updatedHeaderData.location,
      teamMembers: updatedHeaderData.teamMembers
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/headers/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchAllData()
  }

  const deleteF01 = async (id: string) => {
    if (confirm('Are you sure you want to delete permanently?')) {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/headers/${id}`, {
        method: 'DELETE'
      })
      await fetchAllData()
    }
  }

  const openModalF01 = () => {
    isEditingF01.value = false
    editingIdF01.value = null

    // Reset Form
    Object.assign(headerForm, {
      assignmentLetterId: fieldworkStore.selectedAssignmentLetter || '',
      auditPurpose: '',
      businessProcess: '',
      periodStart: '',
      periodEnd: '',
      location: '',
      teamMembers: [
        { id: Date.now(), name: '', role: '' }
      ]
    })
    showModalF01.value = true
  }

  const handleSubmitF01 = async () => {
    try {
      if (isEditingF01.value && editingIdF01.value) {
        await updateF01(editingIdF01.value, { ...headerForm })
        alert("Header Data Updated Successfully!")
      } else {
        await addF01({ ...headerForm })
        alert("Header Data Successfully Saved!")
      }
      closeModalF01()
    } catch (error: any) {
      alert("Failed to save data: " + error.message)
    }
  }

  const handleEditF01 = (header: any) => {
    isEditingF01.value = true
    editingIdF01.value = header.id

    let start = ""
    let end = ""
    if (header.period && header.period.includes(" s/d ")) {
      const parts = header.period.split(" s/d ")
      start = parts[0] || ""
      end = parts[1] || ""
    }

    headerForm.assignmentLetterId = header.assignmentLetterId
    headerForm.auditPurpose = header.auditPurpose
    headerForm.businessProcess = header.businessProcess
    headerForm.periodStart = start
    headerForm.periodEnd = end
    headerForm.location = header.location
    headerForm.teamMembers = header.teamMembers ? header.teamMembers.map((m: any) => ({ ...m })) : []

    showModalF01.value = true
  }

  const handleDeleteF01 = async (id: string | undefined) => {
    if (!id) return
    try {
      await deleteF01(id)
    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF02 = async (riskForm: WorkingPaperRiskForm) => {
    const newRisk = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      risk: riskForm.risk,
      taxonomy: riskForm.taxonomy || 'Operational',
      riskLevel: riskForm.riskLevel || 'High',
      controlDescription: riskForm.controlDescription
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/risks`, {
      method: 'POST',
      body: newRisk
    })
    await fetchAllData()
  }

  const updateF02 = async (id: string, updatedRiskData: WorkingPaperRiskForm) => {
    const payload = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      risk: updatedRiskData.risk,
      taxonomy: updatedRiskData.taxonomy || 'Operational',
      riskLevel: updatedRiskData.riskLevel || 'High',
      controlDescription: updatedRiskData.controlDescription
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/risks/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchAllData()
  }

  const deleteF02 = async (id: string) => {
    if (confirm('Are you sure you want to delete permanently?')) {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/risks/${id}`, {
        method: 'DELETE'
      })
      await fetchAllData()
    }
  }

  const openModalF02 = () => {
    isEditingF02.value = false
    editingIdF02.value = null

    // Reset Form
    Object.assign(riskForm, {
      risk: '',
      taxonomy: 'Operational',
      riskLevel: 'High',
      controlDescription: ''
    })
    showModalF02.value = true
  }

  const handleSubmitF02 = async () => {
    try {
      if (isEditingF02.value && editingIdF02.value) {
        await updateF02(editingIdF02.value, { ...riskForm })
        alert("Risk Data Updated Successfully!")
      } else {
        await addF02({ ...riskForm })
        alert("Risk Data Successfully Saved!")
      }
      closeModalF02()
    } catch (error: any) {
      alert("Failed to save data: " + error.message)
    }
  }

  const handleEditF02 = (risk: any) => {
    isEditingF02.value = true
    editingIdF02.value = risk.id

    riskForm.risk = risk.risk
    riskForm.taxonomy = risk.taxonomy
    riskForm.riskLevel = risk.riskLevel
    riskForm.controlDescription = risk.controlDescription

    showModalF02.value = true
  }

  const handleDeleteF02 = async (id: string | undefined) => {
    if (!id) return
    try {
      await deleteF02(id)
    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF03 = async (sampleForm: WorkingPaperSampleForm) => {
    const newSample = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      population: sampleForm.population,
      sampleSize: sampleForm.sampleSize,
      samples: sampleForm.samples,
      conclusion: sampleForm.conclusion
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/samples`, {
      method: 'POST',
      body: newSample
    })
    await fetchAllData()
  }

  const updateF03 = async (id: string, updatedData: WorkingPaperSampleForm) => {
    const payload = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      population: updatedData.population,
      sampleSize: updatedData.sampleSize,
      samples: updatedData.samples,
      conclusion: updatedData.conclusion
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/samples/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchAllData()
  }

  const deleteF03 = async (id: string) => {
    if (confirm('Are you sure you want to delete permanently?')) {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/samples/${id}`, {
        method: 'DELETE'
      })
      await fetchAllData()
    }
  }

  const openModalF03 = () => {
    isEditingF03.value = false
    editingIdF03.value = null

    // Reset Form
    Object.assign(sampleForm, {
      population: 0,
      sampleSize: 0,
      samples: [],
      conclusion: ''
    })
    showModalF03.value = true
  }

  const handleSubmitF03 = async () => {
    try {
      if (isEditingF03.value && editingIdF03.value) {
        await updateF03(editingIdF03.value, { ...sampleForm })
        alert("Sample Data Successfully Updated!")
      } else {
        await addF03({ ...sampleForm })
        alert("Sample Data Successfully Saved!")
      }
      closeModalF03()
    } catch (error: any) {
      alert("Failed to save data: " + error.message)
    }
  }

  const handleEditF03 = (sample: any) => {
    isEditingF03.value = true
    editingIdF03.value = sample.id

    sampleForm.population = sample.population
    sampleForm.sampleSize = sample.sampleSize
    sampleForm.samples = sample.samples ? sample.samples.map((s: any) => ({ ...s })) : []
    sampleForm.conclusion = sample.conclusion

    showModalF03.value = true
  }

  const handleDeleteF03 = async (id: string | undefined) => {
    if (!id) return
    try {
      await deleteF03(id)
    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF04 = async (causeForm: WorkingPaperCauseForm) => {
    const newCause = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      condition: causeForm.condition,
      criteria: causeForm.criteria,
      impact: causeForm.impact,
      evidenceFile: causeForm.evidenceFile ? causeForm.evidenceFile.name : null,
      rootCause: causeForm.rootCause
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/causes`, {
      method: 'POST',
      body: newCause
    })
    await fetchAllData()
  }

  const updateF04 = async (id: string, updatedData: WorkingPaperCauseForm) => {
    const payload = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      condition: updatedData.condition,
      criteria: updatedData.criteria,
      impact: updatedData.impact,
      evidenceFile: updatedData.evidenceFile ? updatedData.evidenceFile.name : null,
      rootCause: updatedData.rootCause
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/causes/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchAllData()
  }

  const deleteF04 = async (id: string) => {
    if (confirm('Are you sure you want to delete permanently?')) {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/causes/${id}`, {
        method: 'DELETE'
      })
      await fetchAllData()
    }
  }

  const openModalF04 = () => {
    isEditingF04.value = false
    editingIdF04.value = null

    // Reset Form
    Object.assign(causeForm, {
      condition: '',
      criteria: '',
      impact: '',
      evidenceFile: null,
      rootCause: []
    })
    showModalF04.value = true
  }

  const handleSubmitF04 = async () => {
    try {
      if (isEditingF04.value && editingIdF04.value) {
        await updateF04(editingIdF04.value, { ...causeForm })
        alert("Root Cause Data Successfully Updated!")
      } else {
        await addF04({ ...causeForm })
        alert("Root Cause Data Successfully Saved!")
      }
      closeModalF04()
    } catch (error: any) {
      alert("Failed to save data: " + error.message)
    }
  }

  const handleEditF04 = (cause: any) => {
    isEditingF04.value = true
    editingIdF04.value = cause.id

    causeForm.condition = cause.condition
    causeForm.criteria = cause.criteria
    causeForm.impact = cause.impact
    causeForm.evidenceFile = cause.evidenceFile ? new File([], cause.evidenceFile) : null
    causeForm.rootCause = cause.rootCause ? cause.rootCause.map((rca: any) => ({ ...rca })) : []

    showModalF04.value = true
  }

  const handleDeleteF04 = async (id: string | undefined) => {
    if (!id) return
    try {
      await deleteF04(id)
    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF05 = async (planForm: WorkingPaperPlanForm) => {
    const newPlan = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      recommendation: planForm.recommendation,
      response: planForm.response,
      actionDescription: planForm.actionDescription,
      pic: planForm.pic,
      periodAction: planForm.periodAction
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/plans`, {
      method: 'POST',
      body: newPlan
    })
    await fetchAllData()
  }

  const updateF05 = async (id: string, updatedData: WorkingPaperPlanForm) => {
    const payload = {
      workingPaperId: fieldworkStore.selectedAssignmentLetter,
      recommendation: updatedData.recommendation,
      response: updatedData.response,
      actionDescription: updatedData.actionDescription,
      pic: updatedData.pic,
      periodAction: updatedData.periodAction
    }
    const baseUrl = getAuditServiceBaseUrl()
    await $fetch(`${baseUrl}/working-papers/plans/${id}`, {
      method: 'PUT',
      body: payload
    })
    await fetchAllData()
  }

  const deleteF05 = async (id: string) => {
    if (confirm('Are you sure you want to delete permanently?')) {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/working-papers/plans/${id}`, {
        method: 'DELETE'
      })
      await fetchAllData()
    }
  }

  const openModalF05 = () => {
    isEditingF05.value = false
    editingIdF05.value = null

    // Reset Form
    Object.assign(planForm, {
      recommendation: '',
      response: '',
      actionDescription: '',
      pic: '',
      periodAction: ''
    })
    showModalF05.value = true
  }

  const handleSubmitF05 = async () => {
    try {
      if (isEditingF05.value && editingIdF05.value) {
        await updateF05(editingIdF05.value, { ...planForm })
        alert("Action Plan Data Successfully Updated!")
      } else {
        await addF05({ ...planForm })
        alert("Action Plan Data Successfully Saved!")
      }
      closeModalF05()
    } catch (error: any) {
      alert("Failed to save data: " + error.message)
    }
  }

  const handleEditF05 = (plan: any) => {
    isEditingF05.value = true
    editingIdF05.value = plan.id

    planForm.recommendation = plan.recommendation
    planForm.response = plan.response || plan.tanggapanresponse || ""
    planForm.actionDescription = plan.actionDescription
    planForm.pic = plan.pic
    planForm.periodAction = plan.periodAction

    showModalF05.value = true
  }

  const handleDeleteF05 = async (id: string | undefined) => {
    if (!id) return
    try {
      await deleteF05(id)
    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  // --- ACTIONS: Uji Sampel (F-03) ---
  const addSample = () => {
    sampleForm.samples.push({
      id: Date.now(),
      document: '',
      l1: undefined,
      l2: undefined,
      l3: undefined
    })
  }

  const removeSample = (index: number) => {
    sampleForm.samples.splice(index, 1)
  }

  // --- GETTERS: Cek Efektivitas Sampel ---
  // Return true = Efektif, Return false = Tidak Efektif
  const checkSampleStatus = (sampel: SampleItem): boolean => {
    if (sampel.l1 === 'Fail' || sampel.l2 === 'Fail' || sampel.l3 === 'Fail') {
      return false
    }
    return true
  }

  // --- ACTIONS: Root Cause (F-04) ---
  const addRootCause = () => {
    causeForm.rootCause.push({
      id: Date.now(),
      method: 'People',
      w1: '',
      w2: '',
      w3: ''
    })
  }

  const removeRootCause = (index: number) => {
    causeForm.rootCause.splice(index, 1)
  }

  const addTeamMember = () => {
    headerForm.teamMembers.push({
      id: Date.now(),
      name: '',
      role: ''
    })
  }

  const removeTeamMember = (index: number) => {
    if (headerForm.teamMembers.length > 1) {
      headerForm.teamMembers.splice(index, 1)
    }
  }

  const options = {
    get assignmentLetter() {
      return fieldworkStore.publishedAssignmentLetters
    },
    businessProcess: ['Procurement', 'Finance', 'HR', 'IT Operations'],
    location: ['Head Office', 'Jakarta Branch', 'Bandung Branch'],
    risk: [
      'FIN-001: Fluktuasi Nilai Tukar Mata Uang',
      'OPS-001: Gangguan Sistem IT',
      'COM-001: Ketidakpatuhan Regulasi GDPR',
      'STR-001: Perubahan Strategi Kompetitor',
      'OPS-002: Kegagalan Rantai Pasokan',
      'FIN-002: Risiko Kredit Pelanggan',
      'SEC-001: Serangan Siber dan Ransomware',
      'REP-001: Penurunan Reputasi Brand',
      'ENV-001: Dampak Perubahan Iklim',
      'HR-001: Retensi Talenta Kunci',
      'OPS-003: Kegagalan Peralatan Kritis',
      'FIN-003: Kenaikan Suku Bunga',
      'COM-002: Perubahan Regulasi Pajak',
      'STR-002: Disrupsi Teknologi',
      'SEC-002: Kebocoran Data Pelanggan',
      'REP-002: Krisis Reputasi Media Sosial',
      'ENV-002: Bencana Alam Lokal',
      'HR-002: Konflik Industrial',
      'OPS-004: Keterlambatan Proyek Konstruksi',
      'FIN-004: Penipuan Internal',
      'COM-003: Perubahan Regulasi Lingkungan',
      'STR-003: Kegagalan Merger & Akuisisi',
      'SEC-003: Pencurian Kekayaan Intelektual',
      'REP-003: Ulasan Negatif Pelanggan',
      'ENV-003: Kenaikan Biaya Energi',
      'HR-003: Kesenjangan Keterampilan',
      'OPS-005: Kegagalan Sistem Logistik',
      'FIN-005: Penurunan Pendapatan',
      'COM-004: Perubahan Regulasi Pasar',
      'STR-004: Kegagalan Inovasi Produk',
      'SEC-004: Pelanggaran Privasi Data',
      'REP-004: Tuduhan Etika Bisnis',
      'ENV-004: Kerusakan Reputasi Akibat Polusi',
      'HR-004: Korupsi dan Penipuan Internal',
      'OPS-006: Kegagalan Sistem Keamanan Fisik',
      'FIN-006: Penipuan Pelanggan',
      'COM-005: Perubahan Regulasi Pasar',
      'STR-005: Kegagalan Inovasi Produk',
      'SEC-005: Pelanggaran Privasi Data',
      'REP-005: Tuduhan Etika Bisnis',
      'ENV-005: Kerusakan Reputasi Akibat Polusi',
      'HR-005: Korupsi dan Penipuan Internal',
      'OPS-007: Kegagalan Sistem Keamanan Fisik',
      'FIN-007: Penipuan Pelanggan',
    ],
    pic: ['Dimas - IT', 'Budi - Finance', 'Siti - HR'],
    testResult: [...TEST_RESULT_OPTIONS],
    rootCauseMethod: [...ROOT_CAUSE_METHOD_OPTIONS],
  }

  return {
    headerForm, riskForm, sampleForm, causeForm, planForm,
    options, dateErrorMessage, isDateError, tabs, workingItems,
    columnsF01, columnsF02, columnsF03, columnsF04, columnsF05,
    showModalF01, showModalF02, showModalF03, showModalF04, showModalF05,
    isEditingF01, isEditingF02, isEditingF03, isEditingF04, isEditingF05,
    dataF01, dataF02, dataF03, dataF04, dataF05,
    filteredDataF01, filteredDataF02, filteredDataF03, filteredDataF04, filteredDataF05,
    updateF01, updateF02, updateF03, updateF04, updateF05,
    deleteF01, deleteF02, deleteF03, deleteF04, deleteF05,
    openModalF01, openModalF02, openModalF03, openModalF04, openModalF05,
    closeModalF01, closeModalF02, closeModalF03, closeModalF04, closeModalF05,
    handleSubmitF01, handleSubmitF02, handleSubmitF03, handleSubmitF04, handleSubmitF05,
    handleEditF01, handleEditF02, handleEditF03, handleEditF04, handleEditF05,
    handleDeleteF01, handleDeleteF02, handleDeleteF03, handleDeleteF04, handleDeleteF05,
    addSample, removeSample, addRootCause, removeRootCause, triggerUpload, onFileChange,
    checkSampleStatus, addTeamMember, removeTeamMember, getAvailableMembers, removeFile,
    loading, errorMsg, fetchAllData
  }
})