<template>
  <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 relative overflow-hidden bg-[var(--bg-main)] transition-colors duration-300">
    <!-- Animated background subtle gradients aligned with brand secondary/primary -->
    <div class="absolute inset-0 opacity-30 dark:opacity-20" />
    <div
      class="absolute inset-0 opacity-5"
      :style="{
        backgroundImage: 'url(\'../assets/images/audits/bg-login.jpg\')',
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <!-- Subtle grid overlay -->
    <div class="absolute inset-0 bg-[linear-gradient(to_bottom,rgba(var(--ui-primary),0.02)_1px,transparent_1px),linear-gradient(to_right,rgba(var(--ui-primary),0.02)_1px,transparent_1px)] bg-[size:64px_64px] opacity-30" />

    <!-- Glowing orbs matching primary (orange) and secondary (purple) -->
    <div class="absolute top-1/4 -left-32 w-64 h-64 bg-primary-500/10 dark:bg-primary-500/5 rounded-full blur-3xl" />
    <div class="absolute bottom-1/4 -right-32 w-64 h-64 bg-secondary-500/10 dark:bg-secondary-500/5 rounded-full blur-3xl" />

    <div class="relative w-full max-w-md z-10">
      <!-- Card following the design system -->
      <div class="bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl shadow-2xl px-8 py-10 transition-all duration-300">
        <!-- Header -->
        <div class="flex flex-col items-center mb-8">
          <Logo class="mb-5 h-12" />
          <h1 class="text-2xl font-bold text-[var(--text-main)] tracking-tight">
            Selamat Datang Kembali
          </h1>
          <p class="mt-1.5 text-sm text-[var(--text-muted)] text-center">
            Masuk ke Sistem Audit Internal Berbasis Risiko
          </p>
        </div>

        <!-- New device notice -->
        <Transition name="fade">
          <div
            v-if="showNewDeviceHint"
            class="mb-4 flex items-start gap-3 rounded-xl bg-warning-500/10 border border-warning-500/20 px-4 py-3"
          >
            <span class="text-warning-500 mt-0.5">⚠</span>
            <p class="text-xs text-warning-600 dark:text-warning-400">
              Login dari perangkat baru terdeteksi. Verifikasi OTP mungkin diperlukan.
            </p>
          </div>
        </Transition>

        <!-- Form -->
        <UForm
          ref="form"
          :schema="schema"
          :state="state"
          class="space-y-5"
          @submit="handleLogin"
        >
          <!-- Username field -->
          <UFormField name="username">
            <template #label>
              <span class="text-sm font-medium text-[var(--text-main)]">Username</span>
            </template>
            <UInput
              id="username"
              v-model="state.username"
              name="username"
              type="text"
              autocomplete="username"
              placeholder="Masukkan username Anda"
              size="lg"
              class="w-full"
              :ui="{
                base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] placeholder-neutral-400 focus:border-primary-500 focus:ring-primary-500/20 rounded-xl',
              }"
            />
          </UFormField>

          <!-- Password field -->
          <UFormField name="password">
            <template #label>
              <span class="text-sm font-medium text-[var(--text-main)]">Password</span>
            </template>
            <UInput
              id="password"
              v-model="state.password"
              name="password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              placeholder="Masukkan password Anda"
              size="lg"
              class="w-full"
              :ui="{
                base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] placeholder-neutral-400 focus:border-primary-500 focus:ring-primary-500/20 rounded-xl',
              }"
            />
          </UFormField>

          <!-- Remember me -->
          <div class="flex items-center justify-between">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="state.rememberMe"
                type="checkbox"
                class="rounded border-[var(--border-main)] bg-[var(--bg-main)] text-primary-500 focus:ring-primary-500/20"
              >
              <span class="text-sm text-[var(--text-muted)]">Ingat saya</span>
            </label>
            <NuxtLink
              to="/auth/forgot-password"
              class="text-sm text-primary-500 hover:text-primary-600 transition-colors"
            >
              Lupa Password?
            </NuxtLink>
          </div>

          <!-- Error alert -->
          <Transition name="fade">
            <div
              v-if="error"
              class="flex items-start gap-3 rounded-xl bg-error-700 border border-error-500 px-4 py-3"
            >
              <p class="text-sm !text-neutral-100 dark:text-neutral-100">{{ error }}</p>
            </div>
          </Transition>

          <!-- Submit button -->
          <UButton
            id="login-submit-btn"
            type="submit"
            block
            size="lg"
            :loading="loading"
            class="rounded-xl bg-primary-500 hover:bg-primary-600 text-white font-semibold tracking-wide transition-all duration-200 shadow-lg shadow-primary-500/20 hover:shadow-primary-600/30 mt-2"
          >
            <template v-if="!loading">
              <span>Masuk</span>
              <span class="ml-2">→</span>
            </template>
            <template v-else>
              <span>Memverifikasi...</span>
            </template>
          </UButton>
        </UForm>

        <!-- Device info chip -->
        <div class="mt-6 flex items-center justify-center gap-2 text-xs text-[var(--text-muted)]">
          <span>Koneksi terenkripsi · {{ deviceInfo.deviceName }}</span>
        </div>

        <!-- Footer note -->
        <p class="mt-4 text-center text-xs text-[var(--text-muted)]">
          Butuh akun? Hubungi Administrator sistem.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'
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

// Collect device info on mount
const deviceInfo = ref({ deviceFingerprint: '', deviceName: 'Web Browser', deviceType: 'desktop' })
onMounted(() => {
  deviceInfo.value = getDeviceFingerprint()
})

const schema = z.object({
  username: z.string().min(1, 'Username wajib diisi').min(3, 'Username minimal 3 karakter'),
  password: z.string().min(1, 'Password wajib diisi'),
  rememberMe: z.boolean().optional(),
})

type Schema = z.output<typeof schema>

const state = reactive<Partial<Schema>>({
  username: '',
  password: '',
  rememberMe: false,
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)
const showNewDeviceHint = ref(false)

const handleLogin = async (event: FormSubmitEvent<Schema>) => {
  loading.value = true
  error.value = ''

  try {
    const result = await authStore.login({
      username: event.data.username,
      password: event.data.password,
      rememberMe: event.data.rememberMe,
      deviceFingerprint: deviceInfo.value.deviceFingerprint,
      deviceName: deviceInfo.value.deviceName,
      deviceType: deviceInfo.value.deviceType,
    })

    if (result?.mfaRequired) {
      router.push('/auth/mfa-login')
    }
    else if (authStore.needsConfidentialityAgreement) {
      router.push('/auth/confidentiality')
    }
    else {
      router.push('/dashboard')
    }
  }
  catch (err: any) {
    let errMsg = err.message || 'Terjadi kesalahan saat login. Coba lagi.'
    if (errMsg.includes('Failed to fetch') || errMsg.includes('fetch failed')) {
      errMsg = 'Gagal menghubungi server. Pastikan koneksi internet Anda aktif dan server telah berjalan.'
    } else if (errMsg.includes('401') || errMsg.toLowerCase().includes('unauthorized') || errMsg.toLowerCase().includes('invalid credentials')) {
      errMsg = 'Username atau password salah.'
    }
    error.value = errMsg
  }
  finally {
    loading.value = false
  }
}
</script>

<style scoped>
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
