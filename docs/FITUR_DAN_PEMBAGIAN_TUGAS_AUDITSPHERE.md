# PETA FITUR & PEMBAGIAN KONTRIBUTOR AUDITSPHERE
## AuditSphere — Enterprise Risk-Based Internal Audit & ERM Platform

Dokumen ini memuat daftar lengkap seluruh fitur pada ekosistem **AuditSphere**, deskripsi fungsional setiap modul, serta pemetaan kontributor utama (*Lead Developer / Contributor*) antara **RauMomo (Qiko)** dan **Dimas Saifullah Ramadhan** berdasarkan riwayat rekayasa perangkat lunak (*Git commit history & code ownership*).

---

## Ringkasan Eksekutif Pembagian Peran

| Area / Domain | Lead / Kontributor Utama | Kontributor Pendukung / Kolaborasi |
| :--- | :--- | :--- |
| **Arsitektur Keamanan & Gateway (Kong 3.4)** | **RauMomo (Qiko)** | Dimas Saifullah Ramadhan |
| **Auth & Identity Service (MFA, RBAC, Pakta)** | **RauMomo (Qiko)** | Tim Engineering |
| **AI / ML Model Training & Inference Engine** | **Dimas Saifullah Ramadhan** | RauMomo (Qiko) |
| **Risk Service & Analytics Service** | **Dimas Saifullah Ramadhan** | RauMomo (Qiko) |
| **Master Data & Multi-Tenant Architecture** | **RauMomo (Qiko)** | Dimas Saifullah Ramadhan |
| **Audit Planning (Strategic & Annual Plan / PKAT)** | **Dimas Saifullah Ramadhan** | RauMomo (Qiko) |
| **Audit Fieldwork & Digital Working Paper (KKA)** | **Kolaborasi (Qiko & Dimas)** | - |
| **Reporting (LHA, Executive Summary, ATR)** | **Dimas Saifullah Ramadhan** | RauMomo (Qiko) |
| **Quality Assurance (QAIP) & Consulting Service** | **Dimas Saifullah Ramadhan** | RauMomo (Qiko) |
| **Frontend Core UI/UX, Design System & Nuxt 4** | **RauMomo (Qiko)** | Dimas Saifullah Ramadhan |
| **DevOps, CI/CD Pipeline & VPS Deployment** | **RauMomo (Qiko)** | Dimas Saifullah Ramadhan |
| **Postman API Test Suite Collection** | **Dimas Saifullah Ramadhan** | - |

---

## 1. Modul Keamanan, Autentikasi & Infrastruktur Gateway

### 1.1 Kong API Gateway & Perimeter Security
* **Deskripsi**: Single entry-point reverse proxy untuk seluruh microservices. Menyediakan fitur *bot detection* (blokir otomatis `sqlmap`, `nikto`, `nmap`), rate limiting dinamis, CORS header enforcement, dan SSL/TLS termination.
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*, didukung oleh Dimas Saifullah Ramadhan.
* **Lokasi Kode**: `backend/kong-gateway/`

### 1.2 Auth & Identity Management Service
* **Deskripsi**: Layanan autentikasi terpusat berbasis JWT dengan *token revocation/blacklisting* di Redis, *Multi-Factor Authentication* (TOTP / Google Authenticator), *Device Fingerprinting*, penandatanganan Pakta Integritas (*Confidentiality Agreement*), dan otorisasi dinamis Casbin.
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*.
* **Lokasi Kode**: `backend/auth-service/`, `frontend/stores/auth.ts`, `frontend/types/auth.ts`

### 1.3 Master Data & Tenant Management Service
* **Deskripsi**: Pengelolaan data master organisasi (Departemen, Pegawai/Auditor, Entitas Perusahaan, Struktur Organisasi, dan script otomasi *Tenant Onboarding*).
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*, didukung oleh Dimas Saifullah Ramadhan.
* **Lokasi Kode**: `backend/master-service/`, `scripts/onboard-tenant.sh`, `frontend/stores/department.ts`, `frontend/stores/employee.ts`

---

## 2. Modul Artificial Intelligence & Advanced Analytics

### 2.1 AI / ML Model Training Pipeline
* **Deskripsi**: Pelatihan model Machine Learning untuk 4 domain utama:
  1. *Anomaly Prediction* (Deteksi anomali transaksi audit),
  2. *Department Prediction* (Klasifikasi area audit berdasarkan profil risiko),
  3. *Document Prediction* (Analisis dan klasifikasi dokumen bukti audit),
  4. *KPI & Risk Prediction* (Prediksi deviasi target KPI organisasi).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*.
* **Lokasi Kode**: `ai_model_training/` (`anomaly_prediction/`, `department_prediction/`, `document_prediction/`, `kpi_prediction/`, `save_all_models.py`)

