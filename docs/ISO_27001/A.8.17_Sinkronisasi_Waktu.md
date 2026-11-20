# A.8.17 — Sinkronisasi Waktu

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Sinkronisasi waktu |
| **Fokus pemeriksaan** | Sistem menggunakan sumber waktu yang disetujui dan konsisten |
| **Contoh bukti objektif** | NTP config, source list, drift alert, validation result |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi, stack tenant, training, Data Hub (server terpisah), stack monitoring, runner CI |
| **PIC** | DevOps Lead (didukung Lead Information Security dan Platform Engineer) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut seluruh sistem memakai **sumber waktu yang disetujui** dan **konsisten**. Fakta terpenting: **repo tidak memuat satu pun konfigurasi sinkronisasi waktu** (tidak ada `chrony`, `ntpd`, `systemd-timesyncd`, `timedatectl`, daftar sumber waktu, atau pengaturan zona waktu host). Kontainer Docker memakai jam kernel host, sehingga ketepatan waktu seluruh stack bergantung pada host VPS, dan host itu berada di luar repo. Karena itu semua hal tentang host (daemon sinkronisasi, sumber, offset aktual, zona waktu host, zona waktu cron) ditandai **Perlu verifikasi**. Yang dapat dibuktikan dari repo hanyalah **bagaimana aplikasi memakai waktu** (UTC pada penyimpanan, toleransi token, TTL, zona Asia/Jakarta untuk tanggal bisnis) dan **ketiadaan alert selisih waktu**. Tidak ada akses SSH dan tidak ada pengukuran offset yang dilakukan dalam penyusunan dokumen ini. Seluruh angka pada Lampiran A sampai D (ambang selisih, frekuensi, jumlah sumber) adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.17 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.17** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Sinkronisasi waktu** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Sistem menggunakan sumber waktu yang disetujui dan konsisten** | Sesuai baris spreadsheet; tiga gate: sumber disetujui (daftar dan konfigurasi), konsisten (semua host dan layanan, termasuk Data Hub), terpantau (alert selisih dan hasil validasi). |
| Contoh bukti objektif | **NTP config, source list, drift alert, validation result** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | Waktu menentukan keabsahan token JWT dan kode MFA (TOTP), urutan dan keandalan log aktivitas sebagai bukti audit, tanggal jatuh tempo ATR/tindak lanjut, retensi, nama dan umur backup, serta validasi sertifikat TLS (bagian 1). Pengecualian tidak dapat dipertahankan. |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: produk audit yang hasilnya berupa catatan bertanggal; kode MFA dan token bergantung jam server; log forensik (A.5.28, A.8.15 draf) harus dapat dikorelasikan; tanggal jatuh tempo memicu status OVERDUE; ada dua server (aplikasi dan Data Hub) yang jamnya harus konsisten. |
| Status implementasi | **Sebagian** | Sisi aplikasi: penyimpanan waktu dalam UTC, toleransi TOTP satu langkah, leeway 30 detik pada token step-up, `tzdata` di image Go, tanggal bisnis di Asia/Jakarta. Sisi host: tidak ada konfigurasi, sumber, bukti sinkron, maupun alert di repo. Status "Sebagian" hanya untuk sisi aplikasi; sinkronisasi host **belum terbukti**. |
| Ketersediaan bukti | **Sebagian (kode aplikasi) / Tidak ada (konfigurasi NTP, daftar sumber, drift alert, hasil validasi)** | Tidak ada satu pun dari empat contoh bukti objektif yang dapat diserahkan hari ini; yang ada hanyalah bukti kode tentang penggunaan waktu oleh aplikasi. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.17-01 s/d EV-A8.17-10) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **DevOps Lead** (didukung **Lead Information Security**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal setelah validasi sinkronisasi host pertama selesai | Selanjutnya: validasi [bulanan] dan tinjauan standar [tahunan] (Lampiran A, C). |
| Desain kontrol | **Perlu perbaikan** | Aplikasi dirancang cukup baik (UTC, toleransi terbatas), tetapi tidak ada standar sumber waktu, tidak ada konfigurasi yang dapat dibuktikan, zona waktu dan format stempel waktu tidak konsisten antar skrip dan layanan, dan tidak ada pemantauan selisih waktu. |
| Efektivitas operasional | **Belum teruji** | Tidak ada hasil validasi, log sinkronisasi, atau pengukuran offset. Efektivitas tidak dapat diklaim. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: tanpa konfigurasi/sumber waktu di repo; tanpa alert selisih (Prometheus tanpa `rule_files`; tidak ada metrik `timex`/NTP yang dipakai); dua server tanpa jaminan konsistensi; token JWT tanpa leeway; TOTP sensitif jam; stempel waktu backup tanpa zona; log syslog Promtail tanpa parsing waktu. |
| Tindakan perbaikan | **Lihat bagian 5 (12 tindakan)** | Draf Lampiran A sampai D sebagai masukan. |
| Prioritas | **Sedang** | Satu host untuk hampir seluruh layanan menekan risiko selisih antar-layanan, tetapi jam host yang salah langsung memutus login MFA dan merusak nilai bukti log; kontrol ini murah dipenuhi (konfigurasi dan satu alert). Naik menjadi Tinggi bila ada deployment multi-server klien. |
| Target | **[usulan] 31 Desember 2026** untuk bukti sinkronisasi host dan standar disetujui; **[usulan] 31 Januari 2027** untuk alert dan validasi berkala | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "sumber disetujui" dan "terpantau" belum terpenuhi; gate "konsisten" belum terbukti pada server. |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: bukti sinkronisasi host dan Data Hub, standar disetujui, alert selisih aktif. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 4 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.17 **berlaku** untuk AuditSphere karena:

1. **Hasil produk berupa catatan bertanggal.** KKA, LHA, ATR, persetujuan, dan log aktivitas adalah bukti audit; tanggal dan urutan kejadiannya menjadi bagian dari nilai bukti. Log aktivitas memberi stempel `OccurredAt` dari jam layanan penghasil (`backend/auth-service/pkg/activitylog/entry.go:97-109`), sehingga jam yang salah merusak urutan forensik (terkait A.5.28 dan A.8.15, draf).
2. **Autentikasi bergantung pada jam server.** Kode MFA TOTP dihitung dari `time.Now().Unix() / 30` dengan toleransi satu langkah di kedua arah (`backend/auth-service/services/auth/auth_service.go:454-470`, `backend/auth-service/services/mfa/mfa_service.go:384-398`); JWT memuat `exp`, `iat`, dan `nbf` (`backend/auth-service/pkg/utils/jwt.go:36-40`). Jam host yang bergeser lebih dari sekitar 30 detik membuat login MFA gagal atau, ke arah sebaliknya, memperlebar jendela penerimaan kode.
3. **Tanggal jatuh tempo memicu status.** Status ATR/tindak lanjut OVERDUE ditentukan oleh tanggal di zona Asia/Jakarta (`backend/audit-service/models/action-taken-report-status.go:59-81`). Selisih jam atau zona yang tidak konsisten menggeser hari kalender.
4. **Dua server yang harus konsisten.** Aplikasi di VPS produksi memanggil Data Hub di server terpisah melalui jaringan privat (`backend/docker-compose.prod.yml:124`; `data-server/docs/DEPLOY_DATAHUB_SAFE.md`). Korelasi log lintas server, penjadwalan sinkronisasi, dan validasi TLS memerlukan jam yang konsisten di kedua sisi.
5. **TLS dan retensi memakai waktu.** Validitas sertifikat (certbot pada `scripts/onboard-tenant.sh:399-410`; mode `verify-ca`/`verify-full` Data Hub pada `data-server/data-hub-api/core/db.py:64-65`), retensi log 7 hari di Loki (`monitoring/loki/loki-config.yml:36-39`), retensi Kafka 168 jam (`backend/docker-compose.prod.yml:66`), serta retensi backup berbasis `-mtime` (`backend/scripts/backup/backup-prod.sh:115`) semuanya bergantung pada jam yang benar.
6. **Klien dan auditor menanyakan.** Pertanyaan lazim due diligence: "sumber waktu apa yang dipakai, bagaimana selisihnya dipantau". Saat ini tidak ada dokumen atau bukti yang dapat diserahkan; A.5.28 sudah mencatat ketiadaan sinkronisasi waktu terdokumentasi sebagai gap (butir 9 dan tindakan 10 dokumen tersebut).

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti kode, konfigurasi, atau skrip di repo**. Keadaan host dan server adalah **Perlu verifikasi** kecuali dinyatakan lain.

### 2.1 Bagaimana waktu diperoleh dan dipakai di seluruh stack

| Komponen | Perilaku terlihat di repo | Bukti | Status / keterbatasan |
| :--- | :--- | :--- | :--- |
| Host VPS produksi (daemon sinkronisasi, sumber, zona waktu) | **Tidak ada konfigurasi di repo.** Pencarian `chrony`, `ntp`, `timesyncd`, `timedatectl`, `timex`, `clock skew` pada `backend/`, `monitoring/`, `data-server/`, `scripts/`, `.github/` tidak menemukan apa pun (selain komentar leeway di bagian 2.2). | (tidak ditemukan) | **Perlu verifikasi** pada host: daemon aktif, sumber, status sinkron, offset, zona waktu. |
| Kontainer Docker | Tidak ada pengaturan jam sendiri; kontainer memakai jam kernel host. Tidak ada `TZ` pada layanan produksi: `TZ: Asia/Jakarta` hanya pada `kafka-ui` di `docker-compose.yml:92`, `docker-compose.dev.yml:114`, `docker-compose.internal.yml:89` (semuanya di `backend/`); tidak ada di `docker-compose.prod.yml`. | `backend/docker-compose*.yml` | Zona waktu kontainer layanan produksi mengikuti bawaan image (diperkirakan UTC): **Perlu verifikasi** (`date` di kontainer). |
| Image dasar | Layanan Go (auth, audit, master, risk, analytics, kong-gateway) memakai `alpine:latest` dengan `apk add ca-certificates tzdata`. Python (python-ai, data-hub-api, ai-engine) memakai `python:3.10-slim`, frontend `node:20-alpine`, caatt-engine `jupyter/scipy-notebook`; tidak ada pemasangan `tzdata` eksplisit pada image Python/Node. Image `alpine:latest` tidak dipin. | `backend/*/Dockerfile:3` (Go); `backend/python-ai/Dockerfile:1`; `data-server/data-hub-api/Dockerfile:1`; `frontend/Dockerfile:1` | `tzdata` di image Go diperlukan agar `time.LoadLocation("Asia/Jakarta")` berhasil (lihat bawah). Image Python/Node tidak memerlukan karena memakai UTC. |
| Penyimpanan di Go (GORM) | `NowFunc` dipaksa UTC pada auth, audit, master, dan risk-service. Pada analytics-service `NowFunc` tidak ditemukan (waktu mengikuti zona proses, diperkirakan UTC). | `backend/auth-service/pkg/database/postgres.go:28`; `backend/audit-service/pkg/database/postgres.go:28`; `backend/master-service/pkg/database/postgres.go:28`; `backend/risk-service/pkg/database/postgres.go:27` | DSN tidak memuat parameter `TimeZone`; zona sesi Postgres bawaan image `postgres:16` (`backend/docker-compose.prod.yml`): **Perlu verifikasi** (`SHOW timezone`). |
| Tanggal bisnis (jatuh tempo) | Zona Asia/Jakarta untuk keputusan hari kalender ATR dan tindak lanjut; `LoadLocation` dengan fallback bila `tzdata` hilang. RCM COSO memakai Asia/Jakarta untuk `as_of`. | `backend/audit-service/models/action-taken-report-status.go:59-81`; `backend/audit-service/models/follow-up-status.go:29,63`; `backend/risk-service/controllers/rcm_coso.go:451,500` | Konsisten pada layanan Go; tidak berlaku pada python-ai (lihat butir bawah). |
| Python AI | `datetime.now()` tanpa zona untuk `date` keluaran dan sufiks ID. Pada kontainer UTC, tanggal ini dapat berbeda satu hari dari WIB pada jam 00:00-07:00 WIB. | `backend/python-ai/main.py:203,1300,1332`; `data-server/ai-engine/main.py:186,1114,1146` | Risiko kecil (tanggal tampilan/ID), dicatat di bagian 4 butir 9. |
| JWT | Token memuat `exp` (default 24 jam), `iat`, `nbf`; validasi mewajibkan `exp`; **tidak ada leeway** pada validasi token akses. | `backend/auth-service/pkg/utils/jwt.go:36-40,55-56`; `backend/auth-service/pkg/config/config.docker.yaml:22`; middleware: `backend/audit-service/pkg/middleware/auth.go:64`, `backend/master-service/pkg/middleware/auth.go:58`, `backend/risk-service/middleware/auth.go:74`, `backend/analytics-service/middleware/auth.go:70`, `backend/auth-service/pkg/middleware/auth.go:73` | Satu host berarti satu jam, sehingga aman; pada deployment multi-host, `nbf = iat = waktu penerbitan` dan tanpa leeway membuat layanan dengan jam tertinggal menolak token baru ("belum berlaku"). |
| Token Redis | Kunci `token:<jwt>` disimpan dengan TTL `time.Until(expiresAt)`; kunci lain: 10 menit pada alur login, 5 menit untuk setup MFA dan trusted device tertunda. | `backend/auth-service/services/auth/auth_service.go:251-254,188`; `backend/auth-service/services/mfa/mfa_service.go:331`; `backend/auth-service/services/trusted-devices/trusted_devices_service.go:57` | TTL dihitung relatif (durasi), sehingga tahan terhadap zona waktu; sensitif terhadap lompatan jam bila jam host dikoreksi besar-besaran. |
| TOTP (MFA) | Langkah 30 detik; menerima langkah saat ini, sebelumnya, dan berikutnya (jendela efektif sekitar +/-30 detik). Logika ini **diduplikasi** di dua tempat. | `backend/auth-service/services/auth/auth_service.go:454-470`; `backend/auth-service/services/mfa/mfa_service.go:384-398` | Toleransi terbatas itu baik, tetapi menjadi batas kerja: selisih jam host lebih dari sekitar 30 detik memutus login MFA. Tidak ada ambang alert yang selaras. |
| Step-up token | TTL 5 menit, leeway 30 detik antara auth-service dan master-service (komentar eksplisit "tolerates clock skew"), jti sekali pakai di Redis. | `backend/master-service/pkg/middleware/step_up.go:30,36-37,105,129`; `backend/auth-service/pkg/utils/step_up.go:30,55-56` | Satu-satunya tempat leeway diterapkan; menunjukkan pemahaman risiko selisih. |
| Trusted device | Kedaluwarsa 90 hari; pemeriksaan `ExpiresAt.Before(time.Now())`. | `backend/auth-service/repositories/trusted_device_repository.go:35,113`; `backend/auth-service/services/trusted-devices/trusted_devices_service.go:93` | Perbandingan instan, tahan zona; bergantung jam host. |
| Penguncian MFA / step-up | Jendela 15 menit untuk kegagalan dan kunci. | `backend/auth-service/services/auth/mfa_lockout.go:22`; `backend/auth-service/services/auth/step_up_service.go:24` | Durasi relatif. |
| Log aktivitas | `OccurredAt` dinormalisasi ke UTC dan dipotong ke mikrodetik; retensi dijalankan tiap 24 jam dengan `cutoff` dari `time.Now()`. | `backend/auth-service/pkg/activitylog/entry.go:97-109`; `backend/auth-service/pkg/activitylog/retention.go:29,43`; `backend/auth-service/cmd/auth/serve.go:136` | Stempel diberikan oleh jam layanan penghasil (bukan broker/DB), jadi mengikuti jam host. UTC baik untuk korelasi. |
| Kafka | Event membawa `Timestamp: time.Now().UTC()` (master, audit) dan `OccurredAt ... UTC()` (risk); ID event memakai `UnixNano`. Retensi topik 168 jam. | `backend/master-service/pkg/kafka/events.go:133-136`; `backend/audit-service/pkg/kafka/producer.go:118,187`; `backend/risk-service/pkg/kafka/producer.go:49`; `backend/docker-compose.prod.yml:66` | Stempel pesan diberikan produsen; apakah Kafka memakai CreateTime atau LogAppendTime untuk retensi: **Perlu verifikasi** (default topik). Tidak ada `TZ` pada Kafka/ZooKeeper/Redis. |
| Redis | `redis:7-alpine`; tidak ada pengaturan waktu. TTL kunci dihitung oleh Redis dari jam host. | `backend/docker-compose.prod.yml:20` | Bergantung host. |
| Data Hub (server terpisah) | Penjadwal worker memakai **durasi relatif**: `EXTRACT(EPOCH FROM (NOW() - last_sync_at))` pada Postgres Data Hub dan loop polling 10 detik; `last_sync_at` diisi `NOW()`. Muatan Bronze memakai `TIMESTAMPTZ` dan `datetime.now(timezone.utc)`. Watermark disimpan sebagai nilai kolom sumber (`ops.sync_watermarks`). | `data-server/data-hub-api/worker.py:170,200-201,257-260`; `data-server/data-hub-api/ingestion/watermark.py:28-39`; `data-server/data-hub-api/ingestion/loader.py:43`; `data-server/data-hub-api/ingestion/bronze_ddl.py:62` | Desain relatif dan UTC baik. Konsistensi jam server Data Hub dengan VPS aplikasi dan dengan basis data sumber klien **tidak terbukti** (Perlu verifikasi). Autentikasi aplikasi-ke-Data Hub memakai API key (bukan JWT), sehingga leeway JWT tidak menjadi masalah lintas server. |
| Scheduler dalam aplikasi | Hanya ticker berbasis interval (retensi log aktivitas 24 jam; analyzer audit completion; health check dan rate limiter Kong). Tidak ada ekspresi cron atau zona terjadwal di kode Go. | `backend/audit-service/services/audit_completion/analyzer.go:90`; `backend/auth-service/pkg/activitylog/retention.go:40`; `backend/kong-gateway/pkg/gateway/proxy.go:338` | Tidak peka zona. |
| Cron host | Hanya runbook backup yang menyebut cron "02:30 server time"; tidak ada berkas cron di repo; workflow CI tidak memakai `schedule:`. | `docs/backup-b2-setup.md:100` | Zona waktu server dan cron nyata: **Perlu verifikasi**. |
| Backup (nama dan retensi) | Nama berkas dan baris log memakai **waktu lokal host tanpa penanda zona** (`date +%Y%m%d-%H%M%S`; `date '+%F %T'`); retensi memakai `find -mtime` (berbasis epoch, tahan zona). `deploy-prod.sh` dan skrip training memakai **UTC** dengan akhiran `Z`. | `backend/scripts/backup/backup-prod.sh:51,53,115`; `backend/deploy-prod.sh:78,114`; `backend/training/lib.sh:20`; `backend/training/reset-training.sh:28` | Format stempel tidak konsisten antar skrip (bagian 4 butir 8). |
| Monitoring | node-exporter dijalankan tanpa flag yang mematikan kolektor, dan tidak ada aturan/dasbor di repo yang memakai metrik `node_timex_*` atau `node_time_seconds`. Prometheus: tidak ada `rule_files` dan tidak ada `alerting`. Loki menolak sampel lebih tua dari 168 jam. | `monitoring/docker-compose.monitoring.yml:74-86`; `monitoring/prometheus/prometheus.yml` (tidak ada `rule_files`); `monitoring/loki/loki-config.yml:36-39` | Kolektor `timex` aktif secara bawaan pada node-exporter upstream (pengetahuan umum; image dipasang `:latest`, **Perlu verifikasi** versi): metrik mungkin sudah terkumpul tetapi **tidak ada alert dan tidak ada dasbor**. |
| Promtail | Job `docker` memakai stage `docker: {}` (stempel dari log Docker). Job `syslog` menangkap grup `timestamp` lewat regex tetapi **tidak ada stage `timestamp`** yang memakainya, jadi stempel asli (tanpa tahun dan zona) tidak diparse dan stempel Loki adalah waktu ingest. | `monitoring/promtail/promtail-config.yml:28-29,46-52` | Waktu kejadian syslog tidak tersimpan sebagai stempel Loki (bagian 4 butir 10). |
| TLS | Sertifikat diterbitkan certbot saat onboarding tenant; tidak ada pemantauan masa berlaku (A.8.21). | `scripts/onboard-tenant.sh:399-410`; `docs/ISO_27001/A.8.21_Keamanan_Layanan_Jaringan.md` bagian 4 butir 4 | Validitas TLS bergantung jam klien dan server. |
| Stack tenant dan training | Template tenant tidak mengatur `TZ`; semua berjalan di host yang sama dengan produksi (satu jam kernel). Skrip training memakai UTC. | `scripts/templates/docker-compose.tenant.yml.tpl` (tidak ada `TZ`); `backend/training/lib.sh:20` | Satu host, satu jam; tidak ada isolasi waktu antar tenant (tidak diperlukan, tetapi berarti satu kegagalan sinkronisasi memengaruhi semua tenant). |
| CI | Workflow berjalan di `ubuntu-latest` (runner terkelola GitHub). | `.github/workflows/deploy.yml:37`; `.github/workflows/deploy-dev.yml:15` | Sinkronisasi waktu runner adalah tanggung jawab penyedia (di luar kendali AuditSphere; Perlu verifikasi bila ada runner self-hosted). Stempel deploy memakai UTC. |

### 2.2 Kendali toleransi waktu yang terverifikasi dari kode

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Penyimpanan UTC | GORM `NowFunc` UTC; log aktivitas UTC; event Kafka UTC; Data Hub `TIMESTAMPTZ`. | Lihat 2.1 | Analytics-service tidak memaksa UTC (bergantung zona proses). |
| Leeway terbatas | Step-up 30 detik, dan TTL token step-up dibatasi `StepUpTTL + leeway`. | `backend/master-service/pkg/middleware/step_up.go:105,129` | Tidak dipakai pada token akses (nol leeway). |
| Toleransi TOTP | +/-1 langkah (30 detik). | `auth_service.go:454-470`; `mfa_service.go:384-398` | Dua salinan logika; perubahan toleransi harus dilakukan di dua tempat. |
| `tzdata` pada image Go | Terpasang agar `Asia/Jakarta` dapat dimuat. | `backend/*/Dockerfile:3` | Image Python tidak memuat tanggal bisnis. |

### 2.3 Dokumen yang menyinggung waktu

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| A.5.28 (annex ada) | Mencatat ketiadaan sinkronisasi waktu (NTP/chrony) yang terdokumentasi (butir 9), tindakan 10 menugaskan penetapan sinkronisasi waktu dan zona waktu pada log, serta FRP-001 menuntut waktu koleksi "NTP Synchronized" dan baris "sumber waktu dan selisih" pada formulir. | Dokumen ini adalah tindak lanjut tindakan 10 tersebut. |
| A.8.21 (annex ada) | Prometheus tanpa `rule_files`/`alerting`; aturan alert tidak termuat; tidak ada pemantauan masa berlaku TLS. | Berlaku sama untuk alert selisih waktu. |
| A.8.13 (annex ada) | Backup, retensi, dan cron; tidak membahas format stempel waktu. | Lihat bagian 4 butir 8. |

### 2.4 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `backend/`, `scripts/`, `data-server/`, `monitoring/`, `docs/`, `.github/` (19 November 2026) tidak menemukan: standar atau kebijakan sinkronisasi waktu; konfigurasi `chrony`/`ntpd`/`systemd-timesyncd` atau skrip yang memasangnya; daftar sumber waktu yang disetujui; pengaturan zona waktu host (`timedatectl set-timezone`); aturan alert untuk selisih jam atau status sinkronisasi (`node_timex_*`, `node_timex_sync_status`, `node_time_seconds`); dasbor Grafana untuk offset; hasil validasi sinkronisasi; pemeriksaan selisih antar server (aplikasi vs Data Hub); pemeriksaan jam pada skrip deploy (`deploy-prod.sh`) atau skrip provisioning tenant; prosedur koreksi jam setelah reboot tidak bersih (terkait catatan Kong crash loop pada memori operasional); pengaturan `TZ` pada layanan produksi. Seluruh keadaan host **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| NTP config (konfigurasi daemon sinkronisasi) | **Belum tersedia / Perlu verifikasi** | Tidak ada di repo. Ekspor read-only dari host VPS dan server Data Hub (konfigurasi tanpa rahasia; Lampiran C butir 1). |
| Source list (daftar sumber waktu yang disetujui) | **Belum tersedia; draf di Lampiran B** | Tidak ada keputusan tertulis; kandidat sumber di Lampiran B adalah usulan. |
| Standar/kebijakan sinkronisasi waktu | **Belum tersedia; draf di Lampiran A** | Kode usulan SYN-001. Perlu persetujuan; angka [usulan]. |
| Drift alert | **Belum tersedia; persyaratan di Lampiran D** | Prometheus tanpa `rule_files`/Alertmanager; tidak ada aturan selisih waktu. |
| Validation result (hasil validasi) | **Belum tersedia; formulir di Lampiran C** | Tidak ada keluaran `chronyc tracking`/`timedatectl` yang tersimpan. A.5.28 Lampiran memuat templat koleksi bukti yang menyebut `timedatectl` (UTC) dan "sumber waktu dan selisih". |
| Bukti zona waktu host dan cron | **Belum tersedia / Perlu verifikasi** | `timedatectl` dan `crontab -l` (read-only). |
| Bukti offset antar server (aplikasi vs Data Hub) | **Belum tersedia** | Dua pembacaan serentak atau metrik offset kedua host. |
| Bukti perilaku aplikasi (UTC, toleransi, TTL) | **Tersedia (kode)** | Bagian 2.1 dan 2.2; uji unit yang ada tidak khusus menguji selisih jam. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Tidak ada]** tidak ditemukan; **[Perlu verifikasi]** bergantung pada host/server.

