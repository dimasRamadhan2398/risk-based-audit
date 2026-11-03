# A.5.28 — Pengumpulan Bukti

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Pengumpulan bukti (ISO/IEC 27001:2022 A.5.28: Pengumpulan bukti terkait kejadian keamanan informasi) |
| **Fokus pemeriksaan** | Bukti insiden dikumpulkan dan dijaga integritas serta chain of custody |
| **Contoh bukti objektif** | Prosedur forensik, evidence log, hash, chain-of-custody form |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Security Officer [konfirmasi] (pemilik FRP-001; tidak tercantum di baris checklist) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 30 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Dokumen ini bukan nasihat hukum. Pernyataan tentang keabsahan bukti di pengadilan perlu divalidasi Legal.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Insiden pada AuditSphere (kebocoran data audit klien, kompromi akun admin, perubahan data kertas kerja) bisa berujung pada proses hukum, klaim klien, atau pemeriksaan regulator. Bukti harus bisa dikumpulkan tanpa merusak integritasnya. |
| Implementasi status | **Sebagian** | Prosedur forensik tingkat dokumen (FRP-001) dan formulir Chain of Custody contoh sudah ada. Di sisi teknis ada checksum SHA-256 untuk backup, pemindaian ClamAV pada unggahan, dan soft delete pada tabel inti. Evidence log, prosedur operasional yang bisa dijalankan, tabel jejak audit yang terisi, dan hash per berkas unggahan belum ada. |
| Ketersediaan bukti | **Sebagian** | Ada FRP-001 (belum disahkan) dan skrip backup dengan `.sha256`. Tidak ada evidence log, formulir CoC terisi, catatan hash bukti insiden, atau log restore test. Lampiran A sampai D dokumen ini menyediakan templat yang bisa dipakai. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.28-01 s/d EV-A5.28-05) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada catatan pengumpulan bukti nyata atau latihan. ASC-003 NC-04 (LiME/Volatility belum terpasang di VPS produksi) berstatus Open. Restore test belum tercatat hasilnya. |
| Desain kontrol | **Perlu perbaikan** | FRP-001 mengasumsikan lab forensik air-gapped, write-blocker, dan repositori log terpusat yang tidak terbukti ada. Sumber bukti utama AuditSphere (log aplikasi, basis data per tenant, berkas unggahan, backup) belum punya prosedur akuisisi yang spesifik, dan log aplikasi tidak punya retensi terjamin. |
| Status kesiapan | **Belum siap** | Penerapan, bukti, dan pengujian belum memenuhi gate. |
| Prioritas | **Tinggi** | Bukti yang tidak diamankan segera akan hilang: log container ikut hilang saat container dibuat ulang, dan sebagian berkas unggahan hilang saat redeploy (lihat gap 4 dan 5). |
| Review berikutnya | [isi] | Usulan: 3 bulan setelah prosedur, evidence log, dan formulir CoC disahkan, lalu setiap tahun atau setelah insiden signifikan (FRP-001 menetapkan review berikutnya 28 September 2027). |

> Catatan pembacaan checklist: nomor kontrol A.5.28 dikonfirmasi dari `docs/DOKUMEN_KESELARASAN_ISO27001.md:324` (kelompok "A.5.24 s/d A.5.28", status Partially Implemented) dan dari kolom kontrol ISO di FRP-001 (A.5.28, A.5.26, A.8.15). Baris A.5.28 di workbook `03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` tidak dibaca ulang untuk dokumen ini. PIC tidak tercantum, sehingga dokumen ini memakai "Security Officer" (pemilik FRP-001) sebagai usulan dan perlu dikonfirmasi.

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.28 **berlaku** untuk AuditSphere karena:

1. Platform menyimpan data audit, kertas kerja (KKA), Laporan Hasil Audit (LHA), bukti Action Taken Report (ATR), dan data pribadi klien (lihat `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 13.3). Insiden seperti akses tanpa izin ke bukti audit atau perubahan data oleh pihak internal menuntut bukti yang bisa dipertanggungjawabkan.
2. Bagian 11.1 langkah 4 `docs/DOKUMEN_KESELARASAN_ISO27001.md:437` menyebut "meneliti log audit trail, log akses, dan dump memori". Langkah itu hanya berguna jika sumbernya ada dan terjaga. Bagian 2 dokumen ini menunjukkan sumber mana yang benar-benar ada.
3. FRP-001 (`docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx`) mensyaratkan hash SHA-256 pada setiap bukti, formulir CoC lengkap, dan target koleksi ≤ 4 jam sejak eskalasi, dengan rujukan UU ITE No. 19/2016 dan UU PDP No. 27/2022. Dokumen operasional untuk memenuhi syarat itu belum ada.
4. Produksi berjalan sebagai Docker Compose di satu VPS, ditambah satu stack per tenant (`scripts/generate-tenant-stack.sh`, `scripts/templates/docker-compose.tenant.yml.tpl`). Bukti insiden satu tenant harus dikumpulkan tanpa membuka data tenant lain.
5. Klien enterprise dan bank pada due diligence akan menanyakan siapa yang boleh mengambil bukti, bagaimana keutuhannya dibuktikan, dan siapa yang memegangnya.

---

## 2. Implementasi yang Sudah Ada

Setiap butir diverifikasi dari repositori. Path dan baris mengacu pada cabang `staging`. Tidak ada nilai rahasia (kunci API, token, IP server, kredensial) yang dicantumkan di dokumen ini. Status "terverifikasi" berarti diverifikasi dari kode atau konfigurasi di repositori, bukan dari kondisi VPS yang berjalan.

### 2.1 Prosedur dan formulir tingkat organisasi

| Dokumen / bagian | Isi | Lokasi | Status verifikasi |
| :--- | :--- | :--- | :--- |
| Forensic Readiness Procedure (FRP-001, DOK-120, v1.0, berlaku 28 September 2026) | Empat komponen kesiapan, urutan volatilitas (RAM, jaringan, proses, disk, backup/log), atribut formulir Chain of Custody, aspek hukum (UU ITE, UU PDP), dan metrik (integritas hash 100%, CoC 100%, koleksi ≤ 4 jam) | `docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx` | **Ada (dokumen).** Lembar pengesahan berisi placeholder "[Ditandatangani Digital]". Contoh isian memakai host `SRV-PROD-DB01` dan "Evidence Safe Locker #02 (Lab Forensik Lt. 3)", yang tidak sesuai dengan VPS (lihat gap 2). Rujukan IRP-001 tidak ditemukan. |
| Advanced Security Operations Checklist (ASC-003) | NC-04: LiME/Volatility belum terpasang di VPS produksi, berstatus Open | `docs/ISO_27001/advanced_security/Advanced_Security_Operations_Checklist.docx` (bagian 10) | **Ada (checklist).** Dikutip dari verifikasi A.5.24; tidak dibuka ulang di sini. |
| Prosedur insiden tingkat kebijakan | Langkah investigasi dan analisis forensik (11.1 langkah 4) dan klasifikasi severity | `docs/DOKUMEN_KESELARASAN_ISO27001.md:429-462` | **Terverifikasi (teks kebijakan).** Tidak memuat aturan pengumpulan, penyimpanan, atau serah terima bukti. |
| Klaim audit trail | "Perekaman audit trail bawaan pada setiap perubahan data, sentralisasi log, proteksi log dari modifikasi" berstatus Implemented (A.8.15 & A.8.16); "immutable log" pemilihan sampel | `docs/DOKUMEN_KESELARASAN_ISO27001.md:267, 321, 493` | **Klaim dokumen, tidak didukung kode** pada bagian yang dicek (lihat 2.2 dan gap 3). |
| Kontrol terkait | A.5.24 (kesiapan insiden), A.5.26 (respons), A.8.15 (logging), A.8.13 (backup) | `docs/ISO_27001/A.5.24_Persiapan_Manajemen_Insiden.md` | Dokumen annex tetangga; ketergantungan dicatat di bagian 5. |

### 2.2 Sumber bukti di dalam AuditSphere dan keadaan aktualnya

| Sumber bukti | Yang tersedia | Lokasi | Status verifikasi |
| :--- | :--- | :--- | :--- |
| Log aplikasi | `zap` ke stdout di setiap service. Tidak ada konfigurasi `logging:` (log-driver, `max-size`) di compose. Log disimpan oleh Docker pada container, jadi hilang bersama container saat dibuat ulang. | contoh `backend/auth-service/pkg/logger/logger.go`; `backend/docker-compose.prod.yml` | **Sebagian.** Retensi, rotasi, dan pengiriman ke lokasi terpusat tidak ada. Dasar untuk gap 4. |
| Tabel jejak login/keamanan `sys_alert_logs` | Model dan repositori ada (aksi login, logout, gagal login, ubah izin, locked out, role assigned/revoked, rate limit, reset password; kolom user, IP, waktu). Dimigrasi lewat AutoMigrate. | `backend/auth-service/models/sys_alert_log.go`; `backend/auth-service/repositories/sys_alert_log_repository.go`; `backend/auth-service/cmd/auth/migrate.go:160` | **Skema ada, tidak terisi.** Pencarian nama konstanta aksi dan `GetSystemAlertLogRepository()` tidak menemukan pemanggil di luar model dan registry. Jangan dipakai sebagai bukti sebelum ada penulisnya. |
| Event keamanan Kafka | Publisher untuk login, logout, login gagal, MFA, perubahan password, role | `backend/auth-service/pkg/kafka/events.go`, `backend/audit-service/pkg/kafka/events.go:154-285`, `backend/master-service/pkg/kafka/events.go` | **Tidak terpakai sebagai jejak audit.** Di audit-service, metode `Publish*` tidak punya pemanggil di luar `events.go`. Konsumer audit-service dipakai untuk sinkronisasi rencana tahunan (`backend/audit-service/cmd/serve.go:332`), bukan untuk menyimpan jejak audit. Retensi topik Kafka 168 jam (`backend/docker-compose.prod.yml:66`), bukan arsip. |
| Tabel `user_access_audit_logs` | Hanya nama sumber data di kode | `backend/master-service/routes/routes.go:959`, `:1041` | **Jangan dipakai sebagai bukti.** Tidak ada model atau migrasi (sama dengan temuan A.5.24). |
| Jejak perubahan di tabel domain | Soft delete (`gorm.DeletedAt`) pada kertas kerja dan beberapa model audit; `ReviewedByID` dan `Status` pada kertas kerja; `UploadedBy`, `UploadedByUserID`, `UploadedAt`, `FileSize` pada bukti ATR | `backend/audit-service/models/working-paper.go:35,94,165,221,238,251,259,262`; `backend/audit-service/models/action-taken-report.go:68-81` | **Sebagian.** Berguna sebagai petunjuk siapa dan kapan, tetapi bukan jejak perubahan: nilai lama tidak disimpan, `UpdatedAt` menimpa, dan baris yang di-soft-delete masih bisa dihapus permanen. Tidak ada kolom hash atau checksum. |
| Berkas unggahan (bukti audit, bukti ATR, dokumen) | Allowlist ekstensi, pemeriksaan magic bytes, batas 10 MB, nama berkas dibuat server dengan sufiks acak, tulis dengan `O_EXCL`, pemindaian ClamAV mode `required` (fail-closed), unduh hanya lewat endpoint terautentikasi | `backend/audit-service/pkg/upload/upload.go` (komentar paket, `StoredName`, `CheckContent`); `backend/audit-service/pkg/avscan/avscan.go`; `backend/docker-compose.prod.yml:145-150,173-176`; `docs/security-hardening-changes-backend.md:54` | **Terverifikasi (kode).** Mengamankan masuknya berkas, tetapi **tidak menghitung atau menyimpan hash** berkas. Tidak ada cara membuktikan bahwa berkas tidak berubah sejak diunggah. |
| Lokasi penyimpanan unggahan (audit-service, prod) | Volume bernama `audit_uploads` di `/root/uploads`. Deploy menyalin `/root/uploads` keluar dan mengembalikannya setelah recreate. | `backend/docker-compose.prod.yml:179,299`; `backend/deploy-prod.sh:12,250-271`; `docs/deploy-tidy-plan.md:32` | **Ada di repositori.** Apakah VPS memakai volume ini belum diverifikasi (`docs/backend-enhancement-plan.md:59`). Komentar di `backend/scripts/backup/backup-prod.sh` ("not in a volume") sudah usang terhadap compose. |
| Lokasi penyimpanan unggahan (master-service, prod) | Tanpa volume. Jalur `./uploads/qar-reports` dan `./uploads/consulting` ada di writable layer container. | `docs/deploy-tidy-plan.md:33` | **Gap.** Hilang saat container dibuat ulang; skrip deploy tidak menyalinnya. Perbaikan baru berupa usulan (`docs/deploy-tidy-plan.md:85`). |
| Penyimpanan unggahan per tenant | Direktori silo `__STORAGE_DIR__` per tenant dipasang ke `/root/uploads`; Google Drive dimatikan (`GDRIVE_ENABLED: "false"`) agar bukti tidak bercampur antar tenant | `scripts/templates/docker-compose.tenant.yml.tpl:112-128` | **Terverifikasi (templat).** Memudahkan akuisisi per tenant. Backup untuk stack tenant tidak ditemukan (`backup-prod.sh` hanya menargetkan `rb_audit_postgres` dan `rb_audit_audit_service`). |
| Backup basis data dan unggahan | `pg_dumpall` terkompresi, arsip uploads, dan `sha256sum` untuk keduanya (`*.sha256`); retensi lokal 14 hari; salinan off-site opsional lewat rclone dengan `--checksum` | `backend/scripts/backup/backup-prod.sh` | **Skrip ada, belum terjadwal.** Cron belum dipasang dan `RCLONE_REMOTE` belum diset (`docs/backup-b2-setup.md:6`: "DRAFT. Not installed on the VPS."). File `.sha256` disimpan di direktori yang sama dengan datanya, sehingga tidak melindungi dari pelaku yang bisa menulis direktori itu. |
| Uji restore dan verifikasi checksum | Memeriksa `.sha256`, merestore ke container Postgres sementara, memeriksa jumlah baris dan berkas | `backend/scripts/backup/restore-test.sh` | **Skrip ada, belum ada log hasil.** Keluaran skrip ke terminal saja; belum ada penyimpanan hasil. |
| Hash lain di kode | `GenerateFingerprint` (SHA-256 atas subnet dan user agent) untuk perangkat tepercaya; HMAC-SHA256 untuk kunci step-up | `backend/audit-service/pkg/utils/signature.go`; `backend/master-service/pkg/middleware/step_up.go:40` | **Bukan hash integritas bukti.** Dicatat agar tidak disalahartikan sebagai kontrol A.5.28. |
| Isolasi per tenant | Basis data terpisah per tenant (`rb_audit_<svc>_<slug>`), stack terpisah | `scripts/generate-tenant-stack.sh`, `scripts/onboard-tenant.sh` | **Terverifikasi (skrip).** Dasar untuk membatasi cakupan akuisisi. |
| Jejak penerapan (deploy) | Deploy produksi manual dengan ref eksplisit dan environment `production`; skrip deploy mencatat log dan menyalin uploads | `.github/workflows/deploy.yml`, `backend/deploy-prod.sh` | **Sebagian.** Berguna untuk linimasa perubahan, tetapi tidak disimpan sebagai bukti. |
| Kontrol akses ke bukti | RBAC berbasis role di frontend; semua rute `/api/v1` dan `/uploads/*` memerlukan JWT valid. Role tidak diperiksa di sisi server pada audit, master, dan risk. | `docs/security-hardening-changes-backend.md:119`; `docs/backend-enhancement-plan.md` | **Sebagian.** Siapa pun yang punya JWT valid bisa mengunduh berkas bukti jika tahu jalurnya. Berpengaruh pada klaim bahwa bukti tidak dimodifikasi atau diakses pihak lain. |

### 2.3 Hal yang belum ditemukan di repositori

- Evidence log atau register bukti insiden (tidak ada berkas atau templat selain contoh isian di FRP-001).
- Formulir CoC yang sudah terisi, dan templat CoC di luar tabel contoh FRP-001.
- Prosedur akuisisi yang spesifik untuk AuditSphere (container Docker, volume, Postgres per tenant, Kafka, Kong).
- Catatan hash bukti insiden, dan hash per berkas unggahan di basis data.
- Log restore test dan log eksekusi backup.
- Penyimpanan log terpusat dengan retensi dan perlindungan dari perubahan (WORM atau sejenisnya).
- Sinkronisasi waktu (NTP/chrony) yang terdokumentasi untuk VPS; FRP-001 mensyaratkan waktu koleksi "NTP Synchronized".
- Tempat penyimpanan bukti yang terpisah dari produksi dengan akses terbatas.
- Kebijakan retensi bukti insiden dan legal hold.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Prosedur forensik | **Tersedia (dokumen, belum disahkan)** | FRP-001. Perlu disesuaikan dengan lingkungan VPS dan Docker. Lampiran A dokumen ini adalah usulan prosedur operasional yang menurunkannya. |
| Evidence log | **Belum tersedia** | Tidak ada register. Lampiran B menyediakan templat. |
| Hash (SHA-256) | **Sebagian** | Ada `.sha256` untuk backup (`backup-prod.sh`). Tidak ada hash untuk bukti insiden maupun berkas unggahan. Lampiran C menyediakan templat dan perintahnya. |
| Chain-of-custody form | **Sebagian** | Atribut formulir ada sebagai contoh di FRP-001 bagian 4 (tujuh atribut, tanpa kolom serah terima berurutan, saksi, atau tanda tangan). Lampiran D menyediakan formulir lengkap. |
| Catatan pengumpulan bukti nyata | **Belum tersedia / tidak ada di repositori** | Status ini tidak berarti tidak pernah ada insiden. Konfirmasi ke Security Officer [konfirmasi]. |
| Log restore test | **Belum tersedia** | Skrip ada, belum ada log (lihat A.5.24 bagian 3). |
| Bukti latihan pengumpulan bukti | **Belum tersedia** | Tidak ada catatan. Usulan tindakan 9. |
| Dasar hukum pembuktian (UU ITE, UU PDP) | **Sebagian** | Dirujuk di FRP-001 bagian 5. Teks undang-undang belum diverifikasi di repositori dan perlu diperiksa Legal [konfirmasi]. |

---

## 4. Catatan / Gap Terverifikasi

1. **Evidence log dan prosedur operasional belum ada.** FRP-001 hanya memberi kerangka. Tidak ada register bukti, tidak ada langkah akuisisi untuk komponen AuditSphere, dan rujukan IRP-001 tidak ditemukan. Status FRP-001 belum disahkan (placeholder tanda tangan).
2. **FRP-001 tidak sesuai dengan lingkungan produksi.** Dokumen mengasumsikan server `SRV-PROD-DB01`, lab forensik air-gapped, safe locker, write-blocker, FTK Imager, EnCase, dan "repositori log terpusat". Produksi adalah VPS Docker Compose tanpa pengendali itu di repositori, dan ASC-003 NC-04 mencatat LiME/Volatility belum terpasang. Target "kesiapan forensik 100%" dan "koleksi ≤ 4 jam" tidak punya dasar ukur.
3. **Jejak audit tidak terbukti ada.** `sys_alert_logs` tidak terisi (tidak ada penulis), event Kafka tidak diubah menjadi jejak audit, dan `user_access_audit_logs` hanya nama. Klaim "Implemented" untuk A.8.15 dan "immutable log" di `docs/DOKUMEN_KESELARASAN_ISO27001.md:267,321,493` tidak didukung kode. Tabel domain hanya menyimpan siapa dan kapan terakhir, bukan riwayat perubahan, dan soft delete bisa dilewati dengan hapus permanen.
4. **Log aplikasi mudah hilang.** Log hanya di stdout container tanpa rotasi atau retensi yang ditetapkan, tanpa pengiriman terpusat, dan tanpa perlindungan perubahan. Deploy yang membuat ulang container (recreate) menghasilkan container baru dan log lama ikut hilang. Insiden yang ditemukan setelah deploy bisa kehilangan buktinya.
5. **Berkas unggahan tidak seluruhnya terjaga.** Audit-service punya volume `audit_uploads` di compose prod, tetapi pemakaiannya di VPS belum diverifikasi. Master-service (unggahan QAR dan consulting) tidak punya volume dan hilang saat recreate (`docs/deploy-tidy-plan.md:33`). Kondisi aktual di VPS (termasuk riwayat deploy yang pernah menghapus unggahan) perlu dikonfirmasi ke auditsphere-devops. Unggahan juga tidak punya hash, sehingga keutuhan tidak bisa dibuktikan.
6. **Hash hanya ada di backup, dan tidak terlindungi.** `.sha256` disimpan di direktori yang sama dengan berkas, tidak ditandatangani, dan tidak dicatat di tempat terpisah. Pelaku dengan akses tulis bisa mengganti berkas dan hash sekaligus. Cron backup belum terpasang, dan salinan off-site masih DRAFT, sehingga hash belum memberi jaminan nyata.
7. **Backup tidak mencakup stack tenant.** `backup-prod.sh` menargetkan `rb_audit_postgres` dan `rb_audit_audit_service` saja. Basis data dan unggahan tenant di VPS lain atau stack tenant tidak dicakup skrip itu. Perlu diverifikasi cara pencadangan tenant [konfirmasi auditsphere-devops].
8. **Kontrol akses ke bukti lemah di sisi server.** Role hanya diperiksa di frontend untuk modul audit, master, dan risk. Pengguna dengan JWT valid dapat mengunduh bukti lewat URL. Ini melemahkan pernyataan bahwa bukti hanya diakses pihak berwenang. Tidak ada log akses unduhan yang tersimpan.
9. **Tidak ada penyimpanan bukti terpisah, NTP, atau kebijakan retensi.** Tidak ditemukan lokasi tujuan bukti insiden dengan akses terbatas, konfigurasi sinkronisasi waktu, kebijakan retensi bukti, atau legal hold. Retensi log dan arsip belum ditetapkan.
10. **Rahasia pernah tercatat di riwayat git.** `docs/security-hardening-changes-backend.md:9` mencatat nilai JWT_SECRET lama ada di riwayat git. Jika terjadi insiden penyalahgunaan token, bukti sesi lama tidak bisa dibedakan dari token yang dibuat dengan secret yang bocor sampai rotasi dijalankan (`docs/jwt-rotation-runbook.md:3`, masih "PREPARED ONLY"). Nilai tidak dicantumkan di dokumen ini.
11. **Aspek privasi pengumpulan bukti belum diatur.** Bukti dari basis data tenant berisi data pribadi (UU PDP). Belum ada aturan siapa yang boleh mengakses salinan bukti dan berapa lama disimpan, serta bagaimana izin klien diperoleh saat bukti berasal dari data tenant.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Sahkan FRP-001 dan sesuaikan dengan VPS Docker Compose: ganti contoh host dan lab fisik, tetapkan alat yang benar-benar dipakai. Jadikan Lampiran A sebagai dasar prosedur operasional. | Security Officer [konfirmasi] | Tinggi | [isi] | FRP-001 versi baru dengan pengesahan |
| 2 | Tetapkan evidence log (Lampiran B) dan formulir CoC (Lampiran D) di lokasi dengan akses terbatas, bukan di repositori. Buat penomoran Evidence ID. | Security Officer [konfirmasi] | Tinggi | [isi] | Register dan formulir yang disetujui |
| 3 | Putuskan lokasi penyimpanan bukti terpisah dari produksi (host atau bucket dengan akses terbatas), kebijakan retensi, dan legal hold. | Lead Information Security [konfirmasi] bersama Legal | Tinggi | [isi] | Keputusan tertulis |
| 4 | Tetapkan retensi dan rotasi log container (`logging:` dengan `max-size` dan `max-file`, atau log-driver terpusat) dan kirim log ke lokasi di luar container. Jika memakai Loki atau sejenisnya, lakukan lewat auditsphere-platform. | auditsphere-devops bersama auditsphere-platform | Tinggi | [isi] | Konfigurasi dan catatan retensi |
| 5 | Pasang penulis jejak audit: isi `sys_alert_logs` dari jalur login, logout, perubahan role, dan reset password, atau tambahkan tabel jejak audit baru untuk perubahan KKA, LHA, dan ATR (siapa, apa, nilai lama, nilai baru, waktu). Pasang `SecurityAlertHandler` sesuai tindakan A.5.24 nomor 6. | Tim Backend (auth-service dan audit-service) | Tinggi | [isi] | Perubahan kode dan contoh baris terisi di staging |
| 6 | Hitung dan simpan SHA-256 setiap berkas unggahan saat diunggah (kolom hash di `ActionTakenReportEvidence` dan model unggahan lain) dan tampilkan atau verifikasi saat diunduh. | Tim Backend (audit-service) | Sedang | [isi] | Perubahan kode dan uji |
| 7 | Verifikasi di VPS bahwa volume `audit_uploads` benar-benar terpasang, dan selesaikan volume `master_uploads` (`docs/deploy-tidy-plan.md:85`). Jalur deploy yang aman lewat `backend/deploy-prod.sh` dengan `--ref`, bukan redeploy bebas. | auditsphere-devops | Tinggi | [isi] | Hasil `docker inspect` dan catatan keputusan |
| 8 | Pasang cron backup dan salinan off-site (`docs/backup-b2-setup.md`), simpan keluaran skrip dan `restore-test.sh` ke berkas log, dan salin `.sha256` ke lokasi terpisah dari data. Cakup stack tenant. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Log backup dan restore test, daftar manifest hash |
| 9 | Jalankan latihan pengumpulan bukti (tabletop atau praktik pada staging) memakai Lampiran A sampai D, dan catat waktu koleksi serta temuan. Gabungkan dengan tabletop A.5.24 tindakan 11. | Security Officer [konfirmasi] | Tinggi | [isi] | Berita acara latihan, evidence log dan CoC terisi sebagai contoh |
| 10 | Tetapkan sinkronisasi waktu (NTP/chrony) pada VPS dan catat zona waktu yang dipakai pada log dan formulir. | auditsphere-platform | Sedang | [isi] | Hasil `chronyc tracking` atau setara |
| 11 | Tambahkan log akses unduhan berkas bukti (user, berkas, waktu, IP) dan pemeriksaan role di sisi server untuk endpoint unduh. | Tim Backend (audit-service) | Sedang | [isi] | Perubahan kode dan contoh log |
| 12 | Pasang LiME/Volatility atau tetapkan alternatif akuisisi memori pada VPS (ASC-003 NC-04), atau catat keputusan bahwa akuisisi memori tidak dilakukan. | auditsphere-devops bersama Security Officer [konfirmasi] | Sedang | [isi] | Keputusan atau instalasi yang terbukti |
| 13 | Jalankan rotasi JWT_SECRET sesuai runbook dengan persetujuan tertulis, agar token lama tidak bisa dipakai menyamar. Catat hasil tanpa menulis nilai. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Catatan rotasi dan persetujuan |
| 14 | Validasi aspek hukum: keabsahan bukti elektronik (UU ITE Pasal 5 dan 6), pemrosesan data pribadi selama akuisisi (UU PDP), dan izin klien untuk mengambil data tenant. | Legal [konfirmasi] | Tinggi | [isi] | Catatan validasi Legal |
| 15 | Perbaiki klaim A.8.15 dan "immutable log" di `docs/DOKUMEN_KESELARASAN_ISO27001.md` agar sesuai kondisi aktual sampai tindakan 5 selesai. | Lead Information Security [konfirmasi] | Sedang | [isi] | Revisi dokumen |
| 16 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index dan tandai dokumen yang belum disahkan. | Security Officer [konfirmasi] | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.28-01 | Prosedur forensik FRP-001 (urutan volatilitas, atribut CoC, metrik) dan prosedur operasional Lampiran A | `docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx`; Lampiran A dokumen ini | Internal (FRP-001 belum disahkan) |
| EV-A5.28-02 | Pengendalian teknis yang mendukung bukti: pemeriksaan dan pemindaian unggahan, soft delete, skema `sys_alert_logs`, backup dengan `.sha256`, uji restore | `backend/audit-service/pkg/upload/upload.go`, `backend/audit-service/pkg/avscan/avscan.go`, `backend/auth-service/models/sys_alert_log.go`, `backend/scripts/backup/` | Internal |
| EV-A5.28-03 | Evidence log insiden (belum tersedia). Templat ada di Lampiran B. | Belum tersedia. Akan dikumpulkan dari tindakan 2 dan 9. | Restricted |
| EV-A5.28-04 | Manifest hash SHA-256 dan formulir CoC terisi (belum tersedia). Templat ada di Lampiran C dan D. | Belum tersedia. Akan dikumpulkan dari tindakan 2, 8, dan 9. | Restricted |
| EV-A5.28-05 | Log backup dan restore test, konfigurasi retensi log, dan keputusan lokasi penyimpanan bukti (belum tersedia) | Belum tersedia. Akan dikumpulkan dari tindakan 3, 4, dan 8. | Restricted |

Setelah tindakan 1, 2, 8, dan 9 selesai, tambahkan ID untuk FRP-001 yang disahkan, register evidence log, dan berita acara latihan.

---

## Lampiran A. Templat Prosedur Forensik (usulan operasional untuk AuditSphere)

Usulan ini menurunkan FRP-001 ke lingkungan VPS Docker Compose. Belum disahkan dan belum diuji. Sesuaikan nama container dengan hasil `docker ps` pada saat kejadian. Nama di bawah mengikuti `backend/docker-compose.prod.yml` dan harus diperiksa ulang.

**A.1 Prinsip**

1. Amankan bukti sebelum memperbaiki. Jangan restart, recreate, atau deploy sebelum bukti volatil diambil, kecuali menunda itu menambah kerusakan. Catat keputusan dan alasannya.
2. Kerja pada salinan. Hitung hash master sekali, simpan master tanpa dibuka, analisis hanya working copy dan verifikasi hash working copy sama dengan master.
3. Semua tindakan dicatat di evidence log dan CoC dengan waktu (zona waktu dan sumber waktu dinyatakan).
4. Data pribadi di bukti diperlakukan sebagai Confidential. Ambil seminimal mungkin sesuai tujuan insiden.
5. Satu tenant satu cakupan. Jangan menyalin data tenant lain.
6. Gunakan akun individual yang tercatat, bukan akun bersama.

**A.2 Urutan tindakan**

| Langkah | Tindakan | Keterangan |
| :--- | :--- | :--- |
| 1 | Buka rekaman insiden: ID insiden, waktu deteksi, pelapor, tenant terdampak, dugaan jenis | Mengacu pada severity di `DOKUMEN_KESELARASAN_ISO27001.md` bagian 11.3 |
| 2 | Tunjuk Evidence Custodian dan Collector (dua orang berbeda bila memungkinkan), catat di Lampiran B | Pengumpul dan pemegang sebaiknya berbeda |
| 3 | Catat waktu server dan selisihnya dengan sumber waktu tepercaya | Perintah di A.3 |
| 4 | Putuskan apakah sistem perlu diisolasi (blokir di nginx/Kong, hentikan satu service) dan catat persetujuannya | Isolasi dilakukan setelah langkah 5 sampai 7 bila memungkinkan |
| 5 | Ambil bukti volatil (urutan volatilitas FRP-001 bagian 3): proses, koneksi, container, log berjalan | Perintah di A.3 |
| 6 | Ambil log container dan log host sebelum container dibuat ulang | Perintah di A.3 |
| 7 | Ambil salinan basis data terdampak (per tenant) dan unggahan terdampak | Perintah di A.3 |
| 8 | Hitung SHA-256 setiap berkas bukti, catat di manifest (Lampiran C), simpan master di penyimpanan bukti | Lampiran C |
| 9 | Isi formulir CoC (Lampiran D) dan catat setiap perpindahan | Lampiran D |
| 10 | Buat working copy, verifikasi hash, mulai analisis | Catat alat dan versinya |
| 11 | Setelah analisis: ringkas temuan, tentukan retensi atau legal hold, tutup rekaman | Libatkan Legal untuk insiden yang berpotensi proses hukum |
| 12 | Tinjau pasca insiden dan perbarui prosedur | Masukkan ke tindak lanjut A.5.24 |

**A.3 Perintah akuisisi (contoh, dijalankan oleh pengguna yang berwenang pada VPS)**

Tulis keluaran ke direktori kerja yang tidak berada di dalam container, dan jangan menulis ke volume produksi.

```bash
# Direktori kerja bukti (ganti INC-ID). Izin hanya untuk pemilik.
EVD=/root/evidence/INC-YYYYMM-NNN; mkdir -p "$EVD" && chmod 700 "$EVD"

# 1. Waktu dan sinkronisasi
date -u +%FT%TZ | tee "$EVD/00_waktu_server_utc.txt"
timedatectl >> "$EVD/00_waktu_server_utc.txt" 2>&1

# 2. Keadaan sistem (volatil)
{ ps auxww; ss -tunap; docker ps -a --no-trunc; docker stats --no-stream; } > "$EVD/01_state.txt" 2>&1

# 3. Log container sebelum recreate (satu berkas per container)
for c in $(docker ps -a --format '{{.Names}}' | grep '^rb_audit_'); do
  docker logs --timestamps "$c" > "$EVD/02_log_$c.txt" 2>&1
  docker inspect "$c" > "$EVD/02_inspect_$c.json"
done

# 4. Dump basis data terdampak (hanya basis data yang relevan, bukan pg_dumpall)
docker exec rb_audit_postgres pg_dump -U postgres --no-owner rb_audit_audit_<slug> | gzip -9 > "$EVD/03_db_audit_<slug>.sql.gz"

# 5. Unggahan terdampak (audit-service)
docker exec rb_audit_audit_service tar -C /root/uploads -czf - . > "$EVD/04_uploads.tar.gz"

# 6. Hash seluruh berkas bukti
( cd "$EVD" && sha256sum * > SHA256SUMS.txt && sha256sum -c SHA256SUMS.txt )
```

Catatan: nama basis data per tenant mengikuti pola `rb_audit_<svc>_<slug>` (`scripts/generate-tenant-stack.sh`). Untuk stack tenant, gunakan container dan direktori silo milik tenant itu. Verifikasi nama pada saat kejadian.

**A.4 Penyimpanan, retensi, dan pemusnahan**

- Master disalin ke penyimpanan bukti terpisah dengan akses terbatas (keputusan tindakan 3). Working copy hanya di host analisis.
- Retensi: [isi, ditetapkan Legal]. Legal hold menunda pemusnahan.
- Pemusnahan dicatat di bagian akhir CoC (tanggal, penyetuju, metode).

---

## Lampiran B. Templat Evidence Log

Satu baris per item bukti. Simpan di lokasi terbatas, bukan di repositori. Kolom diberi nomor agar bisa dipindah ke spreadsheet.

**Header rekaman insiden**

| Field | Isi |
| :--- | :--- |
| ID insiden | INC-YYYYMM-NNN |
| Tenant / sistem terdampak | [slug tenant, layanan] |
| Waktu deteksi (UTC dan WIB) | [isi] |
| Pelapor | [nama, peran] |
| Severity | Critical / Major / Minor (bagian 11.3) |
| Evidence Custodian | [nama] |
| Collector | [nama] |
| Sumber waktu dan selisih | [chrony/NTP, selisih detik] |
| Status | Terbuka / Dianalisis / Ditutup / Legal hold |

**Daftar bukti**

| No | Evidence ID | Deskripsi | Jenis | Sumber (host, container, path) | Waktu koleksi (UTC) | Metode / perintah | Pengumpul | Ukuran (byte) | SHA-256 | Lokasi master | Lokasi working copy | Klasifikasi | Status | Catatan |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | EVD-YYYYMM-[SISTEM]-[JENIS]-01 | | log / db / file / memori / jaringan / dokumen | | | | | | | | | Confidential | Master tersegel | |
| 2 | EVD-YYYYMM-[SISTEM]-[JENIS]-02 | | | | | | | | | | | | | |

Format Evidence ID mengikuti FRP-001 bagian 4: `EVD-YYYYMM-[SISTEM]-[JENIS]-NN`. Jenis yang dipakai: `LOG`, `DB`, `FILE`, `MEM`, `NET`, `DOC`.

---

## Lampiran C. Templat Pencatatan Hash (SHA-256)

**C.1 Aturan**

1. Hash dihitung pada master segera setelah koleksi, sebelum dibuka atau dipindahkan.
2. Hash master dan hash working copy dicatat dan harus sama. Perbedaan dicatat sebagai insiden integritas dan dilaporkan ke Security Officer.
3. Manifest hash disalin ke minimal dua lokasi berbeda (penyimpanan bukti dan pemegang evidence log). Idealnya dikirim sebagai pesan atau catatan ber-cap waktu agar tidak bisa diubah diam-diam.
4. Hash dihitung ulang setiap kali bukti diserahterimakan (kolom verifikasi di CoC).

**C.2 Perintah**

```bash
# Hitung
sha256sum <berkas> > <berkas>.sha256

# Verifikasi manifest (keluaran harus "OK" untuk setiap baris)
sha256sum -c SHA256SUMS.txt

# macOS (workstation analis)
shasum -a 256 <berkas>
```

**C.3 Manifest hash**

| Field | Isi |
| :--- | :--- |
| ID insiden | INC-YYYYMM-NNN |
| Dibuat oleh | [nama, peran] |
| Waktu pembuatan (UTC) | [isi] |
| Host yang menghitung | [isi] |
| Algoritma | SHA-256 |

| No | Evidence ID | Nama berkas | Ukuran (byte) | SHA-256 master | Waktu hitung (UTC) | SHA-256 working copy | Cocok (Ya/Tidak) | Diverifikasi oleh | Waktu verifikasi (UTC) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | | | | | | |
| 2 | | | | | | | | | |

Tanda tangan manifest: [nama, peran, tanggal]. Salinan manifest disimpan di: [lokasi 1], [lokasi 2].

---

## Lampiran D. Formulir Chain of Custody

Satu formulir per Evidence ID. Setiap perpindahan, akses, atau analisis menambah satu baris. Tidak boleh ada jeda waktu yang tidak tercatat.

**D.1 Identitas bukti**

| Field | Isi |
| :--- | :--- |
| Evidence ID | EVD-YYYYMM-[SISTEM]-[JENIS]-NN |
| ID insiden | INC-YYYYMM-NNN |
| Deskripsi barang bukti | [mis. dump basis data `rb_audit_audit_<slug>`, log container `rb_audit_audit_service`, arsip uploads] |
| Jenis bukti | LOG / DB / FILE / MEM / NET / DOC |
| Host dan lokasi sumber | [host, container, path] |
| Tenant | [slug] |
| Waktu koleksi (UTC dan WIB) | [isi] (sumber waktu: [isi]) |
| Petugas pengumpul | [nama, peran] |
| Saksi | [nama, peran] (bila ada) |
| Metode koleksi | [perintah atau alat, versi] |
| Ukuran (byte) | [isi] |
| Hash SHA-256 (master) | [64 karakter heksadesimal] |
| Kondisi segel / wadah | [mis. berkas read-only, disimpan pada media X, label Y] |
| Lokasi penyimpanan master | [isi] |
| Klasifikasi | Confidential / Restricted |
| Mengandung data pribadi | Ya / Tidak (dasar pemrosesan: [isi]) |

**D.2 Riwayat serah terima dan akses**

| No | Tanggal dan waktu (UTC) | Dari (nama, peran) | Kepada (nama, peran) | Tujuan | Lokasi asal | Lokasi tujuan | Hash diverifikasi (Ya/Tidak, nilai) | Kondisi bukti | Tanda tangan pengirim | Tanda tangan penerima |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | Penyimpanan awal | | | | | | |
| 2 | | | | | | | | | | |
| 3 | | | | | | | | | | |

**D.3 Analisis dan penggunaan**

| No | Tanggal dan waktu (UTC) | Analis | Working copy (path, hash) | Alat dan versi | Ringkasan tindakan | Hasil |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | | | |

**D.4 Pelepasan atau pemusnahan**

| Field | Isi |
| :--- | :--- |
| Keputusan | Disimpan (legal hold) / Diserahkan ke pihak lain / Dimusnahkan |
| Dasar keputusan dan penyetuju | [nama, peran, tanggal] (Legal [konfirmasi]) |
| Penerima (jika diserahkan) | [nama, institusi, dasar hukum] |
| Metode pemusnahan (jika dimusnahkan) | [isi] |
| Tanggal pelaksanaan dan pelaksana | [isi] |
| Hash terakhir yang diverifikasi sebelum dilepas | [isi] |

**D.5 Pengesahan**

| Peran | Nama | Tanda tangan | Tanggal |
| :--- | :--- | :--- | :--- |
| Evidence Custodian | | | |
| Security Officer | | | |
| Legal [konfirmasi] | | | |
