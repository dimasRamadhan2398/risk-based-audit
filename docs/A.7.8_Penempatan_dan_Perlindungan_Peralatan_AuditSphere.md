# A.7.8 Penempatan dan Perlindungan Peralatan

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Penempatan dan perlindungan peralatan |
| **Fokus pemeriksaan** | Peralatan ditempatkan untuk mengurangi gangguan dan akses tidak sah |
| **Contoh bukti objektif** | Layout, rack lock, environmental control, inspection record |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Seluruh perangkat keras pendukung sistem AuditSphere (server cluster VPS 202.10.34.160 dan 202.10.34.166, server hypervisor on-premises Proxmox VE, perangkat jaringan, serta workstation auditor) wajib ditempatkan di lokasi terlindung guna memitigasi bahaya fisik, bencana lingkungan, pencurian, dan akses tanpa izin. |
| **Status implementasi** | **Dalam proses** | Server colocation data center telah berada dalam rak tertutup berkunci dan didukung sistem pendingin presisi; namun dokumentasi tata letak rak (*rack elevation diagram*), SOP perlindungan peralatan, dan log inspeksi fisik berkala belum dibakukan secara formal. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia spesifikasi teknis arsitektur hardware 3-tier on-premises dan kontrak colocation data center; namun dokumen SOP Penempatan & Perlindungan Peralatan (SOP-SEC-006), denah rak, buku log kontrol lingkungan, dan formulir inspeksi rutin belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A7.8-01 s/d EV-A7.8-06 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Peralatan utama beroperasi stabil di data center dengan kontrol daya dan pendingin; namun pemeliharaan fisik workstation pengembang/auditor dan inspeksi server on-premises belum terjadwal secara proaktif. |
| **Desain kontrol** | **Memadai** | Desain mencakup penempatan rak server berpagar/berkunci, redundansi catu daya PDU dan UPS, pendingin ruangan presisi (PAC), proteksi kabel terstruktur, serta kebijakan pencegahan intipan layar (*shoulder surfing*). |
| **Status kesiapan** | **Belum siap** | Pedoman kebijakan SOP perlindungan peralatan yang disahkan pimpinan (Dimas), blueprint diagram elevasi rak, rekam jejak inspeksi berkala, dan bukti pemeliharaan UPS belum memenuhi standar audit sertifikasi ISO 27001. |
| **Prioritas** | **Tinggi** | Kerusakan fisik pada peralatan pemrosesan atau pemadaman listrik tak terkontrol dapat mengakibatkan korupsi basis data audit (PostgreSQL), kegagalan disk storage, dan henti operasional sistem secara total. |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah pengesahan SOP-SEC-006 dan pelaksanaan inspeksi fisik peralatan semesteran pertama. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.7.8 **berlaku** untuk seluruh infrastruktur komputasi, perangkat jaringan, dan peralatan kerja AuditSphere karena:

1. **Mitigasi Kerusakan Fisik dan Kegagalan Komponen**: Perangkat pemrosesan informasi (server, storage SAN/NAS, switch jaringan) rentan terhadap debu berlebih, panas ekstrem, fluktuasi tegangan listrik, dan kelembaban. Penempatan yang tepat dan kontrol lingkungan yang ketat menjamin masa pakai perangkat keras serta mencegah malfungsi tak terduga.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.7.8**: Standar mewajibkan organisasi merancang penempatan peralatan dan memberikan perlindungan fisik yang memadai guna mengurangi risiko dari ancaman lingkungan serta mempersempit peluang terjadinya akses fisik tanpa izin atau perusakan aset.
3. **Perlindungan Terhadap Akses Fisik Tanpa Otorisasi (*Tampering & Theft*)**: Server yang dibiarkan terbuka atau workstation yang tidak terawasi memberikan celah bagi penyerang internal/eksternal untuk menyalin basis data secara langsung melalui drive eksternal, memasang perangkat penyadap keras (*hardware keylogger*), atau merusak kabel jaringan.
4. **Dukungan Arsitektur Deployment Multi-Tier (VPS Colocation, Cloud, dan On-Premises)**: AuditSphere mengoperasikan kluster VPS (202.10.34.160 / 202.10.34.166) dan arsitektur on-premises 3 node fisik (App Tier, DB Tier, AI/ML Tier). Setiap node membutuhkan jaminan rak terkunci, aliran udara terpisah (*hot/cold aisle*), dan proteksi daya tanpa putus (UPS).
5. **Pencegahan Kebocoran Informasi Sensitif pada Layar (*Shoulder Surfing*)**: Penempatan layar monitor kerja di area kantor dan workstation auditor harus diatur agar tidak menghadap jendela luar atau lorong lalu lintas umum, menjaga kerahasiaan kertas kerja investigasi fraud dan data pribadi auditee.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan arsitektur deployment infrastruktur pada repositori `risk-based-audit`, komponen perlindungan peralatan yang telah diterapkan meliputi:

