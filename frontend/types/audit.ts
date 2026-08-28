// Audit Types
import type { RiskLevel, RiskTaxonomy } from "./risk";

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
  PENDING_APPROVAL = "Pending Approval",
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

export type OrganizationIdentity = 'SKAI' | 'DAI' | 'CAE';
export type TeamRole = 'Person Responsible' | 'Supervisor' | 'Head Member' | 'Member';
export type AssignmentLetterStatus = 'Draft' | 'Waiting Approval' | 'Published' | 'Archived';

// 1. Definisikan Array-nya terlebih dahulu
export const TEST_RESULT_OPTIONS = ['Pass', 'Fail', 'N/A'] as const;
export const ROOT_CAUSE_METHOD_OPTIONS = ['People', 'Process', 'Policy', 'System', 'External'] as const;

// 2. Ambil Tipe-nya dari Array tersebut (Otomatis sinkron)
export type TestResult = typeof TEST_RESULT_OPTIONS[number] | undefined;
export type RootCauseMethod = typeof ROOT_CAUSE_METHOD_OPTIONS[number];

export interface AuditPlan {
  id: string;
  year: number;
  activities: AuditActivity[];
  createdAt: string;
  updatedAt: string;
}

export interface AuditActivity {
  id: string;
  annualPlanId: string;
  targetUnitId: string;
  projectCode: string;
  title: string;
  engagementSubject: string;
  auditType: string;
  auditUniverseId: string;
  justification: string;
  auditPurpose: string;
  auditFocus: string;
  plannedStart: string;
  plannedEnd: string;
  teamSize: number;
  totalMandays: number;
  teamLeaderId: string;
  status: AuditStatus;
  objective: string;
  scope: string;
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
  content?: string;
  date: string; // ISO Date string (YYYY-MM-DD)
  uploadedBy: string;
  approvedBy: string;
  isActive: boolean;
  fileName?: string;
  fileUrl?: string;
  fileSize?: string;
}

export interface CharterFormState {
  title: string;
  version: string;
  content?: string;
  date: string;
  uploadedBy: string;
  approvedBy: string;
  isActive: boolean;
  file: File | null;
}

export interface AuditActivities {
  id?: string;
  itemNumber?: string;
  category: AuditCategory | string;
  groupTitle?: string;
  name: string;
  department?: AuditDepartment | string;
  involvedDepartments?: any[];
  timelineText?: string;
  auditorCount?: number;
  totalMandays?: number;
  supervisorName?: string;
  notesObjective?: string;
  riskName?: string;
  riskLevel?: string;
}

export interface RevisionHistory {
  date: string;
  version: string;
  changes: string;
  user: string;
}

export interface AnnualAuditPlan {
  id?: string;
  code: string;
  version?: string;
  parentPlanId?: string;
  revisionHistory?: RevisionHistory[];
  activities: AuditActivities[];
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];
  quarters?: string[];
  auditorCount: number;
  daysPerAuditor: number;
  totalMandays?: number;      // Calculated: count * days
  supervisorId?: string;
  supervisorName?: string;
  year: string;
  notes?: string;
  attachmentCategory: string;
  attachmentUploadedBy: string;
  attachmentUploadDate: string;
  attachments?: AnnualAuditAttachment[];
  isActive: boolean;
  isUsed?: boolean;
}

export interface AnnualAuditAttachment {
  name: string;
  size: string; // e.g., "2.5 MB"
  url: string; // URL objek lokal atau URL unduhan dari server
}

export interface AnnualPlanForm {
  id?: string;
  code: string;
  version?: string;
  parentPlanId?: string;
  revisionHistory?: RevisionHistory[];
  activities: AuditActivities[];
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];
  auditorCount: number;
  daysPerAuditor: number;
  supervisorId?: string;
  notes?: string;
  year: string;
  attachmentCategory: string;
  attachmentUploadedBy: string;
  attachmentUploadDate: string;
  attachments?: AnnualAuditAttachment[];
  file: File[] | null;
  staffApprovalNote: string;
  managerApprovalNote: string;
  chiefApprovalNote: string;
  isActive: boolean;
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

