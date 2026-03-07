<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Pengumpulan Dokumen</h3>
        <p class="text-sm text-gray-500">Kelola dokumen yang dibutuhkan untuk pelaksanaan audit</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Tambah Dokumen" @click="store.openDocumentModal()" />
    </div>

    <!-- Document List -->
    <UCard v-if="store.documents.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.documents" :columns="columns">
        <template #documentName-cell="{ row }">
          <span class="font-medium">{{ row.documentName }}</span>
        </template>
        <template #description-cell="{ row }">
          <span class="text-sm text-gray-600">{{ row.description }}</span>
        </template>
        <template #requiredDate-cell="{ row }">
          <UBadge color="warning" variant="subtle">{{ row.requiredDate }}</UBadge>
        </template>
        <template #file-cell="{ row }">
          <UButton v-if="row.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
            {{ row.file.name }}
          </UButton>
          <span v-else class="text-gray-400 text-sm">-</span>
        </template>
        <template #actions-cell="{ index }">
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteDocument(index)" />
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-document-duplicate" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">Belum ada dokumen yang diminta</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Tambah Dokumen" @click="store.openDocumentModal()" />
    </div>

    <!-- Document Modal -->
    <Teleport to="body">
      <div v-if="store.showDocumentModal" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Form Pengumpulan Dokumen</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.showDocumentModal = false" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveDocument()" class="space-y-4">
            <UFormField label="Nama Dokumen" required>
              <UInput v-model="store.documentForm.documentName" placeholder="Contoh: SOP Pengadaan Barang" required />
            </UFormField>

            <UFormField label="Deskripsi Dokumen" required>
              <UTextarea v-model="store.documentForm.description" placeholder="Jelaskan kegunaan dan scope dokumen ini" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Tanggal Diperlukan" required>
                <UInput v-model="store.documentForm.requiredDate" type="date" required />
              </UFormField>
            </div>

            <UFormField label="Upload File (PDF/DOCX/XLSX)">
              <UInput
                type="file"
                icon="i-heroicons-paper-clip"
                @change="store.handleDocumentFileChange"
                accept=".pdf,.docx,.doc,.xlsx,.xls"
                class="w-full"
              />
              <div v-if="store.documentForm.file" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.documentForm.file.name }}</span>
              </div>
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Batal" @click="store.showDocumentModal = false" />
              <UButton color="primary" label="Simpan" @click="store.saveDocument()" />
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
  { accessorKey: 'description', header: 'Deskripsi' },
  { accessorKey: 'requiredDate', header: 'Tgl Diperlukan' },
  { accessorKey: 'file', header: 'File' },
  { accessorKey: 'actions', header: '' }
]
</script>
