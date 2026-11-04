# A.5.12 — Klasifikasi Informasi

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Klasifikasi informasi |
| **Fokus pemeriksaan** | Informasi diklasifikasikan sesuai risiko dan kebutuhan bisnis |
| **Contoh bukti objektif** | Skema klasifikasi, inventaris informasi, sampel klasifikasi |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 29 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere menyimpan kertas kerja, bukti temuan, dan kredensial klien enterprise; kebutuhan perlindungan berbeda per jenis informasi sehingga klasifikasi diperlukan. |
| Implementasi status | **Dalam proses** | Skema klasifikasi sudah diusulkan di ISO-AS-001 bagian 4.2 dan 9 (*Partially Implemented*), tetapi belum disetujui dan belum diterapkan pada data, kode, maupun keluaran sistem. |
| Ketersediaan bukti | **Sebagian** | Ada usulan klasifikasi per aset dan definisi tingkat dampak di risk-service; belum ada kebijakan klasifikasi resmi, inventaris aset yang terverifikasi, maupun sampel label. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.12-01 s/d EV-A5.12-04) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Tidak efektif / belum teruji** | Tidak ditemukan mekanisme yang membaca atau menerapkan level klasifikasi saat data disimpan, ditampilkan, atau diekspor. |
| Desain kontrol | **Perlu perbaikan** | Skema belum punya definisi tunggal (ISO-AS-001 memakai *Internal / Confidential / Restricted* di bagian 4.2, dan *Public, Internal, Confidential, Restricted* di bagian 9); tidak ada atribut klasifikasi pada model data. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Data audit klien dan kredensial sudah ditetapkan sebagai *Confidential / Restricted* di ISO-AS-001, tetapi penanda ini belum diterapkan di sistem. |
| Review berikutnya | [isi] | Usulan: 12 bulan setelah skema disetujui, dan setiap kali ada jenis data baru (contoh: tenant baru, integrasi baru, ekspor laporan baru). |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.12 **berlaku** untuk AuditSphere karena:

1. AuditSphere menyimpan informasi dengan sensitivitas berbeda dalam satu platform: data audit klien (kertas kerja, bukti temuan, draf laporan), hasil risk scoring dan *exception flag*, log audit trail, serta kredensial dan konfigurasi infrastruktur (ISO-AS-001 bagian 4.2).
2. Keputusan perlindungan (akses, enkripsi, retensi, pembagian ke klien) bergantung pada klasifikasi. Tanpa skema yang disepakati, perlakuan terhadap tiap informasi tidak dapat dibuktikan konsisten.
3. Klien enterprise (lembaga jasa keuangan) pada due diligence akan meminta bukti bahwa informasi audit mereka diberi label dan ditangani sesuai tingkat sensitivitasnya.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Usulan skema dan inventaris aset (dokumen)

`docs/DOKUMEN_KESELARASAN_ISO27001.md` (ISO-AS-001, Draft v0.1) bagian 4.2 memuat tujuh aset informasi dengan klasifikasi usulan dan pemilik aset (*data owner*):

| Aset informasi | Klasifikasi usulan | Pemilik aset (usulan) |
| :--- | :--- | :--- |
| Data audit klien (kertas kerja, bukti temuan, notula, draf laporan) | Confidential / Restricted | Klien / CAE |
| Hasil risk scoring & exception flag | Confidential | Unit Kerja Audit & Risk Management |
| Kredensial pengguna & akses sistem | Restricted | Tim Security & IT Infrastructure |
| Konfigurasi infrastruktur | Internal / Restricted | Tim Infrastructure & DevOps |
| Kode sumber AuditSphere | Confidential | Tim Development / Engineering Lead |
| Log audit trail sistem | Confidential | Tim Keamanan Informasi & Compliance |
| Data tenant (multi-tenant cloud) | Confidential / Restricted | Product Owner & Cloud Lead |

Bagian 4.2 juga mencatat bahwa klasifikasi tersebut masih **[STATUS/PLACEHOLDER]** dan perlu diselaraskan dengan skema resmi.

