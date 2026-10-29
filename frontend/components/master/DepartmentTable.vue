<template>
  <UCard
    class="rounded-xl shadow overflow-hidden"
    variant="soft"
    color="primary"
  >
    <template #header>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">{{ t('masterData.department.listTitle') }}</h2>
        <UButton v-if="canWriteMasterData"
          :label="t('masterData.department.add')"
          icon="i-heroicons-plus"
          color="primary"
          class="w-full sm:w-auto"
          @click="store.openCreateModal"
        />
      </div>
    </template>

    <!-- Search & Filter -->
    <div class="mb-4 flex flex-wrap gap-2 sm:gap-4 items-center w-full">
      <CommonSearchInput
        v-model="searchInput"
        :placeholder="t('masterData.department.searchPlaceholder')"
        class="w-full sm:w-64"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      />
      <UButton
        :label="t('masterData.department.search')"
        color="primary"
        variant="soft"
        @click="handleSearch"
      />
      <UButton
        :label="t('masterData.department.reset')"
        color="neutral"
        variant="outline"
        @click="resetSearch"
      />
    </div>

    <!-- Loading State -->
    <div
      v-if="store.loading"
      class="flex justify-center py-8"
    >
      <ULoadingIcon />
    </div>

    <!-- Error State -->
    <div
      v-else-if="store.errorMsg"
      class="py-4"
    >
      <UAlert
        color="error"
        variant="soft"
        :title="store.errorMsg"
        icon="i-heroicons-exclamation-circle"
      />
    </div>

    <!-- Empty State -->
    <div
      v-else-if="store.departments.length === 0"
      class="py-8 text-center"
    >
      <UIcon
        name="i-heroicons-building-office"
        class="text-4xl text-gray-400 mb-2"
      />
      <p class="text-gray-500">
        {{ t('masterData.department.empty') }}
      </p>
    </div>

    <!-- Table & Pagination via TableEntities -->
    <TableEntities
      v-else
      :data="store.departments"
      :columns="store.columns"
      :loading="store.loading"
      :server-side="true"
      :total="store.pagination.total"
      :items-per-page="store.pagination.page_size"
      :page="store.pagination.page"
      @update:page="(p) => store.setPage(p)"
      @update:items-per-page="(size) => store.setPageSize(size)"
    >
      <template #department_code-cell="{ row }">
        <span class="font-medium text-primary-600">{{ row.original.department_code }}</span>
      </template>

      <template #department_name-cell="{ row }">
        <div class="font-medium">
          {{ row.original.department_name }}
        </div>
      </template>

      <template #company_id-cell="{ row }">
        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ options.labelFor(options.companyOptions, row.original.company_id) || '-' }}
        </span>
      </template>

      <template #department_description-cell="{ row }">
        <span class="text-gray-600 text-sm">
          {{ row.original.department_description || '-' }}
        </span>
      </template>

      <template #level-cell="{ row }">
        <UBadge
          color="neutral"
          variant="soft"
        >
          {{ t('masterData.department.level', { level: row.original.level }) }}
        </UBadge>
      </template>

      <template #is_active-cell="{ row }">
        <UBadge
          :color="row.original.is_active ? 'success' : 'error'"
          variant="subtle"
        >
          {{ row.original.is_active ? t('masterData.department.active') : t('masterData.department.inactive') }}
        </UBadge>
      </template>

      <template #actions-cell="{ row }">
        <div class="flex gap-1">
          <UButton v-if="canWriteMasterData"
            icon="i-heroicons-pencil"
            color="primary"
            variant="ghost"
            size="sm"
            @click="store.handleEdit(row.original)"
          />
          <UButton v-if="canWriteMasterData"
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
import { useRbac } from '~/composables/useRbac'
import { useDepartmentStore } from '~/stores/department'
import { useMasterOptionsStore } from '~/stores/master-options'

const store = useDepartmentStore()
const options = useMasterOptionsStore()
const { t } = useI18n()
const { canWriteMasterData } = useRbac()

// Local search state
const searchInput = ref(store.search)
const pageSize = ref(store.pagination.page_size)

const handleSearch = () => {
  store.setSearch(searchInput.value)
}

const resetSearch = () => {
  searchInput.value = ''
  store.setSearch('')
}

const handlePageSizeChange = (value: any) => {
  store.setPageSize(Number(value))
}

// Fetch data on mount
onMounted(() => {
  options.fetchAll()
  store.fetchDepartments()
})
</script>
