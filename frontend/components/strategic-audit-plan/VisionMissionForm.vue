<template>
  <UModal
    v-model:open="store.isModalOpen"
    title="Visi Misi"
    :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }"
  >
    <template #body>
      <UForm :state="store.form" @submit.prevent="handleSubmit">
        <div class="space-y-6">
          <!-- Error Alert -->
          <UAlert
            v-if="store.errorMsg"
            title="Validation Error"
            :description="store.errorMsg"
            color="error"
            variant="soft"
            icon="i-lucide-alert-circle"
            class="mb-2"
          />

          <!-- Visi Section -->
          <div class="space-y-3">
            <label class="block text-sm font-semibold text-[var(--text-main)]">
              Visi <span class="text-orange-500">*</span>
            </label>
            <div v-for="(visi, index) in store.form.visis" :key="'visi-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="store.form.visis[index]"
                placeholder="Ex: Visi Corporate"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.visis.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.visis.splice(index, 1) }"
                aria-label="Hapus Visi"
              />
            </div>
            <UButton
              label="Tambah Visi"
              color="primary"
              variant="solid"
              size="sm"
              class="mt-1"
              @click="() => { store.form.visis.push('') }"
            />
          </div>

          <!-- Misi Section -->
          <div class="space-y-3">
            <label class="block text-sm font-semibold text-[var(--text-main)]">
              Misi <span class="text-orange-500">*</span>
            </label>
            <div v-for="(misi, index) in store.form.misis" :key="'misi-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="store.form.misis[index]"
                placeholder="Ex: Misi Corporate"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.misis.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.misis.splice(index, 1) }"
                aria-label="Hapus Misi"
              />
            </div>
            <UButton
              label="Tambah Misi"
              color="primary"
              variant="solid"
              size="sm"
              class="mt-1"
              @click="() => { store.form.misis.push('') }"
            />
          </div>

          <!-- Goals Input Section -->
          <div class="space-y-3 pt-4 border-t border-[var(--border-main)]">
            <h4 class="text-sm font-bold text-[var(--text-main)] uppercase tracking-wide">
              Goals
            </h4>
            <p class="text-xs text-[var(--text-muted)]">
              Tambah goal-goal organisasi yang relevan:
            </p>
            <div v-for="(goal, index) in store.form.goals" :key="'goal-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="goal.goal_code"
                placeholder="G-001"
                class="w-24"
                required
              />
              <UInput
                v-model="goal.goal_name"
                placeholder="Nama Goal (Ex: Peningkatan Revenue)"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.goals.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.goals.splice(index, 1) }"
                aria-label="Hapus Goal"
              />
            </div>
            <UButton
              label="Tambah Goal"
              color="primary"
              variant="solid"
              size="sm"
              class="mt-1"
              @click="() => { store.form.goals.push({ goal_code: `G-00${store.form.goals.length + 1}`, goal_name: '' }) }"
            />
          </div>

          <!-- Rentang Tahun Section -->
          <div class="space-y-3 pt-4 border-t border-[var(--border-main)]">
            <p class="text-xs text-[var(--text-muted)] font-semibold">
              Pilih rentang tahun pencapaian target:
            </p>
            <div class="flex gap-4 items-center">
              <div class="flex items-center gap-2">
                <span class="text-sm text-[var(--text-main)]">Tahun</span>
                <USelectMenu
                  v-model="store.form.yearStart"
                  :items="store.yearOptions"
                  value-key="value"
                  class="w-28"
                />
              </div>
              <span class="text-sm text-[var(--text-main)]">s/d</span>
              <USelectMenu
                v-model="store.form.yearEnd"
                :items="store.yearOptions"
                value-key="value"
                class="w-28"
              />
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex justify-end gap-3 pt-4 border-t border-[var(--border-main)]">
            <UButton
              label="Batal"
              color="neutral"
              variant="ghost"
              @click="() => { store.isModalOpen = false }"
            />
            <UButton
              type="submit"
              label="Simpan Data"
              color="primary"
              variant="solid"
              :loading="store.saving"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'

const store = useVisionMissionGoalsStore()

const handleSubmit = async () => {
  const success = await store.saveVmg()
  if (success) {
    // Optionally trigger feedback/notification
  }
}
</script>
