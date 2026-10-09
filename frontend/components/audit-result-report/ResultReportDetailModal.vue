<template>
  <UModal
    :open="open"
    dismissible
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-6 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/60 dark:bg-black/85 backdrop-blur-md'
    }"
    @update:open="emit('update:open', $event)"
  >
    <template #content>
      <div class="flex flex-col h-full max-h-[90vh]">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-850">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-xl bg-primary-100 dark:bg-primary-950/60 text-primary-600 dark:text-primary-400">
              <UIcon name="i-lucide-eye" class="size-6" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">
                  Detail Laporan Hasil Audit (LHA)
                </h3>
                <UBadge :color="report?.status === 'Final' ? 'success' : 'neutral'" variant="subtle" size="xs">
                  {{ (report?.status || 'Draft').toUpperCase() }}
                </UBadge>
              </div>
              <p class="text-xs text-gray-500 font-mono mt-0.5">
                {{ report?.reportNumber || '-' }} • Surat Tugas: <span class="font-semibold text-primary-600">{{ report?.assignmentLetterId || '-' }}</span>
              </p>
            </div>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark"
            @click="emit('update:open', false)"
          />
        </div>

        <!-- Body -->
        <div class="p-6 overflow-y-auto flex-1 space-y-6">
          <!-- Overview Cards -->
          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
            <div class="p-4 rounded-xl bg-slate-50 dark:bg-gray-850 border border-slate-200 dark:border-slate-800">
              <span class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider block">No. LHA</span>
              <span class="text-sm font-mono font-bold text-primary-600 dark:text-primary-400 mt-1 block truncate">
                {{ report?.reportNumber || '-' }}
              </span>
            </div>

            <div class="p-4 rounded-xl bg-slate-50 dark:bg-gray-850 border border-slate-200 dark:border-slate-800">
              <span class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider block">Tanggal Laporan</span>
              <span class="text-sm font-semibold text-gray-900 dark:text-gray-100 mt-1 block">
                {{ formatDate(report?.reportDate) }}
              </span>
            </div>

            <div class="p-4 rounded-xl bg-slate-50 dark:bg-gray-850 border border-slate-200 dark:border-slate-800">
              <span class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider block">Perusahaan / Entitas</span>
              <span class="text-sm font-semibold text-gray-900 dark:text-gray-100 mt-1 block truncate" :title="report?.companyName">
                {{ report?.companyName || 'PT AIFL Indonesia' }}
              </span>
            </div>

            <div class="p-4 rounded-xl bg-slate-50 dark:bg-gray-850 border border-slate-200 dark:border-slate-800">
              <span class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider block">Total Temuan</span>
              <span class="text-sm font-bold text-primary-600 dark:text-primary-400 mt-1 block">
                {{ report?.findingsCount || report?.findings?.length || 0 }} Temuan
              </span>
            </div>
          </div>

          <!-- Report Title Card -->
          <div class="p-5 rounded-xl bg-slate-50 dark:bg-gray-850 border border-slate-200 dark:border-slate-800 space-y-1">
            <span class="text-xs font-bold text-gray-500 uppercase tracking-wider">Judul Laporan</span>
            <h4 class="text-base font-bold text-gray-900 dark:text-white">
              {{ report?.reportTitle || '-' }}
            </h4>
          </div>

          <!-- Executive Summary Section -->
          <div class="space-y-2">
            <h4 class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 flex items-center gap-2 border-b border-gray-100 dark:border-gray-800 pb-2">
              <UIcon name="i-heroicons-document-text" class="size-4 text-primary-600" />
              Ringkasan Eksekutif (Executive Summary)
            </h4>
            <div class="p-4 rounded-xl bg-slate-50/70 dark:bg-gray-850/60 border border-slate-200 dark:border-slate-800 text-xs sm:text-sm text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-line">
              {{ report?.executiveSummary || 'Berdasarkan hasil penugasan audit yang telah diselesaikan sesuai dengan Surat Tugas, ruang lingkup evaluasi pengendalian internal telah dirangkum dalam temuan dan rekomendasi tindak lanjut di bawah ini.' }}
            </div>
          </div>

          <!-- Findings & Actions Section -->
          <div class="space-y-3">
            <div class="flex justify-between items-center border-b border-gray-100 dark:border-gray-800 pb-2">
              <h4 class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 flex items-center gap-2">
                <UIcon name="i-heroicons-list-bullet" class="size-4 text-primary-600" />
                Daftar Temuan Audit & Rekomendasi
              </h4>
              <UBadge color="primary" variant="subtle" size="xs">
                {{ report?.findings?.length || 0 }} Temuan
              </UBadge>
            </div>

            <!-- Findings List Cards -->
            <div v-if="report?.findings && report.findings.length > 0" class="space-y-3">
              <div
                v-for="(finding, idx) in report.findings"
                :key="idx"
                class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-gray-850 space-y-3 shadow-xs"
              >
                <div class="flex justify-between items-start gap-2">
                  <div class="flex items-center gap-2">
                    <span class="w-6 h-6 rounded-full bg-slate-100 dark:bg-gray-700 text-slate-700 dark:text-slate-200 flex items-center justify-center font-bold text-xs">
                      {{ idx + 1 }}
                    </span>
                    <span class="font-bold text-sm text-gray-900 dark:text-white">
                      {{ finding.title }}
                    </span>
                  </div>
                  <UBadge :color="getCategoryColor(finding.category)" variant="subtle" size="xs" class="shrink-0">
                    {{ finding.category }}
                  </UBadge>
                </div>

                <div class="pl-8 text-xs space-y-1.5">
                  <div class="p-3 bg-slate-50 dark:bg-gray-900 rounded-lg border border-slate-100 dark:border-gray-800">
                    <span class="font-semibold text-gray-700 dark:text-gray-300 block mb-0.5">Rekomendasi / Tindak Lanjut:</span>
                    <span class="text-gray-600 dark:text-gray-400">{{ finding.action || 'Tidak ada tindakan yang dicatat.' }}</span>
                  </div>
                  <span v-if="finding.source" class="text-[11px] text-gray-400 italic block">
                    Sumber: {{ finding.source }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Empty State for Findings -->
            <div
              v-else
              class="p-6 bg-slate-50 dark:bg-slate-850/50 rounded-xl border border-dashed border-slate-200 dark:border-slate-800 text-center text-xs text-gray-500 italic"
            >
              Tidak ada rincian temuan audit pada laporan ini.
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-850">
          <UButton
            label="Tutup"
            color="neutral"
            variant="ghost"
            @click="emit('update:open', false)"
          />

          <div class="flex items-center gap-2">
            <UButton
              v-if="report?.id"
              label="Download Docx"
              color="success"
              variant="outline"
              icon="i-heroicons-arrow-down-tray"
              size="sm"
              @click="store.downloadDocx(report.id, report.reportNumber)"
            />
            <UButton
              label="Cetak Dokumen LHA"
              color="primary"
              icon="i-lucide-printer"
              size="sm"
              @click="handleOpenPrint"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAuditResultReportStore, type AuditResultReport } from '~/stores/audit-result-report'

const props = defineProps<{
  open: boolean
  report: AuditResultReport | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'print', report: AuditResultReport): void
}>()

const store = useAuditResultReportStore()

const formatDate = (rawDate?: string) => {
  if (!rawDate) return '-'
  try {
    const clean = rawDate.split('T')[0]
    const d = new Date(clean)
    if (isNaN(d.getTime())) return rawDate
    return d.toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'long',
      year: 'numeric'
    })
  } catch {
    return rawDate
  }
}

const getCategoryColor = (cat?: string) => {
  switch (cat) {
    case 'Very Significant': return 'error'
    case 'Significant': return 'warning'
    case 'Quite Significant': return 'info'
    case 'Not Significant': return 'success'
    default: return 'neutral'
  }
}

const handleOpenPrint = () => {
  if (props.report) {
    emit('print', props.report)
  }
}
</script>
