# A.8.15 Logging

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Logging |
| **Fokus pemeriksaan** | Log aktivitas penting dibuat, dilindungi, dan disimpan |
| **Contoh bukti objektif** | Logging standard, SIEM source list, retention, sample log |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 09 Oktober 2026 (Ref: 46304.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Log dibuat untuk memantau aktivitas histori yang terjadi di aplikasi AuditSphere guna menjamin auditabilitas, deteksi anomali, investigasi insiden, dan akuntabilitas hukum. |
| **Status implementasi** | **Dalam proses** | Audit logging telah aktif pada backend Go melalui Uber Zap terstruktur, streaming Apache Kafka (*topic audit-logs*), dan basis data `sys_alert_logs`; namun agregasi terpusat dan standardisasi retensi masih dikembangkan. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia pipeline log Kafka, repositori `sys_alert_log.go`, middleware HTTP logger, dan konfigurasi Zap; namun dokumen SOP Standar Logging (SOP-SEC-009), daftar sumber SIEM, dan log review berkala belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A8.15-01 s/d EV-A8.15-06 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Pencatatan log transaksi dan anomali keamanan berjalan otomatis pada tingkat aplikasi dan broker pesan; namun peninjauan proaktif berkala dan proteksi anti-tampering log belum terformalisasi. |
| **Desain kontrol** | **Memadai** | Desain menggabungkan logging terstruktur berformat JSON (ISO 8601 UTC), pengiriman sinkron Kafka dengan *RequiredAcks WaitForAll*, dan penyimpanan terpisah tabel peringatan sistem. |
| **Status kesiapan** | **Belum siap** | Pedoman kebijakan SOP logging formal bertanda tangan Dimas, konfigurasi retensi minimal 12 bulan yang teruji, dan rekam jejak review log berkala belum memenuhi standar sertifikasi ISO 27001. |
| **Prioritas** | **Sedang** | Audit logging sudah dibuat sebagian; prioritas difokuskan untuk melanjutkan audit logging guna memantau aktivitas historis website AuditSphere secara keseluruhan. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Peninjauan integrasi sentralisasi log ke dasbor pemantauan dan evaluasi kapasitas retensi. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.15 **berlaku** untuk seluruh lingkungan arsitektur dan sistem operasional AuditSphere karena:

1. **Pemantauan Aktivitas Historis Sistem Secara Menyeluruh**: Log transaksi dan sistemik merupakan instrumen utama untuk memantau rekam jejak historis seluruh aktivitas pengguna, perubahan data kertas kerja audit, proses *approval* laporan eksekutif, serta konfigurasi perizinan.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.8.15**: Standar secara eksplisit mewajibkan organisasi menghasilkan, menyimpan, melindungi dari pemalsuan, dan menganalisis log yang merekam aktivitas pengguna, pengecualian (*exceptions*), kegagalan sistem, dan kejadian keamanan lainnya.
3. **Dasar Pembuktian Forensik Digital & Non-Repudiasi**: Dalam hal terjadi sengketa hasil audit, tuduhan kecurangan (*fraud*), atau insiden kebocoran informasi, catatan log yang tersinkronisasi waktu (NTP) dan berformat terstruktur menjadi bukti digital yang sah di hadapan hukum (UU ITE & UU PDP).
4. **Deteksi Dini Kejadian & Anomali Keamanan (*Early Warning System*)**: Log akses HTTP, kegagalan autentikasi berulang, dan peringatan eskalasi peran menjadi bahan baku utama bagi deteksi intrusi, penanganan insiden (A.5.25 & A.5.26), serta penegakan disipliner (A.6.4).
5. **Pemberian Akuntabilitas Individual (*Traceability*)**: Setiap tindakan pengguna pada antarmuka web, API gateway Kong, dan microservices backend dapat diatribusikan secara pasti ke identitas subjek tertentu (UserID, Username, IP Address, User-Agent).

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode repositori `risk-based-audit`, infrastruktur logging yang telah diimplementasikan mencakup:

### 2.1 Mesin Logging Terstruktur Berbasis Uber Zap (`pkg/logger/logger.go`)
- **Format JSON & Waktu ISO 8601**: Setiap microservice (`auth-service`, `audit-service`, `master-service`, `risk-service`) menggunakan pustaka `go.uber.org/zap` dengan konfigurasi encoder `ISO8601TimeEncoder`, memastikan stempel waktu seragam dengan presisi tinggi dan terstruktur.
- **Pemisahan Level Log Dinamis**: Logger mendukung tingkatan debug, info, warn, error, dan fatal dengan jalur keluaran stdout/stderr yang siap dikonsumsi oleh kontainer Docker dan agregator log.

### 2.2 Middleware Pencatatan Permintaan HTTP (`pkg/middleware/logger.go`)
- **Perekaman Metadata Permintaan Lengkap**: Interseptor Gin pada `middleware/logger.go` secara otomatis merekam status kode HTTP, metode request, path URL, query string, alamat IP klien (`ClientIP`), User-Agent, dan durasi latensi eksekusi.
- **Audit Akses Gateway Kong**: Kong API Gateway mengintersepsi seluruh traffic masuk dan mencatat log akses terpusat sebelum diteruskan ke layanan microservices.

### 2.3 Streaming Audit Event Asinkron via Apache Kafka (*topic audit-logs*)
- **Struktur Pesan Log Standar (`LogMessage`)**: Modul `pkg/kafka/producer.go` mendefinisikan skema `LogMessage` mencakup Timestamp, Level, Service, Message, Fields, TraceID, SpanID, dan Environment.
- **Jaminan Pengiriman Andal (`WaitForAll`)**: Konfigurasi `sarama.WaitForAll` pada producer Kafka menjamin pesan audit berhasil direplikasi ke seluruh broker Kafka aktif sebelum dianggap sukses (*acks = all*).

### 2.4 Tabel Peringatan Keamanan Sistem Terstruktur (`sys_alert_logs`)
- **Pencatatan Event Keamanan Otomatis**: Model `backend/auth-service/models/sys_alert_log.go` mencatat kejadian penting seperti login gagal berulang (`SysAlertFailedAttemptLog`), penguncian akun (`SysAlertLockedOut`), pelanggaran rate limit, dan pencabutan hak akses peran (`SysRoleRevoked`).
- **Repositori Kueri Khusus**: File `repositories/sys_alert_log_repository.go` menyediakan method `FindByAction`, `FindByUserID`, dan `FindByIPAddress` untuk mempermudah investigasi.

### 2.5 Jejak Audit Kertas Kerja & Dokumen Hasil Audit
- **Snapshot Histori Transaksi**: Layanan `audit-service` mencatat riwayat perubahan status dokumen, pembuatan temuan audit, dan persetujuan laporan eksekutif lengkap dengan identitas auditor penanggung jawab.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.8.15 menunjukkan 8 kesenjangan yang harus ditangani:

| No | Parameter Kontrol Logging | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Kebijakan & SOP Standar Logging (SOP-SEC-009)** | Tersedia dokumen SOP resmi yang mengatur format log, event wajib catat, dan siklus peninjauan. | Infrastruktur kode log telah aktif; namun dokumen SOP formal tertulis belum disahkan pimpinan. | **Gap Kebijakan / SOP** |
| 2 | **Daftar Inventaris Sumber Log Sistem (SIEM Sources)** | Katalog formal yang menginventarisasi seluruh sumber log (Kong, microservices, DB, Kafka, OS). | Sumber log tersebar di beberapa modul tanpa dokumen inventaris resmi terpusat. | **Gap Inventaris Sumber** |
| 3 | **Kebijakan Retensi & Pengarsipan Log 12 Bulan** | Log disimpan aktif minimal 90 hari dan diarsipkan (*cold storage*) hingga minimal 365 hari. | Retensi log Kafka dan database belum dikonfigurasi secara formal untuk jangka 12 bulan. | **Gap Kebijakan Retensi** |
| 4 | **Perlindungan Anti-Tampering & Hak Akses Log** | Log dilindungi dari modifikasi atau penghapusan oleh administrator menggunakan hak akses read-only. | Administrator database masih memiliki hak teknis menghapus baris `sys_alert_logs` (belum WORM). | **Gap Proteksi Integritas** |
| 5 | **Dasbor Agregasi Log & SIEM Terpusat** | Sistem visualisasi log terpusat (misal: Grafana Loki/OpenSearch) untuk kueri real-time lintas servis. | Log saat ini dialirkan ke stdout dan Kafka; dasbor agregasi terpadu belum sepenuhnya terintegrasi. | **Gap Sentralisasi SIEM** |
| 6 | **Aturan Deteksi Anomali & Alerting Otomatis** | Pemicu alert otomatis ke tim SecOps saat terdeteksi lonjakan error 500, brute force, atau akses abnormal. | Event alert tersimpan di DB `sys_alert_logs` tetapi belum memicu notifikasi instan (Slack/email). | **Gap Otomasi Alert** |
| 7 | **Logbook Peninjauan Log Berkala (Log Review Records)** | Rekam jejak tertulis pelaksanaan review log berkala mingguan/bulanan oleh tim keamanan. | Peninjauan log masih bersifat insidental/reaktif saat terjadi gangguan, belum terjadwal proaktif. | **Gap Prosedur Review** |
| 8 | **Sinkronisasi Waktu Server Tepercaya (NTP Sync)** | Seluruh server dan container tersinkronisasi ke server NTP stratum-1/2 resmi dengan bukti verifikasi. | Server menggunakan waktu sistem default tanpa dokumen pengujian sinkronisasi NTP berkala. | **Gap Sinkronisasi Waktu** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan implementasi teknis di dalam repositori yang mendukung kontrol logging meliputi:

1. **Konfigurasi Uber Zap Logger**: File `backend/auth-service/pkg/logger/logger.go` dengan encoder ISO8601 dan keluaran JSON terstruktur.
2. **Middleware HTTP Request Logging**: File `backend/auth-service/pkg/middleware/logger.go` yang merekam IP, metode, path, latency, dan user-agent.
3. **Pipeline Streaming Kafka**: Berkas `backend/auth-service/pkg/kafka/producer.go` dengan skema `LogMessage` dan konfigurasi `WaitForAll`.
4. **Basis Data Peringatan Keamanan**: File `backend/auth-service/models/sys_alert_log.go` dan repositori `sys_alert_log_repository.go`.
5. **Pencatatan Event Domain Audit**: Berkas `backend/audit-service/pkg/kafka/events.go` untuk publikasi aktivitas kertas kerja audit.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif formal yang harus disiapkan untuk memenuhi kriteria kelulusan audit sertifikasi ISO/IEC 27001:2022:

1. **SOP Standar Logging & Audit Trail (SOP-SEC-009)**: Dokumen resmi bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) yang mengatur standar logging.
2. **Katalog Inventaris Sumber Log Sistem (SIEM Source Inventory)**: Daftar seluruh titik penghasil log pada arsitektur AuditSphere.
3. **Dokumen Konfigurasi Retensi Log Minimal 365 Hari**: Bukti konfigurasi masa simpan log pada Kafka broker dan skema cold storage archiving.
4. **Sampel Berkas Log Aktivitas & Keamanan Terstruktur**: Tangkapan layar dan ekspor sampel log audit yang membuktikan rekaman aktivitas lengkap.
5. **Bukti Pengaturan Hak Akses Read-Only & Anti-Tampering Log**: Skrip database trigger pencegah modifikasi/penghapusan pada tabel `sys_alert_logs`.
6. **Buku Log & Berita Acara Peninjauan Log Berkala**: Formulir rekam jejak pemeriksaan log mingguan oleh tim DevSecOps.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif yang wajib diregistrasikan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.15-01** | SOP Standar Logging & Audit Trail (SOP-SEC-009) | Dokumen PDF / Word | `docs/policies/SOP-SEC-009-Logging-Standard.pdf` | Lead InfoSec & Reza |
| **EV-A8.15-02** | Katalog Inventaris Sumber Log Sistem (SIEM Sources) | Spreadsheet / PDF | `docs/architecture/SIEM-Log-Sources-Catalog.pdf` | Lead Infrastructure |
| **EV-A8.15-03** | Konfigurasi Retensi Log 365 Hari & Script Archiving | File Konfigurasi & Script | `docs/evidence/Log-Retention-Configuration.pdf` | DevSecOps Engineer |
| **EV-A8.15-04** | Sampel Ekspor Log Terstruktur (JSON & `sys_alert_logs`) | File Ekspor JSON / CSV | `docs/evidence/Sample-Audit-Trail-Export.json` | Lead Backend Dev |
| **EV-A8.15-05** | Konfigurasi Proteksi Anti-Tampering Tabel Log | Skrip SQL Trigger & ACL | `docs/evidence/Append-Only-Log-Protection.sql` | Database Administrator |
| **EV-A8.15-06** | Buku Log Peninjauan Aktivitas Sistem Berkala | Spreadsheet Terproteksi | SecOps Management Portal / Log Review Records | SecOps Team |