### 2.2 Python AI FastAPI Inference Engine
* **Deskripsi**: Layanan microservice berbasis Python (FastAPI + PyTorch/Scikit-Learn) yang menjalankan model AI secara real-time untuk memberikan skor anomali, rekomendasi pengujian audit, dan ekstraksi wawasan otomatis.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung integrasi oleh RauMomo (Qiko).
* **Lokasi Kode**: `backend/python-ai/`

### 2.3 Analytics & Statistical Risk Aggregator Service
* **Deskripsi**: Microservice kalkulasi metrik agregat, tren risiko lintas periode, korelasi temuan audit, dan penyedia data visualisasi grafik untuk dashboard eksekutif.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `backend/analytics-service/`

---

## 3. Modul Enterprise Risk Management (ERM)

### 3.1 Risk Profile & Risk Register
* **Deskripsi**: Registrasi dan pemetaan profil risiko perusahaan, klasifikasi risiko inheren (*Inherent Risk*), efektivitas kontrol internal, dan risiko residual (*Residual Risk*).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `backend/risk-service/`, `frontend/pages/profile-risk/`, `frontend/stores/profile-risk.ts`

### 3.2 Risk Appetite & Tolerance Framework
* **Deskripsi**: Penetapan batas toleransi risiko (*Risk Appetite Statements & Threshold Limits*) beserta matriks eskalasi jika risiko melampaui batas ambang perusahaan.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, diselaraskan oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/risk-appetite/`, `frontend/stores/risk-appetite.ts`

### 3.3 Risk Mitigation Action Plan & RCM (Risk Control Matrix)
* **Deskripsi**: Perumusan rencana aksi mitigasi risiko, penetapan *Risk Owner*, pemantauan progres mitigasi, serta matriks hubungan antara risiko dan kontrol (*Risk & Control Matrix*).
* **Kontributor**: **Dimas Saifullah Ramadhan** & **RauMomo (Qiko)** *(Kolaborasi)*.
* **Lokasi Kode**: `frontend/pages/mitigation-risk/`, `frontend/pages/rcm/`, `frontend/stores/mitigation-risk.ts`, `frontend/stores/rcm.ts`

---

## 4. Modul Tata Kelola & Perencanaan Audit (Audit Governance & Planning)

### 4.1 Audit Universe & Audit Charter
* **Deskripsi**: Inventarisasi seluruh auditable entity (Audit Universe) serta penetapan piagam audit internal (*Audit Charter*), mandat, SOP, dan pedoman kerja auditor (*Audit Guidelines*).
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*, didukung oleh Dimas Saifullah Ramadhan.
* **Lokasi Kode**: `frontend/pages/audit-universe/`, `frontend/pages/audit-charter/`, `frontend/pages/audit-sop/`, `frontend/pages/audit-guideline/`, `frontend/stores/charter.ts`

### 4.2 Strategic Audit Plan (Rencana Audit Jangka Panjang 3-5 Tahun)
* **Deskripsi**: Penyusunan peta jalan audit strategis multi-tahun yang selaras dengan visi, misi, dan sasaran strategis korporasi.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/strategic-audit-plan/`, `frontend/stores/strategic-audit-plan.ts`

### 4.3 Annual Audit Plan (PKAT - Program Kerja Audit Tahunan) & Activity Plan
* **Deskripsi**: Penyusunan PKAT berbasis pemeringkatan risiko (*Risk-Based Prioritization*), estimasi kebutuhan hari-orang (*mandays*), jadwal pelaksanaan audit, dan alokasi tim auditor.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/annual-audit-plan/`, `frontend/pages/audit-activity-plan/`, `frontend/stores/annual-audit.ts`, `frontend/stores/activity-plan.ts`

### 4.4 Assignment Letter (Surat Tugas Audit)
* **Deskripsi**: Pembuatan, penomoran terstruktur, alokasi penugasan Ketua Tim & Anggota Auditor, serta penerbitan surat tugas resmi pelaksanaan audit.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/assignment-letter/`, `frontend/pages/surat-tugas/`, `frontend/stores/assignment-letter.ts`

---

## 5. Modul Pelaksanaan Audit (Audit Fieldwork & Digital Working Paper)

### 5.1 Digital Working Paper (Kertas Kerja Audit / KKA)
* **Deskripsi**: Modul utama pelaksanaan audit lapangan mencakup pengujian substantif dan *compliance test*, dokumentasi kriteria, kondisi, penyebab (*cause*), dan dampak (*effect*), pengunggahan bukti audit (*audit evidence*), serta alur *review* bertingkat (Auditor → Team Leader → Quality Reviewer).
* **Kontributor**: **Kolaborasi Erat**:
  * **RauMomo (Qiko)**: Arsitektur eksekusi pengujian (*fieldwork test controls*), struktur data KKA, *step sampling logic*, dan sinkronisasi Kafka audit trail.
  * **Dimas Saifullah Ramadhan**: Fitur import template working paper, kalkulasi temuan berulang (*repeat findings*), dan siklus approval KKA.
