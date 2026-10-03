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
                  <div class="flex items-center gap-2">
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
                    <span class="font-bold text-md text-slate-800 dark:text-slate-200">{{ node.name }}</span>
                    <div v-if="canEditAuditUniverse" class="flex items-center gap-1">
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
                      <span class="text-md text-slate-600 dark:text-slate-400">{{ sub.name }}</span>
                      <div class="flex items-center gap-1">
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
                <UFormField :label="t('auditUniverse.modals.nameLabel')" class="space-y-1">
                  <UInput
                    v-model="addCustomNodeName"
                    :placeholder="t('auditUniverse.modals.namePlaceholder')"
                    color="neutral"
                    class="w-full"
                    type="text"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('auditUniverse.modals.entityNameValidation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ addCustomNodeName ? addCustomNodeName.length : 0 }}/100
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
                      :items="[2025, 2026, 2027, 2028]"
                      size="sm"
                      color="neutral"
                      class="w-24"
                      @update:model-value="fetchYearlyUniverse"
                    />
                  </div>
                  <UButton
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

const store = useAuditUniverseStore()
const riskFactorsStore = useRiskFactorsStore()
const route = useRoute()
const toast = useToastNotification()
const { t } = useI18n()
const { canEditAuditUniverse } = useRbac()

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
const selectedYear = ref(2026)
const alertMessage = ref('')
const alertType = ref('success')

// Local state for Yearly selection checkboxes
const selectedYearlyIDs = ref<string[]>([])

// Modals
const renameModalOpen = ref(false)
const renameNode = ref<any>(null)
const renameNodeName = ref('')

const addCustomModalOpen = ref(false)
const addCustomParentID = ref<string | null>(null)
const addCustomNodeName = ref('')

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
})

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
  renameModalOpen.value = true
}

const saveRenameNode = async () => {
  if (!renameNodeName.value) return
  await store.saveCorporateNode({
    id: renameNode.value.id,
    name: renameNodeName.value,
    parent_id: renameNode.value.parent_id,
    standard_audit_universe_id: renameNode.value.standard_audit_universe_id
  })
  renameModalOpen.value = false
  toast.showSuccess(t('auditUniverse.messages.entityRenamed'))
}

// Adding custom nodes
const openAddCustomModal = (parentID: string | null) => {
  addCustomParentID.value = parentID
  addCustomNodeName.value = ''
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
    if (match) {
      corpParentID = match.id
    }
  }

  await store.saveCorporateNode({
    name: addCustomNodeName.value,
    parent_id: corpParentID
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
