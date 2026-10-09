-- ============================================================================
-- Migration 007: Composite Unique Keys for Multi-Source Silver & Gold Tables
-- Enables multi-source conflict resolution for set-based upsert operations
-- ============================================================================

DO $$
BEGIN
    -- 0. Ensure source_id is VARCHAR(64) across all tables to allow both UUIDs and strings like 'cbs_simulator'
    BEGIN
        ALTER TABLE bronze.registered_sources ALTER COLUMN source_id TYPE VARCHAR(64) USING source_id::VARCHAR;
        ALTER TABLE silver.external_source_data ALTER COLUMN source_id TYPE VARCHAR(64) USING source_id::VARCHAR;
        ALTER TABLE gold.fact_external_audit_data ALTER COLUMN source_id TYPE VARCHAR(64) USING source_id::VARCHAR;
        ALTER TABLE gold.ai_training_pool ALTER COLUMN source_id TYPE VARCHAR(64) USING source_id::VARCHAR;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;

    -- 1. Silver Transactions
    ALTER TABLE silver.trx_cleaned DROP CONSTRAINT IF EXISTS trx_cleaned_trx_ref_number_key;
    CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_trx_source_ref
    ON silver.trx_cleaned (source_id, trx_ref_number);

    -- 2. Silver Accounts
    ALTER TABLE silver.acct_cleaned DROP CONSTRAINT IF EXISTS acct_cleaned_account_number_key;
    CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_acct_source_num
    ON silver.acct_cleaned (source_id, account_number);

    -- 3. Silver Customers
    ALTER TABLE silver.cust_cleaned DROP CONSTRAINT IF EXISTS cust_cleaned_cif_number_key;
    CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_cust_source_cif
    ON silver.cust_cleaned (source_id, cif_number);

    -- 4. Silver Loans
    ALTER TABLE silver.loan_cleaned DROP CONSTRAINT IF EXISTS loan_cleaned_loan_number_key;
    CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_loan_source_num
    ON silver.loan_cleaned (source_id, loan_number);

    -- 5. Silver GL Entries
    CREATE UNIQUE INDEX IF NOT EXISTS uq_silver_gl_source_id
    ON silver.gl_cleaned (source_id, gl_id);

    -- 6. Gold Loans
    CREATE UNIQUE INDEX IF NOT EXISTS uq_gold_loans_source_num
    ON gold.fact_loans (source_id, loan_number);

    -- 7. Gold GL Entries
    CREATE UNIQUE INDEX IF NOT EXISTS uq_gold_gl_source_id
    ON gold.fact_gl_entries (source_id, gl_id);

    -- 8. Gold Dim Branches
    CREATE UNIQUE INDEX IF NOT EXISTS uq_gold_dim_branches_source_code
    ON gold.dim_branches (source_id, branch_code);

EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Migration 007 notices: %', SQLERRM;
END $$;
