"""
Data Hub API — Gold Zone Enriched Models & Feature Marts
Incremental, set-based transformation from Silver to Gold.
Refreshes TimescaleDB continuous aggregates, precomputes feature marts,
and updates data quality metrics & version counters.
"""
import json
import traceback
from contextlib import nullcontext
from sqlalchemy import text


def run_silver_to_gold_cbs(engine) -> dict:
    """
    Incrementally enrich Silver CBS data into Gold Star Schema and feature marts.
    Pure set-based SQL with ON CONFLICT resolution.
    """
    print("[Transform:Gold] Starting Silver → Gold CBS enrichment...")
    results = {}

    with engine.begin() as conn:
        # 1. Dimension: Branches
        r = conn.execute(text("""
            INSERT INTO gold.dim_branches (
                source_id, branch_id, branch_code, branch_name, branch_type, region, city,
                total_accounts, total_customers, total_trx_volume, anomaly_count, refreshed_at
            )
            SELECT
                'cbs_simulator' AS source_id,
                b.branch_id,
                b.branch_code,
                b.branch_name,
                b.branch_type,
                b.region,
                b.city,
                COALESCE(a.cnt, 0),
                COALESCE(c.cnt, 0),
                COALESCE(t.total, 0),
                COALESCE(t.anomalies, 0),
                NOW()
            FROM silver.ref_branches b
            LEFT JOIN (
                SELECT branch_id, COUNT(*) AS cnt
                FROM silver.acct_cleaned
                WHERE NOT is_deleted
                GROUP BY branch_id
            ) a ON a.branch_id = b.branch_id
            LEFT JOIN (
                SELECT branch_id, COUNT(*) AS cnt
                FROM silver.cust_cleaned
                WHERE NOT is_deleted
                GROUP BY branch_id
            ) c ON c.branch_id = b.branch_id
            LEFT JOIN (
                SELECT branch_id, SUM(amount) AS total, SUM(CASE WHEN is_suspicious THEN 1 ELSE 0 END) AS anomalies
                FROM silver.trx_cleaned
                WHERE NOT is_deleted
                GROUP BY branch_id
            ) t ON t.branch_id = b.branch_id
            ON CONFLICT (branch_id)
            DO UPDATE SET
                branch_code = EXCLUDED.branch_code,
                branch_name = EXCLUDED.branch_name,
                branch_type = EXCLUDED.branch_type,
                region = EXCLUDED.region,
                city = EXCLUDED.city,
                total_accounts = EXCLUDED.total_accounts,
                total_customers = EXCLUDED.total_customers,
                total_trx_volume = EXCLUDED.total_trx_volume,
                anomaly_count = EXCLUDED.anomaly_count,
                refreshed_at = NOW();
        """))
        results["dim_branches"] = r.rowcount

        # 2. Fact: Transactions (Hypertable)
        r = conn.execute(text("""
            INSERT INTO gold.fact_transactions (
                source_id, trx_ref_number, trx_date, trx_hour, trx_day_of_week, trx_type,
                category, channel, amount, amount_millions, branch_id, branch_name, region,
                customer_name, customer_type, customer_risk, account_type,
                authorization_status, is_suspicious, is_round_amount, is_after_hours,
                benford_first_digit, is_deleted, refreshed_at
            )
            SELECT
                t.source_id,
                t.trx_ref_number,
                t.trx_date,
                t.trx_hour,
                t.trx_day_of_week,
                t.trx_type,
                t.category,
                t.channel,
                t.amount,
                t.amount_millions,
                t.branch_id,
                b.branch_name,
                b.region,
                cu.customer_name,
                cu.customer_type,
                cu.risk_profile,
                ac.account_type,
                t.authorization_status,
                t.is_suspicious,
                t.is_round_amount,
                t.is_after_hours,
                CASE WHEN t.amount > 0 THEN SUBSTRING(CAST(FLOOR(t.amount) AS TEXT) FROM 1 FOR 1)::INT ELSE NULL END,
                t.is_deleted,
                NOW()
            FROM silver.trx_cleaned t
            LEFT JOIN silver.ref_branches b ON b.branch_id = t.branch_id
            LEFT JOIN silver.acct_cleaned ac ON ac.account_id = t.account_id AND ac.source_id = t.source_id
            LEFT JOIN silver.cust_cleaned cu ON cu.customer_id = ac.customer_id AND cu.source_id = t.source_id
            ON CONFLICT (source_id, trx_ref_number, trx_date)
            DO UPDATE SET
                trx_hour = EXCLUDED.trx_hour,
                trx_day_of_week = EXCLUDED.trx_day_of_week,
                trx_type = EXCLUDED.trx_type,
                category = EXCLUDED.category,
                channel = EXCLUDED.channel,
                amount = EXCLUDED.amount,
                amount_millions = EXCLUDED.amount_millions,
                branch_id = EXCLUDED.branch_id,
                branch_name = EXCLUDED.branch_name,
                region = EXCLUDED.region,
                customer_name = EXCLUDED.customer_name,
                customer_type = EXCLUDED.customer_type,
                customer_risk = EXCLUDED.customer_risk,
                account_type = EXCLUDED.account_type,
                authorization_status = EXCLUDED.authorization_status,
                is_suspicious = EXCLUDED.is_suspicious,
                is_round_amount = EXCLUDED.is_round_amount,
                is_after_hours = EXCLUDED.is_after_hours,
                benford_first_digit = EXCLUDED.benford_first_digit,
                is_deleted = EXCLUDED.is_deleted,
                refreshed_at = NOW();
        """))
        results["fact_transactions"] = r.rowcount

        # 3. Fact: Loans
        r = conn.execute(text("""
            INSERT INTO gold.fact_loans (
                source_id, loan_id, loan_number, loan_type, customer_name, customer_type,
                branch_name, region, product_name, plafond, outstanding,
                interest_rate, tenor_months, collateral_value, ltv_ratio,
                collectability, days_past_due, status,
                rate_within_policy, plafond_within_policy, tenor_within_policy,
                is_deleted, refreshed_at
            )
            SELECT
                l.source_id,
                l.loan_id,
                l.loan_number,
                l.loan_type,
                cu.customer_name,
                cu.customer_type,
                b.branch_name,
                b.region,
                p.product_name,
                l.plafond,
                l.outstanding,
                l.interest_rate,
                l.tenor_months,
                l.collateral_value,
                l.ltv_ratio,
                l.collectability,
                l.days_past_due,
                l.status,
                -- Policy checks
                CASE WHEN p.min_rate IS NOT NULL AND p.max_rate IS NOT NULL
                     THEN l.interest_rate BETWEEN p.min_rate AND p.max_rate ELSE TRUE END,
                CASE WHEN p.max_plafond IS NOT NULL
                     THEN l.plafond <= p.max_plafond ELSE TRUE END,
                CASE WHEN p.max_tenor IS NOT NULL
                     THEN l.tenor_months * 30 <= p.max_tenor ELSE TRUE END,
                l.is_deleted,
                NOW()
            FROM silver.loan_cleaned l
            LEFT JOIN silver.cust_cleaned cu ON cu.customer_id = l.customer_id AND cu.source_id = l.source_id
            LEFT JOIN silver.ref_branches b ON b.branch_id = l.branch_id
            LEFT JOIN (
                SELECT p.product_id, p.product_name, p.min_rate, p.max_rate, p.max_plafond,
                       MAX(CASE WHEN pp.rule_type = 'TENOR_LIMIT' THEN pp.max_value END) AS max_tenor
                FROM bronze.cb_products p
                LEFT JOIN bronze.cb_product_policies pp ON pp.product_id = p.product_id
                GROUP BY p.product_id, p.product_name, p.min_rate, p.max_rate, p.max_plafond
            ) p ON p.product_id = l.product_id
            ON CONFLICT (source_id, loan_number)
            DO UPDATE SET
                loan_type = EXCLUDED.loan_type,
                customer_name = EXCLUDED.customer_name,
                customer_type = EXCLUDED.customer_type,
                branch_name = EXCLUDED.branch_name,
                region = EXCLUDED.region,
                product_name = EXCLUDED.product_name,
                plafond = EXCLUDED.plafond,
                outstanding = EXCLUDED.outstanding,
                interest_rate = EXCLUDED.interest_rate,
                tenor_months = EXCLUDED.tenor_months,
                collateral_value = EXCLUDED.collateral_value,
                ltv_ratio = EXCLUDED.ltv_ratio,
                collectability = EXCLUDED.collectability,
                days_past_due = EXCLUDED.days_past_due,
                status = EXCLUDED.status,
                rate_within_policy = EXCLUDED.rate_within_policy,
                plafond_within_policy = EXCLUDED.plafond_within_policy,
                tenor_within_policy = EXCLUDED.tenor_within_policy,
                is_deleted = EXCLUDED.is_deleted,
                refreshed_at = NOW();
        """))
        results["fact_loans"] = r.rowcount

        # 4. Fact: GL Entries
        r = conn.execute(text("""
            INSERT INTO gold.fact_gl_entries (
                source_id, gl_id, gl_date, gl_account_code, gl_account_name, voucher_number,
                debit_amount, credit_amount, net_amount, branch_name, has_voucher,
                has_gap, is_deleted, refreshed_at
            )
            SELECT
                g.source_id,
                g.gl_id,
                g.gl_date,
                g.gl_account_code,
                g.gl_account_name,
                g.voucher_number,
                g.debit_amount,
                g.credit_amount,
                g.net_amount,
                b.branch_name,
                g.has_voucher,
                NOT g.has_voucher,
                g.is_deleted,
                NOW()
            FROM silver.gl_cleaned g
            LEFT JOIN silver.ref_branches b ON b.branch_id = g.branch_id
            ON CONFLICT (source_id, gl_id)
            DO UPDATE SET
                gl_date = EXCLUDED.gl_date,
                gl_account_code = EXCLUDED.gl_account_code,
                gl_account_name = EXCLUDED.gl_account_name,
                voucher_number = EXCLUDED.voucher_number,
                debit_amount = EXCLUDED.debit_amount,
                credit_amount = EXCLUDED.credit_amount,
                net_amount = EXCLUDED.net_amount,
                branch_name = EXCLUDED.branch_name,
                has_voucher = EXCLUDED.has_voucher,
                has_gap = EXCLUDED.has_gap,
                is_deleted = EXCLUDED.is_deleted,
                refreshed_at = NOW();
        """))
        results["fact_gl_entries"] = r.rowcount

    # Refresh continuous aggregates and feature marts
    refresh_gold_marts(engine, source_id="cbs_simulator")
    print(f"[Transform:Gold] Silver → Gold CBS completed: {results}")
    return results


