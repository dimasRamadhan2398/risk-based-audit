<template>
  <div class="space-y-6 sm:space-y-8 animate-fade-in">
    <!-- Header: Auditor Profile & Context Bar -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 p-5 sm:p-6 bg-gradient-to-r from-emerald-900/10 via-primary-500/5 to-transparent border border-emerald-500/20 rounded-2xl shadow-xs">
      <div class="flex items-start sm:items-center gap-4">
        <div class="size-14 rounded-2xl bg-emerald-600/10 dark:bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400 shrink-0 shadow-xs">
          <UIcon name="i-lucide-clipboard-check" class="size-7" />
        </div>
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">
              Dashboard Pelaksana Audit (Fieldwork Auditor)
            </h1>
            <UBadge color="success" variant="subtle" size="sm" class="font-bold uppercase tracking-wider">
              Role: Auditor
            </UBadge>
          </div>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mt-1 max-w-3xl">
            Selamat datang, <strong class="text-gray-900 dark:text-gray-200">{{ currentUserFullName }}</strong>. 
            Layar kerja khusus eksekusi penugasan audit lapangan, penyusunan Kertas Kerja Audit dan perumusan draft temuan audit (LHA).
          </p>
        </div>
      </div>
    </div>

    <!-- Notice Card: Fokus Operasional Lapangan (Sesuai Panduan RBAC) -->
    <div class="flex items-center justify-between gap-4 p-4 rounded-xl bg-sky-500/10 border border-sky-500/20 text-sky-900 dark:text-sky-200 text-xs sm:text-sm">
      <div class="flex items-center gap-3">
        <UIcon name="i-lucide-info" class="size-5 text-sky-600 dark:text-sky-400 shrink-0" />
        <div>
          <span class="font-bold">Prinsip Tampilan Auditor:</span>
          <span class="ml-1 text-sky-800 dark:text-sky-300">
            Metrik risiko makro korporasi disederhanakan agar workspace berfokus penuh pada penugasan aktif, kertas kerja, dan dokumen bukti (PBC).
          </span>
        </div>
      </div>
      <div class="hidden md:flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-md bg-sky-600/10 text-sky-700 dark:text-sky-300 shrink-0">
        <UIcon name="i-lucide-shield-check" class="size-3.5" />
        <span>Matriks RBAC Baris 8 & 9</span>
      </div>
    </div>

    <!-- 4 Focused Operational Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Card 1: Active Fieldwork Assignments -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-emerald-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Audit Lapangan Aktif
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-gray-900 dark:text-white mt-1">
              {{ activeAssignmentsCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <UIcon name="i-lucide-briefcase" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
            <UIcon name="i-lucide-activity" class="size-3.5" />
            {{ inProgressAssignmentsCount }} Dalam Pelaksanaan
          </span>
          <span class="text-gray-400">Tahun {{ currentYear }}</span>
        </div>
      </div>

      <!-- Card 2: Working Papers (KKA) Progress -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-primary-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Kertas Kerja
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-gray-900 dark:text-white mt-1">
              {{ totalWorkingPapersCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-primary-50 dark:bg-primary-500/10 text-primary-600 dark:text-primary-400">
            <UIcon name="i-lucide-file-signature" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <div class="flex items-center gap-1.5">
            <div class="w-16 bg-gray-200 dark:bg-gray-800 h-2 rounded-full overflow-hidden">
              <div class="bg-primary-500 h-full rounded-full" :style="{ width: `${kkaCompletionRate}%` }" />
            </div>
            <span class="font-bold text-primary-600 dark:text-primary-400">{{ kkaCompletionRate }}% Selesai</span>
          </div>
          <NuxtLink to="/working-paper" class="text-primary-600 dark:text-primary-400 hover:underline font-semibold">
            Kelola &rarr;
          </NuxtLink>
        </div>
      </div>

      <!-- Card 3: Draft Findings (LHA) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-amber-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Draft Temuan Sementara
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-amber-600 dark:text-amber-400 mt-1">
              {{ draftFindingsList.length }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-amber-50 dark:bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <UIcon name="i-lucide-alert-triangle" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <div class="flex items-center gap-1 text-[11px]">
            <span class="px-1.5 py-0.5 rounded bg-red-100 dark:bg-red-950/60 text-red-700 dark:text-red-300 font-semibold">{{ highFindingsCount }} High</span>
            <span class="px-1.5 py-0.5 rounded bg-amber-100 dark:bg-amber-950/60 text-amber-700 dark:text-amber-300 font-semibold">{{ mediumFindingsCount }} Med</span>
            <span class="px-1.5 py-0.5 rounded bg-blue-100 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 font-semibold">{{ lowFindingsCount }} Low</span>
          </div>
          <span class="text-gray-400 text-[11px]">Draft LHA</span>
        </div>
      </div>

      <!-- Card 4: PBC & Evidence Tracker -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-sky-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Bukti Klien (PBC Tracker)
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-sky-600 dark:text-sky-400 mt-1">
              {{ pbcStats.collected }}/{{ pbcStats.total }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-sky-50 dark:bg-sky-500/10 text-sky-600 dark:text-sky-400">
            <UIcon name="i-lucide-folder-input" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-sky-600 dark:text-sky-400 font-semibold">
            {{ pbcStats.pending }} Dokumen Menunggu Respon
          </span>
          <NuxtLink to="/audit-fieldwork" class="text-sky-600 dark:text-sky-400 hover:underline font-semibold">
            Lihat &rarr;
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- KOMPONEN UTAMA 1: TABEL DAFTAR AUDIT BERJALAN (ACTIVE FIELDWORK) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-5">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-briefcase" class="size-5 text-emerald-600 dark:text-emerald-400" />
            1. Tabel Utama: Daftar Audit Berjalan (Active Fieldwork)
          </h2>
          <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Daftar penugasan audit yang sedang dipegang, timeline tahapan lapangan, dan status progres Kertas Kerja.
          </p>
        </div>

        <div class="flex items-center gap-2.5 w-full sm:w-auto">
          <UInput
            v-model="fieldworkSearch"
            icon="i-lucide-search"
            placeholder="Cari surat tugas / unit..."
            size="sm"
            class="w-full sm:w-64"
          />
          <UButton
            to="/assignment-letter"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-external-link"
            class="shrink-0 font-medium"
          >
            Semua Penugasan
          </UButton>
        </div>
      </div>

      <!-- Table Content -->
      <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
        <table class="w-full text-left text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-700 dark:text-gray-200 border-b border-gray-200 dark:border-gray-700 text-xs uppercase font-bold tracking-wider">
            <tr>
              <th class="py-3 px-4 min-w-[200px]">Surat Tugas & Judul Audit</th>
              <th class="py-3 px-4 min-w-[130px]">Unit Kerja (Auditee)</th>
              <th class="py-3 px-4 min-w-[170px]">Periode Lapangan</th>
              <th class="py-3 px-4 min-w-[240px]">Tahapan Lapangan</th>
              <th class="py-3 px-4 min-w-[120px] text-center">Kelengkapan Kertas Kerja</th>
              <th class="py-3 px-4 min-w-[120px] text-center">Status</th>
              <th class="py-3 px-4 min-w-[160px] text-center">Aksi Operasional</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
            <tr
              v-for="item in filteredFieldworkList"
              :key="item.letterNumber"
              class="hover:bg-emerald-50/30 dark:hover:bg-emerald-950/10 transition-colors"
              :class="{ 'bg-emerald-50/20 dark:bg-emerald-950/20': selectedLetterNumber === item.letterNumber }"
            >
              <!-- Letter Number & Title -->
              <td class="py-3.5 px-4">
                <div class="flex items-start gap-2.5">
                  <div class="size-8 rounded-lg bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0 mt-0.5">
                    <UIcon name="i-lucide-file-text" class="size-4" />
                  </div>
                  <div>
                    <span class="font-mono font-bold text-xs text-primary-600 dark:text-primary-400 block">
                      {{ item.letterNumber }}
                    </span>
                    <span class="font-semibold text-gray-900 dark:text-gray-100 text-xs sm:text-sm line-clamp-1">
                      {{ item.auditTitle }}
                    </span>
                    <span class="text-[11px] text-gray-500 dark:text-gray-400">
                      Tim: {{ item.auditTeam || 'SKAI' }} &bull; {{ item.category || 'Assurance' }}
                    </span>
                  </div>
                </div>
              </td>

              <!-- Auditee Unit -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-1.5 font-medium text-xs text-gray-800 dark:text-gray-200">
                  <UIcon name="i-lucide-building" class="size-3.5 text-gray-400 shrink-0" />
                  <span>{{ item.workingUnit }}</span>
                </div>
              </td>

              <!-- Execution Period -->
              <td class="py-3.5 px-4 text-xs">
                <div class="space-y-1">
                  <div class="flex items-center gap-1 text-gray-700 dark:text-gray-300">
                    <UIcon name="i-lucide-calendar" class="size-3.5 text-gray-400" />
                    <span>{{ formatPeriod(item.startPeriod, item.finishPeriod) }}</span>
                  </div>
                  <UBadge
                    :color="getDeadlineBadgeColor(item.finishPeriod)"
                    variant="soft"
                    size="sm"
                    class="text-[10px] font-semibold"
                  >
                    {{ getDeadlineLabel(item.finishPeriod) }}
                  </UBadge>
                </div>
              </td>

              <!-- Fieldwork Milestones Stepper -->
              <td class="py-3.5 px-4">
                <div class="space-y-1.5">
                  <div class="flex items-center justify-between text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                    <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 1 }">1. Entry</span>
                    <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 2 }">2. Pengujian</span>
                    <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 3 }">3. Kertas Kerja</span>
                    <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 4 }">4. Exit</span>
                  </div>
                  <!-- Progress Tracker Bar -->
                  <div class="w-full h-1.5 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden flex">
                    <div
                      class="bg-emerald-500 h-full transition-all duration-500"
                      :style="{ width: `${(item.stageStep / 4) * 100}%` }"
                    />
                  </div>
                  <div class="text-[11px] text-gray-500 dark:text-gray-400">
                    Tahap saat ini: <strong class="text-gray-800 dark:text-gray-200">{{ item.currentStageName }}</strong>
                  </div>
                </div>
              </td>

              <!-- KKA Progress Bar -->
              <td class="py-3.5 px-4 text-center">
                <div class="inline-flex flex-col items-center gap-1">
                  <span class="font-bold text-xs text-gray-900 dark:text-gray-100">
                    {{ item.kkaCompleted }}/5 Form
                  </span>
                  <div class="w-16 h-1.5 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all"
                      :class="item.kkaCompleted === 5 ? 'bg-emerald-500' : 'bg-primary-500'"
                      :style="{ width: `${(item.kkaCompleted / 5) * 100}%` }"
                    />
                  </div>
                </div>
              </td>

              <!-- Status Badge -->
              <td class="py-3.5 px-4 text-center">
                <UBadge
                  :color="getFieldworkStatusColor(item.status)"
                  variant="subtle"
                  size="sm"
                  class="font-semibold"
                >
                  {{ item.status || 'In Progress' }}
                </UBadge>
              </td>

              <!-- Operational Actions -->
              <td class="py-3.5 px-4 text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <UButton
                    size="xs"
                    color="primary"
                    variant="solid"
                    icon="i-lucide-file-signature"
                    class="font-bold cursor-pointer"
                    title="Buka Kertas Kerja"
                    @click="goToWorkingPaper(item.letterNumber)"
                  >
                    Kertas Kerja
                  </UButton>
                  <UButton
                    size="xs"
                    color="neutral"
                    variant="outline"
                    icon="i-lucide-briefcase"
                    class="font-medium cursor-pointer"
                    title="Buka Pengujian Lapangan & PBC"
                    @click="goToFieldwork(item.letterNumber)"
                  >
                    Uji
                  </UButton>
                  <UButton
                    size="xs"
                    color="neutral"
                    variant="ghost"
                    icon="i-lucide-file-text"
                    class="cursor-pointer"
                    title="Draft Laporan Hasil Audit (LHA)"
                    @click="goToResultReport(item.letterNumber)"
                  />
                </div>
              </td>
            </tr>

            <tr v-if="filteredFieldworkList.length === 0">
              <td colspan="7" class="py-8 text-center text-gray-500 dark:text-gray-400">
                <UIcon name="i-lucide-folder-open" class="size-10 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
                <p class="font-medium">Tidak ada penugasan audit berjalan yang cocok dengan pencarian.</p>
                <p class="text-xs text-gray-400 mt-1">Pastikan Surat Tugas telah berstatus Published di modul Assignment Letter.</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- KOMPONEN UTAMA 2: KERTAS KERJA & TEMPLATE (KKA F-01 S/D F-05) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-6">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-file-signature" class="size-5 text-primary-600 dark:text-primary-400" />
            2. Kertas Kerja & Template
          </h2>
          <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            Akses cepat pembuatan atau kelanjutan formulir kertas kerja pengujian audit dan upload kertas kerja pendukung.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2 w-full sm:w-auto">
          <!-- Assignment Letter Selector for KKA -->
          <div class="w-full sm:w-64">
            <USelectMenu
              v-model="selectedLetterNumber"
              :items="assignmentLetterOptions"
              value-key="value"
              placeholder="Pilih Surat Tugas..."
              class="w-full"
            >
              <template #leading>
                <UIcon name="i-lucide-file-text" class="size-4 text-primary-500" />
              </template>
            </USelectMenu>
          </div>

          <!-- Button Upload KKA (Sesuai Matriks RBAC Row 8-3: UPLOAD) -->
          <UButton
            to="/working-paper/upload"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-upload"
            class="font-semibold cursor-pointer shrink-0"
          >
            Upload Kertas Kerja Pendukung
          </UButton>
        </div>
      </div>

      <!-- Selected Audit Context Banner -->
      <div class="p-3.5 rounded-xl bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700/80 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
        <div class="flex items-center gap-2">
          <UIcon name="i-lucide-target" class="size-4 text-primary-500 shrink-0" />
          <span class="text-gray-600 dark:text-gray-400">Fokus Kertas Kerja Saat Ini:</span>
          <strong class="font-mono text-primary-600 dark:text-primary-400 text-xs sm:text-sm">{{ activeAssignmentData.letterNumber }}</strong>
          <span class="text-gray-500 dark:text-gray-400 hidden sm:inline">&mdash; {{ activeAssignmentData.auditTitle }}</span>
        </div>
        <div class="text-gray-500 dark:text-gray-400 text-right">
          Unit: <strong class="text-gray-800 dark:text-gray-200">{{ activeAssignmentData.workingUnit }}</strong>
        </div>
      </div>

      <!-- 5 Interactive KKA Cards (F-01 s/d F-05) -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <!-- F-01: Header & Latar Belakang -->
        <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-4 hover:border-primary-500 transition-all group">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <UBadge :color="kkaFormStatus.f01 ? 'success' : 'neutral'" variant="soft" size="xs">
                {{ kkaFormStatus.f01 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Header & Latar Belakang
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Proses bisnis, ruang lingkup, lokasi audit, serta tanggal mulai & selesai pengujian.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f01')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-02: Identifikasi Risiko & Pengendalian -->
        <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-4 hover:border-primary-500 transition-all group">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <UBadge :color="kkaFormStatus.f02 ? 'success' : 'neutral'" variant="soft" size="xs">
                {{ kkaFormStatus.f02 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Identifikasi Risiko & RCM
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Pemetaan taksonomi risiko, risk level, dan deskripsi sistem pengendalian internal terkait.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f02')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-03: Pengujian Sampel & Bukti -->
        <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-4 hover:border-primary-500 transition-all group">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <UBadge :color="kkaFormStatus.f03 ? 'success' : 'neutral'" variant="soft" size="xs">
                {{ kkaFormStatus.f03 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Uji Sampel & Bukti
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Populasi audit, penentuan sample size, pengujian substantif/kepatuhan, dan bukti dokumen.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f03')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-04: Analisis Sebab & Kondisi -->
        <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-4 hover:border-primary-500 transition-all group">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <UBadge :color="kkaFormStatus.f04 ? 'success' : 'neutral'" variant="soft" size="xs">
                {{ kkaFormStatus.f04 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Sebab & Kondisi (RCA)
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Kondisi riil vs kriteria standar, analisis akar masalah (5-Whys/Fishbone), dan dampak risiko.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f04')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-05: Rencana Aksi & Rekomendasi -->
        <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-4 hover:border-primary-500 transition-all group">
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <UBadge :color="kkaFormStatus.f05 ? 'success' : 'neutral'" variant="soft" size="xs">
                {{ kkaFormStatus.f05 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Rekomendasi & Rencana Aksi
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Rekomendasi perbaikan auditor, respons auditee, penunjukan PIC, dan target penyelesaian.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f05')"
          >
            Buka Form &rarr;
          </UButton>
        </div>
      </div>
    </div>

    <!-- 2 COLUMN LAYOUT: WIDGET DRAFT TEMUAN & PBC TRACKER -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- KOMPONEN UTAMA 3: WIDGET DRAFT TEMUAN & LHA (7 COLS) -->
      <div class="lg:col-span-7 p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-5">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-alert-triangle" class="size-5 text-amber-500" />
              3. Widget Draft Temuan & LHA
            </h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Ringkasan temuan sementara yang sedang disusun sebelum finalisasi laporan.
            </p>
          </div>

          <NuxtLink
            to="/audit-result-report"
            class="text-xs font-bold text-primary-600 dark:text-primary-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
          >
            <span>Buka Modul LHA (Draft)</span>
            <span>&rarr;</span>
          </NuxtLink>
        </div>

        <!-- Warning Callout: Wewenang Auditor (No Publish Button) -->
        <div class="p-3 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-900 dark:text-amber-200 text-xs flex items-start gap-2.5">
          <UIcon name="i-lucide-lock" class="size-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
          <div>
            <span class="font-bold">Otorisasi Laporan:</span>
            <span class="ml-1 text-amber-800 dark:text-amber-300">
              Hak akses Auditor adalah <strong>DRAFT</strong> (Row 9-1). Tombol finalisasi, otorisasi, dan publikasi LHA resmi merupakan wewenang Audit Manager & Chief Audit Executive (CAE).
            </span>
          </div>
        </div>

        <!-- Findings Category Pill Summary -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-xs">
          <div class="p-2.5 rounded-xl bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900/40">
            <span class="block text-lg font-extrabold text-red-600 dark:text-red-400">{{ highFindingsCount }}</span>
            <span class="text-red-700 dark:text-red-300 text-[11px] font-semibold">Sangat Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40">
            <span class="block text-lg font-extrabold text-amber-600 dark:text-amber-400">{{ mediumFindingsCount }}</span>
            <span class="text-amber-700 dark:text-amber-300 text-[11px] font-semibold">Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40">
            <span class="block text-lg font-extrabold text-blue-600 dark:text-blue-400">{{ lowFindingsCount }}</span>
            <span class="text-blue-700 dark:text-blue-300 text-[11px] font-semibold">Cukup Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
            <span class="block text-lg font-extrabold text-gray-700 dark:text-gray-300">{{ notSignificantCount }}</span>
            <span class="text-gray-600 dark:text-gray-400 text-[11px] font-semibold">Tidak Signifikan</span>
          </div>
        </div>

        <!-- Draft Findings Table -->
        <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
          <table class="w-full text-left text-xs">
            <thead class="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300 uppercase font-bold text-[11px]">
              <tr>
                <th class="py-2.5 px-3">Judul Temuan Sementara</th>
                <th class="py-2.5 px-3">Kategori</th>
                <th class="py-2.5 px-3">Sumber Kertas Kerja</th>
                <th class="py-2.5 px-3 text-center">Status</th>
                <th class="py-2.5 px-3 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
              <tr
                v-for="(finding, idx) in draftFindingsList"
                :key="idx"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors"
              >
                <td class="py-2.5 px-3">
                  <span class="font-semibold text-gray-900 dark:text-white block line-clamp-1">
                    {{ finding.title }}
                  </span>
                  <span class="text-[10px] text-gray-500 dark:text-gray-400">
                    {{ finding.assignmentLetterId || activeAssignmentData.letterNumber }}
                  </span>
                </td>
                <td class="py-2.5 px-3">
                  <UBadge :color="getFindingCategoryColor(finding.category)" variant="subtle" size="xs">
                    {{ finding.category }}
                  </UBadge>
                </td>
                <td class="py-2.5 px-3 text-gray-600 dark:text-gray-400">
                  {{ finding.source || 'Test Controls' }}
                </td>
                <td class="py-2.5 px-3 text-center">
                  <UBadge color="neutral" variant="soft" size="xs">
                    Draft Auditor
                  </UBadge>
                </td>
                <td class="py-2.5 px-3 text-center">
                  <UButton
                    size="xs"
                    color="primary"
                    variant="ghost"
                    icon="i-lucide-pencil"
                    title="Edit Draft Temuan"
                    @click="goToResultReport(finding.assignmentLetterId || activeAssignmentData.letterNumber)"
                  />
                </td>
              </tr>

              <tr v-if="draftFindingsList.length === 0">
                <td colspan="5" class="py-6 text-center text-gray-500 dark:text-gray-400">
                  Belum ada draft temuan untuk penugasan ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex items-center justify-between pt-2">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            Draft siap diajukan ke review supervisor/manager saat Kertas Kerja selesai.
          </span>
          <UButton
            to="/audit-result-report"
            color="primary"
            variant="outline"
            size="sm"
            icon="i-lucide-file-text"
            class="font-semibold"
          >
            Lanjutkan Penyusunan LHA (Draft)
          </UButton>
        </div>
      </div>

      <!-- KOMPONEN PENDUKUNG: PBC TRACKER & DOKUMEN BUKTI (5 COLS) -->
      <div class="lg:col-span-5 p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-5">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-folder-check" class="size-5 text-sky-500" />
              4. Permintaan Bukti Klien (PBC)
            </h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Monitoring dokumen dan data yang diminta ke Auditee.
            </p>
          </div>

          <NuxtLink
            to="/audit-fieldwork"
            class="text-xs font-bold text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1"
          >
            <span>Fieldwork</span>
            <span>&rarr;</span>
          </NuxtLink>
        </div>

        <!-- PBC Summary Stats -->
        <div class="grid grid-cols-3 gap-2 text-center text-xs">
          <div class="p-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
            <span class="block text-base font-bold text-gray-900 dark:text-white">{{ pbcStats.total }}</span>
            <span class="text-[10px] text-gray-500">Diminta</span>
          </div>
          <div class="p-2 rounded-xl bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/40">
            <span class="block text-base font-bold text-emerald-600 dark:text-emerald-400">{{ pbcStats.collected }}</span>
            <span class="text-[10px] text-emerald-700 dark:text-emerald-300 font-semibold">Diterima</span>
          </div>
          <div class="p-2 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40">
            <span class="block text-base font-bold text-amber-600 dark:text-amber-400">{{ pbcStats.pending }}</span>
            <span class="text-[10px] text-amber-700 dark:text-amber-300 font-semibold">Menunggu</span>
          </div>
        </div>

        <!-- PBC Item List -->
        <div class="space-y-2.5 max-h-[290px] overflow-y-auto pr-1">
          <div
            v-for="(pbc, idx) in pbcItemsList"
            :key="idx"
            class="p-3 rounded-xl border border-gray-100 dark:border-gray-800 bg-gray-50/60 dark:bg-gray-800/40 flex items-start justify-between gap-3 text-xs"
          >
            <div class="min-w-0">
              <span class="font-semibold text-gray-900 dark:text-white block line-clamp-1">
                {{ pbc.documentName }}
              </span>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1">
                {{ pbc.description || 'Tidak ada keterangan tambahan' }}
              </p>
              <span class="text-[10px] text-gray-400 mt-1 block">
                Deadline: {{ formatDisplayDate(pbc.requiredDate) }}
              </span>
            </div>

            <UBadge
              :color="pbc.status === 'Received' ? 'success' : 'warning'"
              variant="subtle"
              size="xs"
              class="shrink-0 font-semibold"
            >
              {{ pbc.status === 'Received' ? 'Diterima' : 'Menunggu' }}
            </UBadge>
          </div>

          <div v-if="pbcItemsList.length === 0" class="py-8 text-center text-gray-500 dark:text-gray-400">
            <UIcon name="i-lucide-folder-open" class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
            <p class="text-xs font-medium">Belum ada daftar permintaan bukti (PBC) untuk penugasan ini.</p>
            <p class="text-[11px] text-gray-400 mt-0.5">Kelola permintaan dokumen di modul Audit Fieldwork.</p>
          </div>
        </div>

        <!-- Quick Link to Fieldwork Documents Tab -->
        <UButton
          to="/audit-fieldwork"
          variant="outline"
          color="neutral"
          size="sm"
          block
          icon="i-lucide-arrow-right"
          class="font-semibold cursor-pointer"
        >
          Kelola Permintaan Bukti di Audit Fieldwork
        </UButton>
      </div>
    </div>

    <!-- REFERENSI CEPAT: RCM & FAKTOR RISIKO (READ-ONLY) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-book-open" class="size-5 text-indigo-500" />
            5. Referensi Cepat: Risk Control Matrix (RCM) & Kriteria Audit (Read-Only)
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            Akses rujukan risiko dan standar kontrol (Hak Akses Auditor: READ & FULL RCM) untuk penyusunan kertas kerja pengujian.
          </p>
        </div>
        <NuxtLink
          to="/risk-profile/risk-control-matrix"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Buka Modul RCM Lengkap</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div
          v-for="(rcm, idx) in rcmQuickReferences"
          :key="idx"
          class="p-3.5 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 text-xs space-y-1.5"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono font-bold text-primary-600 dark:text-primary-400 text-[11px]">{{ rcm.code }}</span>
            <UBadge :color="rcm.riskLevel === 'High' ? 'error' : (rcm.riskLevel === 'Moderate' ? 'warning' : 'info')" variant="soft" size="xs">
              {{ rcm.riskLevel }} Risk
            </UBadge>
          </div>
          <h4 class="font-bold text-gray-900 dark:text-white line-clamp-1">
            {{ rcm.process }}
          </h4>
          <p class="text-[11px] text-gray-600 dark:text-gray-300 line-clamp-2">
            <strong class="text-gray-700 dark:text-gray-200">Kontrol:</strong> {{ rcm.controlDescription }}
          </p>
        </div>

        <div v-if="rcmQuickReferences.length === 0" class="col-span-full py-6 text-center text-gray-500 dark:text-gray-400">
          <UIcon name="i-lucide-book-open" class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2" />
          <p class="text-xs font-medium">Belum ada data Risk Control Matrix (RCM) yang tercatat.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useRCMStore } from '~/stores/rcm'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

const router = useRouter()
const toast = useToastNotification()
const authStore = useAuthStore()
const { isAdmin, canManageAudits } = useRbac()

const assignmentLetterStore = useAssignmentLetterStore()
const fieldworkStore = useAuditFieldworkStore()
const workingPaperStore = useWorkingPaperStore()
const auditResultStore = useAuditResultReportStore()
const rcmStore = useRCMStore()

const isSyncing = ref(false)
const fieldworkSearch = ref('')
const selectedLetterNumber = ref<string>('')

// Current User Info
const currentUserFullName = computed(() => {
  return authStore.user?.fullName || authStore.user?.username || 'Auditor SKAI'
})



const currentYear = computed(() => new Date().getFullYear())

// Helper to count how many KKA forms are completed for a specific assignment letter
const getCompletedKkaFormsCount = (letterNumber: string, letterId?: string) => {
  const isMatch = (itemLetter: any) => {
    return itemLetter === letterNumber || (letterId && itemLetter === letterId)
  }

  const hasF01 = workingPaperStore.dataF01.some((d: any) => isMatch(d.assignmentLetterId))
  const hasF02 = workingPaperStore.dataF02.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF03 = workingPaperStore.dataF03.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF04 = workingPaperStore.dataF04.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF05 = workingPaperStore.dataF05.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))

  return (hasF01 ? 1 : 0) + (hasF02 ? 1 : 0) + (hasF03 ? 1 : 0) + (hasF04 ? 1 : 0) + (hasF05 ? 1 : 0)
}

// 100% Real Live Fieldwork Assignments from Assignment Letter Store
const combinedFieldworkList = computed(() => {
  return assignmentLetterStore.assignmentLetterList
    .filter((st: any) => st.status === 'Published')
    .map((st: any) => {
      const kkaCompleted = getCompletedKkaFormsCount(st.letterNumber, st.id)

      // Determine real audit stage based on data lifecycle:
      // 1. Has draft/final LHA report -> Stage 4 (Exit / Reporting)
      // 2. Has Working Paper forms -> Stage 3 (Kertas Kerja)
      // 3. Has Fieldwork observations/interviews/samples/documents -> Stage 2 (Pengujian)
      // 4. Default -> Stage 1 (Entry / Persiapan)
      const hasReport = auditResultStore.reportList.some(
        (r: any) => r.assignmentLetterId === st.letterNumber || (st.id && r.assignmentLetterId === st.id)
      )

      const fwData = fieldworkStore.fieldworkData?.[st.letterNumber]
      const hasFieldworkData = !!(
        fwData?.interviews?.length ||
        fwData?.observations?.length ||
        fwData?.documents?.length ||
        fwData?.samples?.length ||
        fwData?.testControls?.length
      )

      let stageStep = 1
      let currentStageName = 'Entry Meeting & Persiapan'

      if (hasReport) {
        stageStep = 4
        currentStageName = 'Penyusunan LHA & Exit'
      } else if (kkaCompleted > 0) {
        stageStep = 3
        currentStageName = 'Penyusunan Kertas Kerja'
      } else if (hasFieldworkData) {
        stageStep = 2
        currentStageName = 'Pengujian Sampel & Bukti'
      }

      return {
        letterNumber: st.letterNumber,
        auditTitle: st.auditTitle,
        workingUnit: st.workingUnit || 'Unit Kerja',
        auditTeam: st.auditTeam || 'SKAI',
        category: st.category || 'Assurance',
        startPeriod: st.startPeriod || '',
        finishPeriod: st.finishPeriod || '',
        status: stageStep === 4 ? 'Pelaporan (LHA)' : (stageStep >= 2 ? 'Fieldwork Active' : 'Persiapan Lapangan'),
        stageStep,
        currentStageName,
        kkaCompleted
      }
    })
})

// Filtered fieldwork list by search query
const filteredFieldworkList = computed(() => {
  const q = fieldworkSearch.value.toLowerCase().trim()
  if (!q) return combinedFieldworkList.value
  return combinedFieldworkList.value.filter((item) => {
    return (
      (item.letterNumber || '').toLowerCase().includes(q) ||
      (item.auditTitle || '').toLowerCase().includes(q) ||
      (item.workingUnit || '').toLowerCase().includes(q)
    )
  })
})

// Assignment letter dropdown options
const assignmentLetterOptions = computed(() => {
  return combinedFieldworkList.value.map((item) => ({
    label: `${item.letterNumber || '-'} - ${item.auditTitle || 'Penugasan Audit'}`,
    value: item.letterNumber
  }))
})

// Currently selected assignment letter details
const activeAssignmentData = computed(() => {
  const target = selectedLetterNumber.value || combinedFieldworkList.value[0]?.letterNumber
  const found = combinedFieldworkList.value.find(i => i.letterNumber === target)
  return found || combinedFieldworkList.value[0] || {
    letterNumber: '-',
    auditTitle: 'Belum ada surat tugas yang dipilih',
    workingUnit: '-',
    auditTeam: '-',
    category: '-',
    startPeriod: '',
    finishPeriod: '',
    status: '',
    stageStep: 0,
    currentStageName: '-',
    kkaCompleted: 0
  }
})

// Top Stats
const activeAssignmentsCount = computed(() => combinedFieldworkList.value.length)
const inProgressAssignmentsCount = computed(() => {
  return combinedFieldworkList.value.filter(i => i.stageStep >= 2).length
})

const totalWorkingPapersCount = computed(() => {
  return (
    workingPaperStore.dataF01.length +
    workingPaperStore.dataF02.length +
    workingPaperStore.dataF03.length +
    workingPaperStore.dataF04.length +
    workingPaperStore.dataF05.length
  )
})

const kkaCompletionRate = computed(() => {
  if (combinedFieldworkList.value.length === 0) return 0
  const totalPossible = combinedFieldworkList.value.length * 5
  const current = combinedFieldworkList.value.reduce((acc, curr) => acc + (curr.kkaCompleted || 0), 0)
  return Math.min(100, Math.round((current / totalPossible) * 100))
})

// Real Draft Findings (Auditor Scope: Live Working Paper Causes + Ineffective Controls + Draft LHA)
const draftFindingsList = computed(() => {
  const targetLetter = selectedLetterNumber.value
  const list: Array<{
    title: string
    category: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
    assignmentLetterId: string
    source: string
    status: string
  }> = []
  const seen = new Set<string>()

  const addFinding = (title: string, category: any, letterId: string, source: string) => {
    const cleanTitle = (title || '').trim()
    if (!cleanTitle) return
    const key = `${letterId}__${cleanTitle.toLowerCase()}`
    if (seen.has(key)) return
    seen.add(key)
    list.push({
      title: cleanTitle,
      category: (category as any) || 'Significant',
      assignmentLetterId: letterId,
      source: source || 'Kertas Kerja Audit',
      status: 'Draft'
    })
  }

  // 1. Findings from saved reports (Draft status only)
  auditResultStore.reportList
    .filter(r => r.status === 'Draft' && (!targetLetter || r.assignmentLetterId === targetLetter))
    .forEach(report => {
      if (Array.isArray(report.findings)) {
        report.findings.forEach(f => {
          addFinding(f.title, f.category, report.assignmentLetterId, 'Draft LHA')
        })
      }
    })

  // 2. Real-time auto-detected findings from KKA & Fieldwork for selected letter
  if (targetLetter) {
    if (typeof auditResultStore.getLocalAutoFindings === 'function') {
      try {
        const autoFindings = auditResultStore.getLocalAutoFindings(targetLetter)
        if (Array.isArray(autoFindings)) {
          autoFindings.forEach(f => {
            addFinding(f.title, f.category, targetLetter, f.source || 'Digital Working Paper (KKA)')
          })
        }
      } catch (err) {
        console.warn('Error reading auto findings for letter:', targetLetter, err)
      }
    }
  } else {
    combinedFieldworkList.value.forEach(item => {
      if (typeof auditResultStore.getLocalAutoFindings === 'function') {
        try {
          const autoFindings = auditResultStore.getLocalAutoFindings(item.letterNumber)
          if (Array.isArray(autoFindings)) {
            autoFindings.forEach(f => {
              addFinding(f.title, f.category, item.letterNumber, f.source || 'Digital Working Paper (KKA)')
            })
          }
        } catch (err) {
          console.warn('Error reading auto findings for letter:', item.letterNumber, err)
        }
      }
    })
  }

  // 3. Recent findings from backend API matching target letter
  if (Array.isArray(auditResultStore.recentFindings)) {
    auditResultStore.recentFindings
      .filter(rf => !targetLetter || rf.assignmentLetterId === targetLetter)
      .forEach(rf => {
        addFinding(rf.title, rf.category, rf.assignmentLetterId || targetLetter || '-', rf.source || 'Temuan Lapangan')
      })
  }

  return list
})

const highFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Very Significant').length
})

const mediumFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Significant').length
})

const lowFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Quite Significant').length
})

const notSignificantCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Not Significant').length
})

// Real PBC Document Tracker from Fieldwork Store
const pbcItemsList = computed(() => {
  const targetLetter = selectedLetterNumber.value
  if (!targetLetter) {
    return []
  }

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
      status: isCollected ? 'Received' : 'Pending'
    }
  })
})

