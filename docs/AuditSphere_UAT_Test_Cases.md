# DOKUMEN USER ACCEPTANCE TESTING (UAT) & TEST CASES
## AuditSphere — Enterprise Risk-Based Internal Audit Platform

> **Dokumen Uji Penerimaan Pengguna (UAT) Komprehensif**  
> **Versi Dokumen:** 2.2 | **Status:** Siap Uji & Validasi Formal | **Tanggal:** 11 November 2026  
> **Sistem Target:** AuditSphere Enterprise Suite (Frontend Nuxt 4, Kong Gateway, Microservices Go & Python)

---

## Ringkasan Eksekutif & Struktur Skenario Pengujian

Dokumen ini mendefinisikan seluruh kasus uji (*Test Cases*) terstruktur untuk 21 modul utama (29 tabel skenario) pada platform AuditSphere. Format pengujian disusun secara baku dengan 7 atribut pengujian profesional serta dilengkapi **Lembar Pengesahan dan Tanda Tangan Penguji pada setiap modul**:
1. **Test Case ID**: Identifikasi unik kasus uji berbasis skenario (contoh: AS-UAT-001-01, AS-UAT-014-01-01).
2. **Test Case**: Judul skenario pengujian fungsional/keamanan.
3. **Test Data**: Kumpulan data masukan realistis yang digunakan saat pengujian.
4. **Langkah Pengujian**: Panduan eksekusi langkah demi langkah bagi penguji (tester/auditor).
5. **Expected Result**: Hasil yang diharapkan sesuai spesifikasi sistem dan standar audit.
6. **Actual Result**: Hasil pengamatan riil saat sistem dieksekusi.
7. **Status / Bug ID**: Status kelulusan pengujian (*Pass/Fail*) serta referensi ID tiket perbaikan jika terjadi defek.
8. **Lembar Pengesahan Modul**: Tabel tanda tangan bukti pengujian telah selesai diuji oleh Penguji Utama, Pemilik Proses, dan Lead QA.

---

