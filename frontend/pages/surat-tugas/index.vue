<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 dark:border-gray-700 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Daftar Surat Tugas Audit</h1>
      </div>
      <UButton 
        label="Buat Surat Tugas" 
        icon="i-heroicons-plus" 
        color="primary" 
        size="lg" 
        class="font-bold shadow-md"
        @click="store.openModal" 
      />
    </div>

    <SuratTugasTable />

    <SuratTugasForm />

  </div>
</template>

<script setup lang="ts">
import SuratTugasForm from '~/components/surat-tugas/SuratTugasForm.vue';
import SuratTugasTable from '~/components/surat-tugas/SuratTugasTable.vue';
import { useSuratTugasStore } from '~/stores/surat-tugas'

const store = useSuratTugasStore()

</script>