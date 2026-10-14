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
      
      <form @submit.prevent class="space-y-12">
        <!-- 1. Metadata & Document Upload -->
        <section id="sec-upload" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.upload.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.upload.description') }}</p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <UFormField label="Tahun Laporan" required>
              <USelectMenu
                v-model="store.form.tahun"
                :items="[2026, 2025, 2024, 2023]"
                :placeholder="t('executiveSummaryForm.placeholders.year')"
                class="w-full font-semibold"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.quarter')" required>
              <USelectMenu
                v-model="reportingQuarter"
                :items="quarterOptions"
                :placeholder="t('executiveSummaryForm.placeholders.quarter')"
                class="w-full font-semibold"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField label="Nomor Dokumen Kompilasi" required>
              <UInput
                v-model="store.form.nomorDokumen"
                placeholder="Contoh: ESC-001/2026"
                class="w-full font-bold bg-gray-50 dark:bg-gray-800"
                disabled
              />
            </UFormField>
          </div>

          <!-- Kumpulan Semua Data LHA Sesuai Tahun dan Kuartal -->
          <div class="border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50/70 dark:bg-gray-800/40 p-5 space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-gray-200 dark:border-gray-700/80">
              <div class="flex items-center gap-3">
                <div class="p-2.5 bg-primary-100 dark:bg-primary-900/60 text-primary-600 dark:text-primary-400 rounded-lg">
                  <UIcon name="i-lucide-files" class="size-5" />
                </div>
                <div>
                  <div class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
                    <span>Kumpulan Data LHA Terkompilasi</span>
                    <UBadge color="primary" variant="subtle" size="xs">
                      {{ reportingQuarter }} {{ store.form.tahun }}
                    </UBadge>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                    Daftar seluruh LHA dari surat tugas yang otomatis dirangkum ke dalam laporan kompilasi ini.
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <UBadge color="neutral" variant="outline" size="sm" class="font-mono">
                  Total: {{ quarterlyLhas.length }} LHA
                </UBadge>
                <UButton
                  v-if="!isLocked"
                  color="primary"
                  variant="ghost"
                  icon="i-lucide-refresh-cw"
                  label="Segarkan Data"
                  size="xs"
                  @click="syncFromIndividualLha"
                />
              </div>
            </div>

            <!-- List of LHAs (Table or Empty State) -->
            <div v-if="quarterlyLhas.length > 0" class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 shadow-xs">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-xs">
                <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-600 dark:text-gray-300 font-bold uppercase tracking-wider">
                  <tr>
                    <th class="px-4 py-2.5 text-left">No. Dokumen LHA</th>
                    <th class="px-4 py-2.5 text-left">Surat Tugas (ST)</th>
                    <th class="px-4 py-2.5 text-left">Judul / Objek Audit</th>
                    <th class="px-4 py-2.5 text-left">Divisi</th>
                    <th class="px-4 py-2.5 text-left">Tanggal LHA</th>
                    <th class="px-4 py-2.5 text-center">Temuan</th>
                    <th class="px-4 py-2.5 text-center">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800 font-medium">
                  <tr v-for="lha in quarterlyLhas" :key="lha.id || lha.reportNumber" class="hover:bg-gray-50/60 dark:hover:bg-gray-800/50 transition-colors">
                    <td class="px-4 py-2.5 font-mono font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">
                      {{ lha.reportNumber || (lha as any).report_number }}
                    </td>
                    <td class="px-4 py-2.5 font-mono text-gray-600 dark:text-gray-400 whitespace-nowrap">
                      <span v-if="lha.assignmentLetterId" class="inline-flex items-center gap-1">
                        <UIcon name="i-lucide-file-signature" class="size-3 text-primary-500" />
                        {{ lha.assignmentLetterId }}
                      </span>
                      <span v-else class="text-gray-400 italic">-</span>
                    </td>
                    <td class="px-4 py-2.5 text-gray-800 dark:text-gray-200 max-w-xs truncate" :title="lha.reportTitle">
                      {{ lha.reportTitle }}
                    </td>
                    <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                      <UBadge color="neutral" variant="subtle" size="xs">
                        {{ lha.department || 'OP' }}
                      </UBadge>
                    </td>
                    <td class="px-4 py-2.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                      {{ lha.reportDate || '-' }}
                    </td>
                    <td class="px-4 py-2.5 text-center whitespace-nowrap">
                      <UBadge color="warning" variant="subtle" size="xs" class="font-bold">
                        {{ lha.findingsCount || (lha.findings ? lha.findings.length : 0) }} Temuan
                      </UBadge>
                    </td>
                    <td class="px-4 py-2.5 text-center whitespace-nowrap">
                      <UBadge :color="lha.status === 'Final' || (lha.status as string) === 'Approved' ? 'success' : 'neutral'" variant="soft" size="xs">
                        {{ lha.status }}
                      </UBadge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Empty State when no LHAs for the selected quarter & year -->
            <div v-else class="text-center py-6 px-4 bg-white dark:bg-gray-900 rounded-lg border border-dashed border-gray-300 dark:border-gray-700 space-y-1">
              <UIcon name="i-lucide-file-x" class="size-8 text-gray-400 mx-auto" />
              <div class="text-xs font-semibold text-gray-700 dark:text-gray-300">
                Tidak ada dokumen LHA untuk {{ reportingQuarter }} {{ store.form.tahun }}
              </div>
              <p class="text-[11px] text-gray-400">
                Belum ada LHA yang terbit atau sesuai pada periode ini. Data kompilasi akan otomatis terisi setelah LHA diterbitkan.
              </p>
            </div>
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
              {{ t('executiveSummaryForm.narrative.title') }}
            </h2>
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
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryCompilation.stats.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryCompilation.stats.description') }}</p>
            </div>
            <UButton
              v-if="!isLocked"
              color="primary"
              variant="outline"
              icon="i-lucide-refresh-cw"
              :label="t('executiveSummaryCompilation.syncFromIndividual')"
              size="sm"
              :loading="sourcesLoading"
              @click="syncFromIndividualLha"
            />
          </div>

          <!-- Source status: loading / error / no LHA selected / empty sources -->
          <div v-if="sourcesLoading" class="flex items-center gap-2 text-sm text-gray-400">
            <UIcon name="i-lucide-loader-circle" class="size-4 animate-spin" />
            {{ t('executiveSummaryCompilation.sources.loading') }}
          </div>
          <div v-else-if="sourcesError" class="text-sm text-error-500">
            {{ t('executiveSummaryCompilation.sources.error', { message: sourcesError }) }}
          </div>
          <div v-else-if="!quarterlyLhas.length" class="text-sm text-gray-400">
            {{ t('executiveSummaryCompilation.sources.noLhaInQuarter') }}
          </div>
          <ul v-else-if="sourceEmptyNotes.length" class="text-sm text-gray-400 space-y-1">
            <li v-for="note in sourceEmptyNotes" :key="note">{{ note }}</li>
          </ul>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <UFormField :label="t('executiveSummaryCompilation.reportCount')" :hint="t('executiveSummaryCompilation.sources.hintLha')">
              <UInput
                type="number"
                :model-value="store.form.jumlahLaporan"
                :placeholder="t('executiveSummaryCompilation.reportCountPlaceholder')"
                class="w-full font-bold"
                disabled
                readonly
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.recommendationCount')" :hint="t('executiveSummaryCompilation.sources.hintRecommendation')">
              <UInput
                type="number"
                :model-value="store.form.jumlahRekomendasi"
                :placeholder="t('executiveSummaryCompilation.recommendationCountPlaceholder')"
                class="w-full font-bold"
                disabled
                readonly
              />
            </UFormField>

            <!-- Auto calculated fields -->
            <UFormField :label="t('executiveSummaryCompilation.totalFindings')" :hint="t('executiveSummaryCompilation.totalFindingsHint')">
              <UInput
                :model-value="totalTemuanSummary"
                class="w-full font-bold bg-gray-50 dark:bg-gray-800"
                disabled
                readonly
              >
                <template #trailing>
                  <span class="text-md text-gray-400">{{ t('executiveSummaryCompilation.findingsUnit') }}</span>
                </template>
              </UInput>
            </UFormField>
          </div>

          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4">
            <div class="flex justify-between items-center">
              <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryCompilation.riskBreakdown') }}</h4>
              <span class="text-xs text-gray-400">{{ t('executiveSummaryCompilation.sources.hintRisk') }}</span>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <UFormField :label="t('executiveSummaryCompilation.riskHigh')">
                <UInput
                  type="number"
                  :model-value="store.form.risikoTinggi"
                  class="w-full border-l-4 border-error-500 font-bold"
                  disabled
                  readonly
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryCompilation.riskMedium')">
                <UInput
                  type="number"
                  :model-value="store.form.risikoSedang"
                  class="w-full border-l-4 border-warning-500 font-bold"
                  disabled
                  readonly
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryCompilation.riskLow')">
                <UInput
                  type="number"
                  :model-value="store.form.risikoRendah"
                  class="w-full border-l-4 border-success-500 font-bold"
                  disabled
                  readonly
                />
              </UFormField>
            </div>
          </div>
        </section>

        <!-- 4. Section Status Tindak Lanjut -->
        <section id="sec-followup" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.followUp.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.followUp.description') }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ t('executiveSummaryCompilation.sources.hintFollowUp') }}</p>
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
                      :model-value="row.jumlah"
                      placeholder="0"
                      class="font-semibold"
                      disabled
                      readonly
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
                  {{ t('executiveSummaryForm.topFindings.title') }}
                </h2>
                <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.topFindings.description') }}</p>
              </div>
              <div class="flex items-center gap-2">
                <UButton
                  v-if="!isLocked"
                  color="neutral"
                  variant="outline"
                  icon="i-lucide-refresh-cw"
                  :label="t('executiveSummaryForm.sync.findings')"
                  size="sm"
                  @click="syncFromIndividualLha"
                />
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
                      @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.findingTitle'))"
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
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryForm.analysis.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.analysis.description') }}</p>
            </div>
            <UButton
              v-if="!isLocked"
              color="neutral"
              variant="outline"
              icon="i-lucide-refresh-cw"
              :label="t('executiveSummaryForm.sync.analysis')"
              size="sm"
              @click="syncFromIndividualLha"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField :label="t('executiveSummaryForm.analysis.theme')">
              <UTextarea
                v-model="store.form.akarMasalah"
                :placeholder="t('executiveSummaryForm.analysis.rootCausePlaceholder')"
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryForm.analysis.conclusion')">
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
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signaturePlace'))"
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
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signatureName'))"
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
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signatureNik'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNIK ? store.form.signatureNIK.length : 0 }}/50
                </div>
              </UFormField>
            </div>
          </div>
        </section>

        <!-- 7. Section Matriks Kompilasi LHA Individual -->
        <section id="sec-matriks" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryCompilation.matrix.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryCompilation.matrix.description') }}</p>
            </div>
            <UBadge color="primary" variant="subtle" size="md">
              {{ t('executiveSummaryCompilation.matrix.total', { count: store.form.matriksKompilasi.length || store.form.jumlahLaporan }) }}
            </UBadge>
          </div>

          <div v-if="store.form.matriksKompilasi.length > 0" class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800/80">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.docNumber') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.unit') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.title') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.risk') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.recommendation') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.followUpStatus') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800 text-sm">
                <tr v-for="(row, idx) in store.form.matriksKompilasi" :key="idx">
                  <td class="px-4 py-3 font-mono font-semibold text-primary-600 dark:text-primary-400">{{ row.nomor }}</td>
                  <td class="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{{ row.division || 'OP' }}</td>
                  <td class="px-4 py-3 text-gray-700 dark:text-gray-300 max-w-xs truncate" :title="row.judulTemuan">{{ row.judulTemuan }}</td>
                  <td class="px-4 py-3">
                    <UBadge :color="row.nilaiRisiko === 'Tinggi' ? 'error' : row.nilaiRisiko === 'Sedang' ? 'warning' : 'success'" variant="subtle" size="xs">
                      {{ row.nilaiRisiko }}
                    </UBadge>
                  </td>
                  <td class="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate" :title="row.rekomendasi">{{ row.rekomendasi }}</td>
                  <td class="px-4 py-3">
                    <span :class="getStatusBadgeClass(row.status || 'In Progress')">{{ row.status || 'In Progress' }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="text-center py-8 bg-gray-50 dark:bg-gray-800/30 rounded-lg text-sm text-gray-400 border border-dashed">
            {{ t('executiveSummaryCompilation.matrix.empty') }}
            <button v-if="!isLocked" type="button" @click="syncFromIndividualLha" class="text-primary-500 font-bold ml-1 hover:underline">
              {{ t('executiveSummaryCompilation.matrix.emptyAction') }}
            </button>
          </div>
        </section>

        <!-- Catatan Executive untuk Auditor -->
        <section v-if="store.isViewing" id="sec-notes" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-message-square-text" class="size-5 text-primary-500" />
              {{ t('executiveSummaryForm.notes.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.notes.description') }}</p>
          </div>

          <UFormField :label="t('executiveSummaryForm.notes.label')">
            <UTextarea
              v-model="store.form.executiveNote"
              :placeholder="canWriteExecutiveNote ? t('executiveSummaryForm.notes.placeholder') : (store.form.executiveNote ? '' : t('executiveSummaryForm.notes.empty'))"
              :rows="5"
              class="w-full"
              :disabled="!canWriteExecutiveNote"
            />
          </UFormField>
          <div v-if="canWriteExecutiveNote" class="flex justify-end">
            <UButton
              color="primary"
              icon="i-lucide-save"
              :label="t('executiveSummaryForm.notes.save')"
              :loading="store.loading"
              @click="store.saveExecutiveNote"
            />
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
        </template>
        
        <!-- Locked/Unlocked Overrides -->
        <template v-else>
          <UButton
            v-if="isDraft && canApprove"
            color="success"
            variant="solid"
            icon="i-lucide-check"
            :label="t('executiveSummaryForm.actions.approve')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="approveReportDirectly"
          />
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useI18n } from '~/composables/useI18n'
import { useExecutiveSummarySources, type CompilationSources } from '~/composables/useExecutiveSummarySources'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
const store = useExecutiveSummaryStore()
const authStore = useAuthStore()
const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()
const { canReviewExecutiveSummary } = useRbac()
const { t } = useI18n()
const { loading: sourcesLoading, error: sourcesError, fetchCompilationSources } = useExecutiveSummarySources()
const fileInput = ref<HTMLInputElement | null>(null)
const activeSection = ref('sec-upload')
const scrollContainer = ref<HTMLElement | null>(null)

