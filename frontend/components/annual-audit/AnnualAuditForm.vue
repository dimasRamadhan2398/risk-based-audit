<template>
  <UModal
    v-model:open="store.showModal"
    dismissible
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <UForm
        :state="store.form"
        @submit.prevent="store.handleSubmit"
      >
        <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh] transition-colors duration-300">
          <div class="px-6 py-4 rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              Annual Audit Form
            </h3>
            <UIcon
              name="close"
              class="text-2xl cursor-pointer"
              @click="store.closeModal"
            />
          </div>

          <div class="p-6 overflow-y-auto space-y-4">
            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                1. Activity Detail
              </h4>
              <div
                v-if="store.validationErrors.activityDetail"
                class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
              >
                <UIcon
                  name="i-lucide-circle-alert"
                  class="mt-0.5 shrink-0"
                />

                <span>
                  {{ store.validationErrors.activityDetail }}
                </span>
              </div>

              <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-4">
                <UFormField
                  label="Status"
                  size="lg"
                >
                  <USelectMenu
                    v-model="store.form.status"
                    :items="statusOptions"
                    placeholder="Select Status"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  label="Activity Code"
                  size="lg"
                  help="Format Auto: PKAT-Tahun-CodeCategory-No.urut"
                >
                  <UInput
                    v-model="store.form.code"
                    required
                    type="text"
                    placeholder="e.g. PKAT-2026-ASR-001"
                    class="w-full font-mono text-sm"
                  />
                </UFormField>
              </div>

              <div class="space-y-4 mt-6">
                <UCard
                  v-for="(activity, index) in store.form.activities"
                  :key="index"
                  class="relative"
                >
                  <div class="flex justify-between items-center mb-4 border-b border-gray-200  pb-2">
                    <h5 class="font-bold text-gray-700 ">
                      Sub-Aktivitas {{ index + 1 }}
                    </h5>
                    <UButton
                      v-if="store.form.activities.length > 1"
                      icon="i-heroicons-trash"
                      color="error"
                      variant="ghost"
                      size="sm"
                      @click="store.removeActivity(index)"
                    />
                  </div>

                  <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
                    <UFormField
                      label="Activity Name"
                      size="lg"
                    >
                      <UInput
                        v-model="activity.name"
                        type="text"
                        placeholder="e.g. Audit Operasional Div. Keuangan"
                        class="w-full"
                        required
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('Nama aktivitas maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ activity.name ? activity.name.length : 0 }}/100
                      </div>
                    </UFormField>

                    <UFormField
                      label="Category"
                      size="lg"
                    >
                      <USelectMenu
                        v-model="(activity.category as any)"
                        :items="categoryOptions"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField
                      label="Involved Departments"
                      size="lg"
                    >
                      <USelectMenu
                        v-model="activity.involvedDepartments"
                        :items="Object.values(AuditDepartment)"
                        multiple
                        placeholder="Select Departments"
                        class="w-full"
                        @update:model-value="() => { if (activity.involvedDepartments && activity.involvedDepartments.length > 0) activity.department = activity.involvedDepartments[0]; }"
                      />
                    </UFormField>
                  </div>

                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mt-4">
                    <UFormField
                      label="Auditable Entity (from Audit Universe)"
                      size="lg"
                    >
                      <USelectMenu
                        :model-value="auditableUniverseOptions.find(u => u.name === activity.riskName || u.label === activity.riskName) || (activity.riskName ? { label: activity.riskName, name: activity.riskName, riskLevel: activity.riskLevel || 'Custom' } : undefined)"
                        :items="auditableUniverseOptions"
                        label-key="label"
                        placeholder="-- Select Auditable Entity --"
                        class="w-full"
                        :loading="auditUniverseStore.loading"
                        @update:model-value="(val: any) => handleSelectAuditableEntity(activity, val)"
                      >
                        <template #item="{ item }">
                          <div class="flex items-center gap-2 max-w-full w-full py-0.5">
                            <span
                              class="w-2.5 h-2.5 rounded-full shrink-0"
                              :style="{ backgroundColor: getRiskLevelColorHex(item.riskLevel) }"
                            />
                            <span class="text-[10px] font-bold text-gray-500 shrink-0">[{{ item.riskLevel }}]</span>
                            <span class="truncate text-sm font-medium">{{ item.label || item.name }}</span>
                            <span
                              v-if="item.riskIndex > 0"
                              class="text-[10px] text-gray-400 shrink-0 ml-1"
                            >({{ (item.riskIndex).toFixed(1) }}%)</span>
                            <UBadge
                              v-if="item.auditPriority"
                              color="error"
                              size="sm"
                              variant="subtle"
                              class="ml-auto"
                            >
                              Priority
                            </UBadge>
                          </div>
                        </template>
                      </USelectMenu>
                    </UFormField>

                    <UFormField
                      label="Calculated Risk Level"
                      size="lg"
                    >
                      <div class="mt-2 flex items-center gap-2">
                        <UBadge
                          v-if="activity.riskLevel"
                          :color="getRiskLevelColor(activity.riskLevel)"
                          size="lg"
                          variant="solid"
                        >
                          {{ activity.riskLevel }}
                        </UBadge>
                        <UBadge
                          v-if="auditableUniverseOptions.find(u => u.name === activity.riskName)?.auditPriority"
                          color="error"
                          size="md"
                          variant="subtle"
                        >
                          Audit Priority
                        </UBadge>
                        <span
                          v-if="!activity.riskLevel"
                          class="text-gray-400 text-sm"
                        >-</span>
                      </div>
                    </UFormField>
                  </div>
                </UCard>
              </div>

              <div class="flex justify-start pt-2">
                <UButton
                  label="Tambah Aktivitas"
                  icon="i-heroicons-plus"
                  color="primary"
                  variant="soft"
                  @click="store.addActivity"
                />
              </div>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">
                2. Timeline
              </h4>
              <div
                v-if="store.validationErrors.timeline"
                class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
              >
                <UIcon
                  name="i-lucide-circle-alert"
                  class="mt-0.5 shrink-0"
                />

                <span>
                  {{ store.validationErrors.timeline }}
                </span>
              </div>

              <UFormField
                label="Year"
                size="lg"
              >
                <USelectMenu
                  v-model="store.form.year"
                  :items="store.yearOptions"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                label="Select Months"
                size="lg"
              >
                <div class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-12 gap-2 p-4">
                  <div
                    v-for="(month, idx) in store.monthsList"
                    :key="idx"
                    class="cursor-pointer border rounded-lg p-2 text-center text-md font-semibold transition select-none"
                    :class="store.form.selectedMonths.includes(idx) ? 'border-secondary-600 shadow-md transform scale-105' : 'text-gray-500'"
                    @click="store.toggleMonth(idx)"
                  >
                    {{ month }}
                  </div>
                </div>

                <div
                  v-if="store.scheduleWarning"
                  class="flex items-center gap-2 text-warning-600 p-4 m-4 rounded-lg text-sm border border-warning-200"
                >
                  <UIcon
                    name="warning"
                    class=" text-warning-500"
                  />
                  {{ store.scheduleWarning }}
                </div>
                <div
                  v-if="store.quarterAlert"
                  class="flex items-center gap-2 text-error-600 p-4 m-4 rounded-lg text-sm border border-error-200"
                >
                  <UIcon
                    name="alert"
                    class=" text-error-500"
                  />
                  {{ store.quarterAlert }}
                </div>

                <div class="text-sm text-gray-600 p-4 m-4 rounded border">
                  <span class="font-bold">Distribusi Triwulan:</span>
                  {{ store.computedQuarters.length ? store.computedQuarters.join(', ') : '-' }}
                </div>
              </UFormField>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">
                3. Auditor
              </h4>
              <div
                v-if="store.validationErrors.auditor"
                class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
              >
                <UIcon
                  name="i-lucide-circle-alert"
                  class="mt-0.5 shrink-0"
                />

                <span>
                  {{ store.validationErrors.auditor }}
                </span>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-3 gap-6 p-4 rounded-lg border border-primary-200 ">
                <UFormField
                  label="Number of Auditors (1-10)"
                  size="lg"
                >
                  <UInput
                    v-model.number="store.form.auditorCount"
                    type="number"
                    min="1"
                    max="10"
                    class="w-full"
                  />
                  <p class="text-md text-gray-500 mt-1">
                    Advice: High Risk min. 3 auditor
                  </p>
                </UFormField>

                <UFormField
                  label="Duration (Days)"
                  size="lg"
                >
                  <UInput
                    v-model.number="store.form.daysPerAuditor"
                    type="number"
                    min="1"
                  />
                </UFormField>

                <UBadge
                  variant="outline"
                  color="primary"
                  class="p-3 rounded-lg text-center flex flex-col justify-center bg-transparent border border-primary-500/50 dark:border-primary-400/50"
                >
                  <span class="text-md text-gray-500 dark:text-gray-400 uppercase">Total Mandays</span>
                  <span class="text-2xl font-bold text-primary-600 dark:text-primary-400">{{ store.totalMandays }}</span>
                  <span class="text-md text-gray-400 dark:text-gray-500">= {{ store.form.auditorCount }} person × {{ store.form.daysPerAuditor }} day</span>
                </UBadge>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <UFormField
                  label="Supervisor"
                  size="lg"
                >
                  <USelectMenu
                    :model-value="(store.form.supervisorId as any)"
                    :items="store.supervisorOptions"
                    value-key="id"
                    option-key="label"
                    placeholder="-- Choose Supervisor --"
                    class="w-full rounded-md shadow-sm"
                    @update:model-value="(val: any) => store.form.supervisorId = val"
                  />

                  <p
                    v-if="Number(store.selectedSupervisor?.workload || 0) > 6"
                    class="text-md text-error-500 mt-1 font-bold"
                  >
                    ⚠️ Warning: This Supervisor is supervised > 6 Activity!
                  </p>
                </UFormField>

                <UFormField>
                  <div class="flex justify-between items-center mb-1">
                    <label class="label mb-0">Team Estimation Capacity</label>
                    <span
                      class="text-md font-bold"
                      :class="store.utilizationData.color === 'red' ? 'text-error-600' : 'text-success-600'"
                    >
                      {{ store.utilizationData.msg }}
                    </span>
                  </div>
                  <div class="w-full rounded-full h-2.5 ">
                    <div
                      class="h-2.5 rounded-full transition-all duration-500"
                      :class="{
                        'bg-success-500': store.utilizationData.color === 'green',
                        'bg-warning-400': store.utilizationData.color === 'yellow',
                        'bg-error-600': store.utilizationData.color === 'red'
                      }"
                      :style="{ width: `${Math.min(store.utilizationData.percent, 100)}%` }"
                    />
                  </div>
                  <p class="text-md text-gray-400 mt-1">
                    Total Load: {{ store.utilizationData.percent.toFixed(1) }}% from Annual Capacity.
                  </p>
                </UFormField>
              </div>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">
                4. Notes
              </h4>
              <UFormField
                label="Additional Note (Optional)"
                size="lg"
              >
                <UTextarea
                  v-model="store.form.notes"
                  maxlength="500"
                  :rows="5"
                  placeholder="Example: High Priority - external recommendation..."
                  class="w-full"
                  autoresize
                />
                <div class="flex justify-between mt-1">
                  <span class="text-md text-gray-400">Use it for Special Audit / Investigation.</span>
                  <span class="text-md text-gray-400">{{ (store.form.notes || '').length }}/500</span>
                </div>
              </UFormField>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">
                5. Attachment
              </h4>
              <UFormField label="Attachment Category">
                <USelectMenu
                  v-model="store.form.attachmentCategory"
                  :items="store.attachmentCategoryOptions"
                  value-key="id"
                  option-key="label"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                label="Attachment Uploaded By"
                class="block text-sm font-medium"
                size="lg"
              >
                <UInput
                  v-model="store.form.attachmentUploadedBy"
                  placeholder="Attachment Uploaded By"
                  class="w-full"
                />
              </UFormField>
              <UFormField label="Attachment Upload Date">
                <AppDatePicker
                  v-model="store.form.attachmentUploadDate"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                label="Upload your Attachment here"
                size="lg"
              >
                <UFileUpload
                  v-model="store.form.file"
                  layout="list"
                  multiple
                  label="Drop your attachments here"
                  description="You can upload multiple files (max. 2MB each)"
                  class="w-full"
                  :ui="{
                    base: 'min-h-48'
                  }"
                />
              </UFormField>
            </div>
          </div>

          <div class="px-6 py-4 rounded-b-xl flex justify-end gap-3">
            <UButton
              type="submit"
              :label="store.isEditing ? 'Update Plan' : 'Save Plan'"
              color="primary"
              :disabled="!!store.quarterAlert || store.utilizationData.color === 'red'"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { AnnualAuditPlanStatus, AuditCategory, AuditDepartment } from '~/types/audit'

