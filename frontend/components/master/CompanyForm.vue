<template>
  <UModal
    v-model:open="store.showModal"
    dismissible
    @update:open="(val: boolean) => { if (!val) store.closeModal() }"
    :ui="{ content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-2xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden' }"
  >
    <template #content>
      <UForm :state="store.form" @submit.prevent="store.handleSubmit">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">

          <!-- Header -->
          <div class="px-4 sm:px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-building-office-2" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('masterData.company.edit') : t('masterData.company.addNew') }}
              </h3>
            </div>
            <button
              type="button"
              @click="store.closeModal"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1 rounded-lg focus:outline-none"
              :aria-label="t('masterData.company.close')"
            >
              <UIcon name="i-heroicons-x-mark" class="text-2xl" />
            </button>
          </div>

          <!-- Body -->
          <div class="p-4 sm:p-6 overflow-y-auto space-y-5">
            <!-- Error Message -->
            <UAlert
              v-if="store.errorMsg"
              color="error"
              variant="soft"
              :title="store.errorMsg"
              icon="i-heroicons-exclamation-circle"
              class="mb-4"
            />

            <!-- Basic Info -->
            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.company.details') }}
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField :label="t('masterData.company.code')" size="lg">
                  <UInput
                    v-model="store.form.company_code"
                    type="text"
                    :placeholder="t('masterData.company.codePlaceholder')"
                    :disabled="store.isEditing"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.company.codeHint') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.company.name')" size="lg">
                  <UInput
                    v-model="store.form.company_name"
                    type="text"
                    :placeholder="t('masterData.company.namePlaceholder')"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.company.nameHint') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.company.legalName')" size="lg" class="md:col-span-2">
                  <UInput
                    v-model="store.form.legal_name"
                    type="text"
                    :placeholder="t('masterData.company.legalNamePlaceholder')"
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.company.legalNameHint') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.company.type')" size="lg">
                  <USelectMenu
                    v-model="store.form.company_type"
                    :items="companyTypeOptions"
                    value-key="value"
                    :placeholder="t('masterData.company.typePlaceholder')"
                    class="w-full"
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.company.taxId')" size="lg">
                  <UInput
                    v-model="store.form.tax_id"
                    type="text"
                    :placeholder="t('masterData.company.taxIdPlaceholder')"
                  />
                </UFormField>

                <UFormField :label="t('masterData.company.parent')" size="lg" class="md:col-span-2">
                  <USelectMenu
                    v-model="store.form.parent_id"
                    :items="parentCompanyOptions"
                    value-key="value"
                    :clear="true"
                    :placeholder="t('masterData.company.parentPlaceholder')"
                    class="w-full"
                    @update:model-value="onParentCompanyChange"
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.company.parentHint') }}</p>
                </UFormField>
              </div>
            </div>

            <!-- Contact Information -->
            <div class="space-y-4 pt-2">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.company.contact') }}
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                <UFormField :label="t('masterData.company.phone')" size="lg">
                  <UInput
                    v-model="store.form.phone"
                    type="tel"
                    placeholder="+62 21 ..."
                  />
                </UFormField>

                <UFormField :label="t('masterData.company.email')" size="lg">
                  <UInput
                    v-model="store.form.email"
                    type="email"
                    placeholder="corporate@aifl.co.id"
                  />
                </UFormField>

                <UFormField :label="t('masterData.company.website')" size="lg">
                  <UInput
                    v-model="store.form.website"
                    type="url"
                    placeholder="https://www.aifl.co.id"
                  />
                </UFormField>
              </div>
            </div>

            <!-- Status -->
            <div class="pt-2">
              <UFormField :label="t('masterData.company.status')" size="lg">
                <div class="flex items-center gap-6 mt-1">
                  <label class="flex items-center gap-2 cursor-pointer text-sm font-medium">
                    <input
                      type="radio"
                      :checked="store.form.is_active"
                      @change="store.form.is_active = true"
                      class="w-4 h-4 text-primary-600 focus:ring-primary-500"
                    />
                    <span>{{ t('masterData.company.active') }}</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer text-sm font-medium">
                    <input
                      type="radio"
                      :checked="!store.form.is_active"
                      @change="store.form.is_active = false"
                      class="w-4 h-4 text-error-600 focus:ring-error-500"
                    />
                    <span>{{ t('masterData.company.inactive') }}</span>
                  </label>
                </div>
              </UFormField>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-4 sm:px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-3">
            <UButton
              :label="t('masterData.company.cancel')"
              color="neutral"
              variant="soft"
              class="w-full sm:w-auto"
              type="button"
              @click="store.closeModal"
            />
            <UButton
              :label="store.isEditing ? t('masterData.company.update') : t('masterData.company.create')"
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
import { useCompanyStore } from '~/stores/company'
import { useMasterOptionsStore } from '~/stores/master-options'

const store = useCompanyStore()
const masterOptions = useMasterOptionsStore()
const { t } = useI18n()

const companyTypeOptions = computed(() => [
  { label: t('masterData.company.types.HOLDING'), value: 'HOLDING' },
  { label: t('masterData.company.types.SUBSIDIARY'), value: 'SUBSIDIARY' },
  { label: t('masterData.company.types.BRANCH'), value: 'BRANCH' }
])

const noneOption = computed(() => ({ label: t('masterData.company.none'), value: '__none__' }))

const onParentCompanyChange = (val: any) => {
  if (!val || val === '__none__') {
    store.form.parent_id = ''
  }
}

const parentCompanyOptions = computed(() => {
  const currentId = store.editingId
  // Exclude current company from being its own parent
  const list = masterOptions.companies.filter((c: any) => c.id !== currentId)
  return [
    noneOption.value,
    ...list.map((c: any) => ({
      label: c.company_name || c.name || c.company_code || c.code,
      value: c.id
    }))
  ]
})

watch(
  () => store.showModal,
  (open) => {
    if (open) {
      masterOptions.fetchAll()
    }
  },
  { immediate: true }
)
</script>
