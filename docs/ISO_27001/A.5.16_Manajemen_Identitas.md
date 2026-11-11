# A.5.16 — Manajemen Identitas

| Field | Isi |
| :--- | :--- |
| **Tema** | Organizational |
| **Judul kontrol** | Manajemen identitas |
| **Fokus pemeriksaan** | Siklus hidup identitas dikelola dari pembuatan hingga penghapusan |
| **Contoh bukti objektif** | Prosedur IAM, joiner-mover-leaver log, daftar akun |
| **Sistem** | AuditSphere (auditsphere.app), VPS 202.10.34.166 & Cloud SaaS |
| **PIC** | Qiko (tercatat di checklist; konfirmasi peran [isi]) |
| **Reviewer** | [isi saat validasi] |
| **Tanggal validasi** | [isi saat validasi] |

> Ringkasan internal untuk checklist kesiapan bukti. Bukan pengganti Statement of Applicability (SoA) resmi. Isi diverifikasi terhadap repositori (cabang `staging`) pada 29 Oktober 2026; periksa ulang sebelum dipakai sebagai bukti.

---

## Ringkasan Nilai untuk Checklist

| Kolom checklist | Nilai | Catatan singkat |
| :--- | :--- | :--- |
| Relevansi SoA | **Berlaku** | AuditSphere menyimpan akun pengguna internal dan auditee, hubungan akun ke data karyawan, dan hak akses ke kertas kerja serta bukti klien. Siklus hidup akun harus dikendalikan. |
| Implementasi status | **Sebagian** | Pembuatan, perubahan profil, penonaktifan (`is_active`), penghapusan (soft delete) dan reset password oleh admin sudah ada. Pemindahan peran (mover) lewat API, alur leaver otomatis, pencabutan sesi, penguncian akun, dan kebijakan password di ISO-AS-001 belum ada. |
| Ketersediaan bukti | **Sebagian** | Kode dan rute dapat ditunjukkan. Belum ada prosedur IAM tertulis, log joiner-mover-leaver, maupun daftar akun yang diverifikasi. |
| Evidence ID / URL | **Belum terdaftar** (usulan: EV-A5.16-01 s/d EV-A5.16-04) | Lihat bagian Evidence di bawah. ID perlu didaftarkan di Master Audit Evidence Index. |
| Efektivitas operasional | **Sebagian efektif / belum teruji** | Akun yang dinonaktifkan tidak bisa login baru, tetapi sesi yang sudah ada tidak dicabut. Akun default dengan password tetap masih ada di seeder. Tidak ada catatan operasional yang membuktikan proses berjalan. |
| Desain kontrol | **Perlu perbaikan** | Tidak ada titik kontrol untuk leaver dan mover. Kebijakan di ISO-AS-001 bagian 8.2 dan 8.3 belum sesuai dengan kode. |
| Status kesiapan | **Belum siap** | Penerapan dan bukti belum memenuhi seluruh gate. |
| Prioritas | **Tinggi** | Akun default dengan password yang sama (`password123`) dan sesi yang tidak dicabut saat akun dinonaktifkan adalah risiko langsung pada data audit klien. |
| Review berikutnya | [isi] | Usulan: 3 bulan setelah prosedur IAM disetujui, dan setiap kali alur akun atau peran berubah. Review akses triwulanan yang diminta ISO-AS-001 bagian 8.3 belum berjalan. |

> Catatan pembacaan checklist: pada tab `Checklist 93 Kontrol`, baris A.5.16 (baris 22) kolom G s/d Z masih kosong. Nilai "Belum ditentukan" hanya muncul sebagai pilihan di tab `Master Data` dan nilai "Qiko" muncul sebagai PIC pada baris kontrol lain. Nilai "Belum ditentukan" tidak dipakai sebagai status di dokumen ini. Status diisi dari verifikasi kode. PIC "Qiko" diambil dari checklist sebagai usulan dan perlu dikonfirmasi.

---

## 1. Justifikasi Relevansi (Berlaku)

Kontrol A.5.16 **berlaku** untuk AuditSphere karena:

