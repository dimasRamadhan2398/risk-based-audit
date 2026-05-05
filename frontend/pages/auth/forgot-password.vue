<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ t('auth.forgotPassword') }}
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          {{ t('auth.forgotPasswordSubtitle') }}
        </p>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
        <div>
          <label for="email-address" class="sr-only">{{ t('auth.email') }}</label>
          <input
            id="email-address"
            v-model="email"
            name="email"
            type="email"
            autocomplete="email"
            required
            class="appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            :placeholder="t('auth.email')"
          >
        </div>

        <div v-if="success" class="rounded-md bg-green-50 p-4">
          <p class="text-sm text-green-800">{{ t('auth.resetEmailSent') }}</p>
        </div>

        <div v-if="error" class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-800">{{ error }}</p>
        </div>

        <div>
          <button
            type="submit"
            :disabled="loading"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {{ loading ? t('auth.sending') : t('auth.sendResetLink') }}
          </button>
        </div>

        <div class="text-center">
          <NuxtLink to="/auth/login" class="font-medium text-indigo-600 hover:text-indigo-500">
            {{ t('auth.backToLogin') }}
          </NuxtLink>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useI18n } from '~/composables/useI18n'

definePageMeta({
  layout: 'auth',
  middleware: 'guest'
})

const { t } = useI18n()
const authStore = useAuthStore()

const email = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  success.value = false

  try {
    await authStore.forgotPassword(email.value)
    success.value = true
  } catch (err: any) {
    error.value = err.message || t('auth.resetError')
  } finally {
    loading.value = false
  }
}
</script>
