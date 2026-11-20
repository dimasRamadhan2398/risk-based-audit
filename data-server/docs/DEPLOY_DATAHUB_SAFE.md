# Runbook: Deploy Ulang Data Hub (API + Worker) TANPA Menimpa Data

Berlaku untuk server Data Hub (AlmaLinux, `187.127.122.82`, VPN `10.0.0.1`, hanya dapat diakses lewat WireGuard).
Tujuan: membangun ulang image `data-hub-api` dan `datahub-worker`, lalu me-recreate HANYA kedua service itu.
Data lake (TimescaleDB: bronze/silver/gold, skema `ops`, `bronze.registered_sources`) tidak disentuh.

Semua perintah dijalankan sebagai root di server Data Hub kecuali disebut lain.

---

## 0. Di mana data tinggal (baca dulu)

| Data | Lokasi | Pemilik |
|---|---|---|
| Data lake (bronze/silver/gold/ops, sumber terdaftar) | named volume `<project>_datalake_data` -> `/var/lib/postgresql/data` | service `postgres-datalake` (container `auditsphere_datalake`) |
| CBS simulator | named volume `<project>_cbs_data` | `postgres-cbs-sim` |
| Scratch worker/API | named volume `<project>_hub_data` -> `/app/data` | `data-hub-api`, `datahub-worker` |
| Model AI / Jupyter | `<project>_ai_models`, `<project>_jupyter_data` | `ai-engine`, `caatt-engine` |

`<project>` = nama direktori compose (umumnya `auditsphere-data` bila di `/opt/auditsphere-data`). Verifikasi di langkah 1.

Hal yang MENGHAPUS data (jangan dijalankan pada prosedur ini):

1. `docker compose down -v` / `docker volume rm` / `docker volume prune` / `docker system prune --volumes`.
2. Menjalankan compose dari direktori dengan nama berbeda (atau `-p` berbeda): compose membuat volume BARU yang kosong, dan init script `init-datalake.sql` jalan lagi di database kosong itu. Data lama tetap ada di volume lama tetapi aplikasi "kehilangan" data. Selalu jalankan dari direktori yang sama dengan sebelumnya.
3. `setup_server.sh` dijalankan ulang: tidak menghapus named volume, tetapi membuat `volumes/*` (bind mount yang TIDAK dipakai compose; jangan percaya direktori itu berisi data) dan mengubah firewall. Tidak perlu dijalankan untuk deploy.
4. `rsync --delete`, `git clean -fdx`, atau `rm -rf` pada direktori compose: menghapus `.env`, `nginx/htpasswd`, `wireguard/*.key` (semuanya di-gitignore dan tidak ada di repo). Data lake aman (named volume), tetapi stack tidak bisa naik dan kunci WireGuard hilang.
5. `docker compose up -d --force-recreate` TANPA nama service, atau tanpa `--no-deps`: ikut me-recreate `postgres-datalake` (data tetap, tetapi DB restart dan downtime). Selalu sebut nama service dan pakai `--no-deps`.
6. Overlay `docker-compose.prod.yml` saat ini tidak valid: `deploy.replicas: 4` untuk `datahub-worker` bentrok dengan `container_name` (`can't set container_name and datahub-worker as container name must be unique`). Cek overlay mana yang dipakai di server (langkah 1) dan jangan "memperbaiki" dengan menghapus container DB.

---

## 1. Pre-check (read-only)

