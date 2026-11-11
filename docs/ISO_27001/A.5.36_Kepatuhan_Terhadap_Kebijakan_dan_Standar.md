# A.5.36 — Kepatuhan Terhadap Kebijakan dan Standar

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Kepatuhan terhadap kebijakan dan standar |
| **Fokus pemeriksaan** | Kepatuhan internal dipantau dan penyimpangan ditangani |
| **Contoh bukti objektif** | Compliance review, exception register, monitoring report |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Legal & Compliance) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 30 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Hal yang berada di luar repo (konfigurasi GitHub, kondisi server VPS, kontrak, SDM) ditandai **Perlu verifikasi**. Seluruh isian pada Lampiran B (exception register) dan Lampiran C (monitoring report) adalah **draf awal hasil rekonstruksi dari repo**; belum ada persetujuan pemilik risiko dan belum merupakan register resmi.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere memiliki banyak kebijakan dan standar yang harus dipatuhi (CSP-001, AIP-001, ISO-AS-001, A.5.x lainnya, UU PDP, kontrak klien, lisensi). Tanpa pemantauan kepatuhan, penyimpangan tidak terdeteksi dan tidak tercatat. |
| Implementasi status | **Belum diimplementasikan** | Tidak ditemukan prosedur compliance review, exception register, maupun monitoring report. Ada kontrol teknis yang bersifat mencegah dan sebagian pencatat (deploy manual dengan ref eksplisit, backup dan restore-test, `route-coverage.yml`, log perubahan struktur organisasi), tetapi tidak ada yang memantau kepatuhan terhadap kebijakan secara terstruktur. Draf prosedur dan templat diusulkan pada Lampiran A sampai E. |
| Ketersediaan bukti | **Ada sebagian** | Tersedia bahan baku: dokumen kebijakan (`.docx` v1.0 Draft), riwayat git, workflow CI, skrip backup/restore-test, rencana yang mencatat penyimpangan (`backend-enhancement-plan.md`, `deploy-tidy-plan.md`). Compliance review terisi, exception register resmi, dan monitoring report belum ada. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.36-01 s/d EV-A5.36-08) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum efektif** | Penyimpangan yang diketahui (seeder pada compose produksi, port terbuka, tanpa SBOM, CI tanpa test) dicatat tersebar di berbagai dokumen rencana tanpa pemilik, tenggat, persetujuan pengecualian, atau pemantauan status. Tidak ada review berkala yang tercatat. |
| Desain kontrol | **Belum dirancang** | Draf siklus (review → pengecualian → monitoring → tindakan korektif → verifikasi → penutupan) dan RACI diusulkan pada Lampiran A, D, dan E; belum disetujui. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Klien lembaga jasa keuangan akan meminta bukti bahwa kebijakan benar-benar dipatuhi, bukan hanya ditulis. Banyak annex lain (A.5.8, A.5.12, A.5.16, A.5.20, A.5.24, A.5.28, A.5.32) menghasilkan gap yang memerlukan satu tempat pemantauan. |
| Review berikutnya | [isi] | Usulan: compliance review triwulanan; evaluasi desain kontrol ini 12 bulan setelah prosedur disetujui; sinkron dengan audit internal ISMS tahunan (ISO-AS-001 bagian 13.2). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.36 **berlaku** untuk AuditSphere karena:

1. AuditSphere menyimpan kertas kerja (KKA), laporan (LHA), tindak lanjut (ATR), dan kredensial klien enterprise. Kebijakan keamanan yang tidak dipatuhi (misalnya seeder menimpa data produksi, atau port layanan terbuka ke internet) langsung berdampak ke data klien.
2. Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001, v0.1 DRAFT) menetapkan bahwa audit internal ISMS dilakukan minimal sekali setahun (bagian 13.2), review akses triwulanan (bagian 8.3), dan kepatuhan UU PDP (bagian 13.3). Tidak ditemukan catatan bahwa salah satu dari ketiganya pernah dilaksanakan.
3. Annex yang sudah disusun (A.5.1, A.5.8, A.5.12, A.5.16, A.5.20, A.5.24, A.5.28, A.5.32) masing-masing menghasilkan daftar gap dan tindakan perbaikan. Tanpa kontrol A.5.36, gap tersebut tidak punya mekanisme pemantauan terpusat, dan penerimaan risiko (pengecualian) tidak tercatat.
4. Risk register ISO-AS-001 bagian 9.2 mencatat RSK-01 (kerentanan pada modul dalam fase QA Remediation, **High (12)**) dan RSK-03 (kegagalan pemulihan database, **Medium (8)**, status *In Progress*, dengan target "validasi restore mingguan, Q4 2026"). Kemajuan keduanya perlu dipantau, bukan hanya dicatat.
5. Praktik pengembangan melibatkan beberapa kontributor dan dua repositori yang dapat men-deploy produksi (lihat `docs/deploy-tidy-plan.md`). Kepatuhan terhadap aturan deploy aman harus dibuktikan, bukan diasumsikan.
6. Kewajiban eksternal: UU No. 27/2022 tentang Pelindungan Data Pribadi (rujukan di ISO-AS-001 bagian 13.3 dan `A.5.32_Lampiran_Template_Perjanjian_IP.md`), kewajiban kontrak dengan klien, serta syarat lisensi pihak ketiga (A.5.32). Kewajiban regulator sektor keuangan klien (misalnya OJK) **perlu verifikasi** oleh Legal & Compliance; tidak ditemukan pemetaannya di repo.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Kebijakan dan standar yang menjadi objek pemantauan (rujukan)

| Kode | Kebijakan/standar | Status di repo | Annex terkait | Lokasi |
| :--- | :--- | :--- | :--- | :--- |
| CSP-001 | Cloud Security Policy v1.0 | `.docx`, penyetuju tercantum sebagai jabatan, tanpa tanggal tanda tangan | A.5.1, A.5.20 | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` |
| AIP-001 | AI Usage Policy v1.0 | idem | A.5.1, A.5.32 | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx` |
| ISO-AS-001 | Dokumen Keselarasan ISO/IEC 27001:2022 v0.1 DRAFT | Draft; tanggal berlaku dan review placeholder | A.5.1 | `docs/DOKUMEN_KESELARASAN_ISO27001.md` |
| DRC-001, TMT-001, SARC-001, SCG-001, ASC-001 | Templat tinjauan desain, threat modeling, dan pedoman secure coding v1.0 | Draft; pemilik placeholder; tidak ada salinan terisi | A.5.8 | `docs/ISO_27001/threat_modeling/` |
| FRP-001 | Forensic Readiness Procedure | Belum disahkan | A.5.24, A.5.28 | `docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx` |
| (draf) | Kebijakan IP dan penggunaan perangkat lunak | Draf di A.5.32 Lampiran A; belum disetujui | A.5.32 | `docs/ISO_27001/A.5.32_Hak_Kekayaan_Intelektual.md` |
| (draf) | Klasifikasi informasi, IAM, perjanjian pemasok, insiden, bukti forensik | Masing-masing diusulkan di annex | A.5.12, A.5.16, A.5.20, A.5.24, A.5.28 | `docs/ISO_27001/A.5.*.md` |
| Runbook | Deploy produksi aman, rotasi JWT, backup/restore | Ada, berstatus plan atau DRAFT | A.5.8 | `backend/deploy-prod.sh`, `docs/jwt-rotation-runbook.md`, `docs/backup-b2-setup.md`, `docs/deploy-tidy-plan.md` |
| Eksternal | ISO/IEC 27001:2022, UU No. 27/2022 (UU PDP), kontrak klien, lisensi pihak ketiga | Kontrak klien di luar repo (**Perlu verifikasi**) | A.5.31, A.5.32 | ISO-AS-001 bagian 3.1 dan 13.3 |

