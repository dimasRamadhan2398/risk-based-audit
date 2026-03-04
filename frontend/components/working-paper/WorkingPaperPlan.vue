<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm p-8">
        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Rekomendasi & Tanggapan</h2>
        
        <div class="space-y-6 max-w-full">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Rekomendasi Auditor (Solusi) <span class="text-red-500">*</span></label>
            <UTextarea class="md:col-span-3" v-model="store.form.rekomendasi" :rows="3" placeholder="Ex: Tim IT perlu menambahkan fitur hard-block..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Tanggapan Audite (Managemen Response)</label>
            <UTextarea class="md:col-span-3" v-model="store.form.tanggapan" :rows="3" placeholder="Ex: Kami setuju, update akan dilakukan di Q3..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Detail Rencana Aksi (Action Plan)</h2>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">Deskripsi Action <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.form.deskripsiAction" placeholder="Ex: Staf lupa meminta TTD Manager" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">PIC (Penanggung Jawab) <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.form.pic" icon="i-heroicons-magnifying-glass" :items="store.options.pic" placeholder="Cari Nama Karyawan Ex: Dimas - IT" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">Periode Perencanaan <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" type="date" v-model="store.form.periodAction" icon="i-heroicons-calendar" />
            </div>
        </div>

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF05()" />
        </div>
        </UCard>

        <UCard class="shadow-sm mt-10">
        <div class="p-4 border-b border-gray-300 dark:border-gray-600">
            <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Rekomendasi & Tanggapan</h3>
        </div>
        <UTable :data="store.savedF05" :columns="store.columnsF05" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" />
        </UCard>
    </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>