// Auth Types
export interface User {
  id: string;
  email: string;
  fullName: string;
  role: UserRole;
  department?: string;
  createdAt: string;
  updatedAt: string;
}

export enum UserRole {
  ADMIN = "admin",
  AUDITOR = "auditor",
  DEPARTMENT_HEAD = "department_head",
  AUDITEE = "auditee",
  VIEWER = "viewer",
  AUDIT_STAFF = "audit_staff",
  AUDIT_MANAGER = "audit_manager",
  CHIEF_AUDIT_EXECUTIVE = "chief_audit_executive",
}

export interface LoginCredentials {
  email: string;
  password: string;
  rememberMe?: boolean;
}

export interface RegisterData {
  email: string;
  password: string;
  fullName: string;
}