export interface AuditExecutionStatus {
  id: number;
  name: string;
  percentage: number;
}

export interface RecentFinding {
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
}

export interface SampleItem {
  id: number;
  document: string;
  l1: TestResult;
  l2: TestResult;
  l3: TestResult;
}

export interface RootCauseItem {
  id: number;
  method: RootCauseMethod;
  w1: string;
  w2: string;
  w3: string;
}

export interface TeamMember {
  id: number;
  name: string;
  role: string;
}

export interface WorkingPaperHeader {
  id?: string;
  assignmentLetterId: string;
  auditPurpose: string;
  businessProcess: string;
  period: string;
  location: string;
  teamMembers: TeamMember[];
}

export interface WorkingPaperHeaderForm {
  id?: string;
  assignmentLetterId: string;
  auditPurpose: string;
  businessProcess: string;
  periodStart: string;
  periodEnd: string;
  location: string;
  teamMembers: TeamMember[];
}

export interface WorkingPaperRisk {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  risk: string;
  taxonomy: RiskTaxonomy;
  riskLevel: RiskLevel;
  controlDescription: string;
}

export interface WorkingPaperRiskForm {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  risk: string;
  taxonomy: RiskTaxonomy;
  riskLevel: RiskLevel;
  controlDescription: string;
}

export interface WorkingPaperSample {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  population: number | null;
  sampleSize: number | null;
  samples: SampleItem[];
  conclusion: string;
}

export interface WorkingPaperSampleForm {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  population: number | null;
  sampleSize: number | null;
  samples: SampleItem[];
  conclusion: string;
}

export interface WorkingPaperCause {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  condition: string;
  criteria: string;
  impact: string;
  evidenceFile?: File | any | null;
  rootCause: RootCauseItem[];
}

export interface WorkingPaperCauseForm {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  condition: string;
  criteria: string;
  impact: string;
  evidenceFile?: File | any | null;
  rootCause: RootCauseItem[];
}

export interface WorkingPaperPlan {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  recommendation: string;
  response: string;
  actionDescription: string;
  pic: string;
  periodAction: string;
}

export interface WorkingPaperPlanForm {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  recommendation: string;
  response: string;
  actionDescription: string;
  pic: string;
  periodAction: string;
}

export interface TeamAssignmentLetter {
  name: string;
  role: string;
}

// Struktur Form Input (Sama persis dengan v-model di UI)
export interface AssignmentLetterForm {
  auditTitle: string;
  leader: string;
  category: AuditCategory;
  auditYear: string;
  auditTeam: string;
  startPeriod: string;  // Format YYYY-MM-DD
  finishPeriod: string; // Format YYYY-MM-DD
  workingUnit: string;
  auditPurpose: string;
  letterDate?: string;
  caeSignature?: string;
  membersList: TeamAssignmentLetter[];
  purposeList: string[];
  scopeList: string[];
  ccList: string[];
}

// Struktur Data Utama di Database/Store
export interface AssignmentLetter extends AssignmentLetterForm {
  id: string;
  letterNumber: string; // Di-generate otomatis oleh Store (F-01)
  executionPeriod: string;
  status: AssignmentLetterStatus;
  createdAt: string;
}

// --- Audit Activity Plan Types ---

export interface PlannedAuditActivity {
  id?: string;
  auditName: string;
  auditee: string;
  category: AuditCategory;
  riskName?: string;
  riskLevel: RiskLevel;
  duration: number;
  priority: string;
  numberOfAuditors: number;
  estimatedSchedule: string;
  budgetEstimation: string;
}

export interface ResourceAuditor {
  id?: string;
  name: string;
  position: string;
  competence: string;
  availability: string;
}

export interface ActivityPlanBudget {
  totalEstimatedCost: number;
  totalAllocatedBudget: number;
  budgetNotes: string;
}

export interface ActivityPlanReview {
  creatorName: string;
  creatorPosition: string;
  approverName: string;
  approverPosition: string;
  approvalDate: string;
  additionalNotes: string;
}

