<template>
  <div class="min-h-screen flex bg-[var(--bg-main)] transition-colors duration-300 relative overflow-hidden">
    
    <!-- Left panel: Security Context Info (hidden on mobile) -->
    <div class="hidden lg:flex lg:w-4/12 relative overflow-hidden border-r border-[var(--border-main)] p-12 flex-col justify-between select-none" style="background: linear-gradient(160deg, color-mix(in srgb, var(--color-secondary-500) 4%, var(--bg-surface)) 0%, var(--bg-surface) 40%, color-mix(in srgb, var(--color-primary-500) 3%, var(--bg-main)) 100%)">
      <!-- Background subtle grid -->
      <div class="absolute inset-0 bg-[linear-gradient(to_bottom,rgba(var(--ui-primary),0.01)_1px,transparent_1px),linear-gradient(to_right,rgba(var(--ui-primary),0.01)_1px,transparent_1px)] bg-[size:48px_48px] opacity-40 pointer-events-none" />

      <!-- Top brand logo -->
      <div class="relative z-10 flex items-center gap-3">
        <Logo class="h-10" />
      </div>

      <!-- Mid content: Security Info Steps -->
      <div class="relative z-10 my-auto space-y-8 pl-2">
        <div class="flex items-start gap-4">
          <span class="w-6 h-6 rounded-full bg-success-500/10 border border-success-500/20 flex items-center justify-center text-success-600 font-bold text-xs">✓</span>
          <div>
            <h4 class="text-sm font-bold text-[var(--text-main)]">Autentikasi Utama</h4>
            <p class="text-xs text-[var(--text-muted)]">Kredensial dasar berhasil divalidasi</p>
          </div>
        </div>
        
        <div class="flex items-start gap-4">
          <span class="w-6 h-6 rounded-full bg-secondary-500 text-white shadow-lg shadow-secondary-500/25 flex items-center justify-center font-bold text-xs">2</span>
          <div>
            <h4 class="text-sm font-bold text-secondary-500">Verifikasi Dua Faktor (MFA)</h4>
            <p class="text-xs text-[var(--text-muted)]">Masukkan kode OTP dari aplikasi autentikator Anda</p>
          </div>
        </div>
      </div>

      <!-- Bottom metadata -->
      <div class="relative z-10 text-xs text-[var(--text-muted)] opacity-60">
        <span>Sistem Audit Internal Berbasis Risiko v1.0</span>
      </div>
    </div>

    <!-- Right panel: Full screen login verification area -->
    <div class="w-full lg:w-8/12 flex flex-col justify-between bg-[var(--bg-main)] overflow-y-auto min-h-screen">
      
      <!-- Mobile top banner (hidden on large screens) -->
      <div class="lg:hidden px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] flex items-center justify-between text-xs font-semibold select-none">
        <span class="text-secondary-500">Verifikasi MFA</span>
        <span class="text-[var(--text-muted)] font-medium">Langkah 2 dari 2</span>
      </div>

      <!-- Center content section -->
      <div class="my-auto max-w-md w-full mx-auto p-6 sm:p-12 space-y-8 animate-fade-in">
        
        <!-- Header -->
        <div class="space-y-3 inline-flex gap-4">
          <div class="w-12 h-12 rounded-xl bg-secondary-500/10 border border-secondary-500/20 flex items-center justify-center text-secondary-500 px-3">
            <UIcon name="i-lucide-shield-alert" class="w-7 h-7" />
          </div>
          <div>
            <h1 class="text-2xl font-bold text-[var(--text-main)]">Masukkan Kode OTP</h1>
            <p class="text-sm text-[var(--text-muted)]">
              Masukkan 6 digit kode keamanan dari aplikasi autentikator di ponsel Anda untuk memverifikasi.
            </p>
          </div>
        </div>

        <!-- Form -->
        <UForm :state="state" class="space-y-6" @submit="handleVerify">
          
          <!-- OTP Code Input -->
          <UFormField name="code">
            <div class="relative">
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
                  base: 'bg-[var(--bg-surface)] border-[var(--border-main)] text-[var(--text-main)] placeholder-neutral-300 focus:border-secondary-500 focus:ring-secondary-500/20 rounded-xl py-4 text-center',
                }"
                @input="handleOtpInput"
                autofocus
              />
            </div>
          </UFormField>

          <!-- Trust Device Checkbox -->
          <div class="flex items-start gap-3 rounded-xl bg-[var(--bg-surface)] border border-[var(--border-main)] px-4 py-3 shadow-sm">
            <div class="flex items-center h-5 mt-0.5">
              <input
                id="trust-device"
                v-model="state.trustDevice"
                type="checkbox"
                class="rounded border-[var(--border-main)] bg-[var(--bg-main)] text-secondary-500 focus:ring-secondary-500/20"
              >
            </div>
            <div>
              <label for="trust-device" class="text-sm font-semibold text-[var(--text-main)] cursor-pointer select-none">
                Percayakan perangkat ini
              </label>
              <p class="text-xs text-[var(--text-muted)] mt-0.5">
                Jangan minta OTP lagi selama 90 hari di perangkat ini.
              </p>
            </div>
          </div>

          <!-- Timer / Expiry Info -->
          <div class="text-center pt-2">
            <p v-if="timeLeft > 0" class="text-xs text-[var(--text-muted)] flex items-center justify-center gap-1.5">
              Kode berlaku selama: 
              <span class="text-secondary-500 font-mono font-bold">{{ formatTime(timeLeft) }}</span>
            </p>
            <p v-else class="text-xs text-error-600 dark:text-error-400 font-semibold">
              Sesi verifikasi telah kedaluwarsa.
              <NuxtLink to="/auth/login" class="underline hover:opacity-80 ml-1">Ulangi Login</NuxtLink>
            </p>
          </div>

          <!-- Errors -->
          <Transition name="fade">
            <div v-if="error" class="flex items-start gap-3 rounded-xl bg-error-500/10 border border-error-500/20 px-4 py-3">
              <span class="text-error-500 text-lg leading-none">✕</span>
              <p class="text-sm text-error-600 dark:text-error-400 font-medium">{{ error }}</p>
            </div>
          </Transition>

          <!-- Attempts Warning -->
          <Transition name="fade">
            <div v-if="attempts > 0" class="flex items-start gap-3 rounded-xl bg-warning-500/10 border border-warning-500/20 px-4 py-3">
              <span class="text-warning-500 text-lg leading-none">⚠️</span>
              <p class="text-xs text-warning-600 dark:text-warning-400">
                Gagal memverifikasi: {{ attempts }}/3 percobaan. Akun akan terkunci jika terlalu banyak kesalahan.
              </p>
            </div>
          </Transition>

          <!-- Submit Button -->
          <UButton
            id="mfa-verify-btn"
            type="submit"
            block
            size="lg"
            :loading="loading"
            :disabled="state.code.length < 6 || timeLeft === 0"
            class="rounded-xl font-bold tracking-wide transition-all duration-200 text-white disabled:bg-[var(--border-main)] disabled:text-[var(--text-muted)] disabled:cursor-not-allowed"
            :style="state.code.length === 6 && timeLeft > 0 ? { background: 'linear-gradient(135deg, var(--color-secondary-500), var(--color-primary-500))', boxShadow: '0 8px 24px -4px color-mix(in srgb, var(--color-secondary-500) 30%, transparent)' } : {}"
          >
            <template v-if="!loading">
              <span>Verifikasi & Masuk</span>
              <span class="ml-2">→</span>
            </template>
            <template v-else>
              <span>Memverifikasi...</span>
            </template>
          </UButton>

          <!-- Back Button -->
          <div class="text-center pt-2">
            <NuxtLink
              to="/auth/login"
              class="text-sm text-[var(--text-muted)] hover:text-[var(--text-main)] transition-colors duration-150 inline-flex items-center gap-1.5"
            >
              ← Kembali ke halaman login
            </NuxtLink>
          </div>

        </UForm>
      </div>

      <!-- Legal footer -->
      <div class="p-8 text-center text-xs text-[var(--text-muted)] border-t border-[var(--border-main)] bg-[var(--bg-surface)]">
        <span>Keamanan terjamin · Aktivitas verifikasi login dicatat untuk audit kepatuhan.</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'

definePageMeta({
  layout: 'auth',
  middleware: 'guest',
  pageTransition: { name: 'fade', mode: 'out-in' },
})

const authStore = useAuthStore()
const router = useRouter()
const { getDeviceFingerprint } = useDeviceFingerprint()

const state = reactive({
  code: '',
  trustDevice: false,
})

const loading = ref(false)
const error = ref('')
const attempts = ref(0)
const timeLeft = ref(600) // 10 minutes in seconds

// OTP visual indicator
const otpDisplay = computed(() => {
  return Array.from({ length: 6 }, (_, i) => state.code[i] ?? '')
})

// Timer countdown
let timerInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (!authStore.mfaToken) {
    router.push('/auth/login')
    return
  }
  timerInterval = setInterval(() => {
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
  if (state.code.length === 6) {
    handleVerify()
  }
}

const handleVerify = async () => {
  if (state.code.length < 6 || loading.value) return
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
    error.value = err.message || 'Kode verifikasi tidak valid'
    state.code = ''
    attempts.value++
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
