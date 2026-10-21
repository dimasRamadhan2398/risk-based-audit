<template>
  <div class="space-y-6 sm:space-y-8 animate-fade-in pb-12">
    <!-- 1. Header: Profil Unit Kerja & Pengalih Peran (Role Switcher) -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 p-5 sm:p-6 bg-gradient-to-r from-amber-500/10 via-primary-500/5 to-transparent border border-amber-500/20 rounded-2xl shadow-xs">
      <div class="flex items-start sm:items-center gap-4">
        <div class="size-14 rounded-2xl bg-amber-500/10 dark:bg-amber-500/20 border border-amber-500/30 flex items-center justify-center text-amber-600 dark:text-amber-400 shrink-0 shadow-xs">
          <UIcon name="i-lucide-building-2" class="size-7" />
        </div>
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">
              Dashboard Unit Kerja (Auditee & PIC)
            </h1>
            <UBadge color="warning" variant="subtle" size="sm" class="font-bold uppercase tracking-wider">
              Role: Auditee
            </UBadge>
            <span class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-amber-100 dark:bg-amber-950/50 text-amber-800 dark:text-amber-300 border border-amber-300/40">
              <UIcon name="i-lucide-map-pin" class="size-3" />
              {{ currentDepartment }}
            </span>
          </div>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mt-1 max-w-3xl">
            Selamat datang, <strong class="text-gray-900 dark:text-gray-200">{{ currentUserFullName }}</strong>. 
            Layar kerja khusus pemenuhan permintaan bukti dokumen audit (PBC), pembaruan rencana tindak lanjut temuan (CAPA), dan pengisian survei kepuasan audit.
          </p>
        </div>
      </div>
    </div>

    <!-- 2. Banner Survei Kepuasan Layanan Audit (Satisfaction Survey Alert) -->
    <!-- Otomatis muncul jika ada penugasan audit yang baru saja selesai (LHA Terbit) dan belum diisi surveinya -->
    <div
      v-if="hasPendingSurvey"
      class="p-4 sm:p-5 rounded-2xl bg-gradient-to-r from-indigo-500/15 via-primary-500/10 to-amber-500/10 border border-indigo-500/30 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4"
    >
      <div class="flex items-start sm:items-center gap-3.5">
        <div class="size-11 rounded-xl bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 flex items-center justify-center shrink-0 shadow-inner">
          <UIcon name="i-lucide-sparkles" class="size-6 animate-pulse" />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h3 class="text-sm sm:text-base font-bold text-gray-900 dark:text-white">
              Survei Kepuasan Layanan Audit Siap Diisi
            </h3>
            <UBadge color="primary" variant="subtle" size="xs" class="font-bold uppercase tracking-wider">
              Survei Baru
            </UBadge>
          </div>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-300 mt-0.5">
            Penugasan audit terbaru unit kerja Anda telah rampung (LHA Final terbit). Masukan Anda sangat berharga bagi peningkatan mutu dan objektivitas layanan audit internal.
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2 self-start md:self-center shrink-0">
        <UButton
          to="/audit-result-report/satisfaction-survey"
          color="primary"
          size="sm"
          icon="i-lucide-clipboard-pen"
          class="font-bold shadow-xs cursor-pointer"
        >
          Isi Survei Kepuasan Sekarang &rarr;
        </UButton>
      </div>
    </div>

    <!-- Notice Card: Fokus Operasional Auditee (Sesuai Panduan RBAC) -->
    <div class="flex items-center justify-between gap-4 p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-900 dark:text-amber-200 text-xs sm:text-sm">
      <div class="flex items-center gap-3">
        <UIcon name="i-lucide-info" class="size-5 text-amber-600 dark:text-amber-400 shrink-0" />
        <div>
          <span class="font-bold">Prinsip Tampilan Auditee:</span>
          <span class="ml-1 text-amber-800 dark:text-amber-300">
            Fokus penuh pada pemenuhan bukti audit (PBC), tindak lanjut rekomendasi (CAPA), dan pengisian survei kepuasan. Seluruh kertas kerja internal (KKA F-01 s/d F-05) dan statistik risiko makro disembunyikan.
          </span>
        </div>
      </div>
      <div class="hidden md:flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-md bg-amber-600/10 text-amber-700 dark:text-amber-300 shrink-0">
        <UIcon name="i-lucide-shield-check" class="size-3.5" />
        <span>Matriks RBAC Auditee (Row 8, 9, 10)</span>
      </div>
    </div>

    <!-- 3. 4 Kartu Metrik Operasional Auditee -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Card 1: Tindak Lanjut Terbuka (Open CAPA) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-amber-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Temuan Perlu Tindak Lanjut
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-amber-600 dark:text-amber-400 mt-1">
              {{ openFindingsCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-amber-50 dark:bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <UIcon name="i-lucide-alert-circle" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-amber-600 dark:text-amber-400 font-semibold flex items-center gap-1">
            <UIcon name="i-lucide-file-text" class="size-3.5" />
            {{ pendingPlanCount }} Menunggu Rencana Aksi
          </span>
          <span class="text-gray-400">Tahun {{ currentYear }}</span>
        </div>
      </div>

      <!-- Card 2: Sedang Dijalankan (In Progress CAPA) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-sky-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Sedang Dijalankan (CAPA)
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-sky-600 dark:text-sky-400 mt-1">
              {{ inProgressFindingsCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-sky-50 dark:bg-sky-500/10 text-sky-600 dark:text-sky-400">
            <UIcon name="i-lucide-wrench" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <div class="flex items-center gap-1.5">
            <div class="w-16 bg-gray-200 dark:bg-gray-800 h-2 rounded-full overflow-hidden">
              <div class="bg-sky-500 h-full rounded-full" :style="{ width: `${avgProgressRate}%` }" />
            </div>
            <span class="font-bold text-sky-600 dark:text-sky-400">{{ avgProgressRate }}% Rata-rata</span>
          </div>
          <NuxtLink to="/action-taken-report" class="text-sky-600 dark:text-sky-400 hover:underline font-semibold">
            Lihat &rarr;
          </NuxtLink>
        </div>
      </div>

      <!-- Card 3: Mendekati / Lewat Jatuh Tempo (Overdue) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-rose-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Lewat Jatuh Tempo
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-rose-600 dark:text-rose-400 mt-1">
              {{ overdueFindingsCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-rose-50 dark:bg-rose-500/10 text-rose-600 dark:text-rose-400">
            <UIcon name="i-lucide-alert-triangle" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span :class="overdueFindingsCount > 0 ? 'text-rose-600 dark:text-rose-400 font-bold' : 'text-gray-400'">
            {{ overdueFindingsCount > 0 ? 'Perlu Prioritas Segera' : 'Semua Tepat Waktu' }}
          </span>
          <span class="text-gray-400">{{ closedFindingsCount }} Selesai (Closed)</span>
        </div>
      </div>

      <!-- Card 4: Dokumen Audit (PBC) Pending -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-emerald-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Permintaan Bukti (PBC) Pending
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 mt-1">
              {{ pendingPbcCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <UIcon name="i-lucide-file-up" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
            <UIcon name="i-lucide-check-circle-2" class="size-3.5" />
            {{ collectedPbcCount }} Terunggah
          </span>
          <a href="#pbc-section" class="text-emerald-600 dark:text-emerald-400 hover:underline font-semibold">
            Unggah Sekarang &darr;
          </a>
        </div>
      </div>
    </div>

    <!-- 4. WIDGET 1: PERMINTAAN BUKTI DOKUMEN AUDIT (PBC - PROVIDED BY CLIENT TRACKER) -->
    <div id="pbc-section" class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-up" class="size-5 text-emerald-500" />
            1. Pemenuhan Dokumen Bukti Audit (PBC - Provided By Client)
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Daftar berkas atau data yang diminta tim auditor untuk penugasan audit aktif. Unggah dokumen langsung di sini atau kelola di modul Audit Fieldwork.
          </p>
        </div>

        <!-- Assignment Letter Selector -->
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <div class="w-full sm:w-72">
            <USelectMenu
              v-model="selectedLetterNumber"
              :items="assignmentLetterOptions"
              value-key="value"
              placeholder="Pilih Surat Tugas..."
              class="w-full"
            >
              <template #leading>
                <UIcon name="i-lucide-file-text" class="size-4 text-emerald-500" />
              </template>
            </USelectMenu>
          </div>
          <UButton
            to="/audit-fieldwork"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-external-link"
            class="shrink-0 font-medium"
            title="Buka modul Audit Fieldwork"
          >
            Fieldwork
          </UButton>
        </div>
      </div>

      <!-- Quick Stats & Status Bar -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 p-3.5 rounded-xl bg-gray-50/70 dark:bg-gray-800/40 border border-gray-100 dark:border-gray-800 text-xs">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-200">
            <UIcon name="i-lucide-files" class="size-4" />
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400 block text-[11px]">Total Berkas Diminta</span>
            <span class="font-bold text-gray-900 dark:text-white text-sm">{{ pbcStats.total }} Dokumen</span>
          </div>
        </div>

        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <UIcon name="i-lucide-check-check" class="size-4" />
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400 block text-[11px]">Sudah Terunggah (Received)</span>
            <span class="font-bold text-emerald-600 dark:text-emerald-400 text-sm">{{ pbcStats.collected }} Dokumen</span>
          </div>
        </div>

        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <UIcon name="i-lucide-alert-circle" class="size-4" />
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400 block text-[11px]">Menunggu Unggah (Pending)</span>
            <span class="font-bold text-amber-600 dark:text-amber-400 text-sm">{{ pbcStats.pending }} Dokumen</span>
          </div>
        </div>
      </div>

      <!-- Document Collection List Table -->
      <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60 font-semibold text-gray-600 dark:text-gray-300">
              <th class="py-3 px-3.5 w-12 text-center">No</th>
              <th class="py-3 px-3.5">Nama Berkas & Permintaan</th>
              <th class="py-3 px-3.5">Keterangan / Tujuan</th>
              <th class="py-3 px-3.5 text-center">Tenggat Waktu</th>
              <th class="py-3 px-3.5 text-center">Status</th>
              <th class="py-3 px-3.5 text-center w-36">Aksi Auditee</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800/80">
            <tr
              v-for="(doc, idx) in pbcItemsList"
              :key="doc.id || idx"
              class="hover:bg-gray-50/70 dark:hover:bg-gray-800/50 transition-colors"
            >
              <td class="py-3 px-3.5 text-center font-mono text-gray-500">{{ idx + 1 }}</td>
              <td class="py-3 px-3.5">
                <span class="font-semibold text-gray-900 dark:text-white block line-clamp-1">
                  {{ doc.documentName }}
                </span>
                <span v-if="doc.fileName" class="text-[11px] text-emerald-600 dark:text-emerald-400 flex items-center gap-1 mt-0.5">
                  <UIcon name="i-lucide-paperclip" class="size-3" />
                  {{ doc.fileName }}
                </span>
              </td>
              <td class="py-3 px-3.5 text-gray-600 dark:text-gray-300 max-w-xs">
                <p class="line-clamp-2 text-[11px]">
                  {{ doc.description || 'Tidak ada catatan tambahan dari auditor' }}
                </p>
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <UBadge
                  :color="getDeadlineBadgeColor(doc.requiredDate)"
                  variant="subtle"
                  size="xs"
                  class="font-semibold"
                >
                  <UIcon name="i-lucide-calendar" class="size-3 mr-1" />
                  {{ formatDisplayDate(doc.requiredDate) }}
                </UBadge>
                <span class="block text-[10px] text-gray-400 mt-0.5">
                  {{ getDeadlineLabel(doc.requiredDate) }}
                </span>
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <UBadge
                  :color="doc.status === 'Received' ? 'success' : 'warning'"
                  variant="subtle"
                  size="xs"
                  class="font-bold"
                >
                  {{ doc.status === 'Received' ? 'Diterima' : 'Menunggu' }}
                </UBadge>
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <UButton
                  v-if="doc.status === 'Pending'"
                  size="xs"
                  color="warning"
                  icon="i-lucide-upload"
                  class="font-bold cursor-pointer shadow-xs text-white dark:text-gray-900"
                  @click="openPbcUploadModal(doc)"
                >
                  Unggah Bukti
                </UButton>
                <div v-else class="flex items-center justify-center gap-1">
                  <UButton
                    size="xs"
                    variant="ghost"
                    color="neutral"
                    icon="i-lucide-download"
                    title="Unduh Berkas"
                    @click="downloadPbcDoc(doc)"
                  >
                    Unduh
                  </UButton>
                  <UButton
                    size="xs"
                    variant="ghost"
                    color="primary"
                    icon="i-lucide-refresh-cw"
                    title="Ganti Berkas"
                    @click="openPbcUploadModal(doc)"
                  >
                    Ganti
                  </UButton>
                </div>
              </td>
            </tr>

            <tr v-if="pbcItemsList.length === 0">
              <td colspan="6" class="py-10 text-center text-gray-500 dark:text-gray-400">
                <UIcon name="i-lucide-folder-open" class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
                <p class="text-xs font-semibold">Belum ada daftar permintaan bukti (PBC) untuk penugasan ini.</p>
                <p class="text-[11px] text-gray-400 mt-0.5">Pilih surat tugas lain atau tunggu permintaan dokumen dari tim auditor.</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 5. WIDGET 2: TABEL TINDAK LANJUT TEMUAN AUDIT (ACTION TAKEN REPORT / CAPA) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-clipboard-check" class="size-5 text-amber-500" />
            2. Rencana Tindak Lanjut Temuan Audit (Action Taken Report / CAPA)
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Daftar temuan hasil audit yang dialokasikan kepada unit kerja Anda. Perbarui uraian rencana tindakan, persentase progres, dan unggah berkas bukti penyelesaian.
          </p>
        </div>

        <div class="flex items-center gap-2">
          <UButton
            to="/action-taken-report"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-external-link"
            class="font-semibold"
          >
            Buka Halaman ATR Penuh
          </UButton>
        </div>
      </div>

      <!-- Filter Tabs & Search Bar -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-1">
        <div class="flex flex-wrap items-center gap-1.5">
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="selectedAtrFilter === 'all' ? 'bg-amber-500 text-white shadow-xs' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'"
            @click="selectedAtrFilter = 'all'"
          >
            Semua ({{ filteredAtrListTotal }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="selectedAtrFilter === 'active' ? 'bg-amber-500 text-white shadow-xs' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'"
            @click="selectedAtrFilter = 'active'"
          >
            Perlu Tindak Lanjut ({{ openFindingsCount }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="selectedAtrFilter === 'in_progress' ? 'bg-amber-500 text-white shadow-xs' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'"
            @click="selectedAtrFilter = 'in_progress'"
          >
            Sedang Berjalan ({{ inProgressFindingsCount }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="selectedAtrFilter === 'overdue' ? 'bg-rose-600 text-white shadow-xs' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'"
            @click="selectedAtrFilter = 'overdue'"
          >
            Lewat Jatuh Tempo ({{ overdueFindingsCount }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="selectedAtrFilter === 'closed' ? 'bg-emerald-600 text-white shadow-xs' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'"
            @click="selectedAtrFilter = 'closed'"
          >
            Selesai ({{ closedFindingsCount }})
          </button>
        </div>

        <div class="w-full sm:w-64">
          <UInput
            v-model="atrSearchQuery"
            placeholder="Cari no LHA, temuan, atau PIC..."
            icon="i-lucide-search"
            color="neutral"
            variant="outline"
            size="sm"
            class="w-full"
          />
        </div>
      </div>

      <!-- Action Taken Report Table -->
      <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60 font-semibold text-gray-600 dark:text-gray-300">
              <th class="py-3 px-3.5 w-12 text-center">No</th>
              <th class="py-3 px-3.5">No. LHA & Judul Temuan</th>
              <th class="py-3 px-3.5">Signifikansi</th>
              <th class="py-3 px-3.5">PIC & Jatuh Tempo</th>
              <th class="py-3 px-3.5 text-center">Progres Fisik</th>
              <th class="py-3 px-3.5 text-center">Status</th>
              <th class="py-3 px-3.5 text-center w-44">Aksi CAPA</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800/80">
            <tr
              v-for="(item, idx) in displayedAtrList"
              :key="item.id || idx"
              class="hover:bg-gray-50/70 dark:hover:bg-gray-800/50 transition-colors"
            >
              <td class="py-3 px-3.5 text-center font-mono text-gray-500">{{ idx + 1 }}</td>
              <td class="py-3 px-3.5 max-w-sm">
                <span class="font-mono font-bold text-primary-600 dark:text-primary-400 text-[11px] block">
                  {{ item.report_number || (item.lha as any)?.report_number || 'LHA-REF' }}
                </span>
                <span class="font-bold text-gray-900 dark:text-white line-clamp-1 mt-0.5">
                  {{ item.finding_title }}
                </span>
                <p v-if="item.action_plan" class="text-[11px] text-gray-500 dark:text-gray-400 line-clamp-1 mt-0.5">
                  <strong class="text-gray-700 dark:text-gray-300">Rencana:</strong> {{ item.action_plan }}
                </p>
                <p v-else class="text-[11px] text-amber-600 dark:text-amber-400 font-medium italic mt-0.5">
                  Belum ada rencana tindak lanjut
                </p>
              </td>
              <td class="py-3 px-3.5 whitespace-nowrap">
                <UBadge
                  :color="getFindingCategoryColor(item.finding_category)"
                  variant="subtle"
                  size="xs"
                  class="font-semibold"
                >
                  {{ item.finding_category || 'Significant' }}
                </UBadge>
              </td>
              <td class="py-3 px-3.5 whitespace-nowrap">
                <div class="flex items-center gap-1.5 text-gray-900 dark:text-white font-medium">
                  <UIcon name="i-lucide-user" class="size-3.5 text-gray-400 shrink-0" />
                  <span class="line-clamp-1">{{ item.pic_name || 'PIC Unit' }}</span>
                </div>
                <div class="text-[11px] text-gray-500 dark:text-gray-400 flex items-center gap-1 mt-0.5">
                  <UIcon name="i-lucide-calendar" class="size-3 text-gray-400 shrink-0" />
                  <span>Due: {{ formatAtrDisplayDate(item.due_date) }}</span>
                </div>
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <div class="w-24 mx-auto space-y-1">
                  <div class="flex items-center justify-between text-[11px] font-bold">
                    <span :class="getProgressColorText(item.progress)">{{ item.progress || 0 }}%</span>
                  </div>
                  <div class="w-full bg-gray-200 dark:bg-gray-700 h-1.5 rounded-full overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="getProgressColorBg(item.progress)"
                      :style="{ width: `${item.progress || 0}%` }"
                    />
                  </div>
                </div>
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <AtrStatusBadge
                  :status="item.status"
                  :is-overdue="item.is_overdue"
                  :overdue-days="item.overdue_days"
                  size="sm"
                />
              </td>
              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <div class="flex items-center justify-center gap-1.5">
                  <UButton
                    size="xs"
                    color="primary"
                    variant="soft"
                    icon="i-lucide-edit-3"
                    class="font-bold cursor-pointer"
                    title="Perbarui Rencana & Progres Tindak Lanjut"
                    @click="atrStore.openActionPlan(item)"
                  >
                    Tindak Lanjut
                  </UButton>
                  <UButton
                    size="xs"
                    variant="ghost"
                    color="neutral"
                    icon="i-lucide-eye"
                    title="Lihat Detail Temuan"
                    @click="atrStore.openDetail(item)"
                  />
                </div>
              </td>
            </tr>

            <tr v-if="displayedAtrList.length === 0">
              <td colspan="7" class="py-12 text-center text-gray-500 dark:text-gray-400">
                <UIcon name="i-lucide-clipboard-check" class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
                <p class="text-xs font-semibold">Tidak ada data tindak lanjut temuan audit (ATR) yang sesuai filter.</p>
                <p class="text-[11px] text-gray-400 mt-0.5">Semua rekomendasi audit telah selesai atau belum ada temuan baru untuk unit kerja Anda.</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 6. WIDGET 3: PROFIL RISIKO DEPARTEMEN SENDIRI (OWN DEPARTMENT RISK SUMMARY - ROW 3-1) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-shield-alert" class="size-5 text-indigo-500" />
            3. Profil Risiko Terdaftar Unit Kerja (Own Dept. Risk Profile)
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Daftar register risiko operasional khusus untuk departemen Anda (Hak Akses Auditee: EDIT — Own Dept. per Matriks RBAC).
          </p>
        </div>

        <NuxtLink
          to="/risk-profile"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Kelola di Corporate Risk Profile</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div
          v-for="(risk, idx) in departmentRisks"
          :key="idx"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 text-xs space-y-2 hover:border-indigo-500/40 transition-all"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-[11px]">
              {{ risk.code || `RSK-${idx + 1}` }}
            </span>
            <UBadge
              :color="risk.level === 'High' ? 'error' : (risk.level === 'Moderate' ? 'warning' : 'info')"
              variant="subtle"
              size="xs"
              class="font-bold"
            >
              {{ risk.level }} Risk
            </UBadge>
          </div>
          <h4 class="font-bold text-gray-900 dark:text-white line-clamp-1">
            {{ risk.name || 'Risiko Operasional Unit' }}
          </h4>
          <p class="text-[11px] text-gray-600 dark:text-gray-300 line-clamp-2">
            <strong class="text-gray-700 dark:text-gray-200">Kontrol:</strong>
            {{ risk.control || 'Penerapan SOP dan pengendalian internal berkala.' }}
          </p>
        </div>

        <div v-if="departmentRisks.length === 0" class="col-span-full py-8 text-center text-gray-500 dark:text-gray-400">
          <UIcon name="i-lucide-shield-check" class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
          <p class="text-xs font-semibold">Belum ada data risiko terdaftar untuk departemen {{ currentDepartment }}.</p>
          <p class="text-[11px] text-gray-400 mt-0.5">Tambahkan register risiko operasional di modul Corporate Risk Profile.</p>
        </div>
      </div>
    </div>

    <!-- Official Integrated Modals for ATR Action Plan & Detail -->
    <ATRActionPlanForm />
    <ATRDetail />
    <AuditFieldworkDocumentModal />

    <!-- Dedicated Quick Modal for Auditee PBC Upload -->
    <UModal v-model:open="showQuickPbcModal" :ui="{ content: 'sm:max-w-md' }">
      <template #content>
        <div class="p-5 sm:p-6 space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-gray-100 dark:border-gray-800">
            <div class="flex items-center gap-2">
              <div class="p-2 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                <UIcon name="i-lucide-upload-cloud" class="size-5" />
              </div>
              <div>
                <h3 class="text-base font-bold text-gray-900 dark:text-white">Unggah Dokumen Bukti (PBC)</h3>
                <p class="text-xs text-gray-500">{{ activePbcTarget?.documentName || 'Dokumen Permintaan' }}</p>
              </div>
            </div>
            <UButton icon="i-lucide-x" color="neutral" variant="ghost" size="sm" @click="showQuickPbcModal = false" />
          </div>

          <div class="space-y-3 text-xs">
            <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700 space-y-1">
              <span class="text-gray-500 text-[11px] block">Keterangan dari Auditor:</span>
              <p class="text-gray-800 dark:text-gray-200 font-medium">
                {{ activePbcTarget?.description || 'Mohon sediakan berkas pendukung sesuai kriteria pemeriksaan.' }}
              </p>
            </div>

            <div>
              <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
                Pilih Berkas Dokumen (PDF, Excel, Word, Gambar - Maks. 10MB) *
              </label>
              <input
                type="file"
                class="w-full text-xs file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-primary-50 file:text-primary-700 hover:file:bg-primary-100 dark:file:bg-primary-950 dark:file:text-primary-300 border border-gray-200 dark:border-gray-700 rounded-xl p-1 cursor-pointer"
                @change="handlePbcFileSelected"
              />
              <p v-if="pbcSelectedFile" class="text-emerald-600 dark:text-emerald-400 text-[11px] mt-1 flex items-center gap-1">
                <UIcon name="i-lucide-check-circle" class="size-3.5" />
                Berkas siap diunggah: <strong>{{ pbcSelectedFile.name }}</strong> ({{ Math.round(pbcSelectedFile.size / 1024) }} KB)
              </p>
            </div>
          </div>

          <div class="pt-3 border-t border-gray-100 dark:border-gray-800 flex items-center justify-end gap-2">
            <UButton
              variant="outline"
              color="neutral"
              size="sm"
              @click="showQuickPbcModal = false"
            >
              Batal
            </UButton>
            <UButton
              color="primary"
              size="sm"
              icon="i-lucide-upload"
              class="font-bold cursor-pointer"
              :loading="isUploadingPbc"
              :disabled="!pbcSelectedFile"
              @click="submitPbcUpload"
            >
              Kirim Dokumen Bukti
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditeeSurveyStore } from '~/stores/auditee-survey'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import ATRActionPlanForm from '~/components/action-taken-report/ATRActionPlanForm.vue'
import ATRDetail from '~/components/action-taken-report/ATRDetail.vue'
import AuditFieldworkDocumentModal from '~/components/audit-fieldwork/AuditFieldworkDocumentModal.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'

const router = useRouter()
const toast = useToastNotification()
const authStore = useAuthStore()
const { isAdmin, canManageAudits } = useRbac()

const atrStore = useActionTakenReportStore()
const fieldworkStore = useAuditFieldworkStore()
const assignmentLetterStore = useAssignmentLetterStore()
const surveyStore = useAuditeeSurveyStore()
const auditResultStore = useAuditResultReportStore()
const riskProfileStore = useRiskProfileStore()

const isSyncing = ref(false)
const selectedLetterNumber = ref<string>('')
const selectedAtrFilter = ref<'all' | 'active' | 'in_progress' | 'overdue' | 'closed'>('all')
const atrSearchQuery = ref('')

// Quick PBC Upload State
const showQuickPbcModal = ref(false)
const activePbcTarget = ref<any>(null)
const pbcSelectedFile = ref<File | null>(null)
const isUploadingPbc = ref(false)

let isMounted = false

onBeforeUnmount(() => {
  isMounted = false
})

// Current User & Unit Details
const currentUserFullName = computed(() => {
  return authStore.user?.fullName || authStore.user?.username || 'PIC Unit Kerja'
})

const currentDepartment = computed(() => {
  return authStore.user?.department || 'Operasional'
})



const currentYear = computed(() => new Date().getFullYear())

// --- SATISFACTION SURVEY ALERT LOGIC ---
const hasPendingSurvey = computed(() => {
  // Check if there are finalized / published audit reports
  const completedReports = auditResultStore.reportList.filter(
    (r: any) => r.status === 'Final' || r.status === 'Published'
  )
  if (completedReports.length === 0) return false

  // If department has already filled at least one survey for recent audits
  const dept = currentDepartment.value.toLowerCase().trim()
  const deptSurveys = surveyStore.surveys.filter((s: any) => {
    return (s.department || '').toLowerCase().trim() === dept
  })

  // Show banner if there are completed reports and department hasn't submitted enough surveys
  return completedReports.length > deptSurveys.length
})

// --- PBC DOCUMENT COLLECTION (FIELDWORK STORE) ---
const publishedAssignmentLetters = computed(() => {
  const published = assignmentLetterStore.assignmentLetterList.filter((st: any) => st.status === 'Published')
  return published.length > 0 ? published : assignmentLetterStore.assignmentLetterList
})

const assignmentLetterOptions = computed(() => {
  return publishedAssignmentLetters.value.map((st: any) => ({
    label: `${st.letterNumber} - ${st.auditTitle}`,
    value: st.letterNumber
  }))
})

// PBC documents for selected letter
const pbcItemsList = computed(() => {
  const targetLetter = selectedLetterNumber.value
  if (!targetLetter) return []

  const docList = fieldworkStore.fieldworkData?.[targetLetter]?.documents ||
                  (fieldworkStore.selectedAssignmentLetter === targetLetter ? fieldworkStore.documents : [])

  if (!Array.isArray(docList)) return []

  return docList.map((doc: any) => {
    const isCollected = !!(doc.file || doc.fileName || doc.filePath || doc.fileUrl)
    return {
      id: doc.id,
      documentName: doc.documentName || 'Dokumen Permintaan Bukti',
      description: doc.description || '',
      requiredDate: doc.requiredDate || '',
      fileName: doc.fileName || doc.file?.name || '',
      file: doc.file,
      filePath: doc.filePath || doc.fileUrl,
      status: isCollected ? 'Received' : 'Pending'
    }
  })
})

const pbcStats = computed(() => {
  const total = pbcItemsList.value.length
  const collected = pbcItemsList.value.filter(item => item.status === 'Received').length
  const pending = total - collected
  return { total, collected, pending }
})

const pendingPbcCount = computed(() => pbcStats.value.pending)
const collectedPbcCount = computed(() => pbcStats.value.collected)

// --- ACTION TAKEN REPORT / CAPA ---
const allAtrItems = computed(() => {
  const list = atrStore.summaryItems.length > 0 ? atrStore.summaryItems : atrStore.items
  return list || []
})

const filteredAtrListTotal = computed(() => allAtrItems.value.length)

const openFindingsCount = computed(() => {
  return allAtrItems.value.filter((i: any) => {
    const s = String(i.status || '').toUpperCase()
    return s === 'OPEN' || s === 'ASSIGNED' || s === 'PENDING_PLAN'
  }).length
})

const pendingPlanCount = computed(() => {
  return allAtrItems.value.filter((i: any) => !i.action_plan || i.action_plan.trim() === '').length
})

const inProgressFindingsCount = computed(() => {
  return allAtrItems.value.filter((i: any) => {
    const s = String(i.status || '').toUpperCase()
    return s === 'IN_PROGRESS' || s === 'PENDING_REVIEW'
  }).length
})

const overdueFindingsCount = computed(() => {
  return allAtrItems.value.filter((i: any) => i.is_overdue || String(i.status || '').toUpperCase() === 'OVERDUE').length
})

const closedFindingsCount = computed(() => {
  return allAtrItems.value.filter((i: any) => String(i.status || '').toUpperCase() === 'CLOSED').length
})

const avgProgressRate = computed(() => {
  const inProgressList = allAtrItems.value.filter((i: any) => String(i.status || '').toUpperCase() !== 'CLOSED')
  if (inProgressList.length === 0) return 100
  const total = inProgressList.reduce((acc: number, curr: any) => acc + (Number(curr.progress) || 0), 0)
  return Math.min(100, Math.round(total / inProgressList.length))
})

// Filtered and searched ATR list
const displayedAtrList = computed(() => {
  let list = [...allAtrItems.value]

  // Filter tab
  if (selectedAtrFilter.value === 'active') {
    list = list.filter((i: any) => {
      const s = String(i.status || '').toUpperCase()
      return s === 'OPEN' || s === 'ASSIGNED' || s === 'PENDING_PLAN'
    })
  } else if (selectedAtrFilter.value === 'in_progress') {
    list = list.filter((i: any) => {
      const s = String(i.status || '').toUpperCase()
      return s === 'IN_PROGRESS' || s === 'PENDING_REVIEW'
    })
  } else if (selectedAtrFilter.value === 'overdue') {
    list = list.filter((i: any) => i.is_overdue || String(i.status || '').toUpperCase() === 'OVERDUE')
  } else if (selectedAtrFilter.value === 'closed') {
    list = list.filter((i: any) => String(i.status || '').toUpperCase() === 'CLOSED')
  }

  // Search query
  const q = atrSearchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter((i: any) => {
      return (
        (i.finding_title || '').toLowerCase().includes(q) ||
        (i.report_number || '').toLowerCase().includes(q) ||
        (i.pic_name || '').toLowerCase().includes(q) ||
        (i.action_plan || '').toLowerCase().includes(q)
      )
    })
  }

  return list
})

// --- OWN DEPARTMENT RISK PROFILE (ROW 3-1: EDIT OWN DEPT) ---
const departmentRisks = computed(() => {
  const dept = currentDepartment.value.toLowerCase().trim()
  const rawList = riskProfileStore.risks || []
  const matches = rawList.filter((r: any) => {
    const rDept = (r.department || r.business_process || '').toLowerCase().trim()
    return rDept.includes(dept) || dept.includes(rDept)
  })

  // If department matches found, return up to 3; otherwise take first 3 from register
  const source = matches.length > 0 ? matches : rawList.slice(0, 3)

  return source.slice(0, 3).map((r: any) => {
    let level = 'Moderate'
    const score = Number(r.inherent_risk || r.total_weighted_score || 0)
    if (score >= 15 || r.impact >= 4) level = 'High'
    else if (score < 8 && r.impact <= 2) level = 'Low'

    return {
      code: r.code || r.risk_code || 'RSK-UNIT',
      risk_code: r.code || r.risk_code || 'RSK-UNIT',
      name: r.risk_event || r.name || 'Risiko Operasional Departemen',
      risk_event: r.risk_event || r.name || 'Risiko Operasional Departemen',
      level,
      control: r.control_description || r.existing_control || 'Standard Operating Procedure (SOP) unit.',
      control_description: r.control_description || r.existing_control || 'Standard Operating Procedure (SOP) unit.'
    }
  })
})

// --- QUICK ACTIONS & HANDLERS ---
const openPbcUploadModal = (doc: any) => {
  activePbcTarget.value = doc
  pbcSelectedFile.value = null
  showQuickPbcModal.value = true
}

const handlePbcFileSelected = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    if (file.size > 10 * 1024 * 1024) {
      toast.error('Ukuran berkas terlalu besar! Maksimal 10MB.')
      target.value = ''
      return
    }
    pbcSelectedFile.value = file
  }
}

const submitPbcUpload = async () => {
  if (!activePbcTarget.value || !pbcSelectedFile.value) return
  isUploadingPbc.value = true
  try {
    if (selectedLetterNumber.value) {
      fieldworkStore.selectedAssignmentLetter = selectedLetterNumber.value
    }
    fieldworkStore.documentForm.documentName = activePbcTarget.value.documentName
    fieldworkStore.documentForm.description = activePbcTarget.value.description || 'Diunggah oleh Auditee'
    fieldworkStore.documentForm.requiredDate = activePbcTarget.value.requiredDate || new Date().toISOString().split('T')[0]
    fieldworkStore.documentForm.file = pbcSelectedFile.value
    fieldworkStore.documentForm.fileName = pbcSelectedFile.value.name
    fieldworkStore.isEditingDocument = true
    fieldworkStore.editingDocumentId = activePbcTarget.value.id

    await fieldworkStore.saveDocument()
    showQuickPbcModal.value = false
    toast.success('Berkas bukti berhasil diunggah ke tim auditor!')
    if (selectedLetterNumber.value) {
      await fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value)
    }
  } catch (err) {
    console.error('Failed to upload PBC document:', err)
    toast.error('Gagal mengunggah berkas bukti')
  } finally {
    isUploadingPbc.value = false
  }
}

const downloadPbcDoc = (doc: any) => {
  fieldworkStore.downloadDocumentFile(doc)
}

// Helpers
const formatDisplayDate = (d?: string) => {
  if (!d) return 'Tidak ada tenggat'
  try {
    const date = new Date(d)
    if (isNaN(date.getTime())) return d
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return d
  }
}

const formatAtrDisplayDate = (d?: string | null) => {
  if (!d) return '-'
  try {
    const date = new Date(d)
    if (isNaN(date.getTime())) return d
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return d
  }
}

const getDeadlineLabel = (finishPeriod?: string) => {
  if (!finishPeriod) return 'Jadwal belum diatur'
  const end = new Date(finishPeriod)
  if (isNaN(end.getTime())) return '-'
  const now = new Date()
  const diffDays = Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diffDays < 0) return `${Math.abs(diffDays)} Hari Terlambat`
  if (diffDays === 0) return 'Hari Terakhir'
  return `${diffDays} Hari Tersisa`
}

const getDeadlineBadgeColor = (finishPeriod?: string): 'error' | 'warning' | 'success' | 'neutral' => {
  if (!finishPeriod) return 'neutral'
  const end = new Date(finishPeriod)
  if (isNaN(end.getTime())) return 'neutral'
  const now = new Date()
  const diffDays = Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  if (diffDays < 0) return 'error'
  if (diffDays <= 3) return 'error'
  if (diffDays <= 7) return 'warning'
  return 'success'
}

const getFindingCategoryColor = (cat?: string): 'error' | 'warning' | 'info' | 'neutral' => {
  if (cat === 'Very Significant') return 'error'
  if (cat === 'Significant') return 'warning'
  if (cat === 'Quite Significant') return 'info'
  return 'neutral'
}

const getProgressColorBg = (progress?: number) => {
  const p = Number(progress) || 0
  if (p >= 100) return 'bg-emerald-500'
  if (p >= 50) return 'bg-sky-500'
  if (p > 0) return 'bg-amber-500'
  return 'bg-gray-300 dark:bg-gray-600'
}

const getProgressColorText = (progress?: number) => {
  const p = Number(progress) || 0
  if (p >= 100) return 'text-emerald-600 dark:text-emerald-400'
  if (p >= 50) return 'text-sky-600 dark:text-sky-400'
  if (p > 0) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-400'
}

// Initial Data Loader (Silent on Mount)
const loadInitialAuditeeData = async () => {
  try {
    await Promise.allSettled([
      atrStore.fetchSummary(),
      atrStore.fetchReports(),
      assignmentLetterStore.fetchAssignmentLetters(),
      surveyStore.fetchSurveys(),
      auditResultStore.fetchReports(),
      riskProfileStore.fetchRisks()
    ])
    if (!isMounted) return

    if (!selectedLetterNumber.value && publishedAssignmentLetters.value[0]) {
      selectedLetterNumber.value = publishedAssignmentLetters.value[0].letterNumber
    }

    if (selectedLetterNumber.value) {
      fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value).catch(() => {})
    }
  } catch (err) {
    console.warn('Silent auditee dashboard data load notice:', err)
  }
}

// Manual Sync (Button Click with Toast)
const manualSyncAuditeeData = async () => {
  isSyncing.value = true
  try {
    await Promise.allSettled([
      atrStore.fetchSummary(),
      atrStore.fetchReports(),
      assignmentLetterStore.fetchAssignmentLetters(),
      surveyStore.fetchSurveys(),
      auditResultStore.fetchReports(),
      riskProfileStore.fetchRisks()
    ])
    if (!isMounted) return
    if (selectedLetterNumber.value) {
      await fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value, true)
    }
    toast.success('Data operasional unit kerja berhasil disinkronkan!')
  } catch (err) {
    if (isMounted) toast.error('Gagal memperbarui data auditee')
  } finally {
    if (isMounted) isSyncing.value = false
  }
}

// Watch selected letter change to load fieldwork documents (PBC)
watch(
  () => selectedLetterNumber.value,
  (rawVal) => {
    if (!isMounted) return
    const newVal = typeof rawVal === 'string' ? rawVal : (rawVal as any)?.value || ''
    if (newVal && fieldworkStore.selectedAssignmentLetter !== newVal) {
      fieldworkStore.selectedAssignmentLetter = newVal
    }
  }
)

// Auto-select first letter when list is loaded
watch(
  () => publishedAssignmentLetters.value,
  (list) => {
    if (!isMounted) return
    if (!selectedLetterNumber.value && list && list[0]) {
      selectedLetterNumber.value = list[0].letterNumber
    }
  },
  { immediate: true }
)

onMounted(async () => {
  isMounted = true
  await loadInitialAuditeeData()
})
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
