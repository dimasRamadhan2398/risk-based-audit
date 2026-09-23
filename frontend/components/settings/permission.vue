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
              AuditSphere Role-Based Access Control (RBAC) Matrix
            </h3>
            <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
              Official permission matrix defining access rights and operational authority across system modules.
            </p>
          </div>
          <div class="flex items-center gap-2">
            <UButton
              label="Export Matrix"
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
            Access Level Legend
          </h4>
          <div class="flex flex-wrap items-center gap-4 text-xs">
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-semibold text-xs leading-none transition-colors border bg-emerald-800 text-emerald-700 border-emerald-200 dark:bg-emerald-800 dark:text-emerald-300 dark:border-emerald-800">
                FULL
              </span>
              <span class="text-gray-600 dark:text-gray-300">Full Access (Create, Read, Update, Delete, Approval)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-semibold text-xs leading-none transition-colors border bg-amber-800 text-amber-700 border-amber-200 dark:bg-amber-800 dark:text-amber-300 dark:border-amber-800">
                LIMITED / ACTION
              </span>
              <span class="text-gray-600 dark:text-gray-300">Restricted operational action (Draft, Edit, Respond, Review, etc.)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-semibold text-xs leading-none transition-colors border bg-slate-800 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700">
                READ
              </span>
              <span class="text-gray-600 dark:text-gray-300">Read-Only access</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="inline-flex items-center justify-center px-2.5 py-1 rounded-md font-semibold text-xs leading-none transition-colors border bg-slate-800 text-slate-400 border-slate-200 dark:bg-slate-800 dark:text-slate-500 dark:border-slate-800">
                NONE
              </span>
              <span class="text-gray-600 dark:text-gray-300">No access / Hidden from navigation</span>
            </div>
          </div>
        </div>

        <!-- Search Bar -->
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="w-full sm:w-80">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              placeholder="Search module or feature..."
              class="w-full"
            />
          </div>
          <div class="flex items-center gap-2 w-full sm:w-auto text-xs text-gray-500 dark:text-gray-400 font-medium">
            Showing {{ filteredMatrix.length }} of {{ matrixData.length }} module features
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
                Sidebar Module / Feature
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px] bg-emerald-50/50 dark:bg-emerald-950/20">
                Admin / CAE
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px]">
                Audit Manager
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[130px]">
                Auditor
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[150px]">
                Auditee / Dept. Head
              </th>
              <th class="py-3.5 px-4 font-bold text-center min-w-[120px]">
                Viewer
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

              <!-- Admin / CAE Column -->
              <td class="py-3 px-4 text-center bg-emerald-50/20 dark:bg-emerald-950/10">
                <span :class="getBadgeClass(item.admin)">
                  {{ item.admin }}
                </span>
              </td>

              <!-- Audit Manager Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.manager)">
                  {{ item.manager }}
                </span>
              </td>

              <!-- Auditor Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.auditor)">
                  {{ item.auditor }}
                </span>
              </td>

              <!-- Auditee / Dept Head Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.auditee)">
                  {{ item.auditee }}
                </span>
              </td>

              <!-- Viewer Column -->
              <td class="py-3 px-4 text-center">
                <span :class="getBadgeClass(item.viewer)">
                  {{ item.viewer }}
                </span>
              </td>
            </tr>

            <tr v-if="filteredMatrix.length === 0">
              <td
                colspan="7"
                class="py-8 text-center text-gray-500 dark:text-gray-400"
              >
                No matching features found for "{{ searchQuery }}".
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

interface MatrixItem {
  id: string
  num: number
  module: string
  submodule: string
  isSubmodule: boolean
  admin: string
  manager: string
  auditor: string
  auditee: string
  viewer: string
}

const searchQuery = ref('')

