# A.7.4 Pemantauan Keamanan Fisik

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Pemantauan keamanan fisik |
| **Fokus pemeriksaan** | Area penting dipantau dan rekaman dikelola aman |
| **Contoh bukti objektif** | CCTV policy, coverage map, retention config, review log |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Infrastruktur AuditSphere mengoperasikan server hosting VPS (202.10.34.160 dan 202.10.34.166), instalasi server on-premises (hypervisor Proxmox VE), dan ruang operasional/NOC tim pengembang; pemantauan fisik terus-menerus wajib diterapkan guna mencegah akses fisik tidak sah, pencurian perangkat keras, sabotase, dan gangguan ketersediaan sistem. |
| **Status implementasi** | **Dalam proses** | Fasilitas colocation data center telah memiliki pengawasan perimeter CCTV 24/7 dan kantor operasional dilengkapi kamera pengawas; namun SOP formal kebijakan CCTV, denah cakupan kamera (*coverage map*), dan log review berkala belum dibakukan secara tertulis. |
| **Ketersediaan bukti** | **Ada sebagian** | Tersedia rekaman visual operasional CCTV dan kontrol gerbang gedung; namun dokumen Kebijakan Pemantauan Fisik (SOP-SEC-005), peta titik pantau kamera, konfigurasi retensi NVR formal, dan sertifikat audit kepatuhan data center pihak ketiga belum terhimpun lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A7.4-01 s/d EV-A7.4-06 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Kamera pemantau aktif merekam area kerja dan ruang server; namun peninjauan rekaman (*video audit*) masih bersifat reaktif saat terjadi anomali tanpa jadwal review proaktif terstruktur. |
| **Desain kontrol** | **Memadai** | Desain mencakup kombinasi pengawasan visual (CCTV beresolusi tinggi), kontrol akses pintu berbasis kunci elektronik/kartu, serta sertifikasi fasilitas fisik data center tier-3. |
| **Status kesiapan** | **Belum siap** | Pedoman kebijakan pemantauan fisik yang disahkan pimpinan (Dimas), peta cakupan denah kamera, bukti sertifikasi SOC 2 / ISO 27001 data center VPS, dan log review berkala belum siap untuk diaudit sertifikasi eksternal. |
| **Prioritas** | **Tinggi** | Kompromi fisik terhadap perangkat keras server atau konsol manajemen dapat melumpuhkan seluruh lapisan kontrol perangkat lunak (*air-gapped breach*, penyalinan disk fisik langsung, atau pemutusan daya paksa). |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah pengesahan SOP-SEC-005 dan finalisasi pengarsipan sertifikasi fisik penyedia data center. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.7.4 **berlaku** untuk seluruh lingkungan infrastruktur fisik dan fasilitas operasional sistem AuditSphere karena:

1. **Perlindungan Terhadap Ancaman Fisik Langsung**: Sistem perangkat lunak yang aman (enkripsi AES-256, RBAC Casbin, TLS 1.3) dapat ditembus seketika jika pihak yang tidak berwenang memperoleh akses fisik ke unit server fisik, media penyimpanan backup (NAS/SAN), atau konsol workstation administrator. Pemantauan fisik mencegah sabotase, penyadapan perangkat keras (*hardware keylogger*), atau pencurian server.
2. **Mandat Standar ISO/IEC 27001:2022 Klausul A.7.4**: Standar mewajibkan seluruh area sensitif dan perimeter keamanan fisik dipantau secara berkesinambungan menggunakan sistem pengawasan fisik guna mendeteksi masuknya pihak yang tidak berhak dan menjamin keselamatan aset informasi.
3. **Model Deployment Hibrida (Cloud, VPS Colocation dan On-Premises)**: AuditSphere beroperasi pada arsitektur multilingkungan: (1) Cluster VPS 202.10.34.160 / 202.10.34.166 pada data center colocation Indonesia, (2) Cloud SaaS pada AWS Jakarta Region `ap-southeast-3`, dan (3) Model instalasi on-premises berbasis host fisik Proxmox VE (*App Tier*, *DB Tier*, *AI/ML Tier*). Setiap lokasi mensyaratkan jaminan pemantauan fisik yang ketat.
4. **Pencegahan Akses Ilegal ke Konsol Administrasi dan Ruang NOC**: Kantor operasional tim pengembang dan ruang pemantauan jaringan (NOC) mengelola kredensial master root, kunci enkripsi, dan portal administrasi tenant; ruangan ini wajib diawasi CCTV untuk mencegah pihak luar mengamati data rahasia audit (*shoulder surfing*).
5. **Kepatuhan Hukum dan Forensik Digital (UU PDP dan UU ITE)**: Rekaman pemantauan fisik yang terproteksi dan bersinkronisasi waktu (NTP) menjadi bukti sah yang tidak terbantahkan (*non-repudiation*) saat dilakukan penyelidikan forensik terhadap insiden kebocoran data audit atau pelanggaran integritas sistem.

