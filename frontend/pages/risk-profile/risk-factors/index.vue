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
        variant="outline"
        :title="alertType === 'success' ? t('common.success') : t('common.error')"
        :description="alertMessage"
        icon="i-lucide-info"
        :ui="{ description: 'text-white dark:text-white' }"
        class="shadow-md"
        closable
        @close="alertMessage = ''"
      >
        <template #description>
          <span class="text-white">{{ alertMessage }}</span>
        </template>
      </UAlert>
    </Transition>

    <!-- Tabs Navigation -->
    <UTabs :items="tabItems" class="w-full">
      <!-- Tab 1: Corporate Weighting -->
      <template #weighting>
        <div class="space-y-6 mt-6">
          <!-- Validation Status Banner -->
          <UCard 
            class="shadow-sm transition-all duration-300 rounded-xl"
            :class="isValidWeightSum 
              ? 'bg-emerald-600 dark:bg-emerald-600 text-white border-0' 
              : 'bg-rose-600 dark:bg-rose-600 text-white border-0'"
            :ui="{ body: 'p-4 sm:p-4' }"
          >
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div class="flex items-center gap-3">
                <UIcon 
                  :name="isValidWeightSum ? 'i-lucide-check-circle-2' : 'i-lucide-alert-triangle'" 
                  class="w-6 h-6 text-white shrink-0"
                />
                <div>
                  <p class="text-sm font-semibold text-white">
                    <span class="font-bold text-base text-white">{{ t('riskFactors.weighting.totalWeight', { total: totalWeight }) }}</span>
                  </p>
                  <p class="text-sm text-white!">
                    {{ isValidWeightSum ? t('riskFactors.weighting.validSum') : t('riskFactors.weighting.invalidSum') }}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-3 shrink-0">
                <UBadge 
                  :color="isValidWeightSum ? 'success' : 'error'" 
                  size="md" 
                  variant="solid"
                  class="bg-white/20 text-white border border-white/30 backdrop-blur-sm"
                >
                  {{ isValidWeightSum ? t('riskFactors.weighting.badgeValid') : t('riskFactors.weighting.badgeInvalid') }}
                </UBadge>
                <UButton
                  v-if="canEditRiskFactors"
                  icon="i-lucide-save"
                  color="neutral"
                  variant="solid"
                  class="bg-white text-slate-900 hover:bg-slate-100 dark:bg-primary-600/90 dark:text-slate-100 dark:hover:bg-slate-100 font-semibold shadow-sm"
                  :loading="store.loading"
                  :disabled="!isValidWeightSum"
                  @click="saveChanges"
                >
                  {{ t('riskFactors.weighting.saveWeights') }}
                </UButton>
              </div>
            </div>
          </UCard>

          <div class="flex flex-col gap-8">
            <!-- Catalog Explorer -->
            <div class="space-y-6">
              <UCard class="shadow-sm border border-[var(--border-main)]">
                <template #header>
                  <div class="space-y-3">
                    <div class="flex items-center justify-between gap-4">
                      <h2 class="text-base font-bold text-slate-800 dark:text-slate-100 font-space flex items-center gap-2">
                        {{ t('riskFactors.weighting.standardFactorsTitle') }}
                      </h2>
                      <UButton
                        v-if="canEditRiskFactors"
                        icon="i-lucide-plus"
                        color="primary"
                        size="sm"
                        variant="solid"
                        @click="openAddFactorModal"
                      >
                        {{ t('riskFactors.weighting.addFactor') }}
                      </UButton>
                    </div>
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
                    class="py-3 flex items-start justify-between gap-4 hover:bg-slate-50 dark:hover:bg-slate-800/40 p-2 rounded-lg transition-colors duration-200 group"
                  >
                    <div class="cursor-pointer flex-1" @click="openGuidelines(factor)">
                      <h3 class="text-md font-semibold text-slate-800 dark:text-slate-200 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                        {{ factor.name }}
                      </h3>
                      <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5 line-clamp-2">
                        {{ factor.description }}
                      </p>
                    </div>
                    <div class="flex items-center gap-1.5 shrink-0">
                      <UTooltip :text="t('riskFactors.weighting.viewGuidelines')">
                        <UButton
                          icon="i-lucide-book-open"
                          color="neutral"
                          variant="ghost"
                          size="md"
                          @click.stop="openGuidelines(factor)"
                        />
                      </UTooltip>
                      <template v-if="canEditRiskFactors">
                        <UTooltip :text="t('riskFactors.weighting.editFactor')">
                          <UButton
                            icon="i-lucide-edit"
                            color="warning"
                            variant="ghost"
                            size="md"
                            @click.stop="openEditFactorModal(factor)"
                          />
                        </UTooltip>
                        <UTooltip :text="t('riskFactors.weighting.deleteFactor')">
                          <UButton
                            icon="i-lucide-trash-2"
                            color="error"
                            variant="ghost"
                            size="md"
                            @click.stop="openDeleteFactorModal(factor)"
                          />
                        </UTooltip>
                      </template>
                      <div class="ml-1 pl-2 border-l border-slate-200 dark:border-slate-700">
                        <UCheckbox
                          :model-value="isFactorSelected(factor.id)"
                          @update:model-value="toggleFactorSelection(factor)"
                        />
                      </div>
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
                    <UTooltip :text="t('riskFactors.scoring.guidelinesTooltip')">
                      <UButton
                        icon="i-lucide-info"
                        color="neutral"
                        variant="ghost"
                        size="md"
                        @click="openGuidelines(score.rubric)"
                      />
                    </UTooltip>
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
              <div v-if="canEditRiskFactors" class="flex justify-end gap-3">
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
    </UTabs>

    <!-- Scoring Scale Guidelines Modal -->
    <!-- Scoring Scale Guidelines Modal -->
    <UModal 
      v-model:open="guidelinesModalOpen"
      :ui="{ content: 'sm:max-w-xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
    >
      <template #content>
        <div class="relative flex flex-col max-h-[85vh] transition-colors duration-300">
          <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <div class="flex items-center gap-3">
              <div class="p-2.5 rounded-xl bg-primary-500/10 text-primary-500 dark:bg-primary-500/20">
                <UIcon name="i-lucide-book-open" class="w-5 h-5" />
              </div>
              <div>
                <h3 class="font-bold text-base text-[var(--text-main)]">
                  {{ t('riskFactors.guidelines.modalTitle', { name: detailFactor?.name }) }}
                </h3>
                <p class="text-md text-[var(--text-muted)] mt-0.5">{{ t('riskFactors.weighting.guidelinesSection') }}</p>
              </div>
            </div>
            <UButton
              icon="i-lucide-x"
              color="neutral"
              variant="ghost"
              size="sm"
              class="rounded-lg text-[var(--text-muted)] hover:text-[var(--text-main)]"
              @click="() => { guidelinesModalOpen = false; }"
            />
          </div>
          
          <div v-if="detailFactor" class="p-6 overflow-y-auto space-y-4 flex-1">
            <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-900/50 border border-slate-200/60 dark:border-slate-800 text-md text-slate-600 dark:text-slate-300 leading-relaxed">
              <span class="font-semibold text-slate-800 dark:text-slate-200 block mb-0.5">Deskripsi Faktor:</span>
              {{ detailFactor.description || '-' }}
            </div>
            <div class="space-y-2.5">
              <div 
                v-for="guide in parsedGuidelines" 
                :key="guide.score" 
                class="p-3 rounded-xl border border-slate-200/70 dark:border-slate-800 bg-white dark:bg-slate-900/40 flex items-start gap-3 shadow-md hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
              >
                <div 
                  class="w-7 h-7 rounded-lg flex items-center justify-center text-md font-bold shrink-0 mt-0.5 shadow-md"
                  :class="getScoreColor(guide.score)"
                >
                  {{ guide.score }}
                </div>
                <div class="flex-1">
                  <p class="text-md font-bold text-slate-800 dark:text-slate-200">{{ getScoreLabel(guide.score) }}</p>
                  <p class="text-md text-slate-600 dark:text-slate-400 mt-0.5 leading-relaxed">{{ guide.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Add / Edit Standard Risk Factor Modal -->
    <UModal 
      v-model:open="factorModalOpen"
      :ui="{ content: 'sm:max-w-2xl md:max-w-3xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
    >
      <template #content>
        <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
          <div class="flex items-center justify-between px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <div class="flex items-center gap-3">
              <div 
                class="p-2.5 rounded-xl"
                :class="isEditingFactor ? 'bg-amber-500/10 text-amber-500 dark:bg-amber-500/20' : 'bg-primary-500/10 text-primary-500 dark:bg-primary-500/20'"
              >
                <UIcon :name="isEditingFactor ? 'i-lucide-edit-3' : 'i-lucide-plus-circle'" class="w-5 h-5" />
              </div>
              <div>
                <h3 class="font-bold text-base text-[var(--text-main)]">
                  {{ isEditingFactor ? t('riskFactors.weighting.editModalTitle') : t('riskFactors.weighting.addModalTitle') }}
                </h3>
                <p class="text-md text-[var(--text-muted)] mt-0.5">
                  {{ isEditingFactor ? t('riskFactors.weighting.editModalSubtitle') : t('riskFactors.weighting.addModalSubtitle') }}
                </p>
              </div>
            </div>
            <UButton
              icon="i-lucide-x"
              color="neutral"
              variant="ghost"
              size="sm"
              class="rounded-lg text-[var(--text-muted)] hover:text-[var(--text-main)]"
              @click="() => { factorModalOpen = false; }"
            />
          </div>

          <form @submit.prevent="saveFactorForm" class="px-6 py-4 overflow-y-auto space-y-3.5 flex-1 max-h-[calc(90vh-130px)]">
            <!-- Factor Name -->
            <UFormField :label="t('riskFactors.weighting.factorName')" required>
              <UInput
                v-model="factorForm.name"
                :placeholder="t('riskFactors.weighting.factorNamePlaceholder')"
                class="w-full"
                maxlength="200"
                required
              />
            </UFormField>

            <!-- Factor Description -->
            <UFormField :label="t('riskFactors.weighting.factorDesc')">
              <UTextarea
                v-model="factorForm.description"
                :placeholder="t('riskFactors.weighting.factorDescPlaceholder')"
                class="w-full"
                :rows="2"
              />
            </UFormField>

            <!-- Score Guidelines Section -->
            <div class="space-y-2.5 pt-2.5 border-t border-[var(--border-main)]">
              <div class="flex items-center justify-between">
                <div>
                  <label class="text-md font-bold text-[var(--text-main)] uppercase tracking-wider block">
                    {{ t('riskFactors.weighting.guidelinesSection') }}
                  </label>
                  <p class="text-[11px] text-[var(--text-muted)] mt-0.5">
                    {{ t('riskFactors.weighting.guidelinesSubtitle') }}
                  </p>
                </div>
              </div>

              <div class="space-y-2">
                <div
                  v-for="scoreItem in factorForm.scores"
                  :key="scoreItem.score"
                  class="flex flex-col sm:flex-row sm:items-center gap-2.5 p-2 rounded-xl border border-slate-200/70 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/30 hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
                >
                  <div class="flex items-center gap-2 sm:w-44 shrink-0">
                    <div
                      class="w-6 h-6 rounded-md flex items-center justify-center text-md font-bold shrink-0 shadow-md"
                      :class="getScoreColor(scoreItem.score)"
                    >
                      {{ scoreItem.score }}
                    </div>
                    <span class="text-md font-semibold text-slate-700 dark:text-slate-200 truncate">
                      {{ getScoreLabel(scoreItem.score) }}
                    </span>
                  </div>
                  <div class="flex-1">
                    <UInput
                      v-model="scoreItem.desc"
                      size="sm"
                      :placeholder="t('riskFactors.weighting.scoreDescPlaceholder', { score: scoreItem.score })"
                      class="w-full"
                    />
                  </div>
                </div>
              </div>
            </div>
          </form>

          <div class="flex items-center justify-end gap-2.5 px-6 py-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)]">
            <UButton
              color="neutral"
              variant="outline"
              size="md"
              class="rounded-xl font-medium"
              :label="t('riskFactors.weighting.cancel')"
              @click="() => { factorModalOpen = false; }"
            />
            <UButton
              color="primary"
              variant="solid"
              size="md"
              class="rounded-xl font-semibold shadow-sm"
              :loading="factorSubmitting"
              icon="i-lucide-save"
              :label="t('riskFactors.weighting.save')"
              @click="saveFactorForm"
            />
          </div>
        </div>
      </template>
    </UModal>

    <!-- Delete Confirmation Modal -->
    <UModal 
      v-model:open="deleteModalOpen"
      :ui="{ 
        content: 'sm:max-w-md w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' 
      }"
    >
      <template #content>
        <div class="p-6 space-y-4">
          <!-- Header with warning icon -->
          <div class="flex items-start gap-3.5">
            <div class="w-10 h-10 rounded-xl bg-error-50 dark:bg-error-950/50 text-error-600 dark:text-error-400 flex items-center justify-center shrink-0 ring-4 ring-error-500/10">
              <UIcon name="i-lucide-alert-triangle" class="w-5 h-5" />
            </div>
            <div class="flex-1 min-w-0 pt-0.5">
              <h3 class="text-base font-bold text-[var(--text-main)] leading-snug">
                {{ t('riskFactors.weighting.deleteModalTitle') }}
              </h3>
              <p class="text-md text-[var(--text-muted)] mt-1">
                {{ t('riskFactors.weighting.deleteConfirmDesc') }}
              </p>
            </div>
          </div>

          <!-- Highlight Factor Name pill -->
          <div v-if="factorToDelete" class="p-3 rounded-xl bg-error-50/50 dark:bg-error-950/30 border border-error-200 dark:border-error-800/60 flex items-center gap-2.5">
            <UIcon name="i-lucide-activity" class="w-4 h-4 text-error-600 dark:text-error-400 shrink-0" />
            <div class="min-w-0 flex-1">
              <p class="text-md font-semibold text-error-900 dark:text-error-200 truncate">
                {{ factorToDelete.name }}
              </p>
              <p v-if="factorToDelete.description" class="text-[11px] text-error-700/80 dark:text-error-400/80 truncate mt-0.5">
                {{ factorToDelete.description }}
              </p>
            </div>
          </div>

          <!-- Actions Footer -->
          <div class="flex items-center justify-end gap-2.5 pt-2">
            <UButton
              color="neutral"
              variant="outline"
              size="md"
              class="rounded-xl font-medium"
              :label="t('riskFactors.weighting.cancel')"
              :disabled="deleteSubmitting"
              @click="() => { deleteModalOpen = false; }"
            />
            <UButton
              color="error"
              variant="solid"
              size="md"
              class="rounded-xl font-semibold shadow-sm"
              :loading="deleteSubmitting"
              icon="i-lucide-trash-2"
              :label="t('riskFactors.weighting.delete')"
              @click="confirmDeleteFactor"
            />
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRiskFactorsStore } from '~/stores/risk-factors'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

