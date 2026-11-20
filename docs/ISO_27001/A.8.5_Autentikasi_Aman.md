# A.8.5 — Autentikasi Aman

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Autentikasi aman |
| **Fokus pemeriksaan** | Mekanisme autentikasi sesuai risiko dan terlindungi |
| **Contoh bukti objektif** | MFA config, password policy, SSO setting, authentication log |
| **Sistem** | AuditSphere (auditsphere.app), auth-service dan layanan yang memvalidasi JWT, Kong, stack tenant, training, python-ai, Data Hub, akses operator (SSH, GitHub, konsol cloud) |
| **PIC** | Lead Information Security (didukung Backend Lead [konfirmasi peran] dan DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut mekanisme autentikasi **sesuai risiko** dan **terlindungi**. Repo membuktikan **kode dan konfigurasi bawaan**; repo **tidak** membuktikan apa yang berjalan di produksi (variabel lingkungan, ada tidaknya `MFA_ENCRYPTION_KEY`, status migrasi MFA, akun seed yang masih aktif, pengaturan Kong dan nginx terpasang, pengaturan SSH/GitHub/konsol cloud), sehingga semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH dan tidak ada percobaan login ke sistem apa pun dalam penyusunan dokumen ini. Dokumen ini hanya menjalankan `go test` pada paket auth-service yang relevan (lulus, bagian 3) dan membaca kode. Fakta terpenting: **MFA TOTP, penguncian akun setelah kode MFA salah, pembatalan sesi saat logout di auth-service, pin algoritma JWT, validasi panjang `JWT_SECRET`, dan log aktivitas login sudah ada di kode**; tetapi **MFA bersifat opsional untuk semua peran (termasuk ADMIN), kata sandi hanya wajib 6 karakter di backend, tidak ada penguncian atau pembatasan percobaan kata sandi pada login, tidak ada kode pemulihan MFA, tidak ada reset kata sandi mandiri, akun seed berkata sandi default sama, dan tidak ada SSO**. Seluruh angka pada Lampiran A sampai D (panjang kata sandi, ambang penguncian, masa sesi, retensi log, tenggat) adalah **[usulan]** yang belum disetujui manajemen dan **bukan hasil pengukuran**. Nilai rahasia, kata sandi (termasuk kata sandi default seeder), kunci, nilai hash, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.5 menurut data yang diberikan pengguna sudah berisi relevansi "Berlaku", PIC "Qiko", penilaian "Lengkapi assessment", kesiapan "Belum siap", jadwal "Tanpa jadwal", dan perlu tindakan "Ya" (berkas xlsx tidak dibuka ulang oleh dokumen ini). **Nilai usulan di bawah konsisten dengan baris itu; bukti repo tidak bertentangan dengan relevansi "Berlaku", kesiapan "Belum siap", maupun "Perlu tindakan: Ya"** (lihat butir kesiapan di tabel). Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.5** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Autentikasi aman** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Mekanisme autentikasi sesuai risiko dan terlindungi** | Sesuai baris spreadsheet; tiga gate: sesuai risiko (kekuatan faktor per peran dan per data), terlindungi (kata sandi, token, rahasia MFA, kunci layanan), terpantau (log autentikasi dan peninjauan). |
| Contoh bukti objektif | **MFA config, password policy, SSO setting, authentication log** | Sesuai baris spreadsheet; status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | Sesuai spreadsheet. AuditSphere memegang data audit klien lembaga jasa keuangan di balik satu pintu login; autentikasi adalah kendali akses pertama dan menentukan atribusi seluruh jejak audit (bagian 1). Pengecualian tidak dapat dipertahankan. |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: data klien bernilai tinggi, multi-tenant, peran ADMIN memegang kendali penuh, klien lembaga keuangan akan menanyakan MFA dan kebijakan kata sandi, serta log login adalah bukti forensik. |
| Status implementasi | **Sebagian** | Ada: bcrypt, JWT HS256 dengan pin algoritma dan kedaluwarsa wajib, sesi Redis yang dibatalkan saat logout (auth-service), MFA TOTP, penguncian MFA 3 kali salah, step-up kata sandi, perangkat tepercaya, enkripsi rahasia MFA di kode, log login. Belum ada: MFA wajib per peran, kebijakan kata sandi yang memadai, penguncian kata sandi, kode pemulihan, reset mandiri, SSO, standar tertulis, bukti konfigurasi produksi. |
| Ketersediaan bukti | **Sebagian (kode, uji unit, runbook) / Tidak ada (kebijakan, konfigurasi produksi, tinjauan log)** | Dari empat contoh bukti: MFA config **sebagian** (kode ada, status produksi belum terbukti); password policy **tidak ada** (hanya batas minimum 6 karakter di kode); SSO setting **tidak ada** (fitur tidak ada); authentication log **sebagian** (log dihasilkan, belum ada tinjauan atau alert). |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.5-01 s/d EV-A8.5-15) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **Backend Lead** [konfirmasi peran] dan **DevOps Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal setelah keputusan MFA wajib dan rotasi akun seed selesai | Selanjutnya: tinjauan log autentikasi [mingguan] dan tinjauan standar [tahunan] (Lampiran A, C). |
| Desain kontrol | **Perlu perbaikan** | Fondasi teknis cukup baik (fail-closed pada MFA dan sesi, pin algoritma, validasi rahasia), tetapi desain tidak sesuai risiko: faktor kedua tidak wajib bagi ADMIN, kebijakan kata sandi lemah, tidak ada pembatasan tebakan kata sandi per akun, sesi tidak berlaku di layanan lain setelah logout, rahasia MFA belum terbukti terenkripsi di produksi. |
| Efektivitas operasional | **Belum teruji** | Tidak ada uji penetrasi, uji brute-force terkontrol, tinjauan log autentikasi, atau pengukuran cakupan MFA. Uji unit lulus (bagian 3) tetapi bukan bukti operasi. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: MFA opsional (ADMIN tanpa MFA wajib); minimum kata sandi 6 karakter tanpa riwayat/kedaluwarsa/cek kebocoran; tanpa penguncian atau pembatasan tebakan kata sandi; tanpa kode pemulihan MFA dan tanpa rute reset MFA oleh admin; rahasia MFA belum terbukti terenkripsi di produksi (dan template tenant berpotensi gagal start); akun seed berkata sandi default sama; sesi tidak dibatalkan di layanan lain dan tidak dibatalkan saat akun dinonaktifkan; reset mandiri tidak ada walau halaman UI ada; CORS auth-service memantulkan origin apa pun; IP klien dari `X-Forwarded-For` tanpa validasi; literal kunci layanan di compose; tanpa SSO; tanpa alert dan tinjauan log; autentikasi operator tidak terbukti. |
| Tindakan perbaikan | **Lihat bagian 5 (17 tindakan)** | Draf Lampiran A sampai D sebagai masukan. |
| Prioritas | **Tinggi** | Akun ADMIN tanpa MFA wajib, kata sandi minimum 6 karakter, tanpa pembatasan tebakan kata sandi per akun, dan akun seed berkata sandi default sama pada sistem yang memegang data klien lembaga jasa keuangan; klien akan menanyakan MFA dan kebijakan kata sandi. |
| Target | **[usulan] 31 Desember 2026** untuk standar disetujui, akun seed dirotasi, dan verifikasi konfigurasi produksi; **[usulan] 31 Januari 2027** untuk MFA wajib ADMIN, kebijakan kata sandi di backend, dan penguncian | Keputusan pengguna. Perubahan kode memerlukan tiket ke agen backend/frontend, bukan dilakukan dokumen ini. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Konsisten dengan spreadsheet. Gate "sesuai risiko" belum terpenuhi (MFA tidak wajib, kata sandi lemah); gate "terlindungi" belum terbukti di produksi (rahasia MFA, akun seed); gate "terpantau" belum terpenuhi (tanpa alert dan tinjauan). |
| Deadline | **[usulan] 15 Januari 2027** | Batas kesiapan: standar disetujui, akun seed dirotasi, konfigurasi produksi terverifikasi, keputusan MFA wajib ADMIN dengan tanggal pelaksanaan. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Konsisten dengan spreadsheet. Tindakan 1, 2, 5, 6, 16 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.5 **berlaku** untuk AuditSphere karena:

1. **Autentikasi adalah pintu ke data klien.** Seluruh `/api/v1` di audit-service dan layanan lain hanya dilindungi JWT (Kong tidak memvalidasi JWT; setiap layanan memvalidasi sendiri). Data yang dilindungi mencakup KKA, LHA, ATR, berkas bukti, dan kredensial terenkripsi sumber data klien di Data Hub (A.5.12).
2. **Peran ADMIN berdaya besar.** ADMIN membuat dan menghapus akun, mengubah peran, mereset kata sandi, dan membaca log aktivitas (`backend/auth-service/routes/registry.go:181-198`; A.5.16). Akun ADMIN yang diambil alih membuka semua tenant pada stack tersebut; faktor autentikasi untuk peran ini harus paling kuat.
3. **Atribusi jejak audit bergantung pada identitas yang kuat.** Produk ini menghasilkan catatan audit yang dipakai sebagai bukti; log aktivitas dan klaim JWT (`user_id`, peran, kode karyawan) menjadi dasar siapa melakukan apa (A.5.28).
4. **Multi-tenant dan satu Redis bersama.** Stack tenant memakai database per layanan, indeks Redis logis per tenant, dan `JWT_SECRET` per tenant (`scripts/templates/docker-compose.tenant.yml.tpl:26-45`). Rahasia yang dipakai ulang antar-tenant membuat token satu tenant sah di tenant lain karena Kong tidak memvalidasi JWT.
5. **Rahasia autentikasi memiliki jalur kebocoran sekunder.** Rahasia MFA (TOTP) disimpan di basis data dan ikut ke dump backup (A.8.13); catatan operasional menyebut rahasia itu tanpa enkripsi di produksi (bagian 4 butir 8).
6. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim menanyakan MFA untuk akun istimewa, kebijakan kata sandi, penguncian akun, SSO, dan log login. Saat ini tidak ada dokumen kebijakan autentikasi tunggal yang dapat diserahkan; ISO-AS-001 bagian 8 menyebut nilai yang tidak diterapkan kode (A.5.16 bagian 4 butir 5).
7. **Operator mengakses produksi lewat SSH, GitHub Actions, dan konsol pemasok.** Autentikasi operator termasuk lingkup kontrol ini, tetapi tidak terlihat dari repo (bagian 4 butir 17).

