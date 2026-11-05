<template>
  <div class="space-y-4">
    <div>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">{{ t('masterData.employee.listTitle') }}</h2>
        <UButton v-if="canWriteMasterData"
          :label="t('masterData.employee.add')"
          icon="i-heroicons-plus"
          color="primary"
          class="w-full sm:w-auto justify-center text-center"
          @click="store.openCreateModal"
        />
      </div>
    </div>

    <!-- Search & Filter -->
    <div class="mb-4 flex flex-wrap gap-2 sm:gap-4 items-center w-full">
      <CommonSearchInput
        v-model="searchInput"
        :placeholder="t('masterData.employee.searchPlaceholder')"
        class="w-full sm:w-64"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      />
    </div>

    <!-- Loading State -->
    <div
      v-if="store.loading"
      class="flex justify-center py-8"
    >
      <ULoadingIcon />
    </div>

    <!-- Error State -->
    <div
      v-else-if="store.errorMsg"
      class="py-4"
    >
      <UAlert
        color="error"
        variant="soft"
        :title="store.errorMsg"
        icon="i-heroicons-exclamation-circle"
      />
    </div>

    <!-- Empty State -->
    <div
      v-else-if="store.employees.length === 0"
      class="py-8 text-center"
    >
      <UIcon
        name="i-heroicons-user-group"
        class="text-4xl text-gray-400 mb-2"
      />
      <p class="text-gray-500">
        {{ t('masterData.employee.empty') }}
      </p>
    </div>

    <!-- Table & Pagination via TableEntities -->
    <div v-else class="space-y-4">
      <!-- Desktop View (>= md): Full Table Entities -->
      <div class="hidden md:block">
        <TableEntities
          :data="store.employees"
          :columns="store.columns"
          :loading="store.loading"
          :server-side="true"
          :total="store.pagination.total"
          :items-per-page="store.pagination.page_size"
          :page="store.pagination.page"
          @update:page="(p) => store.fetchEmployees(p)"
          @update:items-per-page="(size) => store.setPageSize(size)"
        >
          <template #employee_code-cell="{ row }">
            <span class="font-medium text-primary-600">{{ row.original.employee_code }}</span>
          </template>

          <template #full_name-cell="{ row }">
            <div class="font-medium">{{ row.original.full_name }}</div>
          </template>

          <template #email-cell="{ row }">
            <a :href="`mailto:${row.original.email}`" class="text-blue-600 hover:underline">
              {{ row.original.email }}
            </a>
          </template>

          <template #level_grade-cell="{ row }">
            <UBadge color="neutral" variant="soft">
              {{ t('masterData.employee.level', { level: row.original.level_grade }) }}
            </UBadge>
          </template>

          <template #is_active-cell="{ row }">
            <UBadge
              :color="row.original.is_active ? 'secondary' : 'error'"
              variant="subtle"
            >
              {{ row.original.is_active ? t('masterData.employee.active') : t('masterData.employee.inactive') }}
            </UBadge>
          </template>

          <template #actions-cell="{ row }">
            <div class="flex gap-1">
              <UButton
                v-if="canWriteMasterData"
                icon="i-heroicons-pencil"
                color="primary"
                variant="ghost"
                size="sm"
                @click="store.handleEdit(row.original)"
              />
              <UButton
                v-if="isAdmin"
                icon="i-heroicons-key"
                color="warning"
                variant="ghost"
                size="sm"
                :title="t('masterData.employee.resetPassword')"
                @click="openResetPassword(row.original)"
              />
              <UButton
                v-if="canWriteMasterData"
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="sm"
                @click="store.handleDelete(row.original)"
              />
            </div>
          </template>
        </TableEntities>
      </div>

      <!-- Mobile & Tablet View (< md): Stacked Cards List -->
      <div class="block md:hidden space-y-3">
        <div
          v-for="(emp, index) in store.employees"
          :key="emp.id || index"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
        >
          <!-- Card Header: Code Badge + Active & Level Grade Badges -->
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40">
                <UIcon name="i-heroicons-user" class="size-3" />
                {{ emp.employee_code }}
              </span>
              <UBadge
                :color="emp.is_active ? 'secondary' : 'error'"
                variant="subtle"
                size="sm"
              >
                {{ emp.is_active ? t('masterData.employee.active') : t('masterData.employee.inactive') }}
              </UBadge>
            </div>
            <UBadge color="neutral" variant="soft" size="sm">
              {{ t('masterData.employee.level', { level: emp.level_grade }) }}
            </UBadge>
          </div>

          <!-- Card Body: Full Name & Email -->
          <div>
            <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
              {{ emp.full_name }}
            </h4>
            <div v-if="emp.email" class="flex items-center gap-1.5 text-xs text-primary-600 dark:text-primary-400 mt-1">
              <UIcon name="i-heroicons-envelope" class="size-3.5 shrink-0" />
              <a :href="`mailto:${emp.email}`" class="truncate hover:underline">
                {{ emp.email }}
              </a>
            </div>
          </div>

          <!-- Card Footer: Index & Action Buttons -->
          <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs text-slate-400 font-mono">
              #{{ (store.pagination.page - 1) * store.pagination.page_size + index + 1 }}
            </span>
            <div class="flex items-center gap-1">
              <UButton
                v-if="canWriteMasterData"
                icon="i-heroicons-pencil"
                color="primary"
                variant="ghost"
                size="md"
                @click="store.handleEdit(emp)"
              />
              <UButton
                v-if="isAdmin"
                icon="i-heroicons-key"
                color="warning"
                variant="ghost"
                size="md"
                :title="t('masterData.employee.resetPassword')"
                @click="openResetPassword(emp)"
              />
              <UButton
                v-if="canWriteMasterData"
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="md"
                @click="store.handleDelete(emp)"
              />
            </div>
          </div>
        </div>

        <!-- Mobile Pagination Controls -->
        <div v-if="store.pagination.total > store.pagination.page_size" class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800">
          <span class="text-xs text-slate-500 dark:text-slate-400">
            {{ (store.pagination.page - 1) * store.pagination.page_size + 1 }}-{{ Math.min(store.pagination.page * store.pagination.page_size, store.pagination.total) }} of {{ store.pagination.total }}
          </span>
          <UPagination
            :page="store.pagination.page"
            :items-per-page="store.pagination.page_size"
            :total="store.pagination.total"
            size="sm"
            active-color="primary"
            color="neutral"
            variant="outline"
            class="max-w-full overflow-x-auto"
            @update:page="(p) => store.fetchEmployees(p)"
          />
        </div>
      </div>
    </div>

    <EmployeeResetPasswordModal
      v-if="isAdmin"
      v-model:open="showResetPassword"
      :employee="resetPasswordEmployee"
    />
  </div>
</template>

<script setup lang="ts">
import { useEmployeeStore } from '~/stores/employee'
import EmployeeResetPasswordModal from '~/components/master/EmployeeResetPasswordModal.vue'
import type { Employee } from '~/types/master'
import { useI18n } from '~/composables/useI18n'

const store = useEmployeeStore()
const { isAdmin, canWriteMasterData } = useRbac()
const { t } = useI18n()

// Admin-only: reset the password of the login account linked to an employee
const showResetPassword = ref(false)
const resetPasswordEmployee = ref<Employee | null>(null)

const openResetPassword = (employee: Employee) => {
  resetPasswordEmployee.value = employee
  showResetPassword.value = true
}

// Local search state
const searchInput = ref(store.search)
const pageSize = ref(store.pagination.page_size)

const handleSearch = () => {
  store.setSearch(searchInput.value)
}

const resetSearch = () => {
  searchInput.value = ''
  store.setSearch('')
}

const handlePageSizeChange = (value: any) => {
  store.setPageSize(Number(value))
}

// Fetch data on mount
onMounted(() => {
  store.fetchEmployees()
})
</script>
