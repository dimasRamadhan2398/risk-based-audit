<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Kertas Kerja Audit</h1>
        <p class="text-sm text-gray-500">Digital Working Paper System</p>
      </div>
      <div class="flex gap-2">
        <UBadge :color="store.form.status === 'Draft' ? 'neutral' : 'success'" variant="soft">
          {{ store.form.status }}
        </UBadge>
        <UButton icon="i-heroicons-document-arrow-down" label="Export PDF" class="bg-gray" variant="ghost"/>
        <UButton icon="i-heroicons-check" label="Save & Submit" color="primary" />
      </div>
    </div>

    <UTabs :items="tabs" class="w-full">
      
      <template #header="{ item }">
        <UCard class="p-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormGroup label="No. Surat Tugas" required>
              <UInput v-model="store.form.assignmentNumber" placeholder="ST-001/AUD/2026" />
            </UFormGroup>
            <UFormGroup label="Tim Audit" required>
              <UInput v-model="store.form.teamMembers" placeholder="Ketua & Anggota" />
            </UFormGroup>
            <UFormGroup label="Periode Audit">
              <UInput v-model="store.form.period" placeholder="Januari - Desember 2025" />
            </UFormGroup>
             <UFormGroup label="Total Populasi Data" required help="Batas Maksimum Sampel">
              <UInput v-model.number="store.form.populationSize" type="number" />
            </UFormGroup>
            
            <div class="col-span-1 md:col-span-2 space-y-4 border-t pt-4 mt-2">
              <h4 class="font-bold text-primary-600">Profil Risiko</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <UFormGroup label="Deskripsi Risiko">
                  <UTextarea v-model="store.form.riskDescription" rows?="2" />
                </UFormGroup>
                <UFormGroup label="Taksonomi Risiko">
                   <USelectMenu v-model="store.form.riskTaxonomy" :options="['Strategic', 'Operational', 'Financial', 'Compliance', 'IT']" />
                </UFormGroup>
              </div>
            </div>
          </div>
        </UCard>
      </template>

      <template #steps="{ item }">
        <UCard>
          <div class="flex gap-4 items-end mb-6">
            <UFormGroup label="Tambah Langkah Pengujian Baru" class="flex-1">
              <UInput v-model="newStepInput" placeholder="Contoh: Verifikasi tanda tangan pejabat berwenang..." @keyup.enter="handleAddStep"/>
            </UFormGroup>
            <UButton icon="i-heroicons-plus" label="Tambah Langkah" @click="handleAddStep" />
          </div>

          <div v-if="store.form.testSteps.length === 0" class="text-center text-gray-500 py-8 border-2 border-dashed rounded-lg">
            Belum ada langkah pengujian didefinisikan.
          </div>

          <ul v-else class="space-y-2">
            <li v-for="(step, index) in store.form.testSteps" :key="step.id" class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded border">
              <UBadge :label="`Langkah ${index + 1}`" color="neutral" />
              <span class="font-medium">{{ step.description }}</span>
            </li>
          </ul>
        </UCard>
      </template>

      <template #matrix="{ item }">
        <UCard class="p-0">
          <div class="p-4 border-b flex justify-between items-center bg-gray-50 dark:bg-gray-900">
            <div class="flex gap-2 items-end">
               <UInput v-model="newitemInput" placeholder="ID Sampel / Dokumen" size="sm" />
               <UButton size="sm" icon="i-heroicons-plus" label="Add item" @click="handleAdditem" />
            </div>
            <div class="text-sm">
              Sampel: <b>{{ store.form.items.length }}</b> / {{ store.form.populationSize }} Populasi
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-sm text-left">
              <thead class="text-xs text-gray-700 uppercase bg-gray-100 dark:bg-gray-700 dark:text-gray-400">
                <tr>
                  <th class="px-4 py-3 sticky left-0 bg-gray-100 dark:bg-gray-700 z-10 w-48">Sampel Data</th>
                  <th v-for="(step, idx) in store.form.testSteps" :key="step.id" class="px-4 py-3 min-w-[200px]">
                    Langkah {{ idx + 1 }}
                    <UIcon name="i-heroicons-information-circle" class="ml-1 cursor-help" :title="step.description"/>
                  </th>
                  <th class="px-4 py-3 text-center sticky right-0 bg-gray-100 dark:bg-gray-700 z-10">Kesimpulan</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in store.form.items" :key="item.id" class="border-b dark:border-gray-700 hover:bg-gray-50">
                  <td class="px-4 py-3 font-medium sticky left-0 bg-white dark:bg-gray-800 z-10">
                    {{ item.auditWorkingName }}
                  </td>
                  
                  <td v-for="step in store.form.testSteps" :key="step.id" class="px-4 py-3">
                    <USelectMenu 
                        :model-value="item.results[step.id]"
                        :options="['Pass', 'Fail', 'N/A']"
                        size="xs"
                        @update:model-value="(val) => store.updateItemResult(item.id, step.id, val as 'Pass' | 'Fail' | 'N/A')"
                    >
                        <template #leading>
                            <div 
                            class="h-2 w-2 rounded-full"
                            :class="{
                                'bg-green-500': item.results[step.id] === 'Pass',
                                'bg-red-500': item.results[step.id] === 'Fail',
                                'bg-gray-400': item.results[step.id] === 'N/A'
                            }"
                            />
                        </template>

                        <template #item="{ item: option }">
                            <span :class="{
                            'text-green-600 font-bold': option === 'Pass',
                            'text-red-600 font-bold': option === 'Fail',
                            'text-gray-400': option === 'N/A'
                            }">{{ option }}</span>
                        </template>
                    </USelectMenu>
                  </td>

                  <td class="px-4 py-3 text-center sticky right-0 bg-white dark:bg-gray-800 z-10">
                    <UBadge :color="item.isEffective ? 'success' : 'error'" variant="subtle">
                      {{ item.isEffective ? 'EFEKTIF' : 'TIDAK EFEKTIF' }}
                    </UBadge>
                  </td>
                </tr>
                <tr v-if="store.form.items.length === 0">
                  <td :colspan="store.form.testSteps.length + 2" class="text-center py-8 text-gray-500">
                    Belum ada sampel data. Tambahkan di toolbar atas.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </UCard>
      </template>

      <template #findings="{ item }">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <UCard>
            <template #header><h3 class="font-bold">Area of Improvement (AOI)</h3></template>
            <div class="space-y-4">
              <UFormGroup label="Judul Temuan"><UInput v-model="store.form.findingTitle" /></UFormGroup>
              <UFormGroup label="Kondisi (Fakta)"><UTextarea v-model="store.form.condition" rows?="3" /></UFormGroup>
              <UFormGroup label="Kriteria (Acuan)"><UTextarea v-model="store.form.criteria" rows?="2" placeholder="Sesuai POJK No..."/></UFormGroup>
              <UFormGroup label="Dampak/Risiko"><UTextarea v-model="store.form.effect" rows?="2" /></UFormGroup>
              
              <UFormGroup label="Bukti Pendukung (Gambar)">
                <UInput type="file" icon="i-heroicons-photo" />
              </UFormGroup>
            </div>
          </UCard>

          <UCard>
            <template #header><h3 class="font-bold">Root Cause Analysis (Why Analysis)</h3></template>
            <div class="space-y-4">
              <UFormGroup label="Kategori Penyebab">
                <USelectMenu v-model="store.form.rcaCategory" :options="['People', 'Process', 'Policy', 'System', 'External']" />
              </UFormGroup>
              
              <div class="pl-4 border-l-2 border-primary-200 space-y-3">
                <UFormGroup label="Why 1"><UInput v-model="store.form.why1" placeholder="Mengapa masalah terjadi?" /></UFormGroup>
                <UFormGroup label="Why 2"><UInput v-model="store.form.why2" placeholder="Mengapa penyebab Why 1 terjadi?" /></UFormGroup>
                <UFormGroup label="Why 3 (Root Cause)"><UInput v-model="store.form.why3" placeholder="Akar masalah sesungguhnya..." /></UFormGroup>
              </div>

              <UFormGroup label="Kesimpulan Akar Masalah">
                <UTextarea v-model="store.form.rootCauseConclusion" class="font-semibold" />
              </UFormGroup>
            </div>
          </UCard>
        </div>
      </template>

      <template #action="{ item }">
        <UCard>
           <div class="space-y-6">
             <div class="bg-gray-50 p-4 rounded-lg border">
               <UFormGroup label="Tanggapan Auditee" required>
                 <UTextarea v-model="store.form.auditeeResponse" placeholder="Kami setuju dengan temuan..." />
               </UFormGroup>
             </div>

             <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
               <UFormGroup label="Rencana Aksi (Action Plan)">
                 <UTextarea v-model="store.form.actionPlan" rows?="3" />
               </UFormGroup>
               <UFormGroup label="Target Penyelesaian">
                 <UInput type="date" v-model="store.form.targetDate" />
                 <p v-if="isDateInvalid" class="text-xs text-red-500 mt-1">Target tidak boleh di masa lalu!</p>
               </UFormGroup>
             </div>

             <UDivider label="Approval" />
             
             <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
               <UFormGroup label="PIC (Manager)">
                 <UInput v-model="store.form.pic" icon="i-heroicons-user" />
               </UFormGroup>
               <UFormGroup label="Final Approver">
                 <USelectMenu v-model="store.form.approver" :options="['VP Audit', 'Direktur Utama', 'Komite Audit']" />
               </UFormGroup>
             </div>
           </div>
        </UCard>
      </template>

    </UTabs>
  </div>
