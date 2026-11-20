<script setup lang="ts">
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useAiAnalytics } from '~/composables/useAiAnalytics'

const props = withDefaults(defineProps<{
  modelKey: 'entity' | 'anomaly' | 'nlp' | 'kpi' | 'kpi_auditee' | 'kpi_auditor'
  compact?: boolean
  showRetrainButton?: boolean
  buttonLabel?: string
}>(), {
  compact: false,
  showRetrainButton: false,
  buttonLabel: 'Latih Ulang Model AI'
})

const emit = defineEmits<{
  (e: 'retrained', result: any): void
}>()

const {
  modelStatuses,
  retrainQueue,
  retrainResource,
  checkingRetrainStatus,
  fetchRetrainStatus,
  retrainModel
} = useAiAnalytics()

const resolvedKey = computed<'entity' | 'anomaly' | 'nlp' | 'kpi_auditee' | 'kpi_auditor'>(() => {
  if (props.modelKey === 'kpi') {
    const aud = modelStatuses.value.kpi_auditor
    const aee = modelStatuses.value.kpi_auditee
    if (aud?.state === 'training' || aud?.state === 'queued') return 'kpi_auditor'
    if (aee?.state === 'training' || aee?.state === 'queued') return 'kpi_auditee'
    if (aud?.has_new_data) return 'kpi_auditor'
    return 'kpi_auditee'
  }
  return props.modelKey
})

const currentModel = computed(() => {
  return modelStatuses.value[resolvedKey.value] || {
    key: resolvedKey.value,
    label: props.modelKey.toUpperCase(),
    state: 'trained',
    has_new_data: false,
    new_rows: 0,
    current_rows: 0,
    trained_rows: 0,
    trained_at: null,
    duration_seconds: null,
    samples: 0,
    tier: 'T1',
    metrics: null,
    message: 'Model AI siap digunakan.',
    last_error: null,
    progress: null,
    queue_position: null,
    waiting_reason: null
  }
})

const stateConfig = computed(() => {
  const st = currentModel.value.state
  if (st === 'training') {
    const prog = currentModel.value.progress
    const pct = prog?.percent ?? 15
    const bText = prog?.batch && prog?.total_batches ? `Batch ${prog.batch}/${prog.total_batches} (${pct}%)` : `${pct}%`
    return {
      color: 'info' as const,
      dotClass: 'bg-sky-500 animate-ping',
      badgeText: `Sedang Melatih • ${bText}`,
      title: `Pelatihan Ulang ${currentModel.value.label} Sedang Berjalan (${bText})`,
      icon: 'i-lucide-cpu',
      percent: pct
    }
  }
  if (st === 'queued') {
    const pos = currentModel.value.queue_position || 1
    return {
      color: 'info' as const,
      dotClass: 'bg-indigo-500 animate-pulse',
      badgeText: `Dalam Antrian #${pos}`,
      title: `${currentModel.value.label} Menunggu Giliran Antrian (#${pos})`,
      icon: 'i-lucide-clock',
      percent: 5
    }
  }
  if (st === 'untrained_data' || currentModel.value.has_new_data) {
    const nRows = currentModel.value.new_rows || 1
    return {
      color: 'warning' as const,
      dotClass: 'bg-amber-500 animate-pulse',
      badgeText: `Ada Data Baru (+${nRows})`,
      title: `Ada Data Baru pada ${currentModel.value.label} yang Harus Dilatih (+${nRows} pembaruan)`,
      icon: 'i-lucide-alert-circle',
      percent: 0
    }
  }
  if (st === 'never_trained') {
    return {
      color: 'neutral' as const,
      dotClass: 'bg-gray-400',
      badgeText: 'Model Belum Dilatih',
      title: `Model ${currentModel.value.label} Belum Dilatih`,
      icon: 'i-lucide-circle-dashed',
      percent: 0
    }
  }
  if (st === 'failed') {
    return {
      color: 'error' as const,
      dotClass: 'bg-rose-500',
      badgeText: 'Pelatihan Gagal / Rollback',
      title: `Pelatihan ${currentModel.value.label} Gagal / Rollback Kualitas`,
      icon: 'i-lucide-alert-triangle',
      percent: 0
    }
  }
  return {
    color: 'success' as const,
    dotClass: 'bg-emerald-500',
    badgeText: 'Model Sudah Dilatih',
    title: `Model ${currentModel.value.label} Sudah Dilatih & Siap Digunakan`,
    icon: 'i-lucide-shield-check',
    percent: 100
  }
})

