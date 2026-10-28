<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-8 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[var(--border-main)] pb-5">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white font-space">
          {{ t('auditUniverse.title') }}
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          {{ t('auditUniverse.subtitle') }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <UButton
          to="/risk-profile"
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="outline"
          class="w-full sm:w-auto"
        >
          {{ t('auditUniverse.backToHeatmap') }}
        </UButton>
      </div>
    </div>

    <!-- Alert / Toast -->
    <Transition name="fade">
      <UAlert
        v-if="alertMessage"
        :color="alertType === 'success' ? 'success' : 'error'"
        variant="solid"
        :title="alertType === 'success' ? t('common.success') : t('common.error')"
        :description="alertMessage"
        icon="i-lucide-info"
        class="shadow-md"
        closable
        @close="alertMessage = ''"
      />
    </Transition>

    <!-- Tabs Navigation -->
    <UTabs v-model="activeTab" :items="tabItems" class="w-full">
      
      <!-- Tab 1: Corporate Universe Builder -->
      <template #library>
        <div class="grid grid-cols-1 lg:grid-cols-1 gap-6 mt-6">
          <!-- Left Column: Standard Library Explorer -->
          <div class="space-y-6">
            <UCard class="shadow-sm border border-[var(--border-main)] h-full">
              <template #header>
                <div>
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
                    {{ t('auditUniverse.library.standardTitle') }}
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    {{ t('auditUniverse.library.standardSubtitle') }}
                  </p>
                </div>
              </template>

              <!-- Standard Library Tree -->
              <div class="max-h-[500px] overflow-y-auto pr-2 space-y-4">
                <div 
                  v-for="node in standardUniverse" 
                  :key="node.id"
                  class="border border-slate-100 dark:border-slate-800 rounded-xl p-4 bg-slate-50/30 dark:bg-slate-900/10 space-y-3"
                >
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <UCheckbox
                        :model-value="isStandardNodeSelected(node.id)"
                        :disabled="!canEditAuditUniverse"
                        @update:model-value="toggleStandardSelection(node)"
                      />
                      <span class="font-bold text-sm text-slate-800 dark:text-slate-200">
                        {{ node.name }}
                      </span>
                    </div>
                  </div>

                  <!-- Children (Sub-entities) as checkboxes -->
                  <div v-if="node.children && node.children.length > 0" class="pl-6 border-l border-slate-200 dark:border-slate-800 space-y-2">
                    <div 
                      v-for="sub in node.children" 
                      :key="sub.id"
                      class="flex items-center justify-between p-1 hover:bg-slate-50 dark:hover:bg-slate-800/40 rounded transition-colors"
                    >
                      <div class="flex items-center gap-2">
                        <UCheckbox
                          :model-value="isStandardNodeSelected(sub.id)"
                          :disabled="!canEditAuditUniverse"
                          @update:model-value="toggleStandardSelection(sub)"
                        />
                        <span class="text-md font-semibold text-slate-600 dark:text-slate-400">
                          {{ sub.name }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </UCard>
          </div>

          <!-- Right Column: Corporate Universe Explorer & Editor -->
          <div>
            <UCard class="shadow-sm border border-[var(--border-main)] h-full">
              <template #header>
                <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space">
                    {{ t('auditUniverse.library.corporateTitle') }}
                  </h2>
                  <div class="flex flex-wrap items-center gap-2">
                    <UTooltip :text="t('auditUniverse.library.syncMasterDataTooltip')">
                      <UButton
                        v-if="canEditAuditUniverse"
                        size="md"
                        color="neutral"
                        variant="outline"
                        icon="i-lucide-refresh-cw"
                        :loading="syncingMaster"
                        :label="t('auditUniverse.library.syncMasterData')"
                        class="w-full sm:w-auto"
                        @click="handleSyncMasterData"
                      />
                    </UTooltip>
                    <UButton
                      v-if="canEditAuditUniverse"
                      size="md"
                      color="primary"
                      icon="i-lucide-plus"
                      :label="t('auditUniverse.library.addCustomEntity')"
                      class="w-full sm:w-auto"
                      @click="openAddCustomModal(null)"
                    />
                    <UBadge color="primary" variant="subtle">{{ t('auditUniverse.library.entityCount', { count: corporateUniverse.length }) }}</UBadge>
                  </div>
                </div>
              </template>

              <!-- Corporate Universe Tree -->
              <div v-if="corporateUniverse.length > 0" class="max-h-[500px] overflow-y-auto pr-2 space-y-3">
                <div 
                  v-for="node in corporateUniverse" 
                  :key="node.id"
                  class="border border-slate-150 dark:border-slate-800/70 rounded-xl p-3 bg-white dark:bg-slate-900/40 space-y-2"
                >
                  <div class="flex items-center justify-between">
                    <span class="font-bold text-md text-slate-800 dark:text-slate-200">
                      {{ node.name }}
                      <UBadge v-if="departmentLabel(node)" color="neutral" variant="subtle" size="sm" class="ml-2 align-middle">{{ departmentLabel(node) }}</UBadge>
                    </span>
                    <div v-if="canEditNode(node)" class="flex items-center gap-1">
                      <UTooltip :text="t('auditUniverse.library.addSubEntity')">
                        <UButton
                          icon="i-lucide-plus"
                          color="primary"
                          variant="ghost"
                          size="md"
                        @click="openAddCustomModal(node.id)"
                      />
                      </UTooltip>
                      <UTooltip :text="t('auditUniverse.library.renameParentEntity')">
                        <UButton
                        icon="i-lucide-edit"
                        color="warning"
                        variant="ghost"
                        size="md"
                        @click="openRenameModal(node)"
                      />
                      </UTooltip>
                      <UTooltip :text="t('auditUniverse.library.deleteParentEntity')">
                        <UButton
                          icon="i-lucide-trash-2"
                          color="error"
                          variant="ghost"
                          size="md"
                          @click="deleteCorporateNode(node.id)"
                        />
                      </UTooltip>
                    </div>
                  </div>

                  <!-- Children (Sub-entities) -->
                  <div v-if="node.children && node.children.length > 0" class="pl-4 border-l border-slate-100 dark:border-slate-800 space-y-1">
                    <div 
                      v-for="sub in node.children" 
                      :key="sub.id"
                      class="flex items-center justify-between p-1 hover:bg-slate-50 dark:hover:bg-slate-800/30 rounded"
                    >
                      <span class="text-md text-slate-600 dark:text-slate-400">
                        {{ sub.name }}
                        <UBadge v-if="departmentLabel(sub)" color="neutral" variant="subtle" size="sm" class="ml-2 align-middle">{{ departmentLabel(sub) }}</UBadge>
                      </span>
                      <div v-if="canEditNode(sub)" class="flex items-center gap-1">
                        <UTooltip :text="t('auditUniverse.library.renameSubEntity')">
                          <UButton
                            icon="i-lucide-edit"
                            color="warning"
                            variant="ghost"
                            size="md"
                            @click="openRenameModal(sub)"
                          />
                        </UTooltip>
                        <UTooltip :text="t('auditUniverse.library.deleteSubEntity')">
                          <UButton
                            icon="i-lucide-trash-2"
                            color="error"
                            variant="ghost"
                            size="md"
                            @click="deleteCorporateNode(sub.id)"
                          />
                        </UTooltip>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="text-center py-16 text-slate-400 text-md">
                {{ t('auditUniverse.library.emptyState') }}
              </div>
            </UCard>
          </div>
        </div>

        <!-- Scoped Modals inside Tab 1 library template -->
        <UModal 
          v-model:open="renameModalOpen"
          :ui="{
            content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
          }"
        >
          <template #content>
            <UCard>
              <template #header>
                <h3 class="font-bold text-base text-slate-800 dark:text-slate-100">{{ t('auditUniverse.modals.renameTitle') }}</h3>
              </template>
              <div class="space-y-4">
                <UFormField :label="t('auditUniverse.modals.departmentLabel')" class="space-y-1">
                  <USelectMenu
                    v-model="renameDepartmentId"
                    :items="departmentItems"
                    value-key="value"
                    :disabled="departmentLocked"
                    class="w-full"
                  />
                  <p v-if="departmentLocked" class="text-xs text-gray-500 mt-1">{{ t('auditUniverse.modals.departmentLockedHint') }}</p>
                </UFormField>
                <UFormField :label="t('auditUniverse.modals.entityNameLabel')" class="space-y-2">
                  <UInput
                    v-model="renameNodeName"
                    :placeholder="t('auditUniverse.modals.entityNamePlaceholder')"
                    color="neutral"
                    class="w-full"
                    type="text"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('auditUniverse.modals.entityNameValidation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ renameNodeName ? renameNodeName.length : 0 }}/100
                  </div>
                </UFormField>
              </div>
              <template #footer>
                <div class="flex flex-col-reverse sm:flex-row justify-end gap-3">
                  <UButton
                    color="neutral"
                    variant="outline"
                    :label="t('auditUniverse.modals.cancel')"
                    class="w-full sm:w-auto"
                    @click="() => { renameModalOpen = false }" />
                  <UButton color="primary" :label="t('auditUniverse.modals.save')" class="w-full sm:w-auto" @click="saveRenameNode" />
                </div>
              </template>
            </UCard>
          </template>
        </UModal>

        <UModal 
          v-model:open="addCustomModalOpen"
          :ui="{
            content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
          }"  
        >
          <template #content>
            <UCard>
              <template #header>
                <h3 class="font-bold text-base text-slate-800 dark:text-slate-100">
                  {{ addCustomParentID ? t('auditUniverse.modals.addSubEntityTitle') : t('auditUniverse.modals.addCustomEntityTitle') }}
                </h3>
              </template>
              <div class="space-y-4">
                <UFormField :label="t('auditUniverse.modals.departmentLabel')" class="space-y-1">
                  <USelectMenu
                    v-model="addCustomDepartmentId"
                    :items="departmentItems"
                    value-key="value"
                    :disabled="departmentLocked"
                    class="w-full"
                  />
                  <p v-if="departmentLocked" class="text-xs text-gray-500 mt-1">{{ t('auditUniverse.modals.departmentLockedHint') }}</p>
                </UFormField>
                <UFormField :label="t('auditUniverse.modals.nameLabel')" class="space-y-1">
                  <UInput
                    v-model="addCustomNodeName"
                    :placeholder="t('auditUniverse.modals.namePlaceholder')"
                    color="neutral"
                    class="w-full"
                    type="text"
                    list="master-suggestions-list"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('auditUniverse.modals.entityNameValidation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <datalist id="master-suggestions-list">
                    <option v-for="item in activeSuggestions" :key="item" :value="item" />
                  </datalist>
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ addCustomNodeName ? addCustomNodeName.length : 0 }}/100
                  </div>
                </UFormField>

                <!-- Clickable suggestions chips from Master Data if available -->
                <div v-if="activeSuggestions.length > 0" class="space-y-1.5 pt-1">
                  <span class="text-xs font-semibold text-slate-500 dark:text-slate-400">
                    {{ t('auditUniverse.modals.masterSuggestions') }}
                  </span>
                  <div class="flex flex-wrap gap-1.5 max-h-28 overflow-y-auto pr-1">
                    <UBadge
                      v-for="s in activeSuggestions"
                      :key="s"
                      color="neutral"
                      variant="subtle"
                      class="cursor-pointer hover:bg-primary-50 hover:text-primary-600 dark:hover:bg-primary-950/40 transition-colors py-1 px-2.5 text-xs font-medium border border-slate-200 dark:border-slate-700"
                      @click="addCustomNodeName = s"
                    >
                      + {{ s }}
                    </UBadge>
                  </div>
                </div>
              </div>
              <template #footer>
                <div class="flex flex-col-reverse sm:flex-row justify-end gap-3">
                  <UButton
                    color="neutral"
                    variant="outline"
                    :label="t('auditUniverse.modals.cancel')"
                    class="w-full sm:w-auto"
                    @click="() => { addCustomModalOpen = false }" />
                  <UButton
                    color="primary"
                    :label="t('auditUniverse.modals.addNode')"
                    class="w-full sm:w-auto"
                    @click="saveAddCustomNode" />
                </div>
              </template>
            </UCard>
          </template>
        </UModal>
      </template>

      <!-- Tab 2: Establish Universe -->
      <template #establish>
        <div class="mt-6 space-y-6">
          <UCard class="shadow-sm border border-[var(--border-main)]">
            <template #header>
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                  <h2 class="text-lg font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                    {{ t('auditUniverse.establish.title', { year: selectedYear }) }}
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    {{ t('auditUniverse.establish.subtitle', { year: selectedYear }) }}
                  </p>
                </div>
                <div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">{{ t('auditUniverse.establish.targetYear') }}</span>
                    <USelect
                      v-model.number="selectedYear"
                      :items="fiscalYears"
                      size="sm"
                      color="neutral"
                      class="w-24"
                      @update:model-value="fetchYearlyUniverse"
                    />
                  </div>
                  <UButton v-if="canWriteAuditUniverse"
                    size="sm"
                    color="primary"
                    variant="solid"
                    icon="i-lucide-check-circle"
                    :label="t('auditUniverse.establish.establishBtn')"
                    class="w-full sm:w-auto"
                    @click="saveYearlyEstablishment"
                  />
                </div>
              </div>
            </template>

            <!-- Grid of Corporate Entities -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-h-[460px] overflow-y-auto pr-2">
              <div 
                v-for="node in corporateUniverse" 
                :key="node.id"
                class="border border-slate-100 dark:border-slate-800 rounded-xl p-4 bg-slate-50/20 dark:bg-slate-900/10 space-y-3"
              >
                <div class="flex items-center gap-2">
                  <UCheckbox
                    :model-value="isYearlySelected(node.id)"
                    @update:model-value="toggleYearlySelection(node.id)"
                  />
                  <span class="font-bold text-sm text-slate-800 dark:text-slate-200">
                    {{ node.name }}
                  </span>
                </div>

                <!-- Children / Sub-entities as checkboxes -->
                <div v-if="node.children && node.children.length > 0" class="pl-4 border-l border-slate-200 dark:border-slate-800 space-y-2">
                  <div v-for="sub in node.children" :key="sub.id" class="flex items-center gap-2">
                    <UCheckbox
                      :model-value="isYearlySelected(sub.id)"
                      @update:model-value="toggleYearlySelection(sub.id)"
                    />
                    <span class="text-md text-slate-500">{{ sub.name }}</span>
                  </div>
                </div>
              </div>
            </div>
          </UCard>
        </div>
      </template>
    </UTabs>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useRiskFactorsStore } from '~/stores/risk-factors'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useFiscalYear } from '~/composables/useFiscalYear'
