<template>
  <div>
    <!-- Empty State -->
    <div
      v-if="!store.loading && store.guidelines.length === 0"
      class="flex flex-col items-center justify-center p-12 bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl text-center space-y-6 shadow-sm my-4"
    >
      <div class="w-16 h-16 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
        <UIcon name="i-lucide-book-open" class="w-8 h-8" />
      </div>
      <div class="space-y-2 max-w-md">
        <h2 class="text-xl font-bold text-[var(--text-main)]">{{ t('auditCharter.guideline.emptyTitle') }}</h2>
        <p class="text-sm text-[var(--text-muted)] leading-relaxed">
          {{ t('auditCharter.guideline.emptyDesc') }}
        </p>
      </div>
      <UButton
        v-if="canManageCharter"
        :label="t('auditCharter.guideline.addGuideline')"
        @click="() => { store.showModal = true }"
        color="primary"
        class="w-full sm:w-auto min-w-[120px] rounded-xl px-6 py-3 font-semibold justify-center text-center transition-all duration-200"
        icon="i-lucide-plus"
      />
    </div>

    <!-- Data State -->
    <div v-else class="space-y-4">
      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
        <div>
          <h2 class="text-2xl font-bold text-[var(--text-main)]">{{ t('auditCharter.guideline.title') }}</h2>
          <p class="text-sm text-[var(--text-muted)]">{{ t('auditCharter.guideline.subtitle') }}</p>
        </div>
        <UButton
          v-if="canManageCharter"
          :label="t('auditCharter.guideline.addGuidelineShort')"
          @click="() => { store.showModal = true }"
          color="primary"
          icon="i-lucide-plus"
          class="w-full sm:w-auto min-w-[120px] font-bold justify-center text-center"
        />
      </div>

      <!-- Desktop View: Table -->
      <div class="hidden lg:block">
        <TableEntities
          :data="tableData"
          :columns="columns"
          :loading="store.loading"
          :server-side="true"
          :total="store.pagination.total"
          :items-per-page="store.pagination.page_size"
          :page="store.pagination.page"
          :empty-state="{
            icon: 'i-lucide-book-open',
            label: t('auditCharter.guideline.emptyTable')
          }"
          :ui="{ root: 'overflow-visible' }"
          class="w-full"
          @update:page="(p) => store.fetchGuidelines(p)"
          @update:items-per-page="(size) => store.setPageSize(size)"
        >
          <!-- No slot -->
          <template #no-cell="{ row }">
            <span class="font-medium text-[var(--text-muted)] block text-center">{{ row.original.no }}</span>
          </template>

          <!-- Name slot -->
          <template #name-cell="{ row }">
            <ReadMoreText
              :text="row.original.name"
              :max-length="75"
              text-class="font-semibold text-[var(--text-main)]"
            />
          </template>

          <!-- Status slot -->
          <template #status-cell="{ row }">
            <div class="flex justify-center">
              <UBadge
                :color="row.original.status === 'Aktif' ? 'secondary' : 'warning'"
                variant="subtle"
                class="rounded font-semibold"
              >
                {{ translateStatus(row.original.status) }}
              </UBadge>
            </div>
          </template>

          <!-- Effective date slot -->
          <template #effective_date-cell="{ row }">
            <span class="font-medium text-[var(--text-main)] block text-center">{{
              formatMonthYear(row.original.effective_date)
            }}</span>
          </template>

          <!-- File Name slot -->
          <template #file_name-cell="{ row }">
            <div
              class="line-clamp-2 break-all text-sm font-normal text-[var(--text-main)]"
              :title="row.original.file_name"
            >
              {{ row.original.file_name || '-' }}
            </div>
          </template>

          <!-- Actions slot -->
          <template #actions-cell="{ row }">
            <div class="flex justify-center items-center gap-1">
              <UTooltip v-if="canAccessFiles" :text="t('auditCharter.tooltips.viewGuideline')">
                <UButton
                  v-if="row.original.file_url && row.original.file_url !== '#'"
                  icon="i-lucide-eye"
                  color="neutral"
                  variant="ghost"
                  size="md"
                  @click="openFile(row.original.file_url, row.original.name)"
                />
              </UTooltip>
              <UTooltip :text="t('auditCharter.tooltips.editGuideline')">
                <UButton
                  v-if="canManageCharter"
                  size="md"
                  color="warning"
                  variant="ghost"
                  icon="i-lucide-edit"
                  @click="store.handleEdit(row.original)"
                />
              </UTooltip>
              <UTooltip :text="t('auditCharter.tooltips.deleteGuideline')">
                <UButton
                  v-if="canManageCharter"
                  size="md"
                  color="error"
                  variant="ghost"
                  icon="i-lucide-trash-2"
                  @click="confirmDelete(row.original)"
                />
              </UTooltip>
            </div>
          </template>
        </TableEntities>
      </div>

      <!-- Mobile / Tablet View: Stacked Cards -->
      <div class="block lg:hidden space-y-3">
        <!-- Skeleton Loading -->
        <div v-if="store.loading" class="space-y-3">
          <div
            v-for="i in 3"
            :key="i"
            class="p-4 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] space-y-3 animate-pulse"
          >
            <div class="flex items-center justify-between">
              <div class="h-4 w-10 bg-gray-200 dark:bg-gray-800 rounded"></div>
              <div class="h-5 w-16 bg-gray-200 dark:bg-gray-800 rounded-full"></div>
            </div>
            <div class="h-4 w-3/4 bg-gray-200 dark:bg-gray-800 rounded"></div>
            <div class="h-3 w-1/2 bg-gray-200 dark:bg-gray-800 rounded"></div>
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-else-if="tableData.length === 0"
          class="p-8 text-center text-sm text-[var(--text-muted)] bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-xl"
        >
          <UIcon name="i-lucide-book-open" class="w-8 h-8 mx-auto mb-2 text-gray-400" />
          <p>{{ t('auditCharter.guideline.emptyTable') }}</p>
        </div>

        <!-- Card List -->
        <div v-else class="space-y-3">
          <div
            v-for="item in tableData"
            :key="item.id || item.no"
            class="p-4 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] shadow-xs transition-shadow hover:shadow-sm space-y-3"
          >
            <!-- Card Header: Number & Status Badge -->
            <div class="flex items-center justify-between gap-2">
              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-gray-100 dark:bg-gray-800 text-[var(--text-muted)]">
                #{{ item.no }}
              </span>
              <UBadge
                :color="item.status === 'Aktif' ? 'secondary' : 'warning'"
                variant="subtle"
                class="rounded font-semibold text-xs"
              >
                {{ translateStatus(item.status) }}
              </UBadge>
            </div>

            <!-- Title / Name -->
            <div>
              <h3 class="font-semibold text-[var(--text-main)] text-sm sm:text-base leading-snug">
                {{ item.name }}
              </h3>
            </div>

            <!-- Details: Effective Date & Attachment -->
            <div class="pt-2 border-t border-[var(--border-main)]/60 flex flex-col gap-1.5 text-xs text-[var(--text-muted)]">
              <div class="flex items-center gap-1.5">
                <UIcon name="i-lucide-calendar" class="w-3.5 h-3.5 shrink-0 text-gray-400" />
                <span>{{ t('auditCharter.guideline.columns.effectiveDate') || 'Mulai Berlaku' }}:</span>
                <span class="font-medium text-[var(--text-main)]">{{ formatMonthYear(item.effective_date) }}</span>
              </div>
              <div v-if="item.file_name" class="flex items-start gap-1.5">
                <UIcon name="i-lucide-file-text" class="w-3.5 h-3.5 shrink-0 mt-0.5 text-gray-400" />
                <span class="break-all text-[var(--text-main)]">{{ item.file_name }}</span>
              </div>
            </div>

            <!-- Actions Row -->
            <div class="flex items-center justify-end gap-1.5 pt-2 border-t border-[var(--border-main)]/60">
              <UButton
                v-if="canAccessFiles && item.file_url && item.file_url !== '#'"
                icon="i-lucide-eye"
                color="neutral"
                variant="soft"
                size="xs"
                :label="t('common.actions.view') || 'Lihat'"
                @click="openFile(item.file_url, item.name)"
              />
              <UButton
                v-if="canManageCharter"
                size="xs"
                color="warning"
                variant="soft"
                icon="i-lucide-edit"
                :label="t('common.actions.edit') || 'Ubah'"
                @click="store.handleEdit(item)"
              />
              <UButton
                v-if="canManageCharter"
                size="xs"
                color="error"
                variant="soft"
                icon="i-lucide-trash-2"
                :label="t('common.actions.delete') || 'Hapus'"
                @click="confirmDelete(item)"
              />
            </div>
          </div>

          <!-- Mobile Pagination -->
          <div
            v-if="store.pagination.total > store.pagination.page_size"
            class="flex flex-col sm:flex-row items-center justify-between gap-2 pt-2 px-1"
          >
            <span class="text-xs text-[var(--text-muted)]">
              Total: {{ store.pagination.total }} (Hal. {{ store.pagination.page }})
            </span>
            <UPagination
              v-model:page="store.pagination.page"
              :total="store.pagination.total"
              :items-per-page="store.pagination.page_size"
              size="xs"
              active-color="primary"
              color="neutral"
              variant="outline"
              class="max-w-full overflow-x-auto"
              @update:page="(p) => store.fetchGuidelines(p)"
            />
          </div>
        </div>
      </div>
          </div>

          <!-- Mobile Pagination -->
          <div
            v-if="store.pagination.total > store.pagination.page_size"
            class="flex flex-col sm:flex-row items-center justify-between gap-2 pt-2 px-1"
          >
            <span class="text-xs text-[var(--text-muted)]">
              Total: {{ store.pagination.total }} (Hal. {{ store.pagination.page }})
            </span>
            <UPagination
              v-model:page="store.pagination.page"
              :total="store.pagination.total"
              :items-per-page="store.pagination.page_size"
              size="xs"
              active-color="primary"
              color="neutral"
              variant="outline"
              class="max-w-full overflow-x-auto"
              @update:page="(p) => store.fetchGuidelines(p)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useGuidelineStore } from '~/stores/guideline'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import TableEntities from '~/components/shared/TableEntities.vue'
