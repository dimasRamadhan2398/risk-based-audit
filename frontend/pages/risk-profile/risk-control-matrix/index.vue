<template>
  <div class="p-4 md:p-6 space-y-6 min-h-screen">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <div class="flex items-center gap-2 text-sm text-slate-500 mb-1">
          <NuxtLink to="/risk-profile" class="hover:text-primary-600 transition-colors">Risk Profile</NuxtLink>
          <span>/</span>
          <span class="text-slate-800 font-medium">Risk Control Matrix</span>
        </div>
        <h1 class="text-2xl md:text-3xl font-bold text-slate-900 tracking-tight">
          Risk Control Matrix (RCM)
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Pengukuran efektivitas internal control berbasis 5 dimensi COSO 2013 & evaluasi risiko terintegrasi.
        </p>
      </div>

      <!-- Actions / Filters -->
      <div class="flex flex-wrap items-center gap-3">
        <USelectMenu
          v-model="rcmStore.selectedYear"
          :items="dynamicYears"
          value-key="id"
          label-key="label"
          class="w-36"
        />

        <USelectMenu
          v-model="rcmStore.selectedDepartment"
          :items="dynamicDepartments"
          value-key="id"
          label-key="label"
          class="w-60"
        />

        <UButton
          color="primary"
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
      <div class="lg:col-span-2 bg-white rounded-2xl border border-slate-200 shadow-sm p-6 flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-md font-semibold uppercase tracking-wider text-slate-400">Pengukuran Tutup Buku Akhir Tahun</span>
              
            </div>
            <span class="text-md font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded-md">
              Tahun: {{ rcmStore.selectedYear }} | Dep: {{ rcmStore.selectedDepartment }}
            </span>
          </div>
          <h2 class="text-xl font-bold text-slate-900 mt-1">Internal Control Effectiveness</h2>
        </div>

        <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-50 rounded-xl p-4 border border-slate-100">
          <!-- Big Score % -->
          <div class="flex flex-col justify-center border-b md:border-b-0 md:border-r border-slate-200 pb-3 md:pb-0 md:pr-4">
            <span class="text-md text-slate-500 font-medium">Real Effectiveness Score</span>
            <div class="flex items-baseline gap-2 mt-1">
              <span class="text-4xl font-extrabold tracking-tight text-slate-900">
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
            <div class="flex items-center justify-between text-md">
              <span class="text-slate-600 font-medium">Inherent Risk (Risiko Prioritas Awal Tahun):</span>
              <span class="font-bold text-slate-900 bg-white px-2 py-0.5 rounded border border-slate-200">
                {{ rcmStore.totalInherentRisk }} Risiko
              </span>
            </div>
            <div class="flex items-center justify-between text-md">
              <span class="text-slate-600 font-medium">Residual Risk (Sisa Risiko Tutup Buku):</span>
              <span class="font-bold text-red-600 bg-white px-2 py-0.5 rounded border border-slate-200">
                {{ rcmStore.totalResidualRisk }} Risiko
              </span>
            </div>
            <div class="p-2.5 rounded-lg text-md" :class="rcmStore.effectivenessRating.bgClass">
              <div class="font-bold flex items-center gap-1.5">
                <UIcon name="i-lucide-info" class="size-3.5" />
                Interpretasi Hasil COSO:
              </div>
              <p class="mt-0.5 leading-relaxed">{{ rcmStore.effectivenessRating.interpretation }}</p>
            </div>
          </div>
        </div>

        <div class="mt-3 text-md text-slate-400 flex items-center gap-1">
          <UIcon name="i-lucide-check-circle-2" class="size-3.5 text-emerald-500" />
          <span>Data Inherent & Residual Risk terintegrasi langsung secara otomatis dari Corporate Risk Profile.</span>
        </div>
      </div>

      <!-- Card 2: COSO 2013 5 Dimensions Summary (1 Col) -->
      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6 flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-bold text-slate-900">Rata-Rata COSO 2013</h3>
            <span class="text-md font-bold text-primary-700 bg-primary-50 px-2.5 py-1 rounded-full border border-primary-100">
              {{ rcmStore.cosoAverages.totalWeighted }}%
            </span>
          </div>
          <p class="text-md text-slate-500 mt-0.5">Rata-Rata Bobot 5 Dimensi Kontrol</p>

          <div class="mt-4 space-y-3">
            <div v-for="dim in cosoDimensions" :key="dim.key" class="space-y-1">
              <div class="flex justify-between text-md">
                <span class="font-medium text-slate-700">{{ dim.shortLabel }}</span>
                <span class="font-bold text-slate-900">{{ getDimAverage(dim.key) }}%</span>
              </div>
              <div class="w-full bg-slate-100 h-2 rounded-full overflow-hidden">
                <div
                  class="bg-primary-600 h-full rounded-full transition-all duration-300"
                  :style="{ width: `${getDimAverage(dim.key)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-md text-slate-500">
          <span>Total Kontrol Dievaluasi:</span>
          <span class="font-bold text-slate-900">{{ rcmStore.filteredRCMList.length }} Item</span>
        </div>
      </div>
    </div>

    <!-- Collapsible Standard Interpretation Table Reference -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <button
        class="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-slate-50 transition-colors"
        @click="showRatingTable = !showRatingTable"
      >
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-book-open" class="size-5 text-primary-600" />
          <span class="font-bold text-slate-900 text-sm md:text-base">Tabel Standar Interpretasi Rating Efektivitas Kontrol Internal</span>
        </div>
        <UIcon :name="showRatingTable ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" class="size-5 text-slate-400" />
      </button>

      <div v-if="showRatingTable" class="px-6 pb-6 border-t border-slate-100 pt-4">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-md border-collapse">
            <thead>
              <tr class="bg-slate-50 border-b border-slate-200 text-slate-700 font-semibold">
                <th class="py-2.5 px-4 rounded-l-lg whitespace-nowrap">Total Weighted Score (%)</th>
                <th class="py-2.5 px-4 whitespace-nowrap">Rating</th>
                <th class="py-2.5 px-4 rounded-r-lg">Interpretation</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr class="hover:bg-slate-50/50">
                <td class="py-2.5 px-4 font-bold text-emerald-600 whitespace-nowrap">90 – 100%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-emerald-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Highly Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600">Controls reliably mitigate risk and require only routine monitoring.</td>
              </tr>
              <tr class="hover:bg-slate-50/50">
                <td class="py-2.5 px-4 font-bold text-sky-600 whitespace-nowrap">80 – 89%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-sky-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600">Controls function well; only minor improvements are recommended.</td>
              </tr>
              <tr class="hover:bg-slate-50/50">
                <td class="py-2.5 px-4 font-bold text-amber-600 whitespace-nowrap">70 – 79%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-amber-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Moderately Effective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600">Some weaknesses exist; corrective actions should be planned.</td>
              </tr>
              <tr class="hover:bg-slate-50/50">
                <td class="py-2.5 px-4 font-bold text-orange-600 whitespace-nowrap">60 – 69%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-orange-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Weak</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600">Significant improvements are needed to reduce risk adequately.</td>
              </tr>
              <tr class="hover:bg-slate-50/50">
                <td class="py-2.5 px-4 font-bold text-red-600 whitespace-nowrap">&lt; 60%</td>
                <td class="py-2.5 px-4 whitespace-nowrap">
                  <span class="bg-red-500 text-white font-bold px-2.5 py-1 rounded-md text-md whitespace-nowrap inline-block">Ineffective</span>
                </td>
                <td class="py-2.5 px-4 text-slate-600">Controls do not provide sufficient risk mitigation and require immediate attention.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

        <!-- Main Risk Control Matrix Table -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden p-6 space-y-4">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h3 class="text-lg font-bold text-slate-900">Daftar Risk Control Matrix</h3>
          <p class="text-sm text-slate-500">Hasil evaluasi 5 Dimensi COSO 2013 (Rating 1 - 5 mewakili 4% - 20% per dimensi)</p>
        </div>
        <div class="relative w-full md:w-64">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari risiko / kontrol..."
            class="w-full bg-slate-50 border border-slate-200 rounded-lg pl-9 pr-3 py-1.5 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          <UIcon name="i-lucide-search" class="absolute left-3 top-2.5 size-3.5 text-slate-400" />
        </div>
      </div>

      <TableEntities
        :data="filteredList"
        :columns="columns"
        :items-per-page="10"
        :empty-state="{
          icon: 'i-lucide-shield-alert',
          label: 'Tidak ada data Risk Control Matrix',
          description: 'Tidak ada data Risk Control Matrix yang sesuai dengan filter.'
        }"
      >
        <!-- Risk Code & Event -->
        <template #risk_code-cell="{ row }">
          <div class="space-y-1">
            <span class="inline-block px-2 py-0.5 bg-primary-50 text-primary-700 font-bold rounded text-xs">
              {{ row.original.risk_code }}
            </span>
            <p class="font-medium text-slate-900 text-sm leading-snug line-clamp-2 break-words" :title="row.original.risk_event">
              {{ row.original.risk_event }}
            </p>
          </div>
        </template>

        <!-- Control Code & Description -->
        <template #control_code-cell="{ row }">
          <div class="space-y-1">
            <span class="inline-block px-2 py-0.5 bg-slate-100 text-slate-800 font-extrabold rounded text-xs border border-slate-200">
              {{ row.original.control_code }}
            </span>
            <p class="text-slate-600 text-sm leading-relaxed line-clamp-2 break-words" :title="row.original.control_description">
              {{ row.original.control_description }}
            </p>
          </div>
        </template>

        <!-- Department & PIC -->
        <template #department-cell="{ row }">
          <div>
            <p class="text-sm text-slate-800 font-semibold mb-0.5">{{ row.original.department }}</p>
            <p class="text-xs text-slate-500">PIC: <strong class="text-slate-700">{{ row.original.control_owner }}</strong></p>
          </div>
        </template>

        <!-- Design Rating (1-5) -->
        <template #design_effectiveness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700">
            <span class="px-2 py-1 rounded bg-slate-100 text-xs inline-block" :title="`${row.original.design_effectiveness_rating * 4}%`">
              {{ row.original.design_effectiveness_rating }}
            </span>
          </div>
        </template>

        <!-- Operating Rating (1-5) -->
        <template #operating_effectiveness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700">
            <span class="px-2 py-1 rounded bg-slate-100 text-xs inline-block" :title="`${row.original.operating_effectiveness_rating * 4}%`">
              {{ row.original.operating_effectiveness_rating }}
            </span>
          </div>
        </template>

        <!-- Coverage Rating (1-5) -->
        <template #coverage_completeness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700">
            <span class="px-2 py-1 rounded bg-slate-100 text-xs inline-block" :title="`${row.original.coverage_completeness_rating * 4}%`">
              {{ row.original.coverage_completeness_rating }}
            </span>
          </div>
        </template>

        <!-- Timeliness Rating (1-5) -->
        <template #timeliness_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700">
            <span class="px-2 py-1 rounded bg-slate-100 text-xs inline-block" :title="`${row.original.timeliness_rating * 4}%`">
              {{ row.original.timeliness_rating }}
            </span>
          </div>
        </template>

        <!-- Automation Rating (1-5) -->
        <template #automation_monitoring_rating-cell="{ row }">
          <div class="text-center font-bold text-slate-700">
            <span class="px-2 py-1 rounded bg-slate-100 text-xs inline-block" :title="`${row.original.automation_monitoring_rating * 4}%`">
              {{ row.original.automation_monitoring_rating }}
            </span>
          </div>
        </template>

        <!-- Total Score (%) -->
        <template #total_weighted_score-cell="{ row }">
          <div class="text-center">
            <span class="text-xs font-extrabold text-primary-700 bg-primary-50 px-2.5 py-1 rounded-lg border border-primary-100 inline-block">
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
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              title="Edit Control"
              @click="openEditModal(row.original)"
            />
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              title="Hapus Control"
              @click="confirmDelete(row.original.id)"
            />
          </div>
        </template>
      </TableEntities>
    </div>

    <!-- Add / Edit Modal -->
    <UModal v-model:open="isModalOpen" title="Manage Risk Control Matrix">
      <template #body>
        <div class="p-6 space-y-4">
          <!-- Synchronized Risk Dropdown from Corporate Risk Profile -->
          <div>
            <label class="block text-md font-semibold text-slate-700 mb-1">Pilih Risiko (Corporate Risk Profile)</label>
            <USelectMenu
              v-model="selectedRiskId"
              :items="riskProfileStore.risks.map(r => ({ id: r.id, label: `${riskProfileStore.getFormattedId(r)} - ${r.name} (${r.branch || r.category})` }))"
              value-key="id"
              label-key="label"
              placeholder="-- Pilih Risiko dari Corporate Risk Profile --"
              class="w-full"
              size="lg"
              searchable
              @update:model-value="onRiskSelected"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 mb-1">Kode Risiko (Sinkron)</label>
              <input
                v-model="formData.risk_code"
                type="text"
                readonly
                class="w-full bg-slate-100 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-600 font-bold cursor-not-allowed"
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 mb-1">Departemen / Branch (Sinkron)</label>
              <input
                v-model="formData.department"
                type="text"
                readonly
                class="w-full bg-slate-100 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-600 font-medium cursor-not-allowed"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 mb-1">Kejadian Risiko / Risk Event (Sinkron)</label>
            <textarea
              v-model="formData.risk_event"
              rows="2"
              readonly
              class="w-full bg-slate-100 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-600 font-medium cursor-not-allowed"
            ></textarea>
          </div>

          <!-- Synchronized Control & Mitigation Selection from Risk Mitigation Plans & Controls -->
          <div v-if="availableMitigations.length > 0">
            <label class="block text-md font-semibold text-slate-700 mb-1">Pilih Risk Control ID (Rencana Mitigasi & Kontrol)</label>
            <USelectMenu
              v-model="selectedMitigationId"
              :items="availableMitigations.map(m => ({ id: m.id, label: `${m.riskControlId || 'CTL-001'} - ${m.mitigationPlan} (PIC: ${m.pic})` }))"
              value-key="id"
              label-key="label"
              placeholder="-- Pilih Risk Control ID --"
              class="w-full"
              size="lg"
              searchable
              @update:model-value="onControlSelected"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-md font-semibold text-slate-700 mb-1">Kode Kontrol (Risk Control ID)</label>
              <input
                v-model="formData.control_code"
                type="text"
                class="w-full bg-slate-100 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-600 font-medium cursor-not-allowed"
                readonly
              />
            </div>
            <div>
              <label class="block text-md font-semibold text-slate-700 mb-1">PIC / Owner Kontrol (Sinkron)</label>
              <input
                v-model="formData.control_owner"
                type="text"
                placeholder="Finance Manager"
                class="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-md font-semibold text-slate-700 mb-1">Deskripsi Aktivitas Pengendalian Internal</label>
            <textarea
              v-model="formData.control_description"
              rows="3"
              placeholder="Jelaskan mekanisme kontrol operasional, review, verifikasi, atau sistemik yang dijalankan..."
              class="w-full bg-slate-50 border border-slate-200 rounded-lg px-3 py-2 text-md text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
            ></textarea>
          </div>

          <!-- 5 Dimensions COSO 2013 Rating System -->
          <div class="space-y-4 pt-2 border-t border-slate-100">
            <div class="flex items-center justify-between">
              <h4 class="font-bold text-slate-900 text-md">Evaluasi 5 Dimensi Internal Control (COSO 2013)</h4>
              <span class="text-md text-slate-500">Bobot Tetap: 20% Tiap Dimensi (Skala 1 - 5)</span>
            </div>

            <!-- Dimensi 1: Design -->
            <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 text-md">1. Design Effectiveness (20%)</span>
                  <p class="text-md text-slate-500">Kesesuaian rancangan kontrol terhadap risiko & SOP</p>
                </div>
                <span class="font-extrabold text-primary-700 bg-white px-2 py-0.5 rounded border border-slate-200 text-md">
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
            <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 text-md">2. Operating Effectiveness (20%)</span>
                  <p class="text-md text-slate-500">Konsistensi eksekusi dan ketiadaan deviasi kontrol</p>
                </div>
                <span class="font-extrabold text-primary-700 bg-white px-2 py-0.5 rounded border border-slate-200 text-md">
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
            <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 text-md">3. Coverage & Completeness (20%)</span>
                  <p class="text-md text-slate-500">Cakupan kontrol pada seluruh transaksi/aktivitas</p>
                </div>
                <span class="font-extrabold text-primary-700 bg-white px-2 py-0.5 rounded border border-slate-200 text-md">
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
            <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 text-md">4. Timeliness & Frequency (20%)</span>
                  <p class="text-md text-slate-500">Ketepatan waktu kontrol mendeteksi / mencegah insiden</p>
                </div>
                <span class="font-extrabold text-primary-700 bg-white px-2 py-0.5 rounded border border-slate-200 text-md">
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
            <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200/80 space-y-2">
              <div class="flex items-center justify-between">
                <div>
                  <span class="font-bold text-slate-800 text-md">5. Automation & Monitoring (20%)</span>
                  <p class="text-md text-slate-500">Tingkat otomatisasi sistemik & continuous monitoring</p>
                </div>
                <span class="font-extrabold text-primary-700 bg-white px-2 py-0.5 rounded border border-slate-200 text-md">
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
          <div class="bg-primary-50 p-4 rounded-xl border border-primary-100 flex items-center justify-between">
            <div>
              <span class="text-md text-primary-900 font-bold">Total Weighted Effectiveness Score:</span>
              <p class="text-md text-primary-700">
                Interpretasi: <strong>{{ getItemRating(calculatedModalScorePercent).rating }}</strong>
              </p>
            </div>
            <span class="text-2xl font-extrabold text-primary-700">{{ calculatedModalScorePercent }}%</span>
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