1. **[Tidak ada] Tidak ada konfigurasi, sumber, atau bukti sinkronisasi waktu di repo.** Seluruh akurasi waktu stack bergantung pada daemon host yang tidak terdokumentasi. Apakah VPS produksi, server Data Hub, dan VPS lain menjalankan chrony, `systemd-timesyncd`, atau tidak sama sekali, serta sumber yang dipakai: **Perlu verifikasi**. Sampai itu dilakukan, **jangan klaim** "waktu tersinkronisasi dengan NTP".
2. **[Tidak ada] Tidak ada drift alert.** Prometheus tidak memuat `rule_files` atau `alerting` (`monitoring/prometheus/prometheus.yml`; A.8.21 butir 4), dan tidak ada aturan atau dasbor yang memakai `node_timex_*`. Selisih jam yang membesar tidak akan diketahui sampai pengguna mengeluh login MFA gagal. Metrik `timex` mungkin sudah dikumpulkan node-exporter (bawaan upstream; Perlu verifikasi) sehingga alert dapat ditambahkan tanpa komponen baru.
3. **[Repo] Login MFA sensitif jam.** TOTP menerima +/-1 langkah 30 detik (`auth_service.go:454-470`, `mfa_service.go:384-398`). Jam host yang bergeser lebih dari sekitar 30 detik menyebabkan kegagalan login massal dan, melalui penguncian 15 menit (`mfa_lockout.go:22`), dapat mengunci pengguna sah. Logika TOTP diduplikasi di dua file sehingga toleransi dapat menyimpang.
4. **[Repo] Token akses tanpa leeway.** Semua layanan memakai `WithExpirationRequired()` tanpa leeway, sedangkan token memuat `nbf = iat = sekarang` (`jwt.go:36-40`). Pada satu host ini aman; pada deployment multi-host (mis. stack tenant on-premises klien dengan layanan di beberapa server) layanan yang jamnya tertinggal akan menolak token baru. Hanya step-up yang memiliki leeway 30 detik (`step_up.go:36-37`). Perubahan kode bukan bagian dokumen ini.
5. **[Repo] Dua server tanpa jaminan konsistensi.** VPS aplikasi dan server Data Hub berjalan terpisah; tidak ada pemeriksaan selisih jam, tidak ada sumber waktu bersama yang ditetapkan, dan korelasi log lintas server memerlukan UTC konsisten. Penjadwal Data Hub bersifat relatif (aman terhadap zona), tetapi stempel `_loaded_at`, `last_sync_at`, dan watermark akan dibandingkan dengan jam basis data sumber klien (**Perlu verifikasi** untuk setiap sumber). Pada deployment on-premises klien, kepemilikan sinkronisasi waktu perlu ditetapkan (selaras A.8.13 butir 14).
6. **[Perlu verifikasi] Zona waktu host dan cron tidak diketahui.** Runbook backup menyebut 02:30 "server time"; tanpa zona waktu host yang tercatat, jam eksekusi dan nama berkas backup ambigu (WIB vs UTC, selisih 7 jam). Tidak ada `TZ` pada layanan produksi sehingga zona kontainer mengikuti bawaan image.
7. **[Repo] Image dasar tidak dipin.** `alpine:latest` dan `node-exporter:latest` berarti `tzdata` dan perilaku kolektor dapat berubah antar build. Bukan inti kontrol ini, tetapi memengaruhi konsistensi (selaras A.8.21 tindakan pin versi).
8. **[Repo] Format stempel waktu tidak konsisten antar skrip.** `backup-prod.sh` memakai waktu lokal tanpa penanda zona (`:51,53`), sedangkan `deploy-prod.sh` dan skrip training memakai UTC dengan akhiran `Z` (`deploy-prod.sh:78`, `lib.sh:20`, `reset-training.sh:28`). Mencocokkan berkas backup dengan peristiwa deploy atau insiden menjadi rawan salah tafsir; A.5.28 meminta zona waktu dicatat pada log dan formulir.
9. **[Repo] Python AI memakai waktu lokal tanpa zona.** `datetime.now()` untuk tanggal keluaran dan sufiks ID (`backend/python-ai/main.py:203,1300,1332`; `data-server/ai-engine/main.py:186,1114,1146`). Pada kontainer UTC, tanggal dapat berbeda satu hari dari WIB; layanan Go memakai Asia/Jakarta untuk tanggal bisnis. Dampak rendah, tetapi inkonsistensi.
10. **[Repo] Log syslog Promtail tanpa parsing waktu.** Job `syslog` menangkap `timestamp` tanpa stage yang memakainya (`promtail-config.yml:46-52`), sehingga waktu kejadian syslog host (mis. SSH, kernel) adalah waktu ingest, bukan waktu kejadian; ini melemahkan korelasi forensik (A.8.15, draf). Apakah job tersebut berfungsi di VPS (`/var/log/messages` tidak ada pada distro berbasis Debian): **Perlu verifikasi**.
11. **[Repo] Log aplikasi tidak menjamin zona pada keluarannya.** Stempel tersimpan UTC pada log aktivitas (baik), tetapi log layanan memakai logger terpisah (zap) dan format tampilan tidak diperiksa menyeluruh pada dokumen ini; sebelum klaim "semua log berstempel UTC", periksa konfigurasi encoder waktu (**Perlu verifikasi**).
12. **[Repo] Ketergantungan Kafka pada jam produsen.** Stempel event diberikan oleh produsen (`time.Now().UTC()`); ID event memakai `UnixNano` yang tidak monoton bila jam dikoreksi mundur. Kebijakan stempel topik (CreateTime vs LogAppendTime) dan efeknya pada retensi 168 jam: **Perlu verifikasi**.
13. **[Tidak ada] Tidak ada pemeriksaan jam pada prosedur reboot/pemulihan.** Memori operasional mencatat reboot tidak bersih yang menimbulkan crash loop Kong; tidak ada langkah "verifikasi jam" pada runbook pemulihan. Jam yang salah setelah reboot pada VPS tanpa RTC akurat dapat merusak validasi TLS dan token sampai sinkronisasi pulih.
14. **[Tidak ada] Tidak ada daftar sumber waktu pada pemasok.** Apakah penyedia VPS menyediakan sumber waktu internal, dan apakah paparan jaringan keluar (UDP 123) diizinkan firewall, tidak terdokumentasi (A.8.21 Lampiran D, draf).
15. **[Repo] Runner CI dan stempel.** Runner GitHub terkelola; AuditSphere tidak mengontrol sinkronisasinya, tetapi stempel deploy memakai UTC. Bila kelak ada runner self-hosted, ia harus masuk cakupan standar.
16. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet dalam penyusunan dokumen ini; tidak ada akses SSH.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Ukur keadaan nyata sinkronisasi** (read-only, atas konfirmasi eksplisit, lewat `auditsphere-devops`/`auditsphere-platform`): pada VPS produksi dan server Data Hub catat daemon aktif, status sinkron, sumber, offset, zona waktu (`timedatectl`, `chronyc tracking`/`chronyc sources` atau `timedatectl timesync-status`), serta `crontab -l`; di kontainer, `date` dan `TZ`; Postgres `SHOW timezone`. Jangan simpan alamat IP atau berkas berisi rahasia. | DevOps Lead | Tinggi | [usulan] 30 November 2026 | Ekspor read-only bertanggal, disaring (EV-A8.17-01) |
| 2 | **Tetapkan standar SYN-001** (Lampiran A, B): sumber disetujui, zona waktu baku (UTC pada penyimpanan/log, WIB untuk tampilan bisnis), ambang selisih, kepemilikan (termasuk deployment klien). | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | Standar disetujui (EV-A8.17-04) |
| 3 | **Pasang/seragamkan sinkronisasi pada semua host** sesuai standar (satu daemon, dua atau lebih sumber, konfigurasi tersimpan sebagai ekspor atau sebagai berkas di repo). Perubahan host hanya atas konfirmasi eksplisit, satu host per sesi. | `auditsphere-devops`, DevOps Lead | Tinggi | [usulan] 31 Desember 2026 | Konfigurasi tersanitasi, hasil validasi (EV-A8.17-02, -03) |
| 4 | **Validasi pertama dan hasil tertulis** (Lampiran C), termasuk selisih antara VPS aplikasi dan server Data Hub; dasar pembaruan formulir A.5.28 (baris "sumber waktu dan selisih"). | DevOps Lead (pelaksana), Lead Information Security (peninjau) | Tinggi | [usulan] 31 Desember 2026 | Formulir terisi (EV-A8.17-05) |
| 5 | **Drift alert** (Lampiran D): muat `rule_files` ke Prometheus (bersama tindakan 7 A.8.21), tambahkan aturan berbasis metrik `timex`/`time` node-exporter, uji dengan simulasi, dan kirim ke kanal notifikasi. Perubahan konfigurasi monitoring melalui agen platform atas persetujuan. | `auditsphere-platform`, DevOps Lead | Tinggi | [usulan] 31 Januari 2027 | Berkas aturan, bukti alert uji (EV-A8.17-06) |
| 6 | **Dasbor offset** untuk semua host (satu panel) dan tinjauan bulanan. | `auditsphere-platform` | Sedang | [usulan] 31 Januari 2027 | Tangkapan layar dasbor, catatan tinjauan |
| 7 | **Seragamkan format stempel waktu skrip**: UTC dengan akhiran `Z` pada nama berkas dan log (`backup-prod.sh`) dan catat zona cron; selaraskan runbook backup. Perubahan skrip dilakukan oleh agen devops atas persetujuan, bukan oleh dokumen ini. | `auditsphere-devops` | Sedang | [usulan] 31 Januari 2027 | Skrip dan runbook revisi |
| 8 | **Perbaiki parsing waktu Promtail syslog** (stage `timestamp` dengan format dan zona), atau putuskan job syslog tidak dipakai; verifikasi jalur log host. | `auditsphere-platform` | Sedang | [usulan] 31 Januari 2027 | Konfigurasi dan sampel log |
| 9 | **Evaluasi toleransi waktu pada aplikasi**: tinjau leeway token akses untuk deployment multi-host, konsolidasi logika TOTP menjadi satu implementasi, dan seragamkan tanggal Python AI ke zona yang diputuskan. Keputusan desain; implementasi oleh agen backend/data-analyst atas persetujuan. | `auditsphere-backend`, `auditsphere-data-analyst`, Solution Architect | Sedang | [usulan] 28 Februari 2027 | Keputusan tertulis; perubahan terdokumentasi |
| 10 | **Tambahkan pemeriksaan jam ke runbook** pemulihan setelah reboot dan provisioning tenant/host baru (langkah "verifikasi sinkron sebelum menjalankan stack"), serta syarat sinkronisasi waktu pada deployment on-premises klien. | DevOps Lead, Solution Architect | Sedang | [usulan] 28 Februari 2027 | Runbook revisi |
| 11 | **Kunci image dan versi** (`alpine`, `node-exporter`) agar `tzdata` dan kolektor stabil (selaras A.8.21). | `auditsphere-devops` | Rendah | [usulan] 28 Februari 2027 | Perubahan terdokumentasi |
| 12 | **Daftarkan Evidence ID** pada Master Audit Evidence Index, jadwalkan validasi bulanan, dan komit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 1, 3, 5, 6, 8, dan 10 menyentuh VPS, Data Hub, atau konfigurasi monitoring; semuanya memerlukan konfirmasi eksplisit, dilakukan lewat agen devops/platform, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack. Koreksi jam besar-besaran pada host berjalan dapat menggeser TTL Redis dan memutus sesi; jadwalkan pada jendela perawatan.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.17-01 | Ekspor read-only status sinkronisasi dan zona waktu host (VPS produksi dan server Data Hub), disaring (setelah tindakan 1) | [belum ada keluaran host] | Restricted |
| EV-A8.17-02 | Konfigurasi sinkronisasi waktu (tanpa rahasia), termasuk daftar sumber (setelah tindakan 3) | [belum ada] | Internal |
| EV-A8.17-03 | Hasil validasi setelah pemasangan/penyeragaman (setelah tindakan 3) | [belum ada] | Internal |
| EV-A8.17-04 | Standar SYN-001, daftar sumber dan ambang disetujui (setelah tindakan 2) | `docs/ISO_27001/A.8.17_Sinkronisasi_Waktu.md` (Lampiran A, B); [belum ada versi disetujui] | Internal |
| EV-A8.17-05 | Formulir validasi sinkronisasi terisi, termasuk selisih lintas server (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.17-06 | Aturan drift alert dan bukti alert uji (setelah tindakan 5) | [belum ada]; konfigurasi monitoring di `monitoring/` | Internal |
| EV-A8.17-07 | Kode toleransi waktu aplikasi (JWT, TOTP, step-up, UTC) | `backend/auth-service/pkg/utils/jwt.go`; `backend/auth-service/services/auth/auth_service.go`; `backend/auth-service/services/mfa/mfa_service.go`; `backend/master-service/pkg/middleware/step_up.go` | Internal |
| EV-A8.17-08 | Kode penyimpanan UTC dan log aktivitas | `backend/*/pkg/database/postgres.go`; `backend/auth-service/pkg/activitylog/entry.go` | Internal |
| EV-A8.17-09 | Perilaku waktu Data Hub (penjadwal relatif, watermark, UTC) | `data-server/data-hub-api/worker.py`; `data-server/data-hub-api/ingestion/` | Internal |
| EV-A8.17-10 | Catatan tinjauan dasbor offset bulanan dan keputusan evaluasi toleransi aplikasi (setelah tindakan 6 dan 9) | [belum ada] | Internal |

Setelah tindakan 1 sampai 5 selesai, tambahkan ID per siklus untuk setiap validasi bulanan.

---

## Lampiran A — Draf Standar Sinkronisasi Waktu (SYN-001, usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | SYN-001 Standar Sinkronisasi Waktu (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | DevOps Lead |
| Penyetuju | Lead Information Security; Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah perubahan arsitektur, penambahan server, atau penggantian penyedia VPS) |
| Berlaku untuk | Semua host dan kontainer yang menjalankan AuditSphere: VPS produksi, stack tenant, training, Data Hub, monitoring, serta server klien pada deployment yang dikelola AuditSphere |

