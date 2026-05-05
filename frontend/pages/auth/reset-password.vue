<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ t('auth.resetPassword') }}
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          {{ t('auth.resetPasswordSubtitle') }}
        </p>
      </div>
      <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
        <div class="space-y-4">
          <div>
            <label for="password" class="sr-only">{{ t('auth.newPassword') }}</label>
            <input
              id="password"
              v-model="form.password"
              name="password"
              type="password"
              required
              class="appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              :placeholder="t('auth.newPassword')"
            >
          </div>
          <div>
            <label for="confirm-password" class="sr-only">{{ t('auth.confirmPassword') }}</label>
            <input
              id="confirm-password"
              v-model="form.confirmPassword"
              name="confirmPassword"
              type="password"
              required
              class="appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              :placeholder="t('auth.confirmPassword')"
            >
          </div>
        </div>

        <div v-if="error" class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-800">{{ error }}</p>
        </div>

        <div>
          <button
            type="submit"
            :disabled="loading"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md  bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {{ loading ? t('auth.resetting') : t('auth.resetPasswordButton') }}
          </button>
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
const router = useRouter()
const route = useRoute()

const form = ref({
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref('')

const token = computed(() => route.query.token as string)

const handleSubmit = async () => {
  if (form.value.password !== form.value.confirmPassword) {
    error.value = t('auth.passwordMismatch')
    return
  }

  if (!token.value) {
    error.value = t('auth.invalidResetToken')
    return
  }

  loading.value = true
  error.value = ''

  try {
    await authStore.resetPassword(token.value, form.value.password)
    router.push('/auth/login')
  } catch (err: any) {
    error.value = err.message || t('auth.resetError')
  } finally {
    loading.value = false
  }
}
</script>
