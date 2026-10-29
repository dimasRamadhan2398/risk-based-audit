<template>
  <div class="p-4 md:p-6 space-y-6 min-h-screen">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 mb-1">
          <NuxtLink to="/risk-profile" class="hover:text-primary-600 dark:hover:text-primary-400 transition-colors">{{ t('rcm.breadcrumb.riskProfile') }}</NuxtLink>
          <span>/</span>
          <span class="text-slate-800 dark:text-slate-200 font-medium">{{ t('rcm.breadcrumb.rcm') }}</span>
        </div>
        <h1 class="text-2xl md:text-3xl font-bold text-slate-900 dark:text-white tracking-tight">
          {{ t('rcm.header.title') }}
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          {{ t('rcm.header.subtitle') }}
        </p>
      </div>

      <!-- Actions / Filters -->
      <div class="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center justify-end gap-3 w-full md:w-auto">
        <USelect
          v-model="rcmStore.selectedYear"
          :items="yearOptions"
          value-key="value"
          size="md"
          class="w-full sm:w-36"
        />

        <USelect
          v-model="rcmStore.selectedDepartment"
          :items="departmentOptions"
          value-key="value"
          size="md"
          class="w-full sm:w-64"
        />

        <UButton v-if="canWrite"
          color="primary"
          variant="solid"
          class="font-medium shadow-sm w-full sm:w-auto justify-center"
          @click="openAddModal"
        >
          <UIcon name="i-lucide-plus" class="size-4 mr-1.5" />
          {{ t('rcm.header.addControlMatrix') }}
        </UButton>
      </div>
    </div>

    <!-- Top Cards Grid: Yearly Internal Control Effectiveness & COSO 5 Dimensions Breakdown -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Card 1: Yearly Internal Control Effectiveness (2 Cols) -->
      <div class="lg:col-span-2 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-sm font-semibold tracking-wider text-slate-400 dark:text-slate-400">{{ t('rcm.cards.effectiveness.badge') }}</span>
            </div>
            <span class="text-sm font-medium text-slate-500 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-md border border-transparent dark:border-slate-700">
              {{ t('rcm.cards.effectiveness.periodInfo', { year: rcmStore.selectedYear, department: selectedDepartmentLabel }) }}
            </span>
          </div>
          <h2 class="text-xl font-bold text-slate-900 dark:text-white mt-1">{{ t('rcm.cards.effectiveness.title') }}</h2>
        </div>

        <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-50 dark:bg-slate-800/60 rounded-xl p-4 border border-slate-100 dark:border-slate-700/60">
          <!-- Big Score % -->
          <div class="flex flex-col justify-center border-b md:border-b-0 md:border-r border-slate-200 dark:border-slate-700 pb-3 md:pb-0 md:pr-4">
            <span class="text-sm text-slate-500 dark:text-slate-400 font-medium">{{ t('rcm.cards.effectiveness.realScore') }}</span>
            <div class="flex items-baseline gap-2 mt-1">
              <span class="text-2xl font-extrabold tracking-tight text-slate-900 dark:text-white">
                {{ rcmStore.internalControlEffectiveness }}%
              </span>
            </div>
            <div class="mt-2">
              <span :class="getRatingBadgeClass(rcmStore.effectivenessRating.rating)">
                {{ rcmStore.effectivenessRating.rating }}
              </span>
            </div>
          </div>

          <!-- Synchronized Risk Counts & Interpretation -->
          <div class="md:col-span-2 flex flex-col justify-center space-y-2">
            <div class="flex items-center justify-between text-sm">
              <span class="text-slate-600 dark:text-slate-300 font-medium">{{ t('rcm.cards.effectiveness.inherentRiskLabel') }}</span>
              <span class="font-bold text-slate-900 dark:text-slate-100 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                {{ t('rcm.cards.effectiveness.riskCount', { count: rcmStore.totalInherentRisk }) }}
              </span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-slate-600 dark:text-slate-300 font-medium">{{ t('rcm.cards.effectiveness.residualRiskLabel') }}</span>
              <span class="font-bold text-red-600 dark:text-red-400 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                {{ t('rcm.cards.effectiveness.riskCount', { count: rcmStore.totalResidualRisk }) }}
              </span>
            </div>
            <UAlert
              :color="(rcmStore.effectivenessRating.alertColor as any) || 'info'"
              variant="subtle"
              icon="i-lucide-info"
              :title="t('rcm.cards.effectiveness.cosoInterpretation', { rating: getLocalizedRating(rcmStore.internalControlEffectiveness) })"
              :description="getLocalizedInterpretation(rcmStore.internalControlEffectiveness)"
              class="rounded-xl shadow-xs border-none! bg-transparent! border-transparent!"
            />
          </div>
        </div>

        <div class="mt-3 text-sm text-slate-400 dark:text-slate-400 flex items-center gap-4">
          <UIcon name="i-lucide-check-circle-2" class="size-10 text-emerald-500" />
          <span>{{ t('rcm.cards.effectiveness.syncNote') }}</span>
        </div>
      </div>

      <!-- Card 2: COSO 2013 5 Dimensions Summary (1 Col) -->
      <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-bold text-slate-900 dark:text-white">{{ t('rcm.cards.coso.title') }}</h3>
            <span class="text-md font-bold text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/60 px-2.5 py-1 rounded-full border border-primary-100 dark:border-primary-800">
              {{ rcmStore.cosoAverages.totalWeighted }}%
            </span>
          </div>

          <div class="mt-4 space-y-3">
            <div v-for="dim in cosoDimensions" :key="dim.key" class="space-y-1">
              <div class="flex justify-between text-md">
                <span class="font-medium text-slate-700 dark:text-slate-300">{{ t(`rcm.dimensions.${dim.i18nKey}`) }}</span>
                <span class="font-bold text-slate-900 dark:text-white">{{ getDimAverage(dim.key) }}%</span>
              </div>
              <div class="w-full bg-slate-100 dark:bg-slate-800 h-2 rounded-full overflow-hidden">
                <div
                  class="bg-primary-600 dark:bg-primary-500 h-full rounded-full transition-all duration-300"
                  :style="{ width: `${getDimAverage(dim.key)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-md text-slate-500 dark:text-slate-400">
          <span>{{ t('rcm.cards.coso.totalControls') }}</span>
          <span class="font-bold text-slate-900 dark:text-white">{{ t('rcm.cards.coso.itemsCount', { count: rcmStore.filteredRCMList.length }) }}</span>
        </div>
      </div>
    </div>

    <!-- Collapsible Standard Interpretation Table Reference -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
      <button
        class="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-slate-50 dark:hover:bg-slate-800/60 transition-colors"
        @click="showRatingTable = !showRatingTable"
      >
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-book-open" class="size-5 text-primary-600 dark:text-primary-400" />
          <span class="font-bold text-slate-900 dark:text-white text-sm md:text-base">{{ t('rcm.interpretationTable.toggle') }}</span>
        </div>
        <UIcon :name="showRatingTable ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5 text-slate-400 dark:text-slate-500" />
      </button>

      <div v-if="showRatingTable" class="px-6 pb-6 border-t border-slate-100 dark:border-slate-800 pt-4">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-md border-collapse">
            <thead>
              <tr class="bg-slate-50 dark:bg-slate-800/80 border-b border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 font-semibold">
                <th class="py-2.5 px-4 rounded-l-lg whitespace-nowrap">{{ t('rcm.interpretationTable.headers.totalWeightedScore') }}</th>
                <th class="py-2.5 px-4 whitespace-nowrap">{{ t('rcm.interpretationTable.headers.rating') }}</th>
                <th class="py-2.5 px-4 rounded-r-lg">{{ t('rcm.interpretationTable.headers.interpretation') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-green-700 dark:text-green-400 whitespace-nowrap">90 – 100%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block" :class="getControlEffectivenessColorClass('Highly Effective')">{{ t('rcm.interpretationTable.ratings.highlyEffective') }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ t('rcm.interpretationTable.descriptions.highlyEffective') }}</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-lime-700 dark:text-lime-400 whitespace-nowrap">80 – 89%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block" :class="getControlEffectivenessColorClass('Effective')">{{ t('rcm.interpretationTable.ratings.effective') }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ t('rcm.interpretationTable.descriptions.effective') }}</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-yellow-700 dark:text-yellow-400 whitespace-nowrap">70 – 79%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block" :class="getControlEffectivenessColorClass('Moderately Effective')">{{ t('rcm.interpretationTable.ratings.moderatelyEffective') }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ t('rcm.interpretationTable.descriptions.moderatelyEffective') }}</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-orange-700 dark:text-orange-400 whitespace-nowrap">60 – 69%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block" :class="getControlEffectivenessColorClass('Weak')">{{ t('rcm.interpretationTable.ratings.weak') }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ t('rcm.interpretationTable.descriptions.weak') }}</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-red-700 dark:text-red-400 whitespace-nowrap">&lt; 60%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block" :class="getControlEffectivenessColorClass('Ineffective')">{{ t('rcm.interpretationTable.ratings.ineffective') }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ t('rcm.interpretationTable.descriptions.ineffective') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Main Risk Control Matrix Table -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden p-4 sm:p-6 space-y-4">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h3 class="text-lg font-bold text-slate-900 dark:text-white">{{ t('rcm.table.title') }}</h3>
          <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('rcm.table.subtitle') }}</p>
        </div>
        <div class="relative w-full md:w-64">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('rcm.table.searchPlaceholder')"
            class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg pl-9 pr-3 py-1.5 text-sm text-slate-700 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          <UIcon name="i-lucide-search" class="absolute left-3 top-2.5 size-3.5 text-slate-400 dark:text-slate-500" />
        </div>
      </div>

      <div class="rcm-table w-full min-w-0">
      <TableEntities
        :data="filteredList"
        :columns="rcmStore.columns"
        :items-per-page="10"
        table-layout="auto"
        min-width="1630px"
        :ui="{
          root: 'relative overflow-visible',
          tr: 'hover:bg-slate-50/60 dark:hover:bg-slate-800/40 transition-colors group'
        }"
        :empty-state="{
          icon: 'i-lucide-shield-alert',
          label: t('rcm.table.empty.label'),
          description: t('rcm.table.empty.description')
        }"
      >
        <!-- Risk Code & Event -->
        <template #risk_code-cell="{ row }">
          <div class="w-full whitespace-normal break-words space-y-1 pr-1">
            <span class="inline-block px-2 py-0.5 bg-primary-50 dark:bg-primary-950/60 text-primary-700 dark:text-primary-300 font-bold rounded text-xs">
              {{ row.original.risk_code }}
            </span>
            <p class="font-medium text-slate-900 dark:text-white text-sm leading-snug break-words" :title="row.original.risk_event">
              {{ row.original.risk_event }}
            </p>
          </div>
        </template>

        <!-- Control Code & Description -->
        <template #control_code-cell="{ row }">
          <div class="w-full whitespace-normal break-words space-y-1 pr-1">
            <span class="inline-block px-2 py-0.5 bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200 font-extrabold rounded text-xs border border-slate-200 dark:border-slate-700">
              {{ row.original.control_code }}
            </span>
            <p class="text-slate-600 dark:text-slate-300 text-sm leading-relaxed break-words" :title="row.original.control_description">
              {{ row.original.control_description }}
            </p>
          </div>
        </template>

        <!-- Department & PIC -->
        <template #department-cell="{ row }">
          <div class="min-w-[140px] whitespace-normal break-words">
            <p class="text-sm text-slate-800 dark:text-slate-100 font-semibold mb-0.5">{{ row.original.department }}</p>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('rcm.table.cells.pic') }} <strong class="text-slate-700 dark:text-slate-200">{{ row.original.control_owner }}</strong></p>
          </div>
        </template>

        <!-- Design Rating (1-5) -->
        <template #design_effectiveness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700 dark:text-slate-200">
            <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs inline-block" :title="`${row.original.design_effectiveness_rating * 4}%`">
              {{ row.original.design_effectiveness_rating }}
            </span>
          </div>
        </template>

        <!-- Operating Rating (1-5) -->
        <template #operating_effectiveness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700 dark:text-slate-200">
            <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs inline-block" :title="`${row.original.operating_effectiveness_rating * 4}%`">
              {{ row.original.operating_effectiveness_rating }}
            </span>
          </div>
        </template>

        <!-- Coverage Rating (1-5) -->
        <template #coverage_completeness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700 dark:text-slate-200">
            <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs inline-block" :title="`${row.original.coverage_completeness_rating * 4}%`">
              {{ row.original.coverage_completeness_rating }}
            </span>
          </div>
        </template>

        <!-- Timeliness Rating (1-5) -->
        <template #timeliness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700 dark:text-slate-200">
            <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs inline-block" :title="`${row.original.timeliness_rating * 4}%`">
              {{ row.original.timeliness_rating }}
            </span>
          </div>
        </template>

        <!-- Automation Rating (1-5) -->
        <template #automation_monitoring_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700 dark:text-slate-200">
            <span class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 text-xs inline-block" :title="`${row.original.automation_monitoring_rating * 4}%`">
              {{ row.original.automation_monitoring_rating }}
            </span>
          </div>
        </template>

        <!-- Total Score (%) -->
        <template #total_weighted_score-cell="{ row }">
          <div class="text-center">
            <span class="text-xs font-extrabold text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/60 px-2.5 py-1 rounded-lg border border-primary-100 dark:border-primary-800 inline-block">
              {{ row.original.total_weighted_score }}%
            </span>
          </div>
        </template>

        <!-- Rating Efektivitas Column -->
        <template #rating-cell="{ row }">
          <div class="text-center whitespace-nowrap">
            <span :class="getRatingBadgeClass(getItemRating(row.original.total_weighted_score).rating)">
              {{ getLocalizedRating(row.original.total_weighted_score) }}
            </span>
          </div>
        </template>

        <!-- Actions -->
        <template #actions-cell="{ row }">
          <div class="flex items-center justify-end gap-1">
            <UTooltip v-if="canWrite" :text="t('rcm.table.actions.edit')">
            <UButton
              icon="i-lucide-edit-3"
              color="neutral"
              variant="ghost"
              size="xs"
              :title="t('rcm.table.actions.edit')"
              @click="openEditModal(row.original)"
            />
            </UTooltip>  
            <UTooltip v-if="canWrite" :text="t('rcm.table.actions.delete')">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="xs"
              :title="t('rcm.table.actions.delete')"
              @click="confirmDelete(row.original.id)"
            />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
      </div>
    </div>

    <!-- Add / Edit Modal -->
    <UModal
      :open="isModalOpen"
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
      @update:open="(open: boolean) => { if (!open) isModalOpen = false }"
    >
      <template #content>
        <div class="relative flex max-h-[92vh] flex-col transition-colors duration-300">
          <!-- Header -->
          <div class="shrink-0 flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
            <div>
              <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
                {{ isEditMode ? t('rcm.modal.editTitle') : t('rcm.modal.addTitle') }}
              </h3>
              <p class="mt-0.5 text-xs text-[var(--text-muted)]">
                {{ t('rcm.modal.manageTitle') }} — {{ steps[currentStep]?.label }} ({{ t('rcm.modal.stepNum', { current: currentStep + 1, total: steps.length }) }})
              </p>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark-20-solid"
              class="-my-1"
              aria-label="Tutup form Risk Control Matrix"
              :title="t('common.close')"
              @click="isModalOpen = false" 
            />
          </div>

          <!-- Stepper -->
          <div class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
            <div class="flex min-w-max items-center gap-3 px-1 sm:gap-4">
              <template
                v-for="(step, idx) in steps"
                :key="step.key"
              >
                <div
                  class="group flex shrink-0 cursor-pointer select-none items-center gap-2 sm:gap-2.5"
                  data-testid="form-step"
                  @click="goToStep(idx)"
                >
                  <div
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 sm:h-8 sm:w-8"
                    :class="currentStep === idx
                      ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105'
                      : currentStep > idx
                        ? 'bg-secondary text-white shadow-sm'
                        : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
                  >
                    <UIcon
                      v-if="currentStep > idx"
                      name="i-heroicons-check"
                      class="h-3.5 w-3.5 stroke-[3]"
                    />
                    <span v-else>{{ idx + 1 }}</span>
                  </div>
                  <span
                    class="whitespace-nowrap text-xs font-semibold transition-colors"
                    :class="currentStep === idx
                      ? 'text-primary font-bold'
                      : currentStep > idx
                        ? 'text-secondary font-semibold'
                        : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'"
                  >
                    {{ step.label }}
                  </span>
                </div>
                <div
                  v-if="idx < steps.length - 1"
                  class="h-0.5 w-5 shrink-0 rounded transition-all duration-300 sm:w-8"
                  :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
                />
              </template>
            </div>
          </div>

          <!-- Body -->
          <div
            ref="bodyRef"
            class="flex-1 min-h-0 overflow-y-auto p-4 sm:p-6 space-y-6"
          >
            <!-- Step 1: Corporate Risk Profile -->
            <div
              v-if="currentStep === 0"
              class="space-y-5"
            >
              <UCard
                :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
                class="border border-[var(--border-main)]"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-shield-exclamation"
                      class="h-5 w-5 text-amber-500"
                    />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">
                      {{ t('rcm.modal.steps.riskProfile') }}
                    </h3>
                  </div>
                </template>

                <!-- Risk selection from Corporate Risk Profile -->
                <UFormField
                  :label="t('rcm.modal.riskSelectLabel')"
                  required
                >
                  <USelectMenu
                    v-model="selectedRiskId"
                    :items="riskOptionsForModal"
                    value-key="value"
                    class="w-full"
                    :placeholder="t('rcm.modal.riskSelectPlaceholder')"
                    @update:model-value="onRiskSelected"
                  >
                    <template #item="{ item }">
                      <div class="flex items-center gap-2 max-w-full w-full py-0.5">
                        <span class="text-xs font-bold text-primary-600 dark:text-primary-400 shrink-0">[{{ item.code }}]</span>
                        <span class="truncate text-sm font-medium">{{ item.name }}</span>
                        <span class="text-xs text-slate-400 shrink-0 ml-auto">({{ item.branch }})</span>
                      </div>
                    </template>
                  </USelectMenu>
                </UFormField>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                  <UFormField
                    :label="t('rcm.modal.riskCodeLabel')"
                  >
                    <UInput
                      v-model="formData.risk_code"
                      disabled
                      class="w-full font-bold"
                    />
                  </UFormField>
                  <UFormField
                    :label="t('rcm.modal.departmentLabel')"
                  >
                    <UInput
                      v-model="formData.department"
                      disabled
                      class="w-full"
                    />
                  </UFormField>
                </div>

                <UFormField
                  :label="t('rcm.modal.riskEventLabel')"
                >
                  <UTextarea
                    v-model="formData.risk_event"
                    :rows="3"
                    disabled
                    class="w-full"
                  />
                </UFormField>
              </UCard>
            </div>

            <!-- Step 2: Mitigation Plan & Control -->
            <div
              v-else-if="currentStep === 1"
              class="space-y-5"
            >
              <UCard
                :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
                class="border border-[var(--border-main)]"
              >
                <template #header>
                  <div class="flex items-center gap-2">
                    <UIcon
                      name="i-heroicons-clipboard-document-check"
                      class="h-5 w-5 text-amber-500"
                    />
                    <h3 class="text-lg font-bold text-[var(--text-main)]">
                      {{ t('rcm.modal.steps.controlMitigation') }}
                    </h3>
                  </div>
                </template>

                <!-- Synchronized Control & Mitigation Selection from Risk Mitigation Plans & Controls -->
                <UFormField
                  v-if="availableMitigations.length > 0"
                  :label="t('rcm.modal.controlSelectLabel')"
                >
                  <USelectMenu
                    v-model="selectedMitigationId"
                    :items="mitigationOptionsForModal"
                    value-key="value"
                    class="w-full"
                    :placeholder="t('rcm.modal.controlSelectPlaceholder')"
                    @update:model-value="onControlSelected"
                  />
                </UFormField>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                  <UFormField
                    :label="t('rcm.modal.controlCodeLabel')"
                    required
                  >
                    <UInput
                      v-model="formData.control_code"
                      :placeholder="t('rcm.modal.controlCodePlaceholder')"
                      disabled
                      class="w-full font-bold"
                    />
                  </UFormField>
                  <UFormField
                    :label="t('rcm.modal.controlOwnerLabel')"
                    required
                  >
                    <UInput
                      v-model="formData.control_owner"
                      :placeholder="t('rcm.modal.controlOwnerPlaceholder')"
                      disabled
                      class="w-full"
                    />
                  </UFormField>
                </div>

                <UFormField
                  :label="t('rcm.modal.controlDescLabel')"
                  required
                >
                  <UTextarea
                    v-model="formData.control_description"
                    :rows="3"
                    :placeholder="t('rcm.modal.controlDescPlaceholder')"
                    disabled
                    class="w-full"
                  />
                </UFormField>
              </UCard>
            </div>

            <!-- Step 3: Evaluation of 5 Internal Control Dimensions (COSO) -->
            <div
              v-else-if="currentStep === 2"
              class="space-y-5"
            >
              <UCard
                :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
                class="border border-[var(--border-main)]"
              >
                <template #header>
                  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                    <div class="flex items-center gap-2">
                      <UIcon
                        name="i-heroicons-chart-bar"
                        class="h-5 w-5 text-amber-500"
                      />
                      <h3 class="text-lg font-bold text-[var(--text-main)]">
                        {{ t('rcm.modal.evaluationTitle') }}
                      </h3>
                    </div>
                    <span class="text-xs font-medium text-[var(--text-muted)] bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-md border border-slate-200 dark:border-slate-700 self-start sm:self-auto">
                      {{ t('rcm.modal.evaluationWeight') }}
                    </span>
                  </div>
                </template>

                <div class="space-y-3.5">
                  <div
                    v-for="dim in cosoModalDimensions"
                    :key="dim.key"
                    class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2.5"
                  >
                    <div class="flex items-start sm:items-center justify-between gap-2">
                      <div>
                        <span class="font-bold text-slate-800 dark:text-slate-200 text-sm sm:text-base">{{ dim.name }}</span>
                        <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-0.5">{{ dim.desc }}</p>
                      </div>
                      <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2.5 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-sm shrink-0">
                        {{ (formData[dim.key] || 1) * 4 }}%
                      </span>
                    </div>
                    <div class="grid grid-cols-5 gap-1.5 sm:gap-2">
                      <button
                        v-for="star in [1, 2, 3, 4, 5]"
                        :key="star"
                        type="button"
                        class="py-1.5 rounded-lg text-xs sm:text-sm font-semibold transition-all cursor-pointer"
                        :class="getRatingBtnClass(formData[dim.key] || 1, star)"
                        @click="formData[dim.key] = star"
                      >
                        {{ star }} <span class="hidden sm:inline">({{ star * 4 }}%)</span>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Total Score Summary in Modal -->
                <div class="bg-primary-600 dark:bg-primary-700 text-white p-4 sm:p-5 rounded-xl shadow-md flex items-center justify-between mt-4">
                  <div>
                    <span class="text-sm sm:text-base font-bold text-white">{{ t('rcm.modal.totalWeightedScore') }}</span>
                    <p class="text-xs sm:text-sm text-primary-100 mt-0.5">
                      {{ t('rcm.modal.interpretation') }} <strong class="text-white font-extrabold">{{ getLocalizedRating(calculatedModalScorePercent) }}</strong>
                    </p>
                  </div>
                  <span class="text-2xl sm:text-3xl font-extrabold text-white">{{ calculatedModalScorePercent }}%</span>
                </div>
              </UCard>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex flex-col-reverse items-stretch justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:flex-row sm:items-center">
            <div class="flex flex-col-reverse gap-2 sm:flex-row">
              <UButton
                :label="t('rcm.modal.cancel')"
                color="neutral"
                variant="ghost"
                class="w-full sm:w-auto"
                @click="isModalOpen = false"
              />
            </div>

            <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
              <UButton
                v-if="currentStep > 0"
                icon="i-heroicons-arrow-left"
                :label="t('rcm.modal.previous')"
                color="neutral"
                variant="outline"
                class="flex-1 sm:flex-none"
                @click="currentStep--"
              />
              <UButton
                v-if="currentStep < steps.length - 1"
                trailing-icon="i-heroicons-arrow-right"
                :label="t('rcm.modal.next')"
                color="primary"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                @click="handleNext"
              />
              <UButton
                v-else
                icon="i-heroicons-check"
                :label="t('rcm.modal.save')"
                color="primary"
                class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
                @click="saveForm"
              />
            </div>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<style scoped>
