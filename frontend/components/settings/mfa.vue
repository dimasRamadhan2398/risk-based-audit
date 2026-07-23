<template>
  <div class="space-y-6">
    <UCard class="w-full">
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Autentikasi Dua Faktor (2FA / MFA)</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Tingkatkan keamanan akun Anda dengan verifikasi tambahan.</p>
          </div>
          <UBadge
            :label="mfaStatus?.is_enabled ? 'Aktif' : 'Nonaktif'"
            :color="mfaStatus?.is_enabled ? 'success' : 'neutral'"
            variant="soft"
            size="md"
          />
        </div>
      </template>

      <!-- MFA is Disabled: Show Setup Flow -->
      <div v-if="!mfaStatus?.is_enabled" class="space-y-6">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          Autentikasi dua faktor menambahkan langkah keamanan tambahan saat masuk ke akun Anda. Setelah diaktifkan, Anda harus memasukkan kode 6 digit dari aplikasi authenticator Anda (seperti Google Authenticator atau Microsoft Authenticator).
        </p>

        <!-- Initial state: Click to start setup -->
        <div v-if="!setupData" class="flex justify-start">
          <UButton
            icon="i-lucide-shield-plus"
            class="text-white hover:opacity-90 font-bold transition-all duration-200"
            :style="{ background: 'linear-gradient(135deg, var(--color-secondary-500), var(--color-primary-500))', bomdhadow: '0 4px 12px -2px color-mix(in srgb, var(--color-secondary-500) 25%, transparent)' }"
            :loading="loading"
            @click="handleSetup"
          >
            Aktifkan Autentikasi Dua Faktor
          </UButton>
        </div>

        <!-- Setup in progress: Show QR and Verification Input -->
        <div v-else class="space-y-6 border border-[var(--border-main)] rounded-xl p-6 bg-[var(--bg-surface)]">
          <div class="flex flex-col md:flex-row gap-6 items-center">
            <!-- QR Code container -->
            <div class="bg-white p-4 rounded-xl border border-gray-200 shadow-inner flex-shrink-0">
              <img v-if="qrCodeDataURL" :src="qrCodeDataURL" alt="QR Code" class="w-40 h-40" />
              <div v-else class="w-40 h-40 flex items-center justify-center text-gray-400 text-md animate-pulse">
                Membuat QR Code...
              </div>
            </div>

            <!-- Steps description -->
            <div class="space-y-3 flex-1">
              <h4 class="text-sm font-bold text-gray-900 dark:text-white">Langkah Aktivasi:</h4>
              <ol class="list-decimal list-inside space-y-2 text-md text-gray-600 dark:text-gray-300">
                <li>Buka aplikasi authenticator Anda di smartphone Anda.</li>
                <li>Pindai kode QR di sebelah kiri, atau masukkan kunci rahasia berikut secara manual:</li>
                <li class="list-none pl-4">
                  <code class="bg-[var(--bg-main)] px-2.5 py-1 rounded-md font-mono text-md font-semibold text-primary-500 tracking-wider inline-block mt-1 select-all border border-[var(--border-main)]">
                    {{ setupData.secret }}
                  </code>
                </li>
                <li>Masukkan kode 6 digit yang dihasilkan oleh aplikasi authenticator Anda di bawah ini untuk mengonfirmasi.</li>
              </ol>
            </div>
          </div>

          <!-- Code input form -->
          <div class="border-t border-[var(--border-main)] pt-6 max-w-sm">
            <UFormField label="Kode Verifikasi 6-Digit" help="Masukkan kode 6 digit dari aplikasi authenticator Anda.">
              <UInput
                v-model="verificationCode"
                placeholder="000000"
                maxlength="6"
                class="w-full font-mono text-center text-lg tracking-widest font-bold"
                size="lg"
                :ui="{
                  base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 text-center',
                }"
              />
            </UFormField>

            <div class="flex gap-3 mt-4">
              <UButton
                variant="ghost"
                color="neutral"
                class="flex-1 rounded-xl"
                @click="() => {setupData = null}"
              >
                Batal
              </UButton>
              <UButton
                class="flex-1 rounded-xl text-white font-bold transition-all duration-200"
                :style="{ background: 'linear-gradient(135deg, var(--color-secondary-500), var(--color-primary-500))' }"
                :loading="loading"
                @click="handleVerifySetup"
              >
                Konfirmasi & Aktifkan
              </UButton>
            </div>
          </div>
        </div>
      </div>

      <!-- MFA is Enabled: Show Status and Disable option -->
      <div v-else class="space-y-6">
        <div class="flex items-start gap-4 p-4 rounded-xl bg-success-500/10 border border-success-500/20 text-success-700 dark:text-success-400">
          <UIcon name="i-lucide-shield-check" class="size-6 shrink-0 mt-0.5 text-success-600" />
          <div class="space-y-1">
            <p class="text-sm font-bold">Autentikasi dua faktor (MFA) saat ini aktif.</p>
            <p class="text-md">Akun Anda dilindungi dengan langkah verifikasi ekstra menggunakan TOTP (Aplikasi Authenticator).</p>
          </div>
        </div>

        <div class="pt-2">
          <UButton
            variant="ghost"
            color="error"
            class="font-bold border border-red-500/20 hover:bg-red-500/10 rounded-xl"
            @click="() => {showDisableModal = true}"
          >
            Nonaktifkan MFA
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Disable confirmation modal -->
    <UModal v-model="showDisableModal">
      <UCard class="max-w-md w-full">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center text-red-600">
              <UIcon name="i-lucide-shield-alert" class="size-5" />
            </div>
            <h3 class="text-base font-bold text-gray-900 dark:text-white">Nonaktifkan Keamanan MFA?</h3>
          </div>
        </template>

        <div class="space-y-4">
          <p class="text-sm text-gray-600 dark:text-gray-300">
            Untuk menonaktifkan Autentikasi Dua Faktor, silakan masukkan password akun Anda untuk konfirmasi keamanan.
          </p>

          <UFormField label="Konfirmasi Password">
            <UInput
              v-model="password"
              type="password"
              placeholder="Masukkan password Anda"
              class="w-full"
              :ui="{
                base: 'bg-[var(--bg-main)] border-[var(--border-main)] text-[var(--text-main)] rounded-xl py-3 px-4',
              }"
            />
          </UFormField>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <UButton
              variant="ghost"
              color="neutral"
              class="rounded-xl"
              @click="showDisableModal = false; password = ''"
            >
              Batalkan
            </UButton>
            <UButton
              color="error"
              class="rounded-xl text-white font-bold"
              :loading="loading"
              @click="handleDisable"
            >
              Nonaktifkan MFA
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'
import { useAuthStore } from '~/stores/auth'

