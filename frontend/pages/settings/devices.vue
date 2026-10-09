<template>
  <UDashboardPage>
    <UDashboardHeader :title="t('settings.devices.title')" />

    <UDashboardPanelContent>
      <div class="space-y-6">
        <UCard>
          <template #header>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h3 class="text-lg font-semibold">{{ t('settings.devices.registeredDevices') }}</h3>
              <UButton icon="i-lucide-qr-code" class="w-full sm:w-auto justify-center" @click="handleGenerateQR">{{ t('settings.devices.registerNewDevice') }}</UButton>
            </div>
          </template>

          <UTable :rows="devices" :columns="columns">
            <template #created_at-data="{ row }">
              {{ new Date(row.original.created_at).toLocaleDateString(locale === 'id' ? 'id-ID' : 'en-US') }}
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

        <UModal v-model="showQRModal" :ui="{ content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-md' }">
          <UCard>
            <template #header>{{ t('settings.devices.modal.title') }}</template>
            <div class="flex flex-col items-center py-6">
              <p class="mb-4 text-center">{{ t('settings.devices.modal.scanDesc') }}</p>
              <div class="bg-white p-4 rounded-lg shadow-inner">
                <img v-if="qrCodeDataURL" :src="qrCodeDataURL" :alt="t('settings.devices.modal.qrAlt')" />
              </div>
              <p v-if="qrData" class="mt-4 text-sm text-gray-500 italic">{{ t('settings.devices.modal.expiresAt', { time: new Date(qrData.expires_at).toLocaleString(locale === 'id' ? 'id-ID' : 'en-US') }) }}</p>
            </div>
            <template #footer>
              <UButton block @click="() => { showQRModal = false }">{{ t('common.close') }}</UButton>
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
import { extractErrorMessage } from '~/utils/error'

interface Device {
  id: string
  device_name: string
  ip_address: string
  created_at: string
}

const config = useRuntimeConfig()
const { t, locale } = useI18n()
const devices = ref<Device[]>([])
const showQRModal = ref(false)
const qrData = ref<any>(null)
const qrCodeDataURL = ref('')
const toast = useToast()

const columns = computed<TableColumn<Device>[]>(() => [
  { accessorKey: 'device_name', header: t('settings.devices.table.device') },
  { accessorKey: 'ip_address', header: t('settings.devices.table.ipAddress') },
  { accessorKey: 'created_at', header: t('settings.devices.table.addedOn') },
  { accessorKey: 'actions', header: '' }
])

const fetchDevices = async () => {
  try {
    const { data } = await $fetch<any>(`${config.public.apiBase}/devices`)
    devices.value = data
  } catch (err: any) {
    console.error('Failed to fetch devices:', err)
  }
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
    toast.add({ title: t('common.error'), description: extractErrorMessage(err, t('settings.devices.errorEnroll')), color: 'error' })
  }
}

const handleRemove = async (id: string) => {
  if (!await useGlobalModalStore().confirmDelete({ description: t('settings.devices.confirmDelete') })) return

  try {
    await $fetch(`${config.public.apiBase}/devices/${id}`, { method: 'DELETE' })
    toast.add({ title: t('settings.devices.deviceRemoved') })
    await fetchDevices()
  } catch (err: any) {
    toast.add({ title: t('common.error'), description: extractErrorMessage(err, t('settings.devices.errorRemove')), color: 'error' })
  }
}

onMounted(fetchDevices)
</script>
