<template>
  <div class="p-4 sm:p-6 space-y-6 sm:space-y-8 max-w-full mx-auto min-h-screen min-w-0">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="space-y-1">
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white uppercase tracking-tight">
          {{ t('consultingService.dashboard.title') }}
        </h1>
        <p class="text-sm sm:text-base text-gray-600 dark:text-gray-400 font-medium">
          {{ t('consultingService.dashboard.subtitle') }}
        </p>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 sm:gap-3 w-full sm:w-auto">
        <UButton v-if="canWriteQaConsulting"
          color="neutral"
          variant="outline"
          :label="t('consultingService.dashboard.importDocument')"
          icon="i-lucide-upload"
          class="px-5 py-2.5 font-bold rounded-lg shadow w-full sm:w-auto"
          to="/consulting-service/upload"
        />
        <UButton v-if="canWriteQaConsulting"
          color="primary"
          :label="t('consultingService.dashboard.newAssignment')"
          icon="i-lucide-plus"
          class="px-6 py-2.5 font-bold rounded-lg shadow-lg w-full sm:w-auto"
          @click="store.openForm()"
        />
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Total -->
      <div class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] dashboard-card">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('consultingService.dashboard.summary.total') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ store.services.length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
            <UIcon name="i-lucide-clipboard-list" class="text-indigo-600 size-5 dark:text-indigo-400" />
          </div>
        </div>
      </div>

      <!-- In Progress -->
      <div class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] dashboard-card">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('consultingService.dashboard.summary.inProgress') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ store.services.filter(s => s.status === 'In Progress').length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
            <UIcon name="i-lucide-refresh-cw" class="text-amber-600 size-5 dark:text-amber-400 animate-spin-slow" />
          </div>
        </div>
      </div>

      <!-- Completed -->
      <div class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] dashboard-card">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('consultingService.dashboard.summary.completed') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ store.services.filter(s => s.status === 'Completed').length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
            <UIcon name="i-lucide-check-circle" class="text-emerald-600 size-5 dark:text-emerald-400" />
          </div>
        </div>
      </div>

      <!-- Planned -->
      <div class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] dashboard-card">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('consultingService.dashboard.summary.planned') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ store.services.filter(s => s.status === 'Planned').length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-sky-50 p-2.5 flex items-center justify-center dark:bg-sky-500/10">
            <UIcon name="i-lucide-calendar" class="text-sky-600 size-5 dark:text-sky-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <UCard class="shadow-sm">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <UFormField :label="t('consultingService.dashboard.filters.searchTitle')">
          <CommonSearchInput v-model="filters.search" :placeholder="t('consultingService.dashboard.filters.searchPlaceholder')" class="w-full" />
        </UFormField>
        <UFormField :label="t('consultingService.dashboard.filters.category')">
          <USelectMenu v-model="filters.category" :items="categoryFilterOptions" value-key="value" :placeholder="t('consultingService.dashboard.filters.selectCategory')" class="w-full" />
        </UFormField>
        <UFormField :label="t('consultingService.dashboard.filters.status')">
          <USelectMenu v-model="filters.status" :items="statusFilterOptions" value-key="value" :placeholder="t('consultingService.dashboard.filters.selectStatus')" class="w-full" />
        </UFormField>
        <UFormField :label="t('consultingService.dashboard.filters.department')">
          <USelectMenu v-model="filters.dept" :items="departmentFilterOptions" value-key="value" :placeholder="t('consultingService.dashboard.filters.selectDepartment')" class="w-full" />
        </UFormField>
      </div>
    </UCard>

    <!-- Data Table -->
    <UCard class="shadow-md overflow-hidden">
      <TableEntities :data="filteredServices" :columns="columns" class="w-full">
        <!-- Status Cell -->
        <template #status-cell="{ row }">
          <UBadge :color="store.getStatusColor((row.original as any).status)" variant="subtle" size="md">
            {{ getStatusLabel((row.original as any).status) }}
          </UBadge>
        </template>

        <template #category-cell="{ row }">
          {{ getCategoryLabel((row.original as any).category) }}
        </template>

        <!-- Attachment Cell -->
        <template #attachment-cell="{ row }">
          <span v-if="(row.original as any).attachment && !canAccessFiles" class="text-md">{{ (row.original as any).attachment.name }}</span>
          <div v-else-if="(row.original as any).attachment" class="flex items-center gap-1.5">
            <UButton
              icon="i-lucide-download"
              color="primary"
              variant="ghost"
              size="sm"
              :label="(row.original as any).attachment.name"
              @click="store.viewAttachment((row.original as any).id, (row.original as any).attachment.name)"
            />
          </div>
          <span v-else class="text-gray-400 italic text-md">{{ t('consultingService.dashboard.noFile') }}</span>
        </template>

        <!-- Actions Cell -->
        <template #actions-cell="{ row }">
          <div class="flex items-center gap-1">
            <UTooltip :text="t('consultingService.dashboard.actions.viewDetails')" >
              <UButton
                icon="i-lucide-eye"
                color="neutral"
                variant="ghost"
                size="md"
                @click="store.openDetail(row.original as any)"
              />
            </UTooltip>

            <UTooltip v-if="canWriteQaConsulting" :text="t('consultingService.dashboard.actions.edit')">
              <UButton
                icon="i-lucide-edit"
                color="warning"
                variant="ghost"
                size="md"
              @click="store.editService(row.original as any)"
            />
            </UTooltip>

            <UTooltip v-if="canWriteQaConsulting" :text="t('consultingService.dashboard.actions.delete')">
              <UButton
                icon="i-lucide-trash-2"
                color="error"
                variant="ghost"
                size="md"
                @click="confirmDelete(row.original as any)"
              />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
    </UCard>

    <!-- Form Modal -->
    <UModal
      :open="store.isFormOpen"
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
      @update:open="(open: boolean) => { if (!open) store.closeForm() }"
    >
      <template #content>
        <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
          <!-- Header -->
          <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600 dark:text-primary-400">
                <UIcon
                  name="i-lucide-briefcase"
                  class="size-6"
                />
              </div>
              <div>
                <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                  {{ store.isEditing ? t('consultingService.dashboard.form.editTitle') : t('consultingService.dashboard.form.newTitle') }}
                </h3>
                <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                  {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
                </p>
              </div>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark-20-solid"
              class="-my-1"
              :aria-label="t('common.close')"
              @click="store.closeForm"
            />
          </div>

          <!-- Stepper -->
          <div class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
            <div class="flex min-w-max items-center gap-3 px-1 sm:gap-4">
              <template
                v-for="(step, idx) in steps"
                :key="step.key"
              >
                <div
                  class="group flex shrink-0 cursor-pointer select-none items-center gap-2 sm:gap-2.5"
                  data-testid="form-step"
                  @click="goToStep(idx)"
                >
                  <div
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 sm:h-8 sm:w-8"
                    :class="currentStep === idx
                      ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105'
                      : currentStep > idx
                        ? 'bg-secondary text-white shadow-sm'
                        : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
                  >
                    <UIcon
                      v-if="currentStep > idx"
                      name="i-heroicons-check"
                      class="h-3.5 w-3.5 stroke-[3]"
                    />
                    <span v-else>{{ idx + 1 }}</span>
                  </div>
                  <span
                    class="whitespace-nowrap text-xs font-semibold transition-colors"
                    :class="currentStep === idx
                      ? 'text-primary font-bold'
                      : currentStep > idx
                        ? 'text-secondary font-semibold'
                        : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'"
                  >
                    {{ step.label }}
                  </span>
                </div>
                <div
                  v-if="idx < steps.length - 1"
                  class="h-0.5 w-8 shrink-0 rounded transition-all duration-300 sm:w-12"
                  :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
                />
              </template>
            </div>
          </div>

          <!-- Body -->
          <div
            ref="formBodyRef"
            class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6"
          >
            <!-- Validation / Blocking Alert -->
            <UAlert
              v-if="formErrorMessage || store.errorMsg"
              color="error"
              variant="subtle"
              icon="i-heroicons-exclamation-triangle"
              :title="t('common.error')"
              :description="formErrorMessage || store.errorMsg"
              close
              class="shadow-sm"
              @update:open="(open: boolean) => { if (!open) { formErrorMessage = ''; store.errorMsg = '' } }"
            />

            <!-- Step 1: Project Information (Project Title sampai Consultant / Lead Auditor) -->
            <div v-show="currentStep === 0" class="space-y-5">
              <UCard
                :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
                class="border border-[var(--border-main)]"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-lucide-file-text"
                      class="h-5 w-5 text-primary-500"
                    />
                    <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                      {{ t('consultingService.dashboard.form.stepper.assignmentInfo') }}
                    </h3>
                  </div>
                </template>

                <!-- Project Title -->
                <UFormField :label="t('consultingService.dashboard.form.projectTitle')" required>
                  <UInput
                    v-model="store.newService.title"
                    :placeholder="t('consultingService.dashboard.form.projectPlaceholder')"
                    class="w-full"
                    required
                    maxlength="100"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.newService.title ? store.newService.title.length : 0 }}/100
                  </div>
                </UFormField>

                <!-- Category & Requestor Department -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <UFormField :label="t('consultingService.dashboard.form.category')">
                    <USelect v-model="store.newService.category" :items="categoryOptions" value-key="value" class="w-full" />
                  </UFormField>

                  <UFormField :label="t('consultingService.dashboard.form.requestorDepartment')">
                    <USelect v-model="store.newService.requestorDept" :items="store.departments" class="w-full" />
                  </UFormField>
                </div>

                <!-- Period Quarter, Period Year & Consultant / Lead Auditor -->
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <UFormField :label="t('consultingService.dashboard.form.periodQuarter')">
                    <USelect v-model="store.newService.periodQuarter" :items="store.quarters" class="w-full" />
                  </UFormField>

                  <UFormField :label="t('consultingService.dashboard.form.periodYear')">
                    <USelect v-model="store.newService.periodYear" :items="store.years" class="w-full" />
                  </UFormField>

                  <UFormField :label="t('consultingService.dashboard.form.consultant')" required>
                    <UInput
                      v-model="store.newService.consultantName"
                      :placeholder="t('consultingService.dashboard.form.consultantPlaceholder')"
                      class="w-full"
                      required
                      maxlength="50"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ store.newService.consultantName ? store.newService.consultantName.length : 0 }}/50
                    </div>
                  </UFormField>
                </div>
              </UCard>
            </div>

            <!-- Step 2: Status sampai Upload Support Document / Report -->
            <div v-show="currentStep === 1" class="space-y-5">
              <UCard
                :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
                class="border border-[var(--border-main)]"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-lucide-file-check"
                      class="h-5 w-5 text-secondary-500"
                    />
                    <h3 class="text-base sm:text-lg font-bold text-[var(--text-main)]">
                      {{ t('consultingService.dashboard.form.stepper.statusAndDocument') }}
                    </h3>
                  </div>
                </template>

                <!-- Status -->
                <UFormField :label="t('consultingService.dashboard.form.status')">
                  <USelect v-model="store.newService.status" :items="statusOptions" value-key="value" class="w-full" />
                </UFormField>

                <!-- Notes / Description -->
                <UFormField :label="t('consultingService.dashboard.form.notes')">
                  <UTextarea v-model="store.newService.notes" :rows="3" :placeholder="t('consultingService.dashboard.form.notesPlaceholder')" class="w-full" />
                </UFormField>

                <!-- File Upload -->
                <UFormField :label="t('consultingService.dashboard.form.upload')">
                  <div class="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg p-6 text-center hover:border-primary-500 transition cursor-pointer relative bg-gray-50/50 dark:bg-gray-800/50">
                    <input
                      type="file"
                      class="absolute inset-0 opacity-0 cursor-pointer"
                      @change="(e: any) => store.handleFileUpload(e.target.files)"
                    />
                    <UIcon name="i-lucide-upload-cloud" class="w-10 h-10 text-gray-400 mx-auto mb-2" />
                    <p class="text-sm font-medium text-gray-700 dark:text-gray-300">
                      {{ store.newService.attachment ? store.newService.attachment.name : t('consultingService.dashboard.form.uploadPrompt') }}
                    </p>
                    <p class="text-xs text-gray-500 mt-1">{{ t('consultingService.dashboard.form.uploadHint') }}</p>
                  </div>
                </UFormField>
              </UCard>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
            <div class="flex w-full items-center justify-start gap-2 sm:w-auto">
              <UButton
                :label="t('common.cancel')"
                color="neutral"
                variant="ghost"
                class="w-full sm:w-auto"
                @click="store.closeForm"
              />
            </div>

            <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
              <UButton
                v-if="currentStep > 0"
                icon="i-heroicons-arrow-left"
                :label="t('auditActivityPlan.stepper.previous')"
                color="neutral"
                variant="outline"
                class="flex-1 sm:flex-none"
                @click="() => { currentStep-- }"
              />
              <UButton
                v-if="currentStep < steps.length - 1"
                trailing-icon="i-heroicons-arrow-right"
                :label="t('auditActivityPlan.stepper.next')"
                color="primary"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                @click="handleNext"
              />
              <UButton
                v-else
                icon="i-heroicons-check"
                :label="store.isEditing ? t('common.updateData') : t('consultingService.dashboard.form.save')"
                color="primary"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                :loading="store.loading"
                @click="handleSave"
              />
            </div>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Detail Modal -->
    <UModal
      v-model:open="store.isDetailOpen"
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <UCard
          class="w-full shadow-2xl flex flex-col max-h-[90vh]"
          :ui="{ body: 'p-4 sm:p-6' }"
        >
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('consultingService.dashboard.detail.title') }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.closeDetail" />
            </div>
          </template>

          <div v-if="store.selectedService" class="space-y-6">
            <div>
              <h4 class="text-xl font-bold text-gray-900 dark:text-white">{{ store.selectedService.title }}</h4>
              <UBadge :color="store.getStatusColor(store.selectedService.status)" variant="subtle" size="md" class="mt-2">
                {{ getStatusLabel(store.selectedService.status) }}
              </UBadge>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm bg-gray-50 dark:bg-gray-900 p-4 rounded-lg">
              <div>
                <span class="text-md text-gray-400 block">{{ t('consultingService.dashboard.detail.category') }}</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ getCategoryLabel(store.selectedService.category) }}</span>
              </div>
              <div>
                <span class="text-md text-gray-400 block">{{ t('consultingService.dashboard.detail.period') }}</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.period }}</span>
              </div>
              <div>
                <span class="text-md text-gray-400 block">{{ t('consultingService.dashboard.detail.requestorDepartment') }}</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.requestorDept }}</span>
              </div>
              <div>
                <span class="text-md text-gray-400 block">{{ t('consultingService.dashboard.detail.consultantLead') }}</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.consultantName }}</span>
              </div>
            </div>

            <div>
              <span class="text-md text-gray-400 block">{{ t('consultingService.dashboard.detail.description') }}</span>
              <p class="text-sm text-gray-700 dark:text-gray-300 mt-1 whitespace-pre-line bg-gray-50 dark:bg-gray-900 p-3 rounded">
                {{ store.selectedService.notes || t('consultingService.dashboard.detail.noDescription') }}
              </p>
            </div>

            <!-- Attachment details -->
            <div v-if="store.selectedService.attachment">
              <span class="text-md text-gray-400 block mb-2">{{ t('consultingService.dashboard.detail.attachedDocument') }}</span>
              <div class="flex items-center justify-between p-3 bg-primary-50 dark:bg-primary-950 rounded-lg border border-primary-100">
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-file-text" class="text-primary-500 w-5 h-5" />
                  <div>
                    <span class="font-semibold text-sm text-gray-800 dark:text-gray-200">{{ store.selectedService.attachment.name }}</span>
                    <span class="text-md text-gray-400 block">{{ store.selectedService.attachment.size }} · {{ t('consultingService.dashboard.detail.uploadedAt', { date: store.selectedService.attachment.uploadedAt }) }}</span>
                  </div>
                </div>
                <UButton v-if="canAccessFiles"
                  icon="i-heroicons-document-arrow-down"
                  :label="t('consultingService.dashboard.actions.download')"
                  color="primary"
                  variant="solid"
                  size="sm"
                  @click="store.downloadAttachment(store.selectedService.id, store.selectedService.attachment.name)"
                />
              </div>
            </div>

            <div class="flex justify-end pt-4 border-t">
              <UButton
                :label="t('consultingService.dashboard.actions.close')"
                color="neutral"
                variant="ghost"
                @click="store.closeDetail"
              />
            </div>
          </div>
        </UCard>
      </template>
    </UModal>

  </div>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useConsultingServiceStore } from '~/stores/consulting-service'

