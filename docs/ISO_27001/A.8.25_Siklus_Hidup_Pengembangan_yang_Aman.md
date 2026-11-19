# A.8.25 — Siklus Hidup Pengembangan yang Aman

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Siklus hidup pengembangan yang aman |
| **Fokus pemeriksaan** | Aktivitas keamanan terintegrasi pada seluruh SDLC |
| **Contoh bukti objektif** | Secure SDLC, security gates, threat model, release checklist |
| **Sistem** | AuditSphere (auditsphere.app): repo `risk-based-audit` (cabang `staging`), layanan Go (auth, audit, master, risk, analytics), python-ai, frontend Nuxt, Data Hub, skrip tenant dan deploy, pipeline GitHub Actions, agen pengembangan `.claude/agents/` |
| **PIC** | Lead Information Security (didukung Engineering Lead dan DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut agar aturan pengembangan perangkat lunak dan sistem yang aman **ditetapkan dan diterapkan**: keamanan masuk pada setiap tahap (persyaratan, desain, implementasi, tinjauan, pengujian, rilis, pemeliharaan) melalui gerbang yang dapat dibuktikan, bukan hanya niat baik pengembang. Repo hanya membuktikan **kode, test, workflow, skrip, konfigurasi, dan dokumen**; repo **tidak** membuktikan apa yang berjalan di VPS, pengaturan GitHub (branch protection, required reviewers, daftar akses), hasil run CI di GitHub, atau isi berkas lingkungan. Semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH, GitHub API, atau konsol yang dipakai. Tes **tidak dijalankan** dalam penyusunan dokumen ini: hanya keberadaan dan jumlah berkas test yang dihitung, dan **tidak ada klaim lulus/gagal**. Butir yang bersumber dari memori kerja internal dicantumkan sebagai **[Catatan]** dan **bukan bukti repo**. Berkas `.env`, `claude_code_env.sh`, dan `frontend/.env.production` **tidak dibuka**; hanya status pelacakan git yang diperiksa. Seluruh angka, ambang, frekuensi, dan tanggal pada Lampiran A sampai G dan pada kolom target adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan (pengaturan lokal `.claude/settings.json` yang terlacak memuat alamat host dan jalur kunci; keduanya sengaja tidak dikutip).

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.25 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.25** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Siklus hidup pengembangan yang aman** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Aktivitas keamanan terintegrasi pada seluruh SDLC** | Sesuai baris spreadsheet; tiga gate: aturan tertulis (kebijakan SDLC), gerbang keamanan yang ditegakkan (CI, tinjauan, persetujuan rilis), dan rekaman (threat model, hasil pemindaian, catatan tinjauan). |
| Contoh bukti objektif | **Secure SDLC, security gates, threat model, release checklist** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | AuditSphere dikembangkan sendiri (kode Go, Vue, Python, skrip deploy) dan memproses data audit klien lembaga jasa keuangan multi-tenant; cacat pada kode (otorisasi, isolasi tenant, unggah berkas) langsung menjadi insiden data. Pengecualian tidak dapat dipertahankan (bagian 1). |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: pengembangan internal dan kontributor eksternal, penggunaan agen AI, riwayat rahasia yang pernah masuk repo, CI tanpa gerbang keamanan, dan dokumen panduan keamanan masih Draft tanpa salinan terisi. |
| Status implementasi | **Sebagian** | Ada praktik nyata: ratusan berkas test (213 `*_test.go` dan 125 berkas test frontend), tinjauan keamanan berbasis dokumen dengan catatan perubahan, penguatan JWT/unggah berkas/RBAC dengan test, ClamAV fail-closed, MFA terenkripsi, deploy backend manual terkendali, pemisahan peran agen. Belum ada: kebijakan SDLC disetujui, gerbang CI keamanan, SAST/SCA/secret scan, threat model terisi, persetujuan tinjauan kode yang dapat dibuktikan, uji penetrasi. |
| Ketersediaan bukti | **Sebagian (artefak teknis dan dokumen rencana) / Tidak ada (rekaman gerbang)** | Kode, test, catatan perubahan keamanan, template DRC/TMT/SARC/SCG/ASC (Draft) ada. Hasil CI, tinjauan kode bertanda, threat model terisi, hasil pemindaian, dan laporan uji penetrasi **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.25-01 s/d EV-A8.25-12) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **Engineering Lead** dan **DevOps Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal bila CI gate dan kebijakan SDLC-001 disetujui | Selanjutnya: tinjauan kebijakan [tahunan], tinjauan gerbang dan threat model [per rilis mayor] (Lampiran A, B, G). |
| Desain kontrol | **Perlu perbaikan** | Bahan baku baik (panduan SCG-001, DRC-001, TMT-001, catatan tinjauan keamanan), tetapi belum menjadi siklus: tidak ada titik wajib, penyetuju, atau penegakan di pipeline; alur kerja push langsung ke `staging` tidak memiliki pengganti tinjauan sejawat. |
| Efektivitas operasional | **Belum teruji** | Tidak ada rekaman gerbang, hasil CI, atau tinjauan per rilis. Tidak dapat diklaim bahwa setiap perubahan yang sampai ke produksi telah melewati pemeriksaan keamanan. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: CI hanya satu pemeriksaan fungsional (cakupan rute Kong); tanpa test/vet/lint/SCA/secret scan/SAST di pipeline; tanpa bukti branch protection, `CODEOWNERS`, PR template; panduan keamanan Draft dan tidak terisi; tanpa threat model dan uji penetrasi; rahasia pernah dikomit dan sebagian berkas sensitif/berlebih masih terlacak; endpoint proxy OpenAI tanpa autentikasi dihapus hanya di working tree; nilai bawaan literal; penggunaan AI dan kontributor tanpa aturan tertulis. |
| Tindakan perbaikan | **Lihat bagian 5 (15 tindakan)** | Draf Lampiran A sampai G sebagai masukan. |
| Prioritas | **Tinggi** | Pengembangan berjalan cepat (sekitar 150 commit sejak 1 Oktober 2026 oleh empat identitas penulis) tanpa gerbang otomatis; klien lembaga keuangan akan menanyakan SDLC aman dan uji penetrasi. |
| Target | **[usulan] 31 Desember 2026** untuk kebijakan disetujui, verifikasi pengaturan GitHub, dan gerbang CI minimum; **[usulan] 28 Februari 2027** untuk threat model, SAST, dan uji keamanan independen | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "aturan tertulis" dan "rekaman" belum terpenuhi; gate "ditegakkan" hanya sebagian (test dan penguatan ada, pipeline tidak menegakkan). |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: kebijakan SDLC-001 disetujui, pengaturan GitHub terverifikasi, CI gate minimum berjalan satu siklus rilis, threat model v0.1 terisi. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 6 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.25 **berlaku** untuk AuditSphere karena:

1. **Perangkat lunak dikembangkan sendiri dan menjadi inti layanan.** Lima layanan Go, layanan python-ai, frontend Nuxt, Data Hub, serta skrip deploy dan penyediaan tenant ditulis dan dipelihara oleh tim. Cacat keamanan pada kode ini (otorisasi, isolasi antar-tenant, unggah dan unduh berkas, penanganan token) langsung berdampak pada kerahasiaan data audit klien (A.5.12: *Confidential/Restricted*).
2. **Pengembangan cepat dan melibatkan banyak pihak.** Riwayat git (555 commit, dihitung 19 November 2026) memuat enam identitas penulis, termasuk fork `upstream` milik kontributor, satu bot agen AI (`google-labs-jules[bot]`), dan sekitar 50 commit dengan trailer `Co-Authored-By` model Claude (hitungan perkiraan dari baris trailer). Sejak 1 Oktober 2026 terdapat sekitar 150 commit dari empat identitas. Tanpa aturan dan gerbang, kualitas keamanan bergantung pada kebiasaan masing-masing penulis.
3. **Sudah ada insiden desain yang membuktikan perlunya gerbang.** Rencana `docs/backend-enhancement-plan.md` (15 Oktober 2026) menemukan, antara lain, rahasia JWT yang dikomit, registrasi publik, berkas unggahan yang dapat dibaca tanpa autentikasi, otorisasi yang belum ditegakkan pada rute tulis, dan seeder yang berjalan pada setiap start produksi. Temuan ini lolos karena tidak ada gerbang keamanan pada alur pengembangan; sebagian besar kini ditutup di kode (bagian 2.2), tetapi tidak ada mekanisme yang mencegah kemunculan kembali.
4. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim menanyakan SDLC aman, pemindaian kerentanan, tinjauan kode, dan uji penetrasi. `docs/DOKUMEN_KESELARASAN_ISO27001.md` memuat baris "A.8.25 s/d A.8.34" berstatus *In Progress* yang menyebut peninjauan kode, panduan pengkodean aman, dan SAST/DAST; **SAST/DAST dan peninjauan kode terdokumentasi tidak ditemukan di repo** (bagian 4), sehingga klaim itu tidak boleh dipakai dalam materi eksternal.
5. **Multi-tenant dengan isolasi per stack.** Kesalahan kecil pada generator tenant, kunci per tenant, atau middleware tenant berdampak lintas klien. Kontrol ini terkait langsung dengan A.8.26 (persyaratan keamanan aplikasi), A.8.27 (arsitektur sistem aman), A.8.28 (pengkodean aman), A.8.29 (pengujian keamanan), A.8.31, A.8.32, dan A.8.4 (akses ke kode sumber), semuanya draf.
6. **Penggunaan AI dalam pengembangan.** Agen AI membantu menulis kode (`.claude/agents/`, `JULES_PROMPT.md`, trailer commit). Tanpa aturan, kode hasil AI dapat masuk tanpa tinjauan manusia, dan permintaan (*prompt*) dapat memuat data sensitif.

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti kode, test, workflow, skrip, atau dokumen di repo** pada 19 November 2026. Apakah itu yang berjalan di VPS, dan apakah pengaturan GitHub menegakkannya, adalah **Perlu verifikasi**. Instalasi dan deploy produksi dibahas di A.8.19 dan tidak diulang di sini.

