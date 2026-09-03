<template>
  <AppFormField v-bind="$attrs">
    <template
      v-for="(_, slotName) in $slots"
      #[slotName]="slotProps"
    >
      <slot
        :name="slotName"
        v-bind="slotProps || {}"
      />
    </template>
  </AppFormField>
</template>

<script setup lang="ts">
import AppFormField from './AppFormField.vue'

defineOptions({
  inheritAttrs: false
})
</script>