1. **Tujuan.** Memastikan seluruh sistem memakai waktu yang konsisten dan dapat dibuktikan, agar token, kode MFA, log, tanggal jatuh tempo, retensi, dan sertifikat berfungsi benar, dan agar log dapat dikorelasikan lintas sistem.
2. **Sumber waktu.** Setiap host memakai satu daemon sinkronisasi yang disetujui (Lampiran B) dengan minimal [4] sumber atau kumpulan sumber yang independen; sumber hanya dari daftar yang disetujui; perubahan daftar melalui persetujuan pemilik standar.
3. **Kontainer.** Kontainer tidak menjalankan daemon sinkronisasi sendiri dan tidak mengubah jam; jam kontainer mengikuti host. Kontainer tidak diberi hak mengubah waktu sistem.
4. **Zona waktu dan format.** Host dan kontainer produksi memakai UTC [usulan]; penyimpanan, log, event, dan nama berkas memakai UTC dengan format RFC 3339 atau akhiran `Z`; tanggal bisnis (jatuh tempo) dihitung di zona Asia/Jakarta secara eksplisit di aplikasi dan bukan dari zona host. Zona cron dicatat di runbook.
5. **Ambang selisih.** Mengikuti Lampiran B; selisih yang melampaui ambang dieskalasi (Lampiran D).
6. **Konsistensi lintas server.** Semua server dalam satu deployment memakai sumber dari daftar yang sama; selisih antar server diukur dan dicatat (Lampiran C). Pada deployment on-premises klien, kewajiban sinkronisasi dicantumkan dalam dokumen deployment.
7. **Pemantauan.** Offset dan status sinkron dikumpulkan sebagai metrik, ditinjau [bulanan], dan alert berlaku sesuai Lampiran D (A.8.16, draf).
8. **Validasi.** Validasi sinkronisasi dilakukan setelah pemasangan, reboot tidak bersih, perubahan jaringan/firewall, dan [bulanan] (Lampiran C).
9. **Pemulihan.** Setelah reboot atau pemulihan dari VPS baru, status sinkron diverifikasi sebelum layanan dinyatakan siap (A.5.24, A.8.13).
10. **Peran.** Sesuai RACI Lampiran E.
11. **Pengecualian dan pelanggaran.** Pengecualian tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir. Selisih jam yang berdampak pada bukti diperlakukan sebagai insiden sesuai A.5.24.

