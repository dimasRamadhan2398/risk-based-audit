<template>
  <div>
    <!-- Empty State -->
    <div
      v-if="store.sops.length === 0"
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

      <UCard class="relative overflow-hidden" variant="soft">
        <UTable
          :data="tableData"
          :columns="columns"
          class="w-full text-sm text-left"
        >
          <!-- Parent Guideline Name slot -->
          <template #guideline_name-cell="{ row }">
            <span class="text-gray-800 font-semibold">
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
            <span class="font-medium text-gray-800">{{
              formatMonthYearIndonesian(row.original.effective_date)
            }}</span>
          </template>

          <!-- File / View Dokumen slot -->
          <template #file_name-cell="{ row }">
            <UButton
              v-if="row.original.file_url && row.original.file_url !== '#'"
              :to="row.original.file_url"
              target="_blank"
              icon="i-lucide-external-link"
              color="primary"
              variant="link"
              size="sm"
              class="p-0 font-bold"
            >
              View Dokumen
            </UButton>
            <span v-else class="text-gray-400 italic">No File</span>
          </template>

          <!-- Actions slot -->
          <template #actions-cell="{ row }">
            <div class="flex justify-end gap-2">
              <UButton
                size="sm"
                color="primary"
                variant="outline"
                icon="i-lucide-edit"
                @click="store.handleEdit(row.original)"
              />
              <UButton
                size="sm"
                color="error"
                variant="outline"
                icon="i-lucide-trash"
                @click="confirmDelete(row.original)"
              />
            </div>
          </template>
        </UTable>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useSopStore } from '~/stores/sop'
import { useGuidelineStore } from '~/stores/guideline'

const store = useSopStore()
const guidelineStore = useGuidelineStore()

const columns = [
  { accessorKey: 'no', header: 'No' },
  { accessorKey: 'name', header: 'Nama Petunjuk Teknis' },
  { accessorKey: 'guideline_name', header: 'Nama Pedoman' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'effective_date', header: 'Mulai Berlaku' },
  { accessorKey: 'file_name', header: 'View Dokumen' },
  { accessorKey: 'actions', header: '' }
]

const tableData = computed(() => {
  return store.sops.map((item, index) => ({
    ...item,
    no: index + 1
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
