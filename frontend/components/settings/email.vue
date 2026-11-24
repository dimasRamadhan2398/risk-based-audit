<template>
  <div class="space-y-6 max-w-5xl">
    <!-- Header Status Card -->
    <UCard class="border border-gray-200 dark:border-slate-800 dark:bg-slate-900 dark:ring-slate-800 rounded-2xl shadow-md">
      <template #header>
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-bold text-lg shrink-0">
            <UIcon
              name="i-lucide-mail"
              class="w-5 h-5"
            />
          </div>
          <div>
            <h3 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ t('settings.email.cardTitle') }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('settings.email.cardSubtitle') }}
            </p>
          </div>
        </div>
      </template>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.activeDispatcher') }}</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon
              name="i-lucide-cloud"
              class="w-4 h-4 text-primary"
            />
            {{ t('settings.email.cloudEngine') }}
          </p>
          <span class="text-[11px] text-gray-500">{{ t('settings.email.replacesLegacy') }}</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.tenantProvisioning') }}</span>
          <p class="text-xs font-mono font-bold text-gray-800 dark:text-gray-200 truncate">
            scripts/provision-tenant-email.sh
          </p>
          <span class="text-[11px] text-gray-500">{{ t('settings.email.tenantProvisioningHint') }}</span>
        </div>

        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700/60 space-y-1">
          <span class="text-gray-400 font-medium">{{ t('settings.email.deliveryMode') }}</span>
          <p class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
            <UIcon
              name="i-lucide-shield-check"
              class="w-4 h-4 text-emerald-500"
            />
            {{ t('settings.email.dkimSpfEnforced') }}
          </p>
          <span class="text-[11px] text-gray-500">{{ t('settings.email.highDeliverability') }}</span>
        </div>
      </div>
    </UCard>

    <!-- Tenant sending keys: created by the provisioning job, not from this screen -->
    <UCard class="border border-gray-200 dark:border-slate-800 dark:bg-slate-900 dark:ring-slate-800 rounded-2xl shadow-md">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="flex items-start gap-3">
          <UIcon
            name="i-lucide-info"
            class="w-5 h-5 text-primary shrink-0 mt-0.5"
          />
          <div class="space-y-1 text-sm">
            <p class="font-semibold text-gray-900 dark:text-white">
              {{ t('settings.email.provisionMovedTitle') }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('settings.email.provisionMovedDesc') }}
            </p>
          </div>
        </div>
        <UButton
          v-if="isPlatformAdmin"
          to="/site-generator"
          color="primary"
          variant="soft"
          icon="i-lucide-globe"
          class="rounded-xl font-bold shrink-0"
        >
          {{ t('settings.email.openSiteGenerator') }}
        </UButton>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
// Tenant sending domains/keys are provisioned by the Site Generator job (step "email") or by
// scripts/provision-tenant-email.sh on the VPS. The old POST /api/v1/resend/provision API is gone.
const { t } = useI18n()
const { isPlatformAdmin } = useRbac()
</script>
