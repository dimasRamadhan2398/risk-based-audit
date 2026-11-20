# A.7.7 — Meja Bersih dan Layar Bersih

| Field | Isi |
| :--- | :--- |
| **Tema** | Physical |
| **Judul kontrol** | Meja bersih dan layar bersih |
| **Fokus pemeriksaan** | Informasi dan perangkat tidak ditinggalkan tanpa perlindungan |
| **Contoh bukti objektif** | Policy, spot-check record, screen-lock config, awareness |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Security Officer, HR, dan DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini hampir seluruhnya perilaku dan konfigurasi di perangkat serta tempat kerja personel, sehingga buktinya berada di luar repo: pengaturan kunci layar laptop (waktu idle, kata sandi saat bangun), kondisi meja di kantor/co-working/rumah, printer dan pemusnahan kertas, filter privasi layar, berkas unduhan di laptop, isi `.env` lokal, dan konfigurasi alat AI di laptop. Semuanya ditandai **Perlu verifikasi**; tidak ada yang dapat dibuktikan dari repo. Yang dapat dibuktikan dari repo hanyalah sisi aplikasi dan skrip (masa token, pencabutan sesi saat logout, berkas terautentikasi, soket SSH sementara). Seluruh isian pada Lampiran A sampai G (kebijakan, baseline kunci layar, formulir spot-check, poin awareness, adendum kerja jarak jauh/rumah, RACI) adalah **draf usulan**: belum disetujui manajemen, belum ada salinan terisi, dan angka (menit idle, menit meninggalkan meja, frekuensi spot-check, jam pelaporan) bukan hasil pengukuran. Nilai rahasia, kata sandi, kunci, dan alamat IP tidak dicantumkan (kredensial demo bawaan di panduan demo sengaja tidak disalin).

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet, baris ini saat ini berisi relevansi "Tentukan relevansi", status "Belum dinilai", dan jadwal "Belum dijadwalkan"; PIC pada spreadsheet tercatat Qiko, sedangkan dokumen Annex lain memakai peran (Lead Information Security), sehingga dokumen ini mengikuti konvensi peran dan nama PIC spreadsheet dibiarkan sebagai keputusan pengguna.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Personel bekerja jarak jauh dengan laptop yang memegang kunci SSH, token, `.env`, unduhan KKA/LHA/ATR, dan sesi aplikasi yang terbuka; satu layar atau meja yang ditinggalkan tanpa pengaman sudah cukup membuka data klien lembaga jasa keuangan. Pengecualian "tidak berlaku" tidak disarankan karena justru di tempat kerja jarak jauh risikonya paling besar (A.7.3 bagian 1 butir 3 dan 8; A.6.7 bagian 4 butir 15). |
| Implementasi status | **Sebagian** | Ada kendali teknis pendukung di aplikasi dan skrip (bagian 2.1 dan 2.2): JWT 24 jam, sesi dicek ke Redis dan dicabut saat logout, berkas `/uploads/*` dan bukti ATR di belakang token, `Cache-Control: no-store` pada unduhan ATR, soket SSH deploy dibersihkan lewat `trap`. Tidak ada kebijakan clear desk/clear screen yang berdiri sendiri, tidak ada baseline kunci layar, tidak ada idle timeout di aplikasi, tidak ada spot-check, tidak ada awareness. |
| Ketersediaan bukti | **Ada sebagian (hanya bukti kode)** | Policy: belum ada (draf A.7.3 Lampiran A bagian 5 dan A.6.7 RWP-001 butir 2.4/4.2 ada sebagai draf, belum disetujui). Spot-check record: belum ada. Screen-lock config: belum ada (tangkapan layar per perangkat perlu dikumpulkan). Awareness: belum ada (A.6.3). |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A7.7-01 s/d EV-A7.7-10) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada catatan inspeksi, spot-check, atau pengukuran kepatuhan kunci layar. Kendali aplikasi ada tetapi punya celah terhadap layar tak terjaga: tidak ada idle timeout, cookie sesi 1 hari (30 hari bila "ingat saya"), perangkat tepercaya berlaku 90 hari (bagian 4). |
| Desain kontrol | **Belum dirancang** | Draf CDS-001 (usulan kode), baseline kunci layar, formulir spot-check, dan adendum rumah diusulkan di Lampiran A sampai G; belum disetujui. Idle timeout aplikasi belum dirancang. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Sedang** (naik ke Tinggi untuk personel dengan akses produksi/SSH) | Sama dengan A.7.3: tidak ada kantor tetap yang terbukti, risiko berpusat pada laptop dan rumah. Alasan tidak "Tinggi" penuh: kendali teknis utama (MFA, token, sesi server) sudah ada. Alasan tidak "Rendah": satu laptop terbuka dapat membawa SSH root ke produksi, `.env` dan unduhan klien. |
| Review berikutnya | [isi] | Usulan: spot-check pertama dalam 30 hari setelah kebijakan disetujui, lalu tiap 6 bulan; tinjauan desain 12 bulan setelah kebijakan disetujui; sinkron dengan audit internal ISMS tahunan. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.7.7 **berlaku** untuk AuditSphere karena:

1. **Data yang dilindungi bernilai tinggi.** KKA, LHA, ATR beserta berkas bukti klien lembaga jasa keuangan, serta kredensial dan konfigurasi produksi, tersedia dari layar aplikasi dan dari berkas unduhan. Klasifikasi sensitivitasnya mengikuti A.5.12 (data audit klien: *Confidential / Restricted*; kredensial: *Restricted*), meskipun penandaan itu belum diterapkan di sistem (A.5.12 bagian 4).
2. **Tidak ada kantor yang terbukti; "meja" adalah meja rumah, co-working, atau tempat umum.** A.7.3 bagian 1 butir 3 mencatat bahwa status kantor fisik tidak dapat diverifikasi dari repo, dan A.6.7 bagian 1 butir 1 mencatat seluruh personel bekerja dari lokasi masing-masing. Maka clear desk/clear screen harus dirancang untuk tempat kerja jarak jauh, bukan hanya kantor.
3. **Laptop memegang akses langsung ke produksi.** Skrip deploy operator membuka multiplexing SSH (`ControlMaster`) yang tetap hidup 120 detik setelah perintah terakhir, dan memakai kunci pribadi di laptop (`backend/deploy-prod.sh:79-88`). Laptop yang terbuka dan tak terjaga pada jendela itu, atau yang tidak mengunci layar, memberi akses root tanpa autentikasi tambahan.
4. **Aplikasi menampilkan dan mengeluarkan data penuh.** UI tidak memiliki penyamaran (masking) atau watermark data sensitif; ada pratinjau cetak LHA (`window.print()`), ekspor, dan pembukaan berkas bukti di tab baru (bagian 2.2). Layar yang terlihat orang lain atau rapat yang berbagi layar langsung membuka data tenant.
5. **Berkas lokal di luar kendali aplikasi.** Unduhan KKA/LHA/ATR, ekspor Excel, `.env`, dump basis data, dan hasil agen AI tersimpan di laptop. A.6.7 RWP-001 butir 4.3 melarang data klien disimpan lokal tetapi belum berlaku dan tidak ada kendali teknis.
6. **Agen AI berjalan di laptop.** Agen dapat membaca berkas lokal termasuk `.env`; `.claude/settings.json` di repo hanya memuat kunci `permissions` dan tidak ada aturan `deny` untuk berkas rahasia (diperiksa 19 Nov 2026). Layar dan sesi terminal yang terbuka juga memperluas apa yang dapat dilihat agen atau orang lain (A.6.7 bagian 4 butir 12).
7. **Lingkungan demo, pelatihan, dan klien.** Demo dan pelatihan menampilkan data yang tampak nyata kepada audiens; panduan demo mendokumentasikan kredensial bawaan yang lemah (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:683-686`, nilai tidak disalin). Lingkungan pelatihan terpisah (`training.auditsphere.app`) berada di luar repo dan dikelola lewat skrip server (**Perlu verifikasi**). Layar bersama di sesi demo/pelatihan adalah titik paparan.
8. **Pelaksana kontrol adalah personel.** Kontrol ini hanya efektif bila ada kebijakan, kebiasaan, dan bukti pemeriksaan; karena itu cakupan A.6.3 (awareness) dan A.5.36 (kepatuhan) bergantung padanya.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Kendali teknis pendukung layar bersih (sisi aplikasi dan backend, terverifikasi dari repo)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Masa berlaku token | JWT berlaku 24 jam (default `jwt.expiry_hour`); kedaluwarsa dihitung saat login. | `backend/auth-service/pkg/config/config.go:132`, `:206`; `backend/auth-service/services/auth/auth_service.go:251` | Layar yang ditinggal tetap sah hingga 24 jam. Tidak ada refresh token yang dipakai aplikasi untuk memperpendek sesi (repositori `refresh_token_repository.go` ada, tetapi tidak ditemukan penggunaan di alur frontend; **perlu verifikasi** pemakaiannya). |
| Sesi server dan pencabutan saat logout | Setiap token juga disimpan di Redis `token:<jwt>`; middleware auth gagal-tutup bila sesi tidak ditemukan atau Redis tidak tersedia. Logout menghapus kuncinya. | `backend/auth-service/services/auth/auth_service.go:252`; `backend/auth-service/pkg/middleware/auth.go:83-91`; `backend/auth-service/cmd/auth/serve.go:173`, `:267-276` | Logout hanya terjadi bila pengguna menekan logout. Tidak ada pencabutan otomatis karena idle. Pencabutan sesi saat akun dinonaktifkan tidak diverifikasi ulang di sini (A.6.7 bagian 4 butir 8 menyatakan belum). |
| Penanganan 401 di klien | Respons 401 pada API sesi memanggil `_clearState()` dan mengarahkan ke `/auth/login`. | `frontend/plugins/api-auth.client.ts:69-76` | Baru bereaksi pada request berikutnya; layar yang tampak tidak dikosongkan sebelum ada interaksi atau polling. |
| Cookie sesi | Token disimpan di cookie `auth-token` lewat `useCookie` dengan `maxAge` 1 hari, atau 30 hari bila "ingat saya". | `frontend/stores/auth.ts:334`, `:357-361` | Cookie 30 hari melampaui JWT 24 jam; cookie diset dari klien (tidak `HttpOnly`). Berkas profil browser di laptop yang hilang/terbuka menyimpan sesi. |
| MFA dan perangkat tepercaya | MFA TOTP opsional; token MFA sementara berlaku 10 menit; perangkat tepercaya berlaku 90 hari dan melewati MFA. | `auth_service.go:188`; `backend/auth-service/services/trusted-devices/trusted_devices_service.go:93`; `frontend/pages/settings/devices.vue` | Laptop terbuka yang berstatus perangkat tepercaya tidak meminta MFA selama 90 hari. Menegaskan perlunya kunci layar perangkat sebagai kendali utama. Nilai 90 hari ini melengkapi A.6.7 bagian 2.2 yang menyatakan "perlu verifikasi". |
| Login | Kolom kata sandi memakai `type="password"` dengan tombol tampil/sembunyi. | `frontend/pages/auth/login.vue:106`, `:216` | Hanya untuk kolom login; tidak ada kunci layar aplikasi (lock screen) terpisah dari logout. |
| Berkas terautentikasi | `/uploads/*` memerlukan autentikasi dan menolak peran VIEWER; berkas diambil dengan header `Authorization` lalu ditampilkan lewat blob URL yang dicabut setelah 60 detik. | `backend/audit-service/routes/uploads.go:13-37`; `backend/audit-service/pkg/middleware/viewer.go:81`; `frontend/utils/authedFile.ts:1-60`; `frontend/plugins/api-auth.client.ts:39-44` | Ini memperbarui catatan "`/uploads/*` publik" pada A.6.7/A.5.36 untuk jalur audit-service; periksa ulang sebelum dikutip (lihat bagian 4 butir 11). Setelah pengguna menyimpan berkas ke disk, perlindungan aplikasi berakhir. |
| Unduhan bukti ATR tanpa cache | Bukti ATR dilayani sebagai `attachment`, `nosniff`, dan `Cache-Control: private, no-store`. Dokumen IC assessment dan POF juga `no-store`. | `backend/audit-service/controllers/action_taken_report/evidence.go:290-298`; `backend/audit-service/controllers/ic_assessment/handler.go:474-476`; `.../pof_handler.go:431-433` | `no-store` mencegah cache peramban/proxy, bukan salinan yang disimpan pengguna di folder Unduhan. |
| Ekspor dibatasi peran | `canExport` bernilai salah untuk VIEWER. | `frontend/composables/useRbac.ts:205`; contoh pemakaian `frontend/pages/kpi-performance/index.vue:17`, `:142` | Pembatasan di UI; ekspor hanya menyembunyikan tombol dan tidak ada watermark/penandaan pada berkas hasil ekspor. |
| Cetak LHA | Pratinjau cetak LHA menyembunyikan toolbar dan overlay saat cetak; KPI performance punya gaya `@media print`. | `frontend/components/audit-result-report/ResultReportPrintModal.vue:319-324`; `frontend/pages/kpi-performance/index.vue:212` | Tidak ada watermark, header klasifikasi, atau nama pencetak pada hasil cetak (pencarian `watermark` 19 Nov 2026 hanya menemukan fitur *watermark column* sinkronisasi data di `frontend/components/settings/data-source.vue`, bukan watermark dokumen). Cetak tidak dicatat. |

### 2.2 Hal yang diverifikasi tidak ada di aplikasi

| Yang dicari | Hasil (19 Nov 2026) |
| :--- | :--- |
| Idle timeout / auto-logout / peringatan sesi habis | **Tidak ditemukan.** Pencarian `idle`, `inactiv`, `auto-logout`, `session timeout`, `useIdle`, `visibilitychange` pada `frontend/` (di luar `node_modules`, `.nuxt`) hanya menemukan filter data "inactive" pada uji dan `visibilitychange` di `frontend/composables/useSystemLogs.ts:285-297` yang dipakai untuk jeda polling log, bukan untuk mengunci layar. Konfigurasi backend `max_idle_*` (`config.go:119-120`) adalah pool koneksi basis data, bukan sesi pengguna. |
| Layar kunci aplikasi (lock screen) selain logout | Tidak ditemukan. |
| Masking data sensitif di UI (NIK, nomor akun, email, telepon) | Tidak ditemukan pola penyamaran di komponen (pencarian pola `mask`/`****` selain kolom kata sandi). Data tampil penuh sesuai peran. |
| Watermark pada layar, hasil cetak, atau berkas ekspor | Tidak ditemukan (lihat 2.1). |
| Pencatatan cetak/ekspor/unduhan berkas sebagai peristiwa audit | **Perlu verifikasi.** Tidak ditemukan pemanggilan pencatatan di `ResultReportPrintModal.vue`; pencatatan sisi backend untuk unduhan tidak diperiksa menyeluruh. |

### 2.3 Hal yang ada di skrip dan repo terkait perangkat personel

| Objek | Yang terlihat | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Soket SSH deploy sementara | `ControlMaster=auto`, `ControlPath=/tmp/auditsphere-ssh-%C`, `ControlPersist=120`; folder kerja `mktemp -d`; fungsi `cleanup` dengan `trap ... EXIT` menghapus folder dan menutup soket (`ssh -O exit`). Skrip `deploy-prod-serverbuild.sh` memakai opsi yang sama. | `backend/deploy-prod.sh:79-88`; `backend/deploy-prod-serverbuild.sh:80` | `trap EXIT` tidak berjalan bila proses di-`kill -9` atau laptop mati mendadak; soket `/tmp` baru kedaluwarsa oleh `ControlPersist` (120 detik) dan `%C` adalah hash, bukan nama pengguna. `/tmp` di macOS dapat dibaca pengguna lokal lain hanya bila izin soket mengizinkan; izin soket **perlu verifikasi**. Laptop yang dibiarkan terbuka dalam jendela itu memberi sesi SSH yang sudah terautentikasi. |
| Berkas rahasia di laptop | `.env`, `.env.local`, `backend/tenants/*/.env` di `.gitignore:9-16`. Di laptop yang diperiksa ada `backend/.env` dan `frontend/.env` (tidak dilacak git; isi tidak dibaca). `frontend/.env.production` dilacak git tetapi hanya berisi nama variabel URL dan mode (12 nama variabel, diperiksa tanpa nilai). | `.gitignore:9-16`; `git ls-files`; pencarian nama variabel 19 Nov 2026 | `.gitignore` mencegah komit, bukan keberadaan di disk. Berkas `.env` di laptop adalah sasaran saat layar terbuka atau laptop hilang. Isi `.env` lokal dan apakah berisi nilai produksi: **Perlu verifikasi** (A.6.7 RWP-001 butir 6.3). |
| Pembatas agen AI | `.claude/settings.json` hanya memuat `permissions`, tanpa `deny` pada `.env`/kunci. | `.claude/settings.json` (diperiksa 19 Nov 2026) | Konfigurasi alat AI personel lain dan perangkat kontraktor tidak terlihat. AIP-001 melarang kredensial ke AI publik tetapi tidak mengatur akses berkas lokal (A.6.7 bagian 4 butir 12). |
| Kredensial demo bawaan | Panduan demo mendokumentasikan akun dan kata sandi bawaan lingkungan demo. | `docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:683-686` | Nilai tidak disalin. Berisiko bila layar demo/rekaman memperlihatkannya atau bila instans demo dapat diakses publik; kondisi nyata **perlu verifikasi** (`auditsphere-devops`). |

### 2.4 Dokumen yang menyinggung (belum berlaku formal)

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| A.7.3 Lampiran A bagian 5 dan 6 (PHY-001 usulan) | Kunci layar otomatis maksimum [5] menit; meja bersih bila meninggalkan meja lebih dari [15] menit atau akhir hari; papan tulis dibersihkan; cetak dibatasi, pencacah kertas, PIN printer, media removable. | Draf; belum disetujui. Dokumen ini merinci dan mengoperasionalkan bagian tersebut, bukan menggantikannya. |
| A.6.7 Lampiran A (RWP-001 usulan) | Butir 2.4 kunci layar [5] menit; 4.1 privasi layar; 4.2 perangkat dikunci; 4.3 data klien tidak disimpan lokal; 4.4 tidak mencetak; 4.6 tidak membagikan layar yang menampilkan kredensial; 6.3 `.env` hanya nilai pengembangan; 6.5 batas agen AI. | Draf; belum disetujui. |
| A.5.12 Klasifikasi Informasi | Skema klasifikasi (*Public, Internal, Confidential, Restricted*) yang menentukan perlakuan layar/cetak. | Belum disetujui dan belum diterapkan di sistem. |
| A.6.3 Awareness | Program awareness dan topik T17 (kerja jarak jauh) yang diusulkan; usulan topik T18 pada Lampiran E dokumen ini. | Draf; belum ada materi atau catatan kehadiran. |
| A.5.32 dan AIP-001 | Penggunaan agen AI dan larangan kredensial/data klien ke AI publik. | AIP-001 tidak memuat ketentuan berkas lokal atau layar. |
| CSP-001 | Kebijakan keamanan cloud dengan MFA konsol dan tinjauan akses. | Templat generik; tidak memuat clear desk/clear screen. |
| A.7.3 bagian 4 butir 8 | Menyatakan belum ada kebijakan clear desk/clear screen yang berdiri sendiri. | Tetap benar per 19 Nov 2026 (tidak ada berkas `docs/` yang memuatnya; pencarian `clear desk`, `meja bersih`, `screen lock`, `clear screen` pada `docs/` hanya menemukan referensi pada A.7.3, A.6.7, dan dokumen ini). |

### 2.5 Hal yang diverifikasi **tidak ada** di repo

Kebijakan clear desk/clear screen yang disetujui; baseline kunci layar atau konfigurasi MDM/profil perangkat; formulir atau catatan spot-check; catatan inspeksi meja; kebijakan cetak dan pemusnahan kertas beserta pencacah; filter privasi layar; kebijakan berbagi layar untuk demo dan pelatihan; ketentuan layar saat rapat dengan klien; pelatihan atau acknowledgement khusus clear desk/clear screen. Kondisi nyata di perangkat dan tempat kerja personel **Perlu verifikasi**.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Policy clear desk dan clear screen (disetujui) | **Belum tersedia** | Draf CDS-001 pada Lampiran A (merujuk PHY-001 bagian 5 dan RWP-001). Perlu tanda tangan. |
| Screen-lock config per perangkat (tangkapan layar: waktu idle, kata sandi saat bangun) | **Belum tersedia / Perlu verifikasi** | Baseline usulan Lampiran B; tangkapan layar bertanggal per perangkat dikumpulkan lewat Lampiran C. Tidak ada MDM. |
| Konfigurasi sesi aplikasi (masa token, idle timeout) | **Ada sebagian** | Masa token 24 jam dan cookie ada di kode (bagian 2.1). Idle timeout tidak ada. |
| Spot-check record | **Belum tersedia** | Formulir usulan Lampiran C; belum ada pelaksanaan. |
| Bukti awareness (materi, daftar hadir, quiz) | **Belum tersedia** | Poin awareness usulan Lampiran E dan topik T18 (A.6.3). |
| Acknowledgement personel | **Belum tersedia** | Dapat digabung dengan RWP-001 (A.6.7 Lampiran H) dan PHY-001 (A.7.3 Lampiran A). |
| Bukti kondisi tempat kerja rumah/co-working | **Belum tersedia / Perlu verifikasi** | Swa-lapor A.7.3 Lampiran E.2 dan adendum Lampiran D. |
| Bukti cetak dan pemusnahan (pencacah, log) | **Belum tersedia / Perlu verifikasi** | Tidak ada kebijakan maupun peralatan yang terbukti. |
| Inventaris perangkat | **Belum tersedia** | A.6.7 tindakan 8 dan Lampiran F. |
| Bukti pembersihan lingkungan demo/pelatihan (data nyata tidak tampil, kredensial bawaan diganti) | **Belum tersedia / Perlu verifikasi** | Daftar periksa demo Lampiran F; dikerjakan `auditsphere-devops`. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada kebijakan clear desk/clear screen yang berdiri sendiri maupun yang disetujui.** Hanya draf bagian dalam PHY-001 (A.7.3) dan RWP-001 (A.6.7). Tindakan 1.
2. **Tidak ada idle timeout atau auto-logout di aplikasi.** Sesi hidup hingga 24 jam (JWT) dan cookie 1 atau 30 hari, tanpa kunci layar aplikasi atau peringatan (bagian 2.1 dan 2.2). Selama laptop tidak terkunci, sesi aplikasi ikut terbuka penuh. Perlu tiket ke tim frontend dan backend (tindakan 6).
3. **Perangkat tepercaya 90 hari dan "ingat saya" 30 hari memperlebar jendela paparan.** MFA dilewati pada perangkat tepercaya (`trusted_devices_service.go:93`), dan cookie bertahan 30 hari (`auth.ts:334`). Bila laptop terbuka, tidak ada MFA yang menghalangi. Kunci layar perangkat adalah kendali utama; belum ada bukti penerapannya.
4. **Tidak ada baseline kunci layar yang terukur.** Angka [5] menit pada draf PHY-001/RWP-001 belum disetujui dan tidak ada tangkapan layar perangkat satu pun. Tidak ada MDM (A.6.7 bagian 4 butir 2). Status kunci layar di setiap laptop: **Perlu verifikasi**.
5. **UI tidak menyamarkan atau memberi watermark data sensitif.** Data tampil penuh; hasil cetak/ekspor tidak memuat penanda klasifikasi, nama pengguna, atau tanggal (bagian 2.1 dan 2.2). Terkait A.5.12 yang belum diterapkan. Rapat dengan klien dan berbagi layar membuka data tenant tanpa penghalang aplikasi.
6. **Unduhan dan ekspor menghasilkan salinan lokal tak terkendali.** Setelah `saveBlobAsFile` (`frontend/utils/authedFile.ts`), berkas LHA/KKA/ATR berada di folder Unduhan tanpa masa simpan dan tanpa penandaan. `no-store` hanya mengatur cache. Tidak ada ketentuan hapus setelah selesai kecuali pada draf RWP-001 butir 4.3. Perlu kebiasaan hapus (Lampiran A bagian 6) dan, bila layak, penandaan berkas (tindakan 6).
7. **Soket SSH deploy bertahan di `/tmp` selama 120 detik dan sesi SSH dapat dipakai siapa pun yang menguasai laptop terbuka pada jendela itu.** `cleanup` lewat `trap` menutup soket pada akhir normal (`backend/deploy-prod.sh:84-88`) dan folder kerja memakai `mktemp -d`. Risiko residu rendah bila skrip selesai normal; berlaku lebih tinggi bila proses terputus. Kunci pribadi di laptop tetap bergantung pada kunci layar. Izin soket di `/tmp`: **Perlu verifikasi**. Usulan: operator mengunci layar atau menutup soket manual (`ssh -O exit`) bila meninggalkan laptop saat deploy (Lampiran A bagian 7).
8. **`.env` dan rahasia di laptop.** `.env` ada secara lokal pada laptop yang diperiksa (backend dan frontend); isinya tidak dibaca. Selama laptop terbuka atau dikuasai agen AI, berkas ini terbaca. Kebijakan "hanya nilai pengembangan" belum berlaku (RWP-001 butir 6.3). Tindakan 4 dan 7.
9. **Agen AI dan layar terbuka.** Tidak ada aturan `deny` pada `.claude/settings.json` untuk `.env`/kunci; ketentuan akses berkas lokal oleh agen belum ada di AIP-001 (A.6.7 tindakan 10). Terminal yang menampilkan nilai rahasia, atau sesi agen yang berjalan tanpa pengawasan, adalah paparan layar sekaligus berkas.
10. **Lingkungan demo dan pelatihan.** Panduan demo mendokumentasikan kredensial bawaan lemah (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md:683-686`). Lingkungan pelatihan adalah klon produksi terisolasi di luar repo, dikelola lewat skrip server (**Perlu verifikasi**; tidak ada dokumen di `docs/`). Perlu aturan: data nyata klien tidak tampil pada demo/pelatihan, layar dan rekaman tidak memperlihatkan kredensial atau `.env`, dan akun bawaan diganti sebelum instans dapat dijangkau pihak luar.
11. **Koreksi silang terhadap dokumen Annex lain.** Per 19 Nov 2026 audit-service melayani `/uploads/*` di belakang autentikasi (`routes/uploads.go:13-37`) dan auth-service mengecek sesi di Redis dengan gagal-tutup (`pkg/middleware/auth.go:83-91`). Catatan "`/uploads/*` publik" dan beberapa pernyataan sesi pada A.6.7 dan A.5.36 (31 Okt 2026) kemungkinan sudah usang untuk jalur tersebut; tinjau dan perbarui. Ini tidak mengubah kesimpulan A.7.7 karena kendali itu tidak melindungi layar yang terbuka.
12. **Cetak dan pemusnahan.** Aplikasi menyediakan `window.print()` pada LHA (`ResultReportPrintModal.vue:319`) tanpa penanda atau pencatatan. Tidak ada kebijakan cetak, printer PIN, atau pencacah kertas yang terbukti. Kebijakan "jangan mencetak" (RWP-001 butir 4.4) belum berlaku.
13. **Tidak ada spot-check dan tidak ada awareness.** Tidak ada mekanisme verifikasi kepatuhan meja/layar dan tidak ada materi pelatihan. Tindakan 8 dan 9.
14. **Kondisi fisik di luar repo.** Filter privasi layar, kunci pintu ruang kerja rumah, tamu atau anggota keluarga di dekat perangkat, laci/lemari terkunci, ruang kerja co-working, dan kondisi meja kantor (bila ada) **Perlu verifikasi** dan hanya dapat dibuktikan lewat swa-lapor, foto tanpa data sensitif, atau inspeksi.
15. **Status dokumen.** Seperti A.5.12 s/d A.7.3, dokumen ini belum ter-commit. Tidak ada perubahan kode atau konfigurasi dilakukan. Checklist xlsx tidak diubah. Tidak ada berkas indeks Annex yang ditemukan untuk ditambahi: `docs/DOKUMEN_KESELARASAN_ISO27001.md` tidak merujuk satu pun dokumen `A.x.y` per 19 Nov 2026, sehingga tidak ada entri yang ditambahkan.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Setujui **CDS-001 Kebijakan Meja Bersih dan Layar Bersih** (Lampiran A) atau tetapkan sebagai bagian 5 PHY-001 dan butir RWP-001; tetapkan angka (menit idle, menit meninggalkan meja, jam pelaporan). **Keputusan manusia diperlukan.** | Lead Information Security dengan Legal & Compliance dan HR | Tinggi | [isi] | Kebijakan bertanda tangan |
| 2 | Tetapkan **baseline kunci layar** (Lampiran B) dan kumpulkan **tangkapan layar pengaturan** dari setiap laptop (Lampiran C bagian C.1). | Security Officer, seluruh personel | Tinggi | [isi] | Tangkapan layar bertanggal per perangkat; inventaris (A.6.7 Lampiran F) |
| 3 | Gabungkan **acknowledgement** CDS-001 dengan RWP-001 dan PHY-001 (satu formulir untuk ketiganya). | Lead Information Security, HR | Sedang | [isi] | Acknowledgement tertanda |
| 4 | Pastikan **`.env`, kunci, dan unduhan klien** tidak ada di laptop kecuali yang disetujui; hapus berkas yang tidak perlu; batasi agen AI dari `.env` dan kunci (tambahkan aturan `deny` pada konfigurasi agen, selaras A.6.7 tindakan 10). | Lead Information Security, Security Officer | Tinggi | [isi] | Hasil swa-periksa (Lampiran C bagian C.1 butir 8-10); konfigurasi agen |
| 5 | Lakukan **spot-check pertama** dengan formulir Lampiran C (sampel personel dan sesi demo/pelatihan) dan simpan catatan. | Security Officer | Sedang | [isi] | Formulir terisi; masuk monitoring report A.5.36 |
| 6 | Ajukan **tiket aplikasi** (di luar dokumen ini): idle timeout dan auto-logout dengan peringatan [15 menit usulan], kunci layar aplikasi, pencabutan sesi saat idle, cookie sesi yang lebih pendek untuk peran ADMIN/AUDITOR, masa perangkat tepercaya yang lebih pendek, penandaan (nama pengguna, tanggal, klasifikasi) pada cetak dan ekspor, opsi penyamaran bidang sensitif, pencatatan cetak/ekspor/unduhan. | Tim Backend dan Tim Frontend bersama `auditsphere-solution-architect` | Sedang | [isi] | Tiket, rancangan, perubahan kode dan uji (dikerjakan `auditsphere-frontend`/`auditsphere-backend`) |
| 7 | Tambahkan **ketentuan operator deploy**: kunci layar sebelum meninggalkan laptop saat deploy; tutup soket SSH manual bila skrip terputus; tidak mengunduh dump produksi ke laptop. Verifikasi izin soket di `/tmp` pada macOS/Linux operator. | `auditsphere-devops`, Lead Information Security | Sedang | [isi] | Klausul Lampiran A bagian 7; catatan verifikasi izin |
| 8 | Masukkan **topik T18 Meja bersih dan layar bersih** ke program awareness dan induksi (Lampiran E), termasuk demo, rapat dengan klien, dan berbagi layar. | Security Officer | Sedang | [isi] | Materi, daftar hadir, quiz (A.6.3) |
| 9 | Siapkan **daftar periksa sesi demo/pelatihan** (Lampiran F) dan ganti kredensial bawaan sebelum instans dapat dijangkau pihak luar. Perubahan pada server melalui `auditsphere-devops`. | `auditsphere-devops`, Project Lead | Sedang | [isi] | Daftar periksa terisi; catatan perubahan |
| 10 | Putuskan **kebijakan cetak dan pemusnahan** (cetak seperlunya, pencacah kertas bila mencetak, tanpa printer publik) selaras PHY-001 bagian 6. | Lead Information Security | Rendah | [isi] | Ketentuan di CDS-001 |
| 11 | **Perbarui dokumen yang usang**: perbaiki catatan `/uploads/*` publik dan sesi tanpa pengecekan pada A.6.7 dan A.5.36 setelah diverifikasi ulang. | Project Lead | Rendah | [isi] | Revisi dokumen |
| 12 | Jadwalkan **spot-check berkala** dan masukkan temuan serta pengecualian ke exception register A.5.36. | Security Officer | Sedang | [isi] | Jadwal; register |
| 13 | **Daftarkan Evidence ID** setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |
| 14 | Commit dokumen ini bersama annex lain yang belum ter-commit (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |

Catatan pelaksanaan: tindakan 6 dan 9 menyentuh aplikasi atau server. Perubahan aplikasi melalui agen frontend/backend; perubahan server melalui `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh --ref`, tanpa seeder), satu perubahan per sesi, dan tanpa `docker compose up` seluruh stack di VPS.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A7.7-01 | Masa token, sesi Redis, dan pencabutan saat logout | `backend/auth-service/pkg/config/config.go`, `services/auth/auth_service.go`, `pkg/middleware/auth.go`, `cmd/auth/serve.go` | Internal |
| EV-A7.7-02 | Penyimpanan token di cookie dan penanganan 401 di klien | `frontend/stores/auth.ts`, `frontend/plugins/api-auth.client.ts` | Internal |
| EV-A7.7-03 | MFA sementara dan perangkat tepercaya 90 hari | `backend/auth-service/services/trusted-devices/trusted_devices_service.go`, `frontend/pages/settings/devices.vue` | Internal |
| EV-A7.7-04 | Berkas terautentikasi, unduhan tanpa cache, pembatasan VIEWER | `backend/audit-service/routes/uploads.go`, `controllers/action_taken_report/evidence.go`, `frontend/utils/authedFile.ts`, `frontend/composables/useRbac.ts` | Internal |
| EV-A7.7-05 | Cetak LHA tanpa watermark (bukti ketiadaan penanda) | `frontend/components/audit-result-report/ResultReportPrintModal.vue` | Internal |
| EV-A7.7-06 | Soket SSH deploy sementara dan pembersihan | `backend/deploy-prod.sh`, `backend/deploy-prod-serverbuild.sh` | Internal |
| EV-A7.7-07 | Draf kebijakan, baseline, formulir, awareness, adendum rumah, RACI (setelah tindakan 1 disetujui) | `docs/ISO_27001/A.7.7_Meja_Bersih_dan_Layar_Bersih.md` (Lampiran A sampai G) | Internal |
| EV-A7.7-08 | Tangkapan layar kunci layar per perangkat dan inventaris (setelah tindakan 2) | [belum ada] | Restricted (data per orang) |
| EV-A7.7-09 | Formulir spot-check terisi dan ringkasan (setelah tindakan 5 dan 12) | [belum ada] | Restricted |
| EV-A7.7-10 | Materi T18, daftar hadir, quiz, acknowledgement; daftar periksa demo/pelatihan terisi (setelah tindakan 3, 8, 9) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 5 selesai, tambahkan ID per siklus spot-check.

