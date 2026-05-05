<template>
  <UModal v-model:open="store.isModalOpen" class="w-full sm:max-w-4xl">
    <div></div>
    <template #content>
    <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] z-10 sticky top-0 rounded-t-xl transition-colors duration-300">
        <h3 class="text-xl font-bold text-[var(--text-main)]">
          {{ store.isEditMode ? 'Edit' : 'Buat' }} Rencana Aktivitas Audit
        </h3>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-x-mark-20-solid"
          class="-my-1"
          @click="store.closeModal"
        />
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 space-y-6">
        <UForm :state="store.formState" @submit="store.savePlan" class="space-y-6">
          
          <!-- Informasi Dasar Perencanaan -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">Informasi Dasar Perencanaan</h3>
              </div>
            </template>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormGroup label="Judul Rencana Aktivitas Audit" required>
                <UInput v-model="store.formState.planTitle" placeholder="Rencana Audit Kantor Tahun 2024" />
              </UFormGroup>
              <UFormGroup label="Tahun Perencanaan" required>
                <UInput v-model="store.formState.planYear" />
              </UFormGroup>
              <UFormGroup label="Periode Perencanaan" required class="col-span-1 md:col-span-2">
                <div class="flex items-center gap-2">
                  <UInput v-model="store.formState.planPeriodStart" type="date" class="flex-1" />
                  <span class="text-gray-500">s/d</span>
                  <UInput v-model="store.formState.planPeriodEnd" type="date" class="flex-1" />
                </div>
              </UFormGroup>
              <UFormGroup label="Departemen/Unit Audit" required>
                <UInput v-model="store.formState.department" placeholder="Risk Audit IT" />
              </UFormGroup>
              <UFormGroup label="Dibuat Oleh" required>
                <UInput v-model="store.formState.createdBy" readonly />
              </UFormGroup>
              <UFormGroup label="Tanggal Pembuatan" required>
                <UInput v-model="store.formState.creationDate" type="date" readonly />
              </UFormGroup>
            </div>
          </UCard>

          <!-- Aktivitas Audit yang Direncanakan -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">Aktivitas Audit yang Direncanakan</h3>
                <UButton color="warning" variant="solid" icon="i-heroicons-plus" label="Tambah Aktivitas Audit" @click="store.addPlannedActivity" />
              </div>
            </template>

            <div class="space-y-6">
              <div v-for="(activity, index) in store.formState.plannedActivities" :key="activity.id" class="border border-[var(--border-main)] rounded-lg p-4 relative bg-[var(--bg-surface)] transition-all duration-300">
                <UButton
                  color="error"
                  variant="ghost"
                  icon="i-heroicons-trash"
                  class="absolute top-2 right-2"
                  @click="store.removePlannedActivity(index)"
                />
                <h4 class="font-medium mb-4">Aktivitas Audit #{{ index + 1 }}</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormGroup label="Nama Audit" required>
                    <UInput v-model="activity.auditName" placeholder="Audit IT Q1" />
                  </UFormGroup>
                  <UFormGroup label="Auditor" required>
                    <UInput v-model="activity.auditor" placeholder="Jamil" />
                  </UFormGroup>
                  <UFormGroup label="Area/Unit" class="col-span-1 md:col-span-2">
                    <UInput v-model="activity.area" placeholder="Unit IT" />
                  </UFormGroup>
                  <UFormGroup label="Status Pelaksanaan">
                    <USelect v-model="activity.executionStatus" :options="['Planned', 'In Progress', 'Completed']" />
                  </UFormGroup>
                  <UFormGroup label="Tingkat Risiko">
                    <USelect v-model="activity.riskLevel" :options="['Rendah', 'Sedang', 'Tinggi']" />
                  </UFormGroup>
                  <UFormGroup label="Durasi (hari)">
                    <UInput v-model="activity.duration" type="number" />
                  </UFormGroup>
                  <UFormGroup label="Prioritas">
                    <USelect v-model="activity.priority" :options="['Low', 'Medium', 'High']" />
                  </UFormGroup>
                  <UFormGroup label="Jumlah Auditor">
                    <UInput v-model="activity.numberOfAuditors" type="number" />
                  </UFormGroup>
                  <UFormGroup label="Estimasi Jadwal/Waktu">
                    <UInput v-model="activity.estimatedSchedule" type="date" />
                  </UFormGroup>
                </div>
              </div>
              
              <div v-if="store.formState.plannedActivities.length === 0" class="text-center text-gray-500 py-4">
                Belum ada aktivitas audit yang ditambahkan.
              </div>

              <div class="mt-4 border-t pt-4">
                <h4 class="font-medium mb-2">Ringkasan Aktivitas</h4>
                <div class="grid grid-cols-3 gap-4 text-center">
                  <div class="bg-red-50 text-red-600 rounded-lg p-2  ">
                    <div class="text-xs font-semibold">Risiko Tinggi</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter(a => a.riskLevel === 'Tinggi').length }}</div>
                  </div>
                  <div class="bg-yellow-50 text-yellow-600 rounded-lg p-2  ">
                    <div class="text-xs font-semibold">Risiko Sedang</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter(a => a.riskLevel === 'Sedang').length }}</div>
                  </div>
                  <div class="bg-green-50 text-green-600 rounded-lg p-2  ">
                    <div class="text-xs font-semibold">Risiko Rendah</div>
                    <div class="text-lg font-bold">{{ store.formState.plannedActivities.filter(a => a.riskLevel === 'Rendah').length }}</div>
                  </div>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Sumber Daya & Budget -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">Sumber Daya & Budget</h3>
                <UButton color="warning" variant="solid" icon="i-heroicons-plus" label="Tambah Auditor" @click="store.addResourceAuditor" />
              </div>
            </template>
            
            <div class="space-y-6">
              <div v-for="(auditor, index) in store.formState.resourceAuditors" :key="auditor.id" class="border border-[var(--border-main)] rounded-lg p-4 relative bg-[var(--bg-surface)] transition-all duration-300">
                 <UButton
                  color="error"
                  variant="ghost"
                  icon="i-heroicons-trash"
                  class="absolute top-2 right-2"
                  @click="store.removeResourceAuditor(index)"
                />
                <h4 class="font-medium mb-4">Auditor #{{ index + 1 }}</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormGroup label="Nama">
                    <UInput v-model="auditor.name" placeholder="Jamil" />
                  </UFormGroup>
                  <UFormGroup label="Posisi">
                    <UInput v-model="auditor.position" placeholder="IT Auditor" />
                  </UFormGroup>
                  <UFormGroup label="Departemen">
                    <UInput v-model="auditor.department" />
                  </UFormGroup>
                  <UFormGroup label="Ketersediaan">
                    <UInput v-model="auditor.availability" />
                  </UFormGroup>
                </div>
              </div>

              <div class="mt-6 border-t pt-4">
                <h4 class="font-medium mb-4">Perencanaan Anggaran</h4>
                <div class="grid grid-cols-1 gap-4">
                  <UFormGroup label="Total Estimasi Biaya Aktivitas">
                    <UInput v-model="store.formState.budget.totalEstimatedCost" type="number" />
                  </UFormGroup>
                  <UFormGroup label="Total Budget yang Dialokasikan">
                    <UInput v-model="store.formState.budget.totalAllocatedBudget" type="number" />
                  </UFormGroup>
                  <UFormGroup label="Catatan Anggaran">
                    <UTextarea v-model="store.formState.budget.budgetNotes" />
                  </UFormGroup>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Review & Persetujuan -->
          <UCard :ui="{ body: 'px-4 py-5 sm:p-6' }">
            <template #header>
              <div class="flex justify-between items-center">
                <h3 class="text-lg font-medium">Review & Persetujuan</h3>
              </div>
            </template>
            <div class="space-y-6">
              <div>
                <h4 class="font-medium mb-2">Dibuat Oleh</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormGroup label="Nama Pembuat" required>
                    <UInput v-model="store.formState.review.creatorName" />
                  </UFormGroup>
                  <UFormGroup label="Jabatan" required>
                    <UInput v-model="store.formState.review.creatorPosition" />
                  </UFormGroup>
                </div>
              </div>
              <div class="border-t pt-4">
                <h4 class="font-medium mb-2">Disetujui Oleh</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormGroup label="Nama Penyetuju" required>
                    <UInput v-model="store.formState.review.approverName" />
                  </UFormGroup>
                  <UFormGroup label="Jabatan" required>
                    <UInput v-model="store.formState.review.approverPosition" />
                  </UFormGroup>
                  <UFormGroup label="Tanggal Persetujuan" required>
                    <UInput v-model="store.formState.review.approvalDate" type="date" />
                  </UFormGroup>
                  <UFormGroup label="Catatan Tambahan" class="col-span-1 md:col-span-2">
                    <UTextarea v-model="store.formState.review.additionalNotes" />
                  </UFormGroup>
                </div>
              </div>
            </div>
          </UCard>

          <!-- Buttons -->
          <div class="flex justify-end gap-3 pb-6">
            <UButton label="Batal" color="neutral" variant="ghost" @click="store.closeModal" />
            <UButton type="submit" :label="store.isEditMode ? 'Simpan Perubahan' : 'Buat Rencana'" color="warning" />
          </div>

        </UForm>
      </div>
    </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useActivityPlanStore } from '~/stores/activity-plan'

const store = useActivityPlanStore()
</script>
