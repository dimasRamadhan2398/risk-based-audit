<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 ">Assignment Letter List</h1>
      </div>
      <UButton
        label="Create Assignment Letter"
        icon="i-heroicons-plus"
        color="primary"
        size="lg"
        class="font-bold shadow-md"
        @click="store.openModal"
      />
    </div>

    <AssignmentLetterTable />

    <AssignmentLetterForm />

  </div>
</template>

<script setup lang="ts">
import AssignmentLetterForm from '~/components/assignment-letter/AssignmentLetterForm.vue';
import AssignmentLetterTable from '~/components/assignment-letter/AssignmentLetterTable.vue';
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const store = useAssignmentLetterStore()

</script>
