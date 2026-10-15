<template>
  <UTooltip :text="text" :disabled="!isOverflowing" :content="{ side: 'top' }">
    <span ref="el" :class="textClass" @mouseenter="checkOverflow">{{ text }}</span>
  </UTooltip>
</template>

<script setup lang="ts">
import { ref } from 'vue'

// Shows the full text in a tooltip only when it is clipped (truncate / line-clamp).
defineProps<{
  text: string
  textClass?: string
}>()

const el = ref<HTMLElement | null>(null)
const isOverflowing = ref(false)

const checkOverflow = () => {
  const node = el.value
  if (!node) return
  isOverflowing.value
    = node.scrollWidth > node.clientWidth || node.scrollHeight > node.clientHeight
}
</script>