Pengecualian "tidak berlaku" tidak dapat dipertahankan.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti kode atau konfigurasi di repo (cabang `staging`)**. Apakah itu yang berjalan di produksi adalah **Perlu verifikasi** kecuali dinyatakan lain.

### 2.1 Alur login dan kredensial (auth-service)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Alur login | `POST /api/v1/auth/login` dan `POST /api/v1/auth/verify-mfa-login` adalah satu-satunya rute publik autentikasi; sisanya di belakang `Authenticate()` dan allow-list VIEWER. Kata sandi diperiksa sebelum status MFA dan kunci diungkapkan. | `backend/auth-service/routes/registry.go:80-96`; `backend/auth-service/services/auth/auth_service.go:122-201` | Lihat bagian 4 butir 1 sampai 3. |
| Hashing kata sandi | bcrypt dengan `bcrypt.DefaultCost` (nilai 10 pada `golang.org/x/crypto`; periksa versi di `go.mod`). | `backend/auth-service/pkg/utils/password.go:11-18` | Cost tidak dikonfigurasi; tidak ada batas atas panjang (bcrypt memotong pada 72 byte). Cost 10 adalah bawaan, bukan keputusan yang didokumentasikan. |
| Pesan galat generik | Nama pengguna tidak dikenal dan kata sandi salah sama-sama menghasilkan `ErrInvalidCredentials`; alasan sebenarnya hanya masuk log aktivitas. | `auth_service.go:124-147`, `:614-623` | Akun nonaktif mendapat galat berbeda (bagian 4 butir 3). |
| Kata sandi sementara | Kata sandi sementara dibuat dengan `crypto/rand`, 12 karakter, selalu memuat huruf besar, huruf kecil, dan angka; `must_change_password` diset; admin tidak dapat mereset dirinya atau admin lain. | `backend/auth-service/pkg/utils/password.go:20-61`; `backend/auth-service/services/user/user_service.go:489-530` | Kata sandi sementara dikembalikan ke admin dalam respons API dan tidak memiliki masa berlaku (bagian 4 butir 10). |
| Ganti kata sandi | Memerlukan kata sandi lama; kata sandi baru tidak boleh sama dengan yang lama; kegagalan dan keberhasilan dicatat ke log aktivitas dan Kafka. | `auth_service.go:556-612`; `backend/auth-service/models/dto.go:66-67` | Minimum 6 karakter; tanpa riwayat (bagian 4 butir 1). |
| Paksa ganti kata sandi | Middleware Nuxt mengarahkan akun `mustChangePassword` ke halaman ganti kata sandi; halaman itu mewajibkan 8 karakter dengan huruf besar, kecil, dan angka. | `frontend/middleware/01-force-password-change.global.ts`; `frontend/pages/auth/change-password.vue:109-114` | Aturan 8 karakter hanya di sisi klien; backend menerima 6 karakter (bagian 4 butir 1). |

### 2.2 Token dan sesi

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Algoritma dan kedaluwarsa JWT | Penandatanganan HS256; parsing hanya menerima HS256 dan mewajibkan `exp`; masa berlaku bawaan 24 jam (`jwt.expiry_hour`). Middleware auth-service memakai pin yang sama. | `backend/auth-service/pkg/utils/jwt.go:37-56`; `backend/auth-service/pkg/middleware/auth.go:66-74`; `backend/auth-service/pkg/config/config.go:132`, `:206` | Kunci simetris yang sama dipakai semua layanan (bagian 4 butir 4). |
| Penanganan `JWT_SECRET` | Layanan menolak start bila kosong, kurang dari 64 byte, atau sama dengan nilai publik yang pernah ter-commit (dibandingkan lewat digest, nilai tidak ada di kode ini); dibaca dari lingkungan. Runbook rotasi lima layanan tersedia. | `backend/auth-service/pkg/config/jwt_secret.go:14-44`; `backend/auth-service/cmd/auth/serve.go:56-58`; `docs/jwt-rotation-runbook.md`; `docs/security-hardening-changes-backend.md` bagian A | Apakah rotasi sudah dijalankan di produksi dan secret produksi memenuhi syarat: **Perlu verifikasi**. |
| Rahasia per tenant | Generator stack tenant membuat `JWT_SECRET` acak per tenant (`openssl rand -hex 32`, 64 karakter) dan `APP_SIGNATURE_KEY` terpisah. | `scripts/generate-tenant-stack.sh:221-226`; `scripts/templates/docker-compose.tenant.yml.tpl:43-45` | Tidak ada pemeriksaan otomatis bahwa dua tenant tidak berbagi secret (bagian 4 butir 19). |
| Sesi server dan logout | Login menyimpan `token:<jwt>` di Redis dengan TTL sama dengan JWT; middleware auth-service gagal-tutup bila sesi tidak ada atau Redis tidak tersedia; logout menghapus kunci. | `auth_service.go:250-255`, `:504-535`; `backend/auth-service/cmd/auth/serve.go:173`, `:270-276`; `backend/auth-service/pkg/middleware/auth.go:84-93`; `backend/auth-service/pkg/middleware/session_test.go` | Hanya auth-service yang memeriksa sesi (bagian 4 butir 5). |
| Token tidak lewat URL atau cookie di layanan lain | `?token=` dan cookie token dihapus dari middleware audit, risk, dan analytics; hanya header `Authorization`. Log akses menyamarkan parameter bernama token, secret, key, password. | `docs/security-hardening-changes-backend.md` bagian D; `backend/auth-service/pkg/middleware/redact.go` | Frontend tetap menyimpan token di cookie `auth-token` (bagian 4 butir 14). |
| Refresh token | **Tidak ada alur refresh yang dipakai.** `RefreshTokenRepository` ada tetapi tidak dipanggil; tidak ada rute refresh. | `backend/auth-service/repositories/refresh_token_repository.go`; `docs/security-hardening-changes-backend.md` bagian E | Sesi tetap 24 jam tanpa perpanjangan atau pemendekan. |
| Step-up kata sandi | `POST /auth/verify-password` menerbitkan token step-up berumur 5 menit untuk aksi istimewa (saat ini perubahan struktur organisasi); 5 kata sandi salah per 15 menit mengunci endpoint untuk pengguna itu (HTTP 429). | `backend/auth-service/services/auth/step_up_service.go:20-55`; `backend/auth-service/pkg/utils/step_up.go:30`; `backend/auth-service/routes/registry.go:93-94` | Hanya satu `purpose` pada whitelist; bukan step-up untuk reset kata sandi atau perubahan peran. |

### 2.3 MFA, perangkat tepercaya, dan penguncian

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| MFA TOTP | Enroll, verifikasi, nonaktifkan, status; TOTP 6 digit 30 detik, toleransi satu langkah sebelum dan sesudah, perbandingan waktu-konstan. Tahap kedua login memakai token MFA sekali pakai (Redis, 10 menit). | `backend/auth-service/routes/registry.go:129-140`; `backend/auth-service/services/mfa/mfa_service.go:138-182`, `:384-407`; `auth_service.go:184-197`, `:357-431`, `:453-477` | Opsional per akun (bagian 4 butir 6). Tidak ada perlindungan replay kode dalam jendela yang sama (tidak ditemukan). |
| Fail-closed MFA | Galat selain "tidak ada baris MFA" (basis data, rahasia tidak terdekripsi) menghentikan login, bukan melewati MFA. Penguncian membaca Redis dan gagal-tutup. | `auth_service.go:154-161`; `backend/auth-service/services/auth/mfa_lockout.go:74-87`; `mfa_secret_fail_closed_test.go` | - |
| Penguncian MFA | 3 kode salah per pengguna (bukan per token MFA) mengunci akun 15 menit (HTTP 429, `ACCOUNT_LOCKED`); kunci berlaku untuk tahap kata sandi dan MFA, termasuk perangkat tepercaya; hitungan dihapus saat benar. | `backend/auth-service/services/auth/mfa_lockout.go:18-30`, `:89-128`; `auth_service.go:165-170`, `:368-411`; `mfa_lockout_test.go` | Hanya menjaga tahap MFA; kata sandi tidak memiliki penghitung (bagian 4 butir 2). |
| Nonaktifkan MFA | Memerlukan kata sandi akun; kejadian `mfa.disabled` dikirim. | `mfa_service.go:275-298` | Tidak memerlukan kode MFA saat ini; tidak ada rute admin untuk mereset MFA pengguna yang kehilangan perangkat (bagian 4 butir 7). |
| Email OTP | Kode 6 digit, berlaku 5 menit, sekali pakai, perbandingan waktu-konstan, dikirim lewat Resend/SMTP. | `mfa_service.go:316-382`; `backend/auth-service/services/email/email_service.go:329` | Tidak ditemukan batas percobaan per pengguna (bagian 4 butir 7). Tahap kedua login hanya memverifikasi TOTP (`auth_service.go:398`). |
| Enkripsi rahasia MFA | AES-256-GCM (`enc:v1:`) terikat ke id pengguna; kunci dari `MFA_ENCRYPTION_KEY` (lingkungan); wajib bila `APP_APP_ENV=production` atau `MFA_ENCRYPTION_REQUIRED=true`; baris terenkripsi yang tidak terbuka membuat login gagal-tutup. Perintah `auth migrate encrypt-mfa-secrets` dry-run secara bawaan, satu transaksi, verifikasi baca-ulang. | `backend/auth-service/pkg/secretcrypt/secretcrypt.go`; `backend/auth-service/pkg/config/mfa_encryption.go:34-86`; `backend/auth-service/cmd/auth/serve.go:59-62`; `docs/mfa-encryption-rollout.md` | **Kode dan runbook ada; penerapan di produksi tidak terbukti dan menurut catatan operasional belum dijalankan** (bagian 4 butir 8). |
| Perangkat tepercaya | "Percayai perangkat ini" melewati MFA selama 90 hari; perangkat sesi saja tetap meminta MFA; perangkat tercatat di halaman Activity; email pemberitahuan saat perangkat ditambah atau dihapus. | `backend/auth-service/repositories/trusted_device_repository.go:34-35`; `backend/auth-service/services/trusted-devices/trusted_devices_service.go:93`; `auth_service.go:172-182`; `email_service.go:451`, `:508` | Sidik jari perangkat dikirim klien. Masa 90 hari tidak dapat dikonfigurasi (A.7.7 bagian 2.1). |