import { useCompanyApi } from '~/composables/useCompanyApi'
import { useLocationApi } from '~/composables/useLocationApi'
import { useDepartmentApi } from '~/composables/useDepartmentApi'
import { useAuthStore } from '~/stores/auth'

const store = useAuditUniverseStore()
const riskFactorsStore = useRiskFactorsStore()
const route = useRoute()
const toast = useToastNotification()
const { t } = useI18n()
const authStore = useAuthStore()
const { canEditAuditUniverse, canWriteAuditUniverse, canWriteAuditUniverseFor } = useRbac()

// Departments for the selector (what the user can read) and the fixed department of a DEPARTMENT_HEAD.
const NO_DEPARTMENT = '__none__'
const departments = ref<{ id: string, department_name: string }[]>([])
const ownDepartmentId = computed(() => authStore.user?.departmentId || '')
// Only full writers choose freely; everyone else who can edit here is scoped to their own department.
const departmentLocked = computed(() => !canWriteAuditUniverse.value)
const departmentItems = computed(() => {
  const items = departments.value.map(d => ({ label: d.department_name, value: d.id }))
  if (departmentLocked.value) {
    const own = ownDepartmentId.value
    const label = items.find(i => i.value === own)?.label || authStore.user?.department || own
    return own ? [{ label, value: own }] : []
  }
  return [{ label: t('auditUniverse.modals.noDepartment'), value: NO_DEPARTMENT }, ...items]
})
const departmentLabel = (node: any): string => {
  if (node?.department_name) return node.department_name
  const id = node?.department_id
  return id ? (departments.value.find(d => d.id === id)?.department_name || '') : ''
}
// Value sent as department_id: DEPARTMENT_HEAD is forced to its own department.
const toDepartmentPayload = (selected: string): string | null => {
  if (departmentLocked.value) return ownDepartmentId.value || null
  return selected && selected !== NO_DEPARTMENT ? selected : null
}
const loadDepartments = async () => {
  try {
    const res = await useDepartmentApi().getDepartments({ page_size: 1000 })
    departments.value = res?.departments || []
  } catch (err) {
    console.warn('Could not load departments:', err)
  }
}