/* Freeze the first two columns of the RCM table during horizontal scroll ONLY on desktop.
   On mobile/tablet (< 1280px), disable sticky so all columns (Department, PIC, Ratings, Actions)
   can be freely scrolled to and are not obscured by frozen columns. */
@media (min-width: 1280px) {
  .rcm-table :deep(th:nth-child(-n + 2)),
  .rcm-table :deep(td:nth-child(-n + 2)) {
    position: sticky;
  }

  .rcm-table :deep(th:nth-child(1)),
  .rcm-table :deep(td:nth-child(1)) {
    left: 0;
  }

  .rcm-table :deep(th:nth-child(2)),
  .rcm-table :deep(td:nth-child(2)) {
    left: 280px;
    border-right: 1px solid var(--border-main);
  }

  /* Opaque fills so scrolled cells cannot bleed through. The theme variables
     already flip under .dark, so no dark-mode override is needed. */
  .rcm-table :deep(td:nth-child(-n + 2)) {
    z-index: 1;
    background-color: var(--bg-main);
  }

  .rcm-table :deep(th:nth-child(-n + 2)) {
    z-index: 2;
    background-color: var(--bg-surface);
  }

  .rcm-table :deep(tbody tr:hover td:nth-child(-n + 2)) {
    background-color: var(--bg-surface);
  }
}
</style>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useRCMStore, cosoDimensions, getEffectivenessInterpretation, getEffectivenessKey, type RCMItem } from '~/stores/rcm'
import { getControlEffectivenessColorClass } from '~/utils/riskLevelBadge'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import TableEntities from '~/components/shared/TableEntities.vue'
import type { RiskMitigation } from '~/types/risk'
import { useGlobalModalStore } from '~/stores/global-modal'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

