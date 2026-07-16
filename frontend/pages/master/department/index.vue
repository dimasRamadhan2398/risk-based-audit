<template>
  <div class="max-w-full mx-auto space-y-6">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Department Management</h1>
        <p class="text-gray-500 mt-1">Manage department data for the organization</p>
      </div>
    </div>

    <!-- Department Table -->
    <MasterDepartmentTable />

    <!-- Department Form Modal -->
    <MasterDepartmentForm />
  </div>
</template>

<script setup lang="ts">
import MasterDepartmentTable from '~/components/master/DepartmentTable.vue'
import MasterDepartmentForm from '~/components/master/DepartmentForm.vue'

definePageMeta({
  layout: 'default'
})
</script>