import { useGlobalModalStore } from '~/stores/global-modal'
import ReadMoreText from '~/components/shared/ReadMoreText.vue'
import { openUploadedFile } from '~/utils/documentViewer'
import { extractErrorMessage } from '~/utils/error'

const { t, locale } = useI18n()
const store = useGuidelineStore()
const { canManageCharter, canAccessFiles } = useRbac()

const columns = computed(() => store.columns)

const tableData = computed(() => {
  return store.guidelines.map((item, index) => ({
    ...item,
    no: (store.pagination.page - 1) * store.pagination.page_size + index + 1
  }))
})

const translateStatus = (status: string): string => {
  if (status === 'Aktif') return t('auditCharter.guideline.statusActive')
  if (status === 'Sedang Diperbarui') return t('auditCharter.guideline.statusUnderReview')
  return status
}

const formatMonthYear = (val: string) => {
  if (!val) return '-'
  const parts = val.split('-')
  if (parts.length < 2) return val
  const [year, month] = parts
  const mIndex = parseInt(month || '', 10) - 1
  if (mIndex < 0 || mIndex >= 12) return val

  if (locale.value === 'id') {
    const monthsId = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ]
    return `${monthsId[mIndex]} ${year}`
  }
  const monthsEn = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ]
  return `${monthsEn[mIndex]} ${year}`
}

const openFile = async (fileUrl: string, name?: string) => {
  try {
    await openUploadedFile(fileUrl, getAuditServiceBaseUrl(), name)
  } catch (err) {
    console.error('Failed to open guideline file:', err)
    useAppToast().error(t('auditCharter.toasts.guideline.openFailed'), extractErrorMessage(err, t('auditCharter.toasts.guideline.openFailed')))
  }
}

const confirmDelete = async (item: any) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('auditCharter.guideline.deleteConfirm', { name: item.name }) })) {
    await store.deleteGuideline(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchGuidelines()
})
</script>