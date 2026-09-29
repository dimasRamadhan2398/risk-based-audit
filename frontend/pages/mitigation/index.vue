<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-w-0">

    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 border-b border-gray-200 dark:border-gray-800 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Risk Mitigation Plan</h1>
        <p class="text-sm text-gray-600 dark:text-gray-400 mt-1"><strong>Risk ID:</strong> {{ riskProfileStore.getFormattedId(currentRisk) }} - {{ currentRiskName }}</p>
      </div>
      <UButton
        icon="i-heroicons-arrow-left"
        variant="ghost"
        color="neutral"
        label="Back to Detail"
        class="w-full sm:w-auto"
        @click="goBack"
      />
    </div>

    <MitigationTable :currentRiskId="currentRiskId" />

    <MitigationForm :currentRiskId="currentRiskId" />
    
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { ref, onMounted, computed } from 'vue'
import { useRiskProfileStore } from '~/stores/risk-profile'
import MitigationTable from '~/components/risk-mitigation/MitigationTable.vue'
import MitigationForm from '~/components/risk-mitigation/MitigationForm.vue'

const route = useRoute()
const router = useRouter()
const riskProfileStore = useRiskProfileStore()

// Mengambil ID risiko dari query parameter URL (?id=...)
const currentRiskId = computed(() => route.query.id as string)

/**
 * Mengambil data detail risiko dari store berdasarkan ID
 * Memanfaatkan getter getRiskById yang telah didefinisikan di store
 */
const currentRisk = computed(() => {
  if (!currentRiskId.value) return null
  return riskProfileStore.getRiskById(currentRiskId.value)
})

// Properti pembantu untuk menampilkan nama risiko
const currentRiskName = computed(() => currentRisk.value?.name || '-')

const goBack = () => {
  // Navigasi balik ke index sambil membawa ID risiko di query
  router.push({
    path: '/risk-profile',
    query: { openDetail: currentRiskId.value }
  })
}

// State untuk menyimpan data yang ditangkap dari URL
// const currentRisk = ref({
//   id: '',
//   name: '',
//   // ... data lain yang relevan
// })

// onMounted(() => {
//   // Tangkap query parameter saat halaman dimuat
//   if (route.query.id) {
//     currentRisk.value.id = route.query.id as string
//   }
//   if (route.query.name) {
//     currentRisk.value.name = route.query.name as string
//   }
  
// })
</script>
