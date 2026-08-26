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
      <!-- Language Switcher Toggle -->
      <div class="flex justify-end mb-3">
        <div class="flex items-center gap-1.5 bg-[var(--bg-surface)]/80 backdrop-blur-sm p-1 rounded-xl border border-[var(--border-main)] shadow-sm">
          <button
            type="button"
            class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1"
            :class="locale === 'id' ? 'bg-primary-500 text-white shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
            @click="setLocale('id')"
          >
            <span>🇮🇩</span> ID
          </button>
          <button
            type="button"
            class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all duration-200 flex items-center gap-1"
            :class="locale === 'en' ? 'bg-primary-500 text-white shadow-sm' : 'text-[var(--text-muted)] hover:text-[var(--text-main)]'"
            @click="setLocale('en')"
          >
            <span>🇬🇧</span> EN
          </button>
        </div>
      </div>

      <!-- Card following the design system -->
      <div class="bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl shadow-2xl px-8 py-10 transition-all duration-300">
        <!-- Header -->
        <div class="flex flex-col items-center mb-8">
          <Logo class="mb-5 h-12" />
          <h1 class="text-2xl font-bold text-[var(--text-main)] tracking-tight">
            {{ t('auth.login.title') }}
          </h1>
          <p class="text-sm text-[var(--text-muted)] mt-1.5 text-center">
            {{ t('auth.login.subtitle') }}
          </p>
        </div>

        <!-- New device notice -->
        <Transition name="fade">
          <div
            v-if="showNewDeviceHint"
            class="mb-4 flex items-start gap-3 rounded-xl bg-warning-500/10 border border-warning-500/20 px-4 py-3"
          >
            <span class="text-warning-500 mt-0.5">⚠</span>
            <p class="text-md text-warning-600 dark:text-warning-400">
              {{ t('auth.login.newDeviceHint') }}
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
              <span class="text-sm font-medium text-[var(--text-main)]">{{ t('auth.login.username') }}</span>
            </template>
            <UInput
              id="username"
              v-model="state.username"
              name="username"
              type="text"
              autocomplete="username"
              :placeholder="t('auth.login.usernamePlaceholder')"
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
              <span class="text-sm font-medium text-[var(--text-main)]">{{ t('auth.login.password') }}</span>
            </template>
            <UInput
              id="password"
              v-model="state.password"
              name="password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              :placeholder="t('auth.login.passwordPlaceholder')"
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
              <span class="text-sm text-[var(--text-muted)]">{{ t('auth.login.rememberMe') }}</span>
            </label>
            <NuxtLink
              to="/auth/forgot-password"
              class="text-sm text-primary-500 hover:text-primary-600 transition-colors"
            >
              {{ t('auth.login.forgotPassword') }}
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
              <span>{{ t('auth.login.button') }}</span>
              <span class="ml-2">→</span>
            </template>
            <template v-else>
              <span>{{ t('auth.login.verifying') }}</span>
            </template>
          </UButton>
        </UForm>

        <!-- Device info chip -->
        <div class="mt-6 flex items-center justify-center gap-2 text-md text-[var(--text-muted)]">
          <span>{{ t('auth.login.encryptedConnection', { device: deviceInfo.deviceName }) }}</span>
        </div>

        <!-- Footer note -->
        <p class="mt-4 text-center text-md text-[var(--text-muted)]">
          {{ t('auth.login.needAccount') }}
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
import { useI18n } from '~/composables/useI18n'

definePageMeta({
  layout: 'auth',
  middleware: 'guest',
  pageTransition: { name: 'fade', mode: 'out-in' },
})

const { t, locale, setLocale } = useI18n()
const authStore = useAuthStore()
const router = useRouter()
const { getDeviceFingerprint } = useDeviceFingerprint()

// Collect device info on mount
const deviceInfo = ref({ deviceFingerprint: '', deviceName: 'Web Browser', deviceType: 'desktop' })
onMounted(() => {
  deviceInfo.value = getDeviceFingerprint()
})

const schema = computed(() => z.object({
  username: z.string().min(1, t('auth.login.usernameRequired')).min(3, t('auth.login.usernameMinLength')),
  password: z.string().min(1, t('auth.login.passwordRequired')),
  rememberMe: z.boolean().optional(),
}))

type Schema = {
  username?: string
  password?: string
  rememberMe?: boolean
}

const state = reactive<Schema>({
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
      username: event.data.username || '',
      password: event.data.password || '',
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
    const status = err?.status ?? err?.statusCode ?? err?.response?.status
    const apiMsg: string = err?.data?.error?.message || err?.data?.message || err?.message || ''

    let errMsg: string

    if (status === 404 || apiMsg.toLowerCase().includes('endpoint_not_found') || apiMsg.toLowerCase().includes('not found')) {
      // Backend route not found — likely misconfigured proxy or wrong URL
      errMsg = t('auth.login.errors.notFound')
    } else if (status === 502 || status === 503 || status === 504) {
      // Gateway/proxy or service unavailable
      errMsg = t('auth.login.errors.serviceUnavailable')
    } else if (status === 500) {
      errMsg = t('auth.login.errors.serverError')
    } else if (status === 401 || status === 403 || apiMsg.toLowerCase().includes('invalid credentials') || apiMsg.toLowerCase().includes('unauthorized')) {
      errMsg = t('auth.login.errors.invalidCredentials')
    } else if (status === 429) {
      errMsg = t('auth.login.errors.tooManyRequests')
    } else if (apiMsg.includes('Failed to fetch') || apiMsg.includes('fetch failed') || apiMsg.includes('ECONNREFUSED') || apiMsg.includes('NetworkError')) {
      errMsg = t('auth.login.errors.networkError')
    } else if (apiMsg) {
      errMsg = apiMsg
    } else {
      errMsg = t('auth.login.errors.generic')
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
