<template>
  <div class="wrapper-vertical-expanded space-y-6">
    <!-- Database registration check panel -->
    <UCard class="w-full border-l-4 border-emerald-500 bg-gradient-to-br from-emerald-50/50 to-green-50/50 dark:from-emerald-950/20 dark:to-green-950/20">
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="flex items-start gap-4">
          <div class="rounded-full bg-emerald-500/20 dark:bg-emerald-500/30 p-2.5 text-emerald-600 dark:text-emerald-400 shrink-0 border border-emerald-200 dark:border-emerald-900/50">
            <UIcon name="i-lucide-shield-check" class="size-7" />
          </div>
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-gray-900 dark:text-white">{{ t('settings.activity.verificationTitle') }}</h4>
            <p class="text-md text-gray-600 dark:text-gray-300">
              {{ t('settings.activity.deviceText') }}
              <span class="font-mono font-semibold bg-slate-200 dark:bg-slate-700 px-2 py-1 rounded-lg text-slate-900 dark:text-slate-100">
                {{ currentDevice.deviceName }} ({{ currentDevice.deviceFingerprint.slice(0, 8) }}...)
              </span>
            </p>
          </div>
        </div>

        <div class="shrink-0 self-start sm:self-center">
          <div v-if="dbStatus.checked" class="px-3.5 py-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/60 text-sm text-emerald-700 dark:text-emerald-400 flex items-center gap-2 font-medium">
            <span class="relative flex h-2.5 w-2.5">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
            </span>
            <span>{{ t('settings.activity.registeredYes', { fingerprint: dbStatus.lastLoginFingerprint }) }}</span>
          </div>
          <div v-else class="px-3.5 py-2 rounded-lg bg-slate-100 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 text-sm text-slate-600 dark:text-slate-400 flex items-center gap-2 animate-pulse">
            <span class="flex h-2.5 w-2.5 rounded-full bg-slate-400"></span>
            <span>{{ t('settings.activity.checkingDb') }}</span>
          </div>
        </div>
      </div>
    </UCard>

    <QuickTip
      :title="t('settings.activity.securityTipsTitle')"
      :description="t('settings.activity.securityTipsDesc')"
      icon="i-lucide-alert-triangle"
      color="amber"
      variant="outline"
    />

    <UCard class="w-full">
      <template #header>
        <div class="flex justify-between items-center">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.activity.trustedDevicesTitle') }}</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('settings.activity.trustedDevicesSubtitle') }}</p>
          </div>
          <UButton
            icon="i-lucide-refresh-cw"
            variant="ghost"
            color="primary"
            :loading="loading"
            @click="loadDevices"
          />
        </div>
      </template>

      <div class="space-y-4 w-full">
        <!-- Current active session device (always shown) -->
        <div class="flex items-stretch gap-4 p-4 rounded-xl border border-cyan-200 dark:border-cyan-900/40 bg-gradient-to-br from-cyan-50/50 to-blue-50/50 dark:from-cyan-950/20 dark:to-blue-950/20 w-full shadow-sm">
          <div class="flex flex-row justify-between items-stretch w-full">
            <div class="flex flex-row gap-4">
              <div class="rounded-full bg-cyan-500/20 dark:bg-cyan-500/30 w-10 h-10 flex items-center justify-center shrink-0 border border-cyan-200 dark:border-cyan-900/50">
                <UIcon :name="currentDevice.deviceType === 'mobile' ? 'i-lucide-smartphone' : 'i-lucide-laptop'" class="text-cyan-600 dark:text-cyan-400 size-5" />
              </div>
              <div class="flex-1 min-w-0 w-full">
                <div class="flex items-center gap-3">
                  <p class="font-medium text-gray-900 dark:text-white">{{ currentDevice.deviceName }}</p>
                  <UBadge :label="t('settings.activity.activeSession')" color="primary" variant="solid" size="md" />
                  <UBadge :label="t('settings.activity.online')" color="success" variant="soft" size="md" />
                </div>
                <h6 class="text-sm flex flex-row gap-2 mt-1 text-gray-600 dark:text-gray-300">
                  <UIcon name="i-lucide-fingerprint" class="size-4 text-gray-400" />
                  Fingerprint: {{ currentDevice.deviceFingerprint }}
                </h6>
                <p class="text-md text-gray-500 mt-1 flex flex-row items-center gap-1.5">
                  <UIcon name="i-lucide-clock" class="size-4" />
                  {{ t('settings.activity.justNow') }}
                </p>
              </div>
            </div>
            <div class="text-md text-gray-400 self-center">
              {{ t('settings.activity.thisDevice') }}
            </div>
          </div>
        </div>

        <!-- Loaded list of other trusted devices -->
        <template v-if="trustedDevices.length > 0">
          <div v-for="device in trustedDevices" :key="device.id" class="flex items-stretch gap-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-gradient-to-br from-slate-50/30 to-gray-50/30 dark:from-slate-900/20 dark:to-gray-900/20 w-full hover:shadow-sm transition-shadow">
            <div class="flex flex-row justify-between items-stretch w-full">
              <div class="flex flex-row gap-4">
                <div class="rounded-full bg-slate-200 dark:bg-slate-700 w-10 h-10 flex items-center justify-center shrink-0 border border-slate-300 dark:border-slate-600">
                  <UIcon :name="device.deviceType === 'mobile' ? 'i-lucide-smartphone' : 'i-lucide-laptop'" class="text-slate-600 dark:text-slate-400 size-5" />
                </div>
                <div class="flex-1 min-w-0 w-full">
                  <div class="flex items-center gap-3">
                    <p class="font-medium text-gray-900 dark:text-white">{{ device.deviceName }}</p>
                    <UBadge v-if="device.deviceFingerprint === currentDevice.deviceFingerprint" :label="t('settings.activity.thisDevice')" color="primary" variant="soft" size="md" />
                  </div>
                  <h6 class="text-sm flex flex-row gap-2 mt-1 text-gray-600 dark:text-gray-300">
                    <UIcon name="i-lucide-globe" class="size-4 text-gray-400" />
                    IP: {{ device.ipAddress }} · Fingerprint: {{ device.deviceFingerprint }}
                  </h6>
                  <p class="text-md text-gray-500 mt-1 flex flex-row items-center gap-1.5">
                    <UIcon name="i-lucide-clock" class="size-4" />
                    {{ t('settings.activity.registeredAt', { date: formatTime(device.createdAt) }) }}
                  </p>
                </div>
              </div>
              <div class="self-center">
                <UButton
                  :label="t('settings.activity.remove')"
                  variant="ghost"
                  color="error"
                  size="sm"
                  @click="confirmRelease(device.id, device.deviceName)"
                />
              </div>
            </div>
          </div>
        </template>
        
        <div v-else-if="!loading" class="text-center py-8 text-sm text-slate-600 dark:text-slate-400 border border-dashed border-slate-200 dark:border-slate-700 rounded-xl bg-slate-50/30 dark:bg-slate-900/20">
          <UIcon name="i-lucide-info" class="size-8 text-slate-400 mx-auto mb-3" />
          <p class="font-medium">{{ t('settings.activity.noDevices') }}</p>
          <p class="text-xs text-slate-500 dark:text-slate-500 mt-2">{{ t('settings.activity.noDevicesSub') }}</p>
        </div>
      </div>
    </UCard>

    <UCard class="w-full" variant="outline" color="neutral">
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.activity.aboutTitle') }}</h3>
      </template>  
      <article class="flex flex-col gap-4 text-sm text-gray-700 dark:text-gray-300">
        <p>{{ t('settings.activity.aboutP1') }}</p>
        <p>{{ t('settings.activity.aboutP2') }}</p>
        <ul class="pb-2 px-6 text-sm space-y-1 list-disc list-inside">
          <li>{{ t('settings.activity.aboutTip1') }}</li>
          <li>{{ t('settings.activity.aboutTip2') }}</li>
          <li>{{ t('settings.activity.aboutTip3') }}</li>
        </ul>
      </article>
    </UCard>

    <!-- Confirmation Modal -->
    <ConfirmationPopup
      v-model:isOpen="state.open"
      :title="t('settings.activity.modalTitle')"
      :question="t('settings.activity.modalQuestion')"
      :confirmText="t('settings.activity.confirmRemove')"
      :cancelText="t('settings.activity.cancelRemove')"
      variant="danger"
      @confirm="handleReleaseConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'
