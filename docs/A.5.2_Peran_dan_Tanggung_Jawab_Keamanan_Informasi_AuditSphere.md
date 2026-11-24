# A.5.2 Peran dan Tanggung Jawab Keamanan Informasi

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Peran dan tanggung jawab keamanan informasi |
| **Fokus pemeriksaan** | Peran, kewenangan, dan akuntabilitas keamanan ditetapkan |
| **Contoh bukti objektif** | RACI, job description, surat penunjukan, struktur organisasi |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Peran dan tanggung jawab keamanan informasi adalah fondasi tata kelola ISMS dan prasyarat pembagian wewenang yang aman; kontrol wajib dipenuhi. |
| **Status implementasi** | **Dalam proses** | Matriks RACI teknis (Bab 14 ISO-AS-001) dan pemisahan peran RBAC di aplikasi sudah aktif; job description spesifik keamanan dan SK penunjukan formal belum diterbitkan. |
| **Ketersediaan bukti** | **Ada sebagian** | Matriks RACI di dokumen keselarasan ISO, peta pembagian fitur tim engineering, dan skema seeder RBAC sistem tersedia; SK penunjukan formal bertanda tangan dan bagan struktur organisasi ISMS belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A5.2-01 s/d EV-A5.2-05 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Penegakan kontrol akses berbasis peran (RBAC) pada API Gateway dan backend berjalan efektif; namun evaluasi periodik alokasi peran dan pemisahan tugas belum dilakukan terjadwal. |
| **Desain kontrol** | **Perlu perbaikan** | Desain otorisasi dan pemisahan peran di tingkat sistem sangat rinci; formalisasi tata kelola SDM (SK penunjukan formal, pembentukan komite pengarah keamanan) memerlukan perbaikan. |
| **Status kesiapan** | **Belum siap** | Penerapan tata kelola organisasi dan kelengkapan bukti dokumenter SDM belum memenuhi seluruh gate kesiapan audit. |
| **Prioritas** | **Tinggi** | Fondasi penegakan akuntabilitas untuk seluruh kontrol Annex A lainnya, serta prasyarat kontrol A.5.3 (Pemisahan tugas) dan A.5.4 (Tanggung jawab manajemen). |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah pengesahan SK penunjukan dan job description formal, atau paling lambat saat terjadi perubahan struktur organisasi platform. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.2 **berlaku** untuk AuditSphere karena:

1. **Sensitivitas Data Audit & ERM**: AuditSphere memproses data audit internal rahasia, penilaian risiko (*Enterprise Risk Management*), kertas kerja audit (KKA), temuan kepatuhan, serta kredensial otentikasi klien enterprise. Setiap personel yang merancang, mengoperasikan, mengelola server, atau mengakses data wajib memiliki peran, batas kewenangan, dan akuntabilitas keamanan yang terdokumentasi secara tegas.
2. **Keterikatan Framework ISO/IEC 27001:2022**: Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) secara eksplisit menetapkan klausul A.5.1 s/d A.5.8 sebagai kelompok kontrol kebijakan dan tata kelola keamanan informasi yang mutlak berlaku bagi operasional platform.
3. **Mandat Standar Tata Kelola Klausul 5.3**: Standar ISO/IEC 27001:2022 Klausul 5.3 (*Organizational roles, responsibilities and authorities*) mewajibkan pimpinan puncak mengalokasikan tanggung jawab dan wewenang untuk memastikan SMKI memenuhi persyaratan standar serta melaporkan kinerja SMKI secara berkala.
4. **Kepatuhan Hukum & Regulasi Data**: Kepatuhan terhadap UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi (UU PDP) dan standar audit internal (IIA Global Internal Audit Standards) mewajibkan adanya penunjukan personil yang bertanggung jawab atas pelindungan data pribadi (Pejabat Pelindungan Data / *Data Protection Officer* / *Security Officer*).
5. **Kebutuhan Due Diligence Klien Enterprise**: Calon klien enterprise (Chief Audit Executive, Head of Information Security, Komite Audit) pada tahapan *vendor security risk assessment* dan *due diligence* selalu meminta bukti kepemilikan struktur organisasi keamanan yang jelas, uraian tugas personil pengelola, serta alur eskalasi penanganan insiden.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Matriks RACI Keamanan Informasi Organisasi

Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 14 menetapkan matriks RACI (*Responsible, Accountable, Consulted, Informed*) yang membagi akuntabilitas operasional keamanan antara 4 fungsi internal inti serta jalur pelaporan ke pimpinan puncak:

