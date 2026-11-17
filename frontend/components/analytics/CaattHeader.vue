<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'

const props = defineProps<{
  title: string
  subtitle?: string
  icon?: string
  badgeText?: string
  badgeColor?: 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'error' | 'neutral'
  hideSourceBar?: boolean
}>()

const { t } = useI18n()
const {
  loading,
  runningCaatt,
  dataHubOffline,
  selectedSourceId,
  activeSourceMeta,
  sources,
  setSelectedSourceId,
  refreshSources,
  fetchCaattAnalytics,
  triggerCaattRun
} = useCaattAnalytics()

onMounted(() => {
  if (sources.value.length === 0) {
    refreshSources().catch(() => {})
  }
})

const sourceOptions = computed(() => {
  const opts: Array<{ label: string; value: string }> = [
    { label: t('analytics.caattUi.common.autoLatestSource'), value: 'all' }
  ]
  const seen = new Set<string>(['all'])

  for (const s of sources.value) {
    const sid = s.id || s.source_id
    if (!sid || seen.has(sid)) continue
    seen.add(sid)
    const rawType = s.source_type || s.type
    const dbSuffix = rawType ? ` (${String(rawType).toUpperCase()})` : ''
    opts.push({
      label: `${s.name || sid}${dbSuffix}`,
      value: sid
    })
  }

  const activeId = activeSourceMeta.value.sourceId
  if (activeId && !seen.has(activeId)) {
    seen.add(activeId)
    opts.push({
      label: activeSourceMeta.value.sourceName || activeId,
      value: activeId
    })
  }

  if (!seen.has('cbs_simulator')) {
    opts.push({
      label: t('analytics.caattUi.common.cbsSimulatorSource'),
      value: 'cbs_simulator'
    })
  }

  return opts
})

const currentSourceBinding = computed({
  get: () => selectedSourceId.value || 'all',
  set: (val: string) => {
    setSelectedSourceId(val || 'all')
  }
})
</script>

<template>
  <div class="space-y-4 mb-6">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <!-- Breadcrumbs -->
        <nav class="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mb-2">
          <NuxtLink to="/analytics" class="hover:text-primary transition-colors font-medium">
            {{ t('navigation.analytics') }}
          </NuxtLink>
          <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
          <span class="font-medium text-gray-600 dark:text-gray-300">{{ t('analytics.hub.caattTitle') }}</span>
          <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
          <span class="text-indigo-600 dark:text-indigo-400 font-bold">{{ props.title }}</span>
        </nav>

        <h1 class="text-2xl md:text-3xl font-black tracking-tight text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon :name="props.icon || 'i-heroicons-table-cells'" class="text-indigo-600 dark:text-indigo-400 w-8 h-8 shrink-0" />
          <span>{{ props.title }}</span>
        </h1>
        <p v-if="props.subtitle" class="text-sm md:text-base text-gray-500 dark:text-gray-400 mt-1">
          {{ props.subtitle }}
        </p>
      </div>

      <div v-if="props.badgeText" class="flex items-center gap-2 shrink-0">
        <UBadge :color="props.badgeColor || 'primary'" variant="subtle" size="lg" class="px-3.5 py-1.5 font-bold shadow-xs">
          {{ props.badgeText }}
        </UBadge>
      </div>
    </div>

    <!-- Unified Active Source Selector & CAATT Engine Toolbar -->
    <div
      v-if="!props.hideSourceBar"
      class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3 shadow-xs"
    >
      <div class="flex flex-wrap items-center gap-2.5">
        <span class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
          {{ t('analytics.caattUi.common.sourceLabel') }}:
        </span>
        <USelectMenu
          v-model="currentSourceBinding"
          :items="sourceOptions"
          value-key="value"
          label-key="label"
          size="sm"
          class="w-full sm:w-72"
        />
        <UBadge
          v-if="activeSourceMeta.sourceName || activeSourceMeta.sourceId"
          :color="dataHubOffline ? 'error' : 'success'"
          variant="subtle"
          size="sm"
          class="font-medium"
        >
          <UIcon name="i-lucide-database" class="w-3.5 h-3.5 mr-1" />
          {{ t('analytics.caattUi.common.showingDataFrom') }}:
          {{ activeSourceMeta.sourceName || activeSourceMeta.sourceId }}
        </UBadge>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <UButton
          color="neutral"
          variant="outline"
          size="sm"
          icon="i-lucide-refresh-cw"
          :loading="loading"
          :disabled="loading || runningCaatt"
          @click="fetchCaattAnalytics(true)"
        >
          {{ t('analytics.caattUi.common.refresh') }}
        </UButton>
        <UButton
          color="primary"
          variant="solid"
          size="sm"
          icon="i-lucide-play"
          :loading="runningCaatt"
          :disabled="loading || runningCaatt"
          @click="triggerCaattRun()"
        >
          {{ runningCaatt ? t('analytics.caattUi.common.running') : t('analytics.caattUi.common.rerunEngine') }}
        </UButton>
        <UButton
          to="/settings?tab=datasource"
          color="neutral"
          variant="ghost"
          size="sm"
          icon="i-lucide-settings-2"
        >
          {{ t('analytics.caattUi.common.manageSources') }}
        </UButton>
      </div>
    </div>

    <!-- Shared Offline Banner across all CAATT Tabs -->
    <UAlert
      v-if="dataHubOffline && !loading"
      icon="i-lucide-wifi-off"
      color="error"
      variant="subtle"
      :title="t('analytics.caattUi.common.dataHubOfflineTitle')"
      :description="t('analytics.caattUi.common.dataHubOfflineDesc')"
    />
  </div>
</template>