const pbcStats = computed(() => {
  const total = pbcItemsList.value.length
  const collected = pbcItemsList.value.filter(item => item.status === 'Received').length
  const pending = total - collected
  return {
    total,
    collected,
    pending
  }
})

// Real KKA Module Completion Status (F-01 s/d F-05) for Selected Assignment Letter
const kkaFormStatus = computed(() => {
  const letter = selectedLetterNumber.value
  if (!letter) {
    return {
      f01: false,
      f02: false,
      f03: false,
      f04: false,
      f05: false
    }
  }

  const hasF01 = workingPaperStore.dataF01.some((d: any) => d.assignmentLetterId === letter)
  const hasF02 = workingPaperStore.dataF02.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF03 = workingPaperStore.dataF03.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF04 = workingPaperStore.dataF04.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF05 = workingPaperStore.dataF05.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)

  return {
    f01: hasF01,
    f02: hasF02,
    f03: hasF03,
    f04: hasF04,
    f05: hasF05
  }
})

// Real RCM Quick References from RCM Store
const rcmQuickReferences = computed(() => {
  if (!rcmStore.rcmList || rcmStore.rcmList.length === 0) return []
  return rcmStore.rcmList.slice(0, 3).map((item: any) => {
    const riskLevelScore = item.inherent_risk || item.total_weighted_score || 0
    let riskLevel = 'Moderate'
    if (riskLevelScore >= 15 || item.design_effectiveness_rating === 1) {
      riskLevel = 'High'
    } else if (riskLevelScore < 8 && item.operating_effectiveness_rating >= 4) {
      riskLevel = 'Low'
    }

    return {
      code: item.control_code || item.risk_code || 'RCM-REF',
      process: item.risk_event || item.department || 'Proses Operasional',
      riskLevel,
      controlDescription: item.control_description || 'Pengendalian internal standar operasional.'
    }
  })
})

