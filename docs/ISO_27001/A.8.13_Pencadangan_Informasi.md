# A.8.13 — Pencadangan Informasi

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pencadangan informasi |
| **Fokus pemeriksaan** | Backup sesuai kebutuhan, terlindungi, dan diuji pemulihannya |
| **Contoh bukti objektif** | Backup policy, job log, restore test, encryption, retention |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi, stack tenant, training, Data Hub, Backblaze B2 & Cloud SaaS |
| **PIC** | Lead Information Security (didukung DevOps Lead dan Legal & Compliance) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut backup **sesuai kebutuhan**, **terlindungi**, dan **teruji pemulihannya**. Repo hanya membuktikan **skrip, runbook, dan konfigurasi**; repo **tidak** membuktikan apa yang berjalan di VPS (cron, versi skrip terpasang, isi direktori backup, isi bucket B2, aturan lifecycle, kunci `age`), sehingga semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH dan tidak ada akses konsol Backblaze yang dipakai dalam penyusunan dokumen ini. Butir yang bersumber dari catatan operasional internal (audit read-only VPS 5 Nov 2026, catatan 10 Nov 2026) dicantumkan sumbernya dan **bukan bukti repo**; catatan itu juga saling bertentangan dengan dokumen annex lain (bagian 4 butir 1). Tidak ada bukti uji restore, log pekerjaan backup, atau definisi RPO/RTO yang ditemukan di repo. Seluruh angka pada Lampiran A sampai E (RPO, RTO, retensi, frekuensi, ambang) adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan (nama bucket dan nama kontainer yang tercantum di skrip/runbook repo boleh disebut karena bukan rahasia; kunci aplikasi B2 dan kunci privat `age` tidak pernah ada di repo).

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.13 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.13** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Pencadangan informasi** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Backup sesuai kebutuhan, terlindungi, dan diuji pemulihannya** | Sesuai baris spreadsheet; tiga gate: kebutuhan (RPO/RTO/cakupan), perlindungan (enkripsi, akses, lokasi), pengujian (restore test). |
| Contoh bukti objektif | **Backup policy, job log, restore test, encryption, retention** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | AuditSphere menyimpan data audit klien lembaga jasa keuangan (KKA, LHA, ATR, berkas bukti) di satu VPS tanpa replika; backup adalah satu-satunya jalur pemulihan bila VPS, volume, atau basis data rusak atau diserang. Pengecualian tidak dapat dipertahankan (bagian 1). |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: data klien bernilai tinggi, satu titik kegagalan, multi-tenant dalam satu Postgres, perubahan produksi lewat deploy yang pernah menghapus data (uploads), dan klien akan menanyakan RPO/RTO serta bukti restore. |
| Status implementasi | **Sebagian** | Ada skrip backup terenkripsi `age` dengan fail-closed, checksum, retensi lokal, salinan `rclone` ke B2, skrip restore-test dan dekripsi, serta dump pra-deploy terverifikasi di `deploy-prod.sh`. Belum ada: kebijakan backup, RPO/RTO, bukti skrip terenkripsi terpasang di VPS, bukti job log, bukti restore test, cakupan Redis/Kafka/uploads master/uploads tenant/file konfigurasi/monitoring, dan backup Data Hub terjadwal. |
| Ketersediaan bukti | **Sebagian (skrip dan runbook) / Tidak ada (operasi)** | Skrip, runbook DRAFT, dan prosedur dekripsi ada di repo. Job log, hasil restore test, daftar isi bucket, aturan lifecycle, dan bukti pemasangan cron **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.13-01 s/d EV-A8.13-12) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **DevOps Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal bila backup terenkripsi terpasang dan restore test pertama selesai | Selanjutnya: restore test [bulanan] dan tinjauan kebijakan [tahunan] (Lampiran A, C). |
| Desain kontrol | **Perlu perbaikan** | Mekanisme teknis dirancang cukup baik (fail-closed, kunci privat di luar VPS, restore ke kontainer sementara), tetapi kebutuhan (RPO/RTO) tidak ditetapkan, cakupan tidak lengkap, runbook tidak sinkron dengan skrip, dan tidak ada pemantauan/alert. |
| Efektivitas operasional | **Belum teruji** | Tidak ada bukti restore test, job log, atau peninjauan. Catatan operasional menyebut cron berjalan tetapi dengan skrip lama tanpa enkripsi (bagian 4 butir 1); efektivitas tidak dapat diklaim. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: tanpa kebijakan/RPO/RTO; tanpa bukti restore; enkripsi terpasang tidak terbukti; B2 di luar Indonesia; cakupan tidak lengkap; tanpa alert; kunci `age` satu-satunya. |
| Tindakan perbaikan | **Lihat bagian 5 (14 tindakan)** | Draf Lampiran A sampai E sebagai masukan. |
| Prioritas | **Tinggi** | Klien lembaga jasa keuangan akan meminta bukti backup terenkripsi, retensi, dan restore test; tidak ada replika sehingga backup adalah satu-satunya pemulihan; backup yang tidak terenkripsi memuat rahasia MFA dan data klien. |
| Target | **[usulan] 31 Desember 2026** untuk restore test pertama dan kebijakan disetujui; **[usulan] 31 Januari 2027** untuk cakupan lengkap dan alert | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Field bukti operasional belum dapat diisi: Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "teruji pemulihannya" belum terpenuhi; gate "terlindungi" belum terbukti pada server. |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: kebijakan disetujui, enkripsi terbukti terpasang, restore test pertama terdokumentasi. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 6 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.13 **berlaku** untuk AuditSphere karena:

1. **Satu titik kegagalan.** Produksi berjalan di satu VPS dengan satu instans Postgres, Redis, dan Kafka tanpa replika (`backend/docker-compose.prod.yml:65` `KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1`; A.8.21 bagian 4 butir 13). Tidak ada salinan hidup; bila VPS, disk, atau volume hilang, backup adalah satu-satunya jalur pemulihan.
2. **Data yang dilindungi bernilai tinggi.** Basis data memuat data audit klien lembaga jasa keuangan, kredensial terenkripsi sumber data klien di Data Hub, dan material autentikasi (A.5.12: data audit klien *Confidential/Restricted*). Backup menyalin seluruh data ini, sehingga **perlindungan backup sama pentingnya dengan perlindungan data sumber** (enkripsi, akses, lokasi).
3. **Multi-tenant dalam satu Postgres.** Stack tenant memakai database per layanan pada Postgres bersama (`scripts/templates/docker-compose.tenant.yml.tpl:5-8`; A.8.21 bagian 2.1); `pg_dumpall` dari `backup-prod.sh` otomatis memuat data semua tenant dalam satu berkas. Satu backup yang bocor membuka semua tenant, dan restore penuh memengaruhi semua tenant sekaligus (bagian 4 butir 9).
4. **Perubahan produksi pernah menghapus data.** Catatan operasional dan memori internal mencatat bahwa `docker compose up --force-recreate` pada alur deploy lama menghapus berkas upload dan seeder menimpa data. Alur deploy backend saat ini (`.github/workflows/deploy.yml`, manual, menjalankan `backend/deploy-prod.sh`) menambahkan `pg_dumpall` terverifikasi dan salin-keluar/pulihkan uploads sebelum recreate. Itu bukti bahwa tim memahami risikonya, tetapi dump pra-deploy bukan strategi backup (tidak terenkripsi, tanpa retensi, tanpa salinan off-site; bagian 4 butir 5).
5. **Backup memuat rahasia yang belum terenkripsi di basis data.** Memori internal mencatat secret MFA tersimpan tanpa enkripsi di produksi; skrip migrasi enkripsi MFA ada di repo (`docs/mfa-encryption-rollout.md`) dan rencananya mensyaratkan backup (langkah 6) sebelum migrasi. Karena itu backup harus terenkripsi dan restore harus teruji **sebelum** migrasi MFA dijalankan.
6. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim menanyakan frekuensi backup, enkripsi, lokasi penyimpanan, retensi, dan bukti uji restore terakhir. Saat ini tidak ada dokumen tunggal atau bukti uji yang dapat diserahkan.
7. **Jalur backup melibatkan pemasok.** Salinan off-site dikirim ke Backblaze B2 (tidak ada region Indonesia), sehingga A.5.19 sampai A.5.23 dan klaim residensi data berkaitan langsung dengan kontrol ini (A.5.20 bagian 2.1; bagian 4 butir 4).

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti skrip, runbook, atau konfigurasi di repo**. Apakah itu yang berjalan di VPS adalah **Perlu verifikasi** kecuali dinyatakan lain.