import ConfirmationPopup from '../shared/ConfirmationPopup.vue'
import QuickTip from '../shared/QuickTip.vue'

const authStore = useAuthStore()
const { getDeviceFingerprint } = useDeviceFingerprint()
const toast = useToast()
const { t, locale } = useI18n()

const loading = ref(false)
const trustedDevices = ref<any[]>([])
const currentDevice = ref({ deviceFingerprint: '', deviceName: '', deviceType: '' })

// Database verification status
const dbStatus = reactive({
  checked: false,
  lastLoginFingerprint: '',
  isStored: false,
})

const state = ref({
  open: false,
  targetDeviceId: '',
  targetDeviceName: '',
})

onMounted(async () => {
  currentDevice.value = getDeviceFingerprint()
  await loadDevices()
  await verifyDatabaseRegistration()
})

const loadDevices = async () => {
  loading.value = true
  try {
    const list = await authStore.fetchTrustedDevices()
    // Filter out null values if list has empty spots
    trustedDevices.value = (list || []).filter((d: any) => d && d.id)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const verifyDatabaseRegistration = async () => {
  try {
    const profile = await authStore.fetchUserProfile()
    if (profile) {
      dbStatus.lastLoginFingerprint = profile.last_login_fingerprint || ''
      dbStatus.isStored = (profile.last_login_fingerprint === currentDevice.value.deviceFingerprint)
      dbStatus.checked = true
    }
  } catch (err) {
    console.error(err)
    dbStatus.checked = true
  }
}

const confirmRelease = (deviceId: string, deviceName: string) => {
  state.value.targetDeviceId = deviceId
  state.value.targetDeviceName = deviceName
  state.value.open = true
}

const handleReleaseConfirm = async () => {
  state.value.open = false
  try {
    await authStore.unenrollDevice(state.value.targetDeviceId)
    toast.add({
      title: t('settings.activity.deviceRemovedToast'),
      description: t('settings.activity.deviceRemovedDesc', { name: state.value.targetDeviceName }),
      color: "success",
      icon: "i-lucide-circle-check",
    })
    await loadDevices()
  } catch (err: any) {
    toast.add({
      title: "Gagal Menghapus",
      description: err.message || "Gagal menghapus perangkat terpercaya.",
      color: "error",
      icon: "i-lucide-alert-triangle",
    })
  }
}

const formatTime = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat(locale.value === 'id' ? 'id-ID' : 'en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
</script>