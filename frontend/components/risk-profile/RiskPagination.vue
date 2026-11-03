<template>
  <div
    v-if="total > 0"
    class="flex flex-col sm:flex-row items-center justify-center sm:justify-between gap-2 px-2 sm:px-3 pt-3 overflow-x-auto max-w-full"
  >
    <span class="text-xs text-gray-500 dark:text-gray-400">
      {{ t('riskProfile.pagination.showing', { from, to, total }) }}
    </span>
    <UPagination
      :page="page"
      :items-per-page="pageSize"
      :total="total"
      size="sm"
      active-color="primary"
      color="neutral"
      variant="outline"
      class="max-w-full overflow-x-auto"
      @update:page="(p: number) => emit('update:page', p)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'

const props = defineProps<{ page: number, pageSize: number, total: number }>()
const emit = defineEmits<{ (e: 'update:page', value: number): void }>()
const { t } = useI18n()

const from = computed(() => (props.page - 1) * props.pageSize + 1)
const to = computed(() => Math.min(props.page * props.pageSize, props.total))
</script>
