# Lampiran A.5.32 — Contoh Template Perjanjian IP dan Klausul Kontrak On-Premises

> **Status: DRAFT contoh.** Wajib ditinjau konsultan hukum sebelum dipakai. Isian `[...]` harus diganti. Bukan nasihat hukum.
> Dasar hukum yang relevan (verifikasi oleh counsel): UU No. 28/2014 tentang Hak Cipta, UU No. 27/2022 tentang Pelindungan Data Pribadi (UU PDP), UU ITE.

---

## Bagian 1 — Perjanjian Pengalihan Hak Kekayaan Intelektual dan Kerahasiaan (Kontributor / Kontraktor)

**Para Pihak**
- **Perusahaan:** [Nama badan hukum pemilik AuditSphere], alamat [...], diwakili oleh [nama, jabatan].
- **Kontributor:** [Nama lengkap / badan usaha], NIK/NPWP [...], alamat [...], status: ☐ karyawan ☐ kontraktor/freelance ☐ magang.

**Pasal 1 — Definisi**
1. *Hasil Kerja*: seluruh kode sumber, skrip, model, konfigurasi, desain, dokumentasi, data, terjemahan, dan materi lain yang dibuat Kontributor untuk atau terkait AuditSphere, selama masa kerja atau penugasan.
2. *IP Pihak Ketiga/Open Source*: komponen yang dimiliki atau dilisensikan pihak lain.
3. *Alat AI*: layanan generatif (mis. Claude Code, Jules, Copilot, ChatGPT) yang digunakan untuk menghasilkan atau membantu Hasil Kerja.

**Pasal 2 — Kepemilikan dan Pengalihan**
1. Seluruh hak cipta dan hak ekonomi atas Hasil Kerja menjadi milik Perusahaan sejak tercipta. Sepanjang belum otomatis beralih menurut hukum, Kontributor dengan ini mengalihkan seluruh hak tersebut kepada Perusahaan secara penuh, di seluruh dunia, dan tanpa batas waktu.
2. Kontributor menjamin Hasil Kerja adalah karya orisinalnya dan tidak melanggar hak pihak lain.
3. Hak moral diakui sesuai hukum; Kontributor menyetujui Perusahaan mengubah, mengadaptasi, dan tidak mencantumkan nama Kontributor.
4. Kontributor menandatangani dokumen tambahan yang wajar untuk menyempurnakan pengalihan, termasuk pencatatan hak cipta.

**Pasal 3 — Kekayaan Intelektual Bawaan (Pre-existing IP)**
Kontributor wajib mendaftar IP bawaan yang akan dipakai dalam Hasil Kerja pada Lampiran 1. IP bawaan tetap milik Kontributor; Perusahaan memperoleh lisensi non-eksklusif, bebas royalti, tidak dapat dicabut, dan dapat disublisensikan untuk menggunakannya sebagai bagian dari AuditSphere.

**Pasal 4 — Open Source dan Komponen Pihak Ketiga**
1. Kontributor dilarang memasukkan komponen berlisensi copyleft kuat (GPL, AGPL, SSPL, RSAL) atau lisensi *source-available* tanpa persetujuan tertulis Penanggung Jawab Compliance.
2. Setiap dependensi baru wajib dicatat (nama, versi, lisensi) pada License Register sebelum merge ke cabang `staging`.
3. Dilarang menyalin kode dari sumber tanpa lisensi jelas (mis. Stack Overflow tanpa atribusi, repositori tanpa lisensi, proyek klien/pemberi kerja lain).

