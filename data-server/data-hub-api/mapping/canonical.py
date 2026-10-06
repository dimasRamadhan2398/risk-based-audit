"""
Data Hub API — Canonical Entity Mapping & SQL Generator
Matches arbitrary client columns to AuditSphere standard entities
and generates set-based SQL to feed canonical Silver and Gold models.
"""
from typing import List, Dict, Any, Optional
from mapping.synonyms import CANONICAL_SCHEMAS
from core.identifiers import quote_identifier, sanitize_table_name


def suggest_canonical_mapping(columns: List[str], table_name: str = "") -> dict:
    """
    Suggest best canonical entity and column mappings based on synonym matching.
    Returns matched entity, confidence score, and column-to-column pairs.
    """
    clean_cols = [c.lower().strip() for c in columns if not c.startswith("_")]
    table_lower = table_name.lower().strip()

    best_entity = None
    best_score = -1.0
    best_mapping = {}

    for entity_name, entity_def in CANONICAL_SCHEMAS.items():
        score = 0
        mapping = {}
        fields = entity_def["fields"]
        req_fields = entity_def["required_fields"]

        # Boost score if table name hints at entity
        if entity_name in table_lower or any(syn in table_lower for syn in [entity_name[:4], entity_name[:-1]]):
            score += 2.0

        for canon_field, syn_list in fields.items():
            matched_col = None
            for col in clean_cols:
                if col == canon_field:
                    matched_col = col
                    score += 2.0
                    break
                elif col in syn_list:
                    matched_col = col
                    score += 1.5
                    break
                elif any(syn in col for syn in syn_list if len(syn) > 3):
                    matched_col = col
                    score += 0.8
                    break

            if matched_col:
                mapping[canon_field] = matched_col

        # Penalty if required fields are missing
        req_matched = sum(1 for rf in req_fields if rf in mapping)
        if len(req_fields) > 0 and req_matched == 0:
            score = 0

        confidence = round(min(1.0, score / max(4.0, len(fields))), 2)

        if confidence > best_score and req_matched >= (len(req_fields) // 2):
            best_score = confidence
            best_entity = entity_name
            best_mapping = mapping

    # If no high confidence canonical match, default to generic audit_features
    if not best_entity or best_score < 0.3:
        return {
            "entity": "generic",
            "confidence": 0.0,
            "description": "Generic external dataset (will be stored in silver.external_source_data)",
            "mappings": {},
            "unmapped_columns": clean_cols
        }

    mapped_src_cols = set(best_mapping.values())
    unmapped = [c for c in clean_cols if c not in mapped_src_cols]

    return {
        "entity": best_entity,
        "confidence": best_score,
        "description": CANONICAL_SCHEMAS[best_entity]["description"],
        "required_fields": CANONICAL_SCHEMAS[best_entity]["required_fields"],
        "mappings": best_mapping,
        "unmapped_columns": unmapped
    }


def validate_canonical_mapping(entity: str, mapping_dict: Dict[str, str]) -> Dict[str, Any]:
    """Validate that required fields for the canonical entity are present in the mapping."""
    if entity not in CANONICAL_SCHEMAS:
        return {"valid": False, "error": f"Unknown canonical entity: {entity}"}

    schema = CANONICAL_SCHEMAS[entity]
    missing_required = []
    for req in schema["required_fields"]:
        if req not in mapping_dict or not mapping_dict[req]:
            missing_required.append(req)

    if missing_required:
        return {
            "valid": False,
            "error": f"Missing required canonical fields: {', '.join(missing_required)}",
            "missing_fields": missing_required
        }

    return {"valid": True, "message": "Mapping is valid for canonical ingestion"}


def generate_canonical_sql(
    source_id: str,
    source_name: str,
    table_name: str,
    entity: str,
    mapping_dict: Dict[str, str],
    batch_id: str = None
) -> Optional[str]:
    """
    Generate set-based SQL to load external Bronze table directly into canonical Silver table.
    """
    safe_sname = sanitize_table_name(source_name)
    safe_tname = sanitize_table_name(table_name)
    bronze_tbl = f"src_{safe_sname}_{safe_tname}"
    batch_filter = f"AND b._batch_id = '{batch_id}'" if batch_id else ""

    if entity == "transactions":
        rf = mapping_dict.get("trx_ref_number")
        dt = mapping_dict.get("trx_date")
        amt = mapping_dict.get("amount")
        if not (rf and dt and amt):
            return None

        def col_or_literal(key: str, default_sql: str) -> str:
            val = mapping_dict.get(key)
            if val and not val.startswith("'"):
                return f"COALESCE(b.{quote_identifier(val)}::TEXT, {default_sql})"
            return default_sql

        chn_expr = col_or_literal("channel", "'SYSTEM'")
        cat_expr = col_or_literal("category", "'GENERAL'")
        acc_expr = col_or_literal("account_id", "1")
        br_expr = col_or_literal("branch_id", "1")
        desc_expr = col_or_literal("description", "''")

        return f"""
            INSERT INTO silver.trx_cleaned (
                source_id, trx_ref_number, trx_date, trx_hour, trx_day_of_week, trx_type,
                category, channel, amount, amount_millions, description, account_id,
                branch_id, authorization_status, is_suspicious, is_round_amount,
                is_after_hours, _row_hash, _batch_id, is_deleted, deleted_detected_at, cleaned_at
            )
            SELECT DISTINCT ON ('{source_id}', b.{quote_identifier(rf)}::TEXT)
                '{source_id}',
                b.{quote_identifier(rf)}::TEXT,
                b.{quote_identifier(dt)}::DATE,
                12,
                EXTRACT(DOW FROM b.{quote_identifier(dt)}::DATE)::INT,
                {cat_expr},
                {cat_expr},
                {chn_expr},
                b.{quote_identifier(amt)}::NUMERIC,
                ROUND((b.{quote_identifier(amt)}::NUMERIC) / 1000000.0, 4),
                {desc_expr},
                ({acc_expr})::INT,
                ({br_expr})::INT,
                'APPROVED',
                FALSE,
                CASE WHEN b.{quote_identifier(amt)}::NUMERIC > 0 AND (b.{quote_identifier(amt)}::BIGINT % 100000000 = 0) THEN TRUE ELSE FALSE END,
                FALSE,
                COALESCE(b._row_hash, MD5(ROW(b.*)::TEXT)),
                b._batch_id,
                CASE WHEN b._op = 'D' THEN TRUE ELSE FALSE END,
                CASE WHEN b._op = 'D' THEN NOW() ELSE NULL END,
                NOW()
            FROM bronze.{quote_identifier(bronze_tbl)} b
            WHERE b.{quote_identifier(rf)} IS NOT NULL {batch_filter}
            ON CONFLICT (source_id, trx_ref_number)
            DO UPDATE SET
                trx_date = EXCLUDED.trx_date,
                amount = EXCLUDED.amount,
                amount_millions = EXCLUDED.amount_millions,
                description = EXCLUDED.description,
                _row_hash = EXCLUDED._row_hash,
                is_deleted = EXCLUDED.is_deleted,
                cleaned_at = NOW()
            WHERE silver.trx_cleaned._row_hash IS DISTINCT FROM EXCLUDED._row_hash;
        """

    return None