// DEPARTMENT_HEAD may edit only rows of its own department. Audit universe rows currently carry no department
// field, so for that role this stays false until the backend sends one (department_id / department).
const canEditNode = (node: any) => canWriteAuditUniverseFor(node)

const syncingMaster = ref(false)
const masterCompanies = ref<string[]>([])
const masterLocations = ref<string[]>([])

const tabItems = computed(() => [
  { value: 'library', slot: 'library', label: t('auditUniverse.tabs.library') },
  { value: 'establish', slot: 'establish', label: t('auditUniverse.tabs.establish') }
])

const activeTab = ref(
  route.query.tab === 'establish' ? 'establish' : 'library'
)

watch(() => route.query.tab, (newTab) => {
  if (newTab === 'priority') {
    navigateTo('/risk-profile/audit-priority')
  } else if (newTab === 'establish' || newTab === 'library') {
    activeTab.value = newTab as string
  }
})

// State
const { fiscalYears, selectedFiscalYear } = useFiscalYear()
const selectedYear = selectedFiscalYear
const alertMessage = ref('')
const alertType = ref('success')

watch(selectedYear, async () => {
  await fetchYearlyUniverse()
})

// Local state for Yearly selection checkboxes
const selectedYearlyIDs = ref<string[]>([])

// Modals
const renameModalOpen = ref(false)
const renameNode = ref<any>(null)
const renameNodeName = ref('')
const renameDepartmentId = ref(NO_DEPARTMENT)

