<template>
  <div class="space-y-6 max-w-4xl">
    <!-- General Settings Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xs">
      <template #header>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">General Settings</h3>
      </template>

      <div class="space-y-6">
        <!-- Settings List Container -->
        <div>
          <div class="pb-2">
            <p class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">Email Notifications</p>
          </div>

          <div class="divide-y divide-gray-100 dark:divide-gray-800/70">
            <div
              v-for="item in notificationSettings"
              :key="item.key"
              class="flex items-center justify-between py-3.5"
            >
              <div class="space-y-0.5 min-w-0 flex-1">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.label }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ item.description }}</p>
              </div>
              <USwitch
                :model-value="settings[item.key]"
                :disabled="item.disabled"
                color="primary"
                @update:model-value="(val: boolean) => settings[item.key] = val"
              />
            </div>

            <!-- Dark Mode Toggle -->
            <div class="flex items-center justify-between py-3.5">
              <div class="space-y-0.5 min-w-0 flex-1">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">Dark Mode</p>
                <p class="text-xs text-gray-500 dark:text-gray-400">Toggle dark mode theme interface</p>
              </div>
              <USwitch v-model="settings.darkMode" color="primary" />
            </div>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Language & Region Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xs">
      <template #header>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">Language & Region</h3>
      </template>

      <div class="space-y-4">
        <UFormField label="Timezone" help="Selected timezone will format dates across reports and activity logs.">
          <USelect v-model="settings.timezone" :options="timezones" size="lg" class="w-full max-w-sm" />
        </UFormField>
      </div>
    </UCard>

    <!-- Change Password Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xs">
      <template #header>
        <div class="space-y-2">
          <h3 class="text-lg font-bold text-gray-900 dark:text-white">Change Password</h3>
          <UAccordion :items="items" variant="subtle" color="primary">
            <template #content="{ item }">
              <ul class="pb-3.5 px-4 text-xs sm:text-sm text-gray-700 dark:text-gray-300 space-y-1 list-disc list-inside">
                <li>Preventing unauthorized access from potentially compromised old passwords</li>
                <li>Protecting sensitive company audit data</li>
                <li>Meeting company information security standards</li>
                <li class="list-none mt-2 font-semibold">Recommendation: Use a combination of uppercase letters, lowercase letters, numbers, and symbols. Minimum 8 characters.</li>
              </ul>
            </template>
          </UAccordion>
        </div>
      </template>

      <form @submit.prevent="changePassword" class="space-y-4 max-w-lg">
        <UFormField
          v-for="item in passwordSettings"
          :key="item.key"
          :label="item.label"
          :error="passwordErrors[item.key]"
        >
          <UInput
            type="password"
            :placeholder="`Enter ${item.label.toLowerCase()}`"
            v-model="passwordForm[item.key]"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <div v-if="passwordSuccess" class="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-sm font-medium">
          {{ passwordSuccess }}
        </div>

        <div class="flex justify-end gap-3 pt-2">
          <UButton type="button" variant="subtle" color="neutral" class="rounded-xl font-semibold" @click="resetPasswordForm">Cancel</UButton>
          <UButton type="submit" color="primary" class="rounded-xl font-bold" :loading="isChangingPassword">Change Password</UButton>
        </div>
      </form>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { AccordionItem } from '@nuxt/ui'
import { z } from 'zod'

const { setLocale } = useI18n()
const { copy } = useClipboard()
const toast = useToast()

type Settings = {
  emailNotifications: boolean
  pushNotifications: boolean
  darkMode: boolean
  critical: boolean
  deadlineReminder: boolean
  statusUpdate: boolean
  adminSystem: boolean
  language: string
  timezone: string
  twoFactor: boolean
}

const settings = ref<Settings>({
  emailNotifications: true,
  pushNotifications: false,
  darkMode: false,
  critical: true,
  deadlineReminder: true,
  statusUpdate: true,
  adminSystem: false,
  language: 'en',
  timezone: 'asia/jakarta',
  twoFactor: false,
})

const items = ref<AccordionItem[]>([
  {
    label: 'Mengapa perlu mengubah password secara berkala?',
    icon: 'i-lucide-info',
  },
])

const timezones = [
  { label: 'UTC+7 (Jakarta)', value: 'asia/jakarta' },
  { label: 'UTC+8 (Singapore)', value: 'asia/singapore' },
  { label: 'UTC+0 (London)', value: 'europe/london' },
]

type BooleanSettingsKeys = 'critical' | 'deadlineReminder' | 'statusUpdate' | 'adminSystem'

const notificationSettings: { key: BooleanSettingsKeys; label: string; description: string; disabled: boolean }[] = [
  { key: 'critical', label: 'Critical Alerts', description: 'High priority notifications', disabled: false },
  { key: 'deadlineReminder', label: 'Deadline Reminder', description: 'Remind before due dates', disabled: false },
  { key: 'statusUpdate', label: 'Status Update', description: 'Task status changes', disabled: false },
  { key: 'adminSystem', label: 'Admin & System', description: 'System-level messages', disabled: false },
]

type PasswordSettingsKeys = 'currPassword' | 'newPassword' | 'confirmPassword'

const passwordSettings: { key: PasswordSettingsKeys; label: string; disabled: boolean }[] = [
  { key: 'currPassword', label: 'Current Password', disabled: false },
  { key: 'newPassword', label: 'New Password', disabled: false },
  { key: 'confirmPassword', label: 'Confirm Password', disabled: false },
]

const passwordSchema = z.object({
  currPassword: z.string().min(1, 'Please enter your current password'),
  newPassword: z.string()
    .min(8, 'Password must be at least 8 characters long')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number'),
  confirmPassword: z.string().min(1, 'Please confirm your new password'),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: 'New passwords do not match',
  path: ['confirmPassword'],
}).refine((data) => data.currPassword !== data.newPassword, {
  message: 'New password must be different from current password',
  path: ['newPassword'],
})

type PasswordFormData = z.infer<typeof passwordSchema>

const passwordForm = ref<PasswordFormData>({
  currPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const passwordErrors = ref<Record<PasswordSettingsKeys, string>>({
  currPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const isChangingPassword = ref(false)
const passwordSuccess = ref('')

function resetPasswordForm() {
  passwordForm.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  passwordErrors.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  passwordSuccess.value = ''
}

async function changePassword() {
  passwordErrors.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  passwordSuccess.value = ''

  const result = passwordSchema.safeParse(passwordForm.value)

  if (!result.success) {
    result.error.errors.forEach((error) => {
      const field = error.path[0] as PasswordSettingsKeys
      passwordErrors.value[field] = error.message
    })
    return
  }

  isChangingPassword.value = true

  try {
    await new Promise((resolve) => setTimeout(resolve, 1200))
    passwordSuccess.value = 'Password changed successfully!'
    resetPasswordForm()
    toast.add({
      title: 'Password Updated',
      description: 'Password Anda telah berhasil diperbarui.',
      color: 'success',
      icon: 'i-lucide-check-circle',
    })
  } catch (error) {
    passwordErrors.value.currPassword = 'Failed to change password. Please try again.'
  } finally {
    isChangingPassword.value = false
  }
}
</script>