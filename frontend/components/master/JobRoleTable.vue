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
          class="w-full sm:w-auto justify-center text-center"
          @click="store.openCreateModal"
        />
      </div>
    </div>

    <!-- Search & Filter -->
    <div class="mb-4 flex flex-wrap gap-2 sm:gap-4 items-center w-full">
      <CommonSearchInput
        v-model="searchInput"
        :placeholder="t('masterData.jobRole.searchPlaceholder')"
        class="w-full sm:w-72"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
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

    <!-- Table & Pagination via TableEntities -->
    <div v-else class="space-y-4">
      <!-- Desktop View (>= md): Full Table Entities -->
      <div class="hidden md:block">
        <TableEntities
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
            <span class="font-semibold text-primary-600 dark:text-primary-400">
              {{ row.original.job_role_code }}
            </span>
          </template>

          <template #job_role_name-cell="{ row }">
            <div>
              <div class="font-medium text-gray-900 dark:text-white">
                {{ row.original.job_role_name }}
              </div>
              <div
                v-if="row.original.job_role_description"
                class="text-xs text-gray-500 dark:text-gray-400"
              >
                {{ row.original.job_role_description }}
              </div>
            </div>
          </template>

          <template #job_position_type-cell="{ row }">
            <span class="text-sm text-gray-600 dark:text-gray-400">
              {{ labelOrRaw(t, `masterData.jobRole.positionTypes.${row.original.job_position_type}`, row.original.job_position_type) }}
            </span>
          </template>

          <template #default_access_role-cell="{ row }">
            <UBadge
              v-if="row.original.default_access_role"
              :color="accessRoleColor(row.original.default_access_role)"
              variant="subtle"
              size="sm"
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
              :color="row.original.is_active ? 'secondary' : 'error'"
              variant="subtle"
              size="sm"
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

      <!-- Mobile & Tablet View (< md): Stacked Cards List -->
      <div class="block md:hidden space-y-3">
        <div
          v-for="(jobRole, index) in store.paged"
          :key="jobRole.id || index"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
        >
          <!-- Card Header: Code Badge + Active & Access Role Badges -->
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40">
                <UIcon name="i-heroicons-briefcase" class="size-3" />
                {{ jobRole.job_role_code }}
              </span>
              <UBadge
                :color="jobRole.is_active ? 'secondary' : 'error'"
                variant="subtle"
                size="sm"
              >
                {{ jobRole.is_active ? t('masterData.jobRole.active') : t('masterData.jobRole.inactive') }}
              </UBadge>
            </div>
            <UBadge
              v-if="jobRole.default_access_role"
              :color="accessRoleColor(jobRole.default_access_role)"
              variant="subtle"
              size="sm"
            >
              {{ labelOrRaw(t, `masterData.jobRole.accessRoles.${jobRole.default_access_role}`, jobRole.default_access_role) }}
            </UBadge>
          </div>

          <!-- Card Body: Name & Description -->
          <div>
            <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
              {{ jobRole.job_role_name }}
            </h4>
            <p
              v-if="jobRole.job_role_description"
              class="text-xs text-gray-500 dark:text-gray-400 mt-1 italic line-clamp-2"
            >
              {{ jobRole.job_role_description }}
            </p>
          </div>

          <!-- Metadata: Position Type -->
          <div
            v-if="jobRole.job_position_type"
            class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400 py-1.5 px-2.5 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-800"
          >
            <UIcon name="i-heroicons-user" class="size-3.5 text-gray-400 shrink-0" />
            <span class="text-[11px]">
              <strong class="font-bold text-gray-700 dark:text-gray-300">{{ t('masterData.jobRole.columns.positionType') }}:</strong>
              {{ labelOrRaw(t, `masterData.jobRole.positionTypes.${jobRole.job_position_type}`, jobRole.job_position_type) }}
            </span>
          </div>

          <!-- Card Footer: Index & Action Buttons -->
          <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs text-slate-400 font-mono">
              #{{ (store.page - 1) * store.pageSize + index + 1 }}
            </span>
            <div
              v-show="canWriteMasterData"
              class="flex items-center gap-1"
            >
              <UButton
                icon="i-heroicons-pencil"
                color="warning"
                variant="ghost"
                size="md"
                :aria-label="t('common.edit')"
                @click="store.handleEdit(jobRole)"
              />
              <UButton
                icon="i-heroicons-trash"
                color="error"
                variant="ghost"
                size="md"
                :aria-label="t('common.delete')"
                @click="store.handleDelete(jobRole)"
              />
            </div>
          </div>
        </div>

        <!-- Mobile Pagination Controls -->
        <div
          v-if="store.filtered.length > store.pageSize"
          class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
        >
          <span class="text-xs text-slate-500 dark:text-slate-400">
            {{ (store.page - 1) * store.pageSize + 1 }}-{{ Math.min(store.page * store.pageSize, store.filtered.length) }} of {{ store.filtered.length }}
          </span>
          <UPagination
            :page="store.page"
            :items-per-page="store.pageSize"
            :total="store.filtered.length"
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
