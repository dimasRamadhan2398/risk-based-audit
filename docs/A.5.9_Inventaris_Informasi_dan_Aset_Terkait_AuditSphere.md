# A.5.9 Inventaris Informasi dan Aset Terkait

> *Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.*

---

| Parameter | Nilai |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Inventaris informasi dan aset terkait |
| **Fokus pemeriksaan** | Aset memiliki pemilik, klasifikasi, lokasi, dan status terkini |
| **Contoh bukti objektif** | Asset register, CMDB, daftar pemilik, hasil rekonsiliasi |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.160 / 202.10.34.166 dan Cloud SaaS |
| **PIC** | Lead Information Security / DevSecOps (Reza) |
| **Reviewer** | [isi saat validasi] (Dimas / Group Vision) |
| **Tanggal validasi** | [isi saat validasi] |

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| **Relevansi SoA** | **Berlaku** | Aset informasi audit, model AI, dan infrastruktur komputasi adalah inti operasional AuditSphere; inventarisasi aset wajib dipenuhi untuk penegakan kontrol keamanan. |
| **Status implementasi** | **Dalam proses** | Daftar aset informasi utama (Bab 4 ISO-AS-001) dan inventaris kontainer/layanan di docker-compose sudah terdefinisi; Master Information Asset Register (IAR) terpadu belum diformalkan. |
| **Ketersediaan bukti** | **Ada sebagian** | Inventaris data utama, topologi arsitektur VM/VLAN, dan konfigurasi repo git tersedia; formulir registrasi aset terpadu, CMDB otomatis, dan BA rekonsiliasi tahunan belum lengkap. |
| **Evidence ID / URL** | **Belum terdaftar** | Usulan EV-A5.9-01 s/d EV-A5.9-05 (lihat bagian 6). ID perlu didaftarkan di Master Audit Evidence Index. |
| **Efektivitas operasional** | **Sebagian efektif** | Pemisahan kontainer dan volume data berjalan efektif sesuai spesifikasi; namun pemutakhiran berkala status aset dan verifikasi kepemilikan belum terotomatisasi. |
| **Desain kontrol** | **Perlu perbaikan** | Spesifikasi teknis aset infrastruktur sangat detail; mekanisme formal pemeliharaan siklus hidup aset (pencatatan, perubahan, rekonsiliasi, decommissioning) perlu perbaikan. |
| **Status kesiapan** | **Belum siap** | Daftar aset terpusat bertanda tangan dan mekanisme rekonsiliasi berkala belum memenuhi seluruh gate kesiapan audit sertifikasi. |
| **Prioritas** | **Tinggi** | Prasyarat fundamental untuk kontrol A.5.10 (Penggunaan yang dapat diterima), A.5.12 (Klasifikasi informasi), A.5.13 (Pelabelan), dan A.8.1 (Perangkat titik akhir). |
| **Review berikutnya** | **[isi]** | Usulan: 12 bulan setelah Master IAR disahkan atau setiap kali terjadi penambahan microservice / perubahan topologi server mayor. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.9 **berlaku** untuk AuditSphere karena:

1. **Sensitivitas Nilai Aset Informasi**: AuditSphere mengelola aset informasi sensitif berskala enterprise, termasuk data audit internal rahasia, profil risiko (*Enterprise Risk Management*), kertas kerja audit digital (KKA), temuan kepatuhan, algoritma inferensi Machine Learning, serta kredensial otentikasi. Seluruh aset tersebut memerlukan inventarisasi yang jelas agar dapat dilindungi sesuai nilai kritikalitasnya.
2. **Keterikatan Framework ISO/IEC 27001:2022**: Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) Bagian 4.2 dan 7 menetapkan klausul A.5.9 s/d A.5.14 sebagai kelompok kontrol Manajemen Aset dan Klasifikasi Data yang wajib dipatuhi oleh seluruh sistem dan lingkungan komputasi AuditSphere.
3. **Mandat Standar Tata Kelola ISO 27001:2022 Klausul A.5.9**: Standar ISO/IEC 27001:2022 kontrol Annex A.5.9 mewajibkan organisasi untuk mengidentifikasi informasi dan aset terkait lainnya, serta menyusun dan memelihara inventaris aset tersebut lengkap dengan penunjukan pemilik aset (*asset owner*).
4. **Kompleksitas Lingkungan Hybrid & Multi-Tenant**: Arsitektur AuditSphere bersifat multi-tier dan hybrid (on-premises Proxmox VE dan Cloud SaaS AWS/VPS), sehingga tanpa inventaris aset yang akurat, risiko timbulnya celah keamanan berupa shadow IT, server/container terlantar, data store tak terenkripsi, atau dependensi usang akan meningkat secara signifikan.
5. **Kebutuhan Due Diligence Klien Enterprise**: Klien korporasi pada tahapan *vendor security risk assessment* dan *due diligence* audit keamanan eksternal mewajibkan penyedia solusi SaaS menunjukkan daftar inventaris aset (Asset Register / CMDB) yang terverifikasi dan termutakhirkan secara konsisten.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Inventaris Aset Informasi Utama (Data & Informasi)

Dokumen `docs/DOKUMEN_KESELARASAN_ISO27001.md` Bagian 4.2 telah memetakan inventaris awal aset informasi utama platform AuditSphere beserta deskripsi, klasifikasi, dan pemiliknya:

| ID Aset | Aset Informasi | Deskripsi & Format | Klasifikasi | Pemilik Aset | Lokasi Penyimpanan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **AST-INF-01** | **Data Kertas Kerja Audit (KKA)** | Working papers F-01 s/d F-05, pengujian substantif, notula, format JSON & DB records | **Restricted / Confidential** | Klien / CAE | PostgreSQL (`audit_db`) & S3 Object Storage |
| **AST-INF-02** | **Dokumen Bukti Audit (Evidence)** | Berkas pendukung temuan audit, invoice sampel, screenshoot, format PDF, XLSX, DOCX, JPG | **Restricted / Confidential** | Lead Auditor / Auditee | Encrypted S3 Bucket / NAS Terisolasi |
| **AST-INF-03** | **Register Risiko & Profil ERM** | Data Risk Universe, Inherent/Residual Risk, matriks RCM, format tabel relasional | **Confidential** | Unit Risk Management | PostgreSQL (`risk_db`) |
| **AST-INF-04** | **Model Prediksi & Bobot AI** | Bobot inferensi Machine Learning Anomaly Detection, Department Classifier, format `.pkl` / `.pt` | **Confidential** | AI Lead (Dimas) | Service Container / Storage Volume |
| **AST-INF-05** | **Kredensial & Secrets Sistem** | Password hash Bcrypt, private keys, JWT secret, TLS certs, WireGuard keys, format text/pem | **Restricted** | Lead InfoSec (Reza/Qiko) | Redis Session, Vault / Environment |
| **AST-INF-06** | **Konfigurasi Infrastruktur** | Docker Compose, routing Kong, Nginx, manifest Proxmox & firewall rules, format YAML/Shell | **Internal / Restricted** | DevOps & Infra Lead | Git Repository & Host Production |
| **AST-INF-07** | **Log Audit Trail & Aktivitas** | Perekaman audit trail who-when-what, status perubahan workpaper, auth logs, format JSON | **Confidential** | Compliance & InfoSec | Apache Kafka, PostgreSQL (`sys_logs`) |
| **AST-INF-08** | **Data Master & Multi-Tenant** | Struktur organisasi, data pegawai, konfigurasi domain tenant, format tabel relasional | **Confidential / Restricted** | Master Data Lead (Qiko) | PostgreSQL (`master_db`) |

### 2.2 Inventaris Aset Perangkat Lunak & Microservices (Software Assets)