* **Lokasi Kode**: `backend/audit-service/`, `frontend/pages/audit-fieldwork/`, `frontend/pages/working-paper/`, `frontend/stores/audit-fieldwork.ts`, `frontend/stores/working-paper.ts`

---

## 6. Modul Pelaporan, Monitoring & QA (Reporting & Quality Assurance)

### 6.1 Audit Result Report (LHA - Laporan Hasil Audit)
* **Deskripsi**: Kompilasi temuan audit menjadi laporan formal LHA dengan klasifikasi tingkat signifikansi temuan (*High, Medium, Low*), rekomendasi perbaikan, dan tanggapan manajemen (*Management Response*).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/audit-result-report/`, `frontend/stores/audit-result-report.ts`

### 6.2 Executive Summary & Executive Dashboard
* **Deskripsi**: Ringkasan eksekutif bagi Direksi dan Komite Audit yang merangkum *health score* pengendalian internal, distribusi temuan kritis, dan status area berisiko tinggi.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/executive-summary/`, `frontend/stores/executive-summary.ts`

### 6.3 Action Taken Report (ATR / Pemantauan Tindak Lanjut Temuan)
* **Deskripsi**: Pelacakan komitmen tindak lanjut rekomendasi audit oleh Auditee (*Action Plan Monitoring*), verifikasi bukti penyelesaian oleh auditor, dan status keterlambatan (*overdue tracking*).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/action-taken-report/`, `frontend/stores/action-taken-report.ts`

### 6.4 Quality Assurance and Improvement Program (QAIP) & Auditee Survey
* **Deskripsi**: Penilaian kualitas penugasan audit internal sesuai standar IIA IPPF, survei kepuasan auditee (*Auditee Satisfaction Survey*), dan evaluasi kinerja tim audit (*Performance Report*).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*, didukung oleh RauMomo (Qiko).
* **Lokasi Kode**: `frontend/pages/quality-assurance/`, `frontend/pages/auditee-survey/`, `frontend/pages/performance-report/`, `frontend/stores/quality-assurance.ts`

### 6.5 Consulting & Advisory Service Management
* **Deskripsi**: Modul pencatatan layanan konsultasi non-assurance internal audit (advisory, fasilitasi risk assessment, review kebijakan).
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*.
* **Lokasi Kode**: `frontend/pages/consulting-service/`, `frontend/stores/consulting-service.ts`

---

## 7. Fondasi Frontend UI/UX, Testing & DevOps

### 7.1 Frontend Design System & Framework (Nuxt 4 / Vue 3)
* **Deskripsi**: Pembangunan fondasi arsitektur frontend dengan Nuxt 4, integrasi Nuxt UI & TailwindCSS, tema *Space Grotesk*, state management terpusat (Pinia), i18n multi-bahasa (ID/EN), utilitas format data dan toast notifikasi.
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*, didukung oleh Dimas Saifullah Ramadhan.
* **Lokasi Kode**: `frontend/app.vue`, `frontend/layouts/`, `frontend/components/core/`, `frontend/locales/`, `frontend/utils/`

### 7.2 Automated Testing Suite (Vitest & Playwright)
* **Deskripsi**: Pengujian unit otomatis untuk store/komponen dengan Vitest dan skenario End-to-End (E2E) testing dengan Playwright.
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*.
* **Lokasi Kode**: `frontend/tests/unit/`, `frontend/tests/e2e/`, `frontend/vitest.config.ts`, `frontend/playwright.config.ts`

### 7.3 DevOps, Docker Multi-Service & CI/CD Pipeline
* **Deskripsi**: Konfigurasi Docker compose microservices (Postgres, Redis, Kafka, Zookeeper, Kong, Go services, Python AI, Nuxt Frontend), script deployment VPS otomatis (`deploy-backend.sh`, `deploy-prod.sh`, `deploy-dev.sh`), dan otomasi GitHub Actions (`.github/workflows/deploy.yml`).
* **Kontributor**: **RauMomo (Qiko)** *(Lead)*, didukung oleh Dimas Saifullah Ramadhan.
* **Lokasi Kode**: `docker-compose.yml`, `backend/docker-compose.yml`, `frontend/docker-compose.prod.yml`, `.github/workflows/deploy.yml`, `backend/deploy-backend.sh`

### 7.4 Postman API Collection & Documentation
* **Deskripsi**: Dokumentasi dan koleksi pengujian endpoint REST API menyeluruh untuk seluruh microservices di Postman Workspace.
* **Kontributor**: **Dimas Saifullah Ramadhan** *(Lead)*.
* **Lokasi Kode**: `postman/collections/Risk-Based Audit System/`