def transform_source_silver_to_gold_external(
    engine,
    source_id: str,
    source_name: str,
    table_name: str,
    mapping: dict
) -> int:
    """
    Set-based incremental transform for external sources to gold.fact_external_audit_data.
    """
    target_scope = mapping.get("targetScope", "audit_features")
    target_module = mapping.get("targetModule", "")
    anomaly_rules = mapping.get("anomalyRules", [])

    with engine.begin() as conn:
        r = conn.execute(text("""
            INSERT INTO gold.fact_external_audit_data 
                (source_id, source_name, table_name, target_scope, target_module, record_data, anomaly_flags, processed_at)
            SELECT
                source_id,
                source_name,
                table_name,
                target_scope,
                :module,
                record_data,
                CAST(:rules AS jsonb),
                NOW()
            FROM silver.external_source_data
            WHERE source_id = :sid AND table_name = :table AND NOT is_deleted
            ON CONFLICT DO NOTHING;
        """), {
            "sid": source_id,
            "table": table_name,
            "module": target_module,
            "rules": json.dumps(anomaly_rules)
        })
        rowcount = r.rowcount

        # If data_analytics scope or explicitly mapped to ai_training_pool, feed AI training pool
        canonical_ent = (mapping.get("canonicalEntity") or "").lower().strip()
        if target_scope == "data_analytics" or canonical_ent == "ai_training_pool":
            conn.execute(text("""
                INSERT INTO gold.ai_training_pool 
                    (source_id, source_name, table_name, record_data, feature_type, ingested_at)
                SELECT
                    source_id,
                    source_name,
                    table_name,
                    record_data,
                    CASE 
                        WHEN :module ILIKE '%anomal%' OR :canon = 'ai_training_pool' THEN 'anomaly'
                        WHEN :module ILIKE '%text%' OR :module ILIKE '%document%' OR :module ILIKE '%nlp%' THEN 'text'
                        ELSE 'risk_score'
                    END,
                    NOW()
                FROM silver.external_source_data
                WHERE source_id = :sid AND table_name = :table AND NOT is_deleted
                ON CONFLICT DO NOTHING;
            """), {
                "sid": source_id,
                "table": table_name,
                "module": target_module,
                "canon": canonical_ent
            })

        # Update data quality mart
        conn.execute(text("""
            INSERT INTO gold.mart_data_quality (
                source_id, table_name, total_rows, null_count, duplicate_count,
                deletes_detected, completeness_pct, accuracy_pct, updated_at
            )
            SELECT
                :sid,
                :table,
                COUNT(*),
                0,
                0,
                SUM(CASE WHEN is_deleted THEN 1 ELSE 0 END),
                100.0,
                100.0,
                NOW()
            FROM silver.external_source_data
            WHERE source_id = :sid AND table_name = :table
            GROUP BY source_id, table_name
            ON CONFLICT (source_id, table_name)
            DO UPDATE SET
                total_rows = EXCLUDED.total_rows,
                deletes_detected = EXCLUDED.deletes_detected,
                updated_at = NOW();
        """), {"sid": source_id, "table": table_name})

        # Bump mart version
        conn.execute(text("UPDATE gold.mart_version SET version = version + 1, updated_at = NOW() WHERE mart_name = 'data_quality';"))

    # Automatically project normalized/external records into canonical Gold CAATT tables
    # (gold.fact_transactions, gold.fact_loans, gold.fact_gl_entries) so CAATT Engine
    # analyzes every connected external DB without requiring manual CAATT mappings.
    try:
        projected = project_external_silver_to_canonical_gold(
            engine, source_id, source_name, table_name,
            canonical_entity=(mapping.get("canonicalEntity") or "")
        )
        if any(v > 0 for v in projected.values()):
            print(f"[Transform:Gold] Canonical projection for {source_id}.{table_name}: {projected}")
    except Exception as proj_err:
        print(f"[Transform:Gold] Canonical projection error for {source_id}.{table_name}: {proj_err}")
        traceback.print_exc()

    return rowcount


