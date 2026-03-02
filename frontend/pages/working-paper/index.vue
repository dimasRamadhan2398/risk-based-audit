<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 bg-gray-50 dark:bg-gray-900 min-h-screen">
    
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Kertas Kerja Audit</h1>
        <p class="text-sm text-gray-500">Digital Working Paper System</p>
      </div>
    </div>

    <UTabs :items="tabs" class="w-full">
      
      <template #f01="{ item }">
        <UCard class="mt-4 shadow-sm p-8">
          <div class="justify-between items-center mb-10">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Referensi Penugasan</h2>
          </div>

          <div class="space-y-6">
            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Nomor Surat Tugas <span class="text-red-500">*</span></label>
              <USelectMenu class="md:col-span-3" v-model="store.form.suratTugas" :items="store.options.suratTugas" placeholder="Pilih Nomor Surat Tugas" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tujuan Audit <span class="text-red-500">*</span></label>
              <UInput class="md:col-span-3" v-model="store.form.tujuanAudit" disabled placeholder="(Otomatis terisi saat mengisi surat tugas)" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Proses Bisnis <span class="text-red-500">*</span></label>
              <USelectMenu class="md:col-span-3" v-model="store.form.prosesBisnis" :items="store.options.prosesBisnis" placeholder="Pilih Proses Bisnis" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300">Periode Audit <span class="text-red-500">*</span></label>
              <div class="md:col-span-3">
                <UFormField :error="dateErrorMessage">
                  <div class="flex items-center gap-4 w-full">
                    <UInput 
                      type="date" 
                      v-model="store.form.periodeStart" 
                      icon="i-heroicons-calendar" 
                      class="w-full"
                      :color="isDateError ? 'error' : 'neutral'"
                    />
                    
                    <span class="text-gray-500 font-bold whitespace-nowrap">s/d</span>
                    
                    <UInput 
                      type="date" 
                      v-model="store.form.periodeEnd" 
                      icon="i-heroicons-calendar" 
                      class="w-full"
                      :color="isDateError ? 'error' : 'neutral'"
                    />
                  </div>
                </UFormField>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Lokasi Cabang <span class="text-red-500">*</span></label>
              <USelectMenu class="md:col-span-3" v-model="store.form.lokasi" :items="store.options.lokasi" placeholder="Pilih lokasi" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tim Audit</label>
              <div class="md:col-span-3 space-y-4">
    
                <div v-for="(member, index) in store.form.teamMembers" :key="member.id" class="flex gap-2 items-center">
                  <div class="grid grid-cols-2 gap-2 flex-1">
                    <USelectMenu 
                      v-model="member.name" 
                      :items="getAvailableMembers(index)" 
                      placeholder="Pilih Nama Anggota"
                    />
                    <UInput 
                      v-model="member.role" 
                      placeholder="Jabatan (ex: Ketua Tim)" 
                    />
                  </div>

                  <UButton 
                    v-if="store.form.teamMembers.length > 1"
                    icon="i-heroicons-trash" 
                    color="error" 
                    variant="ghost" 
                    @click="store.removeTeamMember(index)" 
                  />
                </div>

                <UButton 
                  color="primary" 
                  variant="soft"
                  icon="i-heroicons-plus" 
                  label="Tambah Anggota Audit" 
                  @click="store.addTeamMember()"
                />
              </div>
            </div>
          </div>
          <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF01()" />
          </div>
        </UCard>

        <UCard class="shadow-sm mt-10">
          <div class="p-4 border-b border-gray-300 dark:border-gray-600">
            <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Penugasan</h3>
          </div>
          <UTable :data="store.savedF01" :columns="columnsF01" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" >
          
            <template #teamMembers-cell="{ row }">
              <div class="flex flex-wrap gap-1">
                <UBadge 
                  v-for="member in row.original.teamMembers"
                  :key="member.id"
                  color="neutral" 
                  variant="subtle" 
                  size="lg"
                  class="flex flex-col items-start px-2 py-1"
                >
                  <span class="font-bold text-primary-700">{{ member.name }}</span>
                  <span class="text-[10px] opacity-70 italic">{{ member.role }}</span>
                </UBadge>
                <span v-if="!row.original.teamMembers?.length" class="text-gray-400">-</span>
              </div>
            </template>
        
          </UTable>
        </UCard>

      </template>

      <template #f02="{ item }">
        <UCard class="mt-4 shadow-sm p-8">
          <div class="justify-between items-center mb-10">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Risk & Control</h2>
          </div>

          <div class="space-y-6">
            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Resiko (Dari Risk Register) <span class="text-red-500">*</span></label>
              <USelectMenu class="md:col-span-3" v-model="store.form.resiko" :items="store.options.resiko" placeholder="Pilih Resiko" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Taksonomi <span class="text-red-500">*</span></label>
              <UInput class="md:col-span-3" v-model="store.form.taksonomi" disabled placeholder="Otomatis terisi saat memilih resiko" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Tingkat Resiko <span class="text-red-500">*</span></label>
              <UInput class="md:col-span-3" v-model="store.form.tingkatResiko" disabled placeholder="Otomatis terisi saat memilih resiko" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Deskripsi Kontrol yang Diuji</label>
              <UTextarea class="md:col-span-3" v-model="store.form.deskripsiKontrol" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
            </div>
          </div>
          <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF02()" />
          </div>
        </UCard>

        <UCard class="shadow-sm mt-10">
          <div class="p-4 border-b border-gray-300 dark:border-gray-600">
            <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Risk & Control Profile</h3>
          </div>
          <UTable :data="store.savedF02" :columns="columnsF02" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" />
        </UCard>

      </template>

      <template #f03="{ item }">
        <div class="mt-4 space-y-6">

          <UCard class="mt-4 shadow-sm p-8">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-10">Populasi & Sampel</h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-full">
              <UFormField label="Total Populasi" required>
                <UInput type="number" v-model="store.form.populasi" placeholder="Ex: 100" class="w-full"/>
              </UFormField>
              <UFormField label="Jumlah Sampel yang Diuji" required>
                <UInput type="number" v-model="store.form.jmlSampel" placeholder="Ex: 10" class="w-full"/>
              </UFormField>
            </div>

            <div class="justify-between items-center mt-10">
              <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white">Matriks Pengujian Kontrol</h2>
            </div>
            
            <div class="space-y-6 mt-6">
              <div v-for="(sampel, index) in store.form.samples" :key="sampel.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 relative bg-white dark:bg-gray-800">
                <h3 class="text-lg font-bold mb-4">Sampel {{ index + 1 }}</h3>
                
                <div class="space-y-4 max-w-full">
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Nama Dokumen <span class="text-red-500">*</span></label>
                    <UInput class="md:col-span-3" v-model="sampel.dokumen" placeholder="Ex: PO-2026-001" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Langkah 1 <span class="text-red-500">*</span></label>
                    <USelectMenu class="md:col-span-3" v-model="sampel.l1" :items="store.options.testResult" placeholder="Pilih Langkah" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Langkah 2 <span class="text-red-500">*</span></label>
                    <USelectMenu class="md:col-span-3" v-model="sampel.l2" :items="store.options.testResult" placeholder="Pilih Langkah" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Langkah 3 <span class="text-red-500">*</span></label>
                    <USelectMenu class="md:col-span-3" v-model="sampel.l3" :items="store.options.testResult" placeholder="Pilih Langkah" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center mt-2">
                    <label class="font-semibold text-sm">Status</label>
                    <div class="md:col-span-3 flex items-center gap-2">
                      <div class="w-4 h-4 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                      <span class="font-bold">{{ store.checkSampleStatus(sampel) ? 'Efektif' : 'Tidak Efektif' }}</span>
                    </div>
                  </div>
                </div>

                <div class="mt-4 flex justify-end">
                  <UButton label="Hapus Sampel" color="error" variant="ghost" class="font-bold" @click="store.removeSample(index)" />
                </div>
              </div>
              <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Sampel" @click="store.addSample" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
              <label class="font-semibold text-sm mt-2">Kesimpulan</label>
              <UTextarea class="md:col-span-3" v-model="store.form.kesimpulan" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
            </div>

            <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
              <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF03()" />
            </div>
          </UCard>

          <UCard class="shadow-sm mt-10">
            <div class="p-4 border-b border-gray-300 dark:border-gray-600">
              <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Pengujian Kontrol</h3>
            </div>
            <UTable :data="store.savedF03" :columns="columnsF03" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
              <template #samples-cell="{ row }">
                <div class="flex flex-col gap-2 max-w-md">
                  <div 
                    v-for="s in row.original.samples" 
                    :key="s.id" 
                    class="text-sm p-2 border rounded bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700"
                  >
                    <div class="flex justify-between font-bold mb-1">
                      <span>{{ s.dokumen || 'Tanpa Nama Dokumen' }}</span>
                      <UBadge 
                        :color="store.checkSampleStatus(s) ? 'success' : 'error'" 
                        size="md" 
                        variant="subtle"
                      >
                        {{ store.checkSampleStatus(s) ? 'Efektif' : 'Tidak Efektif' }}
                      </UBadge>
                    </div>
                    <div class="text-sm text-gray-500 italic">
                      L1: {{ s.l1 || '-' }} | L2: {{ s.l2 || '-' }} | L3: {{ s.l3 || '-' }}
                    </div>
                  </div>
                  
                  <span v-if="!row.original.samples?.length" class="text-gray-400 italic text-xs">
                    Tidak ada data sampel
                  </span>
                </div>
              </template>
            </UTable>
          </UCard>

        </div>
      </template>

      <template #f04="{ item }">
        <div class="mt-4 space-y-6">
          
          <UCard class="shadow-sm p-8">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Detail Temuan</h2>
            
            <div class="space-y-6 max-w-full">
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                <label class="font-semibold text-sm mt-2">Kondisi (Dampak di Lapangan)</label>
                <UTextarea class="md:col-span-3" v-model="store.form.kondisi" :rows="3" placeholder="Ex: Ditemukan dokumen PO tanpa tanda tangan Manager terkait pada tanggal..." />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-semibold text-sm mt-2">Unggah bukti temuan/foto</label>
              
              <div class="md:col-span-3">
                <div 
                  @click="triggerUpload"
                  class="border-2 border-dashed border-gray-400 rounded-lg p-8 flex flex-col items-center justify-center text-center hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer relative"
                  :class="{ 'border-gray-500 bg-gray-50': store.form.buktiFile }"
                >
                  <input 
                    type="file" 
                    ref="fileInput" 
                    class="hidden" 
                    accept="image/png, image/jpeg" 
                    @change="onFileChange"
                  />

                  <template v-if="!store.form.buktiFile">
                    <UIcon name="i-heroicons-arrow-up-tray" class="w-8 h-8 text-gray-600 mb-2" />
                    <span class="font-bold text-gray-800 dark:text-white">Upload Here or Drag and Drop</span>
                    <span class="text-xs text-gray-500 mt-1">Jpg, Png (Max 10MB)</span>
                  </template>

                  <template v-else>
                    <UIcon name="i-heroicons-document-check" class="w-8 h-8 text-gray-500 mb-2" />
                    <span class="font-bold text-gray-700">{{ store.form.buktiFile.name }}</span>
                    <span class="text-xs text-gray-500 mt-1">
                      {{ (store.form.buktiFile.size / 1024).toFixed(2) }} KB
                    </span>
                    <UButton 
                      label="Ganti File" 
                      variant="link" 
                      color="error" 
                      size="xs" 
                      class="mt-2" 
                      @click.stop="removeFile" 
                    />
                  </template>
                </div>
              </div>
            </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Kriteria (Aturan/SOP Terkait) <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="store.form.kriteria" placeholder="Cari Peraturan Internal (Contoh: SOP Pengadaan Bab IV - Otorisasi)" />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                <label class="font-semibold text-sm mt-2">Dampak (Risk Impact)</label>
                <UTextarea class="md:col-span-3" v-model="store.form.dampak" :rows="3" placeholder="Ex: Potensi fraud atau pembelian fiktif yang dapat merugikan keuangan..." />
              </div>

              <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Root Cause Analysis</h2>
              <div v-for="(rca, index) in store.form.rcaList" :key="rca.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 bg-white dark:bg-gray-800">
                <h3 class="text-lg font-bold mb-4">Analisis {{ index + 1 }}</h3>
                
                <div class="space-y-4 max-w-full">
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Kategori <span class="text-red-500">*</span></label>
                    <USelectMenu class="md:col-span-3" v-model="rca.kategori" :items="store.options.rcaCategory" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Why 1 <span class="text-red-500">*</span></label>
                    <UInput class="md:col-span-3" v-model="rca.w1" placeholder="Ex: Staf lupa meminta TTD Manager" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Why 2</label>
                    <UInput class="md:col-span-3" v-model="rca.w2" placeholder="Ex: Karena staf terburu-buru mengejar kuota pengiriman" />
                  </div>
                  <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                    <label class="font-semibold text-sm">Why 3</label>
                    <UInput class="md:col-span-3" v-model="rca.w3" placeholder="-" />
                  </div>
                </div>

                <div class="mt-4 flex justify-end">
                  <UButton label="Hapus Kategori" color="error" variant="ghost" class="font-bold" @click="store.removeRCA(index)" />
                </div>
              </div>

              <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Kategori" @click="store.addRCA" />
            </div>

            <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
              <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF04()" />
            </div>

          </UCard>

          <UCard class="shadow-sm mt-10">
            <div class="p-4 border-b border-gray-300 dark:border-gray-600">
              <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Temuan</h3>
            </div>
            <UTable :data="store.savedF04" :columns="columnsF04" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
              <template #buktiFile-cell="{ row }">
                <div v-if="row.original.buktiFile" class="flex items-center gap-1 text-blue-600">
                  <UIcon name="i-heroicons-paper-clip" />
                  <span class="text-md truncate max-w-[150px]">{{ row.original.buktiFile.name }}</span>
                </div>
                <span v-else class="text-gray-400">-</span>
              </template>

              <template #rcaList-cell="{ row }">
                <div class="space-y-2 py-2">
                  <div v-for="rca in row.original.rcaList" :key="rca.id" class="text-[11px] leading-tight border-l-2 border-orange-400 pl-2">
                    <div class="font-bold text-lg text-gray-700">{{ rca.kategori }}</div>
                    <div class="text-gray-500 text-sm italic">
                      Why 1: {{ rca.w1 || '-' }} <br>
                      Why 2: {{ rca.w2 || '-' }} <br>
                      Why 3: {{ rca.w3 || '-' }}
                    </div>
                  </div>
                </div>
              </template>
            </UTable>
          </UCard>

        </div>
      </template>

      <template #f05="{ item }">
        <div class="mt-4 space-y-6">
          
          <UCard class="shadow-sm p-8">
            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Rekomendasi & Tanggapan</h2>
            
            <div class="space-y-6 max-w-full">
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                <label class="font-semibold text-sm mt-2">Rekomendasi Auditor (Solusi) <span class="text-red-500">*</span></label>
                <UTextarea class="md:col-span-3" v-model="store.form.rekomendasi" :rows="3" placeholder="Ex: Tim IT perlu menambahkan fitur hard-block..." />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
                <label class="font-semibold text-sm mt-2">Tanggapan Audite (Managemen Response)</label>
                <UTextarea class="md:col-span-3" v-model="store.form.tanggapan" :rows="3" placeholder="Ex: Kami setuju, update akan dilakukan di Q3..." />
              </div>

              <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Detail Rencana Aksi (Action Plan)</h2>
              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Deskripsi Action <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="store.form.deskripsiAction" placeholder="Ex: Staf lupa meminta TTD Manager" />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">PIC (Penanggung Jawab) <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="store.form.pic" icon="i-heroicons-magnifying-glass" :items="store.options.pic" placeholder="Cari Nama Karyawan Ex: Dimas - IT" />
              </div>

              <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Periode Perencanaan <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" type="date" v-model="store.form.periodAction" icon="i-heroicons-calendar" />
              </div>
            </div>

            <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
              <UButton label="Simpan" color="primary" icon="i-heroicons-document-check" @click="store.saveF05()" />
            </div>
          </UCard>

          <UCard class="shadow-sm mt-10">
            <div class="p-4 border-b border-gray-300 dark:border-gray-600">
              <h3 class="font-bold text-gray-700 dark:text-gray-200">Data Rekomendasi & Tanggapan</h3>
            </div>
            <UTable :data="store.savedF05" :columns="columnsF05" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" />
          </UCard>

        </div>
      </template>
    </UTabs>

  </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Panggil Store
