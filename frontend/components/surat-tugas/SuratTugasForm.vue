<template>
    <Teleport to="body">
      <div v-if="store.isModalOpen" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-white dark:bg-gray-900 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-hidden">
          
          <div class="flex justify-between items-center p-6 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <h2 class="text-xl font-bold text-gray-800 dark:text-white flex items-center gap-2">
              <UIcon name="i-heroicons-document-plus" class="w-6 h-6 text-orange-500" />
              Tambah Surat Tugas
            </h2>
            <button @click="store.closeModal" class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">
              <UIcon name="i-heroicons-x-mark" class="w-7 h-7" />
            </button>
          </div>

          <div class="p-8 overflow-y-auto space-y-6 flex-1">
            
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Judul Audit</label>
              <div class="md:col-span-3">
                <UInput v-model="store.form.judulAudit" placeholder="Contoh: Audit Operasional Keuangan" size="lg" class="w-full"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Dipimpin Oleh (PJ)</label>
              <div class="md:col-span-3">
                <UInput v-model="store.form.dipimpinOleh" placeholder="Nama Penanggung Jawab" size="lg" class="w-full"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Kategori Audit</label>
              <div class="md:col-span-3">
                <UFormField size="lg">
                    <select v-model="store.form.category" class="input-field bg-white rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                        <option value="Assurance">Assurance</option>
                        <option value="Special Audit">Special Audit</option>
                        <option value="Specific Reason">Specific Reason</option>
                        <option value="Consulting Services">Consulting Services</option>
                        <option value="Investigation">Investigation</option>
                        <option value="Quality Assurance Review">Quality Assurance Review</option>
                        <option value="Follow-Up Audit">Follow-up Audit</option>
                    </select>
                </UFormField>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Tahun Audit</label>
              <div class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.tahunAudit" type="date" size="lg" class="flex-1 w-full" />
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center border-y border-gray-100 dark:border-gray-800 py-4 my-6">
            <label class="font-bold text-gray-700 dark:text-gray-300">Tim Audit</label>
            <div class="md:col-span-3">
                <URadioGroup 
                orientation="horizontal" 
                variant="list" 
                default-value="System" 
                :items="store.options.timAudit"
                />
            </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Periode Audit</label>
              <div class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.periodeMulai" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                  <span class="font-bold text-gray-500">s/d</span>
                  <UInput v-model="store.form.periodeSelesai" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                </div>
                <p v-if="store.dateError" class="text-red-500 text-sm font-semibold mt-1">{{ store.dateError }}</p>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Unit Kerja</label>
              <div class="md:col-span-3">
                <USelectMenu v-model="store.form.unitKerja" :items="store.options.unitKerja" placeholder="Pilih Unit Kerja (Wajib)" size="lg" class="w-full" :popper="{ strategy: 'absolute' }"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Anggota Tim</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(anggota, index) in store.form.listAnggota" :key="index" class="flex items-center gap-2">
                  <UInput v-model="anggota.nama" placeholder="Nama Personil" class="flex-1" />
                  <USelectMenu v-model="anggota.peran" :items="store.options.peran" placeholder="Peran" class="w-1/3" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.listAnggota, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Tambah Personil" @click="store.addItem(store.form.listAnggota, { nama: '', peran: 'Anggota' })" />
                <p class="text-xs text-orange-600 dark:text-orange-400 font-semibold">* Sesuai template, disarankan minimal 3 anggota tim.</p>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Tujuan Audit</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(tujuan, index) in store.form.listTujuan" :key="index" class="flex items-start gap-2">
                  <span class="mt-2 font-bold text-gray-400">{{ index + 1 }}.</span>
                  <UTextarea v-model="store.form.listTujuan[index]" placeholder="Ketik tujuan audit..." class="flex-1" :rows="2" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.listTujuan, index)" class="mt-1" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Tambah Tujuan" @click="store.addItem(store.form.listTujuan, '')" />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Ruang Lingkup</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(ruang, index) in store.form.listRuangLingkup" :key="index" class="flex items-start gap-2">
                  <UInput v-model="store.form.listRuangLingkup[index]" placeholder="Ketik ruang lingkup..." class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.listRuangLingkup, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Tambah Ruang Lingkup" @click="store.addItem(store.form.listRuangLingkup, '')" />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Tembusan</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(tembusan, index) in store.form.listTembusan" :key="index" class="flex items-center gap-2">
                  <UInput v-model="store.form.listTembusan[index]" placeholder="Jabatan Tembusan (Contoh: Direktur Utama)" class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.listTembusan, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Tambah Tembusan" @click="store.addItem(store.form.listTembusan, '')" />
              </div>
            </div>

          </div>

          <div class="p-6 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 flex justify-end items-center gap-4">
            <button @click="store.closeModal" class="font-bold text-gray-500 hover:text-gray-700 px-4 py-2">
              Batal
            </button>
            <UButton 
                label="Simpan Surat Tugas" color="primary" size="lg" class="font-bold px-8 shadow-md" @click="store.handleSubmit" />
          </div>

        </div>
      </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useSuratTugasStore } from '~/stores/surat-tugas'
import { AuditCategory } from '~/types/audit';

const store = useSuratTugasStore()

</script>