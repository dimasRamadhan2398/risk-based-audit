<template>
  <div class="p-6 max-w-full mx-auto space-y-6">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 ">{{ t('assignmentLetter.title') }}</h1>
      </div>
      <div class="flex items-center gap-2">
        <UButton
          :label="t('assignmentLetter.importDocument')"
          icon="i-lucide-upload"
          color="neutral"
          variant="outline"
          size="lg"
          class="font-bold shadow-md"
          to="/audit-assignment/upload"
        />
        <UButton
          :label="t('assignmentLetter.createAssignmentLetter')"
          icon="i-heroicons-plus"
          color="primary"
          size="lg"
          class="font-bold shadow-md"
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

const { t } = useI18n()
const store = useAssignmentLetterStore()
store.fetchAssignmentLetters()
</script>
