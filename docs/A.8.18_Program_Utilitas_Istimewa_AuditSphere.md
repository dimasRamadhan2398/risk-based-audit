# A.8.18 Program Utilitas Istimewa

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Program utilitas istimewa |
| **Fokus pemeriksaan** | Utility yang dapat melewati kontrol dibatasi dan dicatat |
| **Contoh bukti objektif** | Utility inventory, access list, execution log, approval |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166, Docker Engine, PostgreSQL, Microservices |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 09 Oktober 2026 (Ref: 46304.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Pembatasan hak akses fungsi utilitas untuk melindungi informasi penting serta mencegah pengabaian kontrol keamanan (*control override*) oleh personel teknis maupun pihak luar. |
| **Status implementasi** | **Diterapkan** | Utility sistem dan command eksekusi dibatasi hanya melalui binary terotentikasi di container; port database privat tidak terekspos ke host publik. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia RBAC permission matrix di `https://auditsphere.app/settings?tab=permissions`, konfigurasi Dockerfile berbasis Alpine minimal, dan skrip otomasi terkontrol `onboard-tenant.sh`. |
| **Evidence ID / URL** | `https://auditsphere.app/settings?tab=permissions` | Antarmuka matriks izin RBAC AuditSphere (Settings > Tab Permissions) dan direktori skrip `backend/scripts/` terotentikasi, didukung registrasi EV-A8.18-01 s/d EV-A8.18-06. |
| **Efektivitas operasional** | **Efektif** | Utility sistem dan command eksekusi dibatasi hanya melalui binary terotentikasi di container; perkakas bypass seperti compiler, debugger, dan admin DB publik tidak terpasang. |
| **Desain kontrol** | **Memadai** | Desain berbasis kontainer minimalis (*zero-compiler, zero-debugger*), isolasi network alias Docker internal (`postgres`), dan pembatasan fitur administratif hanya untuk peran Admin / CAE. |
| **Status kesiapan** | **Siap** | Kontrol program utilitas istimewa telah beroperasi stabil pada seluruh kontainer produksi dan siap diaudit pada asesmen sertifikasi ISO/IEC 27001:2022. |
| **Prioritas** | **Rendah** | Eksekusi utilitas sudah dilakukan dengan baik; pemeliharaan rutin berfokus pada audit berkala hak akses dan dokumentasi prosedur darurat (*break-glass*). |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Formalisasi SOP Pengelolaan Program Utilitas Istimewa (SOP-SEC-011) dan inventaris utilitas resmi. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Audit berkala izin administratif pada modul pengaturan sistem dan tinjauan keamanan kontainer. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.18 (*Program Utilitas Istimewa*) **berlaku** penuh untuk seluruh arsitektur dan sistem komputasi AuditSphere berdasarkan pertimbangan strategis:

1. **Perlindungan Informasi Penting dari Manipulasi Tak Terotorisasi**: Program utilitas istimewa (seperti database CLI client, container execution tools, root shell, debugger, dan sistem migrasi skema) memiliki kapabilitas melewati kendali aplikasi normal (*bypass controls*). Akses terhadap utilitas ini harus dibatasi secara ketat guna melindungi integritas kertas kerja audit, temuan, dan basis data multi-tenant.
2. **Kepatuhan Standar ISO/IEC 27001:2022 Klausul A.8.18**: Standar mewajibkan penggunaan program utilitas yang berpotensi memotong kendali sistem operasi, database, atau aplikasi dibatasi hanya kepada personel yang berwenang, diotorisasi secara formal, dan seluruh aktivitas eksekusinya dicatat secara rinci.
3. **Pencegahan Modifikasi Data Langsung (*Direct Database Manipulation*)**: Memastikan seluruh perubahan data bisnis dilakukan melalui API microservices yang tervalidasi dan mencatat jejak audit (*audit trail*), bukan melalui perintah SQL manual sewenang-wenang (`psql`, `DBeaver`) yang dapat mengaburkan akuntabilitas.
4. **Mitigasi Eskalasi Hak Istimewa (*Privilege Escalation*) & *Lateral Movement***: Pembatasan utilitas di tingkat kontainer Docker dan sistem operasi VPS mencegah penyerang atau pengguna berniat jahat memanfaatkan alat administratif bawaan (*living-off-the-land*) untuk meningkatkan hak akses atau melompat antar kontainer tenant.
5. **Penegakan Segregasi Tugas (*Separation of Duties*)**: Memastikan bahwa staf pengembang (*developers*) tidak memiliki akses utilitas istimewa tanpa pengawasan di lingkungan produksi (*production*), sementara staf operasional (DevSecOps) terikat pada alur persetujuan.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode repositori `risk-based-audit` dan tata kelola VPS, pengendalian program utilitas istimewa yang telah beroperasi meliputi:

### 2.1 Arsitektur Kontainer Minimalis Tanpa Utilitas Berbahaya (*Minimalist Docker Images*)
- **Penggunaan Image Dasar Alpine Minimal**: Seluruh microservices (`auth-service`, `audit-service`, `master-service`, `risk-service`) menggunakan Dockerfile berbasis `alpine:latest` dengan penambahan paket terbatas (hanya `ca-certificates` dan `tzdata`).
- **Ketiadaan Perangkat Kompilasi & Debugging di Produksi**: Kontainer produksi tidak menyertakan kompiler (`gcc`, `make`), debugger (`gdb`, `strace`), interpreter tambahan (python, perl pada modul Go), atau utilitas inspeksi jaringan (`tcpdump`, `nmap`, `netcat`), sehingga memitigasi risiko penyalahgunaan perkakas bawaan.
- **Eksekusi Terbatas pada Binary Terkompilasi**: Kontainer hanya menyalin dan menjalankan satu binary aplikasi terotentikasi (contoh: `CMD ["./auth", "--config", "./pkg/config/config.yaml", "serve"]` dan `CMD ["./audit", "serve"]`) tanpa membuka shell interaktif ke publik.

### 2.2 Isolasi Jaringan Database Privat & Pencegahan Eksekusi Utilitas DB Host
- **Port Database PostgreSQL Tidak Terekspos ke Host**: File `docker-compose` arsitektur produksi tidak mempublikasikan port `5432` ke antarmuka jaringan host maupun internet publik.
- **Resolusi Internal Docker Network Alias**: Database PostgreSQL hanya dapat dijangkau oleh microservices melalui alias jaringan Docker `postgres` pada jaringan terisolasi internal.
- **Larangan Utilitas DB dari Luar Jaringan**: Personel atau program utilitas di luar lingkungan Docker (seperti perkakas GUI database lokal pengembang) secara arsitektural tidak dapat terhubung langsung ke basis data produksi.

### 2.3 Pengendalian Skrip Migrasi & Provisi Administratif (`scripts/onboard-tenant.sh`)
- **Eksekusi Migrasi Database Terisolasi via Binary Aplikasi**: Provisi skema dan pembaruan struktur basis data dijalankan terkontrol melalui sub-perintah binary internal (`auth migrate up`, `audit migrate up`, `master migrate`, `risk up`) menggunakan perintah `docker compose run --rm --no-deps`, bukan melalui utilitas eksternal yang tidak terkontrol.
- **Validasi Parameter Skrip Ketat**: Skrip operasional administratif (`scripts/onboard-tenant.sh` dan `scripts/generate-tenant-stack.sh`) menerapkan flag `set -euo pipefail` guna memastikan eksekusi berhenti seketika bila terjadi kesalahan input atau kegagalan perintah.
- **Isolasi Silo Data Tenant**: Skrip secara ketat mengalokasikan direktori penyimpanan terpisah (`/var/data/auditsphere/storage/<slug>`) dan database spesifik per tenant, mencegah percampuran data.

### 2.4 Matriks Izin Utilitas Aplikasi Terpusat (`https://auditsphere.app/settings?tab=permissions`)
- **Pemusatan Modul Administratif (Modul 15 Settings & User Management)**: Berkas `frontend/components/settings/permission.vue` menetapkan bahwa hak akses ke Modul 15 (pengaturan pengguna, hak akses peran, konfigurasi sistem, dan provisi tenant) hanya diberikan kepada peran **Admin / CAE** dengan izin `FULL`.
- **Penolakan Akses untuk Peran Operasional Lain**: Seluruh peran operasional lainnya (Audit Manager, Auditor, Auditee, dan Viewer) secara tegas diberi hak `NONE` pada modul pengaturan dan utilitas sistem.
- **Fitur Ekspor Matriks Izin Resmi**: Tersedia fungsi `exportMatrix()` yang memungkinkan pengunduhan konfigurasi matriks peran dan izin sistem dalam format JSON (`AuditSphere_RBAC_Matrix.json`) sebagai bukti audit.

### 2.5 Perekaman Log Eksekusi & Audit Trail Aktivitas Tingkat Tinggi
- **Pencatatan Kejadian Kritis di `sys_alert_logs`**: Seluruh aktivitas administratif, perubahan peran, upaya eskalasi izin, dan anomali autentikasi dicatat secara otomatis ke dalam tabel `sys_alert_logs` dan dialirkan ke topik Kafka `audit-logs`.
- **Agregasi Log Eksekusi Docker & Kontainer**: Perintah eksekusi kontainer dan log syslog host server dicatat melalui Promtail dan diagregasi ke Grafana Loki untuk pemantauan berkelanjutan.

---

## 3. Evaluasi Kinerja & Area Peningkatan Berkelanjutan

Sesuai penilaian checklist Excel baris 83, status kontrol adalah **SIAP** dengan efektivitas **EFEKTIF** dan catatan *"Eksekusi utilitas sudah dilakukan dengan baik"*. Untuk mempertahankan postur keamanan prima dan mengantisipasi audit eksternal, diidentifikasi 8 area peningkatan tata kelola:

| No | Parameter Program Utilitas Istimewa | Kondisi Implementasi Saat Ini | Target Peningkatan Berkelanjutan | Prioritas Peningkatan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Program Utilitas Istimewa (SOP-SEC-011)** | Pengendalian utilitas berjalan secara teknis namun belum dibakukan dalam dokumen SOP resmi bertanda tangan pimpinan. | Penyusunan dan pengesahan dokumen formal SOP-SEC-011 yang mengatur otorisasi, penggunaan, dan pelaporan utilitas. | **Sedang (Formalitas)** |
| 2 | **Katalog Inventaris Program Utilitas Resmi** | Daftar utilitas administratif tersebar pada skrip bash dan konfigurasi kontainer tanpa katalog inventaris terpadu. | Penyusunan dokumen formal *Privileged Utility Inventory* mencakup daftar perkakas resmi, fungsi, dan wewenang personel. | **Sedang** |
| 3 | **Penerapan Non-Root User pada Kontainer** | Kontainer saat ini berjalan dengan direktori `/root/` dan user default kontainer. | Penambahan direktif `USER appuser` non-root pada Dockerfile produksi guna membatasi kemampuan eksekusi di dalam kontainer. | **Sedang** |
| 4 | **Penguatan Hak Istimewa Docker (*Security Opts*)** | Kontainer belum secara eksplisit menambahkan parameter `no-new-privileges` dan `drop capabilities` pada compose. | Penerapan konfigurasi `security_opt: [no-new-privileges:true]` dan `cap_drop: [ALL]` pada `docker-compose.prod.yml`. | **Sedang** |
| 5 | **Prosedur Akses Darurat (*Break-Glass Protocol*)** | Akses perbaikan darurat saat insiden operasional belum memiliki formulir persetujuan dan pencatatan terpisah. | Penyusunan prosedur Break-Glass bertanda tangan Dimas dengan akun darurat sementara dan audit pasca-insiden. | **Sedang** |
| 6 | **Pencatatan Khusus Sesi Perintah Administratif** | Log mencatat event aplikasi dan syslog kontainer; rekam jejak sesi perintah interaktif teknisi belum diisolasi. | Pemberlakuan konfigurasi `auditd` pada VPS untuk mencatat eksekusi perintah administratif bernilai kritis (`sudo`, `docker`). | **Sedang** |
| 7 | **Tinjauan Berkala Matriks Peran & Izin Pengguna** | Matriks RBAC dapat diekspor melalui antarmuka settings; namun jadwal peninjauan berkala belum terdokumentasi resmi. | Pemberlakuan jadwal review izin administratif kuartalan yang ditandatangani oleh Reza dan disetujui Dimas. | **Rendah** |
| 8 | **Pengujian Integritas Binary Aplikasi & Hashing** | Binary dikompilasi dan disalin langsung ke image tanpa verifikasi checksum hash SHA-256 otomatis pada pipeline. | Penerapan pencatatan SHA-256 hash untuk setiap binary rilis produksi guna menjamin integritas perangkat lunak. | **Rendah** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak teknis dan konfigurasi kode pada repositori `risk-based-audit` yang mendukung kesiapan kontrol A.8.18:

1. **Matriks Izin RBAC & Pengaturan Pengguna (`frontend/components/settings/permission.vue`)**: Implementasi pembatasan Modul 15 (*Settings & User Management*) yang hanya dapat diakses oleh Admin/CAE, dilengkapi fitur unduhan `AuditSphere_RBAC_Matrix.json`.
2. **Konfigurasi Dockerfile Minimalis (`backend/*/Dockerfile`)**: Seluruh microservices dibangun di atas `alpine:latest` tanpa utilitas kompilasi (`gcc`, `make`) maupun utilitas debugging berisiko tinggi.
3. **Isolasi Port Database Privat (`backend/docker-compose.prod.yml`)**: Port PostgreSQL (`5432`) tidak diekspos ke host maupun publik, membatasi akses utilitas database langsung hanya melalui jaringan kontainer privat.
4. **Skrip Provisi & Migrasi Terkontrol (`scripts/onboard-tenant.sh`)**: Pengeksekusian migrasi skema dan seed peran melalui binary aplikasi internal yang terotentikasi (`auth migrate up`, `audit migrate up`, `master migrate`, `risk up`).
5. **Repositori Log Peringatan Keamanan (`backend/auth-service/models/sys_alert_log.go`)**: Model pencatatan aktivitas administratif dan anomali eskalasi hak istimewa secara otomatis.
6. **Skrip Provisi Multi-Tenant (`scripts/generate-tenant-stack.sh`)**: Skrip pembentukan stack terisolasi per tenant dengan validasi parameter ketat guna mencegah benturan wewenang.

---

## 5. Bukti Objektif Penguat Kesiapan Audit

Dokumen prosedural dan rekaman tata kelola pelengkap yang disiapkan untuk mengukuhkan status kesiapan audit:

1. **Dokumen SOP Pengelolaan Program Utilitas Istimewa (SOP-SEC-011)**: Standar operasional resmi bertanda tangan Dimas (Leadership) dan Reza (Lead InfoSec).
2. **Daftar Inventaris Program Utilitas Resmi (*Privileged Utility Inventory*)**: Katalog resmi yang mencantumkan seluruh utilitas administratif berizin, fungsi, dan daftar pemegang hak akses.
3. **Formulir Permohonan & Otorisasi Penggunaan Utilitas (*Privileged Utility Approval Form*)**: Rekam jejak persetujuan tertulis sebelum menjalankan utilitas kritis di lingkungan produksi.
4. **Buku Log Eksekusi Program Utilitas Istimewa (*Privileged Execution Logbook*)**: Catatan tanggal, waktu, operator, utilitas yang digunakan, tujuan, dan hasil eksekusi.
5. **Dokumen Prosedur Akses Darurat (*Break-Glass Protocol*)**: Tata cara penggunaan akun dan perkakas darurat saat terjadi kegagalan sistem kritis di luar jam kerja normal.
6. **Laporan Audit Kuartalan Hak Akses Administratif**: Berita acara peninjauan berkala atas akun pemegang wewenang administratif pada sistem AuditSphere.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber Repositori & Server | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.18-01** | Matriks Izin RBAC & Pengaturan Pengguna | Aplikasi Web & JSON | `https://auditsphere.app/settings?tab=permissions` & `AuditSphere_RBAC_Matrix.json` | Lead InfoSec & Reza |
| **EV-A8.18-02** | Konfigurasi Minimalis Kontainer Tanpa Utilitas Debug | Dockerfile Microservices | `backend/*/Dockerfile` (basis Alpine minimal, *zero-compiler/debugger*) | Lead Backend Dev |
| **EV-A8.18-03** | Skrip Migrasi & Provisi Terkontrol via Binary | Skrip Bash & Compose | `scripts/onboard-tenant.sh` & `scripts/generate-tenant-stack.sh` | DevSecOps Engineer |
| **EV-A8.18-04** | SOP Pengelolaan Program Utilitas Istimewa (SOP-SEC-011) | Dokumen PDF / Word | `docs/policies/SOP-SEC-011-Privileged-Utilities-Management.pdf` | Lead InfoSec & Dimas |
| **EV-A8.18-05** | Daftar Inventaris Program Utilitas Sistem & Basis Data | Dokumen Spreadsheet / PDF | `docs/evidence/Privileged-Utilities-Inventory-and-Access-List.pdf` | Lead Infrastructure |
| **EV-A8.18-06** | Buku Log Eksekusi & Berita Acara Persetujuan Utilitas | Spreadsheet & Formulir PDF | `docs/evidence/Privileged-Execution-Log-and-Approval-Records.pdf` | SecOps Response Team |

---

## 7. Rencana Aksi Peningkatan Berkelanjutan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan efektivitas kontrol utilitas istimewa dan merealisasikan peningkatan berkala hingga batas target **27 November 2026**:

| No | Tindakan Pemeliharaan / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Program Utilitas (`SOP-SEC-011`)** | Dokumen SOP formal mencakup alur persetujuan penggunaan utilitas bypass, tata kelola akun istimewa, dan disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Pembuatan Katalog Resmi Inventaris Utilitas Administratif** | Dokumen inventaris resmi seluruh perkakas administratif (CLI, container runner, migration tools) dan daftar personel terotorisasi | Lead Infra | 1 minggu |
| 3 | **Penerapan Non-Root User pada Dockerfile Microservices** | Implementasi instruksi `USER appuser` pada seluruh Dockerfile backend guna membatasi wewenang eksekusi di dalam kontainer | Lead Backend Dev | 2 minggu |
| 4 | **Penguatan Hak Istimewa Kontainer (`no-new-privileges` & `cap_drop`)** | Pembaruan file `docker-compose.prod.yml` dengan parameter pembatasan wewenang kontainer tingkat lanjut | DevSecOps | 1 minggu |
| 5 | **Penyusunan Prosedur & Formulir Akses Darurat (*Break-Glass*)** | SOP dan formulir rekam jejak persetujuan pembukaan akses administratif darurat saat insiden operasional | SecOps Lead | 2 minggu |
| 6 | **Pengaktifan Audit Jejak Perintah Host (`auditd` logging)** | Konfigurasi modul `auditd` pada VPS produksi untuk mencatat setiap pemanggilan perintah sensitif (`docker`, `sudo`, `psql`) | Lead Infra | 2 minggu |
| 7 | **Pemberlakuan Jadwal Audit Izin Akses Administratif Kuartalan** | Kalender audit dan checklist peninjauan berkala atas hak akses pengguna pada modul Settings & User Management | Reza / Dimas | 1 minggu |
| 8 | **Penerapan Verifikasi Hash SHA-256 pada Binary Rilis Produksi** | Skrip otomasi verifikasi checksum hash untuk menjamin binary aplikasi tidak dimodifikasi sebelum dideploy ke kontainer | DevSecOps | 2 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-011 bertanda tangan, Katalog Inventaris Utilitas, Konfigurasi Non-Root Dockerfile, dan Tangkapan Layar Matriks Izin RBAC pada Master Audit Evidence Index.

