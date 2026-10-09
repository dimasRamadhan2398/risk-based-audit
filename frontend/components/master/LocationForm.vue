<template>
  <UModal v-model:open="store.showModal" dismissible :ui="{ content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden' }">
    <template #content>
      <UForm :state="store.form" @submit.prevent="handleSubmit">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">

          <!-- Header -->
          <div class="px-4 sm:px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-map-pin" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('masterData.location.editTitle') : t('masterData.location.createTitle') }}
              </h3>
            </div>
            <UIcon
              name="i-heroicons-x-mark"
              class="text-primary-400 hover:text-primary-600 text-2xl cursor-pointer"
              @click="store.closeModal"
            />
          </div>

          <!-- Body -->
          <div class="p-4 sm:p-6 overflow-y-auto space-y-4">
            <!-- Error Message -->
            <UAlert
              v-if="store.errorMsg"
              color="error"
              variant="soft"
              :title="store.errorMsg"
              icon="i-heroicons-exclamation-circle"
              class="mb-4"
            />

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.location.sectionInfo') }}
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField :label="t('masterData.location.fieldName')" size="lg" class="md:col-span-2">
                  <UInput
                    v-model="store.form.name"
                    type="text"
                    :placeholder="t('masterData.location.fieldNamePlaceholder')"
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.location.fieldAddress')" size="lg" class="md:col-span-2">
                  <UTextarea
                    v-model="store.form.address"
                    :placeholder="t('masterData.location.fieldAddressPlaceholder')"
                    :rows="2"
                    autoresize
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.location.fieldCity')" size="lg">
                  <UInput
                    v-model="store.form.city"
                    type="text"
                    :placeholder="t('masterData.location.fieldCityPlaceholder')"
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.location.fieldProvince')" size="lg">
                  <UInput
                    v-model="store.form.province"
                    type="text"
                    :placeholder="t('masterData.location.fieldProvincePlaceholder')"
                  />
                </UFormField>

                <UFormField :label="t('masterData.location.fieldPostalCode')" size="lg">
                  <UInput v-model="store.form.postal_code" type="text" placeholder="60234" />
                </UFormField>

                <UFormField :label="t('masterData.location.fieldCountry')" size="lg">
                  <UInput v-model="store.form.country" type="text" placeholder="Indonesia" />
                </UFormField>
              </div>

              <UFormField :label="t('masterData.location.status')" size="lg">
                <div class="flex items-center gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="store.form.is_active"
                      class="w-4 h-4 text-primary-600"
                      @change="store.form.is_active = true"
                    />
                    <span>{{ t('masterData.location.active') }}</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="!store.form.is_active"
                      class="w-4 h-4 text-error-600"
                      @change="store.form.is_active = false"
                    />
                    <span>{{ t('masterData.location.inactive') }}</span>
                  </label>
                </div>
              </UFormField>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-4 sm:px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-3">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="soft"
              class="w-full sm:w-auto"
              @click="store.closeModal"
            />
            <UButton
              :label="store.isEditing ? t('common.update') : t('common.create')"
              color="primary"
              :loading="store.loading"
              class="w-full sm:w-auto font-bold"
              type="submit"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useLocationStore } from '~/stores/location'
import { useI18n } from '~/composables/useI18n'

const store = useLocationStore()
const toast = useAppToast()
const { t } = useI18n()

const handleSubmit = async () => {
  const wasEditing = store.isEditing
  const success = await store.handleSubmit()
  if (success) {
    toast.success(
      wasEditing ? t('masterData.location.updated') : t('masterData.location.created')
    )
  }
}
</script>
