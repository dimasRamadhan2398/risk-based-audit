<template>
  <div class="p-4 md:p-6 space-y-6 min-h-screen">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400 mb-1">
          <NuxtLink to="/risk-profile" class="hover:text-primary-600 dark:hover:text-primary-400 transition-colors">Risk Profile</NuxtLink>
          <span>/</span>
          <span class="text-slate-800 dark:text-slate-200 font-medium">Risk Control Matrix</span>
        </div>
        <h1 class="text-2xl md:text-3xl font-bold text-slate-900 dark:text-white tracking-tight">
          Risk Control Matrix (RCM)
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Pengukuran efektivitas internal control berbasis 5 dimensi COSO 2013 & evaluasi risiko terintegrasi.
        </p>
      </div>

      <!-- Actions / Filters -->
      <div class="flex flex-wrap items-center justify-end gap-3">
        <USelect
          v-model="rcmStore.selectedYear"
          :items="yearOptions"
          value-key="value"
          size="md"
          class="w-36"
        />

        <USelect
          v-model="rcmStore.selectedDepartment"
          :items="departmentOptions"
          value-key="value"
          size="md"
          class="w-64"
        />

        <UButton
          color="primary"
          variant="solid"
          class="font-medium shadow-sm"
          @click="openAddModal"
        >
          <UIcon name="i-lucide-plus" class="size-4 mr-1.5" />
          Tambah Control Matrix
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
              <span class="text-sm font-semibold tracking-wider text-slate-400 dark:text-slate-400">Pengukuran Tutup Buku Akhir Tahun</span>
            </div>
            <span class="text-sm font-medium text-slate-500 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-md border border-transparent dark:border-slate-700">
              Tahun: {{ rcmStore.selectedYear }} | {{ rcmStore.selectedDepartment }}
            </span>
          </div>
          <h2 class="text-xl font-bold text-slate-900 dark:text-white mt-1">Internal Control Effectiveness</h2>
        </div>

        <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-50 dark:bg-slate-800/60 rounded-xl p-4 border border-slate-100 dark:border-slate-700/60">
          <!-- Big Score % -->
          <div class="flex flex-col justify-center border-b md:border-b-0 md:border-r border-slate-200 dark:border-slate-700 pb-3 md:pb-0 md:pr-4">
            <span class="text-sm text-slate-500 dark:text-slate-400 font-medium">Real Effectiveness Score</span>
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
              <span class="text-slate-600 dark:text-slate-300 font-medium">Inherent Risk (Risiko Prioritas Awal Tahun):</span>
              <span class="font-bold text-slate-900 dark:text-slate-100 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                {{ rcmStore.totalInherentRisk }} Risiko
              </span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-slate-600 dark:text-slate-300 font-medium">Residual Risk (Sisa Risiko Tutup Buku):</span>
              <span class="font-bold text-red-600 dark:text-red-400 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
                {{ rcmStore.totalResidualRisk }} Risiko
              </span>
            </div>
            <UAlert
              :color="(rcmStore.effectivenessRating.alertColor as any) || 'info'"
              variant="subtle"
              icon="i-lucide-info"
              :title="'Interpretasi Hasil COSO (' + rcmStore.effectivenessRating.rating + '):'"
              :description="rcmStore.effectivenessRating.interpretation"
              class="rounded-xl shadow-xs border-none! bg-transparent! border-transparent!"
            />
          </div>
        </div>

        <div class="mt-3 text-sm text-slate-400 dark:text-slate-400 flex items-center gap-4">
          <UIcon name="i-lucide-check-circle-2" class="size-10 text-emerald-500" />
          <span>Data Inherent & Residual Risk terintegrasi langsung secara otomatis dari Corporate Risk Profile.</span>
        </div>
      </div>

      <!-- Card 2: COSO 2013 5 Dimensions Summary (1 Col) -->
      <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm p-6 flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-bold text-slate-900 dark:text-white">Rata-Rata COSO 2013</h3>
            <span class="text-md font-bold text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/60 px-2.5 py-1 rounded-full border border-primary-100 dark:border-primary-800">
              {{ rcmStore.cosoAverages.totalWeighted }}%
            </span>
          </div>

          <div class="mt-4 space-y-3">
            <div v-for="dim in cosoDimensions" :key="dim.key" class="space-y-1">
              <div class="flex justify-between text-md">
                <span class="font-medium text-slate-700 dark:text-slate-300">{{ dim.shortLabel }}</span>
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
          <span>Total Kontrol Dievaluasi:</span>
          <span class="font-bold text-slate-900 dark:text-white">{{ rcmStore.filteredRCMList.length }} Item</span>
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
          <span class="font-bold text-slate-900 dark:text-white text-sm md:text-base">Tabel Standar Interpretasi Rating Efektivitas Kontrol Internal</span>
        </div>
        <UIcon :name="showRatingTable ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5 text-slate-400 dark:text-slate-500" />
      </button>

      <div v-if="showRatingTable" class="px-6 pb-6 border-t border-slate-100 dark:border-slate-800 pt-4">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-md border-collapse">
            <thead>
              <tr class="bg-slate-50 dark:bg-slate-800/80 border-b border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 font-semibold">
                <th class="py-2.5 px-4 rounded-l-lg whitespace-nowrap">Total Weighted Score (%)</th>
                <th class="py-2.5 px-4 whitespace-nowrap">Rating</th>
                <th class="py-2.5 px-4 rounded-r-lg">Interpretation</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-emerald-600 dark:text-emerald-400 whitespace-nowrap">90 – 100%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-emerald-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Highly Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">Controls reliably mitigate risk and require only routine monitoring.</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-sky-600 dark:text-sky-400 whitespace-nowrap">80 – 89%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-sky-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">Controls function well; only minor improvements are recommended.</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-amber-600 dark:text-amber-400 whitespace-nowrap">70 – 79%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-amber-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Moderately Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">Some weaknesses exist; corrective actions should be planned.</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-orange-600 dark:text-orange-400 whitespace-nowrap">60 – 69%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-orange-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Weak</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">Significant improvements are needed to reduce risk adequately.</td>
              </tr>
              <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2.5 px-4 font-bold text-red-600 dark:text-red-400 whitespace-nowrap">&lt; 60%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-red-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Ineffective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">Controls do not provide sufficient risk mitigation and require immediate attention.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Main Risk Control Matrix Table -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden p-6 space-y-4">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h3 class="text-lg font-bold text-slate-900 dark:text-white">Daftar Risk Control Matrix</h3>
          <p class="text-sm text-slate-500 dark:text-slate-400">Hasil evaluasi 5 Dimensi COSO 2013 (Rating 1 - 5 mewakili 4% - 20% per dimensi)</p>
        </div>
        <div class="relative w-full md:w-64">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari risiko / kontrol..."
            class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg pl-9 pr-3 py-1.5 text-sm text-slate-700 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          <UIcon name="i-lucide-search" class="absolute left-3 top-2.5 size-3.5 text-slate-400 dark:text-slate-500" />
        </div>
      </div>

      <TableEntities
        :data="filteredList"
        :columns="rcmStore.columns"
        :items-per-page="10"
        :empty-state="{
          icon: 'i-lucide-shield-alert',
          label: 'Tidak ada data Risk Control Matrix',
          description: 'Tidak ada data Risk Control Matrix yang sesuai dengan filter.'
        }"
      >
        <!-- Risk Code & Event -->
        <template #risk_code-cell="{ row }">
          <div class="max-w-[280px] min-w-[220px] whitespace-normal break-words space-y-1">
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
          <div class="max-w-[280px] min-w-[220px] whitespace-normal break-words space-y-1">
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
            <p class="text-xs text-slate-500 dark:text-slate-400">PIC: <strong class="text-slate-700 dark:text-slate-200">{{ row.original.control_owner }}</strong></p>
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
              {{ getItemRating(row.original.total_weighted_score).rating }}
            </span>
          </div>
        </template>

        <!-- Actions -->
        <template #actions-cell="{ row }">
          <div class="flex items-center justify-end gap-1">
            <UTooltip text="Edit Control">
            <UButton
              icon="i-lucide-edit-3"
              color="neutral"
              variant="ghost"
              size="xs"
              title="Edit Control"
              @click="openEditModal(row.original)"
            />
            </UTooltip>  
            <UTooltip text="Hapus Control">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="xs"
              title="Hapus Control"
              @click="confirmDelete(row.original.id)"
            />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
    </div>

    <!-- Add / Edit Modal -->
    <UModal 
      v-model:open="isModalOpen" 
      title="Manage Risk Control Matrix"
      :ui="{
        content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <div class="p-6 space-y-4 max-h-[85vh] overflow-y-auto bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">
          <h3 class="text-lg font-bold text-slate-900 dark:text-white border-b border-slate-100 dark:border-slate-800 pb-2">
            {{ isEditMode ? 'Edit Risk Control Matrix' : 'Tambah Risk Control Matrix Baru' }}
          </h3>

          <!-- Synchronized Branch and Risk Selection from Corporate Risk Profile -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                Branch / Departemen
              </label>
              <USelectMenu
                v-model="selectedBranchInModal"
                :items="branchModalOptions"
                value-key="value"
                size="md"
                class="w-full"
                placeholder="Pilih Branch..."
                @update:model-value="onBranchModalChange"
              />
            </div>

            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1.5">
                Pilih Nama Risiko (Corporate Risk Profile)
              </label>
              <USelectMenu
                v-model="selectedRiskId"
                :items="riskOptionsForModal"
                value-key="value"
                size="md"
                class="w-full"
                placeholder="Pilih Nama Risiko..."
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
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">Kode Risiko (Sinkron)</label>
              <input
                v-model="formData.risk_code"
                type="text"
                readonly
                class="w-full bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-600 dark:text-slate-400 font-bold cursor-not-allowed"
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">Departemen / Branch (Sinkron)</label>
              <input
                v-model="formData.department"
                type="text"
                readonly
                class="w-full bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-600 dark:text-slate-400 font-medium cursor-not-allowed"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">Kejadian Risiko / Risk Event (Sinkron)</label>
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
              Pilih Risk Control ID (Rencana Mitigasi & Kontrol)
            </label>
            <USelectMenu
              v-model="selectedMitigationId"
              :items="mitigationOptionsForModal"
              value-key="value"
              size="md"
              class="w-full"
              placeholder="Pilih Risk Control ID..."
              @update:model-value="onControlSelected"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">Kode Kontrol (Risk Control ID)</label>
              <input
                v-model="formData.control_code"
                type="text"
                placeholder="misal: CTL-FIN-001"
                class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500 font-bold"
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">PIC / Owner Kontrol (Sinkron)</label>
              <input
                v-model="formData.control_owner"
                type="text"
                placeholder="Finance Manager"
                class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 dark:text-slate-300 mb-1">Deskripsi Aktivitas Pengendalian Internal</label>
            <textarea
              v-model="formData.control_description"
              rows="3"
              placeholder="Jelaskan mekanisme kontrol operasional, review, verifikasi, atau sistemik yang dijalankan..."
              class="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-md text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            ></textarea>
          </div>

          <!-- 5 Dimensions COSO 2013 Rating System -->
          <div class="space-y-4 pt-2 border-t border-slate-100 dark:border-slate-800">
            <div class="flex items-center justify-between">
              <h4 class="font-bold text-slate-900 dark:text-white text-md">Evaluasi 5 Dimensi Internal Control (COSO 2013)</h4>
              <span class="text-md text-slate-500 dark:text-slate-400">Bobot Tetap: 20% Tiap Dimensi (Skala 1 - 5)</span>
            </div>

            <!-- Dimensi 1: Design -->
            <div class="bg-slate-50 dark:bg-slate-800/60 p-3.5 rounded-xl border border-slate-200/80 dark:border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">1. Design Effectiveness (20%)</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">Kesesuaian rancangan kontrol terhadap risiko & SOP</p>
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
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">2. Operating Effectiveness (20%)</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">Konsistensi eksekusi dan ketiadaan deviasi kontrol</p>
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
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">3. Coverage & Completeness (20%)</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">Cakupan kontrol pada seluruh transaksi/aktivitas</p>
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
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">4. Timeliness & Frequency (20%)</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">Ketepatan waktu kontrol mendeteksi / mencegah insiden</p>
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
                  <span class="font-bold text-slate-800 dark:text-slate-200 text-md">5. Automation & Monitoring (20%)</span>
                  <p class="text-md text-slate-500 dark:text-slate-400">Tingkat otomatisasi sistemik & continuous monitoring</p>
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
              <span class="text-md font-bold text-white">Total Weighted Effectiveness Score:</span>
              <p class="text-sm text-primary-100 mt-0.5">
                Interpretasi: <strong class="text-white font-extrabold">{{ getItemRating(calculatedModalScorePercent).rating }}</strong>
              </p>
            </div>
            <span class="text-2xl font-extrabold text-white">{{ calculatedModalScorePercent }}%</span>
          </div>

          <!-- Actions -->
          <div class="flex justify-end gap-2 pt-2">
            <UButton variant="ghost" color="neutral" @click="() => { isModalOpen = false }">Batal</UButton>
            <UButton color="primary" class="font-medium" @click="saveForm">
              Simpan Kontrol Matrix
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRCMStore, cosoDimensions, getEffectivenessInterpretation, type RCMItem } from '~/stores/rcm'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useMitigationStore } from '~/stores/mitigation-risk'
import TableEntities from '~/components/shared/TableEntities.vue'
import type { RiskMitigation } from '~/types/risk'
import { useGlobalModalStore } from '~/stores/global-modal'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

