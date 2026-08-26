<template>
  <div class="space-y-6 w-full">
    <!-- General Settings Card -->
    <UCard class="w-full border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.general.cardTitle') }}</h3>
      </template>

      <div class="space-y-6">
        <!-- Settings List Container -->
        <div>
          <div class="pb-2">
            <p class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.general.emailNotifications') }}</p>
          </div>

          <div class="divide-y divide-gray-100 dark:divide-gray-800/70">
            <div
              v-for="item in notificationSettings"
              :key="item.key"
              class="flex items-center justify-between py-3.5"
            >
              <div class="space-y-0.5 min-w-0 flex-1">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ item.label }}</p>
                <p class="text-md text-gray-500 dark:text-gray-400">{{ item.description }}</p>
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
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('settings.general.darkMode') }}</p>
                <p class="text-md text-gray-500 dark:text-gray-400">{{ t('settings.general.darkModeDesc') }}</p>
              </div>
              <USwitch v-model="settings.darkMode" color="primary" />
            </div>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Language & Region Card -->
    <UCard class="w-full border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.general.languageRegion') }}</h3>
      </template>

      <div class="space-y-4">
        <UFormField :label="t('settings.general.language')" :help="t('settings.general.languageHelp')">
          <USelect
            v-model="settings.language"
            :items="languages"
            size="lg"
            class="w-full max-w-sm"
            @update:model-value="onLanguageChange"
          />
        </UFormField>

        <UFormField :label="t('settings.general.timezone')" :help="t('settings.general.timezoneHelp')">
          <USelect v-model="settings.timezone" :items="timezones" size="lg" class="w-full max-w-sm" />
        </UFormField>
      </div>
    </UCard>

    <!-- Change Password Card -->
    <UCard class="w-full border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <div class="space-y-2">
          <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.general.changePassword') }}</h3>
          <UAccordion :items="accordionItems" variant="subtle" color="primary">
            <template #content="{ item }">
              <ul class="pb-3.5 px-4 text-md sm:text-sm text-gray-700 dark:text-gray-300 space-y-1 list-disc list-inside">
                <li>{{ t('settings.general.tip1') }}</li>
                <li>{{ t('settings.general.tip2') }}</li>
                <li>{{ t('settings.general.tip3') }}</li>
                <li class="list-none mt-2 font-semibold">{{ t('settings.general.recommendation') }}</li>
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
            :placeholder="item.placeholder"
            v-model="passwordForm[item.key]"
            size="lg"
            class="w-full"
          />
        </UFormField>

        <div v-if="passwordSuccess" class="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-sm font-medium">
          {{ passwordSuccess }}
        </div>

        <div class="flex justify-end gap-3 pt-2">
          <UButton type="button" variant="subtle" color="neutral" class="rounded-xl font-semibold" @click="resetPasswordForm">{{ t('settings.general.cancel') }}</UButton>
          <UButton type="submit" color="primary" class="rounded-xl font-bold" :loading="isChangingPassword">{{ t('settings.general.submitChangePassword') }}</UButton>
        </div>
      </form>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { AccordionItem } from '@nuxt/ui'
import { z } from 'zod'

const { t, locale, setLocale } = useI18n()
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
  language: locale.value === 'id' ? 'Bahasa Indonesia' : 'English',
  timezone: 'asia/jakarta',
  twoFactor: false,
})

watch(locale, (newLoc) => {
  settings.value.language = newLoc === 'id' ? 'Bahasa Indonesia' : 'English'
}, { immediate: true })

const accordionItems = computed<AccordionItem[]>(() => [
  {
    label: t('settings.general.accordionWhy'),
    icon: 'i-lucide-info',
  },
])

const languages = ['English', 'Bahasa Indonesia']

function onLanguageChange(langVal?: any) {
  const strVal = String(langVal || 'English')
  const localeCode = strVal === 'Bahasa Indonesia' || strVal === 'id' ? 'id' : 'en'
  if (setLocale) {
    try {
      setLocale(localeCode)
    } catch {
      // i18n fallback
    }
  }
  toast.add({
    title: t('settings.general.languageUpdatedToast'),
    description: t('settings.general.languageSetToast', { lang: strVal }),
    color: 'success',
  })
}

const timezones = [
  { label: 'UTC+7 (Jakarta)', value: 'asia/jakarta' },
  { label: 'UTC+8 (Singapore)', value: 'asia/singapore' },
  { label: 'UTC+0 (London)', value: 'europe/london' },
]

type BooleanSettingsKeys = 'critical' | 'deadlineReminder' | 'statusUpdate' | 'adminSystem'

const notificationSettings = computed(() => [
  { key: 'critical' as BooleanSettingsKeys, label: t('settings.general.criticalAlerts'), description: t('settings.general.criticalAlertsDesc'), disabled: false },
  { key: 'deadlineReminder' as BooleanSettingsKeys, label: t('settings.general.deadlineReminder'), description: t('settings.general.deadlineReminderDesc'), disabled: false },
  { key: 'statusUpdate' as BooleanSettingsKeys, label: t('settings.general.statusUpdate'), description: t('settings.general.statusUpdateDesc'), disabled: false },
  { key: 'adminSystem' as BooleanSettingsKeys, label: t('settings.general.adminSystem'), description: t('settings.general.adminSystemDesc'), disabled: false },
])

type PasswordSettingsKeys = 'currPassword' | 'newPassword' | 'confirmPassword'

const passwordSettings = computed(() => [
  { key: 'currPassword' as PasswordSettingsKeys, label: t('settings.general.currentPassword'), placeholder: t('settings.general.enterCurrentPassword'), disabled: false },
  { key: 'newPassword' as PasswordSettingsKeys, label: t('settings.general.newPassword'), placeholder: t('settings.general.enterNewPassword'), disabled: false },
  { key: 'confirmPassword' as PasswordSettingsKeys, label: t('settings.general.confirmPassword'), placeholder: t('settings.general.enterConfirmPassword'), disabled: false },
])

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
    passwordSuccess.value = t('settings.general.successMessage')
    resetPasswordForm()
    toast.add({
      title: 'Password Updated',
      description: t('settings.general.successMessage'),
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