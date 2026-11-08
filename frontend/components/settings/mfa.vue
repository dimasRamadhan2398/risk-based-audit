<template>
  <div class="space-y-6 w-full">
    <UCard variant="outline" color="neutral" class="w-full overflow-hidden shadow-sm rounded-2xl dark:bg-slate-900 dark:ring-slate-800">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <div
              class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 transition-all duration-300 shadow-sm"
              :class="isMfaEnabled
                ? (isDisabling ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-900/50' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-900/50')
                : 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-900/50'"
            >
              <UIcon
                :name="isMfaEnabled ? (isDisabling ? 'i-lucide-shield-alert' : 'i-lucide-shield-check') : 'i-lucide-shield-alert'"
                class="w-6 h-6"
              />
            </div>
            <div class="min-w-0">
              <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate">
                {{ isDisabling ? t('settings.mfa.modalTitle') : t('settings.mfa.title') }}
              </h3>
              <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 break-words">
                {{ isDisabling ? t('settings.mfa.modalSubtitle') : t('settings.mfa.subtitle') }}
              </p>
            </div>
          </div>

          <div v-if="!loadingStatus && !statusError" class="self-start sm:self-center shrink-0">
            <UBadge
              :color="isMfaEnabled ? (isDisabling ? 'error' : 'success') : 'warning'"
              variant="soft"
              size="md"
              class="font-bold px-3.5 py-1.5 rounded-full"
            >
              <template #leading>
                <span
                  class="w-2 h-2 rounded-full"
                  :class="isMfaEnabled
                    ? (isDisabling ? 'bg-rose-500 animate-pulse' : 'bg-emerald-500 animate-pulse')
                    : 'bg-amber-500 animate-pulse'"
                />
              </template>
              {{ isMfaEnabled ? (isDisabling ? t('settings.mfa.confirmDisable') : t('settings.mfa.activeProtected')) : t('settings.mfa.notEnabled') }}
            </UBadge>
          </div>
          <USkeleton v-else-if="loadingStatus" class="h-7 w-28 rounded-full" />
        </div>
      </template>

      <!-- Loading State -->
      <div v-if="loadingStatus" class="p-6 space-y-4">
        <USkeleton class="h-4 w-3/4" />
        <USkeleton class="h-10 w-48 rounded-xl" />
      </div>

      <!-- Status could not be loaded: don't offer setup on a guess -->
      <UAlert
        v-else-if="statusError"
        color="error"
        variant="subtle"
        icon="i-lucide-alert-triangle"
        :title="t('settings.mfa.statusFailedTitle')"
        :description="statusError"
        :actions="[{ label: t('settings.mfa.retry'), color: 'neutral', variant: 'outline', icon: 'i-lucide-refresh-cw', onClick: () => fetchStatus() }]"
      />

      <!-- State 1 & 2: MFA Disabled Flow (Begin Component / Setup Stepper) -->
      <div v-else-if="!isMfaEnabled" class="space-y-6 p-2">
        <div class="p-4 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40 text-amber-900 dark:text-amber-200 text-sm flex items-start gap-3">
          <UIcon name="i-lucide-info" class="w-5 h-5 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
          <p>
            {{ t('settings.mfa.infoText') }}
          </p>
        </div>

        <!-- MFA Begin Component (Start Setup) -->
        <div v-if="!setupData" class="pt-2">
          <UButton
            icon="i-lucide-shield-plus"
            color="primary"
            size="lg"
            class="font-bold px-6 py-3 rounded-xl transition-all duration-200 shadow-md cursor-pointer"
            :loading="loading"
            :disabled="loading"
            @click="handleSetup"
          >
            {{ t('settings.mfa.enableButton') }}
          </UButton>
        </div>

        <!-- MFA Setup Stepper Component -->
        <div v-else class="space-y-6 border border-amber-200 dark:border-amber-900/30 rounded-2xl p-6 bg-gradient-to-br from-amber-50/50 to-yellow-50/50 dark:from-amber-950/20 dark:to-yellow-950/10">
          <div class="flex flex-col lg:flex-row gap-8 items-start">
            <!-- QR Code Box -->
            <div class="flex flex-col items-center gap-2 bg-white dark:bg-slate-800 p-4 rounded-2xl border border-amber-200 dark:border-amber-900/40 shadow-sm shrink-0">
              <img v-if="qrCodeDataURL" :src="qrCodeDataURL" :alt="t('settings.mfa.qrAlt')" class="w-44 h-44 rounded-lg border border-gray-200 dark:border-gray-700" />
              <div v-else class="w-44 h-44 flex items-center justify-center text-gray-400 text-sm animate-pulse">
                {{ t('settings.mfa.generatingQr') }}
              </div>
              <span class="text-md text-gray-600 dark:text-gray-300 font-medium">{{ t('settings.mfa.scanQr') }}</span>
            </div>

            <!-- Steps -->
            <div class="space-y-4 flex-1">
              <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <span class="w-6 h-6 rounded-full bg-amber-500 text-white text-md flex items-center justify-center font-bold">1</span>
                {{ t('settings.mfa.scanTitle') }}
              </h4>
              <p class="text-sm text-gray-600 dark:text-gray-300">
                {{ t('settings.mfa.scanInstructions') }}
              </p>

              <div class="flex items-center gap-2 max-w-md">
                <code class="bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded-xl font-mono text-sm font-semibold text-primary tracking-widest border border-gray-200 dark:border-gray-700 flex-1 truncate select-all">
                  {{ setupData.secret }}
                </code>
                <UButton
                  icon="i-lucide-copy"
                  color="neutral"
                  variant="subtle"
                  class="rounded-xl"
                  :title="t('settings.mfa.copySecret')"
                  @click="copySecret"
                />
              </div>

              <div class="pt-4 border-t border-gray-200 dark:border-gray-800 space-y-4">
                <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                  <span class="w-6 h-6 rounded-full bg-primary text-white text-md flex items-center justify-center font-bold">2</span>
                  {{ t('settings.mfa.verifyTitle') }}
                </h4>

                <!-- Enter and the submit button both submit the form; handleVerifySetup drops overlapping calls -->
                <form class="max-w-md space-y-3" novalidate @submit.prevent="handleVerifySetup">
                  <UAlert
                    v-if="setupChanged"
                    color="warning"
                    variant="subtle"
                    icon="i-lucide-qr-code"
                    :title="t('settings.mfa.setupChangedTitle')"
                    :description="t('settings.mfa.setupChangedDesc')"
                    :actions="[{ label: t('settings.mfa.newQrButton'), color: 'warning', variant: 'solid', icon: 'i-lucide-refresh-cw', loading, disabled: loading, onClick: () => handleSetup() }]"
                  />
                  <UFormField :label="t('settings.mfa.verificationCodeLabel')" :help="verifying ? t('settings.mfa.verifyingCode') : t('settings.mfa.verificationCodeHelp')">
                    <UInput
                      v-model="verificationCode"
                      placeholder="000000"
                      inputmode="numeric"
                      autocomplete="one-time-code"
                      size="lg"
                      :disabled="verifying"
                      class="w-full font-mono text-center text-xl tracking-[0.3em] font-bold"
                    />
                  </UFormField>

                  <div class="flex gap-3 pt-2">
                    <UButton
                      type="button"
                      variant="subtle"
                      color="neutral"
                      class="flex-1 rounded-xl font-semibold"
                      :disabled="verifying"
                      @click="cancelSetup"
                    >
                      {{ t('settings.mfa.cancel') }}
                    </UButton>
                    <UButton
                      type="submit"
                      color="primary"
                      class="flex-1 rounded-xl font-bold transition-all duration-200 shadow-md"
                      :loading="verifying"
                      :disabled="verifying || !canSubmitCode"
                    >
                      {{ t('settings.mfa.verifyEnableButton') }}
                    </UButton>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- State 3: MFA Active Component (When not in disable confirmation state) -->
      <div v-else-if="!isDisabling" class="space-y-6 p-2">
        <!-- MFA Active UCard Component (Natural Surface Theme & Full Width) -->
        <UCard
          variant="outline"
          class="w-full bg-gradient-to-br from-emerald-50/50 to-teal-50/50 dark:from-emerald-950/20 dark:to-teal-950/20 border border-emerald-200 dark:border-emerald-900/40 rounded-xl shadow-xs"
        >
          <div class="flex items-start gap-4">
            <div class="p-2.5 rounded-xl bg-emerald-500/20 dark:bg-emerald-500/30 text-emerald-600 dark:text-emerald-400 shrink-0 border border-emerald-200 dark:border-emerald-900/50">
              <UIcon name="i-lucide-shield-check" class="w-6 h-6" />
            </div>
            <div class="space-y-1">
              <h4 class="text-base font-bold text-gray-900 dark:text-white">{{ t('settings.mfa.mfaActiveTitle') }}</h4>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                {{ t('settings.mfa.mfaActiveDesc') }}
              </p>
            </div>
          </div>
        </UCard>

        <div class="pt-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-t border-gray-200 dark:border-gray-800">
          <div class="min-w-0">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('settings.mfa.disableTitle') }}</p>
            <p class="text-sm text-gray-600 dark:text-gray-400 break-words">{{ t('settings.mfa.disableDesc') }}</p>
          </div>
          <UButton
            icon="i-lucide-shield-off"
            color="error"
            variant="soft"
            class="font-bold rounded-xl px-5 py-2.5 transition-all duration-200 cursor-pointer w-full sm:w-auto shrink-0 justify-center"
            @click="() => {isDisabling = true}"
          >
            {{ t('settings.mfa.disableButton') }}
          </UButton>
        </div>
      </div>

      <!-- State 4: MFA Confirm Disable Component (Displayed conditionally when disabling) -->
      <div v-else class="space-y-6 p-4 rounded-2xl border border-rose-200 dark:border-rose-900/40 bg-gradient-to-br from-rose-50/50 to-red-50/50 dark:from-rose-950/20 dark:to-red-950/20">
        <div class="flex items-start gap-3 p-4 rounded-xl bg-rose-500/15 dark:bg-rose-500/20 border border-rose-200 dark:border-rose-900/40 text-rose-900 dark:text-rose-100">
          <UIcon name="i-lucide-alert-triangle" class="w-6 h-6 text-rose-600 dark:text-rose-400 shrink-0 mt-0.5" />
          <div class="space-y-1">
            <h4 class="text-sm font-bold">{{ t('settings.mfa.modalTitle') }}</h4>
            <p class="text-xs opacity-90">
              {{ t('settings.mfa.modalPrompt') }}
            </p>
          </div>
        </div>

        <form class="max-w-md space-y-4" novalidate @submit.prevent="handleDisable">
          <UFormField :label="t('settings.mfa.passwordLabel')">
            <UInput
              v-model="password"
              type="password"
              autocomplete="current-password"
              :placeholder="t('settings.mfa.passwordPlaceholder')"
              size="lg"
              class="w-full"
              :disabled="loading"
            />
          </UFormField>

          <div class="flex items-center gap-3 pt-2">
            <UButton
              type="button"
              variant="subtle"
              color="neutral"
              class="rounded-xl font-semibold px-5"
              :disabled="loading"
              @click="cancelDisable"
            >
              {{ t('settings.mfa.cancelModal') }}
            </UButton>
            <UButton
              type="submit"
              color="error"
              class="rounded-xl text-white font-bold px-6 cursor-pointer"
              :loading="loading"
              :disabled="loading || !password"
            >
              {{ t('settings.mfa.confirmDisable') }}
            </UButton>
          </div>
        </form>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { useAuthStore } from '~/stores/auth'