definePageMeta({
  middleware: 'auth'
})

const rcmStore = useRCMStore()
const riskProfileStore = useRiskProfileStore()
const mitigationStore = useMitigationStore()
const toast = useToastNotification()

const searchQuery = ref('')
const showRatingTable = ref(true)
const isModalOpen = ref(false)
const isEditMode = ref(false)
const selectedRiskId = ref('')
const selectedMitigationId = ref('')
const selectedBranchInModal = ref('All Branches')


const branchModalOptions = [
  { label: 'Semua Branch / Departemen', value: 'All Branches' },
  { label: 'Head Office', value: 'Head Office' },
  { label: 'Jakarta Branch', value: 'Jakarta Branch' },
  { label: 'Surabaya Branch', value: 'Surabaya Branch' },
  { label: 'Bandung Branch', value: 'Bandung Branch' },
  { label: 'Bali Branch', value: 'Bali Branch' }
]

const riskOptionsForModal = computed(() => {
  let list = riskProfileStore.risks || []
  if (selectedBranchInModal.value && selectedBranchInModal.value !== 'All Branches') {
    list = list.filter(r => (r.branch || r.category || 'Head Office') === selectedBranchInModal.value)
  }
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

const yearOptions = [
  { label: 'Tahun 2026', value: 2026 },
  { label: 'Tahun 2025', value: 2025 },
  { label: 'Tahun 2024', value: 2024 }
]

const departmentOptions = [
  { label: 'Semua Departemen / Branch', value: 'All Departments' },
  { label: 'Head Office', value: 'Head Office' },
  { label: 'Jakarta Branch', value: 'Jakarta Branch' },
  { label: 'Surabaya Branch', value: 'Surabaya Branch' },
  { label: 'Bandung Branch', value: 'Bandung Branch' },
  { label: 'Bali Branch', value: 'Bali Branch' }
]

const formData = ref<Partial<RCMItem>>({
  risk_id: '',
  risk_code: '',
  risk_event: '',
  control_code: '',
  control_description: '',
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
    .map(y => ({ id: y, label: `Tahun ${y}` }))
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
    { id: 'All Departments', label: 'Semua Departemen / Branch' },
    ...Array.from(depts).sort().map(d => ({ id: d, label: d }))
  ]
})

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

const getRatingBadgeClass = (ratingLabel: string) => {
  switch (ratingLabel) {
    case 'Highly Effective':
      return 'bg-emerald-500 text-white font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block'
    case 'Effective':
      return 'bg-sky-500 text-white font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block'
    case 'Moderately Effective':
      return 'bg-amber-500 text-white font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block'
    case 'Weak':
      return 'bg-orange-500 text-white font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block'
    default:
      return 'bg-red-500 text-white font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block'
  }
}

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
      formData.value.control_description = 'Aktivitas mitigasi & pengawasan internal untuk ' + risk.name
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

const onBranchModalChange = (newBranch?: any) => {
  const branch = typeof newBranch === 'object' && newBranch !== null ? (newBranch.value || newBranch.label) : (newBranch || selectedBranchInModal.value)
  selectedBranchInModal.value = branch || 'All Branches'
  
  const available = riskOptionsForModal.value
  if (available.length > 0) {
    const stillValid = available.some(r => r.value === selectedRiskId.value)
    if (!stillValid && available[0]) {
      selectedRiskId.value = available[0].value
      onRiskSelected(available[0].value)
    }
  } else {
    selectedRiskId.value = ''
    formData.value.risk_id = ''
    formData.value.risk_code = ''
    formData.value.risk_event = ''
    formData.value.department = selectedBranchInModal.value !== 'All Branches' ? selectedBranchInModal.value : 'Head Office'
  }
}

const openAddModal = () => {
  isEditMode.value = false
  
  if (rcmStore.selectedDepartment && rcmStore.selectedDepartment !== 'All Departments') {
    selectedBranchInModal.value = rcmStore.selectedDepartment
  } else {
    selectedBranchInModal.value = 'All Branches'
  }

  const defaultRisks = riskOptionsForModal.value
  const defaultRiskItem = defaultRisks.length > 0 && defaultRisks[0] ? riskProfileStore.getRiskById(defaultRisks[0].value) : (riskProfileStore.risks && riskProfileStore.risks.length > 0 ? riskProfileStore.risks[0] : undefined)
  const defaultCode = defaultRiskItem ? riskProfileStore.getFormattedId(defaultRiskItem) : 'FIN-001'
  const defaultEvent = defaultRiskItem ? defaultRiskItem.name : 'Target pendapatan dan laba tidak tercapai'
  const defaultDept = defaultRiskItem ? (defaultRiskItem.branch || defaultRiskItem.category || 'Head Office') : 'Head Office'

  formData.value = {
    risk_id: defaultRiskItem ? String(defaultRiskItem.id) : '1',
    risk_code: defaultCode,
    risk_event: defaultEvent,
    control_code: 'CTL-' + defaultCode,
    control_description: 'Review bulanan pencapaian KPI sales dan monitoring piutang usaha secara ketat.',
    control_owner: 'Finance Manager',
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
  selectedBranchInModal.value = item.department || 'All Branches'
  selectedRiskId.value = item.risk_id || ''
  isModalOpen.value = true
}

const saveForm = async () => {
  if (!formData.value.risk_code || !formData.value.risk_event || !formData.value.control_description) {
    toast.showWarning('Mohon pilih risiko dan lengkapi deskripsi kontrol.')
    return
  }

  if (isEditMode.value && formData.value.id) {
    await rcmStore.updateRCMItem(formData.value as RCMItem)
    toast.showSuccess('Risk Control Matrix berhasil diupdate')
  } else {
    await rcmStore.addRCMItem(formData.value as any)
    toast.showSuccess('Risk Control Matrix berhasil ditambahkan')
  }
  isModalOpen.value = false
}

const confirmDelete = async (id: string) => {
  if (await useGlobalModalStore().confirmDelete({ description: 'Apakah Anda yakin ingin menghapus baris Risk Control Matrix ini?' })) {
    await rcmStore.deleteRCMItem(id)
    toast.showSuccess('Risk Control Matrix berhasil dihapus')
  }
}
</script>
