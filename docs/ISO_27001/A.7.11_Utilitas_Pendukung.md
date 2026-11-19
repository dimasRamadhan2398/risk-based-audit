# A.7.11 — Utilitas Pendukung

| Field | Isi |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Utilitas pendukung |
| **Fokus pemeriksaan** | Listrik, pendingin, dan utilitas pendukung dipantau dan dipelihara |
| **Contoh bukti objektif** | UPS/genset test, maintenance log, capacity report, alarm |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi (termasuk training clone), Data Hub, dan deployment on-premises klien |
| **PIC** | Qiko (sesuai checklist); didukung DevOps Lead, auditsphere-platform, dan penyedia VPS |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026. Kontrol ini hampir seluruhnya bergantung pada penyedia dan lokasi fisik, sehingga bukti fisik berada di luar repo. Catatan operasional dari VPS (audit read-only 5 Nov 2026 dan catatan internal 10 Nov 2026) belum diverifikasi ulang via SSH dan ditandai **Perlu verifikasi**. Nilai angka (ambang, frekuensi) adalah usulan, bukan hasil pengukuran. Alamat IP dan nilai rahasia tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku sebagian** (model tanggung jawab bersama) | Listrik, pendingin, dan UPS/genset data center VPS dan Data Hub ada pada penyedia; listrik dan pendingin server on-premises ada pada klien (A.7.3 Lampiran G, butir 4). Yang tetap menjadi tanggung jawab AuditSphere: pemantauan dan alarm host, prosedur pemulihan setelah gangguan utilitas, kapasitas, dan keputusan risiko single-VPS. |
| Implementasi status | **Belum diimplementasikan** | Ada mekanisme parsial (restart otomatis, node-exporter, alert aplikasi, backup). Tidak ada alarm yang sampai ke penanggung jawab, tidak ada alert host, tidak ada capacity report, dan tidak ada prosedur pemulihan pasca-pemadaman tertulis. |
| Ketersediaan bukti | **Tidak ada / hanya bukti tidak langsung** | Tidak ada hasil uji UPS/genset, log pemeliharaan, atau laporan kapasitas terukur. Ada desain sizing (`data-server/SIZING_GUIDE.md`), bukan pengukuran. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A7.11-01 s/d EV-A7.11-06) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Belum ada uji pemulihan terdokumentasi. Riwayat restart tidak bersih (lihat bagian 4) menunjukkan pemulihan masih manual. |
| Desain kontrol | **Perlu perbaikan** | Alert aplikasi ada, alert host tidak ada, dan alert tidak punya tujuan notifikasi. Satu VPS tanpa replika. |
| Status kesiapan | **Belum siap** | Bukti penyedia, alarm, dan uji pemulihan belum ada. |
| Prioritas | **Tinggi** | Satu VPS tanpa replika berarti satu kegagalan listrik atau pendingin menghentikan seluruh layanan, termasuk Data Hub bila satu penyedia. Klien lembaga jasa keuangan akan menanyakan ketahanan layanan saat due diligence. |
| Review berikutnya | [isi] | Usulan: tinjauan triwulanan kapasitas dan alarm; uji pemulihan tahunan dan setelah perubahan infrastruktur. |

---

## 1. Justifikasi Relevansi (Berlaku sebagian)

Kontrol A.7.11 **berlaku sebagian** untuk AuditSphere karena:

1. **Pelaksana utilitas berbeda per lapisan.** Listrik, pendingin, dan UPS/genset VPS produksi dan VPS Data Hub dijalankan penyedia (sama dengan A.7.3 bagian 1 butir 1). Whois atas IP VPS menunjuk ke PT Fajar Informasi Globalnet Jaya (BOLEH-NET), tetapi kontrak dan lokasi data center belum tercatat di repo (**Perlu verifikasi**). Server on-premises klien dijalankan klien melalui kontrak (A.7.3 Lampiran G).
2. **Satu VPS tanpa replika.** `docker-compose.prod.yml` menjalankan seluruh layanan pada satu host. Dokumen internal mengklaim arsitektur HA dengan minimum dua unit per komponen kritikal (`docs/DOKUMEN_KESELARASAN_ISO27001.md:484`), dan CSP-001 menyebut Multi-AZ untuk tingkat kritikal. Produksi belum memenuhi keduanya. Keputusan menerima risiko ini harus dicatat di SoA.
3. **Gangguan utilitas berdampak langsung pada data klien.** Layanan menyimpan data audit (KKA, LHA, ATR) di PostgreSQL dan volume upload pada satu host. Pemadaman tanpa shutdown bersih berisiko korupsi atau upload tertinggal (lihat A.8.13 dan A.8.14).
4. **Pemantauan ada, tetapi tidak sampai ke orang.** Prometheus dan node-exporter tersedia (lihat bagian 2), tetapi tidak ada Alertmanager dan tidak ada aturan alert untuk kondisi host. Gangguan listrik atau disk penuh tidak memicu notifikasi apa pun.
5. **Kapasitas belum diukur.** Ada desain sizing dan catatan penambahan swap 2 GB dan pembebasan RAM pada 5 Nov 2026. Belum ada laporan kapasitas berkala.
6. **Klien on-premises memerlukan UPS sebagai syarat.** A.7.3 Lampiran G butir 4 mensyaratkan UPS, pendingin, dan pemadam pada server klien. Ketentuan ini masih usulan.

