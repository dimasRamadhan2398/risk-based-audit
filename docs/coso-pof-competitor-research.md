# Riset kompetitor: penilaian COSO 2013 dengan hirarki PoF dan unsur pemenuhan

Dibuat 2026-11-12. Hanya riset, tidak ada kode yang diubah. Pelengkap `docs/coso-2013-component-assessment-plan.md`.

## 0. Batasan riset (baca dulu)

- Pencarian web hanya mengembalikan halaman ringkasan, review pihak ketiga dan halaman vendor. Tidak ada screenshot atau demo produk yang saya lihat langsung. Jadi uraian "bagaimana UI-nya" untuk tiap produk berstatus **dugaan/ulasan pihak ketiga**, kecuali ditandai Terverifikasi.
- Terverifikasi = tertulis di sumber yang saya buka/baca hasilnya. Dugaan = kesimpulan saya dari pola umum.
- Manual kertas kerja BPKP (PDF Pemkot Yogyakarta) tidak terbaca (PDF terkompresi), jadi struktur lembar kerjanya tidak saya verifikasi.
- Tidak ditemukan produk yang menilai seluruh 87 Point of Focus satu per satu dengan bobot dan capaian 0-1 seperti KKE ICON PLUS. Itu temuan penting: pola KKE ini lebih dekat ke kertas kerja Excel (BPKP/AuditNet) daripada ke fitur produk GRC mana pun.

## 1. Temuan per produk

