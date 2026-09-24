"""
Data Hub API — Serve Router
GET endpoints to serve Gold zone data to AuditSphere and other consumers.
"""
from typing import Optional
from fastapi import APIRouter, Query, HTTPException
from sqlalchemy import text

from main import engine

router = APIRouter()


@router.get("/transactions")
def get_gold_transactions(
    limit: int = Query(100, le=5000),
    offset: int = Query(0),
    branch_id: Optional[int] = None,
    category: Optional[str] = None,
    suspicious_only: bool = False,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
):
    """Serve enriched transactions from Gold zone."""
    conditions = []
    params = {"limit": limit, "offset": offset}

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

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(
                text(f"SELECT * FROM gold.fact_transactions {where} ORDER BY trx_date DESC LIMIT :limit OFFSET :offset"),
                params,
            )
            rows = [dict(r._mapping) for r in result]

            count_result = conn.execute(
                text(f"SELECT COUNT(*) FROM gold.fact_transactions {where}"),
                {k: v for k, v in params.items() if k not in ("limit", "offset")},
            )
            total = count_result.scalar()

        return {"data": rows, "total": total, "limit": limit, "offset": offset}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/loans")
def get_gold_loans(
    limit: int = Query(100, le=5000),
    collectability: Optional[int] = None,
    status: Optional[str] = None,
):
    """Serve enriched loan data from Gold zone."""
    conditions = []
    params = {"limit": limit}

    if collectability:
        conditions.append("collectability = :coll")
        params["coll"] = collectability
    if status:
        conditions.append("status = :status")
        params["status"] = status

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(
                text(f"SELECT * FROM gold.fact_loans {where} ORDER BY outstanding DESC LIMIT :limit"),
                params,
            )
            return {"data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/gl-entries")
def get_gold_gl_entries(
    limit: int = Query(500, le=5000),
    gaps_only: bool = False,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
):
    """Serve GL entries with gap analysis from Gold zone."""
    conditions = []
    params = {"limit": limit}

    if gaps_only:
        conditions.append("has_gap = TRUE")
    if date_from:
        conditions.append("gl_date >= :date_from")
        params["date_from"] = date_from
    if date_to:
        conditions.append("gl_date <= :date_to")
        params["date_to"] = date_to

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(
                text(f"SELECT * FROM gold.fact_gl_entries {where} ORDER BY gl_date DESC LIMIT :limit"),
                params,
            )
            return {"data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/statistics")
def get_gold_statistics():
    """Dashboard summary statistics from Gold zone."""
    try:
        with engine.connect() as conn:
            stats = {}

            # Transaction stats
            r = conn.execute(text("SELECT COUNT(*) FROM gold.fact_transactions"))
            stats["total_transactions"] = r.scalar() or 0

            r = conn.execute(text("SELECT COUNT(*) FROM gold.fact_transactions WHERE is_suspicious = TRUE"))
            stats["total_anomalies"] = r.scalar() or 0

            # Loan stats
            r = conn.execute(text("SELECT COUNT(*) FROM gold.fact_loans"))
            stats["total_loans"] = r.scalar() or 0

            r = conn.execute(text("SELECT COUNT(*) FROM gold.fact_loans WHERE collectability >= 3"))
            stats["npl_loans"] = r.scalar() or 0

            # CAATT stats
            r = conn.execute(text("SELECT COUNT(*) FROM gold.caatt_policy_violations"))
            stats["policy_violations"] = r.scalar() or 0

            r = conn.execute(text("SELECT COUNT(*) FROM gold.caatt_duplicate_gap_results WHERE result_type = 'DUPLICATE'"))
            stats["duplicate_transactions"] = r.scalar() or 0

            r = conn.execute(text("SELECT COUNT(*) FROM gold.caatt_duplicate_gap_results WHERE result_type = 'GAP'"))
            stats["document_gaps"] = r.scalar() or 0

            # Data freshness
            r = conn.execute(text("""
                SELECT source_name, MAX(last_refreshed) as last_refreshed
                FROM gold.data_freshness_log
                GROUP BY source_name
                ORDER BY last_refreshed DESC
            """))
            stats["data_freshness"] = [dict(row._mapping) for row in r]

        return {"status": "success", "data": stats}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/branches")
def get_gold_branches():
    """All branches with enriched metrics."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text("SELECT * FROM gold.dim_branches ORDER BY branch_name"))
            return {"data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/policy-violations")
def get_policy_violations(limit: int = Query(100, le=5000)):
    """Policy violations detected by CAATT."""
    try:
        with engine.connect() as conn:
            result = conn.execute(
                text("SELECT * FROM gold.caatt_policy_violations ORDER BY test_date DESC LIMIT :limit"),
                {"limit": limit},
            )
            return {"data": [dict(r._mapping) for r in result]}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
