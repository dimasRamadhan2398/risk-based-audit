# A.5.32 — Hak Kekayaan Intelektual

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Hak kekayaan intelektual |
| **Fokus pemeriksaan** | Penggunaan perangkat lunak dan konten mematuhi hak kekayaan intelektual |
| **Contoh bukti objektif** | License register, bukti pembelian, software inventory, audit lisensi |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Legal & Compliance) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 30 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Kolom "Lisensi" pada Lampiran B berstatus **Terverifikasi** hanya bila dibaca dari berkas paket yang terpasang (`frontend/node_modules/<paket>/package.json`); selain itu ditandai **Perlu verifikasi** (berdasarkan pengetahuan umum tentang paket, belum dipindai alat).

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere dibangun di atas ratusan komponen pihak ketiga, dioperasikan sebagai SaaS multi-tenant, dan menyimpan konten milik klien; kepatuhan lisensi dan kepemilikan IP wajib dikelola. |
| Implementasi status | **Belum diimplementasikan** | Tidak ditemukan kebijakan IP, license register, software inventory/SBOM, berkas `LICENSE`/`NOTICE`, maupun pemindaian lisensi di CI. Inventaris di Lampiran B adalah hasil rekonstruksi pertama dari berkas manifes, bukan register resmi. |
| Ketersediaan bukti | **Ada sebagian** | Manifes dependensi (`package.json`, `go.mod`, `requirements.txt`, berkas compose/Dockerfile) ada dan dapat menjadi dasar inventory; bukti pembelian/langganan, license register, dan catatan audit lisensi belum ada. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.32-01 s/d EV-A5.32-07) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum efektif** | Belum ada proses berulang (audit lisensi berkala, persetujuan sebelum menambah dependensi). Sebagian besar dependensi langsung berlisensi permisif, tetapi ada temuan lisensi yang perlu keputusan (lihat bagian 4). |
| Desain kontrol | **Belum dirancang** | Draf ketentuan kebijakan dan prosedur audit lisensi diusulkan pada Lampiran A dan C; belum disetujui. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Klien lembaga jasa keuangan akan meminta bukti kepatuhan lisensi pada due diligence; tiga temuan berisiko (GPL pada Data Hub, lisensi source-available pada Redis/TimescaleDB/RedisInsight, dan layanan OpenAI tanpa kontrak tercatat) perlu keputusan segera. |
| Review berikutnya | [isi] | Usulan: 12 bulan setelah kebijakan IP disetujui; audit lisensi penuh setiap 6 bulan dan pemindaian otomatis pada setiap rilis (Lampiran C). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.32 **berlaku** untuk AuditSphere karena:

1. Produk berdiri di atas banyak komponen pihak ketiga: 874 paket Node terpasang di `frontend/node_modules` (22 dependensi langsung dan 11 dependensi pengembangan di `frontend/package.json`), enam modul Go (`backend/*/go.mod`), tiga layanan Python di Data Hub (`data-server/*/requirements.txt`), `backend/python-ai/requirements.txt`, serta belasan image Docker (Postgres, Redis, Kafka, Kong, ClamAV, TimescaleDB, Grafana, dan lainnya). Setiap komponen membawa syarat lisensi yang harus dipenuhi.
2. AuditSphere disajikan sebagai SaaS multi-tenant dan sebagai stack per tenant (`scripts/generate-tenant-stack.sh`). Selama perangkat lunak hanya dijalankan di server milik kami, kewajiban copyleft kuat (GPL) umumnya tidak terpicu; namun skema pengiriman on-premises ke klien (dibahas di `docs/architecture_on_premises_vs_cloud.md`) dan paket Data Hub di server klien dapat dianggap **distribusi** sehingga kewajiban lisensi berubah. Perlu penilaian hukum.
3. Sistem memproses konten milik klien (kertas kerja KKA, LHA, data transaksi lewat Data Hub) dan konten standar profesi (IPPF/GIAS dari The IIA, istilah PKAT, Surat Tugas, KKA, LHA, ATR). Kepemilikan konten klien dan batas penggunaan konten standar harus jelas.
4. Tim memakai alat bantu AI dalam pengembangan (riwayat git memuat 14 komit oleh `google-labs-jules[bot]`, dan alur kerja memakai agen Claude Code) dan fitur AI/LLM di produk (`nuxt-chatgpt`, `openai`, `transformers`, IndoBERT). Status hak cipta dan syarat layanan keluaran AI harus ditetapkan.
5. Kontributor kode lebih dari satu orang (riwayat git: 6 nama penulis, termasuk bot). Pengalihan hak kekayaan intelektual dari karyawan dan kontraktor belum terbukti.
6. Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) memetakan A.5.1 s/d A.5.8 sebagai kelompok tata kelola; A.5.32 melengkapi kelompok yang sama (kepatuhan hukum dan kontraktual) dan terhubung ke A.5.20 (perjanjian pemasok) serta A.5.12 (klasifikasi informasi).

---

## 2. Implementasi yang Sudah Ada

### 2.1 Kebijakan dan prosedur

Tidak ditemukan kebijakan IP, kebijakan penggunaan perangkat lunak berlisensi, atau prosedur audit lisensi di `docs/`. Dokumen terkait yang ada hanya bersifat tidak langsung:

| Dokumen | Relevansi terhadap A.5.32 | Lokasi |
| :--- | :--- | :--- |
| AIP-001 AI Usage Policy & Acceptable Use Guidelines v1.0 | Aturan penggunaan AI; **perlu diperiksa** apakah memuat ketentuan kepemilikan/keluaran AI dan larangan memasukkan data klien ke layanan AI publik | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx` |
| CSP-001 Cloud Security Policy v1.0 | Tata kelola penyedia cloud; dasar untuk daftar pemasok | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` |
| CPA-001 Cloud Provider Assessment | Penilaian penyedia cloud; dapat dirujuk untuk bukti kontrak/langganan | `docs/ISO_27001/cloud_saas/CPA-001_Cloud_Provider_Assessment.docx` |
| Confidentiality / Integrity Pact (halaman `frontend/pages/auth/confidentiality.vue`) | Pengguna menyetujui kerahasiaan; tidak mengatur kepemilikan IP | `frontend/pages/auth/confidentiality.vue` |

### 2.2 Manifes dependensi (dasar software inventory)

