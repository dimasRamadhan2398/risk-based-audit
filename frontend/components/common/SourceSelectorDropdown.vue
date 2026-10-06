<template>
  <div class="relative inline-block text-left">
    <UDropdownMenu :items="dropdownMenuItems">
      <UButton
        color="neutral"
        variant="subtle"
        size="sm"
        class="rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] hover:bg-[var(--bg-surface-hover)] shadow-xs transition-all duration-200 flex items-center gap-1.5 px-2.5 py-1.5"
      >
        <div class="flex items-center gap-1.5">
          <UIcon
            :name="isAllSources ? 'i-lucide-layers' : 'i-lucide-database'"
            class="w-4 h-4 text-primary-500 shrink-0"
          />
          <span class="text-xs font-semibold text-[var(--text-main)] max-w-[120px] sm:max-w-[160px] truncate">
            {{ isAllSources ? 'Semua Sumber' : (selectedSource?.name || 'Sumber Data') }}
          </span>
          <span
            v-if="!isAllSources"
            class="w-2 h-2 rounded-full shrink-0"
            :class="selectedSource?.status === 'synced' ? 'bg-emerald-500' : 'bg-amber-500'"
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

const router = useRouter()
const { sources, selectedSourceId, selectedSource, isAllSources, setSelectedSourceId, fetchSources } = useActiveDataSource()

onMounted(() => {
  fetchSources()
})

const dropdownMenuItems = computed(() => {
  const sourceItems = sources.value.map(src => {
    const sid = src.id || src.source_id || ''
    const isSelected = selectedSourceId.value === sid
    return {
      label: src.name,
      icon: isSelected ? 'i-lucide-check-circle-2' : 'i-lucide-database',
      color: isSelected ? 'primary' : 'neutral',
      onSelect: () => setSelectedSourceId(sid)
    }
  })

  return [
    [
      {
        label: 'Semua Sumber Data (Consolidated)',
        icon: isAllSources.value ? 'i-lucide-check-circle-2' : 'i-lucide-layers',
        color: isAllSources.value ? 'primary' : 'neutral',
        onSelect: () => setSelectedSourceId('all')
      }
    ],
    sourceItems.length > 0 ? sourceItems : [
      {
        label: 'Core Banking Simulator',
        icon: 'i-lucide-database',
        onSelect: () => setSelectedSourceId('cbs_simulator')
      }
    ],
    [
      {
        label: 'Kelola Sumber Data...',
        icon: 'i-lucide-settings-2',
        onSelect: () => router.push('/settings')
      }
    ]
  ]
})
</script>
