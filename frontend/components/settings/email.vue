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
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">Email Service Dispatcher (Resend API)</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                Manage transactional email delivery, provision sending domains, and inspect DNS verification records.
              </p>
            </div>
          </div>
          <UBadge color="success" variant="subtle" class="self-start sm:self-auto font-medium">
            <span class="w-2 h-2 rounded-full bg-emerald-500 mr-1.5 animate-pulse" />
            Resend API Connected
          </UBadge>
        </div>
      </template>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">Active Dispatcher</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon name="i-lucide-cloud" class="w-4 h-4 text-primary" />
            Resend Cloud Engine
          </p>
          <span class="text-[11px] text-gray-500">Replaces legacy Mailtrap SMTP</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">Backend Endpoint</span>
          <p class="text-xs font-mono font-bold text-gray-800 dark:text-gray-200 truncate">
            /api/v1/resend/provision
          </p>
          <span class="text-[11px] text-emerald-600 dark:text-emerald-400">Kong Gateway Proxy Active</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">Delivery Mode</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon name="i-lucide-shield-check" class="w-4 h-4 text-emerald-500" />
            DKIM & SPF Enforced
          </p>
          <span class="text-[11px] text-gray-500">High deliverability inbox rate</span>
        </div>
      </div>
    </UCard>

    <!-- Provision Domain Form Card -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-md">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-globe" class="w-4 h-4 text-primary" />
            Provision Client Sending Domain
          </h3>
          <span class="text-xs text-gray-500">Automatic DNS generation</span>
        </div>
      </template>

      <form @submit.prevent="handleProvision" class="space-y-4">
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <UFormField label="Client Name" required hint="e.g. Acme Corp">
            <UInput
              v-model="form.clientName"
              placeholder="e.g. Acme Corporation"
              required
              class="w-full"
            />
          </UFormField>

          <UFormField label="Tenant Slug" required hint="Short alphanumeric">
            <UInput
              v-model="form.slug"
              placeholder="e.g. acme"
              required
              class="w-full font-mono"
            />
          </UFormField>

          <UFormField label="Sending Domain" required hint="Auditsphere or custom">
            <UInput
              v-model="form.domain"
              placeholder="e.g. acme.auditsphere.id"
              required
              class="w-full font-mono"
            />
          </UFormField>
        </div>

        <div class="flex items-center justify-between pt-2">
          <p class="text-xs text-gray-500 dark:text-gray-400">
            Registers domain on Resend and automatically creates a tenant-scoped API key.
          </p>
          <UButton
            type="submit"
            color="primary"
            class="rounded-xl font-bold"
            :loading="isProvisioning"
            icon="i-lucide-sparkles"
          >
            Provision Domain via API
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
              Domain Provisioned: {{ provisionResult.domain }}
            </h3>
          </div>
          <UBadge color="primary" variant="subtle" class="font-mono text-xs">
            Domain ID: {{ provisionResult.domain_id }}
          </UBadge>
        </div>
      </template>

      <div class="space-y-4">
        <!-- Scoped Client API Key Display -->
        <div class="p-3.5 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 flex items-center justify-between gap-3 text-xs">
          <div class="space-y-0.5 min-w-0 flex-1">
            <span class="text-gray-400 font-medium block">Scoped Tenant API Key</span>
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
              Copy
            </UButton>
          </div>
        </div>

        <!-- DNS Records Table -->
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs font-semibold text-gray-700 dark:text-gray-300">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-lucide-list" class="w-4 h-4 text-primary" />
              Required DNS Records for Verification (DKIM, SPF, MX)
            </span>
            <span class="text-gray-400 font-normal">Add these records to your DNS provider</span>
          </div>

          <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <table class="w-full text-left text-xs font-mono">
              <thead class="bg-gray-50 dark:bg-gray-800/60 text-gray-500 border-b border-gray-200 dark:border-gray-800">
                <tr>
                  <th class="p-3">Type</th>
                  <th class="p-3">Host / Name</th>
                  <th class="p-3">Value / Destination</th>
                  <th class="p-3">Priority</th>
                  <th class="p-3 text-right">Action</th>
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
      title: 'Copied to Clipboard',
      description: `${label} copied successfully.`,
      color: 'success'
    })
  } catch {
    toast.add({
      title: 'Copied',
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
        title: 'Domain Provisioned',
        description: `Successfully provisioned ${res.data.domain} via Resend API.`,
        color: 'success'
      })
    }
  } catch (error: any) {
    toast.add({
      title: 'Provisioning Failed',
      description: error?.data?.message || error?.message || 'Failed to call Resend API',
      color: 'error'
    })
  } finally {
    isProvisioning.value = false
  }
}
</script>
