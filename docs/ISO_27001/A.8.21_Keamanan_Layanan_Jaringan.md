# A.8.21 — Keamanan Layanan Jaringan

| Field | Isi |
| :--- | :--- |
| **Tema** | Technological |
| **Judul kontrol** | Keamanan layanan jaringan |
| **Fokus pemeriksaan** | Persyaratan dan tingkat layanan keamanan jaringan ditetapkan |
| **Contoh bukti objektif** | Service catalogue, SLA, network security spec, monitoring |
| **Sistem** | AuditSphere (auditsphere.app), VPS produksi, Data Hub, stack tenant & Cloud SaaS |
| **PIC** | Lead Information Security (didukung DevOps Lead dan Legal & Compliance) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | 19 November 2026 (verifikasi repo cabang `staging`); validasi oleh reviewer [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi dan bukan nasihat hukum. Isi diverifikasi terhadap repositori (cabang `staging`) pada 19 November 2026; periksa ulang sebelum dipakai sebagai bukti. Kontrol ini menuntut persyaratan keamanan dan tingkat layanan jaringan **ditetapkan** (katalog layanan, SLA, spesifikasi keamanan, pemantauan). Repo hanya membuktikan **konfigurasi** (compose, Kong, skrip firewall, stack monitoring); repo **tidak** membuktikan keadaan server yang berjalan, konfigurasi nginx produksi, aturan firewall yang terpasang, konfigurasi DNS/Cloudflare, atau adanya kontrak/SLA dengan pemasok. Semua itu ditandai **Perlu verifikasi**. Butir yang bersumber dari catatan operasional (audit read-only VPS 5 Nov 2026, catatan platform 7 Okt 2026) dicantumkan sumbernya dan bukan bukti repo. Tidak ada SLA, angka, atau komitmen layanan yang ditemukan di repo; seluruh angka pada Lampiran adalah **usulan** dalam [kurung siku], bukan hasil pengukuran dan belum disetujui manajemen. Nilai rahasia, kunci, kata sandi, dan alamat IP tidak dicantumkan (beberapa berkas repo memuat nilai bawaan atau alamat; itu dicatat sebagai gap tanpa mengutip nilainya).

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere menjalankan dan memakai layanan jaringan: gateway Kong, reverse proxy nginx, TLS, DNS, firewall host, jaringan Docker per stack, VPN WireGuard ke Data Hub, dan akses SSH, serta bergantung pada penyedia VPS, DNS, email, dan backup pihak ketiga. Pengecualian tidak disarankan (bagian 1). |
| Implementasi status | **Sebagian** | Ada kendali teknis yang terlihat di repo: port host diikat ke loopback (`backend/docker-compose.prod.yml`), skrip firewall defence-in-depth, CORS terbatas, rate limiting, ClamAV tidak dipublikasikan, stack tenant dan training terpisah. Belum ada katalog layanan jaringan, SLA, spesifikasi keamanan jaringan resmi, atau pemantauan jaringan yang terbukti aktif (aturan alert tidak dimuat Prometheus di repo). |
| Ketersediaan bukti | **Sebagian (konfigurasi) / Tidak ada (kontraktual)** | Bukti konfigurasi tersedia di repo. Katalog layanan, SLA/kontrak pemasok, konfigurasi nginx produksi, bukti firewall terpasang, dan catatan pemantauan **belum ada**. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A8.21-01 s/d EV-A8.21-10) | Lihat bagian 6. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Belum teruji** | Tidak ada uji pemindaian port, uji penetrasi jaringan, atau catatan peninjauan firewall/pemantauan. Catatan 5 Nov 2026 menunjukkan IPv6 sempat melewati filter IPv4 (bagian 4 butir 2), sehingga efektivitas kendali belum dapat diklaim. |
| Desain kontrol | **Perlu perbaikan** | Kendali tersebar di compose, Kong, dan skrip tanpa spesifikasi tunggal; tidak ada persyaratan tingkat layanan; stack tenant berbagi jaringan Docker, Postgres, Redis, dan Kafka; Redis, Kafka, dan Postgres tanpa autentikasi/TLS jaringan di repo. Draf di Lampiran A sampai F. |
| Status kesiapan | **Belum siap** | Gate "persyaratan dan tingkat layanan ditetapkan" belum terpenuhi. |
| Prioritas | **Tinggi** | Klien lembaga jasa keuangan akan meminta bukti segmentasi, TLS, pembatasan akses administrasi, dan SLA. Gap eksposur port dan kredensial bawaan pada monitoring sudah tercatat (bagian 4). |
| Review berikutnya | [isi] | Usulan: tinjauan katalog dan spesifikasi setiap [6 bulan] dan setiap perubahan arsitektur jaringan, uji eksposur port setiap [kuartal], tinjauan pemasok setiap [tahun] (selaras A.5.22, draf). |

> Catatan pembacaan checklist: pada tab `Checklist 93 Kontrol`, baris A.8.21 berstatus "Belum ditentukan", PIC pada spreadsheet "Qiko", relevansi "Tentukan relevansi", penilaian "Belum dinilai", jadwal "Belum dijadwalkan". Dokumen ini memakai PIC berbasis peran (Lead Information Security, DevOps Lead) sebagai usulan yang perlu dikonfirmasi dan tidak mengubah spreadsheet.

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.8.21 **berlaku** karena AuditSphere mengoperasikan sendiri layanan jaringan inti dan memakai layanan jaringan pihak ketiga:

1. **Layanan jaringan yang dioperasikan sendiri.** Gateway Kong (`backend/docker-compose.prod.yml:270-310`), reverse proxy nginx di host (konfigurasi produksi tidak ada di repo; hanya contoh training di `backend/training/nginx-training.auditsphere.app.conf`), jaringan bridge Docker `rb_audit_network` (`:321-324`), firewall host (`backend/scripts/ops/auditsphere-firewall.sh`), VPN WireGuard ke Data Hub (`data-server/wireguard/`), akses SSH untuk deploy (`.github/workflows/deploy.yml`, `backend/deploy-prod.sh`), dan stack monitoring (`monitoring/`).
2. **Layanan jaringan pihak ketiga.** VPS dan konektivitas dari penyedia hosting, DNS di Rumahweb, Cloudflare (nameserver domain landing page; peran WAF/proxy untuk aplikasi belum terbukti), Resend untuk email transaksional (termasuk OTP login), Backblaze B2 untuk backup, GitHub untuk CI/CD. Detail pemasok ada di A.5.20 bagian 2.1; dokumen ini hanya menilai sisi jaringan dan tingkat layanan.
3. **Multi-tenant pada satu host.** Stack tenant (`backend/tenants/<slug>/`) berbagi bridge `rb_audit_network`, Postgres, Redis, Kafka, dan ClamAV dengan stack kontrol (`scripts/templates/docker-compose.tenant.yml.tpl:5-8`), sehingga segmentasi jaringan antar-tenant adalah persoalan A.8.21 (dan A.8.22, draf).
4. **Klaim dokumen internal melampaui bukti.** `docs/DOKUMEN_KESELARASAN_ISO27001.md:401` menyatakan seluruh komunikasi eksternal dan internal dilindungi TLS 1.3 dengan integrasi Cloudflare/AWS WAF; `:320` menyatakan A.8.24 *Implemented*. Repo menunjukkan koneksi Postgres `sslmode: disable` secara bawaan (`backend/docker-compose.prod.yml:97`), Kafka `PLAINTEXT` (`:63`), dan tidak ada konfigurasi Cloudflare/WAF. Klaim itu tidak boleh dipakai sebagai bukti A.8.21 (bagian 4 butir 9).
5. **Klien mengharapkan jawaban.** Due diligence lembaga jasa keuangan lazim menanyakan segmentasi jaringan, perlindungan port administrasi, TLS, dan komitmen ketersediaan. Saat ini tidak ada dokumen tunggal yang dapat diserahkan.

