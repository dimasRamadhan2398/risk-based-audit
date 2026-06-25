<template>
  <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <UContainer class="max-w-md w-full space-y-8 bg-neutral-400 px-8 py-8 rounded-xl relative shadow-2xl">
      <div class="flex flex-col items-center">
        <Logo class="mb-4" />
        <h2 class="mt-2 text-center text-3xl font-extrabold text-gray-900">
          Two-Factor Authentication
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          Please enter the 6-digit code from your authenticator app.
        </p>
      </div>

      <UForm :state="state" class="mt-8 space-y-6" @submit="handleVerify">
        <UFormField label="Verification Code" name="code">
          <UInput
            v-model="state.code"
            placeholder="000000"
            size="lg"
            class="text-center text-2xl tracking-widest"
            maxlength="6"
            required
          />
        </UFormField>

        <UFormField name="trust_device">
          <UCheckbox v-model="state.trustDevice" label="Trust this device for 90 days" />
        </UFormField>

        <div v-if="error" class="rounded-md bg-red-50 p-4">
          <p class="text-sm text-red-800">{{ error }}</p>
        </div>

        <div>
          <UButton
            type="submit"
            block
            color="primary"
            :loading="loading"
            label="Verify & Login"
          />
        </div>

        <div class="text-center">
          <UButton variant="ghost" to="/auth/login">Back to Login</UButton>
        </div>
      </UForm>
    </UContainer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  layout: 'auth',
  middleware: 'guest'
})

const authStore = useAuthStore()
const router = useRouter()

const state = reactive({
  code: '',
  trustDevice: false
})

const loading = ref(false)
const error = ref('')

// Redirect if no MFA token
onMounted(() => {
  if (!authStore.mfaToken) {
    router.push('/auth/login')
  }
})

const handleVerify = async () => {
  loading.value = true
  error.value = ''

  try {
    await authStore.verifyMFALogin(state.code, state.trustDevice)
    router.push('/dashboard')
  } catch (err: any) {
    error.value = err.message || 'Verification failed'
  } finally {
    loading.value = false
  }
}
</script>
