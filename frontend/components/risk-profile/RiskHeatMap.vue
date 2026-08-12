<template>
  <div class="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <!-- Header Card -->
    <UCard :ui="{ body: { padding: 'p-0' } }" class="mb-8 overflow-visible border-gray-200 dark:border-gray-800">
      <div class="p-6 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-primary to-blue-400 flex items-center justify-center shadow-lg shadow-primary/20">
            <UIcon name="i-heroicons-chart-bar-square" class="w-8 h-8" />
          </div>
          <div>
            <h1 class="text-2xl font-extrabold tracking-tight">{{ t('riskProfile.title') }}</h1>
            <p class="text-sm">{{ t('riskProfile.subtitle', { year: store.selectedYear, period: store.selectedPeriod }) }}</p>
          </div>
        </div>
        
        <div class="flex items-center gap-4">
          <div class="flex gap-4">
            <div class="px-4 py-2 rounded-lg border border-gray-100 dark:border-gray-700 text-center min-w-[80px]">
              <span class="block text-xl font-black ">{{ totalRisks }}</span>
              <span class="block text-[10px] font-bold uppercase tracking-widest">{{ t('riskProfile.total') }}</span>
            </div>
            <div class="px-4 py-2 bg-orange-100 dark:bg-orange-900/50 rounded-lg border border-orange-100 dark:border-orange-900/50 text-center min-w-[80px]">
              <span class="block text-xl text-warning-500 font-black ">{{ priorityCount }}</span>
              <span class="block text-[10px] font-bold text-warning-500 uppercase tracking-widest">{{ t('riskProfile.priority') }}</span>
            </div>
          </div>
          
          <UModal 
            v-model:open="isAddModalOpen"
            :title="t('riskProfile.addModal.title')" 
            :description="t('riskProfile.addModal.desc')"
            :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }"
          >
            <UButton 
              icon="i-heroicons-plus"
              :label="t('riskProfile.addRisk')"
              color="primary"
              size="lg"
              class="font-bold shadow-md shadow-primary/20"
            />

            <template #body>
              <form @submit.prevent="submitNewRisk" class="space-y-4">
                <UFormField :label="t('riskProfile.addModal.name')" required>
                  <UInput v-model="newRisk.name" :placeholder="t('riskProfile.addModal.namePlaceholder')" class="w-full" />
                </UFormField>

                <div class="grid grid-cols-2 gap-4">
                  <UFormField :label="t('riskProfile.addModal.category')">
                    <USelect v-model="newRisk.category" :items="categoryOptions" class="w-full" />
                  </UFormField>
                  <UFormField :label="t('riskProfile.addModal.branch')">
                    <USelect v-model="newRisk.branch" :items="store.branches" class="w-full" />
                  </UFormField>
                </div>

                <!-- Quarterly Residual Score Inputs -->
                <div class="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-gray-50/50 dark:bg-gray-900/50 space-y-4">
                  <span class="block text-md font-black uppercase tracking-wider text-primary-500">{{ t('riskProfile.addModal.quarterlyTitle', { year: store.selectedYear }) }}</span>
                  <div class="grid grid-cols-4 gap-4">
                    <div v-for="q in ['q1', 'q2', 'q3', 'q4']" :key="q" class="space-y-2 border-r last:border-0 border-gray-200 dark:border-gray-700 pr-2">
                      <span class="text-md font-black uppercase text-gray-500">{{ q }}</span>
                      <UFormField :label="t('riskProfile.addModal.impact')">
                        <USelect v-model.number="newRisk[`impact_${q}`]" :items="[1,2,3,4,5]" class="w-full" />
                      </UFormField>
                      <UFormField :label="t('riskProfile.addModal.likelihood')">
                        <USelect v-model.number="newRisk[`likelihood_${q}`]" :items="[1,2,3,4,5]" class="w-full" />
                      </UFormField>
                    </div>
                  </div>
                </div>

                <UFormField :label="t('riskProfile.addModal.description')">
                  <UTextarea v-model="newRisk.description" :placeholder="t('riskProfile.addModal.descriptionPlaceholder')" class="w-full" />
                </UFormField>
              </form>
            </template>

            <template #footer>
              <div class="flex justify-end gap-3">
                <UButton :label="t('common.cancel')" color="neutral" variant="ghost" @click="isAddModalOpen = false" />
                <UButton :label="t('riskProfile.addModal.saveBtn')" color="primary" @click="submitNewRisk" />
              </div>
            </template>
          </UModal>
        </div>
      </div>
    </UCard>

    <!-- Controls & Hint -->
    <div class="grid md:grid-cols-4 gap-6 mb-8 items-start">
      <div class="md:col-span-2">
        <QuickTip
          :title="t('riskProfile.quickTip.title')"
          :description="t('riskProfile.quickTip.desc')"
        />
      </div>
      <div>
        <UFormField :label="t('riskProfile.filterBranch')" size="sm" class="font-bold">
          <USelect
            v-model="selectedBranch"
            :items="branchOptions"
            icon="i-heroicons-building-office"
            class="w-full"
          />
        </UFormField>
      </div>
      <div class="flex gap-4">
        <UFormField :label="t('riskProfile.fiscalYear')" size="sm" class="font-bold w-1/2">
          <USelect
            v-model.number="store.selectedYear"
            :items="[2025, 2026, 2027]"
            icon="i-heroicons-calendar"
            class="w-full"
          />
        </UFormField>
        <UFormField :label="t('riskProfile.period')" size="sm" class="font-bold w-1/2">
          <USelect
            v-model="store.selectedPeriod"
            :items="['Q1', 'Q2', 'Q3', 'Q4']"
            icon="i-heroicons-clock"
            class="w-full"
          />
        </UFormField>
      </div>
    </div>

    <!-- Legend -->
    <div class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 mb-10 flex flex-wrap items-center gap-x-8 gap-y-3">
      <span class="text-[10px] font-black uppercase tracking-[0.2em]">{{ t('riskProfile.riskLevels') }}</span>
      <div class="flex flex-wrap gap-6">
        <div v-for="(config, key) in riskLevelConfig" :key="key" class="flex items-center gap-2">
          <div class="w-3.5 h-3.5 rounded-sm shadow-sm" :style="{ background: config.color }"></div>
          <span class="text-md font-bold">{{ getRiskLevelLabel(key) }}</span>
          <UIcon v-if="config.priority" name="i-heroicons-fire" class="w-3.5 h-3.5 text-warning-500" />
        </div>
      </div>
    </div>

    <!-- Heat Map Grid -->
    <div class="relative pl-12 mb-16 select-none">
      <!-- Y-axis Label -->
      <div class="absolute -left-20 top-1/2 -translate-y-1/2 -rotate-90 origin-center whitespace-nowrap text-[10px] font-black uppercase tracking-[0.3em]">
        {{ t('riskProfile.likelihoodLevel') }}
      </div>

      <div class="flex gap-4">
        <!-- Y-axis Ticks -->
        <div class="flex flex-col w-28 shrink-0">
          <div v-for="l in likelihoodLevels" :key="`y-${l}`" class="flex-1 flex items-center justify-end gap-3 pr-2 min-h-[110px]">
            <span class="text-md font-bold text-right leading-tight max-w-[70px] uppercase">{{ getLikelihoodLabel(l) }}</span>
            <span class="text-xl font-black">{{ l }}</span>
          </div>
        </div>

        <!-- Main Grid Area -->
        <div class="flex-1">
          <div class=" p-1 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700">
            <div class="grid grid-rows-5 gap-1 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
              <div v-for="l in likelihoodLevels" :key="`row-${l}`" class="grid grid-cols-5 gap-1">
                <div
                  v-for="i in impactLevels"
                  :key="`cell-${l}-${i}`"
                  :id="`cell-${l}-${i}`"
                  class="relative min-h-[110px] p-2 flex flex-col items-center justify-center transition-all duration-300 group"
                  :class="[
                    getCellBgClass(l, i),
                    dragOverCell === `${l}-${i}` ? 'ring-4 ring-primary ring-inset z-10 scale-[1.02] shadow-2xl' : ''
                  ]"
                  @dragover.prevent="onDragOver($event, l, i)"
                  @dragleave="onDragLeave"
                  @drop="onDrop($event, l, i)"
                >
                  <!-- Cell Labels -->
                  <div class="absolute inset-x-2 top-2 flex justify-between items-start pointer-events-none opacity-20 group-hover:opacity-40 transition-opacity">
                    <span class="text-[8px] font-black uppercase tracking-tighter max-w-[60%] leading-none">{{ getRiskLevelLabel(getRiskLevel(l, i)) }}</span>
                    <span class="text-[10px] font-black">{{ getRiskScore(l, i) }}</span>
                  </div>

                  <!-- Badges Container -->
                  <div class="flex flex-wrap gap-1.5 justify-center items-center py-4">
                    <RiskBadge
                      v-for="(risk, idx) in getCellRisks(l, i)"
                      :key="risk.id"
                      :risk="risk"
                      :z-index="getCellRisks(l, i).length - idx"
                      @drag-start="onRiskDragStart"
                      @drag-end="onRiskDragEnd"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- X-axis Ticks -->
          <div class="flex mt-4 ml-0">
            <div v-for="i in impactLevels" :key="`x-${i}`" class="flex-1 flex flex-col items-center gap-1">
              <span class="text-xl font-black">{{ i }}</span>
              <span class="text-[10px] font-bold uppercase tracking-tighter text-center max-w-[80px] leading-tight">{{ getImpactLabel(i) }}</span>
            </div>
          </div>

          <!-- X-axis Label -->
          <div class="mt-8 text-center text-[10px] font-black uppercase tracking-[0.3em]">
            {{ t('riskProfile.impactLevel') }}
          </div>
        </div>
      </div>
    </div>

    <!-- Risk List Panel -->
    <UCard :ui="{ body: { padding: 'p-0' } }" class="overflow-hidden border-gray-200 dark:border-gray-800">
      <UTabs v-model="activeRiskTab" :items="tabItems" class="w-full">
        <template #content="{ item }">
          <div class="p-6 max-h-[600px] overflow-y-auto">
            
            <!-- Tab: Progress spreadsheet-like table -->
            <div v-if="item.key === 'progress'">
              <div class="overflow-x-auto border border-gray-200 dark:border-gray-800 rounded-xl shadow-inner bg-white dark:bg-gray-900">
                <table class="w-full text-left border-collapse text-md font-sans">
                  <thead>
                    <tr class="bg-gray-100/80 dark:bg-gray-800/80 border-b border-gray-200 dark:border-gray-800 font-bold uppercase tracking-wider text-[9px] text-gray-500 text-center">
                      <th class="py-4 px-3 border-r border-gray-200 dark:border-gray-800 text-left min-w-[40px]" rowspan="2">{{ t('riskProfile.table.no') }}</th>
                      <th class="py-4 px-4 border-r border-gray-200 dark:border-gray-800 text-left min-w-[300px]" rowspan="2">{{ t('riskProfile.table.riskEvent') }}</th>
                      <th class="py-2 px-3 border-r border-gray-200 dark:border-gray-800 border-b" colspan="4">{{ t('riskProfile.table.impactScale') }}</th>
                      <th class="py-2 px-3 border-r border-gray-200 dark:border-gray-800 border-b" colspan="4">{{ t('riskProfile.table.likelihoodScale') }}</th>
                      <th class="py-2 px-3 border-b" colspan="4">{{ t('riskProfile.table.riskLevel') }}</th>
                    </tr>
                    <tr class="bg-gray-50/50 dark:bg-gray-800/40 border-b border-gray-200 dark:border-gray-800 font-bold text-[8px] text-gray-400 text-center uppercase">
                      <!-- Dampak Q1-Q4 -->
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q1</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q2</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q3</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q4</th>
                      <!-- Kemungkinan Q1-Q4 -->
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q1</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q2</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q3</th>
                      <th class="py-2 px-2 border-r border-gray-200 dark:border-gray-800">Q4</th>
                      <!-- Level Q1-Q4 -->
                      <th class="py-2 px-3 border-r border-gray-200 dark:border-gray-800">Q1</th>
                      <th class="py-2 px-3 border-r border-gray-200 dark:border-gray-800">Q2</th>
                      <th class="py-2 px-3 border-r border-gray-200 dark:border-gray-800">Q3</th>
                      <th class="py-2 px-3">Q4</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr 
                      v-for="(risk, index) in filteredRisks" 
                      :key="risk.id" 
                      class="border-b border-gray-200 dark:border-gray-800 hover:bg-gray-50/50 dark:hover:bg-gray-800/20 text-center font-medium"
                    >
                      <td class="py-3.5 px-3 border-r border-gray-200 dark:border-gray-800 text-left font-bold text-gray-400">{{ index + 1 }}</td>
                      <td class="py-3.5 px-4 border-r border-gray-200 dark:border-gray-800 text-left font-bold leading-normal text-gray-700 dark:text-gray-200 max-w-[300px] whitespace-normal">
                        {{ risk.name }}
                      </td>
                      
                      <!-- Skala Dampak Q1-Q4 -->
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'impact', 'q1') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'impact', 'q2') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'impact', 'q3') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'impact', 'q4') }}</td>

                      <!-- Skala Kemungkinan Q1-Q4 -->
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'likelihood', 'q1') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'likelihood', 'q2') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'likelihood', 'q3') }}</td>
                      <td class="py-3.5 px-2 border-r border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 font-semibold">{{ getQVal(risk, 'likelihood', 'q4') }}</td>

                      <!-- Level Risiko Q1-Q4 (Cell Tints + Badge) -->
                      <td class="py-2.5 px-2.5 border-r border-gray-200 dark:border-gray-800" :style="getQLevelCellStyle(risk, 'q1')">
                        <span class="inline-block px-2.5 py-1 rounded text-[10px] font-black tracking-tight" :style="getQLevelBadgeStyle(risk, 'q1')">
                          {{ getQLevelLabel(risk, 'q1') }}
                        </span>
                      </td>
                      <td class="py-2.5 px-2.5 border-r border-gray-200 dark:border-gray-800" :style="getQLevelCellStyle(risk, 'q2')">
                        <span class="inline-block px-2.5 py-1 rounded text-[10px] font-black tracking-tight" :style="getQLevelBadgeStyle(risk, 'q2')">
                          {{ getQLevelLabel(risk, 'q2') }}
                        </span>
                      </td>
                      <td class="py-2.5 px-2.5 border-r border-gray-200 dark:border-gray-800" :style="getQLevelCellStyle(risk, 'q3')">
                        <span class="inline-block px-2.5 py-1 rounded text-[10px] font-black tracking-tight" :style="getQLevelBadgeStyle(risk, 'q3')">
                          {{ getQLevelLabel(risk, 'q3') }}
                        </span>
                      </td>
                      <td class="py-2.5 px-2.5" :style="getQLevelCellStyle(risk, 'q4')">
                        <span class="inline-block px-2.5 py-1 rounded text-[10px] font-black tracking-tight" :style="getQLevelBadgeStyle(risk, 'q4')">
                          {{ getQLevelLabel(risk, 'q4') }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Tabs: Priority / All List -->
            <div v-else>
              <TransitionGroup 
                name="list" 
                tag="div" 
                class="space-y-3"
              >
                <div
                  v-for="risk in getTabRisks(item.key)"
                  :key="risk.id"
                  class="group flex items-center gap-4 p-4 rounded-xl border transition-all duration-200 hover:shadow-md bg-white dark:bg-gray-900"
                  :class="getItemBorderClass(risk)"
                >
                  <!-- Formatted ID Badge -->
                  <div 
                    class="w-14 h-14 shrink-0 rounded-lg flex flex-col items-center justify-center shadow-lg font-mono text-white"
                    :style="{ background: riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].color }"
                  >
                    <span class="text-[8px] font-bold opacity-75 uppercase">{{ getPrefix(risk) }}</span>
                    <span class="text-sm font-black">{{ getNumber(risk) }}</span>
                  </div>

                  <!-- Info -->
                  <div class="flex-1 min-w-0">
                    <h4 class="text-sm font-bold truncate group-hover:text-primary transition-colors text-gray-800 dark:text-gray-100">
                      {{ risk.name }}
                    </h4>
                    <div class="flex items-center gap-3 mt-1 text-gray-500">
                      <span class="text-[10px] font-bold flex items-center gap-1">
                        {{ categoryIcons[risk.category] }} {{ risk.category }}
                      </span>
                      <span>·</span>
                      <span class="text-[10px] font-bold">
                        Severity: {{ risk.severity }}%
                      </span>
                    </div>
                  </div>

                  <!-- Score -->
                  <div class="text-right shrink-0 px-4 border-r border-gray-100 dark:border-gray-800">
                    <div class="text-[10px] font-black uppercase tracking-tighter leading-none" :style="{ color: riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].color }">
                      {{ getRiskLevelLabel(getRiskLevel(risk.likelihood, risk.impact)) }}
                    </div>
                    <div class="text-2xl font-black leading-tight text-gray-700 dark:text-gray-300">
                      {{ getRiskScore(risk.likelihood, risk.impact) }}
                    </div>
                  </div>

                  <!-- Actions -->
                  <div class="flex items-center gap-1.5">
                    <UModal 
                      :title="t('riskProfile.detailModal.title', { id: store.getFormattedId(risk) })"
                      :description="t('riskProfile.detailModal.desc')"
                      :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }"
                    >
                      <UButton
                        icon="i-heroicons-eye"
                        color="neutral"
                        variant="ghost"
                        size="sm"
                      />

                      <template #body>
                        <div class="space-y-6">
                          <!-- Status Banner -->
                          <div class="flex items-center justify-between p-4 rounded-xl border border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/50">
                            <div class="flex items-center gap-3">
                              <span class="text-3xl">{{ categoryIcons[risk.category] }}</span>
                              <div>
                                <div class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.category') }}</div>
                                <div class="text-sm font-bold text-gray-700 dark:text-gray-200">{{ risk.category }}</div>
                              </div>
                            </div>
                            <div class="text-right">
                              <div class="text-[10px] font-black uppercase tracking-widest mb-1 text-gray-400">{{ t('riskProfile.detailModal.riskLevel') }}</div>
                              <UBadge 
                                :style="{ backgroundColor: riskLevelConfig[getRiskLevel(risk.likelihood, risk.impact)].color, color: 'white' }"
                                size="sm"
                                class="font-black"
                              >
                                {{ getRiskLevelLabel(getRiskLevel(risk.likelihood, risk.impact)) }}
                              </UBadge>
                            </div>
                          </div>

                          <!-- Basic Info -->
                          <div class="space-y-1">
                            <div class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.riskEventName') }}</div>
                            <h3 class="text-xl font-black leading-tight text-gray-800 dark:text-gray-100">{{ risk.name }}</h3>
                          </div>

                          <!-- Assessment Grid -->
                          <div class="grid grid-cols-2 gap-6 pt-4 border-t border-gray-100 dark:border-gray-800">
                            <div class="space-y-1.5">
                              <div class="flex items-center justify-between">
                                <span class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.impactPeriod', { period: store.selectedPeriod }) }}</span>
                                <span class="text-md font-bold text-gray-700 dark:text-gray-300">{{ risk.impact }}/5</span>
                              </div>
                              <div class="text-sm font-bold text-gray-700 dark:text-gray-200">{{ getImpactLabel(risk.impact) }}</div>
                            </div>
                            <div class="space-y-1.5">
                              <div class="flex items-center justify-between">
                                <span class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.likelihoodPeriod', { period: store.selectedPeriod }) }}</span>
                                <span class="text-md font-bold text-gray-700 dark:text-gray-300">{{ risk.likelihood }}/5</span>
                              </div>
                              <div class="text-sm font-bold text-gray-700 dark:text-gray-200">{{ getLikelihoodLabel(risk.likelihood) }}</div>
                            </div>
                          </div>

                          <!-- Severity Progress -->
                          <div class="space-y-2 pt-4 border-t border-gray-100 dark:border-gray-800">
                            <div class="flex justify-between items-center">
                              <span class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.severityWeight') }}</span>
                              <span class="text-sm font-black text-gray-700 dark:text-gray-300">{{ risk.severity }}%</span>
                            </div>
                            <UMeter :value="risk.severity" color="primary" size="md" />
                          </div>

                          <!-- Description -->
                          <div v-if="risk.description" class="space-y-1 pt-4 border-t border-gray-100 dark:border-gray-800">
                            <div class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ t('riskProfile.detailModal.description') }}</div>
                            <p class="text-sm leading-relaxed italic text-gray-600 dark:text-gray-400">
                              "{{ risk.description }}"
                            </p>
                          </div>

                          <!-- Mitigation Link -->
                          <div class="pt-6 border-t border-gray-100 dark:border-gray-800">
                            <UButton
                              icon="i-heroicons-shield-check"
                              :label="t('riskProfile.detailModal.mitigationBtn')"
                              color="primary"
                              variant="soft"
                              class="w-full justify-center font-bold animate-pulse"
                              :to="`/mitigation?id=${risk.id}`"
                            />
                          </div>
                        </div>
                      </template>
                    </UModal>

                    <!-- Edit Risk Button -->
                    <UButton
                      icon="i-heroicons-pencil-square"
                      color="warning"
                      variant="ghost"
                      size="sm"
                      @click="handleOpenEditModal(risk)"
                    />
                    
                    <UButton
                      icon="i-heroicons-trash"
                      color="error"
                      variant="ghost"
                      size="sm"
                      @click="handleDeleteRisk(risk.id)"
                    /> 
                  </div>
                </div>  
              </TransitionGroup>
            </div>

          </div>
        </template>
      </UTabs>
    </UCard>

    <!-- Global Edit Risk Modal -->
    <UModal 
      v-model:open="store.isFormOpen"
      :title="store.selectedRisk ? t('riskProfile.editModal.title', { name: store.selectedRisk.name }) : t('riskProfile.editModal.titleDefault')"
      :description="t('riskProfile.editModal.desc')"
      :ui="{ content: 'sm:max-w-2xl' }"
    >
      <template #body>
        <div v-if="store.selectedRisk" class="space-y-4">
          <UFormField :label="t('riskProfile.addModal.name')" required>
            <UInput v-model="store.selectedRisk.name" class="w-full" />
          </UFormField>

          <div class="grid grid-cols-2 gap-4">
            <UFormField :label="t('riskProfile.addModal.category')">
              <USelect v-model="store.selectedRisk.category" :items="categoryOptions" class="w-full" />
            </UFormField>
            <UFormField :label="t('riskProfile.addModal.branch')">
              <USelect v-model="store.selectedRisk.branch" :items="store.branches" class="w-full" />
            </UFormField>
          </div>

          <!-- Quarterly residual inputs for editing -->
          <div class="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-gray-50/50 dark:bg-gray-900/50 space-y-4">
            <span class="block text-md font-black uppercase tracking-wider text-warning-500">{{ t('riskProfile.addModal.quarterlyTitle', { year: store.selectedYear }) }}</span>
            <div class="grid grid-cols-4 gap-4">
              <div v-for="q in ['q1', 'q2', 'q3', 'q4']" :key="q" class="space-y-2 border-r last:border-0 border-gray-200 dark:border-gray-700 pr-2">
                <span class="text-md font-black uppercase text-gray-500">{{ q }}</span>
                <UFormField :label="t('riskProfile.addModal.impact')">
                  <USelect v-model.number="store.selectedRisk[`impact_${q}`]" :items="[1,2,3,4,5]" class="w-full" />
                </UFormField>
                <UFormField :label="t('riskProfile.addModal.likelihood')">
                  <USelect v-model.number="store.selectedRisk[`likelihood_${q}`]" :items="[1,2,3,4,5]" class="w-full" />
                </UFormField>
              </div>
            </div>
          </div>

          <UFormField :label="t('riskProfile.addModal.description')">
            <UTextarea v-model="store.selectedRisk.description" class="w-full" />
          </UFormField>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3 w-full">
          <UButton 
            :label="t('common.cancel')" 
            color="neutral" 
            variant="ghost" 
            @click="store.isFormOpen = false" 
          />
          <UButton 
            :label="t('riskProfile.editModal.updateBtn')" 
            color="warning" 
            @click="submitEditRisk" 
          />
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import RiskBadge from './RiskBadge.vue'
import { useI18n } from '~/composables/useI18n'
import { 
  useRiskProfileStore, 
  riskLevelConfig, 
  categoryIcons, 
  impactLabels, 
  likelihoodLabels 
} from '~/stores/risk-profile'

