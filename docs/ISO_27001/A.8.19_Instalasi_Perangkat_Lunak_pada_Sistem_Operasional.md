# A.8.19 — Instalasi Perangkat Lunak pada Sistem Operasional

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Instalasi perangkat lunak pada sistem operasional |
| **Fokus pemeriksaan** | Instalasi software pada produksi diotorisasi dan dikendalikan |
| **Contoh bukti objektif** | Allowlist, deployment record, approval, software inventory |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi, stack tenant, training, stack dev, Data Hub, pipeline GitHub Actions |
| **PIC** | Lead Information Security (didukung DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut agar instalasi dan perubahan software pada sistem operasional (produksi) **diotorisasi**, **terkendali** (hanya komponen yang disetujui, dari sumber yang dikenal, dapat ditelusuri dan dapat dibatalkan), dan **tercatat**. Repo hanya membuktikan **skrip, workflow, Dockerfile, compose, dan dokumen**; repo **tidak** membuktikan apa yang benar-benar terpasang atau berjalan di VPS (paket host, versi Docker/nginx/rclone, image yang berjalan, siapa yang memegang akses root, pengaturan GitHub seperti required reviewers dan branch protection). Semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH, akses GitHub API, atau akses konsol yang dipakai dalam penyusunan dokumen ini. Butir yang bersumber dari catatan operasional internal (catatan 4 sampai 10 November 2026, memori kerja) dicantumkan sebagai **[Catatan]** dan **bukan bukti repo**. Tidak ada allowlist, software inventory, SBOM, pemindaian kerentanan dependensi, atau rekaman persetujuan deploy yang ditemukan di repo. Seluruh angka dan ambang pada Lampiran A sampai F (frekuensi tinjauan, batas waktu patch, retensi catatan) adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan (skrip repo memuat alamat VPS tertanam; alamat itu sengaja tidak dikutip di sini).

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.19 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.19** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Instalasi perangkat lunak pada sistem operasional** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Instalasi software pada produksi diotorisasi dan dikendalikan** | Sesuai baris spreadsheet; tiga gate: otorisasi (siapa menyetujui), kendali (allowlist, sumber, pin versi, jalur tunggal), rekaman (deployment record, inventory). |
| Contoh bukti objektif | **Allowlist, deployment record, approval, software inventory** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | Produksi memproses data audit klien lembaga jasa keuangan; perubahan software di VPS (image, binary, paket host, konfigurasi gateway) langsung memengaruhi kerahasiaan dan ketersediaan data itu. Pengecualian tidak dapat dipertahankan (bagian 1). |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: satu VPS dengan akses root, banyak jalur deploy yang tidak setara, dua repo pernah dapat men-deploy produksi, tanpa allowlist/inventory/SBOM, dan insiden hilangnya uploads akibat deploy (catatan). |
| Status implementasi | **Sebagian** | Ada jalur deploy backend terkendali (`deploy.yml` manual, `deploy-prod.sh`: build dari `git archive`, checksum binary, dump pra-deploy, rollback otomatis, tanpa seeder, satu layanan per waktu, catatan deploy di server) dan pin host key SSH. Belum ada: allowlist, inventory/SBOM, pemindaian dependensi, bukti persetujuan, jalur tunggal, pin image/versi yang konsisten, prosedur instalasi paket host. |
| Ketersediaan bukti | **Sebagian (skrip dan workflow) / Tidak ada (operasi)** | Workflow, skrip, Dockerfile, dan compose ada. Rekaman persetujuan, log deploy terarsip, daftar paket host, dan pengaturan GitHub **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.19-01 s/d EV-A8.19-12) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **DevOps Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal bila jalur deploy tunggal dan allowlist awal disetujui | Selanjutnya: tinjauan allowlist/inventory [kuartalan] dan kebijakan [tahunan] (Lampiran A, B, C, F). |
| Desain kontrol | **Perlu perbaikan** | Jalur `deploy-prod.sh` dirancang baik, tetapi jalur lain yang tidak terkendali masih ada di repo, tidak ada allowlist/inventory, base image mengambang, dan otorisasi bergantung pada pengaturan GitHub yang tidak terbukti. |
| Efektivitas operasional | **Belum teruji** | Tidak ada rekaman persetujuan, log deploy terarsip, atau tinjauan inventory. Tidak dapat diklaim bahwa hanya perubahan yang disetujui yang sampai ke produksi. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: beberapa jalur deploy produksi dan dua repo; otorisasi tidak terbukti; tanpa allowlist/inventory/SBOM; tanpa pemindaian dependensi dan tanpa test di CI; base image mengambang (`alpine:latest`, `:latest` monitoring/tenant); frontend dibangun di laptop; instalasi host via `curl \| bash`; konfigurasi produksi di luar repo. |
| Tindakan perbaikan | **Lihat bagian 5 (15 tindakan)** | Draf Lampiran A sampai F sebagai masukan. |
| Prioritas | **Tinggi** | Produksi satu host dengan akses root; perubahan software tidak memiliki persetujuan yang dapat dibuktikan; klien akan menanyakan manajemen perubahan dan kerentanan komponen. |
| Target | **[usulan] 31 Desember 2026** untuk jalur deploy tunggal, bukti persetujuan, dan allowlist awal; **[usulan] 31 Januari 2027** untuk inventory/SBOM dan pemindaian CI | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "diotorisasi" dan "tercatat" belum terbukti; gate "terkendali" hanya terpenuhi pada satu jalur. |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: kebijakan disetujui, jalur deploy tunggal ditetapkan, bukti persetujuan satu siklus rilis, allowlist awal. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 6 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.19 **berlaku** untuk AuditSphere karena:

1. **Produksi adalah satu host dengan akses root.** Skrip deploy bersambung ke VPS sebagai `root` (`backend/deploy-prod.sh`, `VPS_HOST`; `.github/workflows/deploy.yml` memakai secret kunci SSH). Siapa pun yang dapat memicu deploy atau memegang kunci itu dapat memasang kode apa pun yang berjalan di host yang sama dengan Postgres multi-tenant (A.8.13 bagian 1).
2. **Perubahan software produksi langsung memengaruhi data klien.** Image layanan Go, Kong, python-ai, ClamAV, dan frontend berjalan di host yang menyimpan data audit klien lembaga jasa keuangan (A.5.12: *Confidential/Restricted*). Komponen yang tidak disetujui atau dependensi yang rentan adalah jalur kompromi.
3. **Ada banyak jalur instalasi ke produksi dengan tingkat kendali berbeda.** Satu jalur terkendali (`deploy.yml` ke `deploy-prod.sh`), dan jalur lain masih ada di repo: skrip root `deploy-prod.sh`/`deploy.sh` ke `backend/deploy-backend.sh` (`docker compose up -d --build` tanpa override tanpa-seeder dan tanpa dump), `frontend/deploy-prod.sh` (build di mesin pengembang, `--force-recreate`), dan `backend/deploy-prod-serverbuild.sh` (salinan mirip tanpa pemanggil). `docs/deploy-tidy-plan.md` mencatat ini sebagai masalah (bagian 4 butir 1).
4. **Pernah terjadi kerugian akibat deploy yang tidak terkendali.** Catatan internal (bukan bukti repo) menyebut `deploy.yml` lama pada push ke `main` melakukan `--force-recreate` dan menghapus berkas upload pada 4 Oktober 2026, serta bahwa dua repo GitHub dapat men-deploy VPS yang sama. Ini contoh nyata mengapa otorisasi dan kendali jalur diperlukan (bagian 4 butir 2).
5. **Rantai pasok software tidak terdokumentasi.** Tidak ada allowlist, inventory, SBOM, atau pemindaian dependensi di CI; base image mengambang pada banyak Dockerfile dan compose (bagian 4 butir 4 sampai 8). Kontrol ini terkait langsung dengan A.5.21 (rantai pasok TIK, draf) dan A.8.8 (manajemen kerentanan teknis, draf).
6. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim menanyakan siapa yang menyetujui perubahan produksi, bagaimana versi dikontrol, dan apakah ada inventory komponen. Saat ini tidak ada dokumen tunggal atau rekaman yang dapat diserahkan; A.5.32 dan A.5.36 sudah mencatat ketiadaan SBOM dan gerbang CI.
7. **Banyak lingkungan berbagi host.** Stack dev, training, tenant, monitoring, dan produksi berjalan di VPS yang sama menurut catatan (bukan bukti repo); instalasi pada satu lingkungan dapat memengaruhi yang lain (sumber daya, jaringan Docker, port).

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti skrip, workflow, Dockerfile, atau konfigurasi di repo**. Apakah itu yang berjalan di VPS adalah **Perlu verifikasi** kecuali dinyatakan lain.

