<template>
  <div class="flex items-center gap-2">
    <img
      :src="auditsphereIconUrl"
      alt="AuditSphere Logo"
      class="shrink-0 h-full w-auto"
      :class="iconOnly ? 'max-h-8' : 'max-h-12'"
    />
    <div v-if="!iconOnly" class="flex flex-col leading-none">
      <div class="font-bold flex" :class="textClass">
        <span class="text-orange-600">Audit</span>
        <span class="text-indigo-700">Sphere</span>
      </div>
      <span v-if="!hideSubtitle" class="text-[8px] text-gray-500 uppercase tracking-tighter font-semibold">Intelligent Risk-Based Audit</span>
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
