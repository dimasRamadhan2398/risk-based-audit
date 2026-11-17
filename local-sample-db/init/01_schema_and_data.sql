-- ============================================================================
-- PT Bank Nusantara Sejahtera (Local Sample Audit Database)
-- Standalone External PostgreSQL Database for Manual AuditSphere Testing
-- ============================================================================

-- Ensure reproducible seed
SELECT setseed(0.4242);

-- 1. Branch Directory (10 Cabang Regional)
CREATE TABLE IF NOT EXISTS branches (
    cabang_id     SERIAL PRIMARY KEY,
    branch_code   VARCHAR(16) UNIQUE NOT NULL,
    branch_name   VARCHAR(100) NOT NULL,
    region        VARCHAR(60) NOT NULL,
    city          VARCHAR(60) NOT NULL,
    risk_tier     VARCHAR(20) NOT NULL DEFAULT 'Medium Risk',
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO branches (branch_code, branch_name, region, city, risk_tier) VALUES
('KCU-JKT1', 'KCU Sudirman Plaza', 'Jabodetabek', 'Jakarta Selatan', 'High Risk'),
('KCP-JKT2', 'KCP Menteng Raya', 'Jabodetabek', 'Jakarta Pusat', 'High Risk'),
('KCU-BDG1', 'KCU Bandung Dago', 'Jawa Barat', 'Bandung', 'Medium Risk'),
('KCP-SMG1', 'KCP Semarang Pandanaran', 'Jawa Tengah', 'Semarang', 'Low Risk'),
('KCU-SBY1', 'KCU Surabaya Tunjungan', 'Jawa Timur', 'Surabaya', 'High Risk'),
('KCP-MKS1', 'KCP Makassar Losari', 'Sulawesi', 'Makassar', 'Medium Risk'),
('KCU-MDN1', 'KCU Medan Kesawan', 'Sumatera Utara', 'Medan', 'Medium Risk'),
('KCP-DPS1', 'KCP Denpasar Renon', 'Bali & Nusa Tenggara', 'Denpasar', 'Low Risk'),
('KCU-BPN1', 'KCU Balikpapan Sudirman', 'Kalimantan', 'Balikpapan', 'Medium Risk'),
('KP-OPR01', 'Kantor Pusat Operasional', 'Head Office', 'Jakarta Selatan', 'High Risk');

-- 2. Corporate Financial Transactions (3,250 rows with realistic Benford baseline + Digit 4/8/9 anomalies)
CREATE TABLE IF NOT EXISTS corporate_transactions (
    trx_id                SERIAL PRIMARY KEY,
    trx_ref_number        VARCHAR(30) NOT NULL,
    trx_date              TIMESTAMP NOT NULL,
    branch_name           VARCHAR(100) NOT NULL,
    customer_name         VARCHAR(150) NOT NULL,
    category              VARCHAR(50) NOT NULL,
    channel               VARCHAR(40) NOT NULL,
    account_type          VARCHAR(30) NOT NULL,
    amount                NUMERIC(18,2) NOT NULL,
    authorization_status  VARCHAR(30) NOT NULL DEFAULT 'AUTHORIZED',
    is_suspicious         BOOLEAN NOT NULL DEFAULT FALSE,
    notes                 TEXT,
    updated_at            TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_corp_trx_date ON corporate_transactions(trx_date);
CREATE INDEX IF NOT EXISTS idx_corp_trx_ref ON corporate_transactions(trx_ref_number);

-- 2A. 2,780 Benford-natural corporate transactions (July - Nov 2026, normal business hours)
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-2026-' || LPAD(gs::TEXT, 6, '0'),
    TIMESTAMP '2026-07-01 08:00:00'
      + ((gs % 135) || ' days')::INTERVAL
      + (((8 + (gs % 9)) || ' hours')::INTERVAL)
      + (((gs * 7) % 55) || ' minutes')::INTERVAL,
    (ARRAY[
        'KCU Sudirman Plaza','KCP Menteng Raya','KCU Bandung Dago','KCP Semarang Pandanaran',
        'KCU Surabaya Tunjungan','KCP Makassar Losari','KCU Medan Kesawan','KCP Denpasar Renon',
        'KCU Balikpapan Sudirman','Kantor Pusat Operasional'
    ])[1 + (gs % 10)],
    (ARRAY[
        'PT Cakrawala Logistik Nusantara','CV Timur Jaya Abadi','PT Graha Energi Prima',
        'PT Mitra Konstruksi Mandiri','PT Samudra Niaga Sejahtera','CV Karya Bhakti Utama',
        'PT Sinar Medika Farma','PT Archipelago Pangan Tbk','PT Bumi Pertiwi Properti',
        'PT Sentra Teknologi Digital','Koperasi Karyawan Sejahtera','PT Pelabuhan Lintas Timur',
        'CV Andalas Agro Makmur','PT Surya Kencana Industri','PT Prima Telekomunikasi'
    ])[1 + (gs % 15)],
    (ARRAY[
        'Vendor Settlement','Commercial Transfer','Payroll Batch','Tax Remittance',
        'Treasury Placement','Procurement Payment'
    ])[1 + (gs % 6)],
    (ARRAY['BI-RTGS','SKNBI','HOST_TO_HOST','TREASURY_DESK','INTERNET_CORP'])[1 + (gs % 5)],
    (ARRAY['GIRO_KORPORASI','ESCROW_PROJECT','OPERASIONAL_CABANG'])[1 + (gs % 3)],
    -- Natural Benford power-law distribution (~Rp 2.1M to Rp 380M)
    ROUND((POWER(10, 6.2 + (random() * 2.35)))::NUMERIC, 0),
    CASE WHEN gs % 25 = 0 THEN 'APPROVED_DUAL' ELSE 'AUTHORIZED' END,
    FALSE,
    'Transaksi operasional rutin korporasi',
    TIMESTAMP '2026-11-16 09:00:00'
FROM generate_series(1, 2780) AS gs;

-- 2B. 190 Intentional Benford Digit-4 Threshold-Splitting Anomalies (Rp 485M - Rp 498M, below Rp 500M dual-control)
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-SPL4-' || LPAD(gs::TEXT, 5, '0'),
    TIMESTAMP '2026-09-01 09:00:00'
      + ((gs % 70) || ' days')::INTERVAL
      + (((9 + (gs % 8)) || ' hours')::INTERVAL)
      + (((gs * 11) % 50) || ' minutes')::INTERVAL,
    CASE WHEN gs % 3 = 0 THEN 'KCP Menteng Raya' ELSE 'KCU Surabaya Tunjungan' END,
    (ARRAY[
        'PT Cakrawala Logistik Nusantara','CV Karya Bhakti Utama','PT Graha Energi Prima',
        'PT Bumi Pertiwi Properti','CV Timur Jaya Abadi'
    ])[1 + (gs % 5)],
    CASE WHEN gs % 2 = 0 THEN 'Vendor Settlement' ELSE 'Procurement Payment' END,
    CASE WHEN gs % 4 = 0 THEN 'BRANCH_OVERRIDE' ELSE 'SKNBI' END,
    'GIRO_KORPORASI',
    (ARRAY[485000000, 490000000, 492500000, 495000000, 498000000])[1 + (gs % 5)],
    CASE WHEN gs % 4 = 0 THEN 'PENDING_OVERRIDE' ELSE 'AUTHORIZED' END,
    (gs % 3 = 0),
    'Indikasi pemecahan transaksi tepat di bawah batas kontrol ganda Rp 500 Juta',
    TIMESTAMP '2026-11-16 09:05:00'
FROM generate_series(1, 190) AS gs;

-- 2C. 135 Intentional Benford Digit-8 After-Hours & High-Value Anomalies (Rp 850M / Rp 880M / Rp 8.5B)
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-AFT8-' || LPAD(gs::TEXT, 5, '0'),
    TIMESTAMP '2026-08-15 22:15:00'
      + ((gs % 85) || ' days')::INTERVAL
      + (((gs % 5) || ' hours')::INTERVAL),
    CASE WHEN gs % 2 = 0 THEN 'Kantor Pusat Operasional' ELSE 'KCU Sudirman Plaza' END,
    (ARRAY[
        'PT Samudra Niaga Sejahtera','PT Mitra Konstruksi Mandiri','PT Archipelago Pangan Tbk',
        'PT Surya Kencana Industri'
    ])[1 + (gs % 4)],
    'Treasury Placement',
    'BRANCH_OVERRIDE',
    'ESCROW_PROJECT',
    CASE WHEN gs % 9 = 0 THEN 8500000000 ELSE (ARRAY[850000000, 875000000, 880000000])[1 + (gs % 3)] END,
    'PENDING_OVERRIDE',
    TRUE,
    'Otorisasi override di luar jam operasional (>22:00 WIB) bernilai di atas batas cabang',
    TIMESTAMP '2026-11-16 09:10:00'
FROM generate_series(1, 135) AS gs;

-- 2D. 75 Intentional Benford Digit-9 High-Risk AML/Structuring Round-Amount Transfers (Rp 950M - Rp 995M)
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-AML9-' || LPAD(gs::TEXT, 5, '0'),
    TIMESTAMP '2026-09-10 15:00:00'
      + ((gs % 60) || ' days')::INTERVAL
      + (((gs % 7) || ' hours')::INTERVAL),
    CASE WHEN gs % 2 = 0 THEN 'KCU Medan Kesawan' ELSE 'KCU Balikpapan Sudirman' END,
    (ARRAY[
        'CV Andalas Agro Makmur','PT Prima Telekomunikasi','PT Pelabuhan Lintas Timur'
    ])[1 + (gs % 3)],
    'Commercial Transfer',
    'BI-RTGS',
    'GIRO_KORPORASI',
    (ARRAY[950000000, 980000000, 990000000, 995000000])[1 + (gs % 4)],
    CASE WHEN gs % 3 = 0 THEN 'REJECTED' ELSE 'PENDING_OVERRIDE' END,
    TRUE,
    'Pola angka bulat signifikan mendekati Rp 1 Miliar dengan tanda mencurigakan',
    TIMESTAMP '2026-11-16 09:15:00'
FROM generate_series(1, 75) AS gs;

-- 2E. 40 High-Value Syndication / Treasury Loans > Rp 1.2B across digits 1, 2, 3
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-BIG1-' || LPAD(gs::TEXT, 5, '0'),
    TIMESTAMP '2026-08-01 10:30:00' + ((gs * 2) || ' days')::INTERVAL,
    'Kantor Pusat Operasional',
    'PT Graha Energi Prima',
    'Loan Disbursement',
    'BI-RTGS',
    'GIRO_KORPORASI',
    (ARRAY[1250000000, 2150000000, 3400000000])[1 + (gs % 3)],
    'APPROVED_DUAL',
    (gs % 5 = 0),
    'Pencairan kredit korporasi berskala besar',
    TIMESTAMP '2026-11-16 09:20:00'
FROM generate_series(1, 40) AS gs;

-- 2F. 18 Intentional Duplicate Same-Customer/Amount/Date Pairs + 12 Reused trx_ref_number Across Dates (CAATT #2)
INSERT INTO corporate_transactions (
    trx_ref_number, trx_date, branch_name, customer_name, category, channel,
    account_type, amount, authorization_status, is_suspicious, notes, updated_at
)
SELECT
    'NSA-DUPA-' || LPAD(gs::TEXT, 4, '0'),
    TIMESTAMP '2026-10-12 11:05:00' + ((gs % 9) || ' days')::INTERVAL,
    'KCU Bandung Dago',
    'PT Sentra Teknologi Digital',
    'Vendor Settlement',
    'HOST_TO_HOST',
    'GIRO_KORPORASI',
    145000000 + ((gs % 9) * 15000000),
    'AUTHORIZED',
    TRUE,
    'Duplikasi pembayaran ganda nasabah & nominal sama (Transaksi #1)',
    TIMESTAMP '2026-11-16 09:25:00'
FROM generate_series(1, 9) AS gs
UNION ALL
SELECT
    'NSA-DUPB-' || LPAD(gs::TEXT, 4, '0'),
    TIMESTAMP '2026-10-12 11:06:30' + ((gs % 9) || ' days')::INTERVAL,
    'KCU Bandung Dago',
    'PT Sentra Teknologi Digital',
    'Vendor Settlement',
    'HOST_TO_HOST',
    'GIRO_KORPORASI',
    145000000 + ((gs % 9) * 15000000),
    'AUTHORIZED',
    TRUE,
    'Duplikasi pembayaran ganda nasabah & nominal sama (Transaksi #2)',
    TIMESTAMP '2026-11-16 09:25:00'
FROM generate_series(1, 9) AS gs
UNION ALL
-- Same trx_ref_number repeated on a different day
SELECT
    'NSA-REUSE-' || LPAD(gs::TEXT, 4, '0'),
    TIMESTAMP '2026-10-20 14:10:00',
    'KCP Makassar Losari',
    'PT Sinar Medika Farma',
    'Procurement Payment',
    'SKNBI',
    'OPERASIONAL_CABANG',
    210000000 + (gs * 1000000),
    'AUTHORIZED',
    TRUE,
    'Referensi transaksi pertama sebelum pengulangan nomor referensi',
    TIMESTAMP '2026-11-16 09:25:00'
FROM generate_series(1, 6) AS gs
UNION ALL
SELECT
    'NSA-REUSE-' || LPAD(gs::TEXT, 4, '0'),
    TIMESTAMP '2026-10-22 14:15:00',
    'KCP Makassar Losari',
    'PT Sinar Medika Farma',
    'Procurement Payment',
    'SKNBI',
    'OPERASIONAL_CABANG',
    210000000 + (gs * 1000000),
    'AUTHORIZED',
    TRUE,
    'Nomor referensi transaksi identik muncul kembali pada tanggal berbeda',
    TIMESTAMP '2026-11-16 09:25:00'
FROM generate_series(1, 6) AS gs;

-- 3. Credit Loan Portfolio (220 rows, including 26 policy & NPL breaches for CAATT #6)
CREATE TABLE IF NOT EXISTS credit_loan_portfolio (
    loan_id           SERIAL PRIMARY KEY,
    loan_number       VARCHAR(24) UNIQUE NOT NULL,
    product_name      VARCHAR(100) NOT NULL,
    customer_name     VARCHAR(150) NOT NULL,
    branch_name       VARCHAR(100) NOT NULL,
    plafond           NUMERIC(18,2) NOT NULL,
    outstanding       NUMERIC(18,2) NOT NULL,
    interest_rate     NUMERIC(5,2) NOT NULL,
    tenor_months      INT NOT NULL,
    collateral_value  NUMERIC(18,2) NOT NULL,
    collectability    INT NOT NULL DEFAULT 1,
    days_past_due     INT NOT NULL DEFAULT 0,
    status            VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    disbursed_at      DATE NOT NULL,
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO credit_loan_portfolio (
    loan_number, product_name, customer_name, branch_name, plafond, outstanding,
    interest_rate, tenor_months, collateral_value, collectability, days_past_due,
    status, disbursed_at, updated_at
)
SELECT
    'LN-NSA-' || LPAD(gs::TEXT, 5, '0'),
    (ARRAY[
        'Kredit Modal Kerja Infrastruktur','Pembiayaan Alat Berat',
        'KPR Komersial Eksekutif','Kredit Sindikasi Energi'
    ])[1 + (gs % 4)],
    (ARRAY[
        'PT Cakrawala Logistik Nusantara','PT Graha Energi Prima','PT Mitra Konstruksi Mandiri',
        'PT Samudra Niaga Sejahtera','CV Timur Jaya Abadi','PT Surya Kencana Industri',
        'PT Bumi Pertiwi Properti','PT Archipelago Pangan Tbk'
    ])[1 + (gs % 8)],
    (ARRAY[
        'KCU Sudirman Plaza','KCP Menteng Raya','KCU Bandung Dago','KCU Surabaya Tunjungan',
        'KCU Medan Kesawan','Kantor Pusat Operasional'
    ])[1 + (gs % 6)],
    -- Breach plafond (> Rp 5B) on every 15th row
    CASE
        WHEN gs % 15 = 0 THEN 7500000000::BIGINT + ((gs % 4)::BIGINT * 1250000000::BIGINT)
        ELSE 350000000::BIGINT + ((gs % 12)::BIGINT * 280000000::BIGINT)
    END,
    CASE
        WHEN gs % 15 = 0 THEN 6800000000::BIGINT + ((gs % 4)::BIGINT * 1100000000::BIGINT)
        ELSE 250000000::BIGINT + ((gs % 12)::BIGINT * 210000000::BIGINT)
    END,
    -- Breach interest rate (> 24%) on every 18th row
    CASE
        WHEN gs % 18 = 0 THEN 27.25
        ELSE 8.75 + ((gs % 7) * 0.85)
    END,
    -- Breach tenor (> 120 months) on every 22nd row
    CASE
        WHEN gs % 22 = 0 THEN 156
        ELSE 24 + ((gs % 6) * 12)
    END,
    CASE
        WHEN gs % 15 = 0 THEN 5200000000::BIGINT
        ELSE 650000000::BIGINT + ((gs % 12)::BIGINT * 350000000::BIGINT)
    END,
    -- NPL collectability >= 3 on every 14th row
    CASE
        WHEN gs % 28 = 0 THEN 5
        WHEN gs % 19 = 0 THEN 4
        WHEN gs % 14 = 0 THEN 3
        WHEN gs % 9 = 0 THEN 2
        ELSE 1
    END,
    CASE
        WHEN gs % 28 = 0 THEN 245
        WHEN gs % 19 = 0 THEN 155
        WHEN gs % 14 = 0 THEN 105
        WHEN gs % 9 = 0 THEN 38
        ELSE 0
    END,
    CASE WHEN gs % 28 = 0 THEN 'NPL_MACET' WHEN gs % 14 = 0 THEN 'RESTRUCTURED' ELSE 'ACTIVE' END,
    DATE '2025-01-15' + (gs * 2),
    TIMESTAMP '2026-11-16 09:30:00'
FROM generate_series(1, 220) AS gs;

-- 4. General Ledger Journal Entries (420 rows with intentional sequence gaps & missing vouchers for CAATT #2 & #5)
CREATE TABLE IF NOT EXISTS gl_journal_entries (
    gl_id            SERIAL PRIMARY KEY,
    gl_date          DATE NOT NULL,
    gl_account_code  VARCHAR(20) NOT NULL,
    gl_account_name  VARCHAR(100) NOT NULL,
    voucher_number   VARCHAR(30),
    debit_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    credit_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    branch_name      VARCHAR(100) NOT NULL,
    posted_by        VARCHAR(60) NOT NULL,
    description      TEXT,
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert 420 sequenced rows while intentionally skipping numbers 42..44, 115..118, 240..242, 351..354
-- and setting voucher_number = NULL on 6 rows to test both sequential gaps and missing vouchers.
WITH seq AS (
    SELECT gs AS n
    FROM generate_series(1, 434) AS gs
    WHERE gs NOT BETWEEN 42 AND 44
      AND gs NOT BETWEEN 115 AND 118
      AND gs NOT BETWEEN 240 AND 242
      AND gs NOT BETWEEN 351 AND 354
)
INSERT INTO gl_journal_entries (
    gl_date, gl_account_code, gl_account_name, voucher_number,
    debit_amount, credit_amount, branch_name, posted_by, description, updated_at
)
SELECT
    DATE '2026-07-01' + (n % 130),
    (ARRAY['1101.01','1205.02','2101.04','4102.01','5103.08'])[1 + (n % 5)],
    (ARRAY['Kas & Penempatan BI','Kredit Yang Diberikan','Simpanan Giro Korporasi','Pendapatan Bunga','Beban Operasional Cabang'])[1 + (n % 5)],
    CASE
        WHEN n IN (65, 148, 212, 289, 317, 405) THEN NULL
        ELSE 'JV-2026-' || LPAD(n::TEXT, 4, '0')
    END,
    CASE WHEN n % 2 = 0 THEN 125000000 + ((n % 15) * 45000000) ELSE 0 END,
    CASE WHEN n % 2 = 1 THEN 125000000 + ((n % 15) * 45000000) ELSE 0 END,
    (ARRAY['KCU Sudirman Plaza','KCP Menteng Raya','KCU Surabaya Tunjungan','KCU Bandung Dago','Kantor Pusat Operasional'])[1 + (n % 5)],
    (ARRAY['acct_spv_01','gl_officer_02','ctrl_mgr_03','sys_eod_batch'])[1 + (n % 4)],
    'Posting jurnal akuntansi harian Bank Nusantara Sejahtera',
    TIMESTAMP '2026-11-16 09:35:00'
FROM seq;

-- 5. Procurement Vendor Contracts & Invoices (150 rows for Working Paper Evidence & AI Anomaly Detection)
CREATE TABLE IF NOT EXISTS vendor_contracts (
    contract_id      SERIAL PRIMARY KEY,
    contract_no      VARCHAR(32) UNIQUE NOT NULL,
    vendor_name      VARCHAR(120) NOT NULL,
    branch_name      VARCHAR(100) NOT NULL,
    contract_value   NUMERIC(18,2) NOT NULL,
    po_number        VARCHAR(32) NOT NULL,
    approval_status  VARCHAR(30) NOT NULL,
    approver_id      VARCHAR(50) NOT NULL,
    signed_at        DATE NOT NULL,
    is_flagged       BOOLEAN NOT NULL DEFAULT FALSE,
    audit_note       TEXT,
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO vendor_contracts (
    contract_no, vendor_name, branch_name, contract_value, po_number,
    approval_status, approver_id, signed_at, is_flagged, audit_note, updated_at
)
SELECT
    'CTR-NSA-' || LPAD(gs::TEXT, 4, '0'),
    (ARRAY[
        'PT Solusi Integrasi Cloud','PT Garda Keamanan Siber','CV Konstruksi Pilar Jaya',
        'PT Data Center Nusantara','PT Logistik Armada Prima'
    ])[1 + (gs % 5)],
    (ARRAY['Kantor Pusat Operasional','KCU Sudirman Plaza','KCU Surabaya Tunjungan','KCU Bandung Dago'])[1 + (gs % 4)],
    CASE
        WHEN gs % 10 = 0 THEN 985000000
        WHEN gs % 7 = 0 THEN 492000000
        ELSE 75000000 + ((gs % 8) * 32000000)
    END,
    'PO-NSA-' || LPAD(gs::TEXT, 5, '0'),
    CASE WHEN gs % 7 = 0 THEN 'PENDING_REVIEW' WHEN gs % 10 = 0 THEN 'OVERRIDE_APPROVED' ELSE 'APPROVED' END,
    CASE WHEN gs % 10 = 0 THEN 'proc_head_override' ELSE 'cfo_committee' END,
    DATE '2026-07-10' + gs,
    (gs % 7 = 0 OR gs % 10 = 0),
    CASE
        WHEN gs % 10 = 0 THEN 'Nilai kontrak mendekati batas wewenang direksi tanpa lelang terbuka'
        WHEN gs % 7 = 0 THEN 'Kontrak dipecah menjadi beberapa PO di bawah batas tender'
        ELSE 'Pengadaan sesuai prosedur komite'
    END,
    TIMESTAMP '2026-11-16 09:40:00'
FROM generate_series(1, 150) AS gs;

-- 6. Audit & Risk Incident Register (40 rows for Risk Management, QAR & Soft-Delete Testing)
CREATE TABLE IF NOT EXISTS audit_risk_incidents (
    incident_id      SERIAL PRIMARY KEY,
    incident_code    VARCHAR(24) UNIQUE NOT NULL,
    entity_unit      VARCHAR(100) NOT NULL,
    risk_category    VARCHAR(60) NOT NULL,
    severity_score   INT NOT NULL,
    exposure_value   NUMERIC(18,2) NOT NULL,
    status           VARCHAR(24) NOT NULL,
    finding_summary  TEXT NOT NULL,
    reported_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    is_deleted       BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO audit_risk_incidents (
    incident_code, entity_unit, risk_category, severity_score, exposure_value,
    status, finding_summary, reported_at, is_deleted
)
SELECT
    'INC-NSA-' || LPAD(gs::TEXT, 3, '0'),
    (ARRAY[
        'KCU Sudirman Plaza','KCP Menteng Raya','KCU Surabaya Tunjungan',
        'Kantor Pusat Operasional','Divisi Kredit Komersial','Divisi Pengadaan & TI'
    ])[1 + (gs % 6)],
    (ARRAY[
        'AML / Structuring Risk','Dual-Control Bypass','Credit Underwriting Breach',
        'GL Reconciliation Variance','Vendor Conflict of Interest'
    ])[1 + (gs % 5)],
    CASE WHEN gs % 3 = 0 THEN 88 ELSE 45 + (gs % 40) END,
    150000000::BIGINT + (gs::BIGINT * 85000000::BIGINT),
    CASE WHEN gs % 3 = 0 THEN 'ESCALATED' WHEN gs % 2 = 0 THEN 'OPEN' ELSE 'IN_REVIEW' END,
    'Temuan pengawasan internal #' || gs || ' pada pengendalian transaksi dan kredit cabang',
    TIMESTAMP '2026-10-01 10:00:00' + (gs || ' days')::INTERVAL,
    (gs IN (39, 40))
FROM generate_series(1, 40) AS gs;

-- Grant read/write privileges cleanly to audit_client
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO audit_client;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO audit_client;

-- Update PostgreSQL planner statistics so Schema Explorer shows exact row counts immediately
ANALYZE;