### 2.1 Jalur deploy produksi terkendali (backend)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Pemicu manual, ref eksplisit | `workflow_dispatch` saja; input `ref` dan `services` wajib; ref harus teresolusi ke commit sebelum jalan. | `.github/workflows/deploy.yml` (blok `on`, langkah "Resolve deploy ref") | `ref` bebas (cabang, SHA apa pun); tidak ada pemeriksaan bahwa commit sudah ditinjau, lulus test, atau berada di cabang tertentu. |
| Satu deploy produksi pada satu waktu | `concurrency: auditsphere-production`, `cancel-in-progress: false`. | `deploy.yml` | Hanya untuk workflow ini; jalur lain tidak tunduk padanya. |
| Persetujuan lingkungan | `environment: production`; komentar di workflow menyatakan lingkungan itu harus dikonfigurasi dengan *required reviewers* sebelum dipakai pertama kali. | `deploy.yml` (`environment`, komentar) | Pengaturan required reviewers ada di GitHub, bukan di repo: **Perlu verifikasi**. Tanpa itu, `--yes` membuat deploy berjalan tanpa persetujuan siapa pun selain pemicu. |
| Izin minimal workflow | `permissions: contents: read`. | `deploy.yml` | Baik; secret `VPS_SSH_KEY` dan `VPS_KNOWN_HOSTS` tetap berada pada lingkungan yang sama. |
| Host key SSH dipin | Langkah instalasi kunci menolak berjalan bila `VPS_KNOWN_HOSTS` kosong dan melarang `ssh-keyscan`. | `deploy.yml` (langkah "Install SSH key and pinned host key") | Nilai secret tidak terlihat dari repo. |
| Build dari commit, bukan working tree | `git archive <SHA>`; binary Go dikompilasi `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`, diperiksa sebagai ELF x86-64; `DEPLOY_SHA` dan `BINARY_SHA256` ikut payload dan diverifikasi di server. | `backend/deploy-prod.sh` (bagian "Building", `deploy_one`, validasi payload) | Pin toolchain Go longgar (bagian 4 butir 6). Build terjadi di runner/laptop, bukan di dalam Dockerfile (binary disalin; lihat 2.3). |
| Tampilan perubahan sebelum deploy | Skrip mencetak commit (hash, penulis, tanggal, judul) antara HEAD server dan SHA target, serta `diff --stat`, untuk jalur layanan yang dipilih. | `deploy-prod.sh` (bagian "Resolving ref") | Hanya tampilan; tidak ada gerbang yang mensyaratkan pemeriksa kedua. Pada CI, `--yes` melewati prompt "deploy" interaktif. |
| Pra-deploy dan keselamatan data | `--dry-run`; pre-flight read-only; `pg_dumpall` terverifikasi; snapshot jumlah baris; salin-keluar uploads audit; tag image rollback `auditsphere-rollback/<svc>-service:<TS>` dan tarball rollback. | `deploy-prod.sh` | Dump tidak terenkripsi dan tanpa retensi (A.8.13 bagian 4 butir 5). |
| Tanpa seeder | Override compose di luar repo menjalankan `migrate` saja; skrip menolak command yang masih memuat `seed`. | `deploy-prod.sh` (OVERRIDE, pemeriksaan command) | Override dihasilkan skrip dan berada di server, bukan di repo (bagian 4 butir 13). |
| Satu layanan per waktu, tanpa efek samping luas | `up -d --no-deps --build <svc>-service`; tidak pernah `down`, `prune`, atau `--remove-orphans`; tidak menyentuh postgres, redis, kafka, kong, python-ai, frontend, dev, atau stack tenant. | `deploy-prod.sh` (kepala skrip) | Sekaligus berarti python-ai, Kong, ClamAV, dan frontend **tidak** memiliki jalur terkendali yang setara (bagian 4 butir 1). |
| Health check dan rollback otomatis | `wait_healthy`; kegagalan memicu `rollback_service`; perbandingan jumlah baris sesudah deploy (tabel yang kehilangan baris membuat exit non-nol). | `deploy-prod.sh` | Teruji tidaknya rollback pada kondisi nyata: **Perlu verifikasi** (tidak ada log di repo). |
| Catatan deploy di server | Setiap deploy menulis direktori `/root/auditsphere-deploy/<TS>/` (log, `rollback-*`, `rows-*`), dan berkas `last-deploy` berisi timestamp, SHA, layanan, HEAD repo sebelumnya, lokasi backup, hasil. Workflow merujuk lokasi log server. | `deploy-prod.sh` (blok akhir); `deploy.yml` (langkah Summary) | Catatan hanya di VPS (satu host), tanpa retensi atau pengarsipan, dan tanpa identitas penyetuju. Bukti nyata: **Perlu verifikasi**. |
| Penyinkronan Kong terkendali | `sync-kong.sh`: hash konfigurasi, validasi Kong DB-less sementara, laporan diff rute, mode `apply` dengan cadangan lokal dan pemulihan bila reload gagal. | `scripts/sync-kong.sh` | Dijalankan manual di VPS; tidak ada rekaman siapa/kapan di repo. |
| Pemeriksaan rute di CI | Workflow `route-coverage.yml` gagal bila frontend memanggil path tanpa rute di `kong/prod/kong.yml`. | `.github/workflows/route-coverage.yml` | Satu-satunya pemeriksaan otomatis di CI; bukan pemeriksaan keamanan komponen. |

### 2.2 Jalur instalasi lain yang ditemukan di repo

