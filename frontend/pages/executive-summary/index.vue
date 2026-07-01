<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Executive Summary Report</h1>
        <p class="text-gray-500">Kompilasi dan visualisasi laporan hasil audit triwulanan untuk Direksi dan Komite Audit.</p>
      </div>
      <UButton
        v-if="!store.showForm"
        color="primary"
        icon="i-heroicons-plus"
        label="Buat Laporan Kompilasi Baru"
        @click="store.initForm(selectedTab)"
      />
    </div>

    <!-- Warning Sync Notification -->
    <div 
      v-if="store.showForm && store.isSyncWarning" 
      class="p-4 bg-amber-50 border-l-4 border-amber-500 rounded-md text-amber-800 flex items-start gap-3 shadow-xs"
    >
      <UIcon name="i-heroicons-exclamation-triangle" class="size-6 text-amber-500 shrink-0 mt-0.5" />
      <div>
        <span class="font-semibold">Peringatan Sinkronisasi:</span>
        Jumlah total temuan di ringkasan statistik (Section II) tidak sinkron dengan total data di Matriks Induk. Harap sesuaikan data di bagian Matriks Induk atau Section II agar selaras.
      </div>
    </div>

    <!-- Active Form View -->
    <div v-if="store.showForm" class="space-y-8">
      <UCard>
        <template #header>
          <div class="flex justify-between items-center">
            <h2 class="text-lg font-semibold text-gray-900">
              {{ store.isEditing ? 'Edit Laporan Kompilasi' : 'Kompilasi Laporan Baru' }} - Triwulan {{ store.currentTriwulan }}
            </h2>
            <UButton 
              color="neutral" 
              variant="ghost" 
              icon="i-heroicons-x-mark" 
              @click="store.showForm = false" 
            />
          </div>
        </template>

        <div class="space-y-8">
          <!-- Document Upload & General Metadata Form -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4 p-4 bg-gray-50 rounded-lg">
            <div>
              <label class="label">Periode Bulan</label>
              <USelectMenu v-model="store.activeReport.bulan" :items="store.bulanOptions" />
            </div>
            <div>
              <label class="label">Tahun</label>
              <USelectMenu v-model="store.activeReport.tahun" :items="store.tahunOptions" />
            </div>
            <div class="col-span-2">
              <label class="label">Nomor Dokumen</label>
              <UInput 
                v-model="store.activeReport.nomorDokumen" 
                placeholder="Format: LKA/B/.../SPI/KAI/.../2026"
              />
            </div>
            <div class="col-span-4 mt-2">
              <label class="label">Upload Dokumen Resmi (PDF/Word, Max 10MB)</label>
              <div class="flex items-center gap-3">
                <input 
                  type="file" 
                  accept=".pdf,.docx" 
                  class="hidden" 
                  ref="fileInput"
                  @change="handleFileChange"
                />
                <UButton 
                  color="neutral" 
                  variant="subtle" 
                  icon="i-heroicons-cloud-arrow-up" 
                  label="Pilih File Laporan Resmi" 
                  @click="$refs.fileInput.click()" 
                />
                <span class="text-sm text-gray-600 font-medium">
                  {{ store.activeReport.fileName || 'Belum ada file yang diunggah' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Section I: Narrative -->
          <div class="space-y-2">
            <h3 class="text-md font-semibold text-gray-800 border-b pb-1">Section I: Executive Summary Narrative</h3>
            <p class="text-xs text-gray-500">Sajikan narasi eksekutif laporan secara terstruktur:</p>
            <UTextarea 
              v-model="store.activeReport.executiveSummaryNarrative" 
              rows="4" 
              class="w-full" 
            />
          </div>

          <!-- Section II: Statistik Kompilasi -->
          <div class="space-y-4">
            <h3 class="text-md font-semibold text-gray-800 border-b pb-1">Section II: Statistik Kompilasi</h3>
            <div class="grid grid-cols-2 md:grid-cols-6 gap-4">
              <div>
                <label class="label">Jumlah Laporan</label>
                <UInput type="number" v-model="store.activeReport.jumlahLaporan" />
              </div>
              <div>
                <label class="label">Risiko Tinggi</label>
                <UInput type="number" v-model="store.activeReport.risikoTinggi" />
              </div>
              <div>
                <label class="label">Risiko Sedang</label>
                <UInput type="number" v-model="store.activeReport.risikoSedang" />
              </div>
              <div>
                <label class="label">Risiko Rendah</label>
                <UInput type="number" v-model="store.activeReport.risikoRendah" />
              </div>
              <div>
                <label class="label">Total Temuan</label>
                <UInput :value="store.totalFindings" disabled class="bg-gray-100" />
              </div>
              <div>
                <label class="label">Rekomendasi</label>
                <UInput type="number" v-model="store.activeReport.jumlahRekomendasi" />
              </div>
            </div>
          </div>

          <!-- Section III: Status Tindak Lanjut -->
          <div class="space-y-3">
            <h3 class="text-md font-semibold text-gray-800 border-b pb-1">Section III: Status Tindak Lanjut</h3>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 border text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-2 text-left font-medium text-gray-700">Status</th>
                    <th class="px-4 py-2 text-left font-medium text-gray-700 w-32">Jumlah</th>
                    <th class="px-4 py-2 text-left font-medium text-gray-700 w-32">Persentase (%)</th>
                    <th class="px-4 py-2 text-left font-medium text-gray-700">Keterangan</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr>
                    <td class="px-4 py-2 font-medium text-gray-900">Closed</td>
                    <td class="px-4 py-2">
                      <UInput type="number" v-model="store.activeReport.statusClosedCount" />
                    </td>
                    <td class="px-4 py-2 text-gray-500 font-medium">{{ store.percentClosed }}%</td>
                    <td class="px-4 py-2">
                      <UInput v-model="store.activeReport.keteranganClosed" />
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 font-medium text-gray-900">In Progress</td>
                    <td class="px-4 py-2">
                      <UInput type="number" v-model="store.activeReport.statusInProgressCount" />
                    </td>
                    <td class="px-4 py-2 text-gray-500 font-medium">{{ store.percentInProgress }}%</td>
                    <td class="px-4 py-2">
                      <UInput v-model="store.activeReport.keteranganInProgress" />
                    </td>
                  </tr>
                  <tr>
                    <td class="px-4 py-2 font-medium text-gray-900">Overdue</td>
                    <td class="px-4 py-2">
                      <UInput type="number" v-model="store.activeReport.statusOverdueCount" />
                    </td>
                    <td class="px-4 py-2 text-gray-500 font-medium">{{ store.percentOverdue }}%</td>
                    <td class="px-4 py-2">
                      <UInput v-model="store.activeReport.keteranganOverdue" />
                    </td>
                  </tr>
                  <tr class="bg-gray-50 font-semibold">
                    <td class="px-4 py-2">Total</td>
                    <td class="px-4 py-2">{{ store.totalStatusCount }}</td>
                    <td class="px-4 py-2">100%</td>
                    <td class="px-4 py-2 text-gray-400 font-normal">Calculated Total</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Section IV: Top 5 Temuan Signifikan -->
          <div class="space-y-3">
            <div class="flex justify-between items-center border-b pb-1">
              <h3 class="text-md font-semibold text-gray-800">Section IV: Top 5 Temuan Signifikan</h3>
              <UButton 
                v-if="store.activeReport.topFindings.length < 5"
                color="primary" 
                variant="ghost" 
                size="xs" 
                icon="i-heroicons-plus" 
                label="Tambah Baris"
                @click="addTopFindingRow"
              />
            </div>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 border text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium text-gray-700 w-40">Unit/Divisi</th>
                    <th class="px-3 py-2 text-left font-medium text-gray-700">Judul Temuan</th>
                    <th class="px-3 py-2 text-left font-medium text-gray-700 w-36">Risiko</th>
                    <th class="px-3 py-2 text-left font-medium text-gray-700 w-36">Status TL</th>
                    <th class="px-3 py-2 text-left font-medium text-gray-700 w-44">Usulan</th>
                    <th class="px-3 py-2 text-center font-medium text-gray-700 w-16">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr v-for="(row, idx) in store.activeReport.topFindings" :key="idx">
                    <td class="px-2 py-1.5">
                      <UInput v-model="row.unit" placeholder="Contoh: OP" />
                    </td>
                    <td class="px-2 py-1.5">
                      <UInput v-model="row.judul" placeholder="Judul temuan utama" />
                    </td>
                    <td class="px-2 py-1.5">
                      <USelectMenu v-model="row.risiko" :items="['Tinggi', 'Sedang', 'Rendah']" />
                    </td>
                    <td class="px-2 py-1.5">
                      <USelectMenu v-model="row.status" :items="['Closed', 'In Progress', 'Overdue']" />
                    </td>
                    <td class="px-2 py-1.5">
                      <UInput v-model="row.usulan" placeholder="Eskalasi Direksi" />
                    </td>
                    <td class="px-2 py-1.5 text-center">
                      <UButton 
                        color="error" 
                        variant="ghost" 
                        icon="i-heroicons-trash" 
                        size="xs" 
                        @click="removeTopFindingRow(idx)"
                      />
                    </td>
                  </tr>
                  <tr v-if="store.activeReport.topFindings.length === 0">
                    <td colspan="6" class="text-center py-4 text-gray-400">Belum ada temuan signifikan dimasukkan.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Section V, VII: Analisis Temuan Berulang & Kesimpulan & TTD -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="label font-semibold">Section V: Analisis Temuan Berulang</label>
              <UTextarea 
                v-model="store.activeReport.temaBerulang" 
                placeholder="Akar masalah & Usulan revisi kebijakan"
                rows="4"
              />
            </div>
            <div class="space-y-2">
              <label class="label font-semibold">Section VII: Kesimpulan & Rekomendasi</label>
              <UTextarea 
                v-model="store.activeReport.kesimpulan" 
                placeholder="Pernyataan umum, permohonan arahan Direksi, usulan task force"
                rows="4"
              />
            </div>
          </div>

          <!-- Section VIII: Tanda Tangan Elektronik -->
          <div class="p-4 bg-gray-50 rounded-lg">
            <h4 class="text-sm font-semibold text-gray-800 mb-3">Tanda Tangan Elektronik</h4>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div>
                <label class="label">Tempat</label>
                <UInput v-model="store.activeReport.ttdTempat" />
              </div>
              <div>
                <label class="label">Tanggal</label>
                <UInput type="date" v-model="store.activeReport.ttdTanggal" />
              </div>
              <div>
                <label class="label">Nama Kepala SPI</label>
                <UInput v-model="store.activeReport.ttdNama" />
              </div>
              <div>
                <label class="label">NIK</label>
                <UInput v-model="store.activeReport.ttdNik" />
              </div>
            </div>
          </div>

          <!-- Section VI: Tren & Grafik Visuals (System Generated preview) -->
          <div class="space-y-3">
            <h3 class="text-md font-semibold text-gray-800 border-b pb-1">Section VI: Tren & Grafik (Pratinjau Visualisasi)</h3>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <!-- Bar Chart: Jumlah Temuan per Kategori Risiko -->
              <div class="border p-4 rounded-xl bg-white shadow-xs">
                <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">Jumlah Temuan per Kategori</h4>
                <div class="flex items-end justify-around h-40 pt-4 px-2 border-b">
                  <div class="flex flex-col items-center w-8">
                    <div 
                      class="bg-error-500 w-full rounded-t transition-all duration-500" 
                      :style="{ height: `${Math.min(store.activeReport.risikoTinggi * 10, 100)}px` }"
                    ></div>
                    <span class="text-xs font-bold text-gray-700 mt-1">{{ store.activeReport.risikoTinggi }}</span>
                    <span class="text-[10px] text-gray-500 mt-1">Tinggi</span>
                  </div>
                  <div class="flex flex-col items-center w-8">
                    <div 
                      class="bg-warning-500 w-full rounded-t transition-all duration-500" 
                      :style="{ height: `${Math.min(store.activeReport.risikoSedang * 10, 100)}px` }"
                    ></div>
                    <span class="text-xs font-bold text-gray-700 mt-1">{{ store.activeReport.risikoSedang }}</span>
                    <span class="text-[10px] text-gray-500 mt-1">Sedang</span>
                  </div>
                  <div class="flex flex-col items-center w-8">
                    <div 
                      class="bg-success-500 w-full rounded-t transition-all duration-500" 
                      :style="{ height: `${Math.min(store.activeReport.risikoRendah * 10, 100)}px` }"
                    ></div>
                    <span class="text-xs font-bold text-gray-700 mt-1">{{ store.activeReport.risikoRendah }}</span>
                    <span class="text-[10px] text-gray-500 mt-1">Rendah</span>
                  </div>
                </div>
              </div>

              <!-- Pie Chart: Persentase Status Tindak Lanjut -->
              <div class="border p-4 rounded-xl bg-white shadow-xs">
                <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">Status Tindak Lanjut (%)</h4>
                <div class="flex flex-col justify-between h-40">
                  <div class="space-y-2">
                    <div>
                      <div class="flex justify-between text-xs font-medium text-gray-700 mb-1">
                        <span>Closed</span>
                        <span>{{ store.percentClosed }}%</span>
                      </div>
                      <div class="w-full bg-gray-100 rounded-full h-2">
                        <div class="bg-success-500 h-2 rounded-full" :style="{ width: `${store.percentClosed}%` }"></div>
                      </div>
                    </div>
                    <div>
                      <div class="flex justify-between text-xs font-medium text-gray-700 mb-1">
                        <span>In Progress</span>
                        <span>{{ store.percentInProgress }}%</span>
                      </div>
                      <div class="w-full bg-gray-100 rounded-full h-2">
                        <div class="bg-primary-500 h-2 rounded-full" :style="{ width: `${store.percentInProgress}%` }"></div>
                      </div>
                    </div>
                    <div>
                      <div class="flex justify-between text-xs font-medium text-gray-700 mb-1">
                        <span>Overdue</span>
                        <span>{{ store.percentOverdue }}%</span>
                      </div>
                      <div class="w-full bg-gray-100 rounded-full h-2">
                        <div class="bg-error-500 h-2 rounded-full" :style="{ width: `${store.percentOverdue}%` }"></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Line Chart: Aging Rekomendasi (Visual Gauge) -->
              <div class="border p-4 rounded-xl bg-white shadow-xs">
                <h4 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Penyelesaian On-Time</h4>
                <div class="flex flex-col items-center justify-center h-40">
                  <div class="relative size-24 flex items-center justify-center">
                    <svg class="size-full transform -rotate-90" viewBox="0 0 36 36">
                      <path
                        class="text-gray-100"
                        stroke-width="3"
                        stroke="currentColor"
                        fill="none"
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                      />
                      <path
                        class="text-primary-500"
                        stroke-width="3"
                        stroke-dasharray="95, 100"
                        stroke-linecap="round"
                        stroke="currentColor"
                        fill="none"
                        d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                      />
                    </svg>
                    <span class="absolute text-lg font-bold text-gray-900">95%</span>
                  </div>
                  <span class="text-[10px] text-gray-400 mt-2 text-center">Target Penyelesaian Standar Operasional</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-3">
            <UButton color="neutral" variant="ghost" label="Batal" @click="store.showForm = false" />
            <UButton color="primary" label="Simpan Laporan Kompilasi" @click="store.saveActiveReport" />
          </div>
        </template>
      </UCard>
    </div>

    <!-- History / Summary List Table -->
    <div v-else class="space-y-4">
      <div class="border-b border-gray-200">
        <nav class="flex gap-4" aria-label="Tabs">
          <button
            v-for="q in ['Q1', 'Q2', 'Q3', 'Q4']"
            :key="q"
            @click="selectedTab = q as any"
            class="px-4 py-2 text-sm font-semibold border-b-2 transition-all duration-200"
            :class="selectedTab === q ? 'border-primary-500 text-primary-500' : 'border-transparent text-gray-500 hover:text-gray-700'"
          >
            Triwulan {{ q }}
          </button>
        </nav>
      </div>

      <div v-if="filteredReports.length > 0" class="space-y-4">
        <div 
          v-for="item in filteredReports" 
          :key="item.id" 
          class="bg-white border rounded-xl overflow-hidden shadow-xs hover:shadow-md transition-all duration-300"
        >
          <!-- Report Header Panel -->
          <div class="p-5 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-gray-50 border-b">
            <div>
              <div class="flex items-center gap-2 flex-wrap">
                <h3 class="text-md font-bold text-gray-900">{{ item.nomorDokumen }}</h3>
                <UBadge :color="item.isApproved ? 'success' : 'neutral'" variant="soft">
                  {{ item.isApproved ? 'Approved (Locked)' : 'Draft' }}
                </UBadge>
              </div>
              <p class="text-xs text-gray-400 mt-1">
                Periode: {{ item.bulan }} {{ item.tahun }} | File Lampiran: 
                <span class="font-medium text-primary-600">{{ item.fileName }}</span>
              </p>
            </div>
            <div class="flex gap-2 shrink-0">
              <UButton 
                :color="item.isApproved ? 'warning' : 'success'"
                variant="subtle"
                size="sm"
                :icon="item.isApproved ? 'i-heroicons-lock-open' : 'i-heroicons-lock-closed'"
                :label="item.isApproved ? 'Unlock (Draft)' : 'Approve & Lock'"
                @click="store.toggleApproval(item.id)"
              />
              <UButton 
                v-if="!item.isApproved"
                color="primary" 
                variant="ghost" 
                size="sm" 
                icon="i-heroicons-pencil-square" 
                label="Edit"
                @click="store.editReport(item)"
              />
              <UButton 
                v-if="!item.isApproved"
                color="error" 
                variant="ghost" 
                size="sm" 
                icon="i-heroicons-trash" 
                label="Hapus"
                @click="store.deleteReport(item.id)"
              />
              <UButton 
                color="neutral" 
                variant="subtle" 
                size="sm" 
                icon="i-heroicons-table-cells" 
                label="Edit Matriks Temuan"
                @click="navigateTo('/executive-summary/matriks')"
              />
            </div>
          </div>

          <!-- Quick statistics overview -->
          <div class="p-5 grid grid-cols-2 md:grid-cols-5 gap-4 divide-y md:divide-y-0 md:divide-x divide-gray-100 text-center">
            <div>
              <span class="block text-xs font-semibold text-gray-400 uppercase tracking-wider">Jumlah Laporan</span>
              <span class="block text-xl font-bold text-gray-800 mt-1">{{ item.jumlahLaporan }} LHA</span>
            </div>
            <div class="pt-3 md:pt-0">
              <span class="block text-xs font-semibold text-gray-400 uppercase tracking-wider">Total Temuan</span>
              <span class="block text-xl font-bold text-gray-800 mt-1">
                {{ Number(item.risikoTinggi) + Number(item.risikoSedang) + Number(item.risikoRendah) }}
              </span>
              <div class="flex justify-center gap-1.5 mt-1 text-[10px]">
                <span class="text-error-600 font-semibold">{{ item.risikoTinggi }} H</span>
                <span class="text-gray-300">|</span>
                <span class="text-warning-600 font-semibold">{{ item.risikoSedang }} M</span>
                <span class="text-gray-300">|</span>
                <span class="text-success-600 font-semibold">{{ item.risikoRendah }} L</span>
              </div>
            </div>
            <div class="pt-3 md:pt-0">
              <span class="block text-xs font-semibold text-gray-400 uppercase tracking-wider">Closed</span>
              <span class="block text-xl font-bold text-success-600 mt-1">{{ item.statusClosedCount }}</span>
              <span class="block text-[10px] text-gray-400">{{ getPercentage(item.statusClosedCount, item) }}%</span>
            </div>
            <div class="pt-3 md:pt-0">
              <span class="block text-xs font-semibold text-gray-400 uppercase tracking-wider">In Progress</span>
              <span class="block text-xl font-bold text-primary-600 mt-1">{{ item.statusInProgressCount }}</span>
              <span class="block text-[10px] text-gray-400">{{ getPercentage(item.statusInProgressCount, item) }}%</span>
            </div>
            <div class="pt-3 md:pt-0">
              <span class="block text-xs font-semibold text-gray-400 uppercase tracking-wider">Overdue</span>
              <span class="block text-xl font-bold text-error-600 mt-1">{{ item.statusOverdueCount }}</span>
              <span class="block text-[10px] text-gray-400">{{ getPercentage(item.statusOverdueCount, item) }}%</span>
            </div>
          </div>

          <!-- Section narrative block preview -->
          <div class="px-5 pb-5 pt-2 border-t border-gray-100">
            <h4 class="text-xs font-bold text-gray-400 uppercase tracking-wider mb-2 text-left">Ringkasan Eksekutif</h4>
            <blockquote class="text-sm italic text-gray-600 border-l-4 border-gray-200 pl-3 py-1">
              "{{ item.executiveSummaryNarrative }}"
            </blockquote>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div 
        v-else 
        class="text-center py-16 bg-gray-50 rounded-xl border-2 border-dashed border-gray-200"
      >
        <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700">Belum ada Laporan Kompilasi</h3>
        <p class="text-gray-500 mt-2 max-w-sm mx-auto mb-6">
          Belum ada ringkasan eksekutif hasil audit triwulanan yang ditambahkan untuk Triwulan {{ selectedTab }} ini.
        </p>
        <UButton
          color="primary"
          icon="i-heroicons-plus"
          label="Buat Laporan Baru"
          @click="store.initForm(selectedTab)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'

const store = useExecutiveSummaryStore()
const selectedTab = ref<'Q1' | 'Q2' | 'Q3' | 'Q4'>('Q1')

const filteredReports = computed(() => {
  return store.reports.filter(r => r.triwulan === selectedTab.value)
})

const fileInput = ref<HTMLInputElement | null>(null)

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    store.activeReport.fileName = target.files[0].name
  }
}

const addTopFindingRow = () => {
  store.activeReport.topFindings.push({
    unit: '',
    judul: '',
    risiko: 'Sedang',
    status: 'In Progress',
    usulan: ''
  })
}

const removeTopFindingRow = (idx: number) => {
  store.activeReport.topFindings.splice(idx, 1)
}

const getPercentage = (value: number, item: any) => {
  const total = Number(item.statusClosedCount) + Number(item.statusInProgressCount) + Number(item.statusOverdueCount)
  if (total === 0) return 0
  return Math.round((value / total) * 1000) / 10
}
</script>