| Ekosistem | Sumber | Catatan |
| :--- | :--- | :--- |
| Node/Nuxt (frontend) | `frontend/package.json`, `frontend/package-lock.json`, `frontend/bun.lock` | Dua berkas kunci (npm dan bun) ada bersamaan; pilih satu sebagai sumber SBOM. Lisensi paket langsung dapat dibaca dari `node_modules`. |
| Go (backend) | `backend/{auth,audit,master,risk,analytics}-service/go.mod`, `backend/kong-gateway/go.mod`, `backend/shared/go.sum` | Versi dependensi berbeda antar service (contoh: `gin` v1.9.1 di auth/audit, v1.12.0 di master/risk/analytics; `sarama` v1.47.0 dan v1.50.2). Tidak ada `go.mod` untuk `backend/shared`. |
| Python (AI) | `backend/python-ai/requirements.txt` (versi dipin) | TensorFlow, XGBoost, transformers, scikit-learn, pandas, numpy. |
| Python (Data Hub) | `data-server/{ai-engine,caatt-engine,data-hub-api}/requirements.txt` | Memuat paket berlisensi GPL dan LGPL (lihat bagian 4). `data-hub-api` memakai rentang versi (`>=`) sehingga tidak deterministik. |
| Image container | `backend/docker-compose*.yml`, `data-server/docker-compose*.yml`, `monitoring/docker-compose.monitoring.yml`, `*/Dockerfile` | Beberapa image memakai tag mengambang (`latest`, `alpine:latest`, `redis:7-alpine`) sehingga lisensi dan versi dapat berubah tanpa disadari. |

### 2.3 Pengendalian teknis yang mendukung

- Penyedia font Nuxt Fonts dimatikan untuk semua penyedia (`frontend/nuxt.config.ts`, blok `fonts.providers`), tetapi font Space Grotesk tetap dimuat dari Google Fonts melalui `<link>` (lisensi font SIL OFL 1.1, memperbolehkan penggunaan komersial).
- Pemindaian berkas unggahan memakai ClamAV sebagai kontainer terpisah (`backend/clamav/clamd.conf`, image dipin dengan digest). Pemisahan proses ini relevan untuk analisis copyleft (komunikasi via jaringan TCP, bukan penautan pustaka).
- Image ClamAV dipin dengan digest sha256; ini praktik yang baik untuk reprodusibilitas inventory dan dapat ditiru untuk image lain.
- Tidak ada pengendalian lisensi otomatis: `.github/workflows/` hanya memuat `deploy.yml`, `deploy-dev.yml`, dan `route-coverage.yml` (lihat A.5.8); tidak ada `license-checker`, `go-licenses`, `pip-licenses`, Dependabot, atau SBOM.

### 2.4 Kepemilikan kode dan kontributor

- `package.json` bertanda `"private": true` (mencegah publikasi npm; visibilitas repositori GitHub **tidak dapat diverifikasi dari repo**). **Tidak ada berkas `LICENSE`, `NOTICE`, atau `THIRD_PARTY_LICENSES`** di akar repositori maupun di `frontend/`.
- Kontributor di riwayat git: Dimas Saifullah Ramadhan, RauMomo, mikhail-karim, Sumanadi Rahmanachika, muhammadrizqiko, dan `google-labs-jules[bot]`. Status hubungan kerja (karyawan/kontraktor) dan klausul pengalihan IP **tidak dapat diverifikasi dari repo**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kebijakan IP dan penggunaan perangkat lunak (disetujui) | **Belum tersedia** | Draf ketentuan diusulkan pada Lampiran A. |
| License register | **Belum tersedia** | Lampiran B adalah draf register awal; belum diverifikasi alat, belum ada pemilik/tanggal persetujuan per komponen. |
| Software inventory / SBOM | Sebagian | Manifes ada; tidak ada SBOM terhasilkan (CycloneDX/SPDX). Inventory image container tersebar di beberapa berkas compose. |
| Bukti pembelian / langganan lisensi komersial | **Belum tersedia** | Tidak ada faktur, kontrak, atau ToS tersimpan di repo. Layanan berbayar yang terlihat dari kode: ImageKit, OpenAI API, SMTP/email, domain/DNS/VPS (lihat Lampiran B.6). |
| Hasil audit lisensi berkala | **Belum tersedia** | Belum pernah dilakukan; prosedur diusulkan pada Lampiran C. |
| Perjanjian kerja/kontraktor dengan klausul IP | **Belum dapat diverifikasi** | Berada di luar repo (HR/Legal). |
| Perjanjian klien (kepemilikan data tenant, hak atas keluaran) | **Belum dapat diverifikasi** | Berada di luar repo; periksa kontrak langganan dan kesepakatan pemrosesan data. |
| Bukti izin penggunaan aset UI/konten (gambar, ikon bendera, templat dokumen) | **Belum tersedia** | Asal `bg-login.jpg`, `english.svg`, `indonesia.svg`, `avatar-placeholder.png` tidak terdokumentasi. |
| Persetujuan sebelum menambah dependensi baru | **Belum tersedia** | Tidak ada gerbang di PR/CI. |

---

## 4. Catatan / Gap Terverifikasi

Temuan berisiko lisensi (diurutkan menurut keparahan):

