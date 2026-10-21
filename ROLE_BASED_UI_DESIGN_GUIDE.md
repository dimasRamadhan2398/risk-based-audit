# Rekomendasi Tampilan UI Berdasarkan Role
*Selaras dengan Matriks Hak Akses (`frontend/components/settings/permission.vue`)*

---

## Prinsip Utama
Tampilan aplikasi disesuaikan dengan tanggung jawab operasional masing-masing role agar pengguna tidak bingung (*information overload*) dan data sensitif tetap terjaga:
- **Auditor**: Fokus pada **eksekusi penugasan lapangan** dan penyusunan kertas kerja.
- **Auditee**: Fokus pada **tindak lanjut rekomendasi (CAPA)** dan pemenuhan dokumen audit.
- **Manager / Lead**: Fokus pada **approval, review mutu, dan monitoring progres tim**.
- **Viewer**: Fokus pada **laporan eksekutif dan analitik hasil akhir (Read-Only)**.
- **Admin**: Akses penuh termasuk **pengaturan sistem & manajemen pengguna**.

---

## 1. Tampilan Dashboard: Auditee (Unit Kerja yang Diaudit)

> **Landasan Matriks `permission.vue`:**
> - Row 8-1 (`Audit Fieldwork`): `RESPOND`
> - Row 8-2 (`Working Papers F-01 to F-05`): `NONE`
> - Row 9-3 (`Client Satisfaction Survey`): `FILL SURVEY`
> - Row 10 (`Action Taken Report / ATR`): `UPDATE CAPA`
> - Row 3-1 (`Corporate Risk Profile`): `EDIT — Own Dept.`

### Komponen Utama:
1. **Header Profil Unit**: Menampilkan nama departemen/unit kerja yang sedang login.
2. **Widget Dokumen Wajib Diunggah (PBC - *Provided by Client*)**:
   - Menampilkan daftar permintaan bukti/data dari auditor yang statusnya masih *Pending* atau mendekati deadline.
   - Tombol langsung untuk *drag & drop* upload file bukti.
3. **Tabel Utama: Action Taken Report (ATR / CAPA)**:
   - Status tindak lanjut temuan audit (*Open, In-Progress, Overdue, Closed*).
   - Kolom untuk input progres perbaikan dan upload dokumen bukti perbaikan (*evidence*).
4. **Banner Survei Kepuasan (*Satisfaction Survey*)**:
   - Otomatis muncul jika ada penugasan audit yang baru saja selesai (*Result Report* terbit).
5. **Elemen yang Disembunyikan (Hidden)**:
   - ✗ Seluruh Kertas Kerja Audit internal (KKA F-01 s/d F-05).
   - ✗ Statistik risiko makro korporat (hanya tampil profil risiko unitnya sendiri).
   - ✗ Menu Rencana Strategis, Audit Universe, dan KPI internal audit.

---

## 2. Tampilan Dashboard: Auditor (Pelaksana Audit)

> **Landasan Matriks `permission.vue`:**
> - Row 8-1 (`Audit Fieldwork`): `FULL`
> - Row 8-2 (`Working Papers F-01 to F-05`): `CREATE`
> - Row 8-3 (`Import Working Papers`): `UPLOAD`
> - Row 9-1 (`Result Report / LHA`): `DRAFT`
> - Row 3-3 (`Risk Factors & Scoring`): `READ`

### Komponen Utama:
1. **Tabel Utama: Daftar Audit Berjalan (*Active Fieldwork*)**:
   - Menampilkan penugasan audit yang sedang dipegang, timeline tahapan, dan status audit lapangan.
2. **Kertas Kerja & Template (KKA F-01 s/d F-05)**:
   - Akses cepat untuk membuat atau melanjutkan kertas kerja pengujian/sampling.
   - Tombol upload/import kertas kerja pendukung.
3. **Widget Draft Temuan & LHA**:
   - Ringkasan temuan sementara yang sedang disusun sebelum finalisasi laporan.
4. **Elemen yang Disederhanakan / Dihilangkan**:
   - ✗ Widget agregat risiko makro korporasi dihilangkan/disederhanakan agar layar kerja fokus ke penugasan audit aktif.
   - ✗ Tombol publish/otorisasi akhir laporan (karena wewenang Manager/CAE).

---

## 3. Tampilan Dashboard: Manager / Team Lead / CAE

> **Landasan Matriks `permission.vue`:**
> - Row 8-2 (`Working Papers`): `REVIEW`
> - Row 9-1 (`Result Report / LHA`): `FULL`
> - Row 10 (`Action Taken Report / ATR`): `REVIEW`
> - Row 5-1 (`Annual Audit Plan`): `DRAFT`

### Komponen Utama:
1. **Kotak Masuk Persetujuan (*Approval Inbox*)**:
   - Notifikasi & daftar review KKA dari auditor.
   - Approval draft LHA sebelum diterbitkan ke Auditee/Direksi.
   - Review efektivitas bukti tindak lanjut CAPA dari Auditee.
2. **Monitoring Rencana Audit Tahunan (PKPT / Annual Plan)**:
   - Realisasi pelaksanaan audit (Persentase audit terencana vs berjalan vs selesai).
3. **Ringkasan Temuan Signifikan (*High Risk Findings*)**:
   - Grafik temuan berisiko tinggi di seluruh departemen.

---

## 4. Tampilan Role Lainnya

- **Viewer (Direksi / Komite Audit)**:
  - Tampilan dashboard eksekutif berbasis grafik (ringkasan status audit, rasio temuan tertutup/terbuka, dan laporan LHA final). Sifatnya *Read-Only* tanpa tombol edit/import/upload.
- **Admin**:
  - Akses penuh ke seluruh menu plus menu eksklusif: **Settings & User Management**.

---

## Ringkasan Hak Akses Navigasi & Menu Sidebar

| Menu / Fitur | Auditee | Auditor | Manager | Viewer | Admin |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Dashboard** | Khusus CAPA & PBC | Khusus Audit Aktif | Khusus Approval | Eksekutif Read-Only | Full |
| **Audit Fieldwork (PBC)** | Respond (Upload) | Full | Full | Read | Full |
| **Working Papers (KKA)** | ✗ Hidden | Create & Edit | Review | Read | Full |
| **LHA / Hasil Audit** | Read | Draft | Full / Approve | Read | Full |
| **Survei Kepuasan** | Fill Survey | Read | Read | Read | Full |
| **Action Taken Report (ATR)** | Update CAPA | Read | Review | Read | Full |
| **Pengaturan & User** | ✗ Hidden | ✗ Hidden | ✗ Hidden | ✗ Hidden | Full |
