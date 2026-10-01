<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 dark:border-gray-800 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('assignmentLetter.title') }}</h1>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full sm:w-auto">
        <UButton
          v-if="canImportPlanDocs"
          :label="t('assignmentLetter.importDocument')"
          icon="i-lucide-upload"
          color="neutral"
          variant="outline"
          size="lg"
          class="font-bold shadow-md w-full sm:w-auto"
          to="/assignment-letter/upload"
        />
        <UButton
          v-if="canManageAssignmentLetter"
          :label="t('assignmentLetter.createAssignmentLetter')"
          icon="i-heroicons-plus"
          color="primary"
          size="lg"
          class="font-bold shadow-md w-full sm:w-auto"
          @click="store.openModal"
        />
      </div>
    </div>

    <AssignmentLetterTable />

    <AssignmentLetterForm />

  </div>
</template>

<script setup lang="ts">
import AssignmentLetterForm from '~/components/assignment-letter/AssignmentLetterForm.vue';
import AssignmentLetterTable from '~/components/assignment-letter/AssignmentLetterTable.vue';
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'

const { t } = useI18n()
const { canManageAssignmentLetter, canImportPlanDocs } = useRbac()
const store = useAssignmentLetterStore()
store.fetchAssignmentLetters()

</script>
