"""
Data Hub API — Transform Service
Bronze → Silver (clean + deduplicate)
Silver → Gold (enrich + aggregate)
"""
import math
from sqlalchemy import text


def run_bronze_to_silver(engine):
    """Transform Bronze zone raw data into Silver zone cleaned data."""
    print("[Transform] Starting Bronze → Silver...")

    with engine.connect() as conn:
        # ── Transactions: clean, deduplicate, add derived columns ─────────
        conn.execute(text("TRUNCATE TABLE silver.trx_cleaned"))
        conn.execute(text("""
            INSERT INTO silver.trx_cleaned
                (trx_ref_number, trx_date, trx_hour, trx_day_of_week, trx_type, category,
                 channel, amount, amount_millions, description, account_id, branch_id,
                 teller_id, authorization_status, is_suspicious, is_round_amount, is_after_hours)
            SELECT DISTINCT ON (trx_ref_number)
                trx_ref_number,
                trx_date,
                EXTRACT(HOUR FROM trx_time)::INT,
                EXTRACT(DOW FROM trx_date)::INT,
                trx_type,
                category,
                channel,
                amount,
                ROUND(amount / 1000000.0, 4),
                description,
                account_id,
                branch_id,
                teller_id,
                authorization_status,
                COALESCE(is_suspicious, FALSE),
                -- Round amount detection: divisible by 100M
                CASE WHEN amount > 0 AND amount::BIGINT % 100000000 = 0 THEN TRUE ELSE FALSE END,
                -- After hours: before 7 AM or after 8 PM
                CASE WHEN EXTRACT(HOUR FROM trx_time) < 7 OR EXTRACT(HOUR FROM trx_time) >= 20 THEN TRUE ELSE FALSE END
            FROM bronze.cb_transactions
            WHERE trx_ref_number IS NOT NULL
            ORDER BY trx_ref_number, _loaded_at DESC
        """))

        # ── Accounts ──────────────────────────────────────────────────────
        conn.execute(text("TRUNCATE TABLE silver.acct_cleaned"))
        conn.execute(text("""
            INSERT INTO silver.acct_cleaned
                (account_id, account_number, account_type, balance, interest_rate, status, customer_id, branch_id, opened_date)
            SELECT DISTINCT ON (account_id)
                account_id, account_number, account_type, balance, interest_rate, status, customer_id, branch_id, opened_date
            FROM bronze.cb_accounts
            WHERE account_id IS NOT NULL
            ORDER BY account_id, _loaded_at DESC
        """))

        # ── Customers ─────────────────────────────────────────────────────
        conn.execute(text("TRUNCATE TABLE silver.cust_cleaned"))
        conn.execute(text("""
            INSERT INTO silver.cust_cleaned
                (customer_id, cif_number, customer_name, customer_type, risk_profile, branch_id, is_active)
            SELECT DISTINCT ON (customer_id)
                customer_id, cif_number, customer_name, customer_type, risk_profile, branch_id, is_active
            FROM bronze.cb_customers
            WHERE customer_id IS NOT NULL
            ORDER BY customer_id, _loaded_at DESC
        """))

        # ── Loans ─────────────────────────────────────────────────────────
        conn.execute(text("TRUNCATE TABLE silver.loan_cleaned"))
        conn.execute(text("""
            INSERT INTO silver.loan_cleaned
                (loan_id, loan_number, loan_type, customer_id, branch_id, product_id,
                 plafond, outstanding, interest_rate, tenor_months, disbursement_date,
                 maturity_date, collateral_value, ltv_ratio, collectability, days_past_due, status)
            SELECT DISTINCT ON (loan_id)
                loan_id, loan_number, loan_type, customer_id, branch_id, product_id,
                plafond, outstanding, interest_rate, tenor_months, disbursement_date,
                maturity_date, collateral_value,
                CASE WHEN collateral_value > 0 THEN ROUND((outstanding / collateral_value * 100)::NUMERIC, 2) ELSE NULL END,
                collectability, days_past_due, status
            FROM bronze.cb_loans
            WHERE loan_id IS NOT NULL
            ORDER BY loan_id, _loaded_at DESC
        """))

        # ── GL Entries ────────────────────────────────────────────────────
        conn.execute(text("TRUNCATE TABLE silver.gl_cleaned"))
        conn.execute(text("""
            INSERT INTO silver.gl_cleaned
                (gl_id, gl_date, gl_account_code, gl_account_name, voucher_number,
                 debit_amount, credit_amount, net_amount, description, branch_id, posted_by, has_voucher)
            SELECT DISTINCT ON (gl_id)
                gl_id, gl_date, gl_account_code, gl_account_name, voucher_number,
                COALESCE(debit_amount, 0), COALESCE(credit_amount, 0),
                COALESCE(debit_amount, 0) - COALESCE(credit_amount, 0),
                description, branch_id, posted_by,
                CASE WHEN voucher_number IS NOT NULL AND voucher_number != '' THEN TRUE ELSE FALSE END
            FROM bronze.cb_gl_entries
            WHERE gl_id IS NOT NULL
            ORDER BY gl_id, _loaded_at DESC
        """))

        # ── Branches reference ────────────────────────────────────────────
        conn.execute(text("TRUNCATE TABLE silver.ref_branches"))
        conn.execute(text("""
            INSERT INTO silver.ref_branches (branch_id, branch_code, branch_name, branch_type, region, city)
            SELECT DISTINCT ON (branch_id)
                branch_id, branch_code, branch_name, branch_type, region, city
            FROM bronze.cb_branches
            WHERE branch_id IS NOT NULL
            ORDER BY branch_id, _loaded_at DESC
        """))

        conn.commit()
    print("[Transform] Bronze → Silver completed.")