### 2.2 Kontrol teknis yang mencegah atau mencatat penyimpangan (bukti praktik)

| Kontrol | Fungsi terhadap kepatuhan | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Deploy produksi manual dengan `ref` eksplisit, `environment: production`, concurrency tunggal | Menegakkan aturan "tanpa seeder di produksi", backup sebelum perubahan, rollback per service | `.github/workflows/deploy.yml`, `backend/deploy-prod.sh` | *Required reviewers* pada environment GitHub **tidak dapat diverifikasi dari repo**. Jalur lain (skrip akar `deploy-prod.sh`, `deploy.sh`, `backend/deploy-backend.sh`) masih ada dan tidak memakai override tanpa seeder (`docs/deploy-tidy-plan.md`). |
| Pin host key SSH untuk deploy (`VPS_KNOWN_HOSTS`), tanpa `ssh-keyscan` | Mencegah deploy ke host palsu | `.github/workflows/deploy.yml` (langkah *Install SSH key and pinned host key*) | `deploy-dev.yml` memakai `appleboy/ssh-action` dengan mekanisme berbeda. |
| `route-coverage.yml` | Gerbang otomatis pada PR/push `staging`/`main`: setiap path API yang dipanggil frontend harus punya route Kong | `.github/workflows/route-coverage.yml`, `scripts/check-kong-routes.sh` | Fungsional, bukan pemeriksaan kepatuhan keamanan. Satu-satunya pemeriksaan otomatis pada PR. |
| Backup malam hari dengan checksum SHA-256 dan retensi lokal 14 hari | Mendukung kepatuhan ketersediaan/pemulihan | `backend/scripts/backup/backup-prod.sh` | Tidak ditemukan bukti cron terpasang. Salinan off-host (Backblaze B2) berstatus DRAFT, "not installed on the VPS" (`docs/backup-b2-setup.md`). |
| Restore-test ke kontainer Postgres sementara (memeriksa checksum, restore tanpa error, jumlah baris, ekstraksi uploads) | Bukti bahwa backup dapat dipulihkan; menjawab RSK-03 | `backend/scripts/backup/restore-test.sh` | Skrip ada; tidak ditemukan **log hasil eksekusi** tersimpan di repo. |
| Step-up password untuk aksi sensitif | Konfirmasi ulang identitas sebelum perubahan struktur organisasi | `POST /auth/verify-password` (`backend/auth-service/routes/registry.go`, `backend/auth-service/models/step_up.go`), komit `bf84bff`, `ab220e1` | Baru mencakup perubahan level organisasi (**cakupan lain perlu verifikasi**). |
| Log perubahan khusus | Jejak perubahan terbatas | `backend/*/models/org_change_log.go`, `risk_approval_log.go`, `sys_alert_log.go` | Tidak ditemukan model *audit log* umum (siapa mengubah apa pada kertas kerja, ekspor laporan, data master) di `backend/*/models`; klaim "built-in audit logging" di ISO-AS-001 bagian 13.1 **perlu verifikasi** terhadap kode. |
| RBAC | Kepatuhan terhadap prinsip hak akses minimum | `backend/auth-service/pkg/database/seeders/seeder.go`; penegakan VIEWER read-only di server (komit `b960dd0`, 28 Okt 2026) | Per `docs/backend-enhancement-plan.md` (15 Okt 2026), audit/master/risk/analytics hanya memeriksa JWT valid; izin seeded `manage_*` tidak ditegakkan. Cakupan penegakan terkini selain VIEWER **perlu verifikasi**. |
| Hardening JWT: startup menolak secret kosong, di bawah 64 byte, atau yang dikenal publik | Menegakkan kebijakan secret | `backend/*/pkg/config/jwt_secret.go` (+ test); `docs/security-hardening-changes-backend.md` | Berkas `config.yaml` di `backend/auth-service/pkg/config/` masih berisi nilai non-kosong pada cabang ini (nilai tidak dibaca); rotasi di produksi berstatus "PREPARED ONLY" (`docs/jwt-rotation-runbook.md`). **Perlu verifikasi** status rotasi di server. |
| Pemindaian unggahan ClamAV | Kepatuhan pada pedoman upload | `backend/docker-compose.prod.yml` (image dipin digest sha256) | `docs/deploy-tidy-plan.md` menyebut ClamAV belum ada di compose produksi pada checkout yang dibaca; di compose cabang ini ada. Status di VPS **perlu verifikasi**. |
| Test otomatis | Bukti verifikasi mutu | 165 berkas `*_test.go` di `backend/`, 86 berkas test di `frontend/` (dihitung 30 Okt 2026) | Tidak dijalankan oleh CI (lihat bagian 4, butir 4). |
| Monitoring stack | Pemantauan operasional | `monitoring/docker-compose.monitoring.yml` (Prometheus, Grafana, Loki, Promtail), aturan alert di `monitoring/prometheus-fieldwork-alerts.yml` dan `prometheus-dashboard-alerts.yml` | Seluruh alert adalah alert performa/ketersediaan (rate limit, latensi, service down, Postgres down). **Tidak ada** alert atau dasbor untuk kepatuhan (login gagal, perubahan peran, deploy tak sah). Apakah stack berjalan di produksi **perlu verifikasi**. |
| Riwayat git dan PR | Jejak perubahan kode | Repo; 107 merge commit; cabang fitur per proyek | Memakai juga komit langsung ke `staging` (praktik tim: tanpa PR, `push-direct-to-staging`); tidak ada `CODEOWNERS` atau PR template di `.github/`. |
| Ketentuan terkait agen dan peran | Pemisahan tugas | `.claude/agents/` | Bukan prosedur yang disetujui manajemen (lihat A.5.8). |

### 2.3 Jadwal pemantauan yang sudah dinyatakan di dokumen (belum berjalan)