</template>

<script setup lang="ts">
import { useAuditWorkingStore } from '~/stores/audit-work'

const store = useAuditWorkingStore()

// Tab Configuration
const tabs = [
  { slot: 'header', label: '1. Profil & Header', icon: 'i-heroicons-identification' },
  { slot: 'steps', label: '2. Langkah Kerja', icon: 'i-heroicons-list-bullet' },
  { slot: 'matrix', label: '3. Matriks Pengujian', icon: 'i-heroicons-table-cells' },
  { slot: 'findings', label: '4. Temuan & RCA', icon: 'i-heroicons-magnifying-glass-circle' },
  { slot: 'action', label: '5. Action Plan', icon: 'i-heroicons-check-badge' }
]

// State Lokal untuk Input Cepat
const newStepInput = ref('')
const newitemInput = ref('')

// Computed Validation
const isDateInvalid = computed(() => {
  if (!store.form.targetDate) return false
  return new Date(store.form.targetDate) < new Date()
})

// Handlers
const handleAddStep = () => {
  if (!newStepInput.value) return alert("Deskripsi langkah tidak boleh kosong!")
  store.addTestStep(newStepInput.value)
  newStepInput.value = ''
}

const handleAdditem = () => {
  if (store.form.testSteps.length === 0) return alert("Harap definisikan Langkah Kerja (Tab 2) terlebih dahulu!") // Error Handling Spec
  if (!newitemInput.value) return alert("ID Sampel tidak boleh kosong!")
  
  try {
    store.addItem(newitemInput.value)
    newitemInput.value = ''
  } catch (e: any) {
    alert(e.message) // Error Handling: Population Limit
  }
}
</script>

<style scoped>
/* Styling khusus untuk tabel matriks agar sticky header/column berfungsi rapi */
th, td {
  white-space: nowrap;
}
</style>