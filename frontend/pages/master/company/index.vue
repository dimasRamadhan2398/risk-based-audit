<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Company Management</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">Manage holding, subsidiary entities, and branch companies</p>
      </div>
    </div>

    <!-- Company Table -->
    <MasterCompanyTable />

    <!-- Company Form Modal -->
    <MasterCompanyForm />
  </div>
</template>

<script setup lang="ts">
import MasterCompanyTable from '~/components/master/CompanyTable.vue'
import MasterCompanyForm from '~/components/master/CompanyForm.vue'

definePageMeta({
  layout: 'default'
})
</script>
