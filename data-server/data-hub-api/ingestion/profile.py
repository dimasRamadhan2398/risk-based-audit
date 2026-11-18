"""
AuditSphere Data Hub — Read-Only Table Profiler & Stratified Sampler
Computes exact row counts, real metric aggregates, dimension breakdowns,
and stratified sample rows (both anomalous/flagged and normal baseline rows)
directly from the connected client database without fake fallbacks.
"""
from typing import Any, Dict, List, Optional
from core.identifiers import quote_identifier, validate_identifier

DIMENSION_CANDIDATES = [
    "branch_name",
    "entity_unit",
    "department",
    "reporting_unit",
    "vendor_name",
    "gl_account_name",
    "region",
    "branch_code",
    "product_name",
    "category",
]

AMOUNT_CANDIDATES = [
    "amount",
    "plafond",
    "outstanding",
    "contract_value",
    "invoice_amount",
    "exposure_value",
    "debit_amount",
    "principal_amount",
    "balance",
]

MEASURE_CANDIDATES = set(
    AMOUNT_CANDIDATES
    + [
        "credit_amount",
        "severity_score",
        "collectability",
        "days_past_due",
        "interest_rate",
        "tenor_months",
        "is_suspicious",
        "is_flagged",
        "is_override_auth",
        "voucher_number",
    ]
)


def _pick_first_column(col_map: Dict[str, str], candidates: List[str]) -> Optional[str]:
    for cand in candidates:
        if cand in col_map:
            return col_map[cand]
    return None


def _build_flag_sql(col_map: Dict[str, str]) -> str:
    clauses: List[str] = []
    if "is_suspicious" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['is_suspicious'])}, FALSE) IS TRUE")
    if "is_flagged" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['is_flagged'])}, FALSE) IS TRUE")
    if "is_override_auth" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['is_override_auth'])}, FALSE) IS TRUE")
    if "authorization_status" in col_map:
        q = quote_identifier(col_map["authorization_status"])
        clauses.append(f"({q} IS NOT NULL AND {q} NOT IN ('AUTHORIZED', 'APPROVED', 'APPROVED_DUAL'))")
    if "approval_status" in col_map:
        q = quote_identifier(col_map["approval_status"])
        clauses.append(f"({q} IS NOT NULL AND {q} NOT IN ('APPROVED'))")
    if "collectability" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['collectability'])}, 1) >= 3")
    if "days_past_due" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['days_past_due'])}, 0) > 90")
    if "interest_rate" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['interest_rate'])}, 0) > 24")
    if "tenor_months" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['tenor_months'])}, 0) > 120")
    if "plafond" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['plafond'])}, 0) > 5000000000")
    if "voucher_number" in col_map:
        q = quote_identifier(col_map["voucher_number"])
        clauses.append(f"({q} IS NULL OR TRIM(CAST({q} AS TEXT)) = '')")
    if "severity_score" in col_map:
        clauses.append(f"COALESCE({quote_identifier(col_map['severity_score'])}, 0) >= 70")
    if "status" in col_map:
        q = quote_identifier(col_map["status"])
        clauses.append(f"({q} IN ('ESCALATED', 'NPL_MACET', 'FLAGGED', 'NON_COMPLIANT'))")
    return " OR ".join(clauses) if clauses else "FALSE"