definePageMeta({
  middleware: 'auth'
})

const { t } = useI18n()
const { canWrite } = useRbac()
const rcmStore = useRCMStore()
const riskProfileStore = useRiskProfileStore()
const mitigationStore = useMitigationStore()
const toast = useToastNotification()
const { fiscalYears } = useFiscalYear()

const searchQuery = ref('')
const showRatingTable = ref(true)
const isModalOpen = ref(false)
const isEditMode = ref(false)
const selectedRiskId = ref('')
const selectedMitigationId = ref('')
const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)

const steps = computed(() => [
  { key: 'riskProfile', label: t('rcm.modal.steps.riskProfile') },
  { key: 'controlMitigation', label: t('rcm.modal.steps.controlMitigation') },
  { key: 'cosoEvaluation', label: t('rcm.modal.steps.cosoEvaluation') }
])

const cosoModalDimensions = computed(() => [
  {
    key: 'design_effectiveness_rating' as const,
    name: t('rcm.dimensions.designFullName'),
    desc: t('rcm.dimensions.designDesc')
  },
  {
    key: 'operating_effectiveness_rating' as const,
    name: t('rcm.dimensions.operatingFullName'),
    desc: t('rcm.dimensions.operatingDesc')
  },
  {
    key: 'coverage_completeness_rating' as const,
    name: t('rcm.dimensions.coverageFullName'),
    desc: t('rcm.dimensions.coverageDesc')
  },
  {
    key: 'timeliness_rating' as const,
    name: t('rcm.dimensions.timelinessFullName'),
    desc: t('rcm.dimensions.timelinessDesc')
  },
  {
    key: 'automation_monitoring_rating' as const,
    name: t('rcm.dimensions.automationFullName'),
    desc: t('rcm.dimensions.automationDesc')
  }
])