const matrixData = ref<MatrixItem[]>([
  { id: '1', num: 1, module: 'Dashboard', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '2', num: 2, module: 'Audit Charter', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'READ', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '3-1', num: 3, module: 'Risk Profile', submodule: 'Corporate Risk Profile', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'FULL', auditee: 'EDIT — Own Dept.', viewer: 'READ' },
  { id: '3-2', num: 3, module: 'Risk Profile', submodule: 'Risk Appetite Statement', isSubmodule: true, admin: 'FULL', manager: 'READ', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '3-3', num: 3, module: 'Risk Profile', submodule: 'Risk Factors & Scoring', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '3-4', num: 3, module: 'Risk Profile', submodule: 'Audit Universe', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '3-5', num: 3, module: 'Risk Profile', submodule: 'Risk Control Matrix (RCM)', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'FULL', auditee: 'READ', viewer: 'READ' },
  { id: '4', num: 4, module: 'Strategic Audit Plan', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'EDIT', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '5-1', num: 5, module: 'Annual Audit Plan', submodule: 'Create / Manage Plan', isSubmodule: true, admin: 'FULL', manager: 'DRAFT', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '5-2', num: 5, module: 'Annual Audit Plan', submodule: 'Import Plan Document', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'NONE', auditee: 'NONE', viewer: 'NONE' },
  { id: '5-3', num: 5, module: 'Annual Audit Plan', submodule: 'Execution Status', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '6-1', num: 6, module: 'Audit Activity Plan', submodule: 'Create Activity Plan', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'DRAFT', auditee: 'NONE', viewer: 'READ' },
  { id: '6-2', num: 6, module: 'Audit Activity Plan', submodule: 'Import Activity Plan', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'NONE', auditee: 'NONE', viewer: 'NONE' },
  { id: '7-1', num: 7, module: 'Assignment Letter', submodule: 'Create / Publish Letter', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '7-2', num: 7, module: 'Assignment Letter', submodule: 'Import Letter', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'NONE', auditee: 'NONE', viewer: 'NONE' },
  { id: '8-1', num: 8, module: 'Audit Fieldwork', submodule: 'Fieldwork — Interviews, Sampling, Testing, etc.', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'FULL', auditee: 'RESPOND', viewer: 'READ' },
  { id: '8-2', num: 8, module: 'Audit Fieldwork', submodule: 'Working Papers — F-01 to F-05', isSubmodule: true, admin: 'FULL', manager: 'REVIEW', auditor: 'CREATE', auditee: 'NONE', viewer: 'READ' },
  { id: '8-3', num: 8, module: 'Audit Fieldwork', submodule: 'Import Working Papers', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'UPLOAD', auditee: 'NONE', viewer: 'NONE' },
  { id: '9-1', num: 9, module: 'Audit Result Report', submodule: 'Result Report / LHA', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'DRAFT', auditee: 'READ', viewer: 'READ' },
  { id: '9-2', num: 9, module: 'Audit Result Report', submodule: 'Executive Summary', isSubmodule: true, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '9-3', num: 9, module: 'Audit Result Report', submodule: 'Client Satisfaction Survey', isSubmodule: true, admin: 'FULL', manager: 'READ', auditor: 'READ', auditee: 'FILL SURVEY', viewer: 'READ' },
  { id: '10', num: 10, module: 'Action Taken Report (ATR)', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'REVIEW', auditor: 'READ', auditee: 'UPDATE CAPA', viewer: 'READ' },
  { id: '11', num: 11, module: 'KPI Performance', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '12', num: 12, module: 'Consulting Service', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'FULL', auditor: 'FULL', auditee: 'REQUEST', viewer: 'READ' },
  { id: '13', num: 13, module: 'Quality Assurance (QA)', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'NONE', viewer: 'READ' },
  { id: '14', num: 14, module: 'Analytics', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'FULL', auditor: 'READ', auditee: 'READ', viewer: 'READ' },
  { id: '15', num: 15, module: 'Settings & User Management', submodule: '', isSubmodule: false, admin: 'FULL', manager: 'NONE', auditor: 'NONE', auditee: 'NONE', viewer: 'NONE' }
])

const filteredMatrix = computed(() => {
  return matrixData.value.filter((item) => {
    const q = searchQuery.value.toLowerCase().trim()
    if (!q) return true
    return (
      item.module.toLowerCase().includes(q)
      || item.submodule.toLowerCase().includes(q)
    )
  })
})

function getBadgeClass(val: string): string {
  const base =
    'inline-flex items-center justify-center px-2.5 py-1 rounded-md font-semibold text-xs leading-none transition-colors border'

  if (val === 'FULL') {
    return (
      base +
      ' bg-emerald-800 text-emerald-700 border-emerald-200' +
      ' dark:bg-emerald-800 dark:text-emerald-300 dark:border-emerald-800'
    )
  }

  if (val === 'READ') {
    return (
      base +
      ' bg-slate-800 text-slate-600 border-slate-200' +
      ' dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700'
    )
  }

  if (val === 'NONE') {
    return (
      base +
      ' bg-slate-800 text-slate-400 border-slate-200' +
      ' dark:bg-slate-800 dark:text-slate-500 dark:border-slate-800'
    )
  }

  // Action-specific / Limited access
  return (
    base +
    ' bg-amber-800 text-amber-700 border-amber-200' +
    ' dark:bg-amber-800 dark:text-amber-300 dark:border-amber-800'
  )
}

function exportMatrix() {
  const jsonStr = JSON.stringify(matrixData.value, null, 2)
  const blob = new Blob([jsonStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'AuditSphere_RBAC_Matrix.json'
  a.click()
  URL.revokeObjectURL(url)
}
</script>
