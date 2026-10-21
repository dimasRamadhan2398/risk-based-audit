# Runbook Deploy — Executive Summary (narasi periode + statistik kompilasi)

Dokumen ini untuk agen dan untuk operator manusia. Setiap langkah diberi label:

- **[AGEN]**: boleh dikerjakan agen sendiri (lokal, read-only, atau dry-run).
- **[KONFIRMASI]**: agen berhenti dan minta persetujuan user sebelum menjalankan.
- **[MANUAL]**: dikerjakan user sendiri (atau agen kalau user menyuruh secara eksplisit).

Kalau ada langkah yang hasilnya tidak sesuai dokumen ini, **berhenti dan laporkan**, jangan improvisasi di prod.

---

## 1. Isi rilis

| Perubahan | Service | Efek ke user |
|---|---|---|
| Narasi executive summary diawali periode. Individual: "Periode Januari 2026, SPI telah menerbitkan …" (dari tanggal surat tugas LHA). Kompilasi: "Periode Kuartal 4 2026, SPI telah menerbitkan …" | frontend | Narasi baru ikut bahasa UI (sebelumnya selalu bahasa Indonesia) |
| Fix "Failed to load source data" di statistik kompilasi. Endpoint pindah dari `/api/v1/executive-summary/compilation-stats` (tidak di-route Kong → 404) ke `/api/v1/executive-summaries/compilation-stats` | audit-service + frontend | Statistik kompilasi termuat lagi |
| Ikut terbawa di audit-service (kode saling bergantung, tidak bisa dipisah): `pkg/serviceurl` baru, `masterclient`/`authclient` meneruskan token JWT, nama perusahaan yang tidak ditemukan jadi kosong (bukan lagi "PT AIFL Indonesia") | audit-service | Nama perusahaan di DOCX LHA sekarang dari master-service; kalau tidak ketemu, kosong |

Yang **tidak** perlu di-deploy: Kong (path baru sudah tercakup prefix `/api/v1/executive-summaries` di semua config), auth, master, risk, analytics, python-ai.

## 2. Aturan yang tidak boleh dilanggar

1. **Jangan push ke `main`.** `origin/main` masih punya workflow lama yang auto-deploy dengan `--force-recreate` dan menjalankan seeder (menimpa data prod dan dulu menghapus upload).
2. **Jangan `docker compose up` seluruh stack**, jangan `down`, `--remove-orphans`, atau `prune` di VPS.
3. **Backend prod hanya lewat `backend/deploy-prod.sh`** (migrate-only, backup, rollback otomatis). Jangan jalankan compose backend tanpa override `/root/auditsphere-deploy/docker-compose.prod.noseed.yml`.
4. **Jangan pakai `frontend/deploy-prod.sh` untuk prod.** Script itu menulis URL API absolut `https://api.auditsphere.app/...` dan menjalankan `compose up --force-recreate`, padahal container prod memakai URL relatif `/api/v1`. Pakai prosedur di langkah 6.
5. **Jangan build Nuxt di VPS** (RAM 3.8 GB, swap sudah terpakai).
6. **Jangan sentuh** container `*_dev`, stack tenant `rb_audit_tenant_bai`, postgres, redis, kafka, kong.
7. SSH ke VPS bisa terblokir sementara kalau terlalu banyak koneksi cepat — gabungkan perintah dalam satu sesi.

VPS: `ssh root@202.10.34.166` · repo di server: `/app/rbia-repo` · situs: https://auditsphere.app

## 3. Persiapan commit — [MANUAL] atau [KONFIRMASI]

Script deploy membangun dari `git archive <ref>`, **bukan** dari working tree. Perubahan harus sudah di-commit dan di-push ke `origin/staging`.

Per 2026-10-22, working tree berisi banyak perubahan lain yang belum di-commit (risk-service scoring, tenant compose, dokumen, dll). Commit hanya file berikut untuk rilis ini:

**audit-service** (harus satu paket, kalau sebagian saja build gagal karena signature `ResolveCompanyName` berubah):
```
backend/audit-service/routes/handler.go
backend/audit-service/routes/executive_summary_routes_test.go
backend/audit-service/controllers/executive_summary/handler.go
backend/audit-service/pkg/serviceurl/
backend/audit-service/pkg/masterclient/client.go
backend/audit-service/pkg/masterclient/client_test.go
backend/audit-service/pkg/authclient/client.go
backend/audit-service/pkg/authclient/baseurl_test.go
backend/audit-service/pkg/docxbuilder/docx_generator.go
backend/audit-service/pkg/docxbuilder/docx_company_name_test.go
```

