<template>
  <UModal 
    v-model:open="store.showModalF03" 
    :dismissible="false" 
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'p-0',
      body: 'p-0 flex-1 min-h-0 flex flex-col overflow-hidden',
      footer: 'p-0',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <UForm :schema="sampleSchema" :state="store.sampleForm" class="flex flex-col h-full max-h-[90vh] overflow-hidden" @submit.prevent="store.handleSubmitF03">
        <!-- Pinned Header -->
        <div class="px-5 sm:px-6 py-4 border-b border-gray-100 dark:border-gray-800 flex justify-between items-center shrink-0 bg-white dark:bg-gray-900">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-xl bg-primary-50 dark:bg-primary-950/60 flex items-center justify-center text-primary-500">
              <UIcon name="i-lucide-clipboard-check" class="w-5 h-5" />
            </div>
            <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">
              {{ store.isEditingF03 ? (t('auditFieldwork.sample.modalEdit') || 'Edit Sample') : (t('auditFieldwork.sample.modalAdd') || 'Sample Data Form') }}
            </h3>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            @click="() => { store.closeModalF03() }"
          />
        </div>

        <!-- Scrollable Body -->
        <div class="p-4 sm:p-6 space-y-6 overflow-y-auto flex-1 min-h-0">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-full">
            <UFormField label="Total Populasi" required name="population">
              <UInput v-model="store.sampleForm.population" placeholder="Contoh: 100 atau 100 Transaksi" class="w-full"/>
            </UFormField>
            <UFormField label="Jumlah Sampel yang Diuji" required name="sampleSize">
              <UInput type="number" v-model.number="store.sampleForm.sampleSize" placeholder="Ex: 10" class="w-full"/>
            </UFormField>
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">{{ t('auditFieldwork.sample.documentListLabel') || 'Daftar Sampel Dokumen' }}</h4>
              <UButton 
                color="primary" 
                icon="i-heroicons-plus" 
                variant="soft" 
                size="sm"
                :label="t('auditFieldwork.sample.addBtn') || 'Tambah Sampel Dokumen'" 
                @click="store.addSample()" 
              />
            </div>

            <div 
              v-for="(sampel, index) in store.sampleForm.samples" 
              :key="sampel.id || index" 
              class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 sm:p-5 relative bg-gray-50/50 dark:bg-gray-800/30 space-y-4"
            >
              <div class="flex justify-between items-center">
                <span class="text-sm font-bold text-gray-800 dark:text-gray-200">Dokumen {{ index + 1 }}</span>
                <UButton 
                  v-if="store.sampleForm.samples.length > 1"
                  icon="i-heroicons-trash" 
                  color="error" 
                  variant="ghost" 
                  size="sm"
                  @click="store.removeSample(index)" 
                />                
              </div>

              <!-- 1. Dropdown Dokumen (Sinkronisasi dengan Sample Data Audit Fieldwork) -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Dokumen" class="font-semibold text-sm" required />
                <div class="md:col-span-3">
                  <USelectMenu 
                    v-model="sampel.fieldworkDocument" 
                    :items="getAvailableFieldworkOptions(index)" 
                    value-key="value"
                    placeholder="Pilih Dokumen dari Sample Data Fieldwork" 
                    class="w-full"
                    @update:model-value="(val: any) => handleFieldworkDocumentChange(sampel, val)"
                  />
                </div>
              </div>

              <!-- 2. Input Field: Sample Dokumen -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Sample Dokumen" class="font-semibold text-sm" required />
                <div class="md:col-span-3">
                  <UInput 
                    v-model="sampel.document" 
                    placeholder="Contoh: INV/2026/03/001 - Faktur Pembelian PT Maju Mundur"
                    required
                    maxlength="100"
                    class="w-full"
                    @invalid="($event.target as any)?.setCustomValidity('Sample Dokumen maksimal 100 karakter dan wajib diisi')"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ (sampel.document || '').length }}/100
                  </div>
                </div>
              </div>

              <!-- 3. Langkah 1 sampai 3 dengan Pass/Fail dropdown di posisikan di bawahnya -->
              <!-- Langkah 1 -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-2 md:gap-4 items-start pt-3 border-t border-gray-100 dark:border-gray-800">
                <UFormField label="Langkah 1" class="font-semibold text-sm md:pt-2" />
                <div class="md:col-span-3 space-y-2">
                  <div>
                    <UInput 
                      v-model="sampel.step1" 
                      placeholder="Ketik prosedur / pengujian langkah 1 (maksimal 100 karakter)..." 
                      maxlength="100"
                      class="w-full"
                      @invalid="($event.target as any)?.setCustomValidity('Langkah 1 maksimal 100 karakter')"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ (sampel.step1 || '').length }}/100
                    </div>
                  </div>
                  <USelectMenu 
                    v-model="sampel.l1" 
                    :items="store.options.testResult" 
                    placeholder="Pilih Hasil Langkah 1 (Pass / Fail)" 
                    class="w-full"
                  />
                </div>
              </div>

              <!-- Langkah 2 -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-2 md:gap-4 items-start pt-3 border-t border-gray-100 dark:border-gray-800">
                <UFormField label="Langkah 2" class="font-semibold text-sm md:pt-2" />
                <div class="md:col-span-3 space-y-2">
                  <div>
                    <UInput 
                      v-model="sampel.step2" 
                      placeholder="Ketik prosedur / pengujian langkah 2 (maksimal 100 karakter)..." 
                      maxlength="100"
                      class="w-full"
                      @invalid="($event.target as any)?.setCustomValidity('Langkah 2 maksimal 100 karakter')"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ (sampel.step2 || '').length }}/100
                    </div>
                  </div>
                  <USelectMenu 
                    v-model="sampel.l2" 
                    :items="store.options.testResult" 
                    placeholder="Pilih Hasil Langkah 2 (Pass / Fail)" 
                    class="w-full"
                  />
                </div>
              </div>

              <!-- Langkah 3 -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-2 md:gap-4 items-start pt-3 border-t border-gray-100 dark:border-gray-800">
                <UFormField label="Langkah 3" class="font-semibold text-sm md:pt-2" />
                <div class="md:col-span-3 space-y-2">
                  <div>
                    <UInput 
                      v-model="sampel.step3" 
                      placeholder="Ketik prosedur / pengujian langkah 3 (maksimal 100 karakter)..." 
                      maxlength="100"
                      class="w-full"
                      @invalid="($event.target as any)?.setCustomValidity('Langkah 3 maksimal 100 karakter')"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ (sampel.step3 || '').length }}/100
                    </div>
                  </div>
                  <USelectMenu 
                    v-model="sampel.l3" 
                    :items="store.options.testResult" 
                    placeholder="Pilih Hasil Langkah 3 (Pass / Fail)" 
                    class="w-full"
                  />
                </div>
              </div>

              <!-- Status Evaluasi 3 Langkah -->
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center pt-3 border-t border-gray-100 dark:border-gray-800">
                <UFormField label="Status" class="font-semibold text-sm" />
                <div class="md:col-span-3 flex items-center gap-2">
                  <div class="w-3.5 h-3.5 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                  <span class="font-bold text-sm" :class="store.checkSampleStatus(sampel) ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'">
                    {{ store.checkSampleStatus(sampel) ? 'Efektif' : 'Tidak Efektif' }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <UFormField 
            label="Kesimpulan" 
            name="conclusion" 
            required
            class="grid grid-cols-1 md:grid-cols-4 items-start gap-2 md:gap-4" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 dark:text-gray-300 md:mt-2' }"
          >
            <UTextarea v-model="store.sampleForm.conclusion" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." class="w-full" />
          </UFormField>
        </div>

        <!-- Pinned Footer -->
        <div class="px-5 sm:px-6 py-3.5 border-t border-gray-100 dark:border-gray-800 flex justify-end items-center gap-3 shrink-0 bg-gray-50/70 dark:bg-gray-800/40">
          <UButton
            color="neutral"
            variant="ghost"
            :label="t('common.cancel')"
            @click="() => { store.closeModalF03() }"
          />
          <UButton 
            type="submit"
            :label="store.isEditingF03 ? t('common.updateData') : t('common.submit')" 
            color="primary"
            class="font-semibold px-4"
          />
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useI18n } from '~/composables/useI18n'
import { useWorkingPaperStore, sampleSchema } from '~/stores/working-paper'

const { t } = useI18n()
const store = useWorkingPaperStore()

const handleFieldworkDocumentChange = (sampel: any, val: any) => {
  sampel.fieldworkDocument = val
}

const getAvailableFieldworkOptions = (currentIndex: number) => {
  const chosenValues = new Set(
    store.sampleForm.samples
      .filter((_, idx) => idx !== currentIndex)
      .map(s => s.fieldworkDocument)
      .filter(Boolean)
  )
  return store.fieldworkSampleOptions.filter(opt => !chosenValues.has(opt.value))
}
</script>