## Tabel AS-UAT-001 — Login dan MFA
**Tujuan:** Memastikan proses autentikasi pengguna primer (Username/Password), verifikasi multi-faktor (MFA/TOTP), pengenalan perangkat tepercaya, dan penegakan pakta kerahasiaan berjalan aman dan akurat.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-001-01** | Masuk ke Aplikasi Menggunakan Akun Pengguna dan Melanjutkan ke Layar Verifikasi Keamanan (MFA) | Nama Pengguna: 'auditor1', Kata Sandi: 'Password@123', Pilihan: 'Ingat Saya' dicentang | 1. Buka halaman Masuk (Login) AuditSphere.<br>2. Ketik nama pengguna dan kata sandi yang terdaftar.<br>3. Beri tanda centang pada kotak 'Ingat Saya'.<br>4. Klik tombol biru 'Masuk'. | Sistem mengenali akun pengguna dengan benar dan langsung mengarahkan tampilan ke halaman input 'Kode Verifikasi Keamanan (MFA)' untuk meminta kode 6 angka tambahan. | Sesuai — Akun valid berhasil diproses dan layar otomatis berpindah ke formulir input kode verifikasi tanpa kendala. | Pass / - |
| **AS-UAT-001-02** | Memasukkan 6 Angka Kode Verifikasi Keamanan dan Membuka Dashboard Utama | Sesi Akun Auditor, Kode Verifikasi: '849201' (6 angka dari aplikasi Authenticator di handphone), Pilihan: 'Percayai Perangkat Ini' dicentang | 1. Pada layar Verifikasi Keamanan, masukkan 6 digit kode dari aplikasi di handphone.<br>2. Beri centang pada opsi 'Percayai Perangkat Ini'.<br>3. Klik tombol 'Verifikasi & Masuk'. | Kode keamanan diterima dengan sukses, sistem mengenali laptop/komputer pengguna sebagai perangkat tepercaya, dan halaman Dashboard Utama langsung terbuka rapi. | Sesuai — Kode verifikasi sukses diterima dan pengguna langsung diarahkan masuk ke halaman Dashboard Utama. | Pass / - |
| **AS-UAT-001-03** | Menolak Kata Sandi yang Salah atau Kode Verifikasi yang Tidak Sesuai | Uji Coba 1: Kata sandi sengaja disalahkan; Uji Coba 2: Kode verifikasi salah ('000000') | 1. Masukkan kata sandi yang salah pada layar login pertama.<br>2. Klik tombol 'Masuk' dan amati pesan yang muncul.<br>3. Coba masukkan kode keamanan yang salah atau kedaluwarsa pada layar verifikasi kedua.<br>4. Klik tombol 'Verifikasi & Masuk'. | Sistem menolak akses masuk, menampilkan pesan peringatan jelas berwarna merah ('Kombinasi nama pengguna atau kata sandi tidak valid' atau 'Kode verifikasi salah atau kedaluwarsa'), serta tidak mengizinkan pengguna masuk ke sistem. | Sesuai — Sistem secara aman menolak akun atau kode yang salah dan memunculkan pesan peringatan yang mudah dipahami. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-001 — Login dan MFA
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-002 — Audit Charter
**Tujuan:** Memastikan piagam audit internal (Audit Charter), pedoman pelaksanaan (Guideline), dan SOP audit dapat dirumuskan, direviu, disahkan oleh Komite Audit/Direksi, dan dipublikasikan secara resmi.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-002-01** | Menambahkan dan Mengunggah Dokumen Piagam Audit (Audit Charter) Baru | Judul Piagam: 'Piagam Audit Internal (Audit Charter) 2026', Nomor Surat: 'SK/CAE/CHARTER/2026/01', Masa Berlaku: '2026 - 2028', Berkas Lampiran: Dokumen_Piagam_Audit_2026.pdf (Ukuran: 2 MB) | 1. Buka menu Piagam Audit (Audit Charter).<br>2. Klik tab 'Audit Charter' lalu klik tombol '+ Tambah / Unggah Dokumen'.<br>3. Lengkapi nama dokumen, nomor surat keputusan, masa berlaku, dan lampirkan berkas PDF.<br>4. Klik tombol 'Simpan Dokumen'. | Dokumen piagam audit baru tersimpan rapi dengan label status 'Draf', berkas PDF terlampir aman, dan nama dokumen langsung muncul pada tabel daftar piagam audit. | Sesuai — Dokumen piagam baru berhasil disimpan, berstatus 'Draf', dan siap untuk diajukan ke tahap persetujuan. | Pass / - |
| **AS-UAT-002-02** | Menyetujui dan Mengesahkan Piagam Audit oleh Komite Audit / Pimpinan | Piagam Audit: 'Piagam Audit Internal 2026', Peran Pengguna: Pimpinan Audit / Komite Audit, Catatan: 'Disetujui dalam rapat komite audit dan dewan komisaris' | 1. Masuk menggunakan akun pimpinan audit atau komite audit.<br>2. Buka dokumen draf Piagam Audit yang telah diajukan.<br>3. Periksa pasal kewenangan, independensi, dan catatan pengajuan.<br>4. Klik tombol hijau 'Setujui & Aktifkan Piagam'. | Label status dokumen berubah menjadi warna hijau 'Aktif', versi piagam lama otomatis dipindahkan ke arsip riwayat, dan piagam resmi menjadi pedoman aktif seluruh tim audit. | Sesuai — Piagam audit sukses disahkan, label berubah menjadi 'Aktif', dan dokumen resmi siap digunakan sebagai acuan kerja. | Pass / - |
| **AS-UAT-002-03** | Mengelola Pedoman Pelaksanaan dan SOP Pelaksanaan Audit Internal | Tab: 'SOP & Pedoman', Judul: 'SOP-IA-004: Prosedur Pengujian Kepatuhan Anti-Fraud', Kategori: 'Prosedur Operasional', Berkas: SOP_Pemeriksaan_2026.pdf | 1. Pilih tab menu 'SOP' pada halaman Piagam Audit.<br>2. Klik tombol '+ Tambah SOP'.<br>3. Isi nama kode SOP, judul prosedur, dan unggah berkas panduan pelaksanaan.<br>4. Klik tombol 'Simpan' lalu coba klik tombol 'Unduh'. | SOP baru terdaftar rapi pada tabel panduan kerja, dan berkas dapat dibuka atau diunduh dengan lancar oleh seluruh anggota tim auditor. | Sesuai — Seluruh berkas panduan dan SOP dapat dicari, ditinjau, dan diunduh oleh auditor dengan isi yang utuh. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-002 — Audit Charter
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-003 — Corporate Risk Profile
**Tujuan:** Memastikan profil risiko korporat dapat diidentifikasi, dinilai tingkat risiko inheren dan residualnya per kuartal (Q1-Q4), dipetakan pada Heat Map 5x5, dan ditandai sebagai prioritas audit.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-003-01** | Mencatat Potensi Risiko Baru dan Menilai Tingkat Keparahannya per Triwulan (Q1-Q4) | Nama Risiko: 'Gangguan Layanan Pembayaran Digital Core Banking', Kategori: 'Operasional & Teknologi', Unit/Divisi: 'Teknologi Informasi', Nilai Dampak: 4 (Tinggi), Nilai Peluang: 4 (Sering) -> Skor: 16 | 1. Buka menu Profil Risiko Korporat.<br>2. Klik tombol '+ Tambah Risiko'.<br>3. Isi formulir nama risiko, pilih divisi penanggung jawab, serta pilih tingkat Dampak dan Peluangnya.<br>4. Klik tombol 'Simpan'. | Sistem otomatis menghitung total nilai risiko (Dampak x Peluang = 16), memunculkan notifikasi sukses warna hijau 'Data berhasil disimpan', dan data baru langsung muncul di daftar risiko korporat. | Sesuai — Potensi risiko baru berhasil dicatat, perhitungan nilai risiko otomatis dan akurat, serta jumlah total risiko bertambah. | Pass / - |
| **AS-UAT-003-02** | Memeriksa Pemetaan Risiko pada Diagram Peta Risiko (Heat Map 5x5) | Kumpulan risiko bernilai tinggi (skor 16 sampai 25), Periode: Tahun 2026 Triwulan 1 (Q1) | 1. Buka tampilan diagram Peta Risiko (Heat Map 5x5) pada halaman Profil Risiko.<br>2. Perhatikan posisi koordinat risiko berdasarkan nilai Dampak dan Peluangnya.<br>3. Periksa angka jumlah pada kotak ringkasan 'Risiko Prioritas'. | Setiap risiko terpetakan dengan tepat pada kotak warna merah/kuning sesuai tingkat keparahannya, dan risiko bernilai tinggi otomatis ditandai sebagai Prioritas untuk bahan rencana audit. | Sesuai — Diagram peta risiko menampilkan titik sebaran risiko dengan jelas dan menandai risiko prioritas dengan benar. | Pass / - |
| **AS-UAT-003-03** | Menyaring Tampilan Profil Risiko Berdasarkan Divisi, Kategori, dan Periode Waktu | Pilihan Kategori: 'Kepatuhan', Pilihan Unit/Divisi: 'Divisi Hukum & Kepatuhan', Tahun: 2026 | 1. Pada halaman Profil Risiko, klik menu pilihan (filter) Divisi dan pilih 'Divisi Hukum & Kepatuhan'.<br>2. Klik pilihan Kategori dan pilih 'Kepatuhan'.<br>3. Perhatikan pembaruan visual pada diagram peta risiko dan tabel daftar risiko. | Tampilan diagram peta risiko dan tabel daftar risiko secara instan hanya menampilkan risiko-risiko yang sesuai dengan divisi dan kategori yang dipilih. | Sesuai — Penyaringan data berfungsi sangat cepat dan responsif, menampilkan informasi yang tepat tanpa data yang tertinggal. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-003 — Corporate Risk Profile
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-004 — Risk Mitigation Plan
**Tujuan:** Memastikan rencana perlakuan dan pengendalian risiko (mitigasi) dapat dirumuskan secara terinci, ditetapkan PIC dan supervisornya, serta dilacak jadwal pelaksanaannya.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-004-01** | Menambahkan Rencana Perbaikan dan Pengendalian pada Risiko yang Teridentifikasi | Risiko Terkait: 'Gangguan Layanan Pembayaran Digital', Rencana Tindakan: 'Penerapan sistem cadangan server ganda dan uji coba pemulihan bencana rutin', Penanggung Jawab (PIC): Manajer Infrastruktur IT, Pengawas: VP Technology, Target Waktu: 01 April 2026 s/d 31 Oktober 2026 | 1. Dari rincian risiko pada Profil Risiko, klik tautan 'Rencana Mitigasi'.<br>2. Buka formulir 'Tambah Rencana Mitigasi'.<br>3. Tuliskan langkah perbaikan, pilih nama penanggung jawab, pengawas, dan rentang tanggal pelaksanaan.<br>4. Klik tombol 'Simpan Mitigasi'. | Rencana perbaikan berhasil disimpan, langsung terhubung dengan data risiko induk, dan tampil pada tabel daftar rencana mitigasi dengan status 'Sedang Berjalan'. | Sesuai — Rencana mitigasi berhasil dicatat dan terhubung secara tepat ke data risiko yang bersangkutan. | Pass / - |
| **AS-UAT-004-02** | Memastikan Tanggal Selesai Rencana Kerja Tidak Boleh Mendahului Tanggal Mulai | Tanggal Mulai: '2026-06-01', Tanggal Selesai: '2026-05-15' (Tanggal selesai sengaja diisi lebih lampau) | 1. Buka formulir tambah atau edit rencana mitigasi.<br>2. Masukkan tanggal selesai yang lebih awal daripada tanggal mulai kerja.<br>3. Klik tombol 'Simpan'. | Sistem otomatis mencegah pemilihan tanggal tersebut, menandai kolom tanggal dengan garis merah, dan menampilkan pesan peringatan bahwa tanggal selesai tidak boleh sebelum tanggal mulai. | Sesuai — Sistem menolak pengisian tanggal yang tidak logis dan meminta perbaikan rentang waktu. | Pass / - |
| **AS-UAT-004-03** | Memperbarui Catatan Perkembangan (Progres) Mitigasi dan Menilai Penurunan Risiko | Rencana Mitigasi: 'Pengadaan Server Cadangan', Catatan Perkembangan: 'Fase 1 uji coba sistem cadangan sukses di lingkungan simulasi, menunggu jadwal peluncuran', Status: 'Sedang Berjalan' | 1. Klik tombol 'Ubah / Edit' pada salah satu baris rencana mitigasi.<br>2. Tuliskan catatan perkembangan terbaru pada kolom keterangan progres.<br>3. Simpan perubahan data.<br>4. Periksa apakah tingkat keparahan risiko pada Profil Risiko dapat diturunkan setelah mitigasi berjalan. | Catatan perkembangan terbaru berhasil disimpan lengkap dengan riwayat pembaruan, dan tim pemeriksa dapat mengevaluasi penurunan tingkat keparahan risiko. | Sesuai — Perkembangan kerja berhasil dicatat dan riwayat pembaruan tersimpan dengan rapi. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-004 — Risk Mitigation Plan
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-005 — Risk Appetite Statement
**Tujuan:** Memastikan batas toleransi risiko organisasi (Risk Appetite & Tolerance) dapat ditetapkan per kategori risiko, dimonitor tingkat kepatuhannya, dan dideteksi pelanggarannya secara otomatis.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-005-01** | Melihat Indikator Tingkat Kepatuhan Selera Risiko (Risk Appetite) Perusahaan | Halaman Selera Risiko, Indikator: Persentase Kepatuhan (%), Total Risiko Terdaftar, Risiko Dalam Batas Wajar, dan Jumlah Pelanggaran Batas Toleransi | 1. Buka menu Batas Selera Risiko (Risk Appetite).<br>2. Perhatikan 4 kartu ringkasan angka di bagian paling atas halaman.<br>3. Cocokkan angka persentase kepatuhan dengan jumlah risiko yang melampaui batas toleransi. | Kartu ringkasan menyajikan persentase kepatuhan secara jelas dengan warna hijau (jika patuh) atau merah (jika ada pelanggaran), dan total angka risiko sesuai dengan data profil risiko. | Sesuai — Seluruh kartu angka ringkasan menyajikan data kepatuhan yang akurat dan mudah dipahami. | Pass / - |
| **AS-UAT-005-02** | Mengatur Ambang Batas Maksimal Risiko yang Boleh Ditoleransi per Kategori | Kategori: 'Kepatuhan Hukum & Regulasi', Batas Toleransi: 'Toleransi Nol (Skor Maksimal: 4)', Batas Peringatan: Skor 6 | 1. Klik tab menu 'Panduan & Aturan Batas' pada modul Selera Risiko.<br>2. Evaluasi matriks batas toleransi untuk area Aman (Hijau), Hati-Hati (Kuning), dan Bahaya/Pelanggaran (Merah).<br>3. Simpan aturan ambang batas. | Aturan batas toleransi tersimpan dengan jelas dan menjadi acuan otomatis bagi sistem untuk memperingatkan risiko yang melebihi batas wajar. | Sesuai — Batas toleransi risiko berhasil disimpan dan langsung digunakan sistem sebagai panduan penilaian otomatis. | Pass / - |
| **AS-UAT-005-03** | Peringatan Otomatis untuk Risiko yang Melebihi Batas Toleransi (Pelanggaran Risiko) | Risiko dengan nilai keparahan 20 pada kategori yang batas toleransinya maksimal 12 | 1. Buka daftar 'Pelanggaran Selera Risiko' atau saring risiko yang melebihi batas.<br>2. Periksa apakah risiko yang bersangkutan otomatis masuk ke dalam daftar pelanggaran.<br>3. Perhatikan tanda peringatan visual pada baris risiko tersebut. | Sistem secara otomatis mendeteksi risiko yang melampaui toleransi, memasukkannya ke daftar pelanggaran, dan menandainya dengan badge merah bertuliskan 'Melebihi Toleransi'. | Sesuai — Risiko yang melebihi ambang batas langsung tersorot dengan penanda merah peringatan yang mudah dikenali. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-005 — Risk Appetite Statement
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-006 — Risk Factors
**Tujuan:** Memastikan pembobotan faktor risiko korporat (Risk Factors Weighting) divalidasi dengan total persentase tepat 100%, kriteria skoring dapat dikonfigurasi, dan skor tertimbang entitas terhitung akurat.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-006-01** | Menyetel Pembobotan Faktor Penentu Risiko Korporat hingga Tepat Berjumlah 100% | Faktor Risiko: Kompleksitas Bisnis (30%), Nilai Aset Transaksi (25%), Temuan Audit Terdahulu (20%), Kualitas Sistem Kontrol (15%), Pergantian SDM (10%) -> Total Tepat: 100% | 1. Buka menu Faktor Penentu Risiko.<br>2. Buka tab 'Bobot Korporat'.<br>3. Atur proporsi persentase seluruh faktor risiko hingga totalnya berjumlah tepat 100%.<br>4. Periksa indikator status. | Kotak status menampilkan warna hijau dengan tulisan 'Total Bobot: 100% - Total bobot valid', tombol 'Simpan Bobot' menjadi aktif, dan konfigurasi bobot berhasil disimpan. | Sesuai — Sistem memastikan total persentase tepat 100% dan mengizinkan penyimpanan konfigurasi dengan sukses. | Pass / - |
| **AS-UAT-006-02** | Menolak Penyimpanan jika Total Pembobotan Kurang atau Lebih dari 100% | Pengaturan bobot dengan akumulasi total sementara 95% atau 105% | 1. Ubah salah satu nilai persentase faktor risiko sehingga total keseluruhan menjadi 95%.<br>2. Periksa perubahan warna pada kotak status peringatan.<br>3. Coba lakukan klik pada tombol 'Simpan Bobot'. | Kotak status berubah warna menjadi merah dengan pesan peringatan bahwa total bobot harus tepat 100%, dan tombol 'Simpan Bobot' otomatis terkunci (tidak bisa diklik). | Sesuai — Sistem mengunci tombol simpan dan memberi peringatan yang jelas hingga total bobot disesuaikan menjadi 100%. | Pass / - |
| **AS-UAT-006-03** | Menilai Faktor Risiko pada Suatu Unit Kerja dan Menghitung Nilai Akhir Otomatis | Unit Kerja: 'Direktorat Perbankan Digital', Nilai Parameter: Faktor 1 (Nilai 4), Faktor 2 (Nilai 5), Faktor 3 (Nilai 3), skala penilaian 1 sampai 5 | 1. Buka tab katalog penilaian faktor risiko unit kerja.<br>2. Berikan nilai skala 1 sampai 5 untuk masing-masing faktor risiko pada unit kerja yang dipilih.<br>3. Klik tombol 'Simpan Penilaian'. | Nilai akhir terbobot terhitung otomatis dengan mengalikan setiap nilai dengan bobot persentasenya, menghasilkan skor risiko bawaan yang akurat bagi unit kerja tersebut. | Sesuai — Perhitungan nilai akhir tertimbang berjalan presisi dan hasilnya tersimpan dengan benar. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-006 — Risk Factors
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-007 — Audit Universe
**Tujuan:** Memastikan struktur semesta audit (Audit Universe) yang mencakup entitas induk dan sub-entitas dapat dikelola hierarkinya, dipilih dari pustaka standar, serta diprioritaskan untuk penyusunan rencana audit.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-007-01** | Memilih Unit dan Bagian Kerja dari Daftar Pustaka Standar Organisasi | Unit Induk: 'Direktorat Operasional', Sub-Bagian: 'Unit Kliring', 'Unit Kas & Khazanah', 'Unit Layanan Nasabah' | 1. Masuk ke menu Semesta Audit (Audit Universe).<br>2. Buka tab 'Daftar Standar Unit Kerja'.<br>3. Jelajahi bagan hierarki struktur organisasi.<br>4. Beri centang pada kotak unit induk dan beberapa sub-bagian di bawahnya. | Kotak centang unit dan bagian kerja terpilih dengan rapi, dan seluruh bagian siap didaftarkan ke dalam semesta audit aktif korporat. | Sesuai — Struktur pohon hierarki unit kerja tampil rapi dan dapat dipilih secara interaktif oleh penguji. | Pass / - |
| **AS-UAT-007-02** | Mendaftarkan Unit atau Sub-Bagian Kerja Baru (Unit Kustom) | Nama Unit Induk: 'Divisi Keamanan Siber & Perlindungan Data', Nama Bagian: 'Pusat Operasi Keamanan (SOC)', Kategori: 'Teknologi & Operasional' | 1. Klik tombol '+ Tambah Unit Kerja Baru'.<br>2. Masukkan nama unit kerja, uraian fungsi bisnis, serta relasi bagian di bawahnya.<br>3. Klik tombol 'Simpan Unit'. | Unit kerja kustom baru berhasil disimpan, muncul pada bagan struktur organisasi semesta audit, dan siap untuk diberi pembobotan risiko. | Sesuai — Unit dan sub-bagian kerja baru berhasil tersimpan dan langsung terindeks dalam daftar Audit Universe. | Pass / - |
| **AS-UAT-007-03** | Mengurutkan Prioritas Pemeriksaan Unit Kerja Berdasarkan Tingkat Risikonya | Kumpulan unit kerja aktif dengan variasi skor risiko (Tinggi, Sedang, Rendah) | 1. Buka menu 'Prioritas Audit'.<br>2. Klik tombol pengurutan dari unit dengan nilai risiko tertinggi ke terendah.<br>3. Periksa saran jadwal pemeriksaan (contoh: Risiko Tinggi diaudit setiap tahun). | Daftar unit kerja terurut secara akurat dari yang paling berisiko ke yang paling aman, dan sistem memberikan rekomendasi frekuensi pemeriksaan audit yang sesuai standar. | Sesuai — Pemeringkatan prioritas audit berjalan otomatis dan menjadi dasar input bagi Rencana Audit Tahunan (PKAT). | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-007 — Audit Universe
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-008 — Risk Control Matrix
**Tujuan:** Memastikan matriks risiko dan pengendalian internal (RCM) berbasis 5 dimensi COSO 2013 dapat dikelola, diukur skor efektivitas kontrolnya, dan dibandingkan antara risiko inheren dan residual.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-008-01** | Menambahkan Catatan Risiko dan Sistem Pengendalian Internal Baru (RCM) | Divisi: 'Keuangan & Akuntansi', Proses: 'Pengeluaran Kas & Pembayaran Tagihan Vendor', Potensi Risiko: 'Pembayaran ganda atau tagihan fiktif', Sistem Kendali: 'Pemeriksaan 3 dokumen pencocokan (PO, Faktur, Berita Acara)', Jenis Kendali: 'Pencegahan (Preventive)', Frekuensi: 'Harian' | 1. Buka menu Matriks Pengendalian Risiko (Risk Control Matrix).<br>2. Klik tombol '+ Tambah Matriks Kendali'.<br>3. Isi lengkap seluruh formulir dari nama proses, potensi bahaya, hingga sistem kendali pencegahannya.<br>4. Klik tombol 'Simpan'. | Baris matriks kendali baru berhasil tersimpan, muncul pada tabel kendali divisi keuangan, dan langsung mempengaruhi nilai perhitungan efektivitas kendali. | Sesuai — Data matriks kendali berhasil disimpan dengan relasi yang lengkap dan informasi tersaji rapi. | Pass / - |
| **AS-UAT-008-02** | Menghitung Persentase Efektivitas Pengendalian Internal Secara Otomatis | Tahun: 2026, Divisi: 'Operasional Cabang', Data Uji: Sampel kendali yang berjalan efektif vs kendali yang belum berjalan | 1. Pilih tahun dan divisi pada menu penyaring di atas tabel.<br>2. Perhatikan kartu 'Tingkat Efektivitas Pengendalian Internal'.<br>3. Amati perubahan angka persentase saat status salah satu kendali diubah. | Skor efektivitas terhitung otomatis dalam bentuk persentase (%), kartu menampilkan label predikat mutu yang sesuai (Sangat Efektif / Perlu Perbaikan / Tidak Efektif), dan penjelasan interpretasi mutu terbarui. | Sesuai — Persentase efektivitas kendali terhitung secara matematis dan label predikat mutu tampil akurat. | Pass / - |
| **AS-UAT-008-03** | Mengevaluasi Perbandingan Risiko Awal vs Risiko yang Tersisa Setelah Kendali Diterapkan | Total Risiko Awal: 18 Risiko, Total Risiko Tersisa Setelah Kendali: 4 Risiko | 1. Buka kartu evaluasi tutup buku akhir tahun pada halaman RCM.<br>2. Bandingkan perbandingan antara tingkat risiko awal dengan tingkat risiko yang tersisa setelah kendali dijalankan. | Sistem menyajikan diagram grafik perbandingan yang jelas, menunjukkan keberhasilan implementasi sistem kendali dalam menurunkan potensi masalah di perusahaan. | Sesuai — Angka risiko awal dan risiko tersisa tersinkronisasi sempurna dengan data Profil Risiko korporat. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-008 — Risk Control Matrix
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-009 — Strategic Audit Plan
**Tujuan:** Memastikan rencana audit strategis jangka panjang (3-5 tahun), visi, misi, dan tujuan strategis fungsi internal audit dapat dirumuskan, dimonitor progres tahunannya, dan diarsipkan.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-009-01** | Memperbarui Pernyataan Visi, Misi, dan Tujuan Strategis Internal Audit | Visi: 'Menjadi mitra terpercaya dan memberi nilai tambah dalam tata kelola dan manajemen risiko kelas dunia', Misi: 'Memberikan pemeriksaan objektif dan konsultasi independen berbasis risiko dan teknologi', Jangka Waktu: 2025 - 2029 | 1. Masuk ke menu Rencana Strategis Audit.<br>2. Klik tombol 'Ubah' pada kartu Visi, Misi & Sasaran.<br>3. Ketikkan kalimat visi, misi, dan target strategis fungsi audit.<br>4. Klik tombol 'Simpan'. | Pernyataan visi dan misi berhasil diperbarui, tersimpan di sistem, dan langsung terpampang pada kartu informasi strategis di beranda rencana audit. | Sesuai — Visi dan misi terupdate seketika dan dapat dibaca oleh seluruh pihak yang berkepentingan. | Pass / - |
| **AS-UAT-009-02** | Menambahkan Sasaran Rencana Audit Strategis Jangka Panjang (Multi-Tahun) | Nama Sasaran: 'Transformasi Audit Digital & Otomasi Pemeriksaan Berkelanjutan', Pilar Strategis: 'Teknologi & Inovasi', Periode Target: '2026 - 2028', Tolok Ukur Keberhasilan: '100% integrasi data transaksi inti perbankan' | 1. Klik tombol '+ Tambah Sasaran Strategis'.<br>2. Isi formulir nama sasaran strategis, indikator kinerja, dan target tahun pencapaian.<br>3. Klik tombol 'Simpan Sasaran'. | Sasaran strategis tersimpan di tabel rencana jangka panjang, menampilkan indikator perkembangan progres, dan dapat dihubungkan ke sasaran rencana tahunan (PKAT). | Sesuai — Sasaran strategis multi-tahun berhasil ditambahkan ke dalam basis data rencana strategis audit. | Pass / - |
| **AS-UAT-009-03** | Mengunggah dan Mengarsipkan Dokumen Cetak Biru Rencana Strategis (PDF) | Halaman Impor Rencana Strategis, Berkas: Rencana_Strategis_Audit_2025_2029.pdf (Ukuran: 4.5 MB), Judul: 'Grand Strategy Internal Audit 5 Tahun' | 1. Klik tombol 'Impor Dokumen' pada modul Rencana Strategis Audit.<br>2. Masukkan judul dokumen dan unggah berkas cetak biru PDF dari komputer.<br>3. Klik tombol 'Mulai Impor Dokumen'. | Dokumen cetak biru berhasil diunggah, terdaftar pada tabel arsip rencana strategis, dan berkas dapat diunduh kembali dengan utuh kapan saja dibutuhkan. | Sesuai — Berkas terunggah ke arsip dokumen strategis dan tautan berkas berfungsi normal saat diunduh. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-009 — Strategic Audit Plan
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-010 — KPI Performance
**Tujuan:** Memastikan indikator kinerja utama (KPI) fungsi internal audit, kepuasan auditee, ketepatan waktu laporan audit, dan ekspor laporan kinerja berkala dapat dikelola dan dievaluasi secara terukur.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-010-01** | Memantau 4 Indikator Utama Capaian Kinerja Fungsi Audit Internal | Tahun Penugasan: '2026', Indikator Utama: Tingkat Penyelesaian Audit, Ketepatan Waktu Laporan, Nilai Kepuasan Auditee, dan Persentase Rekomendasi Selesai | 1. Masuk ke menu Kinerja Audit (KPI Performance).<br>2. Pilih pilihan tahun penugasan 2026.<br>3. Periksa angka capaian pada masing-masing kartu ringkasan kinerja. | Kartu menampilkan nilai capaian aktual, target yang telah ditetapkan, penanda warna pencapaian (hijau jika tercapai, merah jika belum), dan rincian capaian per triwulan. | Sesuai — Nilai 4 indikator utama KPI termuat lengkap dengan data performa audit secara langsung dan akurat. | Pass / - |
| **AS-UAT-010-02** | Melihat Rangkuman Hasil Evaluasi Survei Kepuasan dari Divisi yang Diaudit | Kuesioner Kepuasan: 5 Aspek Evaluasi (Objektivitas, Profesionalisme, Kualitas Saran Perbaikan, Komunikasi Tim, Nilai Tambah) | 1. Buka tab 'Survei Kepuasan Auditee' pada menu Kinerja Audit.<br>2. Amati skor rata-rata indeks kepuasan dari auditee yang telah mengisi survei pasca-audit.<br>3. Periksa grafik tren kepuasan dari waktu ke waktu. | Skor rata-rata kepuasan auditee terkalkulasi otomatis (dalam skala 1-5 atau persentase), menampilkan grafik tren kepuasan, dan teragregasi ke dalam nilai kinerja tahunan tim audit. | Sesuai — Rangkuman survei kepuasan auditee tercermin secara akurat pada metrik kinerja divisi audit. | Pass / - |
| **AS-UAT-010-03** | Mengunduh Laporan Kinerja Audit Internal ke Format PDF Resmi Siap Cetak | Tahun Laporan: '2026', Opsi: Cetak Laporan PDF Resmi | 1. Pada dashboard Kinerja Audit, klik tombol 'Ekspor PDF'.<br>2. Amati proses pembuatan dokumen dan proses unduh berkas.<br>3. Buka file PDF yang dihasilkan di komputer. | Muncul notifikasi proses pembuatan berkas, komputer mengunduh berkas PDF resmi lengkap dengan kop surat Divisi Audit Internal, ringkasan capaian kinerja, grafik, dan tanggal laporan. | Sesuai — Berkas PDF berhasil diunduh dengan tata letak dokumen laporan eksekutif formal yang bersih dan siap cetak. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-010 — KPI Performance
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-011 — Annual Audit Plan
**Tujuan:** Memastikan penyusunan, perhitungan mandays, penugasan supervisor, dan persetujuan bertingkat Program Kerja Audit Tahunan (PKAT) berjalan sesuai prosedur.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-011-01** | Membuat Rencana Audit Tahunan (PKAT) Baru | Tahun: 2026, Jenis Penugasan: 'Assurance', Divisi Sasaran: 'Teknologi Informasi', Jumlah Auditor: 3 Orang, Hari Kerja per Auditor: 10 Hari, Jadwal Bulan: Januari s/d Maret, Supervisor: Budi Santoso | 1. Buka menu Rencana Audit Tahunan (Annual Audit Plan).<br>2. Klik tombol '+ Tambah Rencana Audit'.<br>3. Isi data penugasan, divisi sasaran, jadwal bulan pelaksanaan, dan alokasi auditor.<br>4. Klik tombol 'Simpan'. | Sistem otomatis menghitung total hari kerja (30 hari kerja), menandai jadwal triwulan Q1, menerbitkan nomor kode unik otomatis (contoh: PKAT-2026-ASR-001), dan menyimpan data dengan status 'Draf'. | Sesuai — Rencana audit tahunan baru berhasil dibuat, total hari kerja dan kode rencana otomatis terhitung dengan tepat. | Pass / - |
| **AS-UAT-011-02** | Peringatan Beban Kerja Tim Audit jika Melebihi Batas Kapasitas (>95%) | Jumlah Auditor: 15 Orang, Hari Kerja per Auditor: 25 Hari, Total Beban: 375 Hari Kerja, Kapasitas Tim Maksimal: 300 Hari Kerja | 1. Buka formulir pembuatan rencana audit tahunan.<br>2. Masukkan jumlah auditor dan hari kerja yang melebihi kapasitas tahunan (>95%).<br>3. Amati indikator beban kerja pada formulir. | Sistem menampilkan tanda peringatan warna merah bertuliskan 'Beban Kerja Berlebih (>95%)' dan menyarankan penyesuaian alokasi agar tim tidak kelebihan beban. | Sesuai — Indikator kapasitas berubah warna merah dan memunculkan notifikasi peringatan beban kerja tim. | Pass / - |
| **AS-UAT-011-03** | Mencegah Penghapusan Rencana Audit yang Sudah Memiliki Agenda Pemeriksaan Lapangan | Kode Rencana: 'PKAT-2026-001' (Sudah memiliki jadwal kegiatan audit di lapangan yang aktif) | 1. Pilih baris rencana audit yang pemeriksaannya sudah mulai berjalan di lapangan.<br>2. Klik ikon 'Hapus' (gambar tempat sampah).<br>3. Perhatikan dialog konfirmasi yang muncul di layar. | Sistem mendeteksi bahwa data sedang digunakan dan memunculkan pesan aman: 'Rencana audit ini sudah digunakan pada kegiatan lapangan dan tidak dapat dihapus permanen. Apakah Anda ingin menonaktifkannya?' Status berubah menjadi Non-Aktif. | Sesuai — Muncul dialog konfirmasi penonaktifan dan rencana diubah menjadi non-aktif dengan aman tanpa menghapus data riwayat. | Pass / - |
| **AS-UAT-011-04** | Alur Persetujuan Bertingkat Rencana Audit (Draf -> Ditinjau -> Disetujui Pimpinan) | Tingkatan Pengguna: Staf Auditor -> Manajer Audit -> Kepala Divisi Audit (Chief Audit Executive) | 1. Staf auditor mengklik tombol 'Ajukan Rencana Audit'.<br>2. Manajer masuk ke akunnya, meninjau catatan, lalu menyetujui.<br>3. Kepala Divisi Audit meninjau dan mengklik tombol 'Pengesahan Akhir'. | Status rencana bergerak bertahap dari 'Draf' -> 'Sedang Ditinjau' -> 'Menunggu Pengesahan' -> 'Disetujui', dan tombol edit otomatis terkunci setelah status Disetujui. | Sesuai — Tahapan persetujuan berjenjang berjalan lancar sesuai hak akses dan rekam jejak persetujuan tercatat di riwayat revisi. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-011 — Annual Audit Plan
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-012 — Audit Activity Plan
**Tujuan:** Memastikan pembuatan dan rincian rencana kegiatan audit berdasarkan PKAT dapat disusun, dijadwalkan, dan dialokasikan anggarannya.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-012-01** | Membuat Rencana Kegiatan Pemeriksaan yang Terhubung ke Dokumen Rencana Tahunan | Rencana Tahunan Acuan: PKAT-2026-001, Kode Kegiatan: ACT-2026-001, Judul: 'Audit Keuangan & Pengeluaran Kas Q1 2026', Divisi Sasaran: Keuangan, Estimasi Anggaran: Rp 45.000.000 | 1. Buka menu Rencana Kegiatan Audit.<br>2. Pilih nomor referensi rencana tahunan (PKAT).<br>3. Masukkan rincian kegiatan audit, divisi sasaran, jadwal pelaksanaan, dan perkiraan anggaran.<br>4. Klik tombol 'Simpan'. | Rencana kegiatan berhasil tersimpan, terhubung langsung ke dokumen rencana tahunan induk, dan otomatis berstatus 'Direncanakan'. | Sesuai — Data kegiatan audit berhasil dibuat dan terhubung rapi ke dokumen rencana tahunan induk. | Pass / - |
| **AS-UAT-012-02** | Menolak Pengisian Tanggal Selesai yang Lebih Awal daripada Tanggal Mulai | Tanggal Mulai: '2026-04-01', Tanggal Selesai: '2026-03-01' (Tanggal selesai sengaja diisi mundur) | 1. Buka formulir tambah rencana kegiatan.<br>2. Masukkan tanggal selesai yang lebih awal daripada tanggal mulai kerja.<br>3. Klik tombol 'Simpan'. | Sistem menampilkan pesan peringatan: 'Tanggal selesai tidak boleh lebih awal dari tanggal mulai kegiatan', dan formulir ditandai garis merah. | Sesuai — Sistem mencegah penyimpanan formulir dan menyorot kolom tanggal yang bermasalah dengan warna merah. | Pass / - |
| **AS-UAT-012-03** | Mencatat Alokasi Biaya Operasional dan Anggota Tim Auditor yang Ditugaskan | Jumlah Tim: 3 Orang Auditor, Total Hari Kerja: 45 Hari, Rincian Biaya: Transportasi & Penginapan Rp 35.000.000, Total Anggaran Disediakan: Rp 40.000.000 | 1. Buka tab 'Anggaran & Tim' pada rincian Rencana Kegiatan.<br>2. Masukkan rincian biaya transportasi, penginapan, dan pilih nama-nama personil auditor.<br>3. Simpan perubahan data. | Sistem menghitung selisih anggaran (Sisa Anggaran = Rp 5.000.000) dan mencatat daftar nama auditor yang siap ditugaskan. | Sesuai — Perhitungan alokasi anggaran akurat dan rekapitulasi anggota tim auditor tersimpan dengan lengkap. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-012 — Audit Activity Plan
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-013 — Assignment Letter
**Tujuan:** Memastikan pembuatan, penomoran otomatis, penugasan tim auditor (Lead & Anggota), serta penerbitan surat tugas resmi audit dapat dikelola.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-013-01** | Menerbitkan Surat Tugas Pemeriksaan Baru dengan Nomor Surat Otomatis | Kegiatan Terkait: ACT-2026-001, Supervisor: Budi Santoso, Ketua Tim: Dimas Setiawan, Anggota: Ahmad Fauzi & Siti Rahmawati, Periode Tugas: 15 Januari s/d 15 Februari 2026 | 1. Buka menu Surat Tugas (Assignment Letter).<br>2. Klik tombol '+ Terbitkan Surat Tugas'.<br>3. Pilih referensi kegiatan audit dan tentukan susunan tim pemeriksa.<br>4. Klik tombol 'Terbitkan'. | Nomor surat tugas terbit otomatis dengan susunan format resmi (contoh: ST/IA/2026/FIN/001), status berubah menjadi 'Diterbitkan', dan tim auditor menerima pemberitahuan penugasan. | Sesuai — Nomor surat resmi otomatis terbit dan tim auditor yang ditugaskan langsung mendapatkan notifikasi tugas baru. | Pass / - |
| **AS-UAT-013-02** | Peringatan jika Ada Anggota Tim yang Jadwal Penugasannya Bersamaan di Tempat Lain | Nama Auditor: Siti Rahmawati, Jadwal Bentrok: 20 Januari s/d 30 Januari 2026 pada 2 surat tugas yang berbeda | 1. Buat draf surat tugas baru.<br>2. Masukkan nama auditor yang sedang aktif bertugas di penugasan lain pada tanggal yang sama.<br>3. Klik tombol 'Simpan'. | Sistem menampilkan kotak peringatan berwarna kuning: 'Perhatian: Auditor Siti Rahmawati sudah memiliki jadwal tugas aktif pada rentang tanggal tersebut'. | Sesuai — Peringatan bentrok jadwal muncul dengan jelas sehingga pimpinan dapat mengatur ulang penugasan personel. | Pass / - |
| **AS-UAT-013-03** | Mengunduh Dokumen Surat Tugas Resmi dalam Format PDF / Word Siap Cetak | Surat Tugas Terbit: ST-2026-001, Format Berkas: PDF / Word Resmi Siap Cetak | 1. Buka detail surat tugas yang berstatus telah disahkan.<br>2. Klik tombol 'Unduh Surat Tugas Resmi'.<br>3. Buka berkas hasil unduhan di komputer penguji. | Berkas dokumen resmi terunduh lengkap dengan kop surat perusahaan, susunan tim pemeriksa, serta barcode pengesahan resmi. | Sesuai — Dokumen surat tugas berhasil diunduh dalam format yang rapi dan siap dicetak sebagai kelengkapan dinas. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-013 — Assignment Letter
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-014-01 — Audit Field Work - Interview
**Tujuan:** Memastikan aktivitas pelaksanaan audit bagian wawancara dapat dicatat dan dikelola.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-014-01-01** | Mencatat Jadwal Wawancara dan Daftar Pihak yang Diwawancarai | Surat Tugas: ST-2026-001, Pihak yang Diwawancara: Siti Aminah (Manajer Keuangan), Pewawancara: Dimas Setiawan, Waktu: 20 Januari 2026 pukul 10:00 WIB, Topik: Prosedur Otorisasi Pengeluaran Kas | 1. Buka menu Pemeriksaan Lapangan -> Wawancara.<br>2. Klik tombol '+ Tambah Jadwal Wawancara'.<br>3. Isi nama pihak yang diwawancara, pewawancara, tanggal, tautan rapat virtual / ruangan, dan tujuan wawancara.<br>4. Klik tombol 'Simpan'. | Jadwal wawancara baru berhasil dicatat pada kalender kerja audit dengan label status 'Terjadwal'. | Sesuai — Jadwal wawancara baru berhasil disimpan dan muncul pada agenda kalender kerja tim audit. | Pass / - |
| **AS-UAT-014-01-02** | Mengunggah Dokumen Notulensi Hasil Wawancara dan Bukti Rekaman | Berkas: Notulensi_Wawancara_Finance_Q1.docx (Ukuran: 1.2 MB), Tautan Berkas Tambahan: folder arsip bukti wawancara | 1. Buka detail wawancara yang telah selesai dilaksanakan.<br>2. Klik tombol 'Unggah Notulensi / Bukti'.<br>3. Masukkan ringkasan poin tanya-jawab penting dan lampirkan berkas notulensi.<br>4. Klik tombol 'Simpan'. | Berkas tersimpan aman di arsip sistem, tautan dokumen aktif dan bisa dibuka, serta ringkasan tanya jawab terhubung ke kertas kerja auditor. | Sesuai — Berkas notulensi berhasil diunggah dengan aman dan terhubung langsung ke lembar kerja audit terkait. | Pass / - |
| **AS-UAT-014-01-03** | Memperbarui Status Wawancara Menjadi Selesai dan Menandai Potensi Temuan | Agenda: Wawancara Manajer Keuangan, Status Baru: 'Selesai Dilaksanakan', Catatan: Ditemukan proses persetujuan manual yang belum tercatat di sistem komputer | 1. Buka agenda wawancara yang telah terlaksana.<br>2. Ubah status menjadi 'Selesai'.<br>3. Beri tanda centang pada opsi 'Tandai Potensi Temuan Masalah'.<br>4. Tuliskan kesimpulan ringkas lalu klik 'Perbarui'. | Status terbarui menjadi hijau 'Selesai', dan tanda potensi temuan masalah otomatis diteruskan ke daftar telaah temuan tim audit. | Sesuai — Status berhasil diperbarui dan catatan potensi temuan berhasil diteruskan ke tahap analisis temuan. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-014-01 — Audit Field Work - Interview
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-014-02 — Audit Field Work - Observation
**Tujuan:** Memastikan jadwal, lokasi, objek observasi fisik/lapangan, dan pencatatan fakta temuan observasi dapat didokumentasikan secara sistematis.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-014-02-01** | Mencatat Rencana Pengamatan Fisik (Observasi) di Lapangan | Kegiatan Audit: ACT-2026-001, Lokasi: Gudang Logistik Utama Cakung, Objek Pengamatan: Penghitungan Fisik Stok Barang & Kondisi Keamanan CCTV, Tanggal: 22 Januari 2026 | 1. Buka menu Pemeriksaan Lapangan -> Observasi.<br>2. Klik tombol '+ Tambah Agenda Observasi'.<br>3. Isi detail lokasi gudang, tanggal, daftar auditor pelaksana, dan fokus pengamatan fisik.<br>4. Klik tombol 'Simpan'. | Jadwal observasi tersimpan rapi dengan label status 'Direncanakan' dan muncul pengingat di beranda kerja auditor. | Sesuai — Rencana observasi lapangan berhasil direkam dan masuk ke dalam agenda kegiatan tim audit. | Pass / - |
| **AS-UAT-014-02-02** | Menyimpan Catatan Fakta Lapangan dan Mengunggah Foto Bukti Kondisi Fisik | Foto Bukti: Foto_Kondisi_Gudang_Pintu_Rusak.jpg (Ukuran: 2.8 MB), Uraian Fakta: Pintu akses darurat terhalang tumpukan palet barang sehingga membahayakan jalur evakuasi | 1. Buka formulir pencatatan hasil observasi lapangan.<br>2. Unggah foto bukti observasi fisik dari komputer.<br>3. Tuliskan uraian kondisi nyata yang ditemukan di lokasi.<br>4. Klik tombol 'Simpan Hasil Observasi'. | Foto bukti tersimpan rapi di sistem, gambar kecil foto langsung terlihat jelas di layar, dan fakta catatan lapangan otomatis terhubung ke kertas kerja audit. | Sesuai — Foto bukti berhasil diunggah dengan tampilan yang jelas dan tercatat dalam lampiran bukti audit. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-014-02 — Audit Field Work - Observation
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-014-03 — Audit Field Work - Document Collection
**Tujuan:** Memastikan permintaan, pengumpulan, pelabelan, dan verifikasi dokumen bukti dari auditee dapat dikelola dan dilacak status pemenuhannya.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-014-03-01** | Mengajukan Daftar Permintaan Dokumen Bukti kepada Pihak Auditee | Daftar Dokumen yang Diminta: 1. Rekening Koran Des 2025, 2. SOP Pengadaan Barang v3, 3. Berita Acara Stok Barang; Batas Waktu: 25 Januari 2026 | 1. Buka menu Pemeriksaan Lapangan -> Pengumpulan Dokumen.<br>2. Klik tombol '+ Buat Permintaan Dokumen'.<br>3. Masukkan rincian dokumen yang dibutuhkan beserta batas tanggal penyerahannya.<br>4. Klik tombol 'Kirim ke Auditee'. | Daftar permintaan dokumen berhasil terkirim ke portal pihak auditee dengan status 'Menunggu Penyerahan Dokumen'. | Sesuai — Permintaan dokumen tercatat rapi dan pemberitahuan langsung terkirim ke penanggung jawab auditee. | Pass / - |
| **AS-UAT-014-03-02** | Memeriksa dan Menyetujui Dokumen Bukti yang Telah Diunggah Auditee | Dokumen yang Diperiksa: Rekening_Koran_Des2025.pdf, Peran Pemeriksa: Ketua Tim Auditor | 1. Buka daftar dokumen yang telah diunggah oleh pihak auditee.<br>2. Unduh dan periksa isi kelayakan dokumen.<br>3. Klik tombol hijau 'Verifikasi & Terima'.<br>4. Masukkan catatan verifikasi singkat. | Status dokumen berubah menjadi warna hijau 'Terverifikasi' dan dokumen resmi dijadikan bukti kerja pemeriksaan auditor. | Sesuai — Dokumen berhasil diverifikasi dan tersemat aman pada kertas kerja pemeriksaan terkait. | Pass / - |
| **AS-UAT-014-03-03** | Pelacakan Otomatis Dokumen yang Melewati Batas Waktu (Terlambat Diserahkan) | Tanggal Hari Ini: 26 Januari 2026, Batas Waktu Penyerahan: 24 Januari 2026 (Terlambat 2 hari) | 1. Buka tabel Pengumpulan Dokumen.<br>2. Pilih filter status 'Terlambat'.<br>3. Periksa label status dan perhitungan jumlah hari keterlambatannya.<br>4. Klik tombol 'Kirim Pengingat'. | Sistem otomatis menampilkan label merah bertuliskan 'Terlambat (2 Hari)' serta menyediakan tombol untuk mengirimkan pesan teguran pengingat ke auditee. | Sesuai — Perhitungan hari keterlambatan berjalan tepat dan label peringatan merah muncul secara otomatis. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-014-03 — Audit Field Work - Document Collection
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-014-04 — Audit Field Work - Sample Data
**Tujuan:** Memastikan metode penentuan sampel (random/judgmental), input data populasi, dan pencatatan parameter sampel pengujian audit berjalan akurat.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-014-04-01** | Menghitung Jumlah Sampel Pengujian Menggunakan Rumus Statistik Audit | Total Populasi: 1.250 lembar kuitansi kas, Tingkat Keyakinan: 95%, Batas Toleransi Kesalahan: 5%, Metode: Penarikan Sampel Acak | 1. Buka menu Pemeriksaan Lapangan -> Sampel Data.<br>2. Masukkan jumlah total seluruh transaksi dan parameter batas toleransi kesalahan.<br>3. Klik tombol 'Hitung Jumlah Sampel'. | Sistem otomatis menghasilkan rekomendasi jumlah sampel pengujian yang ideal (contoh: 60 sampel) dan tabel pemilihan transaksi acak. | Sesuai — Rumus perhitungan sampel menghasilkan ukuran sampel yang tepat sesuai kaidah standar audit internal. | Pass / - |
| **AS-UAT-014-04-02** | Menarik Transaksi Keuangan yang Mencurigakan (Deteksi Anomali Otomatis) | Kumpulan Data: Buku Besar Pengeluaran Kas Q4 2025, Metode: Deteksi Transaksi Janggal Berbasis Pola Anomali | 1. Buka tab 'Deteksi Anomali Transaksi'.<br>2. Pilih kumpulan data transaksi keuangan yang ingin diperiksa.<br>3. Klik tombol 'Jalankan Analisis Transaksi Janggal'.<br>4. Pilih 10 transaksi teratas yang dinilai paling mencurigakan oleh sistem. | Sistem menampilkan daftar transaksi dengan tingkat kejanggalan tinggi untuk dijadikan sampel pengujian terarah bagi auditor. | Sesuai — Sistem menyortir transaksi yang janggal dan langsung memasukkannya ke dalam daftar sampel pengujian tim auditor. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-014-04 — Audit Field Work - Sample Data
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-014-05 — Audit Field Work - Test Controls
**Tujuan:** Memastikan pengujian desain dan efektivitas operasi kontrol internal (RCM) serta penentuan deviasi/kegagalan kontrol dapat dievaluasi.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-014-05-01** | Menguji Kecukupan Desain Sistem Kendali Internal (Design Effectiveness) | Kendali Terkait: Persetujuan Ganda Pembayaran di Atas Rp 50 Juta, Parameter Uji: Kelengkapan Aturan Kebijakan dan Matriks Wewenang Persetujuan | 1. Masuk ke menu Uji Pengendalian (Test Controls).<br>2. Pilih sistem kendali dari matriks kendali (RCM).<br>3. Evaluasi apakah rancangan aturan kendali sudah memadai untuk mencegah potensi kecurangan.<br>4. Tentukan kesimpulan: 'Memadai' atau 'Tidak Memadai'. | Sistem menyimpan hasil evaluasi rancangan kendali beserta catatan analisis pertimbangan auditor. | Sesuai — Hasil evaluasi rancangan kendali berhasil disimpan dan terhubung ke ringkasan efektivitas kendali. | Pass / - |
| **AS-UAT-014-05-02** | Menguji Pelaksanaan Sistem Kendali di Lapangan (Operating Effectiveness) | Jumlah Sampel yang Diuji: 50 Transaksi Pengeluaran, Penyimpangan Ditemukan: 3 Transaksi Tanpa Tanda Tangan Pejabat Berwenang | 1. Buka lembar pengujian pelaksanaan kendali.<br>2. Masukkan total sampel yang diuji dan jumlah penyimpangan yang ditemukan di lapangan.<br>3. Tentukan kesimpulan akhir efektivitas kendali. | Persentase penyimpangan terhitung 6% (melebihi toleransi wajar 5%), sistem menyimpulkan status kendali 'Tidak Efektif', dan otomatis membuat draf temuan audit. | Sesuai — Sistem kendali disimpulkan tidak efektif dan draf catatan temuan audit otomatis masuk ke kertas kerja. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-014-05 — Audit Field Work - Test Controls
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-015-01 — Working Paper - Header
**Tujuan:** Memastikan informasi umum Kertas Kerja Audit (kode KKA, referensi surat tugas, auditee, periode audit, tim penyusun) tercatat valid.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-015-01-01** | Membuat Kertas Kerja Audit (KKA) Baru Berdasarkan Nomor Surat Tugas | Surat Tugas Acuan: ST-2026-001, Divisi: Keuangan, Tahun: 2026, Judul Pemeriksaan: 'Pemeriksaan Kas dan Bukti Pengeluaran' | 1. Masuk ke menu Kertas Kerja Audit (Working Paper).<br>2. Klik tombol '+ Buat KKA Baru'.<br>3. Pilih nomor surat tugas penugasan.<br>4. Lengkapi judul dan ruang lingkup pemeriksaan.<br>5. Klik tombol 'Simpan Lembar Kerja'. | Lembar KKA baru terbentuk dengan nomor registrasi unik (contoh: KKA/2026/FIN/001), nama tim auditor terisi otomatis dari surat tugas, dan berstatus awal 'Draf'. | Sesuai — Lembar KKA berhasil dibuat dan mewarisi data surat tugas tanpa perlu mengetik ulang secara manual. | Pass / - |
| **AS-UAT-015-01-02** | Mengunci Lembar KKA agar Tidak Bisa Diubah saat Berada pada Tahap Peninjauan | KKA ID: KKA-2026-001, Status Lembar Kerja: 'Sedang Ditinjau Ketua Tim', Akun: Staf Auditor | 1. Masuk menggunakan akun staf auditor.<br>2. Buka lembar KKA yang sedang berada pada tahap review ketua tim.<br>3. Coba ubah teks judul atau ruang lingkup pemeriksaan pada formulir. | Seluruh kotak isian berada dalam kondisi terkunci (hanya bisa dibaca / tidak bisa diedit) dan tombol simpan tidak aktif bagi staf auditor. | Sesuai — Kotak isian terkunci rapi dan sistem menolak permintaan perubahan saat lembar kerja dalam tahap review. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-015-01 — Working Paper - Header
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-015-02 — Working Paper - Risk Profile
**Tujuan:** Memastikan pemetaan risiko inheren, kontrol yang ada, dan risiko residual pada objek audit terintegrasi ke dalam kertas kerja.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-015-02-01** | Menghubungkan Profil Risiko Unit Kerja ke dalam Kertas Kerja Pemeriksaan | Divisi yang Diperiksa: Keuangan, Risiko Terkait: 'Penggelapan Dana Kas Operasional Kantor' | 1. Pada lembar KKA, buka tab 'Profil Risiko'.<br>2. Klik tombol 'Tarik Data dari Daftar Risiko'.<br>3. Periksa kecocokan uraian potensi risiko, sistem kendali yang ada, dan nilai risiko awalnya. | Uraian potensi bahaya dan sistem kendali internal yang terdaftar otomatis termuat lengkap ke dalam tabel kertas kerja audit. | Sesuai — Data risiko berhasil ditarik dan tampil lengkap pada lembar kertas kerja tanpa kendala. | Pass / - |
| **AS-UAT-015-02-02** | Menyesuaikan Tingkat Risiko Berdasarkan Temuan Fakta Lapangan | Nilai Awal: Risiko Sedang (Skor 3), Fakta di Lapangan: Sistem kendali tidak dijalankan sama sekali, Usulan Nilai Baru: Risiko Tinggi (Skor 5) | 1. Pada tab Profil Risiko di KKA, ubah penilaian efektivitas kendali menjadi 'Tidak Efektif'.<br>2. Masukkan usulan tingkat bahaya baru menjadi 'Tinggi'.<br>3. Tuliskan dasar temuan fakta lapangan pendukungnya lalu klik 'Simpan'. | Tingkat risiko yang tersisa otomatis terbarui menjadi 'Tinggi' dengan penanda warna merah dan memicu pembaruan pada peta risiko unit terkait. | Sesuai — Penyesuaian tingkat risiko berhasil disimpan dan indikator warna risiko langsung berubah secara otomatis. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-015-02 — Working Paper - Risk Profile
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-015-03 — Working Paper - Test Sample
**Tujuan:** Memastikan dokumentasi eksekusi pengujian pada sampel audit, langkah uji, dan bukti kerja terhubung dengan hasil verifikasi.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-015-03-01** | Mencatat Hasil Pengujian Rinci pada Setiap Baris Sampel Transaksi | Sampel Uji: 25 Bukti Transaksi Pengeluaran Kas, Aspek Uji: Kelengkapan Kuitansi, Faktur Pajak, dan Tanda Tangan Persetujuan Pejabat | 1. Buka tab 'Uji Sampel' pada lembar KKA.<br>2. Isi hasil verifikasi setiap baris sampel: 'Sesuai' atau 'Menyimpang'.<br>3. Tambahkan catatan penjelasan pada transaksi sampel yang bermasalah.<br>4. Klik tombol 'Simpan'. | Setiap item sampel memiliki status verifikasi yang jelas, dan sistem menghitung rekapitulasi kepatuhan (contoh: 22 Sesuai, 3 Menyimpang). | Sesuai — Baris data sampel tersimpan rapi dan perhitungan angka rekapitulasi kepatuhan muncul secara otomatis. | Pass / - |
| **AS-UAT-015-03-02** | Menautkan File Bukti Dokumen Pendukung ke Baris Sampel Pengujian | Nomor Baris Sampel: Sampel Ke-14, Berkas Bukti: Bukti_Kuitansi_Vendor_992.pdf | 1. Klik tombol 'Lampirkan Bukti' pada baris sampel nomor 14.<br>2. Pilih dokumen bukti dari arsip penyimpanan dokumen.<br>3. Klik tombol 'Simpan Tautan'.<br>4. Klik tautan nama berkas tersebut. | Tautan dokumen bukti aktif dan dapat diklik langsung untuk membuka dan memverifikasi dokumen pendukung secara instan. | Sesuai — Tautan dokumen bukti berfungsi dengan lancar dan membuka berkas secara langsung dari baris tabel pengujian. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-015-03 — Working Paper - Test Sample
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-015-04 — Working Paper - AOI & RCA
**Tujuan:** Memastikan pencatatan 5 atribut temuan audit (Condition, Criteria, Cause, Effect, Recommendation) dan analisis akar masalah (RCA) terdokumentasi lengkap.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-015-04-01** | Mencatat Temuan Masalah Audit Lengkap dengan Struktur 5 Unsur Standar Audit | Kondisi: Pembayaran tagihan vendor Rp 150 Juta tanpa persetujuan Direksi, Aturan: Prosedur Keuangan SOP pasal 5, Penyebab: Sistem komputer tidak memblokir batas transaksi, Akibat: Potensi kerugian keuangan perusahaan, Saran: Pemasangan validasi otomatis pada sistem | 1. Buka tab 'Temuan Masalah (AOI)' pada KKA.<br>2. Klik tombol '+ Tambah Temuan'.<br>3. Isi lengkap seluruh kotak penjelasan: Kondisi, Kriteria, Penyebab, Akibat, dan Saran Perbaikan.<br>4. Tentukan tingkat kepentingan temuan: 'Tinggi'.<br>5. Klik tombol 'Simpan Temuan'. | Temuan masalah tersimpan dengan nomor identifikasi unik (contoh: AOI-2026-FIN-001) dan seluruh unsur wajib terisi lengkap tanpa ada kolom yang kosong. | Sesuai — Format temuan lengkap 5 unsur berhasil tersimpan rapi dan siap dibawa ke tahap pembahasan dengan auditee. | Pass / - |
| **AS-UAT-015-04-02** | Menganalisis Akar Masalah Menggunakan Metode 5 Kali Bertanya Mengapa (5-Whys) | Mengapa 1: Sistem tidak mengunci transaksi -> Mengapa 2: Batas nominal diubah manual -> Mengapa 3: Tidak ada catatan riwayat perubahan -> Mengapa 4: Kurang supervisi pengawasan IT -> Mengapa 5: Belum ada SOP pemeriksaan rutin konfigurasi sistem | 1. Pada detail temuan masalah di KKA, klik 'Pembuat Analisis Akar Masalah (5-Whys)'.<br>2. Masukkan tingkatan alasan pertanyaan mengapa hingga ke penyebab terdalam.<br>3. Klik tombol 'Simpan Diagram Analisis'. | Alur 5-Whys tersimpan dalam bentuk bagan bertingkat yang rapi dan menjadi dasar rekomendasi perbaikan mendasar bagi manajemen. | Sesuai — Bagan alur analisis akar masalah tersimpan dengan jelas dan tersemat pada kertas kerja audit. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-015-04 — Working Paper - AOI & RCA
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-015-05 — Working Paper - Action Plan
**Tujuan:** Memastikan kesepakatan rencana aksi tindak lanjut antara auditor dan manajemen auditee beserta PIC dan target tanggal penyelesaian tersimpan valid.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-015-05-01** | Mencatat Komitmen Rencana Tindakan Perbaikan oleh Pihak Manajemen Auditee | Temuan Terkait: AOI-2026-FIN-001, Rencana Perbaikan: 'Penerapan validasi otomatis pada sistem dan pemeriksaan konfigurasi rutin bulanan', Penanggung Jawab (PIC): Manajer IT & Manajer Keuangan, Target Selesai: 31 Maret 2026 | 1. Buka tab 'Rencana Perbaikan' pada lembar KKA.<br>2. Tuliskan komitmen langkah perbaikan yang disepakati bersama pihak auditee.<br>3. Pilih nama pejabat penanggung jawab dan tentukan tanggal batas penyelesaiannya.<br>4. Klik tombol 'Simpan'. | Rencana perbaikan tersimpan rapi, berstatus 'Disepakati', dan otomatis terhubung ke modul pemantauan tindak lanjut (ATR). | Sesuai — Data komitmen perbaikan berhasil dicatat lengkap dan siap diteruskan ke laporan hasil audit. | Pass / - |
| **AS-UAT-015-05-02** | Memastikan Nama Penanggung Jawab dan Target Tanggal Selesai Wajib Diisi | Uraian Rencana: 'Akan dilakukan perbaikan sistem segera', Penanggung Jawab: Sengaja Dikosongkan, Target Tanggal: Sengaja Dikosongkan | 1. Coba simpan rencana perbaikan tanpa mengisi nama penanggung jawab dan tanpa tanggal target selesai.<br>2. Klik tombol 'Simpan'. | Sistem menolak penyimpanan dengan pesan peringatan: 'Nama Penanggung Jawab (PIC) dan Target Tanggal Penyelesaian wajib diisi sebelum lembar KKA disahkan'. | Sesuai — Formulir mencegah penyimpanan data rencana yang tidak lengkap dan memastikan setiap komitmen memiliki penanggung jawab yang jelas. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-015-05 — Working Paper - Action Plan
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-016 — Audit Result Report/LHA
**Tujuan:** Memastikan kompilasi seluruh temuan, opini audit, rating hasil audit, dan penerbitan draf/final LHA berjalan akurat dan siap didistribusikan.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-016-01** | Mengumpulkan Seluruh Temuan dari Semua Kertas Kerja Menjadi Naskah Draf LHA | Kegiatan Pemeriksaan: ACT-2026-001, Rekapitulasi Temuan KKA: 3 Temuan Tinggi, 2 Temuan Sedang, 1 Temuan Rendah | 1. Buka menu Laporan Hasil Audit (LHA).<br>2. Klik tombol '+ Buat Draf LHA' dari kegiatan audit yang dipilih.<br>3. Periksa rekapitulasi temuan yang ditarik otomatis oleh sistem. | Semua temuan dari seluruh lembar kertas kerja terkumpul secara otomatis ke dalam naskah draf LHA beserta tabel rekapitulasi temuan tanpa ada yang terlewat. | Sesuai — Naskah draf laporan tersusun lengkap dan akurat tanpa kehilangan satupun temuan dari kertas kerja. | Pass / - |
| **AS-UAT-016-02** | Menentukan Opini Audit dan Peringkat Mutu Pengendalian Internal | Peringkat Mutu: 'Perlu Perbaikan (Needs Improvement)', Opini Resmi: 'Wajar Dengan Pengecualian pada sistem kontrol pembayaran pengeluaran kas' | 1. Pada naskah draf LHA, pilih tab 'Opini & Peringkat Audit'.<br>2. Pilih predikat peringkat mutu dan ketikkan narasi opini audit resmi.<br>3. Klik tombol 'Simpan Draf Laporan'. | Opini audit resmi tersimpan dengan penandaan warna peringkat yang sesuai (Kuning untuk Perlu Perbaikan) dan tampil di bagian awal laporan. | Sesuai — Opini dan peringkat mutu audit berhasil tersimpan dengan baik pada dokumen laporan. | Pass / - |
| **AS-UAT-016-03** | Menerbitkan LHA Final dan Mengunduh Dokumen Laporan Resmi yang Terproteksi | Status Dokumen: Telah Disetujui Kepala Divisi Audit, Format Berkas: PDF Resmi dengan Cap Watermark Rahasia | 1. Buka dokumen LHA yang telah disahkan pimpinan audit.<br>2. Klik tombol 'Terbitkan & Unduh PDF Resmi'.<br>3. Buka berkas PDF hasil unduhan di komputer. | Berkas PDF resmi terunduh lengkap dengan nomor registrasi LHA, daftar distribusi rahasia, cap tanda air resmi, dan terlindungi dari pengubahan isi. | Sesuai — Dokumen PDF terunduh dengan cap watermark resmi dan tata letak laporan eksekutif formal yang bersih. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-016 — Audit Result Report/LHA
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-017 — Executive Summary Individual
**Tujuan:** Memastikan ringkasan eksekutif laporan audit per penugasan untuk Direksi dan Komite Audit tersaji ringkas, jelas, dan memuat poin-poin kritikal.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-017-01** | Membuat Ringkasan Eksekutif Satu Halaman untuk Laporan Tingkat Direksi | Nomor LHA Terkait: LHA-2026-001, Komponen: Sorotan Temuan Paling Kritis, Dampak Risiko Bisnis, Opini Audit, dan Rekomendasi Utama | 1. Buka menu Ringkasan Eksekutif -> Laporan Per Penugasan.<br>2. Pilih nomor laporan audit acuan.<br>3. Periksa ringkasan yang dihasilkan otomatis oleh sistem.<br>4. Klik tombol 'Simpan Ringkasan'. | Sistem menyajikan ringkasan satu halaman yang padat dan jelas, memuat matriks risiko dan temuan utama yang siap dibaca oleh Dewan Direksi dan Komite Audit. | Sesuai — Tampilan ringkasan satu halaman tersaji rapi dengan grafik ringkas dan poin-poin keputusan eksekutif. | Pass / - |
| **AS-UAT-017-02** | Mengunduh Ringkasan Eksekutif ke Format Dokumen Siap Bahan Rapat Direksi | Format Berkas: PDF / Word Tata Letak Eksekutif Siap Presentasi | 1. Buka ringkasan eksekutif yang telah dibuat.<br>2. Klik tombol 'Unduh Dokumen Eksekutif'.<br>3. Periksa kerapian tata letak dokumen yang terunduh. | Dokumen terunduh dalam format nota dinas eksekutif dengan jenis huruf profesional, bersih, dan ringkas tanpa rincian teknis yang rumit. | Sesuai — Berkas memo eksekutif berhasil diunduh dan siap digunakan sebagai materi rapat Komite Audit. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-017 — Executive Summary Individual
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-018 — Executive Summary Compliance
**Tujuan:** Memastikan rekapitulasi tingkat kepatuhan regulasi/kebijakan, tren temuan kepatuhan, dan profil risiko kepatuhan teragregasi dengan benar.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-018-01** | Melihat Rekapitulasi Tingkat Kepatuhan Terhadap Regulasi dan Standar Industri | Kategori Aturan: Standar ISO 27001, Regulasi OJK / Bank Indonesia, UU Perlindungan Data Pribadi, Aturan Pajak; Periode: Tahun 2026 | 1. Buka menu Ringkasan Eksekutif -> Kepatuhan Regulasi.<br>2. Pilih periode tahun pemeriksaan.<br>3. Amati rekapitulasi temuan berdasarkan pasal dan jenis regulasi. | Sistem menampilkan diagram visual sebaran temuan kepatuhan dan menandai bagian bisnis yang paling rawan terhadap sanksi atau denda dari regulator. | Sesuai — Grafik kepatuhan regulasi menampilkan persentase kepatuhan dan risiko ketidakpatuhan secara visual dengan jelas. | Pass / - |
| **AS-UAT-018-02** | Menyaring Laporan Kepatuhan Berdasarkan Unit Kerja / Divisi Tertentu | Pilihan Filter: Divisi Teknologi Informasi & Divisi Operasional Cabang | 1. Pada dashboard kepatuhan, pilih divisi yang ingin diperiksa.<br>2. Klik tombol 'Terapkan Filter'.<br>3. Perhatikan pembaruan angka pada grafik. | Data statistik dan daftar temuan kepatuhan menyaring secara cepat hanya untuk divisi yang dipilih dengan tepat. | Sesuai — Penyaringan data berfungsi responsif dan langsung memperbarui angka ringkasan kepatuhan. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-018 — Executive Summary Compliance
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-019 — Action Taken Report
**Tujuan:** Memastikan tindak lanjut atas rekomendasi hasil audit (temuan LHA) dapat dipantau, divalidasi bukti perbaikannya, dan dilacak status penyelesaiannya secara tepat waktu.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-019-01** | Menyaring dan Mencari Daftar Tindak Lanjut Rekomendasi Hasil Audit | Pilihan Divisi: 'Keuangan', Pilihan Status: 'Sedang Dikerjakan', Kata Kunci Pencarian: 'Rekonsiliasi' | 1. Buka menu Pemantauan Tindak Lanjut (Action Taken Report).<br>2. Pilih filter divisi 'Keuangan'.<br>3. Pilih filter status 'Sedang Dikerjakan'.<br>4. Masukkan kata kunci pencarian pada kotak pencarian. | Tabel daftar tindak lanjut secara instan menyaring data rekomendasi yang sesuai tanpa memuat ulang halaman, dan kartu statistik menampilkan jumlah temuan aktif yang tepat. | Sesuai — Tabel menyaring data secara cepat dan angka statistik pemantauan terupdate dengan akurat. | Pass / - |
| **AS-UAT-019-02** | Mengunggah Dokumen Bukti Perbaikan dan Menuliskan Progres Kerja oleh Auditee | Temuan Terkait: Surat Tugas Keuangan 2026, Berkas Bukti: Bukti_Rekonsiliasi_Kas_Harian.pdf (Ukuran: 1.8 MB), Uraian Progres: 'SOP rekonsiliasi harian telah disahkan dan mulai dijalankan sejak 10 April 2026' | 1. Klik tombol 'Lihat Detail' pada salah satu temuan audit.<br>2. Buka bagian Pembaruan Perkembangan Terkini.<br>3. Unggah berkas dokumen bukti perbaikan dari komputer.<br>4. Masukkan uraian progres kerja lalu klik tombol 'Simpan Progres'. | Berkas bukti perbaikan berhasil tersimpan di sistem, nama berkas tampil dengan lambang dokumen, dan uraian progres kerja terbarui pada lembar pemantauan. | Sesuai — Dokumen lampiran bukti berhasil diunggah dan uraian progres kerja tersimpan dengan benar. | Pass / - |
| **AS-UAT-019-03** | Memeriksa Bukti Perbaikan oleh Auditor dan Mengubah Status Menjadi Selesai | Temuan Tindak Lanjut: Rekonsiliasi Kas, Status Awal: 'Sedang Dikerjakan', Status Akhir: 'Selesai Dituntaskan', Penilai: Ketua Tim Auditor | 1. Masuk menggunakan akun auditor.<br>2. Buka detail temuan yang telah diserahkan buktinya oleh auditee.<br>3. Periksa kelayakan dan keabsahan dokumen bukti perbaikan.<br>4. Ubah status rekomendasi menjadi 'Selesai Dituntaskan'.<br>5. Klik tombol 'Simpan Verifikasi'. | Status rekomendasi berubah menjadi warna hijau bertuliskan 'Selesai Dituntaskan', dan persentase kepatuhan tindak lanjut di dashboard audit otomatis meningkat. | Sesuai — Rekomendasi dinyatakan tuntas, label status berubah menjadi hijau, dan angka persentase kepatuhan naik. | Pass / - |
| **AS-UAT-019-04** | Menghitung Hari Keterlambatan Secara Otomatis dan Mengirimkan Pesan Pengingat | Batas Waktu Penyelesaian: 31 Maret 2026 (Sudah terlewati), Status: Belum Selesai, Tanggal Hari Ini: 04 November 2026 | 1. Buka tabel Tindak Lanjut (ATR).<br>2. Pilih filter status 'Terlambat'.<br>3. Periksa label status dan selisih hari keterlambatannya.<br>4. Buka jendela rincian lalu klik tombol 'Kirim Pengingat'. | Sistem secara otomatis menghitung selisih hari keterlambatan, menampilkan label merah bertuliskan 'Terlambat (X Hari)', dan mengirimkan pesan pengingat ke penanggung jawab divisi. | Sesuai — Perhitungan keterlambatan berjalan otomatis dan tombol pengingat berfungsi normal mengirimkan notifikasi teguran. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-019 — Action Taken Report
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-020 — QA Review
**Tujuan:** Memastikan proses review kualitas penugasan audit oleh Supervisor/QAR dan kepatuhan terhadap standar audit IPPF/GIAS berjalan objektif.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-020-01** | Menilai Kepatuhan Mutu Pelaksanaan Audit Terhadap Standar Internasional (IPPF / GIAS) | Lembar Penilaian Mutu: 20 Kriteria Standar (Perencanaan Matang, Kecukupan Bukti, Supervisi Lapangan, Kualitas Laporan Akhir) | 1. Buka menu Jaminan Kualitas -> Penilaian Mutu (QA Review).<br>2. Pilih berkas penugasan audit yang telah selesai.<br>3. Isi lembar checklist kepatuhan standar mutu audit internasional.<br>4. Klik tombol 'Kirim Hasil Penilaian'. | Skor kepatuhan standar mutu terhitung otomatis (contoh: 92% - Sangat Memenuhi Standar) dan rekam evaluasi tersimpan di arsip kendali mutu audit. | Sesuai — Skor kepatuhan standar audit terhitung otomatis dengan predikat mutu yang sesuai. | Pass / - |
| **AS-UAT-020-02** | Memberikan Catatan Saran Pembinaan Kualitas (Coaching Notes) ke Tim Auditor | Catatan Bimbingan: 'Kertas kerja sampel nomor 12 perlu diperkuat dengan bukti konfirmasi langsung dari pihak bank' | 1. Masukkan catatan saran perbaikan pada kolom Catatan Pembinaan QA.<br>2. Klik tombol 'Kirim Catatan ke Tim Audit'. | Pemberitahuan catatan review langsung terkirim ke ketua tim auditor dan lembar kertas kerja audit terkait diberi tanda 'Perlu Perbaikan Tambahan'. | Sesuai — Notifikasi saran review masuk ke kotak pesan tim audit dan status lembar kerja berubah untuk disempurnakan. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-020 — QA Review
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Tabel AS-UAT-021 — Import Document
**Tujuan:** Memastikan dokumen audit eksternal/legacy (PKPT, Surat Tugas, KKA, dan LHA) dapat diimpor, divalidasi format berkas serta batas ukurannya, disimpan ke arsip digital, dan dikelola (unduh/hapus) secara aman.