watch(isModalOpen, (isOpen) => {
  if (isOpen) {
    currentStep.value = 0
  }
})

watch(currentStep, () => {
  bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
})

const validateStep = (stepIdx: number): boolean => {
  if (stepIdx === 0) {
    if (!formData.value.risk_code || !formData.value.risk_event) {
      toast.showWarning(t('rcm.toasts.validationWarning'))
      return false
    }
  } else if (stepIdx === 1) {
    if (!formData.value.control_code || !formData.value.control_description) {
      toast.showWarning(t('rcm.toasts.validationWarning'))
      return false
    }
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }
  for (let i = currentStep.value; i < targetIdx; i++) {
    if (!validateStep(i)) return
  }
  currentStep.value = targetIdx
}

const handleNext = () => {
  if (!validateStep(currentStep.value)) return
  currentStep.value++
}

// Every risk in the Corporate Risk Profile; the department follows from the chosen risk.
const riskOptionsForModal = computed(() => {
  const list = riskProfileStore.risks || []
  return list.map(r => ({
    id: String(r.id),
    value: String(r.id),
    label: `${riskProfileStore.getFormattedId(r)} - ${r.name}`,
    code: riskProfileStore.getFormattedId(r),
    name: r.name,
    branch: r.branch || r.category || 'Head Office',
    riskLevel: r.riskLevel
  }))
})

