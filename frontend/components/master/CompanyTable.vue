<template>
  <UCard class="rounded-xl shadow overflow-hidden" variant="soft" color="primary">
    <template #header>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">{{ t('masterData.company.list') }}</h2>
        <UButton
          :label="t('masterData.company.add')"
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
        :placeholder="t('masterData.company.searchPlaceholder')"
        icon="i-heroicons-magnifying-glass"
        class="w-full sm:w-72"
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
    <div v-else-if="store.companies.length === 0" class="py-8 text-center">
      <UIcon name="i-heroicons-building-office-2" class="text-4xl text-gray-400 mb-2" />
      <p class="text-gray-500">{{ t('masterData.company.empty') }}</p>
    </div>

    <!-- Table & Pagination via TableEntities -->
    <div v-else class="space-y-4">
      <!-- Desktop View (>= md): Full Table Entities -->
      <div class="hidden md:block">
        <TableEntities
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
              {{ row.original.company_code }}
            </span>
          </template>

          <template #company_name-cell="{ row }">
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ row.original.company_name }}
              </div>
              <div v-if="row.original.legal_name && row.original.legal_name !== row.original.company_name" class="text-xs text-gray-500">
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
              {{ t(`masterData.company.types.${row.original.company_type}`) }}
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
              {{ row.original.is_active ? t('masterData.company.active') : t('masterData.company.inactive') }}
            </UBadge>
          </template>

          <template #actions-cell="{ row }">
            <div class="flex gap-1">
              <UButton
                icon="i-heroicons-pencil"
                color="primary"
                variant="ghost"
                size="md"
                @click="store.handleEdit(row.original)"
              />
              <UButton
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="md"
                @click="store.handleDelete(row.original)"
              />
            </div>
          </template>
        </TableEntities>
      </div>

      <!-- Mobile & Tablet View (< md): Stacked Cards List -->
      <div class="block md:hidden space-y-3">
        <div
          v-for="(company, index) in store.companies"
          :key="company.id || index"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
        >
          <!-- Card Header: Code Badge + Active & Type Badges -->
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40">
                <UIcon name="i-heroicons-building-office-2" class="size-3" />
                {{ company.company_code }}
              </span>
              <UBadge
                :color="company.is_active ? 'success' : 'error'"
                variant="subtle"
                size="sm"
              >
                {{ company.is_active ? t('masterData.company.active') : t('masterData.company.inactive') }}
              </UBadge>
            </div>
            <UBadge
              :color="getTypeBadgeColor(company.company_type)"
              variant="subtle"
              size="sm"
            >
              {{ t(`masterData.company.types.${company.company_type}`) }}
            </UBadge>
          </div>

          <!-- Card Body: Name & Legal Name -->
          <div>
            <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
              {{ company.company_name }}
            </h4>
            <p v-if="company.legal_name && company.legal_name !== company.company_name" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              {{ company.legal_name }}
            </p>
          </div>

          <!-- Metadata: Tax ID -->
          <div v-if="company.tax_id" class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400 py-1.5 px-2.5 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-800">
            <UIcon name="i-heroicons-identification" class="size-3.5 text-gray-400 shrink-0" />
            <span class="font-mono text-[11px]"><strong class="font-bold text-gray-700 dark:text-gray-300">Tax ID:</strong> {{ company.tax_id }}</span>
          </div>

          <!-- Card Footer: Index & Action Buttons -->
          <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs text-slate-400 font-mono">
              #{{ (store.pagination.page - 1) * store.pagination.page_size + index + 1 }}
            </span>
            <div class="flex items-center gap-1">
              <UButton
                icon="i-heroicons-pencil"
                color="warning"
                variant="ghost"
                size="md"
                @click="store.handleEdit(company)"
              />
              <UButton
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="md"
                @click="store.handleDelete(company)"
              />
            </div>
          </div>
        </div>

        <!-- Mobile Pagination Controls -->
        <div v-if="store.pagination.total > store.pagination.page_size" class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800">
          <span class="text-xs text-slate-500 dark:text-slate-400">
            {{ (store.pagination.page - 1) * store.pagination.page_size + 1 }}-{{ Math.min(store.pagination.page * store.pagination.page_size, store.pagination.total) }} of {{ store.pagination.total }}
          </span>
          <UPagination
            :page="store.pagination.page"
            :items-per-page="store.pagination.page_size"
            :total="store.pagination.total"
            size="md"
            active-color="primary"
            color="neutral"
            variant="outline"
            class="max-w-full overflow-x-auto"
            @update:page="(p) => store.setPage(p)"
          />
        </div>
      </div>
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { useCompanyStore } from '~/stores/company'

const store = useCompanyStore()
const { t } = useI18n()

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
