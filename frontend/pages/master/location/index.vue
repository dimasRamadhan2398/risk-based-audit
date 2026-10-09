<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('masterData.location.title') }}
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          {{ t('masterData.location.subtitle') }}
        </p>
      </div>
    </div>

    <!-- Location Table -->
    <MasterLocationTable />

    <!-- Location Form Modal -->
    <MasterLocationForm />
  </div>
</template>

<script setup lang="ts">
import MasterLocationTable from '~/components/master/LocationTable.vue'
import MasterLocationForm from '~/components/master/LocationForm.vue'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()

definePageMeta({
  layout: 'default',
  middleware: 'auth'
})
</script>