| Aktivitas Keamanan | Dev | Infra / Proxmox | Security | QA | Leadership |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Pengembangan Fitur Baru & SDLC** | **R / A** | C | C | I | I |
| **Konfigurasi Jaringan, Firewall & VLAN** | I | **R / A** | C | I | I |
| **Penilaian Risiko & Threat Modeling** | C | C | **R / A** | I | **A** |
| **QA Remediation & Pengujian Keamanan** | **R** | I | C | **R / A** | I |
| **Deteksi & Respons Insiden Keamanan** | C | **R** | **R / A** | I | **A** |
| **Manajemen Kunci Enkripsi & Secrets** | C | **R** | **R / A** | I | I |
| **Pemeliharaan RBAC & Otorisasi Pengguna** | **R** | C | **R / A** | I | I |
| **Audit Kepatuhan & Tinjauan ISO 27001** | C | C | **R** | **R / A** | **A** |

*Keterangan Matriks RACI:*
* **R (Responsible)**: Pelaksana yang mengerjakan tugas secara langsung.
* **A (Accountable)**: Pihak yang memiliki wewenang penuh dan bertanggung jawab atas hasil akhir.
* **C (Consulted)**: Pihak yang dimintai pendapat atau keahliannya dalam proses pengerjaan.
* **I (Informed)**: Pihak yang selalu mendapatkan informasi pembaruan status pengerjaan.

Pemetaan Personel dan Kepemilikan Modul Teknis (berdasarkan `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md`):
* **Dimas Saifullah Ramadhan**: Leadership Approver (Group Vision), Lead AI/ML Inference Pipeline, Risk & Analytics Service, Audit Planning (PKAT), Executive Reporting, dan Quality Assurance (QAIP).
* **RauMomo (Qiko)**: Lead Arsitektur Keamanan & Gateway (Kong 3.4), Auth & Identity Service (MFA, RBAC, Pakta Integritas), Master Data & Tenant Management, Frontend Core Nuxt 4, dan DevOps/Deployment VPS.
* **Reza**: Lead DevSecOps & PIC Checklist Kesiapan Bukti Annex A untuk kontrol organisasi dan teknis (A.5.2, A.5.13, A.5.17, A.5.21, A.5.25, A.5.29, A.8.18).
* **Sumanadi**: Cloud & Infrastructure Security PIC untuk pemeliharaan kontrol kepatuhan (A.5.11, A.5.15, A.5.19, A.5.23, A.5.27, A.8.32).

### 2.2 Struktur Otorisasi & Peran Pengguna pada Sistem (RBAC)

AuditSphere mengimplementasikan *Role-Based Access Control* (RBAC) berbasis Casbin yang terdaftar pada database seeder `backend/auth-service/pkg/database/seeders/seeder.go` dan tabel `rbac_matrix_features` untuk menjamin pemisahan tugas dan akuntabilitas sistem:

* **ADMIN**: Administrator sistem dengan akses penuh terhadap manajemen pengguna, registrasi peran, konfigurasi tenant, dan infrastruktur; secara sadar dibatasi dari hak manipulasi kertas kerja audit (mencegah intervensi independensi audit / *Segregation of Duties*).
* **AUDITOR**: Internal auditor pelaksana audit lapangan, penyusun KKA (F-01 s/d F-05), sampling data transaksi, dan pencatat temuan sementara; akses dibatasi hanya pada penugasan aktifnya.
* **EXECUTIVE**: Pimpinan puncak (Chief Audit Executive, Direksi, Komite Audit) dengan hak akses penuh ke dasbor analitik makro, profil risiko korporasi, serta persetujuan akhir PKAT dan laporan audit final.
* **DEPARTMENT_HEAD**: Kepala unit kerja/departemen; bertanggung jawab mengawasi profil risiko departemen dan persetujuan komitmen mitigasi risiko.
* **AUDITEE**: Perwakilan auditee dengan hak terbatas hanya untuk menanggapi temuan audit (*management response*) dan mengunggah bukti penyelesaian rekomendasi (*Action Taken Report* / CAPA).
* **VIEWER**: Akun peninjau dengan hak baca-saja (*read-only*) terhadap laporan final dan dasbor analitik tanpa wewenang pengubahan data.

Otorisasi peran divalidasi berlapis: di perimeter luar oleh Kong API Gateway (*reverse proxy*) berbasis JWT claims dan di setiap microservice oleh interceptor Casbin RBAC.

