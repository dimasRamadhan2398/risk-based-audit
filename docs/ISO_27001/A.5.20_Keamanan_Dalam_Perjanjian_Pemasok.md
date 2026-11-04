# A.5.20 — Keamanan dalam Perjanjian Pemasok

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Keamanan dalam perjanjian pemasok |
| **Fokus pemeriksaan** | Persyaratan keamanan dimasukkan ke kontrak pemasok |
| **Contoh bukti objektif** | Kontrak, SLA, DPA, security schedule, daftar klausul |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Lead Information Security [konfirmasi] (tidak tercantum di baris checklist) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 29 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere bergantung pada pemasok untuk hosting, DNS, email transaksional, backup, kode sumber, dan CI/CD. Sebagian pemasok memproses data klien atau kredensial. Persyaratan keamanan harus ada di kontrak. |
| Implementasi status | **Sebagian** | Dari sisi teknis, pemisahan kunci email per tenant dan pengikatan host key SSH sudah ada. Dari sisi kontraktual, tidak ditemukan DPA, SLA, atau security schedule di repositori. |
| Ketersediaan bukti | **Belum tersedia** | Tidak ada kontrak, SLA, DPA, security schedule, atau daftar klausul di `docs/` maupun di `frontend/`. Dokumen `docs/ISO_27001/cloud_saas/` menyebut pemasok yang tidak sesuai dengan kondisi produksi (lihat bagian 4). |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.20-01 s/d EV-A5.20-04) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada daftar pemasok yang direview secara berkala dan tidak ada catatan operasional yang membuktikan persyaratan keamanan ditagih atau dipantau. |
| Desain kontrol | **Perlu perbaikan** | Belum ada register pemasok, belum ada daftar sub-prosesor untuk klien, dan belum ada checklist klausul kontrak yang wajib dipenuhi sebelum pemasok baru dipakai. |
| Status kesiapan | **Belum siap** | Penerapan kontraktual dan bukti belum memenuhi gate. |
| Prioritas | **Tinggi** | Email notifikasi dikirim lewat pemasok pihak ketiga, dan kebijakan `CSP-001` mewajibkan DPA untuk setiap penyedia cloud. Belum ada DPA yang bisa ditunjukkan. |
| Review berikutnya | [isi] | Usulan: 3 bulan setelah register pemasok dan klausul kontrak disetujui, lalu setiap tahun atau saat pemasok berubah. |

> Catatan pembacaan checklist: pada tab `Checklist 93 Kontrol`, baris A.5.20 kolom G s/d Z belum diisi. PIC tidak tercantum pada baris ini, sehingga dokumen ini memakai "Lead Information Security" sebagai usulan dan perlu dikonfirmasi. Status diisi dari verifikasi kode dan dokumen di repositori, bukan dari keterangan pihak luar.

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.20 **berlaku** untuk AuditSphere karena:

1. Platform bergantung pada pihak ketiga untuk pengiriman email transaksional (Resend, `backend/auth-service/services/resend/resend_service.go:20`; `backend/audit-service/pkg/mailer/mailer.go:24`), DNS dan domain pengiriman (Rumahweb, `scripts/provision-tenant-email.sh:14`, `:87-88`), backup off-site (Backblaze B2, `docs/backup-b2-setup.md:1-15`), kode sumber dan CI/CD (GitHub, `.github/workflows/deploy.yml`), serta penyimpanan media (ImageKit, `backend/audit-service/cmd/serve.go:195-200`). Lihat daftar lengkap di bagian 2.1.
2. Kebijakan internal `CSP-001` (`docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx`, butir 19 "Manajemen Vendor Cloud & Perjanjian DPA") mewajibkan DPA yang memuat klausul UU PDP No. 27/2022 sebelum PO. Persyaratan ini hanya berlaku jika bukti kontraknya ada.
3. Klien enterprise dan bank pada due diligence akan meminta daftar sub-prosesor, lokasi data, dan DPA. Tanpa dokumen tersebut, klaim keamanan platform tidak bisa dibuktikan kepada pihak ketiga.

---

## 2. Implementasi yang Sudah Ada

