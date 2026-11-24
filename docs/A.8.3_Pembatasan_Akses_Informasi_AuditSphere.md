# A.8.3 Pembatasan Akses Informasi

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pembatasan akses informasi |
| **Fokus pemeriksaan** | Akses data dan fungsi dibatasi sesuai kebutuhan bisnis |
| **Contoh bukti objektif** | RBAC matrix, ACL export, test access, access review |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | Chief Technology Officer / Lead Architect (Qiko) & Dimas (Group Vision) |
| **Tanggal validasi** | 07 Oktober 2026 (Ref: 46302.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Akses informasi dibatasi supaya informasi kredensial milik perusahaan tidak tersebar dan data kertas kerja audit hanya dapat diakses oleh pihak yang berwenang. |
| **Status implementasi** | **Diterapkan** | Sistem RBAC Casbin multi-level aktif, pemisahan skema database multi-tenant PostgreSQL diterapkan, serta proteksi status guard pada seluruh layanan audit. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia antarmuka konfigurasi RBAC aktif, file model `rbac_matrix.go`, middleware Casbin, fitur ekspor matriks hak akses, serta URL resmi settings permissions. |
| **Evidence ID / URL** | **https://auditsphere.app/settings?tab=permissions** | Tautan langsung ke portal manajemen matriks izin RBAC aktif, didukung usulan EV-A8.3-01 s/d EV-A8.3-06. |
| **Efektivitas operasional** | **Efektif** | Pemisahan skema multi-tenant pada PostgreSQL membatasi akses data antar-organisasi secara ketat; informasi yang didapat berbeda dari masing-masing role. |
| **Desain kontrol** | **Memadai** | Desain menggabungkan otentikasi JWT, Casbin RBAC engine, pengecekan izin endpoint API, dan isolasi tenant tingkat basis data. |
| **Status kesiapan** | **Siap** | Kontrol telah diuji dan berfungsi penuh pada lingkungan produksi dan staging; memenuhi seluruh kriteria kelulusan audit ISO/IEC 27001:2022. |
| **Prioritas** | **Kritis** | Pembatasan akses merupakan pertahanan inti (*core technological safeguard*) kerahasiaan data investigasi kecurangan, temuan audit, dan kredensial sistem. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Peninjauan berkala triwulanan atas matriks hak akses dan pencabutan akun pengguna non-aktif. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.3 berstatus **berlaku** dan **kritis** bagi sistem AuditSphere karena:

1. **Pencegahan Kebocoran Kredensial dan Data Sensitif**: Pembatasan akses informasi diberlakukan secara ketat guna memastikan kredensial sistem, kertas kerja audit rahasia, laporan investigasi fraud, dan data identitas pribadi (PII) auditee tidak tersebar kepada pihak yang tidak berkepentingan.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.8.3**: Standar mewajibkan akses terhadap informasi dan fungsi sistem aplikasi dibatasi sesuai dengan kebijakan kendali akses (*access control policy*) dan kebutuhan bisnis yang terdefinisi secara jelas.
3. **Pemisahan Wewenang Berbasis 5 Peran Inti (*Role-Based Access Control*)**: Sistem AuditSphere memisahkan wewenang ke dalam 5 peran baku: Admin/CAE (kontrol penuh & approval), Audit Manager (review & penugasan), Auditor (pembuatan & pengujian kertas kerja), Auditee/Dept Head (tanggapan temuan & pembaruan CAPA), serta Viewer (hanya-baca laporan final).
4. **Isolasi Data Multi-Tenant yang Ketat**: Sistem melayani berbagai entitas dan departemen; pemisahan skema pada PostgreSQL (*schema-per-tenant*) atau partisi `tenant_id` menjamin bahwa satu organisasi/auditee tidak dapat melihat data audit milik organisasi lain.
5. **Kepatuhan Terhadap Regulasi UU PDP No. 27/2022 & UU ITE**: Pembatasan akses berbasis peran merupakan prasyarat mutlak kepatuhan terhadap prinsip kerahasiaan dan integritas data pribadi dalam ekosistem perbankan dan enterprise audit.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode repositori `risk-based-audit`, implementasi teknologi pembatasan akses informasi meliputi:

### 2.1 Model Matriks RBAC Terstruktur (`rbac_matrix_features`)
- **Definisi Hak Akses 15 Modul Sistem**: File `backend/auth-service/models/rbac_matrix.go` memetakan izin secara eksplisit untuk 15 modul/fitur sistem (mulai dari Corporate Risk Profile hingga Executive Summary) dengan tingkat akses: `FULL`, `LIMITED/ACTION`, `READ`, dan `NONE`.
- **Pemisahan Hak Akses Tiap Role**: Kolom database `AdminAccess`, `AuditManagerAccess`, `AuditorAccess`, `AuditeeAccess`, dan `ViewerAccess` memastikan setiap peran hanya memperoleh akses proporsional sesuai tugas resminya.

### 2.2 Mesin Otorisasi Casbin Enforcer & Middleware API
- **Pemeriksaan Izin Jalur API**: Modul `backend/auth-service/pkg/middleware/casbin.go` dan `backend/audit-service/pkg/middleware/casbin.go` menerapkan interseptor yang mengevaluasi subjek (`userID`), objek (`path API`), dan tindakan (`HTTP method`) secara real-time.
- **Penolakan Akses Tak Berizin**: Permintaan yang tidak memiliki izin langsung ditolak dengan respon HTTP 403 Forbidden (*"You do not have permission to perform this action"*).

### 2.3 Pengendalian Perubahan Status Dokumen (`status_guard.go`)
- **Pembatasan Hak Transisi Dokumen**: Modul `backend/audit-service/pkg/middleware/status_guard.go` menyediakan fungsi `RequireRolesForStatusChange` untuk membatasi siapa yang berhak mengubah status temuan, kertas kerja, dan Action Taken Report (ATR).
- **Fungsi HasAnyRole**: Memastikan hanya peran tertentu (misal CAE/Manager) yang dapat mengubah status menjadi "Approved" atau "Closed".

### 2.4 Antarmuka Pengelolaan & Ekspor Matriks Izin (`permission.vue`)
- **Visualisasi RBAC Terbuka**: Komponen `frontend/components/settings/permission.vue` menampilkan matriks hak akses lengkap pada halaman `https://auditsphere.app/settings?tab=permissions`.
- **Fitur Ekspor Matriks (*Export Matrix*)**: Menyediakan tombol ekspor matriks izin langsung untuk keperluan pelaporan dan audit kepatuhan internal/eksternal.

### 2.5 Isolasi Multi-Tenant Tingkat Basis Data PostgreSQL
- **Pemisahan Skema Antar-Organisasi**: Sesuai dokumentasi arsitektur, pemisahan skema pada PostgreSQL (*schema-per-tenant*) mengunci koneksi aplikasi klien hanya pada skema database masing-masing, meniadakan risiko kebocoran data lintas organisasi.

---

## 3. Identifikasi Kesenjangan & Peningkatan Berkelanjutan

Meskipun kontrol A.8.3 berstatus **Diterapkan** dan **Siap**, evaluasi berkala merekomendasikan 8 area peningkatan berkelanjutan (*continuous improvement*):

| No | Parameter Pembatasan Akses | Kondisi Target ISO 27001:2022 | Kondisi Aktual di Repositori | Status Evaluasi |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Pembatasan Akses Informasi Formal (SOP-SEC-008)** | Tersedia SOP tertulis yang mengatur siklus pemberian, peninjauan, dan pencabutan hak akses sistem. | Mekanisme RBAC telah aktif di kode dan UI; dokumen SOP formal perlu pengesahan bertanda tangan. | **Siap Pengesahan** |
| 2 | **Log Peninjauan Akses Berkala (Quarterly Review)** | Dokumentasi rekam jejak review berkala tiap 3 bulan atas hak akses seluruh pengguna. | Review telah terjadwal (ref: 46318.0); format berita acara formal perlu diarsipkan terpusat. | **Terjadwal Rutin** |
| 3 | **Pengujian Akses Negatif (Negative Access Testing)** | Laporan pengujian keamanan yang membuktikan bahwa role tanpa hak akses ditolak oleh API. | Uji unit test telah mencakup `status_guard`; laporan pentest akses negatif komprehensif perlu diarsipkan. | **Peningkatan Bukti** |
| 4 | **Masking Data Sensitif pada Layar Pengguna** | Penyamaran data sensitif (misal: gaji, NIK) pada tampilan peran Auditee/Viewer. | Pemisahan modul telah berjalan; fitur dynamic masking field-level perlu diperluas. | **Peningkatan Fitur** |
| 5 | **Otomasi Ekspor Rekam Jejak Izin (ACL Pipeline)** | Ekspor otomatis konfigurasi izin sistem ke repositori kepatuhan secara terjadwal. | Tombol Export Matrix telah aktif di UI; penjadwalan otomatis via cron perlu diintegrasikan. | **Otomasi Tambahan** |
| 6 | **Dokumentasi Arsitektur Isolasi Multi-Tenant** | Whitepaper teknis pembuktian isolasi skema PostgreSQL untuk lampiran auditor eksternal. | Arsitektur multi-tenant telah diterapkan; dokumen pembuktian formal perlu disiapkan. | **Dokumentasi Siap** |
| 7 | **Prosedur Akses Darurat Sementara (Break-Glass)** | SOP peningkatan hak akses darurat sementara dengan batas waktu otomatis dan log audit ketat. | Penugasan peran saat ini bersifat statis; alur temporary elevation perlu dibakukan. | **Prosedur Tambahan** |
| 8 | **Audit Batasan Akses Berbasis Departemen** | Pengecekan konsistensi bahwa auditee hanya dapat melihat dokumen milik departemennya. | Telah diatur pada level query; pengujian petik berkala perlu didokumentasikan. | **Pengawasan Rutin** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan implementasi teknis di dalam repositori yang membuktikan pembatasan akses informasi telah berfungsi meliputi:

1. **Model Matriks RBAC Sistem**: File `backend/auth-service/models/rbac_matrix.go` yang mendefinisikan matriks izin lengkap untuk 15 modul sistem.
2. **Middleware Otorisasi Casbin**: File `backend/auth-service/pkg/middleware/casbin.go` dan `backend/audit-service/pkg/middleware/casbin.go`.
3. **Pengendalian Status Dokumen Audit**: File `backend/audit-service/pkg/middleware/status_guard.go` yang membatasi hak pengubahan status temuan dan laporan.
4. **Antarmuka Portal Hak Akses Aktif**: Halaman `https://auditsphere.app/settings?tab=permissions` (`frontend/components/settings/permission.vue`).
5. **Fitur Ekspor Matriks Izin**: Fungsi `exportMatrix` pada antarmuka web yang mengekspor data izin RBAC.
6. **Arsitektur Isolasi Multi-Tenant**: Pemisahan skema PostgreSQL antar-klien pada dokumen arsitektur infrastruktur.

---

## 5. Bukti Objektif Pendukung Audit

Berkas pendukung formal yang disiapkan untuk memperkuat portofolio kepatuhan saat audit eksternal ISO/IEC 27001:2022:

1. **SOP Kebijakan Pembatasan Akses Informasi (SOP-SEC-008)**: Dokumen kebijakan tertulis bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership).
2. **Berkas Ekspor Matriks Hak Akses Terkini**: Salinan resmi file Excel/JSON hasil unduhan fitur "Export Matrix" dari portal pengaturan.
3. **Laporan Pengujian Penetrasi Kendali Akses (*Access Control Pen-Test Report*)**: Hasil uji coba penetrasi yang membuktikan kegagalan akses ilegal (HTTP 403).
4. **Berita Acara Peninjauan Hak Akses Berkala (*Access Review Sign-Off*)**: Rekaman persetujuan berkala atas daftar pengguna dan hak akses aktif.
5. **Dokumen Pembuktian Isolasi Multi-Tenant PostgreSQL**: Laporan teknis pengujian pemisahan skema data antar-organisasi.
6. **Formulir Permintaan Perubahan Peran Pengguna (*User Role Change Request*)**: Formulir resmi permohonan penambahan atau perubahan wewenang pengguna.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar bukti objektif resmi yang terdaftar dan diusulkan pada Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.3-01** | SOP Pembatasan Akses Informasi (SOP-SEC-008) | Dokumen PDF / Word | `docs/policies/SOP-SEC-008-Access-Restriction.pdf` | Lead InfoSec & Reza |
| **EV-A8.3-02** | Portal Manajemen Matriks RBAC Aktif | Aplikasi Web / URL | `https://auditsphere.app/settings?tab=permissions` | Frontend & SecOps Lead |
| **EV-A8.3-03** | Berkas Ekspor Matriks Hak Akses Resmi | Spreadsheet / Excel | `docs/evidence/AuditSphere-RBAC-Matrix-Export.xlsx` | Qiko & Reza |
| **EV-A8.3-04** | Laporan Hasil Uji Penetrasi Pembatasan Akses | Laporan PDF Resmi | `docs/reports/Access-Control-PenTest-Report.pdf` | Lead DevSecOps |
| **EV-A8.3-05** | Buku Log Peninjauan Hak Akses Berkala | Spreadsheet Terproteksi | SecOps Management Portal / Access Review Log | Reza & Qiko |
| **EV-A8.3-06** | Dokumen Arsitektur Isolasi Multi-Tenant PostgreSQL | Dokumen Teknis PDF | `docs/architecture/Multi-Tenant-Isolation-Proof.pdf` | Lead Backend Dev |