definePageMeta({
  middleware: 'auth'
})

const rcmStore = useRCMStore()
const riskProfileStore = useRiskProfileStore()
const mitigationStore = useMitigationStore()

const searchQuery = ref('')
const showRatingTable = ref(true)
const isModalOpen = ref(false)
const isEditMode = ref(false)
const selectedRiskId = ref('')
const selectedMitigationId = ref('')

const columns = computed(() => [
  { accessorKey: 'risk_code', id: 'risk_code', header: 'Kode / Risiko', class: 'w-[16%]' },
  { accessorKey: 'control_code', id: 'control_code', header: 'Risk Control ID & Deskripsi Mitigasi', class: 'w-[20%]' },
  { accessorKey: 'department', id: 'department', header: 'Departemen / PIC', class: 'w-[14%]' },
  { accessorKey: 'design_effectiveness_rating', id: 'design_effectiveness_rating', header: 'Design (1-5)', class: 'w-[6%] text-center' },
  { accessorKey: 'operating_effectiveness_rating', id: 'operating_effectiveness_rating', header: 'Operating (1-5)', class: 'w-[6%] text-center' },
  { accessorKey: 'coverage_completeness_rating', id: 'coverage_completeness_rating', header: 'Coverage (1-5)', class: 'w-[6%] text-center' },
  { accessorKey: 'timeliness_rating', id: 'timeliness_rating', header: 'Timeliness (1-5)', class: 'w-[6%] text-center' },
  { accessorKey: 'automation_monitoring_rating', id: 'automation_monitoring_rating', header: 'Automation (1-5)', class: 'w-[6%] text-center' },
  { accessorKey: 'total_weighted_score', id: 'total_weighted_score', header: 'Total Score', class: 'w-[7%] text-center' },
  { accessorKey: 'rating', id: 'rating', header: 'Rating Efektivitas', class: 'w-[9%] text-center' },
  { accessorKey: 'actions', id: 'actions', header: 'Aksi', class: 'w-[4%] text-right' }
])

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
      return 'bg-emerald-500 text-white font-bold px-2.5 py-1 rounded-md text-md shadow-md whitespace-nowrap inline-block'
    case 'Effective':
      return 'bg-sky-500 text-white font-bold px-2.5 py-1 rounded-md text-md shadow-md whitespace-nowrap inline-block'
    case 'Moderately Effective':
      return 'bg-amber-500 text-white font-bold px-2.5 py-1 rounded-md text-md shadow-md whitespace-nowrap inline-block'
    case 'Weak':
      return 'bg-orange-500 text-white font-bold px-2.5 py-1 rounded-md text-md shadow-md whitespace-nowrap inline-block'
    default:
      return 'bg-red-500 text-white font-bold px-2.5 py-1 rounded-md text-md shadow-md whitespace-nowrap inline-block'
  }
}

