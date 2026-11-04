# A.5.24 — Perencanaan dan Persiapan Manajemen Insiden Keamanan Informasi

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Persiapan manajemen insiden (ISO/IEC 27001:2022 A.5.24: Perencanaan dan persiapan manajemen insiden keamanan informasi) |
| **Fokus pemeriksaan** | Peran, proses, kanal, dan alat respons insiden siap digunakan |
| **Contoh bukti objektif** | IR plan, playbook, contact tree, hasil simulasi |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Lead Information Security [konfirmasi] (tidak tercantum di baris checklist) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 29 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere menyimpan data audit klien, kredensial, dan data pribadi. Kejadian seperti kebocoran data, kompromi akun admin, atau kegagalan layanan memerlukan respons yang sudah disiapkan. Kebijakan `CSP-001` butir 20 juga mengatur penanganan insiden di lingkungan cloud. |
| Implementasi status | **Sebagian** | Bagian kebijakan dan istilah sudah ada di `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 11 (deteksi, triage, containment, eskalasi, severity), dan dokumen FRP-001, RTP-001, serta ASC-003 sudah disusun. Tetapi IR plan dan playbook yang bisa dijalankan, contact tree yang berisi kontak, dan kanal deteksi otomatis belum terbukti di repositori. |
| Ketersediaan bukti | **Sebagian** | Ada prosedur tingkat kebijakan (bagian 3). Tidak ada playbook, contact tree, atau catatan simulasi di repositori. File escalation tree yang disebut `CSP-001` tidak ada di repositori. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.24-01 s/d EV-A5.24-04) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada catatan insiden, tabletop, atau drill yang terekam. Dokumen `RTP-001` dan `ASC-003` mencatat simulasi sebagai rencana atau temuan terbuka (NC-02). Restore test backup belum tercatat hasilnya. |
| Desain kontrol | **Perlu perbaikan** | Alur eskalasi ada di `DOKUMEN_KESELARASAN_ISO27001.md` bagian 11, tetapi belum ada kanal deteksi yang terpasang, tidak ada kewajiban notifikasi regulator yang masuk ke bagian 11, dan ada rujukan ke dokumen yang tidak ada (IRP-001, IRP-002, escalation tree). |
| Status kesiapan | **Belum siap** | Penerapan, bukti, dan pengujian belum memenuhi gate. |
| Prioritas | **Tinggi** | Dampak kebocoran data klien dan tenggat notifikasi (CSP-001 butir 20: 1x24 jam ke CSP dan 3x24 jam ke subjek data dan regulator) menuntut respons yang sudah siap sebelum insiden terjadi. |
| Review berikutnya | [isi] | Usulan: 3 bulan setelah IR plan, contact tree, dan playbook disetujui, lalu setiap tahun atau setelah simulasi atau insiden signifikan (sesuai periode review ASC-003). |

> Catatan pembacaan checklist: pada tab `Checklist 93 Kontrol` (sheet3), baris A.5.24 (baris 30) kolom G s/d Z belum diisi. Nomor kontrol A.5.24 dikonfirmasi dari workbook `03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet3, baris 30) dan sesuai dengan `docs/DOKUMEN_KESELARASAN_ISO27001.md` baris 324. PIC tidak tercantum pada baris ini, sehingga dokumen ini memakai "Lead Information Security" sebagai usulan dan perlu dikonfirmasi.

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.24 **berlaku** untuk AuditSphere karena:

