<template>
  <UModal v-model:open="store.isViewModalOpen" class="sm:max-w-4xl bg-[var(--bg-main)] border-[var(--border-main)]">
    <template #content>
    <div class="relative rounded-xl shadow-2xl flex flex-col max-h-[90vh] border">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 sticky top-0 rounded-t-xl z-10">
        <h3 class="text-xl font-bold text-[var(--text-main)]">
          Detail Rencana Aktivitas Audit
        </h3>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-x-mark-20-solid"
          class="-my-1"
          @click="store.closeViewModal"
        />
      </div>

      <!-- Content -->
      <div v-if="store.selectedPlan" class="flex-1 overflow-y-auto p-4 space-y-6">
        <!-- Basic Info -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">Informasi Dasar</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">Judul Rencana:</strong> <span>{{ store.selectedPlan.planTitle }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Tahun Rencana:</strong> <span>{{ store.selectedPlan.planYear }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Periode Rencana:</strong> <span>{{ store.selectedPlan.planPeriodStart }} s/d {{ store.selectedPlan.planPeriodEnd }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Departemen:</strong> <UBadge variant="soft">{{ store.selectedPlan.department }}</UBadge></div>
                <div class="flex"><strong class="w-48 shrink-0">Dibuat oleh:</strong> <span>{{ store.selectedPlan.createdBy }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Tanggal Dibuat:</strong> <span>{{ new Date(store.selectedPlan.creationDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) }}</span></div>
            </div>
        </UCard>

        <!-- Planned Activities -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">Aktivitas Audit Direncanakan ({{ store.selectedPlan.plannedActivities.length }})</h4>
            </template>
            <UTable :data="store.selectedPlan.plannedActivities" :columns="plannedActivitiesColumns" />
        </UCard>

        <!-- Resource Auditors -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">Sumber Daya Auditor ({{ store.selectedPlan.resourceAuditors.length }})</h4>
            </template>
            <UTable :data="store.selectedPlan.resourceAuditors" :columns="resourceAuditorsColumns" />
        </UCard>

        <!-- Budget -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">Perencanaan Budget</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">Total Estimasi Biaya:</strong> <span>{{ new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(store.selectedPlan.budget.totalEstimatedCost) }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Total Alokasi Budget:</strong> <span>{{ new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(store.selectedPlan.budget.totalAllocatedBudget) }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Catatan Budget:</strong> <span class="flex-1">{{ store.selectedPlan.budget.budgetNotes || '-' }}</span></div>
            </div>
        </UCard>

        <!-- Review -->
        <UCard>
            <template #header>
                <h4 class="text-lg font-medium">Review & Persetujuan</h4>
            </template>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm">
                <div class="space-y-2">
                    <p class="font-semibold border-b pb-1 mb-2">Dibuat Oleh</p>
                    <div class="flex"><strong class="w-24 shrink-0">Nama:</strong> <span>{{ store.selectedPlan.review.creatorName }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">Posisi:</strong> <span>{{ store.selectedPlan.review.creatorPosition }}</span></div>
                </div>
                <div class="space-y-2">
                    <p class="font-semibold border-b pb-1 mb-2">Disetujui Oleh</p>
                    <div class="flex"><strong class="w-24 shrink-0">Nama:</strong> <span>{{ store.selectedPlan.review.approverName }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">Posisi:</strong> <span>{{ store.selectedPlan.review.approverPosition }}</span></div>
                    <div class="flex"><strong class="w-24 shrink-0">Tanggal:</strong> <span>{{ new Date(store.selectedPlan.review.approvalDate).toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) }}</span></div>
                </div>
                <div class="col-span-2 mt-2">
                    <strong class="font-semibold">Catatan Tambahan:</strong>
                    <p class="mt-1 text-gray-600 dark:text-gray-300">{{ store.selectedPlan.review.additionalNotes || '-' }}</p>
                </div>
            </div>
        </UCard>

        <!-- Attachments -->
        <UCard v-if="store.selectedPlan.attachmentCategory">
            <template #header>
                <h4 class="text-lg font-medium">Attachment</h4>
            </template>
            <div class="space-y-3 text-sm">
                <div class="flex"><strong class="w-48 shrink-0">Kategori File:</strong> <span>{{ store.selectedPlan.attachmentCategory }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Diupload Oleh:</strong> <span>{{ store.selectedPlan.attachmentUploadedBy }}</span></div>
                <div class="flex"><strong class="w-48 shrink-0">Tanggal Upload:</strong> <span>{{ store.selectedPlan.attachmentUploadDate }}</span></div>
                
                <div class="pt-2">
                  <strong class="block mb-2 text-gray-700">Uploaded Files:</strong>
                  <ul v-if="store.selectedPlan.attachments?.length" class="list-disc list-inside space-y-2">
                    <li v-for="(file, index) in store.selectedPlan.attachments" :key="index" class="flex items-center justify-between p-2 bg-gray-50 rounded-md">
                      <div class="flex items-center gap-2">
                        <UIcon name="i-heroicons-document-text" class="text-gray-500" />
                        <span class="font-semibold text-gray-800">{{ file.name }}</span>
                        <UBadge color="neutral" variant="soft" size="md">{{ file.size }}</UBadge>
                      </div>
                      <UButton :to="file.url" target="_blank" icon="i-heroicons-arrow-down-tray" size="sm" color="primary" variant="link" label="Download" />
                    </li>
                  </ul>
                  <p v-else class="text-gray-500 italic mt-2">No files attached.</p>
                </div>
            </div>
        </UCard>
      </div>
      <div v-else class="flex items-center justify-center h-32">
        <p class="text-gray-500">Memuat data...</p>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-3 p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] sticky bottom-0 rounded-b-xl">
        <UButton label="Tutup" color="neutral" variant="ghost" @click="store.closeViewModal" />
      </div>
    </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useActivityPlanStore } from '~/stores/activity-plan'

const store = useActivityPlanStore()

const plannedActivitiesColumns = [
  { accessorKey: 'auditName', header: 'Nama Audit' },
  { accessorKey: 'auditee', header: 'Auditee' },
  { accessorKey: 'category', header: 'Kategori' },
  { accessorKey: 'riskName', header: 'Risk Name' },
  { accessorKey: 'riskLevel', header: 'Level Risiko' },
  { accessorKey: 'duration', header: 'Durasi (hari)' },
  { accessorKey: 'priority', header: 'Prioritas' },
]

const resourceAuditorsColumns = [
  { accessorKey: 'name', header: 'Nama' },
  { accessorKey: 'position', header: 'Posisi' },
  { accessorKey: 'competence', header: 'Kompetensi' },
  { accessorKey: 'availability', header: 'Ketersediaan' },
]
</script>