const formattedMetric = computed(() => {
  const m = currentModel.value.metrics
  if (!m || m.primary === undefined || m.primary === null) return null
  const name = String(m.primary_name || 'score').toUpperCase().replace('_', '-')
  const val = Number(m.primary)
  if (name === 'RMSE' || name === 'MAE') {
    return `${name}: ${val.toFixed(4)}`
  }
  return `${name}: ${(val * 100).toFixed(1)}%`
})

const formattedTrainedTime = computed(() => {
  const iso = currentModel.value.trained_at
  if (!iso) return null
  try {
    const dt = new Date(iso)
    return dt.toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return null
  }
})

const isBusy = computed(() =>
  currentModel.value.state === 'training' || currentModel.value.state === 'queued'
)

let pollTimer: ReturnType<typeof setInterval> | null = null

const adjustPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  const anyActive = Object.values(modelStatuses.value).some(
    (m: any) => m?.state === 'training' || m?.state === 'queued'
  ) || retrainQueue.value.length > 0
  const intervalMs = anyActive ? 2000 : 30000
  pollTimer = setInterval(() => {
    fetchRetrainStatus().catch(() => {})
  }, intervalMs)
}

watch(
  () => [currentModel.value.state, retrainQueue.value.length],
  () => adjustPolling()
)

onMounted(() => {
  fetchRetrainStatus().catch(() => {})
  adjustPolling()
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})

const handleTriggerModelRetrain = async () => {
  if (isBusy.value) return
  const res = await retrainModel(props.modelKey, { force: true })
  emit('retrained', res)
  adjustPolling()
}
</script>