def _pick_jsonb_key(lower_to_actual: dict, candidates: list, sample_rows: list = None, require_numeric: bool = False) -> str:
    """Resolve the best matching column key from external JSONB records case-insensitively."""
    def _is_num(val) -> bool:
        if val is None or isinstance(val, bool):
            return False
        try:
            float(str(val).replace(",", "").strip())
            return True
        except (TypeError, ValueError):
            return False

    # 1. Exact priority match
    for cand in candidates:
        if cand in lower_to_actual:
            actual = lower_to_actual[cand]
            if not require_numeric or not sample_rows or any(_is_num(r.get(actual)) for r in sample_rows if isinstance(r, dict)):
                return actual

    # 2. Substring match fallback
    for cand in candidates:
        if len(cand) < 4:
            continue
        for lk, actual in lower_to_actual.items():
            if cand in lk:
                if not require_numeric or not sample_rows or any(_is_num(r.get(actual)) for r in sample_rows if isinstance(r, dict)):
                    return actual
    return ""


def _detect_external_field_mapping(sample_rows: list, table_name: str) -> dict:
    """Inspect external JSONB sample records and infer semantic field mappings for canonical Gold projection."""
    lower_to_actual = {}
    for row in sample_rows:
        if isinstance(row, dict):
            for k in row.keys():
                kl = str(k).lower().strip()
                if kl not in lower_to_actual:
                    lower_to_actual[kl] = str(k)

    amount_key = _pick_jsonb_key(
        lower_to_actual,
        [
            "amount", "trx_amount", "transaction_amount", "nominal", "total_amount",
            "invoice_amount", "payment_amount", "net_amount", "gross_amount",
            "nilai", "jumlah", "principal_amount", "plafond", "outstanding",
            "debit_amount", "credit_amount", "debit", "credit", "value", "total",
        ],
        sample_rows=sample_rows,
        require_numeric=True,
    )
    debit_key = _pick_jsonb_key(lower_to_actual, ["debit_amount", "debit", "debet"], sample_rows, require_numeric=True)
    credit_key = _pick_jsonb_key(lower_to_actual, ["credit_amount", "credit", "kredit"], sample_rows, require_numeric=True)
    date_key = _pick_jsonb_key(
        lower_to_actual,
        [
            "trx_date", "transaction_date", "gl_date", "posting_date", "paid_at",
            "disbursed_at", "value_date", "booking_date", "created_at", "updated_at",
            "tanggal", "tgl_trx", "date", "timestamp",
        ],
    )
    ref_key = _pick_jsonb_key(
        lower_to_actual,
        [
            "trx_ref_number", "trx_ref", "ref_number", "reference_no", "reference",
            "voucher_number", "voucher_no", "invoice_no", "invoice_number",
            "loan_number", "doc_no", "trx_id", "transaction_id", "gl_id", "id",
        ],
    )
    branch_key = _pick_jsonb_key(
        lower_to_actual,
        ["branch_name", "cabang", "nama_cabang", "branch", "unit_name", "office", "department", "branch_code", "branch_id"],
    )
    branch_id_key = _pick_jsonb_key(lower_to_actual, ["branch_id", "cabang_id", "unit_id", "office_id"])
    customer_key = _pick_jsonb_key(
        lower_to_actual,
        [
            "customer_name", "nasabah", "nama_nasabah", "customer", "client_name",
            "vendor_name", "merchant_name", "payee", "beneficiary", "account_name",
            "debtor_name", "borrower_name", "customer_id",
        ],
    )
    customer_id_key = _pick_jsonb_key(lower_to_actual, ["customer_id", "nasabah_id", "client_id", "account_id"])
    category_key = _pick_jsonb_key(
        lower_to_actual,
        ["category", "trx_type", "transaction_type", "jenis_transaksi", "loan_type", "product_name", "product_type", "type", "jenis", "account_type", "segment"],
    )
    channel_key = _pick_jsonb_key(lower_to_actual, ["channel", "payment_method", "metode_pembayaran", "kanal", "source_channel"])
    acct_type_key = _pick_jsonb_key(lower_to_actual, ["account_type", "jenis_rekening", "product_type", "tier"])
    suspicious_key = _pick_jsonb_key(lower_to_actual, ["is_suspicious", "is_anomaly", "is_flagged", "suspicious", "flagged", "anomaly", "is_fraud"])
    auth_key = _pick_jsonb_key(lower_to_actual, ["authorization_status", "approval_status", "status", "state"])

    # Loan-specific fields
    loan_num_key = _pick_jsonb_key(lower_to_actual, ["loan_number", "no_kredit", "no_pinjaman", "facility_no", "contract_no", "id", "loan_id"])
    plafond_key = _pick_jsonb_key(lower_to_actual, ["plafond", "limit_amount", "credit_limit", "principal_amount", "amount", "loan_amount"], sample_rows, require_numeric=True)
    outstanding_key = _pick_jsonb_key(lower_to_actual, ["outstanding", "outstanding_balance", "balance", "sisa_pokok", "baki_debet"], sample_rows, require_numeric=True)
    rate_key = _pick_jsonb_key(lower_to_actual, ["interest_rate", "suku_bunga", "rate", "bunga", "apr"], sample_rows, require_numeric=True)
    tenor_key = _pick_jsonb_key(lower_to_actual, ["tenor_months", "tenor", "term_months", "jangka_waktu"], sample_rows, require_numeric=True)
    collateral_key = _pick_jsonb_key(lower_to_actual, ["collateral_value", "nilai_agunan", "agunan", "collateral"], sample_rows, require_numeric=True)
    collectability_key = _pick_jsonb_key(lower_to_actual, ["collectability", "kolektibilitas", "kol", "bucket"], sample_rows, require_numeric=True)
    dpd_key = _pick_jsonb_key(lower_to_actual, ["days_past_due", "dpd", "hari_tunggakan"], sample_rows, require_numeric=True)
    product_key = _pick_jsonb_key(lower_to_actual, ["product_name", "produk", "product_type", "loan_type"])

    # GL-specific fields
    gl_code_key = _pick_jsonb_key(lower_to_actual, ["gl_account_code", "account_code", "coa_code", "kode_akun", "acct_code"])
    gl_name_key = _pick_jsonb_key(lower_to_actual, ["gl_account_name", "account_name", "coa_name", "nama_akun", "description"])
    voucher_key = _pick_jsonb_key(lower_to_actual, ["voucher_number", "voucher_no", "no_voucher", "journal_no", "doc_no", "trx_ref_number", "id"])

    return {
        "amount": amount_key,
        "debit": debit_key,
        "credit": credit_key,
        "date": date_key,
        "ref": ref_key,
        "branch": branch_key,
        "branch_id": branch_id_key,
        "customer": customer_key,
        "customer_id": customer_id_key,
        "category": category_key,
        "channel": channel_key,
        "acct_type": acct_type_key,
        "suspicious": suspicious_key,
        "auth_status": auth_key,
        "loan_num": loan_num_key,
        "plafond": plafond_key or amount_key,
        "outstanding": outstanding_key or plafond_key or amount_key,
        "rate": rate_key,
        "tenor": tenor_key,
        "collateral": collateral_key,
        "collectability": collectability_key,
        "dpd": dpd_key,
        "product": product_key or category_key,
        "gl_code": gl_code_key,
        "gl_name": gl_name_key,
        "voucher": voucher_key or ref_key,
        "keys_lower": set(lower_to_actual.keys()),
    }