| Jalur | Perilaku terverifikasi | Bukti | Kendali yang tidak ada |
| :--- | :--- | :--- | :--- |
| Root `deploy-prod.sh` / `deploy.sh` | Menjalankan `backend/deploy-backend.sh` lalu deploy frontend; alamat VPS tertanam. | `deploy-prod.sh` (root) | Tidak ada persetujuan, dump, atau pencegahan seeder. |
| `backend/deploy-backend.sh` | Mengemas dan mengunggah backend, lalu `docker compose -f docker-compose.prod.yml up -d --build`; compose produksi menjalankan `seed && serve` pada auth, audit, master. | `backend/deploy-backend.sh`; `backend/docker-compose.prod.yml` (`command:` pada layanan) | Seeder menimpa data produksi (memori: prod-deploy-no-seed); tanpa dump, tanpa rollback. |
| `frontend/deploy-prod.sh` | `npm run build` **di mesin pengembang**, unggah arsip `.output`, `docker rm -f` lalu `compose up -d --build --force-recreate` di server. | `frontend/deploy-prod.sh` | Artefak tidak terikat ke SHA commit atau dibuat dari `git archive`; tidak ada persetujuan atau rollback. Menurut catatan (bukan bukti repo), kontainer frontend produksi dibuat manual dan berbeda dari file compose; prosedur aktual memakai image bertag `<tanggal>-<sha>`. |
| `backend/deploy-prod-serverbuild.sh` | Varian `deploy-prod.sh` yang membangun binary di server dalam `golang:1.25-alpine`; tidak dipanggil dari mana pun (`docs/deploy-tidy-plan.md`). | `deploy-prod-serverbuild.sh` | Risiko drift antara dua salinan skrip. |
| Deploy dev otomatis | Push ke `dev`/`develop` menjalankan SSH ke VPS memakai `appleboy/ssh-action@v1.0.3` dan men-deploy dari repo GitHub. | `.github/workflows/deploy-dev.yml` | Tanpa lingkungan/persetujuan; action pihak ketiga dipin tag, bukan SHA; tidak ada input fingerprint host terlihat pada langkah itu. |
| Stack tenant | `generate-tenant-stack.sh` dan `onboard-tenant.sh` menghasilkan stack di `backend/tenants/<slug>/` dan menjalankan `docker compose`; image tenant memakai tag `:latest` (`rb_audit_*:latest`), Kong `3.4`. | `scripts/generate-tenant-stack.sh`; `scripts/onboard-tenant.sh`; `scripts/templates/docker-compose.tenant.yml.tpl` (baris image) | Tidak ada prosedur pembaruan versi lintas tenant atau pencatatan versi per tenant selain registri slug/port (`backend/tenants/registry.tsv`). |
| Monitoring | `deploy-monitoring.sh` menjalankan `compose pull` lalu `up -d`; Prometheus, Grafana, dan node-exporter memakai `:latest`. | `monitoring/deploy-monitoring.sh`; `monitoring/docker-compose.monitoring.yml` | Versi berubah tanpa disengaja setiap `pull`; Loki dan Promtail dipin (`3.7.8`, `3.6.8`). |
| Training | `reset-training.sh` menandai image kontainer produksi yang berjalan (`docker tag`) untuk dipakai training; produksi hanya dibaca. | `backend/training/reset-training.sh:47-49` | Baik (image sama dengan produksi); training tetap salinan data klien (A.8.13 butir 10). |
| Data Hub | Compose terpisah (`nginx:1.27-alpine`, `postgres:16-alpine`, `timescale/timescaledb:2.17.2-pg16`, build lokal untuk API/engine); prosedur deploy aman. | `data-server/docker-compose.yml`; `data-server/docs/DEPLOY_DATAHUB_SAFE.md` | Pin lebih baik daripada compose produksi; tidak ada inventory atau persetujuan. |

### 2.3 Pengendalian komponen (versi, sumber, lockfile)

| Area | Yang terlihat | Bukti | Penilaian |
| :--- | :--- | :--- | :--- |
| Go | `go.sum` per layanan (`auth`, `audit`, `master`, `risk`, `analytics`, `shared`); versi `go` di `go.mod` berbeda (1.25.0 sampai 1.25.8); CI memakai `setup-go` `1.25` dengan `check-latest: true`. | `backend/*/go.mod`, `go.sum`; `deploy.yml` | Dependensi terkunci per modul; toolchain mengambang. |
| Image layanan Go | Dockerfile `FROM alpine:latest`, `apk add ca-certificates tzdata` tanpa pin, `COPY <binary>`; binary dikompilasi di luar Dockerfile. | `backend/{auth,audit,master,risk,analytics}-service/Dockerfile` | Base image mengambang; isi image berubah tiap build tanpa perubahan repo. Build lokal dengan `--build` tanpa kompilasi ulang memuat binary lama (catatan: local-docker-prebuilt-binary). |
| python-ai | `python:3.10-slim`; lapisan awal memasang paket tanpa versi (`typing-extensions`, `jinja2`, `filelock`, `sympy`, `networkx`, `fsspec`, `mpmath`, `pillow`), `numpy==1.26.4`, torch/torchvision/torchaudio CPU `2.2.2`/`0.17.2`/`2.2.2` dari `--find-links` ke mirror PyTorch; `requirements.txt` 14 baris dipin `==`; tanpa hash. `entrypoint.sh` menyalin model bawaan image ke volume dan, bila model tidak ada, menjalankan `python train.py`. | `backend/python-ai/Dockerfile`; `requirements.txt`; `entrypoint.sh` | Dependensi langsung dipin, transitif dan lapisan awal tidak. Berkas model (mis. `xgboost_model.json`, `anomaly_bundle.pkl` disebut di entrypoint) adalah artefak yang dimuat saat runtime; asal dan integritasnya tidak didokumentasikan. |
| Frontend | `node:20-alpine`; `package.json` memakai rentang `^`; ada **dua** lockfile (`package-lock.json` dan `bun.lock`); Dockerfile hanya menyalin `.output`. | `frontend/Dockerfile`; `frontend/package.json`; `frontend/package-lock.json`; `frontend/bun.lock` | Ambiguitas lockfile (dicatat juga di A.5.32); build di luar CI. |
| Image infrastruktur produksi | `postgres:16`, `redis:7-alpine`, `confluentinc/cp-zookeeper:7.6.1`, `confluentinc/cp-kafka:7.6.1`, `kong:3.8`, ClamAV dipin **digest SHA256**. | `backend/docker-compose.prod.yml` (baris image) | Campuran: Kafka/Zookeeper/ClamAV dipin, Postgres dan Redis hanya tag mayor. Tidak ada daftar yang menyatakan pin ini disengaja. |
| Layanan Python Data Hub | `python:3.10-slim`; `jupyter/scipy-notebook:python-3.10` untuk caatt-engine; `requirements.txt` per layanan. | `data-server/*/Dockerfile`; `requirements.txt` | Tag mengambang; versi minor Python 3.10 mencapai akhir dukungan menurut jadwal rilis publik pada Oktober 2026 (**Perlu verifikasi** terhadap jadwal resmi). Node 20 juga mendekati/melewati akhir dukungan menurut jadwal publik (**Perlu verifikasi**). |

### 2.4 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `.github/`, `backend/`, `scripts/`, `monitoring/`, `data-server/`, `docs/` (19 November 2026) tidak menemukan: allowlist software atau daftar komponen yang disetujui; software inventory atau SBOM (CycloneDX/SPDX); Dependabot/Renovate (`.github/dependabot.yml` tidak ada); pemindaian dependensi atau image (`govulncheck`, `npm audit`, `pip-audit`, `trivy`, `grype`) di workflow; langkah `go test`/`go vet`/test frontend di CI (hanya `route-coverage.yml` selain dua workflow deploy); `CODEOWNERS`; dokumen kebijakan atau prosedur manajemen perubahan; templat catatan deploy/persetujuan; prosedur instalasi dan pembaruan paket host (OS, Docker, nginx, certbot, rclone, age) dan inventory paket host; penandatanganan atau verifikasi image/artefak (cosign); pin image berbasis digest selain ClamAV; kebijakan siapa boleh memiliki akses root atau akses tulis ke repo; daftar akses GitHub. Seluruh keadaan VPS dan pengaturan GitHub **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Allowlist software | **Belum tersedia; draf di Lampiran B** | Kode usulan SWI-001. Entri awal diambil dari repo; perlu persetujuan. |
| Kebijakan instalasi software dan manajemen perubahan | **Belum tersedia; draf di Lampiran A** | Kode usulan SWI-001; angka [usulan]. |
| Software inventory / SBOM | **Sebagian** | Manifes dan lockfile ada; tidak ada SBOM terhasilkan. Templat dan baris awal di Lampiran C. |
| Deployment record (per rilis) | **Sebagian** | Skrip menghasilkan `last-deploy`, `deploy.log`, tag rollback di server; **Perlu verifikasi** bahwa berkas ada dan diarsipkan. Riwayat run GitHub Actions (jika dapat diakses) adalah bukti kedua. Templat catatan di Lampiran D. |
| Approval (persetujuan deploy) | **Belum tersedia** | Required reviewers pada lingkungan `production`: **Perlu verifikasi** (tangkapan layar pengaturan oleh pihak berwenang). Tidak ada catatan persetujuan per rilis. |
| Daftar siapa dapat men-deploy | **Belum tersedia** | Akses tulis repo (RauMomo dan upstream), pemegang kunci SSH root, pemegang secret GitHub belum didaftar (Lampiran E). |
| Bukti hanya satu jalur deploy aktif | **Belum tersedia** | Perlu pemeriksaan workflow di `origin/main` dan di repo kedua (**Perlu verifikasi**). |
| Daftar paket host dan versi | **Belum tersedia** | Tidak ada inventory host; perlu ekspor read-only (Lampiran C). |
| Hasil pemindaian kerentanan | **Belum tersedia** | Tidak ada pemindaian di CI atau terjadwal. |
| Bukti uji rollback | **Belum tersedia** | Mekanisme ada; tidak ada log uji. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari catatan operasional internal (tanggal dicantumkan) dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan.

