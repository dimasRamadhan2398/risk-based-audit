-- =========================================================================
-- Deduplication and Data Integrity Script
-- Database: rb_audit_audit_service
-- Cleans duplicates for audit_guidelines, audit_sops, and strategic_plans
-- Adds partial unique indexes to guarantee no duplicates ever recur
-- =========================================================================

BEGIN;

-- 1. GUIDELINES DEDUPLICATION
-- Map each guideline to its canonical record (preferring one referenced by active SOPs)
CREATE TEMP TABLE canonical_guidelines AS
WITH referenced AS (
    SELECT DISTINCT guideline_id FROM audit_sops WHERE deleted_at IS NULL
),
ranked AS (
    SELECT g.id, g.name,
           FIRST_VALUE(g.id) OVER (
               PARTITION BY TRIM(g.name) 
               ORDER BY 
                   CASE WHEN g.deleted_at IS NULL THEN 0 ELSE 1 END,
                   CASE WHEN r.guideline_id IS NOT NULL THEN 0 ELSE 1 END,
                   g.created_at ASC,
                   g.id ASC
           ) as canonical_id,
           ROW_NUMBER() OVER (
               PARTITION BY TRIM(g.name) 
               ORDER BY 
                   CASE WHEN g.deleted_at IS NULL THEN 0 ELSE 1 END,
                   CASE WHEN r.guideline_id IS NOT NULL THEN 0 ELSE 1 END,
                   g.created_at ASC,
                   g.id ASC
           ) as rn
    FROM audit_guidelines g
    LEFT JOIN referenced r ON g.id = r.guideline_id
)
SELECT id, canonical_id, name, rn FROM ranked;

-- Update SOPs referencing duplicate guidelines to point to canonical guideline
UPDATE audit_sops s
SET guideline_id = cg.canonical_id
FROM canonical_guidelines cg
WHERE s.guideline_id = cg.id AND cg.rn > 1;

-- Delete duplicate guidelines
DELETE FROM audit_guidelines
WHERE id IN (SELECT id FROM canonical_guidelines WHERE rn > 1);

-- Remove any soft-deleted guidelines that share a name with an active one
DELETE FROM audit_guidelines
WHERE deleted_at IS NOT NULL
  AND TRIM(name) IN (SELECT TRIM(name) FROM audit_guidelines WHERE deleted_at IS NULL);

-- 2. SOPS DEDUPLICATION
WITH ranked_sops AS (
    SELECT s.id,
           ROW_NUMBER() OVER (
               PARTITION BY s.guideline_id, TRIM(s.name)
               ORDER BY 
                   CASE WHEN s.deleted_at IS NULL THEN 0 ELSE 1 END,
                   s.created_at ASC,
                   s.id ASC
           ) as rn
    FROM audit_sops s
)
DELETE FROM audit_sops
WHERE id IN (SELECT id FROM ranked_sops WHERE rn > 1);

-- Remove soft-deleted SOPs that share a name with an active one
DELETE FROM audit_sops
WHERE deleted_at IS NOT NULL
  AND TRIM(name) IN (SELECT TRIM(name) FROM audit_sops WHERE deleted_at IS NULL);

-- 3. STRATEGIC PLANS CLEANUP
-- Remove soft-deleted duplicate records
DELETE FROM strategic_plans WHERE deleted_at IS NOT NULL;

-- Clean up duplicate SO-IA101 spam (keep only 1 latest)
WITH ranked_sp AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY TRIM(code)
               ORDER BY created_at DESC, id ASC
           ) as rn
    FROM strategic_plans
    WHERE code = 'SO-IA101'
)
DELETE FROM strategic_plans
WHERE id IN (SELECT id FROM ranked_sp WHERE rn > 1);

-- 4. PARTIAL UNIQUE INDEXES TO PERMANENTLY PREVENT DUPLICATES
CREATE UNIQUE INDEX IF NOT EXISTS idx_audit_guidelines_name_unique ON audit_guidelines (name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_audit_sops_name_unique ON audit_sops (guideline_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_strategic_plans_code_unique ON strategic_plans (code) WHERE deleted_at IS NULL;

COMMIT;
