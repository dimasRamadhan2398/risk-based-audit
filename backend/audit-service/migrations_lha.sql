-- Migration to update legacy LHA report numbers to new format: LHA-[NoUrut]/[Audit Team]/[Tahun]

-- 1. Update audit_result_reports.report_number
UPDATE audit_result_reports 
SET report_number = 'LHA-' || SPLIT_PART(report_number, '/', 1) || '/SKAI/' || SPLIT_PART(report_number, '/', 5) 
WHERE report_number LIKE '%/LHA/%';

-- 2. Update audit_result_reports.executive_summary
UPDATE audit_result_reports
SET executive_summary = REGEXP_REPLACE(executive_summary, '([0-9]{3})/LHA/[0-9]{2}/[^/]+/([0-9]{4})', 'LHA-\1/SKAI/\2', 'g')
WHERE executive_summary LIKE '%/LHA/%';

-- 3. Update executive_summaries.nomor_dokumen
UPDATE executive_summaries 
SET nomor_dokumen = 'LHA-' || SPLIT_PART(nomor_dokumen, '/', 1) || '/SKAI/' || SPLIT_PART(nomor_dokumen, '/', 5) 
WHERE nomor_dokumen LIKE '%/LHA/%';

-- 4. Update executive_summaries.narrative
UPDATE executive_summaries
SET narrative = REGEXP_REPLACE(narrative, '([0-9]{3})/LHA/[0-9]{2}/[^/]+/([0-9]{4})', 'LHA-\1/SKAI/\2', 'g')
WHERE narrative LIKE '%/LHA/%';

-- 5. Update executive_summaries.matriks_kompilasi
UPDATE executive_summaries
SET matriks_kompilasi = REGEXP_REPLACE(matriks_kompilasi, '([0-9]{3})/LHA/[0-9]{2}/[^/]+/([0-9]{4})', 'LHA-\1/SKAI/\2', 'g')
WHERE matriks_kompilasi LIKE '%/LHA/%';

UPDATE executive_summaries
SET matriks_kompilasi = REGEXP_REPLACE(matriks_kompilasi, '"nomor":"([0-9]{3})/LHA/[0-9]{2}"', '"nomor":"LHA-\1"', 'g')
WHERE matriks_kompilasi LIKE '%/LHA/%';