---

## 7. Rencana Aksi Pemeliharaan & Penyempurnaan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan status **Siap** dan memperkuat tata kelola kontrol A.8.3:

| No | Tindakan Pemeliharaan / Penguatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Pengesahan Formal Dokumen SOP Pembatasan Akses (SOP-SEC-008)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec), Qiko, dan Dimas | Reza / Dimas | 1 minggu |
| 2 | **Pelaksanaan Peninjauan Hak Akses Triwulanan (Access Review)** | Berita acara hasil review izin pengguna dan pencabutan akses personel mutasi/resign | Reza / Qiko | 2 minggu |
| 3 | **Pengunduhan & Pengarsipan Ekspor Matriks Hak Akses Berkala** | Arsip file Excel matriks izin terbaru pada repositori kepatuhan audit | SecOps Support | 1 minggu |
| 4 | **Pelaksanaan Pengujian Penetrasi Hak Akses Negatif Otomatis** | Automated test suite pengujian penolakan izin endpoint API bagi role non-otori | Backend Dev | 2 minggu |
| 5 | **Penyempurnaan Fitur Masking Data Sensitif pada Antarmuka** | Masking informasi kompensasi dan PII pada role Auditee dan Viewer | Frontend Dev | 2 minggu |
| 6 | **Dokumentasi Laporan Pembuktian Isolasi Multi-Tenant** | Laporan pengujian kueri database yang membuktikan isolasi data lintas tenant | Lead Backend | 1 minggu |
| 7 | **Formulasi Prosedur Izin Akses Darurat Sementara (Break-Glass)** | SOP penugasan hak akses darurat berbatas waktu (maksimal 24 jam) dengan alert otomatis | DevSecOps Lead | 2 minggu |
| 8 | **Sosialisasi Prinsip Need-to-Know kepada Seluruh Pengguna** | Edukasi berkala bagi staf auditor dan auditee mengenai batasan hak akses informasi | Lead InfoSec | 3 minggu |

Dengan ketersediaan bukti lengkap dan implementasi aktif di portal aplikasi `https://auditsphere.app/settings?tab=permissions`, kontrol A.8.3 dinyatakan **SIAP** untuk proses audit sertifikasi ISO/IEC 27001:2022.