1. **Paket GPL-2.0 pada Data Hub (risiko copyleft tinggi bila didistribusikan).** `data-server/caatt-engine/requirements.txt` memuat `fuzzywuzzy==0.18.0` dan `python-Levenshtein==0.25.1`. Kedua paket dikenal berlisensi GPL-2.0 (**perlu verifikasi** dengan `pip-licenses`). Selama CAATT engine hanya dijalankan di server kami dan hasilnya diakses melalui API/Jupyter, kewajiban GPL umumnya tidak terpicu; jika image atau kode CAATT diserahkan ke klien (paket Data Hub on-premises), kewajiban pengungkapan kode sumber seluruh karya turunan dapat terpicu. Alternatif permisif: `rapidfuzz` (MIT) untuk keduanya. **Butuh keputusan.**
2. **Redis, TimescaleDB, RedisInsight: lisensi source-available (bukan OSI open source).** (a) `redis:7-alpine` (tag mengambang) di `backend/docker-compose*.yml`: Redis versi hingga 7.2.x berlisensi BSD-3, sedangkan 7.4 ke atas berlisensi RSALv2/SSPLv1 (Redis 8 menambah opsi AGPLv3); tag mengambang berarti lisensi bergantung pada versi yang ditarik saat deploy. Penggunaan internal sebagai cache umumnya diperbolehkan, tetapi menawarkan Redis sebagai layanan kelola ke pihak lain tidak. (b) `timescale/timescaledb:2.17.2-pg16` di `data-server/docker-compose.yml`: inti Apache-2.0, tetapi fitur kompresi/continuous aggregate berlisensi Timescale License (TSL, source-available); penggunaan internal diperbolehkan, tetapi tidak boleh ditawarkan sebagai DBaaS. Data Hub memakai tabel kolom terkompresi (diagram `data_server_medallion_flow_v2.png`), jadi kemungkinan memakai fitur TSL (**perlu verifikasi**). (c) `redis/redisinsight:latest` (compose dev/internal): berlisensi RedisInsight (SSPL); hanya untuk lingkungan non-produksi, pastikan tidak dikirim ke klien. **Butuh keputusan** apakah pin ke versi BSD (Redis 7.2.x) atau Valkey (BSD-3) atau menerima syarat saat ini dengan catatan.
3. **Grafana, Loki, Promtail: AGPL-3.0 (perlu verifikasi versi).** `monitoring/docker-compose.monitoring.yml` memakai `grafana/grafana:latest`, `grafana/loki:3.7.8`, `grafana/promtail:3.6.8`. Penggunaan tidak dimodifikasi untuk monitoring internal tidak mewajibkan pengungkapan kode; memodifikasi dan menyediakannya ke pengguna jaringan (misalnya klien) akan mewajibkan. `prom/prometheus:latest` dan `prom/node-exporter:latest` berlisensi Apache-2.0. Tag `latest` mengambang. **Kebijakan perlu menyatakan: tidak dimodifikasi dan tidak diekspos ke klien.**
4. **ClamAV: GPL-2.0.** Dijalankan sebagai kontainer terpisah dan diakses via TCP (`clamd.conf`); berstatus *mere aggregation*, risiko rendah selama tidak ditautkan ke kode kami. Tambahan: basis data tanda tangan ClamAV (freshclam) memiliki syarat tersendiri; **perlu verifikasi**.
5. **Pustaka LGPL di Python.** `psycopg2-binary` dan `psycopg[binary]` (LGPL-3.0 dengan pengecualian OpenSSL) dipakai sebagai pustaka dinamis; risiko rendah bila tidak dimodifikasi. **Perlu verifikasi.**
6. **Integrasi OpenAI tanpa kontrak/penggunaan tercatat.** `frontend/server/api/openai.ts` memanggil `client.completions.create` dengan model `gpt-3.5-turbo-instruct` memakai `OPENAI_API_KEY`; modul `nuxt-chatgpt` terdaftar di `frontend/nuxt.config.ts`. Pencarian di `frontend/pages`, `components`, `composables`, dan `stores` tidak menemukan pemakai endpoint ini, sehingga kemungkinan kode tidak terpakai (**perlu konfirmasi**). Jika aktif, data klien dapat terkirim ke pihak ketiga (isu A.5.20 dan AIP-001) dan syarat layanan OpenAI berlaku. Model `gpt-3.5-turbo-instruct` juga berstatus lawas. **Butuh keputusan**: hapus atau dokumentasikan.
7. **Model AI pihak ketiga tanpa catatan lisensi.** `backend/python-ai/train.py` memuat `indobenchmark/indobert-base-p1` dari Hugging Face dan `ai_model_training/save_all_models.py` menyebut bundel IndoBERT/TF-IDF. Lisensi bobot model dan dataset pelatihnya **belum diverifikasi** (kartu model Hugging Face harus diperiksa). Materi pemasaran pada diagram eksekutif menyebut IndoBERT, sehingga klaim kepada klien harus didukung izin pakai komersial. Dataset pelatihan di `ai_model_training/anomaly_prediction/` dibangkitkan oleh skrip `generate_*_dataset.py` (sintetis, **perlu konfirmasi** bahwa tidak ada data klien nyata).
8. **Dokumen klien/pihak ketiga tersimpan di repositori.** `backend/LHA_020_LHA_01_KS_IAD_2023.docx` dan `backend/Test_LHA_Fixed.docx` ter-commit dan tampak berasal dari dokumen audit nyata (nama berkas memuat nomor LHA dan tahun). Jika berasal dari klien, ini adalah isu kepemilikan konten/kerahasiaan (A.5.12, A.5.34) dan hak untuk menggunakan sebagai templat. **Perlu konfirmasi asal dan izin; pertimbangkan penghapusan dari riwayat.**
9. **Aset UI tanpa catatan sumber.** `frontend/assets/images/audits/bg-login.jpg` (juga di `frontend/public/assets/images/audits/`), `frontend/assets/images/languages/english.svg`, `indonesia.svg`, `frontend/assets/images/utils/avatar-placeholder.png`. Ikon `auditsphere-icon.png` dan `favicon.png` diasumsikan milik sendiri (**perlu konfirmasi**). Bendera nasional umumnya domain publik/aman, tetapi foto latar belakang memerlukan bukti lisensi (stok berlisensi atau buatan sendiri).
10. **Sumber ikon dan font dimuat dari CDN pihak ketiga tanpa pin.** `frontend/nuxt.config.ts` memuat `https://unpkg.com/@phosphor-icons/web` (tanpa versi dan tanpa SRI) dan Google Fonts. Lisensi Phosphor (MIT) dan Space Grotesk (SIL OFL 1.1) memperbolehkan, namun hasilnya perlu dicatat di register; risiko rantai pasok adalah isu terpisah (A.8.x). Ikon `i-lucide-*` (964 pemakaian, ISC) dan `i-heroicons-*` (517 pemakaian, MIT) dimuat melalui Nuxt Icon/Iconify; paket `@iconify-json/*` tidak terpasang secara lokal sehingga mode muat (CDN API atau bundel) **perlu verifikasi**.
11. **Konten standar profesi.** Modul Quality Assurance dan Piagam Audit memakai istilah dan kerangka IPPF/GIAS (The IIA). Ringkasan atau rujukan umumnya aman; penyalinan teks standar secara verbatim ke templat/seeder memerlukan izin The IIA. **Perlu pemeriksaan isi** seeder/lokalisasi (`frontend/locales/*/common.json`, `backend/*/pkg/database/seeders`).
12. **Versi dan sumber dependensi tidak seragam.** Dua berkas kunci frontend (`package-lock.json` dan `bun.lock`), versi `gin`/`sarama`/`go-redis` berbeda antar service Go, `data-hub-api` memakai `>=`, `kong-gateway` memakai Go 1.21 dan image `golang:1.21-alpine` (sudah tidak didukung), dan image `alpine:latest`, `kong:3.4`/`3.8`, `confluentinc/cp-kafka` 7.5.0/7.6.1 ada bersamaan. Inventory tidak dapat deterministik; selain itu `confluentinc/cp-kafka` memuat Apache Kafka (Apache-2.0) tetapi paket Confluent tambahan dapat berlisensi Confluent Community License (**perlu verifikasi**).
13. **Tidak ada SBOM, `LICENSE`, `NOTICE`, atau pemindaian lisensi di CI.** Tidak ada Dependabot/Renovate; lisensi dependensi baru tidak diperiksa saat PR (tambahan terhadap item A12 di `backend-enhancement-plan.md`).
14. **Pengalihan IP dari kontributor belum terbukti.** Termasuk pekerjaan oleh agen AI (komit `google-labs-jules[bot]`, Claude Code); status hak cipta atas keluaran AI bervariasi menurut yurisdiksi dan syarat layanan alat. Belum ada aturan yang melarang memasukkan kode atau data klien ke alat AI publik selain AIP-001 (isi **perlu diperiksa**).
15. **Kepemilikan data/konten tenant belum tertulis dalam dokumen internal.** Klausul kepemilikan data klien, lisensi terbatas untuk memproses, hak ekspor, dan penghapusan saat offboarding tenant (`scripts/onboard-tenant.sh`) belum terdokumentasi di repo; kontrak langganan perlu diperiksa.
16. **Layanan pihak ketiga tanpa bukti kontrak.** ImageKit (`backend/shared/media/imagekit.go`, variabel `IMAGEKIT_*` di `docker-compose.prod.yml`), pengiriman email (`backend/shared/email/smtp.go`, penyedia Resend/Titan per catatan tim; **perlu verifikasi**), Cloudflare, Rumahweb, penyedia VPS (petunjuk: `hostinger-api-mcp` di `package.json` akar, **perlu verifikasi**), dan Backblaze B2 (`docs/backup-b2-setup.md`, status DRAFT). Tidak ada faktur, rencana langganan, atau ToS yang tersimpan.
17. **Paket pengembangan tidak terkait produk.** `hostinger-api-mcp` di `package.json` akar (alat bantu pengembang; lisensi **perlu verifikasi**), `frontend/frontend-prod.tar.gz`, `backend/backend-dev.tar.gz` (arsip build di working tree; di-ignore git, tidak ter-commit, tetapi dapat memuat paket pihak ketiga tanpa inventaris).