// Navigation helpers
const goToWorkingPaper = (letterNumber: string) => {
  fieldworkStore.selectedAssignmentLetter = letterNumber
  workingPaperStore.selectedAssignmentLetter = letterNumber
  router.push('/working-paper')
}

const openKkaStep = (stepSlot: string) => {
  if (selectedLetterNumber.value) {
    fieldworkStore.selectedAssignmentLetter = selectedLetterNumber.value
    workingPaperStore.selectedAssignmentLetter = selectedLetterNumber.value
  }
  router.push(`/working-paper?step=${stepSlot}`)
}

const goToFieldwork = (letterNumber: string) => {
  fieldworkStore.selectedAssignmentLetter = letterNumber
  router.push('/audit-fieldwork')
}

const goToResultReport = (letterNumber: string) => {
  auditResultStore.selectedAssignmentLetter = letterNumber
  router.push('/audit-result-report')
}

// Helpers
const formatPeriod = (start?: string, end?: string) => {
  if (!start && !end) return '-'
  if (start && end) return `${start} s/d ${end}`
  return start || end || '-'
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

const getFieldworkStatusColor = (status?: string): 'success' | 'info' | 'warning' | 'neutral' => {
  if (status === 'Fieldwork Active') return 'success'
  if (status === 'Pelaporan (LHA)') return 'info'
  if (status === 'Persiapan Lapangan') return 'warning'
  return 'neutral'
}

const getFindingCategoryColor = (cat?: string): 'error' | 'warning' | 'info' | 'neutral' => {
  if (cat === 'Very Significant') return 'error'
  if (cat === 'Significant') return 'warning'
  if (cat === 'Quite Significant') return 'info'
  return 'neutral'
}

let isMounted = false

onBeforeUnmount(() => {
  isMounted = false
})

// Silent initial load for dashboard on mount (no toast notification)
const loadInitialAuditorData = async () => {
  try {
    await Promise.allSettled([
      assignmentLetterStore.fetchAssignmentLetters(),
      workingPaperStore.fetchAllData(),
      auditResultStore.fetchReports(),
      auditResultStore.fetchRecentFindings(10),
      rcmStore.fetchRCMList()
    ])
    if (!isMounted) return
    if (!selectedLetterNumber.value && combinedFieldworkList.value[0]) {
      selectedLetterNumber.value = combinedFieldworkList.value[0].letterNumber
    }
    if (selectedLetterNumber.value) {
      fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value).catch(() => {})
    }
  } catch (err) {
    console.warn('Initial auditor dashboard data load warning:', err)
  }
}