1. **[Repo] Beberapa jalur deploy produksi dengan tingkat kendali berbeda.** Hanya `deploy.yml` ke `backend/deploy-prod.sh` yang terkendali, dan hanya untuk lima layanan Go. Jalur root `deploy-prod.sh`/`deploy.sh`, `backend/deploy-backend.sh` (seeder, tanpa dump), `frontend/deploy-prod.sh` (build laptop, `--force-recreate`), dan `deploy-prod-serverbuild.sh` (salinan tak terpakai) masih ada di repo dan dapat dijalankan oleh siapa pun yang memegang akses ke VPS. Python-ai, Kong, ClamAV, frontend, tenant, dan monitoring tidak memiliki jalur terkendali yang setara. `docs/deploy-tidy-plan.md` (rencana 27 Oktober 2026, belum diimplementasikan) sudah mengidentifikasi ini; dokumen ini menambahkan sudut pandang otorisasi dan rekaman.
2. **[Catatan] Dua repo GitHub pernah dapat men-deploy VPS yang sama, dan `origin/main` mungkin masih memuat workflow lama.** Catatan 8 dan 21 Oktober 2026 menyebut `origin` dan `upstream` (fork) sama-sama memiliki `deploy.yml` yang berjalan pada push ke `main`, dan workflow `main` lama melakukan `up -d --build --force-recreate --remove-orphans` dengan `seed && serve`. Di cabang `staging` workflow sudah manual. Apakah `main`/fork masih dapat memicu deploy, dan siapa memegang secret SSH di tiap repo: **Perlu verifikasi** sebelum klaim apa pun. Selama ini belum dipastikan, **jangan klaim bahwa hanya jalur terkendali yang dapat mengubah produksi**.
3. **[Repo] Otorisasi deploy bergantung pada pengaturan GitHub yang tidak terbukti.** Workflow meneruskan `--yes` (melewati prompt "deploy"), `ref` bebas, dan hanya komentar yang menyebut *required reviewers*. Tidak ada `CODEOWNERS`, tidak ada pemeriksaan bahwa ref sudah ditinjau atau lulus test, dan tidak ada bukti branch protection. Riwayat git menunjukkan dua penulis pada berkas workflow dan skrip deploy (RauMomo dan Dimas Saifullah Ramadhan), menandakan lebih dari satu orang mengubah jalur instalasi tanpa catatan tinjauan di repo. Untuk A.8.19, "diotorisasi" harus berarti persetujuan oleh peran yang bukan pemicu, tercatat per rilis.
4. **[Tidak ada] Tidak ada allowlist, software inventory, atau SBOM.** Komponen yang boleh berjalan di produksi tidak terdefinisi; tidak ada daftar tunggal image, versi, sumber, pemilik, dan tanggal tinjauan. Inventory image tersebar di beberapa compose (A.5.32 bagian 4 butir 13). Tanpa allowlist, tidak ada dasar untuk menyatakan sebuah instalasi "tidak diizinkan".
5. **[Tidak ada] Tidak ada pemindaian kerentanan dependensi/image dan tidak ada test di CI.** Workflow hanya `deploy.yml`, `deploy-dev.yml`, `route-coverage.yml`. Tidak ada `govulncheck`, `npm audit`, `pip-audit`, `trivy`, Dependabot/Renovate. Pembaruan dependensi dan patch tidak memiliki pemicu atau batas waktu (A.8.8, draf). Gerbang CI sudah direkomendasikan di A.5.36 tindakan 5; dokumen ini menjadikannya prasyarat otorisasi instalasi.
6. **[Repo] Versi base image dan toolchain mengambang.** `alpine:latest` pada lima Dockerfile layanan Go, `apk add` tanpa pin, `python:3.10-slim`, `node:20-alpine`, `jupyter/scipy-notebook:python-3.10`; `prom/prometheus:latest`, `grafana/grafana:latest`, `prom/node-exporter:latest`; tag `:latest` pada image tenant; `postgres:16` dan `redis:7-alpine` hanya tag mayor; Go `check-latest: true`. Build yang sama dua kali pada waktu berbeda dapat menghasilkan isi berbeda, sehingga "yang disetujui" dan "yang terpasang" dapat berbeda tanpa perubahan di repo. Pengecualian baik: ClamAV dipin digest; Kafka/Zookeeper, Loki, Promtail, TimescaleDB dipin versi. Versi Kong juga berbeda antara produksi (`3.8`) dan templat tenant (`3.4`).
7. **[Repo] Binary Go dibangun di luar image dan frontend dibangun di mesin pengembang.** Dockerfile layanan hanya menyalin binary; `deploy-prod.sh` mengurangi risikonya (build dari `git archive`, `BINARY_SHA256`, ELF diperiksa), tetapi compose lokal dengan `--build` menyalin binary yang sudah ada di working tree, termasuk perubahan yang belum ter-commit (catatan: local-docker-prebuilt-binary). Frontend: `npm run build` dijalankan di laptop dan `.output` diunggah; tidak ada SHA atau hash artefak yang dicatat oleh `frontend/deploy-prod.sh`, dan tidak ada CI build. Dependensi frontend memakai rentang `^` dengan dua lockfile; kecocokan hasil build dengan lockfile yang dimaksud tidak terbukti.
8. **[Repo] Python-ai memuat dependensi dan artefak yang kurang terkendali.** Lapisan awal memasang paket tanpa versi dan torch dari mirror `--find-links`; `requirements.txt` tanpa hash; `entrypoint.sh` dapat menjalankan `train.py` saat start bila model tidak ada, dan berkas model disalin dari image ke volume. Provenans model dan checksum tidak dicatat. Python 3.10 dan Node 20 mendekati/melewati akhir dukungan publik (**Perlu verifikasi**).
9. **[Repo] Instalasi paket host memakai skrip dari internet dan tidak terdokumentasi sebagai prosedur.** `docs/backup-b2-setup.md:30` memasang `rclone` lewat `curl -fsSL https://rclone.org/install.sh | bash` (dengan catatan "tinjau skrip"); dokumen setup demo memasang Docker lewat `get.docker.com` dan nginx/certbot lewat `apt-get`, sementara runbook backup menyatakan VPS memakai AlmaLinux 9 tanpa `apt-get`; dua keadaan itu bertentangan, sehingga dokumen tidak dapat dijadikan inventory host. Tidak ada daftar paket host, versi Docker/nginx/certbot/rclone/age, atau kebijakan pembaruan otomatis (**Perlu verifikasi** di VPS).
10. **[Repo] Rekaman deploy hanya di server dan tanpa penyetuju.** `last-deploy`, `deploy.log`, dan direktori rollback berada di `/root/auditsphere-deploy/` pada VPS yang sama dengan produksi (kehilangan host berarti kehilangan rekaman); `docs/deploy-tidy-plan.md` mencatat belum ada aturan retensi untuk direktori rollback. Log GitHub Actions menyimpan pemicu dan ref, tetapi lamanya retensi log dan siapa penyetuju tidak terlihat dari repo.
11. **[Repo]+[Catatan] Deploy dev otomatis dan berbagi host dengan produksi.** `deploy-dev.yml` men-deploy otomatis pada push ke `dev`/`develop` memakai action pihak ketiga (tag `v1.0.3`, bukan SHA) dengan kunci dan passphrase dari secret; catatan 5 November 2026 menyebut stack dev dihentikan di VPS produksi dan hidup lagi bila ada push ke cabang itu (bukan bukti repo). Instalasi non-produksi di host produksi memperluas permukaan perubahan; keputusan memisahkan atau membatasi perlu diambil.
12. **[Repo] Stack tenant dan monitoring tanpa pembaruan terkendali.** Image tenant `:latest` dan monitoring `:latest`: versi yang berjalan bergantung pada kapan image terakhir di-build/pull, bukan pada keputusan. Tidak ada register versi per tenant selain slug/port (`registry.tsv`). Aturan isolasi tenant (tidak berbagi rahasia atau kunci antar-tenant; catatan: auditsphere-tenant-isolation-gotchas) berarti pembaruan versi harus dilakukan per tenant dengan rekaman sendiri.
13. **[Catatan] Konfigurasi produksi yang berada di luar repo.** Override compose tanpa-seeder dihasilkan `deploy-prod.sh` dan disimpan di `/root/auditsphere-deploy/`; catatan 4 November 2026 menyebut python-ai dan `PYTHON_AI_URL` analytics bergantung pada berkas override lain yang **tidak** dilewatkan oleh `deploy-prod.sh`, sehingga deploy analytics dapat mengembalikan target lama dan recreate python-ai tanpa override membuat semua panggilan AI gagal; ClamAV pernah dijalankan manual dengan `docker run` di luar compose (compose produksi di repo kini memuat layanan ClamAV dipin digest; mana yang berjalan: **Perlu verifikasi**). `deploy-tidy-plan.md` mencatat `git reset --hard` pada deploy menghapus perbaikan yang hanya ada di server. Keadaan produksi tidak dapat direkonstruksi dari repo saja.
14. **[Repo] Akses instalasi: root tunggal dan konfigurasi bawaan berisiko.** Seluruh jalur memakai SSH root; `deploy.yml` memakai satu kunci dari secret (tidak terlihat siapa pemegang salinannya). Compose produksi memuat nilai bawaan literal untuk kunci API python-ai (`PYTHON_AI_API_KEY`, nilai tidak dikutip) sehingga instalasi tanpa override memakai kunci yang ada di repo; ini juga perlu dibahas di A.8.9 (manajemen konfigurasi, draf). Siapa yang dapat memicu workflow (akses tulis repo) dan memegang akses root: **Perlu verifikasi**.
15. **[Tidak ada] Tidak ada prosedur persetujuan darurat dan pengecualian.** Tidak ada definisi perubahan darurat (hotfix langsung di server), siapa yang boleh, dan bagaimana dicatat setelahnya. Catatan operasional menunjukkan perbaikan langsung di server pernah dilakukan (mis. perbaikan konfigurasi ClamAV 5 November 2026 yang hilang bila server di-`git reset --hard`); itu wajar dalam praktik, tetapi harus memiliki catatan.
16. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet dalam penyusunan dokumen ini; tidak ada akses SSH atau GitHub API.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Tetapkan jalur deploy tunggal yang sah** per komponen (backend: `deploy.yml` ke `deploy-prod.sh`; tentukan jalur untuk frontend, python-ai, Kong, ClamAV, tenant, monitoring) dan nonaktifkan atau tandai usang jalur lain (`deploy-backend.sh`, root `deploy*.sh`, `deploy-prod-serverbuild.sh`) sesuai `docs/deploy-tidy-plan.md`. Perubahan skrip dilakukan oleh agen devops atas persetujuan, bukan oleh dokumen ini. | DevOps Lead, Lead Information Security | Tinggi | [usulan] 31 Desember 2026 | Keputusan jalur tertulis; riwayat commit |
| 2 | **Pastikan siapa yang dapat memicu deploy** (read-only, oleh pihak berwenang): required reviewers pada lingkungan `production`, branch protection `main`/`staging`, akses tulis repo `origin` dan repo kedua/fork, workflow di `origin/main` dan fork yang masih berpicu pada push, daftar pemegang secret `VPS_SSH_KEY`. Tidak menampilkan nilai secret. | Lead Information Security, DevOps Lead | Tinggi | [usulan] 30 November 2026 | Tangkapan layar/ekspor pengaturan bertanggal (EV-A8.19-05) |
| 3 | **Setujui kebijakan SWI-001 dan prosedur persetujuan perubahan** (Lampiran A dan D): persetujuan oleh peran selain pemicu, catatan per rilis, kriteria perubahan darurat. | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | Kebijakan disetujui; templat Lampiran D dipakai |
| 4 | **Wajibkan persetujuan pada lingkungan `production`** (required reviewers minimal satu peran bukan pemicu) dan batasi `ref` ke commit yang sudah ada di cabang terlindungi; hapus `--yes` dari jalur otomatis bila persetujuan eksplisit dipindahkan ke GitHub. | DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Pengaturan GitHub tersanitasi; satu run uji |
| 5 | **Arsipkan deployment record di luar VPS**: salin `last-deploy`, `deploy.log` ringkas, SHA, dan hash binary ke lokasi terlindungi per rilis (selaras A.8.15); tetapkan retensi direktori rollback [usulan: 12 bulan catatan, 5 rilis terakhir artefak]. | DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Contoh catatan terisi (EV-A8.19-03) |
| 6 | **Setujui allowlist awal** (Lampiran B) dengan keputusan manusia atas base image, sumber, dan versi; jadikan acuan peninjauan setiap perubahan komponen. | Lead Information Security, Solution Architect | Tinggi | [usulan] 31 Desember 2026 | Allowlist v0.1 disetujui (EV-A8.19-02) |
| 7 | **Pin versi dan digest** untuk image produksi dan tenant (ganti `alpine:latest`, `:latest` monitoring dan tenant; pilih tag mayor.minor atau digest), seragamkan versi Kong, tetapkan satu lockfile frontend (npm atau bun) dan pin versi Go/Node di CI. Perubahan kode/konfigurasi oleh agen devops atas persetujuan. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 31 Januari 2027 | Dockerfile/compose terbaru; daftar pin |
| 8 | **Hasilkan software inventory dan SBOM** per komponen (Go, npm, Python, image) dan simpan per rilis; gabungkan dengan inventory image dan paket host (Lampiran C). Pilih satu alat (selaras A.5.32 tindakan 8). | DevOps Lead | Sedang | [usulan] 31 Januari 2027 | Inventory v0.1; SBOM per rilis (EV-A8.19-04) |
| 9 | **Tambahkan gerbang CI** sebelum deploy: `go vet`/`go test`, `govulncheck`, `npm audit` (atau setara), `pip-audit`, pemindaian image (`trivy`/`grype`), secret scanning; tetapkan ambang blok dan batas waktu patch [usulan: kritis 7 hari, tinggi 30 hari] (selaras A.5.36 tindakan 5, A.8.8). | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 31 Januari 2027 | Workflow terbaru; log run lulus/gagal |
| 10 | **Inventarisasi host VPS secara read-only** dan tetapkan prosedur instalasi paket host: ekspor daftar paket dan versi (OS, Docker, nginx, certbot, rclone, age), ganti `curl \| bash` dengan sumber/paket terverifikasi, tetapkan siapa menyetujui paket baru dan kebijakan pembaruan keamanan OS. Hanya atas konfirmasi eksplisit; tidak mengubah paket pada langkah inventarisasi. | `auditsphere-devops`, DevOps Lead | Sedang | [usulan] 31 Januari 2027 | Ekspor tersanitasi (EV-A8.19-06); prosedur host |
| 11 | **Bawa konfigurasi produksi yang hidup di luar repo ke repo atau register**: override tanpa-seeder, override python-ai, ClamAV, nilai bawaan kunci; tambahkan override python-ai ke `deploy-prod.sh` atau dokumentasikan wajib dipakai (selaras `deploy-tidy-plan.md`); hapus nilai bawaan kunci literal dari compose. | `auditsphere-devops`, `auditsphere-backend` | Tinggi | [usulan] 31 Januari 2027 | Commit; daftar konfigurasi di luar repo |
| 12 | **Jadikan deploy frontend dapat ditelusuri**: build dari `git archive <SHA>` di lingkungan terkontrol (CI atau kontainer), catat SHA dan hash `.output`, tag image `<tanggal>-<sha>`, simpan rollback. | `auditsphere-devops`, `auditsphere-frontend` | Sedang | [usulan] 28 Februari 2027 | Skrip/CI frontend; rekaman rilis |
| 13 | **Prosedur pembaruan tenant dan monitoring**: register versi per tenant, urutan pembaruan, uji kesehatan dan rollback; monitoring memakai versi tetap dan `pull` hanya saat perubahan disetujui. | DevOps Lead, Solution Architect | Sedang | [usulan] 28 Februari 2027 | Register versi; prosedur |
| 14 | **Putuskan batas dev/training terhadap host produksi** (apakah deploy otomatis dev tetap di host produksi; pin action pihak ketiga ke SHA commit; batasi akses). | Lead Information Security, DevOps Lead | Sedang | [usulan] 28 Februari 2027 | Keputusan tertulis; workflow terbaru |
| 15 | **Daftarkan Evidence ID** pada Master Audit Evidence Index, jadwalkan tinjauan allowlist/inventory [kuartalan], dan commit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 2, 4, 7, 10, 11, 12, 13, dan 14 menyentuh GitHub, VPS, atau server lain; semuanya memerlukan konfirmasi eksplisit, dilakukan lewat agen devops/platform, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack, dan tidak mengubah skrip produksi tanpa persetujuan. Deploy ke produksi tidak boleh dipicu sebagai bagian dari pengumpulan bukti.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.19-01 | Workflow dan skrip deploy terkendali (`deploy.yml`, `deploy-prod.sh`, `sync-kong.sh`) | `.github/workflows/deploy.yml`; `backend/deploy-prod.sh`; `scripts/sync-kong.sh` | Internal |
| EV-A8.19-02 | Allowlist software awal dan kebijakan SWI-001 disetujui (setelah tindakan 3 dan 6) | `docs/ISO_27001/A.8.19_Instalasi_Perangkat_Lunak_pada_Sistem_Operasional.md` (Lampiran A, B); [belum ada versi disetujui] | Internal |
| EV-A8.19-03 | Contoh deployment record dan catatan persetujuan terisi (Lampiran D) beserta `last-deploy`/log (setelah tindakan 3 dan 5) | [belum ada] | Restricted |
| EV-A8.19-04 | Software inventory dan SBOM per rilis (setelah tindakan 8) | [belum ada] | Internal |
| EV-A8.19-05 | Pengaturan GitHub: required reviewers, branch protection, daftar akses, workflow di `main`/fork (setelah tindakan 2) | [belum ada] | Restricted |
| EV-A8.19-06 | Ekspor read-only inventaris paket host dan image berjalan di VPS (setelah tindakan 10) | [belum ada keluaran VPS] | Restricted |
| EV-A8.19-07 | Rencana perapian deploy produksi | `docs/deploy-tidy-plan.md` | Internal |
| EV-A8.19-08 | Dockerfile, compose, lockfile (bukti pin dan ketiadaan pin) | `backend/*/Dockerfile`; `backend/docker-compose.prod.yml`; `go.sum`; `frontend/package-lock.json`; `frontend/bun.lock` | Internal |
| EV-A8.19-09 | Workflow CI dengan pemindaian dan test (setelah tindakan 9) dan hasil run | [belum ada] | Internal |
| EV-A8.19-10 | Skrip deploy tidak terkendali beserta keputusan penonaktifannya (setelah tindakan 1) | `deploy-prod.sh`; `backend/deploy-backend.sh`; `frontend/deploy-prod.sh`; [belum ada keputusan] | Internal |
| EV-A8.19-11 | Prosedur pemasangan paket host dan runbook rclone terbarui (setelah tindakan 10) | `docs/backup-b2-setup.md`; [belum ada revisi] | Internal |
| EV-A8.19-12 | Catatan tinjauan allowlist/inventory berkala (Lampiran F) | [belum ada] | Internal |