def run_silver_to_gold(engine):
    """Transform Silver zone cleaned data into Gold zone enriched data."""
    print("[Transform] Starting Silver → Gold...")

    with engine.connect() as conn:
        # ── Dimension: Branches (enriched) ────────────────────────────────
        conn.execute(text("TRUNCATE TABLE gold.dim_branches"))
        conn.execute(text("""
            INSERT INTO gold.dim_branches
                (branch_id, branch_code, branch_name, branch_type, region, city,
                 total_accounts, total_customers, total_trx_volume, anomaly_count)
            SELECT
                b.branch_id, b.branch_code, b.branch_name, b.branch_type, b.region, b.city,
                COALESCE(a.cnt, 0),
                COALESCE(c.cnt, 0),
                COALESCE(t.total, 0),
                COALESCE(t.anomalies, 0)
            FROM silver.ref_branches b
            LEFT JOIN (SELECT branch_id, COUNT(*) AS cnt FROM silver.acct_cleaned GROUP BY branch_id) a ON a.branch_id = b.branch_id
            LEFT JOIN (SELECT branch_id, COUNT(*) AS cnt FROM silver.cust_cleaned GROUP BY branch_id) c ON c.branch_id = b.branch_id
            LEFT JOIN (
                SELECT branch_id, SUM(amount) AS total, SUM(CASE WHEN is_suspicious THEN 1 ELSE 0 END) AS anomalies
                FROM silver.trx_cleaned GROUP BY branch_id
            ) t ON t.branch_id = b.branch_id
        """))

        # ── Fact: Transactions (enriched with customer + branch names) ────
        conn.execute(text("TRUNCATE TABLE gold.fact_transactions"))
        conn.execute(text("""
            INSERT INTO gold.fact_transactions
                (trx_ref_number, trx_date, trx_hour, trx_day_of_week, trx_type, category, channel,
                 amount, amount_millions, branch_id, branch_name, region,
                 customer_name, customer_type, customer_risk, account_type,
                 authorization_status, is_suspicious, is_round_amount, is_after_hours,
                 benford_first_digit)
            SELECT
                t.trx_ref_number, t.trx_date, t.trx_hour, t.trx_day_of_week, t.trx_type, t.category, t.channel,
                t.amount, t.amount_millions, t.branch_id,
                b.branch_name, b.region,
                cu.customer_name, cu.customer_type, cu.risk_profile,
                ac.account_type,
                t.authorization_status, t.is_suspicious, t.is_round_amount, t.is_after_hours,
                CASE WHEN t.amount > 0 THEN SUBSTRING(CAST(FLOOR(t.amount) AS TEXT) FROM 1 FOR 1)::INT ELSE NULL END
            FROM silver.trx_cleaned t
            LEFT JOIN silver.ref_branches b ON b.branch_id = t.branch_id
            LEFT JOIN silver.acct_cleaned ac ON ac.account_id = t.account_id
            LEFT JOIN silver.cust_cleaned cu ON cu.customer_id = ac.customer_id
        """))

        # ── Fact: Loans (enriched with policy compliance flags) ───────────
        conn.execute(text("TRUNCATE TABLE gold.fact_loans"))
        conn.execute(text("""
            INSERT INTO gold.fact_loans
                (loan_id, loan_number, loan_type, customer_name, customer_type,
                 branch_name, region, product_name, plafond, outstanding,
                 interest_rate, tenor_months, collateral_value, ltv_ratio,
                 collectability, days_past_due, status,
                 rate_within_policy, plafond_within_policy, tenor_within_policy)
            SELECT
                l.loan_id, l.loan_number, l.loan_type,
                cu.customer_name, cu.customer_type,
                b.branch_name, b.region,
                p.product_name,
                l.plafond, l.outstanding, l.interest_rate, l.tenor_months,
                l.collateral_value, l.ltv_ratio,
                l.collectability, l.days_past_due, l.status,
                -- Policy checks
                CASE WHEN p.min_rate IS NOT NULL AND p.max_rate IS NOT NULL
                     THEN l.interest_rate BETWEEN p.min_rate AND p.max_rate ELSE TRUE END,
                CASE WHEN p.max_plafond IS NOT NULL
                     THEN l.plafond <= p.max_plafond ELSE TRUE END,
                CASE WHEN p.max_tenor IS NOT NULL
                     THEN l.tenor_months * 30 <= p.max_tenor ELSE TRUE END
            FROM silver.loan_cleaned l
            LEFT JOIN silver.cust_cleaned cu ON cu.customer_id = l.customer_id
            LEFT JOIN silver.ref_branches b ON b.branch_id = l.branch_id
            LEFT JOIN (
                SELECT product_id, product_name, min_rate, max_rate, max_plafond,
                       MAX(CASE WHEN rule_type = 'TENOR_LIMIT' THEN max_value END) AS max_tenor
                FROM bronze.cb_products p
                LEFT JOIN bronze.cb_product_policies pp ON pp.product_id = p.product_id
                GROUP BY p.product_id, product_name, min_rate, max_rate, max_plafond
            ) p ON p.product_id = l.product_id
        """))

        # ── Fact: GL Entries (with gap detection) ─────────────────────────
        conn.execute(text("TRUNCATE TABLE gold.fact_gl_entries"))
        conn.execute(text("""
            INSERT INTO gold.fact_gl_entries
                (gl_id, gl_date, gl_account_code, gl_account_name, voucher_number,
                 debit_amount, credit_amount, net_amount, branch_name, has_voucher, has_gap)
            SELECT
                g.gl_id, g.gl_date, g.gl_account_code, g.gl_account_name, g.voucher_number,
                g.debit_amount, g.credit_amount, g.net_amount,
                b.branch_name,
                g.has_voucher,
                NOT g.has_voucher
            FROM silver.gl_cleaned g
            LEFT JOIN silver.ref_branches b ON b.branch_id = g.branch_id
        """))

        # ── Log freshness ─────────────────────────────────────────────────
        conn.execute(text("""
            INSERT INTO gold.data_freshness_log (source_name, table_name, records_count, is_new)
            VALUES
                ('etl_pipeline', 'gold.dim_branches', (SELECT COUNT(*) FROM gold.dim_branches), TRUE),
                ('etl_pipeline', 'gold.fact_transactions', (SELECT COUNT(*) FROM gold.fact_transactions), TRUE),
                ('etl_pipeline', 'gold.fact_loans', (SELECT COUNT(*) FROM gold.fact_loans), TRUE),
                ('etl_pipeline', 'gold.fact_gl_entries', (SELECT COUNT(*) FROM gold.fact_gl_entries), TRUE)
        """))

        conn.commit()

    print("[Transform] Silver → Gold completed.")