const store = useWorkingPaperStore()
const fileInput = ref<HTMLInputElement | null>(null)

// Picu klik pada input file yang tersembunyi
const triggerUpload = () => {
  fileInput.value?.click()
}

// Handle perubahan file
const onFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  const files = target.files

  if (files && files[0]) {
    const file = files[0]
    
    // Validasi Ukuran (Contoh: 10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert('Ukuran file terlalu besar! Maksimal 10MB.')
      return
    }

    // Simpan ke Store
    store.form.buktiFile = file
    console.log('File terpilih:', file.name)
  }
}

// Hapus file
const removeFile = () => {
  store.form.buktiFile = null
  if (fileInput.value) {
    fileInput.value.value = '' // Reset input agar bisa upload file yang sama lagi
  }
}

// Fungsi untuk mendapatkan daftar anggota yang tersedia untuk baris tertentu
const getAvailableMembers = (currentIndex: number) => {
  // 1. Ambil semua nama yang sudah dipilih di baris-baris LAIN
  const selectedNames = store.form.teamMembers
    .filter((_, index) => index !== currentIndex) // Kecualikan baris yang sedang aktif
    .map(member => member.name)
    .filter(name => !!name) // Hanya ambil yang sudah ada isinya

  // 2. Filter master list PIC agar tidak menyertakan nama yang sudah dipilih di baris lain
  return store.options.pic.filter(pic => !selectedNames.includes(pic))
}