### 2.4 Autentikasi antar-layanan dan gerbang

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Kong tidak memvalidasi JWT | Kong prod hanya memakai plugin `prometheus`, `cors`, `rate-limiting`, `bot-detection`; tidak ada `jwt` atau `key-auth`. Validasi ada di tiap layanan. | `backend/kong-gateway/kong/prod/kong.yml:23-110` | Layanan yang lupa memasang `Authenticate()` terbuka; dijaga oleh `.github/workflows/route-coverage.yml` dan uji rute (di luar dokumen ini). |
| CORS Kong | Prod: satu origin (`https://auditsphere.app`), `credentials: true`; template tenant: satu origin per tenant. | `kong.yml:31-62`; `scripts/templates/kong-tenant.yml.tpl:30-40` | Lihat bagian 4 butir 13 untuk CORS di auth-service. |
| Rate limiting Kong | Global 100 per menit dan 5000 per jam per IP (`policy: local`); tidak ada batas khusus untuk `/api/v1/auth`. Batas khusus hanya pada rute fieldwork, dashboard, dan retrain. | `kong.yml:80-86`, `:464-470`, `:656-662`; `scripts/templates/kong-tenant.yml.tpl:75-80` | 100 per menit per IP terlalu longgar untuk login; IP asli di belakang nginx: **Perlu verifikasi** (bagian 4 butir 2). |
| python-ai | Header `X-API-Key` wajib di semua path kecuali `/health`; bila `API_KEY` tidak diset, semua path (kecuali `/health`) menolak; perbandingan waktu-konstan; `/retrain/auto` memerlukan token internal terpisah. Analytics mengirim `X-API-Key`. | `backend/python-ai/security.py`; `backend/python-ai/test_security.py`; `backend/analytics-service/services/python_ai_client.go:260`, `:287`; `docs/security-hardening-changes-backend.md` bagian F | Lihat bagian 4 butir 12 untuk literal bawaan di compose. |
| Data Hub API | Semua router kecuali kesehatan memakai dependensi `verify_api_key` (header `X-API-Key`). | `data-server/data-hub-api/main.py:41-48`, `:103-109` | Nilai bawaan literal dan perbandingan biasa (bagian 4 butir 12). |
| Port layanan | Port auth-service dan layanan lain dipublikasikan hanya ke `127.0.0.1` pada compose produksi. | `backend/docker-compose.prod.yml:93-94`, `:140-141`, `:172-173`, `:218-219`, `:247-248`, `:259-260` | Apakah nginx/ufw di VPS sesuai: **Perlu verifikasi** (A.8.21). |

### 2.5 Pencatatan autentikasi

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Log aktivitas login | Login berhasil, login gagal (dengan alasan: nama pengguna tidak dikenal, akun nonaktif, kata sandi salah, kode MFA salah, token MFA tidak valid, akun terkunci), logout, ganti kata sandi (sukses dan gagal) dicatat dengan IP dan user agent; tersimpan di `sys_alert_logs` lewat Kafka atau tulis langsung bila Kafka mati. | `auth_service.go:128`, `:138`, `:145`, `:258-269`, `:512-522`, `:565-600`, `:614-657`; `backend/auth-service/pkg/activitylog/`; `backend/auth-service/cmd/auth/serve.go:102-151` | Tidak ada alert atau deteksi pola (bagian 4 butir 16). |
| Peristiwa MFA, perangkat, peran, step-up, izin ditolak | Dikirim lewat `EventPublisher` dengan tap ke log aktivitas. | `backend/auth-service/pkg/kafka/events.go:14-51`; `serve.go:142-145` | Peristiwa akun (`account.locked`, `account.deactivated`) didefinisikan tetapi tidak dipanggil dari auth-service (A.5.16 butir 6). |
| Retensi dan akses log | Retensi bawaan 90 hari dengan pembersihan harian; akses lewat `GET /api/v1/system-logs` dengan penyaringan berdasarkan peran. | `backend/auth-service/pkg/config/config.go:50-64`, `:224-228`; `backend/auth-service/pkg/activitylog/retention.go`; `backend/auth-service/routes/registry.go:209-213`; `controllers/systemlog/system_log_controller.go:62-64` | Retensi 90 hari adalah bawaan teknis, bukan keputusan; nilai produksi: **Perlu verifikasi**. |
| Redaksi | Parameter `token`, `secret`, `key`, `password`, `signature`, `auth` disamarkan pada log akses; kata sandi tidak dicatat. | `backend/auth-service/pkg/middleware/redact.go`; `backend/auth-service/pkg/middleware/logger.go` | Nama pengguna yang diketik dicatat apa adanya pada login gagal; bila pengguna salah mengetik kata sandi di kolom nama pengguna, kata sandi itu masuk log (risiko kecil, dicatat). |

### 2.6 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `backend/`, `frontend/`, `scripts/`, `data-server/`, `docs/` (19 November 2026) tidak menemukan: **SSO, SAML, OIDC, OAuth login, atau integrasi IdP** (tidak ada rute, konfigurasi, maupun halaman); kebijakan kata sandi di backend selain `min=6`; riwayat kata sandi, masa berlaku kata sandi, atau pemeriksaan kata sandi bocor; penguncian atau penghitung kegagalan kata sandi pada login (kolom `LockedUntil` tidak pernah diisi, `TrackLoginAttempt` hanya dipakai step-up; `RateLimiter` tollbooth di `pkg/middleware/security.go` tidak dipasang pada rute mana pun); MFA wajib per peran; kode pemulihan MFA yang disimpan (runbook menyatakan hanya dikembalikan di DTO); rute admin untuk mereset MFA; rute `forgot-password` atau `reset-password` di backend; alur undangan dengan token sekali pakai; pembatasan satu sesi aktif; refresh token; WebAuthn/passkey; kebijakan autentikasi tertulis; alert pada kegagalan login; konfigurasi autentikasi operator (SSH, GitHub, konsol cloud). Keadaan produksi **Perlu verifikasi**.

### 2.7 Dokumen yang menyinggung

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| `docs/security-hardening-changes-backend.md` | Daftar perubahan: rotasi `JWT_SECRET`, pin HS256, logout mengakhiri sesi, `/auth/register` dihapus, token tidak lewat query/cookie, python-ai fail-closed; risiko sisa sesi di layanan lain. | Bukti perubahan kode (ditulis 2026); bukan bukti penerapan produksi. |
| `docs/mfa-encryption-rollout.md` | Urutan rollout enkripsi MFA (training dahulu), perintah dry-run, kunci hilang, rollback. | Runbook; keadaan "kunci tidak diset di produksi dan training" disebut sendiri. |
| `docs/jwt-rotation-runbook.md` | Rotasi `JWT_SECRET` lima layanan. | Runbook; eksekusi: **Perlu verifikasi**. |
| ISO-AS-001 bagian 8 (dokumen keselarasan) | Kata sandi 12 karakter, kedaluwarsa 90 hari, riwayat 5, penguncian 5 kali/30 menit, idle timeout 15 menit, satu sesi aktif, MFA wajib ADMIN/CAE/Audit Manager. | **Tidak diterapkan kode** (A.5.16 butir 5); tidak boleh dipakai sebagai bukti A.8.5. |
| `docs/ISO_27001/A.5.16_Manajemen_Identitas.md` | Siklus identitas, akun default, sesi tidak dicabut saat leaver. | Ada; beberapa butir sudah usang (bagian 4 butir 20). |
| `docs/ISO_27001/A.7.7_Meja_Bersih_dan_Layar_Bersih.md` | Tidak ada idle timeout, cookie 1 atau 30 hari, perangkat tepercaya 90 hari. | Ada; **tidak diulang** di sini. |

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| MFA config | **Sebagian** | Kode, rute, dan runbook enkripsi ada. Tidak ada: daftar akun ber-MFA per peran, bukti `MFA_ENCRYPTION_KEY` terpasang di produksi, hasil dry-run migrasi, kebijakan MFA wajib. Ekspor yang disarankan (read-only, oleh pihak berwenang): jumlah `mfa_setups` aktif per peran tanpa kolom rahasia. |
| Password policy | **Belum tersedia; draf di Lampiran A** | Kode hanya `min=6` (`models/dto.go:12`, `:67`, `:76`); halaman ganti kata sandi memeriksa 8 karakter di klien. Draf AUT-001 (usulan kode). |
| SSO setting | **Tidak ada** | Fitur tidak ada. Dokumen perlu menyatakan "tidak diimplementasikan" dan keputusan roadmap (tindakan 15), bukan mengisi bukti palsu. |
| Authentication log | **Sebagian** | Peristiwa dihasilkan dan disimpan (bagian 2.5). Belum ada: contoh ekspor bertanggal, catatan tinjauan, alert, bukti retensi produksi. Daftar periksa tinjauan di Lampiran C. |
| Uji unit autentikasi | **Ada (19 November 2026)** | `go test -count=1 ./services/auth/... ./pkg/config/... ./pkg/middleware/... ./pkg/secretcrypt/... ./pkg/utils/...` di `backend/auth-service` lulus (lima paket). Mencakup penguncian MFA, fail-closed rahasia MFA, JWT, sesi, kebijakan secret. Bukan bukti operasi. |
| Konfigurasi gerbang (Kong rate limit dan CORS) | **Sebagian** | File konfigurasi di repo; versi terpasang di VPS: **Perlu verifikasi** (`scripts/sync-kong.sh`). |
| Akun default dirotasi | **Belum tersedia** | Tidak ada catatan rotasi atau penonaktifan akun seed di produksi, training, atau tenant. |
| Autentikasi operator (SSH, GitHub, konsol cloud) | **Belum tersedia** | Di luar repo. `deploy.yml` memakai kunci SSH dari GitHub secret dengan host key tersemat; MFA GitHub dan konsol pemasok tidak terlihat. |
| Hasil uji penetrasi atau brute-force terkontrol | **Belum tersedia** | Tidak ada. |
| Tinjauan log autentikasi berkala | **Belum tersedia** | Tidak ada catatan tinjauan. |
| Persetujuan standar autentikasi | **Belum tersedia** | AUT-001 adalah draf di dokumen ini. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari catatan operasional internal dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan; **[Luar repo]** tidak dapat dilihat dari repo.