---

## Lampiran A — Draf Kebijakan Meja Bersih dan Layar Bersih (CDS-001, usulan, siap tanda tangan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Nilai pada kurung siku adalah usulan awal, bukan hasil pengukuran; ubah sebelum pengesahan. CDS-001 mengoperasionalkan PHY-001 bagian 5 (A.7.3 Lampiran A) dan RWP-001 butir 2.4 dan 4 (A.6.7 Lampiran A); bila terjadi perbedaan angka, selaraskan ketiganya.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | CDS-001 Kebijakan Meja Bersih dan Layar Bersih (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review tahunan dan setelah insiden) |
| Berlaku untuk | Seluruh karyawan, kontraktor, dan kontributor; semua tempat kerja yang dipakai untuk pekerjaan AuditSphere (kantor, co-working, rumah, tempat umum); semua perangkat, media, dokumen cetak, dan layar yang menampilkan informasi AuditSphere atau klien, termasuk sesi demo dan pelatihan |

**1. Tujuan.** Mencegah informasi dan perangkat dilihat, diambil, atau dipakai pihak tidak berwenang karena ditinggalkan tanpa perlindungan. Informasi diperlakukan menurut klasifikasi A.5.12; data audit klien (KKA, LHA, ATR, bukti) dan kredensial diperlakukan sebagai *Confidential/Restricted* sampai skema klasifikasi disetujui.

