// Audit Types

export enum AuditCategory {
  ASSURANCE = "Assurance",
  SPECIAL_AUDIT = "Special Audit",
  SPECIFIC_REASON = "Specific Reason",
  CONSULTING_SERVICES = "Consulting Services",
  INVESTIGATION = "Investigation",
  QUALITY_ASSURANCE_REVIEW = "Quality Assurance Review",
  FOLLOWUP_AUDIT = "Follow-Up Audit",
}

export enum AuditStatus {
  PLANNED = "Planned",
  IN_PROGRESS = "In Progress",
  FIELDWORK = "Fieldwork",
  REPORTING = "Reporting",
  COMPLETED = "Completed",
  CANCELLED = "Cancelled",
}

export enum AnnualAuditPlanStatus {
  DONE = "Done",
  WORK_IN_PROGRESS = "Work In Progress",
  NOT_AVAILABLE = "Not Available"
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

export enum AuditDepartment {
  IT = "IT",
  FINANCE = "Finance",
  HR = "HR",
  OPS = "Ops"
}

export type IdentitasOrganisasi = 'SKAI' | 'DAI' | 'CAE';
export type PeranTim = 'Penanggung Jawab' | 'Pengawas' | 'Ketua' | 'Anggota';
export type SuratTugasStatus = 'Draft' | 'Waiting Approval' | 'Published' | 'Archived';

// 1. Definisikan Array-nya terlebih dahulu
export const TEST_RESULT_OPTIONS = ['Pass', 'Fail', 'N/A'] as const;
export const RCA_METHOD_OPTIONS = ['People', 'Process', 'Policy', 'System', 'External'] as const;

// 2. Ambil Tipe-nya dari Array tersebut (Otomatis sinkron)
export type TestResult = typeof TEST_RESULT_OPTIONS[number] | undefined;
export type RCAMethod = typeof RCA_METHOD_OPTIONS[number];

export interface AuditPlan {
  id: string;
  year: number;
  activities: AuditActivity[];
  createdAt: string;
  updatedAt: string;
}

export interface AuditActivity {
  id: string;
  category: AuditCategory;
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

export interface AuditActivities {
  name: string;
  category: AuditCategory;
  department: AuditDepartment;
}

export interface AnnualAuditPlan {
  id?: string;
  code: string; 
  activities: AuditActivities[];    
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];  // 0=Jan, 11=Dec
  quarters?: string[];        // Calculated: ['Q1', 'Q2']
  auditorCount: number;      // Jumlah Auditor (1-10)
  daysPerAuditor: number;    // Durasi per auditor
  totalMandays?: number;      // Calculated: count * days
  supervisorId: string;      // ID Supervisor
  supervisorName?: string;    // Nama Supervisor
  year: string;
  notes?: string;
  isActive: boolean; // Status Aktif/Non-aktif
  isUsed?: boolean; // Flag untuk cek apakah sudah dipakai di RKAT (Simulasi)
  //auditUniverse: string;  Unit/Area yang diaudit
  //auditCycle: string;     e.g., "Annually", "2 Years"
  //lastAudit: string;      Tahun terakhir audit
}

export interface AnnualPlanForm {
  id?: string;
  code: string;
  activities: AuditActivities[];
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];
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

export interface SampleItem {
  id: number;
  dokumen: string;
  l1: TestResult;
  l2: TestResult;
  l3: TestResult;
}

export interface RCAItem {
  id: number;
  method: RCAMethod;
  w1: string;
  w2: string;
  w3: string;
}

export interface TeamMember {
  id: number;
  name: string;
  role: string;
}

export interface WorkingPaperForm {
  // F-01: Referensi Penugasan
  suratTugas: string;
  tujuanAudit: string;
  prosesBisnis: string;
  periodeStart: string;
  periodeEnd: string;
  lokasi: string;
  teamMembers: TeamMember[];

  // F-02: Risk Profile
  resiko: string;
  taksonomi: string;
  tingkatResiko: string;
  deskripsiKontrol: string;

  // F-03: Uji Sampel
  populasi: number | null;
  jmlSampel: number | null;
  samples: SampleItem[];
  kesimpulan: string;

  // F-04: AOI & RCA
  kondisi: string;
  kriteria: string;
  dampak: string;
  buktiFile: File | null; // Untuk upload file
  rcaList: RCAItem[];

  // F-05: Action Plan
  rekomendasi: string;
  tanggapan: string;
  deskripsiAction: string;
  pic: string;
  periodAction: string;
}

export interface SuratTugasTim {
  name: string;
  role: string;
}

// Struktur Form Input (Sama persis dengan v-model di UI)
export interface SuratTugasForm {
  auditTitle: string;
  leader: string;
  category: AuditCategory;
  auditYear: string;
  auditTeam: string;
  startPeriod: string;  // Format YYYY-MM-DD
  finishPeriod: string; // Format YYYY-MM-DD
  workingUnit: string;
  membersList: SuratTugasTim[];
  purposeList: string[];
  scopeList: string[];
  ccList: string[];
}

// Struktur Data Utama di Database/Store
export interface SuratTugas extends SuratTugasForm {
  id: string;
  letterNumber: string; // Di-generate otomatis oleh Store (F-01)
  executionPeriod: string;
  status: SuratTugasStatus;
  createdAt: string;
}