| Sumber | Ketentuan | Status |
| :--- | :--- | :--- |
| ISO-AS-001 bagian 13.2 | Audit internal ISMS minimal 1x/tahun atau saat perubahan arsitektur mayor, oleh auditor internal independen | Tidak ada laporan, jadwal, atau rencana audit di repo |
| ISO-AS-001 bagian 8.3 | Review akses triwulanan; penonaktifan otomatis akun tidak aktif > 60 hari | Belum berjalan (lihat A.5.16) |
| ISO-AS-001 bagian 9.3 | Uji restore mingguan di staging, target Q4 2026 | Skrip ada; jadwal dan hasil tidak tercatat |
| CSP-001, AIP-001 | Review dokumen paling lambat 28 September 2027 | Belum jatuh tempo |
| DRC-001, TMT-001, SARC-001, SCG-001 | Review 6 bulan atau setelah insiden | Belum ada pemilik |

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Prosedur compliance review berkala (disetujui) | **Belum tersedia** | Draf di Lampiran A. |
| Compliance review terisi (minimal satu siklus) | **Belum tersedia** | Belum pernah dilakukan. Lampiran B dan C adalah baseline awal, bukan hasil review terdokumentasi dengan penandatangan. |
| Exception register | **Belum tersedia** (draf awal) | Lampiran B memuat templat dan 12 entri awal dari fakta repo; belum ada persetujuan pemilik risiko. |
| Monitoring report | **Belum tersedia** | Templat dan KPI di Lampiran C; laporan pertama (baseline 30 Okt 2026) belum ditandatangani. |
| Daftar kebijakan dan standar yang dipantau, dengan pemilik dan tanggal review | Sebagian | Bagian 2.1; banyak dokumen berstatus Draft dengan pemilik placeholder. |
| Catatan penanganan penyimpangan (temuan → tindakan korektif → verifikasi → penutupan) | **Belum tersedia** | Tindakan perbaikan tersebar di setiap annex; belum ada pelacak tunggal. |
| Bukti operasional deploy (log deploy, SHA, hasil backup, hasil restore-test) | Sebagian | Skrip menulis log ke `/root/auditsphere-deploy/<timestamp>/` di VPS (**Perlu verifikasi**, di luar repo). |
| Hasil pemindaian otomatis (test, vet, SAST, dependensi, lisensi) | **Belum tersedia** | Tidak ada di CI. |
| Notulen/ tanda tangan manajemen atas hasil review | **Belum tersedia** | Persetujuan saat ini tercatat sebagai teks bebas pada dokumen rencana (lihat A.5.8). |
| Audit internal ISMS tahunan | **Belum tersedia** | Lihat bagian 2.3. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada proses compliance review.** Tidak ada prosedur, jadwal, templat, atau catatan review kepatuhan terhadap kebijakan di `docs/`. Jadwal yang dinyatakan di ISO-AS-001 (audit internal tahunan, review akses triwulanan, uji restore mingguan) tidak memiliki bukti pelaksanaan.
2. **Tidak ada exception register.** Pengecualian yang nyata (lihat Lampiran B, misalnya EX-01 s/d EX-12) diterima secara diam-diam atau dicatat sebagai teks dalam rencana tanpa pemilik, alasan bisnis, kontrol kompensasi, atau tanggal kedaluwarsa.
3. **Kebijakan yang dipantau sendiri belum disetujui.** CSP-001 dan AIP-001 tidak memiliki tanggal tanda tangan; ISO-AS-001 masih DRAFT v0.1; DRC-001, TMT-001, SARC-001, SCG-001 berstatus Draft tanpa pemilik. Kepatuhan terhadap dokumen yang belum berlaku tidak dapat diukur secara formal (A.5.1, A.5.8).
4. **Kepatuhan tidak ditegakkan di pipeline.** `.github/workflows/` hanya berisi `deploy.yml`, `deploy-dev.yml`, `route-coverage.yml`. Tidak ada `go vet`, `go test`, `govulncheck`, secret scanning, pemindaian dependensi/lisensi, SBOM, maupun typecheck frontend. `deploy-dev.yml` men-deploy setiap push ke `dev`/`develop` tanpa test. SCG-001 bagian 6.6 menyatakan kegagalan scanner tingkat High memblokir merge, sehingga ada selisih antara kebijakan dan praktik. Item A12 di `docs/backend-enhancement-plan.md` masih terbuka.
5. **Jalur deploy yang melanggar aturan "tanpa seeder" masih ada.** `backend/docker-compose.prod.yml` menjalankan `auth seed`, `audit seed`, `master seed` pada setiap start (baris 106, 180, 205). Jalur aman hanya `backend/deploy-prod.sh` (override tanpa seeder). Skrip akar `deploy-prod.sh`/`deploy.sh` -> `backend/deploy-backend.sh` memakai `up -d --build` tanpa override. Repo kedua yang juga men-deploy produksi tercatat di memori tim dan `docs/deploy-tidy-plan.md` (tidak diverifikasi di sini). Aturan deploy aman dipatuhi hanya jika pelaku memilih jalur yang benar.
6. **Port layanan dipublikasikan ke semua antarmuka.** `backend/docker-compose.prod.yml` memetakan `8001`, `8002`, `8003`, `8004`, `8000`, `8084`, serta Kong `8080`, `8443`, `8009` (admin API), `8444` tanpa binding `127.0.0.1`. Pembatasan firewall/nginx di host **Perlu verifikasi** (kewenangan `auditsphere-platform`). Pada `docs/backend-enhancement-plan.md` ini dicatat sebagai blocker.
7. **Penyimpangan keamanan yang diketahui belum punya status pemantauan.** Dari `docs/backend-enhancement-plan.md` (15 Okt 2026) dan `docs/security-hardening-changes-backend.md`: secret JWT yang pernah ter-commit (rotasi produksi belum dieksekusi per runbook "PREPARED ONLY"), endpoint non-auth tanpa pemeriksaan role (sebagian ditangani: VIEWER read-only), `/uploads/*` publik, akun seeder dengan password default sama (juga dicatat di A.5.16 dan A.5.8), tidak ada penguncian akun. Status terkini masing-masing **perlu verifikasi** terhadap kode dan server.
8. **Tidak ada bukti hasil restore-test dan backup off-host.** Skrip `backup-prod.sh` dan `restore-test.sh` ada, tetapi `docs/backup-b2-setup.md` berstatus DRAFT dan tidak ada log hasil. RSK-03 (Medium 8) bergantung pada ini.
9. **Pemantauan operasional belum mencakup kepatuhan.** Alert Prometheus hanya untuk performa dan ketersediaan. Tidak ada metrik atau tampilan Grafana/Loki untuk indikator kepatuhan (percobaan login gagal, perubahan peran, deploy di luar jalur aman, umur dokumen kebijakan).
10. **Pencatatan jejak audit aplikasi tidak terverifikasi.** ISO-AS-001 bagian 13.1 menyatakan audit trail bawaan; di `backend/*/models` hanya ditemukan log khusus (`org_change_log`, `risk_approval_log`, `sys_alert_log`). Kepatuhan terhadap IIA IPPF 1220 dan permintaan klien atas jejak audit perlu dibuktikan sebelum diklaim.
11. **Kewajiban eksternal belum dipetakan.** Tidak ada daftar peraturan/kontrak yang berlaku (UU PDP, ketentuan sektor keuangan klien, kewajiban notifikasi, kontrak lisensi) beserta pemilik dan bukti pemenuhan. Klausul notifikasi insiden CSP-001 butir 20 (1x24 jam / 3x24 jam) tidak dikaitkan dengan prosedur nyata (A.5.24).
12. **Gap annex lain belum dikonsolidasikan.** Gap A.5.1 (versi acceptance tidak divalidasi server), A.5.8 (tanpa gate/sign-off), A.5.12 (klasifikasi tidak diterapkan), A.5.16 (sesi tidak dicabut), A.5.20 (tanpa DPA), A.5.24 (tanpa IR plan), A.5.28 (tanpa evidence log), A.5.32 (temuan GPL, source-available, OpenAI) tersebar di dokumen masing-masing. Ringkasannya dimasukkan sebagai entri awal Lampiran B.
13. **Persetujuan pengecualian tidak formal.** Persetujuan dalam praktik tim diberikan di percakapan atau pesan komit; tidak tercatat nama penyetuju, peran, tanggal, dan masa berlaku.
14. **Beberapa dokumen kebijakan/rencana belum ter-commit.** Saat ini, berkas annex di `docs/ISO_27001/A.5.12`, `A.5.16`, `A.5.20`, `A.5.24`, `A.5.28`, `A.5.32` (dan lampiran IP) berstatus untracked di git (per `git status` 30 Okt 2026), sehingga belum punya riwayat versi.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Setujui dan tetapkan **Prosedur Compliance Review dan Penanganan Penyimpangan** berdasarkan Lampiran A, D, dan E (frekuensi, penanggung jawab, checklist, eskalasi). | Lead Information Security dengan Legal & Compliance | Tinggi | [isi] | Prosedur bertanda tangan |
| 2 | Resmikan **exception register** (Lampiran B): tetapkan pemilik tiap entri, alasan bisnis, kontrol kompensasi, tanggal kedaluwarsa, dan penyetuju; simpan terversi (di `docs/` dan Master Audit Evidence Index). | Lead Information Security | Tinggi | [isi] | Exception register terversi dan disetujui |
| 3 | Lakukan **compliance review pertama (baseline)** memakai checklist Lampiran A.3; verifikasi butir **Perlu verifikasi** pada VPS, GitHub, dan kontrak. | Lead Information Security, DevOps Lead, Legal & Compliance | Tinggi | [isi] | Laporan compliance review pertama |
| 4 | Terbitkan **monitoring report** pertama (Lampiran C) dan jadwalkan bulanan (teknis) dan triwulanan (manajemen). | Project Lead | Sedang | [isi] | Monitoring report bertanggal |
| 5 | Tambahkan gerbang kepatuhan di CI: `go vet`, `go test`, `govulncheck`, secret scanning, typecheck/test frontend, pemindaian lisensi/SBOM (selaras A.5.8 tindakan 6 dan A.5.32 tindakan 8-9). Hasil CI menjadi sumber KPI otomatis. | DevOps Lead | Tinggi | [isi] | Workflow CI dan log run |
| 6 | Verifikasi dan dokumentasikan konfigurasi GitHub: *required reviewers* pada environment `production`, branch protection pada `staging`/`main`; tambahkan PR template dan `CODEOWNERS`. | DevOps Lead | Sedang | [isi] | Tangkapan layar/ekspor konfigurasi, PR template |
| 7 | Selesaikan **tidy deploy**: satu jalur deploy produksi, nonaktifkan atau beri peringatan pada skrip yang menjalankan seeder, dan selesaikan pertanyaan terbuka dua-repo (EX-01, EX-10). Lewat `auditsphere-devops`, jalur aman `backend/deploy-prod.sh --ref`. | DevOps Lead | Tinggi | [isi] | PR tidy deploy; EX-01/EX-10 ditutup |
| 8 | Tutup atau formalkan pengecualian jaringan: verifikasi firewall/nginx, bind port ke `127.0.0.1`, batasi admin API Kong (EX-02). | DevOps Lead dan Platform | Tinggi | [isi] | Hasil `ss -tlnp`/pemindaian port dan konfigurasi firewall (tanpa kredensial) |
| 9 | Pasang cron backup, selesaikan setup B2, jalankan `restore-test.sh` terjadwal dan simpan log hasil (EX-07, RSK-03). | DevOps Lead | Tinggi | [isi] | Log restore-test bertanggal |
| 10 | Verifikasi kemampuan **audit trail aplikasi** terhadap ISO-AS-001 bagian 13.1; bila kurang, buat rencana (model log umum untuk perubahan data audit, ekspor, dan login) dan koreksi klaim dokumen (EX-08). | Project Lead dengan Tim Backend | Sedang | [isi] | Hasil verifikasi; rencana atau koreksi dokumen |
| 11 | Selesaikan rotasi JWT di produksi sesuai runbook (hanya dengan persetujuan eksplisit, lewat `auditsphere-devops`), hapus nilai literal dari `config.yaml` bila belum (EX-09). | DevOps Lead dan Security Officer | Tinggi | [isi] | Bukti rotasi (digest, bukan nilai) |
| 12 | Tetapkan pemilik dan penyetuju dokumen kebijakan yang dipantau (CSP-001, AIP-001, ISO-AS-001, DRC-001, TMT-001, SARC-001, SCG-001, ASC-001); simpan salinan bertanda tangan (selaras A.5.1 tindakan 1-3, A.5.8 tindakan 3). | Lead Information Security | Tinggi | [isi] | Dokumen dengan document control lengkap |
| 13 | Susun **register kewajiban hukum, regulasi, dan kontrak** (UU PDP, ketentuan sektor keuangan klien, kewajiban lisensi A.5.32, notifikasi insiden) lengkap dengan pemilik dan bukti pemenuhan. | Legal & Compliance | Tinggi | [isi] | Register kewajiban |
| 14 | Tambahkan indikator kepatuhan ke Grafana/Loki (login gagal, perubahan peran, deploy, umur backup, status restore-test) dan alert yang relevan. Koordinasi dengan A.5.24. | Platform dan Tim Backend | Sedang | [isi] | Dasbor dan aturan alert |
| 15 | Jadwalkan dan laksanakan **audit internal ISMS** pertama oleh pihak independen sesuai ISO-AS-001 bagian 13.2. | Lead Information Security | Sedang | [isi] | Rencana dan laporan audit internal |
| 16 | Commit dokumen annex dan rencana yang belum ter-commit agar memiliki riwayat versi (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |
| 17 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

Catatan pelaksanaan: tindakan yang menyentuh produksi (rotasi JWT, bind port, cron, perubahan compose) dilakukan melalui `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh` dengan `--ref`, tanpa seeder), bukan redeploy sembarang.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.36-01 | Daftar kebijakan/standar yang dipantau (bagian 2.1) dan dokumen kebijakan sumber | `docs/ISO_27001/`, `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| EV-A5.36-02 | Workflow CI/CD dan skrip deploy (bukti kontrol pencegah dan batasannya) | `.github/workflows/deploy.yml`, `deploy-dev.yml`, `route-coverage.yml`, `backend/deploy-prod.sh`, `scripts/check-kong-routes.sh` | Internal |
| EV-A5.36-03 | Skrip backup dan restore-test (dan log hasil setelah tindakan 9) | `backend/scripts/backup/backup-prod.sh`, `restore-test.sh`, `docs/backup-b2-setup.md` | Internal |
| EV-A5.36-04 | Prosedur compliance review dan penanganan penyimpangan (setelah tindakan 1) | [belum ada] | Internal |
| EV-A5.36-05 | Exception register (draf awal Lampiran B; versi resmi setelah tindakan 2) | `docs/ISO_27001/A.5.36_Kepatuhan_Terhadap_Kebijakan_dan_Standar.md` (Lampiran B) | Internal |
| EV-A5.36-06 | Compliance review terisi (setelah tindakan 3) | [belum ada] | Internal |
| EV-A5.36-07 | Monitoring report bulanan/triwulanan (draf baseline Lampiran C; resmi setelah tindakan 4) | [belum ada] | Internal |
| EV-A5.36-08 | Rencana dan catatan penyimpangan sumber (kajian keamanan, perubahan hardening, tidy deploy, runbook JWT) | `docs/backend-enhancement-plan.md`, `docs/security-hardening-changes-backend.md`, `docs/deploy-tidy-plan.md`, `docs/jwt-rotation-runbook.md` | Internal |

Setelah tindakan 1 sampai 4 dan 15 selesai, tambahkan ID untuk prosedur bertanda tangan, review per siklus, laporan audit internal, dan salinan hasil verifikasi korektif.

---

## Lampiran A — Prosedur Compliance Review Berkala (usulan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**.

**A.1 Frekuensi dan lingkup**

| Jenis | Frekuensi | Lingkup | Pelaksana | Peninjau/penyetuju |
| :--- | :--- | :--- | :--- | :--- |
| Pemeriksaan teknis otomatis | Setiap PR dan rilis | Test, vet, kerentanan dependensi, secret, lisensi (setelah tindakan 5) | DevOps Lead (pemilik pipeline) | Lead Information Security |
| Pemeriksaan kepatuhan operasional | Bulanan | Jalur deploy, backup dan restore-test, port terbuka, akun default, status pengecualian | DevOps Lead dan Security Officer | Lead Information Security |
| Compliance review manajemen | Triwulanan | Seluruh daftar bagian 2.1, exception register, KPI, tindakan terbuka | Lead Information Security | Direktur Utama (penyetuju kebijakan), Legal & Compliance |
| Audit lisensi | 6 bulan (A.5.32 Lampiran C) | Dependensi, layanan berbayar, konten | DevOps Lead | Lead Information Security |
| Review akses | Triwulanan (ISO-AS-001 bagian 8.3) | Akun aktif vs peran (A.5.16) | Tim Keamanan Informasi bersama pimpinan tim | Lead Information Security |
| Audit internal ISMS | Tahunan (ISO-AS-001 bagian 13.2) | Seluruh ISMS | Auditor internal independen | Direktur Utama |
| Review berbasis kejadian | Setelah insiden, perubahan arsitektur mayor, tenant baru, atau perubahan regulasi | Kebijakan terdampak | Lead Information Security | Legal & Compliance |

**A.2 Langkah satu siklus compliance review**

1. Kumpulkan masukan: hasil CI, log deploy (`/root/auditsphere-deploy/<timestamp>/deploy.log`, **Perlu verifikasi** di VPS), hasil restore-test, insiden periode berjalan (A.5.24), perubahan kebijakan, dan gap annex.
2. Jalankan checklist A.3, catat hasil: Patuh / Tidak patuh / Tidak berlaku / Belum diverifikasi, dengan rujukan bukti (`path:line`, ID log, tangkapan layar).
3. Setiap "Tidak patuh" dicocokkan dengan exception register: ada pengecualian aktif dan belum kedaluwarsa (Lampiran B), atau menjadi temuan baru (Lampiran D).
4. Perbarui status entri exception dan tindakan korektif yang terbuka; tandai yang lewat tenggat.
5. Hitung KPI (Lampiran C) dan terbitkan monitoring report.
6. Tinjau hasil bersama manajemen; catat keputusan (nama, peran, tanggal) di laporan. Penerimaan risiko memerlukan penyetuju yang memiliki wewenang atas risiko.
7. Simpan laporan dan bukti pendukung, daftarkan Evidence ID.

**A.3 Checklist compliance review (titik awal)**

| No | Pertanyaan | Kebijakan/standar sumber | Cara memeriksa | Status baseline 30 Okt 2026 |
| :--- | :--- | :--- | :--- | :--- |
| 1 | Setiap deploy produksi memakai jalur `deploy.yml` dengan `ref` eksplisit dan tanpa seeder? | A.5.8, `deploy-tidy-plan.md` | Riwayat run Actions; log server; periksa skrip lain tidak dipakai | Belum diverifikasi (jalur lain masih ada: EX-01) |
| 2 | *Required reviewers* aktif pada environment `production`? | A.5.8 | Pengaturan GitHub | Perlu verifikasi |
| 3 | Backup malam berjalan, checksum valid, restore-test berhasil dalam 7 hari terakhir? | ISO-AS-001 9.3, 12.2 | Log backup/restore-test | Tidak ada bukti (EX-07) |
| 4 | Salinan backup off-host ada? | ISO-AS-001 12.2 | Konfigurasi rclone/B2 | DRAFT, belum terpasang |
| 5 | Port yang terbuka ke internet sesuai daftar yang disetujui? | CSP-001, ISO-AS-001 10.3 | Pemindaian port dari luar; konfigurasi firewall | Tidak sesuai di compose (EX-02); firewall perlu verifikasi |
| 6 | Tidak ada akun default aktif dengan password bawaan di produksi? | A.5.16 | Daftar akun (tanpa mencetak password), kebijakan penggantian | Perlu verifikasi (EX-05) |
| 7 | Review akses triwulanan dilaksanakan dan dicatat? | ISO-AS-001 8.3 | Laporan review akses | Belum ada |
| 8 | Secret JWT dirotasi dan tidak ada literal di repo/konfigurasi? | A.5.8, runbook JWT | Digest di server; `git grep` pola secret (tanpa mencetak nilai) | Belum selesai (EX-09) |
| 9 | CI menjalankan test, vet, scan kerentanan dan secret sebelum merge? | SCG-001 6.6, ISO-AS-001 9.3 | Workflow | Tidak (EX-04) |
| 10 | Dependensi baru diperiksa lisensinya, SBOM terbit per rilis? | A.5.32 | Artefak CI | Tidak (EX-03) |
| 11 | Dokumen kebijakan memiliki pemilik, penyetuju bertanggal, dan jadwal review? | A.5.1 | Document control | Belum lengkap (EX-06) |
| 12 | Insiden pada periode ini dicatat dan ditutup? | A.5.24, A.5.28 | Register insiden | Register belum ada |
| 13 | Akses ke data pribadi dan klien sesuai klasifikasi dan hak minimum? | A.5.12, ISO-AS-001 13.3 | Uji RBAC (`auditsphere-qa`), tinjauan model data | Penegakan role sebagian (EX-11) |
| 14 | Kewajiban lisensi, kontrak, dan UU PDP dipenuhi (DPA pemasok, notifikasi insiden, kepemilikan data tenant)? | A.5.20, A.5.32, UU PDP | Register kewajiban (tindakan 13) | Belum ada register |
| 15 | Pengecualian aktif belum kedaluwarsa dan memiliki kontrol kompensasi? | Kontrol ini | Exception register | Register belum resmi |
| 16 | Unggahan melewati pemindaian ClamAV dan tersaji dengan autentikasi? | ASC-001, A.5.8 | Uji unggah; konfigurasi | Perlu verifikasi |
| 17 | Tenant baru dibuat melalui skrip yang menjaga isolasi (env key, migrate verbs, secret per tenant)? | RSK-02, A.5.20 | Pemeriksaan `scripts/onboard-tenant.sh` dan hasil | Perlu verifikasi per tenant |

---

## Lampiran B — Exception Register

### B.1 Templat

| Kolom | Isi |
| :--- | :--- |
| ID | EX-nn |
| Kebijakan/standar dan klausul | Kode dokumen dan butir (contoh: A.5.8, ISO-AS-001 9.3) |
| Deskripsi penyimpangan | Fakta terverifikasi dengan rujukan `path:line` atau bukti; tanpa nilai rahasia |
| Sistem/tenant terdampak | Prod, dev, tenant tertentu |
| Alasan/Justifikasi bisnis | Mengapa tidak dapat dipatuhi sekarang |
| Risiko (L x I, skala ISO-AS-001 9.1) | Skor dan tingkat; tautan ke RSK-nn bila ada |
| Kontrol kompensasi | Apa yang mengurangi risiko selama pengecualian berlaku |
| Pemilik risiko | Jabatan |
| Penyetuju (nama, peran, tanggal) | Wajib sebelum status "Disetujui"; pengecualian risiko High memerlukan Direktur Utama |
| Berlaku sampai (tanggal kedaluwarsa) | Maksimum 6 bulan untuk High, 12 bulan untuk lainnya; perpanjangan = persetujuan baru |
| Tindakan penutupan dan tautan ke tindakan annex | Nomor tindakan |
| Status | Diusulkan / Disetujui / Dalam penanganan / Ditutup / Kedaluwarsa |
| Tanggal verifikasi terakhir | Tanggal |

### B.2 Entri awal (draf, hasil verifikasi repo 30 Oktober 2026)

Semua entri berstatus **Diusulkan** (belum ada penyetuju). Skor risiko di bawah adalah usulan awal berdasarkan skala L x I ISO-AS-001 bagian 9.1 dan **perlu ditetapkan** oleh pemilik risiko. Kolom pemilik adalah usulan.

| ID | Kebijakan/klausul | Penyimpangan terverifikasi | Rujukan bukti | Risiko (usulan) | Kontrol kompensasi yang ada | Pemilik (usulan) | Penutupan | Catatan verifikasi |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| EX-01 | A.5.8; aturan produksi "tanpa seeder" | Compose produksi menjalankan `seed` pada setiap start; skrip deploy akar tidak memakai override tanpa seeder; satu repo lain juga dapat men-deploy produksi. Seeder dapat menimpa data produksi. | `backend/docker-compose.prod.yml:106,180,205`; `docs/deploy-tidy-plan.md`; `backend/deploy-prod.sh` | Tinggi | Jalur resmi `deploy.yml` -> `backend/deploy-prod.sh` (migrate-only, backup `pg_dumpall`, cek jumlah baris) | DevOps Lead | A.5.36 tindakan 7 | Compose terverifikasi; repo kedua **perlu verifikasi** |
| EX-02 | CSP-001; ISO-AS-001 10.3 (kontrol jaringan) | Port layanan dipublikasikan ke semua antarmuka, termasuk admin API Kong (`8009`). | `backend/docker-compose.prod.yml` (`ports` untuk 8001, 8002, 8003, 8004, 8000, 8084, 8080, 8443, 8009, 8444); `docs/backend-enhancement-plan.md` blocker 5 | Tinggi | Kemungkinan pembatasan firewall/nginx di host | DevOps Lead dan Platform | Tindakan 8 | **Perlu verifikasi** firewall dan pemindaian port dari luar |
| EX-03 | A.5.32 (inventaris/lisensi); SCG-001 | Tidak ada SBOM, `LICENSE`/`NOTICE`, atau pemindaian lisensi/kerentanan dependensi di CI; ada dependensi GPL-2.0 pada Data Hub (`fuzzywuzzy`, `python-Levenshtein`, lisensi **perlu verifikasi** dengan `pip-licenses`). | `.github/workflows/`; `data-server/caatt-engine/requirements.txt`; A.5.32 bagian 4 | Sedang | Hanya dijalankan di server sendiri (belum didistribusikan ke klien) | Lead Information Security | A.5.32 tindakan 2, 8, 9 | Terverifikasi (tidak ada SBOM/scan) |
| EX-04 | SCG-001 6.6 (Security Definition of Done); ISO-AS-001 9.3 | CI tidak menjalankan test, vet, govulncheck, secret scan; `deploy-dev.yml` men-deploy tanpa test. 165 berkas test Go dan 86 berkas test frontend ada tetapi dijalankan manual. | `.github/workflows/*.yml`; `docs/backend-enhancement-plan.md` A12 | Sedang | Pengujian oleh `auditsphere-qa` sebelum rilis (praktik, tidak tercatat) | DevOps Lead | A.5.8 tindakan 6; A.5.36 tindakan 5 | Terverifikasi |
| EX-05 | A.5.16; praktik akun | Seeder membuat akun dengan password default yang sama dan dijalankan ulang bila start (create-if-missing); tenant baru memakai akun default. Status akun default di produksi/tenant belum terbukti sudah diganti. | `backend/auth-service/pkg/database/seeders/seeder.go:223-304` (nilai tidak disalin); `docs/org-structure-model-plan.md` bagian 11 butir 9 | Tinggi | MFA TOTP tersedia (RSK-04); password minimum 6 karakter (per kajian 15 Okt) | Security Officer | A.5.16 | Seeder terverifikasi; kondisi akun di server **perlu verifikasi** |
| EX-06 | A.5.1; document control | Kebijakan dan templat berstatus Draft/v0.1, tanpa tanggal persetujuan dan pemilik placeholder; kepatuhan terhadap dokumen belum berlaku formal. | A.5.1 bagian 3-4; A.5.8 bagian 2.1; ISO-AS-001 | Sedang | Isi dokumen dipakai sebagai acuan praktik | Lead Information Security | A.5.36 tindakan 12 | Terverifikasi |
| EX-07 | ISO-AS-001 9.3 dan 12.2 (RSK-03) | Backup malam dan restore-test tidak terbukti berjalan terjadwal; salinan off-host (B2) berstatus DRAFT, cron belum dipasang; belum ada log hasil restore-test. | `backend/scripts/backup/*.sh`; `docs/backup-b2-setup.md` | Sedang (Medium 8 per RSK-03) | `pg_dumpall` otomatis sebelum setiap deploy oleh `deploy-prod.sh`; salinan di host yang sama | DevOps Lead | Tindakan 9 | **Perlu verifikasi** crontab dan folder backup di VPS |
| EX-08 | ISO-AS-001 13.1 (audit trail); IIA IPPF 1220 | Tidak ditemukan model audit log umum di backend; hanya log khusus (`org_change_log`, `risk_approval_log`, `sys_alert_log`). Klaim dokumen belum terbukti. | `backend/*/models` | Sedang | Log khusus; riwayat `created_by` pada sebagian model (**perlu verifikasi**) | Project Lead | Tindakan 10 | Pencarian model terverifikasi; fitur lain **perlu verifikasi** |
| EX-09 | A.5.8; runbook rotasi JWT | Secret JWT pernah ter-commit; rotasi produksi belum dieksekusi (runbook "PREPARED ONLY"); `config.yaml` pada cabang ini masih berisi nilai non-kosong. Startup kini menolak secret kosong/pendek/dikenal publik. | `docs/jwt-rotation-runbook.md`; `docs/security-hardening-changes-backend.md`; `backend/auth-service/pkg/config/jwt_secret.go` | Tinggi | Validasi secret saat startup; jalur deploy mensyaratkan `JWT_SECRET` dari env | DevOps Lead dan Security Officer | Tindakan 11 | Nilai tidak dibaca; status server **perlu verifikasi** |
| EX-10 | A.5.8 (gerbang rilis) | Komit langsung ke `staging` tanpa PR/review wajib; tidak ada PR template/`CODEOWNERS`; *required reviewers* produksi belum terbukti. | Praktik tim; `.github/` | Sedang | Deploy produksi manual dengan `ref` eksplisit | Project Lead | A.5.8 tindakan 7; A.5.36 tindakan 6 | **Perlu verifikasi** pengaturan GitHub |
| EX-11 | A.5.16; RBAC | Penegakan role di server belum menyeluruh: audit/master/risk/analytics sebelumnya hanya memeriksa JWT valid; baru VIEWER read-only yang ditegakkan (komit `b960dd0`). Izin seeded tidak dievaluasi. | `docs/backend-enhancement-plan.md` 2.2; `git log` | Tinggi | Penyembunyian tombol di frontend (`useRbac`); VIEWER read-only di server | Tim Backend | A.5.16; A.5.36 tindakan 3 | Cakupan terbaru **perlu verifikasi** oleh `auditsphere-qa` |
| EX-12 | A.5.8; A.5.28 (persistensi bukti) | Unggahan master (`./uploads/qar-reports`, `./uploads/consulting`) tidak memakai volume sehingga hilang saat kontainer dibuat ulang; sebelumnya redeploy menghapus unggahan. | `docs/deploy-tidy-plan.md` bagian 2; `backend/docker-compose.prod.yml` | Sedang | `deploy-prod.sh` menyalin keluar dan memulihkan `audit_uploads`; tidak mencakup unggahan master | DevOps Lead | Tidy deploy | Terverifikasi di dokumen rencana; kondisi VPS **perlu verifikasi** |

Catatan: entri baru dapat berasal dari temuan annex lain (contoh: IR plan belum ada, A.5.24; DPA pemasok belum ada, A.5.20). Entri yang ditutup tidak dihapus; status diubah dan bukti penutupan dilampirkan.

---

## Lampiran C — Format Monitoring Report dan KPI Kepatuhan

### C.1 Templat monitoring report

| Bagian | Isi |
| :--- | :--- |
| Identitas | Periode, tanggal terbit, penyusun, peninjau, versi, klasifikasi (Internal) |
| 1. Ringkasan eksekutif | Tiga hal terpenting: status kepatuhan, risiko tertinggi, keputusan yang diminta |
| 2. Hasil compliance review | Ringkasan checklist A.3: jumlah Patuh / Tidak patuh / Belum diverifikasi |
| 3. KPI kepatuhan | Tabel C.2 dengan nilai periode ini, periode lalu, dan target |
| 4. Exception register | Pengecualian baru, diperpanjang, ditutup, kedaluwarsa; daftar High |
| 5. Temuan dan tindakan korektif | Temuan baru, status tindakan terbuka, yang lewat tenggat |
| 6. Insiden dan near-miss kepatuhan | Rujukan ke register insiden (A.5.24) |
| 7. Perubahan kebijakan/regulasi | Dokumen baru atau direvisi, perubahan regulasi/kontrak |
| 8. Keputusan manajemen | Persetujuan pengecualian, alokasi sumber daya (nama, peran, tanggal) |
| Lampiran | Bukti (ID Evidence), hasil CI, log restore-test |

### C.2 KPI kepatuhan (usulan; target untuk ditetapkan manajemen)

| No | KPI | Definisi | Sumber data | Frekuensi | Target (usulan) | Baseline 30 Okt 2026 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| K1 | Persentase butir checklist Patuh | Patuh / (Patuh + Tidak patuh), tanpa "Belum diverifikasi" | Checklist A.3 | Triwulanan | >= 90% | Belum dapat dihitung (mayoritas belum diverifikasi) |
| K2 | Butir "Belum diverifikasi" | Jumlah butir tanpa bukti | Checklist A.3 | Triwulanan | 0 | 8 dari 17 butir bertanda perlu verifikasi/tidak ada bukti (hitung ulang saat review) |
| K3 | Pengecualian aktif menurut tingkat | Jumlah per High/Sedang/Rendah | Exception register | Bulanan | High: 0 tanpa persetujuan | 12 diusulkan, 0 disetujui |
| K4 | Pengecualian kedaluwarsa | Jumlah melewati tanggal berlaku tanpa perpanjangan/penutupan | Exception register | Bulanan | 0 | Tidak berlaku (belum ada tanggal) |
| K5 | Tindakan korektif tepat waktu | Selesai sebelum tenggat / jatuh tempo | Pelacak tindakan | Bulanan | >= 85% | Tenggat belum ditetapkan (kolom [isi] di annex) |
| K6 | Umur rata-rata tindakan High terbuka | Hari sejak dibuka | Pelacak tindakan | Bulanan | <= 30 hari | Belum dapat dihitung |
| K7 | Deploy produksi lewat jalur resmi | Deploy via `deploy.yml` / seluruh deploy | Riwayat Actions + log server | Bulanan | 100% | Perlu verifikasi |
| K8 | Merge ke `main`/`staging` dengan CI hijau | Merge dengan seluruh pemeriksaan wajib lulus / seluruh merge | GitHub | Bulanan | 100% | Hanya `route-coverage` yang ada; tes belum di CI |
| K9 | Backup sukses dan restore-test berhasil | Hari dengan backup valid, dan restore-test terakhir (hari sejak) | Log backup/restore-test | Mingguan | Backup harian 100%; restore-test <= 7 hari | Tidak ada log |
| K10 | Kerentanan High/Critical terbuka pada dependensi | Jumlah dari scanner | Scan CI | Bulanan | 0 | Tidak ada scanner |
| K11 | Dokumen kebijakan dengan pemilik dan persetujuan bertanggal | Dokumen bertanda tangan / daftar dokumen bagian 2.1 | Document control | Triwulanan | 100% | 0 dari 8 kebijakan inti terbukti bertanda tangan |
| K12 | Dokumen kebijakan melewati tanggal review | Jumlah | Document control | Triwulanan | 0 | 0 (belum ada yang jatuh tempo; CSP-001/AIP-001: 28 Sep 2027) |
| K13 | Akun tidak aktif / default aktif | Jumlah (tanpa menampilkan kredensial) | Review akses | Triwulanan | 0 default | Perlu verifikasi |
| K14 | Port terbuka di luar daftar yang disetujui | Jumlah | Pemindaian port | Bulanan | 0 | Tidak sesuai di compose (EX-02) |
| K15 | Pemenuhan kewajiban eksternal | Kewajiban dengan bukti / seluruh kewajiban | Register kewajiban | Triwulanan | 100% | Register belum ada |

Catatan: nilai baseline yang menyebut "perlu verifikasi" bukan nilai terukur. Target pada tabel adalah usulan dan bukan hasil pengukuran. Metrik K7-K10 dapat diotomatisasi setelah tindakan 5 dan 14.

### C.3 Distribusi

Laporan bulanan (teknis): Lead Information Security, DevOps Lead, Project Lead. Laporan triwulanan: ditambah Legal & Compliance dan Direktur Utama. Laporan disimpan terversi dan didaftarkan di Master Audit Evidence Index.

---

## Lampiran D — Prosedur Penanganan Penyimpangan (usulan)

**D.1 Alur**

1. **Deteksi.** Sumber: compliance review, hasil CI, audit internal/eksternal, insiden, laporan staf, pemeriksaan klien. Setiap orang dapat melaporkan ke Lead Information Security.
2. **Pencatatan temuan.** Buat entri temuan: ID, sumber, kebijakan/klausul, deskripsi dan bukti, tingkat keparahan (usulan: High / Medium / Low sesuai skala ISO-AS-001 9.1), tanggal ditemukan, pelapor.
3. **Triase (maksimum 5 hari kerja).** Lead Information Security menentukan: (a) bukan penyimpangan; (b) perlu tindakan korektif; (c) perlu pengecualian sementara (Lampiran B); (d) insiden keamanan (A.5.24).
4. **Analisis akar masalah.** Tentukan penyebab (proses, teknis, pengetahuan, sumber daya), apakah berulang, dan sistem lain yang terdampak (termasuk tenant lain).
5. **Tindakan korektif dan pencegahan.** Tetapkan PIC, tenggat, dan bukti yang dihasilkan. Tenggat usulan: High 30 hari, Medium 90 hari, Low 180 hari. Tindakan yang menyentuh produksi melalui `auditsphere-devops` dengan jalur deploy aman.
6. **Pengecualian bila tindakan tidak dapat dilakukan segera.** Isi entri exception register dengan kontrol kompensasi dan tanggal kedaluwarsa; persetujuan oleh pemilik risiko (High: Direktur Utama). Pengecualian tidak menghapus temuan.
7. **Verifikasi.** Pihak yang bukan pelaksana (Security Officer, `auditsphere-qa`, atau auditor internal) memeriksa bukti dan, bila relevan, menguji ulang (contoh: pemindaian port ulang, uji RBAC, restore-test).
8. **Penutupan.** Status "Ditutup" hanya setelah verifikasi lulus; catat tanggal, pemverifikasi, dan Evidence ID. Bila verifikasi gagal, buka kembali dengan tenggat baru.
9. **Pelaporan.** Status masuk monitoring report berikutnya; pelajaran yang diperoleh memperbarui checklist A.3 dan kebijakan terkait.

**D.2 Eskalasi**

| Kondisi | Eskalasi ke | Batas waktu |
| :--- | :--- | :--- |
| Temuan High tanpa PIC atau tindakan | Lead Information Security -> Direktur Utama | 5 hari kerja sejak triase |
| Tindakan lewat tenggat | PIC -> Lead Information Security; > 30 hari lewat tenggat -> Direktur Utama | Pada monitoring report berikutnya, atau segera bila High |
| Pengecualian kedaluwarsa tanpa perpanjangan | Pemilik risiko -> Lead Information Security | Hari kedaluwarsa |
| Penyimpangan yang berdampak pada data klien atau berpotensi insiden | Aktifkan A.5.24 (insiden); kumpulkan bukti sesuai A.5.28; notifikasi sesuai CSP-001 butir 20 dan kontrak (1x24 jam / 3x24 jam, **tenggat resmi perlu verifikasi Legal & Compliance**) | Segera |
| Dugaan pelanggaran lisensi/IP | Legal & Compliance (A.5.32 Lampiran A.7) | 2 hari kerja |
| Pelanggaran disengaja oleh staf/kontraktor | HR dan Legal & Compliance (proses disipliner mengikuti kontrak kerja, **di luar repo**) | Sesuai kebijakan SDM |

**D.3 Templat catatan temuan dan tindakan korektif**

| Kolom | Isi |
| :--- | :--- |
| ID temuan / sumber | NC-nn / compliance review, CI, audit, insiden |
| Kebijakan/klausul | |
| Deskripsi dan bukti | |
| Keparahan | High / Medium / Low |
| Akar masalah | |
| Tindakan korektif (PIC, tenggat) | |
| Tindakan pencegahan | |
| Pengecualian terkait | EX-nn atau tidak ada |
| Verifikator dan tanggal | |
| Status | Baru / Dalam penanganan / Menunggu verifikasi / Ditutup / Dibuka kembali |
| Bukti penutupan (Evidence ID) | |

---

## Lampiran E — Penanggung Jawab (RACI usulan)

| Aktivitas | Direktur Utama | Lead Information Security | Legal & Compliance | DevOps Lead | Security Officer | Project Lead | Tim Backend/Frontend | QA Lead |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Prosedur compliance review dan penanganan penyimpangan | A (persetujuan) | R | C | C | C | C | I | I |
| Pemeriksaan teknis otomatis di CI | I | A | I | R | C | I | C | C |
| Pemeriksaan kepatuhan operasional bulanan | I | A | I | R | R | C | C | I |
| Compliance review triwulanan | C | A/R | R | C | C | C | I | I |
| Exception register (persetujuan) | A (risiko High) | R | C | C | C | C | I | I |
| Monitoring report | I | A | C | C | C | R | I | I |
| Register kewajiban hukum, regulasi, kontrak | I | C | A/R | I | C | I | I | I |
| Tindakan korektif teknis | I | A | I | R | C | C | R | C |
| Verifikasi tindakan korektif | I | A | I | C | R | I | I | R |
| Audit internal ISMS tahunan | A | R | C | C | C | I | I | I |
| Eskalasi insiden kepatuhan (A.5.24) | I | A | R | R | R | C | C | I |

R = Responsible, A = Accountable, C = Consulted, I = Informed.

---

Catatan sinkronisasi checklist: baris A.5.36 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan belum diperbarui oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
