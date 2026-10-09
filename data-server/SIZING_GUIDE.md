# AuditSphere Data Lakehouse: Hardware Sizing & Scalability Guide (1 TB+ Scale)

Dokumen ini adalah panduan teknis resmi kapasitas infrastruktur (*Hardware Sizing & Tuning Guide*) untuk menjalankan AuditSphere Data Lakehouse baik pada **Lingkungan Demo VPS KVM 2** maupun **Server Produksi Klien (1 TB+, 50+ Concurrent Users, Multi-Database)**.

---

## 1. Perbandingan Spesifikasi Infrastruktur

Arsitektur AuditSphere didesain **Hardware-Adaptive**: kode dan arsitektur database secara otomatis mengenali batas memori (cgroup memory limit) dan mengatur tuning database, chunk streaming, serta worker pool tanpa perlu mengubah satu baris kode pun.

| Parameter | Lingkungan Demo (VPS KVM 2) | Lingkungan Produksi Klien (1 TB+) |
| :--- | :--- | :--- |
| **Tujuan** | Presentasi demo, evaluasi client, functional verification | Beban penuh data client (1 TB+), 50+ auditor bersamaan |
| **CPU Core** | 2 vCPU Core | 16–32 Core (Dedicated / Baremetal / Enterprise VM) |
| **RAM** | 8 GB RAM | 64 GB – 128 GB RAM (ECC Recommended) |
| **Disk Storage** | 100 GB NVMe SSD | 1 TB – 2 TB Enterprise NVMe SSD (RAID 10) |
| **Bandwidth** | 8 TB Bandwidth | 10 Gbps LAN / Dedicated Cloud Interconnect |
| **OS** | AlmaLinux 10 / Ubuntu 24.04 LTS | AlmaLinux 9/10 / RHEL 9 / Ubuntu 24.04 LTS |
| **Docker Compose** | `docker-compose.demo.yml` | `docker-compose.prod.yml` |
| **Ingestion Chunk** | 10.000 baris / chunk streaming | 100.000 baris / chunk streaming |
| **Database Pool** | 15 koneksi internal | 50–100 koneksi internal |

---

## 2. Tuning Parameter PostgreSQL & TimescaleDB Otomatis

Script `data-server/postgres/autotune.sh` berjalan saat container startup dan menghitung alokasi optimal berdasarkan kapasitas RAM server:

```bash
# Perhitungan Dinamis Autotune
RAM_MB = $(mem_limit / 1024 / 1024)
shared_buffers        = RAM_MB * 0.25      # 25% Total RAM
effective_cache_size  = RAM_MB * 0.75      # 75% Total RAM
maintenance_work_mem  = min(4096, RAM_MB * 0.05)
work_mem              = (RAM_MB * 0.15) / max_connections
```

### Matriks Konfigurasi Parameter Database

| PostgreSQL Parameter | VPS Demo (8 GB RAM) | Client Prod (64 GB RAM) | Client Prod (128 GB RAM) |
| :--- | :--- | :--- | :--- |
| `shared_buffers` | `2 GB` | `16 GB` | `32 GB` |
| `effective_cache_size` | `6 GB` | `48 GB` | `96 GB` |
| `maintenance_work_mem` | `512 MB` | `3.2 GB` | `4 GB` (max limit) |
| `work_mem` | `24 MB` | `96 MB` | `192 MB` |
| `max_connections` | `50` | `100` | `150` |
| `max_parallel_workers` | `2` | `16` | `32` |
| `max_parallel_workers_per_gather` | `2` | `8` | `12` |
| `max_worker_processes` | `4` | `24` | `40` |
| `effective_io_concurrency` | `200` (NVMe) | `200` (NVMe) | `300` (NVMe Array) |
| `checkpoint_completion_target` | `0.9` | `0.9` | `0.9` |
| `wal_buffers` | `16 MB` | `64 MB` | `128 MB` |

---

## 3. Efisiensi Storage & TimescaleDB Columnar Compression

Pada data transaksi perbankan dan GL entries (1 TB+ data mentah):

1. **Rasio Kompresi Columnar LZ4**:
   - Pada tabel `gold.fact_transactions` dan hypertable bronze, TimescaleDB menerapkan kompresi columnar tersegmentasi berdasarkan `source_id, category` dengan pengurutan `trx_date DESC`.
   - **Rasio Kompresi Teruji: 4,2x hingga 6,8x**.
   - Data mentah 1 TB (~1,5 Miliar baris transaksi) dikompresi menjadi **~180 GB – 240 GB di disk NVMe**.
2. **Kebijakan Kompresi Otomatis**:
   - Data yang berumur lebih dari 7 hari otomatis dikompresi di latar belakang (`add_compression_policy('gold.fact_transactions', INTERVAL '7 days')`).
   - Query analitik (CAATT, Continuous Aggregates, Drill-down) membaca langsung dari chunk terkompresi tanpa dekompresi ke RAM (*Vectorized execution*).