Bagian 9 ISO-AS-001 (A.5.9 s/d A.5.14) menyebut skema *Public, Internal, Confidential, Restricted* untuk kertas kerja, bukti audit, dan kredensial, dengan status *Partially Implemented*.

### 2.2 Definisi tingkat dampak di sistem risiko (bukti parsial)

`backend/risk-service/cmd/seed_risk_universe.go` (sekitar baris 353–360) mendefinisikan faktor risiko **Data Sensitivity** dengan empat tingkat deskriptif (High, Medium, Low to Medium, Low) berdasarkan dampak pengungkapan. Ini adalah faktor penilaian risiko audit, **bukan** klasifikasi informasi yang melekat pada data; tidak ada kolom yang menyimpan level ini pada informasi.

### 2.3 Penanda klasifikasi pada keluaran

`backend/audit-service/controllers/performance_stats_controller.go` (baris 706 dan 840) mencetak label **"Classification: Confidential"** dan footer "Confidential & Proprietary" pada laporan performa. Label ini ditulis langsung (*hardcoded*) pada template HTML, tidak diambil dari data atau skema klasifikasi.

### 2.4 Pembatasan akses pengguna (pendukung)

Halaman `frontend/pages/auth/confidentiality.vue` dan tabel `confidentiality_agreements` mewajibkan persetujuan kerahasiaan sebelum memakai sistem (lihat A.5.1 bagian 2.2). Ini mendukung penanganan informasi, tetapi tidak membedakan tingkat klasifikasi per informasi.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Skema klasifikasi informasi (disetujui) | **Belum tersedia** | Hanya usulan di ISO-AS-001 bagian 4.2 dan 9. Tidak ada kebijakan klasifikasi tersendiri di `docs/ISO_27001/`. |
| Definisi tingkat klasifikasi dan perlakuan per tingkat (akses, enkripsi, retensi, pembagian) | **Belum tersedia** | Tidak ditemukan matriks perlakuan per tingkat. |
| Inventaris informasi / aset informasi | **Sebagian** | Tujuh aset di ISO-AS-001 bagian 4.2 dengan pemilik usulan. Belum ada register aset yang terverifikasi dan diperbarui. |
| Atribut klasifikasi pada data | **Belum tersedia** | Tidak ada kolom klasifikasi pada model data. Field `Classification` di `master-service/models/audit_finding.go` dan `audit_issue.go` berisi tipe dan keparahan temuan, bukan klasifikasi informasi. |
| Penanda klasifikasi pada keluaran (label) | **Sebagian** | Hanya laporan performa yang mencetak label, dan labelnya tetap (hardcoded). Ekspor dan laporan lain tidak diberi label. |
| Sampel klasifikasi (contoh data berlabel) | **Belum tersedia** | Tidak ditemukan sampel data berlabel di repositori. Perlu query sampel tanpa data pribadi. |
| Kesesuaian label dengan skema di dokumen klien | **Belum tersedia** | Bagian 9 ISO-AS-001 mencatat perlu diselaraskan dengan skema resmi klien (AIFL); belum ada dokumen skema tersebut di repositori. |

---

## 4. Catatan / Gap Terverifikasi

