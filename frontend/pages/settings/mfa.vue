<template>
  <UDashboardPage>
    <UDashboardHeader title="Security Settings" />

    <UDashboardPanelContent>
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Two-Factor Authentication (MFA)</h3>
            <UBadge :color="mfaStatus?.is_enabled ? 'success' : 'neutral'">
              {{ mfaStatus?.is_enabled ? 'Enabled' : 'Disabled' }}
            </UBadge>
          </div>
        </template>

        <div v-if="!mfaStatus?.is_enabled" class="space-y-6">
          <p class="text-gray-600">
            Secure your account by adding an extra layer of security. Once enabled, you'll need to enter a code from your authenticator app to log in.
          </p>

          <div v-if="!setupData" class="flex justify-center">
            <UButton color="primary" @click="handleSetup">Enable MFA</UButton>
          </div>

          <div v-else class="space-y-4 flex flex-col items-center">
            <p class="font-medium">Scan this QR code with your authenticator app:</p>
            <div class="bg-white p-4 rounded-lg shadow-inner">
              <img :src="qrCodeDataURL" alt="QR Code" />
            </div>
            <p class="text-sm text-gray-500">Or enter this secret manually: <code class="bg-gray-100 px-2 py-1 rounded">{{ setupData.secret }}</code></p>

            <div class="w-full max-w-xs space-y-4 mt-6">
              <UFormField label="Enter Verification Code" help="Enter the 6-digit code from your app to confirm setup.">
                <UInput v-model="verificationCode" placeholder="000000" maxlength="6" />
              </UFormField>
              <UButton block color="primary" :loading="loading" @click="handleVerifySetup">Confirm and Enable</UButton>
            </div>
          </div>
        </div>

        <div v-else class="space-y-4">
          <p class="text-gray-600">Two-factor authentication is currently enabled for your account.</p>
          <UButton color="red" variant="subtle" @click="showDisableModal = true">Disable MFA</UButton>
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
              <UButton color="neutral" variant="ghost" @click="showDisableModal = false">Cancel</UButton>
              <UButton color="red" :loading="loading" @click="handleDisable">Confirm Disable</UButton>
            </div>
          </template>
        </UCard>
      </UModal>
    </UDashboardPanelContent>
  </UDashboardPage>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'

const config = useRuntimeConfig()
const mfaStatus = ref<any>(null)
const setupData = ref<any>(null)
const qrCodeDataURL = ref('')
const verificationCode = ref('')
const loading = ref(false)
const showDisableModal = ref(false)
const password = ref('')
const toast = useToast()

const fetchStatus = async () => {
  try {
    const { data } = await $fetch<any>(`${config.public.apiBase}/mfa/status`)
    mfaStatus.value = data
  } catch (err) {}
}

const handleSetup = async () => {
  try {
    const response = await $fetch<any>(`${config.public.apiBase}/mfa/enroll`, {
      method: 'POST',
      body: { mfa_type: 'TOTP' }
    })
    setupData.value = response.data
    if (setupData.value?.qr_code_url) {
      qrCodeDataURL.value = await QRCode.toDataURL(setupData.value.qr_code_url)
    }
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.message, color: 'red' })
  }
}

// I should probably fix EnrollMFA to return the data.
onMounted(fetchStatus)

const handleVerifySetup = async () => {
  loading.value = true
  try {
    await $fetch(`${config.public.apiBase}/mfa/verify`, {
      method: 'POST',
      body: { code: verificationCode.value }
    })
    toast.add({ title: 'Success', description: 'MFA enabled successfully' })
    setupData.value = null
    await fetchStatus()
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.message, color: 'red' })
  } finally {
    loading.value = false
  }
}

const handleDisable = async () => {
  loading.value = true
  try {
    await $fetch(`${config.public.apiBase}/mfa/disable`, {
      method: 'POST',
      body: { password: password.value }
    })
    toast.add({ title: 'Success', description: 'MFA disabled successfully' })
    showDisableModal.value = false
    password.value = ''
    await fetchStatus()
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.message, color: 'red' })
  } finally {
    loading.value = false
  }
}
</script>
