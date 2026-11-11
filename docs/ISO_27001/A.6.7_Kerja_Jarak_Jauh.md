# A.6.7 — Kerja Jarak Jauh

| Field | Isi |
| :--- | :--- |
| **Tema** | People |
| **Judul kontrol** | Kerja jarak jauh |
| **Fokus pemeriksaan** | Risiko remote working dikendalikan pada perangkat, koneksi, dan lingkungan |
| **Contoh bukti objektif** | Remote work policy, VPN/MFA config, endpoint report, training |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi & Cloud SaaS |
| **PIC** | Lead Information Security (didukung Legal & Compliance, HR, dan DevOps Lead) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 31 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti. Kondisi perangkat personel (enkripsi disk, patch, antivirus, screen lock, MDM), konfigurasi server VPS (firewall, `sshd`, daftar `authorized_keys`), pengaturan GitHub, konsol layanan cloud, dan kontrak/SDM berada di luar repo dan ditandai **Perlu verifikasi**; tidak ada satu pun yang dapat dibuktikan dari repo. Seluruh isian pada Lampiran A sampai I (kebijakan, ambang teknis, templat, opsi VPN, RACI) adalah **draf usulan**: belum disetujui manajemen, belum ada salinan terisi, dan angka (hari patch, menit screen lock, jam pelaporan) bukan hasil pengukuran. Nilai rahasia, kunci, dan alamat IP tidak dicantumkan.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Tim kecil bekerja dari lokasi berbeda tanpa jaringan kantor yang dapat dibuktikan dari repo; seluruh akses ke repo GitHub, VPS (SSH), Kong/monitoring, konsol cloud, dan alat AI dilakukan dari perangkat dan jaringan personel. Pengguna klien/tenant juga mengakses aplikasi dari jaringan dan perangkat mereka sendiri. |
| Implementasi status | **Sebagian** | Ada kendali teknis pada jalur akses jarak jauh: deploy via SSH dengan host key dipin (`VPS_KNOWN_HOSTS`), MFA TOTP dan perangkat tepercaya di aplikasi, rate limit dan CORS di Kong, TLS di edge. Tidak ada kebijakan kerja jarak jauh, VPN/bastion/IP allowlist untuk personel, kendali endpoint, laporan endpoint, maupun pelatihan khusus. Draf kebijakan dan templat diusulkan pada Lampiran A sampai I. |
| Ketersediaan bukti | **Ada sebagian** | Tersedia bukti kode dan konfigurasi di repo (bagian 2). Remote work policy, konfigurasi VPN personel, endpoint report, dan catatan pelatihan **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A6.7-01 s/d EV-A6.7-10) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada catatan tinjauan akses jarak jauh, pengukuran kepatuhan perangkat, ataupun uji pelaporan perangkat hilang. Beberapa kendali teknis ada tetapi punya celah (port admin terbuka di compose, MFA tidak diwajibkan, sesi tidak dicabut saat akun nonaktif, jalur deploy lama tanpa pin host key). |
| Desain kontrol | **Belum dirancang** | Draf kebijakan (RWP-001 usulan), baseline perangkat, model akses privileged, dan opsi VPN diusulkan di Lampiran A sampai C; belum disetujui. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Satu laptop personel yang hilang atau terinfeksi dapat membawa kunci SSH ke VPS produksi, token GitHub, dan akses konsol cloud sekaligus. Klien lembaga jasa keuangan akan menanyakan kendali perangkat dan akses jarak jauh pada due diligence. |
| Review berikutnya | [isi] | Usulan: tinjauan akses jarak jauh triwulanan (Lampiran G), evaluasi desain 12 bulan setelah kebijakan disetujui, dan sinkron dengan audit internal ISMS tahunan (ISO-AS-001 bagian 13.2). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.6.7 **berlaku** untuk AuditSphere karena:

1. AuditSphere menyimpan kertas kerja (KKA), laporan (LHA), tindak lanjut (ATR), dan kredensial klien enterprise (lembaga jasa keuangan). Seluruh personel yang menyentuh data atau produksi bekerja dari lokasi masing-masing; tidak ditemukan di repo jaringan kantor, VPN, bastion, atau IP allowlist untuk personel. Perimeter praktisnya adalah akun dan perangkat personel.
2. Akses privileged ke produksi dilakukan dari jarak jauh: deploy lewat SSH dari GitHub Actions (`.github/workflows/deploy.yml`) dan dari workstation operator (`backend/deploy-prod.sh`), serta sesi SSH manual ke VPS yang di dokumentasi runbook memakai pengguna `root` (`docs/BACKEND_RUN_AND_DEPLOY.md:75-77`, `docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md:33`).
3. Permukaan layanan produksi cukup lebar: compose produksi memetakan port layanan dan admin API Kong tanpa binding `127.0.0.1` (bagian 4 butir 5), sehingga perlindungan sangat bergantung pada firewall host yang **perlu verifikasi**.
4. Personel memakai alat AI pada perangkat kerja (riwayat git memuat agen AI; A.5.32 bagian 1 butir 4). Agen yang berjalan di laptop dapat membaca berkas lokal, termasuk `.env`. AIP-001 melarang kredensial dan data klien ke AI publik, tetapi tidak ada kendali teknis di perangkat (A.6.3 bagian 4 butir 4).
5. Kontributor lebih dari satu orang (6 nama penulis di riwayat git, termasuk bot) dengan status karyawan/kontraktor yang tidak dapat diverifikasi dari repo (A.5.32 bagian 2.4). Perangkat kontraktor dan BYOD paling sulit dikendalikan.
6. Layanan cloud pihak ketiga diakses dari konsol web oleh personel: GitHub, Cloudflare, Rumahweb, Resend, ImageKit, Backblaze B2, dan alat AI. Kompromi satu akun konsol (mis. DNS di Cloudflare, atau kunci pengiriman email) berdampak lintas tenant. Status MFA masing-masing **perlu verifikasi**.
7. Pengguna klien/tenant mengakses aplikasi dari jaringan dan perangkat yang tidak dikelola AuditSphere. Yang dapat dikendalikan AuditSphere adalah sisi aplikasi (MFA, perangkat tepercaya, sesi, rate limit), bukan endpoint klien; pembagian tanggung jawab perlu ada di kontrak/DPA (A.5.20).
8. CSP-001 (bagian 10 butir 2 dan bagian 11 butir 5) mewajibkan MFA untuk seluruh pengguna konsol cloud dan akses administratif lewat VPN gateway atau PrivateLink; ZTAR-001 menyebut MDM dan EDR sebagai pilar perangkat. Keduanya belum punya bukti penerapan di repo.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Kendali teknis akses jarak jauh personel (terverifikasi dari repo)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| Deploy produksi manual dengan persetujuan lingkungan | `workflow_dispatch`, `environment: production`, `concurrency` tunggal. Komentar workflow: konfigurasikan *required reviewers* sebelum dipakai. | `.github/workflows/deploy.yml:14`, `:27`, `:37-39` | Pengaturan *required reviewers* di GitHub **tidak dapat diverifikasi dari repo**. |
| Host key VPS dipin, bukan `ssh-keyscan` | Workflow menolak berjalan bila `VPS_KNOWN_HOSTS` kosong; kunci ditulis dengan mode `600`. | `deploy.yml:63-76` | Runner `ubuntu-latest` (GitHub-hosted, `deploy.yml:37`) berarti VPS harus dapat dijangkau SSH dari internet publik (rentang IP runner berubah). Kunci CI harus tanpa passphrase (`deploy.yml:76`), sehingga satu secret GitHub setara akses SSH penuh. |
| Deploy lewat SSH dari workstation operator | `BatchMode=yes`, multiplexing `ControlMaster` dengan `ControlPersist=120`, kanal kontrol di `/tmp`. Sebelum deploy memeriksa SHA di server dan melakukan *pre-flight*. | `backend/deploy-prod.sh:79-90`, `:579`, `:634` | Memakai kunci pribadi di laptop operator; kebijakan penyimpanan kunci belum ada. |
| Deploy dev via `appleboy/ssh-action` | Memakai secret `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_PORT` (default 22). | `.github/workflows/deploy-dev.yml:15-25` | Mekanisme verifikasi host key berbeda dari `deploy.yml`; target dev/prod di VPS yang sama **perlu verifikasi**. |
| Berkas rahasia tidak ikut git | `.env`, `.env.local`, `backend/tenants/*/.env` ada di `.gitignore`; tidak ada berkas `*.pem`/`id_rsa`/`id_ed25519` yang dilacak git. | `.gitignore:9-16`; pencarian `git ls-files` 31 Okt 2026 | Tidak menghalangi rahasia tersimpan di laptop; secret JWT pernah ter-commit (A.5.36 EX-09). |
| Hardening JWT saat start | Layanan menolak secret kosong, < 64 byte, atau yang dikenal publik. | `backend/*/pkg/config/jwt_secret.go` (A.5.36 bagian 2.2) | Rotasi di produksi berstatus "PREPARED ONLY" (`docs/jwt-rotation-runbook.md`). |
| Runbook deploy aman | Larangan `compose up` penuh/`down`/`prune`, tidak build Nuxt di VPS, gabungkan perintah dalam satu sesi SSH karena SSH bisa terblokir sementara bila koneksi terlalu cepat. | `docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md:26-33` | Menunjukkan ada mekanisme pembatas koneksi di host (jenis dan konfigurasinya **perlu verifikasi**; tidak ditemukan di repo). |

