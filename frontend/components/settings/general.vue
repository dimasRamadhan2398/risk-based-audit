<template>
<div class="space-y-6">
              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold text-gray-900">General Settings</h3>
                </template>

                <div class="space-y-4">
                  <div class="flex flex-col items-start justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Email Notifications</p>
                      <p class="text-sm text-gray-500">Receive email updates about your activity</p>
                    </div>
                    <div class="flex flex-col w-full p-4">
                      <UFormField v-for="item in notificationSettings" :key="item.key" :label="item.label" :description="item.description" orientation="horizontal" class="w-full" :ui="{
                        label: 'min-w-7xl'
                      }">
                        <USwitch :model-value="settings[item.key]" @update:model-value="(val: boolean) => settings[item.key] = val" />
                      </UFormField>
                    </div>
                  </div>

                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Push Notifications</p>
                      <p class="text-sm text-gray-500">Receive push notifications on your device</p>
                    </div>
                    <UFormField>
                      <USwitch v-model="settings.pushNotifications" />
                    </UFormField>
                  </div>

                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Dark Mode</p>
                      <p class="text-sm text-gray-500">Toggle dark mode theme</p>
                    </div>
                    <UFormField>
                      <USwitch v-model="settings.darkMode" />
                    </UFormField>
                  </div>
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <h3 class="text-lg font-semibold text-gray-900">Language & Region</h3>
                </template>

                <div class="space-y-4">
                  <!-- <UFormField label="Language">
                    <USelect v-model="selectedLanguage?.value" :options="languages" @change="changeLanguage" />
                  </UFormField> -->

                  <UFormField label="Timezone">
                    <USelect v-model="settings.timezone" :options="timezones" />
                  </UFormField>
                </div>
              </UCard>

              <UCard>
                <template #header>
                  <div class="pb-2">
                    <h3 class="text-lg font-semibold text-gray-900">Change Password</h3>
                  </div>
                  <UAccordion :items="items" variant="subtle" color="primary">
                    <template #content="{ item }">
                      <ul class="pb-3.5 px-6 text-sm text-neutral-900 space-y-1 list-disc list-inside">
                        <li>Preventing unauthorized access from potentially compromised old passwords</li>
                        <li>Protecting sensitive company audit data</li>
                        <li>Meeting company information security standards</li>
                        <li class="list-none mt-2"><strong>Recommendation:</strong> Use a combination of uppercase letters, lowercase letters, numbers, and symbols. Minimum 8 characters.</li>
                      </ul>
                    </template>
                  </UAccordion>
                </template>

                <form @submit.prevent="changePassword" class="space-y-4">
                  <UFormField
                    v-for="item in passwordSettings"
                    :key="item.key"
                    :label="item.label"
                    orientation="horizontal"
                    class="w-full"
                    :ui="{ label: 'w-48',root: 'justify-center items-center'}"
                    :error="passwordErrors[item.key]"
                  >
                    <UInput
                      type="password"
                      :placeholder="`Enter ${item.label.toLowerCase()}`"
                      v-model="passwordForm[item.key]"
                    />
                  </UFormField>

                  <div v-if="passwordSuccess" class="p-3 rounded-md bg-success/10 text-success text-sm">
                    {{ passwordSuccess }}
                  </div>

                  <div class="flex justify-end gap-3 pt-2">
                    <UButton type="button" variant="outline" @click="resetPasswordForm">Cancel</UButton>
                    <UButton type="submit" color="primary" :loading="isChangingPassword">Change Password</UButton>
                  </div>
                </form>
              </UCard>
              <UCard>
                <template #header>
                  <div class="pb-2">
                    <h3 class="text-lg font-semibold text-gray-900">Two Factor Authentication</h3>
                  </div>
                  <UAccordion :items="twoFactorItem" variant="subtle" color="primary">
                    <template #content="{ item }">
                      <ul class="pb-3.5 px-6 text-sm text-neutral-900 space-y-1 list-disc list-inside">
                        <li class="list-none mt-2">Two Factor Authentication (2FA) is a security method that requires two ways of verification</li>
                        <li>Something you know: Your password</li>
                        <li>Something you have: OTP code from authenticator app</li>
                        <li class="list-none mt-2">With 2FA, even if your password is compromised, your account remains secure because the OTP code only exists on your device.</li>
                      </ul>
                    </template>
                  </UAccordion>
                </template>
               <div class="flex flex-col gap-8">
                 <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium text-gray-900">Status 2FA</p>
                      <p class="text-sm text-gray-500">Nonaktif - Aktifkan untuk keamanan lebih</p>
                    </div>
                    <UFormField>
                      <USwitch :model-value="settings.twoFactor" class="self-end! text-right! content-end justify-between items-end flex-0" @update:model-value="updateTwoFactor"/>
                    </UFormField>
                  </div>
               </div>
               <div v-if="settings.twoFactor">
                  <div class="wrapper-vertical-center">
                    <h2>Scan QR Code dengan Authenticator App</h2>
                    <img src="../../assets/images/dummies/qr-code.png" alt="qr-code" class="w-48 h-48">
                    <h5>Tidak bisa scan? Kode Manual: <span class="font-semibold text-info-800"> {{ manualCode }} <UIcon name="copy" size="20" @click="copyManualCode" class="cursor-pointer"></UIcon></span></h5>
                  </div>
               </div>
            </UCard>
          </div>