---

## Lampiran B — Usulan Sumber Waktu dan Ambang Selisih

**Seluruh angka dan nama sumber adalah [usulan]** dan harus diputuskan manusia setelah pengukuran pertama (tindakan 1). Tidak ada nilai di bawah yang merupakan hasil pengukuran atau pernyataan keadaan saat ini.

### B.1 Daftar sumber waktu yang disetujui (usulan, belum diputuskan)

| No | Kategori sumber | Contoh/Kandidat | Status |
| :--- | :--- | :--- | :--- |
| 1 | Server waktu internal penyedia VPS | [isi: nama dari dokumentasi penyedia; konfirmasi ketersediaan] | [usulan; Perlu verifikasi] |
| 2 | Kumpulan NTP publik | `pool.ntp.org` atau kumpulan regional [isi] | [usulan] |
| 3 | Sumber di Indonesia (bila diwajibkan klien/regulator) | [isi setelah konfirmasi] | [usulan; keputusan Legal & Compliance] |
| 4 | Sumber internal untuk server on-premises klien | Server waktu klien [isi] | [usulan; per deployment] |

Aturan: [minimal 4] sumber independen per host; autentikasi sumber (NTS) bila daemon dan sumber mendukung [pertimbangkan]; akses keluar hanya ke sumber yang disetujui (A.8.21).

