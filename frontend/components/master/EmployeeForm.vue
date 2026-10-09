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
              <UIcon name="i-heroicons-user-group" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? 'Edit Employee' : 'Add New Employee' }}
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
                Basic Information
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField label="Employee Code *" size="lg">
                  <UInput
                    v-model="store.form.employee_code"
                    type="text"
                    placeholder="e.g. EMP-001"
                    :disabled="store.isEditing"
                    required
                  />
                </UFormField>

                <UFormField label="Full Name *" size="lg">
                  <UInput
                    v-model="store.form.full_name"
                    type="text"
                    placeholder="Enter full name"
                    required
                  />
                </UFormField>

                <UFormField label="Email *" size="lg">
                  <UInput
                    v-model="store.form.email"
                    type="email"
                    placeholder="email@example.com"
                    required
                  />
                </UFormField>

                <UFormField label="Phone" size="lg">
                  <UInput
                    v-model="store.form.phone"
                    type="tel"
                    placeholder="+62 xxx xxxx xxxx"
                  />
                </UFormField>

                <UFormField label="Level Grade *" size="lg">
                  <UInput
                    v-model.number="store.form.level_grade"
                    type="number"
                    min="1"
                    required
                  />
                </UFormField>

                <UFormField label="Join Date *" size="lg">
                  <AppDatePicker
                    v-model="store.form.join_date"
                    required
                  />
                </UFormField>
              </div>

              <UFormField label="Status" size="lg">
                <div class="flex items-center gap-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="store.form.is_active"
                      @change="store.form.is_active = true"
                      class="w-4 h-4 text-primary-600"
                    />
                    <span>Active</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      :checked="!store.form.is_active"
                      @change="store.form.is_active = false"
                      class="w-4 h-4 text-error-600"
                    />
                    <span>Inactive</span>
                  </label>
                </div>
              </UFormField>
            </div>

            <!-- Organization -->
            <div class="space-y-4 pt-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                Organization
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField label="Company *" size="lg">
                  <USelectMenu
                    v-model="store.form.company_id"
                    :items="options.companyOptions"
                    value-key="value"
                    :loading="options.loading"
                    placeholder="Select company"
                    class="w-full"
                    required
                  />
                </UFormField>

                <UFormField label="Department *" size="lg">
                  <USelectMenu
                    v-model="store.form.department_id"
                    :items="options.departmentOptions"
                    value-key="value"
                    :loading="options.loading"
                    placeholder="Select department"
                    class="w-full"
                    required
                  />
                </UFormField>

                <UFormField label="Job Role *" size="lg">
                  <USelectMenu
                    v-model="store.form.job_role_id"
                    :items="options.jobRoleOptions"
                    value-key="value"
                    :loading="options.loading"
                    placeholder="Select job role"
                    class="w-full"
                    required
                  />
                </UFormField>

                <UFormField label="Work Location" size="lg">
                  <USelectMenu
                    v-model="store.form.work_location_id"
                    :items="workLocationItems"
                    value-key="value"
                    :loading="options.loading"
                    :clear="true"
                    placeholder="Select work location (optional)"
                    class="w-full"
                    @update:model-value="onWorkLocationChange"
                  />
                </UFormField>

                <UFormField label="Manager" size="lg">
                  <USelectMenu
                    v-model="store.form.manager_id"
                    :items="managerItems"
                    value-key="value"
                    :loading="options.loading"
                    :clear="true"
                    placeholder="Select manager (optional)"
                    class="w-full"
                    @update:model-value="onManagerChange"
                  />
                </UFormField>
              </div>
            </div>

            <!-- Address -->
            <div class="space-y-4 pt-4">
              <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b border-[var(--border-main)] pb-2">
                Address
              </h4>

              <UFormField label="Residence Address" size="lg">
                <UTextarea
                  v-model="store.form.residence_address"
                  placeholder="Street address"
                  :rows="2"
                  autoresize
                />
              </UFormField>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                <UFormField label="City" size="lg">
                  <UInput
                    v-model="store.form.residence_city"
                    type="text"
                    placeholder="City"
                  />
                </UFormField>

                <UFormField label="Province" size="lg">
                  <UInput
                    v-model="store.form.residence_province"
                    type="text"
                    placeholder="Province"
                  />
                </UFormField>

                <UFormField label="Postal Code" size="lg">
                  <UInput
                    v-model="store.form.residence_postal_code"
                    type="text"
                    placeholder="Postal code"
                  />
                </UFormField>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 bg-[var(--bg-surface)] border-t border-[var(--border-main)] rounded-b-xl flex justify-end gap-3">
            <UButton
              label="Cancel"
              color="neutral"
              variant="soft"
              type="button"
              @click="store.closeModal"
            />
            <UButton
              :label="store.isEditing ? 'Update' : 'Create'"
              color="primary"
              :loading="store.loading"
              type="submit"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useEmployeeStore } from '~/stores/employee'
import { useMasterOptionsStore } from '~/stores/master-options'

const store = useEmployeeStore()
const options = useMasterOptionsStore()

// Optional relations get an explicit empty choice so they can be cleared again.
const noneOption = { label: '— None —', value: '__none__' }

const onWorkLocationChange = (val: any) => {
  if (!val || val === '__none__') {
    store.form.work_location_id = ''
  }
}

const onManagerChange = (val: any) => {
  if (!val || val === '__none__') {
    store.form.manager_id = ''
  }
}

const workLocationItems = computed(() => [noneOption, ...options.locationOptions])

// An employee cannot be their own manager.
const managerItems = computed(() => [
  noneOption,
  ...options.employeeOptions.filter((o) => o.value !== store.editingId)
])

// Load the reference lists the first time the modal is opened.
watch(
  () => store.showModal,
  (open) => {
    if (open) options.fetchAll()
  },
  { immediate: true }
)
</script>