### 2.1 Cakupan backup per penyimpanan data

| Penyimpanan data | Backup yang terlihat di repo | Bukti | Status |
| :--- | :--- | :--- | :--- |
| Postgres produksi (semua database, termasuk database per tenant dan peran) | `pg_dumpall -U postgres --clean --if-exists` dari kontainer `rb_audit_postgres`, dialirkan `gzip -9` lalu `age`, ditulis ke `.partial` lalu di-rename; checksum SHA-256 berkas terenkripsi. | `backend/scripts/backup/backup-prod.sh` (bagian 1) | Skrip terverifikasi. Terpasang dan terenkripsi di VPS: **Perlu verifikasi** (bagian 4 butir 1). |
| Uploads audit-service (`/root/uploads`) | `tar` dari dalam kontainer `rb_audit_audit_service`, dienkripsi `age`. Volume `audit_uploads` dideklarasikan pada compose produksi. | `backup-prod.sh` (bagian 2); `backend/docker-compose.prod.yml:189`, `:318` | Skrip terverifikasi. Isi arsip pada backup nyata: **Perlu verifikasi**. |
| Uploads master-service (`/root/uploads`) | **Tidak ada di `backup-prod.sh`.** Compose produksi tidak mendeklarasikan volume uploads untuk master-service (blok `volumes` master hanya memuat `pg-ca`, `docker-compose.prod.yml:209-210`). `reset-training.sh` membaca `/root/uploads` master bila ada (`backend/training/reset-training.sh:88-91`), jadi direktorinya mungkin ada dan berisi berkas. | `backup-prod.sh`; `docker-compose.prod.yml`; `reset-training.sh` | **Gap**: tidak tercakup dan tidak persisten (bagian 4 butir 6). Isi nyata: **Perlu verifikasi**. |
| Uploads stack tenant | Bind mount ke direktori host per tenant (`__STORAGE_DIR__:/root/uploads`). `backup-prod.sh` hanya membaca kontainer audit-service stack kontrol, bukan direktori tenant. | `scripts/templates/docker-compose.tenant.yml.tpl:129-133`; `backup-prod.sh` | **Gap**: tidak tercakup (bagian 4 butir 6). |
| Dump pra-deploy | `pg_dumpall | gzip` ke `/root/db-backups`, verifikasi `gzip -t` dan baris penutup "cluster dump complete", `chmod 600`; salin-keluar uploads audit ke direktori rollback dan dipulihkan setelah recreate. | `backend/deploy-prod.sh:110`, `:254-274`, `:437-453` | Terverifikasi (skrip). **Tidak terenkripsi, tanpa retensi/pembersihan, tanpa salinan off-site** (bagian 4 butir 5). |
| Redis (sesi `token:<jwt>`, cache, indeks logis per tenant) | Hanya volume `redis_data`; tidak ada backup. Tidak ada konfigurasi persistensi eksplisit di compose (perilaku bawaan image). | `docker-compose.prod.yml:22-23`, `:312-319` | Tidak dicadangkan. Sebagai cache/sesi, kehilangan berakibat login ulang; apakah ada data yang harus dipulihkan: **Perlu keputusan** (Lampiran B). |
| Kafka dan ZooKeeper | Hanya volume; retensi topik 168 jam; tanpa backup. | `docker-compose.prod.yml:39-40`, `:66`, `:69-70` | Tidak dicadangkan. Pesan event/log aktivitas yang belum dikonsumsi hilang saat kehilangan volume; dampak: **Perlu keputusan**. |
| Volume ClamAV dan model python-ai | `clamav_db` dan `python_ai_models` hanya volume. Model AI dan data pelatihan tidak ada di backup. | `docker-compose.prod.yml:317`, `:319` | Tidak dicadangkan. `clamav_db` dapat diunduh ulang; model AI **Perlu keputusan**. |
| Konfigurasi dan rahasia deploy (`.env`, file kunci, `age-recipients.txt`, konfigurasi nginx, `rclone.conf`) | Tidak ada prosedur backup di repo. Memori internal mencatat `JWT_SECRET` dan file server berada di luar repo. | Tidak ditemukan | **Gap**: kehilangan VPS berarti kehilangan rahasia yang diperlukan untuk menjalankan ulang stack (mis. `MFA_ENCRYPTION_KEY` bila migrasi MFA dijalankan). Penyimpanan rahasia di luar VPS **Perlu verifikasi**. |
| Konfigurasi Kong tersinkron | `scripts/sync-kong.sh` menyalin konfigurasi aktif sebelumnya ke `/root/kong-backups` sebelum reload (cadangan lokal untuk rollback, bukan backup DR). Sumber utama ada di git. | `scripts/sync-kong.sh:31`, `:97-99` | Terverifikasi (rollback lokal). |
| Monitoring (Prometheus, Grafana, Loki) | Volume `prometheus_data`, `grafana_data`, `loki_data`; retensi 7 hari; tanpa backup. Konfigurasi di repo; salinan server di `/app/rbia-monitoring` bukan checkout git (A.8.21 bagian 4 butir 4). | `monitoring/docker-compose.monitoring.yml:137-139` | Tidak dicadangkan. Dasbor/konfigurasi yang diubah langsung di server dapat hilang: **Perlu verifikasi**. |
| Data Hub (server terpisah: `datalake_data`, `cbs_data`, `hub_data`, `jupyter_data`) | Runbook deploy aman menyediakan `pg_dump -Fc` terenkripsi `openssl enc -aes-256-cbc -pbkdf2` dengan passphrase di file terpisah, dan alternatif snapshot volume; dipakai **sebelum deploy**, tanpa jadwal berkala di repo. Salinan off-box yang disarankan: lewat VPN WireGuard ke server aplikasi, "di Indonesia", dan **bukan** bucket luar negeri. | `data-server/docs/DEPLOY_DATAHUB_SAFE.md:46-108`; `data-server/docker-compose.yml:219-225` | Prosedur ada; **tidak ada backup terjadwal, tidak ada uji restore** (dokumen sendiri menyebut restore hypertable TimescaleDB "hanya jalan darurat", `:98`). Bertentangan dengan kebijakan B2 untuk produksi (bagian 4 butir 4). |
| Stack training (`rb_audit_training`) | Klon produksi, dibuat ulang dari produksi oleh `reset-training.sh` (`pg_dump -Fc` per database ke direktori training, `docker cp` uploads). Dump dihapus secara bawaan (`KEEP_DUMP_SETS=0`). Training **bukan backup**: data dibuang dan diisi ulang saat reset, sekaligus menghapus enrolmen MFA. | `backend/training/reset-training.sh:78-93`, `:100-161` | Bukan sumber pemulihan. Dump sementara tidak terenkripsi (`chmod 600`, direktori 700) selama proses reset; dicatat sebagai gap kecil (bagian 4 butir 10). |

