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
        <UButton
          color="neutral"
          variant="outline"
          :label="t('consultingService.dashboard.importDocument')"
          icon="i-lucide-upload"
          class="px-5 py-2.5 font-bold rounded-lg shadow w-full sm:w-auto"
          to="/consulting-service/upload"
        />
        <UButton
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
          <UInput v-model="filters.search" :placeholder="t('consultingService.dashboard.filters.searchPlaceholder')" icon="i-lucide-search" class="w-full" />
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
          <div v-if="(row.original as any).attachment" class="flex items-center gap-1.5">
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

            <UTooltip :text="t('consultingService.dashboard.actions.edit')">
              <UButton
                icon="i-lucide-edit"
                color="warning"
                variant="ghost"
                size="md"
              @click="store.editService(row.original as any)"
            />
            </UTooltip>

            <UTooltip :text="t('consultingService.dashboard.actions.delete')">
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
    <!-- Form Modal -->
    <UModal
      v-model:open="store.isFormOpen"
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
          :ui="{ body: 'overflow-y-auto flex-1 p-4 sm:p-6' }"
        >
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">
                {{ store.isEditing ? t('consultingService.dashboard.form.editTitle') : t('consultingService.dashboard.form.newTitle') }}
              </h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.closeForm" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveService" class="space-y-4">
            <UFormField :label="t('consultingService.dashboard.form.projectTitle')" required>
              <UInput
                v-model="store.newService.title"
                :placeholder="t('consultingService.dashboard.form.projectPlaceholder')"
                class="w-full"
                required
                maxlength="100"
                @invalid="($event.target as any)?.setCustomValidity(t('consultingService.dashboard.form.titleValidation'))"
                @input="($event.target as any)?.setCustomValidity('')"
              />
              <div class="text-xs text-gray-500 mt-1 text-right">
                {{ store.newService.title ? store.newService.title.length : 0 }}/100
              </div>
            </UFormField>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <UFormField :label="t('consultingService.dashboard.form.category')">
                <USelect v-model="store.newService.category" :items="categoryOptions" value-key="value" class="w-full" />
              </UFormField>

              <UFormField :label="t('consultingService.dashboard.form.requestorDepartment')">
                <USelect v-model="store.newService.requestorDept" :items="store.departments" class="w-full" />
              </UFormField>
            </div>

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
                  @invalid="($event.target as any)?.setCustomValidity(t('consultingService.dashboard.form.consultantValidation'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.newService.consultantName ? store.newService.consultantName.length : 0 }}/50
                </div>
              </UFormField>
            </div>

            <UFormField :label="t('consultingService.dashboard.form.status')">
              <USelect v-model="store.newService.status" :items="statusOptions" value-key="value" class="w-full" />
            </UFormField>

            <UFormField :label="t('consultingService.dashboard.form.notes')">
              <UTextarea v-model="store.newService.notes" :placeholder="t('consultingService.dashboard.form.notesPlaceholder')" class="w-full" />
            </UFormField>

            <!-- File Upload -->
            <UFormField :label="t('consultingService.dashboard.form.upload')">
              <div class="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-primary-500 transition cursor-pointer relative">
                <input
                  type="file"
                  class="absolute inset-0 opacity-0 cursor-pointer"
                  @change="(e: any) => store.handleFileUpload(e.target.files)"
                />
                <UIcon name="i-lucide-upload-cloud" class="w-10 h-10 text-gray-400 mx-auto mb-2" />
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ store.newService.attachment ? store.newService.attachment.name : t('consultingService.dashboard.form.uploadPrompt') }}
                </p>
                <p class="text-md text-gray-500 mt-1">{{ t('consultingService.dashboard.form.uploadHint') }}</p>
              </div>
            </UFormField>

            <div v-if="store.errorMsg" class="text-red-500 text-sm font-medium">
              {{ store.errorMsg }}
            </div>

            <div class="flex flex-col-reverse sm:flex-row justify-end gap-3 pt-4 border-t">
              <UButton :label="t('consultingService.dashboard.form.cancel')" color="neutral" variant="ghost" class="w-full sm:w-auto" @click="store.closeForm" />
              <UButton type="submit" :label="t('consultingService.dashboard.form.save')" color="primary" class="w-full sm:w-auto font-bold" :loading="store.loading" />
            </div>
          </UForm>
        </UCard>
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
                <UButton
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
import { ref, reactive, computed } from 'vue'
import { useConsultingServiceStore } from '~/stores/consulting-service'

const store = useConsultingServiceStore()
const { t } = useI18n()
store.fetchServices()

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