**frontend**:
```
frontend/utils/executiveSummaryNarrative.ts
frontend/utils/executiveSummaryCompilation.ts
frontend/stores/executive-summary.ts
frontend/composables/useExecutiveSummarySources.ts
frontend/components/audit-result-report/ExecutiveSummaryIndividualForm.vue
frontend/components/audit-result-report/ExecutiveSummaryCompilationForm.vue
frontend/locales/en/common.json
frontend/locales/id/common.json
frontend/tests/unit/executiveSummaryNarrative.test.ts
frontend/tests/unit/executiveSummarySources.test.ts
```

Catatan:
- `frontend/locales/*/common.json` juga berisi perubahan lain (penghapusan emoji, key LHA baru). Kalau mau dipisah, pakai `git add -p` (manual, interaktif).
- `staging` lokal sudah **11 commit di depan** `origin/staging` (termasuk merge `upstream/main` dari Dimas: dashboard auditor/auditee, fix dark mode LHA, dll). Push `staging` berarti semua itu ikut ter-deploy di frontend. Pastikan user setuju.

```bash
git add <file-file di atas>
git commit -m "fix(executive-summary): period prefix in narrative, route compilation-stats via Kong"
git push origin staging          # BUKAN main
git rev-parse --short origin/staging   # catat SHA ini, dipakai di semua langkah berikut
```

## 4. Verifikasi lokal sebelum deploy — [AGEN]

```bash
cd backend/audit-service && go build ./... && go test -count=1 ./routes/ ./controllers/executive_summary/ ./pkg/...
cd frontend && npx vitest run tests/unit/executiveSummaryNarrative.test.ts tests/unit/executiveSummarySources.test.ts tests/unit/executiveSummaryRiskBuckets.test.ts
```

Diketahui (bukan regresi rilis ini): `tests/unit/executiveSummaryFeatures.test.ts` gagal 5 test (label bahasa Indonesia yang sudah pindah ke locale, satu cek RBAC), dan `nuxi typecheck` punya baseline 82 error. Rilis ini lolos kalau angka itu tidak bertambah.

## 5. Deploy audit-service (backend) — HARUS sebelum frontend

Frontend baru memanggil path baru. Kalau frontend duluan, backend lama mengartikan `compilation-stats` sebagai `/:id` dan form error.

**5a. Dry-run — [AGEN]**
```bash
backend/deploy-prod.sh --ref <SHA> --dry-run audit
```
Pastikan yang dipilih hanya `audit`, ref = SHA dari langkah 3, dan tidak ada service lain yang tersentuh.

**5b. Deploy — [KONFIRMASI]**
```bash
backend/deploy-prod.sh --ref <SHA> audit
```
Script akan: `pg_dumpall` backup, menyalin `/root/uploads` keluar-masuk, `up -d --no-deps --build audit-service` dengan override migrate-only, health check, rollback otomatis kalau gagal, dan membandingkan jumlah baris. Kalau script melaporkan FAILED, baca alasannya dulu (pernah gagal hanya karena container lain seperti `rbia_loki` crash-loop).

Alternatif lewat GitHub: Actions → "AuditSphere Production Backend Deploy" → `ref=<SHA>`, `services=audit` (manual, butuh approval environment `production`).

**5c. Smoke test — [AGEN]**
```bash
# lewat Kong: harus 401 dari audit-service (bukan 404 "no Route matched")
curl -s -o /dev/null -w '%{http_code}\n' https://auditsphere.app/api/v1/executive-summaries/compilation-stats?year=2026\&quarter=4
curl -s https://auditsphere.app/api/v1/executive-summaries/compilation-stats?year=2026\&quarter=4
```
- `401` / `Missing authorization header` → route sampai ke audit-service. ✅
- `404 no Route matched with those values` → Kong tidak meneruskan. Berhenti.
- `404` lain / `invalid id` → backend lama masih jalan. Berhenti.

## 6. Deploy frontend (manual, tanpa `frontend/deploy-prod.sh`)

Container `rbia-frontend-prod` dibuat manual (bukan dikelola compose). Image hanya `COPY .output`, jadi `.output` dibangun di Mac.