Pengecualian "tidak berlaku" tidak dapat dipertahankan. Bagian tanggung jawab penyedia (jaringan fisik dan backbone) dibagi lewat model tanggung jawab bersama dan dibuktikan lewat A.5.20 dan A.7.3.

---

## 2. Implementasi yang Sudah Ada

Semua butir berikut adalah **bukti konfigurasi di repo**. Apakah konfigurasi itu yang berjalan di VPS adalah **Perlu verifikasi** kecuali dinyatakan lain.

### 2.1 Inventaris layanan jaringan dan kendali keamanan yang terlihat

| Layanan jaringan | Kendali yang terlihat di repo | Bukti | Status |
| :--- | :--- | :--- | :--- |
| Kong gateway (produksi) | Mode DB-less dari `kong.yml`; proxy 8080 dan 8443 (`ssl http2`), admin 8001 dan 8444; host port diikat `127.0.0.1`; `tmpfs` untuk socket (mencegah crash loop setelah reboot tak bersih); healthcheck `kong health`; `client_max_body_size` 100m; `restart: unless-stopped`. | `backend/docker-compose.prod.yml:270-310` | Terverifikasi (konfigurasi). Port berjalan: **Perlu verifikasi**. |
| Kong routing | Route per layanan (auth, audit, master, risk, analytics, python-ai) dengan upstream dan healthcheck aktif/pasif; `/uploads` dan `/api/v1/media` melalui Kong. | `backend/kong-gateway/kong/prod/kong.yml:117-300`, `:310-660`, `:452` | Terverifikasi |
| Kong CORS | Origin tunggal `https://auditsphere.app`, `credentials: true`, daftar metode dan header eksplisit (tanpa `*`). Dev: `dev.auditsphere.app` dan `localhost`. | `kong.yml:31-63`; `backend/kong-gateway/kong/dev/kong-dev.yml:26-29` | Terverifikasi |
| Kong rate limiting dan bot detection | Global 100/menit dan 5000/jam per IP, `policy: local`; deny `sqlmap`, `nikto`, `nmap`; override per route (60/1000, 120/3600, 10/100). | `kong.yml:80-111`, `:466`, `:622`, `:658` | Terverifikasi |
| Kong dan JWT | Kong **tidak** memvalidasi JWT; hanya routing, CORS, rate limit, cache. Validasi JWT ada di tiap layanan dengan `JWT_SECRET` wajib dan bersama (`${JWT_SECRET:?...}`). Tenant memakai secret sendiri. | `kong.yml:481-485` (komentar); `backend/docker-compose.prod.yml:96`, `:121`, `:175`, `:202`, `:231`; `docs/jwt-rotation-runbook.md`; `scripts/templates/docker-compose.tenant.yml.tpl:38-42` | Terverifikasi |
| Kong metrik | Plugin `prometheus` (status code, latency, bandwidth, kesehatan upstream). | `kong.yml:23-29` | Terverifikasi |
| Eksposur port host (produksi) | Semua port terpublikasi diikat `127.0.0.1` (Kong, auth, audit, master, risk, analytics, python-ai). Postgres, Redis, Kafka, ZooKeeper, dan ClamAV **tidak** dipublikasikan ke host. Commit `46753fe`. | `backend/docker-compose.prod.yml:93-94`, `:140-141`, `:172-173`, `:218-219`, `:247-248`, `:259-260`, `:290-294`; layanan data tanpa blok `ports` (`:4-79`, `:149-158`) | Terverifikasi (konfigurasi). Sudah ter-deploy di VPS: **Perlu verifikasi**. |
| Firewall host | Skrip `auditsphere-firewall.sh`: DROP pada `DOCKER-USER` (IPv4 dan IPv6) dan `INPUT` IPv6 untuk 30 port internal pada antarmuka publik dan LAN penyedia; port 22, 80, 443 tidak pernah diblokir; idempoten, ada `--remove`, dijalankan saat boot lewat unit systemd (menurut komentar skrip). Antarmuka WireGuard sengaja tidak difilter. | `backend/scripts/ops/auditsphere-firewall.sh:1-98` | Terverifikasi (skrip). Terpasang dan aktif: **Perlu verifikasi**. |
| ClamAV | Tidak dipublikasikan ke host; hanya dapat dijangkau di jaringan Docker; image dipin dengan digest; `mem_limit`; upload fail-closed (`CLAMAV_MODE=required`). `clamd` mendengarkan TCP 3310 pada semua alamat di dalam container dan `::1`; clamd tidak punya autentikasi. | `backend/docker-compose.prod.yml:146-158`, `:182`, `:208`; `backend/clamav/clamd.conf` | Terverifikasi |
| Jaringan Docker | Satu bridge `rb_audit_network` untuk produksi; training memakai jaringan sendiri `rb_audit_training_network` dengan Kong admin di loopback; tenant memakai bridge produksi. | `backend/docker-compose.prod.yml:321-324`; `backend/training/docker-compose.training.yml:4-16`, `:198`, `:240-242`; `scripts/templates/docker-compose.tenant.yml.tpl:5-8`, `:291-293` | Terverifikasi |
| Isolasi tenant (jaringan/data) | Nama layanan bersufiks slug (alias jaringan unik), database per layanan, indeks Redis logis per tenant, `JWT_SECRET` per tenant, Kong dan frontend per tenant terikat loopback. Postgres, Redis, Kafka, dan python-ai dibagi. | `scripts/templates/docker-compose.tenant.yml.tpl:5-8`, `:22-47`, `:238-243`, `:265-266`; memori internal "tenant isolation gotchas" | Terverifikasi sebagian (lihat bagian 4 butir 5) |
| Reverse proxy host (nginx) | Contoh training: TLS ditambahkan certbot setelah penerbitan, proxy ke `127.0.0.1:3020`, `X-Robots-Tag`, `client_max_body_size 100m`. Konfigurasi produksi (`/etc/nginx/conf.d/auditsphere.conf`) **tidak ada di repo**. | `backend/training/nginx-training.auditsphere.app.conf:1-47` | Hanya contoh. Produksi: **Perlu verifikasi** |
| Frontend proxy | Nuxt `routeRules` meneruskan `/api/v1/**`, `/api/analytics/**`, `/api/python-ai/**`, dan `/uploads/**` ke target internal. | `frontend/nuxt.config.ts:72-87` | Terverifikasi |
| Data Hub (server terpisah) | nginx internal HTTP dengan Basic Auth untuk JupyterLab; API Data Hub dan AI Engine diakses lewat VPN; firewalld membuka 22/tcp dan 51820/udp, serta port 80, 8000, 8100 hanya untuk zona VPN. | `data-server/nginx/nginx.conf:1-72`; `data-server/setup_server.sh:96-110`; `data-server/wireguard/wg0-server.conf` | Terverifikasi (skrip). Aktif di server: **Perlu verifikasi** |
| Akses SSH untuk deploy | CI memasang kunci SSH dari secret dan **host key yang dipin** (`VPS_KNOWN_HOSTS`, tanpa `ssh-keyscan`); `deploy-prod.sh` memakai `BatchMode` (kunci, tanpa kata sandi interaktif). | `.github/workflows/deploy.yml:44-76`; `backend/deploy-prod.sh:79-90` | Terverifikasi (klien). Konfigurasi `sshd` server: **Perlu verifikasi** |
| Backup ke B2 | Skrip backup mengenkripsi dengan `age` sebelum ditulis (fail closed bila kunci tidak ada) dan menyalin ke remote `rclone` bila `RCLONE_REMOTE` diatur. Runbook masih berstatus DRAFT. | `backend/scripts/backup/backup-prod.sh:1-60`, `:120`; `docs/backup-b2-setup.md:1-5` | Skrip terverifikasi. Versi yang berjalan di VPS: lihat bagian 4 butir 8 |
| Pemantauan | Prometheus (retensi 7 hari dan 500 MB), Grafana, Node Exporter, Loki (retensi 7 hari), Promtail; semua dibatasi memori dan port diikat loopback. Prometheus menjangkau jaringan produksi dan dev. File aturan alert ada (Kong down, latensi, error rate, Postgres down). | `monitoring/docker-compose.monitoring.yml:19-148`; `monitoring/prometheus/prometheus.yml`; `monitoring/loki/loki-config.yml:36-47`; `monitoring/prometheus-dashboard-alerts.yml`; `monitoring/prometheus-fieldwork-alerts.yml` | Terverifikasi (konfigurasi). Efektivitas: lihat bagian 4 butir 4 |