const isDateError = computed(() => {
  // Jika salah satu tanggal belum diisi, jangan anggap error dulu
  if (!store.form.periodeStart || !store.form.periodeEnd) return false
  
  const start = new Date(store.form.periodeStart)
  const end = new Date(store.form.periodeEnd)
  
  // Return true jika tanggal akhir LEBIH KECIL dari tanggal mulai
  return end < start
})

// Opsional: Pesan error dinamis
const dateErrorMessage = computed(() => {
  return isDateError.value ? "Tanggal akhir tidak boleh sebelum tanggal mulai" : false
})

// Tabs Configuration
const tabs = [
  { label: 'Referensi', slot: 'f01', icon: 'i-heroicons-document-text' },
  { label: 'Profil Resiko', slot: 'f02', icon: 'i-heroicons-shield-exclamation' },
  { label: 'Uji Sampel', slot: 'f03', icon: 'i-heroicons-table-cells' },
  { label: 'AOI & RCA', slot: 'f04', icon: 'i-heroicons-magnifying-glass-circle' },
  { label: 'Action Plan', slot: 'f05', icon: 'i-heroicons-check-badge' }
]

const columnsF01 = [
  { accessorKey: 'suratTugas', header: 'Surat Tugas' },
  { accessorKey: 'prosesBisnis', header: 'Proses Bisnis' },
  { accessorKey: 'periode', header: 'Periode' },
  { accessorKey: 'lokasi', header: 'Lokasi' },
  { accessorKey: 'teamMembers', header: 'Team' }
]

