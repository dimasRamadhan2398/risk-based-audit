<template>
  <UModal
    :open="open"
    dismissible
    :ui="{
      content: 'sm:max-w-5xl max-h-[95vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden print:max-w-none print:max-h-none print:border-none print:shadow-none print:rounded-none',
      header: 'border-b border-gray-100 dark:border-gray-800 p-4 text-gray-900 dark:text-white font-bold shrink-0 print:hidden',
      body: 'p-8 space-y-6 bg-white text-gray-900 overflow-y-auto max-h-[calc(95vh-130px)] flex-1 print:max-h-none print:overflow-visible print:p-0',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 print:hidden',
      overlay: 'bg-gray-900/60 dark:bg-black/85 backdrop-blur-md print:hidden'
    }"
    @update:open="emit('update:open', $event)"
  >
    <template #content>
      <div class="flex flex-col h-full max-h-[95vh] print:max-h-none print:h-auto">
        <!-- Modal Toolbar (Hidden during print) -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-850 print:hidden">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-100 dark:bg-primary-950 text-primary-600">
              <UIcon name="i-lucide-printer" class="size-5" />
            </div>
            <div>
              <h3 class="text-base font-bold text-gray-900 dark:text-white">
                Dokumen Laporan Hasil Audit (LHA)
              </h3>
              <p class="text-xs text-gray-500 font-mono">
                {{ report?.reportNumber || '-' }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <UButton
              color="primary"
              icon="i-lucide-printer"
              label="Cetak / Simpan PDF"
              @click="handlePrint"
            />
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="emit('update:open', false)"
            />
          </div>
        </div>

        <!-- Printable Document Body -->
        <div id="lha-printable-document" class="p-8 sm:p-12 overflow-y-auto flex-1 bg-white text-slate-900 print:p-0 print:overflow-visible">
          <!-- Kop Surat / Document Header -->
          <div class="border-b-2 border-slate-900 pb-4 mb-6">
            <div class="flex justify-between items-start">
              <div>
                <h2 class="text-xl font-extrabold uppercase tracking-wider text-slate-900">
                  {{ report?.companyName || 'PT AIFL INDONESIA' }}
                </h2>
                <p class="text-xs font-semibold text-slate-600 tracking-widest uppercase">
                  Satuan Kerja Audit Internal (SKAI)
                </p>
              </div>
              <div class="text-right">
                <UBadge :color="report?.status === 'Final' ? 'success' : 'neutral'" variant="subtle" size="sm">
                  STATUS: {{ (report?.status || 'DRAFT').toUpperCase() }}
                </UBadge>
              </div>
            </div>
          </div>

          <!-- Document Title & Numbers -->
          <div class="text-center my-6 space-y-1">
            <h1 class="text-lg sm:text-xl font-black uppercase tracking-wide text-slate-900">
              LAPORAN HASIL AUDIT (LHA)
            </h1>
            <p class="text-sm font-mono font-bold text-slate-700">
              Nomor: {{ report?.reportNumber || '-' }}
            </p>
          </div>

          <!-- Metadata Table -->
          <div class="border border-slate-300 rounded-lg overflow-hidden mb-6 text-xs sm:text-sm">
            <div class="grid grid-cols-1 sm:grid-cols-2 divide-y sm:divide-y-0 sm:divide-x divide-slate-200">
              <div class="p-3 space-y-1.5 bg-slate-50/50">
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Surat Tugas:</span>
                  <span class="font-bold text-slate-900">{{ report?.assignmentLetterId || '-' }}</span>
                </div>
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Tanggal Laporan:</span>
                  <span class="text-slate-900">{{ formatDate(report?.reportDate) }}</span>
                </div>
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Entitas / Perusahaan:</span>
                  <span class="font-semibold text-slate-900">{{ report?.companyName || 'PT AIFL Indonesia' }}</span>
                </div>
              </div>
              <div class="p-3 space-y-1.5">
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Judul Audit:</span>
                  <span class="font-bold text-slate-900">{{ report?.reportTitle || '-' }}</span>
                </div>
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Jumlah Temuan:</span>
                  <span class="font-bold text-primary-700">{{ report?.findingsCount || report?.findings?.length || 0 }} Temuan</span>
                </div>
                <div class="flex">
                  <span class="w-32 font-semibold text-slate-600">Objek Audit:</span>
                  <span class="text-slate-900">{{ report?.department || 'Divisi Keuangan & Operasional' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Section: Ringkasan Eksekutif -->
          <div class="mb-6 space-y-2">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-900 border-b border-slate-200 pb-1">
              I. Ringkasan Eksekutif
            </h3>
            <p class="text-xs sm:text-sm text-slate-700 leading-relaxed whitespace-pre-line">
              {{ report?.executiveSummary || 'Berdasarkan hasil penugasan audit yang telah dilaksanakan sesuai dengan Surat Tugas, seluruh ruang lingkup pekerjaan pemeriksaan pengendalian internal telah diselesaikan dengan rincian temuan dan rekomendasi tindak lanjut sebagaimana diuraikan di bawah ini.' }}
            </p>
          </div>

          <!-- Section: Temuan Hasil Audit -->
          <div class="mb-8 space-y-3">
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-900 border-b border-slate-200 pb-1">
              II. Daftar Temuan & Rekomendasi Audit
            </h3>

            <div v-if="report?.findings && report.findings.length > 0" class="overflow-x-auto">
              <table class="w-full text-left border-collapse border border-slate-300 text-xs sm:text-sm">
                <thead>
                  <tr class="bg-slate-100 text-slate-800 font-bold">
                    <th class="border border-slate-300 p-2.5 w-10 text-center">No</th>
                    <th class="border border-slate-300 p-2.5 w-32">Kategori / Tingkat</th>
                    <th class="border border-slate-300 p-2.5">Deskripsi Temuan</th>
                    <th class="border border-slate-300 p-2.5">Tindak Lanjut / Rekomendasi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-200">
                  <tr v-for="(f, fIdx) in report.findings" :key="fIdx" class="hover:bg-slate-50">
                    <td class="border border-slate-300 p-2.5 text-center font-bold text-slate-700">
                      {{ fIdx + 1 }}
                    </td>
                    <td class="border border-slate-300 p-2.5">
                      <span class="font-bold text-xs" :class="getCategoryTextClass(f.category)">
                        {{ f.category }}
                      </span>
                    </td>
                    <td class="border border-slate-300 p-2.5 font-medium text-slate-900">
                      {{ f.title }}
                      <span v-if="f.source" class="block text-[10px] text-slate-400 italic mt-0.5">
                        Sumber: {{ f.source }}
                      </span>
                    </td>
                    <td class="border border-slate-300 p-2.5 text-slate-700">
                      {{ f.action || '-' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-else class="p-4 bg-slate-50 border border-slate-200 rounded text-center text-xs text-slate-500 italic">
              Tidak ada temuan yang dilaporkan pada periode ini.
            </div>
          </div>

          <!-- ============================================================ -->
          <!-- HALAMAN PALING BAWAH / AKHIR: PENGESAHAN & TANDA TANGAN      -->
          <!-- ============================================================ -->
          <div class="mt-12 pt-8 border-t-2 border-slate-800 break-inside-avoid print:break-before-page">
            <!-- Informasi Tempat & Tanggal & Nama Perusahaan -->
            <div class="text-right space-y-1 mb-8">
              <p class="text-sm font-semibold text-slate-800">
                {{ report?.signaturePlace || 'Jakarta' }}, {{ formatDate(report?.signatureDate || report?.reportDate) }}
              </p>
              <p class="text-base font-extrabold uppercase tracking-wide text-slate-900">
                AUDIT INTERNAL {{ (report?.companyName || 'PT AIFL INDONESIA').toUpperCase() }}
              </p>
            </div>

            <!-- Tanda Tangan Team Member (Grid) -->
            <div class="space-y-4">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-600 border-b border-slate-200 pb-1 mb-4 text-center">
                LEMBAR PENGESAHAN TIM AUDIT
              </h4>

              <div
                v-if="displaySignatures && displaySignatures.length > 0"
                class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-6 justify-center"
              >
                <div
                  v-for="(sig, sIdx) in displaySignatures"
                  :key="sIdx"
                  class="flex flex-col items-center justify-between p-3 border border-slate-200 rounded-lg bg-slate-50/30 text-center"
                >
                  <!-- Peran / Jabatan -->
                  <div class="text-xs font-bold text-slate-600 uppercase mb-2">
                    {{ sig.role || 'Team Member' }}
                  </div>

                  <!-- Kolom Tanda Tangan Gambar / Canvas -->
                  <div class="h-24 w-full flex items-center justify-center border-b border-slate-400 mb-2">
                    <img
                      v-if="sig.signature"
                      :src="sig.signature"
                      alt="Tanda Tangan"
                      class="max-h-20 max-w-full object-contain"
                    />
                    <span v-else class="text-[11px] text-slate-400 italic">
                      (Tanda Tangan)
                    </span>
                  </div>

                  <!-- Nama Team Member -->
                  <div class="font-bold text-xs sm:text-sm text-slate-900 underline">
                    {{ sig.name }}
                  </div>
                </div>
              </div>

              <!-- Fallback if completely no members -->
              <div
                v-else
                class="grid grid-cols-2 sm:grid-cols-3 gap-6 justify-center"
              >
                <div class="flex flex-col items-center p-3 border border-slate-200 rounded-lg text-center">
                  <div class="text-xs font-bold text-slate-600 uppercase mb-2">Ketua Tim Audit</div>
                  <div class="h-20 w-full flex items-center justify-center border-b border-slate-400 mb-2 text-xs text-slate-400 italic">
                    (Belum Ditandatangani)
                  </div>
                  <div class="font-bold text-sm text-slate-900 underline">Ketua Tim</div>
                </div>
                <div class="flex flex-col items-center p-3 border border-slate-200 rounded-lg text-center">
                  <div class="text-xs font-bold text-slate-600 uppercase mb-2">Anggota Tim Audit</div>
                  <div class="h-20 w-full flex items-center justify-center border-b border-slate-400 mb-2 text-xs text-slate-400 italic">
                    (Belum Ditandatangani)
                  </div>
                  <div class="font-bold text-sm text-slate-900 underline">Team Member</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer Actions (Hidden during print) -->
        <div class="px-6 py-4 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center bg-slate-50/70 dark:bg-gray-850 print:hidden">
          <span class="text-xs text-gray-500">
            Pastikan opsi background graphics diaktifkan pada jendela cetak browser.
          </span>
          <div class="flex items-center gap-2">
            <UButton
              label="Tutup"
              color="neutral"
              variant="ghost"
              @click="emit('update:open', false)"
            />
            <UButton
              label="Cetak Dokumen"
              color="primary"
              icon="i-lucide-printer"
              @click="handlePrint"
            />
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditResultReportStore, type AuditResultReport, type MemberSignature } from '~/stores/audit-result-report'

const props = defineProps<{
  open: boolean
  report: AuditResultReport | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const store = useAuditResultReportStore()

const displaySignatures = computed<MemberSignature[]>(() => {
  const r = props.report
  if (!r) return []

  let sigs: any = r.signatures || (r as any).Signatures
  if (typeof sigs === 'string' && sigs.trim().startsWith('[')) {
    try {
      sigs = JSON.parse(sigs)
    } catch {
      sigs = []
    }
  }

  if (Array.isArray(sigs) && sigs.length > 0) {
    return sigs
  }

  // Fallback to Team Members from Assignment Letter
  const letter = r.assignmentLetterId || ''
  const members = store.getTeamMembersForLetter(letter)
  if (members && members.length > 0) {
    return members.map(m => ({
      name: m.name,
      role: m.role || 'Member',
      signature: '',
      signedAt: ''
    }))
  }

  return []
})

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

const getCategoryTextClass = (cat?: string) => {
  switch (cat) {
    case 'Very Significant': return 'text-red-600 font-extrabold'
    case 'Significant': return 'text-amber-600 font-bold'
    case 'Quite Significant': return 'text-blue-600 font-semibold'
    case 'Not Significant': return 'text-green-600 font-semibold'
    default: return 'text-slate-600'
  }
}

const handlePrint = () => {
  window.print()
}
</script>

<style>
@media print {
  body * {
    visibility: hidden;
  }
  #lha-printable-document,
  #lha-printable-document * {
    visibility: visible;
  }
  #lha-printable-document {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    margin: 0;
    padding: 0;
    background: white !important;
    color: black !important;
  }
}
</style>
