<template>
  <div class="p-6 sm:p-8 space-y-5">
    <!-- Job header: slug, mode, overall status, connection -->
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div class="space-y-1 min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white font-mono truncate">
            {{ job?.slug || '…' }}
          </h2>
          <UBadge
            :color="statusColor(job?.status)"
            variant="subtle"
            size="sm"
          >
            <UIcon
              :name="statusIcon(job?.status)"
              :class="['w-3.5 h-3.5', job?.status === 'running' ? 'animate-spin' : '']"
            />
            {{ job ? t(`siteGenerator.job.status.${job.status}`) : t('siteGenerator.job.loading') }}
          </UBadge>
          <UBadge
            v-if="job?.mode"
            color="neutral"
            variant="outline"
            size="sm"
          >
            {{ job.mode === 'local' ? t('siteGenerator.organization.localhost') : t('siteGenerator.organization.production') }}
          </UBadge>
          <UBadge
            v-if="job?.dry_run"
            color="warning"
            variant="solid"
            size="sm"
          >
            <UIcon
              name="i-lucide-flask-conical"
              class="w-3.5 h-3.5"
            />
            {{ t('siteGenerator.job.dryRunBadge') }}
          </UBadge>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 font-mono">
          {{ t('siteGenerator.job.jobId', { id: jobId }) }}
          <template v-if="job?.created_at">
            · {{ formatDate(job.created_at) }}
          </template>
          <template v-if="job?.created_by">
            · {{ job.created_by }}
          </template>
          <template v-if="totalElapsed">
            · {{ t('siteGenerator.job.elapsed', { time: totalElapsed }) }}
          </template>
        </p>
      </div>

      <span
        class="text-[11px] font-medium px-2 py-1 rounded-full border flex items-center gap-1.5 self-start shrink-0"
        :class="connectionClass"
      >
        <span
          class="w-1.5 h-1.5 rounded-full"
          :class="state.connection === 'streaming' ? 'bg-emerald-500 animate-pulse' : state.connection === 'closed' ? 'bg-gray-400' : 'bg-amber-500 animate-pulse'"
        />
        {{ t(`siteGenerator.job.connection.${state.connection}`) }}
      </span>
    </div>

    <!-- Dry run banner -->
    <div
      v-if="job?.dry_run"
      class="p-3 rounded-xl bg-amber-50 dark:bg-amber-950/40 border border-amber-300 dark:border-amber-800 text-xs text-amber-900 dark:text-amber-200 flex items-start gap-2"
    >
      <UIcon
        name="i-lucide-flask-conical"
        class="w-4 h-4 shrink-0 mt-0.5"
      />
      <span>{{ t('siteGenerator.job.dryRunBanner') }}</span>
    </div>

    <!-- Load / connection errors -->
    <div
      v-if="state.error"
      class="p-3 rounded-xl bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/60 text-xs text-red-800 dark:text-red-300 flex items-start gap-2"
    >
      <UIcon
        name="i-lucide-wifi-off"
        class="w-4 h-4 shrink-0 mt-0.5"
      />
      <div class="space-y-0.5">
        <p class="font-semibold">
          {{ t(`siteGenerator.errors.${state.error.kind}`) }}
        </p>
        <p v-if="state.error.message">
          {{ state.error.message }}
        </p>
        <p v-if="state.connection !== 'closed'">
          {{ t('siteGenerator.job.retrying') }}
        </p>
      </div>
    </div>

    <!-- Queued hint -->
    <div
      v-if="job?.status === 'queued'"
      class="p-3 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40 text-xs text-blue-800 dark:text-blue-300 flex items-center gap-2"
    >
      <UIcon
        name="i-lucide-hourglass"
        class="w-4 h-4 shrink-0"
      />
      {{ t('siteGenerator.job.queuedHint') }}
    </div>

    <!-- Steps -->
    <ol
      v-if="job"
      class="rounded-xl border border-gray-200 dark:border-neutral-800 divide-y divide-gray-100 dark:divide-neutral-800 overflow-hidden"
    >
      <li
        v-for="step in job.steps"
        :key="step.key"
        class="px-4 py-3 flex items-start gap-3 text-xs"
        :class="step.status === 'running' ? 'bg-primary-50/60 dark:bg-primary-950/20' : step.status === 'failed' ? 'bg-red-50/70 dark:bg-red-950/20' : ''"
      >
        <div
          class="w-6 h-6 rounded-full flex items-center justify-center shrink-0"
          :class="isStepWarning(step) ? 'bg-amber-500 text-white' : stepBadgeClass(step.status)"
        >
          <UIcon
            :name="isStepWarning(step) ? 'i-lucide-triangle-alert' : stepIcon(step.status)"
            :class="['w-3.5 h-3.5', step.status === 'running' ? 'animate-spin' : '']"
          />
        </div>
        <div class="flex-1 min-w-0 space-y-0.5">
          <div class="flex items-center justify-between gap-2">
            <span class="font-semibold text-gray-900 dark:text-white">
              {{ step.index }}. {{ stepTitle(step) }}
            </span>
            <span class="flex items-center gap-2 shrink-0">
              <span
                v-if="stepElapsed(step)"
                class="font-mono text-gray-500 dark:text-gray-400"
              >{{ stepElapsed(step) }}</span>
              <span
                class="text-[11px] font-medium"
                :class="isStepWarning(step) ? 'text-amber-600 dark:text-amber-400' : stepTextClass(step.status)"
              >{{ isStepWarning(step) ? t('siteGenerator.job.stepWarning') : t(`siteGenerator.job.stepStatus.${step.status}`) }}</span>
            </span>
          </div>
          <p
            v-if="step.message"
            class="text-gray-600 dark:text-gray-400 break-words"
            :class="step.status === 'failed' ? 'text-red-700 dark:text-red-300 font-medium' : isStepWarning(step) ? 'text-amber-700 dark:text-amber-300 font-medium' : ''"
          >
            {{ step.message }}
          </p>
        </div>
      </li>
    </ol>
    <div
      v-else-if="!state.error"
      class="py-10 flex items-center justify-center gap-2 text-xs text-gray-500"
    >
      <UIcon
        name="i-lucide-loader-2"
        class="w-4 h-4 animate-spin"
      />
      {{ t('siteGenerator.job.loading') }}
    </div>

    <!-- Success -->
    <div
      v-if="job?.status === 'succeeded'"
      class="rounded-xl border border-emerald-300 dark:border-emerald-800 bg-emerald-50/60 dark:bg-emerald-950/20 p-4 space-y-3 text-xs"
    >
      <div class="flex items-center gap-2 text-emerald-800 dark:text-emerald-300 font-bold text-sm">
        <UIcon
          name="i-lucide-check-circle"
          class="w-5 h-5"
        />
        {{ job.dry_run ? t('siteGenerator.success.titleDryRun') : t('siteGenerator.success.title') }}
      </div>

      <!-- allow_no_tls: step 9 was skipped, the tenant is live on plain HTTP -->
      <div
        v-if="noTlsFallback"
        class="p-3 rounded-lg border border-amber-300 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/40 text-amber-900 dark:text-amber-200 flex items-start gap-2"
        role="alert"
      >
        <UIcon
          name="i-lucide-shield-alert"
          class="w-4 h-4 shrink-0 mt-0.5"
        />
        <div class="space-y-1 min-w-0">
          <p class="font-bold">
            {{ t('siteGenerator.success.noTlsTitle') }}
          </p>
          <p>{{ t('siteGenerator.success.noTlsBody') }}</p>
          <p
            v-if="ingressMessage"
            class="font-mono text-[11px] break-all"
          >
            {{ ingressMessage }}
          </p>
        </div>
      </div>

      <dl class="grid grid-cols-1 sm:grid-cols-[max-content_1fr] gap-x-4 gap-y-2">
        <template v-if="job.result?.frontend_url">
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('siteGenerator.success.frontendUrl') }}
          </dt>
          <dd class="font-mono break-all">
            <a
              :href="job.result.frontend_url"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary font-semibold hover:underline inline-flex items-center gap-1"
            >
              {{ job.result.frontend_url }}
              <UIcon
                name="i-lucide-external-link"
                class="w-3.5 h-3.5"
              />
            </a>
          </dd>
        </template>
        <template v-if="job.result?.api_url">
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('siteGenerator.success.apiUrl') }}
          </dt>
          <dd class="font-mono break-all text-gray-800 dark:text-gray-200">
            {{ job.result.api_url }}
          </dd>
        </template>
        <template v-if="job.result?.admin_username">
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('siteGenerator.success.adminUsername') }}
          </dt>
          <dd class="font-mono text-gray-800 dark:text-gray-200">
            {{ job.result.admin_username }}
          </dd>
        </template>
        <template v-if="job.result?.admin_email">
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('siteGenerator.success.adminEmail') }}
          </dt>
          <dd class="font-mono text-gray-800 dark:text-gray-200 break-all">
            {{ job.result.admin_email }}
          </dd>
        </template>
      </dl>

      <!-- Initial admin password: shown once -->
      <div
        v-if="state.initialPassword"
        class="rounded-lg border border-amber-300 dark:border-amber-800 bg-white dark:bg-neutral-900 p-3 space-y-2"
      >
        <div class="flex items-center justify-between gap-2">
          <span class="font-semibold text-gray-800 dark:text-gray-200">{{ t('siteGenerator.success.password') }}</span>
          <div class="flex items-center gap-1.5">
            <UButton
              color="neutral"
              variant="ghost"
              size="xs"
              :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
              :aria-label="showPassword ? t('siteGenerator.success.hide') : t('siteGenerator.success.reveal')"
              @click="showPassword = !showPassword"
            />
            <UButton
              color="neutral"
              variant="outline"
              size="xs"
              icon="i-lucide-copy"
              @click="emit('copy', state.initialPassword, t('siteGenerator.success.password'))"
            >
              {{ t('siteGenerator.success.copy') }}
            </UButton>
          </div>
        </div>
        <div class="font-mono text-sm font-bold text-primary break-all select-all">
          {{ showPassword ? state.initialPassword : '•'.repeat(Math.min(state.initialPassword.length, 24)) }}
        </div>
        <p class="text-amber-800 dark:text-amber-300 flex items-start gap-1.5">
          <UIcon
            name="i-lucide-triangle-alert"
            class="w-4 h-4 shrink-0"
          />
          {{ t('siteGenerator.success.passwordWarning') }}
        </p>
      </div>
      <p
        v-else-if="state.connection === 'closed'"
        class="text-gray-600 dark:text-gray-400 flex items-start gap-1.5"
      >
        <UIcon
          name="i-lucide-key-round"
          class="w-4 h-4 shrink-0"
        />
        {{ t('siteGenerator.success.passwordUnavailable') }}
      </p>
    </div>

    <!-- Failure -->
    <div
      v-if="job?.status === 'failed'"
      class="rounded-xl border border-red-300 dark:border-red-900 bg-red-50/70 dark:bg-red-950/20 p-4 space-y-2 text-xs"
    >
      <div class="flex items-center gap-2 text-red-800 dark:text-red-300 font-bold text-sm">
        <UIcon
          name="i-lucide-circle-x"
          class="w-5 h-5"
        />
        {{ t('siteGenerator.failure.title') }}
      </div>
      <p
        v-if="failedStep"
        class="text-gray-800 dark:text-gray-200"
      >
        <span class="text-gray-500 dark:text-gray-400">{{ t('siteGenerator.failure.failedStep') }}</span>
        <span class="font-semibold">{{ failedStep.index }}. {{ stepTitle(failedStep) }}</span>
      </p>
      <p class="font-mono text-red-700 dark:text-red-300 whitespace-pre-wrap break-words">
        {{ failureError || t('siteGenerator.failure.noError') }}
      </p>
      <div class="flex flex-wrap gap-2 pt-1">
        <ReusableButton
          variant="outline"
          color="neutral"
          size="sm"
          icon="i-lucide-scroll-text"
          @click="openLog"
        >
          {{ t('siteGenerator.actions.viewLog') }}
        </ReusableButton>
        <ReusableButton
          variant="fill"
          color="primary"
          size="sm"
          icon="i-lucide-plus"
          @click="emit('start-new')"
        >
          {{ t('siteGenerator.actions.startNew') }}
        </ReusableButton>
      </div>
    </div>

    <!-- Live log tail -->
    <div
      ref="logSection"
      class="rounded-xl border border-gray-200 dark:border-neutral-800 overflow-hidden"
    >
      <button
        type="button"
        class="w-full px-4 py-2.5 flex items-center justify-between text-xs font-semibold text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-neutral-900/80 hover:bg-gray-100 dark:hover:bg-neutral-800/80"
        :aria-expanded="logOpen"
        @click="logOpen = !logOpen"
      >
        <span class="flex items-center gap-1.5">
          <UIcon
            name="i-lucide-terminal"
            class="w-3.5 h-3.5 text-primary"
          />
          {{ t('siteGenerator.job.log') }}
          <span class="font-normal text-gray-400">{{ t('siteGenerator.job.logLines', { count: state.logs.length }) }}</span>
        </span>
        <UIcon
          :name="logOpen ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
          class="w-4 h-4"
        />
      </button>
      <div
        v-if="logOpen"
        ref="logBox"
        class="bg-gray-950 text-gray-200 font-mono text-[11px] leading-relaxed p-3 max-h-80 overflow-auto"
        @scroll="onLogScroll"
      >
        <p
          v-if="!state.logs.length"
          class="text-gray-500"
        >
          {{ t('siteGenerator.job.logEmpty') }}
        </p>
        <div
          v-for="(line, i) in visibleLogs"
          :key="`${line.seq}-${i}`"
          class="whitespace-pre-wrap break-all"
          :class="line.stream === 'stderr' ? 'text-red-300' : ''"
        >
          {{ line.line }}
        </div>
      </div>
    </div>

    <!-- Footer actions (finished jobs) -->
    <div
      v-if="job && job.status === 'succeeded'"
      class="flex justify-end"
    >
      <ReusableButton
        variant="fill"
        color="primary"
        size="sm"
        icon="i-lucide-plus"
        @click="emit('start-new')"
      >
        {{ t('siteGenerator.actions.startNew') }}
      </ReusableButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import ReusableButton from '~/components/shared/ReusableButton.vue'
