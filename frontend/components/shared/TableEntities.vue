<template>
  <div class="w-full space-y-4">
    <!-- Table Container using Nuxt UI UTable -->
    <div class="overflow-x-auto rounded-xl border border-[var(--border-main)] bg-[var(--bg-main)] shadow-xs">
      <UTable
  :data="paginatedData"
  :columns="normalizedColumns"
  :loading="loading"
  :ui="mergedUi"
>
  <template
    v-for="column in normalizedColumns"
    :key="`${column.id}-cell`"
    #[`${column.id}-cell`]="slotProps"
  >
    <slot
      :name="`${column.id}-cell`"
      v-bind="slotProps"
    >
      {{ slotProps.getValue() }}
    </slot>
  </template>

  <template #empty>
    <slot name="empty">
      <div class="py-8 text-center">
        No data available
      </div>
    </slot>
  </template>

  <template #loading>
    <slot name="loading">
      <div class="flex items-center justify-center py-8">
        <UIcon
          name="i-lucide-loader-2"
          class="size-8 animate-spin"
        />
      </div>
    </slot>
  </template>
</UTable>
    </div>

    <!-- Pagination & Controls Footer using Nuxt UI UPagination -->
    <div
      v-if="showPagination && (totalItems > 0 || !serverSide)"
      class="flex flex-col sm:flex-row items-center justify-between gap-4 px-3 py-2.5"
    >
      <!-- Entry Counter & Page Size Selector -->
      <!-- <div class="flex flex-wrap items-center gap-3 text-xs text-[var(--text-muted)]">
        <span>
          Showing <strong class="font-semibold text-[var(--text-main)]">{{ startItem }}</strong> to
          <strong class="font-semibold text-[var(--text-main)]">{{ endItem }}</strong> of
          <strong class="font-semibold text-[var(--text-main)]">{{ totalItems }}</strong> entries
        </span>
        <div
          v-if="showPageSize"
          class="flex items-center gap-1.5 ml-2"
        >
          <span>Per page:</span>
          <USelect
            v-model="perPage"
            :items="pageSizeSelectOptions"
            value-key="value"
            size="xs"
            class="w-20"
          />
        </div>
      </div> -->

      <!-- UPagination Control -->
     <UPagination
        v-if="totalItems > 0"
        v-model:page="currentPage"
        :items-per-page="perPage"
        :total="totalItems"
        :show-controls="showControls"
        :show-edges="showEdges"
        size="sm"
        active-color="primary"
        color="neutral"
        variant="outline"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, useSlots } from 'vue'
import { getPaginationRowModel } from '@tanstack/vue-table'

defineOptions({ inheritAttrs: false })

export interface TableColumnItem {
  key?: string
  accessorKey?: string
  id?: string
  label?: string
  header?: string
  cell?: unknown
  [key: string]: unknown
}

const props = withDefaults(
  defineProps<{
    data?: any[]
    columns?: (string | TableColumnItem)[]
    loading?: boolean
    emptyState?: { icon?: string, label?: string, description?: string }
    itemsPerPage?: number
    pageSizeOptions?: number[]
    showPagination?: boolean
    showPageSize?: boolean
    showEdges?: boolean
    showControls?: boolean
    serverSide?: boolean
    total?: number
    page?: number
    ui?: Record<string, unknown>
  }>(),
  {
    data: () => [],
    columns: () => [],
    loading: false,
    emptyState: () => ({ icon: 'i-lucide-database', label: 'No data available' }),
    itemsPerPage: 10,
    pageSizeOptions: () => [5, 10, 20, 50, 100],
    showPagination: true,
    showPageSize: true,
    showEdges: true,
    showControls: true,
    serverSide: false,
    total: undefined,
    page: 1,
    ui: () => ({})
  }
)

const emit = defineEmits<{
  (e: 'update:page' | 'update:itemsPerPage', value: number): void
}>()

const $slots = useSlots()

// Normalize columns for TanStack Table / Nuxt UI UTable
const normalizedColumns = computed(() => {
  if (!props.columns || !Array.isArray(props.columns)) return []
  return props.columns.map((col) => {
    if (typeof col === 'string') {
      return {
        key: col,
        accessorKey: col,
        id: col,
        label: col,
        header: col
      }
    }
    const key = col.key || col.accessorKey || col.id || ''
    const label = col.label || col.header || key
    const normalized: TableColumnItem = {
      ...col,
      key,
      accessorKey: col.accessorKey || key,
      id: col.id || key,
      label,
      header: col.header || label
    }
    if (typeof normalized.cell === 'string') {
      delete normalized.cell
    }
    return normalized
  })
})

// Dynamic slot resolution helpers
const getCellSlotName = (col: TableColumnItem) => {
  const keys = Array.from(new Set([col.key, col.accessorKey, col.id].filter(Boolean))) as string[]
  for (const k of keys) {
    if ($slots[`${k}-cell`]) return `${k}-cell`
    if ($slots[`${k}-data`]) return `${k}-data`
    if ($slots[k]) return k
  }
  return null
}

const getHeaderSlotName = (col: TableColumnItem) => {
  const keys = Array.from(new Set([col.key, col.accessorKey, col.id].filter(Boolean))) as string[]
  for (const k of keys) {
    if ($slots[`${k}-header`]) return `${k}-header`
    if ($slots[`${k}-header-data`]) return `${k}-header-data`
  }
  return null
}

const getFooterSlotName = (col: TableColumnItem) => {
  const keys = Array.from(new Set([col.key, col.accessorKey, col.id].filter(Boolean))) as string[]
  for (const k of keys) {
    if ($slots[`${k}-footer`]) return `${k}-footer`
  }
  return null
}

const currentPage = ref(props.page || 1)
const perPage = ref(props.itemsPerPage || 10)

watch(
  () => props.page,
  (newPage) => {
    if (newPage !== undefined && newPage !== currentPage.value) {
      currentPage.value = newPage
    }
  }
)

watch(
  () => props.itemsPerPage,
  (newSize) => {
    if (newSize !== undefined && newSize !== perPage.value) {
      perPage.value = newSize
    }
  }
)

watch(currentPage, (newVal) => {
  emit('update:page', newVal)
})

watch(perPage, (newVal) => {
  if (currentPage.value !== 1) {
    currentPage.value = 1
  }
  emit('update:itemsPerPage', newVal)
})

const totalItems = computed(() => {
  if (props.serverSide && props.total !== undefined) {
    return props.total
  }
  return props.data?.length || 0
})

const paginatedData = computed(() => {
  if (!props.data) return []
  if (props.serverSide) return props.data
  const start = (currentPage.value - 1) * perPage.value
  return props.data.slice(start, start + perPage.value)
})

const startItem = computed(() => {
  if (totalItems.value === 0) return 0
  return (currentPage.value - 1) * perPage.value + 1
})

const endItem = computed(() => {
  if (totalItems.value === 0) return 0
  return Math.min(currentPage.value * perPage.value, totalItems.value)
})

const pageSizeSelectOptions = computed(() => {
  return (props.pageSizeOptions || [5, 10, 20, 50, 100]).map(size => ({
    label: `${size}`,
    value: Number(size)
  }))
})

const mergedUi = computed(() => ({
  th: 'px-4 py-3.5 text-xs text-left font-bold uppercase tracking-wider text-[var(--text-muted)] bg-[var(--bg-surface)] border-b border-[var(--border-main)]',
  td: 'px-4 py-3 text-sm text-[var(--text-main)] border-b border-[var(--border-main)]/50',
  ...props.ui
}))
</script>