// Last successful read of the statistics sources (null until the first sync)
const lastSources = ref<CompilationSources | null>(null)

// Empty-source hints shown under the statistics section after a sync
const sourceEmptyNotes = computed<string[]>(() => {
  const sources = lastSources.value
  if (!sources) return []
  const notes: string[] = []
  if (sources.recommendationCount === 0) notes.push(t('executiveSummaryCompilation.sources.emptyRecommendation'))
  if (sources.riskBreakdown.high + sources.riskBreakdown.medium + sources.riskBreakdown.low === 0) {
    notes.push(t('executiveSummaryCompilation.sources.emptyRisk'))
  }
  if (sources.followUp.closed + sources.followUp.inProgress + sources.followUp.overdue === 0) {
    notes.push(t('executiveSummaryCompilation.sources.emptyFollowUp'))
  }
  return notes
})

if (!auditReportStore.loading && auditReportStore.reportList.length === 0) {
  auditReportStore.fetchReports()
}
if (!assignmentLetterStore.loading && assignmentLetterStore.assignmentLetterList.length === 0) {
  assignmentLetterStore.fetchAssignmentLetters()
}

// Navigation links
const sections = computed(() => [
  { id: 'sec-upload', index: '1', title: t('executiveSummaryCompilation.nav.upload') },
  { id: 'sec-narrative', index: '2', title: t('executiveSummaryCompilation.nav.narrative') },
  { id: 'sec-stats', index: '3', title: t('executiveSummaryCompilation.nav.stats') },
  { id: 'sec-followup', index: '4', title: t('executiveSummaryCompilation.nav.followUp') },
  { id: 'sec-topfindings', index: '5', title: t('executiveSummaryCompilation.nav.topFindings') },
  { id: 'sec-analysis', index: '6', title: t('executiveSummaryCompilation.nav.analysis') },
  { id: 'sec-matriks', index: '7', title: t('executiveSummaryCompilation.nav.matrix') },
  ...(store.isViewing ? [{ id: 'sec-notes', index: '8', title: t('executiveSummaryCompilation.nav.notes') }] : [])
])