### B.2 Ambang selisih (offset terhadap sumber)

| Parameter | Peringatan [usulan] | Kritis [usulan] | Dasar / keterangan |
| :--- | :--- | :--- | :--- |
| Offset host terhadap sumber (produksi) | > [0,5 detik] | > [1 detik] | Jauh di bawah toleransi TOTP (+/-30 detik, `auth_service.go:454-470`) dan leeway step-up (30 detik) |
| Offset yang memutus login MFA (batas fungsional) | n/a | > [10 detik] eskalasi segera | Margin pengaman terhadap jendela TOTP; angka batas teknis 30 detik berasal dari kode, bukan ambang operasional |
| Selisih antar server (VPS aplikasi vs Data Hub) | > [1 detik] | > [5 detik] | Korelasi log dan penjadwalan lintas server |
| Host tidak sinkron (tanpa sumber) | > [5 menit] | > [15 menit] | Status tidak sinkron berkepanjangan |
| Usia sinkronisasi terakhir | > [1 jam] | > [6 jam] | Sumber tidak terjangkau |
| Jumlah sumber terjangkau | < [3] | < [2] | Redundansi sumber |
| Lompatan jam (step) oleh daemon | setiap kejadian dicatat | > [1 detik] | Memengaruhi TTL dan urutan log |

