<template>
  <div class="space-y-4">
    <!-- Header with Statistics and Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Pengujian Kontrol</h3>
        <p class="text-sm text-gray-500">Uji efisiensi dan efektivitas kontrol yang ada</p>
      </div>
      <div class="flex gap-2">
        <div v-if="store.testControls.length > 0" class="flex gap-2 mr-2">
          <UBadge color="success" variant="solid">{{ store.effectiveControls }} Efektif</UBadge>
          <UBadge color="error" variant="solid">{{ store.ineffectiveControls }} Tidak Efektif</UBadge>
        </div>
        <UButton color="primary" icon="i-heroicons-plus" label="Tambah Pengujian" @click="store.openTestControlModal()" />
      </div>
    </div>

    <!-- Test Controls List -->
    <UCard v-if="store.testControls.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.testControls" :columns="columns">
        <template #controlName-cell="{ row }">
          <span class="font-medium">{{ row.controlName }}</span>
        </template>
        <template #controlType-cell="{ row }">
          <UBadge :color="getControlTypeColor(row.controlType)" variant="subtle">{{ row.controlType }}</UBadge>
        </template>
        <template #testResult-cell="{ row }">
          <UBadge :color="getResultColor(row.testResult)" variant="solid">{{ row.testResult }}</UBadge>
        </template>
        <template #finding-cell="{ row }">
          <span class="text-sm text-gray-600 line-clamp-2">{{ row.finding || '-' }}</span>
        </template>
        <template #mitigationPlan-cell="{ row }">
          <span class="text-sm text-gray-600 line-clamp-2">{{ row.mitigationPlan || '-' }}</span>
        </template>
        <template #actions-cell="{ index }">
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteTestControl(index)" />
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-shield-check" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">Belum ada data pengujian kontrol</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Tambah Pengujian" @click="store.openTestControlModal()" />
    </div>

    <!-- Test Control Modal -->
    <Teleport to="body">
      <div v-if="store.showTestControlModal" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Form Pengujian Kontrol</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.showTestControlModal = false" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
            <!-- Control Information -->
            <div class="bg-gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Informasi Kontrol</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Nama Kontrol" required>
                  <UInput v-model="store.testControlForm.controlName" placeholder="Contoh: Approval Purchase Order" required />
                </UFormField>
                <UFormField label="Jenis Kontrol" required>
                  <USelectMenu v-model="store.testControlForm.controlType" :items="store.options.controlTypes" placeholder="Pilih jenis kontrol" required />
                </UFormField>
              </div>
              <UFormField label="Deskripsi Kontrol" required>
                <UTextarea v-model="store.testControlForm.controlDescription" placeholder="Jelaskan bagaimana kontrol ini bekerja" required />
              </UFormField>
            </div>

            <!-- Test Procedure -->
            <div class="bg-gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Prosedur Pengujian</h4>
              <UFormField label="Langkah-langkah Pengujian" required>
                <UTextarea v-model="store.testControlForm.testProcedure" placeholder="Jelaskan langkah-langkah yang dilakukan untuk menguji kontrol ini" required />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Hasil Pengujian" required>
                  <USelectMenu v-model="store.testControlForm.testResult" :items="store.options.testResults" placeholder="Pilih hasil" required />
                </UFormField>
              </div>
            </div>

            <!-- Finding and Recommendation -->
            <div class="bg-gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Temuan dan Rekomendasi</h4>
              <UFormField label="Temuan">
                <UTextarea v-model="store.testControlForm.finding" placeholder="Jelaskan temuan dari pengujian (jika ada)" />
              </UFormField>
              <UFormField label="Rekomendasi">
                <UTextarea v-model="store.testControlForm.recommendation" placeholder="Saran perbaikan untuk mengontrol yang tidak efektif" />
              </UFormField>
            </div>

            <!-- Mitigation Plan -->
            <div class="bg-gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Rencana Mitigasi</h4>
              <UFormField label="Rencana Mitigasi">
                <UTextarea v-model="store.testControlForm.mitigationPlan" placeholder="Jelaskan rencana mitigasi untuk mengatasi kelemahan kontrol" />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="PIC (Penanggung Jawab)">
                  <UInput v-model="store.testControlForm.pic" placeholder="Nama penanggung jawab" />
                </UFormField>
                <UFormField label="Target Penyelesaian">
                  <UInput v-model="store.testControlForm.dueDate" type="date" />
                </UFormField>
              </div>
            </div>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Batal" @click="store.showTestControlModal = false" />
              <UButton color="primary" label="Simpan" @click="store.saveTestControl()" />
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

const store = useAuditFieldworkStore()

const columns = [
  { accessorKey: 'controlName', header: 'Nama Kontrol' },
  { accessorKey: 'controlType', header: 'Jenis' },
  { accessorKey: 'testResult', header: 'Hasil' },
  { accessorKey: 'finding', header: 'Temuan' },
  { accessorKey: 'mitigationPlan', header: 'Rencana Mitigasi' },
  { accessorKey: 'actions', header: '' }
]

const getControlTypeColor = (type: string) => {
  const colors: Record<string, string> = {
    'Preventive': 'success',
    'Detective': 'warning',
    'Corrective': 'info',
    'Manual': 'neutral',
    'Automated': 'primary'
  }
  return colors[type] || 'neutral'
}

const getResultColor = (result: string) => {
  const colors: Record<string, string> = {
    'Effective': 'success',
    'Ineffective': 'error',
    'Partially Effective': 'warning',
    'Not Tested': 'neutral'
  }
  return colors[result] || 'neutral'
}
</script>