// Dropdown options
const divisionOptions = ['OP', 'KK/KSD', 'IT', 'FIN', 'HR', 'LEG']

const quarterOptions = ['Kuartal I', 'Kuartal II', 'Kuartal III', 'Kuartal IV']
const reportingQuarter = computed({
  get: (): string => quarterOptions[store.form.quarter - 1] ?? 'Kuartal I',
  set: (value: string) => {
    const quarterIndex = quarterOptions.indexOf(value)
    store.form.quarter = quarterIndex >= 0 ? quarterIndex + 1 : 1
    store.form.periodeBulan = quarterOptions[store.form.quarter - 1] ?? 'Kuartal I'
  }
})

// Role computed checks: Admin, CAE, and Audit Manager can write note & approve draft
const canWriteExecutiveNote = computed(() => Boolean(canReviewExecutiveSummary.value))
const canApprove = computed(() => Boolean(canReviewExecutiveSummary.value))
const isChiefAuditExecutive = computed(() => Boolean(canReviewExecutiveSummary.value))
const isDraft = computed(() => String(store.form.status || '').toLowerCase() === 'draft')

const isLocked = computed(() => {
  return store.form.status === 'Approved'
})

// Helper to generate compilation document number: ESC-[nomor urut]/[tahun]
const generateEscDocNumber = (targetYear: number = 2026): string => {
  if (typeof (store as any).generateEscDocNumber === 'function') {
    return (store as any).generateEscDocNumber(targetYear)
  }
  let maxSeq = 0
  const escRegex = new RegExp(`^ESC-(\\d+)/${targetYear}$`, 'i')
  for (const s of store.summaryList || []) {
    if (s.nomorDokumen) {
      const match = s.nomorDokumen.match(escRegex)
      if (match && match[1]) {
        const val = parseInt(match[1], 10)
        if (!isNaN(val) && val > maxSeq) maxSeq = val
      }
    }
  }
  const nextSeq = (maxSeq + 1).toString().padStart(3, '0')
  return `ESC-${nextSeq}/${targetYear}`
}

