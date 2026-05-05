<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6">

    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 border-b border-gray-200 pb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Rencana Mitigasi Resiko</h1>
        <p><strong>Risk ID:</strong> {{ currentRiskId }} - {{ currentRiskName }}</p>
      </div>
      <UButton
        icon="i-heroicons-arrow-left"
        variant="ghost"
        color="neutral"
        label="Kembali ke Detail"
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
import { useRiskProfileStore } from '~/stores/profile-risk'
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
const currentRiskName = computed(() => currentRisk.value?.risk_name || '-')

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

