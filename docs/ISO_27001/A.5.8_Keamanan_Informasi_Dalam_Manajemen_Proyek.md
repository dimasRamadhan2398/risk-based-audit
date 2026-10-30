# A.5.8 — Keamanan Informasi dalam Manajemen Proyek

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Keamanan informasi dalam manajemen proyek |
| **Fokus pemeriksaan** | Risiko keamanan terintegrasi pada seluruh siklus proyek |
| **Contoh bukti objektif** | Template proyek, risk register, gate review, security sign-off |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 29 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere dikembangkan dalam banyak proyek fitur dan perubahan infrastruktur yang menyentuh data audit klien; kontrol wajib dipenuhi. |
| Implementasi status | **Dalam proses** | Setara *Partially Implemented* pada ISO-AS-001 bagian 7. Templat tinjauan desain (DRC-001, TMT-001, SARC-001) dan praktik dokumen rencana dengan bagian risiko sudah ada; prosedur proyek, gate, dan sign-off formal belum berjalan. |
| Ketersediaan bukti | **Ada sebagian** | Ada templat dan dokumen rencana proyek dengan analisis risiko dan checklist verifikasi; belum ada templat proyek baku, risk register proyek, catatan gate review, maupun sign-off keamanan yang terisi. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.8-01 s/d EV-A5.8-06) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Sebagian efektif** | Proyek berisiko tinggi (hardening backend, rotasi JWT, struktur organisasi) memang diawali kajian keamanan dan diakhiri verifikasi; namun praktik ini tidak konsisten di semua proyek dan tidak ada DRC-001 yang terisi. |
| Desain kontrol | **Perlu perbaikan** | Isi templat memadai (red flag, sign-off, action item log), tetapi tidak terhubung ke siklus proyek: tidak ada titik wajib kapan dipakai, siapa yang menyetujui, dan pipeline CI tidak menegakkan *Security Definition of Done*. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Beberapa proyek yang berdampak keamanan sedang berjalan paralel di `staging` (RBAC write rules, upload hardening/ClamAV, struktur organisasi, tidy deploy) tanpa gate keamanan yang tercatat. |
| Review berikutnya | [isi] | Usulan: 6 bulan setelah prosedur keamanan proyek disetujui, selaras dengan siklus tinjau DRC-001 dan SCG-001 (minimal 6 bulan atau setelah insiden keamanan). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.8 **berlaku** untuk AuditSphere karena:

1. AuditSphere dibangun dan diubah melalui proyek fitur, perubahan skema, integrasi, dan perubahan infrastruktur (contoh: `docs/org-structure-model-plan.md`, `docs/audit-calendar-plan.md`, `docs/deploy-tidy-plan.md`). Setiap proyek dapat mengubah kontrol akses, isolasi tenant, atau penanganan berkas bukti audit klien.
2. Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) memetakan A.5.1 s/d A.5.8 sebagai kelompok kebijakan dan tata kelola dengan status **Partially Implemented**, dan mencatat RSK-01 (kerentanan pada modul dalam fase QA Remediation, tingkat **High (12)**) yang bersumber dari siklus pengembangan.
3. Kajian keamanan internal (`docs/backend-enhancement-plan.md`, 2026-10-15) menemukan masalah yang muncul karena keamanan tidak dinilai sejak awal proyek: endpoint audit, master, dan risk hanya memeriksa validitas JWT tanpa pemeriksaan role, berkas `/uploads/*` tersaji tanpa autentikasi, dan secret JWT ter-commit di repositori.
4. Klien enterprise (lembaga jasa keuangan) pada due diligence akan meminta bukti bahwa perubahan sistem melalui penilaian risiko keamanan dan persetujuan sebelum rilis.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Templat tinjauan keamanan proyek (tertulis)