export interface ActivityPlanFormState {
  planTitle: string;
  planYear: string;
  planPeriodStart: string;
  planPeriodEnd: string;
  department: AuditDepartment;
  createdBy: string;
  creationDate: string;
  plannedActivities: PlannedAuditActivity[];
  resourceAuditors: ResourceAuditor[];
  budget: ActivityPlanBudget;
  review: ActivityPlanReview;
  attachmentCategory?: string;
  attachmentUploadedBy?: string;
  attachmentUploadDate?: string;
  attachments?: AnnualAuditAttachment[];
  file?: File[] | null;
}

export interface ActivityPlan extends ActivityPlanFormState {
  id: string;
  status: string;
  createdAt: string;
}

export interface KPITargetYear {
  year: number;
  value: string;
}

export type HibHigType = 'HIG' | 'HIB';
export type PeriodType = 'Quartal' | 'Yearly';
export type QuarterType = 'Q1' | 'Q2' | 'Q3' | 'Q4';

export interface StrategicAuditPlan {
  id: string;
  code: string;
  goalId?: string;
  strategicObjective: string;
  kpi: string;
  unit: string;
  hibHig: HibHigType;
  periodType: PeriodType;
  selectedPeriod: string; // Q1-Q4 for Quartal, or year string for Yearly
  yearStart?: number;
  yearEnd?: number;
  kpiTargets: Record<string | number, string>;
  kpiActuals?: Record<string | number, string>;
  internalAuditSO: string;
  actual: string;
  target: string;
  calculation: string;
  status: string;
}

export interface ActionTakenReport {
  id: string
  assignment_letter_id?: string
  assignmentLetterId?: string
  assignment_letter?: AssignmentLetter
  assignmentLetter?: AssignmentLetter
  auditRef: AnnualAuditPlan['code']
  title: string
  department: AuditDepartment | string
  deadline: string
  status: AuditStatus | string
  auditObject?: string
  findingCategory?: AuditCategory | string
  condition?: string
  criteria?: string
  recommendation?: string
  pic?: string
  attachment?: string
  progressDescription?: string
}

export interface AuditExecution {
  id: string
  ref: ActionTakenReport['auditRef']
  name: ActionTakenReport['title']
  category: AuditCategory | string
  department?: AuditDepartment | string
  progress: number
  lead_auditor: string
  status: AuditStatus
  status_detail?: 'Late' | 'On Time'
  sample_data_test_controls?: {
    progress: number
    description: string
  }
  working_papers?: {
    condition: string
    criteria: string
  }
  action_plan_improvements?: {
    recommendation: string
    deadline: string
    pic: string
  }
  latest_update_progress?: {
    attachment: string
    description: string
  }
}

export interface ExecutionPhase {
  step: number
  title: string
  shortLabel: string
  description: string
  icon: string
  badgeColor: 'neutral' | 'info' | 'primary' | 'warning' | 'success' | 'error' | 'secondary'
  badgeClass: string
  iconClass: string
  numBgClass: string
  cardClass: string
  minProgress: number
  maxProgress: number
}