const addCustomModalOpen = ref(false)
const addCustomParentID = ref<string | null>(null)
const addCustomNodeName = ref('')
const addCustomDepartmentId = ref(NO_DEPARTMENT)

// Lifecycle
onMounted(async () => {
  if (route.query.tab === 'priority') {
    navigateTo('/risk-profile/audit-priority')
    return
  }
  await store.fetchStandardUniverse()
  await store.fetchCorporateUniverse()
  await riskFactorsStore.fetchCorporateFactors()
  await fetchYearlyUniverse()
  await loadMasterData()
  await loadDepartments()
  if (departmentLocked.value && !authStore.user?.departmentId) {
    try { await authStore.fetchEmployeeLink() } catch (err) { console.warn('Could not link employee department:', err) }
  }
})

const loadMasterData = async () => {
  try {
    const compRes = await useCompanyApi().getCompanies({ page_size: 1000 })
    masterCompanies.value = (compRes?.companies || [])
      .filter((c: any) => c.is_active !== false)
      .map((c: any) => c.company_name || c.name)
      .filter(Boolean)
  } catch (err) {
    console.warn('Could not load master companies:', err)
  }

  try {
    const locRes = await useLocationApi().getLocations()
    masterLocations.value = (locRes || [])
      .filter((l: any) => l.is_active !== false)
      .map((l: any) => l.name)
      .filter(Boolean)
  } catch (err) {
    console.warn('Could not load master locations:', err)
  }
}