Setiap butir diverifikasi dari repositori. Path dan baris mengacu pada cabang `staging`. Tidak ada nilai rahasia (kunci API, token) yang dicantumkan di dokumen ini.

### 2.1 Inventaris pemasok yang terlihat di repositori

| Pemasok / layanan | Fungsi di AuditSphere | Bukti di repositori | Kontrak / DPA / SLA di repositori | Lokasi data dan catatan | Status verifikasi |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Penyedia VPS (nama dan lokasi [isi]) | Host produksi: Docker Compose, Nuxt, Go services, Postgres, Redis, Kafka, Kong | `docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:149-156` (rekomendasi umum, bukan pilihan final); `.github/workflows/deploy.yml:63-76` (SSH ke VPS) | Tidak ada | Lokasi data center tidak tercatat | **Terverifikasi (fungsi), penyedia belum tercatat** |
| Resend | Email transaksional (notifikasi, reset, pengingat ATR, dan email tenant) | `backend/auth-service/services/resend/resend_service.go:20`; `backend/audit-service/pkg/mailer/mailer.go:24`; `scripts/provision-tenant-email.sh:1-30`, `:55` (`RESEND_REGION` default `ap-northeast-1`) | Tidak ada | Region pengiriman diatur ke `ap-northeast-1` secara default. Lokasi pemrosesan akhir dan sub-prosesor Resend belum diverifikasi. | **Terverifikasi (fungsi)** |
| SMTP relay Resend (fallback) | Pengiriman SMTP bila API tidak dipakai | `backend/auth-service/.env.example:34-39`; `backend/shared/email/smtp.go` | Tidak ada | Sama dengan Resend | **Terverifikasi (konfigurasi contoh)** |
| Rumahweb | Registrar domain dan DNS untuk `mail.auditsphere.app` (SPF, DKIM, DMARC) | `scripts/provision-tenant-email.sh:14`, `:87-88` (hanya komentar dan instruksi) | Tidak ada | Tidak ada catatan lokasi atau sub-prosesor | **Terverifikasi (dari komentar skrip)** |
| Cloudflare | WAF/DDoS di edge (rencana) | `docs/DOKUMEN_KESELARASAN_ISO27001.md:401`; `docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md:34`, `:95`; `CSRM-001` dan `CPA-001` (teks dokumen `cloud_saas/`, bagian whitelisting WAF) | Tidak ada | Tidak ada konfigurasi di `backend/` atau `scripts/` | **Hanya dokumen, belum terverifikasi** |
| Backblaze B2 | Penyimpanan backup off-site (Postgres dump dan arsip upload) | `docs/backup-b2-setup.md:1-15` (status **DRAFT, belum terpasang**, cron belum ditambahkan) | Tidak ada | Bucket privat `auditsphere-prod-backups`, retensi 30 hari (rencana) | **Rencana, belum berjalan** |
| ImageKit | Penyimpanan media (opsional, aktif bila env `IMAGEKIT_*` terisi) | `backend/audit-service/cmd/serve.go:195-200`; `backend/audit-service/pkg/config/config.go:107-112`; `backend/docker-compose.prod.yml:172-174`; `backend/audit-service/pkg/media/imagekit.go` | Tidak ada | Lokasi penyimpanan ImageKit belum tercatat. Tidak ada di inventaris `docs/`. | **Terverifikasi (kode), status aktif di produksi [konfirmasi]** |
| Google Drive | Penyimpanan media alternatif (dinonaktifkan) | `backend/audit-service/pkg/config/config.yaml:56-57` (`gdrive.enabled: false`); `backend/audit-service/cmd/serve.go` | Tidak ada | Jika diaktifkan, file media akan disimpan di layanan Google. Tujuan dan lokasi penyimpanan perlu dicatat sebelum diaktifkan. | **Kode ada, nonaktif** |
| OpenAI | Rute Nuxt `/api/openai` (model `gpt-3.5-turbo-instruct`) | `frontend/server/api/openai.ts:1-30`; `frontend/package.json:36` (`openai`); `docs/api/openapi.yaml:1198-1222` | Tidak ada | Prompt dikirim ke layanan eksternal. `OPENAI_API_KEY` tidak ada di compose atau `.env.example`. Tidak ada pemanggil dari halaman `frontend/pages` atau `components` (pencarian `api/openai`). Risiko PII belum diatur (lihat A.8.11 di checklist). | **Orphan (kode ada, tidak dipanggil dari UI)** |
| GitHub dan GitHub Actions | Hosting repositori, CI/CD, deploy ke VPS lewat SSH | `.github/workflows/deploy.yml:44-76` (secret `VPS_SSH_KEY`, `VPS_KNOWN_HOSTS`); `.github/workflows/deploy-dev.yml` | Tidak ada | Kode dan secret CI berada di GitHub | **Terverifikasi** |
| Titan (mail) | Layanan email perusahaan (peran belum diketahui) | Tidak ditemukan di repositori (pencarian `titan` nihil) | Tidak ada | Tidak tercatat | **Tidak ditemukan. Konfirmasi peran dan keberadaannya [konfirmasi]** |
| Google Search Console | Pemantauan indeks SEO landing page | Tidak ditemukan di repositori (pencarian `search console` nihil) | Tidak ada | Tidak tercatat | **Tidak ditemukan. Konfirmasi [konfirmasi]** |
| Kafka, Redis, Postgres, Kong, ClamAV | Layanan inti dijalankan sendiri di container VPS (image pihak ketiga; ClamAV memakai digest tetap) | `backend/docker-compose.prod.yml:5` (postgres:16), `:19` (redis:7-alpine), `:51` (confluentinc/cp-kafka), `:146` (clamav digest) | Tidak perlu (lisensi open source) | Data tetap di VPS. Image pihak ketiga tetap menjadi rantai pasok (lihat A.5.21). | **Self-hosted, bukan pemasok data** |
| Python AI (`backend/python-ai`) | Inferensi model lokal (FastAPI, model di `ai_model_training/`) | `backend/python-ai/main.py`, `requirements.txt` | Tidak perlu | Tidak ditemukan pemanggilan LLM eksternal di `backend/python-ai` | **Self-hosted (model lokal)** |

