<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-2">Audit Fieldwork</h1>
    <p class="text-gray-500 mb-6">Manage audit execution from interview to working papers</p>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Select Assignment Letter (Audit)
          </label>
          <USelectMenu
            v-model="store.selectedAssignmentLetter"
            :items="store.publishedAssignmentLetters"
            placeholder="Select Assignment Letter to audit"
            class="w-full md:w-96"
            :disabled="store.publishedAssignmentLetters.length === 0"
          >
            <template #leading>
              <UIcon name="i-heroicons-document-text" class="size-5" />
            </template>
          </USelectMenu>
        </div>
        <div v-if="store.publishedAssignmentLetters.length === 0" class="text-sm text-amber-600">
          <UIcon name="i-heroicons-exclamation-triangle" class="size-4 inline mr-1" />
          No Assignment Letter with Published status. Please create and publish an Assignment Letter first.
        </div>
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline mr-1" />
          Audit: {{ store.selectedAssignmentLetter }}
        </div>
      </div>
    </UCard>

    <!-- Main Content - Only show when Assignment Letter is selected -->
    <div v-if="store.hasSelectedAssignmentLetter">
      <UTabs :items="store.tabs" class="w-full">
        <template #tab01>
          <AuditFieldworkInterview />
        </template>

        <template #tab02>
          <AuditFieldworkObservation />
        </template>

        <template #tab03>
          <AuditFieldworkDocument />
        </template>

        <template #tab04>
          <AuditFieldworkSample />
        </template>

        <template #tab05>
          <AuditFieldworkTestControls />
        </template>

        <template #tab06>
          <div class="text-center py-8">
            <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
            <h3 class="text-lg font-semibold text-gray-700">Working Papers</h3>
            <p class="text-gray-500 mt-2">Manage working papers through separate menu</p>
            <UButton color="primary" variant="soft" class="mt-4" label="Open Working Papers" to="/working-paper" />
          </div>
        </template>
      </UTabs>
    </div>

    <!-- Empty State - When no Assignment Letter selected -->
    <div v-else class="text-center py-16">
      <UIcon name="i-heroicons-folder-open" class="size-20 text-gray-300 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-700">Select Assignment Letter</h3>
      <p class="text-gray-500 mt-2 max-w-md mx-auto">
        Please select an Assignment Letter (Audit) first to start managing fieldwork.
        Only Assignment Letters with "Published" status can be selected.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

const store = useAuditFieldworkStore()
</script>