Catatan penting: sebagian besar dependensi langsung frontend berlisensi permisif (MIT/Apache-2.0) dan **terverifikasi** dari berkas paket (Lampiran B.1). Tidak ditemukan paket komersial berbayar di frontend; `@nuxt/ui` v4.4.0 berlisensi MIT (terverifikasi), sehingga tidak ada kewajiban lisensi Nuxt UI Pro.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Susun, setujui, dan tanda tangani **Kebijakan Hak Kekayaan Intelektual dan Penggunaan Perangkat Lunak** berdasarkan Lampiran A; rujuk ke A.5.1, A.5.12, A.5.20, AIP-001. | Lead Information Security dengan Legal & Compliance | Tinggi | [isi] | Dokumen kebijakan bertanda tangan |
| 2 | Putuskan nasib `fuzzywuzzy` dan `python-Levenshtein` (GPL): ganti dengan `rapidfuzz`, atau dokumentasikan bahwa CAATT engine tidak pernah didistribusikan ke klien. Verifikasi lisensi dengan `pip-licenses`. | Legal & Compliance bersama Tim Data Hub | Tinggi | [isi] | Keputusan tertulis; perubahan `requirements.txt` bila diganti |
| 3 | Putuskan strategi Redis (pin ke 7.2.x BSD, pindah ke Valkey, atau terima RSALv2/SSPL dengan catatan), TimescaleDB (verifikasi fitur TSL yang dipakai), RedisInsight (batasi non-produksi). Pin tag image, hindari `latest`. | DevOps Lead dengan Legal & Compliance | Tinggi | [isi] | Keputusan tertulis; tag image dipin |
| 4 | Konfirmasi dan putuskan `frontend/server/api/openai.ts` dan `nuxt-chatgpt`: hapus bila tidak dipakai, atau catat kontrak/ToS OpenAI, aturan data, dan penilaian pemasok. | Lead Frontend dan Security Officer | Tinggi | [isi] | PR penghapusan atau entri pemasok di register |
| 5 | Periksa asal `backend/LHA_020_LHA_01_KS_IAD_2023.docx` dan `backend/Test_LHA_Fixed.docx`; jika milik klien, hapus dari repo (dan riwayat) dan gunakan data sintetis. | Project Lead dengan Legal & Compliance | Tinggi | [isi] | Catatan asal/izin; commit penghapusan |
| 6 | Verifikasi lisensi bobot model `indobenchmark/indobert-base-p1`, model lain di `backend/python-ai/models` dan `ai_model_training`, serta dataset pelatihan; catat di register. Pastikan klaim pemasaran sesuai. | Tim AI/ML dengan Legal & Compliance | Sedang | [isi] | Entri register model dan kartu model tersimpan |
| 7 | Bangun **license register** resmi dari Lampiran B: pemilik komponen, versi, SPDX, jenis, tanggal verifikasi, keputusan, dan Evidence ID. Simpan di Master Audit Evidence Index. | Lead Information Security | Tinggi | [isi] | License register terversi |
| 8 | Hasilkan **SBOM** (CycloneDX) per komponen: `cyclonedx-npm`/`syft` untuk frontend dan image, `cyclonedx-gomod` untuk Go, `cyclonedx-py` untuk Python; pilih satu berkas kunci frontend (npm atau bun). | DevOps Lead | Tinggi | [isi] | Berkas SBOM terlampir pada setiap rilis |
| 9 | Tambahkan **gerbang lisensi di CI** (`license-checker`/`license-finder`, `go-licenses`, `pip-licenses`, `trivy`/`grype`) dengan daftar lisensi yang diizinkan, ditolak (AGPL, GPL, SSPL, RSAL, BUSL tanpa persetujuan), dan perlu tinjauan; satukan dengan item A12 dan SCG-001. | DevOps Lead | Sedang | [isi] | Workflow CI dan log run |
| 10 | Kumpulkan **bukti pembelian/langganan** dan ToS untuk ImageKit, penyedia email, Cloudflare, Rumahweb, VPS, Backblaze B2, GitHub, dan layanan AI; simpan di penyimpanan Legal (akses terbatas). | Legal & Compliance bersama Finance | Tinggi | [isi] | Folder faktur/kontrak dan indeks |
| 11 | Dokumentasikan sumber dan lisensi aset UI (`bg-login.jpg`, bendera, avatar) dan ikon/font CDN; ganti aset tak berlisensi. Pin versi Phosphor dan pertimbangkan self-hosting. | Lead Frontend | Sedang | [isi] | Berkas `THIRD_PARTY_LICENSES` dan catatan aset |
| 12 | Tambahkan `LICENSE` (proprietary) dan `THIRD_PARTY_NOTICES` pada repo dan paket yang diserahkan; siapkan halaman "Open-source notices" di aplikasi bila diwajibkan lisensi (atribusi). | Legal & Compliance dan Lead Frontend | Sedang | [isi] | Berkas notice dan halaman notice |
| 13 | Tegaskan di kontrak/SLA klien: kepemilikan data tenant tetap milik klien, lisensi pemrosesan terbatas, hak ekspor dan penghapusan, serta kepemilikan keluaran (KKA/LHA) dan templat bawaan. | Legal & Compliance | Tinggi | [isi] | Klausul standar kontrak |
| 14 | Pastikan semua kontributor (karyawan dan kontraktor) menandatangani klausul **pengalihan IP dan kerahasiaan**; catat hasilnya. Tambahkan aturan penggunaan alat AI untuk pengembangan (jenis data yang dilarang, penandaan, peninjauan manusia). | Legal & Compliance dan HR | Tinggi | [isi] | Daftar penandatangan dan kebijakan alat AI |
| 15 | Jadwalkan **audit lisensi berkala** (Lampiran C) dan buat pengingat; hasil pertama menjadi baseline. | Lead Information Security | Sedang | [isi] | Laporan audit lisensi pertama |
| 16 | Periksa seeder/lokalisasi dan templat dokumen untuk teks verbatim IPPF/GIAS atau materi berhak cipta lain; ganti dengan parafrasa atau dapatkan izin. | Project Lead dengan Legal & Compliance | Sedang | [isi] | Catatan hasil tinjauan konten |
| 17 | Rapikan versi dan image: satu versi major per dependensi lintas service Go bila memungkinkan, ganti `golang:1.21-alpine`, hindari `latest`, dan jangan menyimpan arsip `.tar.gz` build di working tree yang dibagikan. | DevOps Lead dan Tim Backend | Rendah | [isi] | PR rapi-rapi dan daftar tag image |
| 18 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

