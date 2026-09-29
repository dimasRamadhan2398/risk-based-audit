<template>
  <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 relative overflow-hidden bg-[var(--bg-main)] transition-colors duration-300">
    <div class="absolute top-1/4 -left-32 w-64 h-64 bg-primary-500/10 dark:bg-primary-500/5 rounded-full blur-3xl" />
    <div class="absolute bottom-1/4 -right-32 w-64 h-64 bg-secondary-500/10 dark:bg-secondary-500/5 rounded-full blur-3xl" />

    <div class="relative w-full max-w-md z-10">
      <div class="bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl shadow-2xl px-8 py-10 transition-all duration-300">
        <div class="flex flex-col items-center mb-8">
          <Logo class="mb-5 h-12" />
          <h1 class="text-2xl font-bold text-[var(--text-main)] tracking-tight">
            {{ t('auth.forceChangePassword.title') }}
          </h1>
          <p class="text-sm text-[var(--text-muted)] mt-1.5 text-center">
            {{ t('auth.forceChangePassword.subtitle') }}
          </p>
        </div>

        <UForm
          :schema="schema"
          :state="state"
          class="space-y-5"
          @submit="handleSubmit"
        >
          <UFormField
            v-for="field in fields"
            :key="field.key"
            :name="field.key"
          >
            <template #label>
              <span class="text-sm font-medium text-[var(--text-main)]">{{ field.label }}</span>
            </template>
            <UInput
              v-model="state[field.key]"
              :name="field.key"
              type="password"
              :autocomplete="field.autocomplete"
              size="lg"
              class="w-full"
              :ui="{
                base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] placeholder-neutral-400 focus:border-primary-500 focus:ring-primary-500/20 rounded-xl'
              }"
            />
          </UFormField>

          <p class="text-xs text-[var(--text-muted)]">
            {{ t('auth.forceChangePassword.rules') }}
          </p>

          <div
            v-if="error"
            class="rounded-xl bg-error-500/10 border border-error-500/20 px-4 py-3"
          >
            <p class="text-sm text-error-600 dark:text-error-400">
              {{ error }}
            </p>
          </div>

          <UButton
            type="submit"
            color="primary"
            size="lg"
            block
            class="rounded-xl font-bold"
            :loading="loading"
          >
            {{ t('auth.forceChangePassword.submit') }}
          </UButton>

          <UButton
            type="button"
            color="neutral"
            variant="ghost"
            block
            class="rounded-xl"
            @click="authStore.logout()"
          >
            {{ t('auth.forceChangePassword.logout') }}
          </UButton>
        </UForm>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import * as z from 'zod'
import { useAuthStore } from '~/stores/auth'
import { useI18n } from '~/composables/useI18n'

definePageMeta({
  layout: 'auth',
  middleware: 'auth'
})

const { t } = useI18n()
const authStore = useAuthStore()
const router = useRouter()

type FieldKey = 'currentPassword' | 'newPassword' | 'confirmPassword'

const fields = computed(() => [
  { key: 'currentPassword' as FieldKey, label: t('auth.forceChangePassword.currentPassword'), autocomplete: 'current-password' },
  { key: 'newPassword' as FieldKey, label: t('auth.forceChangePassword.newPassword'), autocomplete: 'new-password' },
  { key: 'confirmPassword' as FieldKey, label: t('auth.forceChangePassword.confirmPassword'), autocomplete: 'new-password' }
])

// Same rules as the change password form in Settings
const schema = computed(() => z.object({
  currentPassword: z.string().min(1, t('auth.forceChangePassword.errors.required')),
  newPassword: z.string()
    .min(8, t('auth.forceChangePassword.errors.weak'))
    .regex(/[A-Z]/, t('auth.forceChangePassword.errors.weak'))
    .regex(/[a-z]/, t('auth.forceChangePassword.errors.weak'))
    .regex(/[0-9]/, t('auth.forceChangePassword.errors.weak')),
  confirmPassword: z.string().min(1, t('auth.forceChangePassword.errors.required'))
}).refine(data => data.newPassword === data.confirmPassword, {
  message: t('auth.forceChangePassword.errors.mismatch'),
  path: ['confirmPassword']
}).refine(data => data.currentPassword !== data.newPassword, {
  message: t('auth.forceChangePassword.errors.same'),
  path: ['newPassword']
}))

const state = reactive<Record<FieldKey, string>>({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  loading.value = true
  error.value = ''
  try {
    await authStore.changePassword(state.currentPassword, state.newPassword)
    await router.push(authStore.needsConfidentialityAgreement ? '/auth/confidentiality' : '/dashboard')
  } catch (err: any) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