---

## 7. Rencana Aksi Perbaikan (Action Plan)

Rencana aksi terstruktur guna melanjutkan audit logging dan menutup kesenjangan kontrol A.8.15:

| No | Tindakan Remediasi / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Standar Logging (SOP-SEC-009)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) | Reza / Dimas | 2 minggu |
| 2 | **Pemetaan & Inventarisasi Sumber Log Sistem Terpadu** | Katalog lengkap seluruh endpoint penghasil log (Gateway, Services, Database, Kafka) | Lead Infra | 1 minggu |
| 3 | **Standardisasi Konfigurasi Retensi Log Minimal 365 Hari** | Konfigurasi retensi topik Kafka 90 hari aktif dan arsip cold storage S3/NAS hingga 1 tahun | DevSecOps | 2 minggu |
| 4 | **Penerapan Proteksi Anti-Tampering Tabel sys_alert_logs** | Pemasangan database trigger PostgreSQL untuk mencegah modifikasi atau penghapusan log | Backend Dev | 1 minggu |
| 5 | **Implementasi Agregator Log Terpusat (Grafana Loki / OpenSearch)** | Dasbor sentralisasi log real-time untuk pemantauan traffic dan anomali lintas servis | Infra Team | 2 minggu |
| 6 | **Konfigurasi Rule Alerting Otomatis ke Kanal DevSecOps** | Pemicu alert otomatis saat terjadi lonjakan error rate limit atau kegagalan login massal | SecOps Lead | 2 minggu |
| 7 | **Pemberlakuan Jadwal Review Log Berkala Mingguan** | Prosedur inspeksi proaktif mingguan atas ringkasan log aktivitas dan anomali sistem | Reza / SecOps | 1 minggu |
| 8 | **Verifikasi Sinkronisasi Waktu Server ke NTP Nasional** | Dokumentasi kalibrasi sinkronisasi waktu seluruh node server ke server NTP tepercaya | Lead Infra | 1 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP Standar Logging bertanda tangan, Katalog Sumber Log, Konfigurasi Retensi 365 Hari, dan Sampel Log Terstruktur pada Master Index.

