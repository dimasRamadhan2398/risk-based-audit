"""
Data Hub API — Silver Zone Set-Based Transforms
High-speed in-database incremental merge (Bronze → Silver).
Never truncates tables; updates existing rows only if _row_hash has changed.
Tombstones records if source delete was detected (_op = 'D').
"""
import re
from sqlalchemy import text
from core.identifiers import quote_identifier, sanitize_table_name

def run_bronze_to_silver_cbs(engine, batch_id: str = None) -> dict:
    """
    Incrementally merge CBS Bronze tables into Silver zone using set-based SQL UPSERT.
    No TRUNCATE, no iterrows(), pure PostgreSQL query optimization.
    """
    print(f"[Transform:Silver] Merging CBS Bronze to Silver (batch_id={batch_id or 'all'})...")
    batch_filter = "AND b._batch_id = :bid" if batch_id else ""
    params = {"bid": batch_id} if batch_id else {}

    results = {}
    with engine.begin() as conn:
        # 1. Transactions
        r = conn.execute(text(f"""
            INSERT INTO silver.trx_cleaned (
                source_id, trx_ref_number, trx_date, trx_hour, trx_day_of_week, trx_type,
                category, channel, amount, amount_millions, description, account_id,
                branch_id, teller_id, authorization_status, is_suspicious,
                is_round_amount, is_after_hours, _row_hash, _batch_id, is_deleted,
                deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON (COALESCE(b._source, 'cbs_simulator'), b.trx_ref_number)
                COALESCE(b._source, 'cbs_simulator') AS source_id,
                b.trx_ref_number,
                b.trx_date,
                EXTRACT(HOUR FROM b.trx_time)::INT,
                EXTRACT(DOW FROM b.trx_date)::INT,
                b.trx_type,
                b.category,
                b.channel,
                b.amount,
                ROUND(b.amount / 1000000.0, 4),
                b.description,
                b.account_id,
                b.branch_id,
                b.teller_id,
                b.authorization_status,
                COALESCE(b.is_suspicious, FALSE),
                -- Round amount flag: divisible by 100M
                CASE WHEN b.amount > 0 AND b.amount::BIGINT % 100000000 = 0 THEN TRUE ELSE FALSE END,
                -- After hours flag: outside 07:00 - 20:00
                CASE WHEN EXTRACT(HOUR FROM b.trx_time) < 7 OR EXTRACT(HOUR FROM b.trx_time) >= 20 THEN TRUE ELSE FALSE END,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.cb_transactions b
            WHERE b.trx_ref_number IS NOT NULL {batch_filter}
            ORDER BY COALESCE(b._source, 'cbs_simulator'), b.trx_ref_number, b._loaded_at DESC
            ON CONFLICT (source_id, trx_ref_number)
            DO UPDATE SET
                trx_date = EXCLUDED.trx_date,
                trx_hour = EXCLUDED.trx_hour,
                trx_day_of_week = EXCLUDED.trx_day_of_week,
                trx_type = EXCLUDED.trx_type,
                category = EXCLUDED.category,
                channel = EXCLUDED.channel,
                amount = EXCLUDED.amount,
                amount_millions = EXCLUDED.amount_millions,
                description = EXCLUDED.description,
                account_id = EXCLUDED.account_id,
                branch_id = EXCLUDED.branch_id,
                teller_id = EXCLUDED.teller_id,
                authorization_status = EXCLUDED.authorization_status,
                is_suspicious = EXCLUDED.is_suspicious,
                is_round_amount = EXCLUDED.is_round_amount,
                is_after_hours = EXCLUDED.is_after_hours,
                _row_hash = EXCLUDED._row_hash,
                _batch_id = EXCLUDED._batch_id,
                is_deleted = EXCLUDED.is_deleted,
                deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.trx_cleaned.deleted_detected_at END,
                cleaned_at = NOW()
            WHERE silver.trx_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash
               OR silver.trx_cleaned.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
        """), params)
        results["trx_cleaned"] = r.rowcount

        # 2. Accounts
        r = conn.execute(text(f"""
            INSERT INTO silver.acct_cleaned (
                source_id, account_id, account_number, account_type, balance,
                interest_rate, status, customer_id, branch_id, opened_date,
                _row_hash, _batch_id, is_deleted, deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON (COALESCE(b._source, 'cbs_simulator'), b.account_number)
                COALESCE(b._source, 'cbs_simulator') AS source_id,
                b.account_id,
                b.account_number,
                b.account_type,
                b.balance,
                b.interest_rate,
                b.status,
                b.customer_id,
                b.branch_id,
                b.opened_date,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.cb_accounts b
            WHERE b.account_number IS NOT NULL {batch_filter}
            ORDER BY COALESCE(b._source, 'cbs_simulator'), b.account_number, b._loaded_at DESC
            ON CONFLICT (source_id, account_number)
            DO UPDATE SET
                account_id = EXCLUDED.account_id,
                account_type = EXCLUDED.account_type,
                balance = EXCLUDED.balance,
                interest_rate = EXCLUDED.interest_rate,
                status = EXCLUDED.status,
                customer_id = EXCLUDED.customer_id,
                branch_id = EXCLUDED.branch_id,
                opened_date = EXCLUDED.opened_date,
                _row_hash = EXCLUDED._row_hash,
                _batch_id = EXCLUDED._batch_id,
                is_deleted = EXCLUDED.is_deleted,
                deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.acct_cleaned.deleted_detected_at END,
                cleaned_at = NOW()
            WHERE silver.acct_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash
               OR silver.acct_cleaned.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
        """), params)
        results["acct_cleaned"] = r.rowcount

        # 3. Customers
        r = conn.execute(text(f"""
            INSERT INTO silver.cust_cleaned (
                source_id, customer_id, cif_number, customer_name, customer_type,
                risk_profile, branch_id, is_active, _row_hash, _batch_id, is_deleted,
                deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON (COALESCE(b._source, 'cbs_simulator'), b.cif_number)
                COALESCE(b._source, 'cbs_simulator') AS source_id,
                b.customer_id,
                b.cif_number,
                b.customer_name,
                b.customer_type,
                b.risk_profile,
                b.branch_id,
                b.is_active,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.cb_customers b
            WHERE b.cif_number IS NOT NULL {batch_filter}
            ORDER BY COALESCE(b._source, 'cbs_simulator'), b.cif_number, b._loaded_at DESC
            ON CONFLICT (source_id, cif_number)
            DO UPDATE SET
                customer_id = EXCLUDED.customer_id,
                customer_name = EXCLUDED.customer_name,
                customer_type = EXCLUDED.customer_type,
                risk_profile = EXCLUDED.risk_profile,
                branch_id = EXCLUDED.branch_id,
                is_active = EXCLUDED.is_active,
                _row_hash = EXCLUDED._row_hash,
                _batch_id = EXCLUDED._batch_id,
                is_deleted = EXCLUDED.is_deleted,
                deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.cust_cleaned.deleted_detected_at END,
                cleaned_at = NOW()
            WHERE silver.cust_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash
               OR silver.cust_cleaned.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
        """), params)
        results["cust_cleaned"] = r.rowcount

        # 4. Loans
        r = conn.execute(text(f"""
            INSERT INTO silver.loan_cleaned (
                source_id, loan_id, loan_number, loan_type, customer_id, branch_id,
                product_id, plafond, outstanding, interest_rate, tenor_months,
                disbursement_date, maturity_date, collateral_value, ltv_ratio,
                collectability, days_past_due, status, _row_hash, _batch_id, is_deleted,
                deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON (COALESCE(b._source, 'cbs_simulator'), b.loan_number)
                COALESCE(b._source, 'cbs_simulator') AS source_id,
                b.loan_id,
                b.loan_number,
                b.loan_type,
                b.customer_id,
                b.branch_id,
                b.product_id,
                b.plafond,
                b.outstanding,
                b.interest_rate,
                b.tenor_months,
                b.disbursement_date,
                b.maturity_date,
                b.collateral_value,
                CASE WHEN b.collateral_value > 0 THEN ROUND((b.outstanding / b.collateral_value * 100)::NUMERIC, 2) ELSE NULL END,
                b.collectability,
                b.days_past_due,
                b.status,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.cb_loans b
            WHERE b.loan_number IS NOT NULL {batch_filter}
            ORDER BY COALESCE(b._source, 'cbs_simulator'), b.loan_number, b._loaded_at DESC
            ON CONFLICT (source_id, loan_number)
            DO UPDATE SET
                loan_id = EXCLUDED.loan_id,
                loan_type = EXCLUDED.loan_type,
                customer_id = EXCLUDED.customer_id,
                branch_id = EXCLUDED.branch_id,
                product_id = EXCLUDED.product_id,
                plafond = EXCLUDED.plafond,
                outstanding = EXCLUDED.outstanding,
                interest_rate = EXCLUDED.interest_rate,
                tenor_months = EXCLUDED.tenor_months,
                disbursement_date = EXCLUDED.disbursement_date,
                maturity_date = EXCLUDED.maturity_date,
                collateral_value = EXCLUDED.collateral_value,
                ltv_ratio = EXCLUDED.ltv_ratio,
                collectability = EXCLUDED.collectability,
                days_past_due = EXCLUDED.days_past_due,
                status = EXCLUDED.status,
                _row_hash = EXCLUDED._row_hash,
                _batch_id = EXCLUDED._batch_id,
                is_deleted = EXCLUDED.is_deleted,
                deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.loan_cleaned.deleted_detected_at END,
                cleaned_at = NOW()
            WHERE silver.loan_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash
               OR silver.loan_cleaned.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
        """), params)
        results["loan_cleaned"] = r.rowcount

        # 5. GL Entries
        r = conn.execute(text(f"""
            INSERT INTO silver.gl_cleaned (
                source_id, gl_id, gl_date, gl_account_code, gl_account_name,
                voucher_number, debit_amount, credit_amount, net_amount,
                description, branch_id, posted_by, has_voucher,
                _row_hash, _batch_id, is_deleted, deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON (COALESCE(b._source, 'cbs_simulator'), b.gl_id)
                COALESCE(b._source, 'cbs_simulator') AS source_id,
                b.gl_id,
                b.gl_date,
                b.gl_account_code,
                b.gl_account_name,
                b.voucher_number,
                COALESCE(b.debit_amount, 0),
                COALESCE(b.credit_amount, 0),
                COALESCE(b.debit_amount, 0) - COALESCE(b.credit_amount, 0),
                b.description,
                b.branch_id,
                b.posted_by,
                CASE WHEN b.voucher_number IS NOT NULL AND b.voucher_number != '' THEN TRUE ELSE FALSE END,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.cb_gl_entries b
            WHERE b.gl_id IS NOT NULL {batch_filter}
            ORDER BY COALESCE(b._source, 'cbs_simulator'), b.gl_id, b._loaded_at DESC
            ON CONFLICT (source_id, gl_id)
            DO UPDATE SET
                gl_date = EXCLUDED.gl_date,
                gl_account_code = EXCLUDED.gl_account_code,
                gl_account_name = EXCLUDED.gl_account_name,
                voucher_number = EXCLUDED.voucher_number,
                debit_amount = EXCLUDED.debit_amount,
                credit_amount = EXCLUDED.credit_amount,
                net_amount = EXCLUDED.net_amount,
                description = EXCLUDED.description,
                branch_id = EXCLUDED.branch_id,
                posted_by = EXCLUDED.posted_by,
                has_voucher = EXCLUDED.has_voucher,
                _row_hash = EXCLUDED._row_hash,
                _batch_id = EXCLUDED._batch_id,
                is_deleted = EXCLUDED.is_deleted,
                deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.gl_cleaned.deleted_detected_at END,
                cleaned_at = NOW()
            WHERE silver.gl_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash
               OR silver.gl_cleaned.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
        """), params)
        results["gl_cleaned"] = r.rowcount

        # 6. Reference Branches
        r = conn.execute(text("""
            INSERT INTO silver.ref_branches (
                branch_id, branch_code, branch_name, branch_type, region, city, cleaned_at
            )
            SELECT DISTINCT ON (b.branch_id)
                b.branch_id, b.branch_code, b.branch_name, b.branch_type, b.region, b.city, NOW()
            FROM bronze.cb_branches b
            WHERE b.branch_id IS NOT NULL
            ORDER BY b.branch_id, b._loaded_at DESC
            ON CONFLICT (branch_id)
            DO UPDATE SET
                branch_code = EXCLUDED.branch_code,
                branch_name = EXCLUDED.branch_name,
                branch_type = EXCLUDED.branch_type,
                region = EXCLUDED.region,
                city = EXCLUDED.city,
                cleaned_at = NOW();
        """))
        results["ref_branches"] = r.rowcount

    print(f"[Transform:Silver] CBS Bronze → Silver completed: {results}")
    return results