const config = useRuntimeConfig()
const authStore = useAuthStore()

const mfaStatus = ref<any>(null)
const setupData = ref<any>(null)
const qrCodeDataURL = ref('')
const verificationCode = ref('')
const loading = ref(false)
const showDisableModal = ref(false)
const password = ref('')
const toast = useToast()

const fetchStatus = async () => {
  if (!authStore.token) return
  try {
    const response = await $fetch<any>(`${config.public.authServiceBaseUrl}/mfa/status`, {
      headers: { Authorization: `Bearer ${authStore.token}` }
    })
    mfaStatus.value = response.data ?? response
  } catch (err) {
    console.error('Failed to fetch MFA status:', err)
  }
}

const handleSetup = async () => {
  if (!authStore.token) return
  loading.value = true
  try {
    const response = await $fetch<any>(`${config.public.authServiceBaseUrl}/mfa/enroll`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { mfa_type: 'TOTP' }
    })
    
    setupData.value = response.data ?? response
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

const handleVerifySetup = async () => {
  if (!authStore.token || !verificationCode.value) return
  loading.value = true
  try {
    await $fetch(`${config.public.authServiceBaseUrl}/mfa/verify`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { code: verificationCode.value }
    })
    
    toast.add({
      title: 'MFA Diaktifkan',
      description: 'Autentikasi Dua Faktor berhasil diaktifkan.',
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

const handleDisable = async () => {
  if (!authStore.token || !password.value) return
  loading.value = true
  try {
    await $fetch(`${config.public.authServiceBaseUrl}/mfa/disable`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authStore.token}` },
      body: { password: password.value }
    })
    
    toast.add({
      title: 'MFA Dinonaktifkan',
      description: 'Autentikasi dua faktor akun Anda telah dinonaktifkan.',
      color: 'success',
      icon: 'i-lucide-circle-check'
    })
    
    showDisableModal.value = false
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
