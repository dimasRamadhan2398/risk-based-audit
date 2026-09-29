<template>
  <UModal 
    v-model:open="store.isOpen"
    :ui="{
      content: 'sm:max-w-lg w-[calc(100vw-2rem)] sm:w-full max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-4 sm:p-6 space-y-4 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #header>
      <div class="flex items-center justify-between w-full">
        <div class="flex items-center gap-3 min-w-0">
          <div v-if="store.options?.type === 'submit'" class="flex-shrink-0 w-8 h-8 rounded-full bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center text-primary-500 font-black text-lg leading-none">
            ?
          </div>
          <div v-else class="flex-shrink-0 w-8 h-8 rounded-full bg-error-50 dark:bg-error-900/20 flex items-center justify-center text-error-500 font-black text-lg leading-none">
            !
          </div>
          <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate">
            {{ description }}
          </h3>
        </div>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1 shrink-0" @click="store.resolve(false)" />
      </div>
    </template>

    <template #body>
      <div v-if="store.options?.type === 'submit'">
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Apakah Anda yakin ingin menyimpan data ini?</p>
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Data yang disimpan akan direkam ke dalam sistem.</p>
      </div>
      <div v-else>
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Apakah Anda yakin ingin menghapus data ini?</p>
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">Tindakan ini permanen dan tidak dapat dibatalkan.</p>
      </div>
    </template>

    <template #footer>
      <div class="flex flex-col-reverse sm:flex-row justify-end gap-2 sm:gap-3 w-full">
        <UButton label="Batal" color="neutral" variant="ghost" class="w-full sm:w-auto justify-center" @click="store.resolve(false)" />
        <UButton v-if="store.options?.type === 'submit'" label="Ya, Simpan Data" color="primary" class="w-full sm:w-auto justify-center" @click="store.resolve(true)" />
        <UButton v-else label="Ya, Hapus Data" color="error" class="w-full sm:w-auto justify-center" @click="store.resolve(true)" />
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
  if (store.options?.type === 'submit') {
    if (store.options?.itemName) return `Simpan "${store.options.itemName}"?`
    return 'Simpan Data?'
  }
  if (store.options?.itemName) return `Hapus "${store.options.itemName}"?`
  return 'Hapus Data?'
})

const description = computed(() => {
  if (store.options?.description) return store.options.description
  if (store.options?.type === 'submit') return 'Apakah Anda yakin ingin menyimpan data ini?'
  return 'Apakah Anda yakin ingin menghapus data ini?<br />Tindakan ini permanen dan tidak dapat dibatalkan.'
})
</script>
