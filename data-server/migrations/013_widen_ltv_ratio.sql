-- ============================================================================
-- Migration 013: Widen ltv_ratio Precision for High-LTV Audit Anomalies
-- Allows loans with small collateral values (LTV >= 1000%) to load without
-- numeric(5,2) overflow in silver.loan_cleaned and gold.fact_loans.
-- ============================================================================

ALTER TABLE silver.loan_cleaned ALTER COLUMN ltv_ratio TYPE NUMERIC(10,2);
ALTER TABLE gold.fact_loans    ALTER COLUMN ltv_ratio TYPE NUMERIC(10,2);

