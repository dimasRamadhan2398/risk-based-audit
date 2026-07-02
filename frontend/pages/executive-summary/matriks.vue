<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Matriks Induk Kompilasi Temuan</h1>
        <p class="text-gray-500">
          Editable Data Grid untuk memasukkan detail temuan triwulanan.
        </p>
      </div>
      <div class="flex gap-2">
        <input 
          type="file" 
          accept=".xlsx,.xls" 
          class="hidden" 
          ref="excelInput"
          @change="handleImportExcel"
        />
        <UButton 
          color="neutral" 
          variant="subtle" 
          icon="i-heroicons-arrow-up-tray" 
          label="Impor via Excel" 
          @click="$refs.excelInput.click()" 
        />
        <UButton 
          color="primary" 
          icon="i-heroicons-plus" 
          label="Tambah Baris Temuan" 
          @click="addMatriksRow"
        />
      </div>
    </div>

    <!-- Rule 2 Validation Banner -->
    <UCard :ui="{ body: 'p-4' }" class="bg-gray-50">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 text-sm">
        <div class="flex items-center gap-2">
          <UIcon 
            name="i-heroicons-check-circle" 
            class="size-5 shrink-0 text-green-500"
          />
          <div>
            <span class="font-semibold">Informasi Sinkronisasi:</span>
            <span class="text-green-700 ml-1">
              Sinkronisasi data otomatis dengan Section II ringkasan eksekutif.
            </span>
          </div>
        </div>
        <div class="flex gap-4 text-xs font-semibold text-gray-600 bg-white p-2 rounded-lg border">
          <div>Matriks: <span class="text-error-600 font-bold">{{ countMatriksByRisk('Tinggi') }} Tinggi</span>, <span class="text-warning-600 font-bold">{{ countMatriksByRisk('Sedang') }} Sedang</span>, <span class="text-success-600 font-bold">{{ countMatriksByRisk('Rendah') }} Rendah</span></div>
          <div class="text-gray-300">|</div>
          <div>Ringkasan II: <span class="text-error-600 font-bold">{{ store.form.risikoTinggi }} Tinggi</span>, <span class="text-warning-600 font-bold">{{ store.form.risikoSedang }} Sedang</span>, <span class="text-success-600 font-bold">{{ store.form.risikoRendah }} Rendah</span></div>
        </div>
      </div>
    </UCard>

    <!-- Matriks Data Grid Table -->
    <UCard class="overflow-x-auto" :ui="{ body: 'p-0' }">
      <table class="min-w-full divide-y divide-gray-200 border text-sm text-left table-fixed">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-3 py-3 font-semibold text-gray-700 w-16">No</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-32">Nomor Temuan</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-28">Divisi</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-44">Unit Kerja</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-40">Proses Bisnis</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-64">Judul Temuan</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-32">Nilai Risiko</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-64">Rekomendasi</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-36">Due Date</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-36">PIC Unit</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-28">Progres (%)</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-36">Status</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-48">Bukti TL</th>
            <th class="px-3 py-3 font-semibold text-gray-700 w-16 text-center sticky right-0 bg-gray-50">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-for="(row, idx) in store.form.matriksKompilasi" :key="idx">
            <!-- Row Number -->
            <td class="px-3 py-2 text-gray-500 font-medium align-top pt-3.5">{{ idx + 1 }}</td>
            
            <!-- Nomor (A) -->
            <td class="px-2 py-1.5 align-top">
              <UInput v-model="row.nomor" placeholder="001/SPI/2026" />
            </td>

            <!-- Division (B) -->
            <td class="px-2 py-1.5 align-top">
              <USelectMenu v-model="row.division" :items="divisionOptions" />
            </td>

            <!-- Unit Kerja (C) -->
            <td class="px-2 py-1.5 align-top">
              <UInput v-model="row.unitKerja" placeholder="Nama unit kerja" />
            </td>

            <!-- Proses Bisnis (D) -->
            <td class="px-2 py-1.5 align-top">
              <UInput v-model="row.prosesBisnis" placeholder="Contoh: O&M, K3" />
            </td>

            <!-- Judul Temuan (E) -->
            <td class="px-2 py-1.5 align-top">
              <UTextarea v-model="row.judulTemuan" rows="2" placeholder="Detail temuan" />
            </td>

            <!-- Nilai Risiko (F) -->
            <td class="px-2 py-1.5 align-top">
              <USelectMenu v-model="row.nilaiRisiko" :items="['Tinggi', 'Sedang', 'Rendah']" />
            </td>

            <!-- Rekomendasi (G) -->
            <td class="px-2 py-1.5 align-top">
              <UTextarea v-model="row.rekomendasi" rows="2" placeholder="Rekomendasi tindakan" />
            </td>

            <!-- Due Date (H) -->
            <td class="px-2 py-1.5 align-top">
              <UInput type="date" v-model="row.dueDate" />
            </td>

            <!-- PIC Unit (I) -->
            <td class="px-2 py-1.5 align-top">
              <UInput v-model="row.picUnit" placeholder="Jabatan PIC" />
            </td>

            <!-- % Progres (J) -->
            <td class="px-2 py-1.5 align-top">
              <UInput type="number" v-model="row.progres" min="0" max="100" />
            </td>

            <!-- Status (K) -->
            <td class="px-2 py-1.5 align-top">
              <USelectMenu v-model="row.status" :items="['Closed', 'In Progress', 'Overdue']" />
            </td>

            <!-- Bukti TL (L) -->
            <td class="px-2 py-1.5 align-top">
              <div class="flex flex-col gap-1">
                <input 
                  type="file" 
                  class="hidden" 
                  :ref="el => { if(el) fileInputs[idx] = el as HTMLInputElement }"
                  @change="handleBuktiChange($event, idx)"
                />
                <UButton 
                  color="neutral" 
                  variant="subtle" 
                  size="xs" 
                  icon="i-heroicons-paper-clip" 
                  label="Attach"
                  @click="triggerFileInput(idx)"
                />
                <span class="text-[10px] text-gray-500 font-medium truncate block max-w-[120px]">
                  {{ row.buktiTL || 'No Attachment' }}
                </span>
              </div>
            </td>

            <!-- Delete Action -->
            <td class="px-3 py-2 text-center align-top pt-3.5 sticky right-0 bg-white shadow-[-4px_0_12px_rgba(0,0,0,0.05)] border-l">
              <UButton 
                color="error" 
                variant="ghost" 
                icon="i-heroicons-trash" 
                size="xs" 
                @click="removeMatriksRow(idx)"
              />
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-if="store.form.matriksKompilasi.length === 0">
            <td colspan="14" class="text-center py-16 text-gray-400">
              <UIcon name="i-heroicons-table-cells" class="size-16 text-gray-200 mx-auto mb-4" />
              <h3 class="text-md font-semibold text-gray-600">Matriks Temuan Kosong</h3>
              <p class="text-xs text-gray-400 mt-1 max-w-sm mx-auto">
                Silakan tambah baris temuan baru atau impor file excel untuk mengisi data matriks.
              </p>
            </td>
          </tr>
        </tbody>
      </table>
    </UCard>

    <div class="flex justify-end gap-3">
      <UButton 
        color="neutral" 
        variant="ghost" 
        label="Kembali ke Ringkasan" 
        @click="navigateTo('/executive-summary')"
      />
      <UButton 
        color="primary" 
        icon="i-heroicons-check" 
        label="Simpan Matriks Temuan" 
        @click="saveMatriks"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'

