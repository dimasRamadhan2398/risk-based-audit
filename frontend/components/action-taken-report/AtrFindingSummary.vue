<template>
  <!-- Finding the ATR follows up (copied from the LHA when it was approved) -->
  <section class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4 bg-white dark:bg-gray-900">
    <div class="flex flex-wrap items-start justify-between gap-2">
      <h4 class="text-base font-bold">
        {{ t('actionTakenReport.detail.findingSection') }}
      </h4>
      <UBadge
        v-if="item.finding_category"
        :color="atrFindingCategoryColor(item.finding_category)"
        variant="subtle"
        size="sm"
        :label="categoryLabel"
      />
    </div>

    <dl class="space-y-3 text-sm">
      <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
        <dt class="w-full sm:w-1/3 font-semibold text-gray-600 dark:text-gray-300">
          {{ t('actionTakenReport.detail.finding') }}
        </dt>
        <dd class="w-full sm:w-2/3 whitespace-pre-line break-words">
          {{ item.finding_title || '-' }}
        </dd>
      </div>
      <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
        <dt class="w-full sm:w-1/3 font-semibold text-gray-600 dark:text-gray-300">
          {{ t('actionTakenReport.detail.recommendation') }}
        </dt>
        <dd class="w-full sm:w-2/3 whitespace-pre-line break-words">
          {{ item.recommendation || '-' }}
        </dd>
      </div>
      <div class="flex flex-col sm:flex-row gap-1 sm:gap-4">
        <dt class="w-full sm:w-1/3 font-semibold text-gray-600 dark:text-gray-300">
          {{ t('actionTakenReport.detail.lha') }}
        </dt>
        <dd class="w-full sm:w-2/3 space-y-1">
          <p class="font-mono font-semibold text-primary-600 dark:text-primary-400">
            {{ item.report_number || '-' }}
          </p>
          <p
            v-if="item.report_title"
            class="text-gray-600 dark:text-gray-300 break-words"
          >
            {{ item.report_title }}
          </p>
          <slot name="lha-actions" />
        </dd>
      </div>
    </dl>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ActionTakenReport } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'
import { atrFindingCategoryColor, atrFindingCategoryI18nKey } from '~/utils/actionTakenReport'

const props = defineProps<{ item: ActionTakenReport }>()
const { t } = useI18n()

const categoryLabel = computed(() => {
  const key = atrFindingCategoryI18nKey(props.item.finding_category)
  return key ? t(key) : props.item.finding_category
})
</script>