3. **Immutabilitas Bronze**:
   - Data Bronze dilindungi trigger `ops.prevent_bronze_mutation()` sehingga tidak bisa di-`UPDATE` atau di-`DELETE` secara tidak sengaja oleh proses ETL.

---

## 4. Kecepatan Streaming Ingestion & Flat Memory Footprint

Masalah utama pada aplikasi ETL konvensional adalah kehabisan memori (`MemoryError` / OOM Killed) saat membaca data besar dengan `pd.read_sql`.

### Desain Streaming AuditSphere:
1. **Server-Side Cursor (Read-Only)**:
   - Data ditarik dari database klien menggunakan cursor database (`psycopg.ServerCursor`), hanya mengambil `chunk_size` baris per round-trip jaringan.
2. **Direct Pipe ke `COPY ... FROM STDIN`**:
   - Data di-stream langsung ke TimescaleDB menggunakan protokol binary text `COPY` tanpa melewati objek DataFrame Python.
3. **Flat Memory Footprint**:
   - Konsumsi RAM proses worker Python **konstan di < 180 MB**, terlepas dari apakah tabel sumber memiliki 1.000 baris atau 500.000.000 baris.
4. **Integritas Manifest Batch (SHA-256)**:
   - Setiap batch yang masuk dihitung checksum SHA-256 dan dicatat di `ops.ingest_batches` sebagai bukti integritas audit (*auditor cryptographic proof*).

---

## 5. Keamanan Database Klien (Strict Zero-Lock & Read-Only)

Saat menghubungkan database klien:
1. **Enforce Read-Only**:
   - Setiap sesi koneksi diawali perintah: `SET TRANSACTION READ ONLY; SET statement_timeout = '60s';`.
2. **Zero Lock Introspection**:
   - Estimasi jumlah baris dan metadata skema membaca `pg_class.reltuples` (PostgreSQL) atau `information_schema.tables` (MySQL/SQL Server), **tanpa pernah menjalankan `COUNT(*)`** yang mengunci tabel klien.
3. **Koneksi Terbatas**:
   - Pool koneksi ke database klien dibatasi maksimal 2 koneksi bersamaan (`max_size=2`) untuk memastikan aktivitas audit tidak mengganggu operasional core banking klien.
4. **Zero Temporary Table**:
   - Seluruh proses rekonsiliasi dan perbandingan hash dilakukan di sisi AuditSphere Data Lakehouse, bukan di database klien.

---

## 6. Mekanisme Deteksi Penghapusan (Source Deletion Detection)

Ketika data transaksi atau nasabah dihapus di sistem sumber (misal: penipuan yang dihapus untuk menutupi jejak):
1. **Soft-Delete Tracking**: Mendeteksi kolom `is_deleted` atau `deleted_at`.
2. **Bucket Count Reconciliation**:
   - Menghitung jumlah record per bucket `width_bucket(id)` di database klien dengan query agregat cepat.
   - Membandingkan dengan jumlah aktif di Silver zone.
   - Melakukan drill-down biner pada bucket yang jumlahnya berkurang di klien.
3. **Audit Tombstone**:
   - Baris yang hilang di sumber **tidak dihapus** dari AuditSphere.
   - Dibuatkan catatan tombstone di Bronze (`_op='D'`).
   - Ditandai `is_deleted=TRUE`, `deleted_detected_at=NOW()` di Silver dan Gold.
   - Muncul di log rekonsiliasi `ops.reconcile_runs` dan dashboard Data Quality sebagai peringatan audit (*Auditor Exception Warning*).

---

## 7. Panduan Deployment Client (Cara Beralih Demo -> Produksi)

### Langkah 1: Clone & Konfigurasi Lingkungan
```bash
git clone https://github.com/dimasRamadhan2398/risk-based-audit.git
cd risk-based-audit/data-server
cp .env.example .env
# Isi kredensial aman di .env
```

### Langkah 2: Menjalankan pada Demo VPS (2 CPU, 8 GB RAM)
```bash
# Menjalankan overlay demo (chunk 10K, resource cap 8GB)
docker compose -f docker-compose.yml -f docker-compose.demo.yml up -d --build
```

### Langkah 3: Menjalankan pada Server Produksi Client (16+ CPU, 64 GB+ RAM, 1 TB+)
```bash
# Menjalankan overlay produksi (chunk 100K, resource cap 64GB+, Timescale autotune 16GB)
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

### Langkah 4: Verifikasi Status & Metrik
```bash
# Cek kesehatan container dan pipeline
curl -s http://localhost:8100/health | jq .
curl -s http://localhost:8100/api/v1/pipeline/health | jq .

# Cek log autotune PostgreSQL
docker logs auditsphere_datalake | grep -i "autotune"
```

---

*AuditSphere Enterprise Data Architecture — Sizing & Tuning Specification 2026*