---

## Lampiran C — Daftar Periksa Validasi dan Formulir (templat)

Salin satu formulir per validasi. Jangan mencantumkan kunci, kata sandi, atau alamat IP. Frekuensi: setelah pemasangan, setelah reboot tidak bersih atau perubahan jaringan, dan [bulanan].

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Daemon sinkronisasi aktif dan dikonfigurasi pada host (VPS produksi, Data Hub, [host lain]); konfigurasi diekspor tanpa rahasia. | [isi] | |
| 2 | Status sinkron "ya" dan offset terukur: [isi] detik; di bawah ambang peringatan Lampiran B. | [isi] | |
| 3 | Sumber yang dipakai sesuai daftar yang disetujui; jumlah sumber terjangkau: [isi]. | [isi] | |
| 4 | Zona waktu host sesuai standar (UTC [usulan]); zona cron tercatat: [isi]. | [isi] | |
| 5 | Jam kontainer sama dengan host (`date` pada sampel layanan: auth, audit, master, python-ai); `TZ` sesuai standar. | [isi] | |
| 6 | Selisih antar server (VPS aplikasi vs Data Hub): [isi] detik; di bawah ambang. | [isi] | |
| 7 | Postgres `SHOW timezone` dan sampel stempel `created_at` terbaru konsisten dengan UTC. | [isi] | |
| 8 | Login MFA uji berhasil dan token baru diterima (tanpa kesalahan "belum berlaku"/kedaluwarsa). | [isi] | |
| 9 | Stempel log aktivitas dan log layanan bertanda UTC dan sesuai jam host; format skrip (backup, deploy) berpenanda zona. | [isi] | |
| 10 | Metrik offset/status sinkron tersedia di Prometheus untuk semua host; dasbor menampilkan. | [isi] | |
| 11 | Alert selisih teruji bekerja pada kuartal ini (simulasi); tiket/insiden: [isi]. | [isi] | |
| 12 | Tidak ada lompatan jam tak terduga sejak validasi terakhir; kejadian reboot dan tindak lanjut: [isi]. | [isi] | |