1. **Skema belum disetujui dan belum konsisten.** ISO-AS-001 memakai dua daftar tingkat berbeda: *Internal / Confidential / Restricted* (bagian 4.2) dan *Public, Internal, Confidential, Restricted* (bagian 9). Kebijakan tunggal perlu menetapkan satu daftar.
2. **Klasifikasi tidak melekat pada data.** Tidak ada atribut klasifikasi pada entitas kertas kerja, bukti, laporan, maupun kredensial. Akibatnya sistem tidak dapat menerapkan perlakuan berbeda per tingkat.
3. **Label pada laporan bersifat statis.** Label "Confidential" di `performance_stats_controller.go` ditulis tetap dan tidak diturunkan dari klasifikasi informasi yang dilaporkan.
4. **Inventaris aset belum diverifikasi.** Klasifikasi dan pemilik di ISO-AS-001 bagian 4.2 berstatus placeholder dan belum dicocokkan dengan entitas database atau service yang sebenarnya.
5. **Pemilik aset belum ditetapkan.** Kolom pemilik masih berupa peran usulan ("Klien / CAE", "Tim Development"), tanpa nama dan tanpa persetujuan.
6. **Risiko bercampur dengan klasifikasi.** Faktor *Data Sensitivity* di risk-service dan kolom *Classification* di model temuan dapat disalahartikan sebagai klasifikasi informasi; dokumen kebijakan perlu membedakannya.
7. **Klasifikasi belum dibawa ke ekspor dan integrasi.** Tidak ada aturan untuk unduhan berkas, bukti yang diunggah ke `/uploads/*`, atau data yang dikirim ke layanan pihak ketiga (contoh: AI service).

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Susun dan setujui kebijakan klasifikasi informasi: satu daftar tingkat (usulan: Public, Internal, Confidential, Restricted), definisi tiap tingkat, dan contoh per jenis informasi. | Lead Information Security | Tinggi | [isi] | Kebijakan klasifikasi bertanda tangan |
| 2 | Buat matriks perlakuan per tingkat: akses, enkripsi saat simpan dan kirim, retensi, pembagian ke klien, dan pembuangan. | Lead Information Security bersama Tim Backend | Tinggi | [isi] | Matriks perlakuan |
| 3 | Verifikasi inventaris di ISO-AS-001 bagian 4.2 terhadap entitas database dan service yang sebenarnya; tetapkan pemilik aset dengan nama dan peran. | Security Officer | Tinggi | [isi] | Register aset informasi terverifikasi |
| 4 | Selaraskan skema dengan klasifikasi data resmi AIFL dan kebijakan privasi klien, sebagaimana diminta ISO-AS-001 bagian 4.2. | Lead Information Security | Sedang | [isi] | Berita acara penyelarasan |
| 5 | Tentukan apakah klasifikasi perlu disimpan per data (kolom klasifikasi) atau cukup per jenis entitas; catat keputusan di dokumen desain sebelum implementasi. | Project Lead bersama auditsphere-solution-architect | Sedang | [isi] | Keputusan desain tertulis |
| 6 | Ganti label tetap di `performance_stats_controller.go` dengan nilai dari klasifikasi informasi yang dilaporkan. | Tim Backend (audit-service) | Sedang | [isi] | Perubahan kode dan test |
| 7 | Tambahkan label klasifikasi pada ekspor dan unduhan (laporan, kertas kerja, bukti) sesuai matriks perlakuan. | Tim Frontend & Tim Backend | Sedang | [isi] | Perubahan kode, contoh keluaran berlabel |
| 8 | Ambil sampel data berlabel dan hasil query tanpa data pribadi sebagai bukti operasional. | Tim Backend | Sedang | [isi] | Sampel klasifikasi |
| 9 | Tambahkan aturan klasifikasi untuk integrasi pihak ketiga (layanan AI, email, penyimpanan cadangan). | Lead Information Security | Sedang | [isi] | Daftar integrasi dengan tingkat data |
| 10 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.12-01 | Usulan klasifikasi aset informasi dan pemilik aset | `docs/DOKUMEN_KESELARASAN_ISO27001.md` bagian 4.2 dan 9 | Internal |
| EV-A5.12-02 | Definisi tingkat dampak (Data Sensitivity) | `backend/risk-service/cmd/seed_risk_universe.go` | Internal |
| EV-A5.12-03 | Penanda klasifikasi pada laporan performa (hardcoded) | `backend/audit-service/controllers/performance_stats_controller.go` | Internal |
| EV-A5.12-04 | Sampel data berlabel dan inventaris aset terverifikasi | Database layanan terkait (belum diambil) | Restricted |

Setelah tindakan 1, 2, dan 3 selesai, tambahkan ID untuk kebijakan klasifikasi, matriks perlakuan, dan register aset terverifikasi.
