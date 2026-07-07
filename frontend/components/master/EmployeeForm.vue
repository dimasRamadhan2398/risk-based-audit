<template>
  <UModal v-model:open="store.showModal" dismissible class="w-full sm:max-w-2xl">
    <template #content>
      <UForm :state="store.form" @submit.prevent="store.handleSubmit">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">

          <!-- Header -->
          <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-user-group" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? 'Edit Employee' : 'Add New Employee' }}
              </h3>
            </div>
            <UIcon
              name="i-heroicons-x-mark"
              @click="store.closeModal"
              class="text-primary-400 hover:text-primary-600 text-2xl cursor-pointer"
            />
          </div>

          <!-- Body -->
          <div class="p-6 overflow-y-auto space-y-4">
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
                  <UInput
                    v-model="store.form.join_date"
                    type="date"
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
                <UFormField label="Company ID *" size="lg">
                  <UInput
                    v-model="store.form.company_id"
                    type="text"
                    placeholder="Company UUID"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">Enter Company UUID</p>
                </UFormField>

                <UFormField label="Department ID *" size="lg">
                  <UInput
                    v-model="store.form.department_id"
                    type="text"
                    placeholder="Department UUID"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">Enter Department UUID</p>
                </UFormField>

                <UFormField label="Job Role ID *" size="lg">
                  <UInput
                    v-model="store.form.job_role_id"
                    type="text"
                    placeholder="Job Role UUID"
                    required
                  />
                  <p class="text-xs text-gray-500 mt-1">Enter Job Role UUID</p>
                </UFormField>

                <UFormField label="Work Location ID" size="lg">
                  <UInput
                    v-model="store.form.work_location_id"
                    type="text"
                    placeholder="Location UUID (optional)"
                  />
                </UFormField>

                <UFormField label="Manager ID" size="lg">
                  <UInput
                    v-model="store.form.manager_id"
                    type="text"
                    placeholder="Manager UUID (optional)"
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
          <div class="px-6 py-4 bg-gray-50 border-t border-gray-200 rounded-b-xl flex justify-end gap-3">
            <UButton
              label="Cancel"
              color="neutral"
              variant="soft"
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

const store = useEmployeeStore()
</script>
