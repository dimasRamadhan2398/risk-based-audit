<template>
    <div class="flex flex-col md:flex-row gap-4 mb-4  p-4 rounded-xl shadow-sm border border-gray-100 ">
      <UInput 
        v-model="store.searchCode" 
        icon="i-heroicons-magnifying-glass" 
        placeholder="Cari Code / Nama Aktivitas..." 
        class="flex-1"
        clearable
      />
      
      <USelectMenu 
        v-model="store.selectedDepartment" 
        :items="store.departmentOptions" 
        placeholder="Semua Departemen" 
        class="w-full md:w-60"
        clearable
      />
      
      <USelectMenu 
        v-model="store.selectedStatus" 
        :items="store.statusOptions" 
        placeholder="Semua Status" 
        class="w-full md:w-60"
        clearable
      />

      <UButton 
        v-if="store.searchCode || store.selectedDepartment || store.selectedStatus"
        icon="i-heroicons-x-mark" 
        color="neutral" 
        variant="ghost" 
        @click="store.clearFilters" 
        title="Reset Filters"
      />
    </div>
</template>

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
</script>