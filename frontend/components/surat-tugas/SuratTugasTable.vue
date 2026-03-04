<template>
    <UCard class="rounded-xl shadow-sm overflow-hidden" variant="soft">
      <UTable 
        :columns="store.columns" 
        :data="store.suratTugasList"
        :empty-state="{ icon: 'i-heroicons-document-text', label: 'Belum ada surat tugas yang dibuat. Klik tombol Buat Surat Tugas.' }"
        class="w-full text-sm"
      >
        <template #nomorSurat-cell="{ row }">
          <span class="font-bold text-orange-600 dark:text-orange-400">
            {{ row.original.nomorSurat }}
          </span>
        </template>

        <template #periode-cell="{ row }">
          <div class="flex items-center gap-1 text-gray-600 dark:text-gray-300">
            <UIcon name="i-heroicons-calendar" class="w-4 h-4" />
            <span>{{ row.original.periodeMulai }} s/d {{ row.original.periodeSelesai }}</span>
          </div>
        </template>

        <template #timAudit-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-800 dark:text-gray-200">{{ row.original.dipimpinOleh }} (PJ)</span>
            <span class="text-xs text-gray-500">{{ row.original.timAudit }}</span>
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 'Published' ? 'success' : 'neutral'" variant="subtle" size="lg" class="font-bold">
            {{ row.original.status }}
          </UBadge>
        </template>

      </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useSuratTugasStore } from '~/stores/surat-tugas'

const store = useSuratTugasStore()

</script>