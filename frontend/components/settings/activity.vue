<template>
  <div class="space-y-6 w-full">
    <!-- Devices -->
    <UCard
      variant="outline"
      color="neutral"
      class="w-full overflow-hidden shadow-sm rounded-2xl dark:bg-slate-900 dark:ring-slate-800"
    >
      <template #header>
        <div class="flex items-start sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 bg-secondary-500/10 text-secondary-600 dark:text-secondary-400 border border-secondary-200 dark:border-secondary-900/50">
              <UIcon
                name="i-lucide-monitor-smartphone"
                class="w-6 h-6"
              />
            </div>
            <div class="min-w-0">
              <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate">
                {{ t('settings.activity.devicesTitle') }}
              </h3>
              <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 break-words">
                {{ t('settings.activity.devicesSubtitle') }}
              </p>
            </div>
          </div>
          <UButton
            icon="i-lucide-refresh-cw"
            color="neutral"
            variant="ghost"
            class="shrink-0"
            :loading="refreshing"
            :disabled="refreshing"
            :aria-label="t('settings.activity.refresh')"
            :title="t('settings.activity.refresh')"
            @click="refreshAll"
          />
        </div>
      </template>

      <!-- First load -->
      <div
        v-if="devicesLoading && !devicesLoaded"
        class="divide-y divide-gray-100 dark:divide-gray-800/70"
      >
        <div
          v-for="i in 2"
          :key="i"
          class="flex items-center gap-3.5 py-4 first:pt-0 last:pb-0"
        >
          <USkeleton class="size-10 rounded-xl shrink-0" />
          <div class="flex-1 space-y-2">
            <USkeleton class="h-4 w-48" />
            <USkeleton class="h-3 w-72 max-w-full" />
          </div>
        </div>
      </div>

      <div
        v-else
        class="space-y-4"
      >
        <UAlert
          v-if="devicesError"
          color="error"
          variant="subtle"
          icon="i-lucide-alert-triangle"
          :title="t('settings.activity.devicesErrorTitle')"
          :description="devicesError"
          :actions="[{ label: t('settings.activity.retry'), color: 'neutral', variant: 'outline', icon: 'i-lucide-refresh-cw', onClick: loadDevices }]"
        />

        <ul class="divide-y divide-gray-100 dark:divide-gray-800/70">
          <li
            v-for="row in deviceRows"
            :key="row.key"
            class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 py-4 first:pt-0 last:pb-0"
          >
            <div class="flex items-start gap-3.5 min-w-0 flex-1">
              <div class="size-10 rounded-xl flex items-center justify-center shrink-0 bg-secondary-500/10 text-secondary-600 dark:bg-secondary-400/10 dark:text-secondary-400 ring-1 ring-inset ring-secondary-500/20">
                <UIcon
                  :name="deviceIcon(row.type)"
                  class="size-5"
                />
              </div>
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">
                    {{ row.title }}
                  </p>
                  <UBadge
                    v-if="row.isCurrent"
                    :label="t('settings.activity.thisDevice')"
                    color="secondary"
                    variant="soft"
                    size="sm"
                  />
                  <UBadge
                    v-if="row.trusted"
                    :label="t('settings.activity.trusted')"
                    color="success"
                    variant="soft"
                    size="sm"
                  />
                  <UBadge
                    v-else
                    :label="t('settings.activity.notTrusted')"
                    color="neutral"
                    variant="outline"
                    size="sm"
                  />
                </div>
                <dl class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-600 dark:text-gray-400">
                  <div
                    v-if="row.os"
                    class="flex items-center gap-1.5"
                  >
                    <dt class="sr-only">
                      {{ t('settings.activity.osLabel') }}
                    </dt>
                    <UIcon
                      name="i-lucide-cpu"
                      class="size-3.5 shrink-0 text-gray-400"
                    />
                    <dd>{{ row.os }}</dd>
                  </div>
                  <div
                    v-if="row.ipAddress"
                    class="flex items-center gap-1.5"
                  >
                    <dt class="sr-only">
                      {{ t('settings.activity.ipLabel') }}
                    </dt>
                    <UIcon
                      name="i-lucide-globe"
                      class="size-3.5 shrink-0 text-gray-400"
                    />
                    <dd class="font-mono">
                      {{ row.ipAddress }}
                    </dd>
                  </div>
                  <div
                    v-if="row.location"
                    class="flex items-center gap-1.5"
                  >
                    <dt class="sr-only">
                      {{ t('settings.activity.locationLabel') }}
                    </dt>
                    <UIcon
                      name="i-lucide-map-pin"
                      class="size-3.5 shrink-0 text-gray-400"
                    />
                    <dd>{{ row.location }}</dd>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <dt class="sr-only">
                      {{ t('settings.activity.lastActiveLabel') }}
                    </dt>
                    <UIcon
                      name="i-lucide-clock"
                      class="size-3.5 shrink-0 text-gray-400"
                    />
                    <dd :title="row.activityTitle">
                      {{ row.activity }}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
            <UButton
              v-if="row.removableId"
              :label="t('settings.activity.remove')"
              color="error"
              variant="ghost"
              size="sm"
              icon="i-lucide-trash-2"
              class="shrink-0 self-end sm:self-center"
              :loading="removingId === row.removableId"
              :disabled="!!removingId"
              @click="confirmRemove(row.removableId, row.title)"
            />
          </li>
        </ul>

        <p
          v-if="!devicesError && !hasTrustedDevice"
          class="flex items-start gap-2 text-xs text-gray-500 dark:text-gray-400 rounded-xl border border-dashed border-gray-200 dark:border-gray-800 px-3 py-2.5"
        >
          <UIcon
            name="i-lucide-info"
            class="size-4 shrink-0 mt-px"
          />
          <span>{{ t('settings.activity.noTrustedDevices') }}</span>
        </p>
      </div>
    </UCard>

    <!-- Recent activity -->
    <UCard
      variant="outline"
      color="neutral"
      class="w-full overflow-hidden shadow-sm rounded-2xl dark:bg-slate-900 dark:ring-slate-800"
    >
      <template #header>
        <div class="flex items-start sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300 border border-gray-200 dark:border-gray-700">
              <UIcon
                name="i-lucide-history"
                class="w-6 h-6"
              />
            </div>
            <div class="min-w-0">
              <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate">
                {{ t('settings.activity.recentTitle') }}
              </h3>
              <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 break-words">
                {{ t('settings.activity.recentSubtitle') }}
              </p>
            </div>
          </div>
          <UButton
            v-if="canViewSystemLogs"
            :label="t('settings.activity.openSystemLogs')"
            trailing-icon="i-lucide-arrow-right"
            color="neutral"
            variant="outline"
            size="sm"
            class="shrink-0"
            :to="{ query: { ...route.query, tab: 'systemlogs' } }"
          />
        </div>
      </template>

      <div
        v-if="activityLoading"
        class="space-y-4"
      >
        <div
          v-for="i in 3"
          :key="i"
          class="flex items-center gap-3"
        >
          <USkeleton class="size-8 rounded-full shrink-0" />
          <div class="flex-1 space-y-2">
            <USkeleton class="h-3.5 w-40" />
            <USkeleton class="h-3 w-56 max-w-full" />
          </div>
        </div>
      </div>

      <UAlert
        v-else-if="activityError"
        color="error"
        variant="subtle"
        icon="i-lucide-alert-triangle"
        :title="t('settings.activity.recentErrorTitle')"
        :description="t('settings.activity.recentErrorDesc')"
        :actions="[{ label: t('settings.activity.retry'), color: 'neutral', variant: 'outline', icon: 'i-lucide-refresh-cw', onClick: refreshAll }]"
      />

      <div
        v-else-if="activityItems.length === 0"
        class="text-center py-8"
      >
        <UIcon
          name="i-lucide-inbox"
          class="size-8 text-gray-400 mx-auto mb-2"
        />
        <p class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('settings.activity.recentEmpty') }}
        </p>
      </div>

      <ol
        v-else
        class="divide-y divide-gray-100 dark:divide-gray-800/70"
      >
        <li
          v-for="item in activityItems"
          :key="item.key"
          class="flex items-start gap-3 py-3 first:pt-0 last:pb-0"
        >
          <div class="size-8 rounded-full flex items-center justify-center shrink-0 bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300">
            <UIcon
              :name="item.icon"
              class="size-4"
            />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ item.label }}
            </p>
            <p
              v-if="item.detail"
              class="text-xs text-gray-600 dark:text-gray-400 truncate"
            >
              {{ item.detail }}
            </p>
          </div>
          <time
            class="text-xs text-gray-500 dark:text-gray-400 shrink-0 whitespace-nowrap"
            :datetime="item.at"
            :title="formatAbsoluteTime(item.at, locale)"
          >{{ formatRelativeTime(item.at, now, locale) }}</time>
        </li>
      </ol>
    </UCard>

    <QuickTip
      :title="t('settings.activity.securityTipsTitle')"
      :description="t('settings.activity.securityTipsDesc')"
      icon="i-lucide-lightbulb"
      color="warning"
      variant="subtle"
    />
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'
import { useSystemLogAccess } from '~/composables/useSystemLogs'
import { useGlobalModalStore } from '~/stores/global-modal'
import { extractErrorMessage } from '~/utils/error'
import { formatAbsoluteTime, formatRelativeTime } from '~/utils/systemLogFormat'
import { describeDevice, deviceIcon, type TrustedDeviceView } from '~/utils/trustedDevices'
import QuickTip from '../shared/QuickTip.vue'

