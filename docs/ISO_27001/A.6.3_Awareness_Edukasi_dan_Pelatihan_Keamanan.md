# A.6.3 — Awareness, Edukasi, dan Pelatihan Keamanan

| Field | Isi |
| :--- | :--- |
| **Tema** | People |
| **Judul kontrol** | Awareness, edukasi, dan pelatihan keamanan |
| **Fokus pemeriksaan** | Program pelatihan sesuai peran diberikan dan efektivitasnya diukur |
| **Contoh bukti objektif** | Materi, attendance, quiz, simulasi phishing, training matrix |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Legal & Compliance dan HR) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 31 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Hal yang berada di luar repo (data SDM, kontrak, konfigurasi GitHub/VPS, layanan email, kondisi tenant) ditandai **Perlu verifikasi**. Seluruh isian pada Lampiran A sampai H (training matrix, templat, skenario, KPI, target, jadwal) adalah **draf usulan**: belum disetujui manajemen, belum ada salinan terisi, dan angka target bukan hasil pengukuran. Tidak ada kemampuan simulasi phishing yang ditemukan di repo; bagian terkait menandainya sebagai gap.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Tim kecil dengan banyak peran memegang data audit klien (KKA, LHA, ATR), kredensial, akses produksi, dan alat AI. Kesalahan satu orang (phishing, salah deploy, data klien ke AI publik) langsung berdampak ke klien lembaga jasa keuangan. Kebijakan yang tidak disosialisasikan (A.5.1, AIP-001, A.5.32, A.5.36) tidak dapat dipatuhi. |
| Implementasi status | **Belum diimplementasikan** | Tidak ditemukan program awareness, training matrix, materi induksi, catatan kehadiran, quiz, maupun simulasi phishing. Yang ada hanya bahan baku: pedoman tertulis (AIP-001, SCG-001, runbook), manual pengguna, dan *acknowledgement* Pakta Integritas di aplikasi. Draf program, matriks, dan templat diusulkan pada Lampiran A sampai H. |
| Ketersediaan bukti | **Ada sebagian** | Tersedia materi sumber (bagian 2.1) dan bukti acceptance pengguna sistem (tabel `confidentiality_agreements`). Materi pelatihan yang disusun untuk tujuan awareness, daftar hadir, hasil quiz, laporan simulasi phishing, dan training matrix terisi **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A6.3-01 s/d EV-A6.3-10) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum efektif** | Tidak ada pengukuran apa pun. Satu-satunya temuan eksplisit tentang pelatihan di repo, **NC-06** pada Advanced Security Operations Checklist (pelatihan analisis forensik dan penanganan insiden belum dilaksanakan), menegaskan ketiadaan pelatihan. Gap teknis yang bersumber dari pengetahuan (akun default, secret ter-commit, seeder di produksi) masih terbuka (A.5.16, A.5.36). |
| Desain kontrol | **Belum dirancang** | Draf siklus (induksi → pelatihan berbasis peran → penilaian → simulasi → pengukuran → perbaikan materi) dan matriks pelatihan per peran diusulkan pada Lampiran A, B, dan E; belum disetujui. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Beberapa kontrol lain bergantung pada pengetahuan personel: aturan deploy aman tanpa seeder (A.5.8, A.5.36), larangan data klien ke AI publik (AIP-001), pelaporan insiden (A.5.24), dan klausul IP/lisensi (A.5.32). Auditor klien akan menanyakan bukti pelatihan dan efektivitasnya. |
| Review berikutnya | [isi] | Usulan: evaluasi efektivitas triwulanan; tinjau ulang desain program dan matriks 12 bulan setelah disetujui; sinkron dengan audit internal ISMS tahunan (ISO-AS-001 bagian 13.2). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.6.3 **berlaku** untuk AuditSphere karena:

1. AuditSphere menyimpan kertas kerja (KKA), laporan (LHA), tindak lanjut (ATR), dan kredensial klien enterprise (lembaga jasa keuangan). Orang yang memiliki akses (developer, DevOps, admin/support) adalah titik kegagalan paling mungkin, terutama pada tim kecil di mana satu orang memegang beberapa peran.
2. Praktik kerja menyentuh produksi secara langsung. Seeder yang menimpa data produksi, port yang terbuka, dua repo yang dapat men-deploy, dan unggahan yang tidak persisten (A.5.8, A.5.36 EX-01, EX-02, EX-12) menunjukkan bahwa aturan deploy aman hanya efektif bila pelaku mengetahui dan memilih jalur yang benar (`backend/deploy-prod.sh --ref`).
3. Alat AI dipakai dalam pengembangan (riwayat git memuat komit `google-labs-jules[bot]` dan alur kerja dengan agen Claude Code, lihat A.5.32 bagian 1 butir 4) dan ada fitur AI di produk. AIP-001 melarang memasukkan kredensial, kode proprietary, data klien, dan data pribadi ke AI publik dengan sanksi disipliner berat, tetapi tidak ditemukan bukti bahwa isinya pernah disosialisasikan.
4. Kontributor lebih dari satu orang (6 nama penulis di riwayat git, termasuk bot) dengan status karyawan/kontraktor yang tidak dapat diverifikasi dari repo (A.5.32 bagian 2.4). Kontributor baru perlu diberi aturan sebelum memegang akses.
5. Pengguna klien/tenant juga bagian dari rantai keamanan: akun pertama tenant memakai kredensial awal, MFA tidak diwajibkan per role (A.5.16 bagian 4), dan pengguna AUDITEE mengunggah dokumen bukti. Edukasi pengguna klien mengurangi risiko akun diambil alih dan kebocoran data.
6. ISO-AS-001 (v0.1 DRAFT) bagian 3 dan 13 menetapkan kepatuhan UU PDP No. 27/2022 dan audit internal tahunan. Kewajiban pelatihan UU PDP bagi personel pemroses data pribadi **perlu verifikasi** oleh Legal & Compliance; tidak ditemukan pemetaannya di repo.
7. A.5.1 tindakan 8 (sosialisasi kebijakan ke staf) dan A.5.36 (pemantauan kepatuhan) memerlukan bukti bahwa kebijakan telah dikomunikasikan. Kontrol ini adalah mekanismenya.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Materi sumber yang dapat menjadi bahan pelatihan (belum berupa program)

| Materi | Isi relevan untuk awareness | Status | Lokasi |
| :--- | :--- | :--- | :--- |
| AIP-001 AI Usage Policy v1.0 | Klasifikasi tool AI (Public/Enterprise/Private/Embedded), matriks data vs tool, daftar aktivitas diizinkan dan dilarang, aturan per departemen (Software Engineering dan DevOps: pindai dependensi usulan AI sebelum commit), tanggung jawab dan sanksi. Pengesahan tercantum "Ditandatangani Digital, 28 Sep 2026", tetapi bukti tanda tangan tidak dapat diverifikasi dari repo (lihat A.5.1 bagian 4). | Dokumen ada; sosialisasi tidak terbukti | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx` |
| SCG-001 Secure Coding Guide v1.0 (Draft) | Pedoman Go, Nuxt, Python, SQL/migrasi, infrastruktur; bagian 6.1 manajemen secret, 6.2 masking PII di log, 6.4 UU PDP untuk developer, 6.6 Security Definition of Done. | Draft; pemilik placeholder | `docs/ISO_27001/threat_modeling/SCG-001-Secure-Coding-Guide.docx` |
| DRC-001, TMT-001, SARC-001, ASC-001 | Checklist tinjauan desain, STRIDE, tinjauan arsitektur, keamanan API (OWASP API Top 10 2023). | Draft; tidak ada salinan terisi | `docs/ISO_27001/threat_modeling/` |
| CSP-001 Cloud Security Policy v1.0 | Aturan cloud/infrastruktur, notifikasi insiden butir 20. | Dokumen ada; sosialisasi tidak terbukti | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` |
| FRP-001, RTP-001, TIP-001, BBP-001 | Prosedur forensik, rencana red/purple team (menyebut spear-phishing terotorisasi), intelijen ancaman, bug bounty (rekayasa sosial di luar cakupan). | Belum disahkan atau belum berjalan | `docs/ISO_27001/advanced_security/` |
| Runbook operasional | Deploy produksi aman, rotasi JWT, backup/restore, backfill, ringkasan runbook eksekutif dengan label [AGEN]/[KONFIRMASI]/[MANUAL]. | Ada, sebagian berstatus plan/DRAFT | `backend/deploy-prod.sh`, `docs/jwt-rotation-runbook.md`, `docs/backup-b2-setup.md`, `docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md`, `docs/execution-progress-backfill-runbook.md`, `docs/deploy-tidy-plan.md` |
| Annex ISO (A.5.1, A.5.8, A.5.12, A.5.16, A.5.20, A.5.24, A.5.28, A.5.32, A.5.36) | Ringkasan kebijakan, gap, dan tindakan; basis topik pelatihan. Seluruhnya draf internal dan sebagian belum ter-commit. | Draf | `docs/ISO_27001/A.5.*.md` |
| Dokumentasi & Manual User AuditSphere | Manual pengguna 9 bab (login, MFA, Pakta Integritas, reset password, seluruh modul, matriks RBAC per bab, perangkat tepercaya). Format `.docx` (sekitar 31 MB). | Ada; untuk pengguna produk, bukan materi keamanan | `docs/user-manual/Dokumentasi & Manual User AuditSphere.docx` |
| Panduan pengembang/operasional | Setup backend, demo, provisioning domain klien, pengujian API. | Ada; bukan materi keamanan | `docs/BACKEND_RUN_AND_DEPLOY.md`, `docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md`, `docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md`, `docs/API_Testing_Guide.md` |
| Definisi peran dan pembagian kerja agen | Pemisahan tugas pengembangan, QA, dan produksi (`auditsphere-devops` untuk produksi). | Ada; bukan prosedur yang disetujui | `.claude/agents/` |

