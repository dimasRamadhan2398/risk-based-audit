"""
AuditSphere Data Hub — Identifier Sanitizer & Quoter
Protects against SQL injection when handling dynamic external database tables and columns.
"""
import re

IDENTIFIER_REGEX = re.compile(r"^[a-zA-Z_][a-zA-Z0-9_]*$")

def validate_identifier(name: str) -> str:
    """Validate that identifier consists only of alphanumeric characters and underscores."""
    if not name or not IDENTIFIER_REGEX.match(name):
        raise ValueError(f"Invalid or unsafe SQL identifier: '{name}'")
    return name

def quote_identifier(name: str) -> str:
    """Safely quote a SQL identifier using double quotes."""
    validate_identifier(name)
    # Double quote with internal quote escaping
    escaped = name.replace('"', '""')
    return f'"{escaped}"'

def quote_table_name(table_name: str, schema: str = None) -> str:
    """Safely quote a table name with optional schema."""
    if schema:
        return f"{quote_identifier(schema)}.{quote_identifier(table_name)}"
    return quote_identifier(table_name)

def sanitize_table_name(name: str) -> str:
    """Sanitize string into safe snake_case table name component."""
    cleaned = re.sub(r"[^a-zA-Z0-9_]+", "_", name.lower().strip())
    return cleaned.strip("_") or "default"
