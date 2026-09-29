"""
Data Hub API — CAATT Analytics Router
Endpoints serving CAATT test results from Gold zone to AuditSphere frontend.
These 7 endpoints power the 7 new CAATT Analytics tabs.
"""
from typing import Optional
from fastapi import APIRouter, Query, HTTPException
from sqlalchemy import text

from main import engine

router = APIRouter()


@router.get("/full-population")
def get_full_population_results(limit: int = Query(100, le=5000)):
    """CAATT #1: Full population testing results — violations found across 100% of transactions."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text("""
                SELECT * FROM gold.caatt_full_population_results
                ORDER BY test_date DESC LIMIT :limit
            """), {"limit": limit})
            rows = [dict(r._mapping) for r in result]

            # Summary stats
            summary = conn.execute(text("""
                SELECT
                    SUM(total_population) AS total_tested,
                    SUM(violations_found) AS total_violations,
                    ROUND(AVG(violation_rate) * 100, 2) AS avg_violation_rate,
                    COUNT(DISTINCT branch_name) AS branches_tested,
                    COUNT(DISTINCT category) AS categories_tested
                FROM gold.caatt_full_population_results
            """))
            summary_row = summary.fetchone()

        return {
            "data": rows,
            "summary": dict(summary_row._mapping) if summary_row else {},
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/duplicate-gap")
def get_duplicate_gap_results(
    result_type: Optional[str] = Query(None, description="DUPLICATE or GAP"),
    limit: int = Query(100, le=5000),
):
    """CAATT #2: Duplicate transactions and document number gaps."""
    conditions = []
    params = {"limit": limit}

    if result_type:
        conditions.append("result_type = :rtype")
        params["rtype"] = result_type.upper()

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(text(f"""
                SELECT * FROM gold.caatt_duplicate_gap_results
                {where} ORDER BY test_date DESC LIMIT :limit
            """), params)
            rows = [dict(r._mapping) for r in result]

            summary = conn.execute(text("""
                SELECT
                    SUM(CASE WHEN result_type = 'DUPLICATE' THEN 1 ELSE 0 END) AS total_duplicates,
                    SUM(CASE WHEN result_type = 'GAP' THEN 1 ELSE 0 END) AS total_gaps,
                    COUNT(DISTINCT branch_name) AS branches_affected
                FROM gold.caatt_duplicate_gap_results
            """))
            summary_row = summary.fetchone()

        return {"data": rows, "summary": dict(summary_row._mapping) if summary_row else {}}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/benford-analysis")
def get_benford_results():
    """CAATT #3: Benford's Law digit distribution analysis."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text("""
                SELECT * FROM gold.caatt_benford_results
                ORDER BY digit
            """))
            rows = [dict(r._mapping) for r in result]

            # Check if any digit has significant deviation
            significant = conn.execute(text("""
                SELECT COUNT(*) FROM gold.caatt_benford_results WHERE is_significant = TRUE
            """))
            sig_count = significant.scalar() or 0

        return {
            "data": rows,
            "summary": {
                "total_digits_analyzed": len(rows),
                "significant_deviations": sig_count,
                "conclusion": "ANOMALY DETECTED" if sig_count > 2 else "NORMAL" if sig_count == 0 else "MINOR DEVIATION",
            },
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/stratification")
def get_stratification_results(category: Optional[str] = None):
    """CAATT #4: Transaction stratification and aging analysis."""
    conditions = []
    params = {}

    if category:
        conditions.append("category = :category")
        params["category"] = category

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(text(f"""
                SELECT * FROM gold.caatt_stratification_results
                {where} ORDER BY min_value
            """), params)
            rows = [dict(r._mapping) for r in result]

            # Summary
            summary = conn.execute(text(f"""
                SELECT
                    SUM(trx_count) AS total_transactions,
                    SUM(total_amount) AS total_amount,
                    COUNT(DISTINCT stratum_label) AS total_strata,
                    COUNT(DISTINCT category) AS categories
                FROM gold.caatt_stratification_results {where}
            """), params)
            summary_row = summary.fetchone()

        return {"data": rows, "summary": dict(summary_row._mapping) if summary_row else {}}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/reconciliation")
def get_reconciliation_results():
    """CAATT #5: Cross-system reconciliation results (CBS vs GL vs LOS)."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text("""
                SELECT * FROM gold.caatt_reconciliation_results
                ORDER BY test_date DESC
            """))
            rows = [dict(r._mapping) for r in result]

            summary = conn.execute(text("""
                SELECT
                    AVG(match_rate_pct) AS avg_match_rate,
                    SUM(unmatched_a + unmatched_b) AS total_unmatched,
                    SUM(total_difference) AS total_difference
                FROM gold.caatt_reconciliation_results
            """))
            summary_row = summary.fetchone()

        return {"data": rows, "summary": dict(summary_row._mapping) if summary_row else {}}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/policy-violations")
def get_caatt_policy_violations(
    severity: Optional[str] = None,
    limit: int = Query(100, le=5000),
):
    """CAATT #10: Policy and rule engine violation results."""
    conditions = []
    params = {"limit": limit}

    if severity:
        conditions.append("severity = :severity")
        params["severity"] = severity

    where = "WHERE " + " AND ".join(conditions) if conditions else ""

    try:
        with engine.connect() as conn:
            result = conn.execute(text(f"""
                SELECT * FROM gold.caatt_policy_violations
                {where} ORDER BY test_date DESC LIMIT :limit
            """), params)
            rows = [dict(r._mapping) for r in result]

            summary = conn.execute(text("""
                SELECT
                    COUNT(*) AS total_violations,
                    SUM(CASE WHEN severity = 'Critical' THEN 1 ELSE 0 END) AS critical,
                    SUM(CASE WHEN severity = 'High' THEN 1 ELSE 0 END) AS high,
                    SUM(CASE WHEN severity = 'Medium' THEN 1 ELSE 0 END) AS medium,
                    SUM(CASE WHEN severity = 'Low' THEN 1 ELSE 0 END) AS low,
                    COUNT(DISTINCT rule_name) AS unique_rules_violated,
                    COUNT(DISTINCT branch_name) AS branches_affected
                FROM gold.caatt_policy_violations
            """))
            summary_row = summary.fetchone()

        return {"data": rows, "summary": dict(summary_row._mapping) if summary_row else {}}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/data-quality")
def get_data_quality_metrics():
    """CAATT #9: Data quality dashboard — completeness, accuracy, timeliness per table."""
    try:
        with engine.connect() as conn:
            result = conn.execute(text("""
                SELECT * FROM gold.caatt_data_quality_metrics
                ORDER BY test_date DESC
            """))
            rows = [dict(r._mapping) for r in result]

            # Overall quality score
            summary = conn.execute(text("""
                SELECT
                    ROUND(AVG(completeness_pct), 2) AS avg_completeness,
                    ROUND(AVG(accuracy_pct), 2) AS avg_accuracy,
                    ROUND(AVG(timeliness_days), 1) AS avg_timeliness_days,
                    COUNT(DISTINCT table_name) AS tables_profiled,
                    SUM(total_rows) AS total_rows_profiled
                FROM gold.caatt_data_quality_metrics
            """))
            summary_row = summary.fetchone()

            overall_score = 0.0
            if summary_row and summary_row[0] is not None:
                overall_score = round((float(summary_row[0]) + float(summary_row[1])) / 2, 2)

        return {
            "data": rows,
            "summary": dict(summary_row._mapping) if summary_row else {},
            "overall_quality_score": overall_score,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
