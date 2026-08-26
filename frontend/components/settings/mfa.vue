<template>
  <div class="space-y-6 w-full">
    <UCard class="w-full overflow-hidden border border-gray-200 dark:border-gray-800 shadow-sm rounded-2xl">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <div
              class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 transition-all duration-300"
              :class="isMfaEnabled 
                ? (isDisabling ? 'bg-red-500/10 text-red-600 dark:text-red-400' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400')
                : 'bg-primary-500/10 text-primary-600 dark:text-primary-400'"
            >
              <UIcon 
                :name="isMfaEnabled ? (isDisabling ? 'i-lucide-shield-alert' : 'i-lucide-shield-check') : 'i-lucide-shield-alert'" 
                class="w-6 h-6" 
              />
            </div>
            <div>
              <h3 class="text-lg font-bold text-white">
                {{ isDisabling ? t('settings.mfa.modalTitle') : t('settings.mfa.title') }}
              </h3>
              <p class="text-xs sm:text-sm text-white/80">
                {{ isDisabling ? t('settings.mfa.modalSubtitle') : t('settings.mfa.subtitle') }}
              </p>
            </div>
          </div>

          <div v-if="!loadingStatus" class="self-start sm:self-center">
            <UBadge
              :color="isMfaEnabled ? (isDisabling ? 'error' : 'success') : 'warning'"
              variant="solid"
              size="md"
              class="font-bold px-3.5 py-1.5 rounded-full shadow-md text-white"
            >
              <template #leading>
                <span 
                  class="w-2 h-2 rounded-full bg-white animate-pulse" 
                />
              </template>
              {{ isMfaEnabled ? (isDisabling ? t('settings.mfa.confirmDisable') : t('settings.mfa.activeProtected')) : t('settings.mfa.notEnabled') }}
            </UBadge>
          </div>
          <USkeleton v-else class="h-7 w-28 rounded-full" />
        </div>
      </template>

      <!-- Loading State -->
      <div v-if="loadingStatus" class="p-6 space-y-4">
        <USkeleton class="h-4 w-3/4" />
        <USkeleton class="h-10 w-48 rounded-xl" />
      </div>

      <!-- State 1 & 2: MFA Disabled Flow (Begin Component / Setup Stepper) -->
      <div v-else-if="!isMfaEnabled" class="space-y-6 p-2">
        <div class="p-4 rounded-xl bg-primary-900/40 dark:bg-primary-950/60 border border-primary-100 dark:border-primary-900/40 text-primary-900 dark:text-primary-200 text-sm flex items-start gap-3">
          <UIcon name="i-lucide-info" class="w-5 h-5 text-primary-600 dark:text-primary-400 shrink-0 mt-0.5" />
          <p>
            {{ t('settings.mfa.infoText') }}
          </p>
        </div>

        <!-- MFA Begin Component (Start Setup) -->
        <div v-if="!setupData" class="pt-2">
          <UButton
            icon="i-lucide-shield-plus"
            color="primary"
            size="lg"
            class="font-bold px-6 py-3 rounded-xl transition-all duration-200 shadow-md cursor-pointer"
            :loading="loading"
            @click="handleSetup"
          >
            {{ t('settings.mfa.enableButton') }}
          </UButton>
        </div>

        <!-- MFA Setup Stepper Component -->
        <div v-else class="space-y-6 border border-gray-200 dark:border-gray-800 rounded-2xl p-6 bg-gray-50/50 dark:bg-gray-900/50">
          <div class="flex flex-col lg:flex-row gap-8 items-start">
            <!-- QR Code Box -->
            <div class="flex flex-col items-center gap-2 bg-gray-100 dark:bg-gray-900 p-4 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm shrink-0">
              <img v-if="qrCodeDataURL" :src="qrCodeDataURL" alt="MFA QR Code" class="w-44 h-44 rounded-lg" />
              <div v-else class="w-44 h-44 flex items-center justify-center text-gray-400 text-sm animate-pulse">
                Membuat QR Code...
              </div>
              <span class="text-md text-gray-500 font-medium">Pindai QR Code</span>
            </div>

            <!-- Steps -->
            <div class="space-y-4 flex-1">
              <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <span class="w-6 h-6 rounded-full bg-primary text-white text-md flex items-center justify-center font-bold">1</span>
                {{ t('settings.mfa.scanTitle') }}
              </h4>
              <p class="text-sm text-gray-600 dark:text-gray-300">
                {{ t('settings.mfa.scanInstructions') }}
              </p>

              <div class="flex items-center gap-2 max-w-md">
                <code class="bg-gray-100 dark:bg-gray-800 px-3 py-2 rounded-xl font-mono text-sm font-semibold text-primary tracking-widest border border-gray-200 dark:border-gray-700 flex-1 truncate select-all">
                  {{ setupData.secret }}
                </code>
                <UButton
                  icon="i-lucide-copy"
                  color="neutral"
                  variant="subtle"
                  class="rounded-xl"
                  :title="t('settings.mfa.copySecret')"
                  @click="copySecret"
                />
              </div>

              <div class="pt-4 border-t border-gray-200 dark:border-gray-800 space-y-4">
                <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                  <span class="w-6 h-6 rounded-full bg-primary text-white text-md flex items-center justify-center font-bold">2</span>
                  {{ t('settings.mfa.verifyTitle') }}
                </h4>

                <div class="max-w-md space-y-3">
                  <UFormField :label="t('settings.mfa.verificationCodeLabel')" :help="t('settings.mfa.verificationCodeHelp')">
                    <UInput
                      v-model="verificationCode"
                      placeholder="000 000"
                      maxlength="6"
                      size="lg"
                      class="w-full font-mono text-center text-xl tracking-[0.3em] font-bold"
                    />
                  </UFormField>

                  <div class="flex gap-3 pt-2">
                    <UButton
                      variant="subtle"
                      color="neutral"
                      class="flex-1 rounded-xl font-semibold"
                      @click="setupData = null"
                    >
                      {{ t('settings.mfa.cancel') }}
                    </UButton>
                    <UButton
                      color="primary"
                      class="flex-1 rounded-xl font-bold transition-all duration-200 shadow-md"
                      :loading="loading"
                      @click="handleVerifySetup"
                    >
                      {{ t('settings.mfa.verifyEnableButton') }}
                    </UButton>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- State 3: MFA Active Component (When not in disable confirmation state) -->
      <div v-else-if="!isDisabling" class="space-y-6 p-2">
        <!-- MFA Active UCard Component (Natural Surface Theme & Full Width) -->
        <UCard
          variant="outline"
          class="w-full bg-[var(--bg-surface)]/60 dark:bg-gray-900/60 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xs"
        >
          <div class="flex items-start gap-4">
            <div class="p-2 rounded-xl bg-emerald-500/10 dark:bg-emerald-500/20 text-emerald-500 dark:text-emerald-400 shrink-0">
              <UIcon name="i-lucide-shield-check" class="w-6 h-6" />
            </div>
            <div class="space-y-1">
              <h4 class="text-base font-bold text-gray-900 dark:text-white">{{ t('settings.mfa.mfaActiveTitle') }}</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('settings.mfa.mfaActiveDesc') }}
              </p>
            </div>
          </div>
        </UCard>

        <div class="pt-2 flex items-center justify-between border-t border-gray-200 dark:border-gray-800">
          <div>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('settings.mfa.disableTitle') }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('settings.mfa.disableDesc') }}</p>
          </div>
          <UButton
            icon="i-lucide-shield-off"
            color="error"
            variant="solid"
            class="font-bold rounded-xl px-5 py-2.5 bg-red-600 hover:bg-red-700 active:bg-red-800 text-white shadow-md transition-all duration-200 cursor-pointer"
            @click="isDisabling = true"
          >
            {{ t('settings.mfa.disableButton') }}
          </UButton>
        </div>
      </div>

      <!-- State 4: MFA Confirm Disable Component (Displayed conditionally when disabling) -->
      <div v-else class="space-y-6 p-4 rounded-2xl border border-red-200 dark:border-red-900/50 bg-red-50/30 dark:bg-red-950/20">
        <div class="flex items-start gap-3 p-4 rounded-xl bg-red-500/20 border border-red-500/30 text-white">
          <UIcon name="i-lucide-alert-triangle" class="w-6 h-6 text-red-400 shrink-0 mt-0.5" />
          <div class="space-y-1">
            <h4 class="text-sm font-bold text-white">{{ t('settings.mfa.modalTitle') }}</h4>
            <p class="text-xs text-white/90">
              {{ t('settings.mfa.modalPrompt') }}
            </p>
          </div>
        </div>

        <div class="max-w-md space-y-4">
          <UFormField :label="t('settings.mfa.passwordLabel')">
            <UInput
              v-model="password"
              type="password"
              :placeholder="t('settings.mfa.passwordPlaceholder')"
              size="lg"
              class="w-full"
              @keyup.enter="handleDisable"
            />
          </UFormField>

          <div class="flex items-center gap-3 pt-2">
            <UButton
              variant="subtle"
              color="neutral"
              class="rounded-xl font-semibold px-5"
              @click="cancelDisable"
            >
              {{ t('settings.mfa.cancelModal') }}
            </UButton>
            <UButton
              color="error"
              class="rounded-xl text-white font-bold px-6 cursor-pointer"
              :loading="loading"
              @click="handleDisable"
            >
              {{ t('settings.mfa.confirmDisable') }}
            </UButton>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { useAuthStore } from '~/stores/auth'

