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
        size="lg"
        class="rounded-xl px-6 py-3 font-semibold transition-all duration-200"
        icon="i-lucide-plus"
      />
    </div>

    <!-- Data State -->
    <div v-else class="space-y-4">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-2xl font-bold text-gray-900">{{ t('auditCharter.guideline.title') }}</h2>
          <p class="text-sm text-gray-500">{{ t('auditCharter.guideline.subtitle') }}</p>
        </div>
        <UButton
          v-if="canManageCharter"
          :label="t('auditCharter.guideline.addGuidelineShort')"
          @click="() => { store.showModal = true }"
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
          icon: 'i-lucide-book-open',
          label: t('auditCharter.guideline.emptyTable')
        }"
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
              :color="row.original.status === 'Aktif' ? 'success' : 'warning'"
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
            <UTooltip text="Lihat Pedoman">
            <UButton
              v-if="row.original.file_url && row.original.file_url !== '#'"
              icon="i-lucide-eye"
              color="primary"
              variant="ghost"
              size="md"
              @click="openFile(row.original.file_url)"
            />
            </UTooltip>
            <UTooltip text="Edit Pedoman">
            <UButton
              v-if="canManageCharter"
              size="md"
              color="primary"
              variant="ghost"
              icon="i-lucide-edit"
              @click="store.handleEdit(row.original)"
            />
            </UTooltip>
            <UTooltip text="Hapus Pedoman">
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
import { useGuidelineStore } from '~/stores/guideline'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import TableEntities from '~/components/shared/TableEntities.vue'
import { useGlobalModalStore } from '~/stores/global-modal'
import ReadMoreText from '~/components/shared/ReadMoreText.vue'

const { t, locale } = useI18n()
const store = useGuidelineStore()
const { canManageCharter } = useRbac()

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

const openFile = (fileUrl: string) => {
  if (!fileUrl || fileUrl === '#') return
  if (fileUrl.startsWith('http://') || fileUrl.startsWith('https://')) {
    window.open(fileUrl, '_blank')
    return
  }
  const config = useRuntimeConfig()
  const baseUrl = config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  const finalUrl = fileUrl.startsWith('/') ? `${baseUrl.replace(/\/api\/v1$/, '')}${fileUrl}` : `${baseUrl}/${fileUrl}`
  window.open(finalUrl, '_blank')
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