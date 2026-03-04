// stores/surat-tugas.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import {type SuratTugas, type SuratTugasForm, type SuratTugasStatus } from '~/types/audit'

export const useSuratTugasStore = defineStore('surat-tugas', () => {
  // State: Menyimpan daftar surat tugas
  const suratTugasList = ref<SuratTugas[]>([])
  
  // ==========================================
  // DEFINISI KOLOM TABEL
  // ==========================================
  const columns = [
    { accessorKey: 'nomorSurat', header: 'Nomor Surat' },
    { accessorKey: 'judulAudit', header: 'Judul / Objek Audit' },
    { accessorKey: 'unitKerja', header: 'Unit Kerja' },
    { accessorKey: 'periode', header: 'Periode Pelaksanaan' },
    { accessorKey: 'timAudit', header: 'Tim Audit' },
    { accessorKey: 'status', header: 'Status' }
  ]

  // ==========================================
  // STATE: MODAL & FORM
  // ==========================================
  const isModalOpen = ref(false)

  const form = reactive({
    judulAudit: '',
    dipimpinOleh: '',
    category: '',
    tahunAudit: '2026',
    timAudit: 'SKAI',
    periodeMulai: '',
    periodeSelesai: '',
    unitKerja: '',
    listAnggota: [
      { nama: '', peran: 'Ketua' },
      { nama: '', peran: 'Anggota' }
    ],
    listTujuan: [''],
    listRuangLingkup: [''],
    listTembusan: ['Direktur Utama']
  })

  // ==========================================
  // FUNGSI HELPER FORM (DINAMIS)
  // ==========================================
  const addItem = (list: any[], defaultItem: any) => {
    // Deep clone untuk objek agar tidak ada referensi yang nyangkut
    list.push(typeof defaultItem === 'object' ? { ...defaultItem } : defaultItem)
  }

  const removeItem = (list: any[], index: number) => {
    list.splice(index, 1)
  }

  // ==========================================
  // VALIDASI ERROR HANDLING (F-05)
  // ==========================================
  const dateError = computed(() => {
    if (form.periodeMulai && form.periodeSelesai) {
      const start = new Date(form.periodeMulai)
      const end = new Date(form.periodeSelesai)
      if (end < start) {
        return "Kesalahan: Tanggal selesai tidak boleh sebelum tanggal mulai."
      }
    }
    return null
  })

  // ==========================================
  // AKSI BUKA/TUTUP & SUBMIT
  // ==========================================
  const openModal = () => {
    // Reset form saat modal dibuka
    Object.assign(form, {
      judulAudit: '',
      dipimpinOleh: '',
      category: '',
      tahunAudit: new Date().getFullYear().toString(),
      timAudit: 'SKAI',
      periodeMulai: '',
      periodeSelesai: '',
      unitKerja: '',
      listAnggota: [{ nama: '', peran: 'Ketua' }],
      listTujuan: [''],
      listRuangLingkup: [''],
      listTembusan: ['Direktur Utama']
    })
    isModalOpen.value = true
  }

  const closeModal = () => {
    isModalOpen.value = false
  }

  const handleSubmit = () => {
    // 1. Validasi Unit Kerja Kosong (F-05)
    if (!form.unitKerja) {
      alert("Unit kerja harus diisi untuk menentukan tujuan surat.")
      return
    }

    // 2. Validasi Tanggal (F-05)
    if (dateError.value || !form.periodeMulai || !form.periodeSelesai) {
      alert(dateError.value || "Harap isi periode mulai dan selesai.")
      return
    }

    // 3. Validasi Anggota Tim Minimal 3 (F-02)
    if (form.listAnggota.length < 3) {
      const proceed = confirm("Template menyarankan minimal 3 anggota tim. Lanjutkan menyimpan?")
      if (!proceed) return
    }

    // 4. Generate Nomor Surat Otomatis (F-01)
    // Format: ST-001/SKAI/2026
    const countCount = (suratTugasList.value.length + 1).toString().padStart(3, '0')
    const nomorSuratGenerate = `ST-${countCount}/${form.timAudit}/${form.tahunAudit}`

    // 5. Simpan Data ke Tabel
    suratTugasList.value.unshift({
      id: Date.now().toString(),
      nomorSurat: nomorSuratGenerate,
      status: 'Draft',
      // Clone form data
      ...JSON.parse(JSON.stringify(form))
    })

    // 6. Tutup Modal
    closeModal()
  }

  // Helper: Membuat nomor surat berurutan berdasarkan tim dan tahun
  // Contoh: ST-001/SKAI/2026
  const generateNomorSurat = (timAudit: string, tahun: string) => {
    const nextCount = suratTugasList.value.length + 1
    const paddedCount = nextCount.toString().padStart(3, '0') 
    return `ST-${paddedCount}/${timAudit}/${tahun}`
  }

  // Action: Tambah Surat Tugas (Dari Form)
  // stores/surat-tugas.ts
const addSuratTugas = (form: SuratTugasForm) => {
  // Buat objek baru untuk memastikan semua field 'string' terisi
  const newEntry: SuratTugas = {
    ...form, // Ambil semua field dari form
    id: Date.now().toString(), // Generate ID di sini
    nomorSurat: generateNomorSurat(form.timAudit, form.tahunAudit), // Generate No Surat
    status: 'Draft',
    createdAt: new Date().toISOString()
  }

  suratTugasList.value.unshift(newEntry)
}

  // Action: Edit Data Surat Tugas
// const updateSuratTugas = (id: string, updatedForm: SuratTugasForm) => {
//   const index = suratTugasList.value.findIndex(s => s.id === id)
  
//   if (index !== -1) {
//     const existingSurat = suratTugasList.value[index]
    
//     // Gabungkan: Properti sistem tetap, properti isi form diperbarui
//     suratTugasList.value[index] = {
//       ...existingSurat, // id, nomorSurat, status, createdAt tetap ada
//       ...updatedForm    // timpa dengan isi form terbaru
//     }
//   }
// }

  // Action: Hapus Data
  const deleteSuratTugas = (id: string) => {
    suratTugasList.value = suratTugasList.value.filter(s => s.id !== id)
  }

  // Action: Ubah Status (Misal dari Draft -> Published)
  const changeStatus = (id: string, status: SuratTugasStatus) => {
    const surat = suratTugasList.value.find(s => s.id === id)
    if (surat) {
      surat.status = status
    }
  }

  const options = {
    timAudit: ['SKAI' , 'DAI' , 'CAE'],
    unitKerja: ['Produksi' , 'Pemasaran' , 'Keuangan'],
    peran: ['Penanggung Jawab' , 'Pengawas' , 'Ketua' , 'Anggota'],
  }

  return {
    suratTugasList, columns, isModalOpen, form, dateError, options,
    generateNomorSurat, addSuratTugas, deleteSuratTugas, changeStatus, 
    openModal, closeModal, removeItem, addItem, handleSubmit
  }
})