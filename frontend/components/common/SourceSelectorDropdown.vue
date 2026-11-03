<template>
  <div class="relative inline-block text-left">
    <UDropdownMenu :items="dropdownMenuItems">
      <UButton
        color="neutral"
        variant="subtle"
        size="sm"
        class="rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] hover:bg-[var(--bg-surface-hover)] shadow-xs transition-all duration-200 flex items-center gap-1.5 px-2.5 py-1.5"
      >
        <div class="flex items-center gap-1.5 min-w-0">
          <UIcon
            :name="isAllSources ? 'i-lucide-layers' : 'i-lucide-database'"
            class="w-4 h-4 text-primary-500 shrink-0"
          />
          <span class="text-xs font-semibold text-[var(--text-main)] max-w-[130px] sm:max-w-[180px] truncate">
            {{ isAllSources ? t('common.sourceSelector.allSources') : formatSourceLabel(selectedSource) }}
          </span>
          <span
            v-if="!isAllSources"
            class="w-2 h-2 rounded-full shrink-0"
            :class="getStatusDotClass(selectedSource?.status)"
          />
        </div>
        <UIcon name="i-lucide-chevron-down" class="w-3.5 h-3.5 text-[var(--text-muted)] shrink-0 ml-0.5" />
      </UButton>
    </UDropdownMenu>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useActiveDataSource } from '~/composables/useActiveDataSource'
import { useI18n } from '~/composables/useI18n'
import type { DataSourceItem } from '~/stores/data-source'

const router = useRouter()
const { t } = useI18n()
const { sources, selectedSourceId, selectedSource, isAllSources, setSelectedSourceId, fetchSources } = useActiveDataSource()

onMounted(() => {
  fetchSources()
})

function formatSourceLabel(src?: DataSourceItem | null) {
  if (!src) return t('common.sourceSelector.dataSource')
  const dbName = src.database || src.database_name
  return dbName ? `${src.name} (${dbName})` : src.name
}

function getStatusDotClass(status?: string) {
  if (status === 'Connected' || status === 'synced') return 'bg-emerald-500'
  if (status === 'Syncing') return 'bg-blue-500 animate-pulse'
  return 'bg-amber-500'
}

const dropdownMenuItems = computed(() => {
  const sourceItems = sources.value.map(src => {
    const sid = src.id || src.source_id || ''
    const isSelected = selectedSourceId.value === sid
    const dbName = src.database || src.database_name
    return {
      label: dbName ? `${src.name} (${dbName})` : src.name,
      icon: isSelected ? 'i-lucide-check-circle-2' : 'i-lucide-database',
      color: isSelected ? 'primary' : 'neutral',
      onSelect: () => setSelectedSourceId(sid)
    }
  })

  return [
    [
      {
        label: t('common.sourceSelector.allSourcesConsolidated'),
        icon: isAllSources.value ? 'i-lucide-check-circle-2' : 'i-lucide-layers',
        color: isAllSources.value ? 'primary' : 'neutral',
        onSelect: () => setSelectedSourceId('all')
      }
    ],
    sourceItems.length > 0 ? sourceItems : [
      {
        label: 'Core Banking Simulator (cbs_simulator)',
        icon: 'i-lucide-database',
        onSelect: () => setSelectedSourceId('cbs_simulator')
      }
    ],
    [
      {
        label: t('common.sourceSelector.manage'),
        icon: 'i-lucide-settings-2',
        onSelect: () => router.push('/settings?tab=datasource')
      }
    ]
  ]
})
</script>