const store = useRiskProfileStore()
const { getRiskLevel, getRiskScore } = store
const { t } = useI18n()
const toast = useToast()

// Local state for UI
const dragOverCell = ref(null)
const activeRiskTab = ref('priority')
const isAddModalOpen = ref(false)

// Form state for new risk
const newRisk = ref({
  name: '',
  category: 'Strategic',
  impact: 3,
  likelihood: 3,
  severity: 50,
  description: '',
  branch: 'Head Office',
  impact_q1: 3,
  impact_q2: 3,
  impact_q3: 3,
  impact_q4: 3,
  likelihood_q1: 3,
  likelihood_q2: 3,
  likelihood_q3: 3,
  likelihood_q4: 3
})

// Localized helper labels
function getRiskLevelLabel(levelKey) {
  return t(`riskProfile.riskLevelLabels.${levelKey}`) || riskLevelConfig[levelKey]?.label || levelKey
}

function getImpactLabel(val) {
  return t(`riskProfile.impactLabels.${val}`) || impactLabels[val] || val
}

function getLikelihoodLabel(val) {
  return t(`riskProfile.likelihoodLabels.${val}`) || likelihoodLabels[val] || val
}

// Options for selects
const categoryOptions = Object.keys(categoryIcons).map(cat => ({ 
  label: `${categoryIcons[cat]} ${cat}`, 
  value: cat 
}))