**Pasal 5 — Penggunaan Alat AI**
1. Alat AI hanya boleh digunakan yang terdaftar/disetujui Perusahaan, dengan akun dan pengaturan yang menonaktifkan pelatihan model atas input (*no-training/zero-retention* bila tersedia).
2. Dilarang memasukkan ke Alat AI: kredensial/secret, data pribadi, data tenant/klien, dokumen audit nyata, atau kode pihak ketiga yang dilindungi.
3. Keluaran Alat AI yang masuk ke repositori wajib ditinjau manusia (code review) dan Kontributor bertanggung jawab atas kepatuhan lisensi dan kualitasnya. Kontributor wajib mengungkapkan penggunaan Alat AI yang signifikan pada deskripsi commit/PR bila diminta.
4. Sebagai Hasil Kerja, keluaran Alat AI yang dimasukkan Kontributor diperlakukan sebagai milik Perusahaan sejauh diizinkan hukum dan syarat layanan Alat AI.
5. Akun bot/otomasi (mis. Jules, Claude Code, CI bot) beroperasi atas nama dan di bawah tanggung jawab pemegang akun/penugasnya; keluaran bot tunduk pada Pasal ini.

**Pasal 6 — Kerahasiaan**
Kontributor menjaga kerahasiaan kode sumber, arsitektur, kredensial, data tenant/klien, dokumen audit, dan informasi non-publik selama penugasan dan [3 / 5] tahun sesudahnya (tanpa batas waktu untuk rahasia dagang dan data pribadi).

**Pasal 7 — Pengembalian dan Offboarding**
Saat penugasan berakhir, Kontributor wajib dalam [3] hari kerja: mengembalikan/menghapus seluruh salinan kode dan data, menyerahkan seluruh akses (repositori, server, kredensial), dan menandatangani berita acara. Perusahaan mencabut akses pada hari terakhir.

**Pasal 8 — Pelanggaran dan Ganti Rugi**
Pelanggaran atas Pasal 4–7 dapat mengakibatkan pemutusan, tuntutan ganti rugi, dan upaya hukum lain.

**Pasal 9 — Hukum yang Berlaku**
Hukum Republik Indonesia. Sengketa diselesaikan musyawarah, lalu [BANI / Pengadilan Negeri ...].

| Perusahaan | Kontributor |
|---|---|
| Nama: | Nama: |
| Jabatan: | Tanggal: |
| Tanda tangan / meterai: | Tanda tangan / meterai: |

**Lampiran 1 — Daftar IP Bawaan Kontributor**

| No | Deskripsi | Lisensi | Digunakan di |
|---|---|---|---|
| 1 | [...] | [...] | [...] |

**Catatan penerapan internal:** tandatangani sebelum akses repositori diberikan; simpan salinan di folder HR/Legal; catat kontributor lama (retroaktif) dengan *confirmatory assignment* yang berlaku surut ke tanggal mulai kerja.

---

## Bagian 2 — Klausul Kontrak Klien untuk Penyebaran On-Premises

> Disisipkan pada Perjanjian Lisensi dan Layanan AuditSphere (On-Premises). Istilah: **Pemberi Lisensi** = Perusahaan; **Klien** = pelanggan.

**Pasal A — Kepemilikan Perangkat Lunak**
1. AuditSphere (kode, binari, image container, model, dokumentasi, antarmuka, dan pembaruannya) adalah milik Pemberi Lisensi atau pemberi lisensinya. Kontrak ini hanya memberikan lisensi, bukan penjualan.
2. Klien memperoleh lisensi non-eksklusif, tidak dapat dialihkan, tidak dapat disublisensikan, terbatas pada [jumlah] instalasi/tenant/pengguna, untuk keperluan internal Klien, selama masa kontrak.
3. Klien dilarang: menyalin di luar yang diizinkan, merekayasa balik, mendekompilasi, menghapus pemberitahuan hak cipta, menyewakan atau menyediakan ke pihak ketiga, atau menggunakan untuk membangun produk pesaing.