### 2.2 Pengendalian teknis terkait pemasok yang sudah ada

- **Kunci email per tenant dengan hak kirim saja.** Setiap tenant memakai API key sendiri dengan hak kirim saja. Kunci master tidak dipakai tenant, dan kunci tenant bisa dicabut tanpa mengganggu tenant lain (`scripts/provision-tenant-email.sh:3-10`). Kunci master dibaca dari file root-only di `/etc/auditsphere/email.env` (`scripts/provision-tenant-email.sh:23-30`, bukan dari repo).
- **Pembuatan domain dan kunci Resend tidak lewat API aplikasi.** Auth-service hanya memasang rute masuk webhook Resend (`POST /resend/webhook`, `backend/auth-service/routes/registry.go:208-212`). Pembuatan domain dan kunci per tenant dilakukan lewat skrip di VPS (`scripts/provision-tenant-email.sh`). Dokumen panduan juga menyebut hal ini (`docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md:334-337`, `:498-499`).
- **Host key SSH dipin di CI.** Deploy menolak host key yang tidak dipin (`.github/workflows/deploy.yml:63-70`).
- **Secret tidak masuk repositori.** `backend/.env` dan `.env.example` memakai placeholder (`re_xxx...`). Nilai asli tidak ditampilkan di dokumen ini.
- **Kelemahan yang perlu diperbaiki.** `ResendConfig.GetAPIKey()` jatuh ke `RESEND_MASTER_API_KEY` bila `RESEND_API_KEY` kosong (`backend/auth-service/pkg/config/config.go:132-146`). Akibatnya, kunci master dengan hak penuh bisa dipakai untuk mengirim email. Kunci ini seharusnya hanya dipakai skrip provisioning.

### 2.3 Dokumen kebijakan yang sudah ada (belum berisi klausul kontrak)