const impactOptions = Object.entries(impactLabels).map(([val, label]) => ({ 
  label: `${val} - ${getImpactLabel(val)}`, 
  value: Number(val) 
}))

const likelihoodOptions = Object.entries(likelihoodLabels).map(([val, label]) => ({ 
  label: `${val} - ${getLikelihoodLabel(val)}`, 
  value: Number(val) 
}))

function submitNewRisk() {
  if (!newRisk.value.name) {
    toast.add({
      title: t('riskProfile.toasts.validationError'),
      description: t('riskProfile.toasts.nameRequired'),
      color: 'error'
    })
    return
  }

  // Map fields into structured assessments
  const payload = {
    name: newRisk.value.name,
    category: newRisk.value.category,
    impact: newRisk.value.impact_q1 || 3,
    likelihood: newRisk.value.likelihood_q1 || 3,
    severity: 50,
    description: newRisk.value.description,
    branch: newRisk.value.branch,
    assessments: [
      {
        year: store.selectedYear,
        impact_q1: newRisk.value.impact_q1,
        impact_q2: newRisk.value.impact_q2,
        impact_q3: newRisk.value.impact_q3,
        impact_q4: newRisk.value.impact_q4,
        likelihood_q1: newRisk.value.likelihood_q1,
        likelihood_q2: newRisk.value.likelihood_q2,
        likelihood_q3: newRisk.value.likelihood_q3,
        likelihood_q4: newRisk.value.likelihood_q4
      }
    ]
  }

  store.addRisk(payload)
  toast.add({
    title: t('riskProfile.toasts.riskAdded'),
    description: t('riskProfile.toasts.riskAddedDesc', { name: newRisk.value.name }),
    color: 'success',
    icon: 'i-heroicons-check-circle'
  })
  
  // Reset form
  newRisk.value = {
    name: '',
    category: 'Strategic',
    impact: 3,
    likelihood: 3,
    severity: 50,
    description: '',
    branch: 'Head Office',
    impact_q1: 3,
    impact_q2: 3,
    impact_q3: 3,
    impact_q4: 3,
    likelihood_q1: 3,
    likelihood_q2: 3,
    likelihood_q3: 3,
    likelihood_q4: 3
  }
  isAddModalOpen.value = false
}

