<template>
  <UDashboardPage>
    <UDashboardHeader title="Trusted Devices" />

    <UDashboardPanelContent>
      <div class="space-y-6">
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Registered Devices</h3>
              <UButton icon="i-lucide-qr-code" @click="handleGenerateQR">Register New Device</UButton>
            </div>
          </template>

          <UTable :rows="devices" :columns="columns">
            <template #created_at-data="{ row }">
              {{ new Date(row.original.created_at).toLocaleDateString() }}
            </template>
            <template #actions-data="{ row }">
              <UButton
                color="error"
                variant="ghost"
                icon="i-lucide-trash"
                @click="handleRemove(row.original.id)"
              />
            </template>
          </UTable>
        </UCard>

        <UModal v-model="showQRModal">
          <UCard>
            <template #header>Register Device</template>
            <div class="flex flex-col items-center py-6">
              <p class="mb-4 text-center">Scan this QR code with your mobile app to trust this device.</p>
              <div class="bg-white p-4 rounded-lg shadow-inner">
                <img v-if="qrCodeDataURL" :src="qrCodeDataURL" alt="Enrollment QR" />
              </div>
              <p v-if="qrData" class="mt-4 text-sm text-gray-500 italic">Expires at: {{ new Date(qrData.expires_at).toLocaleString() }}</p>
            </div>
            <template #footer>
              <UButton block @click="() => { showQRModal = false }">Close</UButton>
            </template>
          </UCard>
        </UModal>
      </div>
    </UDashboardPanelContent>
  </UDashboardPage>
</template>

<script setup lang="ts">
import QRCode from 'qrcode'

import type { TableColumn } from '@nuxt/ui'

interface Device {
  id: string
  device_name: string
  ip_address: string
  created_at: string
}

const config = useRuntimeConfig()
const devices = ref<Device[]>([])
const showQRModal = ref(false)
const qrData = ref<any>(null)
const qrCodeDataURL = ref('')
const toast = useToast()

const columns: TableColumn<Device>[] = [
  { accessorKey: 'device_name', header: 'Device' },
  { accessorKey: 'ip_address', header: 'IP Address' },
  { accessorKey: 'created_at', header: 'Added On' },
  { accessorKey: 'actions', header: '' }
]

const fetchDevices = async () => {
  try {
    const { data } = await $fetch<any>(`${config.public.apiBase}/devices`)
    devices.value = data
  } catch (err) {}
}

const handleGenerateQR = async () => {
  try {
    const { data } = await $fetch<any>(`${config.public.apiBase}/devices/enroll`, { method: 'POST' })
    qrData.value = data
    if (qrData.value?.qr_code_url) {
      qrCodeDataURL.value = await QRCode.toDataURL(qrData.value.qr_code_url)
    }
    showQRModal.value = true
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.message, color: 'error' })
  }
}

const handleRemove = async (id: string) => {
  if (!confirm('Are you sure you want to remove this trusted device?')) return

  try {
    await $fetch(`${config.public.apiBase}/devices/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Device removed' })
    await fetchDevices()
  } catch (err: any) {
    toast.add({ title: 'Error', description: err.message, color: 'error' })
  }
}

onMounted(fetchDevices)
</script>