const store = useConsultingServiceStore()
const { t } = useI18n()
const { canWriteQaConsulting, canAccessFiles } = useRbac()
store.fetchServices()

const currentStep = ref(0)
const formBodyRef = ref<HTMLElement | null>(null)
const formErrorMessage = ref('')

const steps = computed(() => [
  { key: 'assignmentInfo', label: t('consultingService.dashboard.form.stepper.assignmentInfo') || 'Informasi Proyek & Konsultan' },
  { key: 'statusAndDocument', label: t('consultingService.dashboard.form.stepper.statusAndDocument') || 'Status & Dokumen Pendukung' }
])

const validateStep = (idx: number): boolean => {
  if (idx === 0) {
    if (!store.newService.title?.trim()) {
      formErrorMessage.value = t('consultingService.dashboard.form.titleRequired') || 'Judul Proyek wajib diisi'
      return false
    }
    if (store.newService.title.length > 100) {
      formErrorMessage.value = t('consultingService.dashboard.form.titleValidation') || 'Judul wajib diisi dan maksimal 100 karakter'
      return false
    }
    if (!store.newService.consultantName?.trim()) {
      formErrorMessage.value = t('consultingService.dashboard.form.consultantRequired') || 'Konsultan / Ketua Auditor wajib diisi'
      return false
    }
    if (store.newService.consultantName.length > 50) {
      formErrorMessage.value = t('consultingService.dashboard.form.consultantValidation') || 'Konsultan wajib diisi dan maksimal 50 karakter'
      return false
    }
    formErrorMessage.value = ''
    return true
  }
  if (idx === 1) {
    formErrorMessage.value = ''
    return true
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    formErrorMessage.value = ''
    return
  }
  for (let i = currentStep.value; i < targetIdx; i++) {
    if (!validateStep(i)) return
  }
  formErrorMessage.value = ''
  currentStep.value = targetIdx
  nextTick(() => {
    formBodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
  })
}

