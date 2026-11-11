# A.7.3 — Pengamanan Kantor, Ruangan, dan Fasilitas

| Field | Isi |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Pengamanan kantor, ruangan, dan fasilitas |
| **Fokus pemeriksaan** | Area kerja dan fasilitas sensitif diamankan sesuai risiko |
| **Contoh bukti objektif** | Floor plan, kunci/akses, inspeksi, daftar area terbatas |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Legal & Compliance, HR, dan DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 31 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini adalah kontrol fisik, sehingga hampir seluruh buktinya berada di luar repo: status dan alamat kantor (atau ketiadaannya), kunci, akses, CCTV, buku tamu, floor plan, kondisi tempat kerja jarak jauh personel, fasilitas data center penyedia VPS (termasuk identitas penyedia, lokasi, sertifikat, dan laporan audit), lokasi fisik server Data Hub, serta ruang server klien on-premises. Semuanya ditandai **Perlu verifikasi**; tidak ada yang dapat dibuktikan dari repo. Nama penyedia hosting, alamat, dan sertifikasi **tidak ditemukan di repo dan tidak diisi** di dokumen ini. Seluruh isian pada Lampiran A sampai H (kebijakan, daftar area terbatas, templat, jadwal, kuesioner, persyaratan klien, RACI) adalah **draf usulan**: belum disetujui manajemen, belum ada salinan terisi, dan angka (frekuensi inspeksi, masa simpan log, durasi) bukan hasil pengukuran. Nilai rahasia, kunci, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku sebagian** (model tanggung jawab bersama) | Keamanan fisik data center produksi dan Data Hub berada pada penyedia VPS pihak ketiga; keamanan fisik deployment on-premises berada pada klien (A.5.32 Lampiran bagian 2, baris 91). Yang tetap menjadi tanggung jawab AuditSphere: tempat kerja personel (kantor kecil, co-working, atau rumah, belum terverifikasi), perangkat dan media yang dibawa, dokumen cetak, serta pengelolaan penyedia (A.5.20). Pengecualian penuh ("tidak berlaku") **tidak** disarankan: lihat bagian 1. |
| Implementasi status | **Belum diimplementasikan (sebagian bergantung penyedia)** | Tidak ada kebijakan keamanan fisik, daftar area terbatas, floor plan, log akses/tamu, atau inspeksi. Kendali fisik data center ada pada penyedia VPS tetapi tidak ada bukti (penyedia belum tercatat di repo). Draf kebijakan (PHY-001 usulan) dan templat diusulkan pada Lampiran A sampai H. |
| Ketersediaan bukti | **Tidak ada / hanya bukti tidak langsung** | Hanya ada bukti kode dan dokumen yang menunjukkan model hosting (bagian 2). Floor plan, daftar area terbatas, catatan akses, inspeksi, dan sertifikat penyedia **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A7.3-01 s/d EV-A7.3-09) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada inspeksi, catatan tamu, atau kunjungan/audit ke penyedia. Klaim fisik pada dokumen internal berasal dari desain on-premises (Proxmox) dan tidak menggambarkan produksi. |
| Desain kontrol | **Belum dirancang** | Draf PHY-001, klasifikasi area, daftar area terbatas, kuesioner penyedia, dan persyaratan klien diusulkan (Lampiran A sampai H); belum disetujui. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. Nilai ini akan berubah bila keputusan "berlaku sebagian" disetujui dan bukti penyedia diterima. |
| Prioritas | **Sedang** | Risiko fisik terbesar bukan di kantor (mungkin tidak ada), melainkan (a) kepercayaan pada penyedia VPS yang belum dinilai dan (b) perangkat/dokumen di tempat kerja jarak jauh yang beririsan dengan A.6.7 (prioritas Tinggi). Klien lembaga jasa keuangan akan menanyakan keamanan fisik hosting dan Data Hub pada due diligence. |
| Review berikutnya | [isi] | Usulan: tinjauan tahunan SoA, inspeksi sesuai Lampiran E, dan penilaian ulang penyedia tahunan (A.5.19/A.5.22). |

---

## 1. Justifikasi Relevansi (Berlaku sebagian)

Kontrol A.7.3 **berlaku sebagian** untuk AuditSphere karena:

1. **Pelaksana kontrol berbeda per lapisan.** Produksi berjalan di satu VPS dengan Docker Compose milik penyedia hosting pihak ketiga (`.github/workflows/deploy.yml:63-76` mengakses VPS lewat SSH; `docs/ISO_27001/A.5.20_Keamanan_Dalam_Perjanjian_Pemasok.md` bagian 2.1: "Penyedia VPS (nama dan lokasi [isi])", "Lokasi data center tidak tercatat"). Ruang server, kunci, CCTV, listrik, dan pengendalian akses data center adalah tanggung jawab penyedia, bukan AuditSphere. Ini sejalan dengan CSRM-001 (`docs/ISO_27001/cloud_saas/CSRM-001_Cloud_Shared_Responsibility_Matrix.docx`): "CSP bertanggung jawab atas keamanan perangkat fisik, fasilitas data center ..." dan baris "Physical Security & Data Center: fasilitas fisik, kontrol biometrik, CCTV, dan listrik ditangani CSP".
2. **AuditSphere tidak bisa mengklaim "tidak berlaku".** Auditor ISO 27001 akan menilai apakah organisasi memastikan kendali fisik pada pihak yang dipercaya (A.5.19, A.5.20, A.5.22, A.5.23) dan apakah personel mengamankan area kerja serta perangkat. Dokumen ini memerlukan bukti dari penyedia (sertifikat dan laporan audit) agar pengecualian sebagian dapat dipertahankan. Tanpa bukti itu, justifikasi "dikendalikan penyedia" hanya klaim.
3. **Personel bekerja jarak jauh.** Tidak ditemukan di repo alamat, sewa, atau fasilitas kantor; A.6.7 bagian 1 butir 1 mencatat seluruh personel bekerja dari lokasi masing-masing. "Kantor" dapat berupa kantor kecil, ruang co-working, atau rumah. **Status ada/tidaknya kantor fisik tidak dapat diverifikasi dari repo** dan menjadi keputusan manusia (bagian 4 butir 1). Bila ada kantor tetap, seluruh kontrol (kunci, tamu, area terbatas) berlaku penuh untuk kantor itu; bila tidak, kontrol bergeser ke tempat kerja jarak jauh (A.6.7 bagian 4 Lampiran A) dan ke ruang penyimpanan perangkat/media.
4. **Data Hub adalah server terpisah.** `data-server/setup_server.sh:4` menyebut target "AlmaLinux 10 (KVM VPS, 2 CPU / 8 GB RAM / 100 GB Disk)", dan `data-server/SIZING_GUIDE.md:11-18` membedakan lingkungan demo (VPS KVM 2) dari produksi klien (server khusus/bare metal/VM enterprise). Artinya Data Hub pada demo berjalan di VPS pihak ketiga yang **terpisah** dari VPS aplikasi (dihubungkan WireGuard, `data-server/wireguard/`), sehingga ada setidaknya dua lokasi fisik penyedia yang perlu dinilai. Apakah penyedia dan data center kedua VPS sama atau berbeda **perlu verifikasi**.
5. **Deployment on-premises di klien.** `docs/DOKUMEN_KESELARASAN_ISO27001.md:113` dan `:192` menawarkan lisensi on-premises minimum 3 server fisik/VM; templat kontrak (A.5.32 Lampiran bagian 2, baris 91) menyatakan klien bertanggung jawab atas "pencadangan, keamanan fisik/jaringan, dan kontrol akses lingkungan on-premises". Fisik server klien berada di luar kendali AuditSphere; AuditSphere perlu menetapkan persyaratan minimum dan klausul kontrak (Lampiran G).
6. **Klaim fisik pada dokumen internal tidak sesuai produksi.** `docs/DOKUMEN_KESELARASAN_ISO27001.md:402` menyebut residensi data di "AWS Jakarta Region `ap-southeast-3`", dan `:405-420` mendeskripsikan 3 server fisik Proxmox, VLAN, dan IPMI. Produksi adalah satu VPS (A.5.20 bagian 2.1, A.6.7 bagian 4 butir 11). Klaim ini tidak boleh dipakai sebagai bukti kontrol fisik (lihat bagian 4 butir 4).
7. **Backup off-host.** Backup ke Backblaze B2 masih **DRAFT dan belum terpasang** (`docs/backup-b2-setup.md:5`; A.5.20 bagian 4 butir 7). Bila berjalan, keamanan fisik salinan ada pada Backblaze (pihak ketiga); sampai itu, tidak ada salinan di lokasi fisik terpisah yang dapat dibuktikan (memori operasional tim: hanya dump manual pra-deploy di VPS yang sama).
8. **Perangkat dan dokumen berpindah tempat.** Laptop personel memegang kunci SSH, token, dan unduhan LHA/KKA (A.6.7 bagian 4 butir 15). Laptop yang ditinggalkan di tempat umum atau cetakan dokumen klien adalah risiko fisik yang termasuk lingkup A.7.3 (area kerja), A.7.7 (clear desk/clear screen), dan A.7.9/A.7.10 (aset di luar kantor, media).

