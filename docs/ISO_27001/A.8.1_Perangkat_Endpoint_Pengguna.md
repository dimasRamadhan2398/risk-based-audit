# A.8.1 — Perangkat Endpoint Pengguna

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Perangkat endpoint pengguna |
| **Fokus pemeriksaan** | Endpoint dikonfigurasi, dipantau, dan dilindungi sesuai standar |
| **Contoh bukti objektif** | Endpoint standard, MDM/EDR report, encryption, patch status |
| **Sistem** | Laptop/workstation personel dan kontraktor yang mengakses repo, VPS, konsol cloud, dan aplikasi AuditSphere (auditsphere.app) |
| **PIC** | Qiko (sesuai checklist); didukung Lead Information Security, Project Lead, dan DevOps Lead |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026. Kontrol ini menilai **keadaan perangkat** (enkripsi disk, patch, proteksi malware, screen lock, inventaris, pelaporan). Tidak satu pun dapat dibuktikan dari repo, jadi seluruh status perangkat ditandai **Perlu verifikasi** dan diperlakukan sebagai gap sampai ada laporan endpoint. Dokumen ini tidak memeriksa atau menyimpan data perangkat siapa pun. Isian pada Lampiran A sampai C adalah **draf usulan** yang belum disetujui manajemen; angka (hari patch, menit screen lock) bukan hasil pengukuran. Melengkapi A.6.7 (kerja jarak jauh; Lampiran E dan F memuat templat checklist dan laporan endpoint) dan A.7.7 (layar bersih). Nilai rahasia, kunci, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Seluruh akses ke repo GitHub, VPS (SSH), konsol cloud, dan alat AI dilakukan dari perangkat personel (A.6.7 bagian 1). Perangkat adalah perimeter praktis. Pengecualian tidak disarankan. |
| Implementasi status | **Belum diimplementasikan** | Tidak ada standar endpoint, inventaris, MDM/EDR, atau laporan kepatuhan. Kendali di sisi aplikasi (MFA opsional, perangkat tepercaya) bukan kendali endpoint. |
| Ketersediaan bukti | **Tidak ada / hanya bukti tidak langsung** | Hanya rujukan di ZTAR-001 (roadmap MDM/EDR, belum dilaksanakan) dan templat A.6.7 Lampiran E dan F (belum terisi). |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.1-01 s/d EV-A8.1-06) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada pengukuran kepatuhan perangkat atau uji perangkat hilang. |
| Desain kontrol | **Belum dirancang** | Standar endpoint dan opsi kendali diusulkan di Lampiran A dan B; belum disetujui. |
| Status kesiapan | **Belum siap** | Gate "dikonfigurasi, dipantau, dilindungi sesuai standar" belum terpenuhi. |
| Prioritas | **Tinggi** | Satu laptop yang hilang atau terinfeksi membawa kunci SSH ke VPS produksi, token GitHub, `.env`, dan akses konsol sekaligus. Tim berkomitmen langsung ke `staging` tanpa PR, sehingga kompromi satu perangkat dapat mengubah kode. |
| Review berikutnya | [isi] | Usulan: laporan endpoint [triwulanan] bersama tinjauan akses jarak jauh (A.6.7 Lampiran G). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.1 **berlaku** karena:

1. **Perangkat personel memegang akses istimewa.** Operator memakai kunci SSH pribadi di laptop untuk deploy (`backend/deploy-prod.sh:79-90`), dan akses GitHub, konsol cloud, serta alat AI juga dari perangkat yang sama (A.6.7 bagian 1 dan 2.1).
2. **Data dan rahasia berada di perangkat.** Repo kerja memuat `.env` per tenant (di `.gitignore`, tetapi tetap ada secara lokal), stack lokal dapat dijalankan penuh (`backend/docker-compose.yml`), dan unduhan LHA/KKA berada di mesin personel. Berkas contoh LHA (`.docx`) terlihat di working tree `backend/`; apakah memuat data klien nyata **Perlu verifikasi**.
3. **Kompromi perangkat berdampak langsung pada kode dan produksi.** Praktik tim adalah komit langsung ke `staging` tanpa PR (memori tim; A.5.36 bagian 2.2). Deploy produksi memakai `workflow_dispatch` atau skrip dari workstation; pengaturan *required reviewers* tidak dapat diverifikasi dari repo.
4. **Token sesi aplikasi bergantung pada keamanan browser.** Token disimpan di cookie `auth-token` yang diset dari sisi klien (tidak `HttpOnly`) dengan `maxAge` hingga 30 hari (`frontend/stores/auth.ts:318`, `:341`; A.6.7 bagian 2.2). Perangkat yang dikompromikan dapat mencuri sesi.
5. **Agen AI berjalan di perangkat.** Riwayat git memuat agen AI. Agen yang berjalan di laptop dapat membaca berkas lokal termasuk `.env` (A.6.7 bagian 1 butir 4; A.6.3 bagian 4 butir 4).
6. **Kontributor dan status perangkat beragam.** Riwayat git menunjukkan lebih dari satu kontributor dengan status karyawan/kontraktor yang tidak dapat diverifikasi (A.5.32 bagian 2.4). Perangkat kontraktor dan BYOD paling sulit dikendalikan.
7. **Klien mengharapkan jawaban.** Due diligence lembaga jasa keuangan lazim menanyakan enkripsi disk, patch, EDR, dan kebijakan BYOD. Endpoint **klien** (perangkat pengguna tenant) bukan lingkup kontrol ini; tanggung jawabnya dibagi lewat kontrak/DPA (A.5.20; A.6.7 tindakan 16).

---

## 2. Implementasi yang Sudah Ada

Semua butir adalah bukti tidak langsung; tidak ada yang membuktikan keadaan perangkat.

| Mekanisme | Yang terlihat | Bukti | Terbukti sebagai kendali endpoint? |
| :--- | :--- | :--- | :--- |
| Rahasia tidak ikut git | `.env`, `.env.local`, `backend/tenants/*/.env` ada di `.gitignore`; tidak ada berkas kunci privat yang dilacak git (diperiksa 31 Okt 2026, A.6.7 bagian 2.1). | `.gitignore:9-16` | Tidak. Hanya mencegah komit, bukan melindungi perangkat. Secret JWT pernah ter-commit (A.5.36 EX-09). |
| Host key SSH dipin di CI | `VPS_KNOWN_HOSTS` wajib; tanpa `ssh-keyscan`. | `.github/workflows/deploy.yml:44-76` | Tidak (kendali jalur deploy, bukan perangkat). |
| Deploy dari workstation | `BatchMode=yes` (kunci saja), multiplexing `ControlPersist=120`. | `backend/deploy-prod.sh:79-90` | Tidak. Kebijakan penyimpanan kunci di laptop tidak ada. |
| MFA TOTP dan perangkat tepercaya di aplikasi | Pengguna dapat mengaktifkan TOTP; perangkat tepercaya melewati MFA; deteksi perangkat baru. | `backend/auth-service/services/auth/auth_service.go:102-135`; `frontend/pages/settings/devices.vue` | Tidak. Ini identitas aplikasi, bukan kepatuhan perangkat. Perangkat hilang tetap tepercaya sampai dihapus (A.6.7 bagian 2.2). |
| Roadmap MDM dan EDR | ZTAR-001 pilar Device: MDM, penilaian postur, EDR; Q3 "MDM & EDR"; perangkat tidak patuh diblokir. | `docs/ISO_27001/` (ZTAR-001, A.6.7 bagian 2.4) | Tidak. Roadmap, belum dilaksanakan. |
| ClamAV | Pemindaian upload di sisi server (fail-closed). | `backend/docker-compose.prod.yml` | Tidak. Melindungi server, bukan endpoint. |
| Templat checklist dan laporan endpoint | Checklist kepatuhan per perangkat dan laporan endpoint triwulanan. | A.6.7 Lampiran E dan F | Templat, belum terisi. |

### 2.1 Hal yang diverifikasi **tidak ada** di repo