| Field | Isi |
| :--- | :--- |
| ID validasi (usulan) | TS-[tahun]-[nomor] |
| Tanggal / jam | [isi] (UTC) |
| Pelaksana / Peninjau (peran) | [DevOps Lead] / [Lead Information Security] |
| Hasil | [Lulus / Lulus dengan catatan / Gagal] |
| Temuan dan tindak lanjut | [isi]; pemilik: [peran]; tenggat: [isi] |
| Lampiran bukti | [EV-A8.17-05-xx] |
| Tanda tangan peninjau (peran) / tanggal | [isi] |

---

## Lampiran D — Persyaratan Drift Alert (usulan)

Alert dikirim ke DevOps Lead dan Lead Information Security melalui kanal yang ditetapkan. Prasyarat teknis: Prometheus memuat `rule_files` dan Alertmanager atau pengganti (tindakan 7 A.8.21; saat ini belum ada, bagian 4 butir 2). Nama metrik di bawah mengikuti node-exporter upstream dan **harus dikonfirmasi** pada versi yang terpasang.

| No | Alert (usulan) | Kondisi [usulan] | Tingkat | Respons |
| :--- | :--- | :--- | :--- | :--- |
| 1 | Offset host melebihi ambang peringatan | offset absolut > [0,5 detik] selama [5 menit] (metrik offset `timex` node-exporter) | Peringatan | Periksa sumber dan daemon dalam [1 hari kerja] |
| 2 | Offset host melebihi ambang kritis | offset absolut > [1 detik] selama [5 menit] | Kritis | Tangani segera; periksa dampak MFA/log |
| 3 | Host tidak sinkron | status sinkron `timex` = tidak sinkron selama > [5 menit] | Peringatan; kritis > [15 menit] | Periksa daemon, firewall UDP 123, sumber |
| 4 | Selisih antar server | selisih jam antara VPS aplikasi dan Data Hub (dari `node_time_seconds` kedua host) > [1 detik] | Peringatan | Periksa kedua daemon |
| 5 | Host tidak melapor | metrik node-exporter hilang > [5 menit] | Peringatan | Periksa exporter/host (A.8.16, draf) |
| 6 | Gelombang kegagalan MFA bersamaan | lonjakan kegagalan verifikasi TOTP pada banyak pengguna [ambang: isi] | Peringatan | Periksa jam host terlebih dahulu (sinyal dari log aktivitas; A.8.16, draf) |
| 7 | Lompatan jam | perubahan mendadak offset > [1 detik] antar sampel | Peringatan | Catat; tinjau dampak TTL/log |

