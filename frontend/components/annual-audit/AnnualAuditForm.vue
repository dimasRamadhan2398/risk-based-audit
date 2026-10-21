<template>
  <UModal
    v-model:open="store.showModal"
    dismissible
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
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
          <div class="px-4 sm:px-6 py-4 rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ t('annualAudit.form.title') }}
            </h3>
            <UIcon
              name="close"
              class="text-2xl cursor-pointer"
              @click="store.closeModal"
            />
          </div>

          <div class="p-4 sm:p-6 overflow-y-auto space-y-4">
            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('annualAudit.form.sectionPlanHeader') }}
              </h4>

              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <UFormField
                  :label="t('annualAudit.form.year')"
                  size="lg"
                  required
                >
                  <USelectMenu
                    v-model="store.form.year"
                    :items="store.yearOptions"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.planTitle')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <UInput
                    v-model="store.form.title"
                    type="text"
                    :placeholder="t('annualAudit.form.planTitlePlaceholder')"
                    maxlength="150"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.status')"
                  size="lg"
                >
                  <USelectMenu
                    v-model="store.form.status"
                    :items="statusOptions"
                    :placeholder="t('annualAudit.form.statusPlaceholder')"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.activityCode')"
                  size="lg"
                  :help="t('annualAudit.form.activityCodeHelp')"
                >
                  <UInput
                    :model-value="activityCodePreview"
                    disabled
                    aria-readonly="true"
                    type="text"
                    :placeholder="t('annualAudit.form.activityCodePending')"
                    class="w-full font-mono text-sm bg-gray-100 dark:bg-gray-800"
                  />
                </UFormField>
              </div>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">
                {{ t('annualAudit.form.sectionActivity') }}
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

              <div class="space-y-4">
                <UCard
                  v-for="(activity, index) in store.form.activities"
                  :key="activity.id || index"
                  class="relative"
                >
                  <div class="flex justify-between items-center mb-4 border-b border-gray-200  pb-2">
                    <h5 class="font-bold text-gray-700 ">
                      {{ t('annualAudit.form.subActivity', { number: index + 1 }) }}
                    </h5>
                    <UTooltip
                      v-if="store.form.activities.length > 1"
                      :text="activity.activityPlanRef ? t('annualAudit.form.rowInUseTooltip') : t('common.delete')"
                    >
                      <UButton
                        icon="i-heroicons-trash"
                        color="error"
                        variant="ghost"
                        size="sm"
                        :disabled="!!activity.activityPlanRef"
                        :aria-label="t('common.delete')"
                        @click="store.removeActivity(index)"
                      />
                    </UTooltip>
                  </div>

                  <!-- The Audit Activity Plan that uses this row: remove it there before removing it here. -->
                  <div
                    v-if="activity.activityPlanRef"
                    class="mb-4 flex items-start gap-2 rounded-lg border border-primary-200 dark:border-primary-800 bg-primary-50 dark:bg-primary-950/40 p-2 text-sm text-primary-700 dark:text-primary-300"
                    data-testid="used-in-activity-plan"
                  >
                    <UIcon
                      name="i-heroicons-link"
                      class="mt-0.5 shrink-0"
                    />
                    <span>{{ t('annualAudit.form.usedInActivityPlan', { title: activity.activityPlanRef.activityPlanTitle || '-', code: activity.activityPlanRef.assignmentCode || '-' }) }}</span>
                  </div>

                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <UFormField
                      :label="t('annualAudit.form.objectName')"
                      size="lg"
                    >
                      <UInput
                        v-model="activity.name"
                        type="text"
                        :placeholder="t('annualAudit.form.activityNamePlaceholder')"
                        class="w-full"
                        required
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity(t('annualAudit.form.activityNameInvalid'))"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-md text-gray-500 mt-1 text-right">
                        {{ activity.name ? activity.name.length : 0 }}/100
                      </div>
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.auditType')"
                      size="lg"
                    >
                      <USelectMenu
                        v-model="(activity.category as any)"
                        :items="categoryOptions"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.department')"
                      size="lg"
                    >
                      <USelectMenu
                        v-model="activity.departmentId"
                        :items="store.departments"
                        value-key="id"
                        label-key="name"
                        :loading="store.departmentOptionsLoading"
                        :placeholder="t('annualAudit.form.departmentsPlaceholder')"
                        class="w-full"
                      >
                        <template #empty>
                          {{ store.departmentOptionsLoading ? t('common.loading') : t('annualAudit.form.departmentsEmpty') }}
                        </template>
                      </USelectMenu>
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.auditableEntityLabel')"
                      size="lg"
                      class="md:col-span-2"
                    >
                      <USelectMenu
                        :model-value="auditableEntityValue(activity)"
                        :items="auditableUniverseOptions"
                        label-key="label"
                        by="value"
                        :placeholder="t('annualAudit.form.auditableEntityPlaceholder')"
                        class="w-full"
                        :loading="auditUniverseStore.loading"
                        @update:model-value="(val: any) => handleSelectAuditableEntity(activity, val)"
                      >
                        <template #empty>
                          {{ auditUniverseStore.loading ? t('common.loading') : t('annualAudit.form.auditableEntityEmpty') }}
                        </template>
                        <template #item="{ item }">
                          <div class="flex items-center gap-2 max-w-full w-full py-0.5">
                            <span
                              class="w-2.5 h-2.5 rounded-full shrink-0"
                              :class="getRiskLevelDotClass(item.riskLevel)"
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
                              {{ t('annualAudit.form.priorityBadge') }}
                            </UBadge>
                          </div>
                        </template>
                      </USelectMenu>
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.calculatedRiskLevel')"
                      size="lg"
                      :help="t('annualAudit.form.riskSnapshotHelp')"
                    >
                      <div class="mt-2 flex flex-wrap items-center gap-2">
                        <span class="text-sm font-medium text-gray-700 dark:text-gray-200">
                          {{ activity.riskName || '-' }}
                        </span>
                        <span
                          v-if="activity.riskLevel"
                          class="inline-flex items-center px-2 py-1 rounded-md text-sm font-semibold"
                          :class="getRiskLevelColorClass(activity.riskLevel)"
                        >
                          {{ activity.riskLevel }}
                        </span>
                        <UBadge
                          v-if="auditableUniverseOptions.find(u => u.id === activity.auditUniverseId)?.auditPriority"
                          color="error"
                          size="md"
                          variant="subtle"
                        >
                          {{ t('annualAudit.form.auditPriorityBadge') }}
                        </UBadge>
                      </div>
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.scheduleText')"
                      size="lg"
                    >
                      <UInput
                        v-model="activity.timelineText"
                        type="text"
                        :placeholder="t('annualAudit.form.scheduleTextPlaceholder')"
                        maxlength="150"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.teamCount')"
                      size="lg"
                    >
                      <UInput
                        v-model.number="activity.auditorCount"
                        type="number"
                        min="1"
                        max="10"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.totalMandays')"
                      size="lg"
                    >
                      <UInput
                        v-model.number="activity.totalMandays"
                        type="number"
                        min="1"
                        class="w-full"
                      />
                    </UFormField>

                    <UFormField
                      :label="t('annualAudit.form.teamLead')"
                      size="lg"
                      class="md:col-span-2"
                    >
                      <UInput
                        v-model="activity.supervisorName"
                        type="text"
                        :placeholder="t('annualAudit.form.supervisorPlaceholder')"
                        class="w-full"
                      />
                    </UFormField>
                  </div>
                </UCard>
              </div>

              <div class="flex justify-start pt-2">
                <UButton
                  :label="t('annualAudit.form.addActivity')"
                  icon="i-heroicons-plus"
                  color="primary"
                  variant="soft"
                  @click="store.addActivity"
                />
              </div>

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

              <div class="grid grid-cols-1 md:grid-cols-2 gap-6 p-4 rounded-lg border border-primary-200">
                <UBadge
                  variant="outline"
                  color="primary"
                  class="p-3 rounded-lg text-center flex flex-col justify-center bg-transparent border border-primary-500/50 dark:border-primary-400/50"
                >
                  <span class="text-md text-gray-500 dark:text-gray-400 uppercase">{{ t('annualAudit.form.totalMandays') }}</span>
                  <span class="text-2xl font-bold text-primary-600 dark:text-primary-400">{{ store.totalMandays }}</span>
                  <span class="text-md text-gray-400 dark:text-gray-500">{{ t('annualAudit.form.mandaysSumHint') }}</span>
                </UBadge>

                <UFormField>
                  <div class="flex justify-between items-center mb-1">
                    <label class="label mb-0">{{ t('annualAudit.form.teamCapacity') }}</label>
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
                    {{ t('annualAudit.form.totalLoad', { percent: store.utilizationData.percent.toFixed(1) }) }}
                  </p>
                </UFormField>
              </div>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">
                {{ t('annualAudit.form.sectionTimeline') }}
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
                :label="t('annualAudit.form.selectMonths')"
                size="lg"
                required
              >
                <div
                  class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-12 gap-2 p-4"
                  role="group"
                  :aria-label="t('annualAudit.form.selectMonths')"
                  aria-required="true"
                >
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

                <div class="text-sm text-gray-600 dark:text-white p-4 m-4 rounded border">
                  <span class="font-bold">{{ t('annualAudit.form.quarterDistribution') }}</span>
                  {{ store.computedQuarters.length ? store.computedQuarters.join(', ') : '-' }}
                </div>
              </UFormField>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">
                {{ t('annualAudit.form.sectionNotes') }}
              </h4>
              <UFormField
                :label="t('annualAudit.form.additionalNote')"
                size="lg"
              >
                <UTextarea
                  v-model="store.form.notes"
                  maxlength="500"
                  :rows="5"
                  :placeholder="t('annualAudit.form.notesPlaceholder')"
                  class="w-full"
                  autoresize
                />
                <div class="flex justify-between mt-1">
                  <span class="text-md text-gray-400">{{ t('annualAudit.form.notesHint') }}</span>
                  <span class="text-md text-gray-400">{{ (store.form.notes || '').length }}/500</span>
                </div>
              </UFormField>
            </div>

            <div class="space-y-4">
              <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">
                {{ t('annualAudit.form.sectionAttachment') }}
              </h4>
              <UFormField :label="t('annualAudit.form.attachmentCategory')">
                <USelectMenu
                  v-model="store.form.attachmentCategory"
                  :items="store.attachmentCategoryOptions"
                  value-key="id"
                  option-key="label"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                :label="t('annualAudit.form.attachmentUploadedBy')"
                class="block text-sm font-medium"
                size="lg"
              >
                <UInput
                  v-model="store.form.attachmentUploadedBy"
                  :placeholder="t('annualAudit.form.attachmentUploadedBy')"
                  class="w-full bg-gray-100 dark:bg-gray-800"
                  disabled
                  aria-readonly="true"
                />
              </UFormField>
              <UFormField :label="t('annualAudit.form.attachmentUploadDate')">
                <AppDatePicker
                  v-model="store.form.attachmentUploadDate"
                  class="w-full"
                />
              </UFormField>
              <UFormField
                :label="t('annualAudit.form.uploadLabel')"
                size="lg"
              >
                <UFileUpload
                  v-model="store.form.file"
                  layout="list"
                  multiple
                  :label="t('annualAudit.form.dropLabel')"
                  :description="t('annualAudit.form.uploadDescription')"
                  class="w-full"
                  :ui="{
                    base: 'min-h-48'
                  }"
                />
              </UFormField>
            </div>
          </div>

          <div class="px-4 sm:px-6 py-4 rounded-b-xl flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3">
            <UButton
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              :label="t('common.cancel')"
              @click="() => store.closeModal()"
            />
            <UButton
              type="submit"
              :label="store.isEditing ? t('annualAudit.form.updatePlan') : t('annualAudit.form.savePlan')"
              color="primary"
              class="w-full sm:w-auto font-bold"
              :disabled="store.utilizationData.color === 'red'"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, watch, onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useAuthStore } from '~/stores/auth'