```bash
cd /opt/auditsphere-data            # direktori compose yang dipakai saat ini (cek: docker inspect di bawah)

# 1a. Container, project, dan overlay yang sedang berjalan
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
docker inspect auditsphere_datalake --format '{{index .Config.Labels "com.docker.compose.project"}} | {{index .Config.Labels "com.docker.compose.project.working_dir"}} | {{index .Config.Labels "com.docker.compose.project.config_files"}}'

# 1b. Pastikan data ada di named volume dan catat namanya
docker inspect auditsphere_datalake --format '{{range .Mounts}}{{.Type}} {{.Name}} -> {{.Destination}}{{"\n"}}{{end}}' | grep postgresql/data
docker volume ls | grep -E 'datalake_data|hub_data'

# 1c. Ruang disk dan ukuran data (menentukan strategi backup di langkah 2)
df -h /var/lib/docker /root
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select pg_size_pretty(pg_database_size(current_database()))"

# 1d. Migrasi yang sudah tercatat. Harus berisi 001..008 (bila kosong/tidak ada: BERHENTI, lihat catatan)
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select version from ops.schema_migrations order by 1"

# 1e. Tidak boleh ada job ingest yang sedang berjalan (worker akan dihentikan)
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select status, count(*) from ops.job_queue group by 1"

# 1f. Catat state awal untuk dibandingkan di post-check
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select 'sources', count(*) from bronze.registered_sources"
docker inspect auditsphere_datalake --format 'DB started={{.State.StartedAt}} id={{.Id}}' | tee /root/datahub-deploy-before.txt
docker inspect auditsphere_datahub_api auditsphere_datahub_worker --format '{{.Name}} image={{.Image}}'
```

Kriteria lanjut: 1b menunjukkan `volume <project>_datalake_data`; 1d berisi `001`..`008`; 1e tidak ada status `running` (bila ada, tunggu selesai).

Catatan 1d: bila `ops.schema_migrations` tidak ada atau tidak lengkap, runner baru akan mendeteksi objek lama dan menandainya applied (backfill), tetapi jangan lanjut deploy sebelum hasilnya diperiksa bersama pemilik kode migrasi.

---

## 2. Backup (wajib sebelum build)

Pilihan: `pg_dump -Fc` (logical). Cocok selama database kecil sampai menengah (kira-kira di bawah 50 GB, sesuai hasil 1c; demo VPS 100 GB disk). Untuk skala besar (ratusan GB, profil `prod`) pakai snapshot volume dengan DB berhenti (bagian 2B) atau snapshot disk VPS dari panel penyedia.

### 2A. pg_dump terenkripsi (default)

```bash
umask 077
mkdir -p /root/backups && cd /root/backups
TS=$(date +%Y%m%d-%H%M%S)

# Passphrase: simpan di file 600 yang dibuat manual (jangan di-echo ke terminal/log)
test -s /root/.datahub_backup_pass || { echo "buat /root/.datahub_backup_pass (chmod 600) dulu"; exit 1; }

docker exec auditsphere_datalake pg_dump -U postgres -d auditsphere_datalake -Fc \
  | openssl enc -aes-256-cbc -pbkdf2 -salt -pass file:/root/.datahub_backup_pass \
  > datalake-$TS.dump.enc

ls -lh datalake-$TS.dump.enc
sha256sum datalake-$TS.dump.enc | tee datalake-$TS.dump.enc.sha256

# Verifikasi dapat dibaca (dekripsi + daftar isi, tanpa restore)
openssl enc -d -aes-256-cbc -pbkdf2 -pass file:/root/.datahub_backup_pass < datalake-$TS.dump.enc \
  | docker exec -i auditsphere_datalake pg_restore -l | head -5
```

Pertimbangan keamanan:
- Dump memuat `bronze.registered_sources` (kolom `password_encrypted` isinya kredensial sumber klien) dan data perbankan. Selalu terenkripsi seperti di atas; jangan simpan dump plaintext.
- Salinan off-box: kirim lewat WireGuard ke audit server (`10.0.0.2`) atau penyimpanan lain yang masih dalam kendali kita dan di Indonesia. Jangan ke bucket di luar negeri (data klien perbankan, kewajiban lokasi data). Contoh: `scp datalake-$TS.dump.enc* root@10.0.0.2:/root/datahub-backups/`.
- Simpan passphrase terpisah dari file dump.
- Restore hypertable TimescaleDB memerlukan `SELECT timescaledb_pre_restore();` sebelum dan `SELECT timescaledb_post_restore();` sesudah `pg_restore` ke database kosong yang sudah `CREATE EXTENSION timescaledb`. Restore tidak dilakukan pada prosedur ini; ia hanya jalan darurat.

### 2B. Snapshot volume (alternatif untuk data besar, perlu downtime DB)

