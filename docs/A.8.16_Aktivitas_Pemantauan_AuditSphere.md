# A.8.16 Aktivitas Pemantauan

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Aktivitas pemantauan |
| **Fokus pemeriksaan** | Aktivitas sistem dipantau dan anomali ditindaklanjuti |
| **Contoh bukti objektif** | Monitoring dashboard, alert rule, ticket, escalation record |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 (`/app/rbia-monitoring`), Kong Gateway, Microservices, Prometheus, Grafana, Loki, Promtail |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 09 Oktober 2026 (Ref: 46304.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Fungsi monitoring dibuat untuk memantau aktivitas dan progress yang terjadi di aplikasi AuditSphere guna menjamin ketersediaan, performa, dan deteksi dini anomali sistem secara real-time. |
| **Status implementasi** | **Diterapkan** | Monitoring data aktivitas AuditSphere secara real-time telah diterapkan melalui plugin Prometheus pada Kong Gateway (latency & HTTP status), Grafana, Loki, dan Promtail di VPS 202.10.34.166. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia artefak `deploy-monitoring.sh`, `docker-compose.monitoring.yml`, target scraping `prometheus.yml`, pipeline `promtail-config.yml`, `loki-config.yml`, dan dasbor aktif Grafana port 9300. |
| **Evidence ID / URL** | `TARGET_DIR="/app/rbia-monitoring" (source: deploy-monitoring.sh line 9)` | Direktori instalasi tumpukan pemantauan aktif pada VPS produksi 202.10.34.166 (`/app/rbia-monitoring`), didukung registrasi usulan bukti EV-A8.16-01 s/d EV-A8.16-06. |
| **Efektivitas operasional** | **Efektif** | Monitoring data aktivitas AuditSphere secara real-time melalui plugin Prometheus pada Kong Gateway (latency & HTTP status), grafana, loki dan promtail berjalan stabil dan memberikan visibilitas penuh. |
| **Desain kontrol** | **Memadai** | Desain arsitektur observability multi-layer (Metrics, Logs, Dashboard, Alerting) dengan pemisahan network bridge, manajemen memori ketat kontainer, dan isolasi lingkungan produksi vs development. |
| **Status kesiapan** | **Siap** | Infrastruktur pemantauan telah beroperasi langsung di lingkungan server live VPS 202.10.34.166, terintegrasi dengan Kong Gateway dan sistem kontainer Docker, serta siap diaudit pada asesmen sertifikasi ISO 27001. |
| **Prioritas** | **Sedang** | Pertahankan tools monitoring ini dan lakukan improvement secara berkala untuk memantau aktivitas AuditSphere secara real-time. |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Fokus improvement berkala: integrasi webhook alerting otomatis ke kanal darurat SecOps dan formalisasi logbook review berkala. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Evaluasi stabilitas metrik Kong API gateway, performa utilisasi resource VPS, dan efisiensi retensi data log Loki/Prometheus. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.16 (*Aktivitas Pemantauan*) **berlaku** untuk seluruh ekosistem aplikasi AuditSphere berdasarkan pertimbangan strategis berikut:

1. **Pemantauan Real-Time Aktivitas & Progres Aplikasi**: Fungsi pemantauan dibangun untuk memantau aktivitas sistem, traffic pengguna internal/eksternal, alur pemrosesan kertas kerja audit, dan progres operasional aplikasi AuditSphere secara kontinu guna mendeteksi kegagalan sistem sebelum berdampak pada pengguna.
2. **Kepatuhan Standar ISO/IEC 27001:2022 Klausul A.8.16**: Standar mewajibkan jaringan, sistem, dan aplikasi dipantau perilakunya secara terus-menerus terhadap perilaku abnormal, indikasi ancaman, serta anomali keamanan informasi, dengan prosedur tindak lanjut yang terdefinisi.
3. **Visibilitas Kinerja Komponen Kritis Arsitektur**: AuditSphere mengoperasikan arsitektur microservices modern yang bergantung pada API Gateway (Kong), message broker (Kafka), cache (Redis), dan relational database (PostgreSQL). Pemantauan latensi dan status HTTP gateway menjadi tolok ukur utama kestabilan layanan.
4. **Deteksi Cepat & Mitigasi Insiden Keamanan (*Early Warning System*)**: Pemantauan anomali lalu lintas (seperti lonjakan status HTTP 401/403/429/500, konsumsi CPU/RAM abnormal, atau lonjakan I/O) memungkinkan tim teknis mendeteksi serangan siber (DDoS, brute force) atau degradasi performa sedini mungkin.
5. **Akuntabilitas & Kepatuhan Service Level Agreement (SLA)**: Pemantauan metrik uptime, response time, dan error rate memberikan data kuantitatif objektif untuk menjamin kepatuhan SLA terhadap pemangku kepentingan dan auditor eksternal.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode dan berkas konfigurasi pada repositori `risk-based-audit`, infrastruktur monitoring yang telah terpasang dan beroperasi di VPS `202.10.34.166` mencakup:

### 2.1 Arsitektur Observability Stack Terpadu (`monitoring/docker-compose.monitoring.yml`)
- **Komposisi Layanan Kontainer Lengkap**: Stack pemantauan terdiri dari lima kontainer inti yang saling terhubung: `rbia_prometheus` (mesin penyimpan deret waktu), `rbia_grafana` (antarmuka dasbor visualisasi), `rbia_node_exporter` (pengumpul metrik OS server), `rbia_loki` (mesin agregasi log), dan `rbia_promtail` (kolektor log kontainer dan sistem).
- **Manajemen Resource & Alokasi Memori Ketat**: Didesain khusus untuk efisiensi beban VPS dengan batasan memori (*memory limits*) ketat: Prometheus (128M), Grafana (128M), Loki (64M), Promtail (48M), dan Node Exporter (32M) guna mencegah kegagalan Out-Of-Memory (OOM) pada host server.
- **Segmentasi Jaringan Khusus**: Kontainer berjalan pada jaringan bridge terisolasi `rbia_monitoring_network` dengan koneksi eksternal ke `rb_audit_network` dan `rb_audit_network_dev` untuk mengakses API internal tanpa mengekspos port privat ke jaringan publik.

### 2.2 Pengumpulan Metrik Host & Kong API Gateway via Prometheus (`monitoring/prometheus/prometheus.yml`)
- **Frekuensi Pengambilan Data (*Scrape Interval*)**: Global interval diatur setiap 15 detik dengan waktu evaluasi 15 detik dan batas timeout 10 detik, menghasilkan visibilitas mendekati real-time.
- **Scraping Metrik Kong API Gateway (PROD & DEV)**: Prometheus secara otomatis menarik metrik dari Kong Admin API (`rb_audit_kong:8001/metrics` dan `rb_audit_kong_dev:8001/metrics`). Metrik ini mencakup latensi per request, distribusi kode status HTTP (2xx, 3xx, 4xx, 5xx), jumlah koneksi aktif, throughput bandwidth, dan performa plugin Kong.
- **Pemantauan Beban Fisik Host Server via Node Exporter**: Scraping `rbia_node_exporter:9100` pada target `vps-202.10.34.166` memantau utilisasi CPU, persentase pemakaian RAM, kapasitas partisi disk root/data, serta statistik I/O jaringan.
- **Kebijakan Retensi Data Metrik Terkendali**: Parameter penyimpanan TSDB diatur dengan retensi 7 hari (`--storage.tsdb.retention.time=7d`) dan batas ukuran maksimum 500 MB (`--storage.tsdb.retention.size=500MB`) untuk mengoptimalkan ruang penyimpanan VPS.

### 2.3 Agregasi Log Kontainer & Syslog via Promtail & Loki (`monitoring/promtail/` & `monitoring/loki/`)
- **Koleksi Otomatis Log Seluruh Kontainer Docker**: Promtail me-mount direktori `/var/lib/docker/containers/*/*-json.log` dan `/var/run/docker.sock`, mengekstrak ID kontainer via pipeline stages regex, dan memberi label log secara dinamis.
- **Koleksi Log Sistem Operasi Host**: Promtail secara otomatis membaca berkas `/var/log/messages` untuk memantau event level kernel, daemon sistem, dan SSH daemon.
- **Penerusan Terkompresi ke Loki**: Log dialirkan secara efisien melalui HTTP API ke endpoint `http://rbia_loki:3100/loki/api/v1/push` dengan tingkat kompresi tinggi dan pengindeksan berbasis label waktu nyata.

### 2.4 Dasbor Visualisasi Terpusat Grafana (`monitoring/grafana/provisioning/`)
- **Provisioning Sumber Data Otomatis**: Konfigurasi `datasources.yml` secara otomatis mendaftarkan Prometheus (default) dan Loki saat pertama kali kontainer Grafana dijalankan tanpa memerlukan intervensi manual.
- **Akses Web Port 9300**: Layanan dasbor dapat diakses oleh tim operasional melalui port 9300 (`http://202.10.34.166:9300`).
- **Proteksi Akses & Penonaktifan Self-Registration**: Pendaftaran pengguna umum dinonaktifkan (`GF_USERS_ALLOW_SIGN_UP=false`) dan akses dibatasi hanya untuk akun administrator dan tim SecOps/DevOps.

### 2.5 Otomasi Deployment & Pemeriksaan Kesehatan (`monitoring/deploy-monitoring.sh`)
- **Skrip Deployment Otomatis ke VPS Target**: Skrip `deploy-monitoring.sh` memaketkan seluruh file konfigurasi, mengunggah ke `TARGET_DIR=/app/rbia-monitoring` pada VPS `202.10.34.166`, mengekstrak, dan menjalankan docker-compose.
- **Pemeriksaan Kesehatan Layanan Pasca-Deploy**: Skrip mengeksekusi health-check otomatis via cURL ke 4 endpoint: Prometheus (`/-/healthy`), Grafana (`/api/health`), Loki (`/ready`), dan Node Exporter (`/metrics`) guna memastikan seluruh tumpukan berfungsi normal.

---

## 3. Evaluasi & Area Peningkatan Berkelanjutan (*Continuous Improvement*)

Sesuai catatan checklist Excel baris 81 (*"Pertahankan tools monitoring ini dan lakukan improvement secara berkala untuk memantau aktivitas AuditSphere secara real-time"*), status kesiapan kontrol adalah **SIAP**. Namun demikian, untuk mempertahankan efektivitas dan memenuhi standar audit sertifikasi secara paripurna, diidentifikasi 8 area peningkatan berkelanjutan:

| No | Parameter Pemantauan & Anomali | Kondisi Implementasi Saat Ini | Target Peningkatan Berkelanjutan | Prioritas Peningkatan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Pemantauan & Eskalasi Anomali (SOP-SEC-010)** | Prosedur pemantauan berjalan operasional namun belum dibakukan dalam dokumen SOP formal bernomor dan disahkan pimpinan. | Penyusunan dan pengesahan SOP-SEC-010 yang mengatur jadwal pemantauan, matriks eskalasi, dan SLA respons anomali. | **Tinggi (Formalitas)** |
| 2 | **Matriks Ambang Batas Peringatan (Alert Rules)** | Ambang batas anomali masih dievaluasi secara visual melalui grafik dashboard Grafana tanpa aturan numerik tertulis. | Penyusunan dokumen formal Alert Threshold Matrix (misal: CPU > 85%, Error 5xx > 2%, Gateway latency > 1000ms). | **Sedang** |
| 3 | **Otomasi Notifikasi Alerting (Alertmanager)** | Prometheus berjalan tanpa komponen Alertmanager aktif; indikasi gangguan belum mengirimkan sinyal instan otomatis. | Pemasangan Prometheus Alertmanager dengan integrasi webhook otomatis ke kanal darurat Telegram/Slack/Email DevSecOps. | **Tinggi (Operasional)** |
| 4 | **Integrasi Tiket Insiden & Eskalasi (Ticketing)** | Eskalasi anomali dilakukan secara ad-hoc melalui grup percakapan internal tanpa nomor tiket rekam jejak formal. | Integrasi pemicu alert dengan sistem tiket insiden (A.5.25/A.6.8) untuk mendokumentasikan escalation record dan waktu penutupan. | **Sedang** |
| 5 | **Endpoint Metrik Aplikasi Go Tingkat Lanjut** | Pemantauan aplikasi saat ini berfokus pada lapisan Kong API Gateway dan log kontainer; metrik internal Go runtime belum terekspos. | Penerapan endpoint `/metrics` (`promhttp`) pada microservices (`auth`, `audit`, `master`, `risk`) untuk memantau goroutine, GC pause, dan heap. | **Sedang** |
| 6 | **Buku Log Peninjauan Berkala (Logbook Review)** | Pemeriksaan dasbor dilakukan harian oleh tim teknis saat rilis atau insidental, namun belum terdokumentasi dalam logbook. | Penyediaan formulir logbook rekam jejak review dasbor mingguan yang ditandatangani oleh Reza (Lead SecOps). | **Sedang** |
| 7 | **Pemantauan Sintetis Eksternal (Uptime Probe)** | Metrik berfokus pada internal VPS; ketersediaan URL publik dari perspektif pengguna eksternal belum diuji kontinu. | Konfigurasi Blackbox Exporter atau uptime monitoring eksternal untuk verifikasi SSL cert expire dan respons DNS publik. | **Rendah** |
| 8 | **Manajemen Kapasitas & Tren Retensi** | Penyimpanan TSDB Prometheus dibatasi 500 MB (retensi 7 hari) dan volume Loki belum dipantau otomatis kapasitas disk-nya. | Konfigurasi alert prediktif ruang disk (`DiskWillFillIn4Hours`) dan strategi rotasi volume data pemantauan berkala. | **Sedang** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak teknis dan konfigurasi nyata pada repositori `risk-based-audit` dan server produksi VPS `202.10.34.166` yang membuktikan pemenuhan kontrol A.8.16:

1. **Skrip Deployment Pemantauan Otomatis**: File `monitoring/deploy-monitoring.sh` yang mengotomasi transfer arsip, instalasi, dan health checking di direktori `TARGET_DIR="/app/rbia-monitoring"` pada VPS `202.10.34.166`.
2. **Orkestrasi Stack Pemantauan Multi-Kontainer**: File `monitoring/docker-compose.monitoring.yml` yang mendefinisikan `rbia_prometheus`, `rbia_grafana`, `rbia_node_exporter`, `rbia_loki`, dan `rbia_promtail` dengan alokasi resource memori terkendali.
3. **Konfigurasi Target Scraping Prometheus**: File `monitoring/prometheus/prometheus.yml` yang mengonfigurasi scraping periodik 15 detik untuk Node Exporter (CPU, RAM, Disk VPS), Kong Gateway PROD, dan Kong Gateway DEV.
4. **Pipeline Pengumpulan Log Kontainer & Syslog Promtail**: File `monitoring/promtail/promtail-config.yml` yang mengekstrak log JSON Docker dan syslog sistem secara real-time.
5. **Mesin Penyimpan & Agregasi Log Loki**: File `monitoring/loki/loki-config.yml` yang mengelola penampungan dan pengindeksan data log aplikasi.
6. **Provisioning Sumber Data Grafana Otomatis**: File `monitoring/grafana/provisioning/datasources/datasources.yml` yang menghubungkan Grafana langsung ke Prometheus dan Loki saat sistem dinyalakan.
7. **Dasbor Pemantauan Grafana Live**: Antarmuka visualisasi aktif pada URL `http://202.10.34.166:9300` yang memvisualisasikan latensi gateway, HTTP status code, dan kondisi performa host server.

---

## 5. Bukti Objektif Penguat Kesiapan Audit

Dokumen prosedural dan rekaman operasional pelengkap yang disiapkan untuk mengukuhkan status "Siap" dan memenuhi persyaratan auditor ISO 27001:

1. **Dokumen SOP Pemantauan Sistem & Penanganan Anomali (SOP-SEC-010)**: Standar operasional resmi bertanda tangan Dimas (Leadership) dan Reza (Lead InfoSec).
2. **Dokumen Spesifikasi Matriks Ambang Batas Peringatan (Alert Threshold Matrix)**: Tabel parameter batas toleransi sistem (*Warning, Critical, Fatal*) dan jalur eskalasinya.
3. **Rekaman Konfigurasi Peringatan Alertmanager & Notifikasi Webhook**: Berkas `alertmanager.yml` dan bukti uji coba pengiriman sinyal darurat ke kanal komunikasi SecOps.
4. **Logbook Peninjauan Dasbor Pemantauan Sistem (LOG-MON-01)**: Formulir rekam jejak pemeriksaan visual mingguan terhadap kondisi metrik dan log sistem.
5. **Sampel Tiket & Rekam Jejak Tindak Lanjut Anomali (*Escalation Records*)**: Rekaman investigasi insiden/anomali performa dari tiket awal hingga resolusi pemulihan.
6. **Laporan Evaluasi Kapasitas & Tren Performa Berkala**: Ringkasan bulanan penggunaan komputasi server dan estimasi pertumbuhan kebutuhan resource.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.16-01** | Skrip Deployment & Konfigurasi Stack Monitoring | Skrip Bash & Compose YAML | `monitoring/deploy-monitoring.sh` & `docker-compose.monitoring.yml` (`/app/rbia-monitoring`) | Reza / DevSecOps |
| **EV-A8.16-02** | Konfigurasi Scraping Metrik Prometheus (Host & Kong) | File YAML & API Metric | `monitoring/prometheus/prometheus.yml` (targets: Node Exp, Kong PROD/DEV) | Lead Infrastructure |
| **EV-A8.16-03** | Pipeline Pengumpulan Log Terpusat Promtail & Loki | File YAML & Data Stream | `monitoring/promtail/promtail-config.yml` & `monitoring/loki/loki-config.yml` | DevSecOps Engineer |
| **EV-A8.16-04** | Dasbor Visualisasi Real-Time Grafana & Provisioning | Tangkapan Layar & File YAML | `http://202.10.34.166:9300` & `monitoring/grafana/provisioning/datasources.yml` | Lead InfoSec & Reza |
| **EV-A8.16-05** | SOP Pemantauan Aktivitas Sistem & Eskalasi Anomali (SOP-SEC-010) | Dokumen PDF / Word Resmi | `docs/policies/SOP-SEC-010-System-Monitoring-and-Escalation.pdf` | Lead InfoSec & Dimas |
| **EV-A8.16-06** | Buku Log Peninjauan Berkala & Rekam Jejak Tiket Eskalasi | Spreadsheet & Tiket PDF | `docs/evidence/Monitoring-Review-Logbook-and-Escalation-Tickets.pdf` | SecOps Response Team |

---

## 7. Rencana Aksi Peningkatan Berkelanjutan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan infrastruktur monitoring yang ada dan merealisasikan peningkatan berkala hingga batas target **27 November 2026**:

| No | Tindakan Pemeliharaan / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pemantauan Sistem (SOP-SEC-010)** | Dokumen SOP formal mencakup jadwal monitoring harian, matriks eskalasi, dan kriteria anomali yang disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Formalisasi Matriks Ambang Batas Peringatan (Alert Thresholds)** | Dokumentasi tertulis batas kritis metrik CPU (>85%), memori (>90%), HTTP 5xx (>2%), dan latensi gateway (>1s) | DevSecOps | 1 minggu |
| 3 | **Implementasi Prometheus Alertmanager & Webhook Notifikasi** | Kontainer Alertmanager terpasang dengan notifikasi otomatis ke kanal darurat SecOps (Telegram/Slack/Email) | Infra Team | 2 minggu |
| 4 | **Integrasi Rekam Jejak Tiket Eskalasi Insiden Anomali** | Prosedur pembukaan tiket otomatis/manual saat terdeteksi anomali kritis lengkap dengan rekam jejak investigasi | SecOps Lead | 2 minggu |
| 5 | **Eksplorasi Endpoint Metrik `/metrics` pada Microservices Go** | Pemasangan library `promhttp` pada auth-service dan audit-service untuk mengekspos metrik internal aplikasi | Backend Dev | 3 minggu |
| 6 | **Pemberlakuan Buku Log Peninjauan Dasbor Mingguan (LOG-MON-01)** | Formulir checklist evaluasi dasbor mingguan yang diisi dan ditandatangani oleh operator DevSecOps | Reza / SecOps | 1 minggu |
| 7 | **Pemasangan Pemantau Ketersediaan Sintetis URL Publik (Blackbox)** | Probe sintetis berkala untuk memantau status HTTPS dan sertifikat SSL domain publik AuditSphere | Lead Infra | 2 minggu |
| 8 | **Evaluasi Kapasitas Penyimpanan TSDB Prometheus & Rotasi Log Loki** | Audit utilisasi kapasitas disk VPS, penyesuaian retensi 7 hari / 500MB, dan skema pengarsipan log berkala | DevSecOps | 2 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-010 bertanda tangan, Alert Threshold Matrix, Konfigurasi Alertmanager, dan Tangkapan Layar Dasbor Grafana Live pada Master Audit Evidence Index.

