// Audit Types

export enum AuditType {
  ASSURANCE = "Assurance",
  SPECIAL_AUDIT = "Special Audit",
  SPECIFIC_REASON = "Specific Reason",
  CONSULTING_SERVICES = "Consulting Services",
  INVESTIGATION = "Investigation",
  QUALITY_ASSURANCE_REVIEW = "Quality Assurance Review",
  FOLLOWUP_AUDIT = "Follow-Up Audit",
  OPERASIONAL = "Operasional",
  FINANCIAL = "Financial",
  IT = "IT",
}

export enum AuditStatus {
  PLANNED = "Planned",
  IN_PROGRESS = "In Progress",
  FIELDWORK = "Fieldwork",
  REPORTING = "Reporting",
  COMPLETED = "Completed",
  CANCELLED = "Cancelled",
}

export interface AuditPlan {
  id: string;
  year: number;
  activities: AuditActivity[];
  createdAt: string;
  updatedAt: string;
}

export interface AuditActivity {
  id: string;
  type: AuditType;
  title: string;
  subject: string;
  auditUniverse: string;
  justification: string;
  purpose: string;
  focus: string;
  scheduleStart: string;
  scheduleEnd: string;
  teamSize: number;
  totalManDays: number;
  teamLeader: string;
  status: AuditStatus;
}

export interface AuditAssignment {
  id: string;
  activityId: string;
  letterNumber: string;
  teamMembers: AuditTeamMember[];
  auditPurpose: string;
  scope: string;
  period: string;
  chiefAuditorExecutive: string;
  issueDate: string;
}

export interface AuditTeamMember {
  name: string;
  role: string;
  scheduleStart: string;
  scheduleEnd: string;
}

export interface AuditFinding {
  id: string;
  auditId: string;
  title: string;
  description: string;
  severity: FindingSeverity;
  recommendation: string;
  status: FindingStatus;
  responsiblePerson: string;
  dueDate: string;
  createdAt: string;
  updatedAt: string;
}

export enum FindingSeverity {
  LOW = "Low",
  MEDIUM = "Medium",
  HIGH = "High",
  CRITICAL = "Critical",
}

export enum FindingStatus {
  OPEN = "Open",
  IN_PROGRESS = "In Progress",
  RESOLVED = "Resolved",
  CLOSED = "Closed",
}

export interface AuditCharter {
  id: string;
  title: string;
  version: string;
  date: string; // ISO Date string (YYYY-MM-DD)
  uploadedBy: string;
  approvedBy: string;
  isActive: boolean;
  fileName?: string;
  fileUrl?: string; // Simulasi URL file
  fileSize?: string;
}

export interface CharterFormState {
  title: string;
  version: string;
  date: string;
  uploadedBy: string;
  approvedBy: string;
  isActive: boolean;
  file: File | null;
}

export interface AnnualAuditPlan {
  id?: string;
  code: string; // Kode Kegiatan (e.g., ASR-01)
  name: string; // Nama Kegiatan
  type: AuditType; // Kategori (Bisa custom, jadi string)
  selectedMonths: number[]; // 0=Jan, 11=Dec
  quarters?: string[]; // Calculated: ['Q1', 'Q2']
  auditorCount: number; // Jumlah Auditor (1-10)
  daysPerAuditor: number; // Durasi per auditor
  totalMandays?: number; // Calculated: count * days
  supervisorId: string; // ID Supervisor
  supervisorName?: string; // Nama Supervisor
  year: string;
  notes?: string;
  status: AuditStatus;
  isActive: boolean; // Status Aktif/Non-aktif
  isUsed?: boolean; // Flag untuk cek apakah sudah dipakai di RKAT (Simulasi)
  //auditUniverse: string;  Unit/Area yang diaudit
  //auditCycle: string;     e.g., "Annually", "2 Years"
  //lastAudit: string;      Tahun terakhir audit
}

export interface AnnualPlanForm {
  id?: string;
  code: string;
  name: string;
  type: AuditType;
  selectedMonths: number[];
  status: AuditStatus;
  auditorCount: number;
  daysPerAuditor: number;
  supervisorId: string;
  notes?: string;
  year: string;
  isActive: boolean;
  // auditUniverse: string;
  // auditCycle: string;
  // lastAudit: string;
}

export interface AuditMainStats {
  plannedAudit: number;
  openFinding: number;
  executionStatus: number; // 0-1 range (e.g., 0.8 = 80%)
  atrCompliance: number; // 0-1 range (e.g., 0.9 = 90%)
}

export interface AuditCoverage {
  plannedAudits: number;
  completedAudits: number;
  remainingAudits: number;
}

export interface AuditCategoryRisk {
  id: number;
  name: string;
  inherentRisk: number;
  residualRisk: number;
}

export interface AtrReport {
  id: string;
  name: string;
  owner: string;
  date: string;
  status: string;
}

export interface AuditExecutionStatus{
  id: number;
  name: string;
  percentage: number;
}

export interface RecentFinding{
  id: number;
  audit_finding: string;
  audit_category: string;
  severity: string;
}

export interface AuditDataState {
  auditMainStats: AuditMainStats;
  auditMainStatsLastMonth: AuditMainStats;
  auditCoverage: AuditCoverage;
  auditCategories: AuditCategoryRisk[];
  auditExecutionStatus: AuditExecutionStatus[];
  atrReports: AtrReport[];
  recentFindings:  RecentFinding[];
}