Catatan pelaksanaan: tindakan yang menyentuh produksi (pin image, perubahan Redis) dilakukan melalui `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh` dengan `--ref`, tanpa seeder), bukan redeploy sembarang.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.32-01 | Manifes dependensi (frontend, Go, Python AI, Python Data Hub) | `frontend/package.json`, `frontend/package-lock.json`, `backend/*/go.mod`, `backend/python-ai/requirements.txt`, `data-server/*/requirements.txt` | Internal |
| EV-A5.32-02 | Daftar image container dan versi (compose dan Dockerfile) | `backend/docker-compose*.yml`, `data-server/docker-compose*.yml`, `monitoring/docker-compose.monitoring.yml`, `*/Dockerfile` | Internal |
| EV-A5.32-03 | Software inventory dan license register (draf awal Lampiran B; versi resmi menyusul) | `docs/ISO_27001/A.5.32_Hak_Kekayaan_Intelektual.md` (Lampiran B) | Internal |
| EV-A5.32-04 | Kebijakan IP dan penggunaan perangkat lunak (setelah tindakan 1) | [belum ada] | Internal |
| EV-A5.32-05 | Bukti pembelian/langganan dan ToS pemasok (setelah tindakan 10) | [penyimpanan Legal, belum ada] | Restricted |
| EV-A5.32-06 | SBOM dan log pemindaian lisensi CI (setelah tindakan 8 dan 9) | [belum ada] | Internal |
| EV-A5.32-07 | Laporan audit lisensi berkala (setelah tindakan 15) | [belum ada] | Internal |

Setelah tindakan 1, 7, 8, 10, dan 15 selesai, tambahkan ID untuk salinan bertanda tangan, SBOM per rilis, dan laporan audit lisensi.

---

## Lampiran A — Draf Ketentuan Kebijakan IP dan Penggunaan Perangkat Lunak (usulan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Ketentuan hukum final ditetapkan Legal & Compliance.

**A.1 Kepemilikan kode dan konten**

1. Kode sumber, desain, basis data skema, model AI hasil pelatihan sendiri, dokumen, dan merek AuditSphere (nama, logo `auditsphere-icon.png`) dimiliki perusahaan. Kontributor mengalihkan hak ekonomi melalui perjanjian kerja atau kontrak.
2. **Data dan konten tenant** (data master, rencana audit, KKA, LHA, ATR, berkas unggahan, data yang masuk ke Data Hub) adalah milik klien. Perusahaan hanya memiliki lisensi terbatas untuk memproses data demi layanan; tidak dipakai untuk pelatihan model lintas tenant tanpa persetujuan tertulis. Data tenant dikembalikan/dihapus saat offboarding sesuai kontrak.
3. Templat bawaan (struktur KKA/LHA, formulir bulk import) dimiliki perusahaan; isi yang diisi klien adalah milik klien. Templat yang berasal dari klien tidak digunakan kembali untuk klien lain tanpa izin.

**A.2 Penggunaan open-source**

1. Penambahan dependensi baru memerlukan pemeriksaan lisensi pada PR. Kategori:
   - **Diizinkan:** MIT, BSD-2/3, ISC, Apache-2.0, PostgreSQL, PSF, Zlib, SIL OFL 1.1 (font).
   - **Perlu persetujuan Legal:** LGPL (hanya sebagai pustaka dinamis tak dimodifikasi), MPL-2.0, EPL, CC-BY (aset), lisensi source-available (SSPL, RSALv2, BUSL, TSL, Confluent Community).
   - **Dilarang pada produk yang didistribusikan atau dilayankan ke klien tanpa persetujuan:** GPL, AGPL.
   - **Alat internal tak terdistribusi** (Grafana, Loki, ClamAV): boleh sebagai kontainer terpisah, tidak dimodifikasi, tidak diekspos ke klien; versi dipin.
2. Atribusi dan notice wajib dipenuhi (lihat tindakan 12).
3. Dilarang menyalin kode dari StackOverflow, GitHub, atau sumber lain tanpa memeriksa lisensinya; potongan berlisensi tidak jelas tidak boleh masuk repo.

**A.3 Lisensi komersial dan bukti pembelian**

1. Setiap layanan/lisensi berbayar atau berkunci akun (ImageKit, penyedia email, VPS, domain, penyimpanan cadangan, layanan AI) terdaftar di license register dengan pemilik, nomor akun, paket, tanggal mulai/berakhir, jumlah kursi/penggunaan, dan lokasi faktur.
2. Pengadaan melalui Finance; faktur dan kontrak disimpan di penyimpanan Legal (akses terbatas). Akun layanan atas nama perusahaan, bukan perorangan.
3. Dilarang memakai perangkat lunak bajakan atau lisensi pribadi untuk kerja perusahaan. Lisensi perangkat pengembang (IDE, alat AI, Office) dicatat juga.

**A.4 Konten pihak ketiga**

1. Gambar, ikon, font, dan teks hanya dipakai bila lisensinya tercatat di `THIRD_PARTY_LICENSES`. Larangan mengambil gambar dari hasil pencarian web.
2. Teks standar (IPPF/GIAS, ISO/IEC 27001, peraturan OJK) tidak disalin utuh; gunakan rujukan atau parafrasa kecuali memiliki izin.
3. Dokumen klien tidak disimpan di repo (A.5.12) dan tidak digunakan sebagai contoh/templat tanpa persetujuan tertulis.