// Helper to determine year and quarter of an individual LHA
const getReportQuarterAndYear = (r: any, letters: any[] = []): { year: number; quarter: number } => {
  let year = 0
  let quarter = 0
  const dateStr = r.reportDate || (r as any).report_date || ''
  if (dateStr) {
    const parts = dateStr.split('T')[0].split('-')
    if (parts[0] && /^\d{4}$/.test(parts[0])) {
      year = parseInt(parts[0], 10)
    }
    if (parts[1]) {
      const month = parseInt(parts[1], 10)
      if (month >= 1 && month <= 12) {
        quarter = Math.ceil(month / 3)
      }
    }
  }

  // Linked Assignment Letter info
  if (r.assignmentLetterId) {
    const st = letters.find(s => s.letterNumber === r.assignmentLetterId)
    if (st) {
      if (!year && st.auditYear) {
        const yMatch = String(st.auditYear).match(/\b(20\d{2})\b/)
        if (yMatch) year = parseInt(yMatch[1] ?? '0', 10)
      }
      if (!quarter && st.startPeriod) {
        const parts = st.startPeriod.split('T')[0].split('-')
        if (parts[1]) {
          const m = parseInt(parts[1], 10)
          if (m >= 1 && m <= 12) quarter = Math.ceil(m / 3)
        }
      }
      if (!quarter && st.executionPeriod) {
        if (/Q1|Kuartal I\b|Triwulan I\b/i.test(st.executionPeriod)) quarter = 1
        else if (/Q2|Kuartal II\b|Triwulan II\b/i.test(st.executionPeriod)) quarter = 2
        else if (/Q3|Kuartal III\b|Triwulan III\b/i.test(st.executionPeriod)) quarter = 3
        else if (/Q4|Kuartal IV\b|Triwulan IV\b/i.test(st.executionPeriod)) quarter = 4
      }
    }
  }

  if (!year) {
    const numMatch = (r.reportNumber || (r as any).report_number || '').match(/\b(20\d{2})\b/)
    if (numMatch) {
      year = parseInt(numMatch[1], 10)
    } else {
      year = 2026
    }
  }
  if (!quarter) {
    quarter = 1
  }
  return { year, quarter }
}