- `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx`: butir 19 "Manajemen Vendor Cloud & Perjanjian DPA" (wajib pre-adoption assessment melalui SAP-001, wajib DPA dengan klausul UU PDP No. 27/2022).
- `docs/ISO_27001/cloud_saas/CPA-001_Cloud_Provider_Assessment.docx`: kriteria penilaian vendor, termasuk sub-prosesor, residensi data, dan SLA ≥ 99,9%.
- `docs/ISO_27001/cloud_saas/CSRM-001_Cloud_Shared_Responsibility_Matrix.docx`: pembagian tanggung jawab. Isinya mengacu pada AWS, GCP, dan Azure (lihat gap 2).
- `docs/ISO_27001/cloud_saas/CES-001_Cloud_Exit_Strategy.docx`: strategi keluar dari vendor (lihat gap 2).
- `docs/ISO_27001/cloud_saas/Cloud_Security_and_Governance_Checklist.docx`: daftar periksa, termasuk "SLA kontrak CSP, DPA, laporan SOC 2 Type II, sertifikat ISO 27001/27017, Cloud Exit Plan" (bagian matriks bukti). Ini daftar yang perlu dipenuhi, bukan bukti bahwa sudah dipenuhi.
- `docs/DOKUMEN_KESELARASAN_ISO27001.md`: tidak merujuk ke dokumen `cloud_saas/` dan tidak membahas A.5.19 s/d A.5.21 secara eksplisit.

### 2.4 Pemberitahuan privasi kepada pengguna