| Dokumen | Versi | Berlaku sejak | Pemilik | Penyetuju | Review berikutnya | Lokasi |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| DRC-001 Design Review Checklist | 1.0 (Draft) | 4 Okt 2026 | [placeholder di dokumen: Security Lead atau Engineering Manager] | Tidak tercantum | Minimal 6 bulan atau setelah insiden | `docs/ISO_27001/threat_modeling/DRC-001-Design-Review-Checklist.docx` |
| TMT-001 Threat Modeling Template (STRIDE) | 1.0 (Draft) | 4 Okt 2026 | [placeholder di dokumen] | Tidak tercantum | Setiap rilis besar, tenant/integrasi baru, insiden, atau minimal 6 bulan | `docs/ISO_27001/threat_modeling/TMT-001_Threat_Modeling_Template.docx` |
| SARC-001 Security Architecture Review Checklist | 1.0 (Draft) | 4 Okt 2026 | [placeholder di dokumen] | Tidak tercantum | Perubahan arsitektur besar, tenant baru, atau minimal tahunan | `docs/ISO_27001/threat_modeling/SARC-001_Security_Architecture_Review_Checklist.docx` |
| SCG-001 Secure Coding Guide | 1.0 (Draft) | 4 Okt 2026 | [placeholder di dokumen: Security Lead atau Engineering Manager] | Tidak tercantum | Minimal 6 bulan atau setelah insiden | `docs/ISO_27001/threat_modeling/SCG-001-Secure-Coding-Guide.docx` |
| ASC-001 API Security Checklist (OWASP API Top 10 2023) | 1.0 (Draft) | 4 Okt 2026 | [placeholder di dokumen] | Tidak tercantum | Tidak tercantum | `docs/ISO_27001/threat_modeling/ASC-001_API_Security_Checklist.docx` |

Isi yang relevan untuk siklus proyek:

- **DRC-001** wajib diisi sebelum fitur, perubahan skema, integrasi, atau perubahan infrastruktur diimplementasikan. Penulis desain menyalinnya ke tiket fitur; minimal dua peninjau (satu teknis, satu keamanan atau DevOps). Bagian 21 berisi *red flag* yang otomatis memblokir persetujuan (RF-01 endpoint tanpa otorisasi, RF-02 berkas tanpa autentikasi, RF-03 error internal ke klien, dan lainnya). Bagian 22 adalah **sign-off** (Disetujui / Disetujui dengan catatan / Ditolak) dan bagian 23 adalah **Action Item Log** dengan kolom "Wajib sebelum rilis".
- **TMT-001** menyediakan metodologi STRIDE, *threat register* (B.6) dengan skor L × I, status mitigasi, dan aturan bahwa risiko *Accepted* memerlukan persetujuan pemilik risiko serta tanggal tinjau ulang.
- **SARC-001** memuat penetapan tingkat risiko komponen (C.1), tanda tangan hasil tinjauan (C.2), dan rencana tindakan (C.3).
- **SCG-001** bagian 6.6 *Security Definition of Done*: tiket belum *Done* bila DRC-001 belum terisi dan disetujui, endpoint baru belum diuji negatif untuk RBAC/objek/tenant, dan seterusnya.

Seluruh dokumen ini berstatus Draft, pemilik masih placeholder, dan **tidak ditemukan satu pun salinan terisi** di repositori.

### 2.2 Praktik dokumen rencana proyek dengan analisis risiko (bukti praktik)

Proyek yang lebih besar didokumentasikan sebagai rencana di `docs/` sebelum dikerjakan. Beberapa memuat unsur yang diminta A.5.8:

| Dokumen rencana | Unsur keamanan dalam siklus proyek | Status di git |
| :--- | :--- | :--- |
| `docs/backend-enhancement-plan.md` (2026-10-15) | Kajian keamanan read-only sebelum perubahan; temuan berprioritas P0/P1/P2 dengan rujukan `path:line`; bagian konflik dengan batasan, pertanyaan terbuka, dan hand-off. | Ter-commit |
| `docs/security-hardening-changes-backend.md` | Catatan perubahan item A s/d G, "Operator actions required before deploy", dan tabel verifikasi (`go build`, `go vet`, `go test` per service) termasuk yang belum terverifikasi. | Ter-commit; perubahan belum di-deploy |
| `docs/jwt-rotation-runbook.md` | Status "PREPARED ONLY"; tidak boleh dieksekusi tanpa persetujuan eksplisit; aturan perbandingan secret hanya dengan digest. | Ter-commit |
| `docs/org-structure-model-plan.md` (FINAL SPEC rev. 3, 2026-10-28) | Keputusan desain keamanan (D4 validasi server-side, D8 pembatasan role, D9 konfirmasi password/step-up dan log perubahan), bagian 11 **Risks** (9 risiko), bagian 14 **Verification checklist**. | **Belum ter-commit** |
| `docs/auditor-auditee-dashboard-spec.md` | Mengidentifikasi risiko cakupan data (endpoint tanpa scoping per pengguna); `auditsphere-solution-architect` meninjau skema sebelum dibangun; `auditsphere-qa` memiliki RBAC checks (bagian 8). | Ter-commit |
| `docs/execution-progress-backfill-runbook.md`, `docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md` | Pre-check (test lokal), aturan yang tidak boleh dilanggar, dan setiap langkah produksi memerlukan persetujuan eksplisit (label [AGEN] / [KONFIRMASI] / [MANUAL]). | Ter-commit |
| `docs/deploy-tidy-plan.md` (2026-10-27) | Inventaris jalur deploy berisiko (seeder menimpa data produksi, dua repo men-deploy produksi). | **Belum ter-commit** |