**2. Layar bersih**

| No | Ketentuan | Catatan |
| :--- | :--- | :--- |
| 2.1 | Kunci layar otomatis dengan kata sandi/biometrik setelah idle maksimum [5] menit pada semua perangkat (Lampiran B). | Selaras RWP-001 butir 2.4 dan PHY-001 butir 5.1. |
| 2.2 | Kunci layar manual setiap kali meninggalkan perangkat, termasuk sebentar (macOS: Ctrl+Cmd+Q; Windows: Win+L; Linux: sesuai lingkungan desktop). | Jangan mengandalkan penghemat layar saja. |
| 2.3 | Logout dari aplikasi AuditSphere dan layanan lain (GitHub, konsol cloud) setelah tugas sensitif selesai; jangan memilih "ingat saya" pada perangkat bersama atau di luar perangkat terdaftar. | Aplikasi belum memiliki idle timeout (bagian 4 butir 2). |
| 2.4 | Tutup soket SSH dan sesi terminal produksi sebelum meninggalkan laptop; saat deploy berlangsung, kunci layar atau tetap di depan perangkat. | `ControlPersist=120` dan `trap` hanya berlaku pada akhir normal. |
| 2.5 | Layar tidak menampilkan data klien atau kredensial di area umum, saat rapat dengan tamu/klien, atau saat berbagi layar. Bagikan satu jendela, bukan seluruh layar; tutup notifikasi dan tab sensitif lebih dahulu. | Bagian 6 untuk demo dan pelatihan. |
| 2.6 | Gunakan filter privasi atau atur posisi bila bekerja di tempat umum; hindari panggilan yang menyebut data klien di tempat umum. | RWP-001 butir 4.1. |
| 2.7 | Notifikasi pratinjau (email, chat) dimatikan atau disembunyikan pada layar terkunci. | |
| 2.8 | Nilai rahasia, `.env`, dan kredensial tidak ditampilkan di terminal, tangkapan layar, rekaman, atau tiket. | RWP-001 butir 6.3 dan 4.6. |

