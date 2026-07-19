<template>
  <div>
    <!-- Empty State -->
    <div
      v-if="store.guidelines.length === 0"
      class="flex flex-col items-center justify-center p-12 bg-[var(--bg-surface)] border border-[var(--border-main)] rounded-2xl text-center space-y-6 shadow-sm my-4"
    >
      <div class="w-16 h-16 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500">
        <UIcon name="i-lucide-book-open" class="w-8 h-8" />
      </div>
      <div class="space-y-2 max-w-md">
        <h2 class="text-xl font-bold text-[var(--text-main)]">Belum Ada Pedoman Audit</h2>
        <p class="text-sm text-[var(--text-muted)] leading-relaxed">
          Pedoman Audit mendefinisikan aturan dan standar pelaksanaan audit di perusahaan Anda.
        </p>
      </div>
      <UButton
        label="Tambah Pedoman Audit"
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
          <h2 class="text-2xl font-bold text-gray-900">Pedoman Audit</h2>
          <p class="text-sm text-gray-500">Daftar seluruh Pedoman Audit yang berlaku di perusahaan</p>
        </div>
        <UButton
          label="Tambah Pedoman"
          @click="() => { store.showModal = true }"
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
          <!-- Status slot -->
          <template #status-data="{ row }">
            <UBadge
              :color="row.original.status === 'Aktif' ? 'success' : 'warning'"
              variant="subtle"
              class="rounded font-semibold"
            >
              {{ row.original.status }}
            </UBadge>
          </template>

          <!-- Effective date slot -->
          <template #effective_date-data="{ row }">
            <span class="font-medium text-gray-800">{{
              formatMonthYearIndonesian(row.original.effective_date)
            }}</span>
          </template>

          <!-- File / View Dokumen slot -->
          <template #file_name-data="{ row }">
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
          <template #actions-data="{ row }">
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
import { useGuidelineStore } from '~/stores/guideline'

const store = useGuidelineStore()

const columns = [
  { accessorKey: 'no', header: 'No' },
  { accessorKey: 'name', header: 'Nama Pedoman' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'effective_date', header: 'Mulai Berlaku' },
  { accessorKey: 'file_name', header: 'View Dokumen' },
  { accessorKey: 'actions', header: '' }
]

const tableData = computed(() => {
  return store.guidelines.map((item, index) => ({
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

const confirmDelete = async (item: any) => {
  if (confirm(`Apakah Anda yakin ingin menghapus Pedoman "${item.name}"?`)) {
    await store.deleteGuideline(item.id || '')
  }
}

onMounted(async () => {
  await store.fetchGuidelines()
})
</script>
