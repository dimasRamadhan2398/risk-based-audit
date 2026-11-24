# A.5.13 Pelabelan Informasi

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Pelabelan informasi |
| **Fokus pemeriksaan** | Label konsisten dengan skema klasifikasi dan media |
| **Contoh bukti objektif** | Prosedur label, sampel dokumen/email, konfigurasi DLP |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Pelabelan informasi audit internal, KKA, LHA, dan kredensial enterprise wajib diterapkan agar seluruh penangan data memahami tingkat kerahasiaan dan perlakuan pengamanan yang disyaratkan. |
| **Status implementasi** | **Dalam proses** | Ketentuan klasifikasi 4-level dan desain kontrol dynamic watermarking telah dicatat pada ISO-AS-001; generator dokumen dan antarmuka belum menyematkan label klasifikasi dan watermark visual secara otomatis. |
| **Ketersediaan bukti** | **Ada sebagian** | Skema klasifikasi (Bab 4.2), modul engine docxbuilder, dan kontrol otorisasi ekspor RBAC tersedia; SOP pelabelan formal, sampel label header/footer Word, dan konfigurasi DLP belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A5.13-01 s/d EV-A5.13-05 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Pembatasan hak akses unduh (RBAC) pada endpoint laporan berjalan efektif; namun berkas yang diunduh belum memuat label klasifikasi visual permanen atau metadata DLP. |
| **Desain kontrol** | **Perlu perbaikan** | Modul docxbuilder mampu menghasilkan dokumen Word, namun template dokumen belum menyertakan header/footer klasifikasi dan dynamic watermark identitas pengunduh. |
| **Status kesiapan** | **Belum siap** | Prosedur operasional pelabelan formal dan implementasi teknis tagging dokumen belum memenuhi seluruh kriteria kesiapan audit sertifikasi. |
| **Prioritas** | **Tinggi** | Prasyarat kepatuhan UU PDP No. 27/2022, pencegahan kebocoran data kertas kerja audit, dan kepatuhan kelompok kontrol A.5.12 s/d A.5.14. |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah SOP Pelabelan Informasi disahkan dan fitur dynamic watermark diintegrasikan ke lingkungan produksi. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.13 **berlaku** untuk AuditSphere karena:

1. **Sensitivitas Dokumen KKA & LHA**: AuditSphere menghasilkan, menyimpan, dan mengekspor dokumen audit yang memuat temuan berisiko tinggi, evaluasi pengendalian internal korporasi, indikasi kecurangan (*fraud*), dan informasi identitas pribadi (PII) auditee. Tanpa penandaan atau pelabelan klasifikasi yang tegas, penerima dokumen dapat lalai dalam memperlakukan kerahasiaan berkas tersebut.
2. **Keterikatan Framework ISO/IEC 27001:2022**: Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) Bagian 6.6 dan 7 secara spesifik menetapkan titik kontrol keamanan pada ekspor laporan (PDF/Excel/DOCX) yang wajib dilengkapi *dynamic watermark* dan pembatasan izin untuk mencegah kebocoran data.
3. **Mandat Standar Tata Kelola ISO 27001:2022 Klausul A.5.13**: Standar ISO/IEC 27001:2022 kontrol Annex A.5.13 mewajibkan organisasi untuk mengembangkan dan mengimplementasikan serangkaian prosedur pelabelan informasi yang konsisten dengan skema klasifikasi informasi yang telah ditetapkan (*Public, Internal, Confidential, Restricted*).
4. **Penegakan Rantai Kepemilikan (Chain of Custody)**: Penandaan identitas penangan data (*dynamic watermark* identitas pengguna, IP, dan timestamp unduh) merupakan kontrol preventif dan detektif kunci dalam menjaga rantai kepemilikan data dan mempermudah investigasi forensik jika terjadi kebocoran laporan audit.
5. **Kepatuhan Regulasi & Standar Perbankan/Enterprise**: Klien enterprise pada sektor jasa keuangan, perbankan, dan BUMN mewajibkan seluruh dokumen kertas kerja audit dan LHA yang dibagikan secara elektronik diberi penandaan klasifikasi kerahasiaan formal sesuai regulasi OJK dan standar IIA IPPF.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Skema Pelabelan Berdasarkan Klasifikasi Informasi

Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 4.2 menetapkan 4 tingkatan klasifikasi keamanan informasi. Penerapan pelabelan untuk masing-masing tingkatan dirancang sebagai berikut:

| Klasifikasi | Label Visual / Penandaan | Label Metadata / Digital | Contoh Aset / Dokumen | Aturan Penanganan |
| :--- | :--- | :--- | :--- | :--- |
| **PUBLIC** | Tanpa label khusus atau penanda `[PUBLIC]` pada bagian footer dokumen | Metadata: `Classification=Public`; Header: Content-Disposition inline | Halaman login, landing page publik, dokumentasi publik API auditsphere.app | Dapat diakses dan didistribusikan secara bebas tanpa pembatasan otentikasi. |
| **INTERNAL** | Header/Footer teks `[INTERNAL USE ONLY]` warna abu-abu netral (`#555555`) | Metadata: `Classification=Internal`; Tagging repositori git privat | Pedoman operasional audit, SOP teknis, template formulir KKA, arsitektur dasar | Hanya untuk staf internal AIFL & pengguna terdaftar; dilarang dibagikan ke publik. |
| **CONFIDENTIAL** | Header tebal `[CONFIDENTIAL / RAHASIA]` warna biru tua/oranye; Dynamic watermark | Header HTTP: `X-Classification=Confidential`; Tagging tabel DB; PDF/Docx metadata | LHA (Laporan Hasil Audit), Register Risiko ERM, Kertas Kerja F-01-F-05, audit log | Wajib enkripsi at-rest/in-transit; unduhan dibatasi peran Auditor/Manager/CAE. |
| **RESTRICTED** | Header & Footer `[STRICTLY RESTRICTED]` warna merah (`#C0392B`); Watermark diagonal pekat | Enkripsi field-level (AES-256); Data masking PII (NIK/rekening); Tagging DLP | Kredensial/Secrets, private keys, bukti investigasi fraud, data gaji, PII auditee | Akses sangat terbatas (*Need-to-Know basis*), wajib otentikasi MFA, dilarang ekspor massal. |

### 2.2 Mekanisme Ekspor Dokumen & Penegakan Otorisasi (RBAC)

AuditSphere menyediakan layanan ekspor laporan audit formal melalui modul `backend/audit-service/pkg/docxbuilder/docx_generator.go`. Endpoint `GET /audit-result-reports/:id/download-docx` menghasilkan dokumen Microsoft Word (`.docx`) lengkap dengan Nomor LHA, Surat Tugas, Data Fieldwork, dan Kertas Kerja. Akses unduh diproteksi secara bertingkat:

1. **Otorisasi API Gateway**: Kong Gateway memvalidasi JWT bearer token dan claims peran pengguna sebelum meneruskan permintaan ke `audit-service`.
2. **Matriks Hak Akses RBAC**: Berdasarkan `seeders/seeder.go` dan tabel `rbac_matrix_features`, hanya peran yang berwenang (`ADMIN`, `AUDITOR`, `EXECUTIVE` / CAE) yang memiliki izin unduh laporan final.
3. **Penamaan Berkas Terstruktur**: Berkas unduhan otomatis diberi nama terstruktur `LHA_{nomor_laporan}.docx` untuk memudahkan identifikasi dokumen oleh auditor.

### 2.3 Desain Kontrol Dynamic Watermarking & Data Masking

Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 6.6 Bagian 3 dan 4 menetapkan kontrol pencegahan kebocoran data pada titik keluaran aplikasi:

1. **Pencegahan Kebocoran Data (Data Masking)**: Data masking diterapkan secara otomatis pada informasi sensitif (PII auditee, nomor identitas/NIK, nomor rekening bank) sesuai izin peran pengguna saat data ditampilkan di antarmuka web.
2. **Dynamic Watermark**: Kontrol ekspor laporan audit (PDF/Excel) dirancang untuk memuat watermark dinamis transparan berisi identitas pengguna (nama & employee ID), alamat IP pemohon, dan timestamp pengunduhan.

