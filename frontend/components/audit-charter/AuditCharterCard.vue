<template>
    <UCard
      v-if="store.activeCharter"
      class="relative group"
      variant="soft"
    >
      <div class="flex justify-between items-center">
        <h1 class="text-2xl font-bold text-gray-900">
          Active Audit Charter
        </h1>
        <UButton
          label="Add Charter"
          @click="store.showModal = true"
          color="primary"
          icon="add"
        >
        </UButton>
      </div>

      <div class="border-t border-gray-400  my-4"></div>

      <div class="flex justify-between items-start">
        <div>
          <UBadge
            label="CURRENTLY ACTIVE"
            class="px-2.5 py-0.5 mb-4 rounded inline-block"
            size="xl"
            color="success"
          >
          </UBadge>
          <h2 class="text-3xl font-bold text-[var(--text-main)] mb-4">
            {{ store.activeCharter.title }}
          </h2>
          <div class="flex items-center gap-4 text-sm text-gray-500 mb-6">
            <UBadge class="rounded inline-block" size="lg" color="error">
              v{{ store.activeCharter.version }}
            </UBadge>
            <UIcon name="calendar" size="md"></UIcon>
            <span class="flex items-center gap-1">{{
              store.activeCharter.date
            }}</span>
            <UIcon name="charter" size="md"></UIcon>
            <span class="flex items-center gap-1">{{
              store.activeCharter.fileName
            }}</span>
          </div>

          <div class="grid grid-cols-2 bg-[var(--bg-surface)] gap-8 mb-6 p-4 rounded-lg transition-colors duration-300">
            <div>
              <p class="text-xs text-[var(--text-muted)] uppercase tracking-wider">
                Uploaded By
              </p>
              <p class="font-medium text-[var(--text-main)]">
                {{ store.activeCharter.uploadedBy }}
              </p>
            </div>
            <div>
              <p class="text-xs text-[var(--text-muted)] uppercase tracking-wider">
                Approved By
              </p>
              <p class="font-medium text-[var(--text-main)]">
                {{ store.activeCharter.approvedBy }}
              </p>
            </div>
          </div>
        </div>
      </div>
      <div class="sm:flex sm:flex-row-reverse gap-4">
        <UButton icon="download" size="md" color="primary" variant="solid">
          Download
        </UButton>
        <UButton
          label="Edit"
          @click="store.handleEdit(store.activeCharter)"
          color="primary"
          icon="edit"
          variant="outline"
        >
        </UButton>
      </div>
    </UCard>

    <UCard v-else class="bg-secondary-50 border-l-4 border-secondary-400 p-4">
      <div class="flex justify-between items-center">
        <UIcon name="warning" size="lg" class="text-warning-600"></UIcon>
        <p class="text-secondary-700">
          Belum ada Audit Charter yang aktif. Silakan upload dokumen baru.
        </p>
        <UButton
          label="Add Charter"
          @click="store.showModal = true"
          color="primary"
          icon="add"
        >
        </UButton>
      </div>
    </UCard>

    <div>
      <UCard class=" relative group" variant="soft">
        <h3 class="text-lg font-semibold text-gray-700  mb-4">
          History Audit Charter
        </h3>
        <UTable
          :data="store.historyCharters"
          :columns="store.columns"
          :empty-state="{
            icon: 'i-heroicons-circle-stack-20-solid',
            label: 'Belum ada data rencana audit.',
          }"
          class="w-full text-sm text-left"
        >
          <template #version-data="{ row }">
            <span class="font-bold text-gray-800 ">{{
              row.original.version
            }}</span>
          </template>
          <template #title-data="{ row }">
            <span class="font-bold text-gray-800 ">{{
              row.original.title
            }}</span>
          </template>
          <template #date-data="{ row }">
            <span class="font-bold text-gray-800 ">{{
              row.original.date
            }}</span>
          </template>
          <template #approvedBy-data="{ row }">
            <span class="font-bold text-gray-800 ">{{
              row.original.approvedBy
            }}</span>
          </template>
          <template #uploadedBy-data="{ row }">
            <span class="font-bold text-gray-800 ">{{
              row.original.uploadedBy
            }}</span>
          </template>
          <template #fileName-cell="{ row }">
            <UButton icon="download" color="primary" size="md">
              {{ row.original.fileName }}
            </UButton>
          </template>
          <template #actions-cell="{ row }">
            <div class="flex justify-end">
              <UButton
                label="Edit"
                size="md"
                color="primary"
                variant="outline"
                icon="edit"
                @click="store.handleEdit(row.original)"
              />
            </div>
          </template>
        </UTable>
      </UCard>
    </div>
</template>

<script setup lang="ts">
import { useCharterStore } from '~/stores/charter'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useCharterStore()

onMounted(async () => {
  await store.fetchCharters()
})
</script>