**3. Meja bersih**

| No | Ketentuan |
| :--- | :--- |
| 3.1 | Saat meninggalkan meja lebih dari [15] menit dan di akhir hari kerja, tidak ada dokumen klien, catatan kredensial, token keamanan (kunci keamanan perangkat keras), media removable, atau catatan tempel berisi detail sistem di atas meja; simpan di laci/lemari terkunci (PHY-001 butir 5.2). |
| 3.2 | Perangkat dan token keamanan disimpan di tempat terkunci saat tidak dipakai, atau dibawa serta. Perangkat tidak ditinggalkan di kendaraan atau tempat umum (PHY-001 bagian 7). |
| 3.3 | Papan tulis dan catatan tempel dibersihkan setelah rapat (PHY-001 butir 5.3). |
| 3.4 | Pengiriman fisik yang berisi perangkat/media diterima dan dicatat (PHY-001 butir 3.4). |

**4. Berkas lokal, unduhan, dan ekspor**

| No | Ketentuan |
| :--- | :--- |
| 4.1 | Data klien tidak disimpan di laptop kecuali tugas yang disetujui; unduhan KKA/LHA/ATR, ekspor Excel, dan hasil cetak ke PDF dihapus setelah tugas selesai dan paling lambat akhir minggu kerja [usulan]; kosongkan juga folder Unduhan dan Sampah. |
| 4.2 | Berkas lokal diberi nama yang tidak mengungkap klien bila memungkinkan dan disimpan di folder khusus pada disk terenkripsi (RWP-001 butir 2.2); tidak disalin ke penyimpanan awan pribadi atau layanan berbagi non-perusahaan. |
| 4.3 | Berkas dari aplikasi tidak dikirim lewat chat/email pribadi; gunakan jalur yang disetujui perusahaan. |
| 4.4 | `.env` lokal hanya berisi nilai pengembangan; dump basis data produksi tidak diunduh ke laptop (RWP-001 butir 6.3). |

