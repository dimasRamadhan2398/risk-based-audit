<template>
  <UModal
    v-model:open="store.showModal"
    dismissible
    @update:open="(val: boolean) => { if (!val) store.closeModal() }"
    :ui="{ content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden' }"
  >
    <template #content>
      <UForm :state="store.form" @submit.prevent="store.handleSubmit">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">

          <!-- Header -->
          <div class="px-4 sm:px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-building-office" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? t('masterData.department.editTitle') : t('masterData.department.createTitle') }}
              </h3>
            </div>
            <button
              type="button"
              @click="store.closeModal"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1 rounded-lg focus:outline-none"
              :aria-label="t('common.close')"
            >
              <UIcon name="i-heroicons-x-mark" class="text-2xl" />
            </button>
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

            <!-- Basic Info -->
            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.department.sectionInfo') }}
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField :label="t('masterData.department.fieldCode')" size="lg">
                  <UInput
                    v-model="store.form.department_code"
                    type="text"
                    :placeholder="t('masterData.department.fieldCodePlaceholder')"
                    :disabled="store.isEditing"
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.department.fieldName')" size="lg">
                  <UInput
                    v-model="store.form.department_name"
                    type="text"
                    :placeholder="t('masterData.department.fieldNamePlaceholder')"
                    required
                  />
                </UFormField>

                <UFormField :label="t('masterData.department.fieldLevel')" size="lg" class="md:col-span-2">
                  <UInput
                    v-model.number="store.form.level"
                    type="number"
                    min="1"
                    required
                  />
                  <p class="text-md text-gray-500 mt-1">{{ t('masterData.department.fieldLevelHelp') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.department.fieldDescription')" size="lg" class="md:col-span-2">
                  <UTextarea
                    v-model="store.form.department_description"
                    :placeholder="t('masterData.department.fieldDescriptionPlaceholder')"
                    :rows="3"
                    autoresize
                  />
                </UFormField>
              </div>

              <UFormField :label="t('masterData.department.status')" size="lg">
                <div class="flex items-center gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="store.form.is_active"
                      @change="store.form.is_active = true"
                      class="w-4 h-4 text-primary-600"
                    />
                    <span>{{ t('masterData.department.active') }}</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="!store.form.is_active"
                      @change="store.form.is_active = false"
                      class="w-4 h-4 text-error-600"
                    />
                    <span>{{ t('masterData.department.inactive') }}</span>
                  </label>
                </div>
              </UFormField>
            </div>

            <!-- Organization -->
            <div class="space-y-4 pt-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('masterData.department.organizationLink') }}
              </h4>

              <div class="grid grid-cols-1 gap-4">
                <UFormField :label="t('masterData.department.companyId')" size="lg">
                  <USelectMenu
                    v-model="store.form.company_id"
                    :items="options.companyOptions"
                    value-key="value"
                    :loading="options.loading"
                    :placeholder="t('masterData.department.companyIdPlaceholder')"
                    class="w-full"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.department.companyIdHelp') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.department.picId')" size="lg">
                  <USelectMenu
                    v-model="store.form.pic_id"
                    :items="picItems"
                    value-key="value"
                    :loading="options.loading"
                    :clear="true"
                    :placeholder="t('masterData.department.picIdPlaceholder')"
                    class="w-full"
                    @update:model-value="onPicChange"
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.department.picIdHelp') }}</p>
                </UFormField>

                <UFormField :label="t('masterData.department.businessUnitId')" size="lg">
                  <USelectMenu
                    v-model="store.form.business_unit_id"
                    :items="businessUnitItems"
                    value-key="value"
                    :loading="options.loading"
                    :clear="true"
                    :placeholder="t('masterData.department.businessUnitIdPlaceholder')"
                    class="w-full"
                    @update:model-value="onBusinessUnitChange"
                  />
                  <p class="text-xs text-gray-500 mt-1">{{ t('masterData.department.businessUnitIdHelp') }}</p>
                </UFormField>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-4 sm:px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-3">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="soft"
              class="w-full sm:w-auto"
              type="button"
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
import { useDepartmentStore } from '~/stores/department'
import { useMasterOptionsStore } from '~/stores/master-options'

const store = useDepartmentStore()
const options = useMasterOptionsStore()
const { t } = useI18n()

const noneOption = { label: '— None —', value: '__none__' }

const onPicChange = (val: any) => {
  if (!val || val === '__none__') {
    store.form.pic_id = ''
  }
}

const onBusinessUnitChange = (val: any) => {
  if (!val || val === '__none__') {
    store.form.business_unit_id = ''
  }
}

const businessUnitItems = computed(() => {
  let list = options.businessUnits
  if (store.form.company_id) {
    list = list.filter((bu: any) => !bu.company_id || bu.company_id === store.form.company_id)
  }
  return [
    noneOption,
    ...list.map((b: any) => ({
      label: b.business_unit_code ? `${b.business_unit_name} (${b.business_unit_code})` : b.business_unit_name,
      value: b.id
    }))
  ]
})

const picItems = computed(() => {
  let list = options.employees
  if (store.form.company_id) {
    list = list.filter((emp: any) => !emp.company_id || emp.company_id === store.form.company_id)
  }
  return [
    noneOption,
    ...list.map((e: any) => ({
      label: e.employee_code ? `${e.full_name} (${e.employee_code})` : e.full_name,
      value: e.id
    }))
  ]
})

watch(
  () => store.showModal,
  (open) => {
    if (open) options.fetchAll()
  },
  { immediate: true }
)
</script>
