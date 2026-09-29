-- ============================================================================
-- AuditSphere Data Hub — Core Banking System (CBS) Simulator
-- Mock banking database with ~10K transactions, ~500 accounts, ~200 customers
-- Includes intentionally injected anomalies for CAATT testing
-- ============================================================================

-- Create dedicated schema
CREATE SCHEMA IF NOT EXISTS cbs;

-- ============================================================================
-- DIMENSION TABLES
-- ============================================================================

-- Branches / Cabang
CREATE TABLE cbs.branches (
    branch_id       SERIAL PRIMARY KEY,
    branch_code     VARCHAR(10) UNIQUE NOT NULL,
    branch_name     VARCHAR(100) NOT NULL,
    branch_type     VARCHAR(20) DEFAULT 'BRANCH', -- HQ, BRANCH, SUB_BRANCH
    region          VARCHAR(50),
    city            VARCHAR(50),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Customers / Nasabah
CREATE TABLE cbs.customers (
    customer_id     SERIAL PRIMARY KEY,
    cif_number      VARCHAR(20) UNIQUE NOT NULL,    -- Customer Information File
    customer_name   VARCHAR(150) NOT NULL,
    customer_type   VARCHAR(20) DEFAULT 'INDIVIDUAL', -- INDIVIDUAL, CORPORATE
    id_type         VARCHAR(20) DEFAULT 'KTP',
    id_number       VARCHAR(30),
    phone           VARCHAR(20),
    email           VARCHAR(100),
    address         TEXT,
    risk_profile    VARCHAR(20) DEFAULT 'Low Risk', -- Low Risk, Medium Risk, High Risk
    branch_id       INT REFERENCES cbs.branches(branch_id),
    registration_date DATE DEFAULT CURRENT_DATE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Accounts / Rekening
CREATE TABLE cbs.accounts (
    account_id      SERIAL PRIMARY KEY,
    account_number  VARCHAR(20) UNIQUE NOT NULL,
    account_type    VARCHAR(30) NOT NULL, -- SAVINGS, CURRENT, DEPOSIT, LOAN
    currency        VARCHAR(3) DEFAULT 'IDR',
    balance         NUMERIC(18,2) DEFAULT 0,
    credit_limit    NUMERIC(18,2),
    interest_rate   NUMERIC(5,2),
    status          VARCHAR(20) DEFAULT 'ACTIVE',
    customer_id     INT REFERENCES cbs.customers(customer_id),
    branch_id       INT REFERENCES cbs.branches(branch_id),
    opened_date     DATE DEFAULT CURRENT_DATE,
    maturity_date   DATE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Products / Produk Perbankan
CREATE TABLE cbs.products (
    product_id      SERIAL PRIMARY KEY,
    product_code    VARCHAR(20) UNIQUE NOT NULL,
    product_name    VARCHAR(100) NOT NULL,
    product_type    VARCHAR(30) NOT NULL, -- FUNDING, LENDING, TREASURY, PAYMENT
    min_balance     NUMERIC(18,2) DEFAULT 0,
    max_balance     NUMERIC(18,2),
    min_rate        NUMERIC(5,2),
    max_rate        NUMERIC(5,2),
    min_tenor_days  INT,
    max_tenor_days  INT,
    max_plafond     NUMERIC(18,2),
    requires_collateral BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- Product Policy Rules
CREATE TABLE cbs.product_policies (
    policy_id       SERIAL PRIMARY KEY,
    product_id      INT REFERENCES cbs.products(product_id),
    rule_name       VARCHAR(100) NOT NULL,
    rule_type       VARCHAR(30) NOT NULL, -- RATE_LIMIT, AMOUNT_LIMIT, TENOR_LIMIT, AUTH_LEVEL
    min_value       NUMERIC(18,4),
    max_value       NUMERIC(18,4),
    description     TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- FACT TABLES
-- ============================================================================

-- Transactions / Transaksi
CREATE TABLE cbs.transactions (
    transaction_id      SERIAL PRIMARY KEY,
    trx_ref_number      VARCHAR(30) UNIQUE NOT NULL,
    trx_date            DATE NOT NULL,
    trx_time            TIME NOT NULL,
    trx_type            VARCHAR(30) NOT NULL, -- DEBIT, CREDIT, TRANSFER, PAYMENT, CASH_WITHDRAWAL, CASH_DEPOSIT
    category            VARCHAR(30) NOT NULL, -- Funding, Lending, Treasury, Payment, KYC, IT Control
    channel             VARCHAR(30) DEFAULT 'Teller', -- Teller, ATM, Mobile, Internet, EDC
    amount              NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) DEFAULT 'IDR',
    description         TEXT,
    account_id          INT REFERENCES cbs.accounts(account_id),
    counterpart_account VARCHAR(20),
    branch_id           INT REFERENCES cbs.branches(branch_id),
    teller_id           VARCHAR(20),
    authorization_status VARCHAR(30) DEFAULT 'Approved',
    is_suspicious       BOOLEAN DEFAULT FALSE,
    created_at          TIMESTAMP DEFAULT NOW()
);

-- Loans / Kredit
CREATE TABLE cbs.loans (
    loan_id             SERIAL PRIMARY KEY,
    loan_number         VARCHAR(20) UNIQUE NOT NULL,
    loan_type           VARCHAR(30) NOT NULL, -- KMK, KI, KPR, KKB, MULTIGUNA
    customer_id         INT REFERENCES cbs.customers(customer_id),
    account_id          INT REFERENCES cbs.accounts(account_id),
    product_id          INT REFERENCES cbs.products(product_id),
    branch_id           INT REFERENCES cbs.branches(branch_id),
    plafond             NUMERIC(18,2) NOT NULL,
    outstanding         NUMERIC(18,2) NOT NULL,
    interest_rate       NUMERIC(5,2) NOT NULL,
    tenor_months        INT NOT NULL,
    disbursement_date   DATE,
    maturity_date       DATE,
    collateral_value    NUMERIC(18,2),
    collectability      INT DEFAULT 1, -- 1=Current, 2=Special Mention, 3=Substandard, 4=Doubtful, 5=Loss
    days_past_due       INT DEFAULT 0,
    last_payment_date   DATE,
    status              VARCHAR(20) DEFAULT 'ACTIVE',
    approval_level      VARCHAR(20),
    created_at          TIMESTAMP DEFAULT NOW()
);

-- General Ledger Entries
CREATE TABLE cbs.gl_entries (
    gl_id               SERIAL PRIMARY KEY,
    gl_date             DATE NOT NULL,
    gl_account_code     VARCHAR(20) NOT NULL,
    gl_account_name     VARCHAR(100),
    voucher_number      VARCHAR(30),
    debit_amount        NUMERIC(18,2) DEFAULT 0,
    credit_amount       NUMERIC(18,2) DEFAULT 0,
    description         TEXT,
    branch_id           INT REFERENCES cbs.branches(branch_id),
    source_system       VARCHAR(30) DEFAULT 'CBS',
    posted_by           VARCHAR(50),
    posting_date        TIMESTAMP DEFAULT NOW(),
    created_at          TIMESTAMP DEFAULT NOW()
);

-- Documents / Dokumen Pendukung
CREATE TABLE cbs.documents (
    document_id         SERIAL PRIMARY KEY,
    document_number     VARCHAR(30) UNIQUE NOT NULL,
    document_type       VARCHAR(30) NOT NULL, -- LOAN_AGREEMENT, COLLATERAL_APPRAISAL, KYC_FORM, TRANSFER_SLIP
    reference_id        INT,                   -- FK flexible (loan_id, customer_id, etc.)
    reference_type      VARCHAR(30),           -- LOAN, CUSTOMER, ACCOUNT
    document_date       DATE NOT NULL,
    status              VARCHAR(20) DEFAULT 'Complete', -- Complete, Incomplete, Missing, Expired
    expiry_date         DATE,
    branch_id           INT REFERENCES cbs.branches(branch_id),
    created_by          VARCHAR(50),
    created_at          TIMESTAMP DEFAULT NOW()
);

-- Failed Login Attempts (for IT Control audit)
CREATE TABLE cbs.failed_logins (
    log_id              SERIAL PRIMARY KEY,
    user_id             VARCHAR(50) NOT NULL,
    ip_address          VARCHAR(45),
    attempt_time        TIMESTAMP NOT NULL,
    reason              VARCHAR(50),  -- WRONG_PASSWORD, LOCKED, EXPIRED
    branch_id           INT REFERENCES cbs.branches(branch_id),
    created_at          TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- SEED DATA — Branches (12 branches)
-- ============================================================================
INSERT INTO cbs.branches (branch_code, branch_name, branch_type, region, city) VALUES
('HQ-001', 'Kantor Pusat Jakarta', 'HQ', 'DKI Jakarta', 'Jakarta'),
('JKT-001', 'Cabang Jakarta Sudirman', 'BRANCH', 'DKI Jakarta', 'Jakarta'),
('JKT-002', 'Cabang Jakarta Kuningan', 'BRANCH', 'DKI Jakarta', 'Jakarta'),
('BDG-001', 'Cabang Bandung Asia Afrika', 'BRANCH', 'Jawa Barat', 'Bandung'),
('BDG-002', 'Cabang Bandung Dago', 'SUB_BRANCH', 'Jawa Barat', 'Bandung'),
('SBY-001', 'Cabang Surabaya Tunjungan', 'BRANCH', 'Jawa Timur', 'Surabaya'),
('SBY-002', 'Cabang Surabaya Darmo', 'SUB_BRANCH', 'Jawa Timur', 'Surabaya'),
('SMG-001', 'Cabang Semarang Simpang Lima', 'BRANCH', 'Jawa Tengah', 'Semarang'),
('MDN-001', 'Cabang Medan Merdeka', 'BRANCH', 'Sumatera Utara', 'Medan'),
('MKS-001', 'Cabang Makassar Losari', 'BRANCH', 'Sulawesi Selatan', 'Makassar'),
('BLI-001', 'Cabang Denpasar Renon', 'BRANCH', 'Bali', 'Denpasar'),
('PLB-001', 'Cabang Palembang Ilir', 'BRANCH', 'Sumatera Selatan', 'Palembang');

-- ============================================================================
-- SEED DATA — Products (15 products)
-- ============================================================================
INSERT INTO cbs.products (product_code, product_name, product_type, min_balance, max_balance, min_rate, max_rate, min_tenor_days, max_tenor_days, max_plafond, requires_collateral) VALUES
('SAV-001', 'Tabungan Reguler', 'FUNDING', 100000, NULL, 1.50, 3.00, NULL, NULL, NULL, FALSE),
('SAV-002', 'Tabungan Bisnis', 'FUNDING', 5000000, NULL, 2.00, 4.00, NULL, NULL, NULL, FALSE),
('DEP-001', 'Deposito Berjangka', 'FUNDING', 10000000, NULL, 4.00, 6.50, 30, 365, NULL, FALSE),
('DEP-002', 'Deposito On Call', 'FUNDING', 100000000, NULL, 3.50, 5.50, 7, 30, NULL, FALSE),
('GIR-001', 'Giro Rupiah', 'FUNDING', 1000000, NULL, 0.50, 1.50, NULL, NULL, NULL, FALSE),
('KMK-001', 'Kredit Modal Kerja', 'LENDING', NULL, NULL, 8.00, 14.00, 365, 1825, 50000000000, TRUE),
('KI-001', 'Kredit Investasi', 'LENDING', NULL, NULL, 9.00, 13.00, 365, 3650, 100000000000, TRUE),
('KPR-001', 'KPR Reguler', 'LENDING', NULL, NULL, 7.50, 12.00, 1825, 7300, 5000000000, TRUE),
('KKB-001', 'Kredit Kendaraan Bermotor', 'LENDING', NULL, NULL, 6.00, 11.00, 365, 2190, 2000000000, TRUE),
('MUL-001', 'Kredit Multiguna', 'LENDING', NULL, NULL, 10.00, 16.00, 365, 1825, 1000000000, TRUE),
('TRS-001', 'Obligasi SBN', 'TREASURY', 1000000, NULL, 5.50, 8.00, 365, 3650, NULL, FALSE),
('TRS-002', 'Repo Surat Berharga', 'TREASURY', 100000000, NULL, 4.00, 7.00, 1, 90, NULL, FALSE),
('PAY-001', 'Transfer Antar Bank', 'PAYMENT', NULL, 500000000, NULL, NULL, NULL, NULL, NULL, FALSE),
('PAY-002', 'RTGS', 'PAYMENT', 100000000, NULL, NULL, NULL, NULL, NULL, NULL, FALSE),
('PAY-003', 'SWIFT International', 'PAYMENT', NULL, NULL, NULL, NULL, NULL, NULL, NULL, FALSE);

-- ============================================================================
-- SEED DATA — Product Policies (rules)
-- ============================================================================
INSERT INTO cbs.product_policies (product_id, rule_name, rule_type, min_value, max_value, description) VALUES
-- Funding
(1, 'Saldo Minimum Tabungan', 'AMOUNT_LIMIT', 100000, NULL, 'Saldo minimum Rp 100.000'),
(3, 'Rate Deposito Range', 'RATE_LIMIT', 4.00, 6.50, 'Suku bunga deposito 4.00%-6.50%'),
(3, 'Minimum Penempatan', 'AMOUNT_LIMIT', 10000000, NULL, 'Minimum Rp 10 juta'),
-- Lending
(6, 'Plafond Maksimum KMK', 'AMOUNT_LIMIT', NULL, 50000000000, 'Plafond maks Rp 50 miliar'),
(6, 'Rate Range KMK', 'RATE_LIMIT', 8.00, 14.00, 'Suku bunga KMK 8%-14%'),
(6, 'Otorisasi Di Atas 10M', 'AUTH_LEVEL', 10000000000, NULL, 'Pencairan >10M perlu Direksi'),
(8, 'LTV Maksimum KPR', 'AMOUNT_LIMIT', NULL, 80.00, 'Loan to Value maks 80%'),
(8, 'Tenor Maksimum KPR', 'TENOR_LIMIT', NULL, 7300, 'Tenor maks 20 tahun'),
-- Payment
(13, 'Limit Transfer Harian', 'AMOUNT_LIMIT', NULL, 500000000, 'Transfer maks Rp 500 juta/hari'),
(14, 'Minimum RTGS', 'AMOUNT_LIMIT', 100000000, NULL, 'Minimum RTGS Rp 100 juta');

-- ============================================================================
-- SEED DATA — Customers (~200 via generate_series)
-- ============================================================================
INSERT INTO cbs.customers (cif_number, customer_name, customer_type, id_number, phone, risk_profile, branch_id, registration_date)
SELECT
    'CIF-' || LPAD(n::TEXT, 6, '0'),
    CASE (n % 20)
        WHEN 0 THEN 'PT Maju Jaya Sentosa'
        WHEN 1 THEN 'Ahmad Hidayat'
        WHEN 2 THEN 'Siti Nurhaliza'
        WHEN 3 THEN 'PT Bintang Utama'
        WHEN 4 THEN 'Budi Santoso'
        WHEN 5 THEN 'CV Karya Mandiri'
        WHEN 6 THEN 'Dewi Sartika'
        WHEN 7 THEN 'PT Nusantara Indah'
        WHEN 8 THEN 'Rudi Hartono'
        WHEN 9 THEN 'Ani Sulistyowati'
        WHEN 10 THEN 'PT Global Teknik'
        WHEN 11 THEN 'Hendra Kusuma'
        WHEN 12 THEN 'Rina Wulandari'
        WHEN 13 THEN 'PT Sejahtera Abadi'
        WHEN 14 THEN 'Joko Widodo'
        WHEN 15 THEN 'CV Putra Bangsa'
        WHEN 16 THEN 'Maya Angelina'
        WHEN 17 THEN 'PT Harapan Baru'
        WHEN 18 THEN 'Dimas Ramadhan'
        ELSE 'Fitri Handayani'
    END || ' - ' || n::TEXT,
    CASE WHEN n % 5 IN (0,3,5) THEN 'CORPORATE' ELSE 'INDIVIDUAL' END,
    '327105' || LPAD((n * 137 % 999999)::TEXT, 6, '0'),
    '0812' || LPAD((n * 7919 % 99999999)::TEXT, 8, '0'),
    CASE WHEN n % 10 = 0 THEN 'High Risk' WHEN n % 4 = 0 THEN 'Medium Risk' ELSE 'Low Risk' END,
    (n % 12) + 1,
    DATE '2020-01-01' + (n * 5 % 1800)
FROM generate_series(1, 200) AS n;

-- ============================================================================
-- SEED DATA — Accounts (~500)
-- ============================================================================
INSERT INTO cbs.accounts (account_number, account_type, balance, interest_rate, status, customer_id, branch_id, opened_date)
SELECT
    '100' || LPAD(n::TEXT, 7, '0'),
    CASE (n % 5) WHEN 0 THEN 'SAVINGS' WHEN 1 THEN 'CURRENT' WHEN 2 THEN 'DEPOSIT' WHEN 3 THEN 'LOAN' ELSE 'SAVINGS' END,
    ROUND((RANDOM() * 5000 + 1)::NUMERIC * 1000000, 2),
    CASE (n % 5) WHEN 2 THEN ROUND((RANDOM() * 2.5 + 4)::NUMERIC, 2) WHEN 3 THEN ROUND((RANDOM() * 6 + 8)::NUMERIC, 2) ELSE ROUND((RANDOM() * 2 + 1)::NUMERIC, 2) END,
    CASE WHEN n % 20 = 0 THEN 'DORMANT' WHEN n % 30 = 0 THEN 'CLOSED' ELSE 'ACTIVE' END,
    (n % 200) + 1,
    (n % 12) + 1,
    DATE '2019-01-01' + (n * 3 % 2000)
FROM generate_series(1, 500) AS n;

-- ============================================================================
-- SEED DATA — Transactions (~10,000 with ~200 injected anomalies)
-- ============================================================================

-- Normal Transactions (~9,800)
INSERT INTO cbs.transactions (trx_ref_number, trx_date, trx_time, trx_type, category, channel, amount, description, account_id, branch_id, teller_id, authorization_status, is_suspicious)
SELECT
    'TRX-' || TO_CHAR(DATE '2025-01-01' + (n % 365), 'YYYYMMDD') || '-' || LPAD(n::TEXT, 5, '0'),
    DATE '2025-01-01' + (n % 365),
    (TIME '08:00:00' + (n % 36000) * INTERVAL '1 second'),
    CASE (n % 6) WHEN 0 THEN 'CREDIT' WHEN 1 THEN 'DEBIT' WHEN 2 THEN 'TRANSFER' WHEN 3 THEN 'PAYMENT' WHEN 4 THEN 'CASH_DEPOSIT' ELSE 'CASH_WITHDRAWAL' END,
    CASE (n % 6) WHEN 0 THEN 'Funding' WHEN 1 THEN 'Lending' WHEN 2 THEN 'Payment' WHEN 3 THEN 'Treasury' WHEN 4 THEN 'Funding' ELSE 'Payment' END,
    CASE (n % 5) WHEN 0 THEN 'Teller' WHEN 1 THEN 'ATM' WHEN 2 THEN 'Mobile' WHEN 3 THEN 'Internet' ELSE 'EDC' END,
    ROUND((RANDOM() * 200 + 0.5)::NUMERIC * 1000000, 2),
    'Transaksi operasional normal #' || n,
    (n % 500) + 1,
    (n % 12) + 1,
    'TLR-' || LPAD(((n % 50) + 1)::TEXT, 3, '0'),
    'Approved',
    FALSE
FROM generate_series(1, 9800) AS n;

-- Injected Anomalies (~200 suspicious transactions)
INSERT INTO cbs.transactions (trx_ref_number, trx_date, trx_time, trx_type, category, channel, amount, description, account_id, branch_id, teller_id, authorization_status, is_suspicious)
SELECT
    'TRX-ANOM-' || LPAD(n::TEXT, 4, '0'),
    DATE '2025-01-01' + (n * 3 % 365),
    -- After-hours transactions (anomaly)
    CASE WHEN n % 3 = 0 THEN TIME '23:30:00' + (n % 30) * INTERVAL '1 minute'
         WHEN n % 3 = 1 THEN TIME '02:15:00' + (n % 45) * INTERVAL '1 minute'
         ELSE TIME '04:00:00' + (n % 60) * INTERVAL '1 minute' END,
    CASE WHEN n % 4 = 0 THEN 'TRANSFER' ELSE 'CASH_WITHDRAWAL' END,
    CASE (n % 6) WHEN 0 THEN 'Funding' WHEN 1 THEN 'Lending' WHEN 2 THEN 'Payment' WHEN 3 THEN 'Treasury' WHEN 4 THEN 'KYC' ELSE 'IT Control' END,
    CASE WHEN n % 2 = 0 THEN 'Internet' ELSE 'Mobile' END,
    -- Large round amounts (anomaly pattern)
    CASE WHEN n % 5 = 0 THEN 500000000.00  -- Rp 500 juta (round)
         WHEN n % 5 = 1 THEN 1000000000.00 -- Rp 1 miliar (round)
         WHEN n % 5 = 2 THEN 750000000.00  -- Rp 750 juta
         WHEN n % 5 = 3 THEN 250000000.00  -- Rp 250 juta
         ELSE 2000000000.00 END,            -- Rp 2 miliar
    CASE WHEN n % 4 = 0 THEN 'Transfer besar ke rekening baru di luar jam kerja'
         WHEN n % 4 = 1 THEN 'Penarikan tunai signifikan melebihi pola historis'
         WHEN n % 4 = 2 THEN 'Pencairan kredit tanpa dokumen lengkap'
         ELSE 'Transaksi split untuk menghindari batas otorisasi' END,
    (n % 500) + 1,
    (n % 12) + 1,
    'TLR-' || LPAD(((n % 10) + 1)::TEXT, 3, '0'),
    CASE WHEN n % 3 = 0 THEN 'Override' WHEN n % 3 = 1 THEN 'Pending' ELSE 'Approved' END,
    TRUE
FROM generate_series(1, 200) AS n;

-- ============================================================================
-- SEED DATA — Loans (~150)
-- ============================================================================
INSERT INTO cbs.loans (loan_number, loan_type, customer_id, account_id, product_id, branch_id, plafond, outstanding, interest_rate, tenor_months, disbursement_date, maturity_date, collateral_value, collectability, days_past_due, last_payment_date, status, approval_level)
SELECT
    'LN-' || LPAD(n::TEXT, 6, '0'),
    CASE (n % 5) WHEN 0 THEN 'KMK' WHEN 1 THEN 'KI' WHEN 2 THEN 'KPR' WHEN 3 THEN 'KKB' ELSE 'MULTIGUNA' END,
    (n % 200) + 1,
    (n % 500) + 1,
    CASE (n % 5) WHEN 0 THEN 6 WHEN 1 THEN 7 WHEN 2 THEN 8 WHEN 3 THEN 9 ELSE 10 END,
    (n % 12) + 1,
    ROUND((RANDOM() * 4000 + 100)::NUMERIC * 1000000, 2),
    ROUND((RANDOM() * 3000 + 50)::NUMERIC * 1000000, 2),
    ROUND((RANDOM() * 6 + 7)::NUMERIC, 2),
    CASE (n % 5) WHEN 2 THEN 240 WHEN 3 THEN 60 ELSE 36 + (n % 24) END,
    DATE '2022-01-01' + (n * 7 % 1000),
    DATE '2025-01-01' + (n * 30 % 3000),
    ROUND((RANDOM() * 5000 + 200)::NUMERIC * 1000000, 2),
    CASE WHEN n % 15 = 0 THEN 3 WHEN n % 10 = 0 THEN 2 WHEN n % 20 = 0 THEN 4 WHEN n % 30 = 0 THEN 5 ELSE 1 END,
    CASE WHEN n % 15 = 0 THEN 120 + (n % 60) WHEN n % 10 = 0 THEN 31 + (n % 30) WHEN n % 20 = 0 THEN 180 + (n % 90) ELSE 0 END,
    CASE WHEN n % 15 = 0 THEN DATE '2024-08-01' ELSE DATE '2025-09-01' - (n % 30) END,
    CASE WHEN n % 20 = 0 THEN 'NPL' WHEN n % 30 = 0 THEN 'RESTRUCTURED' ELSE 'ACTIVE' END,
    CASE WHEN ROUND((RANDOM() * 4000 + 100)::NUMERIC * 1000000, 2) > 10000000000 THEN 'DIREKSI' ELSE 'KOMITE_KREDIT' END
FROM generate_series(1, 150) AS n;

-- ============================================================================
-- SEED DATA — GL Entries (~3,000 with some intentional gaps in voucher numbers)
-- ============================================================================
INSERT INTO cbs.gl_entries (gl_date, gl_account_code, gl_account_name, voucher_number, debit_amount, credit_amount, description, branch_id, posted_by)
SELECT
    DATE '2025-01-01' + (n % 365),
    CASE (n % 8)
        WHEN 0 THEN '1101' WHEN 1 THEN '1201' WHEN 2 THEN '2101' WHEN 3 THEN '2201'
        WHEN 4 THEN '3101' WHEN 5 THEN '4101' WHEN 6 THEN '5101' ELSE '6101' END,
    CASE (n % 8)
        WHEN 0 THEN 'Kas' WHEN 1 THEN 'Penempatan Bank Lain' WHEN 2 THEN 'Dana Pihak Ketiga'
        WHEN 3 THEN 'Pinjaman Diterima' WHEN 4 THEN 'Modal Disetor' WHEN 5 THEN 'Pendapatan Bunga'
        WHEN 6 THEN 'Beban Bunga' ELSE 'Beban Operasional' END,
    -- Intentionally skip some voucher numbers (gaps for audit testing)
    CASE WHEN n % 50 = 0 THEN NULL  -- Missing voucher (anomaly)
         ELSE 'VCR-' || TO_CHAR(DATE '2025-01-01' + (n % 365), 'YYYYMMDD') || '-' || LPAD(n::TEXT, 5, '0') END,
    CASE WHEN n % 2 = 0 THEN ROUND((RANDOM() * 500 + 1)::NUMERIC * 1000000, 2) ELSE 0 END,
    CASE WHEN n % 2 = 1 THEN ROUND((RANDOM() * 500 + 1)::NUMERIC * 1000000, 2) ELSE 0 END,
    'Jurnal posting GL #' || n,
    (n % 12) + 1,
    'USR-' || LPAD(((n % 20) + 1)::TEXT, 3, '0')
FROM generate_series(1, 3000) AS n;

-- ============================================================================
-- SEED DATA — Documents (~300, some intentionally incomplete/missing)
-- ============================================================================
INSERT INTO cbs.documents (document_number, document_type, reference_id, reference_type, document_date, status, expiry_date, branch_id, created_by)
SELECT
    'DOC-' || LPAD(n::TEXT, 5, '0'),
    CASE (n % 4) WHEN 0 THEN 'LOAN_AGREEMENT' WHEN 1 THEN 'COLLATERAL_APPRAISAL' WHEN 2 THEN 'KYC_FORM' ELSE 'TRANSFER_SLIP' END,
    (n % 150) + 1,
    CASE WHEN n % 3 = 0 THEN 'LOAN' WHEN n % 3 = 1 THEN 'CUSTOMER' ELSE 'ACCOUNT' END,
    DATE '2024-01-01' + (n * 3 % 600),
    CASE WHEN n % 15 = 0 THEN 'Missing'
         WHEN n % 12 = 0 THEN 'Incomplete'
         WHEN n % 20 = 0 THEN 'Expired'
         ELSE 'Complete' END,
    CASE WHEN n % 3 = 0 THEN DATE '2025-06-01' + (n % 365) ELSE NULL END,
    (n % 12) + 1,
    'USR-' || LPAD(((n % 20) + 1)::TEXT, 3, '0')
FROM generate_series(1, 300) AS n;

-- ============================================================================
-- SEED DATA — Failed Logins (~100, some branches with excessive failures)
-- ============================================================================
INSERT INTO cbs.failed_logins (user_id, ip_address, attempt_time, reason, branch_id)
SELECT
    'USR-' || LPAD(((n % 20) + 1)::TEXT, 3, '0'),
    '192.168.' || (n % 255) || '.' || ((n * 7) % 255),
    TIMESTAMP '2025-01-01 00:00:00' + (n * 3600 + n * 137) * INTERVAL '1 second',
    CASE (n % 3) WHEN 0 THEN 'WRONG_PASSWORD' WHEN 1 THEN 'LOCKED' ELSE 'EXPIRED' END,
    CASE WHEN n % 5 = 0 THEN 2  -- Concentrate failures on Jakarta Sudirman
         WHEN n % 7 = 0 THEN 6  -- and Surabaya Tunjungan
         ELSE (n % 12) + 1 END
FROM generate_series(1, 100) AS n;

-- ============================================================================
-- INDEXES
-- ============================================================================
CREATE INDEX idx_trx_date ON cbs.transactions(trx_date);
CREATE INDEX idx_trx_category ON cbs.transactions(category);
CREATE INDEX idx_trx_suspicious ON cbs.transactions(is_suspicious);
CREATE INDEX idx_trx_account ON cbs.transactions(account_id);
CREATE INDEX idx_trx_branch ON cbs.transactions(branch_id);
CREATE INDEX idx_loan_customer ON cbs.loans(customer_id);
CREATE INDEX idx_loan_collectability ON cbs.loans(collectability);
CREATE INDEX idx_gl_date ON cbs.gl_entries(gl_date);
CREATE INDEX idx_gl_voucher ON cbs.gl_entries(voucher_number);
CREATE INDEX idx_doc_status ON cbs.documents(status);
CREATE INDEX idx_failed_login_time ON cbs.failed_logins(attempt_time);

-- Done
SELECT 'CBS Simulator initialized successfully' AS status,
       (SELECT COUNT(*) FROM cbs.transactions) AS transactions,
       (SELECT COUNT(*) FROM cbs.customers) AS customers,
       (SELECT COUNT(*) FROM cbs.accounts) AS accounts,
       (SELECT COUNT(*) FROM cbs.loans) AS loans,
       (SELECT COUNT(*) FROM cbs.gl_entries) AS gl_entries;