**A.5 Konten buatan AI**

1. Penggunaan alat AI untuk kode/dokumen mengikuti AIP-001; dilarang memasukkan data klien, kredensial, atau kode rahasia ke layanan AI publik.
2. Keluaran AI ditinjau manusia sebelum masuk repo; reviewer memeriksa tidak adanya potongan yang tampak disalin dari sumber berlisensi (kode, teks).
3. Perusahaan mencatat alat AI yang dipakai (nama, penyedia, rencana, syarat kepemilikan keluaran) di license register; komit yang dihasilkan bot ditandai.
4. Model pra-latih dan dataset pihak ketiga hanya dipakai bila lisensinya mengizinkan penggunaan komersial dan dicatat (nama, versi, sumber, lisensi).

**A.6 Offboarding dan kontraktor**

1. Karyawan/kontraktor menandatangani klausul pengalihan IP dan kerahasiaan sebelum mendapat akses repo.
2. Saat selesai: akses repo, registri, akun layanan, dan perangkat dicabut; kode/dokumen dikembalikan; akun layanan perorangan dialihkan ke akun perusahaan.
3. Kontraktor membuktikan komponen pihak ketiga yang dibawanya (daftar dependensi dan lisensi) pada serah terima.

**A.7 Pelanggaran dan pengecualian**

Dugaan pelanggaran lisensi dilaporkan ke Lead Information Security dan Legal & Compliance, dicatat sebagai insiden (A.5.24), dan dimitigasi (ganti komponen, beli lisensi, atau hentikan penggunaan). Pengecualian terhadap kebijakan memerlukan persetujuan tertulis Legal & Compliance dan tanggal tinjau ulang.

---

## Lampiran B — Software Inventory dan License Register (draf awal, 30 Oktober 2026)

Status verifikasi: **T** = Terverifikasi dari berkas paket terpasang; **PV** = Perlu verifikasi (diisi dari pengetahuan umum, belum dipindai alat); **TV** = Tidak dapat diverifikasi dari repo. Jenis: OSS = open source, SA = source-available, KOM = komersial/layanan berbayar. Risiko: R = Rendah, S = Sedang, T = Tinggi.

### B.1 Frontend (Nuxt 4) — `frontend/package.json`

