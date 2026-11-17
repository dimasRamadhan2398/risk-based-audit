<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('masterData.employee.title') }}</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">{{ t('masterData.employee.subtitle') }}</p>
      </div>

      <!-- User seats (admin only): any active login with a role other than Viewer -->
      <div
        v-if="isAdmin && employeeStore.seats"
        class="flex items-center gap-2"
        :title="t('masterData.employee.seats.hint')"
      >
        <UBadge
          :color="seatStatus"
          variant="subtle"
          size="lg"
          icon="i-heroicons-user-group"
        >
          {{ t('masterData.employee.seats.indicator', { used: employeeStore.seats.used, limit: employeeStore.seats.limit }) }}
        </UBadge>
        <UBadge
          v-if="employeeStore.seats.over_limit"
          color="error"
          variant="solid"
          size="lg"
        >
          {{ t('masterData.employee.seats.overLimit') }}
        </UBadge>
      </div>
    </div>

    <!-- Employee Table -->
    <MasterEmployeeTable />

    <!-- Employee Form Modal -->
    <MasterEmployeeForm />

    <!-- One-time temporary password after creating an employee -->
    <MasterEmployeeAccountCredentialsModal />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '~/composables/useI18n'
import { useEmployeeStore } from '~/stores/employee'
import { getSeatStatus } from '~/utils/employeeSeats'
import MasterEmployeeTable from '~/components/master/EmployeeTable.vue'
import MasterEmployeeForm from '~/components/master/EmployeeForm.vue'

import MasterEmployeeAccountCredentialsModal from '~/components/master/EmployeeAccountCredentialsModal.vue'

const { t } = useI18n()
const { isAdmin } = useRbac()
const employeeStore = useEmployeeStore()

const seatStatus = computed(() => employeeStore.seats ? getSeatStatus(employeeStore.seats) : 'neutral')

// Seats are admin-only on the backend too; the store refreshes them after each change.
// Cleared for non-admins so a previous admin session's numbers never linger.
watch(isAdmin, (admin) => {
  if (admin) employeeStore.fetchSeats()
  else employeeStore.resetSeats()
}, { immediate: true })

definePageMeta({
  layout: 'default'
})
</script>
