<template>
  <div>
    <!-- Loading State -->
    <UCard v-if="store.loading" class="w-full border border-[var(--border-main)] bg-[var(--bg-main)]">
      <div class="space-y-4 py-4 animate-pulse">
        <div class="h-6 bg-gray-200 rounded w-1/4"></div>
        <div class="h-4 bg-gray-200 rounded w-3/4"></div>
        <div class="h-4 bg-gray-200 rounded w-5/6"></div>
      </div>
    </UCard>

    <!-- Empty State -->
    <UCard
      v-else-if="!store.activeVmg"
      class="w-full border border-[var(--border-main)] bg-[var(--bg-main)] shadow-sm"
    >
      <div class="flex flex-col items-center justify-center text-center py-10 px-4 space-y-5">
        <h3 class="text-xl font-bold text-[var(--text-main)]">
          {{ t('strategicPlan.vmg.title') }}
        </h3>
        <div class="space-y-1">
          <h4 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
            {{ t('strategicPlan.vmg.emptyTitle') }}
          </h4>
          <p class="text-sm text-[var(--text-muted)] w-full text-center">
            {{ t('strategicPlan.vmg.emptySubtitle') }}
          </p>
        </div>
        <UButton
          :label="t('strategicPlan.vmg.addVmg')"
          icon="i-lucide-plus"
          color="primary"
          variant="solid"
          size="lg"
          @click="store.openModal"
        />
      </div>
    </UCard>

    <!-- View State -->
    <UCard
      v-else
      class="w-full border border-[var(--border-main)] bg-[var(--bg-main)] shadow-sm"
    >
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-bold text-[var(--text-main)]">
            {{ t('strategicPlan.vmg.title') }}
          </h3>
          <UButton
            icon="i-lucide-pencil"
            color="neutral"
            variant="ghost"
            size="sm"
            @click="store.openModal"
            :aria-label="t('strategicPlan.vmg.editVmg')"
          />
        </div>
      </template>

      <div class="space-y-6">
        <!-- Target Period -->
        <div v-if="parsedPeriod" class="pb-3 border-b border-[var(--border-main)]">
          <p class="text-sm font-semibold text-[var(--text-main)]">
            {{ t('strategicPlan.vmg.targetPeriod', { start: parsedPeriod.start, end: parsedPeriod.end, duration: parsedPeriod.duration }) }}
          </p>
        </div>

        <!-- Visi Section -->
        <div class="space-y-2">
          <h4 class="text-sm font-bold uppercase tracking-wider text-[var(--text-muted)]">
            {{ t('strategicPlan.vmg.vision') }}
          </h4>
          <div class="space-y-2 pl-4 border-l-4 border-orange-500">
            <p
              v-for="(v, index) in store.activeVmg.vision.split('\n')"
              :key="'visi-' + index"
              class="text-base italic font-medium text-[var(--text-main)]"
            >
              ”{{ v }}”
            </p>
          </div>
        </div>

        <!-- Misi Section -->
        <div class="space-y-2">
          <h4 class="text-sm font-bold uppercase tracking-wider text-[var(--text-muted)]">
            {{ t('strategicPlan.vmg.mission') }}
          </h4>
          <div class="space-y-2 pl-1">
            <div
              v-for="(m, index) in store.activeVmg.mission.split('\n')"
              :key="'misi-' + index"
              class="text-sm text-[var(--text-main)] flex gap-2"
            >
              <span class="font-bold text-orange-500">{{ index + 1 }}.</span>
              <span>{{ cleanMissionLine(m) }}</span>
            </div>
          </div>
        </div>

        <!-- Goals Section -->
        <div class="space-y-2 pt-4 border-t border-[var(--border-main)]">
          <h4 class="text-sm font-bold uppercase tracking-wider text-[var(--text-muted)]">
            {{ t('strategicPlan.vmg.goals') }}
          </h4>
          <div v-if="store.activeVmg.goals && store.activeVmg.goals.length > 0" class="mt-3 space-y-2 pl-1">
            <div
              v-for="g in store.activeVmg.goals"
              :key="g.id"
              class="text-sm text-[var(--text-main)] flex gap-2"
            >
              <span class="font-bold text-orange-500">{{ g.goal_code }}</span>
              <span class="text-gray-400">|</span>
              <span class="font-medium">{{ g.goal_name }}</span>
            </div>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useVisionMissionGoalsStore()

// Clean leading numbers like "1. ", "2. " from seeded missions
const cleanMissionLine = (line: string) => {
  return line.replace(/^\d+\.\s*/, '').trim()
}

// Parse period range and calculate duration
const parsedPeriod = computed(() => {
  if (!store.activeVmg?.period) return null
  const parts = store.activeVmg.period.split('-')
  if (parts.length !== 2) return null
  const first = parts[0]
  const second = parts[1]
  if (!first || !second) return null
  const start = parseInt(first.trim(), 10)
  const end = parseInt(second.trim(), 10)
  if (isNaN(start) || isNaN(end)) return null
  const duration = end - start + 1
  return { start, end, duration }
})
</script>
