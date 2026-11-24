# A.8.31 Pemisahan Lingkungan Development, Test, dan Production

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pemisahan lingkungan development, test, dan production |
| **Fokus pemeriksaan** | Lingkungan dan akses dipisahkan untuk mencegah perubahan tidak sah |
| **Contoh bukti objektif** | Environment diagram, IAM role, pipeline control, access test |
| **Sistem** | AuditSphere (auditsphere.app), Docker Compose (dev vs prod), Kong Gateway, PostgreSQL, GitHub Actions |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 12 Oktober 2026 (Ref: 46307.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Pemisahan lingkungan dilakukan untuk menghindari konflik dan crash yang terjadi ketika ada update pada aplikasi AuditSphere serta mencegah kebocoran data audit rahasia. |
| **Status implementasi** | **Diterapkan** | Pemisahan tegas lingkungan development, staging, dan production dengan port dan kredensial terpisah telah aktif secara operasional. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia dokumen arsitektur komparatif, panduan setup demo/staging, panduan provisi domain klien, serta file konfigurasi Docker Compose terpisah dev vs prod. |
| **Evidence ID / URL** | `architecture_on_premises_vs_cloud.md, AUDITSPHERE_DEMO_SETUP_GUIDE.md, NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md` | Dokumentasi arsitektur resmi di folder `docs/` dan repositori docker-compose, didukung registrasi bukti EV-A8.31-01 s/d EV-A8.31-06. |
| **Efektivitas operasional** | **Efektif** | Pemisahan tegas lingkungan development, staging, dan production dengan port dan kredensial terpisah terbukti efektif mencegah interferensi kode uji coba ke sistem aktif. |
| **Desain kontrol** | **Memadai** | Desain multi-layer mencakup pemisahan jaringan bridge Docker, isolasi volume database, kredensial `.env` terpisah, dan segregasi pipeline CI/CD. |
| **Status kesiapan** | **Siap** | Struktur arsitektur pemisahan lingkungan telah teruji, stabil di lingkungan live produksi dan staging, serta siap diaudit pada asesmen ISO/IEC 27001:2022. |
| **Prioritas** | **Kritis** | Pertahankan pemisahan lingkungan development, staging, dan production dan lakukan evaluasi lingkungan secara berkala untuk menjaga integritas data. |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Fokus pemeliharaan: evaluasi berkala konfigurasi firewall, audit kredensial, dan formalisasi SOP pemisahan lingkungan. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Peninjauan integritas isolasi jaringan kontainer dan audit sanitasi data uji coba. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.31 (*Pemisahan Lingkungan Development, Test, dan Production*) **berlaku** penuh untuk seluruh infrastruktur dan proses siklus pengembangan perangkat lunak AuditSphere berdasarkan pertimbangan strategis:

1. **Pencegahan Konflik & Kerusakan Sistem Saat Pembaruan Aplikasi**: Pemisahan lingkungan dilakukan untuk menghindari konflik dan crash yang terjadi ketika ada update pada aplikasi AuditSphere. Fitur baru dan penyesuaian skema database wajib diuji secara tuntas di lingkungan pengembangan dan staging sebelum dideploy ke produksi.
2. **Kepatuhan Standar ISO/IEC 27001:2022 Klausul A.8.31**: Standar mewajibkan pemisahan tegas antara lingkungan pengembangan (*development*), pengujian (*testing/staging*), dan operasional (*production*) guna mencegah akses tidak sah, kontaminasi data, dan perubahan tanpa otorisasi pada sistem produksi.
3. **Perlindungan Kerahasiaan Data Audit Finansial**: Sistem AuditSphere memproses kertas kerja audit sensitif, data investigasi kecurangan (*fraud*), dan informasi kepatuhan perbankan. Data produksi tidak boleh terekspos ke lingkungan pengembangan atau pengujian yang memiliki kendali keamanan lebih longgar.
4. **Pemberian Stabilitas Layanan & Penjaminan SLA 99.9%**: Aktivitas pengujian beban (*load testing*), uji penetrasi (*pentest*), dan eksperimen kecerdasan buatan (AI/ML) tidak boleh mengganggu responsivitas dan stabilitas sistem produksi yang digunakan auditor.
5. **Penegakan Segregasi Hak Akses & Tugas (*Separation of Duties*)**: Memastikan pengembang tidak memiliki akses administratif langsung ke basis data produksi, dan personel operasional terikat pada alur rilis resmi.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran berkas konfigurasi repositori `risk-based-audit` dan server VPS, implementasi pemisahan lingkungan mencakup:

### 2.1 Pemisahan File Konfigurasi Orkestrasi Docker Compose (Prod vs Dev)
- **Orkestrasi Produksi (`backend/docker-compose.prod.yml`)**: Menetapkan project name `rb_audit_backend` dengan jaringan bridge privat `rb_audit_network`, volume persisten terisolasi (`postgres_data`, `redis_data`, `kafka_data`), dan port database `5432` yang sama sekali tidak diekspos ke host.
- **Orkestrasi Development (`backend/docker-compose.dev.yml`)**: Menetapkan project name terpisah `rb_audit_backend_dev` dengan jaringan terisolasi `rb_audit_network_dev`, volume terpisah (`postgres_dev_data`, `redis_dev_data`), dan port forwarding khusus yang terikat ke localhost (`127.0.0.1:5433` untuk PostgreSQL dan `127.0.0.1:6380` untuk Redis) sehingga tidak berbenturan dengan produksi.
- **Orkestrasi Frontend Terpisah**: File `frontend/docker-compose.prod.yml` mengompilasi bundel statis Nuxt dengan mode produksi, sementara `frontend/docker-compose.dev.yml` menjalankan server pengembangan dengan fitur Hot Module Replacement (HMR).

### 2.2 Isolasi Jaringan Kontainer & Database State (Network & Data Silo)
- **Segmentasi Jaringan Bridge Independen**: Jaringan Docker `rb_audit_network` dan `rb_audit_network_dev` terpisah secara logis; kontainer development tidak dapat merutekan paket data ke kontainer produksi.
- **Instans Database dan Kontainer Terpisah**: Mengoperasikan instans PostgreSQL terpisah (`rb_audit_postgres` vs `rb_audit_postgres_dev`) dengan kredensial otentikasi berbeda, mencegah ketidaksengajaan modifikasi data produksi.
- **Isolasi Volume Penyimpanan Bukti Audit**: Direktori penyimpanan bukti audit produksi (`/var/data/auditsphere/storage/<slug>`) terpisah secara fisik dari direktori unggahan development (`backend/audit-service/uploads/<slug>`).

### 2.3 Pemisahan Konfigurasi Kong API Gateway (PROD vs DEV)
- **Kong Gateway Produksi (`backend/kong-gateway/kong/prod/kong.yml`)**: Berjalan pada kontainer `rb_audit_kong` dengan penegakan rate limiting ketat (100 req/min), deteksi bot otomatis (sqlmap, nikto, nmap), dan CORS terikat ke `https://auditsphere.app`.
- **Kong Gateway Development (`backend/kong-gateway/kong/dev/kong.yml`)**: Berjalan pada kontainer `rb_audit_kong_dev` dengan pengaturan rute pengujian dan port Admin API terpisah (8001 dev vs 8001 prod) untuk kebutuhan debugging.

### 2.4 Segregasi Pipeline CI/CD GitHub Actions (`.github/workflows/`)
- **Pipeline Produksi (`.github/workflows/deploy.yml`)**: Hanya dipicu pada push ke branch `main`, menargetkan server VPS produksi `202.10.34.166` dengan kredensial secret `VPS_SSH_KEY_PROD`.
- **Pipeline Development (`.github/workflows/deploy-dev.yml`)**: Dipicu pada branch `develop`/staging, menargetkan lingkungan pengujian dengan kredensial secret `VPS_SSH_KEY_DEV` terpisah.

### 2.5 Pengaturan Dataset Uji & Sanitasi Data (Demo vs Production Data)
- **Skrip Provisi Otomatis (`scripts/onboard-tenant.sh`)**: Menyediakan parameter pembeda tegas antara lingkungan produksi (`--empty-data` menghasilkan 0 record bisnis dan skema bersih) dan lingkungan pengujian/demo (`--with-demo-data` yang menginjeksi dataset sampel sintetis).
- **Larangan Penggunaan Data Nyata di Demo**: Dokumen `docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md` menegaskan bahwa data pengujian hanya berupa data tiruan (*mock data*) dan tidak pernah menyalin basis data produksi.

---

## 3. Evaluasi Kinerja & Area Peningkatan Berkelanjutan

Sesuai catatan checklist Excel baris 96 (*"Pertahankan pemisahan lingkungan development, staging, dan production dan lakukan evaluasi lingkungan secara berkala"*), status kesiapan kontrol adalah **SIAP** dan prioritas **KRITIS**. Untuk menjaga konsistensi isolasi jangka panjang, diidentifikasi 8 area pemeliharaan dan peningkatan berkelanjutan:

| No | Parameter Pemisahan Lingkungan | Kondisi Implementasi Saat Ini | Target Peningkatan Berkelanjutan | Prioritas Peningkatan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Pemisahan Lingkungan (SOP-SEC-013)** | Pemisahan berjalan secara teknis arsitektur namun belum dibakukan dalam dokumen SOP resmi bertanda tangan pimpinan. | Penyusunan dan pengesahan SOP-SEC-013 yang mengatur tata kelola lingkungan, promosi kode, dan aturan isolasi data. | **Tinggi (Formalitas)** |
| 2 | **Kebijakan Sanitasi Data Uji (Data Sanitization)** | Lingkungan demo menggunakan seed sintetis; belum ada aturan tertulis formal yang melarang clone data produksi ke staging. | Penyusunan Pedoman Anonimisasi dan Masking Data Uji bertanda tangan DPO/InfoSec Lead sesuai UU PDP. | **Tinggi (Kepatuhan)** |
| 3 | **Pemisahan Host VPS Fisik/Cloud Staging Terpisah** | Lingkungan dev dan prod berjalan pada compose terpisah namun sebagian masih berbagi server VPS yang sama. | Migrasi penuh staging/demo ke VPS atau instance cloud terpisah mandiri untuk isolasi komputasi total. | **Sedang** |
| 4 | **Manajemen Rahasia & Kredensial (.env Vault)** | File `.env` dipisahkan per direktori; belum menggunakan secret manager terpusat seperti HashiCorp Vault. | Implementasi pemusatan manajemen secret dengan rotasi otomatis dan pelarangan kredensial default di dev. | **Sedang** |
| 5 | **Audit Berkala Firewall & Aturan Port Binding** | Port dev PostgreSQL (`5433`) dibatasi ke `127.0.0.1`; verifikasi binding masih dilakukan secara manual. | Otomasi skrip verifikasi mingguan untuk memastikan tidak ada port internal yang terekspos ke `0.0.0.0`. | **Sedang** |
| 6 | **Segregasi Akun Akses Pengembang ke Server Prod** | Akses SSH produksi dibatasi pada private key tertentu; belum ada pembagian akun individual berbasis IAM. | Pemberlakuan akun SSH personal dengan log session recording untuk teknisi yang memiliki wewenang rilis. | **Sedang** |
| 7 | **Verifikasi Pengujian Regresi Lintas Lingkungan** | Pengujian dilakukan pada dev sebelum merge; belum ada automated smoke test pasca-deploy di staging. | Penambahan automated smoke test terprogram di GitHub Actions pasca deploy ke staging sebelum rilis prod. | **Sedang** |
| 8 | **Simulasi Peninjauan Akses Kuartalan Lintas Lingkungan** | Hak akses ke server dan database diperiksa berkala saat insidental. | Pemberlakuan jadwal review kuartalan formal terhadap daftar pemegang kunci SSH dan akun database dev vs prod. | **Rendah** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak konfigurasi dan dokumen arsitektur pada repositori `risk-based-audit` yang membuktikan kepatuhan kontrol A.8.31:

1. **Dokumen Desain Arsitektur Lingkungan (`docs/architecture_on_premises_vs_cloud.md`)**: Spesifikasi arsitektur komparatif on-premises vs cloud yang merinci pemisahan multi-tier (DMZ, App Tier, Database Tier, AI Tier).
2. **Panduan Setup Lingkungan Demo/Staging (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md`)**: Panduan lengkap isolasi lingkungan demo terhadap produksi, mencakup tabel komparasi Demo vs Production.
3. **Panduan Provisi Domain Klien & Multi-Tenancy (`docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md`)**: Dokumentasi arsitektur silo kontainer yang mereplikasi pemisahan dev/prod per tenant.
4. **File Konfigurasi Docker Compose Produksi (`backend/docker-compose.prod.yml`)**: Konfigurasi stack produksi dengan jaringan `rb_audit_network` dan penutupan port database.
5. **File Konfigurasi Docker Compose Development (`backend/docker-compose.dev.yml`)**: Konfigurasi stack development dengan jaringan terpisah `rb_audit_network_dev` dan port binding lokal (`127.0.0.1:5433`).
6. **Konfigurasi Kong Gateway Dev vs Prod**: Pemisahan `kong.yml` produksi (rate limiting & bot detection) dan development (jalur testing).
7. **Workflow CI/CD Terpisah (`.github/workflows/deploy.yml` vs `deploy-dev.yml`)**: Pipeline rilis independen untuk branch main dan branch develop.

---

## 5. Bukti Objektif Penguat Kesiapan Audit

Dokumen prosedural dan rekaman tata kelola pelengkap yang disiapkan untuk mengukuhkan status "Siap" pada audit sertifikasi ISO 27001:

1. **Dokumen SOP Pemisahan Lingkungan Dev, Test, dan Prod (SOP-SEC-013)**: Standar operasional resmi bertanda tangan Dimas (Leadership) dan Reza (Lead InfoSec).
2. **Kebijakan Sanitasi Data Uji & Larangan Penggunaan Data Produksi**: Kebijakan tertulis yang mengatur pelarangan mutlak transfer data audit riil ke lingkungan non-produksi.
3. **Diagram Arsitektur Lingkungan Terpadu (*Environment Network Topology Diagram*)**: Peta topologi jaringan resmi yang memvisualisasikan segmentasi firewall dev, staging, dan prod.
4. **Matriks Segregasi Hak Akses Kredensial & Kunci SSH**: Daftar personel yang memiliki wewenang akses per lingkungan (Development, Staging, Production).
5. **Berita Acara Evaluasi & Audit Berkala Pemisahan Lingkungan**: Formulir inspeksi kuartalan yang memverifikasi tidak adanya kebocoran rute atau port terbuka.
6. **Laporan Pengujian Penetrasi Isolasi Jaringan Kontainer**: Bukti empiris bahwa kontainer dev tidak dapat mengakses basis data produksi.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber Repositori | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.31-01** | Konfigurasi Docker Compose Dev vs Prod Terpisah | File YAML Compose | `backend/docker-compose.prod.yml` & `backend/docker-compose.dev.yml` | DevSecOps Engineer |
| **EV-A8.31-02** | Dokumen Arsitektur & Topologi Pemisahan Lingkungan | Markdown & Diagram | `docs/architecture_on_premises_vs_cloud.md` & `AUDITSPHERE_DEMO_SETUP_GUIDE.md` | Lead Infrastructure |
| **EV-A8.31-03** | Panduan Provisi Domain & Isolasi Silo Tenant | Markdown Guide | `docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md` & `scripts/onboard-tenant.sh` | Reza / DevSecOps |
| **EV-A8.31-04** | Workflow Segregasi Pipeline CI/CD GitHub Actions | File YAML GitHub Actions | `.github/workflows/deploy.yml` & `.github/workflows/deploy-dev.yml` | Lead Backend Dev |
| **EV-A8.31-05** | SOP Pemisahan Lingkungan Dev, Test, & Prod (SOP-SEC-013) | Dokumen PDF / Word Resmi | `docs/policies/SOP-SEC-013-Environment-Separation.pdf` | Lead InfoSec & Dimas |
| **EV-A8.31-06** | Kebijakan Sanitasi Data Uji & Berita Acara Audit Akses | Dokumen PDF & Matriks | `docs/evidence/Test-Data-Sanitization-and-Access-Matrix.pdf` | SecOps Response Team |

---

## 7. Rencana Aksi Pemeliharaan & Peningkatan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan pemisahan lingkungan development, staging, dan production hingga batas target **27 November 2026**:

| No | Tindakan Pemeliharaan / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pemisahan Lingkungan (`SOP-SEC-013`)** | Dokumen SOP formal mencakup aturan isolasi kode, kredensial, dan tata cara promosi yang disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Formalisasi Kebijakan Sanitasi & Anonimisasi Data Uji** | Aturan tertulis pelarangan penyalinan data audit produksi ke lingkungan dev/test sesuai UU PDP | Lead InfoSec | 1 minggu |
| 3 | **Pengujian Verifikasi Pembatasan Port Binding Host** | Skrip pengujian otomatis untuk memverifikasi seluruh port database dev terikat ketat ke `127.0.0.1` | DevSecOps | 1 minggu |
| 4 | **Pemisahan Total Host Staging ke VPS/Server Fisik Independen** | Migrasi kontainer dev/staging ke host terpisah guna menjamin isolasi komputasi dan memori total | Lead Infra | 3 minggu |
| 5 | **Audit Matriks Hak Akses Kredensial & Kunci SSH Lintas Lingkungan** | Penerbitan daftar otorisasi pemegang kunci SSH produksi vs development bertanda tangan pimpinan | Reza / Dimas | 1 minggu |
| 6 | **Implementasi Automated Smoke Test Pasca-Deploy Staging** | Pemasangan automated smoke test pada workflow GitHub Actions untuk validasi staging sebelum prod | Lead Backend | 2 minggu |
| 7 | **Pembaruan Diagram Topologi Jaringan Lingkungan Terpadu** | Dokumentasi visual arsitektur jaringan resmi yang memetakan subnet, VLAN, dan aturan firewall | Lead Infra | 2 minggu |
| 8 | **Simulasi Audit Kuartalan Evaluasi Pemisahan Lingkungan** | Pelaksanaan uji coba pemutusan koneksi untuk membuktikan ketiadaan interkoneksi dev ke prod DB | SecOps Team | 1 minggu |

Setelah tindakan 1, 2, 3, dan 5 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-013 bertanda tangan, Kebijakan Sanitasi Data Uji, File Compose Dev vs Prod, dan Matriks Hak Akses Kredensial pada Master Audit Evidence Index.