import {
  formatDuration,
  isNoTlsFallback,
  isStepWarning,
  isTerminalJobStatus,
  stepElapsedMs,
  type JobStatus,
  type ProvisioningStep,
  type StepStatus,
  type WatcherState
} from '~/utils/provisioning'

const props = defineProps<{
  jobId: string
  state: WatcherState
}>()

const emit = defineEmits<{
  (e: 'start-new'): void
  (e: 'copy', text: string, label: string): void
}>()

const { t, locale } = useI18n()

const job = computed(() => props.state.job)
const noTlsFallback = computed(() => isNoTlsFallback(job.value))
// Carries the certbot command to run once DNS is ready ("TLS NOT CONFIGURED ... run: certbot ...")
const ingressMessage = computed(() => job.value?.steps?.find(s => s.key === 'ingress')?.message || '')
const showPassword = ref(false)
const logOpen = ref(false)
const logBox = ref<HTMLElement | null>(null)
const logSection = ref<HTMLElement | null>(null)
const stickToBottom = ref(true)

// Wall clock for the elapsed time of running steps (display only)
const now = ref(Date.now())
let clock: ReturnType<typeof setInterval> | null = null
const stopClock = () => {
  if (clock) clearInterval(clock)
  clock = null
}
watch(() => job.value?.status, (status) => {
  now.value = Date.now()
  if (status && !isTerminalJobStatus(status)) {
    if (!clock) {
      clock = setInterval(() => {
        now.value = Date.now()
      }, 1000)
    }
  } else {
    stopClock()
  }
}, { immediate: true })
onBeforeUnmount(stopClock)