Setelah tindakan 1 sampai 9 selesai, tambahkan ID per rilis produksi untuk catatan persetujuan dan deployment record.

---

## Lampiran A — Draf Kebijakan Instalasi Perangkat Lunak pada Sistem Operasional (SWI-001, usulan)

Draf ini adalah masukan untuk tindakan 3 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | SWI-001 Kebijakan Instalasi Perangkat Lunak pada Sistem Operasional (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden, perubahan arsitektur, atau perubahan jalur deploy) |
| Berlaku untuk | Produksi (auditsphere.app), stack tenant, Data Hub milik AuditSphere, monitoring, dan host VPS; training dan dev sejauh berbagi host produksi |

1. **Tujuan.** Memastikan hanya software yang disetujui, dari sumber yang dikenal dan pada versi yang ditetapkan, yang dipasang atau diubah pada sistem operasional, dan setiap perubahan dapat ditelusuri dan dibatalkan.
2. **Definisi sistem operasional.** Setiap host, kontainer, atau layanan yang memproses data klien atau melayani pengguna: VPS produksi, stack tenant, Data Hub produksi. Training dan dev dianggap operasional selama berbagi host atau memuat salinan data produksi.
3. **Jalur sah.** Perubahan software produksi hanya melalui jalur sah per komponen (tindakan 1). Perubahan di luar jalur sah adalah pengecualian darurat (butir 9).
4. **Otorisasi.** Setiap perubahan produksi disetujui oleh peran selain pemicu (pemohon tidak menyetujui perubahannya sendiri) dan dicatat per rilis (Lampiran D). Persetujuan memeriksa: daftar commit dan penulis sejak rilis terakhir, perubahan yang melanggar kompatibilitas, hasil CI, dampak pada data, dan rencana rollback.
5. **Allowlist.** Hanya komponen pada Lampiran B yang boleh dipasang. Komponen baru (image, paket, dependensi langsung, action CI) diajukan, ditinjau (lisensi, pemelihara, kerentanan, kebutuhan), dan ditambahkan sebelum dipasang.
6. **Sumber dan integritas.** Image dan paket berasal dari sumber resmi; versi dipin (tag mayor.minor atau digest) dan lockfile dikomit; tidak ada `latest` di produksi; instalasi dari skrip internet (`curl | bash`) hanya setelah skrip ditinjau dan disimpan, atau diganti paket terverifikasi.
7. **Build yang dapat ditelusuri.** Artefak dibangun dari commit tertentu di lingkungan terkontrol (bukan laptop pribadi), dengan SHA dan hash artefak dicatat; build dari working tree tidak dipakai untuk produksi.
8. **Pengujian dan verifikasi.** Perubahan lulus gerbang CI (test, pemindaian dependensi dan image) sebelum disetujui; pasca-deploy dilakukan health check dan verifikasi alur yang disentuh; batas waktu penambalan kerentanan [usulan: kritis 7 hari, tinggi 30 hari, sedang 90 hari].
9. **Perubahan darurat.** Perbaikan mendesak boleh dilakukan langsung dengan persetujuan lisan DevOps Lead dan Lead Information Security, dicatat paling lambat [1 hari kerja] setelahnya (Lampiran D, bagian darurat), dan diselaraskan ke repo.
10. **Rollback.** Setiap perubahan memiliki rencana dan artefak rollback; rollback diuji [kuartalan] (usulan).
11. **Inventory dan tinjauan.** Software inventory dan SBOM diperbarui per rilis; allowlist dan hak akses deploy ditinjau [kuartalan].
12. **Pembatasan akses.** Akses root dan kemampuan memicu deploy diberikan minimum dan ditinjau [6 bulan]; kunci SSH dan secret dikelola sesuai kebijakan rahasia (tidak ada di repo).
13. **Pemisahan lingkungan.** Dev dan training tidak berbagi host produksi bila memungkinkan; bila berbagi, perubahan di sana mengikuti butir 4 sampai 8.
14. **Peran.** Sesuai RACI Lampiran E.
15. **Pengecualian dan pelanggaran.** Pengecualian tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir. Instalasi tak berizin di produksi diperlakukan sebagai insiden keamanan (A.5.24).