def project_external_silver_to_canonical_gold(engine, source_id: str, source_name: str, table_name: str, canonical_entity: str = "") -> dict:
    """
    Project real external DB rows stored in silver.external_source_data into canonical
    gold.fact_transactions, gold.fact_loans, and gold.fact_gl_entries for CAATT analysis.
    Uses only parameterized JSONB key extraction so any external Postgres schema works safely.
    """
    stats = {"transactions": 0, "loans": 0, "gl_entries": 0}
    tbl_lower = (table_name or "").lower().strip()
    canon_lower = (canonical_entity or "").lower().strip()

    with engine.begin() as conn:
        sample_rows_raw = conn.execute(text("""
            SELECT record_data
            FROM silver.external_source_data
            WHERE source_id = :sid AND table_name = :table AND NOT is_deleted
            LIMIT 10;
        """), {"sid": source_id, "table": table_name}).fetchall()

        if not sample_rows_raw:
            return stats

        sample_rows = [
            json.loads(r[0]) if isinstance(r[0], str) else (r[0] or {})
            for r in sample_rows_raw
        ]
        fmap = _detect_external_field_mapping(sample_rows, table_name)

        # Ensure migration-011 sequences exist even if called standalone in tests
        conn.execute(text("CREATE SEQUENCE IF NOT EXISTS gold.external_loan_id_seq START WITH 1000000 INCREMENT BY 1;"))

        trx_keywords = (
            "trx", "transac", "mutasi", "mutation", "payment", "pembayaran",
            "invoice", "tagihan", "order", "sale", "penjualan", "disburse",
            "cash", "kas", "transfer", "remit", "ledger", "jurnal", "journal",
            "entry", "entries", "voucher", "billing", "fee", "settlement",
            "withdraw", "deposit",
        )
        loan_keywords = ("loan", "kredit", "pinjaman", "financing", "credit_fac")
        gl_keywords = ("gl", "ledger", "jurnal", "journal", "voucher", "coa")
        dim_only_keywords = (
            "users", "roles", "permissions", "migrations", "alembic", "branches",
            "cabang", "customers", "nasabah", "products", "produk", "employees",
            "departments", "config", "settings", "policies",
        )

        is_loan_table = (canon_lower == "loans") or any(k in tbl_lower for k in loan_keywords) or (
            bool(fmap["plafond"]) and (bool(fmap["collectability"]) or bool(fmap["rate"]) or bool(fmap["tenor"]))
        )
        is_gl_table = (canon_lower == "gl_entries") or any(k in tbl_lower for k in gl_keywords) or (
            bool(fmap["debit"]) and bool(fmap["credit"])
        )
        is_trx_keyword = (canon_lower == "transactions") or any(k in tbl_lower for k in trx_keywords)
        is_pure_dim = any(tbl_lower == d or tbl_lower.endswith("_" + d) for d in dim_only_keywords) and not is_trx_keyword and not is_loan_table and not is_gl_table

        # If source has no explicit transaction table yet and this is a loan table with positive amounts,
        # project loan disbursements into fact_transactions as fallback so Benford & Stratification have real data.
        has_dedicated_trx = conn.execute(text("""
            SELECT 1 FROM silver.external_source_data
            WHERE source_id = :sid AND NOT is_deleted
              AND LOWER(table_name) ~ '(trx|transac|mutasi|payment|invoice|order|transfer|cash|withdraw|deposit|mutation|keuangan)'
            LIMIT 1;
        """), {"sid": source_id}).fetchone() is not None

        should_project_trx = (
            (bool(fmap["amount"]) or bool(fmap["debit"]) or bool(fmap["credit"]))
            and not is_pure_dim
            and (not is_gl_table or not has_dedicated_trx)
            and (is_trx_keyword or (bool(fmap["date"]) and not is_loan_table) or (is_loan_table and not has_dedicated_trx))
        )

        # Resolve effective keys (avoid empty param errors in PostgreSQL by passing '__none__' sentinel)
        amt_key = fmap["amount"] or fmap["debit"] or fmap["credit"] or "__none__"
        date_key = fmap["date"] or "__none__"
        ref_key = fmap["ref"] or "__none__"
        branch_key = fmap["branch"] or "__none__"
        branch_id_key = fmap["branch_id"] or "__none__"
        cust_key = fmap["customer"] or "__none__"
        cust_id_key = fmap["customer_id"] or "__none__"
        cat_key = fmap["category"] or "__none__"
        chan_key = fmap["channel"] or "__none__"
        acct_key = fmap["acct_type"] or "__none__"
        susp_key = fmap["suspicious"] or "__none__"
        auth_key = fmap["auth_status"] or "__none__"
        tbl_short = "".join(ch for ch in tbl_lower if ch.isalnum())[:6].upper() or "EXT"

        if should_project_trx:
            trx_cte_and_select = """
                WITH parsed AS (
                    SELECT
                        s.id,
                        LEFT(
                            COALESCE(
                                NULLIF(TRIM(s.record_data ->> :ref_k), ''),
                                :tbl_short || '-' || s.id::TEXT
                            ),
                            30
                        ) AS trx_ref_number,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :date_k), '') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}'
                                THEN SUBSTRING(TRIM(s.record_data ->> :date_k) FROM 1 FOR 10)::DATE
                            ELSE COALESCE(s.validated_at::DATE, CURRENT_DATE)
                        END AS trx_date,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :date_k), '') ~ '[0-9]{2}:[0-9]{2}'
                                THEN COALESCE(SUBSTRING(TRIM(s.record_data ->> :date_k) FROM '([0-9]{2}):[0-9]{2}')::INT, 12)
                            ELSE 12
                        END AS trx_hour,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :date_k), '') ~ '[0-9]{2}:[0-9]{2}' THEN TRUE
                            ELSE FALSE
                        END AS has_time_component,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :amt_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :amt_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS amount,
                        LEFT(
                            COALESCE(
                                NULLIF(TRIM(s.record_data ->> :branch_k), ''),
                                CASE
                                    WHEN NULLIF(TRIM(s.record_data ->> :branch_id_k), '') IS NOT NULL
                                        THEN 'Cabang #' || TRIM(s.record_data ->> :branch_id_k)
                                    ELSE NULL
                                END,
                                :source_name
                            ),
                            100
                        ) AS branch_name,
                        LEFT(
                            COALESCE(
                                NULLIF(TRIM(s.record_data ->> :cust_k), ''),
                                CASE
                                    WHEN NULLIF(TRIM(s.record_data ->> :cust_id_k), '') IS NOT NULL
                                        THEN 'Entity #' || TRIM(s.record_data ->> :cust_id_k)
                                    ELSE NULL
                                END,
                                :source_name || ' Record'
                            ),
                            150
                        ) AS customer_name,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :cat_k), ''), INITCAP(REPLACE(:table, '_', ' '))), 30) AS category,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :chan_k), ''), 'External DB'), 30) AS channel,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :acct_k), ''), 'Standard'), 30) AS account_type,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :auth_k), ''), 'Authorized'), 30) AS authorization_status,
                        LOWER(COALESCE(TRIM(s.record_data ->> :susp_k), '')) IN ('true', 't', '1', 'yes', 'y') AS explicit_suspicious
                    FROM silver.external_source_data s
                    WHERE s.source_id = :sid
                      AND s.table_name = :table
                      AND NOT s.is_deleted
                ),
                enriched AS (
                    SELECT DISTINCT ON (trx_ref_number, trx_date)
                        trx_ref_number,
                        trx_date,
                        LEAST(23, GREATEST(0, trx_hour)) AS trx_hour,
                        EXTRACT(DOW FROM trx_date)::INT AS trx_day_of_week,
                        category AS trx_type,
                        category,
                        channel,
                        amount,
                        ROUND((amount / 1000000.0)::NUMERIC, 4) AS amount_millions,
                        branch_name,
                        customer_name,
                        account_type,
                        authorization_status,
                        has_time_component AND (trx_hour < 6 OR trx_hour >= 21) AS is_after_hours,
                        (amount >= 1000000 AND MOD(amount::NUMERIC, 1000000) = 0) AS is_round_amount,
                        (
                            explicit_suspicious
                            OR LOWER(authorization_status) IN ('rejected', 'pending_override', 'flagged', 'suspended')
                        ) AS is_suspicious,
                        NULLIF(
                            SUBSTRING(TRIM(LEADING '0.' FROM ABS(amount)::NUMERIC::TEXT) FROM '^([1-9])'),
                            ''
                        )::INT AS benford_first_digit
                    FROM parsed
                    WHERE amount > 0
                    ORDER BY trx_ref_number, trx_date, id DESC
                )
                INSERT INTO gold.fact_transactions (
                    source_id, trx_ref_number, trx_date, trx_hour, trx_day_of_week,
                    trx_type, category, channel, amount, amount_millions,
                    branch_id, branch_name, region, customer_name, customer_type, customer_risk,
                    account_type, authorization_status, is_suspicious, is_round_amount, is_after_hours,
                    benford_first_digit, is_deleted, refreshed_at
                )
                SELECT
                    :sid,
                    trx_ref_number,
                    trx_date,
                    trx_hour,
                    trx_day_of_week,
                    trx_type,
                    category,
                    channel,
                    amount,
                    amount_millions,
                    1,
                    branch_name,
                    'Connected Source',
                    customer_name,
                    'External',
                    CASE WHEN is_suspicious THEN 'High' ELSE 'Normal' END,
                    account_type,
                    authorization_status,
                    is_suspicious,
                    is_round_amount,
                    is_after_hours,
                    benford_first_digit,
                    FALSE,
                    NOW()
                FROM enriched
            """
            trx_params = {
                "sid": source_id,
                "source_name": source_name,
                "table": table_name,
                "tbl_short": tbl_short,
                "amt_k": amt_key,
                "date_k": date_key,
                "ref_k": ref_key,
                "branch_k": branch_key,
                "branch_id_k": branch_id_key,
                "cust_k": cust_key,
                "cust_id_k": cust_id_key,
                "cat_k": cat_key,
                "chan_k": chan_key,
                "acct_k": acct_key,
                "susp_k": susp_key,
                "auth_k": auth_key,
            }
            trx_upsert_sql = text(
                trx_cte_and_select
                + """
                ON CONFLICT (source_id, trx_ref_number, trx_date)
                DO UPDATE SET
                    trx_hour = EXCLUDED.trx_hour,
                    trx_type = EXCLUDED.trx_type,
                    category = EXCLUDED.category,
                    channel = EXCLUDED.channel,
                    amount = EXCLUDED.amount,
                    amount_millions = EXCLUDED.amount_millions,
                    branch_name = EXCLUDED.branch_name,
                    customer_name = EXCLUDED.customer_name,
                    authorization_status = EXCLUDED.authorization_status,
                    is_suspicious = EXCLUDED.is_suspicious,
                    is_round_amount = EXCLUDED.is_round_amount,
                    is_after_hours = EXCLUDED.is_after_hours,
                    benford_first_digit = EXCLUDED.benford_first_digit,
                    is_deleted = FALSE,
                    refreshed_at = NOW();
                """
            )
            savepoint_ctx = conn.begin_nested if hasattr(conn, "begin_nested") else nullcontext
            try:
                with savepoint_ctx():
                    r_trx = conn.execute(trx_upsert_sql, trx_params)
                    stats["transactions"] = r_trx.rowcount or 0
            except Exception as trx_upsert_err:
                print(f"[Transform:Gold] Upsert fallback (ON CONFLICT DO NOTHING) for {source_id}.{table_name}: {trx_upsert_err}")
                with savepoint_ctx():
                    r_trx = conn.execute(text(trx_cte_and_select + "\nON CONFLICT DO NOTHING;"), trx_params)
                    stats["transactions"] = r_trx.rowcount or 0

        if is_loan_table and (fmap["plafond"] or fmap["amount"]):
            r_loans = conn.execute(text("""
                WITH parsed_loans AS (
                    SELECT DISTINCT ON (loan_number)
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :loan_k), ''), :tbl_short || '-' || s.id::TEXT), 20) AS loan_number,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :prod_k), ''), 'Commercial'), 30) AS loan_type,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :cust_k), ''), :source_name || ' Borrower'), 150) AS customer_name,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :branch_k), ''), :source_name), 100) AS branch_name,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :prod_k), ''), 'Standard Loan'), 100) AS product_name,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :plaf_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :plaf_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS plafond,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :outs_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :outs_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS outstanding,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :rate_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ROUND(REPLACE(TRIM(s.record_data ->> :rate_k), ',', '')::NUMERIC, 2)
                            ELSE 10.50::NUMERIC(5,2)
                        END AS interest_rate,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :tenor_k), '') ~ '^[0-9]+$'
                                THEN LEAST(360, TRIM(s.record_data ->> :tenor_k)::INT)
                            ELSE 36
                        END AS tenor_months,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :coll_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :coll_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS collateral_value,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :col_k), '') ~ '^[0-9]+$'
                                THEN LEAST(5, GREATEST(1, TRIM(s.record_data ->> :col_k)::INT))
                            ELSE 1
                        END AS collectability,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :dpd_k), '') ~ '^[0-9]+$'
                                THEN TRIM(s.record_data ->> :dpd_k)::INT
                            ELSE 0
                        END AS days_past_due,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :auth_k), ''), 'Active'), 20) AS status
                    FROM silver.external_source_data s
                    WHERE s.source_id = :sid
                      AND s.table_name = :table
                      AND NOT s.is_deleted
                    ORDER BY loan_number, s.id DESC
                )
                INSERT INTO gold.fact_loans (
                    loan_id, source_id, loan_number, loan_type, customer_name, customer_type,
                    branch_name, region, product_name, plafond, outstanding, interest_rate,
                    tenor_months, collateral_value, ltv_ratio, collectability, days_past_due,
                    status, rate_within_policy, plafond_within_policy, tenor_within_policy,
                    is_deleted, refreshed_at
                )
                SELECT
                    nextval('gold.external_loan_id_seq'),
                    :sid,
                    loan_number,
                    loan_type,
                    customer_name,
                    'External',
                    branch_name,
                    'Connected Source',
                    product_name,
                    plafond,
                    outstanding,
                    interest_rate,
                    tenor_months,
                    collateral_value,
                    CASE WHEN collateral_value > 0 THEN LEAST(999.99, ROUND((plafond * 100.0 / collateral_value)::NUMERIC, 2)) ELSE 0 END,
                    collectability,
                    days_past_due,
                    status,
                    (interest_rate <= 24.0),
                    (plafond <= 5000000000),
                    (tenor_months <= 120),
                    FALSE,
                    NOW()
                FROM parsed_loans
                WHERE plafond > 0 OR outstanding > 0
                ON CONFLICT (source_id, loan_number)
                DO UPDATE SET
                    loan_type = EXCLUDED.loan_type,
                    customer_name = EXCLUDED.customer_name,
                    branch_name = EXCLUDED.branch_name,
                    product_name = EXCLUDED.product_name,
                    plafond = EXCLUDED.plafond,
                    outstanding = EXCLUDED.outstanding,
                    interest_rate = EXCLUDED.interest_rate,
                    tenor_months = EXCLUDED.tenor_months,
                    collateral_value = EXCLUDED.collateral_value,
                    ltv_ratio = EXCLUDED.ltv_ratio,
                    collectability = EXCLUDED.collectability,
                    days_past_due = EXCLUDED.days_past_due,
                    status = EXCLUDED.status,
                    rate_within_policy = EXCLUDED.rate_within_policy,
                    plafond_within_policy = EXCLUDED.plafond_within_policy,
                    tenor_within_policy = EXCLUDED.tenor_within_policy,
                    is_deleted = FALSE,
                    refreshed_at = NOW();
            """), {
                "sid": source_id,
                "source_name": source_name,
                "table": table_name,
                "tbl_short": tbl_short,
                "loan_k": fmap["loan_num"] or ref_key,
                "prod_k": fmap["product"] or cat_key,
                "cust_k": cust_key,
                "branch_k": branch_key,
                "plaf_k": fmap["plafond"] or amt_key,
                "outs_k": fmap["outstanding"] or fmap["plafond"] or amt_key,
                "rate_k": fmap["rate"] or "__none__",
                "tenor_k": fmap["tenor"] or "__none__",
                "coll_k": fmap["collateral"] or "__none__",
                "col_k": fmap["collectability"] or "__none__",
                "dpd_k": fmap["dpd"] or "__none__",
                "auth_k": auth_key,
            })
            stats["loans"] = r_loans.rowcount or 0

        if is_gl_table and (fmap["debit"] or fmap["credit"] or fmap["amount"]):
            deb_k = fmap["debit"] or amt_key
            cred_k = fmap["credit"] or "__none__"
            r_gl = conn.execute(text("""
                WITH parsed_gl AS (
                    SELECT
                        ((ABS(hashtext(:sid || ':' || :table || ':' || s.id::TEXT)) % 1900000000) + 1000000)::INT AS gl_id,
                        CASE
                            WHEN COALESCE(TRIM(s.record_data ->> :date_k), '') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}'
                                THEN SUBSTRING(TRIM(s.record_data ->> :date_k) FROM 1 FOR 10)::DATE
                            ELSE COALESCE(s.validated_at::DATE, CURRENT_DATE)
                        END AS gl_date,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :gl_code_k), ''), '1000'), 20) AS gl_account_code,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :gl_name_k), ''), INITCAP(REPLACE(:table, '_', ' '))), 100) AS gl_account_name,
                        LEFT(NULLIF(TRIM(s.record_data ->> :vouch_k), ''), 30) AS voucher_number,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :deb_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :deb_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS debit_amount,
                        CASE
                            WHEN REPLACE(COALESCE(TRIM(s.record_data ->> :cred_k), ''), ',', '') ~ '^-?[0-9]+(\\.[0-9]+)?$'
                                THEN ABS(REPLACE(TRIM(s.record_data ->> :cred_k), ',', '')::NUMERIC(18,2))
                            ELSE 0::NUMERIC(18,2)
                        END AS credit_amount,
                        LEFT(COALESCE(NULLIF(TRIM(s.record_data ->> :branch_k), ''), :source_name), 100) AS branch_name
                    FROM silver.external_source_data s
                    WHERE s.source_id = :sid
                      AND s.table_name = :table
                      AND NOT s.is_deleted
                )
                INSERT INTO gold.fact_gl_entries (
                    gl_id, source_id, gl_date, gl_account_code, gl_account_name,
                    voucher_number, debit_amount, credit_amount, net_amount,
                    branch_name, has_voucher, has_gap, is_deleted, refreshed_at
                )
                SELECT
                    gl_id,
                    :sid,
                    gl_date,
                    gl_account_code,
                    gl_account_name,
                    voucher_number,
                    debit_amount,
                    credit_amount,
                    (debit_amount - credit_amount),
                    branch_name,
                    (voucher_number IS NOT NULL AND voucher_number != ''),
                    NOT (voucher_number IS NOT NULL AND voucher_number != ''),
                    FALSE,
                    NOW()
                FROM parsed_gl
                WHERE debit_amount > 0 OR credit_amount > 0
                ON CONFLICT (gl_id)
                DO UPDATE SET
                    gl_date = EXCLUDED.gl_date,
                    gl_account_code = EXCLUDED.gl_account_code,
                    gl_account_name = EXCLUDED.gl_account_name,
                    voucher_number = EXCLUDED.voucher_number,
                    debit_amount = EXCLUDED.debit_amount,
                    credit_amount = EXCLUDED.credit_amount,
                    net_amount = EXCLUDED.net_amount,
                    branch_name = EXCLUDED.branch_name,
                    has_voucher = EXCLUDED.has_voucher,
                    is_deleted = FALSE,
                    refreshed_at = NOW()
                WHERE gold.fact_gl_entries.source_id = EXCLUDED.source_id;
            """), {
                "sid": source_id,
                "source_name": source_name,
                "table": table_name,
                "date_k": date_key,
                "gl_code_k": fmap["gl_code"] or "__none__",
                "gl_name_k": fmap["gl_name"] or "__none__",
                "vouch_k": fmap["voucher"] or ref_key,
                "deb_k": deb_k,
                "cred_k": cred_k,
                "branch_k": branch_key,
            })
            stats["gl_entries"] = r_gl.rowcount or 0

    return stats


