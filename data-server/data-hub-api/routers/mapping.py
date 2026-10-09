"""
Data Hub API — Canonical Mapping Router
Endpoints for discovering canonical schemas, auto-suggesting mappings, and validating configurations.
"""
from typing import List, Dict, Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from mapping.synonyms import CANONICAL_SCHEMAS
from mapping.canonical import suggest_canonical_mapping, validate_canonical_mapping

router = APIRouter()


class SuggestRequest(BaseModel):
    columns: List[str]
    table_name: Optional[str] = ""


class ValidateRequest(BaseModel):
    entity: str
    mapping: Dict[str, str]


@router.get("/schemas")
def list_canonical_schemas():
    """List all supported canonical entity schemas and their required fields."""
    return {
        "status": "success",
        "entities": CANONICAL_SCHEMAS
    }


@router.post("/suggest")
def get_suggested_mapping(req: SuggestRequest):
    """
    Given an array of column names and table name from a client database,
    suggest the best matching canonical entity and column pairs.
    """
    if not req.columns:
        raise HTTPException(status_code=400, detail="Columns array cannot be empty")
    suggestion = suggest_canonical_mapping(req.columns, req.table_name or "")
    return {"status": "success", "suggestion": suggestion}


@router.post("/validate")
def validate_mapping(req: ValidateRequest):
    """Validate user-configured canonical mapping."""
    result = validate_canonical_mapping(req.entity, req.mapping)
    return {"status": "success" if result["valid"] else "error", **result}
