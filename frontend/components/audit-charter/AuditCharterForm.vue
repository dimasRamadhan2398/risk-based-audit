<template>
    
      <UModal
        v-model:open="store.showModal"
        :dismissible="false"
        class="w-full sm:max-w-4xl bg-[var(--bg-main)] border-[var(--border-main)]"
      >
            <template #content>
              <UForm @submit.prevent="store.handleSubmit">
                <div
                  class="px-4 pt-5 pb-4 sm:p-6 sm:pb-4"
                >
                  <div class="flex justify-between items-center">
                    <h3
                      class="text-lg leading-6 font-medium mb-4"
                      id="modal-title"
                    >
                      {{ store.isEditing ? t('auditCharter.form.editTitle') : t('auditCharter.form.uploadTitle') }}
                    </h3>
                    <UIcon
                      name="close"
                      @click="store.closeModal"
                      class="text-2xl"
                      >&times;</UIcon
                    >
                  </div>

                  <div class="space-y-4">
                    <UFormField
                      :label="t('auditCharter.form.docTitle')"
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
                        :placeholder="t('auditCharter.form.docTitlePlaceholder')"
                      />
                    </UFormField>

                    <div class="grid grid-cols-2 gap-4">
                      <UFormField
                        :label="t('auditCharter.form.versionAuto')"
                        class="block text-sm font-medium"
                        size="lg"
                        disabled
                      >
                        <div
                          class="mt-1 block w-full rounded-md border border-secondary-200 text-primary-900 p-2 sm:text-sm font-bold"
                        >
                          <span v-if="store.isEditing">{{ store.form.version }}</span>
                          <span v-else 
                            >v{{ store.nextVersion }}</span
                          >
                        </div>
                      </UFormField>

                      <UFormField
                        :label="t('auditCharter.form.date')"
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
                        :label="t('auditCharter.form.uploadedBy')"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <div
                          class="mt-1 block w-full rounded-md border border-secondary-200 text-primary-900 p-2 sm:text-sm font-bold"
                        >
                          <span>{{ store.form.uploadedBy }}</span>
                        </div>
                      </UFormField>

                      <UFormField
                        :label="t('auditCharter.form.approvedBy')"
                        class="block text-sm font-medium"
                        size="lg"
                      >
                        <UInput
                          v-model="store.form.approvedBy"
                          required
                          type="text"
                          class="mt-1 block w-full rounded-md sm:text-sm"
                          :placeholder="t('auditCharter.form.approvedByPlaceholder')"
                        />
                      </UFormField>
                    </div>

                    <UFormField
                      :label="t('auditCharter.form.status')"
                      class="block text-sm font-medium"
                      size="lg"
                    >
                      <URadioGroup
                        v-model="store.form.isActive"
                        :items="[
                          { label: t('auditCharter.form.active'), value: true },
                          { label: t('auditCharter.form.inactive'), value: false }
                        ]"
                        orientation="horizontal"
                        class="mt-2"
                      />
                    </UFormField>

                    <UFormField
                      :label="t('auditCharter.form.uploadFile')"
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
                        class="mt-2 flex items-center gap-2 text-sm p-2 rounded"
                      >
                        <UIcon name="i-heroicons-document" />
                        <span class="font-bold">{{ store.form.file.name }}</span>
                      </div>
                    </UFormField>

                  </div>
                </div>
                <div
                  class="px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse"
                >
                  <UButton
                    type="submit"
                    color="primary"
                    class="w-full inline-flex justify-center rounded-md px-4 py-2 sm:ml-3 sm:w-auto sm:text-sm"
                    :label="t('auditCharter.form.submit')"
                  >
                  </UButton>
                </div>
              </UForm>
            </template>
      </UModal>
    
</template>

<script setup lang="ts">
import { useCharterStore } from '~/stores/charter'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useCharterStore()
</script>