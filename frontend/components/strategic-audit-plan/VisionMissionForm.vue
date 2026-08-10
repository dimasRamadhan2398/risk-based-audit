<template>
  <UModal
    v-model:open="store.isModalOpen"
    :title="t('strategicPlan.vmg.modalTitle')"
    :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }"
  >
    <template #body>
      <UForm :state="store.form" @submit.prevent="handleSubmit">
        <div class="space-y-6">
          <!-- Error Alert -->
          <UAlert
            v-if="store.errorMsg"
            :title="t('strategicPlan.vmg.validationError')"
            :description="store.errorMsg"
            color="error"
            variant="soft"
            icon="i-lucide-alert-circle"
            class="mb-2"
          />

          <!-- Visi Section -->
          <div class="space-y-3">
            <label class="block text-sm font-semibold text-[var(--text-main)]">
              {{ t('strategicPlan.vmg.vision') }} <span class="text-orange-500">*</span>
            </label>
            <div v-for="(visi, index) in store.form.visis" :key="'visi-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="store.form.visis[index]"
                :placeholder="t('strategicPlan.vmg.visionPlaceholder')"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.visis.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.visis.splice(index, 1) }"
                :aria-label="t('strategicPlan.vmg.deleteVision')"
              />
            </div>
            <UButton
              :label="t('strategicPlan.vmg.addVision')"
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
              {{ t('strategicPlan.vmg.mission') }} <span class="text-orange-500">*</span>
            </label>
            <div v-for="(misi, index) in store.form.misis" :key="'misi-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="store.form.misis[index]"
                :placeholder="t('strategicPlan.vmg.missionPlaceholder')"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.misis.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.misis.splice(index, 1) }"
                :aria-label="t('strategicPlan.vmg.deleteMission')"
              />
            </div>
            <UButton
              :label="t('strategicPlan.vmg.addMission')"
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
              {{ t('strategicPlan.vmg.goals') }}
            </h4>
            <p class="text-md text-[var(--text-muted)]">
              {{ t('strategicPlan.vmg.goalsSubtitle') }}
            </p>
            <div v-for="(goal, index) in store.form.goals" :key="'goal-' + index" class="flex gap-2 items-center">
              <UInput
                v-model="goal.goal_code"
                :placeholder="t('strategicPlan.vmg.goalCodePlaceholder')"
                class="w-24"
                required
              />
              <UInput
                v-model="goal.goal_name"
                :placeholder="t('strategicPlan.vmg.goalNamePlaceholder')"
                class="flex-1"
                required
              />
              <UButton
                v-if="store.form.goals.length > 1"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="() => { store.form.goals.splice(index, 1) }"
                :aria-label="t('strategicPlan.vmg.deleteGoal')"
              />
            </div>
            <UButton
              :label="t('strategicPlan.vmg.addGoal')"
              color="primary"
              variant="solid"
              size="sm"
              class="mt-1"
              @click="() => { store.form.goals.push({ goal_code: `G-00${store.form.goals.length + 1}`, goal_name: '' }) }"
            />
          </div>

          <!-- Rentang Tahun Section -->
          <div class="space-y-3 pt-4 border-t border-[var(--border-main)]">
            <p class="text-md text-[var(--text-muted)] font-semibold">
              {{ t('strategicPlan.vmg.yearRangeTitle') }}
            </p>
            <div class="flex gap-4 items-center">
              <div class="flex items-center gap-2">
                <span class="text-sm text-[var(--text-main)]">{{ t('strategicPlan.vmg.year') }}</span>
                <USelectMenu
                  v-model="store.form.yearStart"
                  :items="store.yearOptions"
                  value-key="value"
                  class="w-28"
                />
              </div>
              <span class="text-sm text-[var(--text-main)]">{{ t('strategicPlan.vmg.to') }}</span>
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
              :label="t('strategicPlan.vmg.cancel')"
              color="neutral"
              variant="ghost"
              @click="() => { store.isModalOpen = false }"
            />
            <UButton
              type="submit"
              :label="t('strategicPlan.vmg.saveData')"
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
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useVisionMissionGoalsStore()

const handleSubmit = async () => {
  const success = await store.saveVmg()
  if (success) {
    // Optionally trigger feedback/notification
  }
}
</script>