// Manual Sync handler - triggered on explicit button click
const manualSyncAuditorData = async () => {
  isSyncing.value = true
  try {
    await Promise.allSettled([
      assignmentLetterStore.fetchAssignmentLetters(),
      workingPaperStore.fetchAllData(),
      auditResultStore.fetchReports(),
      auditResultStore.fetchRecentFindings(10),
      rcmStore.fetchRCMList()
    ])
    if (!isMounted) return
    if (selectedLetterNumber.value) {
      await fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value, true)
    }
    toast.success('Data operasional auditor berhasil disinkronkan')
  } catch (err) {
    if (isMounted) {
      toast.error('Gagal memperbarui data auditor')
    }
  } finally {
    if (isMounted) {
      isSyncing.value = false
    }
  }
}

// Watch selected letter change to safely synchronize stores
watch(
  () => selectedLetterNumber.value,
  (rawVal) => {
    if (!isMounted) return
    const newVal = typeof rawVal === 'string' ? rawVal : (rawVal as any)?.value || ''
    if (newVal) {
      if (fieldworkStore.selectedAssignmentLetter !== newVal) {
        fieldworkStore.selectedAssignmentLetter = newVal
      }
      if (workingPaperStore.selectedAssignmentLetter !== newVal) {
        workingPaperStore.selectedAssignmentLetter = newVal
      }
    }
  }
)

// Auto-select first letter when list is loaded
watch(
  () => combinedFieldworkList.value,
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
  await loadInitialAuditorData()
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