const store = useAnnualPlanStore()
const riskStore = useRiskProfileStore()
const auditUniverseStore = useAuditUniverseStore()

const statusOptions = Object.values(AnnualAuditPlanStatus) as AnnualAuditPlanStatus[]
const categoryOptions = Object.values(AuditCategory) as AuditCategory[]

riskStore.fetchRisks()

const loadAuditUniverseData = async () => {
  const currentYear = Number(store.form.year) || new Date().getFullYear()
  await Promise.allSettled([
    auditUniverseStore.fetchYearlyUniverse(currentYear),
    auditUniverseStore.fetchCorporateUniverse()
  ])
}

onMounted(() => {
  loadAuditUniverseData()
})

watch(() => store.showModal, (isOpen) => {
  if (isOpen) {
    loadAuditUniverseData()
  }
})

watch(() => store.form.year, async (newYear) => {
  if (newYear) {
    await auditUniverseStore.fetchYearlyUniverse(Number(newYear))
  }
})

interface AuditableOption {
  id?: string
  name: string
  label: string
  value: string
  riskLevel: string
  riskIndex: number
  auditPriority: boolean
  source: string
}

const auditableUniverseOptions = computed<AuditableOption[]>(() => {
  const items: AuditableOption[] = []
  const seenNames = new Set<string>()

  // 1. First priority: Yearly established universe for the selected plan year
  if (auditUniverseStore.yearlyUniverse && auditUniverseStore.yearlyUniverse.length > 0) {
    for (const item of auditUniverseStore.yearlyUniverse) {
      const name = item.corporate_audit_universe?.name
      if (!name || seenNames.has(name.toLowerCase())) continue
      seenNames.add(name.toLowerCase())

      items.push({
        id: item.id,
        name,
        label: name,
        value: name,
        riskLevel: item.risk_level || 'Not Scored',
        riskIndex: Number(item.risk_index) || 0,
        auditPriority: !!item.audit_priority,
        source: 'yearly'
      })
    }
  }

  // 2. Secondary: Corporate Audit Universe library (parent entities and sub-entities)
  if (auditUniverseStore.corporateUniverse && auditUniverseStore.corporateUniverse.length > 0) {
    const traverse = (node: { id?: string, name?: string, children?: unknown[] }, parentName?: string) => {
      if (!node || !node.name) return
      const lowerName = node.name.toLowerCase()
      if (!seenNames.has(lowerName)) {
        seenNames.add(lowerName)
        items.push({
          id: node.id,
          name: node.name,
          label: parentName ? `${node.name} (${parentName})` : node.name,
          value: node.name,
          riskLevel: 'Not Scored',
          riskIndex: 0,
          auditPriority: false,
          source: 'corporate'
        })
      }
      if (Array.isArray(node.children)) {
        for (const child of node.children) {
          traverse(child as { id?: string, name?: string, children?: unknown[] }, node.name)
        }
      }
    }

    for (const node of auditUniverseStore.corporateUniverse) {
      traverse(node)
    }
  }

  return items
})

