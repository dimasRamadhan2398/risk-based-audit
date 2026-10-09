-- Database Indexes for Dashboard Caching Optimization
--
-- These indexes should be created to improve query performance for cached dashboard endpoints.
-- Run these migrations in a transaction and monitor performance before and after.
--
-- WARNING: Creating indexes on production databases should be done during low-traffic periods
-- with adequate monitoring. Consider using CONCURRENTLY option if your PostgreSQL version supports it.

-- ============================================================================
-- AUDIT SERVICE INDEXES
-- ============================================================================

-- Speed up recent findings queries ordered by creation date
-- Used by: GET /api/v1/audit-result-reports/recent-findings
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_result_reports_created_at_desc
ON audit_result_reports(created_at DESC);

-- Speed up assignment filtering and status checks
-- Used by: GET /api/v1/audit-assignments with filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_assignments_audit_plan_id_status
ON audit_assignments(audit_plan_id, status);

-- Speed up activity tracking and sorting
-- Used by: GET /api/v1/audit-activities with filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_activities_created_at_status
ON audit_activities(created_at DESC, status);

-- Speed up working paper sample queries
-- Used by: Fieldwork sample retrieval
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_working_paper_samples_assignment_id
ON working_paper_samples(assignment_letter_id);

-- Speed up audit charter lookups
-- Used by: GET /api/v1/audit-charters with active status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_charters_is_active
ON audit_charters(is_active);

-- Speed up audit mandate lookups
-- Used by: GET /api/v1/audit-mandates with active status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_mandates_is_active
ON audit_mandates(is_active);

-- ============================================================================
-- RISK SERVICE INDEXES
-- ============================================================================

-- Speed up risk filtering by status and risk level (dashboard KPIs)
-- Used by: Risk summary statistics, dashboard cards
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risks_status_risk_level
ON risks(status, risk_level);

-- Speed up RCM matrix lookups and control-risk relationships
-- Used by: GET /api/v1/rcm, /api/v1/rcm/summary
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rcm_matrix_control_id_risk_id
ON rcm_matrix(control_id, risk_id);

-- Speed up mitigation filtering by risk
-- Used by: Mitigation list queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mitigations_risk_id
ON mitigations(risk_id);

-- Speed up risk appetite queries
-- Used by: Risk appetite statements lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_appetite_status
ON risk_appetite(status);

-- Speed up risk factor lookups
-- Used by: GET /api/v1/risk-factors
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_factors_category
ON risk_factors(category);

-- ============================================================================
-- MASTER SERVICE INDEXES
-- ============================================================================

-- Speed up employee filtering by status and department
-- Used by: Employee list queries with filters
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_status_department_id
ON employees(status, department_id);

-- Speed up department filtering by company and status
-- Used by: Department list queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_departments_company_id_status
ON departments(company_id, status);

-- Speed up organizational hierarchy queries
-- Used by: Business unit lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_business_units_company_id
ON business_units(company_id);

-- Speed up job role queries
-- Used by: Job role list and search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_roles_status
ON job_roles(status);

-- Speed up location queries
-- Used by: Location list and filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_locations_status
ON locations(status);

-- ============================================================================
-- COMPOSITE INDEXES FOR COMPLEX QUERIES
-- ============================================================================

-- Speed up filtered risk queries (commonly used in dashboards)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risks_status_level_updated
ON risks(status, risk_level, updated_at DESC)
WHERE status != 'archived';

-- Speed up assignment filtering with auditor and plan
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_assignments_auditor_plan_status
ON audit_assignments(auditor_id, audit_plan_id, status);

-- Speed up finding queries by type and status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_findings_type_status_created
ON findings(finding_type, status, created_at DESC)
WHERE status IN ('open', 'in_progress');

-- ============================================================================
-- ANALYSIS QUERIES - For dashboard performance
-- ============================================================================

-- Check if index exists and show its size
-- Run these queries AFTER creating indexes to verify:

-- Show size of each index
-- SELECT
--     schemaname,
--     tablename,
--     indexname,
--     pg_size_pretty(pg_relation_size(indexrelid)) as index_size
-- FROM pg_stat_user_indexes
-- ORDER BY pg_relation_size(indexrelid) DESC;

-- Check index usage
-- SELECT
--     schemaname,
--     tablename,
--     indexrelname,
--     idx_scan,
--     idx_tup_read,
--     idx_tup_fetch
-- FROM pg_stat_user_indexes
-- WHERE idx_scan = 0  -- Unused indexes
-- ORDER BY pg_relation_size(indexrelid) DESC;

-- Check slow queries
-- SELECT
--     query,
--     calls,
--     total_time,
--     mean_time
-- FROM pg_stat_statements
-- WHERE mean_time > 100  -- Queries taking more than 100ms
-- ORDER BY mean_time DESC;

-- ============================================================================
-- MAINTENANCE QUERIES
-- ============================================================================

-- Analyze tables after creating indexes (for query planner)
-- ANALYZE audit_result_reports;
-- ANALYZE risks;
-- ANALYZE rcm_matrix;
-- ANALYZE employees;
-- ANALYZE departments;

-- Reindex all indexes on a table (if performance degrades)
-- REINDEX TABLE CONCURRENTLY audit_result_reports;
-- REINDEX TABLE CONCURRENTLY risks;

-- Drop unused indexes
-- DROP INDEX idx_unused_index_name;

-- ============================================================================
-- MONITORING QUERIES
-- ============================================================================

-- Monitor cache effectiveness after implementing HTTP caching
-- Track: Are fewer database queries being made?

-- Count queries by table before and after caching:
-- SELECT
--     schemaname,
--     tablename,
--     seq_scan,
--     seq_tup_read,
--     idx_scan,
--     idx_tup_fetch
-- FROM pg_stat_user_tables
-- WHERE schemaname = 'public'
-- ORDER BY seq_scan DESC;

-- ============================================================================
-- NOTES
-- ============================================================================

-- 1. Create indexes during low-traffic periods
-- 2. Use CONCURRENTLY to avoid locking tables (PostgreSQL 9.2+)
-- 3. Run ANALYZE after creating indexes
-- 4. Monitor slow_query_log before and after index creation
-- 5. Review execution plans with EXPLAIN ANALYZE
-- 6. For production, consider using pg_partman for large tables

-- Example: Before and after performance testing
--
-- BEFORE INDEX:
-- =============
-- Seq Scan on audit_result_reports  (cost=0.00..35000.00 rows=50000)
-- Planning Time: 0.100 ms
-- Execution Time: 250.000 ms
--
-- AFTER INDEX:
-- =============
-- Index Scan using idx_audit_result_reports_created_at_desc  (cost=0.42..100.00 rows=50000)
-- Planning Time: 0.050 ms
-- Execution Time: 10.000 ms
--
-- Result: 25x faster query execution (250ms → 10ms)
