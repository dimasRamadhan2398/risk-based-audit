<template>
  <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
    <template #header>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">{{ t('masterData.location.listTitle') }}</h2>
        <UButton
          :label="t('masterData.location.add')"
          icon="i-heroicons-plus"
          color="primary"
          class="w-full sm:w-auto justify-center text-center"
          @click="store.openCreateModal"
        />
      </div>
    </template>

    <!-- Search & Filter -->
    <div class="mb-4 flex flex-wrap gap-2 sm:gap-4 items-center">
      <UInput
        v-model="searchInput"
        :placeholder="t('masterData.location.searchPlaceholder')"
        icon="i-heroicons-magnifying-glass"
        class="w-full sm:w-64"
        @keyup.enter="handleSearch"
      >
        <template v-if="searchInput" #trailing>
          <UButton
            color="neutral"
            variant="ghost"
            size="xs"
            icon="i-heroicons-x-mark"
            aria-label="Clear search"
            @click="resetSearch"
          />
        </template>
      </UInput>
    </div>

    <!-- Loading State -->
    <div v-if="store.loading" class="flex justify-center py-8">
      <ULoadingIcon />
    </div>

    <!-- Error State -->
    <div v-else-if="store.errorMsg" class="py-4">
      <UAlert
        color="error"
        variant="soft"
        :title="store.errorMsg"
        icon="i-heroicons-exclamation-circle"
      />
    </div>

    <!-- Empty State -->
    <div v-else-if="store.filteredLocations.length === 0" class="py-8 text-center">
      <UIcon name="i-heroicons-map-pin" class="text-4xl text-gray-400 mb-2" />
      <p class="text-gray-500">{{ t('masterData.location.empty') }}</p>
    </div>

    <!-- Table & Pagination via TableEntities (client-side: the API returns a flat list) -->
    <div v-else class="space-y-4">
      <!-- Desktop View (>= md): Full Table Entities -->
      <div class="hidden md:block">
        <TableEntities
          v-model:page="locationPage"
          :items-per-page="itemsPerPage"
          :data="store.filteredLocations"
          :columns="store.columns"
          :loading="store.loading"
          :show-pagination="true"
        >
          <template #name-cell="{ row }">
            <div class="font-medium text-primary-600">{{ row.original.name }}</div>
          </template>

          <template #city-cell="{ row }">
            <span class="font-medium">{{ row.original.city || '-' }}</span>
          </template>

          <template #province-cell="{ row }">
            <span class="text-gray-600 text-sm">{{ row.original.province || '-' }}</span>
          </template>

          <template #address-cell="{ row }">
            <span class="text-gray-600 text-sm">{{ row.original.address || '-' }}</span>
          </template>

          <template #is_active-cell="{ row }">
            <UBadge
              :color="row.original.is_active ? 'success' : 'error'"
              variant="subtle"
            >
              {{ row.original.is_active ? t('masterData.location.active') : t('masterData.location.inactive') }}
            </UBadge>
          </template>

          <template #actions-cell="{ row }">
            <div class="flex gap-1">
              <UButton
                icon="i-heroicons-pencil"
                color="primary"
                variant="ghost"
                size="sm"
                @click="store.handleEdit(row.original)"
              />
              <UButton
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="sm"
                @click="handleDelete(row.original)"
              />
            </div>
          </template>
        </TableEntities>
      </div>

      <!-- Mobile & Tablet View (< md): Stacked Cards List -->
      <div class="block md:hidden space-y-3">
        <div
          v-for="(loc, index) in paginatedLocations"
          :key="loc.id || index"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
        >
          <!-- Card Header: City/Name + Active Badge -->
          <div class="flex items-center justify-between gap-2">
            <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40">
              <UIcon name="i-heroicons-map-pin" class="size-3" />
              {{ loc.city || loc.name }}
            </span>
            <UBadge
              :color="loc.is_active ? 'success' : 'error'"
              variant="subtle"
              size="sm"
            >
              {{ loc.is_active ? t('masterData.location.active') : t('masterData.location.inactive') }}
            </UBadge>
          </div>

          <!-- Card Body: Location Name & Province / City -->
          <div>
            <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
              {{ loc.name }}
            </h4>
            <div class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400 mt-1">
              <UIcon name="i-heroicons-globe-asia-australia" class="size-3.5 text-slate-400 shrink-0" />
              <span class="truncate">{{ loc.province || '-' }} · {{ loc.city || '-' }}</span>
            </div>
            <p v-if="loc.address" class="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">
              {{ loc.address }}
            </p>
          </div>

          <!-- Card Footer: Index & Action Buttons -->
          <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs text-slate-400 font-mono">
              #{{ (locationPage - 1) * itemsPerPage + index + 1 }}
            </span>
            <div class="flex items-center gap-1">
              <UButton
                icon="i-heroicons-pencil"
                color="warning"
                variant="ghost"
                size="md"
                @click="store.handleEdit(loc)"
              />
              <UButton
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="md"
                @click="handleDelete(loc)"
              />
            </div>
          </div>
        </div>

        <!-- Mobile Pagination Controls -->
        <div v-if="store.filteredLocations.length > itemsPerPage" class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800">
          <span class="text-xs text-slate-500 dark:text-slate-400">
            {{ (locationPage - 1) * itemsPerPage + 1 }}-{{ Math.min(locationPage * itemsPerPage, store.filteredLocations.length) }} of {{ store.filteredLocations.length }}
          </span>
          <UPagination
            v-model:page="locationPage"
            :items-per-page="itemsPerPage"
            :total="store.filteredLocations.length"
            size="sm"
            active-color="primary"
            color="neutral"
            variant="outline"
            class="max-w-full overflow-x-auto"
          />
        </div>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { Location } from '~/types/master'
import { useLocationStore } from '~/stores/location'
import { useI18n } from '~/composables/useI18n'

const store = useLocationStore()
const toast = useAppToast()
const { t } = useI18n()

// Local search state
const searchInput = ref(store.search)
const locationPage = ref(1)
const itemsPerPage = ref(10)

const paginatedLocations = computed(() => {
  const start = (locationPage.value - 1) * itemsPerPage.value
  return store.filteredLocations.slice(start, start + itemsPerPage.value)
})

watch(() => store.search, () => {
  locationPage.value = 1
})

const handleSearch = () => {
  store.setSearch(searchInput.value)
}

const resetSearch = () => {
  searchInput.value = ''
  store.setSearch('')
}

const handleDelete = async (location: Location) => {
  const deleted = await store.handleDelete(location)
  if (deleted) {
    toast.success(t('masterData.location.deleted'))
  }
}

// Fetch data on mount
onMounted(() => {
  store.fetchLocations()
})
</script>