---

## 2. Implementasi yang Sudah Ada

### 2.1 Model fasilitas yang terlihat dari repo (tidak sama dengan kendali fisik)

| Fasilitas / lapisan | Yang terlihat di repo | Bukti | Kendali fisik terbukti? |
| :--- | :--- | :--- | :--- |
| VPS produksi (aplikasi, database, Kong) | Satu VPS, Docker Compose, akses SSH; penyedia dan lokasi tidak tercatat. | `.github/workflows/deploy.yml:63-76`; `backend/docker-compose.prod.yml`; A.5.20 bagian 2.1 | Tidak. **Perlu verifikasi** (penyedia, lokasi, sertifikat). |
| VPS Data Hub | Server terpisah AlmaLinux 10 (KVM VPS, 2 CPU / 8 GB / 100 GB) untuk demo; produksi klien disarankan server khusus. Terhubung via WireGuard ke server aplikasi; layanan memanggil Data Hub API lewat VPN. | `data-server/setup_server.sh:4`; `data-server/SIZING_GUIDE.md:11-18`; `data-server/wireguard/wg0-server.conf`, `wg0-client.conf` | Tidak. Lokasi fisik dan apakah satu penyedia dengan VPS produksi **perlu verifikasi**. Alamat IP sengaja tidak dicantumkan. |
| Backup off-host | Rencana Backblaze B2 (bucket privat), skrip `backup-prod.sh`; **belum terpasang**. | `docs/backup-b2-setup.md:1-15`; `backend/scripts/backup/` | Tidak. Status DRAFT. |
| Upload/media | Volume `audit_uploads` di VPS; ImageKit opsional (pihak ketiga). Upload tidak dipersist di beberapa jalur dan `/uploads/*` publik tercatat sebagai gap (A.5.36). | `backend/docker-compose.prod.yml`; A.6.7 bagian 4 butir 15 | Tidak (kendali logis, bukan fisik). |
| Tempat kerja personel | Tidak ada alamat atau fasilitas kantor di repo; personel bekerja jarak jauh. | A.6.7 bagian 1 butir 1 | Tidak. **Perlu verifikasi**. |
| Server klien on-premises | Disebut sebagai model deployment; tanggung jawab fisik klien di templat kontrak. | `docs/DOKUMEN_KESELARASAN_ISO27001.md:113`, `:192`, `:402-420`; A.5.32 Lampiran bagian 2 | Tidak ada instalasi referensi di repo; kontrak belum terisi. |

### 2.2 Dokumen yang menyinggung aspek fisik (belum berlaku formal)

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| CSRM-001 | Keamanan fisik, fasilitas data center, biometrik, CCTV, listrik, dan hardware jaringan ditangani CSP; pelanggan bertanggung jawab atas data, identitas, dan konfigurasi. | Templat generik yang mengacu AWS/GCP/Azure (A.5.20 bagian 4 butir 2); belum disesuaikan dengan penyedia VPS nyata. |
| CPA-001 | Kriteria penilaian penyedia: sertifikat ISO 27001 aktif, SOC 2/ISO 27017, lokasi pusat data di Indonesia, audit lokasi data center, DPA. | Kriteria, bukan hasil penilaian; belum ada penilaian terisi untuk penyedia VPS. |
| CSP-001 | Tingkat layanan kritikal mensyaratkan Multi-AZ, DPA, SOC 2, audit kuartalan. | Tidak selaras dengan satu VPS (A.5.20 bagian 4 butir 2). |
| CES-001 | Menyebut model colocation bersertifikasi Tier III untuk kontrol data fisik penuh. | Opsi strategi keluar; tidak menandakan produksi memakai colocation. |
| `DOKUMEN_KESELARASAN_ISO27001.md` bagian 10 | Desain on-premises: 3 server Proxmox, VLAN 10/20/30/40/90, IPMI via VPN Admin; residensi AWS Jakarta. | Desain/penawaran, bukan produksi (bagian 1 butir 6). Tidak ada bagian khusus keamanan fisik/ruang server/kantor (pencarian `fisik`, `datacenter`, `ruang server`, `kantor` 31 Okt 2026 hanya menemukan referensi di atas). |
| ISO-AS-001 v0.1 DRAFT | Klaim segmentasi VLAN dan status *Implemented* untuk kontrol jaringan on-premises. | Tidak didukung bukti repo (A.6.7 bagian 2.4). |
| A.6.7 Lampiran A, RWP-001 usulan | Ketentuan lingkungan kerja jarak jauh (privasi layar, dokumen, perangkat). | Draf; A.7.3 menambahkan sisi fasilitas dan area terbatas. |