const handleSelectAuditableEntity = (activity: { name?: string, riskName?: string, riskLevel?: string, department?: AuditDepartment, involvedDepartments?: AuditDepartment[] }, val: unknown) => {
  if (val) {
    const selectedName = typeof val === 'object' && val !== null ? ((val as { name?: string, label?: string }).name || (val as { name?: string, label?: string }).label || '') : String(val)
    const match = auditableUniverseOptions.value.find(u => u.name === selectedName || u.label === selectedName) || (typeof val === 'object' && val !== null ? val as AuditableOption : null)
    activity.riskName = match?.name || selectedName
    activity.riskLevel = match?.riskLevel || 'Not Scored'

    // Suggest default activity title if empty or using standard default prefix
    if (!activity.name || activity.name.startsWith('Audit Operasional')) {
      activity.name = `Audit Operasional ${activity.riskName}`
    }

    // Auto-detect Department based on entity name
    const nameLower = (activity.riskName || '').toLowerCase()
    let detectedDept = AuditDepartment.OPS
    if (nameLower.includes('it') || nameLower.includes('technology') || nameLower.includes('sistem') || nameLower.includes('cyber') || nameLower.includes('data') || nameLower.includes('security')) {
      detectedDept = AuditDepartment.IT
    } else if (nameLower.includes('finance') || nameLower.includes('keuangan') || nameLower.includes('treasury') || nameLower.includes('tax') || nameLower.includes('pajak') || nameLower.includes('accounting') || nameLower.includes('akuntansi')) {
      detectedDept = AuditDepartment.FINANCE
    } else if (nameLower.includes('hr') || nameLower.includes('human') || nameLower.includes('sumber daya') || nameLower.includes('resource') || nameLower.includes('talent') || nameLower.includes('personalia')) {
      detectedDept = AuditDepartment.HR
    }
    activity.department = detectedDept

    if (!activity.involvedDepartments || activity.involvedDepartments.length === 0) {
      activity.involvedDepartments = [detectedDept]
    }
  } else {
    activity.riskName = ''
    activity.riskLevel = ''
  }
}

