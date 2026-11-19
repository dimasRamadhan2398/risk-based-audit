"""
Data Hub API — CAATT Analytics Router
Endpoints serving CAATT test results from Gold zone to AuditSphere frontend.
Powers CAATT Analytics tabs with multi-source filtering, real Chi-Square forensic origins,
and automatic in-database audit execution.
"""
import json
from typing import Optional
from fastapi import APIRouter, Query, HTTPException
from pydantic import BaseModel
from sqlalchemy import text

from main import engine
from services.caatt_engine import (
    run_all_caatt_tests,
    run_benford_tests,
    CHI_SQUARE_CRIT_DF8_95,
    CHI_SQUARE_CRIT_DF1_95,
)

router = APIRouter()


class CaattRunRequest(BaseModel):
    source_id: Optional[str] = None


def _parse_details(val) -> dict:
    if isinstance(val, dict):
        return val
    if isinstance(val, str) and val.strip():
        try:
            return json.loads(val)
        except Exception:
            return {}
    return {}


def _resolve_active_source(conn, requested_source_id: Optional[str]) -> tuple:
    """
    Resolve effective (source_id, source_name) for CAATT queries.
    When explicit source_id is passed (other than 'all'/''), use it strictly;
    otherwise pick the latest synced external/internal source with Gold transactions.
    """
    if requested_source_id and requested_source_id not in ("all", "undefined", "null"):
        name_row = conn.execute(text("""
            SELECT name FROM bronze.registered_sources WHERE source_id = :sid LIMIT 1;
        """), {"sid": requested_source_id}).fetchone()
        sname = name_row[0] if name_row else (
            "Core Banking Simulator" if requested_source_id == "cbs_simulator" else requested_source_id
        )
        return requested_source_id, sname

    # Prefer latest external registered source that actually has Gold transactions
    ext_row = conn.execute(text("""
        SELECT t.source_id, COALESCE(rs.name, t.source_id)
        FROM gold.fact_transactions t
        LEFT JOIN bronze.registered_sources rs ON rs.source_id = t.source_id
        WHERE NOT t.is_deleted
        GROUP BY t.source_id, rs.name, rs.last_sync_at
        ORDER BY (t.source_id != 'cbs_simulator') DESC, rs.last_sync_at DESC NULLS LAST, COUNT(*) DESC
        LIMIT 1;
    """)).fetchone()
    if ext_row:
        return str(ext_row[0]), str(ext_row[1])

    # Fall back to whatever source already exists in CAATT result tables
    res_row = conn.execute(text("""
        SELECT b.source_id, COALESCE(rs.name, b.source_id)
        FROM gold.caatt_benford_results b
        LEFT JOIN bronze.registered_sources rs ON rs.source_id = b.source_id
        ORDER BY (b.source_id != 'cbs_simulator') DESC, b.test_date DESC
        LIMIT 1;
    """)).fetchone()
    if res_row:
        return str(res_row[0]), str(res_row[1])

    return None, None


def _ensure_caatt_for_source(source_id: str) -> None:
    """
    If source_id has Gold transactions but lacks complete Benford/CAATT results, trigger CAATT computation.
    Executes detection and computation in separate non-overlapping transactions to avoid locks.
    """
    if not source_id:
        return
    needs_run = False
    with engine.connect() as conn:
        chk = conn.execute(text("""
            SELECT
                COUNT(*) AS cnt,
                COALESCE(SUM(CASE WHEN COALESCE(expected_count, 0) > 0 AND details IS NOT NULL AND details != '{}'::jsonb THEN 1 ELSE 0 END), 0) AS valid_cnt
            FROM gold.caatt_benford_results
            WHERE source_id = :sid;
        """), {"sid": source_id}).fetchone()
        cnt = int(chk[0] or 0) if chk else 0
        try:
            valid_cnt = int(chk[1] or 0) if chk is not None else cnt
        except (IndexError, KeyError, TypeError):
            valid_cnt = cnt
        if cnt < 9 or valid_cnt < 9:
            has_trx = conn.execute(text("""
                SELECT 1 FROM gold.fact_transactions
                WHERE source_id = :sid AND NOT is_deleted AND amount > 0
                LIMIT 1;
            """), {"sid": source_id}).fetchone() is not None
            needs_run = has_trx

    if needs_run:
        run_all_caatt_tests(engine, source_id=source_id)


