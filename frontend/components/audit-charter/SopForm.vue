<template>
  <UModal
    v-model:open="store.showModal"
    :dismissible="false"
    class="w-full sm:max-w-4xl bg-[var(--bg-main)] border-[var(--border-main)]"
  >
    <template #content>
      <UForm @submit.prevent="handleSubmit">
        <div class="px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-lg leading-6 font-bold" id="modal-title">
              {{ store.isEditing ? 'Edit Petunjuk Teknis / SOP' : 'Tambah Petunjuk Teknis / SOP Baru' }}
            </h3>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-lucide-x"
              @click="store.closeModal"
            />
          </div>

          <div class="space-y-4">
            <!-- SOP Name -->
            <UFormField
              label="Nama Petunjuk Teknis / SOP"
              class="block text-sm font-medium"
              size="lg"
              required
            >
              <UInput
                v-model="store.form.name"
                required
                type="text"
                maxlength="200"
                placeholder="e.g. SOP Penyusunan Program Kerja Audit Tahunan"
                class="mt-1 block w-full rounded-md"
                @invalid="($event.target as any)?.setCustomValidity('Nama SOP wajib diisi dan maksimal 200 karakter')"
                @input="($event.target as any)?.setCustomValidity('')"
              />
            </UFormField>

            <!-- Parent Guideline Selection -->
            <UFormField
              label="Pedoman Audit Induk"
              class="block text-sm font-medium"
              size="lg"
              required
            >
              <USelect
                v-model="store.form.guideline_id"
                required
                :items="guidelineOptions"
                placeholder="Pilih Pedoman Audit..."
                class="mt-1 block w-full rounded-md"
              />
            </UFormField>

            <div class="grid grid-cols-2 gap-4">
              <!-- Status -->
              <UFormField
                label="Status"
                class="block text-sm font-medium"
                size="lg"
                required
              >
                <USelect
                  v-model="store.form.status"
                  required
                  :items="['Aktif', 'Sedang Diperbarui']"
                  class="mt-1 block w-full rounded-md"
                />
              </UFormField>

              <!-- Effective Date -->
              <UFormField
                label="Mulai Berlaku"
                class="block text-sm font-medium"
                size="lg"
                required
              >
                <UInput
                  v-model="store.form.effective_date"
                  required
                  type="month"
                  class="mt-1 block w-full rounded-md"
                />
              </UFormField>
            </div>

            <!-- File Upload -->
            <UFormField
              label="Dokumen SOP / Juknis (PDF)"
              class="block text-sm font-medium"
              size="lg"
              :required="!store.isEditing"
            >
              <div class="mt-1 flex items-center gap-4">
                <input
                  type="file"
                  accept="application/pdf"
                  @change="store.handleFileChange"
                  class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100"
                  :required="!store.isEditing && !store.form.fileUrl"
                />
              </div>
              <p v-if="store.form.fileName" class="text-md text-gray-500 mt-1">
                File terpilih: <span class="font-semibold text-gray-700">{{ store.form.fileName }}</span>
              </p>
            </UFormField>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="p-3 bg-error-50 text-error-700 rounded-lg text-sm font-semibold">
              {{ store.errorMsg }}
            </div>
          </div>
        </div>

        <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse gap-3 rounded-b-lg">
          <UButton
            type="submit"
            :loading="store.loading"
            color="primary"
            variant="solid"
            size="md"
            class="w-full sm:w-auto font-bold"
          >
            {{ store.isEditing ? 'Simpan Perubahan' : 'Tambah SOP' }}
          </UButton>
          <UButton
            type="button"
            color="neutral"
            variant="outline"
            size="md"
            class="w-full sm:w-auto mt-2 sm:mt-0 font-bold"
            @click="store.closeModal"
          >
            Batal
          </UButton>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSopStore } from '~/stores/sop'
import { useGuidelineStore } from '~/stores/guideline'

const store = useSopStore()
const guidelineStore = useGuidelineStore()

const guidelineOptions = computed(() => {
  return guidelineStore.guidelines.map(g => ({
    value: g.id,
    label: g.name
  }))
})

const handleSubmit = async () => {
  if (store.isEditing) {
    await store.updateSop()
  } else {
    await store.addSop()
  }
}
</script>
