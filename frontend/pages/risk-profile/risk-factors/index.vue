<template>
  <div class="space-y-8 p-6 max-w-full mx-auto">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[var(--border-main)] pb-5">
      <div>
        <h1 class="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-white font-space">
          {{ t('riskFactors.title') }}
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          {{ t('riskFactors.subtitle') }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <UButton
          to="/risk-profile"
          icon="i-lucide-arrow-left"
          color="neutral"
          variant="outline"
        >
          {{ t('riskFactors.backToHeatmap') }}
        </UButton>
      </div>
    </div>

    <!-- Alert Message -->
    <Transition name="fade">
      <UAlert
        v-if="alertMessage"
        :color="alertType === 'success' ? 'success' : 'error'"
        variant="solid"
        :title="alertType === 'success' ? t('common.success') : t('common.error')"
        :description="alertMessage"
        icon="i-lucide-info"
        class="shadow-md"
        closable
        @close="alertMessage = ''"
      />
    </Transition>

    <!-- Tabs Navigation -->
    <UTabs :items="tabItems" class="w-full">
      <!-- Tab 1: Corporate Weighting -->
      <template #weighting>
        <div class="space-y-6 mt-6">
          <!-- Validation Status Banner -->
          <div 
            class="flex items-center justify-between p-4 rounded-xl border shadow-sm transition-all duration-300"
            :class="isValidWeightSum ? 'bg-emerald-50/50 border-emerald-200 dark:bg-emerald-950/20 dark:border-emerald-800' : 'bg-rose-50/50 border-rose-200 dark:bg-rose-950/20 dark:border-rose-800'"
          >
            <div class="flex items-center gap-3">
              <UIcon 
                :name="isValidWeightSum ? 'i-lucide-check-circle-2' : 'i-lucide-alert-triangle'" 
                class="w-6 h-6"
                :class="isValidWeightSum ? 'text-emerald-500' : 'text-rose-500'"
              />
              <div>
                <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">
                  <span class="font-bold text-base" :class="isValidWeightSum ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ t('riskFactors.weighting.totalWeight', { total: totalWeight }) }}</span>
                </p>
                <p class="text-md text-slate-500 dark:text-slate-400">
                  {{ isValidWeightSum ? t('riskFactors.weighting.validSum') : t('riskFactors.weighting.invalidSum') }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <UBadge :color="isValidWeightSum ? 'success' : 'error'" size="md" variant="subtle">
                {{ isValidWeightSum ? t('riskFactors.weighting.badgeValid') : t('riskFactors.weighting.badgeInvalid') }}
              </UBadge>
              <UButton
                icon="i-lucide-save"
                color="primary"
                variant="solid"
                :loading="store.loading"
                :disabled="!isValidWeightSum"
                @click="saveChanges"
              >
                {{ t('riskFactors.weighting.saveWeights') }}
              </UButton>
            </div>
          </div>

          <div class="flex flex-col gap-8">
            <!-- Catalog Explorer -->
            <div class="space-y-6">
              <UCard class="shadow-sm border border-[var(--border-main)]">
                <template #header>
                  <div class="space-y-3">
                    <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                      {{ t('riskFactors.weighting.standardFactorsTitle') }}
                    </h2>
                    <UInput
                      v-model="searchQuery"
                      icon="i-lucide-search"
                      size="sm"
                      :placeholder="t('riskFactors.weighting.searchPlaceholder')"
                      color="neutral"
                      class="w-full"
                    />
                  </div>
                </template>

                <div class="max-h-[500px] overflow-y-auto pr-2 divide-y divide-slate-100 dark:divide-slate-800 space-y-2">
                  <div 
                    v-for="factor in filteredStandardFactors" 
                    :key="factor.id"
                    class="py-3 flex items-start justify-between gap-4 hover:bg-slate-50 dark:hover:bg-slate-800/40 p-2 rounded-lg transition-colors duration-200"
                  >
                    <div class="cursor-pointer flex-1" @click="openGuidelines(factor)">
                      <h3 class="text-md font-semibold text-slate-800 dark:text-slate-200">
                        {{ factor.name }}
                      </h3>
                      <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5 line-clamp-2">
                        {{ factor.description }}
                      </p>
                    </div>
                    <div class="flex items-center gap-2">
                      <UCheckbox
                        :model-value="isFactorSelected(factor.id)"
                        @update:model-value="toggleFactorSelection(factor)"
                      />
                    </div>
                  </div>
                  <div v-if="filteredStandardFactors.length === 0" class="text-center py-8 text-md text-slate-400">
                    {{ t('riskFactors.weighting.noMatchingFactors') }}
                  </div>
                </div>
              </UCard>
            </div>

            <!-- Corporate Weights Table -->
            <div>
              <UCard class="shadow-sm border border-[var(--border-main)]">
                <template #header>
                  <div class="flex items-center justify-between">
                    <div>
                      <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                        {{ t('riskFactors.weighting.selectedWeightsTitle') }}
                      </h2>
                    </div>
                    <UBadge color="neutral" variant="outline">
                      {{ t('riskFactors.weighting.activeFactorsCount', { count: selectedCorporateList.length }) }}
                    </UBadge>
                  </div>
                </template>

                <div v-if="selectedCorporateList.length > 0" class="overflow-x-auto">
                  <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                    <thead class="bg-slate-50 dark:bg-slate-800/50">
                      <tr>
                        <th scope="col" class="px-4 py-3 text-left font-semibold text-slate-700 dark:text-slate-300 w-12">{{ t('riskFactors.weighting.no') }}</th>
                        <th scope="col" class="px-4 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.weighting.riskFactor') }}</th>
                        <th scope="col" class="px-4 py-3 text-left font-semibold text-slate-700 dark:text-slate-300 w-32">{{ t('riskFactors.weighting.weightCol') }}</th>
                        <th scope="col" class="px-4 py-3 text-center font-semibold text-slate-700 dark:text-slate-300 w-20">{{ t('riskFactors.weighting.action') }}</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                      <tr 
                        v-for="(item, idx) in selectedCorporateList" 
                        :key="item.standard_risk_factor_id"
                        class="hover:bg-slate-50/50 dark:hover:bg-slate-800/20"
                      >
                        <td class="px-4 py-3 text-slate-500">{{ idx + 1 }}</td>
                        <td class="px-4 py-3">
                          <span 
                            class="font-medium text-slate-800 dark:text-slate-200 cursor-pointer hover:underline"
                            @click="openGuidelines(item.standard_risk_factor)"
                          >
                            {{ item.standard_risk_factor?.name }}
                          </span>
                          <p class="text-md text-slate-400 mt-0.5 line-clamp-1">
                            {{ item.standard_risk_factor?.description }}
                          </p>
                        </td>
                        <td class="px-4 py-3">
                          <UInput
                            v-model.number="item.weight"
                            type="number"
                            size="sm"
                            placeholder="0"
                            color="neutral"
                            class="w-24"
                            trailing-icon="i-lucide-percent"
                          />
                        </td>
                        <td class="px-4 py-3 text-center">
                          <UButton
                            icon="i-lucide-trash-2"
                            color="error"
                            variant="ghost"
                            size="md"
                            @click="removeFactorFromCorporate(item)"
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <div v-else class="text-center py-16 border-2 border-dashed border-slate-200 dark:border-slate-800 rounded-xl">
                  <UIcon name="i-lucide-alert-circle" class="w-8 h-8 text-slate-400 mx-auto" />
                  <h3 class="mt-2 text-sm font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.weighting.noFactorsTitle') }}</h3>
                  <p class="mt-1 text-md text-slate-500 dark:text-slate-400">
                    {{ t('riskFactors.weighting.noFactorsDesc') }}
                  </p>
                </div>
              </UCard>
            </div>
          </div>
        </div>
      </template>

      <!-- Tab 2: Scoring Workspace -->
      <template #scoring>
        <div class="mt-6">
          <UCard class="shadow-sm border border-[var(--border-main)]">
            <template #header>
              <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                  <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                    {{ t('riskFactors.scoring.title') }}
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    {{ t('riskFactors.scoring.subtitle') }}
                  </p>
                </div>
                
                <div class="flex items-center gap-3">
                  <div class="flex items-center gap-2">
                    <span class="text-md font-semibold text-slate-500">{{ t('riskFactors.scoring.yearLabel') }}</span>
                    <USelect
                      v-model.number="selectedYear"
                      :items="[2025, 2026, 2027, 2028]"
                      size="sm"
                      color="neutral"
                      class="w-20"
                      @update:model-value="fetchYearlyUniverse"
                    />
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-md font-semibold text-slate-500">{{ t('riskFactors.scoring.entityLabel') }}</span>
                    <USelectMenu
                      v-model="selectedEntityId"
                      :items="dropdownEntities"
                      label-key="label"
                      value-key="value"
                      :placeholder="t('riskFactors.scoring.selectEntityPlaceholder')"
                      class="w-56"
                      @update:model-value="selectEntityForScoringById"
                    >
                      <template #item="{ item }">
                        <span>{{ item.label }}</span>
                      </template>
                    </USelectMenu>
                  </div>
                </div>
              </div>
            </template>

            <!-- Scoring Area when activeYearlyEntity is selected -->
            <div v-if="activeYearlyEntity" class="space-y-6">
              <!-- Calculations overview -->
              <div class="grid grid-cols-3 gap-4 bg-slate-50 dark:bg-slate-850/50 p-4 rounded-xl border border-slate-100 dark:border-slate-800">
                <div class="text-center border-r border-slate-200 dark:border-slate-800">
                  <p class="text-md text-slate-500">{{ t('riskFactors.scoring.weightedScoreSum') }}</p>
                  <p class="text-lg font-black text-slate-800 dark:text-slate-100 mt-1">{{ totalWeightedScore.toFixed(2) }}</p>
                </div>
                <div class="text-center border-r border-slate-200 dark:border-slate-800">
                  <p class="text-md text-slate-500">{{ t('riskFactors.scoring.riskIndex') }}</p>
                  <p class="text-lg font-black text-slate-800 dark:text-slate-100 mt-1">{{ activeYearlyEntity.risk_index?.toFixed(1) }}%</p>
                </div>
                <div class="text-center">
                  <p class="text-md text-slate-500">{{ t('riskFactors.scoring.auditPriority') }}</p>
                  <UBadge :color="activeYearlyEntity.audit_priority ? 'success' : 'neutral'" variant="subtle" class="mt-1">
                    {{ activeYearlyEntity.audit_priority ? t('riskFactors.scoring.priorityYes') : t('riskFactors.scoring.priorityNo') }}
                  </UBadge>
                </div>
              </div>

              <!-- Scoring list -->
              <div class="mt-6 space-y-4 max-h-[380px] overflow-y-auto pr-2 divide-y divide-slate-100 dark:divide-slate-800">
                <div 
                  v-for="score in scoringRows" 
                  :key="score.corporate_risk_factor_id"
                  class="pt-3 flex flex-col md:flex-row md:items-center justify-between gap-4 p-2 rounded-lg hover:bg-slate-50/50"
                >
                  <div class="flex-1 cursor-pointer" @click="openGuidelines(score.rubric)">
                    <h3 class="text-md font-bold text-slate-800 dark:text-slate-200">
                      {{ score.factor_name }}
                    </h3>
                    <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5">
                      {{ t('riskFactors.scoring.weightInfo', { weight: (score.weight * 100).toFixed(0), weighted: (score.weight * score.score).toFixed(2) }) }}
                    </p>
                  </div>
                  <div class="flex items-center gap-3">
                    <USelect
                      v-model.number="score.score"
                      :items="[1, 2, 3, 4, 5]"
                      size="sm"
                      color="neutral"
                      class="w-20"
                      @update:model-value="onScoreChange(score)"
                    />
                    <UButton
                      icon="i-lucide-info"
                      color="neutral"
                      variant="ghost"
                      size="md"
                      @click="openGuidelines(score.rubric)"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Empty State when no entity is selected -->
            <div v-else class="text-center py-20 border border-dashed border-slate-200 dark:border-slate-800 rounded-xl bg-slate-50/20 dark:bg-slate-900/10">
              <UIcon name="i-lucide-clipboard" class="w-12 h-12 text-slate-300 mx-auto" />
              <h3 class="mt-4 text-sm font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.scoring.selectEntityTitle') }}</h3>
              <p class="mt-1 text-md text-slate-500 dark:text-slate-400">
                {{ t('riskFactors.scoring.selectEntityDesc') }}
              </p>
            </div>

            <template #footer v-if="activeYearlyEntity">
              <div class="flex justify-end gap-3">
                <UButton
                  icon="i-lucide-save"
                  color="primary"
                  :label="t('riskFactors.scoring.saveBtn')"
                  :loading="auditStore.loading"
                  @click="saveEntityScoring"
                />
              </div>
            </template>
          </UCard>
        </div>
      </template>

      <!-- Tab 3: Rekapitulasi & Priorities -->
      <template #priority>
        <div class="mt-6 space-y-6">
          <UCard class="shadow-sm border border-[var(--border-main)]">
            <template #header>
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 font-space">
                    {{ t('riskFactors.priority.title', { year: selectedYear }) }}
                  </h2>
                  <p class="text-md text-slate-500 mt-0.5">
                    {{ t('riskFactors.priority.subtitle') }}
                  </p>
                </div>
                <div class="flex items-center gap-4">
                  <UBadge color="success" variant="subtle" class="font-bold">
                    {{ t('riskFactors.priority.prioritizedCount', { count: prioritizedCount }) }}
                  </UBadge>
                  <UBadge color="info" variant="solid" class="font-bold">
                    {{ t('riskFactors.priority.priorityRule') }}
                  </UBadge>
                </div>
              </div>
            </template>

            <!-- Recap Table -->
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-sm">
                <thead class="bg-slate-50 dark:bg-slate-850/50">
                  <tr>
                    <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.priority.no') }}</th>
                    <th scope="col" class="px-6 py-3 text-left font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.priority.auditableEntity') }}</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.priority.riskIndex') }}</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.priority.riskLevel') }}</th>
                    <th scope="col" class="px-6 py-3 text-center scope font-semibold text-slate-700 dark:text-slate-300">{{ t('riskFactors.priority.auditPriorityCol') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr 
                    v-for="(ent, idx) in sortedYearlyUniverse" 
                    :key="ent.id"
                    class="hover:bg-slate-50/50 dark:hover:bg-slate-800/10"
                    :class="ent.audit_priority ? 'bg-primary-50/10' : ''"
                  >
                    <td class="px-6 py-4 text-slate-500 font-medium">{{ idx + 1 }}</td>
                    <td class="px-6 py-4 font-bold text-slate-800 dark:text-slate-200">
                      {{ ent.corporate_audit_universe?.name }}
                    </td>
                    <td class="px-6 py-4 text-center font-semibold text-slate-700 dark:text-slate-300">
                      {{ ent.risk_index?.toFixed(1) }}%
                    </td>
                    <td class="px-6 py-4 text-center">
                      <UBadge :color="getRiskLevelBadgeColor(ent.risk_level)" size="md" class="font-bold">
                        {{ formatRiskLevel(ent.risk_level) }}
                      </UBadge>
                    </td>
                    <td class="px-6 py-4 text-center">
                      <div v-if="ent.audit_priority" class="inline-flex items-center gap-1.5 text-success-600 dark:text-success-400 font-bold">
                        <UIcon name="i-lucide-check-circle" class="w-5 h-5 text-success-500" />
                        <span>{{ t('riskFactors.priority.priorityBadge') }}</span>
                      </div>
                      <span v-else class="text-slate-400 text-md">-</span>
                    </td>
                  </tr>
                  <tr v-if="yearlyUniverse.length === 0">
                    <td colspan="5" class="text-center py-10 text-slate-400 text-md">
                      {{ t('riskFactors.priority.noEntities', { year: selectedYear }) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <template #footer>
              <div class="flex items-center justify-between text-[10px] text-slate-400">
                <span>{{ t('riskFactors.priority.sortedNote') }}</span>
              </div>
            </template>
          </UCard>

          <!-- Risk Index Level Info -->
          <UCard class="shadow-sm border border-[var(--border-main)] bg-slate-50/50 dark:bg-slate-900/30">
            <template #header>
              <h3 class="text-sm font-bold text-slate-800 dark:text-slate-100">
                {{ t('riskFactors.priority.levelInfoTitle') }}
              </h3>
            </template>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-800 text-md text-center">
                <thead class="bg-slate-100 dark:bg-slate-800">
                  <tr>
                    <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300 text-center">{{ t('riskFactors.priority.riskIndex') }}</th>
                    <th class="px-4 py-2 font-semibold text-slate-700 dark:text-slate-300 text-center">{{ t('riskFactors.priority.riskLevel') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium text-center">80 - 100%</td>
                    <td class="px-4 py-2 text-center flex justify-center">
                      <span class="inline-flex items-center justify-center font-bold px-3 py-1 rounded text-black text-xs w-36 shadow-sm" style="background-color: #F44336;">
                        {{ t('riskFactors.priority.levels.high') }}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium text-center">60 - 79%</td>
                    <td class="px-4 py-2 text-center flex justify-center">
                      <span class="inline-flex items-center justify-center font-bold px-3 py-1 rounded text-black text-xs w-36 shadow-sm" style="background-color: #FF9800;">
                        {{ t('riskFactors.priority.levels.moderateToHigh') }}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium text-center">40 - 59%</td>
                    <td class="px-4 py-2 text-center flex justify-center">
                      <span class="inline-flex items-center justify-center font-bold px-3 py-1 rounded text-black text-xs w-36 shadow-sm" style="background-color: #FFC107;">
                        {{ t('riskFactors.priority.levels.moderate') }}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium text-center">20 - 39%</td>
                    <td class="px-4 py-2 text-center flex justify-center">
                      <span class="inline-flex items-center justify-center font-bold px-3 py-1 rounded text-black text-xs w-36 shadow-sm" style="background-color: #8BC34A;">
                        {{ t('riskFactors.priority.levels.lowToModerate') }}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 text-slate-600 dark:text-slate-400 font-medium text-center">0 - 19%</td>
                    <td class="px-4 py-2 text-center flex justify-center">
                      <span class="inline-flex items-center justify-center font-bold px-3 py-1 rounded text-black text-xs w-36 shadow-sm" style="background-color: #4CAF50;">
                        {{ t('riskFactors.priority.levels.low') }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </UCard>
        </div>
      </template>
    </UTabs>

    <!-- Scoring Scale Guidelines Modal -->
    <UModal v-model:open="guidelinesModalOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="font-bold text-base text-slate-800 dark:text-slate-100 flex items-center gap-2">
              {{ t('riskFactors.guidelines.modalTitle', { name: detailFactor?.name }) }}
            </h3>
          </template>
          
          <div v-if="detailFactor" class="space-y-4">
            <p class="text-md text-slate-500 dark:text-slate-400">{{ detailFactor.description }}</p>
            <div class="space-y-3 mt-4">
              <div 
                v-for="guide in parsedGuidelines" 
                :key="guide.score" 
                class="p-3 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30 flex items-start gap-3"
              >
                <div 
                  class="w-6 h-6 rounded-full flex items-center justify-center text-md font-bold shrink-0 mt-0.5"
                  :class="getScoreColor(guide.score)"
                >
                  {{ guide.score }}
                </div>
                <div>
                  <p class="text-md font-bold text-slate-700 dark:text-slate-300">{{ getScoreLabel(guide.score) }}</p>
                  <p class="text-md text-slate-500 dark:text-slate-400 mt-0.5 font-sans leading-normal">{{ guide.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRiskFactorsStore } from '~/stores/risk-factors'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useI18n } from '~/composables/useI18n'

const store = useRiskFactorsStore()
const auditStore = useAuditUniverseStore()
const { t } = useI18n()

const tabItems = computed(() => [
  { slot: 'weighting', label: t('riskFactors.tabs.weight'), icon: 'i-lucide-activity' },
  { slot: 'scoring', label: t('riskFactors.tabs.score'), icon: 'i-lucide-award' },
  { slot: 'priority', label: t('riskFactors.tabs.priority'), icon: 'i-lucide-list-checks' }
])

// State
const searchQuery = ref('')
const selectedCorporateList = ref<any[]>([])
const detailFactor = ref<any>(null)
const guidelinesModalOpen = ref(false)
const alertMessage = ref('')
const alertType = ref('success')

// For scoring workspace:
const selectedYear = ref(2026)
const selectedEntityId = ref<string | undefined>(undefined)
const activeYearlyEntity = ref<any>(null)
const scoringRows = ref<any[]>([])
const rubricModalOpen = ref(false)
const rubricFactor = ref<any>(null)

// Lifecycle
onMounted(async () => {
  await store.fetchStandardFactors()
  await store.fetchCorporateFactors()
  await auditStore.fetchStandardUniverse()
  await auditStore.fetchCorporateUniverse()
  await fetchYearlyUniverse()
  
  // Set default details to Financial Materiality
  if (store.standardFactors.length > 0) {
    detailFactor.value = store.standardFactors[0]
  }

  // Populate local corporate selection from store
  selectedCorporateList.value = store.corporateFactors.map(cf => ({
    standard_risk_factor_id: cf.standard_risk_factor_id,
    weight: Math.round(cf.weight * 100), // convert 0.15 to 15
    standard_risk_factor: cf.standard_risk_factor
  }))
})

const filteredStandardFactors = computed(() => {
  if (!searchQuery.value) return store.standardFactors
  const query = searchQuery.value.toLowerCase().trim()
  if (!query) return store.standardFactors

  const matched = store.standardFactors.filter(f => {
    const nameMatch = f.name?.toLowerCase().includes(query)
    const descMatch = f.description?.toLowerCase().includes(query)
    return nameMatch || descMatch
  })

  return matched.slice().sort((a, b) => {
    const nameA = (a.name || '').toLowerCase()
    const nameB = (b.name || '').toLowerCase()

    const getScore = (name: string, desc: string) => {
      if (name.startsWith(query)) return 1
      const words = name.split(/\s+/)
      if (words.some(w => w.startsWith(query))) return 2
      if (name.includes(query)) return 3
      if ((desc || '').toLowerCase().startsWith(query)) return 4
      return 5
    }

    const scoreA = getScore(nameA, a.description)
    const scoreB = getScore(nameB, b.description)

    if (scoreA !== scoreB) {
      return scoreA - scoreB
    }

    return nameA.localeCompare(nameB)
  })
})

const parsedGuidelines = computed(() => {
  if (!detailFactor.value?.score_guidelines) return []
  try {
    return JSON.parse(detailFactor.value.score_guidelines)
  } catch (e) {
    console.error('Failed to parse guidelines:', e)
    return []
  }
})

// Correct reactive calculation of total weight sum
const totalWeight = computed(() => {
  return selectedCorporateList.value.reduce((sum, item) => sum + (Number(item.weight) || 0), 0)
})

const isValidWeightSum = computed(() => {
  return totalWeight.value === 100
})

const yearlyUniverse = computed(() => auditStore.yearlyUniverse)

const dropdownEntities = computed(() => {
  return auditStore.yearlyUniverse.map(item => ({
    label: item.corporate_audit_universe?.name || 'Unknown',
    value: item.id,
    original: item
  }))
})

const sortedYearlyUniverse = computed(() => {
  return [...auditStore.yearlyUniverse]
    .sort((a, b) => (b.risk_index || 0) - (a.risk_index || 0))
})

const prioritizedCount = computed(() => {
  return auditStore.yearlyUniverse.filter(ent => ent.audit_priority).length
})

const totalWeightedScore = computed(() => {
  return scoringRows.value.reduce((sum, item) => sum + (item.score * item.weight || 0), 0)
})

// Methods
const fetchYearlyUniverse = async () => {
  selectedEntityId.value = undefined
  activeYearlyEntity.value = null
  scoringRows.value = []
  await auditStore.fetchYearlyUniverse(selectedYear.value)
}

const isFactorSelected = (id: string): boolean => {
  return selectedCorporateList.value.some(item => item.standard_risk_factor_id === id)
}

const toggleFactorSelection = (factor: any) => {
  const index = selectedCorporateList.value.findIndex(item => item.standard_risk_factor_id === factor.id)
  if (index >= 0) {
    selectedCorporateList.value.splice(index, 1)
  } else {
    const currentSum = selectedCorporateList.value.reduce((s, i) => s + (i.weight || 0), 0)
    const remaining = Math.max(0, 100 - currentSum)
    
    selectedCorporateList.value.push({
      standard_risk_factor_id: factor.id,
      weight: remaining,
      standard_risk_factor: factor
    })
  }
}

const removeFactorFromCorporate = (item: any) => {
  selectedCorporateList.value = selectedCorporateList.value.filter(
    i => i.standard_risk_factor_id !== item.standard_risk_factor_id
  )
}

const openGuidelines = (factor: any) => {
  if (factor) {
    detailFactor.value = factor
    guidelinesModalOpen.value = true
  }
}

const selectEntityForScoringById = (entityId?: string) => {
  if (!entityId) {
    selectEntityForScoring(null)
    return
  }

  const selected = dropdownEntities.value.find(
    item => item.value === entityId
  )

  selectEntityForScoring(selected?.original ?? null)
}

const selectEntityForScoring = (ent: any) => {
  if (!ent) {
    activeYearlyEntity.value = null
    scoringRows.value = []
    return
  }
  activeYearlyEntity.value = ent
  scoringRows.value = store.corporateFactors.map(cf => {
    const matchScore = ent.risk_scores?.find((s: any) => s.corporate_risk_factor_id === cf.id)
    return {
      corporate_risk_factor_id: cf.id,
      factor_name: cf.standard_risk_factor?.name,
      weight: cf.weight,
      score: matchScore ? matchScore.score : 3,
      rubric: cf.standard_risk_factor
    }
  })
}

const onScoreChange = (scoreRow: any) => {
  const sum = scoringRows.value.reduce((s, row) => s + (row.score * row.weight), 0)
  activeYearlyEntity.value.risk_index = (sum / 5.0) * 100.0
  
  if (activeYearlyEntity.value.risk_index >= 80.0) {
    activeYearlyEntity.value.risk_level = 'High'
    activeYearlyEntity.value.audit_priority = true
  } else if (activeYearlyEntity.value.risk_index >= 60.0) {
    activeYearlyEntity.value.risk_level = 'Medium to High'
    activeYearlyEntity.value.audit_priority = true
  } else if (activeYearlyEntity.value.risk_index >= 40.0) {
    activeYearlyEntity.value.risk_level = 'Medium'
    activeYearlyEntity.value.audit_priority = false
  } else if (activeYearlyEntity.value.risk_index >= 20.0) {
    activeYearlyEntity.value.risk_level = 'Low to Medium'
    activeYearlyEntity.value.audit_priority = false
  } else {
    activeYearlyEntity.value.risk_level = 'Low'
    activeYearlyEntity.value.audit_priority = false
  }
}

const saveEntityScoring = async () => {
  if (!activeYearlyEntity.value) return

  const payload = {
    audit_universe_year_id: activeYearlyEntity.value.id,
    scores: scoringRows.value.map(row => ({
      corporate_risk_factor_id: row.corporate_risk_factor_id,
      score: row.score
    }))
  }

  const res = await auditStore.scoreYearlyEntity(selectedYear.value, payload)
  if (res) {
    showAlert(t('riskFactors.messages.scoresSaved'), 'success')
    await fetchYearlyUniverse()
  } else {
    showAlert(auditStore.errorMsg || t('riskFactors.messages.scoresSaveFailed'), 'error')
  }
}

const saveChanges = async () => {
  if (!isValidWeightSum.value) {
    showAlert(t('riskFactors.messages.weightSumMustBe100', { total: totalWeight.value }), 'error')
    return
  }

  const payload = selectedCorporateList.value.map(item => ({
    standard_risk_factor_id: item.standard_risk_factor_id,
    weight: item.weight
  }))

  const success = await store.saveCorporateFactors(payload)
  if (success) {
    showAlert(t('riskFactors.messages.weightsUpdated'), 'success')
  } else {
    showAlert(store.errorMsg || t('riskFactors.messages.weightsUpdateFailed'), 'error')
  }
}

const getScoreColor = (score: number) => {
  switch (score) {
    case 5: return 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300'
    case 4: return 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300'
    case 3: return 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
    case 2: return 'bg-lime-100 text-lime-700 dark:bg-lime-900/40 dark:text-lime-300'
    case 1: return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
    default: return 'bg-slate-100 text-slate-700'
  }
}

const getScoreLabel = (score: number) => {
  const labelKey = `riskFactors.scoreLabels.${score}`
  const label = t(labelKey)
  return label !== labelKey ? label : ''
}

const formatRiskLevel = (level?: string) => {
  if (!level) return 'N/A'
  switch (level) {
    case 'High': return t('riskFactors.priority.levels.high')
    case 'Medium to High': return t('riskFactors.priority.levels.moderateToHigh')
    case 'Medium': return t('riskFactors.priority.levels.moderate')
    case 'Low to Medium': return t('riskFactors.priority.levels.lowToModerate')
    case 'Low': return t('riskFactors.priority.levels.low')
    default: return level
  }
}

const getRiskLevelBadgeColor = (level?: string) => {
  if (!level) return 'neutral'
  switch (level) {
    case 'High': return 'error'
    case 'Medium to High': return 'warning'
    case 'Medium': return 'primary'
    case 'Low to Medium': return 'info'
    case 'Low': return 'success'
    default: return 'neutral'
  }
}

const showAlert = (msg: string, type: string) => {
  alertMessage.value = msg
  alertType.value = type
  setTimeout(() => {
    alertMessage.value = ''
  }, 4000)
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