const activeSuggestions = computed(() => {
  if (!addCustomParentID.value) return []

  const findNode = (nodes: any[]): any => {
    for (const n of nodes) {
      if (n.id === addCustomParentID.value || n.standard_audit_universe_id === addCustomParentID.value) return n
      if (n.children) {
        const found = findNode(n.children)
        if (found) return found
      }
    }
    return null
  }
  const parent = findNode(store.corporateUniverse)
  if (!parent) return []

  const existingChildNames = (parent.children || []).map((c: any) => c.name.toLowerCase().trim())
  const parentName = (parent.name || '').toLowerCase()

  if (parentName.includes('subsidiaries') || parentName.includes('business unit')) {
    return masterCompanies.value.filter(c => !existingChildNames.includes(c.toLowerCase().trim()))
  }
  if (parentName.includes('branch') || parentName.includes('regional')) {
    return masterLocations.value.filter(l => !existingChildNames.includes(l.toLowerCase().trim()))
  }
  return []
})

const handleSyncMasterData = async () => {
  if (syncingMaster.value) return
  syncingMaster.value = true
  try {
    await loadMasterData()
    const success = await store.syncMasterData(masterCompanies.value, masterLocations.value)
    if (success) {
      toast.showSuccess(t('auditUniverse.messages.syncSuccess'))
    } else {
      toast.showError(store.errorMsg || t('auditUniverse.messages.syncFailed'))
    }
  } catch (err: any) {
    console.error('Failed to sync master data:', err)
    toast.showError(t('auditUniverse.messages.syncFailed'))
  } finally {
    syncingMaster.value = false
  }
}