### 2.4 Pelabelan Pesan Log & Streaming Audit Trail

Pada lapisan backend dan pesan asinkronus (Apache Kafka & database `sys_alert_logs`), setiap kejadian audit dan log sistem diberi label klasifikasi terstruktur dalam format JSON mencakup parameter severity (`INFO`, `WARN`, `ERROR`, `CRITICAL`), tipe agreement (`SYSTEM` / `ROLE` / `PROJECT`), serta konteks identitas pengguna untuk memastikan keterlacakan penuh.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :---: | :--- |
| **Skema Klasifikasi & Ketentuan Pelabelan Awal** | **Tersedia** | Tercantum dalam Bab 4.2 & Bab 6.6 `docs/DOKUMEN_KESELARASAN_ISO27001.md`. |
| **Modul Generator Dokumen Audit LHA (docxbuilder)** | **Tersedia** | Engine `backend/audit-service/pkg/docxbuilder/docx_generator.go` mampu menghasilkan berkas `.docx` terstruktur. |
| **Kontrol Otorisasi Pengunduhan Berbasis RBAC** | **Tersedia** | Hak akses unduh dokumen dibatasi melalui claims peran JWT dan otorisasi Casbin. |
| **Standar Operasional Prosedur (SOP) Pelabelan Informasi Formal** | **Belum tersedia** | Belum ada dokumen SOP formal bertanda tangan manajemen mengenai aturan pelabelan informasi lintas media. |
| **Fitur Dynamic Watermark pada Hasil Ekspor** | **Belum tersedia** | Desain watermark (User ID & Timestamp) tercantum di arsitektur namun belum terintegrasi pada generator dokumen produksi. |
| **Header & Footer Klasifikasi pada Dokumen Otomatis** | **Belum tersedia** | Template LHA yang digenerate docxbuilder belum menyematkan label `[CONFIDENTIAL]` pada header/footer Word. |
| **Pelabelan Metadata & Konfigurasi DLP** | **Belum tersedia** | Belum ada konfigurasi sistem pencegah kebocoran data (DLP) atau injeksi metadata klasifikasi pada berkas digital. |
| **Bukti Sosialisasi & Pedoman Pelabelan kepada Pengguna** | **Belum tersedia** | Belum ada dokumentasi pelatihan atau petunjuk teknis pelabelan kepada staf dan auditor. |

---

## 4. Catatan / Gap Terverifikasi

