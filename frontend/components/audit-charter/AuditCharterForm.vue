<template>
  <UModal
    v-model:open="store.showModal"
    :dismissible="false"
    class="w-full sm:max-w-4xl bg-[var(--bg-main)] border-[var(--border-main)]"
  >
    <template #content>
      <UForm @submit.prevent="store.handleSubmit">
        <div class="px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg leading-6 font-bold" id="modal-title">
              {{ store.isEditing ? t('auditCharter.form.editTitle') : t('auditCharter.form.uploadTitle') }}
            </h3>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-lucide-x"
              @click="store.closeModal"
            />
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
                maxlength="200"
                class="mt-1 block w-full rounded-md"
                :placeholder="t('auditCharter.form.docTitlePlaceholder')"
                @invalid="(e: { target: HTMLInputElement; }) => (e.target as HTMLInputElement).setCustomValidity('Judul maksimal 200 karakter dan wajib diisi')"
                @input="(e: { target: HTMLInputElement; }) => (e.target as HTMLInputElement).setCustomValidity('')"
              />
            </UFormField>

            <div class="grid grid-cols-2 gap-4">
              <UFormField
                :label="t('auditCharter.form.versionAuto')"
                class="block text-sm font-medium"
                size="lg"
                disabled
              >
                <div class="mt-1 block w-full rounded-md border border-secondary-200 text-primary-900 p-2 sm:text-sm font-bold bg-gray-50">
                  <span v-if="store.isEditing">{{ store.form.version }}</span>
                  <span v-else>v{{ store.nextVersion }}</span>
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
                  class="mt-1 block w-full rounded-md"
                />
              </UFormField>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <UFormField
                :label="t('auditCharter.form.uploadedBy')"
                class="block text-sm font-medium"
                size="lg"
              >
                <div class="mt-1 block w-full rounded-md border border-secondary-200 text-primary-900 p-2 sm:text-sm font-bold bg-gray-50">
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
                  maxlength="200"
                  class="mt-1 block w-full rounded-md"
                  :placeholder="t('auditCharter.form.approvedByPlaceholder')"
                  @invalid="(e: { target: HTMLInputElement; }) => (e.target as HTMLInputElement).setCustomValidity('Penyetuju maksimal 200 karakter dan wajib diisi')"
                  @input="(e: { target: HTMLInputElement; }) => (e.target as HTMLInputElement).setCustomValidity('')"
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
              class="block text-sm font-medium"
              size="lg"
              :required="!store.isEditing"
            >
              <div class="mt-1 flex items-center gap-4">
                <input
                  type="file"
                  accept=".pdf,.docx,.doc"
                  @change="store.handleFileChange"
                  class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
                  :required="!store.isEditing && !store.form.file"
                />
              </div>
              <p v-if="store.form.file" class="text-md text-gray-500 mt-1">
                File terpilih: <span class="font-semibold text-gray-700">{{ store.form.file.name }}</span>
              </p>
            </UFormField>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="p-3 bg-error-50 text-error-700 rounded-lg text-sm font-semibold">
              {{ store.errorMsg }}
            </div>

          </div>
        </div>
        <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse gap-3 rounded-b-lg">
          <UButton
            type="submit"
            :loading="store.loading"
            color="primary"
            variant="solid"
            size="md"
            class="w-full sm:w-auto font-bold"
          >
            {{ store.isEditing ? t('common.save') : t('common.submit') }}
          </UButton>
          <UButton
            type="button"
            color="neutral"
            variant="outline"
            size="md"
            class="w-full sm:w-auto mt-2 sm:mt-0 font-bold"
            @click="store.closeModal"
          >
            {{ t('common.cancel') }}
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