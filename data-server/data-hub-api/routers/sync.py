"""
Data Hub API — Sync Router
Endpoints for bidirectional master data sync with AuditSphere Audit Server.
"""
from typing import List, Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from sqlalchemy import text
import uuid

from main import engine

router = APIRouter()


# ─── Request Models ──────────────────────────────────────────────────────────
class CompanySync(BaseModel):
    id: Optional[str] = None
    code: str
    name: str
    legal_name: Optional[str] = None
    company_type: Optional[str] = "BRANCH"
    is_active: bool = True


class DepartmentSync(BaseModel):
    id: Optional[str] = None
    department_code: str
    department_name: str
    business_unit_id: Optional[str] = None
    is_active: bool = True


class EmployeeSync(BaseModel):
    id: Optional[str] = None
    employee_id_number: str
    full_name: str
    email: Optional[str] = None
    department_id: Optional[str] = None
    is_active: bool = True


class RiskCategorySync(BaseModel):
    id: Optional[str] = None
    name: str
    description: Optional[str] = None


class RiskRegisterSync(BaseModel):
    id: Optional[str] = None
    code: str
    name: str
    department_id: Optional[str] = None
    risk_category_id: Optional[str] = None
    inherent_score: int = 0
    residual_score: int = 0
    is_active: bool = True


class AuditFindingSync(BaseModel):
    id: Optional[str] = None
    finding_code: str
    title: str
    description: Optional[str] = None
    severity: Optional[str] = "Medium"
    status: Optional[str] = "Open"
    department_id: Optional[str] = None


class MasterDataSyncRequest(BaseModel):
    companies: Optional[List[CompanySync]] = None
    departments: Optional[List[DepartmentSync]] = None
    employees: Optional[List[EmployeeSync]] = None
    risk_categories: Optional[List[RiskCategorySync]] = None
    risk_registers: Optional[List[RiskRegisterSync]] = None
    audit_findings: Optional[List[AuditFindingSync]] = None


