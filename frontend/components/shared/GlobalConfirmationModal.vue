<template>
  <UModal v-model:open="store.isOpen">
    <template #header>
      <div class="flex items-center justify-between w-full">
        <div class="flex items-center gap-3">
          <div class="flex-shrink-0 w-8 h-8 rounded-full bg-error-50 dark:bg-error-900/20 flex items-center justify-center text-error-500 font-black text-lg leading-none">
            !
          </div>
          <h3 class="text-lg font-bold text-gray-900 dark:text-white">
            {{ description }}
          </h3>
        </div>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="store.resolve(false)" />
      </div>
    </template>

    <template #body>
      <p class="text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Apakah Anda yakin ingin menghapus data ini?</p>
      <p class="text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Tindakan ini permanen dan tidak dapat dibatalkan.</p>
    </template>

    <template #footer>
      <div class="flex justify-end gap-3 w-full">
        <UButton label="Batal" color="neutral" variant="ghost" @click="store.resolve(false)" />
        <UButton label="Ya, Hapus Data" color="error" @click="store.resolve(true)" />
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useGlobalModalStore } from '~/stores/global-modal'

const store = useGlobalModalStore()

const title = computed(() => {
  if (store.options?.title) return store.options.title
  if (store.options?.itemName) return `Hapus "${store.options.itemName}"?`
  return 'Hapus Data?'
})

const description = computed(() => {
  if (store.options?.description) return store.options.description
  return 'Apakah Anda yakin ingin menghapus data ini?<br />Tindakan ini permanen dan tidak dapat dibatalkan.'
})
</script>