### 2.2 Kendali pada skrip backup dan restore (terverifikasi dari kode)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Enkripsi sebelum tulis | Aliran dipipa `age -R <berkas penerima>`; teks polos tidak menyentuh disk. Berkas penerima hanya berisi kunci publik (`age1...`); skrip menolak berkas yang memuat `AGE-SECRET-KEY`. Kunci privat disimpan di luar VPS (menurut komentar skrip: pengelola kata sandi dan salinan offline). | `backup-prod.sh` (kepala dan pra-flight) | Kunci privat tidak terlihat di repo (benar). Siapa pemegangnya dan berapa salinannya: **Perlu verifikasi** (bagian 4 butir 8). |
| Fail-closed | Bila `age` atau berkas penerima tidak ada, backup gagal; tidak menulis backup polos kecuali `ALLOW_PLAINTEXT=1` (darurat). | `backup-prod.sh` (pra-flight) | Penggantian `ALLOW_PLAINTEXT=1` pada cron tidak terlihat dari repo. Salinan skrip terpasang berbeda dari repo menurut catatan 10 Nov 2026. |
| Integritas | Penulisan ke `.partial` lalu rename hanya bila sukses; header `age-encryption.org/v1` diperiksa; `pipefail`; checksum SHA-256 per berkas; `rclone copy --checksum`. | `backup-prod.sh` | Checksum dihitung dari berkas terenkripsi; tidak menjamin dump logis valid (itu tugas restore test). |
| Retensi lokal | `find ... -mtime +RETENTION_DAYS -delete`, bawaan 14 hari, hanya untuk `pg-*` dan `uploads-*`. | `backup-prod.sh` (bagian 3) | Retensi off-site diatur lifecycle bucket (bukan skrip). Dump `/root/db-backups` dari `deploy-prod.sh` tidak ikut dibersihkan. |
| Salinan off-site | `rclone copy` ke `RCLONE_REMOTE` bila diatur; kegagalan salin membuat skrip keluar non-nol; bila tidak diatur, hanya peringatan "local only". | `backup-prod.sh` (bagian 4) | Tidak ada `rclone check`/verifikasi sisi remote selain `--checksum` saat salin. Isi bucket: **Perlu verifikasi**. |
| Izin berkas | `umask 077`, direktori backup `chmod 700`. | `backup-prod.sh` | Izin nyata di VPS: **Perlu verifikasi**. |
| Uji restore | `restore-test.sh`: verifikasi checksum, restore ke kontainer Postgres sementara (bukan produksi), cek daftar database dan jumlah tabel/baris, ekstrak arsip uploads ke direktori sementara, keluar non-nol bila gagal; menerima berkas `.age` lewat `AGE_IDENTITY`. | `backend/scripts/backup/restore-test.sh` | Skrip ada; **tidak ada bukti dijalankan**. Hanya memeriksa keberadaan dan jumlah baris (perkiraan `n_live_tup`), bukan kebenaran isi atau waktu pemulihan. Harus dijalankan di mesin admin yang memegang kunci privat, bukan di VPS. |
| Dekripsi dan prosedur restore | `decrypt-backup.sh` memverifikasi checksum sebelum dekripsi dan dapat men-stream ke `stdout`; komentar skrip memuat prosedur restore produksi (dump segar dulu, hentikan layanan aplikasi, restore dengan `ON_ERROR_STOP`). | `backend/scripts/backup/decrypt-backup.sh` | Prosedur restore produksi penuh bersifat destruktif (`pg_dumpall --clean`); belum pernah terbukti dijalankan. |
| Enkripsi backup lama | `encrypt-existing-backups.sh`: dry-run bawaan, `--encrypt`, lalu `--delete-plaintext` hanya bila `.age` valid; catatan bahwa penghapusan pada disk virtual tidak menjamin blok lama tak dapat dipulihkan. | `backend/scripts/backup/encrypt-existing-backups.sh` | Belum terbukti dijalankan di VPS. |