const columnsF02 = [
  { accessorKey: 'resiko', header: 'Resiko' },
  { accessorKey: 'taksonomi', header: 'Taksonomi' },
  { accessorKey: 'tingkatResiko', header: 'Tingkat Resiko' },
  { accessorKey: 'deskripsiKontrol', header: 'Deskripsi Kontrol' }
]

const columnsF03 = [
  { accessorKey: 'populasi', header: 'Populasi' },
  { accessorKey: 'jmlSampel', header: 'Jumlah Sample' },
  { accessorKey: 'samples', header: 'Daftar Sampel' },
  { accessorKey: 'kesimpulan', header: 'Kesimpulan' }
]

const columnsF04 = [
  { accessorKey: 'kondisi', header: 'Kondisi' },
  { accessorKey: 'kriteria', header: 'Kriteria' },
  { accessorKey: 'dampak', header: 'Dampak' },
  { accessorKey: 'rcaList', header: 'Root Cause' },
  { accessorKey: 'buktiFile', header: 'Dokumen Bukti' },
]

const columnsF05 = [
  { accessorKey: 'rekomendasi', header: 'Rekomendasi' },
  { accessorKey: 'tanggapan', header: 'Tanggapan Auditee' },
  { accessorKey: 'deskripsiAction', header: 'Deskripsi' },
  { accessorKey: 'pic', header: 'PIC' },
  { accessorKey: 'periodAction', header: 'Target Selesai' }
]

</script>