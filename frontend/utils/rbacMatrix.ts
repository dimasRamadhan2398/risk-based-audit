/**
 * AuditSphere RBAC Permission Matrix & Utility Helpers.
 * Matches the official RBAC matrix displayed in Account Settings -> Permissions.
 */

export interface MatrixItem {
  id: string
  num: number
  module: string
  submodule: string
  isSubmodule: boolean
  admin: string
  executive: string
  auditor: string
  auditee: string
  departmentHead: string
  viewer: string
}

export interface UserPermissionItem {
  id: string
  num: number
  module: string
  submodule?: string
  displayName: string
  accessLevel: string
  badgeClass: string
  description: string
}

export type RoleColumnKey = 'admin' | 'executive' | 'auditor' | 'auditee' | 'departmentHead' | 'viewer'

/**
 * VIEWER is strictly read-only: it can view business modules but cannot create, edit, delete, approve, publish,
 * import or upload, cannot open or download any file (documents, evidence, working papers, uploads, signatures),
 * cannot export, has no master data, and in Account Settings sees only My Profile, Activity and FAQ.
 * Write rules: master data writes are ADMIN only; QA reports and consulting services writes are ADMIN, EXECUTIVE
 * (CAE) and AUDIT_MANAGER; audit universe edits are ADMIN and EXECUTIVE (CAE), and DEPARTMENT_HEAD
 * may edit rows of its own department only (AUDIT_UNIVERSE_* in viewerPolicy.ts). Rights are the union of a
 * user's roles and VIEWER adds nothing.
 * This grid only documents the policy for the UI. The backend is authoritative and enforces it on every request.
 */
