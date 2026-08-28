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
        <h2 class="text-xl font-bold text-[var(--text-main)]">Belum Ada Petunjuk Teknis / SOP</h2>
        <p class="text-sm text-[var(--text-muted)] leading-relaxed">
          Petunjuk Teknis / SOP merupakan penjabaran detail dari Pedoman Audit untuk panduan teknis auditor.
        </p>
      </div>
      <UButton
        label="Tambah Petunjuk Teknis / SOP"
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
          <h2 class="text-2xl font-bold text-gray-900">Petunjuk Teknis / SOP</h2>
          <p class="text-sm text-gray-500">Daftar seluruh Petunjuk Teknis dan Standar Operasional Prosedur pelaksanaan audit</p>
        </div>
        <UButton
          label="Tambah SOP / Juknis"
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
        :empty-state="{
          icon: 'i-lucide-file-text',
          label: 'Belum ada petunjuk teknis / SOP'
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
          <span class="font-semibold text-[var(--text-main)]">{{ row.original.name }}</span>
        </template>

        <!-- Parent Guideline Name slot -->
        <template #guideline_name-cell="{ row }">
          <span class="text-[var(--text-main)] font-medium">
            {{ row.original.guideline?.name || '-' }}
          </span>
        </template>

        <!-- Status slot -->
        <template #status-cell="{ row }">
          <UBadge
            :color="row.original.status === 'Aktif' ? 'success' : 'warning'"
            variant="subtle"
            class="rounded font-semibold"
          >
            {{ row.original.status }}
          </UBadge>
        </template>

        <!-- Effective date slot -->
        <template #effective_date-cell="{ row }">
          <span class="font-medium text-[var(--text-main)]">{{
            formatMonthYearIndonesian(row.original.effective_date)
          }}</span>
        </template>

        <!-- Actions slot -->
        <template #actions-cell="{ row }">
          <div class="flex justify-end gap-1">
            <UButton
              v-if="row.original.file_url && row.original.file_url !== '#'"
              :to="row.original.file_url"
              target="_blank"
              icon="i-lucide-eye"
              color="primary"
              variant="ghost"
              size="md"
            />
            <UButton
              size="md"
              color="primary"
              variant="ghost"
              icon="i-lucide-edit"
              @click="store.handleEdit(row.original)"
            />
            <UButton
              size="md"
              color="error"
              variant="ghost"
              icon="i-lucide-trash-2"
              @click="confirmDelete(row.original)"
            />
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
import TableEntities from '~/components/shared/TableEntities.vue'

const store = useSopStore()
const guidelineStore = useGuidelineStore()

const columns = [
  { accessorKey: 'no', header: 'No' },
  { accessorKey: 'name', header: 'Nama Petunjuk Teknis' },
  { accessorKey: 'guideline_name', header: 'Nama Pedoman' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'effective_date', header: 'Mulai Berlaku' },
  { accessorKey: 'actions', header: '' }
]

const tableData = computed(() => {
  return store.sops.map((item, index) => ({
    ...item,
    no: (store.pagination.page - 1) * store.pagination.page_size + index + 1
  }))
})

const formatMonthYearIndonesian = (val: string) => {
  if (!val) return '-'
  const parts = val.split('-')
  if (parts.length < 2) return val
  const [year, month] = parts
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  const mIndex = parseInt(month || '', 10) - 1
  if (mIndex >= 0 && mIndex < 12) {
    return `${months[mIndex]} ${year}`
  }
  return val
}

const openAddModal = async () => {
  await guidelineStore.fetchGuidelines()
  store.showModal = true
}

const confirmDelete = async (item: any) => {
  if (confirm(`Apakah Anda yakin ingin menghapus Petunjuk Teknis/SOP "${item.name}"?`)) {
    await store.deleteSop(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchSops()
})
</script>