**Pasal B — Kepemilikan Data Klien**
1. Seluruh data yang dimasukkan, diunggah, atau dihasilkan Klien dalam sistem (data audit, kertas kerja, temuan, laporan, data master, lampiran, data pegawai, dan *output* analitik atas data Klien) (**"Data Klien"**) adalah milik Klien sepenuhnya.
2. Pemberi Lisensi tidak memperoleh hak atas Data Klien kecuali yang tegas diberikan dalam Pasal ini.
3. Data Klien disimpan pada infrastruktur yang dikendalikan Klien. Pemberi Lisensi **tidak** memiliki akses ke Data Klien kecuali Klien memberikan akses secara tertulis dan terbatas (mis. sesi dukungan), dan akses itu dicatat serta dapat dicabut.
4. Klien bertindak sebagai **Pengendali Data Pribadi**; Pemberi Lisensi hanya sebagai **Prosesor** bila ia mengakses data pribadi dalam rangka dukungan, berdasarkan instruksi Klien dan UU PDP.

**Pasal C — Data Turunan, Telemetri, dan Peningkatan Produk**
1. Pemberi Lisensi tidak mengumpulkan Data Klien. Telemetri/log teknis yang bukan Data Klien (versi, status layanan, kode galat tanpa isi data) boleh dikumpulkan bila Klien menyetujuinya [☐ ya ☐ tidak].
2. Data Klien tidak digunakan untuk melatih, menyempurnakan, atau mengevaluasi model AI Pemberi Lisensi atau pihak ketiga, kecuali dengan persetujuan tertulis terpisah.

**Pasal D — Konten Pihak Ketiga dan Open Source**
Komponen open source dalam AuditSphere tunduk pada lisensinya masing-masing. Daftar komponen dan lisensi (SBOM/`NOTICE`) disediakan kepada Klien atas permintaan. Bila Klien menerima paket yang memuat komponen berlisensi copyleft, Pemberi Lisensi menyediakan kewajiban pemenuhannya (mis. penawaran kode sumber) sebagaimana disyaratkan lisensi tersebut.

**Pasal E — Backup, Keamanan, dan Tanggung Jawab**
Klien bertanggung jawab atas pencadangan, keamanan fisik/jaringan, dan kontrol akses lingkungan on-premises. Pemberi Lisensi bertanggung jawab atas kerentanan dalam perangkat lunak yang diserahkan sesuai SLA [...].

**Pasal F — Pengembalian dan Penghapusan Saat Kontrak Berakhir**
1. Data Klien tetap pada Klien; Klien dapat mengekspor data dalam format [CSV/JSON/dump PostgreSQL] kapan pun.
2. Saat kontrak berakhir, hak lisensi berhenti; Klien menghapus perangkat lunak dan menyatakannya tertulis dalam [30] hari. Pemberi Lisensi menghapus setiap salinan Data Klien yang ia pegang dan mengonfirmasinya tertulis.

**Pasal G — Eskro Kode Sumber (opsional)**
[☐ Tidak ada ☐ Kode sumber dititipkan pada agen eskro dan dilepas bila Pemberi Lisensi pailit atau menghentikan dukungan.]

**Pasal H — Jaminan dan Ganti Rugi IP**
Pemberi Lisensi menjamin sepengetahuannya AuditSphere tidak melanggar hak cipta pihak ketiga dan akan membela Klien dari klaim pelanggaran IP atas perangkat lunak yang diserahkan, dengan batasan: tidak berlaku bila pelanggaran timbul dari modifikasi oleh Klien atau kombinasi dengan produk lain. Total tanggung jawab dibatasi sebesar [...].

---

## Bagian 3 — Daftar Periksa Penerapan

| No | Tindakan | PIC | Target | Status |
|---|---|---|---|---|
| 1 | Konsultan hukum meninjau Bagian 1 dan 2 | Legal | [...] | ☐ |
| 2 | Tandatangani Bagian 1 oleh seluruh kontributor aktif (termasuk confirmatory assignment) | HR/Legal | [...] | ☐ |
| 3 | Jadikan Bagian 1 syarat onboarding sebelum akses repo | HR/PM | [...] | ☐ |
| 4 | Daftar Alat AI yang disetujui dan setelan no-training | PM/Security | [...] | ☐ |
| 5 | Masukkan Bagian 2 ke template kontrak on-premises | Legal/Sales | [...] | ☐ |
| 6 | Siapkan SBOM dan `NOTICE` untuk lampiran Pasal D | Engineering | [...] | ☐ |
