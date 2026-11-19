# A.8.8 — Pengelolaan Kerentanan Teknis

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pengelolaan kerentanan teknis |
| **Fokus pemeriksaan** | Kerentanan diidentifikasi, diprioritaskan, dan diperbaiki tepat waktu |
| **Contoh bukti objektif** | Scan report, patch SLA, remediation ticket, exception register |
| **Sistem** | AuditSphere (auditsphere.app): repo `risk-based-audit` (cabang `staging`), lima layanan Go dan modul `kong-gateway`, python-ai, frontend Nuxt, image dan compose (produksi, tenant, training, dev, monitoring), ClamAV, Data Hub, host VPS (OS, Docker, nginx), workflow GitHub Actions; perangkat endpoint personel dirujuk ke A.8.1 |
| **PIC** | Lead Information Security (didukung DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut agar informasi tentang kerentanan teknis pada sistem yang dipakai **diperoleh tepat waktu**, keterpaparan organisasi **dievaluasi**, dan **tindakan yang sesuai diambil** (menambal, memitigasi, atau menerima risiko secara tertulis), dengan peran, tanggung jawab, dan batas waktu yang jelas. Repo hanya membuktikan **manifes dependensi, lockfile, Dockerfile, compose, workflow, skrip, dan dokumen**; repo **tidak** membuktikan paket, versi kernel, atau patch yang terpasang di VPS, versi image yang benar-benar berjalan, status pembaruan signature ClamAV, pengaturan GitHub (Dependabot alerts, secret scanning, branch protection), atau status patch endpoint personel. Semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH, GitHub API, atau konsol yang dipakai. **Tidak ada pemindai yang dijalankan dan tidak ada alat yang dipasang** dalam penyusunan dokumen ini; tidak ada pemeriksaan ke basis data kerentanan eksternal. Karena itu dokumen ini **tidak menyatakan adanya CVE tertentu** pada komponen mana pun dan **tidak menyatakan komponen tertentu bebas kerentanan**. Pernyataan akhir dukungan (*end of support*) untuk versi bahasa, image, atau produk ditandai **Perlu verifikasi** terhadap jadwal resmi pemasok. Butir yang bersumber dari memori kerja internal (catatan 4 sampai 5 November 2026) dicantumkan sebagai **[Catatan]** dan **bukan bukti repo**. Berkas `.env` tidak dibuka. Seluruh angka, ambang, tenggat patch, frekuensi, dan tanggal pada Lampiran A sampai G dan pada kolom target adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan (nilai bawaan literal yang ditemukan di berkas sengaja tidak dikutip).

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.8 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.8** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Pengelolaan kerentanan teknis** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Kerentanan diidentifikasi, diprioritaskan, dan diperbaiki tepat waktu** | Sesuai baris spreadsheet; tiga gate: identifikasi (inventory dan pemindaian, sumber informasi), prioritisasi (klasifikasi tingkat keparahan dan keterpaparan), perbaikan tepat waktu (SLA patch, tiket, pengecualian tertulis). |
| Contoh bukti objektif | **Scan report, patch SLA, remediation ticket, exception register** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | AuditSphere adalah layanan multi-tenant yang memproses data audit klien lembaga jasa keuangan dengan permukaan serangan luas (lima layanan Go, python-ai berbasis ML, frontend Nuxt, Kafka, Postgres, Kong, ClamAV, Data Hub) pada satu VPS. Komponen yang tidak ditambal adalah jalur kompromi langsung. Pengecualian tidak dapat dipertahankan (bagian 1). |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: banyak ekosistem dependensi tanpa pemindaian atau pemicu pembaruan, base image mengambang dan tidak seragam, tidak ada proses patch OS yang terbukti, pola rentan yang pernah ditemukan di repo, dan klien lembaga keuangan akan menanyakan SLA patch dan hasil pemindaian. |
| Status implementasi | **Sebagian (ad hoc)** | Ada penguatan nyata hasil tinjauan satu kali (JWT, unggahan, registrasi, port loopback, firewall host, ClamAV fail-closed), lockfile per modul, dan pin sebagian image. Belum ada proses berulang: tidak ada pemindai, SLA yang disetujui, tiket, register pengecualian, atau pemantauan sumber informasi kerentanan. |
| Ketersediaan bukti | **Sebagian (artefak teknis dan dokumen rencana) / Tidak ada (rekaman proses)** | Manifes, lockfile, Dockerfile, compose, rencana `docs/backend-enhancement-plan.md`, dan dokumen Draft BBP-001/RTP-001/TIP-001 ada. Scan report, tiket remediasi, register pengecualian, hasil pemindaian host, dan bukti patch OS **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.8-01 s/d EV-A8.8-12) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **DevOps Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal bila VUL-001 disetujui dan pemindaian dasar (baseline) pertama selesai | Selanjutnya: tinjauan register dan pengecualian [bulanan], tinjauan kebijakan [tahunan] (Lampiran A, D, F). |
| Desain kontrol | **Belum ada** | Tidak ada kebijakan, kelas keparahan, SLA patch, pemilik per komponen, atau alur triase yang disetujui. Dokumen Draft BBP-001 dan RTP-001 memuat tabel SLA remediasi, tetapi tidak dikaitkan dengan komponen dan angkanya berbeda dari usulan annex lain (bagian 4 butir 16). |
| Efektivitas operasional | **Belum teruji** | Tidak ada scan report, tiket, atau rekaman penutupan. Tidak dapat diklaim bahwa kerentanan pada komponen produksi diketahui, diprioritaskan, atau ditambal dalam batas waktu tertentu. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: tanpa pemindaian dependensi/image/host/rahasia dan tanpa Dependabot yang terbukti; versi Go/Python/Node/Kong tidak seragam dan beberapa mungkin sudah melewati dukungan (Perlu verifikasi); base image `latest` dan kontainer berjalan sebagai root; patch OS/kernel host tidak terbukti; ClamAV tanpa pemantauan kesegaran signature; jalur deploy terkendali hanya untuk lima layanan Go; pola rentan di repo (nilai bawaan literal, `pickle.load`, route OpenAI tanpa autentikasi di HEAD, `curl \| bash`); drift tenant dan training; tanpa uji penetrasi; tanpa register pengecualian. |
| Tindakan perbaikan | **Lihat bagian 5 (15 tindakan)** | Draf Lampiran A sampai G sebagai masukan. |
| Prioritas | **Tinggi** | Klien lembaga keuangan lazim meminta SLA patch dan hasil pemindaian; stack padat komponen berjalan pada satu host dengan akses root tanpa deteksi kerentanan. |
| Target | **[usulan] 31 Desember 2026** untuk verifikasi pengaturan GitHub, pengaktifan alert dependensi, pemindaian dasar pertama, dan VUL-001 disetujui; **[usulan] 28 Februari 2027** untuk gerbang CI, pemindaian image, inventaris host, dan register pengecualian berjalan | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "identifikasi" dan "perbaikan tepat waktu" belum terbukti; "prioritisasi" hanya terjadi ad hoc pada tinjauan satu kali. |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: VUL-001 disetujui, baseline pemindaian semua komponen terdaftar, register temuan dan pengecualian terisi, satu siklus remediasi kritis/tinggi tercatat. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 7 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.8 **berlaku** untuk AuditSphere karena:

1. **Permukaan serangan luas dan beragam ekosistem.** Repo melacak enam modul Go (`go.mod`: lima layanan dan modul `kong-gateway`), tiga set `requirements.txt` Data Hub, satu `requirements.txt` python-ai (termasuk TensorFlow, `transformers`, dan PyTorch yang dipasang terpisah), frontend Nuxt dengan rentang `^`, serta belasan image (Postgres, Redis, Kafka, Zookeeper, Kong, ClamAV, nginx, TimescaleDB, Jupyter, Prometheus, Grafana, Loki). Setiap ekosistem memiliki aliran kerentanan sendiri dan tidak ada satu pun yang dipantau secara terstruktur (bagian 4).
2. **Data klien sensitif pada satu host.** Semua komponen berjalan pada VPS yang sama dengan Postgres multi-tenant (A.8.13 bagian 1) dan akses root (A.8.19 bagian 1). Kerentanan yang dapat dieksploitasi pada gateway, layanan, atau kontainer berdampak ke data banyak klien (A.5.12: *Confidential/Restricted*).
3. **Sudah ada bukti bahwa kerentanan dan pola rentan nyata ada.** Tinjauan satu kali `docs/backend-enhancement-plan.md` (15 Oktober 2026) menemukan rahasia JWT yang dikomit, registrasi publik, berkas unggahan terbuka, dan rute tulis tanpa otorisasi; sebagian besar kini ditutup di kode (A.8.25 bagian 2.2). Pemeriksaan untuk dokumen ini menemukan pola lain yang masih ada di repo (bagian 4 butir 10). Tinjauan satu kali tidak menggantikan proses berulang: tanpa pemindaian dan SLA, kemunculan baru tidak terdeteksi.
4. **Insiden operasional yang beririsan dengan patch dan keterpaparan.** Catatan internal menyebut port internal terjangkau lewat IPv6 meski filter IPv4 berlaku (audit 5 November 2026, A.8.21 butir 2), `sshd` mengizinkan kata sandi dan root tanpa `fail2ban` (**[Catatan]**), dan reboot tak bersih membuat Kong crash loop (23 Oktober 2026). Reboot adalah konsekuensi lazim patch kernel pada host tunggal; tanpa prosedur, patch dihindari karena takut gangguan.
5. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim menanyakan frekuensi pemindaian, SLA patch per tingkat keparahan, dan cara pengecualian dikelola. `docs/DOKUMEN_KESELARASAN_ISO27001.md` memuat risiko RSK-01 ("temuan vulnerability scanner") dan rencana penambalan menyeluruh, tetapi **tidak ada keluaran pemindai atau tiket** yang dapat diserahkan; klaim itu tidak boleh dipakai di luar tanpa bukti (selaras A.5.36 dan A.8.25 bagian 4 butir 2).
6. **Multi-tenant, training, dan klon.** Tenant, training, dan dev berbagi host atau image dengan produksi; komponen rentan yang sama dapat ada pada beberapa tempat, dan training memuat salinan data nyata (bagian 4 butir 13). Kontrol ini terkait langsung dengan A.8.19 (instalasi), A.8.25 (SDLC aman), A.8.9 (konfigurasi), A.8.15/A.8.16 (pencatatan dan pemantauan), A.5.7 (intelijen ancaman), dan A.5.24 (insiden), sebagian besar masih draf.

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti manifes, kode, Dockerfile, compose, workflow, skrip, atau dokumen di repo** pada 19 November 2026. Apakah itu yang berjalan di VPS adalah **Perlu verifikasi** kecuali dinyatakan lain. Jalur instalasi dan deploy dibahas di A.8.19, gerbang CI di A.8.25, dan eksposur port serta jaringan di A.8.21; di sini hanya dirujuk.

### 2.1 Inventaris komponen dan versi yang terlihat di repo (bahan awal untuk identifikasi)