```bash
docker compose stop postgres-datalake            # DB berhenti bersih; API/worker akan error sampai DB hidup lagi
docker run --rm -v <project>_datalake_data:/src:ro -v /root/backups:/out alpine \
  tar czf /out/datalake_data-$(date +%Y%m%d-%H%M%S).tgz -C /src .
docker compose start postgres-datalake
```
Lalu enkripsi tgz-nya (`openssl enc` seperti 2A) dan salin off-box. Pastikan ruang disk cukup (cek 1c).

---

## 3. Sinkronkan kode (tanpa --delete)

Kode di server disalin dari repo (bukan git). `./migrations` di-mount read-only ke container, jadi file migrasi di host HARUS diperbarui sebelum recreate.

Dari mesin pengembang (repo `risk-based-audit`), lewat WireGuard:

```bash
HUB=root@10.0.0.1        # alamat VPN server Data Hub
cd /path/ke/risk-based-audit/data-server

rsync -av --exclude '__pycache__' --exclude 'tests' data-hub-api/ $HUB:/opt/auditsphere-data/data-hub-api/
rsync -av migrations/                                           $HUB:/opt/auditsphere-data/migrations/
# JANGAN tambahkan --delete. JANGAN rsync .env, nginx/htpasswd, wireguard/, volumes/.
```

Perubahan pada `docker-compose*.yml`: bila ikut berubah, bandingkan dulu (`diff`) dan salin manual; perubahan env/volume pada service DB membuat compose me-recreate DB.

---

## 4. Tandai image saat ini (untuk rollback) lalu build

Di server:

```bash
cd /opt/auditsphere-data
P=$(docker inspect auditsphere_datahub_api --format '{{index .Config.Labels "com.docker.compose.project"}}')
DATE=$(date +%Y%m%d)

# Simpan image yang sedang jalan sebagai tag rollback
docker tag "$(docker inspect auditsphere_datahub_api    --format '{{.Image}}')" $P-data-hub-api:rollback-$DATE
docker tag "$(docker inspect auditsphere_datahub_worker --format '{{.Image}}')" $P-datahub-worker:rollback-$DATE
docker images | grep rollback-$DATE

# Build saja (belum mengganti container apa pun)
OVERLAY="-f docker-compose.yml -f docker-compose.demo.yml"   # sesuaikan dengan hasil 1a; hanya -f docker-compose.yml bila tanpa overlay
docker compose $OVERLAY build data-hub-api datahub-worker
```

---

## 5. Recreate HANYA API dan Worker

```bash
docker compose $OVERLAY up -d --no-deps --force-recreate data-hub-api datahub-worker
```

`--no-deps` mencegah compose menyentuh `postgres-datalake`. Jangan pakai `down`, `-v`, `--remove-orphans`, ataupun `up -d` tanpa nama service.

Saat start, API dan worker menjalankan migrasi yang belum tercatat (saat ini `009_registered_sources_ssl_mode.sql`: hanya `ADD COLUMN IF NOT EXISTS` + `UPDATE` ssl_mode dari flag lama; aditif, tidak menghapus data). Runner migrasi memakai advisory lock sehingga API (2 uvicorn worker) dan worker tidak berebut.

---

## 6. Post-check

```bash
sleep 20

# 6a. DB tidak ikut restart (harus sama dengan /root/datahub-deploy-before.txt)
docker inspect auditsphere_datalake --format 'DB started={{.State.StartedAt}} id={{.Id}}'
cat /root/datahub-deploy-before.txt

# 6b. Health (tanpa API key) dan status container
curl -s http://127.0.0.1:8100/health ; echo
docker compose $OVERLAY ps

# 6c. Log migrasi: harus ada "All migrations are up to date" dan TIDAK ada "Migration failed"
docker logs --since 10m auditsphere_datahub_api    2>&1 | grep -i "migrat"
docker logs --since 10m auditsphere_datahub_worker 2>&1 | grep -i "migrat"

# 6d. Catatan migrasi di DB, jumlah sumber tidak berkurang
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select version from ops.schema_migrations order by 1"
docker exec auditsphere_datalake psql -U postgres -d auditsphere_datalake -Atc "select 'sources', count(*) from bronze.registered_sources"
```

PENTING: kegagalan migrasi pada API hanya dicatat sebagai `Warning` dan `/health` tetap `ok`. Karena itu 6c/6d wajib; jangan menyimpulkan sukses dari `/health` saja.

