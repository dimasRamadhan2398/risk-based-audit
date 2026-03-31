<template>
    <UModal v-model:open="store.showViewModal" class="w-full sm:max-w-5xl">
      <div></div>

      <template #content>
      <div v-if="store.selectedPlan" class="relative bg-white dark:bg-gray-900 rounded-xl shadow-2xl flex flex-col max-h-[95vh] overflow-y-auto p-8">
              <template v-if="store.selectedPlan">
                  <div class="flex justify-between items-center">
                    <div class="flex items-center gap-4">
                      <UBadge color="primary" variant="subtle" class="text-xl font-bold text-gray-800 dark:text-white">
                        {{ store.selectedPlan.code }}  
                      </UBadge>
                      <div class="flex items-center gap-2">
                        <span class="w-4 h-4 rounded-full" :class="store.getStatusColor(store.selectedPlan.status)"></span>
                        <span class="font-bold text-lg text-gray-800 dark:text-white">{{ store.selectedPlan.status }}</span>
                      </div>
                    </div>
                    
                    <div class="flex items-center gap-2">
                      <UButton label="Hapus" color="error" icon="i-heroicons-trash" size="md" @click="store.handleDelete(store.selectedPlan.id)"/>
                      <UButton 
                        label="Edit Data" 
                        color="warning" 
                        icon="i-heroicons-pencil-square" 
                        size="md"
                        class="font-bold"
                        @click="store.handleEditFromView(store.selectedPlan)" 
                      />
                      <UIcon name="i-heroicons-x-mark" @click="store.closeViewModal" class="text-gray-400 hover:text-gray-600 text-3xl ml-2"></UIcon>
                    </div>
                  </div>
    

                <USeparator class="pt-6"/>

                <div class="space-y-6 pt-6">
                  
                  <UCard class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                    <template #header>
                      <h3 class="text-lg font-bold text-gray-800 dark:text-white">Status Pelaksanaan</h3>
                    </template>
                    <div class="flex items-center gap-8">
                      <span class="font-bold text-gray-700 dark:text-gray-300 w-32">Progress</span>
                      <div class="flex-1 max-w-xl">
                        <UProgress v-model="store.progressAudit" color="secondary" class="h-3" />
                      </div>
                      <span class="text-secondary-600 font-bold">50 %</span>
                    </div>
                  </UCard>

                  <UCard class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                    <template #header>
                      <h3 class="text-lg font-bold text-gray-800 dark:text-white">1. Detil Aktivitas</h3>
                    </template>
                    
                    <div class="space-y-6">
                      <div class="flex items-center pb-4 border-b border-gray-100 dark:border-gray-800">
                        <span class="font-bold text-gray-700 dark:text-gray-300 w-48">Kode Aktivitas</span>
                        <UBadge color="primary" variant="subtle" size="lg" class="font-black text-primary-600 dark:text-primary-400">{{ store.selectedPlan.code }}</UBadge>
                      </div>

                      <UCard 
                        v-for="(activity, index) in (store.selectedPlan.activities || [])" 
                        :key="index"
                        class="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg space-y-3 border border-gray-100 dark:border-gray-800"
                      >
                        <template #header>
                          <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 pb-2">
                            <span class="text-xs font-black uppercase text-gray-500 tracking-wider">
                              Sub-Aktivitas {{ Number(index) + 1 }}
                            </span>
                          </div>
                        </template>
                        
                        <div class="flex items-start">
                          <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Nama Aktivitas</span>
                          <span class="font-semibold text-gray-800 dark:text-white flex-1">{{ activity.name }}</span>
                        </div>

                        <div class="flex items-center">
                          <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Kategori</span>
                          <UBadge size="md" color="primary" variant="subtle" class="font-bold">
                            {{ activity.category }}
                          </UBadge>
                        </div>

                        <div class="flex items-center">
                          <span class="font-bold text-gray-600 dark:text-gray-400 w-44 text-sm">Departemen</span>
                          <span class="font-semibold text-gray-800 dark:text-white flex items-center gap-1">
                            {{ activity.department }}
                          </span>
                        </div>
                      </UCard>

                      <UCard v-if="!store.selectedPlan.activities?.length" class="text-center py-4 text-gray-400 italic">
                        Tidak ada detail aktivitas yang terdaftar.
                      </UCard>
                    </div>
                  </UCard>

                  <UCard class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                    <template #header>
                      <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">2. Timeline</h3>
                    </template>
                    <div class="space-y-4">
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Tahun</span><span class="font-semibold text-gray-800 dark:text-white">{{ store.selectedPlan.year }}</span></div>
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Distribusi Kuartal</span><span class="font-semibold text-gray-800 dark:text-white">{{ store.selectedPlan.quarters?.map((q: any) => `[ ${q} ]`).join(' ') || '-' }}</span></div>
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Pelaksanaan Bulan</span><span class="font-semibold text-gray-800 dark:text-white">{{ store.selectedPlan.selectedMonths?.slice().sort((a: number, b: number) => a - b).map((m: number) => `[ ${store.monthsList[m]} ]`).join(' ') || '-' }}</span></div>
                    </div>
                  </UCard>

                  <UCard class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                    <template #header>
                      <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">3. Auditor Resources</h3>
                    </template>
                    <div class="space-y-4">
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Supervisor</span><span class="font-semibold text-gray-800 dark:text-white">{{ store.getSupervisorName(store.selectedPlan.supervisorId) }}</span></div>
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Alokasi Waktu</span><span class="font-semibold text-gray-800 dark:text-white">👥 {{ store.selectedPlan.auditorCount }} Auditor ⏱️ {{ store.selectedPlan.daysPerAuditor }} Hari Durasi</span></div>
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Total Mandays</span><span class="font-semibold text-gray-800 dark:text-white">🔥 {{ store.selectedPlan.auditorCount * store.selectedPlan.daysPerAuditor }} Mandays</span></div>
                      <div class="flex items-center"><span class="font-bold text-gray-700 dark:text-gray-300 w-48">Estimasi Kapasitas</span><span class="font-semibold text-gray-800 dark:text-white">[🟢 Optimal (60-80%)] Total Load: 0.6% dari Kap Tahunan</span></div>
                    </div>
                  </UCard>

                  <UCard class="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                    <template #header>
                      <h3 class="text-lg font-bold text-gray-800 dark:text-white mb-6">4. Additional Notes</h3>
                    </template>
                    <p class="font-semibold text-gray-800 dark:text-white">{{ store.selectedPlan.notes || '-' }}</p>
                  </UCard>

                </div>
              </template>
      </div>
      </template>
    </UModal>
  
</template>

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
const open = ref(false)

defineShortcuts({
  o: () => open.value = !open.value
})
</script>