const _getFilteredRisksForDept = (dept: string) => {
  if (!riskStore.risks || riskStore.risks.length === 0) {
    return []
  }
  return riskStore.risks.filter((r) => {
    if (dept === 'IT') return r.category === 'Technology'
    if (dept === 'Finance') return r.category === 'Financial'
    if (dept === 'HR') return r.category === 'Human Resources'
    if (dept === 'Ops') return ['Operations', 'Compliance', 'Strategic', 'Governance'].includes(r.category)
    return true
  })
}

const getRiskLevelColorHex = (level?: string) => {
  if (!level) return '#9E9E9E'
  const lvl = level.toLowerCase()
  if (lvl.includes('moderate to high') || lvl.includes('medium to high')) return '#FF9800'
  if (lvl.includes('low to moderate') || lvl.includes('low to medium')) return '#8BC34A'
  if (lvl.includes('high')) return '#F44336'
  if (lvl.includes('moderate') || lvl.includes('medium')) return '#FFC107'
  if (lvl.includes('low')) return '#4CAF50'
  return '#9E9E9E'
}

const getRiskLevelColor = (level?: string) => {
  if (!level) return 'neutral'
  const lvl = level.toLowerCase()
  if (lvl.includes('moderate to high') || lvl.includes('medium to high')) return 'warning'
  if (lvl.includes('high')) return 'error'
  if (lvl.includes('mod') || lvl.includes('medium')) return 'warning'
  if (lvl.includes('low')) return 'success'
  return 'neutral'
}
</script>