Pembagian peran pengembangan juga terdokumentasi di `.claude/agents/` (ter-commit): perancangan skema oleh `auditsphere-solution-architect` sebelum implementasi, implementasi oleh agen backend/frontend, pengujian oleh `auditsphere-qa` yang secara eksplisit tidak memperbaiki kode, dan seluruh perubahan produksi melalui `auditsphere-devops`. Ini mendukung pemisahan tugas, tetapi bukan prosedur proyek yang disetujui manajemen.

Keterbatasan: format rencana berbeda-beda, bagian keamanan dan risiko tidak wajib (contoh: `docs/audit-calendar-plan.md` membahas fungsi dan persetujuan pengguna, tanpa bagian risiko keamanan), dan persetujuan revisi tercatat sebagai teks bebas ("approved by the user"), bukan sign-off bertanggal dengan nama penyetuju.

### 2.3 Kontrol pada tahap rilis dan riwayat perubahan

- `.github/workflows/deploy.yml`: deploy produksi backend hanya manual (`workflow_dispatch`) dengan `ref` eksplisit, `environment: production`, *concurrency* tunggal, dan menjalankan `backend/deploy-prod.sh` (backup `pg_dumpall`, migrate-only tanpa seeder, satu service per langkah). Komentar workflow meminta *required reviewers* dikonfigurasi pada environment GitHub; konfigurasi ini **tidak dapat diverifikasi dari repo**.
- `.github/workflows/route-coverage.yml`: satu-satunya pemeriksaan otomatis pada pull request (cakupan route Kong). Bersifat fungsional, bukan pemeriksaan keamanan.
- `.github/workflows/deploy-dev.yml`: push ke `dev`/`develop` langsung di-deploy via SSH tanpa test.
- Test otomatis ada (138 berkas `*_test.go` di `backend/`, 86 berkas test di `frontend/`), tetapi dijalankan manual atau oleh `auditsphere-qa`, tidak oleh CI.
- Riwayat git menunjukkan cabang fitur per proyek keamanan (`feat/rbac-write-rules`, `feat/upload-hardening-clamav`, `feat/viewer-readonly`) dan 24 merge melalui pull request, tetapi juga merge langsung antar cabang. Tidak ada PR template atau `CODEOWNERS` di `.github/`.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Prosedur keamanan informasi dalam manajemen proyek (disetujui) | **Belum tersedia** | Tidak ditemukan di repo. ISO-AS-001 bagian 14 (RACI) hanya menyatakan "Pengembangan Fitur Baru": Development R/A, Security C. |
| Template proyek dengan bagian keamanan wajib | **Belum tersedia** | Tidak ada templat rencana proyek baku. DRC-001 dan TMT-001 adalah templat tinjauan keamanan, bukan templat proyek, dan belum terhubung ke templat proyek. |
| Templat tinjauan desain dan threat modeling (DRC-001, TMT-001, SARC-001, SCG-001, ASC-001) | Tersedia | Format `.docx`, versi 1.0, status Draft, pemilik placeholder. |
| Risk register keamanan per proyek | **Belum tersedia** | Risiko proyek tersebar di dokumen rencana (contoh: `org-structure-model-plan.md` bagian 11). Risk register ISO-AS-001 bagian 9.2 bersifat tingkat sistem (RSK-01 s/d RSK-05) dan tidak dirujuk dari proyek. Threat register TMT-001 B.6 hanya baseline platform. |
| Catatan gate review (inisiasi, desain, sebelum rilis, penutupan) | **Belum tersedia** | Tidak ada gate yang didefinisikan. Yang ada: "confirm before build" pada spec dan persetujuan eksplisit sebelum langkah produksi pada runbook, keduanya tidak terdokumentasi sebagai keputusan gate. |
| Security sign-off bertanggal | **Belum tersedia** | Bagian 22 DRC-001 dan C.2 SARC-001 kosong; tidak ada salinan terisi. |
| Bukti verifikasi keamanan sebelum rilis | Sebagian | Tabel verifikasi di `security-hardening-changes-backend.md` dan checklist di `org-structure-model-plan.md` bagian 14; tidak ada hasil CI otomatis. |
| Gerbang rilis produksi | Sebagian | Deploy manual dengan ref eksplisit dan backup (`deploy.yml`); *required reviewers* pada environment `production` belum terverifikasi; tidak ada test atau scan keamanan di jalur deploy. |
| Riwayat versi dokumen rencana proyek | Sebagian | Sebagian besar rencana ter-commit; `org-structure-model-plan.md`, `deploy-tidy-plan.md`, dan `backup-b2-setup.md` belum ter-commit sehingga belum punya riwayat versi. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada templat proyek dengan bagian keamanan wajib.** Rencana proyek di `docs/` ditulis bebas. Sebagian memuat risiko dan verifikasi, sebagian tidak. Tidak ada aturan kapan sebuah perubahan dianggap "proyek" yang wajib melalui tinjauan keamanan.
2. **Tidak ada risk register keamanan proyek.** Risiko dicatat di dalam masing-masing rencana tanpa ID, pemilik, skor, atau status yang dipantau. Risiko yang sudah diketahui, misalnya akun default `admin` / `password123` pada tenant baru (`org-structure-model-plan.md` bagian 11 butir 9) dan item A12 CI gate di `backend-enhancement-plan.md`, tidak dilacak di satu register dan tidak terhubung ke RSK-01 di ISO-AS-001.
3. **Gate review belum didefinisikan dan DRC-001 belum pernah dipakai.** DRC-001 mensyaratkan tinjauan sebelum implementasi, tetapi tidak ada salinan terisi untuk proyek yang sudah atau sedang berjalan (hardening backend, struktur organisasi, upload hardening, RBAC write rules). Beberapa *red flag* DRC-001 ditandai "ada di repo" per 2026-10-04.
4. **Security sign-off tidak terdokumentasi.** Persetujuan diberikan secara lisan atau di percakapan ("approved by the user", "explicit go-ahead") tanpa nama penyetuju, peran, tanggal, dan keputusan. Peran Security Lead / pemilik dokumen pada DRC-001, TMT-001, SARC-001, SCG-001 masih placeholder.
5. **Security Definition of Done tidak ditegakkan di pipeline.** SCG-001 menyatakan kegagalan scanner tingkat High memblokir merge, tetapi `.github/workflows/` tidak menjalankan `go test`, `go vet`, `govulncheck`, atau secret scanning. `deploy-dev.yml` men-deploy setiap push ke `dev` tanpa test. Item A12 di `backend-enhancement-plan.md` masih terbuka.
6. **Perubahan keamanan belum ditutup secara proyek.** `security-hardening-changes-backend.md` mencatat perubahan "in the working tree, nothing committed or deployed"; tidak ada catatan penutupan proyek atau *lessons learned* yang memastikan risiko residual diterima atau dimitigasi.
7. **Sebagian dokumen rencana belum ter-commit.** Tanpa riwayat git, dokumen tersebut belum dapat dijadikan bukti versi.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Susun dan setujui prosedur keamanan informasi dalam manajemen proyek: kriteria perubahan yang dianggap proyek, tahapan (inisiasi, desain, implementasi, rilis, penutupan), gate dan penyetuju per tahap, serta kaitan ke DRC-001, TMT-001, SARC-001, SCG-001. | Lead Information Security | Tinggi | [isi] | Prosedur bertanda tangan |
| 2 | Buat templat rencana proyek di `docs/` dengan bagian wajib: klasifikasi data terdampak, analisis risiko keamanan, keputusan desain keamanan, rencana verifikasi, rencana rollback, dan sign-off. Jadikan format `org-structure-model-plan.md` (bagian 9, 11, 14) sebagai acuan. | Project Lead | Tinggi | [isi] | Templat proyek |
| 3 | Tetapkan pemilik dan penyetuju DRC-001, TMT-001, SARC-001, SCG-001, ASC-001; ubah status dari Draft setelah disetujui. | Lead Information Security | Tinggi | [isi] | Dokumen versi ≥1.0 dengan document control lengkap |
| 4 | Bentuk risk register keamanan proyek (ID, proyek, deskripsi, L × I, pemilik, mitigasi, status, tanggal tinjau) memakai skala ISO-AS-001 bagian 9.1, dan tautkan ke RSK-01. Isi awal dari risiko yang sudah tercatat di rencana yang ada. | Security Officer | Tinggi | [isi] | Risk register proyek |
| 5 | Isi DRC-001 (dan TMT-001 bila relevan) secara retrospektif untuk proyek yang sedang berjalan: struktur organisasi, RBAC write rules, upload hardening/ClamAV, tidy deploy. Catat sign-off di bagian 22 dan action item di bagian 23. | Project Lead bersama Lead Information Security | Tinggi | [isi] | DRC-001 terisi dan ditandatangani |
| 6 | Tambahkan gate CI di jalur deploy dan pull request: `go vet`, `go test`, `govulncheck`, secret scanning, typecheck/test frontend (item A12 `backend-enhancement-plan.md`). | DevOps Lead | Tinggi | [isi] | Workflow CI dan log run |
| 7 | Verifikasi dan dokumentasikan *required reviewers* pada environment GitHub `production`; tambahkan PR template berisi checklist *Security Definition of Done* SCG-001 bagian 6.6. | DevOps Lead | Sedang | [isi] | Tangkapan layar konfigurasi, PR template |
| 8 | Catat persetujuan gate dan sign-off secara tertulis (nama, peran, tanggal, keputusan) di dokumen rencana atau tiket, bukan hanya di percakapan. | Project Lead | Sedang | [isi] | Entri sign-off pada rencana/tiket |
| 9 | Commit dokumen rencana yang belum ter-commit (`org-structure-model-plan.md`, `deploy-tidy-plan.md`, `backup-b2-setup.md`) agar memiliki riwayat versi. | Project Lead | Sedang | [isi] | Riwayat git |
| 10 | Tambahkan langkah penutupan proyek: verifikasi risiko residual, status action item "Wajib sebelum rilis", dan *lessons learned*. | QA Lead | Sedang | [isi] | Catatan penutupan proyek |
| 11 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.8-01 | DRC-001 Design Review Checklist v1.0 (templat, termasuk red flag dan sign-off) | `docs/ISO_27001/threat_modeling/DRC-001-Design-Review-Checklist.docx` | Internal |
| EV-A5.8-02 | TMT-001 Threat Modeling Template v1.0 (threat register baseline) | `docs/ISO_27001/threat_modeling/TMT-001_Threat_Modeling_Template.docx` | Internal |
| EV-A5.8-03 | SCG-001 Secure Coding Guide v1.0 (bagian 6.6 Security Definition of Done) | `docs/ISO_27001/threat_modeling/SCG-001-Secure-Coding-Guide.docx` | Internal |
| EV-A5.8-04 | Contoh rencana proyek dengan analisis risiko dan verifikasi | `docs/backend-enhancement-plan.md`, `docs/security-hardening-changes-backend.md`, `docs/org-structure-model-plan.md` | Internal |
| EV-A5.8-05 | Gerbang rilis produksi (workflow manual dengan ref eksplisit dan environment `production`) | `.github/workflows/deploy.yml`, `backend/deploy-prod.sh` | Internal |
| EV-A5.8-06 | Risk register sistem (RSK-01 s/d RSK-05) dan RACI pengembangan | `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 9.2 dan 14 | Internal |

Setelah tindakan 1, 2, 4, dan 5 selesai, tambahkan ID untuk prosedur proyek, templat proyek, risk register proyek, dan DRC-001 terisi bertanda tangan.
