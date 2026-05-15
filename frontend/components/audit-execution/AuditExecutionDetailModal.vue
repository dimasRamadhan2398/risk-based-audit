<script setup lang="ts">
import type { AuditExecution } from '~/types/audit'

const props = defineProps<{
  open: boolean
  audit?: AuditExecution
}>()

const emit = defineEmits(['update:open', 'remind'])

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})

const sendReminder = () => {
  emit('remind', props.audit)
  isOpen.value = false
}
</script>

<template>
  <UModal v-model:open="isOpen" :ui="{ content: 'sm:max-w-3xl' }" scrollable>
    <template #content>
      <UCard :ui="{ root: 'divide-y divide-gray-100 dark:divide-gray-800' }">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white">Follow-up Detail</h3>
              <p class="text-sm text-gray-500 mt-1">
                Ref: {{ audit?.ref }} | Status: [ <span :class="audit?.status_detail === 'Late' ? 'text-red-500' : 'text-green-500'">
                  <UIcon :name="audit?.status_detail === 'Late' ? 'i-lucide-circle-alert' : 'i-lucide-check-circle'" class="inline-block mr-1" />
                  {{ audit?.status_detail }}
                </span> ]
              </p>
            </div>
            <UButton color="neutral" variant="ghost" icon="i-lucide-x" class="-my-1" @click="isOpen = false" />
          </div>
        </template>

        <div class="space-y-6">
          <!-- Section 1: Sample Data & Test Controls -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">[1] SAMPLE DATA & TEST CONTROLS</h4>
              <div class="grid grid-cols-4 gap-4 items-center">
                <span class="text-sm font-medium text-gray-500">Progress</span>
                <div class="col-span-3 flex items-center gap-3">
                  <UProgress :value="audit?.sample_data_test_controls?.progress || 0" color="secondary" class="flex-1" />
                  <span class="text-sm font-bold text-secondary">{{ audit?.sample_data_test_controls?.progress || 0 }} %</span>
                </div>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Description</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.sample_data_test_controls?.description || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Section 2: Working Papers -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">[2] WORKING PAPERS</h4>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Condition</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.working_papers?.condition || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Criteria</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.working_papers?.criteria || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Section 3: Action Plan Improvements -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">Action Plan Improvements</h4>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Recommendation</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.recommendation || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Deadline</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.deadline || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">PIC</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.pic || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <hr class="border-gray-200 dark:border-gray-700" />

          <!-- Section 4: Latest Update Progress -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">Latest Update Progress</h4>
              <div class="grid grid-cols-4 gap-4 items-center">
                <span class="text-sm font-medium text-gray-500">Attachment</span>
                <div class="col-span-3">
                  <div v-if="audit?.latest_update_progress?.attachment" class="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400">
                    <UIcon name="i-lucide-file-text" />
                    <span>[ {{ audit?.latest_update_progress?.attachment }} ]</span>
                  </div>
                  <span v-else class="text-sm text-gray-400">No attachment</span>
                </div>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">Description</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300 italic">
                  {{ audit?.latest_update_progress?.description || '-' }}
                </p>
              </div>
            </div>
          </UCard>
        </div>

        <template #footer>
          <div class="flex justify-end">
            <UButton
              label="Send Reminder"
              color="primary"
              icon="i-lucide-bell"
              @click="sendReminder"
            />
          </div>
        </template>
      </UCard>
    </template>
  </UModal>
</template>