### 2.2 Kendali akses jarak jauh di aplikasi (untuk personel dan pengguna klien)

| Kendali | Implementasi | Bukti | Keterbatasan |
| :--- | :--- | :--- | :--- |
| MFA TOTP | Pengguna dapat mengaktifkan TOTP (`/mfa/status`, `/mfa/enroll`, dll.). Bila MFA aktif dan perangkat tidak tepercaya, login mengembalikan `MFARequired` dengan token MFA sementara di Redis. | `backend/auth-service/models/user.go:54`; `routes/registry.go:128-134`; `services/auth/auth_service.go:107-135`; UI `frontend/pages/settings/mfa.vue`, `frontend/pages/auth/mfa-login.vue` | MFA **tidak diwajibkan per role**, termasuk ADMIN (A.5.16 bagian 4 butir 5); ISO-AS-001 bagian 8 menyatakan wajib untuk ADMIN/CAE/Audit Manager. |
| Perangkat tepercaya | Pendaftaran lewat QR (`/devices/enroll`, `/devices/enroll/verify/{token}`), daftar dan hapus perangkat (`/devices`, `DELETE /devices/{id}`). Perangkat tepercaya yang belum kedaluwarsa melewati MFA. Kedaluwarsa dicek lewat `ExpiresAt`. | `backend/auth-service/controllers/trusted-devices/trusted_devices_controller.go:23-59`; `repositories/trusted_device_repository.go:88-91`; `auth_service.go:110-117`; `frontend/pages/settings/devices.vue` | Lama masa berlaku bawaan dan nilai `ExpiresAt` **perlu verifikasi** pada kode enrollment. Perangkat tepercaya mengurangi friksi MFA; perangkat yang hilang tetap tepercaya sampai dihapus. |
| Deteksi perangkat baru | Sidik jari perangkat dibandingkan dengan login terakhir; respons memuat `is_new_device`. | `auth_service.go:102-104`, `:130`, `:221`; `frontend/stores/auth.ts:339` | Hanya penanda di respons; tidak ada notifikasi email/alert ke pengguna (**perlu verifikasi** di UI). |
| Sesi dan token | JWT berlaku 24 jam (default `jwt.expiry_hour`), token juga disimpan di Redis `token:<jwt>` dan dihapus saat logout. | `backend/auth-service/pkg/config/config.go:176`; `auth_service.go:186-189`, `:392-395`; `pkg/middleware/auth.go:83-91` | Sesi tidak dicabut saat akun dinonaktifkan/dihapus; tidak ada idle timeout; tidak ada batas satu sesi aktif; lockout tidak berjalan (A.5.16 bagian 4 butir 1 dan 5). |
| Penyimpanan token di browser | Token disimpan di cookie `auth-token` lewat `useCookie`, `maxAge` 30 hari bila "ingat saya", selain itu 1 hari. | `frontend/stores/auth.ts:318`, `:341` | Cookie diset dari sisi klien, sehingga tidak bisa `HttpOnly`; atribut `secure`/`sameSite` tidak diatur eksplisit pada pemanggilan ini. Token dapat dicuri bila perangkat dikompromikan atau terkena XSS. `maxAge` cookie 30 hari melebihi masa JWT 24 jam. |
| Rate limit | Kong produksi: global 100/menit dan 5000/jam per IP; rute tertentu 60/menit dan 1000/jam. `bot-detection` aktif. | `backend/kong-gateway/kong/prod/kong.yml:80-86`, `:101`, `:457` | Per IP; pengguna di balik NAT bersama berbagi kuota. Tidak ada `ip-restriction` per tenant. |
| CORS | Kong produksi hanya mengizinkan origin `https://auditsphere.app` (konfigurasi global). | `kong.yml:31-34` | Origin domain tenant (`*.auditsphere.id`) pada konfigurasi tenant **perlu verifikasi**. |
| TLS di edge | Cloudflare/Nginx host dengan sertifikat Let's Encrypt (panduan provisioning tenant). | `docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md:34`, `:111-138` | Konfigurasi aktual di VPS **perlu verifikasi** (kewenangan `auditsphere-platform`). |
| Pakta Integritas | Pengguna wajib menyetujui sebelum memakai sistem (scroll >= 95%, dicatat dengan IP dan user agent). | `frontend/pages/auth/confidentiality.vue`; A.5.1 bagian 4 | Bukan kendali perangkat. |
| Monitoring | Prometheus, Grafana, Loki, Promtail; Promtail membaca log kontainer dan `/var/log/messages`. | `monitoring/docker-compose.monitoring.yml`; `monitoring/promtail` | Tidak ada alert untuk login SSH, login aplikasi gagal, atau perubahan perangkat tepercaya; log autentikasi SSH tidak terlihat di konfigurasi Promtail. |

### 2.3 Hubungan privat yang ada: hanya server-ke-server, bukan akses personel

Repo memuat WireGuard untuk menghubungkan server audit dengan Data Hub (`data-server/wireguard/wg0-server.conf`, `wg0-client.conf`, port 51820, subnet privat khusus). Ini mengamankan lalu lintas layanan ke Data Hub (A.5.20 dan arsitektur Data Hub), **bukan** akses personel ke VPS atau konsol. Pencarian `wireguard|tailscale|bastion|zerotier|cloudflare access|zero trust` di repo (di luar `.claude/worktrees` yang berisi salinan) hanya menemukan Data Hub, dokumen ZTAR-001, dan definisi agen. Tidak ditemukan VPN personel, bastion, `ip-restriction` Kong, atau konfigurasi `ufw`/`iptables`/`fail2ban`/`sshd_config` untuk VPS produksi (hanya contoh di panduan demo dan panduan dashboard).

### 2.4 Dokumen kebijakan yang menyinggung (belum berlaku formal)

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| CSP-001 Cloud Security Policy v1.0, bagian 10 dan 11 | SSO dan MFA wajib 100% untuk konsol cloud; tidak ada API key statis berumur panjang untuk service account; tinjauan akses triwulanan; akses administratif jaringan privat hanya lewat VPN gateway/PrivateLink. | Dokumen ada; bukti tanda tangan tidak dapat diverifikasi (A.5.36 EX-06). Praktik saat ini (kunci SSH statis tanpa passphrase di CI, akses `root`) tidak selaras. |
| ISO-AS-001 v0.1 DRAFT, bagian 10.2 | Segmentasi VLAN dengan VLAN 90 Management (SSH) "hanya via VPN Admin"; tabel A.8.20-A.8.22 berstatus *Implemented*. | Merupakan desain on-premises (Proxmox). Produksi berjalan di satu VPS dengan Docker Compose, jadi klaim *Implemented* tidak didukung bukti repo (tindakan 11). |
| ZTAR-001 Zero Trust Architecture Roadmap | Pilar Device: MDM, penilaian postur, integrasi EDR; Q3 "MDM & EDR"; perangkat tidak patuh diblokir. | Roadmap, belum dilaksanakan. |
| AIP-001 AI Usage Policy v1.0 | Tidak ada ketentuan perangkat atau kerja jarak jauh (pencarian teks 31 Okt 2026). | Perlu klausul agen AI di perangkat personel (Lampiran A bagian 6). |
| SCG-001 | Menyebut deploy key khusus dan token tidak di URL; bagian 11 mengusulkan `limit_req` Nginx untuk login. | Draft; deskripsi workflow deploy di dalamnya sudah usang (A.5.36). |
| A.5.16, A.5.8, A.5.24, A.5.36, A.6.3 | Gap akun, deploy, insiden, kepatuhan, dan pelatihan yang beririsan dengan kontrol ini. | Draf internal (rujukan bagian 4). |

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Remote work policy (disetujui) | **Belum tersedia** | Tidak ada di `docs/`. Draf RWP-001 pada Lampiran A. |
| VPN/MFA config: akses personel ke VPS dan konsol | **Belum tersedia** | Tidak ada VPN personel. MFA konsol cloud/GitHub **perlu verifikasi** (tangkapan layar pengaturan organisasi). Opsi pada Lampiran C. |
| VPN/MFA config: aplikasi | Tersedia (sebagian) | Kode MFA TOTP dan perangkat tepercaya (bagian 2.2). Tidak ada konfigurasi yang mewajibkan MFA per role. |
| Konfigurasi SSH/firewall VPS (`sshd_config`, `ufw`/`iptables`, daftar `authorized_keys` berupa fingerprint) | **Perlu verifikasi** | Di luar repo; diambil `auditsphere-devops`/`auditsphere-platform` sebagai tangkapan terbatas tanpa nilai kunci. |
| Endpoint report (inventaris dan kepatuhan perangkat) | **Belum tersedia** | Tidak ada MDM/inventaris. Templat Lampiran E dan F. |
| Bukti pelatihan kerja jarak jauh | **Belum tersedia** | Belum ada program pelatihan (A.6.3 bagian 3). Topik diusulkan sebagai T17 (Lampiran A bagian 9). |
| Acknowledgement personel atas kebijakan | **Belum tersedia** | Templat Lampiran H. |
| Tinjauan akses jarak jauh terisi | **Belum tersedia** | Templat Lampiran G; sebanding dengan tinjauan akses A.5.16 yang belum berjalan. |
| Catatan pelaporan perangkat hilang dan uji prosedur | **Belum tersedia** | Prosedur usulan di Lampiran D; register insiden belum ada (A.5.24). |
| Catatan pencabutan akses dan perangkat saat offboarding | **Belum tersedia / Perlu verifikasi** | Berkas SDM di luar repo; langkah di A.6.3 Lampiran H dan Lampiran D di sini. |
| Pengaturan keamanan organisasi GitHub (2FA wajib, *required reviewers*, daftar anggota, deploy key, PAT) | **Perlu verifikasi** | Di luar repo; tangkapan layar bertanggal. |

