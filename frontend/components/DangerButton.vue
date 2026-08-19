<template>
  <DangerButton v-bind="$attrs">
    <template v-for="(_, slot) in $slots" #[slot]="scope">
      <slot :name="slot" v-bind="scope || {}" />
    </template>
  </DangerButton>
</template>

<script setup lang="ts">
import DangerButton from './shared/ReusableButton.vue'
</script>