import { AnnualAuditPlanStatus, AuditCategory, AuditDepartment, type AuditActivities } from '~/types/audit'
import { getRiskLevelColorClass, getRiskLevelDotClass } from '~/utils/riskLevelBadge'
import {
  buildAuditableEntityOptions,
  findAuditableEntityForActivity,
  NOT_SCORED_RISK_LEVEL,
  resolveAuditableEntitySelection,
  type AuditableEntityOption
} from '~/utils/auditableEntityOptions'

const store = useAnnualPlanStore()
const riskStore = useRiskProfileStore()
const auditUniverseStore = useAuditUniverseStore()
const authStore = useAuthStore()
const { t } = useI18n()

const roleLabels: Record<string, string> = {
  ADMIN: 'System Administrator',
  AUDITOR: 'Auditor',
  DEPARTMENT_HEAD: 'Department Head',
  AUDITEE: 'Auditee',
  VIEWER: 'Viewer',
  EXECUTIVE: 'Chief Audit Executive'
}

const signedInRoleLabel = computed(() => {
  const role = authStore.user?.roles?.[0]
  if (!role) return ''
  return roleLabels[role] || role
    .split('_')
    .filter(Boolean)
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
})

// New plans get their number from the server on save, so the field shows the pattern for the selected year
const activityCodePreview = computed(() => {
  if (store.isEditing) return store.form.code || ''
  return `PKAT-${store.form.year || new Date().getFullYear()}-XXX`
})