const getRatingBtnClass = (current: number, star: number) => {
  if (current === star) {
    return 'bg-primary-600 text-white shadow-sm font-bold border border-primary-700'
  }
  return 'bg-white text-slate-700 hover:bg-slate-100 border border-slate-300'
}

const calculatedModalScorePercent = computed(() => {
  return rcmStore.calculateItemScorePercent(formData.value)
})

const onRiskSelected = () => {
  if (!selectedRiskId.value) return
  const risk = riskProfileStore.getRiskById(selectedRiskId.value)
  if (risk) {
    formData.value.risk_id = String(risk.id)
    formData.value.risk_code = riskProfileStore.getFormattedId(risk)
    formData.value.risk_event = risk.name
    formData.value.department = risk.branch || risk.category || 'Head Office'

    const mits = mitigationStore.getMitigationsByRiskId(String(risk.id))
    const firstMit = mits && mits.length > 0 ? mits[0] : undefined
    if (firstMit) {
      selectedMitigationId.value = firstMit.id
      onControlSelected()
    } else {
      selectedMitigationId.value = ''
      formData.value.control_code = 'CTL-' + formData.value.risk_code
      formData.value.control_description = 'Aktivitas mitigasi & pengawasan internal untuk ' + risk.name
      formData.value.control_owner = 'Department Lead'
    }
  }
}

