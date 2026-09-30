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
                {{ store.isEditing ? 'Edit Company' : 'Add New Company' }}
              </h3>
            </div>
            <button
              type="button"
              @click="store.closeModal"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1 rounded-lg focus:outline-none"
              aria-label="Close modal"
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
                Company Details
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField label="Company Code *" size="lg">
                  <UInput
                    v-model="store.form.company_code"
                    type="text"
                    placeholder="e.g. AIFL-HQ, SUB-001"
                    :disabled="store.isEditing"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">Unique company code identifier</p>
                </UFormField>

                <UFormField label="Company Name *" size="lg">
                  <UInput
                    v-model="store.form.company_name"
                    type="text"
                    placeholder="e.g. PT AIFL Indonesia"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">Common operational name</p>
                </UFormField>

                <UFormField label="Legal Name" size="lg" class="md:col-span-2">
                  <UInput
                    v-model="store.form.legal_name"
                    type="text"
                    placeholder="e.g. PT Astra International Financial Logistics Tbk"
                  />
                  <p class="text-xs text-gray-500 mt-1">Official registered entity name (optional)</p>
                </UFormField>

                <UFormField label="Company Type *" size="lg">
                  <USelectMenu
                    v-model="store.form.company_type"
                    :items="companyTypeOptions"
                    value-key="value"
                    placeholder="Select company type"
                    class="w-full"
                    required
                  />
                </UFormField>

                <UFormField label="Tax ID (NPWP)" size="lg">
                  <UInput
                    v-model="store.form.tax_id"
                    type="text"
                    placeholder="e.g. 01.234.567.8-901.000"
                  />
                </UFormField>

                <UFormField label="Parent Company (Optional)" size="lg" class="md:col-span-2">
                  <USelectMenu
                    v-model="store.form.parent_id"
                    :items="parentCompanyOptions"
                    value-key="value"
                    :clear="true"
                    placeholder="Select parent company (if subsidiary / branch)"
                    class="w-full"
                    @update:model-value="onParentCompanyChange"
                  />
                  <p class="text-xs text-gray-500 mt-1">Select holding company if this entity is a subsidiary or branch</p>
                </UFormField>
              </div>
            </div>

            <!-- Contact Information -->
            <div class="space-y-4 pt-2">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                Contact & Online Presence
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                <UFormField label="Phone" size="lg">
                  <UInput
                    v-model="store.form.phone"
                    type="tel"
                    placeholder="+62 21 ..."
                  />
                </UFormField>

                <UFormField label="Email" size="lg">
                  <UInput
                    v-model="store.form.email"
                    type="email"
                    placeholder="corporate@aifl.co.id"
                  />
                </UFormField>

                <UFormField label="Website" size="lg">
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
              <UFormField label="Status" size="lg">
                <div class="flex items-center gap-6 mt-1">
                  <label class="flex items-center gap-2 cursor-pointer text-sm font-medium">
                    <input
                      type="radio"
                      :checked="store.form.is_active"
                      @change="store.form.is_active = true"
                      class="w-4 h-4 text-primary-600 focus:ring-primary-500"
                    />
                    <span>Active</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer text-sm font-medium">
                    <input
                      type="radio"
                      :checked="!store.form.is_active"
                      @change="store.form.is_active = false"
                      class="w-4 h-4 text-error-600 focus:ring-error-500"
                    />
                    <span>Inactive</span>
                  </label>
                </div>
              </UFormField>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-4 sm:px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-3">
            <UButton
              label="Cancel"
              color="neutral"
              variant="soft"
              class="w-full sm:w-auto"
              type="button"
              @click="store.closeModal"
            />
            <UButton
              :label="store.isEditing ? 'Update Company' : 'Create Company'"
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

const companyTypeOptions = [
  { label: 'Holding Company', value: 'HOLDING' },
  { label: 'Subsidiary Entity', value: 'SUBSIDIARY' },
  { label: 'Branch Office', value: 'BRANCH' }
]

const noneOption = { label: '— None (Top Level Holding) —', value: '__none__' }

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
    noneOption,
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