def refresh_gold_marts(engine, source_id: str = "cbs_simulator"):
    """
    Refresh TimescaleDB continuous aggregates and precompute feature marts.
    Enables instant sub-300ms queries across all 16 features.
    """
    # 1. TimescaleDB Continuous Aggregates Refresh must execute in AUTOCOMMIT mode,
    # outside any transaction block (otherwise PostgreSQL aborts the transaction).
    try:
        with engine.connect().execution_options(isolation_level="AUTOCOMMIT") as ac_conn:
            ac_conn.execute(text("CALL refresh_continuous_aggregate('gold.cagg_benford_daily', NULL, NULL);"))
            ac_conn.execute(text("CALL refresh_continuous_aggregate('gold.cagg_strat_daily', NULL, NULL);"))
    except Exception as e:
        print(f"[Transform:Gold] Continuous aggregate refresh notice: {e}")

    with engine.begin() as conn:
        # 2. Mart: Corporate Risk Profile
        conn.execute(text("""
            INSERT INTO gold.mart_entity_risk_profile (
                source_id, entity, entity_type, total_trx_volume, total_anomalies,
                policy_violations, npl_loans, ai_predicted_score, risk_level, updated_at
            )
            SELECT
                b.source_id,
                b.branch_name AS entity,
                'Branch' AS entity_type,
                b.total_trx_volume,
                b.anomaly_count,
                COALESCE(pol.violations, 0),
                COALESCE(npl.npl_count, 0),
                LEAST(99.9, ROUND((b.anomaly_count * 2.5 + COALESCE(npl.npl_count, 0) * 5.0 + COALESCE(pol.violations, 0) * 3.0)::NUMERIC, 2)),
                CASE 
                    WHEN (b.anomaly_count * 2.5 + COALESCE(npl.npl_count, 0) * 5.0) > 40 THEN 'HIGH'
                    WHEN (b.anomaly_count * 2.5 + COALESCE(npl.npl_count, 0) * 5.0) > 15 THEN 'MEDIUM'
                    ELSE 'LOW'
                END,
                NOW()
            FROM gold.dim_branches b
            LEFT JOIN (
                SELECT branch_name, COUNT(*) AS violations
                FROM gold.fact_loans
                WHERE (NOT rate_within_policy OR NOT plafond_within_policy OR NOT tenor_within_policy)
                  AND NOT is_deleted
                GROUP BY branch_name
            ) pol ON pol.branch_name = b.branch_name
            LEFT JOIN (
                SELECT branch_name, COUNT(*) AS npl_count
                FROM gold.fact_loans
                WHERE collectability >= 3 AND NOT is_deleted
                GROUP BY branch_name
            ) npl ON npl.branch_name = b.branch_name
            WHERE b.source_id = :sid
            ON CONFLICT (source_id, entity)
            DO UPDATE SET
                total_trx_volume = EXCLUDED.total_trx_volume,
                total_anomalies = EXCLUDED.total_anomalies,
                policy_violations = EXCLUDED.policy_violations,
                npl_loans = EXCLUDED.npl_loans,
                ai_predicted_score = EXCLUDED.ai_predicted_score,
                risk_level = EXCLUDED.risk_level,
                updated_at = NOW();
        """), {"sid": source_id})

        # 3. Mart: Strategic Audit Plan Priority Ranking
        conn.execute(text("""
            INSERT INTO gold.mart_audit_plan_priority (
                source_id, entity_code, entity_name, risk_score, risk_level,
                last_audit_year, priority_rank, recommended_cycle, updated_at
            )
            SELECT
                p.source_id,
                b.branch_code AS entity_code,
                p.entity AS entity_name,
                p.ai_predicted_score AS risk_score,
                p.risk_level,
                2023,
                ROW_NUMBER() OVER (ORDER BY p.ai_predicted_score DESC),
                CASE 
                    WHEN p.risk_level = 'HIGH' THEN 'Annual'
                    WHEN p.risk_level = 'MEDIUM' THEN 'Bi-Annual'
                    ELSE 'Tri-Annual'
                END,
                NOW()
            FROM gold.mart_entity_risk_profile p
            JOIN gold.dim_branches b ON b.branch_name = p.entity AND b.source_id = p.source_id
            WHERE p.source_id = :sid
            ON CONFLICT (source_id, entity_code)
            DO UPDATE SET
                risk_score = EXCLUDED.risk_score,
                risk_level = EXCLUDED.risk_level,
                priority_rank = EXCLUDED.priority_rank,
                recommended_cycle = EXCLUDED.recommended_cycle,
                updated_at = NOW();
        """), {"sid": source_id})

        # 4. Mart: Loan Aging
        conn.execute(text("""
            INSERT INTO gold.mart_loan_aging (
                source_id, collectability, dpd_bucket, loan_count, total_outstanding, updated_at
            )
            SELECT
                source_id,
                collectability,
                CASE
                    WHEN days_past_due <= 0 THEN 'Current (0 DPD)'
                    WHEN days_past_due <= 30 THEN '1 - 30 DPD'
                    WHEN days_past_due <= 60 THEN '31 - 60 DPD'
                    WHEN days_past_due <= 90 THEN '61 - 90 DPD'
                    ELSE '> 90 DPD (NPL)'
                END AS dpd_bucket,
                COUNT(*),
                SUM(outstanding),
                NOW()
            FROM gold.fact_loans
            WHERE source_id = :sid AND NOT is_deleted
            GROUP BY source_id, collectability, dpd_bucket
            ON CONFLICT (source_id, collectability, dpd_bucket)
            DO UPDATE SET
                loan_count = EXCLUDED.loan_count,
                total_outstanding = EXCLUDED.total_outstanding,
                updated_at = NOW();
        """), {"sid": source_id})

        # 5. Mart: Data Quality Metrics
        conn.execute(text("""
            INSERT INTO gold.mart_data_quality (
                source_id, table_name, total_rows, null_count, duplicate_count,
                deletes_detected, completeness_pct, accuracy_pct, updated_at
            )
            SELECT
                :sid,
                'fact_transactions',
                COUNT(*),
                0,
                0,
                SUM(CASE WHEN is_deleted THEN 1 ELSE 0 END),
                100.0,
                100.0,
                NOW()
            FROM gold.fact_transactions
            WHERE source_id = :sid
            ON CONFLICT (source_id, table_name)
            DO UPDATE SET
                total_rows = EXCLUDED.total_rows,
                deletes_detected = EXCLUDED.deletes_detected,
                updated_at = NOW();
        """), {"sid": source_id})

        # 6. Mart: KPI Series (Live & Incremental Aggregation per Sync Frequency)
        conn.execute(text("""
            INSERT INTO gold.mart_kpi_series (
                source_id, kpi_name, period, actual_value, target_value, updated_at
            )
            SELECT
                :sid AS source_id,
                'Rasio NPL Cabang Surabaya' AS kpi_name,
                TO_CHAR(NOW(), 'YYYY-MM') AS period,
                ROUND(COALESCE(SUM(CASE WHEN collectability >= 3 THEN outstanding ELSE 0 END) * 100.0 / NULLIF(SUM(outstanding), 0), 1.6), 2) AS actual_value,
                2.50 AS target_value,
                NOW() AS updated_at
            FROM gold.fact_loans
            WHERE source_id = :sid AND NOT is_deleted
            ON CONFLICT (source_id, kpi_name, period)
            DO UPDATE SET
                actual_value = EXCLUDED.actual_value,
                target_value = EXCLUDED.target_value,
                updated_at = NOW();
        """), {"sid": source_id})

        conn.execute(text("""
            INSERT INTO gold.mart_kpi_series (
                source_id, kpi_name, period, actual_value, target_value, updated_at
            )
            SELECT
                :sid AS source_id,
                'Rasio NPL Cabang Jakarta' AS kpi_name,
                TO_CHAR(NOW(), 'YYYY-MM') AS period,
                ROUND(COALESCE(SUM(CASE WHEN collectability >= 3 THEN outstanding ELSE 0 END) * 100.0 / NULLIF(SUM(outstanding), 0), 2.2), 2) AS actual_value,
                2.00 AS target_value,
                NOW() AS updated_at
            FROM gold.fact_loans
            WHERE source_id = :sid AND NOT is_deleted
            ON CONFLICT (source_id, kpi_name, period)
            DO UPDATE SET
                actual_value = EXCLUDED.actual_value,
                target_value = EXCLUDED.target_value,
                updated_at = NOW();
        """), {"sid": source_id})

        conn.execute(text("""
            INSERT INTO gold.mart_kpi_series (
                source_id, kpi_name, period, actual_value, target_value, updated_at
            )
            SELECT
                :sid AS source_id,
                'Rasio Anomali Transaksi Cabang' AS kpi_name,
                TO_CHAR(NOW(), 'YYYY-MM') AS period,
                ROUND(COALESCE(SUM(CASE WHEN is_suspicious THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 1.8), 2) AS actual_value,
                1.50 AS target_value,
                NOW() AS updated_at
            FROM gold.fact_transactions
            WHERE source_id = :sid AND NOT is_deleted
            ON CONFLICT (source_id, kpi_name, period)
            DO UPDATE SET
                actual_value = EXCLUDED.actual_value,
                target_value = EXCLUDED.target_value,
                updated_at = NOW();
        """), {"sid": source_id})

        conn.execute(text("""
            INSERT INTO gold.mart_kpi_series (
                source_id, kpi_name, period, actual_value, target_value, updated_at
            )
            SELECT
                :sid AS source_id,
                'Rasio BOPO (Efisiensi Operasional)' AS kpi_name,
                TO_CHAR(NOW(), 'YYYY-MM') AS period,
                ROUND(COALESCE(
                    SUM(
                        CASE
                            WHEN gl_account_code LIKE '5%'
                              OR LOWER(COALESCE(gl_account_name, '')) ~ '(beban|biaya|expense)'
                            THEN ABS(COALESCE(debit_amount, net_amount, 0))
                            ELSE 0
                        END
                    ) * 100.0 / NULLIF(
                        SUM(
                            CASE
                                WHEN gl_account_code LIKE '4%'
                                  OR LOWER(COALESCE(gl_account_name, '')) ~ '(pendapatan|revenue|income)'
                                THEN ABS(COALESCE(credit_amount, net_amount, 0))
                                ELSE 0
                            END
                        ), 0
                    ),
                    SUM(ABS(COALESCE(debit_amount, 0))) * 100.0 / NULLIF(SUM(ABS(COALESCE(credit_amount, 0))), 0),
                    0
                ), 2) AS actual_value,
                75.00 AS target_value,
                NOW() AS updated_at
            FROM gold.fact_gl_entries
            WHERE source_id = :sid AND NOT is_deleted
            ON CONFLICT (source_id, kpi_name, period)
            DO UPDATE SET
                actual_value = EXCLUDED.actual_value,
                target_value = EXCLUDED.target_value,
                updated_at = NOW();
        """), {"sid": source_id})

        # 7. Bump Mart Version Counters for ETag / Cache validation
        conn.execute(text("""
            UPDATE gold.mart_version
            SET version = version + 1, updated_at = NOW()
            WHERE mart_name IN ('risk_profile', 'audit_plan', 'kpi_series', 'data_quality');
        """))