// All LHAs corresponding to the selected Tahun Laporan and Periode Kuartal
const quarterlyLhas = computed(() => {
  const targetYear = Number(store.form.tahun) || 2026
  const targetQuarter = Number(store.form.quarter) || 1
  const validReports = auditReportStore.reportList.filter(r => {
    const num = r.reportNumber || (r as any).report_number
    return (
      num &&
      num !== 'DOC-EXSUM-Q1-2026' &&
      num !== '020/LHA/01/KS IAD/2023' &&
      num !== '019/LHA/01/KS IAD/2025' &&
      num !== 'LHA-020/SKAI/2023' &&
      num !== 'LHA-019/SKAI/2025'
    )
  })
  return validReports.filter(r => {
    const { year, quarter } = getReportQuarterAndYear(r, assignmentLetterStore.assignmentLetterList)
    return year === targetYear && quarter === targetQuarter
  })
})

// Section II computed auto-sum
const totalTemuanSummary = computed(() => {
  return (store.form.risikoTinggi || 0) + (store.form.risikoSedang || 0) + (store.form.risikoRendah || 0)
})

// Section III computed auto-sums
const totalFollowUpCount = computed(() => {
  return store.form.followUpTable.reduce((acc, row) => acc + (row.jumlah || 0), 0)
})

