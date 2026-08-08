<template>
  <div class="space-y-6 max-w-4xl">
    <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800 shadow-sm rounded-2xl">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <div
              class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 transition-all duration-300"
              :class="isMfaEnabled ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-primary-500/10 text-primary-600 dark:text-primary-400'"
            >
              <UIcon :name="isMfaEnabled ? 'i-lucide-shield-check' : 'i-lucide-shield-alert'" class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">Autentikasi Dua Faktor (2FA / MFA)</h3>
              <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400">
                Tingkatkan keamanan akun Anda menggunakan aplikasi authenticator TOTP.
              </p>
            </div>
          </div>

          <div v-if="!loadingStatus" class="self-start sm:self-center">
            <UBadge
              :color="isMfaEnabled ? 'success' : 'warning'"
              variant="subtle"
              size="md"
              class="font-semibold px-3 py-1 rounded-full shadow-xs"
            >
              <template #leading>
                <span class="w-2 h-2 rounded-full" :class="isMfaEnabled ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'" />
              </template>
              {{ isMfaEnabled ? 'Aktif & Terlindungi' : 'Belum Diaktifkan' }}
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

      <!-- MFA Disabled State (Setup Flow) -->
      <div v-else-if="!isMfaEnabled" class="space-y-6 p-2">
        <div class="p-4 rounded-xl bg-primary-900/50 dark:bg-primary-950/60 border border-primary-100 dark:border-primary-900/40 text-primary-900 dark:text-primary-200 text-sm flex items-start gap-3">
          <UIcon name="i-lucide-info" class="w-5 h-5 text-primary-600 dark:text-primary-400 shrink-0 mt-0.5" />
          <p>
            Autentikasi dua faktor menambahkan verifikasi ekstra saat masuk. Setelah diaktifkan, Anda akan memerlukan kode 6-digit dari aplikasi authenticator seperti <strong>Google Authenticator</strong> atau <strong>Microsoft Authenticator</strong>.
          </p>
        </div>

        <!-- Start Setup Button -->
        <div v-if="!setupData" class="pt-2">
          <UButton
            icon="i-lucide-shield-plus"
            color="primary"
            size="lg"
            class="font-bold px-6 py-3 rounded-xl transition-all duration-200 shadow-md"
            :loading="loading"
            @click="handleSetup"
          >
            Aktifkan Autentikasi Dua Faktor
          </UButton>
        </div>

        <!-- Setup Stepper -->
        <div v-else class="space-y-6 border border-gray-200 dark:border-gray-800 rounded-2xl p-6 bg-gray-50/50 dark:bg-gray-900/50">
          <div class="flex flex-col lg:flex-row gap-8 items-start">
            <!-- QR Code Box -->
            <div class="flex flex-col items-center gap-2 bg-white dark:bg-gray-950 p-4 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm shrink-0">
              <img v-if="qrCodeDataURL" :src="qrCodeDataURL" alt="MFA QR Code" class="w-44 h-44 rounded-lg" />
              <div v-else class="w-44 h-44 flex items-center justify-center text-gray-400 text-sm animate-pulse">
                Membuat QR Code...
              </div>
              <span class="text-xs text-gray-500 font-medium">Pindai QR Code</span>
            </div>

            <!-- Steps -->
            <div class="space-y-4 flex-1">
              <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <span class="w-6 h-6 rounded-full bg-primary text-white text-xs flex items-center justify-center font-bold">1</span>
                Pindai atau Masukkan Kunci Rahasia
              </h4>
              <p class="text-sm text-gray-600 dark:text-gray-300">
                Buka aplikasi authenticator di smartphone Anda dan pindai QR Code di sebelah kiri. Jika tidak bisa memindai, gunakan kunci berikut:
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
                  title="Salin Kunci"
                  @click="copySecret"
                />
              </div>

              <div class="pt-4 border-t border-gray-200 dark:border-gray-800 space-y-4">
                <h4 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
                  <span class="w-6 h-6 rounded-full bg-primary text-white text-xs flex items-center justify-center font-bold">2</span>
                  Verifikasi Kode 6-Digit
                </h4>

                <div class="max-w-xs space-y-3">
                  <UFormField label="Kode Verifikasi" help="Masukkan kode 6 digit dari aplikasi Anda.">
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
                      Batal
                    </UButton>
                    <UButton
                      color="primary"
                      class="flex-1 rounded-xl font-bold transition-all duration-200 shadow-md"
                      :loading="loading"
                      @click="handleVerifySetup"
                    >
                      Verifikasi & Aktifkan
                    </UButton>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- MFA Enabled State -->
      <div v-else class="space-y-6 p-2">
        <div class="flex items-start gap-4 p-5 rounded-2xl bg-emerald-50/70 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/50 text-emerald-900 dark:text-emerald-200 shadow-xs">
          <UIcon name="i-lucide-shield-check" class="w-7 h-7 shrink-0 text-emerald-600 dark:text-emerald-400 mt-0.5" />
          <div class="space-y-1">
            <h4 class="text-base font-bold text-emerald-950 dark:text-emerald-100">Keamanan MFA Aktif</h4>
            <p class="text-sm text-emerald-800 dark:text-emerald-300">
              Akun Anda saat ini terlindungi dengan verifikasi tambahan berbasis TOTP (Aplikasi Authenticator). Setiap kali login, Anda akan diminta memasukkan kode verifikasi 6 digit.
            </p>
          </div>
        </div>

        <div class="pt-2 flex items-center justify-between border-t border-gray-100 dark:border-gray-800">
          <div>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">Nonaktifkan Keamanan 2FA</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">Menonaktifkan 2FA akan mengurangi perlindungan keamanan akun Anda.</p>
          </div>
          <UButton
            icon="i-lucide-shield-off"
            color="error"
            variant="soft"
            class="font-bold rounded-xl px-4 py-2 hover:bg-red-500/20 transition-all duration-200"
            @click="showDisableModal = true"
          >
            Nonaktifkan MFA
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Disable Modal -->
    <UModal v-model="showDisableModal">
      <UCard class="max-w-md w-full rounded-2xl">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-red-500/10 text-red-600 flex items-center justify-center shrink-0">
              <UIcon name="i-lucide-alert-triangle" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-base font-bold text-gray-900 dark:text-white">Konfirmasi Nonaktifkan MFA</h3>
              <p class="text-xs text-gray-500 dark:text-gray-400">Verifikasi password Anda sebelum melanjutkan.</p>
            </div>
          </div>
        </template>

        <div class="space-y-4 py-2">
          <p class="text-sm text-gray-600 dark:text-gray-300">
            Demi alasan keamanan, silakan masukkan password akun Anda untuk mengonfirmasi penonaktifan MFA.
          </p>

          <UFormField label="Password Akun">
            <UInput
              v-model="password"
              type="password"
              placeholder="Masukkan password Anda"
              size="lg"
              class="w-full"
            />
          </UFormField>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <UButton
              variant="subtle"
              color="neutral"
              class="rounded-xl font-semibold"
              @click="showDisableModal = false; password = ''"
            >
              Batalkan
            </UButton>
            <UButton
              color="error"
              class="rounded-xl text-white font-bold px-5"
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

const fetchStatus = async () => {
  if (!authStore.token) {
    loadingStatus.value = false
    return
  }
  loadingStatus.value = true
  try {
    const response = await $fetch<any>(`${config.public.authServiceBaseUrl}/mfa/status`, {
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
    const response = await $fetch<any>(`${config.public.authServiceBaseUrl}/mfa/enroll`, {
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
    title: 'Tersalin',
    description: 'Kunci rahasia berhasil disalin ke papan klip.',
    color: 'success',
    icon: 'i-lucide-check'
  })
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