```bash
# 6e. Daftarkan sumber uji (butuh DATA_HUB_API_KEY; baca dari env container, jangan dicetak)
KEY=$(docker exec auditsphere_datahub_api printenv API_KEY)
curl -s -X POST http://127.0.0.1:8100/api/v1/sources/register \
  -H "X-API-Key: $KEY" -H 'Content-Type: application/json' \
  -d '{"source_id":"00000000-0000-0000-0000-00000000dead","name":"deploy-smoke","host":"example.invalid","database_name":"x","ssl_mode":"verify-full"}'
echo
# Harapan: status success DAN respons memuat "ssl_mode":"verify-full" (master-service memeriksa echo ini).

# 6f. Hapus sumber uji
curl -s -X DELETE http://127.0.0.1:8100/api/v1/sources/00000000-0000-0000-0000-00000000dead -H "X-API-Key: $KEY"; echo
unset KEY
```

Gerbang: bila respons 6e tidak memuat `ssl_mode`, master-service akan menolak semua pendaftaran dengan SSL selain `disable` (`SSLModeNotAppliedError`). Jangan lanjut ke master-service sebelum ini hijau.

---

## 7. Urutan deploy: Data Hub dulu, baru master-service

1. Data Hub (runbook ini).
2. Smoke 6e hijau.
3. Baru deploy master-service (lewat `backend/deploy-prod.sh --ref <sha>` sesuai prosedur backend, tanpa seeder).

Alasan:
- master-service versi baru mengirim `ssl_mode` / `ssl_root_cert` saat mendaftarkan sumber dan memverifikasi bahwa Data Hub mengembalikan `ssl_mode` yang sama. Data Hub lama mengabaikan field itu (Pydantic membuang field tak dikenal) sehingga `verify-full` diam-diam turun menjadi `require` lewat flag lama `ssl_enabled`; master-service yang baru memang menolaknya, tetapi pengguna melihat error.
- Sebaliknya Data Hub baru menerima payload master-service lama (`ssl_mode` opsional, diturunkan dari `ssl_enabled`), jadi Data Hub-first kompatibel mundur dan tidak ada jendela rusak.
- Migrasi 009 (kolom `ssl_mode`, `ssl_root_cert`) harus sudah ada di data lake sebelum master mulai mengirim field itu.
- Rollback master-service tidak memerlukan rollback Data Hub; urutan terbalik (master dulu) justru memaksa keduanya di-rollback bersamaan.

---

## 8. Rollback (kembali ke image sebelumnya)

Skema DB tidak perlu di-rollback: migrasi 009 aditif dan kolom tambahan tidak mengganggu kode lama (kode lama tidak menulis `ssl_mode`; kolomnya NULL). File `009` di `./migrations` boleh dibiarkan.

```bash
cd /opt/auditsphere-data
P=$(docker inspect auditsphere_datahub_api --format '{{index .Config.Labels "com.docker.compose.project"}}')
DATE=<tanggal tag rollback dari langkah 4>

docker tag $P-data-hub-api:rollback-$DATE  $P-data-hub-api:latest
docker tag $P-datahub-worker:rollback-$DATE $P-datahub-worker:latest

# --no-build: pakai image yang baru di-retag, jangan build ulang
docker compose $OVERLAY up -d --no-deps --no-build --force-recreate data-hub-api datahub-worker
curl -s http://127.0.0.1:8100/health
```

Bila master-service sudah terlanjur di-deploy, rollback master-service juga (atau terima bahwa pendaftaran SSL non-`disable` akan ditolak).

Restore data (hanya bila data rusak, bukan untuk rollback image): hentikan API/worker, buat DB kosong, `timescaledb_pre_restore()`, `pg_restore`, `timescaledb_post_restore()`, jalankan lagi. Lakukan bersama tim, jangan improvisasi.

---

## 9. Pembersihan

```bash
docker image prune -f        # hanya image dangling; BUKAN `docker system prune --volumes`
# Hapus tag rollback dan file backup lokal setelah beberapa hari stabil, dan pastikan salinan off-box terenkripsi ada.
```