const mitigationOptionsForModal = computed(() => {
  return availableMitigations.value.map(m => ({
    id: m.id,
    value: m.id,
    label: `${m.riskControlId || 'CTL-001'} - ${m.mitigationPlan} (PIC: ${m.pic})`,
    riskControlId: m.riskControlId,
    mitigationPlan: m.mitigationPlan,
    pic: m.pic
  }))
})

const yearOptions = computed(() => {
  return fiscalYears.value.map(y => ({
    label: t('rcm.filters.yearOption', { year: y }),
    value: y
  }))
})

// Computed so the "All departments" label follows the active locale (branch names are master data).
const departmentOptions = computed(() => [
  { label: t('rcm.filters.allDepartments'), value: 'All Departments' },
  { label: 'Head Office', value: 'Head Office' },
  { label: 'Jakarta Branch', value: 'Jakarta Branch' },
  { label: 'Surabaya Branch', value: 'Surabaya Branch' },
  { label: 'Bandung Branch', value: 'Bandung Branch' },
  { label: 'Bali Branch', value: 'Bali Branch' }
])

const formData = ref<Partial<RCMItem>>({
  risk_id: '',
  risk_code: '',
  risk_event: '',
  control_code: '',
  control_description: '',
  control_type: 'Preventive',
  control_owner: 'Finance Manager',
  department: 'Head Office',
  year: 2026,
  design_effectiveness_weight: 20,
  design_effectiveness_rating: 4,
  operating_effectiveness_weight: 20,
  operating_effectiveness_rating: 3,
  coverage_completeness_weight: 20,
  coverage_completeness_rating: 4,
  timeliness_weight: 20,
  timeliness_rating: 3,
  automation_monitoring_weight: 20,
  automation_monitoring_rating: 3,
  notes: ''
})

