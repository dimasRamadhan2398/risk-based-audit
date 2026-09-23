<template>
  <UApp :toaster="toaster">
    <NuxtLayout>
      <NuxtPage />
      <GlobalConfirmationModal />
    </NuxtLayout>
  </UApp>
</template>

<script setup lang="ts">

const toaster = {
  max: 1
}
// Auth is initialized via plugins and middleware
// No need for manual initialization here
</script>