### 2.2 Pengendalian di aplikasi yang bersifat edukatif/pengingat bagi pengguna

| Fitur | Fungsi awareness | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Pakta Integritas / Kerahasiaan saat onboarding pengguna | Pengguna harus menggulir sampai >= 95% dan menyetujui sebelum memakai sistem; dicatat dengan versi, waktu, IP, user agent. | `frontend/pages/auth/confidentiality.vue`; `backend/auth-service/models/confidentiality_agreement.go`; manual bagian 1.3 | Hanya *acknowledgement*, bukan pelatihan atau quiz; versi tidak divalidasi server dan tidak ada re-acceptance saat berubah; kegagalan simpan dapat terlewati (A.5.1 bagian 4 butir 3-5). |
| Paksa ganti password awal | Pengguna yang dibuat admin diarahkan ke `/auth/change-password`. | `frontend/middleware/01-force-password-change.global.ts`; `frontend/pages/auth/change-password.vue` | Akun seeder default tidak diberi `must_change_password` (A.5.16 butir 4). |
| Halaman MFA dan perangkat tepercaya | Pengguna dapat mengaktifkan TOTP dan mengelola perangkat tepercaya. | `frontend/pages/settings/mfa.vue`, `frontend/pages/settings/devices.vue`, `frontend/pages/auth/mfa-login.vue` | Tidak ada teks edukasi kontekstual yang diverifikasi di halaman; MFA tidak diwajibkan per role. |
| Pemulihan akun | Alur lupa/reset password. | `frontend/pages/auth/forgot-password.vue`, `reset-password.vue` | Tidak ada peringatan tentang email phishing (**perlu verifikasi** pada templat email). |
| Step-up password | Konfirmasi ulang sebelum perubahan struktur organisasi. | `POST /auth/verify-password`; komit `bf84bff` | Cakupan terbatas (A.5.36 bagian 2.2). |
| Panduan di dalam aplikasi | Modal "Panduan Siklus Audit" dan tautan "Lihat Panduan" skala penilaian (konten domain audit, bukan keamanan). | `frontend/pages/audit-execution-status/index.vue`; kunci i18n `cycleGuideButton`, `guidelinesSection` di `frontend/locales/{id,en}/common.json` | Tidak ada tur produk, modul pelatihan, atau pusat bantuan keamanan. |
| Matriks RBAC per bab pada manual | Pengguna mengetahui hak aksesnya per modul. | Manual bagian 1.15, 2.6, 3.8, dst. | Penegakan server belum menyeluruh (A.5.36 EX-11); teks manual dapat berbeda dari perilaku. |

### 2.3 Sinyal bahwa pelatihan belum berjalan (bukan bukti pelaksanaan)

| Sumber | Isi | Catatan |
| :--- | :--- | :--- |
| Advanced Security Operations Checklist (`docs/ISO_27001/advanced_security/Advanced_Security_Operations_Checklist.docx`) | Baris "Pelatihan berkala analisis forensik dan penanganan insiden bagi tim SOC/IR (A.6.3)": status NC; **NC-06**, belum dilaksanakan pada siklus tahun berjalan. | Temuan non-konformitas yang sudah tercatat; belum ada tindakan korektif di repo. |
| Advanced Security Operations Checklist | Pernyataan "simulasi spear-phishing terotorisasi dalam jadwal tahunan" dan "sosialisasi menunggu peluncuran program publik" (BBP-001). | Klaim jadwal; **tidak ditemukan bukti pelaksanaan** (RTP-001 tidak berisi laporan latihan). |
| ZTAR-001 Zero Trust Architecture Roadmap | Menyebut "edukasi berkala" sebagai bagian program change management. | Rencana, bukan pelaksanaan. |
| A.5.24 bagian 2.3 | Tidak ada catatan tabletop, drill, atau simulasi insiden. | Terkait pelatihan insiden (Lampiran A, T08). |

### 2.4 Kemampuan simulasi phishing

Tidak ditemukan di repo: platform simulasi phishing, skrip, templat email simulasi, daftar target, atau laporan hasil (pencarian kata `phishing` pada `docs/`, `scripts/`, `backend/`, `monitoring/`, `frontend/` hanya menemukan dokumen kebijakan lanjutan di atas, `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md` dan A.5.32; tidak ada alat atau skrip simulasi). Infrastruktur pengiriman email tenant ada (`scripts/provision-tenant-email.sh`, kunci pengiriman per tenant lewat Resend dengan domain bersama `mail.auditsphere.app`), tetapi:

- Domain dan kunci tersebut adalah jalur email produksi tenant. Menggunakannya untuk simulasi phishing berisiko merusak reputasi pengiriman dan melanggar perjanjian layanan penyedia; **tidak boleh dipakai** untuk simulasi tanpa persetujuan Platform dan Legal & Compliance.
- Apakah penyedia email atau Google Workspace/Microsoft 365 internal perusahaan memiliki fitur simulasi phishing **perlu verifikasi** (di luar repo).