const store = useRiskFactorsStore()
const auditStore = useAuditUniverseStore()
const { t } = useI18n()
const { canEditRiskFactors } = useRbac()
const toast = useToastNotification()

const tabItems = computed(() => [
  { slot: 'weighting', label: t('riskFactors.tabs.weight'), icon: 'i-lucide-activity' },
  { slot: 'scoring', label: t('riskFactors.tabs.score'), icon: 'i-lucide-award' }
])

// State
const searchQuery = ref('')
const selectedCorporateList = ref<any[]>([])
const detailFactor = ref<any>(null)
const guidelinesModalOpen = ref(false)
const alertMessage = ref('')
const alertType = ref('success')

// Modals for Standard Risk Factors CRUD
const factorModalOpen = ref(false)
const isEditingFactor = ref(false)
const editingFactorId = ref<string | null>(null)
const factorSubmitting = ref(false)
const factorForm = ref({
  name: '',
  description: '',
  scores: [
    { score: 5, desc: '' },
    { score: 4, desc: '' },
    { score: 3, desc: '' },
    { score: 2, desc: '' },
    { score: 1, desc: '' }
  ]
})

const deleteModalOpen = ref(false)
const factorToDelete = ref<any>(null)
const deleteSubmitting = ref(false)

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
  toast.showSuccess('Factor berhasil dihapus')
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
    toast.showSuccess('Skor berhasil disimpan')
    await fetchYearlyUniverse()
  } else {
    toast.showError('Skor gagal disimpan')
  }
}

