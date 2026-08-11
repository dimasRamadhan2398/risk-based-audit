<template>
  <UDashboardPage>
    <UDashboardHeader :title="t('settings.titles.mfa')" />

    <UDashboardPanelContent>
      <SettingsMfa />
    </UDashboardPanelContent>
  </UDashboardPage>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
  layout: 'dashboard'
})

const { t } = useI18n()
</script>
