<script setup lang="ts">
import { useI18n } from '~/composables/useI18n'

const props = defineProps<{
  title: string
  subtitle?: string
  icon?: string
  badgeText?: string
  badgeColor?: 'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'error' | 'neutral'
}>()

const { t } = useI18n()
</script>

<template>
  <div class="space-y-4 mb-8">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <!-- Breadcrumbs -->
        <nav class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mb-2">
          <NuxtLink to="/analytics" class="hover:text-primary transition-colors font-medium">
            {{ t('navigation.analytics') }}
          </NuxtLink>
          <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
          <span class="font-medium text-gray-600 dark:text-gray-300">CAATT Insight</span>
          <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
          <span class="text-indigo-600 dark:text-indigo-400 font-bold">{{ props.title }}</span>
        </nav>

        <h1 class="text-2xl md:text-3xl font-black tracking-tight text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon :name="props.icon || 'i-heroicons-table-cells'" class="text-indigo-600 dark:text-indigo-400 w-8 h-8 shrink-0" />
          <span>{{ props.title }}</span>
        </h1>
        <p v-if="props.subtitle" class="text-sm md:text-base text-gray-500 dark:text-gray-400 mt-1">
          {{ props.subtitle }}
        </p>
      </div>

      <div v-if="props.badgeText" class="flex items-center gap-2 shrink-0">
        <UBadge :color="props.badgeColor || 'primary'" variant="subtle" size="lg" class="px-3.5 py-1.5 font-bold shadow-xs">
          {{ props.badgeText }}
        </UBadge>
      </div>
    </div>
  </div>
</template>
