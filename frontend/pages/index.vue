<script setup lang="ts">
definePageMeta({
  middleware: [
    'guest',
    () => navigateTo('/auth/login')
  ]
})
</script>

<template>
  <div />
</template>