// Grid levels
const likelihoodLevels = [5, 4, 3, 2, 1]
const impactLevels = [1, 2, 3, 4, 5]

// Map store state to local-like computed
const risks = computed(() => store.risks)
const selectedBranch = computed({
  get: () => store.selectedBranch,
  set: (val) => store.selectedBranch = val
})

const filteredRisks = computed(() => {
  if (selectedBranch.value === 'All Branches' || selectedBranch.value === t('riskProfile.allBranches')) return risks.value
  return risks.value.filter(r => r.branch === selectedBranch.value)
})

const totalRisks = computed(() => filteredRisks.value.length)

const priorityRisks = computed(() => {
  return filteredRisks.value
    .filter(r => {
      const level = store.getRiskLevel(r.likelihood, r.impact)
      return riskLevelConfig[level].priority
    })
    .sort((a, b) => {
      const scoreA = getRiskScore(a.likelihood, a.impact)
      const scoreB = getRiskScore(b.likelihood, b.impact)
      if (scoreB !== scoreA) return scoreB - scoreA
      return b.severity - a.severity
    })
})

const priorityCount = computed(() => priorityRisks.value.length)

const branchOptions = computed(() => [t('riskProfile.allBranches'), ...store.branches])

const tabItems = computed(() => [
  {
    key: 'priority',
    value: 'priority',
    label: t('riskProfile.tabs.priority', { count: priorityRisks.value.length }),
    icon: 'i-heroicons-fire'
  },
  {
    key: 'all',
    value: 'all',
    label: t('riskProfile.tabs.all', { count: totalRisks.value }),
    icon: 'i-heroicons-list-bullet'
  },
  {
    key: 'progress',
    value: 'progress',
    label: t('riskProfile.tabs.progress'),
    icon: 'i-heroicons-arrow-trending-up'
  }
])

