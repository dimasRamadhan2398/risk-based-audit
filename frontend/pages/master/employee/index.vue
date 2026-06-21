<template>
  <div class="max-w-full mx-auto space-y-6">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Employee Management</h1>
        <p class="text-gray-500 mt-1">Manage employee data for the organization</p>
      </div>
    </div>

    <!-- Employee Table -->
    <MasterEmployeeTable />

    <!-- Employee Form Modal -->
    <MasterEmployeeForm />
  </div>
</template>

<script setup lang="ts">
import MasterEmployeeTable from '~/components/master/EmployeeTable.vue'
import MasterEmployeeForm from '~/components/master/EmployeeForm.vue'

definePageMeta({
  layout: 'default'
})
</script>
