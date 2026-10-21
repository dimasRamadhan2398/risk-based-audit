// Audit Types
import type { RiskLevel, RiskTaxonomy } from "./risk";
import { getExecutionStageIndex } from "~/utils/executionSources";

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
  PLANNING = "planning",
  ENTRY_MEETING = "entry meeting",
  DRAFT_FINDINGS = "draft findings",
}

export enum AuditExecutionStatus {
  PLANNING = "planning",
  ENTRY_MEETING = "entry meeting",
  FIELDWORK = "fieldwork",
  DRAFT_FINDINGS = "draft findings",
  REPORTING = "reporting",
  COMPLETED = "completed",
}

export enum AnnualAuditPlanStatus {
  DRAFT = "Draft",
  PENDING_REVIEW = "Pending Review",
  PENDING_APPROVAL = "Pending Approval",
  APPROVED = "Approved",
  REVISED = "Revised"
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
  fileName?: string;
}

/**
 * Jenis Kegiatan Audit: the header of an Audit Activity Plan (Formulir Rencana Kegiatan Audit).
 * Values are the backend strings.
 */
export const ACTIVITY_PLAN_ACTIVITY_TYPES = ['Assurance', 'Audit Khusus', 'Audit Tujuan Tertentu', 'Audit Investigasi'] as const;
export type ActivityPlanActivityType = typeof ACTIVITY_PLAN_ACTIVITY_TYPES[number];

/** The Audit Activity Plan row that uses an annual plan row (server-computed, read-only). */
export interface AnnualActivityPlanRef {
  activityPlanId: string;
  activityPlanTitle: string;
  auditActivityId: string;
  assignmentCode: string;
}

/**
 * One row of the Annual Audit Plan (PKAT activities[] item): what is audited and with how many
 * resources. The why and how (Kode Penugasan, Justifikasi, Tujuan, Fokus, Ketua Tim) live on the
 * Audit Activity Plan row.
 */
export interface AuditActivities {
  /** Server-assigned UUID. Kept on edit and sent back on PUT. */
  id?: string;
  /** No. */
  itemNumber?: string;
  /** @deprecated Moved to the Audit Activity Plan. Read-only legacy value, never sent; the server keeps it. */
  activityCode?: string;
  /** Objek Penugasan Audit. */
  name: string;
  /** Audit category (Assurance, Special Audit, ...). Decides which Jenis Kegiatan Audit can use the row. */
  category: AuditCategory | string;
  /** Department id from the master list (GET /departments). */
  departmentId?: string;
  /** Department name snapshot, refreshed by the server. Read-only. */
  department?: string;
  /** Audit Universe (corporate) id picked in the form. */
  auditUniverseId?: string;
  /** Snapshots refreshed by the server from the Audit Universe. Read-only, never sent. */
  riskName?: string;
  riskLevel?: string;
  /** @deprecated Moved to the Audit Activity Plan. Read-only legacy value, never sent. */
  justification?: string;
  /** @deprecated Moved to the Audit Activity Plan. Read-only legacy value, never sent. */
  auditObjective?: string;
  /** @deprecated Moved to the Audit Activity Plan. Read-only legacy value, never sent. */
  auditFocus?: string;
  /** Jadwal Penugasan. */
  timelineText?: string;
  /** Jumlah Tim (Orang). */
  auditorCount?: number;
  /** Total Mandays. */
  totalMandays?: number;
  /** Supervisor (Pengendali Teknis). Not the Ketua Tim, which is chosen on the Audit Activity Plan. */
  supervisorName?: string;
  /** Set by the server when an Audit Activity Plan uses this row. Read-only, never sent. */
  activityPlanRef?: AnnualActivityPlanRef | null;
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
  title?: string;
  /** @deprecated Jenis Kegiatan Audit moved to the Audit Activity Plan. Read-only legacy value, never sent. */
  activityType?: string;
  version?: string;
  parentPlanId?: string;
  revisionHistory?: RevisionHistory[];
  activities: AuditActivities[];
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];
  quarters?: string[];
  year: string;
  notes?: string;
  attachmentCategory: string;
  attachmentUploadedBy: string;
  attachmentUploadDate: string;
  attachments?: AnnualAuditAttachment[];
  isActive: boolean;
  /** Set by the server: an Audit Activity Plan or audit activity uses this plan, so it can only be deactivated. */
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
  title: string;
  version?: string;
  parentPlanId?: string;
  revisionHistory?: RevisionHistory[];
  activities: AuditActivities[];
  status: AnnualAuditPlanStatus;
  selectedMonths: number[];
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

