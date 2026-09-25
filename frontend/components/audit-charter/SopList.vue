<template>
  <div>
    <!-- Empty State -->
    <div
      v-if="!store.loading && store.sops.length === 0"
      class="flex flex-col items-center justify-center p-12 bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl text-center space-y-6 shadow-sm my-4"
    >
      <div class="w-16 h-16 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
        <UIcon name="i-lucide-file-text" class="w-8 h-8" />
      </div>
      <div class="space-y-2 max-w-md">
        <h2 class="text-xl font-bold text-[var(--text-main)]">{{ t('auditCharter.sopList.emptyTitle') }}</h2>
        <p class="text-sm text-[var(--text-muted)] leading-relaxed">
          {{ t('auditCharter.sopList.emptyDesc') }}
        </p>
      </div>
      <UButton
        v-if="canManageCharter"
        :label="t('auditCharter.sopList.addSop')"
        @click="openAddModal"
        color="primary"
        size="lg"
        class="rounded-xl px-6 py-3 font-semibold transition-all duration-200"
        icon="i-lucide-plus"
      />
    </div>

    <!-- Data State -->
    <div v-else class="space-y-4">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-2xl font-bold text-gray-900">{{ t('auditCharter.sopList.title') }}</h2>
          <p class="text-sm text-gray-500">{{ t('auditCharter.sopList.subtitle') }}</p>
        </div>
        <UButton
          v-if="canManageCharter"
          :label="t('auditCharter.sopList.addSopShort')"
          @click="openAddModal"
          color="primary"
          icon="i-lucide-plus"
        />
      </div>

      <TableEntities
        :data="tableData"
        :columns="columns"
        :loading="store.loading"
        :server-side="true"
        :total="store.pagination.total"
        :items-per-page="store.pagination.page_size"
        :page="store.pagination.page"
        table-layout="fixed"
        :empty-state="{
          icon: 'i-lucide-file-text',
          label: t('auditCharter.sopList.emptyTable')
        }"
        class="w-full"
        @update:page="(p) => store.fetchSops(p)"
        @update:items-per-page="(size) => store.setPageSize(size)"
      >
        <!-- No slot -->
        <template #no-cell="{ row }">
          <span class="font-medium text-[var(--text-muted)]">{{ row.original.no }}</span>
        </template>

        <!-- Name slot -->
        <template #name-cell="{ row }">
          <ReadMoreText
            :text="row.original.name"
            :max-length="60"
            text-class="font-semibold text-[var(--text-main)]"
          />
        </template>

        <!-- Parent Guideline Name slot -->
        <template #guideline_name-cell="{ row }">
          <ReadMoreText
            :text="row.original.guideline?.name || '-'"
            :max-length="50"
            text-class="text-[var(--text-main)] font-medium"
          />
        </template>

        <!-- Status slot -->
        <template #status-cell="{ row }">
          <UBadge
            :color="row.original.status === 'Aktif' ? 'success' : 'warning'"
            variant="subtle"
            class="rounded font-semibold"
          >
            {{ translateStatus(row.original.status) }}
          </UBadge>
        </template>

        <!-- Effective date slot -->
        <template #effective_date-cell="{ row }">
          <span class="font-medium text-[var(--text-main)]">{{
            formatMonthYear(row.original.effective_date)
          }}</span>
        </template>

        <!-- Actions slot -->
        <template #actions-cell="{ row }">
          <div class="flex justify-center items-center gap-1">
            <UTooltip :text="t('auditCharter.tooltips.viewSop')">
            <UButton
              v-if="row.original.file_url && row.original.file_url !== '#'"
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              @click="openFile(row.original.file_url)"
            />
            </UTooltip>
            <UTooltip :text="t('auditCharter.tooltips.editSop')">
            <UButton
              v-if="canManageCharter"
              size="md"
              color="warning"
              variant="ghost"
              icon="i-lucide-edit"
              @click="store.handleEdit(row.original)"
            />
            </UTooltip>
            <UTooltip :text="t('auditCharter.tooltips.deleteSop')">
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useSopStore } from '~/stores/sop'
import { useGuidelineStore } from '~/stores/guideline'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import TableEntities from '~/components/shared/TableEntities.vue'
import { useGlobalModalStore } from '~/stores/global-modal'
import ReadMoreText from '~/components/shared/ReadMoreText.vue'

const { t, locale } = useI18n()
const store = useSopStore()
const guidelineStore = useGuidelineStore()
const { canManageCharter } = useRbac()

const columns = computed(() => store.columns)

const tableData = computed(() => {
  return store.sops.map((item, index) => ({
    ...item,
    no: (store.pagination.page - 1) * store.pagination.page_size + index + 1
  }))
})

const translateStatus = (status: string): string => {
  if (status === 'Aktif') return t('auditCharter.sopList.statusActive')
  if (status === 'Sedang Diperbarui') return t('auditCharter.sopList.statusUnderReview')
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

const openAddModal = async () => {
  await guidelineStore.fetchGuidelines()
  store.showModal = true
}

const openFile = (fileUrl: string) => {
  if (!fileUrl || fileUrl === '#') return
  if (fileUrl.startsWith('http://') || fileUrl.startsWith('https://')) {
    window.open(fileUrl, '_blank')
    return
  }
  const baseUrl = getAuditServiceBaseUrl()
  const finalUrl = fileUrl.startsWith('/') ? `${baseUrl.replace(/\/api\/v1$/, '')}${fileUrl}` : `${baseUrl}/${fileUrl}`
  window.open(finalUrl, '_blank')
}

const confirmDelete = async (item: any) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('auditCharter.sopList.deleteConfirm', { name: item.name }) })) {
    await store.deleteSop(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchSops()
})
</script>