import { extractErrorMessage } from '~/utils/error'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'
import { TOTP_LENGTH, isMfaAlreadyEnabledError, isMfaSetupChangedError, sanitizeTotp, singleFlight } from '~/utils/mfa'

// The api-auth plugin attaches the bearer token to these $fetch calls
const authStore = useAuthStore()
const { getDeviceFingerprint } = useDeviceFingerprint()
const { t } = useI18n()

const mfaStatus = ref<any>(null)
const loadingStatus = ref(true)
const statusError = ref('')
const setupData = ref<any>(null)
const qrCodeDataURL = ref('')
const verificationCode = ref('')
/** Code the 6th-digit auto-submit last sent (not reactive: only the watcher reads it) */
let lastAutoSubmitted = ''
const loading = ref(false)
const verifying = ref(false)
/** The secret on screen was replaced by a newer setup (409 MFA_SETUP_CHANGED) */
const setupChanged = ref(false)
const isDisabling = ref(false)
const password = ref('')
const toast = useToast()

const isMfaEnabled = computed(() => {
  if (!mfaStatus.value) return false
  return Boolean(mfaStatus.value.is_enabled ?? mfaStatus.value.data?.is_enabled)
})

const canSubmitCode = computed(() => verificationCode.value.length === TOTP_LENGTH)

