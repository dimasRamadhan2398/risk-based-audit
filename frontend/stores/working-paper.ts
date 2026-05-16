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

  const filteredDataF01 = computed(() => {
    if (!fieldworkStore.selectedAssignmentLetter) return dataF01.value
    return dataF01.value.filter(wp => wp.assignmentLetterId === fieldworkStore.selectedAssignmentLetter)
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

  const addF01 = (headerForm: WorkingPaperHeaderForm) => {
    const auditPeriod = `${headerForm.periodStart} s/d ${headerForm.periodEnd}`

    const newHeader: WorkingPaperHeader = {
      id: Date.now().toString(),
      assignmentLetterId: headerForm.assignmentLetterId,
      auditPurpose: headerForm.auditPurpose,
      businessProcess: headerForm.businessProcess,
      period: auditPeriod,
      location: headerForm.location,
      teamMembers: headerForm.teamMembers
    }
    dataF01.value.unshift(newHeader)

  }

  const updateF01 = (id: string, updatedHeaderData: WorkingPaperHeaderForm) => {
    const index = dataF01.value.findIndex(p => p.id === id)
    const isDuplicate = dataF01.value.some(a => a.assignmentLetterId === updatedHeaderData.assignmentLetterId && a.id !== id)
    const auditPeriod = `${updatedHeaderData.periodStart} s/d ${updatedHeaderData.periodEnd}`
    const existingHeader = dataF01.value[index]!

    if (isDuplicate) throw new Error('The assignment letter has been used for other data!')
    if (index === -1) return

    dataF01.value[index] = {
      ...updatedHeaderData,
      id: existingHeader.id,
      period: auditPeriod,
    }

  }

  const deleteF01 = (id: string) => {
    const header = dataF01.value.find(a => a.id === id)
    if (!header) return

    if (isEditingF01.value && editingIdF01.value) {
      if (confirm('Are you sure you want to delete permanently?')) {
        dataF01.value = dataF01.value.filter(a => a.id !== id)

      } else {
        if (confirm('Are you sure you want to delete permanently?')) {
          dataF01.value = dataF01.value.filter(a => a.id !== id)
        }
      }

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

  const handleSubmitF01 = () => {

    if (isEditingF01.value && editingIdF01.value) {
      // Mode EDIT
      updateF01(editingIdF01.value, { ...headerForm })
      alert("Header Data Updated Successfully!")
    } else {
      // Mode ADD
      addF01({ ...headerForm })
      alert("Header Data Successfully Saved!")
    }

    closeModalF01()
  }

  const handleEditF01 = (header: any) => {
    isEditingF01.value = true
    editingIdF01.value = header.id

    headerForm.auditPurpose = header.auditPurpose,
      headerForm.businessProcess = header.businessProcess,
      headerForm.periodStart = header.periodStart,
      headerForm.periodEnd = header.periodEnd,
      headerForm.location = header.location,
      headerForm.teamMembers = header.teamMembers

    showModalF01.value = true
  }

  const handleDeleteF01 = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deleteF01(id)

    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF02 = (riskForm: WorkingPaperRiskForm) => {
    const newRisk: WorkingPaperRisk = {
      id: Date.now().toString(),
      risk: riskForm.risk,
      taxonomy: riskForm.taxonomy || 'Operational',
      riskLevel: riskForm.riskLevel || 'High',
      controlDescription: riskForm.controlDescription
    }
    dataF02.value.unshift(newRisk)
  }

  const updateF02 = (id: string, updatedRiskData: WorkingPaperRiskForm) => {
    const index = dataF02.value.findIndex(p => p.id === id)
    const existingRisk = dataF02.value[index]!
    if (index === -1) return

    dataF02.value[index] = {
      ...updatedRiskData,
      id: existingRisk.id
    }

  }

  const deleteF02 = (id: string) => {
    const risk = dataF02.value.find(a => a.id === id)
    if (!risk) return

    if (confirm('Are you sure you want to delete permanently??')) {
      dataF02.value = dataF02.value.filter(a => a.id !== id)
    }
  }

  const openModalF02 = () => {
    isEditingF02.value = false
    editingIdF02.value = null

    // Reset Form
    Object.assign(riskForm, {
      risk: '',
      taxonomy: '',
      riskLevel: '',
      controlDescription: ''
    })
    showModalF02.value = true
  }

  const handleSubmitF02 = () => {

    if (isEditingF02.value && editingIdF02.value) {
      // Mode EDIT
      updateF02(editingIdF02.value, { ...riskForm })
      alert("Risk Data Updated Successfully!")
    } else {
      // Mode ADD
      addF02({ ...riskForm })
      alert("Risk Data Successfully Saved!")
    }

    closeModalF02()
  }

  const handleEditF02 = (risk: any) => {
    isEditingF02.value = true
    editingIdF02.value = risk.id

    // Isi form dengan data yang dipilih
    riskForm.risk = risk.risk,
      riskForm.taxonomy = risk.taxonomy,
      riskForm.riskLevel = risk.riskLevel,
      riskForm.controlDescription = risk.controlDescription

    showModalF02.value = true
  }

  const handleDeleteF02 = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deleteF02(id)

    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF03 = (sampleForm: WorkingPaperSampleForm) => {
    const newSample: WorkingPaperSample = {
      id: Date.now().toString(),
      population: sampleForm.population,
      sampleSize: sampleForm.sampleSize,
      samples: [...sampleForm.samples],
      conclusion: sampleForm.conclusion
    }
    dataF03.value.unshift(newSample)

  }

  const updateF03 = (id: string, updatedData: WorkingPaperSampleForm) => {
    const index = dataF03.value.findIndex(p => p.id === id)
    const existingSample = dataF03.value[index]!
    if (index === -1) return

    dataF03.value[index] = {
      ...updatedData,
      id: existingSample.id,
      samples: [...updatedData.samples]
    }

  }

  const deleteF03 = (id: string) => {
    const sample = dataF03.value.find(a => a.id === id)
    if (!sample) return

    if (confirm('Are you sure you want to delete permanently?')) {
      dataF03.value = dataF03.value.filter(a => a.id !== id)

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

  const handleSubmitF03 = () => {

    if (isEditingF03.value && editingIdF03.value) {
      // Mode EDIT
      updateF03(editingIdF03.value, { ...sampleForm })
      alert("Sample Data Successfully Updated!")
    } else {
      // Mode ADD
      addF03({ ...sampleForm })
      alert("Sample Data Successfully Saved!")
    }

    closeModalF03()
  }

  const handleEditF03 = (sample: any) => {
    isEditingF03.value = true
    editingIdF03.value = sample.id

    sampleForm.population = sample.population,
      sampleForm.sampleSize = sample.sampleSize,
      sampleForm.samples = sample.samples.map((act: any) => ({ ...act })),
      sampleForm.conclusion = sample.conclusion

    showModalF03.value = true
  }

  const handleDeleteF03 = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deleteF03(id)

    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF04 = (causeForm: WorkingPaperCauseForm) => {
    const newCause: WorkingPaperCause = {
      id: Date.now().toString(),
      condition: causeForm.condition,
      criteria: causeForm.criteria,
      impact: causeForm.impact,
      evidenceFile: causeForm.evidenceFile,
      rootCause: causeForm.rootCause.map(rca => ({ ...rca }))
    }
    dataF04.value.unshift(newCause)
  }

  const updateF04 = (id: string, updatedCauseData: WorkingPaperCauseForm) => {
    const index = dataF04.value.findIndex(p => p.id === id)
    const existingCause = dataF04.value[index]!
    if (index === -1) return

    dataF04.value[index] = {
      ...updatedCauseData,
      id: existingCause.id,
      rootCause: [...updatedCauseData.rootCause]
    }

  }

  const deleteF04 = (id: string) => {
    const cause = dataF04.value.find(a => a.id === id)
    if (!cause) return

    if (confirm('Are you sure you want to delete permanently??')) {
      dataF04.value = dataF04.value.filter(a => a.id !== id)

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

  const handleSubmitF04 = () => {

    if (isEditingF04.value && editingIdF04.value) {
      // Mode EDIT
      updateF04(editingIdF04.value, { ...causeForm })
      alert("Root Cause Data Successfully Updated!")
    } else {
      // Mode ADD
      addF04({ ...causeForm })
      alert("Root Cause Data Successfully Saved!")
    }

    closeModalF04()
  }

  const handleEditF04 = (cause: any) => {
    isEditingF04.value = true
    editingIdF04.value = cause.id

    causeForm.condition = cause.condition,
      causeForm.criteria = cause.criteria,
      causeForm.impact = cause.impact,
      causeForm.evidenceFile = cause.evidenceFile,
      causeForm.rootCause = cause.rootCause.map((rca: any) => ({ ...rca }))


    showModalF04.value = true
  }

  const handleDeleteF04 = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deleteF04(id)

    } catch (error) {
      alert('Failed to delete data: ' + error)
    }
  }

  const addF05 = (planForm: WorkingPaperPlanForm) => {
    const newPlan: WorkingPaperPlan = {
      id: Date.now().toString(),
      recommendation: planForm.recommendation,
      response: planForm.response,
      actionDescription: planForm.actionDescription,
      pic: planForm.pic,
      periodAction: planForm.periodAction
    }
    dataF05.value.unshift(newPlan)

  }

  const updateF05 = (id: string, updatedPlanData: WorkingPaperPlanForm) => {
    const index = dataF05.value.findIndex(p => p.id === id)
    const existingPlan = dataF05.value[index]!
    if (index === -1) return

    dataF05.value[index] = {
      ...updatedPlanData,
      id: existingPlan.id
    }

  }

  const deleteF05 = (id: string) => {
    const plan = dataF05.value.find(a => a.id === id)
    if (!plan) return

    if (confirm('Are you sure you want to delete permanently?')) {
      dataF05.value = dataF05.value.filter(a => a.id !== id)

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

  const handleSubmitF05 = () => {

    if (isEditingF05.value && editingIdF05.value) {
      // Mode EDIT
      updateF05(editingIdF05.value, { ...planForm })
      alert("Action Plan Data Successfully Updated!")
    } else {
      // Mode ADD
      addF05({ ...planForm })
      alert("Action Plan Data Successfully Saved!")
    }

    closeModalF05()
  }

  const handleEditF05 = (plan: any) => {
    isEditingF05.value = true
    editingIdF05.value = plan.id

    planFormrecommendation: plan.recommendation,
      planForm.response = plan.tanggapanresponse,
      planForm.actionDescription = plan.actionDescription,
      planForm.pic = plan.pic,
      planForm.periodAction = plan.periodAction


    showModalF05.value = true
  }

  const handleDeleteF05 = (id: string | undefined) => {
    if (!id) return
    try {
      // 2. Panggil fungsi hapus di store
      deleteF05(id)

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
    filteredDataF01,
    updateF01, updateF02, updateF03, updateF04, updateF05,
    deleteF01, deleteF02, deleteF03, deleteF04, deleteF05,
    openModalF01, openModalF02, openModalF03, openModalF04, openModalF05,
    closeModalF01, closeModalF02, closeModalF03, closeModalF04, closeModalF05,
    handleSubmitF01, handleSubmitF02, handleSubmitF03, handleSubmitF04, handleSubmitF05,
    handleEditF01, handleEditF02, handleEditF03, handleEditF04, handleEditF05,
    handleDeleteF01, handleDeleteF02, handleDeleteF03, handleDeleteF04, handleDeleteF05,
    addSample, removeSample, addRootCause, removeRootCause, triggerUpload, onFileChange,
    checkSampleStatus, addTeamMember, removeTeamMember, getAvailableMembers, removeFile
  }
})