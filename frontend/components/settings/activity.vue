<template>
    <div class="wrapper-vertical-expanded">
      <UAlert
      title="Tips Keamanan"
      description="Jika anda melihat device  atau lokasi yang tidak dikenali, segera lepaskan device tersebut dan ubah password Anda"
      icon="warning"
      color="info"
      variant="subtle"
    />
    <UCard class="w-full">
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900">Devices</h3>
          <p class="text-sm text-gray-600">Your recent actions and logs</p>
        </template>  
        <div class="space-y-4 w-full">
          <div v-for="device in devices" :key="device.id" class="flex items-stretch gap-4 p-4 rounded-lg border border-gray-200 w-full">
            <div class="flex flex-row justify-between items-stretch w-full">
              <div class="flex flex-row gap-4">
                <div class="rounded-full bg-primary-100 w-10 h-10 flex items-center justify-center shrink-0">
                  <UIcon :name="device.icon" class="text-primary-600 size-5" />
                </div>
                <div class="flex-1 min-w-0 w-full">
                  <div class="flex flex-row gap-8">
                    <p class="font-medium text-gray-900">{{ device.title }}</p>
                    <div v-if="device.status == 'online'" class="-translate-y-1">
                      <UBadge :label="capitalizeFirstLetter(device.status)" color="success" variant="soft"></UBadge>
                    </div>
                    <div v-else class="h-8"></div>
                  </div>
                  <h6 class="text-sm flex flex-row gap-2">
                    <UIcon name="pin" color="secondary" class=" size-5 mb-2"  />
                    IP: {{ device.ipAddress }} . {{ device.description || device.location || '' }}
                  </h6>
                  <p class="text-xs text-gray-500 mt-1 flex flex-row items-center gap-2">
                    <UIcon name="clock" color="secondary" class=" size-5"  />
                    {{ displayTime(device.time) }}
                  </p>
                </div>
              </div>
              <div v-if="device.status != 'online'">
                <UButton label="Lepas" variant="ghost" color="error" @click="releaseDevice" />
              </div>
            </div>
          </div>
        </div>
    </UCard>
    <UCard class="w-full">
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900">Tentang Activity Log</h3>
        </template>  
        <article class="flex flex-col gap-4">
          <p>Activity log menampilkan semua device yang pernah login ke akun Anda dalam 30 hari terakhir.</p>
          <p>Jika Anda melihat aktivitas mencurigakan:</p>
          <ul class="pb-3.5 px-6 text-sm text-neutral-900 space-y-1 list-disc list-inside">
            <li>Lepaskan device yang tidak dikenali</li>
            <li>Ubah password akun segera</li>
            <li>Aktifkan Two-Factor Authentication (2FA)</li>
            <li>Hubungi IT Security jika diperlukan</li>
          </ul>
        </article>
    </UCard>
  </div>
  <ConfirmationPopup
    v-model:isOpen="state.open"
    :title="warningReleaseDevice.title"
    :question="warningReleaseDevice.description"
    :confirmText="warningReleaseDevice.confirmText"
    :cancelText="warningReleaseDevice.cancelText"
    variant="danger"
    @confirm="handleReleaseConfirm"
  />
</template>
<script setup lang="ts">
import ConfirmationPopup from '../shared/ConfirmationPopup.vue'

type Device = {
  id: number
  status: 'online' | 'offline'
  deviceType: 'desktop' | 'mobile'
  ipAddress: string
  title: string
  location?: string
  description?: string
  time: Date
  icon: string
}

const releaseDevice = () => {
  state.value.open = true
}

const toast = useToast();

const handleReleaseConfirm = () => {
  console.log('device released')
  releaseDevice()
  
  toast.add({
    title: "Perangkat berhasil dilepas",
    description: "Perangkat Macbook Pro telah dilepaskan dari akun anda",
    color: "success",
    icon: "i-lucide-circle-check",
  })
}

const state = ref({
  open: false
})

const warningReleaseDevice = ref(
  {
    title: 'Apakah anda ingin melepaskan device ini?',
    description: 'Perangkat ini akan dilepas dan perlu login kembali untuk mengakses akun anda.',
    confirmText: 'Keluar',
    cancelText: 'Batalkan',
  },
  
)

const devices: Device[] = [
  {
    id: 1,
    status: "online",
    deviceType: "desktop",
    ipAddress: '192.168.1.123',
    title: 'Safari on Macbook',
    location: 'Jakarta, Indonesia',
    time: new Date(),
    icon: 'i-lucide-laptop',
  },
  {
    id: 2,
    status: "offline",
    deviceType: "mobile",
    ipAddress: '192.168.1.124',
    title: 'Chrome on iPhone',
    location: 'Bekasi, Indonesia',
    time: new Date(Date.now() - 2 * 60 * 60 * 1000), // 2 hours ago
    icon: 'i-lucide-smartphone',
  }
]

const displayTime = (date: Date): string => {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) {
    return diffSecs <= 1 ? 'Just now' : `${diffSecs} seconds ago`
  }
  if (diffMins < 60) {
    return diffMins === 1 ? '1 minute ago' : `${diffMins} minutes ago`
  }
  if (diffHours < 24) {
    return diffHours === 1 ? '1 hour ago' : `${diffHours} hours ago`
  }
  if (diffDays < 7) {
    return diffDays === 1 ? '1 day ago' : `${diffDays} days ago`
  }
  return formatTime(date)
}
const formatTime = (date: Date): string => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

</script>