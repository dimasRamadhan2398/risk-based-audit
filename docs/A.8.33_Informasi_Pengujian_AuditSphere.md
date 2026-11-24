# A.8.33 Informasi Pengujian

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Informasi pengujian |
| **Fokus pemeriksaan** | Data uji dipilih, dilindungi, dan dihapus sesuai kebutuhan |
| **Contoh bukti objektif** | Test data policy, masking, access log, deletion record |
| **Sistem** | AuditSphere (auditsphere.app), Data Hub CBS Simulator, Tenant Database (`rb_audit_audit_accenture`), PostgreSQL, Seeders |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 12 Oktober 2026 (Ref: 46307.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Mencegah Kebocoran Data Sensitif di Lingkungan Uji, Kewajiban Anonimisasi / Masking / Sintetis, Pembersihan Data (Data Sanitization/Purging). |
| **Status implementasi** | **Diterapkan** | Pengujian sistem menggunakan data sintetis dummy kemudian akan diuji menggunakan data riil klien di staging dengan proteksi ketat. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia skrip simulator perbankan `init-cbs-simulator.sql`, isolasi basis data tenant `rb_audit_audit_accenture`, modul seeders backend, dan panduan data demo. |
| **Evidence ID / URL** | `init-cbs-simulator.sql, rb_audit_audit_accenture` | Skrip simulator Core Banking System (`data-server/init-cbs-simulator.sql`) dan basis data tenant uji, didukung registrasi bukti EV-A8.33-01 s/d EV-A8.33-06. |
| **Efektivitas operasional** | **Efektif** | Pengujian sistem menggunakan data sintetis dummy berjalan efektif; integritas data nasabah perbankan dan temuan audit terlindungi dari kebocoran. |
| **Desain kontrol** | **Memadai** | Desain berbasis pembangkitan data sintetis murni (~10K transaksi mock), isolasi silo database tenant, dan opsi pemisahan data bersih (`--empty-data` vs `--with-demo-data`). |
| **Status kesiapan** | **Siap** | Mekanisme pengendalian dan proteksi data uji telah beroperasi stabil pada seluruh komponen pengujian dan siap diaudit pada sertifikasi ISO/IEC 27001:2022. |
| **Prioritas** | **Sedang** | Pertahankan pengujian ini dan lakukan pengujian secara berkala dengan tetap memastikan kepatuhan anonimisasi data. |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Fokus peningkatan: formalisasi SOP Informasi Pengujian (SOP-SEC-014), jadwal purging data uji, dan protokol masking staging. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Evaluasi berkala atas volume data uji coba dan verifikasi ketiadaan PII produksi di lingkungan staging. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.33 (*Informasi Pengujian*) **berlaku** penuh untuk seluruh aktivitas pengujian perangkat lunak, verifikasi analitik AI, dan validasi data pada platform AuditSphere berdasarkan pertimbangan strategis:

1. **Mencegah Kebocoran Data Sensitif di Lingkungan Uji**: Lingkungan pengembangan dan pengujian umumnya memiliki kontrol akses dan visibilitas yang lebih terbuka dibandingkan produksi. Penggunaan data nyata klien tanpa perlindungan menimbulkan risiko tinggi kebocoran informasi kertas kerja, rahasia perbankan, dan data operasional auditee.
2. **Kewajiban Hukum Kepatuhan Perlindungan Data Pribadi (UU PDP No. 27/2022)**: Regulasi secara eksplisit melarang pemrosesan data pribadi (seperti NIK/KTP nasabah, nomor rekening, saldo, dan riwayat transaksi) di lingkungan pengujian tanpa dasar hukum atau tanpa melalui proses anonimisasi/pseudonimisasi yang sah.
3. **Mandat Standar ISO/IEC 27001:2022 Klausul A.8.33**: Standar mewajibkan informasi pengujian dipilih dengan cermat, dilindungi dari pengungkapan yang tidak sah, dan dilakukan pembersihan (*data sanitization/purging*) setelah kegiatan pengujian selesai.
4. **Pemberian Validitas Hasil Uji CAATT & Machine Learning**: Pengujian teknik audit berbantuan komputer (CAATT) dan model deteksi anomali (`python-ai`) membutuhkan data dengan karakteristik riil (volume ~10.000 transaksi) beserta pola anomali injeksi tanpa harus mengekspos data nasabah perbankan yang sebenarnya.
5. **Pencegahan Kompromi Data Finansial Klien**: Memastikan seluruh demonstrasi fitur, pelatihan auditor internal, dan uji fungsional menggunakan data sintetis dummy yang terisolasi dalam silo tenant uji (seperti instance `rb_audit_audit_accenture`).

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode dan basis data repositori `risk-based-audit`, implementasi pengelolaan data pengujian mencakup:

### 2.1 Simulator Core Banking System (CBS) Berbasis Data Sintetis (`data-server/init-cbs-simulator.sql`)
- **Dataset Uji Tiruan Skala Penuh (~10.000 Transaksi)**: File `data-server/init-cbs-simulator.sql` menyediakan skema `cbs` terisolasi lengkap dengan 200 nasabah sintetis, 500 rekening mock, dan sekitar 10.000 transaksi perbankan dummy (ATM, Mobile Banking, Teller, Kliring, RTGS, BI-FAST).
- **Penyusunan Identitas Palsu Sesuai Regulasi**: Seluruh data nasabah menggunakan nama rekaan, NIK acak dummy, nomor telepon fiktif, dan alamat sintetis, memastikan *zero exposure* terhadap data pribadi nyata.
- **Injeksi Anomali Buatan untuk Uji Audit**: Simulator menyertakan transaksi anomali terkontrol (seperti transaksi di luar jam kerja, penarikan tunai beruntun di bawah batas pelaporan PPATK/*structuring*, dan rekening dorman aktif tiba-tiba) guna memvalidasi kemampuan deteksi CAATT AuditSphere.

### 2.2 Isolasi Basis Data Uji Tenant (`rb_audit_audit_accenture`)
- **Pemisahan Silo Basis Data Per Tenant**: Arsitektur AuditSphere mengalokasikan database terpisah untuk setiap tenant (contoh: `rb_audit_audit_accenture`, `rb_audit_auth_accenture`, `rb_audit_master_accenture`, `rb_audit_risk_accenture`).
- **Pengendalian Data Staging**: Pengujian fitur khusus klien dijalankan di dalam silo basis data terisolasi tanpa interkoneksi ke basis data klien lain atau basis data produksi utama.

### 2.3 Modul Seeder Aplikasi Microservices Terstandarisasi (`backend/*/cmd/seed.go`)
- **Pembangkitan Data Referensi Sintetis**: Setiap microservice memiliki sub-command `seed` (`auth seed`, `audit seed`, `master seed`, `risk seed`) yang mengisi peran awal, permission matrix, profil risiko standar, dan kertas kerja audit contoh.
- **Dataset Demo Terkendali (`seed_strategic_plans_100.sql`)**: Pustaka seed SQL khusus yang mengisi data perencanaan audit strategis untuk kebutuhan validasi performa dan pagination tanpa mengorbankan data rahasia.

### 2.4 Mekanisme Segregasi Status Data (`scripts/onboard-tenant.sh`)
- **Parameter `--empty-data` (Default Produksi)**: Provisi tenant produksi hanya menjalankan migrasi skema struktur database dengan 0 catatan bisnis riil, mencegah pencemaran data uji ke lingkungan produksi.
- **Parameter `--with-demo-data` (Khusus Staging/Demo)**: Mengaktifkan seeder data dummy untuk kebutuhan pengujian dan demonstrasi fungsional.

### 2.5 Tata Kelola Data Demo (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md`)
- **Kebijakan Tertulis Data Demo vs Produksi**: Dokumen panduan demo secara eksplisit menetapkan bahwa lingkungan demo hanya memuat data sampel fiktif, terpisah dari data bisnis nyata, dan dapat di-reset/wiped setiap saat tanpa dampak operasional.

---

## 3. Evaluasi Kinerja & Area Peningkatan Berkelanjutan

Sesuai catatan checklist Excel baris 98 (*"Pertahankan pengujian ini dan lakukan pengujian secara berkala"*), status kesiapan kontrol adalah **SIAP** dan efektivitas **EFEKTIF**. Untuk mempertahankan kepatuhan regulasi dan meningkatkan maturitas perlindungan data uji, diidentifikasi 8 area peningkatan berkelanjutan:

| No | Parameter Pengelolaan Informasi Pengujian | Kondisi Implementasi Saat Ini | Target Peningkatan Berkelanjutan | Prioritas Peningkatan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Perlindungan Informasi Pengujian (SOP-SEC-014)** | Penggunaan data sintetis telah berjalan operasional; dokumen SOP formal tata kelola data uji belum disahkan. | Penyusunan dan pengesahan SOP-SEC-014 yang mengatur perolehan, proteksi, dan pemusnahan data uji coba. | **Tinggi (Formalitas)** |
| 2 | **Protokol Masking Data Riil Klien di Staging** | Catatan Excel menyebutkan staging dapat diuji data riil klien; belum ada protokol masking tertulis sebelum data masuk. | Penyusunan Pedoman Masking & Pseudonimisasi Data Klien (Hashing NIK, Pengacakan Nama, Masking Nomor Rekening). | **Tinggi (Kepatuhan UU PDP)** |
| 3 | **Otomasi Skrip Pembersihan Data Uji (Purging Script)** | Pembersihan data uji coba di staging masih dilakukan secara manual melalui `docker compose down -v`. | Penyusunan skrip otomatis terjadwal untuk menghapus data transaksi pengujian yang telah selesai masa validasinya. | **Sedang** |
| 4 | **Perekaman Log Akses Khusus ke Database Uji** | Akses ke database uji dicatat melalui log docker bawaan; belum ada rekam jejak terpisah siapa saja yang mengunduh data uji. | Pemberlakuan logging akses khusus untuk memantau personel yang mengekstraksi atau memodifikasi dataset pengujian. | **Sedang** |
| 5 | **Formulir Persetujuan Penggunaan Data Riil (Approval Form)** | Izin pengujian data staging disepakati secara lisan/rapat; belum ada formulir persetujuan bertanda tangan formal. | Pemberlakuan Formulir Otorisasi Khusus Penggunaan Data Produksi di Staging dengan tanda tangan DPO dan Client Sponsor. | **Sedang** |
| 6 | **Enkripsi Dataset Uji pada Media Penyimpanan (At-Rest)** | Dataset uji tersimpan pada volume Docker lokal; belum diproteksi enkripsi kolom khusus untuk field sensitif mock. | Penerapan enkripsi filesystem pada direktori volume data-server dan penegakan izin baca terbatas (`chmod 600`). | **Sedang** |
| 7 | **Verifikasi Integritas Anomali Injeksi pada Seeder** | Anomali diinjeksikan secara statis pada `init-cbs-simulator.sql`; belum ada variasi skenario fraud dinamis. | Pengembangan generator skenario anomali dinamis berbasis konfigurasi parameter untuk memperkaya variasi pengujian. | **Rendah** |
| 8 | **Audit Kuartalan Masa Simpan Data Uji Staging** | Review kapasitas data staging dilakukan insidental saat penyimpanan server mendekati penuh. | Pemberlakuan audit kuartalan berkala untuk memverifikasi bahwa tidak ada data uji kedaluwarsa yang tertinggal di staging. | **Rendah** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak teknis dan konfigurasi nyata pada repositori `risk-based-audit` yang membuktikan kepatuhan kontrol A.8.33:

1. **Skrip Simulator Core Banking System (`data-server/init-cbs-simulator.sql`)**: Pembangkit 10.000 transaksi sintetis, nasabah tiruan, dan rekening mock yang menjamin *zero-exposure* terhadap data nasabah riil.
2. **Isolasi Basis Data Tenant Uji (`rb_audit_audit_accenture`)**: Bukti struktur basis data terpisah per tenant yang membatasi pengujian hanya di dalam silo mandiri tanpa risiko kontaminasi lintas tenant.
3. **Modul Seeder Aplikasi Microservices (`backend/*/cmd/seed.go`)**: Pembangkit data referensi awal (peran, izin, profil risiko, kertas kerja) dengan data standar tanpa kredensial rahasia.
4. **Skrip Otomasi Onboarding Multi-Tenant (`scripts/onboard-tenant.sh`)**: Opsi tegas pemisahan skema kosong (`--empty-data`) untuk rilis operasional dan skema demo (`--with-demo-data`) untuk pengujian.
5. **Dokumen Panduan Lingkungan Demo (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md`)**: Pedoman resmi yang melarang penggunaan data produksi nyata dan mewajibkan isolasi total lingkungan demo.
6. **Panduan Arsitektur & Provisi Domain Klien (`docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md`)**: Spesifikasi arsitektur silo data yang menjamin data uji tidak dapat diakses dari jaringan publik.

---

## 5. Bukti Objektif Penguat Kesiapan Audit

Dokumen prosedural dan rekaman operasional pelengkap yang disiapkan untuk mengukuhkan status "Siap" dan memenuhi kriteria kelulusan audit ISO 27001:

1. **Dokumen SOP Perlindungan & Pemusnahan Informasi Pengujian (SOP-SEC-014)**: Standar operasional resmi bertanda tangan Dimas (Leadership) dan Reza (Lead InfoSec).
2. **Pedoman Teknis Masking & Anonimisasi Data Produksi ke Lingkungan Staging**: Panduan langkah-demi-langkah penghilangan data sensitif (PII) bila pengujian memerlukan replika struktur produksi.
3. **Formulir Permohonan & Otorisasi Penggunaan Informasi Pengujian (*Test Data Approval Form*)**: Rekam jejak persetujuan formal sebelum data riil disanitasi dan dimasukkan ke staging.
4. **Berita Acara Pembersihan & Pemusnahan Data Uji (*Test Data Purging / Deletion Logbook*)**: Catatan penghapusan dataset pengujian pasca penyelesaian siklus UAT atau demonstrasi.
5. **Laporan Hasil Verifikasi Masking Data Uji**: Bukti audit teknis yang memverifikasi bahwa seluruh data di staging telah tersanitasi secara sempurna.
6. **Matriks Otorisasi Akses ke Repositori & Simulator Data Uji**: Daftar personel yang memiliki wewenang mengelola berkas SQL seeder dan simulator perbankan.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber Repositori | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.33-01** | Skrip Simulator CBS & Data Uji Sintetis | File SQL Skrip | `data-server/init-cbs-simulator.sql` (~10K transaksi mock) | Lead Data Engineer |
| **EV-A8.33-02** | Skema Basis Data Uji Tenant Terisolasi | Database Schema & Skrip | PostgreSQL Instance (`rb_audit_audit_accenture`) & `scripts/onboard-tenant.sh` | DevSecOps Engineer |
| **EV-A8.33-03** | Modul Seeders Data Uji Microservices | Source Code Go & SQL | `backend/*/cmd/seed.go` & `backend/scripts/seed_strategic_plans_100.sql` | Lead Backend Dev |
| **EV-A8.33-04** | SOP Perlindungan Informasi Pengujian (SOP-SEC-014) | Dokumen PDF / Word Resmi | `docs/policies/SOP-SEC-014-Test-Data-Protection-and-Sanitization.pdf` | Lead InfoSec & Dimas |
| **EV-A8.33-05** | Pedoman Masking & Anonimisasi Data Produksi ke Staging | Dokumen PDF Panduan Teknis | `docs/guidelines/Data-Masking-and-Anonymization-Protocol.pdf` | Data Protection Officer (DPO) |
| **EV-A8.33-06** | Buku Log Pembersihan Data Uji & Berita Acara Purging | Spreadsheet & Formulir PDF | `docs/evidence/Test-Data-Purging-and-Deletion-Log.pdf` | SecOps Response Team |

---

## 7. Rencana Aksi Pemeliharaan & Peningkatan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan efektivitas kontrol informasi pengujian dan merealisasikan peningkatan tata kelola hingga batas target **27 November 2026**:

| No | Tindakan Pemeliharaan / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Informasi Pengujian (`SOP-SEC-014`)** | Dokumen SOP formal mencakup siklus hidup data uji, proteksi, otorisasi, dan pemusnahan yang disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Penyusunan Pedoman Teknis Masking & Pseudonimisasi Data** | Prosedur operasional masking data riil klien sebelum direplikasi ke lingkungan staging sesuai UU PDP | Lead InfoSec & DPO | 1 minggu |
| 3 | **Pengembangan Skrip Otomasi Pembersihan Data Uji (*Data Purging*)** | Skrip terjadwal untuk menghapus data audit pengujian dan transaksi sementara pasca kegiatan UAT/demo | DevSecOps | 2 minggu |
| 4 | **Pemberlakuan Formulir Otorisasi Data Uji Staging Bertanda Tangan** | Formulir persetujuan resmi sebelum data riil disanitasi dan dimasukkan ke staging untuk verifikasi akhir | Reza / Dimas | 1 minggu |
| 5 | **Penerapan Pembatasan Hak Akses Berkas Data Uji (`chmod 600`)** | Pengamanan izin berkas `init-cbs-simulator.sql` dan seeder di server agar hanya dapat dibaca oleh service account | Lead Infra | 1 minggu |
| 6 | **Pengayaan Variasi Kasus Anomali Injeksi pada Simulator CBS** | Penambahan skenario uji penipuan perbankan baru pada simulator untuk meningkatkan akurasi modul analitik AI | Lead Data Eng | 2 minggu |
| 7 | **Pemberlakuan Audit Kuartalan Masa Simpan Data Staging** | Pemeriksaan kuartalan berkala terhadap direktori unggahan staging guna memastikan data uji lama telah dibersihkan | SecOps Team | 1 minggu |
| 8 | **Sosialisasi Kebijakan Zero Production Data in Development** | Edukasi kepada seluruh tim rekayasa perangkat lunak mengenai larangan menyalin data produksi ke mesin lokal | Reza / Dimas | 1 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-014 bertanda tangan, Pedoman Masking Data, Skrip Purging Data Uji, dan Berita Acara Pemusnahan pada Master Audit Evidence Index.