### 2.3 Protokol & Jalur Eskalasi Insiden ke Leadership (Leadership OS AIFL)

Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 11 menetapkan rantai komando penanganan insiden. Insiden berklasifikasi Major atau Critical wajib segera dieskalasikan kepada Leadership (**Dimas, Group Vision**) menggunakan format komunikasi baku Leadership OS AIFL yang mencakup:
1. `INCIDENT_ID` & Ringkasan Kejadian
2. Klasifikasi & Severity Level (Major / Critical)
3. Ruang Lingkup & Estimasi Dampak (Data audit, sistem klien, integritas platform)
4. Analisis Penyebab Utama (RCA Awal)
5. Tindakan Mitigasi yang Telah Dijalankan
6. Keputusan / Otorisasi yang Dibutuhkan dari Leadership (Penghentian layanan darurat, pemulihan DR, pelaporan regulator / notifikasi klien).

### 2.4 Pakta Integritas & Akuntabilitas Personal

Halaman `frontend/pages/auth/confidentiality.vue` mewajibkan seluruh pengguna sistem menyetujui *Confidentiality & Integrity Agreement* sebelum memakai aplikasi. Persetujuan disimpan di tabel `confidentiality_agreements` (`user_id`, `version`, `timestamp`, `ip_address`, `user_agent`) sebagai bukti pengikatan tanggung jawab individu atas keamanan informasi dan kerahasiaan data audit yang diaksesnya.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :---: | :--- |
| **Matriks RACI Keamanan Informasi Organisasi** | **Tersedia** | Tercantum pada `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 14 mencakup Development, Infra, Security, dan QA. |
| **Peta Pembagian Peran Teknis & Kepemilikan Sistem** | **Tersedia** | Terdokumentasi rinci dalam `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md`. |
| **Konfigurasi Matriks RBAC & Skema Database Peran** | **Tersedia** | Seeders dan model `auth-service` mendefinisikan 6 peran sistem beserta izin granular pada tabel `rbac_matrix_features`. |
| **Protokol Eskalasi Insiden ke Pimpinan Puncak** | **Tersedia** | Bab 11 ISO-AS-001 mengatur alur pelaporan insiden Leadership OS AIFL ke Dimas (Group Vision). |
| **Akuntabilitas Personal (Pakta Kerahasiaan)** | **Tersedia** | Antarmuka `confidentiality.vue` dan tabel database `confidentiality_agreements` menyimpan audit trail pengguna. |
| **Surat Keputusan (SK) Penunjukan Tim Keamanan Formal** | **Belum tersedia** | Alokasi PIC masih berupa konsensus internal repositori; belum ada SK penunjukan bertanda tangan pimpinan puncak. |
| **Job Description (Uraian Tugas) Tertulis Formal** | **Belum tersedia** | Tanggung jawab keamanan masih melekat pada dokumentasi teknis, belum ada dokumen uraian jabatan SDM resmi. |
| **Bagan Struktur Organisasi Keamanan Informasi (ISMS Org Chart)** | **Belum tersedia** | Belum ada diagram visual yang menggambarkan hierarki dan independensi fungsi keamanan ke Direksi/Komite Audit. |
| **Piagam Komite Pengarah Keamanan (Steering Committee Charter)** | **Belum tersedia** | Koordinasi lintas fungsi sudah berjalan, namun piagam komite formal dan jadwal rapat berkala belum didokumentasikan. |

---

## 4. Catatan / Gap Terverifikasi

1. **Job Description formal belum terpisah**: Tanggung jawab keamanan saat ini tersebar di dokumen teknis dan keselarasan arsitektur; belum ada dokumen uraian jabatan formal dari fungsi SDM yang menguraikan tanggung jawab keamanan spesifik, batas wewenang darurat, dan kualifikasi minimum personel.
2. **Penunjukan PIC belum didukung instrumen legal internal**: Penunjukan PIC di checklist Annex A (Reza, Qiko, Dimas, Sumanadi) berjalan efektif secara teknis, namun belum diformalkan melalui Surat Keputusan (SK) Penunjukan Tim Keamanan Informasi dari pimpinan puncak yang memuat mandat ISMS resmi.
3. **Belum ada piagam Komite Pengarah Keamanan Informasi (ISMS Steering Committee)**: Walaupun eskalasi insiden telah diatur melalui Leadership OS, belum ada piagam formal pembentukan komite pengarah yang mendefinisikan frekuensi pertemuan berkala, kuorum keputusan, dan evaluasi sasaran keamanan organisasi.
4. **Perangkapan peran teknis (Dual Roles) pengembang dan pengelola keamanan**: Pada fase saat ini, pengembangan aplikasi dan administrasi infrastruktur/keamanan dijalankan oleh personel teknis inti yang sama. Perlu adanya dokumentasi kompensasi kontrol yang tegas (misalnya wajib peer-review pull request, larangan direct commit ke branch main, serta log aktivitas server yang tak terhapus).
5. **Peninjauan berkala alokasi peran belum dijadwalkan secara formal**: Belum ada prosedur operasional baku (SOP) atau catatan berkala untuk mereview kecukupan peran keamanan informasi secara rutin (minimal tahunan) atau saat ada pergantian personil kunci.
6. **Akuntabilitas pihak ketiga dan penyedia infrastruktur belum dibakukan**: Batas tanggung jawab keamanan penyedia VPS hosting (IP 202.10.34.166) dan cloud service provider belum diikat dalam matriks tanggung jawab bersama (*Shared Responsibility Matrix*) yang ditandatangani kedua belah pihak.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :-: | :--- | :--- | :-: | :-: | :--- |
| **1** | Susun dan sahkan Surat Keputusan (SK) Penunjukan Tim Keamanan Informasi & Penanggung Jawab ISMS bertanda tangan Leadership (Dimas). | Leadership & Lead InfoSec | **Tinggi** | `[isi]` | Salinan SK Penunjukan bertanda tangan (PDF) |
| **2** | Susun dokumen Job Description formal untuk seluruh peran kunci keamanan (Lead Information Security, DevSecOps Engineer, Cloud Infrastructure Lead, System Administrator, dan Lead Auditor). | Lead Information Security (Reza) | **Tinggi** | `[isi]` | Dokumen Job Description ISMS resmi |
| **3** | Buat dan sahkan Bagan Struktur Organisasi Keamanan Informasi (ISMS Organizational Chart) yang menampilkan jalur pelaporan langsung ke pimpinan puncak. | Lead Information Security | **Sedang** | `[isi]` | Diagram Struktur Organisasi ISMS |
| **4** | Rumuskan Piagam Komite Pengarah Keamanan Informasi (ISMS Steering Committee Charter) yang memuat tugas, wewenang, komposisi anggota, dan frekuensi rapat. | Leadership (Dimas) | **Sedang** | `[isi]` | Piagam Komite Keamanan bertanda tangan |
| **5** | Perluas Matriks RACI pada `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 14 agar mencakup seluruh domain kontrol ISO/IEC 27001:2022 (A.5 s/d A.8). | Lead Information Security (Reza) | **Sedang** | `[isi]` | Pembaruan ISO-AS-001 Bab 14 |
| **6** | Dokumentasikan kebijakan kompensasi pengendalian atas perangkapan tugas pengembang dan administrator (Dual-Role Mitigation SOP: peer review wajib di GitHub, pemisahan credential vault). | Tim DevSecOps & DevOps (Qiko & Reza) | **Tinggi** | `[isi]` | SOP Mitigasi Perangkapan Peran |
| **7** | Buat mekanisme dan formulir tinjauan berkala tahunan atas kecukupan alokasi peran dan tanggung jawab keamanan informasi. | Lead Information Security | **Sedang** | `[isi]` | Prosedur & Formulir Review Peran Tahunan |
| **8** | Daftarkan usulan Evidence ID untuk kontrol A.5.2 ke dalam Master Audit Evidence Index. | Lead Information Security | **Sedang** | `[isi]` | Entri Master Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| **EV-A5.2-01** | Matriks RACI Keamanan Informasi Organisasi (Bab 14) | `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| **EV-A5.2-02** | Peta Pembagian Peran Teknis & Arsitektur Platform | `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md` | Internal |
| **EV-A5.2-03** | Matriks RBAC, Kebijakan Casbin & Database Seeder | `backend/auth-service/pkg/database/seeders/seeder.go` | Internal |
| **EV-A5.2-04** | Protokol & Format Komunikasi Insiden Leadership OS | `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bab 11 | Internal |
| **EV-A5.2-05** | Salinan SK Penunjukan Tim Keamanan & Job Description (Draf Usulan) | `docs/security/` *(lokasi usulan)* | Restricted |

> *Setelah tindakan 1, 2, dan 3 selesai disahkan, tambahkan Evidence ID untuk salinan SK Penunjukan bertanda tangan, dokumen Job Description resmi, dan Bagan Struktur Organisasi ISMS.*