| Area | Yang terlihat | Bukti | Catatan untuk manajemen kerentanan |
| :--- | :--- | :--- | :--- |
| Go (lima layanan) | `go` pada `go.mod`: analytics 1.25.0, auth 1.25.0, master 1.25.0, risk 1.25.3, audit 1.25.8. Ukuran `go.sum` (baris): auth 749, audit 782, master 717, risk 232, analytics 162. CI memakai `setup-go@v5` dengan `go-version: '1.25'` dan `check-latest: true`. | `backend/*-service/go.mod`, `go.sum`; `.github/workflows/deploy.yml` | Dependensi terkunci per modul, tetapi versi toolchain berbeda antarlayanan. Binary dikompilasi di runner/laptop; `deploy-prod.sh` mencatat SHA commit dan hash binary tetapi **tidak** versi Go yang dipakai (diperiksa), sehingga tidak diketahui apakah binary di produksi dibangun dengan toolchain terbaru. |
| Modul `kong-gateway` | `go 1.21`; Dockerfile membangun dengan `golang:1.21-alpine` lalu `alpine:latest`; pengguna non-root (`USER appuser`). Compose produksi/dev hanya memasang direktori konfigurasi deklaratif `kong-gateway/kong/...` ke image `kong:3.8`; compose tidak membangun modul Go ini. | `backend/kong-gateway/go.mod`, `Dockerfile`; `backend/docker-compose.prod.yml` | Apakah modul Go ini dipakai di produksi: **Perlu verifikasi**. Bila tidak dipakai, kandidat dihapus dari pemindaian; bila dipakai, versi 1.21 perlu dinilai terhadap jadwal dukungan (**Perlu verifikasi**). |
| python-ai | `python:3.10-slim`; lapisan awal memasang `pip`, `setuptools`, `wheel`, dan beberapa paket **tanpa versi**; `numpy==1.26.4`; torch/torchvision/torchaudio CPU `2.2.2`/`0.17.2`/`2.2.2` lewat `--find-links` ke mirror PyTorch; `requirements.txt` 14 baris dipin `==` (antara lain `fastapi 0.115.6`, `tensorflow-cpu 2.18.0`, `transformers 4.47.1`, `xgboost 2.1.3`, `Pillow 11.0.0`, `pypdf 5.1.0`); tanpa hash. Berjalan sebagai root (tidak ada `USER`). | `backend/python-ai/Dockerfile`, `requirements.txt` | Permukaan besar (TensorFlow, PyTorch, `transformers`) dan memproses berkas unggahan (PDF, DOCX, gambar). Dependensi transitif dan lapisan awal tidak dipin. |
| Data Hub (Python) | `python:3.10-slim` (`data-hub-api`, `ai-engine`), `jupyter/scipy-notebook:python-3.10` (`caatt-engine`, ada `USER root` saat build). Pin lebih lama daripada python-ai: `fastapi 0.111.0`, `uvicorn 0.30.1`, `pandas 2.2.2`, `torch 2.3.1`, `scikit-learn 1.5.1`; rentang `>=` pada `psycopg`, `psycopg-pool`, `orjson`, `pymysql`, `cryptography`, `pytest` (di daftar runtime). | `data-server/*/Dockerfile`, `requirements.txt` | Versi pustaka yang sama berbeda antara python-ai dan Data Hub (mis. `fastapi`, `pandas`); kerentanan satu sisi tidak otomatis tertangkap di sisi lain. |
| Frontend | `node:20-alpine`, `USER node`, hanya menyalin `.output`; `package.json` memakai rentang `^`, `@nuxt/devtools: latest`, `nuxt ^4.1.2`; **dua** lockfile (`package-lock.json`, `bun.lock`); dependensi `openai` dan modul `nuxt-chatgpt` masih terdaftar. Ada juga `package.json` akar (`hostinger-api-mcp ^1.8.0`) dan `package-lock.json` akar, serta `backend/package-lock.json` **tanpa** `package.json` terlacak di folder itu. | `frontend/Dockerfile`, `package.json`, `package-lock.json`, `bun.lock`; `package.json`, `package-lock.json`, `backend/package-lock.json` (akar) | Hasil build tidak terikat ke lockfile tertentu (A.8.19 bagian 4 butir 7); lockfile yatim tidak dapat diaudit. |
| Image layanan Go | `alpine:latest` pada lima Dockerfile layanan dan tahap runtime `kong-gateway`; `apk add ca-certificates tzdata` tanpa pin; tidak ada `USER` (root, `WORKDIR /root`). | `backend/*-service/Dockerfile` | Isi image berubah tiap build tanpa perubahan repo; tidak ada jadwal rebuild untuk menarik patch base image. |
| Image infrastruktur produksi | `postgres:16`, `redis:7-alpine`, `confluentinc/cp-zookeeper:7.6.1`, `confluentinc/cp-kafka:7.6.1`, `kong:3.8`, ClamAV dipin **digest SHA256**. | `backend/docker-compose.prod.yml` | Tag mayor/minor mengambang (Postgres, Redis, Kong) vs. pin keras (ClamAV). Pin digest ClamAV mengunci **mesin**, bukan signature (bagian 2.2). |
| Image non-produksi dan tenant | Templat tenant: `rb_audit_*:latest` dan `kong:3.4`; compose lokal: `kong:3.4`, `provectuslabs/kafka-ui:latest`; compose dev: `kong:3.8`, `redis/redisinsight:latest`, `kafka-ui:latest`; training: `kong:3.8` dan image hasil re-tag. | `scripts/templates/docker-compose.tenant.yml.tpl`; `backend/docker-compose.yml`, `docker-compose.dev.yml`, `backend/training/docker-compose.training.yml` | Kong 3.4 (tenant/lokal) berbeda dari 3.8 (produksi). Dukungan versi: **Perlu verifikasi**. |
| Monitoring | `prom/prometheus:latest`, `grafana/grafana:latest`, `prom/node-exporter:latest`; `grafana/loki:3.7.8`, `grafana/promtail:3.6.8` dipin. | `monitoring/docker-compose.monitoring.yml` | `latest` berubah pada setiap `pull`; tidak ada catatan versi. |
| Data Hub (image) | `nginx:1.27-alpine`, `postgres:16-alpine`, `timescale/timescaledb:2.17.2-pg16`. | `data-server/docker-compose.yml` | Dipin lebih baik daripada produksi; tidak ada inventory atau pemilik. |
| GitHub Actions | `actions/checkout@v4`, `actions/setup-go@v5`, `appleboy/ssh-action@v1.0.3` (dipin tag, bukan SHA). | `.github/workflows/*.yml` | Rantai pasok CI; tidak ada Dependabot untuk `github-actions`. |
| Host OS | Komentar runbook menyebut VPS **AlmaLinux 9** (`docs/backup-b2-setup.md:31`) sedangkan panduan demo memakai `apt-get` (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:264-293`); skrip server Data Hub memakai `dnf update -y` dan repo Docker CE sekali pada setup (`data-server/setup_server.sh:25-38`). | dokumen dan skrip tersebut | Dokumen saling bertentangan; tidak ada daftar paket host maupun kebijakan pembaruan (bagian 4 butir 7). |

Jumlah dependensi diambil dari `go.sum` dan manifes; **tidak dijalankan pemindai** sehingga tidak ada klaim tentang status kerentanan komponen di atas.

### 2.2 Kendali yang ada dan relevan dengan pengurangan kerentanan

| Kendali | Implementasi terverifikasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Kunci dependensi Go | `go.sum` per layanan memastikan unduhan yang sama diverifikasi hash-nya. | `backend/*/go.sum` | Mengunci versi, bukan memeriksa kerentanan. |
| Pin sebagian image | Kafka/Zookeeper 7.6.1, Loki/Promtail, TimescaleDB dipin versi; ClamAV dipin digest. | compose terkait (bagian 2.1) | Tidak ada daftar yang menyatakan pin ini disengaja (A.8.19 Lampiran B). |
| Malware scanning unggahan | Klien ClamAV (INSTREAM), mode `required` bawaan (fail-closed), di audit-service dan master-service dengan test; `clamav` tidak dipublikasikan ke host; komentar compose menyatakan **freshclam memperbarui signature di dalam kontainer** dan basis data berada di volume `clamav_db`. | `backend/audit-service/pkg/avscan/`, `backend/master-service/pkg/avscan/`; `backend/docker-compose.prod.yml:146-158`; `backend/clamav/clamd.conf` | Memeriksa berkas klien, bukan kerentanan sistem. Kesegaran signature tidak dipantau (bagian 4 butir 8). |
| Penguatan hasil tinjauan | Gerbang rahasia JWT saat start, penyematan algoritma JWT, penutupan registrasi publik, `/uploads` terautentikasi dan terkurung, Viewer hanya-baca, MFA terenkripsi; ada test. | A.8.25 bagian 2.2 | Hasil tinjauan satu kali, bukan proses berulang. |
| Pengurangan keterpaparan | Port terpublikasi diikat `127.0.0.1` di compose produksi; skrip firewall host memblokir port internal (IPv4 dan IPv6) pada antarmuka publik dan LAN penyedia. | `backend/docker-compose.prod.yml`; `backend/scripts/ops/auditsphere-firewall.sh` | Terpasang dan aktif di VPS: **Perlu verifikasi** (A.8.21 butir 2). |
| Kong sebagai lapisan pelindung | Rate limit global dan per rute, penolakan agen pengguna pemindai (`sqlmap`, `nikto`, `nmap`), CORS satu origin. | `backend/kong-gateway/kong/prod/kong.yml` (A.8.21 bagian 2.1) | Kendali kompensasi berbasis agen pengguna yang mudah dilewati; bukan pengganti patch. |
| Pemisahan peran dan jalur deploy terkendali | `deploy-prod.sh` untuk lima layanan Go: build dari commit, checksum binary, dump pra-deploy, rollback otomatis. | `backend/deploy-prod.sh`; A.8.19 bagian 2.1 | Menyediakan **jalur patch** untuk lima layanan saja; python-ai, Kong, ClamAV, frontend, Postgres, Redis, Kafka, tenant, monitoring tidak memiliki jalur setara. |
| Rencana dan referensi dokumen | `docs/backend-enhancement-plan.md` butir A12 (gerbang CI dengan `govulncheck` dan `gitleaks`, USER non-root, pin digest) dan C1 (SBOM, cosign); BBP-001, RTP-001, TIP-001 (docx v1.0). | `docs/backend-enhancement-plan.md:125,152`; `docs/ISO_27001/advanced_security/` | Rencana dan prosedur, belum dilaksanakan; status persetujuan dokumen docx tidak terlihat dari isi yang dibaca. |
| Jejak penggunaan alat keamanan oleh pengembang | Komentar `//nolint:gosec` pada `backend/master-service/pkg/datasource/ssl.go:87,106` menunjukkan `gosec` dikenal oleh pengembang. | berkas tersebut | Tidak ada konfigurasi `golangci-lint`, hasil, atau langkah CI yang terlacak; bukan bukti pemindaian. |

### 2.3 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `.github/`, `backend/`, `frontend/`, `scripts/`, `monitoring/`, `data-server/`, `docs/` (19 November 2026) tidak menemukan: `govulncheck`, `npm audit`, `pip-audit`, `osv-scanner`, `trivy`, `grype`, `gosec`, `semgrep`, `CodeQL`, `gitleaks`/`trufflehog`, `snyk`; `dependabot.yml` atau konfigurasi Renovate; `dnf-automatic`, `unattended-upgrades`, atau konfigurasi pembaruan otomatis host; `lynis`, OpenSCAP, atau pemindaian konfigurasi host; SBOM (CycloneDX/SPDX); `SECURITY.md`, `security.txt`, atau saluran pelaporan kerentanan publik; kebijakan atau prosedur manajemen kerentanan; kelas keparahan dan SLA patch yang disetujui; register temuan, tiket remediasi, atau register pengecualian; laporan pemindaian, uji penetrasi, atau DAST; langganan atau catatan pemantauan advisori pemasok; `rule_files` dan Alertmanager pada Prometheus (A.8.21 bagian 2.3); daftar paket host dan versi kernel. Catatan penting: **ketiadaan `dependabot.yml` tidak membuktikan Dependabot alerts nonaktif**, karena alert dan security updates dapat diaktifkan dari pengaturan repositori GitHub tanpa berkas; status ini **Perlu verifikasi**. Seluruh keadaan VPS dan GitHub **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kebijakan manajemen kerentanan (disetujui) | **Belum tersedia; draf di Lampiran A** | Kode usulan VUL-001; angka [usulan]. |
| Patch SLA | **Belum tersedia sebagai komitmen; draf di Lampiran A bagian A.4** | BBP-001 dan RTP-001 (Draft docx) memuat tabel SLA remediasi 7/14/30/90 hari kalender; belum disetujui dan belum ditautkan ke komponen. |
| Scan report (dependensi, image, host, rahasia, SAST) | **Belum tersedia** | Tidak ada pemindaian. Lingkup dan opsi alat di Lampiran B; hasil dasar (baseline) perlu dijalankan lokal oleh pihak berwenang. |
| Remediation ticket | **Belum tersedia; templat di Lampiran C** | Bukti perbaikan yang ada berupa commit dan catatan perubahan (`docs/security-hardening-changes-backend.md`), bukan tiket dengan SLA. |
| Exception register | **Belum tersedia; templat di Lampiran D** | Tidak ada pengecualian tertulis; penerimaan risiko saat ini implisit (bagian 4 butir 18). |
| Inventory/SBOM sebagai dasar pemindaian | **Sebagian** | Manifes dan lockfile ada; SBOM dan inventory host tidak ada (A.8.19 Lampiran C menyediakan templat). |
| Sumber informasi kerentanan dan alur triase | **Belum tersedia; draf di Lampiran E** | TIP-001 (Draft) menyebut BSSN, OTX, MISP dan umpan STIX/TAXII; tidak ada bukti langganan atau pelaksanaan. |
| Pengaturan GitHub (Dependabot alerts, secret scanning, branch protection) | **Belum tersedia** | Perlu tangkapan layar/ekspor bertanggal oleh pihak berwenang. |
| Bukti patch OS/kernel host dan paket Docker, nginx | **Belum tersedia / Perlu verifikasi** | Ekspor read-only dari VPS oleh pihak berwenang (daftar paket, pembaruan tertunda, uptime). |
| Bukti kesegaran signature ClamAV | **Belum tersedia / Perlu verifikasi** | Keluaran `clamd`/`freshclam` pada kontainer yang berjalan. |
| Bukti patch endpoint personel | **Belum tersedia** | Lihat A.8.1 (tidak ada MDM/EDR, laporan endpoint belum ada). |
| Laporan uji penetrasi / DAST | **Belum tersedia** | Lihat A.8.25 tindakan 13. |
| Rekaman tinjauan berkala | **Belum tersedia; daftar periksa di Lampiran F** | Belum ada periode tinjauan. |
| Bukti penguatan hasil tinjauan | **Tersedia** | A.8.25 bagian 2.2; status deploy ke produksi: **Perlu verifikasi**. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari memori kerja internal (tanggal dicantumkan) dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan.

1. **[Tidak ada] Tidak ada proses manajemen kerentanan teknis.** Tidak ada kebijakan, pemilik per komponen, pemindai, SLA patch yang disetujui, tiket, register temuan, atau register pengecualian. Dokumen ini adalah penilaian awal; draf Lampiran A sampai G adalah masukan.
2. **[Tidak ada] Tidak ada pemindaian otomatis di CI.** Tiga workflow; hanya `route-coverage.yml` yang memeriksa sesuatu (cakupan rute Kong), dan `deploy.yml` hanya mengkompilasi dan men-deploy. Tidak ada SCA, pemindaian image, secret scan, atau SAST (A.8.25 bagian 4 butir 1 dan 2; A.8.19 butir 5). `docs/backend-enhancement-plan.md` butir A12 sudah merekomendasikan `govulncheck` dan `gitleaks` tetapi belum dilaksanakan. Apakah Dependabot alerts, security updates, atau secret scanning aktif di pengaturan GitHub: **Perlu verifikasi**.
3. **[Repo] Ekosistem dependensi luas, tidak seragam, sebagian longgar.** Enam `go.mod` dengan versi `go` berbeda (1.25.0 sampai 1.25.8, plus 1.21 pada `kong-gateway`); python-ai dan Data Hub memakai pin berbeda untuk pustaka yang sama dan Data Hub memakai rentang `>=` termasuk `cryptography` dan `pytest` pada daftar runtime; python-ai memasang paket tanpa versi pada lapisan awal; frontend memakai rentang `^`, `@nuxt/devtools: latest`, dua lockfile, serta lockfile yatim di akar dan `backend/`. Tanpa hash pada `requirements.txt`. Dampak: tidak ada satu sumber kebenaran tentang versi yang berjalan, sehingga advisori tidak dapat dipetakan cepat ke komponen terdampak.
4. **[Repo] Base image mengambang dan kontainer berjalan sebagai root.** `alpine:latest` pada lima Dockerfile layanan Go dan tahap runtime `kong-gateway`; `python:3.10-slim` dan `node:20-alpine` hanya tag; `jupyter/scipy-notebook:python-3.10`; Prometheus, Grafana, node-exporter, dan image tenant `:latest`; tanpa `USER` pada layanan Go, python-ai, `data-hub-api`, dan `ai-engine` (hanya frontend dan `kong-gateway` yang non-root; `caatt-engine` kembali ke `jovyan`). Satu sisi baik: image yang di-rebuild menarik patch base image tanpa tindakan, sisi lain: itu terjadi tidak terjadwal dan tidak tercatat, dan eksploitasi pada layanan berjalan memberi akses root dalam kontainer. Rencana A12 sudah menyebut `USER` non-root dan pin digest.
5. **[Repo]+Perlu verifikasi Komponen yang mungkin melewati dukungan.** A.8.19 bagian 2.3 mencatat bahwa Python 3.10 dan Node 20 mendekati atau melewati akhir dukungan menurut jadwal publik (belum diverifikasi), dan `go 1.21` pada `kong-gateway`, Kong 3.4 pada templat tenant dan compose lokal, serta versi torch 2.2.2/2.3.1 belum dinilai. **Seluruh pernyataan akhir dukungan di dokumen ini adalah Perlu verifikasi**; tidak ada kolom EOL di inventory. Tindakan: tambahkan kolom EOL berdasar sumber resmi pada inventory (A.8.19 Lampiran C).
6. **[Repo] Toolchain Go tidak tercatat per artefak.** CI memakai `check-latest: true` dan versi `go` per modul berbeda; `deploy-prod.sh` mencatat SHA dan hash binary tetapi tidak versi Go. Kerentanan pada pustaka standar Go diperbaiki dengan membangun ulang memakai toolchain baru; tanpa catatan versi toolchain per rilis, tidak dapat dibuktikan bahwa binary produksi sudah dibangun ulang setelah sebuah rilis patch bahasa.
7. **[Catatan]+[Tidak ada] Patch OS, kernel, dan paket host tidak terbukti.** Repo tidak memuat konfigurasi pembaruan otomatis (`dnf-automatic`, `unattended-upgrades`), daftar paket, atau kebijakan reboot; yang ada hanyalah langkah sekali pada saat setup (`dnf update -y` pada skrip Data Hub; `apt-get upgrade -y` pada panduan demo yang bertentangan dengan catatan AlmaLinux 9). Tidak ada catatan operasional yang menyebut unattended-upgrades atau sejenisnya terpasang, sehingga **tidak boleh diklaim ada**; status aktual **Perlu verifikasi** di VPS dan server Data Hub. Komponen host: kernel, Docker Engine (repo Docker CE pada Data Hub), nginx, certbot, `rclone`, `age`, `wg` (WireGuard), `sshd`. Patch kernel menuntut reboot pada host tunggal tanpa HA (A.8.21 butir 13) dan reboot tak bersih pernah memicu Kong crash loop (**[Catatan]** 23 Oktober 2026; mitigasi `tmpfs` ada di compose); tanpa jendela pemeliharaan yang disepakati, patch kernel cenderung tertunda. `docs/backup-b2-setup.md:30` masih memasang `rclone` lewat `curl -fsSL ... | bash` (A.8.19 butir 9).
8. **[Repo]+[Catatan] ClamAV: signature diperbarui oleh kontainer tetapi tidak dipantau.** Compose produksi memakai image dipin digest dan komentar menyatakan freshclam memperbarui signature di volume `clamav_db`; tidak ada `freshclam.conf` terlacak, tidak ada pemantauan umur signature atau kegagalan pembaruan (Prometheus tidak men-scrape ClamAV, A.8.21 butir 4), dan pin digest berarti versi **mesin** ClamAV tidak berubah kecuali digest diganti. Mode `required` bersifat fail-closed: ClamAV mati menolak unggahan (terlihat), tetapi signature basi hanya menurunkan deteksi tanpa gejala. **[Catatan]** 4 November 2026: ClamAV pernah dijalankan manual dengan `docker run` di luar compose dan perbaikan `::1` ada di `clamd.conf`; apakah yang berjalan sekarang kontainer compose atau manual, serta tanggal signature terbaru: **Perlu verifikasi**. Pembaruan signature membutuhkan akses keluar ke server cermin ClamAV (**Perlu verifikasi**).
9. **[Repo]+[Catatan] Keterpaparan jaringan sebagai faktor prioritisasi.** Perbaikan di repo (port loopback pada `docker-compose.prod.yml`, `auditsphere-firewall.sh` dengan `ip6tables`) menutup keterpaparan port internal lewat IPv6 yang ditemukan pada audit 5 November 2026; apakah sudah terpasang di VPS: **Perlu verifikasi** (A.8.21 butir 2). `backend/docker-compose.yml` (lokal, dengan perubahan belum dikomit) masih memakai `"port:port"` tanpa loopback; Kong admin tidak berautentikasi dan terjangkau dari kontainer mana pun di bridge (A.8.21 butir 5); API dan AI Engine Data Hub terbit di semua antarmuka dengan firewalld sebagai pembatas (A.8.21 butir 6). Keterpaparan ini harus menjadi masukan skor prioritas (Lampiran A bagian A.4), bukan dibahas ulang di sini.
10. **[Repo] Pola rentan terverifikasi dari repo.** Tabel berikut mencatat **pola**, bukan CVE; masing-masing perlu dimasukkan ke register temuan awal (Lampiran C).

    | No | Pola | Bukti | Status dan tindak lanjut |
    | :--- | :--- | :--- | :--- |
    | a | Nilai bawaan literal untuk kunci API (python-ai, Data Hub) pada compose produksi dan pada kode python-ai (komentar menyatakan fail-closed, tetapi `API_KEY` memiliki nilai bawaan literal sehingga kondisi "tidak diatur" tidak terjadi); kata sandi admin bawaan Grafana pada compose monitoring; kata sandi bawaan Postgres pada compose; akun bawaan pada seeder dan `onboard-tenant.sh` (nilai tidak dikutip) | `backend/docker-compose.prod.yml:123,125,204`; `backend/python-ai/main.py:68`; `monitoring/docker-compose.monitoring.yml:53-54`; A.8.25 butir 8; A.8.21 butir 7 | Instalasi tanpa override memakai nilai di repo. Nilai di produksi: **Perlu verifikasi**. Ditutup lewat A.8.9 dan A.8.19 tindakan 11. |
    | b | Deserialisasi tidak aman: `pickle.load` pada bundel model saat start (`main.py`) dan pada `retrain_engine.py`; `torch.load` tanpa argumen `weights_only` pada dua checkpoint (torch 2.2.2, nilai bawaan diperiksa **Perlu verifikasi**). Empat bundel `*.pkl` terlacak git dan disalin ke volume oleh `entrypoint.sh`; asal dan checksum tidak dicatat | `backend/python-ai/main.py:87-153`; `retrain_engine.py:310,885-890`; `models/*.pkl` | Berkas model yang diubah dapat mengeksekusi kode saat dimuat. Selaras A.8.25 butir 11. Kendali integritas dan penggantian format. |
    | c | Route Nitro `frontend/server/api/openai.ts` pada HEAD meneruskan `prompt` dari klien ke OpenAI memakai kunci server tanpa autentikasi terlihat; berstatus dihapus hanya di working tree; dependensi `openai` dan modul `nuxt-chatgpt` masih ada | `git show HEAD:frontend/server/api/openai.ts`; `frontend/package.json`; `frontend/nuxt.config.ts:37` | Komit penghapusan dan bersihkan dependensi (A.8.25 tindakan 6). Apakah ada pada build produksi: **Perlu verifikasi**. |
    | d | Rahasia JWT pernah dikomit; digest dipakai menolak nilai itu saat start, nilai lama tetap di riwayat git | `backend/*/pkg/config/jwt_secret.go`; `docs/jwt-rotation-runbook.md` | Rotasi dan pemindaian riwayat belum terbukti (A.8.25 butir 6). |
    | e | Instalasi host lewat skrip internet (`curl ... \| bash`) pada runbook backup | `docs/backup-b2-setup.md:30` | A.8.19 butir 9. |
    | f | Berkas terlacak yang berpotensi sensitif: `frontend/.env.production`, `claude_code_env.sh`, `.claude/settings.json`, dokumen LHA `.docx` (isi tidak dibuka) | A.8.25 butir 7 | Pemeriksaan oleh pihak berwenang tanpa mencetak nilai. |
    | g | Postgres `sslmode` bawaan `disable`, Kafka `PLAINTEXT`, Redis tanpa `requirepass`, ClamAV dan Kong admin tanpa autentikasi (bergantung isolasi bridge) | A.8.21 butir 6 | Dicatat di A.8.21; masuk register sebagai risiko yang diterima sementara bila tidak segera diubah (Lampiran D). |
    | h | Kontainer layanan berjalan sebagai root (butir 4) | Dockerfile | Rencana A12. |
    | i | `InsecureSkipVerify: true` pada konfigurasi TLS koneksi sumber data, dengan verifikasi rantai pada `VerifyConnection` pada salah satu cabang dan `MinVersion` TLS 1.2 | `backend/master-service/pkg/datasource/ssl.go:87,106` | Tampak disengaja (semantik libpq `require`); perlu ditinjau oleh Backend Lead dan dicatat sebagai keputusan, bukan diasumsikan aman. |

11. **[Repo] Jalur patch itu sendiri berisiko dan tidak merata.** Hanya lima layanan Go memiliki jalur terkendali (`deploy-prod.sh`); python-ai, Kong, ClamAV, frontend, Postgres, Redis, Kafka, tenant, dan monitoring dipatch lewat jalur manual (A.8.19 bagian 4 butir 1). Seeder pada start produksi di jalur lama dan unggahan yang tidak dipersist pada recreate (**[Catatan]** 4 Oktober 2026) membuat "tambal dengan membangun ulang" menimbulkan risiko kehilangan data, sehingga patch cenderung tertunda. `deploy-prod.sh` juga tidak melewatkan override python-ai (**[Catatan]** 4 November 2026). Gerbang patch yang dapat dipercaya adalah prasyarat SLA yang realistis.
12. **[Repo] Frontend: pemindaian dan keterikatan build.** Build dilakukan di laptop dan `.output` diunggah; ada dua lockfile sehingga tidak jelas pohon dependensi mana yang dibangun (A.8.19 butir 7). `npm audit` atas keluaran apa pun belum pernah terdokumentasi.
13. **[Repo]+[Catatan] Drift tenant, training, dan dev.** Templat tenant memakai `rb_audit_*:latest` dan `kong:3.4` sedangkan produksi `kong:3.8`; hanya satu tenant terlacak (`bai`, `registry.tsv`) dan registri tidak mencatat versi (A.8.19 butir 12). Isolasi tenant melarang berbagi rahasia atau kunci (**[Catatan]** tenant isolation gotchas), sehingga patch dilakukan per tenant dengan rekaman sendiri. Training adalah re-tag image produksi per 5 November 2026 (`auditsphere-training/<svc>:20261105`) dengan salinan data produksi dan hostname publik: patch pada produksi **tidak** mengalir ke training kecuali `reset-training.sh --with-images`. Stack dev dihentikan 5 November 2026 tetapi hidup lagi pada push ke `dev`/`develop` (A.8.19 butir 11). Setiap lingkungan harus masuk cakupan pemindaian dan SLA yang sama bila memuat data nyata (A.8.31, draf).
14. **[Repo] Data Hub sebagai komponen tersendiri.** `jupyter/scipy-notebook:python-3.10` mengambang, JupyterLab diakses lewat nginx Basic Auth, API dan AI Engine diterbitkan pada semua antarmuka dengan firewalld sebagai pembatas, basis data CBS simulasi dan datalake Timescale pada host terpisah dengan patch OS sendiri (`dnf`) dan WireGuard. Tidak ada inventory, pemilik, atau pemindaian untuk server ini; status paket host: **Perlu verifikasi**.
15. **[Repo] Pemantauan tidak mendeteksi kondisi terkait kerentanan.** Tidak ada `rule_files` dan Alertmanager (A.8.21 butir 4); tidak ada pemeriksaan umur signature ClamAV, kedaluwarsa sertifikat TLS, status "pembaruan tertunda" host, atau uptime kernel; retensi metrik dan log 7 hari. Tanpa A.8.15/A.8.16 (draf), tidak ada deteksi percobaan eksploitasi atas komponen yang belum ditambal selain yang kebetulan terlihat di log.
16. **[Repo] Dokumen yang ada belum menjadi proses, dan angkanya tidak konsisten.** BBP-001 (Program Pengungkapan Kerentanan dan Bug Bounty) menyebut A.8.8 dan ISO/IEC 29147, dengan SLA remediasi 7/14/30/90 hari kalender (Critical/High/Medium/Low) serta SLA pengakuan dan triase; RTP-001 memuat tabel SLA remediasi yang sama; TIP-001 menyebut sumber (BSSN, OTX, MISP) dan umpan otomatis. Namun tidak ada bukti program berjalan: tidak ada `SECURITY.md`/`security.txt`, saluran pelaporan, cakupan publik, atau catatan laporan. Annex lain memakai angka berbeda (A.8.19 Lampiran A dan A.8.25 tindakan 14: kritis 7, tinggi 30, sedang 90 hari); angka perlu diseragamkan (Lampiran A bagian A.4 mengusulkan memakai tabel BBP-001/RTP-001). `DOKUMEN_KESELARASAN_ISO27001.md` (RSK-01, bagian 9.3) menyebut temuan *vulnerability scanner* dan penambalan menyeluruh tanpa keluaran pemindai di repo; tabel keparahan insiden di dokumen itu (`:455-459`) mengatur respons insiden, bukan batas waktu patch.
17. **[Tidak ada] Tidak ada uji penetrasi, DAST, atau latihan red team yang terekam.** RTP-001 adalah rencana; `docs/backend-enhancement-plan.md` butir C5 hanya merekomendasikan. Pemindaian dinamis (DAST, termasuk terautentikasi) dan uji penetrasi terhadap lingkungan sintetis perlu dijadwalkan (A.8.25 tindakan 13) agar kerentanan logika bisnis dan isolasi lintas-tenant ikut tercakup; pemindai dependensi tidak menangkapnya.
18. **[Repo] Tidak ada mekanisme pengecualian atau penerimaan risiko.** Beberapa kondisi sudah diketahui dan diterima secara implisit (kontainer root, Postgres tanpa TLS, nilai bawaan, pin mengambang); tidak ada pemilik, alasan, kendali pengganti, atau tanggal kedaluwarsa. Penerimaan risiko semacam ini harus tertulis dan kedaluwarsa (Lampiran D).
19. **[Repo] Patch endpoint personel di luar bukti repo.** A.8.1 menyatakan tidak ada standar endpoint, MDM/EDR, atau laporan status patch OS dan aplikasi (A.8.1 bagian 2.1 dan 4). Kerentanan pada laptop pengembang (kunci SSH root, token GitHub, kunci layanan; A.8.1 butir 4) adalah jalur ke produksi. Tidak diduplikasi di sini; kendali endpoint dikerjakan melalui A.8.1 dan A.6.7, dan Lampiran F memuat satu butir verifikasi. Patch endpoint **klien** adalah tanggung jawab klien dan tidak tercakup, kecuali agen atau aplikasi yang AuditSphere sediakan untuk dipasang di sisi klien (tidak ditemukan dalam repo).
20. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet; tidak ada akses SSH atau GitHub API; tidak ada pemindai yang dijalankan dan tidak ada alat yang dipasang.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Verifikasi pengaturan GitHub secara read-only** oleh pihak berwenang: Dependabot alerts dan security updates, secret scanning dan push protection, code scanning, daftar akses, branch protection; aktifkan alert dependensi bila belum (perubahan pengaturan atas persetujuan). Tanpa menampilkan nilai secret. Satukan dengan A.8.25 tindakan 1. | Lead Information Security, DevOps Lead | Tinggi | [usulan] 30 November 2026 | Tangkapan layar/ekspor bertanggal (EV-A8.8-05) |
| 2 | **Setujui kebijakan VUL-001** (Lampiran A): kelas keparahan, SLA patch (seragamkan dengan BBP-001/RTP-001), peran, pengecualian, dan jadwal pemindaian; tetapkan jendela pemeliharaan untuk patch host dan reboot. | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | Kebijakan disetujui (EV-A8.8-02) |
| 3 | **Jalankan pemindaian dasar (baseline) lokal** oleh pihak berwenang pada repo dan artefak lokal (bukan produksi/host hidup): `govulncheck` per modul Go, `npm audit`/`osv-scanner` untuk frontend, `pip-audit` untuk python-ai dan Data Hub, `trivy`/`grype` pada image yang di-build lokal, dan `gitleaks` pada riwayat penuh (keluaran tersanitasi). Pasang alat hanya atas persetujuan; hasil masuk register temuan. Lingkup dan opsi di Lampiran B. | Lead Information Security, `auditsphere-devops` | Tinggi | [usulan] 31 Desember 2026 | Laporan baseline tersanitasi (EV-A8.8-03) |
| 4 | **Buka register temuan dan register pengecualian** (Lampiran C dan D); masukkan pola rentan pada bagian 4 butir 10, temuan baseline, dan temuan `backend-enhancement-plan.md` yang belum diverifikasi ulang; tetapkan pemilik dan tenggat sesuai VUL-001. Gabungkan dengan A.8.25 tindakan 14 dan A.5.36. | Lead Information Security | Tinggi | [usulan] 15 Januari 2027 | Register terisi (EV-A8.8-04, EV-A8.8-06) |
| 5 | **Verifikasi status host secara read-only** (hanya atas konfirmasi eksplisit): daftar paket dan pembaruan keamanan tertunda, versi kernel dan uptime, versi Docker/nginx/certbot/`rclone`/`age`, ada tidaknya pembaruan otomatis, konfigurasi `sshd`; juga server Data Hub. Tidak mengubah paket pada langkah ini. Satukan dengan A.8.19 tindakan 10 dan A.8.21 tindakan terkait. | DevOps Lead, `auditsphere-platform` | Tinggi | [usulan] 31 Desember 2026 | Ekspor tersanitasi (EV-A8.8-07) |
| 6 | **Verifikasi ClamAV pada kontainer yang berjalan** (read-only): versi mesin, tanggal signature terakhir, keberhasilan pembaruan, apakah kontainer berasal dari compose atau `docker run` manual; putuskan kebijakan pembaruan digest (mis. [bulanan]) dan tambahkan pemeriksaan umur signature ke pemantauan. | DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Catatan verifikasi; keputusan kebijakan (EV-A8.8-08) |
| 7 | **Tutup pola rentan berisiko tinggi**: komit penghapusan `frontend/server/api/openai.ts` dan dependensi `openai`/`nuxt-chatgpt` (A.8.25 tindakan 6), hapus nilai bawaan literal kunci dan kata sandi dari jalur bawaan (A.8.25 tindakan 9; A.8.19 tindakan 11), ganti `curl \| bash` pada runbook rclone, kendalikan integritas model sebelum `pickle.load` (A.8.25 tindakan 11). Perubahan kode/konfigurasi oleh agen backend/frontend/devops atas persetujuan, satu perubahan per sesi. | `auditsphere-backend`, `auditsphere-frontend`, `auditsphere-devops` | Tinggi | [usulan] 31 Januari 2027 | Commit; temuan di register ditutup dengan bukti |
| 8 | **Tambahkan gerbang CI untuk kerentanan** pada push `staging` dan sebagai prasyarat `deploy.yml`: `govulncheck`, `npm audit`/`osv-scanner`, `pip-audit`, pemindaian image, `gitleaks`; ambang blok [usulan: kritis/tinggi memblokir promosi kecuali ada pengecualian tertulis]. Satu pekerjaan bersama dengan A.8.25 tindakan 3 dan 4 dan A.8.19 tindakan 9. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 28 Februari 2027 | Workflow terbaru; log run (EV-A8.8-09) |
| 9 | **Otomatiskan pembaruan dependensi**: Dependabot (atau Renovate) untuk `gomod`, `npm`, `pip`, `docker`, dan `github-actions` dengan pengelompokan dan jadwal [mingguan]; pin GitHub Actions ke SHA. Perubahan workflow/konfigurasi atas persetujuan. | `auditsphere-devops` | Sedang | [usulan] 31 Januari 2027 | `dependabot.yml`; riwayat PR/alert |
| 10 | **Seragamkan toolchain dan base image**: satu versi `go` (dan catat versi toolchain per rilis di catatan deploy), satu lockfile frontend, pin versi/digest image (ganti `alpine:latest`, `:latest` monitoring dan tenant), seragamkan Kong, putuskan nasib modul `kong-gateway` (hapus dari pemindaian bila tidak dipakai), tambahkan `USER` non-root. Satu pekerjaan dengan A.8.19 tindakan 7. | `auditsphere-devops`, `auditsphere-backend` | Sedang | [usulan] 28 Februari 2027 | Dockerfile/compose terbaru; daftar pin |
| 11 | **Tambahkan kolom EOL pada inventory** berdasarkan sumber resmi pemasok (Python, Node, Go, Postgres, Redis, Kafka/Confluent, Kong, Alpine, torch) dan rencana peningkatan untuk yang mendekati/melewati dukungan; verifikasi dukungan Python 3.10 dan Node 20 dahulu. | DevOps Lead, Data Analyst | Sedang | [usulan] 31 Januari 2027 | Inventory dengan kolom EOL (A.8.19 Lampiran C) |
| 12 | **Tetapkan jalur dan jendela patch** per komponen: lima layanan Go lewat `deploy-prod.sh`; tentukan jalur untuk python-ai (sertakan override), Kong, ClamAV, frontend, Postgres/Redis/Kafka, tenant, monitoring, dan host; tentukan jendela reboot; terapkan prosedur tenant per slug dengan register versi. Selaras A.8.19 tindakan 1, 12, 13. | DevOps Lead, Solution Architect | Tinggi | [usulan] 28 Februari 2027 | Prosedur patch; register versi tenant |
| 13 | **Aktifkan sumber informasi kerentanan dan pelaporan**: langganan advisori pemasok utama (Lampiran E), saluran pelaporan publik (`security.txt` dan alamat surel keamanan; keputusan pengguna), putuskan apakah BBP-001 diaktifkan; selaraskan dengan TIP-001 dan A.5.7 (draf). Perubahan situs/DNS oleh `auditsphere-platform` atas persetujuan. | Lead Information Security, `auditsphere-platform` | Sedang | [usulan] 28 Februari 2027 | Daftar langganan; halaman/berkas pelaporan |
| 14 | **Jadwalkan uji keamanan independen dan pemindaian dinamis** pada lingkungan sintetis (bukan salinan data produksi), termasuk isolasi lintas-tenant; pemindaian terautentikasi atas Kong dan frontend; tindak lanjut ke register (A.8.25 tindakan 13). | Lead Information Security | Sedang | [usulan] 28 Februari 2027 | Laporan dan rencana remediasi (EV-A8.8-10) |
| 15 | **Daftarkan Evidence ID** pada Master Audit Evidence Index, jadwalkan tinjauan berkala (Lampiran F), koreksi klaim `DOKUMEN_KESELARASAN_ISO27001.md` RSK-01 atau dukung dengan bukti, dan commit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 1, 3, 5, 6, 7, 8, 9, 10, 12, dan 13 menyentuh GitHub, VPS, workflow, atau konfigurasi dan memerlukan persetujuan eksplisit; dilakukan lewat agen devops/platform/backend/frontend, satu perubahan per sesi. Pemindaian hanya terhadap repo, image lokal, atau lingkungan sintetis; **tidak ada pemindaian terhadap produksi atau host hidup tanpa izin tertulis dan jendela yang disepakati**. Deploy ke produksi tidak boleh dipicu sebagai bagian dari pengumpulan bukti. Hasil pemindaian dapat memuat rahasia atau jalur sensitif; simpan sebagai Restricted dan sanitasi sebelum dibagikan.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.8-01 | Manifes, lockfile, Dockerfile, dan compose sebagai dasar inventory dan bukti pin/ketiadaan pin | `backend/*/go.mod`, `go.sum`; `backend/python-ai/requirements.txt`, `Dockerfile`; `data-server/*/requirements.txt`, `Dockerfile`; `frontend/package.json`, `package-lock.json`, `bun.lock`; `backend/docker-compose.prod.yml`; `monitoring/docker-compose.monitoring.yml`; `scripts/templates/docker-compose.tenant.yml.tpl` | Internal |
| EV-A8.8-02 | Kebijakan VUL-001 disetujui (setelah tindakan 2) | `docs/ISO_27001/A.8.8_Pengelolaan_Kerentanan_Teknis.md` (Lampiran A); [belum ada versi disetujui] | Internal |
| EV-A8.8-03 | Laporan pemindaian baseline tersanitasi: dependensi, image, rahasia (setelah tindakan 3) | [belum ada] | Restricted |
| EV-A8.8-04 | Register temuan (Lampiran C) dengan tiket remediasi dan bukti penutupan (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.8-05 | Pengaturan GitHub: Dependabot alerts, secret scanning, branch protection (setelah tindakan 1) | [belum ada] | Restricted |
| EV-A8.8-06 | Register pengecualian/penerimaan risiko (Lampiran D) (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.8-07 | Ekspor read-only status paket host, kernel, dan pembaruan tertunda VPS produksi dan server Data Hub (setelah tindakan 5) | [belum ada keluaran VPS] | Restricted |
| EV-A8.8-08 | Verifikasi ClamAV (versi mesin, tanggal signature) dan keputusan kebijakan pembaruan (setelah tindakan 6) | [belum ada] | Restricted |
| EV-A8.8-09 | Workflow CI dengan pemindaian dan hasil run (setelah tindakan 8) | [belum ada] | Internal |
| EV-A8.8-10 | Laporan uji penetrasi/DAST dan rencana remediasi (setelah tindakan 14) | [belum ada] | Restricted |
| EV-A8.8-11 | Dokumen terkait kerentanan: BBP-001, RTP-001, TIP-001; rencana peningkatan dan catatan perubahan keamanan | `docs/ISO_27001/advanced_security/`; `docs/backend-enhancement-plan.md`; `docs/security-hardening-changes-backend.md`; `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| EV-A8.8-12 | Catatan tinjauan berkala (Lampiran F) dan laporan metrik (Lampiran A bagian A.9) | [belum ada] | Internal |

Setelah tindakan 1 sampai 8 selesai, tambahkan ID per periode pemindaian dan per tinjauan.

---

## Lampiran A — Draf Kebijakan Pengelolaan Kerentanan Teknis (VUL-001, usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran. Instalasi dan perubahan produksi diatur di SWI-001 (A.8.19) dan SDLC-001 (A.8.25); dokumen ini merujuk, bukan mengulang.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | VUL-001 Kebijakan Pengelolaan Kerentanan Teknis (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden, perubahan arsitektur, atau penambahan komponen utama) |
| Berlaku untuk | Seluruh komponen yang berjalan dan yang dibangun: layanan Go, python-ai, frontend, image dan compose (produksi, tenant, training, dev yang memuat data nyata, monitoring), ClamAV, Data Hub, host VPS dan server Data Hub, pipeline CI, dan perangkat endpoint personel yang memegang akses produksi |

### A.1 Prinsip dan peran

1. **Tujuan.** Memastikan kerentanan teknis pada komponen yang dipakai diketahui tepat waktu, dievaluasi keterpaparannya, dan ditangani dalam batas waktu yang ditetapkan atau diterima secara tertulis.
2. **Inventory sebagai dasar.** Tidak ada pemindaian atau SLA tanpa daftar komponen dan pemiliknya (A.8.19 Lampiran C). Setiap komponen memiliki pemilik (peran).
3. **Peran.** Lead Information Security memiliki kebijakan, klasifikasi, dan persetujuan pengecualian; DevOps Lead memiliki pemindaian, patch infrastruktur, dan jalur deploy; Backend/Frontend Lead dan Data Analyst memiliki patch dependensi pada kode mereka; Solution Architect menilai dampak arsitektur dan tenant; Direktur Utama menyetujui penerimaan risiko Kritis. Detail pada RACI Lampiran G.
4. **Sumber informasi.** Sumber dan frekuensi pemantauan sesuai Lampiran E; sumber baru ditambahkan lewat Lead Information Security.
5. **Pelaporan dari luar.** Laporan eksternal (peneliti, klien) diperlakukan menurut BBP-001 bila diaktifkan dan, bila belum, melalui saluran keamanan yang ditetapkan (tindakan 13); pengakuan dalam [3 hari kerja] (usulan).

### A.2 Lingkup dan frekuensi pemindaian (usulan)

| Objek | Jenis pemindaian | Frekuensi [usulan] | Pemicu tambahan |
| :--- | :--- | :--- | :--- |
| Dependensi Go, npm, Python | SCA (`govulncheck`, `npm audit`/`osv-scanner`, `pip-audit`) | Setiap push ke `staging` (CI) dan [mingguan] terjadwal | Advisori pemasok untuk komponen inventory |
| Image (build lokal/CI) | Pemindaian image (`trivy`/`grype`) | Setiap build rilis dan [mingguan] untuk image yang berjalan | Perubahan base image |
| Repositori dan riwayat | Secret scan (`gitleaks`) | Setiap push; riwayat penuh sekali, lalu [kuartalan] | Insiden kebocoran |
| Kode sumber | SAST (`gosec`/`semgrep`) | Setiap push ke `staging` | Perubahan jalur sensitif (A.8.25) |
| Host VPS dan server Data Hub | Daftar pembaruan keamanan tertunda dan pemindaian konfigurasi (mis. `lynis`/OpenSCAP) | [mingguan] daftar pembaruan; [kuartalan] audit konfigurasi | Advisori kernel/Docker/nginx/OpenSSH |
| ClamAV | Pemeriksaan umur signature dan versi mesin | [harian otomatis] bila pemantauan aktif; [mingguan] manual sebelum itu | Kegagalan pembaruan |
| Lingkungan sintetis | DAST dan uji penetrasi | [tahunan] dan sebelum rilis mayor | Perubahan arsitektur/auth |
| Endpoint personel | Status patch OS dan aplikasi (laporan endpoint A.6.7/A.8.1) | [triwulanan] | Perangkat baru atau hilang |

### A.3 Klasifikasi keparahan (usulan)

Keparahan dasar dari skor CVSS dan advisori pemasok, lalu **disesuaikan** oleh keterpaparan dan konteks pada triase (A.5). Kelas dan rentang berikut [usulan].

| Kelas | Kriteria dasar [usulan] | Penyesuaian naik | Penyesuaian turun (harus tertulis) |
| :--- | :--- | :--- | :--- |
| **Kritis** | CVSS 9.0 sampai 10.0, atau terbukti dieksploitasi aktif, atau eksekusi kode/ bypass autentikasi jarak jauh tanpa autentikasi pada komponen yang terjangkau internet | Komponen terjangkau internet; dampak lintas-tenant; tanpa kendali kompensasi | Kode rentan tidak terjangkau (hasil *reachability*), komponen tidak dipakai atau tidak dipasang, kendali kompensasi terbukti |
| **Tinggi** | CVSS 7.0 sampai 8.9 | Kerentanan memengaruhi data klien atau rahasia; ada eksploit publik | Hanya terjangkau dari jaringan internal terpisah, autentikasi kuat wajib |
| **Sedang** | CVSS 4.0 sampai 6.9 | Rantai dengan kerentanan lain | Dampak terbatas, kode tidak terjangkau |
| **Rendah** | CVSS di bawah 4.0 atau informasional | | |

Penurunan kelas harus dicatat di tiket dengan alasan dan penyetuju (Lead Information Security). Kerentanan yang terkait rahasia yang bocor (kunci, token, kata sandi) diperlakukan sebagai Kritis dan dilakukan rotasi segera, terlepas dari skor.

### A.4 Batas waktu penanganan (patch SLA, usulan)

Dihitung sejak triase selesai (kelas ditetapkan), dalam hari kalender. Angka diseragamkan dengan tabel SLA remediasi di BBP-001 dan RTP-001 (Critical 7, High 14, Medium 30, Low 90) dan **menggantikan** angka berbeda di A.8.19 Lampiran A butir 8 dan A.8.25 tindakan 14 setelah disetujui. Seluruhnya [usulan].

| Kelas | Pengakuan dan triase | Mitigasi sementara | Perbaikan (patch atau penggantian komponen) | Verifikasi penutupan |
| :--- | :--- | :--- | :--- | :--- |
| **Kritis** | [24 jam] sejak diketahui | [24 jam] (mis. blokir jalur di firewall/Kong, matikan fitur, rotasi rahasia) | [7 hari] | Pindai ulang [dalam 2 hari kerja] setelah patch |
| **Tinggi** | [3 hari kerja] | [7 hari] bila patch belum tersedia | [14 hari] | Pindai ulang pada siklus berikut |
| **Sedang** | [7 hari kerja] | Bila perlu | [30 hari] atau pada rilis terjadwal berikutnya | Pindai ulang pada siklus berikut |
| **Rendah** | [14 hari kerja] | Tidak wajib | [90 hari] atau menjadi bagian pemeliharaan rutin | Pada siklus pemindaian berikut |

Aturan tambahan [usulan]:

- Bila kerentanan dieksploitasi aktif (sumber terpercaya) dan patch tersedia, kelas Kritis berlaku tanpa memandang skor, dan mitigasi sementara dalam [24 jam].
- Bila patch belum tersedia dari pemasok, mitigasi sementara dan pemantauan advisori [mingguan] menjadi pengganti, dengan pengecualian tertulis (Lampiran D) bukan penundaan diam-diam.
- Patch host yang memerlukan reboot dijadwalkan pada jendela pemeliharaan [mingguan/bulanan] yang disepakati; Kritis boleh di luar jendela dengan persetujuan darurat (A.8.19 Lampiran A butir 9).
- Komponen yang melewati akhir dukungan tanpa patch dicatat sebagai pengecualian dengan rencana migrasi dan tenggat; tidak dibiarkan tanpa tanggal.
- SLA berlaku sama untuk produksi, tenant, Data Hub, dan lingkungan non-produksi yang memuat data nyata (training); lingkungan sintetis dapat diberi tenggat [2x].

### A.5 Alur triase dan evaluasi keterpaparan

Alur lengkap pada Lampiran E. Pertanyaan evaluasi minimum [usulan]: apakah komponen terpasang dan versinya terdampak; apakah fungsi rentan dipakai (hasil *reachability*); apakah terjangkau dari internet, jaringan Docker, atau hanya lokal; apakah memproses atau mengakses data klien; apakah ada eksploit publik atau eksploitasi aktif; apakah ada kendali kompensasi (firewall, Kong, autentikasi, isolasi tenant); berapa tenant atau lingkungan terdampak.

### A.6 Jalur perbaikan

Perbaikan memakai jalur sah per komponen (A.8.19) dan gerbang SDLC (A.8.25): perubahan melalui commit, tinjauan, CI, deploy terkendali, dan verifikasi pasca-deploy; patch darurat mengikuti prosedur perubahan darurat. Pembaruan base image dan dependensi dicatat pada inventory dan SBOM. Patch per tenant dicatat per slug dengan register versi.

### A.7 Pengecualian dan penerimaan risiko

Kerentanan yang tidak dapat ditambal dalam SLA dicatat pada register pengecualian (Lampiran D) dengan alasan, kendali pengganti, pemilik, penyetuju, dan **tanggal kedaluwarsa** [usulan: Rendah dan Sedang 6 bulan, Tinggi 3 bulan, Kritis 30 hari dan persetujuan Direktur Utama]. Pengecualian yang kedaluwarsa tanpa tindak lanjut dieskalasi ke Direktur Utama. Tidak ada pengecualian tanpa kendali pengganti.

### A.8 Pelaporan, eskalasi, dan keterkaitan insiden

Kerentanan yang terbukti dieksploitasi atau diduga menyebabkan akses tidak sah diperlakukan sebagai insiden (A.5.24) dan mengikuti tingkat keparahan insiden di `DOKUMEN_KESELARASAN_ISO27001.md` bagian 11. Kerentanan yang memengaruhi data klien dikomunikasikan ke klien sesuai perjanjian dan peraturan yang berlaku (Legal & Compliance).

### A.9 Metrik (usulan)

Jumlah temuan terbuka per kelas; persentase temuan ditutup dalam SLA (selaras KPI di RTP-001); umur rata-rata temuan terbuka; jumlah pengecualian aktif dan yang kedaluwarsa; persentase komponen inventory yang tercakup pemindaian; umur signature ClamAV; jumlah image dengan tag `latest`; jumlah komponen pada atau melewati akhir dukungan. Dilaporkan [bulanan] ke Lead Information Security dan [kuartalan] ke manajemen.

### A.10 Pengecualian dan pelanggaran kebijakan

Pelanggaran SLA tanpa pengecualian dicatat sebagai ketidaksesuaian (A.5.36). Revisi kebijakan disetujui oleh penyetuju.

---

## Lampiran B — Lingkup Pemindaian dan Opsi Alat untuk Tim Kecil (usulan)

Tidak ada alat yang dipasang atau dijalankan untuk dokumen ini. Daftar berikut adalah **opsi** yang perlu dinilai (lisensi, biaya, kesesuaian) sebelum dipilih; ketersediaan fitur pada paket GitHub yang dipakai: **Perlu verifikasi**. Prinsip: mulai dari yang gratis, berjalan di CI atau lokal, dan menghasilkan keluaran yang dapat diarsipkan.

| Objek | Opsi alat | Catatan untuk AuditSphere |
| :--- | :--- | :--- |
| Dependensi Go | `govulncheck` (analisis keterjangkauan), `osv-scanner` | Jalankan per modul (`backend/*-service`); cocok untuk CI. Catat versi toolchain Go yang dipakai membangun. |
| Dependensi npm | `npm audit` (lockfile `package-lock.json`), `osv-scanner` | Tetapkan satu lockfile terlebih dahulu (tindakan 10); frontend dibangun di laptop sehingga pemindaian atas lockfile + keluaran build. |
| Dependensi Python | `pip-audit` (atas `requirements.txt`), `osv-scanner` | python-ai dan tiga layanan Data Hub; lapisan awal tanpa versi perlu dipin agar dapat diaudit; torch dari mirror `--find-links` mungkin tidak tercakup oleh pemindai PyPI. |
| Image kontainer | `trivy`, `grype` (juga untuk filesystem dan IaC), `docker scout` (bila sesuai) | Pindai image yang di-build lokal atau di CI; hasilkan SBOM (`syft`) sekaligus (selaras A.8.19 tindakan 8). |
| Rahasia | `gitleaks`, `trufflehog`; secret scanning dan push protection GitHub | Pemindaian riwayat penuh sekali, lokal, tanpa mencetak nilai; riwayat memuat rahasia JWT lama. |
| SAST | `gosec`, `semgrep` (Go, Python, TS), ESLint plugin keamanan, CodeQL (bila tersedia untuk repo privat: Perlu verifikasi) | Fokus aturan pada otorisasi, SQL, path traversal, deserialisasi. |
| Pembaruan otomatis dependensi | Dependabot atau Renovate | Termasuk `github-actions` dan `docker`; Dependabot alerts dapat aktif tanpa berkas konfigurasi (Perlu verifikasi). |
| Host (OS, paket) | `dnf updateinfo`/`dnf check-update --security` (RHEL-family), `dnf-automatic` untuk pembaruan keamanan otomatis; `lynis`, OpenSCAP untuk audit konfigurasi | Hanya atas persetujuan, baca-saja pada langkah inventaris; kebijakan reboot diputuskan terpisah. Server Data Hub memakai `dnf` menurut skrip; VPS produksi: Perlu verifikasi. |
| ClamAV | Pemeriksaan versi dan tanggal signature lewat `clamdscan --version`/log `freshclam` pada kontainer | Tambahkan metrik/alert bila A.8.16 diaktifkan. |
| Web/API (dinamis) | OWASP ZAP, `nuclei` (templat terkurasi), uji penetrasi independen | **Hanya** pada lingkungan sintetis dengan izin tertulis; jangan terhadap produksi atau training yang berisi data nyata. |
| Sertifikat TLS | Pemeriksa kedaluwarsa (Prometheus blackbox, atau cron `openssl`), pemeriksa TLS publik | Belum ada pemantauan kedaluwarsa (A.8.21 butir 10). |
| Endpoint | Laporan endpoint swa-lapor, MDM/EDR bertahap | Lihat A.8.1 Lampiran B. |
| Agregasi hasil | Register temuan (Lampiran C) di tracker/GitHub Issues dan lembar kerja | Satu tempat; hindari memecah per alat. |

Kriteria pilihan [usulan]: berjalan tanpa server tambahan, keluaran terstruktur (SARIF/JSON), dapat diarsipkan sebagai bukti, tidak mengirim kode atau data klien ke layanan pihak ketiga tanpa persetujuan (A.5.20), dan biaya terjangkau untuk tim kecil.

---

## Lampiran C — Templat Tiket Remediasi Kerentanan

Salin satu formulir per temuan. Jangan mencantumkan nilai rahasia, kunci, kata sandi, atau alamat IP; rujuk ke lokasi aman bila perlu. Format ID dan tenggat [usulan].

| Field | Isi |
| :--- | :--- |
| ID temuan | VUL-[tahun]-[nomor] |
| Tanggal diketahui / dilaporkan | [isi] |
| Sumber | [pemindai (nama), advisori pemasok, laporan eksternal, tinjauan internal, uji penetrasi, insiden] |
| Komponen dan versi terdampak | [mis. `<komponen> <versi>`; ID inventory INV-nomor] |
| Lokasi terdampak | [produksi / tenant `<slug>` / training / dev / Data Hub / host / CI / endpoint] |
| Pengenal publik | [ID CVE atau advisori **bila ada dan diverifikasi dari sumber**; kosongkan bila tidak ada] |
| Deskripsi singkat dan dampak | [isi, tanpa detail eksploit] |
| Keterpaparan | [terjangkau internet / jaringan Docker / lokal]; data klien terdampak: [ya/tidak]; lintas-tenant: [ya/tidak] |
| Eksploit publik / eksploitasi aktif | [ya/tidak/tidak diketahui; sumber] |
| Skor dasar dan kelas akhir | Skor: [isi]; kelas dasar: [Kritis/Tinggi/Sedang/Rendah]; kelas akhir: [isi]; alasan penyesuaian dan penyetuju: [isi] |
| Pemilik (peran) / Pelaksana | [DevOps Lead / Backend Lead / ...] / [isi] |
| Tenggat (dari kelas) | Mitigasi: [tanggal]; perbaikan: [tanggal] |
| Mitigasi sementara | [isi / tidak ada]; bukti: [isi] |
| Rencana perbaikan | [naikkan versi / ganti komponen / ubah konfigurasi / hapus fitur]; jalur sah: [A.8.19] |
| Perubahan terkait | ID perubahan CHG-[tahun]-[nomor]; commit/SHA: [isi] |
| Dampak pada data, tenant, downtime | [isi]; perlu reboot: [ya/tidak] |
| Pengujian dan verifikasi | Hasil CI: [isi]; pindai ulang: [tanggal, hasil]; verifikasi alur yang disentuh: [isi] |
| Status | [Baru / Triase / Dalam perbaikan / Menunggu jendela / Terverifikasi / Ditutup / Pengecualian] |
| Pengecualian | [tidak / ID EXC-nomor] |
| Penutupan | Tanggal: [isi]; penutup (bukan pelaksana): [peran]; lampiran bukti: [EV-A8.8-04-xx] |
| Pelajaran | [apakah proses atau kontrol perlu diperbaiki: ya/tidak; catatan] |

---

## Lampiran D — Templat Register Pengecualian dan Penerimaan Risiko

Satu baris per pengecualian. Pengecualian tidak berlaku tanpa kendali pengganti dan tanggal kedaluwarsa.

| Kolom | Isi / contoh |
| :--- | :--- |
| ID pengecualian | EXC-[tahun]-[nomor] |
| ID temuan terkait | VUL-[tahun]-[nomor] |
| Komponen dan lokasi | [isi] |
| Kelas keparahan | [Kritis / Tinggi / Sedang / Rendah] |
| Alasan tidak dapat ditambal tepat waktu | [patch belum tersedia / risiko kehilangan data saat deploy / ketergantungan versi / sumber daya / komponen melewati dukungan] |
| Kendali pengganti (kompensasi) | [mis. port diikat loopback dan firewall, rate limit Kong, isolasi jaringan, pemantauan, autentikasi tambahan] dan bukti keberadaannya |
| Risiko residual dan analisis dampak | [isi; kaitkan ke register risiko] |
| Pemilik risiko (peran) | [isi] |
| Penyetuju (peran) | [Lead Information Security; Direktur Utama untuk Kritis/Tinggi] |
| Tanggal disetujui / kedaluwarsa | [isi] / [isi; maksimum menurut A.7] |
| Rencana penyelesaian dan tenggat | [isi] |
| Tinjauan terakhir / berikutnya | [isi] / [isi] |
| Status | [Aktif / Kedaluwarsa / Ditutup (bukti)] |

Pengecualian awal yang layak dicatat (bila perbaikan belum selesai) **[usulan, setelah diverifikasi]**: kontainer layanan berjalan sebagai root; Postgres `sslmode` bawaan `disable` pada jaringan Docker; Kafka `PLAINTEXT`; Redis tanpa `requirepass`; image dengan tag `latest`; Python 3.10 dan Node 20 bila dukungan terbukti berakhir (**Perlu verifikasi**); nilai bawaan literal selama belum dihapus.

---

## Lampiran E — Sumber Informasi Kerentanan dan Alur Triase

### E.1 Sumber informasi (usulan)

Sumber berikut adalah **opsi**; keanggotaan, langganan, dan tautan perlu dikonfirmasi oleh Lead Information Security. TIP-001 (Draft) menyebut BSSN, OTX, dan MISP sebagai sumber intelijen ancaman; tidak ada bukti langganan.

| Kategori | Sumber (opsi) | Dipantau oleh | Frekuensi [usulan] |
| :--- | :--- | :--- | :--- |
| Keluaran pemindai | CI (SCA, image, rahasia), pemindaian terjadwal (Lampiran B) | DevOps Lead | Setiap run dan [mingguan] |
| Peringatan platform | GitHub Dependabot alerts, secret scanning (Perlu verifikasi aktif) | DevOps Lead | [Harian] otomatis |
| Basis data kerentanan umum | NVD, OSV, basis data kerentanan Go, GitHub Advisory Database | Lead Information Security | [Mingguan] atau via pemindai |
| Daftar kerentanan terdorong eksploitasi | CISA KEV (daftar kerentanan yang dieksploitasi aktif) | Lead Information Security | [Mingguan] |
| Advisori pemasok komponen | Kong, Docker, nginx, PostgreSQL, Redis, Confluent/Kafka, ClamAV, Node.js, Python, PyTorch/TensorFlow, Nuxt/Vue, Alpine, AlmaLinux, Grafana/Prometheus/Loki, TimescaleDB, Jupyter, OpenSSH, WireGuard | Pemilik komponen (Lampiran G) | Berlangganan surel/RSS; [mingguan] tinjau |
| Intelijen ancaman nasional | BSSN dan sumber lain yang disebut TIP-001 | Lead Information Security | [Bulanan] |
| Laporan eksternal | Saluran pelaporan keamanan, program BBP-001 bila aktif, laporan klien | Lead Information Security | Segera |
| Uji keamanan | Uji penetrasi, DAST, latihan RTP-001 | Lead Information Security | Per kegiatan |
| Internal | Tinjauan kode, tinjauan keamanan, insiden dan log, temuan auditor | Semua peran | Sewaktu-waktu |

### E.2 Alur triase dan perbaikan (usulan)

| Langkah | Aktivitas | Peran | Keluaran |
| :--- | :--- | :--- | :--- |
| 1. Penerimaan | Catat informasi masuk (sumber, tanggal) pada register; hilangkan duplikat | DevOps Lead / Lead Information Security | Entri awal |
| 2. Pencocokan aset | Cocokkan dengan inventory: apakah komponen dan versi terpasang, di lingkungan mana (produksi, tenant, training, Data Hub, host, CI) | DevOps Lead | Daftar aset terdampak |
| 3. Evaluasi keterpaparan | Terjangkau dari internet? Fungsi rentan dipakai? Data klien dan lintas-tenant? Eksploit publik? Kendali kompensasi? (Lampiran A bagian A.5) | Lead Information Security, pemilik komponen | Catatan evaluasi |
| 4. Klasifikasi | Tetapkan kelas dan tenggat (Lampiran A bagian A.3 dan A.4); catat penyesuaian | Lead Information Security | Kelas dan tenggat |
| 5. Tiket | Buat tiket (Lampiran C), tetapkan pemilik | Lead Information Security | Tiket |
| 6. Mitigasi sementara | Bila Kritis/Tinggi dan patch belum siap atau belum dapat dideploy: blokir jalur, nonaktifkan fitur, rotasi rahasia | DevOps Lead | Bukti mitigasi |
| 7. Perbaikan | Perubahan lewat jalur sah dengan CI dan persetujuan (A.8.19, A.8.25); patch per tenant bila relevan | Pemilik komponen, DevOps Lead | Commit/rilis, catatan deploy |
| 8. Verifikasi | Pindai ulang; verifikasi alur yang disentuh; konfirmasi bahwa versi di server sesuai | DevOps Lead, `auditsphere-qa` | Hasil pindai ulang |
| 9. Penutupan atau pengecualian | Tutup oleh peran bukan pelaksana; bila tidak dapat ditambal, ajukan pengecualian (Lampiran D) | Lead Information Security | Tiket ditutup / EXC |
| 10. Pelaporan dan pembelajaran | Perbarui metrik; eskalasi pelanggaran SLA; perbarui inventory/SBOM dan bila perlu kebijakan atau gerbang CI | Lead Information Security | Laporan [bulanan] |

Kerentanan yang diduga sudah dieksploitasi dialihkan ke prosedur insiden (A.5.24) pada langkah 1.

---

## Lampiran F — Daftar Periksa Tinjauan Berkala (usulan)

Frekuensi: tinjauan register [bulanan] dan tinjauan penuh [kuartalan] oleh Lead Information Security dan DevOps Lead. Setiap tinjauan dicatat (tanggal, peninjau, hasil).

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Pemindaian SCA, image, dan rahasia berjalan pada periode ini dan keluarannya diarsipkan. | [isi] | |
| 2 | Dependabot alerts dan secret scanning aktif dan alert terbuka ditinjau. | [isi] | |
| 3 | Inventory/SBOM terbaru mencakup semua komponen, termasuk tenant, training, Data Hub, host, dan GitHub Actions. | [isi] | |
| 4 | Tidak ada temuan Kritis terbuka melampaui [7 hari] dan Tinggi melampaui [14 hari] tanpa pengecualian aktif. | [isi] | |
| 5 | Semua pengecualian aktif memiliki kendali pengganti dan belum kedaluwarsa; yang kedaluwarsa telah ditutup atau diperpanjang oleh penyetuju. | [isi] | |
| 6 | Pembaruan keamanan host (VPS dan server Data Hub) tertunda tidak melampaui SLA; uptime kernel wajar terhadap jendela reboot; daftar paket diekspor. | [isi] | |
| 7 | ClamAV: tanggal signature terbaru dalam batas [usulan: 3 hari]; versi mesin ditinjau; sumber pembaruan dapat dijangkau. | [isi] | |
| 8 | Komponen dalam 12 bulan menuju akhir dukungan (Python, Node, Go, Postgres, Redis, Kafka, Kong, Alpine, torch) memiliki rencana; EOL diverifikasi dari sumber resmi. | [isi] | |
| 9 | Tidak ada image `latest` atau tanpa pin pada produksi dan tenant; versi tenant sesuai register. | [isi] | |
| 10 | Lingkungan training dan dev yang memuat data nyata mendapat patch setara produksi (atau dihentikan/direset dengan image terbaru). | [isi] | |
| 11 | Pola rentan di bagian 4 butir 10 berstatus ditutup atau memiliki pengecualian: nilai bawaan literal, `pickle.load`, route tanpa autentikasi, `curl \| bash`. | [isi] | |
| 12 | Sumber informasi kerentanan (Lampiran E) dipantau dan langganan masih aktif; saluran pelaporan eksternal berfungsi. | [isi] | |
| 13 | Rekaman perubahan (CHG) tersedia untuk setiap patch produksi (A.8.19 Lampiran D). | [isi] | |
| 14 | Patch endpoint personel ditinjau lewat laporan endpoint (A.8.1, A.6.7); perangkat dengan akses produksi tidak tertinggal. | [isi] | |
| 15 | Uji penetrasi/DAST terakhir sesuai jadwal dan temuannya masuk register. | [isi] | |
| 16 | Metrik Lampiran A bagian A.9 dilaporkan dan tren ditinjau. | [isi] | |
| 17 | Klaim eksternal tentang pemindaian dan SLA patch (dokumen keselarasan, materi klien) sesuai bukti yang ada. | [isi] | |

---

## Lampiran G — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Solution Architect | Backend/Frontend Lead | Data Analyst | Legal & Compliance | Direktur Utama |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan VUL-001 | A/R | C | C | C | C | C | A (persetujuan) |
| Inventory dan SBOM sebagai dasar pemindaian | C | A/R | C | R | R | I | I |
| Menjalankan pemindaian dan CI gate | A | R | C | R | R | I | I |
| Memantau sumber informasi dan advisori | A | R | C | R | R | I | I |
| Triase dan klasifikasi | A/R | R | C | C | C | I | I |
| Patch dependensi dan kode | C | C | C | A/R | A/R | I | I |
| Patch image, host, ClamAV, dan stack tenant | C | A/R | R | I | I | I | I |
| Mitigasi sementara dan perubahan darurat | A | R | C | C | C | I | I |
| Pengecualian dan penerimaan risiko | A/R | C | C | C | C | C | A (Kritis/Tinggi) |
| Uji penetrasi dan DAST | A/R | C | C | I | I | I | I |
| Saluran pelaporan eksternal dan komunikasi ke klien | A/R | C | I | I | I | R | I |
| Patch endpoint personel (lewat A.8.1) | A | R | I | I | I | I | I |
| Evidence ID dan review berkala | A/R | C | I | I | I | C | I |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.8.19 Instalasi perangkat lunak pada sistem operasional | Ada (`docs/ISO_27001/A.8.19_Instalasi_Perangkat_Lunak_pada_Sistem_Operasional.md`; belum ter-commit) | Jalur deploy dan persetujuan perubahan; inventory/SBOM (Lampiran C), pin image; angka SLA di A.8.19 Lampiran A perlu diseragamkan dengan Lampiran A dokumen ini. |
| A.8.25 Siklus hidup pengembangan yang aman | Ada | Gerbang CI (SCA, SAST, secret scan), threat model, pola rentan, uji penetrasi; register temuan (tindakan 14). |
| A.8.13 Pencadangan informasi | Ada | Backup pra-patch dan uji pemulihan sebelum patch besar; pemasangan `rclone`/`age` di host. |
| A.8.21 Keamanan layanan jaringan | Ada | Keterpaparan port, firewall, SSH, pemantauan tanpa Alertmanager, TLS dan kedaluwarsa sertifikat. |
| A.8.1 Perangkat endpoint pengguna | Ada | Patch OS dan aplikasi endpoint personel; tidak ada MDM/EDR. |
| A.6.7 Kerja jarak jauh | Ada | Laporan endpoint triwulanan (templat). |
| A.5.24 Persiapan manajemen insiden | Ada | Kerentanan yang dieksploitasi sebagai insiden; RTP-001. |
| A.5.36 Kepatuhan terhadap kebijakan dan standar | Ada | Klaim dokumen keselarasan dan penyimpangan; register temuan. |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada | Pemasok image/registri/GitHub; alat pihak ketiga yang menerima kode. |
| A.5.12 Klasifikasi informasi | Ada | Klasifikasi data yang menentukan prioritas dampak. |
| A.5.8 Keamanan informasi dalam manajemen proyek | Ada | Definition of Done keamanan dan CI. |
| A.5.7 Intelijen ancaman | **Draf, dokumen belum ada** | TIP-001 (docx) ada sebagai prosedur Draft; belum ada dokumen annex. |
| A.5.6 Hubungan dengan kelompok kepentingan khusus | **Draf, dokumen belum ada** | Keanggotaan BSSN/komunitas dan sumber advisori. |
| A.5.19, A.5.21, A.5.22, A.5.23 | **Draf, dokumen belum ada** | Rantai pasok TIK dan pemantauan pemasok (image, paket, GitHub Actions). |
| A.8.9 Manajemen konfigurasi | **Draf, dokumen belum ada** | Nilai bawaan literal, konfigurasi di luar repo, baseline host dan kontainer. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Log akses SSH, deploy, dan percobaan eksploitasi. |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Alert untuk signature ClamAV, kedaluwarsa sertifikat, pembaruan tertunda. |
| A.8.31 Pemisahan lingkungan pengembangan, pengujian, dan produksi | **Draf, dokumen belum ada** | Training dan dev berbagi host/image dan memuat data nyata. |
| A.8.32 Manajemen perubahan | **Draf, dokumen belum ada** | Catatan perubahan untuk patch (A.8.19 Lampiran D sebagai masukan). |
| A.8.34 Perlindungan sistem informasi selama pengujian audit | **Draf, dokumen belum ada** | Izin dan batas pemindaian dan uji penetrasi. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.8	Technological	Pengelolaan kerentanan teknis	Kerentanan diidentifikasi, diprioritaskan, dan diperbaiki tepat waktu	Scan report, patch SLA, remediation ticket, exception register	Berlaku	Layanan multi-tenant memproses data audit klien lembaga jasa keuangan dengan permukaan serangan luas (layanan Go, python-ai, frontend, Kafka, Postgres, Kong, ClamAV, Data Hub) pada satu VPS; komponen tidak ditambal adalah jalur kompromi; tanpa pemindaian, SLA, atau pemantauan advisori; klien akan menanyakan SLA patch (lihat A.8.8 bagian 1)	Sebagian (ad hoc: penguatan hasil tinjauan satu kali, lockfile, pin sebagian image, ClamAV fail-closed; belum ada proses berulang)	Sebagian (manifes, lockfile, Dockerfile, compose, rencana, dan dokumen Draft BBP-001/RTP-001/TIP-001 ada; scan report, tiket remediasi, register pengecualian, bukti patch host belum ada)	Belum terdaftar (usulan EV-A8.8-01 s/d EV-A8.8-12)	Lead Information Security (didukung DevOps Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah VUL-001 disetujui dan pemindaian baseline pertama selesai	Belum ada	Belum teruji	Tanpa pemindaian dependensi/image/host/rahasia dan Dependabot tidak terbukti; versi Go/Python/Node/Kong tidak seragam dan dukungan sebagian Perlu verifikasi; base image latest dan kontainer root; patch OS/kernel host tidak terbukti; ClamAV tanpa pemantauan signature; jalur patch terkendali hanya lima layanan Go; pola rentan di repo (nilai bawaan literal, pickle.load, route OpenAI tanpa autentikasi di HEAD, curl|bash); drift tenant dan training; tanpa uji penetrasi; tanpa register pengecualian; angka SLA BBP-001/RTP-001 berbeda dari annex lain	Verifikasi GitHub (Dependabot, secret scanning); setujui VUL-001; pemindaian baseline lokal; register temuan dan pengecualian; verifikasi host dan ClamAV read-only; tutup pola rentan; gerbang CI; pembaruan otomatis dependensi; seragamkan toolchain dan image; kolom EOL; jalur dan jendela patch; sumber advisori dan pelaporan; uji penetrasi; Evidence ID (15 tindakan di bagian 5)	Tinggi	[usulan] 31 Desember 2026 (GitHub, VUL-001, baseline pemindaian); 28 Februari 2027 (CI gate, pemindaian image, inventaris host, register pengecualian)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
