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
  fileUrl?: string;
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
  recentFindings: RecentFinding[];
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
  risk: string;
  taxonomy: RiskTaxonomy;
  riskLevel: RiskLevel;
  controlDescription: string;
}

export interface WorkingPaperRiskForm {
  id?: string;
  risk: string;
  taxonomy: RiskTaxonomy;
  riskLevel: RiskLevel;
  controlDescription: string;
}

export interface WorkingPaperSample {
  id?: string;
  population: number | null;
  sampleSize: number | null;
  samples: SampleItem[];
  conclusion: string;
}

export interface WorkingPaperSampleForm {
  id?: string;
  population: number | null;
  sampleSize: number | null;
  samples: SampleItem[];
  conclusion: string;
}

export interface WorkingPaperCause {
  id?: string;
  condition: string;
  criteria: string;
  impact: string;
  evidenceFile: File | null;
  rootCause: RootCauseItem[];
}

export interface WorkingPaperCauseForm {
  id?: string;
  condition: string;
  criteria: string;
  impact: string;
  evidenceFile: File | null;
  rootCause: RootCauseItem[];
}

export interface WorkingPaperPlan {
  id?: string;
  recommendation: string;
  response: string;
  actionDescription: string;
  pic: string;
  periodAction: string;
}

export interface WorkingPaperPlanForm {
  id?: string;
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
  id: number;
  code: string;
  strategicObjective: string;
  kpi: string;
  unit: string;
  hibHig: HibHigType;
  periodType: PeriodType;
  selectedPeriod: string; // Q1-Q4 for Quartal, or year string for Yearly
  yearStart?: number;
  yearEnd?: number;
  actual: string;
  target: string;
  calculation: string;
  status: string;
}

export interface ActionTakenReport {
  id: string
  auditRef: string
  title: string
  department: AuditDepartment
  deadline: string
  status: AuditStatus
  auditObject?: string
  findingCategory?: AuditCategory
  condition?: string
  criteria?: string
  recommendation?: string
  pic?: string
  attachment?: string
  progressDescription?: string
}
