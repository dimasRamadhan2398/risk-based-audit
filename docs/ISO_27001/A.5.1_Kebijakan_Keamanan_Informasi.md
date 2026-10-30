# A.5.1 — Kebijakan Keamanan Informasi

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Kebijakan keamanan informasi |
| **Fokus pemeriksaan** | Kebijakan disetujui, dikomunikasikan, dan ditinjau berkala |
| **Contoh bukti objektif** | Kebijakan, approval, riwayat versi, bukti sosialisasi |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Lead Information Security |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | Kebijakan keamanan informasi adalah fondasi ISMS; kontrol wajib dipenuhi. |
| Implementasi status | **Dalam proses** | Kebijakan topik sudah ada (CSP-001, AIP-001); kebijakan induk belum ditemukan. |
| Ketersediaan bukti | **Ada sebagian** | Ada dokumen kebijakan dan bukti acceptance di sistem; approval tertandatangan dan bukti sosialisasi staf belum lengkap. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.1-01 s/d EV-A5.1-04) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Sebagian efektif** | Acceptance pengguna tercatat dengan timestamp, IP, dan user agent; namun tidak ada pemicu re-acceptance saat versi berubah. |
| Desain kontrol | **Perlu perbaikan** | Struktur dokumen kebijakan memadai; mekanisme acceptance di sistem belum memvalidasi versi di sisi server. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Kebijakan induk menjadi prasyarat banyak kontrol A.5.x lainnya. |
| Review berikutnya | [isi] | Usulan: 12 bulan setelah kebijakan induk disetujui. CSP-001 dan AIP-001 ditinjau paling lambat 28 September 2027. |

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.1 **berlaku** untuk AuditSphere karena:

1. AuditSphere menyimpan dan memproses data audit internal, kertas kerja, dan kredensial klien enterprise. Organisasi memerlukan arah kebijakan keamanan yang tertulis, disetujui manajemen, dan dikomunikasikan.
2. Dokumen DOKUMEN_KESELARASAN_ISO27001 (ISO-AS-001) sudah memetakan A.5.1 s/d A.5.8 sebagai kelompok kebijakan dan tata kelola yang harus diselaraskan dengan sistem.
3. Kebijakan topik (CSP-001 Cloud Security Policy, AIP-001 AI Usage Policy) mengacu langsung ke ISO/IEC 27001:2022, dan kebijakan induk diperlukan agar keduanya memiliki dasar yang sama.
4. Persyaratan klien enterprise (Chief Audit Executive, Direktur Kepatuhan, Head of IT Security) pada due diligence memerlukan bukti kebijakan yang disetujui dan ditinjau.

---

## 2. Implementasi yang Sudah Ada

### 2.1 Kebijakan topik (tertulis)

| Dokumen | Versi | Berlaku sejak | Pemilik | Penyetuju | Review berikutnya | Lokasi |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| CSP-001 Cloud Security Policy and Infrastructure Governance Standard | 1.0 | 28 Sep 2026 | Security Officer & DevOps Lead | Direktur Utama | 28 Sep 2027 | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` |
| AIP-001 AI Usage Policy & Acceptable Use Guidelines | 1.0 | 28 Sep 2026 | Security Officer | Direktur Utama | 28 Sep 2027 | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx` |

Kedua dokumen memiliki tabel *document control* (nomor, versi, tanggal berlaku, pemilik, penulis, penyetuju, klasifikasi, review berikutnya) dan tabel riwayat revisi.

### 2.2 Acceptance pengguna di sistem (bukti sosialisasi untuk pengguna sistem)

Halaman `frontend/pages/auth/confidentiality.vue` mewajibkan setiap pengguna menyetujui *Confidentiality / Integrity Pact* sebelum memakai sistem:

- Pengguna harus menggulir hingga ≥95% sebelum tombol "Saya Setuju & Lanjutkan" aktif.
- Persetujuan dikirim ke `POST /confidentiality/accept` di auth-service.
- Disimpan di tabel `confidentiality_agreements` (`backend/auth-service/models/confidentiality_agreement.go`) dengan `user_id`, `agreement_type` (SYSTEM / ROLE / PROJECT), `title`, `content`, `version`, `accepted_at`, `ip_address`, `user_agent`.
- Status dicek saat login melalui `GET /confidentiality/status` (`frontend/stores/auth.ts`).

Bukti ini mendukung kebutuhan *acknowledgement* pengguna, tetapi memiliki keterbatasan (lihat bagian 4).

### 2.3 Riwayat versi

- Tabel riwayat revisi pada CSP-001 dan AIP-001 (saat ini hanya versi 1.0).
- Riwayat perubahan file tersedia di git (`git log -- docs/ISO_27001`), tetapi bukan log persetujuan formal.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Kebijakan keamanan informasi induk (ISMS policy) | **Belum tersedia** | Tidak ditemukan di repo. Yang ada hanya kebijakan topik (cloud, AI). |
| Dokumen kebijakan topik (CSP-001, AIP-001) | Tersedia | Format `.docx`, versi 1.0. |
| Approval tertandatangan dengan tanggal | **Belum lengkap** | Tabel control hanya mencantumkan jabatan penyetuju (Direktur Utama), tanpa tanda tangan atau tanggal persetujuan. Salinan tertandatangan perlu disimpan. |
| Dokumen ISO-AS-001 (keselarasan ISO) | Draft | Versi 0.1 (DRAFT). Tanggal berlaku dan review masih placeholder; penyetuju tercantum sebagai "Dimas (Group Vision)". |
| Riwayat versi | Sebagian | Tabel revisi ada, tetapi hanya baris versi 1.0. |
| Bukti acceptance pengguna sistem | Tersedia | Tabel `confidentiality_agreements` dengan timestamp, IP, user agent. Perlu query sampel untuk bukti operasional. |
| Bukti sosialisasi kebijakan ke staf (sesi, notulen, daftar hadir, acknowledgement tertulis) | **Belum tersedia** | Tidak ditemukan di repo. |
| Jadwal review berkala kebijakan induk | **Belum tersedia** | Belum ada kebijakan induk yang memuat jadwal review. |

