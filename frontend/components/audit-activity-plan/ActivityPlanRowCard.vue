<template>
  <div
    v-if="row"
    :id="`activity-plan-row-${index + 1}`"
    class="rounded-xl border bg-[var(--bg-surface)] p-4 sm:p-5 shadow-sm space-y-4"
    :class="highlighted ? 'border-error-400 ring-2 ring-error-200 dark:ring-error-900' : 'border-[var(--border-main)]'"
    data-testid="activity-plan-row"
  >
    <!-- Header: No, object, controls -->
    <div class="flex flex-wrap items-start justify-between gap-3 border-b border-[var(--border-main)] pb-3">
      <div class="flex min-w-0 items-start gap-3">
        <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 dark:bg-primary-900/30 text-sm font-bold text-primary-600 dark:text-primary-300">
          {{ index + 1 }}
        </span>
        <div class="min-w-0">
          <p class="font-semibold text-[var(--text-main)] break-words">
            {{ row.pkat.auditObject || '-' }}
          </p>
          <p
            v-if="row.pkat.department?.name"
            class="text-xs text-[var(--text-muted)]"
          >
            {{ row.pkat.department.name }}
          </p>
        </div>
      </div>
      <div class="flex shrink-0 items-center gap-1">
        <UTooltip :text="t('auditActivityPlan.rows.moveUp')">
          <UButton
            icon="i-heroicons-arrow-up"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="index === 0"
            :aria-label="t('auditActivityPlan.rows.moveUp')"
            @click="store.moveRow(index, -1)"
          />
        </UTooltip>
        <UTooltip :text="t('auditActivityPlan.rows.moveDown')">
          <UButton
            icon="i-heroicons-arrow-down"
            color="neutral"
            variant="ghost"
            size="sm"
            :disabled="index === total - 1"
            :aria-label="t('auditActivityPlan.rows.moveDown')"
            @click="store.moveRow(index, 1)"
          />
        </UTooltip>
        <UTooltip :text="row.inUse ? t('auditActivityPlan.rows.inUseTooltip') : t('auditActivityPlan.rows.remove')">
          <UButton
            icon="i-heroicons-trash"
            color="error"
            variant="ghost"
            size="sm"
            :disabled="row.inUse"
            :aria-label="t('auditActivityPlan.rows.remove')"
            data-testid="row-remove"
            @click="store.removeRow(index)"
          />
        </UTooltip>
      </div>
    </div>

    <UAlert
      v-if="row.source === 'snapshot'"
      color="warning"
      variant="subtle"
      icon="i-heroicons-exclamation-triangle"
      :description="t('auditActivityPlan.rows.pkatMissing')"
    />

    <!-- Read-only: from the PKAT row -->
    <div
      class="rounded-lg border border-dashed border-[var(--border-main)] bg-gray-50 dark:bg-gray-900/40 p-3 space-y-3"
      data-testid="pkat-strip"
    >
      <p class="text-xs font-semibold text-[var(--text-muted)]">
        {{ t('auditActivityPlan.rows.fromPkat') }}
      </p>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <UFormField
          :label="t('auditActivityPlan.columns.auditObject')"
          class="sm:col-span-2"
        >
          <UInput
            :model-value="row.pkat.auditObject"
            disabled
            class="w-full"
            data-testid="pkat-field"
          />
        </UFormField>
        <UFormField
          :label="t('auditActivityPlan.columns.auditUniverse')"
          class="sm:col-span-2"
        >
          <div class="flex items-center gap-2">
            <UInput
              :model-value="row.pkat.auditUniverse?.name || '-'"
              disabled
              class="w-full"
              data-testid="pkat-field"
            />
            <span
              v-if="row.pkat.auditUniverse?.riskLevel"
              class="inline-flex shrink-0 items-center px-2 py-1 rounded-md text-xs font-semibold"
              :class="getRiskLevelColorClass(row.pkat.auditUniverse?.riskLevel)"
            >{{ row.pkat.auditUniverse?.riskLevel }}</span>
          </div>
        </UFormField>
        <UFormField :label="t('auditActivityPlan.columns.schedule')">
          <UInput
            :model-value="row.pkat.schedule || '-'"
            disabled
            class="w-full"
            data-testid="pkat-field"
          />
        </UFormField>
        <UFormField :label="t('auditActivityPlan.columns.teamSize')">
          <UInput
            :model-value="String(row.pkat.teamSize)"
            disabled
            class="w-full"
            data-testid="pkat-field"
          />
        </UFormField>
        <UFormField :label="t('auditActivityPlan.columns.totalMandays')">
          <UInput
            :model-value="String(row.pkat.totalMandays)"
            disabled
            class="w-full"
            data-testid="pkat-field"
          />
        </UFormField>
        <UFormField :label="t('auditActivityPlan.rows.category')">
          <UInput
            :model-value="row.pkat.auditCategory || '-'"
            disabled
            class="w-full"
            data-testid="pkat-field"
          />
        </UFormField>
      </div>
    </div>

    <!-- Owned by the Audit Activity Plan -->
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
      <UFormField
        :label="t('auditActivityPlan.columns.assignmentCode')"
        :help="row.assignmentCodeEditable ? t('auditActivityPlan.rows.assignmentCodeHint') : t('auditActivityPlan.rows.assignmentCodeLocked')"
      >
        <UInput
          v-model="row.assignmentCode"
          :disabled="!row.assignmentCodeEditable"
          :placeholder="t('auditActivityPlan.rows.assignmentCodeHint')"
          maxlength="50"
          class="w-full font-mono"
          data-testid="assignment-code"
        />
      </UFormField>

      <UFormField
        :label="t('auditActivityPlan.columns.auditKind')"
        required
      >
        <USelectMenu
          :model-value="row.auditKind || undefined"
          :items="kindItems"
          value-key="value"
          label-key="label"
          create-item
          :placeholder="t('auditActivityPlan.rows.auditKindPlaceholder')"
          class="w-full"
          data-testid="audit-kind"
          @update:model-value="(value: string | undefined) => { row!.auditKind = value || '' }"
          @create="(value: string) => { row!.auditKind = value.trim().slice(0, 100) }"
        />
      </UFormField>

      <UFormField
        :label="t('auditActivityPlan.columns.justification')"
        required
        class="md:col-span-2"
      >
        <NumberedListInput
          v-model="row.justification"
          :label="t('auditActivityPlan.columns.justification')"
          test-id="list-justification"
        />
      </UFormField>

      <UFormField
        :label="t('auditActivityPlan.columns.objectives')"
        required
        class="md:col-span-2"
      >
        <NumberedListInput
          v-model="row.objectives"
          :label="t('auditActivityPlan.columns.objectives')"
          test-id="list-objectives"
        />
      </UFormField>

      <UFormField
        :label="t('auditActivityPlan.columns.focus')"
        required
        class="md:col-span-2"
      >
        <NumberedListInput
          v-model="row.focus"
          :label="t('auditActivityPlan.columns.focus')"
          test-id="list-focus"
        />
      </UFormField>

      <UFormField
        :label="t('auditActivityPlan.columns.teamLeader')"
        :help="row.pkat.pkatSupervisorName ? t('auditActivityPlan.rows.pkatSupervisor', { name: row.pkat.pkatSupervisorName }) : undefined"
        class="md:col-span-2"
      >
        <USelectMenu
          :model-value="row.teamLeaderEmployeeId || undefined"
          :items="leaderItems"
          value-key="id"
          label-key="full_name"
          :loading="store.employeesLoading"
          clear
          :placeholder="t('auditActivityPlan.rows.teamLeaderPlaceholder')"
          class="w-full"
          data-testid="team-leader"
          @update:model-value="setTeamLeader"
        >
          <template #empty>
            {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('auditActivityPlan.rows.teamLeaderEmpty')) }}
          </template>
        </USelectMenu>
      </UFormField>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useI18n } from '~/composables/useI18n'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import NumberedListInput from '~/components/shared/NumberedListInput.vue'

