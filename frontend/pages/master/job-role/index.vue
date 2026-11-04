<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('masterData.jobRole.title') }}
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          {{ t('masterData.jobRole.subtitle') }}
        </p>
      </div>
    </div>

    <!-- Job Role Table -->
    <MasterJobRoleTable />

    <!-- Job Role Form Modal -->
    <MasterJobRoleForm />
  </div>
</template>

<script setup lang="ts">
import MasterJobRoleTable from '~/components/master/JobRoleTable.vue'
import MasterJobRoleForm from '~/components/master/JobRoleForm.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'default'
})
</script>