### 2.1 Alur kerja git, tinjauan, dan CI

| Area | Yang terlihat | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Workflow GitHub | Hanya tiga: `deploy.yml` (manual), `deploy-dev.yml` (push `dev`/`develop` dan manual), `route-coverage.yml` (pull request, push `staging`/`main`, manual). | `.github/workflows/` | Satu-satunya pemeriksaan otomatis adalah `route-coverage.yml` (frontend memanggil path yang tidak ada rute Kong-nya; menjalankan `scripts/check-kong-routes.sh`, izin `contents: read`). Bukan pemeriksaan keamanan. |
| Praktik cabang | Riwayat menunjukkan cabang fitur per proyek keamanan (`feat/viewer-readonly`, `feat/upload-hardening-clamav`, `feat/rbac-write-rules`, `feat/dept-scoped-audit-universe`) dan 29 commit "Merge pull request" (terakhir #31, 28 Oktober 2026). Setelah itu commit dilakukan langsung pada `staging`. | `git log` (19 November 2026); **[Catatan]** memori 29 Oktober 2026: push langsung ke `staging`, tanpa cabang fitur atau PR, karena cabang membingungkan | Tidak ada tinjauan sejawat yang tercatat untuk commit langsung. Efek samping: `route-coverage.yml` hanya berpicu pada push `staging`/`main` bila tidak ada PR. |
| Berkas tata kelola repo | Tidak ada `CODEOWNERS`, PR template, template issue, `dependabot.yml`, `SECURITY.md`, atau `CLAUDE.md`. Tidak ada hook pre-commit (hanya contoh bawaan git) dan tidak ada konfigurasi `golangci-lint`. | `.github/` (hanya `workflows/`); pencarian repo 19 November 2026 | Pengaturan *branch protection* dan *required reviewers* ada di GitHub, bukan di repo: **Perlu verifikasi**. |
| Pembaruan upstream | Merge dari fork kontributor dicatat sebagai commit merge ("Merge upstream/main (dimas): ..."). | `git log`, remote `upstream` | Tidak ada langkah tinjauan diff merge yang tercatat. |
| Konvensi commit | Pesan commit berpola `feat/fix/chore(scope):`; commit hasil bantuan AI memuat trailer `Co-Authored-By`. | `git log` | Konvensi, bukan kebijakan tertulis. |

### 2.2 Praktik keamanan yang nyata di kode (bukti penguatan)

| Praktik | Implementasi terverifikasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Kebijakan rahasia JWT saat start | Layanan menolak start bila `JWT_SECRET` kosong, di bawah 64 byte, atau cocok dengan digest SHA-256 dari nilai yang pernah dikomit (hanya digest disimpan). Berlaku di auth, audit, master, risk, analytics; ada test per layanan. | `backend/*/pkg/config/jwt_secret.go` dan `jwt_secret_test.go`; `backend/{analytics,risk}-service/middleware/jwt_secret*.go`; `backend/auth-service/cmd/auth/serve.go` | Nilai lama tetap ada di riwayat git (bagian 4 butir 6). Penyelesaian rotasi di produksi: **Perlu verifikasi**. |
| Penyematan algoritma JWT | Middleware memakai `WithValidMethods` pada seluruh layanan Go yang memvalidasi token. | `backend/*/middleware/auth.go`, `pkg/middleware/auth.go` | Rahasia HMAC dipakai bersama antar-layanan dalam satu stack (rencana B10 mengusulkan asimetris). |
| Penutupan registrasi publik | Tidak ada rute `/auth/register`; `CreateUser` admin mewajibkan peran. | `backend/auth-service/routes/registry.go`; `routes/auth_register_test.go`; `controllers/user/create_user_roles_test.go` (disebut `docs/security-hardening-changes-backend.md`) | Catatan perubahan itu menyatakan perubahan belum di-deploy saat ditulis; status produksi: **Perlu verifikasi**. |
| Berkas terautentikasi dan terkurung | `/uploads/*` mensyaratkan token dan menolak penyeberangan jalur; tipe konten dari allowlist ekstensi; `nosniff` dan CSP sandbox. | `backend/audit-service/pkg/safefile/`, `routes/uploads.go`, `routes/uploads_test.go`, `controllers/media/controller_test.go` | Rute lain dengan token di *query* (rencana A9): tidak diverifikasi ulang. |
| Pemindaian malware unggahan | Klien ClamAV (INSTREAM), mode `required` bawaan (fail-closed), `optional`, `off` untuk lokal; tersedia di audit-service dan master-service dengan test. | `backend/audit-service/pkg/avscan/`, `backend/master-service/pkg/avscan/` | Mode di tiap stack runtime: **Perlu verifikasi**. |
| RBAC dan Viewer hanya-baca | Middleware Viewer, allowlist peran per rute tulis, `status_guard`; test matriks rute memeriksa **setiap** rute terdaftar. | `backend/audit-service/routes/viewer_readonly_test.go` (`TestViewerRouteMatrix`, `TestViewerEveryRegisteredRoute`; berkas ini ada perubahan belum dikomit), `pkg/middleware/viewer.go`, `status_guard_test.go`; master/risk `roles.go`; frontend `tests/unit/rbacMatrix.test.ts` | Matriks rute menyeluruh baru terlihat pada audit-service; layanan lain belum diverifikasi setara. |
| Step-up dan MFA | Middleware step-up di master-service; MFA dengan penguncian, secret TOTP terenkripsi AES-256-GCM (`enc:v1:`) dengan skrip migrasi dry-run-by-default dan dokumen rollout; test gagal-tutup. | `backend/master-service/pkg/middleware/step_up*.go`; `backend/auth-service/pkg/secretcrypt/`; `repositories/mfa_secret_migration.go`; `services/auth/mfa_secret_fail_closed_test.go`; `docs/mfa-encryption-rollout.md` | Dokumen rollout menyatakan tanpa kunci rahasia disimpan plaintext (keadaan prod/training); **Perlu verifikasi**. Lihat juga `docs` keamanan enkripsi di A.8.24 (draf). |
| Redaksi log dan jejak aktivitas | Middleware `redact` di setiap layanan dengan test; paket `activitylog` (konsumer, retensi, redaksi) dan kontroler log sistem dengan test. | `backend/*/pkg/middleware/redact*.go`; `backend/auth-service/pkg/activitylog/`; `controllers/systemlog/` | Ketahanan sink log (Kafka fire-and-forget, rencana B6): tidak diverifikasi ulang. A.8.15 draf. |
| Isolasi dan penyediaan tenant | Generator membuat rahasia JWT per tenant secara acak dan `.env` tenant diabaikan git; model struktur organisasi fleksibel dikendalikan `ORG_STRUCTURE_MODE` dengan penolakan seed pada tenant lama; test migrasi/seed per mode. | `scripts/generate-tenant-stack.sh`; `.gitignore` (`backend/tenants/*/.env`); `backend/master-service/cmd/seed_org_levels.go`, `cmd/org_structure_pg_test.go`, `pkg/middleware/org_structure.go` | Tidak ditemukan test otomatis lintas-tenant (token tenant A ditolak tenant B) (bagian 4 butir 9). |
| Disiplin migrasi dan seeder | `deploy-prod.sh` menolak command yang memuat `seed`, menjalankan migrate-only; seeder risk dapat dimatikan lewat flag; seeder organisasi menolak berjalan pada tenant lama. | `backend/deploy-prod.sh`; `backend/docker-compose.prod.yml` (komentar risk-service); `cmd/seed_catalog_only_test.go` | Compose produksi di repo masih memuat `seed && serve` pada auth, audit, master (bagian 4 butir 8; lihat A.8.19). |
| Test otomatis | 213 berkas `*_test.go` di `backend/` (auth, audit, master, risk, analytics) dan 125 berkas test frontend (`tests/unit`, `tests/integration`, `tests/e2e`; vitest dan Playwright); ESLint terkonfigurasi (`eslint.config.mjs`). Hasil lulus/gagal **tidak dijalankan** untuk dokumen ini. | `find backend -name '*_test.go'`; `frontend/tests/`; `frontend/package.json` (skrip `test`, `test:unit`, `test:e2e`, `lint`) | Tidak dijalankan oleh CI. Jumlah berkas pada dokumen lain (138, 165) berbeda karena tanggal hitung berbeda. Frontend membawa baseline galat tipe (`frontend/ts_errors.txt`, 66 baris, terlacak git). |
| Pelaksana pengujian terpisah | Agen `auditsphere-qa` menjalankan `go build/vet/test`, vitest, typecheck, validasi locale, dan probe smoke baca-saja; secara eksplisit tidak memperbaiki kode. | `.claude/agents/auditsphere-qa.md` | Dijalankan atas permintaan, tanpa rekaman terarsip. |
| Tinjauan keamanan berbasis dokumen | Kajian baca-saja `docs/backend-enhancement-plan.md` (temuan P0/P1/P2 dengan `path:line`), catatan perubahan `docs/security-hardening-changes-backend.md` (dengan "Operator actions required before deploy" dan tabel verifikasi), `docs/jwt-rotation-runbook.md`, `docs/mfa-encryption-rollout.md`; rencana lain memuat bagian risiko (A.5.8 bagian 2.2). | `docs/` | Format tidak baku; persetujuan tercatat sebagai teks bebas; bukan tinjauan independen. |
| Panduan keamanan perancangan dan pengkodean | SCG-001 Secure Coding Guide (termasuk *Security Definition of Done*), DRC-001 Design Review Checklist (dengan *red flag* pemblokir), TMT-001 Threat Modeling Template (STRIDE), SARC-001, ASC-001 (OWASP API Top 10 2023). Semua `.docx` v1.0 Draft. | `docs/ISO_27001/threat_modeling/` (lima berkas; SCG-001 dibaca pembukanya dan statusnya Draft dengan pemilik placeholder) | Tidak ada salinan terisi dan tidak ada pemilik/penyetuju (A.5.8 bagian 2.1). |
| Pemisahan peran agen pengembangan | Perancangan skema oleh `auditsphere-solution-architect`, implementasi oleh agen backend/frontend, pengujian oleh `auditsphere-qa`, perubahan produksi lewat `auditsphere-devops`; `.claude/agents/` terlacak git. | `.claude/agents/*.md` | Konvensi alat, bukan kebijakan yang disetujui; lihat butir 12 bagian 4. |
| Ketergantungan terkunci | `go.sum` per modul, `package-lock.json` dan `bun.lock`, `requirements.txt`. | berkas terlacak (15 berkas manifes/lock) | Detail pin dan SBOM di A.8.19 bagian 2.3 dan 4; tidak diulang. |
| Deploy backend terkendali | `deploy.yml` manual dengan `ref` eksplisit dan `environment: production`; `deploy-prod.sh` dengan dump pra-deploy, tanpa seeder, rollback otomatis. | `.github/workflows/deploy.yml`; `backend/deploy-prod.sh` | Dibahas penuh di A.8.19; di sini hanya sebagai gate rilis (Lampiran C). |

### 2.3 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `.github/`, `backend/`, `frontend/`, `scripts/`, `data-server/`, `docs/` (19 November 2026) tidak menemukan: langkah `go vet`/`go test`/`golangci-lint`/`govulncheck`, `npm audit`, `pip-audit`, `trivy`, `gosec`, `semgrep`, `CodeQL`, `gitleaks`/`trufflehog`, vitest/typecheck/eslint di workflow; Dependabot/Renovate; `CODEOWNERS` dan PR template; hook pre-commit; kebijakan atau prosedur SDLC aman yang disetujui; salinan terisi DRC-001/TMT-001/SARC-001; threat model AuditSphere; laporan uji penetrasi, DAST, atau tinjauan keamanan independen (`docs/backend-enhancement-plan.md` butir C5 hanya merekomendasikan); catatan tinjauan kode per perubahan; definisi *Security Definition of Done* yang ditegakkan; aturan penggunaan AI dalam pengembangan (selain berkas `.docx` AIP-001 di `docs/ISO_27001/ai-ml_security/`, yang statusnya tidak diperiksa dalam tinjauan ini); aturan data uji (data sintetis vs salinan produksi); persyaratan penandatanganan commit; pedoman onboarding kontributor eksternal. Seluruh keadaan GitHub dan VPS **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kebijakan Secure SDLC (disetujui) | **Belum tersedia; draf di Lampiran A** | Kode usulan SDLC-001; angka [usulan]. |
| Gerbang keamanan yang ditegakkan (security gates) | **Belum tersedia** | Hanya `route-coverage.yml`. Rancangan gerbang G1 s/d G6 ada di Lampiran A bagian 4. |
| Threat model | **Sebagian (templat); isi belum ada** | TMT-001 (Draft) ada; Lampiran B menyediakan kerangka dengan batas kepercayaan dan kandidat ancaman dari repo. |
| Release / security checklist | **Belum tersedia; draf di Lampiran C** | Skrip deploy punya pra-deploy dan verifikasi pasca-deploy teknis; belum ada checklist keamanan rilis bertanda. |
| Catatan tinjauan kode / PR | **Belum tersedia** | Riwayat PR (#27 s/d #31, akhir Oktober) ada di GitHub (akses **Perlu verifikasi**); tidak ada template atau catatan tinjauan. Checklist di Lampiran D. |
| Hasil test (CI atau terarsip) | **Belum tersedia** | Test ada dan dijalankan manual; tidak ada log run terarsip. |
| Hasil pemindaian (SAST, SCA, secret, image) | **Belum tersedia** | Tidak ada pemindaian. |
| Bukti penguatan keamanan beserta test | **Tersedia** | Bagian 2.2; catatan `docs/security-hardening-changes-backend.md`. Status deploy ke produksi: **Perlu verifikasi**. |
| Pengaturan GitHub (branch protection, required reviewers, akses) | **Belum tersedia** | Perlu tangkapan layar/ekspor bertanggal oleh pihak berwenang. |
| Laporan uji penetrasi / DAST | **Belum tersedia** | Tidak ada. |
| Panduan keamanan (SCG/DRC/TMT/SARC/ASC) | **Tersedia sebagai Draft** | v1.0 Draft, pemilik placeholder; belum disetujui. |
| Pelatihan pengkodean aman bagi pengembang | **Belum tersedia** | Lihat A.6.3 (awareness) untuk cakupan umum. |
| Aturan penggunaan AI pada pengembangan dan kontributor eksternal | **Belum tersedia; draf di Lampiran A butir 12 s/d 13** | Lihat A.5.20 dan A.6.3 untuk aspek kontrak dan pelatihan. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari memori kerja internal dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan.

1. **[Repo] CI tidak menegakkan keamanan atau mutu.** Tiga workflow; hanya `route-coverage.yml` yang memeriksa sesuatu, dan itu cakupan rute Kong. `deploy-dev.yml` men-deploy pada setiap push ke `dev`/`develop` tanpa test. `deploy.yml` mengkompilasi binary tanpa menjalankan test atau vet. Test (213 + 125 berkas) hanya bernilai bila dijalankan; saat ini bergantung pada kebiasaan pengembang atau `auditsphere-qa`. Gerbang CI sudah direkomendasikan di A.5.8 tindakan 6, A.5.36 tindakan 5, dan A.8.19 tindakan 9; dokumen ini menjadikannya satu pekerjaan bersama dan menambahkan SAST dan pemindaian rahasia (bagian 5 tindakan 3 dan 4).
2. **[Tidak ada] Tidak ada SAST, SCA, secret scanning, DAST, atau uji penetrasi.** Tidak ada hasil pemindaian apa pun. `docs/DOKUMEN_KESELARASAN_ISO27001.md` menyebut peninjauan kode, SAST/DAST untuk A.8.25 sampai A.8.34 berstatus *In Progress*; pencarian repo tidak menemukan satu pun bukti praktik itu. Klaim tersebut perlu dikoreksi atau didukung bukti sebelum dipakai di luar (selaras A.5.36).
3. **[Repo]+[Catatan] Tidak ada gerbang tinjauan sejawat yang dapat dibuktikan.** Tidak ada `CODEOWNERS`, PR template, atau bukti branch protection; commit sejak akhir Oktober langsung ke `staging` (kebijakan kerja pengguna, 29 Oktober 2026, **[Catatan]**) sehingga "tinjauan kode" tidak meninggalkan jejak. Satu penulis menghasilkan sebagian besar commit terbaru (94 dari sekitar 150 sejak 1 Oktober), dan perubahan di jalur sensitif (otorisasi, upload, deploy) tidak memiliki penyetuju kedua. Preferensi menghindari cabang fitur dapat dipertahankan, tetapi perlu pengganti: gerbang pada saat **promosi ke deploy** (CI hijau, tinjauan keamanan tercatat untuk jalur sensitif, persetujuan penyetuju bukan pemicu) dan tinjauan pasca-merge terjadwal [usulan] (Lampiran A butir 6 dan Lampiran D). Pengaturan GitHub: **Perlu verifikasi**.
4. **[Repo] Panduan keamanan ada tetapi belum menjadi proses.** SCG-001, DRC-001, TMT-001, SARC-001, ASC-001 berstatus Draft, pemilik/peninjau placeholder, tanpa salinan terisi. DRC-001 menyatakan *red flag* memblokir persetujuan dan SCG-001 6.6 mendefinisikan *Security Definition of Done*, tetapi tidak ada template tiket/PR atau workflow yang memaksanya (A.5.8 bagian 4 butir 4 dan 5). Tidak ada threat model AuditSphere (batas kepercayaan, aset, ancaman lintas-tenant, ancaman unggah berkas, python-ai, Data Hub).
5. **[Repo] Hasil tinjauan keamanan tidak diverifikasi ulang secara berkala.** Rencana dan catatan perubahan (`backend-enhancement-plan.md`, `security-hardening-changes-backend.md`) tidak memiliki ID temuan yang dilacak sampai tertutup. Beberapa butir kini terlihat tertutup di kode (JWT, registrasi, `/uploads`, ClamAV), tetapi butir lain belum diverifikasi ulang pada tinjauan ini: rate limit login, kebijakan kata sandi, ketahanan token/revokasi, DTO eksplisit untuk CRUD generik, urutan cache terhadap autentikasi, TLS Postgres/Redis auth, rahasia bawaan. Tanpa register temuan, tidak ada bukti penutupan.
6. **[Repo] Rahasia pernah dikomit dan sisa jejaknya belum dibersihkan.** Kode memuat digest dari dua nilai JWT yang pernah dikomit (dipakai untuk menolak nilai itu saat start); catatan perubahan menyebut nilai lama masih ada di riwayat git dan rotasi harus dilakukan (runbook berstatus *PREPARED ONLY* saat ditulis). Apakah rotasi telah dilaksanakan di semua stack, apakah riwayat dibersihkan, dan apakah ada rahasia lain di riwayat tidak dapat dipastikan tanpa pemindaian riwayat (gitleaks atau setara): **Perlu verifikasi**. Pemindaian riwayat belum dilakukan.
7. **[Repo] Berkas terlacak yang berpotensi sensitif atau tidak semestinya.** Diperiksa hanya statusnya, isi tidak dibuka: `frontend/.env.production` terlacak (`.gitignore` hanya mencakup `.env` dan `.env.local`); `claude_code_env.sh` di akar terlacak; `.claude/settings.json` terlacak dan memuat izin perintah yang menyebut jalur kunci SSH dan alamat host produksi serta pola izin luas (mis. `bash -c`, `go test *`); empat bundel model `*.pkl` di `backend/python-ai/models/` terlacak meski `*.pkl` ada di `.gitignore` (kemungkinan ditambahkan paksa); `backend/LHA_020_LHA_01_KS_IAD_2023.docx` dan `backend/Test_LHA_Fixed.docx` terlacak (apakah memuat data klien nyata: **Perlu verifikasi**, lihat A.8.33 draf); serta `backend/inspect.json`, `backend/update.sql`, `scratch/`, `frontend/ts_errors.txt`, `frontend/test_script.js`. Pemeriksaan isi dilakukan oleh pihak berwenang tanpa mencetak nilai.
8. **[Repo] Nilai bawaan literal dan seeder pada jalur bawaan.** Compose produksi memberi nilai bawaan untuk `PYTHON_AI_API_KEY` dan `DATA_HUB_API_KEY` (nilai tidak dikutip) dan menjalankan `seed && serve` pada auth, audit, master; `main.py` python-ai memuat nilai bawaan literal untuk `API_KEY` meskipun komentarnya menyatakan fail-closed (perilaku sebenarnya: **Perlu verifikasi** dengan membaca middleware); seeder auth dan `scripts/onboard-tenant.sh` membuat akun bawaan dengan kata sandi literal (nilai tidak dikutip); catatan A.5.8 butir 2 sudah mencatat risiko ini. Instalasi tanpa override memakai nilai di repo. Selaras A.8.9 (draf), A.8.19 butir 14, dan rencana A6; dokumen ini menempatkannya sebagai pelanggaran prinsip "aman secara bawaan" pada SDLC.
9. **[Repo] Test isolasi lintas-tenant dan matriks RBAC menyeluruh belum ada.** Test per layanan mencakup kebijakan rahasia, Viewer, status guard, struktur organisasi, redaksi; pencarian "tenant" pada berkas test hanya menemukan konfigurasi risk/master dan test organisasi. Tidak ditemukan test yang membuktikan token satu tenant ditolak tenant lain, atau test matriks rute menyeluruh pada master/risk/analytics seperti `TestViewerEveryRegisteredRoute` di audit-service. Catatan memori tentang rahasia JWT identik antar-layanan (19 September 2026) mendahului penguatan JWT; kondisi terkini dibuktikan oleh gerbang start (bagian 2.2), bukan oleh test lintas-tenant.
10. **[Repo] Endpoint proxy OpenAI tanpa autentikasi, dihapus hanya di working tree.** `frontend/server/api/openai.ts` pada HEAD adalah route Nitro yang meneruskan isi `prompt` dari klien ke API OpenAI memakai kunci server dari environment, tanpa pemeriksaan autentikasi terlihat. Berkas itu berstatus `D` (dihapus, belum dikomit); tidak ada pemanggil `/api/openai` di kode frontend dan tidak ada referensi `OPENAI` di `nuxt.config.ts` atau compose; dependensi `openai` masih ada di `frontend/package.json`. Penghapusan sudah benar, tetapi perlu dikomit dan dependensi dibersihkan. Apakah route ini pernah ada pada build produksi (frontend dibangun manual, **[Catatan]**), apakah `OPENAI_API_KEY` pernah diset di server, dan apakah kunci perlu dirotasi: **Perlu verifikasi**; isi environment tidak dibaca.
11. **[Repo] Artefak model dimuat dengan `pickle`.** `backend/python-ai/retrain_engine.py` memuat bundel model dengan `pickle.load`; berkas pickle yang dimodifikasi dapat mengeksekusi kode saat dimuat. Asal dan checksum model tidak dicatat (A.8.19 bagian 4 butir 8), dan `entrypoint.sh` menyalin model dari image ke volume. Ini risiko rantai pasok pengembangan yang perlu kendali integritas.
12. **[Repo]+[Catatan] Penggunaan AI dan kontributor eksternal tanpa aturan tertulis.** Sekitar 50 commit membawa trailer Claude dan 14 commit berasal dari bot agen AI (Juni 2026); `JULES_PROMPT.md` terlacak di akar; `.claude/agents/` mendefinisikan peran tetapi tidak mewajibkan tinjauan manusia atas kode hasil AI, larangan memasukkan data klien/rahasia ke prompt, atau batas izin alat. Fork kontributor (`upstream`) di-merge tanpa tinjauan diff yang tercatat. Tidak ada persyaratan penandatanganan commit atau pedoman onboarding kontributor. Berkas AIP-001 (kebijakan penggunaan AI) ada sebagai Draft `.docx`; kaitannya dengan pengembangan belum diperiksa.
13. **[Repo] Disiplin migrasi, seeder, dan lingkungan hanya sebagian.** Jalur terkendali ada (`deploy-prod.sh` tanpa seeder), tetapi compose produksi di repo masih menjalankan seeder pada start dan jalur deploy lama masih ada (A.8.19 bagian 4 butir 1). Rencana B12 (migrasi sebagai job tersendiri, aditif dahulu) belum diterapkan. Lingkungan dev memakai host produksi dan training adalah klon data produksi: **[Catatan]** memori 5 November 2026 menyebut training berisi salinan DB dan unggahan produksi, serta kunci Resend yang sama sehingga email OTP dapat terkirim ke pengguna nyata. Itu berarti "lingkungan uji" memuat data nyata (A.8.31 dan A.8.33, draf).
14. **[Tidak ada] Tidak ada data uji sintetis terdokumentasi dan tidak ada aturan perlindungan data uji.** Test Go memakai basis data SQLite memori dan nilai sintetis (contoh `viewer_readonly_test.go`), yang baik; tetapi demo/training memakai salinan produksi dan tidak ada aturan tertulis tentang penggunaan data nyata untuk uji.
15. **[Repo] Mutu frontend belum bersih.** `frontend/ts_errors.txt` (66 baris) terlacak dan `auditsphere-qa` menyebut baseline galat tipe yang sudah ada; tanpa gerbang, baseline dapat bertambah tanpa disadari. ESLint ada, belum dijalankan di CI.
16. **[Repo] Frontend dan binary dibangun di luar CI.** Binary Go dikompilasi sebelum image dan frontend dibangun di mesin pengembang (A.8.19 bagian 4 butir 7); hasil build tidak terikat ke hasil test yang lulus. Dari sudut SDLC, tidak ada bukti bahwa artefak yang dirilis adalah artefak yang diuji.
17. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet; tidak ada akses SSH atau GitHub API.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Verifikasi pengaturan GitHub secara read-only** oleh pihak berwenang: branch protection `staging`/`main`, required reviewers lingkungan `production`, siapa memiliki akses tulis (repo `origin` dan fork `upstream`), apakah PR wajib, apakah `deploy-dev.yml` masih berpicu. Tanpa menampilkan nilai secret. | Lead Information Security, DevOps Lead | Tinggi | [usulan] 30 November 2026 | Tangkapan layar/ekspor bertanggal (EV-A8.25-05) |
| 2 | **Setujui kebijakan SDLC-001** (Lampiran A) dengan keputusan eksplisit soal tinjauan sejawat: pertahankan push langsung ke `staging` tetapi pasang gerbang pada promosi ke deploy (CI hijau, tinjauan keamanan tercatat untuk jalur sensitif, penyetuju selain pemicu) dan tinjauan pasca-merge [mingguan]. Keputusan harus ditulis, bukan implisit. | Lead Information Security, Engineering Lead, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | Kebijakan disetujui (EV-A8.25-02) |
| 3 | **Bangun gerbang CI minimum** pada push `staging` dan sebagai prasyarat `deploy.yml`: `go build`, `go vet`, `go test` per layanan, vitest unit, ESLint, typecheck dengan baseline tercatat, `govulncheck`, `npm audit` (atau setara), `pip-audit`. Satu pekerjaan bersama dengan A.5.8 tindakan 6, A.5.36 tindakan 5, dan A.8.19 tindakan 9. Perubahan workflow oleh agen devops atas persetujuan. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Workflow terbaru; log run lulus/gagal (EV-A8.25-06) |
| 4 | **Tambahkan secret scanning dan SAST**: `gitleaks` (atau `trufflehog`) pada push dan pemindaian riwayat penuh sekali; satu alat SAST (mis. `gosec`/`semgrep` untuk Go dan Python, ESLint keamanan untuk Vue/TS) dengan ambang blok [usulan: temuan kritis/tinggi memblokir promosi]. Pemindaian riwayat dilakukan lokal tanpa mencetak nilai. | `auditsphere-devops`, Lead Information Security | Tinggi | [usulan] 31 Januari 2027 | Laporan pemindaian riwayat tersanitasi; konfigurasi CI (EV-A8.25-07) |
| 5 | **Tangani rahasia dan berkas terlacak**: pastikan rotasi JWT selesai di semua stack dan catat; putuskan pembersihan riwayat; periksa isi `frontend/.env.production`, `claude_code_env.sh`, `.claude/settings.json` (hapus alamat host, jalur kunci, dan izin luas), bundel `*.pkl`, dan dua berkas LHA `.docx` oleh pihak berwenang tanpa mencetak nilai; hentikan pelacakan yang tidak perlu dan perluas `.gitignore`. Perubahan kode/konfigurasi atas persetujuan. | Lead Information Security, DevOps Lead | Tinggi | [usulan] 15 Desember 2026 | Daftar berkas ditinjau (tanpa nilai); commit; catatan rotasi (EV-A8.25-08) |
| 6 | **Komit penghapusan `frontend/server/api/openai.ts`** (bersama pembersihan dependensi `openai` dari `package.json` dan lockfile), verifikasi apakah rute itu ada pada build produksi dan apakah `OPENAI_API_KEY` pernah diset di server; bila ya, rotasi kunci dan periksa penagihan. | `auditsphere-frontend`, DevOps Lead | Tinggi | [usulan] 30 November 2026 | Commit; catatan verifikasi tanpa nilai |
| 7 | **Isi threat model v0.1 AuditSphere** memakai TMT-001 dan Lampiran B: batas kepercayaan, aset, ancaman lintas-tenant, unggah berkas, JWT/step-up, python-ai, Data Hub, jalur deploy; daftarkan temuan di register berskor. Libatkan `auditsphere-solution-architect`. | Solution Architect, Lead Information Security | Tinggi | [usulan] 31 Januari 2027 | Threat model terisi dan register (EV-A8.25-04) |
| 8 | **Terapkan DRC-001 dan Security Definition of Done** (SCG-001 6.6) lewat template commit/tiket dan Lampiran D; tentukan "jalur sensitif" (auth, middleware, upload, tenant, deploy, Kong) yang wajib tinjauan keamanan sebelum promosi; tambahkan `CODEOWNERS` untuk jalur itu bila PR dipakai. | Engineering Lead, Lead Information Security | Sedang | [usulan] 31 Januari 2027 | Template; satu contoh DRC-001 terisi (EV-A8.25-09) |
| 9 | **Hapus nilai bawaan literal dan seeder dari jalur bawaan**: kunci python-ai dan Data Hub tanpa default (gagal-tutup), seeder tidak berjalan pada start produksi, akun bawaan wajib ganti/ditiadakan pada prod/tenant; koordinasi dengan A.8.9 dan A.8.19 tindakan 11. | `auditsphere-backend`, `auditsphere-devops` | Tinggi | [usulan] 31 Januari 2027 | Commit; test start tanpa nilai gagal |
| 10 | **Perluas uji keamanan otomatis**: test isolasi lintas-tenant (token tenant A ditolak tenant B pada seluruh layanan), matriks rute RBAC menyeluruh pada master/risk/analytics (meniru `TestViewerEveryRegisteredRoute`), test negatif untuk upload/traversal pada master-service; catat hasil per rilis. | `auditsphere-backend`, `auditsphere-qa` | Sedang | [usulan] 28 Februari 2027 | Test baru dan hasil run (EV-A8.25-10) |
| 11 | **Kendalikan integritas artefak model**: catat asal dan checksum bundel model, verifikasi sebelum `pickle.load`, batasi siapa dapat mengganti model, dan nilai format yang lebih aman. | Data Analyst, Backend Lead | Sedang | [usulan] 28 Februari 2027 | Register model; kode verifikasi |
| 12 | **Tetapkan aturan pengembangan berbantuan AI dan kontributor eksternal**: tinjauan manusia wajib, tanpa data klien/rahasia di prompt, batas izin alat pada `.claude/settings.json`, penanda commit, tinjauan diff pada merge `upstream`, persyaratan akses dan onboarding kontributor; selaraskan dengan AIP-001, A.5.20, A.6.3. | Lead Information Security, Engineering Lead | Sedang | [usulan] 31 Januari 2027 | Aturan disetujui (Lampiran A butir 12 s/d 13) |
| 13 | **Jadwalkan uji keamanan independen**: uji penetrasi eksternal [usulan: sebelum rilis mayor berikutnya dan setidaknya tahunan] dan DAST pada lingkungan sintetis (bukan salinan data produksi). Ruang lingkup mencakup isolasi lintas-tenant. | Lead Information Security | Sedang | [usulan] 28 Februari 2027 | Laporan dan rencana remediasi (EV-A8.25-11) |
| 14 | **Bentuk register temuan keamanan** yang melacak temuan `backend-enhancement-plan.md` dan hasil pemindaian sampai tertutup (ID, tingkat, pemilik, tenggat [usulan: kritis 7 hari, tinggi 30 hari], bukti penutupan). Gabungkan dengan A.5.36. | Lead Information Security | Sedang | [usulan] 31 Januari 2027 | Register terisi |
| 15 | **Daftarkan Evidence ID** pada Master Audit Evidence Index, jadwalkan tinjauan berkala (Lampiran G), dan commit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 1, 3, 4, 5, 6, 9 menyentuh GitHub, workflow, atau konfigurasi dan memerlukan persetujuan eksplisit; dilakukan lewat agen devops/backend/frontend, satu perubahan per sesi. Deploy ke produksi tidak boleh dipicu sebagai bagian dari pengumpulan bukti. Pemindaian riwayat dan pemeriksaan berkas sensitif dilakukan lokal oleh pihak berwenang tanpa mencetak nilai.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.25-01 | Bukti penguatan keamanan di kode dan test (JWT, registrasi, uploads, ClamAV, Viewer, MFA terenkripsi) | `backend/*/pkg/config/jwt_secret*.go`; `backend/audit-service/pkg/safefile/`; `pkg/avscan/`; `routes/viewer_readonly_test.go`; `backend/auth-service/pkg/secretcrypt/` | Internal |
| EV-A8.25-02 | Kebijakan SDLC-001 disetujui (setelah tindakan 2) | `docs/ISO_27001/A.8.25_Siklus_Hidup_Pengembangan_yang_Aman.md` (Lampiran A); [belum ada versi disetujui] | Internal |
| EV-A8.25-03 | Panduan keamanan Draft: SCG-001, DRC-001, TMT-001, SARC-001, ASC-001 | `docs/ISO_27001/threat_modeling/` | Internal |
| EV-A8.25-04 | Threat model AuditSphere terisi dan register ancaman (setelah tindakan 7) | [belum ada] | Restricted |
| EV-A8.25-05 | Pengaturan GitHub: branch protection, required reviewers, akses (setelah tindakan 1) | [belum ada] | Restricted |
| EV-A8.25-06 | Workflow CI dengan test, vet, SCA dan hasil run (setelah tindakan 3) | [belum ada] | Internal |
| EV-A8.25-07 | Konfigurasi dan hasil secret scan dan SAST (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.25-08 | Daftar tinjauan berkas sensitif/terlacak dan catatan rotasi (tanpa nilai; setelah tindakan 5) | [belum ada] | Restricted |
| EV-A8.25-09 | Contoh DRC-001 terisi dan catatan tinjauan keamanan rilis (Lampiran C dan D; setelah tindakan 8) | [belum ada] | Restricted |
| EV-A8.25-10 | Hasil test keamanan otomatis per rilis (setelah tindakan 10) | [belum ada] | Internal |
| EV-A8.25-11 | Laporan uji penetrasi/DAST dan rencana remediasi (setelah tindakan 13) | [belum ada] | Restricted |
| EV-A8.25-12 | Kajian keamanan dan catatan perubahan: rencana dan runbook | `docs/backend-enhancement-plan.md`; `docs/security-hardening-changes-backend.md`; `docs/jwt-rotation-runbook.md`; `docs/mfa-encryption-rollout.md`; `.claude/agents/auditsphere-qa.md` | Internal |

Setelah tindakan 1 sampai 8 selesai, tambahkan ID per rilis produksi untuk checklist rilis dan hasil gerbang.

---

## Lampiran A — Draf Kebijakan Siklus Hidup Pengembangan yang Aman (SDLC-001, usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran. Instalasi dan deploy produksi diatur di SWI-001 (A.8.19); dokumen ini merujuk, bukan mengulang.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | SDLC-001 Kebijakan Siklus Hidup Pengembangan yang Aman (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden, perubahan arsitektur, atau perubahan alat) |
| Berlaku untuk | Seluruh kode dan konfigurasi AuditSphere: layanan Go, python-ai, frontend, Data Hub, skrip, Dockerfile/compose/Kong, workflow; pengembang internal, kontributor eksternal, dan agen AI |

### 1. Tujuan dan prinsip
Memastikan keamanan dirancang, diuji, dan diverifikasi pada setiap tahap, dan setiap perubahan yang dirilis dapat ditelusuri ke bukti gerbang. Prinsip: aman secara bawaan (tanpa nilai bawaan rahasia, tanpa seeder pada produksi), hak akses minimum, tolak secara bawaan, isolasi tenant sebagai invarian, gagal tertutup, pertahanan berlapis.

### 2. Fase siklus hidup dan aktivitas keamanan

| Fase | Aktivitas keamanan wajib | Artefak | Peran utama |
| :--- | :--- | :--- | :--- |
| P1 Persyaratan | Klasifikasi dampak fitur (data, peran, tenant); persyaratan keamanan aplikasi (A.8.26): autentikasi, otorisasi, validasi, log, retensi, residensi data. | Catatan persyaratan pada tiket/rencana | Project Lead, Lead Information Security |
| P2 Desain | DRC-001 untuk fitur, skema, integrasi, atau perubahan infrastruktur; threat model (Lampiran B) untuk perubahan pada batas kepercayaan; prinsip arsitektur aman (A.8.27). | DRC-001 terisi; pembaruan threat model | Solution Architect, Lead Information Security |
| P3 Implementasi | SCG-001; tanpa rahasia di repo; dependensi baru melalui allowlist (A.8.19 Lampiran B); migrasi aditif; tanpa nilai bawaan rahasia; kode hasil AI ditinjau manusia (butir 12). | Commit dengan pesan baku | Pengembang |
| P4 Tinjauan | Tinjauan keamanan untuk jalur sensitif (Lampiran D); pemeriksa bukan penulis. | Catatan tinjauan | Peninjau |
| P5 Verifikasi | Gerbang CI (G2), pengujian keamanan (A.8.29): test negatif RBAC/objek/tenant, SAST, SCA, secret scan; DAST/uji penetrasi berkala pada lingkungan sintetis. | Hasil run; laporan | `auditsphere-qa`, DevOps Lead |
| P6 Rilis | Gerbang rilis (G5) dan checklist (Lampiran C); persetujuan oleh peran selain pemicu; lanjut ke SWI-001 dan A.8.32. | Catatan rilis | DevOps Lead, Lead Information Security |
| P7 Operasi dan pemeliharaan | Pemantauan kerentanan dan patch (A.8.8); umpan balik insiden ke persyaratan; peninjauan register temuan. | Register temuan | DevOps Lead, Lead Information Security |
| P8 Penghentian | Hapus fitur/endpoint/dependensi tidak terpakai (contoh: endpoint tak terpakai), cabut kunci dan akses terkait, retensi data sesuai kebijakan. | Catatan penghentian | Engineering Lead |

### 3. Klasifikasi perubahan [usulan]

| Kelas | Contoh | Gerbang |
| :--- | :--- | :--- |
| Sensitif | Autentikasi/JWT/MFA/step-up, middleware RBAC, unggah dan unduh berkas, isolasi tenant, generator tenant, Kong/CORS, skrip deploy, workflow, migrasi data, kunci dan konfigurasi rahasia | G1 sampai G5 penuh, termasuk tinjauan keamanan tercatat |
| Standar | Fitur fungsional pada pola yang ada | G1 (ringkas), G2, G3, G5 |
| Rendah | Teks, locale, gaya, dokumen | G2 (otomatis), G5 |

### 4. Gerbang keamanan (security gates) [usulan]

| Gerbang | Kriteria lulus | Bukti | Penegakan |
| :--- | :--- | :--- | :--- |
| G1 Desain | DRC-001 terisi dan tidak ada *red flag* terbuka untuk perubahan Sensitif; threat model diperbarui bila batas kepercayaan berubah. | DRC-001, register ancaman | Peninjau desain; tinjauan kelas Sensitif |
| G2 Build dan uji otomatis | `go build`, `go vet`, `go test` hijau per layanan terdampak; vitest, ESLint, dan typecheck tanpa galat baru terhadap baseline; migrasi dapat dijalankan. | Log CI | CI pada push `staging` dan sebagai prasyarat deploy |
| G3 Pemindaian | `govulncheck`, `npm audit`, `pip-audit`, secret scan, SAST tanpa temuan kritis/tinggi terbuka tanpa pengecualian tertulis; tenggat patch [kritis 7 hari, tinggi 30 hari, sedang 90 hari]. | Laporan pemindaian | CI |
| G4 Tinjauan | Perubahan Sensitif ditinjau oleh pemeriksa bukan penulis memakai Lampiran D; tinjauan pasca-merge [mingguan] untuk perubahan Standar bila push langsung dipertahankan. | Catatan tinjauan | Penanggung jawab tinjauan; `CODEOWNERS` bila PR dipakai |
| G5 Rilis | Checklist Lampiran C lengkap; CI hijau pada SHA yang akan dirilis; persetujuan oleh peran selain pemicu; catatan perubahan A.8.19 Lampiran D terisi. | Catatan rilis | Lingkungan `production` dengan required reviewers |
| G6 Pasca-rilis | Smoke test alur yang disentuh; pemantauan [24 jam]; rollback siap. | Hasil smoke | DevOps Lead |

### 5. Aturan repositori dan rahasia [usulan]
Rahasia, kunci, kata sandi, token, dan berkas lingkungan tidak dikomit (tidak ada pengecualian untuk `.env.production` atau skrip lingkungan); rahasia diberikan lewat environment atau penyimpanan rahasia; nilai bawaan literal untuk rahasia dilarang; temuan rahasia di repo dianggap insiden (A.5.24) dan menuntut rotasi, bukan hanya penghapusan. Berkas biner besar dan artefak model dicatat asal dan checksum-nya. Pengaturan lokal alat pengembangan tidak memuat alamat host, jalur kunci, atau izin luas.

### 6. Tinjauan kode dan alur git [usulan]
Perubahan Sensitif tidak boleh dipromosikan tanpa pemeriksa kedua. Bila push langsung ke `staging` dipertahankan, kendali penggantinya adalah gerbang G4/G5 pada saat promosi dan tinjauan pasca-merge terjadwal. Branch protection pada `main` dan `production`; tidak ada force-push pada cabang bersama; merge dari `upstream` ditinjau diff-nya.

### 7. Lingkungan dan data uji [usulan]
Pengembangan dan uji memakai data sintetis. Salinan data produksi hanya pada lingkungan yang diperlakukan sebagai produksi (akses, enkripsi, retensi, dan pengiriman email dinonaktifkan atau diarahkan ke kotak uji); lingkungan training/demo dengan data nyata mengikuti kontrol produksi (A.8.31, A.8.33, draf).

### 8. Pengujian keamanan [usulan]
Setiap endpoint baru memiliki test negatif untuk otorisasi peran, objek, dan tenant; fitur berkas memiliki test penyeberangan jalur, tipe, dan pemindaian malware; perubahan JWT/MFA memiliki test gagal-tutup. DAST dan uji penetrasi eksternal [sebelum rilis mayor dan setidaknya tahunan] dengan ruang lingkup isolasi tenant.

### 9. Ketergantungan dan rantai pasok [usulan]
Dependensi baru melalui allowlist dan pemeriksaan lisensi/kerentanan (A.8.19 Lampiran B; A.5.32); satu lockfile per ekosistem; versi dipin; pembaruan terjadwal [bulanan] dan segera untuk kritis.

### 10. Migrasi dan seeder [usulan]
Migrasi aditif pada satu rilis dan destruktif pada rilis berikutnya; seeder tidak berjalan pada produksi dan tidak menimpa data; migrasi dijalankan sebagai langkah terpisah.

### 11. Penerimaan dan penghentian fitur [usulan]
Fitur/endpoint tak terpakai dihapus beserta dependensi dan kuncinya dalam [satu rilis] sejak ditandai usang.

### 12. Pengembangan berbantuan AI [usulan]
Kode hasil AI diperlakukan sama dengan kode kontributor eksternal: ditinjau manusia sesuai kelasnya; tidak ada data klien, rahasia, atau isi berkas lingkungan di prompt atau konteks alat; izin alat dibatasi pada kebutuhan dan tidak dikomit dengan alamat host atau kunci; commit ditandai; penggunaan alat baru mengikuti A.5.19 sampai A.5.23 (draf) dan AIP-001.

### 13. Pengembangan oleh pihak luar [usulan]
Kontributor eksternal dan fork menerima akses minimum, tunduk pada perjanjian kerahasiaan dan klausul keamanan (A.5.20), mengikuti SDLC-001, dan hasil kerjanya ditinjau sebelum merge; akses dicabut saat selesai. Lihat A.8.30 (draf).

### 14. Peran
Sesuai RACI Lampiran F.

### 15. Pengecualian dan pelanggaran
Pengecualian tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir [maksimum 90 hari]. Rilis yang melewati gerbang tanpa pengecualian dicatat sebagai penyimpangan (A.5.36).

---

## Lampiran B — Templat Threat Model Ringkas (dengan isian awal dari repo)

Templat ini menjembatani TMT-001 (metodologi STRIDE dan register B.6, Draft) untuk threat model v0.1 AuditSphere. Kolom "Kendali yang terlihat" mencatat keadaan di repo; kolom "Gap / pertanyaan" adalah titik awal tinjauan, bukan temuan final. Entri berlabel **[Rencana]** berasal dari `docs/backend-enhancement-plan.md` dan belum diverifikasi ulang.

### B.1 Kepala dokumen

| Field | Isi |
| :--- | :--- |
| ID / versi | TM-AUDITSPHERE-[nomor] / 0.1 |
| Lingkup | [fitur / layanan / seluruh sistem] |
| Pemilik / peninjau | [Solution Architect] / [Lead Information Security] |
| Tanggal / pemicu | [isi] / [fitur baru, perubahan batas kepercayaan, insiden, tinjauan berkala] |
| Aset | Data audit klien (temuan, kertas kerja, laporan), data pegawai, kredensial dan token, secret MFA, berkas unggahan, model AI, kunci per tenant |
| Aktor | Pengguna per peran (admin, manajer, auditor, auditee, viewer), admin tenant, operator, kontributor, agen AI, penyerang eksternal, tenant lain |

### B.2 Batas kepercayaan (diagram data aliran, dari repo)

| Batas | Alur | Catatan dari repo |
| :--- | :--- | :--- |
| TB1 Internet ke gateway | Browser, nginx/TLS, Kong | Kong hanya merutekan dan menerapkan CORS; tidak memvalidasi JWT |
| TB2 Gateway ke layanan | Kong ke auth, audit, master, risk, analytics, python-ai | Validasi JWT di tiap layanan; rahasia HMAC dipakai bersama dalam satu stack |
| TB3 Layanan ke penyimpanan | Postgres, Redis, Kafka, volume unggahan | Database terpisah per layanan/tenant; Redis dan Kafka jaringan internal |
| TB4 Layanan ke layanan | Pemanggilan servis internal (`servicetoken`, `authclient`, `masterclient`, `riskclient`) | Token servis di audit-service |
| TB5 Unggahan ke ClamAV | Audit/master ke klaster ClamAV | Mode `required` fail-closed bawaan |
| TB6 Aplikasi ke Data Hub | Master dan analytics ke Data Hub eksternal | Kunci API Data Hub; server terpisah |
| TB7 Aplikasi ke python-ai | Analytics ke python-ai (model, retrain) | Kunci API; model dimuat dari pickle |
| TB8 Aplikasi ke pemasok | Resend (email OTP), penyimpanan cadangan | Lihat A.5.20 dan A.8.21 |
| TB9 Antar-tenant | Setiap tenant satu stack dan basis data | Kunci JWT per tenant dari generator; Redis berbagi bila tidak dipisah (**Perlu verifikasi**) |
| TB10 Pengembang ke produksi | Repo, CI, SSH deploy | Lihat A.8.19 |

### B.3 Kandidat ancaman per kategori STRIDE

| ID | Kategori | Ancaman | Kendali yang terlihat | Gap / pertanyaan | Sumber |
| :--- | :--- | :--- | :--- | :--- | :--- |
| TH-01 | Spoofing | Pemalsuan JWT bila rahasia bocor atau lemah | Gerbang start rahasia (min. 64 byte, daftar digest), `WithValidMethods` | Rahasia HMAC bersama; nilai lama di riwayat git; rotasi diverifikasi? | Repo; [Rencana] B10 |
| TH-02 | Spoofing | Penyalahgunaan akun bawaan | Registrasi publik ditutup | Seeder dan skrip tenant membuat akun bawaan dengan kata sandi literal | Repo |
| TH-03 | Tampering | Mass assignment/filter pada CRUD generik | `status_guard`, allowlist peran | DTO eksplisit belum diverifikasi | [Rencana] B7 |
| TH-04 | Tampering | Penggantian artefak model atau berkas pickle | Tidak ada | Tidak ada checksum/provenans model | Repo |
| TH-05 | Repudiation | Aksi sensitif tanpa jejak yang tahan ubah | `activitylog`, log sistem, redaksi | Ketahanan sink, retensi, jejak pada rute tulis | Repo; [Rencana] B6 |
| TH-06 | Information disclosure | Baca lintas-tenant | Basis data dan kunci per tenant | Tidak ada test lintas-tenant (gap 9); Redis berbagi | Repo |
| TH-07 | Information disclosure | Unduh/akses berkas tanpa otorisasi atau penyeberangan jalur | `/uploads` terautentikasi, `safefile`, tipe konten allowlist | Rute lain dengan token di *query* | Repo; [Rencana] A9 |
| TH-08 | Information disclosure | Token/rahasia di log atau di repo | Middleware `redact`; digest rahasia | Pemindaian riwayat belum dilakukan | Repo |
| TH-09 | Information disclosure | Proxy ke layanan AI pihak ketiga dengan data/prompt bebas | Endpoint OpenAI dihapus di working tree | Belum dikomit; pernah di build produksi? | Repo |
| TH-10 | Denial of service | Brute force login/MFA, unggahan besar | Penguncian MFA; `tollbooth` pada audit-service | Batas laju login belum diverifikasi; batas ukuran unggah | Repo; [Rencana] B1 |
| TH-11 | Elevation of privilege | Peran Viewer atau peran rendah menulis | Middleware Viewer, allowlist peran, test matriks | Cakupan test pada layanan selain audit | Repo |
| TH-12 | Elevation of privilege | Unggahan berbahaya (malware, HTML/SVG aktif) | ClamAV fail-closed, tipe konten allowlist, `nosniff`, CSP sandbox | Mode ClamAV tiap stack (**Perlu verifikasi**) | Repo |
| TH-13 | Elevation of privilege | Kompromi jalur deploy atau kunci SSH root | `deploy.yml` terkendali | Banyak jalur, otorisasi tidak terbukti (A.8.19) | Repo |
| TH-14 | Tampering | Dependensi atau action CI berbahaya | Lockfile | Tanpa SCA/SBOM, action dipin tag | Repo; A.8.19 |
| TH-15 | Elevation of privilege | Kode atau prompt dari agen AI memasukkan kelemahan | Pemisahan peran agen | Tanpa tinjauan manusia wajib | Repo |

### B.4 Register ancaman (salin per ancaman)

| Kolom | Isi |
| :--- | :--- |
| ID / kategori STRIDE | TH-[nomor] / [S/T/R/I/D/E] |
| Aset dan batas kepercayaan | [isi] |
| Skenario | [isi] |
| Kemungkinan (L) x dampak (I) | [1-5] x [1-5] = [skor] [usulan skala, mengikuti TMT-001] |
| Kendali yang ada | [isi] |
| Mitigasi tambahan | [isi] |
| Pemilik / tenggat / status | [peran] / [tanggal] / [terbuka, dimitigasi, diterima] |
| Penerimaan risiko | [pemilik risiko, tanggal tinjau ulang; risiko diterima memerlukan persetujuan tertulis] |
| Uji yang membuktikan mitigasi | [nama test / hasil pemindaian] |

---

## Lampiran C — Checklist Rilis dan Keamanan (usulan)

Salin satu formulir per rilis produksi dan lampirkan pada catatan perubahan A.8.19 Lampiran D. Jangan mencantumkan kunci, kata sandi, atau alamat IP.

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Bukti |
| :--- | :--- | :--- | :--- |
| 1 | SHA dan ref rilis tercatat; daftar commit dan penulis sejak rilis terakhir ditinjau (A.8.19). | [isi] | |
| 2 | Kelas perubahan ditetapkan (Sensitif/Standar/Rendah); untuk Sensitif, DRC-001 terisi dan tinjauan keamanan tercatat. | [isi] | |
| 3 | CI hijau pada SHA yang sama: build, vet, test, lint, typecheck tanpa galat baru. | [isi] | |
| 4 | Pemindaian: SCA, secret scan, SAST tanpa temuan kritis/tinggi terbuka (atau pengecualian tertulis). | [isi] | |
| 5 | Tidak ada rahasia, kunci, berkas lingkungan, atau nilai bawaan literal baru pada diff. | [isi] | |
| 6 | Dependensi baru atau berubah sesuai allowlist, lisensi dan kerentanan diperiksa. | [isi] | |
| 7 | Test negatif untuk otorisasi peran, objek, dan tenant ada untuk endpoint baru/berubah dan lulus. | [isi] | |
| 8 | Fitur berkas: penyeberangan jalur, tipe konten, dan pemindaian malware diuji. | [isi] | |
| 9 | Migrasi aditif, dapat diulang, dan tidak menyentuh data produksi di luar rencana; seeder tidak berjalan pada produksi. | [isi] | |
| 10 | Perubahan Kong, CORS, rute: `route-coverage` hijau; tidak ada rute baru tanpa autentikasi layanan. | [isi] | |
| 11 | Perubahan konfigurasi/env didokumentasikan; tidak ada konfigurasi di luar repo yang tertinggal tanpa catatan. | [isi] | |
| 12 | Backup pra-deploy terverifikasi dan rencana rollback tersedia (A.8.13, A.8.19). | [isi] | |
| 13 | Persetujuan oleh peran selain pemicu; tanggal tercatat. | [isi] | |
| 14 | Smoke test pasca-rilis untuk alur yang disentuh (login/MFA, peran, unggah/unduh berkas, tenant) selesai; hasil dicatat. | [isi] | |
| 15 | Threat model dan register temuan diperbarui bila batas kepercayaan berubah. | [isi] | |

---

## Lampiran D — Daftar Periksa Tinjauan Kode (PR atau tinjauan pasca-merge, usulan)

Dipakai pada PR bila dipakai, atau pada tinjauan pasca-merge terjadwal bila push langsung ke `staging` dipertahankan. Pemeriksa bukan penulis. Untuk kelas Sensitif wajib sebelum promosi.

| No | Pemeriksaan | Hasil | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Endpoint baru berada di balik autentikasi dan otorisasi peran; ada pemeriksaan objek/departemen/tenant; Viewer tidak dapat menulis. | [isi] | |
| 2 | Tidak ada nilai bawaan rahasia; konfigurasi rahasia dari environment dan gagal-tutup bila kosong. | [isi] | |
| 3 | Masukan divalidasi di sisi server (tipe, panjang, allowlist kolom untuk sort/filter, DTO eksplisit, tidak bind langsung ke model). | [isi] | |
| 4 | Kueri berparameter; tidak ada SQL dari string pengguna; tidak ada kolom pengguna tak terdaftar. | [isi] | |
| 5 | Unggahan: batas ukuran, allowlist ekstensi dan MIME, nama berkas dibuat server, jalur terkurung, pemindaian ClamAV dengan mode `required`. | [isi] | |
| 6 | Respons dan log tidak membocorkan token, kata sandi, rahasia, atau galat internal; PII diredaksi. | [isi] | |
| 7 | Aksi sensitif mencatat aktor, aksi, entitas dan hasil ke log aktivitas; aksi berisiko tinggi memakai step-up. | [isi] | |
| 8 | Tidak ada kunci, token, atau data klien pada komentar, berkas uji, atau fixture; data uji sintetis. | [isi] | |
| 9 | Dependensi baru sesuai allowlist; versi dipin; lockfile konsisten (satu per ekosistem). | [isi] | |
| 10 | Test ditambahkan: kasus positif dan negatif (peran, objek, tenant), dan lulus. | [isi] | |
| 11 | Migrasi aman: aditif, tidak ada seeder pada produksi, tidak menimpa data. | [isi] | |
| 12 | Frontend: tanpa rahasia pada bundel, tanpa `v-html` pada data pengguna tanpa sanitasi, gating UI mengikuti RBAC namun tidak menggantikan pemeriksaan server. | [isi] | |
| 13 | Kode hasil AI telah dibaca dan dipahami penulis dan peninjau; tidak ada data sensitif dikirim ke alat. | [isi] | |
| 14 | Perubahan pada Kong/CORS/Docker/compose/workflow ditinjau terhadap prinsip hak akses minimum (port loopback, non-root bila mungkin). | [isi] | |
| 15 | Penghapusan kode/dependensi/endpoint tak terpakai dilakukan lengkap (route, dependensi, konfigurasi). | [isi] | |

---

## Lampiran E — Aturan Praktis Pengembangan Berbantuan AI dan Kontributor (usulan)

| No | Aturan | Catatan |
| :--- | :--- | :--- |
| 1 | Tidak ada data klien, rahasia, kunci, isi `.env`, atau dump basis data pada prompt, konteks, atau berkas yang dibaca alat AI. | Selaras A.8.12 dan A.5.12. |
| 2 | Kode hasil AI dan kontributor eksternal ditinjau manusia sesuai kelas perubahan. | Lampiran A butir 12. |
| 3 | Izin alat (`.claude/settings.json`) dibatasi pada kebutuhan, tanpa jalur kunci, alamat host, atau pola luas yang mengeksekusi perintah arbitrer; izin pribadi di berkas lokal tidak dikomit. | Berkas yang terlacak saat ini memuatnya (gap 7). |
| 4 | Agen yang menyentuh produksi (devops/platform) bertindak hanya atas konfirmasi eksplisit manusia; pengumpulan bukti tidak memicu deploy. | Selaras A.8.19. |
| 5 | Commit hasil bantuan AI ditandai (trailer). | Konvensi yang sudah berjalan. |
| 6 | Merge dari fork/kontributor mensyaratkan tinjauan diff dan hasil CI. | Tindakan 12. |
| 7 | Kontributor menandatangani kerahasiaan dan klausul keamanan sebelum akses; akses dicabut saat selesai. | A.5.20, A.6.3. |

---

## Lampiran F — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | Engineering Lead | DevOps Lead | Solution Architect | Backend/Frontend Lead | QA | Direktur Utama |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan SDLC-001 | A/R | C | C | C | C | C | A (persetujuan) |
| Threat model dan DRC-001 | A | C | C | R | R | C | I |
| Persyaratan keamanan aplikasi | A | R | I | R | C | I | I |
| Implementasi aman (SCG-001) | C | A | I | C | R | I | I |
| Tinjauan kode (jalur sensitif) | A | R | C | C | R | I | I |
| Gerbang CI, SAST, SCA, secret scan | A | C | R | I | R | C | I |
| Pengujian keamanan dan regresi | C | A | C | I | R | R | I |
| Uji penetrasi/DAST independen | A/R | C | C | C | C | I | I |
| Checklist rilis dan persetujuan | A | C | R | I | C | R | I |
| Register temuan dan penutupan | A | R | R | C | R | C | I |
| Aturan AI dan kontributor eksternal | A | R | C | I | C | I | I |
| Evidence ID dan review berkala | A/R | C | C | I | I | I | I |

---

## Lampiran G — Daftar Periksa Tinjauan Berkala (usulan)

Frekuensi tinjauan: [kuartalan] oleh Lead Information Security dan Engineering Lead. Setiap tinjauan dicatat (tanggal, peninjau, hasil).

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Setiap rilis produksi pada periode ini memiliki Checklist Lampiran C terisi dan penyetuju selain pemicu. | [isi] | |
| 2 | CI gate (G2 dan G3) aktif dan tidak dinonaktifkan; pengecualian tercatat dan belum kedaluwarsa. | [isi] | |
| 3 | Pengaturan GitHub (branch protection, required reviewers, akses tulis) tidak berubah tanpa persetujuan. | [isi] | |
| 4 | Perubahan kelas Sensitif pada periode ini memiliki tinjauan keamanan tercatat; tinjauan pasca-merge dijalankan sesuai jadwal. | [isi] | |
| 5 | Temuan pemindaian kritis/tinggi ditutup dalam [7/30 hari]; register temuan diperbarui. | [isi] | |
| 6 | Threat model diperbarui untuk perubahan batas kepercayaan; risiko diterima ditinjau ulang. | [isi] | |
| 7 | Pemindaian rahasia pada riwayat dan repo bersih; tidak ada berkas sensitif terlacak baru. | [isi] | |
| 8 | Nilai bawaan rahasia dan seeder pada jalur bawaan tidak muncul kembali. | [isi] | |
| 9 | Test keamanan otomatis (lintas-tenant, matriks RBAC, upload) berjalan dan lulus; jumlah dan hasil dicatat. | [isi] | |
| 10 | Aturan AI dan kontributor eksternal dipatuhi; merge upstream memiliki tinjauan diff. | [isi] | |
| 11 | Panduan SCG-001/DRC-001/TMT-001 diperbarui dan versi disetujui ([6 bulan]). | [isi] | |
| 12 | Pelatihan pengkodean aman diikuti pengembang baru dan kontributor (A.6.3). | [isi] | |
| 13 | Uji penetrasi/DAST terakhir dan status remediasi: [isi tanggal]. | [isi] | |
| 14 | Data uji sintetis; lingkungan dengan data nyata mengikuti kontrol produksi. | [isi] | |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.8.19 Instalasi perangkat lunak pada sistem operasional | Ada (`docs/ISO_27001/A.8.19_Instalasi_Perangkat_Lunak_pada_Sistem_Operasional.md`; belum ter-commit) | Jalur deploy, persetujuan rilis, allowlist, pin versi, SBOM; tidak diulang di sini. |
| A.8.13 Pencadangan informasi | Ada (`docs/ISO_27001/A.8.13_Pencadangan_Informasi.md`; belum ter-commit) | Dump pra-deploy dan rencana rollback pada gerbang rilis. |
| A.8.21 Keamanan layanan jaringan | Ada (`docs/ISO_27001/A.8.21_Keamanan_Layanan_Jaringan.md`; belum ter-commit) | Kong, CORS, pemasok layanan jaringan pada batas kepercayaan threat model. |
| A.5.8 Keamanan informasi dalam manajemen proyek | Ada | DRC-001, TMT-001, SCG-001 6.6 dan gap CI/PR template; tindakan 6 dan 7 di sana selaras tindakan 3 dan 8 di sini. |
| A.5.36 Kepatuhan terhadap kebijakan dan standar | Ada | EX-04 (CI tanpa test), register penyimpangan. |
| A.5.32 Hak kekayaan intelektual | Ada | SBOM, lisensi dependensi, lockfile ganda. |
| A.5.24 Persiapan manajemen insiden | Ada | Rahasia yang bocor di repo sebagai insiden; instalasi tak berizin. |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada | Kontributor eksternal, pemasok image/registri dan GitHub. |
| A.6.3 Awareness, edukasi, dan pelatihan keamanan | Ada | Pelatihan pengkodean aman. |
| A.8.12 Pencegahan kebocoran data | Ada | Data klien pada prompt AI dan rahasia di repo. |
| A.5.19, A.5.21, A.5.22, A.5.23 | **Draf, dokumen belum ada** | Rantai pasok TIK dan layanan pihak ketiga termasuk alat AI dan GitHub Actions. |
| A.5.8 (lanjutan), A.8.4 Akses ke kode sumber | **Draf, dokumen belum ada** (A.8.4) | Akses repo, branch protection, penandatanganan commit. |
| A.8.8 Manajemen kerentanan teknis | **Draf, dokumen belum ada** | Pemindaian dan batas waktu patch. |
| A.8.9 Manajemen konfigurasi | **Draf, dokumen belum ada** | Nilai bawaan literal, konfigurasi di luar repo. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Jejak aktivitas dan log CI/deploy. |
| A.8.24 Penggunaan kriptografi | **Draf, dokumen belum ada** | Enkripsi secret MFA, kunci JWT, kebijakan rotasi. |
| A.8.26 Persyaratan keamanan aplikasi | **Draf, dokumen belum ada** | Fase P1. |
| A.8.27 Arsitektur sistem yang aman dan prinsip rekayasa | **Draf, dokumen belum ada** | Fase P2, threat model. |
| A.8.28 Pengkodean aman | **Draf, dokumen belum ada** | SCG-001, aturan repositori. |
| A.8.29 Pengujian keamanan dalam pengembangan dan penerimaan | **Draf, dokumen belum ada** | Gerbang G2 dan G3, uji penetrasi, DAST. |
| A.8.30 Pengembangan yang dialihdayakan | **Draf, dokumen belum ada** | Kontributor eksternal dan fork. |
| A.8.31 Pemisahan lingkungan pengembangan, pengujian, dan produksi | **Draf, dokumen belum ada** | Dev dan training berbagi host/data dengan produksi. |
| A.8.32 Manajemen perubahan | **Draf, dokumen belum ada** | Persetujuan dan rekaman perubahan; Lampiran C dan A.8.19 Lampiran D sebagai masukan. |
| A.8.33 Informasi pengujian | **Draf, dokumen belum ada** | Data uji sintetis vs salinan produksi. |
| A.8.34 Perlindungan sistem informasi selama uji audit | **Draf, dokumen belum ada** | Pengujian keamanan pada lingkungan berisi data nyata. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.25	Technological	Siklus hidup pengembangan yang aman	Aktivitas keamanan terintegrasi pada seluruh SDLC	Secure SDLC, security gates, threat model, release checklist	Berlaku	AuditSphere dikembangkan sendiri (Go, Vue, Python, skrip deploy dan tenant) dan memproses data audit klien lembaga jasa keuangan multi-tenant; cacat kode (otorisasi, isolasi tenant, unggah berkas) langsung menjadi insiden data; banyak penulis termasuk kontributor eksternal dan agen AI; temuan keamanan sebelumnya lolos tanpa gerbang (lihat A.8.25 bagian 1)	Sebagian	Sebagian (kode, test, catatan perubahan keamanan, dan panduan SCG/DRC/TMT Draft ada di repo; hasil CI, tinjauan kode bertanda, threat model terisi, hasil pemindaian, dan laporan uji penetrasi belum ada)	Belum terdaftar (usulan EV-A8.25-01 s/d EV-A8.25-12)	Lead Information Security (didukung Engineering Lead dan DevOps Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah CI gate dan kebijakan SDLC-001 disetujui	Perlu perbaikan	Belum teruji	CI hanya memeriksa cakupan rute Kong (tanpa test, vet, SCA, secret scan, SAST); tanpa bukti branch protection, CODEOWNERS, PR template dan commit langsung ke staging tanpa tinjauan sejawat tercatat; panduan keamanan Draft tanpa salinan terisi; tanpa threat model dan uji penetrasi; rahasia pernah dikomit dan berkas sensitif/berlebih masih terlacak; endpoint proxy OpenAI tanpa autentikasi dihapus hanya di working tree; nilai bawaan literal dan seeder pada jalur bawaan; tanpa test lintas-tenant; model dimuat via pickle tanpa checksum; penggunaan AI dan kontributor tanpa aturan tertulis; klaim SAST/DAST di dokumen keselarasan tidak terbukti	Verifikasi pengaturan GitHub; setujui SDLC-001 dengan gerbang pada promosi ke deploy; CI gate minimum; secret scan dan SAST; tangani rahasia dan berkas terlacak; komit penghapusan openai.ts; threat model v0.1; DRC-001 dan Security DoD; hapus nilai bawaan dan seeder; test lintas-tenant dan RBAC; integritas model; aturan AI dan kontributor; uji penetrasi/DAST; register temuan; Evidence ID (15 tindakan di bagian 5)	Tinggi	[usulan] 31 Desember 2026 (kebijakan, verifikasi GitHub, CI gate minimum); 28 Februari 2027 (threat model, SAST, uji keamanan independen)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