/** One assignment of the Audit Activity Plan form: the PKAT values read-only, the RKA values editable. */
const props = defineProps<{
  index: number
  total: number
  /** Scrolled to and outlined after a server error about this row. */
  highlighted?: boolean
}>()

const { t } = useI18n()
const store = useActivityPlanStore()

const row = computed(() => store.formState.rows[props.index])

// A custom Jenis Audit typed earlier stays selectable.
const kindItems = computed(() => {
  const items = [...store.auditKindItems]
  const current = row.value?.auditKind
  if (current && !items.some(i => i.value === current)) items.push({ value: current, label: current })
  return items
})

// The saved Ketua Tim stays selectable even if the employee is no longer listed.
const leaderItems = computed(() => {
  const items = store.employees.map(e => ({ id: e.id, full_name: e.full_name }))
  const r = row.value
  if (r?.teamLeaderEmployeeId && !items.some(i => i.id === r.teamLeaderEmployeeId)) {
    items.unshift({ id: r.teamLeaderEmployeeId, full_name: r.teamLeaderName || r.teamLeaderEmployeeId })
  }
  return items
})

const setTeamLeader = (id: string | null | undefined) => {
  const r = row.value
  if (!r) return
  r.teamLeaderEmployeeId = id || null
  r.teamLeaderName = id ? (leaderItems.value.find(i => i.id === id)?.full_name || '') : ''
}
</script>
