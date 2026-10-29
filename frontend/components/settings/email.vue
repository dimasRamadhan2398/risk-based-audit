<template>
  <div class="space-y-6 max-w-5xl">
    <!-- Header Status Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-bold text-lg shrink-0">
              <UIcon name="i-lucide-mail" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.email.cardTitle') }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('settings.email.cardSubtitle') }}
              </p>
            </div>
          </div>
          <UBadge color="success" variant="subtle" class="self-start sm:self-auto font-medium">
            <span class="w-2 h-2 rounded-full bg-emerald-500 mr-1.5 animate-pulse" />
            {{ t('settings.email.badgeConnected') }}
          </UBadge>
        </div>
      </template>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.activeDispatcher') }}</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon name="i-lucide-cloud" class="w-4 h-4 text-primary" />
            {{ t('settings.email.cloudEngine') }}
          </p>
          <span class="text-[11px] text-gray-500">{{ t('settings.email.replacesLegacy') }}</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.backendEndpoint') }}</span>
          <p class="text-xs font-mono font-bold text-gray-800 dark:text-gray-200 truncate">
            /api/v1/resend/provision
          </p>
          <span class="text-[11px] text-emerald-600 dark:text-emerald-400">{{ t('settings.email.kongProxyActive') }}</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.deliveryMode') }}</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon name="i-lucide-shield-check" class="w-4 h-4 text-emerald-500" />
            {{ t('settings.email.dkimSpfEnforced') }}
          </p>
          <span class="text-[11px] text-gray-500">{{ t('settings.email.highDeliverability') }}</span>
        </div>
      </div>
    </UCard>

    <!-- Provision Domain Form Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-globe" class="w-4 h-4 text-primary" />
            {{ t('settings.email.provisionTitle') }}
          </h3>
          <span class="text-xs text-gray-500">{{ t('settings.email.autoDnsGen') }}</span>
        </div>
      </template>

      <form @submit.prevent="handleProvision" class="space-y-4">
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <UFormField :label="t('settings.email.clientName')" required :hint="t('settings.email.clientNameHint')">
            <UInput
              v-model="form.clientName"
              :placeholder="t('settings.email.clientNamePlaceholder')"
              required
              class="w-full"
            />
          </UFormField>

          <UFormField :label="t('settings.email.tenantSlug')" required :hint="t('settings.email.tenantSlugHint')">
            <UInput
              v-model="form.slug"
              :placeholder="t('settings.email.tenantSlugPlaceholder')"
              required
              class="w-full font-mono"
            />
          </UFormField>

          <UFormField :label="t('settings.email.sendingDomain')" required :hint="t('settings.email.sendingDomainHint')">
            <UInput
              v-model="form.domain"
              :placeholder="t('settings.email.sendingDomainPlaceholder')"
              required
              class="w-full font-mono"
            />
          </UFormField>
        </div>

        <div class="flex items-center justify-between pt-2">
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('settings.email.provisionDesc') }}
          </p>
          <UButton
            type="submit"
            color="primary"
            class="rounded-xl font-bold"
            :loading="isProvisioning"
            icon="i-lucide-sparkles"
          >
            {{ t('settings.email.provisionButton') }}
          </UButton>
        </div>
      </form>
    </UCard>

    <!-- Provisioning Result & DNS Records -->
    <UCard v-if="provisionResult" class="border border-emerald-500/30 dark:border-emerald-500/20 bg-emerald-50/20 dark:bg-emerald-950/10 rounded-2xl shadow-md">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-check-circle" class="w-5 h-5 text-emerald-500" />
            <h3 class="text-base font-bold text-gray-900 dark:text-white">
              {{ t('settings.email.domainProvisioned', { domain: provisionResult.domain }) }}
            </h3>
          </div>
          <UBadge color="primary" variant="subtle" class="font-mono text-xs">
            {{ t('settings.email.domainId', { id: provisionResult.domain_id }) }}
          </UBadge>
        </div>
      </template>

      <div class="space-y-4">
        <!-- Scoped Client API Key Display -->
        <div class="p-3.5 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 flex items-center justify-between gap-3 text-xs">
          <div class="space-y-0.5 min-w-0 flex-1">
            <span class="text-gray-400 font-medium block">{{ t('settings.email.scopedApiKey') }}</span>
            <div class="font-mono font-bold text-primary truncate">
              {{ showApiKey ? provisionResult.client_api_key : maskKey(provisionResult.client_api_key) }}
            </div>
          </div>
          <div class="flex items-center gap-1.5 shrink-0">
            <UButton
              color="neutral"
              variant="ghost"
              size="xs"
              :icon="showApiKey ? 'i-lucide-eye-off' : 'i-lucide-eye'"
              @click="showApiKey = !showApiKey"
            />
            <UButton
              color="neutral"
              variant="outline"
              size="xs"
              icon="i-lucide-copy"
              @click="copyText(provisionResult.client_api_key, 'Client API Key')"
            >
              {{ t('settings.email.copy') }}
            </UButton>
          </div>
        </div>

        <!-- DNS Records Table -->
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs font-semibold text-gray-700 dark:text-gray-300">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-lucide-list" class="w-4 h-4 text-primary" />
              {{ t('settings.email.requiredDns') }}
            </span>
            <span class="text-gray-400 font-normal">{{ t('settings.email.addDnsHint') }}</span>
          </div>

          <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <table class="w-full text-left text-xs font-mono">
              <thead class="bg-gray-50 dark:bg-gray-800/60 text-gray-500 border-b border-gray-200 dark:border-gray-800">
                <tr>
                  <th class="p-3">{{ t('settings.email.thType') }}</th>
                  <th class="p-3">{{ t('settings.email.thHostName') }}</th>
                  <th class="p-3">{{ t('settings.email.thValue') }}</th>
                  <th class="p-3">{{ t('settings.email.thPriority') }}</th>
                  <th class="p-3 text-right">{{ t('settings.email.thAction') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr
                  v-for="(rec, idx) in provisionResult.dns_records"
                  :key="idx"
                  class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30"
                >
                  <td class="p-3">
                    <UBadge color="primary" variant="subtle" size="xs">{{ rec.type }}</UBadge>
                  </td>
                  <td class="p-3 max-w-[200px] truncate text-gray-800 dark:text-gray-200 font-medium" :title="rec.name">
                    {{ rec.name }}
                  </td>
                  <td class="p-3 max-w-[320px] truncate text-gray-600 dark:text-gray-400" :title="rec.value">
                    {{ rec.value }}
                  </td>
                  <td class="p-3 text-gray-500">
                    {{ rec.priority !== undefined && rec.priority !== null ? rec.priority : '-' }}
                  </td>
                  <td class="p-3 text-right">
                    <UButton
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      icon="i-lucide-copy"
                      @click="copyText(rec.value, `${rec.type} Record Value`)"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const { t } = useI18n()
const toast = useToast()

interface DNSRecord {
  name: string
  type: string
  value: string
  priority?: number
  status?: string
}

interface ProvisionResponse {
  domain_id: string
  domain: string
  status: string
  dns_records: DNSRecord[]
  client_api_key: string
}

const form = reactive({
  clientName: 'AuditSphere Client',
  slug: 'client',
  domain: 'client.auditsphere.id'
})

const isProvisioning = ref(false)
const showApiKey = ref(false)
const provisionResult = ref<ProvisionResponse | null>(null)

const maskKey = (key: string) => {
  if (!key || key.length < 10) return '••••••••••••••••'
  return key.slice(0, 7) + '••••••••••••••••' + key.slice(-4)
}

const copyText = async (text: string, label: string) => {
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    }
    toast.add({
      title: t('settings.email.copiedToast'),
      description: t('settings.email.copiedDesc', { name: label }),
      color: 'success'
    })
  } catch {
    toast.add({
      title: t('settings.email.copiedToast'),
      description: text,
      color: 'success'
    })
  }
}

const handleProvision = async () => {
  isProvisioning.value = true
  try {
    const res = await $fetch<{ success: boolean; data: ProvisionResponse }>('/api/v1/resend/provision', {
      method: 'POST',
      body: {
        client_name: form.clientName,
        slug: form.slug,
        domain: form.domain
      }
    })

    if (res?.data) {
      provisionResult.value = res.data
      toast.add({
        title: t('settings.email.provisionSuccessToast'),
        description: t('settings.email.provisionSuccessDesc', { domain: res.data.domain, count: res.data.dns_records?.length || 0 }),
        color: 'success'
      })
    }
  } catch (error: any) {
    toast.add({
      title: t('settings.email.provisionFailedToast'),
      description: error?.data?.message || error?.message || t('settings.email.provisionFailedDesc'),
      color: 'error'
    })
  } finally {
    isProvisioning.value = false
  }
}
</script>