export const matrixData: MatrixItem[] = [
  { id: '1', num: 1, module: 'Dashboard', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '2', num: 2, module: 'Audit Charter', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '3-1', num: 3, module: 'Risk Profile', submodule: 'Corporate Risk Profile', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'FULL', auditee: 'EDIT — Own Dept.', departmentHead: 'EDIT — Own Dept.', viewer: 'READ' },
  { id: '3-2', num: 3, module: 'Risk Profile', submodule: 'Risk Appetite Statement', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '3-3', num: 3, module: 'Risk Profile', submodule: 'Risk Factors & Scoring', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '3-4', num: 3, module: 'Risk Profile', submodule: 'Audit Universe', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'EDIT — Own Dept.', viewer: 'READ' },
  { id: '3-5', num: 3, module: 'Risk Profile', submodule: 'Risk Control Matrix (RCM)', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'FULL', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '4', num: 4, module: 'Strategic Audit Plan', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '5-1', num: 5, module: 'Annual Audit Plan', submodule: 'Create / Manage Plan', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '5-2', num: 5, module: 'Annual Audit Plan', submodule: 'Import Plan Document', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'NONE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '5-3', num: 5, module: 'Annual Audit Plan', submodule: 'Execution Status', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '6-1', num: 6, module: 'Audit Activity Plan', submodule: 'Create Activity Plan', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'DRAFT', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '6-2', num: 6, module: 'Audit Activity Plan', submodule: 'Import Activity Plan', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'NONE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '7-1', num: 7, module: 'Assignment Letter', submodule: 'Create / Publish Letter', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '7-2', num: 7, module: 'Assignment Letter', submodule: 'Import Letter', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'NONE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '8-1', num: 8, module: 'Audit Fieldwork', submodule: 'Fieldwork — Interviews, Sampling, Testing, etc.', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'FULL', auditee: 'RESPOND', departmentHead: 'RESPOND', viewer: 'READ' },
  { id: '8-2', num: 8, module: 'Audit Fieldwork', submodule: 'Working Papers — F-01 to F-05', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'CREATE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '8-3', num: 8, module: 'Audit Fieldwork', submodule: 'Import Working Papers', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'UPLOAD', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '9-1', num: 9, module: 'Audit Result Report', submodule: 'Result Report / LHA', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'DRAFT', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '9-2', num: 9, module: 'Audit Result Report', submodule: 'Executive Summary', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '9-3', num: 9, module: 'Audit Result Report', submodule: 'Client Satisfaction Survey', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'FILL SURVEY', departmentHead: 'FILL SURVEY', viewer: 'READ' },
  { id: '10', num: 10, module: 'Action Taken Report (ATR)', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'UPDATE CAPA', departmentHead: 'UPDATE CAPA', viewer: 'READ' },
  { id: '11', num: 11, module: 'KPI Performance', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '12', num: 12, module: 'Consulting Service', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '13', num: 13, module: 'Quality Assurance (QA)', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'NONE', departmentHead: 'NONE', viewer: 'READ' },
  { id: '14', num: 14, module: 'Analytics', submodule: '', isSubmodule: false, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '15-1', num: 15, module: 'Settings & User Management', submodule: 'User Management & System Settings', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'NONE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '15-2', num: 15, module: 'Account Settings', submodule: 'My Profile, Activity & FAQ', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'READ' },
  { id: '16-1', num: 16, module: 'Master Data', submodule: 'Company, Business Unit, Department, Location, Job Role, Vision-Mission-Goals', isSubmodule: true, admin: 'FULL', executive: 'READ', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'NONE' },
  { id: '16-2', num: 16, module: 'Master Data', submodule: 'Employee Management', isSubmodule: true, admin: 'FULL', executive: 'READ', auditor: 'NONE', auditee: 'NONE', departmentHead: 'NONE', viewer: 'NONE' },
  { id: '17', num: 17, module: 'Files & Exports', submodule: 'Download / Preview Files, Export Reports', isSubmodule: true, admin: 'FULL', executive: 'FULL', auditor: 'READ', auditee: 'READ', departmentHead: 'READ', viewer: 'NONE' },
]

/**
 * Returns Tailwind badge classes corresponding to access level
 */
export function getBadgeClass(val: string): string {
  const base = 'inline-flex items-center justify-center px-2.5 py-1 rounded-md font-bold text-xs leading-none transition-colors shadow-xs'

  if (val === 'FULL') {
    return base + ' bg-emerald-600 text-white'
  }
  if (val === 'READ') {
    return base + ' bg-sky-600 text-white'
  }
  if (val === 'NONE') {
    return base + ' bg-slate-500 text-white'
  }
  // Action-specific / Limited access (e.g. RESPOND, UPDATE CAPA, EDIT — Own Dept., FILL SURVEY, REQUEST, etc.)
  return base + ' bg-amber-500 text-white'
}

/**
 * Normalizes any role representation into one of the 6 RBAC matrix column keys.
 */
export function normalizeRoleToColumn(role: string): RoleColumnKey | null {
  const cleaned = String(role || '').toLowerCase().replace(/[\s_-]+/g, '')
  if (!cleaned) return null

  // ADMIN
  if (cleaned === 'admin' || cleaned === 'administrator' || cleaned === 'superadmin') {
    return 'admin'
  }

  // EXECUTIVE (Chief Audit Executive / CAE)
  if (
    cleaned === 'executive' ||
    cleaned === 'chiefauditexecutive' ||
    cleaned === 'cae' ||
    cleaned === 'kepalaspi' ||
    cleaned === 'headofskai'
  ) {
    return 'executive'
  }

  // AUDITOR
  if (
    cleaned === 'auditor' ||
    cleaned === 'auditstaff' ||
    cleaned === 'staffaudit' ||
    cleaned === 'leadauditor'
  ) {
    return 'auditor'
  }

  // AUDITEE
  if (cleaned === 'auditee' || cleaned === 'pic') {
    return 'auditee'
  }

  // DEPARTMENT HEAD
  if (
    cleaned === 'departmenthead' ||
    cleaned === 'depthead' ||
    cleaned === 'headofdepartment' ||
    cleaned === 'kepaladepartemen' ||
    cleaned === 'kepalaunit'
  ) {
    return 'departmentHead'
  }

  // VIEWER
  if (
    cleaned === 'viewer' ||
    cleaned === 'direksi' ||
    cleaned === 'komiteaudit' ||
    cleaned === 'board' ||
    cleaned === 'boc' ||
    cleaned === 'bod' ||
    cleaned === 'komisaris'
  ) {
    return 'viewer'
  }

  return null
}

/**
 * Assigns weight to access level to pick the most permissive right when a user has multiple roles.
 */
export function getAccessLevelWeight(level: string): number {
  if (!level || level === 'NONE') return 0
  if (level === 'FULL') return 100
  if (level === 'READ') return 20
  // Custom operational action (EDIT, RESPOND, UPDATE CAPA, FILL SURVEY, REQUEST, DRAFT, CREATE, UPLOAD)
  return 50
}

/**
 * Returns descriptive human-readable label for the permission in Indonesian or English.
 */
export function getPermissionDescription(item: MatrixItem, accessLevel: string, locale: 'id' | 'en' = 'id'): string {
  const isEn = locale === 'en'

  if (accessLevel === 'FULL') {
    return isEn
      ? 'Full access (create, read, update, delete, approval)'
      : 'Akses penuh modul (buat, baca, ubah, hapus, persetujuan)'
  }

  if (accessLevel === 'READ') {
    return isEn
      ? 'Read-only access to view documents and data'
      : 'Akses hanya baca untuk melihat dokumen dan data'
  }

  if (accessLevel.includes('EDIT')) {
    return isEn
      ? 'Fill and edit risk profile for own department'
      : 'Mengisi dan memperbarui profil risiko untuk unit kerja sendiri'
  }

  if (accessLevel === 'RESPOND') {
    return isEn
      ? 'Respond to audit inquiries (PBC), interviews, and sample testing'
      : 'Merespon permintaan data (PBC), wawancara, dan bukti audit'
  }

  if (accessLevel === 'FILL SURVEY') {
    return isEn
      ? 'Complete client satisfaction survey for audit engagements'
      : 'Mengisi survei kepuasan atas pelaksanaan layanan audit'
  }

  if (accessLevel === 'UPDATE CAPA') {
    return isEn
      ? 'Update action plan and upload corrective action evidence (CAPA)'
      : 'Memperbarui rencana tindak lanjut dan bukti perbaikan rekomendasi (CAPA)'
  }

  if (accessLevel === 'REQUEST') {
    return isEn
      ? 'Submit advisory and consulting service requests'
      : 'Mengajukan permohonan layanan konsultasi tata kelola & risiko'
  }

  if (accessLevel === 'DRAFT') {
    return isEn
      ? 'Draft audit activity plan or draft audit result report'
      : 'Menyusun draf rencana kegiatan atau draf laporan hasil audit'
  }

  if (accessLevel === 'CREATE') {
    return isEn
      ? 'Create and fill audit working papers (KKA F-01 to F-05)'
      : 'Menyusun dan mengisi kertas kerja audit (KKA F-01 s/d F-05)'
  }

  if (accessLevel === 'UPLOAD') {
    return isEn
      ? 'Upload working papers and supporting attachments'
      : 'Mengunggah kertas kerja audit dan berkas pendukung'
  }

  return isEn ? `Operational access (${accessLevel})` : `Hak akses operasional (${accessLevel})`
}

/**
 * Derives list of active permissions for the given roles from the RBAC Matrix in Account Settings.
 * Items with 'NONE' access level are excluded.
 */
export function getPermissionsForUserRoles(roles: (string | any)[], locale: 'id' | 'en' = 'id'): UserPermissionItem[] {
  // Normalize user roles to columns
  const matchedColumns = new Set<RoleColumnKey>()

  const extractRoleString = (r: any): string => {
    if (!r) return ''
    if (typeof r === 'string') return r
    if (typeof r === 'object') {
      return r.name || r.role || r.code || r.title || ''
    }
    return String(r)
  }

  for (const rawRole of roles) {
    const roleStr = extractRoleString(rawRole)
    const col = normalizeRoleToColumn(roleStr)
    if (col) {
      matchedColumns.add(col)
    }
  }

  // If no columns matched (e.g. unknown role), fallback check if any role contains auditee/auditor/admin
  if (matchedColumns.size === 0) {
    for (const rawRole of roles) {
      const s = extractRoleString(rawRole).toLowerCase()
      if (s.includes('auditee')) matchedColumns.add('auditee')
      else if (s.includes('auditor')) matchedColumns.add('auditor')
      else if (s.includes('admin')) matchedColumns.add('admin')
      else if (s.includes('exec') || s.includes('cae')) matchedColumns.add('executive')
      else if (s.includes('dept') || s.includes('head')) matchedColumns.add('departmentHead')
      else if (s.includes('view')) matchedColumns.add('viewer')
    }
  }

  // If still empty, return empty list
  if (matchedColumns.size === 0) {
    return []
  }

  const result: UserPermissionItem[] = []

  for (const item of matrixData) {
    let highestLevel = 'NONE'
    let highestWeight = 0

    for (const col of matchedColumns) {
      const level = item[col] || 'NONE'
      const weight = getAccessLevelWeight(level)
      if (weight > highestWeight) {
        highestWeight = weight
        highestLevel = level
      }
    }

    // Only include features where user has active access (weight > 0 and level !== 'NONE')
    if (highestWeight > 0 && highestLevel !== 'NONE') {
      const displayName = item.isSubmodule && item.submodule
        ? `${item.module} — ${item.submodule}`
        : item.module

      result.push({
        id: item.id,
        num: item.num,
        module: item.module,
        submodule: item.submodule || undefined,
        displayName,
        accessLevel: highestLevel,
        badgeClass: getBadgeClass(highestLevel),
        description: getPermissionDescription(item, highestLevel, locale),
      })
    }
  }

  return result
}
