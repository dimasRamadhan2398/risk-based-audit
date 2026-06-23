<template>
  <div class="flex items-center gap-2">
    <svg
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class="shrink-0 h-full w-auto"
    >
      <!-- Globe Background -->
      <circle cx="50" cy="50" r="45" stroke="currentColor" stroke-width="2" class="text-orange-500" />
      <path d="M5 50H95" stroke="currentColor" stroke-width="2" class="text-orange-500" />
      <path d="M50 5V95" stroke="currentColor" stroke-width="2" class="text-orange-500" />
      <path d="M15 25C15 25 30 50 15 75" stroke="currentColor" stroke-width="2" class="text-orange-500" />
      <path d="M85 25C85 25 70 50 85 75" stroke="currentColor" stroke-width="2" class="text-orange-500" />

      <!-- Shield -->
      <path
        d="M50 40L80 50V70C80 85 50 95 50 95C50 95 20 85 20 70V50L50 40Z"
        fill="currentColor"
        class="text-indigo-600"
      />
      <!-- Checkmark -->
      <path
        d="M35 65L45 75L65 55"
        stroke="white"
        stroke-width="5"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
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