// Helper formatting percentage
const formatPercent = (val: number) => {
  return val ? val.toFixed(1) : '0.0'
}

const resetNarrativeToDefault = () => {
  store.form.narrative = store.defaultNarrativeTemplate(reportingQuarter.value, store.form.tahun)
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

const getStatusBadgeClass = (status: 'Closed' | 'In Progress' | 'Overdue') => {
  const colors = {
    Closed: 'bg-success-100 text-success-800 dark:bg-success-950 dark:text-success-300 px-2.5 py-0.5 rounded text-md font-bold uppercase',
    'In Progress': 'bg-info-100 text-info-800 dark:bg-info-950 dark:text-info-300 px-2.5 py-0.5 rounded text-md font-bold uppercase',
    Overdue: 'bg-error-100 text-error-800 dark:bg-error-950 dark:text-error-300 px-2.5 py-0.5 rounded text-md font-bold uppercase'
  }
  return colors[status]
}

// Automatic Compilation Synchronization from all individual LHAs in the selected quarter and year.
// Statistics (recommendations, risk levels, follow-up) are read-only and computed from the source data.
const syncFromIndividualLha = async () => {
  const targetYear = store.form.tahun

  // Auto-set nomorDokumen if empty or follows legacy template
  if (!store.isEditing && !store.isViewing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(Number(targetYear) || 2026)
    }
  }

  const individualReports = quarterlyLhas.value

  // 1. Set jumlahLaporan to exact count of individual LHAs
  store.form.jumlahLaporan = individualReports.length

  if (individualReports.length === 0) {
    lastSources.value = null
    store.form.risikoTinggi = 0
    store.form.risikoSedang = 0
    store.form.risikoRendah = 0
    store.form.jumlahRekomendasi = 0
    store.form.topFindings = []
    store.form.matriksKompilasi = []
    store.form.followUpTable = [
      { status: 'Closed', jumlah: 0, persentase: 0, keterangan: 'Telah diverifikasi' },
      { status: 'In Progress', jumlah: 0, persentase: 0, keterangan: 'On-track pelaksanaan' },
      { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '-' }
    ]
    store.form.akarMasalah = `Belum ada data LHA untuk ${reportingQuarter.value} ${targetYear}.`
    store.form.kesimpulan = `Belum ada hasil audit yang dikompilasi untuk ${reportingQuarter.value} ${targetYear}.`
    store.form.narrative = `Periode ${reportingQuarter.value} ${targetYear}, belum ada LHA yang diterbitkan.`
    return
  }

  // 2. Read-only sources: test-control recommendations, working paper risks, ATR follow-up status
  const sources = await fetchCompilationSources({
    assignmentLetterIds: individualReports.map(r => r.assignmentLetterId),
    lhaIds: individualReports.map(r => r.id)
  })
  if (!sources) return
  lastSources.value = sources
  const totalHigh = sources.riskBreakdown.high
  const totalMedium = sources.riskBreakdown.medium
  const totalLow = sources.riskBreakdown.low
  const totalRecs = sources.recommendationCount
  store.form.risikoTinggi = totalHigh
  store.form.risikoSedang = totalMedium
  store.form.risikoRendah = totalLow
  store.form.jumlahRekomendasi = totalRecs

  // 3. Aggregate LHA findings for the Top 5 list and the compilation matrix
  const allFindings: any[] = []
  const matriks: any[] = []

  individualReports.forEach(r => {
    const findings = r.findings || []
    const rNum = r.reportNumber || (r as any).report_number
    let div = r.department || 'OP'
    if (r.reportTitle.includes('Keuangan')) div = 'FIN'
    else if (r.reportTitle.includes('Sistem') || r.reportTitle.includes('TI') || r.reportTitle.includes('ERP')) div = 'IT'
    else if (r.reportTitle.includes('SDM') || r.reportTitle.includes('Payroll')) div = 'HR'
    else if (r.reportTitle.includes('Procurement') || r.reportTitle.includes('SCM')) div = 'KK/KSD'

    findings.forEach((f: any) => {
      let risiko: 'Tinggi' | 'Sedang' | 'Rendah' = 'Tinggi'
      if (f.category === 'Very Significant' || f.category === 'Significant') {
        risiko = 'Tinggi'
      } else if (f.category === 'Quite Significant') {
        risiko = 'Sedang'
      } else {
        risiko = 'Rendah'
      }

      allFindings.push({
        unitDivision: div,
        judulTemuan: f.title || f.finding || '',
        risiko,
        statusTL: 'In Progress',
        usulan: f.action || f.recommendation || 'Perbaikan SOP dan Kontrol Pengendalian Internal',
        severityRank: risiko === 'Tinggi' ? 3 : risiko === 'Sedang' ? 2 : 1
      })
    })

    const highestRisk = findings.some((f: any) => ['Very Significant', 'Significant'].includes(f.category)) ? 'Tinggi' : 'Sedang'
    matriks.push({
      nomor: rNum,
      division: div,
      unitKerja: r.department || r.reportTitle,
      prosesBisnis: r.reportTitle,
      judulTemuan: (findings[0]?.title) || r.reportTitle,
      nilaiRisiko: highestRisk,
      rekomendasi: (findings[0]?.action) || 'Tindak lanjut rekomendasi audit',
      dueDate: r.reportDate || `${targetYear}-06-30`,
      picUnit: `PIC ${div}`,
      progres: r.status === 'Final' || (r.status as any) === 'Approved' ? 100 : 60,
      status: r.status === 'Final' || (r.status as any) === 'Approved' ? 'Closed' : 'In Progress',
      buktiTL: `LHA_${rNum.replace(/[\/\s]/g, '_')}.pdf`
    })
  })

  store.form.matriksKompilasi = matriks

  // 4. Sort findings by severity (High first) and pick Top 5
  allFindings.sort((a, b) => b.severityRank - a.severityRank)
  store.form.topFindings = allFindings.slice(0, 5).map(f => ({
    unitDivision: f.unitDivision,
    judulTemuan: f.judulTemuan,
    risiko: f.risiko,
    statusTL: f.statusTL,
    usulan: f.usulan
  }))

  // 5. Follow-up table from Action Taken Report statuses (read-only)
  const { closed: closedCount, inProgress: inProgCount, overdue: overdueCount } = sources.followUp
  const totalTl = closedCount + inProgCount + overdueCount
  const percentOfTl = (n: number) => (totalTl > 0 ? (n / totalTl) * 100 : 0)
  store.form.followUpTable = [
    { status: 'Closed', jumlah: closedCount, persentase: percentOfTl(closedCount), keterangan: 'Telah diverifikasi' },
    { status: 'In Progress', jumlah: inProgCount, persentase: percentOfTl(inProgCount), keterangan: 'On-track pelaksanaan' },
    { status: 'Overdue', jumlah: overdueCount, persentase: percentOfTl(overdueCount), keterangan: '-' }
  ]
  const totalFindings = totalHigh + totalMedium + totalLow

  // 6. Section V & VII Qualitative text
  if (store.form.topFindings.length > 0) {
    store.form.akarMasalah = `Analisis kompilasi terhadap ${individualReports.length} LHA Individual menunjukkan tema berulang utama pada: ${store.form.topFindings.slice(0, 2).map(f => f.judulTemuan).join('; ')}. Diperlukan penguatan regulasi operasional serta koordinasi lintas unit.`
    store.form.kesimpulan = `Secara keseluruhan hasil audit kompilasi ${reportingQuarter.value} ${targetYear} atas ${individualReports.length} LHA menunjukkan sistem pengendalian internal berjalan cukup efektif dengan total ${totalFindings} temuan (${totalHigh} risiko tinggi, ${totalMedium} sedang, ${totalLow} rendah) dan ${totalRecs} rekomendasi yang perlu dipantau implementasinya.`
  } else {
    store.form.akarMasalah = `Analisis kompilasi terhadap ${individualReports.length} LHA Individual periode ${reportingQuarter.value} ${targetYear}.`
    store.form.kesimpulan = `Secara keseluruhan hasil audit kompilasi ${reportingQuarter.value} ${targetYear} atas ${individualReports.length} LHA menunjukkan tata kelola dan kontrol internal telah dievaluasi.`
  }

  // 7. Update narrative
  store.form.narrative = `Periode ${reportingQuarter.value} ${targetYear}, SPI telah menerbitkan ${individualReports.length} LHA Individual. Total temuan ${totalFindings} dengan ${totalRecs} rekomendasi perbaikan. Tingkat penyelesaian on-time 95%. Terdapat ${totalHigh} temuan risiko tinggi terkait ${store.form.topFindings[0]?.judulTemuan || 'operasional'} yang memerlukan perhatian dan arahan Direksi.`
}

