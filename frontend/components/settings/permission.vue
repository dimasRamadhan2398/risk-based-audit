<template>
  <div class="space-y-6">
    <!-- Main Card Container -->
    <UCard class="border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xs">
      <template #header>
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h3 class="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon
                name="i-lucide-shield-check"
                class="w-6 h-6 text-primary-600 dark:text-primary-400"
              />
              {{ t('settings.permission.rbacTitle') }}
            </h3>
            <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.permission.rbacSubtitle') }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <UButton
              :label="t('settings.permission.exportMatrix')"
              icon="i-lucide-download"
              color="neutral"
              variant="outline"
              size="sm"
              @click="exportMatrix"
            />
          </div>
        </div>
      </template>

      <!-- Access Legend & Controls -->
      <div class="space-y-6 mb-6">
        <!-- Legend Pill Container -->
        <div class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-800">
          <h4 class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">
            {{ t('settings.permission.legendTitle') }}
          </h4>
          <div class="flex flex-wrap items-center gap-4 text-xs">
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-bold text-xs leading-none transition-colors bg-emerald-600 text-white shadow-xs">
                FULL
              </span>
              <span class="text-gray-600 dark:text-gray-300">{{ t('settings.permission.legendFull') }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-bold text-xs leading-none transition-colors bg-amber-500 text-white shadow-xs">
                LIMITED / ACTION
              </span>
              <span class="text-gray-600 dark:text-gray-300">{{ t('settings.permission.legendLimited') }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-bold text-xs leading-none transition-colors bg-sky-600 text-white shadow-xs">
                READ
              </span>
              <span class="text-gray-600 dark:text-gray-300">{{ t('settings.permission.legendRead') }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-bold text-xs leading-none transition-colors bg-slate-500 text-white shadow-xs">
                NONE
              </span>
              <span class="text-gray-600 dark:text-gray-300">{{ t('settings.permission.legendNone') }}</span>
            </div>
          </div>
        </div>

        <!-- Search Bar -->
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="w-full sm:w-80">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              :placeholder="t('settings.permission.searchPlaceholder')"
              class="w-full"
            />
          </div>
          <div class="flex items-center gap-2 w-full sm:w-auto text-xs text-gray-500 dark:text-gray-400 font-medium">
            {{ t('settings.permission.showingCount', { count: filteredMatrix.length, total: matrixData.length }) }}
          </div>
        </div>
      </div>

      <!-- Feature Access Matrix Table -->
      <div class="overflow-x-auto border border-gray-200 dark:border-gray-800 rounded-xl">
        <table class="w-full text-left text-sm">
          <thead class="bg-gray-100 dark:bg-gray-800/80 text-gray-900 dark:text-gray-100 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th class="py-3.5 px-4 font-bold w-12 text-center">
                #
              </th>
              <th class="py-3.5 px-4 font-bold min-w-[240px]">
                {{ t('settings.permission.thModule') }}
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px] bg-emerald-50/50 dark:bg-emerald-950/20">
                ADMIN
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px]">
                EXECUTIVE
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px]">
                AUDITOR
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px]">
                AUDITEE
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[150px]">
                DEPARTMENT_HEAD
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[120px]">
                VIEWER
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
            <tr
              v-for="item in filteredMatrix"
              :key="item.id"
              class="hover:bg-gray-50/80 dark:hover:bg-gray-800/40 transition-colors"
            >
              <!-- Module Number -->
              <td class="py-3 px-4 text-center font-medium text-gray-500 dark:text-gray-400">
                {{ item.num }}
              </td>

              <!-- Sidebar Module / Feature Name -->
              <td class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <span
                    v-if="item.isSubmodule"
                    class="text-gray-400 dark:text-gray-600 font-mono text-sm ml-2"
                  >
                    ↳
                  </span>
                  <div>
                    <span
                      class="font-semibold text-gray-900 dark:text-gray-100"
                      :class="{ 'text-sm tracking-tight text-gray-500 dark:text-gray-400 font-bold': !item.isSubmodule }"
                    >
                      {{ item.isSubmodule ? item.submodule : item.module }}
                    </span>
                    <p
                      v-if="item.isSubmodule && item.module"
                      class="text-xs text-gray-400 dark:text-gray-500"
                    >
                      {{ item.module }}
                    </p>
                  </div>
                </div>
              </td>

              <!-- ADMIN Column -->
              <td class="py-3 px-4 text-center bg-emerald-50/20 dark:bg-emerald-950/10">
                <span :class="getBadgeClass(item.admin)">
                  {{ item.admin }}
                </span>
              </td>

              <!-- EXECUTIVE Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.executive)">
                  {{ item.executive }}
                </span>
              </td>

              <!-- AUDITOR Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.auditor)">
                  {{ item.auditor }}
                </span>
              </td>

              <!-- AUDITEE Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.auditee)">
                  {{ item.auditee }}
                </span>
              </td>

              <!-- DEPARTMENT_HEAD Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.departmentHead)">
                  {{ item.departmentHead }}
                </span>
              </td>

              <!-- VIEWER Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.viewer)">
                  {{ item.viewer }}
                </span>
              </td>
            </tr>

            <tr v-if="filteredMatrix.length === 0">
              <td
                colspan="8"
                class="py-8 text-center text-gray-500 dark:text-gray-400"
              >
                {{ t('settings.permission.noMatch', { query: searchQuery }) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { matrixData, type MatrixItem, getBadgeClass } from '~/utils/rbacMatrix'

const { t } = useI18n()

// Role columns use the backend role names (auth-service seeder). The values are the static
// matrix as previously published; EXECUTIVE and DEPARTMENT_HEAD reuse the former merged columns
// ("Admin / CAE" and "Auditee / Dept. Head"). The backend role/permission endpoints (GET /roles/all)
// list permissions per resource, not this FULL/READ grid, so the grid is still static.
const searchQuery = ref('')

const filteredMatrix = computed(() => {
  return matrixData.filter((item) => {
    const q = searchQuery.value.toLowerCase().trim()
    if (!q) return true
    return (
      item.module.toLowerCase().includes(q)
      || item.submodule.toLowerCase().includes(q)
    )
  })
})

function exportMatrix() {
  const jsonStr = JSON.stringify(matrixData, null, 2)
  const blob = new Blob([jsonStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'AuditSphere_RBAC_Matrix.json'
  a.click()
  URL.revokeObjectURL(url)
}
</script>