const handleNext = () => {
  if (!validateStep(currentStep.value)) return
  formErrorMessage.value = ''
  currentStep.value++
  nextTick(() => {
    formBodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
  })
}

const handleSave = async () => {
  if (!validateStep(0)) {
    currentStep.value = 0
    return
  }
  if (!validateStep(1)) {
    currentStep.value = 1
    return
  }
  formErrorMessage.value = ''
  await store.saveService()
}

watch(
  () => store.isFormOpen,
  (open) => {
    if (open) {
      currentStep.value = 0
      formErrorMessage.value = ''
    }
  }
)

const columns = computed(() => [
  { accessorKey: 'title', header: t('consultingService.dashboard.columns.projectTitle') },
  { accessorKey: 'category', header: t('consultingService.dashboard.columns.category') },
  { accessorKey: 'requestorDept', header: t('consultingService.dashboard.columns.department') },
  { accessorKey: 'period', header: t('consultingService.dashboard.columns.period') },
  { accessorKey: 'consultantName', header: t('consultingService.dashboard.columns.consultantLead') },
  { accessorKey: 'status', header: t('consultingService.dashboard.columns.status') },
  { accessorKey: 'attachment', header: t('consultingService.dashboard.columns.document') },
  { accessorKey: 'actions', header: t('consultingService.dashboard.columns.actions') }
])