def _ensure_benford_for_source(source_id: str) -> None:
    """Compute Benford Chi-Square results for source_id if missing, partial, or lacking expected_count/details."""
    if not source_id:
        return
    needs_run = False
    with engine.connect() as conn:
        chk = conn.execute(text("""
            SELECT
                COUNT(*) AS cnt,
                COALESCE(SUM(CASE WHEN COALESCE(expected_count, 0) > 0 AND details IS NOT NULL AND details != '{}'::jsonb THEN 1 ELSE 0 END), 0) AS valid_cnt
            FROM gold.caatt_benford_results
            WHERE source_id = :sid;
        """), {"sid": source_id}).fetchone()
        cnt = int(chk[0] or 0) if chk else 0
        try:
            valid_cnt = int(chk[1] or 0) if chk is not None else cnt
        except (IndexError, KeyError, TypeError):
            valid_cnt = cnt
        if cnt < 9 or valid_cnt < 9:
            has_trx = conn.execute(text("""
                SELECT 1 FROM gold.fact_transactions
                WHERE source_id = :sid AND NOT is_deleted AND amount > 0
                LIMIT 1;
            """), {"sid": source_id}).fetchone() is not None
            needs_run = has_trx

    if needs_run:
        run_benford_tests(engine, source_id)


@router.post("/run")
def trigger_caatt_tests(
    req: Optional[CaattRunRequest] = None,
    source_id: Optional[str] = Query(None),
):
    """Trigger the in-database CAATT audit engine across all CAATT audit techniques."""
    req_sid = source_id or (req.source_id if req and req.source_id else None)
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, req_sid)
        if not eff_sid:
            return {"status": "empty", "source_id": None, "message": "No ingested transaction sources available yet.", "results": {}}
        results = run_all_caatt_tests(engine, eff_sid)
        return {"status": "success", "source_id": eff_sid, "source_name": sname, "results": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/full-population")
def get_full_population_results(
    source_id: Optional[str] = Query(None),
    limit: int = Query(100, le=5000),
):
    """CAATT #1: Full population testing results — real limit/timing/structuring exceptions."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"total_tested": 0, "total_violations": 0, "avg_violation_rate": 0, "branches_tested": 0, "categories_tested": 0}}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            trx_rows = conn.execute(text("""
                SELECT
                    COALESCE(trx_ref_number, 'TRX-' || trx_id::TEXT) AS id,
                    trx_date::TEXT AS test_date,
                    COALESCE(branch_name, '-') AS branch_name,
                    COALESCE(trx_ref_number, 'ACC-' || trx_id::TEXT) AS account_id,
                    COALESCE(customer_name, '-') AS customer_name,
                    COALESCE(category, 'General') AS category,
                    COALESCE(amount, 0)::FLOAT AS transaction_amount,
                    500000000.0 AS threshold_limit,
                    GREATEST(0, COALESCE(amount, 0) - 500000000)::FLOAT AS excess_amount,
                    CASE
                        WHEN amount > 500000000 THEN 'Exceeds Single Transaction Threshold (> Rp 500M)'
                        WHEN is_suspicious THEN 'Flagged Suspicious Activity (AML)'
                        WHEN is_after_hours THEN 'After-Hours Transaction Anomaly'
                        WHEN is_round_amount THEN 'Round Amount Structuring Pattern'
                        ELSE 'Rule Exception'
                    END AS violation_type,
                    CASE WHEN is_suspicious THEN 'ESCALATED' ELSE 'FLAGGED' END AS status
                FROM gold.fact_transactions
                WHERE source_id = :sid
                  AND NOT is_deleted
                  AND (amount > 500000000 OR is_suspicious OR is_after_hours OR is_round_amount)
                ORDER BY is_suspicious DESC, amount DESC, trx_date DESC
                LIMIT :limit;
            """), {"sid": eff_sid, "limit": limit}).fetchall()
            rows = [dict(r._mapping) for r in trx_rows]

            pop_stats = conn.execute(text("""
                SELECT
                    COUNT(*) AS total_tested,
                    SUM(CASE WHEN (amount > 500000000 OR is_suspicious OR is_after_hours OR is_round_amount) THEN 1 ELSE 0 END) AS total_violations,
                    COUNT(DISTINCT branch_name) AS branches_tested,
                    COUNT(DISTINCT category) AS categories_tested
                FROM gold.fact_transactions
                WHERE source_id = :sid AND NOT is_deleted;
            """), {"sid": eff_sid}).fetchone()

            tot_tested = int(pop_stats[0] or 0) if pop_stats else 0
            tot_viol = int(pop_stats[1] or 0) if pop_stats else 0
            avg_rate = round((tot_viol / max(1, tot_tested)) * 100.0, 2) if tot_tested > 0 else 0.0

        return {
            "data": rows,
            "summary": {
                "total_tested": tot_tested,
                "total_violations": tot_viol,
                "avg_violation_rate": avg_rate,
                "branches_tested": int(pop_stats[2] or 0) if pop_stats else 0,
                "categories_tested": int(pop_stats[3] or 0) if pop_stats else 0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/duplicate-gap")
def get_duplicate_gap_results(
    result_type: Optional[str] = Query(None, description="DUPLICATE or GAP"),
    source_id: Optional[str] = Query(None),
    limit: int = Query(100, le=5000),
):
    """CAATT #2: Duplicate transactions and document number gaps."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"total_duplicates": 0, "total_gaps": 0, "branches_affected": 0}}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            conditions = ["source_id = :sid"]
            params = {"sid": eff_sid, "limit": limit}
            if result_type:
                conditions.append("result_type = :rtype")
                params["rtype"] = result_type.upper()
            where = "WHERE " + " AND ".join(conditions)

            raw_rows = conn.execute(text(f"""
                SELECT * FROM gold.caatt_duplicate_gap_results
                {where}
                ORDER BY test_date DESC, duplicate_count DESC
                LIMIT :limit;
            """), params).fetchall()

            rows = []
            for r in raw_rows:
                m = dict(r._mapping)
                det = _parse_details(m.get("details"))
                rtype = (m.get("result_type") or "DUPLICATE").upper()
                if rtype == "GAP":
                    gs = m.get("gap_start")
                    ge = m.get("gap_end")
                    if gs and ge and gs != ge:
                        desc = f"Celah nomor urut dokumen hilang dari {gs} s/d {ge}"
                    elif gs:
                        desc = f"Dokumen/voucher bermasalah atau hilang sekitar {gs}"
                    else:
                        desc = "Nomor voucher General Ledger kosong / hilang (MISSING voucher)"
                else:
                    desc = f"Duplikasi pada {m.get('reference_field') or 'reference'} ({m.get('duplicate_count') or 2}x kejadian) - {det.get('customer_name') or m.get('branch_name') or '-'}"
                rows.append({
                    "id": f"DG-{m.get('id')}",
                    "test_date": str(m.get("test_date"))[:10] if m.get("test_date") else "",
                    "result_type": rtype,
                    "branch_name": m.get("branch_name") or "-",
                    "reference_no": m.get("reference_value") or "-",
                    "account_id": det.get("gl_account") or det.get("customer_name") or "-",
                    "amount": float(det.get("total_amount") or 0),
                    "description": desc,
                    "occurrences": int(m.get("duplicate_count") or 1),
                    "status": "OPEN",
                })

            summary_row = conn.execute(text(f"""
                SELECT
                    SUM(CASE WHEN result_type = 'DUPLICATE' THEN GREATEST(1, duplicate_count) ELSE 0 END) AS total_duplicates,
                    SUM(CASE WHEN result_type = 'GAP' THEN 1 ELSE 0 END) AS total_gaps,
                    COUNT(DISTINCT branch_name) AS branches_affected
                FROM gold.caatt_duplicate_gap_results
                WHERE source_id = :sid;
            """), {"sid": eff_sid}).fetchone()

        return {
            "data": rows,
            "summary": {
                "total_duplicates": int(summary_row[0] or 0) if summary_row else 0,
                "total_gaps": int(summary_row[1] or 0) if summary_row else 0,
                "branches_affected": int(summary_row[2] or 0) if summary_row else 0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/benford-analysis")
def get_benford_results(
    source_id: Optional[str] = Query(None),
    digit: Optional[int] = Query(None, ge=1, le=9),
    findings_limit: int = Query(100, ge=1, le=500),
):
    """
    CAATT #3: Real Benford's Law digit distribution, Chi-Square forensic explanation per digit (1-9),
    and actual database transaction rows behind any deviations.
    """
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {
                "data": [],
                "findings": [],
                "summary": {
                    "total_digits_analyzed": 0,
                    "sample_size": 0,
                    "total_chi_square": 0.0,
                    "critical_value": CHI_SQUARE_CRIT_DF8_95,
                    "per_digit_critical_value": CHI_SQUARE_CRIT_DF1_95,
                    "significant_deviations": 0,
                    "conclusion": "NO_DATA",
                    "active_source_id": None,
                    "source_name": None,
                },
            }

        # Run Benford computation in its own transaction before reading results (prevents lock deadlock)
        _ensure_benford_for_source(eff_sid)

        with engine.connect() as conn:
            raw_rows = conn.execute(text("""
                SELECT *
                FROM gold.caatt_benford_results
                WHERE source_id = :sid
                ORDER BY digit ASC;
            """), {"sid": eff_sid}).fetchall()

            data_rows = []
            sig_count = 0
            total_chi_sq = 0.0
            sample_size = 0
            low_sample_warning = False

            for r in raw_rows:
                m = dict(r._mapping)
                det = _parse_details(m.get("details"))
                actual_count = int(m.get("actual_count") or 0)
                expected_count = float(m.get("expected_count") or 0.0)
                chi_val = float(m.get("chi_square") or 0.0)
                total_chi_sq += chi_val
                sample_size = max(sample_size, int(m.get("sample_size") or 0))
                if det.get("low_sample_warning"):
                    low_sample_warning = True
                is_sig = bool(m.get("is_significant"))
                if is_sig:
                    sig_count += 1

                data_rows.append({
                    "digit": int(m.get("digit") or 0),
                    "actual_count": actual_count,
                    "expected_count": round(expected_count, 2),
                    "count_diff": round(float(det.get("count_diff", actual_count - expected_count)), 2),
                    "actual_pct": float(m.get("actual_pct") or 0.0),
                    "expected_pct": float(m.get("expected_pct") or 0.0),
                    "deviation_pct": float(m.get("deviation_pct") or 0.0),
                    "chi_square": round(chi_val, 4),
                    "chi_square_share_pct": float(m.get("chi_square_share_pct") or 0.0),
                    "is_significant": is_sig,
                    "dominant_branch": det.get("dominant_branch") or "-",
                    "dominant_branch_count": int(det.get("dominant_branch_count") or 0),
                    "dominant_category": det.get("dominant_category") or "-",
                    "dominant_category_count": int(det.get("dominant_category_count") or 0),
                    "total_amount": float(det.get("total_amount") or 0.0),
                    "avg_amount": float(det.get("avg_amount") or 0.0),
                    "round_amount_count": int(det.get("round_amount_count") or 0),
                    "after_hours_count": int(det.get("after_hours_count") or 0),
                    "suspicious_count": int(det.get("suspicious_count") or 0),
                    "direction": det.get("direction") or "balanced",
                    "finding_origin": det.get("finding_origin") or "",
                })

            total_chi_sq = round(total_chi_sq, 4)

            # Query REAL underlying transactions from gold.fact_transactions (never simulated).
            # Derive effective leading digit cleanly even when legacy rows lacked benford_first_digit.
            sig_digits = [int(row["digit"]) for row in data_rows if row["is_significant"] and 1 <= int(row["digit"]) <= 9]
            finding_params = {"sid": eff_sid, "limit": findings_limit}
            digit_expr = """COALESCE(
                CASE WHEN benford_first_digit BETWEEN 1 AND 9 THEN benford_first_digit END,
                NULLIF(SUBSTRING(TRIM(LEADING '0.' FROM ABS(amount)::NUMERIC::TEXT) FROM '^([1-9])'), '')::INT
            )"""
            if digit is not None:
                finding_params["digit"] = digit
                real_findings_rows = conn.execute(text(f"""
                    WITH tx_with_digit AS (
                        SELECT
                            trx_id,
                            COALESCE(trx_ref_number, 'TRX-' || trx_id::TEXT) AS trx_ref_number,
                            trx_date::TEXT AS trx_date,
                            {digit_expr} AS digit,
                            amount::FLOAT AS amount,
                            COALESCE(branch_name, '-') AS branch_name,
                            COALESCE(category, '-') AS category,
                            COALESCE(customer_name, '-') AS customer_name,
                            COALESCE(channel, '-') AS channel,
                            COALESCE(is_round_amount, FALSE) AS is_round_amount,
                            COALESCE(is_after_hours, FALSE) AS is_after_hours,
                            COALESCE(is_suspicious, FALSE) AS is_suspicious,
                            COALESCE(authorization_status, 'Authorized') AS authorization_status
                        FROM gold.fact_transactions
                        WHERE source_id = :sid
                          AND NOT is_deleted
                          AND amount > 0
                    )
                    SELECT * FROM tx_with_digit
                    WHERE digit = :digit
                    ORDER BY is_suspicious DESC, is_round_amount DESC, amount DESC, trx_date DESC
                    LIMIT :limit;
                """), finding_params).fetchall()
            else:
                finding_params["per_digit_cap"] = max(12, min(25, findings_limit // 5))
                if sig_digits:
                    sig_list = ",".join(str(d) for d in sig_digits)
                    sig_order_expr = f"(digit IN ({sig_list})) DESC,"
                else:
                    sig_order_expr = ""
                real_findings_rows = conn.execute(text(f"""
                    WITH tx_with_digit AS (
                        SELECT
                            trx_id,
                            COALESCE(trx_ref_number, 'TRX-' || trx_id::TEXT) AS trx_ref_number,
                            trx_date::TEXT AS trx_date,
                            {digit_expr} AS digit,
                            amount::FLOAT AS amount,
                            COALESCE(branch_name, '-') AS branch_name,
                            COALESCE(category, '-') AS category,
                            COALESCE(customer_name, '-') AS customer_name,
                            COALESCE(channel, '-') AS channel,
                            COALESCE(is_round_amount, FALSE) AS is_round_amount,
                            COALESCE(is_after_hours, FALSE) AS is_after_hours,
                            COALESCE(is_suspicious, FALSE) AS is_suspicious,
                            COALESCE(authorization_status, 'Authorized') AS authorization_status
                        FROM gold.fact_transactions
                        WHERE source_id = :sid
                          AND NOT is_deleted
                          AND amount > 0
                    ),
                    ranked AS (
                        SELECT
                            trx_id, trx_ref_number, trx_date, digit, amount,
                            branch_name, category, customer_name, channel,
                            is_round_amount, is_after_hours, is_suspicious, authorization_status,
                            ROW_NUMBER() OVER (
                                PARTITION BY digit
                                ORDER BY is_suspicious DESC, is_round_amount DESC, amount DESC, trx_date DESC
                            ) AS rn
                        FROM tx_with_digit
                        WHERE digit BETWEEN 1 AND 9
                    )
                    SELECT
                        trx_id, trx_ref_number, trx_date, digit, amount,
                        branch_name, category, customer_name, channel,
                        is_round_amount, is_after_hours, is_suspicious, authorization_status
                    FROM ranked
                    WHERE rn <= :per_digit_cap
                    ORDER BY {sig_order_expr} is_suspicious DESC, is_round_amount DESC, amount DESC, trx_date DESC
                    LIMIT :limit;
                """), finding_params).fetchall()
            findings = [dict(fr._mapping) for fr in real_findings_rows]

            gold_trx_row = conn.execute(text("""
                SELECT COUNT(*) FROM gold.fact_transactions
                WHERE source_id = :sid AND NOT is_deleted AND amount > 0;
            """), {"sid": eff_sid}).fetchone()
            gold_transactions_count = int(gold_trx_row[0] or 0) if gold_trx_row else 0

            silver_records_count = 0
            try:
                silver_row = conn.execute(text("""
                    SELECT COUNT(*) FROM silver.external_source_data
                    WHERE source_id = :sid AND NOT is_deleted;
                """), {"sid": eff_sid}).fetchone()
                silver_records_count = int(silver_row[0] or 0) if silver_row else 0
            except Exception:
                silver_records_count = 0

            total_chi_sq = round(total_chi_sq, 4)

            if not data_rows:
                conclusion = "NO_DATA"
            elif total_chi_sq > CHI_SQUARE_CRIT_DF8_95 or sig_count >= 2:
                conclusion = "ANOMALY DETECTED"
            elif sig_count == 1:
                conclusion = "MINOR DEVIATION"
            else:
                conclusion = "NORMAL"

        return {
            "data": data_rows,
            "findings": findings,
            "summary": {
                "total_digits_analyzed": len(data_rows),
                "sample_size": sample_size,
                "total_chi_square": total_chi_sq,
                "critical_value": CHI_SQUARE_CRIT_DF8_95,
                "per_digit_critical_value": CHI_SQUARE_CRIT_DF1_95,
                "significant_deviations": sig_count,
                "conclusion": conclusion,
                "active_source_id": eff_sid,
                "source_name": sname,
                "low_sample_warning": low_sample_warning,
                "gold_transactions_count": gold_transactions_count,
                "silver_records_count": silver_records_count,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/stratification")
def get_stratification_results(
    category: Optional[str] = None,
    source_id: Optional[str] = Query(None),
):
    """CAATT #4: Transaction stratification and aging analysis."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"total_transactions": 0, "total_amount": 0, "total_strata": 0}}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            conditions = ["source_id = :sid"]
            params = {"sid": eff_sid}
            if category:
                conditions.append("category = :category")
                params["category"] = category
            where = "WHERE " + " AND ".join(conditions)

            result = conn.execute(text(f"""
                WITH grouped AS (
                    SELECT
                        stratum_label,
                        MIN(min_value) AS min_value,
                        MAX(max_value) AS max_value,
                        category,
                        SUM(trx_count) AS trx_count,
                        SUM(total_amount)::FLOAT AS total_amount
                    FROM gold.caatt_stratification_results
                    {where}
                    GROUP BY stratum_label, category
                ),
                tot AS (
                    SELECT
                        GREATEST(1, SUM(trx_count)) AS grand_cnt,
                        GREATEST(1.0, SUM(total_amount)) AS grand_amt
                    FROM grouped
                )
                SELECT
                    g.stratum_label,
                    g.category,
                    g.trx_count,
                    g.total_amount,
                    ROUND(((g.trx_count::NUMERIC / t.grand_cnt::NUMERIC) * 100)::NUMERIC, 2)::FLOAT AS pct_count,
                    ROUND(((g.total_amount::NUMERIC / t.grand_amt::NUMERIC) * 100)::NUMERIC, 2)::FLOAT AS pct_amount
                FROM grouped g CROSS JOIN tot t
                ORDER BY g.min_value ASC, g.total_amount DESC;
            """), params)
            rows = [dict(r._mapping) for r in result]

            summary_row = conn.execute(text(f"""
                SELECT
                    COALESCE(SUM(trx_count), 0) AS total_transactions,
                    COALESCE(SUM(total_amount), 0)::FLOAT AS total_amount,
                    COUNT(DISTINCT stratum_label) AS total_strata,
                    COUNT(DISTINCT category) AS categories
                FROM gold.caatt_stratification_results
                {where};
            """), params).fetchone()

        return {
            "data": rows,
            "summary": {
                "total_transactions": int(summary_row[0] or 0) if summary_row else 0,
                "total_amount": float(summary_row[1] or 0.0) if summary_row else 0.0,
                "total_strata": int(summary_row[2] or 0) if summary_row else 0,
                "categories": int(summary_row[3] or 0) if summary_row else 0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/reconciliation")
def get_reconciliation_results(source_id: Optional[str] = Query(None)):
    """CAATT #5: Cross-system reconciliation results."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"avg_match_rate": 0, "total_unmatched": 0, "total_difference": 0}}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            raw_rows = conn.execute(text("""
                SELECT * FROM gold.caatt_reconciliation_results
                WHERE source_id = :sid
                ORDER BY test_date DESC;
            """), {"sid": eff_sid}).fetchall()

            rows = []
            for r in raw_rows:
                m = dict(r._mapping)
                det = _parse_details(m.get("details"))
                matched = int(m.get("matched_count") or 0)
                ua = int(m.get("unmatched_a") or 0)
                ub = int(m.get("unmatched_b") or 0)
                rate = float(m.get("match_rate_pct") or 0.0)
                rows.append({
                    "module": det.get("module") or f"{m.get('source_a')} vs {m.get('source_b')}",
                    "system_a": m.get("source_a") or "System A",
                    "system_b": m.get("source_b") or "System B",
                    "total_records_a": int(det.get("total_records_a", matched + ua)),
                    "total_records_b": int(det.get("total_records_b", matched + ub)),
                    "matched_records": matched,
                    "match_rate_pct": rate,
                    "total_difference": float(m.get("total_difference") or 0.0),
                    "status": "BALANCED" if rate >= 99.5 else "DISCREPANCY",
                })

            summary_row = conn.execute(text("""
                SELECT
                    COALESCE(ROUND(AVG(match_rate_pct), 2), 0)::FLOAT AS avg_match_rate,
                    COALESCE(SUM(unmatched_a + unmatched_b), 0) AS total_unmatched,
                    COALESCE(SUM(total_difference), 0)::FLOAT AS total_difference
                FROM gold.caatt_reconciliation_results
                WHERE source_id = :sid;
            """), {"sid": eff_sid}).fetchone()

        return {
            "data": rows,
            "summary": {
                "avg_match_rate": float(summary_row[0] or 0.0) if summary_row else 0.0,
                "total_unmatched": int(summary_row[1] or 0) if summary_row else 0,
                "total_difference": float(summary_row[2] or 0.0) if summary_row else 0.0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/policy-violations")
def get_caatt_policy_violations(
    severity: Optional[str] = None,
    source_id: Optional[str] = Query(None),
    limit: int = Query(100, le=5000),
):
    """CAATT #10: Policy and rule engine violation results."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"total_violations": 0, "critical": 0, "high": 0, "medium": 0, "branches_affected": 0}}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            conditions = ["source_id = :sid"]
            params = {"sid": eff_sid, "limit": limit}
            if severity:
                conditions.append("LOWER(severity) = LOWER(:severity)")
                params["severity"] = severity
            where = "WHERE " + " AND ".join(conditions)

            raw_rows = conn.execute(text(f"""
                SELECT * FROM gold.caatt_policy_violations
                {where}
                ORDER BY test_date DESC, id DESC
                LIMIT :limit;
            """), params).fetchall()

            rows = []
            for r in raw_rows:
                m = dict(r._mapping)
                det = _parse_details(m.get("details"))
                sev = (m.get("severity") or "Medium").capitalize()
                rows.append({
                    "id": f"POL-{m.get('id')}",
                    "test_date": str(m.get("test_date"))[:10] if m.get("test_date") else "",
                    "rule_name": m.get("rule_name") or m.get("violation_type") or "Policy Breach",
                    "branch_name": m.get("branch_name") or "-",
                    "customer_name": det.get("customer_name") or "-",
                    "severity": sev,
                    "description": det.get("description") or f"{m.get('violation_type')}: value {m.get('actual_value')}",
                    "status": "OPEN",
                })

            summary_row = conn.execute(text("""
                SELECT
                    COUNT(*) AS total_violations,
                    SUM(CASE WHEN LOWER(severity) = 'critical' THEN 1 ELSE 0 END) AS critical,
                    SUM(CASE WHEN LOWER(severity) = 'high' THEN 1 ELSE 0 END) AS high,
                    SUM(CASE WHEN LOWER(severity) = 'medium' THEN 1 ELSE 0 END) AS medium,
                    SUM(CASE WHEN LOWER(severity) = 'low' THEN 1 ELSE 0 END) AS low,
                    COUNT(DISTINCT rule_name) AS unique_rules_violated,
                    COUNT(DISTINCT branch_name) AS branches_affected
                FROM gold.caatt_policy_violations
                WHERE source_id = :sid;
            """), {"sid": eff_sid}).fetchone()

        return {
            "data": rows,
            "summary": {
                "total_violations": int(summary_row[0] or 0) if summary_row else 0,
                "critical": int(summary_row[1] or 0) if summary_row else 0,
                "high": int(summary_row[2] or 0) if summary_row else 0,
                "medium": int(summary_row[3] or 0) if summary_row else 0,
                "low": int(summary_row[4] or 0) if summary_row else 0,
                "unique_rules_violated": int(summary_row[5] or 0) if summary_row else 0,
                "branches_affected": int(summary_row[6] or 0) if summary_row else 0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/data-quality")
def get_data_quality_metrics(source_id: Optional[str] = Query(None)):
    """CAATT #9: Data quality dashboard — completeness, accuracy, timeliness per table."""
    try:
        with engine.connect() as conn:
            eff_sid, sname = _resolve_active_source(conn, source_id)
        if not eff_sid:
            return {"data": [], "summary": {"avg_completeness": 0, "avg_accuracy": 0, "avg_timeliness_days": 0, "tables_profiled": 0, "total_rows_profiled": 0}, "overall_quality_score": 0.0}
        _ensure_caatt_for_source(eff_sid)

        with engine.connect() as conn:
            raw_rows = conn.execute(text("""
                SELECT * FROM gold.caatt_data_quality_metrics
                WHERE source_id = :sid
                ORDER BY total_rows DESC, table_name ASC;
            """), {"sid": eff_sid}).fetchall()

            rows = []
            for r in raw_rows:
                m = dict(r._mapping)
                det = _parse_details(m.get("details"))
                comp = float(m.get("completeness_pct") or 0.0)
                acc = float(m.get("accuracy_pct") or 0.0)
                avg_score = (comp + acc) / 2.0
                rows.append({
                    "table_name": m.get("table_name") or "table",
                    "zone": det.get("zone") or "Gold",
                    "total_rows": int(m.get("total_rows") or 0),
                    "completeness_pct": comp,
                    "accuracy_pct": acc,
                    "timeliness_days": float(m.get("timeliness_days") or 0.0),
                    "duplicate_count": int(m.get("duplicate_count") or 0),
                    "status": "GOOD" if avg_score >= 95.0 else ("WARNING" if avg_score >= 85.0 else "POOR"),
                })

            summary_row = conn.execute(text("""
                SELECT
                    COALESCE(ROUND(AVG(completeness_pct), 2), 0)::FLOAT AS avg_completeness,
                    COALESCE(ROUND(AVG(accuracy_pct), 2), 0)::FLOAT AS avg_accuracy,
                    COALESCE(ROUND(AVG(timeliness_days), 1), 0)::FLOAT AS avg_timeliness_days,
                    COUNT(DISTINCT table_name) AS tables_profiled,
                    COALESCE(SUM(total_rows), 0) AS total_rows_profiled
                FROM gold.caatt_data_quality_metrics
                WHERE source_id = :sid;
            """), {"sid": eff_sid}).fetchone()

            overall_score = 0.0
            if summary_row and summary_row[0] is not None and int(summary_row[3] or 0) > 0:
                overall_score = round((float(summary_row[0]) + float(summary_row[1])) / 2.0, 2)

        return {
            "data": rows,
            "summary": {
                "avg_completeness": float(summary_row[0] or 0.0) if summary_row else 0.0,
                "avg_accuracy": float(summary_row[1] or 0.0) if summary_row else 0.0,
                "avg_timeliness_days": float(summary_row[2] or 0.0) if summary_row else 0.0,
                "tables_profiled": int(summary_row[3] or 0) if summary_row else 0,
                "total_rows_profiled": int(summary_row[4] or 0) if summary_row else 0,
                "active_source_id": eff_sid,
                "source_name": sname,
            },
            "overall_quality_score": overall_score,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