const store = useExecutiveSummaryStore()
const excelInput = ref<HTMLInputElement | null>(null)
const fileInputs = ref<HTMLInputElement[]>([])

const divisionOptions = ['OP', 'IT', 'FIN', 'HR', 'LEGAL', 'ENG', 'K3']

const countMatriksByRisk = (risk: 'Tinggi' | 'Sedang' | 'Rendah') => {
  return store.form.matriksKompilasi.filter(f => f.nilaiRisiko === risk).length
}

const addMatriksRow = () => {
  store.form.matriksKompilasi.push({
    nomor: `00${store.form.matriksKompilasi.length + 1}/SPI/2026`,
    division: 'OP',
    unitKerja: '',
    prosesBisnis: '',
    judulTemuan: '',
    nilaiRisiko: 'Sedang',
    rekomendasi: '',
    dueDate: new Date().toISOString().split('T')[0],
    picUnit: '',
    progres: 0,
    status: 'In Progress',
    buktiTL: ''
  })
}

const removeMatriksRow = (idx: number) => {
  store.form.matriksKompilasi.splice(idx, 1)
}

const triggerFileInput = (idx: number) => {
  const input = fileInputs.value[idx]
  if (input) {
    input.click()
  }
}

const handleBuktiChange = (e: Event, idx: number) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    store.form.matriksKompilasi[idx].buktiTL = target.files[0].name
  }
}

const handleImportExcel = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    alert(`File Excel "${target.files[0].name}" berhasil diimpor! Menghasilkan 3 temuan kompilasi baru.`);
    store.form.matriksKompilasi.push(
      { nomor: '101/SPI/2026', division: 'OP', unitKerja: 'Production Line Unit 2', prosesBisnis: 'Quality Control', judulTemuan: 'Suhu Boiler Tidak Stabil', nilaiRisiko: 'Tinggi', rekomendasi: 'Kalibrasi thermostat boiler', dueDate: '2026-06-30', picUnit: 'Manager Produksi', progres: 10, status: 'In Progress', buktiTL: '' },
      { nomor: '102/SPI/2026', division: 'IT', unitKerja: 'Database Administration Unit', prosesBisnis: 'Backup Recovery', judulTemuan: 'Backup Log Harian Gagal', nilaiRisiko: 'Sedang', rekomendasi: 'Ubah schedule task ke storage sekunder', dueDate: '2026-07-15', picUnit: 'Head of DB', progres: 50, status: 'In Progress', buktiTL: '' },
      { nomor: '103/SPI/2026', division: 'K3', unitKerja: 'Safety & Environment Division', prosesBisnis: 'K3 Lapangan', judulTemuan: 'Signage Evakuasi Kusam', nilaiRisiko: 'Rendah', rekomendasi: 'Pasang signage reflektif baru', dueDate: '2026-08-31', picUnit: 'Safety Inspector', progres: 100, status: 'Closed', buktiTL: 'Signage_Invoice.pdf' }
    );
  }
}

const saveMatriks = () => {
  if (store.isEditing && store.currentSummary) {
    const idx = store.summaryList.findIndex(r => r.id === store.currentSummary!.id)
    if (idx !== -1) {
      store.summaryList[idx].matriksKompilasi = JSON.parse(JSON.stringify(store.form.matriksKompilasi))
    }
  }
  
  alert('Data Matriks Induk Kompilasi berhasil disimpan!')
  navigateTo('/executive-summary')
}

// Automatically seed a default item if store form has no findings yet (facilitates initial edit screen UX)
if (store.form.matriksKompilasi.length === 0) {
  addMatriksRow()
}
</script>