const totalElapsed = computed(() => {
  const j = job.value
  if (!j?.started_at) return ''
  return formatDuration(stepElapsedMs({
    status: isTerminalJobStatus(j.status) ? 'done' : 'running',
    started_at: j.started_at,
    finished_at: j.finished_at
  }, now.value))
})

const stepElapsed = (step: ProvisioningStep) => formatDuration(stepElapsedMs(step, now.value))

const stepTitle = (step: ProvisioningStep) => {
  const key = `siteGenerator.job.steps.${step.key}`
  const label = t(key)
  return label === key ? step.title : label
}

const failedStep = computed(() => {
  const j = job.value
  if (!j) return null
  return j.steps.find(s => s.status === 'failed')
    || j.steps.find(s => s.key === j.current_step)
    || null
})

const failureError = computed(() => job.value?.error || failedStep.value?.message || '')

// Only the tail goes into the DOM; the full buffer stays in state
const visibleLogs = computed(() => props.state.logs.slice(-500))

const formatDate = (iso: string) => {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString(locale.value === 'id' ? 'id-ID' : 'en-GB')
}

const scrollLogToBottom = () => {
  if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
}

const onLogScroll = () => {
  const el = logBox.value
  if (!el) return
  stickToBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight < 24
}

watch(() => props.state.logs.length, async () => {
  if (!logOpen.value || !stickToBottom.value) return
  await nextTick()
  scrollLogToBottom()
})