### 2.2 Dokumen yang menyinggung jaringan (belum berlaku formal)

| Dokumen | Yang dinyatakan | Status |
| :--- | :--- | :--- |
| `docs/DOKUMEN_KESELARASAN_ISO27001.md` | TLS 1.3 internal dan eksternal, Cloudflare/AWS WAF, VLAN, desain on-premises Proxmox. | Desain/penawaran; tidak didukung konfigurasi produksi (bagian 4 butir 9). |
| `CSRM-001`, `CPA-001`, `CSP-001` (`docs/ISO_27001/cloud_saas/`, format docx) | Matriks tanggung jawab, kriteria penilaian penyedia (termasuk keamanan jaringan), kebijakan cloud. | Templat generik; belum disesuaikan dengan satu VPS (A.5.20 bagian 4 butir 2). Isi tidak diperiksa ulang pada dokumen ini. |
| `docs/NEW_CLIENT_DOMAIN_PROVISIONING_GUIDE.md` | Panduan domain, DNS, nginx, certbot untuk klien baru. | Panduan operasional; contoh memakai `sites-available`, sedangkan VPS memakai `conf.d` (catatan internal), jadi perlu disesuaikan. |
| `docs/jwt-rotation-runbook.md` | Rotasi `JWT_SECRET` lima layanan. | Disiapkan, belum dieksekusi (menurut dokumennya). |

### 2.3 Hal yang diverifikasi **tidak ada** di repo

Pencarian pada `backend/`, `scripts/`, `monitoring/`, `data-server/`, `docs/` (19 November 2026) tidak menemukan: katalog layanan jaringan; SLA atau kontrak penyedia VPS, DNS, email, atau backup; konfigurasi nginx produksi (TLS, protokol, cipher, header keamanan, HSTS); konfigurasi `sshd`, `fail2ban`, atau aturan `ufw`/`firewalld` host aplikasi; konfigurasi Cloudflare (proxy/WAF/Turnstile) untuk `auditsphere.app`; `rule_files` atau Alertmanager pada Prometheus; uji pemindaian port atau uji penetrasi jaringan; diagram jaringan; kebijakan layanan jaringan.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Service catalogue (katalog layanan jaringan) | **Belum tersedia; draf di Lampiran B** | Draf diisi dari fakta repo; pemilik, jam layanan, dan target masih [isi]. |
| SLA (internal dan dengan pemasok) | **Belum tersedia** | Tidak ada SLA di repo. Lampiran C menyediakan templat persyaratan; angka [usulan]. SLA penyedia VPS, Rumahweb, Cloudflare, Resend, dan Backblaze: **Perlu verifikasi** (dokumen publik pemasok atau kontrak). |
| Network security spec | **Draf di Lampiran D** | Berdasarkan konfigurasi repo; belum disetujui; bagian yang tidak terbukti ditandai. |
| Monitoring (bukti aktif) | **Sebagian** | Stack ada di repo; bukti alert dievaluasi, ada tinjauan, dan ada pemberitahuan **belum ada** (Lampiran E). |
| Konfigurasi nginx/TLS produksi | **Perlu verifikasi** | Ekspor read-only dari VPS oleh pihak berwenang (tanpa kunci privat); jangan disimpan bersama rahasia. |
| Bukti firewall terpasang dan efektif | **Perlu verifikasi** | Keluaran `iptables`/`ip6tables` dan uji eksternal (curl atau pemeriksaan protokol nyata; `nc -z` dari jaringan lokal dapat memberi hasil palsu menurut catatan internal 5 Nov 2026). |
| Diagram jaringan dan alur data | **Belum tersedia** | Lampiran D bagian D.1 memuat daftar zona; diagram dibuat bersama `auditsphere-solution-architect`. |
| Hasil pemindaian eksternal / uji penetrasi | **Belum tersedia** | Lihat tindakan 5. |
| Tinjauan akses SSH dan konfigurasi `sshd` | **Perlu verifikasi** | Catatan 5 Nov 2026: password dan root login diizinkan, tanpa `fail2ban`. Bukan bukti repo. |

---

## 4. Catatan / Gap Terverifikasi

Penanda: **[Repo]** terbukti dari berkas; **[Catatan]** dari catatan operasional internal (tanggal dicantumkan) dan **Perlu verifikasi**; **[Tidak ada]** tidak ditemukan di repo.

