<template>
  <div class="w-full space-y-4">
    <!-- Table Container using Nuxt UI UTable -->
    <div
      class="rounded-xl border border-[var(--border-main)] bg-[var(--bg-main)] shadow-xs w-full transition-all"
      :class="[
        isScrollable ? 'overflow-x-auto' : 'overflow-x-hidden'
      ]"
    >
      <UTable
        :data="paginatedData"
        :columns="(normalizedColumns as any)"
        :loading="loading"
        :ui="mergedUi"
        :class="[
          'w-full',
          isScrollable
            ? (effectiveTableLayout === 'fixed' ? 'table-fixed' : 'table-auto')
            : 'table-fixed min-w-full',
          effectiveMinWidth ? '' : (isScrollable && effectiveTableLayout === 'auto' ? 'min-w-max' : 'min-w-full'),
          tableClass
        ]"
        :style="effectiveMinWidth ? { minWidth: effectiveMinWidth } : undefined"
      >
        <template
          v-for="col in normalizedColumns"
          #[`${col.id}-cell`]="props"
        >
          <slot
            v-if="$slots[`${col.id}-cell`]"
            :name="`${col.id}-cell`"
            v-bind="props"
          />

          <TeamMembersBadge
            v-else-if="col.id === 'teamMembers' || col.accessorKey === 'teamMembers'"
            :members="props.row?.original?.teamMembers || props.getValue()"
          />

          <span v-else>
            {{ props.getValue() }}
          </span>
        </template>

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
      class="flex flex-col sm:flex-row items-center justify-end gap-4 px-3 py-2.5"
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
    horizontalScroll?: boolean
    scrollable?: boolean
    allowHorizontalScroll?: boolean
    tableLayout?: 'fixed' | 'auto'
    minWidth?: string
    tableClass?: string
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
    ui: () => ({}),
    horizontalScroll: undefined,
    scrollable: undefined,
    allowHorizontalScroll: undefined,
    tableLayout: undefined,
    minWidth: undefined,
    tableClass: ''
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
    
    const colClass = (col as any).class || ''
    const colThClass = (col as any).thClass || ''
    const colTdClass = (col as any).tdClass || ''
    const existingMeta = (col as any).meta || {}
    const existingMetaClass = existingMeta.class || {}

    const thClass = [
      colClass,
      colThClass,
      typeof existingMetaClass === 'string' ? existingMetaClass : existingMetaClass.th
    ].filter(Boolean).join(' ')

    const tdClass = [
      colClass,
      colTdClass,
      typeof existingMetaClass === 'string' ? existingMetaClass : existingMetaClass.td
    ].filter(Boolean).join(' ')

    const normalized: TableColumnItem = {
      ...col,
      key,
      accessorKey: col.accessorKey || key,
      id: col.id || key,
      label,
      header: col.header || label,
      meta: {
        ...existingMeta,
        class: {
          th: thClass,
          td: tdClass
        }
      }
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

const isScrollable = computed(() => {
  if (props.horizontalScroll !== undefined) return props.horizontalScroll
  if (props.scrollable !== undefined) return props.scrollable
  if (props.allowHorizontalScroll !== undefined) return props.allowHorizontalScroll
  return true
})

const effectiveTableLayout = computed(() => {
  if (props.tableLayout) return props.tableLayout
  return isScrollable.value ? 'auto' : 'fixed'
})

const effectiveMinWidth = computed(() => props.minWidth)

const mergedUi = computed(() => {
  const layoutClass = effectiveTableLayout.value === 'auto' ? 'table-auto' : 'table-fixed'
  const widthClass = isScrollable.value
    ? (props.minWidth ? '' : 'min-w-full')
    : 'min-w-full w-full'

  return {
    base: `w-full ${layoutClass} ${widthClass}`.trim(),
    table: `w-full ${layoutClass} ${widthClass}`.trim(),
    th: 'px-4 py-3.5 text-xs text-left font-bold uppercase tracking-wider text-[var(--text-muted)] bg-[var(--bg-surface)] border-b border-[var(--border-main)] whitespace-nowrap',
    td: 'px-4 py-3 text-sm text-[var(--text-main)] border-b border-[var(--border-main)]/50',
    ...props.ui
  }
})
</script>