<template>
  <!-- Compact inline pill mode -->
  <div v-if="compact" class="inline-flex items-center gap-2 flex-wrap">
    <UBadge
      :color="stateConfig.color"
      variant="subtle"
      size="sm"
      class="flex items-center gap-1.5 text-xs font-semibold"
      :title="currentModel.message"
    >
      <span class="size-2 rounded-full" :class="stateConfig.dotClass" />
      <span>{{ stateConfig.badgeText }}</span>
      <span v-if="currentModel.tier" class="opacity-75 font-mono text-[10px]">({{ currentModel.tier }})</span>
    </UBadge>
    <span v-if="formattedMetric" class="text-[11px] font-mono text-gray-500 dark:text-gray-400">
      {{ formattedMetric }}
    </span>
  </div>

  <!-- Full status card / banner mode -->
  <div
    v-else
    class="rounded-xl border p-3.5 transition-all"
    :class="{
      'border-sky-500/30 bg-sky-50/60 dark:bg-sky-950/20': stateConfig.color === 'info',
      'border-amber-500/40 bg-amber-50/60 dark:bg-amber-950/20': stateConfig.color === 'warning',
      'border-emerald-500/30 bg-emerald-50/50 dark:bg-emerald-950/15': stateConfig.color === 'success',
      'border-rose-500/40 bg-rose-50/60 dark:bg-rose-950/20': stateConfig.color === 'error',
      'border-gray-200 dark:border-gray-800 bg-gray-50/60 dark:bg-gray-900/40': stateConfig.color === 'neutral'
    }"
  >
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div class="flex items-start gap-3">
        <div
          class="flex size-9 items-center justify-center rounded-lg shrink-0 mt-0.5"
          :class="{
            'bg-sky-500/15 text-sky-600 dark:text-sky-400': stateConfig.color === 'info',
            'bg-amber-500/15 text-amber-600 dark:text-amber-400': stateConfig.color === 'warning',
            'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400': stateConfig.color === 'success',
            'bg-rose-500/15 text-rose-600 dark:text-rose-400': stateConfig.color === 'error',
            'bg-gray-500/15 text-gray-600 dark:text-gray-400': stateConfig.color === 'neutral'
          }"
        >
          <UIcon :name="stateConfig.icon" class="size-5" :class="{ 'animate-spin': currentModel.state === 'training' }" />
        </div>

        <div class="space-y-1">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-xs sm:text-sm font-bold text-gray-900 dark:text-white">
              {{ stateConfig.title }}
            </span>
            <UBadge :color="stateConfig.color" variant="subtle" size="xs" class="font-semibold">
              <span class="size-1.5 rounded-full mr-1" :class="stateConfig.dotClass" />
              {{ stateConfig.badgeText }}
            </UBadge>
            <UBadge v-if="currentModel.tier" color="neutral" variant="soft" size="xs" class="font-mono text-[10px]">
              Tier {{ currentModel.tier }}
            </UBadge>
            <UBadge v-if="formattedMetric" color="primary" variant="subtle" size="xs" class="font-mono text-[10px]">
              {{ formattedMetric }}
            </UBadge>
          </div>

          <p class="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
            {{ currentModel.message }}
            <span v-if="currentModel.waiting_reason && currentModel.state === 'queued'" class="font-medium text-indigo-600 dark:text-indigo-400">
              • {{ currentModel.waiting_reason }}
            </span>
          </p>

          <div class="flex items-center gap-3 flex-wrap text-[11px] text-gray-500 dark:text-gray-400 font-mono">
            <span v-if="currentModel.samples || currentModel.trained_rows">
              Sampel Terlatih: <strong>{{ currentModel.samples || currentModel.trained_rows }}</strong>
            </span>
            <span v-if="currentModel.duration_seconds !== null && currentModel.duration_seconds !== undefined">
              Durasi: <strong>{{ currentModel.duration_seconds }} dtk</strong>
            </span>
            <span v-if="formattedTrainedTime">
              Update: <strong>{{ formattedTrainedTime }}</strong>
            </span>
            <span v-if="retrainQueue.length > 0">
              Total Antrian Aktif: <strong>{{ retrainQueue.length }} job</strong>
            </span>
            <span v-if="retrainResource?.vcpu">
              Profil Server: <strong>{{ retrainResource.vcpu }} vCPU / {{ Math.round((retrainResource.ram_total_mb || 8192) / 1024) }} GB RAM</strong>
            </span>
          </div>

          <!-- Progress bar while training -->
          <div v-if="currentModel.state === 'training'" class="w-full max-w-md pt-1">
            <div class="h-1.5 w-full rounded-full bg-sky-200/70 dark:bg-sky-900/50 overflow-hidden">
              <div
                class="h-full bg-sky-500 transition-all duration-300 rounded-full"
                :style="{ width: `${Math.max(8, stateConfig.percent)}%` }"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="showRetrainButton" class="flex items-center gap-2 shrink-0 self-start sm:self-center">
        <UButton
          size="xs"
          :color="currentModel.has_new_data ? 'warning' : 'primary'"
          :variant="currentModel.has_new_data ? 'solid' : 'soft'"
          icon="i-lucide-refresh-cw"
          :loading="isBusy || checkingRetrainStatus"
          :disabled="isBusy"
          class="font-semibold rounded-lg"
          @click="handleTriggerModelRetrain"
        >
          {{ isBusy ? stateConfig.badgeText : buttonLabel }}
        </UButton>
      </div>
    </div>
  </div>
</template>
