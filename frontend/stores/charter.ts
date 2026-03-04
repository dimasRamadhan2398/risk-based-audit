// stores/charter.ts
import type { TableColumn } from '@nuxt/ui';
import { defineStore } from 'pinia'
import type { AuditCharter, CharterFormState } from '~/types/audit'

export const useCharterStore = defineStore('charter', () => {

  // Modal State
  const showModal = ref(false);
  const errorMsg = ref("");

  const isEditing = ref(false);
  const editingId = ref<string | null>(null);

  const columns: TableColumn<AuditCharter>[] = [
    { accessorKey: "version", header: "Version" },
    { accessorKey: "title", header: "Charter Name" },
    { accessorKey: "date", header: "Date" },
    { accessorKey: "approvedBy", header: "Approved By" },
    { accessorKey: "uploadedBy", header: "Uploaded By" },
    { accessorKey: "fileName", header: "Actions" },
    { accessorKey: "actions", header: "" },
  ];

  // Form State
  const form = reactive<CharterFormState>({
    title: "",
    version: "", // Tidak perlu diisi user
    date: new Date().toISOString().split("T")[0] || "",
    uploadedBy: "Dimas (HIA)",
    approvedBy: "",
    isActive: true,
    file: null,
  });

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
    form.file = file;
  };

  // Submit Handler
  const handleSubmit = () => {
    // Validasi File: Wajib jika mode ADD, Opsional jika mode EDIT
    if (!isEditing.value && !form.file) {
      errorMsg.value = "Mohon upload file charter.";
      return;
    }

    if (isEditing.value && editingId.value) {
      // MODE EDIT
      updateCharter(editingId.value, { ...form });
      alert("Audit Charter berhasil diperbarui!");
    } else {
      // MODE ADD
      addCharter({ ...form });
      alert("Audit Charter berhasil diupload!");
    }

    closeModal();
  };

  const closeModal = () => {
    showModal.value = false;
    errorMsg.value = "";
    isEditing.value = false; // Reset mode edit
    editingId.value = null;

    // Reset Form
    form.title = "";
    form.version = "";
    form.date = new Date().toISOString().split("T")[0] || "";
    form.approvedBy = "";
    form.isActive = false;
    form.file = null;
  };

  const handleEdit = (charter: any) => {
    isEditing.value = true;
    editingId.value = charter.id;
    showModal.value = true;

    // Isi form dengan data lama
    form.title = charter.title;
    form.version = charter.version;
    form.date = charter.date;
    form.uploadedBy = charter.uploadedBy;
    form.approvedBy = charter.approvedBy;
    form.isActive = charter.isActive;
    form.file = null; // Reset file input karena file tidak wajib diisi saat edit

    // Reset error
    errorMsg.value = "";
  };
  // Mock Data
  const charters = ref<AuditCharter[]>([
    {
      id: '2',
      title: 'Internal Audit Charter 2026',
      version: '1.1', // Anggap ini yang terbaru
      date: '2026-01-15',
      uploadedBy: 'Dimas (HIA)',
      approvedBy: 'Audit Committee',
      isActive: true,
      fileName: 'audit-charter-v1.1_2026.pdf',
      fileSize: '3.1 MB'
    },
    {
      id: '1',
      title: 'Internal Audit Charter 2025',
      version: '1.0',
      date: '2025-01-01',
      uploadedBy: 'Dimas (HIA)',
      approvedBy: 'Board of Directors',
      isActive: false,
      fileName: 'audit-charter-v1.0_2025.pdf',
      fileSize: '2.5 MB'
    }
  ])

  // --- LOGIC BARU: GETTER OTOMATISASI VERSI ---
  const nextVersion = computed(() => {
    // Skenario 1: Jika belum ada data sama sekali, mulai dari 1.0
    if (charters.value.length === 0) return '1.0'

    // Skenario 2: Ambil versi dari data paling atas (index 0 / terbaru)
    // Kita asumsikan data selalu di-unshift (terbaru di atas)
    const latestVerStr = charters.value[0]?.version
    const latestVerNum = parseFloat(latestVerStr!)

    // Tambah 0.1 lalu jadikan string dengan 1 angka di belakang koma (fixed)
    // Contoh: 1.1 + 0.1 = 1.2
    return (latestVerNum + 0.1).toFixed(1)
  })

  // Getters Lainnya
  const activeCharter = computed(() => charters.value.find(c => c.isActive))
  const historyCharters = computed(() => charters.value.filter(c => !c.isActive))

  // Actions
  const addCharter = (form: CharterFormState) => {
    // Non-aktifkan yang lama jika yang baru Active
    if (form.isActive) {
      charters.value.forEach(c => c.isActive = false)
    }

    // Gunakan nextVersion.value untuk versi baru
    const autoVersion = nextVersion.value

    const year = new Date(form.date).getFullYear()
    const fileExt = form.file?.name.split('.').pop() || 'pdf'
    
    // Gunakan autoVersion untuk penamaan file
    const generatedFileName = `audit-charter-v${autoVersion}_${year}.${fileExt}`

    const newCharter: AuditCharter = {
      id: Date.now().toString(),
      title: form.title,
      version: autoVersion, // <--- OTOMATIS DI SINI
      date: form.date,
      uploadedBy: form.uploadedBy,
      approvedBy: form.approvedBy,
      isActive: form.isActive,
      fileName: generatedFileName,
      fileSize: form.file ? `${(form.file.size / 1024 / 1024).toFixed(2)} MB` : '0 MB',
      fileUrl: '#'
    }

    charters.value.unshift(newCharter)
  }

  const updateCharter = (id: string, form: CharterFormState) => {
    const index = charters.value.findIndex(c => c.id === id)
    if (index === -1) return

    // 1. Jika status diubah jadi Active, matikan yang lain
    if (form.isActive) {
      charters.value.forEach(c => c.isActive = false)
    }

    // 2. Cek apakah ada file baru? // Ambil referensi data lama
    const existingCharter = charters.value[index]!
    let updatedFileName = charters.value[index]?.fileName
    let updatedFileSize = charters.value[index]?.fileSize

    if (form.file) {
      const year = new Date(form.date).getFullYear()
      const fileExt = form.file.name.split('.').pop() || 'pdf'
      updatedFileName = `audit-charter-v${form.version}_${year}.${fileExt}`
      updatedFileSize = `${(form.file.size / 1024 / 1024).toFixed(2)} MB`
    }

    // 3. Update Data
    charters.value[index] = {
      ...charters.value[index], // Ambil data lama
      id: existingCharter.id,
      title: form.title,
      version: form.version,
      date: form.date,
      uploadedBy: form.uploadedBy,
      approvedBy: form.approvedBy,
      isActive: form.isActive,
      fileName: updatedFileName,
      fileSize: updatedFileSize
    }
    
    // Sort ulang history jika perlu (opsional)
  }

  // Meta & Stores
  definePageMeta({
    layout: "default",
    // middleware: 'auth' // Aktifkan jika middleware auth sudah siap
  });

  return {
    charters,activeCharter,historyCharters,nextVersion, showModal, 
    columns, form, isEditing, errorMsg, 
    addCharter,updateCharter, handleEdit, 
    handleSubmit, closeModal, handleFileChange
  }
})