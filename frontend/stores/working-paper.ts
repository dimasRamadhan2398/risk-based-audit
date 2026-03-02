import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import type { WorkingPaperForm, SampleItem } from '~/types/audit'
import { RCA_CATEGORY_OPTIONS, TEST_RESULT_OPTIONS } from '~/types/audit'

export const useWorkingPaperStore = defineStore('working-paper', () => {
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
      { id: Date.now(), kategori: 'People', w1: '', w2: '', w3: '' }
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
      kategori: 'People',
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
    suratTugas: ['ST-001/2026', 'ST-002/2026'],
    prosesBisnis: ['Pengadaan Barang', 'Keuangan', 'SDM', 'Operasional IT'],
    lokasi: ['Kantor Pusat', 'Cabang Jakarta', 'Cabang Bandung'],
    resiko: ['R-01: Kebocoran Data', 'R-02: Fraud Pengadaan', 'R-03: Keterlambatan Vendor'],
    pic: ['Dimas - IT', 'Budi - Keuangan', 'Siti - SDM'],
    testResult: [...TEST_RESULT_OPTIONS], 
    rcaCategory: [...RCA_CATEGORY_OPTIONS],
  }

  return {
    form, options,
    savedF01, savedF02, savedF03, savedF04, savedF05,
    saveF01, saveF02, saveF03, saveF04, saveF05,
    addSample, removeSample, addRCA, removeRCA, checkSampleStatus, addTeamMember, removeTeamMember
  }
})