**5. Cetak dan pemusnahan.** Tidak mencetak dokumen klien kecuali perlu; bila mencetak, ambil segera, simpan terkunci, hancurkan dengan pencacah kertas (potong silang) setelah tidak diperlukan, dan jangan memakai printer bersama/publik (PHY-001 butir 6.1-6.2; RWP-001 butir 4.4). Menyimpan sebagai PDF diperlakukan sebagai berkas lokal (bagian 4).

**6. Demo, pelatihan, dan rapat dengan klien**

| No | Ketentuan |
| :--- | :--- |
| 6.1 | Demo dan pelatihan memakai lingkungan demo/pelatihan dengan data sintetis atau tersamar; data nyata klien tidak ditampilkan. |
| 6.2 | Kredensial bawaan lingkungan demo/pelatihan diganti dan tidak ditampilkan di layar, slide, atau rekaman. |
| 6.3 | Sebelum berbagi layar: tutup aplikasi, tab, dan notifikasi lain; periksa nama jendela; gunakan akun demo, bukan akun produksi. |
| 6.4 | Rekaman sesi tidak memuat kredensial, `.env`, atau data klien; simpan dengan akses terbatas dan hapus setelah [usulan 30 hari] kecuali dibutuhkan. |

**7. Operator produksi dan agen AI.** Saat memakai akses produksi: kunci layar sebelum meninggalkan perangkat, tutup soket SSH (`ssh -O exit`) bila skrip terputus, dan jangan membiarkan sesi agen AI berjalan tanpa pengawasan bila agen memiliki akses ke direktori yang memuat `.env` atau kunci. Agen tidak diberi izin membaca `.env` dan kunci produksi (RWP-001 butir 6.5).

**8. Pelaporan.** Perangkat atau dokumen yang ditinggalkan/hilang, layar terlihat pihak tidak berwenang, atau pelanggaran kebijakan dilaporkan ke Lead Information Security dalam [1] jam sejak diketahui (RWP-001 butir 7; A.6.7 Lampiran D) dan ditangani sebagai insiden (A.5.24) tanpa menghukum pelaporan jujur.

**9. Pemeriksaan, penegakan, dan pengecualian.** Kepatuhan diperiksa lewat spot-check (Lampiran C). Pelanggaran ditangani sesuai kontrak dan kebijakan SDM (di luar repo) dan eskalasi A.5.36 Lampiran D. Pengecualian (mis. perangkat yang belum dapat dikunci otomatis) tertulis, bertenggat, dengan kontrol kompensasi, dan dicatat di exception register A.5.36.