---

## 4. Catatan / Gap Terverifikasi

1. **Tidak ada kebijakan kerja jarak jauh, BYOD, atau baseline perangkat.** Pencarian `remote work|kerja jarak jauh|akses jarak jauh|BYOD|laptop|enkripsi disk|antivirus|MDM|screen lock` pada `docs/DOKUMEN_KESELARASAN_ISO27001.md`, seluruh `docs/ISO_27001/*.md`, dan teks CSP-001/AIP-001/SCG-001/ZTAR-001 tidak menemukan ketentuan operasional; satu-satunya rujukan perangkat adalah ZTAR-001 (roadmap) dan soal quiz A.6.3 G.1 no. 4 (laptop hilang).
2. **Kendali endpoint tidak terlihat dari repo.** Enkripsi disk, patch OS, antivirus/EDR, screen lock, manajer password, dan status perangkat kontraktor tidak dapat dibuktikan. Status: **Perlu verifikasi** dan, sampai ada laporan endpoint, diperlakukan sebagai gap.
3. **Tidak ada VPN, bastion, atau IP allowlist untuk personel.** Akses SSH dan konsol dilakukan langsung dari perangkat personel. WireGuard hanya untuk server-ke-Data Hub (bagian 2.3). CSP-001 bagian 11 butir 5 dan ISO-AS-001 "VPN Admin" belum terpenuhi.
4. **Akses privileged ke produksi memakai `root` dan kunci statis.** Runbook memakai `ssh root@<ip>`; CI memakai satu kunci tanpa passphrase di secret GitHub (`deploy.yml:76`); operator memakai kunci pribadi di laptop. Tidak ada akun bernama per orang, `sudo`, atau pencatatan siapa melakukan apa yang dapat dibuktikan dari repo. Tidak selaras dengan CSP-001 bagian 10 butir 3-5. Status `PermitRootLogin`/`PasswordAuthentication`: **Perlu verifikasi**.
5. **Port administratif dan layanan terbuka ke semua antarmuka.** `backend/docker-compose.prod.yml` memetakan `8001`, `8002`, `8003`, `8004`, `8000`, `8084`, Kong `8080`, `8443`, `8009` (admin API Kong, `KONG_ADMIN_LISTEN: 0.0.0.0:8001`), dan `8444` tanpa `127.0.0.1` (baris 94, 137, 169, 204, 229, 241, 271-275; admin listen sekitar baris 262). Postgres dan Redis tidak dipublikasikan. Stack monitoring juga mempublikasikan Prometheus `9090`, Grafana `9300`, node exporter `9100`, Loki `3100` (`monitoring/docker-compose.monitoring.yml:36-37, 61-62, 86-87, 105-106`). Admin API Kong OSS tidak punya autentikasi bawaan; mode DB-less membatasi perubahan tetapi tidak membatasi pembacaan konfigurasi. Apakah firewall host menutup port ini: **Perlu verifikasi** (A.5.36 bagian 4 butir 6). Ini adalah gap akses jarak jauh paling berisiko yang terlihat dari repo.
6. **Jalur deploy lama menonaktifkan pemeriksaan host key.** `backend/deploy-backend.sh:64`, `:68`, `:72` memakai `StrictHostKeyChecking=no`. Dipakai dari jaringan tidak tepercaya (Wi-Fi publik), ini membuka risiko man-in-the-middle terhadap sesi root. Jalur ini juga menjalankan seeder (A.5.36 EX-01). Skrip akar `deploy-prod.sh`/`deploy.sh` meneruskan ke jalur ini.
7. **Pengaturan GitHub tidak dapat diverifikasi.** *Required reviewers* untuk environment `production`, 2FA wajib organisasi, branch protection, dan daftar kolaborator tidak terlihat dari repo. Praktik tim adalah komit langsung ke `staging` tanpa PR (memori tim; A.5.36 bagian 2.2), sehingga satu akun GitHub yang diambil alih dapat memodifikasi kode dan memicu deploy (via `workflow_dispatch` bila akun punya hak).
8. **MFA dan sesi aplikasi belum cukup untuk akses jarak jauh.** MFA tidak diwajibkan (termasuk ADMIN); perangkat tepercaya melewati MFA tanpa masa berlaku yang terverifikasi; sesi tidak dicabut saat akun dinonaktifkan (24 jam); tidak ada idle timeout, lockout, maupun batas sesi; cookie token 30 hari dan tidak `HttpOnly` (bagian 2.2). Event login gagal belum terhubung ke handler (A.5.24 tindakan 6).
9. **Tidak ada kendali akses per tenant berbasis jaringan.** Tidak ada `ip-restriction` per tenant, daftar sesi aktif, atau riwayat login yang dapat dilihat pengguna/admin tenant (hanya penanda `is_new_device`). Tidak ada notifikasi email saat login dari perangkat baru (**perlu verifikasi**).
10. **Pemantauan tidak mencakup akses jarak jauh.** Alert yang ada hanya performa dan ketersediaan (A.5.36 bagian 2.2). Log autentikasi SSH tidak dikumpulkan Promtail (hanya log kontainer dan `/var/log/messages`). Tidak ada deteksi login SSH dari alamat baru atau di luar jam kerja.
11. **Klaim dokumen tidak selaras dengan produksi.** ISO-AS-001 bagian 10.2 (VLAN, VPN Admin) dan tabel A.8.20-A.8.22 berstatus *Implemented* untuk desain on-premises; produksi adalah satu VPS Docker Compose. Perlu dikoreksi agar tidak menyesatkan auditor (A.5.16 tindakan 9).
12. **Agen AI di perangkat personel.** Tidak ada ketentuan bahwa agen tidak boleh membaca `.env` atau memegang kredensial produksi, dan tidak ada konfigurasi pembatas yang dapat diverifikasi dari repo untuk personel lain. AIP-001 melarang kredensial ke AI publik, tetapi tidak mengatur akses berkas lokal oleh agen (A.6.3 Lampiran C M4 menyebut "agen tidak diberi kredensial produksi" sebagai materi, bukan kendali).
13. **MFA konsol layanan cloud tidak terbukti.** GitHub, Cloudflare, Rumahweb, Resend, ImageKit, Backblaze B2, dan alat AI: CSP-001 mewajibkan MFA 100%, tetapi tidak ada bukti. Kunci API yang dipakai sistem (Resend, ImageKit, B2) disimpan sebagai env di server; kunci yang sama mungkin ada di laptop personel (**Perlu verifikasi**).
14. **Tidak ada prosedur perangkat hilang atau offboarding perangkat.** Pengguna dapat menghapus perangkat tepercaya miliknya sendiri (`DELETE /devices/{id}`), tetapi tidak ada alur admin untuk mencabut perangkat/sesi pengguna lain, dan tidak ada langkah operasional yang tertulis.
15. **Lingkungan fisik tidak diatur.** Tidak ada ketentuan privasi layar, area kerja, perangkat bersama, atau penanganan cetakan dokumen klien. Aplikasi menyimpan unggahan di volume `audit_uploads` di server (bukan di laptop), dan `/uploads/*` publik tercatat sebagai gap (A.5.36 bagian 4 butir 7); namun salinan lokal (unduhan LHA/KKA, ekspor, backup di laptop) tidak dikendalikan.
16. **Status dokumen.** Seperti A.5.12 s/d A.6.3, dokumen ini belum ter-commit per 31 Oktober 2026. Tidak ada perubahan kode atau konfigurasi dilakukan dalam penyusunan dokumen ini.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Setujui dan tetapkan **Kebijakan Kerja Jarak Jauh (RWP-001)** berdasarkan Lampiran A: perangkat, koneksi, lingkungan, akses privileged, perangkat hilang, offboarding. Putuskan apakah BYOD diizinkan. | Lead Information Security dengan Legal & Compliance dan HR | Tinggi | [isi] | Kebijakan bertanda tangan |
| 2 | **Verifikasi dan tutup port administratif**: pastikan firewall host menutup `8009`, `8444` (admin Kong), port layanan `8000`-`8004`, `8084`, dan `9090`, `9100`, `9300`, `3100` dari internet; bila belum, tutup lewat firewall host atau binding `127.0.0.1` pada compose. Perubahan produksi hanya lewat `auditsphere-devops` dengan jalur deploy aman (`backend/deploy-prod.sh --ref`, tanpa seeder). | `auditsphere-platform` (firewall), `auditsphere-devops` (compose) | Tinggi | [isi] | Hasil pemindaian port dari luar (sebelum dan sesudah), tangkapan aturan firewall |
| 3 | Putuskan dan terapkan **model akses jarak jauh ke VPS dan konsol** (Lampiran C): VPN/mesh atau allowlist untuk SSH dan port admin; akses web admin (Grafana, Kong admin) di belakang lapisan autentikasi. | Lead Information Security, `auditsphere-devops`, `auditsphere-platform` | Tinggi | [isi] | Keputusan tertulis; konfigurasi (tanpa kunci); uji akses dari luar |
| 4 | **Akun privileged bernama**: akun SSH per orang dengan `sudo`, nonaktifkan login `root` dan password SSH setelah akun bernama berfungsi; kunci `ed25519` dengan passphrase atau kunci perangkat keras; kunci CI terpisah dari kunci personel dan dibatasi (pengguna deploy khusus, perintah/direktori terbatas). Verifikasi `sshd_config` aktual. | `auditsphere-devops` | Tinggi | [isi] | Tangkapan `sshd -T` terbatas, daftar fingerprint `authorized_keys`, catatan pemberian akses |
| 5 | **Aktifkan MFA** pada seluruh akun konsol: GitHub (2FA wajib organisasi), Cloudflare, Rumahweb, Resend, ImageKit, Backblaze B2, email kerja, dan alat AI. Simpan bukti pengaturan organisasi. | DevOps Lead, Platform | Tinggi | [isi] | Tangkapan layar bertanggal per layanan |
| 6 | **Periksa pengaturan GitHub**: *required reviewers* pada environment `production`, branch protection `main`, daftar kolaborator, deploy key, dan PAT; jadwalkan tinjauan triwulanan. | Project Lead, DevOps Lead | Tinggi | [isi] | Tangkapan layar pengaturan; Lampiran G terisi |
| 7 | **Hapus atau amankan jalur deploy lama** yang memakai `StrictHostKeyChecking=no` (`backend/deploy-backend.sh`, skrip akar) sejalan dengan tindakan A.5.36 untuk jalur deploy; sampai itu terjadi, larang pemakaiannya dari jaringan tidak tepercaya. | `auditsphere-devops` | Sedang | [isi] | Perubahan skrip atau catatan penghentian |
| 8 | Tetapkan **baseline perangkat** (enkripsi disk, patch, screen lock, antivirus/EDR bawaan, manajer password), buat **inventaris perangkat** dan **laporan endpoint** pertama (Lampiran E dan F); satu laporan terisi per personel. | Lead Information Security, Project Lead | Tinggi | [isi] | Inventaris dan laporan endpoint terisi |
| 9 | Putuskan **kendali endpoint yang layak untuk tim kecil**: swa-lapor dengan bukti tangkapan layar, atau MDM ringan (mis. fitur bawaan suite email/identitas perusahaan bila ada, **perlu verifikasi**). Bila BYOD diizinkan, wajibkan profil browser/akun kerja terpisah. **Keputusan manusia diperlukan.** | Lead Information Security, Direktur Utama | Sedang | [isi] | Keputusan; bila MDM dipilih, bukti pendaftaran perangkat |
| 10 | Susun **ketentuan agen AI pada perangkat personel**: agen tidak membaca `.env`/kunci produksi, tidak diberi kredensial produksi, dan izin agen dibatasi; selaraskan dengan AIP-001 dan A.5.32. | Lead Information Security, Security Officer | Sedang | [isi] | Klausul AIP-001/RWP-001, panduan konfigurasi |
| 11 | **Koreksi klaim dokumen**: ISO-AS-001 bagian 10.2 (VLAN, VPN Admin) dan tabel A.8.20-A.8.22 disesuaikan dengan produksi (VPS Docker Compose); catat CSP-001 bagian 10 dan 11 sebagai gap dengan tanggal target. | Lead Information Security | Sedang | [isi] | Revisi ISO-AS-001 |
| 12 | Usulkan **perbaikan aplikasi** terkait akses jarak jauh (tiket ke tim backend/frontend): wajibkan MFA per role (ADMIN dahulu), masa berlaku perangkat tepercaya, cabut sesi saat akun dinonaktifkan (A.5.16 tindakan 2), idle timeout, lockout, daftar sesi/login terakhir, notifikasi login dari perangkat baru, pertimbangkan cookie sesi `HttpOnly` lewat BFF, dan opsi IP allowlist per tenant. | Tim Backend dan Tim Frontend bersama `auditsphere-solution-architect` | Sedang | [isi] | Tiket, rancangan, perubahan kode dan uji (di luar dokumen ini) |
| 13 | **Pemantauan akses jarak jauh**: kumpulkan log autentikasi SSH ke Loki, alert untuk login SSH baru/gagal berulang dan login aplikasi gagal (hubungkan dengan A.5.24 tindakan 6 dan 8). | `auditsphere-platform` | Sedang | [isi] | Konfigurasi dan contoh alert |
| 14 | Terapkan **prosedur perangkat hilang** dan **offboarding perangkat** (Lampiran D); uji tabletop satu kali (A.6.3 T08). | Lead Information Security, DevOps Lead, Project Lead | Tinggi | [isi] | Prosedur disetujui, catatan uji |
| 15 | Sisipkan topik **T17 Kerja jarak jauh yang aman** ke training matrix dan laksanakan dalam induksi serta penyegaran tahunan (A.6.3). | Lead Information Security, Security Officer | Sedang | [isi] | Materi, daftar hadir, quiz |
| 16 | Petakan **tanggung jawab keamanan akses pengguna klien** dalam kontrak/DPA (MFA, penonaktifan pengguna keluar, perangkat yang dipakai) dan masukkan ke paket serah terima tenant (A.6.3 Lampiran A.5, F.3). | Legal & Compliance, Project Lead | Sedang | [isi] | Klausul kontrak, materi serah terima |
| 17 | Laksanakan **tinjauan akses jarak jauh triwulanan** pertama (Lampiran G) dan masukkan ringkasannya ke monitoring report A.5.36 Lampiran C. | Security Officer, DevOps Lead | Sedang | [isi] | Berita acara tinjauan |
| 18 | Commit dokumen ini bersama annex lain yang belum ter-commit (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |
| 19 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

Catatan pelaksanaan: tindakan 2, 3, 4, dan 7 menyentuh server produksi. Lakukan lewat `auditsphere-devops`/`auditsphere-platform`, satu perubahan per sesi, dengan sesi SSH cadangan terbuka sebelum mengubah `sshd` atau firewall agar tidak terkunci, dan dengan mempertimbangkan bahwa deploy GitHub Actions memerlukan SSH yang dapat dijangkau dari runner (Lampiran C). Jangan menjalankan seeder dan jangan `docker compose up` seluruh stack di VPS.

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A6.7-01 | Workflow deploy dengan host key dipin dan environment produksi | `.github/workflows/deploy.yml`, `.github/workflows/deploy-dev.yml` | Internal |
| EV-A6.7-02 | Skrip deploy dari workstation operator dan jalur lama (temuan `StrictHostKeyChecking=no`) | `backend/deploy-prod.sh`, `backend/deploy-backend.sh` | Internal |
| EV-A6.7-03 | Kode MFA TOTP, perangkat tepercaya, sesi dan token | `backend/auth-service/services/auth/auth_service.go`, `controllers/trusted-devices/`, `repositories/trusted_device_repository.go`, `frontend/pages/settings/mfa.vue`, `devices.vue`, `frontend/stores/auth.ts` | Internal |
| EV-A6.7-04 | Konfigurasi gerbang: rate limit, CORS, bot detection | `backend/kong-gateway/kong/prod/kong.yml` | Internal |
| EV-A6.7-05 | Pemetaan port produksi dan monitoring (temuan port terbuka) | `backend/docker-compose.prod.yml`, `monitoring/docker-compose.monitoring.yml` | Restricted |
| EV-A6.7-06 | Kebijakan terkait (CSP-001 bagian 10-11, ZTAR-001, ISO-AS-001 bagian 10.2) | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx`, `docs/ISO_27001/advanced_security/ZTAR-001_Zero_Trust_Architecture_Roadmap.docx`, `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| EV-A6.7-07 | Kebijakan Kerja Jarak Jauh RWP-001 disetujui (setelah tindakan 1); draf Lampiran A | `docs/ISO_27001/A.6.7_Kerja_Jarak_Jauh.md` (Lampiran A) | Internal |
| EV-A6.7-08 | Konfigurasi akses jarak jauh: aturan firewall, `sshd -T` terbatas, fingerprint `authorized_keys`, konfigurasi VPN/mesh, tangkapan MFA/2FA organisasi (setelah tindakan 2-6) | [belum ada] | Restricted |
| EV-A6.7-09 | Inventaris perangkat dan laporan endpoint per personel, acknowledgement (setelah tindakan 8, Lampiran E, F, H) | [belum ada] | Restricted (data per orang) |
| EV-A6.7-10 | Tinjauan akses jarak jauh triwulanan, catatan pelatihan T17, dan catatan uji perangkat hilang (setelah tindakan 14, 15, 17) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 9 selesai, tambahkan ID per siklus untuk setiap laporan endpoint dan tinjauan triwulanan.

---

## Lampiran A — Draf Kebijakan Kerja Jarak Jauh (RWP-001, usulan, siap tanda tangan)

Draf ini adalah masukan untuk tindakan 1 dan **belum disetujui**. Nilai pada kurung siku adalah usulan awal; ubah sebelum pengesahan.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | RWP-001 Kebijakan Kerja Jarak Jauh (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review tahunan dan setelah insiden) |
| Berlaku untuk | Seluruh karyawan, kontraktor, dan kontributor yang mengakses sistem, kode, data, atau layanan AuditSphere dari luar fasilitas yang dikelola perusahaan |

**1. Tujuan dan cakupan.** Melindungi informasi AuditSphere dan data klien saat pekerjaan dilakukan dari lokasi mana pun. Kebijakan mencakup perangkat, koneksi, dan lingkungan kerja, serta akses ke repo GitHub, VPS, Kong/monitoring, layanan cloud, dan alat AI. Akses pengguna klien diatur dalam kontrak/DPA, bukan dalam kebijakan ini.

**2. Perangkat**

| No | Ketentuan | Catatan |
| :--- | :--- | :--- |
| 2.1 | Pekerjaan memakai perangkat milik perusahaan **atau** perangkat pribadi (BYOD) yang memenuhi baseline pada Lampiran E. **Keputusan: BYOD [diizinkan / diizinkan terbatas / tidak diizinkan].** | Tim kecil: hampir pasti BYOD; bila diizinkan, perangkat wajib didaftarkan di inventaris. |
| 2.2 | Enkripsi seluruh disk aktif (FileVault, BitLocker, atau LUKS). | Bukti: tangkapan status. |
| 2.3 | OS dan peramban diperbarui; patch kritis dipasang dalam [7] hari dan lainnya dalam [30] hari sejak rilis. Hanya OS yang masih mendapat pembaruan keamanan vendor. | |
| 2.4 | Kunci layar otomatis maksimum [5] menit dengan kata sandi/biometrik; kunci layar manual saat meninggalkan perangkat. | |
| 2.5 | Antivirus/anti-malware bawaan OS atau EDR aktif dan terbarui; firewall perangkat aktif. | |
| 2.6 | Tanpa jailbreak/root; perangkat lunak dari sumber resmi; tidak ada perangkat lunak bajakan (A.5.32). | |
| 2.7 | Akun pengguna terpisah dari anggota keluarga/pihak lain; perangkat dengan akses privileged tidak dipakai bersama. Untuk BYOD, gunakan profil browser/akun OS khusus kerja. | |
| 2.8 | Manajer password wajib; tidak menyimpan kata sandi di berkas teks atau catatan. | |
| 2.9 | Perangkat hilang, dicuri, atau diduga terinfeksi dilaporkan sesuai Lampiran D. | |
| 2.10 | Perangkat lama/baru: data kerja dihapus aman sebelum dijual, diperbaiki pihak ketiga, atau dihentikan. | |

**3. Koneksi**

| No | Ketentuan | Catatan |
| :--- | :--- | :--- |
| 3.1 | Akses ke VPS, port administratif, dan panel admin hanya lewat jalur yang ditetapkan (VPN/mesh/allowlist, Lampiran C) setelah diterapkan; sebelum itu, hanya dari perangkat terdaftar. | Bergantung keputusan tindakan 3. |
| 3.2 | SSH memakai kunci `ed25519` (atau kunci perangkat keras) berpassphrase; login dengan password dan login `root` tidak dipakai (setelah tindakan 4). | |
| 3.3 | MFA wajib pada GitHub, email kerja, Cloudflare, Rumahweb, Resend, ImageKit, Backblaze B2, alat AI, dan akun lain yang menyimpan kode, kunci, atau data klien. | CSP-001 bagian 10 butir 2. |
| 3.4 | Wi-Fi publik: tidak untuk pekerjaan administratif atau akses data klien tanpa VPN; utamakan hotspot seluler pribadi atau jaringan rumah dengan sandi WPA2/WPA3 yang diganti dari bawaan. | |
| 3.5 | Tidak menonaktifkan pemeriksaan host key SSH; tidak menerima peringatan sertifikat/host key tanpa verifikasi. | Gap `StrictHostKeyChecking=no`. |
| 3.6 | Router rumah: kata sandi admin diganti, firmware dijaga, manajemen jarak jauh router dimatikan. | Anjuran. |
| 3.7 | Tidak memakai layanan berbagi layar atau akses desktop jarak jauh pihak ketiga ke perangkat kerja kecuali untuk dukungan yang disetujui. | |

**4. Lingkungan kerja**

| No | Ketentuan |
| :--- | :--- |
| 4.1 | Layar tidak dapat dilihat orang lain; gunakan filter privasi atau posisi yang sesuai di tempat umum. Tidak melakukan panggilan yang membahas data klien di tempat umum. |
| 4.2 | Dokumen kerja tidak ditinggalkan terbuka; perangkat dikunci dan disimpan aman saat tidak dipakai. |
| 4.3 | Data klien dan data produksi **tidak disimpan di perangkat lokal**, tidak disalin ke penyimpanan awan pribadi, USB, atau layanan berbagi berkas non-perusahaan, dan tidak dikirim ke alat AI publik (AIP-001). Pengembangan dan pengujian memakai data sintetis atau yang sudah dimasking. Unduhan LHA/KKA untuk keperluan kerja dihapus setelah selesai. |
| 4.4 | Tidak mencetak dokumen klien. Bila tidak dapat dihindari: cetak seperlunya, simpan terkunci, dan hancurkan (pencacah kertas) setelah tidak diperlukan. Printer bersama/publik tidak dipakai. |
| 4.5 | Perangkat tidak dipakai bersama anggota keluarga/teman untuk pekerjaan; tamu tidak dibiarkan tanpa pengawasan di dekat perangkat yang sedang terbuka. |
| 4.6 | Pesan dan rapat daring: gunakan akun kerja; jangan membagikan layar yang menampilkan kredensial, `.env`, atau data klien; rekaman tidak memuat data klien. |

**5. Akses privileged ke produksi dari jarak jauh**

| No | Ketentuan |
| :--- | :--- |
| 5.1 | Akses produksi (SSH, konsol cloud, admin Kong/Grafana) hanya untuk personel yang disetujui tertulis dan sudah lulus induksi (A.6.3 Lampiran A.1) dan modul deploy aman. |
| 5.2 | Akun bernama per orang; tidak berbagi akun atau kunci. Kunci CI dipisahkan dari kunci personel. |
| 5.3 | Perubahan produksi hanya lewat jalur resmi (`.github/workflows/deploy.yml` atau `backend/deploy-prod.sh --ref`, tanpa seeder) dan dicatat; tidak ada perubahan ad hoc tanpa persetujuan (A.5.8, A.5.36). |
| 5.4 | Akses privileged hanya dari perangkat yang memenuhi baseline penuh dan dari jaringan tepercaya (bukan Wi-Fi publik). |
| 5.5 | Hak akses ditinjau triwulanan (Lampiran G) dan dicabut saat peran berubah atau personel keluar. |

**6. Kunci, token, dan secret di perangkat personel**

| No | Ketentuan |
| :--- | :--- |
| 6.1 | Kunci pribadi SSH hanya di perangkat yang memenuhi baseline, tidak disalin ke penyimpanan awan/sinkronisasi berkas, tidak dikirim lewat chat/email, dan tidak dimasukkan ke repo. |
| 6.2 | Token GitHub (PAT) berhak minimum, berumur pendek, dan dicabut bila tidak diperlukan. Tidak ada token di URL, riwayat shell, atau berkas konfigurasi bersama. |
| 6.3 | Berkas `.env` hanya berisi nilai pengembangan; nilai produksi (JWT, kunci API Resend/ImageKit/B2, kredensial database) tidak disimpan di laptop kecuali untuk tugas yang disetujui dan dihapus setelahnya. Nilai tidak dicetak di terminal, log, tiket, atau chat. |
| 6.4 | Rahasia yang pernah terbuka (terkirim ke chat, ter-commit, terlihat pihak lain) dianggap bocor dan dirotasi lewat runbook (`docs/jwt-rotation-runbook.md`) dengan persetujuan. |
| 6.5 | Agen AI di perangkat tidak diberi kredensial produksi dan tidak diizinkan membaca `.env`/kunci; izin agen dibatasi ke direktori proyek. Hasil keluaran agen yang menyentuh produksi ditelaah manusia (AIP-001, A.5.32). |

**7. Pelaporan.** Perangkat hilang/dicuri/terinfeksi, kunci atau token terpapar, login mencurigakan, atau pelanggaran kebijakan dilaporkan ke Lead Information Security dalam [1] jam sejak diketahui (hari/jam kerja; di luar jam kerja segera setelah sadar). Pelaporan jujur tidak dihukum (*just culture*; A.6.3 Lampiran A.2).

**8. Penegakan dan pengecualian.** Pelanggaran ditangani sesuai kontrak dan kebijakan SDM (**di luar repo**) dan eskalasi A.5.36 Lampiran D.2. Pengecualian (mis. perangkat belum terenkripsi sementara) harus tertulis, bertenggat, dengan kontrol kompensasi, dan dicatat di exception register A.5.36.

**9. Pelatihan.** Topik T17 "Kerja jarak jauh yang aman" menjadi wajib untuk seluruh peran (R1) dan diperdalam untuk R2, R4, R9. Ditambahkan pada matriks A.6.3 Lampiran B:

| Kode | Topik | Materi sumber | Peran wajib | Frekuensi | Penilaian |
| :--- | :--- | :--- | :--- | :--- | :--- |
| T17 | Kerja jarak jauh yang aman: baseline perangkat, Wi-Fi publik, SSH/kunci/token, data klien tidak disimpan lokal, perangkat hilang | Dokumen ini, RWP-001 | R1 (W); R4, R9 (W, praktik: pendaftaran kunci dan uji lapor) | Induksi + tahunan | Quiz >= 80% + uji lapor |

**Blok persetujuan**

| Peran | Nama | Tanda tangan | Tanggal |
| :--- | :--- | :--- | :--- |
| Penyusun (Legal & Compliance / PM) | [isi] | | |
| Pemilik (Lead Information Security) | [isi] | | |
| Penyetuju (Direktur Utama) | [isi] | | |

---

## Lampiran B — Kontrol Akses Privileged ke Produksi (kondisi aktual vs usulan)

| Aspek | Kondisi aktual (bukti) | Usulan |
| :--- | :--- | :--- |
| Identitas SSH | Runbook memakai `root` (`docs/DEPLOY_RUNBOOK_EXECUTIVE_SUMMARY.md:33`); jumlah akun dan kunci di server **perlu verifikasi** | Akun bernama + `sudo`; nonaktifkan `root` login dan password SSH setelah uji |
| Kunci CI | Satu secret `VPS_SSH_KEY` tanpa passphrase (`deploy.yml:69-76`) | Kunci CI terpisah, pengguna deploy khusus, hak terbatas (hanya direktori/skrip deploy); rotasi berkala |
| Kunci personel | Kunci pribadi di laptop operator (`backend/deploy-prod.sh`); kebijakan tidak ada | `ed25519` berpassphrase atau kunci FIDO2; daftar fingerprint di Lampiran G; cabut saat keluar |
| Host key | Dipin di CI (`VPS_KNOWN_HOSTS`); jalur lama `StrictHostKeyChecking=no` | Semua jalur memakai host key dipin; hapus jalur lama |
| Jaringan | SSH diduga terbuka ke internet karena runner GitHub (**perlu verifikasi** firewall) | Lihat Lampiran C |
| Persetujuan | `environment: production`; *required reviewers* tidak terverifikasi | Aktifkan *required reviewers* (minimal 1 orang selain pemicu bila tim cukup) |
| Pencatatan | Log deploy di `/root/auditsphere-deploy/<timestamp>/` (A.5.36; **perlu verifikasi**); tidak ada log SSH terpusat | Kumpulkan log SSH ke Loki; catat siapa memicu deploy (GitHub Actions log + skrip) |
| Sesi | Multiplexing `ControlPersist=120` | Sesi singkat, tutup kontrol setelah deploy (sudah ada pembersihan di skrip) |
| Admin API Kong dan monitoring | Dipublikasikan di semua antarmuka (bagian 4 butir 5) | Bind `127.0.0.1` atau firewall; akses lewat VPN/mesh atau reverse proxy terautentikasi |
| Break-glass | Tidak ada | Akses darurat tertulis: siapa, kapan, dicatat dan ditinjau |

---

## Lampiran C — Opsi Konfigurasi VPN/Akses Jarak Jauh (kondisi aktual vs usulan)

**C.1 Kondisi aktual (terverifikasi dari repo; konfigurasi server perlu verifikasi).** Tidak ada VPN, bastion, `ip-restriction`, atau Cloudflare Access untuk personel. WireGuard hanya server-ke-Data Hub. Deploy GitHub Actions memakai runner publik, sehingga SSH ke VPS kemungkinan terbuka ke internet. MFA hanya ada di aplikasi dan bersifat opsional.

**C.2 Opsi dan tradeoff untuk tim kecil.** Biaya, batas gratis, dan fitur layanan pihak ketiga **perlu diverifikasi** langsung ke penyedia sebelum memutuskan; angka tidak dicantumkan di sini.

| Opsi | Cara kerja | Kelebihan | Kekurangan / biaya operasional | Dampak ke CI (GitHub Actions) |
| :--- | :--- | :--- | :--- | :--- |
| 1. IP allowlist di firewall host | SSH dan port admin hanya dari IP personel | Tanpa alat baru; cepat; murah | IP rumah/seluler berubah; admin harus memperbarui daftar; orang lepas pulang jaringan sulit; tidak ada autentikasi tambahan | **Memutus** `deploy.yml` bila runner tidak diizinkan; rentang IP runner GitHub besar dan berubah |
| 2. WireGuard self-hosted di VPS (atau VPS kecil kedua) | Peer per personel; SSH dan port admin hanya via alamat tunnel | Tim sudah punya pengalaman (Data Hub); tanpa vendor; ringan | Kelola kunci peer dan IP manual; tidak ada SSO/MFA bawaan; tidak ada pencatatan per pengguna yang ramah; satu titik kegagalan bila VPN di VPS yang sama | CI perlu peer sendiri atau jalur lain |
| 3. Tailscale (atau mesh sejenis) | Mesh WireGuard dengan identitas SSO + MFA, ACL per pengguna/tag, opsi SSH berbasis identitas | Cepat dipasang; identitas dan MFA terpusat; ACL dan log akses; cocok untuk tim terdistribusi; GitHub Actions dapat bergabung dengan kunci sementara | Bidang kontrol milik pihak ketiga (perlu penilaian pemasok dan DPA, A.5.20); biaya/batas paket **perlu verifikasi**; ketergantungan pada vendor | Mendukung node CI sementara; perlu secret OAuth/auth key yang dikelola |
| 4. Bastion/jump host | Satu host kecil sebagai satu-satunya titik masuk SSH | Satu titik kontrol dan log; ProxyJump sederhana | Tambahan VPS dan biaya; harus diamankan dan dipatch sendiri; tanpa VPN ia tetap terbuka ke internet (hanya dipersempit) | CI harus melalui bastion |
| 5. Cloudflare Access / Tunnel | Aplikasi web admin (Grafana, panel lain) di belakang SSO + MFA; Tunnel menghapus kebutuhan port masuk | Perusahaan sudah memakai Cloudflare (DNS); tanpa port terbuka; kebijakan per pengguna dan log | Utama untuk HTTP; SSH lewat Tunnel/`cloudflared` lebih rumit; batas paket **perlu verifikasi**; ketergantungan pada satu vendor DNS/edge | Dapat memakai service token untuk CI; perlu desain |
| 6. Status quo + hardening | SSH tetap publik, tetapi hanya kunci, tanpa password/root, `fail2ban`/pembatas, port admin ditutup | Tanpa komponen baru; menutup risiko terbesar dahulu | Tetap terbuka ke internet; bergantung disiplin kunci | Tidak ada perubahan |

**C.3 Rekomendasi bertahap (untuk keputusan manusia).**

1. **Langkah 0, segera (tanpa alat baru; tindakan 2, 4, 6):** tutup/binding port admin (Kong admin, Grafana, Prometheus, Loki, node exporter) dan port layanan; akun SSH bernama, tanpa `root`/password; kunci berpassphrase; aktifkan MFA konsol dan *required reviewers*. Ini menurunkan risiko terbesar dengan biaya terendah (Opsi 6).
2. **Langkah 1 (1-3 bulan):** pilih **mesh dengan identitas dan MFA (Opsi 3, Tailscale atau sejenis)** untuk SSH dan port admin bila penilaian pemasok dapat diterima; bila vendor tidak dapat diterima, pilih **WireGuard self-hosted (Opsi 2)** dengan prosedur peer tertulis. Alasan: dua pendekatan ini mengurangi permukaan SSH ke internet sambil tetap sesuai tim kecil; Tailscale mengurangi beban kunci dan memberi MFA, WireGuard menjaga kemandirian.
3. **Langkah 1b:** tempatkan **panel web admin di belakang Cloudflare Access (Opsi 5)** bila tidak ingin memasang VPN pada semua orang, karena Cloudflare sudah dipakai.
4. **CI:** setelah SSH tertutup dari internet, deploy GitHub Actions memerlukan node sementara di mesh (atau runner *self-hosted*). Rencanakan sebelum menutup port 22; uji di staging lebih dahulu. Jangan menutup SSH publik sebelum jalur CI pengganti siap.
5. **Hindari sekarang:** bastion terpisah (Opsi 4) dan allowlist murni (Opsi 1) sebagai satu-satunya kendali, karena keduanya menambah beban operasional atau merusak CI untuk tim kecil.

Keputusan ini memerlukan persetujuan Lead Information Security dan Direktur Utama, serta penilaian biaya oleh Platform. Setiap perubahan pada server melalui `auditsphere-devops`/`auditsphere-platform`.

**C.4 Konfigurasi MFA yang diusulkan.**

| Area | Usulan |
| :--- | :--- |
| Akun layanan cloud dan GitHub | MFA wajib (kunci keamanan atau TOTP); 2FA wajib tingkat organisasi GitHub; simpan bukti |
| SSH ke VPS | Kunci berpassphrase/FIDO2; bila memakai mesh, MFA di identitas mesh |
| Aplikasi (ADMIN, AUDITOR, EXECUTIVE) | Usul perubahan: wajibkan MFA per role (ADMIN dahulu); batasi masa berlaku perangkat tepercaya; notifikasi login dari perangkat baru (tindakan 12) |
| Pemulihan | Kode cadangan dan prosedur pemulihan MFA disimpan di manajer password; reset MFA oleh admin hanya setelah verifikasi identitas (A.6.3 M9) |

---

## Lampiran D — Perangkat Hilang, Insiden Perangkat, dan Offboarding Perangkat (usulan)

**D.1 Pelaporan dan respons perangkat hilang/dicuri/terinfeksi**

| Langkah | Penanggung jawab | Target usulan | Bukti |
| :--- | :--- | :--- | :--- |
| 1. Personel melapor ke Lead Information Security (kanal yang ditetapkan, bukan sekadar chat grup) | Personel | [1] jam sejak diketahui | Tiket/register insiden (A.5.24) |
| 2. Kunci layar jarak jauh/hapus jarak jauh bila tersedia (Find My, Intune, dsb.); ubah password akun perangkat | Personel dengan bantuan | Segera | Catatan |
| 3. Cabut akses: hapus kunci SSH perangkat dari `authorized_keys` (fingerprint di Lampiran G); cabut PAT GitHub dan sesi; keluarkan perangkat dari akun layanan cloud | DevOps Lead | [4] jam | Daftar pencabutan bertanggal |
| 4. Aplikasi: hapus perangkat tepercaya pengguna terkait (`DELETE /devices/{id}`; admin membantu bila alur admin belum ada); nonaktifkan atau reset password akun; cabut sesi (manual karena belum otomatis, A.5.16 butir 1) | DevOps Lead, Admin | [4] jam | Catatan |
| 5. Rotasi secret yang mungkin ada di perangkat (JWT, kunci Resend/ImageKit/B2, kredensial) lewat jalur aman dan runbook | DevOps Lead dengan `auditsphere-devops` | [24] jam | Bukti rotasi (digest, bukan nilai) |
| 6. Nilai dampak: apakah data klien/personal ada di perangkat? Bila ya, tangani sebagai insiden data (A.5.24; tenggat notifikasi CSP-001 butir 20: 1x24 jam ke CSP, 3x24 jam ke subjek data dan regulator, **tenggat resmi perlu verifikasi** Legal & Compliance) | Lead Information Security, Legal & Compliance | Segera | Penilaian dampak |
| 7. Pertahankan bukti (jangan hapus log/akun sebelum dicatat; A.5.28) | Lead Information Security | | Catatan chain of custody |
| 8. Tinjauan pasca-insiden dan pembaruan materi | Lead Information Security | [14] hari | Laporan |

**D.2 Offboarding perangkat dan akses jarak jauh** (melengkapi A.6.3 Lampiran H.2)

| Langkah | Penanggung jawab | Bukti |
| :--- | :--- | :--- |
| 1. Daftar perangkat dan akses yang dimiliki personel diambil dari inventaris (Lampiran F) dan Lampiran G | Project Lead | Daftar |
| 2. Cabut kunci SSH, peer VPN/mesh, PAT GitHub, keanggotaan organisasi dan layanan cloud, akses kanal komunikasi, dan akun aplikasi; hapus perangkat tepercaya terkait | DevOps Lead | Daftar pencabutan |
| 3. Rotasi secret yang pernah dipegang (tindakan 5 D.1) | DevOps Lead | Bukti rotasi |
| 4. Personel menghapus aman data kerja, repo lokal, `.env`, dan kunci dari perangkat pribadi dan mengembalikan perangkat perusahaan (bila ada); pernyataan tertulis | Personel, Lead Information Security | Pernyataan terisi |
| 5. Verifikasi akun layanan cloud tidak lagi memuat anggota tersebut | Security Officer | Tangkapan layar |
| 6. Perbarui inventaris dan training matrix | Project Lead | Entri |

---

## Lampiran E — Checklist Kepatuhan Perangkat (templat per perangkat)

Diisi personel (swa-lapor) saat induksi dan tiap [6] bulan, atau diverifikasi tim keamanan. Bukti berupa tangkapan layar bertanggal tanpa data sensitif.

| Kolom | Isi |
| :--- | :--- |
| ID perangkat / nama personel | |
| Kepemilikan | Perusahaan / BYOD / kontraktor |
| OS dan versi | |

| No | Pemeriksaan | Ya | Tidak | Bukti (nama berkas) | Catatan |
| :--- | :--- | :---: | :---: | :--- | :--- |
| 1 | Enkripsi seluruh disk aktif (FileVault/BitLocker/LUKS) | | | | |
| 2 | OS masih didukung dan patch terbaru terpasang (<= [7] hari untuk kritis) | | | | |
| 3 | Peramban dan perangkat lunak kerja terbarui | | | | |
| 4 | Kunci layar otomatis <= [5] menit dengan kata sandi/biometrik | | | | |
| 5 | Antivirus/anti-malware (bawaan atau EDR) aktif dan terbarui; firewall perangkat aktif | | | | |
| 6 | Manajer password dipakai; MFA aktif pada akun kerja | | | | |
| 7 | Kunci SSH berpassphrase atau kunci perangkat keras; tidak disinkronkan ke awan | | | | |
| 8 | Tidak ada `.env`/kunci produksi, data klien, atau dump basis data di perangkat | | | | |
| 9 | Tidak ada jailbreak/root atau perangkat lunak bajakan | | | | |
| 10 | Profil/akun kerja terpisah (BYOD) | | | | |
| 11 | Agen AI tidak memiliki akses ke kredensial produksi/`.env` | | | | |
| 12 | Fitur pelacakan/hapus jarak jauh perangkat aktif (bila ada) | | | | |
| 13 | Tidak ada perangkat lunak akses jarak jauh yang tidak disetujui | | | | |
| 14 | Sandi router rumah diganti dan manajemen jarak jauh router dimatikan | | | | |

Hasil: Patuh / Patuh dengan pengecualian (nomor pengecualian) / Tidak patuh. Tanda tangan personel dan pemeriksa, tanggal.

---

## Lampiran F — Templat Laporan Endpoint (inventaris dan kepatuhan)

**F.1 Inventaris perangkat**

| No | ID perangkat | Personel (peran R1-R12) | Kepemilikan | OS/versi | Akses sensitif (SSH prod, GitHub, konsol cloud, data klien) | Tanggal terdaftar | Tanggal cek terakhir | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | | | | | | | | |

**F.2 Laporan endpoint (triwulanan, bersama tinjauan akses)**

| Bagian | Isi |
| :--- | :--- |
| Identitas | Periode, tanggal terbit, penyusun, peninjau, versi, klasifikasi (Restricted untuk data per orang; ringkasan Internal) |
| 1. Ringkasan | Jumlah perangkat, jumlah yang patuh, jumlah pengecualian aktif, keputusan yang diminta |
| 2. Kepatuhan per kontrol | Persentase perangkat patuh untuk tiap butir Lampiran E (enkripsi, patch, kunci layar, antivirus, manajer password, MFA) |
| 3. Perangkat tidak patuh | Daftar tanpa nama pada versi manajemen; tindakan dan tenggat |
| 4. Pengecualian | Nomor pengecualian, alasan, kontrol kompensasi, tanggal kedaluwarsa (exception register A.5.36) |
| 5. Perangkat baru/keluar | Terdaftar, dikeluarkan, hilang/dicuri |
| 6. Insiden perangkat | Jumlah dan ringkasan (Lampiran D) |
| 7. Tren | Dibanding periode lalu |
| 8. Tindakan perbaikan | Daftar, PIC, tenggat, status |
| Lampiran | Evidence ID, tangkapan layar yang tersimpan terbatas |

**F.3 KPI usulan (target adalah usulan tanpa dasar historis; baseline 31 Okt 2026: tidak ada data)**

| No | KPI | Definisi | Target usulan | Baseline |
| :--- | :--- | :--- | :--- | :--- |
| K1 | Cakupan inventaris | Perangkat berakses sensitif yang terdaftar / seluruh perangkat berakses sensitif | 100% | Tidak ada data |
| K2 | Kepatuhan enkripsi disk | Perangkat terenkripsi / terdaftar | 100% | Perlu verifikasi |
| K3 | Kepatuhan patch | Perangkat patch terbaru dalam batas / terdaftar | >= 95% | Perlu verifikasi |
| K4 | MFA layanan cloud | Akun konsol dengan MFA / akun konsol | 100% | Perlu verifikasi |
| K5 | Akun privileged bernama | Akun SSH bernama / seluruh akun SSH (tanpa `root` bersama) | 100% | Belum |
| K6 | Port administratif tertutup dari internet | Port admin terbuka (hasil pemindaian luar) | 0 | Perlu verifikasi |
| K7 | Waktu pelaporan perangkat hilang | Median dari kejadian sampai laporan | <= 1 jam kerja | Tidak ada data |
| K8 | Penyelesaian offboarding perangkat | Offboarding dengan pencabutan tuntas <= [1] hari kerja / seluruh offboarding | 100% | Tidak ada data |

---

## Lampiran G — Tinjauan Akses Jarak Jauh (triwulanan, templat)

| Bagian | Isi pemeriksaan | Hasil | Tindakan |
| :--- | :--- | :--- | :--- |
| G.1 Akun dan kunci SSH | Daftar akun di VPS; fingerprint (bukan isi) kunci di `authorized_keys`; kunci tanpa pemilik atau lama tidak dipakai dicabut; kunci CI dirotasi sesuai jadwal | | |
| G.2 Konfigurasi SSH | `PermitRootLogin`, `PasswordAuthentication`, pembatas koneksi; hasil `sshd -T` terbatas | | |
| G.3 Pemindaian port dari luar | Pastikan port admin Kong, layanan, dan monitoring tidak terjangkau dari internet | | |
| G.4 GitHub | Anggota dan kolaborator, peran, 2FA wajib, deploy key, PAT aktif, *required reviewers*, branch protection, riwayat pemicu `workflow_dispatch` | | |
| G.5 Layanan cloud | Anggota dan peran di Cloudflare, Rumahweb, Resend, ImageKit, Backblaze B2, alat AI; status MFA; kunci API yang tidak terpakai | | |
| G.6 VPN/mesh | Peer atau pengguna aktif; peer tanpa pemilik; ACL | | |
| G.7 Aplikasi (personel dan tenant) | Daftar akun ADMIN dan role tinggi per tenant; status MFA; perangkat tepercaya yang tidak lazim; akun nonaktif yang sesinya belum kedaluwarsa | | |
| G.8 Endpoint | Ringkasan laporan endpoint (F.2); perangkat tidak patuh | | |
| G.9 Pelatihan | Personel yang jatuh tempo T17 | | |
| G.10 Insiden dan pengecualian | Insiden perangkat/akses periode ini; pengecualian aktif dan yang kedaluwarsa | | |
| G.11 Pemantauan | Alert akses jarak jauh yang terpicu dan penanganannya | | |

Penutup: pemeriksa, peninjau (orang berbeda bila memungkinkan), tanggal, daftar tindakan dengan PIC dan tenggat. Ringkasan masuk ke monitoring report A.5.36 Lampiran C dan didaftarkan di Evidence Index.

---

## Lampiran H — Acknowledgement Personel

Saya, [nama], peran [peran], status [karyawan/kontraktor/kontributor], menyatakan telah menerima, membaca, dan memahami Kebijakan Kerja Jarak Jauh (RWP-001) versi [isi] serta bersedia mematuhinya. Secara khusus saya menyatakan bahwa:

| No | Pernyataan | Setuju (paraf) |
| :--- | :--- | :--- |
| 1 | Perangkat yang saya pakai untuk pekerjaan telah memenuhi baseline (Lampiran E) dan terdaftar di inventaris. | |
| 2 | Saya mengaktifkan MFA pada seluruh akun kerja saya dan memakai manajer password. | |
| 3 | Saya tidak menyimpan data klien, data produksi, atau kunci/secret produksi di perangkat saya di luar yang disetujui, dan tidak mengirimnya ke alat AI publik. | |
| 4 | Saya menjaga kunci SSH dan token saya, tidak membagikannya, dan tidak memakai jaringan publik tanpa perlindungan untuk akses administratif. | |
| 5 | Saya akan melapor ke Lead Information Security dalam [1] jam bila perangkat atau kredensial saya hilang, dicuri, atau diduga terkompromi. | |
| 6 | Saat saya keluar atau berganti peran, saya mengembalikan atau menghapus aman data kerja dari perangkat saya dan menerima pencabutan akses. | |
| 7 | Saya memahami pelanggaran dapat dikenai sanksi sesuai kontrak dan kebijakan (AIP-001 bagian 5). | |

Nama, tanda tangan, tanggal. Penerima: Lead Information Security. Salinan disimpan terbatas dan dicatat pada training matrix (A.6.3 Lampiran B.4 dan F.2).

---

## Lampiran I — Penanggung Jawab (RACI usulan)

| Aktivitas | Direktur Utama | Lead Information Security | Legal & Compliance | HR | DevOps Lead | Platform | Security Officer | Project Lead | Tim Backend/Frontend |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Kebijakan RWP-001 dan keputusan BYOD | A (persetujuan) | R | C | C | C | I | C | C | I |
| Baseline perangkat dan inventaris | I | A | I | C | C | I | R | R | I |
| Laporan endpoint dan kepatuhan perangkat | I | A | I | I | C | I | R | C | I |
| Model akses jarak jauh (VPN/mesh/allowlist, Cloudflare Access) | A (persetujuan) | A/R | C | I | R | R | C | C | I |
| Penutupan port admin dan firewall host | I | A | I | I | R | R | C | I | I |
| Akun privileged bernama, kunci SSH, kunci CI | I | A | I | I | R | C | C | I | I |
| MFA konsol dan pengaturan GitHub | I | A | I | I | R | R | C | R | I |
| Perubahan aplikasi (MFA wajib, sesi, notifikasi) | I | C | I | I | C | I | C | A | R |
| Pemantauan akses jarak jauh | I | A | I | I | C | R | C | I | I |
| Perangkat hilang dan respons | I | A | R (dampak hukum/PDP) | I | R | C | R | I | I |
| Offboarding perangkat dan akses | I | A | R | R | R | C | C | R | I |
| Ketentuan agen AI di perangkat | I | A | C | I | C | I | R | C | C |
| Pelatihan T17 dan acknowledgement | I | A | C | C | C | I | R | R | I |
| Tinjauan akses jarak jauh triwulanan | I | A | I | I | R | C | R | C | I |
| Klausul kontrak/DPA akses pengguna klien | I | C | A/R | I | I | I | I | R | I |

R = Responsible, A = Accountable, C = Consulted, I = Informed. Pada tim kecil, satu orang dapat memegang beberapa peran; pemeriksa tinjauan akses sebaiknya bukan orang yang mengelola akses yang diperiksa bila memungkinkan.

---

Catatan sinkronisasi checklist: baris A.6.7 pada `docs/ISO_27001/03_Checklist_Kesiapan_Bukti_93_Kontrol_Annex_A.xlsx` (sheet "Checklist 93 Kontrol") belum diisi dan tidak diubah oleh dokumen ini. Nilai pada bagian "Ringkasan Nilai untuk Checklist" dapat disalin saat validasi.