</template>
<script setup lang="ts">
import type { AccordionItem } from '@nuxt/ui';
import { z } from 'zod'

const { locale, setLocale } = useI18n();

const { copy } = useClipboard();
const toast = useToast();

type Settings = {
  emailNotifications: boolean;
  pushNotifications: boolean;
  darkMode: boolean;
  critical: boolean;
  deadlineReminder: boolean;
  statusUpdate: boolean;
  adminSystem: boolean;
  language: string;
  timezone: string;
  twoFactor: boolean;
};

const settings = ref<Settings>({
  emailNotifications: true,
  pushNotifications: false,
  darkMode: false,
  critical: true,
  deadlineReminder: true,
  statusUpdate: true,
  adminSystem: false,
  language: "en",
  timezone: "asia/jakarta",
  twoFactor: false,
});

const items = ref<AccordionItem[]>([
  {
    label: 'Mengapa perlu mengubah password secara berkala?',
    icon: 'info',
  }
])

const twoFactorItem = ref<AccordionItem[]>([
  {
    label: 'Apa itu Two Factor Authentication dan mengapa penting?',
    icon: 'info',
  }
])

const manualCode = "A1B2C3D4E5"

const selectedLanguage = computed(() => {
  return languages.find((lang) => lang.value === settings.value.language);
})

const languages = [
  { label: "English", value: "en" },
  { label: "Indonesia", value: "id" },
];

const changeLanguage = (item: any) => {
  setLocale(item.value)
}

const timezones = [
  { label: "UTC+7 (Jakarta)", value: "asia/jakarta" },
  { label: "UTC+8 (Singapore)", value: "asia/singapore" },
  { label: "UTC+0 (London)", value: "europe/london" },
];

type BooleanSettingsKeys = 'critical' | 'deadlineReminder' | 'statusUpdate' | 'adminSystem'

const notificationSettings: { key: BooleanSettingsKeys; label: string; description: string; disabled: boolean }[] = [
  { key: 'critical',         label: 'Critical Alerts',   description: 'High priority notifications', disabled: false },
  { key: 'deadlineReminder', label: 'Deadline Reminder', description: 'Remind before due dates',     disabled: false },
  { key: 'statusUpdate',     label: 'Status Update',     description: 'Task status changes',         disabled: false },
  { key: 'adminSystem',      label: 'Admin & System',    description: 'System-level messages',       disabled: true  },
]

type PasswordSettingsKeys = 'currPassword' | 'newPassword' | 'confirmPassword'

const passwordSettings: { key: PasswordSettingsKeys; label: string; disabled: boolean }[] = [
  { key: 'currPassword',     label: 'Current Password',  disabled: false },
  { key: 'newPassword',      label: 'New Password',      disabled: false },
  { key: 'confirmPassword',  label: 'Confirm Password',  disabled: false },
]

// Zod schema for password form validation
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

// Password form
const passwordForm = ref<PasswordFormData>({
  currPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const passwordErrors = ref<Record<PasswordSettingsKeys, string>>({
  currPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const isChangingPassword = ref(false)
const passwordSuccess = ref('')

function resetPasswordForm() {
  passwordForm.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  passwordErrors.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  passwordSuccess.value = ''
}

function copyManualCode(){
  copy(manualCode)

  toast.add({
    title: "Manual Code copied to clipboard. Check the help section",
    color: "success",
    icon: "i-lucide-circle-check",
  });
}

function updateTwoFactor(val: boolean){
  settings.value.twoFactor = val
  toast.add({
    title: val ? '2FA diaktifkan' : '2FA dinonaktifkan',
    description: val 
      ? 'Akun Anda sekarang menggunakan autentikasi dua faktor.'
      : 'Akun Anda tidak lagi menggunakan autentikasi dua faktor.',
    color: val ? 'success' : 'warning',
    icon: val ? 'i-lucide-shield-check' : 'i-lucide-alert-triangle',
  })
}

async function changePassword() {
  // Clear previous errors
  passwordErrors.value = {
    currPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  passwordSuccess.value = ''

  // Validate with Zod
  const result = passwordSchema.safeParse(passwordForm.value)

  if (!result.success) {
    // Format Zod errors for display
    result.error.errors.forEach((error) => {
      const field = error.path[0] as PasswordSettingsKeys
      passwordErrors.value[field] = error.message
    })
    return
  }

  // TODO: Implement actual password change API call
  isChangingPassword.value = true

  try {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1500))

    passwordSuccess.value = 'Password changed successfully!'
    resetPasswordForm()
  } catch (error) {
    passwordErrors.value.currPassword = 'Failed to change password. Please try again.'
  } finally {
    isChangingPassword.value = false
  }
}
</script>