### 2.1 Penempatan Server Colocation pada Rak Tertutup (VPS 202.10.34.160 & 202.10.34.166)
- **Penempatan Server dalam Rak Tertutup Berkunci**: Server hosting produksi dan demo ditempatkan pada kabinet rak server 19 inci standar data center yang dilengkapi kunci mekanis di pintu depan dan belakang, mencegah akses fisik langsung oleh pihak luar.
- **Dukungan Fasilitas Pendingin Presisi & Pemadam Api FM-200**: Fasilitas pusat data colocation tier-3 menyediakan sistem pengkondisi udara presisi (PAC) bersuhu konstan 18–22 derajat Celcius dan sistem proteksi kebakaran gas bersih (*clean agent*) yang aman bagi sirkuit elektronik.

### 2.2 Desain Fisik Arsitektur Server On-Premises 3-Tier (Proxmox VE)
- **Pemisahan Komputasi 3 Server Fisik**: Dokumen `docs/architecture_on_premises_vs_cloud.md` mendefinisikan persyaratan minimum 3 server fisik mandiri: Server 1 (App Tier: Kong Gateway, Nuxt, Go microservices), Server 2 (DB & State Tier: PostgreSQL Patroni HA, Redis, Kafka), dan Server 3 (AI/ML Tier: Python AI, model IndoBERT).
- **Larangan Penggunaan Port USB Liar**: Dokumen keselarasan ISO 27001 (`docs/DOKUMEN_KESELARASAN_ISO27001.md` baris 420) secara tegas melarang port USB pada host hypervisor dibiarkan terhubung pada masa operasional produksi.

### 2.3 Perlindungan Catu Daya Redundan (UPS & Dual Power Supplies)
- **Penyediaan UPS Online & Catu Daya Ganda (*Redundant PSU*)**: Server utama dilengkapi unit pasokan daya ganda (*dual PSU*) yang terhubung ke dua jalur PDU (*Power Distribution Unit*) berbeda dan didukung baterai UPS cadangan untuk mencegah lonjakan tegangan (*surge protection*) dan listrik mati mendadak.

### 2.4 Tata Letak Meja Kerja dan Pengaturan Privasi Layar di Kantor
- **Orientasi Meja Kerja Terkendali**: Meja kerja staf pengembang dan analis keamanan diposisikan membelakangi dinding pembatas dan tidak berhadapan langsung dengan pintu masuk umum, mengurangi risiko intipan layar kerja.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.7.8 menunjukkan 8 kesenjangan yang harus ditangani:

| No | Parameter Perlindungan Peralatan | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Kebijakan & SOP Penempatan Peralatan (SOP-SEC-006)** | Tersedia SOP tertulis yang mengatur standar penempatan, perlindungan fisik rak, dan kontrol lingkungan peralatan. | Belum ada dokumen SOP formal bertanda tangan pimpinan; penataan perangkat masih dilakukan secara informal. | **Gap Kebijakan / SOP** |
| 2 | **Diagram Elevasi Rak Server (Rack Layout / Elevation)** | Dokumentasi visual tata letak rak server (posisi server 1, 2, 3, switch jaringan, PDU, dan UPS dalam satuan unit U). | Belum ada diagram resmi elevasi rak server yang dipetakan secara rapi dalam dokumen teknis. | **Gap Dokumentasi Visual** |
| 3 | **Pencatatan Log Suhu & Kelembaban Berkelanjutan** | Pencatatan log suhu dan kelembaban udara ruang server secara kontinyu dengan ambang batas peringatan otomatis. | Belum ada logbook atau rekaman telemetri berkala suhu ruang server on-premises yang terintegrasi. | **Gap Monitoring Lingkungan** |
| 4 | **Jadwal & Checklist Inspeksi Fisik Peralatan Rutin** | Inspeksi fisik bulanan terhadap kebersihan filter udara, kekencangan kabel, status LED PSU, dan kondisi fisik rak. | Belum tersedia formulir inspeksi fisik rutin; pemeriksaan hanya dilakukan jika terjadi gangguan. | **Gap Prosedur Inspeksi** |
| 5 | **Pengamanan Fisik Port Hardware (USB & IPMI Ports)** | Penutupan fisik port USB dan port manajemen yang tidak digunakan dengan physical port lock atau penonaktifan BIOS. | Port fisik pada server on-premises belum diberi proteksi pengunci fisik atau audit port berkala. | **Gap Proteksi Perangkat** |
| 6 | **Manajemen Penataan Kabel Terstruktur (Cable Dressing)** | Kabel data dan kabel listrik dipisahkan jalurnya, diberi label identitas di kedua ujung, dan diikat rapi. | Penataan kabel pada beberapa rak server masih belum berlabel lengkap dan memerlukan perapihan. | **Gap Manajemen Kabel** |
| 7 | **Uji Beban Rutin & Perawatan Preventif Baterai UPS** | Pelaksanaan uji coba pemutusan daya (*power failover test*) dan pemeriksaan usia pakai baterai UPS tiap 6 bulan. | Belum ada catatan rekaman uji beban (*load test*) baterai UPS dan bukti jadwal pemeliharaan preventif. | **Gap Pemeliharaan Catu Daya** |
| 8 | **Standarisasi Privasi Layar Monitor (Privacy Filters)** | Kewajiban penggunaan filter privasi layar (*privacy screen protector*) bagi laptop auditor saat tugas lapangan. | Belum ada aturan tertulis dan pembagian filter privasi layar resmi bagi auditor yang menangani data klien. | **Gap Proteksi Layar** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan konfigurasi perangkat yang mendukung penempatan dan perlindungan peralatan meliputi:

1. **Dokumentasi Spesifikasi Hardware 3-Tier On-Premises**: Berkas `docs/architecture_on_premises_vs_cloud.md` yang merinci spesifikasi server App Tier, Database Tier, dan AI Tier.
2. **Aturan Larangan Port USB Liar**: Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` baris 420 mengenai pembatasan port fisik pada hypervisor Proxmox VE.
3. **Perangkat Kabinet Rak Server Terkunci**: Fisik rak server tertutup berkunci di data center colocation hosting VPS 202.10.34.160 dan 202.10.34.166.
4. **Perangkat Catu Daya UPS Online**: Unit UPS yang terpasang pada rak server utama untuk menjamin stabilitas voltase dan waktu transisi daya cadangan.
5. **Kontrak Fasilitas Data Center Berpendingin PAC**: Perjanjian kerja sama penyediaan fasilitas colocation dengan jaminan suhu 18–22 derajat Celcius dan sistem pemadam api FM-200.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif formal yang harus disiapkan untuk memenuhi kriteria kelulusan audit sertifikasi ISO/IEC 27001:2022:

1. **SOP Penempatan dan Perlindungan Peralatan TIK (SOP-SEC-006)**: Dokumen resmi bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) yang mengatur standar penempatan rak dan kontrol lingkungan.
2. **Diagram Tata Letak Rak Server (Rack Elevation Diagram)**: Gambar arsitektur teknis penempatan unit server 1, 2, 3, switch, PDU, dan UPS dalam rak server.
3. **Buku Log Pemantauan Suhu & Kelembaban Ruang Server**: Rekaman berkala suhu dan kelembaban ruang server on-premises beserta rekaman anomali suhu.
4. **Formulir & Bukti Checklist Inspeksi Fisik Peralatan**: Lembar rekam jejak pemeriksaan kondisi perangkat keras, kipas pendingin, dan lampu indikator error.
5. **Laporan Pengujian Pemeliharaan Preventif UPS & Catu Daya**: Bukti berita acara uji beban baterai UPS dan simulasi pemadaman listrik tanpa gangguan layanan.
6. **Panduan Penataan Meja Kerja & Privasi Layar Auditor**: Brosur panduan penataan sudut pandang monitor dan aturan pelindung layar anti-intip.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif yang wajib diregistrasikan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A7.8-01** | SOP Penempatan dan Perlindungan Peralatan TIK (SOP-SEC-006) | Dokumen PDF / Word | `docs/policies/SOP-SEC-006-Equipment-Protection.pdf` | Lead InfoSec & Reza |
| **EV-A7.8-02** | Diagram Tata Letak & Elevasi Rak Server (*Rack Layout*) | Blueprint / Diagram PDF | `docs/architecture/Rack-Elevation-Diagram.pdf` | Lead Infrastructure |
| **EV-A7.8-03** | Buku Log Pemantauan Suhu & Kelembaban Lingkungan | Spreadsheet Terproteksi | SecOps Management Portal / Environmental Log | DevSecOps Engineer |
| **EV-A7.8-04** | Formulir & Catatan Checklist Inspeksi Fisik Peralatan | Formulir PDF Terisi | `docs/forms/Quarterly-Physical-Inspection-Log.pdf` | Infra & SecOps Support |
| **EV-A7.8-05** | Laporan Pemeliharaan Preventif UPS & Uji Failover Catu Daya | Berita Acara Resmi PDF | `docs/reports/UPS-Preventive-Maintenance-Report.pdf` | Vendor UPS & Lead Infra |
| **EV-A7.8-06** | Panduan Penataan Meja Kerja & Privasi Layar Auditor | Dokumen PDF / Panduan | `docs/guidelines/Screen-Privacy-and-Siting-Policy.pdf` | General Affairs & SecOps |

---

## 7. Rencana Aksi Perbaikan (Action Plan)

Rencana aksi terstruktur guna menutup seluruh kesenjangan kontrol A.7.8 dalam kurun waktu 3 minggu ke depan:

| No | Tindakan Remediasi / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan & Pengesahan SOP Perlindungan Peralatan (SOP-SEC-006)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) | Reza / Dimas | 2 minggu |
| 2 | **Pembuatan Diagram Elevasi Rak Server On-Premises & Colocation** | Dokumen teknis layout rak server yang memetakan posisi tiap server, switch, dan UPS | Lead Infra | 1 minggu |
| 3 | **Implementasi Sistem Monitoring & Log Suhu/Kelembaban Otomatis** | Pemasangan sensor telemetri suhu dengan ambang batas alert jika temperatur melebihi 24°C | DevSecOps | 2 minggu |
| 4 | **Perapihan & Pelabelan Kabel Jaringan serta Pemisahan Jalur Daya** | Standarisasi penataan kabel terstruktur (*structured cabling*) berlabel pada rak server | Infra Team | 1 minggu |
| 5 | **Pemberlakuan Checklist Inspeksi Fisik Peralatan Berkala** | Formulir checklist inspeksi fisik bulanan terhadap hardware, kipas, dan pengunci rak | SecOps Support | 1 minggu |
| 6 | **Pengamanan Fisik Port USB dan Port Manajemen Perangkat Keras** | Penonaktifan port USB yang tidak terpakai melalui BIOS dan pemasangan *physical port blocker* | Lead Infra | 2 minggu |
| 7 | **Pemberian Filter Privasi Layar & Penataan Sudut Monitor Auditor** | Pengadaan filter privasi layar laptop dan penataan sudut monitor di kantor operasional | General Affairs | 2 minggu |
| 8 | **Pelaksanaan Uji Beban UPS & Simulasi Kegagalan Catu Daya Listrik** | Berita acara hasil uji failover pemutusan sumber listrik PLN ke baterai UPS tanpa downtime | Infra & Reza | 3 minggu |

Setelah tindakan 1, 2, 3, dan 5 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP Penempatan dan Perlindungan Peralatan bertanda tangan, Diagram Elevasi Rak, Log Suhu/Kelembaban, dan Berita Acara Inspeksi Fisik pada Master Index.