1. Setiap pengguna (CAE, auditor, auditee, viewer, admin tenant) masuk dengan akun `users` di `backend/auth-service/models/user.go`. Akun ini menentukan hak akses ke modul audit, dan akun yang tidak dicabut dapat membaca data audit klien.
2. Akun karyawan dihubungkan ke data master lewat kunci bisnis (`users.employee_id` = `employees.employee_code`), bukan foreign key (`docs/org-structure-model-plan.md`, bagian 4.6 "Cross-service integrity", baris 190). Perubahan di data karyawan karena mutasi atau resign tidak otomatis mengubah akun.
3. Klien enterprise pada due diligence akan meminta bukti alur joiner, mover, dan leaver. ISO-AS-001 bagian 8.3 mewajibkan pencabutan akses segera dan review triwulanan.

---

## 2. Implementasi yang Sudah Ada

Setiap butir diverifikasi dari kode. Path dan baris mengacu pada cabang `staging`.

### 2.1 Joiner: pembuatan akun

| Langkah | Implementasi | Bukti kode |
| :--- | :--- | :--- |
| Pembuatan akun oleh admin | `POST /api/v1/users` hanya untuk role `ADMIN` (`RequireRoles("ADMIN")`). Akun wajib punya minimal satu role; jika tidak ada role yang cocok, akun dihapus lagi. | `backend/auth-service/routes/registry.go:181-187`; `backend/auth-service/services/user/user_service.go:124-212` |
| Password awal | Admin mengisi password (minimal 6 karakter) atau memilih `generate_password`. Password sementara dikembalikan sekali dalam respons, tidak dicatat di log, dan `must_change_password` diset `true`. | `user_service.go:138-143`, `user_service.go:182-200`, `user_service.go:211-222` |
| Paksa ganti password | Middleware Nuxt mengarahkan pengguna dengan `mustChangePassword` ke `/auth/change-password`. | `frontend/middleware/01-force-password-change.global.ts` |
| Hubungan ke karyawan | `POST /users/:id/link-employee` dan `GET /users/lookup` (ADMIN). Satu karyawan hanya boleh punya satu akun. | `registry.go:190-193`; `user_service.go:231-258` |
| Joiner otomatis (hanya email `@aifl.co.id`) | Saat karyawan dibuat di master-service, akun login dibuat otomatis dengan role `AUDITEE`. Jika gagal, data karyawan dibatalkan lagi. | `backend/master-service/services/employee/login_account.go:19-22`, `:58-120` |

### 2.2 Mover: perubahan profil dan peran

| Langkah | Implementasi | Bukti kode |
| :--- | :--- | :--- |
| Ubah profil | `PUT /users/:id` untuk diri sendiri atau ADMIN. Field: nama, telepon, departemen, jabatan, `is_active`, avatar. | `registry.go:180`; `user_service.go:303-346`; `backend/auth-service/models/dto.go:98-107` |
| Ganti role | Hanya saat pembuatan akun (`AssignRoles`). Tidak ada rute untuk mengubah role akun yang sudah ada. | `backend/auth-service/repositories/user_repository.go:225-238`; `user_service.go:204` |
| Ganti password oleh pengguna | `ChangePassword` memeriksa password lama dan melarang password baru yang sama dengan yang lama. Minimal 6 karakter. | `backend/auth-service/services/auth/auth_service.go:433-466`; `models/dto.go:62` |
| Reset password oleh admin | Password sementara acak. Akun ADMIN lain tidak bisa di-reset. Admin tidak bisa mereset dirinya sendiri. Kejadian dicatat lewat Kafka. | `user_service.go:435-495`; `backend/auth-service/controllers/user/user_controller.go:397` |

### 2.3 Leaver: penonaktifan dan penghapusan

