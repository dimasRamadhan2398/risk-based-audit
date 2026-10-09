<template>
  <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
    <template #header>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">Company List</h2>
        <UButton
          label="Add Company"
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
        placeholder="Search company code, name, or tax ID..."
        icon="i-heroicons-magnifying-glass"
        class="w-full sm:w-72"
        @keyup.enter="handleSearch"
      />
      <UButton
        label="Search"
        color="primary"
        variant="soft"
        @click="handleSearch"
      />
      <UButton
        label="Reset"
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
    <div v-else-if="store.companies.length === 0" class="py-8 text-center">
      <UIcon name="i-heroicons-building-office-2" class="text-4xl text-gray-400 mb-2" />
      <p class="text-gray-500">No companies found.</p>
    </div>

    <!-- Table & Pagination via TableEntities -->
    <TableEntities
      v-else
      :data="store.companies"
      :columns="store.columns"
      :loading="store.loading"
      :server-side="true"
      :total="store.pagination.total"
      :items-per-page="store.pagination.page_size"
      :page="store.pagination.page"
      @update:page="(p) => store.setPage(p)"
      @update:items-per-page="(size) => store.setPageSize(size)"
    >
      <template #company_code-cell="{ row }">
        <span class="font-semibold text-primary-600 dark:text-primary-400">
          {{ row.original.company_code || row.original.code }}
        </span>
      </template>

      <template #company_name-cell="{ row }">
        <div>
          <div class="font-medium text-gray-900 dark:text-white">
            {{ row.original.company_name || row.original.name }}
          </div>
          <div v-if="row.original.legal_name && row.original.legal_name !== (row.original.company_name || row.original.name)" class="text-xs text-gray-500">
            {{ row.original.legal_name }}
          </div>
        </div>
      </template>

      <template #legal_name-cell="{ row }">
        <span class="text-sm text-gray-600 dark:text-gray-400">
          {{ row.original.legal_name || '-' }}
        </span>
      </template>

      <template #company_type-cell="{ row }">
        <UBadge
          :color="getTypeBadgeColor(row.original.company_type)"
          variant="subtle"
          size="sm"
        >
          {{ row.original.company_type }}
        </UBadge>
      </template>

      <template #tax_id-cell="{ row }">
        <span class="text-sm font-mono text-gray-600 dark:text-gray-400">
          {{ row.original.tax_id || '-' }}
        </span>
      </template>

      <template #is_active-cell="{ row }">
        <UBadge
          :color="row.original.is_active ? 'success' : 'error'"
          variant="subtle"
          size="sm"
        >
          {{ row.original.is_active ? 'Active' : 'Inactive' }}
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
            @click="store.handleDelete(row.original)"
          />
        </div>
      </template>
    </TableEntities>
  </UCard>
</template>

<script setup lang="ts">
import { useCompanyStore } from '~/stores/company'

const store = useCompanyStore()

const searchInput = ref(store.search)

const handleSearch = () => {
  store.setSearch(searchInput.value)
}

const resetSearch = () => {
  searchInput.value = ''
  store.setSearch('')
}

const getTypeBadgeColor = (type?: string) => {
  switch (type?.toUpperCase()) {
    case 'HOLDING':
      return 'secondary'
    case 'SUBSIDIARY':
      return 'primary'
    case 'BRANCH':
      return 'info'
    default:
      return 'neutral'
  }
}

onMounted(() => {
  store.fetchCompanies()
})
</script>