Persyaratan umum: setiap alert memiliki pemilik, runbook satu halaman, dan diuji [kuartalan]; hasil uji dicatat (Lampiran C butir 11). Alert tidak boleh memuat rahasia atau alamat IP pada pesan notifikasi.

---

## Lampiran E — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Platform Engineer | Solution Architect | Legal & Compliance | Direktur Utama | Data Analyst (Data Hub) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan SYN-001 | A/R | R | C | C | C | A (persetujuan) | I |
| Memilih dan menyetujui sumber waktu | A | R | C | C | C | I | I |
| Konfigurasi sinkronisasi host dan kontainer | C | A/R | R | I | I | I | R (Data Hub) |
| Validasi sinkronisasi dan formulir | A | R | C | I | I | I | R (Data Hub) |
| Drift alert dan dasbor offset | C | A | R | I | I | I | C |
| Tinjauan bulanan offset | C | A/R | C | I | I | I | C |
| Konsistensi stempel waktu aplikasi dan skrip | C | R | I | A | I | I | R (Python/Data Hub) |
| Sinkronisasi pada deployment on-premises klien | C | R | I | A | C | I | C |
| Penanganan insiden selisih waktu | A | R | C | I | I | I | C |
| Evidence ID dan review berkala | A/R | C | I | I | I | I | I |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.5.28 Pengumpulan bukti | Ada (`docs/ISO_27001/A.5.28_Pengumpulan_Bukti.md`; belum ter-commit) | Gap NTP (butir 9), tindakan 10, dan baris "sumber waktu dan selisih" pada FRP-001; dokumen ini adalah tindak lanjutnya. |
| A.8.13 Pencadangan informasi | Ada (`docs/ISO_27001/A.8.13_Pencadangan_Informasi.md`; belum ter-commit) | Cron 02:30 "server time", nama berkas backup tanpa zona, retensi `-mtime`. |
| A.8.21 Keamanan layanan jaringan | Ada (`docs/ISO_27001/A.8.21_Keamanan_Layanan_Jaringan.md`; belum ter-commit) | Prometheus tanpa `rule_files`; pemantauan TLS; akses keluar UDP 123; pin versi image. |
| A.5.24 Persiapan manajemen insiden | Ada | Selisih jam yang berdampak pada bukti sebagai insiden; verifikasi jam setelah reboot/pemulihan. |
| A.5.16 Manajemen identitas | Ada | MFA/TOTP dan sesi bergantung waktu. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Stempel UTC dan korelasi log; syslog Promtail tanpa parsing waktu. |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Drift alert dan dasbor offset. |
| A.8.9 Manajemen konfigurasi | **Draf, dokumen belum ada** | Konfigurasi sinkronisasi host dan pin versi image sebagai baseline. |
| A.8.5 Autentikasi aman | **Draf, dokumen belum ada** | Ketergantungan TOTP dan JWT pada jam. |
| A.8.24 Penggunaan kriptografi | **Draf, dokumen belum ada** | Validitas sertifikat TLS bergantung jam. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.17	Technological	Sinkronisasi waktu	Sistem menggunakan sumber waktu yang disetujui dan konsisten	NTP config, source list, drift alert, validation result	Berlaku	Token JWT dan kode MFA (TOTP) bergantung jam server; log aktivitas dan tanggal jatuh tempo ATR adalah bukti audit bertanggal; ada dua server (aplikasi dan Data Hub) yang harus konsisten; klien akan menanyakan sumber waktu dan pemantauan selisih (lihat A.8.17 bagian 1)	Sebagian	Sebagian (kode aplikasi); konfigurasi NTP, daftar sumber, drift alert, dan hasil validasi belum ada	Belum terdaftar (usulan EV-A8.17-01 s/d EV-A8.17-10)	DevOps Lead (didukung Lead Information Security)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah validasi sinkronisasi host pertama	Perlu perbaikan	Belum teruji	Tanpa konfigurasi/sumber waktu di repo (host Perlu verifikasi); tanpa drift alert (Prometheus tanpa rule_files, metrik timex tidak dipakai); Data Hub dan VPS aplikasi tanpa jaminan konsistensi; token akses tanpa leeway; TOTP sensitif jam (+/-30 detik, logika duplikat); stempel waktu backup tanpa zona; syslog Promtail tanpa parsing waktu	Ukur keadaan host (read-only); setujui SYN-001; pasang/seragamkan sinkronisasi; validasi pertama termasuk selisih lintas server; drift alert dan dasbor; seragamkan format stempel skrip; evaluasi toleransi aplikasi (12 tindakan di bagian 5)	Sedang	[usulan] 31 Desember 2026 (bukti host dan standar); 31 Januari 2027 (alert dan validasi berkala)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
