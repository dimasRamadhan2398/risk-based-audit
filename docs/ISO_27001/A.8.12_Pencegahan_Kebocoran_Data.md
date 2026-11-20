# A.8.12 — Pencegahan Kebocoran Data

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Pencegahan kebocoran data |
| **Fokus pemeriksaan** | Kanal kebocoran dipantau dan dikendalikan sesuai klasifikasi |
| **Contoh bukti objektif** | DLP policy, rule set, alert log, incident ticket |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi, stack tenant, training, Data Hub, Resend (email), Backblaze B2 & Cloud SaaS |
| **PIC** | Lead Information Security (didukung DevOps Lead, Backend Lead, dan Legal & Compliance) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`, termasuk perubahan working tree yang belum ter-commit) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut kanal kebocoran **dipantau** dan **dikendalikan** sesuai klasifikasi. Repo membuktikan **kode, konfigurasi, dan skrip**; repo **tidak** membuktikan apa yang berjalan di VPS (versi yang ter-deploy, aturan firewall, isi bucket, isi log), sehingga semua itu ditandai **Perlu verifikasi**. Tidak ada akses SSH dan tidak ada akses konsol penyedia yang dipakai dalam penyusunan dokumen ini. Butir yang bersumber dari catatan operasional internal (audit read-only 5 Nov 2026, catatan 10 Nov 2026) dicantumkan sumbernya dan **bukan bukti repo**. **Tidak ada alat DLP, kebijakan DLP, rule set, log alert kebocoran, atau kategori tiket insiden kebocoran data yang ditemukan di repo**; kontrol yang ada bersifat pencegahan tersebar (autentikasi, peran, pemindaian malware, pembatasan akses berkas) dan bukan program DLP. Seluruh angka dan ambang pada Lampiran A sampai D adalah **[usulan]** yang belum disetujui manajemen. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

Seluruh nilai di bawah adalah **usulan dan menunggu validasi pengguna**. Di spreadsheet (`03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx`, tab `Checklist 93 Kontrol`), baris A.8.12 saat ini berstatus "Belum ditentukan", PIC "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", dan jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran sebagai usulan yang perlu dikonfirmasi; spreadsheet tidak diubah.

| Kolom checklist | Nilai (usulan) | Catatan singkat |
| :--- | :--- | :--- |
| Kontrol | **A.8.12** | Sesuai baris spreadsheet. |
| Tema | **Technological** | Sesuai baris spreadsheet. |
| Judul kontrol | **Pencegahan kebocoran data** | Sesuai baris spreadsheet. |
| Fokus pemeriksaan | **Kanal kebocoran dipantau dan dikendalikan sesuai klasifikasi** | Tiga gate: kanal terdaftar, dikendalikan sesuai klasifikasi, dipantau dengan alert dan tindak lanjut. |
| Contoh bukti objektif | **DLP policy, rule set, alert log, incident ticket** | Status tiap bukti di bagian 3. |
| Relevansi SoA | **Berlaku** | AuditSphere menyimpan data audit klien lembaga jasa keuangan dan dapat mengekspor, mengunduh, mengirim email, dan menyinkronkan data keluar melalui beberapa kanal (bagian 1). Pengecualian tidak dapat dipertahankan. |
| Justifikasi relevansi | **Lihat bagian 1** | Ringkas: data klien bernilai tinggi (Confidential/Restricted, A.5.12), multi-tenant, banyak kanal keluar (unduh, ekspor, email, Data Hub, backup off-site, training), dan klien akan menanyakan kendali kebocoran. |
| Status implementasi | **Sebagian (minimal)** | Ada kendali pencegahan tersebar: `/uploads` wajib autentikasi, peran VIEWER dilarang mengakses berkas/ekspor, unggahan dipindai ClamAV (mode `required`), CORS Kong terbatas, rate limiting Kong, `password_hash` dan secret MFA tidak diserialisasi, redaksi query pada log. Belum ada: kebijakan DLP, alat DLP, rule set, alert, pencatatan unduhan/ekspor, penanda klasifikasi, watermark, masking, kendali endpoint/USB. Ada kebocoran nyata pada API (kata sandi sumber data, bagian 4 butir 2). |
| Ketersediaan bukti | **Sebagian (kode dan konfigurasi) / Tidak ada (kebijakan, rule set, alert log, tiket)** | Bukti kendali teknis ada di repo; bukti operasional DLP **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.12-01 s/d EV-A8.12-12) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| PIC | **Lead Information Security** (didukung **DevOps Lead**, **Backend Lead**) | Berbasis peran, bukan nama pribadi. PIC pada spreadsheet tercatat "Qiko"; konfirmasi peran oleh pengguna. |
| Review terakhir | **19 November 2026** (usulan; tinjauan repo oleh penyusun dokumen, bukan validasi reviewer) | Tidak ada tinjauan sebelumnya yang tercatat untuk kontrol ini. |
| Review berikutnya | **[usulan] 19 Februari 2027**, atau lebih awal bila kebijakan DLP disetujui dan perbaikan kebocoran kredensial sumber data dirilis | Selanjutnya: tinjauan alert [mingguan] dan kebijakan [tahunan] (Lampiran A, C). |
| Desain kontrol | **Perlu perbaikan** | Tidak ada daftar kanal kebocoran, tidak ada matriks kanal x klasifikasi, tidak ada penanda klasifikasi pada data, tidak ada pencatatan unduhan/ekspor, dan tidak ada pemantauan. Kendali yang ada tidak diturunkan dari klasifikasi. |
| Efektivitas operasional | **Belum teruji** | Tidak ada alert log, tiket, atau uji kebocoran. Kendali teknis berbukti unit test sebagian (mis. `controller_test.go` media), tetapi tidak ada uji kanal secara menyeluruh dan status ter-deploy di VPS tidak terbukti. |
| Catatan / gap | **Lihat bagian 4** | Gap utama: tanpa kebijakan dan alat DLP; kata sandi sumber data tersimpan dan dikembalikan oleh API; unduhan/ekspor tidak tercatat; tanpa alert (Prometheus tanpa `rule_files`/Alertmanager); backup off-shore dan tidak terenkripsi menurut catatan; training memuat klon data produksi. |
| Tindakan perbaikan | **Lihat bagian 5 (13 tindakan)** | Draf Lampiran A sampai D sebagai masukan. |
| Prioritas | **Tinggi** | Klien lembaga jasa keuangan akan menanyakan kendali kebocoran; ada kebocoran kredensial sumber data pada respons API yang dapat diperbaiki cepat; tanpa pencatatan unduhan tidak ada jejak bila kebocoran terjadi. |
| Target | **[usulan] 31 Januari 2027** untuk kebijakan DLP disetujui, perbaikan kebocoran API, dan pencatatan unduhan; **[usulan] 31 Maret 2027** untuk alert dan tinjauan berkala | Keputusan pengguna. |
| Reviewer | **[isi saat validasi]** | Diisi reviewer saat validasi; bukan penyusun dokumen. |
| Tanggal validasi | **19 November 2026** (verifikasi repo); validasi reviewer **[isi saat validasi]** | Tanggal verifikasi repo bukan tanggal validasi reviewer. |
| Pemeriksaan field | **Belum lengkap** | Evidence ID belum terdaftar, Reviewer dan validasi reviewer kosong, Target dan Deadline masih usulan. Usulan nilai kolom: "Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)". |
| Kesiapan | **Belum siap** | Gate "dipantau" dan "sesuai klasifikasi" belum terpenuhi; tidak ada rule set maupun alert log. |
| Deadline | **[usulan] 28 Februari 2027** | Batas kesiapan: kebijakan disetujui, kebocoran API diperbaiki, pencatatan unduhan aktif, tinjauan pertama terdokumentasi. Keputusan pengguna. |
| Perlu tindakan | **Ya** | Tindakan 1 sampai 5 bersifat segera (bagian 5). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.12 **berlaku** untuk AuditSphere karena:

1. **Data yang ditangani bernilai tinggi.** Sistem memuat kertas kerja, bukti temuan, laporan audit (LHA), ATR, hasil penilaian pengendalian internal, dan kredensial sumber data klien. Dokumen ISO-AS-001 menetapkan data audit klien sebagai *Confidential / Restricted* dan kredensial sebagai *Restricted* (A.5.12 bagian 2). Kebocoran satu tenant merugikan klien lembaga jasa keuangan secara langsung.
2. **Banyak kanal keluar yang sah.** Aplikasi memang dirancang mengeluarkan data: unduh berkas bukti dan dokumen, ekspor DOCX/XLSX/PDF, cetak laporan dari browser, email lewat Resend, sinkronisasi ke Data Hub, backup off-site, dan klon data untuk training (bagian 2). Setiap kanal perlu dikendalikan dan, bila memungkinkan, dicatat.
3. **Multi-tenant dalam satu host.** Stack tenant berbagi Postgres, Redis, Kafka, ClamAV, dan python-ai (A.8.21 bagian 4 butir 5). Kebocoran antar-tenant adalah bentuk kebocoran data yang paling mahal.
4. **Pernah ada insiden paparan.** Catatan operasional mencatat `/uploads` publik tanpa login (8 Okt 2026), port layanan terbuka lewat IPv6 (5 Nov 2026), dan Grafana/Prometheus/Loki sempat terbuka. Perbaikan ada di repo, tetapi status VPS **Perlu verifikasi** (bagian 4 butir 7).
5. **Klien dan auditor akan menanyakan.** Due diligence lembaga jasa keuangan lazim meminta kebijakan DLP, daftar kanal, dan bukti pemantauan. Saat ini tidak ada dokumen atau bukti yang dapat diserahkan.

Pengecualian "tidak berlaku" tidak dapat dipertahankan. Ukuran organisasi yang kecil membenarkan pendekatan bertahap (kendali aplikasi dan pencatatan dahulu, alat DLP komersial kemudian), bukan pengecualian.

---

## 2. Implementasi yang Sudah Ada (per Kanal Kebocoran)

Semua butir adalah **bukti kode atau konfigurasi di repo**. Apakah versi itu berjalan di VPS adalah **Perlu verifikasi** kecuali dinyatakan lain. Kolom "Jenis" memakai: **Pencegahan** (membatasi), **Deteksi** (memantau), **Tidak ada**.

### 2.1 Unduh berkas dan ekspor dokumen

| Kanal | Yang terlihat di repo | Bukti | Jenis / Keterbatasan |
| :--- | :--- | :--- | :--- |
| Unduh dokumen unggahan (charter, kertas kerja impor, rencana, surat tugas, laporan hasil audit, ringkasan eksekutif, dokumen konsultasi, laporan kinerja) | Rute `GET /:id/download` per modul di bawah grup `apiV1` yang wajib token; peran VIEWER dilarang pada rute unduh/ekspor/berkas. | `backend/audit-service/routes/handler.go:76`, `:86`, `:335`, `:347`, `:359`, `:371`, `:383`, `:395`, `:407`, `:419`, `:431`, `:443` | Pencegahan. Setiap pengguna non-VIEWER terautentikasi dapat mengunduh; tidak ada pemeriksaan klasifikasi per berkas. |
| Ekspor laporan hasil audit DOCX | `GET /:id/download-docx` membangun DOCX dan mengirim sebagai lampiran. | `audit-service/routes/handler.go:454`, `:794-795` | Pencegahan (autentikasi + VIEWER ditolak). Tanpa pencatatan peristiwa ekspor. |
| Ekspor PDF laporan kinerja | `GET /performance/export-pdf`; label "Confidential" ditulis tetap pada templat. | `audit-service/routes/handler.go:535`; A.5.12 bagian 2 | Label statis, bukan turunan klasifikasi. |
| Ekspor/impor Excel penilaian pengendalian internal (PoF) dan unduh bukti PoF | Rute `pof-export`, `pof-export-template`, unduh bukti dan unduh impor; ekspor memakai `excelize`; impor dibatasi ukuran unzip. | `audit-service/controllers/ic_assessment/pof_handler.go:29`, `:44`, `:46`, `:51`, `:431`; `services/ic_assessment/pof_import_parse.go:222` | Pencegahan sebagian. File baru, belum ter-commit (status git `??`). |
| Unduh bukti ATR | Dilayani hanya lewat API terautentikasi, `Content-Disposition: attachment`, `Cache-Control: private, no-store`, `X-Content-Type-Options: nosniff`. Direktori `action-taken-reports` privat. | `audit-service/controllers/action_taken_report/evidence.go:290-298`; `audit-service/pkg/safefile/safefile.go:27` | Pencegahan baik. |
| Unduh media | `GET /media/download/:id` di bawah `apiV1` (wajib token); jalur dikonfirmasi di dalam root unggahan; direktori privat ditolak; tipe konten allowlist, aktif-konten dikirim sebagai lampiran. | `audit-service/routes/handler.go:519-523`; `audit-service/controllers/media/controller.go:88-100`; uji `controller_test.go:56-153` | Pencegahan. Memori internal mencatat sebelumnya rute media di luar autentikasi; di repo kini wajib token (status VPS: **Perlu verifikasi**). |
| Unduh laporan QA dan dokumen konsultasi (master-service) | `qa/:id/download`, `view`, dan `consulting/:id/download` mengirim berkas dari DB atau disk; header `nosniff`. | `backend/master-service/routes/routes.go:344-379`, `:401-415`, `:612-649` | Pencegahan sebagian. master-service dilindungi `DenyViewer` (`routes.go:44`). Berkas master tidak persisten dan tidak dicadangkan (A.8.13 bagian 4 butir 6). |
| Cetak dari browser | Modal cetak laporan hasil audit memanggil `window.print()`. | `frontend/components/audit-result-report/ResultReportPrintModal.vue:319` | **Tidak ada kendali**. Cetak/PDF dari browser tidak dapat dicegah atau dicatat oleh aplikasi. |
| Watermark dokumen | Tidak ditemukan. Istilah "watermark" pada repo adalah penanda sinkronisasi Data Hub (`ops.sync_watermarks`), bukan watermark dokumen. | pencarian repo; `data-server/data-hub-api/routers/sources.py:175-182` | **Tidak ada**. |
| Masking/anonimisasi data | Tidak ditemukan fungsi masking atau anonimisasi data pribadi di backend atau frontend. | pencarian `backend/`, `frontend/`, `data-server/` | **Tidak ada**. |

### 2.2 Akses berkas `/uploads` dan unggahan

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| `/uploads` wajib autentikasi | `RegisterUploads` memasang `authenticate` dan `DenyViewer` pada GET/HEAD; jalur dikonfirmasi (`ResolveUnder`), direktori privat 404, tanpa daftar direktori, tipe konten allowlist. Menggantikan `engine.Static` publik (commit `b960dd0`). | `backend/audit-service/routes/uploads.go:13-37`; `audit-service/cmd/serve.go:252-254`; `audit-service/pkg/safefile/safefile.go:85-104`, `:150-157` | Memori internal (8 Okt 2026) mencatat `/uploads` terbuka tanpa login; repo kini menutupnya. **Perlu verifikasi** bahwa versi ter-deploy sudah memuatnya (uji tanpa token harus 401). Tidak ada pemeriksaan kepemilikan/penugasan per berkas, hanya autentikasi dan peran. |
| Rute proxy | Kong dan Nuxt meneruskan `/uploads` ke audit-service. | `backend/kong-gateway/kong/prod/kong.yml:452`; `frontend/nuxt.config.ts` (`routeRules` `/uploads/**`) | Tidak mengubah autentikasi. |
| Pemindaian malware unggahan | ClamAV (clamd, INSTREAM) memindai byte sebelum ditulis; mode `required` bersifat fail-closed; prod memakai `CLAMAV_MODE=required`. | `audit-service/pkg/avscan/avscan.go:1-16`; `audit-service/pkg/upload/upload.go:10`, `:333-346`; `master-service/pkg/avscan/avscan.go`; `backend/docker-compose.prod.yml:181-182`, `:205-208` | Ini **kendali malware (A.8.7)**, bukan DLP: tidak memeriksa isi untuk data pribadi/rahasia pada unggahan atau unduhan. |
| Persistensi dan lokasi berkas | Uploads audit di volume; uploads master tanpa volume; uploads tenant bind mount host. | A.8.13 bagian 2.1 | Berkas bukti dapat hilang pada recreate; bukan kebocoran, tetapi memengaruhi integritas rantai bukti. |

### 2.3 Peran, RBAC, dan viewer read-only

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| VIEWER tidak boleh menulis, mengakses berkas, atau mengekspor (audit-service) | `DenyViewerWrites` (deny-by-default untuk non-GET) dan `DenyViewerFiles` (rute dengan segmen `view`, `evidence`, `print`, atau mengandung `download`/`export`, serta `/uploads` dan `/api/v1/media`). Caller dengan peran lain tidak terbatasi. | `audit-service/pkg/middleware/viewer.go:43-135`; terpasang `audit-service/routes/handler.go:76` | Pencegahan baik. Pencocokan berdasarkan nama rute: rute ekspor baru yang tidak memakai kata `download`/`export` tidak otomatis tercakup. |
| VIEWER pada layanan lain | master-service: `DenyViewer` penuh. risk-service dan analytics-service: hanya `DenyViewerWrites` (VIEWER dapat membaca). | `master-service/routes/routes.go:44`; `risk-service/routes/routes.go:63`; `analytics-service/routes/routes.go:22` | Risiko: data risiko/analitik dapat dibaca VIEWER. Apakah ada ekspor pada layanan itu: tidak diperiksa (**Perlu verifikasi**). |
| Tidak ada pembatasan per klasifikasi | Tidak ada atribut klasifikasi pada model data. | A.5.12 bagian 4 | Pengendalian "sesuai klasifikasi" tidak dapat diterapkan secara teknis. |

### 2.4 Isolasi tenant

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Stack per tenant | Lima layanan Go, Kong, dan frontend per tenant; `JWT_SECRET` per tenant; database per layanan pada Postgres bersama; indeks Redis logis per tenant. | `scripts/templates/docker-compose.tenant.yml.tpl:5-8`; A.8.21 bagian 2.1 dan bagian 4 butir 5 | Isolasi bergantung pada konfigurasi (env, nama layanan), bukan segmentasi jaringan. Memori internal mencatat gotcha isolasi (URL default menunjuk produksi, kunci env berbagi). Status ter-deploy: **Perlu verifikasi**. Satu `pg_dumpall` memuat semua tenant (A.8.13). |

### 2.5 Email (Resend) dan notifikasi

| Kanal | Yang terlihat di repo | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Pengingat ATR (audit-service) | Mailer ke `https://api.resend.com/emails`; payload hanya `from`, `to`, `subject`, `text` (teks polos, tanpa lampiran); `RESEND_API_KEY` kosong membuat fitur nonaktif. Penerima ditentukan dari peran pada ATR/LHA, bukan input bebas. | `audit-service/pkg/mailer/mailer.go:24`, `:26-28`, `:56-57`, `:62-67`; `controllers/action_taken_report/reminder.go:179`, `:197`, `:209-233` | Isi pesan (`reminderMessage`) tidak ditinjau dalam dokumen ini: **Perlu verifikasi** apakah memuat data sensitif. Tidak ditemukan allowlist domain penerima. Pengiriman tidak dicatat sebagai peristiwa aktivitas. |
| Email OTP/auth (auth-service) | Mengirim lewat Resend dengan retry. | `auth-service/services/email/email_service.go:24`, `:63`, `:148` | Isi dan lampiran tidak diperiksa: **Perlu verifikasi**. Catatan operasional: OTP dikirim lewat Resend; training memakai kunci Resend yang sama sehingga OTP training terkirim ke pengguna nyata (memori "training environment"). |
| Pihak ketiga | Resend (AS) memproses alamat email dan isi pesan. | A.8.21 bagian 2; A.5.20 | Kontrak/DPA: belum ada (A.5.20). |

### 2.6 AI dan LLM keluar

| Kanal | Yang terlihat di repo | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| python-ai | Model lokal (scikit-learn/PyTorch/TensorFlow, berkas model di volume). Pencarian `openai`, `anthropic`, dan panggilan HTTP eksternal pada `backend/python-ai` tidak menemukan pemanggilan ke API AI eksternal; permintaan keluar hanya ke layanan internal (risk/audit/master/data-hub). Semua rute kecuali `/health` wajib `X-API-Key` (fail-closed). | `backend/python-ai/retrain_engine.py:94-97`, `:108`; `python-ai/security.py:1-40`; `python-ai/main.py:71` | Tidak ada bukti data audit dikirim ke LLM eksternal dari backend. Jalur retrain: Kong 10/menit (`kong.yml:650-660`). |
| Handler OpenAI di frontend | **Pada HEAD** `frontend/server/api/openai.ts` meneruskan `body.prompt` ke OpenAI (`gpt-3.5-turbo-instruct`) memakai `OPENAI_API_KEY`. **Di working tree berkas ini dihapus (status `D`, belum ter-commit).** Tidak ditemukan pemanggil `/api/openai` di `frontend/`. Paket `openai` masih ada pada `frontend/package.json:36`. | `git show HEAD:frontend/server/api/openai.ts`; `frontend/package.json:36` | Selama penghapusan belum ter-commit dan ter-deploy, endpoint dapat meneruskan teks yang dikirim pengguna ke pihak ketiga bila `OPENAI_API_KEY` terpasang di server (**Perlu verifikasi**; memori internal mencatat kunci itu tidak diperlukan). Tidak ada kebijakan penggunaan AI (A.6.7 menyebut panduan agen AI belum ada). |

### 2.7 Data Hub dan sumber data eksternal

| Kanal | Yang terlihat di repo | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Registrasi sumber data | master-service menyimpan koneksi sumber data (host, kredensial, SSL) dan mendaftarkannya ke Data Hub, termasuk kata sandi dalam payload. | `master-service/models/data_source.go:60-80`; `master-service/routes/routes.go:1400-1416` | Lihat 4 butir 2: kata sandi tersimpan tanpa enkripsi dan dikembalikan oleh API. |
| SSL ke sumber data | Mode SSL tervalidasi; koneksi ditolak bila server tidak menerima SSL pada mode yang mewajibkan. | `master-service/pkg/datasource/ssl.go:188`; `master-service/routes/data_source_ssl.go` | Pencegahan baik untuk transit. |
| Data Hub list | Menghapus `password_encrypted` dan PEM dari daftar sumber. | `data-server/data-hub-api/routers/sources.py:112-120` | Kolom bernama `password_encrypted` diisi langsung dari kata sandi mentah dan dipakai apa adanya (`core/db.py:145-146`, `:183`); tidak ditemukan kode enkripsi. Nama menyesatkan. |
| Egress data klien | Sinkronisasi menarik data klien (mis. CBS) ke lakehouse di server terpisah; training menonaktifkan `DATA_HUB_URL` pada master-service. | memori "training environment"; `data-server/docs/DEPLOY_DATAHUB_SAFE.md` | Kebijakan lokasi dan pemilik data pada deployment klien: **Perlu verifikasi**. |

### 2.8 Pencatatan log, respons API, CORS, rate limiting

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Redaksi query pada log akses | `RedactQuery` mengganti nilai parameter sensitif; log tidak memuat token. master/risk/analytics memakai `RedactingLogger`. | `audit-service/pkg/middleware/redact.go:17-36`; `audit-service/pkg/middleware/logger.go:12-35`; `master-service/cmd/serve.go:139`; `risk-service/cmd/serve.go:109`; `analytics-service/cmd/main.go:27` | `LoggerMiddleware` mencatat jalur, IP, dan user agent, **tanpa identitas pengguna**; log akses bukan jejak unduhan per pengguna. auth-service juga memakai `LoggerMiddleware` (`auth-service/cmd/auth/serve.go:181`). |
| Log aktivitas pengguna | Tabel `sys_alert_logs` (aktor, aksi, entitas, hasil, IP) ditulis lewat paket `activitylog` pada auth-service dan layanan terkait. | `auth-service/models/sys_alert_log.go:36-60`; `auth-service/services/auth/auth_service.go` | Tidak ditemukan penulisan peristiwa unduh/ekspor dari audit-service (pencarian `activitylog` pada audit-service tidak menemukan referensi). |
| Respons API tidak membocorkan hash/secret | `PasswordHash` dan `SecretKey` MFA memakai `json:"-"`. | `auth-service/models/user.go:16`, `:69` | Baik. Secret MFA tersimpan **tanpa enkripsi** di DB produksi sampai `MFA_ENCRYPTION_KEY` dipasang (`auth-service/cmd/auth/serve.go:62`; memori "prod encryption gaps"); itu risiko pada backup/dump, bukan respons API. |
| Kata sandi sumber data pada respons API | `Password` memakai `json:"password,omitempty"`; daftar, buat, dan ubah sumber data mengembalikan model utuh. | `master-service/models/data_source.go:71`; `master-service/routes/routes.go:666-684`; `master-service/routes/data_source_handlers.go:60-70`, `:127-133` | **Kebocoran terverifikasi dari kode** (bagian 4 butir 2). |
| CORS | Kong prod hanya mengizinkan origin `https://auditsphere.app`. Layanan Go (audit-service) memantulkan header `Origin` dengan `Allow-Credentials: true`. | `kong-gateway/kong/prod/kong.yml:31-43`; `audit-service/pkg/middleware/cors.go:8-20` | CORS layanan lemah bila layanan dapat dijangkau langsung; compose prod mengikat port ke loopback (`docker-compose.prod.yml:94`; A.8.21 butir tabel eksposur). Catatan 5 Nov 2026: IPv6 melewati firewall Docker; perbaikan siap, status VPS **Perlu verifikasi**. |
| Rate limiting | Kong: global 100/menit dan 5000/jam per IP; rute fieldwork 60/menit; analytics 120/menit; retrain 10/menit. | `kong-gateway/kong/prod/kong.yml:80-85`, `:464-466`, `:620-622`, `:656-658` | Per IP dan per permintaan, bukan per pengguna atau volume unduhan; tidak mendeteksi pengunduhan massal pelan-pelan. |

### 2.9 Backup, training, endpoint

| Kanal | Yang terlihat | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Backup off-site ke Backblaze B2 | Skrip terenkripsi `age` ada di repo; catatan 10 Nov 2026 menyebut skrip terpasang lama dan menulis berkas tanpa enkripsi, `age` belum terpasang. B2 tidak punya region Indonesia. | `backend/scripts/backup/backup-prod.sh`; A.8.13 bagian 4 butir 1 dan 4; memori "prod encryption gaps" | Kanal kebocoran terbesar bila backup polos terkirim keluar negeri. Status VPS dan bucket: **Perlu verifikasi**. |
| Training berisi klon data produksi | `reset-training.sh` menyalin DB dan uploads produksi (tanpa enrolmen MFA); dump sementara tidak terenkripsi. Klien nyata dapat login ke training. | `backend/training/reset-training.sh:78-93`, `:100-161`; A.8.13 bagian 4 butir 10 | Data klien nyata di lingkungan pelatihan; tidak ada masking. Kebijakan penggunaan data produksi (A.8.33, A.5.34) belum ada. |
| Endpoint, USB, clipboard, screenshot | Tidak ada MDM/EDR, kendali USB, atau kendali clipboard/screenshot; aplikasi tidak membatasi salin/cetak. | `docs/ISO_27001/A.8.1_Perangkat_Endpoint_Pengguna.md` (status "Belum diimplementasikan"); A.7.7 | **Tidak ada**; sifatnya organisasi dan perangkat, bukan aplikasi. |

### 2.10 Pemantauan dan DLP

| Kendali | Temuan | Bukti |
| :--- | :--- | :--- |
| Alat DLP (jaringan, email, endpoint, cloud) | **Tidak ada.** | pencarian repo |
| Rule set DLP, pola data sensitif, aturan deteksi | **Tidak ada.** | pencarian repo |
| Alert kebocoran | **Tidak ada.** Prometheus tidak memuat `rule_files` maupun Alertmanager. Dua berkas aturan (`prometheus-dashboard-alerts.yml`, `prometheus-fieldwork-alerts.yml`) berfokus pada kinerja dan dipasang manual (`docker cp`); bukan aturan kebocoran dan tidak termuat oleh konfigurasi. | `monitoring/prometheus/prometheus.yml`; `monitoring/prometheus-dashboard-alerts.yml:1-14` |
| Retensi log | Loki menyimpan 7 hari. | `monitoring/loki/loki-config.yml:36` |
| Tiket/insiden kebocoran | Tidak ditemukan kategori atau templat; A.5.24 (persiapan manajemen insiden) adalah tempat kebijakan insiden. | `docs/ISO_27001/A.5.24_Persiapan_Manajemen_Insiden.md` |

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| DLP policy | **Belum tersedia; draf di Lampiran A** | Kode usulan DLP-001. Perlu persetujuan; angka [usulan]. |
| Rule set | **Belum tersedia; matriks kanal x klasifikasi di Lampiran B sebagai dasar** | Tidak ada alat atau aturan deteksi. |
| Alert log | **Tidak ada** | Tidak ada alert kebocoran; Prometheus tanpa `rule_files` dan Alertmanager. |
| Incident ticket | **Tidak ada** | Tidak ada tiket kebocoran data; daftar periksa tinjauan di Lampiran C. |
| Daftar kanal kebocoran | **Sebagian** | Bagian 2 dokumen ini; belum disetujui sebagai daftar resmi. |
| Kendali teknis pada kode | **Tersedia (repo)** | Bagian 2; lihat Evidence EV-A8.12-01 s/d 06. Status ter-deploy: **Perlu verifikasi**. |
| Uji kanal (uji penetrasi/uji kebocoran) | **Belum tersedia** | Mis. uji `/uploads` tanpa token, uji unduh oleh VIEWER, uji respons data source. |
| Bukti pencatatan unduh/ekspor | **Tidak ada** | Tidak ditemukan peristiwa unduh/ekspor pada log aktivitas audit-service. |
| Bukti penanda klasifikasi | **Tidak ada** | A.5.12 bagian 4. |
| Kontrak/DPA pemasok (Resend, Backblaze) | **Belum tersedia** | A.5.20. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari catatan operasional internal dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan.

1. **[Tidak ada] Tidak ada program DLP.** Tidak ada kebijakan, alat, rule set, alert, atau tiket kebocoran. Kendali yang ada bersifat pencegahan tersebar dan tidak diturunkan dari klasifikasi. Nilai efektivitas operasional adalah "Belum teruji".
2. **[Repo] Kata sandi sumber data tersimpan polos dan dikembalikan oleh API.** `DataSourceConnection.Password` bertag `json:"password,omitempty"` (`master-service/models/data_source.go:71`); daftar (`routes.go:666-684`), buat (`data_source_handlers.go:60-70`), dan ubah (`:127-133`) mengembalikan model utuh, sehingga kata sandi basis data klien terkirim ke browser setiap pemanggil non-VIEWER. Kolom disimpan sebagai `varchar(255)` biasa dan tidak ditemukan kode enkripsi pada master-service; Data Hub menamai kolom `password_encrypted` tetapi memakainya sebagai kata sandi mentah (`core/db.py:145-146`). Data Hub sendiri menyaring kolom itu dari responsnya (`sources.py:118-119`), jadi master-service tertinggal. Ini kebocoran kredensial klien (*Restricted*) yang dapat diperbaiki cepat (tindakan 2). Apakah kata sandi tersimpan juga di tabel Data Hub dan backup: **Perlu verifikasi** (A.8.13).
3. **[Tidak ada] Unduhan dan ekspor tidak tercatat per pengguna.** Log akses audit-service memuat jalur, IP, dan user agent tanpa identitas pengguna (`logger.go:26-33`); `sys_alert_logs` tidak menerima peristiwa unduh/ekspor dari audit-service. Bila kebocoran terjadi, tidak ada jejak siapa mengunduh apa dan kapan. Loki hanya menyimpan 7 hari.
4. **[Repo] Tidak ada pembatasan per klasifikasi atau per berkas.** Setiap non-VIEWER terautentikasi pada satu tenant dapat mengunduh berkas mana pun yang jalurnya diketahui (`uploads.go:18-35`) atau berkas dokumen mana pun lewat `/:id/download`. Tidak ada atribut klasifikasi, tidak ada penugasan per audit pada unduhan. Penilaian kecocokan dengan kebutuhan hak akses audit (A.5.15, A.5.18, draf): **Perlu verifikasi**.
5. **[Repo] Cakupan VIEWER tidak merata.** audit-service dan master-service menolak VIEWER pada berkas/ekspor; risk-service dan analytics-service hanya menolak penulisan (`risk-service/routes/routes.go:63`, `analytics-service/routes/routes.go:22`). Pencocokan nama rute (`viewer.go:93-103`) tidak mencakup rute ekspor yang diberi nama lain. Perlu uji regresi per rute baru.
6. **[Repo] AI: endpoint OpenAI pada HEAD, penghapusan belum ter-commit.** `frontend/server/api/openai.ts` ada pada HEAD dan dihapus pada working tree; selama belum ter-commit dan ter-deploy, risikonya tetap ada bila `OPENAI_API_KEY` terpasang di server (**Perlu verifikasi**). Paket `openai` masih di `package.json:36`. Tidak ada kebijakan penggunaan AI untuk data audit (juga disebut A.6.7). Backend python-ai tidak menunjukkan panggilan ke API AI eksternal.
7. **[Catatan] Paparan jaringan dan `/uploads`.** Memori internal mencatat `/uploads` dan `media/download` publik (8 Okt 2026) dan port terbuka lewat IPv6 (5 Nov 2026). Repo memuat perbaikan (uploads wajib token; port loopback; skrip firewall IPv6 disiapkan). Apakah semuanya ter-deploy: **Perlu verifikasi** (uji tanpa token dan uji dari luar, oleh pihak berwenang). Jangan klaim "tidak ada paparan publik" sebelum terbukti.
8. **[Repo]+[Catatan] Email lewat pihak ketiga tanpa kontrol isi.** Pengingat ATR dan OTP terkirim lewat Resend (AS). Payload audit-service teks polos tanpa lampiran, tetapi tidak ada filter isi, allowlist domain, atau pencatatan peristiwa kirim; training mengirim OTP ke pengguna nyata. Isi pesan: **Perlu verifikasi**.
9. **[Catatan] Backup keluar negeri dan tidak terenkripsi menurut catatan.** Lihat A.8.13 bagian 4 butir 1, 4, dan 5. Selama tidak terbukti terenkripsi, backup adalah kanal kebocoran terbesar (satu berkas memuat semua tenant dan secret MFA polos). Jangan menulis "data tidak meninggalkan Indonesia" atau "seluruh backup terenkripsi".
10. **[Repo]+[Catatan] Training memuat data klien nyata tanpa masking.** `reset-training.sh` menyalin DB dan uploads produksi; pengguna nyata dapat login ke training; ClamAV dan alias jaringan dibagi dengan produksi. Tidak ada masking atau anonimisasi (tidak ditemukan kode).
11. **[Repo] CORS layanan memantulkan origin.** `cors.go:10-18` memantulkan `Origin` dengan kredensial; pengaman bergantung pada Kong (`kong.yml:31-34`) dan pengikatan loopback. Bila layanan dapat dijangkau langsung (mis. IPv6, catatan 5 Nov 2026), CORS layanan tidak melindungi. Perbaikan: samakan dengan allowlist.
12. **[Repo] Rate limiting tidak berbasis pengguna atau volume.** Per IP dan per rute (`kong.yml:80-85`); tidak ada ambang unduhan massal per pengguna dan tidak ada alert.
13. **[Tidak ada] Kendali endpoint, USB, clipboard, screenshot, cetak, dan watermark.** Bergantung pada A.8.1 (belum diimplementasikan) dan A.6.7/A.7.7; aplikasi tidak menawarkan watermark dokumen atau pembatasan cetak.
14. **[Repo] Konfigurasi log Loki dan alert.** Prometheus tidak memuat `rule_files`; berkas aturan yang ada dipasang manual dan tidak untuk kebocoran. Tidak ada saluran notifikasi (Alertmanager). Sama dengan temuan A.8.21 butir 4.
15. **[Repo] Kunci/kredensial pemasok tersebar.** Kunci Resend tertanam pada image menurut catatan training (memori); kunci per tenant dipisahkan (A.5.20). Perlu daftar kredensial pemasok dan rotasi (A.8.24, draf).
16. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet dalam penyusunan dokumen ini; tidak ada akses SSH.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Setujui kebijakan DLP-001** (Lampiran A): daftar kanal, matriks kanal x klasifikasi (Lampiran B), peran, pengecualian. Pastikan A.5.12 menetapkan satu daftar tingkat klasifikasi lebih dahulu. | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [usulan] 31 Desember 2026 | Kebijakan disetujui (EV-A8.12-04) |
| 2 | **Perbaiki kebocoran kata sandi sumber data**: jangan kembalikan `password` pada respons master-service (daftar/buat/ubah), enkripsi kolom saat disimpan (kunci terpisah), dan selaraskan penamaan `password_encrypted` pada Data Hub. Perubahan kode oleh agen backend atas persetujuan; deploy hanya atas konfirmasi eksplisit (daftar komit dan penulis, tanpa seeder). | Backend Lead (`auditsphere-backend`), `auditsphere-data-analyst` | Tinggi | [usulan] 15 Januari 2027 | PR/komit, uji respons tanpa `password` |
| 3 | **Pencatatan unduhan dan ekspor**: catat peristiwa (pengguna, jenis, entitas, hasil, waktu) untuk semua rute unduh/ekspor ke log aktivitas; sertakan identitas pengguna pada log akses. | Backend Lead | Tinggi | [usulan] 31 Januari 2027 | Contoh peristiwa; uji |
| 4 | **Verifikasi read-only status ter-deploy** (oleh pihak berwenang): `/uploads` tanpa token harus 401; `media/download` tanpa token; port dan IPv6; `OPENAI_API_KEY` tidak terpasang; versi ter-deploy memuat komit `b960dd0`. Jangan menampilkan rahasia. | DevOps Lead | Tinggi | [usulan] 15 Desember 2026 | Hasil uji bertanggal (tersanitasi) |
| 5 | **Commit penghapusan `frontend/server/api/openai.ts`** (dan hapus dependensi `openai` bila tidak dipakai), lalu tetapkan kebijakan penggunaan AI untuk data audit (tidak mengirim data klien ke LLM eksternal tanpa persetujuan). | Frontend Lead, Lead Information Security | Tinggi | [usulan] 15 Desember 2026 | Komit; kebijakan AI |
| 6 | **Seragamkan kendali VIEWER**: tentukan kebijakan VIEWER untuk risk/analytics (baca dan ekspor), tambahkan uji regresi yang memastikan setiap rute `download`/`export` baru tercakup. | Backend Lead | Sedang | [usulan] 31 Januari 2027 | Uji regresi |
| 7 | **Atribut klasifikasi dan label**: terapkan tingkat klasifikasi pada dokumen/laporan (turunan, bukan hardcode) dan batasi unduhan/ekspor sesuai Lampiran B (terkait A.5.12). | Product Owner, Backend Lead | Sedang | [usulan] 31 Maret 2027 | Spesifikasi dan implementasi |
| 8 | **Alert dan pemantauan**: muat `rule_files` dan Alertmanager (atau saluran setara) dengan aturan awal Lampiran C: lonjakan unduhan per pengguna, banyak 401/403 pada rute berkas, ekspor di luar jam kerja, email massal. | DevOps Lead (`auditsphere-platform`) | Tinggi | [usulan] 31 Maret 2027 | Aturan, alert uji (EV-A8.12-08) |
| 9 | **Selaraskan CORS layanan** dengan allowlist Kong dan batasi Admin API/port (A.8.21). | DevOps Lead, Backend Lead | Sedang | [usulan] 31 Januari 2027 | Konfigurasi; uji CORS |
| 10 | **Keputusan data produksi di training**: masking/anonimisasi atau batasi pengguna training; pisahkan kunci email training. | Lead Information Security, DevOps Lead | Sedang | [usulan] 31 Januari 2027 | Keputusan tertulis |
| 11 | **Kendali email**: tinjau isi pesan Resend, tambahkan allowlist domain/pencatatan kirim, dan DPA Resend (A.5.20). | Backend Lead, Legal & Compliance | Sedang | [usulan] 28 Februari 2027 | Tinjauan isi; DPA |
| 12 | **Backup dan lokasi**: laksanakan tindakan A.8.13 (enkripsi terpasang, keputusan lokasi B2). | DevOps Lead, Legal & Compliance | Tinggi | Mengikuti A.8.13 | Lihat A.8.13 |
| 13 | **Endpoint dan perilaku pengguna**: lanjutkan A.8.1 (baseline/MDM ringan), A.6.7 (kerja jarak jauh), A.7.7 (layar bersih); pertimbangkan watermark dokumen ekspor [keputusan]; daftarkan Evidence ID dan komit dokumen (komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Lead Information Security, Project Lead | Sedang | [usulan] 31 Maret 2027 | Keputusan; entri Evidence Index |

Catatan pelaksanaan: tindakan 2, 3, 5, 6, 9 menyentuh kode atau konfigurasi dan dilakukan agen backend/frontend/devops atas persetujuan, bukan oleh dokumen ini. Tindakan 4, 8, 9, 10, 12 menyentuh VPS atau layanan produksi; semuanya memerlukan konfirmasi eksplisit dan, bila berupa deploy, daftar komit dan penulis sejak deploy terakhir (hindari `push` ke `main` yang memicu deploy lama), satu perubahan per sesi, tanpa seeder.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.12-01 | Akses `/uploads` wajib autentikasi, jalur dikonfirmasi, direktori privat | `backend/audit-service/routes/uploads.go`, `audit-service/pkg/safefile/safefile.go` | Internal |
| EV-A8.12-02 | Kendali VIEWER (tulis, berkas, ekspor) | `backend/audit-service/pkg/middleware/viewer.go`, `backend/master-service/pkg/middleware/viewer.go`, `backend/audit-service/routes/handler.go:76` | Internal |
| EV-A8.12-03 | Pemindaian ClamAV (mode `required`) | `backend/audit-service/pkg/avscan/avscan.go`, `backend/docker-compose.prod.yml:181-182` | Internal |
| EV-A8.12-04 | Kebijakan DLP-001 dan matriks kanal (setelah tindakan 1) | `docs/ISO_27001/A.8.12_Pencegahan_Kebocoran_Data.md` (Lampiran A, B); [belum ada versi disetujui] | Internal |
| EV-A8.12-05 | Redaksi log dan serialisasi aman (`json:"-"`) | `backend/audit-service/pkg/middleware/redact.go`, `backend/auth-service/models/user.go` | Internal |
| EV-A8.12-06 | CORS dan rate limiting Kong | `backend/kong-gateway/kong/prod/kong.yml:31-43`, `:80-85` | Internal |
| EV-A8.12-07 | Hasil verifikasi read-only status ter-deploy (setelah tindakan 4) | [belum ada keluaran VPS] | Restricted |
| EV-A8.12-08 | Aturan alert dan catatan alert uji (setelah tindakan 8) | [belum ada] | Internal / Restricted |
| EV-A8.12-09 | Perbaikan respons sumber data dan uji (setelah tindakan 2) | [belum ada] | Internal |
| EV-A8.12-10 | Peristiwa unduh/ekspor pada log aktivitas (setelah tindakan 3) | [belum ada] | Restricted |
| EV-A8.12-11 | Catatan tinjauan alert/insiden berkala (Lampiran C) | [belum ada] | Restricted |
| EV-A8.12-12 | DPA Resend dan keputusan lokasi backup (setelah tindakan 11, 12) | [belum ada] | Restricted |

---

## Lampiran A — Draf Kebijakan Pencegahan Kebocoran Data (DLP-001, usulan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Nilai pada kurung siku adalah **[usulan]**.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | DLP-001 Kebijakan Pencegahan Kebocoran Data (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah insiden kebocoran atau penambahan kanal baru) |
| Berlaku untuk | Seluruh data yang dikelola AuditSphere (produksi, tenant, training, Data Hub milik AuditSphere), personel, dan pemasok pemroses data |

1. **Tujuan.** Mencegah, mendeteksi, dan menanggapi pengungkapan data klien dan data internal kepada pihak yang tidak berwenang melalui kanal yang diketahui.
2. **Dasar klasifikasi.** Kendali mengikuti tingkat klasifikasi A.5.12 (usulan: Public, Internal, Confidential, Restricted). Data audit klien dan kredensial dianggap Confidential/Restricted sampai klasifikasi per item ditetapkan.
3. **Daftar kanal.** Unduh berkas dan dokumen; ekspor DOCX/XLSX/PDF; cetak dan salin dari browser; email dan notifikasi (Resend); sinkronisasi Data Hub dan sumber data eksternal; panggilan AI/LLM eksternal; backup dan salinan off-site; lingkungan training; log dan metrik; perangkat endpoint dan media removable; media sosial/percakapan eksternal (A.6.7, A.6.3).
4. **Prinsip.** Akses minimal; kanal keluar didaftarkan; kanal tidak terdaftar dilarang; setiap pengiriman data Restricted ke pihak ketiga memerlukan persetujuan dan DPA; data klien tidak dikirim ke layanan AI eksternal tanpa persetujuan tertulis.
5. **Kendali minimum.** Autentikasi pada semua berkas dan ekspor; peran VIEWER tanpa berkas/ekspor; pemindaian malware unggahan; tidak ada kredensial atau secret pada respons API, log, atau email; enkripsi transit dan penyimpanan untuk data Restricted (A.8.24, draf).
6. **Pencatatan dan pemantauan.** Peristiwa unduh/ekspor/pengiriman dicatat dengan pengguna, entitas, dan waktu; log disimpan [12 bulan, usulan]; tinjauan sesuai Lampiran C; alert dieskalasi ke Lead Information Security.
7. **Penanganan insiden.** Dugaan kebocoran diperlakukan sebagai insiden keamanan (A.5.24): kontainmen, pengumpulan bukti (A.5.28), penilaian dampak, notifikasi klien/regulator sesuai kontrak dan UU PDP [Legal & Compliance].
8. **Pengecualian.** Tertulis, dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir.
9. **Peran.** Sesuai RACI Lampiran D.
10. **Sanksi dan kesadaran.** Pelanggaran ditangani sesuai A.6.4 (draf); pelatihan dan pengingat rutin (A.6.3).

---

## Lampiran B — Matriks Kanal x Klasifikasi (usulan)

Semua kendali pada kolom "Target" adalah **[usulan]**. Kolom "Saat ini" adalah temuan repo (status VPS: **Perlu verifikasi**).

| Kanal | Saat ini | Public | Internal | Confidential | Restricted (target [usulan]) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Unduh dokumen/berkas bukti | Autentikasi + non-VIEWER; tanpa klasifikasi | Bebas | Terautentikasi | Terautentikasi + peran audit terkait + dicatat | Peran terkait + persetujuan atau penugasan + dicatat + alert volume |
| Ekspor DOCX/XLSX/PDF | Autentikasi + non-VIEWER; label tetap | Bebas | Terautentikasi | Dicatat; label klasifikasi otomatis | Dicatat; watermark [keputusan]; batas volume |
| Cetak/salin browser | Tidak dikendalikan | Bebas | Bebas | Kebijakan (A.7.7); tidak ada kendali teknis | Kebijakan; evaluasi watermark |
| Email (Resend) | Teks polos tanpa lampiran; tanpa filter | Bebas | Penerima internal | Tanpa data Confidential di isi email; tautan saja | Dilarang; gunakan tautan ke aplikasi |
| Data Hub / sumber data | SSL tervalidasi; kata sandi bocor di API | n/a | Terautentikasi | Enkripsi kredensial; dicatat | Enkripsi kredensial; penyetuju; lokasi sesuai kontrak |
| AI/LLM eksternal | Tidak ada panggilan backend; endpoint frontend dihapus (belum commit) | Diizinkan | Dengan persetujuan | Dilarang tanpa persetujuan | Dilarang |
| Backup off-site | Tidak terbukti terenkripsi; B2 luar Indonesia (catatan) | n/a | Terenkripsi | Terenkripsi; lokasi sesuai kontrak | Terenkripsi; kunci terpisah; lokasi disetujui Legal |
| Training | Klon produksi tanpa masking | n/a | Boleh | Masking | Dilarang tanpa anonimisasi |
| Log/metrik | Redaksi query; 7 hari di Loki | Bebas | Terautentikasi | Tanpa data pribadi di log | Tanpa data pribadi/secret di log |
| Endpoint/USB | Tidak ada kendali | Bebas | Perangkat terdaftar | Perangkat terenkripsi (A.8.1) | Perangkat terkelola; USB dibatasi |
| Pihak ketiga | DPA belum ada | Bebas | DPA | DPA + tinjauan | DPA + persetujuan dan tinjauan tahunan |

---

## Lampiran C — Daftar Periksa Tinjauan Alert dan Insiden Kebocoran (usulan)

Frekuensi: [mingguan oleh DevOps Lead; bulanan ke Lead Information Security]. Setiap tinjauan dicatat (tanggal, peninjau, hasil).

| No | Pemeriksaan | Hasil (ya/tidak/n.a.) | Catatan |
| :--- | :--- | :--- | :--- |
| 1 | Alert unduhan/ekspor yang tidak biasa (volume per pengguna, jam di luar kerja) ditinjau; tiap alert dihubungkan ke tiket atau alasan sah. | [isi] | |
| 2 | Lonjakan 401/403 pada `/uploads`, `/api/v1/media`, dan rute unduh ditinjau. | [isi] | |
| 3 | Uji `/uploads` tanpa token menghasilkan 401; uji VIEWER pada rute unduh/ekspor menghasilkan 403. | [isi] | |
| 4 | Respons daftar/buat/ubah sumber data tidak memuat kata sandi. | [isi] | |
| 5 | Log aplikasi tidak memuat token, kata sandi, atau data pribadi (sampel). | [isi] | |
| 6 | Pengiriman email massal atau penerima luar domain ditinjau. | [isi] | |
| 7 | Tidak ada panggilan keluar baru ke layanan AI/LLM atau domain asing (periksa dependensi dan konfigurasi). | [isi] | |
| 8 | Daftar pengguna aktif dan peran ditinjau (A.5.18, draf); akun tak dikenal dicabut. | [isi] | |
| 9 | Backup terbaru terenkripsi dan lokasi sesuai keputusan (A.8.13 Lampiran D). | [isi] | |
| 10 | Data klien di training sesuai keputusan (masking/pembatasan). | [isi] | |
| 11 | Insiden kebocoran dieskalasi sesuai A.5.24; tiket: [isi]. | [isi] | |
| 12 | Uji alert (kuartalan) berhasil; temuan sebelumnya ditindaklanjuti. | [isi] | |

---

## Lampiran D — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Backend Lead | Legal & Compliance | Product Owner | Direktur Utama |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan DLP-001 | A/R | C | C | C | C | A (persetujuan) |
| Klasifikasi dan label (A.5.12) | A | I | R | C | R | I |
| Perbaikan kebocoran API/kredensial | A | C | R | I | I | I |
| Pencatatan unduh/ekspor | A | C | R | I | C | I |
| Alert dan tinjauan berkala | C | A/R | C | I | I | I |
| Verifikasi status ter-deploy | C | A/R | C | I | I | I |
| Kebijakan AI dan email pihak ketiga | A | C | R | C | C | I |
| Keputusan lokasi dan DPA pemasok | C | C | I | A/R | I | A (persetujuan) |
| Penanganan insiden kebocoran (A.5.24) | A | R | R | C | I | I |
| Evidence ID dan review berkala | A/R | C | C | C | I | I |

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.5.12 Klasifikasi informasi | Ada | Skema klasifikasi dan label; dasar matriks kanal x klasifikasi. |
| A.8.1 Perangkat endpoint pengguna | Ada (belum ter-commit) | Kendali endpoint/USB dan MDM/EDR; belum diimplementasikan. |
| A.8.13 Pencadangan informasi | Ada (belum ter-commit) | Backup B2 off-shore, tidak terenkripsi menurut catatan, dump pra-deploy. |
| A.8.21 Keamanan layanan jaringan | Ada (belum ter-commit) | Eksposur port, CORS, Kong, isolasi tenant, tanpa alert Prometheus. |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada | Resend dan Backblaze; DPA belum ada. |
| A.5.24 Persiapan manajemen insiden | Ada | Penanganan insiden kebocoran. |
| A.6.7 Kerja jarak jauh | Ada | Perilaku pengguna, panduan agen AI, perangkat pribadi. |
| A.7.7 Meja bersih dan layar bersih | Ada (belum ter-commit) | Layar, cetak, dan tampilan data di ruang kerja. |
| A.5.28 Pengumpulan bukti | Ada | Pengumpulan bukti bila terjadi kebocoran. |
| A.5.15, A.5.18 Kendali dan hak akses | **Draf, dokumen belum ada** | Hak akses per audit/berkas. |
| A.5.34 Privasi dan PII | **Draf, dokumen belum ada** | Data pribadi pada ekspor, email, training. |
| A.6.3 Awareness dan pelatihan | Ada | Pelatihan perilaku penanganan data. |
| A.6.4 Proses disipliner | **Draf, dokumen belum ada** | Sanksi pelanggaran. |
| A.8.7 Perlindungan terhadap malware | **Draf, dokumen belum ada** | ClamAV pada unggahan. |
| A.8.15 Pencatatan log | **Draf, dokumen belum ada** | Pencatatan unduh/ekspor dan identitas pengguna. |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Alert kebocoran. |
| A.8.24 Penggunaan kriptografi | **Draf, dokumen belum ada** | Enkripsi kredensial sumber data, secret MFA, backup. |
| A.8.11 Penyamaran data | **Draf, dokumen belum ada** | Masking untuk training. |
| A.8.33 Perlindungan informasi uji | **Draf, dokumen belum ada** | Data produksi pada training. |

---

## Lampiran Akhir — Blok Nilai Usulan untuk Salin-Tempel ke Checklist (satu baris, 25 kolom)

Urutan kolom persis sesuai checklist: Kontrol, Tema, Judul kontrol, Fokus pemeriksaan, Contoh bukti objektif, Relevansi SoA, Justifikasi relevansi, Status implementasi, Ketersediaan bukti, Evidence ID / URL, PIC, Review terakhir, Review berikutnya, Desain kontrol, Efektivitas operasional, Catatan / gap, Tindakan perbaikan, Prioritas, Target, Reviewer, Tanggal validasi, Pemeriksaan field, Kesiapan, Deadline, Perlu tindakan. Pemisah antar-kolom adalah karakter tab asli. Semua nilai adalah **usulan** yang menunggu validasi pengguna; baris pertama adalah header, baris kedua adalah nilai.

```tsv
Kontrol	Tema	Judul kontrol	Fokus pemeriksaan	Contoh bukti objektif	Relevansi SoA	Justifikasi relevansi	Status implementasi	Ketersediaan bukti	Evidence ID / URL	PIC	Review terakhir	Review berikutnya	Desain kontrol	Efektivitas operasional	Catatan / gap	Tindakan perbaikan	Prioritas	Target	Reviewer	Tanggal validasi	Pemeriksaan field	Kesiapan	Deadline	Perlu tindakan
A.8.12	Technological	Pencegahan kebocoran data	Kanal kebocoran dipantau dan dikendalikan sesuai klasifikasi	DLP policy, rule set, alert log, incident ticket	Berlaku	Data audit klien lembaga jasa keuangan (Confidential/Restricted) dengan banyak kanal keluar (unduh, ekspor, email, Data Hub, backup off-site, training); multi-tenant; klien akan menanyakan kendali kebocoran (lihat A.8.12 bagian 1)	Sebagian (minimal)	Sebagian (kendali teknis di kode dan konfigurasi); Tidak ada (kebijakan DLP, rule set, alert log, tiket insiden)	Belum terdaftar (usulan EV-A8.12-01 s/d EV-A8.12-12)	Lead Information Security (didukung DevOps Lead, Backend Lead)	19 November 2026 (usulan; tinjauan repo)	[usulan] 19 Februari 2027, atau setelah kebijakan DLP disetujui dan kebocoran API diperbaiki	Perlu perbaikan	Belum teruji	Tanpa kebijakan/alat/rule set/alert DLP; kata sandi sumber data tersimpan polos dan dikembalikan oleh API master-service; unduhan/ekspor tidak tercatat per pengguna; tanpa penanda klasifikasi, watermark, masking; Prometheus tanpa rule_files/Alertmanager; backup off-shore dan tidak terenkripsi menurut catatan; training memuat klon data produksi; endpoint OpenAI frontend dihapus tetapi belum ter-commit; status VPS (uploads, port) Perlu verifikasi	Setujui DLP-001; perbaiki kebocoran kata sandi sumber data; pencatatan unduh/ekspor; verifikasi status ter-deploy; commit penghapusan endpoint OpenAI dan kebijakan AI; alert dan tinjauan berkala; kendali training, email, backup (13 tindakan di bagian 5)	Tinggi	[usulan] 31 Januari 2027 (kebijakan, perbaikan API, pencatatan unduhan); 31 Maret 2027 (alert dan tinjauan berkala)	[isi saat validasi]	19 November 2026 (verifikasi repo); validasi reviewer [isi saat validasi]	Perlu dilengkapi (Evidence ID, Reviewer, Tanggal validasi reviewer)	Belum siap	[usulan] 28 Februari 2027	Ya
```

Catatan penggunaan: pemisah kolom pada blok di atas adalah karakter tab asli, sehingga dapat ditempel langsung ke sel A pada baris kosong spreadsheet. Bila editor mengubah tab menjadi spasi, tempel ulang dari tampilan mentah berkas. Tidak ada perubahan pada berkas xlsx yang dilakukan oleh dokumen ini.