---

## 2. Implementasi yang Sudah Ada

Berdasarkan arsitektur deployment infrastruktur pada dokumen teknis repositori `risk-based-audit`, implementasi keamanan fisik yang ada meliputi:

### 2.1 Keamanan Fisik Pusat Data Colocation dan VPS (202.10.34.160 dan 202.10.34.166)
- **Pengawasan Tingkat Data Center Tier-3**: Server produksi dan demo berlokasi di fasilitas data center komersial Indonesia yang dilengkapi perimeter keamanan berlapis, petugas keamanan 24/7/365, pintu masuk air-lock berkunci biometrik/kartu pintar, dan kamera CCTV berspesifikasi tinggi di setiap koridor rak.
- **Penyekatan Rak Server (*Rack Enclosure*)**: Server fisik yang menaungi mesin virtual AuditSphere ditempatkan dalam rak tertutup yang terkunci secara mandiri, di mana kunci fisik hanya dipegang oleh petugas berwenang penyedia data center dan Lead DevSecOps (Reza).

### 2.2 Keamanan Fisik Ruang Server On-Premises (*Proxmox VE Architecture*)
- **Isolasi Fisik 3 Server Produksi On-Premises**: Sebagaimana tercantum dalam dokumen arsitektur (`architecture_on_premises_vs_cloud.md`), instalasi on-premises enterprise mensyaratkan minimum 3 server fisik/hypervisor Proxmox VE yang ditempatkan pada ruang server berpendingin khusus dengan akses terbatas.
- **Larangan Penggunaan Port Fisik Tak Terkontrol**: Dokumen keselarasan ISO 27001 (`DOKUMEN_KESELARASAN_ISO27001.md` baris 420) secara eksplisit menetapkan larangan keras pemasangan perangkat USB liar pada server produksi on-premises.

### 2.3 Pemantauan Fasilitas Kantor Operasional dan Ruang Kerja SecOps
- **CCTV Area Kerja dan Ruang Kontrol**: Kantor operasional AuditSphere dilengkapi kamera pengawas CCTV di area pintu masuk utama, lobi penerima tamu, dan lorong menuju ruang kerja tim teknis/NOC.
- **Buku Tamu Resepsionis Fisik**: Setiap pengunjung eksternal diwajibkan mengisi buku registrasi identitas dan didampingi (*escorted*) oleh staf internal selama berada di area kerja.

### 2.4 Pemantauan Lingkungan Fisik Server (*Environmental Sensors*)
- **Sensor Suhu dan Kelembaban**: Fasilitas colocation dan ruang server privat dilengkapi sensor telemetri suhu udara dan sensor asap/kebakaran yang mengirimkan peringatan otomatis ke teknisi jika terjadi anomali pendinginan.

---

## 3. Identifikasi Kesenjangan (Gap Analysis)

Evaluasi kesesuaian implementasi aktual terhadap klausul ISO/IEC 27001:2022 Kontrol A.7.4 menunjukkan 8 kesenjangan yang harus ditangani:

| No | Parameter Pemantauan Fisik | Target Standar ISO 27001:2022 | Kondisi Aktual di Repositori | Status Kesenjangan |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Kebijakan dan SOP Pemantauan Fisik Formal (SOP-SEC-005)** | Tersedia SOP tertulis yang mengatur operasional CCTV, hak akses NVR, dan tata kelola pemantauan fisik. | Belum ada dokumen SOP formal bertanda tangan pimpinan; operasional CCTV masih berjalan ad-hoc. | **Gap Kebijakan / SOP** |
| 2 | **Peta Cakupan Kamera (CCTV Coverage Map)** | Denah resmi (*floor plan*) yang memetakan titik letak kamera, sudut pandang (FOV), dan titik buta (*blind spot*). | Belum tersedia dokumen denah tata letak kamera di kantor operasional maupun ruang server. | **Gap Dokumentasi Visual** |
| 3 | **Standarisasi Retensi Rekaman Video (30-90 Hari)** | Penyimpanan rekaman CCTV dikonfigurasi otomatis tersimpan minimal 30 hingga 90 hari sebelum tertimpa (*overwrite*). | Konfigurasi kapasitas penyimpanan NVR belum distandarisasi dan belum dipantau kuota retensinya. | **Gap Konfigurasi Retensi** |
| 4 | **Log Review Berkala Rekaman CCTV** | Pencatatan peninjauan berkala (mingguan/bulanan) terhadap rekaman untuk mendeteksi anomali fisik. | Belum ada buku log atau formulir review rekaman; rekaman hanya dilihat jika terjadi insiden nyata. | **Gap Prosedur Review** |
| 5 | **Sertifikasi Keamanan Fisik Data Center Pihak Ketiga** | Laporan audit kepatuhan (SOC 2 Type II / ISO 27001) dari penyedia colocation VPS 202.10.34.160/166. | Sertifikat kepatuhan fisik vendor data center belum diarsipkan secara terpusat dalam repositori. | **Gap Bukti Rantai Pasok** |
| 6 | **Keamanan dan Pembatasan Akses Konsol NVR** | Perangkat perekam NVR dilindungi kata sandi kuat, berada di VLAN manajemen tertutup, dan hak akses terbatas. | Pengamanan akses antarmuka web NVR belum diintegrasikan ke dalam matriks inventaris aset keamanan. | **Gap Keamanan Perangkat** |
| 7 | **Prosedur Permintaan dan Ekstraksi Rekaman Forensik** | SOP permintaan cuplikan rekaman video untuk keperluan investigasi insiden dan berita acara serah terima. | Belum ada form permintaan resmi dan rantai penanganan bukti (*chain of custody*) cuplikan video. | **Gap Prosedur Forensik** |
| 8 | **Rambu Pemberitahuan Pengawasan Sesuai UU PDP** | Pemasangan signage/pemberitahuan resmi bahwa area berada di bawah pengawasan kamera CCTV. | Rambu pemberitahuan pengawasan CCTV belum terpasang secara standar di pintu masuk kantor operasional. | **Gap Transparansi Regulasi** |

---

## 4. Bukti Objektif yang Sudah Tersedia

Artefak dan konfigurasi infrastruktur yang mendukung pemantauan keamanan fisik mencakup:

1. **Dokumentasi Arsitektur Infrastruktur Server**: Dokumen `docs/architecture_on_premises_vs_cloud.md` yang mendefinisikan batas fisik dan jaringan server on-premises dan cloud.
2. **Kebijakan Isolasi Port Fisik Server Produksi**: Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` baris 420 yang melarang koneksi USB fisik tak terkontrol pada hypervisor Proxmox VE.
3. **Perangkat Kamera CCTV Aktif di Kantor**: Unit kamera IP/CCTV yang telah terpasang dan beroperasi di area pintu masuk utama dan ruang kerja operasional.
4. **Kontrak Layanan Hosting Colocation Data Center**: Perjanjian sewa infrastruktur rak server dan VPS pada penyedia data center lokal dengan SLA ketersediaan daya dan pendingin.
5. **Buku Registrasi Pengunjung Fisik**: Catatan buku tamu manual di meja resepsionis untuk mencatat identitas pengunjung eksternal yang mendatangi kantor operasional.

---

## 5. Bukti Objektif yang Masih Kurang

Bukti objektif formal yang harus disiapkan untuk memenuhi kriteria kelulusan audit sertifikasi ISO/IEC 27001:2022:

1. **SOP Pemantauan Keamanan Fisik dan Retensi CCTV (SOP-SEC-005)**: Dokumen resmi bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) yang mengatur tata kelola CCTV.
2. **Peta Denah Cakupan Pengawasan Kamera (CCTV Coverage Map)**: Denah visual titik kamera, orientasi pandang, dan cakupan area penting (server room, pintu masuk, ruang arsip).
3. **Laporan Konfigurasi dan Kapasitas Retensi NVR**: Dokumen konfigurasi teknis yang membuktikan kapasitas harddisk NVR mampu menyimpan rekaman minimal 30 hingga 90 hari.
4. **Buku Log Peninjauan Berkala Rekaman CCTV (CCTV Review Logbook)**: Bukti rekam jejak pemeriksaan sampel rekaman video mingguan/bulanan oleh tim keamanan.
5. **Salinan Sertifikat Audit Kepatuhan Data Center Colocation**: Bukti sertifikat ISO 27001 / SOC 2 Type II fasilitas gedung data center penyedia VPS 202.10.34.160/166.
6. **Formulir Permintaan dan Ekstraksi Rekaman Video (Footage Request Form)**: Template formal permohonan pengambilan rekaman untuk kebutuhan penelusuran insiden forensik.

---

## 6. Usulan Bukti Objektif (Evidence ID)

Daftar usulan bukti objektif yang wajib diregistrasikan ke dalam Master Audit Evidence Index sistem AuditSphere:

| Evidence ID | Nama Bukti Objektif | Format / Media | Lokasi / Sumber | PIC Penyedia |
| :--- | :--- | :--- | :--- | :--- |
| **EV-A7.4-01** | SOP Pemantauan Keamanan Fisik dan CCTV (SOP-SEC-005) | Dokumen PDF / Word | `docs/policies/SOP-SEC-005-Physical-Monitoring.pdf` | Lead InfoSec dan Reza |
| **EV-A7.4-02** | Peta Denah Cakupan Kamera Pengawas (CCTV Map) | Blueprint / Diagram PDF | `docs/architecture/CCTV-Coverage-Map.pdf` | Lead Infrastructure |
| **EV-A7.4-03** | Konfigurasi Teknis Retensi dan Enkripsi Rekaman NVR | Tangkapan Layar dan Laporan | `docs/evidence/NVR-Retention-Configuration.pdf` | DevSecOps Engineer |
| **EV-A7.4-04** | Buku Log Peninjauan Berkala Rekaman CCTV | Spreadsheet Terproteksi | SecOps Management Portal / Physical Review Log | SecOps Support |
| **EV-A7.4-05** | Sertifikat Kepatuhan Keamanan Fisik Data Center VPS | Sertifikat Resmi PDF | `docs/compliance/DataCenter-ISO27001-Certificate.pdf` | Vendor Colocation DC |
| **EV-A7.4-06** | Formulir dan Berita Acara Ekstraksi Rekaman Video | Template PDF / Form | `docs/forms/CCTV-Footage-Request-Form.pdf` | Legal dan SecOps Lead |

---

## 7. Rencana Aksi Perbaikan (Action Plan)

Rencana aksi terstruktur guna menutup seluruh kesenjangan kontrol A.7.4 dalam kurun waktu 3 minggu ke depan:

| No | Tindakan Remediasi / Peningkatan | Target Luaran Teknis | PIC | Target Waktu |
| :-: | :--- | :--- | :--- | :--- |
| 1 | **Penyusunan dan Pengesahan SOP Pemantauan Fisik (SOP-SEC-005)** | Dokumen formal SOP bertanda tangan Reza (Lead InfoSec) dan Dimas (Leadership) | Reza / Dimas | 2 minggu |
| 2 | **Pembuatan Peta Denah Titik Kamera (CCTV Coverage Map)** | Gambar denah lantai kantor operasional dan ruang server dengan identifikasi sudut pandang | Lead Infra | 1 minggu |
| 3 | **Konfigurasi Standar Retensi Rekaman NVR Minimum 60 Hari** | Pengaturan kuota harddisk NVR dengan fitur auto-overwrite dan proteksi rekaman insiden | DevSecOps | 1 minggu |
| 4 | **Pengarsipan Sertifikat Kepatuhan Data Center VPS 202.10.34.x** | Meminta dan mendokumentasikan sertifikat ISO 27001/SOC 2 dari vendor pusat data | Reza | 2 minggu |
| 5 | **Penyusunan Buku Log dan Jadwal Review Rekaman Rutin** | Pemberlakuan logbook inspeksi visual bulanan terhadap sampel rekaman CCTV | SecOps Support | 1 minggu |
| 6 | **Pengerasan Keamanan Konsol NVR dan Segmentasi VLAN** | Isolasi antarmuka NVR ke segmen VLAN manajemen tertutup tanpa akses publik | DevSecOps Lead | 2 minggu |
| 7 | **Pemasangan Rambu Informasi Pengawasan Kamera (UU PDP)** | Pemasangan stiker pengawasan CCTV di titik masuk area kerja untuk kepatuhan hukum | General Affairs | 1 minggu |
| 8 | **Pengujian Rutin Fungsi Kamera dan Sinkronisasi Waktu (NTP)** | Uji fungsi infra-merah malam hari dan verifikasi sinkronisasi waktu jam kamera ke NTP server | Infra Team | 2 minggu |

Setelah tindakan 1, 2, 3, dan 4 selesai diimplementasikan, daftarkan Evidence ID untuk salinan SOP Pemantauan Fisik bertanda tangan, Peta Denah Kamera, Konfigurasi Retensi NVR, dan Sertifikat Data Center pada Master Index.

