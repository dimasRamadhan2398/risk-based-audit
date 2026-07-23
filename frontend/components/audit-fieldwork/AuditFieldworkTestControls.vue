<template>
  <div class="space-y-4">
    <!-- Header with Statistics and Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">Test Controls</h3>
        <p class="text-sm text-gray-500">Test efficiency and effectiveness of existing controls</p>
      </div>
      <div class="flex gap-2">
        <div v-if="store.testControls.length > 0" class="flex gap-2 mr-2">
          <UBadge color="success" variant="solid">{{ store.effectiveControls }} Effective</UBadge>
          <UBadge color="error" variant="solid">{{ store.ineffectiveControls }} Ineffective</UBadge>
        </div>
        <UButton color="primary" icon="i-heroicons-plus" label="Test Control" @click="store.openTestControlModal()" />
      </div>
    </div>

    <!-- Test Controls List -->
    <UCard v-if="store.testControls.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.testControls" :columns="columns">
        <template #controlName-cell="{ row }">
          <span class="font-medium">{{ row.original.controlName }}</span>
        </template>
        <template #controlType-cell="{ row }">
          <UBadge :color="getControlTypeColor(row.original.controlType)" variant="subtle">{{ row.original.controlType }}</UBadge>
        </template>
        <template #testResult-cell="{ row }">
          <UBadge :color="getResultColor(row.original.testResult)" variant="solid">{{ row.original.testResult }}</UBadge>
        </template>
        <template #finding-cell="{ row }">
          <span class="text-sm text-gray-600 line-clamp-2">{{ row.original.finding || '-' }}</span>
        </template>
        <template #mitigationPlan-cell="{ row }">
          <span class="text-sm text-gray-600 line-clamp-2">{{ row.original.mitigationPlan || '-' }}</span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton icon="i-heroicons-eye" color="neutral" variant="ghost" size="sm" @click="store.viewTestControl(row.original)" />
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editTestControl(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteTestControl(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-shield-check" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">No test control data yet</p>
      <UButton color="primary" variant="soft" class="mt-2" label="Add Test" @click="store.openTestControlModal()" />
    </div>

    <!-- Test Control Modal -->
    <Teleport to="body">
      <div v-if="store.showTestControlModal" class="fixed inset-0   -gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">
                {{ store.isReadOnlyTestControl ? 'View Test Control' : (store.isEditingTestControl ? 'Edit Test Control' : 'Test Control Form') }}
              </h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showTestControlModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
            <!-- Control Information -->
            <div class="  -gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Control Information</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Control Name" required>
                  <UInput v-model="store.testControlForm.controlName" placeholder="Example: Purchase Order Approval" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
                <UFormField label="Control Type" required>
                  <USelectMenu v-model="store.testControlForm.controlType" :items="store.options.controlTypes" placeholder="Choose Control Type" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
              </div>
              <UFormField label="Control Description" required>
                <UTextarea v-model="store.testControlForm.controlDescription" placeholder="Describe how this control works" class="w-full" :disabled="store.isReadOnlyTestControl" required />
              </UFormField>
            </div>

            <!-- Test Procedure -->
            <div class="  -gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Test Procedure</h4>
              <UFormField label="Test Steps" required>
                <UTextarea v-model="store.testControlForm.testProcedure" placeholder="Describe the steps taken to test this control" class="w-full" :disabled="store.isReadOnlyTestControl" required />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Test Result" required>
                  <USelectMenu v-model="store.testControlForm.testResult" :items="store.options.testResults" placeholder="Choose Test Result" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
              </div>
            </div>

            <!-- Finding and Recommendation -->
            <div class="  -gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Finding and Recommendation</h4>
              <UFormField label="Finding">
                <UTextarea v-model="store.testControlForm.finding" placeholder="Describe the finding from the test (if any)" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <UFormField label="Recommendation">
                <UTextarea v-model="store.testControlForm.recommendation" placeholder="Recommendations for improving ineffective controls" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
            </div>

            <!-- Mitigation Plan -->
            <div class="  -gray-50 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700">Mitigation Plan</h4>
              <UFormField label="Mitigation Plan">
                <UTextarea v-model="store.testControlForm.mitigationPlan" placeholder="Describe the mitigation plan for addressing control weaknesses" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="PIC">
                  <UInput v-model="store.testControlForm.pic" placeholder="PIC Name" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
                <UFormField label="Target Due Date">
                  <UInput v-model="store.testControlForm.dueDate" type="date" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
              </div>
            </div>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" label="Close" v-if="store.isReadOnlyTestControl" @click="() => { store.showTestControlModal = false }" />
              <template v-else>
                <UButton color="neutral" variant="soft" label="Cancel" @click="() => { store.showTestControlModal = false }" />
                <UButton color="primary" :label="store.isEditingTestControl ? 'Update' : 'Submit'" @click="store.saveTestControl()" />
              </template>
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'

const store = useAuditFieldworkStore()

const columns = [
  { accessorKey: 'controlName', header: 'Control Name' },
  { accessorKey: 'controlType', header: 'Control Type' },
  { accessorKey: 'testResult', header: 'Test Result' },
  { accessorKey: 'finding', header: 'Finding' },
  { accessorKey: 'mitigationPlan', header: 'Mitigation Plan' },
  { accessorKey: 'actions', header: 'Actions' }
]

const getControlTypeColor = (type: string) => {
  const colors: Record<string, "success" | "warning" | "info" | "neutral" | "primary"> = {
    'Preventive': 'success',
    'Detective': 'warning',
    'Corrective': 'info',
    'Manual': 'neutral',
    'Automated': 'primary'
  }
  return colors[type] || 'neutral'
}

const getResultColor = (result: string) => {
  const colors: Record<string, "success" | "error" | "warning" | "neutral"> = {
    'Effective': 'success',
    'Ineffective': 'error',
    'Partially Effective': 'warning',
    'Not Tested': 'neutral'
  }
  return colors[result] || 'neutral'
}
</script>