| Langkah | Implementasi | Bukti kode |
| :--- | :--- | :--- |
| Nonaktifkan akun | `PUT /users/:id` dengan `is_active=false` (ADMIN). Login ditolak untuk akun nonaktif. | `user_service.go:327-329`; `backend/auth-service/services/auth/auth_service.go:91-94`; `backend/auth-service/services/auth/step_up_service.go:146` |
| Hapus akun | `DELETE /users/:id` (ADMIN). Ini soft delete karena model `User` memakai `gorm.DeletedAt`. | `registry.go:187`; `user_service.go:349-357`; `backend/auth-service/models/user.go:34`; `backend/auth-service/repositories/base_repository.go:54` |
| Cabut sesi aktif | Login menyimpan token di Redis dengan kunci `token:<jwt>` dan masa berlaku sama dengan JWT. Logout menghapus kunci ini. Middleware memeriksa kunci tersebut, dan pemeriksaan ini aktif di `serve.go`. | `auth_service.go:186-189`, `auth_service.go:392-395`; `backend/auth-service/pkg/middleware/auth.go:83-91`; `backend/auth-service/cmd/auth/serve.go:121`, `:199-204` |
| Pencatatan | Penghapusan dan perubahan profil hanya tercatat di log aplikasi (zap). Login, logout, perubahan password, dan reset password dikirim ke Kafka. | `user_service.go:219`, `:340`, `:356`; `auth_service.go:194-200`, `:400-405`, `:465-470`; `user_service.go:477-489` |

### 2.4 Pengendalian pendukung

- Rute daftar pengguna, pembuatan, penghapusan, reset password, dan link karyawan dibatasi ke ADMIN (`backend/auth-service/routes/registry.go:181-193`). Pembatasan ini diuji dalam `routes/viewer_readonly_test.go`.
- Rute `GET /users/:id` dan `PUT /users/:id` dibatasi ke pemilik akun atau ADMIN (`selfOrAdmin`).
- Hak akses per peran dan per fitur dicatat di seeder (`backend/auth-service/pkg/database/seeders/seeder.go`, RBAC matrix). Penegakan di sisi server dibahas di A.5.15.

---

## 3. Ketersediaan Bukti

| Bukti yang diharapkan | Status | Keterangan |
| :--- | :--- | :--- |
| Prosedur IAM (tertulis, disetujui) | **Belum tersedia** | Tidak ada dokumen IAM di `docs/ISO_27001/`. ISO-AS-001 bagian 8 menjelaskan kebijakan, tetapi belum disetujui dan belum merujuk ke prosedur operasional. |
| Log joiner-mover-leaver | **Belum tersedia** | Tidak ada tabel atau log khusus untuk perubahan akun. Log hanya di zap dan sebagian di Kafka (lihat 4.4). |
| Daftar akun (inventaris) | **Sebagian** | Dapat diekspor dengan `GET /users` (ADMIN), tetapi belum ada tinjauan yang terdokumentasi. Akun seeder tidak diberi tanda. |
| Bukti review akses triwulanan | **Belum tersedia** | ISO-AS-001 bagian 8.3 mewajibkan review triwulanan. Tidak ditemukan catatan review. |
| Bukti pencabutan akun leaver | **Belum tersedia** | Tidak ada catatan tindakan, dan alur leaver otomatis tidak ada (lihat 4.2). |
| Bukti penolakan login akun nonaktif | **Sebagian** | Kode ada (`auth_service.go:91-94`). Belum ada hasil uji atau log operasional. |
| Bukti reset password admin | **Sebagian** | Kode dan rute ada. Log Kafka `Password reset by admin` ada di `user_service.go:477-489`. Belum ada contoh keluaran. |

---

## 4. Catatan / Gap Terverifikasi