const standardUniverse = computed(() => store.standardUniverse)
const corporateUniverse = computed(() => store.corporateUniverse)
const yearlyUniverse = computed(() => store.yearlyUniverse)



// Methods
const fetchYearlyUniverse = async () => {
  await store.fetchYearlyUniverse(selectedYear.value)
  // Populate active selection local checkbox states
  selectedYearlyIDs.value = store.yearlyUniverse.map(e => e.corporate_audit_universe_id)
}

// Check if standard node is selected
const isStandardNodeSelected = (stdID: string): boolean => {
  const checkNode = (nodes: any[]): boolean => {
    for (const n of nodes) {
      if (n.standard_audit_universe_id === stdID) return true
      if (n.children && checkNode(n.children)) return true
    }
    return false
  }
  return checkNode(store.corporateUniverse)
}

const toggleStandardSelection = async (stdNode: any) => {
  const alreadySelected = isStandardNodeSelected(stdNode.id)
  
  if (alreadySelected) {
    const findCorpNode = (nodes: any[]): any => {
      for (const n of nodes) {
        if (n.standard_audit_universe_id === stdNode.id) return n
        if (n.children) {
          const res = findCorpNode(n.children)
          if (res) return res
        }
      }
      return null
    }
    const match = findCorpNode(store.corporateUniverse)
    if (match) {
      await deleteCorporateNode(match.id)
    }
  } else {
    let corpParentID: string | undefined = undefined
    if (stdNode.parent_id) {
      const findCorpParent = (nodes: any[]): any => {
        for (const n of nodes) {
          if (n.standard_audit_universe_id === stdNode.parent_id) return n
          if (n.children) {
            const res = findCorpParent(n.children)
            if (res) return res
          }
        }
        return null
      }
      const parentMatch = findCorpParent(store.corporateUniverse)
      if (parentMatch) {
        corpParentID = parentMatch.id
      }
    }

    await store.saveCorporateNode({
      name: stdNode.name,
      standard_audit_universe_id: stdNode.id,
      parent_id: corpParentID
    })
  }
}