const dynamicYears = computed(() => {
  const years = new Set<number>()
  
  riskProfileStore.risks.forEach(risk => {
    if (risk.assessments) {
      risk.assessments.forEach((ass: any) => {
        if (ass.year) years.add(ass.year)
      })
    }
  })
  
  if (rcmStore.selectedYear) {
    years.add(rcmStore.selectedYear)
  }
  
  if (years.size === 0) {
    years.add(new Date().getFullYear())
  }
  
  return Array.from(years)
    .sort((a, b) => b - a)
    .map(y => ({ id: y, label: t('rcm.filters.yearOption', { year: y }) }))
})

const dynamicDepartments = computed(() => {
  const depts = new Set<string>()
  
  riskProfileStore.risks.forEach(risk => {
    if (risk.branch) {
      depts.add(risk.branch)
    } else if (risk.category) {
      depts.add(risk.category)
    }
  })
  
  if (rcmStore.selectedDepartment && rcmStore.selectedDepartment !== 'All Departments') {
    depts.add(rcmStore.selectedDepartment)
  }
  
  if (depts.size === 0) {
    depts.add('Head Office')
  }
  
  return [
    { id: 'All Departments', label: t('rcm.filters.allDepartments') },
    ...Array.from(depts).sort().map(d => ({ id: d, label: d }))
  ]
})

