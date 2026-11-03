<template>
  <div class="space-y-4">
    <div>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 class="text-lg font-bold">
          {{ t('masterData.jobRole.listTitle') }}
        </h2>
        <UButton
          v-if="canWriteMasterData"
          :label="t('masterData.jobRole.add')"
          icon="i-heroicons-plus"
          color="primary"
          class="w-full sm:w-auto"
          @click="store.openCreateModal"
        />
      </div>
    </div>

    <!-- Search -->
    <div class="mb-4 flex flex-wrap gap-2 sm:gap-4 items-center w-full">
      <CommonSearchInput
        v-model="searchInput"
        :placeholder="t('masterData.jobRole.searchPlaceholder')"
        class="w-full sm:w-64"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      />
      <UButton
        :label="t('masterData.jobRole.search')"
        color="primary"
        variant="soft"
        @click="handleSearch"
      />
      <UButton
        :label="t('masterData.jobRole.reset')"
        color="neutral"
        variant="outline"
        @click="resetSearch"
      />
    </div>

    <div
      v-if="store.loading"
      class="flex justify-center py-8"
    >
      <ULoadingIcon />
    </div>

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

    <div
      v-else-if="store.filtered.length === 0"
      class="py-8 text-center"
    >
      <UIcon
        name="i-heroicons-briefcase"
        class="text-4xl text-gray-400 mb-2"
      />
      <p class="text-gray-500">
        {{ t('masterData.jobRole.empty') }}
      </p>
    </div>

    <TableEntities
      v-else
      :data="store.paged"
      :columns="store.columns"
      :loading="store.loading"
      :server-side="true"
      :total="store.filtered.length"
      :items-per-page="store.pageSize"
      :page="store.page"
      @update:page="(p) => store.setPage(p)"
      @update:items-per-page="(size) => store.setPageSize(size)"
    >
      <template #job_role_code-cell="{ row }">
        <span class="font-medium text-primary-600">{{ row.original.job_role_code }}</span>
      </template>

      <template #job_role_name-cell="{ row }">
        <div class="font-medium">
          {{ row.original.job_role_name }}
        </div>
        <div
          v-if="row.original.job_role_description"
          class="text-xs text-gray-500"
        >
          {{ row.original.job_role_description }}
        </div>
      </template>

      <template #job_position_type-cell="{ row }">
        <span class="text-sm">
          {{ labelOrRaw(t, `masterData.jobRole.positionTypes.${row.original.job_position_type}`, row.original.job_position_type) }}
        </span>
      </template>

      <template #default_access_role-cell="{ row }">
        <UBadge
          v-if="row.original.default_access_role"
          :color="accessRoleColor(row.original.default_access_role)"
          variant="subtle"
        >
          {{ labelOrRaw(t, `masterData.jobRole.accessRoles.${row.original.default_access_role}`, row.original.default_access_role) }}
        </UBadge>
        <span
          v-else
          class="text-gray-400 text-sm"
        >{{ t('masterData.jobRole.noAccessRole') }}</span>
      </template>

      <template #is_active-cell="{ row }">
        <UBadge
          :color="row.original.is_active ? 'success' : 'error'"
          variant="subtle"
        >
          {{ row.original.is_active ? t('masterData.jobRole.active') : t('masterData.jobRole.inactive') }}
        </UBadge>
      </template>

      <template #actions-cell="{ row }">
        <div class="flex gap-1">
          <UButton
            v-if="canWriteMasterData"
            icon="i-heroicons-pencil"
            color="primary"
            variant="ghost"
            size="sm"
            :aria-label="t('common.edit')"
            @click="store.handleEdit(row.original)"
          />
          <UButton
            v-if="canWriteMasterData"
            icon="i-heroicons-trash"
            color="error"
            variant="ghost"
            size="sm"
            :aria-label="t('common.delete')"
            @click="store.handleDelete(row.original)"
          />
        </div>
      </template>
    </TableEntities>
  </div>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { useJobRoleStore } from '~/stores/job-role'
import { accessRoleColor, labelOrRaw } from '~/utils/jobRole'

const store = useJobRoleStore()
const { t } = useI18n()
const { canWriteMasterData } = useRbac()

const searchInput = ref(store.search)

const handleSearch = () => {
  store.setSearch(searchInput.value)
}

const resetSearch = () => {
  searchInput.value = ''
  store.setSearch('')
}

onMounted(() => {
  store.fetchJobRoles()
})
</script>