export interface AuditExecutionStatusItem {
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
  auditExecutionStatus: AuditExecutionStatusItem[];
  atrReports: AtrReport[];
}

export interface SampleItem {
  id: number;
  fieldworkDocument?: string;
  document: string;
  step1?: string;
  l1: TestResult;
  step2?: string;
  l2: TestResult;
  step3?: string;
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

export interface ActivityItem {
  id: number;
  name: string;
}

export interface WorkingPaperHeader {
  id?: string;
  assignmentLetterId: string;
  auditPurpose: string;
  businessProcess: string;
  period: string;
  location: string;
  teamMembers: TeamMember[];
  activities: ActivityItem[];
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
  activities: ActivityItem[];
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
  population: string | number | undefined;
  sampleSize: number | undefined;
  samples: SampleItem[];
  conclusion: string;
}

export interface WorkingPaperSampleForm {
  id?: string;
  workingPaperId?: string;
  assignmentLetterId?: string;
  population: string | number | undefined;
  sampleSize: number | undefined;
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
  companyId?: string;
  companyName?: string;
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
// The Audit Activity Plan is the Formulir Rencana Kegiatan Audit (RKA): one header (Tahun,
// Jenis Kegiatan Audit, Judul) plus one row per assignment. Fields owned by the Annual Audit
// Plan (PKAT) are read-only here and resolved live by the server. API: audit-service
// /activity-plans (docs/audit-activity-plan-rework.md, section 6).

/** formVersion of plans saved with the RKA form. Older plans (1) are legacy: view and delete only. */
export const ACTIVITY_PLAN_FORM_VERSION = 2;

export interface ActivityPlanNamedRef {
  id: string;
  name: string;
}

/** Audit Universe of a row: name and risk level only (it has no code). */
export interface ActivityPlanAuditUniverse {
  id: string;
  name: string;
  riskLevel: string;
}

/** The PKAT an Audit Activity Plan belongs to. */
export interface ActivityPlanAnnualPlan {
  id: string;
  code: string;
  title: string;
  year: number;
  status: string;
  isActive: boolean;
}

export interface ActivityPlanTotals {
  rowCount: number;
  teamSize: number;
  totalMandays: number;
}

/** GET /activity-plans item. */
export interface ActivityPlanSummary {
  id: string;
  formVersion: number;
  /** Client-derived: formVersion is not the RKA form, so the plan is view/delete only. */
  isLegacy: boolean;
  /** 0 when a legacy plan has no year. */
  year: number;
  /** "" for legacy plans. */
  activityType: ActivityPlanActivityType | '';
  title: string;
  /** null for legacy plans. */
  annualPlan: ActivityPlanAnnualPlan | null;
  totals: ActivityPlanTotals;
  status: string;
  attachmentCount: number;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface ActivityPlanTeamLeader {
  /** master-service employee UUID. */
  employeeId: string;
  name: string;
}

/** Fields of a row that come from the PKAT row. Read-only in the Audit Activity Plan. */
export interface ActivityPlanPkatFields {
  itemNumber: string;
  auditObject: string;
  auditCategory: string;
  auditUniverse: ActivityPlanAuditUniverse | null;
  department: ActivityPlanNamedRef | null;
  schedule: string;
  teamSize: number;
  totalMandays: number;
  pkatSupervisorName: string;
}

/** One row of GET /activity-plans/:id (an audit_activities row). */
export interface ActivityPlanRow extends Omit<ActivityPlanPkatFields, 'itemNumber'> {
  id: string;
  no: number;
  annualActivityId: string;
  assignmentCode: string;
  /** false once the engagement has started: the code is locked. */
  assignmentCodeEditable: boolean;
  auditKind: string;
  justification: string[];
  objectives: string[];
  focus: string[];
  teamLeader: ActivityPlanTeamLeader | null;
  /** Execution-driven status of the audit activity (PLANNED, ...). */
  status: string;
  /** The engagement has started: the row cannot be removed and its code is locked. */
  inUse: boolean;
  /** "snapshot" when the PKAT row no longer exists and saved values are shown. */
  source: 'pkat' | 'snapshot';
}

/** GET /activity-plans/:id. */
export interface ActivityPlanDetail {
  id: string;
  formVersion: number;
  isLegacy: boolean;
  annualPlanId: string | null;
  annualPlan: ActivityPlanAnnualPlan | null;
  year: number;
  activityType: ActivityPlanActivityType | '';
  title: string;
  status: string;
  rows: ActivityPlanRow[];
  totals: ActivityPlanTotals;
  attachmentCategory: string;
  attachments: AnnualAuditAttachment[];
  attachmentUploadedBy: string;
  attachmentUploadDate: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  /** Stored data of a legacy (formVersion 1) plan, as-is; null for RKA plans. */
  legacy: LegacyActivityPlan | null;
}

/** A row in the POST/PUT /activity-plans body. PKAT-owned fields are never sent. */
export interface ActivityPlanRowInput {
  /** Present for rows already saved; omitted for new rows. */
  id?: string;
  annualActivityId: string;
  /** "" = the server generates one (new row) or keeps the stored one (existing row). */
  assignmentCode: string;
  auditKind: string;
  justification: string[];
  objectives: string[];
  focus: string[];
  teamLeaderEmployeeId: string | null;
  /** Display snapshot of the picked employee; "" when none. */
  teamLeaderName: string;
}

/** POST/PUT /activity-plans body. */
export interface ActivityPlanInput {
  annualPlanId: string;
  activityType: ActivityPlanActivityType;
  title: string;
  /** Echoed on PUT for the optimistic lock. */
  updatedAt?: string;
  rows: ActivityPlanRowInput[];
  attachmentCategory: string;
  attachments: AnnualAuditAttachment[];
}

/** GET /activity-plans/annual-activity-options item: a PKAT row that can be added. */
export interface AnnualActivityOption extends ActivityPlanPkatFields {
  annualActivityId: string;
  /** false when another Audit Activity Plan already uses the row. */
  available: boolean;
  usedBy: { activityPlanId: string; activityPlanTitle: string; assignmentCode: string } | null;
  /** Legacy PKAT values, applied once when the row is added. */
  prefill: { assignmentCode: string; justification: string[]; objectives: string[]; focus: string[] };
  /** A legacy audit activity linked to this PKAT row; the server adopts it on save. */
  existingRow: { id: string; assignmentCode: string; status: string } | null;
}

/** error.details of a 400/409 from /activity-plans (sent as a JSON string). `row` is 1-based. */
export interface ActivityPlanErrorDetails {
  field?: string;
  row?: number;
  activityPlanId?: string;
  activityPlanTitle?: string;
  assignmentCode?: string;
  [key: string]: unknown;
}

/** One row as edited in the form. */
export interface ActivityPlanRowForm {
  /** Client-side key for v-for; not sent. */
  key: string;
  /** audit_activities id of a saved row. */
  id?: string;
  annualActivityId: string;
  assignmentCode: string;
  assignmentCodeEditable: boolean;
  auditKind: string;
  justification: string[];
  objectives: string[];
  focus: string[];
  teamLeaderEmployeeId: string | null;
  teamLeaderName: string;
  inUse: boolean;
  source: 'pkat' | 'snapshot';
  status: string;
  /** Read-only, shown from the PKAT row. */
  pkat: ActivityPlanPkatFields;
}

export interface ActivityPlanFormState {
  /** Set when editing. */
  id?: string;
  annualPlanId: string;
  /** The PKAT as the server returned it with the plan (edit), for display. */
  annualPlan: ActivityPlanAnnualPlan | null;
  activityType: ActivityPlanActivityType | '';
  title: string;
  updatedAt?: string;
  rows: ActivityPlanRowForm[];
  attachmentCategory: string;
  attachments: AnnualAuditAttachment[];
  /** New files, uploaded on save. */
  file: File[] | null;
}

// --- Legacy Audit Activity Plan (formVersion 1), shown read-only ---

export interface LegacyPlannedAuditActivity {
  id?: string;
  activityCode?: string;
  annualPlanId?: string;
  annualActivityId?: string;
  auditName: string;
  auditee: string;
  category: string;
  riskName?: string;
  riskLevel?: string;
  duration: number;
  priority: string;
  numberOfAuditors: number;
  estimatedSchedule: string;
  budgetEstimation: number;
}

export interface LegacyResourceAuditor {
  id?: string;
  name: string;
  position: string;
  competence: string;
  availability: string;
}

export interface LegacyActivityPlanBudget {
  totalEstimatedCost: number;
  totalAllocatedBudget: number;
  budgetNotes: string;
}

export interface LegacyActivityPlanReview {
  creatorName: string;
  creatorPosition: string;
  approverName: string;
  approverPosition: string;
  approvalDate: string;
  additionalNotes: string;
}

/** `legacy` of GET /activity-plans/:id: the stored JSON of an old-format plan. */
export interface LegacyActivityPlan {
  planTitle: string;
  planYear: string;
  planPeriodStart: string;
  planPeriodEnd: string;
  department: string;
  creationDate: string;
  plannedActivities: LegacyPlannedAuditActivity[];
  resourceAuditors: LegacyResourceAuditor[];
  budget: LegacyActivityPlanBudget;
  review: LegacyActivityPlanReview;
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
  /** "Operational" | "Financial" | "Quality" | "Issue" | "Efficiency", or "" when not set (older rows). */
  category?: string;
}

/** Evidence file attached to an ATR; downloaded through GET /action-taken-reports/:id/evidence/:evidenceId. */
export interface AtrEvidence {
  id: string
  file_name: string
  uploaded_at: string
  uploaded_by: string
}

/**
 * Action Taken Report: the follow-up of one Audit Result Report (LHA) finding, created by the backend
 * when the LHA is approved. Finding fields are a copy taken at creation. Status values are the backend's
 * UPPER_SNAKE (see utils/actionTakenReport.ts).
 */
export interface ActionTakenReport {
  id: string
  audit_result_report_id: string
  report_number: string
  report_title: string
  finding_id: string
  finding_title: string
  finding_category: string
  recommendation: string
  action_plan: string
  pic_user_id: string | null
  pic_name: string
  /** RFC3339 or null when not assigned yet. */
  due_date: string | null
  /** 0-100. */
  progress: number
  status: string
  /** Server-computed: past the due date and not COMPLETED/CANCELLED. */
  is_overdue: boolean
  /** Whole days past the due date when overdue, else 0. */
  overdue_days: number
  evidence: AtrEvidence[]
  review_note: string
  reviewed_by: string
  reviewed_at: string | null
  created_at: string
  updated_at: string
  /**
   * What the caller may do with this item now, from the server ('assign', 'update_action_plan',
   * 'upload_evidence', 'submit', 'review', 'cancel', 'delete', 'remind'); null when the response has no such list.
   */
  allowed_actions: string[] | null
  /** The LHA the ATR follows up (`lha` in the detail response); null when the server sent none. */
  lha: AtrLhaInfo | null
}

/** Header fields of the audit result report (LHA) an ATR belongs to. Strings are '' when the LHA has none. */
export interface AtrLhaInfo {
  status: string
  report_date: string | null
  audit_object: string
  department: string
  company_name: string
  audit_period: string
  executive_summary: string
  scope: string
  methodology: string
  finding_summary: string
  conclusion: string
  prepared_by: string
  reviewed_by: string
  approved_by: string
}

/** What the source modules reported for one execution (backend services/audit_execution_sync.Signals). */
export interface AuditExecutionSources {
  annual_plan: boolean
  assignment_letter: boolean
  fieldwork_records: number
  test_controls: number
  test_controls_tested: number
  test_controls_with_finding: number
  /** Fieldwork test controls with a non-empty mitigation_plan. */
  mitigation_plans?: number
  working_paper_records: number
  working_paper_causes: number
  /** Working paper plans (action plans) recorded on the working papers. */
  working_paper_plans?: number
  result_report: boolean
  result_report_approved: boolean
  /** LHA submitted and waiting review / CAE approval (not approved yet). Optional until the backend sends it. */
  result_report_in_review?: boolean
}

export interface AuditExecution {
  id: string
  ref: string
  name: string
  category: AuditCategory | string
  department?: AuditDepartment | string
  progress: number
  lead_auditor: string
  status: AuditStatus | string
  status_detail?: 'Late' | 'On Time' | string
  /** Per-module source data the progress is derived from (read-only). */
  sources?: AuditExecutionSources | null
  created_at?: string
  createdAt?: string
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
    badgeClass: 'bg-secondary-100 text-secondary-700 dark:bg-secondary-950 dark:text-secondary-100 border border-secondary-300 dark:border-secondary-700',
    iconClass: 'text-secondary-600 dark:text-secondary-400',
    numBgClass: 'bg-secondary-600 text-white',
    cardClass: 'border-secondary-300 dark:border-secondary-800 bg-secondary-50/40 dark:bg-secondary-950/30',
    minProgress: 100,
    maxProgress: 100
  }
]

export const getExecutionPhase = (progress: number = 0): ExecutionPhase =>
  EXECUTION_PHASES[getExecutionStageIndex(progress)]!