def transform_source_bronze_to_silver_external(
    engine,
    source_id: str,
    source_name: str,
    table_name: str,
    mapping: dict,
    batch_id: str = None
) -> int:
    """
    Set-based SQL merge of an external bronze table into silver.external_source_data.
    Zero Python loop / iterrows(). Pure database engine set-operation.
    """
    safe_sname = sanitize_table_name(source_name)
    safe_tname = sanitize_table_name(table_name)
    target_bronze_table = f"src_{safe_sname}_{safe_tname}"

    pk_field = mapping.get("auditField") or mapping.get("primaryKey")
    target_scope = mapping.get("targetScope", "audit_features")

    # If pk_field not explicitly configured, try to inspect the bronze table primary key / first column
    with engine.connect() as conn:
        cols_query = conn.execute(text("""
            SELECT column_name
            FROM information_schema.columns
            WHERE table_schema = 'bronze' AND table_name = :tname
            ORDER BY ordinal_position;
        """), {"tname": target_bronze_table})
        all_cols = [r[0] for r in cols_query.fetchall()]

    if not all_cols:
        print(f"[Transform:Silver] Bronze table bronze.{target_bronze_table} has no columns or not found.")
        return 0

    user_cols = [c for c in all_cols if not c.startswith("_")]
    if not pk_field or pk_field not in user_cols:
        pk_field = user_cols[0] if user_cols else all_cols[0]

    safe_pk_field = quote_identifier(pk_field)
    batch_filter = "AND b._batch_id = :bid" if batch_id else ""
    params = {
        "sid": source_id,
        "sname": source_name,
        "table": table_name,
        "scope": target_scope,
        "bid": batch_id
    }

    # Set-based SQL upsert
    query = text(f"""
        INSERT INTO silver.external_source_data (
            source_id, source_name, table_name, record_data, dedup_key,
            target_scope, record_hash, batch_id, is_deleted, deleted_detected_at, validated_at
        )
        SELECT
            :sid,
            :sname,
            :table,
            to_jsonb(b) - '_load_id' - '_source' - '_loaded_at' - '_row_hash' - '_op' - '_batch_id',
            COALESCE(b.{safe_pk_field}::TEXT, MD5(ROW(b.*)::TEXT)),
            :scope,
            COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
            b._batch_id,
            CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
            CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
            NOW()
        FROM bronze.{quote_identifier(target_bronze_table)} b
        WHERE b.{safe_pk_field} IS NOT NULL {batch_filter}
        ON CONFLICT (source_id, table_name, dedup_key)
        DO UPDATE SET
            record_data = EXCLUDED.record_data,
            target_scope = EXCLUDED.target_scope,
            record_hash = EXCLUDED.record_hash,
            batch_id = EXCLUDED.batch_id,
            is_deleted = EXCLUDED.is_deleted,
            deleted_detected_at = CASE WHEN EXCLUDED.is_deleted THEN NOW() ELSE silver.external_source_data.deleted_detected_at END,
            validated_at = NOW()
        WHERE silver.external_source_data.record_hash IS DISTINCT FROM EXCLUDED.record_hash
           OR silver.external_source_data.is_deleted IS DISTINCT FROM EXCLUDED.is_deleted;
    """)

    with engine.begin() as conn:
        r = conn.execute(query, params)
        rowcount = r.rowcount

    print(f"[Transform:Silver] External {source_name}.{table_name} → Silver: {rowcount} records upserted.")
    return rowcount
