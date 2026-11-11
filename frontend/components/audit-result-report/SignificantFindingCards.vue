<template>
  <div class="space-y-3">
    <UCollapsible
      v-for="(finding, idx) in findings"
      :key="idx"
      :open="isExpanded(finding)"
      class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 shadow-sm overflow-hidden"
      @update:open="setExpanded(finding, $event)"
    >
      <template #default="{ open }">
        <UButton
          color="neutral"
          variant="ghost"
          block
          class="justify-between items-center rounded-none px-4 py-3 h-auto text-left"
          :trailing-icon="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
        >
          <span class="flex min-w-0 flex-1 flex-col items-start gap-1 text-left">
            <span class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              {{ t('executiveSummaryForm.topFindings.cardTitle', { number: idx + 1 }) }}
            </span>
            <span class="max-w-full truncate text-sm font-bold text-gray-900 dark:text-white">
              {{ finding.judulTemuan || t('executiveSummaryForm.topFindings.untitled') }}
            </span>
          </span>
          <span class="hidden shrink-0 items-center gap-2 md:flex">
            <span
              v-if="finding.unitDivision"
              class="max-w-48 truncate text-xs text-gray-500 dark:text-gray-400"
            >
              {{ finding.unitDivision }}
            </span>
            <UBadge
              v-if="finding.risiko"
              :color="riskColor(finding.risiko)"
              variant="subtle"
              size="xs"
            >
              {{ riskLabel(finding.risiko) }}
            </UBadge>
            <FollowUpStatusBadge
              :status="finding.statusTL"
              :reason="finding.noActionReason"
            />
          </span>
        </UButton>
      </template>

      <template #content>
        <div class="grid grid-cols-1 gap-4 border-t border-gray-200 dark:border-gray-800 p-4 md:grid-cols-2">
          <UFormField :label="t('executiveSummaryForm.table.unitDivision')">
            <USelectMenu
              v-model="finding.unitDivision"
              :items="unitOptions"
              :loading="unitLoading"
              :placeholder="t('executiveSummaryForm.placeholders.division')"
              :disabled="disabled"
              class="w-full"
              create-item
              @create="(val: string) => { finding.unitDivision = val.trim() }"
            >
              <template #empty>
                {{ unitLoading ? t('common.loading') : (unitError || t('masterData.department.empty')) }}
              </template>
            </USelectMenu>
          </UFormField>

          <UFormField :label="t('executiveSummaryForm.table.findingTitle')">
            <UInput
              v-model="finding.judulTemuan"
              :placeholder="t('executiveSummaryForm.placeholders.findingTitle')"
              :disabled="disabled"
              maxlength="100"
              class="w-full"
              @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.findingTitle'))"
              @input="($event.target as any)?.setCustomValidity('')"
            />
            <div class="mt-1 text-right text-xs text-gray-500">
              {{ finding.judulTemuan ? finding.judulTemuan.length : 0 }}/100
            </div>
          </UFormField>

          <UFormField :label="t('executiveSummaryForm.table.riskValue')">
            <USelectMenu
              v-model="finding.risiko"
              :items="riskOptions"
              value-key="value"
              :placeholder="t('executiveSummaryForm.placeholders.risk')"
              :disabled="disabled"
              class="w-full"
            />
          </UFormField>

          <!-- A finding taken from the LHA follows the bucket of its ATR; only manual rows pick a status -->
          <UFormField
            :label="t('executiveSummaryForm.table.followUpStatus')"
            :hint="finding.findingId ? t('executiveSummaryForm.followUp.atrLinked') : undefined"
          >
            <USelectMenu
              v-model="finding.statusTL"
              :items="statusOptions"
              value-key="value"
              :placeholder="t('executiveSummaryForm.table.followUpStatus')"
              :disabled="disabled || !!finding.findingId"
              class="w-full"
            />
            <div
              v-if="finding.statusTL === 'No Action' && finding.noActionReason"
              class="mt-1"
            >
              <FollowUpStatusBadge
                status="No Action"
                :reason="finding.noActionReason"
              />
            </div>
          </UFormField>

          <UFormField
            :label="t('executiveSummaryForm.table.proposedAction')"
            class="md:col-span-2"
          >
            <UTextarea
              v-model="finding.usulan"
              :placeholder="t('executiveSummaryForm.placeholders.proposedAction')"
              :disabled="disabled"
              class="w-full"
            />
          </UFormField>

          <div
            v-if="!disabled"
            class="flex justify-end md:col-span-2"
          >
            <UButton
              color="error"
              variant="ghost"
              icon="i-lucide-trash"
              size="sm"
              :label="t('executiveSummaryForm.topFindings.remove')"
              @click="removeAt(idx)"
            />
          </div>
        </div>
      </template>
    </UCollapsible>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import type { SignificantFinding } from '~/stores/executive-summary'
import { createFindingExpansion } from '~/utils/findingCardState'
import { FOLLOW_UP_STATUSES, followUpStatusI18nKey } from '~/utils/executiveSummaryFollowUp'
import FollowUpStatusBadge from './FollowUpStatusBadge.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  findings: SignificantFinding[]
  disabled?: boolean
  unitOptions?: string[]
  unitLoading?: boolean
  unitError?: string
  riskOptions?: { label: string, value: string }[]
}>(), {
  disabled: false,
  unitOptions: () => [],
  unitLoading: false,
  unitError: '',
  riskOptions: () => []
})

const emit = defineEmits<{
  remove: [index: number]
}>()

const statusOptions = computed(() => FOLLOW_UP_STATUSES.map(value => ({ label: t(followUpStatusI18nKey(value)), value })))

// Expanded state keyed by finding identity (see utils/findingCardState.ts). Nothing is open on load.
const expansion = createFindingExpansion(reactive(new Set<object>()))
const { isExpanded, setExpanded, expand } = expansion

const removeAt = (idx: number) => {
  const finding = props.findings[idx]
  if (finding) expansion.forget(finding)
  emit('remove', idx)
}

// Called by the parent right after a finding is added, so only that card opens
defineExpose({ expand })

const riskLabel = (value: string) => props.riskOptions.find(o => o.value === value)?.label || value

const riskColor = (value: string) => {
  switch (value) {
    case 'Very Significant': return 'error'
    case 'Significant': return 'warning'
    case 'Quite Significant': return 'info'
    default: return 'neutral'
  }
}
</script>
