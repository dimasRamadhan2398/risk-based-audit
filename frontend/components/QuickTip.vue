<template>
  <QuickTip v-bind="$attrs">
    <template
      v-for="(_, slot) in $slots"
      #[slot]="scope"
    >
      <slot
        :name="slot"
        v-bind="scope || {}"
      />
    </template>
  </QuickTip>
</template>

<script setup lang="ts">
import QuickTip from './shared/QuickTip.vue'
</script>
