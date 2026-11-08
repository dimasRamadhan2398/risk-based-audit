<template>
  <div class="min-h-screen relative overflow-hidden bg-audit-bg bg-cover bg-center flex flex-col transition-colors duration-300">
    <!-- Solid dark overlay so the photo stays readable behind the card -->
    <div class="absolute inset-0 bg-neutral-950/75 pointer-events-none" />

    <!-- Top bar: language switcher (primary, same as navbar) -->
    <div class="relative z-10 flex justify-end items-center px-6 py-5">
      <div class="flex items-center gap-1 bg-white/10 backdrop-blur-md p-1 rounded-xl border border-white/15 shadow-sm select-none">
        <button
          type="button"
          class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200"
          :class="locale === 'id' ? 'bg-primary-500 text-white shadow-sm' : 'text-white/70 hover:text-white'"
          @click="setLocale('id')"
        >
          ID
        </button>
        <button
          type="button"
          class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200"
          :class="locale === 'en' ? 'bg-primary-500 text-white shadow-sm' : 'text-white/70 hover:text-white'"
          @click="setLocale('en')"
        >
          EN
        </button>
      </div>
    </div>

    <!-- Center card -->
    <div class="relative z-10 flex-1 flex items-center justify-center px-4 sm:px-6 pb-10">
      <div class="w-full max-w-lg animate-fade-in">
        <div class="bg-[var(--bg-surface)]/95 backdrop-blur-md border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden">
          <!-- Brand accent strip -->
          <div class="h-1.5 w-full bg-primary-500" />

          <div class="px-5 sm:px-8 py-7 sm:py-9 space-y-7">
            <!-- Logo -->
            <div class="flex justify-center">
              <Logo class="h-10" />
            </div>

            <!-- Step progress -->
            <div class="flex items-center gap-3 select-none">
              <div class="flex items-center gap-2 min-w-0">
                <span class="w-6 h-6 shrink-0 rounded-full bg-success-500/15 border border-success-500/30 text-success-600 dark:text-success-400 flex items-center justify-center text-xs font-bold">✓</span>
                <span class="text-xs font-semibold text-[var(--text-muted)] whitespace-nowrap">{{ t('auth.mfaLogin.primaryAuth') }}</span>
              </div>
              <div class="flex-1 h-0.5 rounded-full bg-primary-500" />
              <div class="flex items-center gap-2 min-w-0">
                <span class="w-6 h-6 shrink-0 rounded-full bg-primary-500 text-white shadow-md shadow-primary-500/30 flex items-center justify-center text-xs font-bold">2</span>
                <span class="text-xs font-bold text-primary-600 dark:text-primary-400 whitespace-nowrap">{{ t('auth.mfaLogin.mfaTitle') }}</span>
              </div>
            </div>

            <!-- Header -->
            <div class="flex flex-col items-center text-center gap-3">
              <div class="w-12 h-12 rounded-xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
                <UIcon name="i-lucide-shield-check" class="w-6 h-6" />
              </div>
              <div>
                <h1 class="text-2xl font-bold text-[var(--text-main)] tracking-tight">{{ t('auth.mfaLogin.headerTitle') }}</h1>
                <p class="text-sm text-[var(--text-muted)] mt-1.5">
                  {{ t('auth.mfaLogin.headerSubtitle') }}
                </p>
              </div>
            </div>

            <!-- Form -->
            <UForm :state="state" class="space-y-6" @submit="handleVerify">
              <!-- OTP Code Input -->
              <UFormField name="code">
                <UInput
                  id="mfa-code"
                  v-model="state.code"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  maxlength="6"
                  placeholder="000000"
                  size="lg"
                  class="w-full text-center text-3xl tracking-[0.5em] font-mono font-bold"
                  :ui="{
                    base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] placeholder-neutral-300 focus:border-primary-500 focus:ring-primary-500/20 rounded-xl py-4 text-center',
                  }"
                  :disabled="locked"
                  @input="handleOtpInput"
                  autofocus
                />
              </UFormField>

              <!-- Trust Device Checkbox -->
              <label for="trust-device" class="flex items-start gap-3 rounded-xl bg-secondary-50 dark:bg-secondary-500/15 border border-secondary-500 px-4 py-3 cursor-pointer">
                <input
                  id="trust-device"
                  v-model="state.trustDevice"
                  type="checkbox"
                  class="mt-1 rounded border-[var(--border-main)] bg-[var(--bg-surface)] text-primary-500 focus:ring-primary-500/20"
                >
                <div>
                  <span class="text-sm font-semibold text-secondary-700 dark:text-secondary-300 select-none">
                    {{ t('auth.mfaLogin.trustDeviceLabel') }}
                  </span>
                  <p class="text-md text-secondary-700 dark:text-secondary-300 mt-0.5">
                    {{ t('auth.mfaLogin.trustDeviceDesc') }}
                  </p>
                </div>
              </label>

              <!-- Timer / Expiry Info -->
              <div v-if="!locked" class="text-center">
                <p v-if="timeLeft > 0" class="text-md text-[var(--text-muted)] flex items-center justify-center gap-1.5">
                  {{ t('auth.mfaLogin.codeValidTime') }}
                  <span class="text-primary-500 tabular-nums font-bold">{{ formatTime(codeTimeLeft) }}</span>
                </p>
                <p v-else class="text-md text-error-600 dark:text-error-400 font-semibold">
                  {{ t('auth.mfaLogin.sessionExpired') }}
                  <NuxtLink to="/auth/login" class="underline hover:opacity-80 ml-1">{{ t('auth.mfaLogin.retryLogin') }}</NuxtLink>
                </p>
              </div>

              <!-- Error + attempts warning -->
              <Transition name="fade">
                <div v-if="error" class="flex items-start gap-3 rounded-xl bg-error-50 dark:bg-error-500/15 border border-error-500 px-4 py-3">
                  <span class="text-error-600 dark:text-error-400 text-lg leading-none">✕</span>
                  <div class="space-y-1">
                    <p class="text-sm text-error-700 dark:text-error-300 font-semibold">{{ error }}</p>
                    <p v-if="!locked && remainingAttempts !== null" class="text-sm text-error-700/90 dark:text-error-300/90">
                      {{ t('auth.mfaLogin.attemptsWarning', { remaining: remainingAttempts, max: maxAttempts }) }}
                    </p>
                  </div>
                </div>
              </Transition>

              <!-- Submit Button -->
              <UButton
                id="mfa-verify-btn"
                type="submit"
                color="primary"
                block
                size="lg"
                :loading="loading"
                :disabled="state.code.length < 6 || timeLeft === 0 || locked"
                class="rounded-xl font-bold tracking-wide transition-all duration-200 !text-white disabled:bg-[var(--border-main)] disabled:!text-[var(--text-muted)] disabled:cursor-not-allowed"
              >
                <template v-if="!loading">
                  <span>{{ t('auth.mfaLogin.verifyBtn') }}</span>
                  <span class="ml-2">→</span>
                </template>
                <template v-else>
                  <span>{{ t('auth.mfaLogin.verifyingBtn') }}</span>
                </template>
              </UButton>

              <!-- Back Button -->
              <div class="text-center">
                <NuxtLink
                  to="/auth/login"
                  class="text-sm text-[var(--text-muted)] hover:text-primary-500 transition-colors duration-150 inline-flex items-center gap-1.5"
                >
                  {{ t('auth.mfaLogin.backToLogin') }}
                </NuxtLink>
              </div>
            </UForm>
          </div>
        </div>
      </div>
    </div>

    <!-- Legal footer -->
    <div class="relative z-10 px-6 pb-6 text-center text-xs text-white/60 select-none">
      <p>{{ t('auth.mfaLogin.legalFooter') }}</p>
      <p class="opacity-70 mt-1">{{ t('auth.mfaLogin.systemVersion') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'
import { useI18n } from '~/composables/useI18n'

definePageMeta({
  layout: 'auth',
  middleware: 'guest',
  pageTransition: { name: 'fade', mode: 'out-in' },
})

const authStore = useAuthStore()
const router = useRouter()
const { getDeviceFingerprint } = useDeviceFingerprint()
const { t, locale, setLocale } = useI18n()

const state = reactive({
  code: '',
  trustDevice: false,
})

const loading = ref(false)
const error = ref('')
// Attempt state comes from auth-service, which counts wrong codes per user.
const remainingAttempts = ref<number | null>(null)
const maxAttempts = ref(3)
const locked = ref(false)
// MFA login session: matches the 10-minute mfa_login token TTL in auth-service.
const timeLeft = ref(600)

// Authenticator (TOTP) codes rotate every 30 seconds, aligned to the clock,
// so the countdown follows the same window the authenticator app shows.
const TOTP_PERIOD = 30
const totpRemaining = () => TOTP_PERIOD - (Math.floor(Date.now() / 1000) % TOTP_PERIOD)
const codeTimeLeft = ref(totpRemaining())

// OTP visual indicator
const otpDisplay = computed(() => {
  return Array.from({ length: 6 }, (_, i) => state.code[i] ?? '')
})

// Timer countdown
let timerInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (window.location.search.includes('demo=true')) {
    setLocale('id')
    state.code = '582914'
    state.trustDevice = true
  }
  if (!authStore.mfaToken && !window.location.search.includes('demo=true')) {
    router.push('/auth/login')
    return
  }
  timerInterval = setInterval(() => {
    codeTimeLeft.value = totpRemaining()
    if (timeLeft.value > 0) {
      timeLeft.value--
    }
    else {
      clearInterval(timerInterval!)
    }
  }, 1000)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})

