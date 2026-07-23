<template>
  <UModal v-model:open="store.showModal" dismissible class="w-full sm:max-w-xl">
    <template #content>
      <UForm :state="store.form" @submit.prevent="store.handleSubmit">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">

          <!-- Header -->
          <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <div class="flex items-center gap-3">
              <UIcon name="i-heroicons-building-office" class="text-primary-500 text-2xl" />
              <h3 class="text-lg font-bold text-[var(--text-main)]">
                {{ store.isEditing ? 'Edit Department' : 'Add New Department' }}
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
                Department Information
              </h4>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <UFormField label="Department Code *" size="lg">
                  <UInput
                    v-model="store.form.department_code"
                    type="text"
                    placeholder="e.g. DEPT-001"
                    :disabled="store.isEditing"
                    required
                  />
                </UFormField>

                <UFormField label="Department Name *" size="lg">
                  <UInput
                    v-model="store.form.department_name"
                    type="text"
                    placeholder="Enter department name"
                    required
                  />
                </UFormField>

                <UFormField label="Level *" size="lg" class="md:col-span-2">
                  <UInput
                    v-model.number="store.form.level"
                    type="number"
                    min="1"
                    required
                  />
                  <p class="text-md text-gray-500 mt-1">Department hierarchy level</p>
                </UFormField>

                <UFormField label="Description" size="lg" class="md:col-span-2">
                  <UTextarea
                    v-model="store.form.department_description"
                    placeholder="Department description (optional)"
                    :rows="3"
                    autoresize
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
                Organization Link
              </h4>

              <div class="grid grid-cols-1 gap-4">
                <UFormField label="Company ID *" size="lg">
                  <UInput
                    v-model="store.form.company_id"
                    type="text"
                    placeholder="Company UUID"
                    required
                  />
                  <p class="text-md text-gray-500 mt-1">Enter Company UUID</p>
                </UFormField>

                <UFormField label="Person In Charge (PIC) ID *" size="lg">
                  <UInput
                    v-model="store.form.pic_id"
                    type="text"
                    placeholder="Employee UUID (PIC)"
                    required
                  />
                  <p class="text-md text-gray-500 mt-1">Enter Employee UUID who is the PIC</p>
                </UFormField>

                <UFormField label="Business Unit ID" size="lg">
                  <UInput
                    v-model="store.form.business_unit_id"
                    type="text"
                    placeholder="Business Unit UUID (optional)"
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
import { useDepartmentStore } from '~/stores/department'

const store = useDepartmentStore()
</script>
