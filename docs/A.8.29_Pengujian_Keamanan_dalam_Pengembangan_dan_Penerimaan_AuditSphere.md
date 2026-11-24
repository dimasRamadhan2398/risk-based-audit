# A.8.29 Pengujian Keamanan dalam Pengembangan dan Penerimaan

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pengujian keamanan dalam pengembangan dan penerimaan |
| **Fokus pemeriksaan** | Pengujian keamanan direncanakan dan hasilnya ditindaklanjuti |
| **Contoh bukti objektif** | Test plan, SAST/DAST, pentest, defect ticket, sign-off |
| **Sistem** | AuditSphere (auditsphere.app), CI/CD GitHub Actions, Kong Gateway, Go Microservices, Nuxt Frontend, PostgreSQL |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 12 Oktober 2026 (Ref: 46307.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Berbagai pengujian keamanan dilakukan untuk memperkuat keamanan aplikasi AuditSphere serta memastikan tidak ada celah kerentanan kritis sebelum rilis produksi. |
| **Status implementasi** | **Dalam proses** | Pengujian penetrasi pre-pentest dipersiapkan resmi dengan LoA dan Scope & Window baku; namun SAST/DAST otomatis di CI/CD dan pentest eksternal independen masih dalam tahap finalisasi. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia dokumen API Testing Guide, UAT Test Cases, checklist keamanan API (ASC-001), dan kerangka LoA; namun laporan resmi pentest independen dan security sign-off sheet belum lengkap. |
| **Evidence ID / URL** | `risk-based-audit\docs\ISO_27001, API_Testing_Guide.md` | Dokumen panduan pengujian API (`docs/API_Testing_Guide.md`) dan bundel kepatuhan ISO 27001, didukung registrasi usulan bukti EV-A8.29-01 s/d EV-A8.29-06. |
| **Efektivitas operasional** | **Sebagian efektif** | Pengujian penetrasi pre-pentest dipersiapkan resmi dengan LoA dan Scope & Window baku; unit/integration test berjalan, namun pemindaian keamanan otomatis belum terintegrasi di pipeline deploy. |
| **Desain kontrol** | **Memadai** | Desain mencakup tahapan pengujian berlapis (Secure Coding Guide, API Security Checklist OWASP, UAT bisnis, automated unit test, dan rencana penetrasi berkala). |
| **Status kesiapan** | **Belum siap** | Laporan hasil pengujian penetrasi independen pihak ketiga, bukti integrasi SAST/DAST di pipeline CI/CD, serta berita acara Security Sign-Off formal belum terpenuhi. |
| **Prioritas** | **Kritis** | Melanjutkan pengujian dengan melakukan Automated SAST & DAST di CI/CD Pipeline dan Penetration Testing (Pentest) Independen guna menutup celah sebelum audit Stage 2. |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Fokus akselerasi: integrasi alat SAST/DAST di GitHub Actions dan pelaksanaan uji penetrasi eksternal independen. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Tinjauan progres integrasi pipeline SAST, peninjauan draf LoA pentest, dan verifikasi tiket remediasi QA. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.29 (*Pengujian Keamanan dalam Pengembangan dan Penerimaan*) **berlaku** untuk seluruh ekosistem pengembangan perangkat lunak AuditSphere berdasarkan pertimbangan strategis:

1. **Penguatan Ketahanan Sistem Terhadap Eksploitasi Siber**: Berbagai pengujian keamanan dilakukan untuk memperkuat keamanan aplikasi AuditSphere, mendeteksi kerentanan kode, kelemahan arsitektur, dan kesalahan konfigurasi sebelum aplikasi dirilis ke lingkungan produksi multi-tenant.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.8.29**: Standar mewajibkan organisasi menetapkan, merencanakan, mendokumentasikan, dan melaksanakan pengujian keamanan selama siklus hidup pengembangan (SDLC) serta sebagai syarat mutlak penerimaan sistem (*acceptance criteria*).
3. **Perlindungan Data Rahasia Lembaga Jasa Keuangan**: AuditSphere memproses kertas kerja audit sensitif, data audit internal bank/lembaga keuangan, temuan audit kepatuhan, dan data pribadi (UU PDP No. 27/2022). Pengujian keamanan yang ketat menjamin tidak adanya kebocoran data antar-tenant (*cross-tenant leakage*).
4. **Validasi Efektivitas Kontrol Teknis (*Security Assurance*)**: Pengujian keamanan berfungsi sebagai pembuktian empiris bahwa kontrol autentikasi (JWT/MFA), otorisasi (RBAC), enkripsi data, rate limiting Kong, dan logging bekerja sebagaimana mestinya di bawah simulasi serangan nyata.
5. **Kepatuhan Kriteria Penerimaan Rilis Formal**: Memastikan setiap rilis versi baru memiliki berita acara penerimaan keamanan (*Security Sign-Off*) resmi yang ditandatangani oleh pemangku kepentingan sebelum diizinkan melayani traffic operasional.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran dokumentasi dan repositori `risk-based-audit`, artefak dan pengujian yang telah diterapkan meliputi:

### 2.1 Panduan Pengujian API Terstruktur (`docs/API_Testing_Guide.md`)
- **Spesifikasi Pengujian Endpoint Komprehensif**: Dokumen `docs/API_Testing_Guide.md` mendefinisikan skenario uji fungsional, parameter validasi, kode status HTTP, paginasi, dan struktur respons error untuk `master-service` dan `auth-service`.
- **Koleksi Pengujian Otomatis Postman**: Tersedia berkas `docs/Postman_Collection_RBIA_Master_Service.json` untuk pengujian integrasi API otomatis dan validasi payload.
- **Skrip Uji API Otomatis**: Berkas `backend/auth-service/scripts/test-api.sh` menyediakan skrip eksekusi pengujian endpoint autentikasi berbasis bash cURL.

### 2.2 Checklist Keamanan API Berbasis OWASP (`docs/security/ASC-001_API_Security_Checklist.docx`)
- **Pemetaan Risiko OWASP API Security Top 10 (2023)**: Dokumen resmi ASC-001 memetakan seluruh endpoint Kong API Gateway terhadap kerentanan Broken Object Level Authorization (BOLA), Broken Authentication, Unrestricted Resource Consumption, dan Server-Side Request Forgery (SSRF).
- **Evaluasi Pengamanan File Upload & Media**: Analisis keamanan mendalam terhadap handler unggah berkas (`audit-service/controllers/media`) dan pembatasan tipe MIME guna mencegah eksekusi kode berbahaya.

### 2.3 Kerangka Uji Penerimaan Pengguna (`docs/AuditSphere_UAT_Meeting_Ready_dengan_Test_Case.docx`)
- **Skenario Uji Penerimaan Bisnis & Keamanan**: Dokumen UAT mendefinisikan puluhan kasus uji terstruktur (termasuk AS-UAT-026 Document Upload, kontrol hak akses peran, dan validasi kertas kerja) dengan kriteria penerimaan yang jelas.
- **Format Sign-Off Penerimaan Formal**: Templat keputusan penerimaan sistem (Diterima, Diterima dengan Catatan, Ditolak) berbasis bukti eksekusi pengguna.

### 2.4 Persiapan Pengujian Penetrasi Pre-Pentest (LoA & Scope & Window)
- **Standardisasi Letter of Authorization (LoA)**: Kerangka kerja surat izin penetrasi resmi yang mendefinisikan wewenang penguji, batasan pengujian (*rules of engagement*), dan kontak darurat eskalasi.
- **Definisi Scope & Testing Window Baku**: Ruang lingkup target pengujian yang mencakup Kong Gateway, API endpoints, web frontend, dan database dengan alokasi jendela waktu pengujian di luar jam sibuk operasional guna mencegah gangguan layanan.

### 2.5 Pengujian Unit & Integrasi Kode Otomatis (`*_test.go` & `vitest`)
- **Pengujian Unit Backend Go**: Puluhan berkas pengujian unit aktif pada microservices (seperti `auth_test.go`, `password_test.go`, `crud_handler_test.go`, `status_guard_test.go`) yang memvalidasi logika otentikasi, hashing password, dan RBAC guard.
- **Pengujian Unit Frontend Nuxt**: Pustaka vitest pada frontend (`frontend/vitest.config.ts`) dengan test suite unit pengujian validasi input dan logika rendering peran.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.8.29 menunjukkan 8 kesenjangan yang menjadi dasar status **Belum Siap** dan prioritas **Kritis**:

| No | Parameter Pengujian Keamanan | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Automated SAST di CI/CD Pipeline** | Pemindaian kode statis otomatis (`gosec`/`semgrep`/`SonarQube`) pada setiap pull request/push ke main. | Workflow CI/CD (`.github/workflows/deploy.yml`) hanya menjalankan build dan deploy; belum ada stage SAST. | **Gap Kritis (CI/CD)** |
| 2 | **Automated DAST di CI/CD Pipeline** | Pemindaian keamanan dinamis otomatis (OWASP ZAP/Nuclei) terhadap endpoint staging sebelum promosi ke produksi. | DAST otomatis belum diintegrasikan ke dalam pipeline rilis perangkat lunak. | **Gap Kritis (CI/CD)** |
| 3 | **Laporan Penetration Testing Independen** | Laporan uji penetrasi resmi dari lembaga penguji pihak ketiga independen minimal 1 kali per tahun. | Pre-pentest telah disiapkan secara formal dengan LoA, namun laporan final pentest independen belum terbit. | **Gap Kritis (Audit)** |
| 4 | **Pemindaian Kerentanan Dependensi (SCA)** | Pemeriksaan otomatis pustaka pihak ketiga (`govulncheck`, `npm audit`, `Trivy`) untuk mendeteksi CVE terbuka. | Pemeriksaan dependensi masih dilakukan secara manual/insidental tanpa gerbang penahan rilis otomatis. | **Gap Ketergantungan** |
| 5 | **Sistem Tiket Defect Keamanan & Remediasi** | Tiket formal untuk setiap temuan kerentanan pengujian dengan SLA penutupan (Kritis: 7 hari, Tinggi: 14 hari). | Temuan pengujian masih dicatat secara informal dalam catatan review; belum terintegrasi ke sistem issue formal. | **Gap Pelacakan Celah** |
| 6 | **Berita Acara Security Sign-Off Rilis** | Formulir persetujuan rilis formal bertanda tangan Lead InfoSec dan QA sebelum versi baru diaktifkan. | Penerimaan sistem saat ini didasarkan pada kelulusan UAT fungsional tanpa lembar tanda tangan sign-off keamanan khusus. | **Gap Formalitas** |
| 7 | **Pengujian Keamanan Multi-Tenancy Khusus** | Uji penetrasi khusus pembuktian isolasi basis data dan ketiadaan kebocoran data antar-tenant (*cross-tenant leakage*). | Isolasi diatur pada skrip; namun laporan pengujian penetrasi spesifik tenant isolation belum didokumentasikan. | **Gap Uji Isolasi** |
| 8 | **SOP Prosedur Pengujian Keamanan (SOP-SEC-012)** | Dokumen SOP formal yang membakukan metodologi pengujian keamanan perangkat lunak dari dev hingga rilis. | Pedoman pengkodean aman (SCG-001) tersedia; namun dokumen SOP pengujian dan penerimaan formal belum disahkan. | **Gap Tata Kelola** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak teknis dan dokumen perencanaan yang telah tersedia pada repositori `risk-based-audit`:

1. **Dokumen Panduan Pengujian API (`docs/API_Testing_Guide.md`)**: Petunjuk pengujian fungsional dan integritas data untuk `master-service` dan `auth-service`.
2. **Koleksi Pengujian API Postman (`docs/Postman_Collection_RBIA_Master_Service.json`)**: Test suite API terstruktur untuk verifikasi permintaan HTTP dan respons.
3. **Checklist Keamanan API OWASP (`docs/security/ASC-001_API_Security_Checklist.docx`)**: Dokumen evaluasi kepatuhan API Gateway Kong terhadap OWASP API Security Top 10.
4. **Dokumen Pengujian Penerimaan Pengguna (`docs/AuditSphere_UAT_Meeting_Ready_dengan_Test_Case.docx`)**: Skenario kasus uji UAT bisnis dan kriteria penerimaan pengguna.
5. **Kerangka Kerja Pra-Penetration Testing**: Format baku Letter of Authorization (LoA) serta definisi Scope & Testing Window untuk pengujian penetrasi.
6. **Test Suite Pengujian Unit & Integrasi**: Kumpulan berkas pengujian backend (`*_test.go`) dan frontend (`frontend/vitest.config.ts`).
7. **Panduan Desain & Pengkodean Aman (`docs/security/SCG-001` & `DRC-001`)**: Pedoman penulisan kode aman dan checklist tinjauan desain arsitektur.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif kritis yang wajib diselesaikan sebelum audit sertifikasi ISO/IEC 27001:2022 Stage 2:

1. **Berkas Konfigurasi Workflow Pemindaian Keamanan CI/CD (`.github/workflows/security-scan.yml`)**: Pipeline terotomasi untuk SAST (`gosec`), DAST (OWASP ZAP), dan SCA (`govulncheck`/`npm audit`).
2. **Laporan Resmi Penetration Testing Pihak Ketiga Independen**: Laporan hasil pengujian penetrasi komprehensif yang diterbitkan oleh auditor keamanan bersertifikasi (CEH/OSCP/CREST).
3. **Matriks Pelacakan Kerentanan & Tiket Defect Remediasi**: Logbook tiket penanganan kerentanan lengkap dengan status verifikasi penutupan (*closed/re-tested*).
4. **Lembar Tanda Tangan Penerimaan Keamanan Rilis (*Security Acceptance Sign-Off Sheet*)**: Formulir persetujuan rilis bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership).
5. **SOP Prosedur Pengujian Keamanan dalam Siklus Pengembangan (SOP-SEC-012)**: Standar operasional resmi yang mengatur gerbang pengujian keamanan sebelum rilis.
6. **Laporan Pengujian Penetrasi Khusus Isolasi Multi-Tenant**: Bukti uji empiris ketiadaan kebocoran data antar penyewa (*cross-tenant isolation test report*).

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber Repositori | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.29-01** | Rencana Pengujian Keamanan & Panduan Uji API | Markdown & Word | `docs/API_Testing_Guide.md` & `docs/security/ASC-001_API_Security_Checklist.docx` | Lead InfoSec & Reza |
| **EV-A8.29-02** | Dokumen UAT & Kriteria Penerimaan Pengguna | Dokumen Word Resmi | `docs/AuditSphere_UAT_Meeting_Ready_dengan_Test_Case.docx` | QA Lead & Dimas |
| **EV-A8.29-03** | Dokumen Otorisasi Pengujian Penetrasi (LoA & Scope) | Dokumen PDF Bertanda Tangan | `docs/security/LoA-Penetration-Testing-Scope-Window.pdf` | Lead InfoSec & Reza |
| **EV-A8.29-04** | Pipeline Pemindaian SAST, DAST, & SCA CI/CD | File YAML GitHub Actions | `.github/workflows/security-scan.yml` & Laporan Artefak CI/CD | DevSecOps Engineer |
| **EV-A8.29-05** | Laporan Hasil Penetration Testing & Tiket Remediasi | Laporan PDF & Tracker | `docs/evidence/Independent-Pentest-Report-and-Defect-Tickets.pdf` | Third-Party Pentester / Reza |
| **EV-A8.29-06** | Berita Acara Security Acceptance Sign-Off Rilis | Formulir PDF Bertanda Tangan | `docs/evidence/Security-Acceptance-Sign-Off-Sheet.pdf` | Lead InfoSec & Dimas |