const statusOptions = Object.values(AnnualAuditPlanStatus) as AnnualAuditPlanStatus[]
const categoryOptions = Object.values(AuditCategory) as AuditCategory[]

riskStore.fetchRisks()

const loadFormOptions = async () => {
  const currentYear = Number(store.form.year) || new Date().getFullYear()
  await Promise.allSettled([
    auditUniverseStore.fetchYearlyUniverse(currentYear),
    auditUniverseStore.fetchCorporateUniverse(),
    store.fetchDepartmentOptions()
  ])
}

onMounted(() => {
  loadFormOptions()
})

watch([() => store.showModal, signedInRoleLabel], ([isOpen]) => {
  if (isOpen) {
    store.form.attachmentUploadedBy = signedInRoleLabel.value
    loadFormOptions()
  }
})

watch(() => store.form.year, async (newYear) => {
  if (newYear) {
    await auditUniverseStore.fetchYearlyUniverse(Number(newYear))
  }
})

const auditableUniverseOptions = computed<AuditableEntityOption[]>(() =>
  buildAuditableEntityOptions(auditUniverseStore.yearlyUniverse, auditUniverseStore.corporateUniverse)
)

// Pre-selects by auditUniverseId (name only for old rows without an id); the picker shows the row's own
// entity even when it is not in the loaded universe lists.
const auditableEntityValue = (activity: AuditActivities): AuditableEntityOption | undefined =>
  findAuditableEntityForActivity(auditableUniverseOptions.value, activity)

const handleSelectAuditableEntity = (activity: AuditActivities, val: unknown) => {
  // Resolved by corporate id, never by name, so entities sharing a name cannot be confused.
  const match = resolveAuditableEntitySelection(auditableUniverseOptions.value, val)
  if (match) {
    activity.auditUniverseId = match.id || ''
    activity.riskName = match.name
    activity.riskLevel = match.riskLevel || NOT_SCORED_RISK_LEVEL

    // Suggest default activity title if empty or using standard default prefix
    const defaultNamePrefixes = ['Audit Operasional', t('annualAudit.form.defaultActivityName', { entity: '|' }).split('|')[0]!.trim()]
    if (!activity.name || defaultNamePrefixes.some(prefix => activity.name!.startsWith(prefix))) {
      activity.name = t('annualAudit.form.defaultActivityName', { entity: activity.riskName })
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
    // Only use a department the master data actually returned; the guess is kept only when its name matches one.
    const matchedDept = store.departments.find(d => d.name.toLowerCase() === detectedDept.toLowerCase())
    if (matchedDept) {
      activity.departmentId = matchedDept.id
      activity.department = matchedDept.name
    }
  } else {
    activity.auditUniverseId = ''
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
</script>