// Form submission & workflow helpers
const saveReportDraft = async () => {
  if (!store.form.nomorDokumen) {
    store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
  }
  if (!store.form.dokumenPath) {
    alert(t('executiveSummaryForm.messages.documentRequired'))
    return
  }

  // Set status Draft
  store.form.periodeBulan = reportingQuarter.value
  store.form.status = 'Draft'
  await store.saveForm()
}

const approveReportDirectly = async () => {
  if (store.currentSummary) {
    await store.updateStatus(store.currentSummary.id, 'Approved')
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

// Watch quarter & tahun changes to automatically sync data
watch(() => [store.form.quarter, store.form.tahun], ([newQ, newY], [oldQ, oldY]) => {
  if (!store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || (oldY && store.form.nomorDokumen.endsWith(`/${oldY}`) && newY !== oldY)) {
      store.form.nomorDokumen = generateEscDocNumber(Number(newY) || 2026)
    }
    syncFromIndividualLha()
  }
})

// Watch modal open
watch(() => store.showModal, (isOpen) => {
  if (isOpen && !store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
    }
  }
})

// Watch reportList when reports are fetched asynchronously
watch(() => auditReportStore.reportList.length, (newLen) => {
  if (newLen > 0 && !store.isViewing && !store.isEditing) {
    if (store.form.jumlahLaporan === 0 || store.form.matriksKompilasi.length === 0) {
      syncFromIndividualLha()
    }
  }
})

onMounted(() => {
  if (!store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
    }
    syncFromIndividualLha()
  }
})
</script>