export const EXECUTION_PHASES: ExecutionPhase[] = [
  {
    step: 1,
    title: 'Perencanaan & Persiapan',
    shortLabel: 'Perencanaan',
    description: 'Penerbitan surat tugas, alokasi tim audit & pengumpulan data awal auditee.',
    icon: 'i-lucide-calendar-clock',
    badgeColor: 'info',
    badgeClass: 'bg-sky-100 text-sky-700 dark:bg-sky-950 dark:text-sky-300 border border-sky-300 dark:border-sky-700',
    iconClass: 'text-sky-500 dark:text-sky-400',
    numBgClass: 'bg-sky-500 text-white',
    cardClass: 'border-sky-200 dark:border-sky-900/60 bg-sky-50/30 dark:bg-sky-950/20',
    minProgress: 0,
    maxProgress: 0
  },
  {
    step: 2,
    title: 'Entry Meeting & Penyelarasan Scope',
    shortLabel: 'Entry Meeting',
    description: 'Pertemuan pembuka dengan auditee, konfirmasi ruang lingkup & finalisasi program audit.',
    icon: 'i-lucide-users',
    badgeColor: 'info',
    badgeClass: 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border border-blue-300 dark:border-blue-700',
    iconClass: 'text-blue-500 dark:text-blue-400',
    numBgClass: 'bg-blue-500 text-white',
    cardClass: 'border-blue-200 dark:border-blue-900/60 bg-blue-50/30 dark:bg-blue-950/20',
    minProgress: 1,
    maxProgress: 25
  },
  {
    step: 3,
    title: 'Fieldwork & Pengujian Pengendalian',
    shortLabel: 'Fieldwork',
    description: 'Pengujian sampel kontrol, wawancara staf, pengisian kertas kerja & pengumpulan bukti audit.',
    icon: 'i-lucide-clipboard-check',
    badgeColor: 'warning',
    badgeClass: 'bg-violet-100 text-violet-700 dark:bg-violet-950 dark:text-violet-300 border border-violet-300 dark:border-violet-700',
    iconClass: 'text-violet-500 dark:text-violet-400',
    numBgClass: 'bg-violet-500 text-white',
    cardClass: 'border-violet-200 dark:border-violet-900/60 bg-violet-50/30 dark:bg-violet-950/20',
    minProgress: 26,
    maxProgress: 50
  },
  {
    step: 4,
    title: 'Draft Temuan & Rekomendasi',
    shortLabel: 'Draft Temuan',
    description: 'Penyusunan temuan audit 5-C, konfirmasi tanggapan manajemen & komitmen PIC penanggung jawab.',
    icon: 'i-lucide-file-warning',
    badgeColor: 'secondary',
    badgeClass: 'bg-purple-100 text-purple-700 dark:bg-purple-950 dark:text-purple-300 border border-purple-300 dark:border-purple-700',
    iconClass: 'text-purple-500 dark:text-purple-400',
    numBgClass: 'bg-purple-500 text-white',
    cardClass: 'border-purple-200 dark:border-purple-900/60 bg-purple-50/30 dark:bg-purple-950/20',
    minProgress: 51,
    maxProgress: 75
  },
  {
    step: 5,
    title: 'Pelaporan & Exit Meeting',
    shortLabel: 'Pelaporan',
    description: 'Penyusunan draft Laporan Hasil Audit (LHA), pertemuan penutup dengan auditee & persetujuan CAE.',
    icon: 'i-lucide-file-text',
    badgeColor: 'primary',
    badgeClass: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300 border border-indigo-300 dark:border-indigo-700',
    iconClass: 'text-indigo-500 dark:text-indigo-400',
    numBgClass: 'bg-indigo-500 text-white',
    cardClass: 'border-indigo-200 dark:border-indigo-900/60 bg-indigo-50/30 dark:bg-indigo-950/20',
    minProgress: 76,
    maxProgress: 99
  },
  {
    step: 6,
    title: 'Audit Selesai',
    shortLabel: 'Selesai',
    description: 'Laporan Hasil Audit final diterbitkan & siap untuk pemantauan Tindak Lanjut (Action Plan).',
    icon: 'i-lucide-check-circle-2',
    badgeColor: 'secondary',
    badgeClass: 'bg-secondary-100 text-secondary-700 dark:bg-secondary-950 dark:text-secondary-300 border border-secondary-300 dark:border-secondary-700',
    iconClass: 'text-secondary-600 dark:text-secondary-400',
    numBgClass: 'bg-secondary-600 text-white',
    cardClass: 'border-secondary-300 dark:border-secondary-800 bg-secondary-50/40 dark:bg-secondary-950/30',
    minProgress: 100,
    maxProgress: 100
  }
]

export const getExecutionPhase = (progress: number = 0): ExecutionPhase => {
  const p = Number(progress) || 0
  if (p >= 100) return EXECUTION_PHASES[5]!
  if (p >= 76) return EXECUTION_PHASES[4]!
  if (p >= 51) return EXECUTION_PHASES[3]!
  if (p >= 26) return EXECUTION_PHASES[2]!
  if (p >= 1) return EXECUTION_PHASES[1]!
  return EXECUTION_PHASES[0]!
}