const onControlSelected = () => {
  if (!selectedMitigationId.value) return
  const mits = availableMitigations.value
  const found = mits.find(m => m.id === selectedMitigationId.value)
  if (found) {
    formData.value.control_code = found.riskControlId || ('CTL-' + formData.value.risk_code)
    formData.value.control_description = found.mitigationPlan || found.notes || ''
    formData.value.control_owner = found.pic || 'Department Lead'
  }
}

const openAddModal = () => {
  isEditMode.value = false
  selectedRiskId.value = ''
  selectedMitigationId.value = ''
  
  const defaultRisk = riskProfileStore.risks[0]
  const defaultCode = defaultRisk ? riskProfileStore.getFormattedId(defaultRisk) : 'FIN-001'
  const defaultEvent = defaultRisk ? defaultRisk.name : 'Target pendapatan dan laba tidak tercapai'
  const defaultDept = defaultRisk ? (defaultRisk.branch || defaultRisk.category) : 'Head Office'

  formData.value = {
    risk_id: defaultRisk ? String(defaultRisk.id) : '1',
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

  if (defaultRisk) {
    selectedRiskId.value = String(defaultRisk.id)
    onRiskSelected()
  }

  isModalOpen.value = true
}

const openEditModal = (item: RCMItem) => {
  isEditMode.value = true
  formData.value = JSON.parse(JSON.stringify(item))
  selectedRiskId.value = item.risk_id || ''
  isModalOpen.value = true
}

const saveForm = async () => {
  if (!formData.value.risk_code || !formData.value.risk_event || !formData.value.control_description) {
    alert('Mohon pilih risiko dan lengkapi deskripsi kontrol.')
    return
  }

  if (isEditMode.value && formData.value.id) {
    await rcmStore.updateRCMItem(formData.value as RCMItem)
  } else {
    await rcmStore.addRCMItem(formData.value as any)
  }
  isModalOpen.value = false
}

const confirmDelete = async (id: string) => {
  if (await useGlobalModalStore().confirmDelete({ description: 'Apakah Anda yakin ingin menghapus baris Risk Control Matrix ini?' })) {
    await rcmStore.deleteRCMItem(id)
  }
}
</script>