1. **[Repo] Kebijakan kata sandi lemah di backend.** Batas hanya `min=6` untuk login, pembuatan akun, dan ganti kata sandi (`backend/auth-service/models/dto.go:12`, `:67`, `:76`). Tidak ada kompleksitas, daftar kata sandi umum atau bocor, riwayat, atau masa berlaku. Halaman ganti kata sandi memeriksa 8 karakter dengan huruf besar, kecil, dan angka (`frontend/pages/auth/change-password.vue:109-114`) tetapi pemeriksaan itu dapat dilewati langsung ke API. Nilai pada ISO-AS-001 (12 karakter, riwayat 5, kedaluwarsa 90 hari) tidak ditegakkan.
2. **[Repo] Tidak ada penguncian atau pembatasan tebakan kata sandi per akun.** Login tidak memiliki penghitung kegagalan; `LockedUntil` hanya dibersihkan, tidak pernah diisi (`user_service.go:528`); `RateLimiter` ada tetapi tidak dipasang. Satu-satunya batas adalah rate limiting global Kong 100 per menit per IP (`kong.yml:80-86`), yang tidak membatasi serangan terdistribusi atau password spraying per akun, dan efektif hanya jika IP asli klien diteruskan benar oleh nginx ke Kong (**Perlu verifikasi**). Penguncian yang ada hanya untuk tahap MFA (3 kali) dan step-up (5 kali).
3. **[Repo] Enumerasi akun dan timing.** Nama pengguna tidak dikenal dan kata sandi salah menghasilkan galat yang sama (baik), tetapi (a) akun nonaktif mendapat `ErrUnauthorized` yang berbeda **sebelum** kata sandi diperiksa (`auth_service.go:136-140`), sehingga status akun terungkap tanpa mengetahui kata sandi; (b) nama pengguna tidak dikenal kembali tanpa menghitung bcrypt, sehingga waktu respons berbeda; (c) respons terkunci MFA hanya diberikan setelah kata sandi benar (baik). Bentuk respons nyata perlu diuji (**Perlu verifikasi**).
4. **[Repo] JWT HS256 dengan kunci simetris yang dibagi lima layanan.** Layanan mana pun yang dikompromikan dapat menerbitkan token untuk pengguna dan peran apa pun di stack itu. Kebijakan panjang dan blocklist secret sudah ada, serta runbook rotasi; apakah rotasi dijalankan dan secret produksi memenuhi syarat: **Perlu verifikasi**. Peran berada di klaim JWT sehingga perubahan peran baru berlaku setelah login ulang atau token kedaluwarsa (24 jam).
5. **[Repo] Logout dan penonaktifan tidak berlaku di semua layanan.** Pemeriksaan sesi Redis hanya dipasang di auth-service. Audit, master, risk, dan analytics memvalidasi JWT secara lokal sehingga token yang sudah logout tetap sah di sana sampai 24 jam (`docs/security-hardening-changes-backend.md` bagian E, "Residual risk"; `backend/audit-service/pkg/middleware/auth.go` tidak memiliki pemeriksaan sesi). Menonaktifkan akun juga tidak menghapus kunci `token:` (tidak ada pemanggilan `InvalidateUserTokens` pada `user_service.go`; A.5.16 butir 1). Mengganti atau mereset kata sandi pun tidak mencabut token lama. Tidak ada idle timeout (A.7.7, tidak diulang di sini).
6. **[Repo] MFA tidak wajib untuk peran apa pun.** Tahap MFA hanya dijalankan bila pengguna sendiri sudah mengaktifkan MFA (`auth_service.go:161`). Tidak ada kolom, konfigurasi, atau middleware "MFA wajib"; ADMIN dapat login dengan kata sandi saja. Ini bertentangan dengan ISO-AS-001 bagian 8 dan dengan ekspektasi klien untuk akun istimewa. Berapa akun ADMIN/AUDITOR yang sudah ber-MFA di produksi: **Perlu verifikasi**.
7. **[Repo] Pemulihan MFA dan perlindungan kode email.** Tidak ada kode pemulihan tersimpan (runbook: hanya dikembalikan di DTO, tidak disimpan) dan tidak ditemukan rute admin untuk mereset MFA pengguna yang kehilangan perangkat; runbook enkripsi menyebut "admin resets their MFA" tetapi rute itu tidak ada di `registry.go` (hanya `POST /mfa/disable` oleh pengguna sendiri dengan kata sandi, atau hapus permanen akun `?hard=true`). Akibat: pengguna yang kehilangan perangkat bergantung pada intervensi basis data atau akun dihapus-dibuat ulang. Kode email OTP tidak memiliki batas percobaan per pengguna pada `/mfa/verify-email-code` (rute di belakang JWT; risiko dinilai rendah-sedang dan perlu uji). Kode TOTP yang sama dapat dipakai ulang dalam jendela yang sama (tidak ditemukan penyimpanan langkah terpakai).
8. **[Repo]+[Catatan] Rahasia MFA belum terbukti terenkripsi di produksi.** Kode dan runbook ada, tetapi compose produksi tidak mendefinisikan `MFA_ENCRYPTION_KEY` maupun `APP_APP_ENV` (`backend/docker-compose.prod.yml`, auth-service sekitar baris 93-110), sehingga auth-service start dengan peringatan dan menyimpan rahasia baru sebagai teks polos (`serve.go:59-62`). Catatan operasional dan runbook menyatakan produksi dan training dalam keadaan ini. Backup basis data juga membawa rahasia ini (A.8.13: backup terpasang belum terbukti terenkripsi). Prasyarat praktis sebelum migrasi: backup terenkripsi dan restore test (A.8.13 tindakan 5). **Perhatian untuk provisioning tenant:** template tenant menetapkan `APP_APP_ENV: production` (`scripts/templates/docker-compose.tenant.yml.tpl:48`) yang membuat kunci MFA wajib (`pkg/config/mfa_encryption.go:34-40`), sedangkan `generate-tenant-stack.sh` tidak membuat `MFA_ENCRYPTION_KEY` (tidak ditemukan pada `scripts/`). Kecuali kunci disuntikkan lewat jalur lain, auth-service tenant baru akan menolak start. **Perlu verifikasi** oleh devops sebelum tenant berikutnya; ini juga syarat keamanan karena setiap tenant perlu kunci sendiri (runbook).
9. **[Repo] Perangkat tepercaya 90 hari melewati MFA** dan sidik jari dikirim klien (A.7.7 membahas dari sisi layar; dicatat di sini sebagai pelemah faktor kedua). Perlu ditentukan apakah ADMIN boleh memakainya (Lampiran B).
10. **[Repo] Reset kata sandi dan undangan.** Tidak ada reset mandiri: `frontend/pages/auth/forgot-password.vue` dan `reset-password.vue` ada dan `frontend/stores/auth.ts:464`, `:477` memanggil `/auth/forgot-password` dan `/auth/reset-password`, tetapi rute itu tidak ada di backend (`registry.go`; `docs/security-hardening-changes-backend.md` mencatat penghapusan `/auth/register`). Pengguna mendapat galat pada fitur yang tampak tersedia (**Perlu verifikasi** tampilan dan tautan di halaman login). Reset hanya oleh admin: kata sandi sementara acak 12 karakter dikembalikan dalam respons API ke admin (dibacakan ke pengguna), tanpa masa berlaku, tanpa dikirim lewat email, dan `must_change_password` dipaksa. Tidak ada alur undangan dengan token sekali pakai dan berumur pendek; akun dibuat admin dengan kata sandi awal. Admin mereset kata sandi tanpa step-up.
11. **[Repo] Kredensial default seeder.** `backend/auth-service/pkg/database/seeders/seeder.go:223-300` membuat enam akun (peran ADMIN, AUDITOR, DEPARTMENT_HEAD, AUDITEE, VIEWER) dengan **satu kata sandi default yang sama** (nilai tidak dikutip di dokumen ini), tanpa `must_change_password`. `scripts/onboard-tenant.sh` menjalankan `auth seed` untuk tenant baru dan mencetak kredensial default pada keluarannya; panduan demo (`docs/AUDITSPHERE_DEMO_SETUP_GUIDE.md`) mencantumkan nilai itu dan menyuruh dirotasi, dengan contoh perintah rotasi yang rutenya tidak ditemukan di `registry.go` (mis. `PUT /users/me/password`; tidak diverifikasi di tempat lain). Layanan memaksa `seed` pada start container di beberapa alur (lihat memori internal: seeder menimpa data), sehingga akun ini dapat muncul kembali di lingkungan non-produksi. Akun seed yang masih aktif di produksi, training, dan tenant: **Perlu verifikasi** (pemeriksaan read-only: daftar akun seed beserta `last_login_at` oleh pihak berwenang, tanpa mencoba login).
12. **[Repo] Kunci antar-layanan.** `backend/docker-compose.prod.yml:123`, `:125`, `:204` memberi nilai bawaan literal (`:-...`) untuk `PYTHON_AI_API_KEY` dan `DATA_HUB_API_KEY` yang ter-commit di repo (nilai tidak dikutip); siapa pun dengan akses repo mengetahuinya dan, bila lingkungan produksi tidak menimpa, kunci itu sah. Memori internal mencatat produksi memakai override yang tidak dilewatkan `deploy-prod.sh`, jadi keadaan nyata: **Perlu verifikasi**. Data Hub API: `API_KEY` memiliki nilai bawaan literal bila lingkungan kosong dan dibandingkan dengan `!=` tanpa waktu-konstan (`data-server/data-hub-api/main.py:18`, `:46`); berbeda dengan python-ai yang fail-closed. Dua klien memakai satu kunci statis tanpa rotasi atau pemisahan per pemanggil.
13. **[Repo] CORS auth-service memantulkan origin apa pun dengan kredensial.** `backend/auth-service/pkg/middleware/cors.go:11-17` mengisi `Access-Control-Allow-Origin` dengan nilai `Origin` permintaan dan `Allow-Credentials: true` (dan `*` bila tidak ada). Pertahanan bergantung pada Kong yang hanya mengizinkan satu origin dan pada port layanan yang terikat ke loopback. Karena autentikasi memakai header Bearer (bukan cookie) risiko CSRF terbatas, tetapi allow-list per konfigurasi lebih sesuai. Kemungkinan duplikasi header CORS antara Kong dan layanan: **Perlu verifikasi**.
14. **[Repo] Cookie dan penyimpanan token di browser.** Token disimpan di cookie `auth-token` lewat `useCookie` dengan `maxAge` 1 hari atau 30 hari (`frontend/stores/auth.ts:333-334`, `:357-361`), diset dari JavaScript sehingga tidak `HttpOnly`, dan tidak ada atribut `Secure` atau `SameSite` eksplisit pada kode. Kebocoran lewat XSS memberi token yang sah hingga 24 jam. Atribut yang efektif pada respons produksi: **Perlu verifikasi** (pemeriksaan header oleh pihak berwenang). Masa cookie dan idle timeout dibahas A.7.7.
15. **[Repo] IP klien tidak tepercaya.** `getClientIP` mengembalikan nilai mentah `X-Forwarded-For` (bisa berupa daftar dan dapat dipalsukan klien) tanpa daftar proxy tepercaya (`backend/auth-service/controllers/auth/auth_controller.go:15-28`). Alamat IP pada log login dan perangkat tepercaya dapat dipalsukan dan tidak andal untuk forensik (A.5.28). Konfigurasi nginx yang menimpa header: **Perlu verifikasi** (A.8.21).
16. **[Repo]+[Tidak ada] Log autentikasi ada tetapi tidak dipantau.** Peristiwa login, login gagal beralasan, penguncian MFA, ganti kata sandi, dan peristiwa MFA dicatat (bagian 2.5). Tidak ada alert untuk rentetan kegagalan, login ADMIN dari perangkat baru, atau penguncian akun (Prometheus tanpa `rule_files` atau Alertmanager; A.8.21 butir 4), tidak ada tinjauan terjadwal, dan retensi bawaan 90 hari belum diputuskan terhadap kebutuhan (A.8.15 draf). Pembersihan menghapus baris lebih tua dari retensi tanpa salinan arsip. Kegagalan Kafka jatuh ke penulisan langsung sehingga peristiwa tidak hilang, tetapi tidak ada bukti uji.
17. **[Luar repo] Autentikasi operator tidak terbukti.** Repo hanya menunjukkan: deploy produksi lewat SSH dengan kunci dari GitHub secret dan host key tersemat tanpa `ssh-keyscan` (`.github/workflows/deploy.yml:63-76`); workflow dev memakai `appleboy/ssh-action` dengan kunci dan passphrase dari secret (`.github/workflows/deploy-dev.yml:18-25`); pesan galat workflow prod menyatakan kunci berfrasa sandi tidak didukung, sehingga kunci CI tersimpan tanpa frasa sandi di GitHub secret. Yang **tidak** terlihat dan **Perlu verifikasi**: apakah login kata sandi SSH dimatikan, apakah root login dibatasi, MFA organisasi GitHub dan branch protection, MFA pada konsol pemasok (registrar/DNS, Cloudflare, Backblaze, Resend, penyedia VPS), manajer kata sandi, serta siapa saja pemegang kunci. Admin API Kong dipublikasikan ke loopback (`docker-compose.prod.yml:293`); akses dan autentikasinya: **Perlu verifikasi**.
18. **[Tidak ada] Tidak ada SSO/SAML/OIDC.** Satu-satunya autentikasi adalah username dan kata sandi lokal ditambah TOTP. Klien korporat dan perbankan lazim meminta SSO atau integrasi IdP; roadmap Zero Trust di `docs/ISO_27001/advanced_security/ZTAR-001_Zero_Trust_Architecture_Roadmap.docx` tidak dibaca dan tidak dijadikan bukti. Status "tidak diimplementasikan" harus dinyatakan jujur dalam SoA dan jawaban due diligence.
19. **[Repo] Isolasi autentikasi antar-tenant bergantung pada konfigurasi.** Isolasi berasal dari `JWT_SECRET` per tenant dan indeks Redis per tenant (kunci `token:`, `mfa_login:`, penghitung kegagalan tidak berprefiks tenant, template menyebutnya sendiri). Tidak ada validasi otomatis bahwa `JWT_SECRET` dan kunci MFA unik antar-tenant; secret yang disalin akan membuat token satu tenant sah di tenant lain tanpa gejala (Kong tidak memvalidasi JWT; lihat juga memori tentang kunci lingkungan yang diam-diam merusak isolasi). Template mendukung 15 tenant pada Redis bawaan (komentar template). Perlu pemeriksaan unik-secret pada generator dan audit berkala.
20. **[Repo] Dokumen annex lain sebagian usang.** A.5.16 bagian 4 butir 5 menyatakan lockout tidak berjalan dan menyiratkan tidak ada penguncian apa pun; sejak commit `a622877` ada penguncian MFA dan sejak perubahan hardening ada pembatalan sesi logout di auth-service. Yang masih benar: tidak ada penguncian kata sandi dan minimum 6 karakter. Perlu disinkronkan agar dua annex tidak saling bertentangan (tindakan 17).
21. **Status dokumen.** Seperti annex lain di direktori ini, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet dalam penyusunan dokumen ini; tidak ada akses SSH dan tidak ada percobaan login.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Setujui standar autentikasi AUT-001** (Lampiran A) dengan keputusan manusia atas panjang kata sandi, kedaluwarsa, ambang penguncian, dan MFA wajib per peran (Lampiran B). | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 15 Desember 2026 | AUT-001 disetujui (EV-A8.5-08) |
| 2 | **Rotasi atau nonaktifkan akun seed** di produksi, training, dev, dan semua tenant; ubah alur onboarding agar akun admin awal memakai kata sandi acak per tenant dengan `must_change_password` (tiket backend/devops, tanpa menjalankan seeder di produksi). Perintah ke produksi hanya atas konfirmasi eksplisit. | DevOps Lead (operasi), Backend Lead (tiket seeder) | Tinggi | [usulan] 31 Desember 2026 | Daftar akun seed dengan status (tanpa kata sandi), catatan rotasi (EV-A8.5-10) |
| 3 | **Tiket backend: kebijakan kata sandi di server**: minimum [12] karakter, daftar kata sandi umum/bocor, riwayat [5], larangan memuat nama pengguna, validasi sama pada pembuatan akun, ganti, dan reset; selaraskan pesan di frontend. | Backend Lead, Lead Information Security | Tinggi | [usulan] 31 Januari 2027 | PR dan uji unit (EV-A8.5-01) |
| 4 | **Tiket backend dan devops: pembatasan tebakan kata sandi**: penghitung kegagalan per akun dan per IP pada login dengan penguncian atau penundaan progresif dan peristiwa `account.locked`; batas Kong khusus `/api/v1/auth/login` dan `/verify-mfa-login`. Perubahan Kong dilakukan `auditsphere-devops`, bukan dokumen ini. | Backend Lead, DevOps Lead | Tinggi | [usulan] 31 Januari 2027 | PR, konfigurasi Kong (tersanitasi), hasil uji terkontrol (EV-A8.5-11, -14) |
| 5 | **MFA wajib untuk ADMIN** (fase 1), lalu peran lain sesuai Lampiran B, dengan prasyarat: kode pemulihan, rute reset MFA oleh admin yang memerlukan step-up dan tercatat, masa transisi, komunikasi pengguna. | Lead Information Security (keputusan), Backend Lead dan Frontend Lead (implementasi) | Tinggi | [usulan] 31 Januari 2027 | Keputusan tertulis, PR, bukti cakupan MFA (EV-A8.5-03, -09) |
| 6 | **Verifikasi read-only keadaan produksi** (oleh pihak berwenang atau lewat `auditsphere-devops` atas konfirmasi eksplisit): ada tidaknya `MFA_ENCRYPTION_KEY` dan `APP_APP_ENV` pada auth-service, jumlah rahasia MFA teks polos (dry-run migrasi: hanya hitungan), akun seed aktif, rotasi `JWT_SECRET` terjalankan, retensi log aktivitas, atribut cookie dan header CORS pada respons. Jangan menampilkan nilai rahasia. | DevOps Lead | Tinggi | [usulan] 30 November 2026 | Ekspor read-only bertanggal tersanitasi (EV-A8.5-09) |
| 7 | **Enkripsi rahasia MFA di produksi dan training** mengikuti `docs/mfa-encryption-rollout.md` (training dahulu), setelah backup terenkripsi dan restore test (A.8.13 tindakan 4-5); kunci per tenant disimpan di pengelola kata sandi dan file lingkungan root-only. | DevOps Lead (pelaksana), Lead Information Security (penyetuju) | Tinggi | [usulan] 31 Januari 2027 | Keluaran dry-run sebelum dan sesudah (hitungan), uji login TOTP (EV-A8.5-04) |
| 8 | **Perbaiki provisioning tenant**: pastikan `MFA_ENCRYPTION_KEY` unik per tenant dibuat dan disuntikkan (template menetapkan `APP_APP_ENV: production`), tambahkan pemeriksaan bahwa `JWT_SECRET` dan kunci MFA tidak berulang antar-tenant. Flag ke `auditsphere-devops`; dokumen ini tidak mengubah template. | DevOps Lead, Solution Architect | Tinggi | [usulan] 31 Desember 2026 (sebelum tenant berikutnya) | Perubahan template/generator, uji provisioning (EV-A8.5-14) |
| 9 | **Sesi**: putuskan pemeriksaan sesi bersama untuk semua layanan atau masa JWT lebih pendek dengan refresh terkontrol; cabut sesi saat akun dinonaktifkan, kata sandi diganti atau direset, dan peran berubah; idle timeout mengikuti A.7.7 tindakan 6. | Solution Architect (desain), Backend Lead | Sedang | [usulan] 28 Februari 2027 | Desain tertulis, PR, uji (EV-A8.5-02) |
| 10 | **Alur reset dan undangan**: implementasikan reset mandiri dengan token sekali pakai berumur pendek ([30 menit]) dikirim ke email terdaftar tanpa mengungkap keberadaan akun, atau sembunyikan halaman `forgot-password`/`reset-password` sampai tersedia; ganti penyampaian kata sandi sementara dengan tautan undangan sekali pakai; kata sandi sementara kedaluwarsa. | Backend Lead, Frontend Lead | Sedang | [usulan] 28 Februari 2027 | PR dan uji; halaman sesuai fungsi (EV-A8.5-01) |
| 11 | **Hilangkan perbedaan respons dan waktu** untuk akun tidak dikenal dan nonaktif (galat seragam, bcrypt semu); uji dengan pihak berwenang. | Backend Lead | Sedang | [usulan] 31 Januari 2027 | PR, uji unit |
| 12 | **Tepercaya IP klien dan CORS**: gunakan daftar proxy tepercaya (`SetTrustedProxies`) atau satu header yang ditimpa nginx; ganti CORS auth-service dengan allow-list dari konfigurasi. | Backend Lead, Platform Engineer | Sedang | [usulan] 31 Januari 2027 | PR; konfigurasi nginx tersanitasi |
| 13 | **Kunci antar-layanan**: hapus nilai bawaan literal pada compose (wajib dari lingkungan), rotasi `PYTHON_AI_API_KEY` dan `DATA_HUB_API_KEY`, Data Hub API fail-closed dan waktu-konstan seperti python-ai; pertimbangkan kunci per pemanggil. Perubahan compose oleh `auditsphere-devops`, kode Data Hub oleh `auditsphere-data-analyst`. | DevOps Lead, Data Analyst (Data Hub) | Sedang | [usulan] 31 Januari 2027 | Perubahan compose dan `main.py`, catatan rotasi (EV-A8.5-15) |
| 14 | **Alert dan tinjauan log autentikasi**: alert untuk rentetan login gagal, penguncian, login ADMIN dari perangkat baru, penghapusan MFA, dan reset kata sandi; tinjauan [mingguan] memakai Lampiran C; tetapkan retensi log autentikasi [180 hari usulan] atau sesuai keputusan A.8.15. | DevOps Lead, Lead Information Security | Sedang | [usulan] 28 Februari 2027 | Aturan alert dan uji, catatan tinjauan (EV-A8.5-06, -12) |
| 15 | **Keputusan SSO**: tentukan apakah SSO/OIDC masuk roadmap (kebutuhan klien) dan nyatakan status jujur di SoA dan lembar jawab due diligence; bila ya, tugaskan Solution Architect untuk desain. | Solution Architect, Lead Information Security | Sedang | [usulan] 28 Februari 2027 | Keputusan tertulis |
| 16 | **Autentikasi operator**: inventarisasi akun istimewa (SSH, GitHub, registrar/DNS, Cloudflare, Backblaze, Resend, penyedia VPS), wajibkan MFA, nonaktifkan login kata sandi SSH, batasi root login, aktifkan MFA organisasi dan branch protection GitHub; kumpulkan tangkapan layar bertanggal tanpa rahasia. | DevOps Lead, Lead Information Security | Tinggi | [usulan] 31 Desember 2026 | Daftar akun istimewa dan bukti MFA (EV-A8.5-13) |
| 17 | **Sinkronkan dokumen dan daftarkan bukti**: koreksi ISO-AS-001 bagian 8 dan A.5.16 butir 5 agar sesuai kode, daftarkan Evidence ID pada Master Audit Evidence Index, dan commit dokumen ini bersama annex lain (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 15 Januari 2027 | Entri Evidence Index; riwayat git |

Catatan pelaksanaan: tindakan 2, 4, 6, 7, 8, 13, dan 16 menyentuh produksi, tenant, atau server lain; semuanya memerlukan konfirmasi eksplisit, dilakukan lewat agen devops/platform, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack, dan jalur deploy backend mengikuti `backend/deploy-prod.sh` tanpa seeder. Perubahan kode (tindakan 3, 4, 5, 9, 10, 11, 12) diajukan sebagai tiket ke agen backend/frontend; dokumen ini tidak mengubah kode.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.5-01 | Alur login, hashing, ganti kata sandi, reset admin (kode dan uji) | `backend/auth-service/services/auth/auth_service.go`, `backend/auth-service/pkg/utils/password.go`, `backend/auth-service/services/user/user_service.go` | Internal |
| EV-A8.5-02 | JWT (HS256 terpin), middleware dan pemeriksaan sesi, uji sesi | `backend/auth-service/pkg/utils/jwt.go`, `backend/auth-service/pkg/middleware/auth.go`, `backend/auth-service/pkg/middleware/session_test.go`, `backend/auth-service/cmd/auth/serve.go` | Internal |
| EV-A8.5-03 | MFA TOTP, penguncian MFA, perangkat tepercaya, uji | `backend/auth-service/services/mfa/mfa_service.go`, `backend/auth-service/services/auth/mfa_lockout.go`, `backend/auth-service/services/trusted-devices/` | Internal |
| EV-A8.5-04 | Enkripsi rahasia MFA (kode) dan runbook rollout; hasil dry-run setelah tindakan 7 | `backend/auth-service/pkg/secretcrypt/`, `backend/auth-service/pkg/config/mfa_encryption.go`, `docs/mfa-encryption-rollout.md`; [belum ada hasil dry-run] | Internal / Restricted |
| EV-A8.5-05 | Kebijakan `JWT_SECRET` dan runbook rotasi | `backend/auth-service/pkg/config/jwt_secret.go`, `docs/jwt-rotation-runbook.md` | Internal |
| EV-A8.5-06 | Log aktivitas autentikasi: kode pencatat, retensi, contoh ekspor bertanggal (setelah tindakan 14) | `backend/auth-service/pkg/activitylog/`; [belum ada contoh ekspor] | Internal / Restricted |
| EV-A8.5-07 | Catatan perubahan hardening backend | `docs/security-hardening-changes-backend.md` | Internal |
| EV-A8.5-08 | Standar autentikasi AUT-001 disetujui (setelah tindakan 1) | `docs/ISO_27001/A.8.5_Autentikasi_Aman.md` (Lampiran A, B); [belum ada versi disetujui] | Internal |
| EV-A8.5-09 | Ekspor read-only keadaan produksi: kunci MFA terpasang, hitungan rahasia MFA, cakupan MFA per peran (setelah tindakan 5-7) | [belum ada keluaran VPS] | Restricted |
| EV-A8.5-10 | Daftar akun seed dan catatan rotasi/penonaktifan (setelah tindakan 2) | [belum ada] | Restricted |
| EV-A8.5-11 | Hasil uji brute-force/penetrasi terkontrol pada login dan MFA (setelah tindakan 4) | [belum ada] | Restricted |
| EV-A8.5-12 | Catatan tinjauan log autentikasi berkala (Lampiran C) | [belum ada] | Internal / Restricted |
| EV-A8.5-13 | Daftar akun istimewa operator dan bukti MFA SSH/GitHub/konsol pemasok (setelah tindakan 16) | [belum ada] | Restricted |
| EV-A8.5-14 | Konfigurasi Kong (rate limit, CORS) dan template tenant | `backend/kong-gateway/kong/prod/kong.yml`, `scripts/templates/kong-tenant.yml.tpl`, `scripts/templates/docker-compose.tenant.yml.tpl` | Internal |
| EV-A8.5-15 | Autentikasi antar-layanan: python-ai dan Data Hub | `backend/python-ai/security.py`, `data-server/data-hub-api/main.py`, `backend/docker-compose.prod.yml` | Internal |

Setelah tindakan 1 sampai 8 selesai, tambahkan ID per siklus untuk setiap tinjauan log autentikasi dan tinjauan akses akun istimewa.

---

## Lampiran A — Draf Standar Autentikasi (AUT-001, usulan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**, bukan pengukuran, dan setiap nilai yang melampaui kode saat ini memerlukan tiket implementasi (bagian 5). Draf ini tidak menggantikan ISO-AS-001 bagian 8; keputusan manusia diperlukan untuk memilih salah satu atau menggabungkan keduanya.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | AUT-001 Standar Autentikasi (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden, perubahan arsitektur, atau permintaan SSO) |
| Berlaku untuk | Akun pengguna aplikasi AuditSphere (semua tenant, training, demo), akun layanan dan kunci antar-layanan, akun operator (SSH, GitHub, konsol pemasok), Data Hub |

1. **Tujuan.** Memastikan setiap akses ke AuditSphere dan infrastrukturnya diautentikasi dengan kekuatan yang sepadan risiko peran dan data, dan setiap peristiwa autentikasi dapat ditelusuri.
2. **Prinsip.** Tidak ada akun bersama; tidak ada kredensial default di lingkungan yang berisi data klien; kegagalan sistem autentikasi menolak akses (fail-closed); rahasia tidak disimpan di repo, tiket, atau pesan.
3. **Kata sandi pengguna aplikasi [usulan].**
   - Panjang minimum [12] karakter; panjang maksimum diizinkan sampai [64] karakter; semua karakter cetak diizinkan; tidak ada aturan komposisi wajib selain daftar tolak.
   - Ditolak: kata sandi pada daftar kata sandi umum/bocor [sumber daftar diputuskan], yang memuat nama pengguna, email, nama organisasi, atau "AuditSphere".
   - Riwayat [5] kata sandi terakhir tidak boleh dipakai ulang.
   - Kedaluwarsa berkala: [tidak ada kedaluwarsa terjadwal; ganti segera bila dicurigai bocor atau setelah reset admin] atau [90 hari sesuai ISO-AS-001] (keputusan).
   - Disimpan hanya sebagai hash adaptif (bcrypt cost [12] atau argon2id; nilai ditentukan setelah pengukuran waktu login); cost ditinjau [tahunan].
   - Kata sandi sementara dari admin: acak, sekali pakai, berlaku [24 jam], wajib diganti saat login pertama; penyampaian lewat kanal terpisah dari nama pengguna.
4. **Penguncian dan pembatasan [usulan].** Setelah [10] kegagalan kata sandi per akun dalam [15] menit: kunci [15] menit atau penundaan progresif; pembatasan per IP [isi]; kode MFA salah: [3] kali (sesuai kode) mengunci [15] menit; semua penguncian tercatat dan memicu alert (Lampiran C). Pesan galat seragam untuk akun tidak dikenal, nonaktif, atau kata sandi salah.
5. **MFA [usulan].** Wajib sesuai Lampiran B. Faktor yang diterima: TOTP (aplikasi autentikator); email OTP hanya untuk peran/lingkungan yang ditetapkan sebagai pengecualian tertulis. Kode pemulihan sekali pakai ([10] kode, disimpan sebagai hash) dan prosedur reset MFA oleh admin dengan verifikasi identitas, step-up, dan pencatatan. Rahasia MFA terenkripsi saat disimpan dengan kunci per tenant (`MFA_ENCRYPTION_KEY`); hilang kunci berarti enrolmen ulang. Perangkat tepercaya: [30] hari untuk peran non-istimewa; tidak berlaku untuk ADMIN.
6. **Sesi dan token [usulan].** JWT HS256 (atau penerus yang disetujui), masa berlaku [8 jam] atau sesuai desain refresh, pencabutan sesi saat logout, penonaktifan akun, perubahan atau reset kata sandi, dan perubahan peran; idle timeout [15] menit (A.7.7); token hanya di header `Authorization`; cookie dengan `Secure` dan `SameSite` [Strict/Lax] bila dipakai. `JWT_SECRET` minimal 64 byte, unik per tenant dan per lingkungan, dirotasi [tahunan] atau saat ada dugaan kebocoran (runbook rotasi).
7. **Akun istimewa dan step-up [usulan].** Aksi istimewa (ubah peran, reset kata sandi pengguna lain, reset MFA, hapus permanen, ubah struktur organisasi) memerlukan step-up kata sandi dan tercatat; admin tidak dapat mereset dirinya atau admin lain (sesuai kode); akun ADMIN dibatasi pada [jumlah minimum] dan ditinjau [triwulanan] (A.5.16).
8. **Akun layanan dan kunci antar-layanan.** Setiap kunci unik per pemanggil dan per lingkungan, tidak memiliki nilai bawaan di repo atau compose, diverifikasi dengan perbandingan waktu-konstan, diputar [tahunan] atau saat personel dengan akses keluar, dan disimpan di file lingkungan root-only atau pengelola rahasia.
9. **Akun operator.** MFA wajib pada GitHub (organisasi), registrar/DNS, Cloudflare, Backblaze, Resend, dan penyedia VPS; SSH hanya dengan kunci, login kata sandi dimatikan, root login dibatasi; kunci CI dibatasi pada satu fungsi dan diputar [tahunan]; pengelola kata sandi dipakai; akses ditinjau [triwulanan] dan dicabut pada hari personel keluar.
10. **SSO.** Status saat ini: tidak diimplementasikan. Bila klien mensyaratkan, diimplementasikan lewat OIDC/SAML dengan MFA pada IdP; akun lokal ADMIN darurat tetap ada dengan MFA dan tercatat (keputusan roadmap).
11. **Pencatatan.** Semua login berhasil dan gagal, logout, perubahan kata sandi, peristiwa MFA, penguncian, reset admin, dan perubahan perangkat tepercaya dicatat dengan waktu UTC, aktor, hasil, dan IP yang tervalidasi; retensi minimal [180] hari (keputusan, selaras A.8.15 draf); log tidak memuat kata sandi atau token; tinjauan sesuai Lampiran C.
12. **Lingkungan non-produksi.** Demo dan training boleh memakai akun contoh hanya dengan kata sandi unik per lingkungan, tidak dapat diakses dari internet tanpa perlindungan tambahan, dan dirotasi sebelum demo klien; kredensial demo tidak dicetak di skrip atau panduan.
13. **Peran.** Sesuai RACI Lampiran D.
14. **Pengecualian dan pelanggaran.** Pengecualian tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir. Dugaan kredensial bocor diperlakukan sebagai insiden keamanan (A.5.24): reset, cabut sesi, tinjau log.

---

## Lampiran B — Usulan MFA Wajib per Peran dan Fase

**Seluruh keputusan adalah [usulan]** dan memerlukan persetujuan manusia. Peran nyata yang di-seed auth-service: ADMIN, AUDITOR, EXECUTIVE, AUDITEE, VIEWER, DEPARTMENT_HEAD. AUDIT_MANAGER dan CHIEF_AUDIT_EXECUTIVE dirujuk frontend tetapi hanya ada bila admin membuatnya (periksa daftar peran per tenant). Perbandingan peran dilakukan tanpa membedakan huruf besar-kecil.

| Peran | Alasan risiko | MFA [usulan] | Fase [usulan] | Perangkat tepercaya [usulan] |
| :--- | :--- | :--- | :--- | :--- |
| ADMIN | Kendali penuh akun, peran, reset kata sandi, log | **Wajib** | 1 (paling lambat 31 Januari 2027) | Tidak berlaku (MFA setiap login) |
| AUDIT_MANAGER, CHIEF_AUDIT_EXECUTIVE (bila ada) | Menyetujui rencana dan laporan; ISO-AS-001 menyebut wajib | **Wajib** | 1 | [30] hari |
| AUDITOR | Menulis KKA, LHA, dan mengunggah bukti klien | **Wajib** | 2 ([28 Februari 2027]) | [30] hari |
| EXECUTIVE | Membaca ringkasan dan dasbor eksekutif data sensitif | **Wajib** | 2 | [30] hari |
| DEPARTMENT_HEAD | Menyetujui ATR dan data departemen | **Wajib** | 3 ([31 Maret 2027]) | [30] hari |
| AUDITEE | Mengisi tindak lanjut dan mengunggah bukti | **Dianjurkan**; wajib bila tenant memilih | 3, per tenant | [30] hari |
| VIEWER | Baca saja (allow-list rute) | **Dianjurkan**; wajib bila tenant memilih | 3, per tenant | [30] hari |
| Akun demo/training | Berisi data contoh atau klon produksi | Mengikuti peran; pengecualian tertulis bila demo | Per kebutuhan | n/a |

Prasyarat sebelum MFA wajib berlaku: (1) kode pemulihan dan rute reset MFA oleh admin tersedia dan teruji; (2) enkripsi rahasia MFA di produksi atau keputusan tertulis menunda; (3) backup terenkripsi dan restore test (A.8.13); (4) masa transisi dengan pemberitahuan pengguna dan alur enrolmen pada login pertama; (5) kebijakan akun tanpa perangkat (mis. akun layanan atau demo) diputuskan; (6) email OTP diputuskan sebagai faktor yang diterima atau tidak; (7) hitungan akun per peran sebelum dan sesudah (EV-A8.5-09). Kecuali kode diubah, MFA wajib per peran tidak dapat ditegakkan; sampai itu, kendali sementara adalah kebijakan tertulis dan pemeriksaan bulanan daftar ADMIN tanpa MFA.

---

## Lampiran C — Daftar Periksa Tinjauan Log Autentikasi (usulan)

Frekuensi tinjauan: [harian oleh DevOps Lead untuk alert; mingguan ringkasan; bulanan ke Lead Information Security]. Setiap tinjauan dicatat (tanggal, peninjau, hasil). Sumber: halaman log aktivitas admin atau `GET /api/v1/system-logs`.

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Jumlah login gagal per hari dibandingkan baseline; tidak ada lonjakan di atas [ambang: isi]. | [isi] | |
| 2 | Tidak ada satu akun dengan kegagalan beruntun lebih dari [5] kali, dan tidak ada satu IP yang menyasar banyak nama pengguna ([5]+ dalam [15] menit). | [isi] | |
| 3 | Setiap `account_locked_mfa` diselidiki: pemilik akun dikonfirmasi, bukan percobaan pengambilalihan. | [isi] | |
| 4 | Login ADMIN dari perangkat baru atau di luar jam kerja dikonfirmasi dengan pemilik akun. | [isi] | |
| 5 | Peristiwa `mfa.disabled`, penghapusan perangkat tepercaya, `password.changed` gagal, dan reset kata sandi admin dicocokkan dengan tiket atau permintaan sah. | [isi] | |
| 6 | Tidak ada login dengan akun seed/demo di lingkungan berisi data klien; semua akun seed tercatat dan dirotasi (EV-A8.5-10). | [isi] | |
| 7 | Daftar ADMIN dan akun istimewa sesuai daftar yang disetujui; tidak ada akun baru tanpa persetujuan (A.5.16). | [isi] | |
| 8 | Cakupan MFA per peran sesuai Lampiran B (hitungan ADMIN/AUDITOR tanpa MFA = [0] setelah fase 1/2). | [isi] | |
| 9 | Log tidak memuat kata sandi, token, atau kode MFA; entri dengan nama pengguna aneh (kemungkinan kata sandi terketik di kolom nama pengguna) dihapus dari tampilan dan dilaporkan. | [isi] | |
| 10 | IP pada log tampak valid (bukan daftar atau nilai dipalsukan); anomali dilaporkan ke tindakan 12. | [isi] | |
| 11 | Pencatatan berfungsi: tidak ada jeda tak wajar dalam peristiwa login pada jam sibuk; Kafka dan konsumer sehat (fallback tulis langsung tidak aktif terus-menerus). | [isi] | |
| 12 | Retensi berjalan sesuai keputusan; tidak ada pembersihan yang menghapus log sebelum jangka retensi. | [isi] | |
| 13 | Alert autentikasi teruji kuartal ini (kegagalan beruntun, penguncian, ADMIN perangkat baru). | [isi] | |
| 14 | Temuan dieskalasi sebagai insiden sesuai A.5.24; tiket/insiden: [isi]. | [isi] | |
| 15 | Tinjauan akses akun operator (SSH, GitHub, konsol pemasok) dilakukan [triwulan] ini; tanggal terakhir: [isi]. | [isi] | |

---

## Lampiran D — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | Backend Lead | Frontend Lead | DevOps Lead | Platform Engineer | Solution Architect | Legal & Compliance | Direktur Utama | Data Analyst (Data Hub) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan AUT-001 | A/R | C | C | C | I | C | C | A (persetujuan) | I |
| Keputusan MFA wajib per peran | A/R | C | C | C | I | C | C | A (persetujuan) | I |
| Implementasi kebijakan kata sandi dan penguncian | C | A/R | R (pesan UI) | I | I | C | I | I | I |
| Implementasi MFA wajib, kode pemulihan, reset MFA | A | R | R | I | I | C | I | I | I |
| Enkripsi rahasia MFA dan kunci per tenant | A | C | I | R | I | C | I | I | I |
| Rotasi `JWT_SECRET` dan kunci antar-layanan | A | C | I | R | I | C | I | I | R (Data Hub) |
| Rotasi dan penonaktifan akun seed | A | R (alur seeder) | I | R (operasi) | I | C | I | I | I |
| Rate limit login (Kong) dan IP tepercaya (nginx) | C | C | I | A/R | R | I | I | I | I |
| Desain sesi, pencabutan, dan refresh | C | R | C | I | I | A | I | I | I |
| Alur reset dan undangan | C | A/R | R | I | I | C | I | I | I |
| Keputusan dan desain SSO | C | C | C | I | I | A/R | C | A (persetujuan) | I |
| Autentikasi operator (SSH, GitHub, konsol pemasok) | A | I | I | R | R | I | I | I | C |
| Alert dan tinjauan log autentikasi | C | C | I | A/R | R | I | I | I | C |
| Penanganan insiden kredensial | A | R | I | R | C | C | C | I | C |
| Evidence ID dan review berkala | A/R | C | C | C | I | I | C | I | I |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.5.16 Manajemen identitas | Ada (`docs/ISO_27001/A.5.16_Manajemen_Identitas.md`) | Siklus akun, akun default, sesi saat leaver; sebagian butir usang (bagian 4 butir 20) dan perlu disinkronkan. |
| A.7.7 Meja bersih dan layar bersih | Ada (`docs/ISO_27001/A.7.7_Meja_Bersih_dan_Layar_Bersih.md`; belum ter-commit) | Idle timeout, cookie 1 atau 30 hari, perangkat tepercaya 90 hari; **tidak diulang** di sini. |
| A.8.13 Pencadangan informasi | Ada (belum ter-commit) | Backup memuat rahasia MFA; syarat sebelum migrasi enkripsi MFA (tindakan 7). |
| A.8.17 Sinkronisasi waktu | Ada (belum ter-commit) | TOTP dan JWT bergantung jam server. |
| A.8.21 Keamanan layanan jaringan | Ada (belum ter-commit) | Port, nginx, tanpa alert Prometheus (butir 16), Kong tanpa validasi JWT. |
| A.5.12 Klasifikasi informasi | Ada | Klasifikasi data yang dilindungi autentikasi. |
| A.5.24 Persiapan manajemen insiden | Ada | Eskalasi dugaan kredensial bocor. |
| A.5.28 Pengumpulan bukti | Ada | Keandalan IP dan stempel waktu pada log login. |
| A.8.25 Siklus hidup pengembangan yang aman | Ada (belum ter-commit) | Tiket perubahan kode autentikasi mengikuti SDLC dan uji. |
| A.5.17 Informasi autentikasi | **Draf, dokumen belum ada** | Kebijakan kata sandi dan penyampaian kredensial; AUT-001 Lampiran A dapat menjadi masukan. |
| A.5.15 Kontrol akses | **Draf, dokumen belum ada** | Peran dan RBAC. |
| A.5.18 Hak akses | **Draf, dokumen belum ada** | Tinjauan akses akun istimewa. |
| A.8.2 Hak akses istimewa | **Draf, dokumen belum ada** | MFA dan step-up bagi ADMIN. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Retensi dan integritas log autentikasi. |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Alert kegagalan login dan penguncian. |
| A.8.24 Penggunaan kriptografi | **Draf, dokumen belum ada** | Hashing kata sandi, enkripsi rahasia MFA, manajemen `JWT_SECRET` dan kunci. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.5	Technological	Autentikasi aman	Mekanisme autentikasi sesuai risiko dan terlindungi	MFA config, password policy, SSO setting, authentication log	Berlaku	Autentikasi adalah pintu ke data audit klien lembaga jasa keuangan dan dasar atribusi jejak audit; peran ADMIN berdaya besar; multi-tenant dengan secret per tenant; klien akan menanyakan MFA, kebijakan kata sandi, SSO, dan log login (lihat A.8.5 bagian 1)	Sebagian	Sebagian (kode, uji unit, runbook ada); MFA config sebagian, password policy tidak ada, SSO setting tidak ada (fitur tidak ada), authentication log sebagian (dihasilkan, belum ditinjau); konfigurasi produksi belum terbukti	Belum terdaftar (usulan EV-A8.5-01 s/d EV-A8.5-15)	Lead Information Security (didukung Backend Lead dan DevOps Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah keputusan MFA wajib dan rotasi akun seed selesai	Perlu perbaikan	Belum teruji	MFA opsional untuk semua peran termasuk ADMIN; kata sandi minimum 6 karakter di backend tanpa riwayat/kedaluwarsa/cek kebocoran; tanpa penguncian atau pembatasan tebakan kata sandi per akun (hanya MFA 3 kali dan step-up 5 kali); tanpa kode pemulihan dan tanpa rute reset MFA oleh admin; rahasia MFA belum terbukti terenkripsi di produksi dan template tenant berpotensi gagal start tanpa MFA_ENCRYPTION_KEY; akun seed berkata sandi default sama; logout dan penonaktifan tidak berlaku di layanan selain auth-service; reset mandiri tidak ada walau halaman UI ada; CORS auth-service memantulkan origin apa pun; IP klien dari X-Forwarded-For tanpa validasi; literal kunci antar-layanan di compose; tanpa SSO; tanpa alert dan tinjauan log; autentikasi operator tidak terbukti	Setujui AUT-001; rotasi akun seed; kebijakan kata sandi dan penguncian di server; MFA wajib ADMIN dengan kode pemulihan dan reset MFA; verifikasi produksi read-only; enkripsi rahasia MFA; perbaikan provisioning tenant; sesi lintas layanan; reset dan undangan; kunci antar-layanan; alert dan tinjauan log; keputusan SSO; autentikasi operator (17 tindakan di bagian 5)	Tinggi	[usulan] 31 Desember 2026 (standar, rotasi akun seed, verifikasi produksi); 31 Januari 2027 (MFA wajib ADMIN, kebijakan kata sandi, penguncian)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 15 Januari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