| Komponen | Versi terpasang (rentang) | Lisensi (SPDX) | Jenis | Risiko | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| nuxt | 4.3.0 (^4.1.2) | MIT | OSS | R | T |
| @nuxt/ui | 4.4.0 (^4.4.0) | MIT | OSS | R | T |
| @nuxt/image | 2.0.0 | MIT | OSS | R | T |
| @nuxt/fonts (transitif) | 0.12.1 | MIT | OSS | R | T |
| @nuxt/icon (transitif) | 2.2.1 | MIT | OSS | R | T |
| reka-ui (transitif) | 2.7.0 | MIT | OSS | R | T |
| @pinia/nuxt / pinia | 0.11.3 / 3.0.4 | MIT | OSS | R | T |
| @vueuse/core, @vueuse/nuxt | 14.1.0 | MIT | OSS | R | T |
| vue / vue-router / vue-i18n | 3.5.27 / 4.6.4 / 11.4.8 | MIT | OSS | R | T |
| chart.js / vue-chartjs / nuxt-charts | 4.5.1 / 5.3.3 / 2.1.2 | MIT | OSS | R | T |
| nuxt-auth-utils | 0.5.27 | MIT | OSS | R | T |
| nuxt-chatgpt | 0.4.0 | MIT | OSS | R (kode); lihat temuan 6 untuk layanan | T |
| openai (SDK) | 6.19.0 | Apache-2.0 | OSS | R (kode); layanan OpenAI = KOM | T |
| qrcode | 1.5.4 | MIT | OSS | R | T |
| zod | 3.25.76 | MIT | OSS | R | T |
| date-fns | 4.1.0 | MIT | OSS | R | T |
| @rollup/plugin-yaml | 4.1.2 | MIT | OSS | R | T |
| @nuxtjs/tailwindcss / tailwindcss | 6.14.0 / 4.1.18 | MIT | OSS | R | T |
| Dev: typescript | 5.9.3 | Apache-2.0 | OSS | R | T |
| Dev: vitest, @nuxt/eslint | 4.0.18 / 1.13.0 | MIT | OSS | R | T |
| Dev: @playwright/test | 1.58.0 | Apache-2.0 | OSS | R | T |
| Dev: happy-dom, vue-tsc, @types/*, @nuxt/devtools | (lihat package.json) | MIT (umum) | OSS | R | PV |
| Paket transitif lainnya | 874 paket di `node_modules` | beragam | OSS | S | PV (jalankan `license-checker --summary`) |

### B.2 Aset UI, font, ikon

| Aset | Sumber | Lisensi | Jenis | Risiko | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Font Space Grotesk | Google Fonts via `<link>` (`nuxt.config.ts`) | SIL OFL 1.1 | OSS (font) | R | PV |
| Phosphor Icons Web | `https://unpkg.com/@phosphor-icons/web` (tanpa pin) | MIT | OSS | R | PV |
| Ikon Lucide (`i-lucide-*`) | Nuxt Icon / Iconify | ISC | OSS | R | PV (mode muat belum dicek) |
| Ikon Heroicons (`i-heroicons-*`) | Nuxt Icon / Iconify | MIT | OSS | R | PV |
| `bg-login.jpg` | `frontend/assets/images/audits/`, `frontend/public/assets/images/audits/` | Tidak diketahui | Tidak diketahui | S | TV |
| `english.svg`, `indonesia.svg` | `frontend/assets/images/languages/` | Tidak diketahui (bendera) | Tidak diketahui | R | TV |
| `avatar-placeholder.png`, `qr-code.png` | `frontend/assets/images/` | Tidak diketahui | Tidak diketahui | R | TV |
| `auditsphere-icon.png`, `favicon.png` | `frontend/assets/images/utils/`, `frontend/public/` | Milik sendiri (asumsi) | Milik sendiri | R | TV |

### B.3 Backend Go — `backend/*/go.mod`

Versi dibaca dari `go.mod`; lisensi **PV** untuk semua baris (sumber: pengetahuan umum; jalankan `go-licenses report ./...`).

| Pustaka | Versi (service) | Lisensi (SPDX) | Jenis | Risiko |
| :--- | :--- | :--- | :--- | :--- |
| gin-gonic/gin | 1.9.1 (auth, audit, kong-gateway); 1.12.0 (master, risk, analytics) | MIT | OSS | R |
| gin-contrib/cors | 1.7.7 (analytics) | MIT | OSS | R |
| gorm.io/gorm, gorm.io/driver/postgres | 1.31.1 / 1.6.0 | MIT | OSS | R |
| jackc/pgx/v5 | 5.8.0 (master) | MIT | OSS | R |
| glebarez/sqlite (membawa modernc.org/sqlite) | 1.11.0 (audit, master, risk) | MIT (modernc: BSD-3-Clause) | OSS | R |
| golang-jwt/jwt/v5 | 5.3.1 (5.2.0 di kong-gateway) | MIT | OSS | R |
| google/uuid | 1.6.0 | BSD-3-Clause | OSS | R |
| IBM/sarama | 1.47.0; 1.50.2 (master) | MIT | OSS | R |
| redis/go-redis/v9 | 9.4.0; 9.20.0 | BSD-2-Clause | OSS | R |
| casbin/casbin/v2, casbin/gorm-adapter/v3 | 2.135.0 / 3.41.0 | Apache-2.0 | OSS | R |
| didip/tollbooth | 4.0.2 | MIT | OSS | R |
| go-playground/validator/v10 | 10.17.0 s/d 10.30.3 | MIT | OSS | R |
| spf13/cobra, spf13/viper (+remote) | 1.8.0 s/d 1.10.2 / 1.18.2 s/d 1.21.0 | Apache-2.0 / MIT | OSS | R |
| sirupsen/logrus | 1.4.2 s/d 1.9.4 | MIT | OSS | R |
| uber-go/zap | 1.26.0 s/d 1.28.0 | MIT | OSS | R |
| swaggo/swag, files, gin-swagger | 1.16.6 / 1.0.1 / 1.6.x | MIT | OSS | R |
| stretchr/testify | 1.11.1 | MIT | OSS | R |
| golang.org/x/crypto, golang.org/x/time | 0.48.0 s/d 0.53.0; 0.5.0 | BSD-3-Clause | OSS | R |
| imagekit-developer/imagekit-go/v2 | 2.8.0 (audit) | Apache-2.0 (PV) | OSS (SDK); layanan ImageKit = KOM | S |
| Pustaka standar `net/smtp` | Go toolchain 1.25.x | BSD-3-Clause | OSS | R |

### B.4 Python — AI dan Data Hub

| Komponen | Versi | Lisensi (SPDX) | Jenis | Risiko | Lokasi | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| fastapi / uvicorn / pydantic | 0.115.6 / 0.34.0 / 2.10.4 (python-ai); 0.111.0 / 0.30.1 (data-server) | MIT / BSD-3-Clause / MIT | OSS | R | python-ai, data-server | PV |
| tensorflow-cpu | 2.18.0 | Apache-2.0 | OSS | R | python-ai | PV |
| transformers | 4.47.1 | Apache-2.0 | OSS | R | python-ai | PV |
| xgboost | 2.1.3; 2.0.3 | Apache-2.0 | OSS | R | python-ai, ai-engine | PV |
| scikit-learn / pandas / numpy | 1.6.0, 1.5.1 / 2.2.3, 2.2.2 / 1.26.4 | BSD-3-Clause | OSS | R | python-ai, data-server | PV |
| torch (CPU) | 2.3.1 | BSD-3-Clause | OSS | R | ai-engine | PV |
| sqlalchemy / apscheduler / python-dotenv / pytest | 2.0.31 / 3.10.4 / 1.0.1 / >=8.0.0 | MIT / MIT / BSD-3-Clause / MIT | OSS | R | data-server | PV |
| psycopg2-binary / psycopg[binary] / psycopg-pool | 2.9.9 / >=3.1.18 / >=3.2.1 | LGPL-3.0-or-later dengan pengecualian | OSS (copyleft lemah) | S | data-server | PV |
| orjson / pymysql / cryptography / requests | >=3.10.0 / >=1.1.0 / >=42.0.0 / 2.32.3 | Apache-2.0 atau MIT / MIT / Apache-2.0 atau BSD-3-Clause / Apache-2.0 | OSS | R | data-server | PV |
| dlt[postgres] | 0.5.4 | Apache-2.0 | OSS | R | data-hub-api, caatt-engine | PV |
| matplotlib / seaborn / scipy / openpyxl | 3.9.1 / 0.13.2 / 1.14.0 / 3.1.5 | PSF-style (Matplotlib License) / BSD-3-Clause / BSD-3-Clause / MIT | OSS | R | caatt-engine | PV |
| **fuzzywuzzy** | 0.18.0 | **GPL-2.0** | OSS (copyleft kuat) | **T** | caatt-engine | PV |
| **python-Levenshtein** | 0.25.1 | **GPL-2.0-or-later** | OSS (copyleft kuat) | **T** | caatt-engine | PV |
| IndoBERT `indobenchmark/indobert-base-p1` | (tanpa pin revisi) | Perlu diperiksa pada kartu model Hugging Face | Model pihak ketiga | S | `backend/python-ai/train.py` | TV |
| Model/dataset hasil latih sendiri | `backend/python-ai/models`, `ai_model_training/` | Milik sendiri; dataset sintetis dari skrip `generate_*_dataset.py` (konfirmasi) | Milik sendiri | R | | TV |

### B.5 Image container dan infrastruktur

| Komponen | Image/versi (sumber) | Lisensi (SPDX) | Jenis | Risiko | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| PostgreSQL | `postgres:16`, `postgres:16-alpine` | PostgreSQL | OSS | R | PV |
| Redis | `redis:7-alpine` (tag mengambang) | BSD-3-Clause (<=7.2.x) / RSALv2 atau SSPL-1.0 (>=7.4) | OSS atau SA, bergantung versi | **S/T** | PV |
| Apache Kafka & ZooKeeper | `confluentinc/cp-kafka:7.5.0/7.6.1`, `cp-zookeeper:7.5.0/7.6.1` | Apache-2.0 (Kafka/ZooKeeper); komponen Confluent: Confluent Community License (PV) | OSS/SA | S | PV |
| Kong Gateway | `kong:3.4`, `kong:3.8` (prod: 3.8) | Apache-2.0 (edisi OSS) | OSS | R | PV |
| Kong Admin/Go gateway lama | `backend/kong-gateway` (kode sendiri, `golang:1.21-alpine`) | Milik sendiri | Milik sendiri | R | T |
| ClamAV | `clamav/clamav@sha256:982e28f9...` | GPL-2.0 (+ basis data tanda tangan) | OSS (copyleft) | S (terpisah via TCP) | PV |
| TimescaleDB | `timescale/timescaledb:2.17.2-pg16` | Apache-2.0 + Timescale License (TSL) | OSS + SA | S | PV |
| nginx | `nginx:1.27-alpine` | BSD-2-Clause | OSS | R | PV |
| JupyterLab/scipy-notebook | `jupyter/scipy-notebook:python-3.10` | BSD-3-Clause (+ paket campuran) | OSS | S | PV |
| Python, Node, Alpine | `python:3.10-slim`, `node:20-alpine`, `alpine:latest` | PSF / MIT / beragam (GPL userland tak dimodifikasi) | OSS | R | PV |
| Kafka UI | `provectuslabs/kafka-ui:latest` (dev/internal) | Apache-2.0 | OSS | R | PV |
| RedisInsight | `redis/redisinsight:latest` (dev/internal) | RedisInsight License (SSPL) | SA | S | PV |
| Grafana, Loki, Promtail | `grafana/grafana:latest`, `loki:3.7.8`, `promtail:3.6.8` | AGPL-3.0 | OSS (copyleft kuat, jaringan) | S | PV |
| Prometheus, node-exporter | `prom/prometheus:latest`, `prom/node-exporter:latest` | Apache-2.0 | OSS | R | PV |
| WireGuard | `data-server/wireguard` | GPL-2.0 (modul kernel Linux) | OSS | R (tidak didistribusikan) | PV |
| rclone | dipasang di VPS (`docs/backup-b2-setup.md`, DRAFT) | MIT | OSS | R | PV |
| OS server | AlmaLinux 9 (VPS) | GPL/beragam | OSS | R | TV |
| Docker Engine/Compose | Host VPS | Apache-2.0 | OSS | R | TV |

### B.6 Layanan pihak ketiga / berbayar (bukti pembelian belum ada)

| Layanan | Dipakai untuk | Sumber petunjuk | Jenis | Bukti yang dibutuhkan | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| ImageKit | Penyimpanan/CDN media | `backend/shared/media/imagekit.go`, `IMAGEKIT_*` di `docker-compose.prod.yml` | KOM (SaaS) | Faktur/paket, ToS, DPA | TV |
| OpenAI API | Endpoint `/api/openai` (kemungkinan tak terpakai) | `frontend/server/api/openai.ts` | KOM | Kontrak/ToS jika aktif | TV |
| Penyedia email (Resend, Titan; SMTP) | Pengiriman email sistem | `backend/shared/email/smtp.go`; catatan tim | KOM | Faktur, ToS | TV |
| Cloudflare | DNS/CDN | catatan tim | KOM/free | Akun dan paket | TV |
| Rumahweb | Domain/hosting | catatan tim | KOM | Faktur domain, masa berlaku | TV |
| Penyedia VPS | Server produksi 202.10.34.166 | `docs/` (VPS), `hostinger-api-mcp` | KOM | Faktur, ToS | TV |
| Backblaze B2 | Cadangan lepas lokasi (DRAFT) | `docs/backup-b2-setup.md` | KOM | Akun, ToS | TV |
| GitHub | Repositori dan Actions | `.github/workflows/` | KOM | Paket organisasi | TV |
| Google Fonts, unpkg | Font dan ikon via CDN | `frontend/nuxt.config.ts` | Gratis (ToS) | Catatan penggunaan | PV |
| Alat AI pengembangan (Claude Code, Jules) | Pengembangan kode | `.claude/`, riwayat git | KOM | Akun, ToS keluaran | TV |
| Alat dokumen (Word/Excel untuk docx/xlsx) | Dokumen ISO dan checklist | `docs/ISO_27001/*.docx`, `*.xlsx` | KOM (lisensi Office) | Bukti lisensi Office | TV |

---

## Lampiran C — Prosedur Audit Lisensi Berkala (usulan)

1. **Frekuensi.** Audit lisensi penuh setiap 6 bulan; pemindaian otomatis pada setiap PR dan rilis; tinjauan ad hoc sebelum mengirim paket ke klien dan setelah menambah komponen besar.
2. **Penanggung jawab.** Pelaksana: DevOps Lead (pemindaian dan SBOM). Peninjau: Lead Information Security. Keputusan hukum: Legal & Compliance.
3. **Langkah:**
   1. Hasilkan inventory: `cyclonedx-npm` atau `license-checker --json` (frontend), `go-licenses report ./...` per service Go, `pip-licenses --format=json` per layanan Python, `syft`/`trivy image` per image produksi.
   2. Rekonsiliasi dengan license register (Lampiran B): tambah/hapus komponen, perbarui versi dan lisensi.
   3. Klasifikasikan menurut daftar izin (A.2.1); setiap komponen "Perlu persetujuan" atau "Dilarang" diberi keputusan tertulis: ganti, beli, pisahkan (kontainer terpisah), atau terima risiko dengan tanggal tinjau.
   4. Bandingkan layanan berbayar dengan faktur/kontrak: masa berlaku, jumlah penggunaan, siapa pemilik akun.
   5. Periksa konten: aset UI di `THIRD_PARTY_LICENSES`, dokumen klien di repo, teks standar.
   6. Periksa kepatuhan kontributor: perjanjian IP bagi kontributor baru, pencabutan akses kontributor keluar.
   7. Catat temuan dan tindakan di laporan audit lisensi; temuan High dibuka sebagai tindakan dengan PIC dan tenggat.
4. **Keluaran.** Laporan audit lisensi (tanggal, cakupan, alat dan versi, temuan, keputusan, penandatangan), SBOM terlampir, register diperbarui, dan Evidence ID baru.
5. **Eskalasi.** Temuan lisensi High yang menyentuh produksi ditangani lewat `auditsphere-devops` dengan jalur deploy aman; insiden kepatuhan mengikuti A.5.24.

---

## Lampiran D — Penanggung Jawab (RACI usulan)

| Aktivitas | Legal & Compliance | Lead Information Security | DevOps Lead | Tim Backend/Frontend/AI | Project Lead | Finance/HR |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Kebijakan IP dan lisensi | A | R | C | C | C | I |
| License register dan SBOM | C | A | R | C | I | I |
| Persetujuan dependensi baru (PR) | C (kategori perlu persetujuan) | C | R (gerbang CI) | R (pengaju) | I | I |
| Bukti pembelian dan kontrak pemasok | A | C | I | I | I | R |
| Audit lisensi berkala | C | A | R | C | I | I |
| Perjanjian IP kontributor/kontraktor | A | I | I | I | C | R |
| Penggunaan alat AI dan keluaran AI | C | A | C | R | C | I |
| Klausul kepemilikan data tenant | A/R | C | I | I | C | I |
| Offboarding (akses dan serah terima) | C | A | R | R | C | R |

R = Responsible, A = Accountable, C = Consulted, I = Informed.

---

Catatan sinkronisasi checklist: baris A.5.32 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan belum diperbarui oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
