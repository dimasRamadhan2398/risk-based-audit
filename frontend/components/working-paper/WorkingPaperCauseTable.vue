<template>
    <div class="mt-4 space-y-6">
        <div class="mt-10">
        <TableEntities
          :data="store.filteredDataF04"
          :columns="store.columnsF04"
          :empty-state="store.hasAssignmentLetter ? { icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' } : { icon: 'i-heroicons-document-magnifying-glass', label: t('workingPaper.index.selectLetterFirst') }"
          :ui="{ td: '!whitespace-normal' }"
        >
            <template #condition-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.condition || '-' }}
              </div>
            </template>

            <template #riskLevel-cell="{ row }">
              <UBadge
                v-if="row.original.riskLevel"
                :color="getRiskLevelColor(row.original.riskLevel)"
                variant="subtle"
                size="sm"
              >
                Level {{ String(row.original.riskLevel).replace(/^level\s*/i, '') }}
              </UBadge>
              <span v-else class="text-gray-400">-</span>
            </template>

            <template #criteria-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.criteria || '-' }}
              </div>
            </template>

            <template #recommendations-cell="{ row }">
              <div v-if="Array.isArray(row.original.recommendations) && row.original.recommendations.length > 0" class="space-y-1 py-1">
                <div
                  v-for="(rec, idx) in row.original.recommendations"
                  :key="rec.id || idx"
                  class="text-xs leading-relaxed text-gray-700 dark:text-gray-300"
                >
                  <span class="font-semibold">{{ idx + 1 }}.</span> {{ typeof rec === 'string' ? rec : rec.recommendation }}
                </div>
              </div>
              <span v-else class="text-gray-400">-</span>
            </template>

            <template #actionPlan-cell="{ row }">
              <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
                <div class="whitespace-normal break-words leading-relaxed">
                  {{ row.original.actionPlan || '-' }}
                </div>
                <div v-if="row.original.targetDate" class="text-[11px] text-gray-400">
                  Target: {{ row.original.targetDate }}
                </div>
              </div>
            </template>

            <template #pic-cell="{ row }">
              <div class="space-y-0.5 text-xs">
                <div class="font-medium text-gray-700 dark:text-gray-200 whitespace-nowrap">
                  {{ store.formatEmployeeName(row.original.pic) }}
                </div>
                <div v-if="row.original.finalApprover" class="text-[11px] text-gray-400 whitespace-nowrap">
                  Approver: {{ store.formatEmployeeName(row.original.finalApprover) }}
                </div>
              </div>
            </template>

            <template #evidenceFile-cell="{ row }">
                <div v-if="row.original.evidenceFile" class="flex items-center gap-1 text-blue-600">
                    <UIcon name="i-heroicons-paper-clip" />
                    <span class="text-md truncate max-w-[150px]">{{ row.original.evidenceFile.name }}</span>
                </div>
                <span v-else class="text-gray-400">-</span>
            </template>

            <template #rootCause-cell="{ row }">
                <div class="space-y-2 py-2">
                    <div v-for="rca in row.original.rootCause" :key="rca.id" class="text-[11px] leading-tight border-l-2 border-orange-400 pl-2">
                        <div class="font-bold text-lg text-gray-700">{{ rca.method }}</div>
                    <div class="text-gray-500 text-sm italic">
                        {{ t('workingPaper.common.why') }}: {{ rca.w1 || '-' }} <br>
                    </div>
                    </div>
                </div>
            </template>

            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                <UTooltip v-if="canWrite" :text="t('workingPaper.common.edit')">
                <UButton 
                    size="md" 
                    color="warning" 
                    variant="ghost" 
                    icon="i-lucide-edit" 
                    @click="store.handleEditF04(row.original)" 
                />
                </UTooltip>
                <UTooltip v-if="canWrite" :text="t('workingPaper.common.delete')">
                <UButton 
                    size="md" 
                    color="error" 
                    variant="ghost" 
                    icon="i-lucide-trash-2" 
                    @click="store.handleDeleteF04(row.original.id)" 
                />
                </UTooltip>
                </div>
            </template>
        </TableEntities>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRbac } from '~/composables/useRbac'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const { canWrite } = useRbac()
const store = useWorkingPaperStore()

const getRiskLevelColor = (level?: string) => {
  const num = String(level || '').replace(/^level\s*/i, '').trim()
  if (num === '5' || num === '4') return 'error'
  if (num === '3') return 'warning'
  if (num === '2') return 'info'
  return 'success'
}

onMounted(() => {
  store.fetchEmployees?.()
})
</script>