### 2.3 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `docs/`, `data-server/`, `scripts/`, dan `backend/` (31 Oktober 2026) tidak menemukan: nama penyedia VPS produksi, nama atau lokasi data center, sertifikasi penyedia (ISO 27001, SOC 2, Tier), perjanjian kolokasi, alamat kantor, floor plan, daftar area terbatas, kebijakan tamu/buku tamu, kebijakan clear desk/clear screen, kebijakan cetak/pemusnahan dokumen, atau jadwal inspeksi fisik. Dokumen demo (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:149-156`) hanya menyebut beberapa penyedia sebagai **rekomendasi umum**, bukan pilihan produksi; nama itu tidak dipakai di dokumen ini.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Floor plan | **Tidak berlaku sampai ada kantor / Perlu verifikasi** | Tidak ada. Bila ada kantor tetap, buat denah sesuai Lampiran C. Untuk tempat kerja jarak jauh, floor plan tidak diwajibkan; gunakan swa-lapor (Lampiran E). Jangan mengarang. |
| Kunci / kendali akses | **Belum tersedia / Perlu verifikasi** | Kantor: tidak terlihat dari repo. Data center: bukti dari penyedia (Lampiran F). Templat register kunci dan log akses pada Lampiran D. |
| Inspeksi berkala | **Belum tersedia** | Belum ada jadwal maupun catatan. Checklist dan jadwal usulan pada Lampiran E. |
| Daftar area terbatas | **Draf tersedia, belum disetujui** | Lampiran B memuat templat dan isian faktual yang terverifikasi; area kantor fisik masih [isi]. |
| Sertifikat/laporan audit data center penyedia (ISO 27001, SOC 2) | **Belum tersedia** | Penyedia belum tercatat. Minta ke penyedia (Lampiran F); verifikasi masa berlaku, ruang lingkup (apakah mencakup lokasi data center yang dipakai), dan penerbit. |
| Perjanjian/SLA penyedia VPS yang memuat keamanan fisik | **Belum tersedia** | A.5.20 bagian 3: tidak ada kontrak/DPA/SLA penyedia VPS di repo. |
| Penilaian penyedia (CPA-001 terisi) | **Belum tersedia** | Templat ada, belum terisi (A.5.20 bagian 3). |
| Bukti lokasi fisik Data Hub | **Belum tersedia / Perlu verifikasi** | Penyedia, lokasi, dan keterpisahan dari VPS produksi belum tercatat. |
| Bukti keamanan fisik deployment on-premises klien | **Belum tersedia** | Tidak ada instalasi klien terdokumentasi; klausul kontrak pada Lampiran G (usulan) dan templat A.5.32. |
| Acknowledgement personel atas kebijakan fisik/clear desk | **Belum tersedia** | Dapat digabung dengan acknowledgement RWP-001 (A.6.7 Lampiran H) dan PHY-001 (Lampiran A). |
| Catatan tamu/vendor yang masuk area kerja | **Belum tersedia** | Templat Lampiran D. |

---

## 4. Catatan / Gap Terverifikasi