const config = useRuntimeConfig()
const authStore = useAuthStore()
const { t } = useI18n()

const mfaStatus = ref<any>(null)
const loadingStatus = ref(true)
const setupData = ref<any>(null)
const qrCodeDataURL = ref('')
const verificationCode = ref('')
const loading = ref(false)
const isDisabling = ref(false)
const password = ref('')
const toast = useToast()

const isMfaEnabled = computed(() => {
  if (!mfaStatus.value) return false
  return Boolean(mfaStatus.value.is_enabled ?? mfaStatus.value.data?.is_enabled)
})

const fetchStatus = async () => {
  if (!authStore.token) {
    loadingStatus.value = false
    return
  }
  loadingStatus.value = true
  try {
    const response = await $fetch<any>(`${getAuthServiceBaseUrl()}/mfa/status`, {
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    const rawData = response?.data?.data ?? response?.data ?? response
    mfaStatus.value = rawData
  } catch (err) {
    console.error('Failed to fetch MFA status:', err)
  } finally {
    loadingStatus.value = false
  }
}

const handleSetup = async () => {
  if (!authStore.token) return
  loading.value = true
  try {
    const response = await $fetch<any>(`${getAuthServiceBaseUrl()}/mfa/enroll`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { mfa_type: 'TOTP' }
    })
    
    const data = response?.data?.data ?? response?.data ?? response
    setupData.value = data
    if (setupData.value?.qr_code_url) {
      qrCodeDataURL.value = await QRCode.toDataURL(setupData.value.qr_code_url)
    }
  } catch (err: any) {
    toast.add({
      title: 'Setup Gagal',
      description: err.data?.message || err.message || 'Gagal memulai setup MFA.',
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    loading.value = false
  }
}

const copySecret = () => {
  if (!setupData.value?.secret) return
  navigator.clipboard.writeText(setupData.value.secret)
  toast.add({
    title: t('settings.mfa.copiedToast'),
    description: t('settings.mfa.copiedDesc'),
    color: 'success',
    icon: 'i-lucide-check'
  })
}

const handleVerifySetup = async () => {
  if (!authStore.token || !verificationCode.value) return
  loading.value = true
  try {
    await $fetch(`${getAuthServiceBaseUrl()}/mfa/verify`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { code: verificationCode.value }
    })
    
    toast.add({
      title: t('settings.mfa.enabledToast'),
      description: t('settings.mfa.enabledDesc'),
      color: 'success',
      icon: 'i-lucide-circle-check'
    })
    
    setupData.value = null
    verificationCode.value = ''
    await fetchStatus()
  } catch (err: any) {
    toast.add({
      title: 'Verifikasi Gagal',
      description: err.data?.message || err.message || 'Kode verifikasi tidak valid.',
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    loading.value = false
  }
}

const cancelDisable = () => {
  isDisabling.value = false
  password.value = ''
}

const handleDisable = async () => {
  if (!authStore.token || !password.value) return
  loading.value = true
  try {
    await $fetch(`${getAuthServiceBaseUrl()}/mfa/disable`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { password: password.value }
    })
    
    toast.add({
      title: t('settings.mfa.disabledToast'),
      description: t('settings.mfa.disabledDesc'),
      color: 'success',
      icon: 'i-lucide-circle-check'
    })
    
    isDisabling.value = false
    password.value = ''
    await fetchStatus()
  } catch (err: any) {
    toast.add({
      title: 'Gagal Menonaktifkan',
      description: err.data?.message || err.message || 'Password yang Anda masukkan salah.',
      color: 'error',
      icon: 'i-lucide-alert-triangle'
    })
  } finally {
    loading.value = false
  }
}

onMounted(fetchStatus)
</script>