- Halaman `frontend/pages/auth/confidentiality.vue:462` berisi pernyataan kerahasiaan dari pengguna, tanpa daftar sub-prosesor atau pemberitahuan lokasi data.
- Tidak ditemukan halaman privacy atau terms di `frontend/pages/` maupun `frontend/layouts/`.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kontrak pemasok (tertanda) | **Belum tersedia** | Tidak ada di repositori. Dokumen `cloud_saas/` memakai placeholder "[Ditandatangani Digital]" dan tidak merujuk ke kontrak nyata. |
| SLA pemasok | **Belum tersedia** | Tidak ada SLA Resend, Backblaze, ImageKit, Rumahweb, atau penyedia VPS. `CPA-001` menulis target SLA ≥ 99,9%, tetapi itu kriteria penilaian, bukan kontrak. |
| DPA (Data Processing Agreement) | **Belum tersedia** | Tidak ditemukan DPA untuk pemasok mana pun. `CSP-001` mewajibkan DPA, tetapi belum ada yang ditandatangani. |
| Security schedule / lampiran keamanan | **Belum tersedia** | Tidak ada. |
| Daftar klausul kontrak | **Belum tersedia** | Tidak ada. Daftar periksa di `Cloud_Security_and_Governance_Checklist.docx` bisa dipakai sebagai dasar. |
| Register pemasok dan sub-prosesor | **Belum tersedia** | Inventaris di bagian 2.1 adalah hasil verifikasi kode, bukan register resmi. |
| Penilaian pemasok (SAP-001 / CPA-001 terisi) | **Belum tersedia** | Template `CPA-001` ada, tetapi hasil penilaian untuk pemasok nyata tidak ditemukan. |
| Bukti konfigurasi per tenant (kunci Resend terpisah) | **Sebagian** | Skrip dan kode ada (`scripts/provision-tenant-email.sh`). Belum ada contoh keluaran atau catatan rotasi. |
| Bukti uji restore backup dari B2 | **Belum tersedia** | Skrip ada (`backend/scripts/backup/restore-test.sh`), tetapi `docs/backup-b2-setup.md` masih DRAFT dan belum terpasang di VPS. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada dokumen kontrak yang bisa ditunjukkan.** Tidak ditemukan file DPA, kontrak, SLA, atau security schedule di `docs/`, `scripts/`, `backend/`, maupun `frontend/` (pencarian nama file `*dpa*`, `*sla*`, `*contract*`, `*kontrak*`, `*vendor*`, `*supplier*`, `*pemasok*`, `*privacy*`, `*terms*`, `*subprocessor*` nihil, kecuali template di `docs/ISO_27001/cloud_saas/`).
2. **Dokumen `cloud_saas/` tidak sesuai dengan kondisi produksi.** `CSP-001` (butir "Data Residency" dan "Vendor"), `CSRM-001`, `CES-001`, dan `CPA-001` menyebut AWS (EKS, RDS, S3, CloudTrail, KMS), GCP (Vertex AI, Cloud SQL), Azure, HashiCorp Vault, Datadog, Sentry, SendGrid, Jira, Slack, GitHub Enterprise, dan Google Workspace. Hanya GitHub yang terlihat di repositori. Produksi memakai VPS, Resend, Rumahweb, B2, dan ImageKit. Dokumen ini harus direvisi atau diberi catatan bahwa isinya masih template sebelum dipakai sebagai bukti A.5.20 atau A.5.23.
3. **Pemasok nyata belum punya register, lokasi data, atau sub-prosesor yang tercatat.** Resend memproses alamat email dan isi notifikasi dengan region default `ap-northeast-1` (`scripts/provision-tenant-email.sh:55`). Tidak ada dokumen yang menjelaskan apakah data ini boleh keluar dari Indonesia, siapa sub-prosesor Resend, atau bagaimana klien diberi tahu. ImageKit (jika aktif) dan OpenAI (rute orphan di `frontend/server/api/openai.ts`) juga belum tercatat. Rute OpenAI menerima `body.prompt` dan mengirimnya ke pihak ketiga. Dokumen `A.8.11` dan `A.8.23` di checklist sudah mencatat risiko ini, tetapi belum ada tindak lanjut.
4. **Halaman untuk pengguna tidak menyebut sub-prosesor atau lokasi data.** `frontend/pages/auth/confidentiality.vue:462` hanya berisi pernyataan kerahasiaan. Tidak ada halaman privacy atau terms yang mengungkap pemasok.
5. **Titan dan Google Search Console tidak ditemukan di repositori.** Pencarian `titan` dan `search console` nihil. Jika keduanya dipakai, mereka perlu masuk inventaris. Jika tidak, keterangan ini perlu dihapus dari daftar pemasok di dokumen internal.
6. **Rumahweb dan Cloudflare hanya muncul di dokumen dan komentar skrip.** Konfigurasi aktual (DNS, WAF) tidak ada di repositori. Peran dan kontrak keduanya perlu dikonfirmasi.
7. **Backup off-site belum berjalan.** `docs/backup-b2-setup.md` menulis "Status: DRAFT. Not installed on the VPS." Artinya backup Backblaze belum menjadi bukti yang bisa dipakai. Gunakan keterangan ini saat mengisi A.8.13.
8. **Pemasok VPS belum tercatat.** Dokumen demo menyebut beberapa pilihan (AWS, DigitalOcean, Linode, Hetzner) sebagai rekomendasi. Penyedia yang dipakai produksi tidak tercatat di repositori. Isi nama dan lokasi data center di bagian 2.1 setelah dikonfirmasi.
9. **Kunci master Resend masih bisa dipakai untuk mengirim email.** Jika `RESEND_API_KEY` kosong, `GetAPIKey()` memakai `RESEND_MASTER_API_KEY` (`backend/auth-service/pkg/config/config.go:132-146`). Kunci master ini seharusnya hanya dipakai skrip provisioning dan memiliki hak penuh atas akun Resend. Tindakan 12.
10. **Tidak ada proses penilaian atau tinjauan berkala pemasok.** Tidak ada catatan tinjauan, jadwal, atau hasil penilaian pemasok yang terlihat di repositori. Tinjauan ini termasuk bagian A.5.19 dan A.5.22.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Konfirmasi daftar pemasok produksi (VPS, Resend, Rumahweb, Backblaze B2, ImageKit, GitHub, dan status Titan, Cloudflare, Google Search Console, Google Drive, OpenAI). Catat nama legal, lokasi data, dan peran di register pemasok. | Lead Information Security [konfirmasi] bersama Tim Backend dan auditsphere-devops | Tinggi | [isi] | Register pemasok yang disetujui |
| 2 | Kumpulkan DPA, kontrak, dan SLA yang sudah ditandatangani untuk setiap pemasok di register. Simpan di lokasi dengan akses terbatas dan catat tanggal kedaluwarsa. | Lead Information Security [konfirmasi] bersama Legal | Tinggi | [isi] | Salinan DPA dan kontrak (akses restricted), Evidence ID |
| 3 | Tetapkan daftar klausul keamanan wajib untuk pemasok baru (DPA sesuai UU PDP No. 27/2022, lokasi data, notifikasi insiden, hak audit, sub-prosesor, pengembalian dan penghapusan data, SLA ketersediaan). Pakai `Cloud_Security_and_Governance_Checklist.docx` sebagai dasar. | Lead Information Security [konfirmasi] | Tinggi | [isi] | Daftar klausul keamanan yang disetujui |
| 4 | Revisi `CSP-001`, `CSRM-001`, `CES-001`, dan `CPA-001` agar memuat pemasok nyata. Tandai bagian yang masih template. Hapus klaim AWS, GCP, Azure, Vault, Datadog, dan Sentry yang tidak dipakai. | Lead Information Security [konfirmasi] | Tinggi | [isi] | Revisi dokumen `cloud_saas/` dengan tanggal dan pemilik |
| 5 | Putuskan nasib rute OpenAI (`frontend/server/api/openai.ts`). Jika tidak dipakai, hapus rute dan dependensi `openai` setelah koordinasi dengan frontend. Jika dipakai, tambahkan persyaratan PII scrubbing dan DPA OpenAI. | Tim Frontend bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Keputusan tertulis dan perubahan kode |
| 6 | Catat region dan sub-prosesor Resend, serta putuskan apakah email klien yang berisi data audit boleh diproses di luar Indonesia. Tambahkan keterangan di pemberitahuan privasi. | Lead Information Security [konfirmasi] bersama Platform | Tinggi | [isi] | Catatan lokasi data Resend dan keputusan residensi |
| 7 | Tambahkan pengumuman sub-prosesor dan lokasi data di halaman privacy atau terms (ID dan EN), dan tautkan dari `confidentiality.vue`. Buat halaman privacy jika belum ada. | auditsphere-frontend bersama auditsphere-software-translation | Sedang | [isi] | Halaman privacy dengan daftar sub-prosesor |
| 8 | Lengkapi status ImageKit dan Google Drive di produksi (aktif atau nonaktif). Jika aktif, masukkan ke register pemasok dan kontrak. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Sedang | [isi] | Hasil cek env produksi dan entri register |
| 9 | Jalankan backup B2 (dari `docs/backup-b2-setup.md`) dan catat DPA atau kontrak Backblaze sebelum data klien dikirim. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Log backup off-site dan hasil restore test |
| 10 | Tetapkan jadwal tinjauan pemasok (minimal tahunan dan saat ada perubahan) dengan checklist. Catat hasilnya. | Lead Information Security [konfirmasi] | Sedang | [isi] | Berita acara tinjauan pemasok |
| 11 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security [konfirmasi] | Sedang | [isi] | Entri Evidence Index |
| 12 | Hapus fallback `GetAPIKey()` ke `RESEND_MASTER_API_KEY` (`backend/auth-service/pkg/config/config.go:132-146`). Jalur kirim email aplikasi hanya memakai kunci dengan hak kirim saja. | Tim Backend (auth-service) bersama auditsphere-devops | Tinggi | [isi] | Perubahan kode dan uji bahwa pengiriman gagal tanpa kunci kirim |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.20-01 | Inventaris pemasok hasil verifikasi kode (bagian 2.1) | Dokumen ini, bagian 2.1 | Internal |
| EV-A5.20-02 | Pengendalian teknis terkait pemasok: kunci Resend per tenant, pinning host key CI, dan pemisahan rute Resend | `scripts/provision-tenant-email.sh`, `backend/auth-service/services/resend/resend_service.go`, `backend/auth-service/pkg/config/config.go`, `.github/workflows/deploy.yml` | Internal |
| EV-A5.20-03 | Dokumen kebijakan dan checklist vendor cloud (CSP-001, CPA-001, CSRM-001, CES-001, Cloud_Security_and_Governance_Checklist) yang belum sesuai produksi | `docs/ISO_27001/cloud_saas/` | Internal |
| EV-A5.20-04 | DPA, kontrak, SLA, dan security schedule (belum tersedia) | Belum tersedia. Akan dikumpulkan dari tindakan 2 dan 3. | Restricted |

Setelah tindakan 1, 2, dan 3 selesai, tambahkan ID untuk register pemasok, DPA yang ditandatangani, dan daftar klausul keamanan.