const statusKeys: Record<string, string> = {
  Planned: 'planned',
  'In Progress': 'inProgress',
  Completed: 'completed',
  Cancelled: 'cancelled'
}

const categoryKeys: Record<string, string> = {
  'Operational Advisory': 'operationalAdvisory',
  'IT Advisory': 'itAdvisory',
  'Training & Workshops': 'trainingWorkshops',
  'Policy & SOP Review': 'policySopReview',
  'Governance Assessment': 'governanceAssessment'
}

const getStatusLabel = (status: string) => {
  const key = statusKeys[status]
  return key ? t(`consultingService.dashboard.statuses.${key}`) : status
}

const getCategoryLabel = (category: string) => {
  const key = categoryKeys[category]
  return key ? t(`consultingService.dashboard.categories.${key}`) : category
}

const statusOptions = computed(() => store.statuses.map(value => ({ label: getStatusLabel(value), value })))
const categoryOptions = computed(() => store.categories.map(value => ({ label: getCategoryLabel(value), value })))
const statusFilterOptions = computed(() => [{ label: t('consultingService.dashboard.filters.all'), value: 'All' }, ...statusOptions.value])
const categoryFilterOptions = computed(() => [{ label: t('consultingService.dashboard.filters.all'), value: 'All' }, ...categoryOptions.value])
const departmentFilterOptions = computed(() => [
  { label: t('consultingService.dashboard.filters.all'), value: 'All' },
  ...store.departments.map(value => ({ label: value, value }))
])

const filters = reactive({
  search: '',
  category: 'All',
  status: 'All',
  dept: 'All'
})

const filteredServices = computed(() => {
  return store.services.filter(s => {
    const matchesSearch = s.title.toLowerCase().includes(filters.search.toLowerCase()) ||
      s.consultantName.toLowerCase().includes(filters.search.toLowerCase())
    const matchesCategory = filters.category === 'All' || s.category === filters.category
    const matchesStatus = filters.status === 'All' || s.status === filters.status
    const matchesDept = filters.dept === 'All' || s.requestorDept === filters.dept
    return matchesSearch && matchesCategory && matchesStatus && matchesDept
  })
})

const confirmDelete = async (service: any) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('consultingService.dashboard.deleteConfirm') })) {
    store.deleteService(service)
  }
}
</script>