const selectedDepartmentLabel = computed(() =>
  rcmStore.selectedDepartment === 'All Departments'
    ? t('rcm.filters.allDepartments')
    : rcmStore.selectedDepartment
)

onMounted(() => {
  rcmStore.fetchRCMList()
  mitigationStore.fetchMitigations()
})

const filteredList = computed(() => {
  let list = rcmStore.filteredRCMList
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(item =>
      item.risk_code.toLowerCase().includes(q) ||
      item.risk_event.toLowerCase().includes(q) ||
      item.control_code.toLowerCase().includes(q) ||
      item.control_description.toLowerCase().includes(q) ||
      item.department.toLowerCase().includes(q) ||
      item.control_owner.toLowerCase().includes(q)
    )
  }
  return list
})

const availableMitigations = computed(() => {
  if (!selectedRiskId.value) return []
  return mitigationStore.getMitigationsByRiskId(selectedRiskId.value)
})

const getDimAverage = (dimKey: string) => {
  const map: Record<string, number> = {
    design_effectiveness: rcmStore.cosoAverages.design,
    operating_effectiveness: rcmStore.cosoAverages.operating,
    coverage_completeness: rcmStore.cosoAverages.coverage,
    timeliness: rcmStore.cosoAverages.timeliness,
    automation_monitoring: rcmStore.cosoAverages.automation
  }
  return map[dimKey] || 0
}

const getItemRating = (scorePercent: number) => {
  return getEffectivenessInterpretation(scorePercent)
}

const getLocalizedRating = (scorePercent: number) =>
  t(`rcm.interpretationTable.ratings.${getEffectivenessKey(scorePercent)}`)

const getLocalizedInterpretation = (scorePercent: number) =>
  t(`rcm.interpretationTable.descriptions.${getEffectivenessKey(scorePercent)}`)

const getRatingBadgeClass = (ratingLabel: string) =>
  `font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block ${getControlEffectivenessColorClass(ratingLabel)}`

const getRatingBtnClass = (current: number, star: number) => {
  if (current === star) {
    return 'bg-primary-600 dark:bg-primary-600 text-white shadow-sm font-bold border border-primary-700 dark:border-primary-500'
  }
  return 'bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 border border-slate-300 dark:border-slate-700'
}