---

## Lampiran B — Templat Allowlist Software dengan Entri Awal dari Repo

Kolom "Status pin (repo)" mencatat keadaan **saat ini**; kolom "Keputusan" diisi saat persetujuan. Entri adalah **usulan** untuk disetujui, bukan daftar yang sudah disahkan. Versi diambil dari berkas yang disebut; versi nyata di VPS: **Perlu verifikasi**.

| Kategori | Komponen | Sumber/berkas | Status pin (repo) | Pemilik (peran) | Keputusan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Image layanan Go | `alpine` + `ca-certificates`, `tzdata` | `backend/*-service/Dockerfile` | `latest`, `apk` tanpa pin | DevOps Lead | [setujui dengan pin / tolak `latest`] |
| Image | `python:3.10-slim` (python-ai, data-hub-api, ai-engine) | `backend/python-ai/Dockerfile`; `data-server/*/Dockerfile` | Tag minor; EOL menurut jadwal publik [verifikasi] | DevOps Lead | [naikkan versi / setujui sementara] |
| Image | `node:20-alpine` (frontend) | `frontend/Dockerfile` | Tag mayor; dukungan [verifikasi] | DevOps Lead | [isi] |
| Image | `postgres:16`, `redis:7-alpine` | `backend/docker-compose.prod.yml` | Tag mayor | DevOps Lead | [pin minor] |
| Image | `confluentinc/cp-zookeeper:7.6.1`, `confluentinc/cp-kafka:7.6.1` | `docker-compose.prod.yml` | Dipin versi | DevOps Lead | [setujui] |
| Image | `kong:3.8` (produksi), `kong:3.4` (templat tenant) | `docker-compose.prod.yml`; `docker-compose.tenant.yml.tpl` | Dipin tetapi tidak seragam | DevOps Lead | [seragamkan] |
| Image | `clamav/clamav@sha256:...` | `docker-compose.prod.yml` | Dipin digest | DevOps Lead | [setujui; tetapkan jadwal pembaruan] |
| Image | `prom/prometheus`, `grafana/grafana`, `prom/node-exporter` | `monitoring/docker-compose.monitoring.yml` | `latest` | DevOps Lead | [pin versi] |
| Image | `grafana/loki:3.7.8`, `grafana/promtail:3.6.8` | monitoring compose | Dipin versi | DevOps Lead | [setujui] |
| Image | `nginx:1.27-alpine`, `postgres:16-alpine`, `timescale/timescaledb:2.17.2-pg16`, `jupyter/scipy-notebook:python-3.10` | `data-server/docker-compose.yml`; `data-server/caatt-engine/Dockerfile` | Dipin (jupyter: tag mengambang) | Data Analyst, DevOps Lead | [isi] |
| Image tenant | `rb_audit_*:latest` (dibangun dari kode yang sama) | `scripts/templates/docker-compose.tenant.yml.tpl` | `latest` lokal | DevOps Lead | [tag per rilis] |
| Dependensi Go | modul pada `go.mod`/`go.sum` per layanan | `backend/*/go.mod` | Terkunci `go.sum`; `go` 1.25.0 s/d 1.25.8 | Backend Lead | [seragamkan toolchain] |
| Dependensi npm | `frontend/package.json` | `package-lock.json` dan `bun.lock` | Rentang `^`; dua lockfile | Frontend Lead | [pilih satu lockfile] |
| Dependensi Python | `requirements.txt` per layanan | `backend/python-ai`, `data-server/*` | `==` pada daftar langsung (python-ai: 14 baris); tanpa hash | Data Analyst, Backend Lead | [tambah hash/lockfile] |
| Artefak model | berkas model python-ai | `backend/python-ai` (image `pretrained_models`) | Provenans tidak tercatat | Data Analyst | [catat sumber dan checksum] |
| Alat host | Docker Engine, nginx, certbot, rclone, age | Dokumen setup; runbook backup | Tidak ada inventory | DevOps Lead | [isi setelah tindakan 10] |
| CI/CD | `actions/checkout@v4`, `actions/setup-go@v5`, `appleboy/ssh-action@v1.0.3` | `.github/workflows/*.yml` | Dipin tag, bukan SHA | DevOps Lead | [pin SHA; tinjau action pihak ketiga] |

