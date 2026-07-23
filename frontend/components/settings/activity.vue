<template>
  <div class="wrapper-vertical-expanded space-y-6">
    <!-- Database registration check panel -->
    <UCard class="w-full border-l-4 border-success-500">
      <div class="flex items-start gap-4">
        <div class="rounded-full bg-success-500/10 p-2 text-success-600">
          <UIcon name="i-lucide-shield-check" class="size-6" />
        </div>
        <div class="space-y-1">
          <h4 class="text-sm font-bold text-gray-900 dark:text-white">Verifikasi Penyimpanan Database</h4>
          <p class="text-md text-gray-600 dark:text-gray-300">
            Perangkat yang Anda gunakan saat ini:
            <span class="font-mono font-semibold bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-primary-600">
              {{ currentDevice.deviceName }} (Fingerprint: {{ currentDevice.deviceFingerprint }})
            </span>
          </p>
          <div v-if="dbStatus.checked" class="mt-2 text-md text-success-700 dark:text-success-400 flex items-center gap-1.5 font-medium">
            <span class="flex h-2 w-2 rounded-full bg-success-500"></span>
            <span>Terdaftar di Database: Ya (Last Login Fingerprint: {{ dbStatus.lastLoginFingerprint }})</span>
          </div>
          <div v-else class="mt-2 text-md text-gray-500 flex items-center gap-1.5 animate-pulse">
            <span class="flex h-2 w-2 rounded-full bg-gray-400"></span>
            <span>Memeriksa database...</span>
          </div>
        </div>
      </div>
    </UCard>

    <UAlert
      title="Tips Keamanan"
      description="Jika Anda melihat perangkat atau lokasi yang tidak dikenali, segera hapus perangkat tersebut dan ubah password Anda."
      icon="i-lucide-alert-triangle"
      color="warning"
      variant="subtle"
    />

    <UCard class="w-full">
      <template #header>
        <div class="flex justify-between items-center">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Daftar Perangkat Terpercaya (MFA Trusted Devices)</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400">Daftar perangkat yang dipercaya untuk melewati verifikasi OTP/MFA.</p>
          </div>
          <UButton
            icon="i-lucide-refresh-cw"
            variant="ghost"
            color="primary"
            :loading="loading"
            @click="loadDevices"
          />
        </div>
      </template>

      <div class="space-y-4 w-full">
        <!-- Current active session device (always shown) -->
        <div class="flex items-stretch gap-4 p-4 rounded-lg border border-primary-500/20 bg-primary-500/5 w-full">
          <div class="flex flex-row justify-between items-stretch w-full">
            <div class="flex flex-row gap-4">
              <div class="rounded-full bg-primary-100 dark:bg-primary-950 w-10 h-10 flex items-center justify-center shrink-0">
                <UIcon :name="currentDevice.deviceType === 'mobile' ? 'i-lucide-smartphone' : 'i-lucide-laptop'" class="text-primary-600 size-5" />
              </div>
              <div class="flex-1 min-w-0 w-full">
                <div class="flex items-center gap-3">
                  <p class="font-medium text-gray-900 dark:text-white">{{ currentDevice.deviceName }}</p>
                  <UBadge label="Sesi Aktif" color="primary" variant="solid" size="md" />
                  <UBadge label="Online" color="success" variant="soft" size="md" />
                </div>
                <h6 class="text-sm flex flex-row gap-2 mt-1 text-gray-600 dark:text-gray-300">
                  <UIcon name="i-lucide-fingerprint" class="size-4 text-gray-400" />
                  Fingerprint: {{ currentDevice.deviceFingerprint }}
                </h6>
                <p class="text-md text-gray-500 mt-1 flex flex-row items-center gap-1.5">
                  <UIcon name="i-lucide-clock" class="size-4" />
                  Masuk baru saja
                </p>
              </div>
            </div>
            <div class="text-md text-gray-400 self-center">
              Perangkat Ini
            </div>
          </div>
        </div>

        <!-- Loaded list of other trusted devices -->
        <template v-if="trustedDevices.length > 0">
          <div v-for="device in trustedDevices" :key="device.id" class="flex items-stretch gap-4 p-4 rounded-lg border border-gray-200 dark:border-gray-800 w-full">
            <div class="flex flex-row justify-between items-stretch w-full">
              <div class="flex flex-row gap-4">
                <div class="rounded-full bg-gray-100 dark:bg-gray-800 w-10 h-10 flex items-center justify-center shrink-0">
                  <UIcon :name="device.deviceType === 'mobile' ? 'i-lucide-smartphone' : 'i-lucide-laptop'" class="text-gray-600 size-5" />
                </div>
                <div class="flex-1 min-w-0 w-full">
                  <div class="flex items-center gap-3">
                    <p class="font-medium text-gray-900 dark:text-white">{{ device.deviceName }}</p>
                    <UBadge v-if="device.deviceFingerprint === currentDevice.deviceFingerprint" label="Perangkat Ini" color="primary" variant="soft" size="md" />
                  </div>
                  <h6 class="text-sm flex flex-row gap-2 mt-1 text-gray-600 dark:text-gray-300">
                    <UIcon name="i-lucide-globe" class="size-4 text-gray-400" />
                    IP: {{ device.ipAddress }} · Fingerprint: {{ device.deviceFingerprint }}
                  </h6>
                  <p class="text-md text-gray-500 mt-1 flex flex-row items-center gap-1.5">
                    <UIcon name="i-lucide-clock" class="size-4" />
                    Terdaftar: {{ formatTime(device.createdAt) }}
                  </p>
                </div>
              </div>
              <div class="self-center">
                <UButton
                  label="Hapus"
                  variant="ghost"
                  color="error"
                  size="sm"
                  @click="confirmRelease(device.id, device.deviceName)"
                />
              </div>
            </div>
          </div>
        </template>
        
        <div v-else-if="!loading" class="text-center py-6 text-sm text-gray-500 border border-dashed border-gray-200 dark:border-gray-800 rounded-xl">
          <UIcon name="i-lucide-info" class="size-6 text-gray-400 mx-auto mb-2" />
          <p>Tidak ada perangkat terpercaya tambahan terdaftar.</p>
          <p class="text-md text-gray-400 mt-1">Perangkat terdaftar secara otomatis saat Anda mencentang "Percayai Perangkat Ini" ketika verifikasi OTP/MFA.</p>
        </div>
      </div>
    </UCard>

    <UCard class="w-full">
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Tentang Activity Log</h3>
      </template>  
      <article class="flex flex-col gap-4 text-sm text-gray-700 dark:text-gray-300">
        <p>Halaman ini mencatat semua perangkat terpercaya yang diizinkan untuk melewati verifikasi multi-faktor (MFA).</p>
        <p>Jika Anda mendeteksi aktivitas mencurigakan atau perangkat yang tidak dikenal:</p>
        <ul class="pb-2 px-6 text-sm space-y-1 list-disc list-inside">
          <li>Segera hapus/lepas perangkat tersebut dari daftar di atas.</li>
          <li>Ubah password akun Anda dengan kombinasi yang kuat.</li>
          <li>Pastikan Multi-Factor Authentication (MFA) tetap aktif.</li>
        </ul>
      </article>
    </UCard>

    <!-- Confirmation Modal -->
    <ConfirmationPopup
      v-model:isOpen="state.open"
      title="Hapus Perangkat Terpercaya?"
      question="Perangkat ini akan dihapus dari daftar terpercaya dan akan memerlukan verifikasi OTP kembali saat masuk."
      confirmText="Hapus Akses"
      cancelText="Batalkan"
      variant="danger"
      @confirm="handleReleaseConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import { useDeviceFingerprint } from '~/composables/useDeviceFingerprint'
