<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-3 sm:gap-4 w-full">
      <CommonSearchInput
        v-model="search"
        :placeholder="t('assignmentLetter.filters.searchPlaceholder')"
        class="w-full sm:flex-1 sm:min-w-64"
      />
      <USelectMenu
        v-model="workUnit"
        :items="workUnitItems"
        :placeholder="t('assignmentLetter.filters.workUnit')"
        class="w-full sm:w-56"
      />
      <USelectMenu
        v-model="status"
        :items="statusItems"
        value-key="value"
        :placeholder="t('assignmentLetter.filters.status')"
        class="w-full sm:w-52"
      />
      <UButton
        v-if="hasFilters"
        :label="t('assignmentLetter.filters.reset')"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        class="w-full sm:w-auto justify-center"
        @click="resetFilters"
      />
    </div>

    <UCard
      class="rounded-xl shadow-sm overflow-hidden"
      variant="soft"
    >
      <TableEntities
        :columns="columns"
        :data="filteredLetters"
        :empty-state="{ icon: 'i-heroicons-document-text', label: t('assignmentLetter.table.empty') }"
        class="w-full text-sm"
      >
        <template #letterNumber-cell="{ row }">
          <span class="font-bold text-orange-600 ">
            {{ row.original.letterNumber }}
          </span>
        </template>

        <template #auditTitle-cell="{ row }">
          <div
            class="line-clamp-2 overflow-hidden max-w-xs font-medium text-gray-800 dark:text-gray-200"
            :title="row.original.auditTitle"
          >
            {{ row.original.auditTitle || '-' }}
          </div>
        </template>

        <template #executionPeriod-cell="{ row }">
          <div class="flex items-center gap-1 text-gray-600 ">
            <UIcon
              name="i-heroicons-calendar"
              class="w-4 h-4"
            />
            <span>{{ row.original.executionPeriod }}</span>
          </div>
        </template>

        <template #auditTeam-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-800 dark:text-white">{{ row.original.leader }} {{ t('assignmentLetter.formExtra.pic') }}</span>
            <span class="text-md text-gray-500">{{ row.original.auditTeam }}</span>
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge
            :color="row.original.status === 'Published' ? 'success' : 'neutral'"
            variant="subtle"
            size="lg"
            class="font-bold"
          >
            {{ statusLabel(row.original.status) }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center justify-center gap-1">
            <UTooltip
              v-if="canManageAssignmentLetter && row.original.status === 'Draft'"
              :text="t('assignmentLetter.formExtra.publish')"
            >
              <UButton
                size="md"
                color="success"
                variant="ghost"
                icon="i-lucide-send"
                :loading="store.loading"
                @click="store.publishLetter(row.original)"
              />
            </UTooltip>
            <UTooltip v-if="canWrite" :text="t('assignmentLetter.formExtra.edit')">
              <UButton
                size="md"
                color="warning"
                variant="ghost"
                icon="i-lucide-edit"
                @click="store.openEditModal(row.original)"
              />
            </UTooltip>
            <UTooltip v-if="canWrite" :text="t('assignmentLetter.formExtra.delete')">
              <UButton
                size="md"
                color="error"
                variant="ghost"
                icon="i-lucide-trash-2"
                @click="store.deleteSuratTugas(row.original.id)"
              />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const { t } = useI18n()

const store = useAssignmentLetterStore()

// Column keys/classes live in the store; headers are resolved here so they follow the active locale.
const headerKeys: Record<string, string> = {
  letterNumber: 'assignmentLetter.table.columns.letterNumber',
  auditTitle: 'assignmentLetter.table.columns.auditTitle',
  workingUnit: 'assignmentLetter.table.columns.workingUnit',
  executionPeriod: 'assignmentLetter.table.columns.executionPeriod',
  auditTeam: 'assignmentLetter.table.columns.auditTeam',
  status: 'assignmentLetter.table.columns.status',
  actions: 'common.columns.actions'
}
const columns = computed(() =>
  store.columns.map((col: any) => ({
    ...col,
    header: headerKeys[col.accessorKey] ? t(headerKeys[col.accessorKey]!) : col.header
  }))
)

const STATUSES = ['Draft', 'Waiting Approval', 'Published', 'Archived']
const statusLabel = (status: string) =>
  STATUSES.includes(status) ? t(`assignmentLetter.table.status.${status}`) : status

// Client-side filters: the list endpoint is fetched once (page_size=100) without query params.
const search = ref('')
const status = ref<string | undefined>(undefined)
const workUnit = ref<string | undefined>(undefined)

const statusItems = computed(() =>
  STATUSES.map(value => ({ value, label: statusLabel(value) }))
)
const workUnitItems = computed(() =>
  Array.from(new Set(store.assignmentLetterList.map(l => l.workingUnit).filter(Boolean))).sort()
)

const hasFilters = computed(() => !!(search.value.trim() || status.value || workUnit.value))

const filteredLetters = computed(() => {
  const q = search.value.trim().toLowerCase()
  return store.assignmentLetterList.filter((l) => {
    if (status.value && l.status !== status.value) return false
    if (workUnit.value && l.workingUnit !== workUnit.value) return false
    if (!q) return true
    return [l.letterNumber, l.auditTitle, l.workingUnit, l.leader, l.auditTeam]
      .some(v => (v || '').toString().toLowerCase().includes(q))
  })
})

const resetFilters = () => {
  search.value = ''
  status.value = undefined
  workUnit.value = undefined
}

// Publishing is limited to the roles in the RBAC matrix ("Create / Publish Letter")
const { canManageAssignmentLetter, canWrite } = useRbac()

</script>