---

## 4. Catatan / Gap Terverifikasi

1. **Kebijakan induk belum ada.** A.5.1 meminta kebijakan keamanan informasi tingkat organisasi. Yang ada baru kebijakan topik. ISO-AS-001 masih DRAFT v0.1 dengan placeholder.
2. **Approval belum terbukti.** CSP-001 dan AIP-001 mencantumkan penyetuju sebagai jabatan, tanpa tanda tangan atau tanggal persetujuan di dokumen.
3. **Versi acceptance tidak divalidasi server.** `ConfidentialityService.Accept` menyimpan `version` dan `content` dari request klien tanpa memeriksanya terhadap teks master. Riwayat versi acceptance bisa tidak akurat.
4. **Pengguna tidak diminta ulang saat kebijakan berubah.** `HasAcceptedLatest` hanya memeriksa `is_accepted = true` dan tidak membandingkan versi. Pengguna yang sudah menyetujui versi lama tidak akan diminta menyetujui versi baru.
5. **Kegagalan simpan acceptance bisa terlewati.** Di `handleAccept` (`frontend/pages/auth/confidentiality.vue`), blok `catch` tetap memindahkan pengguna ke langkah 3. Jika request accept gagal, pengguna lolos tanpa acceptance tercatat dan tanpa pesan kesalahan.
6. **Teks versi di UI di-hardcode.** `systemVersion: 'Risk-Based Internal Audit System v1.0'` tidak terkait dengan versi dokumen kebijakan.
7. **Tidak ada proses review berkala yang tercatat.** Jadwal review CSP-001 dan AIP-001 ada di dokumen, tetapi belum ada pengingat atau catatan review.
8. **Sosialisasi staf di luar sistem belum terdokumentasi.**

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Susun, setujui, dan tanda tangani kebijakan keamanan informasi induk (ISMS policy) yang mencakup ruang lingkup AuditSphere dan merujuk ke CSP-001, AIP-001, dan ISO-AS-001. | Lead Information Security | Tinggi | [isi] | Dokumen kebijakan induk bertanda tangan |
| 2 | Simpan salinan CSP-001 dan AIP-001 yang sudah ditandatangani beserta tanggal persetujuan. Isi kolom tanggal di document control. | Security Officer | Tinggi | [isi] | PDF tertandatangan (akses terbatas) |
| 3 | Selesaikan ISO-AS-001: isi tanggal berlaku, tanggal review, dan ubah status dari DRAFT setelah disetujui. | Lead Information Security | Sedang | [isi] | ISO-AS-001 versi ≥1.0 |
| 4 | Validasi versi di server: `Accept` harus mencocokkan `version` dan `content` dengan teks master yang tersimpan, bukan menerima dari klien. | Tim Backend (auth-service) | Tinggi | [isi] | Perubahan kode dan test |
| 5 | Ubah `HasAcceptedLatest` agar mengecek versi agreement aktif. Pengguna dengan versi lama harus diminta menyetujui ulang. | Tim Backend (auth-service) | Tinggi | [isi] | Perubahan kode dan test |
| 6 | Ubah `handleAccept` agar tidak pindah langkah saat request gagal, dan tampilkan pesan kesalahan. | Tim Frontend | Tinggi | [isi] | Perubahan kode dan test |
| 7 | Ambil sampel query `confidentiality_agreements` (jumlah pengguna aktif vs. yang sudah menyetujui) sebagai bukti efektivitas operasional. | Tim Backend | Sedang | [isi] | Hasil query (tanpa data pribadi) |
| 8 | Lakukan dan catat sosialisasi kebijakan ke staf (sesi atau acknowledgement tertulis). | Lead Information Security | Sedang | [isi] | Daftar hadir atau acknowledgement |
| 9 | Isi jadwal review berkala di master data dan buat pengingat (minimal tahunan). | Lead Information Security | Sedang | [isi] | Entri jadwal review |
| 10 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.1-01 | CSP-001 Cloud Security Policy v1.0 | `docs/ISO_27001/cloud_saas/CSP-001_Cloud_Security_Policy.docx` | Internal |
| EV-A5.1-02 | AIP-001 AI Usage Policy v1.0 | `docs/ISO_27001/ai-ml_security/AIP-001_AI_Usage_Policy.docx` | Internal |
| EV-A5.1-03 | Kode acceptance pengguna (halaman, store, endpoint, model) | `frontend/pages/auth/confidentiality.vue`, `frontend/stores/auth.ts`, `backend/auth-service/` | Internal |
| EV-A5.1-04 | Tabel `confidentiality_agreements` (sampel hasil query) | Database auth-service | Restricted |

Setelah tindakan 1 dan 2 selesai, tambahkan ID untuk kebijakan induk dan salinan tertandatangan.