const saveChanges = async () => {
  if (!isValidWeightSum.value) {
    toast.showError(`Jumlah bobot harus 100%, total saat ini: ${totalWeight.value}`)
    return
  }

  const payload = selectedCorporateList.value.map(item => ({
    standard_risk_factor_id: item.standard_risk_factor_id,
    weight: item.weight
  }))

  const success = await store.saveCorporateFactors(payload)
  if (success) {
    toast.showSuccess('Bobot berhasil disimpan')
  } else {
    toast.showError('Bobot gagal disimpan')
  }
}

const getDefaultScores = () => [
  { score: 5, desc: 'High – Major contributor to enterprise risk' },
  { score: 4, desc: 'Medium to High – Significant risk affecting key operations' },
  { score: 3, desc: 'Medium – Moderate risk exposure with limited enterprise impact' },
  { score: 2, desc: 'Low to Medium – Low risk operations with minimal impact' },
  { score: 1, desc: 'Low – Administrative or routine activity with minimal risk' }
]

const openAddFactorModal = () => {
  isEditingFactor.value = false
  editingFactorId.value = null
  factorForm.value = {
    name: '',
    description: '',
    scores: getDefaultScores()
  }
  factorModalOpen.value = true
}

const openEditFactorModal = (factor: any) => {
  isEditingFactor.value = true
  editingFactorId.value = factor.id
  let parsedScores = getDefaultScores()
  if (factor.score_guidelines) {
    try {
      const parsed = JSON.parse(factor.score_guidelines)
      if (Array.isArray(parsed)) {
        parsedScores = [5, 4, 3, 2, 1].map(scoreNum => {
          const found = parsed.find((item: any) => item.score === scoreNum)
          return {
            score: scoreNum,
            desc: found?.desc || ''
          }
        })
      }
    } catch (e) {
      console.error('Failed to parse guidelines for editing:', e)
    }
  }

  factorForm.value = {
    name: factor.name || '',
    description: factor.description || '',
    scores: parsedScores
  }
  factorModalOpen.value = true
}