// Helpers
const getPrefix = (risk) => store.getFormattedId(risk).split('-')[0]
const getNumber = (risk) => store.getFormattedId(risk).split('-')[1]

function getCellRisks(likelihood, impact) {
  return filteredRisks.value
    .filter(r => r.likelihood === likelihood && r.impact === impact)
    .sort((a, b) => b.severity - a.severity)
}

function getTabRisks(key) {
  if (key === 'priority') return priorityRisks.value
  return [...filteredRisks.value].sort((a, b) => b.severity - a.severity)
}

const getCellBgClass = (l, i) => {
  const level = store.getRiskLevel(l, i)
  const map = {
    'low': 'bg-green-500/10 dark:bg-green-500/5 hover:bg-green-500/20',
    'low-moderate': 'bg-lime-500/15 dark:bg-lime-500/10 hover:bg-lime-500/25',
    'moderate': 'bg-yellow-500/20 dark:bg-yellow-500/15 hover:bg-yellow-500/30',
    'moderate-high': 'bg-orange-500/25 dark:bg-orange-500/20 hover:bg-orange-500/35',
    'high': 'bg-red-500/30 dark:bg-red-500/25 hover:bg-red-500/40'
  }
  return map[level]
}

const getItemBorderClass = (risk) => {
  const level = store.getRiskLevel(risk.likelihood, risk.impact)
  const map = {
    'low': 'border-green-500/20 bg-green-500/5',
    'low-moderate': 'border-lime-500/20 bg-lime-500/5',
    'moderate': 'border-yellow-500/20 bg-yellow-500/5',
    'moderate-high': 'border-orange-500/20 bg-orange-500/5',
    'high': 'border-red-500/20 bg-red-500/5'
  }
  return map[level]
}

