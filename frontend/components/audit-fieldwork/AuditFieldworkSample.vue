<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Data Sampel</h3>
        <p class="text-sm text-gray-500">Kelola sample data untuk pengujian audit</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Tambah Sample" @click="store.openSampleModal()" />
    </div>

    <!-- Sample List -->
    <UCard v-if="store.samples.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.samples" :columns="columns">
        <template #documentName-cell="{ row }">
          <span class="font-medium">{{ row.documentName }}</span>
        </template>
        <template #documentNumber-cell="{ row }">
          <UBadge color="neutral" variant="subtle">{{ row.documentNumber }}</UBadge>
        </template>
        <template #date-cell="{ row }">
          <span>{{ row.date }}</span>
        </template>
        <template #description-cell="{ row }">
          <span class="text-sm text-gray-600">{{ row.description }}</span>
        </template>
        <template #actions-cell="{ index }">
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteSample(index)" />
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-table-cells" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">Belum ada data sample</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Tambah Sample" @click="store.openSampleModal()" />
    </div>

    <!-- Sample Modal -->
    <Teleport to="body">
      <div v-if="store.showSampleModal" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Form Data Sampel</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.showSampleModal = false" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveSample()" class="space-y-4">
            <UFormField label="Nama Dokumen" required>
              <UInput v-model="store.sampleForm.documentName" placeholder="Contoh: Invoice Pembelian" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Nomor Dokumen" required>
                <UInput v-model="store.sampleForm.documentNumber" placeholder="Contoh: INV/2026/001" required />
              </UFormField>
              <UFormField label="Tanggal Dokumen" required>
                <UInput v-model="store.sampleForm.date" type="date" required />
              </UFormField>
            </div>

            <UFormField label="Deskripsi" required>
              <UTextarea v-model="store.sampleForm.description" placeholder="Jelaskan isi atau detail dari sample ini" required />
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Batal" @click="store.showSampleModal = false" />
              <UButton color="primary" label="Simpan" @click="store.saveSample()" />
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
  { accessorKey: 'documentName', header: 'Nama Dokumen' },
  { accessorKey: 'documentNumber', header: 'Nomor' },
  { accessorKey: 'date', header: 'Tanggal' },
  { accessorKey: 'description', header: 'Deskripsi' },
  { accessorKey: 'actions', header: '' }
]
</script>