1. **[Tidak ada] Tidak ada katalog layanan, SLA, atau spesifikasi keamanan jaringan.** Tidak ada komitmen ketersediaan, RTO/RPO jaringan, jam dukungan, atau jalur eskalasi yang tertulis untuk layanan jaringan. Ini adalah inti kegagalan gate kontrol. Draf Lampiran A sampai F.
2. **[Catatan] Eksposur port: IPv6 melewati filter, dan perbaikan belum terbukti ter-deploy.** Audit read-only 5 Nov 2026 mencatat bahwa filter `DOCKER-USER` hanya menjangkau IPv4, sementara `docker-proxy` mendengarkan `[::]` sehingga Kong admin, layanan, Prometheus, Grafana, Loki, dan frontend dapat dijangkau lewat IPv6, serta antarmuka LAN penyedia dan `wg0` tidak difilter. **[Repo]** Perbaikan ada di commit `46753fe` (port diikat `127.0.0.1` di `backend/docker-compose.prod.yml`; skrip firewall mencakup `ip6tables`, port 8080, dan `eth1`). Apakah compose dan skrip itu sudah dipasang di VPS, dan hasil uji eksternal sesudahnya, **Perlu verifikasi**. Catatan: `backend/docker-compose.yml` (varian lokal/kontrol, berubah dan belum di-commit pada snapshot ini) masih memakai pola `"port:port"` tanpa loopback (misalnya `:277-279` untuk Kong); jangan dipakai di VPS tanpa pengikatan loopback.
3. **[Catatan]+[Repo] Akses SSH.** Catatan 5 Nov 2026: `sshd` mengizinkan login kata sandi dan root, tanpa `fail2ban`. **[Repo]** `backend/deploy-prod.sh` dan CI memakai akun `root` untuk deploy. Dengan host key yang dipin (CI) hal ini mengurangi risiko MITM, tetapi tidak ada bukti pembatasan sumber, MFA SSH, atau pembatasan percobaan login. **Perlu verifikasi** konfigurasi `sshd` saat ini.
4. **[Repo] Pemantauan belum menjamin deteksi gangguan jaringan.** (a) `monitoring/prometheus/prometheus.yml` tidak memiliki `rule_files` maupun `alerting`, dan compose tidak memasang berkas aturan, sehingga aturan alert di `prometheus-dashboard-alerts.yml` dan `prometheus-fieldwork-alerts.yml` tidak dievaluasi menurut repo; tidak ada Alertmanager atau kanal notifikasi. (b) Target scrape hanya Prometheus, Node Exporter, Kong prod dan dev; layanan Go, Postgres, Redis, Kafka, dan ClamAV tidak di-scrape. (c) Tidak ada probe ketersediaan dari luar (eksternal) dan tidak ada pemantauan masa berlaku sertifikat TLS. (d) Retensi metrik dan log hanya 7 hari, di bawah kebutuhan investigasi insiden dan bukti audit (selaras A.8.15 dan A.8.16, draf). (e) Image `prom/prometheus`, `grafana/grafana`, dan `prom/node-exporter` memakai tag `latest` (tidak dipin). (f) Promtail memasang `docker.sock` dan `/var/log` (akses luas; perlu dicatat dalam spesifikasi). Server menjalankan salinan di `/app/rbia-monitoring` (bukan checkout git, menurut catatan platform); isi sebenarnya **Perlu verifikasi**.
5. **[Repo] Segmentasi antar-tenant lemah pada lapisan jaringan.** Semua stack tenant berada di bridge yang sama dengan produksi dan berbagi Postgres, Redis, Kafka, ClamAV, dan python-ai (`scripts/templates/docker-compose.tenant.yml.tpl:5-8`). Admin API Kong (`0.0.0.0:8001` di dalam container, `docker-compose.prod.yml:280`) tidak berautentikasi dan dapat dijangkau dari container mana pun di bridge itu. Isolasi bergantung pada nama layanan, `JWT_SECRET`, `DATABASE_NAME`, dan indeks Redis yang benar, bukan pada pemisahan jaringan. Memori internal mencatat kasus kebocoran antar-tenant yang pernah ditemukan (URL master/auth default mengarah ke produksi; diperbaiki di working tree tetapi belum ter-deploy menurut catatan 22 Okt 2026). Training sudah memakai jaringan terpisah (preseden yang baik).
6. **[Repo] Lalu lintas internal tanpa enkripsi dan autentikasi jaringan.** Postgres `sslmode` bawaan `disable` (`docker-compose.prod.yml:97`, `:176`, `:206`, `:236`; `pkg/config/config.yaml` tiap layanan), Kafka `PLAINTEXT` pada kedua listener (`:63`), Redis tanpa `requirepass` pada compose dan healthcheck (`:19-31`), ClamAV tanpa autentikasi, dan Kong admin tanpa autentikasi. Risiko mengandalkan isolasi bridge (lihat butir 5). Data Hub: nginx hanya HTTP dengan Basic Auth dan API Data Hub/AI Engine diterbitkan pada semua antarmuka lewat compose (`data-server/docker-compose.yml:11-12`, `:91-92`, `:191-192`) dengan firewalld sebagai satu-satunya pembatas; interaksi Docker dan firewalld (Docker menulis aturan iptables sendiri) **Perlu verifikasi**.
7. **[Repo] Kredensial bawaan pada konfigurasi.** `monitoring/docker-compose.monitoring.yml:53-54` menetapkan admin Grafana dengan kata sandi bawaan (nilai tidak dikutip); catatan platform 7 Okt 2026 menyebut Grafana/Prometheus/Loki sempat terbuka di IP publik. Compose produksi memuat kata sandi bawaan Postgres (`:7-9`) dan nilai bawaan kunci API python-ai (`:123`) di dalam berkas (nilai tidak dikutip), dan Compose dev memuat bawaan JWT. Ini beririsan dengan A.8.5 dan A.8.9 (draf) tetapi mengurangi nilai segmentasi jaringan, sehingga dicatat di sini. Gap kredensial dikerjakan bersama tindakan 4 dan 6.
8. **[Catatan] Backup dan transfer ke B2.** `docs/backup-b2-setup.md` berstatus DRAFT ("belum terpasang"), padahal catatan 5 Nov 2026 menyebut cron 02:30 berjalan dan mengirim ke bucket privat B2; catatan 10 Nov 2026 menyebut skrip terpasang di VPS adalah salinan lama yang menulis berkas **tanpa enkripsi** dan `age` belum terpasang. **[Repo]** Skrip repo mengenkripsi dan fail closed. Perlindungan transit bergantung pada TLS bawaan `rclone`/B2 (**Perlu verifikasi**; tidak ada bukti repo). Backblaze tidak memiliki region Indonesia, sehingga klaim residensi tidak dapat dipakai. Detail residensi dan enkripsi berada di A.5.20 dan A.8.24 (draf).
9. **[Repo] Klaim dokumen tidak selaras dengan konfigurasi.** `DOKUMEN_KESELARASAN_ISO27001.md:401` (TLS 1.3 menyeluruh, Cloudflare/AWS WAF) dan `:405-420` (VLAN dan Proxmox) tidak sesuai dengan satu VPS dengan bridge Docker dan tanpa WAF yang terbukti. Tandai sebagai desain on-premises atau koreksi (tindakan 9).
10. **[Catatan] TLS dan protokol di edge belum terbukti.** Catatan platform 7 Okt 2026: nginx memakai certbot (Let's Encrypt), hanya HTTP/1.1 (HTTP/2 belum diaktifkan). Tidak ada bukti versi TLS minimum, cipher, HSTS, redirect 80 ke 443, atau pemantauan kedaluwarsa sertifikat. Kong 8443 `ssl http2` tidak memiliki sertifikat kustom yang dikonfigurasi di repo dan hanya berada di belakang nginx. **Perlu verifikasi** lewat ekspor konfigurasi nginx dan pemeriksaan eksternal (misalnya penguji TLS publik).
11. **[Catatan] Cloudflare/DNS.** Catatan platform: DNS `auditsphere.app` di Rumahweb (A record apex dan `api.` ke VPS); nameserver `auditsphere.id` dipindah ke Cloudflare (halaman landing, rencana Pages/Turnstile, belum terhubung). Jadi untuk aplikasi tidak ada proxy/WAF/DDoS Cloudflare yang terbukti; satu VPS menerima lalu lintas langsung. Perlindungan DDoS dan WAF belum ada kecuali penyedia VPS menyediakannya (**Perlu verifikasi**). Tidak ada konfigurasi DNSSEC, CAA, SPF/DKIM/DMARC yang terbukti di repo (catatan platform: SPF/DMARC belum ada pada 7 Okt 2026; status sekarang **Perlu verifikasi**).
12. **[Tidak ada] SLA pemasok layanan jaringan tidak ada.** Tidak ada SLA atau kontrak untuk penyedia VPS (nama dari whois: PT Fajar Informasi Globalnet Jaya, BOLEH-NET; belum dikonfirmasi kontrak, lihat A.7.3), Rumahweb, Cloudflare, Resend, Backblaze, dan GitHub di repo. Paket gratis Resend (kuota harian dibagi seluruh tenant menurut catatan platform) berdampak langsung pada ketersediaan OTP login. SLA publik pemasok perlu dikumpulkan; SLA yang dikutip harus bersumber dari dokumen pemasok, bukan dari repo ini.
13. **[Repo] Ketersediaan: satu VPS, satu Kong, satu instance tiap layanan data.** Healthcheck dan upstream round-robin Kong hanya menunjuk satu target per layanan; Kong `KONG_NGINX_WORKER_PROCESSES: 1`; tidak ada replika Postgres/Redis/Kafka (`KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1`, `docker-compose.prod.yml:65`). Kejadian 23 Okt 2026 (Kong crash loop setelah reboot tak bersih, API 502; catatan internal) menunjukkan titik kegagalan tunggal pada gateway. Ini menentukan angka SLA yang realistis; tidak ada SLA yang boleh dijanjikan di atas kemampuan arsitektur ini.
14. **[Repo] `/uploads` publik.** Rute `/uploads` ada di Kong (`kong.yml:452`) dan proxy Nuxt (`frontend/nuxt.config.ts:84-86`); catatan 8 Okt 2026 menyebut `/uploads/*` dapat dibaca tanpa login (`engine.Static` di audit-service) dan `media/download/:id` di luar autentikasi. Status saat ini **Perlu verifikasi**; dibahas pada A.5.36 dan akan menjadi syarat spesifikasi (Lampiran D butir D.4).
15. **[Repo] Stack training berbagi alias ClamAV produksi** (`backend/training/docker-compose.training.yml:15`), jadi ada jalur jaringan dari training ke layanan produksi bersama; lihat memori "Training environment". Perlu dicatat di katalog (Lampiran B).
16. **Status dokumen.** Seperti annex lain, dokumen ini belum ter-commit. Tidak ada perubahan kode, konfigurasi, server, atau spreadsheet yang dilakukan dalam penyusunan dokumen ini; tidak ada akses SSH ke server.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Verifikasi eksposur port dari luar** (IPv4 dan IPv6, kedua antarmuka) setelah deploy compose loopback dan skrip firewall; pakai pemeriksaan protokol nyata (curl), bukan `nc -z`. Perbaiki bila ada port terbuka. Perubahan produksi hanya lewat `auditsphere-devops`/`auditsphere-platform`, satu perubahan per sesi, atas konfirmasi eksplisit. | DevOps Lead | Tinggi | [isi] | Hasil uji eksternal bertanggal; keluaran aturan firewall (read-only) |
| 2 | **Setujui dan tetapkan NET-001** (Lampiran A) beserta katalog layanan jaringan (Lampiran B) dan spesifikasi keamanan (Lampiran D). Isi pemilik layanan, jam layanan, dan target dengan keputusan manusia. | Lead Information Security, Direktur Utama [konfirmasi jabatan] | Tinggi | [isi] | Kebijakan dan katalog disetujui |
| 3 | **Ekspor read-only konfigurasi nginx produksi** (TLS, protokol, redirect, header) dan pemeriksaan TLS eksternal; putuskan HTTP/2, TLS minimum, HSTS, dan simpan salinan sanitasi ke repo (atau docs) sebagai bukti. | DevOps Lead (`auditsphere-platform`) | Tinggi | [isi] | Berkas konfigurasi tersanitasi; laporan TLS |
| 4 | **Rapikan kredensial bawaan dan akses administrasi jaringan**: ganti kredensial Grafana, Postgres, dan kunci python-ai yang bernilai bawaan; kendalikan Kong admin; tentukan akses Grafana (hanya lewat tunnel/SSH atau reverse proxy terautentikasi). | DevOps Lead, Lead Information Security | Tinggi | [isi] | Perubahan terdokumentasi (tanpa nilai rahasia) |
| 5 | **Uji pemindaian port dan konfigurasi** secara terjadwal ([kuartalan] usulan) dan uji penetrasi eksternal minimal sekali sebelum klien jasa keuangan pertama (usulan; selaras A.8.8 dan A.8.29, draf). | Lead Information Security | Sedang | [isi] | Laporan pemindaian dan tindak lanjut |
| 6 | **Perkuat SSH**: pastikan hanya kunci, nonaktifkan login kata sandi, batasi root atau ganti dengan akun deploy, pertimbangkan `fail2ban` dan pembatasan sumber. Uji bahwa jalur deploy CI dan `deploy-prod.sh` tetap berfungsi sebelum mengubah. | DevOps Lead | Tinggi | [isi] | Konfigurasi `sshd` (tersanitasi), catatan perubahan |
| 7 | **Aktifkan pemantauan yang efektif**: muat `rule_files` ke Prometheus, tambahkan Alertmanager (atau pengganti) dengan kanal notifikasi, tambahkan scrape layanan dan probe ketersediaan eksternal, pantau masa berlaku TLS, pin versi image, dan tinjau retensi ([usulan] sesuai Lampiran E). Salinan di `/app/rbia-monitoring` harus sinkron dengan repo. | DevOps Lead (`auditsphere-platform`) | Tinggi | [isi] | Bukti alert uji, konfigurasi bertanggal |
| 8 | **Segmentasi jaringan**: nilai dan putuskan pemisahan jaringan/instans data per tenant (selaras A.8.22, draf); sebagai langkah minimal, batasi Kong admin ke loopback/jaringan monitoring dan pisahkan jaringan monitoring. Catat keputusan di spesifikasi. | `auditsphere-solution-architect`, DevOps Lead | Sedang | [isi] | Keputusan arsitektur, diagram |
| 9 | **Koreksi klaim dokumen** (`DOKUMEN_KESELARASAN_ISO27001.md:320`, `:401`, `:405-420`, dan ISO-AS-001): tandai sebagai desain on-premises atau sesuaikan dengan produksi; jangan klaim TLS internal dan WAF sebelum terbukti. | Lead Information Security, Legal & Compliance | Sedang | [isi] | Revisi dokumen |
| 10 | **Kumpulkan SLA/ketentuan layanan pemasok jaringan** (penyedia VPS, Rumahweb, Cloudflare bila dipakai, Resend, Backblaze, GitHub) dan catat di register pemasok; minta SLA tertulis dari penyedia VPS memakai Lampiran C/G; masukkan ke penilaian CPA-001 (A.5.20 tindakan 2 dan 3). | Legal & Compliance, DevOps Lead | Tinggi | [isi] | Register pemasok terisi, SLA/ketentuan tersimpan |
| 11 | **Putuskan SLA yang dapat dijanjikan ke klien** berdasarkan kemampuan arsitektur satu VPS (butir 13) dan SLA pemasok (tindakan 10); tulis di Lampiran C bagian C.1. Jangan menjanjikan di atas kemampuan terbukti. | Direktur Utama, Lead Information Security, DevOps Lead | Tinggi | [isi] | SLA internal disetujui; templat untuk klien |
| 12 | **Pasang backup terenkripsi** dan perbarui runbook (`docs/backup-b2-setup.md`) dari DRAFT; verifikasi transit dan pengendalian akses kunci B2 (selaras A.5.20 tindakan 9). | `auditsphere-devops`, Lead Information Security | Tinggi | [isi] | Log backup, runbook terbaru |
| 13 | **Tinjau Data Hub**: terapkan pengikatan antarmuka VPN pada compose atau firewall host, aktifkan TLS pada nginx Data Hub bila diakses di luar VPN, pastikan Basic Auth memakai kredensial kuat. | `auditsphere-data-analyst`, DevOps Lead | Sedang | [isi] | Konfigurasi dan hasil uji |
| 14 | **Daftarkan Evidence ID** pada Master Audit Evidence Index, dan jadwalkan peninjauan katalog/spesifikasi berkala. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |
| 15 | Commit dokumen ini bersama annex lain (jalur: komit langsung ke `staging`, merge origin lebih dahulu, tanpa force). | Project Lead | Sedang | [isi] | Riwayat git |

Catatan pelaksanaan: tindakan 1, 3, 4, 6, 7, 12, dan 13 dapat menyentuh produksi atau server lain; semuanya memerlukan konfirmasi eksplisit, dilakukan lewat agen devops/platform, satu perubahan per sesi, tanpa seeder dan tanpa `docker compose up` seluruh stack, serta tidak lewat `main` secara tak sengaja (push ke `main` memicu deploy pada workflow lama di `origin/main`; periksa).

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A8.21-01 | Compose produksi: port loopback, jaringan, Kong, ClamAV | `backend/docker-compose.prod.yml` | Internal |
| EV-A8.21-02 | Konfigurasi Kong: routing, CORS, rate limiting, bot detection (prod, dev, template tenant) | `backend/kong-gateway/kong/prod/kong.yml`, `backend/kong-gateway/kong/dev/kong-dev.yml`, `scripts/templates/kong-tenant.yml.tpl` | Internal |
| EV-A8.21-03 | Skrip firewall host dan bukti terpasang/aktif (keluaran read-only dari VPS, setelah tindakan 1) | `backend/scripts/ops/auditsphere-firewall.sh`; [belum ada keluaran VPS] | Restricted |
| EV-A8.21-04 | Konfigurasi nginx produksi tersanitasi dan laporan TLS eksternal (setelah tindakan 3) | [belum ada]; contoh training: `backend/training/nginx-training.auditsphere.app.conf` | Restricted |
| EV-A8.21-05 | Stack monitoring dan aturan alert; bukti alert dievaluasi dan uji notifikasi (setelah tindakan 7) | `monitoring/`; [belum ada bukti operasi] | Internal / Restricted |
| EV-A8.21-06 | Isolasi stack tenant dan training (templat compose, jaringan) | `scripts/templates/docker-compose.tenant.yml.tpl`, `backend/tenants/`, `backend/training/docker-compose.training.yml` | Internal |
| EV-A8.21-07 | Data Hub: nginx, WireGuard (konfigurasi memuat alamat dan kunci; akses terbatas), firewalld | `data-server/nginx/nginx.conf`, `data-server/setup_server.sh`, `data-server/wireguard/` | Restricted |
| EV-A8.21-08 | NET-001, katalog layanan, dan spesifikasi keamanan disetujui (setelah tindakan 2) | `docs/ISO_27001/A.8.21_Keamanan_Layanan_Jaringan.md` (Lampiran A, B, D) | Internal |
| EV-A8.21-09 | SLA/ketentuan layanan pemasok jaringan dan register pemasok (setelah tindakan 10) | [belum ada] | Restricted |
| EV-A8.21-10 | Laporan pemindaian port dan uji penetrasi eksternal (setelah tindakan 5) | [belum ada] | Restricted |

Setelah tindakan 1 sampai 7 selesai, tambahkan ID per siklus untuk setiap pemindaian, peninjauan firewall, dan tinjauan alert.

---

## Lampiran A — Draf Kebijakan Layanan Jaringan (NET-001, usulan)

Draf ini adalah masukan untuk tindakan 2 dan **belum disetujui**. Nilai pada kurung siku adalah usulan awal.

| Kolom dokumen | Isi |
| :--- | :--- |
| Kode / judul | NET-001 Kebijakan Keamanan Layanan Jaringan (usulan kode) |
| Versi / status | 0.1 Draf |
| Pemilik | Lead Information Security |
| Penyetuju | Direktur Utama [konfirmasi jabatan] |
| Tanggal berlaku / review | [isi] / [isi] (usulan: review [tahunan] dan setelah perubahan arsitektur jaringan atau insiden) |
| Berlaku untuk | Seluruh layanan jaringan yang dioperasikan AuditSphere (VPS produksi, training, stack tenant, Data Hub, monitoring) dan layanan jaringan pihak ketiga yang dipakai |

1. **Prinsip.** Semua layanan jaringan memiliki pemilik, persyaratan keamanan, dan tingkat layanan tertulis (Lampiran B dan C). Akses dari internet dibatasi pada port yang memang publik (80 dan 443 pada nginx host); semua port layanan lain hanya di loopback atau jaringan privat. Akses administrasi (SSH, Kong admin, Grafana, Prometheus) tidak terbuka ke publik.
2. **Segmentasi.** Lingkungan produksi, tenant, training, dev, monitoring, dan Data Hub dipisahkan sesuai Lampiran D bagian D.1; pengecualian (jaringan bersama) dicatat beserta kendali pengganti.
3. **Enkripsi in-transit.** Lalu lintas publik memakai TLS dengan versi minimum [TLS 1.2, usulan; keputusan setelah tindakan 3]. Lalu lintas internal yang berisi data klien diprioritaskan terenkripsi bila memungkinkan [keputusan; saat ini bergantung pada isolasi bridge].
4. **Pengelolaan perubahan.** Perubahan firewall, port, Kong, nginx, DNS, atau jaringan Docker dicatat (apa, kapan, siapa, persetujuan) dan diuji eksposurnya setelah perubahan (tindakan 1). Perubahan produksi hanya lewat jalur deploy yang disetujui.
5. **Pemasok jaringan.** Pemasok layanan jaringan (VPS, DNS, CDN/WAF, email, backup) dinilai dan SLA/ketentuannya dicatat di register pemasok (A.5.19 sampai A.5.22, draf; A.5.20 sudah ada). Penggantian penyedia mengikuti strategi keluar.
6. **Pemantauan dan insiden.** Layanan jaringan dipantau sesuai Lampiran E; pelanggaran ambang diperlakukan sebagai kejadian keamanan sesuai A.5.24 (dokumen sudah ada).
7. **Peninjauan.** Katalog layanan, spesifikasi, aturan firewall, dan daftar akun administrasi ditinjau setiap [6 bulan] dan hasilnya disimpan sebagai bukti.
8. **Pengecualian.** Pengecualian (misalnya port terbuka sementara) wajib tertulis dengan alasan, kendali pengganti, penyetuju, dan tanggal berakhir.

---

## Lampiran B — Katalog Layanan Jaringan (templat dengan isian faktual dari repo)

Kolom pemilik, jam layanan, dan target adalah [isi] atau usulan; fakta berasal dari repo dan bagian 2.

| ID (usulan) | Layanan | Deskripsi dan pengguna | Penyedia / pelaksana | Persyaratan keamanan utama | Pemilik layanan | Status bukti |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| NS-01 | Edge HTTPS publik (nginx host, certbot) | Terminasi TLS untuk `auditsphere.app` dan `api.`; meneruskan ke Kong dan frontend lewat loopback | AuditSphere (host) di VPS penyedia | TLS, redirect HTTP ke HTTPS [usulan], header keamanan [usulan], batas ukuran upload (repo: 100m), pemantauan sertifikat | DevOps Lead | Konfigurasi produksi: **Perlu verifikasi** |
| NS-02 | API gateway (Kong) | Routing ke lima layanan Go dan python-ai; CORS; rate limit; cache | AuditSphere (container) | Hanya loopback; origin CORS eksplisit; rate limit; admin tidak publik; JWT divalidasi oleh layanan | DevOps Lead | Terverifikasi (repo) |
| NS-03 | Jaringan layanan internal (`rb_audit_network`) | Komunikasi layanan dengan Postgres, Redis, Kafka, ClamAV | AuditSphere (Docker bridge) | Tidak ada port data di host; segmentasi tenant [belum]; autentikasi dan TLS internal [belum] | DevOps Lead, Solution Architect | Terverifikasi (repo), gap di bagian 4 |
| NS-04 | Jaringan tenant | Kong, layanan, dan frontend per tenant | AuditSphere (templat generator) | Secret dan database per tenant; port loopback; alias unik | DevOps Lead | Terverifikasi (templat) |
| NS-05 | Jaringan training | Salinan terisolasi produksi | AuditSphere | Jaringan terpisah; Kong admin loopback; alias ClamAV bersama dicatat | DevOps Lead | Terverifikasi (compose) |
| NS-06 | Firewall host | Memblokir port internal dari antarmuka publik dan LAN penyedia | AuditSphere (iptables/ip6tables) | Default-deny pada port internal; port 22, 80, 443 dikecualikan; diuji setiap perubahan | DevOps Lead | Skrip terverifikasi; terpasang: **Perlu verifikasi** |
| NS-07 | Akses administrasi (SSH) | Deploy dan operasi | AuditSphere | Kunci saja [usulan], host key dipin, tanpa login kata sandi [usulan], pembatasan percobaan [usulan] | DevOps Lead | Klien terverifikasi; server: **Perlu verifikasi** |
| NS-08 | VPN WireGuard ke Data Hub | Akses API Data Hub dan AI Engine | AuditSphere | Hanya peer terdaftar; kunci dikelola di luar repo; antarmuka VPN dicatat di firewall | DevOps Lead, Data Analyst | Skrip/konfigurasi terverifikasi; aktif: **Perlu verifikasi** |
| NS-09 | Pemindaian malware (ClamAV) | Pemindai setiap upload, fail-closed | AuditSphere | Tidak dipublikasikan; hanya jaringan internal; pembaruan signature | DevOps Lead | Terverifikasi (compose) |
| NS-10 | Pemantauan dan log | Metrik (Prometheus/Grafana), log (Loki/Promtail), host (Node Exporter) | AuditSphere | Akses terautentikasi, kredensial non-bawaan, port loopback, retensi sesuai kebijakan | DevOps Lead | Terverifikasi (konfigurasi); operasi: **Perlu verifikasi** |
| NS-11 | Konektivitas VPS dan jaringan fisik | Daya, backbone, bandwidth, alamat publik | Penyedia VPS [nama dikonfirmasi kontrak] | SLA tertulis, DDoS, kontak insiden | Legal & Compliance | **Perlu verifikasi** (tidak ada kontrak) |
| NS-12 | DNS | Resolusi domain aplikasi (Rumahweb) dan domain landing (Cloudflare) | Rumahweb, Cloudflare | MFA akun registrar, kunci domain, CAA/DNSSEC [usulan], catatan SPF/DKIM/DMARC | Lead Information Security | **Perlu verifikasi** |
| NS-13 | Email transaksional (Resend) | Notifikasi dan OTP login | Resend | Kunci khusus pengiriman per tenant, domain terverifikasi, kuota cukup untuk OTP | DevOps Lead | Repo (kunci per tenant); SLA: **Perlu verifikasi** |
| NS-14 | Transfer backup ke B2 | Salinan terenkripsi off-site | Backblaze | Enkripsi sebelum transfer, kunci aplikasi terbatas bucket, retensi | DevOps Lead | Skrip repo; terpasang: **Perlu verifikasi** |
| NS-15 | CI/CD ke VPS | Deploy via SSH dari GitHub Actions | GitHub, AuditSphere | Secret terproteksi, host key dipin, deploy manual (`workflow_dispatch`) | DevOps Lead | Terverifikasi |

---

## Lampiran C — Templat Persyaratan Tingkat Layanan (SLA/SLO) untuk Layanan Jaringan

Tidak ada SLA di repo. Semua angka adalah **[usulan]** yang harus diputuskan manusia dan tidak boleh dijanjikan ke klien sebelum selaras dengan kemampuan arsitektur (bagian 4 butir 13) dan SLA pemasok.

### C.1 Tingkat layanan internal (usulan)

| Layanan | Ukuran | Target [usulan] | Cara mengukur | Sumber data | Catatan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| NS-01/NS-02 Edge dan API | Ketersediaan bulanan | [isi] % | Probe eksternal | Belum ada (tindakan 7) | Satu VPS dan satu Kong; jangan melebihi SLA penyedia VPS |
| NS-02 API | Latensi p95 | [isi] ms | Metrik Kong | Plugin `prometheus` (ada di repo) | Alert latensi ada di file aturan, belum dimuat |
| Semua | Waktu respons insiden | [isi] menit (jam kerja) / [isi] (di luar jam kerja) | Tiket/alert | Belum ada | Selaras A.5.24 |
| Semua | Pemulihan setelah kegagalan | RTO [isi] jam / RPO [isi] jam | Uji pemulihan | `backend/scripts/backup/restore-test.sh` | Bergantung backup terenkripsi terpasang (tindakan 12) |
| Backup | Frekuensi dan retensi | Harian; retensi lokal 14 hari (nilai bawaan skrip), B2 [30 hari, rencana runbook] | Log backup | Skrip repo dan runbook | Nilai ini dari repo, bukan komitmen |
| Pemantauan | Retensi metrik/log | Saat ini 7 hari (repo); usulan [isi] hari | Konfigurasi | `monitoring/` | Lihat bagian 4 butir 4 |
| Sertifikat TLS | Peringatan kedaluwarsa | [isi] hari sebelum | Probe | Belum ada | |

### C.2 Persyaratan yang diminta dari pemasok layanan jaringan (daftar periksa)

| Butir | Penyedia VPS | DNS/registrar | CDN/WAF | Email (Resend) | Backup (B2) | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| SLA ketersediaan tertulis dan kompensasi | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Perlindungan DDoS | [isi] | n/a | [isi] | n/a | n/a | **Perlu verifikasi** |
| Jendela pemeliharaan dan pemberitahuan | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Kontak dan jalur eskalasi insiden | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Pemberitahuan insiden keamanan ke AuditSphere dalam [isi] jam | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Sertifikasi (ISO 27001/SOC 2) dan ruang lingkup | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Lokasi pemrosesan dan sub-prosesor | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Hak audit/permintaan bukti | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |
| Ketentuan penghentian layanan dan ekspor data | [isi] | [isi] | [isi] | [isi] | [isi] | **Perlu verifikasi** |

---

## Lampiran D — Spesifikasi Keamanan Jaringan (usulan)

Kolom "Terbukti di repo" memakai bagian 2; "Target" adalah persyaratan usulan yang dapat diubah.

### D.1 Zona dan segmentasi

| Zona | Isi | Kondisi di repo | Target [usulan] |
| :--- | :--- | :--- | :--- |
| Internet (publik) | nginx host (80, 443) | nginx produksi tidak ada di repo | Hanya 80 dan 443 terbuka; semua lain diblokir pada host dan penyedia |
| Produksi | Kong, lima layanan Go, python-ai, Postgres, Redis, Kafka, ClamAV | Satu bridge `rb_audit_network`; port host loopback | Port data tidak pernah dipublikasikan; autentikasi/TLS internal sesuai keputusan |
| Tenant | Kong, layanan, frontend per tenant | Berbagi bridge dan data layanan dengan produksi | Pertimbangkan jaringan per tenant atau pemisahan instans (tindakan 8) |
| Training | Klon produksi terisolasi | Jaringan sendiri; alias ClamAV bersama | Dokumentasikan pengecualian ClamAV |
| Dev | Stack dev (dihentikan menurut catatan 5 Nov 2026) | Jaringan dev terpisah; Prometheus terhubung ke jaringan dev | Pisahkan monitoring dari jaringan dev bila dev aktif |
| Monitoring | Prometheus, Grafana, Loki, Promtail, Node Exporter | Jaringan sendiri; Prometheus juga ke jaringan produksi dan dev | Akses hanya terautentikasi; tanpa kredensial bawaan |
| Data Hub | Server terpisah | WireGuard; firewalld zona VPN | API hanya melalui VPN; antarmuka publik hanya SSH dan port VPN |
| Administrasi | SSH, Kong admin | SSH ke akun root; Kong admin tidak diautentikasi di bridge | Lihat D.3 |

### D.2 Aturan port dan firewall (usulan)

| Aturan | Persyaratan |
| :--- | :--- |
| Default | Tolak semua masuk; izinkan 80 dan 443 (dan SSH sesuai D.3) |
| Port layanan internal | Hanya `127.0.0.1` atau jaringan Docker; filter `DOCKER-USER` dan `ip6tables` sebagai defence in depth (skrip ada) |
| Antarmuka | Perlakukan antarmuka publik dan LAN penyedia sebagai tidak tepercaya; keputusan untuk antarmuka VPN dicatat |
| IPv6 | Aturan berlaku sama dengan IPv4; diuji terpisah |
| Verifikasi | Pemindaian eksternal setelah setiap perubahan dan setiap [kuartal] |

### D.3 Akses administrasi

SSH hanya berbasis kunci [usulan]; login root dinonaktifkan atau dibatasi [usulan]; pembatasan percobaan login [usulan]; host key dipin pada klien deploy (sudah ada pada CI); daftar akun dan kunci ditinjau tiap [6 bulan]; akses ke Grafana, Prometheus, dan Kong admin hanya lewat SSH tunnel atau jalur terautentikasi.

### D.4 Aplikasi dan gateway

| Persyaratan | Kondisi di repo |
| :--- | :--- |
| CORS hanya origin yang sah, tanpa `*` | Terpenuhi pada Kong prod dan dev |
| Rate limiting global dan per route sensitif | Ada (nilai pada bagian 2.1) |
| Batas ukuran upload | 100m pada Kong dan nginx training |
| Endpoint file publik (`/uploads`) | Terbuka menurut catatan 8 Okt 2026; keputusan akses dan pemindaian diperlukan (A.5.36) |
| JWT divalidasi pada tiap layanan dengan secret bersama per stack | Ada; Kong tidak memvalidasi, jadi cache Kong menyertakan `Authorization` pada kunci |
| Secret berbeda per tenant | Ada pada templat |

### D.5 Enkripsi dan trust internal

TLS publik (D.1). Untuk internal: Postgres `sslmode` bawaan `disable`, Kafka PLAINTEXT, Redis tanpa autentikasi, ClamAV tanpa autentikasi. Keputusan: terima risiko dengan kendali pengganti (isolasi) atau terapkan autentikasi/TLS bertahap [prioritaskan Postgres dan Redis]. Catat di SoA.

---

## Lampiran E — Persyaratan Pemantauan Jaringan (usulan)

| Aspek | Kondisi di repo | Persyaratan usulan |
| :--- | :--- | :--- |
| Ketersediaan layanan | Aturan alert ada (Kong down, layanan down, Postgres down), tetapi tidak dimuat Prometheus | Muat aturan; notifikasi ke [kanal]; uji alert [kuartalan] |
| Probe eksternal | Tidak ada | Probe HTTPS dari luar VPS untuk edge dan API |
| Latensi dan error | Metrik Kong (status, latensi, bandwidth) | Ambang p95 dan rasio error [isi] |
| Host | Node Exporter | Alert disk, memori, swap [isi] |
| Sertifikat TLS | Tidak ada | Alert [isi] hari sebelum kedaluwarsa |
| Log | Loki/Promtail, retensi 7 hari | Retensi log keamanan [isi] hari; akses dibatasi (A.8.15, A.8.16 draf) |
| Deteksi anomali jaringan (pemindaian port, brute force) | Bot detection Kong; tidak ada IDS/fail2ban di repo | Pertimbangkan pemantauan percobaan login SSH dan lonjakan 401/429 |
| Tinjauan | Tidak ada | Tinjauan dasbor dan alert [mingguan/bulanan] dengan catatan |
| Kapasitas | Batas memori kontainer ditetapkan; Loki pernah OOM (catatan di compose) | Tinjauan kapasitas [kuartalan] |
| Keamanan stack monitoring | Kredensial bawaan Grafana; port loopback | Ganti kredensial; akses terautentikasi |

---

## Lampiran F — Penanggung Jawab (RACI usulan)

R = Responsible, A = Accountable, C = Consulted, I = Informed.

| Aktivitas | Lead Information Security | DevOps Lead | Legal & Compliance | Solution Architect | Direktur Utama | Data Analyst (Data Hub) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Menetapkan NET-001 | A/R | C | C | C | A (persetujuan) | I |
| Katalog layanan jaringan | A | R | I | C | I | C |
| Spesifikasi keamanan jaringan | A | R | I | R | I | C |
| Firewall, port, SSH | C | A/R | I | C | I | C |
| Konfigurasi nginx/TLS/DNS | C | A/R | I | I | I | I |
| Kong, CORS, rate limit | C | A/R | I | C | I | I |
| Segmentasi tenant | A | R | I | R | I | I |
| Pemantauan dan alert | C | A/R | I | I | I | I |
| SLA internal | R | R | C | C | A | I |
| SLA pemasok dan register pemasok | C | C | A/R | I | I | I |
| Pemindaian dan uji penetrasi | A/R | C | I | I | I | I |
| Evidence ID dan review berkala | A/R | C | C | I | I | I |

---

## Lampiran G — Permintaan ke Pemasok Layanan Jaringan (ringkas)

Gunakan bersama kuesioner penyedia di A.7.3 Lampiran F dan A.5.20; simpan jawaban sebagai EV-A8.21-09.

1. Dokumen SLA ketersediaan jaringan dan kompensasi pelanggaran untuk produk yang dipakai (VPS, DNS, email, backup).
2. Perlindungan DDoS dan filter jaringan yang disediakan, serta apakah port tertentu diblokir di sisi penyedia.
3. Pengaturan IPv6, antarmuka LAN privat, dan firewall di tingkat penyedia (apakah tersedia dan siapa mengelola).
4. Kontak dan jalur eskalasi insiden, jendela pemeliharaan, dan cara pemberitahuan gangguan.
5. Sertifikasi (ISO 27001/SOC 2), lokasi data center, dan sub-prosesor.
6. Ketentuan data saat penghentian layanan.
7. Untuk Resend: batas kuota dan paket yang dipakai terhadap kebutuhan OTP login seluruh tenant.
8. Untuk Backblaze: enkripsi di transit dan simpanan, region bucket, kebijakan kunci aplikasi.
9. Untuk Cloudflare (bila dipakai untuk aplikasi): mode proxy, aturan WAF, dan pengaturan TLS; saat ini belum terbukti untuk `auditsphere.app`.

---

## Referensi Silang Annex

| Annex | Status dokumen | Catatan |
| :--- | :--- | :--- |
| A.5.20 Keamanan dalam perjanjian pemasok | Ada (`docs/ISO_27001/A.5.20_Keamanan_Dalam_Perjanjian_Pemasok.md`) | Inventaris pemasok dan gap kontrak |
| A.7.3 Pengamanan kantor, ruangan, dan fasilitas | Ada | Kendali fisik penyedia dan Data Hub |
| A.6.7 Kerja jarak jauh | Ada | Akses jarak jauh personel ke VPS |
| A.5.24 Persiapan manajemen insiden | Ada | Respons terhadap kejadian jaringan |
| A.5.36 Kepatuhan terhadap kebijakan dan standar | Ada | `/uploads` publik dan klaim dokumen |
| A.5.19, A.5.21, A.5.22, A.5.23 | **Draf, dokumen belum ada** | Hubungan pemasok, rantai pasok, pemantauan layanan pemasok, layanan cloud |
| A.8.16 Aktivitas pemantauan | **Draf, dokumen belum ada** | Pemantauan dan deteksi anomali |
| A.8.20 Keamanan jaringan | **Draf, dokumen belum ada** | Kendali jaringan (firewall, segmentasi); dokumen ini berfokus pada layanan jaringan |
| A.8.22 Segregasi jaringan | **Draf, dokumen belum ada** | Segmentasi antar-tenant |
| A.8.23 Penyaringan web | **Draf, dokumen belum ada** | Tidak dibahas di sini |
