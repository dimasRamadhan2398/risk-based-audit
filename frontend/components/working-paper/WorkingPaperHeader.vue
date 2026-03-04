<template>
    <UCard class="mt-4 shadow-sm p-8">
        <div class="justify-between items-center mb-10">
        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Referensi Penugasan</h2>
        </div>

        <div class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Nomor Surat Tugas <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.form.suratTugas" :items="store.options.suratTugas" placeholder="Pilih Nomor Surat Tugas" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tujuan Audit <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.form.tujuanAudit" disabled placeholder="(Otomatis terisi saat mengisi surat tugas)" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Proses Bisnis <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.form.prosesBisnis" :items="store.options.prosesBisnis" placeholder="Pilih Proses Bisnis" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300">Periode Audit <span class="text-red-500">*</span></label>
            <div class="md:col-span-3">
            <UFormField :error="store.dateErrorMessage">
                <div class="flex items-center gap-4 w-full">
                <UInput 
                    type="date" 
                    v-model="store.form.periodeStart" 
                    icon="i-heroicons-calendar" 
                    class="w-full"
                    :color="store.isDateError ? 'error' : 'neutral'"
                />
                
                <span class="text-gray-500 font-bold whitespace-nowrap">s/d</span>
                
                <UInput 
                    type="date" 
                    v-model="store.form.periodeEnd" 
                    icon="i-heroicons-calendar" 
                    class="w-full"
                    :color="store.isDateError ? 'error' : 'neutral'"
                />
                </div>
            </UFormField>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Lokasi Cabang <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.form.lokasi" :items="store.options.lokasi" placeholder="Pilih lokasi" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tim Audit</label>
            <div class="md:col-span-3 space-y-4">

            <div v-for="(member, index) in store.form.teamMembers" :key="member.id" class="flex gap-2 items-center">
                <div class="grid grid-cols-2 gap-2 flex-1">
                <USelectMenu 
                    v-model="member.name" 
                    :items="store.getAvailableMembers(index)" 
                    placeholder="Pilih Nama Anggota"
                />
                <UInput 
                    v-model="member.role" 
                    placeholder="Jabatan (ex: Ketua Tim)" 
                />
                </div>

                <UButton 
                v-if="store.form.teamMembers.length > 1"
                icon="i-heroicons-trash" 
                color="error" 
                variant="ghost" 
                @click="store.removeTeamMember(index)" 
                />
            </div>

            <UButton 
                color="primary" 
                variant="soft"
                icon="i-heroicons-plus" 
                label="Tambah Anggota Audit" 
                @click="store.addTeamMember()"
            />
            </div>
        </div>
        </div>
        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
        <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF01()" />
        </div>
    </UCard>

    <UCard class="shadow-sm mt-10">
        <div class="p-4 border-b border-gray-300 dark:border-gray-600">
        <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Penugasan</h3>
        </div>
        <UTable :data="store.savedF01" :columns="store.columnsF01" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" >
        
        <template #teamMembers-cell="{ row }">
            <div class="flex flex-wrap gap-1">
            <UBadge 
                v-for="member in row.original.teamMembers"
                :key="member.id"
                color="neutral" 
                variant="subtle" 
                size="lg"
                class="flex flex-col items-start px-2 py-1"
            >
                <span class="font-bold text-primary-700">{{ member.name }}</span>
                <span class="text-[10px] opacity-70 italic">{{ member.role }}</span>
            </UBadge>
            <span v-if="!row.original.teamMembers?.length" class="text-gray-400">-</span>
            </div>
        </template>

        </UTable>
    </UCard>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>