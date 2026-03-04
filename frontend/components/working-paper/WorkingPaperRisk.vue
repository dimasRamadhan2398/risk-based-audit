<template>
    <UCard class="mt-4 shadow-sm p-8">
        <div class="justify-between items-center mb-10">
        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Risk & Control</h2>
        </div>

        <div class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Resiko (Dari Risk Register) <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.form.resiko" :items="store.options.resiko" placeholder="Pilih Resiko" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Taksonomi <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.form.taksonomi" disabled placeholder="Otomatis terisi saat memilih resiko" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tingkat Resiko <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.form.tingkatResiko" disabled placeholder="Otomatis terisi saat memilih resiko" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Deskripsi Kontrol yang Diuji</label>
            <UTextarea class="md:col-span-3" v-model="store.form.deskripsiKontrol" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
        </div>
        </div>
        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
        <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF02()" />
        </div>
    </UCard>

    <UCard class="shadow-sm mt-10">
        <div class="p-4 border-b border-gray-300 dark:border-gray-600">
        <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Risk & Control Profile</h3>
        </div>
        <UTable :data="store.savedF02" :columns="store.columnsF02" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" />
    </UCard>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>