1. **Leaver tidak otomatis dan sesi tidak dicabut.** Nonaktifkan atau hapus akun (`user_service.go:327`, `:349`) tidak menghapus kunci `token:<jwt>` di Redis dan tidak memanggil `InvalidateUserTokens`. Token yang sudah ada tetap valid sampai masa berlakunya habis. Default masa berlaku 24 jam (`backend/auth-service/pkg/config/config.go:176`, bisa diubah lewat konfigurasi). Akibatnya, pencabutan akses belum langsung berlaku.
2. **Akun karyawan tidak mengikuti status karyawan.** Mengubah `is_active` atau menghapus karyawan di master-service (`backend/master-service/services/employee/employee.go:269-283`, `:440`) tidak menyentuh akun di auth-service. `AccountClient` hanya punya `LinkEmployee` dan pembuatan akun (`login_account.go:28-34`). Akun karyawan yang sudah mutasi atau resign tetap aktif.
3. **Tidak ada mover lewat API.** Perubahan role untuk akun yang sudah ada tidak punya rute (lihat 2.2). Admin harus menghapus akun lalu membuat ulang, dan riwayatnya hilang.
4. **Akun default dengan password tetap.** `backend/auth-service/pkg/database/seeders/seeder.go:223-300` membuat akun `admin`, `auditor`, `dept_head`, `new_dept_head`, `auditee`, dan `viewer` dengan password `password123`. Akun ini tidak diberi `must_change_password`. `scripts/onboard-tenant.sh` (baris 189) menjalankan `auth seed` untuk setiap tenant baru. Seeder hanya membuat akun jika username belum ada, tetapi di tenant baru akun ini langsung aktif. Password yang sama juga tertulis di `backend/auth-service/README.md:177,190` dan `backend/auth-service/scripts/test-api.sh`.
5. **Kebijakan ISO-AS-001 bagian 8.2 dan 8.3 belum diterapkan.** Dokumen menyebut password 12 karakter, masa berlaku 90 hari, riwayat 5 password, lockout 5 kali gagal selama 30 menit, idle timeout 15 menit, 1 sesi aktif per akun, MFA wajib untuk ADMIN/CAE/Audit Manager, dan nonaktif otomatis setelah 60 hari tidak aktif. Hasil pemeriksaan kode:
   - Password minimal 6 karakter (`models/dto.go:12`, `:62`; `user_service.go:142`). Tidak ada aturan kompleksitas, masa berlaku, atau riwayat.
   - `LockedUntil` ada di model (`models/user.go:28`) dan hanya dibersihkan saat reset (`user_service.go:469`). Tidak ada kode yang mengisinya. `TrackLoginAttempt` ada di `repositories/cache_repository.go:206` tetapi tidak dipanggil. Lockout tidak berjalan.
   - Tidak ditemukan logika idle timeout di `frontend/` (pencarian `idle` dan `inactivity` kosong).
   - Tidak ditemukan batas satu sesi aktif. Setiap login menyimpan token baru.
   - MFA ada (`auth_service.go:118-128`), tetapi tidak diwajibkan per role.
   - Tidak ada job terjadwal untuk nonaktif otomatis setelah 60 hari (tidak ada `cron` atau `ticker` di `backend/auth-service`).
   Dokumen ISO-AS-001 bagian 8 dan tabel A.5.15 s/d A.5.18 bagian 7 (*Implemented*) perlu dikoreksi.
6. **Lifecycle akun tidak masuk Kafka.** Pembuatan, perubahan `is_active`, penghapusan, dan perubahan role tidak dikirim sebagai event (`user_service.go:157-230`, `:303-357`). Event `account.deactivated`, `account.locked`, dan `password.reset_requested` didefinisikan di `backend/master-service/pkg/kafka/events.go:24-36` tetapi tidak dipanggil dari auth-service. Event `role.assigned` dikirim saat peran dibuat dan saat izin peran diubah (`backend/auth-service/services/role/role_service.go:68`, `:194`), bukan saat peran diberikan ke pengguna.
7. **Tabel `user_access_audit_logs` bukan log audit aplikasi.** Entri ini ada di katalog data demo `backend/master-service/routes/routes.go:959-968` dengan jumlah baris tetap (894200). Entri ini tidak boleh dipakai sebagai bukti.
8. **Reset password mandiri belum ada.** Fungsi penyimpanan token reset (`cache_repository.go:304-313`) tidak dipanggil oleh layanan mana pun. Reset hanya lewat admin. Ini mengurangi permukaan serangan, tetapi pengguna yang lupa password harus bergantung pada admin.
9. **Joiner otomatis memakai domain dan role tetap.** Email `@aifl.co.id` selalu mendapat role `AUDITEE` (`login_account.go:19-22`). Konfigurasi ini tidak bisa diubah per tenant.

---

## 5. Tindakan Perbaikan

