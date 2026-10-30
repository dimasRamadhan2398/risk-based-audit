<template>
  <div class="grid grid-cols-1 lg:grid-cols-4 h-full bg-gray-50 dark:bg-gray-950">
    <!-- Sidebar Navigation Index (Corporate Report Builder Layout) -->
    <div class="hidden lg:block lg:col-span-1 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-6 overflow-y-auto space-y-6">
      <div class="text-md font-bold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryForm.navTitle') }}</div>
      <nav class="space-y-1">
        <a
          v-for="sec in sections"
          :key="sec.id"
          :href="'#' + sec.id"
          class="flex items-center gap-2 px-3 py-2 text-sm font-medium rounded-md hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          :class="activeSection === sec.id ? 'text-primary-600 bg-primary-50 dark:bg-primary-950 dark:text-primary-400 font-bold' : 'text-gray-600 dark:text-gray-400'"
          @click.prevent="scrollToSection(sec.id)"
        >
          <span class="size-5 rounded-full flex items-center justify-center border text-md font-mono" :class="activeSection === sec.id ? 'border-primary-500 bg-primary-100 dark:bg-primary-900' : 'border-gray-300 dark:border-gray-700'">
            {{ sec.index }}
          </span>
          {{ sec.title }}
        </a>
      </nav>

      <!-- Locking Info / Actions -->
      <div class="border-t border-gray-200 dark:border-gray-800 pt-6 space-y-4">
        <div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-2 border">
          <div class="text-md font-semibold text-gray-500 flex items-center gap-1.5">
            <UIcon name="i-lucide-shield-alert" class="size-4" />
            {{ t('executiveSummaryForm.lock.title') }}
          </div>
          <div class="text-sm font-bold text-gray-800 dark:text-white flex items-center gap-1.5">
            <span class="size-2 rounded-full" :class="store.form.status === 'Approved' ? 'bg-success-500' : 'bg-warning-500'"></span>
            {{ store.form.status === 'Approved' ? t('executiveSummaryForm.lock.approved') : t('executiveSummaryForm.lock.draft') }}
          </div>
          <p class="text-md text-gray-400 leading-normal">
            {{ store.form.status === 'Approved'
              ? t('executiveSummaryForm.lock.approvedDescription')
              : t('executiveSummaryForm.lock.draftDescription') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="col-span-1 lg:col-span-3 overflow-y-auto p-6 lg:p-8 h-full space-y-8 bg-white dark:bg-gray-900 shadow-sm" ref="scrollContainer" @scroll="onScroll">
      
      <!-- Sync Warning Notification (Rule 2) -->
      <UAlert
        v-if="isSyncWarning"
        icon="i-lucide-alert-triangle"
        color="warning"
        variant="solid"
        :title="t('executiveSummaryForm.extra.syncWarning')"
        :description="t('executiveSummaryForm.extra.syncWarningDescription')"
        class="shadow-sm border-l-4 border-warning-600 mb-2"
      />

      <form @submit.prevent class="space-y-12">
        <!-- 1. Metadata & Document Upload -->
        <section id="sec-upload" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">I.</span> {{ t('executiveSummaryForm.upload.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.upload.description') }}</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <UFormField :label="t('executiveSummaryForm.upload.lhaLabel')">
              <USelectMenu
                v-model="selectedLhaId"
                :items="lhaDropdownOptions"
                :placeholder="t('executiveSummaryForm.upload.lhaPlaceholder')"
                class="w-full font-semibold"
                :disabled="isLocked"
                @update:modelValue="onLhaSelect"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.reportYear')" required>
              <USelectMenu
                v-model="store.form.tahun"
                :items="[2026, 2025, 2024, 2023]"
                :placeholder="t('executiveSummaryForm.placeholders.year')"
                class="w-full font-semibold"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryForm.extra.month')" required>
              <USelectMenu
                v-model="store.form.periodeBulan"
                :items="monthOptions"
                value-key="value"
                :placeholder="t('annualAudit.form.selectMonths')"
                class="w-full font-semibold"
                :disabled="isLocked"
                @update:modelValue="onMonthChange"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryForm.extra.documentNumber')" required>
              <UInput
                v-model="store.form.nomorDokumen"
                :placeholder="t('executiveSummaryForm.extra.documentNumberExample')"
                class="w-full font-mono text-sm"
                disabled
              />
            </UFormField>
          </div>

          <!-- Document Upload Field -->
          <div class="bg-gray-50 dark:bg-gray-800/50 p-6 rounded-xl border border-dashed border-gray-300 dark:border-gray-700">
            <UFormField :label="t('executiveSummaryForm.upload.documentLabel')" required>
              <div class="space-y-4">
                <div class="flex items-center gap-4">
                  <!-- File Input simulation -->
                  <input
                    type="file"
                    ref="fileInput"
                    accept=".pdf,.docx"
                    class="hidden"
                    @change="handleFileUpload"
                    :disabled="isLocked"
                  />
                  <UButton
                    color="neutral"
                    variant="solid"
                    icon="i-lucide-upload"
                    :label="t('executiveSummaryForm.upload.chooseFile')"
                    :disabled="isLocked"
                    @click="fileInput?.click()"
                  />
                  <span v-if="store.form.dokumenPath" class="text-sm text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                    <UIcon name="i-lucide-file-check" class="text-success-500 size-5" />
                    {{ store.form.dokumenPath }}
                  </span>
                  <span v-else class="text-sm text-gray-400">{{ t('executiveSummaryForm.upload.noFile') }}</span>
                </div>
                <p class="text-md text-gray-400">
                  {{ t('executiveSummaryForm.upload.formats') }}
                </p>
              </div>
            </UFormField>
          </div>
        </section>

        <!-- 2. Section Narrative -->
        <section id="sec-narrative" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">II.</span> {{ t('executiveSummaryForm.extra.narrativeSection') }} </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.narrative.description') }}</p>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-md font-bold text-gray-400 uppercase">{{ t('executiveSummaryForm.narrative.label') }}</span>
              <UButton
                v-if="!isLocked"
                color="neutral"
                variant="ghost"
                icon="i-lucide-refresh-cw"
                :label="t('executiveSummaryForm.narrative.reset')"
                size="md"
                @click="resetNarrativeToDefault"
              />
            </div>
            <UTextarea
              v-model="store.form.narrative"
              :placeholder="t('executiveSummaryForm.narrative.placeholder')"
              :rows="6"
              class="w-full"
              :disabled="isLocked"
            />
          </div>
        </section>

        <!-- 3. Section Statistik Kompilasi -->
        <section id="sec-stats" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">III.</span> {{ t('executiveSummaryForm.extra.statisticsSection') }} </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryCompilation.stats.description') }}</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <UFormField :label="t('executiveSummaryCompilation.reportCount')" required>
              <UInput
                type="number"
                v-model.number="store.form.jumlahLaporan"
                :placeholder="t('executiveSummaryCompilation.reportCountPlaceholder')"
                class="w-full font-bold"
                :disabled="isLocked"
                min="0"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.recommendationCount')" required>
              <UInput
                type="number"
                v-model.number="store.form.jumlahRekomendasi"
                :placeholder="t('executiveSummaryCompilation.recommendationCountPlaceholder')"
                class="w-full font-bold"
                :disabled="isLocked"
                min="0"
              />
            </UFormField>

            <!-- Auto calculated fields -->
            <UFormField :label="t('executiveSummaryCompilation.totalFindings')">
              <UInput
                v-model="totalTemuanSummary"
                class="w-full font-bold bg-gray-50 dark:bg-gray-800"
                disabled
                :title="t('executiveSummaryCompilation.totalFindingsHint')"
              >
                <template #trailing>
                  <span class="text-md text-gray-400">{{ t('executiveSummaryCompilation.findingsUnit') }}</span>
                </template>
              </UInput>
            </UFormField>
          </div>

          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4">
            <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryCompilation.riskBreakdown') }}</h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <UFormField :label="t('executiveSummaryForm.extra.highRisk')">
                <UInput
                  type="number"
                  :model-value="riskBuckets.high"
                  class="w-full border-l-4 border-error-500 font-bold bg-gray-50 dark:bg-gray-800"
                  disabled
                  :title="t('executiveSummaryForm.extra.highRiskHint')"
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.extra.mediumRisk')">
                <UInput
                  type="number"
                  :model-value="riskBuckets.medium"
                  class="w-full border-l-4 border-warning-500 font-bold bg-gray-50 dark:bg-gray-800"
                  disabled
                  :title="t('executiveSummaryForm.extra.mediumRiskHint')"
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.extra.lowRisk')">
                <UInput
                  type="number"
                  :model-value="riskBuckets.low"
                  class="w-full border-l-4 border-success-500 font-bold bg-gray-50 dark:bg-gray-800"
                  disabled
                  :title="t('executiveSummaryForm.extra.lowRiskHint')"
                />
              </UFormField>
            </div>
          </div>
        </section>

        <!-- 4. Section Status Tindak Lanjut -->
        <section id="sec-followup" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">IV.</span> {{ t('executiveSummaryForm.extra.followUpSection') }} </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.followUp.description') }}</p>
          </div>

          <div class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800/80">
                <tr>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryForm.table.status') }}</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">{{ t('executiveSummaryForm.table.count') }}</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">{{ t('executiveSummaryForm.table.percent') }}</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryForm.table.notes') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800 font-medium">
                <tr v-for="(row, idx) in store.form.followUpTable" :key="row.status">
                  <td class="px-6 py-4 text-sm font-bold">
                    <span :class="getStatusBadgeClass(row.status)">{{ row.status }}</span>
                  </td>
                  <td class="px-6 py-2">
                    <UInput
                      type="number"
                      v-model.number="row.jumlah"
                      placeholder="0"
                      class="font-semibold"
                      :disabled="isLocked"
                      min="0"
                      @input="recalculatePercentages"
                    />
                  </td>
                  <td class="px-6 py-4 text-sm font-semibold text-gray-600 dark:text-white">
                    {{ formatPercent(row.persentase) }}%
                  </td>
                  <td class="px-6 py-2">
                    <UTextarea
                      v-model="row.keterangan"
                      :placeholder="t('executiveSummaryForm.followUp.notePlaceholder')"
                      class="w-full"
                      :disabled="isLocked"
                    />
                  </td>
                </tr>
                <!-- Total Row -->
                <tr class="bg-gray-50 dark:bg-gray-800/40 font-bold border-t-2">
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryForm.table.total') }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white font-mono font-extrabold">{{ totalFollowUpCount }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white font-mono font-extrabold">100.0%</td>
                  <td class="px-6 py-4 text-md text-gray-400 italic">{{ t('executiveSummaryForm.table.autoCalculated') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- 5. Section Top 5 Significant Findings -->
        <section id="sec-topfindings" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <div class="flex justify-between items-center">
              <div>
                <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                  <span class="text-primary-500">V.</span> {{ t('executiveSummaryForm.extra.topFindingsSection') }} </h2>
                <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.topFindings.description') }}</p>
              </div>
              <UButton
                v-if="!isLocked && store.form.topFindings.length < 5"
                color="primary"
                variant="soft"
                icon="i-lucide-plus-circle"
                :label="t('executiveSummaryForm.topFindings.add')"
                size="sm"
                @click="addTopFinding"
              />
            </div>
          </div>

          <div v-if="store.form.topFindings.length > 0" class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800/80">
                <tr>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">{{ t('executiveSummaryForm.table.unitDivision') }}</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryForm.table.findingTitle') }}</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-32">{{ t('executiveSummaryForm.table.riskValue') }}</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">{{ t('executiveSummaryForm.table.followUpStatus') }}</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryForm.table.proposedAction') }}</th>
                  <th v-if="!isLocked" class="px-4 py-3 text-center text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-16">{{ t('executiveSummaryForm.table.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
                <tr v-for="(finding, idx) in store.form.topFindings" :key="idx">
                  <td class="px-3 py-2">
                    <USelectMenu
                      v-model="finding.unitDivision"
                      :items="divisionOptions"
                      :placeholder="t('executiveSummaryForm.placeholders.division')"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <UInput
                      v-model="finding.judulTemuan"
                      :placeholder="t('executiveSummaryForm.placeholders.findingTitle')"
                      :disabled="isLocked"
                      maxlength="100"
                      @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-gray-500 mt-1 text-right">
                      {{ finding.judulTemuan ? finding.judulTemuan.length : 0 }}/100
                    </div>
                  </td>
                  <td class="px-3 py-2">
                    <USelectMenu
                      v-model="finding.risiko"
                      :items="['Tinggi', 'Sedang', 'Rendah']"
                      :placeholder="t('executiveSummaryForm.placeholders.risk')"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <USelectMenu
                      v-model="finding.statusTL"
                      :items="['Closed', 'In Progress', 'Overdue']"
                      :placeholder="t('executiveSummaryForm.table.followUpStatus')"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <UTextarea
                      v-model="finding.usulan"
                      :placeholder="t('executiveSummaryForm.placeholders.proposedAction')"
                      :disabled="isLocked"
                    />
                  </td>
                  <td v-if="!isLocked" class="px-3 py-2 text-center">
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-lucide-trash"
                      size="sm"
                      @click="removeTopFinding(idx)"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          
          <div v-else class="text-center py-8 bg-gray-50 dark:bg-gray-800/30 rounded-lg text-sm text-gray-400 border border-dashed">
            {{ t('executiveSummaryForm.topFindings.empty') }}
            <button v-if="!isLocked" type="button" @click="addTopFinding" class="text-primary-500 font-bold ml-1 hover:underline">
              {{ t('executiveSummaryForm.topFindings.emptyAction') }}
            </button>
          </div>
        </section>

        <!-- 6. Qualitative Analysis (Section V & VII) -->
        <section id="sec-analysis" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">VI.</span> {{ t('executiveSummaryForm.extra.analysisSection') }} </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.analysis.description') }}</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField :label="t('executiveSummaryForm.extra.recurringThemes')">
              <UTextarea
                v-model="store.form.akarMasalah"
                :placeholder="t('executiveSummaryForm.analysis.rootCausePlaceholder')"
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryForm.extra.managementConclusion')">
              <UTextarea
                v-model="store.form.kesimpulan"
                :placeholder="t('executiveSummaryForm.analysis.conclusionPlaceholder')"
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>
          </div>

          <!-- Electronic Signatures -->
          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4">
            <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryForm.signature.title') }}</h4>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
              <UFormField :label="t('executiveSummaryForm.signature.place')" required>
                <UInput
                  v-model="store.form.signatureTempat"
                  :placeholder="t('executiveSummaryForm.signature.placePlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity('Tempat TTD maksimal 20 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureTempat ? store.form.signatureTempat.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.date')" required>
                <AppDatePicker
                  v-model="store.form.signatureTanggal"
                  :disabled="isLocked"
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.headName')" required>
                <UInput
                  v-model="store.form.signatureNamaKepala"
                  :placeholder="t('executiveSummaryForm.signature.headNamePlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity('Nama Kepala SPI maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNamaKepala ? store.form.signatureNamaKepala.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.headNik')" required>
                <UInput
                  v-model="store.form.signatureNIK"
                  :placeholder="t('executiveSummaryForm.signature.headNikPlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity('NIK Kepala SPI maksimal 20 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNIK ? store.form.signatureNIK.length : 0 }}/50
                </div>
              </UFormField>
            </div>
          </div>
        </section>

        <!-- 7. Section VI: System Generated Charts -->
        <section id="sec-charts" class="space-y-6 scroll-mt-6 bg-gray-50 dark:bg-gray-900/50 p-6 rounded-2xl border border-gray-200 dark:border-gray-800">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">VII.</span> {{ t('executiveSummaryForm.extra.chartsSection') }} </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.extra.chartsDescription') }}</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <!-- Pie Chart: Status Tindak Lanjut -->
            <UCard class="flex flex-col h-80" :ui="{ body: 'flex-1 relative flex items-center justify-center p-4' }">
              <template #header>
                <h5 class="text-md font-bold uppercase tracking-wider text-gray-400 text-center">{{ t('executiveSummaryForm.extra.followUpPercent') }}</h5>
              </template>
              <div class="size-full max-h-52 max-w-52">
                <Doughnut v-if="renderCharts" :data="pieChartData" :options="chartOptions" />
              </div>
            </UCard>

            <!-- Bar Chart: Temuan Per Bulan -->
            <UCard class="flex flex-col h-80" :ui="{ body: 'flex-1 relative flex items-center justify-center p-4' }">
              <template #header>
                <h5 class="text-md font-bold uppercase tracking-wider text-gray-400 text-center">{{ t('executiveSummaryForm.extra.monthlyFindings') }}</h5>
              </template>
              <div class="w-full h-52">
                <Bar v-if="renderCharts" :data="barChartData" :options="{ ...chartOptions, scales: { y: { beginAtZero: true } } }" />
              </div>
            </UCard>

            <!-- Line Chart: Aging Rekomendasi -->
            <UCard class="flex flex-col h-80" :ui="{ body: 'flex-1 relative flex items-center justify-center p-4' }">
              <template #header>
                <h5 class="text-md font-bold uppercase tracking-wider text-gray-400 text-center">{{ t('executiveSummaryForm.extra.recommendationAging') }}</h5>
              </template>
              <div class="w-full h-52">
                <Line v-if="renderCharts" :data="lineChartData" :options="{ ...chartOptions, scales: { y: { min: 0, max: 100 } } }" />
              </div>
            </UCard>
          </div>
        </section>

        <!-- 8. Section VIII: Matriks Induk Kompilasi -->
        <section id="sec-matrix" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
              <div>
                <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                  <span class="text-primary-500">VIII.</span> {{ t('executiveSummaryForm.extra.matrixSection') }} </h2>
                <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.extra.matrixDescription') }}</p>
              </div>
              <div class="flex gap-2 flex-wrap">
                <UButton
                  v-if="canAccessFiles && (!isLocked)"
                  color="success"
                  variant="outline"
                  icon="i-lucide-download"
                  :label="t('executiveSummaryForm.extra.downloadTemplate')"
                  size="sm"
                  @click="downloadExcelTemplate"
                />
              </div>
            </div>
          </div>

          <!-- Matrix Grid Table -->
          <div class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
            <div class="overflow-x-auto w-full max-h-[500px] overflow-y-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 font-medium text-md">
                <thead class="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10">
                  <tr class="divide-x divide-gray-200 dark:divide-gray-800">
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-28">{{ t('executiveSummaryForm.extra.columns.number') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-24">{{ t('executiveSummaryForm.extra.columns.division') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-40">{{ t('executiveSummaryForm.extra.columns.workUnit') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-32">{{ t('executiveSummaryForm.extra.columns.businessProcess') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-52">{{ t('executiveSummaryForm.extra.columns.finding') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-28">{{ t('executiveSummaryForm.extra.columns.risk') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-52">{{ t('executiveSummaryForm.extra.columns.recommendation') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-32">{{ t('executiveSummaryForm.extra.columns.dueDate') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-32">{{ t('executiveSummaryForm.extra.columns.pic') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-20">{{ t('executiveSummaryForm.extra.columns.progress') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-28">{{ t('executiveSummaryForm.extra.columns.status') }}</th>
                    <th class="px-3 py-3 text-left font-bold text-gray-500 uppercase w-32">{{ t('executiveSummaryForm.extra.columns.evidence') }}</th>
                    <th v-if="!isLocked" class="px-3 py-3 text-center font-bold text-gray-500 uppercase w-12 sticky right-0 bg-gray-100 dark:bg-gray-800">X</th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
                  <tr v-for="(row, idx) in store.form.matriksKompilasi" :key="idx" class="divide-x divide-gray-100 dark:divide-gray-800 hover:bg-gray-50/50">
                    <td class="p-1">
                      <UInput 
                        v-model="row.nomor" 
                        size="md" 
                        class="font-mono" 
                        placeholder="001/SPI/2026" 
                        :disabled="isLocked" 
                        maxlength="50"
                        @invalid="($event.target as any)?.setCustomValidity('Nomor temuan maksimal 50 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ row.nomor ? row.nomor.length : 0 }}/50
                      </div>
                    </td>
                    <td class="p-1">
                      <USelectMenu v-model="row.division" :items="divisionOptions" size="md" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <UInput 
                        v-model="row.unitKerja" 
                        size="md" 
                        :placeholder="t('executiveSummaryForm.extra.divisionExample')"
                        :disabled="isLocked" 
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('Unit kerja maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ row.unitKerja ? row.unitKerja.length : 0 }}/100
                      </div>
                    </td>
                    <td class="p-1">
                      <UInput 
                        v-model="row.prosesBisnis" 
                        size="md" 
                        placeholder="O&M" 
                        :disabled="isLocked" 
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('Proses bisnis maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ row.prosesBisnis ? row.prosesBisnis.length : 0 }}/100
                      </div>
                    </td>
                    <td class="p-1">
                      <UTextarea v-model="row.judulTemuan" size="md" :placeholder="t('executiveSummaryForm.extra.findingPlaceholder')" :rows="1" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <USelectMenu v-model="row.nilaiRisiko" :items="['Tinggi', 'Sedang', 'Rendah']" size="md" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <UTextarea v-model="row.rekomendasi" size="md" :placeholder="t('executiveSummaryForm.extra.actionPlaceholder')" :rows="1" :disabled="isLocked" />
                    </td>
                    <td class="p-1 min-w-36">
                      <AppDatePicker v-model="row.dueDate" size="md" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <UInput 
                        v-model="row.picUnit" 
                        size="md" 
                        :placeholder="t('executiveSummaryForm.extra.managerExample')"
                        :disabled="isLocked" 
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('PIC Unit maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ row.picUnit ? row.picUnit.length : 0 }}/100
                      </div>
                    </td>
                    <td class="p-1">
                      <UInput type="number" v-model.number="row.progres" size="md" placeholder="0" min="0" max="100" class="w-16" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <USelectMenu v-model="row.status" :items="['Closed', 'In Progress', 'Overdue']" size="md" :disabled="isLocked" />
                    </td>
                    <td class="p-1">
                      <div class="flex items-center gap-1">
                        <UInput 
                        v-model="row.buktiTL" 
                        size="md" 
                        :placeholder="t('executiveSummaryForm.extra.evidencePlaceholder')"
                        :disabled="isLocked" 
                        maxlength="100"
                        @invalid="($event.target as any)?.setCustomValidity('Bukti TL maksimal 100 karakter dan wajib diisi')"
                        @input="($event.target as any)?.setCustomValidity('')"
                      />
                      <div class="text-xs text-gray-500 mt-1 text-right">
                        {{ row.buktiTL ? row.buktiTL.length : 0 }}/100
                      </div>
                        <UButton
                          v-if="!isLocked"
                          color="neutral"
                          variant="ghost"
                          icon="i-lucide-paperclip"
                          size="md"
                          :title="t('executiveSummaryForm.extra.attachFile')"
                          @click="simulateAttachmentUpload(idx)"
                        />
                      </div>
                    </td>
                    <td v-if="!isLocked" class="p-1 text-center sticky right-0 bg-white dark:bg-gray-900">
                      <UButton
                        color="error"
                        variant="ghost"
                        icon="i-lucide-x"
                        size="md"
                        @click="removeMatrixRow(idx)"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Tombol di paling bawah matrix grid table untuk menambahkan row baru matriks induk kompilasi temuan -->
            <div v-if="!isLocked" class="p-3 bg-gray-50 dark:bg-gray-800/60 border-t border-gray-200 dark:border-gray-800 flex items-center justify-between gap-3">
              <UButton
                type="button"
                color="primary"
                variant="soft"
                icon="i-lucide-plus"
                :label="t('executiveSummaryForm.extra.addFindingRow')"
                size="sm"
                @click="addMatrixRow"
              />
              <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">
                {{ store.form.matriksKompilasi.length }} {{ t('executiveSummaryForm.extra.findingUnit') }} </span>
            </div>
          </div>
          
          <div v-if="store.form.matriksKompilasi.length === 0" class="text-center py-10 bg-gray-50 dark:bg-gray-800/30 rounded-lg text-sm text-gray-400 border border-dashed"> {{ t('executiveSummaryForm.extra.emptyMatrix') }} <button v-if="!isLocked" type="button" @click="addMatrixRow" class="text-primary-500 font-bold ml-1 hover:underline"> {{ t('executiveSummaryForm.extra.addManualRow') }} </button>
          </div>
        </section>
      </form>
    </div>

    <!-- Sticky Bottom Footer Controls -->
    <div class="col-span-1 lg:col-span-4 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex justify-between items-center sticky bottom-0 z-20">
      <div>
        <UButton
          :label="t('executiveSummaryForm.actions.close')"
          color="neutral"
          variant="outline"
          class="font-semibold"
          @click="() => {store.showModal = false}"
        />
      </div>

      <div class="flex items-center gap-3">
        <!-- Error alert message displays if necessary -->
        <span v-if="store.errorMsg" class="text-md text-error-500 font-semibold max-w-md truncate">{{ store.errorMsg }}</span>
        
        <!-- Workflow buttons -->
        <template v-if="!store.isViewing">
          <UButton
            v-if="!isLocked"
            color="primary"
            variant="solid"
            icon="i-lucide-save"
            :label="t('executiveSummaryForm.actions.saveDraft')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="saveReportDraft"
          />
          <UButton
            v-if="!isLocked && isChiefAuditExecutive"
            color="success"
            variant="solid"
            icon="i-lucide-check-circle"
            :label="t('executiveSummaryForm.extra.approveReport')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="approveReportDirectly"
          />
        </template>
        
        <!-- Locked/Unlocked Overrides -->
        <template v-else>
          <UButton
            v-if="store.form.status === 'Draft' && isChiefAuditExecutive"
            color="success"
            variant="solid"
            icon="i-lucide-check"
            :label="t('executiveSummaryForm.actions.approve')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="approveReportDirectly"
          />
          <UButton
            v-if="store.form.status === 'Approved' && isHigherAuthority"
            color="warning"
            variant="solid"
            icon="i-lucide-unlock"
            :label="t('executiveSummaryForm.extra.revertDraft')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="revertToDraft"
          />
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'
import { riskBucketBreakdown, riskBucketTotal, riskFieldsToBreakdown } from '~/utils/executiveSummaryCompilation'
import { Doughnut, Bar, Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

const { t } = useI18n()
const { canAccessFiles } = useRbac()

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, ArcElement, Title, Tooltip, Legend, Filler)

const store = useExecutiveSummaryStore()
const authStore = useAuthStore()

const fileInput = ref<HTMLInputElement | null>(null)
const activeSection = ref('sec-upload')
const renderCharts = ref(false)
const scrollContainer = ref<HTMLElement | null>(null)

// Navigation links
const sections = computed(() => [
  { id: 'sec-upload', index: '1', title: t('executiveSummaryCompilation.nav.upload') },
  { id: 'sec-narrative', index: '2', title: t('executiveSummaryCompilation.nav.narrative') },
  { id: 'sec-stats', index: '3', title: t('executiveSummaryCompilation.nav.stats') },
  { id: 'sec-followup', index: '4', title: t('executiveSummaryCompilation.nav.followUp') },
  { id: 'sec-topfindings', index: '5', title: t('executiveSummaryCompilation.nav.topFindings') },
  { id: 'sec-analysis', index: '6', title: t('executiveSummaryCompilation.nav.analysis') },
  { id: 'sec-charts', index: '7', title: t('executiveSummaryForm.extra.navCharts') },
  { id: 'sec-matrix', index: '8', title: t('executiveSummaryForm.extra.navMatrix') }
])

// Dropdown options
const divisionOptions = ['OP', 'KK/KSD', 'IT', 'FIN', 'HR', 'LEG']

const monthValues = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
const localizeMonth = (month: string) => t(`executiveSummaryForm.extra.months.${month}`)
const quarterMonths = computed(() => monthValues.slice((store.form.quarter - 1) * 3, store.form.quarter * 3))
const monthOptions = computed(() => quarterMonths.value.map(value => ({ label: localizeMonth(value), value })))

// Role computed checks
const isChiefAuditExecutive = computed(() => {
  return authStore.user?.roles.includes(UserRole.EXECUTIVE) || authStore.user?.roles.includes(UserRole.ADMIN)
})

const isHigherAuthority = computed(() => {
  return authStore.user?.roles.includes(UserRole.ADMIN)
})

const isLocked = computed(() => {
  return store.form.status === 'Approved'
})

// Section II display buckets, derived from the five stored risk levels
const riskBuckets = computed(() => riskBucketBreakdown(riskFieldsToBreakdown(store.form)))

// Section II computed auto-sum
const totalTemuanSummary = computed(() => riskBucketTotal(riskBuckets.value))

// Section III computed auto-sums
const totalFollowUpCount = computed(() => {
  return store.form.followUpTable.reduce((acc, row) => acc + (row.jumlah || 0), 0)
})

// Rule 2: Warning sync check
const isSyncWarning = computed(() => {
  const matrixTinggi = store.form.matriksKompilasi.filter(r => r.nilaiRisiko === 'Tinggi').length
  const matrimdedang = store.form.matriksKompilasi.filter(r => r.nilaiRisiko === 'Sedang').length
  const matrixRendah = store.form.matriksKompilasi.filter(r => r.nilaiRisiko === 'Rendah').length

  const buckets = riskBuckets.value

  return buckets.high !== matrixTinggi || buckets.medium !== matrimdedang || buckets.low !== matrixRendah
})

// Force recalculate Section III percentages
const recalculatePercentages = () => {
  const total = totalFollowUpCount.value
  store.form.followUpTable.forEach(row => {
    row.persentase = total > 0 ? (row.jumlah / total) * 100 : 0
  })
}

// Watch matrix changes to sync status counts
watch(() => store.form.matriksKompilasi, () => {
  // Sync status counts in Section III
  const closed = store.form.matriksKompilasi.filter(r => r.status === 'Closed').length
  const inProg = store.form.matriksKompilasi.filter(r => r.status === 'In Progress').length
  const overdue = store.form.matriksKompilasi.filter(r => r.status === 'Overdue').length

  if (store.form.matriksKompilasi.length > 0 && store.form.followUpTable.length >= 3) {
    store.form.followUpTable[0]!.jumlah = closed
    store.form.followUpTable[1]!.jumlah = inProg
    store.form.followUpTable[2]!.jumlah = overdue
    recalculatePercentages()
  }
}, { deep: true })

// Helper formatting percentage
const formatPercent = (val: number) => {
  return val ? val.toFixed(1) : '0.0'
}

// Dropdown change behavior
const onMonthChange = (val: any) => {
  if (!store.form.narrative || store.form.narrative.startsWith('Periode ')) {
    store.form.narrative = store.defaultNarrativeTemplate(store.form.periodeBulan, store.form.tahun)
  }
}

const resetNarrativeToDefault = () => {
  store.form.narrative = store.defaultNarrativeTemplate(store.form.periodeBulan, store.form.tahun)
}

// File Upload Handler (with corrupted file simulation & validation)
const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Size limit 10MB
  if (file.size > 10 * 1024 * 1024) {
    alert(t('executiveSummaryForm.messages.fileTooLarge'))
    return
  }

  // Type limit
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (ext !== 'pdf' && ext !== 'docx' && ext !== 'doc') {
    alert(t('executiveSummaryForm.messages.unsupportedFormat'))
    return
  }

  // Simulated file corruption error handling
  if (file.name.toLowerCase().includes('corrupt') || file.name.toLowerCase().includes('damaged')) {
    alert(t('executiveSummaryForm.messages.corruptedFile'))
    return
  }

  // Simulated success
  store.form.dokumenPath = file.name
}

// Section IV: Top 5 dynamic list methods
const addTopFinding = () => {
  if (store.form.topFindings.length >= 5) return
  store.form.topFindings.push({
    unitDivision: 'OP',
    judulTemuan: '',
    risiko: 'Tinggi',
    statusTL: 'In Progress',
    usulan: ''
  })
}

const removeTopFinding = (idx: number) => {
  store.form.topFindings.splice(idx, 1)
}

// Section VIII: Matriks Induk manual addition methods
const addMatrixRow = () => {
  store.form.matriksKompilasi.push({
    nomor: `00${store.form.matriksKompilasi.length + 1}/SPI/2026`,
    division: 'OP',
    unitKerja: '',
    prosesBisnis: '',
    judulTemuan: '',
    nilaiRisiko: 'Sedang',
    rekomendasi: '',
    dueDate: (new Date().toISOString().split('T')[0]) as string,
    picUnit: '',
    progres: 0,
    status: 'In Progress',
    buktiTL: ''
  })
}

const removeMatrixRow = (idx: number) => {
  store.form.matriksKompilasi.splice(idx, 1)
}

const simulateAttachmentUpload = (idx: number) => {
  const attachmentName = prompt('Masukkan nama file attachment bukti TL:', 'BA Serah Terima Alat.pdf')
  if (attachmentName) {
    const row = store.form.matriksKompilasi[idx]
    if (row) {
      row.buktiTL = attachmentName
    }
  }
}

// Download Excel Template Simulation
const downloadExcelTemplate = () => {
  // Creating virtual CSV template download
  const headers = 'Nomor,Division,Unit Kerja,Proses Bisnis,Judul Temuan,Nilai Risiko,Rekomendasi,Due Date,PIC Unit,% Progres,Status,Bukti TL'
  const sample = '001/SPI/2026,OP,Division Operation Personnel,O&M,Keterlambatan Kalibrasi Alat Berat,Tinggi,Segera lakukan kalibrasi,2026-03-15,Manager O&M,40,Overdue,BA Kalibrasi 1.pdf'
  const csvContent = 'data:text/csv;charset=utf-8,' + headers + '\n' + sample
  const encodedUri = encodeURI(csvContent)
  const link = document.createElement('a')
  link.setAttribute('href', encodedUri)
  link.setAttribute('download', 'template_matriks_kompilasi.csv')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}



// Charting Visuals data providers
const statusBadgeColorMap = {
  Closed: 'success',
  'In Progress': 'info',
  Overdue: 'error'
}

const getStatusBadgeClass = (status: 'Closed' | 'In Progress' | 'Overdue') => {
  const colors = {
    Closed: 'bg-success-100 text-success-800 dark:bg-success-950 dark:text-success-300 px-2.5 py-0.5 rounded text-md font-bold uppercase',
    'In Progress': 'bg-info-100 text-info-800 dark:bg-info-950 dark:text-info-300 px-2.5 py-0.5 rounded text-md font-bold uppercase',
    Overdue: 'bg-error-100 text-error-800 dark:bg-error-950 dark:text-error-300 px-2.5 py-0.5 rounded text-md font-bold uppercase'
  }
  return colors[status]
}

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        boxWidth: 12,
        font: { size: 10 }
      }
    }
  }
}

// Pie Chart (Section III Data)
const pieChartData = computed(() => {
  const labels = ['Closed', 'In Progress', 'Overdue']
  const data = store.form.followUpTable.map(r => r.jumlah || 0)
  
  return {
    labels,
    datasets: [{
      data,
      backgroundColor: ['#1fc16b', '#00d4f9', '#fc423f'],
      hoverOffset: 4
    }]
  }
})

// Bar Chart (Monthly distribution from Matrix)
const barChartData = computed(() => {
  // Extract month count from matrix due dates
  const counts = Array(12).fill(0)

  store.form.matriksKompilasi.forEach(row => {
    if (!row.dueDate) return
    const date = new Date(row.dueDate)
    if (!isNaN(date.getTime())) {
      counts[date.getMonth()]++
    }
  })

  // Determine active quarter's months to display
  const labels = quarterMonths.value.map(localizeMonth)
  const data = counts.slice((store.form.quarter - 1) * 3, store.form.quarter * 3)



  return {
    labels,
    datasets: [{
      label: t('executiveSummaryForm.extra.chartFindingCount'),
      data,
      backgroundColor: '#ff5c02',
      borderRadius: 4
    }]
  }
})

// Line Chart (Aging Recommendation progress by Month)
const lineChartData = computed(() => {
  const labels = quarterMonths.value.map(localizeMonth)
  let data = [0, 0, 0]

  // Calculate actual average progress from matrix if items exist
  if (store.form.matriksKompilasi.length > 0) {
    const totalProg = store.form.matriksKompilasi.reduce((acc, r) => acc + (r.progres || 0), 0)
    const avgProg = totalProg / store.form.matriksKompilasi.length
    // Plot a line trending toward the current average
    data = [Math.round(avgProg * 0.7), Math.round(avgProg * 0.85), Math.round(avgProg)]
  }

  return {
    labels,
    datasets: [{
      label: t('executiveSummaryForm.extra.chartFollowUpProgress'),
      data,
      borderColor: '#4d00ff',
      backgroundColor: 'rgba(77, 0, 255, 0.1)',
      borderWidth: 2,
      fill: true,
      tension: 0.3
    }]
  }
})

// Form submission & workflow helpers
const saveReportDraft = async () => {
  if (!store.form.nomorDokumen) {
    alert(t('executiveSummaryForm.messages.docNumberRequired'))
    return
  }
  if (!store.form.dokumenPath) {
    alert(t('executiveSummaryForm.messages.documentRequired'))
    return
  }
  
  // Set status Draft
  store.form.status = 'Draft'
  await store.saveForm()
}

const approveReportDirectly = async () => {
  if (!store.form.nomorDokumen) {
    alert(t('executiveSummaryForm.messages.docNumberRequired'))
    return
  }
  if (!store.form.dokumenPath) {
    alert(t('executiveSummaryForm.messages.documentRequired'))
    return
  }

  // Set status Approved
  store.form.status = 'Approved'
  await store.saveForm()
}

const revertToDraft = async () => {
  if (await useGlobalModalStore().confirmDelete({ description: t('executiveSummaryForm.extra.revertConfirm') })) {
    if (store.currentSummary) {
      await store.updateStatus(store.currentSummary.id, 'Draft')
    }
  }
}

// Scrolling Table of Contents logic
const scrollToSection = (id: string) => {
  activeSection.value = id
  const target = document.getElementById(id)
  if (target) {
    target.scrollIntoView({ behavior: 'smooth' })
  }
}

const onScroll = () => {
  const scrollOffset = 150
  for (const sec of sections.value) {
    const el = document.getElementById(sec.id)
    if (el) {
      const rect = el.getBoundingClientRect()
      // Section is at the top of the container
      if (rect.top <= scrollOffset && rect.bottom > scrollOffset) {
        activeSection.value = sec.id
        break
      }
    }
  }
}

// LHA Result Reports Store Integration for 2-Way Sync
import { useAuditResultReportStore } from '~/stores/audit-result-report'

const auditReportStore = useAuditResultReportStore()
if (!auditReportStore.loading && auditReportStore.reportList.length === 0) {
  auditReportStore.fetchReports()
}

const lhaDropdownOptions: Ref<Array<{ label: string; value: string; report: any }>> = computed(() => {
  return auditReportStore.reportList.map(r => ({
    label: `${r.reportNumber || (r as any).report_number} - ${r.reportTitle}`,
    value: r.reportNumber || (r as any).report_number,
    report: r
  }))
})

const selectedLhaId = ref<{ label: string; value: string; report: any } | undefined>(
  lhaDropdownOptions.value.find(o => o.value === store.form.nomorDokumen)
)

const onLhaSelect = (val: any) => {
  if (!val) return
  const selectedNum = typeof val === 'object' ? val.value : val
  const item = auditReportStore.reportList.find(r => (r.reportNumber || (r as any).report_number) === selectedNum)
  if (item) {
    store.form.nomorDokumen = item.reportNumber || (item as any).report_number
    if (item.executiveSummary) {
      store.form.narrative = item.executiveSummary
    }
    if (item.findingsCount) {
      store.form.jumlahRekomendasi = item.findingsCount
    }
  }
}

// Chart rendering delay to prevent sizing glitches in modal
onMounted(() => {
  nextTick(() => {
    setTimeout(() => {
      renderCharts.value = true
    }, 300)
  })
})
</script>