const saveFactorForm = async () => {
  if (!factorForm.value.name.trim()) {
    toast.showError('Name is required')
    return
  }

  factorSubmitting.value = true
  try {
    const guidelinesJson = JSON.stringify(factorForm.value.scores.map(s => ({
      score: s.score,
      desc: s.desc.trim()
    })))

    const payload = {
      name: factorForm.value.name.trim(),
      description: factorForm.value.description.trim(),
      score_guidelines: guidelinesJson
    }

    if (isEditingFactor.value && editingFactorId.value) {
      const result = await store.updateStandardFactor(editingFactorId.value, payload)
      if (result) {
        toast.showSuccess(t('riskFactors.messages.factorUpdated'))
        factorModalOpen.value = false
        if (detailFactor.value?.id === editingFactorId.value) {
          detailFactor.value = result
        }
      } else {
        toast.showError(store.errorMsg || t('riskFactors.messages.factorUpdateFailed'))
      }
    } else {
      const result = await store.createStandardFactor(payload)
      if (result) {
        toast.showSuccess(t('riskFactors.messages.factorCreated'))
        factorModalOpen.value = false
      } else {
        toast.showError(store.errorMsg || t('riskFactors.messages.factorCreateFailed'))
      }
    }
  } catch (err: any) {
    toast.showError(err.message || 'Operation failed')
  } finally {
    factorSubmitting.value = false
  }
}

const openDeleteFactorModal = (factor: any) => {
  factorToDelete.value = factor
  deleteModalOpen.value = true
}

const confirmDeleteFactor = async () => {
  if (!factorToDelete.value) return
  deleteSubmitting.value = true
  try {
    const id = factorToDelete.value.id
    const success = await store.deleteStandardFactor(id)
    if (success) {
      toast.showSuccess(t('riskFactors.messages.factorDeleted'))
      const idx = selectedCorporateList.value.findIndex(item => item.standard_risk_factor_id === id)
      if (idx >= 0) {
        selectedCorporateList.value.splice(idx, 1)
      }
      if (detailFactor.value?.id === id) {
        detailFactor.value = store.standardFactors.length > 0 ? store.standardFactors[0] : null
      }
      deleteModalOpen.value = false
      factorToDelete.value = null
    } else {
      toast.showError(store.errorMsg || t('riskFactors.messages.factorDeleteFailed'))
    }
  } catch (err: any) {
    toast.showError(err.message || 'Failed to delete')
  } finally {
    deleteSubmitting.value = false
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