Aturan pengajuan komponen baru [usulan]: nama, versi, sumber resmi, lisensi (A.5.32), alasan, pemilik, hasil pemindaian kerentanan, tanggal persetujuan, penyetuju, tanggal tinjauan berikutnya [12 bulan].

---

## Lampiran C — Templat Software Inventory

Salin satu baris per komponen terpasang. Isian "Contoh" diambil dari repo; kolom lokasi nyata menunggu ekspor read-only (tindakan 10).

| Kolom | Isi / contoh |
| :--- | :--- |
| ID inventory | INV-[nomor] |
| Komponen dan versi terpasang | [mis. `cp-kafka 7.6.1`] |
| Jenis | [image / paket host / dependensi Go/npm/Python / artefak model / action CI] |
| Lokasi terpasang | [VPS produksi / stack tenant `<slug>` / training / Data Hub / runner CI] |
| Sumber dan integritas | [registri, digest atau checksum] |
| Berkas definisi di repo | [mis. `backend/docker-compose.prod.yml`] |
| Entri allowlist (Lampiran B) | [ya/tidak, ID] |
| Pemilik (peran) | [DevOps Lead dsb.] |
| Tanggal pasang / terakhir diubah | [isi] |
| Hasil pemindaian terakhir | [tanggal; jumlah temuan kritis/tinggi] |
| Akhir dukungan (EOL) | [tanggal menurut sumber resmi] |
| Rencana pembaruan | [isi] |
| Catatan | [isi] |

Baris awal dari repo (versi nyata di server **Perlu verifikasi**):

| Komponen | Berkas | Catatan |
| :--- | :--- | :--- |
| 5 binary layanan Go (`auth`, `audit`, `master`, `risk`, `analytics`) | `backend/*-service` | Hash binary dicatat `deploy-prod.sh` (`BINARY_SHA256` di server) |
| Image python-ai (Python 3.10, torch CPU 2.2.2) | `backend/python-ai` | Model bawaan dalam image |
| Frontend Nuxt (`.output`, Node 20) | `frontend` | Build di luar CI |
| Kong, Postgres, Redis, Kafka, Zookeeper, ClamAV | `backend/docker-compose.prod.yml` | Lihat Lampiran B |
| Prometheus, Grafana, node-exporter, Loki, Promtail | `monitoring/docker-compose.monitoring.yml` | `latest` pada tiga komponen |
| Data Hub (nginx, Postgres CBS demo, TimescaleDB, API, CAATT, AI engine) | `data-server/docker-compose.yml` | Server terpisah |
| Paket host (OS, Docker, nginx, certbot, rclone, age) | tidak ada di repo | Ekspor read-only dari VPS |

---

## Lampiran D — Templat Catatan Deployment dan Persetujuan Perubahan

Salin satu formulir per rilis produksi. Jangan mencantumkan kunci, kata sandi, atau alamat IP.

