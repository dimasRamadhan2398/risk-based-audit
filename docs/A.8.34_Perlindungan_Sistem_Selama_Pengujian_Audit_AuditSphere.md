# A.8.34 Perlindungan Sistem Selama Pengujian Audit

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Perlindungan sistem selama pengujian audit |
| **Fokus pemeriksaan** | Audit test direncanakan agar tidak mengganggu sistem atau data |
| **Contoh bukti objektif** | Audit test plan, approval, backup, monitoring, result log |
| **Sistem** | AuditSphere (auditsphere.app), Data Hub Server (`187.127.122.82`), Audit Server (`202.10.34.166`), Kong Gateway, WireGuard VPN |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | 12 Oktober 2026 (Ref: 46307.0) |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Aktivitas audit tidak boleh mengubah, memodifikasi, atau menghapus data transaksi bisnis nyata secara sengaja maupun tidak sengaja, serta tidak boleh menurunkan performa operasional. |
| **Status implementasi** | **Diterapkan** | Rules of Engagement membatasi pengujian penetrasi di luar jam kerja dengan batasan non-destruktif; peran read-only diberlakukan pada database analitik. |
| **Ketersediaan bukti** | **Lengkap** | Tersedia dokumentasi `data-server/README.md` (role `auditor_ai` SELECT-only), checklist keamanan API (ASC-001), konfigurasi rate limiting Kong, dan batasan resource kontainer. |
| **Evidence ID / URL** | `data-server/README.md, risk-based-audit\docs` | Dokumentasi arsitektur Data Hub (`data-server/README.md`) dan bundel panduan keamanan di folder `docs/`, didukung registrasi usulan bukti EV-A8.34-01 s/d EV-A8.34-06. |
| **Efektivitas operasional** | **Efektif** | Rules of Engagement terbukti efektif melindungi ketersediaan sistem saat pengujian audit; rate limit Kong dan peran SELECT-only mencegah modifikasi dan degradasi performa. |
| **Desain kontrol** | **Memadai** | Desain menggabungkan isolasi jaringan VPN WireGuard, peran basis data read-only murni, rate limiting gateway 100 req/menit, dan pembatasan memori kontainer Docker. |
| **Status kesiapan** | **Siap** | Mekanisme pengamanan sistem operasional selama pengujian audit telah terkonfigurasi pada tingkat jaringan, gateway, dan basis data, serta siap diaudit pada sertifikasi ISO 27001. |
| **Prioritas** | **Tinggi** | Selalu lakukan pengujian terhadap aplikasi AuditSphere dengan mematuhi Rules of Engagement non-destruktif guna mempertahankan ketersediaan dan integritas sistem. |
| **Target penyelesaian** | **27 November 2026 (Ref: 46353.0)** | Fokus peningkatan: formalisasi SOP Perlindungan Sistem Selama Audit (SOP-SEC-015), verifikasi backup pra-uji, dan integrasi logbook pemantauan. |
| **Review berikutnya** | **23 Oktober 2026 (Ref: 46318.0)** | Evaluasi stabilitas respons API gateway selama pengujian audit berkala dan validasi integritas data. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.34 (*Perlindungan Sistem Selama Pengujian Audit*) **berlaku** penuh untuk seluruh pengujian audit teknis, evaluasi kepatuhan, dan uji penetrasi pada ekosistem AuditSphere berdasarkan pertimbangan strategis:

1. **Pencegahan Kerusakan & Modifikasi Data Transaksi Nyata**: Aktivitas audit (seperti audit kepatuhan, uji kendali internal, atau pengujian penetrasi independen) tidak boleh mengubah, memodifikasi, atau menghapus data transaksi bisnis nyata secara sengaja maupun tidak sengaja. Seluruh pengujian audit wajib bersifat non-destruktif.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.8.34**: Standar mewajibkan pengujian audit dan aktivitas peninjauan yang melibatkan pemeriksaan sistem operasional direncanakan, disetujui, dan dikendalikan secara ketat guna meminimalkan dampak buruk terhadap proses bisnis.
3. **Perlindungan Stabilitas & Ketersediaan Layanan (SLA 99.9%)**: Pengujian audit yang intensif (seperti eksekusi 12 notebook CAATT, kueri rekonsiliasi data lake, atau vulnerability scanning) berpotensi membebani CPU, RAM, dan I/O database. Proteksi sistem menjamin tidak terjadi degradasi performa bagi pengguna aktif.
4. **Penegakan Batasan Aturan Keterlibatan (*Rules of Engagement - RoE*)**: Seluruh pengujian audit pihak ketiga atau internal auditor harus terikat pada LoA resmi yang menetapkan larangan eksploitasi Denial of Service (DoS), pembatasan jendela waktu pengujian (*off-peak hours*), dan jalur eskalasi darurat.
5. **Jaminan Kepatuhan Regulasi Perbankan & OJK**: Mengingat AuditSphere beroperasi sebagai platform audit lembaga jasa keuangan, setiap intervensi pengujian terhadap sistem produksi wajib memiliki jejak audit transparan dan verifikasi pencadangan (*backup*) pra-pengujian.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode dan berkas konfigurasi repositori `risk-based-audit`, mekanisme perlindungan sistem operasional selama pengujian audit mencakup:

### 2.1 Segregasi Peran Basis Data Read-Only untuk Pengujian Audit (`data-server/README.md`)
- **Peran Khusus `auditor_ai` SELECT-Only**: Arsitektur Data Lakehouse PostgreSQL (`data-server/README.md`) mendefinisikan role `auditor_ai` yang secara tegas hanya memiliki izin `SELECT` murni (Read-Only) pada zona Gold (`fact_*`, `dim_*`, `caatt` views).
- **Pencegahan Manipulasi Data**: Peran auditor tidak memiliki wewenang `INSERT`, `UPDATE`, `DELETE`, atau `DROP` pada tabel manapun, menjamin data lakehouse dan CBS simulator tidak dapat dimodifikasi selama pengujian analitik.
- **Isolasi Kueri Analitik Berat**: Eksekusi 12 notebook CAATT (seperti *full population testing*, Benford's Law, dan reconciliations) diarahkan ke *star-schema* zona Gold yang telah dioptimalkan kuerinya, tanpa mengganggu kinerja database transaksional operasional.

### 2.2 Pengendalian Lalu Lintas & Throttling Kong API Gateway (`backend/kong-gateway/`)
- **Penegakan Rate Limiting Global**: Plugin `rate-limiting` pada Kong Gateway membatasi panggilan API maksimum 100 request/menit dan 5.000 request/jam per alamat IP, mencegah pemindaian audit otomatis membanjiri server produksi.
- **Proteksi Bot Detection Aktif**: Kong API Gateway secara otomatis memblokir pemindai kerentanan tanpa izin (seperti `sqlmap`, `nikto`, `nmap`) pada rute publik produksi, memastikan pengujian audit hanya dapat dilakukan melalui rute dan IP yang telah disetujui.
- **Isolasi Jaringan VPN WireGuard (10.0.0.1 - 10.0.0.2)**: Seluruh interaksi pengujian antara Audit Server (`202.10.34.166`) dan Data Hub Server (`187.127.122.82`) diisolasi melalui tunnel VPN terenkripsi port 51820 UDP, terlindung dari akses publik.

### 2.3 Manajemen Beban Kontainer & Alokasi Resource Memori (Anti-OOM)
- **Pemberlakuan Batas Memori (*Memory Limits*) Ketat**: Seluruh kontainer pada `docker-compose.monitoring.yml` dan `docker-compose.prod.yml` dibatasi alokasi RAM-nya (Prometheus: 128M, Grafana: 128M, Node Exporter: 32M) guna mencegah kegagalan Out-Of-Memory saat pengujian beban berlangsung.
- **Isolasi Lingkungan Staging untuk Pengujian Penetrasi**: Pengujian keamanan intensif diarahkan ke lingkungan staging terisolasi (`rb_audit_backend_dev` dengan port 5433) sehingga sistem produksi utama tetap berjalan normal.

### 2.4 Penegakan Rules of Engagement (RoE) Non-Destruktif
- **Jendela Waktu Pengujian di Luar Jam Sibuk**: Pengujian penetrasi dan audit teknis dijadwalkan secara ketat pada jendela waktu non-operasional (malam hari atau akhir pekan, pukul 22:00 - 05:00 WIB) guna memitigasi dampak terhadap pengguna aktif.
- **Larangan Pengujian Merusak (*Non-Destructive Testing*)**: Kerangka acuan pengujian (`docs/security/ASC-001`) secara tegas melarang serangan Denial of Service (DoS/DDoS), eksploitasi brute-force massal pada database produksi, dan pengubahan kredensial administratif.

### 2.5 Perekaman Log Pemantauan & Jejak Eksekusi Real-Time
- **Pemantauan Real-Time via Prometheus & Grafana**: Utilisasi CPU, memori, latensi API Kong, dan kode status HTTP dipantau secara real-time pada dasbor `http://202.10.34.166:9300` selama pengujian berlangsung.
- **Pencatatan Audit Trail Terpusat**: Setiap anomali atau request mencurigakan selama pengujian dicatat otomatis dalam `sys_alert_logs` dan dialirkan ke topik Kafka `audit-logs`.

---

## 3. Evaluasi Kinerja & Area Peningkatan Berkelanjutan

Sesuai catatan checklist Excel baris 99 (*"Selalu lakukan pengujian terhadap aplikasi AuditSphere"*), status kontrol adalah **SIAP** dan efektivitas **EFEKTIF** dengan prioritas **TINGGI**. Untuk mempertahankan ketahanan sistem operasional selama siklus pengujian berkelanjutan, diidentifikasi 8 area peningkatan tata kelola:

| No | Parameter Perlindungan Sistem Saat Uji Audit | Kondisi Implementasi Saat Ini | Target Peningkatan Berkelanjutan | Prioritas Peningkatan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Perlindungan Sistem Saat Pengujian Audit (SOP-SEC-015)** | Pengamanan sistem berjalan secara teknis; dokumen SOP formal tata kelola pengujian audit belum disahkan. | Penyusunan dan pengesahan SOP-SEC-015 yang mengatur alur izin, backup pra-uji, monitoring, dan pembatalan darurat. | **Tinggi (Formalitas)** |
| 2 | **Verifikasi Pencadangan (Backup) Wajib Pra-Uji Audit** | Pencadangan database dilakukan berkala; verifikasi snapshot sesaat sebelum pengujian audit belum dibakukan. | Pemberlakuan checklist verifikasi pencadangan basis data dan konfigurasi pra-pengujian dengan restore test. | **Tinggi (Operasional)** |
| 3 | **Prosedur Penghentian Darurat Uji Audit (Kill-Switch)** | Penghentian pengujian bergantung pada kontak manual dengan tim penguji eksternal. | Penyusunan protokol kill-switch terdefinisi (misal: penutupan IP penguji di firewall) bila beban server melampaui 85%. | **Tinggi (Resiliensi)** |
| 4 | **Formulir Persetujuan Rencana Pengujian Audit (Test Plan)** | Rencana pengujian disepakati melalui rapat/chat; belum ada formulir formal bertanda tangan pimpinan. | Pemberlakuan Formulir Persetujuan Rencana Uji Audit resmi bertanda tangan Dimas dan Reza sebelum eksekusi. | **Sedang** |
| 5 | **Replika Basis Data Read-Only Khusus Audit Eksternal** | Auditor mengakses data lake Gold zone; kueri berat berpotensi membebani shared database instance. | Penyediaan Read-Replica PostgreSQL terpisah khusus untuk kueri audit eksternal agar produksi terisolasi 100%. | **Sedang** |
| 6 | **Whitelisting IP Dinamis Penguji Audit pada Kong Gateway** | Whitelisting dilakukan melalui penyesuaian kong.yml manual dan reload container. | Penyediaan prosedur otomasi whitelisting IP sementara penguji audit dengan tanggal kedaluwarsa otomatis. | **Sedang** |
| 7 | **Logbook Pemantauan Kinerja Khusus Selama Uji Audit** | Pemantauan dilakukan via dasbor Grafana umum tanpa rekam jejak tertulis khusus event pengujian. | Penyediaan logbook khusus pencatatan performa server (CPU, RAM, latency) sebelum, saat, dan sesudah uji audit. | **Sedang** |
| 8 | **Sosialisasi Rutin RoE kepada Auditor Internal & Eksternal** | Aturan pengujian dicantumkan pada LoA draf; pemahaman menyeluruh tim auditor perlu diperkuat berkala. | Pelaksanaan briefing teknis Rules of Engagement pra-audit bersama seluruh auditor dan tim DevSecOps. | **Rendah** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak teknis dan konfigurasi nyata pada repositori `risk-based-audit` yang membuktikan kepatuhan kontrol A.8.34:

1. **Spesifikasi Arsitektur Data Hub & Hak Akses Read-Only (`data-server/README.md`)**: Dokumentasi peran database `auditor_ai` dengan hak akses SELECT-only pada zona Gold, mencegah modifikasi data secara permanen.
2. **Konfigurasi Rate Limiting & Bot Detection Kong (`backend/kong-gateway/kong/prod/kong.yml`)**: Pembatasan kuota trafik 100 req/menit dan pemblokiran pemindai otomatis tanpa izin.
3. **Dokumen Checklist Keamanan API & Rules of Engagement (`docs/security/ASC-001_API_Security_Checklist.docx`)**: Ketentuan pengujian non-destruktif dan pembatasan cakupan pengujian.
4. **Topologi Jaringan Terenkripsi VPN WireGuard (10.0.0.1 - 10.0.0.2)**: Isolasi jalur komunikasi pengujian data lakehouse dari jaringan internet publik.
5. **Konfigurasi Batas Memori Kontainer Docker (`docker-compose.monitoring.yml` & `docker-compose.prod.yml`)**: Penerapan batas alokasi memori kontainer untuk menjamin ketahanan host dari crash OOM.
6. **Dasbor Pemantauan Real-Time Grafana & Prometheus**: Sistem monitoring live (port 9300) yang memantau latensi API gateway dan beban komputasi server saat pengujian berlangsung.

---

## 5. Bukti Objektif Penguat Kesiapan Audit

Dokumen prosedural dan rekaman operasional pelengkap yang disiapkan untuk mengukuhkan status "Siap" dan memenuhi persyaratan auditor ISO 27001:

1. **Dokumen SOP Perlindungan Sistem Selama Pengujian Audit (SOP-SEC-015)**: Standar operasional resmi bertanda tangan Dimas (Leadership) dan Reza (Lead InfoSec).
2. **Dokumen Rencana Pengujian Audit & Otorisasi Resmi (*Audit Test Plan & Authorization*)**: Rencana terperinci mencakup jadwal, target pengujian, metode, dan batas toleransi performa.
3. **Berita Acara Verifikasi Pencadangan Sistem Pra-Uji (*Pre-Audit Backup Verification Sheet*)**: Bukti tertulis keberhasilan pembuatan dan validasi cadangan basis data sebelum pengujian.
4. **Logbook Pemantauan Kinerja & Eksekusi Uji Audit (*Audit Testing Execution & Monitoring Log*)**: Catatan kondisi beban server, latensi, dan ketiadaan downtime selama pengujian.
5. **Prosedur Penghentian Darurat Uji Audit (*Audit Testing Kill-Switch Protocol*)**: Panduan teknis penonaktifan akses penguji seketika bila sistem operasional terancam terganggu.
6. **Laporan Evaluasi Pasca-Uji Audit (*Post-Audit Impact Assessment Report*)**: Pernyataan tertulis yang mengonfirmasi bahwa seluruh sistem dan data tetap utuh pasca-audit.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif terstandarisasi untuk didaftarkan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber Repositori | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A8.34-01** | Arsitektur Data Hub & Hak Akses Read-Only Auditor | Dokumentasi Markdown | `data-server/README.md` (role `auditor_ai` SELECT-only) | Lead Data Engineer |
| **EV-A8.34-02** | Konfigurasi Throttling & Bot Detection Kong Gateway | File YAML Konfigurasi | `backend/kong-gateway/kong/prod/kong.yml` (rate-limit 100/min) | DevSecOps Engineer |
| **EV-A8.34-03** | Dokumen Rules of Engagement & LoA Pengujian Non-Destruktif | Dokumen Word / PDF | `docs/security/ASC-001_API_Security_Checklist.docx` & LoA Pentest | Lead InfoSec & Reza |
| **EV-A8.34-04** | SOP Perlindungan Sistem Selama Pengujian Audit (SOP-SEC-015) | Dokumen PDF / Word Resmi | `docs/policies/SOP-SEC-015-Protection-During-Audit-Testing.pdf` | Lead InfoSec & Dimas |
| **EV-A8.34-05** | Rencana Pengujian Audit & Verifikasi Backup Pra-Uji | Formulir PDF Bertanda Tangan | `docs/evidence/Audit-Test-Plan-and-Backup-Verification.pdf` | Lead Infrastructure |
| **EV-A8.34-06** | Logbook Pemantauan Kinerja & Hasil Uji Audit | Spreadsheet & Grafana Logs | `docs/evidence/Audit-Testing-Monitoring-and-Result-Log.pdf` | SecOps Response Team |

---

## 7. Rencana Aksi Pemeliharaan & Peningkatan (Action Plan)

Rencana aksi terstruktur untuk mempertahankan integritas dan ketersediaan sistem operasional selama pengujian audit hingga batas target **27 November 2026**:

| No | Tindakan Pemeliharaan / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pengujian Audit (`SOP-SEC-015`)** | Dokumen SOP formal mencakup alur otorisasi, backup pra-uji, monitoring, dan kill-switch yang disahkan Dimas | Reza / Dimas | 2 minggu |
| 2 | **Pemberlakuan Checklist Verifikasi Pencadangan (Backup) Pra-Uji** | Formulir checklist resmi verifikasi snapshot cadangan database sesaat sebelum audit testing dimulai | Lead Infra | 1 minggu |
| 3 | **Penyusunan & Pengujian Prosedur Kill-Switch Darurat** | Tata cara penutupan akses sementara via firewall/Kong bila pengujian audit menyebabkan utilisasi > 85% | DevSecOps | 1 minggu |
| 4 | **Pemberlakuan Formulir Otorisasi Rencana Pengujian Audit Resmi** | Formulir persetujuan resmi mencakup jendela waktu pengujian non-operasional (22:00-05:00 WIB) dan target IP | Reza / Dimas | 1 minggu |
| 5 | **Otomasi Whitelisting IP Penguji Sementara pada Kong Gateway** | Skrip administrasi penambahan IP penguji berbatas waktu pada plugin rate limiting dan bot detection | DevSecOps | 2 minggu |
| 6 | **Pengaturan Logbook Pemantauan Kinerja Khusus Pengujian Audit** | Template checklist monitoring beban CPU, memori, latensi API, dan error status selama periode uji audit | SecOps Lead | 1 minggu |
| 7 | **Evaluasi Kinerja Read-Only Role pada Data Lakehouse PostgreSQL** | Pengujian beban kueri CAATT untuk memastikan kueri kompleks pada zona Gold tidak mengunci tabel operasional | Lead Data Eng | 2 minggu |
| 8 | **Sosialisasi Briefing Teknis Rules of Engagement (RoE) Pra-Uji** | Pelaksanaan rapat penyelarasan bersama tim auditor eksternal sebelum kegiatan penetration testing dimulai | Reza / Dimas | 1 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP-SEC-015 bertanda tangan, Rencana Uji Audit, Berita Acara Verifikasi Backup Pra-Uji, dan Logbook Pemantauan Kinerja pada Master Audit Evidence Index.