**Blok persetujuan**

| Peran | Nama | Tanda tangan | Tanggal |
| :--- | :--- | :--- | :--- |
| Penyusun (Legal & Compliance / PM) | [isi] | | |
| Pemilik (Lead Information Security) | [isi] | | |
| Penyetuju (Direktur Utama) | [isi] | | |

---

## Lampiran B — Baseline Kunci Layar dan Sesi (usulan)

Semua angka adalah **usulan** tanpa dasar historis; baseline 19 Nov 2026: tidak ada data pengaturan perangkat dan belum ada idle timeout aplikasi.

**B.1 Perangkat personel**

| Pengaturan | Usulan | Cara bukti |
| :--- | :--- | :--- |
| Kunci layar otomatis setelah idle | Maksimum [5] menit (sesi demo yang sengaja panjang: pengecualian tertulis) | Tangkapan layar pengaturan (macOS: Pengaturan Sistem > Kunci Layar; Windows: Pengaturan > Akun > Opsi masuk dan Kunci layar; Linux: pengaturan desktop) |
| Kata sandi/biometrik diminta saat bangun atau setelah layar mati | Segera, atau maksimum [5] detik | Tangkapan layar |
| Penghemat layar/mati layar | Maksimum [5] menit, tidak lebih lama dari kunci layar | Tangkapan layar |
| Kunci manual (pintasan) | Diketahui seluruh personel | Quiz T18 |
| Pratinjau notifikasi di layar terkunci | Dimatikan | Tangkapan layar |
| Enkripsi disk | Aktif (A.6.7 Lampiran E butir 1) | Tangkapan layar |
| Akun pengguna | Kata sandi kuat; tidak ada login otomatis (auto-login) | Tangkapan layar |
| Terminal dan SSH | Soket kontrol ditutup saat tidak dipakai; kunci SSH berpassphrase (A.6.7 Lampiran B) | Tangkapan konfigurasi tanpa nilai kunci |

**B.2 Aplikasi AuditSphere (usulan perubahan; belum ada)**

| Pengaturan | Kondisi 19 Nov 2026 | Usulan |
| :--- | :--- | :--- |
| Masa JWT | 24 jam (`config.go:206`) | Pertimbangkan lebih pendek untuk ADMIN/AUDITOR bersama mekanisme refresh |
| Idle timeout aplikasi | Tidak ada | [15] menit dengan peringatan sebelum logout otomatis; konfigurasi per tenant bila perlu |
| Cookie "ingat saya" | 30 hari (`auth.ts:334`) | Perpendek atau nonaktifkan untuk peran istimewa |
| Perangkat tepercaya | 90 hari (`trusted_devices_service.go:93`) | Perpendek; tinjau berkala (A.6.7 tindakan 12) |
| Penanda hasil cetak/ekspor | Tidak ada | Nama pengguna, tanggal/jam, klasifikasi |
| Penyamaran bidang sensitif | Tidak ada | Opsional per peran |

Keputusan nilai ada pada Lead Information Security dan Direktur Utama; implementasi melalui agen frontend/backend dan uji (`auditsphere-qa`).

---

## Lampiran C — Formulir Swa-Lapor dan Spot-Check (templat)

**C.1 Swa-lapor perangkat dan meja (induksi dan tiap [6] bulan)**

| No | Pemeriksaan | Ya | Tidak | Bukti | Catatan |
| :--- | :--- | :---: | :---: | :--- | :--- |
| 1 | Kunci layar otomatis <= [5] menit dengan kata sandi/biometrik | | | Tangkapan layar | |
| 2 | Kata sandi diminta segera saat bangun | | | Tangkapan layar | |
| 3 | Notifikasi pratinjau dimatikan di layar terkunci | | | Tangkapan layar | |
| 4 | Pintasan kunci manual diketahui dan dipakai | | | Pernyataan | |
| 5 | Tidak ada dokumen klien/catatan kredensial/token/media di meja saat ditinggal | | | Foto tanpa data sensitif | |
| 6 | Perangkat dan token disimpan terkunci atau dibawa saat tidak dipakai | | | Pernyataan/foto | |
| 7 | Folder Unduhan dan Sampah tidak memuat berkas klien lama | | | Tangkapan daftar nama berkas | |
| 8 | Tidak ada `.env` berisi nilai produksi, kunci, atau dump basis data di laptop | | | Pernyataan | |
| 9 | Agen AI di laptop tidak dapat membaca `.env` dan kunci | | | Tangkapan konfigurasi | |
| 10 | Soket SSH tidak aktif saat perangkat ditinggal | | | Pernyataan | |
| 11 | Filter privasi atau posisi layar aman di tempat umum | | | Pernyataan | |
| 12 | Tahu cara melapor perangkat/dokumen hilang | | | Quiz | |

**C.2 Spot-check oleh pemeriksa (kantor, co-working, atau panggilan video untuk rumah)**

| Bagian | Isi |
| :--- | :--- |
| Identitas | Tanggal, jam, lokasi (kantor/rumah/co-working/sesi demo), personel, pemeriksa, metode (kunjungan/panggilan video/pengamatan sesi daring) |
| Pemeriksaan | Layar terkunci pada perangkat yang ditinggalkan; meja bersih; token/media/cetakan terlihat; papan tulis bersih; tempat sampah/pencacah tanpa dokumen klien; tidak ada perangkat tak dikenal; layar tidak terlihat pihak lain |
| Cara uji layar | Pemeriksa mengamati apakah layar terkunci dalam waktu idle yang ditetapkan, bukan hanya membaca pengaturan |
| Hasil | Patuh / Patuh dengan pengecualian (nomor) / Tidak patuh |
| Tindakan | Daftar dengan PIC dan tenggat; tindak lanjut dalam [14] hari |
| Penutup | Tanda tangan pemeriksa dan peninjau (orang berbeda bila memungkinkan); sampel tanpa data sensitif; ringkasan masuk monitoring report A.5.36 dan Evidence Index |

**C.3 Frekuensi usulan:** swa-lapor induksi dan tiap [6] bulan; spot-check acak sampel [minimal 1 personel per triwulan, ditambah satu sesi demo/pelatihan per semester]; kantor (bila ada) mengikuti A.7.3 Lampiran E.3 butir 3 dan 4.

---

## Lampiran D — Adendum Kerja Jarak Jauh dan Rumah (usulan)

Melengkapi A.6.7 Lampiran A bagian 4 dan A.7.3 Lampiran E.2. **Draf**; belum disetujui.

