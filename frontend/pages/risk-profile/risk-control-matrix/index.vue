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

        <UButton
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
            <UTooltip :text="t('rcm.table.actions.edit')">
            <UButton
              icon="i-lucide-edit-3"
              color="neutral"
              variant="ghost"
              size="xs"
              :title="t('rcm.table.actions.edit')"
              @click="openEditModal(row.original)"
            />
            </UTooltip>  
            <UTooltip :text="t('rcm.table.actions.delete')">
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
      v-model:open="isModalOpen" 
      :title="t('rcm.modal.manageTitle')"
      :ui="{
        content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <div class="flex flex-col max-h-[85vh] overflow-hidden bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">
          <!-- Header stays fixed; only the form body below scrolls -->
          <div class="shrink-0 flex items-center justify-between gap-4 border-b border-slate-100 dark:border-slate-800 px-6 py-4">
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ isEditMode ? t('rcm.modal.editTitle') : t('rcm.modal.addTitle') }}
            </h3>
            <UButton
              type="button"
              icon="i-lucide-x"
              color="neutral"
              variant="ghost"
              :aria-label="t('common.close')"
              :title="t('common.close')"
              @click="isModalOpen = false"
            />
          </div>

          <div class="flex-1 min-h-0 overflow-y-auto p-6 space-y-4">
          <!-- Risk selection from the Corporate Risk Profile -->
          <div>
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                {{ t('rcm.modal.riskSelectLabel') }}
              </label>
              <USelectMenu
                v-model="selectedRiskId"
                :items="riskOptionsForModal"
                value-key="value"
                size="md"
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
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.riskCodeLabel') }}</label>
              <input
                v-model="formData.risk_code"
                type="text"
                readonly
                class="w-full bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-600 dark:text-slate-400 font-bold cursor-not-allowed"
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.departmentLabel') }}</label>
              <input
                v-model="formData.department"
                type="text"
                readonly
                class="w-full bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-600 dark:text-slate-400 font-medium cursor-not-allowed"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.riskEventLabel') }}</label>
            <textarea
              v-model="formData.risk_event"
              rows="2"
              readonly
              class="w-full bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-600 dark:text-slate-400 font-medium cursor-not-allowed"
            ></textarea>
          </div>

          <!-- Synchronized Control & Mitigation Selection from Risk Mitigation Plans & Controls -->
          <div v-if="availableMitigations.length > 0">
            <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
              {{ t('rcm.modal.controlSelectLabel') }}
            </label>
            <USelectMenu
              v-model="selectedMitigationId"
              :items="mitigationOptionsForModal"
              value-key="value"
              size="md"
              class="w-full"
              :placeholder="t('rcm.modal.controlSelectPlaceholder')"
              @update:model-value="onControlSelected"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.controlCodeLabel') }}</label>
              <input
                v-model="formData.control_code"
                type="text"
                :placeholder="t('rcm.modal.controlCodePlaceholder')"
                class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500 font-bold"
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.controlOwnerLabel') }}</label>
              <input
                v-model="formData.control_owner"
                type="text"
                :placeholder="t('rcm.modal.controlOwnerPlaceholder')"
                class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('rcm.modal.controlDescLabel') }}</label>
            <textarea
              v-model="formData.control_description"
              rows="3"
              :placeholder="t('rcm.modal.controlDescPlaceholder')"
              class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            ></textarea>
          </div>

          <!-- 5 Dimensions COSO 2013 Rating System -->
          <div class="space-y-4 pt-2 border-t border-slate-100 dark:border-slate-800">
            <div class="flex items-center justify-between">
              <h4 class="font-bold text-slate-900 dark:text-white text-md">{{ t('rcm.modal.evaluationTitle') }}</h4>
              <span class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.modal.evaluationWeight') }}</span>
            </div>

            <!-- Dimensi 1: Design -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">{{ t('rcm.dimensions.designFullName') }}</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.dimensions.designDesc') }}</p>
                </div>
                <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-md">
                  {{ (formData.design_effectiveness_rating || 1) * 4 }}%
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  v-for="star in [1, 2, 3, 4, 5]"
                  :key="star"
                  type="button"
                  class="flex-1 py-1.5 rounded-lg text-md font-semibold transition-all"
                  :class="getRatingBtnClass(formData.design_effectiveness_rating || 1, star)"
                  @click="formData.design_effectiveness_rating = star"
                >
                  {{ star }} ({{ star * 4 }}%)
                </button>
              </div>
            </div>

            <!-- Dimensi 2: Operating -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">{{ t('rcm.dimensions.operatingFullName') }}</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.dimensions.operatingDesc') }}</p>
                </div>
                <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-md">
                  {{ (formData.operating_effectiveness_rating || 1) * 4 }}%
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  v-for="star in [1, 2, 3, 4, 5]"
                  :key="star"
                  type="button"
                  class="flex-1 py-1.5 rounded-lg text-md font-semibold transition-all"
                  :class="getRatingBtnClass(formData.operating_effectiveness_rating || 1, star)"
                  @click="formData.operating_effectiveness_rating = star"
                >
                  {{ star }} ({{ star * 4 }}%)
                </button>
              </div>
            </div>

            <!-- Dimensi 3: Coverage -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">{{ t('rcm.dimensions.coverageFullName') }}</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.dimensions.coverageDesc') }}</p>
                </div>
                <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-md">
                  {{ (formData.coverage_completeness_rating || 1) * 4 }}%
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  v-for="star in [1, 2, 3, 4, 5]"
                  :key="star"
                  type="button"
                  class="flex-1 py-1.5 rounded-lg text-md font-semibold transition-all"
                  :class="getRatingBtnClass(formData.coverage_completeness_rating || 1, star)"
                  @click="formData.coverage_completeness_rating = star"
                >
                  {{ star }} ({{ star * 4 }}%)
                </button>
              </div>
            </div>

            <!-- Dimensi 4: Timeliness -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">{{ t('rcm.dimensions.timelinessFullName') }}</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.dimensions.timelinessDesc') }}</p>
                </div>
                <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-md">
                  {{ (formData.timeliness_rating || 1) * 4 }}%
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  v-for="star in [1, 2, 3, 4, 5]"
                  :key="star"
                  type="button"
                  class="flex-1 py-1.5 rounded-lg text-md font-semibold transition-all"
                  :class="getRatingBtnClass(formData.timeliness_rating || 1, star)"
                  @click="formData.timeliness_rating = star"
                >
                  {{ star }} ({{ star * 4 }}%)
                </button>
              </div>
            </div>

            <!-- Dimensi 5: Automation -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">{{ t('rcm.dimensions.automationFullName') }}</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">{{ t('rcm.dimensions.automationDesc') }}</p>
                </div>
                <span class="font-extrabold text-primary-700 dark:text-primary-300 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 text-md">
                  {{ (formData.automation_monitoring_rating || 1) * 4 }}%
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  v-for="star in [1, 2, 3, 4, 5]"
                  :key="star"
                  type="button"
                  class="flex-1 py-1.5 rounded-lg text-md font-semibold transition-all"
                  :class="getRatingBtnClass(formData.automation_monitoring_rating || 1, star)"
                  @click="formData.automation_monitoring_rating = star"
                >
                  {{ star }} ({{ star * 4 }}%)
                </button>
              </div>
            </div>
          </div>

          <!-- Total Score Summary in Modal -->
          <div class="bg-primary-600 dark:bg-primary-700 text-white p-4 rounded-xl shadow-md flex items-center justify-between">
            <div>
              <span class="text-md font-bold text-white">{{ t('rcm.modal.totalWeightedScore') }}</span>
              <p class="text-sm text-primary-100 mt-0.5">
                {{ t('rcm.modal.interpretation') }} <strong class="text-white font-extrabold">{{ getLocalizedRating(calculatedModalScorePercent) }}</strong>
              </p>
            </div>
            <span class="text-2xl font-extrabold text-white">{{ calculatedModalScorePercent }}%</span>
          </div>

          <!-- Actions -->
          <div class="flex justify-end gap-2 pt-2">
            <UButton variant="ghost" color="neutral" @click="() => { isModalOpen = false }">{{ t('rcm.modal.cancel') }}</UButton>
            <UButton color="primary" class="font-medium" @click="saveForm">
              {{ t('rcm.modal.save') }}
            </UButton>
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
import { ref, computed, onMounted } from 'vue'
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