---

## 7. Rencana Aksi Akselerasi & Remediasi (Action Plan)

Rencana aksi terstruktur berprioritas **KRITIS** guna mengubah status kontrol dari **Belum Siap** menjadi **Siap** sebelum batas target **27 November 2026**:

| No | Tindakan Akselerasi / Remediasi | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pengujian Keamanan (`SOP-SEC-012`)** | Dokumen SOP formal mencakup gerbang pengujian SAST, DAST, UAT, dan sign-off yang disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Implementasi Automated SAST pada Pipeline GitHub Actions** | Penambahan step `gosec` (Go) dan `semgrep` (TypeScript/Nuxt) pada `.github/workflows/security-scan.yml` | DevSecOps | 1 minggu |
| 3 | **Implementasi Automated SCA & Dependency Vulnerability Scan** | Penambahan step `govulncheck` dan `npm audit` dengan pemutusan build (*build-breaker*) bila ditemukan celah Kritis | DevSecOps | 1 minggu |
| 4 | **Penyelesaian Finalisasi LoA & Jadwal Penetration Testing** | Penandatanganan Letter of Authorization (LoA) dengan vendor penetration testing independen pihak ketiga | Reza / Dimas | 1 minggu |
| 5 | **Pelaksanaan Pengujian Penetrasi Independen (Black-box & Gray-box)** | Eksekusi pentest komprehensif pada API Gateway Kong, microservices, web app, dan isolasi tenant | Third-Party / Reza | 2 minggu |
| 6 | **Remediasi Celah Temuan Pentest & Penerbitan Tiket Defect** | Patching seluruh temuan tingkat Kritis dan Tinggi serta re-testing oleh penguji independen | Lead Backend / Reza | 2 minggu |
| 7 | **Pemberlakuan Formulir Security Acceptance Sign-Off** | Penerbitan lembar persetujuan formal rilis keamanan per rilis versi mayor AuditSphere | QA Lead / Dimas | 1 minggu |
| 8 | **Pengujian Ulang Regresi & Simulasi Cross-Tenant Security** | Verifikasi akhir bahwa tidak ada potensi eskalasi izin atau kebocoran data antar-tenant | SecOps Team | 1 minggu |

Setelah tindakan 1, 2, 3, 4, dan 5 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-012 bertanda tangan, Workflow SAST/DAST YAML, Laporan Pentest Independen, dan Lembar Security Sign-Off pada Master Audit Evidence Index.

