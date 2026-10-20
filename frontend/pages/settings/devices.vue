<template>
  <UDashboardPage>
    <UDashboardHeader :title="t('settings.devices.title')" />

    <UDashboardPanelContent>
      <div class="space-y-6">
        <UCard>
          <template #header>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h3 class="text-lg font-semibold">{{ t('settings.devices.registeredDevices') }}</h3>
              <UButton icon="i-lucide-qr-code" class="w-full sm:w-auto justify-center" @click="handleGenerateQR">{{ t('settings.devices.registerNew') }}</UButton>
            </div>
          </template>

          <UTable :rows="devices" :columns="columns">
            <template #created_at-data="{ row }">
              {{ formatDate(row.original.created_at) }}
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
            <template #header>{{ t('settings.devices.registerTitle') }}</template>
            <div class="flex flex-col items-center py-6">
              <p class="mb-4 text-center">{{ t('settings.devices.scanInstruction') }}</p>
              <div class="bg-white p-4 rounded-lg shadow-inner">
                <img v-if="qrCodeDataURL" :src="qrCodeDataURL" :alt="t('settings.devices.qrAlt')" />
              </div>
              <p v-if="qrData" class="mt-4 text-sm text-gray-500 italic">{{ t('settings.devices.expiresAt', { date: formatDateTime(qrData.expires_at) }) }}</p>
            </div>
            <template #footer>
              <UButton block @click="() => { showQRModal = false }">{{ t('settings.devices.close') }}</UButton>
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
const devices = ref<Device[]>([])
const showQRModal = ref(false)
const qrData = ref<any>(null)
const qrCodeDataURL = ref('')
const toast = useToast()
const { t, locale } = useI18n()

const columns = computed<TableColumn<Device>[]>(() => [
  { accessorKey: 'device_name', header: t('settings.devices.columns.device') },
  { accessorKey: 'ip_address', header: t('settings.devices.columns.ipAddress') },
  { accessorKey: 'created_at', header: t('settings.devices.columns.addedOn') },
  { accessorKey: 'actions', header: '' }
])

const dateLocale = computed(() => locale.value === 'id' ? 'id-ID' : 'en-US')
const formatDate = (value: string) => new Date(value).toLocaleDateString(dateLocale.value)
const formatDateTime = (value: string) => new Date(value).toLocaleString(dateLocale.value)

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
    toast.add({ title: t('common.error'), description: extractErrorMessage(err, t('settings.devices.enrollFailed')), color: 'error' })
  }
}

const handleRemove = async (id: string) => {
  if (!await useGlobalModalStore().confirmDelete({ description: t('settings.devices.removeConfirm') })) return

  try {
    await $fetch(`${config.public.apiBase}/devices/${id}`, { method: 'DELETE' })
    toast.add({ title: t('settings.devices.removed') })
    await fetchDevices()
  } catch (err: any) {
    toast.add({ title: t('common.error'), description: extractErrorMessage(err, t('settings.devices.removeFailed')), color: 'error' })
  }
}

onMounted(fetchDevices)
</script>
