import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import type { WorkingPaperForm, SampleItem } from '~/types/audit'
import { RCA_METHOD_OPTIONS, TEST_RESULT_OPTIONS } from '~/types/audit'

export const useWorkingPaperStore = defineStore('working-paper', () => {

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
        alert('Ukuran file terlalu besar! Maksimal 10MB.')
        return
      }

      // Simpan ke Store
      form.buktiFile = file
      console.log('File terpilih:', file.name)
    }
  }

  // Hapus file
  const removeFile = () => {
    form.buktiFile = null
    if (fileInput.value) {
      fileInput.value.value = '' // Reset input agar bisa upload file yang sama lagi
    }
  }

  // Fungsi untuk mendapatkan daftar anggota yang tersedia untuk baris tertentu
  const getAvailableMembers = (currentIndex: number) => {
    // 1. Ambil semua nama yang sudah dipilih di baris-baris LAIN
    const selectedNames = form.teamMembers
      .filter((_, index) => index !== currentIndex) // Kecualikan baris yang sedang aktif
      .map(member => member.name)
      .filter(name => !!name) // Hanya ambil yang sudah ada isinya

    // 2. Filter master list PIC agar tidak menyertakan nama yang sudah dipilih di baris lain
    return options.pic.filter(pic => !selectedNames.includes(pic))
  }

  const isDateError = computed(() => {
    // Jika salah satu tanggal belum diisi, jangan anggap error dulu
    if (!form.periodeStart || !form.periodeEnd) return false
    
    const start = new Date(form.periodeStart)
    const end = new Date(form.periodeEnd)
    
    // Return true jika tanggal akhir LEBIH KECIL dari tanggal mulai
    return end < start
  })

  // Opsional: Pesan error dinamis
  const dateErrorMessage = computed(() => {
    return isDateError.value ? "Tanggal akhir tidak boleh sebelum tanggal mulai" : false
  })

  // Tabs Configuration
  const tabs = [
    { label: 'Header', slot: 'f01', icon: 'i-heroicons-document-text' },
    { label: 'Profil Resiko', slot: 'f02', icon: 'i-heroicons-shield-exclamation' },
    { label: 'Uji Sampel', slot: 'f03', icon: 'i-heroicons-table-cells' },
    { label: 'AOI & RCA', slot: 'f04', icon: 'i-heroicons-magnifying-glass-circle' },
    { label: 'Action Plan', slot: 'f05', icon: 'i-heroicons-check-badge' }
  ]

  const columnsF01 = [
    { accessorKey: 'assignmentLetter', header: 'Assignment Letter' },
    { accessorKey: 'businessProcess', header: 'Business Process' },
    { accessorKey: 'period', header: 'Period' },
    { accessorKey: 'location', header: 'Location' },
    { accessorKey: 'teamMembers', header: 'Team' }
  ]

  const columnsF02 = [
    { accessorKey: 'risk', header: 'Risk' },
    { accessorKey: 'taxonomy', header: 'Taxonomy' },
    { accessorKey: 'riskLevel', header: 'Risk Level' },
    { accessorKey: 'controlDescription', header: 'Control Description' }
  ]

  const columnsF03 = [
    { accessorKey: 'population', header: 'Population' },
    { accessorKey: 'sampleSize', header: 'Sample Size' },
    { accessorKey: 'samples', header: 'Sample List' },
    { accessorKey: 'conclusion', header: 'Conclusion' }
  ]

  const columnsF04 = [
    { accessorKey: 'condition', header: 'Condition' },
    { accessorKey: 'criteria', header: 'Criteria' },
    { accessorKey: 'impact', header: 'Impact' },
    { accessorKey: 'rootCause', header: 'Root Cause' },
    { accessorKey: 'evidenceFile', header: 'Evidence Document' },
  ]

  const columnsF05 = [
    { accessorKey: 'recommendation', header: 'Recommendation' },
    { accessorKey: 'response', header: 'Auditee Response' },
    { accessorKey: 'actionDescription', header: 'Description' },
    { accessorKey: 'pic', header: 'PIC' },
    { accessorKey: 'targetDate', header: 'Target Date' }
  ]
  // --- STATE ---
  const form = reactive<WorkingPaperForm>({
    suratTugas: '',
    tujuanAudit: '',
    prosesBisnis: '',
    periodeStart: '',
    periodeEnd: '',
    lokasi: '',
    teamMembers: [
      { id: Date.now(), name: '', role: '' } // Inisialisasi 1 baris kosong
    ],
    
    resiko: '',
    taksonomi: '',
    tingkatResiko: '',
    deskripsiKontrol: '',
    
    populasi: null,
    jmlSampel: null,
    samples: [
      { id: Date.now(), dokumen: '', l1: undefined, l2: undefined, l3: undefined }
    ],
    kesimpulan: '',
    
    kondisi: '',
    kriteria: '',
    dampak: '',
    buktiFile: null,
    rcaList: [
      { id: Date.now(), method: 'People', w1: '', w2: '', w3: '' }
    ],
    
    rekomendasi: '',
    tanggapan: '',
    deskripsiAction: '',
    pic: '',
    periodAction: ''
  })

  const savedF01 = ref<any[]>([])
  const savedF02 = ref<any[]>([])
  const savedF03 = ref<any[]>([])
  const savedF04 = ref<any[]>([])
  const savedF05 = ref<any[]>([])

  const saveF01 = () => {
    savedF01.value.push({
      id: Date.now(),
      suratTugas: form.suratTugas,
      prosesBisnis: form.prosesBisnis,
      periode: `${form.periodeStart} s/d ${form.periodeEnd}`,
      lokasi: form.lokasi,
      teamMembers: form.teamMembers
    })
    alert('Data Referensi Penugasan Berhasil Disimpan!')
  }

  const saveF02 = () => {
    savedF02.value.push({
      id: Date.now(),
      resiko: form.resiko,
      taksonomi: form.taksonomi || 'Operational', // Mock fallback
      tingkatResiko: form.tingkatResiko || 'High',
      deskripsiKontrol: form.deskripsiKontrol
    })
    alert('Data Risk Profile Berhasil Disimpan!')
  }

  const saveF03 = () => {
    savedF03.value.push({
      id: Date.now(),
      populasi: form.populasi,
      jmlSampel: form.jmlSampel,
      samples: form.samples,
      kesimpulan: form.kesimpulan
    })
    alert('Data Risk Profile Berhasil Disimpan!')
  }

  const saveF04 = () => {
    savedF04.value.push({
      id: Date.now(),
      kondisi: form.kondisi,
      kriteria: form.kriteria,
      dampak: form.dampak,
      buktiFile: form.buktiFile,
      rcaList: form.rcaList.map(rca => ({ ...rca })),
      jmlRCA: form.rcaList.length
    })
    alert('Data AOI & RCA Berhasil Disimpan!')
  }

  const saveF05 = () => {
    savedF05.value.push({
      id: Date.now(),
      rekomendasi: form.rekomendasi,
      tanggapan: form.tanggapan,
      deskripsiAction: form.deskripsiAction,
      pic: form.pic,
      periodAction: form.periodAction
    })
    alert('Data Action Plan Berhasil Disimpan!')
  }

  // --- ACTIONS: Uji Sampel (F-03) ---
  const addSample = () => {
    form.samples.push({
      id: Date.now(),
      dokumen: '',
      l1: undefined,
      l2: undefined,
      l3: undefined
    })
  }

  const removeSample = (index: number) => {
    form.samples.splice(index, 1)
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
  const addRCA = () => {
    form.rcaList.push({
      id: Date.now(),
      method: 'People',
      w1: '',
      w2: '',
      w3: ''
    })
  }

  const removeRCA = (index: number) => {
    form.rcaList.splice(index, 1)
  }

  const addTeamMember = () => {
    form.teamMembers.push({
      id: Date.now(),
      name: '',
      role: ''
    })
  }

  const removeTeamMember = (index: number) => {
    if (form.teamMembers.length > 1) {
      form.teamMembers.splice(index, 1)
    }
  }

  const options = {
    assignmentLetter: ['ST-001/2026', 'ST-002/2026'],
    businessProcess: ['Procurement', 'Finance', 'HR', 'IT Operations'],
    location: ['Head Office', 'Jakarta Branch', 'Bandung Branch'],
    risk: ['R-01: Data Leakage', 'R-02: Fraud Pengadaan', 'R-03: Keterlambatan Vendor'],
    pic: ['Dimas - IT', 'Budi - Finance', 'Siti - HR'],
    testResult: [...TEST_RESULT_OPTIONS], 
    rcaMethod: [...RCA_METHOD_OPTIONS],
  }

  return {
    form, options, dateErrorMessage, isDateError, tabs,
    columnsF01, columnsF02, columnsF03, columnsF04, columnsF05,
    savedF01, savedF02, savedF03, savedF04, savedF05,
    saveF01, saveF02, saveF03, saveF04, saveF05,
    addSample, removeSample, addRCA, removeRCA, triggerUpload, onFileChange,
    checkSampleStatus, addTeamMember, removeTeamMember, getAvailableMembers, removeFile
  }
})