<template>
  <div class="w-full space-y-4">
    <!-- Table Container using Nuxt UI UTable -->
    <div class="overflow-x-auto rounded-xl border border-[var(--border-main)] bg-[var(--bg-main)] shadow-xs">
      <UTable
        :data="paginatedData"
        :columns="(normalizedColumns as any)"
        :loading="loading"
        :ui="mergedUi"
        class="w-full"
      >
        <!-- Header slots pass-through: support ${col.key}-header, ${col.key}-header-data, or custom slot -->
        <!-- <template
          v-for="col in normalizedColumns"
          :key="`header-${col.id}`"
          #[`${col.id}-header`]="headerProps"
        >
          <slot
            v-if="getHeaderSlotName(col)"
            :name="getHeaderSlotName(col)!"
            v-bind="headerProps"
          />
          <span v-else>{{ col.label }}</span>
        </template> -->

        <!-- Cell slots pass-through: support ${col.key}-cell, ${col.key}-data, and ${col.key} -->
        <!-- <template
          v-for="col in normalizedColumns"
          :key="`cell-${col.id}`"
          #[`${col.id}-cell`]="cellProps"
        >
          <slot
            v-if="getCellSlotName(col)"
            :name="getCellSlotName(col)!"
            v-bind="cellProps"
            :row="getSlotRow(cellProps)"
            :row-data="cellProps.row?.original"
            :value="getSlotValue(cellProps, col)"
            :column="cellProps.column"
            :cell="cellProps.cell"
          />
          <span v-else>{{ getCellValue(cellProps.row?.original, col) }}</span>
        </template> -->

        <!-- Empty state slot -->
        <template #empty>
          <slot name="empty">
            <div class="flex flex-col items-center justify-center py-8 px-4 text-center">
              <UIcon
                :name="emptyState?.icon || 'i-lucide-database'"
                class="w-10 h-10 text-gray-400 dark:text-gray-500 mb-2 opacity-60"
              />
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">
                {{ emptyState?.label || 'No data available' }}
              </p>
              <p
                v-if="emptyState?.description"
                class="text-xs text-gray-400 dark:text-gray-500 mt-1"
              >
                {{ emptyState.description }}
              </p>
            </div>
          </slot>
        </template>

        <!-- Loading state slot -->
        <template #loading>
          <slot name="loading">
            <div class="flex items-center justify-center py-8 px-4">
              <UIcon
                name="i-lucide-loader-2"
                class="w-8 h-8 text-primary-500 animate-spin"
              />
            </div>
          </slot>
        </template>
        <!-- Expanded slot -->
        <!-- <template #expanded="expandedProps">
          <slot
            name="expanded"
            v-bind="expandedProps"
          />
        </template> -->

        <!-- Caption slot -->
        <!-- <template #caption="captionProps">
          <slot
            name="caption"
            v-bind="captionProps"
          />
        </template> -->
        <!-- Footer slots pass-through -->
        <!-- <template
          v-for="col in normalizedColumns"
          :key="`footer-${col.id}`"
          #[`${col.id}-footer`]="footerProps"
        >
          <slot
            v-if="getFooterSlotName(col)"
            :name="getFooterSlotName(col)!"
            v-bind="footerProps"
          />
        </template> -->
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
    data?: Record<string, unknown>[]
    rows?: Record<string, unknown>[]
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
    data: undefined,
    rows: undefined,
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

const rawData = computed(() => props.data ?? props.rows ?? [])

// Helper to access nested object properties e.g. "dept.name"
const getNestedValue = (obj: unknown, path: string) => {
  if (!obj || !path) return undefined
  if (!path.includes('.')) return (obj as Record<string, unknown>)[path]
  return path.split('.').reduce<unknown>((acc, part) => {
    if (acc && typeof acc === 'object') {
      return (acc as Record<string, unknown>)[part]
    }
    return undefined
  }, obj)
}

// Universal slot row helper: ensures entity properties, row.original, and TanStack row methods are seamlessly accessible
const getSlotRow = (cellProps: unknown) => {
  const propsObj = cellProps as Record<string, unknown> | undefined
  const row = propsObj?.row as Record<string, unknown> | undefined
  if (!row) return cellProps
  const original = row.original ?? row

  if (typeof original !== 'object' || original === null) return row

  return new Proxy(row, {
    get(target, prop, receiver) {
      if (prop === 'original') return original
      if (typeof prop === 'string' && prop in original) {
        return (original as Record<string, unknown>)[prop]
      }
      if (prop in target) {
        return Reflect.get(target, prop, receiver)
      }
      return undefined
    }
  })
}

// Get slot value either from TanStack cell context or direct original property
const getSlotValue = (cellProps: unknown, col: TableColumnItem | string) => {
  const propsObj = cellProps as Record<string, unknown> | undefined
  if (typeof propsObj?.getValue === 'function') {
    const val = (propsObj.getValue as () => unknown)()
    if (val !== undefined && val !== null) return val
  }
  const row = propsObj?.row as Record<string, unknown> | undefined
  const original = row?.original ?? row
  const key = typeof col === 'string' ? col : (col.key || col.accessorKey || col.id)
  return key ? getNestedValue(original, key) : undefined
}

// Fallback cell string formatter
const getCellValue = (original: unknown, col: TableColumnItem | string) => {
  if (!original) return '\u00a0'
  const key = typeof col === 'string' ? col : (col.key || col.accessorKey || col.id)
  if (!key) return '\u00a0'
  const val = getNestedValue(original, key)
  if (val === null || val === undefined || val === '') return '\u00a0'
  if (typeof val === 'object') {
    const valObj = val as Record<string, unknown>
    return String(valObj.name || valObj.label || valObj.title || valObj.code || JSON.stringify(valObj))
  }
  return String(val)
}

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
  if (props.serverSide && typeof props.total === 'number') {
    return props.total
  }
  return rawData.value.length
})

const pageCount = computed(() => {
  return Math.ceil(totalItems.value / (perPage.value || 10)) || 1
})

// Auto-reset page if client-side dataset decreases and currentPage exceeds pageCount
watch(
  () => rawData.value.length,
  () => {
    if (!props.serverSide && currentPage.value > pageCount.value) {
      currentPage.value = Math.max(1, pageCount.value)
    }
  }
)

const paginatedData = computed(() => {
  if (props.serverSide) {
    return rawData.value
  }
  const start = (currentPage.value - 1) * perPage.value
  return rawData.value.slice(start, start + perPage.value)
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