| Produk | Yang terverifikasi | Dugaan / belum terverifikasi | Sumber |
|---|---|---|---|
| Workiva | Template "COSO Testing" (dari AuditNet): langkah uji dan atribut per bagian COSO, bisa diunduh .xlsx. Blog Workiva menyebut 87 PoF. Review pihak ketiga: peta kontrol/risiko ke COSO 2013, AI menandai celah cakupan. | Tampilan UI hirarkinya tidak diketahui. | [template](https://marketplace.workiva.com/en-us/templates/coso-testing), [blog](https://www.workiva.com/blog/unpacking-cosos-new-guidance-internal-control-over-sustainability-reporting-icsr) |
| Diligent HighBond | Pustaka proyek dengan template SOX berbasis COSO 2013 (review). Agregasi defisiensi level program disebut lemah. | Tidak ada template PoF per butir. | [review](https://internalauditguide.com/2026/09/diligent-one-review/), [blog Diligent](https://www.diligent.com/resources/blog/coso-internal-controls-framework) |
| Hyperproof | Template SOX (ICFR + ITGC) memuat kontrol COSO/COBIT yang bisa diedit; satu kontrol dipetakan lintas framework. | Tampilan penilaian PoF tidak ditemukan. | [ringkasan](https://thefinancelead.com/blog/sox-compliance-software/) |
| MetricStream | Uji kontrol berformat survei, unggah sampel via Excel, sub-certification berjenjang. AI Survey Autofill: unggah dokumen, dapat draf jawaban dengan skor keyakinan dan sitasi, diterima satu per satu atau massal; jejak audit membedakan jawaban AI vs manual. Keluhan pengguna: gagal simpan, unggah buruk. | Detail tata letak tidak diketahui. | [SOX](https://www.metricstream.com/products/sox-compliance-management.htm), [review](https://internalauditguide.com/2026/09/metricstream-audit-management-review/) |
| LogicGate | Automated Evidence Testing: lampiran dicek terhadap prosedur, hasil Pass/Fail/Insufficient plus alasan, kesimpulan tetap di auditor. Bulk Link Records. Fitur AI ada di app Controls Compliance/Internal Audit, bukan app SOX. | - | [review](https://internalauditguide.com/2026/09/logicgate-risk-cloud-review/) |
| Archer | Workflow uji kontrol dapat dikonfigurasi, antrean review, penugasan tugas, self-assessment owner dengan form, evidence capture. | Kuat konfigurasi, UX attestation massal tidak terdokumentasi. | [perbandingan](https://internalauditguide.com/2026/09/archer-vs-logicgate/) |
| AuditBoard, Resolver, Riskonnect, CaseWare, TeamMate+, ZenGRC, Pathlock, Fastpath, Vanta/Drata | **Tidak terverifikasi**: pencarian tidak menghasilkan halaman spesifik. Jangan dikutip sebagai rujukan. | Vanta/Drata dugaan: model "kontrol terpetakan ke framework" dengan uji otomatis, bukan penilaian PoF. | - |
| BPKP e-SPIP / Kertas Kerja Maturitas SPIP | 5 unsur dirinci ke 25 subunsur, tiap subunsur 5 parameter (125 parameter, tingkat 0 sampai 5) pada pedoman lama (Perka 4/2016). Aplikasi e-SPIP hanya menampilkan nilai total K/L/D, dikelola pusat; instansi memakai Kertas Kerja Excel untuk nilai per OPD dan sebagai cadangan. Peraturan BPKP 5/2021 menambah MRI, IEPK, kapabilitas APIP. Peran: penilaian mandiri, penjaminan kualitas oleh APIP, evaluasi BPKP. | Warna dan struktur lembar tidak terbaca. | [Perka BPKP 5/2021](https://peraturan.bpk.go.id/Download/298281/Peraturan%20BPKP%20Nomor%205%20Tahun%202021.pdf), [manual Yogyakarta](https://inspektorat.jogjakota.go.id/download/hit/10883/manual-book-mekanisme-pengisian-kk-penilaian-maturitas-spip-10883.pdf), [ringkasan BPKP Jabar](https://inspektorat.kotabogor.go.id/dokumen/Materi_Penilaian_Maturitas_Penyelenggaran_SPIP_Terintegrasi_oleh_BPKP_Perwakilan_Jawa_Barat.pdf) |
| Audithink (lokal) | Mengaku platform audit internal terintegrasi mendukung SPIP, RBIA, manajemen risiko, GCG untuk BUMN/BUMD. | Isi modul penilaian SPIP tidak diverifikasi (hanya blog). | [blog](https://audithink.com/blog/pengendalian-internal/) |
| Template pemerintah kota (contoh COSO worksheet) | Kolom: Points of Focus, Control Activity Focus, Control Type, Audit Testing Procedures, Audit Results, Auditor Notes. | - | [PDF](https://www.cstx.gov/media/k1un1zae/18-02-coso-internal-control-assessment.pdf) |

Catatan angka: Wikipedia menyebut sekitar 80 PoF, Workiva 87; sheet KKE memakai kurang lebih 81 PoF. Gunakan katalog dari sheet user sebagai sumber, bukan angka umum.

## 2. Pola UX anti-overwhelm

Terverifikasi (NN/g, [4 Principles to Reduce Cognitive Load in Forms](https://www.nngroup.com/articles/4-principles-reduce-cognitive-load/)): form panjang dipecah per halaman/bagian, kelompokkan field terkait, tampilkan indikator progres, nyatakan apakah progres tersimpan otomatis. Ringkasan sekunder (bukan halaman NN/g, tidak terverifikasi): lebih dari dua tingkat disclosure menurunkan usability; satu langkah wizard = satu keputusan; ada langkah review sebelum submit.

Pola yang saya sarankan, sebagian adalah dugaan desain, bukan fitur yang terbukti ada di produk di atas:

1. Tampilan default berhenti di PoF; unsur muncul saat PoF dibuka (maksimal 2 tingkat terbuka sekaligus).
2. Progres per komponen dan prinsip ("12/17 PoF dinilai").
3. Filter cepat: Belum dinilai, Kurang (kuning/merah), Butuh bukti, Ditugaskan ke saya.
4. Penugasan per unit/PIC (selaras daftar permintaan dokumen per unit di sheet user).
5. Mode fokus satu PoF dengan tombol Berikutnya.
6. Input cepat: set capaian massal untuk unsur terpilih, template jawaban, salin dari periode lalu (hanya sebagai draf, wajib dikonfirmasi).
7. Auto-isi dari sinyal sistem dengan skor keyakinan dan sitasi, diterima per butir atau massal (pola MetricStream/LogicGate, terverifikasi di review).
8. Warna hanya sebagai ringkasan roll-up (hijau/kuning/merah/abu untuk belum dinilai), selalu disertai angka atau ikon agar tidak bergantung warna.

## 3. Perbandingan pola

| Aspek | Produk GRC (Workiva, MetricStream, LogicGate, Archer) | Kertas kerja BPKP / KKE | Usulan AuditSphere |
|---|---|---|---|
| Granularitas | Kontrol dipetakan ke prinsip; PoF sebagai kriteria/uji | Subunsur + parameter atau PoF + unsur | PoF sebagai satuan default, unsur sebagai detail |
| Tampilan | Survei/form, antrean tugas (dugaan tree) | Spreadsheet datar, ratusan baris | Pohon 2 tingkat, accordion, mode fokus |
| Skor | Pass/Fail, status uji | Capaian x bobot, roll-up | Capaian x bobot (metode KKE) |
| Bukti | Lampiran + uji AI (LogicGate), sitasi (MetricStream) | Kolom penjelasan, nomor dokumen | Bukti per unsur + rujukan dokumen |
| Input cepat | Bulk accept AI, unggah Excel | Salin manual | Bulk set, impor Excel KKE, salin periode lalu |
| Otomasi | AI autofill, uji bukti | Tidak ada | Sinyal dari risiko/temuan/ATR (sudah ada di `coso-signals`) |

## 4. Rekomendasi untuk AuditSphere

### 4.1 Lapisan tampilan

- Level 1 Komponen (5 kartu: skor, bobot, progres, warna).
- Level 2 Prinsip (17 baris, di bawah kartu terpilih).
- Level 3 PoF (default terdalam yang terlihat; capaian turunan, status warna, jumlah unsur belum dinilai).
- Level 4 Unsur (expand/drawer: tipe Present/Function, metode D/K/W/O, capaian, bukti, penjelasan, saran).
- Toggle "Ringkas / Rinci". Ringkas = isi capaian di level PoF saja (fase 1).

### 4.2 Wajib vs opsional

- Wajib: capaian (atau status) per PoF atau per unsur, penjelasan bila capaian di bawah 1 atau 0, PIC.
- Opsional: metode evaluasi, saran, bukti pada capaian penuh (tapi disarankan).
- Capaian di luar kelipatan skala katalog ditolak, kecuali jenis kuesioner (nilai kontinu).

### 4.3 Pemetaan ke yang sudah dibangun

Fakta dari plan dan kode (`backend/risk-service/models/coso.go`, `docs/coso-2013-component-assessment-plan.md` bagian 4): `coso_components` + `coso_principles` (framework, code, bobot belum ada), `ic_assessments` -> `ic_assessment_components` (5) -> `ic_assessment_principles` (17). Level prinsip memakai status `YES/PARTIAL/NO` + severity, bukan capaian numerik berbobot.

Konsekuensi: KKE adalah metode **kuantitatif berbobot**, sedangkan yang ada **kualitatif per prinsip**. Jangan menimpa; tambahkan metode ke header assessment (mis. `method = QUALITATIVE | WEIGHTED_KKE`) dan untuk metode KKE turunkan status prinsip dari persentase capaian (ambang bisa dikonfigurasi, mis. di atas 90% = YES) agar roll-up dan dashboard lama tetap jalan. Ini usulan, perlu keputusan produk.

### 4.4 Usulan model data (additive saja, selaras batasan plan: tidak ada kolom diubah/dihapus)

Katalog (risk-service, per framework, seed insert-if-missing):
- `coso_points_of_focus`: id, framework, principle_code, code (1.1.1), name, weight, sort_order, is_active.
- `coso_pof_elements`: id, pof_id, code (1.1.1.1), description, type (`PRESENT`/`FUNCTION`), weight, default_methods (D/K/W/O), scale_type (`DISCRETE`/`QUESTIONNAIRE`), document_hint.
- Tambah kolom nullable `weight` pada `coso_components` dan `coso_principles`. Bobot hidup di katalog, bukan di jawaban, tapi assessment menyimpan **versi katalog** (`catalog_version`) agar periode lama tidak berubah bila bobot direvisi.
- Invarian divalidasi saat impor katalog: jumlah bobot unsur = bobot PoF, jumlah PoF = prinsip, prinsip = komponen, total = 100.

Jawaban (audit-service, tabel baru):
- `ic_assessment_pofs` (assessment_id, pof_code, achievement turunan, status, pic, due) dan `ic_assessment_elements` (assessment_id, element_code, methods[], achievement numeric(4,3), score turunan, explanation, recommendation, assignee, status `NOT_ASSESSED`/`DRAFT`/`FINAL`, updated_by/at).
- `ic_element_evidence` (element_code, file/ref, doc_number, kind) dan penggunaan ulang `ic_assessment_events` untuk jejak perubahan.
- Skor tidak disimpan sebagai sumber kebenaran; hitung di server agar konsisten (bisa dicache di header).

### 4.5 MVP bertahap

| Fase | Cakupan | Bukti selesai |
|---|---|---|
| 1 | Katalog PoF (impor dari KKE), penilaian di level PoF saja (capaian 0 sampai 1 atau skala diskret), roll-up ke prinsip/komponen/total, progres + filter belum dinilai, ringkasan (setara Lampiran 7-29/30) | Total efektivitas contoh 83,26 tereproduksi dari sheet user |
| 2 | Unsur: expand di PoF, tipe Present/Function, metode D/K/W/O, bukti, penjelasan, saran, bulk set, penugasan PIC per unit, impor Excel, salin periode lalu sebagai draf, daftar permintaan dokumen per unit | Impor ~300 unsur lolos validasi bobot; ekspor Excel cocok dengan format KKE |
| 3 | Auto-signal: sinyal dari `/rcm/coso-signals`, temuan, ATR menjadi **usulan** capaian dengan sitasi dan skor keyakinan, diterima per butir atau massal; wajib konfirmasi auditor (selaras D4-D5 di plan) | Setiap capaian dari sinyal tercatat sebagai usulan di event log |

Urutan serah: `auditsphere-solution-architect` (skema final dan tanggung jawab service), backend (katalog, impor, roll-up, tes), frontend (pohon, mode fokus, filter), i18n (istilah), devops (migrasi per tenant, prefix Kong bila ada route baru), qa (uji roll-up terhadap angka sheet).

### 4.6 Risiko

- Katalog ~81 PoF dan ~300 unsur adalah materi BUMN/klien tertentu; cek lisensi dan kerahasiaan sebelum dijadikan seed bersama untuk semua tenant. Lebih aman sebagai katalog yang dapat diimpor per tenant.
- Dua metode (kualitatif vs berbobot) bisa membingungkan; pilih satu per assessment, jangan campur.
- Bobot berubah antar tahun: perlu versi katalog.
- Auto-signal hanya valid sebagian kecil unsur (tipe Function terkait risiko/temuan/ATR); sisanya tetap D/K/W/O manual. Hindari klaim "otomatis" berlebihan.
- Pembulatan dan akurasi desimal (bobot 0,156 dst.): simpan numeric, bukan float.
- Performa: ~300 baris per assessment per unit, aman bila tidak dirender sekaligus (virtual list/lazy expand).
- Bila impor Excel dipakai, aturan unggah AuditSphere (file tidak dipersist di prod) perlu dicek oleh devops.
- Rilis backend COSO belum dideploy (lihat status di plan); rencana ini bergantung pada deploy tersebut.

## 5. Pertanyaan terbuka

1. Satu metode (KKE berbobot) untuk semua tenant, atau pilihan per tenant?
2. Apakah katalog KKE ICON PLUS boleh dipakai sebagai seed default?
3. Ambang capaian PoF ke status YES/PARTIAL/NO?
4. Siapa yang mengisi unsur: auditor, PIC unit (auditee), atau keduanya dengan review?
