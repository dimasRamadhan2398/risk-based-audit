<template>
  <UDashboardPage>
    <UDashboardHeader title="Security Settings" />

    <UDashboardPanelContent>
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Two-Factor Authentication (MFA)</h3>
            <UBadge v-if="!loadingStatus" :color="isMfaEnabled ? 'success' : 'neutral'">
              {{ isMfaEnabled ? 'Enabled' : 'Disabled' }}
            </UBadge>
            <USkeleton v-else class="h-6 w-20 rounded-full" />
          </div>
        </template>

        <div v-if="loadingStatus" class="p-6 space-y-4">
          <USkeleton class="h-4 w-3/4" />
          <USkeleton class="h-10 w-48 rounded-xl" />
        </div>

        <div v-else-if="!isMfaEnabled" class="space-y-6">
          <p class="text-gray-600 font-medium">
            Secure your account by adding an extra layer of security. Once enabled, you'll need to enter a code from your authenticator app to log in.
          </p>

          <div v-if="!setupData" class="flex justify-start">
            <UButton color="primary" size="lg" class="font-bold" :loading="loading" @click="handleSetup">Enable MFA</UButton>
          </div>

          <div v-else class="space-y-4 flex flex-col items-center">
            <p class="font-medium">Scan this QR code with your authenticator app:</p>
            <div class="bg-white p-4 rounded-lg shadow-inner">
              <img :src="qrCodeDataURL" alt="QR Code" />
            </div>
            <p class="text-sm text-gray-500">Or enter this secret manually: <code class="bg-gray-100 px-2 py-1 rounded">{{ setupData.secret }}</code></p>

            <div class="w-full max-w-md space-y-4 mt-6">
              <UFormField label="Enter Verification Code" help="Enter the 6-digit code from your app to confirm setup.">
                <UInput v-model="verificationCode" placeholder="000000" maxlength="6" />
              </UFormField>
              <UButton block color="primary" size="lg" class="font-bold" :loading="loading" @click="handleVerifySetup">Confirm and Enable</UButton>
            </div>
          </div>
        </div>

        <div v-else class="space-y-4">
          <p class="text-gray-600 font-medium">Two-factor authentication is currently enabled for your account.</p>
          <UButton color="error" variant="subtle" class="font-bold" @click="() => { showDisableModal = true }">Disable MFA</UButton>
        </div>
      </UCard>

      <UModal v-model="showDisableModal">
        <UCard>
          <template #header>Disable Two-Factor Authentication</template>
          <p class="mb-4">Please enter your password to confirm disabling MFA.</p>
          <UFormField label="Password">
            <UInput v-model="password" type="password" />
          </UFormField>
          <template #footer>
            <div class="flex justify-end gap-3">
              <UButton color="neutral" variant="ghost" @click="() => { showDisableModal = false }">Cancel</UButton>
              <UButton color="error" class="font-bold" :loading="loading" @click="handleDisable">Confirm Disable</UButton>
            </div>
          </template>
        </UCard>
      </UModal>
    </UDashboardPanelContent>
  </UDashboardPage>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { useAuthStore } from '~/stores/auth'

const config = useRuntimeConfig()
const authStore = useAuthStore()

const mfaStatus = ref<any>(null)
const loadingStatus = ref(true)
const setupData = ref<any>(null)
const qrCodeDataURL = ref('')
const verificationCode = ref('')
const loading = ref(false)
const showDisableModal = ref(false)
const password = ref('')
const toast = useToast()

const isMfaEnabled = computed(() => {
  if (!mfaStatus.value) return false
  return Boolean(mfaStatus.value.is_enabled ?? mfaStatus.value.data?.is_enabled)
})

const getAuthBaseUrl = () => {
  if (import.meta.client) {
    const hostname = window.location.hostname
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
      return 'http://localhost:8080/api/v1'
    }
  }
  let url = (config.public.authServiceBaseUrl as string) || '/api/v1'
  if (!url.startsWith('http://') && !url.startsWith('https://') && !url.startsWith('/')) {
    url = `/${url}`
  }
  return url.replace(/\/$/, '')
}

const fetchStatus = async () => {
  if (!authStore.token) {
    loadingStatus.value = false
    return
  }
  loadingStatus.value = true
  try {
    const response = await $fetch<any>(`${getAuthBaseUrl()}/mfa/status`, {
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
    const response = await $fetch<any>(`${getAuthBaseUrl()}/mfa/enroll`, {
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
    toast.add({ title: 'Error', description: err.data?.message || err.message, color: 'error' })
  } finally {
    loading.value = false
  }
}

onMounted(fetchStatus)

const handleVerifySetup = async () => {
  if (!authStore.token) return
  loading.value = true
  try {
    await $fetch(`${getAuthBaseUrl()}/mfa/verify`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { code: verificationCode.value }
    })
    toast.add({ title: 'Success', description: 'MFA enabled successfully' })
    setupData.value = null
    await fetchStatus()
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.data?.message || err.message, color: 'error' })
  } finally {
    loading.value = false
  }
}

const handleDisable = async () => {
  if (!authStore.token) return
  loading.value = true
  try {
    await $fetch(`${getAuthBaseUrl()}/mfa/disable`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { password: password.value }
    })
    toast.add({ title: 'Success', description: 'MFA disabled successfully' })
    showDisableModal.value = false
    password.value = ''
    await fetchStatus()
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.data?.message || err.message, color: 'error' })
  } finally {
    loading.value = false
  }
}
</script>