# ─── Sync Endpoint ───────────────────────────────────────────────────────────
@router.post("/master-data")
def sync_master_data(req: MasterDataSyncRequest):
    """
    Receive master data from AuditSphere Audit Server.
    Inserts into Bronze zone sync tables, then triggers Silver zone reference update.
    """
    counts = {}

    try:
        with engine.connect() as conn:
            # Companies
            if req.companies:
                for c in req.companies:
                    cid = c.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_companies (id, company_code, company_name, legal_name, company_type, is_active)
                        VALUES (:id, :code, :name, :legal, :ctype, :active)
                    """), {"id": cid, "code": c.code, "name": c.name, "legal": c.legal_name, "ctype": c.company_type, "active": c.is_active})
                counts["companies"] = len(req.companies)

                # Upsert to Silver reference
                for c in req.companies:
                    cid = c.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO silver.ref_companies (id, company_code, company_name, company_type, is_active, synced_at)
                        VALUES (:id, :code, :name, :ctype, :active, NOW())
                        ON CONFLICT (id) DO UPDATE SET
                            company_code = EXCLUDED.company_code,
                            company_name = EXCLUDED.company_name,
                            company_type = EXCLUDED.company_type,
                            is_active = EXCLUDED.is_active,
                            synced_at = NOW()
                    """), {"id": cid, "code": c.code, "name": c.name, "ctype": c.company_type, "active": c.is_active})

            # Departments
            if req.departments:
                for d in req.departments:
                    did = d.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_departments (id, department_code, department_name, business_unit_id, is_active)
                        VALUES (:id, :code, :name, :buid, :active)
                    """), {"id": did, "code": d.department_code, "name": d.department_name, "buid": d.business_unit_id, "active": d.is_active})
                counts["departments"] = len(req.departments)

                for d in req.departments:
                    did = d.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO silver.ref_departments (id, department_code, department_name, business_unit_id, is_active, synced_at)
                        VALUES (:id, :code, :name, :buid, :active, NOW())
                        ON CONFLICT (id) DO UPDATE SET
                            department_code = EXCLUDED.department_code,
                            department_name = EXCLUDED.department_name,
                            business_unit_id = EXCLUDED.business_unit_id,
                            is_active = EXCLUDED.is_active,
                            synced_at = NOW()
                    """), {"id": did, "code": d.department_code, "name": d.department_name, "buid": d.business_unit_id, "active": d.is_active})

            # Employees
            if req.employees:
                for e in req.employees:
                    eid = e.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_employees (id, employee_id_number, full_name, email, department_id, is_active)
                        VALUES (:id, :eid_num, :name, :email, :dept, :active)
                    """), {"id": eid, "eid_num": e.employee_id_number, "name": e.full_name, "email": e.email, "dept": e.department_id, "active": e.is_active})
                counts["employees"] = len(req.employees)

                for e in req.employees:
                    eid = e.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO silver.ref_employees (id, employee_id_number, full_name, email, department_id, is_active, synced_at)
                        VALUES (:id, :eid_num, :name, :email, :dept, :active, NOW())
                        ON CONFLICT (id) DO UPDATE SET
                            employee_id_number = EXCLUDED.employee_id_number,
                            full_name = EXCLUDED.full_name,
                            email = EXCLUDED.email,
                            department_id = EXCLUDED.department_id,
                            is_active = EXCLUDED.is_active,
                            synced_at = NOW()
                    """), {"id": eid, "eid_num": e.employee_id_number, "name": e.full_name, "email": e.email, "dept": e.department_id, "active": e.is_active})

            # Risk Categories
            if req.risk_categories:
                for rc in req.risk_categories:
                    rcid = rc.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_risk_categories (id, name, description)
                        VALUES (:id, :name, :desc)
                    """), {"id": rcid, "name": rc.name, "desc": rc.description})
                counts["risk_categories"] = len(req.risk_categories)

                for rc in req.risk_categories:
                    rcid = rc.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO silver.ref_risk_categories (id, name, description, synced_at)
                        VALUES (:id, :name, :desc, NOW())
                        ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, synced_at = NOW()
                    """), {"id": rcid, "name": rc.name, "desc": rc.description})

            # Risk Registers
            if req.risk_registers:
                for rr in req.risk_registers:
                    rrid = rr.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_risk_registers (id, code, name, department_id, risk_category_id, inherent_score, residual_score, is_active)
                        VALUES (:id, :code, :name, :dept, :rcat, :inh, :res, :active)
                    """), {"id": rrid, "code": rr.code, "name": rr.name, "dept": rr.department_id, "rcat": rr.risk_category_id, "inh": rr.inherent_score, "res": rr.residual_score, "active": rr.is_active})
                counts["risk_registers"] = len(req.risk_registers)

            # Audit Findings
            if req.audit_findings:
                for af in req.audit_findings:
                    afid = af.id or str(uuid.uuid4())
                    conn.execute(text("""
                        INSERT INTO bronze.sync_audit_findings (id, finding_code, title, description, severity, status, department_id)
                        VALUES (:id, :code, :title, :desc, :sev, :status, :dept)
                    """), {"id": afid, "code": af.finding_code, "title": af.title, "desc": af.description, "sev": af.severity, "status": af.status, "dept": af.department_id})
                counts["audit_findings"] = len(req.audit_findings)

            conn.commit()

        return {
            "status": "success",
            "message": "Master data synced to Bronze and Silver zones",
            "synced_counts": counts,
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Sync failed: {str(e)}")


@router.get("/master-data/status")
def get_sync_status():
    """Check last sync status for each master data type."""
    try:
        with engine.connect() as conn:
            tables = ["sync_companies", "sync_departments", "sync_employees",
                       "sync_risk_categories", "sync_risk_registers", "sync_audit_findings"]
            status = {}
            for t in tables:
                r = conn.execute(text(f"SELECT COUNT(*) AS cnt, MAX(_loaded_at) AS last_sync FROM bronze.{t}"))
                row = r.fetchone()
                status[t.replace("sync_", "")] = {
                    "total_records": row[0] if row else 0,
                    "last_sync": str(row[1]) if row and row[1] else None,
                }
        return {"status": "success", "sync_status": status}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