Karena itu simulasi phishing berstatus **gap** (Lampiran D memuat rencana dan templat).

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kebijakan/prosedur program awareness dan pelatihan (disetujui) | **Belum tersedia** | Draf di Lampiran A. |
| Training matrix per peran (terisi dan disetujui) | **Belum tersedia** | Templat dan matriks usulan di Lampiran B; belum ada catatan per orang. |
| Materi pelatihan (slide/modul/handout) | **Belum tersedia** | Hanya materi sumber (bagian 2.1) yang belum dirakit menjadi modul. Rancangan materi inti di Lampiran C. |
| Attendance / daftar hadir atau bukti penyelesaian | **Belum tersedia** | Templat di Lampiran F. Pakta Integritas di aplikasi hanya mencakup pengguna sistem dan bukan pelatihan. |
| Quiz dan hasil penilaian | **Belum tersedia** | Contoh soal di Lampiran G. |
| Simulasi phishing (rencana, kampanye, hasil, tindak lanjut) | **Belum tersedia** | Tidak ada kemampuan di repo (bagian 2.4); rencana dan skenario di Lampiran D. |
| Laporan efektivitas (KPI, tren, tindakan perbaikan materi) | **Belum tersedia** | Templat dan KPI di Lampiran E. |
| Bukti onboarding/offboarding kontributor (briefing keamanan, NDA, pencabutan akses) | **Belum tersedia / Perlu verifikasi** | Berkas SDM dan kontrak di luar repo. Prosedur usulan di Lampiran H. |
| Bukti acceptance pengguna sistem (Pakta Integritas) | Tersedia (sebagian) | Tabel `confidentiality_agreements`; perlu sampel query (A.5.1 tindakan 7). |
| Sertifikasi/pelatihan eksternal personel (mis. ISO 27001 Lead Implementer/Auditor, OWASP, CISA yang dirujuk ISO-AS-001 bagian 13.2) | **Perlu verifikasi** | Tidak ditemukan di repo. |
| Bukti pelatihan pengguna klien/tenant (sesi serah terima, materi, daftar hadir) | **Belum tersedia** | Manual pengguna ada; sesi pelatihan tenant tidak tercatat. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada program awareness.** Tidak ditemukan prosedur, jadwal, materi induksi, atau catatan pelatihan keamanan di `docs/`, `scripts/`, maupun `frontend/`. NC-06 pada Advanced Security Operations Checklist mengonfirmasi bahwa pelatihan insiden belum dilaksanakan.
2. **Tidak ada training matrix per peran.** Peran dalam praktik: developer backend/frontend, DevOps, QA, Project Lead, Legal & Compliance, admin/support, kontributor/kontraktor, dan pengguna klien (ADMIN, AUDITOR, AUDITEE, EXECUTIVE, VIEWER, DEPARTMENT_HEAD di `backend/auth-service/pkg/database/seeders/seeder.go`). Tidak ada daftar siapa wajib mengikuti apa.
3. **Kebijakan yang menjadi materi belum berlaku formal.** SCG-001, DRC-001, TMT-001, SARC-001, ASC-001 berstatus Draft dengan pemilik placeholder; ISO-AS-001 v0.1 DRAFT; bukti tanda tangan CSP-001 dan AIP-001 tidak dapat diverifikasi (A.5.1, A.5.36 EX-06). Pelatihan atas dokumen yang belum disetujui harus dilabeli draf.
4. **Aturan AIP-001 tidak punya bukti sosialisasi dan tidak ditegakkan teknis.** Tidak ada kontrol teknis yang mencegah pengiriman data klien atau kredensial ke AI publik; `frontend/server/api/openai.ts` memakai `OPENAI_API_KEY` tanpa kontrak yang tercatat (A.5.32 butir 6). Satu-satunya pengendali adalah kesadaran personel.
5. **Gap yang bersumber dari pengetahuan/kebiasaan masih terbuka.** Akun seeder dengan password default yang sama (A.5.16 butir 4, EX-05), secret JWT pernah ter-commit (EX-09), jalur deploy yang menjalankan seeder (EX-01), persetujuan tercatat di percakapan, bukan sign-off (A.5.8 butir 4). Materi deploy aman dan manajemen kredensial menjadi prioritas.
6. **Tidak ada pelatihan insiden maupun latihan (tabletop/drill).** IR plan, playbook, dan contact tree belum ada (A.5.24 bagian 2.3), sehingga tidak ada materi untuk dilatihkan; NC-06 terbuka.
7. **Simulasi phishing tidak ada** dan tidak ada alat (bagian 2.4). Klaim "dalam jadwal tahunan" pada Advanced Security Operations Checklist tidak punya bukti pelaksanaan.
8. **Pakta Integritas di aplikasi bukan pelatihan.** Mekanismenya hanya scroll dan setuju, bersifat sekali, tanpa validasi pemahaman, dan memiliki kelemahan teknis (A.5.1 butir 3-5). Tidak boleh diklaim sebagai pemenuhan A.6.3.
9. **Pengguna klien/tenant tidak mendapat edukasi keamanan terdokumentasi.** Manual pengguna fokus pada fungsi modul; tidak ada bagian khusus phishing, kebersihan kredensial, atau pelaporan insiden. MFA tidak diwajibkan per role (A.5.16). Akun admin tenant baru memakai kredensial awal (`docs/org-structure-model-plan.md` bagian 11 butir 9).
10. **Onboarding/offboarding kontributor tidak terdokumentasi.** Tidak ada checklist briefing keamanan, penerimaan aturan AIP-001/SCG-001, NDA/IP (A.5.32), maupun pencabutan akses (GitHub, VPS, secret, akun aplikasi). Sesi tidak dicabut saat akun dinonaktifkan (A.5.16 butir 1). Status karyawan/kontraktor tiap kontributor tidak dapat diverifikasi dari repo.
11. **Tidak ada metrik efektivitas.** Tidak ada KPI (kehadiran, nilai quiz, laporan insiden oleh staf, tingkat klik phishing) maupun mekanisme meninjau materi setelah insiden.
12. **Tidak ada pelatihan UU PDP.** SCG-001 bagian 6.4 memuat ringkasan untuk developer, tetapi tidak ada sesi atau catatan; kewajiban hukum pelatihan **perlu verifikasi** Legal & Compliance.
13. **Pemetaan topik ke annex belum ada.** Setiap annex menghasilkan tindakan perbaikan yang membutuhkan perubahan perilaku, tetapi tidak ada jalur resmi menyampaikannya ke personel.
14. **Status dokumen.** Annex A.5.12, A.5.16, A.5.20, A.5.24, A.5.28, A.5.32, A.5.36 berstatus untracked di git per 30 Oktober 2026 (A.5.36 butir 14); dokumen ini juga belum ter-commit.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Setujui dan tetapkan **Prosedur Program Awareness dan Pelatihan Keamanan** berdasarkan Lampiran A (induksi, tahunan, berbasis peran, berbasis kejadian), termasuk pemilik, anggaran waktu, dan konsekuensi tidak mengikuti. | Lead Information Security dengan Legal & Compliance dan HR | Tinggi | [isi] | Prosedur bertanda tangan |
| 2 | Tetapkan **daftar personel dan peran** (karyawan, kontraktor, kontributor, bot/agen yang dikelola tim) dan resmikan **training matrix** (Lampiran B). | Lead Information Security, Project Lead | Tinggi | [isi] | Training matrix terisi dan disetujui |
| 3 | Rakit **materi inti** (Lampiran C) dari materi sumber bagian 2.1, label versi, pemilik, dan tanggal review. Mulai dari yang paling mendesak: T02 kredensial/MFA, T04 AI (AIP-001), T06 deploy aman, T07 insiden. | Security Officer, DevOps Lead, Technical Lead | Tinggi | [isi] | Modul/slide/handout bertanggal |
| 4 | Selesaikan dahulu dokumen yang menjadi materi: tetapkan pemilik dan persetujuan AIP-001, CSP-001, SCG-001 dan seterusnya; simpan salinan bertanda tangan (A.5.1 tindakan 2, A.5.36 tindakan 12). | Lead Information Security | Tinggi | [isi] | Dokumen dengan document control lengkap |
| 5 | Laksanakan **sesi induksi/sosialisasi pertama** untuk seluruh personel dan kontributor aktif; catat kehadiran dan *acknowledgement* (Lampiran F). Ini sekaligus memenuhi A.5.1 tindakan 8. | Lead Information Security | Tinggi | [isi] | Lembar kehadiran, acknowledgement |
| 6 | Laksanakan **quiz** pasca pelatihan (Lampiran G), tetapkan batas lulus dan remedial, simpan hasil tanpa memuat jawaban rahasia. | Security Officer | Sedang | [isi] | Rekap hasil quiz |
| 7 | Tutup **NC-06**: rancang dan laksanakan pelatihan/latihan penanganan insiden dan forensik dasar setelah IR plan dan playbook tersedia (A.5.24 tindakan terkait, FRP-001); lakukan tabletop pertama. | Lead Information Security, DevOps Lead | Tinggi | [isi] | Catatan tabletop, daftar hadir, tindak lanjut |
| 8 | Putuskan **pendekatan simulasi phishing** (alat tersedia dari penyedia email/suite kantor, layanan pihak ketiga, atau latihan manual terbatas), dapatkan persetujuan Legal & Compliance dan Platform, lalu jalankan kampanye baseline (Lampiran D). Jangan memakai domain dan kunci email produksi tenant. | Security Officer, Platform, Legal & Compliance | Sedang | [isi] | Rencana kampanye disetujui, hasil, laporan |
| 9 | Terapkan **checklist onboarding dan offboarding** kontributor/kontraktor (Lampiran H): briefing keamanan, penerimaan AIP-001/SCG-001, NDA/IP (selaras A.5.32 tindakan terkait), pemberian akses minimum, pencabutan akses dan rotasi secret saat keluar. | Project Lead, HR, Legal & Compliance, DevOps Lead | Tinggi | [isi] | Checklist terisi per orang |
| 10 | Buat **materi edukasi untuk pengguna klien/tenant** (Lampiran A.5): ringkasan keamanan akun, MFA, phishing, unggah dokumen, pelaporan insiden; sertakan dalam paket serah terima tenant dan sebagai halaman bantuan; catat sesi serah terima. Koordinasi dengan tim frontend dan penerjemahan ID/EN. | Project Lead, Tim Frontend, `auditsphere-software-translation` | Sedang | [isi] | Materi, catatan serah terima |
| 11 | Usulkan perbaikan aplikasi yang mendukung awareness: re-acceptance Pakta Integritas per versi (A.5.1 tindakan 4-6), paksa ganti password akun default (A.5.16), teks edukasi pada halaman MFA dan perangkat tepercaya. | Tim Backend dan Tim Frontend | Sedang | [isi] | Perubahan kode dan test (di luar dokumen ini) |
| 12 | Pastikan kewajiban pelatihan hukum/kontrak dipetakan: UU PDP, ketentuan sektor keuangan klien, permintaan pelatihan di kontrak klien. Masukkan ke register kewajiban (A.5.36 tindakan 13). | Legal & Compliance | Tinggi | [isi] | Register kewajiban dengan kolom pelatihan |
| 13 | Susun **laporan efektivitas** pertama (Lampiran E) dan jadwalkan triwulanan; masukkan KPI ke monitoring report A.5.36 Lampiran C. | Lead Information Security, Project Lead | Sedang | [isi] | Laporan efektivitas bertanggal |
| 14 | Catat pelatihan eksternal/sertifikasi yang dimiliki personel dan rencana pengembangan (ISO 27001, secure coding, forensik, privasi). Verifikasi sertifikasi yang dirujuk ISO-AS-001 bagian 13.2. | Lead Information Security, HR | Rendah | [isi] | Daftar kompetensi |
| 15 | Commit dokumen ini dan dokumen annex yang belum ter-commit (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |
| 16 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

Catatan pelaksanaan: materi deploy aman dan rotasi secret harus selaras dengan runbook aktual dan dilatihkan sebagai simulasi di staging/lokal, bukan di produksi. Setiap tindakan yang menyentuh produksi tetap melalui `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh` dengan `--ref`, tanpa seeder).

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A6.3-01 | Materi sumber kebijakan dan pedoman (AIP-001, CSP-001, SCG-001 dan templat tinjauan desain) | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx`, `cloud_saas/CSP-001_Cloud_Security_Policy.docx`, `threat_modeling/` | Internal |
| EV-A6.3-02 | Runbook operasional yang menjadi bahan pelatihan deploy aman dan insiden | `backend/deploy-prod.sh`, `docs/jwt-rotation-runbook.md`, `docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md`, `docs/deploy-tidy-plan.md` | Internal |
| EV-A6.3-03 | Manual pengguna dan matriks RBAC per bab | `docs/user-manual/Dokumentasi & Manual User AuditSphere.docx` | Internal (dapat dibagikan ke klien setelah ditinjau) |
| EV-A6.3-04 | Acceptance Pakta Integritas di aplikasi (kode dan sampel query tanpa data pribadi) | `frontend/pages/auth/confidentiality.vue`, `backend/auth-service/`, tabel `confidentiality_agreements` | Restricted |
| EV-A6.3-05 | Temuan NC-06 dan checklist operasi keamanan lanjutan | `docs/ISO_27001/advanced_security/Advanced_Security_Operations_Checklist.docx` | Internal |
| EV-A6.3-06 | Prosedur program awareness dan pelatihan (setelah tindakan 1) | [belum ada] | Internal |
| EV-A6.3-07 | Training matrix terisi dan disetujui (setelah tindakan 2); draf usulan Lampiran B | `docs/ISO_27001/A.6.3_Awareness_Edukasi_dan_Pelatihan_Keamanan.md` (Lampiran B) | Internal |
| EV-A6.3-08 | Materi pelatihan, lembar kehadiran, dan hasil quiz per sesi (setelah tindakan 3, 5, 6) | [belum ada] | Restricted (hasil per orang) |
| EV-A6.3-09 | Rencana dan laporan simulasi phishing (setelah tindakan 8) | [belum ada] | Restricted |
| EV-A6.3-10 | Laporan efektivitas triwulanan dan checklist onboarding/offboarding terisi (setelah tindakan 9, 13) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 9 selesai, tambahkan ID per siklus untuk setiap sesi, kampanye, dan laporan.

---

## Lampiran A — Program Awareness dan Pelatihan (usulan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Tim kecil: sesi dirancang singkat, dapat direkam, dan peran dirangkap digabung.

**A.1 Komponen program**

| Komponen | Sasaran | Waktu | Format | Penilaian | Catatan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Induksi keamanan | Karyawan, kontraktor, dan kontributor baru | Sebelum akses produksi/data klien diberikan; maksimum 5 hari kerja sejak mulai | Sesi 60-90 menit + baca AIP-001 dan ringkasan SCG-001 + *acknowledgement* | Quiz induksi (Lampiran G.1), lulus >= 80% | Akses ke repo dan staging boleh lebih awal; akses produksi/data klien menunggu induksi lulus |
| Awareness tahunan (penyegaran) | Seluruh personel | Setahun sekali, dan setelah perubahan kebijakan mayor | Sesi 60 menit + quiz | Quiz tahunan, lulus >= 80% | Jadwalkan sebelum audit internal ISMS (ISO-AS-001 13.2) |
| Pelatihan berbasis peran | Sesuai Lampiran B | Saat peran dimulai dan tiap 12 bulan; setelah perubahan teknis mayor | Workshop praktik (lab), *walkthrough* runbook, tinjauan kode | Praktik terobservasi + quiz peran | Lab deploy dilakukan di lingkungan lokal/staging |
| Pembaruan ringkas (micro-awareness) | Seluruh personel | Bulanan (5 menit) | Pesan/buletin: satu kasus nyata dari repo (anonim) atau insiden industri | Tidak dinilai; tercatat | Sumber: temuan annex, near-miss, siaran CTI (TIP-001) |
| Simulasi phishing | Seluruh personel (dan admin/support prioritas) | Baseline sekali, lalu semi-tahunan | Lampiran D | Metrik Lampiran D.4 | Bukan hukuman |
| Latihan insiden (tabletop) | Tim insiden: Lead Information Security, DevOps, Backend, Legal & Compliance, Project Lead | Semi-tahunan, setelah IR plan tersedia | Skenario 90 menit; catat waktu deteksi dan keputusan | Catatan latihan dan pelajaran | Menutup NC-06 sebagian; lihat A.5.24 |
| Berbasis kejadian | Pihak terdampak | Dalam 14 hari setelah insiden, near-miss kepatuhan, atau temuan berulang | Sesi pelajaran dan pembaruan materi | Ditinjau pada laporan efektivitas | Pemicu dari A.5.36 Lampiran D |

**A.2 Kewajiban dan konsekuensi**

- Pelatihan induksi dan tahunan wajib. Tidak lulus: remedial maksimum 2 kali dalam 14 hari; bila tetap gagal, akses sensitif ditinjau oleh Lead Information Security.
- Tidak hadir tanpa alasan: dijadwalkan ulang, dilaporkan ke atasan/Project Lead. Penegakan disipliner mengikuti kontrak dan kebijakan SDM (**di luar repo**).
- Pelaporan insiden atau kesalahan secara jujur tidak dihukum (*just culture*); pelanggaran disengaja mengikuti eskalasi A.5.36 Lampiran D.2.

**A.3 Pengelolaan program**

| Hal | Usulan |
| :--- | :--- |
| Pemilik program | Lead Information Security |
| Pemilik materi per topik | Lihat Lampiran C |
| Penyimpanan catatan | Folder terbatas yang terversi; satu berkas per sesi (materi, daftar hadir, rekap quiz); hasil per orang bersifat *Restricted* |
| Retensi | Minimal 3 tahun (disarankan; selaras kebutuhan audit sertifikasi), **perlu persetujuan Legal & Compliance** |
| Tinjauan materi | Tahunan dan setelah insiden; pemilik topik memperbarui versi |
| Bahasa | Indonesia sebagai utama; versi Inggris bila ada personel/kontributor non-Indonesia (koordinasi `auditsphere-software-translation` untuk materi di aplikasi) |
| Perlindungan data pribadi | Data kehadiran dan skor adalah data pribadi personel; dikelola dengan hak akses terbatas dan tidak dipublikasikan per individu ke klien (laporan ke klien dalam bentuk agregat) |

**A.4 Pesan kunci yang diulang di setiap sesi**

1. Jangan pernah memasukkan kredensial, kunci API, atau data klien ke alat AI publik (AIP-001 bagian 3).
2. Produksi hanya diubah lewat jalur resmi dengan persetujuan; jangan menjalankan seeder di produksi.
3. Curigai permintaan mendesak yang meminta kredensial, kode MFA, atau pembayaran; verifikasi lewat kanal lain.
4. Laporkan segera hal yang mencurigakan ke Lead Information Security (target < 1 jam pada jam kerja, usulan).
5. Setiap akses diberikan sesuai kebutuhan dan dicabut saat tidak diperlukan.

**A.5 Pelatihan pengguna klien/tenant (usulan)**

| Pengguna | Topik | Format | Waktu | Catatan |
| :--- | :--- | :--- | :--- | :--- |
| Admin tenant (ADMIN) | Pengelolaan akun, penonaktifan pengguna keluar, MFA, ganti password awal, reset password, pembatasan role, pelaporan insiden | Sesi serah terima 60 menit + panduan ringkas | Saat tenant dibuat (`scripts/onboard-tenant.sh`) dan setelah perubahan fitur besar | Wajib; ganti kredensial awal pada sesi ini |
| Auditor, DEPARTMENT_HEAD, EXECUTIVE | Phishing, MFA/perangkat tepercaya, penanganan KKA/LHA, jangan berbagi akun, Pakta Integritas | Panduan ringkas + manual pengguna bab 1 | Saat akun dibuat | Dapat berupa halaman bantuan di aplikasi (tindakan 10) |
| AUDITEE | Unggah dokumen bukti dengan benar (jenis berkas, ukuran, tanpa data pribadi berlebih), tanggapan manajemen, ATR | Panduan satu halaman + manual bab 5 | Saat unit auditee pertama diaktifkan | Banyak pengguna tidak teknis; bahasa sederhana |
| VIEWER | Akses baca saja, batas penggunaan data | Panduan satu halaman | Saat akun dibuat | |

Pelatihan tenant diposisikan sebagai **layanan/pendampingan**; tanggung jawab keamanan tenant ditegaskan dalam kontrak dan DPA (A.5.20 Lampiran terkait; **perlu verifikasi** Legal & Compliance).

---

## Lampiran B — Training Matrix per Peran (usulan)

### B.1 Peran dan kode

| Kode peran | Peran | Catatan dalam AuditSphere |
| :--- | :--- | :--- |
| R1 | Seluruh personel (karyawan, kontraktor, kontributor) | Dasar untuk semua |
| R2 | Developer backend (Go, Python) | auth/master/audit/risk/analytics, python-ai |
| R3 | Developer frontend (Nuxt) | `frontend/` |
| R4 | DevOps / Platform | Compose, Kong, deploy, tenant stack, VPS, email, monitoring |
| R5 | QA | `auditsphere-qa`; tes RBAC, smoke check |
| R6 | Project Lead / PM | Gate proyek, sign-off, onboarding tim |
| R7 | Legal & Compliance | Kontrak, IP, UU PDP, DPA |
| R8 | Admin / Support (operasional tenant) | Reset password, pembuatan akun, penanganan tiket klien |
| R9 | Kontributor/kontraktor eksternal | Akses terbatas dan sementara |
| R10 | Tim insiden (Lead Information Security, DevOps, Backend, Legal, Project Lead) | Peran tambahan |
| R11 | Manajemen (Direktur Utama, Lead Information Security) | Penyetuju kebijakan dan risiko |
| R12 | Pengguna klien/tenant | Lampiran A.5 |

### B.2 Topik

| Kode | Topik | Materi sumber utama |
| :--- | :--- | :--- |
| T01 | Kebijakan keamanan, klasifikasi, dan tanggung jawab personel | A.5.1, A.5.12, ISO-AS-001 |
| T02 | Phishing, rekayasa sosial, kredensial, MFA, kebersihan akun | A.5.16, manual bab 1 |
| T03 | Penanganan data klien dan data pribadi (UU PDP), masking, klasifikasi | SCG-001 6.2 dan 6.4, A.5.12 |
| T04 | Penggunaan AI yang aman | AIP-001 |
| T05 | Secure coding (Go/Nuxt/Python/SQL), Security Definition of Done, API security | SCG-001, ASC-001, DRC-001 |
| T06 | Deploy produksi aman, seeder, secret, backup/restore, jaringan | `deploy-prod.sh`, A.5.8, A.5.36, runbook |
| T07 | Pelaporan dan respons insiden, bukti forensik | A.5.24, A.5.28, FRP-001 |
| T08 | Latihan insiden (tabletop) | A.5.24 |
| T09 | Hak kekayaan intelektual, lisensi, NDA | A.5.32 |
| T10 | Pemasok dan perjanjian, tenant isolation | A.5.20, `onboard-tenant.sh` |
| T11 | Pengujian keamanan dan RBAC | QA, ASC-001 |
| T12 | Dukungan klien yang aman (verifikasi identitas, reset password, tidak membagikan data lintas tenant) | A.5.16, ISO-AS-001 |
| T13 | Manajemen proyek aman (gate, sign-off, risk register) | A.5.8 |
| T14 | Kepatuhan dan pengecualian | A.5.36 |
| T15 | Hukum dan regulasi (UU PDP, sektor keuangan klien) | ISO-AS-001 13.3, register kewajiban |
| T16 | Pemanfaatan produk dengan aman (untuk pengguna klien) | Manual pengguna |

### B.3 Matriks (W = wajib, S = disarankan, - = tidak berlaku)

| Topik | R1 | R2 | R3 | R4 | R5 | R6 | R7 | R8 | R9 | R10 | R11 | R12 | Frekuensi | Metode | Penilaian |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- | :--- | :--- |
| T01 | W | W | W | W | W | W | W | W | W | W | W | - | Induksi + tahunan | Sesi + baca + acknowledgement | Quiz >= 80% |
| T02 | W | W | W | W | W | W | W | W | W | W | W | S | Induksi + tahunan + simulasi | Sesi + simulasi phishing | Quiz + metrik simulasi |
| T03 | W | W | W | W | W | W | W | W | W | W | W | S | Induksi + tahunan | Sesi + kasus | Quiz |
| T04 | W | W | W | W | W | W | W | W | W | W | W | - | Induksi + tahunan + saat kebijakan berubah | Sesi + studi kasus data yang boleh/tidak boleh | Quiz skenario |
| T05 | - | W | W | S | S | S | - | - | W | - | - | - | Induksi peran + tahunan | Workshop + tinjauan kode | Praktik + quiz; hasil tinjauan kode sebagai bukti |
| T06 | - | S | - | W | S | W | - | - | S | W | - | - | Induksi peran + tiap 6 bulan atau setelah perubahan jalur deploy | Walkthrough runbook + lab staging | Praktik terobservasi (deploy uji tanpa seeder) |
| T07 | W | W | W | W | W | W | W | W | W | W | W | S | Induksi + tahunan | Sesi singkat: apa yang dilaporkan, ke siapa, seberapa cepat | Quiz + uji lapor |
| T08 | - | S | - | W | S | W | W | - | - | W | W | - | Semi-tahunan | Tabletop | Catatan latihan |
| T09 | W | W | W | W | W | W | W | S | W | S | W | - | Induksi + tahunan | Sesi + dokumen | Quiz + tanda tangan NDA/IP |
| T10 | - | S | - | W | - | W | W | - | - | S | W | - | Tahunan | Sesi | Quiz |
| T11 | - | S | S | S | W | S | - | - | - | - | - | - | Induksi peran + tahunan | Workshop | Praktik |
| T12 | - | - | - | S | - | S | S | W | - | - | - | - | Induksi peran + tahunan | Skrip dan latihan peran | Observasi + quiz |
| T13 | - | S | S | S | S | W | - | - | - | - | W | - | Induksi peran + tahunan | Sesi | Quiz |
| T14 | - | - | - | W | S | W | W | - | - | S | W | - | Tahunan | Sesi | Tidak dinilai |
| T15 | S | S | S | S | S | S | W | S | S | S | W | - | Tahunan | Sesi dan pembaruan regulasi | Catatan |
| T16 | - | - | - | - | - | - | - | S | - | - | - | W | Saat akun dibuat | Panduan + sesi serah terima (admin) | Pakta Integritas + catatan serah terima |

Catatan: matriks ini usulan awal. Peran dirangkap (tim kecil) mengikuti gabungan topik wajib seluruh perannya. Daftar personel per peran, tanggal selesai, dan skor dicatat pada templat F.4 dan **belum ada**. Jumlah personel: [isi].

### B.4 Templat catatan training matrix per orang

| Kolom | Isi |
| :--- | :--- |
| Nama / ID personel | |
| Status (karyawan / kontraktor / kontributor / lainnya) | |
| Peran (R1-R12) | |
| Tanggal mulai / selesai | |
| Topik wajib (T01-T16) | |
| Tanggal selesai per topik | |
| Skor / hasil per topik | |
| Tanggal jatuh tempo berikutnya | |
| Status (Terpenuhi / Jatuh tempo / Lewat / Dibebaskan dengan alasan) | |
| Akses sensitif diberikan setelah induksi? | Ya / Tidak (tanggal) |
| Catatan | |

---

## Lampiran C — Rancangan Materi Inti (usulan)

Setiap modul memiliki pemilik, versi, tanggal review, dan daftar kasus nyata (anonim). Seluruh materi yang bersandar pada dokumen berstatus Draft diberi label "mengikuti dokumen sumber, draf".

| Modul | Tujuan belajar | Isi minimum | Dasar di repo | Pemilik (usulan) | Durasi |
| :--- | :--- | :--- | :--- | :--- | :--- |
| M1 Phishing dan rekayasa sosial | Mengenali dan melaporkan email/pesan mencurigakan, termasuk yang meniru klien, vendor, atau rekan | Ciri phishing, spear-phishing terhadap admin/DevOps, tautan reset password palsu, permintaan mendesak, kanal verifikasi alternatif, cara melapor | A.5.16, TIP-001, RTP-001 | Security Officer | 30 menit |
| M2 Kredensial, MFA, dan akun | Mengelola kredensial dan MFA dengan benar | Password manager, larangan berbagi akun, TOTP dan perangkat tepercaya (`/settings/mfa`, `/settings/devices`), ganti password awal, tidak menyimpan secret di repo, rotasi, step-up password | A.5.16, SCG-001 6.1, manual 1.2-1.4 | Security Officer | 30 menit |
| M3 Data klien dan data pribadi | Memperlakukan data klien sesuai klasifikasi | Jenis data (KKA, LHA, ATR, data Data Hub), klasifikasi, tenant isolation, masking PII di log/tiket/tangkapan layar, larangan data produksi di lingkungan dev, hak subjek data, UU PDP singkat | SCG-001 6.2 dan 6.4, A.5.12, ISO-AS-001 13.3 | Legal & Compliance dengan Technical Lead | 45 menit |
| M4 Penggunaan AI | Memakai AI tanpa membocorkan data | Kelas alat AI, matriks data vs alat, hal terlarang, review manusia, pindai dependensi usulan AI, pengelolaan prompt untuk kode, penggunaan agen pengembangan (agen tidak diberi kredensial produksi), status hak cipta keluaran AI (A.5.32) | AIP-001 | Security Officer | 30 menit |
| M5 Secure coding | Menulis dan meninjau kode sesuai SCG-001 | Validasi input, otorisasi per endpoint dan objek (RBAC/tenant), error tanpa bocor internal, upload aman, JWT, XSS/`v-html`, token di klien, SSRF, seeder insert-only, Security Definition of Done, uji negatif RBAC | SCG-001, ASC-001, DRC-001, A.5.8 | Technical Lead | 2 jam + lab |
| M6 Deploy produksi aman | Mengubah produksi tanpa merusak data atau membuka akses | Jalur resmi `deploy.yml` dan `backend/deploy-prod.sh --ref`, mengapa seeder tidak boleh jalan di produksi, backup sebelum perubahan, rollback, tidak menyentuh unggahan, jalur lama yang masih ada (EX-01), repo kedua, port dan firewall, rotasi JWT dengan persetujuan | `deploy.yml`, `deploy-prod.sh`, `docs/deploy-tidy-plan.md`, A.5.8, A.5.36 | DevOps Lead | 90 menit + lab |
| M7 Insiden: kenali, laporkan, amankan bukti | Melapor cepat dan tidak merusak bukti | Apa yang dianggap insiden, ke mana dan seberapa cepat melapor, jangan menghapus log atau mereboot tanpa arahan, catat waktu, notifikasi klien (CSP-001 butir 20: 1x24 jam / 3x24 jam, **tenggat resmi perlu verifikasi**), chain of custody | A.5.24, A.5.28, FRP-001 | Lead Information Security | 45 menit |
| M8 IP, lisensi, NDA | Menghindari pelanggaran lisensi dan kebocoran IP | Lisensi dependensi (GPL/AGPL/source-available), SBOM, penggunaan konten IIA/IPPF, pengalihan IP kontributor, NDA, kepemilikan data tenant | A.5.32 | Legal & Compliance | 30 menit |
| M9 Dukungan klien yang aman | Melayani klien tanpa membuka celah | Verifikasi identitas sebelum reset password/MFA, tidak membagikan kredensial lewat chat/email, tidak membuka data lintas tenant, pencatatan tiket tanpa data sensitif, eskalasi | A.5.16, A.5.20 | Project Lead | 30 menit |
| M10 Proyek aman dan kepatuhan | Menjalankan gate dan mencatat pengecualian | DRC-001, TMT-001, sign-off tertulis, risk register proyek, exception register, siklus temuan | A.5.8, A.5.36 | Project Lead | 45 menit |
| M11 Pengguna klien/tenant | Memakai produk dengan aman | Lihat Lampiran A.5; Pakta Integritas, MFA, phishing, unggah dokumen, melapor | Manual pengguna, `confidentiality.vue` | Project Lead dengan Tim Frontend | 30-60 menit |

**C.1 Aturan materi**

- Memakai contoh nyata dari repo secara anonim (tanpa nama orang, host, atau nilai secret). Contoh yang boleh dipakai: seeder menimpa data, secret ter-commit, port terbuka, data klien ke AI publik.
- Tidak menyalin teks standar IIA/ISO secara verbatim; merujuk dan meringkas (A.5.32 butir 11).
- Tidak memuat kredensial, IP, atau token nyata. Alamat server tidak disalin ke materi untuk klien.
- Setiap modul diakhiri 5-10 soal quiz dan satu *takeaway*.

---

## Lampiran D — Simulasi Phishing (rencana dan templat)

Status: **gap**. Bagian ini adalah rencana; belum ada kampanye yang dijalankan.

**D.1 Prasyarat dan batasan**

1. Persetujuan tertulis Direktur Utama dan Legal & Compliance sebelum kampanye. Kampanye hanya menyasar personel internal (dan kontributor yang telah diberi tahu bahwa latihan dapat terjadi dalam briefing induksi). **Pengguna klien/tenant tidak disimulasikan** tanpa kontrak dan persetujuan klien.
2. Tidak menggunakan domain pengiriman produksi (`mail.auditsphere.app`) maupun kunci email tenant (A.5.20); gunakan domain terpisah atau alat milik penyedia simulasi. Koordinasi dengan `auditsphere-platform` terkait SPF/DKIM/DMARC dan reputasi.
3. Halaman jebakan tidak boleh menangkap kata sandi atau kode MFA yang nyata. Hanya catat bahwa tautan diklik atau formulir dikirim (tanpa isi).
4. Data hasil per individu hanya untuk Lead Information Security dan Security Officer; laporan ke manajemen dan klien bersifat agregat. Pemrosesan data personel mengikuti UU PDP (**perlu verifikasi** Legal & Compliance, termasuk pemberitahuan di kebijakan SDM).
5. Tidak ada sanksi atas klik. Tujuan: pembelajaran. Pengulangan klik memicu pelatihan tambahan, bukan hukuman.
6. Tim penanganan (DevOps/Platform) diberi tahu agar tidak mengeskalasi simulasi sebagai insiden nyata; tetapi pelapor tetap menerima konfirmasi.
7. Alat: belum ditetapkan. Pilihan yang perlu dievaluasi: fitur bawaan suite email perusahaan (**perlu verifikasi**), alat sumber terbuka yang dihosting sendiri (perhatikan lisensi, A.5.32), atau layanan pihak ketiga (perlu DPA, A.5.20). **Keputusan manusia diperlukan.**

**D.2 Rencana kampanye (usulan)**

| Kampanye | Waktu | Sasaran | Skenario | Tujuan |
| :--- | :--- | :--- | :--- | :--- |
| P0 Baseline | Setelah induksi pertama | Seluruh personel internal | S1 (notifikasi tiket) | Mengukur tingkat awal |
| P1 | +3 bulan | Seluruh personel | S2 (reset password palsu) | Mengukur dampak pelatihan |
| P2 | +6 bulan | Admin/support dan DevOps | S3 (spear-phishing kredensial) | Peran berisiko tinggi |
| P3 | +9 bulan | Seluruh personel | S4 (lampiran/tautan dokumen) | Variasi teknik |
| Ulang | Semi-tahunan | Seluruh personel | Rotasi skenario | Tren |

**D.3 Templat skenario**

| Kolom | Isi |
| :--- | :--- |
| ID skenario | S-nn |
| Tema dan pemicu | Contoh: pemberitahuan keamanan palsu; tiket klien mendesak |
| Pengirim yang ditiru | Kategori (vendor, klien, atasan, platform); tanpa meniru orang sungguhan secara spesifik tanpa persetujuan |
| Isi pesan | Ringkasan; teks lengkap disimpan terpisah |
| Aksi yang diukur | Buka, klik, kirim formulir, lapor |
| Tingkat kesulitan | Rendah / Sedang / Tinggi |
| Sasaran | Peran/kelompok |
| Persetujuan | Nama, peran, tanggal |
| Tanda-tanda yang dapat dikenali | Daftar untuk umpan balik |
| Umpan balik langsung | Halaman edukasi setelah klik |

**D.4 Skenario awal (contoh; semuanya fiktif)**

| ID | Skenario | Tanda-tanda yang diajarkan |
| :--- | :--- | :--- |
| S1 | "Notifikasi tiket dari sistem dukungan: klien meminta tindak lanjut hari ini", tautan ke halaman jebakan. | Domain pengirim tidak dikenal, nada mendesak, tautan tidak cocok dengan domain sah |
| S2 | "Sesi AuditSphere Anda akan berakhir; verifikasi password", meniru halaman login. | Aplikasi tidak meminta password lewat email; periksa domain; gunakan bookmark |
| S3 | "Repositori/penyedia cloud: aktivitas login baru, konfirmasi MFA", ditujukan ke DevOps/admin. | Jangan membagikan kode MFA; konfirmasi lewat konsol langsung |
| S4 | "Dokumen kertas kerja terlampir untuk ditinjau" dari pengirim yang meniru klien. | Lampiran tak terduga; verifikasi lewat kanal lain; ekstensi mencurigakan |
| S5 | "Permintaan pembayaran atau perubahan rekening vendor" ditujukan ke manajemen. | Verifikasi lewat telepon ke nomor yang sudah dikenal |

**D.5 Metrik simulasi**

| Metrik | Definisi | Target usulan |
| :--- | :--- | :--- |
| Tingkat klik | Klik / penerima | Baseline dicatat; <= 15% pada P1, <= 10% pada P3 |
| Tingkat pengiriman data | Formulir dikirim / penerima | <= 5% pada P1; menurun tiap kampanye |
| Tingkat pelaporan | Pelapor / penerima | >= 40% pada P1; >= 60% pada P3 |
| Waktu median pelaporan | Dari pengiriman sampai laporan pertama | <= 30 menit pada jam kerja |
| Klik berulang | Orang yang klik pada >= 2 kampanye berturut-turut | 0; ditindaklanjuti dengan pelatihan tambahan |

Target di atas adalah **usulan** tanpa dasar historis; tetapkan ulang setelah baseline P0.

**D.6 Templat laporan kampanye**

| Bagian | Isi |
| :--- | :--- |
| Identitas | ID kampanye, tanggal, penanggung jawab, persetujuan |
| Skenario dan sasaran | Ringkasan dan jumlah penerima |
| Hasil | Metrik D.5, dibanding kampanye sebelumnya |
| Temuan | Pola (peran, jam, jenis tema) tanpa nama |
| Tindak lanjut | Pelatihan tambahan, perubahan materi, perubahan teknis (misalnya filter email) |
| Pelajaran | Apa yang diubah pada M1 |
| Distribusi | Agregat ke manajemen; rincian terbatas |

---

## Lampiran E — Pengukuran Efektivitas (KPI dan laporan)

### E.1 KPI (usulan; target untuk ditetapkan manajemen)

| No | KPI | Definisi | Sumber data | Frekuensi | Target (usulan) | Baseline 31 Okt 2026 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| E1 | Cakupan induksi | Personel baru yang menyelesaikan induksi sebelum akses sensitif / personel baru | Training matrix, daftar hadir | Triwulanan | 100% | Tidak ada data |
| E2 | Cakupan pelatihan wajib | Personel yang tuntas seluruh topik wajib / seluruh personel | Training matrix | Triwulanan | >= 95% | Tidak ada data |
| E3 | Penyelesaian tepat waktu | Pelatihan selesai sebelum jatuh tempo / seluruh yang jatuh tempo | Training matrix | Triwulanan | >= 90% | Tidak ada data |
| E4 | Tingkat kelulusan quiz | Lulus percobaan pertama / peserta | Rekap quiz | Per sesi | >= 85% | Tidak ada data |
| E5 | Skor rata-rata quiz | Rata-rata nilai | Rekap quiz | Per sesi | >= 85 (usulan) | Tidak ada data |
| E6 | Peningkatan pengetahuan | Selisih skor pasca-sesi dan pra-sesi (bila pra-tes dipakai) | Rekap quiz | Per sesi | Positif | Tidak ada data |
| E7 | Tingkat klik simulasi phishing | Lampiran D.5 | Laporan kampanye | Per kampanye | Menurun tiap kampanye | Belum ada kampanye |
| E8 | Tingkat pelaporan simulasi | Lampiran D.5 | Laporan kampanye | Per kampanye | >= 40% lalu naik | Belum ada kampanye |
| E9 | Pelaporan insiden/near-miss oleh staf | Jumlah laporan oleh personel / periode | Register insiden (A.5.24) | Triwulanan | > 0 dan tidak menurun | Register belum ada |
| E10 | Waktu laporan insiden | Median waktu dari kejadian ke pelaporan | Register insiden | Triwulanan | <= 1 jam kerja | Register belum ada |
| E11 | Insiden/temuan berakar kelalaian pengetahuan | Jumlah insiden atau NC yang akar masalahnya "pengetahuan/kebiasaan" | Analisis akar masalah (A.5.36 D.1 langkah 4) | Triwulanan | Menurun | Belum diukur; contoh historis: EX-01, EX-05, EX-09 |
| E12 | Pelanggaran aturan AI/deploy | Kasus pelanggaran AIP-001 atau jalur deploy tidak resmi | Compliance review (A.5.36 K7) | Bulanan | 0 | Perlu verifikasi |
| E13 | Akun default / MFA | Admin tenant dan personel yang memakai MFA; akun default aktif | Review akses | Triwulanan | MFA personel 100%; default 0 | Perlu verifikasi |
| E14 | Re-acceptance Pakta Integritas | Pengguna aktif yang menyetujui versi terbaru / pengguna aktif | Tabel `confidentiality_agreements` | Triwulanan | 100% | Perlu verifikasi (sampel query, A.5.1 tindakan 7) |
| E15 | Umpan balik peserta | Skor kepuasan/kejelasan | Survei singkat | Per sesi | >= 4 dari 5 | Tidak ada data |
| E16 | Latihan insiden | Jumlah tabletop dan tindakan perbaikan tertutup | Catatan latihan | Semi-tahunan | >= 2 per tahun; tindakan tertutup <= 60 hari | 0 (NC-06 terbuka) |
| E17 | Kesegaran materi | Modul yang melewati tanggal review | Daftar materi | Triwulanan | 0 | Belum ada materi |

Catatan: efektivitas tidak cukup dinilai dari kehadiran. Kombinasikan hasil pengetahuan (E4-E6), perilaku (E7-E10, E12-E13), dan hasil (E11). Metrik E7-E10 bergantung pada tindakan 8 dan register insiden A.5.24. Seluruh baseline berarti tidak ada data, bukan nilai terukur.

### E.2 Templat laporan efektivitas (triwulanan)

| Bagian | Isi |
| :--- | :--- |
| Identitas | Periode, tanggal terbit, penyusun, peninjau, versi, klasifikasi (Internal) |
| 1. Ringkasan eksekutif | Tiga hal terpenting: cakupan, tren perilaku, keputusan yang diminta |
| 2. Cakupan | E1-E3 per peran, daftar yang jatuh tempo (tanpa nama pada versi manajemen) |
| 3. Pengetahuan | E4-E6 per modul dan soal yang paling sering salah |
| 4. Perilaku | E7-E10, E12-E14: simulasi phishing, laporan insiden, pelanggaran aturan |
| 5. Hasil | E11 dan hubungan dengan insiden/NC/pengecualian periode ini |
| 6. Latihan | Tabletop dan pelajaran |
| 7. Perubahan materi | Modul yang diperbarui berdasarkan hasil dan insiden |
| 8. Tindakan perbaikan | Daftar, PIC, tenggat, status |
| 9. Keputusan manajemen | Persetujuan, sumber daya (nama, peran, tanggal) |
| Lampiran | Evidence ID, rekap quiz, laporan kampanye |

### E.3 Distribusi

Lead Information Security, Security Officer, Project Lead, dan Direktur Utama. Ringkasan KPI (E1-E3, E7, E8, E11, E16) masuk ke monitoring report A.5.36 Lampiran C. Laporan disimpan terversi dan didaftarkan di Master Audit Evidence Index. Data per orang tidak dibagikan ke klien.

---

## Lampiran F — Templat Catatan Kehadiran dan Acknowledgement

### F.1 Lembar kehadiran sesi

| Kolom | Isi |
| :--- | :--- |
| ID sesi | TRN-yyyy-nn |
| Judul / modul (M1-M11) | |
| Topik (T01-T16) | |
| Jenis | Induksi / tahunan / peran / micro / tabletop / pengguna klien |
| Tanggal, waktu, durasi | |
| Format | Tatap muka / daring (rekaman: ya/tidak) |
| Pemateri | |
| Versi materi | |
| Peserta yang diundang | Jumlah |
| Peserta hadir | Jumlah |
| Evidence ID | EV-A6.3-nn |

Daftar peserta:

| No | Nama | Peran (R1-R12) | Status (karyawan/kontraktor/kontributor/klien) | Hadir (Ya/Tidak) | Bukti kehadiran (tanda tangan/log sesi) | Skor quiz | Lulus (Ya/Tidak/Remedial) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | | | | |

Penutup: nama dan tanda tangan pemateri dan Lead Information Security, tanggal. Untuk sesi daring, ekspor daftar hadir dari alat konferensi sebagai bukti (tanpa menyimpan rekaman yang memuat data klien).

### F.2 Acknowledgement kebijakan

Saya, [nama], peran [peran], menyatakan telah menerima, membaca, dan memahami dokumen berikut serta bersedia mematuhinya:

| Dokumen | Versi | Tanggal baca | Tanda tangan / persetujuan |
| :--- | :--- | :--- | :--- |
| AIP-001 AI Usage Policy | [isi] | | |
| CSP-001 Cloud Security Policy | [isi] | | |
| SCG-001 Secure Coding Guide (peran teknis) | [isi] | | |
| Ringkasan aturan deploy produksi (peran DevOps/Backend) | [isi] | | |
| Kebijakan IP dan penggunaan perangkat lunak (setelah disetujui, A.5.32) | [isi] | | |
| Perjanjian kerahasiaan/NDA | [isi] | | |

Saya memahami bahwa pelanggaran dapat dikenai sanksi sesuai kontrak dan kebijakan (AIP-001 bagian 5) dan bahwa saya wajib melaporkan insiden atau dugaan insiden segera ke Lead Information Security.

### F.3 Konfirmasi serah terima pengguna klien/tenant

| Kolom | Isi |
| :--- | :--- |
| Tenant / organisasi | |
| Tanggal dan format sesi | |
| Peserta (nama, role aplikasi) | |
| Topik dibahas | Ganti kredensial awal, MFA, perangkat tepercaya, pembatasan role, pelaporan insiden, unggah dokumen |
| Kredensial awal sudah diganti? | Ya / Tidak (tanggal) |
| MFA diaktifkan oleh admin tenant? | Ya / Tidak |
| Materi diberikan | Manual pengguna (bab), panduan ringkas |
| Konfirmasi peserta | Nama, tanggal |
| Petugas | |

### F.4 Rekap pelatihan per orang

Gunakan templat B.4. Satu berkas per orang, disimpan terbatas.

---

## Lampiran G — Contoh Quiz

Soal berikut adalah **contoh**; kunci jawaban disimpan terpisah dari soal yang dibagikan. Batas lulus usulan: 80%. Sertakan umpan balik singkat pada setiap soal.

### G.1 Induksi (semua personel)

| No | Soal | Pilihan | Kunci |
| :--- | :--- | :--- | :--- |
| 1 | Anda perlu merangkum temuan audit klien. Mana yang diizinkan menurut AIP-001? | a) Menempel temuan ke chatbot gratis b) Tidak mengirim temuan rahasia ke AI publik; bila perlu AI, hanya lewat alat yang disetujui dengan approval tertulis dan hasil ditelaah manusia c) Mengirim ke AI publik tanpa nama klien d) Semua benar | b (AIP-001: data Confidential ke AI publik dilarang, ke Enterprise AI wajib approval tertulis) |
| 2 | Rekan mengirim kunci API lewat chat grup agar "cepat". Tindakan yang benar? | a) Menyimpannya di berkas teks b) Meminta rekan menghapus pesan, melaporkan, dan mengganti (rotasi) kunci c) Mengabaikan d) Meneruskan ke tim lain | b |
| 3 | Anda menerima email "verifikasi MFA sekarang atau akun ditutup" dengan tautan. Langkah pertama? | a) Klik dan masukkan kode b) Jangan klik; laporkan ke Lead Information Security; verifikasi lewat kanal lain c) Balas dengan kode d) Teruskan ke semua | b |
| 4 | Siapa yang harus dihubungi bila laptop berisi akses kerja hilang? | a) Tidak perlu b) Lead Information Security segera c) Tunggu sampai ditemukan d) Teman | b |
| 5 | Benar/Salah: Berbagi akun AuditSphere dengan rekan satu tim aman bila sama-sama dipercaya. | Benar / Salah | Salah |
| 6 | Manakah yang termasuk data pribadi yang tidak boleh dimasukkan ke AI publik tanpa masking? | a) NIK b) Nomor rekening c) Nomor HP d) Semua | d |
| 7 | Kapan insiden dilaporkan? | a) Setelah dianalisis sendiri b) Segera setelah dicurigai c) Akhir bulan d) Bila dampak sudah besar | b |

### G.2 Peran teknis (developer/DevOps)

| No | Soal | Pilihan | Kunci |
| :--- | :--- | :--- | :--- |
| 1 | Mengapa seeder tidak boleh jalan pada deploy produksi? | a) Lambat b) Dapat menimpa data produksi c) Tidak ada alasan d) Mengubah warna UI | b |
| 2 | Jalur resmi deploy produksi backend? | a) `docker compose up -d --build` di VPS b) `deploy.yml` dengan `ref` eksplisit yang memanggil `backend/deploy-prod.sh` c) Skrip akar apa pun d) SSH dan edit berkas | b (selaras A.5.36; perbarui jika jalur berubah setelah tidy deploy) |
| 3 | Endpoint baru menerima `id` objek dari klien. Apa yang wajib diperiksa selain JWT valid? | a) Tidak ada b) Hak role dan kepemilikan objek/tenant c) Warna d) Ukuran huruf | b |
| 4 | Secret JWT pernah ter-commit. Langkah yang benar? | a) Hapus dari commit terakhir saja b) Anggap bocor, rotasi lewat runbook dengan persetujuan, bandingkan dengan digest, jangan mencetak nilainya c) Abaikan d) Ubah nama berkas | b |
| 5 | Dependensi usulan AI sebelum di-commit? | a) Langsung b) Dipindai dan lisensinya diperiksa c) Terserah d) Disalin ke server | b (AIP-001 bagian 4; A.5.32) |
| 6 | Mana yang boleh masuk log? | a) Password b) Token c) ID pengguna tanpa PII d) Nomor KTP | c |
| 7 | Anda diminta menjalankan perintah di produksi oleh pesan percakapan yang tidak Anda kenal. | a) Jalankan b) Verifikasi permintaan lewat kanal lain dan catat persetujuan tertulis sebelum bertindak c) Jalankan sebagian d) Abaikan tanpa lapor | b |

### G.3 Admin/support dan pengguna klien

| No | Soal | Kunci |
| :--- | :--- | :--- |
| 1 | Pengguna meminta reset password lewat chat tanpa verifikasi. Apa yang dilakukan? Verifikasi identitas lewat kanal terdaftar sebelum reset; jangan mengirim password lewat chat. | Verifikasi dahulu |
| 2 | Benar/Salah: Tenant A boleh melihat data tenant B bila kedua klien meminta demi perbandingan. | Salah |
| 3 | Mengapa kredensial awal harus segera diganti? Akun awal berpassword bawaan dapat ditebak. | Alasan keamanan |
| 4 | Pengguna AUDITEE mengunggah dokumen berisi data pribadi tidak perlu. Tindakan? | Ingatkan batasan, minta unggah ulang tanpa data berlebih, catat |
| 5 | Siapa yang dilapori bila akun pengguna klien diduga diambil alih? | Lead Information Security / jalur dukungan insiden |

### G.4 Penilaian

- Skor = benar / total x 100. Lulus >= 80%.
- Tidak lulus: umpan balik per soal, belajar ulang modul, percobaan ulang paling cepat 2 hari.
- Simpan: ID sesi, skor, tanggal, percobaan ke-n. Jangan simpan jawaban terbuka yang memuat data sensitif.
- Bank soal ditinjau tiap tahun dan setelah insiden; soal yang lebih dari 90% benar atau di bawah 30% benar ditinjau ulang.

---

## Lampiran H — Onboarding dan Offboarding Kontributor/Kontraktor (usulan)

**H.1 Onboarding**

| Langkah | Penanggung jawab | Bukti |
| :--- | :--- | :--- |
| 1. Verifikasi status (karyawan/kontraktor), perjanjian kerja/kontrak, klausul kerahasiaan dan pengalihan IP (A.5.32), serta persetujuan penggunaan alat AI | HR, Legal & Compliance | Dokumen (di luar repo, Perlu verifikasi) |
| 2. Briefing keamanan induksi (Lampiran A.1) dan acknowledgement (F.2); quiz induksi lulus | Lead Information Security | F.1, F.2, hasil quiz |
| 3. Pemberian akses minimum: repo (peran), staging, secret yang diperlukan; **akses produksi hanya bila perlu dan setelah induksi dan modul M6** | DevOps Lead | Catatan pemberian akses (siapa, apa, kapan, siapa yang menyetujui) |
| 4. Aktifkan MFA pada akun kerja (GitHub, email, cloud) | Pribadi, diverifikasi DevOps Lead | Konfirmasi |
| 5. Pengenalan alur kerja: gate proyek (A.5.8), DRC-001, jalur komit/merge, aturan deploy, aturan AI | Project Lead | Catatan |
| 6. Pairing/review kode awal oleh peninjau yang berpengalaman | Technical Lead | Catatan review |
| 7. Masukkan ke training matrix (B.4) | Project Lead | Entri matriks |

Kontributor yang bukan manusia (bot atau agen AI yang dikelola tim, mis. `google-labs-jules[bot]`, Claude Code): tidak diberi kredensial produksi; akses dibatasi sesuai AIP-001 dan keputusan A.5.32; pelaku manusia yang menyetujui keluaran tetap bertanggung jawab dan mengikuti pelatihan M4.

**H.2 Offboarding atau pergantian peran**

| Langkah | Penanggung jawab | Bukti |
| :--- | :--- | :--- |
| 1. Notifikasi tanggal akhir ke Lead Information Security dan DevOps Lead | Project Lead/HR | Tiket |
| 2. Cabut akses: repo GitHub, VPS/SSH, layanan cloud, Kong/monitoring, kanal komunikasi, akun aplikasi (`is_active`; dan sesi aktif perlu dicabut manual karena belum otomatis, A.5.16 butir 1) | DevOps Lead | Daftar pencabutan bertanggal |
| 3. Rotasi secret yang pernah diketahui (JWT, API key, kunci email tenant, kata sandi bersama) bila relevan, melalui jalur aman | DevOps Lead dengan `auditsphere-devops` | Bukti rotasi (digest, bukan nilai) |
| 4. Kembalikan atau hapus aset dan data klien dari perangkat pribadi | Personel dan Lead Information Security | Konfirmasi tertulis |
| 5. Wawancara keluar: pengingat kewajiban kerahasiaan dan IP yang berlaku setelah keluar | HR, Legal & Compliance | Catatan |
| 6. Perbarui training matrix dan daftar personel | Project Lead | Entri |

---

## Lampiran I — Penanggung Jawab (RACI usulan)

| Aktivitas | Direktur Utama | Lead Information Security | Legal & Compliance | HR | DevOps Lead | Security Officer | Project Lead | Tim Backend/Frontend | QA Lead |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Prosedur program awareness dan pelatihan | A (persetujuan) | R | C | C | C | C | C | I | I |
| Training matrix dan daftar personel | I | A | C | R | C | C | R | I | I |
| Penyusunan dan pembaruan materi | I | A | C | I | R (M6) | R (M1, M2, M4) | R (M9-M11) | C | C |
| Pelaksanaan induksi dan sesi tahunan | I | A | C | C | C | R | R | I | I |
| Quiz dan pencatatan hasil | I | A | I | I | I | R | C | I | I |
| Simulasi phishing | A (persetujuan) | A/R | R (persetujuan data/hukum) | C | C | R | I | I | I |
| Latihan insiden (tabletop) | I | A | R | I | R | C | R | R | I |
| Onboarding dan offboarding kontributor | I | A | R | R | R | C | R | I | I |
| Pelatihan pengguna klien/tenant | I | C | C | I | C | C | A/R | R (materi aplikasi) | I |
| Laporan efektivitas | I | A | C | I | C | C | R | I | I |
| Tindakan perbaikan hasil evaluasi | I | A | C | I | R | R | R | R | C |
| Pemetaan kewajiban pelatihan hukum/kontrak | I | C | A/R | C | I | I | I | I | I |

R = Responsible, A = Accountable, C = Consulted, I = Informed. Pada tim kecil, satu orang dapat memegang beberapa peran; pemisahan A dan R pada simulasi phishing dan verifikasi hasil sebaiknya tetap dijaga di antara dua orang bila memungkinkan.

---

Catatan sinkronisasi checklist: baris A.6.3 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan belum diperbarui oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