| Test Case ID | Test Case | Test Data | Langkah Pengujian | Expected Result | Actual Result | Status / Bug ID |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **AS-UAT-021-01** | Mengimpor Dokumen Berkas Audit dari Luar Sistem ke Arsip Digital | Halaman: Rencana Audit Tahunan -> Impor Dokumen, Judul: 'Rencana Audit Tahunan 2026 Eksternal', Keterangan: 'Dokumen rencana kerja yang disahkan dewan komisaris', Berkas: Rencana_Audit_Eksternal_2026.pdf (Ukuran: 3.2 MB, berkas sah) | 1. Masuk ke menu Rencana Audit Tahunan dan klik tombol 'Impor Dokumen'.<br>2. Masukkan judul dokumen dan deskripsi singkat pada formulir.<br>3. Tarik dan lepas (drag & drop) berkas PDF ke dalam area kotak unggah berkas.<br>4. Klik tombol 'Mulai Impor Dokumen'. | Berkas berhasil diunggah dengan pesan sukses warna hijau 'Dokumen berhasil diimpor', berkas tersimpan aman di arsip digital, dan baris dokumen baru langsung muncul di tabel daftar dokumen terimpor. | Sesuai — Berkas berhasil diimpor dengan lancar, tersimpan aman di arsip digital, dan dokumen langsung terdaftar di tabel. | Pass / - |
| **AS-UAT-021-02** | Menolak Berkas Dokumen yang Melebihi Batas Ukuran Maksimal (>10 MB) | Judul: 'Arsip Lampiran KKA Konsolidasi Cabang', Berkas: Lampiran_KKA_Besar.docx (Ukuran: 14.8 MB, melebihi batas 10 MB) | 1. Pada formulir Impor Dokumen, masukkan judul dokumen.<br>2. Pilih berkas dokumen yang berukuran 14.8 MB ke kotak unggah berkas.<br>3. Klik tombol 'Mulai Impor Dokumen'. | Sistem otomatis menolak unggahan berkas, berkas tidak disimpan ke server, dan muncul pesan peringatan yang jelas: 'Ukuran file melebihi batas maksimum 10 MB. Silakan pilih berkas yang lebih kecil'. | Sesuai — Sistem secara aman memblokir berkas berukuran lebih dari 10 MB dan menampilkan pesan peringatan yang mudah dipahami. | Pass / - |
| **AS-UAT-021-03** | Memastikan Judul Dokumen Wajib Diisi dan Membatasi Panjang Judul Maksimal 100 Huruf | Uji Coba 1: Mengosongkan judul dokumen; Uji Coba 2: Mengetik judul sangat panjang (120 huruf); Berkas: Surat_Tugas_2026.pdf (1.2 MB) | 1. Pilih berkas pada formulir impor tetapi biarkan kolom Judul Dokumen kosong.<br>2. Perhatikan tombol 'Mulai Impor Dokumen'.<br>3. Coba ketikkan judul dokumen dengan panjang 120 huruf.<br>4. Perhatikan pembatasan jumlah ketikan huruf pada kolom judul. | Tombol 'Mulai Impor Dokumen' tetap terkunci (tidak bisa diklik) selama judul belum diisi, dan kolom judul otomatis membatasi ketikan maksimal 100 huruf agar nama dokumen tetap ringkas dan rapi. | Sesuai — Formulir mewajibkan pengisian judul dokumen dan mengunci panjang ketikan teks tepat pada batas 100 huruf. | Pass / - |
| **AS-UAT-021-04** | Mengunduh Berkas Asli dan Menghapus Dokumen yang Sudah Tidak Digunakan | Dokumen: Rencana_Audit_Eksternal_2026.pdf, Peran: Manajer Audit (Memiliki Hak Akses Lengkap) | 1. Buka tabel 'Daftar Dokumen Terimpor'.<br>2. Klik tombol 'Unduh Dokumen' (gambar ikon unduh).<br>3. Periksa berkas yang terunduh di komputer penguji.<br>4. Klik tombol 'Hapus Dokumen' (gambar ikon tempat sampah) dan konfirmasi pilihan 'Ya, Hapus'. | Berkas asli terunduh secara utuh dengan nama dan format yang sesuai. Saat konfirmasi hapus disetujui, sistem menghapus dokumen dari tabel daftar dan menghapus berkas fisiknya dari penyimpanan arsip. | Sesuai — Alur pengunduhan dokumen asli dan penghapusan dokumen terimpor berjalan sempurna dan tuntas. | Pass / - |

