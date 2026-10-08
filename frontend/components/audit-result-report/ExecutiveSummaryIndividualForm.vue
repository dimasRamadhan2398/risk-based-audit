<template>
  <div class="grid grid-cols-1 lg:grid-cols-4 h-full bg-gray-50 dark:bg-gray-950">
    <!-- Sidebar Navigation Index (Corporate Report Builder Layout) -->
    <div class="hidden lg:block lg:col-span-1 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-6 overflow-y-auto space-y-6">
      <div class="text-md font-bold uppercase tracking-wider text-gray-400">Daftar Bagian Laporan</div>
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
            Status Penguncian
          </div>
          <div class="text-sm font-bold text-gray-800 dark:text-white flex items-center gap-1.5">
            <span class="size-2 rounded-full" :class="store.form.status === 'Approved' ? 'bg-success-500' : 'bg-warning-500'"></span>
            {{ store.form.status === 'Approved' ? 'Terkunci (Approved)' : 'Terbuka (Draft)' }}
          </div>
          <p class="text-md text-gray-400 leading-normal">
            {{ store.form.status === 'Approved' 
              ? 'Laporan ini telah disetujui oleh Kepala SPI dan tidak dapat diedit.' 
              : 'Silakan isi semua data sebelum mengajukan persetujuan.' }}
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
              <span class="text-primary-500">I.</span> Navigation & Document Upload
            </h2>
            <p class="text-sm text-gray-400">Detail identitas laporan kompilasi dan unggahan dokumen resmi.</p>
          </div>

          <div class="w-full">
            <UFormField label="Pilih Surat Tugas (Assignment Letter) / LHA">
              <USelectMenu
                v-model="selectedLhaId"
                :items="lhaDropdownOptions"
                placeholder="Pilih ID LHA..."
                class="w-full font-semibold"
                :disabled="isLocked"
                @update:modelValue="onLhaSelect"
              />
            </UFormField>
          </div>

          <!-- Assignment Letter Linked Context Card -->
          <div v-if="linkedAssignmentLetter" class="bg-primary-50/60 dark:bg-primary-950/30 border border-primary-200 dark:border-primary-800/60 rounded-xl p-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div class="flex items-center gap-3">
              <div class="p-2.5 bg-primary-100 dark:bg-primary-900/60 text-primary-600 dark:text-primary-400 rounded-lg">
                <UIcon name="i-lucide-file-signature" class="size-5" />
              </div>
              <div>
                <div class="text-xs font-bold uppercase tracking-wider text-primary-700 dark:text-primary-300 flex items-center gap-2">
                  <span>Surat Tugas: {{ linkedAssignmentLetter.letterNumber }}</span>
                  <UBadge color="primary" variant="subtle" size="xs">{{ linkedAssignmentLetter.status || 'Published' }}</UBadge>
                </div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white mt-0.5">
                  {{ linkedAssignmentLetter.auditTitle }}
                </div>
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <div v-if="linkedAssignmentLetter.leader">
                <span class="text-gray-400">Ketua Tim:</span> <span class="font-medium text-gray-700 dark:text-gray-300">{{ linkedAssignmentLetter.leader }}</span>
              </div>
              <div v-if="linkedAssignmentLetter.executionPeriod || (linkedAssignmentLetter as any).auditYear">
                <span class="text-gray-400">Periode:</span> <span class="font-medium text-gray-700 dark:text-gray-300">{{ linkedAssignmentLetter.executionPeriod || (linkedAssignmentLetter as any).auditYear }}</span>
              </div>
            </div>
          </div>

          <!-- Document Upload Field -->
          <div class="bg-gray-50 dark:bg-gray-800/50 p-6 rounded-xl border border-dashed border-gray-300 dark:border-gray-700">
            <UFormField label="Upload Dokumen Executive Summary Resmi" required>
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
                    label="Pilih File (.pdf, .docx)"
                    :disabled="isLocked"
                    @click="fileInput?.click()"
                  />
                  <span v-if="store.form.dokumenPath" class="text-sm text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                    <UIcon name="i-lucide-file-check" class="text-success-500 size-5" />
                    {{ store.form.dokumenPath }}
                  </span>
                  <span v-else class="text-sm text-gray-400">Belum ada file yang dipilih (Maks. 10MB)</span>
                </div>
                <p class="text-md text-gray-400">
                  Dukungan format file PDF dan Word yang valid, tidak terkunci password.
                </p>
              </div>
            </UFormField>
          </div>
        </section>

        <!-- 2. Section Narrative -->
        <section id="sec-narrative" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">II.</span> Section I: Executive Summary Narrative
            </h2>
            <p class="text-sm text-gray-400">Narasi bebas untuk ringkasan eksekutif hasil audit triwulanan.</p>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-md font-bold text-gray-400 uppercase">Teks Narasi Ringkasan</span>
              <UButton
                v-if="!isLocked"
                color="neutral"
                variant="ghost"
                icon="i-lucide-refresh-cw"
                label="Reset Ke Template Default"
                size="md"
                @click="resetNarrativeToDefault"
              />
            </div>
            <UTextarea
              v-model="store.form.narrative"
              placeholder="Isi teks ringkasan naratif..."
              :rows="6"
              class="w-full"
              :disabled="isLocked"
            />
          </div>
        </section>

        <!-- 4. Section Status Tindak Lanjut -->
        <section id="sec-followup" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-primary-500">III.</span> Section III: Status Tindak Lanjut
            </h2>
            <p class="text-sm text-gray-400">Kalkulasi persentase dan rekap status penyelesaian temuan audit.</p>
          </div>

          <div class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800/80">
                <tr>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">Status</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">Jumlah</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">% (Persentase)</th>
                  <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">Keterangan</th>
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
                      placeholder="Keterangan tindak lanjut..."
                      class="w-full"
                      :disabled="isLocked"
                    />
                  </td>
                </tr>
                <!-- Total Row -->
                <tr class="bg-gray-50 dark:bg-gray-800/40 font-bold border-t-2">
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white uppercase tracking-wider">Total</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white font-mono font-extrabold">{{ totalFollowUpCount }}</td>
                  <td class="px-6 py-4 text-sm text-gray-900 dark:text-white font-mono font-extrabold">100.0%</td>
                  <td class="px-6 py-4 text-md text-gray-400 italic">Dihitung otomatis</td>
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
                  <span class="text-primary-500">IV.</span> Section IV: Top 5 Temuan Signifikan
                </h2>
                <p class="text-sm text-gray-400">Matriks temuan kritikal terpenting yang butuh eskalasi/tindakan jajaran direksi.</p>
              </div>
              <div class="flex items-center gap-2">
                <UButton
                  v-if="!isLocked"
                  color="neutral"
                  variant="outline"
                  icon="i-lucide-refresh-cw"
                  label="Sinkronkan Temuan dari LHA"
                  size="sm"
                  @click="syncFindingsFromLha()"
                />
                <UButton
                  v-if="!isLocked && store.form.topFindings.length < 5"
                  color="primary"
                  variant="soft"
                  icon="i-lucide-plus-circle"
                  label="Tambah Temuan"
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
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">Unit Divisi</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">Judul Temuan</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-32">Nilai Risiko</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">Status TL</th>
                  <th class="px-4 py-3 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">Usulan Tindakan</th>
                  <th v-if="!isLocked" class="px-4 py-3 text-center text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-16">Aksi</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800">
                <tr v-for="(finding, idx) in store.form.topFindings" :key="idx">
                  <td class="px-3 py-2">
                    <USelectMenu
                      v-model="finding.unitDivision"
                      :items="divisionOptions"
                      placeholder="Divisi"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <UInput
                      v-model="finding.judulTemuan"
                      placeholder="Judul temuan utama..."
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
                      placeholder="Risiko"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <USelectMenu
                      v-model="finding.statusTL"
                      :items="['Closed', 'In Progress', 'Overdue']"
                      placeholder="Status TL"
                      :disabled="isLocked"
                    />
                  </td>
                  <td class="px-3 py-2">
                    <UTextarea
                      v-model="finding.usulan"
                      placeholder="Contoh: Eskalasi Direksi"
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
            Belum ada temuan signifikan teratas yang ditambahkan.
            <button v-if="!isLocked" type="button" @click="addTopFinding" class="text-primary-500 font-bold ml-1 hover:underline">
              Klik di sini untuk menambahkan.
            </button>
          </div>
        </section>

        <!-- 5. Recurring findings analysis & conclusion (report Sections V & VII) -->
        <section id="sec-analysis" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                <span class="text-primary-500">V.</span> Analisis Temuan Berulang & Kesimpulan
              </h2>
              <p class="text-sm text-gray-400">Deskripsi tema berulang, akar masalah, dan usulan task force atau arah kebijakan manajemen.</p>
            </div>
            <UButton
              v-if="!isLocked"
              color="neutral"
              variant="outline"
              icon="i-lucide-refresh-cw"
              label="Sinkronkan Analisis dari LHA"
              size="sm"
              @click="syncFindingsFromLha()"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField label="Section V: Tema Berulang (Akar masalah & Usulan revisi kebijakan)">
              <UTextarea
                v-model="store.form.akarMasalah"
                placeholder="Tuliskan analisis akar masalah dan usulan kebijakan..."
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField label="Section VII: Kesimpulan & Rekomendasi Manajemen">
              <UTextarea
                v-model="store.form.kesimpulan"
                placeholder="Tuliskan kesimpulan umum dan arahan Direksi yang diperlukan..."
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>
          </div>

          <!-- Electronic Signatures -->
          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4">
            <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">Tanda Tangan Elektronik SPI</h4>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
              <UFormField label="Tempat TTD" required>
                <UInput
                  v-model="store.form.signatureTempat"
                  placeholder="Contoh: Jakarta"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity('Tempat TTD maksimal 20 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureTempat ? store.form.signatureTempat.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField label="Tanggal TTD" required>
                <AppDatePicker
                  v-model="store.form.signatureTanggal"
                  :disabled="isLocked"
                />
              </UFormField>

              <UFormField label="Nama Kepala SPI" required>
                <UInput
                  v-model="store.form.signatureNamaKepala"
                  placeholder="Contoh: Budi Santoso, CIA"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity('Nama Kepala SPI maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNamaKepala ? store.form.signatureNamaKepala.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField label="NIK Kepala SPI" required>
                <UInput
                  v-model="store.form.signatureNIK"
                  placeholder="Contoh: SPI-77621"
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

        <!-- Catatan Executive untuk Auditor -->
        <section v-if="store.isViewing" id="sec-notes" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-message-square-text" class="size-5 text-primary-500" />
              Noted dari Executive untuk Auditor
            </h2>
            <p class="text-sm text-gray-400">Catatan tindak lanjut atau arahan Executive kepada Auditor.</p>
          </div>

          <UFormField label="Noted">
            <UTextarea
              v-model="store.form.executiveNote"
              :placeholder="canWriteExecutiveNote ? 'Tuliskan catatan untuk Auditor...' : (store.form.executiveNote ? '' : 'Belum ada catatan dari Executive.')"
              :rows="5"
              class="w-full"
              :disabled="!canWriteExecutiveNote"
            />
          </UFormField>
          <div v-if="canWriteExecutiveNote" class="flex justify-end">
            <UButton
              color="primary"
              icon="i-lucide-save"
              label="Simpan Catatan"
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
          label="Tutup"
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
            label="Simpan Draft"
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
            label="Setujui (Approve)"
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
import { ref, computed } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { UserRole } from '~/types/auth'
const store = useExecutiveSummaryStore()
const authStore = useAuthStore()
const { canReviewExecutiveSummary } = useRbac()

const fileInput = ref<HTMLInputElement | null>(null)
const activeSection = ref('sec-upload')
const scrollContainer = ref<HTMLElement | null>(null)

// Navigation links
const sections = computed(() => [
  { id: 'sec-upload', index: '1', title: 'Navigation & Upload' },
  { id: 'sec-narrative', index: '2', title: 'Narrative Summary' },
  { id: 'sec-followup', index: '3', title: 'Status Tindak Lanjut' },
  { id: 'sec-topfindings', index: '4', title: 'Temuan Signifikan' },
  { id: 'sec-analysis', index: '5', title: 'Temuan Berulang & Kesimpulan' },
  ...(store.isViewing ? [{ id: 'sec-notes', index: '6', title: 'Noted' }] : [])
])

// Dropdown options
const divisionOptions = ['OP', 'KK/KSD', 'IT', 'FIN', 'HR', 'LEG']

// Role computed checks: Admin, CAE, and Audit Manager can write note & approve draft
const canWriteExecutiveNote = computed(() => Boolean(canReviewExecutiveSummary.value))
const canApprove = computed(() => Boolean(canReviewExecutiveSummary.value))
const isChiefAuditExecutive = computed(() => Boolean(canReviewExecutiveSummary.value))
const isDraft = computed(() => String(store.form.status || '').toLowerCase() === 'draft')

const isLocked = computed(() => {
  return store.form.status === 'Approved'
})

// Section III computed auto-sums
const totalFollowUpCount = computed(() => {
  return store.form.followUpTable.reduce((acc, row) => acc + (row.jumlah || 0), 0)
})

// Force recalculate Section III percentages
const recalculatePercentages = () => {
  const total = totalFollowUpCount.value
  store.form.followUpTable.forEach(row => {
    row.persentase = total > 0 ? (row.jumlah / total) * 100 : 0
  })
}

// Helper formatting percentage
const formatPercent = (val: number) => {
  return val ? val.toFixed(1) : '0.0'
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
    alert('File size exceeds the 10MB limit.')
    return
  }

  // Type limit
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (ext !== 'pdf' && ext !== 'docx' && ext !== 'doc') {
    alert('Format file tidak didukung. Harap unggah PDF/Word.')
    return
  }

  // Simulated file corruption error handling
  if (file.name.toLowerCase().includes('corrupt') || file.name.toLowerCase().includes('damaged')) {
    alert('Gagal mengunggah file. Pastikan format file adalah PDF/Word yang valid dan tidak terkunci password.')
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

// Form submission & workflow helpers
const saveReportDraft = async () => {
  if (!store.form.nomorDokumen) {
    alert('Nomor dokumen wajib diisi.')
    return
  }
  if (!store.form.dokumenPath) {
    alert('Unggah dokumen resmi wajib diisi.')
    return
  }
  
  // Set status Draft
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

// Assignment Letter & LHA Stores Integration for 2-Way Sync
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()

if (!auditReportStore.loading && auditReportStore.reportList.length === 0) {
  auditReportStore.fetchReports()
}

const assignmentLetterDropdownOptions = computed(() => {
  const lettersFromStore = assignmentLetterStore.assignmentLetterList
    .filter((st: any) => st.letterNumber && st.letterNumber.startsWith('ST-') && st.letterNumber.includes('2026'))
    .map((st: any) => ({
      label: `${st.letterNumber} - ${st.auditTitle || ''}`,
      value: st.letterNumber,
      letter: st
    }))
  const uniqueNumbers = new Set(lettersFromStore.map(l => l.value))
  
  auditReportStore.reportList.forEach(r => {
    if (r.assignmentLetterId && r.assignmentLetterId.startsWith('ST-') && r.assignmentLetterId.includes('2026') && !uniqueNumbers.has(r.assignmentLetterId)) {
      uniqueNumbers.add(r.assignmentLetterId)
      lettersFromStore.push({
        label: `${r.assignmentLetterId} - ${r.reportTitle || ''}`,
        value: r.assignmentLetterId,
        letter: { letterNumber: r.assignmentLetterId, auditTitle: r.reportTitle, leader: 'Lead Auditor' } as any
      })
    }
  })

  // Ensure default known letters strictly format ST-XXX/SKAI/2026
  const defaults = ['ST-001/SKAI/2026', 'ST-002/SKAI/2026', 'ST-003/SKAI/2026', 'ST-004/SKAI/2026', 'ST-005/SKAI/2026']
  defaults.forEach(defNum => {
    if (!uniqueNumbers.has(defNum)) {
      uniqueNumbers.add(defNum)
      lettersFromStore.push({
        label: defNum,
        value: defNum,
        letter: { letterNumber: defNum, auditTitle: 'Audit Penugasan', leader: 'Head of SKAI' } as any
      })
    }
  })

  return lettersFromStore
})

interface AssignmentLetterDropdownItem {
  label: string
  value: string
  letter: any
}

const selectedAssignmentLetter = ref<AssignmentLetterDropdownItem | undefined>(
  assignmentLetterDropdownOptions.value.find(o => o.value === (store.form.assignmentLetterId || 'ST-001/SKAI/2026')) ||
  assignmentLetterDropdownOptions.value[0]
)

const linkedAssignmentLetter = computed(() => {
  const stNum = selectedAssignmentLetter.value?.value || store.form.assignmentLetterId || 'ST-001/SKAI/2026'

  if (!stNum) return null
  return (assignmentLetterStore.assignmentLetterList.find(st => st.letterNumber === stNum) as any) || {
    letterNumber: stNum,
    auditTitle: store.form.narrative ? store.form.narrative.slice(0, 70) + '...' : 'Penugasan Audit SPI',
    status: 'Published' as any,
    leader: 'Head of SKAI',
    executionPeriod: `${store.form.periodeBulan} ${store.form.tahun}`,
    auditYear: String(store.form.tahun || new Date().getFullYear())
  }
})

const lhaDropdownOptions: Ref<Array<{ label: string; value: string; report: any }>> = computed(() => {
  return auditReportStore.reportList
    .filter(r => {
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
    .map(r => ({
      label: `${r.reportNumber || (r as any).report_number} - ${r.reportTitle}`,
      value: r.reportNumber || (r as any).report_number,
      report: r
    }))
})

const selectedLhaId = ref<{ label: string; value: string; report: any } | undefined>(
  lhaDropdownOptions.value.find(o => o.value === store.form.nomorDokumen)
)

const onAssignmentLetterSelect = (val: any) => {
  if (!val) return
  const selectedNum = typeof val === 'object' ? val.value : val
  store.form.assignmentLetterId = selectedNum
  const found = assignmentLetterDropdownOptions.value.find(o => o.value === selectedNum)
  if (found) {
    selectedAssignmentLetter.value = found
  }

  // Auto-find and link corresponding LHA if exists
  const matchingLha = auditReportStore.reportList.find(r => r.assignmentLetterId === selectedNum)
  if (matchingLha) {
    onLhaSelect(matchingLha.reportNumber)
  }
}

const syncFindingsFromLha = (lhaItem?: any) => {
  const targetDoc = typeof selectedLhaId.value === 'object'
    ? (selectedLhaId.value as any)?.value
    : (selectedLhaId.value || store.form.nomorDokumen)

  const item = lhaItem || auditReportStore.reportList.find(r => 
    (r.reportNumber || (r as any).report_number) === targetDoc
  )
  if (!item) return

  store.form.nomorDokumen = item.reportNumber || (item as any).report_number
  if (item.assignmentLetterId) {
    store.form.assignmentLetterId = item.assignmentLetterId
    const found = assignmentLetterDropdownOptions.value.find(o => o.value === item.assignmentLetterId)
    if (found) selectedAssignmentLetter.value = found
  }
  if (item.executiveSummary) {
    store.form.narrative = item.executiveSummary
  }

  const findings = item.findings || []
  if (findings.length > 0) {
    store.form.topFindings = findings.slice(0, 5).map((f: any) => {
      let div = item.department || 'OP'
      if (item.reportTitle?.includes('Keuangan')) div = 'FIN'
      else if (item.reportTitle?.includes('Sistem') || item.reportTitle?.includes('ERP') || item.reportTitle?.includes('TI')) div = 'IT'
      else if (item.reportTitle?.includes('SDM') || item.reportTitle?.includes('Payroll')) div = 'HR'
      else if (item.reportTitle?.includes('Procurement') || item.reportTitle?.includes('SCM')) div = 'KK/KSD'

      let risiko: 'Tinggi' | 'Sedang' | 'Rendah' = 'Tinggi'
      if (f.category === 'Very Significant' || f.category === 'Significant') risiko = 'Tinggi'
      else if (f.category === 'Quite Significant') risiko = 'Sedang'
      else if (f.category === 'Not Significant') risiko = 'Rendah'

      return {
        unitDivision: div,
        judulTemuan: f.title || f.finding || '',
        risiko,
        statusTL: 'In Progress',
        usulan: f.action || f.recommendation || 'Perbaikan SOP dan Kontrol Pengendalian Internal'
      }
    })

    store.form.risikoTinggi = findings.filter((f: any) => ['Very Significant', 'Significant'].includes(f.category)).length
    store.form.risikoSedang = findings.filter((f: any) => f.category === 'Quite Significant').length
    store.form.risikoRendah = findings.filter((f: any) => f.category === 'Not Significant').length
  }

  if (item.findingsCount) {
    store.form.jumlahRekomendasi = item.findingsCount
  } else if (findings.length > 0) {
    store.form.jumlahRekomendasi = findings.length
  }

  // Section V & VII: Analisis Temuan Berulang & Kesimpulan
  if ((item as any).findingSummary || (item as any).finding_summary) {
    store.form.akarMasalah = (item as any).findingSummary || (item as any).finding_summary
  } else if (findings.length > 0) {
    store.form.akarMasalah = `Analisis temuan berulang mengidentifikasi kelemahan pada kepatuhan SOP dan monitoring: ${findings.slice(0, 2).map((f: any) => f.title).join('; ')}.`
  }

  if ((item as any).conclusion) {
    store.form.kesimpulan = (item as any).conclusion
  } else if (!store.form.kesimpulan && item.reportTitle) {
    store.form.kesimpulan = `Tata kelola dan pengendalian internal pada ${item.reportTitle} berjalan cukup memadai dengan prioritas penyelesaian rekomendasi temuan signifikan.`
  }

  selectedLhaId.value = lhaDropdownOptions.value.find(o => o.value === store.form.nomorDokumen)
}

const onLhaSelect = (val: any) => {
  if (!val) return
  const selectedNum = typeof val === 'object' ? val.value : val
  const item = auditReportStore.reportList.find(r => (r.reportNumber || (r as any).report_number) === selectedNum)
  if (item) {
    syncFindingsFromLha(item)
  }
}

// Watch initial state when opening form
watch(() => store.form.assignmentLetterId, (newVal) => {
  if (newVal && selectedAssignmentLetter.value?.value !== newVal) {
    const found = assignmentLetterDropdownOptions.value.find(o => o.value === newVal)
    if (found) selectedAssignmentLetter.value = found
  }
}, { immediate: true })

watch(() => store.form.nomorDokumen, (newDoc) => {
  if (newDoc) {
    selectedLhaId.value = lhaDropdownOptions.value.find(o => o.value === newDoc)
    if (!store.form.assignmentLetterId) {
      const match = auditReportStore.reportList.find(r => r.reportNumber === newDoc)
      if (match && match.assignmentLetterId) {
        store.form.assignmentLetterId = match.assignmentLetterId
        const found = assignmentLetterDropdownOptions.value.find(o => o.value === match.assignmentLetterId)
        if (found) selectedAssignmentLetter.value = found
      }
    }
    // If findings are empty, auto-sync from LHA
    if (store.form.topFindings.length === 0) {
      const match = auditReportStore.reportList.find(r => r.reportNumber === newDoc)
      if (match) {
        syncFindingsFromLha(match)
      }
    }
  }
}, { immediate: true })

</script>