watch(logOpen, async (open) => {
  if (!open) return
  stickToBottom.value = true
  await nextTick()
  scrollLogToBottom()
})

const openLog = async () => {
  logOpen.value = true
  await nextTick()
  logSection.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

// A failed job opens its log so the error context is visible
watch(() => job.value?.status, (status) => {
  if (status === 'failed') logOpen.value = true
})

onMounted(() => {
  if (job.value?.status === 'failed') logOpen.value = true
})

const statusColor = (s?: JobStatus) => {
  if (s === 'succeeded') return 'success' as const
  if (s === 'failed') return 'error' as const
  if (s === 'running') return 'primary' as const
  return 'neutral' as const
}

const statusIcon = (s?: JobStatus) => {
  if (s === 'succeeded') return 'i-lucide-check-circle'
  if (s === 'failed') return 'i-lucide-circle-x'
  if (s === 'running') return 'i-lucide-loader-2'
  return 'i-lucide-hourglass'
}

const stepIcon = (s: StepStatus) => ({
  pending: 'i-lucide-circle-dashed',
  running: 'i-lucide-loader-2',
  done: 'i-lucide-check',
  skipped: 'i-lucide-skip-forward',
  failed: 'i-lucide-x'
})[s] || 'i-lucide-circle-dashed'

const stepBadgeClass = (s: StepStatus) => ({
  pending: 'bg-gray-100 dark:bg-neutral-800 text-gray-400',
  running: 'bg-primary text-white',
  done: 'bg-emerald-500 text-white',
  skipped: 'bg-gray-200 dark:bg-neutral-700 text-gray-500 dark:text-gray-300',
  failed: 'bg-red-500 text-white'
})[s] || ''

const stepTextClass = (s: StepStatus) => ({
  pending: 'text-gray-400',
  running: 'text-primary',
  done: 'text-emerald-600 dark:text-emerald-400',
  skipped: 'text-gray-500 dark:text-gray-400',
  failed: 'text-red-600 dark:text-red-400'
})[s] || ''

const connectionClass = computed(() => {
  const c = props.state.connection
  if (c === 'streaming') return 'border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40'
  if (c === 'closed' || c === 'idle') return 'border-gray-200 dark:border-neutral-700 text-gray-500 bg-gray-50 dark:bg-neutral-900'
  return 'border-amber-200 dark:border-amber-800 text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40'
})
</script>