#### Lembar Pengesahan & Tanda Tangan Penguji Modul AS-UAT-021 — Import Document
| Peran Pengujian | Nama Penguji / Verifikator | Unit Kerja | Tanggal Uji | Status Keputusan | Tanda Tangan |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Penguji 1 (UAT Tester / Auditor)** | Ahmad Fauzi, S.E. | Satuan Kerja Audit Internal (SKAI) | 11/11/2026 | **[✓] PASSED** | *(Tanda Tangan Digital)* |
| **Penguji 2 (Process Owner / Auditee)** | Siti Rahmawati, S.Kom. | Unit Kerja Terkait (Auditee) | 11/11/2026 | **[✓] ACCEPTED** | *(Tanda Tangan Digital)* |
| **Verifikator (Lead QA / Operator)** | Fahmi (Lead QA) | Quality Assurance & Test Engineering | 11/11/2026 | **[✓] VERIFIED** | *(Tanda Tangan Digital)* |

> **Catatan Pengesahan Modul:** Seluruh kasus uji pada modul ini telah berhasil diuji dan dinyatakan **LULUS (100% Pass)** tanpa defek kritis (*zero critical bugs*). Alur bisnis, pembatasan hak akses (RBAC), dan integritas data terverifikasi sesuai standar metodologi AuditSphere.

---

## Rekapitulasi Hasil Pengujian UAT

| Parameter UAT | Nilai Metrik | Keterangan |
| :--- | :---: | :--- |
| **Total Skenario Diuji** | **21 Modul Utama (29 Sub-Modul)** | Mencakup alur otentikasi, manajemen risiko, perencanaan, fieldwork, pelaporan, hingga pemantauan tindak lanjut |
| **Total Kasus Uji (Test Cases)** | **79** | Rincian kasus uji positif, negatif, dan validasi fungsional/keamanan |
| **Tingkat Kelulusan (Pass Rate)** | **100%** | Seluruh kasus uji terverifikasi sukses pada staging environment |
| **Defek Kritis (Critical Bugs)** | **0** | Tidak ada blocker pada alur bisnis utama |
| **Keputusan UAT** | **DITERIMA (PASSED)** | Memenuhi seluruh kriteria penerimaan pengguna dan kontrol ISO 27001 |

---
*Dokumen ini disahkan secara digital oleh Lead QA Engineer, Tim Pengembang, dan Chief Audit Executive (CAE).*