const calculatedModalScorePercent = computed(() => {
  return rcmStore.calculateItemScorePercent(formData.value)
})

const onRiskSelected = (newVal?: any) => {
  const riskId = typeof newVal === 'object' && newVal !== null ? (newVal.value || newVal.id) : (newVal || selectedRiskId.value)
  if (!riskId) return
  selectedRiskId.value = String(riskId)

  const risk = riskProfileStore.getRiskById(riskId)
  if (risk) {
    formData.value.risk_id = String(risk.id)
    formData.value.risk_code = riskProfileStore.getFormattedId(risk)
    formData.value.risk_event = risk.name
    formData.value.department = risk.branch || risk.category || 'Head Office'

    const mits = mitigationStore.getMitigationsByRiskId(String(risk.id))
    const firstMit = mits && mits.length > 0 ? mits[0] : undefined
    if (firstMit) {
      selectedMitigationId.value = firstMit.id
      onControlSelected(firstMit.id)
    } else {
      selectedMitigationId.value = ''
      formData.value.control_code = 'CTL-' + formData.value.risk_code
      formData.value.control_description = t('rcm.defaults.mitigationDescPrefix') + risk.name
      formData.value.control_owner = 'Department Lead'
    }
  }
}

const onControlSelected = (newVal?: any) => {
  const mitId = typeof newVal === 'object' && newVal !== null ? (newVal.value || newVal.id) : (newVal || selectedMitigationId.value)
  if (!mitId) return
  selectedMitigationId.value = String(mitId)

  const mits = availableMitigations.value
  const found = mits.find(m => m.id === mitId)
  if (found) {
    formData.value.control_code = found.riskControlId || ('CTL-' + formData.value.risk_code)
    formData.value.control_description = found.mitigationPlan || found.notes || ''
    formData.value.control_owner = found.pic || 'Department Lead'
  }
}

const openAddModal = () => {
  isEditMode.value = false
  currentStep.value = 0

  const defaultRisks = riskOptionsForModal.value
  const defaultRiskItem = defaultRisks.length > 0 && defaultRisks[0] ? riskProfileStore.getRiskById(defaultRisks[0].value) : (riskProfileStore.risks && riskProfileStore.risks.length > 0 ? riskProfileStore.risks[0] : undefined)
  const defaultCode = defaultRiskItem ? riskProfileStore.getFormattedId(defaultRiskItem) : ''
  const defaultEvent = defaultRiskItem ? defaultRiskItem.name : ''
  const defaultDept = defaultRiskItem ? (defaultRiskItem.branch || defaultRiskItem.category || 'Head Office') : 'Head Office'

  formData.value = {
    risk_id: defaultRiskItem ? String(defaultRiskItem.id) : '',
    risk_code: defaultCode,
    risk_event: defaultEvent,
    control_code: defaultCode ? 'CTL-' + defaultCode : '',
    control_description: '',
    control_type: 'Preventive',
    control_owner: '',
    department: defaultDept,
    year: rcmStore.selectedYear,
    design_effectiveness_weight: 20,
    design_effectiveness_rating: 4,
    operating_effectiveness_weight: 20,
    operating_effectiveness_rating: 3,
    coverage_completeness_weight: 20,
    coverage_completeness_rating: 4,
    timeliness_weight: 20,
    timeliness_rating: 3,
    automation_monitoring_weight: 20,
    automation_monitoring_rating: 3,
    notes: ''
  }

  if (defaultRiskItem) {
    selectedRiskId.value = String(defaultRiskItem.id)
    onRiskSelected(String(defaultRiskItem.id))
  }

  isModalOpen.value = true
}

const openEditModal = (item: RCMItem) => {
  isEditMode.value = true
  currentStep.value = 0
  formData.value = JSON.parse(JSON.stringify(item))
  if (!formData.value.control_type) {
    formData.value.control_type = 'Preventive'
  }
  selectedRiskId.value = item.risk_id || ''
  isModalOpen.value = true
}

const saveForm = async () => {
  if (!formData.value.risk_code || !formData.value.risk_event || !formData.value.control_description) {
    toast.showWarning(t('rcm.toasts.validationWarning'))
    return
  }

  try {
    if (isEditMode.value && formData.value.id) {
      await rcmStore.updateRCMItem(formData.value as RCMItem)
      toast.showSuccess(t('rcm.toasts.updateSuccess'))
    } else {
      await rcmStore.addRCMItem(formData.value as any)
      toast.showSuccess(t('rcm.toasts.addSuccess'))
    }
    isModalOpen.value = false
  } catch (err: any) {
    toast.showError(rcmStore.errorMsg || err?.message || t('errors.unknown'))
  }
}

const confirmDelete = async (id: string) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('rcm.confirmDelete.description') })) {
    try {
      await rcmStore.deleteRCMItem(id)
      toast.showSuccess(t('rcm.toasts.deleteSuccess'))
    } catch (err: any) {
      toast.showError(rcmStore.errorMsg || err?.message || 'Gagal menghapus data dari server.')
    }
  }
}
</script>