// Mapped quarterly values helpers
function getQVal(risk, type, quarter) {
  const ast = risk.assessments?.find(a => a.year === store.selectedYear)
  if (!ast) return risk[type] || 0
  return ast[`${type}_${quarter}`] || 0
}

function getQLevelLabel(risk, quarter) {
  const impact = getQVal(risk, 'impact', quarter)
  const likelihood = getQVal(risk, 'likelihood', quarter)
  const level = store.getRiskLevel(likelihood, impact)
  return getRiskLevelLabel(level)
}

function getQLevelBadgeStyle(risk, quarter) {
  const impact = getQVal(risk, 'impact', quarter)
  const likelihood = getQVal(risk, 'likelihood', quarter)
  const level = store.getRiskLevel(likelihood, impact)
  return {
    backgroundColor: riskLevelConfig[level]?.color || '#4CAF50',
    color: '#ffffff'
  }
}

function getQLevelCellStyle(risk, quarter) {
  const impact = getQVal(risk, 'impact', quarter)
  const likelihood = getQVal(risk, 'likelihood', quarter)
  const level = store.getRiskLevel(likelihood, impact)
  return {
    backgroundColor: (riskLevelConfig[level]?.color || '#4CAF50') + '15' // Tint cell background
  }
}

function submitEditRisk() {
  store.updateRisk(store.selectedRisk);
  toast.add({ 
    title: t('riskProfile.toasts.riskUpdated'), 
    description: t('riskProfile.toasts.riskUpdatedDesc'), 
    color: 'success' 
  });
  store.isFormOpen = false;
}