AuditSphere tersusun atas komponen perangkat lunak independen berbasis microservices dan *containerized architecture* yang terdefinisi pada berkas `backend/docker-compose.yml`:

1. **Frontend Application**: Nuxt 4 / Vue 3 Framework (Nuxt UI, Tailwind CSS, Pinia state management) beroperasi pada port 3000.
2. **API Gateway & Reverse Proxy**: Kong Gateway 3.4 (declarative mode, port 8000/8443) dengan plugin bot detection, TLS termination, dan rate limiting.
3. **Auth & Identity Service**: Golang 1.21+ microservice (port 8081) mencakup modul RBAC Casbin, MFA TOTP, Refresh Token, dan Confidentiality Agreement.
4. **Audit Planning & Fieldwork Service**: Golang 1.21+ microservice (port 8082) mencakup Audit Universe, Charter, PKAT, Digital Workpaper F-01 s/d F-05, LHA, dan ATR.
5. **Risk Management Service**: Golang 1.21+ microservice (port 8083) mencakup Risk Profile, Risk Matrix, Mitigation Plans, dan RCM.
6. **Master Data & Tenant Service**: Golang 1.21+ microservice (port 8084) mencakup pengelolaan Company, Department, Employee, dan skrip otomasi onboarding tenant.
7. **Analytics & Aggregator Service**: Golang 1.21+ microservice (port 8085) untuk kalkulasi agregat analitik dan penyedia data dasbor eksekutif.
8. **AI / ML Inference Engine**: Python 3.10+ FastAPI microservice (port 8000 internal) menjalankan algoritma Machine Learning (Isolation Forest, XGBoost, PyTorch, IndoBERT).
9. **Databases & Message Broker**: PostgreSQL 16 (Patroni HA, port 5432), Redis 7 (In-Memory Cache & Session Store, port 6379), serta Apache Kafka & Zookeeper (port 9092) untuk streaming audit event.

### 2.3 Inventaris Aset Komputasi, Perangkat Fisik & Cloud (Infrastructure Assets)

Berdasarkan dokumen `docs/architecture_on_premises_vs_cloud.md` dan panduan deployment backend, aset infrastruktur terbagi ke dalam lingkungan berikut:

1. **Server Produksi VPS**: Server Linux Ubuntu 22.04 LTS beralamat IP `202.10.34.160` / `202.10.34.166` yang menjalankan seluruh stack kontainer Docker aplikasi.
2. **Infrastruktur Virtualisasi On-Premises**: Host Proxmox VE (3 Physical Nodes HA) yang dipartisi menjadi Virtual Machines terisolasi per-tier (Load Balancer Keepalived/HAProxy, K8s Worker Nodes, Database State Tier, dan Message Broker Tier).
3. **Segmentasi Jaringan VLAN**: Segmentasi 5 zona jaringan (VLAN Ingress/DMZ, VLAN Application, VLAN Secure Data Tier, VLAN AI/ML, dan VLAN Management).
4. **Aset Jaringan & Nama Domain**: Domain publik `auditsphere.app`, domain wildcard `*.auditsphere.io`, sertifikat TLS 1.3 (Let's Encrypt / Automated Certbot), dan WireGuard VPN Tunnel.
5. **Cloud SaaS Resources**: AWS Jakarta Region (`ap-southeast-3`) mencakup S3 Storage Bucket, Route 53 DNS, CloudFront CDN untuk static assets, dan Application Load Balancer.

### 2.4 Mekanisme Penelusuran Kepemilikan (Asset Ownership) & Kontrol Perubahan

Dokumen `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md` menetapkan penanggung jawab teknis atas setiap modul perangkat lunak (Qiko memegang modul Gateway, Auth, Master Data, DevOps; Dimas memegang modul AI/ML, Risk, Analytics, Audit Planning, QAIP; Reza dan Sumanadi memegang kepatuhan kontrol Annex A). Seluruh perubahan konfigurasi dan kode sumber dikendalikan melalui sistem kontrol versi Git dengan audit log commit yang tidak dapat diubah.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :---: | :--- |
| **Inventaris Aset Informasi Utama (Data Inventory)** | **Tersedia** | Terdokumentasi pada Bab 4 `docs/DOKUMEN_KESELARASAN_ISO27001.md` mencakup 8 kategori aset utama. |
| **Daftar Perangkat Lunak & Microservices (Software Inventory)** | **Tersedia** | Tercantum lengkap dalam arsitektur sistem, `docker-compose.yml`, dan `go.mod`/`package.json`. |
| **Inventaris Infrastruktur Server, VM & Jaringan** | **Tersedia** | Spesifikasi VPS 202.10.34.160/166, Proxmox VM tiers, konfigurasi VLAN, dan topologi cloud tercatat. |
| **Peta Kepemilikan Aset Teknis (Asset Ownership Mapping)** | **Tersedia** | Terdokumentasi rinci dalam `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md`. |
| **Master Information Asset Register (IAR) Terpadu SMKI** | **Belum tersedia** | Belum ada satu dokumen induk spreadsheet/database Asset Register terpadu berformat ISO 27001. |
| **Surat Penunjukan Pemilik Aset Formal (Formal Asset Owner)** | **Belum tersedia** | Penetapan kepemilikan aset masih bersifat operasional internal, belum ada SK/SKK formal dari pimpinan puncak. |
| **Bukti Rekonsiliasi & Verifikasi Aset TIK Berkala** | **Belum tersedia** | Belum pernah dilakukan dan didokumentasikan stock opname aset TIK atau rekonsiliasi container/database berkala. |
| **Inventaris Lisensi Pihak Ketiga & SBOM (Software Bill of Materials)** | **Ada sebagian** | Dependensi tercatat pada lockfiles (`package-lock.json`, `go.sum`), namun belum dikompilasi ke laporan SBOM resmi. |
| **SOP Siklus Hidup & Pemusnahan Aset (Asset Lifecycle & Disposal SOP)** | **Belum tersedia** | Belum ada prosedur tertulis formal mengenai onboarding, perubahan, hingga penghapusan/disposal aman aset informasi usang. |

---

## 4. Catatan / Gap Terverifikasi

1. **Belum tersedianya Master Information Asset Register (IAR) terpadu**: Inventaris aset saat ini tersebar di berbagai berkas dokumentasi terpisah (`DOKUMEN_KESELARASAN_ISO27001.md`, `architecture_on_premises_vs_cloud.md`, `docker-compose.yml`), dan belum dikonsolidasikan ke dalam satu register induk resmi yang memuat metadata SMKI lengkap.
2. **Atribut aset belum memenuhi parameter ISO/IEC 27001:2022 secara menyeluruh**: Atribut penting seperti penilaian tingkat kritikalitas CIA (*Confidentiality, Integrity, Availability*), batas retensi data, metode enkripsi at-rest, persyaratan cadangan (*backup frequency*), serta prosedur pemusnahan (*disposal method*) belum distandarisasi pada setiap entri aset.
3. **Belum ada prosedur rekonsiliasi aset berkala (Stock Opname TIK)**: Belum ada mekanisme rutin untuk merekonsiliasi aset aktif (seperti container yang berjalan, database volume, VM Proxmox, sertifikat domain, dan akun cloud) dengan daftar aset tercatat, sehingga berpotensi menimbulkan shadow service atau zombie resource yang tidak terkelola.
4. **Kompilasi SBOM dan inventaris lisensi perangkat lunak belum otomatis**: Dependensi open-source pada service Go, Python, dan frontend Nuxt belum dipindai secara berkala menggunakan perangkat *Software Composition Analysis* (SCA) untuk mendeteksi lisensi berisiko (misal GPL non-kompatibel) atau kerentanan CVE pada library pihak ketiga.
5. **Prosedur penonaktifan dan sanitasi data (Decommissioning & Secure Disposal) belum dibakukan**: Belum ada SOP tertulis mengenai pembersihan aman data klien saat masa sewa/tenant berakhir, penghapusan *cryptographic key*, atau sanitasi media penyimpanan bekas sesuai standar sanitasi data (NIST SP 800-88).

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :-: | :--- | :--- | :-: | :-: | :--- |
| **1** | Susun, konsolidasikan, dan sahkan Master Information Asset Register (IAR) terpadu mencakup aset informasi, perangkat lunak, infrastruktur, dan repositori data. | Lead Information Security (Reza) | **Tinggi** | `[isi]` | Dokumen Master IAR disetujui |
| **2** | Tetapkan dan dokumentasikan surat/berita acara penunjukan Pemilik Aset (*Asset & Data Owner*) definitif untuk setiap kelompok aset informasi. | Leadership (Dimas) & Lead InfoSec | **Tinggi** | `[isi]` | Formulir Penetapan Asset Owner |
| **3** | Lakukan inventarisasi fisik & logis perdana serta susun Berita Acara Rekonsiliasi Aset TIK (verifikasi VPS, container, database, dan SSL). | Tim DevOps & Infra (Qiko / Sumanadi) | **Tinggi** | `[isi]` | BA Rekonsiliasi Aset TIK |
| **4** | Susun Standar Operasional Prosedur (SOP) Siklus Hidup Aset Informasi (Pengadaan, Registrasi, Perubahan, Retensi, dan Pemusnahan Aman). | Lead Information Security (Reza) | **Sedang** | `[isi]` | SOP Siklus Hidup Aset SMKI |
| **5** | Integrasikan perkakas pemindaian dependensi otomatis (SCA) pada pipeline CI/CD untuk menghasilkan Software Bill of Materials (SBOM) dan lisensi pihak ketiga. | Tim DevSecOps (Reza) | **Sedang** | `[isi]` | Laporan SBOM & SCA di CI/CD |
| **6** | Otomatisasi pencatatan aset database tenant baru ke dalam katalog data saat skrip `scripts/onboard-tenant.sh` dijalankan. | Tim Backend (Qiko) | **Sedang** | `[isi]` | Skrip provisioning termutakhirkan |
| **7** | Tetapkan jadwal rutin rekonsiliasi dan review inventaris aset minimal satu kali setiap 6 bulan atau saat perubahan topologi mayor. | Lead Information Security | **Sedang** | `[isi]` | Jadwal & Notulen Review Aset |
| **8** | Daftarkan usulan Evidence ID untuk kontrol A.5.9 ke dalam Master Audit Evidence Index. | Lead Information Security | **Sedang** | `[isi]` | Entri Master Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| **EV-A5.9-01** | Daftar Aset Informasi Utama Organisasi (Bab 4.2) | `docs/DOKUMEN_KESELARASAN_ISO27001.md` | Internal |
| **EV-A5.9-02** | Spesifikasi Infrastruktur & Pembagian VM/VLAN | `docs/architecture_on_premises_vs_cloud.md` | Internal |
| **EV-A5.9-03** | Berkas Orkestrasi Microservices & Inventaris Container | `backend/docker-compose.yml` | Internal |
| **EV-A5.9-04** | Peta Pembagian Kepemilikan Modul Perangkat Lunak | `docs/FITUR_DAN_PEMBAGIAN_TUGAS_AUDITSPHERE.md` | Internal |
| **EV-A5.9-05** | Master Information Asset Register (IAR) AuditSphere *(Draf Usulan)* | `docs/security/` *(lokasi usulan)* | Restricted |

> *Setelah tindakan 1, 2, dan 3 selesai dilaksanakan, tambahkan Evidence ID untuk salinan Master IAR bertanda tangan, formulir penunjukan Asset Owner, dan Berita Acara Rekonsiliasi Aset TIK.*