/** `silent` refreshes in place (no skeleton; a failure keeps the current view). */
const fetchStatus = async ({ silent = false } = {}) => {
  if (!authStore.token) {
    loadingStatus.value = false
    return
  }
  if (!silent) loadingStatus.value = true
  try {
    const response = await $fetch<any>(`${getAuthServiceBaseUrl()}/mfa/status`)
    mfaStatus.value = response?.data?.data ?? response?.data ?? response
    statusError.value = ''
  } catch (err) {
    console.error('Failed to fetch MFA status:', err)
    if (!silent) statusError.value = extractErrorMessage(err, t('settings.mfa.statusFailedDesc'))
  } finally {
    loadingStatus.value = false
  }
}

const resetSetup = () => {
  setupData.value = null
  qrCodeDataURL.value = ''
  verificationCode.value = ''
  lastAutoSubmitted = ''
  setupChanged.value = false
}

/**
 * MFA is on: leave the setup flow, show the enabled view right away, then reload the status from the
 * server and the trusted device list (Settings > Activity reads it from the auth store).
 */
const finishEnable = async () => {
  resetSetup()
  mfaStatus.value = { ...(mfaStatus.value ?? {}), is_enabled: true }
  toast.add({
    title: t('settings.mfa.enabledToast'),
    description: t('settings.mfa.enabledDesc'),
    color: 'success',
    icon: 'i-lucide-circle-check'
  })
  await Promise.allSettled([fetchStatus({ silent: true }), authStore.fetchTrustedDevices()])
}