| Field | Isi |
| :--- | :--- |
| ID perubahan (usulan) | CHG-[tahun]-[nomor] |
| Tanggal / jam rencana dan realisasi | [isi] |
| Pemohon (peran) / Pelaksana (peran) | [isi] / [DevOps Lead] |
| Penyetuju (peran, bukan pemohon) | [Lead Information Security / peran lain]; tanggal persetujuan: [isi] |
| Komponen dan jalur deploy | [auth / audit / master / risk / analytics / frontend / Kong / python-ai / tenant `<slug>` / host] melalui [jalur sah] |
| Ref dan SHA yang di-deploy | [isi]; HEAD sebelumnya: [isi] |
| Daftar commit dan penulis sejak rilis terakhir | [isi / tautan keluaran pra-deploy] |
| Perubahan yang melanggar kompatibilitas dan migrasi | [ya/tidak]; ringkasan: [isi] |
| Komponen baru atau berubah versi (terhadap allowlist) | [daftar; ID allowlist] |
| Hasil CI (test, pemindaian) | [lulus/gagal; tautan run] |
| Dampak pada data dan tenant | [isi] |
| Rencana rollback dan artefak | [tag image, tarball; lokasi] |
| Backup pra-deploy terverifikasi | [ya/tidak; nama berkas tanpa isi] |
| Hasil deploy | [berhasil / rollback / gagal]; log: [lokasi] |
| Verifikasi pasca-deploy (alur yang disentuh) | [daftar periksa; hasil] |
| Perubahan darurat | [ya/tidak]; bila ya: alasan, penyetuju lisan, waktu pencatatan, sinkronisasi ke repo: [isi] |
| Pembaruan inventory/SBOM | [ya/tidak; ID] |
| Lampiran bukti | [EV-A8.19-03-xx] |

---

## Lampiran E — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Solution Architect | Backend/Frontend Lead | Legal & Compliance | Direktur Utama |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan SWI-001 | A/R | C | C | C | C | A (persetujuan) |
| Menyetujui perubahan produksi | A | R | C | C | I | I |
| Memicu dan menjalankan deploy | I | A/R | I | C | I | I |
| Allowlist dan persetujuan komponen baru | A | R | R | C | C (lisensi) | I |
| Software inventory dan SBOM | C | A/R | C | R | C | I |
| Gerbang CI dan pemindaian kerentanan | A | R | C | R | I | I |
| Instalasi dan pembaruan paket host | C | A/R | I | I | I | I |
| Akses root, secret deploy, dan hak akses repo | A | R | I | I | I | I |
| Perubahan darurat dan catatan sesudahnya | A | R | I | C | I | I |
| Pembaruan stack tenant | A | R | R | I | I | I |
| Evidence ID dan review berkala | A/R | C | I | I | C | I |

---

## Lampiran F — Daftar Periksa Tinjauan Berkala (usulan)

Frekuensi tinjauan: [kuartalan] oleh Lead Information Security dan DevOps Lead. Setiap tinjauan dicatat (tanggal, peninjau, hasil).

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Setiap rilis produksi pada periode tinjauan memiliki catatan Lampiran D dengan penyetuju selain pemohon. | [isi] | |
| 2 | Tidak ada deploy produksi yang dipicu dari jalur di luar daftar jalur sah (periksa riwayat run, `last-deploy`, dan login SSH; A.8.15, draf). | [isi] | |
| 3 | Pengaturan GitHub tidak berubah tanpa persetujuan (required reviewers, branch protection, akses tulis, secret). | [isi] | |
| 4 | Allowlist, inventory, dan SBOM terbaru sama dengan yang berjalan di server (bandingkan daftar image berjalan dan versi). | [isi] | |
| 5 | Tidak ada image `latest` atau tanpa pin pada produksi dan tenant. | [isi] | |
| 6 | Temuan kerentanan kritis/tinggi telah ditangani dalam [7/30 hari]; pengecualian tercatat. | [isi] | |
| 7 | Komponen mendekati akhir dukungan (Python, Node, Postgres, Kafka) memiliki rencana pembaruan. | [isi] | |
| 8 | Konfigurasi produksi di luar repo (override, ClamAV, rahasia) terdaftar dan sinkron. | [isi] | |
| 9 | Perubahan darurat periode ini memiliki catatan dan sudah diselaraskan ke repo. | [isi] | |
| 10 | Rollback diuji pada kuartal ini. | [isi] | |
| 11 | Daftar pemegang akses root dan kunci SSH ditinjau [6 bulan]; tinjauan terakhir: [isi]. | [isi] | |
| 12 | Instalasi pada host (paket, skrip internet) hanya dari daftar yang disetujui. | [isi] | |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.8.13 Pencadangan informasi | Ada (`docs/ISO_27001/A.8.13_Pencadangan_Informasi.md`; belum ter-commit) | Dump pra-deploy dan jalur deploy; pemasangan `rclone`/`age` di host. |
| A.8.21 Keamanan layanan jaringan | Ada | Layanan pemasok dan Kong/nginx; versi Kong tidak seragam. |
| A.5.32 Hak kekayaan intelektual | Ada | Ketiadaan SBOM/lisensi dan dua lockfile frontend. |
| A.5.36 Kepatuhan terhadap kebijakan dan standar | Ada | Gerbang CI (tindakan 5), penyimpangan EX-03 dan EX-04. |
| A.5.24 Persiapan manajemen insiden | Ada | Instalasi tak berizin sebagai insiden; kehilangan uploads akibat deploy. |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada | Pemasok image/registri dan GitHub. |
| A.5.8 Keamanan informasi dalam manajemen proyek | Ada | Definition of Done keamanan dan CI. |
| A.5.19, A.5.21, A.5.22, A.5.23 | **Draf, dokumen belum ada** | Rantai pasok TIK dan pemantauan pemasok (image, paket, GitHub Actions). |
| A.8.8 Manajemen kerentanan teknis | **Draf, dokumen belum ada** | Pemindaian dan batas waktu patch. |
| A.8.9 Manajemen konfigurasi | **Draf, dokumen belum ada** | Konfigurasi di luar repo, nilai bawaan kunci, override. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Log deploy dan akses SSH. |
| A.8.31 Pemisahan lingkungan pengembangan, pengujian, dan produksi | **Draf, dokumen belum ada** | Dev/training berbagi host dengan produksi. |
| A.8.32 Manajemen perubahan | **Draf, dokumen belum ada** | Persetujuan dan rekaman perubahan; Lampiran D menjadi masukan. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.19	Technological	Instalasi perangkat lunak pada sistem operasional	Instalasi software pada produksi diotorisasi dan dikendalikan	Allowlist, deployment record, approval, software inventory	Berlaku	Produksi satu VPS dengan akses root memproses data audit klien lembaga jasa keuangan; perubahan software langsung memengaruhi data itu; banyak jalur deploy dan dua repo pernah dapat men-deploy; tanpa allowlist/inventory/SBOM (lihat A.8.19 bagian 1)	Sebagian	Sebagian (workflow, skrip, Dockerfile, compose ada di repo; rekaman persetujuan, log deploy terarsip, daftar paket host, dan pengaturan GitHub belum ada)	Belum terdaftar (usulan EV-A8.19-01 s/d EV-A8.19-12)	Lead Information Security (didukung DevOps Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah jalur deploy tunggal dan allowlist awal disetujui	Perlu perbaikan	Belum teruji	Beberapa jalur deploy produksi (hanya deploy-prod.sh terkendali) dan dua repo; otorisasi bergantung pada pengaturan GitHub yang tidak terbukti; tanpa allowlist/inventory/SBOM; tanpa pemindaian dependensi dan test di CI; base image mengambang (alpine:latest, latest monitoring/tenant); frontend dibangun di laptop; instalasi host via curl|bash; konfigurasi produksi di luar repo; rekaman deploy hanya di VPS	Pastikan siapa dapat memicu deploy (read-only); jalur deploy tunggal; setujui SWI-001 dan catatan persetujuan; wajibkan required reviewers; arsipkan deployment record; allowlist awal; pin versi/digest; inventory dan SBOM; gerbang CI; inventaris host (15 tindakan di bagian 5)	Tinggi	[usulan] 31 Desember 2026 (jalur tunggal, persetujuan, allowlist awal); 31 Januari 2027 (inventory/SBOM dan pemindaian CI)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