1. Platform menyimpan data audit dan data pribadi klien (lihat `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 13.3) serta kredensial pengguna. Kebocoran atau kompromi akun admin termasuk kategori **Critical** di `docs/DOKUMEN_KESELARASAN_ISO27001.md:458`.
2. Kebijakan `CSP-001` (`docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx`, butir 20 "Manajemen Insiden Keamanan di Lingkungan Cloud") mewajibkan koordinasi insiden sesuai IRP-002, eskalasi ke CSP maksimal 1x24 jam, dan pemberitahuan ke subjek data serta regulator maksimal 3x24 jam sesuai UU PDP.
3. Dokumen `FRP-001` dan `ASC-003` (`docs/ISO_27001/advanced_security/`) juga mensyaratkan prosedur insiden sebagai dasar kesiapan forensik dan simulasi. Tanpa IR plan dan playbook yang berlaku, kedua dokumen itu tidak bisa dijalankan.
4. Klien enterprise dan bank pada due diligence akan meminta bukti kesiapan respons insiden, termasuk siapa yang dihubungi dan kapan klien diberi tahu.

---

## 2. Implementasi yang Sudah Ada

Setiap butir diverifikasi dari repositori. Path dan baris mengacu pada cabang `staging`. Tidak ada nilai rahasia (kunci API, token, nomor kontak) yang dicantumkan di dokumen ini.

### 2.1 Kebijakan dan prosedur tingkat organisasi

| Dokumen / bagian | Isi | Lokasi | Status verifikasi |
| :--- | :--- | :--- | :--- |
| Manajemen Insiden Keamanan, bagian 11 (`DOKUMEN_KESELARASAN_ISO27001.md`) | Prosedur deteksi dan eskalasi (11.1), eskalasi ke leadership dengan format Leadership OS AIFL (11.2), dan klasifikasi severity Critical, Major, Minor dengan target respons (11.3) | `docs/DOKUMEN_KESELARASAN_ISO27001.md:429-462` (11.1 di baris 431, 11.2 di baris 440, 11.3 di baris 454-460) | **Terverifikasi (teks kebijakan).** Belum berupa IR plan operasional. |
| Peran respons insiden (RACI) | "Respons Insiden Keamanan": Security R/A, Infra R, Development C, QA I | `docs/DOKUMEN_KESELARASAN_ISO27001.md:507-520` (baris 515) | **Terverifikasi.** Kolom "Infra / Proxmox" (baris 509) tidak sesuai dengan produksi yang berjalan di VPS (lihat gap 5). |
| Status kontrol di dokumen keselarasan | "A.5.24 s/d A.5.28: Partially Implemented" | `docs/DOKUMEN_KESELARASAN_ISO27001.md:324` | **Terverifikasi (klaim dokumen).** Sejalan dengan status Sebagian di checklist ini. |
| Forensic Readiness Procedure (FRP-001, DOK-120) | Akuisisi bukti digital, formulir Chain of Custody, target waktu koleksi bukti ≤ 4 jam sejak eskalasi. Rujukan ke IRP-001. | `docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx` | **Ada (dokumen).** Lembar pengesahan berisi placeholder "[Ditandatangani Digital]". Rujukan IRP-001 tidak ditemukan. |
| Red/Purple Team Exercise Plan (RTP-001, DOK-118) | Jenis latihan termasuk **Tabletop Exercise** (simulasi skenario krisis berbasis diskusi). Ada metrik deteksi dan containment. | `docs/ISO_27001/advanced_security/RTP-001_Red_Team_Purple_Team_Exercise_Plan.docx` | **Ada (rencana).** Tidak ada catatan pelaksanaan. Lembar pengesahan berisi placeholder. |
| Advanced Security Operations Checklist (ASC-003, v1.0, berlaku 28 September 2026) | Status implementasi "Siap Diisi / Sedang Berjalan", daftar NC dengan status Open, target MTTR < 4 jam | `docs/ISO_27001/advanced_security/Advanced_Security_Operations_Checklist.docx` (NC-01 s/d NC-04 di bagian 10) | **Ada (checklist).** NC-02 (simulasi Red/Purple Team belum dieksekusi di produksi) dan NC-04 (LiME/Volatility belum terpasang di VPS produksi) berstatus Open. |
| Kebijakan cloud (CSP-001 butir 20) | IRP-002 sebagai dasar, eskalasi CSP 1x24 jam, notifikasi subjek data dan regulator 3x24 jam | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` (butir 20.1, 20.2, 20.4) | **Ada (kebijakan).** IRP-002 tidak ditemukan di repositori. Butir ini tidak tercermin di bagian 11 `DOKUMEN_KESELARASAN_ISO27001.md`. |
| Rujukan escalation tree | `CSP-001` (lampiran bukti, baris "Tanggap Darurat (A.5.24)") menyebut "Penyusunan kontak darurat (DevOps & Security) serta mekanisme Emergency Stop" dengan bukti `ORG-IncidentPrep-202610-EscalationTree-v1.pdf` | `CSP-001` (lampiran) | **File tidak ada di repositori** (pencarian nama `*EscalationTree*` nihil). Status "Siap" di lampiran tidak bisa diverifikasi. |

### 2.2 Pengendalian teknis yang bisa dipakai saat respons

| Pengendalian | Fungsi saat insiden | Lokasi | Status verifikasi |
| :--- | :--- | :--- | :--- |
| Kafka event untuk kejadian keamanan auth (publisher) | Menyediakan sinyal `user.login_failed`, `account.locked`, `mfa.failed`, dan lainnya | `backend/auth-service/pkg/kafka/events.go:196` (`PublishLoginFailed`); `backend/auth-service/pkg/kafka/events.go:82`; `backend/auth-service/services/mfa/mfa_service.go:161` (MFA gagal) | **Sebagian.** Publisher dibuat di `backend/auth-service/pkg/registry/registry.go:47`. Fungsi `PublishLoginFailed` tidak punya pemanggil di luar definisinya (pencarian `PublishLoginFailed` dan `PublishAccountLocked` hanya menemukan definisi). Sinyal login gagal belum terkirim. |
| Handler keamanan (SecurityAlertHandler) | Mencatat event keamanan ke logger | `backend/auth-service/pkg/kafka/events.go:407` (`NewSecurityAlertHandler`), `:410-427` | **Belum terpasang.** Konstruktor hanya didefinisikan. Tidak ditemukan pemanggil di `cmd/` atau `registry/` (pencarian `NewSecurityAlertHandler`). Pola yang sama ada di `master-service` dan `audit-service`. |
| Pembatasan laju di gateway | Membatasi percobaan berlebih per menit dan per jam | `backend/kong-gateway/kong/prod/kong.yml:80-84` (rate-limiting, 100/menit dan 5000/jam) | **Terverifikasi (konfigurasi).** Bukan pengganti deteksi. Tidak ada kunci akun per pengguna. |
| Penguncian akun setelah gagal login | Mengunci akun setelah percobaan gagal | `backend/auth-service/models/user.go:28` (kolom `LockedUntil`); `backend/auth-service/services/user/user_service.go:469` (hanya mengosongkan) | **Tidak berfungsi.** Tidak ada kode yang mengisi `LockedUntil`. Tabel severity Minor di `DOKUMEN_KESELARASAN_ISO27001.md:460` menyebut "percobaan brute force yang berhasil diblokir", yang belum didukung kode. |
| Pemantauan (monitoring) dan alert | Mendeteksi gangguan dan anomali sesuai bagian 11.1 | Tidak ada service Prometheus, Grafana, Loki, atau Alertmanager di `backend/docker-compose*.yml`. Kong hanya mengaktifkan plugin `prometheus` untuk ekspor metrik (`backend/kong-gateway/kong/prod/kong.yml:23`). Healthcheck container ada di `backend/docker-compose.prod.yml` (baris 24, 41, 71, 276). | **Tidak terpasang.** `DOKUMEN_KESELARASAN_ISO27001.md:432` menyebut "alert sistem monitoring (Prometheus/Grafana)", padahal tidak ada di repositori. |
| Notifikasi alert ke kanal (Telegram, Slack, email on-call) | Mengirim alert ke tim respons | Pencarian `telegram`, `slack`, `pagerduty`, `opsgenie`, `alertmanager` di `backend/`, `scripts/`, `.github/`: tidak ditemukan pengiriman alert. Satu-satunya rujukan Slack ada di `docs/DASHBOARD-OPERATIONS-QUICK-REFERENCE.md:431` dan `docs/DASHBOARD-IMPLEMENTATION-SUMMARY.md:272`, sebagai rencana. | **Tidak terpasang.** |
| Notifikasi kegagalan deploy | Memberi tahu tim jika deploy gagal | `.github/workflows/deploy.yml:88` (`if: always()`, hanya ringkasan di log job) | **Tidak ada notifikasi.** Hanya ringkasan status di log. |
| Log aplikasi | Bahan investigasi forensik | `zap` ke stdout di setiap service (contoh `backend/auth-service/pkg/logger/logger.go`). Tidak ada konfigurasi retensi atau log-driver di `backend/docker-compose.prod.yml`. | **Sebagian.** Retensi, rotasi, dan pengiriman log ke lokasi terpusat belum terkonfirmasi. |
| Tabel `user_access_audit_logs` | Dugaan sumber jejak akses | Nama tabel hanya muncul sebagai string nama sumber data di `backend/master-service/routes/routes.go:959` dan `:1041`. Tidak ditemukan model atau migrasi yang membuatnya di `backend/auth-service` atau `backend/audit-service`. | **Jangan dipakai sebagai bukti.** Tabel ini tidak terbukti ada sebagai jejak akses. |
| Backup produksi (pemulihan) | Pemulihan data setelah insiden atau ransomware | `backend/scripts/backup/backup-prod.sh` (pg_dumpall dan arsip uploads); `backend/scripts/backup/restore-test.sh` (uji restore ke container sementara) | **Skrip ada, belum berjalan terjadwal.** Lihat bagian 3 dan gap 3. |
| Backup off-site (Backblaze B2) | Salinan di luar VPS | `docs/backup-b2-setup.md:6` ("Status: DRAFT. Not installed on the VPS."); jadwal uji restore bulanan di `docs/backup-b2-setup.md:109` | **Rencana.** Belum terpasang. |
| Runbook rotasi JWT_SECRET | Containment: membatalkan semua sesi dan token yang diduga bocor | `docs/jwt-rotation-runbook.md:3` ("PREPARED ONLY. Production has not been changed.") | **Disiapkan, belum dijalankan.** `docs/security-hardening-changes-backend.md:9` mencatat bahwa nilai lama pernah tercatat di riwayat git. Nilai tidak dicantumkan di dokumen ini. |
| Kemampuan pemutusan koneksi dan token | Containment sesuai bagian 11.1 langkah 3 | Tidak ditemukan prosedur atau fitur "emergency stop" di `backend/`, `scripts/`, atau `frontend/` (pencarian `emergency stop`, `kill switch` nihil). Revokasi sesi massal hanya lewat rotasi secret. | **Belum terbukti.** |
| Kanal pelaporan insiden dari pengguna | Pintu masuk laporan (security report ticket, bagian 11.1 langkah 1) | Tidak ditemukan alamat atau formulir pelaporan keamanan di `frontend/pages/`, `docs/`, atau `backend/` (pencarian `security@`, `security.txt`, `laporan keamanan` nihil). | **Belum ada.** |

### 2.3 Hal yang belum ditemukan di repositori

- IR plan operasional (IRP-001 dan IRP-002 tidak ditemukan).
- Playbook per skenario (kebocoran data, ransomware, kompromi akun admin, outage).
- Contact tree atau daftar kontak darurat (file `ORG-IncidentPrep-202610-EscalationTree-v1.pdf` tidak ada).
- Template laporan insiden (incident record, timeline, decision log, formulir notifikasi).
- Catatan tabletop, drill, atau simulasi yang sudah dijalankan.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| IR plan (IRP-001 / IRP-002) | **Belum tersedia** | Tidak ditemukan. Bagian 11 `DOKUMEN_KESELARASAN_ISO27001.md` berupa ringkasan kebijakan, bukan rencana operasional. |
| Playbook per skenario | **Belum tersedia** | Tidak ditemukan di `docs/`, `scripts/`, atau `backend/`. |
| Contact tree / escalation tree | **Belum tersedia** | Disebut di `CSP-001` tetapi file tidak ada di repositori. Kontak tidak boleh ditulis di repositori. |
| Kriteria severity dan SLA | **Tersedia (kebijakan)** | `DOKUMEN_KESELARASAN_ISO27001.md:454-460`. Nilai SLA perlu disamakan dengan FRP-001 dan ASC-003 (lihat gap 4). |
| Prosedur forensik (FRP-001) | **Tersedia (dokumen, belum disahkan)** | `docs/ISO_27001/advanced_security/FRP-001_Forensic_Readiness_Procedure.docx`. Lembar pengesahan berisi placeholder. |
| Rencana tabletop / red team (RTP-001) | **Tersedia (rencana)** | `docs/ISO_27001/advanced_security/RTP-001_Red_Team_Purple_Team_Exercise_Plan.docx`. Tidak ada catatan pelaksanaan. |
| Hasil simulasi / tabletop | **Belum tersedia** | Tidak ada catatan. ASC-003 NC-02 berstatus Open. |
| Catatan insiden nyata | **Belum tersedia / tidak ada di repositori** | Tidak ditemukan catatan insiden di repositori. Status ini tidak berarti tidak pernah ada insiden. Konfirmasi ke Lead Information Security [konfirmasi]. |
| Hasil restore test backup | **Belum tersedia** | Skrip `backend/scripts/backup/restore-test.sh` ada. Tidak ada log hasil. Jadwal bulanan di `docs/backup-b2-setup.md:109` belum berjalan karena dokumen masih DRAFT. |
| Bukti containment (rotasi JWT) | **Belum tersedia** | Runbook ada (`docs/jwt-rotation-runbook.md`), belum dijalankan. |
| Kebijakan notifikasi (CSP 1x24 jam, subjek data dan regulator 3x24 jam) | **Sebagian** | Tercantum di `CSP-001` butir 20. Tidak ada di `DOKUMEN_KESELARASAN_ISO27001.md` bagian 11 maupun 13.3. Teks UU PDP belum diverifikasi di repositori. |

---

## 4. Catatan / Gap Terverifikasi

1. **Dokumen respons insiden operasional belum ada.** Tidak ditemukan IR plan (IRP-001 atau IRP-002), playbook, atau contact tree. Bagian 11 `DOKUMEN_KESELARASAN_ISO27001.md` adalah ringkasan kebijakan. `FRP-001`, `CSP-001` butir 20, dan lampiran `CSP-001` merujuk ke dokumen yang tidak ada, sehingga status "Siap" di lampiran `CSP-001` tidak bisa dibuktikan dari repositori.
2. **Sinyal deteksi belum tersambung.** Fungsi `PublishLoginFailed` dan `PublishAccountLocked` tidak punya pemanggil, `SecurityAlertHandler` tidak dipasang, dan tidak ada monitoring atau alert yang dikirim ke tim. Klaim "alert sistem monitoring (Prometheus/Grafana)" di `DOKUMEN_KESELARASAN_ISO27001.md:432` tidak didukung repositori. Kesimpulannya, untuk saat ini insiden hanya akan terdeteksi dari laporan manual.
3. **Belum ada simulasi atau bukti latihan.** RTP-001 mendefinisikan tabletop dan red team, tetapi tidak ada catatan pelaksanaan. ASC-003 NC-02 (simulasi belum dieksekusi di produksi) dan NC-04 (LiME/Volatility belum terpasang) berstatus Open. Restore test backup juga belum tercatat hasilnya.
4. **Severity, SLA, dan tenggat notifikasi belum selaras antar dokumen.** Di `DOKUMEN_KESELARASAN_ISO27001.md:458`, Critical harus direspons dalam < 15 menit (eskalasi ke Dimas). FRP-001 menargetkan koleksi bukti ≤ 4 jam. ASC-003 menargetkan MTTR < 4 jam. `CSP-001` butir 20 menetapkan CSP 1x24 jam dan subjek data serta regulator 3x24 jam. Bagian 11 dan 13.3 `DOKUMEN_KESELARASAN_ISO27001.md` tidak memuat tenggat notifikasi. Teks UU PDP belum diverifikasi di repositori dan perlu diperiksa Legal [konfirmasi].
5. **Peran RACI belum sesuai dengan kondisi produksi.** Kolom "Infra / Proxmox" di `DOKUMEN_KESELARASAN_ISO27001.md:509` tidak sesuai dengan VPS yang menjalankan Docker Compose. Penanggung jawab respons insiden juga belum tercatat dengan nama. Dokumen memakai "Security" dan "Lead Information Security / Tech Lead AIFL", dan FRP-001 memakai "Security Officer". Pemetaan peran perlu disamakan.
6. **Containment belum siap dijalankan.** Runbook rotasi JWT hanya disiapkan (`docs/jwt-rotation-runbook.md:3`). Tidak ada prosedur pemutusan akses atau sesi per pengguna, dan tidak ada mekanisme "emergency stop" (pencarian nihil). Penguncian akun setelah gagal login tidak berfungsi (`LockedUntil` tidak pernah diisi).
7. **Jejak audit untuk investigasi belum terkonfirmasi.** Tabel `user_access_audit_logs` tidak ditemukan sebagai model atau migrasi (hanya nama sumber data di `backend/master-service/routes/routes.go:959`). Log aplikasi belum punya kebijakan retensi atau pengiriman terpusat. Bukti untuk A.8.15 dan A.5.28 perlu dikaji ulang sebelum dipakai sebagai dasar respons.
8. **Kanal pelaporan dari pengguna belum ada.** Tidak ditemukan alamat, formulir, atau halaman untuk melaporkan insiden atau kerentanan. Bagian 11.1 langkah 1 mengandalkan "security report ticket" yang belum tersedia.
9. **Notifikasi kegagalan deploy tidak ada.** `.github/workflows/deploy.yml:88` hanya mencatat ringkasan di log. Jika deploy gagal saat insiden, tim tidak diberi tahu secara otomatis.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Susun IR plan operasional (IRP-001) dari bagian 11 `DOKUMEN_KESELARASAN_ISO27001.md`. Sertakan peran, langkah deteksi sampai pemulihan, kanal, dan tenggat. Samakan dengan rujukan IRP-002 di `CSP-001` atau ubah rujukan itu. | Lead Information Security [konfirmasi] | Tinggi | [isi] | IR plan yang disetujui, dengan tanggal dan pemilik |
| 2 | Bangun ulang contact tree dan escalation tree. Simpan kontak di lokasi terbatas, bukan di repositori. Pulihkan file `ORG-IncidentPrep-202610-EscalationTree-v1.pdf` atau catat bahwa file tersebut belum ada. | Lead Information Security [konfirmasi] bersama Dimas (Group Vision) | Tinggi | [isi] | Contact tree (akses restricted), tanggal verifikasi kontak |
| 3 | Tetapkan playbook minimal untuk skenario kebocoran data klien, kompromi akun admin, ransomware, dan outage total. Masukkan langkah containment, termasuk rotasi JWT dari `docs/jwt-rotation-runbook.md`. | Lead Information Security [konfirmasi] bersama auditsphere-devops | Tinggi | [isi] | Playbook per skenario yang disetujui |
| 4 | Tambahkan tenggat notifikasi (CSP 1x24 jam; subjek data dan regulator 3x24 jam sesuai `CSP-001` butir 20) ke bagian 11 dan 13.3 `DOKUMEN_KESELARASAN_ISO27001.md`. Verifikasi teks UU PDP. Samakan SLA dengan FRP-001 dan ASC-003. | Lead Information Security [konfirmasi] bersama Legal | Tinggi | [isi] | Revisi dokumen dengan tanggal, dan catatan verifikasi UU PDP |
| 5 | Perbaiki kolom RACI "Infra / Proxmox" menjadi peran yang sesuai dengan VPS produksi. Catat nama penanggung jawab. | Lead Information Security [konfirmasi] | Sedang | [isi] | Revisi bagian 14 |
| 6 | Hubungkan sinyal login gagal, akun terkunci, dan MFA gagal ke handler yang aktif. Panggil `PublishLoginFailed` dan publisher lain dari jalur login, lalu pasang `SecurityAlertHandler` di registry. | Tim Backend (auth-service) bersama auditsphere-devops | Tinggi | [isi] | Perubahan kode dan bukti event tercatat di log saat percobaan login gagal (uji di staging) |
| 7 | Putuskan kebijakan penguncian akun setelah gagal login (mengisi `LockedUntil` atau menghapus klaim "brute force diblokir" dari tabel severity). | Tim Backend (auth-service) bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Keputusan tertulis dan perubahan kode atau revisi dokumen |
| 8 | Pasang monitoring dan alert minimal (ketersediaan layanan, laju error, disk, dan status Kafka) dengan kanal yang sudah disepakati. Perbaiki klaim Prometheus/Grafana di bagian 11.1 jika belum dipasang. | auditsphere-platform bersama auditsphere-devops | Tinggi | [isi] | Daftar alert aktif dan contoh notifikasi yang terkirim |
| 9 | Tentukan kebijakan retensi log container dan pengiriman ke lokasi terpusat. Catat periode retensi di bagian 2.2. | auditsphere-devops | Sedang | [isi] | Konfigurasi log dan catatan retensi |
| 10 | Tetapkan kanal pelaporan insiden dan kerentanan dari pengguna (alamat atau formulir), lalu tampilkan di halaman yang relevan. | Lead Information Security [konfirmasi] bersama auditsphere-frontend | Sedang | [isi] | Kanal yang sudah aktif dan dokumennya |
| 11 | Jalankan tabletop exercise sesuai RTP-001 dengan skenario kebocoran data dan kompromi akun admin. Catat peserta, waktu, keputusan, dan temuan. | Lead Information Security [konfirmasi] bersama Dimas (Group Vision) | Tinggi | [isi] | Berita acara tabletop dan daftar tindak lanjut |
| 12 | Jalankan restore test (`backend/scripts/backup/restore-test.sh`) secara terjadwal dan simpan log hasilnya. Selesaikan tindakan backup off-site dari `docs/backup-b2-setup.md` sebelum mengklaim pemulihan. Ini juga menjadi bukti A.8.13. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Log restore test dan catatan jadwal |
| 13 | Tentukan apakah rotasi JWT_SECRET dijalankan sebagai latihan containment. Jalankan hanya dengan persetujuan tertulis dan jadwal yang disepakati. Catat hasilnya tanpa menulis nilai secret. | auditsphere-devops bersama Lead Information Security [konfirmasi] | Tinggi | [isi] | Catatan latihan atau rotasi, dan keputusan persetujuan |
| 14 | Buat template laporan insiden (incident record, timeline, decision log, formulir notifikasi, dan Chain of Custody dari FRP-001). Simpan di lokasi restricted. | Lead Information Security [konfirmasi] | Sedang | [isi] | Template yang disetujui |
| 15 | Tambahkan notifikasi otomatis saat deploy gagal ke kanal tim. Ini belum ada di `.github/workflows/deploy.yml:88`. | auditsphere-devops | Rendah | [isi] | Perubahan workflow dan bukti notifikasi uji |
| 16 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index, dan tandai dokumen yang belum disahkan (FRP-001, RTP-001, ASC-003). | Lead Information Security [konfirmasi] | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.24-01 | Prosedur insiden tingkat kebijakan (deteksi, triage, containment, eskalasi, severity) dan RACI | `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 11 dan 14 | Internal |
| EV-A5.24-02 | Dokumen kesiapan: FRP-001 (forensik), RTP-001 (tabletop dan red team), ASC-003 (checklist dan NC) | `docs/ISO_27001/advanced_security/` | Internal (sebagian belum disahkan) |
| EV-A5.24-03 | Pengendalian teknis yang bisa dipakai saat respons: event keamanan auth, rate limit gateway, runbook rotasi JWT, skrip backup dan restore test | `backend/auth-service/pkg/kafka/events.go`, `backend/kong-gateway/kong/prod/kong.yml`, `docs/jwt-rotation-runbook.md`, `backend/scripts/backup/` | Internal |
| EV-A5.24-04 | IR plan, playbook, contact tree, dan hasil simulasi (belum tersedia). Escalation tree dirujuk di `CSP-001` tetapi tidak ada di repositori. | Belum tersedia. Akan dikumpulkan dari tindakan 1, 2, 3, dan 11. | Restricted |

Setelah tindakan 1, 2, dan 11 selesai, tambahkan ID untuk IR plan yang disahkan, contact tree (restricted), dan berita acara tabletop.
