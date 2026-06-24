<script setup lang="ts">
definePageMeta({
  middleware: 'guest'  // Prevent already-logged-in users from seeing redirect
})

// Redirect to login
navigateTo('/auth/login')
</script>

<template>
  <div></div>
</template>