| No | Ketentuan |
| :--- | :--- |
| D.1 | Pilih ruang atau sudut kerja yang layarnya tidak terlihat dari jendela, pintu, atau area lalu lalang; atur posisi layar membelakangi ruang bersama. |
| D.2 | Saat berada di rumah dengan anggota keluarga, tamu, atau teknisi: kunci layar sebelum berdiri, dan jangan biarkan mereka memakai perangkat kerja. |
| D.3 | Bila ruang kerja dapat dikunci, kunci saat ditinggal; bila tidak, simpan perangkat, token keamanan, dan dokumen cetak di laci/lemari terkunci atau bawa serta. |
| D.4 | Panggilan dan rapat yang membahas data klien dilakukan dengan headset; pastikan latar layar dan papan tulis tidak menampilkan data klien saat kamera aktif; pertimbangkan latar belakang yang diburamkan. |
| D.5 | Perangkat pintar dengan mikrofon (speaker, asisten suara) dimatikan atau dijauhkan saat membahas data klien [anjuran]. |
| D.6 | Tidak mencetak di rumah kecuali disetujui; bila mencetak, hancurkan dengan pencacah kertas dan jangan membuang ke sampah rumah. |
| D.7 | Perangkat tidak dibawa ke tempat umum tanpa kebutuhan; di kafe/co-working, jangan meninggalkan perangkat atau titip ke orang lain, bahkan sebentar. |
| D.8 | Dalam perjalanan: perangkat sebagai bagasi kabin, layar dikunci, tidak bekerja dengan data klien di transportasi umum yang ramai. |
| D.9 | Kehilangan, pencurian, atau pembukaan layar oleh pihak lain dilaporkan sesuai A.6.7 Lampiran D (bagian 8 CDS-001). |

Bukti kondisi rumah dikumpulkan lewat swa-lapor (Lampiran C bagian C.1 dan A.7.3 Lampiran E.2); AuditSphere tidak masuk ke rumah personel, sehingga pemeriksaan memakai pernyataan, foto tanpa data sensitif, atau panggilan video dengan persetujuan personel.

---

## Lampiran E — Poin Awareness (topik T18 usulan untuk A.6.3)

Ditambahkan pada matriks A.6.3 Lampiran B; **draf**, belum disetujui.

| Kode | Topik | Peran wajib | Frekuensi | Penilaian |
| :--- | :--- | :--- | :--- | :--- |
| T18 | Meja bersih dan layar bersih: kunci layar, tempat kerja rumah/umum, unduhan dan ekspor, cetak, demo/pelatihan, agen AI | R1 (W); diperdalam untuk R2, R4, R9 | Induksi + tahunan | Quiz >= 80% + pemeriksaan pengaturan perangkat |

**Poin inti materi:**

1. Layar terbuka tanpa kunci adalah akses penuh: aplikasi memberi sesi hingga 24 jam dan perangkat tepercaya melewati MFA hingga 90 hari; kunci layar adalah pagar terakhir.
2. Kunci layar saat berdiri, apa pun alasannya; pintasan keyboard per sistem operasi.
3. Folder Unduhan adalah arsip tersembunyi: hapus berkas klien setelah selesai; jangan menyimpan dump atau ekspor berisi data klien.
4. `.env`, kunci, dan token jangan ditampilkan di layar, tangkapan layar, atau rekaman; jangan biarkan agen AI membaca `.env` atau memegang kredensial produksi.
5. Saat berbagi layar atau demo: bagikan satu jendela, tutup tab/notifikasi, pakai data demo dan akun demo, ganti kredensial bawaan.
6. Meja rumah dan meja kafe sama pentingnya dengan meja kantor; layar tidak boleh terlihat orang lain.
7. Cetak hanya bila perlu; hancurkan dengan pencacah; tidak ada printer publik.
8. Laporkan segera perangkat atau dokumen yang hilang atau ditinggalkan; pelaporan jujur tidak dihukum.
9. Soket SSH deploy tetap terbuka singkat setelah perintah terakhir: kunci layar atau tutup sesi.

Contoh soal quiz (draf): (1) Berapa maksimum menit idle sebelum layar harus terkunci menurut baseline? [5, usulan]; (2) Apa yang harus dilakukan terhadap berkas LHA yang diunduh untuk dikerjakan setelah tugas selesai?; (3) Apa yang harus ditutup sebelum berbagi layar dalam demo?; (4) Ke siapa dan dalam berapa jam melaporkan laptop yang tertinggal di kafe? [Lead Information Security, 1 jam, usulan].

Catatan: pengukuran efektivitas (cakupan induksi, skor quiz, persentase perangkat patuh pada spot-check) mengikuti KPI A.6.3 Lampiran E dan A.6.7 Lampiran F.3; target adalah usulan tanpa data baseline.

---

## Lampiran F — Daftar Periksa Demo dan Pelatihan (usulan)

Dipakai sebelum setiap sesi demo/pelatihan, terutama yang menampilkan layar ke pihak luar.

| No | Pemeriksaan | Ya | Tidak | Catatan |
| :--- | :--- | :---: | :---: | :--- |
| 1 | Instans yang dipakai adalah lingkungan demo/pelatihan, bukan produksi tenant klien | | | |
| 2 | Data yang tampil sintetis/tersamar; tidak ada nama klien atau data nyata | | | |
| 3 | Kredensial bawaan telah diganti dan tidak tampak di layar, slide, atau rekaman | | | |
| 4 | Instans demo tidak dapat diakses publik tanpa kendali akses (**perlu verifikasi** `auditsphere-devops`) | | | |
| 5 | Layar yang dibagikan: hanya jendela peramban demo; notifikasi dimatikan; tab lain ditutup | | | |
| 6 | Tidak ada terminal berisi `.env` atau kunci yang terlihat | | | |
| 7 | Rekaman, bila ada, disimpan terbatas dan tidak memuat data sensitif | | | |
| 8 | Setelah sesi: logout, hapus berkas unduhan/ekspor demo, hentikan berbagi layar | | | |
| 9 | Akun peserta pelatihan dinonaktifkan atau diganti sesuai masa pelatihan | | | |

---

## Lampiran G — Penanggung Jawab (RACI usulan)

| Aktivitas | Direktur Utama | Lead Information Security | Legal & Compliance | HR | DevOps Lead | Security Officer | Project Lead | Tim Backend/Frontend | Personel |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Kebijakan CDS-001 (dan selaras PHY-001/RWP-001) | A (persetujuan) | R | C | C | C | C | C | I | I |
| Baseline kunci layar dan pengumpulan bukti perangkat | I | A | I | I | C | R | C | I | R |
| Spot-check dan swa-lapor | I | A | I | I | I | R | C | I | R |
| Awareness T18, acknowledgement | I | A | C | C | I | R | R | I | R |
| Idle timeout, penanda cetak/ekspor, penyamaran (tiket aplikasi) | I | C | I | I | C | C | A | R | I |
| Ketentuan operator deploy dan soket SSH | I | A | I | I | R | C | C | I | R |
| Batas agen AI dan `.env` di laptop | I | A | C | I | C | R | C | C | R |
| Daftar periksa demo/pelatihan dan pembersihan kredensial bawaan | I | C | I | I | R | C | A | I | R |
| Kebijakan cetak dan pemusnahan | I | A | C | I | I | R | C | I | R |
| Pelaporan dan penanganan insiden perangkat/dokumen | I | A | R (dampak hukum/PDP) | I | R | R | I | I | R (melapor) |

R = Responsible, A = Accountable, C = Consulted, I = Informed. Pada tim kecil, satu orang dapat memegang beberapa peran; pemeriksa spot-check sebaiknya bukan orang yang diperiksa dan, bila memungkinkan, bukan orang yang mengelola area/perangkat yang diperiksa.

---

Catatan sinkronisasi checklist: baris A.7.7 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan tidak diubah oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" adalah usulan dan dapat disalin setelah divalidasi pengguna.