1. **Template generator docxbuilder belum menyematkan label klasifikasi formal pada header/footer dokumen**: Modul `docx_generator.go` saat ini telah menghasilkan sampul dan bab isi LHA, namun belum menambahkan penanda teks `[CONFIDENTIAL / RAHASIA]` pada header atau footer setiap halaman dokumen Word yang diekspor.
2. **Fitur dynamic watermark pada ekspor laporan belum terintegrasi ke kode produksi**: Prinsip dynamic watermark (User ID, IP, dan timestamp) telah ditegaskan pada dokumen arsitektur dan keselarasan ISO, tetapi implementasi kode rendering watermark transparan pada file DOCX dan PDF unduhan belum aktif.
3. **Format penamaan berkas unduhan belum mencantumkan prefiks klasifikasi standar**: Penamaan file saat ini masih menggunakan format umum `LHA_{nomor}.docx`, belum menerapkan konvensi penamaan berlabel seperti `[CONFIDENTIAL]_LHA_{nomor}.docx` yang memudahkan penyaringan oleh sistem DLP pengguna.
4. **Metadata dokumen digital belum memuat tag klasifikasi SMKI**: Properti dokumen Word (`docProps/core.xml` atau Dublin Core tags) belum disuntikkan tag metadata klasifikasi (misal `SecurityClassification = Confidential`), sehingga alat pemantau data eksternal tidak dapat membaca tingkat sensitivitas dokumen secara otomatis.
5. **Belum ada integrasi kebijakan Data Loss Prevention (DLP)**: Belum ada rule filtering atau policy DLP (misal pada email gateway atau reverse proxy) untuk mendeteksi, memperingatkan, atau memblokir pengiriman dokumen berlabel Restricted/Confidential ke luar jaringan resmi.
6. **Belum ada pedoman pelabelan media penyimpanan fisik dan backup**: Penandaan fisik untuk media pencadangan (seperti hard drive eksternal, kartu arsip, atau media penyimpanan offline) belum diatur dalam prosedur operasional tertulis.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :-: | :--- | :--- | :-: | :-: | :--- |
| **1** | Susun dan sahkan Standar Operasional Prosedur (SOP) Pelabelan Informasi mencakup format label visual, metadata digital, dan penanganan media fisik. | Lead Information Security (Reza) | **Tinggi** | `[isi]` | Dokumen SOP Pelabelan Informasi bertanda tangan |
| **2** | Perbarui modul backend docxbuilder agar menyematkan header dan footer label klasifikasi (`[CONFIDENTIAL / RAHASIA]`) secara otomatis pada setiap halaman dokumen LHA. | Tim Backend (audit-service) | **Tinggi** | `[isi]` | Commit kode docxbuilder & sampel dokumen LHA berlabel |
| **3** | Implementasikan dynamic watermark visual transparan berisi User ID, Nama Pengunduh, IP Address, dan Timestamp pada dokumen PDF/DOCX yang diekspor. | Tim Backend & Frontend | **Tinggi** | `[isi]` | Hasil uji coba fitur watermark pada file ekspor |
| **4** | Standarisasi penamaan file unduhan otomatis dengan mencantumkan prefiks klasifikasi (misal: `[CONFIDENTIAL]_LHA_{nomor}.docx`). | Tim Frontend & Backend | **Sedang** | `[isi]` | Kode penamaan unduhan di store/controller |
| **5** | Sematkan properti metadata klasifikasi informasi (DocProps / Dublin Core Metadata) ke dalam file digital yang dihasilkan aplikasi. | Tim Backend (audit-service) | **Sedang** | `[isi]` | Pengujian metadata berkas via script inspeksi |
| **6** | Tambahkan indikator visual label klasifikasi data pada antarmuka web (UI Badge: Confidential, Restricted, Internal) di halaman KKA dan LHA. | Tim Frontend (Nuxt 4) | **Sedang** | `[isi]` | Tangkapan layar antarmuka dengan label klasifikasi |
| **7** | Lakukan sosialisasi aturan pelabelan informasi kepada seluruh auditor dan personil pengelola sistem. | Lead Information Security | **Sedang** | `[isi]` | Daftar hadir dan materi sosialisasi |
| **8** | Daftarkan usulan Evidence ID untuk kontrol A.5.13 ke dalam Master Audit Evidence Index. | Lead Information Security | **Sedang** | `[isi]` | Entri Master Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| **EV-A5.13-01** | Ketentuan Klasifikasi & Watermarking (Bab 4.2 & Bab 6.6) | `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| **EV-A5.13-02** | Kode Sumber Mesin Pembuat Dokumen Audit (docxbuilder) | `backend/audit-service/pkg/docxbuilder/docx_generator.go` | Internal |
| **EV-A5.13-03** | Endpoint & Log Otorisasi Unduh LHA | `backend/audit-service/routes/handler.go` | Internal |
| **EV-A5.13-04** | Standar Operasional Prosedur (SOP) Pelabelan Informasi *(Draf Usulan)* | `docs/security/` *(lokasi usulan)* | Restricted |
| **EV-A5.13-05** | Sampel Dokumen LHA dengan Label Klasifikasi & Dynamic Watermark *(Draf Usulan)* | `docs/security/` *(lokasi usulan)* | Restricted |

> *Setelah tindakan 1, 2, dan 3 selesai dilaksanakan, tambahkan Evidence ID untuk salinan SOP Pelabelan Informasi bertanda tangan, tangkapan layar fitur dynamic watermark, dan sampel dokumen berlabel resmi.*

