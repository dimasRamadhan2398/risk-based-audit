<template>
    
      <UModal
        v-model:open="store.showModal"
        :dismissible="false"
        class="w-full sm:max-w-4xl"
      >
      <div></div>
            <template #content>
              <UForm @submit.prevent="store.handleSubmit">
                <div
                  class="bg-secondary-50 dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4"
                >
                  <div class="flex justify-between items-center">
                    <h3
                      class="text-lg leading-6 font-medium text-gray-900 dark:text-white mb-4"
                      id="modal-title"
                    >
                      Upload New Charter
                    </h3>
                    <UIcon
                      name="close"
                      @click="store.closeModal"
                      class="text-primary-400 hover:text-primary-600 text-2xl"
                      >&times;</UIcon
                    >
                  </div>

                  <div class="space-y-4">
                    <UFormField
                      label="Judul Dokumen"
                      class="block text-sm font-medium"
                      size="lg"
                    >
                      <UInput
                        v-model="store.form.title"
                        required
                        type="text"
                        name="title"
                        id="title"
                        class="mt-1 block w-full rounded-md sm:text-sm p-2"
                        placeholder="e.g. Internal Audit Charter 2026"
                      />
                    </UFormField>

                    <div class="grid grid-cols-2 gap-4">
                      <UFormField
                        label="Versi (Auto)"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <div
                          class="mt-1 block w-full rounded-md border border-gray-200 bg-gray-200 p-2 text-gray-600 sm:text-sm font-bold"
                        >
                          <span v-if="store.isEditing">{{ store.form.version }}</span>
                          <span v-else class="text-primary-600"
                            >v{{ store.nextVersion }}</span
                          >
                        </div>
                      </UFormField>

                      <UFormField
                        label="Tanggal"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <UInput
                          v-model="store.form.date"
                          required
                          type="date"
                          class="mt-1 block w-full rounded-md sm:text-sm"
                        />
                      </UFormField>
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                      <UFormField
                        label="Diupload Oleh"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <div
                          class="mt-1 block w-full rounded-md border border-gray-200 bg-gray-200 p-2 text-gray-600 sm:text-sm font-bold"
                        >
                          <span>{{ store.form.uploadedBy }}</span>
                        </div>
                      </UFormField>

                      <UFormField
                        label="Disetujui Oleh"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <UInput
                          v-model="store.form.approvedBy"
                          required
                          type="text"
                          class="mt-1 block w-full rounded-md sm:text-sm"
                          placeholder="e.g. Audit Committee"
                        />
                      </UFormField>
                    </div>

                    <UFormField
                      label="Status"
                      class="block text-sm font-medium"
                      size="lg"
                    >
                      <URadioGroup
                        v-model="store.form.isActive"
                        :items="[
                          { label: 'Active', value: true },
                          { label: 'Inactive', value: false }
                        ]"
                        orientation="horizontal"
                        class="mt-2"
                      />
                    </UFormField>

                    <UFormField
                      label="Upload File Charter (PDF/DOCX)"
                      :error="store.errorMsg"
                      class="block text-sm font-medium"
                      size="lg"
                    >
                      <UInput
                        type="file"
                        icon="i-heroicons-paper-clip"
                        @change="store.handleFileChange"
                        accept=".pdf,.docx,.doc"
                        class="w-full"
                      />

                      <div
                        v-if="store.form.file"
                        class="mt-2 flex items-center gap-2 text-sm text-primary-600 bg-primary-50 p-2 rounded"
                      >
                        <UIcon name="i-heroicons-document" />
                        <span class="font-bold">{{ store.form.file.name }}</span>
                      </div>
                    </UFormField>

                  </div>
                </div>
                <div
                  class="bg-secondary-50 dark:bg-gray-700 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse"
                >
                  <UButton
                    type="submit"
                    color="primary"
                    class="w-full inline-flex justify-center rounded-md px-4 py-2 sm:ml-3 sm:w-auto sm:text-sm"
                    label="Submit"
                  >
                  </UButton>
                </div>
              </UForm>
            </template>
      </UModal>
    
</template>

<script setup lang="ts">
import { useCharterStore } from '~/stores/charter'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useCharterStore()
</script>