const formatTime = (seconds: number) => {
  const m = Math.floor(seconds / 60).toString().padStart(2, '0')
  const s = (seconds % 60).toString().padStart(2, '0')
  return `${m}:${s}`
}

const handleOtpInput = (e: Event) => {
  const val = (e.target as HTMLInputElement).value.replace(/\D/g, '')
  state.code = val.substring(0, 6)
  // Auto-submit when 6 digits entered
  if (state.code.length === 6 && !window.location.search.includes('demo=true')) {
    handleVerify()
  }
}

const handleVerify = async () => {
  if (state.code.length < 6 || loading.value || locked.value) return
  loading.value = true
  error.value = ''

  const deviceInfo = getDeviceFingerprint()

  try {
    await authStore.verifyMFALogin({
      mfaToken: authStore.mfaToken!,
      code: state.code,
      trustDevice: state.trustDevice,
      deviceFingerprint: deviceInfo.deviceFingerprint,
      deviceName: deviceInfo.deviceName,
      deviceType: deviceInfo.deviceType,
    })

    if (authStore.needsConfidentialityAgreement) {
      router.push('/auth/confidentiality')
    }
    else {
      router.push('/dashboard')
    }
  }
  catch (err: any) {
    state.code = ''
    const details = err?.details ?? {}
    if (err?.code === 'ACCOUNT_LOCKED') {
      locked.value = true
      remainingAttempts.value = 0
      const minutes = Math.max(1, Math.ceil((Number(details.retry_after_seconds) || 900) / 60))
      error.value = t('auth.mfaLogin.accountLocked', { minutes })
      if (timerInterval) clearInterval(timerInterval)
      return
    }
    error.value = err?.code === 'INVALID_CODE'
      ? t('auth.mfaLogin.invalidCode')
      : (err?.message || t('auth.mfaLogin.invalidCode'))
    if (typeof details.remaining_attempts === 'number') {
      remainingAttempts.value = details.remaining_attempts
      if (typeof details.max_attempts === 'number') maxAttempts.value = details.max_attempts
    }
  }
  finally {
    loading.value = false
  }
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