1. **Status kantor fisik tidak diketahui.** Tidak ada alamat, sewa, denah, atau penanggung jawab fasilitas di repo. Perlu keputusan manusia: apakah AuditSphere memiliki kantor tetap, co-working, atau sepenuhnya jarak jauh. Jawabannya menentukan apakah Lampiran C, D, dan E untuk kantor dipakai atau dinyatakan tidak berlaku (dengan alasan tertulis di SoA).
2. **Penyedia VPS produksi dan lokasi data center tidak tercatat** (A.5.20 bagian 4 butir 8). Tanpa nama penyedia, tidak ada dasar untuk meminta sertifikat/laporan audit atau menilai lokasi data center. Tindakan 2 dan 3.
3. **Tidak ada bukti keamanan fisik dari penyedia.** Tidak ada sertifikat ISO 27001/SOC 2, laporan audit, kontrak, atau penilaian CPA-001. Pernyataan "keamanan fisik ditangani penyedia" belum dapat dibuktikan; sampai bukti diterima, perlakukan sebagai asumsi tak terverifikasi.
4. **Klaim dokumen tidak selaras dengan produksi.** `DOKUMEN_KESELARASAN_ISO27001.md:402-420` (AWS Jakarta, Proxmox, VLAN, IPMI) dan CSRM-001/CPA-001/CSP-001 (AWS/GCP/Azure, Multi-AZ) tidak menggambarkan satu VPS. Hanya boleh dipakai sebagai desain on-premises atau templat. Koreksi bersama tindakan A.5.20 dan A.6.7.
5. **Data Hub: lokasi fisik dan keterpisahan tidak tercatat.** Diketahui terpisah secara logis (server berbeda, WireGuard), tetapi apakah penyedia/data center sama dengan VPS produksi tidak terlihat dari repo. Bila sama, kegagalan fasilitas penyedia berdampak pada keduanya (risiko ketersediaan, A.7.5, A.8.14). Data Hub memuat data lakehouse klien (Bronze/Silver/Gold) sehingga sensitivitas fisiknya tinggi untuk klien yang memakainya.
6. **Backup off-host belum berjalan** (`docs/backup-b2-setup.md:5`). Tidak ada salinan di lokasi fisik terpisah yang dapat dibuktikan. Setelah B2 berjalan, penyedia dan lokasi bucket B2 perlu dinilai (A.5.20) dan enkripsi backup perlu dipastikan.
7. **Tidak ada kebijakan keamanan fisik, daftar area terbatas, atau floor plan.** Draf pada Lampiran A, B, C.
8. **Tidak ada kebijakan clear desk/clear screen, tamu, atau penanganan cetakan.** A.6.7 Lampiran A butir 4.1-4.5 menyentuh sebagian (privasi layar, tidak mencetak); belum ada kebijakan yang berdiri sendiri maupun acknowledgement. Draf di Lampiran A bagian 5 sampai 7.
9. **Tidak ada inventaris perangkat dan media** (A.6.7 tindakan 8). Tanpa inventaris, area penyimpanan perangkat/media (laptop cadangan, token keamanan, USB, cetakan) tidak dapat diinspeksi. Aset media yang berisi data klien perlu daftar.
10. **Tidak ada ketentuan fisik untuk deployment on-premises di klien** selain satu kalimat tanggung jawab di templat kontrak (A.5.32 Lampiran). Persyaratan minimum (ruang terkunci, akses terdaftar, CCTV, pengendalian USB/konsol, pemusnahan media) belum ada. Draf Lampiran G.
11. **Server on-premises dengan akses staf AuditSphere.** Sesi dukungan/deploy ke server klien berarti personel AuditSphere memasuki ruang fisik klien atau mengakses konsol. Perlu aturan kunjungan (identitas, pengawalan, log) dan pembatasan perangkat bawaan. Lampiran G bagian G.3.
12. **Pengelolaan tamu dan vendor tidak diatur.** Termasuk teknisi perangkat, tim klien yang berkunjung, dan penyedia jasa kebersihan/co-working. Lampiran D.
13. **Kerentanan lingkungan (listrik, pendingin, bencana) pada VPS adalah tanggung jawab penyedia** (A.7.5, A.7.8, A.7.11 di luar dokumen ini). Perlu SLA dan bukti redundansi dari penyedia; satu VPS tunggal tanpa replika berarti tidak ada ketahanan terhadap kegagalan fasilitas (CSP-001 menyatakan Multi-AZ untuk tingkat kritikal; produksi belum memenuhi).
14. **Status dokumen.** Seperti A.5.12 s/d A.6.7, dokumen ini belum ter-commit per 31 Oktober 2026. Tidak ada perubahan kode atau konfigurasi dilakukan dalam penyusunan dokumen ini. Checklist xlsx tidak diubah.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Putuskan model fasilitas** dan catat di SoA: apakah ada kantor tetap/co-working; bila tidak, nyatakan A.7.3 "berlaku sebagian" (penyedia, klien on-premises, tempat kerja jarak jauh) dengan alasan tertulis. **Keputusan manusia diperlukan.** | Direktur Utama, Lead Information Security, Legal & Compliance | Tinggi | [isi] | Keputusan tertulis; baris SoA |
| 2 | **Catat penyedia dan lokasi** VPS produksi dan VPS Data Hub (nama legal, kota/negara data center, akun kontrak) di register pemasok. | DevOps Lead, Legal & Compliance | Tinggi | [isi] | Register pemasok terisi |
| 3 | **Minta dan verifikasi bukti keamanan fisik dari penyedia** memakai Lampiran F: sertifikat ISO 27001/SOC 2 yang mencakup lokasi data center yang dipakai, ringkasan kendali fisik, SLA, dan DPA. Simpan dengan tanggal terima dan masa berlaku. | Legal & Compliance, DevOps Lead | Tinggi | [isi] | Sertifikat/laporan, kuesioner terisi |
| 4 | **Isi CPA-001** untuk penyedia VPS (dan Backblaze, bila B2 dipakai) dengan kriteria fisik; selaraskan CSRM-001 dengan model satu VPS (hapus acuan AWS/GCP/Azure yang tidak dipakai). | Lead Information Security | Sedang | [isi] | CPA-001 dan CSRM-001 terisi dan direvisi |
| 5 | **Setujui dan tetapkan PHY-001** (Lampiran A), termasuk klasifikasi area dan clear desk/clear screen; kumpulkan acknowledgement (bisa digabung dengan RWP-001). | Lead Information Security dengan Legal & Compliance dan HR | Sedang | [isi] | Kebijakan bertanda tangan; acknowledgement |
| 6 | **Tetapkan daftar area terbatas** (Lampiran B) dan, bila ada kantor, floor plan (Lampiran C), register kunci, dan buku tamu (Lampiran D). | Lead Information Security, penanggung jawab fasilitas [isi] | Sedang | [isi] | Daftar area, floor plan, register |
| 7 | **Jadwalkan inspeksi berkala** (Lampiran E): swa-lapor tempat kerja jarak jauh per personel, inspeksi kantor bila ada, tinjauan bukti penyedia tahunan. Laksanakan inspeksi pertama dan simpan catatan. | Security Officer | Sedang | [isi] | Checklist inspeksi terisi |
| 8 | **Tentukan lokasi fisik Data Hub** dan apakah satu penyedia/data center dengan VPS produksi; pertimbangkan pemisahan lokasi bila risiko ketersediaan tinggi (selaraskan A.8.14). | DevOps Lead, `auditsphere-solution-architect` | Sedang | [isi] | Catatan arsitektur dan keputusan |
| 9 | **Jalankan backup B2** (`docs/backup-b2-setup.md`) dan catat keamanan fisik/DPA Backblaze sebelum data klien dikirim (selaras A.5.20 tindakan 9). Perubahan produksi hanya lewat `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh --ref`, tanpa seeder). | `auditsphere-devops`, Lead Information Security | Tinggi | [isi] | Log backup off-site, DPA |
| 10 | **Masukkan persyaratan fisik on-premises** (Lampiran G) ke templat kontrak on-premises dan paket serah terima; sertakan hak meminta bukti kepatuhan fisik dan aturan kunjungan staf AuditSphere. | Legal & Compliance, Project Lead | Sedang | [isi] | Klausul kontrak, formulir pra-instalasi |
| 11 | **Inventaris perangkat dan media** yang menyimpan data kerja (bersama A.6.7 tindakan 8); tetapkan tempat penyimpanan terkunci untuk perangkat cadangan, token keamanan, dan media cetak/USB. | Lead Information Security, Project Lead | Sedang | [isi] | Inventaris |
| 12 | **Koreksi klaim dokumen**: tandai `DOKUMEN_KESELARASAN_ISO27001.md` bagian 10 (AWS Jakarta, Proxmox, VLAN) dan ISO-AS-001 sebagai desain on-premises, bukan produksi; tambahkan bagian keamanan fisik per model hosting. | Lead Information Security, Legal & Compliance | Sedang | [isi] | Revisi dokumen |
| 13 | **Pelatihan**: masukkan clear desk/clear screen, tamu, perangkat di tempat umum, dan pelaporan kehilangan perangkat ke materi A.6.3 (bersama topik T17). | Security Officer | Rendah | [isi] | Materi, daftar hadir |
| 14 | **Daftarkan Evidence ID** setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |
| 15 | Commit dokumen ini bersama annex lain yang belum ter-commit (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |

Catatan pelaksanaan: tindakan 8 dan 9 dapat menyentuh produksi; lakukan lewat `auditsphere-devops`/`auditsphere-platform`, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack di VPS.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A7.3-01 | Model hosting satu VPS produksi dan akses SSH (bukti bahwa keamanan fisik ada pada penyedia) | `.github/workflows/deploy.yml`, `backend/docker-compose.prod.yml` | Internal |
| EV-A7.3-02 | Data Hub: target server terpisah dan koneksi WireGuard | `data-server/setup_server.sh`, `data-server/SIZING_GUIDE.md`, `data-server/wireguard/` (konfigurasi memuat alamat; akses terbatas) | Restricted |
| EV-A7.3-03 | Rencana backup off-host (status DRAFT) | `docs/backup-b2-setup.md`, `backend/scripts/backup/` | Internal |
| EV-A7.3-04 | Dokumen kebijakan/templat terkait (CSRM-001, CPA-001, CSP-001, CES-001; bagian 10 `DOKUMEN_KESELARASAN_ISO27001.md`) beserta catatan ketidaksesuaian | `docs/ISO_27001/cloud_saas/`, `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| EV-A7.3-05 | Kebijakan PHY-001 disetujui (setelah tindakan 5); draf Lampiran A; daftar area terbatas dan floor plan (setelah tindakan 6) | `docs/ISO_27001/A.7.3_Pengamanan_Kantor_Ruangan_dan_Fasilitas.md` (Lampiran A, B, C) | Internal |
| EV-A7.3-06 | Sertifikat ISO 27001/SOC 2 penyedia VPS dan Data Hub, kuesioner terisi, SLA/DPA (setelah tindakan 3) | [belum ada] | Restricted |
| EV-A7.3-07 | CPA-001 terisi dan register pemasok dengan nama dan lokasi data center (setelah tindakan 2 dan 4) | [belum ada] | Restricted |
| EV-A7.3-08 | Catatan inspeksi (swa-lapor tempat kerja jarak jauh, inspeksi kantor bila ada), register kunci, buku tamu (setelah tindakan 6 dan 7; Lampiran D, E) | [belum ada] | Restricted |
| EV-A7.3-09 | Klausul kontrak on-premises dan formulir pra-instalasi klien terisi (setelah tindakan 10; Lampiran G) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 7 selesai, tambahkan ID per siklus untuk setiap inspeksi dan tinjauan bukti penyedia.

---

## Lampiran A — Draf Kebijakan Keamanan Fisik (PHY-001, usulan, siap tanda tangan)

Draf ini adalah masukan untuk tindakan 5 dan **belum disetujui**. Nilai pada kurung siku adalah usulan awal; ubah sebelum pengesahan.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | PHY-001 Kebijakan Keamanan Fisik dan Lingkungan (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review tahunan dan setelah insiden fisik atau perubahan lokasi) |
| Berlaku untuk | Seluruh karyawan, kontraktor, dan kontributor; tempat kerja yang dipakai untuk pekerjaan AuditSphere (kantor, co-working, rumah); perangkat, media, dan dokumen cetak; penyedia hosting; serta persyaratan bagi klien on-premises |

**1. Tujuan dan cakupan.** Mencegah akses fisik tanpa izin, kerusakan, pencurian, dan gangguan terhadap informasi, perangkat, dan fasilitas AuditSphere. Kebijakan mengikuti model tanggung jawab bersama: kendali fisik data center dijalankan penyedia hosting dan dipastikan lewat penilaian pemasok (A.5.20, CPA-001); kendali fisik deployment on-premises dijalankan klien dan ditetapkan lewat kontrak; kendali tempat kerja, perangkat, media, dan dokumen dijalankan AuditSphere dan personel.

**2. Klasifikasi area**

| Kelas | Contoh | Perlakuan minimum |
| :--- | :--- | :--- |
| Terbatas (Restricted) | Ruang server/rack data center penyedia (dikelola penyedia); ruang penyimpanan perangkat cadangan, token keamanan, dan media berisi data klien; ruang server klien on-premises (dikelola klien) | Akses hanya terdaftar, pengawalan tamu, log akses, pintu terkunci |
| Kerja (Internal) | Kantor/co-working yang dipakai tim; tempat kerja jarak jauh saat memproses data klien | Pintu/area dapat dikunci, layar tidak terlihat orang lain, tidak ada tamu tanpa pengawasan |
| Umum (Public) | Lobi, ruang tunggu, ruang rapat bersama, kafe/co-working area umum | Tidak untuk data klien atau kredensial; tidak meninggalkan perangkat |

**3. Kantor dan ruang kerja** (berlaku bila ada kantor tetap/co-working; **Keputusan: [ada / tidak ada / co-working]**)

| No | Ketentuan |
| :--- | :--- |
| 3.1 | Pintu masuk kantor dikunci di luar jam kerja; kunci/kartu dicatat dan dikembalikan saat personel keluar (Lampiran D). |
| 3.2 | Ruang penyimpanan perangkat/media terkunci; hanya penanggung jawab fasilitas dan Lead Information Security yang memegang kunci. |
| 3.3 | Tamu dan vendor mendaftar, mengenakan penanda, didampingi, dan tidak dibiarkan sendiri di area kerja (Lampiran D). Mereka tidak terhubung ke jaringan kerja; Wi-Fi tamu dipisah. |
| 3.4 | Pengiriman dan kurir diterima di area umum; paket berisi perangkat/media dicatat. |
| 3.5 | Alarm/CCTV (bila ada) diatur sesuai ketentuan privasi dan PDP; rekaman disimpan [30] hari dan diakses terbatas. |
| 3.6 | Fasilitas berbagi (co-working): gunakan loker/ruang privat bagi pekerjaan sensitif; periksa kebijakan keamanan pengelola dan catat di Lampiran B. |

**4. Tempat kerja jarak jauh.** Mengacu RWP-001 (A.6.7 Lampiran A bagian 4). Tambahan: ruang kerja yang dapat dikunci atau perangkat yang dikunci dan disimpan aman bila ditinggal; anggota keluarga/tamu tidak mengakses perangkat atau layar kerja; panggilan yang membahas data klien tidak dilakukan di tempat umum.

**5. Clear desk dan clear screen**

| No | Ketentuan |
| :--- | :--- |
| 5.1 | Kunci layar otomatis maksimum [5] menit (selaras RWP-001 butir 2.4); kunci manual saat meninggalkan perangkat. |
| 5.2 | Di akhir hari kerja atau saat meninggalkan meja lebih dari [15] menit, tidak ada dokumen klien, catatan kredensial, token keamanan, atau media di atas meja; simpan di laci/lemari terkunci. |
| 5.3 | Papan tulis dan catatan tempel berisi detail sistem/klien dihapus setelah rapat. |
| 5.4 | Layar tidak menampilkan data klien saat berbagi layar, rapat tamu, atau di area umum. |

**6. Dokumen cetak, printer, dan media**

| No | Ketentuan |
| :--- | :--- |
| 6.1 | Tidak mencetak dokumen klien kecuali perlu (RWP-001 butir 4.4); printer bersama/publik tidak dipakai. Bila dicetak: ambil segera, simpan terkunci, hancurkan dengan pencacah kertas (potong silang) setelah tidak diperlukan; catat bila memuat data klien. |
| 6.2 | Printer/pemindai kantor (bila ada): cetak dengan PIN atau ambil langsung; fitur penyimpanan dokumen dimatikan. |
| 6.3 | Media removable (USB, disk eksternal) tidak dipakai untuk data klien kecuali disetujui, terenkripsi, dan dicatat; dimusnahkan atau dihapus aman saat tidak dipakai. |
| 6.4 | Perangkat atau media yang dibuang atau diperbaiki pihak ketiga dihapus aman lebih dahulu (RWP-001 butir 2.10). |

**7. Perangkat di luar kantor.** Perangkat tidak ditinggalkan di kendaraan atau tempat umum; dibawa sebagai bagasi kabin; tidak diserahkan ke pihak ketiga. Kehilangan dilaporkan sesuai A.6.7 Lampiran D.

**8. Fasilitas penyedia hosting dan Data Hub.** Penyedia VPS wajib memenuhi persyaratan minimum pada Lampiran F (sertifikat ISO 27001/SOC 2 yang mencakup lokasi, kontrol akses fisik, CCTV, listrik/pendingin, pemusnahan media). Penyedia dinilai sebelum dipakai dan ditinjau tahunan; bukti dikumpulkan dan disimpan. Lokasi data center dicatat di register pemasok. AuditSphere tidak mengklaim kendali fisik penyedia sebagai kendali sendiri.

**9. Deployment on-premises.** Klien bertanggung jawab atas keamanan fisik sesuai Lampiran G, dicantumkan dalam kontrak dan formulir pra-instalasi. Staf AuditSphere yang berkunjung mengikuti aturan klien dan Lampiran G bagian G.3.

**10. Pelaporan.** Kehilangan/pencurian perangkat atau dokumen, akses fisik tidak sah, kerusakan fasilitas, atau pelanggaran kebijakan dilaporkan ke Lead Information Security dalam [1] jam sejak diketahui (A.6.7 butir 7); ditangani sebagai insiden (A.5.24) tanpa menghukum pelaporan jujur.

**11. Peninjauan dan pengecualian.** Inspeksi mengikuti Lampiran E. Pengecualian tertulis, bertenggat, dengan kontrol kompensasi, dicatat di exception register A.5.36.

**Blok persetujuan**

| Peran | Nama | Tanda tangan | Tanggal |
| :--- | :--- | :--- | :--- |
| Penyusun (Legal & Compliance / PM) | [isi] | | |
| Pemilik (Lead Information Security) | [isi] | | |
| Penyetuju (Direktur Utama) | [isi] | | |

---

## Lampiran B — Daftar Area Terbatas dan Klasifikasi Area (templat + isian faktual)

**B.1 Isian faktual yang terverifikasi dari repo (tanpa alamat dan IP)**

| ID | Area / fasilitas | Kelas | Pengelola fisik | Konten sensitif | Status verifikasi |
| :--- | :--- | :--- | :--- | :--- | :--- |
| AR-01 | Data center VPS produksi (rak/host server aplikasi, database, Kong) | Terbatas | Penyedia VPS [nama: **Perlu verifikasi**; lokasi: **Perlu verifikasi**] | Data tenant (KKA, LHA, ATR), kredensial, backup lokal pra-deploy | Fungsi terverifikasi (A.5.20); penyedia belum tercatat |
| AR-02 | Data center VPS Data Hub (AlmaLinux, TimescaleDB lakehouse, AI engine) | Terbatas | Penyedia [nama/lokasi: **Perlu verifikasi**; apakah sama dengan AR-01: **Perlu verifikasi**] | Data lakehouse klien (Bronze/Silver/Gold) | Target server terverifikasi (`data-server/setup_server.sh:4`); penyedia belum tercatat |
| AR-03 | Penyimpanan backup off-host (Backblaze B2, rencana) | Terbatas | Backblaze [lokasi/region bucket: **Perlu verifikasi**] | Dump Postgres dan arsip upload | Rencana; DRAFT, belum terpasang |
| AR-04 | Penyimpanan media pihak ketiga (ImageKit, opsional) | Terbatas | ImageKit [lokasi: **Perlu verifikasi**] | Berkas media/unggahan | Opsional; status produksi [konfirmasi] |
| AR-05 | Server klien on-premises (per klien) | Terbatas | Klien | Seluruh data klien | Per kontrak; Lampiran G |

**B.2 Templat isian area yang dikelola AuditSphere (isi saat keputusan tindakan 1 dan 6)**

| ID | Area | Kelas | Penanggung jawab | Kendali akses (kunci/kartu/PIN) | Siapa berhak masuk | Tamu | Konten sensitif | Inspeksi terakhir |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| AR-10 | Kantor / ruang kerja tim [alamat: **Perlu verifikasi / bila ada**] | Kerja | [isi] | [isi] | [isi] | [isi] | [isi] | [isi] |
| AR-11 | Ruang/lemari penyimpanan perangkat cadangan, token keamanan, media | Terbatas | [isi] | [isi] | [isi] | Tidak | Perangkat, media berisi data | [isi] |
| AR-12 | Ruang cetak/pencacah kertas (bila ada) | Kerja | [isi] | [isi] | [isi] | Dengan pengawalan | Cetakan | [isi] |
| AR-13 | Ruang rapat / co-working (bila dipakai) | Umum | [isi] | [isi] | Pengelola | Ya | Tidak boleh menampung data klien terbuka | [isi] |
| AR-14 | Tempat kerja jarak jauh personel (per personel; swa-lapor) | Kerja | Personel | Kunci pintu/ruang, perangkat terkunci | Personel | Tidak mengakses perangkat kerja | Perangkat, cetakan | [isi] |

Ubah klasifikasi hanya dengan persetujuan Lead Information Security. Area yang tidak ada (mis. AR-10 bila sepenuhnya jarak jauh) dicatat "tidak ada" dengan tanggal keputusan, bukan dihapus.

---

## Lampiran C — Floor Plan (templat dan panduan)

Floor plan **tidak ada** di repo dan **tidak boleh diisi tanpa survei nyata**. Hanya dibuat untuk lokasi yang dikelola AuditSphere (kantor tetap/co-working bila ada). Untuk tempat kerja jarak jauh floor plan tidak diwajibkan; gunakan swa-lapor (Lampiran E). Untuk data center penyedia dan klien, minta ringkasan zonasi dari pihak itu (Lampiran F dan G); jangan menyalin denah rinci ke dokumen ini.

| Elemen | Isi |
| :--- | :--- |
| Identitas | Nama lokasi, alamat [isi], pemilik/pengelola fasilitas, versi denah, tanggal survei, pembuat, klasifikasi dokumen (Restricted) |
| Konten denah | Skala, pintu masuk/keluar, jalur darurat, area dengan kelas (Lampiran B), lokasi penyimpanan perangkat/media terkunci, lokasi printer/pencacah kertas, lokasi CCTV/alarm (bila ada), lokasi pemadam api, titik berkumpul |
| Legenda | Warna/pola per kelas area (Terbatas/Kerja/Umum) |
| Pengendalian | Siapa memegang kunci/kartu tiap pintu; titik pendaftaran tamu |
| Pemeliharaan | Diperbarui bila ada perubahan tata ruang; ditinjau tahunan dan disimpan dengan tanda tangan penanggung jawab |

---

## Lampiran D — Kontrol Kunci, Akses, Tamu, dan Log (templat)

**D.1 Register kunci/kartu/PIN**

| No | ID kunci/kartu | Area | Pemegang | Tanggal terbit | Tanggal kembali/dicabut | Penerbit | Catatan |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | | | | |

Aturan usulan: kunci tidak diduplikasi tanpa persetujuan; kunci hilang dilaporkan segera dan kunci/PIN diganti; kunci dikembalikan atau dicabut saat offboarding (bersama A.6.7 Lampiran D.2); register ditinjau [triwulanan].

**D.2 Log akses dan tamu**

| Tanggal | Nama tamu/vendor | Perusahaan | Identitas diperiksa (Ya/Tidak) | Tujuan | Tuan rumah | Area | Masuk | Keluar | Penanda/kartu | Perangkat dibawa | Paraf tuan rumah |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| | | | | | | | | | | | |

Aturan usulan: tamu selalu didampingi di area Kerja dan dilarang masuk area Terbatas; teknisi pihak ketiga (printer, jaringan) diawasi; log disimpan [12] bulan (dihapus aman setelahnya); data pribadi tamu diminimalkan sesuai UU PDP (konfirmasi Legal & Compliance). Bila kantor co-working, log pengelola dapat dirujuk sebagai bukti dengan permintaan tertulis.

**D.3 Insiden fisik.** Akses tidak sah, pintu tidak terkunci di luar jam, kehilangan kunci/perangkat/dokumen, kerusakan, atau tailgating dilaporkan sesuai PHY-001 bagian 10 dan dicatat di register insiden (A.5.24).

---

## Lampiran E — Checklist Inspeksi dan Jadwal (templat)

**E.1 Jadwal usulan (frekuensi adalah usulan tanpa dasar historis; baseline 31 Okt 2026: tidak ada inspeksi)**

| Objek | Pemeriksa | Frekuensi usulan | Metode | Bukti |
| :--- | :--- | :--- | :--- | :--- |
| Tempat kerja jarak jauh (per personel) | Personel (swa-lapor), diverifikasi sampel oleh Security Officer | Induksi + tiap [6] bulan | Checklist E.2 dan foto tanpa data sensitif / panggilan video | Checklist terisi |
| Kantor / ruang kerja tim (bila ada) | Security Officer | [Bulanan] sederhana; [tahunan] menyeluruh | Walkthrough E.3 | Checklist terisi, foto temuan |
| Penyimpanan perangkat/media terkunci | Security Officer | [Triwulanan] | Cocokkan dengan inventaris | Catatan |
| Register kunci dan log tamu | Lead Information Security | [Triwulanan] | Tinjauan | Berita acara |
| Penyedia VPS/Data Hub/B2 (bukti fisik) | Legal & Compliance, DevOps Lead | Tahunan dan saat perubahan lokasi/sertifikat | Lampiran F | Sertifikat baru, kuesioner |
| Server klien on-premises | Klien; AuditSphere meminta pernyataan | Saat instalasi dan tahunan | Lampiran G | Formulir terisi |

**E.2 Checklist tempat kerja jarak jauh (swa-lapor)**

| No | Pemeriksaan | Ya | Tidak | Bukti | Catatan |
| :--- | :--- | :---: | :---: | :--- | :--- |
| 1 | Ada ruang/sudut kerja yang layarnya tidak terlihat orang lain | | | | |
| 2 | Perangkat dikunci dan disimpan di tempat aman (laci/lemari/ruang terkunci) saat ditinggal | | | | |
| 3 | Pintu ruang kerja/rumah dapat dikunci; kunci tidak dibagikan di luar anggota yang berhak | | | | |
| 4 | Tidak ada dokumen klien tercetak yang dibiarkan terbuka; pencacah kertas tersedia bila mencetak | | | | |
| 5 | Token keamanan, catatan kredensial, dan media tidak ditinggalkan terbuka | | | | |
| 6 | Tamu/anggota keluarga tidak mengakses perangkat kerja | | | | |
| 7 | Kunci layar otomatis diatur (selaras RWP-001) | | | | |
| 8 | Perangkat tidak ditinggalkan di kendaraan/tempat umum | | | | |
| 9 | Router/Wi-Fi rumah aman (A.6.7 Lampiran E butir 14) | | | | |
| 10 | Tahu cara melaporkan kehilangan perangkat/dokumen | | | | |

**E.3 Checklist walkthrough kantor / co-working (bila ada)**

| No | Pemeriksaan | Ya | Tidak | N/A | Temuan / tindakan |
| :--- | :--- | :---: | :---: | :---: | :--- |
| 1 | Pintu masuk dan area Terbatas terkunci; kunci sesuai register | | | | |
| 2 | Tidak ada tamu tanpa pendamping atau tanpa log | | | | |
| 3 | Meja bersih: tidak ada dokumen klien, kredensial, token, atau media terbuka | | | | |
| 4 | Layar terkunci pada perangkat yang ditinggal | | | | |
| 5 | Ruang/lemari penyimpanan perangkat/media terkunci dan sesuai inventaris | | | | |
| 6 | Printer/pencacah kertas: tidak ada cetakan tertinggal; tempat sampah tidak berisi dokumen klien | | | | |
| 7 | Papan tulis dan catatan tempel bersih dari detail sensitif | | | | |
| 8 | Alarm/CCTV/pemadam api berfungsi dan terawat (bila ada) | | | | |
| 9 | Jalur darurat tidak terhalang; denah mutakhir (Lampiran C) | | | | |
| 10 | Wi-Fi tamu dipisah dari jaringan kerja | | | | |
| 11 | Tidak ada perangkat tak dikenal terhubung (USB, colokan jaringan) | | | | |
| 12 | Kunci/kartu hilang atau belum dikembalikan oleh personel keluar | | | | |

Hasil: Patuh / Patuh dengan pengecualian (nomor pengecualian) / Tidak patuh. Tanda tangan pemeriksa dan peninjau (orang berbeda bila memungkinkan), tanggal, daftar tindakan dengan PIC dan tenggat. Ringkasan masuk ke monitoring report A.5.36 Lampiran C dan didaftarkan di Evidence Index.

---

## Lampiran F — Kuesioner dan Permintaan Bukti ke Penyedia Hosting (VPS, Data Hub, B2)

Kirim ke penyedia (satu salinan per penyedia). Jangan mengarang jawaban; tandai "tidak diberikan" bila penyedia menolak. Selaraskan dengan CPA-001. Simpan balasan dengan tanggal dan masa berlaku.

**F.1 Identitas dan lokasi**

| No | Pertanyaan | Jawaban penyedia | Bukti |
| :--- | :--- | :--- | :--- |
| 1 | Nama legal penyedia dan kontak keamanan/kepatuhan | | |
| 2 | Lokasi (kota/negara) data center yang menjalankan VPS kami; apakah dapat dipastikan tidak berpindah tanpa pemberitahuan | | |
| 3 | Apakah data center dimiliki sendiri atau memakai kolokasi/penyedia lain (sub-prosesor); nama dan lokasi | | |
| 4 | Klasifikasi fasilitas (mis. Tier) bila ada, dan dasar klaimnya (sertifikat/pernyataan) | | |

**F.2 Sertifikasi dan laporan audit**

| No | Pertanyaan | Jawaban | Bukti |
| :--- | :--- | :--- | :--- |
| 5 | Apakah memiliki sertifikat ISO/IEC 27001; nomor, penerbit/badan sertifikasi, ruang lingkup, masa berlaku; apakah ruang lingkup mencakup lokasi data center kami | | Salinan sertifikat |
| 6 | Apakah memiliki laporan SOC 2 Type II (atau ISAE 3402) terbaru; cakupan kontrol fisik; periode | | Laporan atau ringkasan (NDA) |
| 7 | Sertifikat/standar lain (ISO 27017/27018, PCI DSS, sertifikasi pusat data lokal) | | |
| 8 | Tanggal audit eksternal terakhir dan apakah ada temuan keamanan fisik yang terbuka | | |

**F.3 Kendali fisik**

| No | Pertanyaan | Jawaban | Bukti |
| :--- | :--- | :--- | :--- |
| 9 | Kendali akses ke gedung dan ruang server (kartu, biometrik, pengawal 24 jam, mantrap); siapa yang berhak masuk | | |
| 10 | Pencatatan dan peninjauan akses; masa simpan log | | |
| 11 | CCTV: cakupan, masa simpan rekaman | | |
| 12 | Penanganan tamu/teknisi pihak ketiga di ruang server | | |
| 13 | Verifikasi latar belakang dan pelatihan staf data center | | |
| 14 | Pemisahan rak/host pelanggan; siapa yang dapat memegang hardware yang menjalankan VPS kami | | |
| 15 | Pemusnahan/penghapusan aman disk dan media saat didaur ulang atau dikembalikan | | |

**F.4 Lingkungan dan ketahanan**

| No | Pertanyaan | Jawaban | Bukti |
| :--- | :--- | :--- | :--- |
| 16 | Redundansi listrik (UPS, genset) dan pendingin | | |
| 17 | Deteksi dan pemadaman kebakaran, deteksi kebocoran, mitigasi banjir/gempa | | |
| 18 | SLA ketersediaan jaringan/host dan kompensasi; riwayat insiden fasilitas | | |
| 19 | Rencana pemulihan bencana penyedia; apakah ada opsi replika di lokasi lain | | |

**F.5 Kontraktual dan insiden**

| No | Pertanyaan | Jawaban | Bukti |
| :--- | :--- | :--- | :--- |
| 20 | Apakah DPA tersedia; sub-prosesor; residensi data | | |
| 21 | Kewajiban dan tenggat notifikasi insiden keamanan fisik yang berdampak pada kami | | |
| 22 | Hak kami melakukan audit/kunjungan atau menerima laporan audit; ketentuan pengakhiran dan penghapusan data (CES-001) | | |

**F.6 Ringkasan evaluasi (diisi AuditSphere)**

| Kriteria | Dipenuhi (Ya/Sebagian/Tidak/Tidak diberikan) | Catatan | Keputusan |
| :--- | :--- | :--- | :--- |
| Sertifikat ISO 27001 aktif dan mencakup lokasi | | | |
| Laporan SOC 2 / audit independen terbaru | | | |
| Kendali akses fisik dan CCTV | | | |
| Listrik/pendingin/pemadam | | | |
| SLA dan DPA | | | |
| Lokasi di wilayah yang dapat diterima (keputusan Legal & Compliance; selaras UU PDP) | | | |

Penilai, peninjau, tanggal, tanggal tinjauan ulang berikutnya [tahunan].

---

## Lampiran G — Persyaratan Fisik untuk Klien On-Premises dan Klausul Kontrak (usulan)

**G.1 Persyaratan minimum (draf; dikonfirmasi Legal & Compliance dan Lead Information Security)**

| No | Persyaratan | Bukti dari klien |
| :--- | :--- | :--- |
| 1 | Server (aplikasi, database, AI/ML, Data Hub bila dipakai) ditempatkan di ruang server/rak terkunci atau data center dengan akses terbatas pada personel terdaftar. | Pernyataan, denah ringkas, daftar personel berhak |
| 2 | Kendali akses fisik (kunci/kartu/biometrik), catatan akses, dan tinjauan berkala. | Prosedur, contoh log |
| 3 | Pengawasan (CCTV/penjaga) pada area server dan masa simpan rekaman. | Pernyataan |
| 4 | Proteksi lingkungan: listrik cadangan (UPS), pendingin, deteksi/pemadam kebakaran. | Pernyataan atau laporan inspeksi |
| 5 | Akses konsol/IPMI/port USB dikendalikan; tidak ada akses fisik oleh pihak tidak berwenang; penyambungan media tidak sah dilarang. | Kebijakan klien |
| 6 | Pengelolaan tamu dan teknisi (pendampingan, pencatatan). | Prosedur |
| 7 | Pemisahan jaringan manajemen dari jaringan umum (selaras desain pada `DOKUMEN_KESELARASAN_ISO27001.md` bagian 10.2 sebagai referensi arsitektur, bukan klaim implementasi). | Diagram jaringan |
| 8 | Media (disk, backup) dihapus/dimusnahkan aman saat diganti atau dibuang; pencadangan di lokasi terpisah secara fisik. | Prosedur, bukti pemusnahan |
| 9 | Bila memakai kolokasi atau penyedia cloud privat, klien memastikan penyedia itu memenuhi persyaratan setara Lampiran F. | Sertifikat penyedia |
| 10 | Pelaporan insiden fisik yang berdampak pada sistem AuditSphere ke AuditSphere dalam [24] jam. | Kontak dan prosedur |

**G.2 Klausul kontrak (usulan, melengkapi A.5.32 Lampiran bagian 2; konsultasikan penasihat hukum)**

> **Keamanan Fisik dan Lingkungan.** Klien bertanggung jawab atas keamanan fisik dan lingkungan infrastruktur tempat perangkat lunak AuditSphere dipasang, sekurang-kurangnya sesuai persyaratan minimum pada Lampiran [G] perjanjian ini, termasuk ruang atau rak terkunci dengan akses terbatas pada personel terdaftar, pencatatan akses, perlindungan lingkungan, dan penghapusan/pemusnahan aman media. Pemberi Lisensi tidak bertanggung jawab atas kerugian yang timbul dari tidak terpenuhinya persyaratan tersebut oleh Klien. Atas permintaan wajar, Klien menyampaikan pernyataan tertulis tahunan atau bukti kepatuhan yang relevan. Klien memberitahu Pemberi Lisensi tanpa penundaan yang tidak wajar, dan dalam [24] jam, atas insiden fisik yang berdampak pada sistem atau data. Akses personel Pemberi Lisensi ke fasilitas Klien mengikuti prosedur Klien (identitas, pendampingan, pencatatan) dan hanya dilakukan sesuai kebutuhan dukungan yang disetujui tertulis.

**G.3 Aturan kunjungan staf AuditSphere ke lokasi klien (usulan)**

1. Kunjungan diminta dan disetujui tertulis oleh klien dan Project Lead; daftar nama dan perangkat yang dibawa disampaikan lebih dahulu.
2. Patuhi prosedur tamu klien; jangan membawa media removable yang tidak perlu; jangan menyambungkan perangkat ke jaringan klien tanpa izin.
3. Tidak meninggalkan perangkat atau dokumen di lokasi klien; data klien tidak dibawa pulang.
4. Catat kunjungan dan tindakan (siapa, kapan, apa) dalam laporan dukungan.

**G.4 Formulir pra-instalasi (diisi klien, usulan)**

| Item | Jawaban klien | Tanggal / Penandatangan |
| :--- | :--- | :--- |
| Lokasi server (gedung/ruang/rak) | | |
| Pengelola fisik dan kontak keamanan | | |
| Kendali akses fisik dan pencatatan | | |
| Perlindungan lingkungan (UPS, pendingin, pemadam) | | |
| Pencadangan lokasi terpisah | | |
| Penyedia kolokasi/cloud (bila ada) dan sertifikatnya | | |
| Pernyataan pemenuhan persyaratan G.1 (ya/sebagian/tidak, dengan pengecualian) | | |

---

## Lampiran H — Penanggung Jawab (RACI usulan)

| Aktivitas | Direktur Utama | Lead Information Security | Legal & Compliance | HR | DevOps Lead | Platform | Security Officer | Project Lead | Penanggung jawab fasilitas [isi] |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Keputusan model fasilitas dan SoA A.7.3 | A (persetujuan) | R | C | C | C | I | C | C | C |
| Kebijakan PHY-001 | A (persetujuan) | R | C | C | C | I | C | C | C |
| Daftar area terbatas, floor plan | I | A | I | I | C | I | R | C | R |
| Register kunci, akses, dan log tamu | I | A | I | C | I | I | R | C | R |
| Inspeksi berkala (kantor, tempat kerja jarak jauh) | I | A | I | I | I | I | R | C | R |
| Penilaian penyedia VPS/Data Hub/B2 dan bukti fisik (Lampiran F) | I | A | R | I | R | C | C | C | I |
| Lokasi fisik Data Hub dan pemisahan | I | C | I | I | A/R | C | C | C | I |
| Persyaratan fisik on-premises dan klausul kontrak | I | C | A/R | I | I | I | I | R | I |
| Clear desk/clear screen, cetakan, media | I | A | I | C | I | I | R | R | R |
| Inventaris perangkat dan media | I | A | I | I | C | I | R | R | C |
| Pelaporan dan penanganan insiden fisik | I | A | R (dampak hukum/PDP) | I | R | C | R | I | R |
| Pelatihan dan acknowledgement | I | A | C | C | I | I | R | R | I |

R = Responsible, A = Accountable, C = Consulted, I = Informed. Pada tim kecil, satu orang dapat memegang beberapa peran, dan "penanggung jawab fasilitas" dapat berupa Project Lead atau pengelola co-working; pemeriksa inspeksi sebaiknya bukan orang yang mengelola area yang diperiksa bila memungkinkan.

---

Catatan sinkronisasi checklist: baris A.7.3 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan tidak diubah oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