const authStore = useAuthStore()
const route = useRoute()
const { getDeviceFingerprint } = useDeviceFingerprint()
const { canViewSystemLogs } = useSystemLogAccess()
const toast = useToast()
const { t, locale } = useI18n()

const currentDevice = getDeviceFingerprint()
const currentUserAgent = import.meta.client ? navigator.userAgent : ''

// The list lives in the auth store so the MFA form can refresh it after MFA is enabled
const devices = computed(() => authStore.trustedDevices)
const devicesLoading = ref(false)
const devicesLoaded = ref(false)
const devicesError = ref('')
const removingId = ref('')

// Last sign-in from the user record (users.last_login_at / last_login_fingerprint)
const lastLogin = ref<{ at: string, fingerprint: string } | null>(null)
const profileLoading = ref(false)
const profileFailed = ref(false)

const now = ref(Date.now())
let clock: ReturnType<typeof setInterval> | null = null

const refreshing = computed(() => devicesLoading.value || profileLoading.value)

const formatDate = (iso: string) => formatAbsoluteTime(iso, locale.value)

const loadDevices = async () => {
  devicesLoading.value = true
  devicesError.value = ''
  try {
    await authStore.fetchTrustedDevices()
  } catch (err) {
    devicesError.value = extractErrorMessage(err, t('settings.activity.devicesErrorDesc'))
  } finally {
    devicesLoading.value = false
    devicesLoaded.value = true
  }
}