**6a. Build lokal — [AGEN]**
```bash
REF=<SHA>
WORK=$(mktemp -d)
git archive "$REF" frontend | tar -x -C "$WORK"
docker run --rm -v "$WORK/frontend":/app -w /app node:20-alpine \
  sh -c "npm ci && npx nuxi build --dotenv .env.production"

# Wajib lolos dua cek ini:
grep -rhoE '"apiBase": ?"[^"]*"' "$WORK/frontend/.output/server" | sort -u   # harus "/api/v1"
grep -rl 'api.auditsphere.app' "$WORK/frontend/.output/server" && echo "SALAH: URL absolut ter-bake" || echo OK
```
Build di container Linux supaya binary `sharp` cocok dengan server (linux-x64/musl). Jangan build dengan `frontend/.env` lokal.

**6b. Upload & ganti container — [MANUAL] atau [KONFIRMASI]**
```bash
rsync -a --delete "$WORK/frontend/.output/" root@202.10.34.166:/app/rbia-repo/frontend/.output/
```
Lalu dalam **satu** sesi SSH:
```bash
cd /app/rbia-repo/frontend
docker compose -f docker-compose.prod.yml build            # image saja, JANGAN up
IMG=$(docker compose -f docker-compose.prod.yml config --images)

# simpan chunk lama supaya tab browser yang masih terbuka tidak rusak
cp -rn .output/public/_nuxt/. /var/www/auditsphere-chunk-archive/_nuxt/

docker rm -f rbia-frontend-prod-old 2>/dev/null || true
docker rename rbia-frontend-prod rbia-frontend-prod-old
docker stop rbia-frontend-prod-old
docker update --restart=no rbia-frontend-prod-old
docker run -d --name rbia-frontend-prod --network rb_audit_network -p 3000:3000 \
  --restart unless-stopped --env-file /root/rbia-frontend-prod.env "$IMG"

sleep 5; curl -s -o /dev/null -w '%{http_code}\n' http://localhost:3000/
```
`/root/rbia-frontend-prod.env` harus berformat `KEY=VALUE` (isi = `frontend/.env.production` + `HOST`/`PORT`), bukan dump JSON `docker inspect`.

**Rollback frontend:**
```bash
docker rm -f rbia-frontend-prod && docker rename rbia-frontend-prod-old rbia-frontend-prod \
  && docker update --restart=unless-stopped rbia-frontend-prod && docker start rbia-frontend-prod
```
Hapus `rbia-frontend-prod-old` hanya setelah user mengonfirmasi rilis aman.

## 7. Cek di browser — [MANUAL]

Login ke https://auditsphere.app → Executive Summary → buat baru (hard refresh dulu):

**Individual**
1. Pilih LHA dengan surat tugas satu bulan → narasi diawali "Periode Januari 2026, SPI telah menerbitkan …".
2. LHA lintas bulan → "Periode Januari – Maret 2026"; lintas tahun → "Periode Desember 2025 – Januari 2026".
3. LHA tanpa tanggal → narasi tanpa awalan, tidak ada teks "undefined".

**Kompilasi**
4. Statistik kompilasi termuat untuk kuartal default; tidak ada "Failed to load source data"; "Last synced" muncul.
5. Ganti kuartal/tahun → statistik dimuat ulang dan awalan berubah ("Periode Kuartal 3 2026, …").
6. Ketik teks sendiri, lalu ganti kuartal → teks tetap. Tombol "Reset to Default Template" memakai periode yang dipilih.
7. Ganti UI ke English → narasi "For the period of Q4 2026, SPI has issued …".

**Regresi audit-service**
8. Unduh DOCX salah satu LHA → nama perusahaan benar (bukan "PT AIFL Indonesia", bukan kosong untuk perusahaan yang ada di master data).

Statistik bisa telat sampai 5 menit kalau proxy-cache Kong aktif (belum dicek di prod).

## 8. Tenant BAI — di luar runbook ini

Tenant `bai` menjalankan audit-service sendiri (`rb_audit_bai_audit_service`, image `rb_audit_audit_service:latest`). Rilis ini belum menyentuhnya, jadi form kompilasi di tenant BAI masih error sampai audit-service BAI di-rebuild. `backend/tenants/bai/docker-compose.yml` di repo memakai path volume Mac lokal, jadi **bukan** file yang dipakai di VPS. Jangan deploy tenant dari file ini; minta user/agen devops menyusun langkah khusus tenant.

## 9. Laporan akhir yang harus diberikan agen

- SHA yang di-deploy dan service apa saja.
- Output ringkas `deploy-prod.sh` (backup path, hasil row-count, health).
- Hasil smoke test 5c dan cek `apiBase` 6a.
- Item checklist langkah 7 yang sudah/belum dicek, dan oleh siapa.
- Apakah `rbia-frontend-prod-old` masih ada (untuk rollback).
