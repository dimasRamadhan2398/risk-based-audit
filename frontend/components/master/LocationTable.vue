<template>
  <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
    <template #header>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">{{ t('masterData.location.listTitle') }}</h2>
        <UButton
          :label="t('masterData.location.add')"
          icon="i-heroicons-plus"
          color="primary"
          class="w-full sm:w-auto"
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
      />
      <UButton
        :label="t('masterData.location.search')"
        color="primary"
        variant="soft"
        @click="handleSearch"
      />
      <UButton
        :label="t('masterData.location.reset')"
        color="neutral"
        variant="ghost"
        @click="resetSearch"
      />
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
    <TableEntities
      v-else
      :data="store.filteredLocations"
      :columns="store.columns"
      :loading="store.loading"
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
  </UCard>
</template>

<script setup lang="ts">
import type { Location } from '~/types/master'
import { useLocationStore } from '~/stores/location'
import { useI18n } from '~/composables/useI18n'

const store = useLocationStore()
const toast = useAppToast()
const { t } = useI18n()

// Local search state
const searchInput = ref(store.search)

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