const loadProfile = async () => {
  profileLoading.value = true
  profileFailed.value = false
  try {
    // fetchUserProfile returns null when the request fails
    const profile = await authStore.fetchUserProfile()
    if (!profile) {
      profileFailed.value = true
      return
    }
    lastLogin.value = profile.last_login_at
      ? { at: String(profile.last_login_at), fingerprint: String(profile.last_login_fingerprint || '') }
      : null
  } finally {
    profileLoading.value = false
  }
}

const refreshAll = async () => {
  now.value = Date.now()
  await Promise.all([loadDevices(), loadProfile()])
}

interface DeviceRow {
  key: string
  /** Set for trusted devices: removing one revokes its MFA skip */
  removableId: string
  title: string
  os: string
  type: string
  ipAddress: string
  location: string
  activity: string
  activityTitle: string
  isCurrent: boolean
  trusted: boolean
}

const deviceTitle = (browser: string, os: string, fallback: string) =>
  browser && os ? t('settings.activity.browserOnOs', { browser, os }) : browser || fallback || t('settings.activity.unknownDevice')

const titleOf = (d: TrustedDeviceView) => {
  const desc = describeDevice(d)
  return deviceTitle(desc.browser, desc.os, d.name)
}

/** "Last active" for a listed device: last_seen_at, else the user's last sign-in if it came from it, else when it was added. */
const lastActivity = (d: TrustedDeviceView) => {
  const lastSignIn = lastLogin.value
  if (d.lastSeenAt) return { key: 'settings.activity.lastActive', at: d.lastSeenAt }
  if (lastSignIn && d.fingerprint && d.fingerprint === lastSignIn.fingerprint) return { key: 'settings.activity.lastSignIn', at: lastSignIn.at }
  if (d.createdAt) return { key: d.trusted ? 'settings.activity.trustedOn' : 'settings.activity.firstSeen', at: d.createdAt }
  return null
}