function handleOpenEditModal(risk) {
  const ast = risk.assessments?.find(a => a.year === store.selectedYear)
  const mappedRisk = {
    ...risk,
    impact_q1: ast ? ast.impact_q1 : risk.impact,
    impact_q2: ast ? ast.impact_q2 : risk.impact,
    impact_q3: ast ? ast.impact_q3 : risk.impact,
    impact_q4: ast ? ast.impact_q4 : risk.impact,
    likelihood_q1: ast ? ast.likelihood_q1 : risk.likelihood,
    likelihood_q2: ast ? ast.likelihood_q2 : risk.likelihood,
    likelihood_q3: ast ? ast.likelihood_q3 : risk.likelihood,
    likelihood_q4: ast ? ast.likelihood_q4 : risk.likelihood
  }
  store.openEditModal(mappedRisk)
}

// Event Handlers
function onRiskDragStart(risk) {
  // Drag start
}

function onRiskDragEnd() {
  dragOverCell.value = null
}

function onDragOver(e, l, i) {
  e.dataTransfer.dropEffect = 'move'
  dragOverCell.value = `${l}-${i}`
}

function onDragLeave() {
  dragOverCell.value = null
}

function onDrop(e, newLikelihood, newImpact) {
  dragOverCell.value = null
  try {
    const riskDataStr = e.dataTransfer?.getData('application/json')
    if (!riskDataStr) return

    const droppedRisk = JSON.parse(riskDataStr)
    if (droppedRisk.likelihood === newLikelihood && droppedRisk.impact === newImpact) return

    store.updateRisk({ ...droppedRisk, likelihood: newLikelihood, impact: newImpact })
    
    toast.add({
      title: t('riskProfile.toasts.positionUpdated'),
      description: t('riskProfile.toasts.positionUpdatedDesc', { name: droppedRisk.name, l: newLikelihood, i: newImpact, period: store.selectedPeriod }),
      color: 'primary',
      icon: 'i-heroicons-arrows-right-left'
    })
  } catch (err) {
    console.error('Drop failed:', err)
  }
}

function handleDeleteRisk(id) {
  const risk = store.risks.find(r => r.id === id)
  if (!risk) return
  
  store.deleteRisk(id)
  toast.add({
    title: t('riskProfile.toasts.riskDeleted'),
    description: t('riskProfile.toasts.riskDeletedDesc', { name: risk.name }),
    color: 'error',
    icon: 'i-heroicons-trash'
  })
}

</script>

<style scoped>
.list-enter-active { transition: all 0.3s ease-out; }
.list-leave-active { transition: all 0.2s ease-in; }
.list-enter-from { opacity: 0; transform: translateX(-20px); }
.list-leave-to { opacity: 0; transform: translateX(20px); }
.list-move { transition: transform 0.3s ease; }

.grid-cell:hover {
  z-index: 50;
}
</style>
