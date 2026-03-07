<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center p-4">
      <div>
        <h2 class="text-lg font-semibold">Daftar Interview</h2>
        <p class="text-sm text-gray-500">Kelola daftar interview dengan interviewee dan interviewer</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" label="Tambah Interview" @click="store.openInterviewModal()" />
    </div>

    <!-- Interview List -->
    <UCard v-if="store.interviews.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.interviews" :columns="columns">
        <template #interviewee-cell="{ row }">
          <div>
            <p class="font-medium">{{ row.interviewee }}</p>
            <p class="text-xs text-gray-500">{{ row.intervieweePosition }}</p>
          </div>
        </template>
        <template #interviewer-cell="{ row }">
          <div>
            <p class="font-medium">{{ row.interviewer }}</p>
            <p class="text-xs text-gray-500">{{ row.interviewerPosition }}</p>
          </div>
        </template>
        <template #topic-cell="{ row }">
          <UBadge color="primary" variant="subtle">{{ row.topic }}</UBadge>
        </template>
        <template #file-cell="{ row }">
          <UButton v-if="row.file" icon="i-heroicons-document-arrow-down" color="neutral" variant="ghost" size="sm">
            {{ row.file.name }}
          </UButton>
          <span v-else class="text-gray-400 text-sm">-</span>
        </template>
        <template #actions-cell="{ index }">
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteInterview(index)" />
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-microphone" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">Belum ada data interview</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Tambah Interview" @click="store.openInterviewModal()" />
    </div>

    <!-- Interview Modal -->
    <Teleport to="body">
      <div v-if="store.showInterviewModal" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Form Interview</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.showInterviewModal = false" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveInterview()" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Interviewee (Nama yang diinterview)" required>
                <USelectMenu
                  v-model="store.interviewForm.interviewee"
                  :items="store.memberOptions"
                  placeholder="Pilih nama dari tim audit"
                  :disabled="!store.hasSelectedAssignmentLetter"
                  required
                />
              </UFormField>
              <UFormField label="Jabatan Interviewee" required>
                <USelectMenu v-model="store.interviewForm.intervieweePosition" :items="store.options.positions" placeholder="Pilih jabatan" required />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Interviewer (Pewawancara)" required>
                <USelectMenu
                  v-model="store.interviewForm.interviewer"
                  :items="store.memberOptions"
                  placeholder="Pilih nama dari tim audit"
                  :disabled="!store.hasSelectedAssignmentLetter"
                  required
                />
              </UFormField>
              <UFormField label="Jabatan Interviewer" required>
                <USelectMenu v-model="store.interviewForm.interviewerPosition" :items="store.options.positions" placeholder="Pilih jabatan" required />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Tanggal Interview" required>
                <UInput v-model="store.interviewForm.date" type="date" required />
              </UFormField>
              <UFormField label="Topik Interview" required>
                <USelectMenu v-model="store.interviewForm.topic" :items="store.options.auditTopics" placeholder="Pilih topik" required />
              </UFormField>
            </div>

            <UFormField label="Upload File (PDF/DOCX)">
              <UInput
                type="file"
                icon="i-heroicons-paper-clip"
                @change="store.handleInterviewFileChange"
                accept=".pdf,.docx,.doc"
                class="w-full"
              />
              <div v-if="store.interviewForm.file" class="mt-2 flex items-center gap-2">
                <UIcon name="i-heroicons-document" />
                <span class="font-bold text-sm">{{ store.interviewForm.file.name }}</span>
              </div>
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Batal" @click="store.showInterviewModal = false" />
              <UButton color="primary" label="Simpan" @click="store.saveInterview()" />
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
  { accessorKey: 'interviewee', header: 'Interviewee' },
  { accessorKey: 'interviewer', header: 'Interviewer' },
  { accessorKey: 'date', header: 'Tanggal' },
  { accessorKey: 'topic', header: 'Topik' },
  { accessorKey: 'file', header: 'File' },
  { accessorKey: 'actions', header: '' }
]
</script>