---

## 2. Implementasi yang Sudah Ada

| Mekanisme | Yang terlihat di repo / catatan | Bukti | Terbukti untuk utilitas? |
| :--- | :--- | :--- | :--- |
| Restart otomatis layanan | `restart: unless-stopped` pada layanan produksi. Berjalan hanya jika Docker daemon ikut start saat boot. | `backend/docker-compose.prod.yml` (baris 17, 31, 49, 79, 113, 144, 158) | Parsial. Auto-start daemon di VPS **Perlu verifikasi**. |
| Healthcheck | Beberapa layanan memiliki healthcheck (tiga blok di compose produksi; layanan yang dicakup belum dipetakan satu per satu). | `backend/docker-compose.prod.yml:24, 41, 71` | Parsial (deteksi, bukan pemulihan utilitas). |
| Metrik host | `node-exporter` (CPU, RAM, disk, swap) dengan scrape job di Prometheus. | `monitoring/prometheus/prometheus.yml:23-25`; `monitoring/docker-compose.monitoring.yml:74-76` | Ada, tetapi hanya bila stack monitoring berjalan di VPS (**Perlu verifikasi**). |
| Aturan alert | 30+ aturan untuk Kong, audit-service, analytics, Postgres, dan fieldwork/dashboard (misalnya `KongDown`, `AuditServiceDown`, `PostgresDown`). Tidak ada aturan untuk host (node down, disk, memori, swap). | `monitoring/prometheus-fieldwork-alerts.yml`, `monitoring/prometheus-dashboard-alerts.yml` | Tidak untuk utilitas. |
| Notifikasi alert | Tidak ada Alertmanager di `monitoring/docker-compose.monitoring.yml`. Alert hanya terlihat di UI Prometheus. | Pencarian `alertmanager` pada `monitoring/` (19 Nov 2026) | **Tidak.** |
| Watchdog VPS | Catatan operasional menyebut skrip yang memeriksa port di 127.0.0.1 dan me-restart container yang tidak merespons. Skrip tidak ada di repo. | Catatan internal (diperbarui 5 Nov 2026) | Tidak terverifikasi (**Perlu verifikasi** via SSH). |
| Backup dan pemulihan data | Cron 02:30 di VPS menjalankan `backup-prod.sh` (catatan 5 Nov 2026). Skrip `restore-test.sh` ada di repo, tanpa catatan uji. | `backend/scripts/backup/`; lihat A.7.3 bagian 2.1 | Sebagian (pemulihan data, bukan utilitas). Uji restore belum tercatat. |
| Mitigasi kapasitas memori | Swap 2 GB dan `vm.swappiness=10` ditambahkan. Ruang RAM tersedia naik dari 1,9 ke 3,6 GB setelah dev stack dihentikan (5 Nov 2026). | Catatan operasional 5 Nov 2026 | Mitigasi tanpa laporan kapasitas. |
| Desain sizing | Spesifikasi referensi demo (2 vCPU, 8 GB, 100 GB) dan produksi klien (16–32 core, 64–128 GB, RAID 10). | `data-server/SIZING_GUIDE.md` bagian 1 | Desain, bukan pengukuran. |

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Hasil uji UPS/genset | **Belum tersedia** | Dimiliki penyedia VPS dan Data Hub. Minta melalui A.7.3 Lampiran F bagian 16. |
| Log pemeliharaan utilitas (penyedia) | **Belum tersedia** | Tidak ada perjanjian atau SLA di repo (A.5.20 bagian 3). |
| Laporan kapasitas (disk, RAM, CPU, swap, volume upload) | **Belum tersedia** | Hanya desain sizing dan catatan perubahan. Perlu laporan terukur dari node-exporter. |
| Alarm / notifikasi gangguan host | **Belum tersedia** | Tidak ada Alertmanager dan tidak ada aturan host. |
| Catatan insiden listrik atau restart tidak bersih | **Sebagian** | Catatan operasional mencatat restart tidak bersih yang memerlukan perbaikan manual (bagian 4 butir 4). Belum ada berita acara atau tiket. |
| Prosedur pemulihan pasca-pemadaman | **Belum tersedia** | Belum ada dokumen tertulis. |
| Sertifikat/laporan Tier, SLA, atau redundansi penyedia | **Belum tersedia** | Lihat A.7.3 bagian 3. |
| Ketentuan UPS pada server klien on-premises | **Belum tersedia** | Hanya usulan di A.7.3 Lampiran G. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada bukti utilitas dari penyedia.** Pernyataan "listrik dan pendingin ditangani penyedia" belum dapat dibuktikan. Perlakukan sebagai asumsi tak terverifikasi sampai bukti diterima (sama dengan A.7.3 gap 3).
2. **Satu VPS tanpa replika.** Kegagalan listrik atau pendingin pada satu host menghentikan semua layanan. Klaim HA di `DOKUMEN_KESELARASAN_ISO27001.md:484` tidak berlaku untuk produksi.
3. **Alarm tidak sampai ke orang.** Tanpa Alertmanager dan aturan host, gangguan listrik, disk penuh, atau node-exporter mati tidak terdeteksi kecuali ada yang membuka UI Prometheus.
4. **Pemulihan setelah restart tidak bersih masih manual.** Catatan operasional menyebut Kong gagal start setelah reboot tidak bersih (socket usang, perlu dihapus dan dibuat ulang) dan Kafka perlu waktu lama untuk pulih setelah dihentikan paksa. Keduanya belum punya prosedur tertulis dan belum diuji sebagai bagian dari kontrol utilitas. Catatan ini berasal dari operasi lokal dan VPS dan perlu dikonfirmasi.
5. **Kapasitas tidak dipantau berdasarkan ambang.** Swap dan pembebasan RAM adalah respons, bukan pengukuran. Pertumbuhan volume upload dan database tidak punya proyeksi.
6. **Dua klaim ketersediaan tidak selaras.** CSP-001 (Multi-AZ) dan dokumen HA tidak sesuai dengan satu VPS. Koreksi bersama A.7.3 butir 4.
7. **Uji pemulihan tidak dapat dilakukan dengan reboot di training clone.** Training clone berjalan di VPS yang sama dengan produksi (A.7.3 AR-06). Reboot host atau pemutusan daya akan memengaruhi produksi. Uji pemulihan harus memakai simulasi di tingkat container atau dilakukan di lingkungan terpisah.
8. **Ketentuan UPS dan pendingin untuk klien on-premises belum final.** Ada di usulan A.7.3 Lampiran G, belum disetujui.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Minta bukti utilitas penyedia** VPS produksi dan Data Hub (uji UPS/genset, pemeliharaan pendingin, SLA, sertifikat) memakai A.7.3 Lampiran F bagian 16–19. | Legal & Compliance, DevOps Lead | Tinggi | [isi] | Balasan penyedia, tanggal terima |
| 2 | **Pasang tujuan notifikasi alert.** Tambahkan Alertmanager (atau integrasi notifikasi) ke stack monitoring dan uji sampai alert benar-benar terkirim. | auditsphere-platform | Tinggi | [isi] | Konfigurasi, tangkapan notifikasi uji |
| 3 | **Tambah aturan alert host**: node-exporter down, disk di atas ambang, RAM dan swap tinggi, dan `up` untuk layanan inti. Ambang diputuskan Lead Information Security. | auditsphere-platform | Tinggi | [isi] | File aturan Prometheus, hasil uji |
| 4 | **Verifikasi auto-start Docker daemon dan restart policy** di VPS (`systemctl is-enabled docker`) dan catat hasilnya. | auditsphere-devops | Tinggi | [isi] | Tangkapan output |
| 5 | **Tulis prosedur pemulihan pasca-pemadaman**: urutan start, pemeriksaan Kafka dan Kong, pemeriksaan migrasi, dan langkah untuk socket Kong yang usang. Uji lewat simulasi container di lingkungan selain produksi. | auditsphere-devops | Tinggi | [isi] | Prosedur tertulis, catatan uji |
| 6 | **Buat laporan kapasitas bulanan** (disk, RAM, swap, volume upload, pertumbuhan DB) dari node-exporter dan tinjau dengan A.8.6. | auditsphere-platform, Dimas | Sedang | [isi] | Laporan kapasitas |
| 7 | **Putuskan risiko single-VPS** (replika/HA atau risiko diterima) dan catat di SoA dan CSP-001. | Project Lead, Direktur Utama | Sedang | [isi] | Keputusan tertulis |
| 8 | **Masukkan syarat UPS dan pendingin** ke Lampiran G klien (A.7.3) dan formulir pra-instalasi. | Legal & Compliance | Sedang | [isi] | Klausul kontrak |
| 9 | **Koreksi klaim HA** di `DOKUMEN_KESELARASAN_ISO27001.md:484` agar sesuai produksi. | Lead Information Security | Sedang | [isi] | Revisi dokumen |
| 10 | **Daftarkan Evidence ID** setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A7.11-01 | Konfigurasi restart dan healthcheck layanan produksi | `backend/docker-compose.prod.yml` | Internal |
| EV-A7.11-02 | Stack monitoring (scrape node-exporter dan aturan alert) | `monitoring/prometheus/prometheus.yml`, `monitoring/prometheus-*-alerts.yml` | Internal |
| EV-A7.11-03 | Konfirmasi auto-start Docker dan output `systemctl` di VPS (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A7.11-04 | Bukti utilitas penyedia: uji UPS/genset, SLA, dan sertifikat (setelah tindakan 1) | [belum ada] | Restricted |
| EV-A7.11-05 | Konfigurasi notifikasi alert dan hasil uji alarm (setelah tindakan 2 dan 3) | [belum ada] | Restricted |
| EV-A7.11-06 | Prosedur pemulihan pasca-pemadaman, catatan uji, dan laporan kapasitas (setelah tindakan 5 dan 6) | [belum ada] | Restricted |

---

## Lampiran A — Checklist Uji Utilitas dan Pemulihan (templat, usulan)

Frekuensi adalah usulan. Baseline 19 November 2026: belum ada uji tercatat.

| No | Uji | Dilakukan oleh | Frekuensi usulan | Hasil (Lulus/Gagal) | Bukti |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Uji UPS/genset (dari penyedia) | Penyedia; diterima DevOps Lead | Tahunan atau sesuai penyedia | | Laporan penyedia |
| 2 | Alert host terkirim saat node-exporter dihentikan di lingkungan uji | auditsphere-platform | Triwulanan | | Tangkapan notifikasi |
| 3 | Restart container inti dan pemeriksaan healthcheck (simulasi di tingkat container) | auditsphere-devops | Triwulanan | | Output `docker compose ps` |
| 4 | Pemulihan setelah simulasi pemadaman (urutan start dan pemeriksaan Kafka, Kong, dan migrasi) | auditsphere-devops | Tahunan | | Catatan uji |
| 5 | Uji restore backup (`backend/scripts/backup/restore-test.sh`) | auditsphere-devops | Bulanan | | Hasil restore, waktu pemulihan |
| 6 | Tinjauan laporan kapasitas | Platform, Dimas | Bulanan | | Laporan kapasitas |

Hasil uji yang gagal dicatat sebagai temuan di A.5.36 Lampiran C dan dibuka sebagai tindakan dengan PIC dan tenggat.

---

## Lampiran B — Acuan Silang

| Topik | Lokasi |
| :--- | :--- |
| Pertanyaan penyedia (listrik, pendingin, pemadam, SLA) | A.7.3 Lampiran F bagian 16–19 |
| Persyaratan UPS, pendingin, dan pemadam klien on-premises | A.7.3 Lampiran G bagian G.1 butir 4 |
| Kapasitas dan pemantauan sumber daya | A.8.6 |
| Backup dan pemulihan data | A.8.13; A.7.3 bagian 2.1 |
| Pemulihan layanan setelah gangguan | A.5.30 |

---

Catatan sinkronisasi checklist: baris A.7.11 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan tidak diubah oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
