"""
Data Hub API — Serve Router
High-performance REST endpoints serving Gold zone analytical marts and fact tables.
Features keyset cursor pagination, sub-second continuous aggregates, multi-source filtering,
and fast catalog-based row estimations to safely scale to 1 TB+ without query timeouts.
"""
import base64
from typing import Optional, List
from fastapi import APIRouter, Query, HTTPException, Response, Request
from sqlalchemy import text

from main import engine

router = APIRouter()


def _get_estimated_count(conn, schema: str, table_name: str) -> int:
    """Return instant O(1) row estimation using PostgreSQL catalog without table scan."""
    try:
        r = conn.execute(text("""
            SELECT reltuples::BIGINT 
            FROM pg_class c 
            JOIN pg_namespace n ON n.oid = c.relnamespace
            WHERE n.nspname = :schema AND c.relname = :table;
        """), {"schema": schema, "table": table_name})
        row = r.fetchone()
        return max(0, row[0]) if row and row[0] is not None else 0
    except Exception:
        return 0


# ─── 1. Transactions Fact Table (Hypertable with Keyset Cursor Pagination) ────
@router.get("/transactions")
def get_gold_transactions(
    limit: int = Query(100, le=5000),
    cursor: Optional[str] = Query(None, description="Keyset cursor in format base64(trx_date:trx_id)"),
    source_id: Optional[str] = Query(None),
    branch_id: Optional[int] = None,
    category: Optional[str] = None,
    suspicious_only: bool = False,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
):
    """Serve enriched transactions with keyset cursor pagination for 1 TB+ scale."""
    conditions = ["NOT is_deleted"]
    params = {"limit": limit}

    if source_id and source_id != "all":
        conditions.append("source_id = :sid")
        params["sid"] = source_id

    if branch_id:
        conditions.append("branch_id = :branch_id")
        params["branch_id"] = branch_id

    if category:
        conditions.append("category = :category")
        params["category"] = category

    if suspicious_only:
        conditions.append("is_suspicious = TRUE")

    if date_from:
        conditions.append("trx_date >= :date_from")
        params["date_from"] = date_from

    if date_to:
        conditions.append("trx_date <= :date_to")
        params["date_to"] = date_to

    # Keyset cursor decoding
    if cursor:
        try:
            decoded = base64.b64decode(cursor).decode("utf-8")
            cursor_date, cursor_id = decoded.split(":")
            conditions.append("(trx_date, trx_id) < (:cur_date, :cur_id)")
            params["cur_date"] = cursor_date
            params["cur_id"] = int(cursor_id)
        except Exception:
            pass

    where = "WHERE " + " AND ".join(conditions)

    try:
        with engine.connect() as conn:
            # Query with keyset ordering
            result = conn.execute(
                text(f"""
                    SELECT trx_id, source_id, trx_ref_number, trx_date, trx_hour, trx_day_of_week,
                           trx_type, category, channel, amount, amount_millions, branch_id,
                           branch_name, region, customer_name, customer_type, customer_risk,
                           account_type, authorization_status, is_suspicious, is_round_amount,
                           is_after_hours, benford_first_digit
                    FROM gold.fact_transactions
                    {where}
                    ORDER BY trx_date DESC, trx_id DESC
                    LIMIT :limit;
                """),
                params,
            )
            rows = [dict(r._mapping) for r in result]

            # Next cursor calculation
            next_cursor = None
            if len(rows) == limit:
                last_row = rows[-1]
                raw_cursor = f"{last_row['trx_date']}:{last_row['trx_id']}"
                next_cursor = base64.b64encode(raw_cursor.encode("utf-8")).decode("utf-8")

            # Instant estimated total when unfiltered, bounded total when filtered
            if len(conditions) == 1:
                total = _get_estimated_count(conn, "gold", "fact_transactions")
            else:
                # Fast bounded count to avoid full scan timeouts on multi-terabyte tables
                bounded_count = conn.execute(
                    text(f"""
                        SELECT COUNT(*) FROM (
                            SELECT 1 FROM gold.fact_transactions {where} LIMIT 50001
                        ) sub;
                    """),
                    {k: v for k, v in params.items() if k not in ("limit", "cur_date", "cur_id")},
                ).scalar() or 0
                total = bounded_count

        return {
            "status": "success",
            "data": rows,
            "next_cursor": next_cursor,
            "limit": limit,
            "estimated_total": total,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ─── 2. Loans Fact Table ──────────────────────────────────────────────────────
@router.get("/loans")
def get_gold_loans(
    limit: int = Query(100, le=5000),
    source_id: Optional[str] = Query(None),
    collectability: Optional[int] = None,
    status: Optional[str] = None,
):
    """Serve enriched loan portfolio data from Gold zone."""
    conditions = ["NOT is_deleted"]
    params = {"limit": limit}

    if source_id and source_id != "all":
        conditions.append("source_id = :sid")
        params["sid"] = source_id

    if collectability:
        conditions.append("collectability = :coll")
        params["coll"] = collectability

    if status:
        conditions.append("status = :status")
        params["status"] = status

    where = "WHERE " + " AND ".join(conditions)

    try:
        with engine.connect() as conn:
            result = conn.execute(
                text(f"""
                    SELECT loan_id, source_id, loan_number, loan_type, customer_name, customer_type,
                           branch_name, region, product_name, plafond, outstanding, interest_rate,
                           tenor_months, collateral_value, ltv_ratio, collectability, days_past_due,
                           status, rate_within_policy, plafond_within_policy, tenor_within_policy
                    FROM gold.fact_loans
                    {where}
                    ORDER BY outstanding DESC
                    LIMIT :limit;
                """),
                params,
            )
            return {"status": "success", "data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ─── 3. GL Entries Fact Table ────────────────────────────────────────────────
@router.get("/gl-entries")
def get_gold_gl_entries(
    limit: int = Query(500, le=5000),
    source_id: Optional[str] = Query(None),
    gaps_only: bool = False,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
):
    """Serve GL journal entries with sequential gap flags."""
    conditions = ["NOT is_deleted"]
    params = {"limit": limit}

    if source_id and source_id != "all":
        conditions.append("source_id = :sid")
        params["sid"] = source_id

    if gaps_only:
        conditions.append("has_gap = TRUE")

    if date_from:
        conditions.append("gl_date >= :date_from")
        params["date_from"] = date_from

    if date_to:
        conditions.append("gl_date <= :date_to")
        params["date_to"] = date_to

    where = "WHERE " + " AND ".join(conditions)

    try:
        with engine.connect() as conn:
            result = conn.execute(
                text(f"""
                    SELECT gl_id, source_id, gl_date, gl_account_code, gl_account_name, voucher_number,
                           debit_amount, credit_amount, net_amount, branch_name, has_voucher, has_gap
                    FROM gold.fact_gl_entries
                    {where}
                    ORDER BY gl_date DESC
                    LIMIT :limit;
                """),
                params,
            )
            return {"status": "success", "data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ─── 4. Precomputed Feature Marts (Sub-300ms SLA) ─────────────────────────────
@router.get("/risk-profile")
def get_entity_risk_profile(source_id: Optional[str] = Query(None)):
    """Precomputed Corporate Risk Profile from gold.mart_entity_risk_profile."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            res = conn.execute(text(f"""
                SELECT source_id, entity, entity_type, total_trx_volume, total_anomalies,
                       policy_violations, npl_loans, ai_predicted_score, risk_level, updated_at
                FROM gold.mart_entity_risk_profile
                {where}
                ORDER BY ai_predicted_score DESC;
            """), params)
            return {"status": "success", "data": [dict(r._mapping) for r in res]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/audit-plan")
def get_audit_plan_priorities(source_id: Optional[str] = Query(None)):
    """Precomputed Strategic Audit Plan priorities from gold.mart_audit_plan_priority."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            res = conn.execute(text(f"""
                SELECT source_id, entity_code, entity_name, risk_score, risk_level,
                       last_audit_year, priority_rank, recommended_cycle, updated_at
                FROM gold.mart_audit_plan_priority
                {where}
                ORDER BY priority_rank ASC;
            """), params)
            return {"status": "success", "data": [dict(r._mapping) for r in res]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/loan-aging")
def get_loan_aging_mart(source_id: Optional[str] = Query(None)):
    """Precomputed Loan Aging and NPL stratification from gold.mart_loan_aging."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            res = conn.execute(text(f"""
                SELECT source_id, collectability, dpd_bucket, loan_count, total_outstanding, updated_at
                FROM gold.mart_loan_aging
                {where}
                ORDER BY collectability ASC;
            """), params)
            return {"status": "success", "data": [dict(r._mapping) for r in res]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/data-quality")
def get_data_quality_mart(source_id: Optional[str] = Query(None)):
    """Precomputed data quality & source deletion monitoring from gold.mart_data_quality."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            res = conn.execute(text(f"""
                SELECT source_id, table_name, total_rows, null_count, duplicate_count,
                       deletes_detected, completeness_pct, accuracy_pct, timeliness_days, updated_at
                FROM gold.mart_data_quality
                {where}
                ORDER BY updated_at DESC;
            """), params)
            return {"status": "success", "data": [dict(r._mapping) for r in res]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ─── 5. Summary Statistics & Dimensions ──────────────────────────────────────
@router.get("/statistics")
def get_gold_statistics(source_id: Optional[str] = Query(None)):
    """Dashboard summary statistics with O(1) row estimates and multi-source awareness."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    and_where = "AND source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            stats = {}

            if not source_id or source_id == "all":
                stats["total_transactions"] = _get_estimated_count(conn, "gold", "fact_transactions")
                stats["total_loans"] = _get_estimated_count(conn, "gold", "fact_loans")
            else:
                stats["total_transactions"] = conn.execute(
                    text("SELECT COUNT(*) FROM gold.fact_transactions WHERE source_id = :sid AND NOT is_deleted"),
                    params
                ).scalar() or 0
                stats["total_loans"] = conn.execute(
                    text("SELECT COUNT(*) FROM gold.fact_loans WHERE source_id = :sid AND NOT is_deleted"),
                    params
                ).scalar() or 0

            stats["total_anomalies"] = conn.execute(
                text(f"SELECT COUNT(*) FROM gold.fact_transactions WHERE is_suspicious = TRUE AND NOT is_deleted {and_where}"),
                params
            ).scalar() or 0

            stats["npl_loans"] = conn.execute(
                text(f"SELECT COUNT(*) FROM gold.fact_loans WHERE collectability >= 3 AND NOT is_deleted {and_where}"),
                params
            ).scalar() or 0

            stats["policy_violations"] = conn.execute(
                text(f"SELECT COUNT(*) FROM gold.caatt_policy_violations {where}"),
                params
            ).scalar() or 0

            # Freshness logs
            r = conn.execute(text("""
                SELECT source_name, MAX(last_refreshed) as last_refreshed
                FROM gold.data_freshness_log
                GROUP BY source_name
                ORDER BY last_refreshed DESC
                LIMIT 10;
            """))
            stats["data_freshness"] = [dict(row._mapping) for row in r]

        return {"status": "success", "data": stats}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/branches")
def get_gold_branches(source_id: Optional[str] = Query(None)):
    """Branch dimension enriched metrics."""
    where = "WHERE source_id = :sid" if source_id and source_id != "all" else ""
    params = {"sid": source_id} if source_id and source_id != "all" else {}

    try:
        with engine.connect() as conn:
            result = conn.execute(text(f"""
                SELECT branch_id, source_id, branch_code, branch_name, branch_type, region, city,
                       total_accounts, total_customers, total_trx_volume, anomaly_count
                FROM gold.dim_branches
                {where}
                ORDER BY branch_name;
            """), params)
            return {"status": "success", "data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
