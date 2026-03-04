<template>
    <div class="mt-4 space-y-6">
        <UCard class="mt-4 shadow-sm p-8">
        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-10">Populasi & Sampel</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-full">
            <UFormField label="Total Populasi" required>
            <UInput type="number" v-model="store.form.populasi" placeholder="Ex: 100" class="w-full"/>
            </UFormField>
            <UFormField label="Jumlah Sampel yang Diuji" required>
            <UInput type="number" v-model="store.form.jmlSampel" placeholder="Ex: 10" class="w-full"/>
            </UFormField>
        </div>

        <div class="justify-between items-center mt-10">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Matriks Pengujian Kontrol</h2>
        </div>
        
        <div class="space-y-6 mt-6">
            <div v-for="(sampel, index) in store.form.samples" :key="sampel.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 relative bg-white dark:bg-gray-800">
            <h3 class="text-lg font-bold mb-4">Sampel {{ index + 1 }}</h3>
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Nama Dokumen <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="sampel.dokumen" placeholder="Ex: PO-2026-001" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 1 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l1" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 2 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l2" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 3 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l3" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center mt-2">
                <label class="font-semibold text-sm">Status</label>
                <div class="md:col-span-3 flex items-center gap-2">
                    <div class="w-4 h-4 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                    <span class="font-bold">{{ store.checkSampleStatus(sampel) ? 'Efektif' : 'Tidak Efektif' }}</span>
                </div>
                </div>
            </div>

            <div class="mt-4 flex justify-end">
                <UButton label="Hapus Sampel" color="error" variant="ghost" class="font-bold" @click="store.removeSample(index)" />
            </div>
            </div>
            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Sampel" @click="store.addSample" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm mt-2">Kesimpulan</label>
            <UTextarea class="md:col-span-3" v-model="store.form.kesimpulan" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
        </div>

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF03()" />
        </div>
        </UCard>

        <UCard class="shadow-sm mt-10">
        <div class="p-4 border-b border-gray-300 dark:border-gray-600">
            <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Pengujian Kontrol</h3>
        </div>
        <UTable :data="store.savedF03" :columns="store.columnsF03" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
            <template #samples-cell="{ row }">
            <div class="flex flex-col gap-2 max-w-md">
                <div 
                v-for="s in row.original.samples" 
                :key="s.id" 
                class="text-sm p-2 border rounded bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700"
                >
                <div class="flex justify-between font-bold mb-1">
                    <span>{{ s.dokumen || 'Tanpa Nama Dokumen' }}</span>
                    <UBadge 
                    :color="store.checkSampleStatus(s) ? 'success' : 'error'" 
                    size="md" 
                    variant="subtle"
                    >
                    {{ store.checkSampleStatus(s) ? 'Efektif' : 'Tidak Efektif' }}
                    </UBadge>
                </div>
                <div class="text-sm text-gray-500 italic">
                    L1: {{ s.l1 || '-' }} | L2: {{ s.l2 || '-' }} | L3: {{ s.l3 || '-' }}
                </div>
                </div>
                
                <span v-if="!row.original.samples?.length" class="text-gray-400 italic text-xs">
                Tidak ada data sampel
                </span>
            </div>
            </template>
        </UTable>
        </UCard>
    </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>