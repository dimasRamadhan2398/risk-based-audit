<template>
  <div class="flex items-center gap-2 min-w-0 select-none">
    <img
      :src="auditsphereIconUrl"
      alt="AuditSphere Logo"
      class="shrink-0 w-auto"
      :class="iconOnly ? 'h-7 sm:h-8' : 'h-8 sm:h-9'"
    />
    <div v-if="!iconOnly" class="flex flex-col justify-center min-w-0">
      <div class="font-bold flex leading-none" :class="textClass">
        <span class="text-orange-600">Audit</span>
        <span class="text-indigo-700 dark:text-indigo-400">Sphere</span>
      </div>
      <span v-if="!hideSubtitle" class="text-[8px] sm:text-[9px] text-gray-500 dark:text-gray-400 uppercase tracking-tight font-semibold whitespace-nowrap leading-tight mt-0.5">Intelligent Risk-Based Audit</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import auditsphereIconUrl from '@/assets/images/utils/auditsphere-icon.png'

interface Props {
  iconOnly?: boolean
  hideSubtitle?: boolean
  textClass?: string
}

withDefaults(defineProps<Props>(), {
  iconOnly: false,
  hideSubtitle: false,
  textClass: 'text-lg'
})
</script>
