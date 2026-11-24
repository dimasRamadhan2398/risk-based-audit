# A.6.8 Pelaporan Kejadian Keamanan

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | People |
| **Judul kontrol** | Pelaporan kejadian keamanan |
| **Fokus pemeriksaan** | Personel mengetahui kanal dan kewajiban melaporkan event keamanan |
| **Contoh bukti objektif** | Prosedur pelaporan, kanal hotline, awareness, ticket sample |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Personel (Auditor, CAE, Developer, Sysadmin, dan pengguna entitas audit) memegang peran vital sebagai garis pertahanan terdepan (*human sensor*); kewajiban dan mekanisme pelaporan yang cepat atas anomali keamanan sangat penting guna mencegah eksfiltrasi data audit dan gangguan operasional. |
| **Status implementasi** | **Dalam proses** | Infrastruktur penangkapan log anomali mesin dan audit trail (`sys_alert_logs`, Kafka `security.suspicious`, *trusted devices*) telah berjalan, namun SOP formal pelaporan oleh personel (*human-initiated reporting*) dan kanal hotline/helpdesk keamanan terpusat belum lengkap. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia log sistemik `sys_alert_logs`, streaming event Kafka `security.rate_limit`, dan model log error; namun SOP Pelaporan Kejadian Keamanan Informasi (SOP-SEC-004), tiket hotline/helpdesk, dan bukti sosialisasi awareness kewajiban pelaporan belum tersedia lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A6.8-01 s/d EV-A6.8-06 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Pelaporan berbasis telemetri sistem berjalan otomatis secara real-time; namun pelaporan manual oleh pengguna masih bersifat situasional tanpa SLA respon awal terstandarisasi dan belum terhubung ke sistem tiket khusus insiden keamanan. |
| **Desain kontrol** | **Perlu perbaikan** | Belum ada alur pelaporan formal yang menjamin budaya bebas sanksi (*no-blame culture*) bagi pelapor yang beritikad baik, belum ada integrasi form laporan mandiri di aplikasi, serta belum tersedia kontak darurat hotline 24/7. |
| **Status kesiapan** | **Belum siap** | Pedoman kebijakan SOP pelaporan formal yang disahkan pimpinan (Dimas), kanal hotline terdedikasi, sampel tiket penanganan laporan pengguna, dan rekaman sosialisasi kesadaran pelaporan belum mencukupi standar audit sertifikasi ISO 27001. |
| **Prioritas** | **Tinggi** | Keterlambatan pelaporan insiden oleh personel dapat memperbesar radius dampak serangan siber (misalnya phishing kredensial, kebocoran kertas kerja audit, infeksi malware); respon cepat sangat bergantung pada deteksi dini oleh manusia. |
| **Review berikutnya** | **[isi]** | Usulan: 6 bulan setelah peresmian kanal hotline dan pelaksanaan simulasi/drill pelaporan insiden keamanan kepada seluruh staf. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.6.8 **berlaku** untuk seluruh lingkungan kerja dan ekosistem aplikasi AuditSphere karena:

1. **Peran Kritis Personel sebagai Detektor Garis Depan (*Human Firewall*)**: Sistem keamanan teknis otomatis (firewall, SIEM, WAF) tidak dapat mendeteksi seluruh bentuk ancaman secara sempurna, khususnya serangan rekayasa sosial (*social engineering*), *spear phishing*, penyusupan fisik, kehilangan laptop dinas, atau kebocoran kredensial auditor. Deteksi oleh personel sangat menentukan keberhasilan respon insiden.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.6.8**: Standar secara eksplisit mewajibkan organisasi menyediakan mekanisme yang mudah diakses bagi personel untuk melaporkan kejadian keamanan informasi yang diamati atau dicurigai secara tepat waktu melalui saluran yang ditetapkan.
3. **Pintu Masuk bagi Siklus Manajemen Insiden (A.5.25 & A.5.26)**: Pelaporan kejadian keamanan (*event reporting*) merupakan gerbang utama yang memicu proses penilaian kejadian (A.5.25) dan respon terhadap insiden (A.5.26). Tanpa laporan masuk yang cepat, tim tanggap insiden (CSIRT/DevSecOps) tidak dapat melakukan penahanan dampak (*containment*).
4. **Pencegahan Eskalasi Kerusakan & Dampak Finansial**: Deteksi dini melalui laporan pengguna dalam kurun waktu menit pertama memungkinkan pemutusan sesi JWT (Redis blacklist), penguncian akun (`LockedUntil`), dan isolasi jaringan sebelum data rahasia kertas kerja audit dieksfiltrasi keluar.
5. **Kepatuhan Terhadap Regulasi Pelindungan Data Pribadi (UU PDP No. 27/2022 Pasal 46)**: Kegagalan pelindungan data pribadi wajib diberitahukan kepada lembaga otoritas dan subjek data dalam kurun waktu maksimal 3 x 24 jam. Kepatuhan hukum ini mensyaratkan personel segera melapor saat pertama kali mencurigai terjadinya pelanggaran kerahasiaan.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur kode repositori `risk-based-audit`, layanan backend microservices, dan antarmuka pengguna, komponen teknis pendukung pelaporan mencakup:

### 2.1 Telemetri & Pencatatan Event Keamanan Otomatis (`sys_alert_logs`)
- **Struktur Pencatatan Peringatan Keamanan**: Model `backend/auth-service/models/sys_alert_log.go` mendefinisikan enumerasi `SysAlertAction` yang menangkap kejadian keamanan otomatis: `SysAlertFailedAttemptLog` (percobaan login gagal), `SysAlertLockedOut` (penguncian akun otomatis), `SysAlertRateLimitExceeded` (serangan brute-force/DoS), `SysAlertChangePermissionLog` (perubahan wewenang), dan `SysAlertRoleRevokedLog`.
- **Penyimpanan Terstruktur Log Peringatan**: Repositori `backend/auth-service/repositories/sys_alert_log_repository.go` menyediakan fungsi `Create`, `FindByAction`, `FindByIPAddress`, dan `FindRecentByUserID` untuk keperluan kueri cepat saat menindaklanjuti laporan kejanggalan dari pengguna.

### 2.2 Streaming Peringatan Keamanan Real-Time via Apache Kafka (`security.suspicious`)
- **Publikasi Kejadian Keamanan Kritis**: File `backend/auth-service/pkg/kafka/events.go` mendefinisikan tipe event `EventSuspiciousActivity` (`"security.suspicious"`), `EventRateLimitExceeded` (`"security.rate_limit"`), `EventNewDeviceLogin` (`"security.new_device"`), dan `EventPermissionDenied` (`"permission.denied"`).
- **Fungsi PublishSecurityEvent**: Sistem secara otomatis memublikasikan muatan data peringatan keamanan ke Kafka broker yang dapat dikonsumsi oleh layanan pemantauan terpusat untuk peringatan dini Tim DevSecOps.

### 2.3 Pemantauan Login dari Perangkat Baru (*Trusted Devices Tracking*)
- **Deteksi Akses Tak Dikenal**: Layanan `backend/auth-service/services/trusted-devices/trusted_devices_service.go` memantau sidik jari perangkat pengguna. Jika akun diakses dari browser atau komputer asing yang belum terdaftar, sistem menerbitkan event keamanan dan mengirimkan peringatan ke log audit.

### 2.4 Penangkapan Indikasi Anomali & Galat pada Antarmuka Frontend
- **Intersepsi Galat Otorisasi & Jaringan**: Berkas `frontend/utils/error.ts` dan composable `useAppToast.ts` menangkap kesalahan HTTP 401 Unauthorized, 403 Forbidden, dan 429 Too Many Requests secara terpusat, memberikan indikasi visual kepada pengguna serta merekam galat ke console browser.

### 2.5 Jejak Audit Kertas Kerja & Riwayat Laporan Hasil Audit
- **Audit Trail Modifikasi Kertas Kerja**: Layanan backend `audit-service` mencatat setiap perubahan pada matriks risiko, kertas kerja audit (*working paper*), dan laporan hasil audit (*audit result report*), sehingga kejanggalan perubahan data yang dilaporkan personel dapat diverifikasi melalui riwayat log snapshot.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.6.8 menunjukkan 8 kesenjangan yang harus ditangani:

| No | Parameter Pelaporan Kejadian | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **SOP Pelaporan Kejadian Formal (SOP-SEC-004)** | Tersedia SOP tertulis yang disahkan pimpinan mengenai tata cara dan kewajiban pelaporan kejadian keamanan. | Belum ada dokumen SOP formal bertanda tangan; mekanisme pelaporan masih bersifat lisan atau chat ad-hoc. | **Gap Tata Kelola / SOP** |
| 2 | **Kanal Hotline & Kontak Darurat 24/7** | Tersedia nomor hotline darurat (telepon/WhatsApp) dan email insiden resmi (misal: `incident@auditsphere.app`). | Belum ada kanal kontak resmi yang dipublikasikan secara eksplisit untuk situasi darurat keamanan. | **Gap Kanal Komunikasi** |
| 3 | **Formulir Laporan Mandiri di Aplikasi (In-App Form)** | Pengguna dapat melaporkan indikasi kecurigaan keamanan secara langsung melalui antarmuka web AuditSphere. | Frontend belum memiliki menu atau tombol "Laporkan Insiden / Kejanggalan Keamanan". | **Gap Fitur Antarmuka** |
| 4 | **Panduan Indikator Kejadian bagi Pengguna Awam** | Panduan visual ringkas (*cheat sheet*/infografis) mengenai tanda-tanda insiden keamanan yang wajib dilaporkan. | Belum ada materi edukasi atau petunjuk praktis mengenai contoh kejadian yang wajib dilaporkan staf. | **Gap Panduan Edukasi** |
| 5 | **Kebijakan Budaya Bebas Sanksi (No-Blame Culture)** | Kebijakan perlindungan bagi pelapor yang beritikad baik untuk mencegah ketakutan melapor jika melakukan kesalahan. | Belum ada klausul tertulis yang menjamin pelapor jujur tidak dikenai sanksi disipliner otomatis. | **Gap Kebijakan Perlindungan** |
| 6 | **Sistem Tiket & SLA Konfirmasi Respon Tim SecOps** | Setiap laporan masuk mendapatkan nomor tiket resmi dengan SLA konfirmasi respon awal dalam waktu < 15 menit. | Belum ada tiket terintegrasi khusus insiden; respon penanganan bergantung pada ketersediaan personel. | **Gap Sistem & SLA** |
| 7 | **Register Log Pelaporan Kejadian (Intake Log)** | Buku register terpusat untuk mencatat seluruh laporan masuk dari personel beserta klasifikasi awal dan disposisinya. | Belum ada dokumen atau basis data Intake Log yang memisahkan event sistem dengan laporan manual staf. | **Gap Rekam Jejak / Evidence** |
| 8 | **Simulasi & Uji Petik Pelaporan Rutin (Drills)** | Pelaksanaan simulasi pelaporan berkala (misal: simulasi phishing) untuk menguji kecepatan dan kepatuhan staf. | Belum pernah diselenggarakan program pengujian atau simulasi pelaporan insiden keamanan terencana. | **Gap Pengujian / Validasi** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan implementasi teknis di dalam repositori yang mendukung pencatatan kejadian keamanan mencakup:

1. **Model Perekaman Peringatan Keamanan**: File `backend/auth-service/models/sys_alert_log.go` yang mencatat percobaan intrusi, penguncian akun, dan pelanggaran rate limit.
2. **Infrastruktur Publikasi Event Real-Time**: Modul `backend/auth-service/pkg/kafka/events.go` dengan tipe event `EventSuspiciousActivity` dan metode `PublishSecurityEvent`.
3. **Fitur Deteksi Akses Perangkat Baru**: Modul `backend/auth-service/services/trusted-devices/trusted_devices_service.go` yang memicu peringatan saat akun diakses dari peramban asing.
4. **Repositori Akses Data Log Sistemik**: Berkas `backend/auth-service/repositories/sys_alert_log_repository.go` untuk kueri riwayat peringatan keamanan per pengguna dan alamat IP.
5. **Jejak Log Audit Transaksi Kertas Kerja**: Log snapshot dan riwayat revisi pada layanan `audit-service` untuk memverifikasi laporan ketidakwajaran data audit.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif formal yang harus disiapkan untuk memenuhi kriteria kelulusan audit sertifikasi ISO/IEC 27001:2022:

1. **SOP Pelaporan Kejadian Keamanan Informasi (SOP-SEC-004)**: Dokumen resmi bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) yang mengatur alur, saluran, dan batas waktu pelaporan.
2. **Buku Register Log Pelaporan Kejadian Keamanan (Security Event Intake Log)**: Rekam jejak pencatatan laporan kejadian yang masuk dari personel, status verifikasi awal, dan tiket tindak lanjutnya.
3. **Dokumentasi Kanal Hotline & Alur Panggilan Siaga**: Bukti konfigurasi alias email `incident@auditsphere.app` dan nomor WhatsApp hotline siaga insiden.
4. **Panduan Cepat Indikator Kejadian bagi Pengguna (Security Cheat Sheet)**: Lembar panduan visual contoh anomali keamanan yang wajib dilaporkan oleh auditor dan karyawan.
5. **Sampel Tiket Pelaporan oleh Personel (Ticket Samples)**: Bukti korespondensi tiket penanganan laporan kejadian keamanan yang diinisiasi oleh pengguna hingga selesai ditangani.
6. **Laporan Pelaksanaan Simulasi / Uji Petik Pelaporan**: Rekaman hasil drill pelaporan insiden (misalnya simulasi phishing email triwulanan) dan tingkat partisipasi staf.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif yang wajib diregistrasikan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A6.8-01** | SOP Pelaporan Kejadian Keamanan Informasi (SOP-SEC-004) | Dokumen PDF / Word | `docs/policies/SOP-SEC-004-Event-Reporting.pdf` | Lead InfoSec & Reza |
| **EV-A6.8-02** | Panduan Visual & Indikator Kejadian Keamanan Pengguna | Dokumen PDF / Infografis | `docs/guidelines/Security-Event-Indicators.pdf` | DevSecOps & HRD |
| **EV-A6.8-03** | Buku Register Pelaporan Kejadian Keamanan (Intake Log) | Spreadsheet Terproteksi | SecOps Management Portal / Event Intake Register | SecOps Lead (Reza) |
| **EV-A6.8-04** | Sampel Tiket Pelaporan Kejadian Keamanan dari Personel | Tangkapan Layar & File PDF | Helpdesk Ticketing System (Sample Tickets) | DevSecOps Support |
| **EV-A6.8-05** | Konfigurasi Kanal Hotline & Email Pelaporan Insiden | Tangkapan Layar & Config | `docs/evidence/Hotline-Configuration-Proof.pdf` | Lead Infrastructure |
| **EV-A6.8-06** | Laporan Hasil Simulasi Pelaporan Insiden / Phishing Drill | Laporan Resmi PDF | `docs/reports/Q3-Incident-Reporting-Drill.pdf` | Lead InfoSec (Reza) |

---

## 7. Rencana Aksi Perbaikan (Action Plan)

Rencana aksi terstruktur guna menutup seluruh kesenjangan kontrol A.6.8 dalam kurun waktu 3 minggu ke depan:

| No | Tindakan Remediasi / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pelaporan Kejadian (SOP-SEC-004)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) | Reza / Dimas | 2 minggu |
| 2 | **Penyediaan Kanal Komunikasi Resmi Hotline & Email Insiden** | Aktivasi email `incident@auditsphere.app` dan hotline darurat dengan alur siaga 24/7 | Infra & SecOps | 1 minggu |
| 3 | **Pengembangan Fitur Form Laporan Mandiri pada Dashboard Web** | Penyediaan tombol "Laporkan Masalah Keamanan" di antarmuka frontend AuditSphere | Frontend Dev | 2 minggu |
| 4 | **Pembuatan & Distribusi Panduan Indikator Kejadian bagi Staf** | Brosur digital/infografis "Kenali & Laporkan Kejadian Keamanan" ke seluruh pengguna | DevSecOps & HR | 1 minggu |
| 5 | **Pemberlakuan Kebijakan Perlindungan Pelapor (No-Blame Culture)** | Klausul resmi perlindungan pelapor beritikad baik dalam aturan operasional keamanan | HRD & Legal | 2 minggu |
| 6 | **Implementasi Sistem Tiket & Register Intake dengan SLA Ketat** | Sistem intake dengan SLA respon awal < 15 menit dan pembuatan register kasus terpusat | DevSecOps Lead | 1 minggu |
| 7 | **Penyelenggaraan Simulasi / Uji Petik Pelaporan Phishing** | Kampanye simulasi phishing email terencana untuk mengukur kesadaran dan kecepatan lapor staf | Reza | 3 minggu |
| 8 | **Integrasi Alur Pelaporan ke Asesmen A.5.25 & Penanganan A.5.26** | SOP operasional penyerahan tiket laporan yang terverifikasi ke alur triase insiden resmi | Reza / SecOps | 2 minggu |

Setelah tindakan 1, 2, 3, dan 6 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP Pelaporan bertanda tangan, tangkapan layar kanal hotline, dan register tiket laporan pada Master Index.