/** This browser first (marked "This device"), then the other devices, most recently active first. */
const deviceRows = computed<DeviceRow[]>(() => {
  const rows: DeviceRow[] = []
  const current = devices.value.find(d => d.fingerprint && d.fingerprint === currentDevice.deviceFingerprint)

  const currentDesc = describeDevice({ userAgent: currentUserAgent, name: currentDevice.deviceName, type: currentDevice.deviceType })
  rows.push({
    key: 'current',
    removableId: current?.trusted ? current.id : '',
    title: deviceTitle(currentDesc.browser, currentDesc.os, currentDevice.deviceName),
    os: currentDesc.os,
    type: currentDesc.type,
    ipAddress: current?.ipAddress ?? '',
    location: current?.location ?? '',
    activity: t('settings.activity.activeNow'),
    activityTitle: '',
    isCurrent: true,
    trusted: !!current?.trusted
  })

  for (const d of devices.value) {
    if (d === current) continue
    const desc = describeDevice(d)
    const last = lastActivity(d)
    rows.push({
      key: d.id,
      removableId: d.trusted ? d.id : '',
      title: deviceTitle(desc.browser, desc.os, d.name),
      os: desc.os,
      type: desc.type,
      ipAddress: d.ipAddress,
      location: d.location,
      activity: last ? t(last.key, { date: formatDate(last.at) }) : '-',
      activityTitle: last ? formatDate(last.at) : '',
      isCurrent: false,
      trusted: d.trusted
    })
  }
  return rows
})

const hasTrustedDevice = computed(() => devices.value.some(d => d.trusted))

interface ActivityItem {
  key: string
  icon: string
  label: string
  detail: string
  at: string
}

/**
 * What auth-service exposes to every user about their own account: the last sign-in and when each device was
 * first seen or trusted. (The full event log is GET /system-logs, which needs view_system_logs and records a
 * "viewed" event on every load, so it stays on the System Logs tab.)
 */
const activityItems = computed<ActivityItem[]>(() => {
  const items: ActivityItem[] = []
  if (lastLogin.value) {
    const fp = lastLogin.value.fingerprint
    const device = fp ? devices.value.find(d => d.fingerprint === fp) : undefined
    let detail = ''
    if (fp && fp === currentDevice.deviceFingerprint) detail = t('settings.activity.onThisDevice')
    else if (device) detail = titleOf(device)
    items.push({ key: 'last-login', icon: 'i-lucide-log-in', label: t('settings.activity.eventLastSignIn'), detail, at: lastLogin.value.at })
  }
  for (const d of devices.value) {
    if (!d.createdAt) continue
    const name = titleOf(d)
    items.push({
      key: `device-${d.id}`,
      icon: d.trusted ? 'i-lucide-shield-check' : 'i-lucide-monitor-smartphone',
      label: t(d.trusted ? 'settings.activity.eventDeviceTrusted' : 'settings.activity.eventNewDevice'),
      detail: d.ipAddress ? `${name} · ${d.ipAddress}` : name,
      at: d.createdAt
    })
  }
  return items
    .sort((a, b) => (Date.parse(b.at) || 0) - (Date.parse(a.at) || 0))
    .slice(0, 10)
})

const activityLoading = computed(() => (profileLoading.value && !lastLogin.value) || (devicesLoading.value && !devicesLoaded.value))
// Only the sign-in record is required; device events are an addition (their error shows on the card above)
const activityError = computed(() => profileFailed.value && activityItems.value.length === 0)

const confirmRemove = async (deviceId: string, name: string) => {
  const ok = await useGlobalModalStore().confirmDelete({
    description: t('settings.activity.modalTitle'),
    body: [t('settings.activity.modalQuestion')],
    confirmLabel: t('settings.activity.confirmRemove')
  })
  if (!ok) return
  removingId.value = deviceId
  try {
    await authStore.unenrollDevice(deviceId)
    toast.add({
      title: t('settings.activity.deviceRemovedToast'),
      description: t('settings.activity.deviceRemovedDesc', { name }),
      color: 'success',
      icon: 'i-lucide-circle-check'
    })
    await loadDevices()
  } catch (err) {
    toast.add({
      title: t('settings.activity.removeFailedTitle'),
      description: extractErrorMessage(err, t('settings.activity.removeFailedDesc')),
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    removingId.value = ''
  }
}

onMounted(() => {
  void refreshAll()
  clock = setInterval(() => {
    now.value = Date.now()
  }, 60_000)
})

onBeforeUnmount(() => {
  if (clock) clearInterval(clock)
})
</script>