def profile_client_table(conn, table_name: str, sample_limit: int = 50) -> Dict[str, Any]:
    """
    Execute read-only SQL profiling and stratified sampling on `table_name`.
    Returns:
      - tableName: str
      - total_rows: int
      - rows: List[Dict[str, Any]]
      - profile: Dict[str, Any]
    """
    validate_identifier(table_name)
    quoted_table = quote_identifier(table_name)
    sample_limit = max(1, min(int(sample_limit or 50), 200))

    with conn.cursor() as cur:
        cur.execute(f"SELECT * FROM {quoted_table} LIMIT 1;")
        col_names = [desc[0] for desc in (cur.description or [])]

    col_map: Dict[str, str] = {}
    for c in col_names:
        validate_identifier(c)
        col_map[c.lower()] = c

    active_where = "TRUE"
    if "is_deleted" in col_map:
        active_where = f"COALESCE({quote_identifier(col_map['is_deleted'])}, FALSE) IS FALSE"

    dim_col = _pick_first_column(col_map, DIMENSION_CANDIDATES)
    amt_col = _pick_first_column(col_map, AMOUNT_CANDIDATES)
    has_numeric_measure = any(c in col_map for c in MEASURE_CANDIDATES)
    flag_sql = _build_flag_sql(col_map)

    total_rows = 0
    flagged_rows = 0
    total_amount = 0.0
    avg_amount = 0.0
    max_amount = 0.0
    metrics: Dict[str, Any] = {}

    with conn.cursor() as cur:
        amt_expr = f"COALESCE(CAST({quote_identifier(amt_col)} AS DOUBLE PRECISION), 0)" if amt_col else "0.0"
        cur.execute(
            f"""
            SELECT
                COUNT(*) AS total_rows,
                COUNT(*) FILTER (WHERE {flag_sql}) AS flagged_rows,
                COALESCE(SUM({amt_expr}), 0.0) AS total_amount,
                COALESCE(AVG({amt_expr}), 0.0) AS avg_amount,
                COALESCE(MAX({amt_expr}), 0.0) AS max_amount
            FROM {quoted_table}
            WHERE {active_where};
            """
        )
        agg = cur.fetchone()
        if agg:
            total_rows = int(agg[0] or 0)
            flagged_rows = int(agg[1] or 0)
            total_amount = float(agg[2] or 0.0)
            avg_amount = float(agg[3] or 0.0)
            max_amount = float(agg[4] or 0.0)

        # Detailed per-domain metric counters
        metric_selects: List[str] = []
        metric_keys: List[str] = []
        if "is_suspicious" in col_map:
            metric_keys.append("suspicious_count")
            metric_selects.append(
                f"COUNT(*) FILTER (WHERE COALESCE({quote_identifier(col_map['is_suspicious'])}, FALSE) IS TRUE)"
            )
        if "authorization_status" in col_map:
            q = quote_identifier(col_map["authorization_status"])
            metric_keys.append("unauthorized_count")
            metric_selects.append(
                f"COUNT(*) FILTER (WHERE {q} IS NOT NULL AND {q} NOT IN ('AUTHORIZED', 'APPROVED', 'APPROVED_DUAL'))"
            )
        if "collectability" in col_map or "days_past_due" in col_map:
            npl_parts = []
            if "collectability" in col_map:
                npl_parts.append(f"COALESCE({quote_identifier(col_map['collectability'])}, 1) >= 3")
            if "days_past_due" in col_map:
                npl_parts.append(f"COALESCE({quote_identifier(col_map['days_past_due'])}, 0) > 90")
            metric_keys.append("npl_count")
            metric_selects.append(f"COUNT(*) FILTER (WHERE {' OR '.join(npl_parts)})")
        if any(k in col_map for k in ("plafond", "interest_rate", "tenor_months")):
            breach_parts = []
            if "plafond" in col_map:
                breach_parts.append(f"COALESCE({quote_identifier(col_map['plafond'])}, 0) > 5000000000")
            if "interest_rate" in col_map:
                breach_parts.append(f"COALESCE({quote_identifier(col_map['interest_rate'])}, 0) > 24")
            if "tenor_months" in col_map:
                breach_parts.append(f"COALESCE({quote_identifier(col_map['tenor_months'])}, 0) > 120")
            metric_keys.append("policy_breach_count")
            metric_selects.append(f"COUNT(*) FILTER (WHERE {' OR '.join(breach_parts)})")
        if "voucher_number" in col_map:
            q = quote_identifier(col_map["voucher_number"])
            metric_keys.append("missing_voucher_count")
            metric_selects.append(f"COUNT(*) FILTER (WHERE {q} IS NULL OR TRIM(CAST({q} AS TEXT)) = '')")
        if "is_flagged" in col_map or "approval_status" in col_map:
            flg_parts = []
            if "is_flagged" in col_map:
                flg_parts.append(f"COALESCE({quote_identifier(col_map['is_flagged'])}, FALSE) IS TRUE")
            if "approval_status" in col_map:
                q = quote_identifier(col_map["approval_status"])
                flg_parts.append(f"({q} IS NOT NULL AND {q} NOT IN ('APPROVED'))")
            metric_keys.append("flagged_contract_count")
            metric_selects.append(f"COUNT(*) FILTER (WHERE {' OR '.join(flg_parts)})")
        if "severity_score" in col_map:
            metric_keys.append("high_severity_count")
            metric_selects.append(
                f"COUNT(*) FILTER (WHERE COALESCE({quote_identifier(col_map['severity_score'])}, 0) >= 70)"
            )

        if metric_selects:
            cur.execute(
                f"SELECT {', '.join(metric_selects)} FROM {quoted_table} WHERE {active_where};"
            )
            mrow = cur.fetchone()
            if mrow:
                for idx, key in enumerate(metric_keys):
                    metrics[key] = int(mrow[idx] or 0)

    dimension_breakdown: List[Dict[str, Any]] = []
    if dim_col:
        q_dim = quote_identifier(dim_col)
        amt_expr = f"COALESCE(CAST({quote_identifier(amt_col)} AS DOUBLE PRECISION), 0)" if amt_col else "0.0"
        with conn.cursor() as cur:
            cur.execute(
                f"""
                SELECT
                    CAST({q_dim} AS TEXT) AS dim_name,
                    COUNT(*) AS row_count,
                    COUNT(*) FILTER (WHERE {flag_sql}) AS flagged_count,
                    COALESCE(SUM({amt_expr}), 0.0) AS total_amount,
                    COALESCE(MAX({amt_expr}), 0.0) AS max_amount
                FROM {quoted_table}
                WHERE {active_where} AND {q_dim} IS NOT NULL AND TRIM(CAST({q_dim} AS TEXT)) <> ''
                GROUP BY {q_dim}
                ORDER BY flagged_count DESC, row_count DESC, dim_name ASC
                LIMIT 20;
                """
            )
            for drow in cur.fetchall():
                dimension_breakdown.append(
                    {
                        "name": str(drow[0]),
                        "row_count": int(drow[1] or 0),
                        "flagged_count": int(drow[2] or 0),
                        "total_amount": float(drow[3] or 0.0),
                        "max_amount": float(drow[4] or 0.0),
                    }
                )

    # Stratified sampling: include both anomalous/flagged rows and normal baseline rows
    rows: List[Dict[str, Any]] = []
    order_col = amt_col or ("severity_score" if "severity_score" in col_map else None)
    order_clause = f"ORDER BY {quote_identifier(order_col)} DESC NULLS LAST" if order_col else ""

    with conn.cursor() as cur:
        if flag_sql != "FALSE" and flagged_rows > 0:
            flagged_quota = min(max(10, sample_limit // 2), sample_limit)
            cur.execute(
                f"""
                SELECT * FROM {quoted_table}
                WHERE ({active_where}) AND ({flag_sql})
                {order_clause}
                LIMIT %s;
                """,
                (flagged_quota,),
            )
            f_cols = [desc[0] for desc in (cur.description or [])]
            for r in cur.fetchall():
                rows.append(dict(zip(f_cols, r)))

            normal_quota = max(0, sample_limit - len(rows))
            if normal_quota > 0:
                cur.execute(
                    f"""
                    SELECT * FROM {quoted_table}
                    WHERE ({active_where}) AND NOT ({flag_sql})
                    LIMIT %s;
                    """,
                    (normal_quota,),
                )
                n_cols = [desc[0] for desc in (cur.description or [])]
                for r in cur.fetchall():
                    rows.append(dict(zip(n_cols, r)))
        else:
            cur.execute(
                f"SELECT * FROM {quoted_table} WHERE {active_where} {order_clause} LIMIT %s;",
                (sample_limit,),
            )
            s_cols = [desc[0] for desc in (cur.description or [])]
            for r in cur.fetchall():
                rows.append(dict(zip(s_cols, r)))

    profile = {
        "table_name": table_name,
        "total_rows": total_rows,
        "flagged_rows": flagged_rows,
        "has_numeric_measure": has_numeric_measure,
        "dimension_column": dim_col,
        "amount_column": amt_col,
        "columns": col_names,
        "total_amount": total_amount,
        "avg_amount": avg_amount,
        "max_amount": max_amount,
        "metrics": metrics,
        "dimension_breakdown": dimension_breakdown,
    }

    return {
        "tableName": table_name,
        "total_rows": total_rows,
        "rows": rows,
        "profile": profile,
    }