| No | Tindakan | PIC | Prioritas | Target | Bukti yang dihasilkan |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | Hapus atau paksa ganti akun seeder `password123` di setiap tenant, dan hentikan penyebaran password default di README dan skrip. Pastikan `auth seed` tidak membuat akun dengan password tetap di produksi. | Tim Backend (auth-service) bersama auditsphere-devops | Tinggi | [isi] | Perubahan kode dan hasil pemeriksaan tanpa akun default |
| 2 | Cabut sesi saat akun dinonaktifkan atau dihapus: hapus kunci `token:*` milik user, atau gunakan `InvalidateUserTokens` di `UpdateUser` dan `DeleteUser`. | Tim Backend (auth-service) | Tinggi | [isi] | Perubahan kode, uji bahwa token lama ditolak setelah nonaktif |
| 3 | Tetapkan prosedur IAM tertulis: joiner (siapa membuat akun, dasar persetujuan), mover (perubahan peran wajib lewat tiket), leaver (batas waktu pencabutan, misalnya 1 hari kerja). Sahkan oleh pemilik proses. | Lead Information Security bersama HR dan CAE | Tinggi | [isi] | Prosedur IAM bertanda tangan |
| 4 | Sinkronkan status karyawan dengan akun: saat karyawan dinonaktifkan atau dihapus di master-service, nonaktifkan akun terkait lewat event atau panggilan auth-service (lihat 4.2). | Tim Backend (master-service dan auth-service) | Tinggi | [isi] | Perubahan kode dan uji integrasi |
| 5 | Sediakan rute perubahan peran untuk akun yang sudah ada (ADMIN), dengan catatan perubahan. | Tim Backend (auth-service) bersama auditsphere-solution-architect | Sedang | [isi] | Rute dan catatan perubahan peran |
| 6 | Kirim event lifecycle akun (create, update, deactivate, delete, role change) ke Kafka dan simpan sebagai log audit yang tidak bisa diubah. Hentikan penggunaan entri `user_access_audit_logs` sebagai bukti. | Tim Backend (auth-service dan master-service) | Tinggi | [isi] | Contoh event dan log joiner-mover-leaver |
| 7 | Terapkan atau koreksi kebijakan password, lockout, dan idle timeout di ISO-AS-001 bagian 8.2. Putuskan mana yang dipenuhi sekarang dan mana yang ditunda. | Lead Information Security | Tinggi | [isi] | Keputusan kebijakan dan perubahan kode |
| 8 | Buat daftar akun dan review akses triwulanan (ADMIN bersama pimpinan audit), dengan catatan hasil. | Security Officer bersama Pimpinan Tim Audit | Sedang | [isi] | Daftar akun dan berita acara review |
| 9 | Perbarui ISO-AS-001 bagian 8 dan tabel A.5.15 s/d A.5.18 agar status sesuai kode. | Lead Information Security | Sedang | [isi] | Revisi ISO-AS-001 |
| 10 | Jadikan domain dan role joiner otomatis dapat diatur per tenant. | Tim Backend (master-service) | Rendah | [isi] | Konfigurasi tenant dan uji |
| 11 | Pertimbangkan reset password mandiri lewat token yang sudah disiapkan, atau tetapkan bahwa reset hanya lewat admin. | Lead Information Security bersama Tim Backend | Rendah | [isi] | Keputusan dan (jika dipilih) fitur reset |
| 12 | Daftarkan Evidence ID untuk setiap bukti di Master Audit Evidence Index. | Lead Information Security | Sedang | [isi] | Entri Evidence Index |

---

## 6. Evidence (usulan, belum terdaftar)

| Evidence ID (usulan) | Bukti | Lokasi | Akses |
| :--- | :--- | :--- | :--- |
| EV-A5.16-01 | Siklus akun di auth-service: create, update, delete, link karyawan, dan reset password | `backend/auth-service/services/user/user_service.go`, `backend/auth-service/controllers/user/user_controller.go`, `backend/auth-service/routes/registry.go` | Internal |
| EV-A5.16-02 | Alur login akun nonaktif dan pencabutan sesi (Redis `token:*`) | `backend/auth-service/services/auth/auth_service.go`, `backend/auth-service/pkg/middleware/auth.go`, `backend/auth-service/cmd/auth/serve.go` | Internal |
| EV-A5.16-03 | Akun seeder dan default password (temuan, perlu dihapus) | `backend/auth-service/pkg/database/seeders/seeder.go`, `scripts/onboard-tenant.sh` | Restricted |
| EV-A5.16-04 | Prosedur IAM, joiner-mover-leaver log, dan daftar akun (belum ada) | Belum tersedia. Akan dibuat dari tindakan 3, 6, dan 8. | Restricted |

Setelah tindakan 1, 2, dan 3 selesai, tambahkan ID untuk prosedur IAM, log joiner-mover-leaver, dan review akses.