const handleSetup = singleFlight(async () => {
  if (!authStore.token) return
  loading.value = true
  try {
    const response = await $fetch<any>(`${getAuthServiceBaseUrl()}/mfa/enroll`, {
      method: 'POST',
      body: { mfa_type: 'TOTP' }
    })

    resetSetup()
    setupData.value = response?.data?.data ?? response?.data ?? response
    if (setupData.value?.qr_code_url) {
      qrCodeDataURL.value = await QRCode.toDataURL(setupData.value.qr_code_url)
    }
  } catch (err) {
    // The status shown was stale (e.g. enabled in another tab): show the enabled view, not an error
    if (isMfaAlreadyEnabledError(err)) {
      mfaStatus.value = { ...(mfaStatus.value ?? {}), is_enabled: true }
      await fetchStatus({ silent: true })
      return
    }
    toast.add({
      title: t('settings.mfa.setupFailedTitle'),
      description: extractErrorMessage(err, t('settings.mfa.setupFailedDesc')),
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    loading.value = false
  }
})

const cancelSetup = () => {
  if (!verifying.value) resetSetup()
}

const copySecret = () => {
  if (!setupData.value?.secret) return
  navigator.clipboard.writeText(setupData.value.secret)
  toast.add({
    title: t('settings.mfa.copiedToast'),
    description: t('settings.mfa.copiedDesc'),
    color: 'success',
    icon: 'i-lucide-check'
  })
}

// Enter, the submit button and the auto-submit on the 6th digit all go through this one guarded call:
// a trigger while a request is in flight reuses it instead of sending the code again
// (that second request is what came back as "MFA is already enabled").
const handleVerifySetup = singleFlight(async () => {
  const code = sanitizeTotp(verificationCode.value)
  if (!authStore.token || !setupData.value || code.length !== TOTP_LENGTH) return
  verifying.value = true
  try {
    // Same fingerprint as the login flow, so this browser is listed under Settings > Activity.
    // A retried or double-submitted verify answers 200 with already_enabled; finishEnable covers both.
    const device = getDeviceFingerprint()
    await $fetch(`${getAuthServiceBaseUrl()}/mfa/verify`, {
      method: 'POST',
      body: {
        code,
        device_fingerprint: device.deviceFingerprint,
        device_name: device.deviceName,
        device_type: device.deviceType
      }
    })
    await finishEnable()
  } catch (err) {
    // Enabled by an earlier request (or another tab): that is the outcome the user wanted
    if (isMfaAlreadyEnabledError(err)) {
      await finishEnable()
      return
    }
    // A newer setup (another tab, or a second Enable click) replaced the secret behind this QR code
    if (isMfaSetupChangedError(err)) {
      setupChanged.value = true
      verificationCode.value = ''
      lastAutoSubmitted = ''
      return
    }
    toast.add({
      title: t('settings.mfa.verifyFailedTitle'),
      description: extractErrorMessage(err, t('settings.mfa.verifyFailedDesc')),
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    verifying.value = false
  }
})

// Digits only; auto-submit once per complete code (deleting a digit and retyping re-arms it)
watch(verificationCode, (value) => {
  const clean = sanitizeTotp(value)
  if (clean !== value) {
    verificationCode.value = clean
    return
  }
  if (clean.length < TOTP_LENGTH) {
    lastAutoSubmitted = ''
    return
  }
  if (clean === lastAutoSubmitted || handleVerifySetup.running) return
  lastAutoSubmitted = clean
  void handleVerifySetup()
})

const cancelDisable = () => {
  isDisabling.value = false
  password.value = ''
}

const handleDisable = singleFlight(async () => {
  if (!authStore.token || !password.value) return
  loading.value = true
  try {
    await $fetch(`${getAuthServiceBaseUrl()}/mfa/disable`, {
      method: 'POST',
      body: { password: password.value }
    })

    toast.add({
      title: t('settings.mfa.disabledToast'),
      description: t('settings.mfa.disabledDesc'),
      color: 'success',
      icon: 'i-lucide-circle-check'
    })

    isDisabling.value = false
    password.value = ''
    await fetchStatus({ silent: true })
  } catch (err) {
    toast.add({
      title: t('settings.mfa.disableFailedTitle'),
      description: extractErrorMessage(err, t('settings.mfa.disableFailedDesc')),
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    loading.value = false
  }
})

onMounted(() => fetchStatus())
</script>