// Renaming
const openRenameModal = (node: any) => {
  renameNode.value = node
  renameNodeName.value = node.name
  renameDepartmentId.value = departmentLocked.value ? ownDepartmentId.value : (node.department_id || NO_DEPARTMENT)
  renameModalOpen.value = true
}

const saveRenameNode = async () => {
  if (!renameNodeName.value) return
  await store.saveCorporateNode({
    id: renameNode.value.id,
    name: renameNodeName.value,
    parent_id: renameNode.value.parent_id,
    standard_audit_universe_id: renameNode.value.standard_audit_universe_id,
    department_id: toDepartmentPayload(renameDepartmentId.value)
  })
  renameModalOpen.value = false
  toast.showSuccess(t('auditUniverse.messages.entityRenamed'))
}

// Adding custom nodes
const findCorporateNode = (nodes: any[], id: string): any => {
  for (const n of nodes) {
    if (n.id === id || n.standard_audit_universe_id === id) return n
    const found = n.children ? findCorporateNode(n.children, id) : null
    if (found) return found
  }
  return null
}
const openAddCustomModal = (parentID: string | null) => {
  addCustomParentID.value = parentID
  addCustomNodeName.value = ''
  const parent = parentID ? findCorporateNode(store.corporateUniverse, parentID) : null
  addCustomDepartmentId.value = departmentLocked.value ? ownDepartmentId.value : (parent?.department_id || NO_DEPARTMENT)
  addCustomModalOpen.value = true
}

const saveAddCustomNode = async () => {
  if (!addCustomNodeName.value) return

  let corpParentID: string | undefined = undefined
  if (addCustomParentID.value) {
    const findCorpParent = (nodes: any[]): any => {
      for (const n of nodes) {
        if (n.standard_audit_universe_id === addCustomParentID.value || n.id === addCustomParentID.value) return n
        if (n.children) {
          const res = findCorpParent(n.children)
          if (res) return res
        }
      }
      return null
    }
    const match = findCorpParent(store.corporateUniverse)
    corpParentID = match ? match.id : addCustomParentID.value
  }

  await store.saveCorporateNode({
    name: addCustomNodeName.value,
    parent_id: corpParentID,
    department_id: toDepartmentPayload(addCustomDepartmentId.value)
  })
  addCustomModalOpen.value = false
  toast.showSuccess(t('auditUniverse.messages.entityAdded'))
}

const deleteCorporateNode = async (id: string) => {
  if (!await useGlobalModalStore().confirmDelete({ description: t('auditUniverse.messages.deleteConfirmDesc') })) return
  await store.deleteCorporateNode(id)
  toast.showSuccess(t('auditUniverse.messages.entityDeleted'))
  await fetchYearlyUniverse()
}

// Yearly establishment
const isYearlySelected = (id: string): boolean => {
  return selectedYearlyIDs.value.includes(id)
}

const toggleYearlySelection = (id: string) => {
  const idx = selectedYearlyIDs.value.indexOf(id)
  if (idx >= 0) {
    selectedYearlyIDs.value.splice(idx, 1)
  } else {
    selectedYearlyIDs.value.push(id)
  }
}

const saveYearlyEstablishment = async () => {
  const success = await store.establishYearlyUniverse(selectedYear.value, selectedYearlyIDs.value)
  if (success) {
    toast.showSuccess(t('auditUniverse.messages.establishSuccess', { year: selectedYear.value }))
    await fetchYearlyUniverse()
  } else {
    toast.showError(store.errorMsg || t('auditUniverse.messages.establishFailed'))
  }
}



// Alert helper
const showAlert = (msg: string, type: string) => {
  alertMessage.value = msg
  alertType.value = type
  setTimeout(() => {
    alertMessage.value = ''
  }, 4000)
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
