# A.7.13 Pemeliharaan Peralatan

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Pemeliharaan peralatan |
| **Fokus pemeriksaan** | Maintenance dilakukan oleh pihak berwenang dan tercatat |
| **Contoh bukti objektif** | Maintenance schedule, work order, vendor record, approval |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Seluruh peralatan pemrosesan informasi sistem AuditSphere (server fisik, switch, storage, UPS, dan workstation) wajib dipelihara secara berkala guna menjamin ketersediaan (*availability*), keandalan operasional, dan mencegah kegagalan perangkat keras mendadak yang dapat mengganggu layanan audit. |
| **Status implementasi** | **Dalam proses** | Pemeliharaan berkala infrastruktur colocation data center VPS 202.10.34.160/166 berjalan melalui kontrak pihak ketiga dan update software dilakukan tim dev; namun SOP formal pemeliharaan peralatan, jadwal preventif terpadu, dan formulir *work order* belum dibakukan. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia kontrak colocation data center dan riwayat pembaruan sistem; namun dokumen SOP Pemeliharaan Peralatan (SOP-SEC-007), jadwal maintenance tahunan, surat perintah kerja (*work order*), dan berita acara servis teknisi berizin belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A7.13-01 s/d EV-A7.13-06 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Perawatan perangkat lunak dan OS berjalan rutin melalui CI/CD; namun pemeliharaan fisik hardware (pembersihan debu, pengecekan baterai UPS, uji diagnostik disk NVMe) masih bersifat reaktif jika timbul kendala teknis. |
| **Desain kontrol** | **Memadai** | Desain mencakup jadwal pemeliharaan preventif, otorisasi teknisi resmi, perlindungan data sebelum servis luar (sanitasi/pelepasan disk), serta pengawasan ketat saat teknisi pihak ketiga bekerja. |
| **Status kesiapan** | **Belum siap** | Pedoman kebijakan SOP pemeliharaan formal yang disahkan pimpinan (Dimas), jadwal preventif tahunan, formulir izin kerja vendor terverifikasi, dan rekam jejak tiket perbaikan belum memenuhi standar audit sertifikasi ISO 27001. |
| **Prioritas** | **Tinggi** | Kegagalan pemeliharaan perangkat keras (misal: kerusakan disk tanpa peringatan SMART, kipas pendingin mati, atau kegagalan baterai UPS) dapat mengakibatkan kehilangan data audit kritis dan downtime berkepanjangan. |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah pengesahan SOP-SEC-007 dan pelaksanaan siklus pemeliharaan preventif semesteran pertama. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.7.13 **berlaku** untuk seluruh siklus hidup peralatan pemrosesan informasi pada sistem AuditSphere karena:

1. **Jaminan Ketersediaan Layanan Tanpa Henti (*High Availability & SLA*)**: Sistem AuditSphere memproses kertas kerja audit, evaluasi risiko *continuous auditing*, dan analitik kecurangan internal secara waktu nyata. Pemeliharaan preventif yang teratur meminimalkan risiko kegagalan fatal perangkat keras (*hardware outage*) yang dapat melanggar SLA operasional sistem.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.7.13**: Standar mewajibkan seluruh peralatan pemrosesan informasi dipelihara secara benar oleh pihak yang berwenang untuk memastikan ketersediaan, integritas, dan kerahasiaan informasi yang diproses di dalamnya tetap terjaga.
3. **Pencegahan Kebocoran Data Sensitif Saat Perbaikan (*Data Leakage Protection*)**: Perangkat yang memerlukan perbaikan (terutama perbaikan di luar fasilitas / *off-site service*) berpotensi membawa data audit rahasia perusahaan dan data identitas pribadi (PII) auditee. Prosedur pemeliharaan memastikan media penyimpanan dilepas atau dienkripsi sebelum diserahkan ke teknisi luar.
4. **Pengendalian Pihak Ketiga & Teknisi Vendor (*Vendor Escort & Accountability*)**: Pemeliharaan perangkat keras sering kali melibatkan teknisi resmi dari vendor pihak ketiga (server manufacturer, teknisi pendingin data center, teknisi baterai UPS). Kontrol ini menjamin teknisi terverifikasi identitasnya, terikat perjanjian kerahasiaan (NDA), dan selalu didampingi staf internal saat bekerja.
5. **Kepatuhan Ketenagakerjaan dan Tata Kelola Perangkat Kerja**: Workstation dan laptop operasional yang digunakan para auditor dan teknisi DevSecOps memerlukan perawatan rutin (pembersihan fisik, pemeriksaan baterai, pembaruan firmware) guna mendukung kinerja profesional dan keamanan komputasi bergerak.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan penelusuran arsitektur deployment infrastruktur pada dokumen repositori `risk-based-audit`, implementasi pemeliharaan peralatan saat ini mencakup:

### 2.1 Perjanjian Layanan Pemeliharaan Data Center Colocation (VPS 202.10.34.160 / 166)
- **Kontrak Perawatan Fasilitas Tingkat Data Center**: Infrastruktur server VPS AuditSphere di-hosting pada data center komersial tier-3 dengan perjanjian SLA pemeliharaan fasilitas secara berkala mencakup pembersihan filter udara PAC, pengujian genset otomatis, dan pemeliharaan bank baterai UPS.
- **Pemberitahuan Jendela Pemeliharaan Terencana**: Vendor penyedia data center colocation selalu menerbitkan notifikasi tertulis sebelum melakukan kegiatan pemeliharaan (*maintenance window*) dengan jadwal di luar jam sibuk.

### 2.2 Pengelolaan Pemeliharaan Host Hypervisor On-Premises (Proxmox VE)
- **Isolasi Pemeliharaan Modular 3-Tier**: Sebagaimana didefinisikan pada dokumen `docs/architecture_on_premises_vs_cloud.md`, pemisahan arsitektur ke dalam Server 1 (App Tier), Server 2 (DB Tier), dan Server 3 (AI Tier) memungkinkan pelaksanaan pemeliharaan perangkat keras secara bergantian (*rolling maintenance*) tanpa mematikan seluruh ekosistem sekaligus.
- **Pencadangan Penuh Sebelum Tindakan Pemeliharaan**: Sebelum pembaruan sistem atau perbaikan perangkat keras dilakukan, tim operasional melakukan backup snapshot penuh basis data PostgreSQL dan VM state ke penyimpanan sekunder NAS/SAN.

### 2.3 Pemeliharaan Perangkat Lunak & Sistem Operasi Berbasis Pipeline CI/CD
- **Penerapan Patch Keamanan Rutin**: Pembaruan dependensi, security patch kernel Linux, dan patching container dilakukan secara terotomatisasi melalui pipeline build dan testing, memastikan tidak ada kerentanan perangkat lunak yang tertinggal.

### 2.4 Dukungan Perangkat Keras Workstation & Garansi Resmi Vendor
- **Cakupan Garansi Resmi (*OEM Warranty*)**: Komputer kerja dan server utama dilindungi garansi pabrikan resmi (*Next Business Day on-site support*), memastikan penggantian suku cadang yang rusak menggunakan komponen orisinal dari distributor resmi.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.7.13 menunjukkan 8 kesenjangan yang harus ditangani:

| No | Parameter Pemeliharaan Peralatan | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Kebijakan & SOP Pemeliharaan Formal (SOP-SEC-007)** | Tersedia SOP tertulis yang mengatur alur pengajuan, pelaksanaan, persetujuan, dan pencatatan pemeliharaan peralatan. | Belum ada dokumen SOP formal bertanda tangan pimpinan; proses servis hardware masih dilakukan secara ad-hoc. | **Gap Kebijakan / SOP** |
| 2 | **Jadwal Pemeliharaan Preventif (Preventive Schedule)** | Kalender jadwal pemeliharaan preventif tahunan terencana untuk server, UPS, pendingin ruangan, dan laptop. | Belum ada jadwal preventif terstruktur; pemeliharaan umumnya baru dilakukan setelah terjadi anomali/kerusakan. | **Gap Jadwal Preventif** |
| 3 | **Formulir Surat Perintah Kerja & Berita Acara Servis** | Dokumen formulir Work Order (WO) resmi dan Berita Acara Penyelesaian Servis bertanda tangan teknisi dan pengawas. | Perbaikan fisik perangkat belum dilengkapi tiket WO dan formulir serah terima formal. | **Gap Dokumentasi Kerja** |
| 4 | **Prosedur Sanitasi & Pelepasan Disk Servis Luar** | Pelepasan harddisk atau penghapusan data aman (*data wiping*) sebelum peralatan dibawa keluar kantor (*off-site repair*). | Belum ada aturan baku dan formulir serah terima pelepasan storage saat laptop/server diservis ke pusat servis. | **Gap Perlindungan Data** |
| 5 | **Buku Log Riwayat Pemeliharaan (Maintenance Logbook)** | Buku register terpusat yang mencatat riwayat kerusakan, suku cadang yang diganti, nama teknisi, dan tanggal servis. | Riwayat servis perangkat keras belum diarsipkan dalam buku log inventaris yang terpelihara. | **Gap Rekam Jejak / Log** |
| 6 | **Prosedur Pendampingan Teknisi Pihak Ketiga (Escort)** | Kewajiban verifikasi tanda pengenal, penandatanganan NDA, dan pendampingan terus-menerus bagi teknisi vendor. | Belum ada pedoman tertulis mengenai protokol pengawasan teknisi eksternal di ruang kerja/server. | **Gap Pengawasan Pihak Ketiga** |
| 7 | **Verifikasi Keamanan Pasca-Pemeliharaan** | Pemeriksaan fungsi keamanan dan pemindaian integritas sistem sebelum perangkat hasil servis disambung ke jaringan. | Perangkat yang selesai diservis langsung digunakan kembali tanpa melalui checklist verifikasi keamanan. | **Gap Verifikasi Pasca-Servis** |
| 8 | **Pengendalian Sesi Diagnostik Remote Vendor** | Akses jarak jauh teknisi vendor untuk pemeliharaan firmware diatur dengan otorisasi satu kali pakai dan diawasi. | Akses remote pemeliharaan belum memiliki mekanisme approval berbatas waktu dan perekaman sesi resmi. | **Gap Keamanan Akses Remote** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan catatan teknis di dalam repositori yang mendukung pemeliharaan peralatan mencakup:

1. **Dokumentasi Arsitektur Deployment On-Premises**: Berkas `docs/architecture_on_premises_vs_cloud.md` yang memetakan komponen fisik server untuk perencanaan pemeliharaan modular.
2. **Prosedur Backup & Restore Pra-Pemeliharaan**: Berkas `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bagian 12 mengenai SOP pencadangan data sebelum tindakan pemeliharaan sistem.
3. **Perjanjian SLA Colocation Data Center**: Kontrak sewa rak dan VPS 202.10.34.160/166 yang memuat klausul pemeliharaan preventif fasilitas pusat data secara berkala.
4. **Riwayat Log Pembaruan Sistem dan Dependensi**: Catatan commit git dan log pipeline CI/CD yang mendokumentasikan pembaruan software dan security patch secara berkala.
5. **Sertifikat Garansi Resmi Perangkat Keras**: Dokumentasi garansi resmi pabrikan untuk unit server utama dan laptop dinas personel.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif formal yang harus disiapkan untuk memenuhi kriteria kelulusan audit sertifikasi ISO/IEC 27001:2022:

1. **SOP Pemeliharaan Peralatan TIK (SOP-SEC-007)**: Dokumen resmi bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) yang mengatur tata kelola servis preventif dan kuratif.
2. **Jadwal Pemeliharaan Preventif Tahunan (Preventive Maintenance Schedule)**: Rencana kerja dan kalender servis berkala seluruh aset perangkat keras pemrosesan informasi.
3. **Formulir Surat Perintah Kerja (Work Order) & Berita Acara Servis**: Berkas administrasi resmi permintaan perbaikan, catatan teknisi, dan persetujuan penyelesaian pekerjaan.
4. **Buku Log Riwayat Pemeliharaan Peralatan (Maintenance Logbook)**: Basis data/register terpusat yang mencatat riwayat perawatan dan pergantian suku cadang per unit perangkat.
5. **Formulir Berita Acara Pelepasan & Sanitasi Media Penyimpanan**: Bukti tertulis bahwa storage drive dilepas atau di-wipe sebelum perangkat dibawa keluar kantor.
6. **Laporan Riwayat Servis Vendor Fasilitas Colocation Data Center**: Rekaman arsip pemberitahuan maintenance window dan laporan penyelesaian servis fasilitas data center.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif yang wajib diregistrasikan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A7.13-01** | SOP Pemeliharaan Peralatan TIK (SOP-SEC-007) | Dokumen PDF / Word | `docs/policies/SOP-SEC-007-Equipment-Maintenance.pdf` | Lead InfoSec & Reza |
| **EV-A7.13-02** | Jadwal Pemeliharaan Preventif Tahunan (Annual Schedule) | Kalender / Spreadsheet | `docs/operations/Preventive-Maintenance-Schedule.xlsx` | Lead Infrastructure |
| **EV-A7.13-03** | Paket Formulir Surat Perintah Kerja & Berita Acara Servis | Formulir PDF Terisi | `docs/forms/Work-Order-and-Service-Reports.pdf` | DevSecOps Support |
| **EV-A7.13-04** | Buku Log Riwayat Pemeliharaan Aset (Maintenance Log) | Spreadsheet Terproteksi | SecOps Management Portal / Maintenance Register | Infra & Hardware Team |
| **EV-A7.13-05** | Berita Acara Pelepasan & Sanitasi Storage Servis Luar | Dokumen Berita Acara | `docs/forms/Storage-Sanitization-Signoff.pdf` | DevSecOps Engineer |
| **EV-A7.13-06** | Rekaman Laporan Servis Vendor Data Center Colocation | Email & Laporan PDF | `docs/vendor/DataCenter-Maintenance-Advisories.pdf` | Vendor Colocation DC |

---

## 7. Rencana Aksi Perbaikan (Action Plan)

Rencana aksi terstruktur guna menutup seluruh kesenjangan kontrol A.7.13 dalam kurun waktu 3 minggu ke depan:

| No | Tindakan Remediasi / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Pemeliharaan Peralatan (SOP-SEC-007)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) | Reza / Dimas | 2 minggu |
| 2 | **Penyusunan Jadwal Pemeliharaan Preventif Tahunan** | Kalender jadwal servis berkala untuk server, UPS, pendingin ruangan, dan laptop | Lead Infra | 1 minggu |
| 3 | **Pembuatan Formulir Work Order & Berita Acara Servis Resmi** | Template form standar untuk permohonan perbaikan dan serah terima teknisi | DevSecOps | 1 minggu |
| 4 | **Implementasi Buku Log Pemeliharaan Aset Terintegrasi** | Register pencatatan digital riwayat pemeliharaan yang terhubung ke inventaris aset A.5.9 | Hardware Team | 2 minggu |
| 5 | **Pemberlakuan Prosedur Pelepasan & Sanitasi Disk Servis Luar** | Standar pelepasan harddisk sebelum laptop/server diservis ke pihak ketiga | SecOps Support | 1 minggu |
| 6 | **Pemberlakuan Aturan Pengawasan & Pendampingan Teknisi Pihak Ketiga** | Formulir izin masuk teknisi vendor dengan kewajiban NDA dan pendampingan staf IT | Reza / HRD | 2 minggu |
| 7 | **Penyusunan Checklist Verifikasi Keamanan Pasca-Pemeliharaan** | Formulir checklist uji fungsi dan pemindaian integritas sebelum perangkat aktif kembali | DevSecOps Lead | 1 minggu |
| 8 | **Pengendalian & Perekaman Sesi Remote Maintenance Vendor** | Penerapan akses VPN/SSH sementara dengan otorisasi satu kali pakai dan log sesi | Lead Infra | 3 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP Pemeliharaan Peralatan bertanda tangan, Kalender Jadwal Preventif, Sampel Work Order, dan Buku Log Pemeliharaan pada Master Index.

