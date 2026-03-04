<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm p-8">
        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Detail Temuan</h2>
        
        <div class="space-y-6 max-w-full">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Kondisi (Dampak di Lapangan)</label>
            <UTextarea class="md:col-span-3" v-model="store.form.kondisi" :rows="3" placeholder="Ex: Ditemukan dokumen PO tanpa tanda tangan Manager terkait pada tanggal..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Unggah bukti temuan/foto</label>
            
            <div class="md:col-span-3">
            <div 
                @click="store.triggerUpload"
                class="border-2 border-dashed border-gray-400 rounded-lg p-8 flex flex-col items-center justify-center text-center hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer relative"
                :class="{ 'border-gray-500 bg-gray-50': store.form.buktiFile }"
            >
                <input 
                type="file" 
                ref="fileInput" 
                class="hidden" 
                accept="image/png, image/jpeg" 
                @change="store.onFileChange"
                />

                <template v-if="!store.form.buktiFile">
                <UIcon name="i-heroicons-arrow-up-tray" class="w-8 h-8 text-gray-600 mb-2" />
                <span class="font-bold text-gray-800 dark:text-white">Upload Here or Drag and Drop</span>
                <span class="text-xs text-gray-500 mt-1">Jpg, Png (Max 10MB)</span>
                </template>

                <template v-else>
                <UIcon name="i-heroicons-document-check" class="w-8 h-8 text-gray-500 mb-2" />
                <span class="font-bold text-gray-700">{{ store.form.buktiFile.name }}</span>
                <span class="text-xs text-gray-500 mt-1">
                    {{ (store.form.buktiFile.size / 1024).toFixed(2) }} KB
                </span>
                <UButton 
                    label="Ganti File" 
                    variant="link" 
                    color="error" 
                    size="xs" 
                    class="mt-2" 
                    @click.stop="store.removeFile" 
                />
                </template>
            </div>
            </div>
        </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">Kriteria (Aturan/SOP Terkait) <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.form.kriteria" placeholder="Cari Peraturan Internal (Contoh: SOP Pengadaan Bab IV - Otorisasi)" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Dampak (Risk Impact)</label>
            <UTextarea class="md:col-span-3" v-model="store.form.dampak" :rows="3" placeholder="Ex: Potensi fraud atau pembelian fiktif yang dapat merugikan keuangan..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Root Cause Analysis</h2>
            <div v-for="(rca, index) in store.form.rcaList" :key="rca.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 bg-white dark:bg-gray-800">
            <h3 class="text-lg font-bold mb-4">Analisis {{ index + 1 }}</h3>
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Kategori <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="rca.method" :items="store.options.rcaMethod" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 1 <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="rca.w1" placeholder="Ex: Staf lupa meminta TTD Manager" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 2</label>
                <UInput class="md:col-span-3" v-model="rca.w2" placeholder="Ex: Karena staf terburu-buru mengejar kuota pengiriman" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 3</label>
                <UInput class="md:col-span-3" v-model="rca.w3" placeholder="-" />
                </div>
            </div>

            <div class="mt-4 flex justify-end">
                <UButton label="Hapus Kategori" color="error" variant="ghost" class="font-bold" @click="store.removeRCA(index)" />
            </div>
            </div>

            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Kategori" @click="store.addRCA" />
        </div>

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF04()" />
        </div>

        </UCard>

        <UCard class="shadow-sm mt-10">
        <div class="p-4 border-b border-gray-300 dark:border-gray-600">
            <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Temuan</h3>
        </div>
        <UTable :data="store.savedF04" :columns="store.columnsF04" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
            <template #buktiFile-cell="{ row }">
            <div v-if="row.original.buktiFile" class="flex items-center gap-1 text-blue-600">
                <UIcon name="i-heroicons-paper-clip" />
                <span class="text-md truncate max-w-[150px]">{{ row.original.buktiFile.name }}</span>
            </div>
            <span v-else class="text-gray-400">-</span>
            </template>

            <template #rcaList-cell="{ row }">
            <div class="space-y-2 py-2">
                <div v-for="rca in row.original.rcaList" :key="rca.id" class="text-[11px] leading-tight border-l-2 border-orange-400 pl-2">
                <div class="font-bold text-lg text-gray-700">{{ rca.method }}</div>
                <div class="text-gray-500 text-sm italic">
                    Why 1: {{ rca.w1 || '-' }} <br>
                    Why 2: {{ rca.w2 || '-' }} <br>
                    Why 3: {{ rca.w3 || '-' }}
                </div>
                </div>
            </div>
            </template>
        </UTable>
        </UCard>
    </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

const store = useWorkingPaperStore()

</script>