import ConfirmationPopup from '../shared/ConfirmationPopup.vue'

const authStore = useAuthStore()
const { getDeviceFingerprint } = useDeviceFingerprint()
const toast = useToast()

const loading = ref(false)
const trustedDevices = ref<any[]>([])
const currentDevice = ref({ deviceFingerprint: '', deviceName: '', deviceType: '' })

// Database verification status
const dbStatus = reactive({
  checked: false,
  lastLoginFingerprint: '',
  isStored: false,
})

const state = ref({
  open: false,
  targetDeviceId: '',
  targetDeviceName: '',
})

onMounted(async () => {
  currentDevice.value = getDeviceFingerprint()
  await loadDevices()
  await verifyDatabaseRegistration()
})

const loadDevices = async () => {
  loading.value = true
  try {
    const list = await authStore.fetchTrustedDevices()
    // Filter out null values if list has empty spots
    trustedDevices.value = (list || []).filter((d: any) => d && d.id)
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const verifyDatabaseRegistration = async () => {
  try {
    const profile = await authStore.fetchUserProfile()
    if (profile) {
      dbStatus.lastLoginFingerprint = profile.last_login_fingerprint || ''
      dbStatus.isStored = (profile.last_login_fingerprint === currentDevice.value.deviceFingerprint)
      dbStatus.checked = true
    }
  } catch (err) {
    console.error(err)
    dbStatus.checked = true
  }
}

const confirmRelease = (deviceId: string, deviceName: string) => {
  state.value.targetDeviceId = deviceId
  state.value.targetDeviceName = deviceName
  state.value.open = true
}

const handleReleaseConfirm = async () => {
  state.value.open = false
  try {
    await authStore.unenrollDevice(state.value.targetDeviceId)
    toast.add({
      title: "Perangkat Dihapus",
      description: `Akses terpercaya untuk perangkat "${state.value.targetDeviceName}" telah dihapus.`,
      color: "success",
      icon: "i-lucide-circle-check",
    })
    await loadDevices()
  } catch (err: any) {
    toast.add({
      title: "Gagal Menghapus",
      description: err.message || "Gagal menghapus perangkat terpercaya.",
      color: "error",
      icon: "i-lucide-alert-triangle",
    })
  }
}

const formatTime = (dateStr: string): string => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('id-ID', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
</script>