### 2.3 Dokumen yang menyinggung backup

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| `docs/backup-b2-setup.md` (runbook B2) | Cron 02:30 harian, bucket privat `auditsphere-prod-backups`, kunci aplikasi terbatas satu bucket, lifecycle 30 hari ditambah 1 hari (sekitar 31 hari titik pulih), restore test bulanan, rotasi kunci aplikasi. | **Status DRAFT, "Not installed on the VPS"** (`:5`). Tidak sinkron dengan skrip terenkripsi (bagian 4 butir 2). |
| A.5.20 (annex ada) | Backblaze B2 terdaftar sebagai pemasok backup; "Backup off-site belum berjalan"; tindakan 9 menjalankan backup B2 dan mencatat DPA/kontrak. | Bertentangan dengan catatan operasional (bagian 4 butir 1). |
| A.5.24 (annex ada) | Backup produksi sebagai kendali pemulihan setelah insiden/ransomware; tindakan 12 menjadwalkan restore test dan menyimpan log. | Selaras dengan dokumen ini. |
| A.7.3 (annex ada) | Backup off-host berjalan menurut catatan, tidak terenkripsi, di luar Indonesia. | Selaras dengan catatan; belum dikonfirmasi lewat SSH. |
| A.8.21 (annex ada) | Backup ke B2 sebagai layanan jaringan; menyebut runbook DRAFT; tindakan 12 (pasang backup terenkripsi dan perbarui runbook); mengutip default retensi 14 hari lokal dan 30 hari B2. | Dokumen ini adalah tindak lanjutnya; angka yang dikutip dari repo, bukan komitmen. |
| `docs/mfa-encryption-rollout.md` | Langkah 6: backup basis data sebelum migrasi enkripsi MFA; dry-run bawaan. | Backup terenkripsi dan teruji adalah prasyarat praktis (bagian 1 butir 5). |
| `data-server/docs/DEPLOY_DATAHUB_SAFE.md` | Prosedur backup Data Hub terenkripsi dengan salinan off-box di Indonesia. | Prosedur pra-deploy; bukan program backup (bagian 2.1). |
| `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Klaim residensi data dan cadangan di Indonesia (AWS Jakarta) dan klaim kontrol Annex A. | Desain on-premises/penawaran; tidak boleh dipakai sebagai bukti A.8.13 (A.7.3 bagian 4 butir 4). |

### 2.4 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `backend/`, `scripts/`, `data-server/`, `monitoring/`, `docs/` (19 November 2026) tidak menemukan: kebijakan backup atau dokumen RPO/RTO; berkas cron, unit atau timer systemd untuk `backup-prod.sh`; log hasil backup atau restore test; rencana pemulihan bencana (DR) atau runbook pemulihan penuh dari nol; backup terjadwal Data Hub; backup Redis, Kafka, uploads master dan tenant, atau berkas rahasia/konfigurasi server; metrik atau alert keberhasilan backup (tidak ada scrape, `rule_files`, atau Alertmanager; lihat A.8.21 bagian 4 butir 4); pembersihan retensi untuk dump pra-deploy; kontrak/DPA Backblaze; prosedur serah-terima kunci `age` (pemilik, salinan, rotasi); pengujian ketahanan terhadap penghapusan (mis. object lock/versioning bucket). Seluruh keadaan VPS dan bucket **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Backup policy | **Belum tersedia; draf di Lampiran A** | Kode usulan BCK-001. Perlu persetujuan; angka [usulan]. |
| Definisi RPO/RTO dan retensi | **Belum tersedia; usulan di Lampiran B** | Tidak ada nilai tertulis. Nilai retensi 14 hari (lokal) dan 30 hari (B2) di skrip/runbook adalah bawaan teknis, bukan keputusan. |
| Job log (hasil backup harian) | **Belum tersedia / Perlu verifikasi** | Runbook mengarahkan keluaran cron ke `/var/log/auditsphere-backup.log`; berkas itu ada atau tidak di VPS tidak diketahui. Contoh tinjauan di Lampiran D. |
| Restore test | **Belum tersedia** | Skrip ada; tidak ada log atau formulir terisi. Formulir usulan di Lampiran C. |
| Encryption | **Sebagian** | Skrip terenkripsi ada di repo. Bukti terpasang di VPS dan salinan lama terenkripsi: **belum ada**. Bukti pemilik kunci privat dan salinan offline: **belum ada**. |
| Retention | **Sebagian** | Bawaan skrip dan runbook ada. Aturan lifecycle bucket dan isi bucket: **Perlu verifikasi** (tangkapan layar konsol atau `rclone lsl` oleh pihak berwenang tanpa menampilkan kunci). |
| Bukti cron aktif | **Belum tersedia / Perlu verifikasi** | `crontab -l` root (read-only); jangan menyertakan variabel bernilai rahasia. |
| Kontrol akses backup (siapa dapat membaca) | **Belum tersedia** | Izin direktori, pemegang kunci aplikasi B2, dan pemegang kunci `age` belum didaftar (Lampiran A butir 6, Lampiran E). |
| Kontrak/DPA dan ketentuan Backblaze | **Belum tersedia** | A.5.20 bagian 3 dan tindakan 9. |
| Rencana pemulihan penuh (dari VPS baru) | **Belum tersedia** | Draf kerangka di Lampiran A butir 9; belum diuji. |
| Bukti backup Data Hub | **Belum tersedia** | Hanya prosedur pra-deploy. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari catatan operasional internal (tanggal dicantumkan) dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan.

1. **[Catatan] Status backup di VPS tidak jelas dan bertentangan antar dokumen.** Catatan audit read-only 5 Nov 2026 menyebut cron 02:30 menjalankan `backup-prod.sh` dan mengirim ke bucket privat B2 (A.7.3, A.8.21 butir 8). Catatan 10 Nov 2026 menyebut skrip terpasang adalah salinan lama yang menulis berkas **tanpa enkripsi** dan `age` belum terpasang. Sementara itu runbook (`docs/backup-b2-setup.md:5`) dan A.5.20 menyebut "belum terpasang". Tiga sumber, tiga kesimpulan; keadaan nyata hanya dapat dipastikan lewat pemeriksaan read-only oleh pihak berwenang (cron, versi skrip, isi direktori dan bucket). Sampai itu dilakukan, **jangan klaim backup berjalan, terenkripsi, atau teruji**.
2. **[Repo] Runbook tidak sinkron dengan skrip terenkripsi.** Runbook (a) tidak memuat pemasangan `age`, berkas penerima, atau kunci privat; (b) langkah 4 dan 6 mengharapkan `pg-*.sql.gz` dan `uploads-*.tar.gz` (bukan `.age`), sehingga `ls ... pg-*.sql.gz` tidak menemukan berkas terenkripsi; (c) memerintahkan `restore-test.sh` dijalankan di VPS, padahal skrip menyatakan harus di mesin admin yang memegang kunci privat dan VPS hanya memiliki kunci publik; (d) pemasangan `rclone` memakai instalasi dari internet (`curl ... | bash`) dengan catatan "tinjau skrip"; (e) kunci aplikasi B2 disimpan di `rclone.conf` milik root pada VPS (izin 600), sehingga siapa pun yang menguasai root VPS menguasai kunci tulis/baca bucket. Runbook perlu diperbarui sebelum dipasang (tindakan 3).
3. **[Tidak ada] Tidak ada bukti restore test sama sekali.** Skrip `restore-test.sh` ada, tetapi tidak ada log, formulir, atau hasil di repo. Pengujian "pemulihan" adalah inti kontrol ini; tanpa bukti, nilai efektivitas operasional adalah "Belum teruji". Tidak ada pengukuran waktu pemulihan (RTO) maupun usia data terbaru yang dapat dipulihkan (RPO).
4. **[Repo]+[Catatan] Lokasi penyimpanan off-site di luar Indonesia dan kebijakan yang bertentangan.** Backblaze tidak memiliki region Indonesia. Dokumen Data Hub secara eksplisit melarang salinan data klien perbankan ke bucket luar negeri (`DEPLOY_DATAHUB_SAFE.md:96`), sedangkan backup produksi ke B2 menyalin data klien lain ke luar negeri. Karena `pg_dumpall` memuat data seluruh tenant, keputusan lokasi harus diambil per kelas data dan kontrak klien (kewajiban lokasi data lembaga jasa keuangan: **Perlu verifikasi** dengan Legal & Compliance). Klaim residensi cadangan di Indonesia pada dokumen keselarasan tidak boleh dipakai. Enkripsi `age` menurunkan risiko kebocoran, tetapi tidak menghilangkan pertanyaan lokasi; hindari frasa "data tidak meninggalkan Indonesia" dan "seluruh backup terenkripsi" sampai terbukti.
5. **[Repo] Dump pra-deploy tidak terenkripsi, tanpa retensi, hanya lokal.** `deploy-prod.sh` menulis `pg_dumpall | gzip` (hanya `chmod 600`) ke `/root/db-backups` dan tidak membersihkannya (tidak ada `find -delete`); direktori itu juga memuat salinan uploads rollback. Dump ini memuat data semua tenant dan rahasia MFA. Selain tidak terenkripsi, ukuran direktori bertambah setiap deploy (risiko disk penuh pada satu VPS). Apakah isi `/root/db-backups` saat ini bertumpuk dan siapa yang dapat membacanya: **Perlu verifikasi**.
6. **[Repo] Cakupan backup tidak lengkap.** Tidak tercakup: uploads master-service (tidak ada volume dan tidak dicadangkan), uploads stack tenant (bind mount host), Redis, Kafka/ZooKeeper, model python-ai, monitoring, konfigurasi nginx dan berkas rahasia/server. Catatan: nilai Redis dan Kafka mungkin dapat dibangun ulang; itu perlu diputuskan tertulis (Lampiran B). Uploads audit-service tercakup bila volume `audit_uploads` benar dipasang di VPS dan direktori yang diarsipkan sama.
7. **[Repo]+[Catatan] Uploads tidak persisten pada recreate.** Memori internal mencatat `deploy.yml` lama (`docker compose up -d --build --force-recreate`) menghapus uploads dan bahwa dua repo dapat men-deploy produksi. Alur saat ini di repo ini (manual, `deploy-prod.sh`) menyalin keluar dan memulihkan uploads audit dan volume `audit_uploads` kini dideklarasikan; **master-service tidak** diberi perlindungan serupa. Apakah workflow lama di `origin/main` atau repo lain masih aktif dan dapat memicu deploy otomatis: **Perlu verifikasi** (periksa sebelum klaim). Untuk A.8.13: tanpa backup berkas yang terbukti, recreate yang salah dapat menghilangkan berkas bukti klien tanpa jalur pemulihan.
8. **[Repo] Kunci `age` adalah titik kegagalan tunggal dan belum ada tata kelolanya.** Komentar skrip menyatakan kehilangan semua salinan kunci privat berarti kehilangan semua backup; menyebut penerima kedua opsional. Tidak ada daftar pemegang, salinan offline, uji dekripsi berkala, atau prosedur rotasi di repo. Penerima kedua (admin kedua) disarankan. Pengujian dekripsi dengan kunci nyata harus bagian dari restore test.
9. **[Repo] Restore selektif per tenant tidak didukung.** Backup berupa satu `pg_dumpall` semua database, dan prosedur restore produksi di `decrypt-backup.sh` menghapus dan membuat ulang semua database (`--clean`). Memulihkan satu tenant atau satu tabel memerlukan restore ke instans sementara lalu ekstraksi manual; belum ada prosedur, dan tenant isolation (memori internal) berarti restore ke database yang salah dapat menimpa data tenant lain. Prosedur ini harus ditulis dan diuji.
10. **[Repo] Salinan data produksi pada jalur non-backup.** `reset-training.sh` membuat dump produksi tidak terenkripsi sementara di direktori training (izin 600/700, dihapus bawaan). Training mendapatkan data klien nyata (kecuali MFA); kebijakan penggunaan data produksi untuk training adalah hal A.8.33 dan A.5.34 (draf, dokumen belum ada) dan di luar dokumen ini, tetapi dicatat karena dump itu berfungsi sebagai salinan lain dari backup.
11. **[Tidak ada] Tidak ada job log, pemantauan, atau alert backup.** Runbook hanya mengarahkan keluaran cron ke berkas log dan menyatakan "check the log, not just cron"; tidak ada peninjauan terjadwal, metrik, probe "backup terakhir lebih tua dari X", atau notifikasi gagal. Prometheus di repo tidak memuat `rule_files` atau Alertmanager (A.8.21 butir 4). Backup yang gagal dapat tidak diketahui selama berminggu-minggu.
12. **[Tidak ada] Tidak ada RPO/RTO dan tidak ada rencana pemulihan penuh.** Jadwal harian (runbook) berarti RPO secara teknis sampai sekitar satu hari ditambah jeda tulis, tetapi itu bukan keputusan atau komitmen. Tidak ada urutan pemulihan dari VPS baru (instal Docker, restore Postgres, uploads, kunci, Kong, nginx, DNS, sertifikat). Terkait A.5.30 (kesiapan TIK untuk kelangsungan bisnis; dokumen belum ada).
13. **[Tidak ada] Perlindungan terhadap penghapusan/ransomware tidak terdokumentasi.** Kunci aplikasi B2 memiliki capability `writeFiles` dan `listFiles` (runbook); apakah kunci dapat menghapus berkas, apakah versioning/object lock bucket aktif, dan apakah ada salinan tak dapat diubah: **Perlu verifikasi**. Bila root VPS dikuasai, kunci di `rclone.conf` memungkinkan penyerang menimpa atau menyembunyikan backup selama lifecycle.
14. **[Tidak ada] Data Hub tanpa backup terjadwal dan tanpa uji restore.** Prosedur ad hoc sebelum deploy; passphrase `openssl` dikelola manual di server yang sama (`/root/.datahub_backup_pass`, nilai tidak dikutip); `openssl enc` tidak memberi autentikasi integritas selain checksum terpisah. Data Hub memuat data lakehouse klien dan kredensial sumber data klien. Apakah Data Hub produksi klien ada, dan siapa pemilik backupnya (AuditSphere atau klien pada deployment on-premises, A.5.32 Lampiran), **Perlu verifikasi**.
15. **[Repo] Backup mengandung data yang tunduk pada penghapusan.** Data subjek yang dihapus dari aplikasi tetap ada di backup sampai retensi habis (14 hari lokal, sekitar 31 hari B2 menurut runbook). Kebijakan retensi dan penghapusan (A.8.10, draf, dokumen belum ada) perlu menyatakan perilaku ini dan cara menangani permintaan penghapusan terhadap backup.
16. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet dalam penyusunan dokumen ini; tidak ada akses SSH.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Pastikan keadaan nyata backup di VPS** (read-only, dilakukan pihak berwenang atau lewat `auditsphere-devops` atas konfirmasi eksplisit): isi cron root, versi skrip terpasang vs repo, ada tidaknya `age` dan berkas penerima, daftar berkas di direktori backup (nama, ukuran, tanggal, ekstensi `.age`), isi bucket B2 (daftar nama saja), aturan lifecycle, izin direktori. Jangan menampilkan atau menyimpan `rclone.conf` atau kunci. | DevOps Lead | Tinggi | [usulan] 30 November 2026 | Ekspor read-only bertanggal (disaring dari rahasia) |
| 2 | **Setujui kebijakan BCK-001** (Lampiran A) dengan keputusan manusia atas RPO, RTO, retensi, cakupan, dan lokasi (Lampiran B). | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | Kebijakan dan tabel RPO/RTO disetujui |
| 3 | **Perbarui runbook** `docs/backup-b2-setup.md` dari DRAFT agar sesuai skrip terenkripsi: pemasangan `age`, penerima, ekstensi `.age`, restore test di mesin admin, peninjauan log, rotasi kunci, penyimpanan kunci B2 dan kunci privat. Ganti `curl | bash` dengan sumber terverifikasi. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 15 Desember 2026 | Runbook revisi (Evidence EV-A8.13-02) |
| 4 | **Pasang atau perbarui backup terenkripsi di VPS** dan jalankan enkripsi/penghapusan polos terkontrol untuk backup lama (`encrypt-existing-backups.sh`: dry-run, `--encrypt`, uji dekripsi di mesin admin, baru `--delete-plaintext`), termasuk `/root/db-backups`. Perubahan produksi hanya atas konfirmasi eksplisit, satu perubahan per sesi, tanpa seeder. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Daftar berkas `.age` dan keluaran dry-run; log eksekusi |
| 5 | **Lakukan restore test pertama** dari salinan off-site memakai kunci privat nyata di mesin admin (`restore-test.sh`), isi Formulir Lampiran C, ukur waktu pemulihan nyata, dan catat hasil sebagai dasar RTO. Lakukan sebelum migrasi enkripsi MFA dijalankan. | DevOps Lead (pelaksana), Lead Information Security (peninjau) | Tinggi | [usulan] 31 Desember 2026 | Formulir terisi, log `restore-test.sh` (EV-A8.13-05) |
| 6 | **Tata kelola kunci `age`**: daftarkan pemegang (peran), tambahkan penerima kedua, simpan salinan offline sesuai komentar skrip, uji dekripsi berkala, tulis prosedur rotasi dan pencabutan. | Lead Information Security | Tinggi | [usulan] 31 Desember 2026 | Daftar pemegang kunci (tanpa nilai kunci), catatan uji dekripsi |
| 7 | **Perluas cakupan**: putuskan dan implementasikan backup uploads master-service dan uploads tenant, berkas konfigurasi/rahasia server (di penyimpanan terenkripsi terpisah), konfigurasi nginx tersanitasi, dan keputusan tertulis untuk Redis, Kafka, python-ai, monitoring. Tambahkan volume persisten untuk uploads master bila dibutuhkan (perubahan kode/konfigurasi dilakukan oleh agen backend/devops atas persetujuan, bukan oleh dokumen ini). | `auditsphere-devops`, Solution Architect | Tinggi | [usulan] 31 Januari 2027 | Keputusan cakupan; perubahan terdokumentasi |
| 8 | **Job log dan alert**: arahkan log ke lokasi terproteksi, tinjau sesuai Lampiran D, tambahkan alert "backup gagal atau lebih tua dari [usulan: 26] jam" dan "restore test terlambat" (selaras tindakan pemantauan A.8.21 butir 7). | DevOps Lead | Tinggi | [usulan] 31 Januari 2027 | Bukti alert uji, catatan tinjauan |
| 9 | **Retensi dan pembersihan**: tetapkan retensi per kelas data (Lampiran B), tambahkan pembersihan dump pra-deploy, verifikasi lifecycle bucket, dan definisikan penanganan permintaan penghapusan terhadap backup. | DevOps Lead, Legal & Compliance | Sedang | [usulan] 31 Januari 2027 | Konfigurasi dan keputusan retensi |
| 10 | **Putuskan lokasi dan pemasok off-site**: nilai B2 terhadap kewajiban lokasi data klien dengan Legal & Compliance; bila tidak dapat diterima untuk data klien, pilih penyimpanan lain (lihat `DEPLOY_DATAHUB_SAFE.md:96`). Catat DPA/kontrak dan sub-prosesor (A.5.20 tindakan 9). | Legal & Compliance, Lead Information Security | Tinggi | [usulan] 31 Januari 2027 | Keputusan tertulis; kontrak/DPA tersimpan |
| 11 | **Kendali akses backup**: daftar pemegang kunci aplikasi B2 (hanya tulis bila memungkinkan), batasi baca pada mesin admin, tinjau izin direktori, aktifkan versioning/object lock bila tersedia dan diputuskan; tinjau akses setiap [6 bulan]. | DevOps Lead, Lead Information Security | Sedang | [usulan] 31 Januari 2027 | Daftar akses, konfigurasi bucket (tersanitasi) |
| 12 | **Rencana pemulihan penuh dan restore selektif per tenant**: tulis dan uji prosedur membangun ulang dari VPS kosong dan memulihkan satu tenant tanpa menimpa tenant lain (gandengkan dengan A.5.30, draf). | DevOps Lead, Solution Architect | Sedang | [usulan] 28 Februari 2027 | Prosedur tertulis, catatan uji |
| 13 | **Backup Data Hub terjadwal** dan uji restore, termasuk keputusan lokasi dan pemilik pada deployment klien. | `auditsphere-data-analyst`, DevOps Lead | Sedang | [usulan] 28 Februari 2027 | Jadwal, log, hasil uji |
| 14 | **Daftarkan Evidence ID** pada Master Audit Evidence Index dan jadwalkan restore test bulanan. Commit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 1, 4, 7, 8, 9, 11, dan 13 menyentuh VPS, bucket, atau server lain; semuanya memerlukan konfirmasi eksplisit, dilakukan lewat agen devops/platform, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack, dan tidak mengubah skrip produksi tanpa persetujuan. Restore penuh ke produksi bersifat destruktif dan **tidak** boleh dijadikan bagian uji; uji hanya ke kontainer sementara.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.13-01 | Skrip backup terenkripsi dan integritas (`backup-prod.sh`, `decrypt-backup.sh`, `encrypt-existing-backups.sh`) | `backend/scripts/backup/` | Internal |
| EV-A8.13-02 | Runbook backup B2 (DRAFT; setelah tindakan 3 menjadi versi berlaku) | `docs/backup-b2-setup.md` | Internal |
| EV-A8.13-03 | Skrip restore test dan prosedur restore | `backend/scripts/backup/restore-test.sh`, `decrypt-backup.sh` | Internal |
| EV-A8.13-04 | Backup policy BCK-001, tabel RPO/RTO/retensi disetujui (setelah tindakan 2) | `docs/ISO_27001/A.8.13_Pencadangan_Informasi.md` (Lampiran A, B); [belum ada versi disetujui] | Internal |
| EV-A8.13-05 | Formulir dan log restore test terisi (setelah tindakan 5) | [belum ada] | Restricted |
| EV-A8.13-06 | Bukti cron aktif, versi skrip terpasang, dan daftar berkas backup di VPS (setelah tindakan 1 dan 4) | [belum ada keluaran VPS] | Restricted |
| EV-A8.13-07 | Daftar isi bucket B2 (nama saja), aturan lifecycle, dan kebijakan kunci aplikasi (tersanitasi) | [belum ada] | Restricted |
| EV-A8.13-08 | Catatan tinjauan job log berkala dan alert uji (setelah tindakan 8) | [belum ada] | Internal / Restricted |
| EV-A8.13-09 | Daftar pemegang kunci `age` dan catatan uji dekripsi (tanpa nilai kunci; setelah tindakan 6) | [belum ada] | Restricted |
| EV-A8.13-10 | Dump pra-deploy dan salin-keluar uploads (`deploy-prod.sh`) | `backend/deploy-prod.sh`, `.github/workflows/deploy.yml` | Internal |
| EV-A8.13-11 | Prosedur backup Data Hub dan hasil uji (setelah tindakan 13) | `data-server/docs/DEPLOY_DATAHUB_SAFE.md`; [belum ada hasil uji] | Internal / Restricted |
| EV-A8.13-12 | Kontrak/DPA dan keputusan lokasi off-site (setelah tindakan 10) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 8 selesai, tambahkan ID per siklus untuk setiap restore test dan setiap tinjauan job log bulanan.

---

## Lampiran A — Draf Kebijakan Pencadangan Informasi (BCK-001, usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | BCK-001 Kebijakan Pencadangan Informasi (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden, perubahan arsitektur, atau penggantian pemasok backup) |
| Berlaku untuk | Basis data dan berkas yang dikelola AuditSphere di produksi, stack tenant, Data Hub (milik AuditSphere), dan konfigurasi/rahasia yang diperlukan memulihkan layanan |

1. **Tujuan.** Menjamin informasi yang diperlukan untuk memulihkan layanan dapat dipulihkan dalam RPO dan RTO yang ditetapkan, tanpa membuka data klien kepada pihak yang tidak berwenang.
2. **Cakupan dan klasifikasi.** Cakupan mengikuti Lampiran B. Backup diklasifikasikan setara data sumber tertinggi yang dikandungnya (data audit klien dan rahasia: *Restricted*; A.5.12, draf penerapan).
3. **Frekuensi dan RPO/RTO.** Mengikuti Lampiran B; RPO/RTO disetujui manajemen dan ditinjau setelah tiap restore test.
4. **Enkripsi.** Semua backup dienkripsi sebelum ditulis ke disk dan sebelum meninggalkan VPS (`age`, atau alat setara yang disetujui). Kunci privat tidak disimpan di server yang dicadangkan. Backup polos hanya dengan pengecualian tertulis, penyetuju, dan tanggal berakhir (tidak ada "darurat" tanpa catatan). Dump pra-deploy dan salinan sementara mengikuti aturan yang sama.
5. **Lokasi.** Minimal satu salinan di luar VPS produksi. Pemilihan lokasi dan pemasok mempertimbangkan kewajiban lokasi data klien dan kontrak; keputusan dicatat. Tidak ada klaim residensi yang tidak terbukti.
6. **Kendali akses.** Hak baca backup hanya untuk peran yang ditunjuk (DevOps Lead dan satu pengganti); kunci aplikasi penyimpanan off-site memiliki cakupan terkecil; tidak ada kunci privat di VPS; kunci dan akses ditinjau setiap [6 bulan].
7. **Retensi.** Mengikuti Lampiran B; berkas yang melewati retensi dihapus otomatis; perilaku data yang sudah dihapus dari aplikasi tetapi masih ada di backup didokumentasikan (A.8.10, draf).
8. **Pengujian.** Restore test ke lingkungan sementara dengan kunci nyata, minimal [bulanan] untuk basis data dan [kuartalan] untuk uploads dan restore selektif per tenant; setiap uji dicatat (Lampiran C). Uji juga dilakukan setelah perubahan skrip backup, pergantian kunci, atau perubahan cakupan.
9. **Pemulihan.** Urutan pemulihan dari VPS kosong ditulis dan diuji [tahunan] (A.5.30, draf): server, Docker, Postgres, uploads, rahasia, Kong, nginx/TLS, DNS, verifikasi per layanan, dan komunikasi ke klien (A.5.24).
10. **Pemantauan.** Hasil setiap job dicatat, ditinjau sesuai Lampiran D, dan kegagalan atau keterlambatan menimbulkan alert ke DevOps Lead dan Lead Information Security (A.8.16, draf).
11. **Peran.** Sesuai RACI Lampiran E.
12. **Pengecualian dan pelanggaran.** Pengecualian tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir. Kehilangan atau kebocoran backup atau kunci diperlakukan sebagai insiden keamanan (A.5.24).

---

## Lampiran B — Usulan RPO, RTO, Retensi, dan Cakupan

**Seluruh angka adalah [usulan]** dan harus diputuskan manusia setelah restore test pertama memberi waktu pulih yang terukur. Kolom "Dasar di repo" memuat nilai bawaan teknis saat ini; itu bukan komitmen. RTO tidak boleh dijanjikan ke klien sebelum diukur.

| Kelas data | Frekuensi backup [usulan] | RPO [usulan] | RTO [usulan] | Retensi lokal [usulan] | Retensi off-site [usulan] | Dasar di repo / keterangan |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Postgres produksi (semua database, termasuk tenant) | [harian; tambah dump sebelum tiap deploy] | [24 jam] | [isi; ukur dulu pada restore test] | [14 hari] | [30 hari; pertimbangkan titik bulanan untuk [12 bulan]] | Harian 02:30 (runbook), retensi lokal 14 hari (bawaan skrip), lifecycle B2 30+1 hari (runbook). RPO sebenarnya sekitar satu siklus plus jeda; tidak ada pencadangan berkelanjutan (WAL) |
| Uploads audit-service (berkas bukti) | [harian] | [24 jam] | [isi] | [14 hari] | [30 hari] | Tercakup skrip (bila volume benar); dump pra-deploy menyalin keluar uploads |
| Uploads master-service | [harian setelah dicakup] | [24 jam] | [isi] | [14 hari] | [30 hari] | **Belum tercakup** (bagian 4 butir 6) |
| Uploads dan basis data stack tenant | [harian; restore selektif per tenant] | [24 jam] | [isi] | [14 hari] | [30 hari] | Basis data tercakup `pg_dumpall` bila di Postgres bersama; uploads tenant belum tercakup |
| Konfigurasi dan rahasia server (`.env`, kunci, konfigurasi nginx, `rclone.conf`) | [setelah setiap perubahan] | [perubahan terakhir] | [isi] | n/a | Penyimpanan rahasia terenkripsi terpisah dari backup data [keputusan] | Tidak ada prosedur. Jangan dicampur ke backup basis data |
| Redis (sesi dan cache) | [tidak dicadangkan; keputusan tertulis] | n/a | Dibangun ulang; pengguna login ulang | n/a | n/a | Keputusan perlu dicatat di SoA/Kebijakan; periksa apakah ada data yang tidak dapat dibangun ulang |
| Kafka dan ZooKeeper | [tidak dicadangkan; keputusan tertulis] | n/a | Dibangun ulang | n/a | n/a | Retensi topik 168 jam; event yang belum diproses dapat hilang |
| Model python-ai | [setelah pelatihan ulang] | [isi] | [isi] | [isi] | [isi] | Hanya volume; tidak ada backup |
| Monitoring (konfigurasi dan dasbor) | [setelah perubahan; konfigurasi di git] | n/a | [isi] | n/a | n/a | Metrik/log 7 hari tidak dicadangkan; konfigurasi server vs repo perlu disinkronkan |
| Data Hub (`datalake`, `hub_data`) | [harian bila ada klien produksi; ad hoc sebelum deploy saat ini] | [24 jam] | [isi] | [14 hari] | [30 hari; lokasi sesuai keputusan klien] | `DEPLOY_DATAHUB_SAFE.md` bagian 2; tanpa jadwal berkala |
| Training | Tidak dicadangkan (dibuat ulang dari produksi) | n/a | [isi] | n/a | n/a | `reset-training.sh`; dump sementara dihapus |
| Dump pra-deploy (`/root/db-backups`) | per deploy | n/a | n/a | [7 hari, dienkripsi] | Tidak dikirim | Saat ini tanpa retensi dan tanpa enkripsi |

Catatan keputusan yang diperlukan: (1) apakah RPO 24 jam diterima klien lembaga jasa keuangan atau diperlukan backup lebih sering/WAL; (2) apakah salinan off-site boleh di luar Indonesia untuk data klien; (3) berapa lama titik pulih bulanan/tahunan disimpan; (4) siapa yang berwenang memerintahkan restore produksi penuh.

---

## Lampiran C — Formulir Uji Restore (templat)

Salin satu formulir per pengujian. Jangan mencantumkan kunci, kata sandi, atau alamat IP.

| Field | Isi |
| :--- | :--- |
| ID uji (usulan) | RT-[tahun]-[nomor] |
| Tanggal / jam mulai - selesai | [isi] |
| Pelaksana (peran) / Peninjau (peran) | [DevOps Lead] / [Lead Information Security] |
| Jenis uji | [Basis data penuh / Uploads / Restore selektif satu tenant / Pemulihan dari VPS kosong] |
| Lingkungan uji | [mesin admin / kontainer sementara]; dikonfirmasi bukan produksi: [ya/tidak] |
| Berkas yang diuji (nama dan tanggal; tanpa isi) | pg: [isi] ; uploads: [isi] ; sumber: [lokal / off-site] |
| Usia berkas saat uji (indikasi RPO nyata) | [isi] jam |
| Checksum cocok | [ya/tidak] |
| Dekripsi berhasil dengan kunci milik: | [peran pemegang kunci]; kunci privat tidak di VPS: [ya/tidak] |
| Restore tanpa error | [ya/tidak]; ringkasan: [isi] |
| Database yang dipulihkan | [daftar nama] |
| Jumlah tabel/baris (perkiraan) vs sumber atau ekspektasi | [isi] |
| Pemeriksaan isi (sampel): login uji pada salinan, satu LHA/ATR terbaca, berkas bukti terbuka | [isi] |
| Jumlah berkas uploads dipulihkan vs sumber | [isi] |
| Waktu pemulihan terukur (indikasi RTO) | [isi] menit; tahapan: unduh [isi], dekripsi [isi], restore [isi], verifikasi [isi] |
| Hasil | [Lulus / Lulus dengan catatan / Gagal] |
| Temuan dan tindak lanjut | [isi]; pemilik: [peran]; tenggat: [isi] |
| Pembersihan (kontainer dan berkas sementara dihapus) | [ya/tidak] |
| Lampiran bukti (nama berkas log, tangkapan layar) | [EV-A8.13-05-xx] |
| Tanda tangan peninjau (peran) / tanggal | [isi] |

---

## Lampiran D — Daftar Periksa Tinjauan Job Log Backup (usulan)

Frekuensi tinjauan: [harian oleh DevOps Lead untuk status; mingguan ringkasan; bulanan ke Lead Information Security]. Setiap tinjauan dicatat (tanggal, peninjau, hasil).

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Job malam terakhir selesai dengan status sukses dan baris "done" tercatat; tidak ada baris `ERROR` atau `WARNING` (mis. peringatan backup polos atau "local only"). | [isi] | |
| 2 | Berkas `pg-*` dan `uploads-*` terbaru berekstensi `.age` dan memiliki berkas `.sha256`; ukuran tidak turun tajam dibandingkan hari sebelumnya [ambang: isi]. | [isi] | |
| 3 | Umur backup terbaru di lokal dan off-site tidak lebih dari [26 jam, usulan]. | [isi] | |
| 4 | Salinan off-site terbaru ada di bucket dan checksum cocok. | [isi] | |
| 5 | Retensi bekerja: tidak ada berkas lokal lebih tua dari retensi; direktori dump pra-deploy tidak menumpuk; ruang disk cukup [ambang: isi]. | [isi] | |
| 6 | Tidak ada berkas `.partial` tertinggal; tidak ada berkas polos baru di direktori backup. | [isi] | |
| 7 | Tidak ada perubahan tak terduga pada skrip terpasang (bandingkan hash dengan repo) dan pada cron. | [isi] | |
| 8 | Tidak ada akses atau kunci aplikasi baru pada bucket tanpa persetujuan; tidak ada login SSH ke akun backup yang tidak dikenal (A.8.15, draf). | [isi] | |
| 9 | Restore test bulan ini sudah dijalankan dan formulirnya terisi; temuan sebelumnya ditindaklanjuti. | [isi] | |
| 10 | Alert backup (gagal/terlambat) teruji bekerja pada kuartal ini. | [isi] | |
| 11 | Kegagalan atau keterlambatan dieskalasi sebagai insiden sesuai A.5.24; tiket/insiden: [isi]. | [isi] | |
| 12 | Daftar pemegang kunci dan akses ditinjau [6 bulan]; tinjauan terakhir: [isi]. | [isi] | |

---

## Lampiran E — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Legal & Compliance | Solution Architect | Direktur Utama | Data Analyst (Data Hub) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan BCK-001 | A/R | C | C | C | A (persetujuan) | I |
| Menetapkan RPO, RTO, dan retensi | R | R | C | C | A | C |
| Operasi backup harian dan retensi | C | A/R | I | I | I | R (Data Hub) |
| Enkripsi dan tata kelola kunci `age` | A | R | I | I | I | C |
| Akses backup dan kunci aplikasi B2 | A | R | I | I | I | C |
| Restore test dan formulir | A | R | I | C | I | R (Data Hub) |
| Tinjauan job log dan alert | C | A/R | I | I | I | C |
| Pemilihan lokasi dan pemasok backup, DPA | C | C | A/R | I | A (persetujuan) | I |
| Rencana pemulihan penuh | A | R | C | R | I | C |
| Restore produksi darurat | C | R | I | C | A (otorisasi) | I |
| Cakupan backup tenant | A | R | I | R | I | I |
| Evidence ID dan review berkala | A/R | C | C | I | I | I |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.8.21 Keamanan layanan jaringan | Ada (`docs/ISO_27001/A.8.21_Keamanan_Layanan_Jaringan.md`; belum ter-commit) | Backup ke B2 sebagai layanan jaringan; tindakan 12 dan runbook DRAFT; pemantauan dan alert (butir 4 dan tindakan 7). |
| A.5.12 Klasifikasi informasi | Ada | Klasifikasi backup mengikuti data sumber tertinggi. |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada | Backblaze B2 sebagai pemasok, DPA, dan tindakan 9. |
| A.5.24 Persiapan manajemen insiden | Ada | Backup sebagai kendali pemulihan; tindakan 12 (restore test). |
| A.7.3 Pengamanan kantor, ruangan, dan fasilitas | Ada | Lokasi fisik salinan, backup tidak terenkripsi menurut catatan. |
| A.5.19, A.5.21, A.5.22, A.5.23 | **Draf, dokumen belum ada** | Hubungan pemasok, rantai pasok, pemantauan layanan pemasok, layanan cloud (B2). Hanya A.5.20 yang ada. |
| A.5.30 Kesiapan TIK untuk kelangsungan bisnis | **Draf, dokumen belum ada** | RPO/RTO dan rencana pemulihan penuh akan menjadi masukan. |
| A.8.14 Redundansi fasilitas pemrosesan informasi | **Draf, dokumen belum ada** | Satu VPS tanpa replika; backup bukan pengganti redundansi. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Log backup dan akses ke backup. |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Alert backup gagal/terlambat. |
| A.8.24 Penggunaan kriptografi | **Draf, dokumen belum ada** | Enkripsi backup, pengelolaan kunci `age`, kunci aplikasi B2. |
| A.8.10 Penghapusan informasi | **Draf, dokumen belum ada** | Retensi dan penghapusan data pada backup. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.13	Technological	Pencadangan informasi	Backup sesuai kebutuhan, terlindungi, dan diuji pemulihannya	Backup policy, job log, restore test, encryption, retention	Berlaku	Data audit klien lembaga jasa keuangan di satu VPS tanpa replika (multi-tenant dalam satu Postgres); backup adalah satu-satunya jalur pemulihan; klien akan meminta bukti enkripsi, retensi, dan restore test (lihat A.8.13 bagian 1)	Sebagian	Sebagian (skrip dan runbook DRAFT ada di repo; job log, restore test, dan bukti pemasangan di VPS belum ada)	Belum terdaftar (usulan EV-A8.13-01 s/d EV-A8.13-12)	Lead Information Security (didukung DevOps Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah restore test pertama	Perlu perbaikan	Belum teruji	Tanpa kebijakan/RPO/RTO; tanpa bukti restore test; enkripsi terpasang di VPS tidak terbukti (catatan saling bertentangan); B2 di luar Indonesia; cakupan tidak lengkap (uploads master/tenant, Redis, Kafka, rahasia, Data Hub); tanpa job log/alert; kunci age tanpa tata kelola; dump pra-deploy tidak terenkripsi tanpa retensi	Pastikan keadaan VPS (read-only); setujui BCK-001; perbarui runbook; pasang backup terenkripsi; restore test pertama; tata kelola kunci age; perluas cakupan; job log dan alert; keputusan lokasi off-site (14 tindakan di bagian 5)	Tinggi	[usulan] 31 Desember 2026 (restore test dan kebijakan); 31 Januari 2027 (cakupan dan alert)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
