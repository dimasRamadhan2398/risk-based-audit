<template>
  <div class="space-y-6">
    <section v-for="code in COSO_COMPONENT_CODES" :key="code" class="space-y-2">
      <h4 class="text-sm font-bold text-gray-900 dark:text-white">
        <span class="font-mono text-gray-500 dark:text-gray-400 mr-1">{{ code }}</span>{{ componentLabel(t, code) }}
      </h4>
      <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
        <table class="min-w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800/60 text-left text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
            <tr>
              <th class="px-3 py-2 w-[30%]">{{ t('icAssessment.principles.code') }}</th>
              <th class="px-3 py-2 w-[18%]">{{ t('icAssessment.principles.status') }}</th>
              <th class="px-3 py-2 w-[18%]">{{ t('icAssessment.principles.severity') }}</th>
              <th class="px-3 py-2">{{ t('icAssessment.principles.rationale') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="p in byComponent(code)" :key="p.code" class="align-top">
              <td class="px-3 py-3">
                <p class="font-semibold text-gray-900 dark:text-white">
                  <span class="font-mono text-gray-500 dark:text-gray-400 mr-1">{{ p.code }}</span>{{ principleTitle(t, p.code) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ t(`coso.principles.${p.code}.description`) }}</p>
                <div class="mt-1.5 flex flex-wrap items-center gap-x-2 text-xs text-gray-600 dark:text-gray-300">
                  <span v-if="p.taggedControlCount > 0">
                    {{ t('icAssessment.principles.taggedControls', { count: p.taggedControlCount }) }}<template v-if="formatPct(p.taggedControlAvgPct)">, {{ t('icAssessment.principles.taggedAvg', { pct: formatPct(p.taggedControlAvgPct)! }) }}</template>
                  </span>
                  <span v-else>{{ t('icAssessment.principles.noTaggedControls') }}</span>
                  <ULink
                    :to="{ path: '/risk-profile/risk-control-matrix', query: { department_id: departmentId || undefined, principle: p.code } }"
                    class="text-primary-600 dark:text-primary-400 hover:underline"
                  >
                    {{ t('icAssessment.principles.openRcm') }}
                  </ULink>
                </div>
              </td>
              <td class="px-3 py-3">
                <USelect
                  v-if="editable"
                  :model-value="p.status"
                  :items="statusItems"
                  class="w-full"
                  @update:model-value="(v: unknown) => onStatus(p, v as PrincipleStatus)"
                />
                <IcStatusText v-else :status="p.status" kind="principle" />
              </td>
              <td class="px-3 py-3">
                <USelect
                  v-if="editable"
                  :model-value="p.severity"
                  :items="severityItems(p.status)"
                  :disabled="ALLOWED_SEVERITIES[p.status].length < 2"
                  class="w-full"
                  @update:model-value="(v: unknown) => emit('update', p.code, { severity: v as Severity })"
                />
                <span v-else class="text-gray-800 dark:text-gray-200">{{ t(`icAssessment.severity.${p.severity}`) }}</span>
                <p v-if="showErrors && principleIssues(p).includes('severityNotAllowed')" class="text-xs text-red-700 dark:text-red-400 mt-1">
                  {{ t('icAssessment.principles.severityNotAllowed') }}
                </p>
              </td>
              <td class="px-3 py-3">
                <template v-if="editable">
                  <UTextarea
                    :model-value="p.rationale"
                    :rows="2"
                    autoresize
                    :maxlength="4000"
                    :placeholder="t('icAssessment.principles.rationalePlaceholder')"
                    :color="rationaleMissing(p) ? 'error' : undefined"
                    :highlight="rationaleMissing(p)"
                    class="w-full"
                    @update:model-value="(v: unknown) => emit('update', p.code, { rationale: String(v ?? '') })"
                  />
                  <p v-if="rationaleMissing(p)" class="text-xs text-red-700 dark:text-red-400 mt-1">
                    {{ p.status === 'NOT_APPLICABLE' ? t('icAssessment.principles.notApplicableRequired') : t('icAssessment.principles.rationaleRequired') }}
                  </p>
                </template>
                <p v-else class="text-gray-800 dark:text-gray-200 whitespace-pre-line">{{ p.rationale || '-' }}</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import IcStatusText from '~/components/ic-assessment/IcStatusText.vue'
import {
  ALLOWED_SEVERITIES,
  COSO_COMPONENT_CODES,
  PRINCIPLE_STATUSES,
  coerceSeverity,
  formatPct,
  principleIssues,
  type ComponentCode,
  type IcPrinciple,
  type PrincipleStatus,
  type Severity
} from '~/utils/icAssessment'
import { componentLabel, principleTitle } from '~/utils/icAssessmentLabels'

const props = defineProps<{
  principles: IcPrinciple[]
  departmentId: string
  editable: boolean
  showErrors: boolean
}>()

const emit = defineEmits<{
  update: [code: string, patch: Partial<IcPrinciple>]
}>()

const { t } = useI18n()

const byComponent = (code: ComponentCode) => props.principles.filter(p => p.componentCode === code)

const statusItems = computed(() => PRINCIPLE_STATUSES.map(s => ({ label: t(`icAssessment.principleStatus.${s}`), value: s })))
const severityItems = (status: PrincipleStatus) =>
  ALLOWED_SEVERITIES[status].map(s => ({ label: t(`icAssessment.severity.${s}`), value: s }))

// Changing the status moves the severity into the allowed range (NO -> MAJOR, NOT_APPLICABLE -> NONE ...)
const onStatus = (p: IcPrinciple, status: PrincipleStatus) =>
  emit('update', p.code, { status, severity: coerceSeverity(status, p.severity) })

// Shown while typing too once a rationale is needed, so the requirement is visible before submit
const rationaleMissing = (p: IcPrinciple) =>
  principleIssues(p).includes('rationaleRequired') && (props.showErrors || p.status !== 'NOT_ASSESSED')
</script>