Tidak ditemukan: standar atau baseline perangkat; inventaris perangkat; konfigurasi atau laporan MDM/EDR; bukti enkripsi disk, status patch OS, antivirus, atau screen lock; kebijakan BYOD; panduan agen AI pada perangkat; prosedur perangkat hilang yang berjalan (usulan di A.6.7 Lampiran D); atau catatan pelatihan perangkat. Pencarian teks `MDM|EDR|BYOD|endpoint|FileVault|BitLocker` pada `docs/` di luar annex hanya menemukan rujukan ZTAR-001.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Endpoint standard | **Draf di Lampiran A** | Belum disetujui. Selaras A.6.7 Lampiran A bagian 2 (RWP-001, usulan). |
| MDM/EDR report | **Belum tersedia** | Tidak ada MDM/EDR. Lampiran B memuat opsi untuk tim kecil. |
| Bukti enkripsi disk per perangkat | **Belum tersedia / Perlu verifikasi** | Bukti per perangkat (tangkapan layar status enkripsi); templat di A.6.7 Lampiran E. |
| Status patch OS dan aplikasi | **Belum tersedia / Perlu verifikasi** | Sama. |
| Inventaris perangkat | **Belum tersedia** | Templat A.6.7 Lampiran F.1. |
| Laporan endpoint triwulanan | **Belum tersedia** | Templat A.6.7 Lampiran F.2 dan KPI F.3. |
| Acknowledgement personel atas standar endpoint | **Belum tersedia** | Dapat digabung dengan RWP-001 (A.6.7 Lampiran H). |
| Catatan perangkat hilang atau uji prosedur | **Belum tersedia** | Prosedur usulan di A.6.7 Lampiran D; register insiden belum ada (A.5.24). |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada standar endpoint, inventaris, atau laporan kepatuhan.** Gate kontrol (endpoint dikonfigurasi, dipantau, dilindungi) belum terpenuhi. Draf Lampiran A sampai C.
2. **Kondisi perangkat tidak terlihat.** Enkripsi disk, patch OS, antivirus/EDR, screen lock, dan manajer password tidak dapat dibuktikan dari repo. Perlakukan sebagai gap sampai laporan endpoint pertama terisi.
3. **Tidak ada MDM/EDR.** ZTAR-001 menyebutnya sebagai roadmap; tidak ada tindakan. Untuk tim kecil, swa-lapor dengan bukti tangkapan layar adalah langkah awal yang realistis (Lampiran B).
4. **Kunci dan token istimewa tersimpan di laptop tanpa kebijakan.** Kunci SSH ke VPS produksi (runbook memakai `root`), token GitHub, dan kunci API layanan (Resend, ImageKit, B2) mungkin ada di mesin personel (A.6.7 bagian 4 butir 4 dan 13). Tidak ada ketentuan passphrase, kunci perangkat keras, atau pencabutan saat perangkat hilang.
5. **Agen AI di perangkat tidak dibatasi.** Tidak ada ketentuan bahwa agen tidak membaca `.env` atau memegang kredensial produksi (A.6.7 bagian 4 butir 12; AIP-001).
6. **Perangkat hilang tidak punya alur.** Pengguna dapat menghapus perangkat tepercaya miliknya sendiri, tetapi tidak ada alur admin untuk mencabut sesi atau perangkat pengguna lain, dan tidak ada langkah operasional tertulis (A.6.7 bagian 4 butir 14).
7. **BYOD dan perangkat kontraktor tidak diatur.** Status karyawan/kontraktor dan perangkat yang dipakai tidak terdokumentasi.
8. **Salinan data lokal tidak dikendalikan.** Unduhan LHA/KKA, ekspor, stack lokal, dan potensi data uji dari klien dapat berada di laptop. Tidak ada ketentuan penghapusan atau enkripsi (A.6.7 bagian 4 butir 15).
9. **Dokumen ini tidak mengubah apa pun.** Tidak ada perangkat yang diperiksa, tidak ada konfigurasi atau spreadsheet yang diubah. Dokumen belum ter-commit.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Putuskan lingkup perangkat dan BYOD**: perangkat apa saja yang boleh mengakses repo, VPS, dan konsol; apakah BYOD dan perangkat kontraktor diizinkan. **Keputusan manusia diperlukan.** | Lead Information Security, Direktur Utama | Tinggi | [isi] | Keputusan tertulis |
| 2 | **Setujui standar endpoint** (Lampiran A) dan masukkan ke RWP-001 (A.6.7 tindakan 1). | Lead Information Security | Tinggi | [isi] | Standar bertanda tangan |
| 3 | **Buat inventaris perangkat** dan **laporan endpoint pertama** (swa-lapor dengan tangkapan layar, A.6.7 Lampiran E dan F); satu laporan per personel. | Lead Information Security, Project Lead | Tinggi | [isi] | Inventaris dan laporan terisi |
| 4 | **Pilih kendali endpoint** yang layak untuk tim kecil (Lampiran B): swa-lapor berkala atau MDM ringan, dan proteksi malware bawaan OS. | Lead Information Security | Sedang | [isi] | Keputusan; bukti pendaftaran bila MDM dipilih |
| 5 | **Amankan kunci istimewa di perangkat**: kunci SSH `ed25519` berpassphrase atau kunci perangkat keras, kunci CI terpisah, manajer password; cabut kunci saat perangkat hilang atau personel keluar (A.6.7 Lampiran B). | `auditsphere-devops`, personel | Tinggi | [isi] | Daftar fingerprint, catatan rotasi |
| 6 | **Terapkan prosedur perangkat hilang dan offboarding perangkat** (A.6.7 Lampiran D) dan uji tabletop satu kali. | Lead Information Security, DevOps Lead | Tinggi | [isi] | Prosedur disetujui, catatan uji |
| 7 | **Susun ketentuan agen AI di perangkat**: tidak membaca `.env` atau kunci produksi, tidak diberi kredensial produksi, izin dibatasi (selaras AIP-001, A.6.7 tindakan 10). | Lead Information Security, Security Officer | Sedang | [isi] | Klausul AIP-001/RWP-001 |
| 8 | **Atur data lokal**: larang menyimpan data klien nyata di perangkat kecuali perlu dan terenkripsi; hapus saat selesai. Verifikasi apakah berkas contoh LHA di `backend/` memuat data nyata. | Lead Information Security, Project Lead | Sedang | [isi] | Ketentuan, hasil verifikasi |
| 9 | **Tinjau triwulanan** laporan endpoint dan KPI (A.6.7 Lampiran F.3) bersama tinjauan akses jarak jauh. | Security Officer | Sedang | [isi] | Berita acara tinjauan |
| 10 | **Daftarkan Evidence ID** pada Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |
| 11 | Commit dokumen ini bersama annex lain (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.1-01 | Standar endpoint disetujui (setelah tindakan 2) dan draf Lampiran A | `docs/ISO_27001/A.8.1_Perangkat_Endpoint_Pengguna.md`, A.6.7 Lampiran A | Internal |
| EV-A8.1-02 | Keputusan BYOD dan lingkup perangkat (setelah tindakan 1) | [belum ada] | Internal |
| EV-A8.1-03 | Inventaris perangkat dan laporan endpoint per personel (setelah tindakan 3) | [belum ada] | Restricted (data per orang) |
| EV-A8.1-04 | Bukti MDM/EDR atau swa-lapor berkala (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.1-05 | Daftar fingerprint kunci SSH dan catatan pencabutan (setelah tindakan 5) | [belum ada] | Restricted |
| EV-A8.1-06 | Prosedur perangkat hilang dan catatan uji tabletop (setelah tindakan 6) | [belum ada]; draf A.6.7 Lampiran D | Restricted |

---

## Lampiran A — Standar Endpoint (draf usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah usulan awal. Berlaku untuk perangkat yang dipakai mengakses repo, VPS, konsol cloud, atau data klien.

| Area | Persyaratan [usulan] | Bukti per perangkat |
| :--- | :--- | :--- |
| Enkripsi disk | Aktif (FileVault di macOS, BitLocker di Windows, LUKS di Linux); kunci pemulihan disimpan di manajer password. | Tangkapan layar status |
| Patch OS | Pembaruan keamanan dipasang dalam [14] hari; OS masih mendapat dukungan vendor. | Tangkapan layar versi dan pembaruan |
| Patch aplikasi | Browser, Docker, IDE, dan klien Git diperbarui dalam [30] hari. | Tangkapan layar atau daftar versi |
| Proteksi malware | Perlindungan bawaan OS aktif (atau EDR bila dipilih); pembaruan otomatis. | Tangkapan layar |
| Screen lock | Otomatis dalam maksimum [5] menit, perlu kata sandi (selaras A.7.7 dan A.6.7). | Tangkapan layar pengaturan |
| Akun dan autentikasi | Akun pengguna terpisah dari admin; kata sandi dan biometrik perangkat; manajer password; MFA pada akun GitHub dan konsol. | Pernyataan, tangkapan layar |
| Kunci istimewa | Kunci SSH `ed25519` berpassphrase atau kunci perangkat keras; tidak ada kunci produksi di perangkat bersama; `.env` tidak disalin ke folder sinkron. | Daftar fingerprint |
| Firewall | Firewall host aktif. | Tangkapan layar |
| Perangkat lunak | Hanya perangkat lunak yang diperlukan; ekstensi browser ditinjau; agen AI sesuai ketentuan AIP-001. | Daftar aplikasi |
| Data lokal | Data klien nyata tidak disimpan di perangkat tanpa kebutuhan; media removable disetujui dan terenkripsi (selaras A.7.3 butir 6). | Pernyataan |
| Perangkat hilang | Lapor ke Lead Information Security dalam [1] jam; pencabutan kunci, sesi, dan token mengikuti A.6.7 Lampiran D. | Catatan insiden |
| Pembuangan | Hapus aman sebelum dijual, dibuang, atau diperbaiki pihak ketiga. | Catatan |

---

## Lampiran B — Opsi Kendali Endpoint untuk Tim Kecil (usulan)

Biaya dan fitur layanan pihak ketiga **perlu diverifikasi** langsung ke penyedia; angka tidak dicantumkan.

| Opsi | Cara kerja | Kelebihan | Kekurangan |
| :--- | :--- | :--- | :--- |
| 1. Swa-lapor berkala | Personel mengisi checklist (A.6.7 Lampiran E) dengan tangkapan layar tiap [triwulan] | Tanpa alat atau biaya; cepat | Bergantung kejujuran; tidak ada pemantauan terus-menerus |
| 2. MDM ringan dari suite identitas/email perusahaan (bila tersedia) | Pendaftaran perangkat, kebijakan sandi dan enkripsi, wipe jarak jauh | Wipe jarak jauh; laporan terpusat | Perlu verifikasi apakah paket yang dipakai menyediakannya; biaya |
| 3. MDM khusus (Intune, Jamf, atau sejenis) | Kebijakan baseline, kepatuhan, wipe | Kontrol dan laporan penuh | Biaya dan beban operasional untuk tim kecil |
| 4. EDR | Deteksi dan respons pada perangkat | Visibilitas ancaman | Biaya; perlu kapasitas pemantauan |
| 5. Perangkat milik perusahaan | AuditSphere menyediakan laptop standar | Kontrol paling mudah | Biaya; tidak cocok untuk kontraktor |

Rekomendasi bertahap untuk keputusan manusia: mulai dari Opsi 1 (tindakan 3 dan 4) dalam 1 bulan, evaluasi Opsi 2 bila suite perusahaan menyediakannya, dan pertimbangkan Opsi 3 atau 4 sebelum klien jasa keuangan pertama atau bila jumlah personel bertambah. Ini selaras ZTAR-001 (MDM dan EDR) tetapi tidak melampaui kemampuan tim.

---

## Lampiran C — Acuan Silang

| Topik | Lokasi |
| :--- | :--- |
| Checklist kepatuhan per perangkat, inventaris, laporan endpoint, KPI | A.6.7 Lampiran E, F |
| Perangkat hilang dan offboarding perangkat | A.6.7 Lampiran D |
| Akses istimewa ke produksi (SSH, kunci CI) | A.6.7 Lampiran B |
| Opsi VPN dan akses jarak jauh | A.6.7 Lampiran C |
| Screen lock, meja bersih, layar bersih | A.7.7 |
| Keamanan fisik tempat kerja dan perangkat | A.7.3 |
| Pelatihan perangkat dan phishing | A.6.3 |
| Penggunaan alat AI | AIP-001, A.5.32 |
| Insiden perangkat | A.5.24 |
| Kontrol teknis aplikasi (MFA, sesi) | A.5.16, A.8.5 (draf, dokumen belum ada) |

---

Catatan sinkronisasi checklist: baris A.8.1 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi (PIC "Qiko", relevansi "Tentukan relevansi") dan tidak diubah oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
